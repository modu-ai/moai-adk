---
id: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
title: "CLI config-system integrity / mental-model alignment (P0 fixes)"
version: "0.1.0"
status: draft
created: 2026-06-23
updated: 2026-06-24
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: "M"
tags: "cli, config, profile, update, init-wizard, mental-model, p0"
---

# SPEC-V3R6-CLI-CONFIG-INTEGRITY-001

## HISTORY

- 2026-06-23: 최초 draft. `moai profile setup` 실행 후 사용자가 요청한 종합 분석(4-track 병렬 Explore + `internal/cli/update.go` 직독)에서 P0 결함 4종 발견. 본 SPEC은 고영향·저비용 P0 수정만 소유하며, 구조적 P1/P2 작업은 명시적으로 제외(## Exclusions 참조).

## §A. 문제 서술 (Problem)

사용자가 `moai profile setup`을 실행한 뒤 `moai profile` / `moai update -c` / `moai init` 세 진입점과 `.moai/config` / `settings.json` / `settings.local.json` 구조에 대한 종합 분석을 요청했다. 코드 검증 결과, 사용자의 mental model과 코드 동작이 **4개 지점**에서 갈라진다. 각각은 저비용·고영향(사용자 혼란/데이터 손실 방지)이므로 P0로 분류한다.

### F1 / P0-1 — `moai update -c` help/doc 오해

`-c`/`--config` 플래그 help 문구는 "Edit project configuration (same as init wizard)"라고 표기하지만, 사용자는 이를 `.moai/config`의 template 동기화로 잘못 해석한다. 실제로는 `runInitWizard(cmd, true)`(reconfigure 모드)로 short-circuit하며 template 재렌더링/재적용은 수행하지 않는다. template 동기화는 bare `moai update`(`runTemplateSync`, 3-way merge)가 담당한다.

**증거**:
- `internal/cli/update.go:88` — `updateCmd.Flags().BoolP("config", "c", false, "Edit project configuration (same as init wizard)")`
- `internal/cli/update.go:128` — `editConfig := getBoolFlag(cmd, "config")`
- `internal/cli/update.go:166-170` — `if editConfig { return runInitWizard(cmd, true) }` (주석: "Handle --config / -c mode (edit configuration only, no template updates)")

**수정 방향**: help 문구 + `update.go:111-122` 명령 doc comment + README를 `-c`는 init wizard 재실행(GitHub/GitLab 자격증명·project mode 편집)이며 template 동기화는 bare `moai update`임을 명확히 기재. 행위 rename은 P2(Out of Scope).

### F2 / P0-2 — profile model alias 확장 no-op (RESIDUAL-RISK 좁힘)

`expandModelString()`(launcher.go:704-706)은 사실상 no-op(`return model`)다. 주석(700-703)은 `[1m]` 접미사가 Claude Code native 지원이므로 pass-through한다고 명시하므로, **"alias가 Claude Code에 전달되지 않는다"는 framing은 부정확**하다(`opus[1m]`은 유효 CC model string).

**실제 residual-risk (검증 후 좁힘)**:
1. 별칭 `opus`/`sonnet`/`haiku`/`opusplan`이 CC 런타임 버전에 따라 canonical id로 해석되지 않을 수 있음. 현재 매핑 테이블이 없어 잠재적 silent-failure.
2. 별칭→id 매핑이 코드 전역에 scatter될 위험(`profile_setup.go:70`, `settings/schema.go:140`, `launcher_test.go:601`에 동일 리터럴 배열이 반복). 이는 CLAUDE.local.md §14 hardcoding-prevention 위반.
3. `launcher_test.go:601`(`{"opusplan", "opusplan", "opusplan"}`)이 현재 no-op 동작을 단언하므로, 별칭 해석 추가 시 테스트 갱신 필요.

**수정 방향**: `internal/template/model_policy.go` 인근에 단일 중앙 alias→canonical-id 테이블을 두고 `expandModelString()`이 이를 참조. model id는 release마다 변경되므로(`opus-4-8`, `claude-fable-5` 등) 매핑은 반드시 단일 SSOT 테이블이어야 함.

### F3 / P0-3 — defaultMode "acceptEdits" 정규화 투명성

wizard는 `profile_setup.go:594`에서 "acceptEdits"를 `""`로 정규화하고, `syncPermissionModeToSettingsLocal`(launcher.go:645-694)는 mode가 "acceptEdits"가 아닐 때만 settings.local.json에 기록한다. "acceptEdits"를 선택한 사용자는 아무것도 기록되지 않은 것을 보게 되어 "내 선택이 반영되지 않았다"고 오인한다.

**증거**:
- `internal/cli/profile_setup.go:592-595` — `if permissionMode == defaultPermissionMode { permissionMode = "" }` (주석: "acceptEdits is the project default, so store empty string")
- `internal/cli/launcher.go:669` — `if permissionMode != "" && permissionMode != "acceptEdits" {` (조건부 기록/삭제)

**수정 방향**: (a) 빈문자열 정규화를 유지하되 wizard에 명시적 확인 라인 추가("acceptEdits는 project default — settings.local.json에 미기록함"), 또는 (b) 명시적 값 보존. 최소놀라움(minimal-surprise) 옵션 선택. P2-1 3-source 통합은 Out of Scope.

### P0-4 — db.yaml `_TBD_` placeholder

local `.moai/config/sections/db.yaml`에 미해결 `_TBD_` 값 3종(`engine`, `migration_tool`, `orm`)이 존재. template source(`internal/template/templates/.moai/config/sections/db.yaml`)에도 동일 `_TBD_`가 있으므로 Template-First Rule(§2)에 따라 template source에서 수정.

**증거**:
- local `.moai/config/sections/db.yaml` — `engine: _TBD_`, `orm: _TBD_`, `migration_tool: _TBD_`
- template `internal/template/templates/.moai/config/sections/db.yaml` — 동일 3종 `_TBD_` (주석: "[System] Primary database engine (set during /moai db init interview)")

**수정 방향**: `_TBD_`를 concrete 기본값 또는 명시적 disabled 상태로 교체. `/moai db init` interview가 값을 채우는 구조라면 placeholder를 빈 문자열 또는 `null`로 변경하여 YAML linter/소비자가 placeholder를 real 값으로 오인하지 않게 함.

## §B. 범위 (Scope)

### In Scope (4 P0 결함)

- F1: `update -c` help/doc/README 명확화 (행위 변경 없음)
- F2: profile model alias → canonical id 중앙 테이블 + `expandModelString()` 확장
- F3: defaultMode "acceptEdits" wizard 확인 라인 (또는 값 보존)
- P0-4: db.yaml `_TBD_` placeholder 제거 (template source + local 동기화)

### Out of Scope

`## Exclusions` 섹션 참조.

## §C. GEARS 요구사항 (Requirements)

> 본 SPEC은 GEARS(current notation)를 사용. 모든 REQ는 §A 증거 file:line을 인용.

### REQ-CCI-001 (F1 — update -c help 명확화, Ubiquitous)

The `moai update -c` / `--config` flag help string shall unambiguously state that the flag re-runs the init wizard to edit project configuration (GitHub/GitLab credentials, project mode) and does NOT synchronize templates — template synchronization is the bare `moai update`.

### REQ-CCI-002 (F1 — update.go doc comment + README, Ubiquitous)

The `update.go` command doc comment (lines 111-122) and the project README shall explicitly distinguish `-c` (reconfigure wizard, no template update) from bare `moai update` (template sync via 3-way merge), citing the short-circuit at `internal/cli/update.go:166-170`.

### REQ-CCI-003 (F2 — central alias→canonical-id table, Ubiquitous)

The profile model alias resolution shall reference a single central alias→canonical-id table located at or near `internal/template/model_policy.go`, eliminating the scattered duplicate alias literals across the verified scatter-site inventory (8 source locations, file:line verified by direct read on 2026-06-24):

- `internal/cli/profile_setup.go:70` — `normalizeModel` switch `case` line embedding all 6 aliases + 1 `[1m]` variant as a single literal array (original SPEC citation)
- `internal/cli/profile_setup.go:74,76,78,80,82` — the 5 `return "opus"` / `"opus[1m]"` / `"sonnet"` / `"sonnet[1m]"` / `"haiku"` reverse-mapping literals inside `normalizeModel` itself (the function this SPEC refactors)
- `internal/cli/profile_setup.go:421-426` — 6 `huh.NewOption(label, "alias")` literals in the wizard select widget; these are the USER-FACING model picker values (the canonical alias surface users see)
- `internal/cli/wizard/advanced_gate.go:143-147` — 3 `Value: "sonnet"`/`"opus"`/`"haiku"` literals + `Default: "sonnet"` in a THIRD wizard gate (the `workflow_team_default_model` advanced gate — previously unmentioned surface)
- `internal/settings/schema.go:140` — `modelOptions()` literal array `[]string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}` (original SPEC citation)
- `internal/settings/accessors_test.go:51,70` — test mirrors embedding the same 6-alias expected-value array; these MUST update in lockstep when `schema.go:140` is refactored to call the central table
- `internal/cli/launcher_test.go:601` — `{"opusplan", "opusplan", "opusplan"}` no-op assertion (original SPEC citation; updated by REQ-CCI-005)

### REQ-CCI-004 (F2 — expandModelString uses table, Event-driven)

**When** `expandModelString(model string)` is invoked with a non-empty alias (`opus`, `sonnet`, `haiku`, `opusplan`, or any `[1m]` variant), the function shall resolve the alias to its canonical Claude Code model id via the central table (REQ-CCI-003), and shall pass through the `[1m]` suffix unchanged when the suffix is present.

### REQ-CCI-005 (F2 — test update for no-op→resolution, State-driven)

**While** `expandModelString()` transitions from no-op to table-backed resolution, the existing assertion at `internal/cli/launcher_test.go:601` (`{"opusplan", "opusplan", "opusplan"}`) shall be updated to assert the resolved canonical id rather than the raw alias, and additional table-driven cases shall cover every alias enumerated in the central table.

### REQ-CCI-006 (F3 — acceptEdits wizard confirmation, Event-driven)

**When** the user selects "acceptEdits" as `permissionMode` in the profile wizard, the wizard shall emit an explicit confirmation line stating that "acceptEdits" is the project default and that `settings.local.json` will not receive a `defaultMode` override, so the user does not perceive the selection as a no-op.

### REQ-CCI-007 (F3 — syncPermissionMode transparency comment, Ubiquitous)

The `syncPermissionModeToSettingsLocal` function comment (`internal/cli/launcher.go:645-652`) shall cross-reference REQ-CCI-006 so future readers understand the empty-string normalization is intentional and is surfaced to the user via the wizard confirmation, not silently.

### REQ-CCI-008 (P0-4 — db.yaml template source fix, Where)

**Where** the template source `internal/template/templates/.moai/config/sections/db.yaml` carries `_TBD_` placeholder values for `engine`, `orm`, `migration_tool`, the template source shall replace `_TBD_` with explicit disabled/empty state (empty string or `null`) so downstream consumers and YAML linters do not interpret the placeholder as a real value, while preserving the `/moai db init` interview contract that fills these fields.

### REQ-CCI-009 (P0-4 — db.yaml local sync, Event-driven)

**When** the template source db.yaml is updated per REQ-CCI-008, a subsequent `moai update` (bare, template sync) shall propagate the placeholder removal to the local `.moai/config/sections/db.yaml`, leaving no residual `_TBD_` values in the local file.

### REQ-CCI-010 (Hardcoding-prevention conformance, Ubiquitous)

All model id literals introduced or modified by this SPEC shall reside in the single central table (REQ-CCI-003), conforming to CLAUDE.local.md §14 hardcoding-prevention, and no new scattered model-id literals shall be introduced in `internal/cli/`, `internal/settings/`, or `internal/template/`.

## §D. 제약사항 (Constraints)

- **행위 변경 최소**: F1은 help/doc만 수정, 코드 행위 변경 없음. F3의 (a) 옵션(확인 라인 추가)이 (b)(값 보존)보다 우선.
- **Template-First (CLAUDE.local.md §2)**: db.yaml `_TBD_` 수정은 template source에서 먼저. local은 `moai update`로 동기화.
- **Hardcoding prevention (§14)**: 모든 model id 리터럴은 단일 중앙 테이블.
- **Template neutrality (§25)**: template source 수정 시 내부 SPEC ID/REQ 토큰/commit SHA 누출 금지.
- **Verification-claim integrity**: 본 SPEC의 모든 결함 단언은 file:line 직독으로 검증됨(§A 증거).
- **Go code 표준**: `internal/cli/` 수정은 `internal/cli/CLAUDE.md` subagent boundary(C-HRA-008), error wrapping, cross-platform 준수.

## §E. Out of Scope (Exclusions)

> 구조적 P1/P2 작업은 별도 후속 SPEC으로 이연. 본 SPEC은 P0 정합만 소유.

### Out of Scope — P1-1 책임 매트릭스 문서

- 3개 진입점(`moai profile` / `moai update -c` / `moai init`)의 책임 매트릭스 독립 문서. 본 SPEC은 F1의 help/doc 명확화만 수행.

### Out of Scope — P1-2 profile→runtime 기록 통합

- 4-way profile 기록 경로를 단일 `applyProfileToProject()`로 통합. 본 SPEC은 F2/F3의 최소 수정만 수행.

### Out of Scope — P1-3 statusline 기본 세그먼트 축소

- statusline 기본 세그먼트 15→5-6 축소. 본 SPEC과 무관.

### Out of Scope — P1-4 namespace 보존 dry-run 가시성

- namespace 보존 dry-run 가시성 개선. 본 SPEC과 무관.

### Out of Scope — P2-1 defaultMode 3-source 통합

- `profile_setup.go` / `settings.json` / `settings.local.json` 3-source defaultMode 통합. F3은 wizard 확인 라인만 추가.

### Out of Scope — P2-2 TemplateContext 변수 추가

- `Framework`, `ProjectDescription` 등 TemplateContext 변수 추가. 본 SPEC과 무관.

### Out of Scope — P2-3 init→update -c handoff guidance

- init 종료 후 `update -c` handoff 안내. 본 SPEC은 help/doc 명확화만.

### Out of Scope — P2-4 language registry 16-vs-23 정합

- 16-vs-23 language registry 정합. 본 SPEC과 무관.

### Out of Scope — F1 행위 rename

- `update -c` 플래그를 `update --reconfigure` 등으로 rename하는 행위 변경. F1은 help/doc/README만 수정.

## §F. 가정 (Assumptions)

- `normalizeModel()`(`profile_setup.go:67-87`)의 역방향(full-id→alias) 매핑이 본 SPEC의 정방향(alias→id) 매핑과 충돌하지 않는다. 충돌 시 중앙 테이블에서 양방향을 통합 관리.
- `/moai db init` interview가 `engine`/`orm`/`migration_tool`을 채우는 기존 계약은 유지되며, placeholder 제거는 interview 동작에 영향을 주지 않는다.

## §G. 검증 가능 조항 (Testable Hooks)

- F1: README grep `-c` → "init wizard" + "no template" 키워드 동시 존재.
- F2: `model_policy.go` 주변 단일 테이블 존재 + scatter 리터럴이 테이블 1소스로 수렴. **Canonical grep (D3 fix, 동일 pattern을 plan.md §E / acceptance.md AC-003·AC-010에 사용 — 3 아티팩트 단일 패턴 수렴)**:

  ```bash
  grep -rnE '"(opus|sonnet|haiku|opusplan)(\[1m\])?"' internal/cli/ internal/settings/
  ```

  관측된 live match 수(plan-audit iteration-1 직접 실행, 2026-06-24): **50건**. 모든 매칭은 REQ-CCI-003 scatter inventory의 8개 소스에 대응하며, `claude-opus-4-7` 등 full model id를 substring로 오매칭하지 않음(ERE가 닫는 따옴표 `"`를 요구하므로 `"claude-opus-4-7"`의 `"opus` prefix는 매칭 불가 — D3 escaped-pipe vacuity trap 회피). post-refactor 기대값: 비-테스트·비-`model_policy` Go 파일에서 0건.
- F3: `profile_setup.go` wizard 확인 라인 존재 + `launcher.go:645` 주석에 REQ-CCI-006 cross-ref.
- P0-4: `grep '_TBD_' .moai/config/sections/db.yaml internal/template/templates/.moai/config/sections/db.yaml` → 0건.

## §H. Cross-References

- `internal/cli/CLAUDE.md` — subagent boundary, hardcoding prevention, settings.json mutation helpers.
- `CLAUDE.local.md` §2 (Template-First), §14 (Hardcoding prevention), §22 (settings.local.json separation), §25 (Template neutrality).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12 canonical frontmatter fields.
- 후속 SPEC(미번호): P1-1 책임 매트릭스 문서, P1-2 profile 기록 통합, P2-1 defaultMode 3-source 통합.
