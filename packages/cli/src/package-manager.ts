import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { FolkwaysError } from './errors.ts';

export type PackageManager = 'npm' | 'pnpm';
export const PACKAGE_MANAGERS: readonly PackageManager[] = ['npm', 'pnpm'];

type Command = [string, string[]];

function supported(name: string): PackageManager {
  if ((PACKAGE_MANAGERS as readonly string[]).includes(name)) return name as PackageManager;
  throw new FolkwaysError(`${name} is not supported; use ${PACKAGE_MANAGERS.join(' or ')}`);
}

/** The team repo's package manager: `packageManager` field, then lockfile, then npm. */
export function detectPackageManager(team: string): PackageManager {
  const manifest = join(team, 'package.json');
  if (existsSync(manifest)) {
    const field = (JSON.parse(readFileSync(manifest, 'utf8')) as { packageManager?: string }).packageManager;
    if (field) return supported(field.split('@')[0]!);
  }
  if (existsSync(join(team, 'pnpm-lock.yaml'))) return 'pnpm';
  return 'npm';
}

/** The package manager that launched us (`npx` or `pnpm dlx`), from npm_config_user_agent. */
export function fromUserAgent(ua: string | undefined = process.env.npm_config_user_agent): PackageManager | undefined {
  const name = ua?.split('/')[0];
  return name && (PACKAGE_MANAGERS as readonly string[]).includes(name) ? (name as PackageManager) : undefined;
}

export function installCommand(pm: PackageManager): Command {
  return [pm, ['install']];
}

export function addDevExactCommand(pm: PackageManager, spec: string): Command {
  return [pm, [pm === 'pnpm' ? 'add' : 'install', '--save-dev', '--save-exact', spec]];
}

/** `pnpm@<installed version>` for the packageManager field. Undefined for npm. */
export function packageManagerField(pm: PackageManager): string | undefined {
  if (pm === 'npm') return undefined;
  try {
    return `pnpm@${execFileSync('pnpm', ['--version'], { encoding: 'utf8' }).trim()}`;
  } catch {
    throw new FolkwaysError('pnpm is not installed; install it or use --package-manager npm');
  }
}
