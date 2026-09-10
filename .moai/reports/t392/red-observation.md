# t392 — B.3 Gap 닫음: golden 스탬프 사이트의 실행 관측

`baseline.md` §B.3 이 미관측으로 남긴 것을 실행으로 관측했다. baseline 은 그대로 두고
이 파일이 그 Gap 만 닫는다.

측정 트리: worktree `.claude/worktrees/t392`, HEAD `9a3e2dabe`
측정 시각: 2026-09-01

## R.1 셀렉터 0매치 함정을 먼저 밟았다 (기록)

첫 시도는 `-run 'TestStatus_Golden|TestDoctor_Golden'` 이었고 결과가
`ok ... [no tests to run]` 이었다. 초록이 아니라 **비실행**이다. 실제 함수명은
`TestStatus_Current_{Light,Dark}` · `TestStatus_NoColor` · `TestDoctorGolden_{Light,Dark,NoColor}`
로, `grep -n '^func Test'` 로 확인한 뒤 셀렉터를 고쳤다.

## R.2 기준선 — 트리 버전(v3.1.3)에서 6/6 통과

    go test ./internal/cli/ -count=1 -v \
      -run 'TestStatus_Current|TestStatus_NoColor|TestDoctorGolden_'

    --- PASS: TestDoctorGolden_Light / _Dark / _NoColor
    --- PASS: TestStatus_Current_Light / _Current_Dark / _NoColor
    ok  github.com/modu-ai/moai-adk/internal/cli  0.792s

## R.3 범프를 모사하면 6/6 실패

같은 트리, `version.Version` 만 v3.1.4 로 주입(golden 은 그대로):

    go test -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=v3.1.4" \
      ./internal/cli/ -count=1 -v \
      -run 'TestStatus_Current|TestStatus_NoColor|TestDoctorGolden_'

    --- FAIL: TestDoctorGolden_Light / _Dark / _NoColor
    --- FAIL: TestStatus_Current_Light / _Current_Dark / _NoColor
    FAIL  github.com/modu-ai/moai-adk/internal/cli  0.948s

여섯 파일 전부가 범프에 반응한다. **스탬프 사이트다** — 코드 판독이 아니라 실행으로 확인됐다.

## R.4 [운영 사안] `release/v3.1.4` 의 CI 가 지금 이 결함으로 빨간불이다

    gh run list --branch release/v3.1.4 --limit 5
      → CI                                  failure   head 26898312e
      → Release PR Multi-OS Verification     failure   head 26898312e
      → Test Installation Scripts / SPEC Lint / CodeQL  success

    gh run view 33382844571 --log-failed | grep -oE '\-\-\- FAIL: [A-Za-z_0-9]+' | sort -u
      → TestDoctorGolden_Dark / _Light / _NoColor
      → TestStatus_Current_Dark / _Current_Light / _NoColor

CI 로그에서 실패한 테스트는 **정확히 이 여섯 개뿐**이다. 다른 실패는 없다.

즉 준비된 v3.1.4 릴리스는 t392 가 탐지하려는 바로 그 누락 때문에 **현재 막혀 있다**.
`61921f1ba` 가 스탬프 7파일만 고치고 golden 6파일을 빠뜨린 결과다.

이 카드의 수리(테스트에서 버전 pin + golden 재생성)가 그대로 이 적색을 닫는다. 다만
릴리스 일정이 t392 착지를 기다릴 이유는 없으므로, 릴리스 쪽에서 먼저 golden 을 재생성해
푸는 선택지가 있다 — 그 경우 이 RED 픽스처가 사라지므로 **관측은 이 파일에 이미
고정되어 있다**(R.3 은 이 워크트리에서 언제든 재현 가능하고 릴리스 브랜치에 의존하지 않는다).

## R.5 [정정·범위] §R.4 의 「여섯 개뿐」은 CI **잡 하나**에 한정된 주장이다

초판 §R.4 는 「CI 로그에서 실패한 테스트는 정확히 이 여섯 개뿐이다. 다른 실패는 없다」로
적었다. 그 문장이 훑은 범위는 **run `33382844571` 하나**이고, 브랜치 전체가 아니다.

lead-1 이 PR 체크 전량을 다시 훑어 **두 번째 적색 계열**을 찾았다 — 별도 run
`33382844542`(Graph Freshness), 원인이 다르다:

    ##[error]codemaps provenance stamp 7fc0af324... is NOT an ancestor of PR base origin/main
             — orphan-bound stamp (a squash merge would strand it)

이 관측은 이 세션의 것이 아니라 lead-1 의 것이며, 카드 **t407** 로 발행됐다. 여기에는
출처를 밝혀 옮겨 적는다.

교훈: 부재 주장(「X 뿐이다」)은 훑은 범위를 함께 말하지 않으면 더 넓은 주장으로 읽힌다.

## R.6 golden diff 의 변수는 하나가 아니다 (t406 첫 단계)

`UPDATE_GOLDEN=1` 재생성이 답이라고 단정하면 안 된다. 같은 CI 로그(`/tmp` 아닌 재조회 시
`gh run view 33382844571 --log-failed`)에 **두 종류의 차이**가 함께 들어 있다:

1. **버전 문자열** — 여섯 테스트 전부에서 `moai-adk v3.1.4`(got) 대 `v3.1.3`(want).
   §R.3 이 이 축을 격리해 증명했다: 같은 트리에서 `version.Version` 만 주입해도 6/6 이
   뒤집힌다.
2. **doctor 집계 수치** — `5 ok, 0 warn` / `9 ok, 3 warn` / `3 ok, 8 warn`,
   `Pass 17 Warn 11` 류. lead-1 관측. 이 수치가 **환경 의존이면** 로컬에서 재생성한
   golden 이 CI 에서 또 갈린다.

즉 축이 최소 둘이고, 1번은 재생성으로 닫히지만 2번은 재생성이 오히려 **로컬 고유 값을
박아 넣을** 수 있다. t406 은 이 분류를 먼저 해야 한다.

## Gaps

- windows / darwin 러너에서의 동일 실패는 CI 로그에서 확인하지 않았다. ubuntu 잡 하나만 읽었다
- `Release PR Multi-OS Verification`(33382844521) 의 실패 원인은 귀속하지 않았다 — 같은
  golden 일 개연성이 높지만 관측하지 않았다
- R.6 의 2번(집계 수치)이 실제로 환경 의존인지는 **관측하지 않았다**. 로그에 차이가 있다는
  것까지만 안다 — 차이의 원인이 환경인지 트리인지는 t406 이 가려야 한다
