import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

/** Root of this package: holds package.json, assets/, templates/. Same from src/ and dist/. */
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
