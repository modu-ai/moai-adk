# SPEC-DOCS-TABCOUNT-DRIFT-001 — progress

카드 t530 / 워크트리 `.claude/worktrees/t530` / 브랜치 `WT-web-tab-docs` / base `1d150a27d`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, 열거 산출물 `.moai/reports/t530/tab-count-sites.md`,
  그리고 인수 판정의 패턴 원본 `.moai/reports/t530/count-literals.txt`(16줄).
- 기준값 14는 이 트리에서 재도출했다. **원천 1(구현) + 미러 1(테스트 리터럴) + 둘의 일치 테스트 1** 이며,
  "서로 독립인 세 측정" 이 아니다(`spec.md §1.1`).
- 미결 없음: llm 탭 이름은 2026-09-12 에 "문서를 코드에 맞춘다" 로 해소됨 — 근거는 `plan.md §C`.
- 범위 판단 1건: 스크린샷 재생성은 별건으로 분리, 근거는 `spec.md §5`.

### plan-audit iter1 (FAIL 0.66) 수리 라운드 — 재측정 기록

감사 보고: `.moai/reports/t530/plan-audit.md`. 아래 수치는 **수리 시점에 다시 잰 값**이다(base `1d150a27d`).

| 측정 | 명령 | 관측 |
|---|---|---|
| A군 리터럴 적중 | `head -4 count-literals.txt > /tmp/t530-a.txt; grep -rnF -f /tmp/t530-a.txt <4파일> \| wc -l` | `4` (en 포함 — 종전 창 기반 정규식은 3) |
| 전체 리터럴 적중 | `grep -rnF -f count-literals.txt <12파일> \| wc -l` | `20` (집합이 열거 20자리와 동일) |
| D군 이름 적중 | `grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' <8파일> \| wc -l` | `12` |
| 가드 층 숫자 스윕(낱말 경계 포함) | `grep -rnE '[0-9]+[^.。]{0,12}(\btabs?\b\|탭\|タブ\|标签页)' <12파일>` | `15`행 = 열거 14 + 허용 1(`ja/advanced/moai-web-console.md:149`) |
| 낱말 경계 없는 스윕의 오탐 | 같은 스윕에서 `\b` 제거 | `README.md:698` (`selec**tab**le`) 추가 적중 |
| 가드 선택자(현재) | `go test ./internal/web/ -run 'TestDocsTabContract' -v` | `testing: warning: no tests to run` / `PASS` / `ok … [no tests to run]`, 종료코드 `0` |
| 패리티 대조군 | `git diff --name-only 1d150a27d -- <README 4본> docs-site/content \| wc -l` | `0` → 측정 불가 = AC-TCD-010 FAIL |
| 스크린샷 | `git diff --quiet 1d150a27d -- assets/images/` | 종료코드 `0` (무변경) |
| 열거 행 수 | `grep -c '^| [ABCD][0-9]' tab-count-sites.md` | `32` (28 → 32, D군 산문 4자리 승격) |
| 열거 경로 완전성 | 표에서 뽑은 고유 문서 경로 수 + 실재 확인 | `12`, 누락 없음 |

**Gap(수리 라운드에서 관측하지 못한 것)** — 조용히 통과시키지 않는다:

- **가드 자체의 동작**: `internal/web/docs_tab_contract_test.go` 는 아직 없다. AC-TCD-004~009 는 실행으로 판정하지 못했고,
  RED-now 셀은 "가드 부재" 로만 기록했다. 변이체 A(README 만 읽는 가드)·B(서브테스트 없음)가 실제로 막히는지는
  가드 작성 후 M5 에서 확인해야 결론이 확정된다.
- **hugo 무경고의 인과**: 수정 후에도 무경고인지는 수정이 없어 확인 불가(RG-TCD-002).
- **넓힌 정규식의 새 미열거 적중**: 낱말 축을 열면 `ko:149 탭 두 곳`·`en:149 two tabs`·`zh:149 两个标签页` 가 모두 적중함을
  실측했다. 이 때문에 낱말 축은 정규식에 넣지 않기로 했고(`plan.md §A.4`), 그 결정의 대가는 `spec.md §7` 에 적었다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
