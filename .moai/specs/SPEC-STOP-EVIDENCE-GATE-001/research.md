# Research — SPEC-STOP-EVIDENCE-GATE-001

> Codebase analysis grounding the plan/design. Status: draft.
> Confirms ground-truth so run-phase does not re-derive; identifies the brownfield-extension precedent for the data-model fork.

## 1. `internal/hook/stop.go` — current `Handle()` flow (PRESERVE target)

`stopHandler.Handle(ctx, input)` 현재 실행 순서 (verified, line refs):

| Step | Line | Action | Disposition |
|------|------|--------|-------------|
| 1 | L36-39 | `slog.Info("stop requested", ...)` | PRESERVE |
| 2 | L44-47 | `if input.StopHookActive { return &HookOutput{}, nil }` (infinite-loop guard) | PRESERVE |
| 3 | L49-52 | `projectDir = input.ProjectDir` then `input.CWD` fallback | PRESERVE (reuse) |
| 4 | L56-60 | `telemetry.PruneOldFiles(projectDir, 90)` (best-effort, errors logged) | PRESERVE |
| 5 | L64-69 | `if countSessionRecords(...) >= minToolInvocationsForReflection { AnalyzeSessionAndLog(...) }` | PRESERVE |
| 6 | L72-74 | `_ = telemetry.PruneOldFiles(projectDir, 30)` | PRESERVE |
| 7 | L79 | `return &HookOutput{}, nil` (always allow — fail-open today) | PRESERVE |

핵심 관찰:
- 핸들러는 **이미 fail-open** (어떤 경로에서도 stop을 차단하지 않음). 본 SPEC의 게이트는 이 fail-open 성질을 유지한다 (REQ-SEG-005).
- `countSessionRecords` (L89-95)는 **이미 `telemetry.LoadBySession(projectRoot, sessionID)`를 호출**한다 (L90). 즉 세션 원장 적재 경로가 이미 존재 — 본 SPEC의 게이트는 같은 호출을 재사용한다 (REQ-SEG-001, 새 read 경로 불필요).
- `minToolInvocationsForReflection = 3` (L85). reflective learning은 ≥3 레코드 세션에만 발화.
- 삽입 지점: step 6(30d pruning) 이후, step 7(final return) 직전 — 모든 기존 단계를 보존하는 additive 위치 (design §1.1).

## 2. `internal/telemetry/types.go` — `UsageRecord` + `Event` (data-model fork target)

`UsageRecord` (JSONL 영속):
```
Timestamp(ts) / SessionID(session_id) / SkillID(skill_id) / Trigger(trigger) /
ContextHash(context_hash) / AgentType(agent_type) / Phase(phase) /
DurationMs(duration_ms) / Outcome(outcome)
```
- `Outcome` ∈ {success, partial, error, unknown} — **soft** success 신호.
- `Phase` ∈ {plan, run, sync, none} — path-kind 추론 입력 (sync ⇒ docs-ish).
- `AgentType` — path-kind 추론 입력 (manager-docs ⇒ docs, manager-develop ⇒ code).

`Event` (JSONL **미영속** — outcome heuristic 전용; types.go L33-44 주석 명시):
```
ToolName / IsError / IsTestPass / IsTestFail
```
- `IsTestPass`/`IsTestFail`는 **이진 증거** 개념을 이미 정의하고 있으나 **영속화되지 않는다**.

**Fork의 본질**: 게이트가 필요로 하는 (a) 이진 증거, (b) path-kind는 `UsageRecord`에 없고, 이진 증거 개념(`IsTestPass`/`IsTestFail`)은 `Event`에 있으나 디스크에 남지 않는다. 따라서 게이트가 stop-time에 읽을 수 있는 영속 신호가 부족하다 → design §0의 Approach 선택 필요.

## 2.1 Production writer ground-truth — the dormancy proof (D1, plan-auditor 확인)

[HARD — 게이트의 현재 dormancy 근거] 유일한 production `UsageRecord` writer는 `logSkillUsage` (`internal/hook/post_tool_metrics.go:72-82`)이며, 다음을 **하드코딩**으로 기록한다 (verified):

```go
r := telemetry.UsageRecord{
    Timestamp:   time.Now().UTC(),
    SessionID:   input.SessionID,
    SkillID:     si.Skill,
    Trigger:     telemetry.TriggerExplicit,
    ContextHash: contextHash,
    AgentType:   input.AgentType,   // non-empty ONLY under --agent
    Phase:       "none",            // hardcoded
    DurationMs:  0,                 // hardcoded
    Outcome:     telemetry.OutcomeUnknown,  // hardcoded
}
```

- `Phase`는 **항상 `"none"`** (하드코딩) — `HookInput`(`internal/hook/types.go:180`)에 **`Phase` 필드가 존재하지 않으므로** 어떤 입력으로도 채워질 수 없다.
- `AgentType`는 `--agent` flag 사용 시에만 non-empty (`types.go:198` 주석 "Custom agent name if --agent flag used").
- `Outcome`은 **항상 `OutcomeUnknown`** (하드코딩) → `hasSuccessClaim`(Outcome ∈ {success, partial})은 현재 스트림에서 **항상 false**.

**함의 (게이트 dormancy)**:
1. `hasSuccessClaim`이 항상 false → 게이트는 현재 텔레메트리 스트림에서 **결코 발화하지 않는다(dormant)**.
2. `Phase="none"` + `AgentType` 대부분 empty → `inferPathKind`는 사실상 모든 실제 세션에 대해 `unknown`을 반환 → REQ-SEG-011의 conservative `unknown` 버킷으로 빠지고 no finding.
3. 따라서 본 게이트의 가치는 **record-time writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked** (spec.md §A.4 + REQ-SEG-012). 그 writer가 code-change 세션에 `PathKind="code-change"` + `Outcome="success"` + `IsTestPass`를 set해야 게이트가 activate된다.

[HARD — originating incident 구분] lesson `L_manager_docs_false_backfill_report`의 originating incident는 **sync-phase(manager-docs)** 사건이다. 본 게이트의 path-kind 분류상 sync 세션은 **docs-only로 EXEMPT**(REQ-SEG-003)이므로, **본 게이트는 그 originating incident를 그대로 잡지 못한다**. 본 게이트가 겨냥하는 것은 그와 **관련되지만 별개인** "code-session false-success" 형태(code-change 세션에서 success 기록 + observed test-pass 없음)이다. originating incident의 sync git-state false-backfill 형태는 IMP-06 또는 writer 후속 SPEC 소관이다. `L_manager_docs_false_backfill_report`는 일반적 결함 클래스("관측하지 않은 검증을 주장")의 **동기 사례**로만 인용되며, 본 게이트가 그 정확한 사건을 탐지한다는 주장은 아니다 (spec.md §A.1 정합).

## 3. `internal/telemetry/recorder.go` — `LoadBySession` (REUSE) + `RecordSkillUsage` (append)

- `LoadBySession(projectRoot, sessionID) ([]UsageRecord, error)` (L73-105): 오늘+어제 `usage-YYYY-MM-DD.jsonl`을 읽어 SessionID 필터. nil-safe (파일 부재 시 continue, 항상 `(result, nil)`). **이것이 게이트의 단일 read 경로** (REQ-SEG-001/008).
- `RecordSkillUsage(projectRoot, r UsageRecord) error` (L38-69): append 경로, mutex 보호, daily rotation (`usage-<UTC-day>.jsonl`). `json.Marshal(r)` → 한 줄 append. **omitempty 필드 추가 시 기존 레코드 영향 없음** (marshal은 zero-value omitempty 필드를 출력 안 함).
- 저장 dir: `<projectRoot>/.moai/evolution/telemetry/`.

## 4. brownfield-extension 선례 — SPEC-HARNESS-OUTCOME-CAPTURE-001

MEMORY.md 확인: SPEC-HARNESS-OUTCOME-CAPTURE-001(completed)이 **동일 JSONL store에 brownfield `Event` omitempty 확장**을 적용한 선례 (`Event omitempty + EventTypeApplyOutcome + outcome.go RecordOutcome`). 즉:
- 동일 telemetry/observer JSONL store에 omitempty 필드를 추가하는 것은 검증된 brownfield 패턴.
- constraint #1("새 store 금지")은 **새 파일/store**를 금지하는 것이지, **기존 레코드 schema의 backward-compatible omitempty 확장**을 금지하지 않는다 → Approach A의 절반이 constraint #1과 양립 (design §0.2).
- 그 SPEC의 "scaffold + dormant, 0 production caller" 정직 framing(MEMORY.md)은 본 SPEC의 dormant framing과 동형 (design §3 scope note + spec.md §A.3). [D1-corrected 구분] 게이트의 **동작(behavior)**은 graceful degradation으로 wiring-독립이나, 게이트의 **가치(value, 표적 결함 탐지 능력)는 record-time writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked**된다 — 현재 production writer 하드코딩(`Outcome: OutcomeUnknown`/`Phase: "none"`)으로 dormant이기 때문 (§2.1).

## 5. `internal/hook/CLAUDE.md` 제약 (verified)

- **≤5s budget** (MoAI 정책; Claude Code 플랫폼 기본 10분을 5s로 tighten). 게이트는 `LoadBySession`(in-memory) 외 heavy work 금지 (REQ-SEG-008).
- **Stop self-gate**: Stop 훅은 매 turn-end마다 발화(완료 시점뿐 아님). agent-common-protocol.md § Hook Invocation Surface "Stop self-gate caveat" 정합. 본 게이트는 advisory라 self-gate 부담이 낮음 — 어떤 turn-end든 원장이 비었거나 finding 없으면 조용히 통과.
- **C-HRA-008**: 훅 핸들러는 AskUserQuestion/mcp__askuser 금지. static guard: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` = 0 (REQ-SEG-007).
- **`$CLAUDE_PROJECT_DIR` 해석**: observer.go helper 경유. 단, stop.go는 이미 `input.ProjectDir`/`input.CWD` fallback을 사용하므로 게이트는 그 `projectDir`를 재사용(새 해석 불필요).

## 6. `internal/hook/reflective_write.go` — `AnalyzeSessionAndLog` (PRESERVE)

- `AnalyzeSessionAndLog(projectRoot, sessionID)` (L333) → `AnalyzeSession` (L34) → `telemetry.LoadBySession` (L35). reflective learning은 **이미 같은 원장을 읽는다**.
- 본 SPEC의 게이트는 reflective learning과 **독립**된 advisory 단계 — `AnalyzeSessionAndLog`를 수정하지 않고, 같은 `LoadBySession` 데이터를 별도로 평가한다 (REQ-SEG-009 PRESERVE).
- 두 기능 모두 같은 세션 원장을 소비하므로, 미래에 한 번의 `LoadBySession` 결과를 공유하는 최적화는 가능하나 본 SPEC scope 밖(현재는 게이트가 독립 호출, ≤5s 내 충분).

## 7. SPEC ID collision check

- `ls .moai/specs/ | grep -iE 'STOP|EVIDENCE|LEDGER|GATE'`: 기존 `SPEC-STOP-HOOK-001`, `SPEC-GATE-001`, `SPEC-WF-AUDIT-GATE-001`, `SPEC-HARNESS-REGRESSION-GATE-001` 등 존재. `SPEC-STOP-EVIDENCE-GATE-001`은 **충돌 없음** (FREE 확인).
- `STOP-HOOK` 도메인은 `SPEC-STOP-HOOK-001`과 충돌하므로 회피. `STOP-EVIDENCE-GATE` 채택.

## 8. 데이터 모델 결정 (fork resolution summary → design §0)

| 신호 필요 | 현재 영속? | 본 SPEC 해법 (Approach C) |
|-----------|------------|---------------------------|
| 이진 증거 (test pass/fail) | NO (`Event`에 있으나 JSONL 미영속) | `UsageRecord`에 `IsTestPass`/`IsTestFail` omitempty 추가 (record-time half) + 부재 시 "관측 불가"로 graceful (REQ-SEG-010) |
| path-kind | NO | `UsageRecord`에 `PathKind` omitempty 추가 + 부재 시 `Phase`/`AgentType`에서 추론 (Approach B fallback) |

채택: **Approach C (Hybrid)** — 신규 정확성 + 레거시 graceful degradation. (design §0.1 상세, one-line why: A 단독은 레거시 false-flag로 REQ-SEG-010 위반, B 단독은 진짜 이진 증거 영구 부재.)

## 9. Cross-References

- spec.md §A.1 (originating-defect framing — related but DISTINCT), §A.3 (dormant scaffold), §A.4 (writer successor), §D (REQ-SEG-001..012)
- design.md §0 (Approach C 결정), §1-§4
- plan.md §B (risks), §F (milestones)
- `internal/hook/stop.go`, `internal/hook/post_tool_metrics.go:72-82` (production writer ground-truth, §2.1), `internal/hook/types.go` (HookInput no Phase field), `internal/telemetry/types.go`, `internal/telemetry/recorder.go`, `internal/hook/reflective_write.go`, `internal/hook/CLAUDE.md`
- successor SPEC id `SPEC-STOP-EVIDENCE-WRITER-001` (record-time evidence/path-kind writer; FREE, not yet authored)
- MEMORY.md: SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001 (scaffold-honest + omitempty brownfield 선례); lesson `L_manager_docs_false_backfill_report` (motivating case ONLY — its sync-phase incident is docs-only-EXEMPT here, NOT caught by this gate)
