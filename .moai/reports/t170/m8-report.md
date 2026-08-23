# M8 검증 보고 — SPEC-FEEDBACK-AUTO-SUBMIT-001

카드 t170 · 브랜치 `WT-auto-feedback` · 워크트리 `.claude/worktrees/t170` · 착수 HEAD `23c5c18fa` · base `3210da7d3` · M8 커밋 = 이 보고서를 담고 있는 커밋(자기 참조라 SHA 를 본문에 고정하지 않는다)

---

## 1. Claim (주장)

| # | 주장 |
|---|---|
| C1 | `feedback.auto_submit: false` 가 배포 템플릿 섹션 YAML 에 중립 주석과 함께 들어갔고, 템플릿 밖 사본 2개가 바이트 동일하게 미러됐다 |
| C2 | `feedback.auto_submit` 항목이 shipped-key 인벤토리에 등록됐고, `evidence` 는 design.md §8 대로 "스킬 본문이 소비, Go 호출자 없음"을 적었다 |
| C3 | `internal/settings/schema_sections_test.go` per-key 기대값 맵에 `"feedback.auto_submit": "false"` 1줄이 추가됐다 |
| C4 | 두 anti-rot 가드가 통과한다 |
| C5 | `make build` 가 exit 0 이고, 그 결과로 커밋해야 할 `catalog.yaml` 변경은 발생하지 않았다 |
| C6 | 템플릿 중립성이 유지된다 (SPEC ID · REQ 토큰 0건, strict 리크 테스트 통과) |
| C7 | docs-site 4로케일에 절 1개씩 동일하게 들어갔다 |
| C8 | 건드린 4개 패키지가 회귀 없이 통과하고, `go vet` · `GOOS=windows go vet` 둘 다 exit 0 이다 |
| C9 | 형제 SPEC `SPEC-TODO-ENABLE-FLAG-001` 과의 공유 파일에서 기존 항목을 재배치·재서식하지 않았다 (AP-10) |
| C10 | M7 이 인계한 `sec.feedback.desc` 4로케일 문구를 두 필드를 포괄하도록 넓혔다 |

**AC-F-023 템플릿 절반 판정: PASS.** M7 이 웹 절반을, M8 이 템플릿 절반을 졌으므로 AC-F-023 양쪽이 모두 관측됐다.

---

## 2. Evidence (증거)

### C1 — 템플릿 + 사본 2개

```
$ cat internal/template/templates/.moai/config/sections/feedback.yaml
feedback:
    # Target repository for the /moai feedback workflow (GitHub owner/repo slug).
    # Default: the remote MoAI-ADK tool repository — bug reports and feature
    # requests about the MoAI-ADK tool itself, NOT the user's own project.
    # Fork maintainers override this to redirect feedback to their own fork.
    repository: modu-ai/moai-adk

    # Whether the /moai feedback workflow may create the issue without asking
    # first. Default false: before anything leaves the machine the workflow
    # shows the masked title, the masked body in full, and a summary of what
    # was masked, then waits for an explicit choice. Set this to true only if
    # that per-submission confirmation is not wanted.
    auto_submit: false
```

세 사본의 변경량 (diff --stat, base 대비):

```
 .moai/config/sections/feedback.yaml                             | 7 +++++++
 internal/settings/testdata/sections/feedback.yaml               | 7 +++++++
 internal/template/templates/.moai/config/sections/feedback.yaml | 7 +++++++
```

세 사본은 `cp` 로 미러했으므로 바이트 동일하다.

### C2 — 인벤토리 항목

```
$ sed -n '362,364p' internal/config/testdata/shipped_key_inventory.yaml
- path: "feedback.auto_submit"
  class: R
  evidence: "consumed by the skill body (.claude/skills/moai/workflows/feedback.md); no Go production caller"
```

`class: R` 인 이유는 design.md §8 — 이 키는 Go 코드가 읽지 않고 스킬 본문이 설정 파일을 읽어 분기하며, `FeedbackAutoSubmit()` 접근자는 현재 프로덕션 호출자가 없다. `evidence` 필드는 가드가 파싱만 하고 값은 검사하지 않으므로(`loadAllowlistWithCount` 는 `Path` 만 읽는다) `none` 으로 뭉갤 수 있었지만, 그러면 나중에 이 키를 보는 사람이 "접근자가 있으니 누군가 읽겠지"로 오독한다.

### C3 — per-key 맵

```
@@ -460,6 +460,7 @@ func TestSchemaCurrentValuesReadsAllSections(t *testing.T) {
 		"ralph.lint_as_instruction":                    "true",
 		"ralph.warn_as_instruction":                    "false",
 		"feedback.repository":                          "modu-ai/moai-adk",
+		"feedback.auto_submit":                         "false",
 		"observability.retention_days":                 "30",
```

`gofmt -w` 이후에도 diff 는 `1 insertion(+)` 단일 줄이다 — 기존 줄의 정렬 열이 움직이지 않았다.

### C4 — 두 anti-rot 가드

plan.md §B 4번이 지목한 가드 2건 중 두 번째(`schema_label_test.go:96`)는 **경로가 다르다**: `internal/settings/` 가 아니라 `internal/web/schema_label_test.go` 이고, 그 안에 `TestSchemaLabel` 이라는 이름의 테스트는 없다.

```
$ grep -n '^func Test' internal/web/schema_label_test.go
16:func TestSchemaEmptyLabelParity(t *testing.T) {
74:func TestI18nKeySetParity(t *testing.T) {
133:func TestI18nSegmentKeysRemovedFromWebDictionary(t *testing.T) {
```

`-run 'TestSchemaLabel'` 로 돌렸다면 0개 매칭으로 조용히 `ok` 가 찍혔을 것이므로, 실재하는 이름으로 돌렸다.

```
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -v
=== RUN   TestShippedConfigKeysHaveReaders/non_vacuous_inventory
    shipped_key_reader_test.go:132: non-vacuity: 896 shipped keys, 959 inventory entries, 329 struct fields
=== RUN   TestShippedConfigKeysHaveReaders/collision_resolution
=== RUN   TestShippedConfigKeysHaveReaders/accessor_indirection
=== RUN   TestShippedConfigKeysHaveReaders/unbound_classification
--- PASS: TestShippedConfigKeysHaveReaders (0.68s)
    --- PASS: TestShippedConfigKeysHaveReaders/non_vacuous_inventory (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/collision_resolution (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/accessor_indirection (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/unbound_classification (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	1.057s

$ go test ./internal/web/ -run 'TestSchemaEmptyLabelParity|TestI18nKeySetParity|TestI18nSegmentKeysRemovedFromWebDictionary|TestScopeContract' -v
--- PASS: TestSchemaEmptyLabelParity (0.01s)
--- PASS: TestI18nKeySetParity (0.05s)
--- PASS: TestI18nSegmentKeysRemovedFromWebDictionary (0.00s)
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.499s
```

키 수 895 → 896, 인벤토리 958 → 959 — 이번 마일스톤이 넣은 키 1개와 항목 1개다.

### C5 — make build

```
$ make build
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "-s -w -X ...version.Version=v3.1.2 -X ...version.Commit=23c5c18fa -X ...version.Date=2026-08-23T10:39:24Z" -o bin/moai ./cmd/moai
EXIT=0
```

직후 워킹 트리 상태 (포슬린):

```
 M .moai/config/sections/feedback.yaml
 M docs-site/content/en/utility-commands/moai-feedback.md
 M docs-site/content/ja/utility-commands/moai-feedback.md
 M docs-site/content/ko/utility-commands/moai-feedback.md
 M docs-site/content/zh/utility-commands/moai-feedback.md
 M internal/config/testdata/shipped_key_inventory.yaml
 M internal/settings/schema_sections_test.go
 M internal/settings/testdata/sections/feedback.yaml
 M internal/template/templates/.moai/config/sections/feedback.yaml
 M internal/web/assets/i18n.js
```

`internal/template/catalog.yaml` 은 목록에 없다 — 카탈로그는 에이전트·스킬만 해시하고 config 섹션 YAML 은 대상이 아니라서 이번 편집으로 해시가 움직이지 않았다. (움직였다면 같은 커밋에 실어야 CI parity 가 통과한다.)

### C6 — 중립성

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	21.247s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	(cached)
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]

$ grep -rn 'SPEC-FEEDBACK-AUTO-SUBMIT\|REQ-' internal/template/templates/.moai/config/sections/feedback.yaml internal/template/templates/.claude/skills/moai/workflows/feedback.md
                                                       # 무출력, exit 1 (0건)

$ grep -rn 'SPEC-\|REQ-\|CLAUDE.local' internal/template/templates/.moai/config/sections/feedback.yaml
                                                       # 무출력, exit 1 (0건)
```

### C7 — docs-site 4로케일

```
$ grep -c '^### ' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
en:14  ko:14  zh:14  ja:14

$ grep -c '^## ' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
ja:10  zh:10  ko:10  en:10

$ grep -c 'auto_submit' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
en:2  ko:2  zh:2  ja:2
```

절 개수 13 → 14 로 네 로케일 동일, `##` 개수 무변경, `auto_submit` 출현 2회(산문 1 + YAML 예시 1)로 네 로케일 동일. 절 제목은 en `Confirming Before Submission` · ko `제출 전 확인` · ja `送信前の確認` · zh `提交前确认`. 본문 장식 이모지 없음, Mermaid 없음, YAML 예시 블록은 네 로케일에서 바이트 동일.

### C8 — 회귀 + 크로스 플랫폼

```
$ go test ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/config	1.309s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.545s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	(cached)
ok  	github.com/modu-ai/moai-adk/internal/template	22.277s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.564s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
ok  	github.com/modu-ai/moai-adk/internal/settings	0.995s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	(cached)
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	(cached)
ok  	github.com/modu-ai/moai-adk/internal/web	2.954s

$ go vet ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...
VET_EXIT=0                                             # 무출력

$ GOOS=windows go vet ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...
WINVET_EXIT=0                                          # 무출력
```

### C9 — 형제 SPEC 공유 파일

```
$ grep -n 'todo\.' internal/config/testdata/shipped_key_inventory.yaml
                                                       # 무출력 (0건)
```

`SPEC-TODO-ENABLE-FLAG-001` 은 아직 인벤토리에 항목을 넣지 않았다 — 중복 삽입은 발생하지 않았고, 세 공유 파일에는 신규 항목만 덧붙였다. `schema_sections_test.go` diff 가 `1 insertion(+)` 이고 `i18n.js` diff 가 로케일당 1줄 치환뿐인 것이 그 관측이다.

### C10 — sec.feedback.desc 확장

```
$ grep -n 'sec.feedback.desc' internal/web/assets/i18n.js
206:    "sec.feedback.desc": "Feedback workflow target repository and pre-submission confirmation.",
922:    "sec.feedback.desc": "피드백 워크플로우 대상 저장소와 제출 전 확인 여부.",
1543:    "sec.feedback.desc": "フィードバックワークフローの対象リポジトリと送信前の確認。",
2164:    "sec.feedback.desc": "反馈工作流目标仓库与提交前确认。",
```

M7 이 [HARD] "공유 파일에는 신규 항목만" 규율을 지켜 이 4줄을 손대지 않고 잔여 위험으로 남겼고, 카드가 M8 에 인계했다. 이 SPEC 자신의 변경(필드가 둘이 됨)이 기존 설명 1개를 부정확하게 만든 경우라 재배치·재서식이 아니다.

### AC-F-023 지정 5개 선택자 (§D.3 [HARD] 0개-실행 통과 방지)

```
$ go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestI18nKeySetParity|TestRouteForSectionTable|TestExcludedSectionsAllRejected|TestScopeContract' -v
=== RUN   TestSchemaCurrentValuesReadsAllSections
=== RUN   TestRouteForSectionTable
=== RUN   TestExcludedSectionsAllRejected
--- PASS: TestExcludedSectionsAllRejected (0.00s)
--- PASS: TestRouteForSectionTable (0.00s)
--- PASS: TestSchemaCurrentValuesReadsAllSections (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.351s
=== RUN   TestI18nKeySetParity
--- PASS: TestI18nKeySetParity (0.05s)
=== RUN   TestScopeContractEditableSections
=== RUN   TestScopeContractExclusions
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.597s
```

5개 선택자가 각각 `=== RUN` 을 찍었다 (`TestScopeContract` 는 접두사 매칭으로 2함수).

---

## 3. Baseline-attribution (baseline 귀속)

- 모든 명령은 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170` 에서, 브랜치 `WT-auto-feedback` 위에서 실행했다. 실행 시점 HEAD 는 `23c5c18fa`(M7 착지분), 커밋 후 HEAD 는 이 보고서를 담은 M8 커밋이다(자기 참조 — `git log -1` 로 확인).
- 시작 시 `pwd` 와 `git rev-parse --show-toplevel` 로 워크트리 루트를 확인했고, 이후 `cd` 를 쓰지 않았다 — 측정 범위가 조용히 움직이지 않았다.
- 인벤토리 카운트 895 → 896 / 958 → 959 는 이번 회차 실측치다: 편집 전 항목 수 `grep -c 'path:'` 가 958, 편집 후 가드 로그가 959 를 보고했다.
- docs 절 개수 13 → 14 는 편집 후 실측(14) + 삽입 절 1개로부터의 산술이다. 편집 **전** 13 을 직접 측정하지는 않았다 — 아래 미검증 2번.
- 스테이징 직전 `git status --short` 와 `git rev-parse --short HEAD` 를 다시 읽어 트리·HEAD 가 움직이지 않았음을 확인한 뒤 명시 pathspec 11개로 스테이징했다 (`git add -A` 미사용).

---

## 4. Gaps (미검증)

1. **`TestConstitutionCrossReference` 를 이번 회차에 재관측하지 않았다.** `internal/cli/agentlint` 패키지 소속이라 M8 이 건드린 4개 패키지 스코프 밖이고, 카드 지시가 "base 에서도 붉다, 고치지 말라"였다. 붉다는 사실은 M6·M7 기록과 카드 지시를 근거로 삼았을 뿐 이번에 직접 돌리지 않았다 — 인용이지 관측이 아니다.
2. **docs 절 개수의 편집 전 값(13)을 직접 측정하지 않았다.** 편집 후 14 만 실측했고 13 은 산술 역산이다. 네 로케일이 모두 14 로 동일하다는 점은 관측됐으므로 로케일 간 패리티 주장은 영향받지 않는다.
3. **Hugo 빌드를 돌리지 않았다.** docs-site 편집이 렌더 단계에서 경고 없이 통과하는지는 이번 회차에 검증하지 않았다. 추가한 것은 표준 markdown(`###` 제목 + 문단 + yaml 코드블록)뿐이고 shortcode·Mermaid·이모지는 없다.
4. **전체 스위트(`go test ./...`)를 돌리지 않았다.** CLAUDE.local.md §4 규율에 따라 4개 패키지 스코프로만 검증했다. 전 패키지 판정은 CI 몫이다.
5. **인벤토리 헤더의 `Total entries: 958` 을 고치지 않았다.** 항목이 959 가 됐으므로 헤더 숫자는 스테일하다. 고치지 않은 근거는 아래 잔여 위험 2번.
6. **`bin/moai` 를 `~/go/bin/` 에 재설치하지 않았다.** `make build` 는 워크트리의 `bin/moai` 만 갱신한다. 이 변경은 config 키 1개라 CLI 동작 검증이 불필요했으나, 설치된 바이너리로 스모크를 돌리지는 않았다.
7. **웹 콘솔을 실제로 띄워 새 문구를 눈으로 확인하지 않았다.** `sec.feedback.desc` 4로케일 확장은 파일 내용 grep 으로만 관측했다.
8. **`golangci-lint` 를 돌리지 않았다.** 카드가 지정한 검증 목록에 없었고, 이번 Go 변경은 테스트 파일 맵 1줄뿐이다.

---

## 5. Residual-risk (잔여 위험)

1. **형제 SPEC 과의 3개 공유 파일 충돌.** `SPEC-TODO-ENABLE-FLAG-001` 이 `shipped_key_inventory.yaml` · `schema_sections_test.go` · `i18n.js` 에 자기 항목을 넣으면 같은 영역에 인접 삽입이 발생한다. 알파벳 순 삽입(`feedback.*` vs `todo.*`)이라 인벤토리는 멀리 떨어져 있고 `schema_sections_test.go` 는 맵이라 순서 무관이지만, 머지 시 diff 컨텍스트가 겹칠 수 있다.
2. **인벤토리 헤더 카운트 스테일.** `Total entries: 958` 이 실제 959 와 어긋난다. 고치지 않은 이유는 두 가지다 — (a) 헤더는 `Code baseline: <sha>` 와 한 덩어리인 M1 시점 스냅샷 기록이고 어떤 테스트도 이 숫자를 읽지 않는다(가드는 항목을 세어 `minimumShippedKeys` 하한만 본다), (b) 형제 SPEC 도 항목을 1개 늘리므로 두 카드가 같은 줄을 고치면 충돌한다. 헤더를 현재 값으로 되살릴지는 후속 판단 사항이다.
3. **docs 문구가 스킬 본문 동작과 어긋날 여지.** 문서는 게이트가 "마스킹된 제목 + 본문 전문 + 가려낸 값 요약"을 보여준다고 적었는데, 이는 `.claude/skills/moai/workflows/feedback.md` Step 3 를 읽고 옮긴 것이다. 스킬 본문이 나중에 바뀌면 문서가 조용히 어긋난다 — 문서와 스킬 본문 사이에 기계적 패리티 검사는 없다.
4. **`evidence` 자유 서술의 내구성.** 가드가 `evidence` 값을 검사하지 않으므로, 나중에 접근자에 실제 호출자가 생겨도 이 문자열은 자동 갱신되지 않는다. `class: R` → `W` 재분류는 사람이 해야 한다.
5. **4로케일 문구의 뉘앙스 차이.** en/ko/ja/zh 문구는 같은 사실을 담되 각 언어의 자연스러운 표현으로 썼다 — 축자 번역이 아니다. 사실·수치·코드 블록은 동일하지만, 표현 차이를 로케일 간 불일치로 읽는 리뷰어가 있을 수 있다.
6. **docs 에 스크러버 서사를 새로 열지 않은 판단.** 이 페이지는 지금까지 마스킹을 한 번도 언급한 적이 없어, 확인 게이트가 무엇을 보여주는지 설명하는 최소한만 적고 스크러버 자체의 절은 만들지 않았다. 사용자가 마스킹 규칙 전체를 문서에서 찾으려 하면 이 페이지에는 없다.
