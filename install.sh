#!/bin/sh
# Installs the folkways binary from a GitHub release: macOS and Linux, amd64 and arm64.
#   curl -fsSL https://github.com/sumsolution/folkways/releases/latest/download/install.sh | sh
# Environment:
#   FOLKWAYS_VERSION      the version to install, e.g. 0.2.0 (default: the latest release)
#   FOLKWAYS_INSTALL_DIR  where to put the binary (default: ~/.local/bin)
#   FOLKWAYS_BASE_URL     where the archives and checksums.txt are (tests; needs FOLKWAYS_VERSION)
# It checks the archive against the release's checksums.txt and refuses on a mismatch.
set -eu

repo=https://github.com/sumsolution/folkways

say() { printf 'folkways install: %s\n' "$*" >&2; }
die() {
	say "$*"
	exit 1
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		die "needs sha256sum or shasum to check the download"
	fi
}

# The whole script is in main, so a download cut short runs nothing.
main() {
	command -v curl >/dev/null 2>&1 || die "needs curl"
	command -v tar >/dev/null 2>&1 || die "needs tar"

	case "$(uname -s)" in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	*) die "no build for $(uname -s); with Go 1.26+: go install github.com/sumsolution/folkways/packages/folkways/cmd/folkways@latest" ;;
	esac
	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) die "no build for $(uname -m); with Go 1.26+: go install github.com/sumsolution/folkways/packages/folkways/cmd/folkways@latest" ;;
	esac

	version=${FOLKWAYS_VERSION:-}
	version=${version#v}
	if [ -z "$version" ]; then
		[ -z "${FOLKWAYS_BASE_URL:-}" ] || die "FOLKWAYS_BASE_URL needs FOLKWAYS_VERSION"
		# releases/latest redirects to releases/tag/v<version>; no API call, no token.
		latest=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$repo/releases/latest") || die "found no release at $repo/releases (or can't reach GitHub)"
		version=${latest##*/v}
		case "$version" in
		[0-9]*) ;;
		*) die "found no release at $repo/releases" ;;
		esac
	fi
	base=${FOLKWAYS_BASE_URL:-$repo/releases/download/v$version}
	dir=${FOLKWAYS_INSTALL_DIR:-$HOME/.local/bin}
	archive=folkways_${version}_${os}_${arch}.tar.gz

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT INT TERM
	say "downloading folkways $version for $os/$arch"
	curl -fsSL -o "$tmp/$archive" "$base/$archive" || die "can't download $base/$archive"
	curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" || die "can't download $base/checksums.txt"

	want=$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")
	[ -n "$want" ] || die "checksums.txt has no entry for $archive"
	got=$(sha256 "$tmp/$archive")
	[ "$got" = "$want" ] || die "checksum mismatch for $archive (expected $want, got $got); nothing installed"

	tar -xzf "$tmp/$archive" -C "$tmp" folkways
	mkdir -p "$dir"
	# Copy, then rename, so a running folkways is never half-written.
	cp "$tmp/folkways" "$dir/.folkways.new"
	chmod 755 "$dir/.folkways.new"
	mv -f "$dir/.folkways.new" "$dir/folkways"
	say "installed $("$dir/folkways" version) to $dir/folkways"

	case ":$PATH:" in
	*":$dir:"*) ;;
	*)
		say "$dir isn't on your PATH. Add it in your login shell's profile (~/.zprofile, ~/.bash_profile or ~/.profile):"
		say "  export PATH=\"$dir:\$PATH\""
		say "Hooks run in tools started from the Dock or a launcher too, which read the login shell's PATH."
		;;
	esac
}

main "$@"
