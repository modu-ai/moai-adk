#!/bin/sh
# Validate SSoT integrity of .github/required-checks.yml
# SPEC-V3R3-CI-AUTONOMY-001 W4-T05; hardened by SPEC-CI-VERDICT-INTEGRITY-001 M2.
#
# Dimensions:
#   (0) parser availability + YAML parse success — a missing yq or a parse
#       failure is a LOUD failure (E10/E11 repairs); an empty parse result can
#       never produce a passing verdict.
#   (a) auxiliary: items match actual workflow names
#   (b) branches.main.contexts ∩ auxiliary: = ∅ (no overlap)
#   (c) branches.release/*.contexts ∩ auxiliary: = ∅ (no overlap)
#   (d) PUBLISHABILITY: every required context must be among the check names
#       the repository's workflows can publish — job display names, with
#       ${{ matrix.key }} composites expanded over declared matrix arrays AND
#       matrix.include entries, and bare names on matrix jobs suffixed with
#       the matrix values (the CodeQL `Analyze (Go)` + language: [go] →
#       `Analyze (Go) (go)` shape). A phantom context fails naming it (E24).
#       KNOWN LIMITATION (documented debt): dimension D is the UNION of all
#       workflows' publishable names — it does not yet verify per-event /
#       per-branch publishability (a schedule-only check required on main, or
#       a main-only check required on release/*, passes this dimension).
#       Follow-up: parse each workflow's `on:` triggers + branch filters
#       per required context (gate-measured gap, card t1534 close round).
#
# Usage:
#   ./scripts/ci-mirror/validate-required-checks.sh   # from the repo root
#   echo $?  # 0 = pass, 1 = fail, 2 = missing file
set -eu

REQUIRED_CHECKS_FILE=".github/required-checks.yml"

# ---- (0) parser availability + parse success (LOUD failures) --------------
if ! command -v yq >/dev/null 2>&1; then
	echo "✗ yq not found — cannot validate $REQUIRED_CHECKS_FILE (refusing a vacuous pass)" >&2
	exit 1
fi
if ! yq eval '.' "$REQUIRED_CHECKS_FILE" >/dev/null 2>&1; then
	echo "✗ $REQUIRED_CHECKS_FILE failed to parse as YAML — refusing the vacuous pass" >&2
	exit 1
fi

exit_code=0
fail() {
	echo "✗ $*" >&2
	exit_code=1
}

# ---- Build the PUBLISHED check-name set from all workflow files ------------
# POSIX-awk: no subscripted split targets; matrix values are space-joined.
# Per-job matrix memory resets at every job boundary and at every top-level
# key (LG/Addendum: comment lines at column 0 do NOT reset the in-jobs scan,
# but stale matrix memory from a previous job must never leak into the next).
# P2-U: the published-name list goes to an UNPREDICTABLE mktemp path — a
# predictable /tmp/t1534-published-$$ lets a planted symlink overwrite an
# outside target (reviewer-measured sentinel rewrite, validator exit 0).
published="$(mktemp "${TMPDIR:-/tmp}/t1534-published-XXXXXXXX")"
trap 'rm -f "$published"' EXIT INT TERM
: > "$published"
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
	[ -f "$wf" ] || continue
	awk '
	# GATE-8: strip a trailing UNQUOTED comment (`name: Lint # required`) —
	# quote-aware: a # inside quotes is literal value text.
	function strip_comment(s,   i, c, q) {
		q = ""
		for (i = 1; i <= length(s); i++) {
			c = substr(s, i, 1)
			if (q != "") {
				if (c == q) q = ""
			} else {
				if (c == "\"" || c == "\047") q = c
				else if (c == "#" && i > 1 && substr(s, i - 1, 1) == " ") {
					# rtrim too — the comment strip leaves `Lint ` and the
					# trailing space judged a valid check phantom
					s = substr(s, 1, i - 1)
					sub(/[[:space:]]+$/, "", s)
					return s
				}
			}
		}
		return s
	}
	function emit() {
		if (!has_name) return
		# include-only matrices (the matrix.include form) have nk == 0 — the
		# tuple loop below is their publish path; never early-return. Plain
		# jobs (no matrix, no include) print their bare name once.
		if (nk == 0 && inc_n == 0) { print name; has_name = 0; return }
		if (nk > 0) {
		n = 1
		lines[1] = name
		sufs[1] = ""
		for (i = 1; i <= nk; i++) {
			k = dims[i]
			m = split(mvals[k], vals_arr, " ")
			newn = 0
			for (j = 1; j <= n; j++) {
				for (q = 1; q <= m; q++) {
					s = lines[j]
					# GATE-8: the expression whitespace is OPTIONAL —
					# `${{matrix.os}}` is a valid Actions expression too.
					gsub("\\$\\{\\{[[:space:]]*matrix\\." k "[[:space:]]*\\}\\}", vals_arr[q], s)
					newn++
					newlines[newn] = s
					ns = sufs[j]
					newsufs[newn] = (ns == "") ? vals_arr[q] : ns " " vals_arr[q]
				}
			}
			n = newn
			for (j = 1; j <= n; j++) { lines[j] = newlines[j]; sufs[j] = newsufs[j] }
		}
		for (j = 1; j <= n; j++) {
			# GATE-4 P2: matrix.exclude subtraction — GitHub does NOT publish
			# a combination when SOME exclude entry matches it on EVERY
			# key:value pair (partial matches do not remove). Pre-repair the
			# validator counted excluded combinations as publishable and a
			# required context GitHub can never run passed Dimension D
			# silently (run-matrix-exclude.sh repro: false-green exit 0).
			excluded = 0
			nsuf = split(sufs[j], svals, " ")
			if (nsuf == nk) {
				for (e = 1; e <= ex_n && !excluded; e++) {
					matches = 1
					for (pk in exval) {
						split(pk, pr, SUBSEP)
						if (pr[1] + 0 != e) continue
						hit = 0
						for (d = 1; d <= nk; d++)
							if (dims[d] == pr[2] && svals[d] == exval[pk]) { hit = 1; break }
						if (!hit) { matches = 0; break }
					}
					if (matches) excluded = 1
				}
			}
			if (excluded) continue
			# GATE-6: include tuples whose in-matrix keys all match this
			# combination MERGE into it — GitHub adds their out-of-matrix
			# key:value pairs to the combination and never overwrites
			# original values. GATE-7: the merge applies to EVERY compatible
			# combination (merged[t] marks only the standalone-suppression
			# flag, never a per-combination skip — a tuple merged into the
			# ubuntu leg still merges into the windows leg), and a later
			# tuple assigning the same key OVERWRITES the earlier value
			# (GitHub keeps the last include value).
			# (Runs AFTER the exclude check: excluded combinations no longer
			# exist for include to merge into.)
			en = 0
			for (t = 1; t <= inc_n; t++) {
				allmatch = 1
				for (pk in incval) {
					split(pk, pr2, SUBSEP)
					if (pr2[1] + 0 != t) continue
					d2 = -1
					for (dd = 1; dd <= nk; dd++)
						if (dims[dd] == pr2[2]) { d2 = dd; break }
					if (d2 > 0 && svals[d2] != incval[pk]) { allmatch = 0; break }
				}
				if (!allmatch) continue
				for (pk in incval) {
					split(pk, pr2, SUBSEP)
					if (pr2[1] + 0 != t) continue
					indim = 0
					for (dd = 1; dd <= nk; dd++)
						if (dims[dd] == pr2[2]) { indim = 1; break }
					if (indim) continue
					seen = 0
					for (q = 1; q <= en; q++) {
						if (ek[q] == pr2[2]) { ev[q] = incval[pk]; seen = 1; break }
					}
					if (!seen) {
						en++
						ek[en] = pr2[2]
						ev[en] = incval[pk]
					}
				}
				merged[t] = 1
			}
			# P2-B: the bare-name suffix applies ONLY when the job HAS a
			# matrix but its name never carried a ${{ matrix.* }} expression
			# (CodeQL `Analyze (Go)` + language: [go] → `Analyze (Go) (go)`).
			# P2-Q: the suffix is PER-COMBINATION — each emitted context keeps
			# its own tuple of matrix values (first[i]-only dropped valid
			# combinations like `Test (windows-latest)` on the floor).
			# GATE-6: apply merged include values — out-of-matrix keys fill
			# their ${{ matrix.* }} expressions and extend the suffix.
			outl = lines[j]
			for (q = 1; q <= en; q++)
				gsub("\\$\\{\\{[[:space:]]*matrix\\." ek[q] "[[:space:]]*\\}\\}", ev[q], outl)
			sfx2 = sufs[j]
			for (q = 1; q <= en; q++)
				sfx2 = (sfx2 == "") ? ev[q] : sfx2 " " ev[q]
			if (nk > 0 && !had_ref) {
				print outl " (" sfx2 ")"
			} else {
				print outl
			}
		}
		}
		for (t = 1; t <= inc_n; t++) {
			# GATE-6: a tuple that merged into a product combination does
			# not ALSO emit as a standalone combination.
			if (merged[t]) continue
			# GATE-5: exclude does NOT apply to include tuples — GitHub
			# processes exclude BEFORE include, so an include tuple can
			# RE-ADD a combination the exclude removed; the round-4
			# subtraction here deleted such re-added combinations and
			# judged a publishable context phantom.
			line = name
			for (kk in incval) {
				split(kk, pair, SUBSEP)
				if (pair[1] + 0 != t) continue
				if (incval[kk] == "") continue
				gsub("\\$\\{\\{[[:space:]]*matrix\\." pair[2] "[[:space:]]*\\}\\}", incval[kk], line)
			}
			if (had_ref) {
				# expression-bearing name: print only when every matrix
				# expression was substituted (residual ${{ }} prints raw)
				if (line !~ /\$\{\{/) print line
			} else {
				# GATE-5: a BARE job name on a matrix job is published with
				# the per-tuple suffix — values in matrix-dims order when
				# declared arrays exist, else the tuple key order (an
				# apostrophe here would break the single-quoted awk)
				# (include-only matrices have no dims). Pre-fix this
				# printed the bare name only and judged the suffixed
				# context phantom.
				sfx = ""
				if (nk > 0) {
					for (d = 1; d <= nk; d++) {
						tv = incval[t, dims[d]]
						if (tv != "") sfx = (sfx == "") ? tv : sfx " " tv
					}
				}
				if (sfx == "") {
					for (i = 1; i <= ikt[t]; i++) {
						tv = incval[t, incord[t, i]]
						if (tv != "") sfx = (sfx == "") ? tv : sfx " " tv
					}
				}
				if (sfx == "") print line
				else print line " (" sfx ")"
			}
		}
		has_name = 0
	}
	function reset_job_mem() {
		nk = 0; inc_n = 0; had_ref = 0; ex_n = 0; mmode = "inc"; bdim_key = ""
		delete dims; delete mvals; delete incval; delete exval; delete ikt; delete incord; delete merged
	}
	BEGIN { in_jobs = 0; has_name = 0; nk = 0; inc_n = 0; in_steps = 0; in_strategy = 0; in_matrix = 0; mmode = "inc"; ex_n = 0; bdim_key = "" }
	{
		ind = 0
		while (substr($0, ind + 1, 1) == " ") ind++
	}
	# GATE-7: comments and blank lines are NOT structural at ANY indent — a
	# two-space-indented comment between name: and strategy: hit the
	# job-boundary rule (ind == 2) and reset the matrix memory, judging a
	# valid workflow context phantom. (Extends GATE-2, which guarded only
	# column 0.)
	$0 ~ /^[[:space:]]*$/ { next }
	$0 ~ /^[[:space:]]*#/ { next }
	ind == 0 {
		if (has_name) emit()
		if ($0 ~ /^[A-Za-z_][A-Za-z_0-9-]*:/) { in_jobs = ($0 ~ /^jobs:/) ? 1 : 0; reset_job_mem() }
		in_steps = 0; in_strategy = 0; in_matrix = 0
		next
	}
	!in_jobs { next }
	ind == 2 && $0 !~ /^[[:space:]]*$/ {
		if (has_name) emit()
		reset_job_mem()
		# GATE-6: the job ID is the DEFAULT name — `name:` is optional in
		# GitHub workflows and a nameless job publishes its ID as the
		# check name. A later name: line overwrites this.
		jid = $0
		sub(/^[[:space:]]*/, "", jid)
		sub(/:.*/, "", jid)
		gsub(/^["\047]|["\047]$/, "", jid)
		name = jid
		has_name = 1
		had_ref = 0
		in_steps = 0; in_strategy = 0; in_matrix = 0
		next
	}
	ind == 4 && $0 ~ /^[[:space:]]*name:/ {
		name = strip_comment($0)
		sub(/^[[:space:]]*name:[[:space:]]*/, "", name)
		# GATE-5: strip BOTH quote styles — a single-quoted name (Lint in
		# ASCII single quotes) was
		# previously read as the literal string WITH the quotes still on it and a
		# perfectly valid job name judged phantom (octal \047 = single
		# quote; the awk program itself is single-quote-wrapped).
		gsub(/^["\047]|["\047]$/, "", name)
		has_name = 1
		had_ref = (name ~ /\$\{\{[[:space:]]*matrix\./) ? 1 : 0
		next
	}
	ind == 4 && $0 ~ /^[[:space:]]*steps:/ { in_steps = 1; next }
	ind == 4 && $0 ~ /^[[:space:]]*strategy:/ { in_strategy = 1; in_matrix = 0; next }
	ind == 4 { in_matrix = 0; next }
	in_strategy && ind == 6 && $0 ~ /^[[:space:]]*matrix:/ { in_matrix = 1; mmode = "inc"; next }
	# GATE-4 P2: `include:` / `exclude:` section markers under matrix: set the
	# tuple mode — without this an exclude entry was parsed as an include
	# tuple and the excluded combination stayed in the publishable set.
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*include:/ { mmode = "inc"; bdim_key = ""; next }
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*exclude:/ { mmode = "excl"; bdim_key = ""; next }
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9]*:[[:space:]]*$/ {
		# GATE-6: block-form array — `os:` with the values as `- ` items
		# below (the same YAML meaning as the flow form `os: [a, b]`;
		# pre-fix only the flow form parsed and the block form judged
		# valid contexts phantom).
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line); sub(/:[[:space:]]*$/, "", line)
		bdim_key = line
		next
	}
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9]*:[[:space:]]*\[/ {
		bdim_key = ""
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^[]*\[/, "", v); sub(/\][[:space:]]*$/, "", v)
		gsub(/[[:space:]]/, "", v); gsub(/["\047]/, "", v)
		nk++
		dims[nk] = k
		mvals[k] = v
		gsub(/,/, " ", mvals[k])
		next
	}
	in_matrix && ind == 10 && $0 ~ /^[[:space:]]*- / {
		line = strip_comment($0); sub(/^[[:space:]]*-[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); gsub(/^["\047]|["\047]$/, "", v)
		is_pair = (line ~ /:/)
		if (bdim_key != "" && !is_pair) {
			# GATE-6: block-form dim item — a bare value appended to the
			# dim declared by the `key:` line above.
			gsub(/[[:space:]]/, "", line); gsub(/["\047]/, "", line)
			if (!(bdim_key in mvals)) { nk++; dims[nk] = bdim_key; mvals[bdim_key] = "" }
			if (mvals[bdim_key] == "") mvals[bdim_key] = line
			else mvals[bdim_key] = mvals[bdim_key] " " line
			next
		}
		if (mmode == "excl") { ex_n++; exval[ex_n, k] = v }
		else {
			# record the tuple key order — the bare-name suffix
			# (GATE-5) needs it for include-only matrices with no dims
			inc_n++
			incval[inc_n, k] = v
			ikt[inc_n]++
			incord[inc_n, ikt[inc_n]] = k
		}
		next
	}
	in_matrix && ind == 12 && mmode == "excl" && ex_n > 0 && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); gsub(/^["\047]|["\047]$/, "", v)
		exval[ex_n, k] = v
		next
	}
	in_matrix && ind == 12 && inc_n > 0 && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); gsub(/^["\047]|["\047]$/, "", v)
		incval[inc_n, k] = v
		ikt[inc_n]++
		incord[inc_n, ikt[inc_n]] = k
		next
	}
	in_matrix && ind <= 6 && $0 !~ /^[[:space:]]*$/ { in_matrix = 0; bdim_key = "" }
	END { if (has_name) emit() }
	' "$wf" >> "$published"
done
sort -u -o "$published" "$published"

# ---- Dimension D: every required context must be publishable ---------------
echo "=== Dimension D: every required context must be workflow-publishable ==="
branch_keys="$(yq -r '.branches | keys | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null || true)"
for branch_key in $branch_keys; do
	while IFS= read -r context; do
		[ -n "$context" ] || continue
		if grep -qxF "$context" "$published"; then
			echo "✓ [$branch_key] $context is publishable"
		else
			fail "[$branch_key] required context '$context' is NOT publishable by any workflow (phantom context)"
		fi
	done <<EOF
$(yq -r ".branches[\"$branch_key\"].contexts // [] | .[]" "$REQUIRED_CHECKS_FILE" 2>/dev/null || true)
EOF
done

# ---- Dimension A: auxiliary items must correspond to workflow files --------
echo "=== Dimension A: Validating auxiliary → workflow name mapping ==="
auxiliary_items="$(yq -r '.auxiliary // [] | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null || true)"
while IFS= read -r aux_item; do
	[ -n "$aux_item" ] || continue
	if [ -f ".github/workflows/${aux_item}.yml" ] || [ -f ".github/workflows/${aux_item}.yaml" ]; then
		echo "✓ $aux_item (workflow file found)"
	else
		fail "auxiliary item '$aux_item' has no matching workflow file"
	fi
done <<EOF
$auxiliary_items
EOF

# ---- Dimensions B/C: auxiliary ∩ required contexts = ∅ ---------------------
echo "=== Dimension B: branches.main.contexts ∩ auxiliary = ∅ ==="
while IFS= read -r aux_item; do
	[ -n "$aux_item" ] || continue
	if yq -r '.branches.main.contexts // [] | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null | grep -qxF "$aux_item"; then
		fail "auxiliary item '$aux_item' found in branches.main.contexts"
	else
		echo "✓ $aux_item not in branches.main.contexts"
	fi
done <<EOF
$auxiliary_items
EOF

echo "=== Dimension C: branches.release/*.contexts ∩ auxiliary = ∅ ==="
while IFS= read -r aux_item; do
	[ -n "$aux_item" ] || continue
	if yq -r '.branches["release/*"].contexts // [] | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null | grep -qxF "$aux_item"; then
		fail "auxiliary item '$aux_item' found in branches.release/*.contexts"
	else
		echo "✓ $aux_item not in branches.release/*.contexts"
	fi
done <<EOF
$auxiliary_items
EOF

rm -f "$published"

echo ""
if [ "$exit_code" -eq 0 ]; then
	echo "✅ All validations passed"
else
	echo "❌ Validation failed" >&2
fi
exit "$exit_code"
