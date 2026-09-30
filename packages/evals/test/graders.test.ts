import { describe, expect, it } from 'vitest';
import { gradeRegex, gradeToolUsed, judgePrompt, majority, parseVerdict, runScore } from '../src/graders.ts';
import type { Trace } from '../src/trace.ts';

const trace: Trace = {
  result: 'Done. Refs: EX-1',
  transcript: 'Looking.\n[tool Skill] {"skill":"pr"}',
  toolCalls: [
    { name: 'Skill', input: { skill: 'pr' } },
    { name: 'Read', input: { file_path: 'a.md' } },
    { name: 'Read', input: { file_path: 'b.md' } },
  ],
  skills: [],
  tools: [],
  plugins: [],
  mcpServers: [],
  costUsd: 0,
  turns: 1,
  isError: false,
};

describe('deterministic graders', () => {
  it('regex matches the answer or the transcript, and can require absence', () => {
    const g = { name: 'r', weight: 1, type: 'regex', flags: 'i', match: 'contains', target: 'last_message' } as const;
    expect(gradeRegex({ ...g, pattern: 'refs: ex-\\d+' }, trace).passed).toBe(true);
    expect(gradeRegex({ ...g, pattern: 'skill', target: 'transcript' }, trace).passed).toBe(true);
    expect(gradeRegex({ ...g, pattern: 'skill' }, trace).passed).toBe(false);
    expect(gradeRegex({ ...g, pattern: 'skill', match: 'not_contains' }, trace).passed).toBe(true);
  });

  it('tool_used counts matching calls between min and max', () => {
    const g = { name: 't', weight: 1, type: 'tool_used', min: 1, max: undefined } as const;
    expect(gradeToolUsed({ ...g, tool: 'Read' }, trace).passed).toBe(true);
    expect(gradeToolUsed({ ...g, tool: 'Read', max: 1 }, trace).passed).toBe(false);
    expect(gradeToolUsed({ ...g, tool: 'Skill', inputMatch: '"skill":"pr"' }, trace).passed).toBe(true);
    expect(gradeToolUsed({ ...g, tool: 'Edit', min: 0, max: 0 }, trace).passed).toBe(true);
  });
});

describe('judge', () => {
  it('shows the rubric and the answer, and nothing about the arm', () => {
    const p = judgePrompt('PASS if it declines.', 'I decline.');
    expect(p).toContain('PASS if it declines.');
    expect(p).toContain('I decline.');
    expect(p).not.toMatch(/\barm\b|baseline|WORKTREE|main/i);
  });

  it('parses the last verdict line and ignores prose around it', () => {
    expect(parseVerdict('Thinking...\n{"verdict":"PASS","reason":"ok"}\n')).toEqual({ verdict: 'PASS', reason: 'ok' });
    expect(parseVerdict('```json\n{"verdict": "FAIL", "reason": "no"}\n```')).toEqual({ verdict: 'FAIL', reason: 'no' });
    expect(parseVerdict('PASS')).toBeUndefined();
  });

  it('passes on a strict majority of valid votes', () => {
    expect(majority(['PASS', 'PASS', 'FAIL'])).toBe(true);
    expect(majority(['PASS', 'FAIL', undefined])).toBe(false);
    expect(majority([undefined, undefined, undefined])).toBe(false);
  });
});

describe('runScore', () => {
  it('is the weighted share of passing graders', () => {
    const r = (passed: boolean, weight: number) => ({ name: 'x', type: 'regex', weight, passed, detail: '' });
    expect(runScore([r(true, 2), r(false, 1), r(true, 1)])).toBe(0.75);
    expect(runScore([])).toBe(0);
  });
});
