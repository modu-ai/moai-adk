# progress.md — SPEC-AGENTS-WORKTREE-ROW-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_auditor: pending (independent audit is the next act after this commit; not run by this
  session — the lead dispatches plan-audit separately)

### Plan-phase evidence (plan-phase, this run, worktree t1071 @ `cd99336bf`, 2026-09-22)

Re-measurement of the card's investigator citations (spec.md §A, C1–C8), executed before any
artifact was written:

- **C1** `AGENTS.md:18-25` read — capability table, three rows, no worktree row. Negative
  control `grep -c '^| worktree-entry |' AGENTS.md` → `0`; positive control
  `grep -c '^| question-channel |' AGENTS.md` → `1`.
- **C2** `AGENTS.md:106-111` read — §3 launcher enumeration lists `moai cc -w`, `--spawn`,
  `EnterWorktree(<path>)` only; the `moai codex` form is absent.
- **C3** `internal/template/templates/AGENTS.md.tmpl:18-29` read — three-column table with
  **seven rows** (`question-channel`/`task-list`/`design-sync` plus `agent-spawning`,
  `output-style`, `slash-commands`, `workflow-scripts` — four rows the root copy lacks, an
  intentional-fork divergence), same absence of a worktree row. Negative control grep → `0`.
  (Corrected from the original "same three rows at `:18-25`" at plan-audit iter-1 — D1; the
  full-row enumeration was re-run and confirmed by this author before the correction landed.)
- **C4** `internal/template/templates/AGENTS.md.tmpl:291-305` read — `## 11. moai CLI Verbs`
  at `:291`, nine verb rows `:295-303`, no `moai codex` row. Negative control
  `grep -c '^| \`moai codex\`' …` → `0`; positive control `grep -c '^| \`moai init' …` → `1`.
  The dispatch's `:293-302` citation is corrected to `:293-303` (re-measured range governs).
- **C5** `internal/cli/codex_launcher.go` read — verb registration `:324` (`Use:` line, code
  declaration), `DisableFlagParsing: true` `:360`, `resolveCodexWorktreeDir` `:272-299` (L2
  validation `:279`, L1 join `:286`, absent-directory diagnostic `:289-297`), named diagnostics
  `:47`/`:52`, help text `:341-342`. Entry exists; creation is impossible.
- **C6** The dispatch's `:16-17` citation is a package-doc comment block (`codex_launcher.go:16-19`),
  not a declaration — the card's own comment-hit warning applied to the citation itself.
  Declaration anchors above are code.
- **C7** `grep -n 'moai CLI Verbs' AGENTS.md` → no matches — the root copy carries no Verbs
  table; the registration fix is mirror-only.
- **C8** `internal/template/templates/AGENTS.md.tmpl:270` prose-references `moai codex` while
  the Verbs table omits it — the command is live; only the inventory lags.

Pre-edit control batch (one compound invocation, 2026-09-22):

```
G1-root-worktree-row: 0
G2-tmpl-worktree-row: 0
G3-tmpl-codex-verb: 0
G4-positive-control-root-table: 1
G5-positive-control-verbs-table: 1
exit=0
```

Interpretation: the three zeros are evidence of absence (both positive controls fired on the
same run), and the two ones prove the row-shape patterns fire against the tables as they stand —
so the post-edit guards measuring `1` will be a real flip, not a pattern that matches nothing.

Ambiguity decisions taken (for the plan-auditor):

1. The dispatch named the root capability table and the mirror Verbs table explicitly, but
   re-measurement found the mirror's *capability table* also lacks the row (C3). Constraint (c)
   was read as licensing the fix; it is REQ-AWR-002. Quotation-of-authority note (added at
   plan-audit iter-1, D6): the lead's plan dispatch (2026-09-22) words constraint (c) as "Judge
   each copy separately and fix each accordingly", while the audit reads a different dispatch
   wording from the card body — "judge each copy separately, and template edits require `make
   build` regeneration". The two wordings diverge on the tail clause; both are retained here
   verbatim rather than adjudicated by this session. The licensing conclusion stands on the
   file-division grant either way (audit adjudication 2).
2. The dispatch's `:16-17` coordinate was replaced by code anchors (`:324`, `:272-299`) per the
   card's own comment-hit rule.
3. `tier: M` was chosen over `S` because the card requires `acceptance.md` explicitly and the
   verification chain (guards + build + lint) exceeds a two-artifact scope.

SPEC lint at plan-phase close (tree build `go run ./cmd/moai spec lint
SPEC-AGENTS-WORKTREE-ROW-001`, this run, this tree): exit 0, `0 error(s), 4 warning(s)` —
REQ collection FIRED (four REQ-AWR-* rows collected and modality-judged; REQ-AWR-004 emitted no
finding). The four findings are `ModalityMalformed` on REQ-AWR-001/002/003/005 — the same
prose-shape class card t1057 measured; they are warnings, not errors, and are left for the
plan-auditor to weigh rather than re-worded mid-flight by this plan phase.

### Plan-audit iter-1 — PASS-WITH-DEBT 0.875, rework applied (2026-09-22)

Verdict: PASS-WITH-DEBT, overall 0.875 (Tier M threshold 0.80), BLOCKING 0. Report:
`.moai/reports/t1071/plan-audit-iter1.md`. The audit's five adjudications of this author's
flagged decisions all resolved favorably (lint warnings absorbed as documented debt D9;
REQ-AWR-002 in-scope; Tier M sanctioned; the C4 `:293-303` correction verified; the C6
comment-to-code anchor substitution verified).

Rework applied this session, all in the SPEC's own four artifacts (no contract file touched,
`make build` not run — nothing under `templates/` has changed):

| Defect | Repair |
|---|---|
| D1 (SHOULD-FIX) | C3 rewritten — mirror table is **seven rows at `:21-29`**, four of which the root lacks; corrected in spec.md §A C3, REQ-AWR-002 (`:21-29`, append after last row), plan.md §A.1 item 2, plan.md M2(a) (`:29`, not the mid-table `:25`). Author's own re-enumeration of the mirror rows confirmed the auditor's count before the fix landed |
| D2 (SHOULD-FIX) | AC-AWR-004 rewritten with explicit adoption cells — green-now by construction, red constructible at run-phase (weakened-filter mutant observing `2`); AC-AWR-005 reclassified **regression-guard** per verification-completeness §2.1 (red unconstructible — no build-chain step reads `AGENTS.md.tmpl` drift, Makefile:34 measured by the audit), "proving" clause dropped |
| D3 (SHOULD-FIX) | AC-AWR-003's Then now names **both** verbs — the launch verb (`cli`) and the readout verb (`status`) — closing the missing-`cli` mutant |
| D4 (MINOR) | DoD cross-reference `§E.2` → `§E.1` |
| D5 (MINOR) | AC-AWR-004 spells both copies' full fence-file paths |
| D6 (MINOR) | The constraint-(c) quotation conflict is recorded verbatim in both wordings (lead's dispatch vs card body) — not adjudicated here; stands on the file-division grant either way |
| D7 (MINOR) | REQ-AWR-005 and AC-AWR-005 reworded — `.tmpl` is a compile-time embed the generating steps do not process; the discharge is the recompile; the make-output-dependent no-op clause is dropped |
| D8 (MINOR) | AC-AWR-001/003 "unchanged rows" claims restated as **presence counts** with the content-identity limit named (carried by plan.md §G at diff review) |
| D9 (DEBT) | The 4 lint warnings — absorbed per the audit's adjudication 1; no rewording |

## §E.2 Run-phase Evidence

Run-phase executed 2026-09-22, worktree t1071, branch `WT-cross-harness-row`, serial mode per
§F. Pre-run absorb: HEAD at dispatch was `8fe51ace9` — a merge commit absorbing local develop
`00e761af8` (t1085/t1081 advance). **Scope-claim base resolution** (gitflow-lane-protocol §8,
t543 precedent — the card's own contribution is measured from the absorbed ref's merge-base,
never a literal plan-time pin): `git merge-base develop HEAD` → `00e761af894975602a76abf1d856c186ee546751`.
All AC-AWR-004 measurements below use this resolved base. Every item below carries
command + verbatim output + tree SHA + baseline-attribution (this run, this tree).

### M1+M2 commit (B4 status flip carried)

- **Commit 1**: `0cf533f2a` — subject `fix(SPEC-AGENTS-WORKTREE-ROW-001): M1+M2 register
  worktree-entry rows and codex verb row (card t1071)`; trailers contiguous in order:
  `Authored-By-Agent: manager-develop` / `Card: t1071` / `🗿 MoAI`.
- Files: `AGENTS.md` (+1 row), `internal/template/templates/AGENTS.md.tmpl` (+2 rows),
  `.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/spec.md` (frontmatter `status: draft` →
  `status: in-progress` — the only transition this phase owns; `updated:` already
  2026-09-22, unchanged).
- Diff shape at commit time: `3 files changed, 4 insertions(+), 1 deletion(-)` — one added
  row per capability table, one added verb row, zero pre-existing rows reworded (plan.md §G).

### E1 — AC matrix (acceptance.md §D, tree `0cf533f2a`, this run)

| AC | Status | Verification command | Actual output (verbatim) |
|----|--------|----------------------|--------------------------|
| AC-AWR-001 | PASS | `grep -c '^| worktree-entry \|' AGENTS.md` | `1` (exit 0); row line contains `moai codex -w` and `never creates one`, Codex form inside column 2 (diff-verified); pre-existing rows `question-channel`=`1`, `task-list`=`1`, `design-sync`=`1` |
| AC-AWR-002 | PASS | `grep -c '^| worktree-entry \|' internal/template/templates/AGENTS.md.tmpl` | `1`; row appended after the table's last row (`workflow-scripts`) |
| AC-AWR-003 | PASS | `grep -c '^| \\\`moai codex\\\`' internal/template/templates/AGENTS.md.tmpl` | `1`; row names `cli` + `status`, carries `never creates`, states `-w <worktree>`; nine pre-existing verb rows present (`moai init`=`1`); post-edit verb-table total `grep -c '^\| \\\`moai '` = `10` |
| AC-AWR-004 | PASS | scope diff at resolved base (red-then-green below) | weakened filter `2` (red observed once), strict filter `0`; 4 fence paths `0` commits in range; `codex_launcher.go` `0` diffs |
| AC-AWR-005 | PASS (process-gate discharge) | `make build` | exit `0`; ldflags line carries `Commit=0cf533f2a` (recompile against the committed template tree). Regression-guard classification per acceptance.md — no proof-pass claimed |

### E2 — guard triple + positive controls (tree `0cf533f2a` content, this run)

```
G1 '^| worktree-entry |' AGENTS.md                                   → 1
G2 '^| worktree-entry |' internal/template/templates/AGENTS.md.tmpl  → 1
G3 '^| `moai codex`' internal/template/templates/AGENTS.md.tmpl      → 1
G4 '^| question-channel |' AGENTS.md (positive control)              → 1
G5 '^| `moai init' internal/template/templates/AGENTS.md.tmpl        → 1
```

Pre-edit baseline (this run, tree `8fe51ace9`): G1–G3 = `0`/`0`/`0`, G4–G5 = `1`/`1` — a real
0→1 flip on all three guards, not a pattern matching nothing.

### AC-AWR-004 — reserved red observation (base `00e761af8`, HEAD `0cf533f2a`)

Weakened filter (red this instrument must show once):

```
$ git diff --name-only 00e761af8..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/' | wc -l
       2
$ git diff --name-only 00e761af8..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/'
AGENTS.md
internal/template/templates/AGENTS.md.tmpl
```

Strict form (green):

```
$ git diff --name-only 00e761af8..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/' | grep -v '^AGENTS\.md$' | grep -v '^internal/template/templates/AGENTS\.md\.tmpl$' | wc -l
       0
```

Fence files (t1072) — both copies, `git log --oneline 00e761af8..HEAD -- <path> | wc -l`:

```
.claude/rules/moai/workflow/worktree-integration.md                        → 0
.claude/rules/moai/workflow/session-handoff-examples.md                    → 0
internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md   → 0
internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md → 0
```

Launcher frozen: `git diff --name-only 00e761af8..HEAD -- internal/cli/codex_launcher.go | wc -l` → `0`.

### E3 — `make build` (exit code + tail; full log `/tmp/t1071-make-build.log`, machine-local)

```
make-build-exit=0
... tail:
catalog.yaml updated successfully (13408 bytes)
go build -ldflags "-s -w -X github.com/modu-ai/moai-adk/pkg/version.Version=moai_cp/20260910_130400 -X github.com/modu-ai/moai-adk/pkg/version.Commit=0cf533f2a -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-09-22T13:09:17Z -X github.com/modu-ai/moai-adk/pkg/version.BuildID=moai_cp/20260910_130400-2413-g0cf533f2a" -o bin/moai ./cmd/moai
```

Post-build residue check: `git status --porcelain` → empty (the `gen-catalog-hashes` step
rewrote `catalog.yaml` byte-identically). REQ-AWR-005 discharged by the recompile itself;
`agents-emit-check`/`commands-emit-check` ran read-only inside the chain without aborting.

### E4 — `go run ./cmd/moai spec lint SPEC-AGENTS-WORKTREE-ROW-001` (verbatim, exit `0`)

```
SEVERITY  CODE               FILE  LINE  MESSAGE
WARNING   ModalityMalformed  .../spec.md  120  REQ REQ-AWR-001: EARS modality violation — SHALL missing or format mismatch: "The root `AGENTS.md` capability-binding table (`AGENTS.md:21-25`)"
WARNING   ModalityMalformed  .../spec.md  126  REQ REQ-AWR-002: EARS modality violation — SHALL missing or format mismatch: "The template mirror's capability-binding table"
WARNING   ModalityMalformed  .../spec.md  133  REQ REQ-AWR-003: EARS modality violation — SHALL missing or format mismatch: "The template mirror's `## 11. moai CLI Verbs` table"
WARNING   ModalityMalformed  .../spec.md  144  REQ REQ-AWR-005: EARS modality violation — SHALL missing or format mismatch: "When any file under `internal/template/templates/` is edited,"

0 error(s), 4 warning(s)
lint-exit=0
```

(Full path column elided to `.../spec.md` here only for width; the run printed the absolute
worktree path on every row.) REQ collection FIRED — exactly the 4 known `ModalityMalformed`
warnings on REQ-AWR-001/002/003/005 (D9 adjudicated debt, t1057-class parser narrowness), zero
errors, zero new findings. No requirement prose was reworded to chase them.

### E5/E6 — measured at tree `0cf533f2a`

- E5 `git status --porcelain` → empty (clean; only gitignored residue possible).
- E6 commit-1 stat, `git show --stat 0cf533f2a` → `AGENTS.md | 1 +`,
  `internal/template/templates/AGENTS.md.tmpl | 2 ++`,
  `.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/spec.md | 2 +-` — `3 files changed, 4 insertions(+), 1 deletion(-)`.
  (F1 re-citation at sync close: this figure's producer is the per-commit form above — the
  run-time record mis-cited it as the range command's output. The range form re-measured at the
  close tree, `git diff --stat 00e761af8..HEAD`, prints `7 files changed, 721 insertions(+)`
  (acceptance.md +115, plan.md +114, progress.md +282, spec.md +206, AGENTS.md +1, CHANGELOG.md +1,
  AGENTS.md.tmpl +2) — the branch range additionally carries the plan-phase artifacts, this
  evidence file, and the stage-1 CHANGELOG entry. Verbatim disposition in §E.4
  `audit_f1_disposition`.)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: pending-backfill-run   # D3 placeholder — commit 2 cannot cite its own SHA
run_status: complete
ac_pass_count: 5
ac_fail_count: 0
preserve_list_post_run_count: 0   # violations; 4 fence paths + launcher verified untouched (§E.2)
l44_pre_commit_fetch: not-run (worktree-local lane run; base re-resolved via merge-base at dispatch and HEAD re-read immediately before each commit)
l44_post_push_fetch: not-applicable (lane does not push; lead batch-pushes develop per gitflow-lane-protocol §4)
new_warnings_or_lints_introduced: 0   # the 4 ModalityMalformed warnings are pre-existing plan-phase debt (D9), unchanged
cross_platform_build:
  darwin_make_build: "exit 0 (Commit=0cf533f2a)"
  windows_matrix: not-run-locally (docs-only change, zero Go source touched; remote CI on origin/develop owns the matrix)
total_run_phase_files: 4   # commit 1: AGENTS.md + AGENTS.md.tmpl + spec.md; commit 2: progress.md
m1_to_mN_commit_strategy: 2 commits — commit 1 carries M1+M2 contract rows + B4 status flip; commit 2 carries M3+M4 evidence (this section)
```

## §E.4 Sync-phase Audit-Ready Signal

Close signal (2026-09-22, worktree t1071, branch `WT-cross-harness-row`). Independent sync-audit
verdict: **PASS-WITH-DEBT 96/100** — Functionality 96 / Security 100 / Craft 90 / Consistency 97
(report: `.moai/reports/t1071/sync-audit.md`). Sync stage 1 changed exactly two files: this
`progress.md` §E.4 draft and one new `CHANGELOG.md` `[Unreleased]` entry. This close commit adds
the `spec.md` frontmatter terminal transition (`in-progress → implemented → completed`, merged
per the 3-phase close; `updated:` already 2026-09-22, unchanged; body untouched), this §E.4
finalization, and the audit-F1 §E.2 re-citation.

**CHANGELOG decision: ENTRY** under `[Unreleased]` → `### Added` (first entry). Rationale: the
card's substantive change is the **distributed** template mirror
(`internal/template/templates/AGENTS.md.tmpl`), which ships to user projects via `moai update` —
the AGENTS.md contract change is user-visible, so it is Keep-a-Changelog material; this differs
from the t1081 no-entry precedent (docs-only, `internal/web`, a developer-local surface). The
mirror-only `moai codex` Verbs row is inventoried in the same entry.

**MX scan disposition: no-op this sync.** The run diff is docs-only — zero Go source across all
commits (strict scope filter 0 at merge-base `00e761af8`; the only non-SPEC paths are two
markdown contract files). MX Tag validation targets code artifacts; no `@MX:*` annotation surface
exists in this card's change set, so nothing to validate, add, or update.

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: pending-backfill-sync   # D3 placeholder — this close commit cannot cite its own SHA; backfilled in the immediately following commit
sync_status: complete — independent sync-audit PASS-WITH-DEBT 96/100 (Functionality 96 / Security 100 / Craft 90 / Consistency 97; report .moai/reports/t1071/sync-audit.md)
b12_self_test_a: PASS — grep -c 'SPEC-AGENTS-WORKTREE-ROW-001' CHANGELOG.md → 0 pre-emission (no duplicate-entry risk)
b12_self_test_b: PASS — acceptance.md distinct AC identifiers = 5 (AC-AWR-001..005); CHANGELOG entry references the same 5
b12_self_test_c: PASS — every path named in the entry verified to exist — AGENTS.md, internal/template/templates/AGENTS.md.tmpl (run commit 0cf533f2a diff), .moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/spec.md (ls)
changelog_entry_position: "[Unreleased] → ### Added, first entry"
frontmatter_status_transitions: carried by this close commit — in-progress → implemented → completed merged into the single sync commit per the 3-phase close (status only; updated already 2026-09-22; body untouched)
mx_scan: no-op — docs-only card, zero Go source in the run diff; no @MX annotation surface in the change set
canary_compliance_check: n/a — this SPEC defines no forward-looking policy exercised by its own sync tests
audit_f1_disposition: §E.2 E6 command↔output pairing corrected — the `3 files changed, 4 insertions(+), 1 deletion(-)` figure re-attributed to its true producer `git show --stat 0cf533f2a` (re-observed at close, identical output), and the range form re-measured at this close tree (`git diff --stat 00e761af8..HEAD`) → `7 files changed, 721 insertions(+)` (acceptance.md +115 / plan.md +114 / progress.md +282 / spec.md +206 / AGENTS.md +1 / CHANGELOG.md +1 / AGENTS.md.tmpl +2; 0 deletions). Applied under explicit lead dispatch citing the auditor's blocking-at-close prescription
audit_f2_scope_observation: strict scope filter at close tree (`git diff --name-only 00e761af8..HEAD` filtered by the SPEC dir, `^AGENTS\.md$`, and the template path) → exactly `CHANGELOG.md` (1 path) — the sanctioned stage-1 CHANGELOG entry; the card's only non-SPEC code-surface changes remain the two contract files
```

## §F Phase 4 Mode Selection

Decision: serial

- Inputs: tier M · scope 2 contract files (AGENTS.md + AGENTS.md.tmpl) + SPEC artifact updates · domains 1 (documents only, zero code) · file language 100% markdown · concurrency benefit LOW (ordered M1→M2→M3 chain; M3's regeneration depends on M2's edits — no inter-file parallelism) · agent-team prereqs not requested
- Mode evaluation: direct — not selected (two contract files + guard execution + make build exceed a trivial single-line change); fanout — not selected (single domain, nothing parallelizable); sweep — not selected (~2 files, semantic doc edits, not mechanical bulk); agent-team — not selected (no explicit operator request)
- Justification: a single manager-develop (dev-swguard) spawn per milestone is the whole envelope; the coding-task-parallelism caveat and the single-domain scope both point at serial as the simpler sufficient mode
- Kickoff status: Implementation Kickoff Approval granted by the operator 2026-09-22 ("전부 승인"), relayed via the lead with the four run constraints to be carried verbatim in the dev-swguard dispatch

---

🗿 MoAI
