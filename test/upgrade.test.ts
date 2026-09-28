import { chmodSync, existsSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';
import { git, isClean } from '../src/git.ts';
import { type Installer, upgrade } from '../src/upgrade.ts';
import { addFrameworkContent, json, makePackage, makeTeam, put, read } from './helpers.ts';

/** Stands in for `npm install`: pins the version and links node_modules/.bin/folkways to a package copy. */
function fakeInstaller(packages: Record<string, string>): Installer {
  return (team, version) => {
    const pkg = packages[version];
    if (!pkg) throw Object.assign(new Error('npm error 404'), { stderr: `No matching version ${version}` });
    const manifest = json(join(team, 'package.json'));
    manifest.devDependencies['@folkways/cli'] = version;
    writeFileSync(join(team, 'package.json'), `${JSON.stringify(manifest, null, 2)}\n`);
    const bin = join(team, 'node_modules/.bin/folkways');
    put(bin, `#!/bin/sh\nexec node "${pkg}/src/cli.ts" "$@"\n`);
    chmodSync(bin, 0o755);
  };
}

let team: string;
let install: Installer;

beforeEach(() => {
  team = makeTeam(makePackage('0.1.0'));
  const next = makePackage('0.2.0');
  addFrameworkContent(next);
  install = fakeInstaller({ '0.1.0': makePackage('0.1.0'), '0.2.0': next });
});

describe('upgrade', () => {
  it("runs the new version's sync on a branch and commits", () => {
    expect(upgrade(team, 'v0.2.0', install)).toEqual({ branch: 'chore/folkways-0.2.0', from: '0.1.0', to: '0.2.0' });
    expect(git(team, 'branch', '--show-current')).toBe('chore/folkways-0.2.0');
    expect(git(team, 'log', '-1', '--format=%s')).toBe('chore(folkways): upgrade to 0.2.0');
    expect(isClean(team)).toBe(true);
    expect(json(join(team, 'package.json')).devDependencies['@folkways/cli']).toBe('0.2.0');
    expect(json(join(team, '.folkways/manifest.json')).version).toBe('0.2.0');
    expect(existsSync(join(team, '.claude/skills/start/SKILL.md'))).toBe(true);
  });

  it('works from a subfolder', () => {
    expect(upgrade(join(team, 'knowledge'), '0.2.0', install).branch).toBe('chore/folkways-0.2.0');
  });

  it('is a no-op at the same version', () => {
    expect(upgrade(team, '0.1.0', install).branch).toBeUndefined();
    expect(git(team, 'branch', '--show-current')).toBe('main');
    expect(git(team, 'branch', '--list', 'chore/*')).toBe('');
  });

  it('refuses a dirty tree', () => {
    put(join(team, 'scratch.md'), 'x');
    expect(() => upgrade(team, '0.2.0', install)).toThrow(/dirty/);
    expect(git(team, 'branch', '--show-current')).toBe('main');
  });

  it('refuses edited framework files before installing', () => {
    put(join(team, '.folkways/instructions.md'), 'mine\n');
    git(team, 'commit', '-qam', 'edit');
    expect(() => upgrade(team, '0.2.0', install)).toThrow(/modified {2}\.folkways\/instructions\.md/);
    expect(git(team, 'branch', '--show-current')).toBe('main');
  });

  it('refuses an existing branch', () => {
    git(team, 'branch', 'chore/folkways-0.2.0');
    expect(() => upgrade(team, '0.2.0', install)).toThrow(/already exists/);
  });

  it('goes back and cleans up when the install fails', () => {
    expect(() => upgrade(team, '9.9.9', install)).toThrow(/upgrade to 9\.9\.9 failed; back on main:\nNo matching version/);
    expect(git(team, 'branch', '--show-current')).toBe('main');
    expect(git(team, 'branch', '--list', 'chore/*')).toBe('');
    expect(isClean(team)).toBe(true);
    expect(read(join(team, 'package.json'))).toContain('"0.1.0"');
  });

  it('refuses outside a team repo', () => {
    expect(() => upgrade('/', '0.2.0', install)).toThrow(/inside the team repo/);
  });
});
