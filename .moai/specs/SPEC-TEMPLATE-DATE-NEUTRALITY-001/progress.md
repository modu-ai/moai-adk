# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-25
plan_iteration: 2
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, classify.sh]
requirements: 20
acceptance_criteria: 23
open_questions: 0
deferred_questions: 1
```

Plan-phase artifacts authored in the isolated worktree `.claude/worktrees/debt-clear` on branch `spec/template-date-neutrality`, based on `origin/main` at `c7309aeb6`.

**Iteration 2** — revised after plan-audit FAIL (0.55, Tier L threshold 0.85). Four user decisions absorbed as binding requirements (hybrid carve-out, DC-1 preserve, mirror-stamp preserve, CI isolated target); the fifth marker re-posed as an M2 step rather than a pre-Kickoff gate.

All 23 acceptance-criteria judgment commands were executed during plan phase; baselines are recorded verbatim in `acceptance.md` §B. The classification recipe is committed as `classify.sh` and reproduces the guard's 135-finding set exactly (verified by set diff).

Zero clarification markers remain in `plan.md`. One question is explicitly **deferred** to M2 (internal-incident prose disposition) because it requires the per-row inventory that M2 produces; it is not a blocker on Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

> M1-M3 evidence rows are not recorded in this section; the entries below cover M4-M5 only.

### M4-M5 — carve-out + report cap (2026-07-27)

Commits: `0888cf9ad` (M4), `b933c28e4` (M5). Single file changed: `internal/template/internal_content_leak_test.go`.

| AC | Status | Judgment command | Actual output |
|---|---|---|---|
| AC-TDN-009 | PASS | `awk '/^func (collectLeakViolations\|isPedagogicallyAllowed\|isDateAllowlisted)/,/^}/' internal/template/internal_content_leak_test.go \| grep -cE 'LineStart\|LineEnd\|lineNo\|LineNumber'` | `0` (window 68 lines, all 3 functions captured; injection probe returned `1`, revert returned `0`) |
| AC-TDN-010 half 1 | PASS | `grep -c "limit := 50" internal/template/internal_content_leak_test.go` | `0` |
| AC-TDN-010 half 2 | PASS | injection recipe on `NOTICE.md`, then `grep -oE 'full listing: [^ ]+'` | `full listing: /var/folders/.../T/moai-template-leak-4123967095.log`; `test -f` succeeded; probe reverted |
| AC-TDN-012 | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `ok`, `exit=0` (baseline was `78 occurrences, mode=strict`) |
| AC-TDN-002 (sanity) | PASS | narrow tier, same target | `ok`, `exit=0` |
| AC-TDN-013 (sanity) | PASS | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok`, `exit=0` |

Carve-out accounting — the 78 strict findings mapped 1:1 onto the 78 distinct `(file, date)` PRESERVE pairs of `triage.tsv`, all of class `S1-internal-date`; the S2 short-sha class contributed 0 both before and after:

| Mechanism | Categories | Findings carved |
|---|---|---|
| structural gate | DC-1 (48) + DC-4 (3) | 51 |
| content-anchored allowlist (27 entries) | DC-3 (9) + DC-2b (11) + DC-5 (7) | 27 |

Over-carving probes (each injected, observed, reverted; `git status --porcelain` clean after each):

| Probe | Expectation | Result |
|---|---|---|
| dated prose line in a skill body | flagged | flagged (`match=2029-12-31`) |
| different date inside an allowlisted file | flagged | flagged (`match=2027-05-05`) — allowlist is `(file, date)`-pinned, not file-wide |
| date repeated on a gated frontmatter line and a prose line | flagged | flagged (`match=2026-01-11`) — the structural gate does not poison per-file dedup |
| frontmatter-shaped line inside a fence | flagged | flagged (`match=2029-11-11`) — fence tracking holds |
| future attribution line in `NOTICE.md` (AC-TDN-015 A) | clean | clean |
| `**Import Date (probe)**: 2027-09-09` in `NOTICE.md` | clean | clean |
| future frontmatter bump (AC-TDN-015 B) | clean | clean |
| arbitrary dated comment in `NOTICE.md` | flagged | flagged — DC-4 gate is `(file AND attribution-line-shape)`, not whole-file |

Regression + build + lint: `go test -count=1 ./...` exit 0 (105 ok / 0 FAIL / 3 no-test); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m` exit 0, `0 issues` (no new, no pre-existing baseline); `make build` exit 0 with `catalog.yaml` SHA256 unchanged (`999046874371ce42d930b097af68b7101812790d4e8963675dacf0bc8b895225` before and after), as expected for a `_test.go`-only change.

### M6 — CI enforcement, adopted (2026-07-27)

Single file changed: `.github/workflows/template-neutrality-check.yaml` (one step appended to the `template-neutrality` job; the two pre-existing steps and the `paths:` filters are byte-unchanged).

**Adoption decision.** `design.md` §C gates the workflow edit on P1-P3. All three were re-measured against this tree before the edit, so M6 takes the **adopted** branch and no `precondition_failed:` line is written (writing one while adopting would make AC-TDN-011's two-branch check ambiguous).

| Precondition | Command | Observed |
|---|---|---|
| P1 — strict tier reports zero findings | `MOAI_TEMPLATE_LEAK_STRICT=1 go test -count=1 -run TestTemplateNoInternalContentLeak ./internal/template/` | `ok  github.com/modu-ai/moai-adk/internal/template  1.004s`, `exit=0` |
| P2 probe A — future attribution record | append `The following docs (imported 2029-03-03) are incorporated:` to `NOTICE.md`, re-run strict tier | `ok … 0.787s`, `probeA_exit=0` — clean; reverted |
| P2 probe B — future frontmatter bump | `moai-workflow-tdd/SKILL.md:22` `updated: "2026-02-03"` → `"2028-05-09"`, re-run strict tier | `ok … 0.932s`, `probeB_exit=0` — clean; reverted |
| P3 — narrow tier + neutrality target | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1`; `… -run TestTemplateNeutralityAudit -count=1` | `ok … 0.683s` / `ok … 0.330s`, both `exit=0` |

`git status --porcelain` after each probe revert showed only ` M .github/workflows/template-neutrality-check.yaml` — no probe residue.

**Step shape.** The test name is written **unquoted** (`-run TestTemplateNoInternalContentLeak`) so AC-TDN-019's grep matches the executable step rather than a prose comment. The two pre-existing steps quote their targets (`-run 'TestTemplateNeutralityAudit'` at line 61, `-run 'TestTemplateNoInternalContentLeak'` at line 71) and therefore do not match that unquoted pattern — which is why the criterion read `0` before this milestone even though a narrow-tier step already existed. The new step's comment deliberately omits the `-run <testname>` literal so the count stays at exactly 1. `MOAI_TEMPLATE_LEAK_STRICT` is **step-scoped**, not job-wide: a job-wide value would silently place the pre-existing narrow-tier step into strict mode and change a different SPEC's contract.

The workflow's in-file comment about "pre-existing failures unrelated to this SPEC" is left in place per `design.md` §C — that claim is withdrawn *as this SPEC's rationale* without being asserted false in general, because a stray `.moai` marker inside `internal/template/` genuinely does make two output-styles tests fail.

| AC | Status | Judgment command | Actual output |
|---|---|---|---|
| AC-TDN-011 | PASS | `grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . \| grep -v "internal_content_leak_test.go"` | `.github/workflows/template-neutrality-check.yaml:85:          MOAI_TEMPLATE_LEAK_STRICT: '1'` — 1 match, inside `.github/workflows/` (adopted branch satisfied) |
| AC-TDN-011 (not-adopted counter) | PASS | `grep -cE '^precondition_failed: (P1\|P2\|P3)' progress.md` | `0` — correct for the adopted branch |
| AC-TDN-015 | PASS | probes A and B above, each against the real file, each reverted | both clean (`exit=0`); working tree clean after each revert |
| AC-TDN-019 | PASS | `grep -c -- '-run TestTemplateNoInternalContentLeak' .github/workflows/template-neutrality-check.yaml` | `1`; the match is line 88, `            -run TestTemplateNoInternalContentLeak -v` — the new step's `run:` body, not a comment |

YAML validity (not covered by any AC — a syntactically broken workflow would silently stop running in CI): `yaml.safe_load` parses the file and reports 5 steps, with `env={'MOAI_TEMPLATE_LEAK_STRICT': '1'}` present on step 4 only and `env=None` on the other four.

### M7 — full acceptance matrix + non-regression sweep (2026-07-27)

Every criterion below was executed **literally** as written in `acceptance.md` §B, from the worktree root. Two criteria carry a cwd-scope defect in their published command (they name a bare filename or `go test .`, which presumes `cwd=internal/template` — an invocation `acceptance.md` §Scope itself forbids); for those, both the literal result and the repo-root-corrected result are recorded, and the literal result is marked as such rather than reported as a pass.

| AC | Status | Judgment command | Actual output |
|---|---|---|---|
| AC-TDN-001 | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `ok … 1.004s`, `exit=0` today. The criterion is the M1 **baseline capture**: the `135 occurrences, mode=strict` baseline is recorded verbatim in `acceptance.md` §B and reconciled by AC-TDN-021. Today's `exit=0` is the post-M4 state, not a re-measurement of the baseline |
| AC-TDN-002 | PASS | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `ok … 0.683s`, `exit=0` |
| AC-TDN-003 | PASS (M2-time) | classifier row count vs `triage.tsv` data rows | `classifier rows: 88` / `tsv data rows: 180`. Equal at M2 (180 = 180); the M7 delta is exactly `92`, the REMOVE count — `180 − 92 = 88`. The inventory is a record of the pre-remediation tree, so post-M3 divergence by precisely the REMOVE count is the expected behaviour and is itself corroborating evidence |
| AC-TDN-004 | PASS | `awk -F'\t' 'NR>1 && ($6=="" \|\| $6=="TODO" \|\| $6=="UNTRIAGED" \|\| $6=="TBD")' triage.tsv \| wc -l` | `0`. Non-vacuous: `triage.tsv` exists with 180 data rows and column 6 is fully populated — `88 PRESERVE` / `92 REMOVE`, no third value |
| AC-TDN-005 | PASS | classifier `$5=="DC-3"` row count | `13` (baseline `13`) |
| AC-TDN-005 (2nd form) | PASS | `grep -rhE '\b2026-11-22\b' internal/template/templates --include=… \| wc -l` | `13` (baseline `13`) |
| AC-TDN-006 | PASS | `grep -cE 'imported 202[6-9]-…' internal/template/templates/.claude/rules/moai/NOTICE.md` | `3` (baseline `3`) |
| AC-TDN-007 | PASS | classifier `$5=="DC-2a"` row count | `0` (baseline `80`). Non-vacuous: the same classifier invocation emits 88 rows and returns `48` for DC-1 / `13` for DC-3, so the pipeline is live and the DC-2a filter genuinely matches nothing |
| AC-TDN-007 (2nd form) | PASS | three-shape `Updated:` grep, excluding `moai-foundation-cc/reference/` | `0` (baseline `78`). Non-vacuous: dropping only the `grep -v` exclusion returns `11`, proving the regex still matches real lines |
| AC-TDN-008 | PASS | classifier `$5=="DC-1"` row count | `48` (baseline `48`) |
| AC-TDN-008 (2nd form) | PASS | whole-file indented `updated:` grep | `49` (baseline `49`) |
| AC-TDN-009 | PASS (corrected form) | literal: `awk '…' internal_content_leak_test.go \| grep -cE 'LineStart\|…'` | literal → `awk: can't open file internal_content_leak_test.go` then `0`. **The literal `0` is vacuous** — `test -f internal_content_leak_test.go` fails from the repo root, so the count is an artifact of the missing file, and `0` happens to be the PASS condition. Repo-root-corrected (`internal/template/` prefix) → `0` with a non-empty 68-line awk window containing all `3` `^func ` headers |
| AC-TDN-010 half 1 | PASS | `grep -c "limit := 50" internal/template/internal_content_leak_test.go` | `0` (baseline `1`). Non-vacuous: the file exists and the cap now lives at line 832 as `const leakReportConsoleCap = 50`, consumed at line 1104 — the literal is absent because it was renamed, not because the file is unreadable |
| AC-TDN-010 half 2 | PASS (corrected form) | literal: inject probe → `go test .` → `grep -oE 'full listing: …'` | literal → `NO-PATH-EMITTED`, because `go test .` from the repo root reports `no Go files in …` / `FAIL . [setup failed]` — the same cwd-scope defect as AC-TDN-009, not a guard failure. Corrected (`./internal/template/`) → `full listing: /var/folders/…/T/moai-template-leak-2898943849.log`; `test -f` succeeded (3 lines) and the listing named the injected probe (`class=S1-internal-date \| match=2029-12-31`). Probe reverted; tree clean |
| AC-TDN-011 | PASS | see M6 table above | 1 match inside `.github/workflows/`; not-adopted counter `0` |
| AC-TDN-012 | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `ok … 0.779s`, `exit=0` (baseline `exit=1`, `135 occurrences`) |
| AC-TDN-013 | PASS | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok … 0.330s`, `exit=0` |
| AC-TDN-014 | PASS | `go build ./...` | `exit=0`, no output |
| AC-TDN-015 | PASS | probes A and B (M6 table) | both clean, both reverted |
| AC-TDN-016 | PASS | M7 re-run over the **actual** edited-file list, not the 4-file sample | the mirror allowlist holds `7` entries; the run-phase edit set under `internal/template/templates/` holds `67` files; `comm -12` of the two sorted lists is **empty** — intersection `0`, so no paired local edit is owed. Non-vacuous: the positive control `.claude/rules/moai/workflow/session-handoff.md` returns `IN-ALLOWLIST`, and `0` of the 67 edited files sit under `.claude/rules/` at all |
| AC-TDN-017 | PASS | classifier `$5=="DC-2b"` row count | `11` (baseline `11`) |
| AC-TDN-017 (2nd form) | PASS | `grep -rlE '^(\*\*)?Updated(\*\*)?:…' …/moai-foundation-cc/reference --include='*.md' \| wc -l` | `11` (baseline `11`) |
| AC-TDN-018 | PASS | placeholder-token grep over `internal/template/templates` | `0` (baseline `0`). Non-vacuous: the identical `grep -rlE` shape with a pattern known present returns `135` files, so the scan is live |
| AC-TDN-019 | PASS | `grep -c -- '-run TestTemplateNoInternalContentLeak' .github/workflows/template-neutrality-check.yaml` | `1`, matching line 88 (the new step's `run:` body). Positive control `grep -c -- '-run TestTemplateNeutralityAudit'` → `1` |
| AC-TDN-020 | PASS | `git status --porcelain -- '.claude/' '.moai/' \| grep -v 'SPEC-TEMPLATE-DATE-NEUTRALITY-001' \| wc -l` | `0`. Non-vacuous: with this progress.md edit in the tree the unfiltered pathspec form reports ` M .moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-001/progress.md`, so the pathspec is live and the `0` is produced by the SPEC-directory filter, not by an inert query |
| AC-TDN-021 | PASS (M2-time) | classifier distinct `(file, date)` count | `78` today vs the `135` M2 baseline that matched the guard exactly. The guard now reports `0` **by design** (the M4 carve-out), so the M7 re-run cannot re-assert equality against the guard; instead it reconciles against the inventory — `awk '$6=="PRESERVE"' triage.tsv \| cut -f1,2 \| sort -u \| wc -l` also returns `78`, an exact match |
| AC-TDN-022 | PASS | `git diff --name-only $(git merge-base origin/main HEAD)..HEAD` with the 5 excludes | `0`. Non-vacuous: the same diff without excludes reports `73` changed files, so every one of the 73 falls inside the declared scope |
| AC-TDN-023 | PASS | classifier dual-shape findings (any category pair) | `0` (baseline `13`) |
| AC-TDN-023 (PASS-condition cmd) | PASS | classifier `DC-1`/`DC-2a` conflicts | `0` (baseline `12`) |
| AC-TDN-023 (2nd form) | PASS | `find … \| while read` frontmatter-and-prose same-literal scan | `0` (baseline `12`). Non-vacuous: the loop walks `310` markdown files and its frontmatter-side predicate fires `49` times, so only the conjunction is empty |
| AC-TDN-023 (conditional) | PASS | `DC-5` rows sharing a finding with a `DC-1` row must cite `REQ-TDN-019` | `0` unannotated. Non-vacuous: `triage.tsv` holds `22` DC-5 rows of which `1` carries a `REQ-TDN-019` rationale, and `grep -P` is confirmed supported on this host (a `grep` without `-P` would make the `\Q…\E` check silently vacuous) |

Result: **23 / 23 PASS, 0 FAIL.** AC-TDN-003 and AC-TDN-021 are M2-milestone criteria whose PASS was established against the pre-remediation tree; both re-reconcile exactly at M7 (`180 − 92 = 88`; `78 = 78`).

**AC-TDN-012 joint-reachability arithmetic, closed with the measured `k`.** `k` (DC-5 rows adjudicated REMOVE at M2) is **12**, measured from `triage.tsv` rather than carried over:

| Category | PRESERVE | REMOVE |
|---|---:|---:|
| DC-1 | 48 | 0 |
| DC-2a | 0 | 80 |
| DC-2b | 11 | 0 |
| DC-3 | 13 | 0 |
| DC-4 | 6 | 0 |
| DC-5 | 10 | 12 |
| **total** | **88** | **92** |

Deleted by remediation `= 80 + k = 92`; carved out by the guard `= 100 − k = 88`; total `180`. Both figures are independently confirmed by the tree: the classifier emits `88` rows today (matching the PRESERVE total) and `180 − 88 = 92` (matching the REMOVE total).

**Non-regression sweep.**

| Check | Command | Result |
|---|---|---|
| Full suite | `go test -count=1 ./...` | `exit=0` — 105 ok / **0 FAIL** / 3 no-test-files |
| Build | `go build ./...` | `exit=0` |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | `exit=0` |
| Lint | `golangci-lint run --timeout=3m` | `0 issues.` |
| Vet | `go vet ./...` | `exit=0`, 0 bytes of output |

`make build` was **not** re-run at M6/M7 and `catalog.yaml` was **not** regenerated: `git status --porcelain -- internal/template/templates/ internal/template/catalog.yaml` is empty, because M6 touches only `.github/`, so the template-content hashes cannot have shifted.

**Modernizer decision (recorded, not applied).** `internal/template/internal_content_leak_test.go:606` holds a `for _, f := range dc4AttributionFiles { if relPath == f { return true } }` loop that a `slices.Contains` modernizer would rewrite. It was **not** changed: `golangci-lint` reports `0 issues` and `go vet` exits 0, so the pattern is below this project's own gate, and `internal_content_leak_test.go` is M4/M5 territory — editing it at M7 would be an out-of-scope drive-by. No repo-wide `make fix` was run.

## §F Phase 4 Mode Selection

**Input parameters**

| Signal | Value |
|---|---|
| tier | L |
| scope (files) | 67 template files carry >=1 REMOVE row; 116 files carry >=1 finding |
| domain count | 2 (template markdown/yaml tree; Go guard file + CI workflow) |
| file language mix | markdown / yaml / tmpl (M3), Go (M4-M5), yaml workflow (M6) |
| concurrency benefit | LOW — M3 carries two distinct transform shapes plus a per-edit surrounding-block judgment (research.md gap G5) |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| 1 trivial | no | 92 line-level deletions across 67 files plus a Go structural change; not a single-line edit |
| 2 background | no | write work; the orchestrator needs each milestone's result before sequencing the next |
| 3 agent-team | no | RETIRED (orchestration-mode-selection.md §C.1) |
| 4 parallel | no | coding-heavy and write-scoped; the Anthropic coding-task parallelism caveat routes this away from parallel fan-out |
| 5 sub-agent | **yes** | sequential `manager-develop` per milestone; each milestone's exit condition gates the next |
| 6 workflow | no | file count clears the ~30 soft boundary, but M3 is NOT a single uniform transform rule: 80 DC-2a stamp deletions and 12 bespoke DC-5 edits are two shapes, and the G5 orphan-block check requires per-edit judgment. §B.2 tie-breaker (multi-rule / semantic work prefers Mode 5) applies |

**Decision: sub-agent**

**Justification.** Mode 6 was evaluated seriously because the raw file count (67) clears the `~30` soft boundary, but the Mode 6 capability gate in §C.3 requires a *single uniform mechanical transform rule with no inter-file dependency*. M3 fails that on two counts: it carries two distinct edit shapes (a standalone footer-stamp line delete versus a mid-paragraph date-phrase excision), and every deletion needs the surrounding block inspected for the orphaned-header regression `research.md` gap G5 names, which no acceptance criterion mechanizes. Per the §B.2 tie-breaker, multi-rule work at any file count stays on Mode 5. Milestones M4-M6 are additionally sequential by construction (M5's report-cap change is untestable until M4 drives the strict tier to zero findings), so a sequential sub-agent chain matches the dependency graph.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-27
milestones: [M1, M2, M3, M4, M5, M6, M7]
milestone_summary:
  M1: rulings confirmed; strict-tier baseline captured (135 findings)
  M2: triage.tsv authored — 180 occurrence-class rows, 135 (file,date) findings, k=12
  M3: 92 REMOVE rows remediated — 135 -> 78 strict findings
  M4: hybrid carve-out (structural gate + content-anchored allowlist) — 78 -> 0
  M5: report cap replaced by a named constant + full-listing path
  M6: CI strict-tier step adopted (P1-P3 all verified)
  M7: full 23-criterion acceptance matrix + non-regression sweep
k: 12
occurrence_class_rows: 180
findings_baseline: 135
findings_final: 0
rows_removed: 92
rows_preserved: 88
acceptance_criteria_total: 23
ac_pass_count: 23
ac_fail_count: 0
ac_pass_with_debt_count: 0
m6_branch: adopted
precondition_failures: 0
run_phase_commits:
  - e9ae9923f  # M2 triage inventory + draft -> in-progress
  - 12c29b811  # M3 remediate the 92-row REMOVE set
  - 5d49c1ab8  # AC-TDN-022 exclude-list amendment + 2025-stamp scope boundary
  - 0888cf9ad  # M4 hybrid carve-out
  - b933c28e4  # M5 report cap names the full-listing path
  - e8644a8ef  # M4-M5 run-phase evidence in progress.md
  - pending-backfill-m6-m7  # M6 CI step + M7 acceptance matrix (this commit)
total_run_phase_files: 70
files_breakdown:
  template_tree: 67
  guard_file: 1          # internal/template/internal_content_leak_test.go
  ci_workflow: 1         # .github/workflows/template-neutrality-check.yaml
  generated_catalog: 1   # internal/template/catalog.yaml (make build output, M3)
mN_commit_strategy: per-milestone commits on feat/SPEC-TEMPLATE-DATE-NEUTRALITY-001; no squash, no amend, no force-push
full_test_suite: "go test -count=1 ./... exit 0 — 105 ok / 0 FAIL / 3 no-test"
cross_platform_build:
  darwin_amd64_native: "go build ./... exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
new_warnings_or_lints_introduced: 0
lint_status: "golangci-lint run --timeout=3m -> 0 issues.; go vet ./... exit 0"
catalog_regenerated_at_m6_m7: false
pushed: false
```

The run phase is complete and audit-ready. `status: in-progress` is retained deliberately — the `in-progress -> implemented -> completed` transition and the `sync_commit_sha` population in §E.4 belong to manager-docs at sync phase.

`run_commit_sha` for the M6/M7 commit is recorded as `pending-backfill-m6-m7` because a commit cannot reference its own hash; it is backfilled in a follow-up commit per the SHA-placeholder backfill exemption in `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
