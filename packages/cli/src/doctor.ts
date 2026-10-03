import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join } from 'node:path';
import { HOOK_NAMES, hookWarnings } from './hooks.ts';
import { MANIFEST, readManifest } from './manifest.ts';
import type { LintFinding } from './overlays.ts';
import { PACKAGE_NAME, PACKAGE_ROOT, packageVersion } from './package.ts';
import { detectPackageManager, installCommand } from './package-manager.ts';
import { findEdits, lintPersonalOverlays, staleOverlays } from './sync.ts';

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
  let manifest: ReturnType<typeof readManifest>;
  try {
    manifest = readManifest(team);
  } catch (e) {
    findings.push(error((e as Error).message));
  }
  findings.push(...versionFindings(team, pkg, manifest?.version));
  if (manifest) {
    findings.push(...findEdits(team, manifest).map((e) => warning(`${e} is a framework file; sync refuses until you move the change to an overlay, or run 'folkways sync --force' to discard it`)));
    findings.push(...staleOverlays(team, manifest).map((p) => error(`${p} changed since the last sync compiled it; run 'folkways sync'`)));
  } else if (!findings.some((f) => f.message.startsWith(MANIFEST))) {
    findings.push(error(`no ${MANIFEST}; run 'folkways sync'`));
  }
  findings.push(...lintPersonalOverlays(team));
  findings.push(...shadowFindings(team, home));
  findings.push(...settingsFindings(team));
  findings.push(...hookWarnings(team).map(error));
  return findings;
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

  const findings: Finding[] = [];
  const installed = (readJson(join(team, 'node_modules', PACKAGE_NAME, 'package.json')) as { version?: string } | undefined)?.version;
  if (!installed) findings.push(error(`${PACKAGE_NAME} is not installed; run '${install}'`));
  else if (installed !== pinned) {
    findings.push(error(`node_modules has ${PACKAGE_NAME} ${installed} but package.json pins ${pinned} (a failed upgrade leaves this); run '${install}'`));
  }
  const running = packageVersion(pkg);
  if (running !== pinned) {
    findings.push(warning(`this is folkways ${running}, but the repo pins ${pinned}; run the pinned one with 'npx folkways doctor' from the team repo`));
  }
  if (synced && synced !== pinned) findings.push(error(`the repo was synced from ${synced}, but package.json pins ${pinned}; run 'folkways sync'`));
  return findings;
}

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

/**
 * Checks the `hooks:` block of folkways.yaml: each value is `true` or `false`. The hook binary (yaml.v3) also reads
 * YAML 1.1 words (`off`, `no`) and empty values as booleans, and anything else leaves the hook on; `run.sh` in the
 * plugin tests only some of these, so doctor asks for the two plain words.
 */
export function hookValueFindings(yaml: string, known: readonly string[]): Finding[] {
  const lines = yaml.split('\n');
  const start = lines.findIndex((l) => /^hooks:/.test(l));
  if (start < 0) return [];
  const inline = stripComment(lines[start]!.slice('hooks:'.length));
  if (inline !== '') {
    return inline === '{}' ? [] : [warning('folkways.yaml hooks: write one hook per line (git-guard: false) so doctor can check the values')];
  }
  const findings: Finding[] = [];
  for (const line of lines.slice(start + 1)) {
    if (line.trim() === '' || line.trim().startsWith('#')) continue;
    const entry = /^\s+([^\s:#][^:]*?):(?:\s+(.*))?$|^\s+([^\s:#][^:]*?):$/.exec(line);
    if (!/^\s/.test(line) || !entry) break;
    const name = entry[1] ?? entry[3]!;
    const raw = stripComment(entry[2] ?? '');
    const shown = raw === '' ? '(empty)' : raw;
    const where = `folkways.yaml hooks.${name}`;
    if (!known.includes(name)) {
      findings.push(warning(`${where}: no hook by that name (hooks: ${[...known].sort().join(', ')})`));
      continue;
    }
    if (raw === 'true' || raw === 'false') continue;
    const value = hookValue(raw);
    if (value === undefined) findings.push(error(`${where}: ${shown} is not a boolean, so the hook stays on; write true or false`));
    else findings.push(error(`${where}: ${shown} reads as ${value} (hook ${value ? 'on' : 'off'}); write ${value}`));
  }
  return findings;
}
