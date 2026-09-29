import { createHash } from 'node:crypto';
import { copyFileSync, existsSync, readdirSync, readFileSync, renameSync, rmdirSync, statSync } from 'node:fs';
import { join, sep } from 'node:path';

/** File operations sync stages with. A seam, so tests can inject failures. */
export const fsOps = {
  copy: (src: string, dest: string): void => copyFileSync(src, dest),
  rename: (from: string, to: string): void => renameSync(from, to),
};

const IGNORED = new Set(['.gitkeep', '.DS_Store']);

export function sha256(file: string): string {
  return sha256Text(readFileSync(file));
}

export function sha256Text(content: string | Buffer): string {
  return createHash('sha256').update(content).digest('hex');
}

/** Files under `dir`, as sorted POSIX paths relative to it. Empty if `dir` is missing. */
export function listFiles(dir: string): string[] {
  if (!existsSync(dir)) return [];
  return readdirSync(dir, { recursive: true, encoding: 'utf8' })
    .filter((rel) => statSync(join(dir, rel)).isFile())
    .filter((rel) => !IGNORED.has(rel.split(sep).at(-1) ?? ''))
    .map((rel) => rel.split(sep).join('/'))
    .sort();
}

/** Delete empty directories under `dir`, bottom up. Keeps `dir` itself unless it is empty too. */
export function pruneEmptyDirs(dir: string): void {
  if (!existsSync(dir) || !statSync(dir).isDirectory()) return;
  for (const entry of readdirSync(dir)) pruneEmptyDirs(join(dir, entry));
  if (readdirSync(dir).length === 0) rmdirSync(dir);
}
