---
id: SPEC-JEV-GUARD-001
title: "plan — Jev Consumer B withdrawal (consumer-guard contract restoration)"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
tier: S
---

# plan.md — SPEC-JEV-GUARD-001

## A. Context

- Card t1083 (Class C), worktree `.claude/worktrees/t1083`, branch `WT-jev-guard-green`, HEAD `cd99336bf` (local develop).
- RED: `TestNoConsumerCallPathShips` (observed this session, exit 1, SkillSuggest marker only). Cause chain inherited from t1066 — no re-diagnosis (card [HARD] #1).
- Resolution axis chosen: **withdraw** the Consumer B (SkillSuggest) family. The guard contract ("consumers ship only after measurement") wins because measurement is blocked upstream (t1066 F2 wire-format defect, separate card).
- Harness: minimal. Run-phase mode: serial. Tier S — 2 milestones.

## B. Known Issues

| Issue | Disposition |
|---|---|
| Dispatch premise "SKILL.md copies are byte-identical" is false on this tree | Measured pre-existing divergence: 21 changed lines / 41 content lines / 4 categories (spec.md §D.3, corrected per plan-audit D1); identity criterion rescoped to prose absence + shape-pinned no-widening (AC-JEVG-004) |
| `.moai/project/codemaps/` 4 files will go stale after withdrawal | Sync-phase obligation (manager-docs, `moai codemaps` regeneration) — run-phase does not touch them |
| CONSUMERS SPEC REQ-JEVN-016 left with no live instance | Documented, not amended — decision (a), rationale spec.md §D.5 |

## C. Pre-flight (run-phase entry checks)

1. `git rev-parse --short HEAD` → `cd99336bf` expected; if moved, stop and re-read (one-writer rule).
2. `git status --porcelain` → clean expected (artifacts committed by the lane orchestrator before run entry).
3. Confirm RED still reproduces: `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips` → exit 1, SkillSuggest hit. If not red, stop — the cause chain assumption is void.
4. Confirm guard test file SHA matches the plan-phase tree: `git diff cd99336bf -- internal/jevmeasure/gate_demo_test.go` → empty.

## D. Constraints

- [HARD] Zero edits to `internal/jevmeasure/gate_demo_test.go` and any test asserting the guard.
- [HARD] Go diff limited to the Consumer B family (`jev_skill_suggest.go`, `jev_skill_suggest_test.go`, `root.go:218`).
- [HARD] No suppression by any mechanism: weakening / re-scoping / marker edit / relocation / build tags / skip logic.
- [HARD] Do NOT run `go test ./...` locally (lane-local verification only; full suite is CI's).
- [HARD] Do not touch t1068's axis or the codemap files.
- Template-First: template SKILL.md edit → `make build` (embedded FS regeneration).
- `make build` carries `agents-emit-check` (read-only) — SKILL.md is not an agent `.md`, so no `make agents-emit` is owed; the check passing is expected and recorded.

## E. Self-Verification (run-phase evidence plan)

Full-AC machine-verifiable surface (all on this worktree, after M1; every AC of acceptance.md is covered):

| # | AC | Check | Pass criterion |
|---|----|-------|----------------|
| a | AC-JEVG-001 | `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips` | exit 0 (RED→GREEN flip observed in M2) |
| b | AC-JEVG-005a | `go build ./...` | exit 0 |
| c | AC-JEVG-005b | `go test ./internal/jevmeasure/... ./internal/cli/...` | exit 0 |
| d | AC-JEVG-002 | `grep -rn 'SkillSuggest' --include='*.go' internal/ cmd/ \| grep -v _test` | 0 hits (count is the evidence) |
| e | AC-JEVG-004 | (a)(b) `grep -c 'jev-suggest'` = 0 on both SKILL.md copies; (c) the normalized no-widening check of acceptance.md §D.4 (sed-collapsed diff, shape assertion prints 0) | all three hold |
| f | AC-JEVG-007 | `golangci-lint run ./internal/cli/... ./internal/jevmeasure/...` | clean |
| g | AC-JEVG-003 | `git diff <pre-M1-SHA> -- internal/jevmeasure/gate_demo_test.go` | empty (guard byte-identical) |
| h | AC-JEVG-005c | `go build -o /tmp/t1083-bin ./cmd/moai && /tmp/t1083-bin jev-suggest --help` | unknown-command (non-zero exit, no jev-suggest usage); scratch binary cleaned after |
| i | AC-JEVG-006 | `grep -n 'func jevNotice' internal/cli/todo_jev_finding.go` + `grep -n 'func jevEnabled' internal/cli/doctor_jev.go` | both definitions present (behavioral backup: row c) |
| j | AC-JEVG-008 | spec.md §B REQ-JEVG-006 + §F chain present (read check); backstop = rows a/d | contract recorded + guard live |

Evidence: write every command + verbatim output + exit code into `.moai/specs/SPEC-JEV-GUARD-001/progress.md` §E.2 (run-phase), citing the post-M1 tree SHA.

## F. Milestones

### M1 — Consumer B withdrawal (Priority High)

1. `git rm internal/cli/jev_skill_suggest.go internal/cli/jev_skill_suggest_test.go`
2. Remove `rootCmd.AddCommand(newJevSuggestCmd())` from `internal/cli/root.go:218` (one line, nothing else in root.go).
3. Remove the `### Skill Suggestion (gated — default off)` subsection (lines 115-123: heading + intro paragraphs + 3-item list) from `.claude/skills/moai/SKILL.md`.
4. Apply the identical removal to `internal/template/templates/.claude/skills/moai/SKILL.md` (same line range; the two edited regions must match each other).
5. Run `make build` — embedded FS regeneration (Template-First).
6. Expected diff footprint: 2 deletions + 2 one-region edits. Nothing else.
7. **Commit guidance (restoration pointer, plan-audit D5)** — the deletion commit's message BODY carries a restoration pointer, since the deletion commit is the only marker a future re-lander will find by `git log --follow -- internal/cli/jev_skill_suggest.go`: name REQ-JEVG-006 (restoration contract, this SPEC), the three preconditions (t1066 F2 repair / measurement gate run / baseline beaten), and that re-landing requires a successor SPEC owning the guard-mechanism edit. One-traceability chain, two carriers: spec.md §F (SPEC side) + this commit body (git side).

### M2 — Verification + evidence (Priority High; complete A, then start B)

1. Run the full-AC 10-row surface (a-j) of plan §E; record every command + verbatim output + exit code + tree SHA in `progress.md` §E.2.
2. RED→GREEN pair: cite the plan-phase RED observation (exit 1, `cd99336bf`) and the post-M1 GREEN (exit 0) in the same evidence block.
3. `## @MX Tag Report` section in the run report: 1 `@MX:NOTE` removed with file deletion; zero added/updated.
4. Populate `progress.md` §E.3 (run-phase audit-ready signal). Commit(s) staged by explicit pathspec; commit subject per the transition matrix (`fix(SPEC-JEV-GUARD-001): M1 ...`).
5. Note for the lane: sync-phase carries the codemap regeneration obligation (manager-docs).

## G. Anti-Patterns (prohibited in run-phase)

- Editing the guard test "just to keep it meaningful" — the test is already meaningful; the tree violated it.
- Re-adding the consumer behind a renamed symbol or a new directory — marker/relocation games.
- "While I'm in root.go" cleanups of unrelated registrations — scope discipline.
- Forcing whole-file byte-identity of the two SKILL.md copies — would touch the 21 pre-existing divergent lines across 4 baseline categories (spec.md §D.3) — axis violation.
- "Normalizing" either SKILL.md copy to erase the pre-existing baseline divergence (e.g., rewriting `${CLAUDE_SKILL_DIR}` lines or the `moai cg/cc` line while "in there") — the AC-JEVG-004 check normalizes the COMPARISON, never the files; the 21 baseline lines are out of this card's scope by design.
- Deleting the codemap lines in run-phase — sync-phase surface.

## H. Cross-References

- spec.md §D (scope inventory + rejected alternatives + SKILL.md baseline), §F (traceability chain)
- research.md — full observation log with commands and exit codes
- `internal/jevmeasure/gate_demo_test.go:97-158` — the guard contract being restored
