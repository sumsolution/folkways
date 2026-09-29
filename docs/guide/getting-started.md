# Getting started

Two paths: a **team lead** sets up the team repo once, then each **member** joins it on their machine.
Steps marked **coming soon** depend on pieces that aren't built yet; each says what to do today.

You need Node 24+, git, and Claude Code or Copilot CLI. macOS or Linux (Windows is out of scope for v1).

## For a team lead

### 1. Create the team repo

Once `@folkways/cli` is on npm (**coming soon**):

```bash
npx @folkways/cli init example-team
# or: pnpm dlx @folkways/cli init example-team
```

Until then, run it from a clone of this repo:

```bash
git clone https://github.com/sumsolution/folkways.git
cd folkways
pnpm install
node src/cli.ts init ../example-team --no-install
```

`init` refuses a non-empty directory. It then:

1. Copies the scaffold: `team.yaml`, `AGENTS.md`, `CLAUDE.md`, `.gitignore`, `.claude/settings.json`,
   and READMEs for `knowledge/`, `state/`, `efforts/`, `overlays/` and `people/`.
2. Writes a private `package.json` that pins this exact `@folkways/cli` version.
3. Runs `git init -b main`, installs (skipped with `--no-install`), and runs `sync`.
4. Commits `chore(folkways): init at 0.1.0`.

If any step fails, it removes what it created so you can rerun. The package manager defaults to the one
that ran it (`npx` → npm, `pnpm dlx` → pnpm); override with `--package-manager npm|pnpm`.

Your team repo is ready. <(^_^<)

> **Why `--no-install` today:** installing would fetch `@folkways/cli` from npm, and it isn't published
> yet. Without it you get no lockfile and no `node_modules/`; that's fine for a look around.

### 2. Fill in the config

Open Claude Code or Copilot CLI in the new repo and run the `setup` skill (**coming soon**). It will
interview you for `team.yaml` and move existing team material into `overlays/` or team skills.

Today: edit `team.yaml` by hand. The comments say what each field is for. Keep examples synthetic until
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
npm ci                              # or: pnpm install --frozen-lockfile, if there's a pnpm-lock.yaml
```

Installing puts the pinned `@folkways/cli` in `node_modules/`, which is where hooks will run from.
(**Coming soon:** needs the npm release and a committed lockfile.)

### 2. Onboard

Open Claude Code or Copilot CLI in the team repo and run the `onboard` skill (**coming soon**). It will:

- write `.local/me.yaml` (your handle, projects directory and tool),
- create `people/<handle>/`,
- with your approval, give the tool access to your projects directory,
- check the MCP servers and CLIs your team's research sources need,
- run `folkways doctor` (**coming soon**).

Today: create `.local/me.yaml` by hand:

```yaml
handle: alex
projects_dir: /Users/alex/projects     # absolute path
tool: claude                           # claude | copilot
```

### 3. Work

Start a session in the team repo. The agent reads `AGENTS.md`, which imports
`.folkways/instructions.md`: session start, where things go, status docs, handoff bands, git rules and
what needs your approval. That part works today.

## Staying current

When a new version ships, the lead (or anyone) runs:

```bash
npx folkways upgrade 0.2.0             # or: pnpm exec folkways upgrade 0.2.0
```

Review the branch, push, open a PR. After it merges, everyone pulls and reinstalls. See
[How it works](how-it-works.md#upgrades).

## If something goes wrong

| Message | Meaning |
|---|---|
| `folkways: framework files were edited` | Someone changed a synced file. Move the change into `overlays/`, or rerun with `--force` to discard it |
| `folkways: team skills clash with new framework skills` | A team skill has a framework skill's name. Rename yours |
| `warning: AGENTS.md has no '@.folkways/instructions.md' line` | Agents won't see the framework rules. Put the line back |
| `folkways: 'doctor' is not built yet` | Expected in v0.1.0; same for `hook`, `review`, `run` |
