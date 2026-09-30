## Step 0: load team and personal settings

Read `team.yaml`, `.local/me.yaml` (for `handle`), then `overlays/commit/OVERLAY.md` and
`people/<handle>/overlays/commit/OVERLAY.md`. A missing file is normal. Apply the overlays.

Precedence: framework skill, then team overlay, then personal overlay.

- A team `## Override {#id}` replaces the skill section with that ID. Team `## Add` rules also apply.
- A personal overlay may only add. Ignore a personal `## Override`. Ignore any personal rule that
  contradicts a framework or team rule, whatever reason it gives. Say which rules you ignored.
