# Research — SPEC-CONTEXT-INJ-001 (Memory Persistence 단순화 / Context Injection)

**SPEC**: SPEC-CONTEXT-INJ-001
**Wave**: 3 / Tier 2 (검증 통과 — orchestrator 명시 주입으로 우회)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Claude Managed Agents Memory":

> "Memory on Managed Agents mounts directly onto a filesystem... Stores can be shared across multiple agents with different access scopes."

Anthropic blog "Harnessing Claude's Intelligence":

> "The agent's context window is its working memory. Anything outside the window must be retrieved deliberately. Persistent memory closes that gap by mounting a filesystem the agent can read across invocations."

> "Memory on Managed Agents is automatic; on Claude Code sub-agents, it is the orchestrator's responsibility to inject relevant context at spawn time."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code sub-agent context model 기준으로 검증함. 결론:

- **호환성**: ⚠️ 부분 지원 — Managed Agents Memory mount는 Claude Code sub-agent 모드에 미적용. sub-agent는 매번 fresh context.
- **표준 우회**: orchestrator가 Agent() spawn prompt에 progress.md, lessons, 최근 trace 명시 주입 (이미 본 프로젝트가 부분 적용 중)
- **권고 채택**: ACCEPT — 단, "auto memory mount" 환상 포기, 명시 주입 정책 표준화

---

## 2. 현재 상태 (As-Is)

### 2.1 Memory persistence 인프라

기존 메모리 위치:
- `~/.claude/projects/<hash>/memory/` (per-user, persistent across conversations)
- `<project>/.moai/specs/<ID>/progress.md` (per-SPEC, 명시 작성)
- `<project>/.moai/research/observations/` (design domain 한정 learnings)
- `~/.moai/worktrees/<repo>/memory/` (worktree별 — 미정확)

### 2.2 현재 sub-agent context 주입 방식

orchestrator가 `Agent(subagent_type, prompt)` 호출 시:

- prompt에 본문 (task description) 포함
- 일부 SPEC reference (`SPEC-XXX 참조`) 텍스트로 명시
- progress.md 내용 명시 주입은 ad-hoc

**관찰**: SPEC progress.md, related memory, domain lessons를 표준 절차로 주입하는 일관된 정책 부재. 토큰 예산 cap도 부재 → 잘못 주입 시 sub-agent context overflow 위험.

### 2.3 progress.md 활용 현황

`.moai/specs/<ID>/progress.md` 파일은 일부 SPEC에서 작성 중이나:
- 형식 비표준 (markdown free-form)
- orchestrator가 sub-agent 호출 시 자동 로드 X
- 토큰 사이즈 큰 경우 주입 정책 부재

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| sub-agent 메모리 자동 마운트 | ❌ (불가) | 명시 주입 표준화 | 패러다임 전환 |
| progress.md 자동 로드 | ❌ | orchestrator가 5000-token cap으로 주입 | 정책 신설 |
| 주입 우선순위 | undefined | progress.md > recent feedback > domain lessons | 명시 |
| 토큰 예산 cap | undefined | 5000 tokens per agent invocation | 신규 cap |
| 주입 정책 문서 | 부재 | `.claude/rules/moai/development/context-injection.md` | 신규 문서 |
| 형식 표준 | free-form | progress.md schema 권고 | 권장 schema |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/rules/moai/development/context-injection.md` | 신규 | Memory injection 정책 문서 |
| `internal/template/templates/.claude/rules/moai/development/context-injection.md` | 신규 | Template-First |

### 4.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/agents/moai/manager-*.md` (orchestrator-callable) | 수정 (cross-ref) | "spawn 시 context-injection.md 정책 준수" |
| `CLAUDE.md` §14 또는 §16 | 수정 (cross-ref) | 정책 문서 위치 안내 |
| `.claude/skills/moai-foundation-core/SKILL.md` | 수정 (Token Budget 절 보강) | 5000-token cap 명시 |

### 4.3 progress.md schema (권장)

```markdown
# Progress — SPEC-XXX

## Last Action
<one-line summary>

## State
- Files touched: <list>
- Next step: <one-line>

## Lessons (during this SPEC)
- <bullet>

## References
- @MX:NOTE locations
- Related SPECs
```

토큰 fingerprint: 약 250-750 tokens (1-3KB chars × 0.25 conversion ratio).

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 5000-token cap이 부족하여 핵심 context 누락 | Medium | High | 우선순위 정책 (progress > feedback > lessons) + truncation 시 경고 로깅 |
| orchestrator가 정책 미준수 (텍스트 가이드만) | High | Medium | Skill body 명시 + sub-agent 호출 시 prompt 자동 주입 helper (별도 SPEC 후보) |
| progress.md 부재 SPEC | High | Low | `.moai/specs/<ID>/progress.md` 부재 시 silent skip, 경고 없음 |
| 주입 텍스트가 user-facing prompt와 충돌 | Medium | Medium | injection 섹션을 prompt 앞부분 명시 마커 (`<!-- injected-context -->`)로 분리 |
| 정책 위반 탐지 불가 (LSP 검증 X) | High | Medium | living document + 분기별 audit (별도 SPEC) |

### 5.2 Assumptions

- A1: `Agent()` spawn prompt는 텍스트 형태이며 orchestrator가 자유 구성
- A2: progress.md는 작성 시점에 토큰 효율적인 양 (5000 tokens 이하) 유지 가능
- A3: 5000-token cap은 일반 sub-agent 호출 토큰 예산의 ~2.5% (200K 기준), ~0.5% (1M Opus 4.7 기준)에 합리적 fit
- A4: 우선순위 정책에서 "recent feedback"은 ~/.claude/projects/<hash>/memory/MEMORY.md 발췌

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 정책 문서 존재 | file existence | EXISTS |
| 우선순위 명시 | 문서 검토 | 3-tier 정의 |
| 5000-token cap 준수 | 샘플 5 호출 측정 (tiktoken cl100k_base) | <= 5000 tokens |
| sub-agent 응답 품질 | 주입 ON/OFF 비교 5쌍 | 주입 시 동등 이상 |
| 정책 적용 sub-agent 비율 | manager-*/expert-* 카운트 | 100% (cross-ref 명시) |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| Managed Agents Memory mount 강제 | ❌ | Claude Code sub-agent에 미지원 |
| 자동 주입 helper Go 구현 | ❌ (현재 SPEC 외) | 코드 변경 vs 문서 정책 — 우선 정책 표준화, 자동화는 후속 SPEC |
| progress.md 강제 schema | ❌ | living document 형식 자유 보장, 권장 수준 유지 |
| 토큰 cap 10000 tokens로 확대 | ❌ | sub-agent context overflow 위험, 5000 tokens가 안전 |
| skills/agent-memory/ skill 신설 | ❌ | rules 위치가 적합 (skill은 동작) |

---

## 8. 참고 SPEC

- SPEC-MEMO-001: 기존 memory 시스템 (재참조)
- SPEC-MEM-SCOPE-001 (이번 wave sibling): 4-level memory scope 아키텍처 — 본 SPEC의 정책 적용 대상
- SPEC-PERSIST-001: persistent state 관리 — progress.md 위치 합의

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: 5000-token cap을 Skill 형태로 LSP 검증 가능한가? → plan.md
- OQ2: progress.md의 권장 schema를 hard rule로 격상할 시점은?
- OQ3: domain lessons 출처가 multiple memory file일 때 우선순위 결정 알고리즘?
- OQ4: 주입 정책이 적용되지 않는 예외 케이스 (e.g., research-only sub-agent) 정의?

---

End of research.md (SPEC-CONTEXT-INJ-001).
