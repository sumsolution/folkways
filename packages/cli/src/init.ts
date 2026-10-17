import { execFileSync } from 'node:child_process';
import { copyFileSync, existsSync, mkdirSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { basename, dirname, join, resolve } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { listFiles } from './files.ts';
import { PNPM_WORKSPACE } from './hooks.ts';
import { git } from './git.ts';
import { PACKAGE_NAME, PACKAGE_ROOT, packageVersion } from './package.ts';
import { installCommand, type PackageManager, packageManagerField } from './package-manager.ts';
import { sync } from './sync.ts';

/** Template paths: `dot-x` becomes `.x`, `name.tmpl` becomes `name`. */
export function templateTarget(rel: string): string {
  return rel.replace(/(^|\/)dot-/g, '$1.').replace(/\.tmpl$/, '');
}

const RESERVED_NAMES = new Set(['node_modules', 'favicon.ico']);

/**
 * An npm package name for a team repo in `folder`: lowercase, URL-safe, no leading `.`, `_` or `-`,
 * at most 214 characters (https://docs.npmjs.com/cli/configuring-npm/package-json#name).
 */
export function packageName(folder: string): string {
  const name = folder
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9._-]/g, '-')
    .replace(/^[-._]+/, '')
    .slice(0, 214);
  if (name === '' || RESERVED_NAMES.has(name)) {
    throw new FolkwaysError(`can't derive a valid npm package name from '${folder}'; use another folder name`);
  }
  return name;
}

export interface InitOptions {
  /** Install dependencies to create the lockfile. Default true. */
  install?: boolean;
  /** Default npm. The team repo's lockfile and `packageManager` field follow this. */
  packageManager?: PackageManager;
  pkg?: string;
}

/** Scaffold a team repo at `path`, pin this package version, sync, and commit. */
export function init(path: string, options: InitOptions = {}): string {
  const pkg = options.pkg ?? PACKAGE_ROOT;
  const dest = resolve(path);
  const name = packageName(basename(dest));
  const existed = existsSync(dest);
  if (existed && readdirSync(dest).length > 0) throw new FolkwaysError(`${dest} exists and is not empty`);
  mkdirSync(dest, { recursive: true });
  try {
    scaffold(dest, name, pkg, options);
  } catch (e) {
    // Leave nothing half-built, so a rerun works.
    if (existed) for (const entry of readdirSync(dest)) rmSync(join(dest, entry), { recursive: true, force: true });
    else rmSync(dest, { recursive: true, force: true });
    throw e;
  }
  return dest;
}

function scaffold(dest: string, name: string, pkg: string, options: InitOptions): void {
  for (const rel of listFiles(join(pkg, 'templates'))) {
    const out = join(dest, templateTarget(rel));
    mkdirSync(dirname(out), { recursive: true });
    copyFileSync(join(pkg, 'templates', rel), out);
  }

  const pm = options.packageManager ?? 'npm';
  const pmField = packageManagerField(pm);
  const version = packageVersion(pkg);
  const manifest = {
    name,
    private: true,
    ...(pmField && { packageManager: pmField }),
    devDependencies: { [PACKAGE_NAME]: version },
  };
  writeFileSync(join(dest, 'package.json'), `${JSON.stringify(manifest, null, 2)}\n`);
  if (pm === 'pnpm') writeFileSync(join(dest, 'pnpm-workspace.yaml'), PNPM_WORKSPACE);

  git(dest, 'init', '-q', '-b', 'main');
  if (options.install ?? true) {
    const [cmd, args] = installCommand(pm);
    try {
      execFileSync(cmd, args, { cwd: dest, stdio: ['ignore', 'inherit', 'pipe'] });
    } catch (e) {
      const detail = (e as { stderr?: Buffer }).stderr?.toString().trim() || (e as Error).message;
      throw new FolkwaysError(`${cmd} ${args.join(' ')} failed; removed what init created:\n${detail}`);
    }
  }
  sync(dest, pkg);
  git(dest, 'add', '-A');
  try {
    git(dest, 'commit', '-q', '-m', `chore(folkways): init at ${version}`);
  } catch (e) {
    const detail = (e as { stderr?: Buffer | string }).stderr?.toString().trim() || (e as Error).message;
    throw new FolkwaysError(`git commit failed; removed what init created:\n${detail}`);
  }
}
