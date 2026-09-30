#!/usr/bin/env node
import { resolve } from 'node:path';
import { parseArgs } from 'node:util';
import { spawnClaude } from './claude.ts';
import { EvalError } from './errors.ts';
import { gradeSuite, runSuite } from './run.ts';
import { loadSuite } from './suite.ts';

const USAGE = `usage: folkways-eval <command> <suite-dir> [options]

  run <suite-dir>     build a team repo per arm x case x run, run claude -p in it, grade, summarize
  grade <suite-dir>   re-grade saved runs (after a rubric change); runs no subjects

  --model <m>         subject model (default: suite.json model, else sonnet)
  --judge-model <m>   judge model for llm graders (default: suite.json judge_model, else sonnet)
  --runs <n>          runs per case per arm (default: each case's runs, else 3)
  --arms a,b          only these arms          --cases x,y   only these cases
  --jobs <n>          parallel claude processes (default 4)
  --out <dir>         results dir (default: <suite-dir>/results/<model>)
  --dry-run           print the plan and exit

Every run and judge vote is a model call on your account. See packages/evals/README.md.`;

async function main(argv: string[]): Promise<void> {
  const { positionals, values } = parseArgs({
    args: argv,
    allowPositionals: true,
    options: {
      model: { type: 'string' },
      'judge-model': { type: 'string' },
      runs: { type: 'string' },
      arms: { type: 'string' },
      cases: { type: 'string' },
      jobs: { type: 'string', default: '4' },
      out: { type: 'string' },
      'dry-run': { type: 'boolean' },
      help: { type: 'boolean', short: 'h' },
    },
  });
  const [command, dir] = positionals;
  if (values.help || !command) return console.log(USAGE);
  if ((command !== 'run' && command !== 'grade') || !dir) throw new EvalError(USAGE);

  const suite = loadSuite(resolve(dir));
  const model = values.model ?? suite.model ?? 'sonnet';
  const judgeModel = values['judge-model'] ?? suite.judgeModel ?? 'sonnet';
  const common = {
    suite,
    out: resolve(values.out ?? `${suite.dir}/results/${model}`),
    judgeModel,
    jobs: Number(values.jobs),
    runner: spawnClaude,
    log: (line: string) => console.log(line),
    ...(values.arms && { arms: values.arms.split(',') }),
    ...(values.cases && { cases: values.cases.split(',') }),
  };
  for (const name of common.arms ?? []) {
    if (!suite.arms.some((a) => a.name === name)) throw new EvalError(`no arm ${name} in suite.json`);
  }
  for (const name of common.cases ?? []) {
    if (!suite.cases.some((c) => c.name === name)) throw new EvalError(`no case ${name}`);
  }

  const arms = suite.arms.filter((a) => !common.arms || common.arms.includes(a.name));
  const cases = suite.cases.filter((c) => !common.cases || common.cases.includes(c.name));
  const runs = values.runs === undefined ? undefined : Number(values.runs);
  const subjects = arms.length * cases.reduce((s, c) => s + (runs ?? c.runs), 0);
  const votes = arms.length * cases.reduce((s, c) => s + (runs ?? c.runs) * c.graders.reduce((v, g) => v + (g.type === 'llm' ? g.votes : 0), 0), 0);
  console.log(
    `${suite.name}: ${arms.map((a) => `${a.name}=${a.ref}`).join(', ')}; ${cases.length} case(s); ` +
      `${command === 'run' ? `${subjects} subject runs (${model}), ` : ''}${votes} judge votes (${judgeModel}) -> ${common.out}`,
  );
  if (values['dry-run']) return;

  const summary =
    command === 'run'
      ? await runSuite({ ...common, repoRoot: resolve(import.meta.dirname, '../../..'), model, ...(runs !== undefined && { runs }) })
      : await gradeSuite(common);
  for (const r of summary.rows) console.log(`${r.case}  ${r.arm}  ${r.score.toFixed(2)}  (${r.perfect}/${r.runs} perfect)`);
  for (const f of summary.flags) console.log(`flag: ${f}`);
  console.log(`summary: ${common.out}/summary.md`);
}

try {
  await main(process.argv.slice(2));
} catch (e) {
  if (e instanceof EvalError || (e as { code?: string }).code?.startsWith('ERR_PARSE_ARGS')) {
    console.error(`folkways-eval: ${(e as Error).message}`);
    process.exit(1);
  }
  throw e;
}
