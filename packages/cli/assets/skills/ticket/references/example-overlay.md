# Example team overlay for ticket

A synthetic example of a team overlay. It isn't applied: to customize the skill, copy the block
below to `overlays/ticket/OVERLAY.md` in the team repo, drop what you don't want, and open a PR.
`## Add` rules apply on top of the skill; each `## Override {#id}` replaces the section with that ID
in `SKILL.md` or `references/templates.md`. Keep an override complete: it replaces the whole section,
not a line of it. Team examples and templates go in files next to `OVERLAY.md`, referenced from it.

```markdown
## Add
- Every story names the feature flag it ships behind, or says "no flag".
- Bugs reported by customers link the support request; ask for it if missing.
- Gold-standard examples: read `overlays/ticket/examples/` before drafting.

## Override {#estimation}
Estimate stories and tasks in points: 1, 2, 3, 5, 8. Anything above 8 gets split.
Bugs and spikes are not estimated; a spike has a time-box instead.

## Override {#acceptance-criteria}
Every story uses a Given / When / Then table, one row per scenario, with real field names
and real values. Non-functional criteria (performance, accessibility) go in a separate
checklist below the table. Each row must be automatable as an end-to-end test; say so in
Notes if one isn't.

## Override {#titles}
Start with the product area and a colon, then the outcome: "Checkout: total includes
discount codes on saved carts". At most 70 characters.
```
