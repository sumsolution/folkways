#!/bin/sh
# Smoke-tests the release archives in <dist> (goreleaser's output) on the platform it runs on: installs with
# install.sh from the local files, checks the version, that a tampered checksum is refused, that init, sync
# and doctor work on a PATH with only the binary and the system dirs, and runs the hook contract suite
# against the installed binary when Go is on PATH (FOLKWAYS_HOOK_BUDGET_MS adds the latency check).
# Usage: sh scripts/smoke-release.sh dist
set -eu

[ $# -eq 1 ] || {
	echo "usage: $0 <dist>" >&2
	exit 2
}
dist=$(cd "$1" && pwd)
root=$(cd "$(dirname "$0")/.." && pwd)
step() { printf 'smoke: %s\n' "$*"; }
fail() {
	printf 'smoke: FAIL: %s\n' "$*" >&2
	exit 1
}

# metadata.json is one line of JSON; "version" is the release (or snapshot) version without the v.
version=$(sed -n 's/.*"version":"\([^"]*\)".*/\1/p' "$dist/metadata.json")
[ -n "$version" ] || fail "no version in $dist/metadata.json"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM
export GIT_AUTHOR_NAME=smoke GIT_AUTHOR_EMAIL=smoke@example.com GIT_COMMITTER_NAME=smoke \
	GIT_COMMITTER_EMAIL=smoke@example.com GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

step "install $version with install.sh"
FOLKWAYS_BASE_URL="file://$dist" FOLKWAYS_VERSION="$version" FOLKWAYS_INSTALL_DIR="$tmp/bin" sh "$root/install.sh"
bin=$tmp/bin/folkways

step "version"
got=$("$bin" version)
[ "$got" = "$version" ] || fail "folkways version printed $got, expected $version"

step "install.sh refuses an archive that doesn't match checksums.txt"
mkdir "$tmp/bad"
cp "$dist"/*.tar.gz "$dist/checksums.txt" "$tmp/bad/"
for f in "$tmp"/bad/*.tar.gz; do printf x >>"$f"; done
if FOLKWAYS_BASE_URL="file://$tmp/bad" FOLKWAYS_VERSION="$version" FOLKWAYS_INSTALL_DIR="$tmp/bad/bin" \
	sh "$root/install.sh" 2>"$tmp/bad/err"; then
	fail "install.sh installed despite a wrong checksum"
fi
grep -q 'checksum mismatch' "$tmp/bad/err" || fail "unexpected install.sh error: $(cat "$tmp/bad/err")"
[ ! -e "$tmp/bad/bin/folkways" ] || fail "install.sh left a binary after a checksum mismatch"

step "init, sync and doctor with only the binary and the system dirs on PATH"
(
	PATH="$tmp/bin:/usr/bin:/bin"
	cd "$tmp"
	folkways init team
	cd team
	folkways sync
	folkways doctor
	grep -q 'command -v folkways >/dev/null 2>&1 || ' .claude/settings.json ||
		fail ".claude/settings.json has no guarded hook command"
)

if command -v go >/dev/null 2>&1; then
	step "hook contract suite against the installed binary"
	(cd "$root/packages/folkways" && FOLKWAYS_HOOK_CMD="$bin hook" go test -count=1 ./contract)
else
	step "skipped the hook contract suite: no go on PATH"
fi

step "ok: folkways $version"
