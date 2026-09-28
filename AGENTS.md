# folkways: framework development

Instructions for working on the framework itself. Not vendored; team repos get `vendor/instructions.md`.

## Layout

| Path | Holds |
|---|---|
| `bin/folkways` | The CLI. Vendored to `.folkways/bin/folkways` in team repos |
| `vendor/` | Everything copied into team repos (spec §2.1) |
| `templates/` | One-time scaffold for `init`. `dot-x` becomes `.x`; `name.tmpl` becomes `name` |
| `tests/run.sh` | Test suite. Run it before every commit |
| `docs/design/` | Design record. `00-STATUS.md` is the entry point; `06-v1-spec.md` is the spec |

## Rules

- **No company-specific content.** Names, portals, project keys, internal URLs and paths belong in team
  repos. Examples use synthetic values (`Example Team`, `EX`, `example.com`).
- **Scripts:** bash 3.2 (macOS default) plus `git` and `jq`. No associative arrays, `mapfile`, or GNU-only
  flags. Hash with the `sha256` helper.
- **TDD:** add a failing case to `tests/run.sh` first, then fix.
- **Skills:** frontmatter is `name` and `description` only. Overridable sections carry an ID:
  `## Commit format {#commit-format}`. Step 0 reads `team.yaml`, `.local/me.yaml` and both overlays.
- **Commits:** Angular conventional commits, module scope (`cli`, `vendor`, `templates`, `skills`, `hooks`,
  `docs`, `tests`). Squash PRs.
- **Releases:** bump `VERSION`, add a `CHANGELOG.md` entry, tag `v<VERSION>`. `folkways upgrade` fetches by tag.
- Keep `docs/design/00-STATUS.md` current at milestones.
