---
id: SPEC-STOP-EVIDENCE-GATE-001
title: "Stop-hook Verification-Evidence Completion Gate + Session Ledger Reader (advisory, warn-first)"
version: "0.1.1"
status: draft
created: 2026-06-15
updated: 2026-06-15
author: GOOS행님
priority: P2
phase: "v0.2.0 target"
module: "internal/hook, internal/telemetry"
lifecycle: spec-anchored
tags: "hooks, stop-hook, telemetry, evidence-gate, session-ledger, advisory, warn-first"
tier: M
---

# SPEC-STOP-EVIDENCE-GATE-001 — Stop-hook Verification-Evidence Completion Gate + Session Ledger Reader

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-15 | GOOS행님 | Initial draft. Plan-phase artifacts (spec/plan/acceptance/design/research). Tier M. Direct follow-up to SPEC-HOOK-DISCIPLINE-WIRING-001 (IMP-01). Realizes IMP-02 (Stop-hook evidence gate) + IMP-03 (session ledger reader). |
| 0.1.1 | 2026-06-15 | manager-spec | amend-1 (plan-auditor iter-1 PASS-WITH-DEBT 0.82 → "amend then re-audit"). D1: §A.3 dormant-scaffold framing + §A.4/REQ-SEG-012 falsifiable value condition naming successor `SPEC-STOP-EVIDENCE-WRITER-001` + §A.1 originating-defect re-framing (sync-phase incident is docs-only-EXEMPT). D5: REQ-SEG-002 reworded (`Event.IsTestPass`→`UsageRecord.IsTestPass`). D2/D3/D4: acceptance.md grep-idiom fixes. Artifact editing only — Tier M unchanged, no code. |

---

## A. Background — IMP-02/03 Realization (additive advisory layer, NOT a bug fix)

본 SPEC은 SPEC-HOOK-DISCIPLINE-WIRING-001(이하 "IMP-01", status: completed, Mx close commit `8a9c1062f`)의 **직접 후속**입니다. IMP-01은 두 개의 휴면(dormant) discipline 훅을 settings.json에 warn-first 철학으로 배선했습니다(status-transition → PostToolUse advisory; sync-phase-quality-gate → Stop warn-first; exit-2 blocking은 env 플래그 뒤에 휴면 보존). 본 SPEC은 외부 분석 로드맵의 IMP-02(게이트) + IMP-03(세션 원장)을 **하나의 SPEC으로 묶어** 구현합니다.

### A.1 The originating defect class — related but DISTINCT from `L_manager_docs_false_backfill_report`

이 SPEC의 **동기(motivation)**는 lesson `L_manager_docs_false_backfill_report`입니다: 한 에이전트(manager-docs)가 `sync_commit_sha` "backfill"을 **완료했다고 보고했으나 실제로는 여전히 placeholder**였던 사건 — 즉, 에이전트가 **실제로 관측하지 않은(NOT actually observed) 검증/완료를 주장(CLAIMED)**한 것입니다.

[HARD — 정확한 결함 형태 구분] 그 originating incident는 **sync-phase(manager-docs)** 사건이며, 본 SPEC의 path-kind 분류상 **docs-only로 EXEMPT**됩니다(REQ-SEG-003). 따라서 **본 게이트는 그 originating incident를 그대로 잡지 못합니다** — 그것은 "sync git-state false-backfill" 형태이고, 본 게이트가 겨냥하는 것은 그와 **관련되지만 별개인(related but DISTINCT)** "**code-session false-success 형태**"입니다:

> **본 게이트가 겨냥하는 결함 형태**: code-change 세션(`Phase=run`/`AgentType=manager-develop`)에서 `Outcome=success`를 기록했으나 관측된 test-pass 신호가 없는 경우.

[HARD] 본 SPEC은 originating incident(sync git-state false-backfill 형태)를 잡는다고 **주장하지 않습니다**. 그 sync-phase 형태는 본 SPEC scope 밖이며, IMP-06(baseline-integrity attribution + invariant) 또는 record-time-writer 후속 SPEC(§A.4) 소관입니다. `L_manager_docs_false_backfill_report`는 본 게이트가 표적으로 삼는 일반적 결함 클래스("관측하지 않은 검증을 주장")의 **동기 사례**로서 인용될 뿐, 본 게이트가 그 정확한 사건을 탐지한다는 주장은 아닙니다.

본 SPEC은 session-stop 시점에 기계적·**advisory** 체크를 추가하여, **code-change 세션이 주장한 성공(claimed success)이 관측된 이진 증거(observed binary evidence)로 뒷받침되지 않을 때 이를 표면화(surface)**합니다. 이 체크는 차단하지 않으며(fail-open), 구조화된 출력/stderr로만 신호를 보냅니다.

### A.3 [HARD] Knowingly-dormant scaffold (current telemetry stream에 대해 dormant)

[HARD] 본 게이트는 **현재 텔레메트리 스트림에 대해 의도적으로 휴면(knowingly-dormant)** 상태로 배송됩니다. 이는 SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001의 **scaffold-honest 선례**와 동형입니다(과대광고 금지). Ground-truth (plan-auditor 확인):

- 유일한 production `UsageRecord` writer는 `logSkillUsage`(`internal/hook/post_tool_metrics.go:72-82`)이며, **하드코딩으로** `Outcome: OutcomeUnknown`, `Phase: "none"`, `DurationMs: 0`을 기록한다.
- `HookInput`에는 `Phase` 필드가 없으므로(`internal/hook/types.go`) `Phase`는 항상 `"none"`이다. `AgentType`는 `--agent` flag 사용 시에만 non-empty이다.
- 따라서 현재 스트림에서 `hasSuccessClaim`(Outcome ∈ {success, partial})은 **항상 false** → 게이트는 현재 발화하지 않으며(dormant), `inferPathKind`는 사실상 모든 실제 세션에 대해 `unknown`을 반환한다.

[HARD] 본 게이트의 **가치(value)는 record-time-writer 후속 SPEC(§A.4)에 blocked**되어 있습니다. 그 writer가 landing하기 전까지 본 게이트는 현재 텔레메트리 스트림에 대해 dormant입니다. 본 SPEC의 deliverable는 **(a) 게이트/리더의 read-side 로직 + (b) 이진 증거/path-kind를 read하고 부재 시 graceful degradation하는 schema + 분류**이며, **현재 스트림에서 finding을 실제로 발화하는 것은 deliverable가 아닙니다**. 이 정직한 framing은 progress.md §B에도 기록됩니다.

### A.4 [HARD] Record-time-writer 후속 SPEC — falsifiable value condition

[HARD] 본 게이트가 표적 결함 클래스(code-session false-success)를 **실제로 잡으려면**, record-time writer가 다음을 충족해야 합니다:

1. (a) code-change 세션의 `UsageRecord`에 `PathKind="code-change"` (또는 code-bearing `Phase`)를 set, AND
2. (b) `Outcome="success"` + `IsTestPass`(이진 증거)를 record time에 set.

이 두 조건이 충족되기 전까지 본 게이트는 dormant이며 표적 결함을 잡지 못합니다. 이로써 "게이트가 결함을 잡는다"는 주장은 **falsifiable(반증 가능)**해집니다 — 구조적으로 반증 불가능한 주장이 아닙니다.

[HARD] 후속 SPEC id: **`SPEC-STOP-EVIDENCE-WRITER-001`** (record-time evidence/path-kind writer). 본 SPEC은 그 writer의 read-side 소비자(scaffold)이며, writer SPEC이 위 (a)(b)를 구현하면 본 게이트가 activate된다. 의존 관계는 frontmatter `depends_on` 역방향(writer가 본 SPEC을 enabler로 소비)이 아니라, **본 게이트의 가치가 writer에 blocked**되는 관계다 (REQ-SEG-012 + progress.md §dependency note 참조).

### A.2 Confirmed ground-truth (재발견하지 말 것 — DO read the files to design precisely)

1. **`internal/hook/stop.go`** — `stopHandler.Handle()`는 현재: stop을 로깅 → `input.StopHookActive == true`일 때 L46에서 `&HookOutput{}` early-return(무한 루프 가드) → telemetry pruning(90d + 30d) → reflective learning(`countSessionRecords >= minToolInvocationsForReflection`=3일 때 `AnalyzeSessionAndLog`) → 최종 L79에서 `&HookOutput{}` 반환(항상 allow). `projectDir`는 `input.ProjectDir` 우선, 비면 `input.CWD` fallback. 핸들러는 오늘 **fail-open**(stop을 절대 차단하지 않음). `countSessionRecords`는 L90에서 이미 `telemetry.LoadBySession(projectRoot, sessionID)`를 호출함.

2. **`internal/telemetry/types.go`** — `UsageRecord` JSON struct(JSONL): `Timestamp(ts)` / `SessionID(session_id)` / `SkillID(skill_id)` / `Trigger(trigger: explicit|auto)` / `ContextHash(context_hash)` / `AgentType(agent_type)` / `Phase(phase: plan|run|sync|none)` / `DurationMs(duration_ms)` / `Outcome(outcome: success|partial|error|unknown)`. 별도로 `Event` struct(JSONL 미저장 — outcome heuristic 전용): `ToolName` / `IsError` / `IsTestPass` / `IsTestFail`.

3. **`internal/telemetry/recorder.go`** — `LoadBySession(projectRoot, sessionID) ([]UsageRecord, error)`는 오늘+어제 `.moai/evolution/telemetry/usage-YYYY-MM-DD.jsonl`을 읽어 SessionID로 필터; nil-safe. `RecordSkillUsage`는 append 경로(mutex 보호, daily rotation). 저장 dir: `<projectRoot>/.moai/evolution/telemetry/`.

4. **`internal/hook/CLAUDE.md`** — 훅은 ≤5s MoAI 예산; Stop 훅은 self-gate 필요(완료 시점뿐 아니라 매 turn-end마다 발화); subagent boundary C-HRA-008: 훅 핸들러는 AskUserQuestion / mcp__askuser 호출 금지(static guard: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` = 0 매치). `$CLAUDE_PROJECT_DIR` 해석은 observer.go helper 경유.

---

## B. Problem Statement

session-stop 시점에 MoAI는 세션의 텔레메트리(원장)를 이미 적재하지만(`countSessionRecords` → `LoadBySession`), **세션이 주장한 성공이 관측된 이진 증거로 뒷받침되는지**는 검사하지 않습니다. 그 결과 `L_manager_docs_false_backfill_report` 같은 "관측하지 않은 검증을 주장" 패턴이 stop 시점에 표면화되지 않고 통과합니다.

추가로, IMP-03(세션 원장)은 **새 원장 저장소를 만들지 않고** 기존 `.moai/evolution/telemetry/usage-*.jsonl` 위에 read-only 뷰로 구축되어야 합니다. 현재 `UsageRecord`는 `Outcome`을 기록하지만 (a) 테스트 pass/fail 이진 증거를 별도 필드로 저장하지 않고, (b) 건드린 파일 경로/path-kind를 기록하지 않습니다 — 이 둘이 본 SPEC의 핵심 설계 긴장(§D 결정 사항)입니다.

---

## C. Goal

advisory, warn-first 증거 게이트 + 세션 원장 리더를 하나의 SPEC으로 구현합니다:

1. **IMP-02 — Stop-hook evidence gate**: `stopHandler.Handle()`의 기존 동작(early-return 가드 / 90d·30d pruning / reflective learning)을 **변경 없이 보존**한 채로, session-stop 시점에 세션 원장을 읽어 "성공 주장 + 이진 증거 없음" 조합을 표면화하는 **additive advisory 단계**를 추가합니다. 절대 stop을 차단하지 않습니다(fail-open).

2. **IMP-03 — Session ledger reader**: 기존 `telemetry.LoadBySession` 위에 read-only로 구축되는 세션 원장 뷰(path-kind 분류 + docs-only 버킷 포함). **새 원장 저장소 파일을 만들지 않습니다.**

3. **Path-kind classification (docs-only bucket 포함)**: 세션 작업을 path-kind로 분류하며, "docs-only" 버킷은 코드 변경과 **동일한 이진 검증 증거를 요구하지 않습니다**(문서 작업은 test-pass 요건 없음).

4. **C-HRA-008 compliance**: 게이트 코드는 AskUserQuestion / mcp__askuser를 호출하지 않습니다. static grep 가드를 acceptance criterion으로 인코딩합니다.

5. **≤5s budget**: 훅 timeout 예산 내. 테스트 재실행 / 네트워크 / stop 시점 heavy rescan 금지.

---

## D. Requirements (GEARS)

### REQ-SEG-001 — session ledger reader built on LoadBySession (Ubiquitous)

The session-ledger reader **shall** be built on top of the existing `telemetry.LoadBySession(projectRoot, sessionID)` read path and **shall not** create a new on-disk ledger storage file (no new JSONL / DB / state file).

- 근거(HARD constraint #1): 세션 원장은 기존 `.moai/evolution/telemetry/usage-*.jsonl` 위에 read-only로 구축한다. 새 저장소 파일은 IMP-01의 dormant/advisory 철학 및 single-source-of-truth를 위반한다.

### REQ-SEG-002 — binary evidence first (Ubiquitous)

When judging whether a session's "success" claim is backed, the evidence gate **shall** prefer concrete binary pass/fail signals — persisted as the new `UsageRecord.IsTestPass` / `UsageRecord.IsTestFail` fields per design §3 — over soft signals (`Outcome=success` alone).

- 근거(HARD constraint #2): 이진 증거 개념(`IsTestPass`/`IsTestFail`)은 기존 `Event` struct에 정의되어 있으나 **`Event`는 JSONL에 영속화되지 않는다**(types.go 주석). 따라서 본 SPEC은 그 두 이진 신호를 design §3에서 **`UsageRecord`의 신규 omitempty 필드**로 영속화한다 (struct-name 혼동 방지: 게이트가 read하는 것은 `UsageRecord.IsTestPass`이지 `Event.IsTestPass`가 아니다). 표적 결함 형태("code-session success claim ≠ observed binary pass")는 이진 증거를 우선 신뢰해야 표면화할 수 있다.

### REQ-SEG-003 — docs-only path-kind exemption (State-driven)

While a session's classified path-kind is `docs-only`, the evidence gate **shall not** demand binary test-pass evidence for that session (a docs-only session with no test-pass signal **shall not** be flagged as an unbacked success claim).

- 근거(HARD constraint #4): 문서 작업은 test-pass 요건이 없다. docs-only 세션을 코드 변경과 동일 기준으로 flag하면 false-positive가 폭증한다.

### REQ-SEG-004 — code-change path-kind demands evidence (State-driven)

While a session's classified path-kind is a code-change bucket (NOT `docs-only`) AND the session ledger contains at least one `success`/`partial` outcome claim, the evidence gate **shall** check for the presence of a binary pass signal and **shall** surface an advisory finding when a success claim is present but no binary pass signal is observed.

- 근거: 이것이 게이트의 핵심 detection. "success claimed + NO binary evidence" 조합(`L_manager_docs_false_backfill_report` 케이스)을 code-change 세션에서 표면화한다.

### REQ-SEG-005 — fail-open (Ubiquitous, unwanted-behavior)

The evidence gate **shall not** block the Stop event; on every path the `stopHandler.Handle()` return value **shall** remain `&HookOutput{}` (allow). Any exit-2 / blocking behavior **shall** stay dormant (behind an environment flag if introduced at all, OFF by default), matching IMP-01's sync-phase-quality-gate warn-first philosophy.

- 근거(HARD constraint #6): warn-first 점진적 도입. stop을 차단하면 사용자 세션이 예기치 않게 멈춘다. 발견 사항은 구조화된 출력 / stderr로만 표면화한다.

### REQ-SEG-006 — advisory output surface (Event-driven)

When the evidence gate detects an unbacked success claim, the gate **shall** write the advisory finding to stderr (structured, human-readable) and **shall not** alter the JSON stdout HookOutput contract that Claude Code consumes for the stop decision.

- 근거(HARD constraint #6 + C-HRA-008): Stop 훅의 stdout은 Claude Code stop 결정 채널이다. advisory 신호는 stderr 진단 채널로 분리해 stop 결정 계약을 오염시키지 않는다.

### REQ-SEG-007 — C-HRA-008 subagent boundary (Ubiquitous, unwanted-behavior)

The evidence-gate code added to `internal/hook` **shall not** invoke `AskUserQuestion` or `mcp__askuser__*`; the static guard `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` **shall** return zero matches after implementation.

- 근거(HARD constraint #5): 훅은 subagent context로 사용자 상호작용 채널이 없다. C-HRA-008은 binary constraint이며 CI guard로 강제된다.

### REQ-SEG-008 — ≤5s budget preservation (Ubiquitous)

The evidence gate **shall** complete within the existing ≤5s MoAI hook budget by reading only the already-loaded session ledger (`LoadBySession`); it **shall not** re-execute tests, make network calls, or perform a heavy filesystem rescan at stop-time.

- 근거(HARD constraint #7): stop 시점 heavy work는 사용자 세션을 stall시킨다. 게이트는 이미 적재된 원장만 in-memory로 평가한다.

### REQ-SEG-009 — behavior preservation of existing stop.go (Ubiquitous)

The existing `stopHandler.Handle()` behavior — the `StopHookActive` early-return guard, the 90d + 30d telemetry pruning, and the reflective-learning `AnalyzeSessionAndLog` invocation (gated by `minToolInvocationsForReflection`) — **shall** be preserved unchanged; the new evidence gate **shall** be purely additive within the same handler.

- 근거(behavior preservation critical): 기존 stop.go 동작 변경은 회귀 위험. 게이트는 기존 단계를 건드리지 않고 추가만 한다.

### REQ-SEG-010 — graceful degradation for legacy JSONL (Event-driven)

When session-ledger records lack the binary-evidence / path-kind fields (records written before this SPEC, or by older agents), the evidence gate **shall** degrade gracefully — treat the absent signal as "evidence not observable" (NOT as "evidence failed") and **shall not** emit a false unbacked-claim finding solely due to absent new fields.

- 근거(backward-compat): 기존 JSONL 레코드는 새 필드가 없다. 부재를 "검증 실패"로 오해하면 모든 과거 세션이 false-flag된다. 부재는 conservative하게 "관측 불가"로 처리한다.

### REQ-SEG-011 — path-kind classification taxonomy (Ubiquitous)

The session-ledger reader **shall** classify each session into exactly one path-kind bucket from a fixed taxonomy that includes at minimum: `docs-only`, `code-change`, and `unknown` (when the path-kind signal is absent or ambiguous), using the chosen data-model approach (§F design.md DD-1).

- 근거: 분류가 게이트 판단의 입력이다. 고정 taxonomy + `unknown` fallback이 REQ-SEG-010의 graceful degradation을 가능케 한다.

### REQ-SEG-012 — falsifiable value condition (dormant scaffold + named writer successor) (Ubiquitous)

The evidence gate **shall** be documented as a knowingly-dormant scaffold whose target-defect detection (code-session false-success) is BLOCKED on a named record-time-writer successor SPEC (`SPEC-STOP-EVIDENCE-WRITER-001`); the gate **shall** detect the target defect ONLY when that writer sets, at record time, both (a) `PathKind="code-change"` (or a code-bearing `Phase`) and (b) `Outcome="success"` + `IsTestPass` on a code-change session's `UsageRecord`.

- 근거(D1 falsifiability): "게이트가 결함을 잡는다"는 주장을 **반증 가능(falsifiable)**하게 만든다 — 현재 production writer(`logSkillUsage`)는 `Outcome: OutcomeUnknown` / `Phase: "none"` 하드코딩이므로 게이트는 dormant이다(§A.3). detection을 가능케 하는 정확한 record-time 조건 (a)(b)를 명시함으로써, 게이트의 가치가 구조적으로 unfalsifiable해지는 것을 방지한다. writer SPEC(`SPEC-STOP-EVIDENCE-WRITER-001`)이 (a)(b)를 구현하면 본 게이트가 activate된다. (§A.4 + progress.md §dependency note 참조.)

---

## Exclusions (What NOT to Build)

[HARD] 본 SPEC은 아래를 **포함하지 않습니다**:

### E.1 Out of Scope

1. **Coverage-relation correlation (DEFERRED)** — coverage-percentage 상관 분석은 본 SPEC scope 밖입니다. 게이트는 coverage 수치를 읽거나 상관시키지 않습니다. IMP-06 또는 후속 SPEC으로 명시 연기됩니다(HARD constraint #3).

2. **IMP-06 — baseline-integrity attribution + 5-section report format + "no unobserved-verification-claim" invariant (OUT OF SCOPE)** — baseline 무결성 귀속, 5-섹션 리포트 포맷, "관측하지 않은 검증을 주장하지 말 것" invariant는 모두 별도 후속 SPEC(IMP-06)으로 명시 연기됩니다. 본 SPEC은 그 invariant의 **기계적 advisory 표면화(detection)**만 제공하며, invariant 자체의 정책 codification은 IMP-06 소관입니다.

3. **exit-2 blocking activation (kept dormant)** — 어떤 exit-2 / blocking 경로도 본 SPEC에서 활성화하지 않습니다. 게이트는 항상 fail-open. blocking 경로를 도입하더라도(선택) env 플래그 뒤에 휴면(OFF default)으로 두며, 본 SPEC의 wired default는 절대 차단하지 않습니다(REQ-SEG-005).

4. **신규 원장 저장소 파일 (EXCLUDED)** — 새 JSONL / DB / state 파일을 만들지 않습니다. 세션 원장은 기존 `.moai/evolution/telemetry/usage-*.jsonl` 위 read-only 뷰입니다(REQ-SEG-001).

5. **Stop 시점 test 재실행 / network / heavy rescan (OUT OF SCOPE)** — 게이트는 이미 적재된 원장만 평가합니다. 새 증거 수집(테스트 실행, 파일 스캔)은 record-time(다른 훅) 책임이며 stop-time이 아닙니다(REQ-SEG-008).

6. **기존 reflective-learning / pruning 로직 수정 (PRESERVE)** — `AnalyzeSessionAndLog`, `PruneOldFiles`(90d/30d), `StopHookActive` 가드는 수정하지 않습니다. 게이트는 순수 additive입니다(REQ-SEG-009).

7. **다른 훅 이벤트의 advisory 게이트 (OUT OF SCOPE)** — 본 SPEC은 Stop 이벤트만 다룹니다. PostToolUse / SubagentStop / TaskCompleted에 동일 게이트를 추가하지 않습니다.

---

## E. Cross-References

- SPEC-HOOK-DISCIPLINE-WIRING-001 (IMP-01) — predecessor; warn-first / dormant / advisory 어휘 + sync-phase-quality-gate Stop wiring 선례
- **SPEC-STOP-EVIDENCE-WRITER-001 (successor, §A.4 + REQ-SEG-012)** — record-time evidence/path-kind writer that unblocks this gate's value (currently FREE id; not yet authored)
- SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001 — scaffold-honest precedent (dormant + 0 production caller framing; §A.3)
- `internal/hook/stop.go` — `stopHandler.Handle()` (PRESERVE 대상 + EXTEND 지점)
- `internal/hook/post_tool_metrics.go:72-82` — `logSkillUsage`, the ONLY production `UsageRecord` writer (hardcodes `Outcome: OutcomeUnknown` / `Phase: "none"` — the dormancy ground-truth, §A.3)
- `internal/hook/types.go` — `HookInput` (no `Phase` field; `AgentType` non-empty only under `--agent` — §A.3 ground-truth)
- `internal/hook/reflective_write.go` — `AnalyzeSessionAndLog` / `AnalyzeSession` (PRESERVE 대상)
- `internal/telemetry/types.go` — `UsageRecord` + `Event` (데이터 모델 fork 대상; §F design.md DD-1)
- `internal/telemetry/recorder.go` — `LoadBySession` (원장 read 경로, REUSE) + `RecordSkillUsage` (append 경로)
- `internal/hook/CLAUDE.md` — ≤5s budget / Stop self-gate / C-HRA-008 / `$CLAUDE_PROJECT_DIR` 해석
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface — Stop 훅 self-gate caveat (매 turn-end 발화)
- CLAUDE.local.md §2 (settings.local separation), §14 (하드코딩 방지 — envkeys.go)
