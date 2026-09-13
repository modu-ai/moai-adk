# t547 판정서 — 착수 전 확인: 진입 조건 미충족으로 수집을 시작하지 않음

- 카드: t547 (SPEC-JUDGMENT-FIRST-MODE-001 M6 · AC-JFM-018, Class B)
- 트리: `.claude/worktrees/t547`, 브랜치 `WT-pull-falsifier-window`, 기반 `bdd5afc95`(두 번째 부모 = 로컬 develop `c7dd269f3`)
- 측정 시점: 2026-09-10, lane-6

## 1. 주장 (Claim)

1. 카드의 [HARD] 진입 조건이 충족되지 않았다. 카드는 "primary 체크아웃의 `interview.yaml` 에서 `grep recommendation_mode` 가 `pull` 을 보여줄 때만 착수한다"고 정한다. 지금 primary 의 `interview.yaml` 에는 이 키가 없다.
2. 따라서 이 카드는 수집·export·판정을 시작하지 않았다. `pull-window.jsonl` 도, 그 provenance 기록도 만들지 않았다. 빈 분모로 `violations: 0` 을 만드는 일을 피하려는 것이다.
3. 이 상태는 develop 병합만으로는 풀리지 않는다. primary 체크아웃은 `main` 에 체크아웃돼 있고, `pull` 키는 develop 에만 있으며 `main` 에는 없다. 관측기는 질문하는 세션의 프로젝트 루트 설정을 읽으므로, primary 에서 도는 세션은 지금도 `push` 로 기록한다.

## 2. 증거 (Evidence)

| 확인 | 명령 / 방법 | 결과 | 증거 파일 |
|---|---|---|---|
| primary 설정 | `grep -n recommendation_mode <primary>/.moai/config/sections/interview.yaml` | 출력 없음, EXIT=1 (파일은 있고 키가 없음) | `entry-check-primary-config.txt` |
| 브랜치별 설정 (대조) | develop 사본과 `origin/main` 사본에 같은 grep | develop 사본 6행 `recommendation_mode: pull` 이 잡힘, `main` 사본은 잡히지 않음 | `entry-check-branch-configs.txt` |
| primary 체크아웃 위치 | primary 의 `.git/HEAD` 판독, 이 트리에서 `git rev-parse main` | `ref: refs/heads/main`, `main` = `2213871af` | (이 판정서에 기록) |
| 관측기 로그 위치 | primary 와 `~/.moai/worktrees` 아래 `askuser-observations.jsonl` 검색(깊이 7) | primary 의 파일 하나만 있다. 검색이 존재하는 대상을 잡는다는 대조는 그 파일 자체다 | `entry-check-observer-files.txt` |
| 관측기 로그 내용 | JSONL 행 집계 | 46행, 전부 `mode: push` · `label_present: true`, `pull` 행 0, session_id 18개, 2026-09-03T19:19:29Z ~ 2026-09-10T09:10:37Z, 2026-09-08 이후 행 13개(전부 push) | `entry-check-observer-log.txt` |
| 모드 결정 방식 | develop 의 `internal/hook/askuser_observer.go` `resolveRecommendationMode` 판독 | 질문하는 세션의 프로젝트 루트에서 `interview.yaml` 을 읽고, 키가 없으면 `push` 로 정한다 | (코드 판독) |

2026-09-08 이후에도 push 행이 13개 쌓였다는 점은 1-3번 주장을 뒷받침한다. M3 에서 설정이 pull 로 바뀐 곳은 develop 이고, 실제로 질문이 기록되는 primary 체크아웃에는 반영되지 않았다.

## 3. 기준 귀속 (Baseline-attribution)

- 설정 확인은 primary 작업 트리의 파일을 이번 실행에서 직접 읽었다. develop·main 사본은 `c7dd269f3` 흡수 직후의 로컬 `develop` ref 와 `origin/main`(`2213871af`) 에서 `git show` 로 꺼냈다.
- 관측기 로그는 이번 실행에서 primary 의 원본 파일을 읽어 집계했다. export 하지 않았다.

## 4. 미검증 (Gaps)

- AC-JFM-018 은 여전히 RED 다. pull 분모 0 은 판정이 아니라 측정 불가 상태다.
- 질문하는 세션별 `calls_issued`(각 세션 트랜스크립트의 `AskUserQuestion` 호출 수)는 세지 않았다. 카드가 판정 시점에 세도록 정했고, 판정은 시작되지 않았다.
- AC-JFM-023(푸시 기준선 export 와 4-way 대조)도 이 카드 범위에 합류돼 있지만 착수하지 않았다. 기존 push 46행을 세션별로 나누고 `calls_issued` 와 대조하는 방식으로 진행할 수는 있다. 다만 이것이 AC-JFM-023 의 "관례가 들어오기 전 기준선" 조건을 만족하는지는 판단하지 않았다.
- primary 의 워크트리 밖 세션(다른 머신, 클라우드) 로그 존재 여부는 확인하지 않았다.

## 5. 잔여 위험 (Residual-risk)

- 풀어 가는 경로는 리드·운영자가 정할 몫이라 여기서 고르지 않는다. 확인된 사실 기준으로 가능한 경로는 셋이다.
  1. release 로 `pull` 키가 `main` 에 들어가고 primary 가 그것을 받은 뒤 창을 연다.
  2. 운영자가 primary 의 로컬 `interview.yaml` 에 `pull` 을 적용한다. 이 경우 `moai update` 가 `.moai/config` 를 통째로 다시 깔아 키가 사라질 수 있으니, update 뒤에 다시 적용해야 한다.
  3. AC-JFM-023 기준선 export 만 먼저 진행한다(위 Gaps 의 조건 판단이 선행돼야 함).
- 경로가 정해지기 전에 primary 에서 질문이 계속 쌓이면 push 행만 늘어난다. 이 행들은 pull 분모에 들어가지 않으므로 창이 열린 뒤의 표본에는 영향이 없다.

## 6. 전제 충족 기록 (2026-09-13 갱신 — lane 재배차)

§5 의 경로 2 가 운영자 결정으로 채택됐다.

| 확인 | 명령 / 방법 | 결과 |
|---|---|---|
| 적용 시각 | primary `interview.yaml` mtime | `2026-09-13T23:07:40+0900` |
| 키 존재 | `grep -n recommendation_mode <primary>/interview.yaml` (2026-09-13 23:09) | 6행 `recommendation_mode: pull` 적중 |
| 미커밋 로컬 | `git status --porcelain -- <primary>/interview.yaml` | ` M` — 커밋 안 된 로컬 수정 (moai update 뒤 재적용 대상) |
| 적용 주체 | 리드 전달(운영자 결정) | 리드가 수동 추가 — 레인은 리드 보고를 믿지 않고 위 세 행으로 독립 재측정 |

전제는 이제 **충족**이다. 이후 이 판정서의 1-3항은 창 개시 이전 상태의 기록으로 보존한다.

## 7. 수집 대기 상태 (2026-09-13 23:11 기준)

- 분모(`mode=="pull"`) 실측: **0행** — 전 트리 스캔(primary + `.claude/worktrees/*` +
  `~/.moai/worktrees/*`), 관측 로그 파일 자체가 primary 에만 존재(76행, 전부 push).
- 적용 뒤 시작한 새 세션이 아직 `AskUserQuestion` 을 발화하지 않았으므로 당장 판정 대상이 없다.
  → **수집 대기**로 보고하고 정지. 수집 도구(`collect-pull-window.sh`, `--selftest` 양성 대조
  포함)와 판정 기준(`judgment-criteria.md`)은 같은 커밋으로 확정했다.
- 선행 조건 플래그: AC-JFM-023 이 여전히 RED — pull 행이 20 을 넣어도 023 이 녹색이기 전에는
  018 판정이 서지 않는다(`judgment-criteria.md` §6). 023 처분은 리드·운영자 몫.
