import { chmodSync, existsSync, lstatSync, mkdirSync, readdirSync, readlinkSync, rmSync, statSync, symlinkSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { fsOps } from '../src/files.ts';
import { staleOverlays, sync } from '../src/sync.ts';
import { addFrameworkContent, commitAll, json, makePackage, makeTeam, put, read, snapshot } from './helpers.ts';

let team: string;
let next: string;

beforeEach(() => {
  team = makeTeam(makePackage('0.1.0'));
  next = makePackage('0.2.0');
  addFrameworkContent(next);
});

const hookCommand = (settings: any, event: string) => settings.hooks?.[event]?.[0]?.hooks?.[0]?.command;

describe('sync to a new version', () => {
  it('adds new framework files and records them', () => {
    const result = sync(team, next);
    expect(result).toMatchObject({ version: '0.2.0', previous: '0.1.0', discarded: [], warnings: [] });
    expect(read(join(team, '.claude/skills/start/references/a.md'))).toBe('ref\n');
    expect(existsSync(join(team, '.folkways/prompts/review.md'))).toBe(true);
    expect(existsSync(join(team, '.github/workflows/folkways-knowledge-audit.yml'))).toBe(true);
    const manifest = json(join(team, '.folkways/manifest.json'));
    expect(manifest.version).toBe('0.2.0');
    expect(manifest.skills).toEqual(['start']);
    expect(Object.keys(manifest.files)).toContain('.claude/skills/start/SKILL.md');
    expect(hookCommand(json(join(team, '.claude/settings.json')), 'SessionStart')).toBe(
      '"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook session-start',
    );
  });

  it('keeps team customizations', () => {
    put(join(team, 'AGENTS.md'), `${read(join(team, 'AGENTS.md'))}\nTeam footer.\n`);
    put(join(team, 'overlays/start/OVERLAY.md'), 'team overlay\n');
    put(join(team, '.claude/skills/team-skill/SKILL.md'), 'team skill\n');
    const settings = json(join(team, '.claude/settings.json'));
    settings.permissions.allow = ['Bash(ls:*)'];
    settings.hooks = { Stop: [{ hooks: [{ type: 'command', command: 'scripts/team.sh' }] }] };
    writeFileSync(join(team, '.claude/settings.json'), JSON.stringify(settings));

    sync(team, next);

    expect(read(join(team, 'AGENTS.md'))).toContain('Team footer.');
    expect(read(join(team, 'overlays/start/OVERLAY.md'))).toBe('team overlay\n');
    expect(read(join(team, '.claude/skills/team-skill/SKILL.md'))).toBe('team skill\n');
    const after = json(join(team, '.claude/settings.json'));
    expect(after.permissions).toEqual({ ask: ['Bash(git push:*)'], allow: ['Bash(ls:*)'] });
    expect(hookCommand(after, 'Stop')).toBe('scripts/team.sh');
  });

  it('removes files and hooks the new version dropped', () => {
    sync(team, next);
    const later = makePackage('0.3.0');
    const { hooks: _, ...rest } = json(join(later, 'assets/settings.json'));
    put(join(later, 'assets/settings.json'), JSON.stringify(rest));
    sync(team, later);
    expect(existsSync(join(team, '.claude/skills/start'))).toBe(false);
    expect(existsSync(join(team, '.folkways/prompts'))).toBe(false);
    expect(existsSync(join(team, '.github'))).toBe(false);
    expect(json(join(team, '.claude/settings.json')).hooks).toBeUndefined();
  });

  it('preserves the executable bit', () => {
    put(join(next, 'assets/skills/start/scripts/run.sh'), '#!/bin/sh\n');
    chmodSync(join(next, 'assets/skills/start/scripts/run.sh'), 0o755);
    sync(team, next);
    expect(statSync(join(team, '.claude/skills/start/scripts/run.sh')).mode & 0o111).not.toBe(0);
  });
});

describe('sync stops on edits to framework files', () => {
  beforeEach(() => {
    sync(team, next);
    commitAll(team);
  });

  it.each([
    ['modified', () => put(join(team, '.folkways/instructions.md'), 'mine\n'), 'modified  .folkways/instructions.md'],
    ['deleted', () => rmSync(join(team, '.claude/skills/start/SKILL.md')), 'deleted   .claude/skills/start/SKILL.md'],
    ['added', () => put(join(team, '.claude/skills/start/extra.md'), 'x'), 'added     .claude/skills/start/extra.md'],
  ])('%s', (_, edit, line) => {
    edit();
    expect(() => sync(team, next)).toThrow(FolkwaysError);
    expect(() => sync(team, next)).toThrow(line);
  });

  it('--force discards the edits and reports them', () => {
    put(join(team, '.folkways/instructions.md'), 'mine\n');
    const result = sync(team, next, { force: true });
    expect(result.discarded).toEqual(['modified  .folkways/instructions.md']);
    expect(read(join(team, '.folkways/instructions.md'))).not.toBe('mine\n');
  });

});

describe('sync refusals and warnings', () => {
  it('refuses a team skill with a new framework skill name', () => {
    put(join(team, '.claude/skills/start/SKILL.md'), 'mine\n');
    expect(() => sync(team, next)).toThrow(/clash.*\.claude\/skills\/start/s);
    expect(read(join(team, '.claude/skills/start/SKILL.md'))).toBe('mine\n');
  });

  it('rejects invalid settings before writing anything', () => {
    put(join(team, '.claude/settings.json'), '{ not json');
    expect(() => sync(team, next)).toThrow(/settings\.json is not valid JSON/);
    expect(existsSync(join(team, '.claude/skills/start'))).toBe(false);
    expect(json(join(team, '.folkways/manifest.json')).version).toBe('0.1.0');
  });

  it('warns when dependencies are installed without the hook binary', () => {
    mkdirSync(join(team, 'node_modules/.bin'), { recursive: true });
    expect(sync(team, next).warnings).toEqual([expect.stringMatching(/node_modules\/\.bin\/folkways-hook is missing/)]);
    put(join(team, 'node_modules/.bin/folkways-hook'), '');
    expect(sync(team, next).warnings).toEqual([]);
  });

  it('does not look for the hook binary before dependencies are installed', () => {
    expect(sync(team, next).warnings).toEqual([]);
  });

  it.each([
    ['no pnpm-workspace.yaml', undefined, true],
    ['another pattern', 'publicHoistPattern:\n  - "*eslint*"\n', true],
    ['the pattern elsewhere', 'catalog:\n  x: "@folkways/hooks-*"\n', true],
    ['the pattern in a list', 'packages: []\npublicHoistPattern:\n  - "*eslint*"\n  - \'@folkways/hooks-*\'\n', false],
    ['the pattern inline', 'publicHoistPattern: ["@folkways/hooks-*"]\n', false],
    ['hoist everything', 'publicHoistPattern:\n  - "*"\n', false],
    ['shamefullyHoist', 'shamefullyHoist: true\n', false],
  ])('pnpm team with %s', (_, workspace, warns) => {
    const pkg = json(join(team, 'package.json'));
    put(join(team, 'package.json'), JSON.stringify({ ...pkg, packageManager: 'pnpm@10.0.0' }));
    if (workspace) put(join(team, 'pnpm-workspace.yaml'), workspace);
    const warnings = sync(team, next).warnings;
    if (warns) expect(warnings).toEqual([expect.stringMatching(/publicHoistPattern.*@folkways\/hooks-\*/)]);
    else expect(warnings).toEqual([]);
  });

  it('skips the pnpm check for a package manager it does not support', () => {
    const pkg = json(join(team, 'package.json'));
    put(join(team, 'package.json'), JSON.stringify({ ...pkg, packageManager: 'yarn@4.0.0' }));
    expect(sync(team, next).warnings).toEqual([]);
  });

  it('warns when AGENTS.md lost the import line', () => {
    put(join(team, 'AGENTS.md'), '# Team\n');
    expect(sync(team, next).warnings[0]).toMatch(/@\.folkways\/instructions\.md/);
  });

  it.each([
    [{ skills: ['../../outside'], files: {} }, /unsafe skill name "\.\.\/\.\.\/outside"/],
    [{ skills: [], files: { '../../outside': 'a'.repeat(64) } }, /unsafe path "\.\.\/\.\.\/outside"/],
    [{ skills: ['start'], files: { '.claude/skills/other/x.md': 'a'.repeat(64) } }, /unsafe path/],
    [{ skills: [], files: { 'team.yaml': 'a'.repeat(64) } }, /unsafe path "team\.yaml"/],
    [{ skills: [], files: {}, overlays: { '../x': 'a'.repeat(64) } }, /unsafe overlay path "\.\.\/x"/],
    [{ skills: [], files: {}, overlays: { 'overlays/a/b/OVERLAY.md': 'a'.repeat(64) } }, /unsafe overlay path/],
    [{ skills: [], files: {}, overlays: { 'overlays/a/OVERLAY.md': 'nope' } }, /bad checksum for overlays\/a\/OVERLAY\.md/],
    [{ skills: [], files: {}, overlays: [] }, /bad overlays/],
  ])('rejects an unsafe manifest before forced deletion: %j', (changes, message) => {
    const outside = join(team, '..', 'outside');
    put(outside, 'keep\n');
    put(
      join(team, '.folkways/manifest.json'),
      `${JSON.stringify({ version: '0.1.0', ...changes })}\n`,
    );
    expect(() => sync(team, next, { force: true })).toThrow(message);
    expect(() => sync(team, next, { force: true })).toThrow("Restore it with 'git checkout -- .folkways/manifest.json'");
    expect(read(outside)).toBe('keep\n');
    expect(existsSync(join(team, 'team.yaml'))).toBe(true);
  });

  it('rejects a manifest that is not JSON', () => {
    put(join(team, '.folkways/manifest.json'), '{ nope');
    expect(() => sync(team, next)).toThrow(/manifest\.json: not a JSON object/);
  });
});

describe('sync without a manifest (adopting a repo)', () => {
  beforeEach(() => {
    rmSync(join(team, '.folkways/manifest.json'));
    commitAll(team, 'drop manifest');
  });

  it('keeps files that already match', () => {
    expect(sync(team, next).discarded).toEqual([]);
    expect(json(join(team, '.folkways/manifest.json')).version).toBe('0.2.0');
  });

  it('refuses to overwrite files that differ, and lists them', () => {
    put(join(team, '.folkways/instructions.md'), 'ours\n');
    put(join(team, '.github/workflows/folkways-knowledge-audit.yml'), 'name: ours\n');
    const before = snapshot(team);
    expect(() => sync(team, next)).toThrow(FolkwaysError);
    expect(() => sync(team, next)).toThrow(
      /replaced  \.folkways\/instructions\.md\n.*replaced  \.github\/workflows\/folkways-knowledge-audit\.yml[\s\S]*--force/,
    );
    expect(snapshot(team)).toEqual(before);
  });

  it('refuses a folder where a file goes, without crashing', () => {
    mkdirSync(join(team, '.folkways/prompts/review.md'), { recursive: true });
    put(join(team, '.folkways/prompts/review.md/x'), 'x');
    expect(() => sync(team, next)).toThrow(/replaced  \.folkways\/prompts\/review\.md/);
  });

  it.each(['.folkways', '.claude'])('refuses a file where the %s folder goes', (dir) => {
    rmSync(join(team, dir), { recursive: true });
    put(join(team, dir), 'x\n');
    expect(() => sync(team, next)).toThrow(FolkwaysError);
    expect(() => sync(team, next)).toThrow(`replaced  ${dir}\n`);
    const result = sync(team, next, { force: true });
    expect(result.discarded).toContain(`replaced  ${dir}`);
    expect(json(join(team, '.folkways/manifest.json')).version).toBe('0.2.0');
  });

  it('--force overwrites them and reports them', () => {
    put(join(team, '.folkways/instructions.md'), 'ours\n');
    const result = sync(team, next, { force: true });
    expect(result.discarded).toEqual(['replaced  .folkways/instructions.md']);
    expect(read(join(team, '.folkways/instructions.md'))).toBe(read(join(next, 'assets/instructions.md')));
  });
});

describe('sync without a manifest, with framework skills in place', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  beforeEach(() => {
    sync(team, next);
    rmSync(join(team, '.folkways/manifest.json'));
    commitAll(team, 'drop manifest');
  });

  it('keeps the adopted files when the swap fails', () => {
    const before = snapshot(team);
    let calls = 0;
    const real = fsOps.rename;
    vi.spyOn(fsOps, 'rename').mockImplementation((from, to) => {
      if (++calls === 5) throw new Error('injected rename failure');
      real(from, to);
    });
    expect(() => sync(team, next)).toThrow(/injected/);
    vi.restoreAllMocks();
    expect(snapshot(team)).toEqual(before);
  });

  it('adopts identical skills', () => {
    expect(sync(team, next).discarded).toEqual([]);
    expect(json(join(team, '.folkways/manifest.json')).skills).toEqual(['start']);
  });

  it('refuses differing or extra skill files, and --force replaces them', () => {
    put(join(team, '.claude/skills/start/SKILL.md'), 'ours\n');
    put(join(team, '.claude/skills/start/extra.md'), 'extra\n');
    expect(() => sync(team, next)).toThrow(/replaced  \.claude\/skills\/start\/SKILL\.md\n.*replaced  \.claude\/skills\/start\/extra\.md/);
    const result = sync(team, next, { force: true });
    expect(result.discarded).toEqual(['replaced  .claude/skills/start/SKILL.md', 'replaced  .claude/skills/start/extra.md']);
    expect(existsSync(join(team, '.claude/skills/start/extra.md'))).toBe(false);
    expect(sync(team, next).discarded).toEqual([]);
  });
});

describe('sync compiles team overlays (ADR 0001)', () => {
  const skill = '---\nname: start\ndescription: test\n---\n\n## Format {#format}\n\nframework format\n';
  const overlay = '## Add\n\n- team rule\n\n## Override {#format}\n\nteam format\n';

  beforeEach(() => {
    put(join(next, 'assets/skills/start/SKILL.md'), skill);
    put(join(next, 'assets/skills/start/references/a.md'), '## Template {#template}\n\nframework template\n');
  });

  it('writes the compiled skill and records both checksums', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), overlay);
    expect(sync(team, next).warnings).toEqual([]);
    const text = read(join(team, '.claude/skills/start/SKILL.md'));
    expect(text).toContain('team format');
    expect(text).not.toContain('framework format');
    expect(text).toMatch(/## Team additions\n\n- team rule\n$/);
    const manifest = json(join(team, '.folkways/manifest.json'));
    expect(manifest.overlays).toEqual({ 'overlays/start/OVERLAY.md': expect.stringMatching(/^[a-f0-9]{64}$/) });
    expect(staleOverlays(team, manifest)).toEqual([]);
  });

  it('overrides a section in a reference file', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), '## Override {#template}\n\nteam template\n');
    sync(team, next);
    expect(read(join(team, '.claude/skills/start/references/a.md'))).toContain('team template');
    expect(read(join(team, '.claude/skills/start/SKILL.md'))).toBe(skill);
  });

  it('stops before writing anything on an unknown section ID', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), '## Override {#nope}\n\nx\n');
    commitAll(team);
    const before = snapshot(team);
    expect(() => sync(team, next)).toThrow(/overlays\/start\/OVERLAY\.md: .*#nope/);
    expect(snapshot(team)).toEqual(before);
  });

  it('recompiles after the overlay changes, without reporting an edit', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), overlay);
    sync(team, next);
    commitAll(team);
    put(join(team, 'overlays/start/OVERLAY.md'), '## Add\n\n- newer rule\n');
    expect(staleOverlays(team, json(join(team, '.folkways/manifest.json')))).toEqual(['overlays/start/OVERLAY.md']);
    expect(sync(team, next).discarded).toEqual([]);
    expect(read(join(team, '.claude/skills/start/SKILL.md'))).toContain('- newer rule');
    expect(read(join(team, '.claude/skills/start/SKILL.md'))).toContain('framework format');
  });

  it('reports a compiled skill edited by hand like any framework edit', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), overlay);
    sync(team, next);
    put(join(team, '.claude/skills/start/SKILL.md'), 'hand edit\n');
    expect(() => sync(team, next)).toThrow(/modified  \.claude\/skills\/start\/SKILL\.md/);
  });

  it('adopts a repo whose skills are already compiled', () => {
    put(join(team, 'overlays/start/OVERLAY.md'), overlay);
    sync(team, next);
    rmSync(join(team, '.folkways/manifest.json'));
    expect(sync(team, next).discarded).toEqual([]);
  });

  it('warns about an overlay for a skill the framework does not ship', () => {
    put(join(team, 'overlays/nope/OVERLAY.md'), '## Add\n\n- x\n');
    expect(sync(team, next).warnings).toEqual([
      'overlays/nope/OVERLAY.md: no framework skill named nope; the overlay is not applied',
    ]);
  });

  it('warns about personal overlays that try to override', () => {
    put(join(team, 'people/alex/overlays/start/OVERLAY.md'), '## Override {#format}\n\nmine\n');
    expect(sync(team, next).warnings).toEqual([
      'people/alex/overlays/start/OVERLAY.md: personal overlays may only add; remove ## Override {#format}',
    ]);
  });
});

describe('sync links .local/me to the person\'s folder', () => {
  const link = () => join(team, '.local/me');

  it('links people/<handle> from .local/me.yaml', () => {
    put(join(team, '.local/me.yaml'), 'handle: alex # me\nprojects_dir: /tmp\n');
    mkdirSync(join(team, 'people/alex/overlays'), { recursive: true });
    sync(team, next);
    expect(lstatSync(link()).isSymbolicLink()).toBe(true);
    expect(readlinkSync(link())).toBe('../people/alex');
  });

  it('repoints a link to another handle', () => {
    put(join(team, '.local/me.yaml'), 'handle: "sam"\n');
    mkdirSync(join(team, 'people/sam'), { recursive: true });
    symlinkSync('../people/alex', link());
    sync(team, next);
    expect(readlinkSync(link())).toBe('../people/sam');
  });

  it('leaves a correct link alone', () => {
    put(join(team, '.local/me.yaml'), 'handle: alex\n');
    mkdirSync(join(team, 'people/alex'), { recursive: true });
    symlinkSync('../people/alex', link());
    const before = lstatSync(link()).ino;
    sync(team, next);
    expect(lstatSync(link()).ino).toBe(before);
  });

  it('removes a link to another person when this handle has no folder, and warns', () => {
    put(join(team, '.local/me.yaml'), 'handle: sam\n');
    mkdirSync(join(team, 'people/alex'), { recursive: true });
    symlinkSync('../people/alex', link());
    expect(sync(team, next).warnings).toEqual([
      "people/sam/ not found; run the onboard skill. Skills won't attach personal overlays until it exists.",
    ]);
    expect(lstatSync(link(), { throwIfNoEntry: false })).toBeUndefined();
  });

  it('does nothing without a handle or a people folder', () => {
    sync(team, next);
    expect(existsSync(link())).toBe(false);
    put(join(team, '.local/me.yaml'), 'handle: alex\n');
    sync(team, next);
    expect(lstatSync(link(), { throwIfNoEntry: false })).toBeUndefined();
  });

  it('warns and leaves a real folder at .local/me', () => {
    put(join(team, '.local/me.yaml'), 'handle: alex\n');
    mkdirSync(join(team, 'people/alex'), { recursive: true });
    put(join(team, '.local/me/keep.md'), 'mine\n');
    expect(sync(team, next).warnings).toEqual([
      '.local/me is not a symlink; skills attach personal overlays through it. Move it and rerun sync.',
    ]);
    expect(read(join(team, '.local/me/keep.md'))).toBe('mine\n');
  });
});

describe('sync after an interrupted sync', () => {
  it('refuses and says how to recover', () => {
    put(join(team, '.folkways-sync-abc123/old/.folkways/manifest.json'), '{}');
    expect(() => sync(team, next)).toThrow(FolkwaysError);
    expect(() => sync(team, next, { force: true })).toThrow(/interrupted[\s\S]*\.folkways-sync-abc123/);
  });
});

describe('sync with a manifest', () => {
  it('refuses to overwrite a team file at a new framework path', () => {
    put(join(team, '.folkways/prompts/review.md'), 'team prompt\n');
    expect(() => sync(team, next)).toThrow(/replaced  \.folkways\/prompts\/review\.md/);
    expect(read(join(team, '.folkways/prompts/review.md'))).toBe('team prompt\n');
  });
});

describe('sync is atomic', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  beforeEach(() => {
    sync(team, next);
    commitAll(team);
    // The next version drops the skill's reference file and changes a prompt.
    next = makePackage('0.3.0');
    addFrameworkContent(next);
    rmSync(join(next, 'assets/skills/start/references'), { recursive: true });
    put(join(next, 'assets/prompts/review.md'), 'review prompt v3\n');
    put(join(next, 'assets/prompts/new.md'), 'new prompt\n');
  });

  const failOnCall = (op: 'copy' | 'rename', n: number) => {
    let calls = 0;
    const real = fsOps[op];
    vi.spyOn(fsOps, op).mockImplementation(((...args: [string, string]) => {
      if (++calls === n) throw new Error(`injected ${op} failure`);
      return real(...args);
    }) as any);
  };

  const expectUntouched = (before: Record<string, string>) => {
    expect(snapshot(team)).toEqual(before);
    expect(readdirSync(team).filter((f) => f.startsWith('.folkways-'))).toEqual([]);
    vi.restoreAllMocks();
    expect(sync(team, next)).toMatchObject({ version: '0.3.0', previous: '0.2.0', discarded: [] });
    expect(read(join(team, '.folkways/prompts/review.md'))).toBe('review prompt v3\n');
    expect(existsSync(join(team, '.claude/skills/start/references'))).toBe(false);
  };

  it.each([1, 3, 5])('leaves the old files when copy %i fails', (n) => {
    const before = snapshot(team);
    failOnCall('copy', n);
    expect(() => sync(team, next)).toThrow(/injected copy failure/);
    expectUntouched(before);
  });

  it.each([1, 4, 8, 12])('leaves the old files when rename %i fails', (n) => {
    const before = snapshot(team);
    failOnCall('rename', n);
    expect(() => sync(team, next)).toThrow(/injected rename failure/);
    expectUntouched(before);
  });
});
