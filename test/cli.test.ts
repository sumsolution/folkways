import { spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { REPO, json, put, tmp } from './helpers.ts';

const cli = (cwd: string, ...args: string[]) =>
  spawnSync(process.execPath, [join(REPO, 'src/cli.ts'), ...args], { cwd, encoding: 'utf8' });

describe('cli', () => {
  it('prints usage', () => {
    const r = cli(REPO, 'help');
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('usage: folkways');
  });

  it('prints the package version', () => {
    expect(cli(REPO, 'version').stdout.trim()).toBe(json(join(REPO, 'package.json')).version);
  });

  it('init then sync', () => {
    const dest = join(tmp(), 'team');
    const r = cli(REPO, 'init', dest, '--no-install');
    expect(r.stderr).toBe('');
    expect(r.stdout).toContain(`Created ${dest}`);
    const s = cli(join(dest, 'knowledge'), 'sync');
    expect(s.status).toBe(0);
    expect(s.stdout).toContain('Synced folkways');
  });

  it('prints expected errors without a stack trace', () => {
    const dest = tmp();
    put(join(dest, 'x'), '');
    const r = cli(REPO, 'init', dest, '--no-install');
    expect(r.status).toBe(1);
    expect(r.stderr).toBe(`folkways: ${dest} exists and is not empty\n`);
  });

  it('rejects unknown options', () => {
    const r = cli(REPO, 'sync', '--nope');
    expect(r.status).toBe(1);
    expect(r.stderr).toMatch(/^folkways: Unknown option '--nope'/);
  });

  it('exits 1 on an unknown command', () => {
    expect(cli(REPO, 'nope').status).toBe(1);
  });
});
