# SPEC Review Report: SPEC-LEARN-CHANNEL-SCOPE-001

Iteration: **2 — post-ceiling verification verdict** (Tier S `plan_audit_tier_ceilings` S=1은 iter-1의 전수 감사에서 소진됐다. 본 iter-2는 천장 이후 오케스트레이터 재량으로 수행된 **결함-델타 재검증**이다 — iter-1 결함 목록(D1-D7)의 폐쇄 검증 + 신규 결함 스윕이 범위 전부이며, 재전수 감사가 아니다. 재검증 반복 예산을 리셋하지 않는다. 판정 권한은 본 감사 에이전트에 있다.)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.86** (harmonic mean — Tier S threshold 0.75 이상, iter-1 0.63 대비 단조 개선, STOP 트리거 없음)
대상: v0.1.1 (spec.md + plan.md, 미커밋, worktree `WT-learn-channel-gap` @ `d7ce6c6bd` — 본 라운드 `git rev-parse` 재확인, D5 핀 유효 유지)
잔여 결함(debt): **N1 1건 (major — kickoff 전 필수 수정)** + N2 1건 (minor, optional)

> 요약 한 줄: iter-1의 7건 결함이 전부 실제로 폐쇄됐다 — D1은 방향까지 올바르게(구성/능력 구분 + 철회 기록). 수정 라운드가 심은 신규 결함은 1건(N1: RED 프로브 마커 언어 불일치)이며, 이것이 본 판정의 유일한 부채다.

---

## Must-Pass Results (v0.1.1 재판정)

- **[PASS] MP-1 REQ 번호 일관성**: spec.md:88-94 연속 001-007, :170-176 AC 연속 001-007 — 갭·중복 0 (그 외 토큰은 교차참조: :24 HISTORY, :99 제약, :116-119 §F, :155-168 §I 헤더·RED 표). `grep -on` 본 라운드 실행.
- **[PASS] MP-2 GEARS (요구 레이어)**: 7 REQ 전부 패턴 적합 — REQ-001/002/003 ubiquitous, REQ-004 **Where**(capability gate 표기 명시화), REQ-005 unwanted, REQ-006 state-driven, REQ-007 Where.
- **[PASS] MP-3 frontmatter**: 12 정본 필드 유지 + `version: "0.1.1"` 승격(quoted semver) + snake_case 별칭 0. `updated: 2026-09-01` — spec.md mtime 2026-09-01 23:59와 일치(plan.md는 09-02 00:01로 자정 2분 경과 — 관측 기록, 결함 아님).
- **[N/A] MP-4**: 단일 도메인 문서 SPEC — 변화 없음.
- **[PASS] MP-5 D7**: 참조 SPEC 2건 상태 불변(iter-1 판독: completed) — 관련 frontmatter 미변경.
- **[PASS] MP-6 D8**: `grep -c 'syscall'` → 0 (본 라운드 재실행).
- **[PASS] MP-7**: `grep -rn 'NEEDS CLARIFICATION'` (spec/plan/progress) → 0매치 (본 라운드 재실행).

M5 방화벽 전부 통과.

---

## Regression Check (Iteration 2 — iter-1 결함 D1-D7 폐쇄 검증)

저자 자가보고를 채택하지 않고 iter-1 좌표 기준 직접 재판독했다. 판정:

- **D1 (critical — writer 단일성) — [RESOLVED]**, 방향 포함 올바름:
  - `spec.md:50` — stub 생산자 "정확히 2개"(`:77` tool_failure / `:111` test_fail), 같은 `appendLessonsInboxStub`(:129), 배선 `evidence_writer.go:583-591` `rec.IsTestFail`, 도입 `e70c77576` — 본 감사 iter-1 판독과 정확히 일치.
  - `spec.md:51` — **구조 vs 구성 구분 행 신설**: "배선(능력)과 구성(측정)은 다른 진술" + v0.1.0의 D1을 "본 SPEC이 고치려는 결함과 같은 형태를 스스로 저지른 것"으로 기록. 요구한 재프레임 그 자체다.
  - `spec.md:69` — §B.2에서 "미구현 서술" 가설 **명시 철회** + constitution-detail 문구를 "능력 기준으로 정확하다"로 정정 + 실제 갭 3건 재정의(능력·구성 미구분 / 경계 선언 부재 / 인간 루프 미명명).
  - `spec.md:125`(§G) — "제2 패밀리는 이미 트리에 배선돼 있다… 추가하거나 제거하지 않는다". `spec.md:142`(§H) — "유일 배출" 삭제, 2함수·배선·커밋 출처 명기. REQ-LCS-001(:88) 능력+구성형 재작성, REQ-LCS-005(:92) "no stub family beyond the two already wired", 제약 2(:99) 동조.
  - `spec.md:52` — 신규 관측 행: `failure_observer.go:80` 헤더 주석이 :109-111의 인박스 append를 빠뜨리는 문서-코드 드리프트 — 본 감사 iter-1 판독으로 실재 확인(:80-84는 usage-log.jsonl만 언급). kickoff 조건부 제안으로 올바르게 라우팅(본문 결정 아님).
- **D2 (major — AC-001 취약 불변식) — [RESOLVED]**: AC-LCS-001(:170) "zero rows **outside the `{tool_failure, test_fail}` families**" — 처방대로. §F(:118)·plan §C(:105-106, `cut -d: -f1-2 | sed 's/:$//' | sort -u` + `grep -c 'test_fail:'`) 동조. 재실행 가능성 본 라운드 검증(아래).
- **D3 (major — AC-004 범위 과잉) — [RESOLVED]**: REQ-LCS-004(:91) 트리거 "Where a prose surface **describes** the lessons-inbox's capture scope … the enumerated list in plan.md §A.5"; AC-LCS-004(:173) "the claim-surface list enumerated in plan.md §A.5 … (surfaces outside the §A.5 list are out of this AC's scope)". plan §A.5(:64-86) 신설 — claim 표면 4건 명시 열거 + 스윕 절차 + 배제표(navigator 배제 선언·미러 캐스케이드 사유, apply_test.sh 픽스처, frozen-allowlist, CHANGELOG, 코드/테스트, 본문 0히트) — **감사 iter-1 표면 분석을 정확히 계승**.
- **D4 (minor — 스테일 카운트) — [RESOLVED]**: §B.2(:70) 스테일 판정 명기, 제약 7(:104) "SPILL: curator SKILL.md:34의 스테일 카운트는 M2에서 anchor 포인터로 교체", AC-LCS-003(:172) "'624 stubs' is replaced by the anchor pointer (no live count in prose)", plan M2(:134)·EXTEND(:61)·pre-flight(:111 `sed -n '34p'`) 5개 표면 일관.
- **D5 (minor — SHA 핀) — [RESOLVED]**: `spec.md:40` §B 전체 문서-수준 핀(HEAD `d7ce6c6bd`, 2026-09-01) + 무엇을 핀가 정확히 스코핑("인박스·메모리 저장소는 primary live runtime state — 워크트리에 없고 복사 안 함") — verification-completeness §4의 핀 판별식을 올바르게 적용. §I:155 섹션-수준 핀이 전 RED 셀을 묶는다. exit 코드 셀에 기록(`0`/exit 1, 무출력/exit 1).
- **D6 (minor — 카드 id 배제) — [RESOLVED]**: AC-LCS-002(:171)·REQ-LCS-007(:94)·E3(plan:126) 3개 목록 모두 "card ids" 추가.
- **D7 (minor — two-cell 형식) — [RESOLVED]**: RED-now 표(§I:159-166) — 명령/원문 출력/exit/flipping 4요소 + test_fail baseline 행. 전 AC flipping 마일스톤 태그(AC-001→M1, 002/003/007→M2, 004/005/006→M3) — plan §F의 flip 주석과 일치. AC-005/006은 regression-guard로 명시 분류 + 사유 기술(:168 "사전 나무에서 적색이 성립하는 내용 주장이 없다") — verification-completeness §2.1의 undecidable/regression-guard 처분을 정확히 적용.

**7/7 RESOLVED. 미해결 잔여 0.**

---

## 신규 결함 (수정 라운드가 심은 것)

**N1.** LCS-RED-MARKER-LANGUAGE — `spec.md:162,163` (RED 표 AC-LCS-002/003 행) 및 `spec.md:171,172` ("the claim marker") — AC-002/003의 RED 프로브가 한국어 마커 `'인간 매개 루프'`를 **영어 문서 표면**에 grep한다. 실측: `.claude/rules/moai/core/moai-constitution-detail.md`는 **한글 0행**(순수 영어 파일, 미러도 0), `.claude/skills/hns-lsel-curator/SKILL.md`는 한글 2행뿐(:364-365, 인용된 보고 발췌) — §27 규약("all skill bodies in English")과 룰 트리 영어 관례상 올바른 구현은 영어로 쓰인다. 따라서 이 프로브는 **올바른 구현으로도 뒤집힐 수 없다**(verification-completeness §2가 RED-now 단독으로는 못 잡는다고 경고하는 impossible 방향 — red at arrival, red after correct work). AC-007의 한국어 마커는 CLAUDE.local.md가 한국어 표면이라 정상이고, AC-004의 `'learning-channel-scope'`(anchor 파일명)은 언어 중립적이라 정상. AC 본문의 Then 절은 내용 기반(언어 불가지론)이라 기준 자체는 측정 가능 — 결함은 프로브 키와 미정의 "claim marker"에 국한된다. — Severity: **major** — Class: **blocking(debt — kickoff 전 필수 수정)** — Required fix: AC-LCS-002/003 본문에 claim marker를 영어 토큰으로 정의(예: `human-mediated loop`)하고 RED 표 두 행의 grep 패턴을 재키잉. 4행 수정. 한국어 마커는 AC-007(CLAUDE.local.md)에만 유지.

**N2.** LCS-AC003-COMPOSITION-POINTER-TENSION — `spec.md:172` — AC-003 green 형태가 SKILL.md에 "the two wired families **with the measured composition**" 보유를 요구하는 반면, 제약 7(:104)과 REQ-004(:91)의 선택절("canonical bounded claim **or** anchor pointer")은 수치의 산문 유입을 pointer 쪽으로 기운다. 구현자가 양쪽을 다 만족하는 "요지 문장 + 포인터"로 해소할 가능성이 높아 실질 혼란은 작다. — Severity: **minor** — Class: **optional** — Required fix: AC-003을 "bounded claim in brief + anchor pointer (numbers live only in the anchor doc)" 형태로 정렬.

**관측 (결함 아님)**: plan.md mtime 2026-09-02 00:01 vs 기재일 2026-09-01 — 수정 라운드가 자정을 2분 넘김. 날짜 귀속 관용 범위 내.

---

## RED 셀 재현성 (본 라운드 샘플 재실행 — 축 3)

저자가 RED 셀에 기록한 관측값 6건을 현재 트리에서 전부 재실행했다:

| 프로브 (spec.md §I RED 표) | 셀 기록 | 본 라운드 재관측 (2026-09-02, HEAD `d7ce6c6bd`) | 판정 |
|---|---|---|---|
| `test -f .moai/docs/learning-channel-scope.md` | 무출력 / exit 1 | `ls` 대체 확인 — "No such file or directory", exit 1 | 재현 ✓ |
| `grep -c '인간 매개 루프' …moai-constitution-detail.md` | `0` / exit 1 | `0` | 재현 ✓ |
| `grep -c '인간 매개 루프' …hns-lsel-curator/SKILL.md` | `0` / exit 1 | `0` | 재현 ✓ |
| `grep -c 'learning-channel-scope' …SKILL.md` | `0` / exit 1 | `0` | 재현 ✓ |
| `grep -c '인간 매개 루프' CLAUDE.local.md` | `0` / exit 1 | `0` | 재현 ✓ |
| `grep -c 'test_fail:' <primary>/.moai/lessons-inbox.jsonl` | `0` / exit 1 | `0` | 재현 ✓ |

구성 불변식 재측정: live 인박스 `jq -r '.event_key' | cut -d: -f1 | sort | uniq -c` → **5,952행 전부 `tool_failure`**(2026-09-02). 측정 사슬 5,916 → 5,919 → 5,930/5,932(iter-1) → 5,942(v0.1.1) → 5,952(iter-2) — append-only 성장, `{tool_failure, test_fail}` 집합 밖 0행 불변. **패밀리-집합 판정식은 살아있는 채널에서 유지 가능함이 5중 baseline으로 확인** — 행수 변동은 예상된 생장이며 결함이 아니다. 저자의 사슬 인용(감사 iter-1 = 5,932)은 본 감사의 실제 마지막 관측값과 일치 — 귀속 정확.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | REQ 레이어 정밀. AC-002/003의 "the claim marker"가 RED 셀을 통해서만(잘못된 언어로) 정의됨 — N1 |
| Completeness | 1.00 | 1.00 | 능력/구성 구분 행 + 철회 + 출처 + 열거 스윕/배제 사유 + 스테일 수치 소관 + 문서 수준 핀 — §B가 이제 정직하고 완전 |
| Testability | 0.75 | 0.75 | 5/7 기기 건전(패밀리-집합·스코프·regression-guard 처분). 2/7(AC-002/003, release-blocking) 프로브 오키잉 — 기준 실질은 측정 가능, 프로브 재키잉 필요 |
| Traceability | 1.00 | 1.00 | 7↔7, 마일스톤 태그 spec↔plan 일치, HISTORY + 감사 상호참조, §A.5 열거, 사슬 인용 귀속 정확 |

**하모닉 평균**: 4 / (1/0.75 + 1/1.00 + 1/0.75 + 1/1.00) = **0.86** (iter-1 0.63 → 단조 개선)

---

## Recommendation

**PASS-WITH-DEBT.** 부채 N1 1건 — **Implementation Kickoff Approval 전에 수정을 마쳐야 한다**:

1. **(N1, kickoff-blocker)** manager-spec 경유 4행 수정: AC-LCS-002/003에 영어 claim marker 정의(예: `human-mediated loop`) + RED 표 두 행 grep 패턴 재키잉. 수정은 plan-artifact 해시를 무효화하므로 run-gate가 재검증한다 — 조용한 스킵 경로는 없다.
2. (N2, optional) AC-003 문구 정렬 — N1 수정 시 같이 1행.
3. N1/N2는 본 감사 에이전트가 수정하지 않는다(감사 경계 — plan-auditor는 감사만, manager-spec이 수정한다). 수정 후 판정은 기계적 확인으로 족하다(AC 본문에 영어 마커 존재 + RED 표 패턴 일치 grep).

카드의 기준 — "무엇이 실제로 학습되는지를 정직하게 정하는 것" — 은 v0.1.1이 이제 충족한다: 능력(배선 2패밀리)과 구성(측정 100% `tool_failure`)이 구분돼 선언되고, v0.1.0의 반대 방향 오류가 스스로 기록됐으며, 도구·테스트 실패 어느 쪽도 아닌 결함 계열이 인간 루프로 흐른다는 경계가 측정과 함께 고정됐다.
