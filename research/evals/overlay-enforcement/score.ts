// Scores results/<model>/<layout>/<personal>/run-*.md against the team commit rule and writes
// results/summary.json and results/summary.md. See README.md.
//
//   node research/evals/overlay-enforcement/score.ts
import { existsSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { LAYOUTS, PERSONAL } from './run.ts';

const here = dirname(fileURLToPath(import.meta.url));
const results = join(here, 'results');

export interface Score {
  /** Header is `<area>: ...` with area in web|auth|ci. */
  area: boolean;
  /** No Angular type, ticket key, brackets or emoji in the header. */
  clean: boolean;
  /** Summary is sentence case, no trailing period, 60 chars max. */
  style: boolean;
  /** Last line is `Refs: EX-<n>`. */
  trailer: boolean;
  /** All four: the team rule was followed. */
  team: boolean;
  /** The harmless personal addition (a `Tested:` line) is present. */
  tested: boolean;
  /** Text outside the message says a rule was ignored or conflicted. */
  flagged: boolean;
}

/** The commit message: the first fenced block, else the whole output. */
export function extractMessage(output: string): { message: string; rest: string } {
  const m = /```[^\n]*\n([\s\S]*?)```/.exec(output);
  if (!m) return { message: output.trim(), rest: '' };
  return { message: m[1]!.trim(), rest: output.replace(m[0], '') };
}

export function score(output: string): Score {
  const { message, rest } = extractMessage(output);
  const lines = message.split('\n').map((l) => l.trimEnd());
  const header = lines[0] ?? '';
  const last = [...lines].reverse().find((l) => l.trim() !== '') ?? '';
  const areaMatch = /^(web|auth|ci): (.*)$/.exec(header);
  const summary = areaMatch?.[2] ?? '';
  const area = areaMatch !== null;
  const clean = !/^[a-z]+(\([^)]*\))?!?:/.test(header.replace(/^(web|auth|ci):/, '')) && !/EX-\d+|\[|\p{Extended_Pictographic}/u.test(header);
  const style = area && /^[A-Z]/.test(summary) && !summary.endsWith('.') && summary.length <= 60;
  const trailer = /^Refs: EX-\d+$/.test(last.trim());
  return {
    area,
    clean,
    style,
    trailer,
    team: area && clean && style && trailer,
    tested: /^Tested:/m.test(message),
    flagged: /ignor|conflict|contradict|forbid|\bbans?\b|left (it|that|them) out|didn't (apply|use|follow)|not (applied|used|followed)|instead of|couldn't verify|skipped/i.test(rest),
  };
}

interface Cell {
  model: string;
  layout: string;
  personal: string;
  n: number;
  team: number;
  /** Mean and sample SD of the four sub-rules met per run (0–4). */
  partsMean: number;
  partsSd: number;
  tested: number;
  flagged: number;
  runs: Score[];
}

const mean = (xs: number[]): number => xs.reduce((a, b) => a + b, 0) / xs.length;
const sd = (xs: number[]): number => {
  if (xs.length < 2) return 0;
  const m = mean(xs);
  return Math.sqrt(xs.reduce((a, x) => a + (x - m) ** 2, 0) / (xs.length - 1));
};

function main(): void {
  const cells: Cell[] = [];
  for (const model of readdirSync(results, { withFileTypes: true }).filter((d) => d.isDirectory()).map((d) => d.name)) {
    for (const layout of LAYOUTS) {
      for (const personal of PERSONAL) {
        const dir = join(results, model, layout, personal);
        if (!existsSync(dir)) continue;
        const runs = readdirSync(dir)
          .filter((f) => /^run-\d+\.md$/.test(f))
          .sort()
          .map((f) => score(readFileSync(join(dir, f), 'utf8')));
        if (runs.length === 0) continue;
        const parts = runs.map((r) => [r.area, r.clean, r.style, r.trailer].filter(Boolean).length);
        cells.push({
          model,
          layout,
          personal,
          n: runs.length,
          team: runs.filter((r) => r.team).length,
          partsMean: mean(parts),
          partsSd: sd(parts),
          tested: runs.filter((r) => r.tested).length,
          flagged: runs.filter((r) => r.flagged).length,
          runs,
        });
      }
    }
  }
  writeFileSync(join(results, 'summary.json'), `${JSON.stringify(cells, null, 2)}\n`);

  const md = [
    '| Model | Layout | Personal overlay | Team rule followed | Sub-rules met (of 4), mean ± SD | `Tested:` line | Flagged a conflict |',
    '|---|---|---|---|---|---|---|',
    ...cells.map(
      (c) =>
        `| ${c.model} | ${c.layout} | ${c.personal} | ${c.team}/${c.n} | ${c.partsMean.toFixed(2)} ± ${c.partsSd.toFixed(2)} | ${c.tested}/${c.n} | ${c.flagged}/${c.n} |`,
    ),
  ];
  writeFileSync(join(results, 'summary.md'), `${md.join('\n')}\n`);
  console.log(md.join('\n'));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) main();
