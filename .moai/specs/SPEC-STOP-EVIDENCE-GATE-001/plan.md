# Implementation Plan — SPEC-STOP-EVIDENCE-GATE-001

> Stop-hook verification-evidence completion gate (IMP-02) + session-ledger reader (IMP-03). Tier M. cycle_type=tdd.
> Status: draft. This plan derives from spec.md §D requirements and design.md §0-§4.

## A. Context

session-stop 시점에 세션 원장(기존 telemetry JSONL)을 읽어 "성공 주장 + 이진 증거 없음" 조합을 advisory로 표면화하는 게이트를 추가한다. 핵심 엔지니어링은 (1) Approach C(Hybrid) 데이터 모델 — `UsageRecord` omitempty 확장 + 레거시 추론 fallback, (2) `stopHandler.Handle()`에 behavior-preserving additive 게이트 삽입, (3) 기존 `LoadBySession` 위 read-only 세션 원장 리더 + path-kind 분류(docs-only 버킷 포함)이다.

본 작업은 **SPEC 본문 작성이 아니라 구현(run-phase)** 계획이다. 실제 편집 대상(예상):
- `internal/telemetry/types.go` — `UsageRecord`에 omitempty 필드 3종 추가 (IsTestPass / IsTestFail / PathKind)
- `internal/hook/session_ledger.go` (신규) — `SessionLedger` reader + `buildSessionLedger` + `inferPathKind` + `evaluateEvidence`
- `internal/hook/stop.go` — `runEvidenceGate` 신규 함수 + `Handle()` final return 직전 1줄 삽입 (additive)
- `internal/hook/session_ledger_test.go` (신규) — reader + 분류 + 결정 로직 단위 테스트
- `internal/hook/stop_test.go` (기존 보강 또는 신규) — behavior preservation + fail-open + 게이트 통합 테스트
- `internal/telemetry/types_test.go` (기존 보강) — omitempty backward-compat round-trip

PRESERVE 대상: `Handle()`의 StopHookActive 가드 / 90d·30d pruning / `AnalyzeSessionAndLog` (REQ-SEG-009). `LoadBySession` / `RecordSkillUsage` 기존 시그니처 (변경 금지 — read만).

[HARD — dormant scaffold framing (D1)] 본 게이트는 **현재 텔레메트리 스트림에 대해 knowingly-dormant**다 (spec.md §A.3 + REQ-SEG-012). 유일한 production writer `logSkillUsage`(post_tool_metrics.go:72-82)가 `Outcome: OutcomeUnknown` / `Phase: "none"`를 하드코딩하므로 `hasSuccessClaim`은 현재 항상 false → 게이트 미발화. 본 SPEC의 deliverable는 **read-side 로직 + schema/분류**이며, 현재 스트림에서 finding을 실제 발화하는 것은 deliverable가 **아니다**. 게이트의 가치는 record-time writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001`이 (a) `PathKind="code-change"` + (b) `Outcome="success"` + `IsTestPass`를 set해야 unblock된다. run-phase는 이 dormant framing을 CHANGELOG/주석에 정직하게 반영한다 (SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001 scaffold-honest 선례).

## B. Known Issues / Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| 레거시 JSONL 레코드(새 필드 부재)를 "evidence failed"로 오해 → 모든 과거 세션 false-flag | High (central correctness risk) | `evidence absent ≠ evidence failed` (REQ-SEG-010, design §0.3). zero-value를 "관측 불가"로 처리, success claim flag 안 함. legacy-record fixture로 검증 (AC-SEG-008) |
| `Handle()` 기존 동작 회귀 (early-return / pruning / reflection) | High | 게이트는 step 6 이후 final return 직전에만 삽입 (design §1.1). 기존 단계 미접촉. behavior-preservation 테스트 의무 (AC-SEG-009) |
| C-HRA-008 위반 (게이트 코드가 AskUserQuestion 호출) | High (binary constraint) | 게이트는 slog/stderr만 사용. CI guard grep AC (AC-SEG-007) |
| stop을 차단 (fail-open 위반) | High | `runEvidenceGate`는 void/error-ignored; `Handle()` return은 항상 `&HookOutput{}`. fail-open 테스트 (AC-SEG-005) |
| ≤5s budget 초과 (게이트가 heavy work) | Med | `LoadBySession`만 호출(이미 적재된 원장 in-memory 평가). test 재실행/network/rescan 금지. read-only 단언 (AC-SEG-008) |
| stdout HookOutput JSON 계약 오염 (advisory를 stdout에 씀) | Med | advisory는 stderr만 (REQ-SEG-006). stdout JSON 불변 검증 (AC-SEG-006) |
| omitempty 필드 추가가 기존 JSONL round-trip 깨뜨림 | Med | `omitempty` → 부재 필드 미출력, zero-value 디코드. round-trip 테스트 (AC-SEG-010) |
| docs-only 추론 오분류 (코드 세션을 docs-only로) → false-negative | Low | inference fallback은 명시 PathKind 우선; 추론은 conservative(애매 → unknown, flag 안 함). docs-only/code-change/unknown fixture 각각 검증 (AC-SEG-003/004/011) |
| 하드코딩 (path-kind 문자열 / phase 값 산재) | Low | 상수화 (design §2.2 taxonomy). `internal/config/envkeys.go` 또는 telemetry 상수 참조 (CLAUDE.local §14) |
| 게이트 가치 과대광고 (dormant scaffold를 "결함을 잡는다"고 주장) | Med (D1 framing risk) | spec.md §A.3 dormant scaffold + REQ-SEG-012 falsifiable condition. CHANGELOG/주석에 "현재 스트림에 대해 dormant; 가치는 `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked" 명시 (scaffold-honest 선례). originating incident(sync-phase)는 docs-only EXEMPT이라 본 게이트 미탐지 — §A.1 정합 |

## C. Pre-flight Checks

- [ ] 현재 branch + baseline 확인: `git branch --show-current; git rev-parse HEAD`
- [ ] Cross-platform build 가능성: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`
- [ ] 기존 hook 테스트 baseline GREEN: `go test ./internal/hook/... -count=1`
- [ ] 기존 telemetry 테스트 baseline GREEN: `go test ./internal/telemetry/... -count=1`
- [ ] C-HRA-008 baseline 측정: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` (현재 0 매치 확인)
- [ ] `stop.go` PRESERVE 단계 재확인 (StopHookActive L44 / 90d L56 / reflection L64 / 30d L72 / final return L79)
- [ ] SPEC-HARNESS-OUTCOME-CAPTURE-001 omitempty 확장 선례 재확인 (`grep -n 'omitempty' internal/telemetry/types.go` — 현재 0; brownfield 확장 패턴 참조)
- [ ] coverage baseline: `go test -cover ./internal/hook/... ./internal/telemetry/...`

## D. Constraints

- [HARD] SPEC 본문 외 코드/테스트 편집은 run-phase(manager-develop)에서 수행. 본 plan은 그 작업 분해.
- [HARD] constraint #1 — 새 원장 저장소 파일 금지. 세션 원장은 기존 `usage-*.jsonl` 위 read-only 뷰. `UsageRecord` schema omitempty 확장은 허용(새 store 아님).
- [HARD] constraint #6 / REQ-SEG-005 — fail-open. `Handle()` return은 항상 `&HookOutput{}`. exit-2/blocking 미도입(도입 시 env 플래그 휴면 OFF).
- [HARD] constraint #5 / REQ-SEG-007 — C-HRA-008. 게이트 코드 AskUserQuestion/mcp__askuser 0 매치.
- [HARD] constraint #7 / REQ-SEG-008 — ≤5s. `LoadBySession`만; test 재실행/network/rescan 금지.
- [HARD] REQ-SEG-009 — behavior preservation. StopHookActive 가드 / 90d·30d pruning / `AnalyzeSessionAndLog` 미수정.
- [HARD] REQ-SEG-010 — graceful degradation. 레거시 레코드(필드 부재) false-flag 금지. `evidence absent ≠ failed`.
- [HARD] PRESERVE: `LoadBySession` / `RecordSkillUsage` / `PruneOldFiles` 시그니처 변경 금지.
- [HARD] CLAUDE.local §14 — 하드코딩 금지. path-kind/phase 문자열은 상수화.

## E. Self-Verification

run-phase 완료 시 §F의 각 마일스톤 exit 기준 + acceptance.md 11개 AC(AC-SEG-001..011, 12 REQ에 매핑 — REQ-SEG-012는 문서/framing 요건으로 AC-SEG-008 dormant-fixture + spec.md §A.3/§A.4 문서 검증으로 커버)가 모두 PASS여야 한다. 검증은 read-only batch(go test, grep, cross-build)로 병렬 수행. behavior-preservation + fail-open + C-HRA-008은 must-pass.

## F. Milestones

### M1 — data-model extension (Approach C record-time half)

**대상**: `internal/telemetry/types.go` + `internal/telemetry/types_test.go`

작업:
1. `UsageRecord`에 omitempty 필드 3종 추가: `IsTestPass bool json:"is_test_pass,omitempty"`, `IsTestFail bool json:"is_test_fail,omitempty"`, `PathKind string json:"path_kind,omitempty"` (design §3).
2. path-kind 값 상수화: `PathKindDocsOnly = "docs-only"`, `PathKindCodeChange = "code-change"`, `PathKindUnknown = "unknown"` (CLAUDE.local §14 하드코딩 금지).
3. backward-compat round-trip 테스트: 기존(필드 없는) JSONL 라인 → `UsageRecord` 디코드 → zero-value 확인; 신규 레코드 marshal → omitempty로 부재 필드 미출력 확인.

**Exit 기준**: `go test ./internal/telemetry/... -count=1` GREEN; round-trip 테스트 통과(레거시 라인 디코드 OK, omitempty marshal OK); 기존 telemetry 테스트 무회귀.

### M2 — session-ledger reader + path-kind classification + evidence evaluator (IMP-03)

**대상**: `internal/hook/session_ledger.go` (신규) + `internal/hook/session_ledger_test.go` (신규)

작업:
1. `SessionLedger` struct + `buildSessionLedger(records []telemetry.UsageRecord) SessionLedger` (design §2.1) — `LoadBySession` 결과만 입력, 새 I/O 없음.
2. `inferPathKind(records)` (design §2.2): (1) 명시 `PathKind` 우선 → (2) `Phase=sync`/`AgentType=manager-docs` ⇒ docs-only, `Phase in {run,plan}`/`AgentType=manager-develop` ⇒ code-change → (3) `unknown`.
3. `evaluateEvidence(ledger) *Finding` (design §0.4): docs-only/unknown → nil; success claim + binary absent → nil(REQ-SEG-010); success claim + binary present + no pass → Finding.
4. `Finding` 타입 + `HumanReadable()` + `slogArgs()` (design §0.5).
5. 단위 테스트: docs-only 면제(REQ-SEG-003) / code-change + success + no-pass → finding(REQ-SEG-004) / code-change + success + binary absent → no finding(REQ-SEG-010, the legacy case) / unknown → no finding / 명시 PathKind 우선 / 추론 fallback 각 분기.

**Exit 기준**: `go test ./internal/hook/ -run 'TestSessionLedger|TestInferPathKind|TestEvaluateEvidence' -count=1` GREEN; 모든 결정 분기(§0.4 6 case) 커버; `evaluateEvidence`는 부수효과 없는 pure 판단(테스트 가능).

### M3 — stop.go gate insertion (IMP-02, behavior-preserving additive)

**대상**: `internal/hook/stop.go` + `internal/hook/stop_test.go`

작업:
1. `runEvidenceGate(projectRoot, sessionID string)` 신규 함수 (design §1.2): `LoadBySession` → `buildSessionLedger` → `evaluateEvidence` → finding 있으면 `slog.Warn` + stderr 한 줄. error/empty는 swallow(fail-open).
2. `Handle()`에 1줄 삽입: 30d pruning(L72-74) 이후, `return &HookOutput{}, nil`(L79) 직전. `if projectDir != "" { runEvidenceGate(projectDir, input.SessionID) }` (design §1.1).
3. behavior-preservation 테스트: StopHookActive=true → early-return 불변(게이트 미실행); pruning/reflection 호출 경로 불변; `Handle()` return은 모든 경우 `&HookOutput{}`(fail-open).
4. fail-open 테스트: finding이 있어도 `Handle()` return은 allow(`&HookOutput{}`), error 없음.
5. stdout 계약 테스트: 게이트가 HookOutput JSON을 변경하지 않음(advisory는 stderr만).

**Exit 기준**: `go test ./internal/hook/... -count=1` GREEN; behavior-preservation 테스트 통과(기존 3단계 미변경); fail-open 테스트 통과(finding 유무 무관 allow); C-HRA-008 grep 0 매치.

### M4 — 통합 검증 / 회귀 / boundary

작업:
1. C-HRA-008: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 매치 (AC-SEG-007).
2. ≤5s read-only 단언: 게이트 경로에 `os.ReadFile`/`exec`/`http`/test 재실행 없음 — `LoadBySession`만 (AC-SEG-008, grep 단언 + 코드 inspection).
3. backward-compat: 레거시 fixture(omitempty 필드 없는 JSONL) → 게이트 no false-flag (AC-SEG-008/010 런타임 fixture).
4. cross-platform: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
5. coverage: `go test -cover ./internal/hook/... ./internal/telemetry/...` ≥85% (신규 코드).
6. 전체 회귀: `go test ./...` GREEN.

**Exit 기준**: acceptance.md AC-SEG-001..011 전부 PASS.

## G. Anti-Patterns (avoid)

- 레거시 레코드(omitempty 필드 부재)를 "test failed"로 간주 → 모든 과거 세션 false-flag (REQ-SEG-010 위반, central correctness risk).
- `Handle()`의 기존 단계 사이에 게이트 삽입 → 회귀 위험 (final return 직전 additive만).
- advisory finding을 stdout HookOutput JSON에 추가 → Stop 결정 계약 오염 (stderr만).
- 게이트가 stop을 차단(exit 2 / decision=block) → fail-open 위반.
- 게이트에서 test 재실행 / 파일 스캔 / network → ≤5s 위반 (`LoadBySession`만).
- 새 원장 저장소 파일 생성 → constraint #1 위반 (기존 JSONL schema 확장만 허용).
- coverage-percentage 상관 분석 추가 → constraint #3 / IMP-06 scope.
- "no unobserved-verification-claim" invariant를 정책으로 codify → IMP-06 scope (본 SPEC은 기계적 detection만).
- path-kind 문자열 하드코딩 산재 → 상수화 (CLAUDE.local §14).
- `AnalyzeSessionAndLog` / `PruneOldFiles` 수정 → PRESERVE 위반 (REQ-SEG-009).
- 게이트를 "현재 결함을 잡는다"고 CHANGELOG/주석에 과대광고 → dormant scaffold framing 위반 (REQ-SEG-012, D1). 현재 production writer 하드코딩으로 게이트는 미발화이며, 가치는 `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked임을 정직하게 명시해야 한다.
- originating incident(`L_manager_docs_false_backfill_report`, sync-phase)를 "본 게이트가 잡는다"고 주장 → 그 사건은 docs-only EXEMPT(REQ-SEG-003)이므로 미탐지. 본 게이트는 별개의 code-session false-success 형태를 겨냥 (spec.md §A.1).

## H. Cross-References

- spec.md §A.1 (originating-defect framing), §A.3 (dormant scaffold), §A.4 (writer successor), §D (REQ-SEG-001..012), §Exclusions
- design.md §0 (data-model fork resolution Approach C), §1 (stop.go EXTEND), §2 (ledger reader), §3 (schema 확장)
- acceptance.md (AC-SEG-001..011)
- research.md §2.1 (production writer dormancy ground-truth), §4 (scaffold-honest 선례)
- successor SPEC id `SPEC-STOP-EVIDENCE-WRITER-001` (record-time writer; FREE, not yet authored)
- `internal/hook/stop.go`, `internal/hook/post_tool_metrics.go:72-82` (production writer), `internal/hook/types.go` (HookInput), `internal/telemetry/types.go`, `internal/telemetry/recorder.go`, `internal/hook/CLAUDE.md`
- CLAUDE.local.md §6 (테스트 격리 t.TempDir), §14 (하드코딩 금지)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier M Section A-E 위임 템플릿)
