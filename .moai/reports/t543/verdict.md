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

## 6. 리드 판정 뒤 수리 (같은 날)

**판정.** 리드가 ①을 채택했다 — 범위 판정식의 왼쪽 끝은 `git merge-base develop HEAD`(흡수한 ref) 이다. ②(흡수 ref 기록)는 상태 파일을 늘리고, ③(흡수 전 한정)은 흡수 후 재측정 규율과 부딪혀 기각됐다. 한계 둘을 규율에 적으라는 지시가 붙었다.

| 산출물 | 커밋 | 내용 |
|---|---|---|
| 규율 | `5f65c1420` | `.claude/rules/local/gitflow-lane-protocol.md` §8 에 [HARD] 항목 추가. 이 룰은 `paths:` 한정이라 상시 로드 표면을 늘리지 않는다. S2 재현을 대조군으로 인용하고, 한계 (a) 병합 뒤 사용 불가 · (b) 흡수 ref 원칙을 적었다 |
| 한계 (a) 실측 | `5f65c1420` | 이미 develop 에 병합된 카드 브랜치 `WT-develop-guide-conflicts` 에서 `git merge-base develop <브랜치>` 가 브랜치 tip `8ce9c5620` 자신이었다. `git diff --name-only 8ce9c5620..WT-develop-guide-conflicts` 는 빈 출력, EXIT=0 이었다 (`repro/limit-a-post-merge-all.txt`) |
| 스윕 분류 | `5f65c1420`, `848579f1f` | `sweep-classification.md`. 1차 인벤토리 22줄·SPEC 10개, 넓힌 인벤토리 80줄·SPEC 27개. 흡수 뒤 평가 대상은 draft 인 `SPEC-UPDATE-DOC-DRIFT-001` 하나다(나머지 26개는 completed) |
| SPEC 수리 1 | `8f92320de` (manager-spec) | AC-UDD-023 두 명령과 plan.md §C 한 줄을 `CARD_BASE` 형태로 바꾸고 무필터 대조군을 붙였다. 옛 값은 날짜 붙은 참고 판독으로 내렸다. 0.3.1 |
| SPEC 수리 2 | `2fbf48512` (manager-spec) | AC-UDD-021 두 명령(`diff --stat`, `log`)을 같은 형태로 바꿨다. unstaged 확인 줄은 그대로 두었고, 0.3.1 HISTORY 행을 보강했다 |

**수리 확인(lane-6 이 이번 실행에서 직접 확인).** 두 수리 커밋의 diff 를 읽었다. `acceptance.md`·`plan.md` 에서 `7f61332ef` 를 찾으면, 명령으로 남은 것은 없고 모두 기준값을 적은 문장(`Baseline at ...`, `Observed at ...`)이나 참고 판독이다. SPEC lint 는 두 번 모두 트리에서 빌드한 도구로 돌렸고 결과는 `0 error(s), 1 warning(s)` 으로 편집 전과 같았다(`spec-repair-lint.txt`, `spec-repair2-lint.txt`).

**계기 결함 한 건(내 몫).** 1차 인벤토리는 `git diff --name-only <SHA>..HEAD` 형태만 잡아서, 같은 SPEC 의 AC-UDD-021(`diff --stat` · `log` 형태)을 놓쳤다. 수리 워커가 찾아냈고, 형태를 넓혀 다시 쟀다.

### 6.1 이 절의 Gaps

- ~~Definition of Done 산문이 옛 기준을 말함~~ — **닫힘.** 리드 지시로 이 카드에서 고쳤다. 커밋은 `bc8a78b0a`(manager-spec)이고, lane-6 이 diff 를 직접 읽어 확인했다. `acceptance.md` §D 의 해당 항목은 이제 "AC-UDD-021 이 재는 이 SPEC 자신의 범위(읽는 시점의 `CARD_BASE`, 병합 전)"를 기준으로 말하고, 0.3.1 HISTORY 행에도 이 수정을 덧붙였다. 수리 뒤 §D 안에는 `7f61332ef` 가 한 건도 남지 않았다(`spec-repair3-grep.txt`: §D 범위 적중 0, 같은 식으로 파일 전체를 세면 30 이고 이는 `grep -c` 결과와 같다). lint 는 트리에서 빌드한 도구로 돌려 `0 error(s), 1 warning(s)` 으로 이전과 같았다(`spec-repair3-lint.txt`).
- 이 SPEC 의 lint 는 원래부터 `StatusGitConsistency` 경고를 낸다(frontmatter `draft` 대 git 이 가리키는 `implemented`). **리드 판단: 살아 있는 draft 다. 근거는 develop 이력 `ddfe2253f`**("v0.3.0 staleness rewrite — retire 4, re-anchor 3, keep 5 live", #1515)와 frontmatter `status: draft` 다. 따라서 수리 전제는 유지된다. 경고가 왜 나는지는 이 카드 범위 밖이라 판정하지 않았다. 리드는 그 커밋 제목의 `feat(...)` 접두를 구현 신호로 읽은 것이라고 추정했지만, 측정한 것은 아니다.
- 넓힌 grep 도 오른쪽 끝이 `HEAD` 가 아닌 범위, SHA 를 변수에 담은 경우, 세 점(`...`) 범위는 잡지 않는다.
- 규율에 따른 실행(카드들이 실제로 `CARD_BASE` 형태를 쓰는지)은 문서 수정으로 보장되지 않는다. 카드 본문이 요구한 "실행 규율로 세운다"는 이번에는 규칙 문서화까지만 했다.
- §4 Gaps(원격 흡수, 흡수 중첩, `--first-parent`)는 리드 판정에 따라 그대로 둔다.

## 7. 병합 트리 재측정 — 통합 창 안 (lane-6)

리드 지명 후 창을 잡고(`moai integration acquire --name lane-6`) 로컬 develop 을 흡수한 트리에서 다시 쟀다. 흡수 전 측정은 병합 뒤 근거로 재사용하지 않는다.

**흡수 대상과 흡수.** `git fetch origin develop` 후 `git rev-list --count --left-right origin/develop...develop` → `0 189` 이므로 흡수 대상은 로컬 develop `0dfb3f605` 다. `git merge --no-edit develop` → HEAD `7bef04ebd`, 트리 `4ea3985a3`. 흡수 전 `git merge-tree --write-tree --name-only develop HEAD` 가 예측한 트리와 같다(충돌 파일 0). 도구체인은 `go1.26.8` 이다(`window-go-version.txt`).

**델타 판정 — 이 카드가 세운 규율로 잰다(§8, 병합 전 평가).**

| 시점 | `CARD_BASE = git merge-base develop HEAD` | 무필터 대조군 | Go 프로브 | 증거 |
|---|---|---|---|---|
| 흡수 직전 (HEAD `a28e462bf`) | `4b82591cb` | 32 | 0 | `window-pre-absorb-scope-all.txt`, `-go.txt` |
| 흡수 직후 (HEAD `7bef04ebd`) | `0dfb3f605` (방금 흡수한 develop tip) | 32 | 0 | `window-post-absorb-scope-all.txt`, `-go.txt` |

두 시점의 파일 집합은 정확히 같다(차집합 양방향 모두 공집합). 흡수로 develop 의 커밋 189개가 들어왔는데도, 규율대로 다시 구한 범위는 카드 자기 기여만 가리키고 Go 변경은 0 이다. S2 재현에서 리터럴 핀이 51개를 냈던 바로 그 상황에서 규율이 올바른 값을 낸다는 것을, 이 카드 자신의 흡수로 한 번 더 관측했다.

- 카드 경로의 흡수 영향: `git diff --name-only 4b82591cb 0dfb3f605 -- .claude/rules/local/gitflow-lane-protocol.md .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/` → 빈 출력, EXIT=0 (`window-delta-card-paths.txt`). develop 은 이 카드의 규칙 파일과 SPEC 파일을 건드리지 않았다.
- 파일 동일성: 규칙 파일과 SPEC 파일 3개의 blob 이 흡수 전(`window-blobs-before.txt`)과 흡수 후(`window-blobs-after.txt`)에 모두 같다.
- SPEC lint: `go run ./cmd/moai spec lint .moai/specs/SPEC-UPDATE-DOC-DRIFT-001`(흡수 트리에서 빌드) → `0 error(s), 1 warning(s)`, EXIT=0. 경고는 기존 `StatusGitConsistency` 하나로 이전 세 번의 실행과 같다(`window-spec-lint.txt`).

이 카드는 Go 코드를 바꾸지 않으므로 Go 테스트는 재측정 대상이 아니다.

**이 절의 Gaps.** 병합 뒤에는 §8 규율의 한계 (a)에 따라 범위 판정식을 쓰지 않는다. 병합 후 근거는 develop 병합 트리와 카드 브랜치 트리의 동일성으로 대신한다(병합 뒤 보고에 적는다).
