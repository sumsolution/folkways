import { spawnSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import {
  ASSETS,
  buildPlugin,
  frameworkSkills,
  HOOK_TARGETS,
  hostTarget,
  pluginHooks,
  rewriteForPlugin,
} from '../src/build.ts';

const tmp = mkdtempSync(join(tmpdir(), 'folkways-claude-plugin-'));
afterAll(() => rmSync(tmp, { recursive: true, force: true }));

const SKILLS = ['commit', 'onboard', 'start'];

describe('rewriteForPlugin', () => {
  it('namespaces framework skill names', () => {
    expect(rewriteForPlugin('run the `onboard` skill first; the `start` skills', SKILLS)).toBe(
      'run the `folkways:onboard` skill first; the `folkways:start` skills',
    );
  });

  it('leaves other backticked names alone', () => {
    expect(rewriteForPlugin('the `deploy` skill and `start`', SKILLS)).toBe('the `deploy` skill and `start`');
  });

  it('points framework skill paths at the plugin root', () => {
    expect(rewriteForPlugin('(`.claude/skills/start/references/x.md`)', SKILLS)).toBe(
      '(`${CLAUDE_PLUGIN_ROOT}/skills/start/references/x.md`)',
    );
  });

  it('keeps team skill paths', () => {
    expect(rewriteForPlugin('a team skill in `.claude/skills/<name>/`', SKILLS)).toBe(
      'a team skill in `.claude/skills/<name>/`',
    );
  });

  it('says where the skills are', () => {
    expect(rewriteForPlugin('Skills in `.claude/skills/` hold the detail.', SKILLS)).toBe(
      'The `folkways:` skills hold the detail.',
    );
  });

  it('namespaces the framework skill list in the instructions', () => {
    expect(rewriteForPlugin('Framework skills: `start`, `commit`,\n`onboard`, `deploy`.', SKILLS)).toBe(
      'Framework skills: `folkways:start`, `folkways:commit`,\n`folkways:onboard`, `deploy`.',
    );
  });
});

describe('pluginHooks', () => {
  const settings = JSON.parse(readFileSync(join(ASSETS, 'settings.json'), 'utf8'));
  const hooks = pluginHooks(settings).hooks;
  const commands = (event: string) => hooks[event]!.flatMap((g) => g.hooks.map((h) => h.command));

  it('runs every team-repo hook, on the same event and matcher, through the plugin runner', () => {
    for (const [event, groups] of Object.entries(settings.hooks as Record<string, any[]>)) {
      for (const [i, group] of groups.entries()) {
        const plugin = hooks[event]!.filter((g) => !g.hooks.some((h) => h.command.endsWith(' instructions')));
        expect(plugin[i]!.matcher).toBe(group.matcher);
        for (const [j, h] of group.hooks.entries()) {
          const name = h.command.split(' ').at(-1);
          expect(plugin[i]!.hooks[j]!.command).toBe(`sh "\${CLAUDE_PLUGIN_ROOT}"/hooks/run.sh ${name}`);
        }
      }
    }
  });

  it('adds the instructions at session start', () => {
    expect(commands('SessionStart')).toContain('sh "${CLAUDE_PLUGIN_ROOT}"/hooks/run.sh instructions');
  });

  it('never points at a team repo', () => {
    expect(JSON.stringify(hooks)).not.toContain('node_modules');
  });
});

describe('buildPlugin', () => {
  const out = join(tmp, 'dist');
  const bins = join(tmp, 'bins');
  const host = hostTarget();

  beforeAll(() => {
    mkdirSync(bins);
    const binaries = new Map<string, string>();
    for (const t of HOOK_TARGETS) {
      const path = join(bins, t);
      writeFileSync(path, `#!/bin/sh\necho "stub ${t} $1"\n`);
      chmodSync(path, 0o755);
      binaries.set(t, path);
    }
    buildPlugin({ out, version: '1.2.3', binaries });
  });

  const read = (path: string) => readFileSync(join(out, path), 'utf8');
  const run = (hook: string, project: string, env: Record<string, string> = {}) =>
    spawnSync('/bin/sh', [join(out, 'hooks/run.sh'), hook], {
      input: '{}',
      encoding: 'utf8',
      env: { ...process.env, CLAUDE_PLUGIN_ROOT: out, CLAUDE_PROJECT_DIR: project, ...env },
    });

  it('writes the manifest at the package version', () => {
    expect(JSON.parse(read('.claude-plugin/plugin.json'))).toMatchObject({ name: 'folkways', version: '1.2.3' });
  });

  it('writes a publishable package.json without the dev marketplace', () => {
    const pkg = JSON.parse(read('package.json'));
    expect(pkg).toMatchObject({ name: '@folkways/claude-plugin', version: '1.2.3', license: 'Apache-2.0' });
    expect(pkg).not.toHaveProperty('private');
    expect(pkg.files).toContain('.claude-plugin/plugin.json');
    expect(pkg.files).not.toContain('.claude-plugin');
    expect(existsSync(join(out, 'LICENSE'))).toBe(true);
  });

  it('writes a dev marketplace that loads the build in place', () => {
    expect(JSON.parse(read('.claude-plugin/marketplace.json')).plugins).toEqual([
      expect.objectContaining({ name: 'folkways', source: './' }),
    ]);
  });

  it('copies every framework skill, rewritten for the plugin', () => {
    const skills = frameworkSkills();
    expect(skills.length).toBeGreaterThan(10);
    for (const skill of skills) {
      const text = read(`skills/${skill}/SKILL.md`);
      for (const s of skills) {
        expect(text, `${skill} still names the bare \`${s}\` skill`).not.toMatch(new RegExp(`\`${s}\` skill`));
        expect(text, `${skill} still points at .claude/skills/${s}/`).not.toContain(`.claude/skills/${s}/`);
      }
    }
    expect(existsSync(join(out, 'skills/start/references/status-template.md'))).toBe(true);
  });

  it('writes the rewritten instructions with a plugin preface', () => {
    const text = read('instructions.md');
    expect(text).toMatch(/^# folkways \(plugin\)/);
    expect(text).toContain('`folkways:onboard` skill');
  });

  it('ships a binary per platform, executable', () => {
    for (const t of HOOK_TARGETS) {
      expect(statSync(join(out, `hooks/bin/folkways-hook-${t}`)).mode & 0o111).not.toBe(0);
    }
  });

  describe('run.sh', () => {
    const plain = join(tmp, 'plain');
    const synced = join(tmp, 'synced');
    beforeAll(() => {
      mkdirSync(plain);
      mkdirSync(join(synced, '.folkways'), { recursive: true });
      writeFileSync(join(synced, '.folkways/manifest.json'), '{}');
    });

    it('prints the instructions', () => {
      const r = run('instructions', plain);
      expect(r.status).toBe(0);
      expect(r.stdout).toBe(read('instructions.md'));
    });

    it.skipIf(!host)('runs the host binary with the hook name', () => {
      const r = run('git-guard', plain);
      expect(r.status).toBe(0);
      expect(r.stdout.trim()).toBe(`stub ${host} git-guard`);
    });

    it('stays quiet in a synced team repo', () => {
      for (const hook of ['instructions', 'session-start', 'git-guard']) {
        const r = run(hook, synced);
        expect(r.status).toBe(0);
        expect(r.stdout).toBe('');
      }
    });

    it('without a binary, says so at session start and is silent otherwise', () => {
      const missing = { FOLKWAYS_HOOK_CMD: join(tmp, 'nope') };
      const start = run('session-start', plain, missing);
      expect(start.status).toBe(0);
      expect(start.stdout).toMatch(/hooks? .*off/);
      const guard = run('git-guard', plain, missing);
      expect(guard.status).toBe(0);
      expect(guard.stdout).toBe('');
    });
  });
});
