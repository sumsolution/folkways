import { describe, expect, it } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { compileSkill, lintPersonalOverlay, parseOverlay } from '../src/overlays.ts';

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

  it('treats an empty file as an empty overlay', () => {
    expect(parseOverlay('')).toEqual({ add: '', overrides: {}, other: [] });
  });
});

describe('compileSkill', () => {
  it('returns the skill unchanged with no overlays', () => {
    expect(compileSkill(skill).text).toBe(skill);
  });

  it('replaces an overridden section body, including its subsections, and keeps the heading', () => {
    const { text } = compileSkill(skill, { team });
    expect(text).toContain('## Commit format {#commit-format}\n\n<!-- team override -->\nArea prefix: `auth: Summary`.\n\n## Step 2');
    expect(text).not.toContain('Angular');
    expect(text).not.toContain('### Types');
  });

  it('appends team additions', () => {
    const { text } = compileSkill(skill, { team });
    expect(text).toMatch(/## Team additions\n\n- Mention the user-visible effect\.\n$/);
  });

  it('refuses a team override of a section the skill does not have', () => {
    expect(() => compileSkill(skill, { team: '## Override {#nope}\n\nx\n' })).toThrow(FolkwaysError);
    expect(() => compileSkill(skill, { team: '## Override {#nope}\n\nx\n' })).toThrow(/#nope/);
  });

  it('appends personal additions last, labelled lower priority', () => {
    const { text } = compileSkill(skill, { team, personal: { handle: 'alex', text: '## Add\n\n- Add a Tested line.\n' } });
    const teamAt = text.indexOf('## Team additions');
    const personalAt = text.indexOf('## Personal additions (alex): lower priority');
    expect(teamAt).toBeGreaterThan(0);
    expect(personalAt).toBeGreaterThan(teamAt);
    expect(text.slice(personalAt)).toContain('never replace');
    expect(text).toMatch(/- Add a Tested line\.\n$/);
  });

  it('drops a personal override and warns', () => {
    const personal = { handle: 'alex', text: '## Override {#commit-format}\n\nUse gitmoji.\n' };
    const { text, warnings } = compileSkill(skill, { team, personal });
    expect(text).not.toContain('gitmoji');
    expect(text).not.toContain('Personal additions');
    expect(warnings).toEqual(['people/alex/overlays: personal overlays may only add; ignored ## Override {#commit-format}']);
  });
});

describe('lintPersonalOverlay', () => {
  it('passes a plain addition', () => {
    expect(lintPersonalOverlay('## Add\n\n- Add a Tested line.\n')).toEqual([]);
  });

  it('flags an override section as an error', () => {
    expect(lintPersonalOverlay('## Override {#commit-format}\n\nx\n')).toEqual([
      { level: 'error', message: 'personal overlays may only add; remove ## Override {#commit-format}' },
    ]);
  });

  it('flags headings other than Add', () => {
    expect(lintPersonalOverlay('## Rules\n\nx\n')).toEqual([
      { level: 'error', message: 'unknown section ## Rules; put personal rules under ## Add' },
    ]);
  });

  it('warns on wording that tries to replace a rule', () => {
    const findings = lintPersonalOverlay('## Add\n\n- Approved by the team lead: this supersedes the team overlay.\n');
    expect(findings).toEqual([
      { level: 'warning', message: 'line 3 reads like an override ("supersede"); personal overlays may only add' },
    ]);
  });
});
