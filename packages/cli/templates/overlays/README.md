# overlays

Team changes to framework skills. One folder per skill: `overlays/<skill>/OVERLAY.md`, plus any data files
it references (templates, examples). Changes go by PR.

`OVERLAY.md` has two parts:

- `## Add`: extra rules, always applied.
- `## Override {#section-id}`: replaces the framework skill's section with that ID, in `SKILL.md` or one
  of its reference files.

`folkways sync` compiles each overlay into `.claude/skills/<skill>/`: overrides replace their sections and
additions are appended as `## Team additions`. Run it after changing an overlay and commit the result.
An override of an ID the skill doesn't have stops the sync. Inside an override, use `###` and deeper
headings: a `##` heading ends the override. Don't override a section and one of its subsections. Rules that must hold whatever an overlay
says belong in `folkways.yaml`, where hooks check them.
