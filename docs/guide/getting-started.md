# Getting started

Two paths: a **team lead** sets up the team repo once, then each **member** joins it on their machine.
Working alone? See [On your own](#on-your-own).
Steps marked **coming soon** depend on pieces that aren't built yet; each says what to do today.

You need git, Claude Code or Copilot CLI, and the `folkways` binary on your `PATH`: [Install](../../README.md#install).
macOS or Linux (Windows is out of scope for v1).

## For a team lead

### 1. Create the team repo

```bash
folkways init example-team
```

`init` refuses a non-empty directory. It then:

1. Copies the scaffold: `README.md`, `folkways.yaml`, `AGENTS.md`, `CLAUDE.md`, `.gitignore`,
   `.claude/settings.json`, and READMEs for `knowledge/`, `state/`, `efforts/`, `overlays/` and
   `people/`.
2. Runs `git init -b main` and `sync`, which records this `folkways` version in `.folkways/manifest.json`: the
   team's pin.
3. Commits `chore(folkways): init at 0.1.0`.

If any step fails, it removes what it created so you can rerun. No Node, no `package.json`.

Your team repo is ready. <(^_^<)

### 2. Fill in the config

Open Claude Code or Copilot CLI in the new repo and run the `setup` skill. It interviews you for `folkways.yaml` and moves existing team material into
`overlays/` or team skills.

Or edit `folkways.yaml` by hand. The comments say what each field is for. Keep examples synthetic until
you're in your own private repo.

### 3. Push it

Create an empty repo on GitHub, then:

```bash
cd example-team
git remote add origin git@github.com:example/example-team.git
git push -u origin main
```

Protect `main` and require linear history. When the knowledge-audit workflow ships (**coming soon**),
it needs one repo secret for the agent CLI it runs.

## For a member

### 1. Clone and install

```bash
git clone git@github.com:example/example-team.git
cd example-team
```

[Install](../../README.md#install) `folkways` once, the version in `.folkways/manifest.json`. The hooks in
`.claude/settings.json` call it on `PATH`; without it they do nothing, and each session starts with a line saying
how to install it.

### 2. Onboard

Open Claude Code or Copilot CLI in the team repo and run the `onboard` skill. It:

- writes `.local/me.yaml` (your handle, projects directory and tool),
- creates `people/<handle>/`,
- with your approval, gives the tool access to your projects directory,
- checks the MCP servers and CLIs your team's research sources need,
- runs `folkways doctor` and reports what it finds.

Or create `.local/me.yaml` by hand:

```yaml
handle: alex
projects_dir: /Users/alex/projects     # absolute path
tool: claude                           # claude | copilot
```

### 3. Work

Start a session in the team repo. The agent reads `AGENTS.md`, which imports
`.folkways/instructions.md`: session start, where things go, status docs, handoff bands, git rules and
what needs your approval. That part works today.

## On your own

Install the Claude Code plugin (`claude plugin marketplace add sumsolution/folkways`, then
`claude plugin install folkways@folkways`) and open a session. The first session offers to set folkways
up; run `/folkways:setup`, choose solo, and answer a few questions: where your brain repo goes (a new
folder, never a code repo), where your project clones live, which repos you work on, tracker, commit
format, review gate, which hooks to turn off. It writes `folkways.yaml`, `.local/me.yaml` and a
`.gitignore`, and commits `folkways.yaml` and `.gitignore` (`.local/` stays on your machine). From then
on, start sessions in the brain repo; code changes happen in worktrees of your projects. Folders such as
`work/` and `inbox/` appear when a skill first needs them. Not now? Ignore the offer; it won't repeat.

**The plugin won't be published:** it's being removed ([ADR 0012](../adr/0012-cli-only.md)). Until then, build it
from a clone of this repo
(`pnpm --filter "@folkways/claude-plugin..." run build`, needs Go) and start Claude Code with
`claude --plugin-dir packages/claude-plugin/dist`.

## Staying current

When a newer version ships, the lead (or anyone) installs it, then runs, in the team repo:

```bash
folkways upgrade 0.2.0
```

It refuses unless the running `folkways` is that version. Review the branch, push, open a PR. After it merges,
everyone pulls and installs the same version; until they do, each session starts with a warning that names both
versions. See
[How it works](how-it-works.md#upgrades).

## If something goes wrong

| Message | Meaning |
|---|---|
| `folkways: framework files were edited:` | Someone changed a synced file. Move the change into `overlays/`, or rerun with `--force` to discard it |
| `folkways: team skills clash with new framework skills: ...` | A team skill has a framework skill's name. Rename yours |
| `warning: AGENTS.md has no '@.folkways/instructions.md' line; agents won't see the framework instructions.` | Agents won't see the framework rules. Put the line back |
| `folkways: 'review' is not built yet` | Expected in v0.1.0; same for `run` |
