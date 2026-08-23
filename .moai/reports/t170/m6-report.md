# M6 보고 — 스킬 본문 + 마법사 질문 (SPEC-FEEDBACK-AUTO-SUBMIT-001)

트리: 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, 착수 HEAD `d2063308b`, base `3210da7d3`.

## §1 Claim (주장)

1. AC-F-019 가 M6 에 넘긴 부채 2건(②의 앵커 부재, ③의 한국어 리터럴)을 **구현 판정 이전에** 재작성했고, 재작성 토큰 4종 전부 base 에서 소스/미러 양쪽 0 을 실측했다.
2. 스킬 본문 두 사본에 확인 게이트 절을 넣었다 — 스크러버 경유(제목 포함) · verbatim 규칙의 명시적 예외 · fail-closed 3문장 · 3옵션 게이트(라벨/요약은 `conversation_language`) · 제출 실패 큐잉(초안 경로와의 D4 분기) · 재전송 직전 재스크럽.
3. 마법사 질문 `feedback_auto_submit` 을 **살아 있는 경로**에만 배선했고, 답이 실제로 `feedback.yaml` 에 도달함을 로더로 되읽어 관측했다.
4. AC-F-002 / F-019 / F-020 / F-021 / F-022 전부 PASS.
5. 두 사본 드리프트는 기존 1줄 그대로이며 확대되지 않았다.

## §2 Evidence (증거)

### 부채 재작성의 base 실측 (판정보다 먼저)

```
$ git show 3210da7d3:.claude/skills/moai/workflows/feedback.md > base_src.md
$ git show 3210da7d3:internal/template/templates/.claude/skills/moai/workflows/feedback.md > base_tpl.md
$ grep -c -- '--title' base_src.md base_tpl.md                          → 0 / 0   (구 ②)
$ grep -cE 'moai feedback scrub[^\n]*--title' base_src.md base_tpl.md   → 0 / 0   (신 ②, 채택)
$ grep -c 'MUST NOT submit' base_src.md base_tpl.md                     → 0 / 0   (신 ③-a, 채택)
$ grep -c '60 seconds' base_src.md base_tpl.md                          → 0 / 0   (신 ③-b, 채택)
```

구 ② 는 base 0 이라는 채택 조건은 만족하지만 **관측력이 없다** — 파일 어디의 `--title` 이든 잡으므로 본문만 스크럽하고 `gh issue create ... --title` 만 쓰는 구현이 통과한다. 동일 줄 앵커가 그 구멍을 닫는다. 구 ③ 의 보조 검사는 한국어 리터럴로 영어 전용 표면을 겨냥해 **통과시키려면 템플릿 언어 규칙을 어겨야 하는** 자기모순이었고, 형제 SPEC `SPEC-TODO-ENABLE-FLAG-001`(acceptance.md `N1`)의 선례대로 영어 토큰 + 건수 하한으로 바꿨다.

### AC-F-002 — 편집 전 FAIL / 편집 후 PASS

```
편집 전(base):  auto_submit 0건. gh issue create :84. 앞선 AskUserQuestion :52 는 필드 수집 라운드.
편집 후:        auto_submit 조건부 AskUserQuestion :108  <  첫 gh issue create :143
```

### AC-F-019 — 사본별 전수 (7관측)

| 관측 | 토큰 | SRC | TPL | 하한 |
|---|---|---|---|---|
| ① | `moai feedback scrub` | 1 | 1 | 1 |
| ② | `moai feedback scrub[^\n]*--title` | 1 | 1 | 1 |
| ③ | `verdict` | 3 | 3 | 1 |
| ③-a | `MUST NOT submit` | 3 | 3 | 3 |
| ③-b | `60 seconds` | 1 | 1 | 1 |
| ④ | `queue.json` | 1 | 1 | 1 |
| ⑤ | `label`↔`conversation_language` 동일 줄 | 1 | 1 | 1 |

verbatim 예외 문장은 ① 과 같은 절(`Step 1`) 안에 있다 — 두 사본 모두 `82 / 87 / 94`.

```
$ diff <소스> <템플릿 미러>
245d244
< Last Updated: 2026-02-07
$ grep -rnE 'SPEC-[A-Z]|REQ-[A-Z0-9]|AC-[A-Z]|CLAUDE\.local' <템플릿 미러>   → 무출력
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Leak|Neutral|Internal'  → ok
```

### RED (구현 이전, verbatim)

```
internal/cli/wizard/feedback_auto_submit_test.go:38:7: r.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
… (5건)
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard [build failed]

internal/core/project/initializer_feedback_test.go:34:50: undefined: defs.FeedbackYAML
internal/core/project/initializer_feedback_test.go:42:3: unknown field FeedbackAutoSubmit in struct literal of type InitOptions
… (5건)
FAIL	github.com/modu-ai/moai-adk/internal/core/project [build failed]
```

### GREEN + 회귀

```
$ go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmitQuestion|TestSaveBoolAnswerFeedbackAutoSubmit|TestFeedbackAutoSubmitTranslationsExist|TestQuestionOrder|TestReconfigureQuestions|TestWizardQuestionTranslationCompleteness' -v
  6개 전부 --- PASS, ok … 0.471s
$ go test ./internal/core/project/ -run 'TestWritePhase1Configs(PersistsFeedbackAutoSubmit|SkipsFeedbackWhenUnanswered|FeedbackNoFile)' -v
  3개 전부 --- PASS, ok … 0.418s
$ go test -timeout 900s ./internal/cli/wizard/... ./internal/core/project/... ./internal/defs/...  → 전부 ok
$ go test -timeout 900s ./internal/cli/...        → ok, 단 TestConstitutionCrossReference 1건 FAIL(선재)
$ go test -timeout 900s ./internal/template/...   → 전부 ok
$ go build ./...                                  → exit 0
$ go vet / GOOS=windows go vet (cli, core/project, defs)  → 둘 다 exit 0
$ golangci-lint run --timeout=3m (같은 3패키지)    → 0 issues.
$ make build                                      → exit 0
```

## §3 Baseline-attribution (baseline 귀속)

- 모든 수치는 이번 회차, 이 트리(`WT-auto-feedback`, 착수 HEAD `d2063308b`)에서 실행한 결과다. 다른 마일스톤·다른 시점의 측정을 옮겨오지 않았다.
- 부재 baseline: 위 4개 토큰의 `git show 3210da7d3:<path> | grep -c` 실측(전부 0/0). AC-F-002 의 편집 전 grep 출력도 같은 base 사본 대상.
- 선재 FAIL 귀속: `TestConstitutionCrossReference` 의 단언 대상은 `.claude/rules/moai/core/moai-constitution.md` 이며, 해당 인용 줄은 커밋 `243eb07ef`(별도 카드, 이미 `release/v3.1.3` 착지)가 제거했다. base 에서도 붉고, M5 보고서 §2 가 같은 FAIL 을 이미 기록했다. M6 diff 는 이 규칙 파일을 건드리지 않는다.
- 개수 고정 테스트 4건의 갱신 전 실패 출력(16→17, 10→11, 15→16, 멤버십)은 이번 회차에 직접 관측한 것이며, 갱신 후 같은 명령이 초록임을 재측정했다.

## §4 Gaps (미검증)

1. **대화형 `moai init` 을 실제로 돌려보지 않았다.** 질문이 자기 로케일로 뜨는지는 번역 존재(단위 테스트)로만 관측했다 — acceptance.md §D.5 의 수동 확인 항목 그대로 남는다.
2. **스킬 본문의 조항이 런타임에 지켜지는지는 관측 불가.** 산문은 기계 검증 대상이 아니다. grep 은 조항이 **쓰여 있음**만 본다.
3. **실제 피드백 1건 제출 왕복(게이트 표시 → 제출 → 큐잉)은 하지 않았다.** M6 은 그 경로를 서술했을 뿐 실행하지 않았다.
4. **`make build` 후 `catalog.yaml` 은 git 상 변경이 없다** — 재생성은 실행됐고 템플릿 테스트도 초록이지만, 스킬 워크플로 파일 편집이 카탈로그 해시에 반영되는지는 별도로 확인하지 않았다.
5. **`gofmt -l` 이 무관한 선재 파일 4개를 계속 보고한다**(`wizard/mcp_audit_test.go`, `core/project/initializer_audit_test.go`, `initializer_audit_wiring_test.go`, `initializer_workflow_toggles_test.go`). M6 편집 파일은 전부 깨끗하다. 선재 부채로 남긴다.
6. **전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다** — CLAUDE.local.md §4 규율. 전 패키지 판정은 CI 몫.
7. **Windows 실동작 미관측.** `GOOS=windows go vet` 은 컴파일만 증명한다.
8. **M7·M8·M9 범위**(웹 노출, 템플릿 `auto_submit` 키·인벤토리·docs-site 4로케일, 최종 스윕)는 M6 밖이다. 배포되는 `feedback.yaml` 에는 아직 `auto_submit` 키가 없다.

## §5 Residual-risk (잔여 위험)

1. **스크러버 도입은 규약 강제이지 샌드박스가 아니다.** `moai feedback` 을 거치지 않고 직접 이슈를 여는 경로는 그대로 열려 있다. M6 이 만든 것은 산문 조항이며, 조항을 어긴 실행을 막는 장치는 없다.
2. **게이트 3옵션의 `(권장)` 기본이 "제출하지 않음"이라 이탈률이 오를 수 있다.** 되돌리기 어려운 공개 게시 쪽으로 기본을 두지 않은 의도된 선택이지만, 사용자에게는 마찰로 보인다. `auto_submit: true` 가 그 탈출구다.
3. **마법사 질문 문구가 부정형이다**("확인 절차 없이 제출할까요?"). Confirm 의 "예"가 **게이트 해제**를 뜻하므로 오독 여지가 있다. 기본이 `false` 라 오독의 비용은 안전한 쪽(게이트 유지)으로 떨어진다.
4. **`feedback.auto_submit` 을 Go 코드가 읽지 않는다**(design.md §8). 스킬 본문이 설정을 직접 읽어 분기하므로, 설정과 산문 사이에 드리프트가 생겨도 컴파일러도 테스트도 잡지 않는다.
5. **형제 SPEC 과 공유 파일 9종이 겹친다.** M6 은 `questions.go` / `types.go` / `wizard.go` / `translations.go` / `init.go` / `initializer.go` / `initializer_expansion.go` 에 항목만 추가했고 기존 항목을 재배치·재서식하지 않았다. 다만 개수 고정 테스트 4건은 숫자가 바뀌었으므로, 형제가 나중에 착지하면 같은 4건을 다시 조정해야 한다.
