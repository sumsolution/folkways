import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import { HOOK_NAMES, hookWarnings } from './hooks.ts';
import { MANIFEST, readManifest } from './manifest.ts';
import type { LintFinding } from './overlays.ts';
import { PACKAGE_NAME, PACKAGE_ROOT, packageVersion } from './package.ts';
import { detectPackageManager, installCommand } from './package-manager.ts';
import { findEdits, staleOverlays } from './sync.ts';

export type Finding = LintFinding;

export interface DoctorOptions {
  /** The running CLI's package root. */
  pkg?: string;
  /** Where personal skills live (`~/.claude/skills`, ...). */
  home?: string;
}

const error = (message: string): Finding => ({ level: 'error', message });
const warning = (message: string): Finding => ({ level: 'warning', message });

/** Checks a team repo (spec §7.2): config, versions, overlays, skills, settings, hook install. Each message names its fix. */
export function doctor(team: string, options: DoctorOptions = {}): Finding[] {
  const pkg = options.pkg ?? PACKAGE_ROOT;
  const home = options.home ?? homedir();
  const config = join(team, 'folkways.yaml');
  if (!existsSync(config)) return [error(`no folkways.yaml in ${team}; run the setup skill, or run doctor in the team repo`)];

  const findings = hookValueFindings(readFileSync(config, 'utf8'), HOOK_NAMES);
  // No package.json: the plugin set this repo up and there is no CLI install, sync or settings to check.
  if (!existsSync(join(team, 'package.json'))) return [...findings, ...shadowFindings(team, home)];
  let manifest: ReturnType<typeof readManifest>;
  try {
    manifest = readManifest(team);
  } catch (e) {
    findings.push(error((e as Error).message));
  }
  findings.push(...versionFindings(team, pkg, manifest?.version));
  if (manifest) {
    findings.push(...findEdits(team, manifest).map(editFinding));
    findings.push(...staleOverlays(team, manifest).map((p) => error(`${p} changed since the last sync compiled it; run 'folkways sync'`)));
  } else if (!findings.some((f) => f.message.startsWith(MANIFEST))) {
    findings.push(error(`no ${MANIFEST}; run 'folkways sync'`));
  }
  findings.push(...shadowFindings(team, home));
  findings.push(...settingsFindings(team));
  findings.push(...hookWarnings(team).map(error));
  return findings;
}

/** `modified  .folkways/instructions.md` (from `findEdits`) as an error: sync refuses until it is resolved. */
function editFinding(edit: string): Finding {
  const [, how, path] = /^(\S+)\s+(.*)$/.exec(edit)!;
  return error(`${path} was ${how}, so sync refuses; move the change into an overlay, or run 'folkways sync --force' to discard it`);
}

function readJson(path: string): Record<string, unknown> | undefined {
  try {
    return JSON.parse(readFileSync(path, 'utf8')) as Record<string, unknown>;
  } catch {
    return undefined;
  }
}

/** The pinned, installed and running CLI versions, and the version the repo was last synced from. */
function versionFindings(team: string, pkg: string, synced: string | undefined): Finding[] {
  let install: string;
  try {
    install = installCommand(detectPackageManager(team)).flat().join(' ');
  } catch {
    install = 'npm install';
  }
  const manifest = readJson(join(team, 'package.json')) as
    | { devDependencies?: Record<string, string>; dependencies?: Record<string, string> }
    | undefined;
  const pinned = manifest?.devDependencies?.[PACKAGE_NAME] ?? manifest?.dependencies?.[PACKAGE_NAME];
  if (!pinned) return [error(`package.json doesn't pin ${PACKAGE_NAME}; add it as an exact devDependency ('folkways upgrade <version>' does this)`)];
  if (!EXACT.test(pinned)) {
    return [error(`package.json pins ${PACKAGE_NAME} ${pinned}, not an exact version; pin one with 'folkways upgrade <version>'`)];
  }

  const findings: Finding[] = [];
  const installed = (readJson(join(team, 'node_modules', PACKAGE_NAME, 'package.json')) as { version?: string } | undefined)?.version;
  if (!installed) findings.push(error(`${PACKAGE_NAME} is not installed; run '${install}'`));
  else if (installed !== pinned) {
    findings.push(error(`node_modules has ${PACKAGE_NAME} ${installed} but package.json pins ${pinned} (a failed upgrade leaves this); run '${install}'`));
  }
  const running = packageVersion(pkg);
  // A running CLI that is the installed one is covered by the node_modules error above.
  if (running !== pinned && running !== installed) {
    findings.push(warning(`this is folkways ${running}, but the repo pins ${pinned}; run the pinned one with 'npx folkways doctor' from the team repo`));
  }
  if (synced && synced !== pinned) findings.push(error(`the repo was synced from ${synced}, but package.json pins ${pinned}; run 'folkways sync'`));
  return findings;
}

const EXACT = /^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;

const PERSONAL_SKILL_DIRS = ['.claude/skills', '.agents/skills', '.copilot/skills'];

const dirs = (path: string): string[] =>
  existsSync(path)
    ? readdirSync(path, { withFileTypes: true })
        .filter((e) => e.isDirectory() || e.isSymbolicLink())
        .map((e) => e.name)
    : [];

/** Personal skills with the name of a framework or team skill (ADR 0001 M10). */
function shadowFindings(team: string, home: string): Finding[] {
  const repoSkills = new Set(dirs(join(team, '.claude/skills')));
  return PERSONAL_SKILL_DIRS.flatMap((d) =>
    dirs(join(home, d))
      .filter((name) => repoSkills.has(name))
      .sort()
      .map((name) => warning(`~/${d}/${name} shadows the skill ${name} in this repo; rename or remove it`)),
  );
}

/** `disableAllHooks` (ADR 0001 M11): the project sets false; a local true turns every hook off. */
function settingsFindings(team: string): Finding[] {
  const findings: Finding[] = [];
  for (const [file, required] of [
    ['.claude/settings.json', true],
    ['.claude/settings.local.json', false],
  ] as const) {
    const path = join(team, file);
    if (!existsSync(path)) {
      if (required) findings.push(error(`${file} is missing; run 'folkways sync'`));
      continue;
    }
    const settings = readJson(path);
    if (!settings) findings.push(error(`${file} is not valid JSON; fix it, or hooks and permissions may not load`));
    else if (required && settings.disableAllHooks !== false) {
      findings.push(error(`${file} must set "disableAllHooks: false" so a user setting can't turn hooks off; run 'folkways sync'`));
    } else if (!required && settings.disableAllHooks === true) {
      findings.push(error(`${file} sets disableAllHooks: true, which turns every folkways hook off; remove it`));
    }
  }
  return findings;
}

// How gopkg.in/yaml.v3 decodes a scalar into the hooks' `map[string]bool` (checked against yaml.v3 3.0.1).
const YAML_TRUE = new Set(['y', 'Y', 'yes', 'Yes', 'YES', 'on', 'On', 'ON']);
const YAML_FALSE = new Set(['n', 'N', 'no', 'No', 'NO', 'off', 'Off', 'OFF']);
const PLAIN_TRUE = new Set(['true', 'True', 'TRUE']);
const PLAIN_FALSE = new Set(['false', 'False', 'FALSE', '', '~', 'null', 'Null', 'NULL']);

/** What the hook binary reads for a raw `hooks:` value: true, false, or undefined (not a boolean: the hook stays on). */
function hookValue(raw: string): boolean | undefined {
  const quoted = /^(["'])(.*)\1$/.exec(raw);
  const s = quoted ? quoted[2]! : raw;
  if (!quoted && PLAIN_TRUE.has(s)) return true;
  if (!quoted && PLAIN_FALSE.has(s)) return false;
  if (YAML_TRUE.has(s)) return true;
  if (YAML_FALSE.has(s)) return false;
  return undefined;
}

/** A raw YAML value without its trailing comment. */
const stripComment = (v: string): string => (/^(["']).*?\1/.exec(v)?.[0] ?? v.replace(/(^|\s)#.*$/, '')).trim();

const NULL_MAP = new Set(['', '{}', '~', 'null', 'Null', 'NULL']);
const KEY = /^(?:"([^"]*)"|'([^']*)'|([^\s"'#][^:#]*?))\s*:(?:\s+(.*))?$/;

/**
 * Checks the `hooks:` block of folkways.yaml: each value is `true` or `false`. The hook binary (yaml.v3) also reads
 * YAML 1.1 words (`off`, `no`) and empty values as booleans, and anything else leaves the hook on; `run.sh` in the
 * plugin tests only some of these, so doctor asks for the two plain words.
 */
export function hookValueFindings(yaml: string, known: readonly string[]): Finding[] {
  const lines = yaml.split(/\r?\n/);
  const start = lines.findIndex((l) => /^hooks\s*:/.test(l));
  if (start < 0) return [];
  const inline = stripComment(lines[start]!.replace(/^hooks\s*:/, ''));
  if (!NULL_MAP.has(inline)) return [warning('folkways.yaml hooks: write one hook per line (git-guard: false) so doctor can check the values')];

  const entries: { name: string; raw: string; nested: boolean }[] = [];
  let indent = -1;
  for (let i = start + 1; i < lines.length; i++) {
    const line = lines[i]!;
    if (line.trim() === '' || line.trim().startsWith('#')) continue;
    if (/^ *\t/.test(line)) {
      return [error(`folkways.yaml hooks: line ${i + 1} is indented with a tab, which YAML forbids, so the hooks ignore folkways.yaml; indent with spaces`)];
    }
    const depth = /^ */.exec(line)![0].length;
    if (depth === 0) break;
    if (indent < 0) indent = depth;
    if (depth > indent) {
      // Deeper lines belong to the entry above: a nested map, or a plain value continued on the next line.
      const last = entries[entries.length - 1];
      if (last && KEY.test(line.trim())) last.nested = true;
      else if (last && !last.nested) last.raw = `${last.raw} ${stripComment(line)}`.trim();
      continue;
    }
    if (depth < indent) break;
    const m = KEY.exec(line.trim());
    if (!m) break;
    entries.push({ name: (m[1] ?? m[2] ?? m[3]!).trim(), raw: stripComment(m[4] ?? ''), nested: false });
  }

  const findings: Finding[] = [];
  for (const { name, raw, nested } of entries) {
    const where = `folkways.yaml hooks.${name}`;
    if (!known.includes(name)) {
      findings.push(warning(`${where}: no hook by that name (hooks: ${[...known].sort().join(', ')})`));
      continue;
    }
    if (nested && raw === '') {
      findings.push(error(`${where}: a nested map is not a boolean, so the hook stays on; write true or false`));
      continue;
    }
    if (raw === 'true' || raw === 'false') continue;
    if (/^[!*&]/.test(raw)) {
      findings.push(error(`${where}: doctor can't check ${raw} (a tag, anchor or alias); write true or false`));
      continue;
    }
    const shown = raw === '' ? '(empty)' : raw;
    const value = hookValue(raw);
    if (value === undefined) findings.push(error(`${where}: ${shown} is not a boolean, so the hook stays on; write true or false`));
    else findings.push(error(`${where}: ${shown} reads as ${value} (hook ${value ? 'on' : 'off'}); write ${value}`));
  }
  return findings;
}
