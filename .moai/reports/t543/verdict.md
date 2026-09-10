# t543 판정서 — 리터럴 base SHA 범위 AC 의 흡수 후 오탐: 재현, 그리고 카드가 제안한 수리도 현행 절차에서는 틀림

- 카드: t543 (재현 먼저, 설계 판단이 생기면 멈추고 보고)
- 트리: `.claude/worktrees/t543`, 브랜치 `WT-literal-base-scope`
- 측정: 2026-09-10, lane-6, 이 트리에서 실제로 흡수를 수행하며 세 시점(S0·S1·S2)을 쟀다.

## 1. 주장 (Claim)

1. **결함이 재현됐다.** Go 를 한 줄도 고치지 않은 카드가, 흡수 뒤에는 리터럴 base SHA 로 잰 범위 AC 에서 Go 변경 51개를 가진 것으로 판정된다.
2. **흡수 전에는 등가성이 성립한다.** 카드 커밋 뒤, 흡수 전(S1)에는 리터럴 SHA · `git merge-base origin/develop HEAD` · `git merge-base develop HEAD` 가 모두 같은 분기점을 낸다. t533 이 실측한 흡수 전 등가성과 같다.
3. **카드가 제안한 수리(`CARD_BASE=$(git merge-base origin/develop HEAD)`)도 현행 절차에서는 같은 오탐을 낸다.** 지금 레인 절차는 원격이 아니라 **로컬** `develop` 을 흡수한다(`CLAUDE.local.md` §4.1, develop 판 389행). 로컬 develop 이 원격보다 앞서 있으면, 흡수 뒤 원격 기준 merge-base 는 흡수 이전 분기점에 머문다. 그래서 범위에 다른 레인들의 미푸시 커밋이 그대로 남는다. 이번 측정에서 원격 기준 merge-base 로 잰 Go 목록은 리터럴 SHA 로 잰 목록과 같은 51개였다.
4. **흡수한 바로 그 ref 와의 merge-base 는 올바른 왼쪽 끝을 낸다.** `git merge-base develop HEAD` 는 흡수 뒤 흡수한 로컬 develop tip 을 내고, 그 범위는 카드 자신의 기여(파일 4개, Go 0)와 같았다.
5. 따라서 수리 방향에 설계 판단이 생긴다(§5). 카드 지시에 따라 여기서 멈추고 보고한다. 절차 문서도, 스윕 대상 AC 도 고치지 않았다.

## 2. 증거 (Evidence)

모든 파일은 `.moai/reports/t543/repro/` 에 있다. 리터럴 핀은 이 워크트리를 만든 시점의 HEAD `d060e0d13`(= `origin/develop`)이다.

| 시점 | 상태 | 리터럴 핀 전체 / Go | `merge-base origin/develop HEAD` | `merge-base develop HEAD` |
|---|---|---|---|---|
| S0 | 핀 = HEAD, 카드 커밋 없음 | 0 / 0 (`s0-literal-*`) — 대조군 0 이라 **측정 불가** | `d060e0d13` | `d060e0d13` |
| S1 | 카드 커밋 `a71faa685`(S0 증거 파일 4개, Go 없음) | 4 / 0 (`s1-literal-*`) | `d060e0d13` | `d060e0d13` |
| S2 | 로컬 develop `4b82591cb` 흡수, HEAD `6f3159d38` (`s2-head.txt`, HEAD^2 = `4b82591cb`) | **540 / 51** (`s2-literal-*`) | `d060e0d13` → Go **51**, 리터럴과 같은 집합 (`s2-proposal-origin-go.txt`) | `4b82591cb` → 4 / 0 (`s2-local-mergebase-*`) |

- S2 에서 리터럴 핀으로 잰 Go 파일의 첫 셋은 `.moai/reports/t582/repro/main.go`, `internal/cli/codex_blank_review_test.go`, `internal/cli/codex_skills_disable.go` 다. 모두 다른 카드의 작업이다.
- 흡수 시점에 로컬 develop 은 원격보다 159커밋 앞서 있었다(`git rev-list --count origin/develop..develop` → `159`). 원격 기준 merge-base 가 움직이지 않은 이유다.
- 스윕 대상 인벤토리 — develop 의 SPEC `acceptance.md`·`plan.md` 에서 `git diff --name-only <리터럴 SHA>..HEAD` 형태가 **22줄**(SPEC 10개) 잡힌다(`sweep-inventory.txt`). 이 grep 은 형태만 센 것이다. 줄마다 흡수 뒤에 평가되는지, 이미 흡수 전 한정이라고 명시돼 있는지는 판정하지 않았다. 예를 들어 `SPEC-FMT-GATE-001` 51행은 "develop 흡수" 를 문구 안에서 이미 다룬다.

## 3. 기준 귀속 (Baseline-attribution)

- 세 시점 모두 이 워크트리에서 순서대로 쟀다. S0 증거는 S1 카드 커밋으로 들어갔고, S1·S2 증거는 S2 흡수 뒤에 커밋한다. 흡수는 리드 배차문이 정한 기반 정렬 병합(`git merge --no-ff 4b82591cb`)을 그대로 쓴 것이다. 다만 재현을 위해 카드 커밋 하나를 병합보다 먼저 두었다.
- `develop` 과 `origin/develop` 값은 흡수 직전 이 트리에서 `git rev-parse` 로 읽었다(`d060e0d13`, `4b82591cb`).

## 4. 미검증 (Gaps)

- 흡수 대상이 원격 develop 인 경우(로컬 == 원격, 또는 원격만 앞선 경우)는 재지 않았다. 그 경우에는 카드가 제안한 원격 기준 merge-base 도 올바를 것으로 보이지만 이는 추론이다.
- 흡수를 두 번 이상 겹친 경우(카드 커밋 → 흡수 → 카드 커밋 → 재흡수)의 merge-base 동작은 재지 않았다.
- 스윕 인벤토리 22줄 각각이 실제로 흡수 뒤에 평가되는지, 이미 닫힌 SPEC 이라 다시 평가될 일이 없는지는 분류하지 않았다.
- `--first-parent` 계열 대안은 측정하지 않았다.

## 5. 잔여 위험 — 리드가 정할 설계 판단

- **올바른 왼쪽 끝이 무엇인지가 흡수 대상에 달려 있다.** 이번 측정으로는 "흡수한 바로 그 ref 와의 merge-base" 가 올바른 값을 냈다. 현행 절차에서는 그 ref 가 로컬 `develop` 이다. 그러나 절차가 원격 흡수로 바뀌거나, 로컬 develop 이 원격보다 뒤처진 상태에서 흡수하면 답도 달라진다. 규율을 "`git merge-base <흡수한 ref> HEAD`" 로 세울지, 흡수 ref 를 기록해 두는 방식으로 할지, 범위 AC 를 흡수 전 평가로 한정할지는 설계 선택이다.
- **카드 문안의 수리를 그대로 스윕하면 오탐을 옮겨 심는다.** 카드 본문은 `origin/develop` 기준 merge-base 를 권한다. 이번 측정에서 그 형태는 로컬 develop 흡수 뒤 리터럴 핀과 똑같이 51개를 냈다.
- **스윕 범위.** 형태가 맞는 22줄 가운데 무엇을 고치고 무엇을 흡수 전 한정으로 둘지는, 위 규율이 정해진 뒤에야 판정할 수 있다.
