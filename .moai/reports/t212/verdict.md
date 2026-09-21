# t212 — report.format 옵션 설명 키가 `.opt.` 가드에 걸려 영어로 얼어붙던 문제

card: t212 · branch: `WT-report-format-opt-key` · base: `origin/main@cd0cee1b8` (#1620 포함)

## Claim

1. `report.format`의 두 옵션 설명 키(`f.report.format.opt.{html_md,md}.desc`)가 `.opt.`를 포함해, `applyI18n`의 G1-2 가드에 의해 로케일과 무관하게 영어 사전으로 해석됐다 — ko/ja/zh 번역이 `i18n.js`에 있는데도 렌더되지 않았다.
2. 스윕 결과, 같은 함정에 빠진 키는 이 둘뿐이다.
3. 키를 `.option.` 형태로 개명해 해소했고, 가드 커버리지를 감사 필드 4개에서 **스키마 전체 섹션**으로 넓혔다.

## Evidence

### 재현 (수정 전 트리, RED)

```
$ go test ./internal/web/ -run TestEveryOptionDesc -count=1
--- FAIL: TestEveryOptionDescKeyAvoidsOptGuard (0.01s)
    option_desc_test.go:54: field "report.format" option "html+md" OptionDesc "f.report.format.opt.html_md.desc" contains ".opt." — …
    option_desc_test.go:54: field "report.format" option "md" OptionDesc "f.report.format.opt.md.desc" contains ".opt." — …
FAIL	github.com/modu-ai/moai-adk/internal/web	0.846s
```

baseline 0 관측: 새 테스트는 수정 전 트리에서 정확히 2건을 잡고 실패한다 — 공허하게 통과하지 않는다.

### 스윕 (전 로케일, `.desc` ∧ `.opt.`)

```
$ grep -on '"[^"]*\.opt\.[^"]*\.desc"' internal/web/assets/i18n.js | sed 's/.*:"//;s/"$//' | sort -u
f.report.format.opt.html_md.desc
f.report.format.opt.md.desc
```

Go 쪽 `OptionDesc` 대입 지점도 `schema_sections.go:510,512` 둘뿐(`grep -rn '\.opt\.' internal/ --include='*.go'`). 감사 필드는 이미 `.option.` 접두사를 쓴다.

### 수정 후 (GREEN)

```
$ gofmt -l internal/settings internal/web
internal/settings/tier_test.go        ← 기존 미포맷 (이 카드 미변경)
internal/web/viewmodel_ops.go         ← 기존 미포맷 (이 카드 미변경)

$ go vet ./internal/settings/... ./internal/web/...     # 무출력

$ go test ./internal/web/... ./internal/settings/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	14.266s
ok  	github.com/modu-ai/moai-adk/internal/settings	2.693s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.403s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.751s
```

## Baseline-attribution

전부 이 트리(`.claude/worktrees/t212`), base `cd0cee1b8`, 이번 실행에서 측정. RED는 키 개명 **전**, GREEN은 개명 **후** 같은 트리.

## 변경 내용

- `internal/settings/schema_sections.go` — `f.report.format.opt.{html_md,md}.desc` → `f.report.format.option.{html_md,md}.desc`. 왜 `.option.`이어야 하는지 주석으로 고정(다음 복사자가 라벨 키 모양을 다시 베끼지 않도록).
- `internal/web/assets/i18n.js` — 4개 로케일 블록 동시 개명(번역 텍스트 무변경).
- `internal/web/option_desc_test.go` (신규) — `AllSections()` 전수 스윕 3종: `.opt.` 금지 / 4로케일 존재 / 렌더 경로. 필드 목록을 손으로 관리하지 않으므로 앞으로 추가되는 설명 보유 필드가 자동으로 같은 보호를 받는다. 스윕이 공허해지지 않도록 알려진 5개 필드 floor 단언 포함.

라벨 키(`f.report.format.opt.html+md` 등)는 손대지 않았다 — G1-2에 따라 열거형 **라벨**은 영어 고정이 의도된 동작이다. G1-2 가드 자체도 무변경.

## Gaps

- ko 라벨이 브라우저에서 실제로 렌더되는지는 **관측하지 않았다**. `applyI18n`은 JS 런타임이라 Go 테스트로 실행 경로를 태울 수 없다 — 기계적 근거는 키 모양 가드 + 4로케일 사전 존재 + 렌더 경로 `data-i18n` 방출 3개다. 감사 필드(t206)와 동일한 근거 수준.
- 전체 스위트는 로컬에서 돌리지 않았다(레인 규율). 전 패키지 판정은 CI 몫.

## Residual-risk

- `.opt.` 가드가 문자열 부분일치라, 훗날 `.opti…` 류 세그먼트를 쓰면 의도치 않게 걸릴 수 있다. 이번 스윕 범위 밖.
- 개명된 키를 참조하는 외부 소비자는 없다(웹 콘솔 사전 전용, `grep -rn` 확인).
