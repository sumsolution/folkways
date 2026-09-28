import { describe, expect, it } from 'vitest';
import { mergeSettings } from '../src/settings.ts';

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
