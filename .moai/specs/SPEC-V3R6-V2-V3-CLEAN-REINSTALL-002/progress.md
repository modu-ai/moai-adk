# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 — Progress

**Status**: in-progress (run-phase M1 pushed 73e7798ba)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-16
- plan-auditor verdict: PASS
- plan-auditor score: 0.95 (Tier M threshold 0.80)
- skip-eligible: YES (score ≥ 0.90; governs ONLY Phase 1 re-execution — Implementation Kickoff Approval remains mandatory, obtained 2026-07-16)
- probes: P1 GEARS PASS / P2 root-cause PASS / P3 parent-contract non-weakening PASS / P4 Reproduction-First PASS
- blocking defects: none
- SHOULD-FIX: S1 (acceptance.md HARD-5 number collision, cosmetic) — deferred

## §E.2 Run-phase Evidence

### M1 — v3-version negative-override (REQ-CRR-001) — commit 73e7798ba (pushed)

**Scope**: 3 files touched (pathspec-only; unrelated `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` and `llm.yaml`/`README.ko.md` excluded).

| File | Change |
|------|--------|
| `internal/cli/v2_detection.go` | REQ-CRR-001 implementation: `probeVersionSignal` signature `(bool, string)` → `(bool, bool, string)`; `V2Fingerprint.V3VersionConfirmed` field added; `IsV2` aggregation changed from pure disjunction to `!V3VersionConfirmed && (S1 \|\| S2 \|\| S3)` |
| `internal/cli/v2_detection_test.go` | AC-CRR-002 reproduction test `TestDetectV2Fingerprint_V3Override_AC_CRR_002`; 2 aggregation cases updated to `wantIsV2: false` for v3+residue scenarios |
| `internal/cli/update_clean_install_test.go` | `makeScenarioB` fixture: `v3.0.0-rc2` → `v2.16.1` (partial-v2 project, not v3+residue — per AC-CRR-007 + Edge-2 design intent) |

**AC PASS/FAIL matrix (M1-relevant ACs)**:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CRR-001 | PASS | `go test -run TestDetectV2Fingerprint_V3Override ./internal/cli/` | `--- PASS: TestDetectV2Fingerprint_V3Override_AC_CRR_002` — V3VersionConfirmed=true, IsV2=false |
| AC-CRR-002 | PASS | `go test -run TestDetectV2Fingerprint_V3Override ./internal/cli/` | reproduction test: v3.0.0 + .agency/ + deprecated path → IsV2=false (loop-termination contract) |
| AC-CRR-007 | PASS | `go test -run TestRunCleanReinstall_ScenarioB ./internal/cli/` | `--- PASS: TestRunCleanReinstall_ScenarioB` — v2.* project + .agency/ → IsV2=true (clean-reinstall runs); v3+residue → IsV2=false (NOT activated) |

**Full-suite verification**:
- `go test ./internal/cli/... -count=1` → exit 0 (21.1s; all ScenarioA/B/C + v2_detection tests pass)
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./internal/cli/...` → exit 0
- `golangci-lint run ./internal/cli/...` → exit 0 (no NEW issues; 6 pre-existing errcheck in `merge_test.go` untouched)
- Coverage: detectV2Fingerprint 95.0%, probeVersionSignal 80.0% (new v3.* branch covered by 5 test cases; pre-existing gaps in error/parse branches unchanged), probeDeprecatedPathSignal 100.0%

**Pending**: M2-M5 ACs not yet addressed (remaining milestones: M2 update.go integration, M3-M5 clean-reinstall path wiring).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M
- scope: ~5-10 files (internal/cli/update.go, v2_detection.go, update_clean_install.go, update_preserve_inventory.go, update_cleanup.go + tests)
- domain count: 1 (internal/cli clean-reinstall code path)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, sequential dependency between milestones)

**Mode evaluation**:
- Mode 1 (trivial): not selected — semantic regression repair, not a typo
- Mode 2 (background): not selected — write-capable implementation work
- Mode 3 (agent-team): RETIRED — never selected
- Mode 4 (parallel): not selected — single domain, coding-heavy (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): **selected** — single sequential manager-develop per milestone
- Mode 6 (workflow): not selected — scope < ~30 files, not mechanical-uniform

**Decision**: sub-agent (Mode 5)

**Justification**: Coding-heavy single-domain regression repair. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. Milestones M1→M5 have sequential dependencies (M1 fingerprint fix is the highest-irreversibility data-model change that M2-M5 build on). Tier M Section A-E delegation template applies.

**Implementation Kickoff Approval**: obtained 2026-07-16 (user selected "run-phase 진입 (권장)"). cycle_type=tdd (existing v2_detection_test.go / update_clean_install_test.go baseline).
