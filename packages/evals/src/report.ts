import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import type { GraderResult } from './graders.ts';
import type { SavedRun } from './run.ts';

export interface Row {
  arm: string;
  case: string;
  runs: number;
  /** Mean run score, 0 to 1. */
  score: number;
  /** Runs where every grader passed. */
  perfect: number;
  /** Passes per grader, in grader order. */
  graders: { name: string; passed: number }[];
  costUsd: number;
}

export interface Summary {
  rows: Row[];
  /** Runs whose result can't be trusted as is: errors, timeouts, skills that didn't load, MCP servers. */
  flags: string[];
  judgeCostUsd: number;
}

interface Graded {
  saved: SavedRun;
  results: GraderResult[];
  score: number;
}

export function summarize(graded: Graded[], arms: string[], cases: string[], judgeCostUsd: number): Summary {
  const rows: Row[] = [];
  for (const arm of arms) {
    for (const c of cases) {
      const g = graded.filter((x) => x.saved.arm === arm && x.saved.case === c);
      if (g.length === 0) continue;
      rows.push({
        arm,
        case: c,
        runs: g.length,
        score: g.reduce((s, x) => s + x.score, 0) / g.length,
        perfect: g.filter((x) => x.results.every((r) => r.passed)).length,
        graders: g[0]!.results.map((r, i) => ({ name: r.name, passed: g.filter((x) => x.results[i]?.passed).length })),
        costUsd: g.reduce((s, x) => s + x.saved.trace.costUsd, 0),
      });
    }
  }
  const flags = graded.flatMap(({ saved: s }) => {
    const id = `${s.arm}/${s.case}/run-${s.run}`;
    const out: string[] = [];
    if (s.timedOut) out.push(`${id}: timed out`);
    else if (s.code !== 0 || s.trace.isError) out.push(`${id}: exit ${s.code}${s.trace.isError ? ', error result' : ''}`);
    if (s.missingSkills.length > 0) out.push(`${id}: skills not loaded: ${s.missingSkills.join(', ')}`);
    if (s.extraTools?.length) out.push(`${id}: tools beyond the case: ${s.extraTools.join(', ')}`);
    if (s.trace.mcpServers.length > 0) out.push(`${id}: MCP servers loaded: ${s.trace.mcpServers.join(', ')}`);
    return out;
  });
  return { rows, flags, judgeCostUsd };
}

export function writeSummary(out: string, s: Summary, meta: { suite: string; judgeModel: string }): void {
  writeFileSync(join(out, 'summary.json'), `${JSON.stringify({ ...meta, ...s }, null, 2)}\n`);
  writeFileSync(join(out, 'summary.md'), toMarkdown(s, meta));
}

function toMarkdown(s: Summary, meta: { suite: string; judgeModel: string }): string {
  const lines = [`# ${meta.suite}: results`, '', `Judge: ${meta.judgeModel}. Scores are mean weighted grader passes per run.`, ''];
  for (const c of [...new Set(s.rows.map((r) => r.case))]) {
    const rows = s.rows.filter((r) => r.case === c);
    const first = rows[0]!;
    const names = first.graders.map((g) => g.name);
    lines.push(`## ${c}`, '');
    lines.push(`| case | arm | runs | score | Δ vs ${first.arm} | perfect | ${names.join(' | ')} | cost |`);
    lines.push(`|---|---|---|---|---|---|${names.map(() => '---|').join('')}---|`);
    for (const r of rows) {
      const delta = r === first ? '' : signed(r.score - first.score);
      const cells = r.graders.map((g) => `${g.passed}/${r.runs}`);
      lines.push(`| ${c} | ${r.arm} | ${r.runs} | ${r.score.toFixed(2)} | ${delta} | ${r.perfect}/${r.runs} | ${cells.join(' | ')} | $${r.costUsd.toFixed(2)} |`);
    }
    lines.push('');
  }
  lines.push(`Judge cost: $${s.judgeCostUsd.toFixed(2)}.`, '');
  if (s.flags.length > 0) lines.push('## Flags', '', ...s.flags.map((f) => `- ${f}`), '');
  return lines.join('\n');
}

function signed(n: number): string {
  return `${n >= 0 ? '+' : ''}${n.toFixed(2)}`;
}
