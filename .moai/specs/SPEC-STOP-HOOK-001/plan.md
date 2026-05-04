---
id: SPEC-STOP-HOOK-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-STOP-HOOK-001

## 1. Overview

Claude Code Stop hook event를 활용해 orchestrator turn 종료 시 자동 테스트 차단 메커니즘을 구축. Shell wrapper + Go binary 구조로 cross-platform 호환. 기존 `/moai gate`의 16-language test runner 재사용.

## 2. Approach Summary

**전략**: Reuse-First, Hook-Wrapper-Pattern, JSON-Contract-Output.

1. 기존 gate 로직을 `internal/gate/runner.go` 또는 동등 패키지에서 재사용
2. `internal/hook/stop.go` 신규: stop_hook_active 가드 + 언어 감지 → gate 호출
3. `cmd/moai/hook.go`에 stop 서브커맨드 라우팅
4. `.claude/hooks/moai/handle-stop.sh` shell wrapper
5. `settings.json`에 Stop event entry 등록 (template + local 동기화)

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight Verification (Priority: Critical)

- [ ] 기존 `/moai gate` 로직 위치 확인 (internal/gate, pkg/gate 또는 cmd/moai)
- [ ] 16-language detection 로직 위치 확인 (internal/lang/detect.go 등)
- [ ] Claude Code v2.0+ Stop hook spec verbatim 캡처
- [ ] 기존 hook handlers structure 분석 (`handle-*.sh` 패턴)
- [ ] 현재 settings.json hook 등록 방식 verbatim 캡처

**Exit Criteria**: 재사용 가능 패키지 확인, hook 패턴 표준화

### M1 — Go Hook Handler Skeleton (Priority: High)

- [ ] `internal/hook/stop.go` 신규 작성:
  - struct `StopRequest` (JSON unmarshal: stop_hook_active, project_dir 등)
  - struct `StopResponse` (JSON marshal: decision, reason)
  - `func Handle(req StopRequest) StopResponse` 시그니처
- [ ] stop_hook_active 가드 로직: true이면 즉시 `{decision: approve, reason: "stop_hook_active guard"}`
- [ ] `MOAI_STOP_HOOK_DISABLED` 환경변수 체크 → `{decision: approve, reason: "disabled by env"}`

**Exit Criteria**: handler skeleton + recursive guard + env opt-out 동작 확인

### M2 — Language Detection + Test Runner Integration (Priority: High)

- [ ] `internal/lang/detect.go` 또는 동등 함수 호출하여 project marker 기반 언어 결정
- [ ] 16-language test command 매핑 테이블 작성 (research.md §2.4 표 활용)
- [ ] 기존 gate runner 재사용 (또는 동등 로직 inline) — 60s timeout 적용
- [ ] 인식 불가 언어는 silent pass (`{decision: approve, reason: "no test command detected"}`)

**Exit Criteria**: 16개 project marker 모두 정확 감지, test 명령 매핑 검증

### M3 — Test Execution + JSON Output (Priority: High)

- [ ] os/exec 또는 Go gate library 호출하여 test suite 실행
- [ ] context.WithTimeout 60s 적용
- [ ] non-zero exit → `{decision: block, reason: "<test failure summary>"}`
- [ ] zero exit → `{decision: approve}`
- [ ] timeout → `{decision: block, reason: "test timeout"}` + partial output

**Exit Criteria**: 의도적 fail/pass/timeout 시나리오에서 올바른 JSON 출력

### M4 — CLI Subcommand Routing (Priority: High)

- [ ] `cmd/moai/hook.go`에 `stop` 서브커맨드 추가:
  - stdin JSON read
  - `internal/hook.Handle()` 호출
  - stdout JSON write
- [ ] error path 처리 (JSON malformed, panic recover)
- [ ] verbose mode (-v flag for diagnostics)

**Exit Criteria**: `moai hook stop < input.json` 명령으로 정상 동작

### M5 — Shell Wrapper + settings.json Registration (Priority: High)

- [ ] `.claude/hooks/moai/handle-stop.sh` 신규:
  - bash header + stdin pass-through
  - `moai hook stop` 호출
  - exit code + stdout 그대로 전달
- [ ] `internal/template/templates/.claude/hooks/moai/handle-stop.sh` 동일 동기화
- [ ] `internal/template/templates/.claude/settings.json`에 Stop event entry 추가:
  ```json
  "Stop": [{ "hooks": [{ "command": "...", "timeout": 60 }] }]
  ```
- [ ] Local `.claude/settings.json`도 동일 추가 (또는 `moai update` 안내)

**Exit Criteria**: settings.json 등록 + shell wrapper 정상 호출 확인

### M6 — Cross-Platform Compatibility (Priority: Medium)

- [ ] macOS bash 호환 확인
- [ ] Linux bash 호환 확인
- [ ] Windows: shell wrapper 대신 Go binary 직접 호출 fallback path
- [ ] CI runner 3종 (ubuntu/macos/windows) 모두 PASS 확인

**Exit Criteria**: 3 OS 모두 hook 트리거 + JSON 출력 검증

### M7 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `make build` 실행 → `internal/template/embedded.go` 재생성
- [ ] CHANGELOG entry (Unreleased)
- [ ] `.claude/rules/moai/core/settings-management.md`에 Stop hook 등록 패턴 명시
- [ ] docs-site 4개국어 reference (별도 PR via /moai sync)

**Exit Criteria**: embedded.go 재생성 완료, 정책 문서 cross-ref

### M8 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md Given-When-Then 8개 시나리오 모두 PASS
- [ ] M0 baseline 대비 추가 회귀 없음 (기존 hook 동작 유지)
- [ ] Privacy: log scan에서 user content 누출 없음
- [ ] plan-auditor 검증 PASS

**Exit Criteria**: 모든 acceptance 시나리오 PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 stop.go 핵심 로직 (의사코드)

```go
type StopRequest struct {
    StopHookActive bool   `json:"stop_hook_active"`
    ProjectDir     string `json:"project_dir"`
}
type StopResponse struct {
    Decision string `json:"decision"` // "block" or "approve"
    Reason   string `json:"reason,omitempty"`
}

func Handle(req StopRequest) StopResponse {
    if req.StopHookActive {
        return StopResponse{Decision: "approve", Reason: "stop_hook_active guard"}
    }
    if os.Getenv("MOAI_STOP_HOOK_DISABLED") == "1" {
        return StopResponse{Decision: "approve", Reason: "disabled by env"}
    }
    lang := detect.ProjectLanguage(req.ProjectDir)
    cmd, ok := testCommandMap[lang]
    if !ok {
        return StopResponse{Decision: "approve", Reason: "no test command detected"}
    }
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    err := runTest(ctx, req.ProjectDir, cmd)
    if errors.Is(err, context.DeadlineExceeded) {
        return StopResponse{Decision: "block", Reason: "test timeout"}
    }
    if err != nil {
        return StopResponse{Decision: "block", Reason: summarize(err)}
    }
    return StopResponse{Decision: "approve"}
}
```

### 4.2 settings.json entry 형식

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh\"",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

### 4.3 16-language test command 매핑 (table)

(research.md §2.4 표 그대로 적용)

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 60s timeout 부족 (대형 test suite) | Medium | High | settings.json `timeout` 사용자 조정 가능 + CI는 별도 처리 |
| stop_hook_active 미구현으로 무한 루프 | Low | Critical | unit test 강제, M1에서 명시적 검증 |
| Windows shell wrapper 비호환 | Medium | High | Go binary 직접 호출 fallback |
| 기존 사용자 경험 저하 (매 turn 차단) | High | Medium | `MOAI_STOP_HOOK_DISABLED=1` 환경변수 명시 + CHANGELOG 안내 |
| 16-language detection 오감지 | Low | Medium | unknown은 silent pass, 강제 차단 안 함 |
| privacy: test output에 user data | Low | High | reason 필드는 첫 200자 truncate |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 재사용 패키지: `internal/gate/`, `internal/lang/`, 기존 hook handler 패턴
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (auto-retry vs user 개입): Claude Code의 기본 정책 따름 (block 시 사용자 개입), retry 자동화 본 SPEC 외
- **OQ2** (pre-commit 자동 vs 명시): default 자동 감지 (`git diff --cached`), 명시 환경변수는 추가 옵션
- **OQ3** (SPEC-level opt-out): 본 SPEC scope 외 (frontmatter 확장은 claude-code-guide 거부 — research.md §1.2)
- **OQ4** (실패 테스트 이름 포함): `reason` 필드에 첫 1-3개 실패 테스트 이름 (200자 truncate)

## 8. Rollout Plan

1. M1-M5 구현 후 dogfooding: 본 SPEC 자체에 Stop hook 적용
2. 5개 sample 프로젝트 (Go/Python/Node/Rust/Java)에서 회귀 테스트
3. CHANGELOG 명시 후 v2.x.0 minor release
4. 부정적 시그널 (DX 저하 보고 다수) 발견 시 default 비활성화 hotfix (`MOAI_STOP_HOOK_DEFAULT_OFF=1`)

End of plan.md (SPEC-STOP-HOOK-001).
