# Plan — SPEC-UPDATE-BACKUP-RETENTION-001

Tier L. Milestones are ordered by reversibility: the decisions most likely to change (a shipped configuration key, the retention semantics users see, the report text) come first; mechanical steps come last. Priorities are labels; no durations.

## §A Context

### A.1 Starting state

- Work tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1733`, branch `WT-10-14-class-3`. Code identical at `e669da81b3f865488dd40f840623bed69839f459` and at the SPEC commit `3b133bedf`. It already contains the t1687 and t1799 code.
- Baseline on main: main `ec5a4f392` produces the same loss as this tree (`research.md` baseline table). The defect is pre-existing; commit `8e43f064a` removed the pre-revert `retention.go`, which is the reference behavior (design D0).
- The pruner today is the run-count pruner in `internal/cli/update/backup/backup.go:210-252` with the run collection at `:293-328`. It is called from `internal/cli/update_clean_install.go:559` and `internal/cli/update_template_sync.go:909`, both with keepCount 5.
- `moai update` is forbidden in this repository until this card lands. Every measurement is a function-level harness on copies.

### A.2 Assumptions (resolved; listed for the record)

1. The operator's per-file budget ruling (a263 C) governs; its template-owned exclusion for `pre-clean/` was withdrawn (fd8b). The budget applies to `pre-clean/` content like any other (REQ-UBR-006); "at least one copy" is the budget floor of 1.
2. Archive and link removal are not reintroduced (operator ruling Q3, 2026-10-16). The residue passes are report-only.
3. The retention unit stays the run (SPEC-UPDATE-REPAIR-RC30-001).
4. Go 1.26.8 `os.Root` (with `Rename`, `Lstat`, `Mkdir`, `MkdirAll`, `Link`) is available on every supported platform. Windows behavior of `Root.Link` was not observed.
5. F26 needs no work (fixed by t1799, E9). F20, F22, F23 need no work beyond REQ-UBR-019.
6. F18 and F19 are in scope.
7. The mo.ai.kr measurement uses a copy of the backup roots, never the live project.

## §B Open Clarifications

None. The three clarifications that stood here (cap strictness, F21 mechanism, archive reintroduction) were answered by the operator on 2026-10-16 and are recorded in `decision-index.md` Q1-Q3: soft cap; report-only for the zero-byte database; no reintroduction of archive or link removal. The retained-excess count (Q4) is a measured output of AC-UBR-06, not a question.

## §C Pre-flight (run-phase starts here)

1. `git rev-parse --short HEAD` and `git branch --show-current` in the card tree; confirm `3b133bedf` or a descendant and branch `WT-10-14-class-3`.
2. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` (the new code touches only portable `os` APIs; FIFO tests stay behind a unix build tag as `restore_unmanaged_fifo_unix_test.go` does).
3. Baseline for the packages touched: `go test -count=1 ./internal/cli/update/... ./internal/config/...` and `go test -count=1 -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/`, each once, recorded. Do not run the full `internal/cli` suite on a loaded machine.
4. Re-measure the RED-now cells of AC-UBR-01 to AC-UBR-06 on the starting tree with the probe files recreated from `research.md` Appendix A (placement steps: `research.md` § Evidence ledger) before changing any code.
5. Environment: `TMPDIR` / `GOTMPDIR` stay on the default temporary area (never the home directory or the tree); never run `moai update`.

## §D Constraints

- PRESERVE: the writers of backup directories, run-marker semantics, the restore paths, `internal/cli/update/deploy` (the `pre-clean/` producer), and the behaviors named in REQ-UBR-013 and REQ-UBR-014.
- Principle: no behavior that deletes data is added in 3.2. Every requirement was read against it (`research.md` § Principle check).
- Template-First: the new configuration key is edited in `internal/template/templates/.moai/config/sections/state.yaml` and `internal/config`, then `make build` regenerates the embedded files.
- One writer per working tree; commits only on the card branch; no push; commit subjects carry `(card t1733)` and the trailer the repository requires.
- Stage by explicit pathspec in the primary checkout only; in this card tree stage the files the milestone touched.
- No time estimates in any artifact or report.

## §E Self-Verification (run-phase report, each as command plus verbatim output)

E1 AC matrix (24 rows); E2 cross-platform build; E3 coverage `go test -cover ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/...` at or above 85%; E4 subagent-boundary grep (no `AskUserQuestion` in the touched non-test code); E5 `gofmt -l` empty, `go vet` clean, `golangci-lint run` on the touched packages; E6 branch HEAD and commit list (no push); E7 blocker report if any; E8 verbatim RED output for AC-UBR-01 to AC-UBR-03 captured before the fix, each probe test in a commit that precedes its fix commit (AC-UBR-24).

## §F Milestones

### M1 — Retention contract and shipped key (Priority High; the decisions most likely to change)

Scope: the configuration key `state.backup_past_versions_keep` (template, struct, default defined once, validation), the version identity (Q5 default), the budget and cover semantics as a selection step that can be tested without a filesystem (including the exempt-run disposition of REQ-UBR-008), and the report row vocabulary.
Order: commit the RED probes P1, P2, P3 first as tests with assertions (each in its own commit, output recorded in progress.md §E.2); then the key; then the selection step with table tests that include the F13 and F14 shapes.
Exit: AC-UBR-07, AC-UBR-10 green; AC-UBR-02 and AC-UBR-03 green against the selection step.

### M2 — Pruner integration and reporting (Priority High)

Scope: wire the selection into the pruner; census with single-read hashing; fail-closed classification (REQ-UBR-008); bookkeeping recognition by writer-produced shape (REQ-UBR-009); `.template-defaults/` exemption; per-run report rows and the unmanaged-entries summary through both call sites; the destructive-registry row for `CleanupOldBackups`; containment and keepCount 0 / negative regression tests.
Exit: AC-UBR-01, 08, 09, 11, 12, 13, 14, 19 green.

### M3 — Residue passes stay report-only; rename bound to the root (Priority Medium)

Scope: P4a and P4b committed as regression guards first (green on the starting tree); the held-back rename performed through a project-root handle with exclusive-create reservation for temporary names (F24, F25); the registry row and the pin test for the changed move site; a test that no removal or move of a retired folder or link exists (AC-UBR-15).
Exit: AC-UBR-04, 05, 15, 16, 17, 20 green. (AC-UBR-20 is bound here, not to M2, because it covers the rename site this milestone changes; the `CleanupOldBackups` row is AC-UBR-19 at M2.)

### M4 — Empty database file is reported (Priority Medium)

Scope: the state-residue pass appends the "empty DB file present" report row and no longer removes a zero-byte `*.db` (REQ-UBR-018, operator ruling Q1); the `backlog.json.lock` removal is untouched.
Exit: AC-UBR-18 green.

### M5 — Mechanical closure (Priority Medium)

Scope: struct-YAML symmetry coverage for the state key and the cli tests asserting the retention report rows on the full-reinstall and template-sync paths (F19 ii, iii); platform guards; `make build` and template tests; coverage, gofmt, vet, lint.
Exit: AC-UBR-21, 22, 23 green.

### M6 — Real-data remeasure and close evidence (Priority High; the card's completion gate)

Scope: make a modification-time-preserving copy of the mo.ai.kr backup roots (read-only on the source), run the harness of AC-UBR-06, write the numbers into progress.md §E.2 (command and verbatim output).
Exit: AC-UBR-06, AC-UBR-24 satisfied.

Dependencies: M1 before M2; M2 before M6; M3 and M4 are independent of M1 and M2 and may run in either order after M1's RED commits; M5 last before M6. AC-UBR-20 depends only on M3's own change; it does not require M2.

## §G Anti-Patterns

- Treating the audit's "0 of 42 removed" symptom as the present defect; the present defect on this tree is over-deletion (E1, E2, E3, E6), and it is pre-existing on main.
- Describing the loss as introduced by t1687 or "compared with main" in a way the reproduction report refutes.
- Fixing F13 and F14 with a test shape that has only a monotone version history; the probe shapes (a current version that also sits in excess runs, repeated copies) are the point.
- Using a modification-time-resetting copy for the real-data measurement; it changes which unmarked runs are "newest".
- Registering a destructive site by editing the registry row alone while the scan counts zero sites (the F22 shape); the pin test must see the new move calls.
- Matching bookkeeping by name only (F18 shape).
- Reintroducing a removal path in the residue passes (operator ruling Q3).

## §H Cross-References

`spec.md` (requirements), `acceptance.md` (24 criteria and Definition of Done), `design.md` (algorithm, reference behavior, post-landing verification), `research.md` (evidence ledger, probe sources), `decision-index.md` (Q1-Q6). Rules: `.claude/rules/moai/development/verification-completeness.md` (two-cell adoption), `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering attribution).
