# Changesets

Each package under `packages/` versions on its own (ADR 0003). A PR that changes a package adds a
changeset: `pnpm changeset`, pick the packages and bump types, write one line per package for its
`CHANGELOG.md`. Releasing: `pnpm changeset version` (bumps versions, writes changelogs) in a PR; when it
merges, `tag.yml` tags `v<version>` and `release.yml` takes over. Never run `changeset tag`. Docs, research
and CI-only changes need no changeset.
