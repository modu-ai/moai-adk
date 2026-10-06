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
normalized="$(mktemp "${TMPDIR:-/tmp}/t1534-normalized-XXXXXXXX")"
trap 'rm -f "$published" "$normalized"' EXIT INT TERM
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
	[ -f "$wf" ] || continue
	# GATE-9: parse YAML STRUCTURE, not line shapes — the workflow is
	# re-rendered to its canonical BLOCK form by yq first (flow-form
	# matrices, inline collections and shorthand all expand), and the
	# line parser below reads one canonical shape. A yq failure aborts:
	# an unparseable workflow must not silently contribute no names
	# (the same vacuous-pass class as an unreadable SSoT).
	# GATE-13: expand YAML aliases BEFORE parsing — yq preserves `os: *oses`
	# verbatim and the line parser then misses the axis and judges a real
	# check phantom; explode(.) resolves every alias to its anchor value.
	if ! yq -P 'explode(.)' "$wf" > "$normalized" 2>/dev/null; then
		fail "workflow $wf failed to parse as YAML (yq) — its checks cannot be verified publishable"
		continue
	fi
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
	# GATE-14: strip quotes only when they are PAIRED (same char leading
	# and trailing) — a bare trailing apostrophe is VALUE TEXT
	# (a trailing apostrophe is VALUE TEXT, not a quoting character)
	# GATE-15: a single-quoted string decodes its ESCAPE — YAML writes an
	# inner apostrophe as two, so the parsed value restores one.
	function strip_quotes(s,   f) {
		if (length(s) >= 2) {
			f = substr(s, 1, 1)
			if ((f == "\"" || f == "\047") && substr(s, length(s), 1) == f) {
				s = substr(s, 2, length(s) - 2)
				if (f == "\047") gsub(/\047\047/, "\047", s)
			}
		}
		return s
	}
	# GATE-14: bracket form normalizes to dot form — `${{ matrix['os'] }}`
	# is the same reference as `${{ matrix.os }}`. Manual capture via
	# match/substr: BSD awk has no backreferences in gsub replacements.
	function norm_bracket(s,   m) {
		while (match(s, /matrix\[["\047]?[A-Za-z_][A-Za-z_0-9-]*["\047]?\]/)) {
			m = substr(s, RSTART, RLENGTH)
			sub(/^matrix\[["\047]?/, "", m)
			sub(/["\047]?\]$/, "", m)
			s = substr(s, 1, RSTART - 1) "matrix." m substr(s, RSTART + RLENGTH)
		}
		return s
	}
	# GATE-11: literal-expression substitution — gsub treats & and
	# backslash in its replacement as grammar (an ampersand expands to the
	# whole match), and the escaping dance is a self-referential trap on
	# BSD awk (the escaped replacement is recovered as a bare ampersand
	# again). Split on the pattern instead: the replacement text enters as
	# plain data, ampersands and all.
	function subst_literal(str, ere, lit,   a, i, out, n) {
		n = split(str, a, ere)
		out = a[1]
		for (i = 2; i <= n; i++) out = out lit a[i]
		return out
	}
	function emit() {
		if (!has_name) return
		# GATE-14/17: object axes. SOLE object axis (no product dims, no
		# include tuples): each item is ONE combination built from THAT
		# item only — a mixed field-set would approve combinations GitHub
		# never publishes. Combined with product dims or include tuples,
		# the first item resolves the sub-field references (documented
		# approximation).
		if (n_oaxes > 0 && nk == 0 && inc_n == 0) {
			n2 = 1
			ob_line[1] = name
			ob_ids[1] = ""
			for (oa = 1; oa <= n_oaxes; oa++) {
				axis = oaxes[oa]
				base = n2
				for (c = 1; c <= base; c++) {
					for (idx = 1; idx <= oin[axis]; idx++) {
						n2++
						ob_line[n2] = ob_line[c]
						ob_ids[n2] = ob_ids[c] SUBSEP axis SUBSEP idx
					}
				}
			}
			# collect the REFERENCED sub-fields from the template — a field
			# an item lacks evaluates as the EMPTY string on GitHub, so
			# every reference substitutes (value or empty) and nothing
			# else: a mixed field-set can never approve a combination.
			nref = 0
			s = name
			while (match(s, /matrix\.[A-Za-z_][A-Za-z_0-9-]*\.[A-Za-z_][A-Za-z_0-9-]*/)) {
				ref = substr(s, RSTART + 7, RLENGTH - 7)
				nref++
				refax[nref] = substr(ref, 1, index(ref, ".") - 1)
				refsub[nref] = substr(ref, index(ref, ".") + 1)
				s = substr(s, RSTART + RLENGTH)
			}
			for (c = 1; c <= n2; c++) {
				# GATE-18: exclude applies to object-axis combinations too.
				# An entry matches when EVERY stored pair hits the
				# combination: a dotted pair (group.field, from the ind-12
				# handler under a bare group key) reads that item sub-field;
				# a flat key or a zero-pair entry never matches here.
				excl_c = 0
				for (e = 1; e <= ex_n && !excl_c; e++) {
					em = 1
					enp = 0
					for (pk in exval) {
						split(pk, pr, SUBSEP)
						if (pr[1] + 0 != e) continue
						enp++
						path = pr[2]
						if (index(path, ".") == 0) { em = 0; break }
						ax2 = substr(path, 1, index(path, ".") - 1)
						sub2 = substr(path, index(path, ".") + 1)
						hit2 = 0
						cnt2 = split(ob_ids[c], oidp2, SUBSEP)
						for (p = 2; p <= cnt2; p += 2) {
							if (oidp2[p] != ax2) continue
							for (fk2 in objitems) {
								split(fk2, fp2, SUBSEP)
								if (fp2[1] != ax2 || fp2[2] + 0 != oidp2[p + 1] + 0) continue
								if (fp2[3] == sub2 && objitems[fk2] == exval[pk]) hit2 = 1
							}
						}
						if (!hit2) { em = 0; break }
					}
					if (em && enp > 0) excl_c = 1
				}
				if (excl_c) continue
				outl = ob_line[c]
				cnt = split(ob_ids[c], oidp, SUBSEP)
				for (p = 2; p <= cnt; p += 2) {
					ax = oidp[p]
					ix = oidp[p + 1]
					for (r = 1; r <= nref; r++) {
						if (refax[r] != ax) continue
						lit = ""
						for (fk in objitems) {
							split(fk, fp, SUBSEP)
							if (fp[1] != ax || fp[2] + 0 != ix || fp[3] != refsub[r]) continue
							lit = objitems[fk]
						}
						outl = subst_literal(outl, "\\$\\{\\{[[:space:]]*matrix\\." ax "\\." refsub[r] "[[:space:]]*\\}\\}", lit)
					}
				}
				if (outl !~ /\$\{\{/) print outl
			}
			has_name = 0
			return
		}
		if (n_oaxes > 0) {
			for (oa = 1; oa <= n_oaxes; oa++) {
				axis = oaxes[oa]
				for (fk in objitems) {
					split(fk, fp, SUBSEP)
					if (fp[1] != axis || fp[2] + 0 != 1) continue
					name = subst_literal(name, "\\$\\{\\{[[:space:]]*matrix\\." axis "\\." fp[3] "[[:space:]]*\\}\\}", objitems[fk])
				}
			}
		}
		# include-only matrices (the matrix.include form) have nk == 0 — the
		# tuple loop below is their publish path; never early-return. Plain
		# jobs (no matrix, no include) print their bare name once.
		if (nk == 0 && inc_n == 0) { print name; has_name = 0; return }
		if (nk > 0) {
		n = 1
		lines[1] = name
		sufs[1] = ""
		sufcnt[1] = 0
		for (i = 1; i <= nk; i++) {
			k = dims[i]
			m = split(mvals[k], vals_arr, SUBSEP)
			# GATE-17: a stored EMPTY value splits to zero parts — it is
			# still ONE combination (`option: [""]` publishes `Test ()`).
			# GATE-18: an explicitly EMPTY array (zcombo) contributes ZERO
			# combinations — no fixup there.
			if (m == 0 && mvals[k] == "" && !zcombo) { vals_arr[1] = ""; m = 1 }
			newn = 0
			for (j = 1; j <= n; j++) {
				for (q = 1; q <= m; q++) {
					s = lines[j]
					# GATE-8: the expression whitespace is OPTIONAL —
					# `${{matrix.os}}` is a valid Actions expression too.
					s = subst_literal(s, "\\$\\{\\{[[:space:]]*matrix\\." k "[[:space:]]*\\}\\}", vals_arr[q])
					newn++
					newlines[newn] = s
					ns = sufs[j]
					# GATE-18: positional PAIR COUNT, not string emptiness —
					# a leading EMPTY value leaves the suffix string empty
					# and the old ns == "" test lost the position. The count
					# lands in newsufcnt (never in place: newn == j would
					# corrupt the not-yet-read source rows).
					newsufs[newn] = (sufcnt[j] == 0) ? vals_arr[q] : ns SUBSEP vals_arr[q]
					newsufcnt[newn] = sufcnt[j] + 1
				}
			}
			n = newn
			for (j = 1; j <= n; j++) { lines[j] = newlines[j]; sufs[j] = newsufs[j]; sufcnt[j] = newsufcnt[j] }
		}
		for (j = 1; j <= n; j++) {
			# GATE-4 P2: matrix.exclude subtraction — GitHub does NOT publish
			# a combination when SOME exclude entry matches it on EVERY
			# key:value pair (partial matches do not remove). Pre-repair the
			# validator counted excluded combinations as publishable and a
			# required context GitHub can never run passed Dimension D
			# silently (run-matrix-exclude.sh repro: false-green exit 0).
			excluded = 0
			nsuf = split(sufs[j], svals, SUBSEP)
			# GATE-18: the positional gate counts PAIRS (sufcnt), not the
			# split count — a leading empty value made nsuf undercount and
			# silently skip the exclude check. A dotted key is an
			# object-axis pair and matches no scalar dimension here.
			if (sufcnt[j] == nk) {
				for (e = 1; e <= ex_n && !excluded; e++) {
					matches = 1
					np = 0
					for (pk in exval) {
						split(pk, pr, SUBSEP)
						if (pr[1] + 0 != e) continue
						np++
						hit = 0
						for (d = 1; d <= nk; d++)
							if (dims[d] == pr[2] && svals[d] == exval[pk]) { hit = 1; break }
						if (!hit) { matches = 0; break }
					}
					if (matches && np > 0) excluded = 1
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
				outl = subst_literal(outl, "\\$\\{\\{[[:space:]]*matrix\\." ek[q] "[[:space:]]*\\}\\}", ev[q])
			# GATE-18: a field the matrix never declares evaluates EMPTY on
			# GitHub — a residual expression after all substitution is an
			# unset field, not an unverifiable name; a partial include
			# tuple must not poison the whole published name.
			# GATE-18b: an EXPRESSION-declared axis (exprdim) has real
			# values this parser cannot compute — its reference parks on a
			# SUBSEP frame through the fill and is restored raw, so no
			# empty-valued combination is fabricated for it.
			for (xd in exprdim)
				outl = subst_literal(outl, "\\$\\{\\{[[:space:]]*matrix\\." xd "[[:space:]]*\\}\\}", SUBSEP xd SUBSEP)
			outl = subst_literal(outl, "\\$\\{\\{[[:space:]]*matrix\\.[A-Za-z_][A-Za-z_0-9-]*[[:space:]]*\\}\\}", "")
			for (xd in exprdim)
				outl = subst_literal(outl, SUBSEP xd SUBSEP, "${{ matrix." xd " }}")
			# GATE-11: the include out-of-matrix fields feed ONLY the
			# expression substitution — the auto suffix reflects the
			# ORIGINAL matrix axes alone (GitHub names
			# os:[ubuntu]+include[extra:smoke] as `Test (ubuntu-latest)`,
			# never `Test (ubuntu-latest smoke)`).
			# GATE-12: GitHub joins MULTIPLE matrix values with comma+space
			# (`Test (ubuntu-latest, 18)`), not a bare space.
			if (nk > 0 && !had_ref) {
				sfx2 = sufs[j]
				gsub(SUBSEP, ", ", sfx2)
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
				# GATE-15: an EXPLICIT empty value substitutes as empty —
				# only an ABSENT field is skipped (presence, not value).
				if (!((kk) in incset)) continue
				line = subst_literal(line, "\\$\\{\\{[[:space:]]*matrix\\." pair[2] "[[:space:]]*\\}\\}", incval[kk])
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
						if (tv != "") sfx = (sfx == "") ? tv : sfx SUBSEP tv
					}
				}
				if (sfx == "") {
					for (i = 1; i <= ikt[t]; i++) {
						tv = incval[t, incord[t, i]]
						if (tv != "") sfx = (sfx == "") ? tv : sfx SUBSEP tv
					}
				}
				if (sfx == "") print line
				else {
					gsub(SUBSEP, ", ", sfx)
					print line " (" sfx ")"
				}
			}
		}
		has_name = 0
	}
	function reset_job_mem() {
		nk = 0; inc_n = 0; had_ref = 0; ex_n = 0; mmode = "inc"; bdim_key = ""; zcombo = 0
		delete dims; delete mvals; delete incval; delete exval; delete ikt; delete incord; delete merged; delete objitems; delete oin; delete oaxes; delete incset; delete mcnt; delete exgrp; delete sufcnt; delete newsufcnt; delete exprdim
	}
	BEGIN { in_jobs = 0; has_name = 0; nk = 0; inc_n = 0; in_steps = 0; in_strategy = 0; in_matrix = 0; mmode = "inc"; ex_n = 0; bdim_key = ""; n_oaxes = 0; zcombo = 0 }
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
		# GATE-13: decode the single-quote ESCAPE before comparing —
		# YAML writes an inner apostrophe as two apostrophes
		# (Lint-double-apostrophe-strict-double resolves to Lint-quote-strict).
		# Quote stripping is PAIRED — the trailing-quote strip alone bit
		# the trailing apostrophe off an UNquoted value like
		# Lint-quote-strict (yq renders the escaped name unquoted).
		sq_quoted = 0
		if (name ~ /^\047/ && name ~ /\047$/) {
			sq_quoted = 1
			name = substr(name, 2, length(name) - 2)
		} else if (name ~ /^"/ && name ~ /"$/) {
			name = substr(name, 2, length(name) - 2)
		}
		if (sq_quoted) gsub(/\047\047/, "\047", name)
		name = norm_bracket(name)
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
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9-]*:[[:space:]]*$/ {
		# GATE-6: block-form array — `os:` with the values as `- ` items
		# below (the same YAML meaning as the flow form `os: [a, b]`;
		# pre-fix only the flow form parsed and the block form judged
		# valid contexts phantom).
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line); sub(/:[[:space:]]*$/, "", line)
		bdim_key = line
		next
	}
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9-]*:[[:space:]]*\[/ {
		bdim_key = ""
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^[]*\[/, "", v); sub(/\][[:space:]]*$/, "", v)
		# GATE-9: values may contain SPACES — split on commas and trim each
		# item; pre-fix the whole list was whitespace-stripped, mangling a
		# value like "Build (linux amd64)" into "Build(linuxamd64)". Note
		# the space-joined suffix store then cannot distinguish a space
		# inside a value, so the exclude subtraction conservatively skips
		# such combinations (no false green — a kept combination is
		# over-inclusive, never under).
		parts_n = split(v, parts, ",")
		# GATE-17: a WHOLE-EMPTY value list is one empty combination —
		# split("", sep) returns 0 parts and would drop the combination.
		# GATE-18: a genuinely EMPTY array (`os: []`) declares ZERO values
		# — GitHub runs the job zero times and publishes nothing. (The
		# GATE-17 shape is a list holding one empty string, raw v == two
		# quote characters — that still publishes `Test ()`.)
		if (parts_n == 0 && v == "") { zcombo = 1 }
		mvals[k] = ""
		mcnt[k] = 0
		for (pp = 1; pp <= parts_n; pp++) {
			pv = parts[pp]
			gsub(/^[[:space:]]+/, "", pv)
			gsub(/[[:space:]]+$/, "", pv)
			pv = strip_quotes(pv)
			# GATE-16: an EMPTY value is a real combination (`option:
			# ["", race]` publishes `Test ()` first) — count-tracked, not
			# dropped for being falsy.
			mcnt[k]++
			# GATE-9: the value store joins on SUBSEP — a space now stays
			# INSIDE a value; split and print convert it back.
			mvals[k] = (mcnt[k] == 1) ? pv : mvals[k] SUBSEP pv
		}
		nk++
		dims[nk] = k
		next
	}
	in_matrix && ind == 8 && $0 ~ /^[[:space:]]*[A-Za-z_][A-Za-z_0-9-]*:[[:space:]]*[^[]/ {
		# GATE-18b: an axis whose value is neither a flow array nor empty
		# carries an EXPRESSION (`os: ${{ fromJSON(...) }}`) — the real
		# values are uncomputable here, so the axis is recorded (never
		# declared as a dimension) and the unset-field fill PARKS its
		# reference: GitHub publishes the evaluated name no line parser
		# can compute, and a raw name never matches a concrete required
		# context — the conservative direction.
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		exprdim[k] = 1
		next
	}
	in_matrix && ind == 10 && $0 ~ /^[[:space:]]*- / {
		line = strip_comment($0); sub(/^[[:space:]]*-[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v)
		# GATE-18: the RAW value separates a nested group opener (`target:`,
		# nothing after the colon) from an explicit empty string (`os: ""`).
		vraw = v
		v = strip_quotes(v)
		# GATE-10: a colon does NOT make a mapping — in YAML, `- node:20`
		# is a STRING scalar (no space after the colon) while `- color: green`
		# is a mapping. The bare is_pair colon test mis-routed spaced-out
		# string values into the tuple path and judged their checks phantom.
		# GATE-11: a QUOTED item is a scalar by node type — yq re-renders a
		# value such as node-colon-space as a QUOTED string whose inner
		# colon+space is literal value text, not a mapping separator.
		is_pair = (line !~ /^["\047]/) && (line ~ /:[[:space:]]/ || line ~ /:$/)
		if (bdim_key != "") {
			# GATE-14: a list item under a DECLARED axis key belongs to
			# that axis, never to include — an object item
			# (target: [{os: ubuntu-latest}]) is a sub-field the template
			# reads as matrix.target.os.
			if (is_pair) {
				# GATE-17: the FIRST field of an item opens a NEW object —
				# follow-up fields (ind 12) extend the SAME object, so the
				# item index is allocated here and never in the ind-12
				# path. Object axes expand per-item (not first-value).
				oi = ++oin[bdim_key]
				objitems[bdim_key SUBSEP oi SUBSEP k] = v
				if (oin[bdim_key] == 1) { n_oaxes++; oaxes[n_oaxes] = bdim_key }
				next
			}
			# GATE-6: block-form dim item — a bare value appended to the
			# dim declared by the `key:` line above.
			# GATE-9: a value may contain SPACES — trim, never strip (the
			# whole-string whitespace strip mangled "ubuntu 24.04" into
			# "ubuntu24.04" and judged the real check phantom).
			# GATE-14: quotes strip only when PAIRED — a bare trailing
			# apostrophe is value text, not a quoting character
			gsub(/^[[:space:]]+/, "", line)
			gsub(/[[:space:]]+$/, "", line)
			line = strip_quotes(line)
			if (!(bdim_key in mvals)) { nk++; dims[nk] = bdim_key; mvals[bdim_key] = ""; mcnt[bdim_key] = 0 }
			# GATE-16: the value COUNT tracks membership — an EMPTY first
			# value (`option: ["", race]`) is a real combination, not an
			# uninitialized slot.
			mcnt[bdim_key]++
			if (mcnt[bdim_key] == 1) mvals[bdim_key] = line
			else mvals[bdim_key] = mvals[bdim_key] SUBSEP line
			next
		}
		if (mmode == "excl") {
			ex_n++
			# GATE-18: a bare key with NO raw value opens a NESTED group —
			# `exclude: - target:` + ind-12 `os: windows` is the dotted
			# pair target.os. An explicit empty string (os: "") is a PAIR
			# whose value is empty, never a group opener.
			if (vraw == "") exgrp[ex_n] = k
			else exval[ex_n, k] = v
		}
		else {
			# record the tuple key order — the bare-name suffix
			# (GATE-5) needs it for include-only matrices with no dims
			inc_n++
			incval[inc_n, k] = v
			incset[inc_n, k] = 1
			ikt[inc_n]++
			incord[inc_n, ikt[inc_n]] = k
		}
		next
	}
	# GATE-15: FOLLOW-UP fields of an object axis item sit at ind 12 —
	# target: [{os: ubuntu-latest, version: "18"}] normalizes with
	# `version:` on the next line, and it is still a sub-field of the AXIS,
	# never an include tuple.
	in_matrix && ind == 12 && bdim_key != "" && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); v = strip_quotes(v)
		objitems[bdim_key SUBSEP oi SUBSEP k] = v
		next
	}
	# GATE-18: a NESTED group field under a bare exclude key renders at
	# ind 14 in the yq-normalized form (`- target:` + child `os:` sits two
	# deeper than the tuple flat fields) — the dotted pair would never be
	# stored if only ind 12 matched.
	in_matrix && (ind == 12 || ind == 14) && mmode == "excl" && ex_n > 0 && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		# GATE-16: strip_quotes, not a trailing-quote strip — a value
		# ending in an apostrophe keeps it, or the exclude never matches
		# the real matrix value and an EXCLUDED combination passes as a
		# required check.
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); v = strip_quotes(v)
		# GATE-18: under a bare group key the field stores as the DOTTED
		# path (group.field) — the object-axis exclude matches on the dot
		# path, and the scalar matcher never confuses it with a dimension.
		if (exgrp[ex_n] != "") exval[ex_n, exgrp[ex_n] "." k] = v
		else exval[ex_n, k] = v
		next
	}
	in_matrix && ind == 12 && inc_n > 0 && $0 ~ /^[[:space:]]*[A-Za-z_]/ {
		line = strip_comment($0); sub(/^[[:space:]]*/, "", line)
		k = line; sub(/:.*/, "", k)
		v = line; sub(/^[^:]*:[[:space:]]*/, "", v); v = strip_quotes(v)
		incval[inc_n, k] = v
		incset[inc_n, k] = 1
		ikt[inc_n]++
		incord[inc_n, ikt[inc_n]] = k
		next
	}
	in_matrix && ind <= 6 && $0 !~ /^[[:space:]]*$/ { in_matrix = 0; bdim_key = "" }
	END { if (has_name) emit() }
	' "$normalized" >> "$published"
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
