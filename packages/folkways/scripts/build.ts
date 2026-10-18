// Build the per-platform npm packages `@folkways/hooks-<os>-<cpu>` (ADR 0002): cross-compile folkways-hook
// for each target and write its package.json, LICENSE and README next to it.
// Usage: node scripts/build.ts [out-dir]   (default npm/; version from this package's package.json)
import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';

const PACKAGE = resolve(import.meta.dirname, '..');
const LICENSE = resolve(PACKAGE, '..', '..', 'LICENSE');

export interface Target {
  /** npm `os` value; also GOOS. */
  os: 'darwin' | 'linux';
  /** npm `cpu` value. */
  cpu: 'arm64' | 'x64';
  goarch: 'arm64' | 'amd64';
}

export const TARGETS: readonly Target[] = [
  { os: 'darwin', cpu: 'arm64', goarch: 'arm64' },
  { os: 'darwin', cpu: 'x64', goarch: 'amd64' },
  { os: 'linux', cpu: 'arm64', goarch: 'arm64' },
  { os: 'linux', cpu: 'x64', goarch: 'amd64' },
];

export const packageName = (t: Target): string => `@folkways/hooks-${t.os}-${t.cpu}`;

/** The target this machine installs, or undefined on an unsupported platform. */
export const hostTarget = (): Target | undefined =>
  TARGETS.find((t) => t.os === process.platform && t.cpu === process.arch);

export function platformManifest(t: Target, version: string): Record<string, unknown> {
  return {
    name: packageName(t),
    version,
    description: `The folkways-hook binary for ${t.os}-${t.cpu}. Installed by @folkways/cli; not for direct use.`,
    license: 'Apache-2.0',
    repository: {
      type: 'git',
      url: 'git+https://github.com/sumsolution/folkways.git',
      directory: 'packages/folkways',
    },
    os: [t.os],
    cpu: [t.cpu],
    bin: { 'folkways-hook': 'bin/folkways-hook' },
    files: ['bin'],
    publishConfig: { access: 'public' },
  };
}

const readme = (t: Target): string =>
  `# ${packageName(t)}\n\nThe \`folkways-hook\` binary for ${t.os}-${t.cpu}. \`@folkways/cli\` installs it as an optional\n` +
  `dependency; team repos don't depend on it directly. See https://github.com/sumsolution/folkways.\n`;

/** Build every target into `out/<os>-<cpu>/`, replacing what was there. Returns the package directories. */
export function build(out: string, version: string): string[] {
  return TARGETS.map((t) => {
    const dir = join(out, `${t.os}-${t.cpu}`);
    rmSync(dir, { recursive: true, force: true });
    mkdirSync(join(dir, 'bin'), { recursive: true });
    const r = spawnSync(
      'go',
      ['build', '-trimpath', '-ldflags', `-s -w -X main.version=${version}`, '-o', join(dir, 'bin', 'folkways-hook'), './cmd/folkways-hook'],
      { cwd: PACKAGE, env: { ...process.env, CGO_ENABLED: '0', GOOS: t.os, GOARCH: t.goarch }, encoding: 'utf8' },
    );
    if (r.error || r.status !== 0) {
      throw new Error(`go build for ${t.os}-${t.cpu} failed: ${r.error?.message ?? r.stderr}`);
    }
    writeFileSync(join(dir, 'package.json'), `${JSON.stringify(platformManifest(t, version), null, 2)}\n`);
    writeFileSync(join(dir, 'README.md'), readme(t));
    copyFileSync(LICENSE, join(dir, 'LICENSE'));
    return dir;
  });
}

if (import.meta.main) {
  const version = (JSON.parse(readFileSync(join(PACKAGE, 'package.json'), 'utf8')) as { version: string }).version;
  const out = resolve(process.argv[2] ?? join(PACKAGE, 'npm'));
  for (const dir of build(out, version)) console.log(dir);
}
