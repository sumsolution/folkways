import { copyFileSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { listFiles, pruneEmptyDirs, sha256 } from './files.ts';
import { MANIFEST, type Manifest, readManifest, writeManifest } from './manifest.ts';
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
    else if (path === MANIFEST ? manifestHash(manifest) !== hash : sha256(full) !== hash) edits.push(`modified  ${path}`);
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

function manifestHash(manifest: Manifest): string {
  const files = { ...manifest.files };
  delete files[MANIFEST];
  return createHash('sha256')
    .update(`${JSON.stringify({ ...manifest, files }, null, 2)}\n`)
    .digest('hex');
}

function isWithin(root: string, target: string): boolean {
  const path = relative(resolve(root), resolve(target));
  return path !== '' && !path.startsWith(`..${sep}`) && path !== '..' && !path.includes(`..${sep}`);
}

function validSkill(skill: unknown): skill is string {
  return typeof skill === 'string' && /^[A-Za-z0-9][A-Za-z0-9._-]*$/.test(skill);
}

function validFile(team: string, path: unknown, skills: string[]): path is string {
  if (typeof path !== 'string' || path.includes('\\') || path.split('/').some((part) => !part || part === '.' || part === '..')) {
    return false;
  }
  if (path === MANIFEST || (path.startsWith('.folkways/') && isWithin(join(team, '.folkways'), join(team, path)))) return true;
  if (
    /^\.github\/workflows\/folkways-[^/]+\.yml$/.test(path) &&
    isWithin(join(team, '.github', 'workflows'), join(team, path))
  ) {
    return true;
  }
  return skills.some(
    (skill) =>
      path.startsWith(`.claude/skills/${skill}/`) && isWithin(join(team, '.claude', 'skills', skill), join(team, path)),
  );
}

function validateManifest(team: string, manifest: unknown): asserts manifest is Manifest {
  if (
    !manifest ||
    typeof manifest !== 'object' ||
    typeof (manifest as Manifest).version !== 'string' ||
    !Array.isArray((manifest as Manifest).skills) ||
    !(manifest as Manifest).skills.every(validSkill) ||
    !(manifest as Manifest).files ||
    typeof (manifest as Manifest).files !== 'object' ||
    Array.isArray((manifest as Manifest).files)
  ) {
    throw new FolkwaysError('manifest is invalid');
  }
  const { skills, files } = manifest as Manifest;
  if (
    new Set(skills).size !== skills.length ||
    !Object.entries(files).every(([path, hash]) => validFile(team, path, skills) && typeof hash === 'string' && /^[a-f0-9]{64}$/.test(hash)) ||
    !(MANIFEST in files)
  ) {
    throw new FolkwaysError('manifest is invalid');
  }
}

function removeSynced(team: string, manifest: Manifest): void {
  for (const skill of manifest.skills) {
    const target = join(team, '.claude/skills', skill);
    if (isWithin(join(team, '.claude', 'skills'), target)) rmSync(target, { recursive: true, force: true });
  }
  for (const path of Object.keys(manifest.files)) {
    const target = join(team, path);
    if (validFile(team, path, manifest.skills)) rmSync(target, { force: true });
  }
  pruneEmptyDirs(join(team, '.folkways'));
  pruneEmptyDirs(join(team, '.github'));
}

/** The merged .claude/settings.json content. Reads only, so sync can fail before writing anything. */
function mergedSettings(team: string, pkg: string): string {
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
  return `${JSON.stringify(mergeSettings(current, fragment), null, 2)}\n`;
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
  let manifest: Manifest | undefined;
  try {
    manifest = readManifest(team);
  } catch {
    throw new FolkwaysError('manifest is invalid');
  }
  if (manifest) validateManifest(team, manifest);
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

  const settings = mergedSettings(team, pkg);
  if (manifest) removeSynced(team, manifest);
  const files: Record<string, string> = {};
  for (const { src, dest } of assetMap(pkg)) {
    const full = join(team, dest);
    mkdirSync(dirname(full), { recursive: true });
    copyFileSync(src, full);
    files[dest] = sha256(full);
  }
  mkdirSync(join(team, '.claude'), { recursive: true });
  writeFileSync(join(team, '.claude/settings.json'), settings);
  const version = packageVersion(pkg);
  const nextManifest = { version, skills, files };
  files[MANIFEST] = manifestHash(nextManifest);
  writeManifest(team, nextManifest);

  const warnings: string[] = [];
  const agents = join(team, 'AGENTS.md');
  const imports = existsSync(agents) && readFileSync(agents, 'utf8').split('\n').some((l) => l.trim() === INSTRUCTIONS_IMPORT);
  if (!imports) warnings.push(`AGENTS.md has no '${INSTRUCTIONS_IMPORT}' line; agents won't see the framework instructions.`);

  return { version, previous: manifest?.version, discarded, warnings };
}
