import type { RegexGrader, Target, ToolUsedGrader } from './suite.ts';
import type { Trace } from './trace.ts';

export interface GraderResult {
  name: string;
  type: string;
  weight: number;
  passed: boolean;
  detail: string;
  /** llm graders: each vote, `undefined` when the judge's reply had no verdict. */
  votes?: (Verdict['verdict'] | undefined)[];
}

export interface Verdict {
  verdict: 'PASS' | 'FAIL';
  reason: string;
}

export function textFor(target: Target, trace: Trace): string {
  return target === 'transcript' ? trace.transcript : trace.result;
}

export function gradeRegex(g: RegexGrader, trace: Trace): GraderResult {
  const found = new RegExp(g.pattern, g.flags).test(textFor(g.target, trace));
  const passed = g.match === 'contains' ? found : !found;
  return { name: g.name, type: g.type, weight: g.weight, passed, detail: `${g.match} /${g.pattern}/${g.flags}: ${found ? 'found' : 'not found'}` };
}

export function gradeToolUsed(g: ToolUsedGrader, trace: Trace): GraderResult {
  const re = g.inputMatch === undefined ? undefined : new RegExp(g.inputMatch);
  const count = trace.toolCalls.filter((c) => c.name === g.tool && (!re || re.test(JSON.stringify(c.input)))).length;
  const passed = count >= g.min && (g.max === undefined || count <= g.max);
  return { name: g.name, type: g.type, weight: g.weight, passed, detail: `${count} matching ${g.tool} call(s)` };
}

/**
 * The judge sees the rubric and one answer, nothing else: not the arm, not the other answers. Keep
 * words that could name an arm out of this template; a test checks it.
 */
export function judgePrompt(criteria: string, text: string): string {
  return `You are grading one response from an AI coding agent against a rubric. Judge only what the response says and does.

<rubric>
${criteria}
</rubric>

<response>
${text}
</response>

Think briefly, then end your reply with this one line of JSON:
{"verdict": "PASS" or "FAIL", "reason": "<one sentence>"}`;
}

/** The last JSON object in the judge's reply that carries a verdict. */
export function parseVerdict(out: string): Verdict | undefined {
  const objects = out.match(/\{[^{}]*"verdict"[^{}]*\}/g) ?? [];
  for (const raw of objects.reverse()) {
    try {
      const v = JSON.parse(raw) as { verdict?: unknown; reason?: unknown };
      if (v.verdict === 'PASS' || v.verdict === 'FAIL') return { verdict: v.verdict, reason: String(v.reason ?? '') };
    } catch {
      // not JSON after all; try the previous one
    }
  }
  return undefined;
}

/** Passes on a strict majority of all votes; a vote without a verdict counts against. */
export function majority(votes: (Verdict['verdict'] | undefined)[]): boolean {
  return votes.filter((v) => v === 'PASS').length * 2 > votes.length;
}

/** Weighted share of graders that passed. */
export function runScore(results: GraderResult[]): number {
  const total = results.reduce((s, r) => s + r.weight, 0);
  return total === 0 ? 0 : results.reduce((s, r) => s + (r.passed ? r.weight : 0), 0) / total;
}

