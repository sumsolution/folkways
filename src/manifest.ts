import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

export const MANIFEST = '.folkways/manifest.json';

/** What the last sync wrote: version, framework skills, and a checksum per file. */
export interface Manifest {
  version: string;
  skills: string[];
  files: Record<string, string>;
}

export function readManifest(team: string): Manifest | undefined {
  const path = join(team, MANIFEST);
  if (!existsSync(path)) return undefined;
  return JSON.parse(readFileSync(path, 'utf8')) as Manifest;
}

export function writeManifest(team: string, manifest: Manifest): void {
  writeFileSync(join(team, MANIFEST), `${JSON.stringify(manifest, null, 2)}\n`);
}
