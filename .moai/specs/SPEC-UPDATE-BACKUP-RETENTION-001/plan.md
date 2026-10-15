# Plan — SPEC-UPDATE-BACKUP-RETENTION-001

Tier L. Milestones are ordered by reversibility: the decisions most likely to change (a shipped configuration key, the retention semantics users see, the report text) come first; mechanical steps come last. Priorities are labels; no durations.

## §A Context

### A.1 Starting state

- Work tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1733`, branch `WT-10-14-class-3`, HEAD `e669da81b3f865488dd40f840623bed69839f459` (clean). It already contains the t1687 and t1799 code (merge `e669da81b`).
- The retention policy (`retention.go`, the `state.backup_past_versions_keep` key, per-item reporting) and the archive / link-removal path were reverted by `8e43f064a`. Nothing of them remains in non-test code (`research.md` § Finding verification).
- The pruner today is the run-count pruner in `internal/cli/update/backup/backup.go:210-252` with the run collection at `:293-328`. It is called from `internal/cli/update_clean_install.go:559` and `internal/cli/update_template_sync.go:909`, both with keepCount 5.
- `moai update` is forbidden in this repository until this card lands. Every measurement is a function-level harness on copies.

### A.2 Assumptions (surface list; correct any that is wrong)

1. The operator's per-file budget ruling (a263 C) still governs; only its template-owned exclusion for `pre-clean/` was withdrawn (fd8b). The budget applies to `pre-clean/` content like any other (REQ-UBR-006); "at least one copy" in the card's completion text is read as the budget floor of 1, not as an unbounded keep.
2. Reintroducing the archive and link removal is in scope (progress.md handoff memo of t1687; board a041). Decision-index Q3 asks the operator to confirm because the committed tree does not carry that authority.
3. The retention unit stays the run (SPEC-UPDATE-REPAIR-RC30-001); the budget and protection operate on runs, never on single directories.
4. Go 1.26.8 `os.Root` (with `Rename`, `Remove`, `Lstat`, `Mkdir`, `MkdirAll`, `Link`, `Symlink`) is available on every supported platform. Confirmed for the toolchain with `go doc os.Root`; Windows behavior of `Root.Link` was not observed.
5. F26 needs no work (fixed by t1799, E9). F20, F22, F23 need no work beyond REQ-UBR-020.
6. F18 and F19 are in scope: both are cheaper to close while the policy is rewritten than to carry as debt (F18 changes the identity of one input to the key; F19 is about the one new constant and two tests).
7. The mo.ai.kr measurement uses a copy of the backup roots, never the live project.

## §B Open Clarifications

These block Implementation Kickoff, not the plan audit. Each maps to a row in `decision-index.md`.

- [NEEDS CLARIFICATION: cap strictness — may protected excess runs exceed keepCount (soft cap, REQ-UBR-010 as drafted) or must the total run count stay at or under keepCount (decision-index Q2)]. The orchestrator's completion wording ("leaves both backup roots at or under their cap") and the budget ruling cannot both hold for every input; AC-UBR-06 is drafted for the soft cap and reports the retained-excess count so the operator can judge Q2 on a measured number (Q4).
- [NEEDS CLARIFICATION: F21 mechanism — keep the zero-byte database removal as is, report-only, or a narrowed removal list (decision-index Q1)]. REQ-UBR-019 fixes only the outcome; AC-UBR-19 passes under (b) and (c) and fails under (a).
- [NEEDS CLARIFICATION: confirm that the retired-skill archive and `.agents/skills` link removal are reintroduced (decision-index Q3)]. If the answer is "report-only", REQ-UBR-015 and REQ-UBR-017 and AC-UBR-05, AC-UBR-15 and AC-UBR-16 collapse to a regression guard and AC-UBR-04 remains.

Not a user question: the retained-excess count on the real roots (Q4) is measured by M6.

## §C Pre-flight (run-phase starts here)

1. `git rev-parse --short HEAD` and `git branch --show-current` in the card tree; confirm `e669da81b` or a descendant and branch `WT-10-14-class-3`.
2. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` (cross-platform build tags: the new code touches only portable `os` APIs; FIFO tests stay behind a unix build tag as `restore_unmanaged_fifo_unix_test.go` does).
3. Baseline for the packages touched: `go test -count=1 ./internal/cli/update/... ./internal/config/...` and `go test -count=1 -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/`, each once, recorded. Do not run the full `internal/cli` suite on a loaded machine.
4. Re-measure the RED-now cells of AC-UBR-01 to AC-UBR-06 on the starting tree with the probe sources in `research.md` Appendix A before changing any code.
5. Environment: `TMPDIR` / `GOTMPDIR` stay on the default temporary area (never the home directory or the tree); never run `moai update`.

## §D Constraints

- PRESERVE: the writers of backup directories, run-marker semantics, the restore paths, `internal/cli/update/deploy` (the `pre-clean/` producer), and the behaviors named in REQ-UBR-013 and REQ-UBR-014.
- Template-First: the new configuration key is edited in `internal/template/templates/.moai/config/sections/state.yaml` and `internal/config`, then `make build` regenerates the embedded files; no direct edit of the local `.moai/config/sections/state.yaml` copy as the source of truth.
- One writer per working tree; commits only on the card branch; no push; commit subjects carry `(card t1733)` and the trailer the repository requires.
- Stage by explicit pathspec in the primary checkout only; in this card tree stage the files the milestone touched.
- No time estimates in any artifact or report.

## §E Self-Verification (run-phase report, each as command plus verbatim output)

E1 AC matrix (24 rows); E2 cross-platform build; E3 coverage `go test -cover ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/...` at or above 85%; E4 subagent-boundary grep (no `AskUserQuestion` in the touched non-test code); E5 `gofmt -l` empty, `go vet` clean, `golangci-lint run` on the touched packages; E6 branch HEAD and commit list (no push); E7 blocker report if any; E8 verbatim RED output for AC-UBR-01 to AC-UBR-05 captured before the fix, each probe test in a commit that precedes its fix commit (AC-UBR-24).

## §F Milestones

### M1 — Retention contract and shipped key (Priority High; the decisions most likely to change)

Scope: the configuration key `state.backup_past_versions_keep` (template, struct, default defined once, validation), the version identity (Q5 default), the budget and cover semantics as a selection step that can be tested without a filesystem, and the report row vocabulary.
Order: commit the RED probes P1, P2, P3 first as tests (each in its own commit, output recorded in progress.md §E.2); then the key; then the selection step with table tests that include the F13 and F14 shapes.
Exit: AC-UBR-07, AC-UBR-10 green; AC-UBR-02 and AC-UBR-03 green against the selection step.

### M2 — Pruner integration and reporting (Priority High)

Scope: wire the selection into the pruner; census with single-read hashing; fail-closed classification (REQ-UBR-008); bookkeeping recognition by writer-produced shape (REQ-UBR-009); `.template-defaults/` exemption; per-run report rows and the unmanaged-entries summary through both call sites; destructive-registry row(s); containment and keepCount 0 / negative regression tests.
Exit: AC-UBR-01, 08, 09, 11, 12, 13, 14, 20 green.

### M3 — Skill-folder residue (Priority High)

Scope: RED probes P4a, P4b first; extend the root refusal to `.agents` and `.agents/skills` (Q6 default); reintroduce the archive move with a per-component real-directory check; link removal limited to the exact link; perform every move and removal through a project-root handle with exclusive-create reservation for temporary names (F24, F25); update the registry row and the pin test for the move site(s).
Exit: AC-UBR-04, 05, 15, 16, 17, 18 green.

### M4 — F21 per the operator's ruling (Priority Medium)

Scope: implement the ruling on decision-index Q1; add the regression guard AC-UBR-19. If the ruling is (a), REQ-UBR-019 is withdrawn by an amendment before M4 starts and this milestone is empty.

### M5 — Mechanical closure (Priority Medium)

Scope: struct-YAML symmetry coverage for the state key and the cli tests asserting the retention report rows on the full-reinstall and template-sync paths (F19 ii, iii); platform guards; `make build` and template tests; coverage, gofmt, vet, lint.
Exit: AC-UBR-21, 22, 23 green.

### M6 — Real-data remeasure and close evidence (Priority High; the card's completion gate)

Scope: make a modification-time-preserving copy of the mo.ai.kr backup roots (read-only on the source), run the harness of AC-UBR-06, write the numbers into progress.md §E.2 (command and verbatim output), and hand the retained-excess count to the operator for Q2 and Q4.
Exit: AC-UBR-06, AC-UBR-24 satisfied.

Dependencies: M1 before M2; M2 before M6; M3 and M4 are independent of M1 and M2 and may run in either order after M1's RED commits; M5 last before M6.

## §G Anti-Patterns

- Treating the audit's "0 of 42 removed" symptom as the present defect; the present defect on this tree is over-deletion (E1, E2, E3, E6).
- Fixing F13 and F14 with a test shape that has only a monotone version history; the probe shapes (a current version that also sits in excess runs, repeated copies) are the point.
- Using a modification-time-resetting copy for the real-data measurement; it changes which unmarked runs are "newest".
- Registering a destructive site by editing the registry row alone while the scan counts zero sites (the F22 shape); the pin test must see the new move calls.
- Matching bookkeeping by name only (F18 shape).

## §H Cross-References

`spec.md` (requirements), `acceptance.md` (24 criteria and Definition of Done), `design.md` (algorithm and options), `research.md` (evidence ledger, probe sources), `decision-index.md` (Q1-Q6). Rules: `.claude/rules/moai/development/verification-completeness.md` (two-cell adoption), `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering attribution).
