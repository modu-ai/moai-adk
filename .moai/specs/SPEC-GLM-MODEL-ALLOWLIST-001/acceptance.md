# Acceptance Criteria — SPEC-GLM-MODEL-ALLOWLIST-001

모든 AC는 기계적으로 검증 가능해야 한다(명령 출력, 파일 단언, grep 결과, git diff). 각 AC는
spec.md의 REQ-GMA-001~008에 매핑된다. 채택 접근은 **Approach C**(정적 템플릿 `availableModels`
확장)이며, settings.local.json 런타임 쓰기(Approach A)는 기각되었다(spec.md §F).

## §D. AC Matrix

| AC | REQ | 검증 방식 | Severity |
|----|-----|-----------|----------|
| AC-GMA-001 | REQ-GMA-001 | M1 차단-해제 실측 (opus[1m] in availableModels → GLM 리다이렉트 허용, Sonnet 폴백 부재) | MUST |
| AC-GMA-002 | REQ-GMA-002 | `git diff` — availableModels 줄만 변경, model/enforce byte-unchanged | MUST |
| AC-GMA-003 | REQ-GMA-003 | #676 회귀 테스트 PASS + settings.local.json 모델 키 부재 grep | MUST |
| AC-GMA-004 | REQ-GMA-004 | 렌더된 템플릿에 [1m] variant 존재 단언 | MUST |
| AC-GMA-005 | REQ-GMA-005 | 추가 alias가 modelCanonical 부분집합임을 grep으로 확인 | MUST |
| AC-GMA-006 | REQ-GMA-006 | M1 /tmp-project 차단-해제 실측 (run-phase 명명 전제) | MUST |
| AC-GMA-007 | REQ-GMA-007 | `TestSettingsTemplateDefaultModelLever` PASS (갱신된 기대값) | MUST |
| AC-GMA-008 | REQ-GMA-008 | `TestRuleTemplateMirrorDrift` PASS + model-policy.md 미러 등록 | MUST |

## §D.1 Given-When-Then 시나리오

### 시나리오 1 — GLM 모드 진입 시 high 모델 비차단 (REQ-GMA-001)

```
Given  프로젝트 settings.json에 enforceAvailableModels:true가 배포되어 있고
  And  availableModels가 ["sonnet","opus","haiku","opus[1m]","sonnet[1m]"]로 확장되어 있고
  And  GLM high 모델이 glm-5.2[1m](UI 표면상 opus[1m])로 설정되어 있을 때
When   moai glm (또는 moai cg의 GLM teammate 경로)로 GLM 모드를 활성화하면
Then   ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2[1m] 리다이렉트가 enforceAvailableModels에
       의해 차단되지 않고 (opus[1m]가 allowlist에 포함되므로)
  And  glm-5.2[1m]가 활성 모델로 사용 가능하며 Sonnet 폴백 경고가 발생하지 않는다
```

### 시나리오 2 — 추적 settings.json 비용 레버 불변 (REQ-GMA-002)

```
Given  본 SPEC의 모든 변경(Approach C)이 적용된 후
When   git diff로 settings.json.tmpl을 확인하면
Then   "availableModels" 줄만 변경되고 (["sonnet","opus","haiku"] → [1m] variant 포함)
  And  "model": "sonnet"(line 377)과 "enforceAvailableModels": true(line 379)는
       byte-unchanged로 유지되어 Claude-backend 비용 레버가 손상되지 않는다
```

### 시나리오 3 — settings.local.json 비오염 / #676 회귀 방지 (REQ-GMA-003)

```
Given  Approach C는 정적 템플릿만 변경하고 어떤 Go 런타임 경로도 건드리지 않을 때
When   moai glm 진입/종료 사이클을 거치고 settings.local.json을 검사하면
  And  TestApplyGLMMode_NoSettingsLocalPollution을 실행하면
Then   settings.local.json에 모델 allowlist 관련 키(enforceAvailableModels 등)가
       일절 기록되지 않으며 (Approach C는 settings.local.json에 쓰지 않음)
  And  #676 회귀 테스트가 그대로 통과한다 (solo clean 설계 보존)
```

## §D.2 기계적 검증 명령 (run-phase에서 실행)

**AC-GMA-001 / AC-GMA-006 — M1 차단-해제 실측 (run-phase 전제, 최우선 게이트):**
```bash
# Claude Code가 availableModels에 opus[1m]를 포함하면 GLM redirect
# (ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2[1m], 표면상 opus[1m])를
# enforceAvailableModels:true 하에서도 차단하지 않는지 실측.
# 검증: /tmp 테스트 프로젝트에 확장된 availableModels를 가진 settings.json 구성 후
#       moai glm 진입 시 glm-5.2[1m]가 차단되지 않음(Sonnet 폴백 경고 부재)을 확인.
# 차단-해제 실패(여전히 Sonnet 폴백) 시 → Approach A fallback (blocker → orchestrator 재위임)
```

**AC-GMA-002 — 추적 settings.json 비용 레버 불변:**
```bash
git diff internal/template/templates/.claude/settings.json.tmpl
# 기대: "availableModels" 줄만 변경([1m] variant 추가).
#       "model": "sonnet" (line 377) 와 "enforceAvailableModels": true (line 379) 는
#       diff에 나타나지 않음 (byte-unchanged).
grep -n '"model": "sonnet"\|"enforceAvailableModels": true' \
  internal/template/templates/.claude/settings.json.tmpl
# 기대: 두 줄 모두 변경 없이 존재.
```

**AC-GMA-003 — #676 회귀 테스트 + settings.local.json 모델 키 부재:**
```bash
go test -run TestApplyGLMMode_NoSettingsLocalPollution ./internal/cli/ -v
# 기대: PASS (settings.local.json에 GLM 관련 키 부재 단언 유지).
grep -rn 'enforceAvailableModels\|availableModels' internal/cli/ | grep -v '_test.go'
# 기대: NO output — 모델 allowlist 키를 settings.local.json(또는 어떤 Go 경로)에도 쓰지 않음.
```

**AC-GMA-004 / AC-GMA-005 — 정적 템플릿 전파 + canonical alias:**
```bash
# 렌더된 템플릿(또는 .tmpl 직접)에 [1m] variant 존재 확인:
grep -n 'availableModels' internal/template/templates/.claude/settings.json.tmpl
# 기대: ["sonnet","opus","haiku","opus[1m]","sonnet[1m]"] (또는 동등 [1m] 포함 형태).
# 추가 alias가 modelCanonical 부분집합임을 확인:
grep -n 'modelCanonical' internal/web/validate.go
# 기대: opus[1m], sonnet[1m] 모두 modelCanonical = ["opus","opus[1m]","sonnet",
#       "sonnet[1m]","haiku","opusplan"]에 포함 (비표준 모델 id 미도입).
```

**AC-GMA-007 — 템플릿 비용 레버 회귀 테스트 (갱신된 기대값):**
```bash
go test -run TestSettingsTemplateDefaultModelLever ./internal/template/ -v
# 기대: PASS — 확장된 availableModels 기대값으로 갱신되었고,
#       model:"sonnet" + enforceAvailableModels:true 단언은 불변 유지.
```

**AC-GMA-008 — 문서 + 미러 byte-parity + 미러 등록:**
```bash
go test -run TestRuleTemplateMirrorDrift ./internal/template/ -count=1
# 기대: PASS — model-policy.md가 workflowOptMirroredPaths에 등록되어 byte-parity 강제.
grep -n 'model-policy.md' internal/template/rule_template_mirror_test.go
# 기대: workflowOptMirroredPaths(line 42-71)에 ".claude/rules/moai/development/model-policy.md" 등록.
grep -n 'GLM-mode reconciliation' \
  internal/template/templates/.claude/rules/moai/development/model-policy.md \
  .claude/rules/moai/development/model-policy.md
# 기대: 라이브 + 템플릿 두 사본 모두 "GLM-mode reconciliation" 섹션 보유.
```

## §D.3 Edge Cases

- **GLM high 모델이 `[1m]` 없는 매핑인 경우**: 비기본 매핑(예: `glm-4.7`)이면 `availableModels`에
  해당 alias도 포함되어야 차단되지 않음. 본 SPEC은 기본 매핑(`glm-5.2[1m]` → 표면상 `opus[1m]`)에
  대한 차단 해제를 보장한다. 비기본 매핑은 spec.md §F("새로운 GLM 모델 등록/매핑")로 out-of-scope.
- **사용자가 settings.json을 수동 편집한 프로젝트**: `moai update`가 비실행 상태면 사용자
  settings.json은 구 allowlist 유지. `moai update` 재렌더 시 확장 allowlist 전파(spec.md §A.2.6).
- **렌더링 변수 보존**: `availableModels` 줄은 Go 템플릿 변수를 포함하지 않으므로 `.tmpl`과
  렌더 산출물이 동일 — `make build` 후 embedded.go에도 반영됨.
- **modelCanonical 외 alias 시도**: `availableModels`에 `modelCanonical` 부분집합이 아닌
  값(예: `glm-5.2[1m]` 원본 id)을 추가하는 것은 REQ-GMA-005 위반 — UI 표면 alias(`opus[1m]`/
  `sonnet[1m]`)만 추가한다.

## §D.4 Quality Gate

- `go test ./internal/cli/... ./internal/template/...` 전부 PASS
- `go vet ./...` clean
- `golangci-lint run` 신규 위반 0 (Approach C는 Go 런타임 코드 무변경 → 신규 위반 기대 0)
- `make build` 성공 (embedded.go 재생성, settings.json.tmpl + model-policy.md 미러 반영)
- 신규/변경 패키지 회귀 0 (런타임 코드 무변경이므로 커버리지 델타는 테스트 기대값 갱신에 한정)

## §D.5 Definition of Done

- [ ] M1 차단-해제 실측 통과 (또는 Approach A fallback 전환 결정 + 재위임)
- [ ] REQ-GMA-001~008 전부 AC PASS (기계적 증거 첨부)
- [ ] 추적 settings.json.tmpl `model:"sonnet"` + `enforceAvailableModels:true` byte-unchanged
- [ ] `availableModels`가 [1m] variant 포함으로 확장 (modelCanonical 부분집합)
- [ ] settings.local.json에 모델 allowlist 키 부재 (#676 회귀 테스트 PASS)
- [ ] model-policy.md 라이브 + 템플릿 미러 byte-parity + workflowOptMirroredPaths 등록 + 템플릿 중립성
- [ ] 전체 테스트 스위트 + lint + vet PASS + `make build` 성공
- [ ] status: draft → (run-phase에서 manager-develop이 in-progress 전이)

## §E. Exclusions (Out of Scope)

본 acceptance가 **검증하지 않는** 범위. spec.md §F와 정합한다.

### Out of Scope — settings.local.json 런타임 오버라이드 (Approach A 기각)

- GLM 모드에서 settings.local.json에 `enforceAvailableModels:false`를 기록하는 런타임 접근을
  검증하는 AC. Approach A는 기각되었으므로(#676 누출 클래스 재도입, spec.md §B.1) 해당 상태를
  단언하는 AC(구 AC-GMA-001b/AC-GMA-004 settings.local.json 상태)는 제거되었다.

### Out of Scope — Claude-backend 비용 레버 재설계

- `model: "sonnet"` Default-model 비용 레버 자체의 변경/제거를 검증하는 AC.
- `enforceAvailableModels: true` 강제 해제(전역 또는 Claude 모드)를 검증하는 AC.

### Out of Scope — 4번째 GLM env 사이트 injectGLMEnv

- `injectGLMEnv`(glm.go:923+)의 수정/dead 확인을 검증하는 AC. Approach C는 어떤 env 사이트도
  수정하지 않으므로(spec.md §F / B.2) 본 acceptance 범위 밖이다.
