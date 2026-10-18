import { spawn } from 'node:child_process';
import type { Case } from './suite.ts';

export interface RunResult {
  code: number;
  stdout: string;
  stderr: string;
  timedOut: boolean;
}

/**
 * Runs `claude` with args in cwd. `input` goes to stdin. `configDir` becomes `CLAUDE_CONFIG_DIR`, so no
 * user skills, plugins, CLAUDE.md or memory load. Injected so tests never call a model.
 */
export type Runner = (
  args: string[],
  options: { cwd: string; configDir: string; input?: string; timeoutMs: number },
) => Promise<RunResult>;

/** An empty config dir has no login, so auth must come from the environment. */
export const AUTH_VARS = ['CLAUDE_CODE_OAUTH_TOKEN', 'ANTHROPIC_API_KEY'];

export function claudeEnv(configDir: string, env: NodeJS.ProcessEnv = process.env): NodeJS.ProcessEnv {
  return { ...env, CLAUDE_CONFIG_DIR: configDir };
}

/** Tools that `--tools` leaves in place (claude 2.1.161, seen 2026-10-18), so they are denied by name. */
export const UNLISTED_TOOLS = ['Monitor', 'PushNotification', 'RemoteTrigger', 'DesignSync'];

/**
 * The subject: one headless session in the team repo. Project settings, CLAUDE.md and skills load;
 * user settings, hooks and MCP servers don't; only the case's tools exist.
 */
export function subjectArgs(c: Case, model: string): string[] {
  return [
    '-p', c.prompt,
    '--model', c.model ?? model,
    '--max-turns', String(c.maxTurns),
    '--output-format', 'stream-json',
    '--verbose',
    '--setting-sources', 'project,local',
    '--strict-mcp-config',
    '--no-session-persistence',
    '--tools', c.allowedTools.join(','),
    '--disallowed-tools', UNLISTED_TOOLS.filter((t) => !c.allowedTools.includes(t)).join(','),
  ];
}

/** The judge: no tools, no settings, run in an empty directory. The prompt comes on stdin. */
export function judgeArgs(model: string): string[] {
  return [
    '-p',
    '--model', model,
    '--output-format', 'json',
    '--tools', '',
    '--setting-sources', '',
    '--strict-mcp-config',
    '--no-session-persistence',
  ];
}

export const spawnClaude: Runner = (args, { cwd, configDir, input, timeoutMs }) =>
  new Promise((resolve) => {
    const child = spawn('claude', args, { cwd, env: claudeEnv(configDir), stdio: [input === undefined ? 'ignore' : 'pipe', 'pipe', 'pipe'] });
    let stdout = '';
    let stderr = '';
    let timedOut = false;
    const timer = setTimeout(() => {
      timedOut = true;
      child.kill('SIGTERM');
    }, timeoutMs);
    child.stdout?.on('data', (d) => (stdout += d));
    child.stderr?.on('data', (d) => (stderr += d));
    child.on('error', (e) => (stderr += String(e)));
    child.on('close', (code) => {
      clearTimeout(timer);
      resolve({ code: code ?? 1, stdout, stderr, timedOut });
    });
    if (input !== undefined) child.stdin?.end(input);
  });
