#!/bin/sh
# MoAI installer for the Rust generation (`moai`, `moai-cli`), Issue #2871 I1-a.
#
#   curl -fsSL https://adk.mo.ai.kr/install.sh | sh
#   curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --version 0.1.0
#   sh install.sh [--version <v>] [--channel stable|rc]
#
# It finds the newest `moai-cli-v*` GitHub release (drafts never; pre-releases only with
# `--channel rc`), downloads `moai-cli-<target>.tar.xz` and its `.sha256`, stops when the
# checksum is missing or different, and unpacks into `<state home>/versions/<v>/bin/`
# with links in `<state home>/path/`. It never edits a shell file or PATH; it prints the
# line to add instead.
#
# Environment:
#   MOAI_VERSION       pin a version (same as --version)
#   MOAI_STATE_HOME    where versions/ and path/ go (default: see state_home)
#   MOAI_RELEASE_BASE  release mirror: <base>/releases.json (GitHub list shape) and
#                      <base>/download/<tag>/<asset>; https:// or file:// only
#
# Everything is in functions and `main` runs on the last line, so a download cut short
# runs nothing.

set -eu

REPO="modu-ai/moai-adk"
TAG_PREFIX="moai-cli-v"

say() {
    printf '%s\n' "$*"
}

die() {
    printf 'moai install: error: %s\n' "$*" >&2
    exit 1
}

usage() {
    cat <<'EOF'
Usage: install.sh [--version <v>] [--channel stable|rc]

  --version <v>     install this version (also MOAI_VERSION); "0.1.0", "v0.1.0" or
                    "moai-cli-v0.1.0"
  --channel <name>  stable (default): newest release that is not a pre-release
                    rc: newest release, pre-releases included
  -h, --help        show this help

No shell file and no PATH entry is changed.
EOF
    say "Installs into $(state_home 2>/dev/null || echo "\$MOAI_STATE_HOME"): versions/<v>/bin/ and path/."
}

usage_error() {
    printf 'moai install: %s\n' "$*" >&2
    usage >&2
    exit 2
}

# The one place that names the state home (stage I2 may move it).
state_home() {
    if [ -n "${MOAI_STATE_HOME:-}" ]; then
        printf '%s\n' "$MOAI_STATE_HOME"
        return
    fi
    [ -n "${HOME:-}" ] || die "HOME is not set; set HOME or MOAI_STATE_HOME"
    printf '%s\n' "$HOME/.moai-cli"
}

detect_target() {
    os=$(uname -s)
    arch=$(uname -m)
    case "$os" in
        Darwin)
            os_part=apple-darwin
            # An x86_64 shell under Rosetta on Apple silicon still gets the native build.
            if [ "$arch" = x86_64 ] && [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" = 1 ]; then
                arch=arm64
            fi
            ;;
        Linux) os_part=unknown-linux-gnu ;;
        *) die "unsupported system: $os (this script installs macOS and Linux builds)" ;;
    esac
    case "$arch" in
        arm64 | aarch64) arch=aarch64 ;;
        x86_64 | amd64) arch=x86_64 ;;
        *) die "unsupported CPU: $arch" ;;
    esac
    printf '%s-%s\n' "$arch" "$os_part"
}

# fetch <url> <file>: https:// and file:// only, no plain http, redirects stay on https.
fetch() {
    case "$1" in
        https://* | file://*) ;;
        *) die "refusing to download from $1 (only https:// or file://)" ;;
    esac
    if command -v curl >/dev/null 2>&1; then
        curl --proto '=https,file' --proto-redir '=https' -fsSL "$1" -o "$2"
    elif command -v wget >/dev/null 2>&1; then
        case "$1" in
            file://*) die "wget cannot read $1; install curl" ;;
        esac
        wget -q --https-only -O "$2" "$1"
    else
        die "curl or wget is needed to download"
    fi
}

sha256_of() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    elif command -v openssl >/dev/null 2>&1; then
        openssl dgst -sha256 -r "$1" | awk '{print $1}'
    else
        die "no sha256 tool (sha256sum, shasum or openssl); refusing to install an unverified file"
    fi
}

# A GitHub releases list (JSON) on stdin, one "<tag> <draft> <prerelease>" line per release
# on stdout, in list order (newest first). Reads the keys of each top-level object only, in
# any order and layout; text inside strings never counts. A missing flag prints "?".
release_rows() {
    awk '
    function flush() {
        if (tag != "") print tag, draft, pre
        tag = ""; draft = "?"; pre = "?"; key = ""
    }
    { buf = buf $0 "\n" }
    END {
        n = length(buf); depth = 0; i = 1; flush()
        while (i <= n) {
            c = substr(buf, i, 1)
            if (c == "\"") {
                j = i + 1
                while (j <= n) {
                    d = substr(buf, j, 1)
                    if (d == "\\") { j += 2; continue }
                    if (d == "\"") break
                    j++
                }
                s = substr(buf, i + 1, j - i - 1); i = j + 1
                if (depth == 2) {
                    if (key == "") key = s
                    else { if (key == "tag_name") tag = s; key = "" }
                }
                continue
            }
            if (c == "{" || c == "[") {
                if (depth == 1 && c == "{") flush()
                else if (depth == 2) key = ""
                depth++
            } else if (c == "}" || c == "]") {
                depth--
                if (depth == 1 && c == "}") flush()
            } else if (c == ",") {
                if (depth == 2) key = ""
            } else if (depth == 2 && key != "" && c ~ /[a-z0-9-]/) {
                j = i
                while (j <= n && substr(buf, j, 1) ~ /[A-Za-z0-9.+-]/) j++
                w = substr(buf, i, j - i); i = j
                if (key == "draft") draft = w
                else if (key == "prerelease") pre = w
                key = ""
                continue
            }
            i++
        }
    }'
}

# pick_version <rows file> <channel>: the first moai-cli release that is not a draft, and
# not a pre-release unless the channel is rc. An unknown flag counts as "no".
pick_version() {
    while read -r tag draft pre; do
        case "$tag" in
            "$TAG_PREFIX"*) ;;
            *) continue ;;
        esac
        [ "$draft" = false ] || continue
        if [ "$2" != rc ]; then
            [ "$pre" = false ] || continue
        fi
        printf '%s\n' "${tag#"$TAG_PREFIX"}"
        return 0
    done <"$1"
    return 1
}

# link_into <dir> <name> <target>: points <dir>/<name> at <target> in one rename. A regular
# file already there is kept as <name>.before-install-<UTC>, never deleted.
link_into() {
    link="$1/$2"
    if [ -d "$link" ]; then
        die "$link is a folder; move it away and run again"
    fi
    if [ -e "$link" ] && [ ! -L "$link" ]; then
        aside="$link.before-install-$(date -u +%Y%m%dT%H%M%SZ)"
        mv "$link" "$aside"
        say "kept the earlier $link as $aside"
    fi
    ln -s "$3" "$1/.$2.new.$$"
    mv -f "$1/.$2.new.$$" "$link"
}

main() {
    version="${MOAI_VERSION:-}"
    channel=stable
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                [ $# -ge 2 ] || usage_error "--version needs a value"
                version="$2"
                shift 2
                ;;
            --version=*)
                version="${1#--version=}"
                shift
                ;;
            --channel)
                [ $# -ge 2 ] || usage_error "--channel needs a value"
                channel="$2"
                shift 2
                ;;
            --channel=*)
                channel="${1#--channel=}"
                shift
                ;;
            -h | --help)
                usage
                exit 0
                ;;
            *) usage_error "unknown option: $1" ;;
        esac
    done
    case "$channel" in
        stable | rc) ;;
        *) usage_error "unknown channel: $channel (stable or rc)" ;;
    esac

    state=$(state_home)
    target=$(detect_target)
    if [ -n "${MOAI_RELEASE_BASE:-}" ]; then
        base="${MOAI_RELEASE_BASE%/}"
        list_url="$base/releases.json"
    else
        base="https://github.com/$REPO/releases"
        list_url="https://api.github.com/repos/$REPO/releases?per_page=100"
    fi

    work=$(mktemp -d "${TMPDIR:-/tmp}/moai-install.XXXXXX")
    trap 'rm -rf "$work"' EXIT
    trap 'exit 1' HUP INT TERM

    if [ -n "$version" ]; then
        version="${version#"$TAG_PREFIX"}"
        version="${version#v}"
    else
        fetch "$list_url" "$work/releases.json" || die "could not read the release list at $list_url"
        release_rows <"$work/releases.json" >"$work/rows"
        version=$(pick_version "$work/rows" "$channel") ||
            die "no $channel release tagged $TAG_PREFIX* found at $list_url"
    fi
    case "$version" in
        "" | .* | *[!0-9A-Za-z.+-]*) die "not a version: $version" ;;
    esac

    tag="$TAG_PREFIX$version"
    asset="moai-cli-$target.tar.xz"
    say "moai install: $tag for $target"
    fetch "$base/download/$tag/$asset" "$work/$asset" ||
        die "could not download $base/download/$tag/$asset"
    fetch "$base/download/$tag/$asset.sha256" "$work/$asset.sha256" ||
        die "the checksum file $asset.sha256 could not be downloaded; refusing to install an unverified archive"

    expected=$(awk 'NR == 1 { print tolower($1) }' "$work/$asset.sha256")
    case "$expected" in
        "" | *[!0-9a-f]*) die "the checksum file $asset.sha256 holds no sha256; refusing to install" ;;
    esac
    [ ${#expected} -eq 64 ] || die "the checksum file $asset.sha256 holds no sha256; refusing to install"
    actual=$(sha256_of "$work/$asset")
    [ "$actual" = "$expected" ] ||
        die "checksum mismatch for $asset (expected $expected, got $actual); nothing was installed"

    mkdir "$work/unpack"
    if ! tar -xJf "$work/$asset" -C "$work/unpack" 2>/dev/null; then
        command -v xz >/dev/null 2>&1 || die "cannot unpack $asset: tar has no xz support and xz is missing"
        xz -dc "$work/$asset" | tar -xf - -C "$work/unpack" || die "cannot unpack $asset"
    fi
    new="$work/unpack/moai-cli-$target/moai-cli"
    [ -f "$new" ] && [ ! -L "$new" ] || new="$work/unpack/moai-cli"
    [ -f "$new" ] && [ ! -L "$new" ] || die "$asset holds no moai-cli program"

    bin="$state/versions/$version/bin"
    mkdir -p "$state/versions" "$state/path"
    if [ -e "$state/versions/$version" ]; then
        # Re-run: the same build stays as it is; a different build is never overwritten.
        cmp -s "$new" "$bin/moai-cli" ||
            die "$state/versions/$version holds a different build; remove it or pick another version"
        if [ ! -e "$bin/moai" ]; then
            ln "$bin/moai-cli" "$bin/moai"
        fi
        [ "$bin/moai" -ef "$bin/moai-cli" ] ||
            die "$bin/moai is not the same file as $bin/moai-cli; remove it and run again"
        say "moai $version was already installed in $bin"
    else
        stage="$state/versions/.$version.new.$$"
        mkdir -p "$stage/bin"
        cp "$new" "$stage/bin/moai-cli"
        chmod 755 "$stage/bin/moai-cli"
        ln "$stage/bin/moai-cli" "$stage/bin/moai"
        mv "$stage" "$state/versions/$version"
        say "moai $version is installed in $bin"
    fi
    link_into "$state/path" moai "$bin/moai"
    link_into "$state/path" moai-cli "$bin/moai-cli"

    say ""
    say "No shell file and no PATH entry was changed. To run moai, add this line to your"
    say "shell setup (or run it in this shell):"
    say ""
    say "  export PATH=\"$state/path:\$PATH\""
    say ""
    say "Next: moai install-path-plan  (shows what it would change in PATH; changes nothing)"
}

main "$@"
