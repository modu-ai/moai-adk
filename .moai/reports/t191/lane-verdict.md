# t191 — lane verdict (SPEC-PROJECT-CONTINUATION-KEY-001)

- 레인: lane-4
- 트리: `.claude/worktrees/t191`, 브랜치 `WT-project-continuation`, HEAD `af6d2dcf1`
- 분기 기준: 로컬 `develop` @ `2660bcd09` (배차 시점 `b7462203a`에서 +2 전진한 값)
- 측정일: 2026-09-02

이 문서는 **레인이 독립적으로 재측정한 값**만 담는다. plan-audit 2회와 sync-audit의
판정은 각자의 보고서(`plan-audit.md`, `plan-audit-iter2.md`, `sync-audit.md`)가 소유하며,
여기서 되풀이하지 않는다.

---

## 정정 — 소유권 경계 주장의 baseline

리드 정정(2026-09-02)을 수용해 기록한다. 레인이 완료 보고에서 이렇게 적었다:

> `plan.md`·`acceptance.md` diff **0**, `spec.md`는 1행(status)뿐

**문장은 참이지만 잰 대상이 빠져 있었다.** 그대로 인용되면 "이 브랜치가 SPEC 문서를
1행만 고쳤다"로 읽히고, 그건 아래 두 번째 표와 정면으로 어긋난다. 두 값 모두 이 트리에서
직접 측정했다.

**(A) sync 커밋 기준 — `5645a0ccb..HEAD`**

```
$ git diff --stat 5645a0ccb..HEAD -- .moai/specs/SPEC-PROJECT-CONTINUATION-KEY-001/
 .../progress.md  | 59 +++++++++++++++++++++-
 .../spec.md      |  2 +-
 2 files changed, 58 insertions(+), 3 deletions(-)
```

`plan.md`와 `acceptance.md`는 출력에 **나타나지 않는다** = 이 구간에서 바이트 무변경.
`spec.md`는 1 insertion / 1 deletion = `status:` 한 줄. 이것이 원래 주장이 재려던 것이다 —
**sync-phase에서 manager-docs가 소유권 경계를 지켰는가**. 지켰다.

**(B) develop 기준 — `refs/heads/develop...HEAD`**

```
$ git diff --stat refs/heads/develop...HEAD -- .moai/specs/SPEC-PROJECT-CONTINUATION-KEY-001/
 .../acceptance.md | 227 ++++++++++++++++++
 .../plan.md       | 144 +++++++++++
 .../progress.md   | 173 ++++++++++++++
 .../spec.md       | 266 +++++++++++++++++++++
 4 files changed, 810 insertions(+)
```

신규 SPEC이라 `develop`에 존재하지 않는다 — 4개 파일 전부 신규 작성분이다.

**둘 다 참이고 잰 대상이 다르다.** 소유권 경계를 말할 때는 (A)를, 브랜치가 무엇을 만들었는지
말할 때는 (B)를 인용한다. baseline 없이 (A)만 옮기는 것이 이 정정이 막으려는 오독이다.

---

## 레인 재측정 목록 — 값마다 잰 대상을 붙인다

| # | 주장 | 명령 | 관측 | 잰 대상 |
|---|---|---|---|---|
| V1 | 브랜치 상태 | `git rev-list --count 2660bcd09..HEAD` | `15` | 분기 기준 `2660bcd09` |
| V1b | 작업 트리 | `git status --short` | 빈 출력 | HEAD `af6d2dcf1` |
| V2 | SPEC 종결 | `grep -n '^status:' …/spec.md` | `5:status: completed` | HEAD |
| V3 | 소유권 경계 | 위 (A) | plan/acceptance 부재, spec 1행 | **sync 커밋 `5645a0ccb..HEAD`** |
| V4 | sync SHA 실물 | `grep -n sync_commit_sha …/progress.md` + `git rev-parse --short c7ac03fe8` | `c7ac03fe8` 양쪽 일치 | HEAD |
| V5 | CHANGELOG 중복 | `git diff --stat 5645a0ccb..HEAD -- CHANGELOG.md` / `grep -c 'SPEC-PROJECT-CONTINUATION-KEY-001' CHANGELOG.md` | `1 insertion(+), 1 deletion(-)` / `1` | sync 커밋 구간 / HEAD |
| V6 | AC-PCK-008 델타 | `grep -c "Implementation Kickoff Approval" …/doc-generation.md` | `3` | HEAD (SPEC이 기준선 `1`을 `2660bcd09`에 고정) |
| V7 | 빌드 | `go build ./...` | exit `0` | HEAD |
| V8 | 신규 enforcer | `go test ./internal/web/ -run TestProjectContinuationI18nKeysInAllLocales -count=1` | `ok … 0.709s` | HEAD |
| V9 | 차분 쌍 | Step 4.2 표 + 세 값 블록 원문 판독 | 아래 인용 | HEAD |
| V10 | F2 회귀 가드 부재 | `grep -rn "doc-generation" --include='*_test.go' internal/` | `0` | HEAD, 범위 `internal/` |
| V11 | 임시 파일 | `find . -name 'i18n_step*.py'` / `git ls-files \| grep -c i18n_step` | 빈 출력 / `0` | HEAD |
| V12 | 발산 | `git rev-list --count --left-right refs/heads/develop...HEAD` | `50 15` | 로컬 `develop` 현재 팁 |

**V9 인용** — 차분 쌍이 기계적으로 성립함을 보이는 근거다. `doc-generation.md` Step 4.2:

- `card` 행: "**The session stops when `/moai plan` returns**; run-phase entry is a
  separately-initiated operator action in some later turn, and this branch neither proceeds
  to it nor emits its gate."
- `pipeline` 행: "**The session does not stop when `/moai plan` returns**: it continues past
  plan completion to the run-phase boundary and **emits the Implementation Kickoff Approval
  gate in this same session**"

한 행 텍스트가 둘을 동시에 만족할 수 없다. `AC-PCK-005`(card는 `/moai plan`에서 종료)와
`AC-PCK-006` 결합 1(pipeline은 그것을 지나감)이 상호배타이므로, `pipeline`을 문구 변경으로만
구현하면 둘 중 하나가 반드시 red가 된다.

**V10의 성격** — 부재 주장이므로 범위를 밝힌다. `internal/` 아래 `*_test.go` 전체에서
`doc-generation`을 참조하는 파일이 0건이다. 이는 sync-audit F2가 지적한 것과 같은 관측이며,
차분 쌍이 **기준으로는** 성립하지만 **회귀 가드로는** 아무것도 지키지 않는다는 뜻이다.

---

## 레인이 관측하지 않은 것 (Gaps)

- **G1** AC 14건 중 레인이 개별 재측정한 것은 하중을 받는 일부다(V6·V8·V9). 나머지는
  sync-audit이 14건 전부를 재측정했다고 보고했고, 레인은 그 보고를 **재현하지 않았다**.
- **G2** `make build`와 `cmp` 미러 행렬을 레인이 직접 돌리지 않았다. run-phase와 sync-audit의
  측정에 의존한다.
- **G3** 커버리지·lint 수치를 레인이 재지 않았다. sync-audit이 선재로 귀속했다고 보고했다.
- **G4** `GOOS=windows` 빌드를 레인이 직접 돌리지 않았다.
- **G5** 전체 스위트를 로컬에서 돌리지 않았다(레인 규율 — 여러 레인이 공유하는 머신).
  이 브랜치에 대한 CI 판정은 **존재하지 않는다**: 미푸시 상태이고, push는 리드의 일괄 행위다.

## 잔여 위험

- **R1** `pipeline`의 동작은 전적으로 오케스트레이터가 읽는 산문에 산다. `pipeline`인데도
  오케스트레이터가 `/moai plan`에서 멈추는 것을 막는 기계 장치가 없다. sync-audit F1은 여기에
  더해 **`none`·`card`는 안전하게 실패하는데 `pipeline`만 조용히 실패한다**는 비대칭을
  지적했다 — 미매칭 값에는 정확히 그 이유로 보고 줄을 넣었으면서 이 축에는 넣지 않았다.
- **R2** V10의 가드 부재 때문에, 나중 편집이 두 행을 다시 합쳐도 어떤 기준도 red가 되지 않는다.
- **R3** 병합 시 CHANGELOG가 t315(`WT-release-notes-gitflow`)의 형제 편집과 충돌할 수 있다.
  양측 항목을 한 줄도 지우지 않는 처리가 필요하다.
- **R4** 이 문서의 모든 수치는 HEAD `af6d2dcf1`과 로컬 `develop` 현재 팁에 귀속된다.
  병합 창에서 흡수를 마치면 **트리가 달라지므로 재측정이 필요하다** — 창 밖에서 잰 값을
  창 안의 근거로 재사용하지 않는다.

---

## 후속 카드 후보 (레인이 발행하지 않음 — 발행은 운영자 소관)

- **F1 축** — `pipeline`의 조용한 퇴화에 보고 신호를 넣을 것인가. 미매칭 값 경로가 이미 그
  형태를 갖고 있어 선례가 있다.
- **F2 축** — 차분 쌍에 회귀 가드를 세울 것인가. 산문 계약을 읽어 두 행의 종단 지시가 갈리는지
  단언하는 테스트가 형태로 가능하다.
