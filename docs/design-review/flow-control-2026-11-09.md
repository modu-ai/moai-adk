# 흐름 제어 설계 — WIP 상한·PR 전 자체 감사·통합 게이트 판정 op·흐름 지표 (rc.2 A13 #2279)

- 상위: #2153 (EP-A). 근거: 사용자 결정 06:02(2026-11-09) 2건 — #1975 comment 6733458567(claim 범위), `docs/decisions-2026-11-08.md`(WIP 상한 값·수리 인계).
- 카드가 요구한 것: 관측(출처 표기) → 흐름 모델 → 정책 4종(채택/기각 근거) → 제품화(#2279 본 PR) → 기준선 1회 실측.
- 구현 분해: `src/flow.rs`(WIP·게이트 판정, GET-only), `src/main.rs` op 2개, `tools/check_pr_self_audit.py`, `tools/flow_metrics.py`, 본 문서 + `flow-baseline-2026-11-09.json`.

## 1. 관측 (행마다 출처)

| # | 관측 | 출처 |
|---|---|---|
| O1 | 감사 PASS를 받은 PR 7개가 tip 이동만으로 병합 대기에 머돌았다 | 원장 `.moai-cli/manifests/github-issues.json` key `pass-awaiting-sync`(#2261 #2249 #2250 #2253 #2266 #2257 #2271, 2026-11-09 06:0xZ 스냅) |
| O2 | 감사 FAIL이 실제 결함이었던 사례 다수 — 명령 실패를 passed로 기록(#2269), revert→re-apply 데이터 삭제(#2235), /tmp symlink 포함검사 우회(#2259), 공백 콜론·중복 키 적격 판정(#2236) — 레인 자체 check set은 전부 통과 상태였다 | 각 Issue 판정 원문 댓글: https://github.com/modu-ai/moai-adk/issues/2269 , /issues/2235 , /issues/2259 , /issues/2236 |
| O3 | FAIL 1건마다 인계 ACK → 새 epoch claim → 재구성 → 재감사가 이어져 한 레인이 claim 여러 개를 쌓고 lease 만료 위험 | #2203 인계 사례(2026-11-09), 수리 인계 규칙(첫 FAIL 저자·두 번째부터 비저자)은 사용자 결정 06:02 |
| O4 | 병합 1건이 대기 중 PR 전부의 tip 포함을 깨뜨렸다 — 직렬 게이트 + "branch 최신 유지" 규칙의 상호작용 | #2226 병합 뒤 대기열 전체 재동기화(2026-11-09) |
| O5 | 공용 등록부(`src/main.rs` OPERATIONS, `src/contract.rs`) 동시 수정 → 병합 뒤 의미 충돌 | #2251 rc-ci 실패, #2260 동기화 중 #2239×#2238 상호작용 결함 |
| O6 | 열린 PR 수·나이, PR→첫 판정, PASS→병합, FAIL 비율·사유의 정량 기준선은 세션 내내 not_run이었다 | 카드 #2279 본문 "정량 기준선 … not_run"(2026-11-09 기재) |
| O7 | **신규(본 카드 실측)**: 열린 PR 저자가 GitHub 계정 기준으로는 2개(GoosLab 16, app/dependabot 1) — 모든 레인이 한 계정으로 PR을 만들므로 author login 기준 lane 집계는 이 저장소에서 레인을 구분하지 못한다 | `flow-baseline-2026-11-09.json` `per_author` (2026-11-09 06:41Z 측정, `python3 -I tools/flow_metrics.py --owner modu-ai --repo moai-adk --out … --max-gh-calls 4`) |
| O8 | **신규(본 카드 실측)**: 열린 PR 17개 — 정책 적용 즉시 총 상한(8) 초과 상태. branch 이름(issue/N) 기준이면 16개 이슈로 분해된다 | 같은 기준선 `open_pr_count`·`open_issues` |

## 2. 흐름 모델

Little's Law: 진행 중 일 L = 도착률 λ × 평균 소요 W.

- 이 세션의 실패 양상(O2·O3)은 W를 늘린다: PR 뒤에 적대적 검증이 있으면 FAIL 1건이 재구성·재감사 사이클을 더한다.
- O4는 λ를 흔든다: 병합 1건이 대기 전부의 tip을 깨뜨려 재동기화(λ 급증, L 유지 못함)를 유발한다.
- 정책 방향은 두 갈래다.
  1. **L에 상한**을 둔다(범위 2, WIP 상한) — λ를 인위적으로 늘리는 대신 L이 상한을 넘으면 새 일 수용을 거부 신호로 바꾼다. 상한값은 측정값이 아니라 사용자 지정값(레인당 감사 전 PR 1, 전체 8)이다.
  2. **W의 앞부분을 자른다**(범위 3, PR 전 자체 감사) — 적대적 검증(변이·새 clone·가짜 실행 파일·receipt 대조)을 PR 열기 전으로 당겨 FAIL 사이클(O3)을 준다. 독립 감사는 유지(대체 아님)다.
- 범위 4(게이트 판정 op)는 O4의 비용을 낮춘다: 판정을 읽기 전용 op으로 제품화해 리더가 병합 전 상태를 명령 한 번으로 본다. 범위 6(지표)은 L·λ·W의 기준선을 남긴다(O6 해소).

## 3. 정책 4종 — 채택 근거와 구현

### 3.1 WIP 상한 (배정 거부) — 채택

- 근거: O1·O3(대기 누적과 claim 폭증) + 사용자 결정 06:02의 값(레인당 감사 전 PR 1, 전체 열린 PR 8). 기본값은 `DEFAULT_LANE_CAP=1`, `DEFAULT_TOTAL_CAP=8`로 코드에 상수로 들어가며 **사용자 승인값이지 측정 기준선이 아님**을 문서·주석에 명시(#2279 요구).
- 구현: `flow.wip_check` op — 열린 PR 목록 GET 1회로 lane_open·total_open을 세어 `exceeded`와 `verdict: allow|refuse_assignment`를 낸다. **거부 행위 자체는 리더의 것** — 제품은 신호만 낸다(`writes: false`).
- 시험: `tests/flow_cli.rs::wip_cap_refuses_at_boundary_and_allows_below` — 1/1 거부, 0/1 허용, 2/1 거부, 타 레인 8/8 거부(레인 0개여도). 경계 N(도달) 거부·N-1 허용이 AC-2의 N/N+1 요구를 충족한다.
- 변이 실측: 상한 비교(`lane_open >= lane_cap || total_open >= total_cap`) 제거 → 동일 시험 FAILED(2026-11-09 실측, 7 passed 1 failed).

### 3.2 PR 전 자체 감사 산출물 — 채택

- 근거: O2 — 레인 check set만으로는 변이·환경 우회(symlink, 가짜 실행 파일, receipt 위조)를 못 잡았다. 감사 전 이동이 W를 줄인다.
- 절(카드 텍스트 그대로 4항목): (1) 가드마다 변이 1회 실측, (2) /tmp 등 symlink 경로가 아닌 새 clone에서 check set, (3) 외부 host는 가짜 실행 파일 시험, (4) receipt 거짓 통과 점검.
- 구현: PR 본문 필수 절 `## Self-Audit` + 검사 도구 `tools/check_pr_self_audit.py`(stdin·`--file`, 태그 `mutation:|clone:|host:|receipt:` 4줄, mutation 줄은 FAILED 보고 필수). exit 0/1/2.
- 시험: `tests/check_pr_self_audit_test.py` 7건 — 전절 통과, 절 누락, 항목 누락(receipt), 미실측 mutation, 빈 본문, 사용법. 변이 실측: FAILED 요구 제거 → `test_unmeasured_mutation_blocked` FAILED(2026-11-09 실측).
- 독립 감사는 그대로 유지(대체 아님 — 카드 명시).

### 3.3 통합 게이트 판정 op (읽기 전용) — 채택

- 근거: 결정 06:02 — 게이트 규칙(tip 포함, 감사 head 대비 변경, claim 줄, CI, head 안정)을 제품 코드로 옮기되 **status 게시·병합은 리더 경로 유지**(판정만).
- 구현: `flow.gate_verdict` op — 입력 PR 번호·감사 head·기저 ref, GET-only로 6 검사: `tip_contained`(뒤처짐=update-branch 필요), `changed_since_audit`(감사 head 이후 파일 변화), `commit_claim_lines`(레인 작성 커밋 각각 `Moai-Claim:` — **merge 커밋은 기록이므로 면제**, 부모 2개 이상), `pr_body_claim`, `ci_status`(commit status success), `head_recheck`(재조회 동일). 실패 시 `reasons`에 검사명+사유, `verdict: ok|blocked`.
- gh 호출 수: PR 2회(조회+재조회), 커밋 목록 1회, status 1회, compare 1~2회 — 최대 7 GET, 페이지네이션은 10페이지 상한. Issue별 반복 조회 없음.
- 시험: `tests/flow_cli.rs` gate 7건 — ok(merge 면제 포함), 뒤처짐, 감사 뒤 변경(파일 목록), claim 없는 레인 커밋(sha 나열, merge는 비지명), PR 본문 claim 누락, CI 실패, head 이동. 변이 실측 2건: tip 포함 검사 제거 → behind 시험 FAILED, claim 줄 검사 제거 → claim 시험 FAILED(2026-11-09 실측).

### 3.4 흐름 지표 (읽기 전용) — 채택, 호출 상한 포함

- 근거: O6 — 기준선 부재. gh 호출 수 상한(카드 요구)을 `GhCounter`로 구조화: 모든 gh 호출이 카운터를 지나고 상한 초과 호출은 시작 전 거부(exit 1).
- 구현: `tools/flow_metrics.py` — `gh pr list --state open --json` **1회**로 열린 PR 수·나이·저자별 수·branch 기준 이슈 목록을 측정, JSON 파일로 기록. PR→첫 판정·PASS→병합·FAIL 비율·사유·동기화 횟수는 `not_run` 배열에 명시(§4 R1).
- 시험: `tests/flow_metrics_test.py` 6건 — 카운트·나이(`--now` 고정), ceiling 0 거부, ceiling 1 정확히 충족, gh 실패 exit 1, 사용법, 정확히 1회 호출(두 번째 호출 시 stub이 exit 7). 변이 실측: 상한 검사 제거 → ceiling-0 시험 FAILED(2026-11-09 실측).

## 4. 기각안 (근거)

| 안 | 내용 | 기각 근거 |
|---|---|---|
| R1 | PR→첫 판정·PASS→병합·FAIL 비율을 제품 지표로 실측 | Issue 댓글 크롤링이 필요해 Issue별 gh 호출 반복 → 호출 상한 위반(범위 6 요구와 정면 충돌). 기준선에서 `not_run`으로 기록하고 정책 적용 뒤 재측정 카드에서 방법(예: 리더 원장 기반) 재설계 |
| R2 | status 게시·자동 병합·ruleset 변경을 제품 코드로 | 사용자 결정 06:02 — 판정·제안만, 행위는 리더 경로(#2279 blocked 조항). 두 op 모두 `writes: false` |
| R3 | 흐름 제어를 새 crate로 분리 | 리더 제약(새 crate 금지). `src/flow.rs` 모듈 + `src/main.rs` op 등록으로 충분 |
| R4 | draft PR을 WIP 집계에서 제외 | 현재 저장소 draft 0개(기준선) — 제외 효과 측정 불가, 구현 단순성 우선. 정책 적용 뒤 재측정 시 재검토 |
| R5 | merge 커밋에도 claim 줄 강제 | 결정 06:02 — 도구 merge는 기록. 게이트의 `commit_claim_lines`는 부모 2개 이상 커밋을 면제하고 fixture로 면제를 검증 |
| R6 | WIP 초과 시 카드 자동 재배정·유휴 레인 자동 감사 지정 | 거부 신호만 제품이 내고 배정 거부·감사 지정은 리더 행위(3.1). 카드의 "제안만" 문구와 일치 |

## 5. 기준선 실측 (AC-4, 2026-11-09 06:41Z)

- 명령: `python3 -I tools/flow_metrics.py --owner modu-ai --repo moai-adk --out docs/design-review/flow-baseline-2026-11-09.json --max-gh-calls 4`
- gh 호출: 1/4(읽기 전용 GET). 결과 파일: `docs/design-review/flow-baseline-2026-11-09.json`.
- 결과: 열린 PR 17, 저자 GoosLab 16·dependabot 1, branch 이슈 16개, 최대 나이 0.18d. **총 상한 8 초과 — 정책 적용 즉시 `refuse_assignment` 상태**. 대기 열의 대부분이 O1의 동기화 대기열임이 수치로 확인됨.
- not_run(기록됨): pr_to_first_verdict, pass_to_merge, fail_ratio_and_causes, sync_count — §4 R1.
- 정책 적용 뒤 재측정: 별도 not_run(결과 파일 notes에 명시).

## 6. claim 범위 정렬 (범위 5) — blocked-by #2255

`claim-verify`(#2255, PR #2270)의 `--hook` 검사가 `--first-parent` 전 커밋을 보는 부분을 결정 06:02 범위(레인 작성 커밋, merge 면제)와 맞추는 작업. **#2255 병합 뒤 본 카드에서 정렬하거나 #2255 쪽에서 반영 시 not_needed 기록.** 본 PR 시점: not_run(대기).

## 7. 제품 한계와 후속

- **lane 집계 한계(O7)**: `flow.wip_check --lane`이 author login 기준이라, 모든 레인이 한 계정(GoosLab)으로 PR을 만드는 이 저장소에서는 lane_open이 전부 합산된다. branch 이름(issue/N)은 레인을 구분하므로, 후속 카드에서 `--branch-prefix` 입력(예: 리더가 lane→issue 번호 목록 전달)이나 원장 기반 매핑을 제안한다. 본 PR은 총 상한 판단에만 wip_check를 쓸 수 있다(레인 상한 판단은 branch 기준 flow_metrics `open_issues`로 보조). not_run: 후속 구현.
- **CI 상태 검사의 범위**: `ci_status`는 커밋 status(`commits/{sha}/status`)의 통합 state만 본다 — 체크 런(check suites)은 본다. 후속 카드 후보. not_run.
- 상한 기본값은 사용자 승인값(2026-11-09 06:02Z)이지 측정 기준선이 아니며, 정책 적용 뒤 재측정으로 재조정한다.
