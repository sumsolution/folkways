import { existsSync, readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

/** Root of this package: holds package.json, templates/, and assets/ once packed. Same from src/ and dist/. */
export const PACKAGE_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');
export const PACKAGE_NAME = '@folkways/cli';

export function packageVersion(pkg: string = PACKAGE_ROOT): string {
  const json = JSON.parse(readFileSync(join(pkg, 'package.json'), 'utf8')) as { version: string };
  return json.version;
}

/** The exact `@folkways/hooks-<os>-<cpu>` version this CLI pins (`folkways.hooks` in package.json, ADR 0003). */
export function hooksVersion(pkg: string = PACKAGE_ROOT): string {
  const json = JSON.parse(readFileSync(join(pkg, 'package.json'), 'utf8')) as { folkways: { hooks: string } };
  return json.folkways.hooks;
}

/**
 * The framework content (skills, instructions, prompts, settings fragment, workflows). A packed CLI
 * carries it as `assets/` (copied from `packages/content` at pack time); in the workspace it is the
 * sibling `packages/content` (ADR 0003).
 */
export function assetsDir(pkg: string = PACKAGE_ROOT): string {
  const packed = join(pkg, 'assets');
  return existsSync(packed) ? packed : join(pkg, '..', 'content');
}
