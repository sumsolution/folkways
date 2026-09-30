import { execFileSync } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { EvalError } from './errors.ts';

/** The arm ref that means "the CLI package in the working tree, uncommitted changes included". */
export const WORKTREE = 'WORKTREE';

/** Git without the developer's config (hooks, signing, helpers), with a fixed identity. */
const GIT_ENV = {
  ...process.env,
  GIT_CONFIG_GLOBAL: '/dev/null',
  GIT_CONFIG_NOSYSTEM: '1',
  GIT_AUTHOR_NAME: 'eval',
  GIT_AUTHOR_EMAIL: 'eval@example.com',
  GIT_COMMITTER_NAME: 'eval',
  GIT_COMMITTER_EMAIL: 'eval@example.com',
};

function git(cwd: string, ...args: string[]): string {
  return execFileSync('git', args, { cwd, env: GIT_ENV, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
}

/**
 * The `packages/cli` directory for an arm: the working tree for `WORKTREE`, else extracted from
 * `ref` into `scratch`. Refs before the workspace move (ADR 0003) have no `packages/cli`.
 */
export function armPackage(repoRoot: string, ref: string, scratch: string): string {
  if (ref === WORKTREE) return join(repoRoot, 'packages/cli');
  if (ref.startsWith('-')) throw new EvalError(`bad ref ${ref}`);
  let sha: string;
  try {
    sha = git(repoRoot, 'rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`);
  } catch {
    throw new EvalError(`${ref} is not a commit in ${repoRoot}`);
  }
  const dest = join(scratch, sha);
  if (!existsSync(join(dest, 'packages/cli/src/cli.ts'))) {
    mkdirSync(dest, { recursive: true });
    let tar: Buffer;
    try {
      tar = execFileSync('git', ['archive', '--format=tar', sha, 'packages/cli'], {
        cwd: repoRoot,
        env: GIT_ENV,
        maxBuffer: 1 << 28,
        stdio: ['ignore', 'pipe', 'pipe'],
      });
    } catch {
      throw new EvalError(`${ref} (${sha.slice(0, 7)}) has no packages/cli; arms need a ref after the workspace move`);
    }
    execFileSync('tar', ['-x', '-C', dest], { input: tar });
  }
  return join(dest, 'packages/cli');
}

/** A team repo at `dest`: the arm's own `folkways init --no-install`, then the case's files, committed. */
export function buildTeamRepo(pkg: string, dest: string, caseRepo?: string): void {
  execFileSync(process.execPath, [join(pkg, 'src/cli.ts'), 'init', dest, '--no-install'], {
    env: GIT_ENV,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  if (caseRepo) {
    cpSync(caseRepo, dest, { recursive: true });
    git(dest, 'add', '-A');
    git(dest, 'commit', '-q', '--allow-empty', '-m', 'test: case fixture');
  }
}

/** Skill folder names in a team repo's `.claude/skills/`. */
export function workspaceSkills(dest: string): string[] {
  const dir = join(dest, '.claude/skills');
  return existsSync(dir) ? readdirSync(dir, { withFileTypes: true }).filter((d) => d.isDirectory()).map((d) => d.name).sort() : [];
}
