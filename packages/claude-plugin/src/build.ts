// Build the folkways Claude Code plugin into dist/: the framework's skills and instructions rewritten for
// the plugin, hooks generated from the team-repo settings, and folkways-hook for every platform.
// The binaries come from `@folkways/hooks`'s own build, which this reads and never runs: `pnpm -r run build`
// builds the hooks first (a workspace devDependency), and `pnpm --filter "@folkways/claude-plugin..." run build`
// builds both.
// Usage: node src/build.ts [out-dir] [hooks-build]   (defaults dist/ and packages/hooks/npm/)
import { chmodSync, copyFileSync, cpSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';

const PACKAGE = resolve(import.meta.dirname, '..');
const REPO = resolve(PACKAGE, '..', '..');
/** The framework content (`@folkways/content`): the files `sync` copies into team repos. Read as data, never imported. */
export const ASSETS = join(REPO, 'packages', 'content');
/** Where `@folkways/hooks` builds its platform packages by default. */
const HOOKS_BUILD = join(REPO, 'packages', 'hooks', 'npm');

/** `<os>-<cpu>` of every folkways-hook build, as `@folkways/hooks` names its platform packages. */
export const HOOK_TARGETS = ['darwin-arm64', 'darwin-x64', 'linux-arm64', 'linux-x64'] as const;

export const hostTarget = (): string | undefined =>
  HOOK_TARGETS.find((t) => t === `${process.platform}-${process.arch}`);

/** The framework skills: one folder each under the framework content. */
export function frameworkSkills(): string[] {
  return readdirSync(join(ASSETS, 'skills'), { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();
}

const escape = (s: string): string => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

/**
 * Text written for a synced team repo, made right for the plugin: plugin skills are namespaced
 * (`folkways:start`) and live under the plugin root, not `.claude/skills/`.
 */
export function rewriteForPlugin(text: string, skills: readonly string[]): string {
  const names = skills.map(escape).join('|');
  return text
    .replace(new RegExp(`\`(${names})\`(\\s+)skill`, 'g'), '`folkways:$1`$2skill')
    .replace('Skills in `.claude/skills/` hold', 'The `folkways:` skills hold')
    .replace(new RegExp(`\\.claude/skills/(${names})/`, 'g'), '$${CLAUDE_PLUGIN_ROOT}/skills/$1/')
    .replace(/^Framework skills:[^]*?\.$/m, (list) => list.replace(new RegExp(`\`(${names})\``, 'g'), '`folkways:$1`'));
}

const PREFACE = `# folkways (plugin)

The folkways plugin is enabled. Its skills are namespaced (\`folkways:start\`, \`folkways:handoff\`, …).
Team overlays in \`overlays/\` aren't compiled in; that needs a team repo and \`folkways sync\`. Personal
overlays still attach.
Without a \`folkways.yaml\` in the repo the plugin stays quiet. folkways runs from a brain repo, which
holds no code; the \`folkways:setup\` skill creates a solo one or sets up a team one.

`;

interface HookEntry {
  type: string;
  command: string;
}
interface HookGroup {
  matcher?: string;
  hooks: HookEntry[];
}
interface Hooks {
  description?: string;
  hooks: Record<string, HookGroup[]>;
}

const runner = (hook: string): HookEntry => ({ type: 'command', command: `sh "\${CLAUDE_PLUGIN_ROOT}"/hooks/run.sh ${hook}` });

/** The plugin's hooks.json: every hook in the team-repo settings, on the same event, run by `run.sh`. */
export function pluginHooks(settings: { hooks: Record<string, HookGroup[]> }): Hooks {
  const hooks: Record<string, HookGroup[]> = { SessionStart: [{ hooks: [runner('instructions')] }] };
  for (const [event, groups] of Object.entries(settings.hooks)) {
    const mapped = groups.map((g) => ({
      ...(g.matcher === undefined ? {} : { matcher: g.matcher }),
      hooks: g.hooks.map((h) => runner(h.command.trim().split(/\s+/).at(-1)!)),
    }));
    hooks[event] = [...(hooks[event] ?? []), ...mapped];
  }
  return { description: 'Generated from packages/content/settings.json by @folkways/claude-plugin', hooks };
}

const json = (value: unknown): string => `${JSON.stringify(value, null, 2)}\n`;

const DESCRIPTION = 'The folkways way of working: framework skills, session instructions and hooks';

export interface BuildOptions {
  out: string;
  version: string;
  /** `<os>-<cpu>` → a built folkways-hook. */
  binaries: Map<string, string>;
}

/** Write the complete plugin to `out`, replacing what was there. */
export function buildPlugin({ out, version, binaries }: BuildOptions): void {
  rmSync(out, { recursive: true, force: true });
  const skills = frameworkSkills();
  const write = (path: string, content: string) => {
    mkdirSync(dirname(join(out, path)), { recursive: true });
    writeFileSync(join(out, path), content);
  };

  for (const skill of skills) {
    const src = join(ASSETS, 'skills', skill);
    cpSync(src, join(out, 'skills', skill), { recursive: true });
    for (const f of readdirSync(src, { recursive: true, encoding: 'utf8' }).filter((f) => f.endsWith('.md'))) {
      write(join('skills', skill, f), rewriteForPlugin(readFileSync(join(src, f), 'utf8'), skills));
    }
  }
  write('instructions.md', PREFACE + rewriteForPlugin(readFileSync(join(ASSETS, 'instructions.md'), 'utf8'), skills));

  write('hooks/hooks.json', json(pluginHooks(JSON.parse(readFileSync(join(ASSETS, 'settings.json'), 'utf8')))));
  copyFileSync(join(PACKAGE, 'src', 'run.sh'), join(out, 'hooks', 'run.sh'));
  mkdirSync(join(out, 'hooks', 'bin'));
  for (const target of HOOK_TARGETS) {
    const bin = binaries.get(target);
    if (!bin) throw new Error(`no folkways-hook built for ${target}`);
    const dest = join(out, 'hooks', 'bin', `folkways-hook-${target}`);
    copyFileSync(bin, dest);
    chmodSync(dest, 0o755);
  }

  write(
    '.claude-plugin/plugin.json',
    json({
      name: 'folkways',
      version,
      description: DESCRIPTION,
      author: { name: 'Summer Jordan' },
      homepage: 'https://github.com/sumsolution/folkways',
      repository: 'https://github.com/sumsolution/folkways',
      license: 'Apache-2.0',
      keywords: ['team', 'workflow', 'skills'],
    }),
  );
  write(
    '.claude-plugin/marketplace.json',
    json({
      name: 'folkways-dev',
      description: 'This build of the folkways plugin, loaded in place (development only)',
      owner: { name: 'Summer Jordan' },
      plugins: [{ name: 'folkways', source: './', description: DESCRIPTION }],
    }),
  );
  write(
    'package.json',
    json({
      name: '@folkways/claude-plugin',
      version,
      description: `${DESCRIPTION}. A Claude Code plugin; install it from the folkways marketplace.`,
      license: 'Apache-2.0',
      repository: { type: 'git', url: 'git+https://github.com/sumsolution/folkways.git', directory: 'packages/claude-plugin' },
      files: ['.claude-plugin/plugin.json', 'skills', 'hooks', 'instructions.md'],
      publishConfig: { access: 'public' },
    }),
  );
  copyFileSync(join(PACKAGE, 'README.md'), join(out, 'README.md'));
  copyFileSync(join(REPO, 'LICENSE'), join(out, 'LICENSE'));
}

/** The folkways-hook binaries in a `@folkways/hooks` build (`<dir>/<os>-<cpu>/bin/folkways-hook`). */
export function hookBinaries(dir: string): Map<string, string> {
  if (!existsSync(dir)) {
    throw new Error(`no @folkways/hooks build in ${dir}: run \`pnpm --filter "@folkways/claude-plugin..." run build\``);
  }
  const built = readdirSync(dir).sort();
  if (JSON.stringify(built) !== JSON.stringify([...HOOK_TARGETS])) {
    throw new Error(`@folkways/hooks built ${built.join(', ')}; HOOK_TARGETS lists ${HOOK_TARGETS.join(', ')}`);
  }
  return new Map(HOOK_TARGETS.map((t) => [t, join(dir, t, 'bin', 'folkways-hook')]));
}

if (import.meta.main) {
  const version: string = JSON.parse(readFileSync(join(PACKAGE, 'package.json'), 'utf8')).version;
  const out = resolve(process.argv[2] ?? join(PACKAGE, 'dist'));
  try {
    buildPlugin({ out, version, binaries: hookBinaries(resolve(process.argv[3] ?? HOOKS_BUILD)) });
  } catch (error) {
    console.error(`folkways: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  }
  console.log(out);
}
