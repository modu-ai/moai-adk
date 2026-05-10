# SPEC-V3R2-SPC-002 Progress Tracker

> Phase tracker shell for **@MX TAG v2 with hook JSON integration and sidecar index**.
> Updated by run-phase agents during M1..M6 execution.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial progress.md shell. plan_complete_at populated post-PR-merge. plan_status: audit-ready. |

---

## 1. Phase Status

| Phase   | Status       | Started     | Completed | Notes |
|---------|--------------|-------------|-----------|-------|
| Plan    | in-progress  | 2026-05-10  | (TBD)     | plan/SPEC-V3R2-SPC-002 branch open; PR pending squash-merge |
| Run     | pending      | -           | -         | Awaits plan PR merge + worktree setup |
| Sync    | pending      | -           | -         | Awaits run PR merge |
| Cleanup | pending      | -           | -         | Awaits sync PR merge |

Plan-phase status fields:
- `plan_complete_at`: (set when plan PR merged into main)
- `plan_status`: draft → **audit-ready** → audited → merged

---

## 2. Plan Phase Checkpoints

- [x] spec.md authored (already present at HEAD `fcb486c87`)
- [x] research.md authored (Phase 1A — 30 evidence anchors)
- [x] plan.md authored (Phase 1B — 6 milestones, 22 tasks, 22 REQs traced)
- [x] acceptance.md authored (15 ACs with Given/When/Then)
- [x] tasks.md authored (22 tasks across M1..M6)
- [x] progress.md shell authored
- [x] issue-body.md authored (PR body draft)
- [ ] plan-auditor PASS verdict (target ≥ 0.85 first iteration)
- [ ] plan PR squash-merged into main
- [ ] plan_complete_at + plan_status = audit-ready persisted

---

## 3. Milestone Tracker (Run Phase)

### M1: PostToolUse handler + types — Priority P0

- [ ] T-SPC002-01: post_tool_mx.go handler (4 sub-tests)
- [ ] T-SPC002-02: HookSpecificOutput.MxTags field
- [ ] T-SPC002-03: formatTagsForContext + budget cap (3 sub-tests)
- [ ] T-SPC002-15: 16-language fixture

Verification gate: `go test ./internal/hook/ ./internal/mx/ -run "TestPostToolUseHandler|TestHookSpecificOutput_MxTagsField|TestScanner_AllSixteenLanguages"` → all PASS.

### M2: `/moai mx` flag dispatcher — Priority P0

- [ ] T-SPC002-04: `--full` flag (3 sub-tests)
- [ ] T-SPC002-05: `--json` flag (2 sub-tests)
- [ ] T-SPC002-06: `--anchor-audit` flag scaffold (3 sub-tests; full wire in M5)

Verification gate: `go test ./internal/cli/ -run TestMxCmd` → all PASS.

### M3: silent env + mx.yaml ignore — Priority P1

- [ ] T-SPC002-07: MOAI_MX_HOOK_SILENT env (3 sub-tests)
- [ ] T-SPC002-08: mx.yaml config loader + Scanner ignore wire-up (4 sub-tests)

Verification gate: `go test ./internal/hook/ ./internal/mx/ -run "TestPostToolUseHandler_HookSilent|TestLoadConfig|TestScanner_RespectsMxYaml"` → all PASS.

### M4: Scanner correctness fixtures — Priority P1

- [ ] T-SPC002-09: MissingReasonForWarn 3-line lookahead (3 sub-tests)
- [ ] T-SPC002-10: DuplicateAnchorID refuse-write (2 sub-tests)
- [ ] T-SPC002-11: Corrupt sidecar repair suggestion (2 sub-tests)
- [ ] T-SPC002-16: HookSpecificOutput mismatch validator (3 sub-tests)

Verification gate: `go test ./internal/mx/ ./internal/hook/ -run "TestScanner_MissingReason|TestScanner_DuplicateAnchor|TestSidecar_Corrupt|TestHookOutput_MxTagsWith"` → all PASS.

### M5: Archive sweep + atomic verify — Priority P1

- [ ] T-SPC002-12: Atomic write no-partial-reads (race fixture)
- [ ] T-SPC002-13: 7-day stale preservation
- [ ] T-SPC002-14: 8-day stale archive sweep (2 sub-tests)
- [ ] T-SPC002-06 (continued): anchor-audit fully wired with fanin.CountFanIn

Verification gate: `go test ./internal/mx/ -run "TestSidecar_Atomic|TestSidecar_Stale|TestRunAnchorAudit" -race -count=10` → all PASS.

### M6: Verification + audit — Priority P0

- [ ] T-SPC002-17: `go test -race -count=1 ./...` PASS
- [ ] T-SPC002-18: `golangci-lint run` clean
- [ ] T-SPC002-19: `make build` exits 0; embedded.go regenerated; `diff -r` clean
- [ ] T-SPC002-20: CHANGELOG.md updated (4 entries)
- [ ] T-SPC002-21: @MX tags applied per plan §6 (6 tags)
- [ ] T-SPC002-22: end-to-end smoke test (real claude session)

Verification gate: All AC-SPC-002-01..15 verified per acceptance.md.

---

## 4. AC Status Tracker

| AC ID | Status | Verified by | Verified at |
|---|---|---|---|
| AC-01 | pending | T-SPC002-15 | (TBD) |
| AC-02 | pending | T-SPC002-04, T-SPC002-12 | (TBD) |
| AC-03 | pending | T-SPC002-12 | (TBD) |
| AC-04 | pending | T-SPC002-01..03 | (TBD) |
| AC-05 | pending | T-SPC002-09 | (TBD) |
| AC-06 | pending | T-SPC002-10 | (TBD) |
| AC-07 | pending | T-SPC002-13 | (TBD) |
| AC-08 | pending | T-SPC002-14 | (TBD) |
| AC-09 | pending | T-SPC002-11 | (TBD) |
| AC-10 | pending | T-SPC002-08 | (TBD) |
| AC-11 | pending | T-SPC002-07 | (TBD) |
| AC-12 | pending | T-SPC002-05 | (TBD) |
| AC-13 | pending | T-SPC002-16 | (TBD) |
| AC-14 | pending | T-SPC002-06 | (TBD) |
| AC-15 | pending | T-SPC002-15 | (TBD) |

---

## 5. Iteration Tracker (Re-planning Gate)

Per `.claude/rules/moai/workflow/spec-workflow.md` § Re-planning Gate, append per-iteration acceptance criteria completion count + error count delta to detect stagnation (3+ iterations no AC progress → trigger AskUserQuestion gap analysis).

| Iteration | Date | AC completed | Error delta | Notes |
|---|---|---|---|---|
| (TBD)     | (TBD) | 0/15        | 0           | First run-phase iteration baseline |

---

## 6. Risk Watch

Per spec §8 + plan §7 + research §7:

| Risk | Status | Mitigation |
|---|---|---|
| Wave 3 already implemented 80% — REQ redundancy concern | MITIGATED | plan §1.2.1 acknowledged; sync HISTORY reconcile |
| SPC-004 schema break risk | MITIGATED | schema_version: 2 preserved; new fields omitempty only |
| PostToolUse handler crashes Claude session | MONITORED | T-SPC002-01 graceful no-op + race detector (T-SPC002-12) |
| Token budget overflow | MITIGATED | T-SPC002-03 (budget cap) + T-SPC002-07 (silent env) |
| Atomic write race | MITIGATED | sidecar Manager has sync.RWMutex (research [E-21]) + T-SPC002-12 fixture |
| Cross-language fixture drift | MITIGATED | T-SPC002-15 16-lang regression guard |
| HookSpecificOutput mismatch | MITIGATED | T-SPC002-16 validator + sentinel error |

---

## 7. Notes

- Run-phase methodology: TDD per `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- Worktree base: `origin/main` HEAD `fcb486c87` (plan-phase author baseline).
- Template-first discipline: this SPEC touches `internal/` Go code only; no template asset changes expected. `make build` regenerates `embedded.go` for safety.
- Estimated artifacts: 5 new Go source files + 5 new test files + 4 modified Go files + CHANGELOG = ~1,250 LOC delta (production ~480 + test ~770).
- 80% of code surface already merged via Wave 3 PR #741 (`3f0933550`) + SPC-004 PR #746 (`68795dbe3`); this SPEC closes 8 격차 (G-01..G-08) per research.md §10.

---

End of progress.

Version: 0.1.0
Status: Progress shell for SPEC-V3R2-SPC-002
