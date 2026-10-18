import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import { EvalError } from './errors.ts';
import { parseDoc, type Value } from './frontmatter.ts';

/** One version of the framework under test: a git ref of this repo, or `WORKTREE` for the working tree. */
export interface Arm {
  name: string;
  ref: string;
}

export type Target = 'last_message' | 'transcript';

interface GraderBase {
  name: string;
  weight: number;
}
export interface RegexGrader extends GraderBase {
  type: 'regex';
  pattern: string;
  flags: string;
  match: 'contains' | 'not_contains';
  target: Target;
}
export interface ToolUsedGrader extends GraderBase {
  type: 'tool_used';
  tool: string;
  inputMatch?: string;
  min: number;
  max: number | undefined;
}
export interface LlmGrader extends GraderBase {
  type: 'llm';
  criteria: string;
  focus: Target;
  votes: number;
}
export type Grader = RegexGrader | ToolUsedGrader | LlmGrader;

export interface Case {
  name: string;
  dir: string;
  prompt: string;
  runs: number;
  model?: string;
  maxTurns: number;
  timeoutSeconds: number;
  allowedTools: string[];
  graders: Grader[];
  /** Files laid over the team repo that `folkways init` makes. */
  repo?: string;
}

export interface Suite {
  dir: string;
  name: string;
  arms: Arm[];
  cases: Case[];
  model?: string;
  judgeModel?: string;
}

const READ_ONLY = ['Read', 'Glob', 'Grep', 'Skill'];
const PROMPT_KEYS = ['description', 'runs', 'model', 'max_turns', 'timeout_seconds', 'allowed_tools'];
const GRADER_KEYS: Record<Grader['type'], string[]> = {
  regex: ['pattern', 'flags', 'match', 'target'],
  tool_used: ['tool', 'input_match', 'min', 'max'],
  llm: ['focus', 'votes'],
};

/**
 * Load `<dir>/suite.json` and every case under `<dir>/cases/`. The case format follows
 * `claude plugin eval` (prompt.md + graders/*.md), so cases read the same in both.
 */
export function loadSuite(dir: string): Suite {
  const config = JSON.parse(readFileSync(join(dir, 'suite.json'), 'utf8')) as {
    arms?: Arm[];
    model?: string;
    judge_model?: string;
  };
  const arms = config.arms ?? [];
  if (arms.length === 0) throw new EvalError(`${dir}/suite.json: needs at least one arm`);
  if (new Set(arms.map((a) => a.name)).size !== arms.length) throw new EvalError('arm names must be unique');
  for (const a of arms) {
    if (!/^[\w.-]+$/.test(a.name ?? '') || typeof a.ref !== 'string') {
      throw new EvalError(`arm ${JSON.stringify(a)}: needs a name ([\\w.-]+) and a ref`);
    }
  }
  const casesDir = join(dir, 'cases');
  const cases = readdirSync(casesDir, { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => loadCase(join(casesDir, d.name)))
    .sort((a, b) => a.name.localeCompare(b.name));
  return {
    dir,
    name: basename(dir),
    arms,
    cases,
    ...(config.model && { model: config.model }),
    ...(config.judge_model && { judgeModel: config.judge_model }),
  };
}

function loadCase(dir: string): Case {
  const name = basename(dir);
  const { data, body } = parseDoc(readFileSync(join(dir, 'prompt.md'), 'utf8'));
  checkKeys(`${name}/prompt.md`, data, PROMPT_KEYS);
  const gradersDir = join(dir, 'graders');
  const files = existsSync(gradersDir) ? readdirSync(gradersDir).filter((f) => f.endsWith('.md')).sort() : [];
  if (files.length === 0) throw new EvalError(`case ${name}: no graders in graders/`);
  const model = str(data.model);
  return {
    name,
    dir,
    prompt: body,
    runs: num(data.runs, 3),
    ...(model && { model }),
    maxTurns: num(data.max_turns, 30),
    timeoutSeconds: num(data.timeout_seconds, 600),
    allowedTools: Array.isArray(data.allowed_tools) ? data.allowed_tools.map(String) : READ_ONLY,
    graders: files.map((f) => loadGrader(`${name}/graders/${f}`, join(gradersDir, f))),
    ...(existsSync(join(dir, 'repo')) && { repo: join(dir, 'repo') }),
  };
}

function loadGrader(label: string, path: string): Grader {
  const { data, body } = parseDoc(readFileSync(path, 'utf8'));
  const type = data.type;
  if (type !== 'regex' && type !== 'tool_used' && type !== 'llm') {
    throw new EvalError(`${label}: unknown grader type ${JSON.stringify(type)} (regex, tool_used, llm)`);
  }
  checkKeys(label, data, ['type', 'weight', ...GRADER_KEYS[type]]);
  const base = { name: basename(path, '.md'), weight: num(data.weight, 1) };
  if (type === 'regex') {
    const pattern = str(data.pattern) ?? '';
    const flags = str(data.flags) ?? '';
    try {
      new RegExp(pattern, flags);
    } catch (e) {
      throw new EvalError(`${label}: bad pattern: ${(e as Error).message}`);
    }
    const match = data.match === 'not_contains' ? 'not_contains' : 'contains';
    return { ...base, type, pattern, flags, match, target: target(data.target) };
  }
  if (type === 'tool_used') {
    const inputMatch = str(data.input_match);
    const min = num(data.min, 1);
    const max = data.max === undefined ? undefined : num(data.max, 0);
    if (max !== undefined && max < min) throw new EvalError(`${label}: max ${max} is below min ${min} (min defaults to 1)`);
    return { ...base, type, tool: str(data.tool) ?? '', ...(inputMatch !== undefined && { inputMatch }), min, max };
  }
  if (body === '') throw new EvalError(`${label}: an llm grader needs a rubric in its body`);
  return { ...base, type, criteria: body, focus: target(data.focus), votes: num(data.votes, 3) };
}

function checkKeys(label: string, data: Record<string, Value>, allowed: string[]): void {
  const unknown = Object.keys(data).filter((k) => !allowed.includes(k));
  if (unknown.length > 0) throw new EvalError(`${label}: unknown key ${unknown.join(', ')} (allowed: ${allowed.join(', ')})`);
}

function num(v: Value | undefined, fallback: number): number {
  return typeof v === 'number' ? v : fallback;
}

function str(v: Value | undefined): string | undefined {
  return v === undefined ? undefined : String(v);
}

function target(v: Value | undefined): Target {
  return v === 'transcript' ? 'transcript' : 'last_message';
}
