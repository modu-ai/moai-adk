# 배치 착지 모양 측정, 그리고 릴리스 태그 안전장치

> 신설 2026-10-17 (card t1833, 운영자 지시 P2). 이 문서는 **로컬 운용 지침**이며 템플릿에 미러하지 않는다.
> 대상은 두 가지다. 로컬 `main`과 origin 배치 병합이 어떤 모양으로 쌓이는지에 대한 측정, 그리고 릴리스 태그를
> CI가 통과한 커밋에만 얹는 확인 단계. 근거 자료: GoReleaser 빠른 시작, Fowler의 release-ready mainline.

> **철회 안내.** 이 카드는 감사 세 번(0.62, 0.76, 0.75)을 거치는 동안 `moai integration bisect` 하위 명령에서 매번 새 결함이 나와
> 범위를 줄였다. 이 하위 명령과 이를 쓰는 원인 카드 탐색 절차는 이 카드에서 철회했고, 재설계한 후속 카드로 돌아온다.
> 감사 지적 `.moai/reports/t1833/sync-audit-iter3.md`의 N7부터 N14까지가 그 후속 카드의 설계 입력이며, 철회한 코드는
> 커밋 `c78e02e68`, `e86799369`, `7c5b5cfc5`에 남아 있다. 아래 내용은 철회와 무관하게 유효한 두 부분만 담는다.

## 1. 착지 모양 (측정)

카드 단위로 원인을 좁히려면 first-parent 한 걸음이 카드 하나여야 한다. 이 저장소의 로컬 `main`이 실제로
그런 모양인지 2026-10-17 기준 트리 `fb749d372`에서 측정했다.

명령: `git log --first-parent --format='%P|%s' main -n <N>` 결과를 부모 수와 제목의 `(card t<번호>`로 분류.

| 창 | 두 부모 + 카드 id | 두 부모, 카드 id 없음 | 직접 커밋(부모 1) |
|---|---|---|---|
| 최근 40개 | 40 | 0 | 0 |
| 최근 100개 (2026-10-07까지) | 64 | 4 | 32 |

- 최근 40개는 전부 카드 단위 `--no-ff` 병합이다. 제목 예:
  `Merge WT-10-15-release-3 into main (card t1760, integration merge)`.
- 100개 창의 나머지는 로컬 main 통합 흐름이 자리 잡기 전의 기록이다. 직접 커밋 32개는 GitHub Flow 시기의
  PR 스쿼시 커밋(`... (#1815)`)이 대부분이고, 부모 하나짜리 `chore` 커밋 3개(`chore(config)` 설정 커밋 2개와 `chore: commit the primary checkout's pending changes onto main` 1개)가 섞여 있다. 카드 id 없는
  두 부모 병합 4개는 PR 병합 3개와 카드 id 없이 쓴 통합 병합 1개다.
- 코드 근거: `internal/factory/integration_merge_step.go`가 병합 제목을
  `Merge <카드 브랜치> into <대상> (card <id>, integration merge)`로 쓰고 `git merge --no-ff`로 병합한다.
  카드 id가 제목에 실리는 것은 이 한 곳의 형식 덕분이다.

**배치가 origin에 들어가는 모양.** 로컬 `main`을 `release/main-batch-YYYYMMDD`로 푸시하고, 그 PR을 병합
커밋으로 병합한다(스쿼시 금지, `AGENTS.local.md` §4.0 §6). 측정 사례: 병합 커밋 `6dde1c1c1`
(`Merge pull request #1826 ...`)에서 `M^1`은 `M^2`의 조상이고, `M`의 트리는 `M^2`의 트리와 같다.
`M^1..M^2`를 first-parent로 걸으면 63단위가 나온다. 따라서 **origin/main을 first-parent로 걸으면 배치
하나가 한 걸음이고, 카드 단위는 `M^2` 쪽 사슬 안에 있다.**

## 2. 릴리스 태그 안전장치

푸시된 `v*` 태그는 지우거나 옮길 수 없다. 그래서 `scripts/release.sh`는 태그를 만들기 전에 **태그할 커밋의
origin CI 결과**를 확인한다(검증 9).

- 읽는 곳: `gh api --paginate repos/modu-ai/moai-adk/commits/<sha>/check-runs`. 검증 6이 HEAD가
  origin의 끝과 같음을 이미 보장하므로 `<sha>`는 태그될 커밋이다.
- 통과 조건: check run이 하나 이상 있고, 전부 `completed`이며, 전부 `success`, `neutral`, `skipped` 중
  하나로 끝났고, 이 저장소 CI 워크플로(`.github/workflows/ci.yml`, `name: CI`)가 게시하는 check run 세 개가
  이름까지 일치하고 `github-actions` 앱이 게시한 것으로 모두 있다: `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`. 이름 목록은 `.github/required-checks.yml`의
  `release_gate` 목록 한 곳이고(스크립트는 이름을 직접 갖지 않고 이 파일을 읽는다), 잡 이름이 바뀌면 그 목록만 고친다.
  파일이 없거나 이름을 하나도 읽지 못하면 스크립트는 태그를 만들지 않고 거부한다. 다른 앱의 녹색 check run
  하나만으로는, 그리고 필수 이름을 쓰는 다른 앱의 check run으로는 통과하지 못한다(`app.slug`를 읽어 비교한다. 앱 정보가 없는 check run도 세지 않는다).
- 거부(종료 코드 1, 조건 이름을 출력): check run 없음, 진행 중인 것이 있음, 실패한 것이 있음, 읽기 실패,
  `gh` 없음, CI 워크플로의 check run이 빠져 있음(없는 이름을 나열한다). 실패와 진행 중이 함께 있으면
  실패로 보고한다.
- 우회: `--skip-ci-check` 하나뿐이다. 우회하면 실행 중 `CI gate OVERRIDDEN` 경고가 나오고, 릴리스 요약에
  `CI gate:     OVERRIDDEN (--skip-ci-check)`가 남는다.
- 확인 프롬프트 뒤 `git tag` 직전에 `HEAD`가 확인한 커밋 그대로인지 한 번 더 확인하고, 태그는
  `git tag -a <태그> -F <노트> <확인한 커밋>`처럼 그 커밋을 직접 지정해 만든다. 확인과 태그 사이에 `HEAD`가
  움직여도 태그는 확인한 커밋에 붙는다.

이전 검증은 결합 상태 엔드포인트(`commits/<sha>/status`)를 읽었다. 그 엔드포인트는 커밋 상태(status)만
집계하고 Actions의 check run은 집계하지 않는 것이 GitHub 문서의 정의다. CodeRabbit 같은 외부 상태 하나가
녹색이면 Actions가 붉어도 `success`가 나올 수 있었다. 또 `pending`은 확인 프롬프트로 통과시켰고 `gh`가 없으면
경고만 하고 넘어갔다.

**GoReleaser는 이 확인을 맡을 수 없다.** `.goreleaser.yml`은 이미 푸시된 태그를 입력으로 받아 빌드하는
설정이라 태그를 만들거나 거부하는 지점이 없다. 태그를 거부할 수 있는 마지막 지점은 `scripts/release.sh`다.
`release.yml`의 provenance 게이트는 태그가 생긴 뒤에 돌아서 게시를 막을 수는 있어도 태그를 되돌리지는
못한다. 손으로 만든 태그는 `release.sh`를 거치지 않으므로 이 확인을 지나지 않는다(후속 카드 후보).

## 3. 한계 (이 확인이 보지 못하는 것)

1. **태그 확인의 범위.** check run 전부를 요구하므로 보조 점검의 실패도 태그를 막는다. 필수 점검만 요구하도록
   좁히지 않은 이유는, 태그가 되돌릴 수 없는 쪽이라 과잉 거부가 과잉 허용보다 싸기 때문이다. 상태(status)로
   보고하는 외부 점검(CodeRabbit)은 이 확인이 읽지 않는다. 이름으로 요구하는 세 개는 `ci.yml`이 push에서
   게시하는 잡 이름이라 `Guard Bundle`, `Analyze (Go) (go)`, 다른 `Build` 행렬 칸은 이름으로는 요구하지 않는다
   (실패하면 전부 요구 조건에 걸려 막는다). `release_gate`와 실제 게시 이름이 어긋나면 안전한 쪽으로,
   즉 태그를 거부하는 쪽으로 실패한다.
2. **GitHub 응답 형식과 이름의 실측.** 이 확인의 jq 식과 `release_gate`의 세 이름은 가짜 `gh`와
   `ci.yml`·`.github/required-checks.yml` 읽기로만 정했다. 실제 GitHub가 main push 커밋에 게시하는 check run
   이름은 원격에 접속해야 하므로 이 카드에서 확인하지 않았다. 다음 실제 릴리스의 `--dry-run`이 첫 실측이다.
   이름이 다르면 그 실행이 `expected check run(s) not reported`로 거부하므로 거짓 통과는 나지 않는다.
   `app.slug` 값 `github-actions`도 같은 방식으로 실측하지 않았다.
3. **check run 이름 파싱.** 스크립트는 check run 이름을 탭으로 나눈 행으로 읽어 필수 이름과 비교한다(`scripts/release.sh`).
   이름에 탭이나 줄바꿈이 든 check run은 행을 위조할 수 있다(sync-audit-iter3 N10). GitHub가 앱에게 그런 이름을 허용하는지는
   원격에 접속하지 않아 확인하지 않았다. 고치는 방법은 판정을 jq의 JSON 객체에서 직접 만드는 것이며 후속으로 미뤘다.

## 4. 검증 근거

- 릴리스 시험: `internal/template/release_ci_gate_test.go`. 가짜 `gh`로 녹색, 진행 중, 실패, 없음, 읽기 실패,
  `gh` 없음, 우회, 다른 앱의 녹색 check run만 있는 경우, CI 워크플로 check run 일부가 빠진 경우, 필수 이름이지만 다른 앱이 게시한 경우,
  앱 정보가 없는 경우를 시험하고,
  확인 뒤 `HEAD`가 움직여도 태그가 확인한 커밋에 붙는지를 `git` 대역으로 시험한다. 증거 기록은 `.moai/reports/t1833/progress.md`.
