import { execFileSync } from 'node:child_process';

/** Run git in `cwd` and return trimmed stdout. Throws on a non-zero exit. */
export function git(cwd: string, ...args: string[]): string {
  return execFileSync('git', args, { cwd, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
}

export function isClean(cwd: string): boolean {
  return git(cwd, 'status', '--porcelain') === '';
}

export function branchExists(cwd: string, branch: string): boolean {
  try {
    git(cwd, 'rev-parse', '--quiet', '--verify', `refs/heads/${branch}`);
    return true;
  } catch {
    return false;
  }
}
