# Example ADR

A synthetic example at the right altitude: concrete context, three real options, a trade-off column
that judges rather than repeats, and consequences that include the costs.

```markdown
# Serve all settings from the settings API, read only

| Property | Value |
|---|---|
| Status | accepted |
| Date | 2026-03-12 |
| Owner | Alex Doe |
| Reviewers | Sam Lee (tech lead), Riya Patel (platform, outside the team) |
| Group | Web |
| Review window | until 2026-03-10 |

## Issue

`web-app` reads feature settings three ways: environment variables at build time, direct queries on
the `settings` table in the shared database, and `GET /v1/settings` on the settings API. The same
setting can differ between them, and two incidents this quarter came from a stale build-time value.
The platform team is moving the `settings` table to a new schema in Q3, which will break the direct
queries. Where should `web-app` read settings from?

## Relevant context

- Settings API: `GET /v1/settings?scope=web` returns all web settings with an `ETag`
  (https://docs.example.com/settings-api). p99 latency 40 ms.
- Direct queries: 14 call sites, all through `src/config/db-settings.ts` in `web-app`.
- Incidents EX-2210 and EX-2291: a flag changed in the admin tool had no effect until the next deploy.
- Settings are written only through the admin tool, which already calls the API.

## Decision

Option 2: read every setting from the settings API, read only. `web-app` fetches on start-up and
refreshes every 60 seconds using the `ETag`, keeping the last good copy if the API is down. It never
writes settings. This removes the three-way drift and the dependency on the table schema, with one
client to maintain; the 60-second staleness is acceptable for every current setting.

## Options considered

| # | Option | Pros | Cons | Trade-off |
|---|---|---|---|---|
| 1 | Keep the three sources | No work now | Drift stays; direct queries break with the Q3 schema change | Keeps the status quo until Q3 forces a rushed change. |
| 2 | Settings API only, read only | One source; decoupled from the schema; changes apply within 60 s | New runtime dependency; up to 60 s stale; start-up needs the API or a cached copy | Adds a dependency, but ends the drift and the schema coupling. |
| 3 | Settings API for reads and writes | One client for everything | Moves admin logic into `web-app`; the admin tool already writes | Solves a problem we don't have and widens `web-app`'s permissions. |

## Consequences

- Settings changes take effect within 60 seconds without a deploy.
- `web-app` can't start cleanly on a cold cache while the settings API is down; start-up falls back
  to the last copy written to disk and logs a warning. EX-2301 covers this.
- The 14 direct-query call sites are removed (EX-2302); build-time settings are removed (EX-2303).
- Revisit if a setting needs to change faster than 60 seconds (for example, a kill switch).

## Interested parties

- Platform team (owns the settings API and the `settings` table schema change).
- Support (settings changes no longer wait for a deploy).
```
