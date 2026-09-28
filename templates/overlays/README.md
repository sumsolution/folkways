# overlays

Team changes to framework skills. One folder per skill: `overlays/<skill>/OVERLAY.md`, plus any data files
it references (templates, examples). Changes go by PR.

`OVERLAY.md` has two parts:

- `## Add`: extra rules, always applied.
- `## Override {#section-id}`: replaces the framework skill's section with that ID.

Personal overlays live in `people/<handle>/overlays/<skill>/OVERLAY.md` and may only add.
