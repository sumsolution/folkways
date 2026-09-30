// Bench prototype (not a production hook). Same source runs on Node (types stripped), Bun and Deno.
// Spec: ../README.md#the-bench-hook
import { readFileSync } from 'node:fs';
import { parse } from 'yaml';

interface Payload {
  tool_name?: string;
  tool_input?: { command?: string };
}
interface Team {
  git?: { scopes?: string[] };
}

const TYPES = 'feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert';
const HEADER = new RegExp(`^(?:${TYPES})(?:\\(([a-z0-9._-]+)\\))?!?: \\S`);
const COMMIT = /\bgit\s+commit\b[^|;&]*?\s-m\s*(?:"([^"]*)"|'([^']*)')/;

function deny(reason: string): void {
  const out = { hookSpecificOutput: { hookEventName: 'PreToolUse', permissionDecision: 'deny', permissionDecisionReason: reason } };
  process.stdout.write(JSON.stringify(out) + '\n');
}

export function run(): void {
  const payload = JSON.parse(readFileSync(0, 'utf8')) as Payload;
  const command = payload.tool_name === 'Bash' ? (payload.tool_input?.command ?? '') : '';
  const match = COMMIT.exec(command);
  if (!match) return;
  const team = (parse(readFileSync('team.yaml', 'utf8')) ?? {}) as Team;
  const scopes = team.git?.scopes ?? [];
  const subject = (match[1] ?? match[2] ?? '').split('\n')[0] ?? '';
  const header = HEADER.exec(subject);
  if (!header) {
    deny(`Commit message is not Angular "type(scope): subject": "${subject}".`);
  } else if (scopes.length > 0 && !scopes.includes(header[1] ?? '')) {
    deny(`Scope "${header[1] ?? ''}" is not in team.yaml git.scopes (${scopes.join(', ')}).`);
  }
}
