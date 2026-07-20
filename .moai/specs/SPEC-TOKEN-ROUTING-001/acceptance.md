# Acceptance Criteria — SPEC-TOKEN-ROUTING-001

> 본 문서는 Token-Economy Epic 2/4(B)의 관측 가능한 완료 기준을 정의한다.
> 모든 AC는 기계적 검증(명령 출력 / 파일 존재 / grep hit / 테스트 PASS)을
> 만족해야 하며, "should" / "might" / "usually" 같은 모호어를 사용하지 않는다.

## §D AC Matrix

> REQ-TR-001 부터 REQ-TR-012 까지 12개 요구사항에 대해 12개 AC가 1:1로 대응.
> 각 AC는 Tier/Phase/우선순위/검증 기계를 명시.

| AC-ID | 대응 REQ | 시나리오 | Severity | 검증 기계 |
|-------|----------|----------|----------|-----------|
| AC-TR-001 | REQ-TR-001 | `model_routing` 블록이 12 엔트리(3 Tier × 4 Phase)를 가짐 | MUST | `grep -c` 카운트 + YAML parse |
| AC-TR-002 | REQ-TR-002 | 누락 (Tier,Phase) 쌍 조회 시 fallback 적용 + `fallback_applied: true` | MUST | loader unit test |
| AC-TR-003 | REQ-TR-003 | `RouteModelFor(tier, phase)` typed entry 반환 | MUST | loader unit test |
| AC-TR-004 | REQ-TR-004 | closed-set 위반(model/effort) 시 validation error | MUST | loader unit test |
| AC-TR-005 | REQ-TR-005 | orchestrator spawn call-site이 `RouteModelFor` 호출해 override 주입 | MUST | call-site grep + 통합 테스트 |
| AC-TR-006 | REQ-TR-006 | explicit user override 시 matrix 양보(미교체) | MUST | 통합 테스트 |
| AC-TR-007 | REQ-TR-007 | `model_routing` closed-set이 `workflow_agents`와 동일 | MUST | shared validation 테스트 |
| AC-TR-008 | REQ-TR-008 | template mirror가 동일 `model_routing` 블록 가짐(neutrality PASS) | MUST | template mirror grep + CI |
| AC-TR-009 | REQ-TR-009 | matrix 매핑 시 AskUserQuestion 없이 자동 적용 | MUST | 통합 테스트 + grep (no AskUserQuestion on path) |
| AC-TR-010 | REQ-TR-010 | `moai cg` default launcher 전환 부재 | MUST | grep + trigger 추적 |
| AC-TR-011 | REQ-TR-011 | Phase 0.95 mode-selection 간섭 부재(직교성 보존) | MUST | orchestration-mode-selection.md diff 비어있음 |
| AC-TR-012 | REQ-TR-012 | unavailable model 시 advisory + session-inherited fallback | SHOULD | loader unit test |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — 선언적 매트릭스 로딩 (happy path) [AC-TR-001, AC-TR-003]

**Given** `workflow.yaml` 이 최상위 `model_routing` 블록을 가지며, 3 Tier × 4 Phase
= 12 엔트리 각각이 closed-set `{model, effort}` 값을 가진다.

**When** `internal/config.RouteModelFor("S", "sync")` 를 호출한다.

**Then**:
- 반환된 `ModelRoutingEntry` 의 `model` 필드가 `workflow.yaml` 에 선언된
  Tier=S / Phase=sync 엔트리의 값과 일치한다.
- `fallback_applied` 가 `false` 이다.
- `grep -c 'model_routing' workflow.yaml` 이 1 이상이다.
- `grep -c 'tier:' workflow.yaml`(template mirror 포함)이 template과 local이
  동일한 카운트를 보고한다.

**검증 명령**:
```bash
go test -run TestRouteModelForHappyPath ./internal/config/
grep -c 'model_routing:' .moai/config/sections/workflow.yaml
grep -c 'model_routing:' internal/template/templates/.moai/config/sections/workflow.yaml
```

### Scenario 2 — 폐쇄집합 검증 + fallback [AC-TR-002, AC-TR-004, AC-TR-007]

**Given** 다음 3가지 하위 케이스:
- (a) `model_routing` 에 허용되지 않은 model 값(예: `gpt-4`)이 포함된 workflow.yaml
- (b) `model_routing` 에 허용되지 않은 effort 값(예: `ultra`)이 포함된 workflow.yaml
- (c) (Tier=L, Phase=mx) 가 매트릭스에서 누락됨

**When**:
- (a)/(b): loader가 해당 YAML을 로드한다.
- (c): `RouteModelFor("L", "mx")` 를 호출한다.

**Then**:
- (a) `model: gpt-4` 로드 시 validation error 반환, entry yield 거부.
- (b) `effort: ultra` 로드 시 validation error 반환, entry yield 거부.
- (c) fallback entry 반환 + 반환된 entry의 `fallback_applied: true`.
- closed-set 검증이 `workflow_agents` 의것과 **동일한 집합**을 사용함을
  shared validation 테스트가 확인(REQ-TR-007).

**검증 명령**:
```bash
go test -run 'TestRouteModelForClosedSet|TestRouteModelForFallback' ./internal/config/
go test -run TestSharedClosedSetWithWorkflowAgents ./internal/config/
```

### Scenario 3 — 기본 동작 계약 (default-behavior) [AC-TR-005, AC-TR-006, AC-TR-009, AC-TR-010]

**Given** orchestrator가 Tier S sync-phase 작업을 `Agent()`로 spawn하려 한다.
매트릭스는 (Tier=S, Phase=sync) → `{model: haiku, effort: low}` 를 선언한다.

**When**:
- (default path) 사용자의 explicit override가 없는 상태에서 spawn.
- (override path) 사용자가 `model: sonnet` 을 미리 명시한 상태에서 spawn.

**Then**:
- (default) spawn 파라미터가 `model: haiku`, `effort: low` 를 carry한다.
  **`AskUserQuestion` 호출이 없다**(REQ-TR-009) — 자동 적용.
- (override) spawn 파라미터가 `model: sonnet` 을 유지한다(사용자 override).
  matrix의 `haiku` 가 sonnet을 밀어내지 않는다(REQ-TR-006).
- 어떤 코드 경로도 `moai cg` 를 default launcher로 전환하는 트리거를 포함하지
  않는다(REQ-TR-010).

**검증 명령**:
```bash
# call-site이 RouteModelFor를 호출함을 grep (run-phase 구현 후)
grep -rn 'RouteModelFor' internal/ | grep -v '_test.go'
# AskUserQuestion이 spawn default path에 없음을 grep
grep -rn 'AskUserQuestion' internal/harness/ internal/orchestrator/ 2>/dev/null \
  | grep -v '_test.go' | grep -i 'model_routing\|RouteModelFor'
# moai cg default launcher 부재
grep -rn 'moai cg' .moai/config/ internal/template/templates/ 2>/dev/null \
  | grep -iE 'default|auto.*launch' || echo "no default-launcher reference"
```

### Scenario 4 — template mirror + neutrality [AC-TR-008]

**Given** run-phase가 `model_routing` 블록을 local `workflow.yaml` 에 추가했다.

**When** template mirror(`internal/template/templates/.moai/config/sections/workflow.yaml`)
을 확인한다.

**Then**:
- template mirror가 동일한 `model_routing` 블록을 가진다.
- `internal_content_leak_test.go` 가 PASS — SPEC-ID/내부날짜/commit SHA가
  template tree에 누출되지 않는다.
- `.github/workflows/template-neutrality-check.yaml` 이 PASS.
- 주석이 generic("Tier×Phase → {model, effort} routing defaults")하다.

**검증 명령**:
```bash
# template mirror에 동일 블록 존재
diff <(grep -A20 'model_routing:' .moai/config/sections/workflow.yaml) \
     <(grep -A20 'model_routing:' internal/template/templates/.moai/config/sections/workflow.yaml) \
  | grep -v '^---' | grep -v '^$' | head -5
# neutrality CI guard
go test ./internal/template/ -run TestInternalContentLeak
# SPEC-ID 누출 grep (self-check)
grep -rE 'SPEC-[A-Z0-9-]+' internal/template/templates/.moai/config/sections/workflow.yaml \
  | grep -v '^[^:]*:[0-9]*:[ \t]*#' || echo "no SPEC-ID leak"
```

### Scenario 5 — 직교성 보존 + deployment neutrality [AC-TR-011, AC-TR-012]

**Given**:
- 본 SPEC run-phase가 완료된 상태.
- GLM 환경 변수(ANTHROPIC_AUTH_TOKEN/ANTHROPIC_BASE_URL 중 하나)가 부재한 세션.

**When**:
- (직교성) `orchestration-mode-selection.md` 의 Mode 1-6 카탈로그/결정트리/
  auto-select thresholds(domains ≥ 3 / files ≥ 10 / score ≥ 7)의 diff를 확인.
- (neutrality) 매트릭스가 `glm` model을 반환하는 (Tier,Phase) 쌍을 조회하되,
  GLM env가 부재한 상황.

**Then**:
- (직교성) 본 SPEC 진행 전후로 `orchestration-mode-selection.md` 의 diff가 비어있다
  (mode-selection 축에 간섭하지 않음, REQ-TR-011).
- (neutrality) loader가 advisory를 반환한다("glm referenced but GLM env absent;
  falling back to session-inherited model"). orchestrator는 session-inherited
  model로 폴백하되, advisory가 `.moai/logs/` 에 기록된다(REQ-TR-012, SHOULD).

**검증 명령**:
```bash
# 직교성 — orchestration-mode-selection.md 무결정성 (본 SPEC 커밋 전후 비교)
git log --oneline -- .claude/rules/moai/workflow/orchestration-mode-selection.md \
  | head -5
git diff <commit-before-B>..<head> -- .claude/rules/moai/workflow/orchestration-mode-selection.md \
  | grep -E '^[+-]' | grep -v '^[+-][+-][+-]' || echo "no mode-selection interference"
# deployment neutrality
go test -run TestRouteModelForUnavailableModel ./internal/config/
ls -la .moai/logs/ 2>/dev/null | grep -i model_routing || echo "advisory logged at sync-phase"
```

## §D.2 Edge Cases

| Edge Case | 기대 동작 | 관련 AC |
|-----------|-----------|---------|
| 빈 `model_routing` 블록(0 엔트리) | fallback entry만 반환 + 모든 조회가 `fallback_applied: true` | AC-TR-002 |
| (Tier,Phase) 가 closed-set 밖(예: Tier=X) | 호출자의 입력 검증 에러(tier ∈ {S,M,L} 위반) | AC-TR-004 |
| 중복 키(YAML이 같은 (Tier,Phase)를 두 번 선언) | YAML loader가 에러 반환(duplicate key) | AC-TR-004 |
| `model: inherit` + GLM 환경 | `inherit` 은 session default를 의미하므로 advisory 없이 그대로 반환 | AC-TR-012 |
| archived agent 이름이 매트릭스 주석에 등장 | CI guard FAIL(또는 self-check grep이 사전 차단) | AC-TR-008(neutrality의 확장) |
| 병렬 세션이 `workflow.yaml` 을 동시 편집 | pathspec-only commit으로 충돌 회피 | (운영 제약, plan.md §B-scope-discipline) |

## §D.3 Quality Gate

- [ ] `go test ./internal/config/...` PASS(0 failures)
- [ ] `go test ./internal/template/...` PASS(neutrality leak 0)
- [ ] `golangci-lint run` 0 issues
- [ ] `go vet ./...` 0 issues
- [ ] `gofmt -l .` empty
- [ ] coverage: `internal/config` 패키지 본 SPEC 추가 코드 ≥ 85%(quality.yaml 목표)
- [ ] build: host + `GOOS=windows GOARCH=amd64` exit 0
- [ ] `.github/workflows/template-neutrality-check.yaml` PASS
- [ ] `internal/template/internal_content_leak_test.go` PASS

## §D.4 Definition of Done

**DoD 충족 조건(전부 MUST)**:

1. **기능 완결**:
   - `workflow.yaml`(local + template mirror)에 `model_routing` 블록이 12 엔트리를 가진다(AC-TR-001, AC-TR-008).
   - `RouteModelFor(tier, phase)` typed loader가 closed-set 검증과 fallback을 제공한다(AC-TR-003, AC-TR-004, AC-TR-007).
   - orchestrator call-site이 loader를 호출해 per-spawn override를 주입한다(AC-TR-005).

2. **계약 준수**:
   - default-behavior(REQ-TR-009/010)가 정확히 (a) auto-route에 한정됨을 테스트가 검증한다(AC-TR-009, AC-TR-010).
   - 사용자 explicit override가 matrix를 이긴다(AC-TR-006).
   - 직교성(REQ-TR-011)이 보존된다(AC-TR-011).

3. **품질 게이트**:
   - §D.3 모든 체크박스 PASS.
   - 8-agent 현행 카탈로그 외의 archived 이름이 매트릭스/주석에 0 노출.

4. **문서화**:
   - `workflow.yaml` 주석이 generic하고 SPEC-ID가 없다(neutrality).
   - template mirror가 local과 동일 블록을 가진다.

5. **Epic 위치**:
   - Epic A(`SPEC-TOKEN-ACCOUNTING-001`)의 completed 상태를 훼손하지 않는다.
   - Epic C/D가 본 SPEC의 매트릭스를 소비할 수 있는 공개 API(`RouteModelFor`)를 남긴다.

6. ** close**:
   - sync-phase에서 `status: draft → in-progress → implemented → completed` 전이(manager-docs 소관).
   - `sync_commit_sha` 가 리터럴 YAML 필드로 progress.md에 기록된다(era V3R6 계약).
   - §I Token Accounting에 A의 writer contract로 측정된 tokens_spent 값이 들어간다.
