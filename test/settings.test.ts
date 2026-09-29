import { describe, expect, it } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { FRAMEWORK_HOOK_PREFIX, isFrameworkHook, mergeSettings } from '../src/settings.ts';

const ours = (name: string) => ({ type: 'command', command: `node_modules/.bin/folkways hook ${name}` });
const theirs = { type: 'command', command: 'scripts/team.sh' };

describe('mergeSettings', () => {
  it('adds framework hooks and ask rules to empty settings', () => {
    const out = mergeSettings({}, { permissions: { ask: ['Bash(git push:*)'] }, hooks: { Stop: [{ hooks: [ours('a')] }] } });
    expect(out).toEqual({ permissions: { ask: ['Bash(git push:*)'] }, hooks: { Stop: [{ hooks: [ours('a')] }] } });
  });

  it('replaces framework hooks and keeps team hooks, even in a shared group', () => {
    const current = {
      hooks: {
        PreToolUse: [{ matcher: 'Bash', hooks: [theirs, ours('old')] }],
        SessionStart: [{ hooks: [ours('old')] }],
      },
    };
    const out = mergeSettings(current, { hooks: { PreToolUse: [{ matcher: 'Bash', hooks: [ours('new')] }] } });
    expect(out.hooks).toEqual({
      PreToolUse: [
        { matcher: 'Bash', hooks: [theirs] },
        { matcher: 'Bash', hooks: [ours('new')] },
      ],
    });
  });

  it('drops the hooks key when nothing is left', () => {
    expect(mergeSettings({ hooks: { Stop: [{ hooks: [ours('x')] }] } }, {})).toEqual({});
  });

  it('keeps other keys and does not duplicate ask rules', () => {
    const current = { model: 'opus', permissions: { allow: ['Bash(ls:*)'], ask: ['Bash(git push:*)', 'Bash(rm:*)'] } };
    const out = mergeSettings(current, { permissions: { ask: ['Bash(git push:*)'] } });
    expect(out).toEqual(current);
  });

  it('does not mutate its input', () => {
    const current = { hooks: { Stop: [{ hooks: [ours('x'), theirs] }] } };
    mergeSettings(current, {});
    expect(current.hooks.Stop[0]!.hooks).toHaveLength(2);
  });
});

describe('isFrameworkHook', () => {
  it('owns commands that start with the framework prefix', () => {
    expect(FRAMEWORK_HOOK_PREFIX).toBe('node_modules/.bin/folkways hook ');
    expect(isFrameworkHook('node_modules/.bin/folkways hook session-start')).toBe(true);
  });

  it.each([
    'scripts/run-folkways hook x',
    'echo folkways hook x',
    'npx folkways hook x',
    'scripts/team.sh && node_modules/.bin/folkways hook x',
    'node_modules/.bin/folkways hooks x',
    '',
  ])('does not own %j', (command) => expect(isFrameworkHook(command)).toBe(false));

  it('keeps team hooks that mention folkways hook', () => {
    const team = { type: 'command', command: 'scripts/wrap.sh folkways hook x' };
    expect(mergeSettings({ hooks: { Stop: [{ hooks: [team] }] } }, {}).hooks).toEqual({ Stop: [{ hooks: [team] }] });
  });
});

describe('mergeSettings with unexpected shapes', () => {
  it('keeps a hook group without a hooks array untouched', () => {
    const current = { hooks: { Stop: [{ matcher: 'x' }, { hooks: [ours('old'), theirs] }] } } as any;
    const out = mergeSettings(current, { hooks: { Stop: [{ hooks: [ours('new')] }] } });
    expect(out.hooks).toEqual({ Stop: [{ matcher: 'x' }, { hooks: [theirs] }, { hooks: [ours('new')] }] });
  });

  it('keeps non-array events, non-object groups and non-string commands untouched', () => {
    const current = {
      hooks: { Stop: { hooks: [] }, PreToolUse: [null, 'x', { hooks: [null, { command: 1 }, ours('old')] }] },
    } as any;
    expect(mergeSettings(current, {}).hooks).toEqual({
      Stop: { hooks: [] },
      PreToolUse: [null, 'x', { hooks: [null, { command: 1 }] }],
    });
  });

  it.each([
    [{ hooks: [] }, /hooks must be a JSON object/],
    [{ permissions: [] }, /permissions must be a JSON object/],
    [{ permissions: { ask: 'Bash(rm:*)' } }, /permissions\.ask must be an array/],
    [{ hooks: { Stop: {} } }, /hooks\.Stop must be an array/],
    [null, /must be a JSON object/],
  ])('refuses what it cannot merge: %j', (current, message) => {
    const fragment = { permissions: { ask: ['Bash(git push:*)'] }, hooks: { Stop: [{ hooks: [ours('a')] }] } };
    expect(() => mergeSettings(current as any, fragment)).toThrow(FolkwaysError);
    expect(() => mergeSettings(current as any, fragment)).toThrow(message);
  });
});
