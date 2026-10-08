# SPEC-GFD-PATCHID-VERBATIM-001 — acceptance

> 판정 명령은 모두 카드 워크트리 루트에서 실행한다. 전체 스위트는 CI 몫이다 — 로컬 판정은 패키지 스코프만 한다.

## §A AC ↔ REQ 대응 (traceability)

| AC | maps REQ | 성격 |
|---|---|---|
| AC-GPV-001 | REQ-GPV-001 | release-blocking (RED→GREEN 관문) |
| AC-GPV-002 | REQ-GPV-004 | release-blocking (회귀 무변경) |
| AC-GPV-003 | REQ-GPV-004 | regression-guard (품질 게이트 — release-blocking 제외, §D 셀에 실측 기준선) |
| AC-GPV-004 | REQ-GPV-001 | release-blocking (잔존 grep) |
| AC-GPV-005 | REQ-GPV-003·REQ-GPV-005 | release-blocking (fail-closed) |
| AC-GPV-006 | REQ-GPV-001·REQ-GPV-002 | release-blocking (다중 커밋 보강) |
| AC-GPV-007 | REQ-GPV-001·REQ-GPV-003 | release-blocking (세션 종료 제2 지점 관문) |

배차 지시의 AC-01..AC-05 대응: AC-GPV-001=AC-01, AC-GPV-002=AC-02, AC-GPV-003=AC-03, AC-GPV-004=AC-04, AC-GPV-005=AC-05, AC-GPV-006=보강 추가분.

## §B 기준선 핀

모든 RED-now 관측의 핀 트리는 `cad44a75163b6f0f056551354ef8008fe799090f`(`WT-landing-patchid`)다 — 최초 관측 2026-10-06T21:36Z 이전, 종료코드 보강 재관측 2026-10-07. 행마다 관측 시점을 병기한다. 관측 명령·원문 출력:

- RED 테스트 실패: `go test ./internal/cli/worktree/ -run '^TestLandingPredicateWhitespaceDivergenceKeepsTree$' -count=1 -v` → `sweep: landed="yes" verdict="DISPOSE" reason="", want preserve — a whitespace-divergent tip is not on the ref and the patch-id normalization must not read it as landed` / `--- FAIL`, 종료코드 1 (저작자 재관측 2026-10-07, 파이프 없는 재실행; 원문: `.moai/reports/t1561/red-reproduction.md`).
- AC-GPV-004 RED-now: `grep -c '\-\-stable' internal/cli/worktree/landing_predicate.go` → `2`, 종료코드 0 (AC-GPV-004 의 When 명령과 같은 형태 — 2026-10-07 관측).
- AC-GPV-005 RED-now: `grep -rc 'verbatim' internal/cli/worktree/landing_predicate.go` → `internal/cli/worktree/landing_predicate.go:0`, 종료코드 1 (공백 충실 지원 부재 — 미지원 경로 자체가 없어 fail-closed 계약의 verbatim 쪽이 미구현; 2026-10-07 관측).
- AC-GPV-006 RED-now: `grep -c 'VerbatimUnsupported\|WhitespaceDivergenceMultiCommit' internal/cli/worktree/landing_predicate_test.go` → `0`, 종료코드 1 (보강 관문 부재).
- AC-GPV-007 RED-now: (a) `grep -c 'TestSessionExitWhitespaceDivergencePreserves' internal/cli/session_worktree_landing_test.go` → `0`, 종료코드 1 (세션 관문 부재, 2026-10-07 관측) (b) 제2 지점 프로브 — 단일 커밋 카드 vs 공백 변형 squash 에서 `git cherry master card` → `- a82b4346…`("-" = 동치 = 착지 판정; 관측 주체 plan-audit 1회차 감사자, 같은 트리 cad44a751, git 2.54.0 한정 — `plan-audit-iter1.md` 증거 표 9행) + 제2 지점 코드 직독(`gitBranchLandedReal` ②팔 조기 `return true`, :902-904).
- AC-GPV-003 기준선(regression-guard 실측 셀, 2026-10-07, 트리 `cad44a75163b6f0f056551354ef8008fe799090f`): `go vet ./internal/cli ./internal/cli/worktree/` → stdout 없음, 종료코드 0 · `golangci-lint run ./internal/cli ./internal/cli/worktree/` → `0 issues.`, 종료코드 0.

## §D AC 상세

### AC-GPV-001 — 공백 갈림 tip 을 sweep 가 보존한다 (RED→GREEN)

**Given** 공백만 다른 squash 커밋이 원격 착지 상태로 있는 gfd 픽스처(RED 테스트와 동일 시나리오) **When** `go test ./internal/cli/worktree/ -run '^TestLandingPredicateWhitespaceDivergenceKeepsTree$' -count=1 -timeout 30m` **Then** exit 0, 실패 0. RED-now는 §B 첫 행 — RED-now 셀에 상한 없음: M1 이 뒤집는다.

### AC-GPV-002 — 기존 착지 판정 스위트 무변경 통과

**Given** 수리가 반영된 트리 **When** (a) `go test ./internal/cli/worktree/ -count=1 -timeout 30m` (b) `git diff cad44a751 -- internal/cli/worktree/landing_predicate_test.go` **Then** (a) exit 0, 실패 0 (b) 기존 테스트 함수(F1~F9·LaterChange·CommitCap·GH 2종·RED)의 단언 행 삭제·수정 0 — diff 는 신규 테스트 추가와 RED 테스트 블록 내 주석 변경만 담는다(diff 판독은 progress §E.2 원문으로 남긴다). RED-now: 수리 전 기준선에서 (a)는 RED 테스트 1건으로 실패(§B).

### AC-GPV-003 — vet·lint 클린 (regression-guard — release-blocking 제외)

**Given** 수리가 반영된 트리 **When** `go vet ./internal/cli ./internal/cli/worktree/` 와 `golangci-lint run ./internal/cli ./internal/cli/worktree/` **Then** 둘 다 신규 지적 0. 이 AC 는 keep-clean 게이트로 regression-guard 다 — 기준선이 이미 클린(§B 셀: vet 빈 출력 종료코드 0, lint `0 issues.` 종료코드 0)이라 이 수리가 뒤집을 적색이 존재하지 않고, release-blocking 의 RED-now 채택 규율이 요구하는 출발 관측이 성립하지 않기 때문이다. 판정은 DoD 게이트로 유지된다(클린 상실 = 회귀).

### AC-GPV-004 — 착지 경로에 공백 정규화 patch-id 잔존 0

**Given** M1 이후의 트리 **When** `grep -c '\-\-stable' internal/cli/worktree/landing_predicate.go` **Then** exit 1, 0히트(주석 포함 최엄계). RED-now: §B 2번째 행(2히트).

### AC-GPV-005 — 미지원 git fail-closed

**Given** 지원 감지 시임이 미지원을 답하게 대입된 테스트 **When** `go test ./internal/cli/worktree/ -run '^TestLandingPredicateVerbatimUnsupportedIsFailClosed$' -count=1 -timeout 30m -v` **Then** exit 0이고 `-v` 출력에 이 관문의 명명된 RUN 행 3개가 나타난다 — `TestLandingPredicateVerbatimUnsupportedIsFailClosed/landedbypatchid_errors`(`LandedByPatchID` 오류 cannot answer), `…/sweep_preserves_without_pr`(PR 없으면 sweep preserve), `…/layer3_still_decides`(병합 PR 있으면 계층 3 DISPOSE). **공집행은 실패다** — 출력에 `[no tests to run]`이 있거나 RUN 행이 3개 미만이면 exit 0과 무관하게 FAIL이다. RED-now: §B 3번째 행 + verbatim 구현 부재.

### AC-GPV-006 — 다중 커밋 누적 공백 갈림 보존

**Given** 두 커밋 카드의 누적 변경과 공백만 다른 squash 커밋이 원격 착지한 픽스처 **When** `go test ./internal/cli/worktree/ -run '^TestLandingPredicateWhitespaceDivergenceMultiCommitKeepsTree$' -count=1 -timeout 30m -v` **Then** exit 0이고 RUN 행 `TestLandingPredicateWhitespaceDivergenceMultiCommitKeepsTree` 가 나타나며 — sweep 가 preserve 를 답한다. **공집행은 실패다** — `[no tests to run]` 또는 RUN 행 부재는 exit 0과 무관하게 FAIL이다. RED-now: §B 4번째 행(관문 부재) + AC-GPV-001 의 단일 커밋 형태가 실측 RED.

### AC-GPV-007 — 세션 종료가 공백 갈림 카드를 보존한다 (제2 지점 관문)

**Given** 단일 커밋 카드와 그 squash 커밋을 공백만 변형해 원격 착지시킨 세션 워크트리 픽스처(`session_worktree_landing_test.go` 의 `TestCleanupSessionWorktree_*` 어법) **When** `go test ./internal/cli -run '^TestSessionExitWhitespaceDivergencePreserves$' -count=1 -timeout 30m -v` **Then** exit 0이고 RUN 행 `TestSessionExitWhitespaceDivergencePreserves` 가 나타나며, 세션 출구 정리가 preserve 를 답한다 — ②팔이 공백 정규화 동치로 조기 "착지"를 답하지 않는다. **공집행은 실패다** — 출력에 `[no tests to run]`이 있거나 RUN 행이 없으면 exit 0과 무관하게 FAIL이다. **Green path**: plan §F-M1.4(커밋별 공백 충실 동치 술어 공개)와 §F-M2.3(세션 종료 경로 관문 신설)이 이 셀을 뒤집는다. RED-now: §B AC-GPV-007 행(관문 부재 + 제2 지점 프로브).

## §E 에지 케이스

- **바이트 동일 패치**: `--verbatim` 전환 뒤에도 F1·F2(단일·이중 커밋 squash)가 계층 2 로 착지 확인된다(AC-GPV-002가 덮는다).
- **커밋 상한 초과 + 미지원 git**: 상한 cannot answer와 미지원 cannot answer가 겹쳐도 답은 하나 — preserve, 계층 3 판정.
- **세션 종료 정리(자체 3팔 술어)**: ②팔이 공백 충실 동치로 바뀐 뒤에도 개별 커밋이 upstream 에 각자 존재하는 카드는 계속 확인된다(의미 보존 — REQ-GPV-002). 미지원 git 이면 ②·③팔 모두 오류 → 세션 출구는 보존하고 다음 sweep 의 gh 계층에 맡긴다 — REQ-WSS-304(SPEC-WEB-SETTINGS-SAVE-001) 네트워크 금지와 REQ-WSS-302 오류 시 preserve 계약은 그대로다.
- **`--verbatim` 이 있어도 문맥이 갈린 squash(F5)**: verbatim 패치-id 가 어차피 불일치 → 기존대로 계층 3 이 판정한다.

## §F 품질 게이트·DoD

- 게이트: AC-GPV-001..007 전부 PASS(AC-GPV-003 은 regression-guard 분류로 게이트 유지) + `progress.md` §E.2·§E.3에 5-섹션 증거 기재 + TRUST 5 (Tested·Readable·Unified·Secured·Trackable) 자점 없음.
- DoD: 커밋이 `(card t1561)` + 🗿 MoAI 트레일러를 운반하고, 변경 파일이 `internal/cli/worktree/landing_predicate.go`·`internal/cli/worktree/landing_predicate_test.go`·`internal/cli/session_worktree.go`·`internal/cli/session_worktree_landing_test.go` 4개로 끝난다.

🗿 MoAI
