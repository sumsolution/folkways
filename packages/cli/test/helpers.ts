import { cpSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, realpathSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { afterEach } from 'vitest';
import { listFiles } from '../src/files.ts';
import { git } from '../src/git.ts';
import { init } from '../src/init.ts';

export const REPO = resolve(import.meta.dirname, '..');

const dirs: string[] = [];
afterEach(() => {
  for (const d of dirs.splice(0)) rmSync(d, { recursive: true, force: true });
});

/** A temp dir, removed after the test. Resolved, so paths compare equal to git's. */
export function tmp(): string {
  const d = realpathSync(mkdtempSync(join(tmpdir(), 'folkways-')));
  dirs.push(d);
  return d;
}

/**
 * A copy of this package at `version`: package.json, assets/, templates/, src/.
 * `src/` lets tests run the copy's CLI directly with Node's type stripping.
 * Skills, prompts and workflows are left out so tests control them (`addFrameworkContent`).
 */
export function makePackage(version: string): string {
  const pkg = join(tmp(), 'pkg');
  mkdirSync(pkg);
  const content = ['skills', 'prompts', 'workflows'].map((d) => join(REPO, 'assets', d));
  for (const d of ['assets', 'templates', 'src']) {
    cpSync(join(REPO, d), join(pkg, d), { recursive: true, filter: (src) => !content.includes(src) });
  }
  writeFileSync(join(pkg, 'package.json'), JSON.stringify({ name: '@folkways/cli', version, type: 'module' }));
  return pkg;
}

/** Write a file, creating parent dirs. */
export function put(path: string, content: string): void {
  mkdirSync(resolve(path, '..'), { recursive: true });
  writeFileSync(path, content);
}

export const read = (path: string): string => readFileSync(path, 'utf8');
export const json = (path: string): any => JSON.parse(read(path));

/** Add a skill, prompt, workflow and hook entry to a package copy. */
export function addFrameworkContent(pkg: string): void {
  put(join(pkg, 'assets/skills/start/SKILL.md'), '---\nname: start\ndescription: test\n---\nbody\n');
  put(join(pkg, 'assets/skills/start/references/a.md'), 'ref\n');
  put(join(pkg, 'assets/prompts/review.md'), 'review prompt\n');
  put(join(pkg, 'assets/workflows/knowledge-audit.yml'), 'name: audit\n');
  const settings = json(join(pkg, 'assets/settings.json'));
  settings.hooks = {
    SessionStart: [{ hooks: [{ type: 'command', command: '"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook session-start' }] }],
  };
  writeFileSync(join(pkg, 'assets/settings.json'), JSON.stringify(settings));
}

/** A team repo created by `init` from `pkg`, without npm install. */
export function makeTeam(pkg: string): string {
  return init(join(tmp(), 'team'), { install: false, pkg });
}

export function commitAll(team: string, message = 'team change'): void {
  git(team, 'add', '-A');
  git(team, 'commit', '-q', '-m', message);
}

/** Every file in a team repo outside .git, with its content. */
export function snapshot(team: string): Record<string, string> {
  // Never walk .git: background git maintenance (macOS) adds and removes lock files there.
  const files = readdirSync(team)
    .filter((e) => e !== '.git')
    .flatMap((e) => (statSync(join(team, e)).isDirectory() ? listFiles(join(team, e)).map((f) => `${e}/${f}`) : [e]));
  return Object.fromEntries(files.sort().map((f) => [f, read(join(team, f))]));
}
