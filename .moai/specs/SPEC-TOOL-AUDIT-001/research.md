# Research — SPEC-TOOL-AUDIT-001 (Tool Minimalism Audit)

**SPEC**: SPEC-TOOL-AUDIT-001
**Wave**: 3 / Tier 2 (검증 통과 — read-only audit, no auto-removal)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Seeing Like an Agent":

> "Low bar for tool removal, high bar for addition."

> "Tool overload without specialization: Providing one agent with 40+ tools across unrelated domains causes selection errors. Agents pick tools based on description match, and overlapping descriptions degrade discrimination."

> "Audit tool usage. If a tool has been declared but never invoked across realistic workloads, the description is either wrong or the tool is unnecessary."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code agent frontmatter (`tools:` field) 기준으로 검증함. 결론:

- **호환성**: ✅ 완전 지원 — `tools:` 필드는 frontmatter 표준, 사용률 측정은 hook 인프라로 가능
- **표준 우회 불필요**: 측정은 PostToolUse hook에서 tool name 집계 가능
- **권고 채택**: ACCEPT — 단, auto-removal은 강하게 거부 (human approval required)

---

## 2. 현재 상태 (As-Is)

### 2.1 agent catalog 규모

CLAUDE.md §4 "Agent Catalog":
- Manager Agents: 8 (spec, ddd, tdd, docs, quality, project, strategy, git)
- Expert Agents: 8 (backend, frontend, security, devops, performance, debug, testing, refactoring)
- Builder Agents: 3 (agent, skill, plugin)
- Evaluator Agents: 2 (evaluator-active, plan-auditor)
- Agency Agents: 2 (copywriter, designer fallback)

**Total**: 24 정의된 sub-agent 종류 (Dynamic team teammate 제외)

### 2.2 skill catalog 규모

`.claude/skills/moai*/`: 100+ skill (SPEC-V3R2 Wave 1에서 consolidation 진행)

### 2.3 tool 선언 현황

각 agent 또는 skill의 frontmatter `tools:` 또는 `allowed-tools:` 필드:

```yaml
---
tools: Read, Write, Edit, Grep, Glob, Bash, Task, TodoWrite
---
```

**관찰**:
- 대부분 agent가 동일한 8-12 도구 묶음 선언
- 사용률 측정 인프라 부재 → 어떤 agent가 어떤 도구를 실제로 자주 호출하는지 미파악
- 미사용 도구가 prompt에 부담으로 남음 (각 도구 schema가 token 차지)

### 2.4 PostToolUse hook (재사용 후보)

`.claude/hooks/moai/handle-post-tool.sh`: 매 도구 호출 후 트리거. JSON에 `tool_name`, `agent_name` 포함.

**관찰**: hook이 호출 정보를 기록하면 audit 데이터로 재가공 가능. 현재 기록은 선택적 (telemetry 켰을 때만).

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| 도구 사용률 측정 | 없음 | per agent / per tool 카운트 | 측정 인프라 |
| 사용률 임계값 | 없음 | 5% 미만은 review flag | 임계값 정의 |
| audit 보고서 | 없음 | `.moai/research/tool-usage/audit-YYYY-MM-DD.md` | 신규 산출 |
| 20+ 도구 agent 감지 | 없음 | 자동 감지 + 분리 권장 | 신규 |
| auto-removal | 없음 | **금지** (human approval) | 보호 강제 |
| 측정 윈도 | undefined | 100 sessions (default) | parameterize |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `cmd/moai/audit.go` 또는 기존 audit 명령 확장 | 신규 또는 수정 | `moai audit tools` 서브커맨드 |
| `internal/audit/tools/aggregator.go` | 신규 | hook log → 사용률 집계 로직 |
| `internal/audit/tools/report.go` | 신규 | markdown report 생성기 |
| `.moai/research/tool-usage/` | 디렉토리 신규 | 보고서 출력 위치 |
| `.claude/rules/moai/quality/tool-audit.md` | 신규 | audit 정책 문서 |

### 4.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `internal/hook/post_tool.go` | 수정 (보강) | `tool_invocations.jsonl` 누적 기록 (옵션 ON 권장) |
| `internal/template/templates/.moai/config/sections/observability.yaml` | 수정 | `tool_audit.enabled: true` 키 추가 |
| `.moai/observability/tool-invocations.jsonl` | runtime artifact | 측정 데이터 저장소 |

### 4.3 입력 데이터 schema

```jsonl
{"timestamp":"2026-04-30T12:34:56Z","agent":"manager-ddd","tool":"Edit","duration_ms":12,"success":true}
{"timestamp":"2026-04-30T12:34:57Z","agent":"manager-ddd","tool":"Read","duration_ms":3,"success":true}
```

### 4.4 출력 보고서 schema (`audit-YYYY-MM-DD.md`)

```markdown
# Tool Usage Audit — 2026-04-30

## Summary
- Sessions analyzed: 100
- Total invocations: 4523
- Agents covered: 18

## Per-Agent Tool Usage
### manager-ddd (612 invocations)
| Tool | Count | % |
|------|-------|---|
| Edit | 245 | 40% |
| Read | 198 | 32% |
| Write | 88 | 14% |
| Grep | 65 | 11% |
| TodoWrite | 16 | 3% ⚠️ (below 5% threshold) |

## Removal Candidates
- agent-X / TodoWrite: 3% usage → review for removal

## High-Tool-Count Agents (>= 20 tools)
- expert-backend (24 tools): Recommendation — split into expert-backend-api + expert-backend-data
```

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 측정 데이터 부족 (100 session 미달) | High | Medium | minimum threshold (e.g., 30 sessions) 미달 시 audit 보류 메시지 |
| 사용률 5% 미만이지만 critical edge case에서 필수 | High | High | auto-removal 절대 금지, human approval 필수 — REQ로 강제 |
| 측정 인프라가 hook 부담 가중 | Medium | Medium | JSONL append-only, 비동기 쓰기, 실패 시 silent (hook 차단 X) |
| 프라이버시 (도구 인자가 민감 정보) | High | High | tool_name + count만 기록, 인자/응답 절대 미기록 |
| 20+ 도구 agent 분리가 호출 사이트 깨짐 | Low | High | audit는 권장만, 분리는 별도 PR (이번 SPEC 외) |

### 5.2 Assumptions

- A1: PostToolUse hook이 안정 작동 중
- A2: `.moai/observability/` 디렉토리 쓰기 가능
- A3: 본 audit는 read-only 분석 + report 생성만 수행
- A4: `moai audit tools` 명령은 사용자 명시 호출 (자동 실행 X)
- A5: 100 sessions baseline은 설정 가능한 default

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| audit 보고서 생성 시간 | 100 sessions 데이터 | < 5s |
| 보고서 정확도 | manual cross-check | 100% match |
| auto-removal 시도 차단 | unit test (auto-remove 호출 시 reject) | 100% block |
| 5% 임계값 검출 정확도 | 의도적 low-usage tool 시드 | 100% flagged |
| 20+ 도구 agent 감지 | 의도적 24-tool agent | flagged |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| auto-removal 자동화 | ❌ | Anthropic 권고 위반: human approval 필수 |
| 임계값 1% (더 엄격) | ❌ | edge case 도구 잘못 제거 위험 |
| 임계값 10% (더 관대) | ❌ | minimalism 의도 약화 |
| skill 사용률도 audit | ❌ (이번 SPEC 외) | tool과 skill은 다른 추적 모델 — 후속 SPEC 후보 |
| Real-time dashboard | ❌ | over-engineering, JSONL + audit 명령으로 충분 |

---

## 8. 참고 SPEC

- SPEC-OBSERVE-001: observability 인프라 — hook log 재사용
- SPEC-TELEMETRY-001: telemetry 시스템 — JSONL 출처
- SPEC-LSP-QGATE-004: LSP-based quality gate — audit는 보완

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: 임계값 5% 외 dimension (예: avg duration > 1s) 추가 여부?
- OQ2: audit 보고서를 PR comment로 자동 게시할지, 수동 review만 할지?
- OQ3: 100 sessions의 정의 (Claude Code session vs MoAI run vs SPEC execution)?
- OQ4: 보고서 보존 정책 (1주일 / 1개월 / 영구)?

---

End of research.md (SPEC-TOOL-AUDIT-001).
