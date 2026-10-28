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

/**
 * The CLI's version in `root`: the release-please manifest (ADR 0013), else `packages/folkways/package.json` for refs
 * from before it.
 */
function cliVersion(root: string): string {
  const manifest = join(root, '.release-please-manifest.json');
  if (existsSync(manifest)) return (JSON.parse(readFileSync(manifest, 'utf8')) as { '.': string })['.'];
  return (JSON.parse(readFileSync(join(root, 'packages/folkways/package.json'), 'utf8')) as { version: string }).version;
}

/** Builds the Go CLI in `root` (a checkout of this repo) to `out`, with its version. */
function buildCli(root: string, out: string): void {
  const version = cliVersion(root);
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
      git(repoRoot, 'merge-base', '--is-ancestor', FIRST_GO_INIT, sha);
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

/** A team repo at `dest`: the arm's own `folkways init`, then the case's files, committed. */
export function buildTeamRepo(cli: string, dest: string, caseRepo?: string): void {
  // Refs before Go init dropped npm need `--no-install`; later ones reject it before touching `dest`. Without
  // the caller's package manager (pnpm under `pnpm test`), the older init writes the same repo everywhere.
  const env: NodeJS.ProcessEnv = { ...GIT_ENV };
  delete env.npm_config_user_agent;
  const init = (...flags: string[]) => execFileSync(cli, ['init', dest, ...flags], { env, stdio: ['ignore', 'pipe', 'pipe'] });
  try {
    init('--no-install');
  } catch (e) {
    if (!String((e as { stderr?: unknown }).stderr).includes("Unknown option '--no-install'")) throw e;
    init();
  }
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
