# AC 카운트 baseline 재생성 — 수명주기 연동 cascade 절차

> SPEC-AC-BASELINE-REFRESH-001 (card t1068, 2026-09-23) 이 정본. `internal/spec/ac_count_clause_test.go` 의 `TestACCounterFullCorpusMatchesBaseline` 이 지키는 스냅샷 `.moai/reports/t338/ac-count-baseline.txt` 을 **언제, 어떻게** 다시 찍는지의 절차다. 이 문서가 존재하기 전에는 그 방법이 트리에 없었다 — 스크래치 스크립트로 손수 재생성하다 그 스크립트가 유실되는 사고가 두 번 반복됐다(아래 §5 사고 기록).

---

## 1. 재생성 명령

```
MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1
```

- 재생성 장치는 테스트 패키지 안에 있다. `extractCounterCommand` 으로 B12 절에서 카운터를 **추출**해 쓰므로, 카운터가 바뀌면 재생성 결과도 자동으로 따라온다 — 계측기가 두 벌로 갈라질 여지가 없다.
- `MOAI_AC_BASELINE_REGENERATE=1` 없이는 **아무것도 쓰지 않는다**(기본 off, `TestACRegenerationGateDefaultsOff` 가 단언). CI 가 이 변수를 설정하는 일은 없다.
- 재생성 출력은 `parseACBaseline` 으로 왕복 round-trip 된다(`TestACBaselineEmitterRoundTrip`). 파일 형식이 소비자와 어긋날 수 없다.

## 2. 방아쇠 이벤트 — 이것들이 일어나면 cascade 의무가 생긴다

다음 네 가지는 공통점이 있다: **커밋된 스냅샷의 관측 대상 집단이나 관측값을 움직인다.**

| 이벤트 | 게이트에 나타나는 모양 | 실제 전례 |
|---|---|---|
| `acceptance.md` 의 superseded/split 로 인한 삭제 | `:479` vanish — "present in the snapshot but no longer matched by the corpus glob" (하드 실패) | `20cdeb6bd` — SPEC-MODEL-PROFILE-MATRIX-002 를 네 successor 로 분해하며 삭제 |
| SPEC 디렉터리의 `_archive/` 이동 | 같은 vanish — depth-1 glob 이 더는 그 파일을 매치하지 않는다 | (아직 없음 — 첫 발생이 이 절차의 첫 시험이다) |
| AC 개수에 영향을 주는 corpus 재작성 (B12 절 카운터 문법 변경, corpus glob 변경) | 광범위한 count-move 하드 실패 | t573 (`d9b472409`) — corpus 기준 재작성 후 cascade 누락 |
| 기존 `acceptance.md` 를 제자리에서 고쳐 AC 수가 바뀌는 경우 (SPEC 개정) | 그 파일 한 행의 count-move 하드 실패 — 파일은 코퍼스에 그대로 있는데 기록된 개수만 어긋난다 | `0fbc75afc` — SPEC-APPJS-FIRE-GUARD-001 을 0.1.0 → 0.2.0 으로 개정하며 AC 를 추가하고 cascade 를 빠뜨렸다 |

**네 번째 행은 파일이 사라지지도 코퍼스 기준이 바뀌지도 않는 유일한 방아쇠다.** 세 번째 행과 갈리는 지점은 코퍼스 집단이 아니라 관측값이다: corpus 재작성은 여러 파일의 개수를 한꺼번에 움직이고, SPEC 개정은 **한 파일의 개수만** 움직인다. 두 경우 모두 하드 실패이므로 cascade 의무는 같다.

개정 절차와의 접점: `completed → in-progress (amendment)` 전이(`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix)를 거쳐 기존 `acceptance.md` 의 AC 를 더하거나 빼는 개정은 그 자체로 네 번째 행에 해당한다. 개정 커밋에 §3 **같은 커밋 규칙**이 그대로 적용된다 — 재생성된 스냅샷은 개정 커밋과 같은 커밋에 들어가야 하고, 나누면 그 사이 어떤 HEAD 에서든 게이트가 빨간불이다.

**새 `acceptance.md` 의 추가는 방아쇠가 아니다.** 부재(absent) 행은 v0.5.0 협정상 "report, not fail" 이고 다음 재생성에서 흡수되는 것이 문서화된 수명주기다. 다만 부재 행은 적체로 불어나기만 하므로(아래 §5), 추가만 있던 기간에도 가끔은 소모성 재생성을 해 주는 것이 보고의 가독성을 지킨다.

## 3. 같은 커밋 규칙 (same-commit rule)

[HARD] **방아쇠 이벤트의 커밋에 재생성된 스냅샷이 같이 들어가야 한다.** 커밋을 나누면 사이에 있는 어떤 HEAD 에서든 게이트가 빨간불이고, 그 빨간불은 §5 에서 봤듯이 실제 결함 신호를 잡는 채널을 오염시킨다.

순서:

1. HEAD 를 다시 읽는다 (`git rev-parse --short HEAD`) — 재생성 헤더에 기록되는 SHA 가 곧 커밋되는 트리여야 한다.
2. §1 의 명령을 실행한다.
3. **diff 를 줄마다 읽는다** (`git diff -- .moai/reports/t338/ac-count-baseline.txt`). §4 의 귀속 술어를 통과할 때까지 커밋하지 않는다.
4. 같은 커밋에 넣는다. 스냅샷 커밋은 **스냅샷만** 운반한다(무관한 변경 혼입 금지).

## 4. diff 검토 + 명명-원인 커밋 규율

[HARD] **이름 붙일 수 없는 diff 행은 정규화하지 않고 멈춘다.** 재생성은 측정이지 승인이 아니므로, 변경된 모든 행의 원인을 커밋 메시지가 명명해야 한다. 숙어:

- **추가 행 (+N)**: 각 행이 구 스냅샷에 없던 SPEC 디렉터리로 귀속되는지 확인한다 — 구 스냅샷의 디렉터리 목록과 겹침이 0이어야 한다. HALT 행이 새로 생겼다면 그것부터 의심한다: 정규화된 것인지(마킹 정리 전) 실제 결함인지 커밋 전에 판정하고 메시지에 명명한다. 절대 조용히 COUNT 로 넘기지 않는다.
- **삭제 행 (−1)**: 어떤 커밋이 그 `acceptance.md` 를 지웠는지 `git log --diff-filter=D -- <path>` 로 답을 내고 그 커밋을 메시지에 적는다.
- **count/state 이동**: 원인을 알 수 있는 이동(그 SPEC 의 실제 본문 편집)만 허용되며, 커밋 메시지가 파일별 원인을 명명한다. 원인을 알 수 없는 이동은 재생성이 아니라 **조사 대상**이다.
- **예상 총수를 미리 못 박지 않는다**: corpus 집단은 `acceptance.md` 가 작성될 때마다 움직인다. 멈춰야 할 판정식은 "숫자가 맞는가"가 아니라 **귀속 술어**(위 세 항)다.

커밋 메시지 형식 모델: 이 SPEC 의 M2 커밋 (`chore(SPEC-AC-BASELINE-REFRESH-001): M2 catch-up cascade`) — 행 클래스별 원인, 귀속 근거(겹침 0·중복 0·HALT 0), 제거 행의 원인 커밋 SHA 를 본문에 적었다.

## 5. 사고 기록 — 이 절차가 없어서 반복된 실패 클래스

- **t348** (`23df21c9e`): 게이트와 최초 스냅샷 생성. 재생성 경로가 git 추적 밖의 임시 스크래치 스크립트였다 — 저장소에 커밋되지 않은 경로.
- **t573** (`d9b472409` → 수리 `5f546af2c`): corpus 기준을 재작성하고 cascade 를 빠뜨렸다. **develop tip 자체에서 게이트가 빨개졌고**, 수리는 손수 재생성으로 했는데 그때 쓴 스크래치 스크립트는 유실됐다.
- **t1068 의 발단** (`20cdeb6bd`): superseded-split 이 `acceptance.md` 를 지우고 cascade 를 빠뜨렸다 — 같은 클래스 두 번째. 방치 비용의 실측: 부재 행은 68(primary@main, 09-21) → 75(t1058 병합 트리, 09-21) → 83(cd99336bf, 09-22) → 84(동일 트리, 본 SPEC plan 산출물이 집단에 편입된 직후) → 91(본 SPEC 실행 트리, 09-22)로 단조 증가했고, 줄일 수 있는 유일한 행위는 재생성뿐이다. 게이트가 benign 한 이유로 빨간 상태인 동안, 게이트가 잡으려는 실제 회귀(count/halt/vanish)는 구별 불가능한 빨간불로 묻힌다.

## 6. 판정 금지선 — 재생성이 바꾸지 않는 것

스냅샷 재생성은 **관측을 다시 찍는 것**이지 판정을 다시 쓰는 것이 아니다. 부재는 report-only 로 남고(필수 출력), 기록된 파일의 vanish / count-move / state-move / halt 식별자 집합 이동은 하드 실패로 남는다 (REQ-ABR-006; 원본 계약 `SPEC-AC-COUNT-DISCRIMINATOR-001` spec.md §3.5 rules 1–4). 자동 흡수는 없다 — 축복 행위는 변수를 걸고 diff 를 읽고 커밋하는 **사람의 검토된 실행**이다.

## 7. 커밋 시점 기계 게이트 — `ac-baseline-guard`

> SPEC-ACSNAPSHOT-COMMIT-GUARD-001 (card t1150, 2026-09-24). §2 의 네 번째 행(제자리 개정)이 문서만으로는 막히지 않았다 — t1106 개정과 t1139 개정이 이틀 사이에 develop CI 를 두 번 빨갛게 만들었다. 이 게이트는 그 행 하나를 커밋 순간에 기계로 막는다.

### 7.1 무엇을 검사하는가

git 의 config 정의 훅(`hook.ac-baseline-guard.event = pre-commit`)이 커밋마다 `scripts/ac-baseline/check-staged.sh` 를 실행한다.

- **대상**: 인덱스에서 `HEAD` 대비 상태가 `M` 인 depth-1 `.moai/specs/<dir>/acceptance.md` 만 본다. `_archive/` 와 depth-2 경로, 새 파일(`A`), 삭제(`D`)는 대상이 아니다 — 삭제와 `_archive/` 이동은 여전히 CI 게이트가 잡는다.
- **판정 재료는 전부 인덱스에서 읽는다**: 카운터(스테이징된 `manager-docs.md` 의 sentinel 블록 — 사본이 따로 없다), 개정된 `acceptance.md` 블롭, 스냅샷 `.moai/reports/t338/ac-count-baseline.txt`. 그래서 §3 같은 커밋 규칙을 지킨 커밋, 곧 재생성된 스냅샷을 **같이 스테이징한** 커밋은 구조상 통과한다. 작업 트리에서만 고치고 스테이징하지 않은 스냅샷은 소용이 없다.
- **비교 규칙은 CI 와 같다**(`acComparison`): `live` 나 `excluded` 가 움직이면, COUNT↔HALT 상태가 바뀌면, HALT 식별자 집합이 바뀌면 거절한다. 스냅샷에 기록이 없는 파일은 보고만 하고 통과시킨다.
- **거절 메시지**는 파일 경로, 기록값과 스테이징값, §1 재생성 명령, 이 문서 경로를 함께 적는다. 게이트는 스냅샷을 쓰거나 재생성하지 않는다 — 재생성은 여전히 §1·§6 의 사람 검토 절차다.
- 대상 파일이 없는 커밋은 아무것도 출력하지 않고 통과한다(체커가 있는 트리 기준).

### 7.2 설치 — 리드가 develop 병합 뒤 한 번

[HARD] **설치 주체는 리드이며, 이 카드 브랜치가 로컬 `develop` 에 병합된 뒤 한 번만 실행한다.** 레인은 공유 git 설정을 건드리지 않는다.

```
sh scripts/ac-baseline/install-hook.sh
```

- `.git/config`(모든 링크된 워크트리가 공유하는 설정)에 `hook.ac-baseline-guard.event` 와 `hook.ac-baseline-guard.command` 두 키만 쓴다. 여러 번 실행해도 키마다 값은 하나로 남는다.
- git 2.54 미만이면 발견한 버전과 요구 버전을 적고 아무것도 쓰지 않은 채 실패한다.
- `.git/hooks/*`, 관리형 훅의 출처 파일 `.git/hooks/.moai-pre-commit.sha256`, `core.hooksPath` 는 건드리지 않는다. 이 저장소는 `core.hooksPath=/dev/null` 이라 관리형 hookdir 훅은 돌지 않지만, config 정의 훅은 그 설정과 무관하게 실행된다.
- 설치 확인: `git config --get-regexp '^hook\.ac-baseline-guard\.'` 가 두 줄을 출력해야 한다. 설치 후 실제 거절 커밋 한 번과 통과 커밋 한 번을 develop 을 흡수한 트리에서 관찰해 기록하는 것은 리드의 병합 후 관찰이며, 이 SPEC 의 완료 조건은 아니다.

### 7.3 `NOT CHECKED` — 검사하지 못했다는 뜻이지 통과가 아니다

게이트는 도구 결함에서 **열린 채로 실패한다**(fail open). 카운터 sentinel 이 없거나 중복되거나 순서가 뒤집혔을 때, 카운터 본문이 비었을 때, 스냅샷이 인덱스에 없을 때, 그 파일의 스냅샷 행이 깨졌을 때, 카운터가 0·3 이외의 코드로 끝났을 때는 `ac-baseline-guard: NOT CHECKED (<원인>): <경로>` 한 줄을 stderr 에 쓰고 커밋을 막지 않는다. 막지 않는 이유는 공유 설정에 걸린 훅이 결함 하나로 모든 레인의 커밋을 한꺼번에 세우지 않게 하기 위해서다 — CI 게이트는 그대로 최종 판정자로 남는다. 다만 같은 커밋의 다른 파일에서 실측 불일치가 나오면 그 불일치가 우선해 커밋은 거절된다.

체커 스크립트 자체가 죽는 경우도 같다. 설치된 훅 명령은 체커의 종료 코드 가운데 **1 만** 거절로 넘긴다. 구문 오류(종료 코드 2), 도구 부재, 시그널처럼 그 밖의 0 이 아닌 코드는 `ac-baseline-guard: NOT CHECKED (checker exited <코드>)` 한 줄을 쓰고 커밋을 통과시킨다. 체커 사본 하나가 깨졌다고 그 사본을 실은 트리의 모든 커밋이 막히지 않게 하기 위해서다. 체커를 손으로 하위 디렉터리에서 실행해도 트리 최상위 기준으로 같은 경로를 본다.

체커 스크립트가 없는 트리에서도 같은 줄이 나온다. 설치 시점부터 **아직 develop 을 흡수하지 않은 레인 워크트리**와 **`main` 에 체크아웃된 primary 체크아웃**은 커밋할 때마다 `NOT CHECKED` 를 출력하고 막히지 않는다. 트리가 develop 을 흡수하면 사라지고, primary 체크아웃은 스크립트를 실은 릴리스가 `main` 에 들어갈 때까지 계속 출력한다. 의도된 소음이다.

[HARD] **레인의 완료 보고는 커밋 중에 본 `NOT CHECKED` 줄을 그대로 인용한다.** 이 줄은 커밋한 세션의 stderr 에만 나타나므로, 보고에 옮겨 적지 않으면 "검사하지 못함"이 "통과"로 읽힌다(`.claude/rules/moai/core/verification-claim-integrity.md` §1).

### 7.4 우회

- 유일한 우회는 `git commit --no-verify` 다. 터미널의 사람은 쓸 수 있지만, Claude 세션은 쓸 수 없다 — PreToolUse 가드(`internal/hook/pre_tool.go`)가 `git commit` 과 `--no-verify` 를 함께 담은 명령을 거부한다. 에이전트 세션에서 거절을 풀려면 스냅샷을 재생성해 같은 커밋에 스테이징하는 것이 정석이다.
- 그 가드의 거부 메시지가 권하는 `SKIP_MOAI_PRECOMMIT=1` 은 관리형 hookdir 훅만 읽는 변수라 **이 게이트를 우회하지 못한다.** 게이트 전용 우회 변수는 따로 두지 않았다.

### 7.5 막지 않는 것

- 충돌 없이 끝나는 `git merge` 는 `pre-commit` 이 아니라 `pre-merge-commit` 을 실행하므로 통합 창의 병합은 게이트를 거치지 않는다. 병합되는 브랜치의 커밋들은 만들어질 때 이미 검사를 받았다. 충돌을 풀고 `git commit` 으로 마무리하는 병합은 검사된다.
- 카운터 문법이나 corpus glob 을 바꾸는 §2 세 번째 행은 대상이 아니다. 게이트는 스테이징된 카운터를 읽지만 비교는 개정된 파일에만 한다.
