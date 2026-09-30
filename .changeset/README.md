# Changesets

Each package under `packages/` versions on its own (ADR 0003). A PR that changes a package adds a
changeset: `pnpm changeset`, pick the packages and bump types, write one line per package for its
`CHANGELOG.md`. Releasing: `pnpm changeset version` (bumps versions, writes changelogs) in a PR; after it
merges, `pnpm changeset tag` and push the tags. Docs, research and CI-only changes need no changeset.
