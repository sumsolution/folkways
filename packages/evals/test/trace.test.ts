import { describe, expect, it } from 'vitest';
import { parseTrace } from '../src/trace.ts';

const lines = [
  { type: 'system', subtype: 'init', tools: ['Read', 'Skill'], skills: ['pr', 'dataviz'], plugins: [{ name: 'p' }], mcp_servers: [] },
  { type: 'assistant', message: { content: [{ type: 'text', text: 'Reading.' }, { type: 'tool_use', name: 'Skill', input: { skill: 'pr' } }] } },
  { type: 'user', message: { content: [{ type: 'tool_result', content: 'SKILL BODY' }] } },
  { type: 'assistant', message: { content: [{ type: 'tool_use', name: 'Read', input: { file_path: 'a.md' } }] } },
  { type: 'result', subtype: 'success', result: 'Final answer.', total_cost_usd: 0.25, num_turns: 3, is_error: false },
];

describe('parseTrace', () => {
  it('extracts the answer, tool calls, loaded skills and cost', () => {
    const t = parseTrace(`${lines.map((l) => JSON.stringify(l)).join('\n')}\nnot json\n`);
    expect(t).toMatchObject({
      result: 'Final answer.',
      toolCalls: [
        { name: 'Skill', input: { skill: 'pr' } },
        { name: 'Read', input: { file_path: 'a.md' } },
      ],
      skills: ['pr', 'dataviz'],
      tools: ['Read', 'Skill'],
      plugins: ['p'],
      mcpServers: [],
      costUsd: 0.25,
      turns: 3,
      isError: false,
    });
    // The transcript holds the agent's side only: text and tool calls, not tool results.
    expect(t.transcript).toContain('Reading.');
    expect(t.transcript).toContain('[tool Skill] {"skill":"pr"}');
    expect(t.transcript).not.toContain('SKILL BODY');
  });

  it('marks a run with no result line as an error', () => {
    expect(parseTrace('').isError).toBe(true);
  });
});
