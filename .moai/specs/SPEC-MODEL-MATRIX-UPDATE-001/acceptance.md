---
spec: SPEC-MODEL-MATRIX-UPDATE-001
tier: M
created: 2026-09-30
author: manager-spec
---

# acceptance.md — SPEC-MODEL-MATRIX-UPDATE-001

> 3차 개정 (wave-2, 운영자 게이트 2026-09-30): NC-2 전면 삭제 재정 반영 — 구 AC-011의 보존 단정을 **삭제** 단정으로 반전, 레거시 폴백 신설, 하행 재번호(총 16 = Tier M 한도). 검증 테스트 열거의 8→10파일 확장은 2차 개정에서 반영돼 있으며 본 개정에서 삭제 스윕 테스트를 포함해 13파일로 확장.

## §A 검증 원칙

모든 PASS 판정은 본 워크트리에서 실행한 명령과 그 관측 출력으로 귀속된다(verification-claim-integrity §1/§2). 검증 예산: `go build ./...` + grep/sed. `go test ./...` 금지(레인 규율) — 테스트 파일의 갱신 자체가 AC 대상이고, 실행 판정은 CI 몫이다.

## §B AC 매트릭스 (16)

### REQ-MMU-001 (codex 핀 신설) — AC-001..004

- **AC-MMU-001**: Given 템플릿 workflow.yaml(`internal/template/templates/.moai/config/sections/workflow.yaml`)과 로컬 도그푸드 workflow.yaml(`.moai/config/sections/workflow.yaml`)의 audit 블록을 읽을 때, When 양쪽의 `codex` 하위 키를 확인하면, Then 템플릿은 `model: gpt-6.1-sol` + `effort: high`를 담고, 로컬은 `model: gpt-6.1-sol`이다(로컬의 구형 `gpt-5.6-sol` 잔존 0행).
- **AC-MMU-002**: Given `internal/config/defaults.go`의 `NewDefaultWorkflowConfig`를 읽을 때, When `Audit` 블록을 확인하면, Then `Codex: ModelEffort{Model: "gpt-6.1-sol", Effort: "high"}`가 설정돼 있다.
- **AC-MMU-003**: Given `internal/cli/mcp_codex.go`의 `resolveCodexAuditModelEffort`를 읽을 때, When 프로젝트 핀과 명시 인자가 모두 빈 경로를 추적하면, Then 단말 폴백이 `{gpt-6.1-sol, high}`를 반환하는 구조(상수 또는 동등 기본값)로 바뀌어 있다 — `internal/cli/model_backend_default_test.go`의 zero-value 단정(`:36-37`)이 같은 방향으로 갱신돼 있다.
- **AC-MMU-004**: Given `internal/cli/mcp_codex.go`를 읽을 때, When `codexServableModelPrefixes`를 확인하면, Then 집합은 `{"gpt-", "o1", "o3", "o4", "codex"}` 그대로다(무변경 — `gpt-6.1-sol` 통과는 기존 접두사로 충분).

### REQ-MMU-002 (claude 핀 승격) — AC-005..007

- **AC-MMU-005**: Given `internal/cli/mcp_claude.go` 상단 상수와 템플릿·로컬 workflow.yaml의 `audit.claude`를 읽을 때, When 각 표면의 model/effort를 확인하면, Then `claudeAuditDefaultModel`/`claudeAuditDefaultEffort` 상수와 템플릿·로컬 양쪽 핀이 모두 `{claude-opus-5-5}` / `{medium}`으로 일치한다(결정 기록 DR-1 확정값).
- **AC-MMU-006**: Given `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`와 그 템플릿 미러를 읽을 때, When "defaults to" 감사 기본 서술 행(`:70`)을 확인하면, Then 두 사본이 같은 커밋에서 새 값을 서술하며 byte-parity가 유지된다.
- **AC-MMU-007**: Given `internal/web/assets/i18n.js`를 읽을 때, When `f.workflow.audit.claude.model.desc`(en :586, ko :1481, ja :2248, zh :3015)와 `f.workflow.audit.claude.effort.desc`(en :588, ko :1483, ja :2250, zh :3017) 8개 행을 확인하면, Then **4 로케일 전부**에서 model 기본값과 effort 기본값 문구가 새 값을 반영한다 — 하나의 로케일이라도 구형(sonnet/high 기본) 서술을 유지하면 이 AC는 FAIL이다(plan-audit D4).

### REQ-MMU-003 (glm effort 통일) — AC-008..009

- **AC-MMU-008**: Given `internal/settings/schema_sections.go`의 `glmDefaultTierEffort`를 읽을 때, When 4티어(high/medium/low/fable) 각각을 평가하면, Then 모든 티어가 `template.GLMStateMax`를 반환하고, 그 근거 주석(`:169-172` 구획)이 전 티어 max 기준으로 갱신돼 있다.
- **AC-MMU-009**: Given 템플릿 llm.yaml의 effort 블록(`:55-95`)을 읽을 때, When 내용을 확인하면, Then Claude-effort→z.ai 상태 collapse map이 변화 없이 유효하다(verify-only — tier 기본값 서술이 없어 갱신 대상이 아님, plan-audit D6 재분류).

### REQ-MMU-004 (구형 모델 전면 삭제 + 레거시 폴백) — AC-010..012

- **AC-MMU-010**: Given `internal/config/closed_sets.go`의 `ValidGLMModels`를 읽을 때, When 반환 집합과 주석을 확인하면, Then 집합은 정확히 {glm-5.3-flash, glm-5.3}이고, 철수·삭제 기록 주석이 존재하며, `internal/config/defaults.go:249-254`의 보존 근거 주석이 삭제 기록으로 대체돼 있다.
- **AC-MMU-011**: Given 변경 후 트리에서 `grep -rn "DefaultGLM45\|DefaultGLM46\|DefaultGLM47\|DefaultGLM45Air\|DefaultGLM51\|DefaultGLM52\|DefaultGLM5Turbo" internal/`과 `grep -rn "Models.Opus\|Models.Sonnet\|Models.Haiku" internal/`를 실행할 때, When 출력을 관측하면, Then **양쪽 모두 0행**이다 — 상수(defaults.go:255-261), statusline 구형 항목(memory.go:36-42), legacy struct 필드(types.go:361-364 + defaults.go:971-974 + glm.go:746-753·847 소비자 + defaults_test.go:362-379 접근)가 전부 삭제됐다(DR-2 — 보존 잔존은 AC 위반이다; grep-B 범위는 cli/ 한정이 아니라 internal/ 전체, plan-audit iter-2 N2).
- **AC-MMU-012**: Given 삭제된 모델 id(glm-4.7 등)를 tier 슬롯 값으로 지명하는 llm.yaml을 읽을 때, When 해당 값이 해석되는 경로(런처 또는 로더의 resolved-value 소비 지점)를 읽을 때, Then 티어 기본 모델로 폴백하며 1행 경고를 출력하는 구현이 존재한다(fail-open — 침묵 통과·하드 오류 어느 쪽도 아님, DR-2 관리 장치).

### REQ-MMU-005/006 (불변) — AC-013

- **AC-MMU-013**: Given `internal/statusline/memory.go`의 glmContextWindows(`:34-35`)와 `internal/template/templates/.claude/settings.json.tmpl`의 `"model"` 키를 읽을 때, When 변경 전후를 비교하면, Then glm-5.3·glm-5.3-flash 모두 `1_000_000`이고 `"model": "sonnet"`도 동일 바이트로 유지된다(두 keep-invariant — 변경 전후 동일).

### REQ-MMU-007 (동반 갱신) — AC-014..015

- **AC-MMU-014**: Given 변경 후 트리에서 `grep -rn "gpt-5.6-sol" internal/ .moai/config/`를 실행할 때, When 출력을 관측하면, Then 적중 0행이다(테스트 픽스처 전면 갱신 포함).
- **AC-MMU-015**: Given 감사·집합·삭제 대상을 단정하던 테스트 17파일 — `audit_models_test.go`, `mcp_audit_config_test.go`, `audit_pin_test.go`, `model_backend_default_test.go`, `mcp_codex_audit_pin_test.go`, `factory_operational_fixture_test.go`, `internal/settings/audit_pin_fields_test.go`, `schema_sections_test.go`, `internal/web/glm_tier_test.go`, `internal/config/defaults_test.go`, `internal/web/schema_select_preserve_test.go`, `internal/cli/glm_persist_gate_test.go`, `internal/template/glm_effort_overlay_test.go`, `internal/cli/glm_team_test.go`, `internal/cli/glm_autocompact_test.go`, `internal/cli/glm_max_context_test.go`, `internal/cli/cg_mode_hardening_test.go` — 을 읽을 때, When 각 파일이 새 핀·새 집합·삭제 후 기준을 단정하는지 확인하면, Then 17파일 모두 갱신돼 있다. 단, 마지막 3파일(glm_autocompact/glm_max_context/cg_mode_hardening)은 열거 갱신이 아니라 **삭제 후 동작 기준의 단정 재작성**이다 — 윈도우 해석 리터럴(5.2→1M, 5.1→200K, 4.7→128K)이 삭제로 소멸하기 때문이다(plan-audit iter-2 N1).

### REQ-MMU-008 (Template-First · 빌드) — AC-016

- **AC-MMU-016**: Given 모든 변경이 착지한 뒤 `make build`와 `go build ./...`를 실행할 때, When 두 exit code를 관측하면, Then 모두 0이다(make build는 agents-emit-check 포함, 임베드 템플릿 재생성 수반).

## §C 엣지 케이스

| 케이스 | 기대 동작 |
|---|---|
| 기존 프로젝트가 workflow.yaml에 구형 codex 핀(gpt-5.6-sol)을 유지 | 프로젝트 핀이 우선 — 강제 마이그레이션 없음(핀 우선순위 불변). 신설 기본값은 핀 부재 프로젝트에만 적용 |
| 기존 llm.yaml이 glm-4.7을 tier 슬롯 값으로 지명 | 티어 기본 모델로 폴백 + 1행 경고(DR-2 관리 전환 — 구형 값으로의 계속 적용은 없다) |
| 기존 llm.yaml이 레거시 별칭 키(`opus:`/`sonnet:`/`haiku:`)를 지명 | 비엄격 로더 기존 의미대로 무오류 무시(필드 소멸 — 수용된 위험의 무음 절반, DR-2) |
| 기존 llm.yaml이 glm-4.5를 지명하고 statusline이 렌더 | 구형 테이블 항목이 삭제됐으므로 내장 매칭 불일치 — `llm.glm.context_windows` 사용자 오버라이드가 유일한 지정 경로(키 유지), 미설정 시 일반 절차 |
| 사용자가 web 위젯에서 glm effort를 high로 저장 | 저장 상태로 유효(선택 집합 유지). wire 수렴은 기존 collapse 오버레이가 지배 — 본 SPEC 변경 밖 |
| claude 감사 핀에 존재하지 않는 id 입력 | 해석기 검증 없이 통과하는 기존 동작 유지(fail-open) — 단, 첫 감사 스폰 수용 여부는 E8 관찰 항목(DR-1) |
| audit.glm.model web 셀렉트의 기존 저장값이 삭제된 id | 표시 유지(fail-open 렌더), 새 저장만 제공 집합으로 제한 — web 스키마의 기존 동작 수용 |

## §D 품질 게이트

- TRUST 5 — Tested: AC-MMU-015(테스트 동반 갱신) · Readable: 갱신 주석이 glm-5.2 선례 문체 유지 · Unified: gofmt · Secured: 핀 값은 하드코딩 허용 영역인가 — **아니다**: 모델 id는 defaults.go/CLI 상수로의 추출이 기존 관례(claudeAuditDefault* 패턴 준용, §14 하드코딩 방지) · Trackable: Conventional Commits + 카드 id.
- LSP 게이트: run-phase 임계(0 error/type-error/lint-error).

## §E Definition of Done

1. §B 전 AC에 대한 관측 증거가 `.moai/reports/t1368/`에 존재한다.
2. `make build` + `go build ./...` 녹색(AC-MMU-016).
3. REQ-MMU-005/006 불변 재측정 기록(AC-MMU-013).
4. plan.md 결정 기록(DR-1/2/3, operator gate 2026-09-30)이 착지돼 있고, DR-1의 값이 AC-005·DR-2의 삭제+폴백이 AC-010..012에 반영돼 있다.
5. E8 관찰 항목(첫 감사 스폰 수용)의 확인 결과가 완료 보고에 서술로 기록돼 있다.
6. 미푸시 커밋 수와 브랜치/HEAD가 완료 보고에 기록돼 있다(통합은 리드 주도 창 경유).
