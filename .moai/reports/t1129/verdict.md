# t1129 — 위저드 빈 effort 라벨 현지화 + 런타임 기본값 단일 원천화

- 카드: t1129 (Tier S, 클래스 B, SPEC 없음)
- 브랜치: `WT-wizard-effort-labels` (base = 로컬 develop `5f264c381`)
- 코드 커밋: `2cf3a42db` (tree `51910ca13f7f6e360f4a1f457deae3be21bca26d`)
- 출처: t1114 sync-audit 후속 채무 F1/F2/F3

## Claim

1. F1 — ko/ja/zh 위저드의 빈 effort 옵션이 영어 스키마 리터럴 대신 각 언어 문구를 렌더하며, 각 문구는 두 폴백(모델 정책 → Claude Code 기본값)을 모두 말한다. en 라벨은 스키마 상수와 바이트 동일하다.
2. F2 — `internal/web/schema_label_test.go` 의 낡은 `// "(runtime default)"` 주석을 현재 의미로 고쳤다. 같은 블록의 나머지 두 주석(`"(project default)"`)은 `schema.go` 의 `emptyLabelProjectDefault` 와 일치해 그대로 두었다.
3. F3 — "Opus 5.5에서 medium" 사실은 `settings.RuntimeDefaultEffort`/`RuntimeDefaultEffortModel` 두 상수에만 정의되고, 스키마 라벨은 그 상수로 조립된다. i18n.js 의 opt.runtime_default 4개와 위저드 4개 로케일 라벨은 가드 테스트 두 개가 상수에 묶는다. 두 가드 모두 변이에 대해 실패함을 확인했다.

F1 메커니즘: 기존 `schemaOptionBridge`(i18n 키 → `profileSetupText` 접근자) 재사용. 빈 옵션 라벨을 필드의 `EmptyLabelKey`(`opt.runtime_default`, 웹 콘솔이 이미 현지화하는 키)로 브리지에서 찾고, 해당 로케일 값이 비어 있으면 스키마 `EmptyLabel` 로 폴백한다. en 값은 비워 두어 스키마 상수가 그대로 쓰인다. 새 필드는 `EffortLevelEmpty` 하나이며, 새 조회 경로를 만들지 않고 옵션 라벨이 이미 쓰는 브리지에 키 한 줄을 더했다.

## Evidence

모든 로그는 `.moai/reports/t1129/` 아래, 워크트리 트리에서 실행했다.

RED (컴파일 단계 — 상수 도입 전 가드 테스트), `red.log`:
```
$ go test -count=1 ./internal/cli -run 'TestEffortEmptyLabel'
internal/cli/profile_setup_effort_empty_label_test.go:54:42: undefined: settings.RuntimeDefaultEffort
internal/cli/profile_setup_effort_empty_label_test.go:54:73: undefined: settings.RuntimeDefaultEffortModel
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
exit=1
```

RED (동작 — 상수 도입 후, 수정 전), `red-behavioral.log`:
```
$ go test -count=1 ./internal/cli -run 'TestEffortEmptyLabel'
--- FAIL: TestEffortEmptyLabelLocalized (0.00s)
    profile_setup_effort_empty_label_test.go:32: ko: empty effort label is the untranslated English schema literal "(model policy, else Claude Code default: medium on Opus 5.5)"
    profile_setup_effort_empty_label_test.go:32: ja: empty effort label is the untranslated English schema literal "(model policy, else Claude Code default: medium on Opus 5.5)"
    profile_setup_effort_empty_label_test.go:32: zh: empty effort label is the untranslated English schema literal "(model policy, else Claude Code default: medium on Opus 5.5)"
FAIL
exit=1
```

골든 재생성(손편집 없음), `golden-update.log`:
```
$ go test -count=1 ./internal/cli -run 'TestProfileWizardGolden_LocaleFrames' -update-golden
ok  	github.com/modu-ai/moai-adk/internal/cli	2.618s
exit=0
```
골든 diff 는 로케일마다 빈 effort 행 1줄씩(3 files, 3 insertions, 3 deletions).

GREEN, `green-settings-web.log`:
```
$ go test -count=1 ./internal/settings/... ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/settings	0.909s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.318s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.326s
ok  	github.com/modu-ai/moai-adk/internal/web	46.370s
exit=0
```

GREEN, `green-cli.log`:
```
$ go test -count=1 ./internal/cli -run 'Profile|TUI|Wizard|Golden|Effort|Schema|Bridge|GetProfileText'
ok  	github.com/modu-ai/moai-adk/internal/cli	25.874s
exit=0
```

GREEN, `green-cli-wizard.log`:
```
$ go test -count=1 ./internal/cli/wizard/...
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	4.097s
exit=0
```

GREEN 이름 지정(-v), `green-cli-named.log` / `green-web-named.log`:
```
--- PASS: TestProfileWizardGolden_LocaleFrames (0.46s)
--- PASS: TestProfileWizardGolden_NoEnglishLeak (0.51s)
--- PASS: TestSchemaSelectOptions_DerivedFromSchema (0.00s)
--- PASS: TestSchemaSelectOptions_EmptyLabelFromSchema (0.00s)
--- PASS: TestSchemaSelectOptions_ModelValuesSurviveNormalizer (0.00s)
--- PASS: TestSchemaSelectOptions_Localized (0.00s)
--- PASS: TestProfileTextFields_Referenced (0.07s)
--- PASS: TestEffortEmptyLabelLocalized (0.00s)
--- PASS: TestTUIEmptyLabelsSchemaSourced (0.00s)
--- PASS: TestEffortEmptyLabelCarriesRuntimeDefaultFact (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.772s
exit=0
--- PASS: TestEffortOptRecommendationLabels (0.00s)
--- PASS: TestSchemaEmptyLabelParity (0.01s)
--- PASS: TestEffortEmptyLabelNamesBothFallbacks (0.00s)
--- PASS: TestRuntimeDefaultI18nCarriesEffortFact (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.462s
exit=0
```

vet / lint, `vet.log` / `lint.log`:
```
$ go vet ./internal/settings/... ./internal/web/... ./internal/cli/
exit=0
$ golangci-lint run ./internal/settings/... ./internal/web/... ./internal/cli/
0 issues.
exit=0
```

F3 가드 변이 검사, `mutation.log`:
```
68678e155b0003074cf3365e9e49ae27e3ca0ea69ebc4bb7d95bca8d81cf2b27  internal/web/assets/i18n.js
48d9a785d8afa9ee506f65f3f9188363eec748225ec7c29c086a5cfbf2d10fa3  internal/cli/profile_setup_translations.go
--- mutant 1: remove 'Opus 5.5' from ko opt.runtime_default in i18n.js
    "opt.runtime_default": "(moai profile setup의 모델 정책, 없으면 Claude Code 기본값: medium)",
--- FAIL: TestRuntimeDefaultI18nCarriesEffortFact (0.00s)
    schema_label_test.go:186: opt.runtime_default "(moai profile setup의 모델 정책, 없으면 Claude Code 기본값: medium)" does not carry "Opus 5.5"
mutant1 exit=1
git diff --stat i18n.js exit=0 (no stat line above = byte-identical to HEAD)
68678e155b0003074cf3365e9e49ae27e3ca0ea69ebc4bb7d95bca8d81cf2b27  internal/web/assets/i18n.js
--- mutant 2: remove 'Opus 5.5' from zh wizard EffortLevelEmpty
467:		EffortLevelEmpty:      "(模型策略，未设置则用 Claude Code 默认值: medium)",
--- FAIL: TestEffortEmptyLabelCarriesRuntimeDefaultFact (0.01s)
    profile_setup_effort_empty_label_test.go:56: zh: empty effort label "(模型策略，未设置则用 Claude Code 默认值: medium)" does not carry "Opus 5.5"
mutant2 exit=1
48d9a785d8afa9ee506f65f3f9188363eec748225ec7c29c086a5cfbf2d10fa3  internal/cli/profile_setup_translations.go
```
두 파일 모두 복원 후 SHA-256 이 변이 전과 같다.

템플릿 미러 점검, `mirror-check.log`:
```
$ find internal/template/templates -name 'i18n.js' -o -name 'schema.go' -o -name 'profile_setup*' -o -name 'schema_bridge.go' -o -name 'schema_label_test.go'
find exit=0 matches=       0
```

## Baseline-attribution

- 모든 측정은 이 실행에서, 워크트리 `.claude/worktrees/t1129` 트리에 대해 수행했다. base 는 로컬 develop `5f264c381`.
- GREEN·vet·lint 는 커밋 `2cf3a42db` 에 들어간 바이트와 같은 워킹 트리에서 돌았다. 측정 뒤 편집은 변이 검사뿐이고, 두 파일 모두 변이 전 SHA-256 으로 복원됨을 위 로그가 보인다. 커밋 직후 `git status --short` 는 비어 있었다.
- RED 는 수정 전 트리(테스트 파일 추가 + 상수 도입만 된 상태)에서 측정했다.

## Gaps

- `internal/cli` 패키지 전체 스위트는 돌리지 않았다(10분 타임아웃 위험). 실행한 것은 `-run 'Profile|TUI|Wizard|Golden|Effort|Schema|Bridge|GetProfileText'` 와 `./internal/cli/wizard/...` 뿐이다. 이 패턴 밖에서 `schemaSelectOptions` 나 `schemaOptionBridge` 를 쓰는 테스트가 있다면 측정되지 않았다. 전 패키지 판정은 develop push 뒤 CI 몫이다.
- darwin 외 플랫폼(linux/windows) 빌드·테스트는 측정하지 않았다.
- 실제 터미널에서 위저드를 띄워 본 시각 확인은 하지 않았다. 골든(ptycaptest 프레임)이 대신한다.
- ko/ja/zh 문구의 자연스러움은 원어민 검수를 거치지 않았다. i18n.js 의 기존 opt.runtime_default 문구를 줄인 형태다.

## Residual-risk

- 가드는 두 상수 값이 문자열에 "들어 있는지"만 본다. 상수가 바뀌어도 옛 값과 새 값이 문자열에 함께 남으면 통과한다. 모델명 교체 시 문장 전체를 사람이 다시 읽어야 한다.
- `schemaOptionBridge` 에 `f.*.opt.*` 형태가 아닌 키(`opt.runtime_default`)가 처음 들어갔다. 브리지 키를 i18n 옵션 키 전수와 대조하는 향후 테스트가 생기면 이 키를 예외로 다뤄야 할 수 있다.
- 다른 빈 옵션 라벨(`(project default)`, `(unset)`)은 여전히 모든 로케일에서 영어로 렌더된다. 이 카드 범위 밖이며, 같은 메커니즘(브리지에 키 추가 + 로케일 값)으로 확장 가능하다.

## Lane independent recheck (after agent return, HEAD 94604338f)

- `go test -count=1 ./internal/settings/... ./internal/web/...` -> exit 0 (settings 0.402s, agentfm, yamlpatch, web 24.264s ok) — lane-recheck-settings-web.log
- `go test -count=1 -run 'Profile|TUI|Wizard|Golden|Effort|Schema|Bridge|GetProfileText' ./internal/cli` -> `ok github.com/modu-ai/moai-adk/internal/cli 14.728s`
- `golangci-lint run ./internal/settings/... ./internal/web/... ./internal/cli/` -> `0 issues.` — lane-recheck-lint.log
- Diff read by the lane: 9 code/golden files, en label byte-identical (constant concatenation), bridge fallback keeps untranslated empty keys on the schema label.

## Sync audit (sync-auditor, lens --i18n, HEAD e0edeb3c0)

- Verdict: **PASS-WITH-DEBT** — Functionality 95 / Security 98 / Craft 90 / Consistency 90 (harmonic ≈ 93.1); 0 blocking.
- Auditor-run: `go test -count=1 ./internal/settings/... ./internal/web/...` -> `ok .../settings 0.827s`, `ok .../web 25.927s`; `go test -count=1 -run 'Effort|Golden|TUI|Schema|Bridge|English|Leak' ./internal/cli -v` -> `ok .../internal/cli 2.941s` (TestEffortEmptyLabelLocalized, TestEffortEmptyLabelCarriesRuntimeDefaultFact, TestProfileWizardGolden_NoEnglishLeak, TestTUIEmptyLabelsSchemaSourced PASS); gofmt/vet clean; en label byte-identical; bridge key `opt.runtime_default` used only by effort_level.
- Debt:
  - D1 [Medium, pre-existing, out of scope] `(project default)` empty label still English on ko/ja/zh wizard (model, development_mode).
  - D2 [Medium, pre-existing, out of scope] model_policy wizard labels English on all locales; `NoEnglishLeak` only checks strings differing from en, so locale-invariant English is structurally invisible.
  - D3 [Low] `internal/web/console_ux_fix_test.go:252` hard-coded `Opus 5\.5` + comment :233 — **fixed in-card** (see below).
  - D4 [Low] guards use `strings.Contains`; a model rename to a prefix of the old name (e.g. "Opus 5") would pass on stale strings.
  - D5 [Info] same fact outside the guard: `.claude/rules/moai/development/model-policy.md:19` (+ template mirror), `docs-site/content/{en,ko,ja,zh}/multi-llm/model-policy.md`, `README{,.ko,.ja,.zh}.md:246`.

## D3 fix (lane)

- `internal/web/console_ux_fix_test.go`: regex now uses `regexp.QuoteMeta(settings.RuntimeDefaultEffortModel)`; comment points at the constant.
- `gofmt -l internal/web/` -> empty; `go vet ./internal/web/` -> exit 0; `go test -count=1 -run TestEffortOptRecommendationLabels -v ./internal/web/` -> `--- PASS` / `ok .../internal/web 0.707s`; `golangci-lint run ./internal/web/...` -> `0 issues.`
