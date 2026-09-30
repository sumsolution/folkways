// Harness for the black-box hook contract suite (README.md). Builds a synthetic team repo per fixture,
// runs `$FOLKWAYS_HOOK_CMD <hook>` through `sh -c` like the hook runners do, and checks the result.
import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { git, tmp } from './util.ts';

export const FIXTURES = resolve(import.meta.dirname, 'fixtures');
export const HOOKS = ['context-band', 'precompact', 'session-start', 'session-end', 'git-guard'] as const;

/** One call of the hook. `payload`, `cwd`, `files` and `expect` may use the placeholders in `Vars`. */
export interface Step {
  payload: unknown;
  /** Process cwd. Default: the team repo. Claude Code's hook cwd follows the Bash tool's `cd`. */
  cwd?: string;
  /** Replace the transcript before this step (token counts of assistant lines, oldest first). */
  transcript?: TranscriptLine[];
  expect: Expect;
}
export interface Expect {
  /** Default 0. */
  exit?: number;
  /** `null`: stdout is empty. Object: stdout is JSON containing this (deep subset; `re:` strings are regexes, case-insensitive). */
  stdout: null | Record<string, unknown>;
  /** Paths that must exist / must not exist after the step. */
  exists?: string[];
  absent?: string[];
  /** Path → regex its content must match. */
  contains?: Record<string, string>;
}
export interface Fixture {
  description: string;
  env?: {
    /** Deep-merged into the base team.yaml. */
    team?: Record<string, unknown>;
    /** Deep-merged into the base .local/me.yaml; `null` removes the file. */
    me?: Record<string, unknown> | null;
    /** Extra files, path relative to the team repo → content. */
    files?: Record<string, string>;
    /** Default true. False: CLAUDE_PROJECT_DIR is unset (Copilot CLI sets it for plugin hooks only). */
    claudeProjectDir?: boolean;
    transcript?: TranscriptLine[];
  };
  steps: Step[];
}
/** A number is a main-chain assistant line with that many tokens in context; an object is written as is. */
export type TranscriptLine = number | { sidechain: number } | Record<string, unknown>;

export interface Vars {
  TEAM: string;
  PROJECTS: string;
  MAIN: string;
  WORKTREE: string;
  HEAD: string;
  HEAD7: string;
  OLD_HEAD: string;
  OLD_HEAD7: string;
  TRANSCRIPT: string;
}

const BASE_TEAM = {
  team: { name: 'Example Team', slug: 'example' },
  members: [{ handle: 'alex', github: 'alex-gh', name: 'Alex Doe' }],
  git: { commit_convention: 'angular', scopes: ['web-app', 'auth', 'ci'], branch_pattern: '{type}/{ticket}-{slug}', merge: 'squash' },
  review: { trigger: 'pre-push' },
  handoff: { bands: [125000, 150000, 200000] },
};

export function loadFixtures(hook: string): [string, Fixture][] {
  const dir = join(FIXTURES, hook);
  if (!existsSync(dir)) return [];
  return readdirSync(dir)
    .filter((f) => f.endsWith('.json'))
    .sort()
    .map((f) => [f.replace(/\.json$/, ''), JSON.parse(readFileSync(join(dir, f), 'utf8')) as Fixture]);
}

type Plain = Record<string, unknown>;
const isPlain = (v: unknown): v is Plain => typeof v === 'object' && v !== null && !Array.isArray(v);

export function merge(base: Plain, over: Plain): Plain {
  const out: Plain = { ...base };
  for (const [k, v] of Object.entries(over)) out[k] = isPlain(v) && isPlain(out[k]) ? merge(out[k] as Plain, v) : v;
  return out;
}

/** Replace `${NAME}` placeholders in every string inside `value`. */
export function fill<T>(value: T, vars: Vars): T {
  if (typeof value === 'string') {
    return value.replace(/\$\{(\w+)\}/g, (m, k: string) => (k in vars ? vars[k as keyof Vars] : m)) as T;
  }
  if (Array.isArray(value)) return value.map((v) => fill(v, vars)) as T;
  if (isPlain(value)) return Object.fromEntries(Object.entries(value).map(([k, v]) => [fill(k, vars), fill(v, vars)])) as T;
  return value;
}

/** Deep subset match. Returns the first mismatch as a path + reason, or undefined. */
export function mismatch(expected: unknown, actual: unknown, path = '$'): string | undefined {
  if (typeof expected === 'string' && expected.startsWith('re:')) {
    return typeof actual === 'string' && new RegExp(expected.slice(3), 'is').test(actual)
      ? undefined
      : `${path}: ${JSON.stringify(actual)} does not match /${expected.slice(3)}/i`;
  }
  if (isPlain(expected)) {
    if (!isPlain(actual)) return `${path}: expected an object, got ${JSON.stringify(actual)}`;
    for (const [k, v] of Object.entries(expected)) {
      const m = mismatch(v, actual[k], `${path}.${k}`);
      if (m) return m;
    }
    return undefined;
  }
  if (Array.isArray(expected)) {
    if (!Array.isArray(actual) || actual.length !== expected.length) return `${path}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`;
    for (const [i, v] of expected.entries()) {
      const m = mismatch(v, actual[i], `${path}[${i}]`);
      if (m) return m;
    }
    return undefined;
  }
  return Object.is(expected, actual) ? undefined : `${path}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`;
}

/** Minimal YAML writer for the config shapes above (flow style, JSON-compatible). */
function yaml(value: Plain): string {
  return Object.entries(value).map(([k, v]) => `${k}: ${JSON.stringify(v)}`).join('\n') + '\n';
}

export function transcriptJsonl(lines: TranscriptLine[]): string {
  return lines
    .map((l) => {
      if (typeof l === 'number') return assistant(l, false);
      if ('sidechain' in l && typeof l.sidechain === 'number') return assistant(l.sidechain, true);
      return l;
    })
    .map((l) => JSON.stringify(l))
    .join('\n') + '\n';
}

function assistant(tokens: number, sidechain: boolean): Plain {
  // Realistic split: most of the context is a cache read.
  const input = Math.min(tokens, 8);
  const creation = Math.min(tokens - input, 1500);
  return {
    type: 'assistant',
    isSidechain: sidechain,
    message: {
      role: 'assistant',
      model: 'claude-example',
      usage: { input_tokens: input, cache_creation_input_tokens: creation, cache_read_input_tokens: tokens - input - creation, output_tokens: 120 },
    },
  };
}

/**
 * The synthetic environment: a team repo with team.yaml and .local/me.yaml, and a projects dir holding a main
 * clone (`web-app`, two commits on main) and a worktree of it (`.worktrees/web-app/feat-orders`).
 */
export function buildEnv(fixture: Fixture): Vars {
  const root = tmp();
  const team = join(root, 'team');
  const projects = join(root, 'projects');
  const main = join(projects, 'web-app');
  const worktree = join(projects, '.worktrees', 'web-app', 'feat-orders');
  mkdirSync(team, { recursive: true });
  mkdirSync(main, { recursive: true });

  git(main, 'init', '-q', '-b', 'main');
  writeFileSync(join(main, 'README.md'), 'one\n');
  git(main, 'add', '-A');
  git(main, 'commit', '-q', '-m', 'chore: first');
  const oldHead = git(main, 'rev-parse', 'HEAD');
  writeFileSync(join(main, 'README.md'), 'two\n');
  git(main, 'commit', '-q', '-am', 'chore: second');
  git(main, 'worktree', 'add', '-q', '-b', 'feat/EX-1-orders', worktree);
  writeFileSync(join(worktree, 'orders.md'), 'orders\n');
  git(worktree, 'add', '-A');
  git(worktree, 'commit', '-q', '-m', 'feat(web-app): add orders');
  const head = git(worktree, 'rev-parse', 'HEAD');

  const vars: Vars = {
    TEAM: team, PROJECTS: projects, MAIN: main, WORKTREE: worktree, HEAD: head, HEAD7: head.slice(0, 7), OLD_HEAD: oldHead, OLD_HEAD7: oldHead.slice(0, 7),
    TRANSCRIPT: join(root, 'transcript.jsonl'),
  };
  const env = fixture.env ?? {};
  writeFileSync(join(team, 'team.yaml'), yaml(merge(BASE_TEAM, env.team ?? {})));
  if (env.me !== null) {
    const me = merge({ handle: 'alex', projects_dir: projects, tool: 'claude' }, fill(env.me ?? {}, vars));
    mkdirSync(join(team, '.local'), { recursive: true });
    writeFileSync(join(team, '.local', 'me.yaml'), yaml(me));
  }
  for (const [path, content] of Object.entries(fill(env.files ?? {}, vars))) {
    mkdirSync(resolve(team, path, '..'), { recursive: true });
    writeFileSync(join(team, path), content);
  }
  if (env.transcript) writeFileSync(vars.TRANSCRIPT, transcriptJsonl(env.transcript));
  return vars;
}

export interface Result {
  exit: number | null;
  stdout: string;
  stderr: string;
}

/**
 * The hook process environment. `HOME` is the temp root that holds the team repo and the projects
 * directory, so fixtures can `cd ~/projects/web-app` or `cd "$HOME/projects/web-app"`.
 */
export function hookEnv(vars: Vars, claudeProjectDir: boolean): NodeJS.ProcessEnv {
  const env: NodeJS.ProcessEnv = { ...process.env };
  delete env.CLAUDE_PROJECT_DIR;
  if (claudeProjectDir) env.CLAUDE_PROJECT_DIR = vars.TEAM;
  env.HOME = resolve(vars.PROJECTS, '..');
  return env;
}

export function runHook(command: string, hook: string, step: Step, vars: Vars, claudeProjectDir: boolean): Result {
  if (step.transcript) writeFileSync(vars.TRANSCRIPT, transcriptJsonl(step.transcript));
  const env = hookEnv(vars, claudeProjectDir);
  const r = spawnSync('/bin/sh', ['-c', `${command} ${hook}`], {
    cwd: fill(step.cwd, vars) ?? vars.TEAM,
    env,
    input: JSON.stringify(fill(step.payload, vars)),
    encoding: 'utf8',
    timeout: 10_000,
  });
  return { exit: r.status, stdout: r.stdout ?? '', stderr: r.stderr ?? (r.error?.message ?? '') };
}

/** Every problem with a result, empty when it meets the expectation. */
export function check(expect: Expect, r: Result, vars: Vars): string[] {
  const problems: string[] = [];
  const want = fill(expect, vars);
  if (r.exit !== (want.exit ?? 0)) problems.push(`exit ${r.exit}, want ${want.exit ?? 0} (stderr: ${r.stderr.trim().slice(0, 200)})`);
  if (want.stdout === null) {
    if (r.stdout.trim() !== '') problems.push(`stdout should be empty, got: ${r.stdout.trim().slice(0, 300)}`);
  } else {
    let got: unknown;
    try {
      got = JSON.parse(r.stdout);
    } catch {
      problems.push(`stdout is not JSON: ${r.stdout.trim().slice(0, 300)}`);
    }
    if (got !== undefined) {
      const m = mismatch(want.stdout, got);
      if (m) problems.push(`stdout ${m}`);
    }
  }
  for (const p of want.exists ?? []) if (!existsSync(p)) problems.push(`missing file ${p}`);
  for (const p of want.absent ?? []) if (existsSync(p)) problems.push(`unexpected file ${p}`);
  for (const [p, re] of Object.entries(want.contains ?? {})) {
    const text = existsSync(p) ? readFileSync(p, 'utf8') : undefined;
    if (text === undefined || !new RegExp(re, 'is').test(text)) problems.push(`${p} does not match /${re}/i`);
  }
  return problems;
}
