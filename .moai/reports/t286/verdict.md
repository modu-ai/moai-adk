# t286 — 위험명령 가드 양방향 결함 수리 (이슈 #1658)

- 카드: t286 · Class B (plan 생략, 원인 미확정 → 재현 우선)
- 워크트리: `.claude/worktrees/t286` · 브랜치 `WT-danger-cmd-guard`
- 기준: develop `1e5199b88`

## Claim

`moai hook pre-tool` 의 위험명령 가드가 양방향으로 틀린 문제를 닫았다. 플래그 순서를 바꾸면
빠져나가던 우회가 사라졌고, 따옴표 안 데이터를 명령으로 오탐하던 차단이 사라졌다. 완화는
정규식을 느슨하게 하는 방향이 아니라 **데이터 컨텍스트 파악**(토큰화 + 따옴표 접기)으로 했다.

## 무엇이 틀렸었나

`internal/hook/pre_tool.go` 의 `dangerousBashPatterns` 에 있던 6개 정규식:

    rm\s+-rf\s+/      rm\s+-rf\s+~      rm\s+-rf\s+\*
    rm\s+-rf\s+\.\*   rm\s+-rf\s+\.git\b   rm\s+-rf\s+node_modules\s*$

- **너무 좁다**: 플래그 뭉치를 리터럴 순서로 물어서 `-fr` 로 뒤집으면 그대로 통과.
- **너무 넓다**: 끝이 맨슬래시라 **모든 절대경로**가 사정권. 임시 디렉터리 정리가 차단됨.
- 그리고 명령 텍스트 전체를 훑어서, 위험 형태를 **문자열로 인용만 해도** 차단됨.

한쪽만 고치면 반대쪽이 나빠지는 구조라 정규식 폭을 조절하는 방향은 애초에 답이 아니다.

## 조치

정규식 6개를 걷어내고 **구조적 판정**으로 바꿨다 (`internal/hook/dangerous_removal.go`).

1. 명령을 셸처럼 세그먼트(`;` `&&` `||` `|` `$(` …)로 나누고 토큰화하며, 이때 따옴표는 **벗긴다**.
   각 세그먼트의 첫 낱말이 `rm` 일 때만 삭제 명령으로 본다.
   → `echo "rm -rf /"` 는 첫 낱말이 `echo` 이므로 데이터다. 오탐이 구조적으로 사라진다.
2. 판정은 **플래그가 아니라 대상 경로**로 한다. 보호 대상(루트 / 홈 / 최상위 한 칸 절대경로 /
   `.git` / `node_modules` / `*` 류)을 겨눈 삭제는 어떤 플래그가 실렸든 거부.
   → 뒤집을 순서 자체가 없어져서 우회가 사라진다. 플래그 파싱 코드는 판정에 아무것도
   기여하지 않아 들어내지 않고 처음부터 넣지 않았다.
3. 대상 판정을 "슬래시로 시작하는가"에서 "**얼마나 깊은가**"로 바꿨다. `/usr` 는 최상위라
   보호, `/tmp/build-123` 은 두 칸 아래라 통과.
4. 나머지 정규식 스캔에는 `substituteQuotedArguments`(브랜치 가드가 이미 쓰는 것)를 태워
   따옴표 구간을 자리표시자로 접는다. 저장소 안에서 두 가드의 따옴표 처리가 반대였던 것도
   이걸로 맞춰진다.

## Evidence

### 재현 (수리 전, RED)

    go test ./internal/hook/ -run TestDangerousRemoval -count=1

우회 방향 — 15형태가 `decision = ""`(통과):
`rm -fr /`, `rm -r -f /`, `rm -f -r /`, `rm -r /`, `rm -f /`,
`rm --recursive --force /`, `rm --force --recursive /`, `rm -fr ~`, `rm -fr $HOME`,
`rm -fr .git`, `rm -fr node_modules`, `rm -fr *`, 따옴표 씌운 루트 2형태,
`&&` / `;` 뒤에 붙인 2형태.

오탐 방향 — 6형태가 `decision = "deny"`:
`echo` 로 문자열만 출력하는 2형태, `printf` 로 SQL 문자열 출력, 그리고 임시경로 정리
`rm -rf /tmp/moai-test-123`, `rm -rf /var/folders/…/build`, `rm -rf ~/go/pkg/mod/cache/tmp`.

**세션 내 실기 관측 — 거절 2건, 서로 다른 가드, 문면으로 귀속됨**

작업 중 두 번 거절당했고 문면이 서로 다르다. 둘을 구분해 적는다.

1. 이 테스트 파일을 heredoc 으로 쓰려던 명령이 거절됐다. 문면:
   `Dangerous command blocked: (?i)DROP\s+DATABASE`
   이슈 본문의 문제 3(가드가 자기 자신에 관한 기록을 막는다)이 수리 작업 중에 그대로
   재현된 것이다.
2. 뮤테이션 4건을 한 스크립트로 묶어 돌리려던 명령이 거절됐다. 문면:
   `this command is too complex to verify that it stays inside the worktree`
   이건 워크트리 격리 가드이며 t287 계열이다. 뮤테이션은 단순 명령으로 쪼개 돌렸다.

**귀속은 미확정이 아니다.** 두 문면은 각각 유일한 발신처를 갖는다:

    grep -rn 'Dangerous command blocked' --include='*.go' internal/ | grep -v _test.go
      → internal/hook/pre_tool.go:938, 951   (이 카드가 고친 가드)
    grep -rn 'too complex to verify' --include='*.go' internal/ pkg/ cmd/ | wc -l
      → 0                                     (moai 소스 아님 = Claude Code 바이너리 소관)

`Dangerous command blocked: %s` 는 `pre_tool.go:951` 의 포맷이고, `%s` 자리에 들어간
`(?i)DROP\s+DATABASE` 는 `compilePatterns` 가 `(?i)` 를 붙여 컴파일한 패턴의 `String()`
이다. 즉 1번은 **moai 자신의 pre-tool 가드**가 낸 것이고, 2번과는 다른 가드다.

수리 후에도 이 세션에서는 1번이 계속 재현된다 — 설치본 바이너리가 아직 옛것이라
반영은 재빌드·재설치 이후이며, 그건 여기서 재지 않았다(Gaps 참조).

### 수리 후 (GREEN)

    go test ./internal/hook/ -run TestDangerousRemoval -count=1
    ok  github.com/modu-ai/moai-adk/internal/hook  0.99s

    go vet ./internal/hook/            → rc 0
    gofmt -l <touched 3 files>         → 출력 없음
    go test ./internal/hook/ -count=1  → ok  31.151s   (패키지 전수, 회귀 0)

### 양방향 뮤테이션 — 4/4 적발

가드가 공허하지 않음을 보이기 위해 결함을 **양쪽 방향으로** 심고 각각 붉어지는지 봤다.
매 회 원본 복원 후 다음 뮤턴트를 심었고, 마지막에 `grep -rn MUTANT internal/hook/` = 0 확인.

| 뮤턴트 | 방향 | 결과 |
|---|---|---|
| M1 `-rf` 리터럴 포함일 때만 판정 (원 결함 복원) | 우회 | RED — 15형태 적발 |
| M2 토큰화에서 따옴표 안 벗김 (`strings.Fields`) | 우회 | RED — 따옴표 씌운 루트/홈 4형태 적발 |
| M3 대상 판정을 "모든 절대경로"로 되돌림 (원 과폭 복원) | 오탐 | RED — 임시경로 정리 3형태 적발 |
| M4 정규식 스캔 전 따옴표 접기 제거 | 오탐 | RED — 인용된 SQL 문자열 적발 |

M1 과 M3 은 각각 원 결함을 그대로 복원한 뮤턴트다. 즉 이 테스트는 이슈 #1658 이 다시
들어오면 반드시 붉어진다.

### 완화가 보안 축을 깎지 않았음

따옴표 접기는 허용을 넓히는 변경이므로, **여전히 무엇을 막는가**를 별도 테스트로 고정했다
(`TestDangerousRemoval_StillBlocksAfterQuoteFolding`). 보호 대상에 따옴표를 씌우는 것은
실행 경로가 되지 못한다 — `rm -rf "$HOME"`, `rm -rf '~'`, `rm "-rf" /`, 그리고 앞에 무해한
`echo` 를 붙인 복합 명령까지 전부 deny 다. M2 가 이 테스트를 붉히는 것이 그 증거다.

## Baseline-attribution

- 측정 트리: 워크트리 `.claude/worktrees/t286`, 기준 커밋 `1e5199b88` (develop) 위의 작업본
- 모든 수치는 이 트리에서 이 실행으로 잰 것이며, 다른 트리·시점에서 가져오지 않았다

## Gaps

- **CI 판정 없음** — 검증은 건드린 패키지(`internal/hook`)만 돌렸다. 전 패키지·크로스플랫폼
  판정은 CI 몫이며 이 시점에 관측되지 않았다.
- **실제 셸 발사 없음** — 이슈 보고자가 한 실물 `rm` 발사는 재현하지 않았다. 판정은 훅
  경로(`checkBashCommand`)에 payload 를 직접 먹여 받았다.
- **훅 바이너리 경로 미검증** — 설치된 `moai` 바이너리로의 반영은 재빌드·재설치 이후의
  일이며 여기서 재지 않았다.
- **이슈 제안 중 미이행** — 없음. 제안 1(플래그 파싱) 은 "플래그를 판정에서 제외"라는 더 강한
  형태로, 제안 2(경로 판정)와 3(따옴표 접기)은 그대로 이행했다.

## Residual-risk

- **따옴표 접기의 대가**: 접기는 `git push "--force" origin main` 처럼 **플래그를 따옴표로
  감싼** 형태를 정규식 스캔에서 놓치게 한다. 브랜치 가드가 이미 감수하고 문서화한 트레이드오프와
  같은 성질이며, 삭제 명령은 접기 이전에 구조적으로 판정하므로 이 경로에 걸리지 않는다.
- **셸 래퍼 우회**: `bash -c "rm -rf /"` 는 첫 낱말이 `bash` 라 데이터로 읽힌다. fail-open
  가드에서 난독화 형태를 덜 잡는 방향은 의도한 오차 방향이지만, 우회 경로로 남아 있다.
- **`node_modules` 보호 유지**: 원 정규식이 막던 대로 계속 막는다. 흔한 정상 명령이라 마찰이
  남지만, 완화 방향 변경은 카드 지시상 금지라 폭을 줄이지 않았다.
- **최상위 한 칸 규칙**: `/tmp` 자체는 보호, `/tmp/x` 는 통과. 최상위 디렉터리를 통째로 지우는
  것은 위험하다는 판단이며, 이 경계에 이견이 있으면 규칙 한 줄로 조정 가능하다.

## 인접 카드 (t287, lane-6) 파일 겹침

건드린 파일 3개:

- `internal/hook/dangerous_removal.go` (신규)
- `internal/hook/dangerous_removal_test.go` (신규)
- `internal/hook/pre_tool.go` (수정 +19/-9 — `dangerousBashPatterns` 6줄 제거,
  `checkBashCommand` 에 구조 판정 + 따옴표 접기 삽입)

t287 은 워크트리 가드(`internal/hook/branch_guard.go`) 소관이다. 나는 그 파일의
`substituteQuotedArguments` 를 **읽어서 재사용만** 했고 수정하지 않았다.
겹칠 수 있는 유일한 지점은 `internal/hook/pre_tool.go` 이며, 내 변경은 `checkBashCommand`
함수 하나에 국한된다. t287 이 같은 함수를 건드리면 리드의 순서 지정이 필요하다.
