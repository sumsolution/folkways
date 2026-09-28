# Changelog

## 0.1.0 (unreleased)

- `bin/folkways init <path>`: scaffold a team repo from `templates/` and vendor the framework.
- `bin/folkways upgrade <version>`: vendor a tagged version on `chore/folkways-<version>`. Stops on edited
  framework files (`--force` discards them) and on team skills that clash with new framework skills.
- `vendor/instructions.md`: the way-of-working block for team `AGENTS.md`.
- `vendor/settings.hooks.json`: asks before `git push`. No hooks yet.
