# t558 판정 — GH #1675 는 폴백 프로필에서 재현되지 않는다

- 카드: t558 (리드 발행 2026-09-08, t546 이슈 스윕 §D)
- 트리: `.claude/worktrees/t558`, 브랜치 `WT-fallback-statusline`, base `d60b2956b` (로컬 develop)
- 측정 일자: 2026-09-10
- 바이너리: 이 워크트리에서 `make build` 로 만든 `bin/moai` (`BuildID=moai_cp/20260910_130400-18-gd60b2956b`)

## Claim

폴백 프로필(`CLAUDE_CONFIG_DIR=~/.moai/claude-profiles/__no_such_profile__`)에서도 statusline 은
프로젝트 `statusline.yaml` 을 **읽는다.** 껐을 때 gh 폴링은 일어나지 않고, 켰을 때는 일어난다 —
즉 프로필이 아니라 설정이 결정한다. GH #1675 가 보고한 "폴백에서 설정 무시 + gh 폴링 부활"은
이 트리에서 재현되지 않는다.

이미 `SPEC-STATUSLINE-PROFILE-RESPECT-001`(status `completed`, updated 2026-09-02,
`issue_number: 1675`)이 이 이슈를 다뤘다. 다만 그 SPEC 은 §E.4 Gaps 에 **"No integration render —
`moai statusline` 을 실제로 실행한 적이 없고, 어느 파일을 읽는지 추적한 것에 근거한다"** 고
스스로 적어뒀다. 이 카드가 채운 것이 정확히 그 구멍이다 — 런타임 실행 + 자식 프로세스 관측.

## Evidence

### 계기 설계

gh 호출을 세는 방식: `PATH` 앞에 가짜 `gh` 를 놓고 호출을 파일에 적게 했다.

```sh
#!/bin/sh
echo "gh $*" >> "$GH_CALL_LOG"
exit 1
```

픽스처는 스크래치패드의 `fx-sl` — `git init` 한 저장소에 `origin` 을
`https://github.com/modu-ai/moai-adk.git` 로 박고, `.moai/config/sections/statusline.yaml` 만
셀마다 바꿨다. stdin 은 Claude Code 가 넘기는 모양의 JSON(`workspace.current_dir` 포함)을 줬다.

### 네 셀

| 셀 | 프로필 | statusline.yaml | github 세그먼트 | gh 호출 |
|---|---|---|---|---|
| **P** (대조군) | 폴백 | `segments.github: true` | 미렌더(캐시 없음) | **2회** |
| A1 | 정상 | `segments.github: false` | 미렌더 | 0회 |
| A2 | **폴백** | `segments.github: false` | 미렌더 | 0회 |
| F | **폴백** | `forge: none` + `github: true` | 미렌더 | 0회 |

P 의 실제 호출 내역:

```
gh api rate_limit --jq .resources.graphql.remaining
gh repo view --json issues,pullRequests --jq .issues.totalCount, .pullRequests.totalCount
```

A1 과 A2 의 렌더 출력은 `diff` 로 바이트 동일했다. A2·F 는 20초 유계 대기 후에도 로그가 비어
있었다(대기 rc=124).

### 대조군이 왜 필수였는가 — 첫 측정은 공허했다

처음 잰 A/B 는 둘 다 "gh 0회" 였지만 **그 수치는 아무것도 증명하지 못했다.** 같은 조건의
대조군(P, `github: true`)도 0회를 찍었기 때문이다. 원인은 둘이었고 둘 다 계기 쪽 결함이다:

1. **stdin 부재.** 렌더의 github 블록 전체가 `if input != nil` 안에 있다
   (`internal/statusline/builder.go:257`). `< /dev/null` 로 돌리면 boardRoot 자체가 계산되지
   않아 스폰 지점에 도달조차 못 한다.
2. **자식이 비동기다.** 갱신은 `exec.Command(self, "statusline", "--refresh-github", ...)` 로
   분리 실행되고 즉시 `Release()` 된다(`github.go:174-183`). 스폰 직후 로그를 읽으면 아직
   비어 있다.

둘을 고친 뒤에야 P 가 2회를 찍었고, 그 시점부터 A2·F 의 0회가 의미를 갖는다. 대조군 없이
"0회 → 결함 없음" 이라고 적었다면 그것은 공허한 초록이었다.

### 코드 경로 확인 (카드 [HARD] 요구)

폴백 프로필 경로가 프로젝트 설정을 읽는지 코드로 확인했다. `statusline` 명령은
`findProjectRootFn()` → `project.FindProjectRoot()` 로 **CWD 에서 위로 올라가며** 프로젝트
루트를 찾고(`internal/core/project/root.go:20-26`), 그 루트로
`loadStatuslineFileConfig(projectRoot)` 가 `.moai/config/sections/statusline.yaml` 을 읽는다
(`internal/cli/statusline.go:92-95, 188-215`). 이 경로에 `CLAUDE_CONFIG_DIR` 도 프로필도
등장하지 않는다 — 설계상 프로필과 무관하다. 위 실측이 그 설계대로 동작함을 보인다.

폴링을 막는 게이트는 둘이고 서로 다른 지점이다:
- 세그먼트 게이트 — `if b.renderer.isSegmentEnabled(SegmentGitHub)` 가 렌더가 아니라 **스폰**을
  감싼다(`builder.go:269-271`). A1/A2 가 이 게이트를 탄다.
- forge 게이트 — `forgeOverride` 가 forge 아닌 값을 내면 스폰 전에 반환한다
  (`github.go:154-160`). F 가 이 게이트를 탄다.

### 회귀 스위트

```
$ go test ./internal/statusline/...                                   ok  30.685s
$ go test ./internal/statusline/ -run 'Forge|SpawnGate|Suppress|GitHub' -v -count=1 \
    | grep -c '^--- PASS'                                             → 38
```

셀렉터 0매치를 배제하려고 `--- PASS` 를 셌다: 38건이 실제로 돌았다.

## Baseline-attribution

- 트리: `d60b2956b` (로컬 develop, 이 워크트리 HEAD). 측정 전 추적 파일 수정 0.
- 바이너리: 위 트리에서 이번에 빌드한 것 — 설치본이 아니다.
- 모든 수치는 이 실행에서 관측했다. SPEC progress.md 나 t293 보고서에서 옮겨온 값이 아니며,
  SPEC 의 `completed` 표기는 단서로만 썼다.

## Gaps

- **제보자의 실제 환경(`~/MoAI/mo.ai.kr`)에서 재현하지 않았다.** 합성 픽스처를 썼다. 그 저장소
  고유의 설정 배치(워크트리 vs primary 의 `statusline.yaml` 분기 — SPEC §E.4 가 F3 로 남긴
  레버 비대칭)가 원인이었다면 이 픽스처는 그것을 재현하지 못한다.
- **`__no_such_profile__` 이 애초에 왜 생기는지는 건드리지 않았다.** 이슈 §기대동작 3번 항목이자
  SPEC 이 Out of Scope 로 명시한 축이다. 이 카드는 "폴백 상태에서 설정이 존중되는가"만 답한다.
- **v3.1.2 에서의 재현은 시도하지 않았다.** 제보 버전이며, 수리는 그 이후에 들어갔다.
- **`0/0` 표시 증상은 관측하지 않았다.** 캐시가 없으면 세그먼트가 통째로 미렌더라 이 픽스처에서는
  나타나지 않는다. SPEC §E.4 도 이 증상은 "우회했지 고치지 않았다" 고 적어뒀다.
- 전체 스위트는 로컬에서 돌리지 않았다(§4.1 규율). 코드 변경 0.

## Residual-risk

- **레버 비대칭(SPEC F3)은 그대로다.** 워크트리의 `statusline.yaml` 에 `forge: none` 을 써도
  아무 오류 없이 아무 효과가 없는 경우가 있다. 이 픽스처는 단일 디렉터리라 그 비대칭을 건드리지
  않는다 — 제보자가 워크트리에서 설정을 편집했다면 증상이 되살아날 수 있다.
- **opt-out 은 `moai update` 에 휘발한다.** `.moai/config` 는 `CleanMoaiManagedPaths` 가 통째로
  지우는 뿌리다. 사용자가 껐다고 믿는 폴링이 업데이트 한 번에 되살아나는 경로가 남아 있고,
  이것은 이 카드가 아니라 update 축의 문제다.

## 권고

1. 카드 t558 은 **코드 변경 없이** 종결한다. 이 카드의 산출물은 SPEC 이 남긴 "런타임 미검증"
   Gap 을 닫은 실측이다.
2. GH #1675 는 수리 완료로 회신하고 출시 때 종결한다(제보 버전 v3.1.2 에는 미출시). 회신에
   위 네 셀 표를 인용하고, **워크트리에서 편집한 `statusline.yaml` 은 아직 비대칭이 있으니
   primary 쪽 설정을 확인해달라**는 단서를 함께 남긴다 — 제보자가 여전히 증상을 본다면 그
   경로일 가능성이 가장 높다.
3. 별도 카드 후보 2건: (a) 레버 비대칭 F3 를 오류로 드러내기, (b) `moai update` 가 opt-out 을
   지우는 경로. 둘 다 이 카드 범위 밖이다.
