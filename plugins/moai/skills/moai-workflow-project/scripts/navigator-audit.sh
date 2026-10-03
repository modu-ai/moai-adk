#!/usr/bin/env bash
# navigator-audit.sh — Project Navigator audit (drift / completeness diff).
#
# Reads the project's design docs (.moai/project/{product,structure,tech}.md)
# and the capability-map produced by navigator-regen.sh, computes a
# bidirectional drift diff (Missing SPECs + Orphan SPECs + Matched), and
# writes audit-report.md + audit-report.json under .moai/project/navigator/.
#
# Design properties (parallel to navigator-regen.sh):
#   * Self-contained bash (git + awk + sed + grep; NO jq, NO yq, NO moai binary).
#     Keeps the script independent of the moai-binary resolution chain.
#   * Language-neutral: no Go-specific assumption. Inputs are design docs +
#     capability-map + SPEC registry + git log, all project-agnostic.
#   * Atomic writes: each file is written <file>.tmp then mv'd into place, so a
#     concurrent reader observes either the old or the new version, never partial.
#   * Fail-open on missing inputs: writes a minimal "no inputs available"
#     report naming the missing input, appends a warning to
#     .moai/logs/navigator-warnings.log, exit 0 (fail-open contract).
#   * Idempotent: provenance sourced from git log (commit SHA + committer
#     date), NOT wall-clock, so two runs on the same commit produce
#     byte-identical output (idempotence contract).
#   * Read-only over inputs: NEVER modifies design docs, capability-map,
#     progress-map, navigator.md, or SPEC frontmatter (read-only contract).
#
# Boundary (boundary non-overlap contract): touches ONLY .moai/project/navigator/audit-report.{md,json}
# + .moai/logs/navigator-warnings.log (write set). NEVER reads or writes any
# LSEL surface (.moai/lessons-inbox.jsonl, .moai/state/lsel/, memory/feedback_*.md,
# hns-lsel-*) or any SPEC-003 surface (tree-sitter grammars, AST helpers).
#
# Header-driven column resolution (header-driven contract): the capability-map column
# order is NOT frozen — the audit parses the header row and resolves columns
# by NAME (case-insensitive, -/_/space equivalent). Rows whose header lacks a
# required column (no feature/name column OR no spec-id column) are skipped
# with a warning.
#
# Optional env vars:
#   CLAUDE_PROJECT_DIR  project root (defaults to $PWD).
#
# Exit codes: 0 always (fail-open). Errors are logged, never fatal.
set -u

ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
NAV_DIR="$ROOT/.moai/project/navigator"
WARN_LOG="$ROOT/.moai/logs/navigator-warnings.log"
SPECS_DIR="$ROOT/.moai/specs"
PRODUCT_MD="$ROOT/.moai/project/product.md"
STRUCTURE_MD="$ROOT/.moai/project/structure.md"
TECH_MD="$ROOT/.moai/project/tech.md"
CAP_MAP="$NAV_DIR/capability-map.md"
OVERRIDE_FILE="$NAV_DIR/audit-known-matches.yaml"
AUDIT_MD="$NAV_DIR/audit-report.md"
AUDIT_JSON="$NAV_DIR/audit-report.json"

mkdir -p "$NAV_DIR" "$(dirname "$WARN_LOG")" 2>/dev/null || true

# --- helpers ---------------------------------------------------------------

is_git_repo() {
    [ -n "$(git -C "$ROOT" rev-parse --is-inside-work-tree 2>/dev/null)" ]
}

head_or_empty() {
    if is_git_repo; then
        git -C "$ROOT" rev-parse HEAD 2>/dev/null || echo ""
    else
        echo ""
    fi
}

now_iso() {
    if is_git_repo; then
        git -C "$ROOT" log -1 --format=%cI 2>/dev/null || echo '(no-git)'
    else
        echo '(no-git)'
    fi
}

warn() {
    local msg="$1"
    {
        printf '[%s] WARN %s\n' \
            "$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo UNKNOWN)" "$msg"
    } >>"$WARN_LOG" 2>/dev/null || true
}

atomic_write() {
    local dest="$1"
    local tmp
    tmp="$(mktemp "${dest}.XXXXXX" 2>/dev/null)" || tmp="${dest}.tmp"
    cat >"$tmp" || { rm -f "$tmp"; return 1; }
    mv -f "$tmp" "$dest" 2>/dev/null || cp "$tmp" "$dest" 2>/dev/null || true
    rm -f "$tmp" 2>/dev/null || true
}

# normalize <string>: lowercase, replace non-alphanumeric runs with single
# hyphen, collapse, strip leading/trailing hyphens.
normalize() {
    printf '%s' "$1" | tr '[:upper:]' '[:lower:]' \
        | tr -c 'a-z0-9\n' '-' \
        | tr -s '-' '-' \
        | sed 's/^-*//; s/-*$//'
}

# json_escape <string>: emit a JSON-safe double-quoted string. Escapes
# backslash, double-quote, and control chars. Assumes ASCII input (audit
# content is design-doc names + SPEC titles + paths).
json_escape() {
    printf '%s' "$1" | sed \
        -e 's/\\/\\\\/g' \
        -e 's/"/\\"/g' \
        -e 's/	/\\t/g'
}

# --- provenance ------------------------------------------------------------

HEAD_SHA="$(head_or_empty)"
NOW_ISO="$(now_iso)"

# --- work files ------------------------------------------------------------

WORK="$(mktemp -d 2>/dev/null || echo /tmp/nav-audit.$$)"
trap 'rm -rf "$WORK" 2>/dev/null' EXIT

DESIGN_TSV="$WORK/design.tsv"
CAP_TSV="$WORK/cap.tsv"
OV_MATCH_TSV="$WORK/ov-match.tsv"
OV_IGNORE="$WORK/ov-ignore.txt"
MATCHED_TSV="$WORK/matched.tsv"
MISSING_TSV="$WORK/missing.tsv"
ORPHAN_TSV="$WORK/orphan.tsv"
: >"$DESIGN_TSV"
: >"$CAP_TSV"
: >"$OV_MATCH_TSV"
: >"$OV_IGNORE"
: >"$MATCHED_TSV"
: >"$MISSING_TSV"
: >"$ORPHAN_TSV"

# --- fail-open detection ---------------------------------------------------

designs_present=0
for f in "$PRODUCT_MD" "$STRUCTURE_MD" "$TECH_MD"; do
    [ -f "$f" ] && { designs_present=1; break; }
done

cap_present=0
[ -f "$CAP_MAP" ] && cap_present=1

specs_present=0
if [ -d "$SPECS_DIR" ]; then
    # shellcheck disable=SC2012
    if ls "$SPECS_DIR"/SPEC-*/spec.md >/dev/null 2>&1; then
        specs_present=1
    fi
fi

if [ "$designs_present" -eq 0 ] || [ "$cap_present" -eq 0 ] || [ "$specs_present" -eq 0 ]; then
    missing=""
    [ "$designs_present" -eq 0 ] && missing="${missing} design-docs"
    [ "$cap_present" -eq 0 ] && missing="${missing} capability-map"
    [ "$specs_present" -eq 0 ] && missing="${missing} spec-registry"
    warn "audit fail-open — missing inputs:${missing}"
    {
        printf '# Navigator Audit Report\n\n'
        printf '> Audit at %s (commit %s).\n\n' "$NOW_ISO" "${HEAD_SHA:-(no-git)}"
        printf '## Missing SPECs\n\n'
        printf 'no inputs available —%s\n\n' "$missing"
        printf '## Orphan SPECs\n\n'
        printf '_no inputs available_\n\n'
        printf '## Matched\n\n'
        printf '_no inputs available_\n'
    } | atomic_write "$AUDIT_MD"
    {
        printf '{"audit_at":"%s","audit_commit":"%s","inputs":{"design_docs":[],"capability_map":null,"override_file":null},"missing":[],"orphan":[],"matched":[]}\n' \
            "$NOW_ISO" "${HEAD_SHA:-(no-git)}"
    } | atomic_write "$AUDIT_JSON"
    exit 0
fi

# --- step 1: extract design-intent features --------------------------------
#
# For each design doc, walk the heading stack. Inside any section whose
# heading text contains a sentinel phrase (Core Features / Features /
# Capabilities / Modules / Functionality / Subsystems / Components), extract
# bolded bullets `- **<name>**`. Output TSV: name \t file \t heading_path.

extract_design_features() {
    local file="$1"
    [ -f "$file" ] || return 0
    local rel
    rel="$(basename "$file")"
    awk -v FILE="$rel" '
        function push_heading(level, text) {
            while (depth > 0 && levels[depth] >= level) { depth-- }
            depth++
            levels[depth] = level
            heads[depth] = text
        }
        function heading_path(    s, i) {
            s = ""
            for (i = 1; i <= depth; i++) {
                if (s != "") s = s " > "
                s = s heads[i]
            }
            return s
        }
        function in_feature_section(    i, t) {
            for (i = 1; i <= depth; i++) {
                t = tolower(heads[i])
                if (t ~ /core features/) return 1
                if (t ~ /features/) return 1
                if (t ~ /capabilities/) return 1
                if (t ~ /modules/) return 1
                if (t ~ /functionality/) return 1
                if (t ~ /subsystems/) return 1
                if (t ~ /components/) return 1
            }
            return 0
        }
        /^#+[[:space:]]/ {
            match($0, /^#+/)
            lvl = RLENGTH
            text = substr($0, lvl + 1)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", text)
            push_heading(lvl, text)
            next
        }
        in_feature_section() && /^[[:space:]]*-[[:space:]]+\*\*/ {
            line = $0
            if (match(line, /\*\*[^*]+\*\*/)) {
                bold = substr(line, RSTART + 2, RLENGTH - 4)
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", bold)
                if (bold != "") {
                    printf "%s\t%s\t%s\n", bold, FILE, heading_path()
                }
            }
        }
    ' "$file"
}

extract_design_features "$PRODUCT_MD" >>"$DESIGN_TSV"
extract_design_features "$STRUCTURE_MD" >>"$DESIGN_TSV"
extract_design_features "$TECH_MD" >>"$DESIGN_TSV"

# --- step 2: header-driven capability-map parse ----------------------------
#
# Parse the header row (the line above the first |---| separator), resolve
# column indices by NAME, emit data rows as TSV. Skip rows whose status is
# superseded/archived/rejected.
#
# Output TSV: spec-id \t title \t implementation-path

parse_cap_map() {
    awk '
        function norm(s,    t) {
            t = tolower(s)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", t)
            gsub(/[-_]/, " ", t)
            gsub(/  +/, " ", t)
            return t
        }
        function is_name_col(n) {
            return (n == "capability" || n == "name" || n == "feature" || n == "title")
        }
        function is_specid_col(n) {
            return (n == "owning spec" || n == "spec id" || n == "specid" || n == "spec")
        }
        function is_status_col(n) {
            return (n == "status")
        }
        function is_path_col(n) {
            return (n == "implementation path" || n == "path" || n == "module path")
        }
        BEGIN { hdr_seen = 0; sep_seen = 0; name_idx = -1; specid_idx = -1; status_idx = -1; path_idx = -1 }
        # detect the separator row |---|---|
        /^\|[[:space:]]*:?-+/ && !sep_seen {
            sep_seen = 1
            # parse the saved header (the prior non-empty table line)
            if (hdr_row != "" && !hdr_seen) {
                row = hdr_row
                gsub(/^\|/, "", row); gsub(/\|$/, "", row)
                n = split(row, cells, /\|/)
                for (i = 1; i <= n; i++) {
                    cn = norm(cells[i])
                    if (name_idx < 0 && is_name_col(cn)) name_idx = i
                    if (specid_idx < 0 && is_specid_col(cn)) specid_idx = i
                    if (status_idx < 0 && is_status_col(cn)) status_idx = i
                    if (path_idx < 0 && is_path_col(cn)) path_idx = i
                }
                hdr_seen = 1
            }
            next
        }
        # before separator: save as candidate header
        !sep_seen && /^\|/ {
            hdr_row = $0
            next
        }
        # data rows after separator + header
        hdr_seen && /^\|/ {
            if (name_idx < 0 || specid_idx < 0) next
            row = $0
            gsub(/^\|/, "", row); gsub(/\|$/, "", row)
            nf = split(row, d, /\|/)
            for (i = 1; i <= nf; i++) { gsub(/^[[:space:]]+|[[:space:]]+$/, "", d[i]) }
            title = (name_idx <= nf) ? d[name_idx] : ""
            sid = (specid_idx <= nf) ? d[specid_idx] : ""
            status = (status_idx > 0 && status_idx <= nf) ? d[status_idx] : ""
            path = (path_idx > 0 && path_idx <= nf) ? d[path_idx] : ""
            if (status == "superseded" || status == "archived" || status == "rejected") next
            if (sid == "" && title == "") next
            printf "%s\t%s\t%s\n", sid, title, path
        }
    ' "$CAP_MAP"
}

parse_cap_map >>"$CAP_TSV" 2>/dev/null

if [ ! -s "$CAP_TSV" ]; then
    # Header parse may have failed (missing required column). Warn but continue
    # (every design feature becomes Missing, zero Orphans).
    warn "audit: capability-map parse yielded zero rows (missing required column or empty map)"
fi

# --- step 3: load override file (minimal YAML reader) ---------------------
#
# match:    list of { design_name: "...", spec_id: "..." }
# ignore:   list of literal strings (design names OR spec-ids)
#
# Single awk pass emits to two output files (override-match TSV +
# override-ignore list) using awk's `print > path` redirection. The schema is
# small enough that this pure-awk extractor suffices (no jq/yq/python).

if [ -f "$OVERRIDE_FILE" ]; then
    awk -v MATCH_OUT="$OV_MATCH_TSV" -v IGNORE_OUT="$OV_IGNORE" '
        BEGIN { in_match = 0; in_ignore = 0; cur_dn = "" }
        /^[[:space:]]*#/ { next }
        /^[[:space:]]*match:/ { in_match = 1; in_ignore = 0; cur_dn = ""; next }
        /^[[:space:]]*ignore:/ { in_match = 0; in_ignore = 1; next }
        in_match && /^[[:space:]]*-[[:space:]]+design_name:/ {
            line = $0
            sub(/.*design_name:[[:space:]]*/, "", line)
            gsub(/^["'\'']|["'\'']$/, "", line)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
            cur_dn = line
            next
        }
        in_match && cur_dn != "" && /spec_id:/ {
            line = $0
            sub(/.*spec_id:[[:space:]]*/, "", line)
            gsub(/^["'\'']|["'\'']$/, "", line)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
            print cur_dn "\t" line > MATCH_OUT
            cur_dn = ""
            next
        }
        in_match && /^[[:space:]]*-[[:space:]]*\{/ {
            line = $0
            dn = ""; sp = ""
            if (match(line, /design_name:[[:space:]]*[^,}]+/)) {
                dn = substr(line, RSTART, RLENGTH)
                sub(/design_name:[[:space:]]*/, "", dn)
                gsub(/^["'\'']|["'\'']$/, "", dn)
            }
            if (match(line, /spec_id:[[:space:]]*[^,}]+/)) {
                sp = substr(line, RSTART, RLENGTH)
                sub(/spec_id:[[:space:]]*/, "", sp)
                gsub(/^["'\'']|["'\'']$/, "", sp)
            }
            if (dn != "" && sp != "") print dn "\t" sp > MATCH_OUT
            next
        }
        in_ignore && /^[[:space:]]*-[[:space:]]+/ {
            line = $0
            sub(/^[[:space:]]*-[[:space:]]+/, "", line)
            gsub(/^["'\'']|["'\'']$/, "", line)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
            if (line != "") print line > IGNORE_OUT
            next
        }
    ' "$OVERRIDE_FILE" 2>/dev/null
fi

# --- step 4: compute the diff ---------------------------------------------
#
# Bash 3.2 (macOS default /usr/bin/bash) does not support `declare -A`, so the
# entire diff is computed in a single awk invocation — awk has native
# associative arrays. The awk reads design.tsv + cap.tsv + override TSVs and
# writes matched.tsv + missing.tsv + orphan.tsv. All override matches are
# applied BEFORE the heuristic (override-before-heuristic contract), and the ignore set excludes
# entries from BOTH the missing and orphan candidate lists.

awk \
    -v DESIGN_PATH="$DESIGN_TSV" \
    -v CAP_PATH="$CAP_TSV" \
    -v OVMATCH_PATH="$OV_MATCH_TSV" \
    -v OVIGNORE_PATH="$OV_IGNORE" \
    -v MATCHED_OUT="$MATCHED_TSV" \
    -v MISSING_OUT="$MISSING_TSV" \
    -v ORPHAN_OUT="$ORPHAN_TSV" '
    function normalize(s,    t) {
        t = tolower(s)
        gsub(/[^a-z0-9]/, "-", t)
        gsub(/--+/, "-", t)
        sub(/^-+/, "", t)
        sub(/-+$/, "", t)
        return t
    }
    function heuristic_match(dname, ctitle, cpath, dnorm,    cnorm, shorter, longer, last_seg, lastnorm) {
        cnorm = normalize(ctitle)
        if (dnorm != "" && dnorm == cnorm) return "exact"
        if (length(dnorm) <= length(cnorm)) { shorter = dnorm; longer = cnorm }
        else { shorter = cnorm; longer = dnorm }
        if (length(shorter) >= 4 && shorter != "" && index(longer, shorter) > 0) return "substring"
        last_seg = cpath
        sub(/^.*\//, "", last_seg)
        if (length(last_seg) >= 4) {
            lastnorm = normalize(last_seg)
            if (length(lastnorm) >= 4 && index(dnorm, lastnorm) > 0) return "module-token"
        }
        return ""
    }
    BEGIN {
        # Load override ignore entries.
        while ((getline line < OVIGNORE_PATH) > 0) {
            if (line != "") ignore[line] = 1
        }
        close(OVIGNORE_PATH)
        # Load override matches (design_name \t spec_id pairs).
        nmatch = 0
        while ((getline line < OVMATCH_PATH) > 0) {
            if (line == "") continue
            split(line, mp, "\t")
            if (mp[1] == "" || mp[2] == "") continue
            nmatch++
            om_dn[nmatch] = mp[1]
            om_sid[nmatch] = mp[2]
        }
        close(OVMATCH_PATH)
        # Load capability-map rows.
        ncaps = 0
        while ((getline line < CAP_PATH) > 0) {
            if (line == "") continue
            split(line, c, "\t")
            ncaps++
            cap_sid[ncaps] = c[1]
            cap_title[ncaps] = c[2]
            cap_path_arr[ncaps] = c[3]
        }
        close(CAP_PATH)
        # Apply override matches first (override-before-heuristic contract). Skip if either side is
        # ignored.
        for (i = 1; i <= nmatch; i++) {
            dn = om_dn[i]; sid = om_sid[i]
            if (dn in ignore) continue
            if (sid in ignore) continue
            if (dn in design_matched) continue
            if (sid in cap_matched) continue
            print dn "\t" sid "\toverride" > MATCHED_OUT
            design_matched[dn] = 1
            cap_matched[sid] = 1
        }
        # Process design features via heuristic.
        while ((getline line < DESIGN_PATH) > 0) {
            if (line == "") continue
            split(line, d, "\t")
            dname = d[1]; dfile = d[2]; dhp = d[3]
            if (dname in design_matched) continue
            if (dname in ignore) continue
            dnorm = normalize(dname)
            basis = ""; matched_sid = ""
            for (i = 1; i <= ncaps; i++) {
                sid = cap_sid[i]
                if (sid in cap_matched) continue
                if (sid in ignore) continue
                b = heuristic_match(dname, cap_title[i], cap_path_arr[i], dnorm)
                if (b != "") { basis = b; matched_sid = sid; break }
            }
            if (basis != "") {
                print dname "\t" matched_sid "\t" basis > MATCHED_OUT
                design_matched[dname] = 1
                cap_matched[matched_sid] = 1
            } else {
                print dname "\t" dfile "\t" dhp > MISSING_OUT
            }
        }
        close(DESIGN_PATH)
        # Orphans: cap rows not matched and not ignored.
        for (i = 1; i <= ncaps; i++) {
            sid = cap_sid[i]
            if (sid in cap_matched) continue
            if (sid in ignore) continue
            print sid "\t" cap_title[i] "\t" cap_path_arr[i] > ORPHAN_OUT
        }
        close(MATCHED_OUT)
        close(MISSING_OUT)
        close(ORPHAN_OUT)
    }
' 2>/dev/null

# --- step 5: emit report ---------------------------------------------------

# Sort each TSV deterministically (by first column).
sort -t$'\t' -k1,1 "$MATCHED_TSV" -o "$MATCHED_TSV" 2>/dev/null || true
sort -t$'\t' -k1,1 "$MISSING_TSV" -o "$MISSING_TSV" 2>/dev/null || true
sort -t$'\t' -k1,1 "$ORPHAN_TSV" -o "$ORPHAN_TSV" 2>/dev/null || true

override_present="null"
[ -f "$OVERRIDE_FILE" ] && override_present="\"audit-known-matches.yaml\""

# --- audit-report.md ---
{
    printf '# Navigator Audit Report\n\n'
    printf '> Audit at %s (commit %s).\n' "$NOW_ISO" "${HEAD_SHA:-(no-git)}"
    printf '> Inputs: design-docs (product/structure/tech.md) + capability-map.md'
    [ -f "$OVERRIDE_FILE" ] && printf ' + audit-known-matches.yaml'
    printf '.\n\n'
    printf '## Missing SPECs\n\n'
    if [ -s "$MISSING_TSV" ]; then
        printf '| design name | source file | heading path |\n'
        printf '|-------------|-------------|--------------|\n'
        while IFS=$'\t' read -r dname dfile dpath; do
            [ -z "$dname" ] && continue
            printf '| %s | %s | %s |\n' "$dname" "$dfile" "$dpath"
        done <"$MISSING_TSV"
    else
        printf '_(none — every design-named feature has a capability-map match)_\n'
    fi
    printf '\n## Orphan SPECs\n\n'
    if [ -s "$ORPHAN_TSV" ]; then
        printf '| spec-id | title | implementation-path |\n'
        printf '|---------|-------|---------------------|\n'
        while IFS=$'\t' read -r sid ctitle cpath; do
            [ -z "$sid" ] && continue
            printf '| %s | %s | %s |\n' "$sid" "$ctitle" "$cpath"
        done <"$ORPHAN_TSV"
    else
        printf '_(none — every capability-map row is anchored in a design doc)_\n'
    fi
    printf '\n## Matched\n\n'
    if [ -s "$MATCHED_TSV" ]; then
        printf '| design name | spec-id | match basis |\n'
        printf '|-------------|---------|-------------|\n'
        while IFS=$'\t' read -r dn sid basis; do
            [ -z "$dn" ] && continue
            printf '| %s | %s | %s |\n' "$dn" "$sid" "$basis"
        done <"$MATCHED_TSV"
    else
        printf '_(none — no design feature matched a capability-map row)_\n'
    fi
} | atomic_write "$AUDIT_MD"

# --- audit-report.json ---
{
    printf '{"audit_at":"%s","audit_commit":"%s","inputs":{"design_docs":["product.md","structure.md","tech.md"],"capability_map":"capability-map.md","override_file":%s},' \
        "$NOW_ISO" "${HEAD_SHA:-(no-git)}" "$override_present"
    # missing[]
    printf '"missing":['
    first=1
    while IFS=$'\t' read -r dname dfile dpath; do
        [ -z "$dname" ] && continue
        [ "$first" -eq 1 ] || printf ','
        first=0
        printf '{"design_name":"%s","source":{"file":"%s","heading_path":"%s"},"closest_match":null}' \
            "$(json_escape "$dname")" "$(json_escape "$dfile")" "$(json_escape "$dpath")"
    done <"$MISSING_TSV"
    printf '],'
    # orphan[]
    printf '"orphan":['
    first=1
    while IFS=$'\t' read -r sid ctitle cpath; do
        [ -z "$sid" ] && continue
        [ "$first" -eq 1 ] || printf ','
        first=0
        printf '{"spec_id":"%s","title":"%s","implementation_path":"%s"}' \
            "$(json_escape "$sid")" "$(json_escape "$ctitle")" "$(json_escape "$cpath")"
    done <"$ORPHAN_TSV"
    printf '],'
    # matched[]
    printf '"matched":['
    first=1
    while IFS=$'\t' read -r dn sid basis; do
        [ -z "$dn" ] && continue
        [ "$first" -eq 1 ] || printf ','
        first=0
        printf '{"design_name":"%s","spec_id":"%s","match_basis":"%s"}' \
            "$(json_escape "$dn")" "$(json_escape "$sid")" "$(json_escape "$basis")"
    done <"$MATCHED_TSV"
    printf ']}'
    printf '\n'
} | atomic_write "$AUDIT_JSON"

exit 0
