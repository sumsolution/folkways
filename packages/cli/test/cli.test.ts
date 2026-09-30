import { spawnSync } from 'node:child_process';
import { chmodSync } from 'node:fs';
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

  it('hook runs the installed binary with the payload on stdin', () => {
    const dest = join(tmp(), 'team');
    cli(REPO, 'init', dest, '--no-install');
    put(join(dest, 'node_modules/.bin/folkways-hook'), '#!/bin/sh\necho "ran $1 with $(cat)"\nexit 3\n');
    chmodSync(join(dest, 'node_modules/.bin/folkways-hook'), 0o755);
    const r = spawnSync(process.execPath, [join(REPO, 'src/cli.ts'), 'hook', 'git-guard'], {
      cwd: join(dest, 'knowledge'),
      input: '{"x":1}',
      encoding: 'utf8',
    });
    expect(r.stdout).toBe('ran git-guard with {"x":1}\n');
    expect(r.status).toBe(3);
  });

  it('hook explains a missing binary', () => {
    const dest = join(tmp(), 'team');
    cli(REPO, 'init', dest, '--no-install');
    const r = cli(dest, 'hook', 'git-guard');
    expect(r.status).toBe(1);
    expect(r.stderr).toMatch(/^folkways: node_modules\/\.bin\/folkways-hook is missing/);
    expect(cli(dest, 'hook').stderr).toMatch(/usage: folkways hook <name>/);
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
