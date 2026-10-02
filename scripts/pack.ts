// Pack the published packages and check them: shared by smoke.ts and release.ts, so what the smoke test
// installs is what a release publishes.
import { spawnSync } from 'node:child_process';
import { mkdirSync, readFileSync, readdirSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { packageName, TARGETS } from '../packages/hooks/scripts/build.ts';

export const CLI = resolve(import.meta.dirname, '..', 'packages', 'cli');
export const HOOKS = resolve(import.meta.dirname, '..', 'packages', 'hooks');
export const CLAUDE_PLUGIN = resolve(import.meta.dirname, '..', 'packages', 'claude-plugin');
/** Every `@folkways/hooks-<os>-<cpu>` package, sorted. */
export const PLATFORMS: readonly string[] = TARGETS.map(packageName).sort();

export function run(cmd: string, args: string[], cwd: string, env: NodeJS.ProcessEnv = process.env): string {
  const r = spawnSync(cmd, args, { cwd, env, encoding: 'utf8' });
  if (r.error || r.status !== 0) {
    const why = r.error ? r.error.message : `exit ${r.status}`;
    throw new Error(`${cmd} ${args.join(' ')} (in ${cwd}) failed: ${why}\n--- stdout ---\n${r.stdout ?? ''}\n--- stderr ---\n${r.stderr ?? ''}`);
  }
  return r.stdout;
}

export function assert(ok: boolean, message: string): asserts ok {
  if (!ok) throw new Error(message);
}

/**
 * Skipped tests in vitest output, from its `Tests` summary line. Test names can say "skipped" (vitest
 * lists slow tests by name), so only the summary counts.
 */
export function skippedTests(output: string): number {
  const plain = output.replace(/\x1b\[[0-9;]*m/g, '');
  const summary = /^\s*Tests\s+(.*)$/m.exec(plain);
  assert(summary !== null, `no vitest Tests summary in:\n${output}`);
  return Number(/(\d+) skipped/.exec(summary[1]!)?.[1] ?? 0);
}

export const readJson = (path: string) => JSON.parse(readFileSync(path, 'utf8'));
export const tarballs = (dir: string): string[] => readdirSync(dir).filter((f) => f.endsWith('.tgz'));
/** package.json inside a packed tarball. */
export const packedManifest = (tgz: string) => JSON.parse(run('tar', ['-xzOf', tgz, 'package/package.json'], process.cwd()));

export interface HookPackages {
  version: string;
  /** Platform package name → its tarball. */
  tarballs: Map<string, string>;
  /** Platform package name → its built directory (holds bin/folkways-hook). */
  dirs: Map<string, string>;
}

/** Build the `@folkways/hooks-<os>-<cpu>` packages into `build`, and `npm pack` each into `dest`. */
export function packHooks(build: string, dest: string): HookPackages {
  const version: string = readJson(join(HOOKS, 'package.json')).version;
  run(process.execPath, ['scripts/build.ts', build], HOOKS);
  mkdirSync(dest, { recursive: true });
  const result: HookPackages = { version, tarballs: new Map(), dirs: new Map() };
  for (const dir of readdirSync(build).map((d) => join(build, d))) {
    const before = new Set(tarballs(dest));
    run('npm', ['pack', '--pack-destination', dest], dir);
    const added = tarballs(dest).filter((f) => !before.has(f));
    assert(added.length === 1, `npm pack in ${dir} made ${added.length} tarballs`);
    const tgz = join(dest, added[0]!);
    const manifest = packedManifest(tgz);
    assert(manifest.version === version, `${manifest.name} is ${manifest.version}, expected ${version}`);
    result.tarballs.set(manifest.name, tgz);
    result.dirs.set(manifest.name, dir);
  }
  const built = [...result.tarballs.keys()].sort();
  assert(JSON.stringify(built) === JSON.stringify(PLATFORMS), `built ${built}, expected ${PLATFORMS}`);
  return result;
}

export interface CliPackage {
  tarball: string;
  /** The hooks version it pins (`folkways.hooks`). */
  hooks: string;
}

/** `pnpm pack` the CLI into `dest` and check `.pnpmfile.mjs` pinned every platform package to `folkways.hooks`. */
export function packCli(dest: string): CliPackage {
  run('pnpm', ['pack', '--pack-destination', dest], CLI);
  const found = tarballs(dest).filter((f) => f.startsWith('folkways-cli-'));
  assert(found.length === 1, `expected one CLI .tgz in ${dest}, found ${found.length}`);
  const tarball = join(dest, found[0]!);
  const hooks: string = readJson(join(CLI, 'package.json')).folkways.hooks;
  const optional: Record<string, string> = packedManifest(tarball).optionalDependencies ?? {};
  assert(
    JSON.stringify(Object.keys(optional).sort()) === JSON.stringify(PLATFORMS),
    `packed CLI optionalDependencies ${JSON.stringify(optional)} don't match the platforms ${PLATFORMS}; see .pnpmfile.mjs`,
  );
  assert(
    Object.values(optional).every((v) => v === hooks),
    `packed CLI optionalDependencies ${JSON.stringify(optional)} should all pin folkways.hooks ${hooks}`,
  );
  return { tarball, hooks };
}

/**
 * Build `@folkways/claude-plugin` into `build`, with the binaries from the `@folkways/hooks` build in `hooks`,
 * and `npm pack` it into `dest`. Checks the tarball carries the manifest at the package version and an
 * executable folkways-hook per platform, and leaves out the dev marketplace.
 */
export function packClaudePlugin(hooks: string, build: string, dest: string): string {
  const version: string = readJson(join(CLAUDE_PLUGIN, 'package.json')).version;
  run(process.execPath, ['src/build.ts', build, hooks], CLAUDE_PLUGIN);
  mkdirSync(dest, { recursive: true });
  const before = new Set(tarballs(dest));
  run('npm', ['pack', '--pack-destination', dest], build);
  const added = tarballs(dest).filter((f) => !before.has(f));
  assert(added.length === 1, `npm pack in ${build} made ${added.length} tarballs`);
  const tgz = join(dest, added[0]!);
  assert(packedManifest(tgz).version === version, `@folkways/claude-plugin packed as ${packedManifest(tgz).version}, expected ${version}`);
  const plugin = JSON.parse(run('tar', ['-xzOf', tgz, 'package/.claude-plugin/plugin.json'], process.cwd()));
  assert(plugin.version === version, `plugin.json version is ${plugin.version}, expected ${version}`);
  const listing = run('tar', ['-tvzf', tgz], process.cwd());
  assert(!listing.includes('marketplace.json'), 'the packed plugin includes the dev marketplace');
  for (const target of PLATFORMS.map((p) => p.replace('@folkways/hooks-', ''))) {
    const line = listing.split('\n').find((l) => l.endsWith(`package/hooks/bin/folkways-hook-${target}`));
    assert(line !== undefined, `the packed plugin has no folkways-hook for ${target}`);
    assert(/^-rwx/.test(line), `the packed folkways-hook for ${target} isn't executable: ${line}`);
  }
  return tgz;
}
