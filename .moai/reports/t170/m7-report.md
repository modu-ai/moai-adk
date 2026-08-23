# SPEC-FEEDBACK-AUTO-SUBMIT-001 M7 — 웹 콘솔 노출

작업 트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, 착수 HEAD `38705eb85`, base `3210da7d3`.

## 1. Claim (주장)

| # | 주장 | 판정 |
|---|---|---|
| C1 | `feedback` 섹션이 `RouteExcluded` → `RouteSeam` 으로 재개방됐고, 이 반전이 SPEC-WEBCONF-SIMPLIFY-001 M3 의 기록된 결정을 뒤집는다는 사실이 커밋 본문에 명시됐다 | PASS |
| C2 | `feedback.auto_submit` 이 seam 영속화 필드로 등록되고 실제 seam 쓰기 왕복이 성립한다 | PASS |
| C3 | feedback 탭·패널과 두 필드 위젯이 렌더된 콘솔 HTML 에 실제로 나타난다 | PASS |
| C4 | `f.feedback.auto_submit.title`/`.desc` 가 en/ko/ja/zh 4로케일 전부에 존재한다 | PASS |
| C5 | AC-F-023 의 **웹 절반**(스키마·라우트·i18n)이 충족된다 | PASS |
| C6 | `internal/settings` · `internal/web` 두 패키지 전 테스트가 초록이고 `go vet` / `GOOS=windows go vet` 이 exit 0 이다 | PASS |

C5 는 **웹 절반에 한정된 판정**이다. AC-F-023 의 템플릿 절반(`internal/template/templates/.moai/config/sections/feedback.yaml` 미러 · `shipped_key_inventory.yaml` 항목 · `make build` · 중립성 grep)은 M8 소관이며 이 마일스톤은 그것을 주장하지 않는다.

## 2. Evidence (증거 — 실행 명령과 그 출력)

### 2.1 RED — 구현 이전 트리에서의 실패

```
$ go test ./internal/settings/ -run 'TestRouteForSectionTable|TestSeamSectionsMatchesRoutes|TestExcludedSectionsAllRejected|TestFeedbackAutoSubmitFieldRegistered|TestFeedbackSectionSeamWritable'
--- FAIL: TestRouteForSectionTable (0.00s)
    sectionroute_test.go:49: RouteForSection("feedback") = 0, want 2
--- FAIL: TestSeamSectionsMatchesRoutes (0.00s)
    sectionroute_test.go:62: SeamSections() length = 1, want 2 (workflow + feedback)
--- FAIL: TestFeedbackAutoSubmitFieldRegistered (0.00s)
    feedback_autosubmit_test.go:26: SectionFields(SectionFeedback) has no feedback.auto_submit field
--- FAIL: TestExcludedSectionsAllRejected (0.00s)
    sectionroute_test.go:94: ExcludedSections() length = 18, want 17 (19 M3 - workflow - feedback)
--- FAIL: TestFeedbackSectionSeamWritable (0.00s)
    feedback_autosubmit_test.go:56: ApplySchemaEdits(feedback.auto_submit): settings: unknown schema field "feedback.auto_submit"
FAIL	github.com/modu-ai/moai-adk/internal/settings	0.428s
```

```
$ go test ./internal/web/ -run 'TestScopeContractEditableSections|TestScopeContractExclusions|TestFeedbackPanelRendered|TestFeedbackAutoSubmitI18nKeysInAllLocales|TestFeedbackPanelFieldsWired'
--- FAIL: TestFeedbackPanelRendered (0.01s)
    feedback_panel_test.go:26: rendered console missing "data-tab=\"feedback\""
    feedback_panel_test.go:26: rendered console missing "data-panel=\"feedback\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.repository\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.auto_submit\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.auto_submit__present\""
--- FAIL: TestFeedbackAutoSubmitI18nKeysInAllLocales (0.00s)
    feedback_panel_test.go:39: i18n.js missing feedback key "f.feedback.auto_submit.title" in all 4 locales
    feedback_panel_test.go:39: i18n.js missing feedback key "f.feedback.auto_submit.desc" in all 4 locales
--- FAIL: TestFeedbackPanelFieldsWired (0.00s)
    feedback_panel_test.go:49: schemaPanelMeta(feedback).ID = "", want "feedback"
    feedback_panel_test.go:53: SectionFields(SectionFeedback) = 1 fields, want >= 2 (repository + auto_submit)
--- FAIL: TestScopeContractEditableSections (0.00s)
    scope_contract_test.go:43: section "feedback": route = 0, want RouteSeam (M7 reopened)
FAIL	github.com/modu-ai/moai-adk/internal/web	0.462s
```

RED 는 두 축을 모두 관측한다: 반전된 라우팅 기대(고정 테스트) **와** 아직 존재하지 않는 필드·패널·i18n 키(신규 테스트). 구현 코드는 이 출력을 얻은 뒤에 작성했다.

### 2.2 GREEN — AC-F-023 선택자 `-v` 실행 (`=== RUN` 확인)

acceptance.md §D.3 이 요구한 "선택자 실재 확인" 절차대로, 각 테스트가 실제로 실행됐는지 `=== RUN` 줄로 센다(0개 매칭 `ok` 통과 방지).

```
$ go test -count=1 ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestI18nKeySetParity|TestRouteForSectionTable|TestExcludedSectionsAllRejected|TestScopeContract|TestFeedbackPanelRendered|TestFeedbackAutoSubmitFieldRegistered|TestFeedbackSectionSeamWritable|TestFeedbackAutoSubmitI18nKeysInAllLocales|TestFeedbackPanelFieldsWired' -v
=== RUN   TestFeedbackAutoSubmitFieldRegistered
=== RUN   TestFeedbackSectionSeamWritable
=== RUN   TestSchemaCurrentValuesReadsAllSections
=== RUN   TestRouteForSectionTable
=== RUN   TestExcludedSectionsAllRejected
--- PASS: TestRouteForSectionTable (0.00s)
--- PASS: TestExcludedSectionsAllRejected (0.00s)
--- PASS: TestFeedbackAutoSubmitFieldRegistered (0.00s)
--- PASS: TestFeedbackSectionSeamWritable (0.00s)
--- PASS: TestSchemaCurrentValuesReadsAllSections (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.219s
=== RUN   TestFeedbackPanelRendered
--- PASS: TestFeedbackPanelRendered (0.01s)
=== RUN   TestFeedbackAutoSubmitI18nKeysInAllLocales
--- PASS: TestFeedbackAutoSubmitI18nKeysInAllLocales (0.00s)
=== RUN   TestFeedbackPanelFieldsWired
--- PASS: TestFeedbackPanelFieldsWired (0.00s)
=== RUN   TestI18nKeySetParity
--- PASS: TestI18nKeySetParity (0.05s)
=== RUN   TestScopeContractEditableSections
=== RUN   TestScopeContractExclusions
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.455s
```

### 2.3 전 패키지 회귀 + vet

```
$ go test -count=1 ./internal/settings/... ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/settings	1.018s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.347s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.692s
ok  	github.com/modu-ai/moai-adk/internal/web	2.899s

$ go vet ./internal/settings/... ./internal/web/...               → exit 0 (무출력)
$ GOOS=windows go vet ./internal/settings/... ./internal/web/...  → exit 0 (무출력)
```

### 2.4 4로케일 i18n 실측

```
$ grep -c '"f.feedback.auto_submit.title"' internal/web/assets/i18n.js   → 4
$ grep -c '"f.feedback.auto_submit.desc"'  internal/web/assets/i18n.js   → 4
```

`i18nKeyInAllLocales` 헬퍼의 판정 기준(로케일 블록 4개 각각에 `"<key>":` 1회 = 총 4회 이상)과 같은 수치이며, `TestFeedbackAutoSubmitI18nKeysInAllLocales` 와 스키마 전 필드를 훑는 기존 `TestI18nKeySetParity` 가 함께 이를 고정한다.

## 3. Baseline-attribution (baseline 귀속)

- 트리: `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`
- 착수 HEAD: `38705eb85` (M1~M6 착지 상태), base: `3210da7d3`
- 위 §2 의 모든 출력은 **이 실행, 이 트리** 에서 관측했다. RED 출력은 구현 편집 이전 상태, GREEN·회귀·vet 출력은 구현 편집 이후 상태이며, 캐시를 배제하기 위해 회귀와 선택자 실행에 `-count=1` 을 걸었다.
- 형제 SPEC 병합 확인도 실측 기반이다: RED 의 `RouteForSection("feedback") = 0` 이 곧 "`SPEC-TODO-ENABLE-FLAG-001` 이 아직 `feedback` 을 `ExcludedSections()` 에서 옮기지 않았다"는 관측이므로, 중복 제거는 발생하지 않았다.

## 4. Gaps (미검증 — 명시적으로 관측하지 않은 것)

1. **AC-F-023 의 템플릿 절반은 검증하지 않았다.** 템플릿 미러(`feedback.yaml` 의 `auto_submit: false`), `shipped_key_inventory.yaml` 항목, `make build`, 중립성 grep — 전부 M8 소관이다. `make build` 는 이번에 **실행하지 않았다**.
2. **`golangci-lint run` 을 실행하지 않았다.** 카드가 지정한 로컬 검증은 두 패키지의 테스트와 vet 2종이었고, 린트는 plan.md M9 검증 스윕에 배정돼 있다.
3. **브라우저 수동 확인을 하지 않았다.** 토글이 실제로 눌리고 저장되는지는 acceptance.md §D.5 의 머지 이후 수동 항목이다. 이번 관측은 서버가 만들어 내는 HTML 문자열까지다.
4. **`.moai/config/sections/feedback.yaml`(로컬 사본)에 `auto_submit` 키를 추가하지 않았다.** M8 소관이며, 키가 없어도 seam 이 upsert 하므로 콘솔 저장은 동작한다 — 다만 저장 전 콘솔이 읽는 초기 표시값은 빈 값이다(아래 잔여 위험 3).
5. **알려진 선행 실패 `TestConstitutionCrossReference`(`internal/cli/agentlint`)는 재확인하지 않았다.** M7 이 건드린 두 패키지 밖이고 base 에서도 붉다(원인: 커밋 `243eb07ef` 가 `moai-constitution.md` 에서 `agent-authoring.md` 인용 제거). 이번 결과에 섞이지 않았다.
6. **`internal/cli` · `internal/config` 등 다른 소비자 패키지는 돌리지 않았다.** 로컬 검증은 패키지 스코프로만 수행하라는 규율(CLAUDE.local.md §4)에 따랐고, 전 패키지 판정은 CI 몫이다.

## 5. Residual-risk (잔여 위험)

1. **고정 테스트 6곳 중 하나를 놓쳤을 가능성.** 카드는 2곳을 지목했고 실측으로 6곳을 찾았지만, 이 목록은 "두 패키지의 테스트를 돌려 붉어진 지점"에서 역산한 것이다. 다른 패키지가 `ExcludedSections()` 나 탭 목록을 복사해 고정하고 있다면 CI 에서만 드러난다. 완화: CI 전 패키지 실행이 그 지점을 잡는다.
2. **형제 SPEC 과의 병합 충돌.** `i18n.js` 와 `schema_sections.go` 는 `SPEC-TODO-ENABLE-FLAG-001` 과 공유한다. 신규 항목만 덧붙였으므로 충돌은 인접 줄 수준이지만, 두 카드가 같은 앵커(`f.feedback.repository.*` 직후 / feedback 섹션 필드 목록) 근처를 건드리면 수동 해소가 필요할 수 있다.
3. **`sec.feedback.desc` 4줄이 좁아진 채로 남아 있다.** 값이 "피드백 워크플로우 대상 저장소"인데 패널에는 이제 필드가 둘이다. 카드의 [HARD](공유 파일에 신규 항목만 추가)를 지키느라 기존 줄을 고치지 않았다. 사용자에게는 설명이 부정확해 보일 수 있으며, 후속 카드나 M8 에서 4로케일 동시 수정으로 처리하는 편이 안전하다.
4. **재개방이 위조 POST 표면을 다시 연다.** `feedback` 이 `RouteSeam` 이 되면서 `feedback.repository` 와 `feedback.auto_submit` 이 seam 쓰기 대상이 됐다. 이는 D5 가 의도한 결과이자 `lens-web-todo.md` §A.3 의 하드 에러를 해소하는 방향이지만, 동의 토글이 웹에서 켜질 수 있다는 뜻이기도 하다 — 전송 자체의 fail-closed 3조항(M6 스킬 본문)이 여전히 유일한 방어선이다.
5. **`auto_submit` 을 읽는 Go 소비자는 없다.** 키는 스킬 본문이 소비하며(`design.md` §8 이 정직하게 기록), 콘솔은 값을 기록만 한다. 콘솔에서 켠 값이 실제 전송 동작으로 이어지는지는 스킬 실행 경로에서만 관측되고, 이번 마일스톤은 그것을 관측하지 않았다.
