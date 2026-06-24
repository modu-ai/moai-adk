# Acceptance Criteria — SPEC-STOP-EVIDENCE-GATE-001

> GEARS-format AC. Each criterion is verifiable by a concrete shell / grep / go-test command an auditor can run from the repo root.
> Status: draft. cycle_type=tdd — each AC is RED before its milestone and GREEN after.

## D. AC Matrix

| AC | Maps to REQ | Severity | Summary |
|----|-------------|----------|---------|
| AC-SEG-001 | REQ-SEG-001 | MUST | session-ledger reader uses `LoadBySession`; NO new on-disk store file created |
| AC-SEG-002 | REQ-SEG-002 | MUST | binary evidence preferred — evaluator keys on `IsTestPass`/`IsTestFail`, not `Outcome` alone |
| AC-SEG-003 | REQ-SEG-003 | MUST | docs-only path-kind exempt — docs-only session with no test-pass is NOT flagged |
| AC-SEG-004 | REQ-SEG-004 | MUST | code-change + success claim + no binary pass → advisory finding emitted (the L_manager_docs_false_backfill_report shape) |
| AC-SEG-005 | REQ-SEG-005 | MUST | fail-open — `Handle()` returns `&HookOutput{}` (allow) on every path, finding or not |
| AC-SEG-006 | REQ-SEG-006 | MUST | advisory → stderr only; stdout HookOutput JSON contract unchanged |
| AC-SEG-007 | REQ-SEG-007 | MUST | C-HRA-008 — `grep AskUserQuestion\|mcp__askuser internal/hook/` (non-test, non-comment) = 0 matches |
| AC-SEG-008 | REQ-SEG-008, 010 | MUST | ≤5s read-only — gate calls only `LoadBySession`; legacy fixture (absent new fields) NOT false-flagged |
| AC-SEG-009 | REQ-SEG-009 | MUST | behavior preservation — StopHookActive guard / 90d·30d pruning / AnalyzeSessionAndLog unchanged |
| AC-SEG-010 | REQ-SEG-010 | MUST | graceful degradation — omitempty round-trip; absent binary fields ≠ "failed" |
| AC-SEG-011 | REQ-SEG-011 | MUST | path-kind taxonomy — docs-only / code-change / unknown buckets; ambiguous → unknown (no finding) |
| (REQ-SEG-012) | REQ-SEG-012 | — (dependency-declaration) | No run-phase AC — falsifiable-value / dormant-scaffold framing requirement; dormancy witnessed by the AC-SEG-008 legacy/dormant fixture; successor writer = `SPEC-STOP-EVIDENCE-WRITER-001` (spec.md §A.3/§A.4) |

---

## AC-SEG-001 — session-ledger reader built on LoadBySession, no new store

| Field | Value |
|-------|-------|
| **Given** | M2 is complete and `internal/hook/session_ledger.go` exists with `buildSessionLedger`. |
| **When** | An auditor inspects the ledger reader's data source and checks that no new storage file is written by the gate path. |
| **Then** | `buildSessionLedger` consumes a `[]telemetry.UsageRecord` produced by `telemetry.LoadBySession` (no other read path), and the gate path opens no new file for writing. |
| **Evidence** | `grep -c 'LoadBySession' internal/hook/session_ledger.go internal/hook/stop.go` returns ≥ 1 (reader reuses LoadBySession). AND `grep -cE 'os\.Create|os\.OpenFile|os\.WriteFile|os\.MkdirAll' internal/hook/session_ledger.go` returns `0` (reader writes no file — bare `-E` alternation `|`, dots escaped `os\.`). AND `grep -c 'func buildSessionLedger' internal/hook/session_ledger.go` returns ≥ 1. |
| **Pass criterion** | LoadBySession referenced ≥ 1 in the gate/reader path AND zero file-write primitives in `session_ledger.go` AND `buildSessionLedger` accepts `[]telemetry.UsageRecord` (verify signature via `grep 'func buildSessionLedger' internal/hook/session_ledger.go` shows a `[]telemetry.UsageRecord` parameter). |

---

## AC-SEG-002 — binary evidence preferred over soft signal

| Field | Value |
|-------|-------|
| **Given** | M2 is complete; `evaluateEvidence` is implemented. |
| **When** | The evaluator is given a code-change session with `Outcome=success` but no `IsTestPass`, vs one with `IsTestPass=true`. |
| **Then** | The evaluator's decision keys on the binary `IsTestPass`/`IsTestFail` signals (not on `Outcome` alone): a `success` Outcome alone does NOT count as "backed"; an observed `IsTestPass=true` does count. |
| **Evidence** | `go test ./internal/hook/ -run 'TestEvaluateEvidence' -v -count=1` shows TWO passing subtests: (1) `success_no_pass_binary_present → finding` (a `success` Outcome with a binary signal observed but no `IsTestPass` yields a finding), and (2) `success_test_pass_observed → nil_finding` (the same `success` Outcome with `IsTestPass=true` yields NO finding). The behavioral delta between (1) and (2) is the binary signal alone — proving the decision pivots on `IsTestPass`/`IsTestFail`, not on `Outcome`. |
| **Pass criterion** | Both subtests pass: flipping only `IsTestPass` (false→true) flips the verdict (finding→nil) while `Outcome=success` is held constant — demonstrating binary-evidence-first per REQ-SEG-002. (No grep token-count proxy: a `grep -c 'IsTestPass'` is satisfiable by dead code and is NOT used as evidence here; the behavioral subtest delta is the proof.) |

---

## AC-SEG-003 — docs-only path-kind exempt

| Field | Value |
|-------|-------|
| **Given** | M2 is complete; `inferPathKind` + `evaluateEvidence` implemented. |
| **When** | The evaluator is given a docs-only session (path-kind = `docs-only`, e.g. `Phase=sync` / `AgentType=manager-docs`) that has a `success` claim but NO binary test-pass signal. |
| **Then** | The evaluator returns no finding (`nil`) — a docs-only session is NOT required to show binary test-pass evidence. |
| **Evidence** | `go test ./internal/hook/ -run 'TestEvaluateEvidence' -v -count=1` shows a passing subtest `docs_only_success_no_test_pass → nil_finding`. AND a direct ledger test: a record with `Phase=sync` (no PathKind) classifies as `docs-only` (`TestInferPathKind` subtest `phase_sync → docs-only`). |
| **Pass criterion** | The docs-only-exempt subtest passes (no finding for docs-only + success + no test-pass) AND `inferPathKind` classifies `Phase=sync` / `AgentType=manager-docs` as `docs-only` per REQ-SEG-003. |

---

## AC-SEG-004 — code-change unbacked success → advisory finding

| Field | Value |
|-------|-------|
| **Given** | M2 is complete; `evaluateEvidence` implemented. |
| **When** | The evaluator is given a code-change session (path-kind = `code-change`) with a `success` Outcome claim AND a binary signal observed (e.g. `IsTestFail=true`) but NO `IsTestPass=true`. |
| **Then** | The evaluator returns a non-nil `Finding` (advisory), capturing the "success claimed but no observed pass" shape that `L_manager_docs_false_backfill_report` represents. |
| **Evidence** | `go test ./internal/hook/ -run 'TestEvaluateEvidence' -v -count=1` shows a passing subtest `code_change_success_no_pass_binary_present → finding` returning a non-nil Finding. AND `Finding.HumanReadable()` includes the strings `path-kind` and a success-claim count (verify via the test asserting on the human-readable output). |
| **Pass criterion** | The unbacked-claim subtest returns a non-nil Finding for a code-change session with success + observed binary + no pass, AND the Finding's human-readable form names the path-kind and success-claim count per REQ-SEG-004 + design §0.5. |

---

## AC-SEG-005 — fail-open (Handle always returns allow)

| Field | Value |
|-------|-------|
| **Given** | M3 is complete; `runEvidenceGate` wired into `Handle()`. |
| **When** | `stopHandler.Handle()` runs against (a) a session that produces a finding, (b) a session that produces no finding, (c) a session with `StopHookActive=true`. |
| **Then** | In all three cases `Handle()` returns `&HookOutput{}` (the empty allow output) with a nil error — the gate NEVER blocks the stop event. |
| **Evidence** | `go test ./internal/hook/ -run 'TestStopHandler.*FailOpen\|TestStopEvidenceGate' -v -count=1` shows passing subtests asserting the returned `*HookOutput` is the empty/allow value and err is nil for finding-present, finding-absent, and StopHookActive cases. |
| **Pass criterion** | All fail-open subtests pass: `Handle()` returns an empty `&HookOutput{}` (no `decision: block`) and nil error regardless of whether the gate produced a finding, per REQ-SEG-005. |

---

## AC-SEG-006 — advisory to stderr only, stdout contract unchanged

| Field | Value |
|-------|-------|
| **Given** | M3 is complete. |
| **When** | The gate detects an unbacked success claim during a `Handle()` invocation. |
| **Then** | The advisory finding is written to stderr (via `slog.Warn` + a human-readable line) and the stdout `HookOutput` JSON the handler returns is the unchanged empty allow object (no `decision`/`reason` fields added by the gate). |
| **Evidence** | `grep -cE 'os\.Stdout|fmt\.Fprintln\(os\.Stdout' internal/hook/session_ledger.go` returns `0` (gate never writes stdout — bare `-E` alternation, dots escaped). AND `grep -cE 'os\.Stderr|slog\.Warn' internal/hook/session_ledger.go internal/hook/stop.go` returns ≥ 1 (advisory → stderr/slog). AND a test asserts `Handle()`'s returned `HookOutput` has no `Decision`/`Reason` populated when a finding fires. |
| **Pass criterion** | Zero stdout writes from the gate AND advisory routed to stderr/slog AND the returned HookOutput carries no stop-decision fields when a finding fires, per REQ-SEG-006. |

---

## AC-SEG-007 — C-HRA-008 subagent boundary

| Field | Value |
|-------|-------|
| **Given** | M1-M3 complete. |
| **When** | An auditor runs the C-HRA-008 static guard over `internal/hook/`. |
| **Then** | No non-test, non-comment source line in `internal/hook/` invokes `AskUserQuestion` or `mcp__askuser`. |
| **Evidence** | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v '_test.go' \| grep -v "^[^:]*:[0-9]*:[[:space:]]*//"` returns no output. |
| **Pass criterion** | The grep returns zero matches (exit 1 = no match = pass). The new `session_ledger.go` + `runEvidenceGate` additions introduce no AskUserQuestion / mcp__askuser call, per REQ-SEG-007. |

---

## AC-SEG-008 — ≤5s read-only + legacy fixture not false-flagged

| Field | Value |
|-------|-------|
| **Given** | M2-M4 complete. |
| **When** | (a) An auditor inspects BOTH the reader (`session_ledger.go`) AND the gate entry (`runEvidenceGate` in `stop.go`) for heavy operations; (b) the gate runs against a legacy session ledger whose JSONL records LACK the new `is_test_pass`/`is_test_fail`/`path_kind` fields but DO carry `outcome:"success"`. |
| **Then** | (a) Neither the reader NOR the gate-entry function performs test re-execution, network calls, or filesystem rescan beyond the single `LoadBySession` read; (b) the gate emits NO finding for the legacy success-claim session (absent binary fields treated as "not observable", not "failed"). |
| **Evidence** | (a) **Reader heavy-op grep** (bare `-E` alternation, dots escaped): `grep -cE 'exec\.Command|http\.|os\.ReadDir|filepath\.Walk|go test ' internal/hook/session_ledger.go` returns `0`. (a-2) **Gate-entry heavy-op grep on the `runEvidenceGate` body** (D3): extract `runEvidenceGate` body via `awk '/^func runEvidenceGate/{f=1} f{print} /^}/{if(f)exit}' internal/hook/session_ledger.go > /tmp/gate.txt`; then `grep -cE 'exec\.Command|http\.|os\.ReadDir|filepath\.Walk|go test ' /tmp/gate.txt` returns `0` AND `grep -c 'LoadBySession' /tmp/gate.txt` returns ≥ 1 (the gate-entry's ONLY data acquisition is `LoadBySession`, which reads ≤2 small JSONL files). (b) `go test ./internal/hook/ -run 'TestEvaluateEvidence.*Legacy|TestEvidenceGateLegacy' -v -count=1` shows a passing subtest where a record with `Outcome=success` and zero-value `IsTestPass`/`IsTestFail`/`PathKind` (legacy) → no finding. Optionally a real JSONL fixture: write a `usage-<today>.jsonl` line WITHOUT the new fields under a temp `.moai/evolution/telemetry/`, run `buildSessionLedger`(records from `LoadBySession`) → evaluator returns nil. |
| **Pass criterion** | (a) Zero heavy-op primitives in the reader AND (a-2) zero heavy-op primitives in the `runEvidenceGate` body with `LoadBySession` as its only data acquisition (D3 — `runEvidenceGate` is defined in session_ledger.go; stop.go holds only the void call, so the heavy-op assertion targets the runEvidenceGate body in session_ledger.go) AND (b) the legacy-record subtest yields no finding (absent fields ≠ failed) per REQ-SEG-008 + REQ-SEG-010. |

---

## AC-SEG-009 — behavior preservation of existing stop.go

| Field | Value |
|-------|-------|
| **Given** | M3 is complete; the gate is wired into `Handle()`. |
| **When** | An auditor compares the pre-existing stop.go steps against the post-edit version. |
| **Then** | The `StopHookActive` early-return guard, the 90d + 30d `PruneOldFiles` calls, and the `AnalyzeSessionAndLog` invocation (gated by `minToolInvocationsForReflection`) are all still present and unmodified; the gate is purely additive (inserted after 30d pruning, before the final return). |
| **Evidence** | `grep -c 'StopHookActive' internal/hook/stop.go` ≥ 1 AND `grep -c 'PruneOldFiles' internal/hook/stop.go` returns ≥ 2 (90d + 30d) AND `grep -c 'AnalyzeSessionAndLog' internal/hook/stop.go` ≥ 1 AND `grep -c 'minToolInvocationsForReflection' internal/hook/stop.go` ≥ 1. AND a test asserts: with `StopHookActive=true`, `Handle()` early-returns and `runEvidenceGate` is NOT reached (gate after the guard). |
| **Pass criterion** | All four pre-existing constructs (StopHookActive / PruneOldFiles×2 / AnalyzeSessionAndLog / minToolInvocationsForReflection) remain present AND a StopHookActive=true test confirms the early-return guard still short-circuits before the gate, per REQ-SEG-009. |

---

## AC-SEG-010 — graceful degradation (omitempty round-trip)

| Field | Value |
|-------|-------|
| **Given** | M1 is complete; `UsageRecord` extended with omitempty fields. |
| **When** | (a) A legacy JSONL line WITHOUT the new fields is unmarshaled into `UsageRecord`; (b) a new `UsageRecord` with zero-value new fields is marshaled. |
| **Then** | (a) The legacy line decodes successfully with `IsTestPass=false`, `IsTestFail=false`, `PathKind=""` (zero values, no error); (b) the marshaled output OMITS the new fields entirely (omitempty), so existing JSONL parsers see no schema change. |
| **Evidence** | `go test ./internal/telemetry/ -run 'TestUsageRecord.*Omitempty\|TestUsageRecordBackwardCompat' -v -count=1` shows: a legacy-line decode subtest (no `is_test_pass` key → zero values) AND a marshal subtest asserting the JSON output does NOT contain `is_test_pass`/`is_test_fail`/`path_kind` when those fields are zero-value. AND `grep -c 'omitempty' internal/telemetry/types.go` returns ≥ 3. |
| **Pass criterion** | The legacy-decode subtest produces zero-value new fields without error AND the zero-value marshal omits the three new JSON keys AND `types.go` carries ≥ 3 `omitempty` tags, per REQ-SEG-010. |

---

## AC-SEG-011 — path-kind taxonomy (docs-only / code-change / unknown)

| Field | Value |
|-------|-------|
| **Given** | M2 is complete; `inferPathKind` implemented. |
| **When** | `inferPathKind` is given: (a) a record with explicit `PathKind="code-change"`; (b) a legacy record set with `Phase=sync`; (c) a legacy record set with `Phase=run`; (d) an ambiguous record set with `Phase=none` / no AgentType signal. |
| **Then** | (a) returns `code-change` (explicit wins); (b) returns `docs-only` (inference); (c) returns `code-change` (inference); (d) returns `unknown` (ambiguous fallback — and the evaluator emits NO finding for `unknown`). |
| **Evidence** | `go test ./internal/hook/ -run 'TestInferPathKind' -v -count=1` shows passing subtests `explicit_pathkind_wins → code-change`, `phase_sync → docs-only`, `phase_run → code-change`, `ambiguous → unknown`. AND `go test ./internal/hook/ -run 'TestEvaluateEvidence' -v -count=1` shows `unknown_pathkind → nil_finding`. AND `grep -cE '"docs-only"|"code-change"|"unknown"|PathKindDocsOnly|PathKindCodeChange|PathKindUnknown' internal/hook/session_ledger.go internal/telemetry/types.go` returns ≥ 3 (three buckets present, ideally via constants — bare `-E` alternation). |
| **Pass criterion** | All four `inferPathKind` subtests pass AND the `unknown → nil_finding` subtest passes (ambiguous never flagged) AND the three bucket identifiers are present (preferably as constants per CLAUDE.local §14), per REQ-SEG-011. |

---

## D.6 Definition of Done

- [ ] AC-SEG-001..011 모두 PASS (각 evidence 명령 재현 가능)
- [ ] `go test ./...` GREEN (전체 회귀 없음)
- [ ] behavior preservation: StopHookActive 가드 / 90d·30d pruning / AnalyzeSessionAndLog 미변경 (AC-SEG-009)
- [ ] fail-open: `Handle()`가 모든 경로에서 `&HookOutput{}` 반환 (AC-SEG-005)
- [ ] C-HRA-008: AskUserQuestion/mcp__askuser 0 매치 (AC-SEG-007)
- [ ] ≤5s read-only: 게이트 경로에 heavy op 없음, `LoadBySession`만 (AC-SEG-008)
- [ ] graceful degradation: 레거시 레코드 no false-flag, omitempty round-trip (AC-SEG-008, AC-SEG-010)
- [ ] advisory → stderr only, stdout HookOutput 계약 불변 (AC-SEG-006)
- [ ] 새 원장 저장소 파일 0 (기존 JSONL schema 확장만) (AC-SEG-001)
- [ ] cross-platform build: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] coverage ≥85% (신규 코드)

## D.7 Edge Cases

1. **빈 세션 (records 0개)**: `LoadBySession`가 빈 슬라이스 반환 → `runEvidenceGate` 즉시 return (no finding). fail-open 불변.
2. **StopHookActive=true**: early-return 가드가 게이트보다 먼저 단락 → 게이트 미실행 (AC-SEG-009 검증).
3. **projectDir 빈 문자열**: `input.ProjectDir`/`input.CWD` 모두 빈 경우 게이트 skip (기존 pruning/reflection 동일 가드).
4. **모노레포 / 혼합 path-kind**: 한 세션에 docs + code 레코드 혼재 시 — `inferPathKind`는 code-change 신호(`Phase=run`/`manager-develop`)가 하나라도 있으면 code-change 우선(보수적: 코드 변경이 있으면 증거 요구). docs-only는 모든 레코드가 docs 신호일 때만.
5. **success claim 0개 (모두 error/unknown)**: 게이트는 finding 없음 (주장이 없으면 unbacked-claim도 없음).
6. **IsTestPass=true AND IsTestFail=true 동시 (혼합 세션)**: pass가 하나라도 관측되면 backed로 처리 (no finding) — 일부 테스트 통과 = 검증 관측됨.
7. **레거시 record-time wiring 미완 (모든 신규 필드 비어 있음)**: 게이트는 inference fallback(path-kind)으로 동작하되 binary 신호 부재 → success claim flag 안 함(REQ-SEG-010). 게이트 가치는 record-time wiring 완료 시 점증 (정직 framing — design §3 scope note).
