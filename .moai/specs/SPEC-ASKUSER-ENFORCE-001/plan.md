---
id: SPEC-ASKUSER-ENFORCE-001
version: "1.0.0"
status: draft
created: "2026-04-25"
updated: "2026-04-25"
author: GOOS
priority: high
issue_number: 0
depends_on: []
---

# SPEC-ASKUSER-ENFORCE-001 Implementation Plan

> 구현 전략, 영향 받는 파일, 단계별 실행 계획
> 최종 갱신: 2026-04-25
> Phase 개수: **8** (Phase A → H)

---

## 구현 전략

본 SPEC은 코드를 추가하지 않는다. 분산된 규범을 **canonical reference**(신규 `askuser-protocol.md`) 한 곳으로 모은 뒤, 4개 기존 핵심 지침 파일이 이 canonical reference를 cross-reference하도록 보강한다. 이후 Template-First 원칙에 따라 `internal/template/templates/*` 미러를 동기화하고 `make build`로 임베디드 산출물을 갱신한다.

### 전략 원칙

1. **Single Source of Truth**: 새로 작성하는 `askuser-protocol.md`가 모든 절차의 단일 출처이다. 기존 4개 파일은 핵심 사실을 간결히 명시하고 askuser-protocol.md를 cross-reference한다 — 본문 중복을 최소화한다.
2. **Additive Change**: 기존 본문은 삭제하지 않는다. 신규 절차(ToolSearch preload, Socratic interview 구조 제약, blocker report 형식)는 기존 `User Interaction Boundary`, `MoAI Orchestrator`, §7 Rule 5, §8, output-styles §3 / §10 단락에 **추가**된다.
3. **Backward Compatibility**: 기존 `AskUserQuestion` 사용 패턴은 모두 유효하다. 본 SPEC은 의무를 추가하지만, 새로운 호출 형태나 인자 변경을 도입하지 않는다.
4. **Template-First Discipline**: `internal/template/templates/*` 동기화를 별도 phase(Phase F)로 분리하여 누락을 방지한다. `make build` 검증을 마지막 단계로 강제한다.
5. **Memory Graduation Hygiene**: dead lesson은 삭제 대신 `[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001]` 마커로 표기한다 (moai-constitution.md Lessons Protocol additive 원칙 준수).
6. **Self-Dogfood**: 본 SPEC 자체가 SPEC-WF-AUDIT-GATE-001의 audit gate 기준을 충족하는지 self-check한다 (plan-auditor 8 criteria).

### 비-목표 (Non-goals)

- `AskUserQuestion` 도구 자체 구현 변경 — Claude Code 런타임 영역
- 신규 CLI 플래그, 신규 agent, 신규 skill 도입
- 본 SPEC 외부의 다른 trigger / 다른 ambiguity 정의 일반화
- `evaluator-active` 또는 `plan-auditor`의 채점 기준 변경

---

## 영향 받는 파일

### 핵심 지침 (5)

| 파일 | 변경 유형 | 변경 범위 |
|------|-----------|-----------|
| `CLAUDE.md` | 부분 수정 | §7 Rule 5 (ToolSearch precondition 추가, 첫 행동 시퀀스 명시), §8 (User Interaction Architecture 단락 보강) |
| `.claude/rules/moai/core/moai-constitution.md` | 부분 수정 | `MoAI Orchestrator` 단락 강화 — ToolSearch preload 의무 + askuser-protocol.md cross-reference |
| `.claude/rules/moai/core/agent-common-protocol.md` | 부분 수정 | `User Interaction Boundary` 단락을 양방향(orchestrator 의무 + subagent 금지)으로 codify, blocker report 형식 명시 |
| `.claude/output-styles/moai/moai.md` | 부분 수정 | §3 Stage 1 행동 시퀀스에 ToolSearch select step 삽입, §10 Output Rules에 free-form interrogative anti-pattern 명시 |
| `.claude/rules/moai/core/askuser-protocol.md` | **신규 생성** | canonical reference, 7개 섹션 (Channel Monopoly / ToolSearch Preload / Socratic Interview / Option Description / Orchestrator–Subagent Boundary / Ambiguity Triggers / Free-form Circumvention) |

### Template 미러 (5)

| 파일 | 변경 유형 |
|------|-----------|
| `internal/template/templates/CLAUDE.md` | 원본과 동일하게 동기화 |
| `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | 원본과 동일하게 동기화 |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | 원본과 동일하게 동기화 |
| `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` | **신규 생성** (원본 미러) |
| `internal/template/templates/.claude/output-styles/moai/moai.md` | 원본과 동일하게 동기화 |

### 빌드 산출물 (1)

| 파일 | 변경 유형 |
|------|-----------|
| `internal/template/embedded.go` | `make build` 후 자동 갱신 — 수동 편집 금지 |

### Memory (1)

| 파일 | 변경 유형 |
|------|-----------|
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` | `[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001]` 마커 추가 + canonical reference cross-link |
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` | 인덱스 항목에 `[GRADUATED to SPEC-ASKUSER-ENFORCE-001]` 표시 추가 |

---

## 단계 (Phases)

### Phase A: askuser-protocol.md 신규 작성 (canonical text)

**목표**: 단일 출처 canonical reference 파일 작성.

**산출물**: `.claude/rules/moai/core/askuser-protocol.md`

**섹션 구성** (각 섹션은 REQ 1개 이상에 대응):

1. `## Channel Monopoly` — REQ-AUE-001, REQ-AUE-007 본문
2. `## ToolSearch Preload Procedure` — REQ-AUE-002 본문 + 일반화 표현 ("deferred tool requires ToolSearch select preload")
3. `## Socratic Interview Structure` — REQ-AUE-003 본문 (4 questions / 4 options / `(권장)` / conversation_language)
4. `## Option Description Standards` — REQ-AUE-004 본문 (description 필수 항목 + bias prevention)
5. `## Orchestrator–Subagent Boundary` — REQ-AUE-005 본문 (양방향 codify + blocker report 형식)
6. `## Ambiguity Triggers and Exceptions` — REQ-AUE-006 본문 (4 trigger + 5 exception)
7. `## Free-form Circumvention Prohibition` — REQ-AUE-007 본문 (Other 옵션 안내 포함)

**의존성**: 없음 (root phase).

**검증**: AC-AUE-008-5 (7개 섹션 존재), AC-AUE-001 ~ AC-AUE-007 (canonical 표현 등장).

### Phase B: CLAUDE.md §7 Rule 5 / §8 보강

**목표**: 기존 ambiguity trigger 정의를 유지하면서 ToolSearch preload step과 첫 행동 시퀀스를 추가한다.

**변경 위치**:
- §7 Rule 5 "Discovery process" 7-step 시퀀스의 두 번째 step("Conduct Socratic interview via AskUserQuestion") 직전에 새 step "First, preload AskUserQuestion schema via ToolSearch(query: \"select:AskUserQuestion\")" 삽입.
- §8 "AskUserQuestion is the ONLY User Question Channel" 단락 끝에 ToolSearch precondition 단락 추가 + askuser-protocol.md cross-reference.

**의존성**: Phase A 완료 (cross-reference 대상).

**검증**: AC-AUE-008-1, AC-AUE-006.

### Phase C: moai-constitution.md MoAI Orchestrator 단락 강화

**목표**: `MoAI Orchestrator` 섹션의 Rules 목록에 ToolSearch preload 의무를 추가하고 askuser-protocol.md를 canonical reference로 연결한다.

**변경 위치**: `MoAI Orchestrator` 단락의 Rules 리스트 끝에 신규 항목 2개 추가.
- "[HARD] AskUserQuestion is a deferred tool — invoke `ToolSearch(query: \"select:AskUserQuestion\")` immediately before each AskUserQuestion call"
- "Canonical reference: `.claude/rules/moai/core/askuser-protocol.md`"

**의존성**: Phase A 완료.

**검증**: AC-AUE-008-2.

### Phase D: agent-common-protocol.md User Interaction Boundary 양방향 codify

**목표**: 기존 subagent 금지 규범에 orchestrator 의무를 대칭으로 추가하고, blocker report 형식을 명시한다.

**변경 위치**: `User Interaction Boundary` 단락 본문.
- 기존 "Subagents MUST NOT prompt the user" 유지.
- 신규: orchestrator 측 규범 단락 추가 — "Orchestrator MUST preload AskUserQuestion via ToolSearch ... see askuser-protocol.md"
- 신규: subagent의 올바른 응답 형식 — "Return a structured blocker report with `## Missing Inputs` section listing required parameters" 명시.
- 신규: orchestrator의 후속 재위임 절차 — "On receiving a blocker report, orchestrator runs an AskUserQuestion round, injects results into a fresh subagent prompt, re-delegates."

**의존성**: Phase A 완료.

**검증**: AC-AUE-008-3, AC-AUE-005.

### Phase E: output-styles/moai/moai.md §3 Stage 1 / §10 Output Rules 갱신

**목표**: turn-by-turn 행동 절차 문서에 ToolSearch select step과 free-form anti-pattern을 명시한다.

**변경 위치**:
- §3 Stage 1 "Clarify"의 행동 시퀀스 — "Ask via AskUserQuestion" 한 줄 직전에 새 step "First: ToolSearch(query: \"select:AskUserQuestion\")" 삽입.
- §10 Output Rules 끝에 신규 anti-pattern 항목 추가 — "Free-form interrogative prose in response body is prohibited as a question channel. Use AskUserQuestion."

**의존성**: Phase A 완료.

**검증**: AC-AUE-008-4.

### Phase F: Template 미러 동기화 + make build

**목표**: `internal/template/templates/*` 5개 미러를 원본과 byte-for-byte 일치시키고, `make build`로 `internal/template/embedded.go`를 갱신한다.

**절차**:
1. 5개 원본 파일 → 5개 미러 파일로 복사 (`cp` 또는 `Read` + `Write`).
2. `make build` 실행. 빌드 실패 시 phase 중단.
3. `git diff internal/template/embedded.go` 확인 — 5 파일 변경분만 반영되었는지 검증.
4. `go build ./...` 실행으로 회귀 부재 확인.

**의존성**: Phase A ~ E 완료.

**검증**: AC-AUE-009-1, AC-AUE-009-2, AC-AUE-009-3.

### Phase G: Memory Dead Lesson SUPERSEDED 처리

**목표**: 기존 dead lesson을 정식 SUPERSEDED 처리하고, MEMORY.md 인덱스에 graduation 표시를 추가한다.

**절차**:
1. `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md`:
    - 본문 상단에 `> [SUPERSEDED by SPEC-ASKUSER-ENFORCE-001]` 줄 추가.
    - 본문 끝에 `> Graduated to canonical guidance: see .claude/rules/moai/core/askuser-protocol.md` 줄 추가.
    - 원본 lesson 본문은 변경 금지 (additive only — moai-constitution.md Lessons Protocol).
2. `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`:
    - "AskUserQuestion Enforcement" 항목 hook 텍스트를 "v3.4.0 CLAUDE.local.md §19" → "GRADUATED to SPEC-ASKUSER-ENFORCE-001 — see askuser-protocol.md"로 갱신.

**의존성**: Phase A 완료 (cross-reference 대상 존재).

**검증**: AC-AUE-010, AC-AUE-011.

### Phase H: 통합 검증 + Lint + Self-Audit

**목표**: 모든 AC를 단일 검증 시퀀스로 실행하고, plan-auditor의 8 criteria를 self-audit한다.

**절차**:
1. AC-AUE-012의 bash 검증 시퀀스 실행 (acceptance.md Part E 참조).
2. `golangci-lint run ./...` 실행 — 회귀 부재 확인.
3. `go vet ./...` 실행.
4. plan-auditor 8 criteria self-audit (본 plan.md 하단 "Self-Audit Summary" 섹션 작성).

**의존성**: Phase A ~ G 완료.

**검증**: AC-AUE-012 + 모든 AC 재실행.

---

## 테스트 전략

본 SPEC은 코드를 추가하지 않으므로 **테스트는 정합성 검증(grep / diff / build)에 한정**된다. 그러나 acceptance 단계에서는 정합성 검증을 사실상 integration test 형태로 운영한다.

### TDD 적용 가능 영역

본 SPEC의 모든 검증은 declarative — 즉, "각 파일에 해당 마커가 존재하는가" 형태의 grep 검증이다. TDD의 RED-GREEN-REFACTOR 사이클은 다음과 같이 변형 적용된다:

- **RED 변형**: 각 Phase 시작 전 acceptance.md의 해당 AC grep 명령을 실행하여 fail을 확인.
- **GREEN 변형**: Phase 작업 수행 후 동일 grep 명령이 pass됨을 확인.
- **REFACTOR 변형**: 본 SPEC의 본문 / 미러 / canonical reference의 어휘 일관성을 재점검 (특히 한국어/영어 혼용 표현이 일관되도록).

### 검증 명령 모음 (acceptance.md AC-AUE-012 기반)

```bash
set -euo pipefail
# Phase A 검증
test -f .claude/rules/moai/core/askuser-protocol.md
grep -cE '^## (Channel Monopoly|ToolSearch Preload|Socratic Interview|Option Description|Orchestrator.*Subagent|Ambiguity Triggers|Free-form)' .claude/rules/moai/core/askuser-protocol.md  # ≥ 7
# Phase B-E 검증
grep -q 'ToolSearch.*select:AskUserQuestion' CLAUDE.md
grep -q 'ToolSearch' .claude/rules/moai/core/moai-constitution.md
grep -qE 'orchestrator.*ToolSearch' .claude/rules/moai/core/agent-common-protocol.md
grep -q 'ToolSearch.*select' .claude/output-styles/moai/moai.md
# Phase F 검증
diff -q CLAUDE.md internal/template/templates/CLAUDE.md
diff -q .claude/rules/moai/core/askuser-protocol.md \
        internal/template/templates/.claude/rules/moai/core/askuser-protocol.md
# Phase G 검증
grep -q '\[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001\]' \
        ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md
# Phase H 검증 (회귀 부재)
go build ./...
go vet ./...
```

### TRUST 5 매핑

| Pillar | 적용 |
|--------|------|
| Tested | grep / diff 기반 정합성 검증이 모든 AC에 매핑됨 (acceptance.md) |
| Readable | canonical reference 단일 출처 + 4개 보조 파일 cross-reference로 중복 최소화 |
| Unified | 5 파일이 동일 어휘로 동일 사실을 표현 (AC-AUE-008-1 ~ -5) |
| Secured | 본 SPEC은 사용자 입력 처리 변경 없음, OWASP 영향 없음 |
| Trackable | 모든 변경이 단일 SPEC ID로 추적, memory dead lesson도 graduation 추적 |

---

## 마이그레이션

본 SPEC은 **기존 사용 패턴을 변경하지 않으므로 마이그레이션이 필요 없다**. 다만 다음 두 가지 운영적 변화가 발생한다:

1. **Orchestrator의 첫 행동**: AskUserQuestion 호출이 예정된 turn에서, orchestrator는 호출 직전에 ToolSearch select를 추가한다. 사용자에게 노출되는 형태는 변경 없음 (AskUserQuestion 모달은 동일하게 표시됨).
2. **Memory dead lesson 정리**: `feedback_askuserquestion_enforcement.md`에 SUPERSEDED 마커가 추가되며, 새로운 lesson은 작성되지 않는다 (이미 canonical guidance로 graduate됨).

---

## 백워드 호환성

- 기존 `AskUserQuestion` 호출 코드 / 패턴 / 인자 형식 — 모두 유효.
- 기존 `CLAUDE.md` §7 Rule 5 / §8 본문 — 모두 유지, 추가만 발생.
- 기존 `agent-common-protocol.md` `User Interaction Boundary` "Subagents MUST NOT prompt the user" 규범 — 유지.
- 기존 `output-styles/moai/moai.md` §3 Stage 1 / §10 Output Rules — 유지, 추가만 발생.
- 기존 memory dead lesson — 본문 변경 없음, additive marker만 추가.

---

## Self-Audit Summary (plan-auditor 8 criteria 대비)

본 SPEC이 SPEC-WF-AUDIT-GATE-001의 audit gate를 통과할 수 있도록 self-check한다.

| Criterion | 결과 | 근거 |
|-----------|------|------|
| 1. EARS format 준수 | PASS | REQ-AUE-001 ~ 009가 5개 EARS 패턴(Ubiquitous / Event-driven / State-driven / Unwanted) 모두 활용 |
| 2. Acceptance criteria 측정 가능 | PASS | 20개 AC 모두 grep / diff / build exit code로 자동 검증 가능 |
| 3. Out-of-scope 명시 | PASS | spec.md §4.2에 5개 항목 명시, plan.md "비-목표"에 4개 항목 재명시 |
| 4. Risk + Mitigation | PASS | spec.md §5에 6개 risk 매트릭스, 각 risk마다 구체 완화 |
| 5. Cross-reference 단일 출처 | PASS | askuser-protocol.md가 canonical, 4개 파일이 cross-reference |
| 6. Backward compatibility | PASS | 본 plan.md "백워드 호환성" 섹션에 5개 보장 명시 |
| 7. SPEC 자체 dogfood | PASS | self-audit 섹션 (본 표) + AC-AUE-012 통합 검증 |
| 8. Bias prevention | PASS | REQ-AUE-004에 description bias prevention, REQ-AUE-003에 first-option `(권장)` 강제 |

---

## 참조

- `.claude/rules/moai/core/moai-constitution.md` Lessons Protocol — graduation 절차
- `CLAUDE.local.md` §2 Template-First Rule — 미러 동기화 의무
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` — graduation 대상 dead lesson
- SPEC-WF-AUDIT-GATE-001 — 독립 SPEC, 본 SPEC과 cross-overlap 부재 (Return §4 참조)
