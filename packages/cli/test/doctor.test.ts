import { chmodSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { doctor, hookValueFindings } from '../src/doctor.ts';
import { HOOK_BIN, HOOK_NAMES } from '../src/hooks.ts';
import { addFrameworkContent, json, makePackage, makeTeam, put, read, ASSETS, tmp } from './helpers.ts';

/** What `npm install` leaves: the CLI at `version` and the hook binary. */
function install(team: string, version = '1.0.0'): void {
  put(join(team, 'node_modules/@folkways/cli/package.json'), JSON.stringify({ name: '@folkways/cli', version }));
  put(join(team, HOOK_BIN), '#!/bin/sh\n');
  chmodSync(join(team, HOOK_BIN), 0o755);
}

function setup(): { pkg: string; team: string; home: string } {
  const pkg = makePackage('1.0.0');
  addFrameworkContent(pkg);
  const team = makeTeam(pkg);
  install(team);
  return { pkg, team, home: tmp() };
}

const messages = (team: string, pkg: string, home: string, level?: 'error' | 'warning'): string[] =>
  doctor(team, { pkg, home })
    .filter((f) => !level || f.level === level)
    .map((f) => f.message);

describe('doctor', () => {
  it('finds nothing in a freshly synced, installed team repo', () => {
    const { pkg, team, home } = setup();
    expect(doctor(team, { pkg, home })).toEqual([]);
  });

  it('stops at a repo with no folkways.yaml', () => {
    const { pkg, home } = setup();
    const empty = tmp();
    expect(doctor(empty, { pkg, home })).toEqual([
      { level: 'error', message: expect.stringMatching(/no folkways\.yaml.*setup skill/) },
    ]);
  });

  it('checks only folkways.yaml and personal skills in a repo without package.json (the plugin, no CLI)', () => {
    const { pkg, home } = setup();
    const solo = tmp();
    put(join(solo, 'folkways.yaml'), 'mode: solo\nhooks:\n  git-guard: off\n');
    expect(doctor(solo, { pkg, home })).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: off reads as false (hook off); write false' },
    ]);
  });

  it('reports a missing manifest', () => {
    const { pkg, team, home } = setup();
    rmSync(join(team, '.folkways/manifest.json'));
    expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/manifest\.json.*folkways sync/)]);
  });

  describe('versions', () => {
    it('reports node_modules on another version than package.json pins (a failed upgrade)', () => {
      const { pkg, team, home } = setup();
      install(team, '1.1.0');
      expect(messages(team, pkg, home, 'error')).toEqual([
        expect.stringMatching(/node_modules has @folkways\/cli 1\.1\.0 but package\.json pins 1\.0\.0.*npm install/),
      ]);
    });

    it('reports the CLI not installed', () => {
      const { pkg, team, home } = setup();
      rmSync(join(team, 'node_modules/@folkways'), { recursive: true });
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/not installed.*npm install/)]);
    });

    it('reports a CLI that is not pinned in package.json', () => {
      const { pkg, team, home } = setup();
      writeFileSync(join(team, 'package.json'), JSON.stringify({ name: 'team', private: true }));
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/package\.json doesn't pin @folkways\/cli/)]);
    });

    it('warns when the running CLI is not the pinned one (npx drift)', () => {
      const { team, home } = setup();
      const other = makePackage('2.0.0');
      expect(messages(team, other, home, 'warning')).toEqual([
        expect.stringMatching(/this is folkways 2\.0\.0.*pins 1\.0\.0.*npx folkways doctor/),
      ]);
    });

    it('reports a pin that is not an exact version, and skips the comparisons', () => {
      const { pkg, team, home } = setup();
      const pj = json(join(team, 'package.json'));
      pj.devDependencies['@folkways/cli'] = '^1.0.0';
      writeFileSync(join(team, 'package.json'), JSON.stringify(pj));
      expect(doctor(team, { pkg, home })).toEqual([
        { level: 'error', message: expect.stringMatching(/pins @folkways\/cli \^1\.0\.0, not an exact version.*folkways upgrade/) },
      ]);
    });

    it('does not send a failed upgrade to npx when the running CLI is the installed one', () => {
      const { team, home } = setup();
      install(team, '1.1.0');
      const installed = makePackage('1.1.0');
      expect(messages(team, installed, home, 'warning')).toEqual([]);
    });

    it('reports a sync from another version than the pinned one', () => {
      const { pkg, team, home } = setup();
      const manifest = json(join(team, '.folkways/manifest.json'));
      manifest.version = '0.9.0';
      writeFileSync(join(team, '.folkways/manifest.json'), JSON.stringify(manifest));
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/synced from 0\.9\.0.*pins 1\.0\.0.*folkways sync/)]);
    });
  });

  it('reports edited framework files, which sync refuses', () => {
    const { pkg, team, home } = setup();
    writeFileSync(join(team, '.folkways/instructions.md'), 'edited\n');
    expect(messages(team, pkg, home, 'error')).toEqual([
      ".folkways/instructions.md was modified, so sync refuses; move the change into an overlay, or run 'folkways sync --force' to discard it",
    ]);
  });

  it('reports team overlays changed since the last sync', () => {
    const { pkg, team, home } = setup();
    put(join(team, 'overlays/start/OVERLAY.md'), '## Add\n\n- x\n');
    expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/overlays\/start\/OVERLAY\.md.*folkways sync/)]);
  });

  it('warns about personal skills that shadow framework or team skills (M10)', () => {
    const { pkg, team, home } = setup();
    put(join(team, '.claude/skills/deploy/SKILL.md'), '---\nname: deploy\n---\n');
    put(join(home, '.claude/skills/start/SKILL.md'), 'x');
    put(join(home, '.copilot/skills/deploy/SKILL.md'), 'x');
    put(join(home, '.agents/skills/unrelated/SKILL.md'), 'x');
    expect(messages(team, pkg, home, 'warning')).toEqual([
      expect.stringMatching(/~\/\.claude\/skills\/start shadows the skill start/),
      expect.stringMatching(/~\/\.copilot\/skills\/deploy shadows the skill deploy/),
    ]);
  });

  describe('disableAllHooks (M11)', () => {
    it('reports true in .claude/settings.local.json', () => {
      const { pkg, team, home } = setup();
      put(join(team, '.claude/settings.local.json'), JSON.stringify({ disableAllHooks: true }));
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/settings\.local\.json sets disableAllHooks/)]);
    });

    it('reports project settings that no longer set it to false', () => {
      const { pkg, team, home } = setup();
      const path = join(team, '.claude/settings.json');
      const settings = json(path);
      delete settings.disableAllHooks;
      writeFileSync(path, JSON.stringify(settings));
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/settings\.json.*disableAllHooks: false.*folkways sync/)]);
    });

    it('reports unreadable settings', () => {
      const { pkg, team, home } = setup();
      put(join(team, '.claude/settings.local.json'), '{');
      expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/settings\.local\.json is not valid JSON/)]);
    });
  });

  it('reports a missing hook binary and pnpm hoisting', () => {
    const { pkg, team, home } = setup();
    rmSync(join(team, HOOK_BIN));
    const pj = json(join(team, 'package.json'));
    writeFileSync(join(team, 'package.json'), JSON.stringify({ ...pj, packageManager: 'pnpm@10.0.0' }));
    expect(messages(team, pkg, home, 'error')).toEqual([
      expect.stringMatching(/publicHoistPattern/),
      expect.stringMatching(/folkways-hook is missing/),
    ]);
  });

  it('checks the hooks: values in folkways.yaml', () => {
    const { pkg, team, home } = setup();
    const path = join(team, 'folkways.yaml');
    writeFileSync(path, read(path).replace('  git-guard: true', '  git-guard: off'));
    expect(messages(team, pkg, home, 'error')).toEqual([expect.stringMatching(/hooks\.git-guard: off/)]);
  });
});

it('HOOK_NAMES are the hooks assets/settings.json wires', () => {
  const commands = JSON.stringify(json(join(ASSETS, 'settings.json')).hooks);
  const wired = [...commands.matchAll(/folkways-hook ([a-z-]+)/g)].map((m) => m[1]);
  expect([...new Set(wired)].sort()).toEqual([...HOOK_NAMES]);
});

describe('hookValueFindings', () => {
  const known = ['git-guard', 'context-band', 'precompact', 'session-start'];
  const check = (block: string) => hookValueFindings(`mode: team\nhooks:\n${block}review:\n  trigger: pre-push\n`, known);

  it('accepts true and false, with comments', () => {
    expect(check('  git-guard: true\n  context-band: false # quiet\n\n  # a comment\n')).toEqual([]);
    expect(hookValueFindings('mode: team\n', known)).toEqual([]);
    expect(hookValueFindings('hooks: {}\n', known)).toEqual([]);
  });

  it('flags YAML 1.1 booleans: the binary reads them, run.sh may not', () => {
    expect(check('  git-guard: off\n  context-band: No\n  precompact: False\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: off reads as false (hook off); write false' },
      { level: 'error', message: 'folkways.yaml hooks.context-band: No reads as false (hook off); write false' },
      { level: 'error', message: 'folkways.yaml hooks.precompact: False reads as false (hook off); write false' },
    ]);
    expect(check('  git-guard: yes\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: yes reads as true (hook on); write true' },
    ]);
  });

  it('flags values that leave the hook on', () => {
    expect(check('  git-guard: "false"\n  context-band: 0\n  precompact: falsey\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: "false" is not a boolean, so the hook stays on; write true or false' },
      { level: 'error', message: 'folkways.yaml hooks.context-band: 0 is not a boolean, so the hook stays on; write true or false' },
      { level: 'error', message: 'folkways.yaml hooks.precompact: falsey is not a boolean, so the hook stays on; write true or false' },
    ]);
  });

  it('flags empty and null values and quoted YAML 1.1 words, which the binary reads as booleans', () => {
    expect(check('  git-guard:\n  context-band: ~ # unset\n  precompact: "off"\n  session-start: \'on\'\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: (empty) reads as false (hook off); write false' },
      { level: 'error', message: 'folkways.yaml hooks.context-band: ~ reads as false (hook off); write false' },
      { level: 'error', message: 'folkways.yaml hooks.precompact: "off" reads as false (hook off); write false' },
      { level: 'error', message: "folkways.yaml hooks.session-start: 'on' reads as true (hook on); write true" },
    ]);
  });

  it('reads CRLF files', () => {
    expect(hookValueFindings('hooks:\r\n  git-guard: "false"\r\n  precompact: false\r\n', known)).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: "false" is not a boolean, so the hook stays on; write true or false' },
    ]);
  });

  it('reports tabs, which make the hook binary ignore folkways.yaml', () => {
    expect(hookValueFindings('hooks:\n\tgit-guard: false\n', known)).toEqual([
      { level: 'error', message: 'folkways.yaml hooks: line 2 is indented with a tab, which YAML forbids, so the hooks ignore folkways.yaml; indent with spaces' },
    ]);
  });

  it('reports a nested map under a hook, which leaves it on', () => {
    expect(check('  git-guard:\n    enabled: false\n  precompact: false\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: a nested map is not a boolean, so the hook stays on; write true or false' },
    ]);
  });

  it('reads a plain value continued on the next line, and does not guess at tags or aliases', () => {
    expect(check('  git-guard:\n    off\n  precompact: !!bool "false"\n  context-band: *f\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: off reads as false (hook off); write false' },
      { level: 'error', message: 'folkways.yaml hooks.precompact: doctor can\'t check !!bool "false" (a tag, anchor or alias); write true or false' },
      { level: 'error', message: "folkways.yaml hooks.context-band: doctor can't check *f (a tag, anchor or alias); write true or false" },
    ]);
  });

  it('unquotes and trims keys, and accepts an empty or null hooks map', () => {
    expect(check('  "git-guard": off\n  precompact : false\n')).toEqual([
      { level: 'error', message: 'folkways.yaml hooks.git-guard: off reads as false (hook off); write false' },
    ]);
    expect(hookValueFindings('hooks: null\n', known)).toEqual([]);
    expect(hookValueFindings('hooks: ~ # none\n', known)).toEqual([]);
  });

  it('warns about unknown hook names', () => {
    expect(check('  git_guard: false\n')).toEqual([
      { level: 'warning', message: 'folkways.yaml hooks.git_guard: no hook by that name (hooks: context-band, git-guard, precompact, session-start)' },
    ]);
  });

  it('flags an inline hooks map it cannot check', () => {
    expect(hookValueFindings('hooks: { git-guard: false }\n', known)).toEqual([
      { level: 'warning', message: 'folkways.yaml hooks: write one hook per line (git-guard: false) so doctor can check the values' },
    ]);
  });
});
