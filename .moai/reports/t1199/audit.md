# t1199 독립 감사 — deny 규칙 `*`·`:*` 혼용 제거

- 감사 대상: `879f100c8` (브랜치 `WT-deny-wildcard-syntax`), 기준 `85d99c6b4`
- 기준→대상 커밋: `879f100c8` 1건뿐 (`git log --oneline 85d99c6b4..879f100c8`)
- 감사 환경: Claude Code `2.1.283`, darwin, 워크트리 `.claude/worktrees/t1199`
- 감사자는 추적 파일을 수정하지 않았다. 스크래치 산출물은 세션 스크래치패드 `t1199-audit/` 에 있으며 휘발성이다 — 판정에 쓴 원출력은 아래에 그대로 옮겨 적었다.

## 판정

**PASS-WITH-DEBT** — 차단(blocking) 결함 0건, 선택(optional) 4건.

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 경고 3→0 재현, 거부 집합이 구 규칙과 일치(3개 경로 모양 전부), 테스트가 두 변이를 모두 잡음 |
| Security (25%) | 88 | PASS | 구 범위 복원이 카드 의도와 일치. 단 `C\:/` 규칙은 구·신 모두 실제 `C:/` 명령에 무력(기존 결함, F2) |
| Craft (20%) | 88 | PASS | 패키지 테스트·vet·lint·gofmt 통과. 테스트 함수 위 주석이 새 의도와 모순(F1), 레인 원증거 휘발(F4) |
| Consistency (15%) | 95 | PASS | 템플릿·tool-policy.yaml·로컬 settings.json 일관, 드리프트 검사 통과, 혼용 규칙 잔존 0 |

조화평균: 4 / (1/95 + 1/88 + 1/88 + 1/95) = **91.4**. 필수 통과 차원(Functionality·Security) 모두 PASS.

## 레인 주장별 검증

### 주장 1 — 구 `/*:*` 는 리터럴 `*`, t1185 의 `/*` 는 범위 확대 → **확인**

에코 대리 규칙으로 독립 측정 (`--permission-mode bypassPermissions`, 각 픽스처 공통 양성 대조 `Bash(echo ctrl:*)`, 음성 대조 `echo free ok`). 모델이 시도한 명령과 거부된 명령을 stream-json 에서 그대로 추출했다.

명령:
```
claude -p --setting-sources user --settings <f_{legacy,new,wide}.json> --permission-mode bypassPermissions --model haiku --max-turns 20 --output-format stream-json --verbose < prompt.txt
```
픽스처 규칙: legacy `echo zz /*:*`·`echo zz ~/*:*` / new `echo zz /\* *`·`echo zz ~/\* *` / wide `echo zz /*`·`echo zz ~/*`

원출력 (세 픽스처 모두 attempted 는 8개 명령 전부 동일):
```
== legacy
denied:
echo ctrl ok
echo zz /*
echo zz /* x
echo zz ~/*
== new
denied:
echo ctrl ok
echo zz /*
echo zz /* x
echo zz ~/*
== wide
denied:
echo ctrl ok
echo zz /*
echo zz /* x
echo zz /*x
echo zz /etc
echo zz ~/*
echo zz ~/foo
```
양성 대조는 세 픽스처 모두 거부, 음성 대조 `echo free ok` 는 세 픽스처 모두 허용. wide 형태만 `/*x`·`/etc`·`~/foo` 까지 거부한다 — 범위 확대가 실재한다.

실제 `rm` 픽스처(스크래치 안에 만든 빈 디렉터리, 절대 경로)로도 확인:
```
new denials: []
wide denials: ["rm -rf /private/tmp/claude-501/.../t1199-audit/fx_wide"]
ls: fx_new: No such file or directory
drwxr-xr-x@ 3 goos  wheel  96 Sep 26 08:57 fx_wide
```
t1185 형태는 무관한 절대 경로 삭제까지 막고, 수리본은 막지 않는다.

### 주장 2 — 새 형태는 경고 0 이면서 구 규칙과 같은 거부 집합 → **확인**

거부 집합은 위 표에서 legacy = new. 기동 경고:
```
claude -p --setting-sources user --settings <fixture> --model haiku --max-turns 1 < ok.txt   (stdout+stderr → 파일)
grep -c mixes warn_base.txt warn_local.txt warn_tmpl.txt warn_newecho.txt
warn_base.txt:3
warn_tmpl.txt:0
warn_local.txt:0
warn_newecho.txt:0
```
`warn_base` = 기준 `85d99c6b4` 의 로컬 settings.json permissions 블록(양성 대조), `warn_local` = 이 트리의 로컬 permissions 블록, `warn_tmpl` = 템플릿 deny 배열(55개 규칙, `{{` 없음 확인 후 추출). 양성 대조에서 나온 경고 원문:
```
Bash(rm -rf /*:*) mixes
Bash(rm -rf C\\:/*:*) mixes
Bash(rm -rf ~/*:*) mixes
```

### 주장 3 — tool-policy.yaml ↔ 로컬 settings.json 일관, 드리프트 검사 통과 → **확인 (재생성 자체는 미재현)**

```
go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$' -count=1 -v
--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.077s
```
이것은 `make build` 의 `tool-policy-drift-check` 가 실행하는 명령 그대로다. `moai tool-policy build --local-only` 재생성은 추적 파일을 쓰므로 재실행하지 않았다(Gaps). tool-policy.yaml 은 템플릿 미러가 없는 로컬 전용 파일이다(`internal/template/templates/.moai/config/sections/` 에 tool 계열 파일 없음).

### 주장 4 — 템플릿 테스트가 새 형태를 요구하고 확대 형태를 거부 → **확인 (변이 2종 사멸)**

```
go test ./internal/template -run '^TestSettingsTemplateDenyWildcardSyntax$' -count=1 -v
--- PASS: TestSettingsTemplateDenyWildcardSyntax (0.00s)
    --- PASS: TestSettingsTemplateDenyWildcardSyntax/darwin (0.00s)
    --- PASS: TestSettingsTemplateDenyWildcardSyntax/linux (0.00s)
    --- PASS: TestSettingsTemplateDenyWildcardSyntax/windows (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.172s
```
변이 시험: `git archive 879f100c8 go.mod go.sum internal pkg` 를 스크래치에 풀고 템플릿만 바꿔 같은 테스트를 돌렸다(추적 파일 무변경).

- 변이 A — 기준 `85d99c6b4` 의 템플릿(t1185 확대 형태)으로 교체: darwin/linux/windows 모두 FAIL
  ```
  settings_test.go:90: missing Bash deny rule "Bash(rm -rf /\\* *)"
  ...
  settings_test.go:95: Bash deny rule "Bash(rm -rf /*)" widens the literal-* scope to every path
  FAIL	github.com/modu-ai/moai-adk/internal/template	0.326s
  ```
- 변이 B — `rm -rf /*` 를 구 혼용 형태 `rm -rf /*:*` 로 교체: 세 플랫폼 모두 FAIL
  ```
  settings_test.go:85: Bash deny rule mixes wildcard with legacy prefix syntax: "Bash(rm -rf /*:*)"
  ```

패키지 전체: `go test ./internal/template -count=1` → `ok  github.com/modu-ai/moai-adk/internal/template	64.991s`. `go vet ./internal/template/` → `vet-ok`, `gofmt -l internal/template/settings_test.go` → 출력 없음, `golangci-lint run ./internal/template/` → `0 issues.`

### 잔존 혼용 규칙·형제 사본 스윕 → **0건**

로컬 settings.json·템플릿·settings.local.json 의 권한 규칙에서 `:*` 앞에 `*` 가 있는 규칙, tool-policy.yaml 의 같은 모양 `args_pattern` 을 검색 → 모두 출력 없음. 확대 형태 문자열(`rm -rf /*)`·`~/*)`·`C\:/*)`)은 테스트의 `widened` 목록 3줄에만 남아 있다. `.codex` 템플릿·`agentemit`·`codexadapter` 는 settings deny 목록을 소비하지 않고, moai 자체 매처(`internal/permission`)는 `doctor permission` 에서 빈 `RulesByTier` 로만 쓰여 settings deny 를 평가하지 않는다 — 즉 이 규칙들의 의미를 정하는 것은 Claude Code 뿐이다.

## Findings

- **F1** [Low] [optional] `internal/template/settings_test.go:51-53` — 테스트 함수 위 주석 "Keep the root, home, and Windows path variants as wildcards." 가 t1185 시절 문구 그대로 남아, 바로 아래 새 주석("`\*` keeps that literal match … an unescaped `/*` would widen")과 정반대를 말한다. 확신도 높음. 수리: 첫 주석을 "리터럴 `*` 매칭을 공백-접미 문법으로 유지한다"는 취지로 한 줄 고친다.
- **F2** [Medium, 기존 결함] [optional — 이 카드 범위 밖] 템플릿·로컬의 `C\:/` 계열 deny 규칙(`rm -rf C\:/:*`, `rm -rf C\:/\* *`, 그리고 `del /S /Q C\:/:*` 등 형제 규칙까지 추정)은 CC 2.1.283 에서 `\:` 를 풀지 않고 **리터럴 역슬래시**로 맞춘다. 실제 `C:/` 명령은 구·신 어느 형태로도 막히지 않는다. 확신도: `echo` 대리 기준 높음, Windows 실기 미측정.
  ```
  (규칙 echo zz C\:/*:* / echo zz C\:/\* * 각각)
  attempted: echo ctrl ok|echo zz C:/*|echo zz C:/* x|echo zz C:/*x|echo zz C:/Windows|echo free ok
  denied: echo ctrl ok                       ← legacy, new 동일
  (규칙에 echo yy C\:/:* 추가, 명령에 역슬래시 포함)
  attempted: echo ctrl ok|echo zz C\:/*|echo yy C:/|echo yy C\:/|echo free ok
  denied: echo ctrl ok|echo zz C\:/*|echo yy C\:/   ← legacy, new 동일
  ```
  이 카드의 "범위를 바꾸지 않는다"는 요건은 지켜졌다(구·신 모두 무력). 수리: 별도 카드로 `C:/` 규칙의 올바른 표기(콜론 비이스케이프 공백-접미 형태 등)를 실측해 정한다. `del`/`rmdir`/`Remove-Item` 형제 규칙도 같은 스윕에 넣는다.
- **F3** [Low] [optional] `.moai/reports/t1199/verdict.md` Claim 3 "`~/`·`C\:/` 도 같은 모양이다"는 `C\:/` 에 대해 추론이었다(Gaps 에 정직하게 기재됨). 감사 측정으로 "동일"은 참이 됐지만 "둘 다 무력"이라는 사실은 Residual-risk 에 없다. 수리: F2 를 판정서 Residual-risk 에 한 줄 추가하거나 후속 카드로 넘긴다.
- **F4** [Low] [optional] 판정서의 거부 집합 표와 rm 픽스처 결과는 요약이며, 원출력은 "세션 스크래치패드에 있었다(휘발)". 증거 영속 의무(`agent-common-protocol.md` § Parallel Execution — Evidence persistence)에 못 미친다. 이 감사 보고가 같은 측정을 원출력과 함께 재현해 공백을 메운다. 수리 불필요(이 문서로 대체), 다음 카드부터 원출력을 `.moai/reports/<card>/` 에 남긴다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim** — `879f100c8` 은 혼용 경고 3건을 0건으로 만들면서 `/`·`~/` 경로의 거부 집합을 구 규칙과 동일하게 유지하고, t1185 가 넓힌 범위를 되돌린다. 템플릿·yaml·로컬 사본이 일관되며, 테스트는 두 방향의 회귀를 모두 잡는다.

**Evidence** — 위 각 절의 명령과 원출력.

**Baseline-attribution** — 모든 측정은 이번 실행에서 이 워크트리(HEAD `879f100c8`, 작업트리 깨끗)와 기준 `85d99c6b4` 를 대상으로 했다. `302cd9e04`(t1185)는 기준의 조상(`merge-base --is-ancestor` → 0)이고 `origin/main` 의 조상이 아니다(→ 1) — 판정서 Residual-risk 의 해당 문장과 일치.

**Gaps**
- `moai tool-policy build --local-only` 와 `make build` 는 재실행하지 않았다(추적 파일 쓰기). 드리프트 검사 테스트는 직접 돌렸다.
- Windows 실기, 실제 `rm -rf /*`·`rm -rf ~/*` 는 측정하지 않았다(대리 규칙만).
- `~/` 경로의 실제 `rm` 픽스처는 재현하지 않았다(에코 대리로만).
- 각 픽스처 1회 실행. CI·CodeRabbit 은 보지 않았다.
- `\*` 이스케이프의 공식 문서 근거는 확인하지 않았다 — 관측 동작에만 기댄다.

**Residual-risk**
- `\*` 의미가 이후 CC 버전에서 바뀌면 범위가 조용히 달라진다. 테스트는 문자열만 보므로 그 변화를 잡지 못한다.
- 권한 규칙은 문자열 패턴이라 `rm -rf "/"*`·`rm -rf //*` 같은 다른 철자는 어느 형태로도 막히지 않는다 — 보안 경계가 아니다.
- develop 에서 빌드된 rc 설치본이 이미 t1185 확대 형태로 배포돼 있을 수 있다. 이 커밋이 develop 에 착지하고 `moai update` 가 settings.json 을 재배포해야 사라진다.
- F2 의 `C:/` 규칙 무력은 이 카드와 무관하게 남는다.
