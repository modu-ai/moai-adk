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

## Gaps

- windows / darwin 러너에서의 동일 실패는 CI 로그에서 확인하지 않았다. ubuntu 잡 하나만 읽었다
- `Release PR Multi-OS Verification`(33382844521) 의 실패 원인은 귀속하지 않았다 — 같은
  golden 일 개연성이 높지만 관측하지 않았다
