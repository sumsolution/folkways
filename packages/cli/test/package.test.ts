import { existsSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { assetsDir, hooksVersion } from '../src/package.ts';
import { REPO, json, put, tmp } from './helpers.ts';

// ADR 0003: the CLI pins one hooks version, in its own package.json, independent of its own version.
describe('hooksVersion', () => {
  it('reads the pinned hooks version', () => {
    expect(hooksVersion()).toBe(json(join(REPO, 'package.json')).folkways.hooks);
  });

  it('pins an exact version the hooks package has reached', () => {
    const pin = hooksVersion().split('.').map(Number);
    const current = (json(join(REPO, '..', 'hooks', 'package.json')).version as string).split('.').map(Number);
    expect(pin).toHaveLength(3);
    expect(pin.every(Number.isInteger)).toBe(true);
    const cmp = pin.map((n, i) => n - current[i]!).find((d) => d !== 0) ?? 0;
    expect(cmp, `pin ${hooksVersion()} is ahead of packages/hooks`).toBeLessThanOrEqual(0);
  });
});

// A packed CLI carries `assets/`; in the workspace the content lives in the sibling `packages/content`.
describe('assetsDir', () => {
  // A local pack leaves a copy in assets/; it must not hide later edits to packages/content.
  it('prefers the sibling content package over a packed copy', () => {
    const pkg = join(tmp(), 'cli');
    put(join(pkg, 'assets/instructions.md'), 'packed\n');
    put(join(pkg, '../content/instructions.md'), 'sibling\n');
    expect(assetsDir(pkg)).toBe(join(pkg, '..', 'content'));
  });

  it('uses the package\'s own assets/ when there is no sibling content', () => {
    const pkg = join(tmp(), 'cli');
    put(join(pkg, 'assets/instructions.md'), 'packed\n');
    expect(assetsDir(pkg)).toBe(join(pkg, 'assets'));
  });

  it('finds the real content from this checkout', () => {
    expect(existsSync(join(assetsDir(), 'instructions.md'))).toBe(true);
  });
});
