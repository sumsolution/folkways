// Small helpers for the contract suite. Kept here so the suite doesn't import from another package.
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, realpathSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach } from 'vitest';

/** Run git in `cwd` and return trimmed stdout. Throws on a non-zero exit. */
export function git(cwd: string, ...args: string[]): string {
  return execFileSync('git', args, { cwd, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
}

const dirs: string[] = [];
afterEach(() => {
  for (const d of dirs.splice(0)) rmSync(d, { recursive: true, force: true });
});

/** A temp dir, removed after the test. Resolved, so paths compare equal to git's. */
export function tmp(): string {
  const d = realpathSync(mkdtempSync(join(tmpdir(), 'folkways-')));
  dirs.push(d);
  return d;
}

export const read = (path: string): string => readFileSync(path, 'utf8');
