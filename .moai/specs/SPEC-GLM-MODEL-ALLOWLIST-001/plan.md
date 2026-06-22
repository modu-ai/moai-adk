# Implementation Plan — SPEC-GLM-MODEL-ALLOWLIST-001

## §A. Context

`moai glm` 세션이 `enforceAvailableModels:true`(Claude-backend 비용 레버, 추적
`settings.json`) 때문에 GLM high 모델(`glm-5.2[1m]`)을 차단하고 Sonnet으로 폴백하는
회귀를 정합한다. 근본 원인 인과 사슬은 spec.md §A.2에 검증된 형태로 기록되어 있다.

### A.5 PRESERVE 대상 (Approach C는 Go 런타임 코드를 건드리지 않는다)

Approach C는 **정적 템플릿 + 템플릿 테스트 + 문서**만 변경하고, 모든 Go 런타임 경로는
무변경으로 유지한다. 다음 항목은 PRESERVE 대상이다:

- `settings.json.tmpl`의 `"model": "sonnet"`(line 377) — 불변. Default-model 비용 레버.
- `settings.json.tmpl`의 `"enforceAvailableModels": true`(line 379) — 불변. allowlist 강제.
- `internal/cli/glm.go` `setGLMEnv`(190-205) — 무변경. solo in-process `os.Setenv`.
- `internal/cli/glm.go` `injectGLMEnvForTeam`(595-638) — 무변경. team settings.local.json env.
- `internal/cli/launcher.go` `removeGLMEnv`(237) / `applyGLMMode`(110-151) — 무변경.
- `internal/cli/glm.go` `injectGLMEnv`(923+) — 무변경(B.2 / spec.md §F out-of-scope).
- 모든 `settings.local.json` 쓰기 경로 — 무변경(Approach C는 settings.local.json에
  모델 allowlist 키를 일절 쓰지 않는다 — #676 누출 클래스 회피, D1).
- `internal/web/validate.go:35` `modelCanonical` 집합 — *참조*만, 무변경.
- `internal/config/defaults.go`의 GLM 모델 tier 매핑 — 무변경.

## §B. Approach 평가 (채택: Approach C)

### 채택 — Approach C (정적 템플릿 allowlist 확장)

`internal/template/templates/.claude/settings.json.tmpl`의 `availableModels`를
`["sonnet", "opus", "haiku"]`에서 `[1m]` variant를 포함하도록 확장한다 (예:
`["sonnet", "opus", "haiku", "opus[1m]", "sonnet[1m]"]`).

근거(plan-auditor iter-1 D1 반영):
- **런타임 코드 ZERO 변경.** `availableModels`는 추적 `settings.json`(템플릿 렌더 산출물)에
  위치하며, `[1m]` variant를 미리 허용하면 GLM `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2[1m]`
  리다이렉트(UI 표면상 `opus[1m]`)가 `enforceAvailableModels:true` 하에서도 차단되지 않는다.
- **#676 누출 클래스 회피.** settings.local.json에 어떤 키도 쓰지 않으므로, solo
  `moai glm` 경로의 의도적 "settings.local.json clean" 설계(launcher.go:130-135 주석)와
  회귀 테스트 `TestApplyGLMMode_NoSettingsLocalPollution`(launcher_test.go:699)를 보존한다.
  Approach C에는 inject↔clear 패리티 부담 자체가 존재하지 않는다(쓰는 키가 없으므로).
- **Default 비용 레버 보존.** `model: "sonnet"` + `enforceAvailableModels: true`는
  byte-unchanged로 유지된다. allowlist는 `[1m]` variant를 **허용**할 뿐, Default를 바꾸거나
  강제를 해제하지 않는다 → Claude 세션은 여전히 Sonnet으로 해소되고 강제가 유효(REQ-GMA-002).
- **canonical alias 일관성.** 추가하는 `opus[1m]` / `sonnet[1m]`는 모두
  `internal/web/validate.go:35` `modelCanonical = ["opus","opus[1m]","sonnet","sonnet[1m]",
  "haiku","opusplan"]`의 부분집합 — 비표준/임의 모델 id를 도입하지 않는다(REQ-GMA-005).
- **전 프로젝트 전파.** 템플릿 배포 자산이므로 `moai update`가 사용자 프로젝트를
  새 allowlist로 재렌더한다 → 시스템적 정합(spec.md §A.2.6).

### 기각 — Approach A (settings.local.json 런타임 오버라이드)

iter-1에서 권장했던 Approach A(GLM 모드에서 settings.local.json에
`enforceAvailableModels:false`를 기록)는 **기각**한다. 기각 근거(D1):
- solo `moai glm` 경로는 의도적으로 settings.local.json을 깨끗하게 둔다
  (`launcher.go:130-135` 주석 — settings.local.json injection은 의도적으로 생략;
  `setGLMEnv`가 설정한 process env를 `syscall.Exec`가 `claude`로 상속). settings.local.json에
  키를 기록하면 `moai glm` 종료 후 후속 `claude` 호출에 GLM 상태가 누출된다(#676 회귀).
- 회귀 테스트 `internal/cli/launcher_test.go:699` `TestApplyGLMMode_NoSettingsLocalPollution`
  (라인 742/745)이 settings.local.json에 GLM 관련 키가 **없어야** 함을 단언한다.
- solo→plain-`claude` 종료 경로에는 제거 경로가 없다(`removeGLMEnv`는 `moai cc`에서만 실행)
  → settings.local.json 필드는 지속되어 누출 클래스를 재도입한다(REQ-GMA-002/003 위반).
- **fallback로만 재검토 가능**: Approach C의 M1 차단-해제 실측이 실패할 경우에 한해,
  team/CG 경로 한정 + solo는 누출-free 메커니즘 분리를 전제로 재검토한다(spec.md §F).

## §C. Tier 판정

**Tier M (standard)** — 확정.

근거:
- 다중 파일 변경: `settings.json.tmpl`(allowlist 1줄) + `settings_test.go`(기대값 갱신)
  + `model-policy.md` 문서(라이브 + 템플릿 미러 2곳) + `rule_template_mirror_test.go`
  (`workflowOptMirroredPaths` 등록). 5 - 15 files 구간.
- 단일 도메인(GLM/모델-policy settings 정합), 신규 아키텍처 없음, 신규 Go 런타임 코드 ZERO.
- Milestone 4개로 충분(M1~M4).
- Tier L의 cascade/대규모 codegen 특성에 미달; Tier S(2-artifact)보다는 문서 미러 + 테스트
  갱신 + 실측 게이트가 얽혀 있어 3-artifact가 적절.

## §D. Pre-flight (검증된 사실 — plan-phase Read/grep)

| 사실 | 출처 (verified this session) |
|------|------------------------------|
| 추적 비용 레버 위치 | `settings.json.tmpl:377-379` (`model:"sonnet"` / `availableModels:["sonnet","opus","haiku"]` / `enforceAvailableModels:true`) |
| CC 2.1.176 리다이렉트 차단 의미론 | `model-policy.md:87` (verbatim) |
| GLM 기본 high 모델 | `defaults.go:46,57` `glm-5.2[1m]` |
| canonical alias 집합 | `internal/web/validate.go:35` `modelCanonical` (`opus[1m]`/`sonnet[1m]` 포함 확인) |
| #676 solo clean 설계 | `launcher.go:130-135` 주석 + `launcher_test.go:699` `TestApplyGLMMode_NoSettingsLocalPollution` |
| 템플릿 비용 레버 테스트 | `internal/template/settings_test.go` `TestSettingsTemplateDefaultModelLever` (13 참조) |
| 미러 패리티 테스트 | `internal/template/rule_template_mirror_test.go:110` `TestRuleTemplateMirrorDrift` (`workflowOptMirroredPaths` 순회, line 42-71) |
| `model-policy.md` 미러 등록 상태 | `workflowOptMirroredPaths`에 **미포함** → run-phase 등록 필요 (B.3) |
| SPEC-ID 충돌 | 없음 (`SPEC-GLM-MODEL-ALLOWLIST-001` 디렉토리만 존재) |

## §E. Self-Verification

plan-phase audit-ready 신호 → `progress.md` §E.1.

## §F. Milestones (우선순위 기반, 시간 추정 없음)

- **M1 — 차단-해제 실측 게이트(REQ-GMA-006).** Approach C의 전제 검증: `availableModels`에
  `opus[1m]`를 추가하면 `enforceAvailableModels:true` 하에서 GLM
  `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2[1m]` 리다이렉트(표면상 `opus[1m]`)가 실제로 차단
  해제되는지를 /tmp 테스트 프로젝트에서 commit 전 실측한다. **전제가 거짓이면 Approach C는
  무효** → blocker → orchestrator 재위임(Approach A fallback 재검토). 이 게이트는 이후 모든
  M의 선결 조건이다.
- **M2 — settings.json.tmpl allowlist 편집 + 템플릿 테스트 갱신(REQ-GMA-001/004/005/007).**
  `settings.json.tmpl:378` `availableModels`를 `[1m]` variant 포함으로 확장(예:
  `["sonnet","opus","haiku","opus[1m]","sonnet[1m]"]`). `model:"sonnet"`(line 377) +
  `enforceAvailableModels:true`(line 379)는 byte-unchanged. `make build`로 embedded.go
  재생성. `internal/template/settings_test.go` `TestSettingsTemplateDefaultModelLever`를
  확장된 `availableModels` 기대값으로 갱신하되, `model:"sonnet"` + `enforceAvailableModels:true`
  단언은 **불변 유지**.
- **M3 — model-policy.md 문서 정합 + 미러 등록(REQ-GMA-008).** "GLM-mode reconciliation"
  하위 섹션을 라이브(`.claude/rules/moai/development/model-policy.md`) + 템플릿
  (`internal/template/templates/.claude/rules/moai/development/model-policy.md`) 양쪽에 추가.
  `internal/template/rule_template_mirror_test.go`의 `workflowOptMirroredPaths`(line 42-71)에
  `.claude/rules/moai/development/model-policy.md`를 **등록**하여 byte-parity가
  `TestRuleTemplateMirrorDrift`로 강제되도록 한다(현재 미등록 → 등록 없이는 패리티 vacuous).
  템플릿 중립성 유지(SPEC-ID/날짜/commit SHA 누출 금지, CLAUDE.local.md §25).
- **M4 — sync/검증 마일스톤(REQ-GMA-002/003).** 추적 `settings.json.tmpl` 비용 레버 불변을
  `git diff`로 검증(`availableModels` 줄만 변경, `model`/`enforceAvailableModels` byte-unchanged).
  #676 회귀 테스트 `TestApplyGLMMode_NoSettingsLocalPollution` 통과 재확인(모델 allowlist
  키가 settings.local.json 경로에 일절 부재함을 grep). 전체 테스트 스위트 + lint + vet PASS.

## §G. Anti-Patterns (회피)

- settings.local.json에 `enforceAvailableModels` 등 모델 allowlist 키 기록(Approach A 잔재 —
  #676 누출 클래스 재도입, D1 위반).
- `model:"sonnet"` 또는 `enforceAvailableModels:true` 줄 변경/제거(Default 비용 레버 손상,
  REQ-GMA-002 위반).
- `availableModels`에 비표준 모델 id 추가(`modelCanonical` 집합 외 — REQ-GMA-005 위반).
- `make build`(embedded.go 재생성) 누락 → 템플릿 편집이 배포에 반영 안 됨.
- model-policy.md 템플릿 미러에 내부 SPEC-ID/날짜/commit SHA 기록(템플릿 중립성 §25 위반).
- model-policy.md를 편집하면서 `workflowOptMirroredPaths` 등록을 누락 → 미러 drift 미검출(B.3).
- M1 차단-해제 실측 없이 바로 commit → Approach C가 무효일 경우 전체 재작업.

## §H. Cross-References

- `spec.md` §A.2 (검증된 인과 사슬), §A.3 (선행 SPEC 관계), §B (Known Issues), §F (Exclusions).
- SPEC-CC2178-MODEL-POLICY-REPAIR-001 (비용 레버 도입 — depends_on/후속 관계).
- `.claude/rules/moai/core/glm-web-tooling.md` § GLM-Backend Detection (GLM 모드 판정 기준).
- `internal/web/validate.go:35` `modelCanonical` (canonical alias 집합 — 참조 SSOT).
- CLAUDE.local.md §2 (settings.local.json 분리), §22 (dev settings intent), §25 (템플릿 중립성),
  §2.1 (템플릿 콘텐츠 중립성).
- `internal/template/CLAUDE.md` § Mirror parity checks + Template-First Rule(`make build`).
