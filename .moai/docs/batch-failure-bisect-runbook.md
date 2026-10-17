# 배치 실패 시 카드 단위 원인 찾기, 그리고 릴리스 태그 안전장치

> 신설 2026-10-17 (card t1833, 운영자 지시 P2). 이 문서는 **로컬 운용 지침**이며 템플릿에 미러하지 않는다.
> 대상은 두 가지다. 배치가 CI에서 깨졌을 때 어느 카드가 깼는지 좁히는 절차, 그리고 릴리스 태그를
> CI가 통과한 커밋에만 얹는 확인 단계. 근거 자료: bors-ng 배치 이분 탐색, Google TAP 배치 분할
> (SWE at Google 23장), GoReleaser 빠른 시작, Fowler의 release-ready mainline.

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

## 2. 절차

### 2.1 양 끝 ref 고르기

| 상황 | `--bad` | `--good` |
|---|---|---|
| origin/main의 배치 PR 병합 `M`이 CI에서 깨졌다 | `M^2` | `M^1` |
| 로컬 main이 어느 시점부터 깨졌다 | `main` | 마지막으로 녹색이던 착지 |

[HARD] 양 끝 ref에서 같은 시험 명령을 먼저 한 번씩 돌려 `bad`는 실패, `good`은 통과를 눈으로 확인한다.
이 하위 명령은 양 끝 ref를 검증하지 않는다(§4 한계 1).

### 2.2 실행

카드 worktree 안에서 실행한다(주 체크아웃에서는 거부된다. bisect는 HEAD를 옮기기 때문이다).

```bash
moai integration bisect --good <ref> [--bad <ref>] [--dry-run] -- <시험 명령>
```

`--bad`의 기본값은 `HEAD`다. `--dry-run`은 정확히 실행될 git 명령 세 줄을 출력하고 HEAD를 옮기지 않는다.
이 하위 명령이 하는 일은 아래 git 명령과 같다. 더하는 것은 거부 조건과 마무리뿐이다.

```bash
git bisect start --first-parent <bad> <good>
git bisect run sh -c '<시험 명령>'
git bisect reset
```

하위 명령이 거부하는 경우: 주 체크아웃, 커밋 안 된 변경이 있는 트리, 이미 진행 중인 bisect, 해석되지 않는 ref.
성공이든 실패든 끝에서 `git bisect reset`을 실행하고, HEAD가 시작 전의 브랜치 또는 커밋으로 돌아왔는지와
bisect가 끝났는지를 확인한다. 돌아오지 못했으면 종료 코드 1로 실패하고 오류에 이유와 수동 복구 방법을 적는다.
`git bisect start`가 도중에 실패해 HEAD가 옮겨졌을 때도 같은 복원과 확인을 한다.
원인 카드는 git이 남기는 bisect 기록(`git bisect log`의 `# first bad commit:` 줄)에서만 읽는다. 시험 명령이
무엇을 출력하든 결과에 영향을 주지 않는다. `git bisect`의 세 단계는 `GIT_DIR`, `GIT_WORK_TREE` 같은
저장소 지정 환경 변수를 뺀 환경에서 돌고, 사전 확인과 같은 트리를 본다. 시험 명령은 그 변수들을 뺀
나머지 환경을 그대로 받는다.

**시험 명령 고르기.** 실패한 CI 잡이 돌린 테스트 하나를 그대로 쓴다(`go test ./internal/x -run '^TestY$'`).
저장소 전체 테스트는 돌리지 않는다. 종료 코드는 git 규약을 따른다. 0은 good, 1-124는 bad,
125는 건너뜀(이 커밋은 빌드가 안 된다), 126·127과 128 이상은 탐색 중단이다. 실측(git 2.54.0):
`git bisect run sh -c nonexistent-tool`은 `error: bogus exit code 127 for good revision`으로 멈춘다.
시험 도구가 없거나(127) 실행할 수 없으면(126) 커밋을 bad로 치지 않고 탐색 자체가 중단된다.

### 2.3 결과 읽기

출력 형식 예시(제목은 실제 로그의 것이고, 결과 자체는 가정이다).

```text
integration bisect: first bad first-parent commit 4a39d0e42
  subject: Merge WT-card-branch-slug into main (card t1713, integration merge)
  card: t1713
```

- `card:` 줄이 원인 카드다. 그 카드 병합이 처음으로 시험을 깬 first-parent 단위다.
- `card: none in the subject`는 직접 커밋이거나 배치 조립 단계다. 예를 들어 배치 브랜치에서 origin/main을
  흡수하는 병합(`merge: absorb origin/main into the batch landing branch`)이나 배치 수리 커밋이 여기에 걸린다.
  이 경우 원인은 카드가 아니라 조립이다.
- 탐색이 수렴하지 않으면(`did not converge`) 이름을 대지 않고 실패한다. 시험 명령이 126·127(도구 없음 포함)
  또는 128 이상으로 끝났거나, 모든 후보가 125로 건너뛰어졌거나, 양 끝 ref가 변화를 사이에 두고 있지 않다는
  뜻이다.

### 2.4 카드 안쪽으로 한 단계 더

first-parent 탐색은 카드 병합 하나를 지목한다. 그 카드의 어느 커밋이 깼는지는 병합 `C`의 두 부모로 다시
좁힌다.

```bash
git bisect start <C>^2 <C>^1     # first-parent 없이 카드 브랜치 안쪽을 걷는다
git bisect run sh -c '<시험 명령>'
git bisect reset
```

이 단계는 하위 명령이 감싸지 않는다. 필요한 일이 드물고 git 명령 세 줄로 충분하기 때문이다.

## 3. 릴리스 태그 안전장치

푸시된 `v*` 태그는 지우거나 옮길 수 없다. 그래서 `scripts/release.sh`는 태그를 만들기 전에 **태그할 커밋의
origin CI 결과**를 확인한다(검증 9).

- 읽는 곳: `gh api --paginate repos/modu-ai/moai-adk/commits/<sha>/check-runs`. 검증 6이 HEAD가
  origin의 끝과 같음을 이미 보장하므로 `<sha>`는 태그될 커밋이다.
- 통과 조건: check run이 하나 이상 있고, 전부 `completed`이며, 전부 `success`, `neutral`, `skipped` 중
  하나로 끝났고, 이 저장소 CI 워크플로(`.github/workflows/ci.yml`, `name: CI`)가 게시하는 check run 세 개가
  이름까지 일치해 모두 있다: `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`. 이름 목록은 스크립트
  검증 9 맨 위의 `CI_REQUIRED_CHECKS` 한 곳이고, 잡 이름이 바뀌면 그 한 줄을 고친다. 다른 앱의 녹색 check run
  하나만으로는 통과하지 못한다.
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

## 4. 한계 (이 방법이 보지 못하는 것)

1. **양 끝 ref 가정.** git은 `bad`를 실패, `good`을 통과로 가정한다. 둘 중 하나가 틀리면 엉뚱한 단위를 지목하거나
   수렴하지 않는다. §2.1의 사전 확인이 이 한계의 유일한 방어선이다.
2. **조합 실패.** 카드 둘이 각각은 멀쩡하고 합칠 때만 깨지면, bisect는 조합을 완성한 두 번째 카드를 지목한다.
   잘못이 그 카드에만 있다는 뜻은 아니다.
3. **불안정한 시험.** 결정적이지 않은 시험은 어느 단계에서든 방향을 틀 수 있다. 의심되면 같은 단위에서 시험을
   반복해 본다.
4. **CI에서만 나는 실패.** darwin/windows 매트릭스, 깨끗한 환경, 시간 의존 실패는 이 머신에서 재현되지
   않는다. 재현이 안 되면 "로컬에서 재현 안 됨"이 결과이고, 카드를 지목하지 않는다.
5. **first-parent가 가리는 것.** 한 단위는 카드 병합 하나다. 카드 안쪽 커밋은 §2.4로 따로 좁힌다.
6. **카드가 아닌 단위.** 직접 커밋과 배치 조립 커밋은 `card: none`으로 보고된다.
7. **빌드가 안 되는 단위.** 시험 명령이 125로 끝나게 해서 건너뛴다. 건너뛴 단위가 많으면 결과의 정밀도가
   떨어진다.
8. **태그 확인의 범위.** check run 전부를 요구하므로 보조 점검의 실패도 태그를 막는다. 필수 점검만 요구하도록
   좁히지 않은 이유는, 태그가 되돌릴 수 없는 쪽이라 과잉 거부가 과잉 허용보다 싸기 때문이다. 상태(status)로
   보고하는 외부 점검(CodeRabbit)은 이 확인이 읽지 않는다. 이름으로 요구하는 세 개는 `ci.yml`이 push에서
   게시하는 잡 이름이라 `Guard Bundle`, `Analyze (Go) (go)`, 다른 `Build` 행렬 칸은 이름으로는 요구하지 않는다
   (실패하면 전부 요구 조건에 걸려 막는다). `CI_REQUIRED_CHECKS`와 실제 게시 이름이 어긋나면 안전한 쪽으로,
   즉 태그를 거부하는 쪽으로 실패한다.
9. **GitHub 응답 형식과 이름의 실측.** 이 확인의 jq 식과 `CI_REQUIRED_CHECKS`의 세 이름은 가짜 `gh`와
   `ci.yml`·`.github/required-checks.yml` 읽기로만 정했다. 실제 GitHub가 main push 커밋에 게시하는 check run
   이름은 원격에 접속해야 하므로 이 카드에서 확인하지 않았다. 다음 실제 릴리스의 `--dry-run`이 첫 실측이다.
   이름이 다르면 그 실행이 `expected check run(s) not reported`로 거부하므로 거짓 통과는 나지 않는다.

## 5. 검증 근거

- 하위 명령 시험: `internal/cli/integration_bisect_test.go`. 실제 git 저장소에 카드 병합 5개를 만들고(내부 커밋 2개씩),
  세 번째가 깨는 상황에서 first-parent 병합 제목과 `card: t1003`이 출력되고 HEAD가 복원됨을 확인한다. 주
  체크아웃, 더러운 트리, 진행 중인 bisect, 중단된 탐색, `--dry-run`도 각각 시험한다.
- 릴리스 시험: `internal/template/release_ci_gate_test.go`. 가짜 `gh`로 녹색, 진행 중, 실패, 없음, 읽기 실패,
  `gh` 없음, 우회, 다른 앱의 녹색 check run만 있는 경우, CI 워크플로 check run 일부가 빠진 경우를 시험하고,
  확인 뒤 `HEAD`가 움직여도 태그가 확인한 커밋에 붙는지를 `git` 대역으로 시험한다. 증거 기록은 `.moai/reports/t1833/progress.md`.
