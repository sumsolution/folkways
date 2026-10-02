// At pack and publish time, pin @folkways/cli's hook binaries as exact optionalDependencies (ADR 0002).
// Only the packed package.json carries them: committed pins to unpublished packages would break
// `pnpm install` here (plan 11, decision 3). The version is `folkways.hooks` in packages/cli/package.json.
// Keep PLATFORMS in step with TARGETS in packages/hooks/scripts/build.ts; scripts/smoke.ts checks it.
const PLATFORMS = ['darwin-arm64', 'darwin-x64', 'linux-arm64', 'linux-x64'];

export const hooks = {
  beforePacking(pkg) {
    if (pkg.name !== '@folkways/cli') return pkg;
    const version = pkg.folkways?.hooks;
    if (!version) throw new Error('@folkways/cli: package.json has no folkways.hooks version');
    pkg.optionalDependencies = Object.fromEntries(PLATFORMS.map((p) => [`@folkways/hooks-${p}`, version]));
    return pkg;
  },
};
