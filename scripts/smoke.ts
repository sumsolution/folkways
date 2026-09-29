// Smoke test: pack the CLI, install the tarball into a scratch project, and drive it end to end.
// Usage: pnpm run smoke   (KEEP=1 keeps the temp dir)
import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, readFileSync, readdirSync, realpathSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const repo = resolve(import.meta.dirname, '..');
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

function run(cmd: string, args: string[], cwd: string): string {
  const r = spawnSync(cmd, args, { cwd, env, encoding: 'utf8' });
  if (r.error || r.status !== 0) {
    const why = r.error ? r.error.message : `exit ${r.status}`;
    throw new Error(`${cmd} ${args.join(' ')} (in ${cwd}) failed: ${why}\n--- stdout ---\n${r.stdout ?? ''}\n--- stderr ---\n${r.stderr ?? ''}`);
  }
  return r.stdout;
}

function assert(ok: boolean, message: string): void {
  if (!ok) throw new Error(message);
}

try {
  step('pack');
  const packDir = join(root, 'pack');
  run('pnpm', ['pack', '--pack-destination', packDir], repo);
  const tarballs = readdirSync(packDir).filter((f) => f.endsWith('.tgz'));
  assert(tarballs.length === 1, `expected one .tgz in ${packDir}, found ${tarballs.length}`);
  const tarball = join(packDir, tarballs[0]!);

  step('install tarball');
  const project = join(root, 'project');
  mkdirSync(project);
  run('npm', ['init', '-y'], project);
  run('npm', ['install', '--no-audit', '--no-fund', tarball], project);
  const bin = join(project, 'node_modules', '.bin', 'folkways');

  step('init');
  const team = join(root, 'team');
  run(bin, ['init', team, '--no-install', '--package-manager', 'npm'], root);

  step('sync');
  const manifest = () => readFileSync(join(team, '.folkways', 'manifest.json'), 'utf8');
  const status = () => run('git', ['status', '--porcelain'], team);
  const out = run(bin, ['sync'], team);
  assert(out.includes('Synced folkways'), `sync output lacks "Synced folkways":\n${out}`);
  const statusBefore = status();
  const manifestBefore = manifest();
  assert(statusBefore === '', `tree not clean after init + sync:\n${statusBefore}`);

  step('sync again (idempotent)');
  const again = run(bin, ['sync'], team);
  assert(again.includes('Synced folkways'), `second sync output lacks "Synced folkways":\n${again}`);
  assert(status() === statusBefore, `git status changed after second sync:\n${status()}`);
  assert(manifest() === manifestBefore, '.folkways/manifest.json changed after second sync');

  step('version');
  const expected = (JSON.parse(readFileSync(join(repo, 'package.json'), 'utf8')) as { version: string }).version;
  const actual = run(bin, ['version'], team).trim();
  assert(actual === expected, `version is ${actual}, expected ${expected}`);

  console.log('smoke: ok');
} catch (error) {
  console.error(`smoke: FAILED\n${error instanceof Error ? error.message : String(error)}`);
  process.exitCode = 1;
} finally {
  if (process.env.KEEP === '1') console.log(`smoke: kept ${root}`);
  else rmSync(root, { recursive: true, force: true });
}
