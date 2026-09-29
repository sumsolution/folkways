import { execFileSync } from 'node:child_process';
import { chmodSync, existsSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { git, isClean } from '../src/git.ts';
import { init, packageName, templateTarget } from '../src/init.ts';
import { json, makePackage, makeTeam, put, read, tmp } from './helpers.ts';

describe('templateTarget', () => {
  it.each([
    ['dot-gitignore', '.gitignore'],
    ['dot-claude/settings.json', '.claude/settings.json'],
    ['AGENTS.md.tmpl', 'AGENTS.md'],
    ['knowledge/README.md', 'knowledge/README.md'],
    ['docs/dot-x/y.md.tmpl', 'docs/.x/y.md'],
  ])('%s -> %s', (from, to) => expect(templateTarget(from)).toBe(to));
});

describe('init', () => {
  it('scaffolds, pins, syncs and commits', () => {
    const team = makeTeam(makePackage('1.2.3'));
    for (const f of ['team.yaml', '.gitignore', 'CLAUDE.md', 'knowledge/README.md', '.folkways/instructions.md']) {
      expect(existsSync(join(team, f)), f).toBe(true);
    }
    expect(read(join(team, 'CLAUDE.md'))).toContain('@AGENTS.md');
    expect(read(join(team, 'AGENTS.md'))).toMatch(/^@\.folkways\/instructions\.md$/m);
    expect(json(join(team, 'package.json'))).toMatchObject({ private: true, devDependencies: { '@folkways/cli': '1.2.3' } });
    expect(json(join(team, '.folkways/manifest.json'))).toMatchObject({ version: '1.2.3', skills: [] });
    expect(json(join(team, '.claude/settings.json')).permissions.ask).toContain('Bash(git push:*)');
    expect(read(join(team, '.gitignore'))).toContain('node_modules/');
    expect(git(team, 'log', '--format=%s')).toBe('chore(folkways): init at 1.2.3');
    expect(git(team, 'branch', '--show-current')).toBe('main');
    expect(isClean(team)).toBe(true);
  });

  it('names the package after the folder', () => {
    const dest = init(join(tmp(), 'My Team'), { install: false, pkg: makePackage('1.0.0') });
    expect(json(join(dest, 'package.json')).name).toBe('my-team');
  });

  it('creates missing parent folders', () => {
    const dest = init(join(tmp(), 'a/b/team'), { install: false, pkg: makePackage('1.0.0') });
    expect(statSync(dest).isDirectory()).toBe(true);
  });

  it('refuses a non-empty folder', () => {
    const dest = join(tmp(), 'team');
    put(join(dest, 'x'), '');
    expect(() => init(dest, { install: false, pkg: makePackage('1.0.0') })).toThrow(/not empty/);
  });
});

describe('packageName', () => {
  it.each([
    ['My Team', 'my-team'],
    ['.hidden', 'hidden'],
    ['__.team', 'team'],
    ['Team_A.B~c', 'team_a.b-c'],
    ['équipe', 'equipe'],
    ['(x)', 'x-'],
    ['-.-team', 'team'],
    ['x'.repeat(300), 'x'.repeat(214)],
  ])('%s -> %s', (folder, name) => expect(packageName(folder)).toBe(name));

  it.each(['...', '._', '', 'node_modules', 'favicon.ico', '中文', '!', ' node_modules'])('refuses %j', (folder) => {
    expect(() => packageName(folder)).toThrow(FolkwaysError);
    expect(() => packageName(folder)).toThrow(/valid npm package name/);
  });
});

describe('init names the package', () => {
  it('after a hidden folder', () => {
    const dest = init(join(tmp(), '.hidden'), { install: false, pkg: makePackage('1.0.0') });
    expect(json(join(dest, 'package.json')).name).toBe('hidden');
  });

  it('or refuses before creating anything', () => {
    const parent = tmp();
    expect(() => init(join(parent, '._'), { install: false, pkg: makePackage('1.0.0') })).toThrow(/valid npm package name/);
    expect(existsSync(join(parent, '._'))).toBe(false);
  });
});

describe('init with pnpm', () => {
  it('pins pnpm in the packageManager field', () => {
    const dest = init(join(tmp(), 'team'), { install: false, pkg: makePackage('1.0.0'), packageManager: 'pnpm' });
    expect(json(join(dest, 'package.json')).packageManager).toMatch(/^pnpm@\d+\.\d+\.\d+$/);
  });

  it('leaves packageManager unset for npm', () => {
    const dest = init(join(tmp(), 'team'), { install: false, pkg: makePackage('1.0.0'), packageManager: 'npm' });
    expect(json(join(dest, 'package.json')).packageManager).toBeUndefined();
  });
});

describe('init when the install fails', () => {
  const path = process.env.PATH;
  afterEach(() => {
    process.env.PATH = path;
  });

  function failingNpm(): void {
    const bin = join(tmp(), 'bin');
    put(join(bin, 'npm'), '#!/bin/sh\necho "npm error network" >&2\nexit 1\n');
    chmodSync(join(bin, 'npm'), 0o755);
    process.env.PATH = `${bin}:${path}`;
  }

  it('removes the folder it created and says why', () => {
    failingNpm();
    const dest = join(tmp(), 'team');
    expect(() => init(dest, { pkg: makePackage('1.0.0') })).toThrow(FolkwaysError);
    expect(() => init(dest, { pkg: makePackage('1.0.0') })).toThrow(/npm install failed/);
    expect(existsSync(dest)).toBe(false);
  });

  it('empties a folder that already existed, and keeps it', () => {
    failingNpm();
    const dest = tmp();
    expect(() => init(dest, { pkg: makePackage('1.0.0') })).toThrow(/npm install failed/);
    expect(readdirSync(dest)).toEqual([]);
  });
});

describe('init when the first commit fails', () => {
  const path = process.env.PATH;
  afterEach(() => {
    process.env.PATH = path;
  });

  it('removes the scaffold and reports the git error concisely', () => {
    const realGit = execFileSync('sh', ['-c', 'command -v git'], { encoding: 'utf8' }).trim();
    const bin = join(tmp(), 'bin');
    put(join(bin, 'git'), `#!/bin/sh\nif [ "$1" = commit ]; then echo "commit rejected" >&2; exit 1; fi\nexec ${realGit} "$@"\n`);
    chmodSync(join(bin, 'git'), 0o755);
    process.env.PATH = `${bin}:${path}`;
    const dest = join(tmp(), 'team');
    expect(() => init(dest, { install: false, pkg: makePackage('1.0.0') })).toThrow(FolkwaysError);
    expect(() => init(dest, { install: false, pkg: makePackage('1.0.0') })).toThrow(/git commit failed;[\s\S]*commit rejected/);
    expect(existsSync(dest)).toBe(false);
  });
});

describe('init without a git identity', () => {
  const saved = { ...process.env };
  afterEach(() => {
    process.env = { ...saved };
  });

  it('fails with git\'s message and removes the scaffold', () => {
    for (const k of ['GIT_AUTHOR_NAME', 'GIT_AUTHOR_EMAIL', 'GIT_COMMITTER_NAME', 'GIT_COMMITTER_EMAIL', 'EMAIL']) delete process.env[k];
    Object.assign(process.env, { GIT_CONFIG_COUNT: '1', GIT_CONFIG_KEY_0: 'user.useConfigOnly', GIT_CONFIG_VALUE_0: 'true' });
    const dest = join(tmp(), 'team');
    let error: unknown;
    try {
      init(dest, { install: false, pkg: makePackage('1.0.0') });
    } catch (e) {
      error = e;
    }
    expect(error).toBeInstanceOf(FolkwaysError);
    expect((error as Error).message).toMatch(/git commit failed;[\s\S]*(identity|user\.email|name)/i);
    expect(existsSync(dest)).toBe(false);
  });
});
