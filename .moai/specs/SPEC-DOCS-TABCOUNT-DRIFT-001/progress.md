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

### plan-audit iter2 (FAIL 0.775) 수리 라운드 — 재측정 기록

감사 보고: `.moai/reports/t530/plan-audit-iter2.md`(iter1 보고 `plan-audit.md` 는 보존). HEAD `28c1ea062`.
왼쪽 끝은 `git merge-base develop HEAD` = `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09`(측정 시점 값).

| 측정 | 명령 | 관측 |
|---|---|---|
| **낱말 축 원시 스윕** | 로케일별 수사 클래스 + 낱말 경계 탭 명사, 대상 12파일 | **8행** |
| − 서수 접두(`第`) 제외 | 같은 스윕에서 `第[一二三四五六七八九十]` 치환 후 재적용 | **7행** (`zh/advanced/…:165` 의 `第三方` 빠짐 — 단독 검증도 `0`) |
| − `codex` 허용 규칙(**파일 범위 포함**) | `grep -vE '^docs-site/content/[a-z]+/advanced/moai-web-console\.md:[0-9]+:.*codex'` | **6행 = A2·A4·B4·C1·C2·C4**, 허용되지 않은 오탐 **0** |
| 같은 규칙을 **파일 범위 없이** 적용 | `grep -v 'codex'` | **3행** — README 3자리가 함께 면제된다(파일 범위가 필수인 이유) |
| 허용 규칙의 면제 표면 | `grep -ci codex` × 4 console 파일 | 각 **4**, 합계 **16줄** (규칙 수 1 과 다른 값) |
| merge-base | `git merge-base develop HEAD` | `1d150a27d4c5…` |
| 좁힌 pathspec, 카드 기여 | `git diff --name-only <merge-base> -- <12파일>` | **0** → 측정 불가 = AC-TCD-010 FAIL (붉은 이유가 "이 카드가 아직 안 고쳤다" 로 바뀜) |
| 좁힌 pathspec, develop 쪽 churn | `git diff --name-only <merge-base> develop -- <12파일>` | **8파일** — 흡수 전에는 보이고 흡수 후 merge-base 전진으로 사라지는 값 |
| 넓은 pathspec 대비(이전 판의 결함) | `git diff --name-only 1d150a27d develop -- <README 4본> docs-site/content` | **155파일**, `1d150a27d..develop` **40커밋** |
| 스크린샷 | `git diff --quiet <merge-base> -- assets/images/` | 종료코드 `0` (무변경) |
| 열거 행/경로(같은 영역) | `grep '^| [ABCD][0-9]'` → 행 수 / 그 행에서 뽑은 경로 수 | **32행 / 12경로**, 누락 0 |
| 구조 불변식 | REQ / AC / RG / `### Out of Scope —` 계수 | **9 / 12 / 3 / 5**, `status:` 는 `spec.md` 에만, clarification 토큰 0 |

**Gap(이 라운드에서 관측하지 못한 것)**:

- **가드는 여전히 없다.** AC-TCD-004~009·012 의 가드 쪽 단정은 명세 수준 논증이며, 낱말 축·서수 제외·파일 범위
  허용 규칙은 **shell 재현**으로만 확인했다. Go 구현이 같은 집합을 내는지는 M1 에서 확인해야 결론이 확정된다.
- **병합 후 거동은 시뮬레이션이다.** develop 을 실제로 흡수하지 않았다(감사 트리를 변형하지 않는다는 같은 이유).
  "흡수 후 merge-base 가 전진해 남의 커밋이 범위에서 빠진다" 는 git 의 성질이지 이 트리에서 관측한 사실이 아니다.
- **hugo 는 이번에도 실행하지 않았다**(RG-TCD-002).
- **낱말 수사 클래스의 완전성은 미확인.** 오늘 0 오탐을 낸 클래스가 앞으로 들어올 표기까지 덮는다는 보장은 없다 —
  `spec.md §7` 의 유지 비용 항목이 이 미확인을 가리킨다.

### plan-audit iter3 (PASS 0.8375) — 판정과 인계되는 부채

감사 보고: `.moai/reports/t530/plan-audit-iter3.md`. 점수 추이 0.66 → 0.775 → **0.8375**(Tier M 임계 0.80),
must-pass 실패 0. D1~D4 전부 REPAIRED 이며, 그중 셋은 감사가 문서 독해가 아니라 **실행**으로 확인했다 —
재현 명령 5 축어 실행 6행, 스크래치 트리에서 카드 처분대로 편집 후 6 → 0(green path 실재),
그 초록 트리에 V4 변이 주입 시 1행 적중(iter2 가 통과시킨 변이체가 이제 잡힌다).

**[부채 N1 — run-phase M1 의 입구 작업] 가드의 낱말 *클래스* 를 구속하는 AC 가 없다.**

`REQ-TCD-005` 는 "숫자와 낱말 모두" 를 명시하지만, 그것을 기계적으로 구속하는 자리가 없다:
`AC-TCD-012` 의 When 은 셸 파이프라인이어서 **셸**을 구속하고 Go 가드를 구속하지 않으며,
`AC-TCD-006 V4` 는 변이가 `fourteen` 한 건이다. 실측으로도 acceptance.md 가 이름을 대는 수사 낱말은
`nine` / `fourteen` / `열네` 뿐이고 `十四`·`九个`·`아홉` 은 어디에도 없다.

결과: 낱말 클래스를 `{fourteen}` 으로만 구현한 가드가 전 기준을 통과하면서 6자리 중 5자리를 놓친다 —
**이 카드가 지난 청소의 실패 원인으로 지목한 바로 그 부류**다.

네 번째 감사 반복은 돌리지 않는다(상한 3 도달, 임계는 이미 넘었다). PASS-with-debt / 범위 축소 /
운영자 override 는 임계 **미달** 시의 선택지라 여기 해당하지 않는다. 고칠 것은 `acceptance.md` 한 문장이고
그것이 구속하는 대상이 M1 이 지금 쓰려는 가드 파일 자신이므로, M1 의 첫 작업으로 라우팅한다.
**그때까지 이 축의 방어는 사람 독해뿐이다** — 그 사실을 여기 남기는 것이 조건이다.

기타(optional, 차단 아님): N2 `§2.1` RED-now 4요소 중 축어 stdout·종료코드가 AC-004 외 11개 셀에서 누락
(감사가 셀을 전부 재실행해 재현했고, DoD 2번이 run-phase 에 그 기록을 이미 요구한다) · N3 수사 토큰에 낱말
경계가 없어 `세부`·`네트워크`·`열기`·`phone` 이 적중(오늘 오탐 실측 0, 미래 노출) · N4 재현 명령 5 의 2단계
정규식이 1단계의 부분집합이 아니라 `twenty tabs` 가 소실 · N5 D12 트레일러가 전이를 운반하지 않는 커밋에
붙어 린트가 여전히 `OwnershipTransitionUnmeasured` INFO 를 낸다(`c205eeeff` 는 이미 착지, 부채로 기록).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
