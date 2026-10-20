import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, statSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';
import { build, hostTarget, packageName, platformManifest, TARGETS } from '../scripts/build.ts';

describe('TARGETS', () => {
  it('covers the four supported platforms', () => {
    expect(TARGETS.map(packageName).sort()).toEqual([
      '@folkways/hooks-darwin-arm64',
      '@folkways/hooks-darwin-x64',
      '@folkways/hooks-linux-arm64',
      '@folkways/hooks-linux-x64',
    ]);
  });
});

describe('platformManifest', () => {
  const manifest = platformManifest({ os: 'darwin', cpu: 'x64', goarch: 'amd64' }, '1.2.3');

  it('restricts installs to its platform', () => {
    expect(manifest.os).toEqual(['darwin']);
    expect(manifest.cpu).toEqual(['x64']);
  });

  it('links the binary as folkways-hook, with no install scripts', () => {
    expect(manifest.bin).toEqual({ 'folkways-hook': 'bin/folkways-hook' });
    expect(manifest).not.toHaveProperty('scripts');
  });

  it('carries the version, license and repository for provenance', () => {
    expect(manifest).toMatchObject({
      name: '@folkways/hooks-darwin-x64',
      version: '1.2.3',
      license: 'Apache-2.0',
      repository: { url: 'git+https://github.com/sumsolution/folkways.git', directory: 'packages/folkways' },
      publishConfig: { access: 'public' },
    });
  });
});

describe('build', () => {
  const out = mkdtempSync(join(tmpdir(), 'folkways-hooks-build-'));
  afterAll(() => rmSync(out, { recursive: true, force: true }));

  it('writes a package per target whose host binary runs and reports the version', () => {
    const dirs = build(out, '9.8.7');
    expect(dirs).toHaveLength(TARGETS.length);
    for (const dir of dirs) {
      const manifest = JSON.parse(readFileSync(join(dir, 'package.json'), 'utf8'));
      expect(manifest.version).toBe('9.8.7');
      expect(existsSync(join(dir, 'LICENSE'))).toBe(true);
      expect(existsSync(join(dir, 'README.md'))).toBe(true);
      expect(statSync(join(dir, 'bin', 'folkways-hook')).mode & 0o111).not.toBe(0);
    }
    const host = hostTarget();
    if (!host) return;
    const bin = join(out, `${host.os}-${host.cpu}`, 'bin', 'folkways-hook');
    const r = spawnSync(bin, ['--version'], { encoding: 'utf8' });
    expect(r.stdout.trim()).toBe('9.8.7');
  }, 180_000);
});
