# SPEC-WORKTREE-BASEREF-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: CLOSED — plan-audit complete, operator kickoff approval granted, run-phase entry authorized
plan_complete_at: 2026-08-27
plan_audit_iter: 2 complete — verdict **PASS-WITH-DEBT 0.92 harmonic, 7/7 must-pass clear, zero blocking defects** (`.moai/reports/t313/plan-audit-iter2.md`). Iteration 1: FAIL 0.78 harmonic, 7/7 must-pass clear (`.moai/reports/t313/plan-audit-iter1.md`).
kickoff_approval: GRANTED by the operator, conditional on repairing debt item N1 first. N1 repaired in spec version 0.3.1 (below) before run-phase entry; the condition is discharged. Run-phase progression mode: **autonomous** — there is no milestone checkpoint at which a human reviews an intermediate reading, so criterion wording that admits a permissive interpretation was pinned rather than left open (see N2).
artifacts: spec.md (0.3.1, draft), plan.md, acceptance.md, progress.md — Tier M
tree: 48eb945df (worktree t313, branch WT-worktree-baseref)
counts: 16 GEARS requirements / 16 acceptance criteria (all MUST) — exactly at the Tier M ceiling of 16 REQ / 16 AC (`spec-workflow.md:148`). Unchanged by the 0.3.1 repairs: both are criterion-layer wording fixes, no id added, renamed, or renumbered.
iter2_repairs: D1 (REQ-WBR-004 now covered by AC-WBR-016, read-seam call count), D2 (vacuous `-run` pass mode closed in §D.3), D3 (AC-WBR-012 restated as REQ-WBR-015's three-part conjunction + guard-existence + mutation check), D4 (R1 premise corrected, consumer 1 narrowed to the primary checkout), A5 (REQ-WBR-011 bound to REQ-WBR-009's predicate as sole authority, AC-WBR-008 asserts the shared helper), plus debt D5-D10 folded in
iter2_changes: `.moai/reports/t313/spec-iter2-changes.md`
post_audit_repairs (spec 0.3.1): `.moai/reports/t313/spec-n1n2-changes.md`
  - **N1 (major, was blocking-carried)** — AC-WBR-013 check (1) probed only `internal/template/templates/$f`, so a correct implementation of plan §D emitted a false `NO-TEMPLATE-COUNTERPART` for `.moai/config/sections/git-strategy.yaml` (counterpart ships only as `.tmpl`; plan §B G5). Probe now accepts the plain OR the `.tmpl` form and reports only when neither changed. Re-measured against real history: old probe emits the false line on `664cd6eae^..664cd6eae`, new probe emits nothing; negative control `63b4628a6^..63b4628a6` (genuine orphan) still emits it, so detection was not weakened.
  - **N2 (minor-to-major, was blocking-carried)** — AC-WBR-016's "read seam" pinned to the **alignment-entry (configured-value) read**; the narrowing alternative was rejected because it would let the criterion lapse on the shipped empty default, which is the common case. Recorded in the criterion text and at plan.md §A D3.2.
debt_carried_into_run_phase (7 items, from the iter-2 verdict § Debt Carried Into Run-Phase):
  1. N1 — REPAIRED in 0.3.1 (no longer carried).
  2. N2 — REPAIRED in 0.3.1 (no longer carried).
  3. **Seam obligation is not optional (M2/M3)** — `internal/hook` carries no function-variable seam today; M2 must introduce one in its own new file. Budget for it; it is what makes AC-WBR-016 and AC-WBR-008's third assertion testable at all.
  4. **N4 — AC-WBR-013 check (2) is dormant, not correct.** Path mixing: the `[ -f ]` guard reconstructs a template-tree path while `diff -q` compares a sibling in `$b`'s own tree. Enumerates nothing for this SPEC (plan §D lists no hook wrapper); fix before any scope growth that touches one.
  5. **N5 — AC-WBR-002's first command has a prose-only pass condition.** Carry-over, not an iter-2 regression. Add an explicit value-emptiness assertion if run-phase wants it mechanical.
  6. **N3 — REQ-WBR-012 carries two GEARS modalities in one id.** Cosmetic; no run-phase action.
  7. **Folding traceability thinness (§ B1)** — REQ-WBR-004's narrowing and REQ-WBR-013's preservation clause are each carried by one AC half. Treat **both halves of AC-WBR-016 as separately mandatory**; an id-level matrix check will not notice a dropped half.
  Plus inherited and unchanged from iteration 1: **G2** (`EnterWorktree`'s `origin/HEAD` read is inferred, not read from source — the doctor item is the stated fallback), **G4** (`moai doctor` never executed), **G6** (rendered attribute order unmeasured — AC-WBR-011's two-branch condition MUST be collapsed to the single true form in run-phase).
blocker: none — D2 (widget shape) CLOSED by operator ruling 2026-08-27: `TypeText` with `main` / `develop` named in the field description; the new-combo-widget alternative rejected on template-neutrality grounds. Ruling + rationale recorded at plan.md §A D2.1; consequences at REQ-WBR-009 / REQ-WBR-012 / REQ-WBR-014, AC-WBR-009 / AC-WBR-011 / AC-WBR-015.

## §E.2 Run-phase Evidence

Full verdict with per-AC commands, exit codes, `=== RUN` counts, and evidence paths: `.moai/reports/t313/run-verdict.md`.
Verbatim command output: `.moai/state/verify/t313/` (persistent — survives `/tmp` clearance).
Measured in worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`, at HEAD `8c46460ff` unless a row says otherwise.

### AC matrix

| AC | Status | Verification command | Actual output | Evidence |
|---|---|---|---|---|
| AC-WBR-001 | PASS | `go test ./internal/config -run 'GitStrategy' -count=1 -v` | exit 0, 35 `=== RUN`, 0 vacuous | `ac-wbr-001.txt` |
| AC-WBR-002 | PASS | value-side grep + `go test ./internal/template -run 'WorktreeBaseBranchTemplate' -v` | branch-grep exit 1 (no match); guard exit 0, 1 `=== RUN` | `ac-wbr-002.txt`, `ac-wbr-002-guard.txt` |
| AC-WBR-003 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Unset\|Empty)' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-003.txt` |
| AC-WBR-004 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Match' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-004.txt` |
| AC-WBR-005 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Mismatch' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-005.txt` |
| AC-WBR-006 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(FailOpen\|GitError)' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-006.txt` |
| AC-WBR-007 | PASS | `go test ./internal/cli -run 'SessionWorktree.*Base' -count=1 -v -timeout 600s` | exit 0, 5 `=== RUN`, 0 vacuous | `ac-wbr-007.txt` |
| AC-WBR-008 | PASS | `... -run 'SessionWorktree.*(NoBase\|Unresolvable)'` + `... -run 'SessionWorktree.*(SharedPredicate\|Resolver)'` | exit 0/0, 3 + 2 `=== RUN`, 0 vacuous | `ac-wbr-008a.txt`, `ac-wbr-008b.txt` |
| AC-WBR-009 | PASS | `go test ./internal/cli -run 'Doctor.*WorktreeBaseBranch' -v` + `./bin/moai doctor --check 'Worktree Base Branch'` | exit 0, 7 `=== RUN`; CLI exit 0, item reachable by exact name | `ac-wbr-009.txt`, `ac-wbr-009-doctor-run.txt` |
| AC-WBR-010 | PASS | `go test ./internal/web -run 'WorktreeBaseBranch' -v` + `go test ./internal/settings -run 'AllFields' -v` | exit 0/0, 2 + 1 `=== RUN`, 0 vacuous | `ac-wbr-010-web.txt`, `ac-wbr-010-allfields.txt` |
| AC-WBR-011 | PASS | `go test ./internal/settings -run 'WorktreeBaseBranch.*Type' -v` + `go test ./internal/web -run 'WorktreeBaseBranch.*(Text\|FreeText)' -v` | exit 0/0, 1 + 1 `=== RUN`, 0 vacuous | `ac-wbr-011-schema.txt`, `ac-wbr-011-render.txt` |
| AC-WBR-012 | PASS | `go test ./internal/web ./internal/hook ./internal/cli -run 'WorktreeBaseBranch' -v` + mutation check | exit 0, 21 `=== RUN`, 0 vacuous; mutant (FieldDef removed) exit 1 | `ac-wbr-012.txt`, `ac-wbr-012-mutation.txt` |
| AC-WBR-013 | PASS | `make build` + diff-scoped parity probes (N1-repaired form) | exit 0; 0 `NO-TEMPLATE-COUNTERPART`, 0 `DRIFT`; mirror byte-identical | `ac-wbr-013-parity.txt`, `make-build-final.txt` |
| AC-WBR-014 | PASS | `go test ./internal/settings -run 'GitStrategy.*RoundTrip' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-014.txt` |
| AC-WBR-015 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Unresolvable' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-015.txt` |
| AC-WBR-016 | PASS (both halves) | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Fires\|Registered\|Once\|LinkedWorktree\|NotPrimary)' -v` | exit 0, 4 `=== RUN`, 0 vacuous | `ac-wbr-016.txt` |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| `go build ./...` clean | PASS (exit 0) | `go-build.txt` |
| Cross-platform build (windows/amd64, linux/amd64) | PASS (exit 0, 0) | `go-build.txt` |
| `go vet` clean on the five affected packages | PASS (exit 0) | `go-vet.txt` |
| `golangci-lint run` clean on the five affected packages | PASS (exit 0, `0 issues.`) | `golangci-lint.txt` |
| No file under `.claude/skills/moai-workflow-project/schemas/` modified (t316 boundary) | PASS (diff empty) | recorded in run-verdict |
| Full-package regression: config / settings / web / hook / cli | PASS (all `ok`) | `coverage.txt`, `coverage-cli.txt` |

### Scope note — one file outside the plan §D write list

`internal/config/types.go` `ModeProfile` gained three pass-through fields (`develop_branch`, `release_branch_prefix`, `rc_version_format`). spec.md §C lists the ModeProfile schema gap as out of scope, and REQ-WBR-013 requires this SPEC's write path not to drop those keys — the two are only jointly satisfiable by modelling them, because `saveSection` re-marshals the struct with no merge step (read at `internal/config/manager.go:418`). MEASURED before the change: the typed save deleted all three (`m5-settings-1.txt`). No accessor and no consumer was added; the fields exist only to survive the round trip. Recorded here rather than absorbed silently.

Three further cascade edits, all inside the SPEC's own envelope: the shipped-key triage inventory (a new shipped template key must be classified), the free-text whitelist guard (a new `TypeText` field must be justified), and the doctor golden snapshots (a new diagnostic row).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: pending-backfill-run-final   # M6 docs commit 8c46460ff; the evidence commit that carries this block cannot name its own hash
run_status: COMPLETE — 16/16 AC PASS, zero blockers
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt_count: 0
milestone_commits:
  M1: 81808d85b   # config schema + neutral template default
  M2: cf2955ed5   # SessionStart origin/HEAD alignment
  M3: 9e1ea4226   # git worktree add base operand
  M4: 5658988be   # Worktree Base Branch doctor diagnostic
  M5: 04c645a68   # web free-text field + guards + round-trip preservation
  M6: 8c46460ff   # worktree rule documentation + mirror
preserve_list_post_run_count: 0 unrelated files modified
l44_pre_commit_fetch: HEAD + branch re-read immediately before each of the 6 commits; unchanged each time (d0bc4bba5 → 8c46460ff, linear)
l44_post_push_fetch: n/a — NOT pushed; integration is the orchestrator's call per the dispatch
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; the 2 errcheck findings this run introduced were fixed before the final measurement)
cross_platform_build:
  darwin_native: exit 0
  windows_amd64: exit 0
  linux_amd64: exit 0
coverage:
  internal/config: 80.6%
  internal/hook: 85.1%
  internal/settings: 90.3%
  internal/cli: 79.6%
  internal/web: 66.8%
  note: package-level pre-existing baselines, not a per-change figure; no baseline was measured before this SPEC, so no delta is claimed
total_run_phase_files: 21 changed (11 production, 7 test, 3 golden snapshots) across 6 commits
m1_to_mN_commit_strategy: one commit per milestone, each naming card t313 in its body
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: de2042416   # backfilled in the immediately following commit (a commit cannot name its own hash)
sync_status: COMPLETE — sync-audit PASS 0.90, zero blocking findings, merge-ready into develop
sync_audit_verdict: "PASS 0.90 harmonic (Functionality 0.95 / Security 1.00 / Craft 0.80 / Consistency 0.90); must-pass firewall clear on both dimensions; evaluator profile `default`, weights 40/25/20/15"
sync_audit_report: .moai/reports/t313/sync-audit.md
sync_audit_ac_reproduction: 16/16 MUST criteria re-executed independently in this tree; 0 vacuous passes; all five touched package trees green in full (config, settings, web, hook, cli); go vet rc 0; golangci-lint "0 issues."
b12_self_test_a: "pre-emission grep — `grep -c 'SPEC-WORKTREE-BASEREF-001' CHANGELOG.md` returned 0 before the write (no duplicate entry from a parallel session)"
b12_self_test_b: "AC count match — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` returned 16; the CHANGELOG entry states 16 acceptance criteria (AC-WBR-001..016)"
b12_self_test_c: "file-path verification — all 12 paths cited in the CHANGELOG entry confirmed present via `ls -1` (spec.md, git-strategy.yaml, types.go, worktree_base_branch.go, session_worktree.go, doctor_worktree_base.go, schema_sections.go, git-strategy.yaml.tmpl, worktree-integration.md, sync-audit.md, session_start.go, i18n.js)"
changelog_entry_position: "CHANGELOG.md [Unreleased] → ### Added, first entry (line 12)"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (merged into this single sync commit per the 3-phase close); `updated:` refreshed to 2026-08-27"
  plan_md: "no frontmatter block — no transition applicable"
  acceptance_md: "no frontmatter block — no transition applicable"
  progress_md: "no frontmatter block — no transition applicable"
docs_review:
  m6_commit: 8c46460ff
  surfaces: ".claude/rules/moai/workflow/worktree-integration.md + internal/template/templates/ mirror (byte-identical, `diff -q` clean)"
  verdict: "ACCURATE AND COMPLETE — no doc repair needed. Five claims re-verified against the shipped code in this tree: (1) the key sits at the root of `git_strategy` in `.moai/config/sections/git-strategy.yaml` (confirmed in the local file and at `internal/config/types.go:174`); (2) the session-start half fires only from the primary checkout (gate precedes the read, `internal/hook/worktree_base_branch.go:91-100`); (3) the `git worktree add` half is honoured from any tree (`internal/cli/session_worktree.go:215-221`); (4) the unresolvable value falls back to the no-operand form (`session_worktree.go:216-218`, `gitWorktreeAddArgs`); (5) the doctor item is reachable by the exact name `Worktree Base Branch` (`internal/cli/doctor_worktree_base.go`). The shipped template default is empty (`git-strategy.yaml.tmpl:12`), matching the documented neutral default."
  docs_site: "no gap — docs-site does not enumerate individual `git_strategy` keys (no `worktree_root` / `github_username` page mentions either); the web-console page describes the Git & Worktree group at group level. The user-facing field strings ship in all four locales via `internal/web/assets/i18n.js`."
debt_findings_recorded (6, all optional, none blocking):
  F1: "MINOR — `internal/config/types.go:121,174`. AC-WBR-001's second command expects exactly 1 grep hit; the tree returns 2 (line 121 is a doc-comment mention added by the M5 pass-through block). Substantive reading passes; a verbatim re-run reads red on a correct tree. Disposition: DEFERRED — criterion-text repair (narrow to `grep -c '^\\s*WorktreeBaseBranch string'`), no runtime effect."
  F2: "MINOR — coverage. Package aggregates below the 85% Craft threshold (config 80.6%, cli 79.6%, web 66.8%) are pre-existing baselines, not this delta; the added code is well covered (RunWorktreeBaseAlignment 100.0%, gitWorktreeAddArgs 100.0%, checkWorktreeBaseBranch 90.0%). The four real git-wrapper seams have no standing automated guard — the auditor exercised them by hand (A6). Disposition: DEFERRED — a `testing.Short()`-skipped integration test over a throwaway bare-origin clone would convert those manual runs into a guard. Highest-leverage follow-up alongside F4."
  F3: "MINOR — `acceptance.md:243`. AC-WBR-013 check (1) excludes only `^.moai/specs/`, so this diff's `.moai/reports/t313/*.md` evidence files emit false NO-TEMPLATE-COUNTERPART lines; the implementer added `grep -v '^.moai/reports/'` and disclosed the deviation in `ac-wbr-013-parity.txt`. Auditor reproduced both forms and agreed. Disposition: DEFERRED — fold `^.moai/reports/` into the criterion's exclusion."
  F4: "MINOR — `internal/settings/worktree_base_branch_test.go:94-98`. AC-WBR-014's guard asserts the three unmodelled key NAMES survive the typed save, not their VALUES; a regression writing `develop_branch: \"\"` would stay green. The values do survive today (read path verified, and the local `git-strategy.yaml` is intact after this SPEC's one-line edit) — a guard-strength gap, not a live defect. Disposition: DEFERRED — assert full `key: value` pairs. Highest-leverage follow-up alongside F2."
  F5: "INFORMATIONAL — `internal/hook/worktree_base_branch.go:171,190`, `internal/cli/session_worktree.go:250-254`. No `--` end-of-options separator on the three git invocations carrying the free-text value. Every call uses `exec.Command` with a separate argument vector — no shell, no interpolation, so no injection surface; the two option-position calls are additionally gated behind the resolvability predicate, so an option-shaped value is reachable only if a remote-tracking ref of that literal name exists. Disposition: DEFERRED — optional hardening, add `--` before the branch operand."
  F6: "MINOR — `acceptance.md:335`. §D.3's `grep -q '[no tests to run]'` VACUOUS clause is over-broad: it fails every multi-package invocation in the file (10 such lines per `./internal/hook/...` run) including correct ones, because sibling packages legitimately match no test. The operative clause is the `=== RUN` count and the auditor applied it that way. Disposition: DEFERRED — scope the string check to the owning package, or drop it. Affects any future SPEC copying the stanza."
audit_gaps_explicitly_not_checked (5, stated so they are not mistaken for passes):
  1: "AC-WBR-012's mutation was NOT independently reproduced. Re-running it requires deleting the `FieldDef` from `internal/settings/schema_sections.go`, a source edit the audit is forbidden from making. The auditor read the implementer's capture (`ac-wbr-012-mutation.txt`), judged it internally consistent with the guard's source, and attributed the claim to the implementer rather than to an independent run."
  2: "Cross-platform. All measurement is darwin/arm64. No Windows or Linux run: `git rev-parse --git-common-dir` path comparison and `git worktree add` operand handling were not exercised there. CI covers this on the PR head."
  3: "Concurrency. Two simultaneous SessionStarts against one repository were not run; the design argument (the second finds a match and no-ops; the primary-checkout gate leaves exactly one writer) was read but not stress-tested."
  4: "The web console was not driven through a browser. The render half is asserted against `renderConsolePage(t)` output; no end-to-end form submit was performed, so the POST → `applyGitStrategyKey` → file path is covered by the round-trip unit test only."
  5: "Inherited G2 — `EnterWorktree`'s actual read of `refs/remotes/origin/HEAD` remains inferred from behaviour rather than read from Claude Code source. Unchanged by this SPEC and explicitly mitigated by the doctor item."
residual_risk:
  - "The four real git-wrapper seams have no standing automated guard (F2); the auditor's A6 runs prove they work today, nothing prevents a silent regression tomorrow."
  - "AC-WBR-014's value-preservation property rests on a key-name assertion (F4) — the property holds, the guard is weaker than the property."
  - "`refs/remotes/origin/HEAD` is repository-global metadata: the primary-checkout gate confines moai's writers to one, but an external actor (another tool, a manual `git remote set-head`, a fresh clone) can still move it between sessions. Inherent to the approach, accepted by design — realignment is idempotent and costs one notice line, and the doctor item surfaces the state."
sync_phase_files: "CHANGELOG.md (1 entry), .moai/specs/SPEC-WORKTREE-BASEREF-001/{spec.md frontmatter, progress.md §E.4}, .moai/reports/t313/{sync-audit.md, sync-summary.md} — no Go, template, or docs file changed at sync (the M6 docs review found nothing to repair)"
push_state: "NOT pushed, no PR opened — integration is the orchestrator's call per the dispatch"
t316_boundary: "no file under `.claude/skills/moai-workflow-project/schemas/` touched (both `tab_schema.json` untouched)"
```
