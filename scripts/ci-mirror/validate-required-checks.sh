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
					gsub("\\$\\{\\{ matrix\\." k " }}", vals_arr[q], s)
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
			# P2-B: the bare-name suffix applies ONLY when the job HAS a
			# matrix but its name never carried a ${{ matrix.* }} expression
			# (CodeQL `Analyze (Go)` + language: [go] → `Analyze (Go) (go)`).
			# P2-Q: the suffix is PER-COMBINATION — each emitted context keeps
			# its own tuple of matrix values (first[i]-only dropped valid
			# combinations like `Test (windows-latest)` on the floor).
			if (nk > 0 && !had_ref) {
				print lines[j] " (" sufs[j] ")"
			} else {
				print lines[j]
			}
		}
		}
		for (t = 1; t <= inc_n; t++) {
			line = name
			fully = 1
			for (kk in incval) {
				split(kk, pair, SUBSEP)
				if (pair[1] + 0 != t) continue
				if (incval[kk] == "") { fully = 0; continue }
				gsub("\\$\\{\\{ matrix\\." pair[2] " }}", incval[kk], line)
			}
			if (line !~ /\$\{\{/) print line
		}
		has_name = 0
	}
	function reset_job_mem() {
		nk = 0; inc_n = 0; had_ref = 0
		delete dims; delete mvals; delete incval
	}
	BEGIN { in_jobs = 0; has_name = 0; nk = 0; inc_n = 0; in_steps = 0; in_strategy = 0; in_matrix = 0 }
	{
		ind = 0
		while (substr($0, ind + 1, 1) == " ") ind++
	}
	ind == 0 {
		# GATE-2: blank lines and column-0 comments must be skipped BEFORE any
		# emit — a blank line inside a job block (between name: and strategy:)
		# used to fire emit() with no matrix memory yet, publishing the bare
		# name and making the real suffixed context judge phantom.
		if ($0 ~ /^[[:space:]]*$/) next
		if ($0 ~ /^[[:space:]]*#/) next
		if (has_name) emit()
		if ($0 ~ /^[A-Za-z_][A-Za-z_0-9-]*:/) { in_jobs = ($0 ~ /^jobs:/) ? 1 : 0; reset_job_mem() }
		in_steps = 0; in_strategy = 0; in_matrix = 0
		next
	}
	!in_jobs { next }
	ind == 2 && $0 !~ /^[[:space:]]*$/ {
		if (has_name) emit()
		reset_job_mem()
		in_steps = 0; in_strategy = 0; in_matrix = 0
		next
	}
	ind == 4 && $0 ~ /^[[:space:]]*name:/ && !has_name {
		name = $0
		sub(/^[[:space:]]*name:[[:space:]]*/, "", name)
		gsub(/^"|"$/, "", name)
		has_name = 1
		had_ref = (name ~ /\$\{\{ matrix\./) ? 1 : 0
		next
	}
	ind == 4 && $0 ~ /^[[:space:]]*steps:/ { in_steps = 1; next }
	ind == 4 && $0 ~ /^[[:space:]]*strategy:/ { in_strategy = 1; in_matrix = 0; next }
	ind == 4 { in_matrix = 0; next }
	in_strategy && ind == 6 && $0 ~ /^[[:space:]]*matrix:/ { in_matrix = 1; next }
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9]*:[[:space:]]*\[/ {
		line = $0; sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^[]*\[/, "", v); sub(/\][[:space:]]*$/, "", v)
		gsub(/[[:space:]]/, "", v); gsub(/"/, "", v)
		nk++
		dims[nk] = k
		mvals[k] = v
		gsub(/,/, " ", mvals[k])
		next
	}
	in_matrix && ind == 10 && $0 ~ /^[[:space:]]*- / {
		inc_n++
		line = $0; sub(/^[[:space:]]*-[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); gsub(/^"|"$/, "", v)
		incval[inc_n, k] = v
		next
	}
	in_matrix && ind == 12 && inc_n > 0 && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = $0; sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); gsub(/^"|"$/, "", v)
		incval[inc_n, k] = v
		next
	}
	in_matrix && ind <= 6 && $0 !~ /^[[:space:]]*$/ { in_matrix = 0 }
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
