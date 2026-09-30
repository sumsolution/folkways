import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { detectPackageManager } from './package-manager.ts';

/** The Go hook binary, linked by the `@folkways/hooks-<os>-<cpu>` optional dependency (ADR 0002). */
export const HOOK_BIN = 'node_modules/.bin/folkways-hook';

/** pnpm links a dependency's dependencies' bins only when hoisted; the pattern that hoists ours. */
export const HOIST_PATTERN = '@folkways/hooks-*';

/** `pnpm-workspace.yaml` for a new pnpm team repo. */
export const PNPM_WORKSPACE = `publicHoistPattern:\n  - "${HOIST_PATTERN}"\n`;

const unquote = (s: string): string => s.trim().replace(/^(["'])(.*)\1$/, '$2');

/**
 * Whether a `pnpm-workspace.yaml` hoists the hook binary: `shamefullyHoist: true`, or a
 * `publicHoistPattern` list (block or inline) with our pattern or `*`.
 */
export function hoistsHooks(yaml: string): boolean {
  if (/^shamefullyHoist:\s*true\s*(#.*)?$/m.test(yaml)) return true;
  const lines = yaml.split('\n');
  const start = lines.findIndex((l) => /^publicHoistPattern:/.test(l));
  if (start < 0) return false;
  const inline = /^publicHoistPattern:\s*\[(.*)\]/.exec(lines[start]!)?.[1];
  const items = inline?.split(',') ?? [];
  if (inline === undefined) {
    for (const line of lines.slice(start + 1)) {
      if (line.trim() === '') continue;
      const item = /^\s+-\s*(.*)$/.exec(line)?.[1];
      if (item === undefined) break;
      items.push(item);
    }
  }
  return items.map(unquote).some((p) => p === HOIST_PATTERN || p === '*');
}

/** What would stop hooks from running in this team repo; `sync` prints these as warnings. */
export function hookWarnings(team: string): string[] {
  const warnings: string[] = [];
  let pm: string;
  try {
    pm = detectPackageManager(team);
  } catch {
    pm = 'other';
  }
  if (pm === 'pnpm') {
    const path = join(team, 'pnpm-workspace.yaml');
    if (!existsSync(path) || !hoistsHooks(readFileSync(path, 'utf8'))) {
      warnings.push(
        `pnpm-workspace.yaml needs publicHoistPattern: ["${HOIST_PATTERN}"], or pnpm won't link ${HOOK_BIN}. Add it and run pnpm install.`,
      );
    }
  }
  if (existsSync(join(team, 'node_modules')) && !existsSync(join(team, HOOK_BIN))) {
    warnings.push(
      `${HOOK_BIN} is missing, so hooks will fail. Reinstall without --omit=optional, and check that your registry serves ${HOIST_PATTERN}.`,
    );
  }
  return warnings;
}

/** Run a hook by hand (spec §7.2): the binary reads the payload on our stdin. Returns its exit code. */
export function runHook(team: string, name: string): number {
  const bin = join(team, HOOK_BIN);
  if (!existsSync(bin)) {
    throw new FolkwaysError(`${HOOK_BIN} is missing; install dependencies (npm install or pnpm install) in ${team}`);
  }
  const r = spawnSync(bin, [name], { stdio: 'inherit', env: { ...process.env, CLAUDE_PROJECT_DIR: team } });
  if (r.error) throw new FolkwaysError(`could not run ${HOOK_BIN}: ${r.error.message}`);
  return r.status ?? 1;
}
