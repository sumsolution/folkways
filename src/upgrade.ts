import { execFileSync } from 'node:child_process';
import { join } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { branchExists, git, isClean } from './git.ts';
import { readManifest } from './manifest.ts';
import { PACKAGE_NAME } from './package.ts';
import { addDevExactCommand, detectPackageManager } from './package-manager.ts';
import { editsMessage, findEdits } from './sync.ts';

export type Installer = (team: string, version: string) => void;

/** Pin `version` with the team repo's own package manager. */
export const packageInstall: Installer = (team, version) => {
  const [cmd, args] = addDevExactCommand(detectPackageManager(team), `${PACKAGE_NAME}@${version}`);
  execFileSync(cmd, args, { cwd: team, stdio: ['ignore', 'pipe', 'pipe'] });
};

const SEMVER = /^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$/;

export interface UpgradeResult {
  branch: string | undefined;
  from: string;
  to: string;
}

/**
 * Install `version` on branch chore/folkways-<version>, run that version's sync, and commit (spec §3.1).
 * The new version's own `sync` runs, so its assets and rules apply.
 */
export function upgrade(cwd: string, versionArg: string, install: Installer = packageInstall): UpgradeResult {
  const version = versionArg.replace(/^v/, '');
  if (!SEMVER.test(version)) throw new FolkwaysError(`'${versionArg}' is not a version; use an exact one, e.g. 1.2.0`);
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
  if (start === 'HEAD') throw new FolkwaysError('detached HEAD; check out a branch first');
  // The tree was clean before we started, so every untracked file is ours. Ignored files
  // (node_modules) stay.
  const abandon = (): void => {
    git(team, 'checkout', '-q', '-f', start);
    git(team, 'clean', '-q', '-f', '-d');
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
