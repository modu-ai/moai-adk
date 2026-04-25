---
spec_id: SPEC-GLM-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact.기능은 PR #581 및 5개 follow-up 커밋에서
  이미 출시됨. spec.md와 실제 구현(`internal/cli/glm.go`, `launcher.go`,
  `config/defaults.go`)을 비교하여 관찰된 동작을 AC로 역도출함.
  plan-auditor 2026-04-24 감사로 acceptance.md 부재가 확인되어 backfill.
---

# Acceptance Criteria — SPEC-GLM-001

GLM 호환성 자동화(`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` + `DISABLE_PROMPT_CACHING=1` 자동 주입/제거, `glm-4.5` 모델 추가)의 관찰 가능한 인수 기준.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| REQ-1 (auto-inject in GLM modes) | AC-001 ~ AC-004 | `internal/cli/glm.go:165-177, 348-372, 497-540, 794-810` |
| REQ-2 (auto-remove on CC) | AC-005, AC-006 | `internal/cli/launcher.go:204-245`, `glm.go:374-405` |
| REQ-3 (glm-4.5 defaults) | AC-007, AC-008 | `internal/config/defaults.go:47` |
| REQ-4 (tests) | AC-009 | `internal/cli/glm_{team,model_override,new,compat}_test.go` |

## AC-001: setGLMEnv은 DISABLE_BETAS와 DISABLE_PROMPT_CACHING를 프로세스 env에 설정한다

**Given** 사용자가 `moai glm` 또는 `moai cg` 모드를 실행한 상황에서,
**When** `setGLMEnv(glmConfig, apiKey)`가 호출되면,
**Then** 현재 프로세스의 환경변수에 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`과 `DISABLE_PROMPT_CACHING=1`이 동시에 설정되어야 한다.

**Verification**: `internal/cli/glm.go:165-177` — `setGLMEnv` 내부에서 `os.Setenv("CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS", "1")` 및 `os.Setenv("DISABLE_PROMPT_CACHING", "1")` 호출 확인.

## AC-002: injectTmuxSessionEnv은 tmux 세션 env에 두 호환 변수를 포함한다

**Given** `moai cg` 모드가 tmux 세션 안에서 실행되는 상황에서,
**When** `injectTmuxSessionEnv(glmConfig, apiKey)`가 호출되면,
**Then** 해당 tmux 세션의 `set-environment` 맵에 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`와 `DISABLE_PROMPT_CACHING`이 모두 `"1"`로 주입되어야 한다.

**Verification**: `internal/cli/glm.go:348-372` — `injectTmuxSessionEnv`의 env 맵 리터럴에 `"DISABLE_PROMPT_CACHING": "1"` 포함(line 365).

## AC-003: injectGLMEnvForTeam은 settings.local.json에 두 호환 변수를 기록한다

**Given** 팀 모드(`moai glm --team` / `moai cg --team`) 실행에서,
**When** `injectGLMEnvForTeam(settingsPath, glmConfig, apiKey)`가 호출되면,
**Then** `settings.local.json`의 `env` 섹션에 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`과 `DISABLE_PROMPT_CACHING=1`이 모두 기록되어야 한다.

**Verification**: `internal/cli/glm.go:497-540` — 라인 533에서 `settings.Env["DISABLE_PROMPT_CACHING"] = "1"` 및 대응 라인에서 DISABLE_BETAS 기록. 테스트: `glm_team_test.go`.

## AC-004: buildGLMEnvVars 반환 맵은 두 호환 변수를 포함한다

**Given** GLM env 변수 맵을 생성해야 하는 호출에서,
**When** `buildGLMEnvVars(glmConfig, apiKey)`가 실행되면,
**Then** 반환된 `map[string]string`은 `"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1"` 및 `"DISABLE_PROMPT_CACHING": "1"` 키 쌍을 포함해야 한다.

**Verification**: `internal/cli/glm.go:794-810` — 리터럴 맵 초기화부에 두 키 모두 포함. 테스트: `glm_model_override_test.go`.

## AC-005: removeGLMEnv은 settings.local.json에서 두 변수를 제거한다

**Given** `settings.local.json`이 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`와 `DISABLE_PROMPT_CACHING`을 포함한 상태에서,
**When** `moai cc` 실행으로 `removeGLMEnv(settingsPath)`가 호출되면,
**Then** 두 키 모두 `settings.Env` 맵에서 삭제되고 파일이 원자적으로 rewrite되어야 한다.

**Verification**: `internal/cli/launcher.go:204-245` — `delete(settings.Env, "DISABLE_PROMPT_CACHING")` (line 241) 및 대응 DISABLE_BETAS 삭제 로직. `launcher.go:189`에서 CC 모드 진입 시 호출 확인.

## AC-006: clearTmuxSessionEnv은 tmux에서 두 변수를 제거한다

**Given** tmux 세션 env에 두 호환 변수가 등록된 상태에서,
**When** `clearTmuxSessionEnv()`가 호출되면,
**Then** `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`와 `DISABLE_PROMPT_CACHING`이 tmux `set-environment -u` 대상 리스트에 포함되어 제거되어야 한다.

**Verification**: `internal/cli/glm.go:374-405` — clear 리스트에 `"DISABLE_PROMPT_CACHING"` 포함(line 398). `launcher.go:75`에서 CC 모드 진입 시 호출.

## AC-007: DefaultGLM45 상수가 "glm-4.5" 값으로 정의된다

**Given** GLM 모델 기본값 상수 정의 모듈이 컴파일된 상태에서,
**When** 소스에서 `DefaultGLM45` 상수를 참조하면,
**Then** 값은 `"glm-4.5"` 문자열이어야 한다.

**Verification**: `internal/config/defaults.go:47` — `DefaultGLM45 = "glm-4.5"`.

## AC-008: 템플릿 llm.yaml 및 display 메시지가 glm-4.5를 참조한다

**Given** `moai glm` 실행 시 사용자가 보는 사용 가능 모델 안내에서,
**When** CLI가 모델 옵션을 출력하면,
**Then** `glm-4.5`가 가용 모델 목록/주석에 포함되어 기존 `glm-4.5-air`와 함께 노출되어야 한다.

**Verification**: `internal/cli/glm.go` 사용자 메시지 및 `internal/template/templates/.moai/config/sections/llm.yaml`의 available-models 주석.

## AC-009: 테스트 스위트가 compatibility assertions을 포함한다

**Given** GLM 관련 테스트 파일(`glm_team_test.go`, `glm_model_override_test.go`, `glm_new_test.go`, `glm_compat_test.go`)에서,
**When** `go test ./internal/cli/...`를 실행하면,
**Then** 테스트들이 `DISABLE_BETAS`와 `DISABLE_PROMPT_CACHING`의 존재/부재를 검증하며 모두 통과해야 하고, `go vet ./...`과 `make build`도 성공해야 한다.

**Verification**: 테스트 파일들 (`internal/cli/glm_compat_test.go` 등) 존재 및 CI pass. plan-auditor 2026-04-24 기준 `go test ./internal/cli/...` green.

## Edge Cases

- **EC-01**: 기존 Claude OAuth 토큰이 존재할 때 → `MOAI_BACKUP_AUTH_TOKEN`으로 백업 후 GLM 토큰으로 덮어쓰기. `removeGLMEnv` 호출 시 백업에서 복원 (`glm.go:514-520, 826-836`).
- **EC-02**: tmux 세션 밖에서 `injectTmuxSessionEnv` 호출 → `TMUX` env 미설정 시 graceful skip.
- **EC-03**: `settings.local.json` 부재 → 빈 구조체로 생성 후 env 주입.

## Definition of Done

- [x] `setGLMEnv`, `injectTmuxSessionEnv`, `injectGLMEnvForTeam`, `buildGLMEnvVars` 4곳 모두 두 변수 주입 확인 (AC-001 ~ AC-004)
- [x] `removeGLMEnv`, `clearTmuxSessionEnv` 2곳 모두 두 변수 제거 확인 (AC-005, AC-006)
- [x] `DefaultGLM45 = "glm-4.5"` 상수 정의 (AC-007)
- [x] CI green (`go test`, `go vet`, `make build`) (AC-009)
- [x] PR #581 + 5 follow-ups merged to main
