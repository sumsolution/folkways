import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { read } from './util.ts';
import { buildEnv, check, fill, merge, mismatch, transcriptJsonl, type Vars } from './harness.ts';

describe('mismatch', () => {
  it('matches a deep subset', () => {
    expect(mismatch({ a: { b: 1 } }, { a: { b: 1, c: 2 }, d: 3 })).toBeUndefined();
    expect(mismatch({ a: { b: 1 } }, { a: { b: 2 } })).toMatch(/\$\.a\.b/);
  });
  it('treats re: strings as case-insensitive regexes', () => {
    expect(mismatch({ r: 're:scope.*api' }, { r: 'Scope "API" is not allowed' })).toBeUndefined();
    expect(mismatch({ r: 're:scope' }, { r: 'nothing' })).toBeDefined();
    expect(mismatch({ r: 're:scope' }, {})).toBeDefined();
  });
});

describe('fill and merge', () => {
  it('replaces placeholders in keys and values', () => {
    const vars = { TEAM: '/t', HEAD7: 'abc1234' } as Vars;
    expect(fill({ '${TEAM}/x': ['${HEAD7}', 1] }, vars)).toEqual({ '/t/x': ['abc1234', 1] });
  });
  it('merges objects and replaces arrays', () => {
    expect(merge({ git: { scopes: ['a'], merge: 'squash' } }, { git: { scopes: [] } })).toEqual({ git: { scopes: [], merge: 'squash' } });
  });
});

describe('buildEnv', () => {
  it('builds a team repo, a main clone and a worktree', () => {
    const vars = buildEnv({ description: '', env: { team: { git: { scopes: [] } }, files: { 'efforts/x/r-${HEAD7}.md': 'sha: ${HEAD}\n' } }, steps: [] });
    expect(read(join(vars.TEAM, 'team.yaml'))).toMatch(/scopes":\[\]/);
    expect(read(join(vars.TEAM, '.local/me.yaml'))).toContain(`projects_dir: ${JSON.stringify(vars.PROJECTS)}`);
    expect(read(join(vars.TEAM, `efforts/x/r-${vars.HEAD7}.md`))).toBe(`sha: ${vars.HEAD}\n`);
    expect(read(join(vars.WORKTREE, '.git'))).toMatch(/^gitdir:/);
    expect(vars.HEAD).not.toBe(vars.OLD_HEAD);
    expect(vars.OLD_HEAD7).toBe(vars.OLD_HEAD.slice(0, 7));
  });
  it('omits me.yaml when me is null', () => {
    const vars = buildEnv({ description: '', env: { me: null }, steps: [] });
    expect(() => read(join(vars.TEAM, '.local/me.yaml'))).toThrow();
  });
});

describe('transcriptJsonl', () => {
  it('writes main-chain and sidechain assistant lines whose usage sums to the token count', () => {
    const [main, side] = transcriptJsonl([130_000, { sidechain: 5 }]).trim().split('\n').map((l) => JSON.parse(l));
    const u = main.message.usage;
    expect(u.input_tokens + u.cache_read_input_tokens + u.cache_creation_input_tokens).toBe(130_000);
    expect(main.isSidechain).toBe(false);
    expect(side.isSidechain).toBe(true);
  });
});

describe('check', () => {
  const vars = { TEAM: '/nonexistent' } as Vars;
  it('passes empty stdout when null is expected', () => {
    expect(check({ stdout: null }, { exit: 0, stdout: '\n', stderr: '' }, vars)).toEqual([]);
  });
  it('reports exit code, stdout and file problems', () => {
    const problems = check({ exit: 0, stdout: { a: 1 }, exists: ['${TEAM}/x'] }, { exit: 2, stdout: 'nope', stderr: 'boom' }, vars);
    expect(problems).toHaveLength(3);
    expect(problems[0]).toMatch(/exit 2/);
    expect(problems[2]).toContain('/nonexistent/x');
  });
});
