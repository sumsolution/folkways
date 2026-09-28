# Changelog

## 0.1.0 (unreleased)

- `@folkways/cli` npm package. Node 24+.
- `folkways init <path>`: scaffold a team repo from `templates/`, pin this version in its `package.json`,
  install, sync, and make the first commit.
- `folkways sync [--force]`: copy the installed version's instructions, skills, prompts and workflows into
  the team repo and merge its settings. Stops on edited framework files (`--force` discards them) and on
  team skills that clash with new framework skills.
- `folkways upgrade <version>`: install a version on `chore/folkways-<version>`, run its sync, and commit.
- Team `AGENTS.md` imports the framework instructions with `@.folkways/instructions.md`.
- Settings ask before `git push`. No hooks yet.
