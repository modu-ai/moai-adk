# t541 verdict — plan-auditor 교차 산출물 순서 모순 감사 항목

트리: `.claude/worktrees/t541`, 브랜치 `WT-auditor-order-conflict`. 문서 커밋 `542f7dbc2`, 테스트·증거는 후속 커밋.
설계 결정(D-1~D-4)은 리드 승인분 — Group 6 CN-4 + 검증 동사, MP-9 신설, 재감사 회차 전체 재실행, 기계적 CONFLICT 는 `Exit:` 결속 plan 에만.

## 주장

1. 수정 전 교리에는 교차 산출물 순서 모순 검사가 없었다.
2. 실제 SPEC 에 동시에 따를 수 없는 순서 의무 쌍이 있었다.
3. 이제 plan-auditor 세 사본(템플릿·로컬·codex)이 Group 6 CN-4, MP-9, 재감사 예외를 함께 담는다.
4. CN-4 동사는 실제 충돌을 CONFLICT 로 잡고, 방향만 뒤집은 쌍은 잡지 않으며, 순서 절이 없으면 NONE, 마일스톤이 없으면 GAP 을 낸다.
5. 새 테스트 15개는 동사의 핵심 분기 4개와 MP-9 문구 하나를 각각 지킨다 — 뮤턴트 5개가 모두 해당 테스트를 빨갛게 만들었다.
6. MP-9 추가가 기존 RED-now 계약 테스트를 깨지 않았고, MP 개수를 세는 다른 테스트·문서는 없다.

## 증거

- 주장 1: `repro/doctrine-absence.txt` — 두 사본 `ordering conflict|cross-artifact|jointly satisfiable` 0건(exit 1). 대조 `repro/doctrine-control.txt` — `### Group` 두 사본 8건.
- 주장 2: `repro.md` §2.2 — plan §F M1(L87)→M2(L113, `Exit: AC-SSF-001 green`), acceptance §D.4 L247 "captured BEFORE the M1 render change".
- 주장 3: 커밋 `542f7dbc2` 4파일. `gen-entry.log`(exit 0, catalog 1줄), `agents-emit.log`·`agents-emit-check.log`(exit 0), 템플릿↔로컬 diff 는 기존 의도 차이 2곳뿐.
- 주장 4: 커밋된 동사를 추출해 실제 SPEC 에 실행 — `repro/run-cn4-committed-t534.txt` CONFLICT 1줄(acceptance.md:247), `repro/run-cn4-committed-stopchain.txt` `Milestone M<n>` 형식에서 5 milestones 수집(시제품은 GAP, `repro/run-control-stopchain.txt`). 픽스처 단위는 `slot/green.txt` 15 PASS.
- 주장 5: `slot/summary.md` 표 — m1~m4 각각 목표 테스트 + VerbIdenticalAcrossCopies FAIL, m5 ClausesPresent FAIL. 복원 cmp exit 전부 0, 해시 재검증 OK, 복원 후 15 PASS.
- 주장 6: `rednow-test.txt` — `internal/spec` RED-now 20개 PASS 20 / FAIL 0, exit 0. `template-test.txt` — `go test ./internal/template/` exit 0. MP 참조 전수 grep: `MP-8`·"eight criteria" 는 plan-auditor 세 사본과 `red_now_cell_test.go`(MP-8 행 접두사·스팬만)에만 있고, 대조 `MP-7` 은 같은 세 사본에 6건.

## 기준선 귀속

- 재현 측정: HEAD `39b32b11d` (수정 전 트리).
- 문서·동사·템플릿 테스트·RED-now: 수정 후 작업 트리, 커밋 `542f7dbc2` 와 동일 내용.
- internal/cli 슬롯: `542f7dbc2` + 미커밋 테스트 파일. 사전 확인에서 다른 internal/cli 프로세스 0(대조 52).
- 모든 grep 은 `/usr/bin/grep`(셸 래퍼 아님). 동사는 awk 만 쓴다.

## 미검증

- 동사는 macOS 기본 awk 에서만 돌렸다. gawk·mawk·busybox awk 실행은 보지 않았다.
- 코퍼스 전체(830 SPEC)에 동사를 돌린 오탐·미탐 측정은 하지 않았다.
- 감사자 LLM 이 CANDIDATE 를 읽고 올바르게 판정하는지는 Go 로 검증할 수 없다.
- 커밋된 동사로 flipped 픽스처를 실제 SPEC 경로에서 다시 돌린 것은 시제품(`repro/run-fixture-after.txt`)뿐이다. 커밋된 동사의 방향 판별은 Go 픽스처(SatisfiableAfterIsClean)로만 확인했다.
- 병합 트리 재측정은 통합 창에서 할 일이다.

## 잔여 위험

- CONFLICT 는 AC·순서 단어·마일스톤이 한 레코드에 있고 plan 이 `Exit:` 로 결속할 때만 난다(코퍼스 plan 765개 중 19개). 나머지 모양은 CANDIDATE 로만 올라와 감사자 판독에 의존한다.
- 순서 단어를 소문자 `before`/`after`/`first` 까지 받으므로 CANDIDATE 잡음이 크다(t534 실사례 5건 중 1건만 실제).
- `Exit:` 뒤 쉼표 목록의 숫자 확장은 뒤따르는 낱말을 보지 않아, `AC-X-001, 3 times` 같은 문장이 plan 의 `Exit:` 줄에 오면 잘못 결속할 수 있다.
- MP-9 는 CONFLICT 줄에 FAIL 을 강제하므로, 결속을 잘못 읽은 CONFLICT 는 거짓 FAIL 이 된다(리드 결정으로 수용).
