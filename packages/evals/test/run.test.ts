import { mkdirSync, mkdtempSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import type { Runner } from '../src/claude.ts';
import { gradeSuite, runSuite } from '../src/run.ts';
import { loadSuite } from '../src/suite.ts';

const repoRoot = resolve(import.meta.dirname, '../../..');

function makeSuite(): string {
  const dir = mkdtempSync(join(tmpdir(), 'evals-run-'));
  const files: Record<string, string> = {
    'suite.json': JSON.stringify({ arms: [{ name: 'before', ref: 'HEAD' }, { name: 'after', ref: 'WORKTREE' }] }),
    'cases/demo/prompt.md': '---\nruns: 2\n---\nSay hello.\n',
    'cases/demo/graders/hello.md': '---\ntype: regex\npattern: hello\n---\n',
    'cases/demo/graders/polite.md': '---\ntype: llm\nvotes: 3\n---\nPASS if the answer is polite.\n',
  };
  for (const [p, t] of Object.entries(files)) {
    mkdirSync(dirname(join(dir, p)), { recursive: true });
    writeFileSync(join(dir, p), t);
  }
  return dir;
}

/** A fake claude: subjects in the "after" arm's repo say hello; the judge passes answers containing "hello". */
function fakeClaude(): { runner: Runner; judgeInputs: string[]; subjectCwds: string[]; configDirs: string[] } {
  const judgeInputs: string[] = [];
  const subjectCwds: string[] = [];
  const configDirs: string[] = [];
  const line = (o: unknown) => JSON.stringify(o);
  const runner: Runner = async (args, { cwd, configDir, input }) => {
    expect(readdirSync(configDir)).toEqual([]);
    configDirs.push(configDir);
    if (input !== undefined) {
      judgeInputs.push(input);
      const verdict = input.includes('hello') ? 'PASS' : 'FAIL';
      return { code: 0, stdout: `{"verdict":"${verdict}","reason":"r"}`, stderr: '', timedOut: false };
    }
    subjectCwds.push(cwd);
    const text = subjectCwds.length % 2 === 0 ? 'hello there' : 'go away';
    const stdout = [
      line({ type: 'system', subtype: 'init', tools: ['Read', 'Monitor'], skills: ['pr', 'review'], plugins: [], mcp_servers: [] }),
      line({ type: 'result', result: text, total_cost_usd: 0.01, num_turns: 1, is_error: false }),
    ].join('\n');
    expect(args).toContain('--setting-sources');
    return { code: 0, stdout, stderr: '', timedOut: false };
  };
  return { runner, judgeInputs, subjectCwds, configDirs };
}

describe('runSuite', () => {
  it('runs every arm x case x run in a fresh team repo, grades blind and writes a summary', async () => {
    const suite = loadSuite(makeSuite());
    const out = mkdtempSync(join(tmpdir(), 'evals-out-'));
    const fake = fakeClaude();
    const summary = await runSuite({
      suite,
      repoRoot,
      out,
      model: 'sonnet',
      judgeModel: 'haiku',
      jobs: 2,
      runner: fake.runner,
      log: () => {},
    });

    expect(fake.subjectCwds).toHaveLength(4);
    expect(new Set(fake.subjectCwds).size).toBe(4);
    const subjectConfigs = fake.configDirs.filter((d) => !d.includes('folkways-judge-config-'));
    expect(new Set(subjectConfigs).size).toBe(4); // an empty config dir per run, outside the team repo
    for (const [i, d] of subjectConfigs.entries()) expect(d.startsWith(fake.subjectCwds[i]!)).toBe(false);
    expect(fake.judgeInputs).toHaveLength(12);
    for (const input of fake.judgeInputs) expect(input).not.toMatch(/before|after|WORKTREE|HEAD/);

    expect(summary.rows.map((r) => [r.arm, r.case, r.runs])).toEqual([
      ['before', 'demo', 2],
      ['after', 'demo', 2],
    ]);
    const total = summary.rows.reduce((s, r) => s + r.score * r.runs, 0);
    expect(total).toBe(2); // two "hello" answers, each passing both graders

    const run = JSON.parse(readFileSync(join(out, 'before/demo/run-1/run.json'), 'utf8'));
    expect(run.trace.skills).toEqual(['pr', 'review']);
    expect(run.missingSkills.length).toBeGreaterThan(0); // the fake loaded 2 of the repo's skills
    expect(run.extraTools).toEqual(['Monitor']);
    expect(summary.flags).toContain('before/demo/run-1: tools beyond the case: Monitor');
    expect(readFileSync(join(out, 'summary.md'), 'utf8')).toContain('| demo |');
  });

  it('re-grades saved runs without running subjects again', async () => {
    const suite = loadSuite(makeSuite());
    const out = mkdtempSync(join(tmpdir(), 'evals-out-'));
    const first = fakeClaude();
    await runSuite({ suite, repoRoot, out, model: 'sonnet', judgeModel: 'haiku', jobs: 1, runner: first.runner, log: () => {} });

    const second = fakeClaude();
    await gradeSuite({ suite, out, judgeModel: 'haiku', jobs: 1, runner: second.runner, log: () => {} });
    expect(second.subjectCwds).toHaveLength(0);
    expect(second.judgeInputs).toHaveLength(12);
  });
});
