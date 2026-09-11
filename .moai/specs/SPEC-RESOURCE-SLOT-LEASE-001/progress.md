# SPEC-RESOURCE-SLOT-LEASE-001 — progress (card t607)

plan 단계 산출물을 트리 `c4ce42eca` @ `WT-heavy-test-slot`(worktree t607)에서 작성했다. Tier M. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- 산출물(Tier M): spec.md, plan.md, acceptance.md, progress.md(이 파일).
- SPEC ID 정규식 검사 — 실행한 명령과 출력:
  `ID="SPEC-RESOURCE-SLOT-LEASE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`
- ID 고유성: 이 워크트리 `.moai/specs/`에서 `RESOURCE`·`SLOT`·`LEASE` 이름을 가진 디렉터리는 `SPEC-SYNC-SHA-SLOT-FORMAT-001` 하나뿐이며 ID가 다르다.
- 프론트매터: 12개 정식 필드 + `tier: M`. `status: draft`. `phase`는 릴리스 대상(`"v3.2.0 target"`)이다.
- 요구사항: REQ-RSL-001..016(GEARS, IF/THEN 없음, Tier M 상한 16 이내).
- 인수 기준: AC-RSL-001..016(Tier M 상한 16 이내). 릴리스 차단 기준의 RED-now 셀은 acceptance.md 증거 원장(EL-1..EL-6)을 인용한다.
- Out of Scope: `### Out of Scope — <주제>` H3 7개, 각각 `-` 항목 보유.
- spec lint — 이 트리(`c4ce42eca`)에서 빌드한 바이너리를 경로로 호출했다(설치본 `ed71054d3`은 HEAD의 조상이라 뒤처진 빌드다: `git merge-base --is-ancestor ed71054d3 HEAD` → 종료 코드 0):
  - `go build -o <scratch>/moai ./cmd/moai` → 성공
  - `<scratch>/moai spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`
  - 음성 대조군: 같은 산출물의 사본에 `phase: plan`을 심고 같은 바이너리로 lint → `ERROR FrontmatterPhaseInvalid ... 1 error(s)`. lint가 이 디렉터리 모양을 실제로 읽는다는 증거다.
  - 참고: 요구 문장을 한국어로만 썼던 첫 초안은 설치본 lint에서 `ModalityUnjudged` 경고 16건을 받았다. 정본 요구 문장을 GEARS 영어로 바꾸고 한국어 설명을 하위 항목으로 옮긴 뒤 위 결과가 나왔다.
- M6(레인 문서 반영) 게이트 상태: `git merge-base --is-ancestor WT-acquire-branch-record develop` → 종료 코드 1(`WT-acquire-branch-record` = `f680dab46`, `develop` = `eb50af5a8`). 게이트 닫힘.

## §E.2 Run-phase Evidence

_<run 단계 대기>_

## §E.3 Run-phase Audit-Ready Signal

_<run 단계 대기>_

## §E.4 Sync-phase Audit-Ready Signal

_<sync 단계 대기>_
