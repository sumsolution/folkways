import { chmodSync, existsSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { sync } from '../src/sync.ts';
import { addFrameworkContent, commitAll, json, makePackage, makeTeam, put, read } from './helpers.ts';

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
      'node_modules/.bin/folkways hook session-start',
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

  it('warns when AGENTS.md lost the import line', () => {
    put(join(team, 'AGENTS.md'), '# Team\n');
    expect(sync(team, next).warnings[0]).toMatch(/@\.folkways\/instructions\.md/);
  });

  it.each([
    [{ skills: ['../../outside'], files: {} }, /unsafe skill name "\.\.\/\.\.\/outside"/],
    [{ skills: [], files: { '../../outside': 'a'.repeat(64) } }, /unsafe path "\.\.\/\.\.\/outside"/],
    [{ skills: ['start'], files: { '.claude/skills/other/x.md': 'a'.repeat(64) } }, /unsafe path/],
    [{ skills: [], files: { 'team.yaml': 'a'.repeat(64) } }, /unsafe path "team\.yaml"/],
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
