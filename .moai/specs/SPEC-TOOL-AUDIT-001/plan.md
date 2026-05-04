---
id: SPEC-TOOL-AUDIT-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-TOOL-AUDIT-001

## 1. Overview

PostToolUse hook 데이터를 활용해 sub-agent의 도구 사용률을 측정하고, 5% 미만 사용 도구 + 20+ 도구 보유 agent를 자동 감지하여 markdown report 생성. Auto-removal 절대 금지 — human approval 필수.

## 2. Approach Summary

**전략**: Read-Only-Audit, Hook-Reuse, Privacy-First, Human-Approval-Gated.

1. PostToolUse hook에 invocation 기록 (옵션 ON, 비동기 append)
2. `internal/audit/tools/` 신규 패키지: aggregator + report
3. `cmd/moai/audit.go`에 `tools` 서브커맨드 추가
4. 5% / 20+ 임계값 default 적용
5. tool argument / response 절대 미기록 (privacy)

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] PostToolUse hook 현재 구조 (`internal/hook/post_tool.go`) 검토
- [ ] hook이 받는 JSON payload schema 확인 (tool_name, agent_name 가용 여부)
- [ ] `.moai/observability/` 디렉토리 사용 사례 확인
- [ ] 기존 telemetry 시스템 (SPEC-TELEMETRY-001) 검토 → 중복 회피
- [ ] 24 agent + 100+ skill 카탈로그의 declared tools 카운트 baseline 측정

**Exit Criteria**: hook payload 확인, baseline measurement 보유

### M1 — Tool Invocation Logger (Priority: High)

- [ ] `internal/hook/post_tool.go`에 invocation logger 추가:
  - `tool_audit.enabled = true`일 때만 활성
  - `.moai/observability/tool-invocations.jsonl`에 append-only 기록
  - 기록 필드: `{ts, agent, tool, duration_ms, success}` (인자/응답 절대 미기록)
  - 비동기 / silent failure (hook 차단 X)
- [ ] `.moai/config/sections/observability.yaml`에 `tool_audit.enabled` 키 추가 (default false)
- [ ] Template-First 동기화

**Exit Criteria**: hook이 정상 기록, privacy 위반 0건

### M2 — Aggregator 구현 (Priority: High)

- [ ] `internal/audit/tools/aggregator.go`:
  - JSONL 파일 읽기 + N session window 필터링
  - per-agent / per-tool 카운트 집계
  - 사용률 = (agent's tool count) / (agent's total invocations)
  - 20+ tools 보유 agent 감지 (frontmatter `tools:` 파싱)
- [ ] session 정의: 단일 Claude Code session ID (PostToolUse payload에서 추출)
- [ ] minimum threshold 30 sessions (parameterize: `--min-sessions`)

**Exit Criteria**: aggregator unit test PASS

### M3 — Report Generator (Priority: High)

- [ ] `internal/audit/tools/report.go`:
  - markdown 템플릿 작성:
    ```markdown
    # Tool Usage Audit — <DATE>
    
    ## Summary
    - Sessions analyzed: N
    - Total invocations: M
    - Agents covered: K
    
    ## Per-Agent Tool Usage
    ### <agent-name> (X invocations)
    | Tool | Count | % |
    |...|...|...|
    
    ## Removal Candidates
    - <agent>/<tool>: X% usage → review for removal
    
    ## High-Tool-Count Agents (>= 20 tools)
    - <agent> (N tools): Recommendation — split
    
    ## No-Action Notes
    - (if no candidates) "No removal candidates found"
    ```
- [ ] 보고서 파일명: `audit-YYYY-MM-DD.md` (날짜는 명령 실행 시점)
- [ ] 출력 위치: `.moai/research/tool-usage/`

**Exit Criteria**: report 파일 정상 생성

### M4 — CLI Subcommand (Priority: High)

- [ ] `cmd/moai/audit.go`에 `tools` 서브커맨드:
  - 기본: 100 session window
  - `--window <N>`: 사용자 정의 window
  - `--min-sessions <N>`: 최소 데이터 임계값 (default 30)
  - `--output <path>`: 사용자 정의 출력 위치
- [ ] auto-remove 시도 (예: `--auto-remove` flag) → 명시적 reject + error message
- [ ] verbose mode (-v)

**Exit Criteria**: `moai audit tools` 호출 가능

### M5 — Auto-removal Block 강제 (Priority: Critical)

- [ ] `--auto-remove` flag 또는 환경변수 시도 시 명시 reject:
  ```
  ERROR: Auto-removal is prohibited per SPEC-TOOL-AUDIT-001 REQ-TA-014.
  Tool removal requires human approval. Review the audit report and modify
  agent frontmatter manually.
  ```
- [ ] unit test: auto-remove 호출 시 100% block

**Exit Criteria**: auto-removal 100% 차단 검증

### M6 — Privacy Enforcement (Priority: Critical)

- [ ] hook logger에 인자 / 응답 기록 금지 strict 적용
- [ ] code review checklist에 "log scan: no user content" 추가
- [ ] grep test: tool-invocations.jsonl에 file path / user prompt 누출 없음

**Exit Criteria**: privacy 위반 0건

### M7 — 정책 문서 + Cross-reference (Priority: Medium)

- [ ] `.claude/rules/moai/quality/tool-audit.md` 정책 문서 신규
  - audit 정책 (5%, 20 tools 임계값)
  - human approval 의무
  - privacy 정책
  - 보고서 출력 위치
- [ ] CLAUDE.md §6 Quality Gates에 cross-ref 추가
- [ ] Template-First 동기화

**Exit Criteria**: 정책 문서 작성 + cross-ref

### M8 — Cross-Platform 호환 (Priority: Medium)

- [ ] macOS/Linux/Windows에서 JSONL append 동작 확인
- [ ] Windows 특수 경로 (`%CLAUDE_PROJECT_DIR%`) 처리
- [ ] CI 3 runners PASS

**Exit Criteria**: 3 OS PASS

### M9 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 의도적 low-usage tool 시드 검출 100%
- [ ] 의도적 24-tool agent 검출
- [ ] auto-remove 차단 unit test PASS
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 JSONL schema (입력)

```jsonl
{"ts":"2026-04-30T12:34:56Z","session_id":"abc123","agent":"manager-ddd","tool":"Edit","duration_ms":12,"success":true}
```

### 4.2 사용률 계산 (의사코드)

```go
func ComputeUsageRate(invocations []Invocation, agent string, tool string) float64 {
    agentTotal := 0
    toolCount := 0
    for _, inv := range invocations {
        if inv.Agent == agent {
            agentTotal++
            if inv.Tool == tool {
                toolCount++
            }
        }
    }
    if agentTotal == 0 {
        return 0.0
    }
    return float64(toolCount) / float64(agentTotal)
}
```

### 4.3 임계값 매핑

| Threshold | Default | Override flag |
|-----------|---------|---------------|
| Removal candidate | < 5% | `--threshold-fp <X>` |
| High tool count | >= 20 | `--threshold-tools <N>` |
| Min sessions | 30 | `--min-sessions <N>` |
| Window | 100 sessions | `--window <N>` |

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 측정 데이터 부족 | High | Medium | min-sessions 30 강제, partial report 명시 |
| auto-removal 우회 시도 | Low | Critical | M5 strict block + unit test |
| privacy 위반 (인자 누출) | Low | Critical | M6 strict policy + grep test |
| hook 부담 가중 | Medium | Medium | 비동기 / silent / 옵트인 (default false) |
| 사용률 5%지만 critical edge case에서 필수 | High | High | auto-removal 절대 금지 — human review 의무 |
| Windows JSONL append race | Low | Medium | atomic write 패턴 (O_APPEND) |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 의존: PostToolUse hook (기존 인프라), `internal/hook/post_tool.go`
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (추가 dimension): 본 SPEC scope 외, 후속 SPEC 후보
- **OQ2** (PR comment 자동 게시): 명시 reject (REQ-TA-017), 사용자 명시 호출만
- **OQ3** (100 sessions 정의): Claude Code session ID (PostToolUse payload `session_id`)
- **OQ4** (보고서 보존): 1년 보존, 이후 archive (`.moai/research/tool-usage/archive/`)

## 8. Rollout Plan

1. M1-M5 구현 후 dogfooding: 본 프로젝트에서 `moai audit tools` 실행
2. 100+ session 누적 후 첫 audit run
3. 결과 분석 → human review로 미사용 도구 제거 PR 시범
4. CHANGELOG + v2.x.0 minor release

End of plan.md (SPEC-TOOL-AUDIT-001).
