#!/usr/bin/env bash
# navigator-regen.sh — Project Navigator regeneration (deterministic core).
#
# Reads the project's SPEC registry (.moai/specs/SPEC-*/spec.md), each SPEC's
# progress.md, and git log for provenance; writes three living documents under
# .moai/project/navigator/ (navigator.md, capability-map.md, progress-map.md)
# plus a last-regen-commit sentinel at .moai/state/navigator/last-regen-commit.txt.
#
# Design properties:
#   * Self-contained bash (git + awk + sed + grep; NO jq, NO moai binary). This
#     keeps the script independent of the moai-binary resolution chain.
#   * Language-neutral: no Go-specific assumption. Inputs are SPEC registry +
#     git log, both project-agnostic.
#   * Atomic writes: each file is written <file>.tmp then mv'd into place, so a
#     concurrent reader observes either the old or the new version, never a
#     partial. The last-regen-commit sentinel is written LAST.
#   * Fail-open on malformed SPEC frontmatter: the offending row is skipped, a
#     warning is appended to .moai/logs/navigator-warnings.log, and regeneration
#     continues.
#   * Idempotent: provenance is sourced from git log (commit SHA + committer
#     date), NOT wall-clock, so two runs on the same commit produce byte-identical
#     output.
#
# Boundary: touches ONLY .moai/project/navigator/ + .moai/state/navigator/ +
# .moai/logs/navigator-warnings.log (write set) and reads ONLY .moai/specs/ +
# git log. It NEVER reads or writes any LSEL surface.
#
# Optional env vars:
#   CLAUDE_PROJECT_DIR  project root (defaults to $PWD).
#   NAVIGATOR_PRE_RENAME_BARRIER  path; when set, the script writes "ready" to
#       this path after creating each <file>.tmp and blocks (poll loop) until
#       the file is removed before the mv lands. Test-only synchronized barrier
#       for the atomic-rename fixture.
#
# Exit codes: 0 always (fail-open). Errors are logged, never fatal to the caller.
set -u

ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
NAV_DIR="$ROOT/.moai/project/navigator"
STATE_DIR="$ROOT/.moai/state/navigator"
WARN_LOG="$ROOT/.moai/logs/navigator-warnings.log"
SPECS_DIR="$ROOT/.moai/specs"

mkdir -p "$NAV_DIR" "$STATE_DIR" "$(dirname "$WARN_LOG")" 2>/dev/null || true

# is_git_repo returns 0 if ROOT is inside a git work tree.
is_git_repo() {
    [ -n "$(git -C "$ROOT" rev-parse --is-inside-work-tree 2>/dev/null)" ]
}

# head_or_empty echoes HEAD sha or empty string when not a repo / no commits.
head_or_empty() {
    if is_git_repo; then
        git -C "$ROOT" rev-parse HEAD 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# file_provenance <path> echoes "<sha>\t<iso8601>" from git log for the path's
# last commit. Empty fields when the path is untracked or not in a repo.
file_provenance() {
    local p="$1"
    if is_git_repo; then
        # --format=%H%x09%cI  (sha TAB committer-date-ISO-8601)
        local line
        line="$(git -C "$ROOT" log -1 --format='%H%x09%cI' -- "$p" 2>/dev/null || true)"
        if [ -n "$line" ]; then
            printf '%s' "$line"
            return
        fi
    fi
    printf '%s\t%s' "" ""
}

# extract_frontmatter_field <file> <fieldname> — prints the value of a top-level
# YAML scalar `fieldname:` from the leading `---`-delimited frontmatter block.
# Strips surrounding double quotes. Returns 1 (nonzero) if the frontmatter is
# absent or the field missing — callers treat that as "malformed".
extract_frontmatter_field() {
    local file="$1" field="$2"
    # awk: print the first `---`-bounded block, then grep the field.
    local val
    val="$(awk -v f="$field" '
        /^---[[:space:]]*$/ { fm++; if (fm == 2) exit; next }
        fm == 1 && $0 ~ "^" f ":" {
            sub("^" f ":[[:space:]]*", "")
            gsub(/^"/, ""); gsub(/"$/, "")
            print; exit
        }
    ' "$file" 2>/dev/null || true)"
    if [ -z "$val" ]; then
        return 1
    fi
    printf '%s' "$val"
}

# has_valid_frontmatter <file> — returns 0 if the file begins with `---` and
# closes the frontmatter block with a second `---`.
has_valid_frontmatter() {
    local file="$1"
    local fmcount
    fmcount="$(awk '
        /^---[[:space:]]*$/ { c++; if (c == 2) { print c; exit } }
        END { if (c < 2) print c }
    ' "$file" 2>/dev/null || echo 0)"
    [ "${fmcount:-0}" -ge 2 ]
}

# barrier_wait — atomic-rename test fixture hook. If NAVIGATOR_PRE_RENAME_BARRIER is
# set AND a companion "<barrier>.armed" sentinel exists, write "ready" to the
# barrier path, consume the armed sentinel (so the barrier fires EXACTLY ONCE
# across the four atomic_write calls), and block until the barrier path is
# removed by the test driver. Production callers never set the env var, so this
# is a no-op in normal operation.
barrier_wait() {
    local barrier="${NAVIGATOR_PRE_RENAME_BARRIER:-}"
    [ -z "$barrier" ] && return 0
    local armed="${barrier}.armed"
    [ -e "$armed" ] || return 0
    rm -f "$armed" 2>/dev/null || true
    mkdir -p "$(dirname "$barrier")" 2>/dev/null || true
    printf 'ready' >"$barrier" 2>/dev/null || true
    local i=0
    while [ -e "$barrier" ] && [ "$i" -lt 1500 ]; do
        sleep 0.05 2>/dev/null || sleep 1
        i=$((i + 1))
    done
}

# atomic_write <dest_path> <content_via_stdin> — write stdin to <dest>.tmp,
# arm the barrier, then mv into place.
atomic_write() {
    local dest="$1"
    local tmp
    tmp="$(mktemp "${dest}.XXXXXX" 2>/dev/null)" || tmp="${dest}.tmp"
    cat >"$tmp" || return 1
    barrier_wait
    mv -f "$tmp" "$dest" 2>/dev/null || cp "$tmp" "$dest" 2>/dev/null || true
    rm -f "$tmp" 2>/dev/null || true
}

# --- collect SPEC rows -------------------------------------------------------

# Each discovered SPEC emits a TSV row:
#   id \t title \t status \t phase \t module \t spec_sha \t spec_iso8601
collect_spec_rows() {
    [ -d "$SPECS_DIR" ] || return 0
    local specfile id title status phase module prov sha ts
    for specfile in "$SPECS_DIR"/SPEC-*/spec.md; do
        [ -f "$specfile" ] || continue
        id=""
        if ! has_valid_frontmatter "$specfile"; then
            {
                printf '[%s] WARN skipped malformed SPEC (no closed frontmatter): %s\n' \
                    "$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo UNKNOWN)" "$specfile"
            } >>"$WARN_LOG" 2>/dev/null || true
            continue
        fi
        id="$(extract_frontmatter_field "$specfile" id)" || {
            {
                printf '[%s] WARN skipped SPEC missing required id field: %s\n' \
                    "$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo UNKNOWN)" "$specfile"
            } >>"$WARN_LOG" 2>/dev/null || true
            continue
        }
        title="$(extract_frontmatter_field "$specfile" title 2>/dev/null || echo "")"
        status="$(extract_frontmatter_field "$specfile" status 2>/dev/null || echo "")"
        phase="$(extract_frontmatter_field "$specfile" phase 2>/dev/null || echo "")"
        module="$(extract_frontmatter_field "$specfile" module 2>/dev/null || echo "")"
        prov="$(file_provenance "$specfile")"
        sha="${prov%%$'\t'*}"
        ts="${prov#*$'\t'}"
        # Skip rows that failed provenance AND are untracked — still emit with
        # empty provenance, since a draft SPEC may be brand-new. Row carries
        # the literal "(untracked)" sentinel in those fields so the reader can
        # see the claim is provisional.
        [ -z "$sha" ] && sha="(untracked)"
        [ -z "$ts" ] && ts="(untracked)"
        printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$id" "$title" "$status" "$phase" "$module" "$sha" "$ts"
    done
}

# --- main regeneration -------------------------------------------------------

HEAD_SHA="$(head_or_empty)"
# NOW_ISO is sourced from the HEAD commit's committer date (NOT wall-clock) so
# regeneration is idempotent — two runs on the same commit produce byte-identical
# output. When there is no git repo / no commits, it degrades to "(no-git)".
NOW_ISO="$(git -C "$ROOT" log -1 --format=%cI 2>/dev/null || echo '(no-git)')"

# Collect rows (sort by id for deterministic ordering).
ROWS_FILE="$(mktemp 2>/dev/null || echo /tmp/nav-rows.$$)"
trap 'rm -f "$ROWS_FILE" 2>/dev/null' EXIT
collect_spec_rows | sort -t$'\t' -k1,1 >"$ROWS_FILE" 2>/dev/null || true
ROW_COUNT="$(wc -l <"$ROWS_FILE" 2>/dev/null | tr -dc '0-9' || echo 0)"
[ -z "$ROW_COUNT" ] && ROW_COUNT=0

# --- navigator.md (entry point) ----------------------------------------------
{
    printf '# Project Navigator\n\n'
    printf '> Living project reorientation brief. Regenerated from the SPEC registry + git log.\n'
    printf '> Last regeneration: commit `%s` at %s\n\n' "${HEAD_SHA:-(no-git)}" "$NOW_ISO"
    if [ "$ROW_COUNT" -eq 0 ]; then
        printf '## Current frontier\n\n'
        printf 'no features tracked yet\n\n'
        printf 'The Navigator is initialized; no SPEC documents were found under '
        printf '`.moai/specs/`. Run `/moai project` after authoring your first SPEC to populate this brief.\n\n'
    else
        printf '## Current frontier\n\n'
        # Frontier = SPECs that are not yet completed (status != completed).
        awk -F'\t' '$3 != "completed" && $3 != "superseded" && $3 != "archived" && $3 != "rejected" { printf "- **%s** (%s) — %s — module `%s` — [%s @ %s]\n", $1, $3, $2, $5, substr($6,1,8), $7 }' "$ROWS_FILE"
        if [ -z "$(awk -F'\t' '$3 != "completed" && $3 != "superseded" && $3 != "archived" && $3 != "rejected" { print; exit }' "$ROWS_FILE")" ]; then
            printf '_All tracked SPECs are in a terminal state._\n'
        fi
        printf '\n## Next task\n\n'
        # Next task = the first non-terminal SPEC by sort order.
        next_line="$(awk -F'\t' '$3 != "completed" && $3 != "superseded" && $3 != "archived" && $3 != "rejected" { print; exit }' "$ROWS_FILE")"
        if [ -n "$next_line" ]; then
            nid="$(printf '%s' "$next_line" | awk -F'\t' '{print $1}')"
            printf 'Advance **%s** toward its next milestone. See `.moai/project/navigator/progress-map.md` for its frontier milestone.\n\n' "$nid"
        else
            printf 'No active SPEC. Consider opening a new SPEC via `/moai plan`.\n\n'
        fi
        printf 'Full entry brief: this file. Full progress rollup: `progress-map.md`.\n\n'
    fi
} | atomic_write "$NAV_DIR/navigator.md"

# --- capability-map.md (feature inventory) -----------------------------------
{
    printf '# Capability Map\n\n'
    printf '| spec-id | title | status | implementation-path | commit-sha | captured-at |\n'
    printf '|---------|-------|--------|----------------------|------------|-------------|\n'
    if [ "$ROW_COUNT" -eq 0 ]; then
        printf '| _(none)_ | | | | | |\n'
    else
        awk -F'\t' '{ printf "| %s | %s | %s | %s | %s | %s |\n", $1, $2, $3, $5, $6, $7 }' "$ROWS_FILE"
    fi
} | atomic_write "$NAV_DIR/capability-map.md"

# --- progress-map.md (per-SPEC rollup) ---------------------------------------
{
    printf '# Progress Map\n\n'
    printf '| spec-id | status | phase | last-commit-sha | last-commit-at | frontier-milestone |\n'
    printf '|---------|--------|-------|-----------------|----------------|--------------------|\n'
    if [ "$ROW_COUNT" -eq 0 ]; then
        printf '| _(none)_ | | | | | |\n'
    else
        # Frontier milestone: derived from progress.md if present (best-effort,
        # non-fatal), else the SPEC status itself.
        while IFS=$'\t' read -r id title status phase module sha ts; do
            fmilestone="$status"
            progfile="$SPECS_DIR/$id/progress.md"
            if [ -f "$progfile" ]; then
                # Pull the highest-numbered M<n> token referenced as "Milestone".
                m="$(grep -oE 'Milestone M[0-9]+' "$progfile" 2>/dev/null | tail -1 | sed 's/Milestone //')"
                [ -n "$m" ] && fmilestone="$m"
            fi
            printf '| %s | %s | %s | %s | %s | %s |\n' "$id" "$status" "$phase" "$sha" "$ts" "$fmilestone"
        done <"$ROWS_FILE"
    fi
} | atomic_write "$NAV_DIR/progress-map.md"

# --- last-regen-commit sentinel (written LAST) -------------------------------
{
    printf '%s\n' "${HEAD_SHA:-(no-git)}"
} | atomic_write "$STATE_DIR/last-regen-commit.txt"

exit 0
