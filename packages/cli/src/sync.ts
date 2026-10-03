import {
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from 'node:fs';
import { dirname, join } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { fsOps, listFiles, pruneEmptyDirs, sha256, sha256Text } from './files.ts';
import { hookWarnings } from './hooks.ts';
import { MANIFEST, type Manifest, readManifest, writeManifest } from './manifest.ts';
import { compileSkillFiles } from './overlays.ts';
import { assetsDir, PACKAGE_ROOT, packageVersion } from './package.ts';
import { mergeSettings, type Settings } from './settings.ts';

export const INSTRUCTIONS_IMPORT = '@.folkways/instructions.md';

/** Staging folder for a sync in progress, inside the team repo. */
const STAGE_PREFIX = '.folkways-sync-';

interface FileCopy {
  src: string;
  dest: string;
  /** Compiled content that replaces `src` (a skill file with a team overlay applied). */
  content?: string;
}

const copyHash = (c: FileCopy): string => (c.content === undefined ? sha256(c.src) : sha256Text(c.content));

/** Every file `sync` copies from the package into a team repo. */
export function assetMap(pkg: string): FileCopy[] {
  const assets = assetsDir(pkg);
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
  const dir = join(assetsDir(pkg), 'skills');
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
    'Move team changes into overlays/, then rerun.',
    fix,
  ].join('\n');
}

/** The first existing parent of `dest` that isn't a folder, e.g. a `.folkways` file. */
function blockingParent(team: string, dest: string): string | undefined {
  const parts = dest.split('/');
  for (let i = 1; i < parts.length; i++) {
    const path = parts.slice(0, i).join('/');
    const stat = lstatSync(join(team, path), { throwIfNoEntry: false });
    if (!stat) return undefined;
    if (!stat.isDirectory()) return path;
  }
  return undefined;
}

/**
 * Team files this version would overwrite that the manifest doesn't track, e.g. `replaced  .folkways/x.md`.
 * Without a manifest (adopting a repo) that is every destination, plus extra files in framework skill folders.
 * Files that already match are fine.
 */
function findOverwrites(team: string, pkg: string, manifest: Manifest | undefined, copies: FileCopy[]): string[] {
  const tracked = new Set(Object.keys(manifest?.files ?? {}));
  const ownedSkills = (manifest?.skills ?? []).map((s) => `.claude/skills/${s}/`);
  const out = new Set<string>();
  for (const dest of [...copies.map((c) => c.dest), '.claude/settings.json', MANIFEST]) {
    const parent = blockingParent(team, dest);
    if (parent) out.add(`replaced  ${parent}`);
  }
  for (const c of copies) {
    const { dest } = c;
    if (tracked.has(dest) || ownedSkills.some((dir) => dest.startsWith(dir)) || blockingParent(team, dest)) continue;
    const full = join(team, dest);
    const stat = lstatSync(full, { throwIfNoEntry: false });
    if (stat && (!stat.isFile() || sha256(full) !== copyHash(c))) out.add(`replaced  ${dest}`);
  }
  if (!manifest) {
    // Adopting: files in framework skill folders that this version doesn't ship would be left behind.
    const shipped = new Set(copies.map((c) => c.dest));
    for (const skill of skillNames(pkg)) {
      if (blockingParent(team, `.claude/skills/${skill}/x`)) continue;
      for (const f of listFiles(join(team, '.claude/skills', skill))) {
        const path = `.claude/skills/${skill}/${f}`;
        if (!shipped.has(path)) out.add(`replaced  ${path}`);
      }
    }
  }
  return [...out];
}

/**
 * Paths this sync replaces or removes: the last sync's files, edits in its skills, forced overwrites,
 * and every existing destination, so a failed swap can put all of them back.
 */
function replacedPaths(team: string, manifest: Manifest | undefined, overwrites: string[], incoming: string[]): string[] {
  const paths = new Set([...Object.keys(manifest?.files ?? {}), ...incoming]);
  for (const skill of manifest?.skills ?? []) {
    for (const f of listFiles(join(team, '.claude/skills', skill))) paths.add(`.claude/skills/${skill}/${f}`);
  }
  for (const line of overwrites) paths.add(line.slice('replaced  '.length));
  paths.add('.claude/settings.json');
  paths.add(MANIFEST);
  return [...paths].filter((p) => !blockingParent(team, p) && lstatSync(join(team, p), { throwIfNoEntry: false }));
}

/** A failed sync whose undo failed too: the staging folder holds the old files, so it stays. */
class RestoreError extends FolkwaysError {}

/**
 * Move `outgoing` team paths aside and the staged `incoming` paths (manifest last) into place.
 * On any failure, undo every move, so the team repo keeps the old set of files.
 */
function swap(team: string, stage: string, outgoing: string[], incoming: string[]): void {
  const done: { from: string; to: string }[] = [];
  const created: string[] = [];
  const move = (from: string, to: string): void => {
    const made = mkdirSync(dirname(to), { recursive: true });
    if (made) created.push(made);
    fsOps.rename(from, to);
    done.push({ from, to });
  };
  try {
    for (const p of outgoing) move(join(team, p), join(stage, 'old', p));
    for (const p of incoming) move(join(stage, 'new', p), join(team, p));
  } catch (e) {
    try {
      for (const { from, to } of done.reverse()) {
        mkdirSync(dirname(from), { recursive: true });
        fsOps.rename(to, from);
      }
      for (const dir of created.reverse()) pruneEmptyDirs(dir);
    } catch (undo) {
      throw new RestoreError(
        `sync failed and could not restore the old files; they are in ${stage}/old.\n` +
          `${(e as Error).message}\n${(undo as Error).message}`,
      );
    }
    throw e;
  }
}

/** Team overlays: `overlays/<skill>/OVERLAY.md` in the team repo. */
function teamOverlays(team: string): { path: string; skill: string }[] {
  return listFiles(join(team, 'overlays')).flatMap((f) => {
    const skill = /^([^/]+)\/OVERLAY\.md$/.exec(f)?.[1];
    return skill ? [{ path: `overlays/${f}`, skill }] : [];
  });
}

/**
 * Compile each team overlay into the copies of its framework skill (spec §3.2, ADR 0001). Reads only,
 * so an unknown section ID stops the sync before anything is written.
 */
function compileOverlays(team: string, pkg: string, copies: FileCopy[]): { overlays: Record<string, string>; warnings: string[] } {
  const shipped = new Set(skillNames(pkg));
  const overlays: Record<string, string> = {};
  const warnings: string[] = [];
  for (const { path, skill } of teamOverlays(team)) {
    if (!shipped.has(skill)) {
      warnings.push(`${path}: no framework skill named ${skill}; the overlay is not applied`);
      continue;
    }
    const prefix = `.claude/skills/${skill}/`;
    const own = copies.filter((c) => c.dest.startsWith(prefix));
    const files = Object.fromEntries(
      own.filter((c) => c.dest.endsWith('.md')).map((c) => [c.dest.slice(prefix.length), readFileSync(c.src, 'utf8')]),
    );
    let compiled: ReturnType<typeof compileSkillFiles>;
    try {
      compiled = compileSkillFiles(files, readFileSync(join(team, path), 'utf8'));
    } catch (e) {
      if (e instanceof FolkwaysError) throw new FolkwaysError(`${path}: ${e.message}`);
      throw e;
    }
    for (const c of own) c.content = compiled.files[c.dest.slice(prefix.length)] ?? c.content;
    warnings.push(...compiled.warnings.map((w) => `${path}: ${w.replace(/^overlays: /, '')}`));
    overlays[path] = sha256(join(team, path));
  }
  return { overlays, warnings };
}

/** Team overlays changed, added or removed since the last sync compiled them; `doctor` reports these. */
export function staleOverlays(team: string, manifest: Manifest): string[] {
  const recorded = manifest.overlays ?? {};
  const onDisk = teamOverlays(team)
    .filter((o) => manifest.skills.includes(o.skill))
    .map((o) => o.path);
  const paths = new Set([...Object.keys(recorded), ...onDisk]);
  return [...paths].filter((p) => !existsSync(join(team, p)) || recorded[p] !== sha256(join(team, p))).sort();
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
  const fragment = JSON.parse(readFileSync(join(assetsDir(pkg), 'settings.json'), 'utf8')) as Settings;
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
  const leftover = readdirSync(team).find((f) => f.startsWith(STAGE_PREFIX));
  if (leftover) {
    throw new FolkwaysError(
      `a previous sync was interrupted and left ${leftover}/ (the files it replaced are in ${leftover}/old).\n` +
        `Restore the committed files with 'git checkout -- .', delete ${leftover}/, then rerun.`,
    );
  }
  const manifest = readManifest(team);
  const discarded = manifest ? findEdits(team, manifest) : [];
  if (discarded.length > 0 && !options.force) {
    throw new FolkwaysError(editsMessage(discarded, 'Or rerun with --force to discard the edits.'));
  }

  const owned = new Set(manifest?.skills ?? []);
  const skills = skillNames(pkg);
  // Without a manifest (adopting a repo), findOverwrites compares skill folders file by file instead.
  const clashes = manifest ? skills.filter((s) => !owned.has(s) && existsSync(join(team, '.claude/skills', s))) : [];
  if (clashes.length > 0) {
    throw new FolkwaysError(
      `team skills clash with new framework skills: ${clashes.map((s) => `.claude/skills/${s}`).join(', ')}\n` +
        'Rename the team skills, then rerun.',
    );
  }

  const copies = assetMap(pkg);
  const compiled = compileOverlays(team, pkg, copies);
  const overwrites = findOverwrites(team, pkg, manifest, copies);
  if (overwrites.length > 0 && !options.force) {
    throw new FolkwaysError(
      [
        'sync would overwrite team files that differ from this version:',
        ...overwrites.map((o) => `  ${o}`),
        'Move team changes into overlays/, then rerun.',
        'Or rerun with --force to overwrite them.',
      ].join('\n'),
    );
  }

  const settings = mergedSettings(team, pkg);
  const version = packageVersion(pkg);

  // Stage the new files inside the team repo (same filesystem, so renames are atomic), then swap.
  const stage = mkdtempSync(join(team, STAGE_PREFIX));
  try {
    const files: Record<string, string> = {};
    for (const { src, dest, content } of copies) {
      const staged = join(stage, 'new', dest);
      mkdirSync(dirname(staged), { recursive: true });
      if (content === undefined) fsOps.copy(src, staged);
      else writeFileSync(staged, content);
      files[dest] = sha256(staged);
    }
    mkdirSync(join(stage, 'new/.claude'), { recursive: true });
    writeFileSync(join(stage, 'new/.claude/settings.json'), settings);
    mkdirSync(join(stage, 'new/.folkways'), { recursive: true });
    const overlays = Object.keys(compiled.overlays).length > 0 ? { overlays: compiled.overlays } : {};
    writeManifest(join(stage, 'new'), { version, skills, files, ...overlays });

    const incoming = [...copies.map((c) => c.dest), '.claude/settings.json', MANIFEST];
    swap(team, stage, replacedPaths(team, manifest, overwrites, incoming), incoming);
  } catch (e) {
    if (!(e instanceof RestoreError)) rmSync(stage, { recursive: true, force: true });
    throw e;
  }
  rmSync(stage, { recursive: true, force: true });
  for (const skill of manifest?.skills ?? []) pruneEmptyDirs(join(team, '.claude/skills', skill));
  pruneEmptyDirs(join(team, '.folkways'));
  pruneEmptyDirs(join(team, '.github'));

  const warnings = [...compiled.warnings, ...hookWarnings(team)];
  if (!existsSync(join(team, 'folkways.yaml'))) warnings.push('no folkways.yaml; the hooks stay silent until the repo has one.');
  const agents = join(team, 'AGENTS.md');
  const imports = existsSync(agents) && readFileSync(agents, 'utf8').split('\n').some((l) => l.trim() === INSTRUCTIONS_IMPORT);
  if (!imports) warnings.push(`AGENTS.md has no '${INSTRUCTIONS_IMPORT}' line; agents won't see the framework instructions.`);

  return { version, previous: manifest?.version, discarded: [...discarded, ...overwrites], warnings };
}
