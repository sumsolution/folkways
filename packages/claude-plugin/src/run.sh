#!/bin/sh
# Runs one folkways hook for the Claude Code plugin: run.sh <hook> < payload.json
#
# `instructions` prints the way of working as session context. Every other hook runs this platform's
# folkways-hook from hooks/bin/; FOLKWAYS_HOOK_CMD overrides it. The hook names skills as the plugin
# does (`folkways:start`). Synced team repos are left to their own settings; a repo without folkways.yaml
# gets only session-start, for the one-time setup offer. Always exits 0: a broken runner must not block the session.

hook=$1
root=${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}
project=${CLAUDE_PROJECT_DIR:-$PWD}

# A synced team repo runs these from its own .claude/settings.json and imports the instructions.
[ -f "$project/.folkways/manifest.json" ] && exit 0

# Without a folkways.yaml the plugin stays quiet: no instructions, and only session-start runs the binary,
# which decides whether to offer one-time setup (spec 2.3, 7.1).
configured=1
[ -f "$project/folkways.yaml" ] || configured=
[ -z "$configured" ] && [ "$hook" != session-start ] && exit 0

if [ "$hook" = instructions ]; then
  cat "$root/instructions.md"
  exit 0
fi

case $(uname -s) in Darwin) os=darwin ;; Linux) os=linux ;; *) os=unknown ;; esac
case $(uname -m) in x86_64 | amd64) cpu=x64 ;; arm64 | aarch64) cpu=arm64 ;; *) cpu=unknown ;; esac
bin=${FOLKWAYS_HOOK_CMD:-$root/hooks/bin/folkways-hook-$os-$cpu}

# Archives don't always keep the executable bit.
[ -f "$bin" ] && [ ! -x "$bin" ] && chmod +x "$bin" 2>/dev/null

if [ ! -x "$bin" ]; then
  # Without the binary nothing parses folkways.yaml; a line test is enough to honor `session-start: false`.
  [ -n "$configured" ] && [ "$hook" = session-start ] &&
    ! grep -Eq '^[^#]*session-start[[:space:]]*:[[:space:]]*false' "$project/folkways.yaml" &&
    echo "folkways: no hook binary for $(uname -s)-$(uname -m), so git-guard, context bands and precompact are off. Follow the \`folkways:start\` skill for this session."
  exit 0
fi

FOLKWAYS_SKILL_NAMESPACE=folkways "$bin" "$hook"
exit 0
