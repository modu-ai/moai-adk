# t1224 판정서 — 훅 matcher PowerShell 대칭 (SPEC-HOOK-MATCHER-POWERSHELL-001)

브랜치 `WT-hook-matcher-powershell` · 흡수 기준 로컬 develop `553e224f3` · 카드 HEAD `f1de3a8bb` (카드 범위 12커밋)

## Claim

Claude Code 의 PowerShell 도구 호출은 기존 PreToolUse matcher `Write|Edit|Bash` 에 걸리지 않아 MoAI 훅 가드(위험 명령 차단, 브랜치 가드, 통합 락, 슬롯 임대, 증거 기록)를 전혀 거치지 않았다. 이 카드는 다음 세 가지를 반영했다.

- `PowerShell` 전용 PreToolUse 블록을 추가했다(D1).
- Go 핸들러의 `ToolName=="Bash"` 분기 10곳을 `IsShellTool` 술어로 바꿨다.
- 판별하지 못한 PowerShell 간접 호출(`iex`, `Start-Process`, `-EncodedCommand`)은 허용하되 사유 `unclassifiable` 로 감사 로그를 한 줄 남긴다(D2).

셸 래퍼의 위험 경고는 Bash 에만 둔다(D6). D1·D2·D6 은 리드가 2026-09-26 운영자 위임 범위에서 결정했고, 착수 승인은 운영자가 레인에서 직접 했다.

## Evidence

**LIVE 측정** (CC 2.1.283, pwsh 7, `--tools PowerShell`, bypassPermissions, `--safe-mode` 미사용, `timeout 180`, `--max-turns 3`, 경우마다 새 스크래치 git 프로젝트, 포그라운드 실행)

| 경우 | 설정 | 관측 | sha256 |
|---|---|---|---|
| A | matcher `Write\|Edit\|Bash` | PowerShell `git status --short` 는 실행됐으나 훅 로그 없음 | run.jsonl `219c4d3b…` |
| B | matcher `PowerShell` | PreToolUse 에서 `tool_name:"PowerShell"`, `tool_input{command,description}`. PostToolUse 에서 `tool_response{stdout,stderr,interrupted,isImage}` | run `0277043a…`, hook `92540172…`, post `7ccb4ff2…` |
| C | 브랜치 바이너리 + branch_guard on | `binary-version.txt` = `-g3a4b623a6`(HEAD 일치). 결과 `BRANCH_GUARD_VIOLATION: git switch in primary checkout`. `probe` 브랜치 생성 안 됨 | run `0e9bd141…`, ver `d7a4f1ee…` |

원본 파일은 `.moai/reports/t1224/live/{A,B,C}/` 에 있다(로컬, gitignore).

**테스트와 정적 검사** (sync-auditor 재측정, HEAD `f1de3a8bb`)

- `go test -count=1 ./internal/hook/ ./internal/cli/ ./internal/template/ -run 'TestHMP'`: 24건 PASS, 세 패키지 모두 ok
- 기존 가드 회귀: hook 선택 실행 ok, cli 73건, template 73건 PASS
- golangci-lint `0 issues.`, vet·gofmt 깨끗, `GOOS=windows go build` exit 0
- `moai spec lint`: `✓ No findings`
- 인수 기준 13/13 PASS: `progress.md` §E.2/§E.3

**감사**

- plan-audit: 1차 FAIL 0.77, 2차 PASS 0.87(조건 N1~N3 은 `adc493d48` 에서 해소)
- sync-audit: **PASS-WITH-DEBT 88.7** (기능 90 / 보안 85 / 품질 88 / 일관성 92). 보고서 `.moai/reports/t1224/sync-audit.md` (sha256 `5f7064e3…`)

## Baseline-attribution

- 테스트는 이 워크트리 HEAD `f1de3a8bb` 에서 sync-auditor 가 다시 쟀다.
- LIVE 는 2026-09-26 에 레인 세션이 워크트리 밖 스크래치 프로젝트에서 실행했다. 경우 C 의 바이너리는 `3a4b623a6` 에서 빌드했고, 그 뒤 커밋(`b9471ba14`, `f1de3a8bb`)은 문서만 바꿨다.

## Gaps

- pwsh `-EncodedCommand` 가 받아들이는 철자는 측정하지 않았다. 검출기는 문서에 적힌 별칭(`-e`, `-ec`, `-en…` 접두)을 기준으로 한다. 유니코드 대시 형태도 확인하지 않았다(F5).
- REQ-HMP-016 이 요구한 「사이트 변환 전에 페이로드 캡처」 순서는 지키지 못했다. 캡처는 M3 뒤에 했고, 사후에 가정과 일치함을 확인했다.
- 경우 C 의 exit 0, `probe` 부재, 면제 변수 미설정은 레인 세션이 관측했지만 증거 파일에는 남기지 않았다(F6).
- `make build` 는 감사 단계에서 다시 돌리지 않았다. hook 패키지 전체 스위트와 패키지 전체 커버리지도 재지 않았다. CI 가 판정한다.
- 측정은 macOS 에서만 했다. Windows 동작은 빌드 성공 외에는 추론이다.

## Residual-risk

- 브랜치 가드와 통합 락을 로그 없이 통과하는 PowerShell 형태가 남아 있다(차단 목록은 이와 무관하게 먼저 적용된다). 후속 카드 후보다.
  - `git.exe switch …`, `& "git.exe" …` (F1)
  - `pwsh -Command "…"`, `powershell -c "…"` (F2)
  - `& (Get-Command git) …` (F3)
- 호출 대상을 인용부호로 감싼 형태(`& 'git' …`)와 `saps` 별칭도 검출하지 못한다.
- 주석과 이름이 실제 동작과 어긋난 곳이 남아 있다.
  - F4: 통합 락 로그 주석이 실제 동작보다 좁게 쓰여 있다.
  - F7: `pre_tool.go:507` 의 ANCHOR 문구와 `LogBashEvidence` 등 이름이 여전히 Bash 만 가리킨다.
- F8: 소스 가드는 문자열 리터럴만 본다.
- 형제 카드 후보 D4 두 건(`IsWriteOperation` 의 PowerShell 판정, Git Bash 없는 Windows 의 훅 실행)은 발행 여부가 운영자 몫이다.

## 후속 카드용 재현 (F1~F3, EncodedCommand)

이 절은 리드의 요청(2026-09-26)으로 추가했다. 프로브는 트리에 쓰지 않고 `go test -overlay` 로 컴파일할 때만 끼워 넣는다. 프로브 사본은 `.moai/reports/t1224/followup-probe_test.go.txt` (컴파일 대상에서 빠지도록 .txt 확장자 — 오버레이에 쓸 때 원래 이름으로 지정) 에 있다(sha256 `868c34db…`). 오버레이 파일에는 `{"Replace":{"<worktree>/internal/hook/zz_audit_probe_test.go":"<probe path>"}}` 한 항목만 둔다.

```
go test -count=1 -overlay <overlay.json> ./internal/hook/ -run TestAuditProbe -v -o /dev/null
```

HEAD `f4d62e5c3` 에서 실행했고 결과는 `ok … 5.989s` 이다. 관측 출력은 원문 그대로 옮겼다. 브랜치 가드 쪽은 브랜치 가드가 켜진 primary 체크아웃 픽스처에서 도구를 PowerShell 로 두고 쟀다.

```
BRANCH "pwsh -Command \"git switch probe\""       -> decision="allow" unclassifiedLines=0 construct=""   # F2
BRANCH "powershell -c \"git switch probe\""       -> decision="allow" unclassifiedLines=0 construct=""   # F2
BRANCH "& 'git' switch probe"                     -> decision="allow" unclassifiedLines=0 construct=""   # 인용 호출 대상
BRANCH "& \"git.exe\" switch probe"               -> decision="allow" unclassifiedLines=0 construct=""   # F1
BRANCH "saps git -ArgumentList 'switch','probe'"  -> decision="allow" unclassifiedLines=0 construct=""   # Start-Process 별칭
BRANCH "git swi`tch probe"                        -> decision="allow" unclassifiedLines=0 construct=""   # 백틱 이스케이프
BRANCH "git.exe switch probe"                     -> decision="allow" unclassifiedLines=0 construct=""   # F1
BRANCH "& (Get-Command git) switch probe"         -> decision="allow" unclassifiedLines=0 construct=""   # F3
BRANCH "cmd /c git switch probe"                  -> decision="deny"  (BRANCH_GUARD_VIOLATION)            # 대조군: 이미 잡힘
BRANCH "$c = 'git switch probe'; iex $c"          -> decision="allow" unclassifiedLines=1 construct="invoke-expression"  # D2 대로 동작
ILOCK  "iex \"git status\""                       -> decision="allow" unclassifiedLines=1   # F4: merge 가 아닌 명령도 기록
ILOCK  "Start-Process git -ArgumentList 'log'"    -> decision="allow" unclassifiedLines=1   # F4
DENYLIST Bash       "eval \"terraform destroy\""  -> allow   # 기존 Bash 동작(기준선)
DENYLIST PowerShell "iex \"terraform destroy\""   -> allow   # Bash eval 과 같음
DENYLIST PowerShell "terraform destroy"           -> deny  "Dangerous command blocked: (?i)terraform\\s+destroy"
DENYLIST PowerShell "Start-Process terraform -ArgumentList 'destroy'" -> allow
```

위험 명령 차단 목록도 간접 호출(`iex "…"`, `Start-Process … -ArgumentList`) 안의 명령은 보지 못한다. Bash 의 `eval "…"` 과 같은 수준이고 이 카드가 새로 만든 약점은 아니지만, 후속 카드에서 함께 다룰 만하다.

**-EncodedCommand 철자는 측정하지 않았다.** 워크트리 세션 가드가 `pwsh` 호출을 거부하므로 워크트리 밖 스크래치 폴더에서 재현한다. 아래 명령은 `ok` 를 base64(UTF-16LE)로 인코딩한 값을 쓴다.

```
B64=$(printf 'Write-Output ok' | iconv -t UTF-16LE | base64)
for s in -e -ec -en -enc -enco -encodedc -EncodedCommand -ENC --EncodedCommand /enc –enc; do
  printf '%s -> ' "$s"; pwsh -NoProfile -NonInteractive $s "$B64" 2>&1 | head -1
done
```

`ok` 를 출력하는 철자가 pwsh 가 받아들이는 집합이다. 이 집합을 `powershell_indirection.go` 의 검출기와 대조한다. 목록 끝의 `–enc` 는 유니코드 대시(U+2013)로 쓴 형태다.
