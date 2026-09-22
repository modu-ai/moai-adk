# t774 판정서 — 무거운 테스트 실행 순번: slot 표면 재사용 확정과 §8 규율 확장

- 카드: t774 (Tier S, Class C — plan → Kickoff(리드·운영자 승인) → run → sync)
- 브랜치: `WT-heavy-test-queue` (워크트리 `.claude/worktrees/t774`) · 기점 `d416f8162`
- 산출: SPEC-HEAVY-TEST-SLOT-001 v0.2.0 (completed) + `.claude/rules/local/gitflow-lane-protocol.md` §8 확장(3문장) — 제품 코드 0
- push 금지 · 병합 창 요청은 완료 보고 동봉

## Claim (주장)

1. **카드가 요구한 표면은 이미 존재하며 가동 중이다** — `moai slot`(internal/cli/slot.go): 교차 세션 임대(acquire/status/release), 임대 레코드는 1차 체크아웃 `.moai/state/slot-leases` 공유. 최초 감사 기록 2026-09-12T12:28:42Z(`heavy-test`), plan 시점 19 이벤트(`heavy-test`×13·`t675-cli-tests`×4·`gotest-gateway`×2), 이후에도 증가. 카드 발행(09-10) 시점엔 없었으므로 침입 당시의 "표면 부재" 전제는 참이었다.
2. **카드 3판정의 답**: (a) 별도 표면은 필요 없음 — 존재; 결핍은 레인이 언제 임대를 잡아야 하는지 묶는 규율 문면의 세 항목이었다. (b) `moai integration` 재사용 불가 확인 — 병합 창과 실행 순번은 기록·수명이 다르며 이번 배치에서 실제로 동시 필요가 있었다. (c) 기록 필드 전부 충족 — status 출력이 보유자 세션+pid·이름·명령·시작·bound/ends(예상 종료)를 지닌다(측정 출력은 investigation.md).
3. **[HARD] 통제군 이행**: "표면이 없을 때 침범이 난다"의 증거는 09-10 실측(세 레인 독립 관측, load 8~21, 판정 뒤집힘)이며, "표면이 있을 때 거부된다"의 증거는 이번 생시 재현(제2 세명 획득 → **exit 3**, 무음, 대체 없음, 이후 정상 반납)이다. 동시 600초 스위트 재실행에 의한 재시연은 부하 규율 위반이라 하지 않았다 — 문서에 그 금지를 명문화했다.
4. **run 산출 = 기존 §8 단락 확장 3문장뿐이다.** plan-audit 1회(FAIL 0.70)가 v0.1.0의 조사 누수(§8 기존 절차·미러 문서 미확인)를 잡았고, v0.2.0은 신규 문서를 폐기하고 §8에 진짜 덧붙일 항목(거부 exit 3 · WHEN 예시 패키지 · 통제군 기록)만 확장하는 형태로 재설계됐다. 리드·운영자 Kickoff 승인(2026-09-14)으로 run 진입.
5. **배포 판정**: `moai slot` 동사는 제품 CLI 표면. 본 카드의 산출은 팩토리/칸반 운영 규율로 로컬 전용(`.claude/rules/local/`)이며, 템플릿 미러 문서(`resource-slot-lease.md`, 배포본)에 내부 카드·사고 기록을 진입시키지 않았다 — §25 중립성 준수.

## Evidence (증거 — 이번 실행, 이 트리)

merge-base 앵커: `git merge-base d416f8162 HEAD` = `d416f816201e1337ef421aba4fec9f5d8cd4cd7b` (변경 없음 — 흡수 없었음).

- **AC-HTS-001** `grep -c 'exit 3' .claude/rules/local/gitflow-lane-protocol.md` → **1**; `grep -c 'internal/cli' …` → **1** (≥1 요구 충족)
- **AC-HTS-002** `grep -c '2026-09-10' …` → **2**; `grep -cE '재시연|재현' …` → **2** (≥1 충족)
- **AC-HTS-003** `git diff --name-only <merge-base> HEAD` → 6파일 전부 `.claude/rules/local/gitflow-lane-protocol.md`·`.moai/reports/t774/*`·`.moai/specs/SPEC-HEAVY-TEST-SLOT-001/*` — 금지 접두(`internal/template/templates/`·`docs-site/`·`resource-slot-lease`) **0건**
- **AC-HTS-004** `git diff <merge-base> HEAD -- internal/cli/slot.go internal/cli/slot_test.go` → **0행**
- 강제성 재현(2026-09-14): acquire 성공 → 제2 세명 획득 `exit=3` 무음 → status에 원 보유자 유지 → release 정상. 전체 출력은 `.moai/reports/t774/investigation.md` §2.

## Baseline-attribution (baseline 귀속)

모든 측정·편집은 이번 턴에 워크트리 `.claude/worktrees/t774`(기점 `d416f8162`)에서 실행했고, §8 존재·미러 여부는 plan-audit 지적을 제 손 grep으로 재확인했다(감사 보고 인용에 그치지 않음).

## Gaps (미검증)

- 규율 문서가 레인 행동을 실제로 바꾸는지는 이 카드가 측정하지 않았다 — 다음 배치에서 침범 0회 관측이 그 판정이고, 실패 시의 상승 단계(선택형 slot 가드 활성)는 SPEC 비목표로 기록돼 있다.
- `resource-slot-lease.md`(배포본)에 exit 코드 표를 넣을지는 제품 표면 결정으로 본 카드 범위 밖 — 사용자 관점 문서는 현재 exit 코드를 기술하지 않는다.
- CHANGELOG 미작성 — 제품/배포본 변경이 0이라 [Unreleased] 항목 대상이 아니라고 판정했다(판정 근거: diff 파일 집합 전부 로컬 전용·SPEC 산출물).

## Residual-risk (잔여 위험)

- §8의 WHEN 목록은 예시 나열이지 기계 문턱이 아니다 — 새로운 무거운 패키지가 생기면 목록 갱신은 수동이다(기계 문턱은 제품 변경으로 본 SPEC 비목표).
- slot 임대는 자율 규율이다 — 가드가 꺼져 있는 한 획득 없이 실행해도 기계적 거절은 없다(그때의 상승로는 비목표 절에 기록).
- `resource-slot-lease.md` 미러 본과 §8의 역할 분담(사용자 표면 vs 레인 규율)은 유지보수 시 갈라질 수 있다 — 갱신은 어느 쪽인지 먼저 판별하고 진행할 것.
