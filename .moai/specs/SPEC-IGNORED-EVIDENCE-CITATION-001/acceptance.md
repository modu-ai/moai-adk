---
id: SPEC-IGNORED-EVIDENCE-CITATION-001
title: "Acceptance criteria — ignored evidence citation repair"
version: "0.4.0"
created: 2026-08-31
updated: 2026-08-31
---

# Acceptance Criteria — SPEC-IGNORED-EVIDENCE-CITATION-001

## §A. Verification contract

**Every MUST criterion below is decided by a plain single command invocation whose output or exit
code decides it.** No MUST criterion is decided by a human reading a context window. The central
gate (AC-IEC-001) is a binary `grep -l | xargs grep -L` composition whose PASS condition is an empty
output (iter2 D2 replaced the original prose read; iter3 N1 made the obligation conditional on a
citation still being present, which is what makes it satisfiable under §C.3 treatment (a)).

**Execution-form constraint — this is load-bearing, not style.** This card runs in a
worktree-isolated session whose guard **refuses** any command it cannot statically verify. Measured
in this tree: a bare `for f in a b; do echo "$f"; done` and a bare `printf '%s' "$(grep -c x f)"` are
both refused before execution. Therefore every command below obeys:

- **Permitted**: plain invocations, multi-path arguments (`git grep -c … -- a b c`), multi-file
  `grep -c a b`, pipes (`| sort -u | wc -l`), and `echo "exit=$?"`.
- **Prohibited**: `for … done`, `$(…)` command substitution, and subshells `( … )`.
- **Prohibited: moving commands into a script file.** The guard cannot read inside a script, so
  every check would be bypassed for that payload. Routing around the guard is not satisfying it.
  Where a check cannot be expressed as one plain invocation, the criterion is made smaller.

Where a check needs N paths, it uses **one invocation with N path arguments** (which prints per-path
counts) rather than a loop.

**Pattern-scope constraint (iter2).** Every citation check matches `.moai/state/verify`, NOT the
broader `.moai/state/`. Measured: `.moai/reports/template-skill-improvement-plan-20260710.html`
contains **two** `.moai/state/` occurrences — line 684 (in scope) and line 529
(`.moai/state/loop-verdict-<id>.json`, a mechanism reference, census class B2, explicitly out of
scope). A `.moai/state/` pattern would sweep line 529 into this card and demand a repair the SPEC
forbids.

**Evidence location (AC-IEC-010 enforces this on itself).** Every command's verbatim output is
written under `.moai/reports/t381/verify/` — a **tracked** path. Nothing is written under
`.moai/state/`.

**Baseline — three-dot, resolved at verification time (iter4, P3).** The two diff criteria
(AC-IEC-006, AC-IEC-007) use `origin/develop...HEAD`, whose three-dot form diffs the **merge base**
of `origin/develop` and `HEAD` against `HEAD`. That means "this branch's own edits", which is the
question both criteria actually ask.

`3f03d9c36` remains the tree census.md measured and the tree every RED-now figure below was observed
in, but it is **no longer used as a diff baseline**. Measured 2026-08-31:

```
$ git merge-base origin/develop HEAD          → 3f03d9c369…   (identical to the old frozen SHA, today)
$ git rev-list --count 3f03d9c36..origin/develop   → 11
```

> **Why a frozen SHA was wrong, and why a *fresh* frozen SHA would be wrong too.** `origin/develop`
> advanced to `9328a5242`, and all four of AC-IEC-007's do-not-touch files changed in those 11
> commits — they are **t375's own landed edits**. Against a frozen `3f03d9c36`, the moment this lane
> absorbs `origin/develop` (which `CLAUDE.local.md` §4.1 `[HARD]` requires before integration) the
> criterion reports those four files as modified and **fails for changes this card never made**.
> Measured contrast, same files, same run:
>
> ```
> $ git diff --exit-code --stat 3f03d9c36 origin/develop -- .claude/agents/moai/manager-lead.md   → exit=1
> $ git diff --exit-code --stat origin/develop...HEAD -- <all 7 do-not-touch paths>                → exit=0
> ```
>
> Re-pinning to a newer SHA would reproduce the same defect at the next move. The three-dot form
> re-resolves the merge base on every run, so absorbing `origin/develop` does not invalidate it: after
> the absorb the merge base becomes `origin/develop` itself, and the diff still shows only this
> card's edits.

**RED-now column.** Each MUST criterion records what it prints **today, before any repair**, measured
in this tree. A criterion whose today-value already equals its PASS-value cannot go RED from this
card's work and is not a MUST — those are demoted to §D structural checks (iter2, D3).

---

## §B. AC matrix

| AC | Backing requirement | Decides | Class | RED today? |
|---|---|---|---|---|
| AC-IEC-001 | REQ-IEC-001 | every in-scope citation carries a non-resolution marker | MUST | **YES** (4 files listed) |
| AC-IEC-002 | REQ-IEC-002 | the false resolvability assertion is gone, the location statement survives | MUST | **YES** (count 1) |
| AC-IEC-003 | REQ-IEC-003 | `mcp_glm.go` figures survive verbatim | MUST | no — regression guard (see note) |
| AC-IEC-004 | REQ-IEC-004 | path-dependent origins preserved, not deleted | MUST | no — regression guard |
| AC-IEC-005 | REQ-IEC-005 | the glob names one file, or records the inability | MUST | **YES** (glob=1, marker=0) |
| AC-IEC-006 | REQ-IEC-006 | the 12 carve-out lines are byte-unchanged | MUST | no — measured negative |
| AC-IEC-007 | REQ-IEC-008 | every t375-owned file is untouched | MUST | no — measured negative |
| AC-IEC-010 | REQ-IEC-007 | the evidence M5 actually wrote is tracked | MUST | **YES** (10 named evidence files absent; `exit=1`) |
| AC-IEC-011 | REQ-IEC-010 | touched Go packages still build and pass | MUST | no — regression guard |
| AC-IEC-012 | REQ-IEC-009 | no citation names a line number without its tree | MUST | **YES** (stale `:284` present) |

Ten MUST criteria. Five are RED today and go green only through M2-M5 work; four are regression
guards or measured negatives (they go RED **if** the repair breaks something — the counter-example
class the auditor confirmed sound for AC-IEC-006/007); AC-IEC-010 is RED until M5 writes evidence.

The two former MUST criteria that asserted properties of the plan-phase artifacts themselves
(iter1 AC-IEC-008 and AC-IEC-009) are demoted to §D — see §D for why and for their commands.

---

## §C. MUST criteria

### AC-IEC-001 — every in-scope citation carries a non-resolution marker (REQ-IEC-001)

**Given** the in-scope files that **still contain a citation**,
**When** those of them lacking a non-resolution marker are listed,
**Then** the list is empty.

The obligation is conditional on a citation still being present. `grep -l` selects the files that
still carry a citation; `grep -L` then reports which of *those* lack a marker. The pipe composes them
into one binary decision — the violation set is exactly "has a citation **and** lacks a marker".

```bash
grep -l '\.moai/state/verify' internal/cli/mcp_glm.go internal/cli/audit_pin_live_test.go internal/hook/evidence_writer_zeroexec_test.go .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt .moai/reports/template-skill-improvement-plan-20260710.html | xargs grep -LiE 'gitignored|does not resolve|not exported|machine-local|scratch'
```

Observed today (2026-08-31, tree `3f03d9c36`) — **RED**, four files carry a citation with no marker:

```
internal/cli/audit_pin_live_test.go
internal/cli/mcp_glm.go
internal/hook/evidence_writer_zeroexec_test.go
.moai/reports/template-skill-improvement-plan-20260710.html
```

PASS when the output is **empty**.

> **iter3 correction (N1).** iter2 used a bare `grep -L` over all five files, which required **every**
> file to contain a marker word. That is unsatisfiable under the treatment `spec.md` §C.3 actually
> selects: treatment **(a)** for `internal/cli/mcp_glm.go` deletes the path and adds no marker —
> measured, `grep -ciE '<markers>' internal/cli/mcp_glm.go` → `0` — so the file stayed in the `-L`
> output after a *correct* repair and the criterion could never print empty. iter2's stated resolution
> ("deleting the citation removes the obligation") was **false about what the command does**: `grep -L`
> cannot see whether a citation is present, so deletion removed nothing from its input. As written it
> forced treatment (b) onto `mcp_glm.go`, contradicting §C.3, and `plan.md` gates four of the five
> repairs on it. The `grep -l | xargs grep -L` form above makes the obligation conditional, so
> treatment (a) reaches PASS by dropping out of the input set — which is what the prose always claimed.

Edge cases verified in this tree before adoption:

- **All-pass**: `grep -l … e2e-lint-4paths.extract.txt | xargs grep -LiE …` → empty output, `exit=0`.
- **Empty input**: if no file matched `grep -l`, `xargs` receives nothing. Measured on this platform
  it invokes nothing and exits 0 with no output (no hang, no `(standard input)` false positive). This
  case is unreachable in practice regardless: `e2e-lint-4paths.extract.txt` retains its citation under
  its §C.3 treatment, so the input set is never empty.

`mcp_glm.go` leaving the input set is guarded independently: AC-IEC-003 requires its five figures to
survive, so treatment (a) cannot quietly take the claim with the path.

Companion — the citation count must not silently grow:

```bash
git grep -c '\.moai/state/verify' -- internal/cli/mcp_glm.go internal/cli/audit_pin_live_test.go internal/hook/evidence_writer_zeroexec_test.go .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt .moai/reports/template-skill-improvement-plan-20260710.html
```

Observed today: `1` for each of the five files. PASS when each printed count is `≤ 1` (a file
repaired by treatment (a) drops out of the listing entirely).

### AC-IEC-002 — the false resolvability assertion is gone (REQ-IEC-002)

**Given** `internal/cli/audit_pin_live_test.go`,
**When** the assertion and the output-location statement are counted separately,
**Then** the assertion is absent and the location statement survives.

```bash
grep -c 'still resolve at audit time' internal/cli/audit_pin_live_test.go
grep -c 'Evidence lands in' internal/cli/audit_pin_live_test.go
```

Observed today: `1` and `1` — **RED** on the first.

PASS when the first prints `0` **and** the second prints `1`. Both printing `0` is FAIL: it means
the whole sentence was deleted (treatment (a)) where treatment (d) was chosen.

### AC-IEC-003 — `mcp_glm.go` figures survive verbatim (REQ-IEC-003)

**Given** the repaired `internal/cli/mcp_glm.go`,
**When** the distinct figures the claim rests on are counted,
**Then** all five are still present.

```bash
grep -oE '3667|3480|3072|1024|1\.02' internal/cli/mcp_glm.go | sort -u | wc -l
```

Observed today: `5`. PASS when it prints `5`. This is a regression guard: it goes RED precisely if
the M2 edit deletes a figure along with the path.

### AC-IEC-004 — path-dependent origins preserved (REQ-IEC-004)

**Given** the two files whose values are path-dependent,
**When** any surviving origin pointer is counted,
**Then** each file still carries at least one.

```bash
grep -cE '\.moai/reports/|\.moai/state/verify' internal/hook/evidence_writer_zeroexec_test.go .moai/reports/template-skill-improvement-plan-20260710.html
```

Observed today: `evidence_writer_zeroexec_test.go:1`, `template-skill-improvement-plan-20260710.html:1`.

> **iter3 correction (N5)**: iter2 recorded `…html:2`. Re-measured, the value is **1** — `grep -c`
> counts matching *lines*, and both pattern matches sit on line 684. The gate is unaffected (PASS
> needs `≥1`), but the recorded figure was wrong and §A promises every today-value is measured here.

PASS when each printed count is `≥ 1`. A `0` means the pointer was deleted outright — FAIL under
REQ-IEC-004. Goes RED if M3 over-applies treatment (a).

### AC-IEC-005 — single-file naming, or a recorded inability (REQ-IEC-005)

**Given** `.moai/reports/template-skill-improvement-plan-20260710.html`,
**When** the surviving glob and the inability marker are counted,
**Then** either the glob is gone, or it is accompanied by an explicit inability marker.

```bash
grep -c 'skill-audit/\*' .moai/reports/template-skill-improvement-plan-20260710.html
grep -ciE 'not exported|반출되지|cannot be named|식별 불가' .moai/reports/template-skill-improvement-plan-20260710.html
```

Observed today: `1` and `0` — **RED**.

PASS when the first prints `0`, or when the first is `≥1` **and** the second is `≥1`. First `≥1`
with second `0` is FAIL — the glob was silently retained.

### AC-IEC-006 — the 12 carve-out lines are byte-unchanged (REQ-IEC-006)

**Given** the eight carve-out files,
**When** their diff against the baseline is taken,
**Then** it is empty.

```bash
git diff --exit-code --stat origin/develop...HEAD -- internal/verify/store.go internal/verify/schema.go internal/verify/store_test.go internal/web/events.go internal/goal/evaluate_snapshot_test.go internal/session/ignored_content_test.go .moai/reports/moai-autonomy-workflow-redesign-20260803.html .moai/reports/model-tier-redesign-20260712.html
echo "exit=$?"
```

Observed today: empty output, `exit=0`. PASS when `exit=0` and output is empty. `--exit-code` makes
this decide by exit code rather than by reading a diff.

Paired positive control — "unchanged" must not be satisfied by deleting the files:

```bash
git grep -c '\.moai/state/verify' -- internal/verify internal/web/events.go internal/goal internal/session .moai/reports/moai-autonomy-workflow-redesign-20260803.html .moai/reports/model-tier-redesign-20260712.html
```

Observed today: `store.go:1`, `schema.go:1`, `store_test.go:3`, `events.go:1`,
`evaluate_snapshot_test.go:3`, `ignored_content_test.go:1`, and `1` for each `.html` — summing to
**12**. PASS when the eight printed counts still sum to 12.

### AC-IEC-007 — every t375-owned file is untouched (REQ-IEC-008)

**Given** the t375-owned file list of spec.md §C.5 (path-corrected per iter2 D5),
**When** their diff against the baseline is taken and the guard file's absence is checked,
**Then** the diff is empty and the guard file does not exist.

```bash
git diff --exit-code --stat origin/develop...HEAD -- .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/agent-common-protocol-reference.md .claude/agents/moai/manager-lead.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.codex/agents/moai/manager-lead.toml
echo "exit=$?"
```

Observed today: empty output, `exit=0`.

```bash
ls internal/template/evidence_citation_guard_test.go
```

Observed today: `No such file or directory`, `exit=1` (BSD `ls`; the PASS condition is the file being absent, not a numeric rc).

PASS when the diff exits 0 with empty output **and** the `ls` reports the file absent.

> **iter2 D5 correction**: iter1 diffed the top-level `.codex/agents/moai/manager-lead.toml`, which
> does not exist in this tree — `git ls-files '*manager-lead.toml'` returns exactly one path,
> `internal/template/templates/.codex/agents/moai/manager-lead.toml`. Diffing a non-existent path
> exits 0 with empty output and therefore asserted nothing. The path is corrected above and in
> spec.md §C.5.

### AC-IEC-010 — the evidence M5 actually wrote is tracked, not ignored (REQ-IEC-007)

**Given** the evidence directory after M5 has run,
**When** the per-criterion evidence files are named and the ignore status is checked,
**Then** every named file exists, the directory is tracked, and nothing was written under
`.moai/state/`.

```bash
ls .moai/reports/t381/verify/ac-iec-001.txt .moai/reports/t381/verify/ac-iec-002.txt .moai/reports/t381/verify/ac-iec-003.txt .moai/reports/t381/verify/ac-iec-004.txt .moai/reports/t381/verify/ac-iec-005.txt .moai/reports/t381/verify/ac-iec-006.txt .moai/reports/t381/verify/ac-iec-007.txt .moai/reports/t381/verify/ac-iec-010.txt .moai/reports/t381/verify/ac-iec-011.txt .moai/reports/t381/verify/ac-iec-012.txt
echo "exit=$?"
```

```bash
git check-ignore -v .moai/reports/t381/verify
echo "exit=$?"
```

```bash
ls .moai/state/verify
```

PASS when: the named-file `ls` exits **0** (all ten evidence files exist — one per MUST criterion);
`git check-ignore` exits **1** (the path is NOT ignored); and `ls .moai/state/verify` reports
`No such file or directory`, confirming the run wrote nothing there.

Observed today (2026-08-31) — **RED** on the first check:

```
ls: .moai/reports/t381/verify/ac-iec-001.txt: No such file or directory
…
exit=1
```

`git check-ignore` → `exit=1` (not ignored, correct). `ls .moai/state/verify` → `No such file or
directory`. The directory currently holds three files — `spec-lint-full.txt`,
`spec-lint-targeted.txt`, `spec-lint-verification.md` — none of which is a per-criterion evidence
file, so the naming check is genuinely RED and goes green only once M5 writes all ten.

> **iter3 correction (N2).** iter2 recorded this criterion as "RED until M5" and its matrix cell as
> "(dir not yet populated)". Both were false: all three sub-checks passed before M5, and the directory
> already held three files. The deeper defect was in the **deciding command**, not the wording —
> `git status --short` collapses an untracked directory to a single `??` line regardless of contents,
> so it could not distinguish twelve fresh evidence files from one stale one, and correcting the prose
> alone would have left the criterion unable to test what iter2's re-scope claimed it tested. Naming
> the ten expected files is what makes it discriminate, and what makes it RED today.

### AC-IEC-011 — no behavior change in the touched Go packages (REQ-IEC-010)

**Given** the two Go packages containing in-scope edits,
**When** they are built and tested,
**Then** both succeed.

```bash
go build ./internal/cli/... ./internal/hook/...
echo "exit=$?"
```

```bash
go test ./internal/cli/... ./internal/hook/...
```

PASS when the build exits `0` and the test run reports no `FAIL` line for these two packages.
`internal/cli` is slow — allow a 600s timeout floor. Full-suite judgment belongs to CI, not to this
lane (`CLAUDE.local.md` §4).

### AC-IEC-012 — no citation names a line number without its tree (REQ-IEC-009)

**Given** `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt`,
**When** the stale coordinate and the surviving non-resolution statement are counted,
**Then** the stale coordinate is gone and the statement remains.

```bash
grep -c 'gitignore:284' .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt
grep -ci 'gitignore' .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt
```

Observed today: `1` and `1` — **RED** on the first.

PASS when the first prints `0` and the second prints `≥1`.

> **iter2 D6 correction**: iter1 backed this criterion with a document section (`§C.3`) rather than a
> requirement, and REQ-IEC-001 does not reach this file — its header already declares non-resolution,
> so it satisfies AC-IEC-001 today, repaired or not (measured: `grep -ci 'gitignore'` = 1). The new
> **REQ-IEC-009** (spec.md §B) binds stale coordinates in the Unwanted form and is what this
> criterion now gates.

---

## §D. Structural checks (not MUST — recorded, not gating)

These assert properties of the plan-phase artifacts. They are true on arrival, and **no milestone
M1-M5 can falsify them**, because no milestone edits `spec.md`. A criterion that cannot go RED at any
point in the run measures nothing, so they are recorded here rather than counted as gates (iter2,
D3). They are retained because they document what the SPEC must continue to carry across future
revisions.

**S-1 — the probe and its blind spots are pinned into spec.md** (was AC-IEC-008):

```bash
grep -c "git grep -n '..moai/state/verify' -- . ':!\*.md'" .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
grep -cE '^\| B[1-6] \|' .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
```

Expected `≥1` and `6`.

**S-2 — the C4 exclusion is stated with its reason** (was AC-IEC-009):

```bash
grep -c '^### Out of Scope' .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
grep -c 'class C4' .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
grep -c 're-produced' .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
```

Expected `5`, `≥1`, `≥1`.

---

## §E. Edge cases

- **The baseline moves — and that is now handled, not merely noticed.** The iter3 wording of this
  edge case keyed on `HEAD ≠ 3f03d9c36` and prescribed re-attributing "to the actual merge base". That
  prescription was circular: the merge base of an absorbed tree and `3f03d9c36` **is** `3f03d9c36`, so
  it returned the same failing comparison. The three-dot baseline (§A) resolves the merge base at
  verification time, which is what actually closes it. **Absorbing `origin/develop` does not
  invalidate AC-IEC-006 or AC-IEC-007** — after the absorb the merge base becomes `origin/develop`,
  and both criteria still see only this card's edits.
- **Treatment (a) vs a deleted sentence.** AC-IEC-002's two-part check separates "the assertion was
  corrected" from "the sentence was removed". A single count cannot tell them apart.
- **Treatment (a) vs a lost figure.** AC-IEC-001 goes green for `mcp_glm.go` when the citation is
  deleted; AC-IEC-003 is what stops that deletion from taking the figures with it. The two must be
  read together.
- **Pattern breadth.** Any criterion widened from `.moai/state/verify` to `.moai/state/` will sweep
  `template-skill-improvement-plan-20260710.html:529` (`loop-verdict`, census B2) into scope and
  demand a repair the SPEC forbids.
- **A carve-out file edited for an unrelated reason.** AC-IEC-006 fails on any diff, including a
  benign one. Intended: this lane has no reason to touch those files at all.
- **t375 lands first and changes its wording.** spec.md §D requires re-reading §A.4 rather than
  assuming. No criterion here asserts a REQ-ECC line number as a pass condition, so a renumbering in
  t375 does not silently break this gate.

---

## §F. Definition of Done

- All ten MUST criteria PASS, each with its command's verbatim output persisted under
  `.moai/reports/t381/verify/`.
- Every command was executed **in this worktree** and observed to run — not merely written.
- The five in-scope files carry the treatment selected in spec.md §C.3, and no other file in the
  repository is modified except this SPEC directory and the evidence directory.
- The open question of spec.md §A.4 (does REQ-ECC-004 reach an output-location statement?) has been
  answered by the lead, or is recorded in progress.md as an unresolved carry-forward with the
  treatment chosen under the conservative reading.
- Residual risk recorded: the C4 instruction is untouched, so the same citations can be re-produced.
