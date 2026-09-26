# t1207 — Windows 경로 deny 규칙이 실제 `C:/` 명령을 막도록 수리

## Claim

1. 템플릿·로컬 deny 규칙 중 Windows 드라이브 경로 계열 5건(`rm -rf` 2건, `del`, `rmdir`, `Remove-Item`)은 `C\:/` 처럼 콜론을 이스케이프했다. Claude Code 2.1.283은 이 역슬래시를 글자 그대로 맞춘다. 그래서 5건 모두 실제 `C:/` 명령을 하나도 막지 못했다.
2. 이스케이프를 뺀 형태(`Bash(rm -rf C:/:*)` 등)는 의도한 범위만 막는다. 드라이브 루트와 그 뒤에 인자가 붙은 형태, 그리고 리터럴 `C:/*`를 막고, `C:/Windows` 같은 하위 경로는 통과시킨다. 문법은 유닉스 형제 규칙(`rm -rf /:*`, `rm -rf /\* *`)과 같고, 기동 경고는 0건이다.
3. 수리는 `.moai/config/sections/tool-policy.yaml`(SSOT)에서 시작해 템플릿, 로컬 `settings.json`(`moai tool-policy build --local-only`로 재생성) 순으로 반영했다. 테스트는 새 형태를 요구하고, `\:` 가 다시 들어오면 실패한다.

## Evidence

측정 조건: Claude Code 2.1.283, `claude -p --setting-sources user --settings <fixture> --permission-mode bypassPermissions --model haiku --output-format json`. 거부 여부는 `.permission_denials[].tool_input.command` 로 읽었다. 규칙 5건은 각각 앞에 `echo ` 만 붙인 대리 규칙으로 옮겨 쟀다. 문자열 모양은 원래 규칙과 같고, 실행해도 해가 없다. 양성 대조는 `Bash(echo ctrl ok)` 이며 두 실행 모두에서 거부됐다.

**수리 전후 거부 집합** (명령 17개, 각 픽스처 1회)

| 대리 명령 | 수리 전 (`C\:/`) | 수리 후 (`C:/`) |
|---|---|---|
| `echo ctrl ok` (양성 대조) | 거부 | 거부 |
| `echo rm -rf C:/` | 통과 | **거부** |
| `echo rm -rf C:/ x` | 통과 | **거부** |
| `echo rm -rf C:/Windows` | 통과 | 통과 |
| `echo rm -rf C\:/` | 거부 | 통과 |
| `echo rm -rf C:/*` | 통과 | **거부** |
| `echo rm -rf C:/* x` | 통과 | **거부** |
| `echo rm -rf C:/*x` | 통과 | 통과 |
| `echo del /S /Q C:/` | 통과 | **거부** |
| `echo del /S /Q C:/Windows` | 통과 | 통과 |
| `echo del /S /Q C\:/` | 거부 | 통과 |
| `echo rmdir /S /Q C:/` | 통과 | **거부** |
| `echo rmdir /S /Q C:/Windows` | 통과 | 통과 |
| `echo rmdir /S /Q C\:/` | 거부 | 통과 |
| `echo Remove-Item -Recurse -Force C:/` | 통과 | **거부** |
| `echo Remove-Item -Recurse -Force C:/Windows` | 통과 | 통과 |
| `echo Remove-Item -Recurse -Force C\:/` | 거부 | 통과 |

수리 전에 거부된 것은 역슬래시가 글자 그대로 들어간 명령(`C\:/`)뿐이다. Windows에서 사람이나 모델이 쓸 형태가 아니다.

**기동 경고** — 후보 픽스처와 재생성한 로컬 `settings.json` 으로 각각 기동: 규칙 경고 0건.

**테스트·빌드**

- `TestSettingsTemplateDenyWildcardSyntax` 에 새 형태 5건을 요구하고, `\:` 를 담은 Bash 규칙을 거부하는 단언을 추가했다. 수리 전 실행 → darwin, linux, windows 모두 FAIL. 수리 후 `go test ./internal/template -count=1` → `ok … 62.962s`.
- `moai tool-policy build --local-only` → `regenerated 1 target(s)`, `deny=55`. `make build` → exit 0(`tool-policy-drift-check` 통과). `go vet`·`gofmt -l` → 출력 없음.
- 저장소 전체에서 `C\:` 규칙 사본을 고정 문자열로 찾았다. `.moai/reports/` 밖에는 남지 않았다.

**이력** — `C\:/` 형태는 `e21d85d8d`(2026-02-07, 설정 템플릿 이관)가 Go 코드에서 그대로 옮겨 온 것이다. 이스케이프의 설계 근거는 남아 있지 않다.

## Baseline-attribution

작업트리 `.claude/worktrees/t1207`, 브랜치 `WT-windows-deny-rules`, 기준 로컬 develop `9d88a99e9`. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다.

## Gaps

- **실제 명령 형태 측정은 얻지 못했다.** 재생성한 로컬 설정으로 `rm -rf C:/` 등 6개 명령을 돌리려 했지만, 모델이 삭제 명령이라며 실행을 거절했다(도구 호출 0회). 판정은 `echo` 대리 측정에만 기댄다. 매칭은 문자열 기준이라 앞에 붙인 `echo ` 가 결과를 바꾸지 않는다는 추론에 의존한다.
- Windows 실기에서 측정하지 않았다. Windows의 Claude Code가 Bash 도구 명령 문자열이나 경로를 정규화하는지는 모른다.
- 역슬래시 경로(`C:\`), 소문자 드라이브(`c:/`), 다른 드라이브(`D:/`)는 의도한 범위에 없어서 추가하지 않았다. 그런 명령은 여전히 막히지 않는다.
- 각 픽스처는 1회씩만 돌렸다.

## Residual-risk

- Windows의 Claude Code는 PowerShell 도구를 따로 가진다. PowerShell 도구로 실행되는 `Remove-Item` 이나 `del` 은 `Bash(...)` 규칙의 적용 대상이 아니다. 그 경로는 이 카드로 막히지 않는다.
- 권한 규칙은 명령 문자열 패턴이라 보안 경계가 아니다. 인용부호나 다른 철자를 쓰면 빠져나간다.
- 역슬래시가 글자 그대로 들어간 `C\:/` 명령은 이제 통과한다. 실제로 쓰이지 않는 형태라 위험은 낮다고 보지만, 측정한 사실로 남긴다.
