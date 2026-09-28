#!/usr/bin/env bash
# Tests for bin/folkways init and upgrade. Usage: tests/run.sh
# Each test builds a throwaway framework repo (tagged) and team repo under a temp dir.
set -uo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd -P)
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Isolate git from the user's config (signing, hooks, identity).
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.com
export GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.com

pass=0
fail=0
current=''

check() { # check <description> <command...>
  local desc=$1
  shift
  if "$@" >/dev/null 2>&1; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    printf 'FAIL  %s: %s\n' "$current" "$desc"
  fi
}
negate() { ! "$@"; }
refute() { local desc=$1; shift; check "$desc" negate "$@"; }
contains() { grep -qF -- "$2" "$1"; }

# new_framework <dir>: copy this repo's framework files, commit, tag v<VERSION>.
new_framework() {
  local fw=$1
  mkdir -p "$fw"
  cp -R "$ROOT/bin" "$ROOT/vendor" "$ROOT/templates" "$ROOT/VERSION" "$fw/"
  git -C "$fw" init -q -b main
  commit_tag "$fw"
}

commit_tag() { # commit_tag <fw>: commit everything and tag v<VERSION>
  git -C "$1" add -A
  git -C "$1" commit -q -m "release $(cat "$1/VERSION")"
  git -C "$1" tag "v$(cat "$1/VERSION")"
}

# release_0_2_0 <fw>: a new version with a skill, a hook, a workflow and new instructions.
release_0_2_0() {
  local fw=$1
  echo 0.2.0 >"$fw/VERSION"
  mkdir -p "$fw/vendor/skills/start/references"
  printf -- '---\nname: start\ndescription: test\n---\nbody\n' >"$fw/vendor/skills/start/SKILL.md"
  echo ref >"$fw/vendor/skills/start/references/a.md"
  printf '#!/bin/sh\nexit 0\n' >"$fw/vendor/hooks/session-start.sh"
  chmod +x "$fw/vendor/hooks/session-start.sh"
  echo 'name: audit' >"$fw/vendor/workflows/knowledge-audit.yml"
  echo 'New rule for 0.2.0.' >>"$fw/vendor/instructions.md"
  jq '.hooks.SessionStart = [{"hooks": [{"type": "command", "command": ".folkways/hooks/session-start.sh"}]}]' \
    "$fw/vendor/settings.hooks.json" >"$fw/s.json" && mv "$fw/s.json" "$fw/vendor/settings.hooks.json"
  commit_tag "$fw"
}

setup_case() { # setup_case <name>: fresh framework at $FW, team repo at $TEAM
  current=$1
  FW="$TMP/$1/fw"
  TEAM="$TMP/$1/team"
  new_framework "$FW"
  "$FW/bin/folkways" init "$TEAM" >/dev/null
}

upgrade() { (cd "$TEAM" && .folkways/bin/folkways upgrade "$@" --from "$FW"); }

# --- init --------------------------------------------------------------------

setup_case init
check "scaffolds team.yaml" test -f "$TEAM/team.yaml"
check "renames dot- templates" test -f "$TEAM/.gitignore"
check "strips .tmpl" test -f "$TEAM/CLAUDE.md"
check "CLAUDE.md imports AGENTS.md" contains "$TEAM/CLAUDE.md" '@AGENTS.md'
check "vendors the CLI" test -x "$TEAM/.folkways/bin/folkways"
check "writes the manifest" jq -e '.version == "0.1.0" and (.files | has(".folkways/bin/folkways"))' "$TEAM/.folkways/manifest.json"
check "AGENTS.md keeps team text" contains "$TEAM/AGENTS.md" '## Team notes'
check "AGENTS.md has the managed block" contains "$TEAM/AGENTS.md" '<!-- folkways:begin -->'
check "AGENTS.md has the instructions" contains "$TEAM/AGENTS.md" '# Way of working'
check "settings get the push ask rule" jq -e '.permissions.ask | index("Bash(git push:*)")' "$TEAM/.claude/settings.json"
check "commits on main" test "$(git -C "$TEAM" log --format=%s main)" = "chore(folkways): init at 0.1.0"
check "leaves a clean tree" test -z "$(git -C "$TEAM" status --porcelain)"
check "team CLI reports version" test "$("$TEAM/.folkways/bin/folkways" version)" = 0.1.0

current=init-refusals
mkdir -p "$TMP/nonempty" && touch "$TMP/nonempty/x"
refute "refuses a non-empty path" "$FW/bin/folkways" init "$TMP/nonempty"
refute "refuses a path inside the framework" "$FW/bin/folkways" init "$FW/team"
refute "refuses to run from a team repo" "$TEAM/.folkways/bin/folkways" init "$TMP/other"

# --- upgrade -----------------------------------------------------------------

setup_case upgrade
# Team customizations the upgrade must keep.
printf '\nTeam footer.\n' >>"$TEAM/AGENTS.md"
jq '.permissions.allow = ["Bash(ls:*)"] | .hooks.Stop = [{"hooks": [{"type": "command", "command": "scripts/team.sh"}]}]' \
  "$TEAM/.claude/settings.json" >"$TMP/s.json" && mv "$TMP/s.json" "$TEAM/.claude/settings.json"
mkdir -p "$TEAM/overlays/start" "$TEAM/.claude/skills/team-skill"
echo 'team overlay' >"$TEAM/overlays/start/OVERLAY.md"
echo 'team skill' >"$TEAM/.claude/skills/team-skill/SKILL.md"
git -C "$TEAM" add -A && git -C "$TEAM" commit -q -m "team changes"
release_0_2_0 "$FW"
check "upgrade succeeds" upgrade 0.2.0
check "is on the upgrade branch" test "$(git -C "$TEAM" branch --show-current)" = chore/folkways-0.2.0
check "commits the upgrade" test "$(git -C "$TEAM" log -1 --format=%s)" = "chore(folkways): upgrade to 0.2.0"
check "leaves a clean tree" test -z "$(git -C "$TEAM" status --porcelain)"
check "adds the new skill" test -f "$TEAM/.claude/skills/start/references/a.md"
check "adds the new hook, executable" test -x "$TEAM/.folkways/hooks/session-start.sh"
check "prefixes workflows" test -f "$TEAM/.github/workflows/folkways-knowledge-audit.yml"
check "updates the managed block" contains "$TEAM/AGENTS.md" 'New rule for 0.2.0.'
check "keeps team text before the block" contains "$TEAM/AGENTS.md" '## Team notes'
check "keeps team text after the block" contains "$TEAM/AGENTS.md" 'Team footer.'
check "one managed block" test "$(grep -c 'folkways:begin' "$TEAM/AGENTS.md")" = 1
check "keeps team overlays" contains "$TEAM/overlays/start/OVERLAY.md" 'team overlay'
check "keeps team skills" test -f "$TEAM/.claude/skills/team-skill/SKILL.md"
check "keeps team allow rules" jq -e '.permissions.allow == ["Bash(ls:*)"]' "$TEAM/.claude/settings.json"
check "keeps team hooks" jq -e '.hooks.Stop[0].hooks[0].command == "scripts/team.sh"' "$TEAM/.claude/settings.json"
check "adds framework hooks" jq -e '.hooks.SessionStart[0].hooks[0].command == ".folkways/hooks/session-start.sh"' "$TEAM/.claude/settings.json"
check "no duplicate ask rule" jq -e '(.permissions.ask | length) == 1' "$TEAM/.claude/settings.json"
check "manifest lists the skill" jq -e '.version == "0.2.0" and .skills == ["start"]' "$TEAM/.folkways/manifest.json"

current=upgrade-removes
echo 0.3.0 >"$FW/VERSION"
rm -rf "$FW/vendor/skills/start" "$FW/vendor/hooks/session-start.sh"
jq 'del(.hooks.SessionStart)' "$FW/vendor/settings.hooks.json" >"$FW/s.json" && mv "$FW/s.json" "$FW/vendor/settings.hooks.json"
commit_tag "$FW"
check "upgrade succeeds" upgrade 0.3.0
refute "removes a dropped skill" test -e "$TEAM/.claude/skills/start"
refute "removes a dropped hook" test -e "$TEAM/.folkways/hooks/session-start.sh"
check "removes the dropped hook entry" jq -e '.hooks | has("SessionStart") | not' "$TEAM/.claude/settings.json"
check "still keeps team hooks" jq -e '.hooks.Stop | length == 1' "$TEAM/.claude/settings.json"

setup_case upgrade-noop
check "same version is a no-op" upgrade 0.1.0
check "stays on main" test "$(git -C "$TEAM" branch --show-current)" = main
refute "leaves no branch behind" git -C "$TEAM" rev-parse -q --verify refs/heads/chore/folkways-0.1.0

setup_case upgrade-dirty
release_0_2_0 "$FW"
echo x >"$TEAM/scratch.md"
refute "refuses a dirty tree" upgrade 0.2.0
check "stays on main" test "$(git -C "$TEAM" branch --show-current)" = main

setup_case upgrade-edited
release_0_2_0 "$FW"
echo '# hack' >>"$TEAM/.folkways/bin/folkways"
git -C "$TEAM" commit -qam "edit framework file"
refute "stops on an edited framework file" upgrade 0.2.0
check "names the file" sh -c "cd '$TEAM' && .folkways/bin/folkways upgrade 0.2.0 --from '$FW' 2>&1 | grep -q 'modified  .folkways/bin/folkways'"
check "stays on main" test "$(git -C "$TEAM" branch --show-current)" = main
check "--force discards the edit" upgrade 0.2.0 --force
refute "edit is gone" contains "$TEAM/.folkways/bin/folkways" '# hack'

setup_case upgrade-block
release_0_2_0 "$FW"
sed -i.bak 's/^# Way of working$/# Our way/' "$TEAM/AGENTS.md" && rm "$TEAM/AGENTS.md.bak"
git -C "$TEAM" commit -qam "edit managed block"
refute "stops on an edited managed block" upgrade 0.2.0

setup_case upgrade-added
release_0_2_0 "$FW"
upgrade 0.2.0 >/dev/null
git -C "$TEAM" checkout -q main && git -C "$TEAM" merge -q --ff-only chore/folkways-0.2.0
echo extra >"$TEAM/.claude/skills/start/extra.md"
git -C "$TEAM" add -A && git -C "$TEAM" commit -qm "add file to framework skill"
echo 0.2.1 >"$FW/VERSION" && commit_tag "$FW"
refute "stops on a file added to a framework skill" upgrade 0.2.1

setup_case upgrade-clash
mkdir -p "$TEAM/.claude/skills/start" && echo mine >"$TEAM/.claude/skills/start/SKILL.md"
git -C "$TEAM" add -A && git -C "$TEAM" commit -qm "team skill named start"
release_0_2_0 "$FW"
refute "refuses a team skill clash" upgrade 0.2.0
check "team skill untouched" contains "$TEAM/.claude/skills/start/SKILL.md" mine

setup_case upgrade-missing-tag
refute "fails on an unknown version" upgrade 9.9.9
check "stays on main" test "$(git -C "$TEAM" branch --show-current)" = main

printf '%d passed, %d failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
