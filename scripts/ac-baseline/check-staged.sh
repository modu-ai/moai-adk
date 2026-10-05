#!/bin/sh
# check-staged.sh — commit-time AC-snapshot guard (card t1150,
# SPEC-ACSNAPSHOT-COMMIT-GUARD-001). Local-only dev tool: no template mirror,
# never distributed.
#
# Runs as the git config-defined pre-commit hook `ac-baseline-guard` (armed by
# install-hook.sh). It rejects a commit in which an in-place amendment of a
# depth-1 .moai/specs/<dir>/acceptance.md moves that file's AC count away from
# its record in the STAGED corpus snapshot. Counter, acceptance blob, and
# snapshot are all read from the index git designates for this commit, so a
# regenerated snapshot staged in the same commit passes by construction.
#
# Trust boundary: the only file-derived text handed to a shell parser is the
# counter body extracted from the staged manager-docs.md (the same trust the
# corpus test already extends). Paths travel NUL-delimited, reach the counter
# only through AC_FILE (a temp copy), and are matched against the snapshot by
# exact string equality through ENVIRON — never eval, never a pattern.
#
# Exit status: 0 = nothing to check / pass / tool fault (NOT CHECKED, fail
# open); 1 = a measured mismatch in at least one file (it wins over faults).
#
# No cd happens before any git call: GIT_INDEX_FILE may be relative.

carrier=.claude/agents/moai/manager-docs.md
baseline=.moai/reports/t338/ac-count-baseline.txt
regen='MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1'
doc=.moai/docs/ac-count-baseline-refresh.md
tag='ac-baseline-guard:'

say() { printf '%s\n' "$*" >&2; }

# qualifies PATH — true for exactly .moai/specs/<one segment>/acceptance.md with
# a segment other than _archive. A git pathspec '*' crosses '/', so the
# depth-1 rule is enforced here rather than trusted to a pathspec.
qualifies() {
	case $1 in
	.moai/specs/*/acceptance.md) ;;
	*) return 1 ;;
	esac
	q_seg=${1#.moai/specs/}
	q_seg=${q_seg%/acceptance.md}
	case $q_seg in
	'' | */* | _archive) return 1 ;;
	esac
	return 0
}

# ---------------------------------------------------------------------------
# Enumeration pass: list the index's in-place modifications against HEAD,
# NUL-delimited, renames off, and hand them to the judging pass as argv.
# ---------------------------------------------------------------------------
if [ "${1:-}" != "--files" ]; then
	# An initial commit has no HEAD, so nothing can be modified against it.
	git rev-parse -q --verify HEAD >/dev/null 2>&1 || exit 0

	list=$(mktemp "${TMPDIR:-/tmp}/ac-baseline-guard-list.XXXXXX") || {
		say "$tag NOT CHECKED (cannot create a temp file): staged acceptance.md amendments were not compared"
		exit 0
	}
	state=$(mktemp "${TMPDIR:-/tmp}/ac-baseline-guard-state.XXXXXX") || {
		rm -f "$list"
		say "$tag NOT CHECKED (cannot create a temp file): staged acceptance.md amendments were not compared"
		exit 0
	}
	trap 'rm -f "$list" "$state"' EXIT
	trap 'exit 1' INT TERM

	# ':/' anchors the pathspec at the tree root, so a manual run from a
	# subdirectory enumerates the same paths the hook does from the top.
	if ! git diff --cached --name-only -z --no-renames --diff-filter=M HEAD -- :/.moai/specs >"$list" 2>/dev/null </dev/null; then
		say "$tag NOT CHECKED (git diff --cached failed): staged acceptance.md amendments were not compared"
		exit 0
	fi
	[ -s "$list" ] || exit 0

	AC_BASELINE_GUARD_STATE=$state xargs -0 sh "$0" --files <"$list"
	rc=$?
	if [ -s "$state" ]; then
		exit 1
	fi
	if [ "$rc" -ne 0 ]; then
		say "$tag NOT CHECKED (judging pass exited $rc): staged acceptance.md amendments were not compared"
	fi
	exit 0
fi

# ---------------------------------------------------------------------------
# Judging pass: argv holds candidate paths (verbatim, never shell-parsed).
# ---------------------------------------------------------------------------
shift
state=${AC_BASELINE_GUARD_STATE:-}

n=0
for p do
	qualifies "$p" && n=$((n + 1))
done
# No qualifying file: exit before touching the counter or the snapshot.
[ "$n" -eq 0 ] && exit 0

# fault_all REASON PATH... — every qualifying file is left unchecked; fail open.
fault_all() {
	f_reason=$1
	shift
	for f_p do
		qualifies "$f_p" && say "$tag NOT CHECKED ($f_reason): $f_p"
	done
	exit 0
}

tmpd=$(mktemp -d "${TMPDIR:-/tmp}/ac-baseline-guard.XXXXXX") || fault_all "cannot create a temp directory" "$@"
trap 'rm -rf "$tmpd"' EXIT
trap 'exit 1' INT TERM

# Counter: exactly one BEGIN/END sentinel pair, END after BEGIN, non-empty body
# (the same rule as extractCounterCommand in internal/spec/ac_count_clause_test.go).
git show ":$carrier" >"$tmpd/carrier" 2>/dev/null </dev/null ||
	fault_all "counter carrier $carrier absent from the index" "$@"
if ! reason=$(OUT="$tmpd/counter" LC_ALL=C awk '
	{
		line[NR] = $0
		s = $0
		sub(/^[ \t\r]+/, "", s)
		sub(/[ \t\r]+$/, "", s)
		if (s == "# MOAI-AC-COUNTER-BEGIN") { nb++; bl = NR }
		else if (s == "# MOAI-AC-COUNTER-END") { ne++; el = NR }
	}
	END {
		if (nb == 0 && ne == 0) { print "counter sentinel pair absent from the staged carrier"; exit 1 }
		if (nb != 1 || ne != 1) { print "expected exactly one counter sentinel pair, found BEGIN=" nb " END=" ne; exit 1 }
		if (el <= bl) { print "counter END sentinel precedes BEGIN"; exit 1 }
		out = ENVIRON["OUT"]
		body = 0
		printf "" > out
		for (i = bl + 1; i < el; i++) {
			print line[i] > out
			if (line[i] ~ /[^ \t\r]/) body = 1
		}
		if (!body) { print "counter sentinel pair delimits an empty body"; exit 1 }
	}
' "$tmpd/carrier"); then
	fault_all "$reason" "$@"
fi
counter=$(cat "$tmpd/counter")

git show ":$baseline" >"$tmpd/baseline" 2>/dev/null </dev/null ||
	fault_all "snapshot $baseline absent from the index" "$@"

checked=0
mismatch=0
reports=''
for p do
	qualifies "$p" || continue

	if ! git show ":$p" >"$tmpd/blob" 2>/dev/null </dev/null; then
		say "$tag NOT CHECKED (staged blob unreadable): $p"
		continue
	fi

	# The only shell-parsed file-derived text: the counter body. The target is
	# a temp copy named through AC_FILE, never the path itself.
	AC_FILE="$tmpd/blob" sh -c "$counter" >"$tmpd/out" 2>"$tmpd/err" </dev/null
	rc=$?
	case $rc in
	0)
		st_live=$(LC_ALL=C awk 'NF { n += NF; v = $1 } END { if (n != 1 || v !~ /^[0-9]+$/) exit 1; print v + 0 }' "$tmpd/out") || {
			say "$tag NOT CHECKED (counter stdout is not one integer): $p"
			continue
		}
		st_exc=$(LC_ALL=C sed -n 's/.*live=[0-9][0-9]* excluded=\([0-9][0-9]*\) ambiguous=0.*/\1/p' "$tmpd/err" | head -n 1)
		if [ -z "$st_exc" ]; then
			say "$tag NOT CHECKED (counter per-state tally absent from stderr): $p"
			continue
		fi
		staged="COUNT $st_live $((st_exc + 0))"
		;;
	3)
		st_ids=$(LC_ALL=C awk 'NR == 1 && $1 == "AMBIGUOUS" { for (i = 2; i <= NF; i++) print $i }' "$tmpd/out" |
			LC_ALL=C sort | LC_ALL=C awk '{ printf "%s%s", (NR > 1 ? " " : ""), $0 }')
		if [ -z "$st_ids" ]; then
			say "$tag NOT CHECKED (counter exited 3 without an AMBIGUOUS line): $p"
			continue
		fi
		staged="HALT $st_ids"
		;;
	*)
		say "$tag NOT CHECKED (counter exited $rc): $p"
		continue
		;;
	esac

	# Exact-string lookup: the path reaches awk through ENVIRON only.
	rec=$(P="$p" LC_ALL=C awk '
		function parse(   i, k, n, ex, v, s, tmp, j, ids) {
			if (NF < 2) return "MALFORMED"
			if ($2 == "COUNT") {
				if (NF < 5 || $3 !~ /^[0-9]+$/) return "MALFORMED"
				n = $3 + 0
				ex = 0
				for (i = 4; i <= NF; i++) {
					if ($i ~ /^live=/) {
						v = substr($i, 6)
						if (v !~ /^[0-9]+$/ || v + 0 != n) return "MALFORMED"
					} else if ($i ~ /^excluded=/) {
						v = substr($i, 10)
						if (v !~ /^[0-9]+$/) return "MALFORMED"
						ex = v + 0
					}
				}
				return "COUNT " n " " ex
			}
			if ($2 == "HALT") {
				k = 0
				for (i = 3; i <= NF; i++) {
					if (index($i, "=")) break
					ids[++k] = $i ""
				}
				if (k == 0 || index($0, "owner=") == 0 || index($0, "reason=") == 0) return "MALFORMED"
				for (i = 2; i <= k; i++) {
					tmp = ids[i]
					for (j = i - 1; j >= 1 && (ids[j] "") > (tmp ""); j--) ids[j + 1] = ids[j]
					ids[j + 1] = tmp
				}
				s = ids[1]
				for (i = 2; i <= k; i++) s = s " " ids[i]
				return "HALT " s
			}
			return "MALFORMED"
		}
		BEGIN { want = ENVIRON["P"] ""; st = "ABSENT" }
		{ sub(/[ \t]+$/, "") }
		/^#/ || NF == 0 { next }
		($1 "") == want { st = parse() }
		END { print st }
	' "$tmpd/baseline") || rec=''

	case $rec in
	'COUNT '* | 'HALT '* | ABSENT | MALFORMED) ;;
	*)
		# The lookup itself failed: a fault, never an implicit pass.
		say "$tag NOT CHECKED (snapshot lookup failed): $p"
		continue
		;;
	esac

	case $rec in
	MALFORMED)
		say "$tag NOT CHECKED (snapshot record malformed in $baseline): $p"
		continue
		;;
	ABSENT)
		# Unrecorded: a new observation, reported, never a failure.
		checked=$((checked + 1))
		reports="$reports; unrecorded (report only): $p $(printf '%s' "$staged" | LC_ALL=C sed 's/^COUNT \([0-9]*\) .*/COUNT \1/')"
		continue
		;;
	esac

	checked=$((checked + 1))
	problem=''
	case $staged in
	COUNT*)
		s_rest=${staged#COUNT }
		s_live=${s_rest%% *}
		s_exc=${s_rest#* }
		case $rec in
		HALT*)
			problem="snapshot records ${rec}, staged content counts live=$s_live excluded=$s_exc"
			;;
		COUNT*)
			r_rest=${rec#COUNT }
			r_live=${r_rest%% *}
			r_exc=${r_rest#* }
			if [ "$r_live" != "$s_live" ] || [ "$r_exc" != "$s_exc" ]; then
				problem="snapshot live=$r_live excluded=$r_exc, staged live=$s_live excluded=$s_exc"
			fi
			;;
		esac
		;;
	HALT*)
		case $rec in
		COUNT*)
			r_rest=${rec#COUNT }
			problem="snapshot records live=${r_rest%% *} excluded=${r_rest#* }, staged content HALTs (${staged#HALT })"
			;;
		HALT*)
			if [ "$rec" != "$staged" ]; then
				problem="halting identifier set moved: snapshot ${rec#HALT }, staged ${staged#HALT }"
			fi
			;;
		esac
		;;
	esac

	if [ -n "$problem" ]; then
		mismatch=$((mismatch + 1))
		say "$tag REJECT $p: $problem"
		say "  regenerate the snapshot and stage it in this same commit: $regen"
		say "  cascade procedure: $doc"
	fi
done

if [ "$mismatch" -gt 0 ]; then
	[ -n "$state" ] && printf 'mismatch\n' >>"$state"
	say "$tag commit rejected: $mismatch amended acceptance.md file(s) moved their AC count without a staged snapshot refresh"
	exit 1
fi
if [ "$checked" -gt 0 ]; then
	say "$tag checked $checked staged acceptance.md file(s) against $baseline, no count moved$reports"
fi
exit 0
