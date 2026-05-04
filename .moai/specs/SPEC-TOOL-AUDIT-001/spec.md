---
id: SPEC-TOOL-AUDIT-001
status: draft
version: "0.1.0"
priority: Medium
labels: [tool-audit, observability, minimalism, agent-quality, post-tool-hook, wave-3, tier-2]
issue_number: null
scope: [internal/audit, cmd/moai, internal/hook, .moai/research/tool-usage, .claude/rules/moai/quality]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-TOOL-AUDIT-001: Tool Minimalism Audit

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic "Seeing Like an Agent" 권고에 따라 sub-agent의 도구 사용률을 측정하고 미사용 도구를 review flag하는 audit 시스템 신설. Auto-removal은 명시적 금지.

---

## 1. Goal (목적)

본 프로젝트의 24 agent + 100+ skill 환경에서 도구 사용률을 측정하여 5% 미만의 저사용 도구를 review flag하고, 20+ 도구를 보유한 agent에 대해 specialization 권장을 자동 생성하는 read-only audit 시스템을 신설한다. 핵심 안전장치: human approval 없는 auto-removal 금지.

### 1.1 배경

- Anthropic blog "Seeing Like an Agent": "Low bar for tool removal, high bar for addition." / "Tool overload without specialization: Providing one agent with 40+ tools across unrelated domains causes selection errors."
- 본 프로젝트의 agent/skill 카탈로그에서 각 도구 실제 사용률 측정 부재
- 미사용 도구가 prompt에 부담으로 남아 selection 정확도 저하 위험

### 1.2 비목표 (Non-Goals)

- 자동 도구 제거 (human approval 필수)
- 도구 사용률 실시간 dashboard / metric 시각화
- skill 사용률 audit (별도 SPEC 후보)
- 도구 인자 / 응답 내용 기록 (프라이버시)
- agent 자동 분리 (audit는 권장만)
- 임계값 글로벌 변경 자동화

---

## 2. Scope (범위)

### 2.1 In Scope

- `cmd/moai/audit.go`에 `moai audit tools` 서브커맨드 추가 (또는 신규 파일)
- `internal/audit/tools/aggregator.go` 신규 — hook log → 사용률 집계
- `internal/audit/tools/report.go` 신규 — markdown report 생성기
- `.moai/research/tool-usage/audit-YYYY-MM-DD.md` 보고서 출력
- `.moai/observability/tool-invocations.jsonl` runtime 데이터 저장소
- `internal/hook/post_tool.go`에 tool invocation 기록 로직 추가 (옵션 ON)
- `.moai/config/sections/observability.yaml`에 `tool_audit.enabled` 키
- `.claude/rules/moai/quality/tool-audit.md` 정책 문서
- 5% 임계값 + 20+ 도구 임계값 자동 감지
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 자동 도구 제거 (human approval 필수)
- 실시간 dashboard
- skill 사용률 audit
- 도구 인자/응답 기록 (privacy violation)
- agent 자동 분리
- 임계값 자동 조정 알고리즘
- PR comment 자동 게시 (manual 호출만)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- 의존: PostToolUse hook (기존 인프라 재사용)
- 영향 디렉터리: `internal/audit/`, `cmd/moai/`, `internal/hook/`, `.moai/observability/`, `.moai/research/tool-usage/`
- 입력 데이터: `.moai/observability/tool-invocations.jsonl` (JSONL append-only)
- 출력 보고서: `.moai/research/tool-usage/audit-YYYY-MM-DD.md`

---

## 4. Assumptions (가정)

- A1: PostToolUse hook이 안정 작동하며 tool name, agent name 기록 가능
- A2: `.moai/observability/` 디렉토리에 쓰기 권한 보유
- A3: audit는 read-only 분석 + report 생성만 수행
- A4: 100 sessions baseline은 default, 사용자 변경 가능 (`--window 100`)
- A5: tool invocation 데이터는 JSONL append-only, atomic write 보장

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-TA-001**: THE COMMAND `moai audit tools` SHALL produce a markdown report at `.moai/research/tool-usage/audit-<DATE>.md` summarizing per-agent tool usage rates.
- **REQ-TA-002**: THE AUDIT REPORT SHALL include a "Removal Candidates" section listing tools with usage rate below 5% per agent.
- **REQ-TA-003**: THE AUDIT REPORT SHALL include a "High-Tool-Count Agents" section listing agents with 20 or more declared tools.
- **REQ-TA-004**: THE AUDIT SYSTEM SHALL NOT remove tools automatically — all removal is human-approval-only.

### 5.2 Event-Driven Requirements

- **REQ-TA-005**: WHEN `moai audit tools` is invoked, THE COMMAND SHALL aggregate tool invocations from the past N sessions (default 100) recorded in `.moai/observability/tool-invocations.jsonl`.
- **REQ-TA-006**: WHEN a tool's usage rate per agent is below 5% over the audit window, THE REPORT SHALL flag that tool as a removal candidate with explicit "review required" annotation.
- **REQ-TA-007**: WHEN an agent has 20 or more declared tools, THE REPORT SHALL recommend specialization (e.g., split into focused sub-agents).
- **REQ-TA-008**: WHEN the PostToolUse hook fires AND `tool_audit.enabled = true`, THE HOOK SHALL append a tool invocation record to `.moai/observability/tool-invocations.jsonl`.

### 5.3 State-Driven Requirements

- **REQ-TA-009**: WHILE `tool_audit.enabled = false` in observability.yaml, THE POSTTOOLUSE HOOK SHALL NOT record tool invocations (zero overhead).
- **REQ-TA-010**: WHILE the audit window contains fewer than 30 sessions, THE COMMAND SHALL emit a "data insufficient" warning and produce a partial report.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-TA-011**: WHERE the user supplies `--window <N>` flag, THE COMMAND SHALL aggregate the past N sessions instead of the default 100.
- **REQ-TA-012**: IF `.moai/observability/tool-invocations.jsonl` is malformed or missing entries, THEN THE COMMAND SHALL skip corrupt entries and log a warning to stderr.
- **REQ-TA-013**: WHERE the audit detects no removal candidates, THE REPORT SHALL state "no removal candidates" explicitly (no false silence).

### 5.5 Unwanted (Negative) Requirements

- **REQ-TA-014**: THE AUDIT SHALL NOT auto-remove any tool from any agent's frontmatter.
- **REQ-TA-015**: THE TOOL INVOCATION LOG SHALL NOT record tool arguments, return values, or any user-supplied content (only tool_name, agent_name, timestamp, duration, success).
- **REQ-TA-016**: THE AUDIT SHALL NOT split agents automatically (only recommendations in the report).
- **REQ-TA-017**: THE AUDIT REPORT SHALL NOT be auto-published to PR comments (manual review only).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Audit 보고서 생성 | `moai audit tools` 실행 | success exit |
| 5% 임계값 검출 | 의도적 low-usage tool 시드 | 100% flagged |
| 20+ 도구 감지 | 의도적 24-tool agent 테스트 | flagged |
| Auto-removal 차단 | unit test (auto-remove 호출 시) | 100% block |
| 데이터 부족 경고 | 30 미만 session 시뮬레이션 | warning emitted |
| 프라이버시 준수 | log scan for forbidden fields | 0건 발견 |
| Cross-platform | macOS/Linux/Windows | 3/3 PASS |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Auto-removal 절대 금지 (Anthropic 권고 핵심 원칙)
- C2: tool argument / return value 기록 절대 금지 (privacy)
- C3: PostToolUse hook 부담 최소화 (비동기 append, 실패 시 silent)
- C4: Template-First Rule 준수
- C5: 임계값 5% / 20 tools는 default; 향후 SPEC에서 조정 가능

End of spec.md (SPEC-TOOL-AUDIT-001 v0.1.0).
