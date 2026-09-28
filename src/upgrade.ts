import { execFileSync } from 'node:child_process';
import { join } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { branchExists, git, isClean } from './git.ts';
import { readManifest } from './manifest.ts';
import { PACKAGE_NAME } from './package.ts';
import { editsMessage, findEdits } from './sync.ts';

export type Installer = (team: string, version: string) => void;

export const npmInstall: Installer = (team, version) => {
  execFileSync('npm', ['install', '--save-dev', '--save-exact', `${PACKAGE_NAME}@${version}`], {
    cwd: team,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
};

export interface UpgradeResult {
  branch: string | undefined;
  from: string;
  to: string;
}

/**
 * Install `version` on branch chore/folkways-<version>, run that version's sync, and commit (spec §3.1).
 * The new version's own `sync` runs, so its assets and rules apply.
 */
export function upgrade(cwd: string, versionArg: string, install: Installer = npmInstall): UpgradeResult {
  const version = versionArg.replace(/^v/, '');
  let team: string;
  try {
    team = git(cwd, 'rev-parse', '--show-toplevel');
  } catch {
    throw new FolkwaysError('run this inside the team repo');
  }
  const manifest = readManifest(team);
  if (!manifest) throw new FolkwaysError('no .folkways/manifest.json here; is this a folkways team repo?');
  if (!isClean(team)) throw new FolkwaysError('working tree is dirty; commit or stash first');
  const edits = findEdits(team, manifest);
  if (edits.length > 0) {
    throw new FolkwaysError(editsMessage(edits, "Or run 'folkways sync --force' to discard them, commit, and rerun."));
  }
  const branch = `chore/folkways-${version}`;
  if (branchExists(team, branch)) throw new FolkwaysError(`branch ${branch} already exists`);

  const start = git(team, 'rev-parse', '--abbrev-ref', 'HEAD');
  const abandon = (): void => {
    git(team, 'checkout', '-q', '-f', start);
    git(team, 'branch', '-q', '-D', branch);
  };

  git(team, 'checkout', '-q', '-b', branch);
  try {
    install(team, version);
    execFileSync(join(team, 'node_modules/.bin/folkways'), ['sync'], {
      cwd: team,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
  } catch (e) {
    abandon();
    const detail = (e as { stderr?: Buffer | string }).stderr?.toString().trim() || (e as Error).message;
    throw new FolkwaysError(`upgrade to ${version} failed; back on ${start}:\n${detail}`);
  }

  git(team, 'add', '-A');
  try {
    git(team, 'diff', '--cached', '--quiet');
    abandon();
    return { branch: undefined, from: manifest.version, to: version };
  } catch {
    // staged changes exist
  }
  git(team, 'commit', '-q', '-m', `chore(folkways): upgrade to ${version}`, '-m', `From ${manifest.version}.`);
  return { branch, from: manifest.version, to: version };
}
