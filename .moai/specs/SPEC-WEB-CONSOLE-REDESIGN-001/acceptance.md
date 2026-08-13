# SPEC-WEB-CONSOLE-REDESIGN-001 — 인수 기준

> 대응 SPEC: `spec.md` §B (REQ-WCR-001 ~ REQ-WCR-063)
> 구현 계획: `plan.md` §D (M1 ~ M7)

---

## §A 판정 규약

### §A.1 공허한 GREEN 방지 (AC 명령 작성 규칙)

아래 규칙을 어긴 판정 명령은 실행이 통과해도 **증거로 인정하지 않는다**.

1. **표 셀 안의 `\|`를 `grep -E` 패턴에 넣지 말 것** — ERE에서 `\|`는 리터럴 파이프이며 교대(alternation)가 아니다. 교대가 필요하면 `grep -E 'a|b'`를 코드 블록 안에 쓴다.
2. **패턴 언어와 파일 언어를 일치시킬 것** — `.claude/` 하위(skills·agents·rules)는 영어 전용이다. 한국어 토큰으로 그 트리를 grep하면 항상 0건이 나온다. 반대로 `.moai/specs/` 본문은 한국어다.
3. **부정형(부재) AC는 RED fixture로 증명할 것** — "X가 없어야 한다"는 명령이 X가 있는 상태에서 실제로 실패하는지 먼저 확인한다. 확인 없이 0건을 PASS로 읽으면 오탈자 하나로 영구 GREEN이 된다.
4. **가드를 재구현하지 말 것** — 기존 가드 테스트(i18n governance, 템플릿 중립성)가 있으면 그 테스트 실행이 권위다. 정규식을 면제 목록 없이 재구현하면 거짓 실패 전용 기계가 된다.
5. **커밋 후 토큰으로 판정하지 말 것** — 변경 후 상태의 토큰을 세는 AC는 구현이 그 토큰을 만들면 자동 통과한다. 베이스라인 델타(변경 전 N → 변경 후 M) 또는 독립 불변식으로 판정한다.

### §A.2 실행 기준

- 모든 명령은 워크트리 루트(`.claude/worktrees/web-redesign`)에서 실행한다.
- 라인 앵커는 사용하지 않는다. content-token으로 판정한다.
- Go 테스트 이름은 run-phase에서 실제 작성되는 이름으로 고정하며, 여기 기재된 이름은 계약이다(구현이 이름을 바꾸면 AC도 함께 갱신).

---

## §B 실측 증거 (베이스라인)

plan-phase 시점에 오케스트레이터가 수행한 실측. run-phase 착수 전 `plan.md` §C Pre-flight로 재검증한다.

| ID | 주장 | 명령 | 관측 |
|----|------|------|------|
| E-1 | 섹션 등록 11개 vs 렌더 3개 | `internal/settings/schema_sections.go` `SchemaSectionIDs()` 판독 + `internal/web/schemaform.go` `schemaSectionMetas()` 판독 | 등록: quality_extras, git_strategy, llm, workflow, harness, ralph, feedback, observability, security, handoff, cache (11). 렌더: LLM, Workflow, Report (3) |
| E-2 | `token_budget` 접근자 호출자 0건 | `grep -rn "WorkflowPlanTokens\|WorkflowRunTokens\|WorkflowSyncTokens" internal/ pkg/ cmd/ --include='*.go' \| grep -v _test.go` | 정의부(`internal/config/workflow_accessors.go`)만; 호출자 0 |
| E-3 | `auto_clear` 접근자 호출자 0건 | `grep -rn "WorkflowAutoClearEnabled" internal/ pkg/ cmd/ --include='*.go' \| grep -v _test.go` | 정의부만; 호출자 0. 산문 소비자 0 |
| E-4 | `agentic_loop` / `loop_prevention` 산문 소비 실존 | `.claude/skills/moai/**` + docs-site 우선순위 표 판독 | `moai.md`(agentic_loop), `loop.md` + docs-site(loop_prevention) |
| E-5 | worktree / branch_guard Go 소비자 실존 | `internal/cli/worktree_advisory.go`, `internal/cli/session_worktree.go`, `internal/hook/pre_tool.go` 판독 | 3개 실소비 지점 |
| E-6 | GLM 모델 티어 주입 | `internal/cli/glm.go` `setGLMEnv` 판독 | `llm.glm.models.{high,medium,low,fable}` → `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` |
| E-7 | 추론 강도 채널 단일 | `internal/template/glm_effort_overlay.go` + `internal/cli/glm.go` + `internal/cli/launcher.go` 판독 | `SessionGLMReasoningState()`는 조건 없이 `reasoning-max` 반환; `SessionGLMReasoningStateForEffort(effort)`가 세션 effort 환경설정을 collapse. 티어별 채널 부재. 코드 주석이 z.ai 준수 여부를 UNVERIFIED로 표기 |
| E-8 | GLM 키 마스킹 계약 | `internal/web/glmkey.go` + `fieldsetGLMKey` 판독 | `value=""` 무조건; 4자 초과 시 마지막 4자만, 4자 이하는 힌트 억제 |
| E-9 | 프로필 UI 중복 | `internal/web/root.templ` `profileSwitch` + `profileManager` 판독 | select 기반 전환과 카드 기반 전환이 병존 |
| E-10 | autonomy stub 미완 | `internal/web/handlers.go` `autonomyToggleLinkHTML` + `internal/web/autonomy.go` `handleAutonomyTiers` 판독 | GET 전용, form/action 부재, CSS 클래스·`data-i18n` 부재 |
| E-11 | autonomy 영속화 대상 부재 | `grep -rn "AutonomyTier" internal/ --include='*.go' \| grep -v _test.go` | 런타임 reader는 env(`MOAI_AUTONOMY_TIER`) 하나; writer는 init 경로 `ApplyAutonomyTierBundle`; config yaml 필드 부재 |
| E-12 | i18n 4-로케일 구조 | `internal/web/assets/i18n.js` 판독 | `window.MOAI_I18N`에 en/ko/ja/zh 4블록. governance 테스트 3종 실존 |

---

## §C 품질 게이트

| 게이트 | 명령 | 통과 기준 |
|--------|------|-----------|
| 빌드 | `go build ./...` | exit 0 |
| 크로스 플랫폼 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 테스트 | `go test ./...` | exit 0 |
| 커버리지 | `go test -cover ./internal/web/...` | ≥ 90.0% |
| 린트 | `golangci-lint run --timeout=2m` | NEW 이슈 0 (베이스라인 대비) |
| i18n | `go test ./internal/web/ -run 'TestI18n'` | exit 0, 예외 0건 |
| templ 동기화 | `templ generate && git diff --exit-code -- 'internal/web/*_templ.go'` | exit 0 |
| SPEC 린트 | `moai spec lint .moai/specs/SPEC-WEB-CONSOLE-REDESIGN-001/spec.md --strict` | error 0 |

---

## §D AC 매트릭스

### §D.1 M1 — 죽은 설정 철거 + 렌더 표면 복구

**AC-WCR-001** (REQ-WCR-001)
*Given* 스키마가 `workflow.token_budget.{plan,run,sync}` 필드를 등록하지 않은 상태에서,
*When* `settings.AllFields()`의 이름 집합을 조회하면,
*Then* 세 이름 중 어느 것도 존재하지 않는다.

```bash
go test ./internal/web/ -run TestDeadConfigAbsentFromSchema -v
# 테스트 본문: settings.AllFields()에 "workflow.token_budget.plan|run|sync" 부재 단언
```

**AC-WCR-002** (REQ-WCR-002)
*Given* 스키마가 `workflow.auto_clear.*` 4개 필드를 등록하지 않은 상태에서,
*When* `settings.AllFields()`의 이름 집합을 조회하면,
*Then* `workflow.auto_clear.enabled` / `.after_plan` / `.after_run` / `.token_threshold` 모두 부재한다.

```bash
go test ./internal/web/ -run TestDeadConfigAbsentFromSchema -v
```

**AC-WCR-003** (REQ-WCR-003)
*Given* `token_budget`와 `auto_clear` 키를 담은 기존 `workflow.yaml`이 존재하는 상태에서,
*When* config 로더가 그 파일을 로드하면,
*Then* 오류 없이 로드되고 해당 값이 struct에 바인딩되며, 접근자 함수가 그 값을 반환한다.

```bash
go test ./internal/config/ -run TestWorkflowBackwardCompat -v
# 테스트 본문: t.TempDir()에 두 키를 담은 workflow.yaml을 쓰고 로드 → 접근자 반환값 단언
```

**AC-WCR-004** (REQ-WCR-004)
*Given* `git_strategy` 섹션이 렌더 메타에 등록된 상태에서,
*When* 콘솔 페이지를 렌더하면,
*Then* `git_strategy.mode`와 3개 profile의 `merge_method` 입력 컨트롤이 HTML에 존재한다.

```bash
go test ./internal/web/ -run TestGitStrategyRendered -v
# 테스트 본문: 렌더된 HTML에 name="git_strategy.mode" + name="git_strategy.<p>.merge_method" ×3 존재 단언
```

### §D.2 M2 — 탭 재편

**AC-WCR-010** (REQ-WCR-010)
*Given* `consoleTabs()`가 9개 탭을 반환하는 상태에서,
*When* 탭 목록을 순서대로 조회하면,
*Then* identity, language, launch, llm, workflow, git-worktree, audit, agentfm, report 순으로 정확히 9개다.

```bash
go test ./internal/web/ -run TestConsoleTabsOrder -v
# 테스트 본문: len(consoleTabs()) == 9 AND 각 인덱스의 ID를 순서대로 단언 (집합이 아니라 순서)
```

**AC-WCR-011** (REQ-WCR-011)
*Given* 워크플로우 탭이 렌더된 상태에서,
*When* 해당 패널의 입력 이름 집합을 조회하면,
*Then* `workflow.execution_mode`, `workflow.default_mode`, `workflow.agentic_loop.max_iterations`, `workflow.loop_prevention.*`가 모두 존재한다 (F3 — 산문 소비 실존이므로 유지).

```bash
go test ./internal/web/ -run TestWorkflowTabRetainsProseConsumedFields -v
```

**AC-WCR-012** (REQ-WCR-012)
*Given* Git·워크트리 탭이 렌더된 상태에서,
*When* 해당 패널의 입력 이름 집합을 조회하면,
*Then* `git_strategy.mode`, 3× `merge_method`, `workflow.worktree.*` 4종, `workflow.branch_guard.enabled`가 모두 존재한다.

```bash
go test ./internal/web/ -run TestGitWorktreeTabFields -v
```

**AC-WCR-013** (REQ-WCR-013)
*Given* 감사 탭이 렌더된 상태에서,
*When* 해당 패널의 입력 이름 집합을 조회하면,
*Then* `workflow.audit.model`과 `workflow.audit.gates.{claude,codex,glm}`가 존재하며, 이 필드들의 영속화 섹션 ID는 변경 전과 동일하다.

```bash
go test ./internal/web/ -run TestAuditTabFields -v
# 테스트 본문: 렌더 존재 단언 + settings.FieldDef의 Section/Persist가 SectionWorkflow/seam 유지 단언
```

### §D.3 M3 — 위젯 정책

**AC-WCR-020** (REQ-WCR-020)
*Given* 모든 bool FieldDef가 렌더된 상태에서,
*When* 렌더된 HTML에서 `type="checkbox"` 출현 횟수를 세면,
*Then* 0이고, 각 bool 필드마다 `type="radio"` 입력이 정확히 2개 존재한다.

```bash
go test ./internal/web/ -run TestBoolFieldsRenderAsRadio -v
# 테스트 본문: (a) 전체 렌더 HTML의 type="checkbox" 개수 == 0
#              (b) settings.AllFields() 중 TypeBool 각각에 대해 radio 입력 2개 단언
#              (c) RED fixture: schemaToggleRow를 체크박스로 되돌린 분기에서 (a)가 실패함을 별도 케이스로 증명
```

**AC-WCR-021** (REQ-WCR-021)
*Given* bool 필드가 라디오로 렌더된 상태에서,
*When* hidden companion 입력의 존재와 `parseSchemaForm`의 동작을 검사하면,
*Then* 모든 bool 필드에 `name="<field>__present"` hidden 입력이 존재하고, 해당 필드를 전혀 담지 않은 POST는 값 보존으로 해석되며, `parseSchemaForm` 본문은 변경되지 않았다.

```bash
go test ./internal/web/ -run 'TestSchemaTogglePresentCompanion|TestParseSchemaFormUnchanged' -v
git diff origin/main -- internal/web/schemaform.go | grep -c '^[+-].*parseSchemaForm' || true
# parseSchemaForm 함수 본문에 대한 diff가 0 라인임을 리뷰에서 확인 (함수 시그니처/본문 무변경)
```

**AC-WCR-022** (REQ-WCR-022)
*Given* 닫힌 집합 필드들이 select/라디오로 전환된 상태에서,
*When* 각 필드의 `FieldDef.Type`과 `Validate`를 조회하면,
*Then* `workflow.execution_mode`, `workflow.default_mode`, `workflow.audit.model`, `workflow.audit.gates.*`, `harness.*`가 모두 `TypeSelect` 또는 `TypeRadio`이고 `Validate != nil`이며, 집합 밖 값의 저장이 거부된다.

```bash
go test ./internal/web/ -run TestClosedSetFieldsAreClosedWidgets -v
go test ./internal/settings/ -run TestClosedSetRejectsOutOfSetValue -v
```

**AC-WCR-023** (REQ-WCR-023)
*Given* 위젯 전환이 완료된 상태에서,
*When* `TypeText`로 남은 필드 이름 집합을 조회하면,
*Then* 그 집합은 화이트리스트(`feedback.repository`, `observability.report_dir`, `observability.trace_dir`, GLM 키)의 부분집합이다.

```bash
go test ./internal/web/ -run TestFreeTextWhitelist -v
# 테스트 본문: 화이트리스트 외 TypeText 필드가 발견되면 그 이름을 출력하며 실패
```

### §D.4 M4 — 서드파티 LLM 탭

**AC-WCR-030** (REQ-WCR-030)
*Given* GLM 모델 티어 4개 필드가 select로 전환된 상태에서,
*When* 각 필드의 옵션 값 집합을 조회하면,
*Then* 정확히 `{glm-5.2, glm-5.1, glm-4.7, glm-4.5-air}`이며, 이 리터럴은 스키마 파일이 아니라 단일 명명 상수 컬렉션에서 파생된다.

```bash
go test ./internal/web/ -run TestGLMModelSelectOptions -v
grep -rn 'glm-4.5-air' internal/ --include='*.go' | grep -v _test.go | wc -l
# 기대: 1 (단일 SSOT 선언). 2 이상이면 리터럴 재선언(AP-2) 위반
```

**AC-WCR-031** (REQ-WCR-031)
*Given* 티어별 추론 강도 필드 4종이 존재하는 상태에서,
*When* 각 필드의 옵션 집합과 기본 선택값을 조회하면,
*Then* 옵션은 `{Max, High, None}`이고 기본값은 fable=Max, high=Max, medium=High, low=None이다.

```bash
go test ./internal/web/ -run TestGLMEffortTierDefaults -v
```

**AC-WCR-032** (REQ-WCR-032)
*Given* 티어 라벨이 Claude 대응 이름으로 표시되는 상태에서,
*When* 렌더된 HTML과 폼 입력 이름을 대조하면,
*Then* 표시 라벨은 Opus/Sonnet/Haiku/Fable이고 `name` 속성의 키는 `high`/`medium`/`low`/`fable`을 유지한다.

```bash
go test ./internal/web/ -run TestGLMTierLabelVsKey -v
# 테스트 본문: name 속성에 ".high"/".medium"/".low"/".fable" 존재 AND 라벨 i18n 키가 Claude 이름에 매핑
```

**AC-WCR-033** (REQ-WCR-033)
*Given* 서드파티 LLM 탭이 렌더된 상태에서,
*When* 탭 헤더 영역의 배지 텍스트를 조회하면,
*Then* 적용 원천이 세션 단위 effort 설정임을 명시하는 i18n 키가 존재하고, 4개 티어 추론 강도 필드가 store-only로 표시되며, "티어 값이 적용된다"는 취지의 문구는 존재하지 않는다.

```bash
go test ./internal/web/ -run TestGLMEffortScopeBadge -v
# 테스트 본문: (a) 배지 data-i18n 키 존재
#              (b) 4개 티어 effort 필드 각각에 store-only 마커 존재
#              (c) 배지 문구 i18n 값 4로케일 전부 존재
```

**AC-WCR-034** (REQ-WCR-034)
*Given* GLM 키가 저장된 상태에서,
*When* 기본 페이지를 렌더하고, 이어서 reveal 엔드포인트를 루프백/비루프백 출처로 각각 호출하면,
*Then* 기본 렌더 HTML에 평문 키가 부재하고, 루프백 호출은 평문을 반환하며, 비루프백 호출은 거부된다.

```bash
go test ./internal/web/ -run 'TestGLMKeyNeverEchoedByDefault|TestGLMKeyRevealLoopbackOnly' -v
# TestGLMKeyNeverEchoedByDefault는 기존 계약의 회귀 핀 — reveal 추가로 깨지지 않음을 증명
```

### §D.5 M5 — 프로필 UI

**AC-WCR-040** (REQ-WCR-040)
*Given* `profileManager` 컴포넌트가 제거된 상태에서,
*When* 렌더된 콘솔 HTML을 검사하면,
*Then* `class="profilemgr"` 및 그 하위 클래스가 부재한다.

```bash
go test ./internal/web/ -run TestProfileManagerCardAbsent -v
grep -c 'profilemgr' internal/web/root.templ
# 기대: 0
```

**AC-WCR-041** (REQ-WCR-041)
*Given* 프로필 CRUD 컨트롤이 프로필 바에 배치된 상태에서,
*When* 렌더된 HTML의 form 중첩 구조를 검사하면,
*Then* 생성/이름변경/삭제 `<form>` 요소 중 어느 것도 메인 설정 `<form class="form">` 내부에 위치하지 않는다.

```bash
go test ./internal/web/ -run TestProfileFormsNotNested -v
# 테스트 본문: HTML 파서로 <form class="form"> 서브트리를 잘라내고 그 안에 <form> 노드가 0개임을 단언
#              (문자열 grep이 아니라 파싱 — 중첩은 구조 문제이지 텍스트 문제가 아니다)
```

**AC-WCR-042** (REQ-WCR-042)
*Given* rename 핸들러가 등록된 상태에서,
*When* 정상 이름변경 / 기본 프로필 / 현재 활성 프로필 / 이름 충돌 / 경로 탈출 문자를 각각 POST하면,
*Then* 첫 케이스만 성공하고 나머지 넷은 거부되며 각각 사유 배너를 반환한다.

```bash
go test ./internal/web/ -run TestProfileRename -v
# 서브테스트 5종: ok / default / current / conflict / traversal
```

### §D.6 M6 — autonomy stub

**AC-WCR-050** (REQ-WCR-050)
*Given* autonomy stub 제거가 적용된 상태에서,
*When* 렌더된 콘솔 HTML과 라우트 테이블을 검사하면,
*Then* form/action 없는 맨 autonomy 링크가 부재하고 `/autonomy/tiers` 라우트가 부재한다.

```bash
go test ./internal/web/ -run TestAutonomyStubResolved -v
# 본문: 렌더 HTML에 "/autonomy/tiers" 부재 AND mux에 해당 라우트 미등록
# (plan.md §B D3 결정 (c) 제거로 고정)
```

### §D.7 M7 — 횡단

**AC-WCR-060** (REQ-WCR-060)
*Given* 신규 i18n 키가 추가된 상태에서,
*When* i18n 키 집합 동등성 및 governance 테스트를 실행하면,
*Then* 전부 통과하고 예외(allowlist) 추가 건수가 0이다.

```bash
go test ./internal/web/ -run 'TestI18n' -v
git diff origin/main --stat -- internal/web/i18n_untranslated_allowlist_test.go
# 기대: allowlist 파일 무변경 (신규 예외 0건)
```

**AC-WCR-061** (REQ-WCR-061)
*Given* `.moai/config/sections/*.yaml`에 편집이 발생한 경우,
*When* 해당 커밋의 파일 목록을 검사하면,
*Then* `internal/template/templates/` 하위 대응 파일과 `make build` 재생성 산출물이 같은 커밋에 포함되어 있다. yaml 편집이 없었다면 이 AC는 공백 통과(vacuous)임을 명시적으로 기록한다.

```bash
git show --name-only HEAD | grep -E '^\.moai/config/sections/' || echo "NO_YAML_EDIT (vacuous pass — 근거 기록 필요)"
# yaml 편집이 있으면: 같은 출력에 internal/template/templates/ 경로가 함께 나타나야 한다
go test ./internal/template/ -run TestRuleTemplateMirror -v
```

**AC-WCR-062** (REQ-WCR-062)
*Given* 템플릿 미러가 갱신된 경우,
*When* 템플릿 중립성 가드를 실행하면,
*Then* 통과한다.

```bash
go test ./internal/template/ -run 'TestInternalContentLeak|TestSplitHarnessNamespaceNoLeak' -v
# 기존 가드 실행이 권위 (§A.1 규칙 4 — 정규식 재구현 금지)
```

**AC-WCR-063** (REQ-WCR-063) — **PASS-WITH-DEBT (사용자 승인 2026-08-13)**
*Given* 구현이 완료된 상태에서,
*When* `internal/web` 패키지 커버리지를 측정하면,
*Then* 90.0% 이상이다.

```bash
go test -cover ./internal/web/... 2>&1 | grep -E 'coverage: [0-9.]+%'
```

**실측**: 73.5% — 90% 기준 미달, 그러나 **구조적 상한 80.7%**이므로 90%는 도달 불가. 패키지
구문의 83%(4323/5208)가 `templ generate` 산출물이며, 그중 927구문(패키지 전체의 17.8%)은
templ이 모든 write 뒤에 기계적으로 붙이는 `if templ_err != nil { return }` / `ctx.Err()` /
`NopComponent` 보일러플레이트로 `strings.Builder`·`httptest` 버퍼 렌더에서 도달 불가하다.
손으로 쓴 코드만 **91.1%**(이미 90% hand-written-code 기준 충족). 상세한 실측과 권고는
`progress.md §E.2.3`에 있다. **DEBT**: AC-WCR-063의 분모 재정의(`*_templ.go` 제외 또는
hand-written-code ≥ 90%로 재정의)는 SPEC 본문 변경이므로 manager-spec 소관이며, 본
sync-phase에서는 사용자 승인된 PASS-WITH-DEBT로 기록한다. 본 sync에서 코드를 추가하지
않는다 — 이 판정은 이미 merged된 run-phase 코드(`b0d3b61f8`)와 orchestrator의
`go test ./internal/web/...` 실측 결과에 기반한 것이다.

---

## §E Definition of Done

1. AC-WCR-001 ~ AC-WCR-063 전 25건이 PASS이거나, FAIL 항목이 사용자 승인된 debt로 기록되어 있다.
2. §C 품질 게이트 8종이 전부 통과한다.
3. `plan.md` §B D3(autonomy 결말) · D4(bool 3-상태 여부) 두 미해소 표식이 해소되어 있다 — D3=(c) 제거, D4=(i) 2-option radio + `__present` 보존으로 확정.
4. `plan.md` §G 안티패턴 7종 중 어느 것도 최종 diff에 존재하지 않는다.
5. `progress.md` §E.2/§E.3에 run-phase 증거가 기록되어 있다.
