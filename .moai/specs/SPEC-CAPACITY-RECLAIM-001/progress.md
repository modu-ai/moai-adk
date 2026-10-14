# SPEC-CAPACITY-RECLAIM-001 — Progress

status: draft
plan_status: audit-ready
plan_complete_at: 2026-10-14T02:03:04Z

## §E.1 Plan-phase Audit-Ready Signal

Plan artifacts (spec.md, plan.md, acceptance.md, decision-index.md) are revised to SPEC v0.2.0 on card worktree HEAD `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`, branch `WT-capacity-reclaim`). This is the FINAL authorized repair round: leader decision d-20261010T080627Z-987a allows exactly one repair inside fix_scope and one final delta re-audit after it; a FAIL on that audit returns the card to the operator.

Basis. The revision answers the iter3 verdict (`.moai/reports/t1631/plan-audit-iter3.md`: FAIL 0.56, 11 blocking, MP-8) restricted to card A's subset of the operator's narrowed scope (decision recorded on the card queue record, 2026-10-10): D1 (AC-CAP-008's approved-branch four-element cell EV-007, command unescaped, verbatim output, exit per the §0 rule, tree pin), D2 (dissolved — its criterion moved to card B with the sweep scope), D3 (AC-CAP-004 restored to RB with a four-element cell; the G-EXIT premise withdrawn), D4 (the §0 exit-capture rule restated from this session's runner observations; the `bash -c` wrapper withdrawn; no "runner-captured" label without a runner-printed status), D7 (plan §D.1 rotation mechanics deleted; retention ownership and the ordering guarantee stated instead), D8 (new AC-CAP-009: the archive-before-rewrite characterization test on the default branch), D9 (new AC-CAP-012: the SessionStart split behavior), D11 (§F absent-key residual rewritten: no notice exists in that case). D5, D6, and D10 left this SPEC by excision, not by rewrite: the worktree-sweep scope (disposal narrowing, last-activity field, report fields, `moai worktree list`, bulk-run evidence, 10 GiB target) moved to card B; the doctor cap read-path (`internal/cli/doctor_disk.go:155`, `:261`) and the compiled-constant removal moved to card D; the inbox rotation moved to card t1653.

Absorb and pins. The card worktree merged local `main` (merge commit `e2658f3012`, 689 commits, zero conflicts, tree clean) and every pin this SPEC carries was re-measured on that tree in this session. Of the measured paths, only `internal/config/defaults.go` (constants moved 451→456, 455→460) and the template `workflow.yaml` (hygiene mode line 318→331) differ from the iter3-audited tree `ab9f5a6d3`; all other cited lines reproduce unchanged. The template `.gitignore` still holds no rule for the state files or archives; the five state files are still tracked, now totalling 132,585,731 B across 191 files.

Scope addition flagged for operator ratification. The merged tree carries a new tracked monthly archive (`learning-history/archive/2026-09.jsonl.gz`, 2,142,019 B; six `.gz` archives now tracked, 4,929,798 B total), which pushed the post-untracking residual to 5,230,544 B — above the operator's 5,000,000-byte completion criterion 1. REQ-CAP-010 (new) adds the archive untracking (local copies kept; template rule `.moai/harness/learning-history/archive/*.jsonl.gz`), restoring the criterion (residual 300,746 B). This widens the M2 untracking set from five to eleven files and was not in the operator's enumerated list; it is forced by the operator's own PASS definition on the merged tree and is recorded in spec §A v0.2.0, §E A14, and §F.

Q3. The clone-measurement gate is satisfied: the per-clone recipe was executed on 2026-10-10 and recorded at `.moai/reports/t1631/q3-clone-measurement.md`; the untracking is authorized by the narrowed-scope decision. B1 lands through the card's normal landing path.

Counts. 10 requirements (REQ-CAP-001–010) and 13 acceptance criteria (AC-CAP-001–013), against the Tier M ceilings of 16 and 16. AC renumbering and the v0.1.2→v0.2.0 mapping are recorded in acceptance.md §0.

v0.2.1 — the authorized one-clause D12 repair (operator decision d-20261014T025128Z-0aa1, kind ceiling-exception: repair only D12, then exactly one re-audit; another FAIL returns the card to hold): acceptance.md AC-CAP-012 clause (1) now seeds one `verify` residue-class target in the fixture project and requires its `"unit":"gc"` / `"mode":"apply"` audit row, making the cache/GC apply half observable (D12, `.moai/reports/t1631/plan-audit-final.md`); premise cells EV-015 (no hook fixture seeds the class) and EV-016 (the GC's decision-gated apply-row write, `gc.go:395–404`) were observed on this tree `111493b71`, EV-008 re-observed there unchanged, and the evidence carriers they cite were updated accordingly. Nothing outside the clause changed: REQ-CAP-010 and every other REQ, AC, and milestone are untouched, no `internal/` code changed, and the REQ-CAP-010 ratification remains with the operator. The re-audit is the next act.

Plan-audit outcome (2026-10-14): the single authorized re-audit PASSED — score 0.86 against the Tier M threshold 0.80, zero blocking defects, MP-1..9 all pass, codex required backend `pass` (receipt `rcpt-3e5c619ef067d45b28756f2f`), D12 mutant closed. Verdict: `.moai/reports/t1631/plan-audit-final-2.md`. Skip-eligibility: plan-artifact hash `8ecff932e63f1cb5c708ba7b2166c52ef1c3a104ca959f80519545d22a76690c` stable at HEAD `e993fff98999e66a549971017b04d3e449725177`. Run entry authorized by decision d-20261014T025128Z-0aa1 ("then exactly one re-audit"; PASS advances the card).

## §E.2 Run-phase Evidence

Run entered on tree `e993fff98999e66a549971017b04d3e449725177` (short `e993fff989`, branch `WT-capacity-reclaim`, clean except pre-existing progress.md plan-phase edit). All commands below ran in this worktree from its root; exit codes are the runner observations described in acceptance.md §0 (this runner prints `Exit code N` for a non-zero bare invocation and `(Bash completed with no output)` for empty output; below, "exit" labelled *convention* means the runner showed no status line).

### Pre-flight measurements (plan §C, run as written)

| Command | Observed output (verbatim) |
|---|---|
| `git rev-parse HEAD` | `e993fff98999e66a549971017b04d3e449725177` |
| `git status --short` | ` M .moai/specs/SPEC-CAPACITY-RECLAIM-001/progress.md` (plan-phase edit, pre-existing) |
| `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` | `132585731 total` |
| `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl .moai/harness/learning-history/archive` | the eleven paths of EV-002, reproduced identically |
| `grep -n -E 'usage-log\|lessons-inbox' internal/hygiene/sinks.go` | `90:` lessons-inbox owner exception; `91:` usage-log owner exception (matches AC-CAP-006/007 green premises) |
| `grep -rn -E 'DefaultProfileMaxBytes' internal` | 5 lines: `defaults.go:460` (const), `doctor_disk_test.go:183`, `clean_home.go:243`, `doctor_disk.go:155`, `doctor_disk.go:261` |
| `grep -n -E 'mode: (report\|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` | `331:        mode: report              # report \| apply (auto path only)` (EV-007 reproduced) |
| `grep -rn 'TestRetentionPruneArchivesBeforeRewrite' internal` | (empty) — EV-006 re-observed, exit 1 by convention, runner showed `(Bash completed with no output)` |
| `grep -rn 'TestSessionStartAutoPathSplit' internal` | (empty) — EV-008 re-observed on `e993fff989` (superset of the `111493b71` pin; no `internal/` path differs), exit 1 by convention |
| `grep -rn 'profile_max_bytes' internal/hook` | (empty) — EV-009 re-observed, exit 1 by convention |

Additional premises re-observed: `grep -n -E 'usage-log\|lessons\|corrupt' internal/template/templates/.gitignore` → (empty), exit 1 by convention (EV-001 reproduced). Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (verbatim tails: `BUILD-MAIN-OK` / `BUILD-WINDOWS-OK` echoed after each).

Affected-family test baseline (before any edit, this tree): `go test ./internal/harness/ -count=1` → first run FAIL, then three consecutive `ok` runs (the sole first-run failure coincided with the template suite running concurrently in a parallel shell; judged machine contention, not a code defect — green on repeated runs per the concurrency-testing discipline). `go test ./internal/template/ -count=1` → FAIL, exactly `--- FAIL: TestRuleDateProvenance` (`.claude/rules/moai/core/hooks-system.md:368` date-provenance leak) — **pre-existing baseline red**, unrelated to this SPEC's template edits (`.moai/config/sections/*` and `.gitignore`), left untouched per scope discipline. `go test ./internal/hook/ -count=1` → `ok` (316.7s). `go test ./internal/hygiene/ -count=1` → `ok`.

### M0/B0 baseline (measured pre-B0; recorded as the B0 progress.md-only commit, which precedes the M2 untracking commit B1)

The baseline numbers below were measured before any tracking change; they are recorded verbatim in the B0 commit (which touches only `progress.md`).

### RED / mutant observations (E8; captured BEFORE the implementation that flips each)

- **EV-006 premise + mutant (AC-CAP-009).** Absence premise re-observed (table above). Test `TestRetentionPruneArchivesBeforeRewrite` written against the unchanged prune → `ok github.com/modu-ai/moai-adk/internal/harness 0.842s` (both subtests PASS — characterization of the default branch, as planned). Mutant: `prune` reordered to `overwriteWithEvents` before `archiveEvents` → FAIL, verbatim: `--- FAIL: TestRetentionPruneArchivesBeforeRewrite/failed_archive_leaves_log_byte-identical — log was rewritten despite the failed archive (pruneErr=retention: 아카이브 실패: 월별 아카이브 2026-08 기록 실패: 아카이브 임시 파일 생성: open .../archive/archive-452868282.jsonl.gz.tmp: permission denied): before 297 bytes, after 141 bytes`. Mutant reverted (`git diff` on retention.go empty); test green again (`ok ... 0.586s`). `retention.go` carries no change in the M1 commit.
- **EV-008 premise + behavioral RED (AC-CAP-012, clauses 1-2).** Test written against the pre-flip template (mode: report) → FAIL, verbatim: `no rotator row carries mode=apply (shipped template default did not reach the pass); rows=[{Unit:rotator Mode:report Outcome:summary Path:} {Unit:gc Mode:report Outcome:summary Path:}]`. After the M1 flip → PASS.
- **EV-001 premise + pre-edit RED (AC-CAP-003).** `TestTemplateGitignoreIgnoresLocalArtifacts` extended with the AC-CAP-003 cases, run before the template `.gitignore` edit → FAIL with exactly the four positive cases (`usage-log` / `lessons-inbox` / `*.corrupt` / monthly `.gz`: `ignored=false, want true`); all controls (tier-promotions, product.md, scaffolds) passed.
- **AC-CAP-010 RED.** `TestHomeCleanKeepFloors` written and run against the age-only backups selection → FAIL both fixtures, verbatim: `backup candidates = [removed-old-a removed-old-b removed-old-c removed-old-d removed-old-e], want [removed-old-a removed-old-b]` (fixture i) and `backup candidates = [removed-old-1 removed-old-2 removed-old-3], want [removed-old-2 removed-old-3]` (fixture ii); the release arm produced no failure (already correct, as the acceptance predicts).
- **EV-009/EV-010 premises + compile RED (AC-CAP-011).** Premises re-observed (table above). Final-form `TestProfileCapWarnsWithoutDeleting` run before the loader/advisory existed → build failure, verbatim: `undefined: hygiene.LoadProfileCapMaxBytes` / `undefined: hygiene.ProfileCapAdvisories` (×2 each). Post-green mutant (loader with compiled fallback) observed after M3 — see the M3 section below.

### M1 — decision-gated tests + template default flip

- Template `workflow.yaml` hygiene mode flipped `report` → `apply` (line 331) with the comment updated to state the new default; `make build` re-embedded the templates (exit 0).
- `go test ./internal/hook/ -run '^TestSessionStartAutoPathSplit$' -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 0.897s` (clauses 1-2 PASS: rotator apply rows, seeded verify target's `unit:gc / mode:apply` row, path-scan clause, fixture home byte-identical).
- `go test ./internal/harness/ -run '^TestRetentionPruneArchivesBeforeRewrite$' -count=1` → `ok` (0.833s).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Inputs: tier=M, scope ≈10 files (template .gitignore, workflow.yaml template, state.yaml template, gitignore_local_artifacts_test.go, retention_test.go, session_start_hygiene tests, clean_home.go, session_start_hygiene.go, state loader), domains=2 (Go code + template assets), language mix=Go+YAML, concurrency benefit=LOW (coding-heavy, interdependent milestone order), Agent Teams prereqs=not requested.
Decision: serial — one manager-develop carries M1→M0→M2→M3→M4 in order.
Justification: coding-heavy work takes the sequential path per the Anthropic coding-task caveat; milestones share files and ordering constraints (B0 before B1), so fan-out adds coordination without parallelism. Sweep preconditions (≥30 mechanical files) are not met.
Mode Selection recorded before the first run-phase Agent() spawn (grep token: serial).
