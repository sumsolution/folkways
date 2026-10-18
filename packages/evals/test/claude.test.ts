import { describe, expect, it } from 'vitest';
import { claudeEnv, judgeArgs, subjectArgs } from '../src/claude.ts';
import type { Case } from '../src/suite.ts';

const c: Case = {
  name: 'a',
  dir: '/x',
  prompt: 'Do it.',
  runs: 3,
  maxTurns: 20,
  timeoutSeconds: 600,
  allowedTools: ['Read', 'Skill'],
  graders: [],
};

describe('subjectArgs', () => {
  it('isolates the run: project settings only, no MCP, no saved session, only the case tools', () => {
    const args = subjectArgs(c, 'sonnet');
    expect(args.slice(0, 2)).toEqual(['-p', 'Do it.']);
    const flag = (name: string) => args[args.indexOf(name) + 1];
    expect(flag('--model')).toBe('sonnet');
    expect(flag('--max-turns')).toBe('20');
    expect(flag('--output-format')).toBe('stream-json');
    expect(flag('--setting-sources')).toBe('project,local');
    expect(flag('--tools')).toBe('Read,Skill');
    expect(args).toContain('--verbose');
    expect(args).toContain('--strict-mcp-config');
    expect(args).toContain('--no-session-persistence');
    expect(args).not.toContain('--mcp-config');
  });

  it('uses the case model over the default', () => {
    const args = subjectArgs({ ...c, model: 'opus' }, 'sonnet');
    expect(args[args.indexOf('--model') + 1]).toBe('opus');
  });
});

describe('judgeArgs', () => {
  it('gives the judge no tools and no project or user settings; the prompt comes on stdin', () => {
    const args = judgeArgs('sonnet');
    expect(args[0]).toBe('-p');
    expect(args[args.indexOf('--tools') + 1]).toBe('');
    expect(args[args.indexOf('--setting-sources') + 1]).toBe('');
    expect(args).toContain('--strict-mcp-config');
  });
});

describe('claudeEnv', () => {
  it('points CLAUDE_CONFIG_DIR at the run dir and keeps the rest, auth included', () => {
    const env = claudeEnv('/tmp/run-1/claude-config', { CLAUDE_CONFIG_DIR: '/home/me/.claude', CLAUDE_CODE_OAUTH_TOKEN: 't' });
    expect(env.CLAUDE_CONFIG_DIR).toBe('/tmp/run-1/claude-config');
    expect(env.CLAUDE_CODE_OAUTH_TOKEN).toBe('t');
  });
});
