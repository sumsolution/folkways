import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { judgeArgs, type Runner, subjectArgs } from './claude.ts';
import { type GraderResult, gradeRegex, gradeToolUsed, judgePrompt, majority, parseVerdict, runScore, textFor, type Verdict } from './graders.ts';
import { summarize, type Summary, writeSummary } from './report.ts';
import type { Arm, Case, LlmGrader, Suite } from './suite.ts';
import { parseTrace, type Trace } from './trace.ts';
import { armPackage, buildTeamRepo, workspaceSkills } from './workspace.ts';

export interface Common {
  suite: Suite;
  /** Results go to `<out>/<arm>/<case>/run-N/`. */
  out: string;
  judgeModel: string;
  jobs: number;
  runner: Runner;
  log: (line: string) => void;
  /** Filters by name; all when omitted. */
  arms?: string[];
  cases?: string[];
}

export interface RunOptions extends Common {
  repoRoot: string;
  model: string;
  /** Overrides each case's `runs`. */
  runs?: number;
}

/** What `run.json` holds for one subject run. */
export interface SavedRun {
  arm: string;
  ref: string;
  case: string;
  run: number;
  model: string;
  code: number;
  timedOut: boolean;
  stderr: string;
  /** Skills in the team repo that the session didn't load: a broken arm, not a result. */
  missingSkills: string[];
  /** Tools the session had beyond the case's `allowed_tools`: the run wasn't held to the case. */
  extraTools: string[];
  trace: Trace;
}

/** Run every arm x case x run in a fresh team repo, then grade. */
export async function runSuite(o: RunOptions): Promise<Summary> {
  const arms = pick(o.suite.arms, o.arms);
  const cases = pick(o.suite.cases, o.cases);
  const scratch = mkdtempSync(join(tmpdir(), 'folkways-eval-'));
  const pkgs = new Map(arms.map((a) => [a.name, armPackage(o.repoRoot, a.ref, join(scratch, 'refs'))]));
  o.log(`workspaces in ${scratch}`);

  const work = arms.flatMap((arm) =>
    cases.flatMap((c) => Array.from({ length: o.runs ?? c.runs }, (_, i) => ({ arm, c, run: i + 1 }))),
  );
  await pool(work, o.jobs, async ({ arm, c, run }) => {
    const ws = join(scratch, arm.name, c.name, `run-${run}`, 'team');
    buildTeamRepo(pkgs.get(arm.name)!, ws, c.repo);
    const configDir = join(scratch, arm.name, c.name, `run-${run}`, 'claude-config');
    mkdirSync(configDir);
    const res = await o.runner(subjectArgs(c, o.model), { cwd: ws, configDir, timeoutMs: c.timeoutSeconds * 1000 });
    const trace = parseTrace(res.stdout);
    const saved: SavedRun = {
      arm: arm.name,
      ref: arm.ref,
      case: c.name,
      run,
      model: c.model ?? o.model,
      code: res.code,
      timedOut: res.timedOut,
      stderr: res.stderr.slice(-2000),
      missingSkills: workspaceSkills(ws).filter((s) => !trace.skills.includes(s)),
      extraTools: trace.tools.filter((t) => !c.allowedTools.includes(t)),
      trace,
    };
    const dir = runDir(o.out, arm.name, c.name, run);
    mkdirSync(dir, { recursive: true });
    writeFileSync(join(dir, 'answer.md'), `${trace.result.trimEnd()}\n`);
    writeFileSync(join(dir, 'run.json'), `${JSON.stringify(saved, null, 2)}\n`);
    o.log(`ran ${arm.name}/${c.name}/run-${run}: exit ${res.code}${res.timedOut ? ' (timed out)' : ''}, $${trace.costUsd.toFixed(3)}`);
  });
  return gradeSuite({ ...o, arms: arms.map((a) => a.name), cases: cases.map((c) => c.name) });
}

/**
 * Grade saved runs. llm votes go to the judge one answer at a time, shuffled across arms and runs,
 * with nothing that names the arm.
 */
export async function gradeSuite(o: Common): Promise<Summary> {
  const arms = pick(o.suite.arms, o.arms);
  const cases = pick(o.suite.cases, o.cases);
  const runs = loadRuns(o.out, arms, cases);

  const votes = new Map<string, (Verdict['verdict'] | undefined)[]>();
  const reasons = new Map<string, string[]>();
  const tasks = runs.flatMap(({ saved, c }) =>
    c.graders
      .filter((g): g is LlmGrader => g.type === 'llm')
      .flatMap((g) => Array.from({ length: g.votes }, () => ({ key: voteKey(saved, g.name), g, text: textFor(g.focus, saved.trace) }))),
  );
  shuffle(tasks);
  const judgeDir = mkdtempSync(join(tmpdir(), 'folkways-judge-'));
  const judgeConfig = mkdtempSync(join(tmpdir(), 'folkways-judge-config-'));
  let judgeCost = 0;
  await pool(tasks, o.jobs, async ({ key, g, text }) => {
    const res = await o.runner(judgeArgs(o.judgeModel), { cwd: judgeDir, configDir: judgeConfig, input: judgePrompt(g.criteria, text), timeoutMs: 180_000 });
    let reply = res.stdout;
    try {
      const json = JSON.parse(res.stdout) as { result?: string; total_cost_usd?: number };
      if (typeof json.result === 'string') reply = json.result;
      judgeCost += json.total_cost_usd ?? 0;
    } catch {
      // not the JSON envelope: read the raw reply
    }
    const v = parseVerdict(reply);
    votes.set(key, [...(votes.get(key) ?? []), v?.verdict]);
    reasons.set(key, [...(reasons.get(key) ?? []), v ? `${v.verdict}: ${v.reason}` : 'no verdict']);
  });

  const graded = runs.map(({ saved, c }) => {
    const results: GraderResult[] = c.graders.map((g) => {
      if (g.type === 'regex') return gradeRegex(g, saved.trace);
      if (g.type === 'tool_used') return gradeToolUsed(g, saved.trace);
      const key = voteKey(saved, g.name);
      const v = votes.get(key) ?? [];
      return { name: g.name, type: g.type, weight: g.weight, passed: majority(v), detail: (reasons.get(key) ?? []).join(' | '), votes: v };
    });
    const score = runScore(results);
    writeFileSync(join(runDir(o.out, saved.arm, saved.case, saved.run), 'grades.json'), `${JSON.stringify({ score, results }, null, 2)}\n`);
    return { saved, results, score };
  });

  const summary = summarize(graded, arms.map((a) => a.name), cases.map((c) => c.name), judgeCost);
  writeSummary(o.out, summary, { suite: o.suite.name, judgeModel: o.judgeModel });
  o.log(`graded ${graded.length} runs, ${tasks.length} judge votes`);
  return summary;
}

function loadRuns(out: string, arms: Arm[], cases: Case[]): { saved: SavedRun; c: Case }[] {
  const found: { saved: SavedRun; c: Case }[] = [];
  for (const arm of arms) {
    for (const c of cases) {
      for (let run = 1; ; run++) {
        let text: string;
        try {
          text = readFileSync(join(runDir(out, arm.name, c.name, run), 'run.json'), 'utf8');
        } catch {
          break;
        }
        found.push({ saved: JSON.parse(text) as SavedRun, c });
      }
    }
  }
  return found;
}

function runDir(out: string, arm: string, c: string, run: number): string {
  return join(out, arm, c, `run-${run}`);
}

function voteKey(s: SavedRun, grader: string): string {
  return `${s.arm}\0${s.case}\0${s.run}\0${grader}`;
}

function pick<T extends { name: string }>(all: T[], names: string[] | undefined): T[] {
  return names ? all.filter((x) => names.includes(x.name)) : all;
}

function shuffle<T>(xs: T[]): void {
  for (let i = xs.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [xs[i], xs[j]] = [xs[j]!, xs[i]!];
  }
}

async function pool<T>(items: T[], jobs: number, fn: (item: T) => Promise<void>): Promise<void> {
  let next = 0;
  const worker = async (): Promise<void> => {
    while (next < items.length) await fn(items[next++]!);
  };
  await Promise.all(Array.from({ length: Math.max(1, jobs) }, worker));
}
