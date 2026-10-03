import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { FolkwaysError } from '../src/errors.ts';
import { FRAMEWORK_HOOK_PREFIX, isFrameworkHook, mergeSettings } from '../src/settings.ts';
import { ASSETS, json } from './helpers.ts';

const ours = (name: string) => ({ type: 'command', command: `${FRAMEWORK_HOOK_PREFIX}${name}` });
const legacy = (name: string) => ({ type: 'command', command: `node_modules/.bin/folkways hook ${name}` });
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

  it('removes entries in the old `folkways hook` form', () => {
    const current = { hooks: { SessionStart: [{ hooks: [legacy('session-start'), theirs] }] } };
    const out = mergeSettings(current, { hooks: { SessionStart: [{ hooks: [ours('session-start')] }] } });
    expect(out.hooks).toEqual({ SessionStart: [{ hooks: [theirs] }, { hooks: [ours('session-start')] }] });
  });

  it('sets disableAllHooks from the fragment, over the team value', () => {
    expect(mergeSettings({ disableAllHooks: true }, { disableAllHooks: false })).toEqual({ disableAllHooks: false });
    expect(mergeSettings({ disableAllHooks: true }, {})).toEqual({ disableAllHooks: true });
  });

  it('does not mutate its input', () => {
    const current = { hooks: { Stop: [{ hooks: [ours('x'), theirs] }] } };
    mergeSettings(current, {});
    expect(current.hooks.Stop[0]!.hooks).toHaveLength(2);
  });
});

describe('isFrameworkHook', () => {
  it('owns commands that start with the framework prefix, or the old one', () => {
    expect(FRAMEWORK_HOOK_PREFIX).toBe('"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook ');
    expect(isFrameworkHook('"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook git-guard')).toBe(true);
    expect(isFrameworkHook('node_modules/.bin/folkways hook session-start')).toBe(true);
  });

  it.each([
    'scripts/run-folkways hook x',
    'echo folkways hook x',
    'npx folkways hook x',
    'scripts/team.sh && node_modules/.bin/folkways hook x',
    'node_modules/.bin/folkways hooks x',
    'node_modules/.bin/folkways-hook x',
    'scripts/team.sh && "${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook x',
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

describe('assets/settings.json', () => {
  const settings = json(join(ASSETS, 'settings.json'));
  const entries = Object.entries(settings.hooks as Record<string, { matcher?: string; hooks: { command: string }[] }[]>).flatMap(
    ([event, groups]) => groups.flatMap((g) => g.hooks.map((h) => ({ event, matcher: g.matcher, command: h.command }))),
  );

  it('wires each hook to its events (spec §7.1)', () => {
    const wired = entries.map((e) => `${e.event}${e.matcher ? `(${e.matcher})` : ''} ${e.command.slice(FRAMEWORK_HOOK_PREFIX.length)}`);
    expect(wired.sort()).toEqual([
      'PostToolBatch context-band',
      'PreCompact precompact',
      'PreToolUse(Bash) git-guard',
      'SessionStart session-start',
      'UserPromptSubmit context-band',
    ]);
  });

  it('owns every entry and keeps hooks on (ADR 0001 M11)', () => {
    expect(entries.every((e) => isFrameworkHook(e.command))).toBe(true);
    expect(settings.disableAllHooks).toBe(false);
  });
});
