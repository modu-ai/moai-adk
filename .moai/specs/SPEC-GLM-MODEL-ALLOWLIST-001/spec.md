---
id: SPEC-GLM-MODEL-ALLOWLIST-001
title: "GLM 모드 모델 allowlist 정합 — enforceAvailableModels와 GLM 모델 활성화 충돌 해소"
version: "0.2.0"
status: completed
created: 2026-06-22
updated: 2026-07-01
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tags: "glm, model-policy, settings-allowlist, cost-lever, regression"
tier: M
depends_on: [SPEC-CC2178-MODEL-POLICY-REPAIR-001]
---

# SPEC-GLM-MODEL-ALLOWLIST-001 — GLM 모드 모델 allowlist 정합

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-22 | manager-spec | 최초 작성 (plan-phase draft) — `enforceAvailableModels` Claude-backend 비용 레버와 GLM 모델 활성화 메커니즘의 충돌로 모든 `moai glm` 세션이 Sonnet으로 폴백되는 회귀 정합. Approach A(settings.local.json 런타임 오버라이드) 채택. |
| 0.2.0 | 2026-06-22 | manager-spec | iter-2 — plan-auditor iter-1 FAIL 0.74 5개 결함 반영. **D1: 채택 접근을 Approach C(정적 템플릿 allowlist 확장)로 전환** — Approach A는 solo `moai glm` 경로의 의도적 "settings.local.json clean" 설계(#676 회귀)와 충돌하고 solo→plain-`claude` 종료 시 제거 경로가 없어 누출 클래스를 재도입. D2 근본원인 진술 정정. D3 미러 테스트명 정정 + model-policy.md 미러 등록. D4 `tier`/`depends_on` 프론트매터 추가. D5 4번째 GLM env 사이트(`injectGLMEnv`) out-of-scope 명시. |

## §A. Context (배경 및 검증된 근본 원인)

### A.1 사용자 가시 증상

두 개의 실제 프로젝트(`~/moai/claude.mo.ai.kr`, `~/moai/MoAI-Cowork-Plugins`)에서
`moai glm` 세션 진입 시 다음 경고가 출력되고, 설정된 GLM 모델 대신 Claude Sonnet으로
조용히 폴백되는 현상이 관찰되었다:

```
⚠ Model "opus[1m]" is restricted by your organization's settings. Using claude-sonnet-4-6 instead.
```

`moai glm`로 GLM 백엔드 세션을 띄워도 활성 모델이 GLM 모델(`glm-5.2[1m]`)이 아니라
Claude Sonnet으로 강제 전환된다. 이는 GLM 비용 최적화 목적 자체를 무력화한다.

### A.2 검증된 인과 사슬 (이번 세션 Read/grep으로 전수 확인)

1. **템플릿이 전 프로젝트에 비용 레버를 배포한다.**
   `internal/template/templates/.claude/settings.json.tmpl:377-379`은 모든
   `moai init`/`moai update` 프로젝트에 다음을 배포한다:
   ```json
   "model": "sonnet",
   "availableModels": ["sonnet", "opus", "haiku"],
   "enforceAvailableModels": true
   ```
   이 레버는 SPEC-CC2178-MODEL-POLICY-REPAIR-001(commit 2509871bd, "M4 Default-model
   cost lever (CC 2.1.175)")이 **Claude 백엔드 비용 레버**로 도입한 것이다.

2. **CC 2.1.176 강제 의미론이 alias 리다이렉트를 차단한다.**
   `internal/template/templates/.claude/rules/moai/development/model-policy.md:87`에
   verbatim으로 기록된 대로: _"alias model picks can no longer be redirected to a
   blocked model via ANTHROPIC_DEFAULT_*_MODEL env vars."_

3. **GLM 활성화가 바로 그 리다이렉트다.**
   `internal/cli/glm.go:190-205` `setGLMEnv`는
   `ANTHROPIC_DEFAULT_OPUS_MODEL = glmConfig.Models.High`(기본값 `glm-5.2[1m]`,
   `internal/config/defaults.go:46,57)를 설정한다. 팀 경로(`injectGLMEnvForTeam`,
   glm.go:595-638)도 동일 env 변수를 `settings.local.json` env에 기록한다.

4. **충돌.** `glm-5.2[1m]`(UI 표면상 `"opus[1m]"`)는 `availableModels`
   `["sonnet","opus","haiku"]`에 포함되지 않으므로, `enforceAvailableModels:true`가
   리다이렉트를 차단하고 첫 번째 허용 모델(sonnet)로 폴백한다. 이것이 A.1의 증상이다.

5. **GLM 활성화 경로(glm.go/launcher.go)는 allowlist를 건드리지 않는다 (정정 D2).**
   `enforceAvailableModels` / `availableModels`에 대한 Go 참조는 **settings.local.json
   /GLM 쓰기 경로(glm.go / launcher.go)에 ZERO**다. 유일한 Go 참조는
   `internal/template/settings_test.go:291-349`
   (`TestSettingsTemplateDefaultModelLever`, 13개 참조)로, 이는 **렌더된 템플릿** 비용
   레버를 단언하는 테스트일 뿐 런타임 GLM 경로가 아니다. 즉 GLM 활성화와 비용 레버
   allowlist는 런타임 어느 지점에서도 정합되지 않는다.

6. **영향 범위는 프로젝트별이 아니라 시스템적이다.**
   allowlist는 템플릿 배포 자산이므로, 현재 바이너리로 init/update된 모든 moai
   프로젝트가 영향을 받는다. (이 호스트에 `managed-settings.json`은 부재 —
   "your organization's settings" 라벨은 org 정책 파일이 아니라 프로젝트
   `settings.json`의 `enforceAvailableModels` 키가 생성한 것이다.)

### A.3 SPEC-CC2178-MODEL-POLICY-REPAIR-001과의 관계 (supersession 아님)

선행 SPEC의 설계 근거(model-policy.md:89)는 Claude `[1m]` entitlement-inheritance
버그(#45847/#51060/#36670)만 분석했으며 GLM `ANTHROPIC_DEFAULT_*_MODEL` 리다이렉트
경로는 분석하지 않았다. 본 SPEC은 그 사각지대를 정합한다. 이는 **후속 정합**이지
supersession이 아니다 — Claude-backend Default-model 비용 레버(Default = `sonnet`)는
Claude 세션에 대해 그대로 유효하게 유지된다. 프론트매터의
`depends_on: [SPEC-CC2178-MODEL-POLICY-REPAIR-001]`로 추적한다.

## §B. Known Issues (선행 상태 — #676 회귀 제약 포함)

### B.1 #676 누출 클래스 (Approach A 기각의 핵심 근거 — D1)

solo `moai glm` 경로는 **의도적으로 settings.local.json을 깨끗하게 둔다.**
- `internal/cli/launcher.go:110-151` `applyGLMMode`의 주석(launcher.go:130-135):
  _"settings.local.json injection is intentionally omitted here: setGLMEnv() already
  sets env for the current process which syscall.Exec inherits into `claude`. Writing
  to settings.local.json (as previous behavior) would leak GLM env to subsequent
  `claude` invocations after `moai glm` exits."_
- 회귀 테스트 `internal/cli/launcher_test.go:699`
  `TestApplyGLMMode_NoSettingsLocalPollution`(라인 742/745)이
  `ANTHROPIC_BASE_URL` / `ANTHROPIC_AUTH_TOKEN`가 settings.local.json에 **있으면 안
  됨**을 단언한다.

따라서 settings.local.json 필드는 (syscall.Exec로 상속되는 휘발성 process env와 달리)
**지속**되며, solo→plain-`claude` 종료 경로에는 제거 경로가 없다(`removeGLMEnv`는
`moai cc`에서만 실행). settings.local.json에 `enforceAvailableModels` 키를 기록하는
어떤 솔로-경로 접근도 #676 누출 클래스를 재도입하고 REQ-GMA-002/003을 위반한다.

### B.2 GLM env 변이 사이트 (4개)

| 사이트 | 위치 | 경로 | 본 SPEC 범위 |
|--------|------|------|--------------|
| `setGLMEnv` | glm.go:190-205 | solo, in-process `os.Setenv` (settings.local.json 미기록) | Approach C에서는 무변경 |
| `injectGLMEnvForTeam` | glm.go:595-638 | team/CG, settings.local.json 기록 | Approach C에서는 무변경 |
| `removeGLMEnv` | launcher.go:237 | `moai cc` 클리어 | Approach C에서는 무변경 |
| `injectGLMEnv` | glm.go:923+ | **legacy/미호출 추정** | **out-of-scope** (run-phase M1에서 dead 확인 또는 미사용 명시 — D5) |

### B.3 미러 패리티 테스트 (D3)

- byte-parity 테스트는 `internal/template/rule_template_mirror_test.go:110`
  `TestRuleTemplateMirrorDrift`이며, `workflowOptMirroredPaths` allowlist(라인 42-71)만
  순회한다. 이 allowlist에는 `model-policy.md`가 **미포함**이다 — Approach C가
  model-policy.md를 편집하려면 run-phase에서 이 allowlist에 등록해야 패리티가 강제된다.
  (현재 라이브↔템플릿 두 트리가 byte-identical이므로 등록 없이 테스트만 돌리는 것은
  vacuous — 등록 자체를 AC로 단언해야 함.)

## §C. Requirements (GEARS)

### C.1 핵심 동작 요구사항

**REQ-GMA-001 (Event-driven, GLM 활성 모델 사용 가능).**
**When** 세션이 `moai glm`(전체 세션 GLM) 또는 `moai cg`의 GLM teammate 패널을 통해
활성화되면, the rendered project settings **shall** 설정된 GLM high/opus 모델(예:
`glm-5.2[1m]`, UI 표면상 `opus[1m]`)이 활성 모델로 사용 가능한 상태가 되도록 한다 — 즉
`enforceAvailableModels`에 의해 차단되어 Sonnet으로 폴백되지 않아야 한다.

**REQ-GMA-002 (Ubiquitous, Default-model 비용 레버 보존).**
The rendered project settings **shall** Default-model 비용 레버(`model: "sonnet"`)와
`enforceAvailableModels: true` 강제를 그대로 유지한다 — 비-GLM(`moai cc` / plain
Claude) 세션에서 Default가 여전히 Sonnet으로 해소되고 allowlist 강제가 유효해야 한다.
allowlist 확장은 `[1m]` variant를 **허용**할 뿐 Default를 바꾸거나 강제를 해제하지 않는다.

**REQ-GMA-003 (Ubiquitous, settings.local.json 비오염 / #676 회귀 방지).**
The GLM activation paths **shall** settings.local.json에 모델 allowlist 관련 키를
기록하지 않는다 — solo `moai glm`의 의도적 "settings.local.json clean" 설계(#676)와
`TestApplyGLMMode_NoSettingsLocalPollution` 회귀 테스트가 그대로 통과해야 한다.

**REQ-GMA-004 (Where, 정적 템플릿 적용 / 전 프로젝트 전파).**
**Where** `moai init` 또는 `moai update`가 settings.json을 렌더하면, the rendered
project settings **shall** `availableModels`에 `[1m]` variant(`opus[1m]`, `sonnet[1m]`)를
포함하여, 런타임 로직이나 settings.local.json 쓰기 없이 GLM 리다이렉트가 허용되도록 한다.

**REQ-GMA-005 (Ubiquitous, canonical alias 일관성).**
The expanded `availableModels` list **shall** moai의 canonical Claude 모델 alias 집합
(`internal/web/validate.go:35` `modelCanonical = ["opus","opus[1m]","sonnet","sonnet[1m]",
"haiku","opusplan"]`)과 일관된 alias만 추가한다 — 비표준/임의 모델 id를 도입하지 않는다.

### C.2 검증 가능성 요구사항

**REQ-GMA-006 (Ubiquitous, 차단 해제 실측 검증 포인트 / M1 게이트).**
The plan **shall** `availableModels`에 `opus[1m]`를 추가하면 `enforceAvailableModels:true`
하에서 GLM `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2[1m]` 리다이렉트(표면상 `opus[1m]`)가
실제로 차단 해제되는지를 run-phase 첫 게이트(M1)로 실측한다 — 이 전제가 거짓이면 채택
접근이 무효이므로(그 경우 fallback 접근으로 전환, blocker → orchestrator 재위임).

**REQ-GMA-007 (Ubiquitous, 템플릿 비용 레버 회귀 테스트 갱신).**
The settings template test **shall** 확장된 `availableModels` 내용(+`[1m]` variant)을
기대하도록 갱신되고, `model: "sonnet"`(Default 비용 레버)과 `enforceAvailableModels:true`가
**불변**임을 계속 단언한다 (`TestSettingsTemplateDefaultModelLever`,
`internal/template/settings_test.go`).

### C.3 문서 요구사항

**REQ-GMA-008 (Ubiquitous, 문서 사각지대 정합 + 미러 등록).**
The model-policy 문서 **shall** "GLM-mode reconciliation" 하위 섹션을 통해 충돌과
선택된 수정안(allowlist `[1m]` variant 확장)을 기술하고, run-phase에서
`.claude/rules/moai/development/model-policy.md`를
`internal/template/rule_template_mirror_test.go`의 `workflowOptMirroredPaths`에
**등록**하여 라이브↔템플릿 byte-parity가 `TestRuleTemplateMirrorDrift`로 강제되도록 한다.

## §D. Constraints (HARD)

- [HARD] **settings.local.json에 모델 allowlist 키를 쓰지 말 것** — #676 누출 클래스
  재도입 금지(B.1). solo `moai glm` 경로의 "settings.local.json clean" 설계와
  `TestApplyGLMMode_NoSettingsLocalPollution`를 보존(REQ-GMA-003).
- [HARD] Default-model 비용 레버(`model: "sonnet"`)와 `enforceAvailableModels: true`
  강제는 불변 — Claude 세션에 대해 손상 없이 유지(REQ-GMA-002). allowlist는 확장만 할 뿐.
- [HARD] `availableModels`에 추가하는 항목은 `modelCanonical` 집합과 일관된 `[1m]`
  variant로 한정 — 비표준 모델 id 금지(REQ-GMA-005).
- [HARD] 템플릿 중립성: settings.json.tmpl + model-policy.md 편집은 16-language 중립
  상태를 유지 — 내부 SPEC-ID / 날짜 / commit SHA 등 금지 클래스 누출 불가
  (CLAUDE.local.md §25). 본 SPEC-ID 자체를 템플릿 산출물에 기록하지 말 것.
- [HARD] model-policy.md 편집 시 라이브 사본과 템플릿 사본은 byte-parity를 유지하고
  `workflowOptMirroredPaths`에 등록되어야 한다(REQ-GMA-008 / B.3).

## §E. Self-Verification (plan-phase)

본 SPEC의 plan-phase audit-ready 신호는 `progress.md` §E.1에 기록된다.
run-phase / sync-phase 증거는 각각 §E.2~§E.4(placeholder)에 채워진다.

## §F. Exclusions (Out of Scope)

본 SPEC이 **만들지 않는** 것을 명시한다. 아래 항목은 본 SPEC 범위 밖이다.

### Out of Scope — settings.local.json 런타임 오버라이드 (Approach A 기각)

- GLM 모드에서 settings.local.json에 `enforceAvailableModels:false`를 기록하는 런타임
  접근(iter-1 Approach A). #676 누출 클래스를 재도입하므로 채택하지 않는다(B.1).
- 단, Approach C의 M1 차단-해제 실측이 실패할 경우의 fallback으로만 재검토 가능하며,
  그 경우에도 team/CG 경로 한정 + solo는 누출-free 메커니즘 분리 필요.

### Out of Scope — Claude-backend 비용 레버 재설계

- `settings.json.tmpl`의 `model: "sonnet"` Default-model 비용 레버 자체의 변경/제거.
  본 SPEC은 GLM 모드 충돌만 정합하며, Claude 세션의 Default 라우팅 정책은 그대로 둔다.
- `enforceAvailableModels: true` 강제 자체의 해제(전역 또는 Claude 모드).

### Out of Scope — 새로운 GLM 모델 등록/매핑

- `defaults.go`의 GLM 모델 tier 매핑(`glm-5.2[1m]`/`glm-4.7`/`glm-4.5-air`) 변경.
- 신규 GLM 모델 variant 추가나 `[1m]` 접미사 파싱 로직 변경.
- `modelCanonical`(validate.go:35) 집합 자체의 변경 — 본 SPEC은 그 기존 집합을 *참조*만 한다.

### Out of Scope — 4번째 GLM env 사이트 injectGLMEnv

- `injectGLMEnv`(glm.go:923+)는 legacy/미호출로 추정되며 본 SPEC 범위 밖이다(B.2 / D5).
  run-phase M1에서 dead임을 확인하거나, 미사용임을 명시한다 — 어느 쪽이든 본 SPEC은 이
  사이트를 수정하지 않는다(Approach C는 어떤 env 사이트도 수정하지 않음).

### Out of Scope — 1M context / auto-compact 윈도우 동작

- `glmAutoCompactWindow` / `CLAUDE_CODE_AUTO_COMPACT_WINDOW` 동작 변경
  (SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001 소관). 본 SPEC은 모델 allowlist 정합에 한정.

### Out of Scope — managed-settings.json / org 정책

- org-level `managed-settings.json` 처리. 본 호스트에는 부재하며, 증상은 프로젝트
  `settings.json`의 키가 생성한 것이다. org 정책 우선순위 처리는 범위 밖이다.
