# SPEC-WEB-CONSOLE-REDESIGN-001 — 구현 계획

> 대응 SPEC: `.moai/specs/SPEC-WEB-CONSOLE-REDESIGN-001/spec.md`
> AC 매트릭스: `.moai/specs/SPEC-WEB-CONSOLE-REDESIGN-001/acceptance.md`

---

## §A Context

### §A.1 작업 위치

- 브랜치: `feat/web-console-redesign` (base: `origin/main` @0de8517e5)
- 워크트리: `.claude/worktrees/web-redesign`
- 모든 경로는 프로젝트 루트 상대 경로로 기술한다.

### §A.2 아키텍처 요약 (실측)

폼은 스키마 주도다. 세 계층이 분리되어 있고, 변경 지점을 고르는 일이 이 SPEC의 난이도 대부분을 차지한다.

| 계층 | 파일 | 역할 |
|------|------|------|
| 스키마 SSOT | `internal/settings/schema_sections.go` | `FieldDef` 정의 + `SchemaSectionIDs()` 등록 |
| 렌더 선택 | `internal/web/schemaform.go` | `consoleTabs()` (탭 nav) · `schemaSectionMetas()` (렌더 대상 섹션) · `parseSchemaForm` (파싱) |
| 위젯 | `internal/web/fieldsets.templ` | `schemaTextRow` / `schemaNumberRow` / `schemaToggleRow` / `schemaSelectRow` / `schemaRadioRow` |
| 셸 | `internal/web/root.templ` | 탭 nav 마크업 · tabpanel · 프로필 UI · 메인 form |
| 영속화 | `internal/settings` (`ApplySchemaEdits`) | `PersistSeam`(yamlpatch) / `PersistTypedSection`(typed) |

**핵심 함의**: 필드 하나를 없애거나 위젯 하나를 바꾸는 일은 대개 **한 곳**에서 이루어진다. 필드 목록을 손대면 스키마에서, 위젯 모양을 손대면 templ에서, 무엇을 보여줄지를 손대면 `schemaform.go`에서. 이 경계를 흐리는 변경(예: 특정 필드를 위해 위젯에 이름 기반 분기 추가)은 §G 안티패턴이다.

### §A.3 PRESERVE 목록 (건드리지 않는다)

- `.moai/config/sections/workflow.yaml`의 `token_budget` / `auto_clear` 키 (REQ-WCR-003)
- `internal/config`의 대응 struct 멤버 + `workflow_accessors.go`의 접근자 함수 시그니처
- `parseSchemaForm`의 bool 분기(`__present` companion 해석) — REQ-WCR-021
- `glmkey.go`의 `value=""` 무조건 계약 + 4자 이하 키 힌트 억제 로직 — REQ-WCR-034
- `internal/web/board.templ` / `board.go` (SPEC 보드) — §C Out of Scope
- 미렌더 섹션 8개 중 `git_strategy` 외 나머지의 FieldDef

---

## §B 되돌리기 어려운 결정 (먼저 검토할 것)

아래 D1-D4는 **변경 가능성이 가장 높은 결정**이라 맨 앞에 둔다. 나머지 밀스톤은 이 결정들이 확정된 뒤의 기계적 실행이다.

### D1 — GLM 추론 강도 배지의 문구 (브리프 정정 반영)

브리프는 "티어 중 하나만 실제 적용되니 그 티어를 배지로 명시하라"고 지시했다. 실측은 이보다 강하다.

```
SessionGLMReasoningState()            → 조건 없이 reasoning-max 반환
SessionGLMReasoningStateForEffort(e)  → e != "" 이면 CollapseClaudeEffortToGLM(e)
                                        e == "" 이면 위 세션 기본값
```

`e`는 **LLM 탭의 `effort_level` 환경설정**(웹에서 설정하는 세션 단위 값)이며, `llm.glm.models.{high,medium,low,fable}`와 아무 관계가 없다. 따라서 "어느 티어가 적용된다"는 문장 자체가 참이 아니다.

**결정**: 배지는 적용 **원천**을 명시한다 — "추론 강도는 LLM 탭의 세션 effort 설정에서 파생되며, 아래 티어별 값은 저장만 됩니다". 4개 티어 effort select는 모두 store-only로 라벨링한다.

**되돌리기 비용**: 문구를 나중에 바꾸면 4개 로케일 × 최소 2개 키를 재작업해야 한다. 지금 확정하는 편이 싸다.

### D2 — 신규 탭의 i18n 키 정책

기존 `consoleTabs()`는 탭 라벨에 `sec.<section>.title`을 재사용하며, 선행 SPEC은 "0 NEW `sec.*` keys"를 제약으로 걸었다. 신규 탭 2종(Git·워크트리 / 감사)은 기존 섹션에 1:1 대응하지 않는다 — Git·워크트리 탭은 `git_strategy` 섹션과 `workflow` 섹션의 필드를 **함께** 담고, 감사 탭은 `workflow` 섹션의 하위 트리만 담는다.

**결정**: 탭은 섹션과 분리된 개념으로 승격한다. 신규 `tab.<id>.title` / `tab.<id>.desc` 키를 도입하고, 기존 7탭은 하위 호환을 위해 현행 `sec.*` 키를 계속 참조한다(혼재 허용). "0 NEW sec.* keys" 제약은 신규 탭에 적용되지 않는다 — 그 제약은 섹션 재사용이 가능했던 시절의 것이다.

**되돌리기 비용**: 키 네임스페이스는 4개 로케일과 governance 테스트에 동시 반영되므로 나중 변경 비용이 크다.

### D3 — autonomy tier의 영속화 대상 [NEEDS CLARIFICATION: autonomy tier 승격 vs 제거]

실측 결과:

- 런타임 reader: `config.AutonomyTier()` → `os.Getenv(MOAI_AUTONOMY_TIER)` 하나뿐 (`internal/config/autonomy.go`)
- init 경로 writer: `ApplyAutonomyTierBundle(projectRoot, userSettingsPath, projectSettingsPath, persistedTier)` — `opts.AutonomyTier`를 받아 배포된 settings의 `defaultMode`로 번역 (`internal/core/project/autonomy_bundle.go`)
- **config yaml 필드는 존재하지 않는다.** `.moai/config/sections/*.yaml` 어디에도 autonomy tier를 담는 키가 없다

따라서 콘솔 필드로 승격하려면 둘 중 하나가 필요하다:

- **(a)** `settings.local.json`의 env 블록에 `MOAI_AUTONOMY_TIER`를 기록 — 런타임 reader와 직결되지만, `settings.local.json`은 CLAUDE.local.md §2에서 **런타임 관리 파일**로 규정되어 콘솔이 쓰는 것이 계약상 애매하다
- **(b)** 저장 시 `ApplyAutonomyTierBundle`을 호출해 settings의 `defaultMode`를 갱신 — init 전용으로 설계된 함수를 런타임 저장 경로에서 재사용하게 되며, 사용자가 손으로 바꾼 `defaultMode`를 덮어쓸 위험이 있다
- **(c)** stub과 핸들러(`internal/web/autonomy.go` + `handlers.go`의 링크 주입 3함수)를 제거하고 사유를 기록

**권고: (c) 제거.** (a)는 파일 소유권 계약 위반이고 (b)는 사용자 편집 덮어쓰기 위험이 있다. 두 위험 모두 "설정 하나를 콘솔에서 고를 수 있게 한다"는 이득보다 크다. REQ-WCR-050의 두 갈래 중 후자를 택하고, 제거 사유를 이 문단으로 인용한다.

**이 마커는 Implementation Kickoff Approval 이전에 오케스트레이터가 사용자에게 확인해야 한다.** (a)/(b)를 택하면 M6의 파일 목록과 AC-WCR-050이 달라진다.

### D4 — bool 라디오의 "미설정" 표현 [NEEDS CLARIFICATION: bool 3-상태 여부]

현행 `schemaToggleRow`는 체크박스 + hidden `__present` companion 조합으로 두 상태를 구분한다: 체크 해제(false 기록) vs 미제출(값 보존). 2옵션 라디오(사용/미사용)로 바꾸면 **둘 다 항상 제출되므로** "값 보존" 상태를 사용자가 선택할 수 없게 된다.

두 해석이 가능하다:

- **(i)** 2옵션 유지 — 라디오는 항상 하나가 선택되므로 disk 값이 초기 선택으로 렌더되고, 저장은 항상 명시적 true/false를 기록한다. 브리프가 지시한 형태. `__present`는 계속 제출되어 파서는 무변경.
- **(ii)** 3옵션(사용 / 미사용 / 프로젝트 기본값) — 키 부재 상태를 UI에 노출. 파서 변경 필요.

**권고: (i).** 브리프가 명시적으로 2옵션을 지시했고, `__present`를 유지하면 `parseSchemaForm`이 무변경으로 남는다(REQ-WCR-021). 다만 **부작용을 명시한다**: 지금까지 "키 부재"였던 bool 필드가 첫 저장 시 명시적 값으로 기록된다. `workflow.branch_guard.enabled`처럼 배포 기본이 "키 부재"인 필드는 이 변화가 관측 가능한 동작 변화를 낳을 수 있다 — AC-WCR-021이 이 경계를 고정한다.

---

## §C Pre-flight (구현 착수 전 실행)

```bash
# 1. 베이스라인
git branch --show-current
git rev-parse HEAD

# 2. 빌드 + 크로스 플랫폼
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. 린트 베이스라인 (NEW vs pre-existing 구분용)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. 현재 커버리지 베이스라인
go test -cover ./internal/web/... ./internal/settings/...

# 5. i18n 게이트 베이스라인
go test ./internal/web/ -run 'TestI18n' -v 2>&1 | tail -20

# 6. 죽은 설정 재확인 (라인 앵커가 아닌 content-token으로)
grep -rn "WorkflowPlanTokens\|WorkflowRunTokens\|WorkflowSyncTokens\|WorkflowAutoClearEnabled" \
  internal/ pkg/ cmd/ --include='*.go' | grep -v '_test.go'
# 기대: 정의부만 (workflow_accessors.go), 호출자 0건

# 7. 살아 있는 설정 재확인 (F3 — 제거 금지 대상)
grep -rn "loop_prevention\|agentic_loop" .claude/skills/ .claude/rules/ docs-site/ | head
grep -rn "BranchGuard\|WorktreeAutoCreate\|worktree.auto" internal/ --include='*.go' | grep -v '_test.go' | head
```

Pre-flight 6번이 호출자를 하나라도 반환하면 **M1을 중단하고 blocker를 보고한다** — 죽은 설정 판정의 전제가 무너진 것이다.

---

## §D 밀스톤

### M1 — 죽은 설정 철거 + git_strategy 렌더 표면 복구

REQ-WCR-001 / -002 / -003 / -004.

작업:

1. `internal/settings/schema_sections.go` `seamSectionFields()`에서 `token_budget` 3행 + `auto_clear` 4행 제거. 제거 사유를 주석으로 남긴다(선례: 같은 파일의 M4 다이어트 주석 스타일).
2. `internal/config`의 struct 멤버와 `workflow_accessors.go` 접근자는 **무접촉**. 접근자가 호출자 0건이 되므로 린트가 unused를 경고할 수 있다 — exported 함수이므로 경고 대상이 아님을 확인하고, 경고가 발생하면 `//nolint` 대신 주석으로 보존 의도를 명시한다.
3. `internal/web/schemaform.go` `schemaSectionMetas()`에 `settings.SectionGitStrategy` 항목 추가.
4. 회귀 가드 테스트: 제거된 7개 이름이 `settings.AllFields()`에 부재함을 고정.

파일:

- `internal/settings/schema_sections.go` (수정)
- `internal/web/schemaform.go` (수정)
- `internal/settings/schema_sections_test.go` 또는 신규 `internal/web/dead_config_guard_test.go` (신규/수정)

### M2 — 탭 7→9 재편

REQ-WCR-010 / -011 / -012 / -013.

작업:

1. `consoleTabs()`를 9개 항목으로 재작성. D2 결정에 따라 신규 탭은 `tab.*` 키, 기존 탭은 현행 `sec.*` 키.
2. 탭↔패널 매핑을 재정의한다. 현재 `root.templ`은 `schemaSectionMetas()`를 순회하며 `data-panel={string(meta.ID)}`로 패널을 만든다 — 즉 **패널 ID가 섹션 ID와 동일**하다는 전제가 박혀 있다. 신규 탭은 여러 섹션의 필드를 섞으므로 이 전제를 깨야 한다.
   - 접근: `schemaform.go`에 탭→필드 매핑을 명시하는 함수(예: `tabFields(tabID) []settings.FieldDef`)를 도입하고, `root.templ`은 탭 목록을 순회하며 패널을 만든다. 섹션 메타는 패널 내부의 fieldset 헤더로 남긴다.
3. 감사 탭의 4개 필드는 현재 `SectionWorkflow`의 seam 필드다(`workflow.audit.*`) — 섹션 재분류 없이 탭 배치만 바꾼다. 섹션 ID를 바꾸면 영속화 경로가 바뀌므로 **금지**.
4. 워크플로우 탭에서 worktree/branch_guard/audit 필드가 빠지고 Git·워크트리/감사 탭으로 이동한다. 이동은 렌더 배치 변경일 뿐 영속화 경로 무변경.

파일:

- `internal/web/schemaform.go` (수정 — `consoleTabs()` + 탭→필드 매핑 신규)
- `internal/web/root.templ` (수정 — 패널 생성 루프)
- `internal/web/fieldsets.templ` (수정 — 탭 단위 fieldset 렌더 진입점)
- `internal/web/assets/i18n.js` (수정 — 신규 `tab.*` 키 × 4 로케일)
- `internal/web/restyle_test.go` / `console_ux_fix_test.go` (수정 — 탭 수·순서 단언)

### M3 — 위젯 정책

REQ-WCR-020 / -021 / -022 / -023.

작업:

1. `schemaToggleRow`를 2옵션 라디오로 재작성(D4 (i)). hidden `__present` 입력은 **그대로 유지**한다. 라벨 키는 신규 `opt.enabled` / `opt.disabled`(사용/미사용) 2종 — 전 필드가 공유하므로 필드당 키 폭증이 없다.
2. `parseSchemaForm` **무변경**을 명시적으로 검증한다(diff 0 라인).
3. 닫힌 집합 필드를 `withSelect`/`withRadio`로 전환:
   - `workflow.execution_mode`, `workflow.default_mode` — 옵션 집합의 SSOT를 찾아 파생한다. 리터럴 재선언 금지(선례: `merge_method`가 `config.ValidMergeMethods()`에서 파생).
   - `workflow.audit.model`, `workflow.audit.gates.{claude,codex,glm}` — 현재 `TypeText`이며 enum 검증이 M3 config-read 계층(`activeAuditBackend`)에 있다. 옵션 집합을 그 계층에서 파생하고, 파생 불가하면 config 상수를 SSOT로 노출한 뒤 참조한다.
   - `harness.default_profile`, `harness.effort_mapping.*`, `harness.evaluator.memory_scope`, `harness.mode_defaults.*` — 각각의 유효 집합을 harness 패키지에서 파생.
4. 자유 텍스트 잔류 화이트리스트를 테스트로 고정(REQ-WCR-023).

파일:

- `internal/web/fieldsets.templ` (수정 — `schemaToggleRow`)
- `internal/settings/schema_sections.go` (수정 — `withSelect`/`withRadio` 적용)
- `internal/config/*` 또는 `internal/harness/*` (수정 — 옵션 집합 SSOT 노출자가 없으면 추가)
- `internal/web/assets/i18n.js` (수정 — `opt.enabled`/`opt.disabled` + 신규 옵션 라벨 × 4 로케일)
- `internal/web/widget_policy_test.go` (신규)

### M4 — 서드파티 LLM 탭

REQ-WCR-030 / -031 / -032 / -033 / -034.

작업:

1. `llmFields()`의 4개 `TypeText`를 `withSelect`로 전환. 모델 집합 `{glm-5.2, glm-5.1, glm-4.7, glm-4.5-air}`는 명명 상수 컬렉션으로 정의한다(CLAUDE.local.md §14 — 리터럴 산재 금지). 기존 `glmContextWindows` 같은 GLM 모델 테이블이 있으면 그것을 SSOT로 삼는다.
2. 티어별 추론 강도 필드 4종 신규. 값 집합 `{Max, High, None}`, 기본값 Fable=Max / high=Max / medium=High / low=None. 영속화 경로는 기존 `llm` typed 경로를 재사용한다.
3. 티어 라벨을 Claude 대응 이름으로 표시하되 내부 키는 불변(REQ-WCR-032) — 표시 라벨만 i18n 키로 매핑한다.
4. D1 배지 렌더. 배지는 탭 헤더 수준 1회 렌더(필드마다 반복 금지 — 선례: `agentfm-gridnote`가 반복 힌트를 헤더로 승격한 사례).
5. GLM 키 reveal 토글:
   - 서버: 신규 엔드포인트가 루프백 바인드에서만 응답하고 평문을 1회 반환. `glmcred.Load()` 재사용.
   - 클라이언트: 토글 조작 시에만 fetch. 기본 렌더는 `value=""` 유지.
   - 보안 계약 문서화: 미표시 경로의 never-echo 계약은 무손상이며, reveal은 루프백 전용 콘솔이라는 전제에 기댄다.

파일:

- `internal/settings/schema_sections.go` (수정 — `llmFields()`)
- `internal/web/glmkey.go` (수정 — reveal 핸들러 + 뷰모델)
- `internal/web/app.go` (수정 — 라우트 등록)
- `internal/web/fieldsets.templ` (수정 — 배지 + reveal 컨트롤)
- `internal/web/assets/app.js` (수정 — reveal 토글)
- `internal/web/assets/i18n.js` (수정 — 배지/토글/티어 라벨 × 4 로케일)
- `internal/web/glmkey_test.go` (수정)
- `internal/web/glm_badge_test.go` (신규)

### M5 — 프로필 UI 통합

REQ-WCR-040 / -041 / -042.

작업:

1. `root.templ`에서 `profileManager` 컴포넌트와 그 호출부 제거.
2. `profileSwitch`를 프로필 바로 확장: select + 생성/이름변경/삭제 버튼. **HTML 제약**: 생성·삭제·이름변경은 각각 독립 POST 폼이므로 `<form>` 요소는 메인 설정 폼 바깥에 둔다. 버튼은 프로필 바에 두고 `form=` 속성으로 바깥 폼을 참조하거나, 프로필 바 자체를 메인 폼 바깥에 배치한다(현행 `profileSwitch`가 이미 메인 폼 바깥에 있으므로 후자가 자연스럽다).
3. `handleProfileRename` 신규. 거부 조건: 기본 프로필 / 현재 활성 프로필 / 이름 충돌 / 경로 탈출 문자. 경로 탈출 검증은 기존 `profile_traversal_test.go`가 고정한 규칙을 재사용한다.
4. `handleProfileDelete`의 기존 거부 규칙(`profileCanDelete`)과 대칭이 되도록 rename의 거부 규칙을 배치한다.

파일:

- `internal/web/root.templ` (수정 — `profileManager` 제거, `profileSwitch` 확장)
- `internal/web/profile_crud.go` (수정 — `handleProfileRename` 추가)
- `internal/web/app.go` (수정 — `/profile/rename` 라우트)
- `internal/web/assets/i18n.js` (수정 — rename 라벨/오류 × 4 로케일)
- `internal/web/profile_crud_test.go` (수정)
- `internal/web/profile_traversal_test.go` (수정 — rename 경로 탈출 케이스)

### M6 — autonomy stub 결말

REQ-WCR-050. **D3 확정 후 착수한다.**

권고안 (c) 기준 작업:

1. `internal/web/handlers.go`에서 `autonomyToggleLinkHTML` 상수와 `injectAutonomyToggleLink` 함수, 그리고 렌더 경로의 호출부 제거.
2. `internal/web/autonomy.go` (`handleAutonomyTiers` + `renderAutonomyToggle`) 제거.
3. `internal/web/app.go`에서 `/autonomy/tiers` 라우트 제거.
4. `internal/config/autonomy_tiers.go`의 `TierToggleOptions` 등은 **무접촉 보존** — init 경로(`ApplyAutonomyTierBundle`)와 CLI 플래그가 여전히 소비한다. 웹 표면만 제거한다.
5. 제거 사유를 `internal/web/autonomy_console_test.go`를 회귀 가드로 전환해 고정한다: 렌더된 페이지에 `/autonomy/tiers` 문자열이 부재.

파일:

- `internal/web/handlers.go` (수정)
- `internal/web/autonomy.go` (삭제)
- `internal/web/app.go` (수정)
- `internal/web/autonomy_test.go` (삭제 또는 축소)
- `internal/web/autonomy_console_test.go` (수정 — 부재 가드로 전환)

### M7 — 횡단 마감

REQ-WCR-060 / -061 / -062 / -063.

작업:

1. i18n 4-로케일 파리티: 신규 키 전부가 en/ko/ja/zh에 존재. `TestI18nKeySetParity` 및 governance 테스트를 예외 0건으로 통과.
2. `.moai/config/sections/*.yaml`을 건드린 경우에만 Template-First 미러 + `make build` (M1의 죽은 설정 제거는 yaml 키를 남기므로 미러 대상이 아닐 수 있다 — 실제 yaml 편집이 발생했는지로 판정).
3. 템플릿 중립성: 미러된 파일에 SPEC ID / REQ 토큰 / 날짜 / SHA 부재.
4. 커버리지 90% 달성.
5. `templ generate` 산출물(`*_templ.go`) 동기화 — templ 소스를 고쳤으면 생성 파일도 같은 커밋에 담는다.

파일:

- `internal/web/assets/i18n.js`
- `internal/template/templates/**` (yaml 편집이 있었을 때만)
- `internal/web/*_templ.go` (생성)

---

## §E 밀스톤 의존 관계

```
M1 (스키마 필드 집합 확정)
 └→ M2 (탭 배치 — M1이 정한 필드 집합을 배치)
      ├→ M3 (위젯 — M2가 정한 배치 위에서 위젯 교체)
      └→ M4 (LLM 탭 — M2의 탭 구조 + M3의 select 전환 관례를 따름)
M5 (프로필 UI) — 독립. 병렬 가능
M6 (autonomy) — 독립. D3 확정 후 착수
M7 — 전 밀스톤 이후
```

M5·M6는 M1-M4와 파일이 겹치지 않는다(`root.templ`은 M2와 M5가 함께 건드리므로 **순차 처리**한다). M2 → M5 순서를 권고한다.

---

## §F 위험

| 위험 | 영향 | 완화 |
|------|------|------|
| 탭↔패널 전제 변경(§D M2-2)이 예상보다 광범위 | `root.templ` + `fieldsets.templ` 동시 재작업, 기존 렌더 테스트 다수 실패 | M2를 단독 커밋으로 분리하고 렌더 테스트를 먼저 갱신 |
| bool 라디오 전환이 "키 부재" 상태를 소멸시킴(D4) | `branch_guard.enabled`처럼 부재가 기본인 필드의 관측 가능한 동작 변화 | AC-WCR-021이 경계를 고정; 배포 기본이 부재인 필드 목록을 M3 착수 시 실측 |
| 닫힌 집합의 SSOT가 존재하지 않는 필드(harness.*, audit.*) | 리터럴 재선언 유혹 → §G 안티패턴 | SSOT가 없으면 **먼저 config/harness 패키지에 노출자를 추가**하고 참조 |
| GLM 키 reveal이 새 유출 표면 | 평문 노출 경로 신설 | 루프백 바인드 검증 + 기본 렌더 무변경 + reveal 전용 엔드포인트 분리 |
| `templ generate` 산출물 누락 | CI 미러/빌드 실패 | M7 체크리스트에 명시 |
| 병렬 세션이 같은 체크아웃을 만짐 | 커밋 오염 | 워크트리에서만 작업, `git add`는 명시 pathspec |

---

## §G 안티패턴

- **AP-1 — 위젯에 필드 이름 분기 추가**: `schemaToggleRow`나 `schemaSelectRow` 안에 `if f.Name == "..."` 분기를 넣는 것. 위젯은 `FieldDef`만 보고 렌더해야 한다. 필드별 차이는 `FieldDef`의 필드(Options / Description / EmptyLabel)로 표현한다.
- **AP-2 — 옵션 리터럴 재선언**: `{"squash","merge","rebase"}` 같은 집합을 스키마 파일에서 다시 쓰는 것. 기존 선례는 `config.ValidMergeMethods()` 파생이다. SSOT가 없으면 만들고 나서 참조한다.
- **AP-3 — 죽은 설정과 산문 소비 설정을 한 번에 제거**: F2(제거)와 F3(유지)는 인접하지만 처분이 정반대다. `auto_clear`를 지우면서 옆줄의 `agentic_loop`까지 지우는 사고가 이 SPEC의 가장 그럴듯한 실패다.
- **AP-4 — 섹션 ID 재분류로 탭 배치 해결**: `workflow.audit.*`를 신규 섹션으로 옮기면 영속화 경로가 바뀐다. 탭은 렌더 배치이고 섹션은 영속화 단위다.
- **AP-5 — yaml 키까지 삭제**: REQ-WCR-003 위반. 웹 표면 제거 ≠ 설정 제거.
- **AP-6 — i18n 키를 en만 추가하고 나머지 3개를 나중에**: governance 테스트가 즉시 실패하고, 부분 커밋이 CI를 빨갛게 만든다.
- **AP-7 — 중첩 form**: 프로필 생성/삭제/이름변경 폼을 메인 설정 form 안에 넣는 것. 유효하지 않은 HTML이며 브라우저가 조용히 폼을 잘라낸다.

---

## §H 교차 참조

- `spec.md` §A.1 (F1-F6 실측), §B (REQ), §C (Out of Scope)
- `acceptance.md` §B (실측 증거), §D (AC 매트릭스)
- `CLAUDE.local.md` §2 (Template-First), §6 (커버리지 90%), §14 (하드코딩 금지), §25 (템플릿 중립성)
- 선행 SPEC: WEB-CONSOLE-011(스키마 주도 확립) / -012(legacy alias 제거) / -013(handoff·cache) / -014(정직화 3계층)

---

## §G.1 Tier 편차 기록

본 SPEC은 `tier: L`로 분류되나(> 15 파일 · 다중 도메인 · 7 밀스톤) 산출물은 **3종**(spec.md / plan.md / acceptance.md)으로 한정한다. Tier L 표준 산출물의 `design.md` / `research.md`는 생략한다:

- `research.md`의 역할(코드베이스 실측)은 오케스트레이터가 위임 브리프에서 이미 수행했고, 그 결과가 spec.md §A.1 + acceptance.md §B에 흡수되어 있다.
- `design.md`의 역할(설계 결정)은 본 plan.md §B(되돌리기 어려운 결정 D1-D4)가 담당한다.

선례: SPEC-CLI-TUX-V3-003이 동일한 Tier L 편차를 명기하고 진행했다. plan-audit의 Tier L 산출물 완비 체크는 이 문단을 근거로 판정한다.
