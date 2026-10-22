#!/bin/bash

# MoAI Hook Wrapper - handle-codex-review-entry.sh
# Forwards stdin JSON to `moai hook codex-review-entry` — the delayed review
# block's NEXT-TURN-ENTRY enforcement handler. The Stop gate (handle-codex-review-gate.sh) allows the turn
# and kicks a background review; this wrapper is the entry-side half: it reads
# the receipt that review recorded and delivers a fresh FAIL's detail as
# additional context — it never blocks a prompt.
# Composes with handle-user-prompt-submit.sh — it is a SEPARATE
# entry reading the same stdin JSON independently. It never runs a review
# synchronously; on a cache-MISS state the gate-ON shell fast path below fires
# the same detached, marker-guarded background kick the Stop gate uses, then
# allows.
#
# The handler is opt-in (workflow.codex.review_gate.enabled — the SAME flag
# the Stop gate reads) and fail-open on every read path.
#
# Capture stdin once (hooks may have multiple composited readers).
INPUT=$(cat)

# --- shell-layer self-gate (the handle-codex-review-gate.sh precondition
# pattern) --- The gate ships OFF, so being registered in the UserPromptSubmit
# array must NOT add a moai cold start to every prompt for every user: a hook
# that ships OFF should cost nothing per turn. Read the opt-in here, in pure
# shell, and exit 0 before any binary resolution unless it is explicitly true.
#
# The parse is deliberately conservative: awk walks the nested
# workflow: -> codex: -> review_gate: -> enabled: indentation and accepts only
# a literal true (the values the Go yaml decoder accepts for a bool). Anything
# it cannot read — an unusual layout, a flow mapping, an absent awk — counts
# as OFF. The failure direction is the point: being fooled into OFF costs one
# maintainer an opt-in to debug, while being fooled into ON costs every user a
# cold start and a possible prompt block. Same fail-CLOSED direction as the
# Go reader. No jq/yq dependency (per hook independence: a precondition MUST
# NOT introduce a new shared failure mode).
CODEX_GATE_PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
CODEX_GATE_WORKFLOW_YAML="$CODEX_GATE_PROJECT_ROOT/.moai/config/sections/workflow.yaml"

# --- gate-ON shell fast-path helpers ---
# Defined BEFORE the early exits below so a test can `source` this file with
# the gate OFF and call them directly (the D4 key-parity gate).
#
# The fast path's dependency set is exactly git + one sha256 tool (shasum,
# openssl fallback) + awk + date for Unix seconds, plus the coreutils the kick arm
# shares with the Stop gate (mkdir/rm/cat). Every absent or failing probe
# FALLS THROUGH to the binary below — the shell may skip the
# binary only on positive evidence of an ALLOW state.

# The runtime-config exclusion lists, mirrored from internal/cli
# codex_review_gate.go (reviewGateRuntimePrefixes / reviewGateTreeConfigExactPaths /
# reviewGateTreeConfigDirPrefixes) and pinned by
# TestCodexEntryFastpath_GoldenPrefixes — the config-only probe can only
# narrow the nothing-reviewable detector's answer, never widen it.
CODEX_ENTRY_RUNTIME_PREFIXES=".moai/state/ .moai/cache/ .moai/reports/ .moai/logs/ .moai/harness/ .claude/agent-memory/"
CODEX_ENTRY_CONFIG_EXACT=".claude/settings.json"
CODEX_ENTRY_CONFIG_DIR=".moai/config/"

# codex_entry_key emits the exact byte stream internal/verify Key hashes
# (tree scope): `git status --porcelain=v2` output, one NUL, `git diff HEAD`
# output, then one leg per non-ignored untracked file — rel path, NUL, type
# marker ('f' + content | 'd' | 'l' + link target), NUL — the same chaining
# key.go:40-110 uses. Byte parity with verify.Key is pinned by
# TestCodexEntryFastpath_KeyParity. Fails (non-zero) on any probe error; the
# caller falls through.
codex_entry_key() {
	local root="$1"
	local porcelain diffh rc
	porcelain=$(git -C "$root" status --porcelain=v2 2>/dev/null) || return 1
	diffh=$(git -C "$root" diff HEAD 2>/dev/null) || return 1
	# $( ) strips the trailing newline git terminates its output with; emit it
	# back exactly once when the output is non-empty (byte parity with the Go
	# hasher, which sees the raw stream).
	[ -n "$porcelain" ] && printf '%s\n' "$porcelain"
	printf '\000'
	[ -n "$diffh" ] && printf '%s\n' "$diffh"
	git -C "$root" ls-files --others --exclude-standard -z 2>/dev/null |
		while IFS= read -r -d '' name || [ -n "$name" ]; do
			[ -n "$name" ] || continue
			local path="$root/$name"
			if [ -L "$path" ]; then
				# (R3) the link target's raw bytes are not portably readable:
				# BSD readlink prints the target verbatim (no terminator),
				# GNU readlink appends a newline, and $( ) would strip
				# trailing newlines from a target that genuinely ends in one
				# (os.Readlink keeps them). No shell form reproduces the
				# os.Readlink leg byte-exactly on every platform — decline;
				# Go owns the leg. Untracked non-ignored symlinks are rare
				# (node_modules-class trees are gitignored), so the cost is
				# the cold start on those trees only.
				exit 1
			elif [ -d "$path" ]; then
				# (R2) verify.Key WalkDirs a directory entry: the 'd' marker
				# PLUS one leg per descendant (git lists an embedded repo as
				# the directory itself). The walk's lexical order is not
				# reproducible byte-exactly here, and a near-miss key silently
				# kills the kick arm (plan §G) — decline; Go owns the walk.
				exit 1
			elif [ -f "$path" ]; then
				# key.go writes path, NUL, 'f', content, NUL — the 'f' marker
				# is followed DIRECTLY by the content (no separator), so the
				# closing NUL is emitted only after cat.
				printf '%s\000f' "$name" || exit 1
				cat "$path" 2>/dev/null || exit 1
				printf '\000' || exit 1
			else
				exit 1 # unsupported file type — Go Key errors; the caller falls through
			fi
		done
	rc=$?
	[ "$rc" -eq 0 ] || return 1
	return 0
}

# codex_entry_sha_extract normalizes `shasum -a 256` and `openssl dgst
# -sha256` output to the bare hex digest: both print the 64-hex digest as a
# field; take the first 64-hex-char field on the line.
codex_entry_sha_extract() {
	awk '{ for (i = 1; i <= NF; i++) if (length($i) == 64 && $i ~ /^[0-9a-f]+$/) { print $i; exit } }'
}

# codex_entry_config_digest emits the byte stream verify.ConfigDigest hashes
# for the tree-scope review binding: the names sorted (gate, method, target),
# each length-prefixed the way writeLenPrefixed does — one little-endian
# length byte, seven zero bytes, then the bytes (internal/verify/receipt.go).
# The literals are pinned to the Go output by
# TestCodexEntryFastpath_GoldenConfigDigest; any drift breaks that equality.
codex_entry_config_digest() {
	awk '
function lp(s) {
	printf "%c", length(s)
	printf "%c%c%c%c%c%c%c", 0, 0, 0, 0, 0, 0, 0
	printf "%s", s
}
BEGIN {
	lp("gate");   lp("codex-review")
	lp("method"); lp("review/start")
	lp("target"); lp("uncommittedChanges")
}'
}

# codex_entry_resolve_path makes a possibly-relative git output absolute and
# symlink-real (the reviewScopeGitPath mirror). Fails when the path cannot be
# resolved at all.
codex_entry_resolve_path() {
	local p="$2"
	case "$p" in
		/*) : ;;
		*) p="$1/$p" ;;
	esac
	local real
	if real=$(cd "$p" 2>/dev/null && pwd -P); then
		printf '%s\n' "$real"
		return 0
	fi
	case "$p" in
		/*) printf '%s\n' "$p" ;; # unresolvable absolute path: clean it lexically (EvalSymlinks fallback)
		*) return 1 ;;
	esac
}

# Read Unix seconds portably; awk's RNG seed is not a wall clock. Reject
# uncertain or noncanonical values before receipt arithmetic or marker naming.
codex_entry_epoch() {
	local epoch
	epoch=$(date +%s 2>/dev/null) || return 1
	case "$epoch" in ''|*[!0-9]*|0[0-9]*) return 1 ;; esac
	[ "$epoch" -ge 0 ] 2>/dev/null || return 1
	printf '%s\n' "$epoch"
}

# Test seam: when sourced with CODEX_ENTRY_SOURCE_ONLY=1, yield to the caller
# after the helper definitions (the D4 key-parity test calls codex_entry_key
# directly). Unset — the normal hook path — this changes nothing: `return`
# fails outside a sourced script and the regular early exit applies.
if [ "${CODEX_ENTRY_SOURCE_ONLY:-}" = "1" ]; then
	return 0 2>/dev/null || exit 0
fi

[ -f "$CODEX_GATE_WORKFLOW_YAML" ] || exit 0
command -v awk >/dev/null 2>&1 || exit 0

CODEX_GATE_ENABLED=$(awk -v gate="codex" '
{
    line = $0
    sub(/#.*/, "", line)
    if (line ~ /^[[:space:]]*$/) next
    match(line, /^ */); ind = RLENGTH
    key = substr(line, ind + 1)
    sub(/[[:space:]]+$/, "", key)

    # Leaving a block: a line at or left of its indent closes it.
    if (in_r && ind <= r_ind) in_r = 0
    if (in_g && ind <= g_ind) { in_g = 0; in_r = 0 }
    if (in_w && ind <= w_ind) { in_w = 0; in_g = 0; in_r = 0 }

    if (!in_w) { if (ind == 0 && key == "workflow:") { in_w = 1; w_ind = ind } next }
    if (!in_g) { if (key == gate ":") { in_g = 1; g_ind = ind } next }
    if (!in_r) { if (key == "review_gate:") { in_r = 1; r_ind = ind } next }
    if (key ~ /^enabled:[[:space:]]*(true|True|TRUE)$/) { print "true"; exit }
}
' "$CODEX_GATE_WORKFLOW_YAML" 2>/dev/null)

# Gate off (or unreadable) ⇒ ALLOW with zero moai cold-start.
[ "$CODEX_GATE_ENABLED" = "true" ] || exit 0

# --- gate-ON shell fast path ---
# Resolves the common ALLOW states without the synchronous moai cold start:
# (1) eligibility — a WT- branch (card scope) and either entry-side skip
#     policy (tree_scope: skip; the primary-checkout primary_scope: skip
#     default) fall through, mirroring "the two halves skip together"
#     (codex_review_delay.go:217, codex_review_tree_scope.go);
# (2) head + tree key in shell (D4 byte parity with verify.Key);
# (3) the receipt probe over .moai/state/verify/snapshots/<head12>-<digest>.json
#     — an unreadable or unparseable snapshot is receipt-ABSENT (the shared
#     verify.LoadReceipt reading, receipt.go:113-117) and takes (4)'s miss arm;
# (4) cache-MISS states fire the marker-guarded background kick, gated on the
#     two kick-region preconditions (codex_review_delay.go:214-225):
#     (4a) nothing-reviewable — change presence AND the config-only narrowing
#          (clean tree or config-only-only tree: exit 0 WITHOUT the kick);
#     (4b) binary-age — HEAD-match pre-check FIRST (StatusFresh,
#          binlag.go:139-141), then merge-base --is-ancestor: a decidable
#          behind verdict exits 0 WITHOUT the kick; undecidable age proceeds
#          per binlag's leniency (codex_review_binary_age.go:33-37);
#     (4c) both satisfied — exclusive-create the .kick-<epoch/900> marker
#          (the shared Stop-arm marker), kick detached, exit 0;
# (5) fresh pass/inconclusive → exit 0 WITHOUT kicking (codex_review_gate.go:234-244);
# (6) fresh FAIL → fall through (the Go handler owns fresh-failure delivery);
# (7) every failure → fall through.
codex_entry_fastpath() {
	# --- (1) eligibility ---
	# Config and session roots have different payload precedence. Both must
	# agree with this wrapper tree before it can decide without the handler.
	if ! printf '%s' "$INPUT" | awk -v root="$CODEX_GATE_PROJECT_ROOT" '
{
	if (match($0, /"cwd"[[:space:]]*:[[:space:]]*"/)) {
		rest = substr($0, RSTART + RLENGTH)
		end = index(rest, "\"")
		v = (end > 1 ? substr(rest, 1, end - 1) : "")
		if (v != "" && v != root) mismatch = 1
	}
	if (match($0, /"project_dir"[[:space:]]*:[[:space:]]*"/)) {
		rest = substr($0, RSTART + RLENGTH)
		end = index(rest, "\"")
		v = (end > 1 ? substr(rest, 1, end - 1) : "")
		if (v != "" && v != root) mismatch = 1
	}
}
END { exit mismatch ? 1 : 0 }
'; then
		return 1
	fi

	command -v git >/dev/null 2>&1 || return 1

	local branch
	branch=$(git -C "$CODEX_GATE_PROJECT_ROOT" branch --show-current 2>/dev/null) || return 1
	case "$branch" in
		WT-*) return 1 ;; # card scope — the Go handler owns card sessions
	esac

	# Read only literal skip policies here. Aliases, anchors, merges, and
	# other unsupported values belong to the authoritative YAML reader.
	local tree_scope primary_scope
	tree_scope=$(awk -v gate="codex" -v want="tree_scope" '
{
    line = $0
    sub(/#.*/, "", line)
    if (line ~ /^[[:space:]]*$/) next
    match(line, /^ */); ind = RLENGTH
    key = substr(line, ind + 1)
    sub(/[[:space:]]+$/, "", key)
    if (in_r && ind <= r_ind) in_r = 0
    if (in_g && ind <= g_ind) { in_g = 0; in_r = 0 }
    if (in_w && ind <= w_ind) { in_w = 0; in_g = 0; in_r = 0 }
    if (!in_w) { if (ind == 0 && key == "workflow:") { in_w = 1; w_ind = ind } next }
    if (!in_g) { if (key == gate ":") { in_g = 1; g_ind = ind } next }
    if (!in_r) { if (key == "review_gate:") { in_r = 1; r_ind = ind } next }
    if (key ~ /^<<[[:space:]]*:/) { print "unresolved"; exit }
    if (key ~ "^" want ":[[:space:]]") {
        v = key
        sub("^" want ":[[:space:]]*", "", v)
        q = sprintf("%c", 39)
        if (substr(v, 1, 1) == "\"" || substr(v, 1, 1) == q) v = substr(v, 2)
        if (substr(v, length(v), 1) == "\"" || substr(v, length(v), 1) == q) v = substr(v, 1, length(v) - 1)
        # (F8) normalize like the Go readers: EqualFold(TrimSpace(v)) —
        # SKIP, " skip " and friends must compare equal to the lowercase
        # literal, or a policy value is mis-read and the shell kicks where
        # the Go design allows (tree_scope) or stays suppressed
        # (primary_scope).
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", v)
        v = tolower(v)
        print v
        exit
    }
}
' "$CODEX_GATE_WORKFLOW_YAML" 2>/dev/null)
	case "$tree_scope" in
		""|review) : ;;
		*) return 1 ;; # literal skip or a value requiring full YAML decoding
	esac

	primary_scope=$(awk -v gate="codex" -v want="primary_scope" '
{
    line = $0
    sub(/#.*/, "", line)
    if (line ~ /^[[:space:]]*$/) next
    match(line, /^ */); ind = RLENGTH
    key = substr(line, ind + 1)
    sub(/[[:space:]]+$/, "", key)
    if (in_r && ind <= r_ind) in_r = 0
    if (in_g && ind <= g_ind) { in_g = 0; in_r = 0 }
    if (in_w && ind <= w_ind) { in_w = 0; in_g = 0; in_r = 0 }
    if (!in_w) { if (ind == 0 && key == "workflow:") { in_w = 1; w_ind = ind } next }
    if (!in_g) { if (key == gate ":") { in_g = 1; g_ind = ind } next }
    if (!in_r) { if (key == "review_gate:") { in_r = 1; r_ind = ind } next }
    if (key ~ "^" want ":[[:space:]]") {
        v = key
        sub("^" want ":[[:space:]]*", "", v)
        q = sprintf("%c", 39)
        if (substr(v, 1, 1) == "\"" || substr(v, 1, 1) == q) v = substr(v, 2)
        if (substr(v, length(v), 1) == "\"" || substr(v, length(v), 1) == q) v = substr(v, 1, length(v) - 1)
        # (F8) normalize like the Go readers: EqualFold(TrimSpace(v)) —
        # SKIP, " skip " and friends must compare equal to the lowercase
        # literal, or a policy value is mis-read and the shell kicks where
        # the Go design allows (tree_scope) or stays suppressed
        # (primary_scope).
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", v)
        v = tolower(v)
        print v
        exit
    }
}
' "$CODEX_GATE_WORKFLOW_YAML" 2>/dev/null)

	# Primary-checkout detection, the Go comparison (isPrimaryCheckoutGit):
	# --git-dir == --git-common-dir, both resolved against the session tree.
	# Any ambiguity falls through.
	local gd gcd gd_abs gcd_abs
	gd=$(git -C "$CODEX_GATE_PROJECT_ROOT" rev-parse --git-dir 2>/dev/null) || return 1
	gcd=$(git -C "$CODEX_GATE_PROJECT_ROOT" rev-parse --git-common-dir 2>/dev/null) || return 1
	gd_abs=$(codex_entry_resolve_path "$CODEX_GATE_PROJECT_ROOT" "$gd") || return 1
	gcd_abs=$(codex_entry_resolve_path "$CODEX_GATE_PROJECT_ROOT" "$gcd") || return 1
	if [ "$gd_abs" = "$gcd_abs" ]; then
		[ "$primary_scope" = "review" ] || return 1
	fi

	# --- (2) head + tree key ---
	local head root
	head=$(git -C "$CODEX_GATE_PROJECT_ROOT" rev-parse HEAD 2>/dev/null) || return 1
	[ -n "$head" ] || return 1
	# The receipt, markers, and kick all anchor to the GIT ROOT — the Go
	# entry anchors the scope's dir with reviewExclusionRoot (--show-toplevel).
	root=$(git -C "$CODEX_GATE_PROJECT_ROOT" rev-parse --show-toplevel 2>/dev/null) || return 1
	[ -n "$root" ] || return 1

	local sha_bin
	if command -v shasum >/dev/null 2>&1; then
		sha_bin="shasum -a 256"
	elif command -v openssl >/dev/null 2>&1; then
		sha_bin="openssl dgst -sha256"
	else
		return 1
	fi

	local hex key_rc
	set -o pipefail
	hex=$(codex_entry_key "$root" | $sha_bin 2>/dev/null | codex_entry_sha_extract)
	key_rc=$?
	set +o pipefail
	[ "$key_rc" -eq 0 ] || return 1
	[ "${#hex}" -eq 64 ] || return 1
	case "$hex" in *[!0-9a-f]*) return 1 ;; esac
	local digest="${hex:0:16}"
	local snapshot="$root/.moai/state/verify/snapshots/${head:0:12}-$digest.json"

	# --- (3) the receipt probe ---
	local now receipt state verdict
	now=$(codex_entry_epoch) || return 1
	receipt=$(awk -v cmd="moai verify codex-review" -v now="$now" -v ttl=86400 -v wantkey="$head:$digest" '
function dc(y, m, d,    yy, era, yoe, doy, doe) {
	yy = y
	if (m <= 2) yy -= 1
	era = int(yy / 400)
	yoe = yy - era * 400
	doy = int((153 * (m + (m > 2 ? -3 : 9)) + 2) / 5) + d - 1
	doe = yoe * 365 + int(yoe / 4) - int(yoe / 100) + doy
	return era * 146097 + doe - 719468
}
function rfc(ts,    y, mo, d, h, mi, s, rest, off, oh, om, md) {
	# A valid prefix is insufficient: reject unsupported date/zone syntax
	# and calendar overflow before interpreting a receipt as evidence.
	if (ts !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]/) {
		invalid_time = 1
		return 0
	}
	y = substr(ts, 1, 4) + 0
	mo = substr(ts, 6, 2) + 0
	d = substr(ts, 9, 2) + 0
	h = substr(ts, 12, 2) + 0
	mi = substr(ts, 15, 2) + 0
	s = substr(ts, 18, 2) + 0
	md = 31
	if (mo == 2) md = 28 + (y % 4 == 0 && (y % 100 != 0 || y % 400 == 0))
	else if (mo == 4 || mo == 6 || mo == 9 || mo == 11) md = 30
	if (y < 1 || mo < 1 || mo > 12 || d < 1 || d > md || h > 23 || mi > 59 || s > 59) {
		invalid_time = 1
		return 0
	}
	rest = substr(ts, 20)
	sub(/^[.][0-9]+/, "", rest)
	if (rest !~ /^(Z|[+-][0-9][0-9]:[0-9][0-9])$/) {
		invalid_time = 1
		return 0
	}
	off = 0
	if (rest ~ /^[+-]/) {
		oh = substr(rest, 2, 2) + 0
		om = substr(rest, 5, 2) + 0
		if (oh > 23 || om > 59) {
			invalid_time = 1
			return 0
		}
		off = oh * 3600 + om * 60
		if (substr(rest, 1, 1) == "-") off = -off
	}
	return dc(y, mo, d) * 86400 + h * 3600 + mi * 60 + s - off
}
function extract(name,    v) {
	v = $0
	sub("^[[:space:]]*\"" name "\"[[:space:]]*:[[:space:]]*\"", "", v)
	sub("\"[[:space:]]*,?[[:space:]]*$", "", v)
	return v
}
# The snapshot loader decodes every typed timestamp, including the snapshot
# timestamp and checks unrelated to this command. An invalid one anywhere
# must not leave a seemingly fresh selected verdict usable here.
/^[[:space:]]*"recorded_at":[[:space:]]*"/ { rfc(extract("recorded_at")) }
# (F7) document-level validation, the verify.Load mirror: a parsable receipt
# requires the WHOLE document — the checks array closed and the document
# closed — AND the stored key equal to the key this shell computed (the
# filename-collision defense). A truncated or mismatched file reads as
# receipt-ABSENT (cache-miss), never as a fresh verdict.
/^\{$/ { doc_open = 1 }
doc_open && /^  "key":/ { stored_key = extract("key") }
doc_open && /^  "checks": \[/ { checks_open = 1 }
/^  \],?[[:space:]]*$/ { checks_closed = 1 }
/^\},?[[:space:]]*$/ { doc_closed = 1 }
/^    \{/ { ec = 0; rec = ""; verd = ""; cfg = ""; tl = "" }
ec == 0 && /^      "command":/ { if (extract("command") == cmd) ec = 1 }
ec == 1 && /^      "recorded_at":/ { rec = extract("recorded_at") }
ec == 1 && /^      "verdict":/ { verd = extract("verdict") }
ec == 1 && /^      "config_digest":/ { cfg = extract("config_digest") }
ec == 1 && /^      "tool_version":/ { tl = extract("tool_version") }
/^    \},?[[:space:]]*$/ {
	if (ec) {
		# (D1) FindCommand returns the FIRST matching entry; a document with
		# TWO review-command entries is a shape this shell must not read
		# last-wins — count the matches and require exactly one below.
		matches++
		gsub(/^[[:space:]]+|[[:space:]]+$/, "", verd)
		verd = tolower(verd)
		eps = rfc(rec)
		f_state = ((now - eps) > ttl ? "stale" : "fresh")
		f_verd = verd
		f_cfg = cfg
		f_tl = tl
		found = 1
	}
}
END {
	if (invalid_time) print "invalid-time"
	else if (found && matches == 1 && doc_closed && checks_closed && stored_key == wantkey)
		print f_state " " f_verd " " f_cfg " " f_tl
}
' "$snapshot" 2>/dev/null)
	[ "$receipt" = "invalid-time" ] && return 1
	# receipt: "fresh <verdict> <config_digest> <tool_version>" | "stale ..."
	# | "" (receipt-ABSENT — no snapshot, unreadable, or unparseable: the
	# LoadReceipt reading). The tool_version value may contain spaces; it is
	# the LAST field and is taken whole.
	state=${receipt%% *}
	if [ -n "$state" ]; then
		local rest stored_cfg stored_tool
		rest=${receipt#* }
		verdict=${rest%% *}
		rest=${rest#* }
		stored_cfg=${rest%% *}
		stored_tool=${rest#* }
		# (R6) the two remaining CheckReceipt bound fields (receipt.go:73/:75):
		# an UNBOUND stored field is never evidence — fall through and let the
		# Go handler apply its own reading; a MISMATCH means the receipt is not
		# valid for the current reviewer/config — the same cache-miss the Stop
		# side re-kicks, so the miss arm proceeds.
		if [ -z "$stored_cfg" ] || [ -z "$stored_tool" ]; then
			return 1
		fi
		# (C2) a stored value carrying JSON escapes (a backslash in the raw
		# text — a multi-line reviewer version, an HTML-escaped character)
		# cannot be decoded byte-exactly in this awk; comparing it against the
		# Go-decoded current value diverges and would mis-class a VALID fresh
		# FAIL as a cache miss (delivery skipped, re-kick every bucket, no
		# self-healing). Decline the binding — the Go handler JSON-decodes
		# before comparing and owns the state.
		case "$stored_tool$stored_cfg" in
			*\\*) return 1 ;;
		esac
		local cur_cfg cur_tool
		cur_cfg=$(codex_entry_config_digest | $sha_bin 2>/dev/null | codex_entry_sha_extract) || return 1
		[ "${#cur_cfg}" -eq 64 ] || return 1
		cur_cfg="${cur_cfg:0:16}" # ConfigDigest binds 16 hex chars (digestHexLen)
		if ! command -v codex >/dev/null 2>&1; then
			return 1 # no reviewer: the Go entry allows at its reviewer-lookup step
		fi
		cur_tool=$(codex --version 2>/dev/null | awk '
{ sub(/^[[:space:]]+|[[:space:]]+$/, "")
  if (NR == 1) out = $0
  else out = out "\n" $0 }
END { print out }') || return 1
		[ -n "$cur_tool" ] || return 1
		if [ "$stored_cfg" != "$cur_cfg" ] || [ "$stored_tool" != "$cur_tool" ]; then
			state="" # bound-field mismatch ⇒ the receipt is not valid here ⇒ miss arm
		fi
		if [ -n "$state" ] && [ "$state" = "fresh" ]; then
			case "$verdict" in
				fail*) return 1 ;; # (6) fresh FAIL: the Go handler owns the delivery
				*) exit 0 ;;       # (5) fresh pass/inconclusive: allow, never re-kicked
			esac
		fi
	fi
	# (4) cache-MISS arm: no receipt, stale receipt, bound-field mismatch, or
	# receipt-absent.

	# --- (4a) nothing-reviewable precondition ---
	local p1
	p1=$(git -C "$root" status --porcelain 2>/dev/null) || exit 0 # detector fail-open ⇒ allow
	if ! printf '%s\n' "$p1" | awk -v rt="$CODEX_ENTRY_RUNTIME_PREFIXES" '
BEGIN { n = split(rt, RT, " ") }
function managed(p,   i) { for (i = 1; i <= n; i++) if (index(p, RT[i]) == 1) return 1; return 0 }
{
	line = $0
	if (line ~ /^[[:space:]]*$/) next
	if (length(line) <= 3) next
	path = substr(line, 4)
	gsub(/^[[:space:]]+|[[:space:]]+$/, "", path)
	if ((idx = index(path, " -> ")) > 0) path = substr(path, 1, idx - 1)
	sub(/^[[:space:]]+/, "", path)
	sub(/[[:space:]]+$/, "", path)
	if (path == "" || managed(path)) next
	found = 1
	exit
}
END { exit found ? 0 : 1 }
'; then
		exit 0 # clean tree — nothing was kicked to enforce
	fi
	local pa
	pa=$(git -C "$root" status --porcelain --untracked-files=all 2>/dev/null) || pa=""
	if [ -n "$pa" ]; then
		if printf '%s\n' "$pa" | awk -v rt="$CODEX_ENTRY_RUNTIME_PREFIXES" -v cx="$CODEX_ENTRY_CONFIG_EXACT" -v cd="$CODEX_ENTRY_CONFIG_DIR" '
BEGIN { n = split(rt, RT, " "); m = split(cx, CX, " "); k = split(cd, CD, " ") }
function managed(p,   i) { for (i = 1; i <= n; i++) if (index(p, RT[i]) == 1) return 1; return 0 }
function excluded(p,   i) {
	if (managed(p)) return 1
	for (i = 1; i <= m; i++) if (p == CX[i]) return 1
	for (i = 1; i <= k; i++) if (index(p, CD[i]) == 1) return 1
	return 0
}
{
	line = $0
	if (line ~ /^[[:space:]]*$/) next
	if (length(line) <= 3) next
	path = substr(line, 4)
	gsub(/^[[:space:]]+|[[:space:]]+$/, "", path)
	ex = 0
	if ((idx = index(path, " -> ")) > 0) {
		oldp = substr(path, 1, idx - 1)
		newp = substr(path, idx + 4)
		gsub(/^[[:space:]]+|[[:space:]]+$/, "", oldp)
		gsub(/^[[:space:]]+|[[:space:]]+$/, "", newp)
		ex = excluded(oldp) && excluded(newp)
	} else {
		ex = excluded(path)
	}
	if (!ex) { bad = 1; exit }
	any = 1
}
END { exit (any && !bad) ? 0 : 1 }
'; then
			exit 0 # config-only-only tree — the Stop gate owns the decision
		fi
	fi

	# --- (4b) binary-age precondition ---
	local moai_bin=""
	if [ -n "$CLAUDE_PROJECT_DIR" ] && [ -f "$CLAUDE_PROJECT_DIR/bin/moai" ] && [ -x "$CLAUDE_PROJECT_DIR/bin/moai" ]; then
		moai_bin="$CLAUDE_PROJECT_DIR/bin/moai"
	fi
	if [ -z "$moai_bin" ]; then
		local path_bin
		path_bin=$(command -v moai 2>/dev/null) || path_bin=""
		if [ -n "$path_bin" ] && [ -f "$path_bin" ] && [ -x "$path_bin" ]; then
			moai_bin="$path_bin"
		elif [ -f "$HOME/go/bin/moai" ] && [ -x "$HOME/go/bin/moai" ]; then
			moai_bin="$HOME/go/bin/moai"
		fi
	fi
	[ -n "$moai_bin" ] || return 1 # unresolvable binary → fall through

	local ver_out bin_commit
	ver_out=$("$moai_bin" version 2>/dev/null) || return 1
	# The version stamp embeds the build commit the way git describe does
	# (<tag>-<n>-g<sha>[-dirty]).
	bin_commit=$(printf '%s\n' "$ver_out" | awk '
	{ if (match($0, /-g[0-9a-f][0-9a-f]*/)) { print substr($0, RSTART + 2, RLENGTH - 2); exit } }')
	# (R5) Without a -g component the shell cannot distinguish "tag-only
	# build, no commit metadata" from "unparseable output" — fall through and
	# let Go's staleBinarySkipApplies own the leniency decision with the real
	# ldflags commit (an ancestor binary is no longer mis-classed).
	[ -n "$bin_commit" ] || return 1
	case "$head" in
		"$bin_commit"*) : ;; # HEAD-match pre-check FIRST: StatusFresh → the kick proceeds
		*)
			if git -C "$root" merge-base --is-ancestor "$bin_commit" "$head" >/dev/null 2>&1; then
				exit 0 # decidable behind → allow without the early kick
			fi
			;; # divergent / ahead → binlag leniency → the kick proceeds
	esac

	# --- (4c) the marker-guarded kick (the shared Stop-arm marker,
	# kickInFlight / codexReviewKickMarkerPath) ---
	local epoch bucket marker_dir marker owned
	epoch=$(codex_entry_epoch) || return 1
	bucket=$((epoch / 900))
	marker_dir="$root/.moai/state/verify/codex-review"
	marker="$marker_dir/$head-$digest.kick-$bucket"
	# The state dir is ensured BEFORE the exclusive create (kickInFlight's
	# MkdirAll-before-OpenFile order); a create failure here then means an
	# existing marker or an unwritable dir, never a missing parent.
	mkdir -p "$marker_dir" 2>/dev/null
	rm -f "$marker_dir/$head-$digest.kick-$((bucket - 1))" 2>/dev/null # GC the aged-out bucket, best-effort
	owned=0
	set -C
	if printf '%s\n' "$epoch" >"$marker" 2>/dev/null; then
		owned=1
	elif [ -e "$marker" ]; then
		set +C
		exit 0 # a kick for this tree state is already in flight this bucket
	fi
	set +C
	# owned=0 with no marker on disk = unwritable state dir ⇒ fail-open toward
	# reviewing: the kick proceeds undeduplicated (kickInFlight).
	if [ "$owned" -eq 0 ] && [ ! -e "$marker" ]; then
		: # unwritable marker ⇒ fail-open toward reviewing (kickInFlight)
	fi
	# The kick: the same detached `moai verify codex-review --project-root
	# <dir>` the Stop gate launches, output appended to the same log.
	# (R4) the log file is probed openable BEFORE the spawn: a failing `>>`
	# redirect (e.g. the log path exists as a directory) kills the async
	# command before exec and strands the marker for the whole bucket with no
	# review behind it. An unopenable log routes through the /dev/null arm —
	# the Go kick's log failure is best-effort and never stops the kick.
	# (C1) non-regular log paths are rejected BEFORE the probe: opening a
	# FIFO for write blocks in the foreground until a reader appears (repro
	# exit 124) — the hook would hang to its 30s budget with the marker
	# stranded. A missing path passes (the probe creates it); directories and
	# FIFOs take the /dev/null arm.
	local log_dir="$root/.moai/logs" log_target
	log_target="$log_dir/codex-review-bg.log"
	mkdir -p "$log_dir" 2>/dev/null
	if { [ ! -e "$log_target" ] || [ -f "$log_target" ]; } && ( : >>"$log_target" ) 2>/dev/null; then
		"$moai_bin" verify codex-review --project-root "$root" >>"$log_target" 2>&1 &
	else
		"$moai_bin" verify codex-review --project-root "$root" >/dev/null 2>&1 &
	fi
	# (R4) a spawn that died before exec must not strand the marker: release
	# it so the Stop side can retry within the bucket (the Go kickErr
	# semantic). Best-effort — a live child (zombie included) reads as started.
	local kick_pid=$!
	kill -0 "$kick_pid" 2>/dev/null || rm -f "$marker" 2>/dev/null
	exit 0
}
codex_entry_fastpath || :
# Everything above either exited 0 (an ALLOW state resolved in shell) or fell
# through to here: the synchronous Go handler owns the verdict delivery.

# Resolve the moai binary (3-tier: $CLAUDE_PROJECT_DIR-relative, PATH, $HOME).
MOAI_BIN=""
if [ -n "$CLAUDE_PROJECT_DIR" ] && [ -f "$CLAUDE_PROJECT_DIR/bin/moai" ] && [ -x "$CLAUDE_PROJECT_DIR/bin/moai" ]; then
	# CLAUDE_PROJECT_DIR is the project root, not the hook script directory.
	MOAI_BIN="$CLAUDE_PROJECT_DIR/bin/moai"
fi
if [ -z "$MOAI_BIN" ]; then
	MOAI_PATH_BIN="$(command -v moai 2>/dev/null)"
	if [ -f "$MOAI_PATH_BIN" ] && [ -x "$MOAI_PATH_BIN" ]; then
		MOAI_BIN="$MOAI_PATH_BIN"
	elif [ -f "$HOME/go/bin/moai" ] && [ -x "$HOME/go/bin/moai" ]; then
		MOAI_BIN="$HOME/go/bin/moai"
	fi
fi

# Fail-open: if the moai binary is unavailable, ALLOW (never break the prompt
# pipeline). The handler itself is fail-open, so an absent binary is a no-op.
if [ -z "$MOAI_BIN" ]; then
	exit 0
fi

printf '%s' "$INPUT" | "$MOAI_BIN" hook codex-review-entry
# Always exit 0: the BLOCK decision rides the JSON Decision field on stdout
# (exit 2 would discard stdout JSON per Claude Code semantics).
exit 0
