import { describe, expect, it } from 'vitest';
import { compileSkillFiles, parseOverlay } from '../src/overlays.ts';

const skill = `---
name: commit
description: Write a commit message.
---

# Commit

## Step 1: read the change

Read the diff.

## Commit format {#commit-format}

Angular: \`type(scope): subject\`.

### Types

feat, fix.

## Step 2: write the message

Write it.
`;

const team = `# Team overlay: commit

## Add

- Mention the user-visible effect.

## Override {#commit-format}

Area prefix: \`auth: Summary\`.
`;

describe('parseOverlay', () => {
  it('splits Add and Override sections', () => {
    expect(parseOverlay(team)).toEqual({
      add: '- Mention the user-visible effect.',
      overrides: { 'commit-format': 'Area prefix: `auth: Summary`.' },
      other: [],
    });
  });

  it('reports headings that are neither Add nor Override', () => {
    expect(parseOverlay('## Add\n\nx\n\n## Notes\n\ny\n').other).toEqual(['Notes']);
  });

  it('joins repeated Add sections', () => {
    expect(parseOverlay('## Add\n\n- one\n\n## Add\n\n- two\n').add).toBe('- one\n\n- two');
  });

  it('refuses the same override ID twice', () => {
    expect(() => parseOverlay('## Override {#x}\n\na\n\n## Override {#x}\n\nb\n')).toThrow(/#x.*twice/);
  });

  it('ignores headings inside a fence, matching fence character and length', () => {
    const text = '## Add\n\n````md\n```\n## Inner\n```\n````\n\n~~~\n```\n## Also inner\n~~~\n';
    expect(parseOverlay(text).other).toEqual([]);
  });

  it('treats an empty file as an empty overlay', () => {
    expect(parseOverlay('')).toEqual({ add: '', overrides: {}, other: [] });
  });
});

describe('compileSkillFiles', () => {
  const template = '# Templates\n\n## Story {#story-template}\n\nAs a user.\n';
  const files = { 'SKILL.md': skill, 'references/templates.md': template, 'scripts/run.sh': '#!/bin/sh\n' };

  it('returns nothing changed with an empty overlay', () => {
    expect(compileSkillFiles(files, '')).toEqual({ files: {}, warnings: [] });
  });

  it('overrides a section in SKILL.md and appends team additions there', () => {
    const out = compileSkillFiles(files, team).files;
    expect(Object.keys(out)).toEqual(['SKILL.md']);
    expect(out['SKILL.md']).toContain('## Commit format {#commit-format}\n\n<!-- team override -->\nArea prefix: `auth: Summary`.\n\n## Step 2');
    expect(out['SKILL.md']).not.toContain('Angular');
    expect(out['SKILL.md']).not.toContain('### Types');
    expect(out['SKILL.md']).toMatch(/## Team additions\n\n- Mention the user-visible effect\.\n$/);
  });

  it('overrides a section in a reference file', () => {
    const out = compileSkillFiles(files, '## Override {#story-template}\n\nGiven, when, then.\n').files;
    expect(Object.keys(out)).toEqual(['references/templates.md']);
    expect(out['references/templates.md']).toContain('<!-- team override -->\nGiven, when, then.');
    expect(out['references/templates.md']).not.toContain('As a user.');
  });

  it('refuses an ID no file has, and an ID in two files', () => {
    expect(() => compileSkillFiles(files, '## Override {#nope}\n\nx\n')).toThrow(/#nope/);
    const twice = { ...files, 'references/other.md': '## Again {#commit-format}\n\ny\n' };
    expect(() => compileSkillFiles(twice, '## Override {#commit-format}\n\nx\n')).toThrow(
      /#commit-format.*SKILL\.md.*references\/other\.md/,
    );
  });

  it('refuses a section ID the skill uses twice in one file', () => {
    const dup = { 'SKILL.md': '## A {#x}\n\none\n\n## B {#x}\n\ntwo\n' };
    expect(() => compileSkillFiles(dup, '## Override {#x}\n\nnew\n')).toThrow(/#x.*more than one/);
  });

  it('says when an override targets a section a parent override already replaced', () => {
    const nested = { 'SKILL.md': '## A {#a}\n\none\n\n### B {#b}\n\ntwo\n' };
    expect(() => compileSkillFiles(nested, '## Override {#a}\n\nnew\n\n## Override {#b}\n\nx\n')).toThrow(
      /#b sits inside #a/,
    );
    expect(() => compileSkillFiles(nested, '## Override {#b}\n\nx\n\n## Override {#a}\n\nnew\n')).toThrow(
      /#b sits inside #a/,
    );
  });

  it('ignores IDs and headings in non-markdown files', () => {
    const shell = { ...files, 'scripts/run.sh': '## Fake {#fake}\n' };
    expect(() => compileSkillFiles(shell, '## Override {#fake}\n\nx\n')).toThrow(/#fake/);
  });

  it('warns about sections it does not apply', () => {
    expect(compileSkillFiles(files, '## Notes\n\nx\n').warnings).toEqual([
      'overlays: ignored ## Notes; team overlays hold ## Add and ## Override {#id} only',
    ]);
  });
});
