// Smoke test: pack @folkways/cli and the @folkways/hooks-<os>-<cpu> packages, install them into scratch team
// repos with npm and with pnpm, and drive them end to end, including the hook contract suite against the
// installed binary (ADR 0002). Nothing comes from the registry: overrides point each platform package at its
// tarball, so the CLI's optional dependencies resolve as they will once published.
// Usage: pnpm run smoke   (KEEP=1 keeps the temp dir; needs Go)
import { appendFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, realpathSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { assert, CLI, HOOKS, packCli, packHooks, readJson, run as runIn } from './pack.ts';

const root = realpathSync(mkdtempSync(join(tmpdir(), 'folkways-smoke-')));
const env = {
  ...process.env,
  GIT_AUTHOR_NAME: 'smoke',
  GIT_AUTHOR_EMAIL: 'smoke@example.com',
  GIT_COMMITTER_NAME: 'smoke',
  GIT_COMMITTER_EMAIL: 'smoke@example.com',
  GIT_CONFIG_GLOBAL: '/dev/null',
  GIT_CONFIG_NOSYSTEM: '1',
};

function step(name: string): void {
  console.log(`smoke: ${name}`);
}

const run = (cmd: string, args: string[], cwd: string, extra: Record<string, string> = {}): string =>
  runIn(cmd, args, cwd, { ...env, ...extra });

try {
  step('build and pack hook packages');
  const hooks = packHooks(join(root, 'hooks-npm'), join(root, 'pack'));
  /** Each platform package → `file:` its tarball, for npm or pnpm overrides. */
  const overrides = Object.fromEntries([...hooks.tarballs].map(([name, tgz]) => [name, `file:${tgz}`]));

  step('pack cli');
  const cli = packCli(join(root, 'pack'));

  step('install cli tarball');
  const project = join(root, 'project');
  mkdirSync(project);
  writeFileSync(join(project, 'package.json'), JSON.stringify({ name: 'project', private: true, overrides }, null, 2));
  run('npm', ['install', '--no-audit', '--no-fund', cli.tarball], project);
  const bin = join(project, 'node_modules', '.bin', 'folkways');
  const installed = join(project, 'node_modules', '@folkways', 'cli');
  assert(existsSync(join(installed, 'LICENSE')), `the tarball has no LICENSE (${installed})`);
  const license = readJson(join(installed, 'package.json')).license;
  assert(license === 'Apache-2.0', `package.json license is ${license}, expected Apache-2.0`);

  step('version');
  const expected: string = readJson(join(CLI, 'package.json')).version;
  const actual = run(bin, ['version'], project).trim();
  assert(actual === expected, `version is ${actual}, expected ${expected}`);

  for (const pm of ['npm', 'pnpm'] as const) {
    step(`${pm}: init`);
    const team = join(root, `team-${pm}`);
    run(bin, ['init', team, '--no-install', '--package-manager', pm], root);

    step(`${pm}: install from tarballs`);
    const manifestPath = join(team, 'package.json');
    const manifest = readJson(manifestPath);
    manifest.devDependencies['@folkways/cli'] = `file:${cli.tarball}`;
    if (pm === 'npm') manifest.overrides = overrides;
    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
    if (pm === 'pnpm') {
      const lines = Object.entries(overrides).map(([name, spec]) => `  "${name}": "${spec}"`);
      appendFileSync(join(team, 'pnpm-workspace.yaml'), `overrides:\n${lines.join('\n')}\n`);
      run('pnpm', ['install'], team);
    } else {
      run('npm', ['install', '--no-audit', '--no-fund'], team);
    }

    step(`${pm}: hook binary is linked`);
    const hookBin = join(team, 'node_modules', '.bin', 'folkways-hook');
    assert(existsSync(hookBin), `${hookBin} is missing after ${pm} install`);
    const hookVersion = run(hookBin, ['--version'], team).trim();
    assert(hookVersion === hooks.version, `folkways-hook --version is ${hookVersion}, expected ${hooks.version}`);
    const lockfile = readFileSync(join(team, pm === 'pnpm' ? 'pnpm-lock.yaml' : 'package-lock.json'), 'utf8');
    // A package entry, not the overrides or the CLI's optionalDependencies, which name every platform anyway.
    const entry = (name: string) =>
      pm === 'pnpm' ? new RegExp(`^\\s+'?${name}@`, 'm').test(lockfile) : lockfile.includes(`"node_modules/${name}"`);
    const missing = [...hooks.tarballs.keys()].filter((name) => !entry(name));
    assert(missing.length === 0, `${pm}'s lockfile drops ${missing.join(', ')}; other platforms would get no hook binary`);
    run('git', ['add', '-A'], team);
    run('git', ['commit', '-q', '-m', 'install from tarballs'], team);

    step(`${pm}: sync`);
    const teamBin = join(team, 'node_modules', '.bin', 'folkways');
    const status = () => run('git', ['status', '--porcelain'], team);
    const manifestJson = () => readFileSync(join(team, '.folkways', 'manifest.json'), 'utf8');
    const out = run(teamBin, ['sync'], team);
    assert(out.includes('Synced folkways'), `sync output lacks "Synced folkways":\n${out}`);
    assert(!out.includes('warning:'), `sync warned after a full install:\n${out}`);
    const statusBefore = status();
    const manifestBefore = manifestJson();
    assert(statusBefore === '', `tree not clean after init + install + sync:\n${statusBefore}`);

    step(`${pm}: sync again (idempotent)`);
    const again = run(teamBin, ['sync'], team);
    assert(again.includes('Synced folkways'), `second sync output lacks "Synced folkways":\n${again}`);
    assert(status() === statusBefore, `git status changed after second sync:\n${status()}`);
    assert(manifestJson() === manifestBefore, '.folkways/manifest.json changed after second sync');

    step(`${pm}: hook contract suite against the installed binary`);
    const suite = run('pnpm', ['exec', 'vitest', 'run', 'contract'], HOOKS, { FOLKWAYS_HOOK_CMD: hookBin });
    assert(!/skipped/.test(suite), `the contract suite skipped cases:\n${suite}`);
  }

  console.log('smoke: ok');
} catch (error) {
  console.error(`smoke: FAILED\n${error instanceof Error ? error.message : String(error)}`);
  process.exitCode = 1;
} finally {
  if (process.env.KEEP === '1') console.log(`smoke: kept ${root}`);
  else rmSync(root, { recursive: true, force: true });
}
