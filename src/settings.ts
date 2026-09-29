/** Hook entries whose command contains this are framework-owned (spec §3.1). */
export const HOOK_MARK = 'folkways hook ';

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

const isOurs = (h: HookEntry): boolean => (h.command ?? '').includes(HOOK_MARK);

/**
 * Merge the framework settings fragment into a team's .claude/settings.json.
 * Replaces framework hook entries, adds missing framework `ask` rules, keeps everything else.
 */
export function mergeSettings(current: Settings, fragment: Settings): Settings {
  const out = structuredClone(current);

  const hooks: Record<string, HookGroup[]> = {};
  for (const [event, groups] of Object.entries(out.hooks ?? {})) {
    const kept = groups
      .map((g) => ({ ...g, hooks: g.hooks.filter((h) => !isOurs(h)) }))
      .filter((g) => g.hooks.length > 0);
    if (kept.length > 0) hooks[event] = kept;
  }
  for (const [event, groups] of Object.entries(fragment.hooks ?? {})) {
    hooks[event] = [...(hooks[event] ?? []), ...structuredClone(groups)];
  }
  if (Object.keys(hooks).length > 0) out.hooks = hooks;
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
