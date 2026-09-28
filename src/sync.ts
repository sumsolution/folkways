import { copyFileSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { listFiles, pruneEmptyDirs, sha256 } from './files.ts';
import { type Manifest, readManifest, writeManifest } from './manifest.ts';
import { PACKAGE_ROOT, packageVersion } from './package.ts';
import { mergeSettings, type Settings } from './settings.ts';

export const INSTRUCTIONS_IMPORT = '@.folkways/instructions.md';

interface FileCopy {
  src: string;
  dest: string;
}

/** Every file `sync` copies from the package into a team repo. */
export function assetMap(pkg: string): FileCopy[] {
  const assets = join(pkg, 'assets');
  const out: FileCopy[] = [{ src: join(assets, 'instructions.md'), dest: '.folkways/instructions.md' }];
  for (const f of listFiles(join(assets, 'prompts'))) {
    out.push({ src: join(assets, 'prompts', f), dest: `.folkways/prompts/${f}` });
  }
  for (const f of listFiles(join(assets, 'skills'))) {
    if (f.includes('/')) out.push({ src: join(assets, 'skills', f), dest: `.claude/skills/${f}` });
  }
  for (const f of listFiles(join(assets, 'workflows'))) {
    if (!f.includes('/') && f.endsWith('.yml')) {
      out.push({ src: join(assets, 'workflows', f), dest: `.github/workflows/folkways-${f}` });
    }
  }
  return out;
}

export function skillNames(pkg: string): string[] {
  const dir = join(pkg, 'assets', 'skills');
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();
}

/** Framework-owned files changed since the last sync, e.g. `modified  .folkways/instructions.md`. */
export function findEdits(team: string, manifest: Manifest): string[] {
  const edits: string[] = [];
  for (const [path, hash] of Object.entries(manifest.files)) {
    const full = join(team, path);
    if (!existsSync(full)) edits.push(`deleted   ${path}`);
    else if (sha256(full) !== hash) edits.push(`modified  ${path}`);
  }
  for (const skill of manifest.skills) {
    for (const f of listFiles(join(team, '.claude/skills', skill))) {
      const path = `.claude/skills/${skill}/${f}`;
      if (!(path in manifest.files)) edits.push(`added     ${path}`);
    }
  }
  return edits;
}

export function editsMessage(edits: string[], fix: string): string {
  return [
    'framework files were edited:',
    ...edits.map((e) => `  ${e}`),
    'Move team changes into overlays/ (team) or people/<handle>/overlays/ (personal), then rerun.',
    fix,
  ].join('\n');
}

function removeSynced(team: string, manifest: Manifest): void {
  for (const skill of manifest.skills) rmSync(join(team, '.claude/skills', skill), { recursive: true, force: true });
  for (const path of Object.keys(manifest.files)) rmSync(join(team, path), { force: true });
  pruneEmptyDirs(join(team, '.folkways'));
  pruneEmptyDirs(join(team, '.github'));
}

function syncSettings(team: string, pkg: string): void {
  const path = join(team, '.claude/settings.json');
  let current: Settings = {};
  if (existsSync(path)) {
    try {
      current = JSON.parse(readFileSync(path, 'utf8')) as Settings;
    } catch {
      throw new FolkwaysError(`${path} is not valid JSON`);
    }
  }
  const fragment = JSON.parse(readFileSync(join(pkg, 'assets/settings.json'), 'utf8')) as Settings;
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, `${JSON.stringify(mergeSettings(current, fragment), null, 2)}\n`);
}

export interface SyncOptions {
  /** Discard edits to framework-owned files instead of stopping. */
  force?: boolean;
}

export interface SyncResult {
  version: string;
  previous: string | undefined;
  discarded: string[];
  warnings: string[];
}

/** Copy the package's text files into a team repo (spec §3.1). Does not commit. */
export function sync(team: string, pkg: string = PACKAGE_ROOT, options: SyncOptions = {}): SyncResult {
  const manifest = readManifest(team);
  const discarded = manifest ? findEdits(team, manifest) : [];
  if (discarded.length > 0 && !options.force) {
    throw new FolkwaysError(editsMessage(discarded, 'Or rerun with --force to discard the edits.'));
  }

  const owned = new Set(manifest?.skills ?? []);
  const skills = skillNames(pkg);
  const clashes = skills.filter((s) => !owned.has(s) && existsSync(join(team, '.claude/skills', s)));
  if (clashes.length > 0) {
    throw new FolkwaysError(
      `team skills clash with new framework skills: ${clashes.map((s) => `.claude/skills/${s}`).join(', ')}\n` +
        'Rename the team skills, then rerun.',
    );
  }

  if (manifest) removeSynced(team, manifest);
  const files: Record<string, string> = {};
  for (const { src, dest } of assetMap(pkg)) {
    const full = join(team, dest);
    mkdirSync(dirname(full), { recursive: true });
    copyFileSync(src, full);
    files[dest] = sha256(full);
  }
  syncSettings(team, pkg);
  const version = packageVersion(pkg);
  writeManifest(team, { version, skills, files });

  const warnings: string[] = [];
  const agents = join(team, 'AGENTS.md');
  const imports = existsSync(agents) && readFileSync(agents, 'utf8').split('\n').some((l) => l.trim() === INSTRUCTIONS_IMPORT);
  if (!imports) warnings.push(`AGENTS.md has no '${INSTRUCTIONS_IMPORT}' line; agents won't see the framework instructions.`);

  return { version, previous: manifest?.version, discarded, warnings };
}
