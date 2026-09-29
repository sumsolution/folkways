import { execFileSync } from 'node:child_process';
import { copyFileSync, existsSync, mkdirSync, readdirSync, writeFileSync } from 'node:fs';
import { basename, dirname, join, resolve } from 'node:path';
import { FolkwaysError } from './errors.ts';
import { listFiles } from './files.ts';
import { git } from './git.ts';
import { PACKAGE_NAME, PACKAGE_ROOT, packageVersion } from './package.ts';
import { installCommand, type PackageManager, packageManagerField } from './package-manager.ts';
import { sync } from './sync.ts';

/** Template paths: `dot-x` becomes `.x`, `name.tmpl` becomes `name`. */
export function templateTarget(rel: string): string {
  return rel.replace(/(^|\/)dot-/g, '$1.').replace(/\.tmpl$/, '');
}

export interface InitOptions {
  /** Install dependencies to create the lockfile. Default true. */
  install?: boolean;
  /** Default npm. The team repo's lockfile and `packageManager` field follow this. */
  packageManager?: PackageManager;
  pkg?: string;
}

/** Scaffold a team repo at `path`, pin this package version, sync, and commit (spec §8.1). */
export function init(path: string, options: InitOptions = {}): string {
  const pkg = options.pkg ?? PACKAGE_ROOT;
  const dest = resolve(path);
  if (existsSync(dest) && readdirSync(dest).length > 0) throw new FolkwaysError(`${dest} exists and is not empty`);
  mkdirSync(dest, { recursive: true });

  for (const rel of listFiles(join(pkg, 'templates'))) {
    const out = join(dest, templateTarget(rel));
    mkdirSync(dirname(out), { recursive: true });
    copyFileSync(join(pkg, 'templates', rel), out);
  }

  const pm = options.packageManager ?? 'npm';
  const pmField = packageManagerField(pm);
  const version = packageVersion(pkg);
  const name = basename(dest).toLowerCase().replace(/[^a-z0-9._-]/g, '-');
  const manifest = {
    name,
    private: true,
    ...(pmField && { packageManager: pmField }),
    devDependencies: { [PACKAGE_NAME]: version },
  };
  writeFileSync(join(dest, 'package.json'), `${JSON.stringify(manifest, null, 2)}\n`);

  git(dest, 'init', '-q', '-b', 'main');
  if (options.install ?? true) {
    const [cmd, args] = installCommand(pm);
    execFileSync(cmd, args, { cwd: dest, stdio: 'inherit' });
  }
  sync(dest, pkg);
  git(dest, 'add', '-A');
  git(dest, 'commit', '-q', '-m', `chore(folkways): init at ${version}`);
  return dest;
}
