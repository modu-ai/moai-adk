# Design — SPEC-STOP-EVIDENCE-GATE-001

> Technical design for (1) the data-model fork resolution, (2) the Stop-hook evidence gate, and (3) the session-ledger reader.
> Status: draft. WHAT/WHY lives in spec.md; this is the HOW for run-phase implementation.

## 0. The Design Tension (must resolve explicitly)

현재 원장(`UsageRecord`)은 `Outcome`(success|partial|error|unknown)을 invocation별로 기록하지만:
- (a) **테스트 pass/fail 이진 증거**를 별도 discrete 필드로 저장하지 않는다.
- (b) 건드린 **파일 경로 / path-kind**를 기록하지 않는다.

그런데 spec.md의 constraint #2(binary evidence)와 #4(docs-only path-kind)는 이 두 신호에 의존한다. `Event` struct(`ToolName`/`IsError`/`IsTestPass`/`IsTestFail`)는 outcome heuristic용으로 존재하지만 **JSONL에 영속화되지 않는다**(types.go 주석: "NOT persisted to JSONL — used for outcome heuristics").

### 0.1 Resolution — Approach C (Hybrid: brownfield-extend + inference fallback)

본 SPEC은 **Approach C (Hybrid)**를 채택한다.

**한 줄 요약**: record-time에 기존 `UsageRecord` schema를 backward-compatible `omitempty` 필드로 brownfield-extend(증거가 관측될 때만 기록)하고, 그 필드가 없는 레거시 레코드에는 기존 필드(`Outcome` + `Phase` + `AgentType`)로부터 path-kind/evidence를 **추론(infer)**한다.

선택지 비교:

| Approach | Mechanism | 장점 | 채택? |
|----------|-----------|------|-------|
| (A) brownfield-extend `UsageRecord` `omitempty` 필드 (binary-evidence + path-kind, record-time) | `IsTestPass`/`IsTestFail`/`PathKind`를 `omitempty`로 영속화 | 정확한 이진 증거; SPEC-HARNESS-OUTCOME-CAPTURE-001이 동일 JSONL store에 `Event` omitempty 확장 선례 | 부분 채택 (C의 절반) |
| (B) 기존 필드만으로 추론 (no schema change) | `Phase=sync` OR `AgentType=manager-docs` ⇒ docs-only; `Outcome=success`만으로 soft claim | schema 무변경; 레거시 호환 100% | 부분 채택 (C의 절반, 레거시 fallback) |
| **(C) Hybrid** | record-time omitempty 확장 **+** 필드 부재 시 추론 fallback | 신규 레코드는 정확한 이진 증거, 레거시 레코드는 graceful degradation | **채택** |

**채택 이유(one-line why)**: Approach A 단독은 모든 레거시 JSONL을 "evidence absent → false-flag"로 만들어 REQ-SEG-010을 위반하고, Approach B 단독은 `L_manager_docs_false_backfill_report`가 요구하는 진짜 이진 증거(`IsTestPass`)를 영원히 확보하지 못한다. Hybrid만이 신규 정확성과 레거시 graceful degradation을 동시에 만족한다.

### 0.2 constraint #1 호환성 (NEW store 금지 vs schema 확장)

[HARD] constraint #1은 **새 원장 저장소 파일**을 금지한다. Approach C의 record-time 확장은 **기존** `usage-*.jsonl`의 **레코드 schema에 backward-compatible omitempty 필드를 추가**하는 것이며, 새 파일/새 store를 만들지 않는다. 선례: SPEC-HARNESS-OUTCOME-CAPTURE-001이 동일 JSONL store에 `Event` omitempty 확장을 적용했다. 따라서 schema 확장은 constraint #1 위반이 아니다 — 명시적으로 허용된다.

### 0.3 backward-compatibility implications (graceful degradation)

기존 `usage-*.jsonl` 레코드는 새 필드(`IsTestPass`/`IsTestFail`/`PathKind`)가 없다. `encoding/json`은 부재 필드를 zero-value로 디코드한다(`bool` → `false`, `string` → `""`).

[HARD] 게이트는 zero-value를 **"evidence failed"로 해석해서는 안 된다**. 구분:
- `PathKind == ""` (부재) → path-kind를 `Phase`/`AgentType`에서 **추론** (Approach B fallback); 추론 불가 시 `unknown` 버킷.
- `IsTestPass == false` AND `IsTestFail == false` (둘 다 부재/false) → **"binary signal not observable"** (NOT "test failed"). REQ-SEG-010에 따라 이 경우 success claim을 unbacked로 flag하지 **않는다** (관측 불가는 실패가 아니다).
- `IsTestFail == true` → 진짜 관측된 실패 (이 경우 success claim과의 모순은 표면화 대상).

### 0.4 게이트의 핵심 판단 로직 (claimed success vs binary evidence)

```
classify session:
  pathKind = inferPathKind(records)   # docs-only | code-change | unknown
  hasSuccessClaim = any(rec.Outcome in {success, partial})
  hasBinaryPass   = any(rec.IsTestPass == true)
  hasBinaryFail   = any(rec.IsTestFail == true)
  binaryObservable = hasBinaryPass OR hasBinaryFail

decision:
  if pathKind == "docs-only":                         → no finding (REQ-SEG-003: docs exempt)
  elif pathKind == "unknown":                          → no finding (conservative; REQ-SEG-010)
  elif hasSuccessClaim AND NOT binaryObservable:       → no finding (REQ-SEG-010: absent ≠ failed)
  elif hasSuccessClaim AND hasBinaryFail AND NOT hasBinaryPass:  → ADVISORY FINDING (success claimed, only failure observed)
  elif hasSuccessClaim AND binaryObservable AND NOT hasBinaryPass: → ADVISORY FINDING (success claimed, no pass observed — the L_manager_docs_false_backfill_report shape)
  else:                                                → no finding (success backed by observed pass)
```

핵심: "success claimed + binary evidence present + NO pass" → flag. "success claimed + binary evidence ABSENT" → NO flag(conservative, REQ-SEG-010). 이것이 `L_manager_docs_false_backfill_report`(완료 주장 but 증거 없음)와 정상 완료를 구분하는 지점이다.

### 0.5 advisory warning — 내용과 출력 위치 (be concrete)

[HARD] 출력 위치: **stderr** (REQ-SEG-006). stdout HookOutput JSON 계약은 건드리지 않는다 (Stop 결정 채널 보존).

advisory warning 내용(human-readable, 구조화):

```
[evidence-gate] session <session_id> claimed success without observed binary evidence
  path-kind:        code-change
  success claims:   <N> record(s) with outcome=success|partial
  binary pass:      observed=false
  binary fail:      observed=<true|false>
  hint:             a 'success' outcome was recorded but no test-pass signal was
                    observed in this session's telemetry. Verify the claimed
                    completion was actually observed (cf. L_manager_docs_false_backfill_report).
```

`slog.Warn("evidence-gate: unbacked success claim", "session_id", ..., "path_kind", ..., "success_claims", ..., "binary_pass_observed", false, ...)` 형태로 emit (slog는 stderr로 라우팅). human-readable 한 줄 요약을 함께 stderr에 쓴다.

## 1. Stop-hook Evidence Gate (`internal/hook/stop.go` EXTEND)

### 1.1 Insertion point (additive, behavior-preserving)

기존 `Handle()` 실행 순서(PRESERVE 대상 — REQ-SEG-009):
1. `slog.Info("stop requested", ...)` (L36)
2. `if input.StopHookActive { return &HookOutput{}, nil }` (L44-47) — early-return 가드 (PRESERVE)
3. `projectDir` 해석 (`input.ProjectDir` → `input.CWD`) (L49-52)
4. 90d pruning (L56-60) (PRESERVE)
5. reflective learning `AnalyzeSessionAndLog` (L64-69) (PRESERVE)
6. 30d pruning (L72-74) (PRESERVE)
7. `return &HookOutput{}, nil` (L79) (PRESERVE)

[HARD] 신규 게이트 삽입 위치: **step 6 이후, step 7(final return) 직전**. 즉 모든 기존 단계가 완료된 뒤 게이트를 호출하고, 게이트는 항상 `&HookOutput{}`로 흐름을 이어간다.

```go
// (after 30d pruning, before final return)
// Evidence gate: advisory-only. Surfaces unbacked success claims.
// NEVER blocks stop (fail-open per REQ-SEG-005). All findings go to stderr.
if projectDir != "" {
    runEvidenceGate(projectDir, input.SessionID)  // void; logs advisory to stderr only
}

return &HookOutput{}, nil
```

- `runEvidenceGate`는 `error`를 반환하지 않거나, 반환하더라도 `Handle()`이 무시한다(advisory). 어떤 경우에도 `Handle()`의 return은 `&HookOutput{}` (allow) 불변.
- `projectDir == ""`일 때 게이트 skip (기존 pruning/reflection도 동일 가드).

### 1.2 `runEvidenceGate` (신규 함수, 같은 패키지)

```go
// runEvidenceGate reads the already-available session ledger and surfaces an
// advisory finding when a success claim is not backed by observed binary
// evidence. It NEVER blocks the stop event (fail-open). All output goes to
// stderr via slog. Errors are swallowed (best-effort, like pruning/reflection).
func runEvidenceGate(projectRoot, sessionID string) {
    records, err := telemetry.LoadBySession(projectRoot, sessionID)  // REUSE — REQ-SEG-001
    if err != nil || len(records) == 0 {
        return  // nothing to evaluate; fail-open
    }
    ledger := buildSessionLedger(records)        // IMP-03 reader (§2)
    finding := evaluateEvidence(ledger)          // §0.4 decision logic
    if finding != nil {
        slog.Warn("evidence-gate: unbacked success claim", finding.slogArgs()...)
        // also write the human-readable one-liner to stderr
        fmt.Fprintln(os.Stderr, finding.HumanReadable())
    }
}
```

- `LoadBySession`만 호출 — 새 I/O / network / test 재실행 없음 (REQ-SEG-008, ≤5s).
- C-HRA-008: AskUserQuestion / mcp__askuser 미호출 (REQ-SEG-007).

## 2. Session Ledger Reader (IMP-03, built on LoadBySession)

### 2.1 `SessionLedger` (read-only view, no new store)

```go
// SessionLedger is a read-only, in-memory view over a session's telemetry
// records. It is built from telemetry.LoadBySession — it does NOT create or
// write any storage file (REQ-SEG-001).
type SessionLedger struct {
    SessionID     string
    Records       []telemetry.UsageRecord
    PathKind      string   // "docs-only" | "code-change" | "unknown"
    SuccessClaims int      // count of Outcome in {success, partial}
    BinaryPass    bool     // any IsTestPass == true
    BinaryFail    bool     // any IsTestFail == true
}

func buildSessionLedger(records []telemetry.UsageRecord) SessionLedger { ... }
```

배치: `internal/hook/session_ledger.go` (게이트와 같은 패키지; telemetry import는 read-only). 또는 `internal/telemetry`에 reader를 두고 hook이 호출 — design 결정: **`internal/hook`에 둔다** (stop.go와 응집; telemetry는 storage 책임만 유지). 단, `UsageRecord`의 신규 omitempty 필드(§3)는 `internal/telemetry/types.go`에 추가한다.

### 2.2 path-kind 분류 (`inferPathKind`)

```
inferPathKind(records):
  # (1) explicit signal first (Approach A — neW omitempty PathKind field)
  for rec in records:
    if rec.PathKind != "":  return rec.PathKind   # docs-only | code-change

  # (2) inference fallback (Approach B — legacy records, no PathKind field)
  if all(rec.Phase == "sync") OR any(rec.AgentType == "manager-docs"):
    return "docs-only"
  if any(rec.Phase in {run, plan}) OR any(rec.AgentType == "manager-develop"):
    return "code-change"

  # (3) ambiguous / absent
  return "unknown"
```

- (1) 신규 레코드: 명시 `PathKind` 우선.
- (2) 레거시 레코드: `Phase`/`AgentType`에서 추론 (`Phase=sync` 또는 `AgentType=manager-docs` ⇒ docs-only; `Phase in {run,plan}` 또는 `AgentType=manager-develop` ⇒ code-change).
- (3) 추론 불가 → `unknown` (conservative, REQ-SEG-011 + REQ-SEG-010).

### 2.3 docs-only 버킷의 면제 (REQ-SEG-003)

`PathKind == "docs-only"`이면 `evaluateEvidence`는 즉시 `nil`(no finding) 반환. 문서 세션은 binary test-pass 증거를 요구받지 않는다.

## 3. Data-model extension (`internal/telemetry/types.go`)

`UsageRecord`에 backward-compatible `omitempty` 필드 3종 추가 (Approach C record-time 절반):

```go
type UsageRecord struct {
    Timestamp   time.Time `json:"ts"`
    SessionID   string    `json:"session_id"`
    SkillID     string    `json:"skill_id"`
    Trigger     string    `json:"trigger"`
    ContextHash string    `json:"context_hash"`
    AgentType   string    `json:"agent_type"`
    Phase       string    `json:"phase"`
    DurationMs  int64     `json:"duration_ms"`
    Outcome     string    `json:"outcome"`
    // --- NEW (SPEC-STOP-EVIDENCE-GATE-001, omitempty, backward-compatible) ---
    IsTestPass  bool      `json:"is_test_pass,omitempty"`  // binary evidence: a test-pass was observed
    IsTestFail  bool      `json:"is_test_fail,omitempty"`  // binary evidence: a test-fail was observed
    PathKind    string    `json:"path_kind,omitempty"`     // "docs-only" | "code-change" (empty = infer)
}
```

[HARD] `omitempty`로 추가 → 기존 레코드 디코드 시 zero-value, 기존 JSONL 파일 영향 없음 (REQ-SEG-010). 기존 `RecordSkillUsage` append 경로는 새 필드를 옵션으로 채울 수 있으나, **본 SPEC은 record-time 채움(populate)을 강제하지 않는다** — record-time 채움은 record-time-writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001`(spec.md §A.4 + REQ-SEG-012)의 소관이고, 본 SPEC의 게이트는 채워진 경우(신규)와 비워진 경우(레거시) 모두 graceful하게 **동작(crash 없이 평가)**한다.

> **Scope note (D1-corrected — value blocked, behavior graceful)**: 본 SPEC의 deliverable는 (a) schema 확장(필드 정의) + (b) 게이트/리더가 그 필드를 read하고 부재 시 추론. record-time에 누가 이 필드를 write하는지(어느 훅이 `IsTestPass`/`PathKind=code-change`/`Outcome=success`를 set하는지)의 production wiring은 본 SPEC scope **밖**이며 후속 writer SPEC `SPEC-STOP-EVIDENCE-WRITER-001` 소관이다. [HARD 구분] 게이트의 **동작(behavior)은 wiring에 blocked되지 않는다** (필드 부재 시 graceful degradation으로 crash 없이 평가) — 그러나 게이트의 **가치(value, 즉 표적 결함을 실제로 탐지하는 능력)는 그 writer에 blocked된다**: 현재 production writer(`logSkillUsage`, post_tool_metrics.go:72-82)가 `Outcome: OutcomeUnknown`/`Phase: "none"`를 하드코딩하므로 게이트는 현재 스트림에서 dormant이다 (spec.md §A.3, research.md §2.1). 이 dormant framing은 SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001의 "scaffold + dormant, 0 production caller" 정직 framing과 동형이다 (과대광고 금지).

## 4. Design Decisions / Trade-offs

| Decision | Rationale | Alternative rejected |
|----------|-----------|----------------------|
| Approach C (Hybrid) | 신규 정확성 + 레거시 graceful degradation 동시 만족 | A 단독(레거시 false-flag), B 단독(이진 증거 영구 부재) |
| 게이트 삽입 = step 6 이후 final return 직전 | 모든 기존 단계 보존 후 additive | 기존 단계 사이 삽입 (회귀 위험) |
| advisory → stderr (slog.Warn) | stdout stop 결정 계약 보존 (REQ-SEG-006) | stdout JSON에 finding 추가 (계약 오염) |
| `evidence absent ≠ evidence failed` | 레거시 레코드 false-flag 방지 (REQ-SEG-010) | 부재를 실패로 간주 (모든 과거 세션 flag) |
| reader를 `internal/hook`에 배치 | stop.go와 응집; telemetry는 storage 책임만 | `internal/telemetry`에 reader (storage/logic 혼재) |
| record-time populate 미강제 (writer SPEC로 분리) | 게이트 **동작**(behavior)은 graceful degradation으로 wiring-독립; 게이트 **가치**(value)는 writer SPEC `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked (정직 framing, D1) | record-time wiring 본 SPEC에 포함 (scope creep, production writer `logSkillUsage` 수정 — 별도 SPEC가 적절) |
| exit-2 / blocking 미도입 (fail-open) | warn-first 점진적 도입 (IMP-01 정합) | 즉시 blocking (예기치 않은 stop 차단) |

## 5. Cross-References

- spec.md §A.1 (originating-defect framing), §A.3 (dormant scaffold), §A.4 (writer successor), §D (REQ-SEG-001..012), §Exclusions
- acceptance.md (AC-SEG-001..011)
- research.md §2.1 (production writer dormancy ground-truth), §4 (scaffold-honest 선례)
- successor SPEC id `SPEC-STOP-EVIDENCE-WRITER-001` (record-time evidence/path-kind writer; FREE, not yet authored)
- `internal/hook/stop.go` (EXTEND 지점 L72-79), `internal/hook/post_tool_metrics.go:72-82` (production writer dormancy), `internal/hook/types.go` (HookInput no Phase field), `internal/telemetry/types.go` (schema 확장), `internal/telemetry/recorder.go` (LoadBySession REUSE)
- `internal/hook/CLAUDE.md` (≤5s / C-HRA-008 / $CLAUDE_PROJECT_DIR)
- lesson `L_manager_docs_false_backfill_report` (motivating case ONLY — sync-phase incident is docs-only-EXEMPT here, NOT caught by this gate)
