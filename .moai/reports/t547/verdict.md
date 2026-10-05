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

## 8. 창 재측정 (2026-09-18T01:10 KST — 리드 재배차)

### Claim

AC-JFM-018 은 **녹색이 아니다**. 동시에 이번 행들을 **유효한 반증으로도 읽을 수 없다** — 창이
오염돼 있다. 판정은 **gap(창 무효)** 이며 처분은 리드·운영자 몫이다.

### Evidence

| 확인 | 명령 | 관측 |
|---|---|---|
| 필터 양성 대조 | `collect-pull-window.sh --selftest` | `selftest PASS: fixture 3 rows (2 pull, 1 push) -> export rows=2 violations=1` |
| 전 트리 스캔 | primary + `.claude/worktrees/*` + `~/.moai/worktrees/*` 의 `askuser-observations.jsonl` 에 `jq 'select(.mode=="pull")'` | 로그 파일은 primary 에만 존재: `total=81 pull=5 viol=4 last=2026-09-17T16:07:14Z` |
| 구성 | `grep -n recommendation_mode` primary / `git show develop:` | 둘 다 `recommendation_mode: pull` |
| 질문 규칙 트리 | `git grep -n -i "recommendation_mode\|pull" <ref> -- .claude/rules/moai/core/askuser-protocol.md` | `develop`: 64·117·120·265행 pull 분기 존재 / `HEAD`(primary=`main`) 와 primary 워킹 사본: **0건** |

pull 행 5건(전부 창 앵커 `2026-09-13T14:07:40Z` 이후):

```
2026-09-13T17:17:35Z b6556473 label_present=false q=1 opt=3
2026-09-13T20:17:26Z 764312c6 label_present=true  q=4 opt=11
2026-09-13T20:19:53Z 41de9342 label_present=true  q=1 opt=3
2026-09-13T20:31:58Z 764312c6 label_present=true  q=4 opt=11
2026-09-17T16:07:14Z 1f4f7f28 label_present=true  q=2 opt=5
```

### Baseline-attribution

이 실행(2026-09-18 01:10 KST), primary 로그 81행과 위 두 ref 에 대해 직접 잰 값이다. §7 의 0행은
2026-09-13 23:11 측정이며 이번 값과 비교 대상일 뿐 재사용하지 않았다.

### 판독

1. **분모 부족**: n=5 < 20 — §3 에 따라 gap.
2. **오염 원인(관측)**: 관측기는 `mode` 를 질문 세션 트리의 `interview.yaml` 로 찍는다. primary
   체크아웃은 `main` 에 있고, 그 트리의 `askuser-protocol.md` 에는 pull 분기가 없어 `(권장)` 라벨을
   여전히 [HARD] 로 요구한다. 즉 primary 세션은 **구성은 pull, 로드된 질문 규칙은 push** 인 상태로
   질문했다. `label_present:true` 4건은 이 불일치와 정합한다 — 규칙을 어긴 증거가 아니라 규칙이
   pull 을 모르는 트리에서 나온 행이다.
3. 따라서 이 행들은 "pull 규칙 하에서의 준수/위반" 표본이 아니다. 계속 모아도 같은 오염이 쌓인다.

### Gaps

- 각 세션(`764312c6` 등)이 실제로 어느 트리의 규칙을 로드했는지는 세션 내부를 보지 못해 직접
  확인하지 않았다 — 로그가 primary 에만 있다는 사실과 primary 규칙 트리 판독으로 추론했다.
- 세션별 `calls_issued` 는 묻지 않았다(§5-3 은 n≥20 일 때만 수행).
- AC-JFM-023 상태는 재측정하지 않았다 — §6 플래그 그대로.
- export(`.moai/reports/t401/pull-window.jsonl`)·provenance 는 만들지 않았다 — 판정 가능한 창이
  아니므로 표본으로 오인될 산출물을 남기지 않는다.

### Residual-risk

- 추론 2가 틀렸다면(세션이 develop 규칙 트리를 로드했다면) 4건은 실제 위반이고 018 은 반증된다.
  이 가능성은 세션 트랜스크립트 확인 없이는 배제되지 않는다.

### 처분 선택지 (리드·운영자 결정 — 레인은 고르지 않는다)

- 창 재개시: pull 규칙이 있는 트리(develop 기반 워크트리)에서 도는 세션만 표본으로 인정하고, 그
  트리 로그를 수집 대상으로 삼는다. 앵커 시각은 그 결정 시점으로 다시 잡는다.
- 또는 primary 에서 pull 구성을 되돌려 오염 행 생성을 멈춘다.

## 9. 창 재개시 — 운영자 판정 (a) 적용 (2026-09-18T01:15 KST)

운영자 판정(리드 전달): **develop 기반 트리의 세션만 표본으로 인정하고 앵커를 다시 잡아 창을
재개시**한다. primary `interview.yaml` 은 건드리지 않는다. 이 절이 `judgment-criteria.md` §1 의
앵커와 표본 범위를 대체한다 — 분모 규칙(`mode=="pull"`, `question_type` 필터 없음)·N=20 하한·
export 금지 조건은 그대로다.

### 9.1 기존 pull 5행 제외

§8 의 pull 5행(`2026-09-13T17:17:35Z` ~ `2026-09-17T16:07:14Z`, primary 로그)은 **표본에서 제외**한다.
근거: 이 행들이 기록된 primary 체크아웃은 `main` 이고, 그 트리의 `askuser-protocol.md` 에는 pull
분기가 0건이다(§8 Evidence 4행). 구성은 pull 인데 로드된 질문 규칙은 push 였으므로 이 행들은
pull 규칙 하의 표본이 아니다.

### 9.2 새 앵커

**`2026-09-17T16:15:27Z` (= `2026-09-18T01:15:27+0900`)** — 운영자 판정 적용 시각, `date -u` 실측.
이 시각 이전 행은 어느 트리에서 나왔든 표본이 아니다.

### 9.3 표본 인정 조건 (세 가지 모두)

1. `mode == "pull"` 이고 `timestamp >= 2026-09-17T16:15:27Z`.
2. 행이 **primary 체크아웃이 아닌** 트리의 로그(`.claude/worktrees/*/` 또는 `~/.moai/worktrees/**/`
   아래 `.moai/logs/askuser-observations.jsonl`)에 있다. primary 로그는 primary 가 `main` 에 있는 한
   통째로 제외한다.
3. 그 트리의 `.claude/rules/moai/core/askuser-protocol.md` 에 pull 분기가 있다 — 판독식
   `grep -c recommendation_mode <tree>/.claude/rules/moai/core/askuser-protocol.md` ≥ 1.

트리 구분 방법: 행 자체가 아니라 **그 행이 담긴 로그 파일의 위치**로 가른다. 관측기는
`<projectRoot>/.moai/logs/` 에 기록하고 `projectRoot` 는 `CLAUDE_PROJECT_DIR`(없으면 cwd)다
(`develop:internal/hook/askuser_observer.go:136-151`, `path_resolve.go:66-71`).

### 9.4 재개시 시점 실측

| 확인 | 관측 |
|---|---|
| 워크트리 수(`askuser-protocol.md` 보유) | 294 |
| 그중 pull 분기 보유 | 195 (99개는 조건 3 불충족) |
| 워크트리 안 관측 로그 파일 | **0** — 재개시 시점 분모 0, 판정 대상 없음 |
| 이 카드 워크트리 규칙 | `grep -c recommendation_mode` = 4 (조건 3 충족) |

`collect-pull-window.sh` 는 아직 primary 를 스캔 대상에 포함하고 앵커 필터가 없다 — 조건 1·2 를
반영하기 전에는 그 스크립트의 READING 줄을 판정 근거로 쓰지 않는다(수리는 n≥20 판독 전에 한다).

### 9.5 발견 — 행에 트리 식별자가 없다 (구현은 별도 판정)

관측 행 스키마(`develop:internal/hook/askuser_observer.go:49-60`)는 `timestamp`·`session_id`·`mode`·
`label_present`·`option_count`·`question_count`·`payload_parsed`·`question_type` 뿐이며, **트리 경로·
브랜치·HEAD·로드된 규칙 판을 담지 않는다.** 그래서 9.3 은 파일 위치를 대리 지표로 쓰고, 다음 구멍이
남는다:

- **로드 규칙과 기록 위치의 불일치**: 기록 위치는 `CLAUDE_PROJECT_DIR` 로 정해지지만 세션이 로드한
  규칙은 세션 시작 시점 트리의 것이다. 세션 도중 워크트리를 옮기면 둘이 갈라질 수 있다(미실측).
- **시점 불일치**: 조건 3 은 판독 시점의 트리 규칙을 본다. 행이 기록될 때 그 트리가 pull 분기를
  이미 갖고 있었는지는 행에서 복원되지 않는다.
- **mode 는 구성만 반영**: `mode` 는 `interview.yaml` 판독값이지 로드된 규칙 판이 아니다 — §8 오염이
  바로 이 틈에서 생겼다.

이 구멍들을 닫으려면 행에 트리/규칙 식별 필드가 필요하다. 추가 여부와 방식은 이 카드가 정하지
않는다.

### 9.6 수집 스크립트 수리 (commit `3ba2dc6ac`) — Gaps

- 수리: primary 제외 · `timestamp >= 2026-09-17T16:15:27Z` · pull 분기 없는 트리 제외 · n<20 이면 export
  미작성 · 깊은 `find` → 트리 glob.
- selftest: `admitted rows=2 violations=1` PASS. 규칙 4종(rule2·rule3·rule1-anchor·rule1-mode)을 하나씩
  제거한 변이는 모두 `selftest FAIL: rows=3 violations=2`.
- **Gap**: 워크트리 로그 발견 경로(glob)는 양성 실측이 없다 — 수리 시점 대상 로그 0개
  (`scanned_trees=0`). selftest 는 트리별 필터만 검증한다. 첫 워크트리 로그가 생기면 `scanned_trees ≥ 1`
  로 발견 경로를 확인한 뒤에 READING 을 근거로 쓴다.
- AC-JFM-018 은 n≥20 판독 전까지 gap 이며 카드는 열린 채로 둔다.
