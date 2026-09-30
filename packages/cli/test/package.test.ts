import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { hooksVersion } from '../src/package.ts';
import { REPO, json } from './helpers.ts';

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
