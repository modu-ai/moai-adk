#!/bin/sh
# MoAI-ADK installer (Rust edition) — served at https://adk.mo.ai.kr/install.sh (#2935).
#
#   curl -fsSL https://adk.mo.ai.kr/install.sh | sh
#
# Downloads the `moai-cli` release archive for this machine, verifies its SHA-256 before
# anything is written (a missing, malformed or different checksum stops the install), and
# places it in the version folder the binary itself names:
#
#   <state home>/versions/<version>/bin/moai-cli   (+ bin/moai, a hard link to it)
#   <link dir>/moai-cli, <link dir>/moai          (symbolic links; default <state home>/path)
#
# <state home> is the `home_state_root` that `moai-cli install-inspect` reports, so the one
# rule for it lives in the binary (#2871 I2), not here. The shell configuration and PATH are
# never edited, and an existing Go-edition `moai` is never touched: the script prints what to
# run next instead. POSIX sh, so it runs under `sh` and `bash` alike.

set -eu

REPO_URL="https://github.com/modu-ai/moai-adk"
RELEASE_BASE="${MOAI_RELEASE_BASE:-$REPO_URL/releases/download}"
RELEASE_FEED="${MOAI_RELEASE_FEED:-$REPO_URL/releases.atom}"
RELEASE_API="${MOAI_RELEASE_API:-https://api.github.com/repos/modu-ai/moai-adk/releases?per_page=100}"
TAG_PREFIX="moai-cli-v"

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
    RED=$(printf '\033[0;31m'); GREEN=$(printf '\033[0;32m'); YELLOW=$(printf '\033[1;33m')
    BLUE=$(printf '\033[0;34m'); NC=$(printf '\033[0m')
else
    RED=""; GREEN=""; YELLOW=""; BLUE=""; NC=""
fi

print_info() { printf '%s[INFO]%s %s\n' "$BLUE" "$NC" "$1"; }
print_success() { printf '%s[SUCCESS]%s %s\n' "$GREEN" "$NC" "$1"; }
print_warning() { printf '%s[WARNING]%s %s\n' "$YELLOW" "$NC" "$1"; }
print_error() { printf '%s[ERROR]%s %s\n' "$RED" "$NC" "$1" >&2; }

TMP_DIR=""
STAGE_DIR=""
cleanup() {
    if [ -n "$STAGE_DIR" ] && [ -d "$STAGE_DIR" ]; then rm -rf "$STAGE_DIR"; fi
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then rm -rf "$TMP_DIR"; fi
}
trap cleanup EXIT
trap 'exit 1' INT TERM

die() {
    print_error "$1"
    exit 1
}

usage() {
    cat <<EOF
Usage: install.sh [OPTIONS]

Options:
  --version VERSION     Install a specific version (default: latest)
  --install-dir DIR     Put the moai and moai-cli commands in DIR (default: <state home>/path)
  --channel CHANNEL     stable (default) or rc: rc also takes release candidates
  -h, --help            Show this help message

Examples:
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --version 2026.11.0
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --channel rc
EOF
}

# A version is SemVer-shaped text only: it becomes part of a URL and a folder name.
valid_version() {
    case "$1" in
        "" | *[!0-9A-Za-z.+-]* | .* | *..*) return 1 ;;
    esac
    case "$1" in
        [0-9]*) return 0 ;;
    esac
    return 1
}

detect_target() {
    os=$(uname -s)
    arch=$(uname -m)
    case "$os" in
        Darwin) os_part="apple-darwin" ;;
        Linux)
            os_part="unknown-linux-gnu"
            if [ -n "${WSL_DISTRO_NAME:-}" ]; then
                print_info "Detected WSL environment: $WSL_DISTRO_NAME"
            fi
            ;;
        *)
            print_info "Supported operating systems: macOS, Linux (Windows: install.ps1)"
            die "Unsupported OS: $os"
            ;;
    esac
    case "$arch" in
        x86_64 | amd64) arch_part="x86_64" ;;
        arm64 | aarch64) arch_part="aarch64" ;;
        *)
            print_info "Supported architectures: x86_64 (amd64), arm64 (aarch64)"
            die "Unsupported architecture: $arch"
            ;;
    esac
    TARGET="$arch_part-$os_part"
    print_success "Detected platform: $TARGET"
}

# fetch URL FILE — curl first, wget second; file:// works with curl (tests use it).
fetch() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --connect-timeout 10 --max-time 300 "$1" -o "$2"
    elif command -v wget >/dev/null 2>&1; then
        wget -q --timeout=30 --tries=2 "$1" -O "$2"
    else
        die "Neither curl nor wget found. Please install one of them."
    fi
}

# Takes the first candidate line `<tag> <draft> <prerelease>` (newest first) that is a
# `moai-cli-v<version>` release for the channel. Sets VERSION, or leaves it empty.
pick_version() {
    while read -r tag draft pre; do
        case "$tag" in "$TAG_PREFIX"*) ;; *) continue ;; esac
        [ "$draft" = "true" ] && continue
        v="${tag#"$TAG_PREFIX"}"
        valid_version "$v" || continue
        if [ "$CHANNEL" != "rc" ]; then
            [ "$pre" = "true" ] && continue
            case "$v" in *-*) continue ;; esac
        fi
        VERSION="$v"
        return 0
    done
    return 0
}

# The newest `moai-cli-v<version>` release. `/releases/latest` cannot be used: the Rust
# edition's final releases are published with `--latest=false` so the Go edition's updater
# keeps its latest (docs/release-checklist.md). First the REST list (100 newest, with the
# draft and prerelease fields); when it cannot be read (no network to the API, or the
# unauthenticated limit of 60 requests an hour is spent), the releases feed, which is not
# rate limited but holds only the 10 newest entries (audit #2943 B1).
latest_version() {
    VERSION=""
    api_file="$TMP_DIR/releases.json"
    if fetch "$RELEASE_API" "$api_file" 2>/dev/null; then
        # One `"key": value` per line; a release's tag_name, draft and prerelease come in
        # that order, so each prerelease line closes one candidate. Escaped quotes inside
        # text (\"tag_name\") do not match.
        pick_version <<EOF
$(tr ',{}' '\n\n\n' <"$api_file" | sed -n -E 's/^[[:space:]]*"(tag_name|draft|prerelease)"[[:space:]]*:[[:space:]]*"?([^"]*)"?[[:space:]]*$/\1 \2/p' |
    awk '$1 == "tag_name" { t = $2; d = "" } $1 == "draft" { d = $2 } $1 == "prerelease" && t != "" { print t, d, $2; t = "" }')
EOF
    fi
    if [ -z "$VERSION" ]; then
        feed_file="$TMP_DIR/releases.atom"
        if ! fetch "$RELEASE_FEED" "$feed_file"; then
            print_info "You can install a specific version: install.sh --version <version>"
            die "Could not read the release list from GitHub"
        fi
        # The feed lists published releases only and has no prerelease field: a `-` in the
        # version marks a release candidate.
        pick_version <<EOF
$(tr '<>"' '\n\n\n' <"$feed_file" | sed -n "s|^.*/releases/tag/\\($TAG_PREFIX[^/?#]*\\)\$|\\1 false false|p")
EOF
    fi
    if [ -z "$VERSION" ]; then
        if [ "$CHANNEL" = "stable" ]; then
            print_info "No stable release yet. For a release candidate: install.sh --channel rc"
        fi
        die "No $TAG_PREFIX* release found for channel $CHANNEL"
    fi
    print_success "Latest version ($CHANNEL): $VERSION"
}

sha256_of() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        die "sha256sum/shasum not found: the download cannot be verified, so nothing is installed"
    fi
}

download_and_verify() {
    asset="moai-cli-$TARGET.tar.xz"
    url="$RELEASE_BASE/$TAG_PREFIX$VERSION/$asset"
    ARCHIVE="$TMP_DIR/$asset"
    print_info "Downloading from: $url"
    fetch "$url" "$ARCHIVE" || die "Download failed"
    fetch "$url.sha256" "$ARCHIVE.sha256" ||
        die "Could not download the checksum $asset.sha256: nothing is installed without it"
    expected=$(awk 'NR == 1 {print tolower($1)}' "$ARCHIVE.sha256")
    case "$expected" in
        *[!0-9a-f]* | "") die "The checksum file is not a SHA-256 digest: nothing is installed" ;;
    esac
    [ "${#expected}" -eq 64 ] || die "The checksum file is not a SHA-256 digest: nothing is installed"
    actual=$(sha256_of "$ARCHIVE")
    if [ "$expected" != "$actual" ]; then
        print_error "Expected: $expected"
        print_error "Actual:   $actual"
        die "Checksum mismatch!"
    fi
    print_success "Checksum verified"
}

extract() {
    mkdir "$TMP_DIR/x"
    tar -xJf "$ARCHIVE" -C "$TMP_DIR/x" ||
        die "Failed to extract the archive (tar needs xz support: install xz-utils)"
    NEW_BIN=$(find "$TMP_DIR/x" -maxdepth 2 -type f -name moai-cli | head -n 1)
    [ -n "$NEW_BIN" ] || die "moai-cli not found in the archive"
    chmod 755 "$NEW_BIN"
}

# Asks the downloaded binary where its state lives and what `moai` on PATH is. Read only.
inspect() {
    out=$("$NEW_BIN" install-inspect 2>/dev/null) || die "The downloaded moai-cli does not run on this machine"
    STATE_HOME=$(printf '%s' "$out" | sed -n 's/.*"home_state_root":"\([^"\\]*\)".*/\1/p')
    case "$STATE_HOME" in
        /*) ;;
        *) die "The downloaded moai-cli did not name an absolute state folder: nothing is installed" ;;
    esac
    GO_ON_PATH=""
    case "$out" in
        *'"path_first_moai":{'*'"generation":"go"'*) GO_ON_PATH=1 ;;
    esac
}

place() {
    versions="$STATE_HOME/versions"
    DEST="$versions/$VERSION"
    if [ -e "$DEST" ] || [ -L "$DEST" ]; then
        if [ -f "$DEST/bin/moai-cli" ] && [ "$(sha256_of "$DEST/bin/moai-cli")" = "$(sha256_of "$NEW_BIN")" ]; then
            print_success "$VERSION is already installed: $DEST"
            return
        fi
        print_error "$DEST already exists with different contents; it is left as it is."
        die "Remove or rename that folder yourself if you want this download in its place"
    fi
    mkdir -p "$versions"
    STAGE_DIR=$(mktemp -d "$versions/.install-$VERSION.XXXXXX")
    mkdir "$STAGE_DIR/bin"
    cp "$NEW_BIN" "$STAGE_DIR/bin/moai-cli"
    chmod 755 "$STAGE_DIR/bin/moai-cli"
    ln "$STAGE_DIR/bin/moai-cli" "$STAGE_DIR/bin/moai"
    chmod 755 "$STAGE_DIR"
    mv "$STAGE_DIR" "$DEST"
    STAGE_DIR=""
    print_success "Installed to: $DEST/bin"
}

link_commands() {
    LINK_DIR="${INSTALL_DIR:-$STATE_HOME/path}"
    mkdir -p "$LINK_DIR"
    for name in moai-cli moai; do
        link="$LINK_DIR/$name"
        if [ -L "$link" ]; then
            current=$(readlink "$link")
            [ "$current" = "$DEST/bin/$name" ] && continue
            # Only a link this installer made (into a version folder) is replaced: a link to
            # anything else — the Go edition's `moai`, say — is someone else's.
            case "$current" in
                "$STATE_HOME/versions/"*) ;;
                *)
                    print_warning "$link points to $current, not a version this installer placed; left as it is"
                    continue
                    ;;
            esac
        elif [ -e "$link" ]; then
            print_warning "$link is not a link made by this installer; left as it is"
            continue
        fi
        # Made under a temporary name and renamed over the old link, so `moai` never goes
        # missing in between (rename is atomic within a folder).
        tmp_link="$LINK_DIR/.$name.install-$$"
        ln -s "$DEST/bin/$name" "$tmp_link"
        mv -f "$tmp_link" "$link"
        print_success "Linked: $link -> $DEST/bin/$name"
    done
}

finish() {
    "$DEST/bin/moai-cli" --version || print_warning "moai-cli did not report its version"
    if [ -n "$GO_ON_PATH" ]; then
        echo ""
        print_warning "The moai first on your PATH is the Go edition; it was not changed."
        print_info "Preview what moving to the Rust edition changes (writes nothing):"
        echo "  $DEST/bin/moai-cli update"
    fi
    case ":$PATH:" in
        *":$LINK_DIR:"*) ;;
        *)
            echo ""
            print_info "$LINK_DIR is not on your PATH. Add it to ~/.zshrc or ~/.bashrc:"
            echo "  export PATH=\"$LINK_DIR:\$PATH\""
            ;;
    esac
    echo ""
    print_success "Installation complete!"
    print_info "Documentation: $REPO_URL"
}

main() {
    VERSION=""
    INSTALL_DIR="${MOAI_INSTALL_DIR:-}"
    CHANNEL="stable"
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                [ $# -ge 2 ] || die "--version needs a value"
                # `v2026.11.0` and the tag `moai-cli-v2026.11.0` name the same version.
                VERSION="${2#"$TAG_PREFIX"}"
                VERSION="${VERSION#v}"
                valid_version "$VERSION" || die "Not a version: '$2'"
                shift 2
                ;;
            --install-dir)
                [ $# -ge 2 ] || die "--install-dir needs a value"
                INSTALL_DIR="$2"
                shift 2
                ;;
            --channel)
                [ $# -ge 2 ] || die "--channel needs a value"
                CHANNEL="$2"
                case "$CHANNEL" in stable | rc) ;; *) die "Unknown channel: $CHANNEL (stable or rc)" ;; esac
                shift 2
                ;;
            -h | --help)
                usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Releases are read over https only (file:// for the tests): the checksum comes from the
    # same place as the archive, so a plain-http source would make the check meaningless.
    for url in "$RELEASE_BASE" "$RELEASE_FEED" "$RELEASE_API"; do
        case "$url" in
            https://* | file://*) ;;
            *) die "Release sources must be https:// (or file://): $url" ;;
        esac
    done

    echo ""
    echo "MoAI-ADK installer (Rust edition)"
    echo ""
    detect_target
    TMP_DIR=$(mktemp -d)
    if [ -z "$VERSION" ]; then
        latest_version
    else
        print_info "Installing version: $VERSION"
    fi
    download_and_verify
    extract
    inspect
    place
    link_commands
    finish
}

main "$@"
