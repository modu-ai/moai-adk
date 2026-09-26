# t1207 감사 — Windows 드라이브 deny 규칙 수리 (`513c3fdca`, 기준 `9d88a99e9`)

감사자: sync-auditor (독립·회의적 감사). 구현 파일은 읽기만 했고, 이 파일 말고는 추적 파일을 쓰지 않았다. 스크래치: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2cdcb7ed-5896-4d85-a8b7-9a19fb627907/scratchpad/t1207-audit/`.

## 종합 판정: **PASS**

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 94 | PASS | 레인의 17행 표를 28개 명령·2회 반복으로 독립 재현했고, 한 칸도 어긋나지 않았다. 새 테스트는 옛 템플릿에서 실패한다 |
| Security (25%) | 88 | PASS | 의도한 범위가 유닉스 형제 규칙과 정확히 같다. 넓어진 곳도, 좁아진 곳도 없다. 남은 우회 경로(인용부호·소문자·`/c/`·PowerShell 도구)는 이 카드 전부터 있던 구조적 한계다 |
| Craft (20%) | 90 | PASS | 가드 테스트가 새 형태 5건을 요구하고 `\:` 를 거부한다. vet·lint·gofmt 모두 깨끗하다. yaml·로컬 쪽에는 `\:` 가드가 없다(선택 항목) |
| Consistency (15%) | 95 | PASS | SSOT·템플릿·로컬 세 곳의 5개 규칙이 모두 같고, 드리프트 테스트도 통과한다 |

조화평균 = 4 / (1/94 + 1/88 + 1/90 + 1/95) ≈ **91.7**. 필수 통과 차원(Functionality, Security)은 둘 다 기준을 넘었다. blocking 결함은 없다.

## 주장별 검증

### 주장 1 — 옛 `C\:/` 규칙은 역슬래시가 글자 그대로 든 명령만 막는다 → **확인**

Claude Code 2.1.283(`claude --version` → `2.1.283 (Claude Code)`)에서 옛 규칙 모양을 `echo` 대리 규칙으로 옮긴 픽스처 `old.json` 으로 쟀다. 2회차는 stream-json 으로 실행해, 모델이 28개 명령을 모두 적힌 그대로 시도했음을 확인했다.

명령: `claude -p --setting-sources user --settings old.json --permission-mode bypassPermissions --model haiku --max-turns 40 --output-format stream-json --verbose < prompt.txt`

거부 목록(`.permission_denials[].tool_input.command`). 1회차와 2회차가 같다.
```
echo ctrl ok
echo rm -rf C\:/
echo rm -rf /
echo rm -rf / x
echo rm -rf /*
echo rm -rf /* x
```
C 드라이브 계열에서 거부된 것은 `echo rm -rf C\:/` 하나다. 실제 `C:/` 명령과 `ls -ld C:/` 비-echo 대리 명령은 하나도 거부되지 않았다.

### 주장 2 — 새 규칙이 루트, 루트+인자, 리터럴 `C:/*`(+인자)만 막는다 → **확인. 범위가 유닉스 형제 규칙과 같다**

픽스처 `new.json` 도 같은 명령으로 두 번 돌렸다(1회차 json, 2회차 stream-json). 두 번 모두 28개 명령을 적힌 그대로 시도했고, 거부 목록도 같았다.
```
echo ctrl ok
echo rm -rf C:/
echo rm -rf C:/ x
echo rm -rf C:/*
echo rm -rf C:/* x
echo del /S /Q C:/
echo rmdir /S /Q C:/
echo Remove-Item -Recurse -Force C:/
echo rm -rf /
echo rm -rf / x
echo rm -rf /*
echo rm -rf /* x
echo rm -rf  C:/
ls -ld C:/
ls -ld C:/ x
echo rm -rf C:/ && echo after
```
통과한 명령: `C:/Windows`(rm·del·rmdir·Remove-Item 4건), `C:/*x`, `C\:/`, `/tmp`, `/*x`, `c:/`, `"C:/"`, `D:/`, `ls -ld C:/Windows`.

유닉스 형제와 나란히 놓으면 다음과 같다.

| 형태 | 유닉스 (`rm -rf /:*`, `/\* *`) | Windows 신규 (`C:/:*`, `C:/\* *`) |
|---|---|---|
| 루트 | 거부 | 거부 |
| 루트 + 인자 | 거부 | 거부 |
| 하위 경로 (`/tmp`, `C:/Windows`) | 통과 | 통과 |
| 리터럴 `*` (+인자) | 거부 | 거부 |
| `*x` | 통과 | 통과 |

두 행렬이 칸마다 같으므로 의도한 범위와 맞는다. 레인 표에 없던 것도 새로 쟀다.
- 공백 두 칸(`rm -rf  C:/`)은 거부된다. CC가 공백을 정규화한다.
- 복합 명령(`… && echo after`)은 거부된다. CC가 하위 명령으로 나눠 매칭한다.
- 첫 토큰이 `echo` 가 아닌 `ls -ld C:/` 대리 명령도 똑같이 거부된다. 앞에 붙인 `echo` 가 매칭을 바꾸지 않는다는 레인의 추론을 뒷받침한다.
- 인용부호(`"C:/"`), 소문자(`c:/`), 다른 드라이브(`D:/`)는 통과한다. 레인 Gaps·Residual-risk 와 일치한다.

넓어지거나 좁아진 것: diff는 규칙 5건의 `\:` → `:` 치환뿐이다(`git show 513c3fdca`). 새로 통과하게 된 것은 리터럴 `C\:/` 명령 하나이며, 레인이 Residual-risk에 이미 적었다.

**기동 경고 0건 — 양성 대조로 확인했다.**
- 대조군: 경고를 내야 하는 `Bash(echo rm -rf /*:*)` 픽스처로 기동하면 다음 경고가 나온다.
  ```
  Permission deny rule (…/badwarn.json): Bash(echo rm -rf /*:*) mixes * with the trailing :* prefix syntax, so it is matched as a literal prefix (the * is not expanded) …
  OK
  ```
- 커밋된 로컬 `.claude/settings.json` 으로 기동(`claude -p --setting-sources user --settings <worktree>/.claude/settings.json --model haiku --max-turns 1 < ok.txt`) → 출력은 `OK` 뿐이다. `grep -ci 'warn|invalid|unexpected|permission rule'` → `0`.
- `new.json` 으로 기동 → 출력 `OK` 뿐.

### 주장 3 — SSOT·템플릿·로컬이 일치한다 → **확인**

```
$ go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$' -count=1 -v
--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.089s
$ go test ./internal/config/toolpolicy/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.129s
```
세 곳을 grep으로 대조했다. yaml `160/167/176/183/190`, 템플릿 `553-557`, 로컬 `533/538/568/569/572` 에 같은 5개 규칙(`rm -rf C:/:*`, `rm -rf C:/\* *`, `del /S /Q C:/:*`, `rmdir /S /Q C:/:*`, `Remove-Item -Recurse -Force C:/:*`)이 있다. 템플릿의 이 5행은 `{{- if eq .Platform "windows"}}` 블록(396-400행대) 밖에 있어서 모든 플랫폼에 배포된다. 테스트의 darwin·linux·windows 하위 테스트가 모두 이 규칙을 요구하고 통과한다. `settings.local.json` 에서 `C\\:` 는 `0` 건이다.

남은 사본은 고정 문자열로 찾았다(양성 대조 포함).
```
$ git grep -nF 'C\:' -- .moai/reports | head -3      # 대조: 적중
.moai/reports/t1199/audit.md:15:…
$ git grep -nF 'C\:' -- ':!.moai/reports'
Binary file assets/images/home-hygiene-infographic-en.png matches
Binary file docs-site/static/images/sections/workflow-commands-en.png matches
$ grep -rlF 'C\:' . --exclude-dir=.git | grep -v '^./.moai/reports/'
(출력 없음)
```
PNG 두 건은 바이너리 우연 적중이라 규칙 사본이 아니다. `tool-policy.yaml` 의 템플릿 사본은 없다(`git ls-files | grep tool-policy` → yaml 1개, SPEC·docs-site 문서만).

### 주장 4 — 템플릿 테스트가 새 형태를 요구하고 `\:` 를 거부하며, 옛 템플릿에서 실패한다 → **확인(스크래치 사본에서 실측)**

현재 트리에서:
```
$ go test ./internal/template -run '^TestSettingsTemplateDenyWildcardSyntax$' -count=1 -v
--- PASS: TestSettingsTemplateDenyWildcardSyntax (0.00s)
    --- PASS: …/darwin (0.00s)
    --- PASS: …/linux (0.00s)
    --- PASS: …/windows (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.157s
```
기준 커밋 `9d88a99e9` 를 `git archive` 로 스크래치에 풀었다. 그 위에 `513c3fdca` 의 `settings_test.go` 만 덮어써서 실행했다(추적 파일은 건드리지 않았다).
```
--- FAIL: TestSettingsTemplateDenyWildcardSyntax (0.00s)
    --- FAIL: TestSettingsTemplateDenyWildcardSyntax/darwin (0.00s)
        settings_test.go:95: Bash deny rule escapes ':' and cannot match a real drive path: "Bash(rm -rf C\\:/:*)"
        … (5건)
        settings_test.go:100: missing Bash deny rule "Bash(rm -rf C:/:*)"
        … (5건)
    --- FAIL: …/linux  (같은 10건)
    --- FAIL: …/windows (같은 10건)
```
새 규칙 `Bash(rm -rf C:/:*)` 는 기존의 「`*` + `:*` 혼용」 검사에도 걸리지 않는다. `TrimSuffix(":*)")` 결과인 `Bash(rm -rf C:/` 에 `*` 가 없기 때문이다.

그 밖의 검사:
```
$ go test ./internal/template -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	72.907s
$ gofmt -l internal/template/settings_test.go      → (출력 없음)
$ go vet ./internal/template/                        → vet exit=0
$ golangci-lint run ./internal/template/...          → 0 issues.
```
이력 주장도 확인했다. `git log -S'C\\:/' -- …settings.json.tmpl` 에서 가장 오래된 커밋이 `e21d85d8d 2026-02-07 refactor(template): migrate settings.json to template-based generation` 이다.

## 발견 사항 (구조화 결함 목록)

- **F1** [Low] [optional] `internal/config/toolpolicy` — `\:` 거부 가드는 렌더된 **템플릿**만 검사한다. `tool-policy.yaml` 에 `\:` 가 다시 들어와 `moai tool-policy build --local-only` 로 로컬만 재생성하면, 드리프트 테스트는 집합이 같다는 것만 보므로 통과한다. 그러면 로컬 settings에 무력한 규칙이 조용히 돌아온다. 확신도: 중간(코드 경로 추론, 실행 재현 없음). 필요한 수리(선택): drift 테스트나 yaml 로더 테스트에 「Bash `args_pattern` 에 `\:` 금지」 단언을 추가한다.
- **F2** [Low] [optional] `.moai/reports/t1207/verdict.md` Gaps — 빠진 철자 목록에 Git Bash 고유의 드라이브 표기 `/c/` 가 없다. Windows의 Claude Code Bash 도구는 Git Bash로 실행되므로 `rm -rf /c/` 가 자연스러운 형태다. `/c/` 는 `rm -rf /:*` 에 걸리지 않는다(단어 경계 조건 때문에 `/tmp` 가 통과하는 것과 같은 이유이며, 이번 측정의 `/tmp` 행이 그 근거다). 확신도: 매칭 규칙 기준 높음, Windows 실기 미측정. 필요한 수리(선택): Gaps에 한 줄을 추가하거나 후속 카드로 넘긴다.
- **F3** [Info] [optional] 인용부호 우회를 실측했다. `echo rm -rf "C:/"` 는 새 규칙에서 통과한다. 레인 Residual-risk의 「인용부호로 빠져나간다」를 실측으로 뒷받침한다. 수리는 필요 없다. 권한 규칙이 보안 경계가 아니라는 기존 인식과 같다.

blocking 결함은 없다.

## 증거 충분성 판단 — echo 대리 측정

실제 삭제 명령은 측정하지 못했다(레인 Gaps). 그래도 판정 근거로 충분하다고 본다.
1. deny 규칙 매칭은 명령 문자열 위에서 이루어진다. 이번에 첫 토큰이 다른 두 대리 명령 계열(`echo …`, `ls -ld …`)이 같은 경계에서 같은 결과를 냈다.
2. 공백 정규화와 복합 명령 분할이 대리 명령에서도 작동했다. 매처가 실제 명령과 같은 파싱 경로를 탄다는 정황이다.
3. 실제 `rm` 에 CC가 따로 안전 검사를 붙이더라도, 그것은 거부를 **더하는** 방향이다. deny 규칙 적중을 없애는 방향이 아니다.

남는 불확실성은 Windows 실기에서 명령 문자열이나 경로를 정규화하는지다. 레인 Gaps에 정직하게 적혀 있다.

## 레인 판정서의 정직성

Claim·Evidence·Gaps·Residual-risk 모두 이번 재측정과 모순되지 않는다. 17행 표는 칸마다 재현됐다. Gaps의 「각 픽스처 1회」는 이번 감사의 2회 반복으로 보강했다. 빠진 것은 F2의 `/c/` 한 줄뿐이다.

## Baseline-attribution

작업트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1207`, 브랜치 `WT-windows-deny-rules`, HEAD `513c3fdca`(`git rev-parse --short HEAD`), 감사 시작 시 `git status --short` 는 비어 있었다. 모든 측정은 이번 실행에서 이 트리와 CC 2.1.283으로 관측했다. 옛 템플릿 실패는 `9d88a99e9` 스크래치 사본에서 쟀다.

## Gaps

- Windows 실기는 측정하지 않았다(PowerShell 도구 경로, Git Bash 경로 정규화 포함).
- 실제 `rm`/`del`/`rmdir`/`Remove-Item` 명령 문자열은 쓰지 않았다. 지시에 따라 `echo`/`ls` 대리 명령만 썼다.
- F1은 코드 경로 추론이며, 재현 실행은 하지 않았다.
- `make build` 는 다시 실행하지 않았다. 같은 검사인 `tool-policy-drift-check` 의 테스트 두 건을 직접 실행했다.

## Residual-risk

- 권한 규칙은 문자열 패턴이다. 인용부호(F3 실측), 소문자·다른 드라이브·`C:\`·`/c/`(F2) 철자는 여전히 막지 못한다.
- Windows에서 CC가 명령 문자열을 정규화하는 방식이 macOS와 다르면, 대리 측정 결과가 옮겨가지 않을 수 있다.
