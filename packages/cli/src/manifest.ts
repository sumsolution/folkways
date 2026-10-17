import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { FolkwaysError } from './errors.ts';

export const MANIFEST = '.folkways/manifest.json';

/** What the last sync wrote: version, framework skills, a checksum per file, and per team overlay it compiled. */
export interface Manifest {
  version: string;
  skills: string[];
  files: Record<string, string>;
  /** `overlays/<skill>/OVERLAY.md` → checksum at the last sync. Absent when no overlay was compiled. */
  overlays?: Record<string, string>;
}

const SKILL_NAME = /^[A-Za-z0-9][A-Za-z0-9._-]*$/;
const SHA256 = /^[a-f0-9]{64}$/;
const OVERLAY_PATH = /^overlays\/[A-Za-z0-9][A-Za-z0-9._-]*\/OVERLAY\.md$/;

/**
 * Whether `path` is a plain relative path in a framework-owned location.
 * sync deletes these, so a tampered manifest must not reach outside them.
 */
export function isOwnedPath(path: string, skills: string[]): boolean {
  const parts = path.split('/');
  if (path.includes('\\') || parts.some((p) => p === '' || p === '.' || p === '..')) return false;
  if (parts[0] === '.folkways') return parts.length > 1;
  if (/^\.github\/workflows\/folkways-[^/]+\.yml$/.test(path)) return true;
  return parts[0] === '.claude' && parts[1] === 'skills' && parts.length > 3 && skills.includes(parts[2]!);
}

function validate(value: unknown): Manifest {
  const invalid = (why: string) =>
    new FolkwaysError(`${MANIFEST}: ${why}. Restore it with 'git checkout -- ${MANIFEST}'.`);
  if (typeof value !== 'object' || value === null) throw invalid('not a JSON object');
  const m = value as Partial<Manifest>;
  if (typeof m.version !== 'string') throw invalid('missing version');
  if (!Array.isArray(m.skills)) throw invalid('missing skills');
  for (const skill of m.skills) {
    if (typeof skill !== 'string' || !SKILL_NAME.test(skill)) throw invalid(`unsafe skill name ${JSON.stringify(skill)}`);
  }
  if (typeof m.files !== 'object' || m.files === null || Array.isArray(m.files)) throw invalid('missing files');
  for (const [path, hash] of Object.entries(m.files)) {
    if (!isOwnedPath(path, m.skills)) throw invalid(`unsafe path ${JSON.stringify(path)}`);
    if (typeof hash !== 'string' || !SHA256.test(hash)) throw invalid(`bad checksum for ${path}`);
  }
  if (m.overlays !== undefined) {
    if (typeof m.overlays !== 'object' || m.overlays === null || Array.isArray(m.overlays)) throw invalid('bad overlays');
    for (const [path, hash] of Object.entries(m.overlays)) {
      if (!OVERLAY_PATH.test(path)) throw invalid(`unsafe overlay path ${JSON.stringify(path)}`);
      if (typeof hash !== 'string' || !SHA256.test(hash)) throw invalid(`bad checksum for ${path}`);
    }
  }
  return m as Manifest;
}

export function readManifest(team: string): Manifest | undefined {
  const path = join(team, MANIFEST);
  if (!existsSync(path)) return undefined;
  let value: unknown;
  try {
    value = JSON.parse(readFileSync(path, 'utf8'));
  } catch {
    value = undefined;
  }
  return validate(value);
}

export function writeManifest(team: string, manifest: Manifest): void {
  writeFileSync(join(team, MANIFEST), `${JSON.stringify(manifest, null, 2)}\n`);
}
