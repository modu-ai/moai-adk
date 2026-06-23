---
id: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
title: "CLI config-system integrity / mental-model alignment (P0 fixes)"
version: "0.1.0"
status: completed
created: 2026-06-23
updated: 2026-06-24
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: "M"
tags: "cli, config, profile, update, init-wizard, mental-model, p0"
era: V3R6
---

# Acceptance Criteria — SPEC-V3R6-CLI-CONFIG-INTEGRITY-001

## §A. AC 매트릭스 (Summary)

| AC ID | REQ ID | 결함 | 심각도 | 검증 방식 |
|-------|--------|------|--------|-----------|
| AC-CCI-001 | REQ-CCI-001 | F1 help string | MUST | CLI `--help` 출력 검사 |
| AC-CCI-002 | REQ-CCI-002 | F1 doc/README | MUST | README grep + doc comment 검사 |
| AC-CCI-003 | REQ-CCI-003 | F2 중앙 테이블 | MUST | grep scatter 0건 + 테이블 존재 |
| AC-CCI-004 | REQ-CCI-004 | F2 expandModelString | MUST | 단위 테스트 |
| AC-CCI-005 | REQ-CCI-005 | F2 테스트 갱신 | MUST | launcher_test.go 통과 |
| AC-CCI-006 | REQ-CCI-006 | F3 wizard 확인 | MUST | wizard 실행 출력 검사 |
| AC-CCI-007 | REQ-CCI-007 | F3 주석 cross-ref | SHOULD | 주석 grep |
| AC-CCI-008 | REQ-CCI-008 | P0-4 template source | MUST | template grep `_TBD_` 0건 |
| AC-CCI-009 | REQ-CCI-009 | P0-4 local 전파 | MUST | local grep `_TBD_` 0건 |
| AC-CCI-010 | REQ-CCI-010 | Hardcoding 정합 | MUST | scatter 리터럴 grep |

## §B. Given-When-Then 시나리오

### AC-CCI-001 — F1 help string (MUST)

**Given** 사용자가 터미널에서 `moai update --help`를 실행한다.
**When** cobra가 `-c, --config` 플래그 help 문구를 stdout에 출력한다.
**Then** help 문구에 "init wizard"(또는 동등 의미)와 "no template sync"/"does not synchronize templates"(또는 동등 의미) 키워드가 동시에 포함된다.
**Evidence**: `internal/cli/update.go:88` flag help string.

### AC-CCI-002 — F1 doc comment + README (MUST)

**Given** 개발자가 `internal/cli/update.go`의 `runUpdate` doc comment(라인 111-122)와 프로젝트 README(`README.md`, `README.ko.md`)를 읽는다.
**When** `-c`/`--config` 플래그 설명을 확인한다.
**Then** doc comment와 README 양쪽 모두에 bare `moai update`(template sync via 3-way merge)와 `-c`(reconfigure wizard, no template update)의 구분이 명시되어 있으며, `internal/cli/update.go:166-170`의 short-circuit을 인용한다.
**Evidence**: `update.go:111-122`, README `moai update` 섹션.

### AC-CCI-003 — F2 중앙 테이블 (MUST)

**Given** `internal/template/model_policy.go`(또는 동일 패키지 파일)에 단일 `ModelAliasTable`(또는 동등 명칭)이 정의되어 있다.
**When** 아래 **canonical grep**(D3 fix — spec.md §G / plan.md §E / AC-010과 동일 pattern)을 실행한다:

```bash
grep -rnE '"(opus|sonnet|haiku|opusplan)(\[1m\])?"' internal/cli/ internal/settings/
```

관측된 live match 수(plan-audit iteration-1 + 본 fix 과정에서 직접 실행, 2026-06-24): **50건 pre-refactor**. 패턴은 닫는 따옴표 `"`를 요구하므로 `"claude-opus-4-7"` 등 full model id를 substring로 오매칭하지 않는다(구 `"opus\|"sonnet\|...` 패턴이 84건으로 팽창한 D3 escaped-pipe vacuity trap 회피).
**Then** 매칭 결과가 중앙 테이블 1소스로 수렴한다. REQ-CCI-003에 열거된 scatter-site inventory 8개 소스가 모두 중앙 테이블을 참조하도록 전환된다:

- `internal/cli/profile_setup.go:70` (`case` 라인), `:74,76,78,80,82` (`return` 리터럴 5건), `:421-426` (`huh.NewOption` 6건 — user-facing picker surface)
- `internal/cli/wizard/advanced_gate.go:143-147` (THIRD wizard gate — `workflow_team_default_model`)
- `internal/settings/schema.go:140` (`modelOptions()` 배열), `internal/settings/accessors_test.go:51,70` (test mirror — lockstep 갱신 필수)
- `internal/cli/launcher_test.go:601` (no-op 단언 → canonical-id 해석 단언으로 갱신, REQ-CCI-005)

post-refactor 기대: 비-테스트·비-`model_policy` Go 파일에서 0건 (테스트 파일의 expected-value 배열과 model_policy 테이블 자체는 남는다 — 이들은 SSOT 소스/단언이므로 제거 대상 아님).
**Evidence**: 8 소스 전부 file:line 직독으로 검증 완료(2026-06-24, verification-claim-integrity 준수).

### AC-CCI-004 — F2 expandModelString 해석 (MUST)

**Given** `expandModelString(model string)`이 중앙 테이블(REQ-CCI-003)을 참조한다.
**When** `expandModelString("opus")`, `expandModelString("sonnet")`, `expandModelString("haiku")`, `expandModelString("opusplan")`을 각각 호출한다.
**Then** 각 별칭이 중앙 테이블에 정의된 canonical Claude Code model id로 해석되어 반환된다.
**And When** `expandModelString("opus[1m]")`, `expandModelString("sonnet[1m]")`을 호출한다.
**Then** canonical id에 `[1m]` 접미사가 보존된 형태로 반환된다(pass-through 유지).
**Evidence**: `internal/cli/launcher.go:704-706`.

### AC-CCI-005 — F2 테스트 갱신 (MUST)

**Given** `internal/cli/launcher_test.go:601`이 기존 `{"opusplan", "opusplan", "opusplan"}` no-op 단언을 포함한다.
**When** F2 구현이 완료된다.
**Then** 동일 케이스가 `opusplan` 별칭의 canonical id 해석을 단언하도록 갱신되며, `opus`/`sonnet`/`haiku` 및 `[1m]` 변형을 포함한 table-driven 케이스가 추가된다.
**And** `go test ./internal/cli/...`가 통과한다.
**Evidence**: `internal/cli/launcher_test.go:601`.

### AC-CCI-006 — F3 wizard 확인 라인 (MUST)

**Given** 사용자가 profile wizard에서 `permissionMode`로 "acceptEdits"를 선택한다.
**When** wizard가 `profile_setup.go:594`에서 "acceptEdits"를 `""`로 정규화한다.
**Then** wizard 종료 전 stdout에 "acceptEdits는 project default이며 settings.local.json에 defaultMode override가 기록되지 않는다"는 취지의 명시적 확인 라인이 출력된다.
**Evidence**: `internal/cli/profile_setup.go:592-595`, `internal/cli/launcher.go:669`.

### AC-CCI-007 — F3 주석 cross-ref (SHOULD)

**Given** `internal/cli/launcher.go:645-652`의 `syncPermissionModeToSettingsLocal` 함수 주석이 빈 문자열 정규화 의도를 설명한다.
**When** 개발자가 해당 주석을 읽는다.
**Then** 주석에 REQ-CCI-006(wizard 확인 라인)에 대한 cross-reference가 포함되어, 정규화가 사용자에게 silence되지 않음을 명시한다.
**Evidence**: `internal/cli/launcher.go:645-652`.

### AC-CCI-008 — P0-4 template source (MUST)

**Given** template source `internal/template/templates/.moai/config/sections/db.yaml`이 `engine`, `orm`, `migration_tool`에 `_TBD_` placeholder를 포함한다.
**When** F4 구현이 완료된다.
**Then** template source의 해당 3개 필드가 empty string(`""`) 또는 `null`로 교체되며, `grep -rn '_TBD_' internal/template/templates/.moai/config/sections/db.yaml` 결과가 0건이다.
**And** `/moai db init` interview가 여전히 해당 필드를 채우는 정상 동작을 유지한다.
**Evidence**: template `db.yaml`(검증 완료).

### AC-CCI-009 — P0-4 local 전파 (MUST)

**Given** template source가 REQ-CCI-008에 따라 갱신되었다.
**When** 사용자가 bare `moai update`(template sync)를 실행한다.
**Then** local `.moai/config/sections/db.yaml`에 `_TBD_`가 0건이 되며, `grep -rn '_TBD_' .moai/config/sections/db.yaml` 결과가 0건이다.
**Evidence**: local `db.yaml`(검증 완료, 현재 3건 `_TBD_`).

### AC-CCI-010 — Hardcoding 정합 (MUST)

**Given** 본 SPEC이 신규 model id 리터럴을 도입 또는 수정한다.
**When** 아래 **canonical grep**(D3 fix — AC-003 / spec.md §G / plan.md §E와 동일 pattern; `internal/template/` 포함 확장)을 실행한다:

```bash
grep -rnE '"(opus|sonnet|haiku|opusplan)(\[1m\])?"' internal/cli/ internal/settings/ internal/template/
```

관측된 live match 수(2026-06-24 직접 실행): **148건 pre-refactor** (`internal/template/`의 `model_policy.go` 테이블 자체 + 빌더 코드 포함). 구 `"opus\|"sonnet\|...` 패턴(124건)과 달리 닫는 따옴표 `"` 요구로 `"claude-opus-4-7"` substring 오매칭을 배제한다.
**Then** 모든 매칭이 중앙 `ModelAliasTable`(REQ-CCI-003)의 단일 SSOT에서 비롯되며, CLAUDE.local.md §14 hardcoding-prevention을 준수한다.
**Evidence**: `internal/template/model_policy.go` 존재(검증 완료).

## §C. Edge Cases

1. **F2 — 알 수 없는 별칭**: `expandModelString("claude-fable-5")` 등 미래 별칭이 들어올 경우. → 중앙 테이블의 default 케이스가 원본을 pass-through 또는 empty string 반환(현재 `normalizeModel`의 default 패턴 준용).
2. **F2 — `[1m]` + 알 수 없는 별칭**: `"claude-fable-5[1m]"`. → default pass-through 시 접미사 보존 정책 명시 필요.
3. **F3 — non-interactive 모드**: `--yes` 플래그 또는 비-TTY 환경에서 wizard 확인 라인이 stdout에 출력되는지 여부. → 확인 라인은 stdout 프린트이므로 TTY 여부와 무관하게 출력.
4. **P0-4 — `/moai db init` interview 미실행 프로젝트**: db.yaml의 3개 필드가 계속 empty로 남아있는 경우. → empty string은 `_TBD_`보다 명시적 disabled 상태이므로 허용.

## §D. Quality Gate Criteria

- **Test**: `go test ./internal/cli/... ./internal/settings/... ./internal/template/...` 100% 통과.
- **Lint**: `golangci-lint run` clean.
- **Coverage**: 수정된 패키지 85% 이상 유지.
- **Template neutrality**: CI guard `template-neutrality-check.yaml` 통과(template source 수정분).
- **Build**: `make build` 성공(embedded template 재생성 확인).

## §E. Definition of Done

- [ ] 모든 MUST AC(001-006, 008-010) 충족
- [ ] SHOULD AC(007) 충족 또는 명시적 사유와 함께 이연
- [ ] spec.md / plan.md / acceptance.md 3파일 모두 존재
- [ ] 코드 수정이 `internal/cli/CLAUDE.md` conventions(subagent boundary, hardcoding, settings helpers) 준수
- [ ] template source 수정이 Template-First(§2) + Template neutrality(§25) 준수
- [ ] verification-claim-integrity: 모든 단언이 관측된 명령 출력에 기반(§A Evidence 인용)

## §F. Forward-Looking Checks (후속 SPEC 연계)

- P1-1 책임 매트릭스 문서화 후보: 본 SPEC F1의 help/doc 명확화가 완료된 뒤, 3개 진입점 전체 매트릭스 독립 문서의 필요성 재평가.
- P1-2 profile 기록 통합 후보: F2/F3 최소 수정 완료 후 4-way 기록 경로 통합 가능성 평가.
- P2-1 defaultMode 3-source 통합: F3 (a) 확인 라인이 사용자 오인을 해소하는지 관찰 후, (b) 값 보존 또는 3-source 통합 필요성 판단.
