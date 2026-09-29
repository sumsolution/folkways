// Hook runtime benchmark. Usage: node bench/hooks/run.ts [--runs 50] [--only id,id] [--keep]
// Builds every candidate whose toolchain is installed, installs them into throwaway npm and pnpm team repos,
// times each call path, and prints Markdown tables. See README.md.
import { spawnSync, type SpawnSyncReturns } from 'node:child_process';
import {
  chmodSync, copyFileSync, cpSync, existsSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync,
  realpathSync, rmSync, statSync, writeFileSync,
} from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import { arch, cpus, platform, release, tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { parseArgs } from 'node:util';

const HERE = import.meta.dirname;
const BUILD = join(HERE, '.build');
const FIXTURE = join(HERE, 'fixture');
const PAYLOAD = readFileSync(join(FIXTURE, 'payload.json'), 'utf8');
const EXPECTED = JSON.parse(readFileSync(join(FIXTURE, 'expected.json'), 'utf8'));
const PLATFORM = `${platform()}-${arch()}`;
const IS_MAC = platform() === 'darwin';

const { values: opts } = parseArgs({
  options: {
    runs: { type: 'string', default: '50' },
    only: { type: 'string' },
    keep: { type: 'boolean', default: false },
  },
});
const RUNS = Number(opts.runs);
const ONLY = opts.only ? new Set(opts.only.split(',')) : undefined;

// ---------- helpers ----------

function sh(cmd: string, args: string[], cwd = HERE, env: NodeJS.ProcessEnv = {}): SpawnSyncReturns<string> {
  const r = spawnSync(cmd, args, { cwd, encoding: 'utf8', env: { ...process.env, ...env }, maxBuffer: 64 << 20 });
  if (r.error || r.status !== 0) {
    const out = `${r.stdout ?? ''}${r.stderr ?? ''}`.trim().split('\n').slice(-15).join('\n');
    throw new Error(`${cmd} ${args.join(' ')} failed (${r.error?.message ?? `exit ${r.status}`})\n${out}`);
  }
  return r;
}

function version(cmd: string, args = ['--version']): string | undefined {
  const r = spawnSync(cmd, args, { encoding: 'utf8' });
  return r.status === 0 ? (r.stdout || r.stderr).trim().split('\n')[0] : undefined;
}

function timed<T>(fn: () => T): [T, number] {
  const t = process.hrtime.bigint();
  const v = fn();
  return [v, Number(process.hrtime.bigint() - t) / 1e6];
}

const pct = (sorted: number[], p: number): number => sorted[Math.max(0, Math.ceil((p / 100) * sorted.length) - 1)] ?? NaN;
const ms = (n: number | undefined): string => (n === undefined || Number.isNaN(n) ? '–' : n.toFixed(1));
const mb = (bytes: number | undefined): string => (bytes === undefined ? '–' : (bytes / 1e6).toFixed(1));

function dirSize(path: string): number {
  const s = lstatSync(path);
  if (!s.isDirectory()) return s.size;
  return readdirSync(path).reduce((n, f) => n + dirSize(join(path, f)), 0);
}

function withShebang(src: string, dest: string): void {
  writeFileSync(dest, `#!/usr/bin/env node\n${readFileSync(src, 'utf8')}`);
  chmodSync(dest, 0o755);
}

// ---------- toolchains ----------

const tools = {
  node: process.version,
  npm: version('npm'),
  pnpm: version('pnpm'),
  go: version('go', ['version']),
  cargo: version('cargo'),
  bun: version('bun'),
  deno: version('deno'),
};
const notes: string[] = [];

// ---------- candidates ----------

interface Candidate {
  id: string;
  /** How to run it directly: [command, ...args]. */
  direct: string[];
  /** Files that make up what ships (for size). */
  size?: number;
  /** Native executable to put in a per-platform package (the proposed distribution). */
  exe?: string;
  /** JS file to expose as a `.bin` entry (Node candidates). */
  jsBin?: string;
  buildMs?: number;
}
const candidates: Candidate[] = [];
const want = (id: string): boolean => !ONLY || ONLY.has(id);

function build(id: string, fn: () => Omit<Candidate, 'id' | 'buildMs'>): void {
  if (!want(id)) return;
  try {
    const [c, t] = timed(fn);
    candidates.push({ id, ...c, buildMs: t });
    console.error(`built ${id} (${ms(t)} ms)`);
  } catch (e) {
    notes.push(`\`${id}\` skipped: ${(e as Error).message.split('\n')[0]}`);
    console.error(`skipped ${id}: ${(e as Error).message}`);
  }
}

rmSync(BUILD, { recursive: true, force: true });
mkdirSync(BUILD, { recursive: true });
const JS = join(BUILD, 'js');

const jsWanted = ['node', 'node-bundle', 'node-cache', 'node-snapshot', 'node-sea', 'bun', 'bun-compile', 'deno-compile'].some(want);
if (jsWanted) {
  cpSync(join(HERE, 'js'), JS, { recursive: true, filter: (p) => !p.includes('node_modules') });
  // No install scripts: esbuild and postject work without theirs (esbuild falls back to its JS launcher).
  sh('npm', ['ci', '--ignore-scripts', '--no-audit', '--no-fund', '--loglevel=error'], JS);
  for (const f of ['hook-lib', 'hook']) {
    const src = readFileSync(join(JS, `${f}.ts`), 'utf8').replace(/\.ts'/g, ".mjs'");
    writeFileSync(join(JS, `${f}.mjs`), stripTypeScriptTypes(src));
  }
  const esbuild = join(JS, 'node_modules/.bin/esbuild');
  const bundle = (entry: string, out: string, format: string): void =>
    void sh(esbuild, [
      entry, '--bundle', '--platform=node', '--target=node24', `--format=${format}`, `--outfile=${out}`, '--log-level=warning',
      // `yaml` resolves to its CommonJS build under the `node` condition; give the ESM bundle a real `require`.
      ...(format === 'esm' ? ["--banner:js=import { createRequire } from 'node:module'; const require = createRequire(import.meta.url);"] : []),
    ], JS);
  bundle('hook.ts', 'hook.bundle.mjs', 'esm');
  bundle('snapshot.ts', 'hook.snapshot.cjs', 'cjs');
  writeFileSync(join(JS, 'entry.mjs'), "import module from 'node:module';\nmodule.enableCompileCache?.();\nawait import('./hook.bundle.mjs');\n");
}

build('node', () => ({ direct: [process.execPath, join(JS, 'hook.mjs')], jsBin: 'hook.mjs' }));
build('node-bundle', () => ({
  direct: [process.execPath, join(JS, 'hook.bundle.mjs')], jsBin: 'hook.bundle.mjs', size: statSync(join(JS, 'hook.bundle.mjs')).size,
}));
build('node-cache', () => ({ direct: [process.execPath, join(JS, 'entry.mjs')], jsBin: 'entry.mjs' }));
build('node-snapshot', () => {
  const blob = join(JS, 'hook.blob');
  sh(process.execPath, ['--snapshot-blob', blob, '--build-snapshot', 'hook.snapshot.cjs'], JS);
  return { direct: [process.execPath, '--snapshot-blob', blob], size: statSync(blob).size };
});
build('node-sea', () => {
  const exe = join(JS, 'hook-sea');
  writeFileSync(join(JS, 'sea.json'), JSON.stringify({
    main: 'hook.snapshot.cjs', output: 'sea.blob', disableExperimentalSEAWarning: true, useCodeCache: true,
  }));
  sh(process.execPath, ['--experimental-sea-config', 'sea.json'], JS);
  copyFileSync(process.execPath, exe);
  chmodSync(exe, 0o755);
  if (IS_MAC) sh('codesign', ['--remove-signature', exe]);
  sh(join(JS, 'node_modules/.bin/postject'), [
    exe, 'NODE_SEA_BLOB', 'sea.blob', '--sentinel-fuse', 'NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2',
    ...(IS_MAC ? ['--macho-segment-name', 'NODE_SEA'] : []),
  ], JS);
  if (IS_MAC) sh('codesign', ['--sign', '-', exe]);
  return { direct: [exe], exe, size: statSync(exe).size };
});
if (tools.bun) {
  build('bun', () => ({ direct: ['bun', join(JS, 'hook.ts')] }));
  build('bun-compile', () => {
    const exe = join(JS, 'hook-bun');
    sh('bun', ['build', '--compile', '--minify', 'hook.ts', '--outfile', exe], JS);
    return { direct: [exe], exe, size: statSync(exe).size };
  });
} else notes.push('Bun not installed: `bun`, `bun-compile` skipped.');
if (tools.deno) {
  build('deno-compile', () => {
    const exe = join(JS, 'hook-deno');
    sh('deno', ['compile', '--quiet', '--no-check', '--allow-read', '--allow-env', '--output', exe, 'hook.ts'], JS);
    return { direct: [exe], exe, size: statSync(exe).size };
  });
} else notes.push('Deno not installed: `deno-compile` skipped.');
if (tools.go) {
  build('go', () => {
    const exe = join(BUILD, 'go', 'folkways-bench-hook');
    sh('go', ['build', '-trimpath', '-ldflags=-s -w', '-o', exe, '.'], join(HERE, 'go'), { CGO_ENABLED: '0' });
    return { direct: [exe], exe, size: statSync(exe).size };
  });
} else notes.push('Go not installed: `go` skipped.');
if (tools.cargo) {
  build('rust', () => {
    const target = join(BUILD, 'rust-target');
    sh('cargo', ['build', '--release', '--quiet', '--target-dir', target], join(HERE, 'rust'));
    const exe = join(target, 'release', 'folkways-bench-hook');
    return { direct: [exe], exe, size: statSync(exe).size };
  });
} else notes.push('Rust (cargo) not installed: `rust` skipped.');

// ---------- packages and team repos ----------

// Local tarballs mimic the published layout:
//   bench-hook-node           one package, `.bin` entries point at JS files (today's design)
//   bench-hook-<id>-<plat>    one package per platform, `os`/`cpu` set, `bin` = the native executable
//   bench-hook-meta           the main package: lists every platform package as an optional dependency (proposed
//                             design). npm hoists the platform package and links its bin; pnpm does with
//                             `publicHoistPattern` in the team's pnpm-workspace.yaml
//   bench-hook-shim           esbuild/Biome pattern: `.bin` points at a JS launcher that spawns the platform binary
const PKGS = join(BUILD, 'pkgs');
mkdirSync(PKGS, { recursive: true });
const tarballs: Record<string, string> = {};
const tarSizes: Record<string, number> = {};

function pack(dir: string, manifest: Record<string, unknown>): void {
  writeFileSync(join(dir, 'package.json'), JSON.stringify({ version: '1.0.0', license: 'UNLICENSED', ...manifest }, null, 2));
  const r = sh('npm', ['pack', '--silent', '--pack-destination', PKGS], dir);
  const file = join(PKGS, r.stdout.trim().split('\n').pop()!);
  tarballs[manifest.name as string] = file;
  tarSizes[manifest.name as string] = statSync(file).size;
}

const jsBins = candidates.filter((c) => c.jsBin);
if (jsBins.length > 0) {
  const dir = join(BUILD, 'pkg-node');
  mkdirSync(dir);
  const bin: Record<string, string> = {};
  for (const f of ['hook-lib.mjs', 'hook.bundle.mjs']) copyFileSync(join(JS, f), join(dir, f));
  for (const c of jsBins) {
    const file = `bin-${c.id}.mjs`;
    const src = c.jsBin === 'entry.mjs' ? join(JS, 'entry.mjs') : join(JS, c.jsBin!);
    withShebang(src, join(dir, file));
    bin[`bench-hook-${c.id}`] = file;
  }
  pack(dir, { name: 'bench-hook-node', type: 'module', bin, dependencies: { yaml: '2.9.0' } });
}

const natives = candidates.filter((c) => c.exe);
const shimFor = natives.filter((c) => c.id === 'go' || c.id === 'rust');
for (const c of natives) {
  for (const plat of [PLATFORM, 'aix-ppc64']) {
    // The aix package never installs here; it proves a same-named bin in a skipped platform package is harmless.
    const name = `bench-hook-${c.id}-${plat}`;
    const dir = join(BUILD, `pkg-${c.id}-${plat}`);
    mkdirSync(join(dir, 'bin'), { recursive: true });
    copyFileSync(c.exe!, join(dir, 'bin', 'hook'));
    chmodSync(join(dir, 'bin', 'hook'), 0o755);
    const [os, cpu] = plat.split('-');
    pack(dir, { name, os: [os], cpu: [cpu], bin: { [`bench-hook-${c.id}`]: 'bin/hook' } });
  }
}
if (natives.length > 0) {
  const dir = join(BUILD, 'pkg-meta');
  mkdirSync(dir);
  const optionalDependencies = Object.fromEntries(
    Object.entries(tarballs).filter(([name]) => /-(?:linux|darwin|aix)-/.test(name)).map(([name, file]) => [name, `file:${file}`]),
  );
  pack(dir, { name: 'bench-hook-meta', optionalDependencies });
}
if (shimFor.length > 0) {
  const dir = join(BUILD, 'pkg-shim');
  mkdirSync(dir);
  const bin: Record<string, string> = {};
  for (const c of shimFor) {
    // Same work as esbuild's lib/npm/node-shim + node-platform: resolve the platform package, spawn it.
    const file = `shim-${c.id}.mjs`;
    writeFileSync(join(dir, file), [
      '#!/usr/bin/env node',
      "import { spawnSync } from 'node:child_process';",
      "import { createRequire } from 'node:module';",
      'const require = createRequire(import.meta.url);',
      `const exe = require.resolve(\`bench-hook-${c.id}-\${process.platform}-\${process.arch}/bin/hook\`, { paths: [process.cwd()] });`,
      "const r = spawnSync(exe, process.argv.slice(2), { stdio: 'inherit' });",
      'process.exitCode = r.status ?? 1;',
      '',
    ].join('\n'));
    chmodSync(join(dir, file), 0o755);
    bin[`bench-hook-shim-${c.id}`] = file;
  }
  pack(dir, { name: 'bench-hook-shim', type: 'module', bin });
}

interface Team { pm: 'npm' | 'pnpm'; dir: string; installMs?: number; binKind?: string; error?: string }
const teams: Team[] = [];
const TEAMS = realpathSync(mkdtempSync(join(tmpdir(), 'folkways-bench-')));

function makeTeam(pm: 'npm' | 'pnpm'): Team {
  // Outside the repo, so pnpm doesn't pick up this repo's pnpm-workspace.yaml.
  const dir = join(TEAMS, `team-${pm}`);
  mkdirSync(dir);
  copyFileSync(join(FIXTURE, 'team.yaml'), join(dir, 'team.yaml'));
  const deps = Object.fromEntries(
    Object.entries(tarballs).filter(([name]) => !/-(?:linux|darwin|aix)-/.test(name)).map(([name, file]) => [name, `file:${file}`]),
  );
  writeFileSync(join(dir, 'package.json'), JSON.stringify({ name: 'example-team', private: true, dependencies: deps }, null, 2));
  if (pm === 'pnpm') writeFileSync(join(dir, 'pnpm-workspace.yaml'), 'publicHoistPattern:\n  - "bench-hook-*"\n');
  const team: Team = { pm, dir };
  try {
    const args = pm === 'npm' ? ['install', '--no-audit', '--no-fund', '--loglevel=error'] : ['install', '--reporter=silent'];
    [, team.installMs] = timed(() => sh(pm, args, dir));
    const probe = join(dir, 'node_modules/.bin', natives[0] ? `bench-hook-${natives[0].id}` : 'bench-hook-node');
    if (existsSync(probe)) team.binKind = lstatSync(probe).isSymbolicLink() ? 'symlink' : 'shim script';
  } catch (e) {
    team.error = (e as Error).message.split('\n')[0];
    notes.push(`${pm} install failed: ${team.error}`);
  }
  return team;
}
teams.push(makeTeam('npm'));
if (tools.pnpm) teams.push(makeTeam('pnpm'));
else notes.push('pnpm not installed: `pnpm-bin` paths skipped.');

// Per-candidate install: a team repo with only that candidate's platform package (or the JS package).
interface Install { candidate: string; npm?: number; pnpm?: number }
const installs: Install[] = [];
for (const c of [...(tarballs['bench-hook-node'] ? [{ id: 'node (JS)', pkg: 'bench-hook-node' }] : []),
  ...natives.map((n) => ({ id: n.id, pkg: `bench-hook-${n.id}-${PLATFORM}` }))]) {
  const row: Install = { candidate: c.id };
  for (const pm of ['npm', 'pnpm'] as const) {
    if (pm === 'pnpm' && !tools.pnpm) continue;
    const dir = join(TEAMS, `install-${c.pkg}-${pm}`);
    mkdirSync(dir);
    writeFileSync(join(dir, 'package.json'), JSON.stringify({ name: 'example-team', private: true, dependencies: { [c.pkg]: `file:${tarballs[c.pkg]}` } }));
    const args = pm === 'npm' ? ['install', '--no-audit', '--no-fund', '--loglevel=error'] : ['install', '--reporter=silent'];
    try { [, row[pm]] = timed(() => sh(pm, args, dir)); } catch (e) { notes.push(`${pm} install of ${c.pkg} failed: ${(e as Error).message.split('\n')[0]}`); }
  }
  installs.push(row);
}

// ---------- call paths ----------

interface Path {
  candidate: string; path: string; cmd: string[]; cwd: string; extraEnv?: Record<string, string>; times: number[]; first?: number; error?: string;
}
const paths: Path[] = [];
const benchCwd = teams.find((t) => !t.error)?.dir ?? FIXTURE;
paths.push({ candidate: 'floor', path: 'spawn `true`', cmd: ['true'], cwd: benchCwd, times: [] });
for (const c of candidates) {
  paths.push({ candidate: c.id, path: 'direct', cmd: [...c.direct, 'hook', 'git-guard'], cwd: benchCwd, times: [] });
  for (const t of teams) {
    if (t.error) continue;
    const bin = (name: string): string[] => [join(t.dir, 'node_modules/.bin', name), 'hook', 'git-guard'];
    if (c.jsBin || c.exe) paths.push({ candidate: c.id, path: `${t.pm}-bin`, cmd: bin(`bench-hook-${c.id}`), cwd: t.dir, times: [] });
    if (shimFor.includes(c)) paths.push({ candidate: c.id, path: `${t.pm}-shim`, cmd: bin(`bench-hook-shim-${c.id}`), cwd: t.dir, times: [] });
    if (c.id === 'go' || c.id === 'rust') {
      // The proposed committed command as the hook runners start it: through `sh -c`, anchored at the project dir.
      const command = `"\${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/bench-hook-${c.id} hook git-guard`;
      paths.push({
        candidate: c.id, path: `${t.pm}-bin, sh -c`, cmd: ['/bin/sh', '-c', command], cwd: t.dir,
        extraEnv: { CLAUDE_PROJECT_DIR: t.dir }, times: [],
      });
    }
  }
}

function once(p: Path): number {
  const [cmd, ...args] = p.cmd;
  const [r, t] = timed(() => spawnSync(cmd!, args, { cwd: p.cwd, env: { ...process.env, ...p.extraEnv }, input: PAYLOAD, encoding: 'utf8' }));
  if (p.candidate !== 'floor') {
    if (r.error || r.status !== 0) throw new Error(r.error?.message ?? `exit ${r.status}: ${r.stderr.trim().split('\n')[0]}`);
    let got: unknown;
    try { got = JSON.parse(r.stdout); } catch { throw new Error(`stdout is not JSON: ${r.stdout.slice(0, 80)}`); }
    if (JSON.stringify(got) !== JSON.stringify(EXPECTED)) throw new Error(`wrong output: ${r.stdout.slice(0, 120)}`);
  }
  return t;
}

for (const p of paths) {
  try { p.first = once(p); } catch (e) { p.error = (e as Error).message; console.error(`${p.candidate} ${p.path}: ${p.error}`); }
}
const live = paths.filter((p) => !p.error);
console.error(`timing ${live.length} paths x ${RUNS} runs`);
for (let i = 0; i < RUNS; i++) for (const p of live) p.times.push(once(p));

// ---------- report ----------

const cpu = cpus()[0]?.model ?? 'unknown';
const date = new Date().toISOString().slice(0, 10);
const lines: string[] = [];
const row = (cells: (string | number)[]): number => lines.push(`| ${cells.join(' | ')} |`);

lines.push(`## Hook benchmark: ${PLATFORM}, ${date}`, '');
lines.push(`${cpu}, ${cpus().length} CPUs, ${platform()} ${release()}. ${RUNS} runs per path, wall clock ms.`, '');
lines.push(`Toolchains: ${Object.entries(tools).map(([k, v]) => `${k} ${v ?? 'missing'}`).join('; ')}.`, '');
row(['Candidate', 'Path', 'p50', 'p95', 'min', 'first']);
row(['---', '---', '--:', '--:', '--:', '--:']);
for (const p of paths) {
  const s = [...p.times].sort((a, b) => a - b);
  if (p.error) row([p.candidate, p.path, `failed: ${p.error.replace(/\|/g, '\\|').slice(0, 80)}`, '', '', '']);
  else row([p.candidate, p.path, ms(pct(s, 50)), ms(pct(s, 95)), ms(s[0]), ms(p.first)]);
}
lines.push('', '### Size and build', '');
row(['Candidate', 'Executable or bundle (MB)', 'npm tarball (MB)', 'Build (s)']);
row(['---', '--:', '--:', '--:']);
for (const c of candidates) {
  const tar = c.exe ? tarSizes[`bench-hook-${c.id}-${PLATFORM}`] : undefined;
  row([c.id, mb(c.size), mb(tar), ((c.buildMs ?? 0) / 1000).toFixed(1)]);
}
if (tarSizes['bench-hook-node']) lines.push('', `\`bench-hook-node\` tarball (all JS variants): ${mb(tarSizes['bench-hook-node'])} MB, plus \`yaml\` from the registry.`);
lines.push('', '### Install (all bench packages from local tarballs, one team repo)', '');
row(['Package manager', 'Install (s)', '`.bin` entry', 'node_modules (MB)', 'Note']);
row(['---', '--:', '---', '--:', '---']);
for (const t of teams) {
  const size = existsSync(join(t.dir, 'node_modules')) ? dirSize(join(t.dir, 'node_modules')) : undefined;
  row([t.pm, t.installMs ? (t.installMs / 1000).toFixed(1) : '–', t.binKind ?? '–', mb(size), t.error ?? '']);
}
lines.push('', '### Install per candidate (local tarball, warm package manager cache)', '');
row(['Candidate', 'npm install (s)', 'pnpm install (s)']);
row(['---', '--:', '--:']);
for (const i of installs) row([i.candidate, i.npm ? (i.npm / 1000).toFixed(1) : '–', i.pnpm ? (i.pnpm / 1000).toFixed(1) : '–']);
if (notes.length > 0) lines.push('', '### Notes', '', ...notes.map((n) => `- ${n}`));

const report = lines.join('\n') + '\n';
process.stdout.write(report);
const outDir = join(HERE, 'results');
mkdirSync(outDir, { recursive: true });
// A subset or short run never overwrites a full run's results.
const partial = ONLY !== undefined || RUNS < 50 ? '-partial' : '';
const base = join(outDir, `${PLATFORM}-${date}${partial}`);
writeFileSync(`${base}.md`, report);
// Only numbers and names: no environment, no local paths (they would leak into a committed file).
const summary = {
  platform: PLATFORM, date, cpu, runs: RUNS, tools, notes,
  teams: teams.map(({ pm, installMs, binKind, error }) => ({ pm, installMs, binKind, error })),
  installs,
  candidates: candidates.map(({ id, size, buildMs }) => ({ id, size, buildMs })),
  paths: paths.map(({ candidate, path, times, first, error }) => ({ candidate, path, times, first, error })),
};
writeFileSync(`${base}.json`, JSON.stringify(summary, null, 2) + '\n');
console.error(`wrote ${resolve(base)}.md and .json`);
if (opts.keep) console.error(`kept team repos in ${TEAMS}`);
else rmSync(TEAMS, { recursive: true, force: true });
