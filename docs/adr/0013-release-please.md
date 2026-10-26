# 0013. Should conventional commits decide folkways' version, and which tool releases it?

| | |
|---|---|
| Status | proposed |
| Date | 2026-10-26 |
| Owner | Summer Jordan |
| Reviewers | Summer Jordan (one-person team, `adr.reviewers.min: 0`) |
| Review window | until Summer's review of the PR that adds this record |
| Supersedes | In part, [0003](0003-repo-layout.md): "Independent versions with changesets" and the changesets row of its options |

## Issue

folkways ships one thing now: the `folkways` binary, released by a `v<version>` tag ([ADR 0011](0011-cli-in-go.md),
[ADR 0012](0012-cli-only.md)). Its version still comes from changesets: a PR adds `.changeset/*.md`, someone runs
`pnpm changeset version` in a version PR, and `tag.yml` tags `v<version>` when that PR changes
`packages/folkways/package.json` ([AGENTS.md][c-agents-rel], [tag.yml][c-tag]). `package.json` survives only to hold
that version, so the release path still needs Node ([tag.yml][c-tag-node], [release.yml][c-rel-node]). Every PR
title is already a conventional commit (62 of 63 on `main`; the other is `doc:`, [research 02][r-02-commits]), and that
information goes unused: the bump is a person's choice in each change file. 0.2.0 shipped this way on 2026-10-23.
Changing tools is cheapest before the next release.

Should conventional commits decide the version, with CI writing the changelog and tagging, and which tool does it?

## Context

- **Rulesets on `sumsolution/folkways`** (gh api, 2026-10-23): `main` takes changes only by PR (squash), with linear
  history and the `ci` check; admins bypass. `release tags` (`refs/tags/v*`) lets only admins and the tap App create
  tags ([research 01 §2][r-01-2]).
- **Squash commits carry the PR body.** The repo sets `squash_merge_commit_title: PR_TITLE`,
  `squash_merge_commit_message: PR_BODY` (gh api, 2026-10-23).
- **A tag made with the App token starts `release.yml`.** `tag.yml` creates the tag with the refs API and the tap
  App's token, from the `tag` environment ([tag.yml][c-tag]); that started the `v0.2.0` release run (2026-10-23). A
  tag or PR made with `GITHUB_TOKEN` starts no workflows ([release-please-action README][rpa-token]).
- **`release.yml` reuses an existing GitHub release:** it creates a draft only if none exists, uploads, then
  publishes ([release.yml][c-rel-reuse]). Notes come from the `## <version>` section of
  `packages/folkways/CHANGELOG.md` ([release.yml][c-rel-notes]).
- **What depends on the tag stays the same whatever we pick:** GoReleaser's version ([.goreleaser.yaml][c-gr]),
  `go install …@v<version>` and `folkways version` from build info ([main.go][c-main]).
- **Change-file tools must commit back.** knope deletes the change files it uses and has no other record of them; a
  run that doesn't commit the deletions uses them again ([research 01 §2][r-01-2], tested). Any tool that writes
  `CHANGELOG.md` in the repo needs a commit on `main`, so a PR or a ruleset bypass.
- **release-please was dry-run against the repo** after `v0.2.0` was published ([research 02, Dry run][r-02-dry]).
  It found the `v0.2.0` release, proposed nothing for the `build:` commit after it, proposed 0.3.0 for test
  `feat` + `fix` commits, skipped a `packages/evals`-only `feat`, and turned a `fix:` paragraph in a commit body
  into a changelog entry.

## Decision

Option 4: release-please in manifest mode. We will let conventional commits on `main` decide the bump and let
release-please keep one release PR open, which updates `packages/folkways/CHANGELOG.md` and
`.release-please-manifest.json`. Merging it makes release-please create the `v<version>` tag and a draft GitHub
release with the App token. The tag starts `release.yml`, which builds, waits for approval, and publishes into that
draft. Changesets, `tag.yml` and the release path's Node calls go.

It is the only option that uses the commit types we already write and runs the release PR and the tag itself, with
no bypass on `main`: the human step stays one PR merge, as today. knope with conventional commits gets the version
from tags alone, without a manifest file, but leaves the release PR and post-merge tagging to scripts we'd write, and
we found no setting in it to keep a breaking change below 1.0 ([research 02, knope comparison][r-02-knope]).

Config (as dry-run):

```jsonc
// release-please-config.json
{ "packages": { ".": {
  "release-type": "go", "package-name": "folkways",
  "changelog-path": "packages/folkways/CHANGELOG.md",
  "include-component-in-tag": false,   // tags stay v<version>
  "bump-minor-pre-major": true,        // breaking stays 0.x until we choose 1.0 (Release-As: 1.0.0)
  "draft": true, "force-tag-creation": true,
  "pull-request-title-pattern": "build: release folkways ${version}",
  "exclude-paths": ["packages/evals"]
} } }
// .release-please-manifest.json
{ ".": "0.2.0" }
```

Non-goals: automating prereleases (`v<version>-rc.<n>` stays a hand-pushed tag), the plugin's release, choosing when
to reach 1.0, linting PR titles in CI.

Size: two JSON files; a `release-please.yml` workflow of about 25 lines; delete `tag.yml`; about 10 changed lines in
`release.yml` (tag check, and the notes step's heading match, since release-please writes
`## [<version>](<compare link>) (<date>)`); delete `.changeset/` and `@changesets/cli`; `AGENTS.md` rules. One PR.

## Options considered

Analysis: [research 01, knope][r-01] and [research 02, release-please][r-02].

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Keep changesets (as shipped for 0.2.0); optionally `changesets/action` to open the version PR | Works today; notes are hand-written per change; no new tool | `package.json` stays the version record and Node stays in the release path; commit types unused; version PR by hand unless the action is added (then the App needs Pull requests: write too) | The status quo, paid for in a hand-run step and a Node file that exists only to hold a number. |
| 2 | knope with change files, release PR | Version from the last `v*` tag, no version file; keeps change files | Entries must be renamed `default:`; `.changeset/README.md` must move; release PR and post-merge tag are our scripts; `Release` alone can't tag without a version file | Gets tags as the record but keeps the per-PR change file and adds glue we own. |
| 2a | knope with change files, committing to `main` | Exactly "merge a feature PR → tag", no second PR | The App must bypass the `main` ruleset (PR, `ci`, review); unreviewed bot commits; push races | Saves one merge per release by letting a CI key write `main`. |
| 3 | knope with conventional commits | Version from tags only; runs locally; no Node | Same release-PR and tagging glue as 2; a breaking change always bumps major, no pre-1.0 setting found | Closest to "version from tags", at the cost of scripts and of 1.0 arriving by accident. |
| 4 | release-please, manifest mode | Commit types decide the bump; release PR, labels, tag and draft release built in; pre-1.0 settings; `exclude-paths`; dry-run matched expectations | Version also recorded in `.release-please-manifest.json`; notes are PR titles, thinner than hand-written ones; PR-body paragraphs starting `feat:`/`fix:` become entries; GitHub-API tool, no local run; App needs Pull requests: write | Turnkey for the flow we want, if we accept PR titles as release notes and a bot-written manifest. |

Plain tags with hand-written notes was dismissed: the bump is a person's call each time and nothing asks a PR to
describe itself.

## Consequences

- **PR titles become the release notes.** They need the care a changeset line gets today. Richer notes go in a
  `BEGIN_COMMIT_OVERRIDE … END_COMMIT_OVERRIDE` block in the PR body, or an edit to the release PR before merging.
- **PR bodies change meaning:** a paragraph starting `feat:`, `fix:` etc. adds an entry and can raise the bump
  ([research 02 finding 8][r-02]). Agents writing PR bodies need to know (the `pr` skill or `AGENTS.md`).
- **Bumps follow commit types:** any `feat` gives a minor bump at 0.x, so most releases are minor; breaking changes
  stay minor until we set `Release-As: 1.0.0`.
- **Two records of the version:** the manifest and the tag. A hand-pushed final tag without a release PR would put
  them out of step; release-please trusts the manifest.
- **The tap App gains Pull requests: write** on `folkways`, alongside Contents: write. No new ruleset bypass.
- **Removed:** `tag.yml`, `.changeset/` (only its README and config are left), `@changesets/cli` and the
  `changeset` script, the version meaning of `packages/folkways/package.json`, the `node -p` calls in `tag.yml` and
  `release.yml`. `release.yml`'s tag check compares with the manifest via `jq`. Node stays for the TypeScript tests
  ([ci.yml][c-ci]).
- **Changed:** `AGENTS.md` package and Releasing rules ([AGENTS.md][c-agents-rel]); ADR 0003's changesets bullet
  marked superseded by this record.
- **Not yet proven end to end:** `ci` on the App's release PR, the draft hand-off to `release.yml`, the notes step
  finding release-please's `## [<version>](…)` heading (today's awk match would fail the build), and the merged
  PR's title parsing back (the dry run warned the pattern lacks `${component}`). The first real release is the test;
  if tagging fails, the tag can be pushed by hand as with `tag.yml`.
- **Revisit if** the notes prove too thin to read, we need more than one released package, release-please stops
  being maintained, or we want local runs without the GitHub API (option 3).

## Related decisions

- [0003](0003-repo-layout.md): superseded in part (independent versions with changesets).
- [0011](0011-cli-in-go.md), [0012](0012-cli-only.md): one Go binary, released from a `v<version>` tag; unchanged.

[c-agents-rel]: https://github.com/sumsolution/folkways/blob/09cd64c/AGENTS.md#L65-L74
[c-tag]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/tag.yml#L1-L62
[c-tag-node]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/tag.yml#L35
[c-rel-node]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/release.yml#L44-L55
[c-rel-reuse]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/release.yml#L153-L164
[c-rel-notes]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/release.yml#L68-L85
[c-gr]: https://github.com/sumsolution/folkways/blob/09cd64c/.goreleaser.yaml#L15-L16
[c-main]: https://github.com/sumsolution/folkways/blob/09cd64c/packages/folkways/cmd/folkways/main.go#L24-L43
[c-ci]: https://github.com/sumsolution/folkways/blob/09cd64c/.github/workflows/ci.yml#L78-L110
[rpa-token]: https://github.com/googleapis/release-please-action/blob/0b6b3fc0186a2f7118bfd88fab9ea481e1839504/README.md#L106-L116
[r-01]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/01-research-knope.md
[r-01-2]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/01-research-knope.md#2-where-the-changelog-and-consumed-files-go
[r-02]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/02-research-release-please.md
[r-02-commits]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/02-research-release-please.md#folkways-commits-as-input
[r-02-dry]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/02-research-release-please.md#dry-run-2026-10-23
[r-02-knope]: https://github.com/sumsolution/folkways-command/blob/66d490a/people/summer/work/knope-versioning/02-research-release-please.md#knope-with-conventional-commits-for-comparison
