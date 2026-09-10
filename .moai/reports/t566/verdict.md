# t566 판정서 — codemaps fold 단위를 구속하는 검사 부재: 재현, 그리고 수리 위치는 설계 판단

- 카드: t566 (재현 먼저)
- 트리: `.claude/worktrees/t566`, 브랜치 `WT-fold-unit-guard`, HEAD `cc5006429`(두 번째 부모 = 로컬 develop `92bf71523`)
- 측정: 2026-09-10, lane-6. 모든 측정은 `.moai/reports/t566/repro/` 에 있다.

## 1. 주장 (Claim)

1. **재현됐다.** fold 로 판정된 단위에 codemaps 산문을 넣는 뮤턴트를 적용하면, fold 불변식("판정 뒤에도 히트 0")은 깨진다. 그런데 SPEC-CODEMAPS-REFRESH-002 의 검사 가운데 기계로 다시 돌릴 수 있는 것은 전부 뮤턴트 전과 같은 결과를 낸다. AC-CM2-004(omission 편입), AC-CM2-007 의 omission 규칙, 게이트(`moai graph check`)의 codemaps·citations 계층이 모두 그렇다.
2. **게이트도 이 위반을 보지 못한다.** `moai graph check` 의 codemaps 계층 값은 뮤턴트 전후 모두 `47`(임계 40)이다. 이 계층은 산문 내용이 아니라 스탬프 이후의 트리 변경을 세기 때문이다.
3. **수리 위치는 설계 판단이다.** SPEC-CODEMAPS-REFRESH-002 는 `status: completed` 라 그 AC 는 다시 평가되지 않는다. 결함이 문제가 되는 곳은 **다음** codemaps 갱신이다. 그런데 그 갱신이 어떤 산출물(새 SPEC, `/moai codemaps` 워크플로, 게이트 계층)의 규칙을 따를지는 이 카드가 정할 수 없다. 카드 지시("처방은 카드에서 판단")가 있지만, 셋 중 무엇을 고르느냐에 따라 산출물이 달라지므로 여기서 멈추고 보고한다.
4. **곁가지 두 층을 확인했다.** (1) AC 축의 id 정규식은 접미 문자 id 를 흡수한다(12 대 13). 이는 결함으로 재현된다. (2) REQ 추출 축에서는 접미 문자 id 정의 줄이 추출되지 않는다. 다만 이것은 코드 주석에 **의도된 잔여 형태**로 적혀 있다. 카드가 결함으로 부른 전제는 이 문서화 사실과 함께 판단돼야 한다.

## 2. 증거 (Evidence)

### 2.1 fold 뮤턴트 재현

집계 규약은 SPEC 자신의 것을 그대로 썼다. `census.py` 가 적용한 규약은 셋이다. 단위별 적중 행 수는 §A.3(a) 의 `grep -c -F` 이다. 파일 적중은 AC-CM2-004 의 `grep -rl -F` 이다. 적중 0 패키지 목록은 §A.3(a) 의 `go list` 모집합이다. fold 5·omission 15 단위는 `.moai/reports/t475/verdict.md` 관측 ② 의 최종 판정에서 가져왔다.

| 측정 | 뮤턴트 전 (`census-baseline`) | 뮤턴트 후 (`census-mutant`) |
|---|---|---|
| fold 단위 중 적중 0 | 5 / 5 | **3 / 5** (`internal/core/git` · `internal/kanban/prlink_landedref.go` 가 적중) |
| omission 단위 중 파일 적중 없음 (AC-CM2-004 FAIL 조건) | 0 | 0 → **AC-CM2-004 통과 조건 유지** |
| 적중 0 패키지 목록에 남은 omission 단위 (AC-CM2-007 FAIL 조건) | 0 | 0 → **AC-CM2-007 통과 조건 유지** |
| 적중 0 패키지 목록에 남은 fold 패키지 | `internal/core/git` | 없음 — 목록에서 사라졌지만 이를 요구하는 검사가 없다 |
| 적중 0 패키지 수 | 38 | 37 |
| `go list` 패키지 수 / exit | 137 / 0 | 137 / 0 |

- 뮤턴트: `mutant_apply.py` 가 `modules.md` 끝에 fold 단위 둘을 기술하는 문장 하나를 붙였다. 패키지 입도 하나(`internal/core/git`), 파일 입도 하나(`internal/kanban/prlink_landedref.go`)다. t475 첫 통과가 저지르고 되돌린 위반과 같은 모양이다.
- 게이트 (`go run ./cmd/moai graph check --json`, 이 트리에서 빌드):

| 계층 | 뮤턴트 전 | 뮤턴트 후 |
|---|---|---|
| codemaps | stale, 47 / 40 | stale, 47 / 40 |
| citations | fresh, 0 / 0 | fresh, 0 / 0 |
| mx-index · edges | absent (새 워크트리 예상 상태) | absent |
| exit | 1 | 1 |

  두 실행의 계층 값이 같다. codemaps 계층이 뮤턴트 전부터 stale 인 것은 t475 스탬프 이후 develop 변경 때문이며, 이 카드와 무관하다.
- 원복: `mutant_revert.py` 로 저장해 둔 원본을 되돌렸다. 원본 `d57e59f5…7c97`, 뮤턴트 `db1cb4ca…4d2f`, 복원 `d57e59f5…7c97`, `RESTORED_MATCHES_ORIGINAL True`(`mutant-revert.txt`). 원복 뒤 `git status` 에는 `.moai/reports/t566/` 만 새로 보인다.
- 기계로 다시 돌리지 않은 검사: AC-CM2-002(판정 행 수 = 후보 수)는 t475 증거 파일을 세는 명령이라 codemaps 산문과 무관하다. AC-CM2-006·008 은 증거 파일의 표를 대상으로 한다.

### 2.2 곁가지 (1) — AC id 정규식

`ac-id-regex-count.txt`: SPEC-CODEMAPS-REFRESH-002 `acceptance.md` 에서 서로 다른 id 의 수를 셌다. canonical `AC-([A-Z0-9]+-)*[0-9]+` → **12**, 접미 허용 `AC-([A-Z0-9]+-)*[0-9]+[a-z]?` → **13**. 차이는 `AC-CM2-003a` 하나다. canonical 형태는 `.claude/agents/moai/plan-auditor.md` Group A 검사의 세 번째 `Grep` 줄에 있다(develop 판독에서 227행. sync-audit 는 t475 시점의 228행으로 인용했다). sync-audit 가 적은 배포 템플릿·codex 방출본 사본도 같은 줄로 있다고 기록돼 있지만, 두 사본은 이번에 다시 읽지 않았다.

### 2.3 곁가지 (2) — REQ 추출

`req-extract-suffix.txt`: `internal/spec/lint_req_widen.go` 의 `reqLineWidePattern` 을 Python `re` 로 옮겨 적용했다(RE2 전용 문법 없음).

- 대조: `spec.md:305` 의 실제 `REQ-CM2-014` 정의 줄 → MATCH, 추출 id `REQ-CM2-014`.
- 같은 줄에서 id 만 `REQ-CM2-003a` 로 바꾸면 → **NO MATCH.**
- 같은 파일의 주석은 이 형태를 이렇게 적는다: "Known residual forms, deliberately OUT of scope for this widening (a line-based parser must not chase them): lowercase-token suffixes (`REQ-BDR-005b`) …".
- 첫 시도는 `REQ-CM2-014` 가 **언급된** 다른 줄을 대조로 골라 추출 id 가 `REQ-CM2-003` 으로 나왔다. 대조가 틀렸으므로 버리고, 정의 줄만 잡는 선택식으로 다시 쟀다(위 결과). 파일에는 두 번째 결과만 남아 있다.

## 3. 기준 귀속 (Baseline-attribution)

- 모든 측정은 이 워크트리에서, 로컬 develop `92bf71523` 을 병합한 HEAD `cc5006429` 위에서 했다. 게이트는 `go run ./cmd/moai` 로 이 트리에서 빌드했고, 설치된 바이너리는 쓰지 않았다.
- codemaps 파일은 뮤턴트 전후와 원복 뒤 해시로 추적했다(`modules-sha-before.txt`, `mutant-apply.txt`, `mutant-revert.txt`).

## 4. 전제 정정

- 카드는 근거를 "sync-audit **F2**"로 적었다. 그러나 t475 `sync-audit.md` 에서 fold 간극을 다룬 절은 **Flag 2**(142행)다. 같은 파일의 F2 는 AC-CM2-009 regression-guard 처분에 관한 다른 finding 이다(230행). 내용 인용은 맞고 절 이름만 다르다.
- 카드의 곁가지 (2)는 REQ 추출 실패를 "더 나쁜 쪽" 결함으로 부른다. 측정은 카드와 같지만(NO MATCH), 코드는 그 형태를 의도적으로 범위 밖에 두었다고 문서화한다. 이것을 결함으로 고칠지, 문서화된 한계를 유지하고 SPEC 작성 규칙에서 접미 문자 REQ id 를 막을지는 판단이 필요하다.

## 5. 미검증 (Gaps)

- 수리 방향을 고르지 않았으므로 새 검사를 만들지도, 그 검사의 RED/GREEN 을 관측하지도 않았다.
- AC id 정규식의 템플릿·codex 사본 두 곳은 이번 실행에서 다시 읽지 않았다.
- REQ 추출 실패가 lint 출력에서 실제로 조용한지(어떤 finding 도 나오지 않는지)는 `moai spec lint` 를 접미 id SPEC 픽스처로 돌려 확인하지 않았다. 확인한 것은 추출 정규식 한 줄의 매치 여부다.
- 뮤턴트는 파일 하나(`modules.md`)에 한 문장을 붙인 형태 하나만 썼다. 다른 문서나 입도(예: 부모 산문을 고치는 형태)는 재지 않았다.

## 6. 잔여 위험 — 리드가 정할 설계 판단

fold 불변식을 어디에 둘지 선택지는 셋이다. 이 카드는 고르지 않는다.

1. **다음 codemaps 갱신 SPEC 의 AC 로 둔다.** 저비용이지만, 다음 SPEC 작성자가 이 AC 를 기억해야 한다. t475 에서 불변식이 실행자 머릿속에만 있었던 것과 같은 구조가 한 층 위로 옮겨질 뿐이다.
2. **`/moai codemaps` 워크플로(스킬)의 검증 단계로 둔다.** 갱신 절차에 붙어 매번 돈다. 다만 fold 판정 목록이 어디에 기록되는지(판정 결과의 저장 위치)가 먼저 정해져야 기계 검사가 가능하다.
3. **`moai graph check` 에 계층을 추가한다.** 가장 강하지만, 판정 결과를 읽을 입력 형식이 필요하고 Go 코드 변경이다.

어느 쪽이든 카드가 요구한 **입도 결정**이 따라온다. 패키지 입도 fold 는 적중 0 패키지 목록에 나타나지만, 파일 입도 fold 4개는 그 목록에 원래 나타나지 않는다(이번 측정에서도 목록에 남은 fold 는 `internal/core/git` 하나였다). 따라서 검사는 목록이 아니라 판정 단위 전수에 대해 `grep -c -F` 적중 0 을 확인하는 형태여야 한다.
