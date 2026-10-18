import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { loadSuite } from '../src/suite.ts';

function suiteDir(files: Record<string, string>): string {
  const dir = mkdtempSync(join(tmpdir(), 'evals-suite-'));
  for (const [path, text] of Object.entries(files)) {
    mkdirSync(dirname(join(dir, path)), { recursive: true });
    writeFileSync(join(dir, path), text);
  }
  return dir;
}

const SUITE = JSON.stringify({ arms: [{ name: 'main', ref: 'main' }, { name: 'branch', ref: 'WORKTREE' }] });

describe('loadSuite', () => {
  it('loads arms, cases and graders with defaults', () => {
    const dir = suiteDir({
      'suite.json': SUITE,
      'cases/first/prompt.md': '---\nmax_turns: 20\nallowed_tools: [Read, Skill]\n---\nDo the thing.\n',
      'cases/first/graders/right.md': '---\ntype: llm\nweight: 2\n---\nPASS if right.\n',
      'cases/first/graders/no-edit.md': '---\ntype: tool_used\ntool: Edit\nmin: 0\nmax: 0\n---\n',
      'cases/first/graders/says.md': '---\ntype: regex\npattern: done\nflags: i\n---\n',
      'cases/first/repo/team.yaml': 'name: Example Team\n',
    });
    const suite = loadSuite(dir);
    expect(suite.arms.map((a) => a.name)).toEqual(['main', 'branch']);
    const [c] = suite.cases;
    expect(c).toMatchObject({
      name: 'first',
      prompt: 'Do the thing.',
      runs: 3,
      maxTurns: 20,
      timeoutSeconds: 600,
      allowedTools: ['Read', 'Skill'],
      repo: join(dir, 'cases/first/repo'),
    });
    expect(c!.graders).toEqual([
      { name: 'no-edit', type: 'tool_used', weight: 1, tool: 'Edit', min: 0, max: 0 },
      { name: 'right', type: 'llm', weight: 2, criteria: 'PASS if right.', focus: 'last_message', votes: 3 },
      { name: 'says', type: 'regex', weight: 1, pattern: 'done', flags: 'i', match: 'contains', target: 'last_message' },
    ]);
  });

  it('defaults to read-only tools', () => {
    const dir = suiteDir({ 'suite.json': SUITE, 'cases/a/prompt.md': 'Hi.\n', 'cases/a/graders/x.md': '---\ntype: regex\npattern: x\n---\n' });
    expect(loadSuite(dir).cases[0]!.allowedTools).toEqual(['Read', 'Glob', 'Grep', 'Skill']);
  });

  it('rejects unknown keys, unknown grader types and cases without graders', () => {
    const base = { 'suite.json': SUITE, 'cases/a/graders/x.md': '---\ntype: regex\npattern: x\n---\n' };
    expect(() => loadSuite(suiteDir({ ...base, 'cases/a/prompt.md': '---\nmax_turn: 3\n---\nHi.\n' }))).toThrow(/max_turn/);
    expect(() =>
      loadSuite(suiteDir({ ...base, 'cases/a/prompt.md': 'Hi.\n', 'cases/a/graders/x.md': '---\ntype: vibes\n---\n' })),
    ).toThrow(/vibes/);
    expect(() => loadSuite(suiteDir({ 'suite.json': SUITE, 'cases/a/prompt.md': 'Hi.\n' }))).toThrow(/no graders/);
    expect(() =>
      loadSuite(suiteDir({ ...base, 'cases/a/prompt.md': 'Hi.\n', 'cases/a/graders/x.md': '---\ntype: tool_used\ntool: Edit\nmax: 0\n---\n' })),
    ).toThrow(/below min/);
  });

  it('rejects a bad regex and an empty rubric at load time', () => {
    const p = { 'suite.json': SUITE, 'cases/a/prompt.md': 'Hi.\n' };
    expect(() => loadSuite(suiteDir({ ...p, 'cases/a/graders/x.md': '---\ntype: regex\npattern: "("\n---\n' }))).toThrow(
      /pattern/,
    );
    expect(() => loadSuite(suiteDir({ ...p, 'cases/a/graders/x.md': '---\ntype: llm\n---\n' }))).toThrow(/rubric/);
  });

  it('requires unique arm names and at least one arm', () => {
    const c = { 'cases/a/prompt.md': 'Hi.\n', 'cases/a/graders/x.md': '---\ntype: regex\npattern: x\n---\n' };
    expect(() => loadSuite(suiteDir({ ...c, 'suite.json': '{"arms": []}' }))).toThrow(/arm/);
    const dup = JSON.stringify({ arms: [{ name: 'a', ref: 'main' }, { name: 'a', ref: 'WORKTREE' }] });
    expect(() => loadSuite(suiteDir({ ...c, 'suite.json': dup }))).toThrow(/unique/);
  });
});
