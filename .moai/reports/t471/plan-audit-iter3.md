# SPEC Review Report: SPEC-LEAD-DEPUTY-001 — iter-3 delta re-audit

- **SPEC-ID**: SPEC-LEAD-DEPUTY-001 (card t471)
- **Tier**: M (threshold 0.80)
- **측정 트리**: `.claude/worktrees/t471` @ `e77ac4bb6120a39e5937ebb334eb36931d4c1d39` (branch `WT-lead-bottleneck`, develop 흡수 병합 커밋 — 리드 지시 HEAD와 일치 확인)
- **날짜**: 2026-09-04
- **감사자**: plan-auditor (iter-1 감사자와 동일 절차·판정 형식; M1 Context Isolation 유지)
- **반복**: 3 (리드 지시 delta 재감사 — iter-1 PASS-WITH-DEBT 0.925의 부채 D1·D2가 `b7370371b`로 수리된 것의 확인 패스. 하드 캡 3 이내. delta 범위: `b7370371b`의 acceptance.md 2줄 수리만 — 7건 optional 발견과 SPEC 전체는 재심의하지 않음)

## Verdict

**PASS** — 점수 **0.938** (threshold 0.80 이상) — **skip-eligible**

- skip 3조건: (1) verdict `PASS` ✓ (2) 점수 0.938 ≥ 0.80 ✓ (3) artifact-hash — 본 판정 시점 기준이며, run-gate가 `ComputeHash`로 재계산해 대조한다(본 판정 이후 plan 아티팩트 추가 변경 시 무효).
- 판정 토큰은 리드 지시대로 PASS/FAIL 2토큰 중 하나로 발행했다.

## Delta 범위 선언

`git log 109a4615d..e77ac4bb6 -- .moai/specs/SPEC-LEAD-DEPUTY-001/` 실측: **변경 커밋은 `b7370371b` 하나**, `acceptance.md` 2줄(2+/2−)뿐 — spec.md·plan.md·progress.md·frontmatter 무접촉. 감사 대상은 이 2줄의 수리 품질과 그로 인한 신규 결함뿐이다.

## 결함별 처분 (리드 체크리스트 순)

### D1 (AC-LDP-002) — **RESOLVED**

| 리드 체크 항목 | 판정 | 근거 |
|---|---|---|
| 기계 검증이 처방된 grep 형태 | ✓ | 수리 셀: `grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md` — iter-1 처방과 문자 단위 일치 |
| RED-now 셀이 원문 출력 + 종료코드 + 트리 SHA 핀 | **부분** | 원문 출력(`:0`/`:0`) ✓, 트리 SHA(`109a4615d`) ✓ — 단, **종료코드 필드 없음**(신규 발견 NEW-1, 아래). 현재 트리 재측정으로 요소 보완 기록: **exit=1** |
| green path가 M3 착지 내용을 명명 | ✓ | "M3가 `cross-session-messaging.md` § An idle notice is a scheduling hint 상호참조를 양쪽에 새기면 GREEN으로 뒤집는다" — 앵커 실측 생존(`cross-session-messaging.md:91`, 핀 이후 무변경) |
| 현재 트리 재실행 | ✓ | 출력: `.claude/agents/moai/manager-lead.md:0`, `.claude/rules/moai/workflow/kanban-dispatch.md:0` — **exit=1**. RED 상태 유지. 핀 대상 파일 2개는 `109a4615d..e77ac4bb6` 사이 무변경(empty diff) → develop 흡수가 핀을 훼손하지 않음 |

**판정**: 본질적 결함(공허-불가 판별자 + 거짓 RED-now)은 해소됐다. 계수기가 이제 진짜로 실패할 수 있고(현재 0/0=RED), M3 착지 시 뒤집힌다. verification-completeness.md §2 two-cell 목적(올바른 이유로 관측된 red + 핀 + green 경로)은 충족.

### D2 (AC-LDP-007) — **RESOLVED**

| 리드 체크 항목 | 판정 | 근거 |
|---|---|---|
| ls 기반 재서술 | ✓ | "회차 디렉터리 파일 목록으로 확인: `ls .moai/reports/lead/`에서 회차당 신규 항목 ≤2 — 회차 파일 1 + 인덱스 1" |
| `git diff --stat` 폐기 + 사유 기록 | ✓ | "`git diff --stat`은 쓰지 않는다: 회차 파일은 추적되지 않음" — 사유가 셀 안에 원문으로 남는다 |
| RED-now 핀 | ✓ | 트리 `109a4615d` 실측: `git ls-files .moai/reports/lead/` = 0행, `ls` = "No such file or directory" — 현재 트리 재측정 동일(0행 / 부재, exit=1). 핀 유효 |

**판정**: iter-1의 공허 검증(git diff on 비추적 = 항상 빈 출력)은 제거됐고, 관측 가능한 목록 기반 계수로 대체됐다.

### 신규 발견 (수리가 심은 것만 — 2건, 모두 optional)

**NEW-1** — `acceptance.md` AC-LDP-002 RED-now 셀 — 심각도: minor — 클래스: optional
§2.1 4요소(명령/원문 출력/종료코드/트리 SHA) 중 **종료코드 필드가 없다**(3/4). AC-LDP-002는 §D.1 release-blocking 분류라 요소 규정의 글자에 걸린다. 다만 종료코드는 기록된 명령+출력에서 결정적으로 도출된다(`grep -c` 0매치 = exit 1)且 본 판정서가 현재 트리 관측치(exit=1)를 기록해 두었다 — red가 보이지 않았거나 잘못 귀속된 경우(규정이 막으려는 실패)는 없다. **처분: optional. 최소 수정: 셀에 `exit 1` 1토큰 추가.** (글자 엄수를 택할 경우 blocking으로 재분류할 여지를 본 감사자가 명시적으로 남긴다 — 판단 근거는 비례성 브레이크 M6.)

**NEW-2** — `acceptance.md` AC-LDP-007 — 심각도: minor — 클래스: optional
정적 `ls` 목록에서 "회차당 신규 항목 ≤2"를 세려면 **이전 회차 목록과의 비교 기준**이 필요하다(연속 회차 보고가 목록을 운반하면 판정 가능하나, 셀은 비교 기준을 한 절 덜 명시한다). 최소 수정: "직전 회차 목록 대비" 1절 추가. 선택.

수리 diff 자체의 부작용 스윕: SPEC 디렉터리 변경은 `b7370371b` 단일 커밋·acceptance.md 2줄로 한정(실측) — 다른 산출물·frontmatter·AC ID 오염 0건.

## Must-Pass (delta 영향 반영)

MP-1..MP-7 전 항목은 iter-1 판정을 **그대로 승계**한다 — delta가 요구사항·frontmatter·참조 관계를 건드리지 않았기 때문이다(실측: 변경은 acceptance.md 본문 2줄). delta로 새로 판정한 것만:

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 | 승계 PASS | 무접촉 |
| MP-2 GEARS (요구 계층) | 승계 PASS | 무접촉 |
| MP-3 frontmatter | 승계 PASS | acceptance.md frontmatter 무변경(diff 실측) |
| MP-4 | 승계 N/A | — |
| MP-5 D7 | 승계 PASS | 참조 SPEC 무변경 |
| MP-6 D8 | 승계 PASS auto | 무접촉 |
| MP-7 clarification | 승계 PASS | plan.md 무접촉 |

## 차원 점수 (iter-1 대비 delta만 이동)

| 차원 | iter-1 | iter-3 | 변동 근거 |
|---|---|---|---|
| Clarity | 0.95 | 0.95 | 무변동 (delta가 요구사항 문안 미접촉) |
| Completeness | 0.95 | 0.95 | 무변동 (D3-iter1 등 optional 공백은 그대로 — delta 범위 밖) |
| Testability | 0.85 | **0.90** | AC-LDP-002·007 계측기가 모두 판별 가능·관측 가능으로 전환. 잔여 감점: NEW-1(종료코드 요소)·NEW-2(비교 기준)·AC-LDP-010의 판단 개입 요소(승계) |
| Traceability | 0.95 | 0.95 | 무변동 |
| **종합** | 0.925 | **0.938** | 산술 평균 — 0.80 이상 |

## Evidence (본 트리 e77ac4bb6에서 실행한 명령과 원문 출력)

| # | 확인 | 명령 | 원문 출력 / 종료코드 |
|---|---|---|---|
| E1 | 트리·HEAD | `git rev-parse --show-toplevel` / `git rev-parse HEAD` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t471` / `e77ac4bb6120a39e5937ebb334eb36931d4c1d39` |
| E2 | 수리 델타 | `git show b7370371b --stat` + full diff | acceptance.md 1 file, 2+/2− — AC-LDP-002·007 셀만 교체 |
| E3 | D1 재측정 | `grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md` | `.claude/agents/moai/manager-lead.md:0` / `.claude/rules/moai/workflow/kanban-dispatch.md:0` — **exit=1** |
| E4 | D2 재측정 | `git ls-files .moai/reports/lead/ \| wc -l` ; `ls .moai/reports/lead/` | `0` ; `ls: .moai/reports/lead/: No such file or directory` — **exit=1** |
| E5 | 핀 조상성 | `git merge-base --is-ancestor 109a4615d e77ac4bb6` | rc=0 — ANCESTOR, 핀 해석 가능 |
| E6 | 핀 이후 대상 파일 드리프트 | `git diff --stat 109a4615d..e77ac4bb6 -- <kanban-dispatch.md> <manager-lead.md>` | 빈 출력 — 무변경, RED 핀 비스테일 |
| E7 | SPEC 디렉터리 delta | `git log --oneline 109a4615d..e77ac4bb6 -- .moai/specs/SPEC-LEAD-DEPUTY-001/` + `git diff --stat` | 커밋 `b7370371b` 단일, acceptance.md 2줄만 |
| E8 | green-path 앵커 | `grep -n '## An idle notice is a scheduling hint' .claude/rules/moai/workflow/cross-session-messaging.md` | `91:## An idle notice is a scheduling hint`; 핀 이후 무변경(empty diff) |

## Gaps (관측하지 않은 것)

- AC-LDP-002의 GREEN 전환(양쪽 상호참조 신설)은 M3 착지 후에야 관측 가능 — 본 감사는 RED 고정과 전환 조건의 정의만 판정했다.
- `cross-session-messaging.md`의 로컬 사본과 템플릿 사본이 서로 다르다(`diff -q` 실측 불일치) — 본 SPEC이 편집하지 않는 파일이고 수리 delta 밖의 선존 상태라 finding으로 채택하지 않았으나, Template-First(REQ-LDP-010)가 `moai update` 흡수 경로를 전제하는 이상 이 divergence의 귀속은 delta 밖 미확인 사항으로 남는다.
- iter-1의 7건 optional 발견(D3~D9)은 리드 지시대로 재심의하지 않았다 — 미수리 상태로 승계된다.
- `moai spec lint` 재실행 안 함(iter-1과 동일 — MP-3은 수동 필드 대조로 별도 검증).

## Residual-risk

- NEW-1·NEW-2는 optional로 남겼다 — run-phase 진행자가 §2.1 글자 엄수를 택하면 AC-LDP-002 RED-now 셀은 요소 미완으로 읽힐 수 있다. 1토큰 수정이므로 run-phase 진입 전 처리가 비용 대비 안전하다.
- skip-eligible은 본 판정 이후 plan 아티팩트 무변경을 전제로 한다 — run-gate가 hash로 강제한다.
- AC-LDP-001의 실측(50% 감소)은 여전히 첫 채택 배치의 운영 데이터에 달려 있다(iter-1과 동일 — protocol-readiness만 판정됨).
