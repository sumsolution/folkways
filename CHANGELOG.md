# Changelog

## Unreleased

- `sync` refuses to overwrite team files that differ at framework paths the manifest doesn't track,
  including every path when there is no manifest (adopting a repo). `--force` overwrites them and lists
  them as `replaced`. Adopting a repo whose framework skill folders already match now works.
- `sync` stages new files inside the team repo and swaps them in by rename, manifest last. A failure
  undoes the swap, so the repo keeps the old files and a rerun works. After a hard kill, the next
  `sync` refuses and says how to recover.
- Framework hooks are those whose command starts with `node_modules/.bin/folkways hook `
  (`FRAMEWORK_HOOK_PREFIX`, `isFrameworkHook`). Team hooks that mention `folkways hook` stay.
- Settings merge keeps hook groups and entries of unexpected shape instead of crashing, and reports
  settings it can't merge into as a `folkways:` error.
- `init` derives a valid npm package name from the folder (`init .hidden` → `hidden`, `équipe` → `equipe`), or refuses before
  creating anything.
- `upgrade` reports a failed commit (e.g. no git identity) cleanly and goes back to the starting branch.
- CI: `pnpm run smoke` packs the tarball, installs it, runs `init --no-install`, `sync` twice and
  `version`. Test and smoke run on Linux and macOS; every job has a timeout.

## 0.1.0 (unreleased)

- `@folkways/cli` npm package. Node 24+.
- `folkways init <path>`: scaffold a team repo from `templates/`, pin this version in its `package.json`,
  install, sync, and make the first commit.
- `folkways sync [--force]`: copy the installed version's instructions, skills, prompts and workflows into
  the team repo and merge its settings. Stops on edited framework files (`--force` discards them) and on
  team skills that clash with new framework skills.
- `folkways upgrade <version>`: install a version on `chore/folkways-<version>`, run its sync, and commit.
- Team repos use npm or pnpm, detected from `packageManager` or the lockfile. `init --package-manager`
  defaults to the one that ran it.
- Team `AGENTS.md` imports the framework instructions with `@.folkways/instructions.md`.
- `upgrade` refuses non-exact versions and a detached HEAD, and cleans up fully when install or sync fails.
  `init` removes what it created when the install fails, so a rerun works.
- `sync` and `upgrade` reject a manifest with paths outside framework-owned folders, so `--force` can't
  delete team files. `init` reports a failed first commit cleanly. Versions must be strict SemVer.
- Settings ask before `git push`. No hooks yet.
