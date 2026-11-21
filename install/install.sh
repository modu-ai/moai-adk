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
#
# The language of the conversation with the model (#2935 L1) comes from --language, else
# MOAI_LANGUAGE, else (at a terminal) a numbered menu, else the locale, else English. After the
# commands are linked the script runs `moai-cli language-set --home <code>` once; that never
# overwrites an earlier choice, and a moai-cli without it only costs a warning. Everything this
# script prints stays English whatever the language.
#
# A release archive carries the skill package and the base agents beside the binary (`skills/`,
# `agents/`); they go into the version folder next to `bin/`, where `moai update` reads them
# (SPEC-20261117-installer-skill-agents). Into the hosts (Claude Code, Codex, OMP) they go only
# through the product's own preview and confirmed apply, `moai-cli update --json` and then
# `moai-cli update --confirm <plan_digest> --json`: this script copies nothing into a host
# folder. It shows the folders the preview names, asks once at a terminal (--yes answers for a
# run without one) and then names the next steps for each host it found.

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
PART_TMP=""
cleanup() {
    if [ -n "$PART_TMP" ] && [ -d "$PART_TMP" ]; then rm -rf "$PART_TMP"; fi
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
  --version VERSION     Install a specific version, for example 4.2647.12 (default: latest)
  --install-dir DIR     Put the moai and moai-cli commands in DIR (default: <state home>/path)
  --channel CHANNEL     stable (default) or rc: rc also takes releases marked as pre-releases
  --language CODE       Language of your conversation with the model, for example en, ko, ja or
                        zh (default: asked at a terminal, else MOAI_LANGUAGE, else your locale,
                        else en). MoAI's own messages stay English.
  --yes                 Install the MoAI skill and agents into the hosts found without asking
                        (the preview is still shown); without a terminal and without --yes
                        nothing is written into your home folder's host settings
  -h, --help            Show this help message

Examples:
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --version 4.2647.12
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --channel rc
  curl -fsSL https://adk.mo.ai.kr/install.sh | sh -s -- --language ko
EOF
}

# A language code is BCP-47-shaped text only: 2-3 letters, then groups of 2-8 letters or digits
# after a `-`, at most 35 characters (#2935 L1). It is passed to moai-cli as written, so nothing
# else gets through (a newline would let a multi-line value slip past the one-line match, so it
# is refused first).
valid_language() {
    case "$1" in
        "" | *[!A-Za-z0-9-]*) return 1 ;;
    esac
    [ "${#1}" -le 35 ] || return 1
    printf '%s\n' "$1" | grep -Eq '^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$'
}

# The language part of the first of LC_ALL, LC_MESSAGES, LANG that is set (`ko_KR.UTF-8` -> ko),
# as written; nothing when it is C, POSIX or not a code. Always succeeds.
locale_language() {
    for loc in "${LC_ALL:-}" "${LC_MESSAGES:-}" "${LANG:-}"; do
        [ -n "$loc" ] || continue
        part="${loc%%[_.@]*}"
        case "$part" in C | POSIX) return 0 ;; esac
        if valid_language "$part"; then printf '%s' "$part"; fi
        return 0
    done
    return 0
}

# Where a question's answer is read from (SPEC-20261117-installer-tty): standard input when it is
# a terminal; else the session's terminal, /dev/tty, when it can be opened — the documented
# `curl ... | sh` gives this script on standard input while the person sits at that terminal, as
# rustup's installer does; else nowhere. A question is asked only when the output is a terminal
# too, so the person sees it. ANSWER_FROM is `stdin`, `tty` or empty (ask nothing).
answer_source() {
    ANSWER_FROM=""
    [ -t 1 ] || return 0
    if [ -t 0 ]; then
        ANSWER_FROM='stdin'
    elif (exec </dev/tty) 2>/dev/null; then
        ANSWER_FROM='tty'
    fi
}

# Reads one answer line into `answer` from ANSWER_FROM; fails at the end of input. /dev/tty is
# opened for this read only, so standard input (the rest of a piped script) is never read.
ask_line() {
    if [ "$ANSWER_FROM" = tty ]; then
        read -r answer </dev/tty
    else
        read -r answer
    fi
}

# Asks for the language at a terminal; the answer goes to LANGUAGE_CODE. A wrong answer is asked
# again, three answers at most, then the default is used; an empty answer or the end of input
# takes the default too.
choose_language() {
    default=$1
    echo ""
    echo "Which language should the model use when it talks with you?"
    echo "  1) English"
    echo "  2) 한국어"
    echo "  3) 日本語"
    echo "  4) 中文"
    tries=0
    while [ "$tries" -lt 3 ]; do
        tries=$((tries + 1))
        printf 'Choose 1-4 [default: %s]: ' "$default"
        if ! ask_line; then
            echo ""
            break
        fi
        case "$answer" in
            "") LANGUAGE_CODE="$default"; return 0 ;;
            1) LANGUAGE_CODE="en"; return 0 ;;
            2) LANGUAGE_CODE="ko"; return 0 ;;
            3) LANGUAGE_CODE="ja"; return 0 ;;
            4) LANGUAGE_CODE="zh"; return 0 ;;
            *) print_warning "Please type a number from 1 to 4." ;;
        esac
    done
    LANGUAGE_CODE="$default"
    print_info "Using $default."
}

# Settles LANGUAGE_CODE before anything is downloaded or written: --language, else
# MOAI_LANGUAGE, else the menu at a terminal, else the locale, else en.
resolve_language() {
    if [ -z "$LANGUAGE_CODE" ] && [ -n "${MOAI_LANGUAGE:-}" ]; then
        valid_language "$MOAI_LANGUAGE" || die "Not a language code: '$MOAI_LANGUAGE' (for example en, ko, ja, zh)"
        LANGUAGE_CODE="$MOAI_LANGUAGE"
    fi
    [ -z "$LANGUAGE_CODE" ] || return 0
    default=$(locale_language)
    [ -n "$default" ] || default="en"
    answer_source
    if [ -n "$ANSWER_FROM" ]; then
        choose_language "$default"
    else
        LANGUAGE_CODE="$default"
    fi
}

# Records the language once, with the installed moai-cli. It never overwrites a value that is
# already there, and a moai-cli that has no language-set (an older pinned version) or refuses
# the code costs a warning, not the install.
record_language() {
    if ! out=$("$DEST/bin/moai-cli" language-set --home "$LANGUAGE_CODE" </dev/null 2>/dev/null); then
        print_warning "Could not record the conversation language ($LANGUAGE_CODE): this moai-cli has no language-set or does not accept the code. The install is complete."
        return 0
    fi
    case "$out" in
        *'"written":true'*) print_success "Conversation language recorded: $LANGUAGE_CODE" ;;
        *'"reason":"exists"'*) print_info "A conversation language is already recorded; left as it is" ;;
        *'"reason":"no_home_state"'*) print_info "No MoAI state folder in your home yet; the conversation language was not recorded" ;;
        *) print_warning "The conversation language answer was not understood; it may not be recorded." ;;
    esac
}

# A version is SemVer-shaped text only: it becomes part of a URL and a folder name. A HeadVer
# version (`4.2647.12`, `{head}.{yearweek}.{build}`) is one; so is `2026.11.0-rc.1` of the earlier
# scheme.
valid_version() {
    case "$1" in
        "" | *[!0-9A-Za-z.+-]* | .* | *..*) return 1 ;;
    esac
    case "$1" in
        [0-9]*) return 0 ;;
    esac
    return 1
}

# The supported platforms (user decision 2026-11-17, SPEC-20261117-supported-platforms): macOS on
# Apple silicon, Linux x86_64 and Windows x86_64 (install.ps1). Anything else stops here, before a
# download or a write.
detect_target() {
    os=$(uname -s)
    arch=$(uname -m)
    machine="$arch"
    # An x86_64 shell under Rosetta 2 on Apple silicon reports x86_64; the machine is arm64.
    # The system's own sysctl, by absolute path: a `sysctl` earlier on PATH is not asked.
    if [ "$os" = "Darwin" ] && [ "$arch" = "x86_64" ] &&
        [ "$(/usr/sbin/sysctl -n sysctl.proc_translated 2>/dev/null || true)" = "1" ]; then
        machine="arm64"
    fi
    case "$os/$machine" in
        Darwin/arm64 | Darwin/aarch64) TARGET="aarch64-apple-darwin" ;;
        Linux/x86_64 | Linux/amd64)
            TARGET="x86_64-unknown-linux-gnu"
            if [ -n "${WSL_DISTRO_NAME:-}" ]; then
                print_info "Detected WSL environment: $WSL_DISTRO_NAME"
            fi
            ;;
        *)
            print_info "Supported platforms: macOS arm64 (Apple silicon), Linux x86_64, Windows x86_64 (install.ps1)"
            die "Unsupported platform: $os $arch"
            ;;
    esac
    print_success "Detected platform: $TARGET"
}

# fetch URL FILE — curl first, wget second; file:// works with curl (tests use it).
fetch() {
    if command -v curl >/dev/null 2>&1; then
        # No redirect may leave https (audit #2943 note 2); file:// is for the tests.
        curl -fsSL --proto '=https,file' --proto-redir '=https' --connect-timeout 10 --max-time 300 "$1" -o "$2"
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
            # A pre-release is told by the release's own mark; a `-` in the version marks one of
            # the earlier scheme only, because a HeadVer version has no suffix.
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
$(tr ',' '\n' <"$api_file" | tr '{' '\n' | tr '}' '\n' | sed -n -E 's/^[[:space:]]*"(tag_name|draft|prerelease)"[[:space:]]*:[[:space:]]*"?([^"]*)"?[[:space:]]*$/\1 \2/p' |
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
        # version marks a release candidate of the earlier scheme, and a HeadVer pre-release
        # (no suffix) cannot be told here, so the newest `moai-cli-v*` is taken. Pin a version
        # with --version when the REST list is out of reach (docs/release-checklist.md).
        pick_version <<EOF
$(tr '<' '\n' <"$feed_file" | tr '>' '\n' | tr '"' '\n' | sed -n "s|^.*/releases/tag/\\(${TAG_PREFIX}[^/?#]*\\)\$|\\1 false false|p")
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
    ARCHIVE_ROOT=$(dirname "$NEW_BIN")
}

# Copies the release parts the archive carries beside the binary (`skills/`, `agents/`) into the
# version folder $1, each under a temporary name renamed into place. A part the folder already
# has, or one the archive does not carry (an older release), is left alone.
add_parts() {
    for part in skills agents; do
        src="$ARCHIVE_ROOT/$part"
        if [ ! -d "$src" ] || [ -L "$src" ] || [ -e "$1/$part" ] || [ -L "$1/$part" ]; then
            continue
        fi
        PART_TMP="$1/.$part.install-$$"
        cp -R "$src" "$PART_TMP" || die "Could not copy $part/ into $1"
        mv "$PART_TMP" "$1/$part"
        PART_TMP=""
        PARTS_ADDED="$PARTS_ADDED $part/"
    done
}

# Asks the downloaded binary where its state lives and what `moai` on PATH is. Read only.
#
# Every `moai-cli` child below reads its standard input from /dev/null
# (SPEC-20261118-installer-stdin-null): under `curl ... | sh` this script itself is standard
# input, and a child that read it would eat the part not yet run. Answers are read by `ask_line`,
# from standard input or /dev/tty, never by a child.
inspect() {
    out=$("$NEW_BIN" install-inspect </dev/null 2>/dev/null) || die "The downloaded moai-cli does not run on this machine"
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
            # A folder an earlier installer placed holds bin/ only: add what this archive carries.
            PARTS_ADDED=""
            add_parts "$DEST"
            [ -z "$PARTS_ADDED" ] || print_success "Added to $DEST:$PARTS_ADDED"
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
    PARTS_ADDED=""
    add_parts "$STAGE_DIR"
    mv "$STAGE_DIR" "$DEST"
    STAGE_DIR=""
    print_success "Installed to: $DEST/bin"
    [ -z "$PARTS_ADDED" ] || print_success "With the release parts:$PARTS_ADDED"
}

# The hosts this machine has: the host's command on PATH, or its folder in HOME.
detect_hosts() {
    HOSTS=""
    for host in claude codex omp; do
        if command -v "$host" >/dev/null 2>&1 || [ -d "$HOME/.$host" ]; then
            HOSTS="$HOSTS $host"
        fi
    done
}

# A failed `moai-cli update`: the binary stays installed, and the product's own message is shown.
update_failed() {
    msg=$(printf '%s' "$1" | sed -n 's/.*"message":"\([^"]*\)".*/\1/p' | head -n 1)
    [ -n "$msg" ] || msg=$(head -n 1 "$TMP_DIR/update.err" 2>/dev/null || true)
    print_warning "The MoAI skill and agents were not installed: ${msg:-moai-cli update did not succeed}"
    print_info "The moai command is installed; try again later with: moai update"
}

# Installs the skill and the agents into the hosts found, through `moai-cli update` run from the
# version folder (so the hooks of the project the user stands in are not part of the plan).
install_skill() {
    SKILL_READY=""
    detect_hosts
    echo ""
    if [ -z "$HOSTS" ]; then
        print_warning "No Claude Code, Codex or OMP found (no claude, codex or omp command, no ~/.claude, ~/.codex or ~/.omp)."
        print_info "Install one of them, then install the MoAI skill and agents with: moai update"
        return 0
    fi
    print_info "Previewing the MoAI skill and agents for your hosts (writes nothing):"
    # Both `moai-cli update` calls keep their stderr in a file: on the person's terminal, what
    # the product prints there (or a question it asks) would wait unseen behind this script's
    # own prompt (SPEC-20261117-installer-tty REQ-3).
    if ! preview=$(cd "$DEST" && "$DEST/bin/moai-cli" update --json </dev/null 2>"$TMP_DIR/update.err"); then
        update_failed "$preview"
        return 0
    fi
    case "$preview" in
        *'"ok":true'*) ;;
        *) update_failed "$preview"; return 0 ;;
    esac
    case "$preview" in
        *'"nothing_to_apply":true'*)
            print_success "The MoAI skill and agents are already installed and up to date"
            SKILL_READY=1
            return 0
            ;;
    esac
    digest=$(printf '%s' "$preview" | sed -n 's/.*"confirm_with":"moai update --confirm \(sha256:[0-9a-f]\{64\}\)".*/\1/p')
    if [ -z "$digest" ]; then
        update_failed '"message":"the preview named no plan to confirm"'
        return 0
    fi
    printf '%s' "$preview" | grep -o '"projection_root":"[^"]*"' | sed 's/^"projection_root":"\(.*\)"$/  skill:  \1/' || true
    printf '%s' "$preview" | grep -o '"dir":"[^"]*"' | sed 's/^"dir":"\(.*\)"$/  agents: \1/' || true
    echo "  and MoAI's sections in those hosts' settings files (moai update lists every file)"
    if [ -n "$YES" ]; then
        print_info "--yes: installing them"
    elif answer_source && [ -n "$ANSWER_FROM" ]; then
        printf 'Install them now? [y/N]: '
        if ! ask_line; then
            echo ""
            answer=""
        fi
        case "$answer" in
            y | Y | yes | Yes | YES) ;;
            *)
                print_info "Nothing was written to your home folder. Install them later with: moai update"
                return 0
                ;;
        esac
    else
        print_info "Nothing was written to your home folder (no terminal to ask at). Install them with: moai update (or run this installer with --yes)"
        return 0
    fi
    if ! out=$(cd "$DEST" && "$DEST/bin/moai-cli" update --confirm "$digest" --json </dev/null 2>"$TMP_DIR/update.err"); then
        update_failed "$out"
        return 0
    fi
    case "$out" in
        *'"ok":true'*'"applied":true'*)
            print_success "Installed the MoAI skill and agents"
            SKILL_READY=1
            ;;
        *) update_failed "$out" ;;
    esac
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
    "$DEST/bin/moai-cli" --version </dev/null || print_warning "moai-cli did not report its version"
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
    echo ""
    print_info "Next steps:"
    n=1
    if [ -z "$SKILL_READY" ]; then
        echo "  $n. Install the MoAI skill and agents (shows what it writes, then asks): moai update"
        n=$((n + 1))
    fi
    echo "  $n. Make a project folder and enter it: mkdir my-project && cd my-project"
    n=$((n + 1))
    echo "  $n. Start a session there, then set the project up inside it:"
    # One entry per host found, or all three when none was: the list is split on purpose.
    # shellcheck disable=SC2086
    for host in ${HOSTS:-claude codex omp}; do
        case "$host" in
            claude) echo "     Claude Code: moai cc (or moai glm; -l leader, -a agent), then /moai project" ;;
            codex) echo "     Codex: moai codex, then \$moai project" ;;
            omp) echo "     OMP: omp, then /skill:moai project" ;;
        esac
    done
    echo ""
    print_info "Documentation: $REPO_URL"
}

main() {
    VERSION=""
    INSTALL_DIR="${MOAI_INSTALL_DIR:-}"
    CHANNEL="stable"
    LANGUAGE_CODE=""
    YES=""
    HOSTS=""
    SKILL_READY=""
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
            --language)
                [ $# -ge 2 ] || die "--language needs a value"
                valid_language "$2" || die "Not a language code: '$2' (for example en, ko, ja, zh)"
                LANGUAGE_CODE="$2"
                shift 2
                ;;
            --yes)
                YES=1
                shift
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
    resolve_language
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
    record_language
    install_skill
    finish
}

main "$@"
