import { FolkwaysError } from './errors.ts';

/**
 * Hook entries whose command starts with this are framework-owned (spec §3.1).
 * The hook runtime (ADR 0002) may change the command form; change it here only.
 */
export const FRAMEWORK_HOOK_PREFIX = 'node_modules/.bin/folkways hook ';

export function isFrameworkHook(command: string): boolean {
  return command.startsWith(FRAMEWORK_HOOK_PREFIX);
}

interface HookEntry {
  command?: string;
  [key: string]: unknown;
}
interface HookGroup {
  hooks: HookEntry[];
  [key: string]: unknown;
}
export interface Settings {
  hooks?: Record<string, HookGroup[]>;
  permissions?: { ask?: string[]; [key: string]: unknown };
  [key: string]: unknown;
}

const isObject = (v: unknown): v is Record<string, unknown> => typeof v === 'object' && v !== null && !Array.isArray(v);

const isOurs = (h: unknown): boolean => isObject(h) && typeof h.command === 'string' && isFrameworkHook(h.command);

/** Drop framework entries from a group. Groups and events of unexpected shape stay as they are. */
function withoutOurs(group: unknown): unknown {
  if (!isObject(group) || !Array.isArray(group.hooks)) return group;
  return { ...group, hooks: group.hooks.filter((h) => !isOurs(h)) };
}

const isEmptyGroup = (g: unknown): boolean => isObject(g) && Array.isArray(g.hooks) && g.hooks.length === 0;

/**
 * Merge the framework settings fragment into a team's .claude/settings.json.
 * Replaces framework hook entries, adds missing framework `ask` rules, keeps everything else.
 * Throws a FolkwaysError when the team settings have a shape it can't merge into.
 */
export function mergeSettings(current: Settings, fragment: Settings): Settings {
  if (!isObject(current)) throw new FolkwaysError('.claude/settings.json must be a JSON object');
  if (current.hooks !== undefined && !isObject(current.hooks)) {
    throw new FolkwaysError('.claude/settings.json: hooks must be a JSON object');
  }
  if (current.permissions !== undefined && !isObject(current.permissions)) {
    throw new FolkwaysError('.claude/settings.json: permissions must be a JSON object');
  }
  if (current.permissions?.ask !== undefined && !Array.isArray(current.permissions.ask)) {
    throw new FolkwaysError('.claude/settings.json: permissions.ask must be an array');
  }
  const out = structuredClone(current);

  const hooks: Record<string, unknown> = {};
  for (const [event, groups] of Object.entries(out.hooks ?? {})) {
    if (!Array.isArray(groups)) {
      hooks[event] = groups;
      continue;
    }
    const kept = groups.map(withoutOurs).filter((g) => !isEmptyGroup(g));
    if (kept.length > 0) hooks[event] = kept;
  }
  for (const [event, groups] of Object.entries(fragment.hooks ?? {})) {
    const existing = hooks[event] ?? [];
    if (!Array.isArray(existing)) {
      throw new FolkwaysError(`.claude/settings.json: hooks.${event} must be an array`);
    }
    hooks[event] = [...existing, ...structuredClone(groups)];
  }
  if (Object.keys(hooks).length > 0) out.hooks = hooks as Settings['hooks'];
  else delete out.hooks;

  const ask = fragment.permissions?.ask ?? [];
  if (ask.length > 0) {
    const permissions = out.permissions ?? {};
    const existing = permissions.ask ?? [];
    permissions.ask = [...existing, ...ask.filter((rule) => !existing.includes(rule))];
    out.permissions = permissions;
  }
  return out;
}
