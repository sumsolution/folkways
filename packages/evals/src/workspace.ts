import { execFileSync } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { EvalError } from './errors.ts';

/** The arm ref that means "the CLI built from the working tree, uncommitted changes included". */
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

/** The ref where the Go CLI got `init` (#58); arms need it or later. */
const FIRST_GO_INIT = '141a6f4';

/** Builds the Go CLI in `root` (a checkout of this repo) to `out`, with the version changesets keep. */
function buildCli(root: string, out: string): void {
  const { version } = JSON.parse(readFileSync(join(root, 'packages/folkways/package.json'), 'utf8')) as { version: string };
  execFileSync('go', ['build', '-ldflags', `-X main.version=${version}`, '-o', out, './packages/folkways/cmd/folkways'], {
    cwd: root,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
}

/**
 * The `folkways` binary for an arm: built from the working tree for `WORKTREE`, else from `ref` extracted into
 * `scratch`. Refs before the Go CLI's `init` are refused.
 */
export function armCli(repoRoot: string, ref: string, scratch: string): string {
  if (ref === WORKTREE) {
    const bin = join(scratch, WORKTREE, 'folkways');
    buildCli(repoRoot, bin);
    return bin;
  }
  if (ref.startsWith('-')) throw new EvalError(`bad ref ${ref}`);
  let sha: string;
  try {
    sha = git(repoRoot, 'rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`);
  } catch {
    throw new EvalError(`${ref} is not a commit in ${repoRoot}`);
  }
  const dest = join(scratch, sha);
  const bin = join(dest, 'folkways');
  if (!existsSync(bin)) {
    try {
      git(repoRoot, 'cat-file', '-e', `${sha}:packages/folkways/internal/scaffold`);
    } catch {
      throw new EvalError(`${ref} (${sha.slice(0, 7)}) has no Go CLI init; arms need ${FIRST_GO_INIT} or later`);
    }
    mkdirSync(dest, { recursive: true });
    const tar = execFileSync('git', ['archive', '--format=tar', sha, 'go.mod', 'go.sum', 'packages/folkways', 'packages/content'], {
      cwd: repoRoot,
      env: GIT_ENV,
      maxBuffer: 1 << 28,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    execFileSync('tar', ['-x', '-C', dest], { input: tar });
    buildCli(dest, bin);
  }
  return bin;
}

/** A team repo at `dest`: the arm's own `folkways init --no-install`, then the case's files, committed. */
export function buildTeamRepo(cli: string, dest: string, caseRepo?: string): void {
  // Without the caller's package manager (pnpm, when run by `pnpm test`), init writes the same repo everywhere.
  const env: NodeJS.ProcessEnv = { ...GIT_ENV };
  delete env.npm_config_user_agent;
  execFileSync(cli, ['init', dest, '--no-install'], {
    env,
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
