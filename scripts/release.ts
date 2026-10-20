// Build and check the tarballs a release tag publishes (ADR 0002, ADR 0003). The release workflow runs this,
// uploads <out>/*.tgz, and a separate job publishes them after approval. Nothing here publishes.
//   @folkways/hooks@x.y.z → the four @folkways/hooks-<os>-<cpu> packages; the contract suite runs against
//                           the host's binary
//   @folkways/cli@x.y.z   → @folkways/cli; every platform package it pins must already be on npm
//   @folkways/claude-plugin@x.y.z → the Claude Code plugin, with folkways-hook built from this commit
// Usage: node scripts/release.ts <tag> <out>
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { hostTarget, packageName } from '../packages/folkways/scripts/build.ts';
import {
  assert,
  CLAUDE_PLUGIN,
  CLI,
  HOOKS,
  packClaudePlugin,
  packCli,
  packHooks,
  PLATFORMS,
  readJson,
  run,
  skippedTests,
  tarballs,
} from './pack.ts';

const PACKAGES = { '@folkways/hooks': HOOKS, '@folkways/cli': CLI, '@folkways/claude-plugin': CLAUDE_PLUGIN } as const;
type Released = keyof typeof PACKAGES;

/** `@folkways/cli@1.2.3` → the package and version. Throws on a tag that publishes nothing. */
function parseTag(tag: string): { name: Released; version: string } {
  const m = /^(@folkways\/[a-z-]+)@(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)$/.exec(tag);
  if (!m || !(m[1]! in PACKAGES)) {
    throw new Error(`${tag} is not a release tag; expected ${Object.keys(PACKAGES).map((p) => `${p}@x.y.z`).join(' or ')}`);
  }
  return { name: m[1] as Released, version: m[2]! };
}

/** Whether `name@version` is on the registry. */
function published(name: string, version: string): boolean {
  try {
    return run('npm', ['view', `${name}@${version}`, 'version'], process.cwd()).trim() === version;
  } catch {
    return false;
  }
}

function releaseHooks(out: string): void {
  const build = mkdtempSync(join(tmpdir(), 'folkways-release-'));
  try {
    const hooks = packHooks(build, out);
    const host = hostTarget();
    assert(host !== undefined, `no folkways-hook target for ${process.platform}-${process.arch}`);
    const bin = join(hooks.dirs.get(packageName(host))!, 'bin', 'folkways-hook');
    const suite = run('pnpm', ['exec', 'vitest', 'run', 'contract'], HOOKS, { ...process.env, FOLKWAYS_HOOK_CMD: bin });
    assert(skippedTests(suite) === 0, `the contract suite skipped cases:\n${suite}`);
  } finally {
    rmSync(build, { recursive: true, force: true });
  }
}

function releaseCli(out: string): void {
  const cli = packCli(out);
  const missing = PLATFORMS.filter((name) => !published(name, cli.hooks));
  assert(
    missing.length === 0,
    `@folkways/cli pins hooks ${cli.hooks}, but npm lacks ${missing.join(', ')}. Release @folkways/hooks@${cli.hooks} first.`,
  );
}

function releaseClaudePlugin(out: string): void {
  const build = mkdtempSync(join(tmpdir(), 'folkways-release-plugin-'));
  try {
    run(process.execPath, ['scripts/build.ts', join(build, 'hooks')], HOOKS);
    packClaudePlugin(join(build, 'hooks'), join(build, 'plugin'), out);
  } finally {
    rmSync(build, { recursive: true, force: true });
  }
}

if (import.meta.main) {
  const [tag, outArg] = process.argv.slice(2);
  try {
    assert(tag !== undefined && outArg !== undefined, 'usage: node scripts/release.ts <tag> <out>');
    const { name, version } = parseTag(tag);
    const actual: string = readJson(join(PACKAGES[name], 'package.json')).version;
    assert(actual === version, `tag ${tag} doesn't match ${name}'s package.json version ${actual}`);
    const out = resolve(outArg);
    mkdirSync(out, { recursive: true });
    assert(tarballs(out).length === 0, `${out} already holds tarballs`);
    if (name === '@folkways/hooks') releaseHooks(out);
    else if (name === '@folkways/cli') releaseCli(out);
    else releaseClaudePlugin(out);
    console.log(`release: ${tag} ready:\n${tarballs(out).map((f) => `  ${join(out, f)}`).join('\n')}`);
  } catch (error) {
    console.error(`release: FAILED\n${error instanceof Error ? error.message : String(error)}`);
    process.exitCode = 1;
  }
}
