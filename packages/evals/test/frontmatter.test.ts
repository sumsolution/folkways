import { describe, expect, it } from 'vitest';
import { parseDoc } from '../src/frontmatter.ts';

describe('parseDoc', () => {
  it('splits frontmatter from the body', () => {
    const doc = parseDoc('---\ntype: llm\nweight: 2\n---\n\nPASS if it works.\n');
    expect(doc.data).toEqual({ type: 'llm', weight: 2 });
    expect(doc.body).toBe('PASS if it works.');
  });

  it('parses flow lists, booleans and quoted strings', () => {
    const doc = parseDoc(
      "---\nallowed_tools: [Read, Glob, 'Skill']\narm: both\nok: true\ninput_match: '\"skill\"\\s*:\\s*\"pr\"'\ntitle: \"a: b\"\n---\n",
    );
    expect(doc.data).toEqual({
      allowed_tools: ['Read', 'Glob', 'Skill'],
      arm: 'both',
      ok: true,
      input_match: '"skill"\\s*:\\s*"pr"',
      title: 'a: b',
    });
  });

  it('returns the whole text as the body without frontmatter', () => {
    expect(parseDoc('just a prompt\n')).toEqual({ data: {}, body: 'just a prompt' });
  });

  it('rejects lines it cannot parse', () => {
    expect(() => parseDoc('---\nnot a pair\n---\n')).toThrow(/frontmatter line/);
  });
});
