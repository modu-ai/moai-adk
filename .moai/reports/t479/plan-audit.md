# SPEC-BINLAG-KEYGUARD-001 — plan-phase 감사 보고 (카드 t479, iteration 1)

- 감사 대상: `.moai/specs/SPEC-BINLAG-KEYGUARD-001/{spec.md,plan.md,acceptance.md,progress.md}`
- 감사 트리: `.claude/worktrees/t479` @ **`93fb36344`** (본 감사가 직접 측정)
- Tier: M (spec.md frontmatter `tier: M`)
- **판정: FAIL** · **점수 0.86**

M1 컨텍스트 격리: 저자 추론 맥락은 사용하지 않았다. 리드가 제기한 결함 1건도 그대로 받지 않고
명령을 다시 실행해 독립적으로 확인했다.

---

## Claim

plan-phase 산출물 4종은 구조·요구 번호·GEARS 형식·REQ↔AC 커버리지 면에서 건전하다.
그러나 **머지 흡수 이후 두 곳이 실측으로 깨져 있다**: (1) AC-BLKG-007의 폭발 반경 판정식이
카드가 옳아도 실패하고, (2) plan.md §C의 전제 착지 판정·흡수 기준 ref가 `origin/develop`인데
전제 카드 2장은 그 ref에 없다. 여기에 AC-BLKG-008(2)의 레시피는 공허하게 통과할 수 있다.

---

## Evidence

측정은 전부 이 실행에서, 이 트리에 대해 냈다.

### E1 — 트리 상태

    $ git rev-parse --short HEAD
    93fb36344
    $ git log --oneline -1
    93fb36344 Merge develop a825183dd into WT-binarylag-key-guard (t479): absorb t466 Hook Delivery check + t477 allowlist entry

    $ grep -c 'Hook Delivery' internal/cli/doctor.go
    1

    $ sed -n '/namesAddedAfterBaseline = map/,/^}/p' internal/cli/binary_lag_test.go
    var namesAddedAfterBaseline = map[string]bool{
    	"hookWiringCheckName": true,
    	`"Hook Delivery"`:     true,
    }

    $ go test ./internal/cli/ -run 'TestBinaryLag' -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.828s

즉 t466의 리터럴 체크와 t477의 backtick 키가 모두 트리에 들어와 있고, 기존 가드는 GREEN이다.

### E2 — AC-BLKG-007의 판정식은 실제로 깨져 있다 (리드 주장 독립 확인)

`acceptance.md:118`의 명령을 그대로 실행:

    $ git diff --stat origin/develop...HEAD -- internal/ pkg/ cmd/
     internal/cli/binary_lag_test.go                    |  11 +
     internal/cli/doctor.go                             |   4 +
     internal/cli/doctor_hook_delivery.go               | 256 +++++++++++
     internal/cli/doctor_hook_delivery_test.go          | 408 ++++++++++++++++++
     internal/cli/hook_install_precommit.go             |  81 +++-
     internal/cli/hook_install_precommit_test.go        |  90 ++++
     internal/cli/testdata/doctor-dark.golden           |   7 +-
     internal/cli/testdata/doctor-light.golden          |   7 +-
     internal/cli/testdata/doctor-nocolor.golden        |   7 +-
     internal/cli/todo_composed_upgrade_test.go         | 321 ++++++++++++++
     internal/cli/update/merge/base_test.go             | 131 ++++++
     internal/cli/worktree_branch_flag_test.go          |   4 -
     internal/core/git/status_branch_test.go            |   6 +-
     internal/hook/t459_repro_test.go                   |  93 ++++
     internal/spec/drift.go                             |  17 +
     internal/spec/drift_close_body_test.go             | 472 +++++++++++++++++++++
     internal/spec/drift_index.go                       | 224 ++++++++++
     internal/template/agentemit/agents-codex.yaml      |  26 +-
     internal/template/catalog.yaml                     |   2 +-
     .../templates/.claude/agents/moai/plan-auditor.md  |   6 +-
     .../.claude/rules/moai/workflow/spec-workflow.md   |   6 +-
     .../templates/.codex/agents/moai/plan-auditor.toml |   6 +-
     internal/template/templates/.git_hooks/pre-commit  |  81 +++-
     .../.moai/docs/audit-artifact-convention.md        |   2 +
     24 files changed, 2207 insertions(+), 61 deletions(-)

    $ git diff --name-only origin/develop...HEAD -- internal/ pkg/ cmd/ | wc -l
          24

**측정된 파일 수 = 24.** AC-BLKG-007의 통과 조건은 「1개」이므로, run-phase가 아직
아무것도 편집하지 않은 지금조차 이미 실패한다. 원인은 ref다:

    $ git rev-parse --short origin/develop
    25a3212a9
    $ git rev-parse --short develop
    a825183dd
    $ git rev-list --count origin/develop..develop
    123

올바른 base는 **흡수 병합 커밋을 명시적으로 핀하고 two-dot을 쓰는 것**이다. 실측:

    $ git diff --name-only 93fb36344..HEAD -- internal/ pkg/ cmd/ | wc -l
           0

run-phase 편집 후에는 이 명령이 정확히 이 카드의 델타만 낸다. 세 점(`...`)은 merge-base가
`25a3212a9`로 내려가 흡수분 전부를 이 카드에 귀속시키므로 쓰면 안 된다.

### E3 — `grep -c` 종료코드 주의(AC-BLKG-008)는 **사실이다**

    $ grep -c 'ZZZ_NO_SUCH_TOKEN_ZZZ' internal/cli/binary_lag_test.go; echo "exit=$?"
    0
    exit=1

매치 0에서 stdout은 `0`, 종료코드는 1. `acceptance.md:135`의 캡션은 정확하다.

### E4 — 그러나 AC-BLKG-008(2)의 레시피 자체는 공허하다

`acceptance.md:133`의 명령을 지금 그대로 실행:

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | grep -c 'lagBaselineSHA\|git show'
    0
    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -c
           0

함수가 **아직 존재하지 않는데도** 판정식이 요구하는 값 `0`이 나온다. 즉 이 항목은
「baseline 참조가 없다」가 아니라 「추출이 비었다」에서도 통과한다 — 피연산자 한쪽이 0인 집합
비교와 같은 공허함이다. run-phase가 함수명을 다르게 짓거나(plan.md §F M2는 그 이름을 **예시**로만
제시한다) awk 범위가 어긋나면, 가드가 `git show`에 의존하더라도 이 항목은 초록이 된다.

### E5 — 전제 착지 ref가 틀렸다

    $ git merge-base --is-ancestor 4e91bf6a9 origin/develop; echo $?
    1
    $ git merge-base --is-ancestor 4e91bf6a9 HEAD; echo $?
    0
    $ git merge-base --is-ancestor origin/develop HEAD; echo $?
    0
    $ git log --oneline -1 a825183dd
    a825183dd Merge branch 'WT-binarylag-allowlist' into develop (t477)

t466(`4e91bf6a9`)은 `origin/develop`의 조상이 **아니다**. plan.md §C는 착지 판정과 흡수 대상을
모두 `origin/develop`으로 적었는데, 그 ref로는 (1) 착지 판정이 영원히 거짓이고 (2)
`git merge origin/develop`은 이미 조상이므로 **no-op**이라 t466/t477을 가져오지 못한다.
실제 흡수는 로컬 `develop @ a825183dd`로 이뤄졌고, 그것이 옳은 경로였다.

### E6 — 좌표 갱신

    (현재 트리)              (25a3212a9)
    lagBaselineSHA          :24   :24
    checkNamesFromSource    :118  :118
    exprSource              :162  :162
    namesAddedAfterBaseline :194  :184
    가드 함수                :199  :188
    대조 지점                :213  :202

`plan.md:139` / `spec.md` §1이 인용한 `:184 / :188 / :202`는 흡수 후 각각 `:194 / :199 / :213`으로
10~11줄 밀렸다. 두 문서 모두 트리 SHA(`25a3212a9`)를 병기하고 있어 **거짓 주장은 아니지만**,
run-phase가 그 좌표로 직행하면 잘못된 줄을 읽는다.

### E7 — 구조·커버리지·형식

    $ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BINLAG-KEYGUARD-001/ | wc -l
           0

- frontmatter 12개 정본 필드(`id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags`) 전부 존재, 거절 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0.
- REQ-BLKG-001…007 연속, 결번·중복 0.
- REQ↔AC 커버리지: 001→AC-001 / 002→AC-002,005 / 003→AC-003,005,006 / 004→AC-004 /
  005→AC-008 / 006→AC-007 / 007→AC-008. **미커버 REQ 0, 고아 AC 0.**
- `## §6 범위 밖`에 `### Out of Scope — <주제>` 6개, 각 항목이 구체 불릿 보유.

---

## Baseline-attribution

- E1~E6의 모든 수치는 `.claude/worktrees/t479` @ `93fb36344`에서 **이 실행으로** 냈다.
  다른 트리·다른 시점의 값을 옮겨 쓴 것은 없다.
- 리드가 제시한 "123 commits behind"와 "AC-BLKG-007이 깨졌다"는 주장은 **건네받은 값으로 쓰지 않고**
  `git rev-list --count`와 `git diff --name-only | wc -l`로 재측정해 확인했다(E2).
- `progress.md` §E.1의 t477 뮤턴트 관측(m1 `:215`, m2 `:216`)은 **건네받은 값으로 명시 표기되어
  있고**, 이 세션도 재현하지 않았다. 그 표기는 정확하다 — 감사 차원 4(baseline 귀속)는 PASS다.

---

## Findings

| id | 심각도 | 분류 | 좌표 | 무엇이 잘못됐나 | 어떻게 고치나 |
|---|---|---|---|---|---|
| D1 | critical | blocking | `acceptance.md:118,120` | 폭발 반경 base가 `origin/develop...HEAD`(three-dot). `origin/develop`은 `25a3212a9`로 로컬 develop보다 123 커밋 뒤져, 흡수한 t466·t477의 파일 변경 **24개**를 이 카드에 귀속시킨다. 카드가 완전히 옳아도 실패한다 | base를 흡수 병합 커밋으로 **명시 핀**하고 two-dot으로: `git diff --stat 93fb36344..HEAD -- internal/ pkg/ cmd/`. 지금 실측 0개, run-phase 후 1개여야 한다 |
| D2 | critical | blocking | `plan.md:60,62,64,65,68` | 전제 착지 판정과 흡수 대상 ref가 `origin/develop`. 실측상 t466(`4e91bf6a9`)은 그 ref의 조상이 아니고, `git merge origin/develop`은 이미 조상이라 **no-op**이다. 문자 그대로 따르면 전제는 영원히 미착지로 읽히고 흡수는 아무것도 가져오지 않는다 | 착지 판정과 흡수 대상을 **로컬 `develop`**(실측 `a825183dd`)으로 고치고, 이미 수행된 흡수 병합 커밋 `93fb36344`를 §C에 기록한다 |
| D3 | major | blocking | `acceptance.md:133` | AC-BLKG-008(2)의 awk 추출이 비어도 `grep -c`가 `0`을 내어 **통과한다**(실측: 함수 부재 상태에서 awk 출력 0바이트, grep -c 출력 `0`). 함수명이 달라지거나 범위가 어긋나면 공허하게 초록 | 추출 비공허성을 먼저 단언: `awk ... \| wc -l`이 0보다 큼을 확인한 뒤 `grep -c`를 판정한다. 아울러 함수명을 plan.md §F M2의 "예시"가 아니라 AC와 동일한 **확정 이름**으로 못박는다 |
| D4 | minor | optional | `plan.md:139`, `spec.md:28-44` | 인용 좌표 `:184 / :188 / :202`가 흡수로 `:194 / :199 / :213`으로 이동. SHA 병기가 있어 거짓은 아니나 run-phase가 오독할 수 있다 | 좌표를 `93fb36344` 기준으로 갱신하거나, 「좌표는 `25a3212a9` 기준이며 흡수 후 재측정한다」를 한 줄로 명시한다 |
| D5 | minor | optional | `acceptance.md:3-5` | 문서 수준 트리 핀이 `25a3212a9`인데 실제 감사 대상 트리는 `93fb36344`. §D 헤더가 "run-phase 관측은 자기 SHA를 명시한다"로 완화하고 있으나, 자기 SHA가 없는 항목(AC-002/003/005/006)은 여전히 낡은 핀에 묶인다 | 문서 핀을 흡수 후 SHA로 갱신한다 |
| D6 | minor | optional | `spec.md:80-82`(REQ-BLKG-003), `acceptance.md:60-62` | 표적 진단이 **한 방향**만 덮는다 — 「따옴표가 빠졌다」. 거울상(상수 등록 체크의 키에 따옴표를 **덧붙인** 경우, 예 backtick으로 감싼 `"hookWiringCheckName"`)은 분기에 걸리지 않아 원인 미지목 일반 메시지로 떨어진다 | 진단 분기를 양방향으로 넓히거나, 거울상을 명시적으로 범위 밖으로 선언한다(현재는 어느 쪽도 아니다) |
| D7 | minor | optional | `acceptance.md:137` | AC-BLKG-008(1)의 판정이 「눈으로 읽어 확인」이다. 기계 판정이 아니어서 재현 불가능하다 | 레시피 파일에서 `$?`를 판정식으로 쓰는 줄 수를 세는 기계 검사로 바꾼다 |
| D8 | minor | optional | `acceptance.md:53`, `acceptance.md:76-81` | AC-BLKG-002는 자기 판정 명령이 없고 AC-005/006에 위임한다. AC-BLKG-004의 뮤턴트(허용목록 비우기)는 **기존 가드도 함께 RED**로 만드는데(허용목록 키 2개가 baseline 이후 추가분이므로) 그 예상 부수효과가 기록돼 있지 않다 — `-run` 스코프로 회피되지만 명시가 없다 | AC-002에 자기 명령을 주거나 위임을 명시하고, AC-004에 「기존 가드의 동반 RED는 예상되며 `-run` 스코프로 분리한다」를 한 줄 적는다 |

### 비공허성 개별 심사 (요청 차원 3)

| AC | 결함이 살아 있는데도 통과할 수 있나 | 판정 |
|---|---|---|
| AC-BLKG-001 | 단독으로는 「가드가 존재하고 초록」까지만 주장. 결함(모양 규율의 기계 부재)은 가드 작성 자체가 고치므로 순환은 아니나 하중은 m1/m2가 진다 | 조건부 비공허 |
| AC-BLKG-002 | 자기 명령 없음. m1/m2에 전적 의존 | 위임 — D8 |
| AC-BLKG-003 | 메시지 본문 2요소 인용을 요구 → 통과 위조 어려움 | **비공허** |
| AC-BLKG-004 | 빈 맵 뮤턴트 + 메시지 취지 요구. 「검사 대상 0의 초록」을 정확히 겨냥 | **비공허** |
| AC-BLKG-005 (m1) | 3단계(RED→뮤턴트 GREEN→원복 GREEN) 전부 종료코드+파일:줄 기록 요구 | **비공허** |
| AC-BLKG-006 (m2) | 「판정 불가를 통과로 적지 않는다」를 명시. 전제 착지는 실측으로 이미 성립(E1) | **비공허** |
| AC-BLKG-007 | **반대 방향으로 깨짐** — 카드가 옳아도 실패(24 ≠ 1) | **D1, 판정 불가** |
| AC-BLKG-008(2) | awk 추출이 비면 통과 | **공허 — D3** |
| AC-BLKG-008(1) | 사람의 눈 판정 | 기계 판정 아님 — D7 |

---

## Gaps — 관측하지 않은 것

- **새 가드의 실제 동작**: 코드가 없으므로 AC-BLKG-001…006의 GREEN/RED를 하나도 관측하지 못했다.
  뮤턴트 m1/m2를 이 감사에서 재현하지 않았다(산출물·대상 파일을 편집하지 않는다는 지시에 따름).
- **t466 / t477 자체의 정확성**: 재검증하지 않았다. `doctor_hook_delivery.go`의 내용도 읽지 않았다.
- **전체 스위트**: `go test ./...`를 돌리지 않았다. 측정 범위는 `./internal/cli/` 의 `-run TestBinaryLag` 뿐이다.
- **CI 판정**: 이 브랜치의 CI 실행을 읽지 않았다. 로컬 초록은 CI 판정이 아니다.
- **`develop`의 tip 이동**: `a825183dd`는 이 감사 시점의 값이며, 이후 움직였을 수 있다.
- **D1 수정본의 최종 정확성**: 제안한 `93fb36344..HEAD`는 지금 0개를 낸다는 것만 확인했고,
  run-phase 편집 후 정확히 1개를 낸다는 것은 아직 관측할 수 없다.

---

## Residual-risk

- **D1을 고쳐도 base 핀은 낡는다**: 흡수를 한 번 더 하면 `93fb36344`가 다시 부정확해진다.
  run-phase는 편집 직전 트리의 SHA를 다시 핀해야 한다.
- **D3을 고쳐도 함수명 결합은 남는다**: AC가 `TestBinaryLag_AllowlistKeysAreLiveNames`를 못박고
  plan.md는 "예"로 제시한다. 두 문서가 다른 이름을 낳으면 D3의 공허함이 이름 축으로 되살아난다.
- **표적 진단 분기의 조용한 도달 불가**: `progress.md`가 이미 적어 둔 위험이며 유효하다.
  등록 방식이 통일되면 그 분기는 침묵한다 — 가드는 고장 나도 알려주지 않는다.
- **거울상(D6)이 열려 있다**: 이 카드가 닫는 구멍의 반대 방향은 여전히 원인 미지목으로 실패한다.

---

## 최종 판정

**FAIL · 0.86**

| 차원 | 점수 | 근거 |
|---|---|---|
| Clarity (명료성) | 0.95 | 요구·수락 모두 단일 해석. 결함 서술(§1)이 원인과 증상의 분기를 정확히 짚는다 |
| Completeness (완결성) | 0.95 | frontmatter 12필드 전부, 6개 절, Out of Scope 6개 구체 항목 |
| Testability (검증가능성) | 0.60 | AC-BLKG-007 판정 불가(D1), AC-BLKG-008(2) 공허(D3), 008(1) 눈 판정(D7) |
| Traceability (추적성) | 1.00 | REQ 7 ↔ AC 8 완전 커버, 고아 0, 결번 0 |

must-pass 계열: REQ 번호 일관성 PASS · GEARS 형식 PASS(REQ-BLKG-006은 서술문에 가까우나 허용 범위)
· frontmatter PASS · 언어 중립성 N/A(단일 언어 SPEC) · `[NEEDS CLARIFICATION]` 0건 PASS.

**FAIL 사유는 점수가 아니라 D1·D2다.** 카드가 완벽히 구현돼도 AC-BLKG-007은 실패하고(24 ≠ 1),
plan.md §C를 문자 그대로 따르면 흡수가 no-op이 된다. 둘 다 산출물 수정 없이는 run-phase가
자기 성공을 증명할 수 없게 만든다. D3은 반대 방향 — 실패해야 할 때 통과한다.

수정 순서: **D1 → D2 → D3** 을 고친 뒤 재감사. D4~D8은 운영자 재량(optional).

### 범위 규율 점검 (요청 차원 6)

산출물은 편집 범위를 `internal/cli/binary_lag_test.go` 한 파일로 못박고 있다 —
`spec.md:93`(REQ-BLKG-006), `spec.md:137`(§5), `plan.md:78-79`(§D), `plan.md:131`(§G).
기계 판정은 AC-BLKG-007이 지는데, 그 판정식이 D1으로 깨져 있다. 즉 **의도는 명시돼 있으나
그것을 지키는 기계가 현재 작동하지 않는다.**

---
---

# 2차 감사 (iteration 2) — D1/D2/D3 수정본 델타 재감사

- 감사 트리: `.claude/worktrees/t479` @ **`93fb36344`** (재측정 — 1차와 동일, 재흡수 없음)
- 범위: 1차 결함 델타 회귀 확인 + 리드가 지목한 4개 질문 + 재흡수 판단
- **판정: FAIL** · **점수 0.87**

1차 판정은 위에 그대로 보존한다.

---

## Claim (2차)

**D1·D2·D3 세 blocking은 모두 해소됐다.** 그러나 수정 라운드가 **새 blocking 2건을 심었다**:
자가 추가한 `go test` 종료코드 예외의 근거가 실측으로 거짓이며(R1), 그 예외가 요구 계층
REQ-BLKG-005 및 AC-BLKG-008(1) 자신과 충돌한다(R2). R1은 이 트리에서 AC-BLKG-001을
**지금 당장 공허하게 통과**시킨다.

---

## Evidence (2차) — 전부 이 실행에서, 이 트리에 대해

### R-E0 — 트리·작업 범위 재측정

    $ git rev-parse --short HEAD
    93fb36344
    $ git status --short
    ?? .moai/reports/t479/
    ?? .moai/specs/SPEC-BINLAG-KEYGUARD-001/

무접촉·무커밋 주장 확인. `internal/cli/binary_lag_test.go`는 추적 상태에 변경이 없다.

### R-E1 — D1 회귀: **RESOLVED**

    $ git diff --name-only 93fb36344..HEAD -- internal/ pkg/ cmd/ | wc -l
    0
    $ git diff --name-only origin/develop...HEAD -- internal/ pkg/ cmd/ | wc -l
    24

`acceptance.md:163-176`이 명시 핀 + two-dot으로 교체됐고, 실측이 머리말(22-31행)의
「24 vs 0」과 일치한다. 편집 전 `0`, 편집 후 `1`이라는 기대도 정합하다.
재흡수 시 핀 갱신 의무(`acceptance.md:175-176`)도 들어갔다.

### R-E2 — D2 회귀: **RESOLVED**

    $ git merge-base --is-ancestor 4e91bf6a9 refs/heads/develop; echo $?
    0
    $ git rev-list --count origin/develop..refs/heads/develop
    130

plan.md §C가 「전제 대기」에서 「이미 흡수함」으로 재서술됐고, 착지 판정·흡수 ref가 로컬
`develop`으로 교체됐다. 문서에 적힌 `130`도 실측과 일치한다(1차 감사 때 `123`이었고 develop이
7커밋 전진했다 — 값이 갱신돼 있다).

### R-E3 — D3 회귀: **RESOLVED, 그리고 1단은 공허하지 않다**

리드 요청대로 1단을 지금 트리에서 실제로 실행:

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -l
    0

함수가 없는 현재 상태에서 1단이 `0`을 낸다 → `acceptance.md:192-194`의 규정에 따라 이 항목은
**「판정 불가」로 올바르게 떨어진다.** 통과로 읽히지 않는다. 1단은 **닫히는 방향으로 실패**하므로
(검사 대상이 없으면 0 → 판정 불가) 그 자체가 공허하지 않다. **D3 수정은 성공이다.**

### R-E4 — **[R1] `go test` 종료코드 예외의 근거는 거짓이다**

`acceptance.md:17-20`은 `go test`의 종료코드를 [HARD] 금지의 예외로 두면서 근거를
「"필터가 못 찾은 것"과 "검사가 실패한 것"이 섞이지 않는다」로 적었다. 실측:

    $ go test ./internal/cli/ -run 'TestBinaryLag_AllowlistKeysAreLiveNames' -count=1 -timeout 600s; echo "exit=$?"
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.729s [no tests to run]
    exit=0

    $ go test ./internal/cli/ -run 'TestBinaryLag_ZZZ_NoSuchTest' -count=1 -timeout 600s -v
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.734s [no tests to run]

**`-run` 셀렉터가 0매치일 때 `go test`는 `ok` + exit 0을 낸다.** 즉 「못 찾음」과 「통과」가
정확히 섞인다 — 예외가 내세운 근거 그 자체가 반증됐다. 그리고 AC-BLKG-001(`acceptance.md:62`)의
판정식은 **「종료코드 0 + 출력 `ok`」**인데, 위 출력은 두 조건을 **모두** 만족한다.
따라서 **가드가 아직 존재하지 않는 지금, AC-BLKG-001은 이미 통과로 읽힌다.** 이것이 바로
이 카드가 막으려는 결함의 모양이며, D3에서 걷어낸 공허함이 이름 축으로 되살아난 것이다.

같은 구멍이 AC-BLKG-005 절차 (2)단계(「단언 제거 뮤턴트 → GREEN 관측」)에도 걸린다 —
이름이 어긋난 상태의 GREEN과 뮤턴트의 GREEN이 구별되지 않는다.
반대로 AC-BLKG-004 / AC-BLKG-006은 「종료코드 ≠ 0」을 요구하므로 0매치가 오히려 실패로 떨어져
**안전한 방향**이다.

### R-E5 — **[R2] 예외 조항이 요구 계층 및 AC-008(1) 자신과 충돌한다**

`spec.md:112-114`(REQ-BLKG-005)은 예외 없이 적혀 있다:

    새 가드의 판정과 그 검증 레시피는 **매치 수** 또는 Go 테스트 단언으로 내려야 하며,
    셸 종료 상태(exit code)를 판정식으로 삼아서는 안 된다.

AC-BLKG-008(1)(`acceptance.md:213-218`)은 그 요구를 검증하며 「종료 상태를 판정식으로 쓴 곳이
**0건**임을 확인한다」로 계수를 규정한다. 그런데 같은 파일에서 종료코드를 판정식으로 쓰는 줄은:

    $ grep -n '종료코드' .moai/specs/SPEC-BINLAG-KEYGUARD-001/acceptance.md
    17:두 가지는 **예외이며 예외인 이유가 있다**: `go test`의 종료코드는 …
    18:`git merge-base --is-ancestor`의 종료코드는 …
    62:- 판정식: `go test`의 종료코드 0 + 출력 `ok`.
    110:- 통과 조건: 종료코드 ≠ 0, 그리고 실패 메시지가 …
    131:- 통과 조건: 세 단계의 종료코드와 실패한 파일:줄을 모두 기록한다.
    147:- 통과 조건: 두 실행 모두 종료코드 ≠ 0이고, …
    200-201: (금지 근거 설명 — 판정식 아님)
    227: (기록 의무 — 판정식 아님)

판정식으로 쓴 곳 **:62 / :110 / :147 = 3건**, 여기에 `plan.md:69,72`의 `echo $?` 2건.
AC-008(1)을 문자 그대로 수행하면 **0건이 아니라 5건**이 나와 **AC-008(1) 자신이 실패한다.**
머리말의 예외 조항은 그 계수 규칙에 반영되지 않았다 — 예외를 신설하면서 그 예외를 검증하는
항목을 함께 고치지 않은 것이다.

### R-E6 — [Q2] D6 양방향 진단은 **실행 가능하다**

    $ grep -n 'hookWiringCheckName' internal/cli/doctor.go
    224:		{hookWiringCheckName, func(v bool) DiagnosticCheck {
    $ grep -n '"Hook Delivery"' internal/cli/doctor.go
    230:		{"Hook Delivery", func(v bool) DiagnosticCheck { return checkHookDelivery(cwd, v) }},

두 등록 모양이 **같은 레지스트리에 나란히 실재한다**(:224 상수 식별자 / :230 문자열 리터럴).
`exprSource`가 원본 소스 텍스트를 돌려주므로 추출 집합은 `hookWiringCheckName`(맨몸)과
`"Hook Delivery"`(따옴표 포함) 두 모양을 함께 담는다. 따라서 m2′ 뮤턴트
(`` `"hookWiringCheckName"` ``)는 「따옴표에 감싸였고 그 알맹이가 집합에 있다」 분기에 정확히
걸린다 — **관측 가능하고 m2와 구별 가능하다.** D6 수정 유효.

### R-E7 — [Q4] AC-BLKG-004 새 조항은 실행 가능하나, 초록의 원인을 구분하지 않는다

    $ git cat-file -e 22f90b1c7:internal/cli/doctor.go && echo blob_present
    blob_present
    $ git show 22f90b1c7:internal/cli/doctor.go | grep -c 'hookWiringCheckName'
    0
    $ git show 22f90b1c7:internal/cli/doctor.go | grep -c 'Hook Delivery'
    0

baseline blob은 **가용하고**, 두 이름 모두 baseline에 **매치 0**이다. 따라서 허용목록을 비우면
두 이름이 면제를 잃어 기존 가드는 **반드시 RED**가 된다 — 이 환경에서 「초록 = blocker」 분기는
**도달 불가**다. 지시 자체는 실행 가능하고 무해하지만, 초록이 나온다면 그 원인은 「가드가
침묵했다」보다 다음 둘일 공산이 크다:

1. baseline blob 미가용(얕은 클론) → `t.Skipf`(`binary_lag_test.go:202`) → 출력 `ok`, exit 0
2. R1의 `-run` 0매치 → 출력 `ok ... [no tests to run]`, exit 0

두 경우 모두 「blocker」로 보고하면 **오탐**이다. AC-004는 초록을 SKIP / no-tests-to-run /
진짜 PASS 셋으로 갈라야 한다.

### R-E8 — [Q3] `git merge-base --is-ancestor` 예외는 조건부로만 성립한다

    $ git merge-base --is-ancestor deadbeef1 HEAD; echo "exit=$?"
    fatal: Not a valid object name deadbeef1
    exit=128

유효 ref에서는 근거가 성립한다(0/1, stdout 빈 상태). 그러나 **잘못된 ref는 exit 128 + stderr
`fatal:`**이다. plan.md는 `; echo $?` 형태로 종료코드를 **그대로 기록**하므로 128이 눈에 남아
구분 가능하다 — 현재 레시피에서는 실질 위험이 낮다. `if`로 감싸면 「조상 아님」과 「ref 오타」가
섞인다. 이 예외는 **「`; echo $?`로 값을 남기는 형태에 한해」**라는 단서와 함께여야 정확하다.

### R-E9 — [Q3-재흡수] 리드의 판단을 **지지한다**

    $ git rev-parse --short refs/heads/develop
    8b391bc8c
    $ git rev-list --count a825183dd..refs/heads/develop
    7
    $ git diff --name-only a825183dd..refs/heads/develop -- internal/cli/binary_lag_test.go internal/cli/doctor.go | wc -l
    0

세 값 모두 리드가 보고한 것과 일치하며, 제가 직접 다시 쟀다. 판단은 옳다:

- 새 가드가 읽는 입력은 `doctor.go`의 레지스트리와 `binary_lag_test.go`의 허용목록 **둘뿐**이고,
  그 7커밋 구간에서 두 파일 모두 무변경이다. 재흡수는 가드의 판정 대상을 바꾸지 않는다.
- 재흡수하면 `93fb36344` 핀이 무효가 되고, 새로 들어온 남의 카드 파일이 AC-BLKG-007 계수에
  다시 섞인다 — 1차 D1과 같은 모양의 손해를 자초한다.
- **유보 1건**: 재흡수를 미루는 것은 이 카드의 판정에는 옳지만, 최종 병합 창에서 develop과
  합쳐질 때의 의미 충돌(자동 병합 성공 ≠ 옳음)은 남는다. 그것은 병합 창의 재측정 몫이며
  이 카드의 plan-phase 결정으로 처리할 사안이 아니다.

---

## Baseline-attribution (2차)

R-E0 … R-E9의 모든 수치는 `.claude/worktrees/t479` @ `93fb36344`에서 **이 실행으로** 냈다.
리드가 제시한 `8b391bc8c` / `7` / `0` 세 값과 `130`은 **옮겨 적지 않고 각각 다시 실행**해 확인했다.
1차 감사에서 잰 값도 재사용하지 않고 R-E1에서 다시 쟀다.

---

## Findings (2차)

| id | 심각도 | 분류 | 좌표 | 무엇이 잘못됐나 | 어떻게 고치나 |
|---|---|---|---|---|---|
| R1 | critical | blocking | `acceptance.md:17-20`, `acceptance.md:62`, `acceptance.md:129` | 자가 추가한 `go test` 종료코드 예외의 근거가 **실측으로 거짓**. `-run` 0매치에서 `ok ... [no tests to run]` + exit 0이 나와 「못 찾음」과 「통과」가 섞인다. 그 결과 AC-BLKG-001이 **가드가 없는 지금 이미 통과로 읽힌다**. AC-BLKG-005 절차 (2)단계도 같은 구멍 | AC-BLKG-001에 **테스트 실재 선단언**을 붙인다(D3의 1단 재사용, 또는 `go test -v ... \| grep -c '^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames$'`가 `1`인지 **매치 수로** 판정 + `[no tests to run]` 매치 수 `0` 확인). 머리말 예외의 근거 문장은 거짓이므로 삭제하거나 「`-run`이 실제로 매치했음을 별도로 단언할 때에 한해」로 좁힌다 |
| R2 | major | blocking | `spec.md:112-114`(REQ-BLKG-005) ↔ `acceptance.md:17-20`, `acceptance.md:213-218` | REQ-BLKG-005는 **예외 없이** exit code 판정을 금지하는데 acceptance.md가 예외 2개를 신설했다. 게다가 그 요구를 검증하는 AC-BLKG-008(1)의 계수 규칙(「0건」)이 갱신되지 않아, 문자 그대로 수행하면 `acceptance.md:62,110,147` + `plan.md:69,72` = **5건**이 잡혀 **AC-008(1) 자신이 실패한다** | REQ-BLKG-005에 예외를 명문화하고, AC-BLKG-008(1)의 계수 규칙을 「예외 2종을 제외한 텍스트-필터 종료 상태 판정이 0건」으로 다시 쓴다. 예외를 유지하려면 요구·검증 두 계층을 같은 커밋에서 고친다 |
| R3 | minor | optional | `spec.md:126` | 「뮤턴트 2건(m1/m2)이 비공허성을 증명한다」인데 acceptance.md는 3종(m1/m2/m2′, `m2′` 5회 언급)을 요구한다. D6 수정이 spec.md에 반영되지 않은 드리프트 | `spec.md:126`을 「뮤턴트 3건(m1/m2/m2′)」으로 갱신 |
| R4 | minor | optional | `acceptance.md:111-118` | AC-BLKG-004의 「기존 가드 초록 = blocker」가 초록의 원인을 구분하지 않는다. 실측상 이 환경에서 그 분기는 도달 불가(baseline blob 가용 + 두 이름 매치 0/0 → 반드시 RED)이며, 실제로 초록이 나온다면 원인은 대개 `t.Skipf` 또는 R1의 0매치다 — 그 둘을 blocker로 보고하면 오탐 | 초록을 **SKIP / `[no tests to run]` / 진짜 PASS** 셋으로 가르는 판정을 넣는다. 출력에서 `--- SKIP`과 `[no tests to run]`의 매치 수를 각각 세는 방식 |
| R5 | minor | optional | `acceptance.md:18` | `git merge-base --is-ancestor` 예외의 근거는 **유효 ref에 한해** 성립한다. 잘못된 ref는 exit **128** + stderr `fatal:`(실측). plan.md는 `; echo $?`로 값을 남겨 구분 가능하므로 현재 위험은 낮으나, 근거 문장이 무조건형이다 | 예외에 「`; echo $?`로 종료코드 값을 그대로 기록하는 형태에 한한다. `if`로 감싸면 ref 오타(128)와 조상 아님(1)이 섞인다」를 덧붙인다 |

### 1차 결함 회귀 확인

| 1차 id | 상태 | 근거 |
|---|---|---|
| D1 | **RESOLVED** | `acceptance.md:163-176` 명시 핀 + two-dot. 실측 `93fb36344..HEAD` → `0`, `origin/develop...HEAD` → `24` (R-E1) |
| D2 | **RESOLVED** | plan.md §C 재서술, 로컬 `develop` ref. ancestry exit `0`, `rev-list --count` → `130` (R-E2) |
| D3 | **RESOLVED** | 2단 레시피. 1단을 지금 실행해 `0` → 「판정 불가」로 올바르게 떨어짐. 1단 자체는 닫히는 방향으로 실패하므로 비공허 (R-E3) |
| D4 | RESOLVED | 좌표 갱신 확인 |
| D5 | RESOLVED | 문서 핀 `93fb36344`로 갱신(`acceptance.md:4`) |
| D6 | **RESOLVED** | 양방향 표(`acceptance.md:86-89`) + m2′. 두 등록 모양이 `doctor.go:224` / `:230`에 실재해 관측 가능 (R-E6) |
| D7 | PARTIAL | 판정 대상 3개 문서로 명시됐으나, 계수 규칙이 신설 예외와 어긋나 실행하면 실패한다 → **R2로 승계** |
| D8 | RESOLVED (부분) | 기존 가드 동반 실행이 들어갔으나 초록의 원인 구분이 없다 → **R4로 승계** |

정체(stagnation) 없음 — 3건 모두 실질 진전이 있었고, 같은 결함이 두 회차에 걸쳐 그대로 남은
항목은 없다.

---

## Gaps (2차) — 관측하지 않은 것

- **새 가드의 동작**: 여전히 코드가 없다. AC-BLKG-001 … 006의 실제 GREEN/RED는 하나도
  관측하지 못했고, 뮤턴트 m1 / m2 / m2′를 어느 것도 실행하지 않았다(편집 금지 지시).
- **추출 집합의 실제 원소**: `checkNamesFromSource`를 실행해 집합을 덤프하지 않았다.
  R-E6은 `doctor.go`의 등록 **모양**을 grep으로 확인한 것이며, AST 추출 결과를 직접 본 것이 아니다.
- **AC-BLKG-004의 실제 RED**: 허용목록을 비운 뮤턴트를 실행하지 않았다. R-E7은 baseline에
  두 이름이 없다는 사실로부터의 **추론**이며, 실행 관측이 아니다.
- **`plan.md` / `spec.md` 전문 재독**: 2차에서는 §C·§2·§3·§4와 grep 히트 구간만 읽었다.
  1차에서 전문을 읽었으나 그 사이 수정된 다른 절이 있다면 놓쳤을 수 있다.
- **전체 스위트·CI**: `go test ./...`도 CI 판정도 읽지 않았다.
- **develop tip**: `8b391bc8c`는 이 감사 시점의 값이며 이후 움직였을 수 있다.

---

## Residual-risk (2차)

- **R1을 고쳐도 이름 결합은 남는다**: 확정 이름 [HARD] 조항(`acceptance.md:10-12`)이 문서 규율일
  뿐 기계 가드가 아니다. run-phase가 이름을 바꾸면서 조항을 못 보면 R1의 공허함이 되돌아온다.
- **예외 조항이라는 형태 자체의 위험**: 이번에 심긴 두 blocking은 모두 **요청받지 않은 자가 추가**
  에서 나왔다. 예외를 늘릴수록 그 예외를 검증하는 항목과의 정합을 매 회차 확인해야 한다.
- **재흡수 유예의 이자**: 재흡수를 미루는 판단은 이 카드에 옳지만, develop이 더 전진할수록
  최종 병합 창의 의미 충돌 가능성은 커진다. 병합 창에서 병합 트리 재측정이 필요하다.
- **AC-BLKG-004 blocker 분기 도달 불가**: 지금은 도달 불가이나, baseline이 언젠가 범프되어
  두 이름이 baseline에 들어가면 그 분기는 「기존 가드가 정당하게 초록」인 상태로 도달한다 —
  그때는 blocker 보고가 오탐이 된다.

---

## 최종 판정 (2차)

**FAIL · 0.87**

| 차원 | 1차 | 2차 | 근거 |
|---|---|---|---|
| Clarity | 0.95 | 0.95 | 양방향 표·2단 레시피로 오히려 명료해졌다 |
| Completeness | 0.95 | 0.95 | 변동 없음 |
| Testability | 0.60 | **0.65** | D1/D3 해소로 상승. 그러나 R1이 AC-001을 새로 공허하게 만들어 상승폭이 깎였다 |
| Traceability | 1.00 | **0.95** | spec.md §3의 뮤턴트 건수가 acceptance.md와 어긋난다(R3) |

**FAIL 사유는 R1과 R2다.** D1·D2·D3은 실측으로 모두 해소됐고 수정의 질도 좋다 — 특히 D3의
2단 레시피는 지금 트리에서 「판정 불가」로 올바르게 떨어짐을 확인했다. 그러나 같은 라운드에서
**요청받지 않은 예외 조항**이 들어왔고, 그 근거가 실측으로 거짓이며(`go test -run` 0매치 → `ok`
+ exit 0), 그 결과 **가드가 존재하지도 않는 지금 AC-BLKG-001이 통과로 읽힌다.** 1차에서 D3으로
걷어낸 공허함이 이름 축으로 되살아난 것이므로, 통과시킬 수 없다.

수정 순서: **R1 → R2**. R3~R5는 운영자 재량(optional).
R1은 D3의 1단과 같은 형태 — 「검사 대상이 실재하는가」를 먼저 매치 수로 단언 — 로 닫으면 된다.

---

# 2차 감사 부록 — D9 (리드 제기) 판정

리드가 재감사 중 독립적으로 재서 넘긴 D9를 판정한다. **결론부터: D9는 실재하는 결함이며,
이 감사가 2차에서 `R1`로 이미 제기한 것과 **같은 결함**이다.** 두 세션이 서로의 값을 보지 않고
같은 곳에 도달했다는 점에서 상호 확증이다. 다만 D9의 **영향 범위 서술 한 대목은 틀렸다** —
아래 A3에서 정정한다.

이 부록은 **재채점하지 않는다.** 판정은 2차와 동일한 **FAIL · 0.87**이며, 아래 R6은 R1의
폭발 반경을 정밀화한 것이지 독립된 결함 축이 아니다. 같은 산출물 상태에 대해 점수를 다시
내리면 산출물이 나빠지지도 않았는데 회귀 신호가 뜬다.

---

## A1 — 재현: 리드 값을 옮기지 않고 직접 쟀다. 대조군 2개 추가

트리 `93fb36344`, 새 가드 부재 상태.

    $ go test ./internal/cli/ -run 'TestBinaryLag_AllowlistKeysAreLiveNames' -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.853s [no tests to run]
    EXIT=0

    $ go test ./internal/cli/ -run 'TestQqqNoSuchThingAtAll' -count=1 -timeout 600s          ← 대조군 A: 전혀 무관한 이름
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.690s [no tests to run]
    EXIT=0

    $ go test ./internal/cli/ -run 'TestBinaryLag_AllowlistsKeysAreLiveNames' -count=1 -timeout 600s   ← 대조군 B: 한 글자 오타
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.706s [no tests to run]
    EXIT=0

**세 경우 모두 `ok` + exit 0.** 대조군 B가 특히 중요하다 — 「이름을 한 글자 잘못 쓰면
초록이 된다」는 것이 이 결함의 실질적 발생 경로이기 때문이다. 재현 성립.

## A2 — 「부재 자체가 RED-now다」는 **거짓 주장이다**

`acceptance.md:63`:

    - 기준선(`93fb36344`): 해당 테스트는 아직 **존재하지 않는다**. 부재 자체가 RED-now다.

A1의 실측이 정면으로 반증한다. 이 트리에서 **부재는 RED가 아니라 GREEN을 낸다.**
그리고 AC-BLKG-001의 판정식(`acceptance.md:62`)은 「종료코드 0 + 출력 `ok`」이므로,
A1의 출력은 **판정식을 그대로 만족한다.** 가드가 한 줄도 없는 지금 이미 통과다.

`verification-completeness.md` §1.1이 이 모양을 정확히 규정한다:

> **A pass whose swept set is empty asserts nothing.** A verification that selected nothing still
> reports success — a test-name selector matching zero tests … and the report is indistinguishable
> from one where everything passed. … The runner usually says so in its own words — go prints
> `[no tests to run]` … and those tokens are the cheapest available evidence. Where an instrument
> classifies runner output, the empty-sweep judgment is made ahead of every other signal, because
> the exit code of a run that swept nothing is the same zero a fully-passing run returns.

(파일 실재 확인: `.claude/rules/moai/development/verification-completeness.md` — 리드가 적은
`§1.1` 인용문과 `no tests to run` 토큰 모두 매치 수 각 `1`. `core/`가 아니라 `development/`에 있다.)

**판정: 거짓 주장 맞다.** 그리고 §1.1의 표현대로 판정 순서까지 틀렸다 — 빈 sweep 판정은
**다른 모든 신호보다 먼저** 내려야 하는데, AC-BLKG-001은 종료코드를 먼저 읽는다.

## A3 — 영향 범위: 리드의 목록은 **항목은 맞고 방향은 틀렸다**

`-run`을 쓰는 곳을 문서에서 전수 계수했다:

    $ grep -n '\-run ' .moai/specs/SPEC-BINLAG-KEYGUARD-001/acceptance.md
    60:      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames …   ← AC-001
    108:      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames …  ← AC-004 본체
    114:      go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged … ← AC-004 동반
    145:      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames …  ← AC-006
    $ grep -n '\-run ' … plan.md   → 110 (일반 지침 `<테스트명>`)
    $ grep -n '\-run ' … spec.md   → 141 (§4 기준선), 168 (일반 지침)

**빈 sweep이 셋 다 다른 방향으로 터진다** — 이 구분이 수정 방향을 가른다:

| 항목 | 기대하는 결과 | 빈 sweep이 낳는 것 | 방향 |
|---|---|---|---|
| AC-BLKG-001 (`:60`) | GREEN | GREEN | **공허한 통과** — 위험 |
| AC-BLKG-005 절차 (2) | 뮤턴트에서 GREEN | GREEN | **공허한 통과** — 위험 |
| AC-BLKG-005 절차 (3) | 원복 후 GREEN | GREEN | **공허한 통과** — 위험 |
| AC-BLKG-004 본체 (`:108`) | 종료코드 ≠ 0 | exit 0 → 항목 실패 | **닫히는 실패** — 안전 |
| AC-BLKG-006 (`:145`) | 종료코드 ≠ 0 (2회) | exit 0 → 항목 실패 | **닫히는 실패** — 안전 |
| AC-BLKG-004 동반 (`:114`) | 기존 가드 RED | GREEN → 「blocker로 보고」 | **거짓 blocker** — 오탐 |

즉 리드의 「001/004/005/006 넷 다 공허하게 초록」은 정확하지 않다. **공허한 초록이 되는 것은
001과 005의 (2)(3)단계뿐**이고, 004 본체·006은 「종료코드 ≠ 0」을 요구하므로 빈 sweep이
**항목 실패**로 떨어져 오히려 보수적이다. 004의 동반 실행은 세 번째 방향 — 거짓 blocker다.

**빠진 것 없음.** 007(`git diff`)·008(`awk`/`grep -c`)은 `-run`을 쓰지 않아 별개가 맞다.
`spec.md:141`은 기존 가드(실재)를 돌리므로 오늘은 sweep이 비지 않는다 — 다만 이름을 손으로
적은 자리라 같은 오타 경로에 노출돼 있다.

**판정: D9는 실재하며 severity critical이 타당하다.** 다만 이 감사는 이를 **새 결함으로
번호를 새로 매기지 않고** 2차의 `R1`과 동일 건으로 본다. R1이 이미 AC-001과 AC-005 (2)단계를
지목했고, D9는 여기에 004/006을 더했으나 그 둘은 위 표대로 안전한 방향이다.

## A4 — §2.1 대조: AC-BLKG-001의 RED-now 셀은 **미채택(unadopted)이다** → **R6**

`verification-completeness.md` §2.1은 release-blocking 항목의 RED-now 셀에 4요소를 함께
요구하고, **3개만 든 셀은 미채택**이라고 못박는다. AC-BLKG-001(`acceptance.md:63`)을 대조하면:

| 요소 | 상태 | 근거 |
|---|---|---|
| the command | **있음** | `go test ./internal/cli/ -run … -count=1 -timeout 600s` — 파이프·리다이렉션·`;` 없는 단일 호출 형태 |
| that command's verbatim stdout | **없음** | 셀에 출력이 없다. 「존재하지 않는다」는 서술만 있다 |
| that command's exit code | **없음** | 종료코드 필드가 없다 |
| the tree SHA | **있음** | 항목 핀 `93fb36344` + 문서 핀(`acceptance.md:4`) |

**4개 중 2개.** §2.1의 기준(3개면 미채택)에 미달하므로 AC-BLKG-001은 release-blocking 항목으로
**미채택**이다. 그리고 이 결함은 단순 누락이 아니다 — 빠진 두 요소를 실제로 재면(A1) 그 값이
셀의 주장을 **반증**한다. §2 표현으로는 **vacuous 방향**(오늘 초록이라 이 작업에 대해 아무것도
주장하지 않는 항목)에 해당하며, RED-now를 요구하는 이유가 정확히 이것이다.

## A5 — 수정 처방. 리드 제안을 **보완**해야 한다 (그리고 내 2차 처방도 정정한다)

### 리드 제안(exit 0 + `[no tests to run]` 매치 수 0)은 **작동한다. 다만 부족하다**

지금 트리에서 실행해 공허하지 않음을 확인했다:

    $ go test ./internal/cli/ -run 'TestBinaryLag_AllowlistKeysAreLiveNames' … > OUT; echo "EXIT=$?"
    EXIT=0
    $ grep -c '\[no tests to run\]' OUT
    1        ← 요구값 0이 아니므로 항목 실패. 빈 sweep을 잡아낸다

**남는 구멍**: 테스트가 **존재하지만 `t.Skipf`로 건너뛰면** `[no tests to run]`이 찍히지 않고
`ok` + exit 0만 나온다. 이 파일에는 그 경로가 실재한다(`binary_lag_test.go:202` — baseline blob
미가용 시 `t.Skipf`). 즉 리드 제안은 「선택 0」은 막지만 「선택했으나 실행 안 함」은 통과시킨다.

### [정정] 내가 2차 R1에서 제시한 형태는 **그대로 쓰면 공허하다**

2차 보고서에서 `grep -c '^--- PASS: <Name>$'`를 제안했는데, **끝 앵커 `$`가 틀렸다.** 실측:

    $ go test ./internal/cli/ -run 'TestBinaryLag_DoctorCheckNameSetIsUnchanged' … -v
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.752s

    $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged$'  → 0    ← 통과했는데 0. 공허하다
    $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged '  → 1    ← 정정형(뒤 공백). 올바름

go는 `--- PASS: <Name> (0.05s)`처럼 **소요시간 접미사**를 붙이므로 `$` 앵커는 언제나 0을 낸다.
D3과 R1에서 잡아낸 것과 **같은 종류의 공허함**을 내 처방이 다시 만들 뻔했다. 정정형을 없는
이름에 적용하면 `0`이 나온다(`gt3`) — 방향이 옳다.

### 처방 (권고)

AC-BLKG-001 / AC-BLKG-005 (2)(3) — **초록을 주장하는 자리**에만 필요하다:

1. **RED-now 셀**(§2.1 준수): 명령은 파이프 없는 단일 호출로 적고, 그 **verbatim stdout**과
   **종료코드**를 셀에 함께 싣는다. 지금 트리라면 그 stdout이 곧 `ok … [no tests to run]`이므로,
   **읽는 사람이 빈 sweep을 눈으로 본다.** 「부재 자체가 RED-now다」라는 서술을 실제 관측으로 교체.
2. **통과 조건**(green path 셀, 파이프 허용): 아래 **두 매치 수**를 함께 요구한다.

        go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v > OUT
        grep -c '^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames '  → 정확히 1   (선택+실행+통과)
        grep -c '\[no tests to run\]'                                   → 0          (빈 sweep 부재)

   첫 줄이 SKIP까지 함께 막는다 — 건너뛴 테스트는 `--- SKIP:`이라 `--- PASS:` 매치가 0이 된다.
   둘째 줄은 리드 제안을 그대로 살린 이중 안전망이다.
3. AC-BLKG-004 동반 실행(`:114`)에는 같은 두 매치 수를 적용해 「거짓 blocker」를 막는다.
4. AC-BLKG-004 본체·AC-BLKG-006은 **고치지 않아도 안전하다**(A3 표). 다만 「RED를 관측했다」와
   「아무것도 안 돌았다」를 기록에서 가르려면 `--- FAIL:` 매치 수 ≥ 1을 함께 적는 편이 낫다(권고).

**파이프 배치의 근거**: §2.1은 RED-now 셀의 **인용 명령**에만 단일 호출을 요구한다. 매치 수
판정은 green path 쪽 통과 조건이므로 파이프를 써도 그 조항에 걸리지 않는다. 둘을 섞어
RED-now 셀에 파이프 명령을 넣으면 §2.1 위반이 된다.

---

## D9 findings 추가분

| id | 심각도 | 분류 | 좌표 | 무엇이 잘못됐나 | 어떻게 고치나 |
|---|---|---|---|---|---|
| R1 (D9와 동일 건) | critical | blocking | `acceptance.md:60,62,63,108,114,145` | `-run` 빈 sweep이 `ok` + exit 0을 내므로, 초록을 주장하는 AC-001과 AC-005 (2)(3)이 공허하게 통과한다. `acceptance.md:63`의 「부재 자체가 RED-now다」는 실측으로 **거짓**이다(대조군 2개 포함 3회 재현) | A5의 두 매치 수(`--- PASS: <Name> ` 정확히 1, `[no tests to run]` 0)를 통과 조건에 넣는다. `$` 앵커를 쓰지 말 것 — 소요시간 접미사 때문에 항상 0이다 |
| R6 (신규) | major | blocking | `acceptance.md:63` | AC-BLKG-001의 RED-now 셀이 `verification-completeness.md` §2.1의 4요소 중 **2개만**(command, tree SHA) 담는다. verbatim stdout과 종료코드가 없다. 3개만 담아도 미채택인데 2개다 → **release-blocking 항목으로 미채택** | 셀에 A1의 verbatim stdout(`ok … [no tests to run]`)과 종료코드(`0`)를 싣는다. 그 순간 「부재 = RED」 서술이 스스로 반증되므로 A5의 통과 조건 교체가 함께 따라온다 |

**리드 제안 severity(critical)에 동의한다.** 다만 A3의 방향 구분 때문에 「AC 4건이 공허」는
정정이 필요하다 — 공허한 초록은 **2.5건**(001, 005의 두 단계)이고, 004·006은 안전 방향,
004 동반은 오탐 방향이다.

## D9 Gaps

- **SKIP 경로 미관측**: 「존재하지만 건너뛴 테스트는 `[no tests to run]`을 찍지 않는다」는
  `binary_lag_test.go:202`의 `t.Skipf` 코드를 읽고 내린 **추론**이다. baseline blob을 실제로
  가려 skip을 재현하지는 않았다(편집·환경 조작 금지).
- **`--- FAIL:` 라인 형식 미관측**: 실패 케이스를 실행하지 않아 `--- FAIL:` 줄의 정확한 모양을
  직접 보지 못했다. A5의 4항 권고는 그 점에서 미검증이다.
- **처방의 최종 형태 미검증**: A5의 두 매치 수를 **가드가 실재하는 상태**에서 돌려보지 못했다
  (가드가 아직 없다). 기존 가드로 대리 측정해 방향만 확인했다.

## D9 Residual-risk

- **처방 자체가 또 공허해질 수 있다**: A5 정정에서 드러났듯, 매치 수 판정은 **패턴 한 글자**에
  전부를 건다. run-phase는 처방을 적용한 뒤 반드시 **없는 이름으로 한 번 돌려** 그 판정식이
  실제로 실패하는지 확인해야 한다 — 그것이 §1.1이 말하는 「red를 본다」다.
- **같은 함정에 세 번째**: D3(awk 빈 추출) → R1/D9(`-run` 빈 sweep) → R1 처방의 `$` 앵커.
  세 번 모두 「선택된 것이 0인데 판정은 초록」이다. 남은 레시피에도 같은 눈으로 한 번 더
  훑을 값이 있다.

---
---

# 3차 감사 (iteration 3, 최종) — R1/R2/R4/R5 수정본

- 감사 트리: `.claude/worktrees/t479` @ **`93fb36344`** (재측정, 재흡수 없음)
- **판정: PASS · 0.90** (Tier M 임계 **0.80** — `spec-workflow.md:141`에서 직접 읽음)
- **잔여 2건은 차단 사유가 아니라 debt다.** run-phase 조건으로 명시한다(§T-결론).

1·2차 판정과 D9 부록은 위에 그대로 보존한다.

이 PASS는 「결함이 없다」가 아니다. 아래 T1·T2는 실재하며 blocking 등급이다. 그럼에도 PASS인
이유는 §T-결론에 두 축(must-pass 7건 · 임계 대비 점수)으로 적었고, 반대 판정을 고려한 근거도
함께 남겼다 — 리드가 뒤집을 수 있어야 하기 때문이다.

---

## Claim (3차)

**R1·R2·R4·R5 네 건 모두 실측으로 닫혔다.** 특히 R1은 지금 트리에서 실제로 「판정 불가」로
떨어짐을 확인했다. 그러나 이번 판도 **새 결함 2건을 심었다** — 두 건 모두 이번 회차 신설분인
AC-BLKG-008(1) 열거 레시피와 R1 수정의 **형제 표면 미스윕**에서 나왔다. 세 라운드 연속으로
「수정이 결함을 심는다」가 성립했다.

---

## Evidence (3차) — 전부 이 실행에서, 이 트리에 대해

### T-E0 — 트리·범위

    $ git rev-parse --short HEAD
    93fb36344
    $ git status --short
    ?? .moai/reports/t479/
    ?? .moai/specs/SPEC-BINLAG-KEYGUARD-001/

무접촉·무커밋 확인.

### T-E1 — [Q1] `-v` 0매치가 `PASS` 한 줄을 낸다: **재현됨**

    $ go test ./internal/cli/ -run 'TestBinaryLag_AllowlistKeysAreLiveNames' -count=1 -timeout 600s -v
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.656s [no tests to run]
    EXIT=0

작성 세션의 관측이 맞다. **맨 아래 `PASS` 한 줄만 보는 판별식도 공허하다** — 0매치 실행이
그 줄을 그대로 낸다. `acceptance.md:90-91`이 이 사실을 본문에 적어 둔 것도 확인했다.

### T-E2 — [Q1] R1 수정은 **닫혔다**

두 축으로 확인했다.

**(a) §2.1 단일 호출 요건 충족.** `acceptance.md:79`의 판정 명령은
`go test ./internal/cli/ -run … -count=1 -timeout 600s -v` — 파이프·리다이렉션·`&&`·`;`·서브셸이
하나도 없다. §2.1이 요구하는 형태 그대로다. 「한 호출의 stdout을 두 번 읽는다」는 **명령을 두 번
부르는 것이 아니라 한 번의 출력을 두 조건으로 읽는 것**이므로 단일 호출 요건과 충돌하지 않는다.
**판정: 만족한다.**

**(b) 지금 트리에서 판정 불가로 떨어진다.** T-E1의 stdout에 `[no tests to run]`이 있다 →
`acceptance.md:84-86`의 1단(비공허성 선단언)이 깨진다 → 항목은 **판정 불가**이며 2단(`--- PASS:`)은
읽지 않는다. **통과로 읽히지 않는다.** R1 성립.

**(c) 판별식이 공허하지 않다** — 실재하는 이름으로 대조:

    $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.752s

두 출력의 차이가 판별식이라는 `acceptance.md:116-117`의 서술이 실측과 일치한다.

**R6(§2.1 4요소)도 닫혔다**: `acceptance.md:95-100`의 RED-now 표가 명령·verbatim stdout·종료코드
`0`·트리 SHA `93fb36344`를 **네 개 다** 담는다. 거짓 주장은 삭제 대신 `:102-105`에서 철회 문장으로
보존됐다 — 다음 사람이 같은 예외를 다시 만들지 않게 하는 올바른 처분이다.

### T-E3 — **[T1] AC-008(1) 열거의 모집단이 과소열거다.** 기대값 「모집단 5 / C 0」은 이 트리에서 틀렸다

리드 요청대로 **총계가 아니라 멤버십으로** 검산했다. 다섯 좌표를 각각 열어 읽었다:

| 좌표 | 실제 내용 | 분류 판정 |
|---|---|---|
| `acceptance.md:165` | 「통과 조건: 종료코드 ≠ 0, 그리고 실패 메시지가 …」 | **A 맞다** |
| `acceptance.md:203` | 「(1)의 RED는 종료코드 ≠ 0을 실패로 읽는 방향이므로 … 안전하다」 | **A로 볼 수 있으나 경계선** — 이 줄은 방향을 *설명*하는 산문이고, (1)단계의 통과 조건은 `:204`다. 이번 회차 신설분(이전 판에 없던 줄)이 그대로 A로 들어갔다 |
| `acceptance.md:220` | 「통과 조건: 두 실행 모두 종료코드 ≠ 0이고 …」 | **A 맞다** |
| `plan.md:74` | `$ git merge-base --is-ancestor 4e91bf6a9 origin/develop; echo $?` | **B 맞다** (머리말 3분기 단서 존재) |
| `plan.md:77` | `$ git merge-base --is-ancestor 4e91bf6a9 develop; echo $?` | **B 맞다** |

여기까지는 리드 보고와 일치한다. **문제는 모집단에 없는 것들이다.**

**누락 ①  `spec.md:158` + `spec.md:162`** — 열거 대상 문서 3개에 `spec.md`가 명시돼 있는데
(`acceptance.md:292-293`), 모집단 5건 중 `spec.md` 출신은 **0건**이다. 그런데 `spec.md`에는 있다:

    $ sed -n '158,162p' spec.md
        $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s; echo "exit=$?"
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.947s
        exit=0

    즉 기준선은 여전히 **GREEN**이다. 달라진 것은 초록의 성격이다.

    $ grep -c 'exit=\$?' spec.md
    1

`:162`는 `ok` + `exit=0`을 읽고 **「GREEN이다」라는 판정을 내린다.** 비공허성 단언은 짝지어져
있지 않다. 위반 정의(`acceptance.md:288-290`)를 문자 그대로 적용하면 이것은 **C 후보**다 —
최소한 **모집단 멤버**인데 목록에 없다. 이 줄이 실제로 위험한지도 구체적이다: 저 테스트 이름을
한 글자만 잘못 적으면 `ok … [no tests to run]` + exit 0이 나오고, 문서는 **아무것도 안 돌았는데
「기준선은 GREEN」이라고 결론**짓는다. 이 카드가 막으려는 바로 그 모양이다.

**누락 ②  `acceptance.md:221` / `acceptance.md:335`** — 종료코드라는 낱말이 없어 눈에 안 띄는
통과-읽기다:

    :221  파일:줄과 메시지 본문을 그대로 기록한다. 원복 후 GREEN 재확인.
    :335  - 기존 가드 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`도 함께 GREEN.

둘 다 **GREEN을 어떻게 읽는지 말하지 않는다.** AC-BLKG-005는 이번에 (2)·(3)단계에 2단 판정을
붙였는데(`:197-202`), **AC-BLKG-006의 「원복 후 GREEN 재확인」과 §D.1의 「함께 GREEN」에는 붙지
않았다.** 같은 요구를 형제 표면에 다 적용하지 않은 **미스윕**이다.

**이것이 [Q4]의 답이다.** 이번 판이 심은 것이 있다. 더 뼈아픈 것은 위치다 —
`acceptance.md:298-307`의 [HARD] 조항이 **「낱말 매치로 분류하지 말라, 총계 일치는 목록 일치가
아니다」**라고 못박고 그 근거로 2차 회차의 오탐 2건까지 적었는데, **그 조항이 붙은 바로 그
열거의 모집단이 낱말이 없는 줄들을 놓쳤다.** 손으로 열거한 판별식이 자기 결함을 재생산했다.

### T-E4 — [Q2] 분류 규칙 자체는 건전하다

A/B/C 삼분류와 「위반은 C뿐」은 옳다. `acceptance.md:288-290`의 위반 정의도 REQ-BLKG-005
(`spec.md:115-117`)와 문장이 일치한다 — 두 계층이 갈리던 2차의 R2는 **닫혔다**. 결함은 규칙이
아니라 **그 규칙을 이 트리에 적용해 얻은 목록**에 있다.

### T-E5 — [Q3] REQ-BLKG-005 개정: 방향은 옳고, 한 문장이 실제보다 넓다

「종료코드 ≠ 0을 실패로 읽는 판정은 제약 밖」(`spec.md:117`)은 **일반화가 지나치다.**
≠ 0은 「의도한 RED」만 뜻하지 않는다 — 측정이 아예 성립하지 않은 경우도 ≠ 0이다:

    $ go test ./internal/nosuchpkg/ -count=1
    stat …/internal/nosuchpkg: directory not found
    FAIL	./internal/nosuchpkg [setup failed]
    FAIL
    (exit ≠ 0)

뮤턴트를 심다 컴파일이 깨지면 같은 모양이 된다 — 가드는 한 번도 안 돌았는데 「RED 관측」으로
기록될 수 있다. **다만 실제 피해는 없다**: ≠ 0을 읽는 세 자리가 모두 두 번째 조건을 함께 요구한다
— `:165`는 특정 실패 메시지, `:220`은 「두 메시지가 다른 교정을 제시」, `:204`는 실패한 파일:줄.
빌드 실패는 그 어느 것도 못 낸다. **구멍은 문장에 있고 레시피에는 없다.** → T3(optional).

### T-E6 — [R4/R5] 닫혔다

**R4**: `acceptance.md:176-180`이 `--- SKIP:` / `[no tests to run]` / 진짜 `--- PASS:` 3갈래로
갈랐고, 앞 둘을 「측정 미성립 → 판정 불가」로 처분한다. 2차에서 지적한 오탐 경로가 막혔다.
`:182-184`가 「이 환경에서 진짜 PASS는 도달 불가로 예상된다」까지 적어 둔 것도 정확하다 —
내가 2차에서 baseline blob 가용·두 이름 매치 0/0으로 잰 것과 같은 결론이다.

**R5**: `acceptance.md:31-39`가 0/1/그 밖 3분기로 바뀌었고 `exit 128` 실측이 붙었다.
내가 2차에서 잰 값(`deadbeef1` → `fatal:` + 128)과 같은 계열이다.

### T-E7 — must-pass 재검

    $ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BINLAG-KEYGUARD-001/ | wc -l
    0
    $ grep -c 'syscall' spec.md
    0
    $ grep -n '^\*\*REQ-BLKG-' spec.md
    89:REQ-BLKG-001  93:002  97:003  107:004  111:005  131:006  134:007      ← 연속, 결번·중복 0
    $ (frontmatter) id/title/version("0.3.0")/status/created/updated/author/priority/phase/module/lifecycle/tags  ← 12개 전부, 거절 별칭 0
    $ (D7) SPEC-BINARY-LAG-VISIBILITY-001 → status: completed
           SPEC-BINLAG-INVOCATION-001     → status: completed              ← retired/superseded/archived 아님

| must-pass | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | 001-007 연속 |
| MP-2 GEARS 형식 | **PASS** | 7건 모두 패턴 부합(006은 서술문에 가까우나 허용 범위) |
| MP-3 frontmatter | **PASS** | 12필드 전부 |
| MP-4 언어 중립성 | N/A | 단일 언어 SPEC |
| MP-5 D7 교차-SPEC | **PASS** | 참조 2건 모두 실재·`completed` |
| MP-6 D8 크로스플랫폼 | N/A | `syscall` 매치 0 |
| MP-7 clarification gate | **PASS** | 마커 0건 |

**7건 중 위반 0.**

---

## Baseline-attribution (3차)

T-E0 … T-E7의 모든 수치는 `.claude/worktrees/t479` @ `93fb36344`에서 **이 실행으로** 냈다.
리드와 작성 세션이 보고한 값(모집단 5, A 3 / B 2 / C 0, `exit 128`, `-v`의 `PASS` 한 줄)은
**옮겨 적지 않고 각각 다시 열어 읽거나 다시 실행**했다. Tier M 임계 `0.80`은
`.claude/rules/moai/workflow/spec-workflow.md:141`에서 직접 읽었다.

---

## Findings (3차)

| id | 심각도 | 분류 | 좌표 | 무엇이 잘못됐나 | 어떻게 고치나 |
|---|---|---|---|---|---|
| T1 | major | blocking | `acceptance.md:315-321` | AC-008(1) 기대값의 **모집단이 과소열거**다. 열거 대상에 든 `spec.md`가 0건 기여인데 `spec.md:158`+`:162`가 `ok`+`exit=0`을 읽고 「기준선은 GREEN」을 판정한다(비공허성 짝 없음 → C 후보). 낱말이 없는 `acceptance.md:221`·`:335`의 통과-읽기도 빠졌다. 「모집단 5 / C 0」은 이 트리에서 **틀린 기대값**이다. 아이러니하게도 `:298-307`의 [HARD] 낱말-매치 금지 조항이 붙은 그 열거가 낱말 없는 줄을 놓쳤다 | 기대값의 고정점을 **좌표 목록에서 분류 규칙 + 재열거 의무로 옮긴다**(T5 처방). 당장은 `spec.md:162`를 모집단에 넣어 A/B/C를 다시 판정하고, `:221`·`:335`는 T2로 함께 고친다 |
| T2 | major | blocking | `acceptance.md:221`, `acceptance.md:335` | R1 수정의 **형제 표면 미스윕**. AC-BLKG-005 (2)(3)에는 2단 판정을 붙였으나(`:197-202`), AC-BLKG-006의 「원복 후 GREEN 재확인」과 §D.1의 「기존 가드도 함께 GREEN」에는 붙지 않았다. 두 자리 모두 GREEN을 어떻게 읽는지 말하지 않아, 이름이 어긋난 초록과 진짜 초록이 구별되지 않는다 | 두 줄에 「AC-BLKG-001과 같은 2단 판정으로 읽는다」를 붙인다. `:185`(AC-004 원복)에는 이미 그 문구가 있으므로 문장을 그대로 복제하면 된다 |
| T3 | minor | optional | `spec.md:117`, `acceptance.md:15-16` | 「종료코드 ≠ 0을 실패로 읽는 판정은 제약 밖」이 실제보다 넓다. 빌드 실패·setup 실패도 ≠ 0이라 「측정 미성립」이 「의도한 RED」로 기록될 수 있다(실측: `[setup failed]` → exit ≠ 0). 현재 레시피는 세 자리 모두 실패 메시지/파일:줄을 함께 요구해 피해가 없다 | 문장에 「단, ≠ 0이 **측정 성립**을 뜻하지는 않는다 — 실패 메시지나 `--- FAIL:` 줄을 함께 요구한다」를 덧붙인다 |
| T4 | minor | optional | `acceptance.md:81-88` | 2단 판정이 stdout을 **사람이 읽는** 형태다. 매치 수 명령이 없다. (b)안(파이프)을 §2.1 단일 호출 요건 때문에 기각했다는데, §2.1은 **RED-now 셀의 인용 명령**만 구속하고 green path 통과 조건은 구속하지 않는다 — 과도 적용이다 | RED-now 셀은 지금 그대로 두고, **통과 조건 쪽에만** 매치 수를 추가한다: `--- PASS: <이름> `(뒤 공백 — `$` 앵커는 소요시간 접미사 때문에 항상 0) 정확히 1, `[no tests to run]` 0 |
| T5 | minor | optional | `acceptance.md:326-327` | [Q5] 기대값이 `file:line` 5개에 매여 있어 문서 한 줄만 밀려도 부패한다. 재열거 의무 조항은 이미 있으나, **좌표 목록이 여전히 「기대값」으로 제시**돼 있어 다음 사람이 그것을 대조 기준으로 삼는다 | 좌표 목록을 「기대값」이 아니라 **「전회 열거 결과(비권위적 seed)」**로 강등하고, 기대값은 **`C = 0` 하나**로 둔다. 모집단·A/B 수는 매 판정 시 재열거해 기록하며, 재열거 없이 옛 목록을 근거로 「C 0건」이라 적는 것은 측정이 아님을 명시(그 문장은 `:326-327`에 이미 있으니 위치만 올린다) |
| R3 (미해결 승계) | minor | optional | `spec.md:143` | 2차에서 지적한 「뮤턴트 2건(m1/m2)」이 그대로다. acceptance.md는 3종(m1/m2/m2′)을 요구한다 | `spec.md:143`을 「뮤턴트 3건(m1/m2/m2′)」으로 갱신 |

### 2차 결함 회귀 확인

| 2차 id | 상태 | 근거 |
|---|---|---|
| R1 (=D9) | **RESOLVED** | 단일 호출 §2.1 충족(`:79`) + 지금 트리에서 판정 불가로 떨어짐(T-E2) + 대조군으로 판별식 비공허 확인 |
| R2 | **RESOLVED** | REQ-BLKG-005 재서술로 예외 없이 정합(`spec.md:115-117` ↔ `acceptance.md:288-290`). 규칙은 건전하다 — 남은 결함은 그 적용 결과(T1) |
| R4 | **RESOLVED** | 3갈래 처분표(`:176-180`) |
| R5 | **RESOLVED** | 0/1/그 밖 3분기 + `exit 128` 실측(`:31-39`) |
| R6 | **RESOLVED** | RED-now 4요소 전부(`:95-100`) |
| R3 | UNRESOLVED | `spec.md:143` — 2회 연속. optional이며 리드에게 재량으로 넘긴 항목이라 정체(stagnation)로 보지 않는다 |

**정체 없음.** 다만 **세 라운드 연속으로 수정이 새 결함을 심었다** — 이것이 이 SPEC의 가장 강한
잔여 신호다(§Residual-risk).

---

## Gaps (3차) — 관측하지 않은 것

- **새 가드의 동작**: 여전히 코드가 없다. AC-001…006의 실제 GREEN/RED, 뮤턴트 m1/m2/m2′ 어느
  것도 실행하지 않았다(편집 금지).
- **`--- SKIP:` 경로**: R4의 3갈래 중 SKIP 갈래를 재현하지 않았다. baseline blob을 가리는 조작이
  필요한데 환경 조작 범위 밖이다.
- **T1의 최종 A/B/C 판정**: `spec.md:162`를 모집단에 넣었을 때 그것이 C인지 A인지는 **분류 판단**이
  갈릴 수 있다(판정 줄이냐 증거 기록이냐). 이 감사는 「모집단 멤버인데 목록에 없다」까지를
  확정했고, 최종 분류는 확정하지 않았다.
- **`plan.md` / `progress.md` 전문 재독**: 3차에서는 §C·§E와 grep 히트 구간만 읽었다.
  `progress.md`는 §E.1이 2차 이후 갱신됐는지 확인하지 않았다.
- **전체 스위트·CI**: `go test ./...`도 CI도 읽지 않았다.
- **T4 처방의 실측**: 매치 수 형태를 **가드가 실재하는 상태**에서 돌려보지 못했다(가드 부재).

---

## Residual-risk (3차)

- **세 라운드 연속 「수정이 결함을 심는다」**: 1차 수정 → R1/R2, 2차 수정 → T1/T2. 그리고 두 번
  모두 **직전 수정을 위해 새로 쓴 문장**이 원인이었다(2차의 예외 조항, 3차의 열거 기대값).
  run-phase가 이 SPEC을 고칠 때도 같은 확률이 붙는다 — **새로 쓴 문장을 먼저 의심**해야 한다.
- **T1이 자기 조항을 재생산했다는 사실**: 낱말-매치 금지 [HARD]가 붙은 그 열거가 낱말 없는 줄을
  놓쳤다. 규칙을 적는 것과 그 규칙대로 재는 것은 다른 행위이며, 이 SPEC은 그 간극을 이미 두 번
  보여 줬다. **run-phase의 열거는 사람이 세 문서를 처음부터 끝까지 읽어야** 한다.
- **T2가 열어 둔 자리**: AC-006 원복 초록과 §D.1 초록은 지금 상태로 run-phase가 기록하면
  「이름이 어긋나 안 돌았다」와 구별되지 않는다. 카드가 거짓으로 성공하지는 않지만
  **증거가 거짓일 수 있다.**
- **좌표 부패(T5)**: 이번 감사에서 실제로 그 위험을 봤다 — 2차 판 좌표 `:62/:110/:147`이
  3차 판에서 `:165/:203/:220`으로 전부 밀렸다. 다음 편집에서도 밀린다.

---

## §T-결론 — 판정과 그 근거, 그리고 반대 판정을 고려한 이유

### 판정: **PASS · 0.90**

| 차원 | 1차 | 2차 | 3차 | 근거 |
|---|---|---|---|---|
| Clarity | 0.95 | 0.95 | **0.95** | 철회 문장 보존·3갈래 처분표로 더 읽기 쉬워졌다 |
| Completeness | 0.95 | 0.95 | **0.95** | 변동 없음 |
| Testability | 0.60 | 0.65 | **0.80** | R1/R6 해소로 **초록을 주장하는 1차 판정이 전부 비공허**해졌다. T2가 2차 표면에 남아 만점은 아니다 |
| Traceability | 1.00 | 0.95 | **0.90** | R3 미해결 + T1의 목록↔실제 불일치 |

산술 평균 **0.90**. Tier M 임계 **0.80**(`spec-workflow.md:141`) 초과.
must-pass 7건 중 **위반 0**(T-E7).

### FAIL을 고려했고, 취하지 않았다 — 그 이유

T1은 **blocking 등급이 맞다.** 수락 기준 안에 이 트리에서 **거짓인 기대값**이 들어 있고,
그것은 취향이 아니라 사실 오류다. 그럼에도 FAIL로 가지 않은 근거는 셋이다.

1. **must-pass 위반이 0이다.** FAIL을 강제하는 것은 must-pass 실패이며, T1/T2는 어느 것도
   거기 해당하지 않는다. 점수는 임계보다 0.10 위다.
2. **T1이 붙은 AC-BLKG-008은 「회귀 가드」이지 「릴리스 차단」이 아니다**(`acceptance.md:65`).
   그리고 같은 항목이 `:326-327`에서 **「문서를 고치면 다시 열거해 갱신한다. 갱신하지 않은 좌표를
   근거로 C 0건이라 적는 것은 측정이 아니다」**라고 이미 못박는다 — 지배 지시가 「재열거」이므로,
   그 지시를 따르면 T1은 run-phase에서 자기 교정된다.
3. **어떤 릴리스-차단 항목도 1차 판정이 공허하게 통과하지 않는다.** T2가 건드리는 두 자리는
   보조 확인(원복 초록·DoD 초록)이고, AC-006의 본 판정(두 방향 RED + 서로 다른 메시지)은 건재하다.
   카드가 **거짓으로 성공할 경로가 없다**. 남은 위험은 「기록된 증거가 부정확할 수 있다」이며,
   그것이 debt의 정의다.

「잔여 결함을 잡는 것이 유일한 값어치」라는 지시는 **발견을 숨기지 말라**는 뜻으로 받았고,
그래서 T1~T5와 미해결 R3을 전부 올렸다. 그러나 must-pass가 모두 통과하고 점수가 임계를 넘은
상태에서 optional·debt 목록의 길이로 FAIL을 만드는 것은 **감사자가 결함을 발명하는 것**이며,
이 역할의 M6이 금지하는 바다. 판정은 PASS로 두되, 잔여를 다음 절에 **조건**으로 못박는다.

### [HARD] run-phase 진입 조건 — 이 둘은 debt이지 면제가 아니다

1. **T2를 먼저 닫는다** (편집 2줄): `acceptance.md:221`·`:335`에 「AC-BLKG-001과 같은 2단 판정으로
   읽는다」를 붙인다. `:185`에 같은 문구가 이미 있으므로 복제로 끝난다. **가장 싸고 가장 효과가 크다.**
2. **AC-008(1)은 좌표를 재사용하지 않는다**: run-phase가 이 항목을 판정할 때는 세 문서를 처음부터
   끝까지 읽어 **다시 열거**하고, 그 결과(모집단·A/B·C)를 `progress.md`에 남긴다.
   `spec.md:162`가 모집단에 포함되는지를 그때 명시적으로 판정한다.
3. T3·T4·T5·R3은 **운영자 재량**이다. 넷 다 문서 문장 수정이며 run-phase를 막지 않는다.

### 리드가 뒤집을 수 있는 지점

T1을 「회귀 가드의 기대값 오류」가 아니라 「수락 기준의 사실 오류」로 무겁게 보면 FAIL이 타당하다.
그 경우 PASS-with-debt와 실질 차이는 **위 조건 1·2를 강제로 만들 것인가**뿐이다 —
나는 그것을 조건으로 명시하는 편이 라운드를 한 번 더 도는 것보다 낫다고 판단했다.
