import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { addDevExactCommand, detectPackageManager, fromUserAgent, installCommand } from '../src/package-manager.ts';
import { put, tmp } from './helpers.ts';

describe('detectPackageManager', () => {
  it('defaults to npm', () => {
    expect(detectPackageManager(tmp())).toBe('npm');
  });

  it('reads the packageManager field first', () => {
    const team = tmp();
    put(join(team, 'package.json'), JSON.stringify({ packageManager: 'pnpm@12.4.1+sha512.abc' }));
    put(join(team, 'package-lock.json'), '{}');
    expect(detectPackageManager(team)).toBe('pnpm');
  });

  it.each([
    ['pnpm-lock.yaml', 'pnpm'],
    ['package-lock.json', 'npm'],
  ])('falls back to %s', (lockfile, pm) => {
    const team = tmp();
    put(join(team, 'package.json'), '{}');
    put(join(team, lockfile), '');
    expect(detectPackageManager(team)).toBe(pm);
  });

  it('rejects an unsupported packageManager', () => {
    const team = tmp();
    put(join(team, 'package.json'), JSON.stringify({ packageManager: 'yarn@4.0.0' }));
    expect(() => detectPackageManager(team)).toThrow(/yarn is not supported; use npm or pnpm/);
  });
});

describe('fromUserAgent', () => {
  it.each([
    ['pnpm/12.4.1 npm/? node/v24.1.0 darwin arm64', 'pnpm'],
    ['npm/11.19.0 node/v24.14.1 darwin arm64 workspaces/false', 'npm'],
    ['yarn/4.0.0 npm/? node/v24.1.0', undefined],
    ['', undefined],
  ])('%s -> %s', (ua, pm) => expect(fromUserAgent(ua)).toBe(pm));
});

describe('commands', () => {
  it('installs from the lockfile', () => {
    expect(installCommand('npm')).toEqual(['npm', ['install']]);
    expect(installCommand('pnpm')).toEqual(['pnpm', ['install']]);
  });

  it('adds an exact dev dependency', () => {
    expect(addDevExactCommand('npm', 'x@1.0.0')).toEqual(['npm', ['install', '--save-dev', '--save-exact', 'x@1.0.0']]);
    expect(addDevExactCommand('pnpm', 'x@1.0.0')).toEqual(['pnpm', ['add', '--save-dev', '--save-exact', 'x@1.0.0']]);
  });
});
