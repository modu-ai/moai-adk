# SPEC-V3R2-SPC-004 Progress Tracker

> Phase tracker shell for **@MX anchor resolver (query by SPEC ID, fan_in, danger category)**.
> Updated by run-phase agents during M1..M6 execution.

## HISTORY

| Version | Date       | Author                                    | Description |
|---------|------------|-------------------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B)     | Initial progress.md shell. plan_complete_at populated post-PR-merge. plan_status: audit-ready. |

---

## 1. Phase Status

| Phase   | Status       | Started     | Completed | Notes |
|---------|--------------|-------------|-----------|-------|
| Plan    | completed    | 2026-05-10  | 2026-05-10 | PR #837 admin merged into main |
| Run     | in-progress  | 2026-05-13  | -         | M1+M2+M3+M4 COMPLETE (G-01/G-02/G-03/G-04) |
| Sync    | pending      | -           | -         | Awaits run PR merge |
| Cleanup | pending      | -           | -         | Awaits sync PR merge |

Plan-phase status fields:
- `plan_complete_at`: (set when plan PR merged into main)
- `plan_status`: draft → **audit-ready** → audited → merged

---

## 2. Plan Phase Checkpoints

- [x] spec.md authored (already present at HEAD `73742e3ee`)
- [x] research.md authored (Phase 1A — 45 evidence anchors)
- [x] plan.md authored (Phase 1B — 6 milestones, 20 tasks, 18 REQs traced)
- [x] acceptance.md authored (15 ACs with Given/When/Then; 11 covered by existing tests)
- [x] tasks.md authored (20 tasks across M1..M6, no wave-split required)
- [x] progress.md shell authored
- [x] issue-body.md authored (PR body draft)
- [ ] plan-auditor PASS verdict (target ≥ 0.85 first iteration)
- [ ] plan PR squash-merged into main
- [ ] plan_complete_at + plan_status = audit-ready persisted

---

## 3. Milestone Tracker (Run Phase)

### M1: LSP `find-references` 통합 (G-01) — Priority P0 ✅ COMPLETE

- [x] T-SPC004-01: LSPFanInCounter struct + interface implementation
  - commit: 25a9f6594 (GREEN), 8c439a4ac (REFACTOR)
  - `internal/mx/fanin_lsp.go` (135 LOC)
- [x] T-SPC004-02: LSPFanInCounter RED tests (7 sub-tests) + strictMode 강화
  - commit: 7237ae2bc (RED)
  - `internal/mx/fanin_lsp_test.go` (246 LOC)

M1 Status: COMPLETE (2026-05-13)
Commits: RED `7237ae2bc`, GREEN `25a9f6594`, REFACTOR `8c439a4ac`

Verification gate: `go test ./internal/mx/ -run "TestLSPFanInCounter|TestResolver_AC9_StrictMode"` → all PASS.
Coverage: 88.8% (목표 85% 초과). Race detector: PASS.

### M2: `mx.yaml` `danger_categories:` user wire-up (G-02) — Priority P1 ✅ COMPLETE

- [x] T-SPC004-03: LoadDangerConfig helper + 2 RED tests
  - commit: 94586497f (RED), 6534a2097 (GREEN)
  - `internal/mx/danger_category.go` +24 LOC (LoadDangerConfig)
- [x] T-SPC004-06: validateQuery danger 분기 + 1 RED test
  - commit: 94586497f (RED), 6534a2097 (GREEN)
  - `internal/mx/resolver_query.go` +23 LOC (danger validation branch)

M2 Status: COMPLETE (2026-05-13)
Commits: RED `94586497f`, GREEN `6534a2097`

Verification gate: `go test -race ./internal/mx/... -count=1` → all PASS.
Coverage: 88.4% (M1 end: 88.8% — slight decrease from new LoadDangerConfig statements; >85% target met).
Race detector: PASS. go vet: PASS.

### M3: `.moai/specs/*/spec.md` `module:` 자동 로드 (G-03) — Priority P1 ✅ COMPLETE

- [x] T-SPC004-04: spec_loader.go LoadSpecModules helper
  - commit: fe57d9107 (GREEN)
  - `internal/mx/spec_loader.go` (126 LOC)
- [x] T-SPC004-05: spec_loader RED tests (5 cases) + SpecAssociator integration
  - commit: adfa3a53a (RED), fe57d9107 (GREEN)
  - `internal/mx/spec_loader_test.go` (161 LOC)

M3 Status: COMPLETE (2026-05-13)
Commits: RED `adfa3a53a`, GREEN `fe57d9107` (REFACTOR skipped: clean implementation)

Verification gate: `go test ./internal/mx/ -run "TestLoadSpecModules|TestSpecAssociator_PathBased_FromLoader"` → all PASS.
Coverage: 88.0% (stable vs M2 88.4%). Race detector: PASS. go vet: PASS.

### M4: `Resolver.ResolveAnchorCallsites()` API parity (G-04) — Priority P1 ✅ COMPLETE

- [x] T-SPC004-10: Callsite struct + helpers
  - commit: d41fccfad (GREEN)
  - `internal/mx/callsite.go` (20 LOC)
- [x] T-SPC004-11: ResolveAnchorCallsites + 5 RED tests (LSP/textual/exclude-tests/backward-compat/not-found)
  - commit: 7b032e38b (RED), d41fccfad (GREEN)
  - `internal/mx/callsite_test.go` (229 LOC)
  - `internal/mx/resolver.go` +117 LOC (ResolveAnchorCallsites, resolveCallsitesTextual, callerLinesInFile)

M4 Status: COMPLETE (2026-05-13)
Commits: RED `7b032e38b`, GREEN `d41fccfad` (REFACTOR skipped: coverage 87.9% ≥ 85%; uncovered branch is error-resilience only)

Verification gate: `go test ./internal/mx/ -run "TestResolver_ResolveAnchorCallsites|TestResolver_ResolveAnchor_BackwardCompat"` → all PASS.
Coverage: 87.9% (M3 end: 88.0% — -0.1% from uncovered walkErr path; >85% target met). Race detector: PASS. go vet: PASS.

### M5: test_paths glob (G-05) + stderr verify (G-06) — Priority P1

- [ ] T-SPC004-07: isTestFile glob 패턴 wire-up + TextualFanInCounter 확장
- [ ] T-SPC004-08: isTestFile RED tests (3 sub-tests) + integration
- [ ] T-SPC004-09: stderr format regression fixture (G-06)

Verification gate: `go test ./internal/mx/ ./internal/cli/ -run "TestIsTestFile_UserPattern|TestTextualFanInCounter_RespectsUserTestPaths|TestSidecarUnavailable_StderrFormat"` → all PASS.

### M6: 16-언어 sweep (G-08) + benchmark (G-07) + verification — Priority P0

- [ ] T-SPC004-12: CLI wire-up — LoadDangerConfig + LoadSpecModules + LSP detect
- [ ] T-SPC004-13: Performance benchmark fixtures (advisory)
- [ ] T-SPC004-15: 16-language sweep test
- [ ] T-SPC004-14: `go test -race -count=1 ./...` PASS
- [ ] T-SPC004-16: `golangci-lint run` clean
- [ ] T-SPC004-17: `make build` exits 0; embedded.go regenerated; `diff -r` clean
- [ ] T-SPC004-18: CHANGELOG.md updated (4 entries)
- [ ] T-SPC004-19: @MX tags applied per plan §6 (6 tags)
- [ ] T-SPC004-20: end-to-end manual verification (real LSP + project)

Verification gate: All AC-SPC-004-01..15 verified per acceptance.md.

---

## 4. AC Status Tracker

| AC ID | Status | Verified by | Existing test? | Verified at |
|---|---|---|---|---|
| AC-01 | partial | T-SPC004-04+05 DONE (path loader+associator); T-SPC004-12 (M6 CLI wire-up pending) | YES | 2026-05-13 M3 |
| AC-02 | verified+ | T-SPC004-01+02 (fan_in count); T-SPC004-10+11 (callsite locations, M4) | YES | 2026-05-13 M4 |
| AC-03 | partial | T-SPC004-03, T-SPC004-06 (M2 DONE); T-SPC004-12 (M6 CLI wire-up pending) | YES | 2026-05-13 M2 |
| AC-04 | pending | T-SPC004-09, T-SPC004-12 | PARTIAL → fixture added | (TBD) |
| AC-05 | pending | T-SPC004-15 (sweep) | YES | (TBD) |
| AC-06 | pending | (existing) | YES | (TBD) |
| AC-07 | verified | T-SPC004-01, T-SPC004-02 | YES | 2026-05-13 M1 |
| AC-08 | pending | T-SPC004-13 | YES | (TBD) |
| AC-09 | verified | T-SPC004-02 (LSP-detect path) | strictMode 강화 완료 | 2026-05-13 M1 |
| AC-10 | pending | (existing) | YES | (TBD) |
| AC-11 | pending | T-SPC004-07, T-SPC004-08 | YES → user_paths added | (TBD) |
| AC-12 | pending | (existing) | YES | (TBD) |
| AC-13 | partial | T-SPC004-06 (danger InvalidQuery done); T-SPC004-12 (CLI exit-2 pending M6) | PARTIAL → validateQuery danger branch done | 2026-05-13 M2 |
| AC-14 | pending | (existing) | YES | (TBD) |
| AC-15 | partial | T-SPC004-04+05 DONE (path-based associator + body-based both active); T-SPC004-15 (M6 sweep pending) | YES | 2026-05-13 M3 |

---

## 5. Iteration Tracker (Re-planning Gate)

Per `.claude/rules/moai/workflow/spec-workflow.md` § Re-planning Gate, append per-iteration acceptance criteria completion count + error count delta to detect stagnation (3+ iterations no AC progress → trigger AskUserQuestion gap analysis).

| Iteration | Date | AC completed | Error delta | Notes |
|---|---|---|---|---|
| M1        | 2026-05-13 | 3/15 (AC-02, AC-07, AC-09) | 0 | RED 7237ae2bc → GREEN 25a9f6594 → REFACTOR 8c439a4ac |
| M2        | 2026-05-13 | +2 partial (AC-03, AC-13) | 0 | RED 94586497f → GREEN 6534a2097 (REFACTOR skipped: no cleanup opportunity) |
| M3        | 2026-05-13 | +2 partial (AC-01, AC-15) | 0 | RED adfa3a53a → GREEN fe57d9107 (REFACTOR skipped: clean implementation) |
| M4        | 2026-05-13 | AC-02 → verified+ (Callsite location list) | 0 | RED 7b032e38b → GREEN d41fccfad (REFACTOR skipped: 87.9% ≥ 85%) |

---

## 6. Risk Watch

Per spec §8 + plan §7 + research §10:

| Risk | Status | Mitigation |
|---|---|---|
| SPC-004 already implemented 90%+ — REQ redundancy concern | MITIGATED | plan §1.2.1 acknowledged; sync HISTORY reconcile (Expose → complete to parity) |
| SPC-002 schema break risk | MITIGATED | schema_version: 2 preserved; no sidecar field added by this SPEC; read-only consumer |
| LSP server unavailable for 16-lang on user host | MITIGATED | Silent fallback to TextualFanInCounter + annotation; strictMode opt-in |
| `Resolver.ResolveAnchor()` signature mismatch with spec.md | MITIGATED | G-04 additive `ResolveAnchorCallsites()`; existing API unchanged |
| Performance regression on 50-anchor LSP path | MONITORED | T-SPC004-13 advisory benchmark |
| isTestFile user pattern over-match | MITIGATED | hard-coded fallback preserved; user pattern is additive |
| `validateQuery` dangerMatcher nil panic | MITIGATED | T-SPC004-06 default matcher fallback inside validateQuery |
| spec_loader.go yaml parse failure | MITIGATED | graceful skip per file (T-SPC004-04); empty map on missing dir |

---

## 7. Notes

- Run-phase methodology: TDD per `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- Worktree base: `origin/main` HEAD `73742e3ee` (plan-phase author baseline; SPC-002 plan PR #836 merge commit).
- Template-first discipline: this SPEC touches `internal/` Go code only; no template asset changes expected. `make build` regenerates `embedded.go` for safety.
- Estimated artifacts: 7 new Go source files + extensions to 4 existing test files + 5 modified Go files + CHANGELOG = ~1,400 LOC delta (production ~480 + test ~920).
- 90%+ of code surface already merged via SPC-004 PR #746 (`68795dbe3`); this SPEC closes 8 격차 (G-01..G-08) per research.md §12.
- 11 of 15 ACs already covered by existing tests; 4 are PARTIAL/NEW (AC-04 stderr fixture, AC-09 LSP-detect, AC-11 test_paths glob, AC-13 exit code 2).

---

End of progress.

Version: 0.1.0
Status: Progress shell for SPEC-V3R2-SPC-004
