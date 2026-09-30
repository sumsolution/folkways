// THROWAWAY reference implementation of the folkways hooks, to validate the contract fixtures.
// Not production code. Usage: node hooks.mjs <hook>  (payload JSON on stdin). Always exits 0.
import { appendFileSync, existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

// FIXTURE-ONLY parser: the harness writes each top-level key as `key: <JSON value>` on one line.
// The real hooks use a YAML library.
function readYaml(path) {
  const out = {};
  try {
    for (const line of readFileSync(path, 'utf8').split('\n')) {
      const m = /^([A-Za-z_][\w-]*):\s*(.*)$/.exec(line);
      if (!m) continue;
      try { out[m[1]] = JSON.parse(m[2]); } catch { /* skip */ }
    }
  } catch { /* missing */ }
  return out;
}

const readStdin = () => { try { return readFileSync(0, 'utf8'); } catch { return ''; } };
const emit = (obj) => process.stdout.write(JSON.stringify(obj) + '\n');
const ctx = (event, text, top) => emit({ hookSpecificOutput: { hookEventName: event, additionalContext: text }, ...(top ? { additionalContext: text } : {}) });

function sessionStart(p, team, sid) {
  const dir = join(team, '.local', 'sessions');
  let unclosed = false;
  try { unclosed = readdirSync(dir).some((f) => f.endsWith('.unclosed')); } catch { /* none */ }
  const parts = ['Run the `start` skill (start procedure) before doing anything else.'];
  if (p.source === 'compact') parts.push('Context was just compacted: update 00-STATUS.md now.');
  if (unclosed) parts.push('A previous session ended without closing: offer the user the `session-close` skill.');
  ctx('SessionStart', parts.join(' '), true);
}

function sessionEnd(p, team, sid) {
  if (!sid) return;
  const dir = join(team, '.local', 'sessions');
  mkdirSync(dir, { recursive: true });
  writeFileSync(join(dir, `${sid}.unclosed`), '');
}

function precompact(p, team, sid) {
  mkdirSync(join(team, '.local'), { recursive: true });
  appendFileSync(join(team, '.local', 'precompact.log'), `${new Date().toISOString()} ${sid ?? 'unknown'} ${p.trigger ?? 'unknown'}\n`);
}

function tokens(path) {
  const lines = readFileSync(path, 'utf8').split('\n');
  for (let i = lines.length - 1; i >= 0; i--) {
    let l;
    try { l = JSON.parse(lines[i]); } catch { continue; }
    if (l?.type !== 'assistant' || l.isSidechain !== false) continue;
    const u = l.message?.usage;
    if (!u) continue;
    const n = (u.input_tokens || 0) + (u.cache_read_input_tokens || 0) + (u.cache_creation_input_tokens || 0);
    if (n > 0) return n;
  }
  return undefined;
}

function bandText(label, i) {
  if (i === 0) return `Context passed ${label} tokens. Look for a stopping point.`;
  if (i === 1) return `Context passed ${label} tokens. Start no new multi-step work; propose a handoff at the next checkpoint.`;
  return `Context passed ${label} tokens. Write 00-STATUS.md (handoff skill) and end the session.`;
}

function contextBand(p, team, sid) {
  if (!sid || !p.transcript_path || !existsSync(p.transcript_path)) return;
  const n = tokens(p.transcript_path);
  if (n === undefined) return;
  const cfg = readYaml(join(team, 'team.yaml'));
  let bands = cfg.handoff?.bands;
  if (!Array.isArray(bands) || !bands.length) bands = [125000, 150000, 200000];
  bands = bands.map(Number).sort((a, b) => a - b);
  let idx = -1;
  bands.forEach((b, i) => { if (n >= b) idx = i; });
  const file = join(team, '.local', 'sessions', `${sid}.band`);
  let stored = -1; // index of last announced band
  try { stored = parseInt(readFileSync(file, 'utf8'), 10); if (Number.isNaN(stored)) stored = -1; } catch { /* none */ }
  if (idx === stored) return;
  mkdirSync(join(team, '.local', 'sessions'), { recursive: true });
  writeFileSync(file, String(idx));
  if (idx > stored) {
    const label = `${bands[idx] / 1000}K`;
    ctx(p.hook_event_name || 'UserPromptSubmit', bandText(label, Math.min(idx, 2)), false);
  }
}

function main() {
  const hook = process.argv[2];
  if (hook === 'git-guard') return;
  let p = {};
  try { p = JSON.parse(readStdin()) ?? {}; } catch { return; }
  if (typeof p !== 'object') return;
  const team = process.env.CLAUDE_PROJECT_DIR || process.cwd();
  const sid = p.session_id ?? p.sessionId;
  const fn = { 'session-start': sessionStart, 'session-end': sessionEnd, precompact, 'context-band': contextBand }[hook];
  fn?.(p, team, sid);
}

try { main(); } catch { /* fail open */ }
process.exit(0);
