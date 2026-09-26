# plan.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §A. 맥락

- 카드 t1236 (AUTONOMY-A3), 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`, plan 작성 기준 HEAD `ca1d5dc43`.
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L).
- 개발 방식: 문서 층 변경 + 가드 테스트 1개. 가드 테스트는 TDD(RED 먼저), 문서 블록은 acceptance 스크립트가 RED → GREEN 을 판정한다.

## §B. 미해결 질문 (Implementation Kickoff 전에 해소)

**[NEEDS CLARIFICATION: NC-1 모드 판독 방식]** 오케스트레이터가 contract 모드임을 아는 방법. 선택지: (a) plan→run 경계에서 `.moai/config/sections/workflow.yaml` 의 `workflow.autonomy.mode` 를 읽는다, (b) `moai contract show --json` 의 `mode` 필드를 읽는다, (c) SessionStart 가 모드를 주입한다(A1 초안에 없음 → 새 작업). 권장 (b) — 설정 판독 규칙(무효값 → guided)을 A1 코드가 한 곳에서 적용하므로 문서가 규칙을 다시 적지 않아도 된다. 단 SPEC 이 정해지기 전(plan phase 초반)에는 (a) 가 필요할 수 있다.

**[NEEDS CLARIFICATION: NC-2 plan phase 소크라테스 인터뷰]** 계약은 plan-audit 뒤에 서명되므로, plan phase 초반의 인터뷰 시점에는 계약이 없다. contract 모드가 (a) 서명 **이후**의 인터뷰만 흡수하는가(권장 — plan phase 인터뷰는 계약 내용을 만드는 입력이다), (b) plan phase 인터뷰까지 생략하는가. (b) 면 `design.md §2` 15행(`clarity-interview.md`)이 편집 대상에 들어간다.

**[NEEDS CLARIFICATION: NC-3 G4 원문 위치]** 가정 확인 대기의 원문은 `moai-constitution.md` Agent Core Behaviors 1 (`CONST-V3R2-030`) 이다. 헌법 파일을 고치지 않고 SSOT·CLAUDE.md 블록의 등가 조항으로만 처리하는 안(권장)이 충분한가, 아니면 헌법 파일에 포인터 한 줄을 두는 개정을 별도 카드로 여는가.

**[NEEDS CLARIFICATION: NC-4 서명 후 goal 무장]** contract 모드에서 진행 방식 축(autonomous / semi-autonomous)은 묻지 않는다. 서명 검증 직후 `ac_converge` goal 을 자동으로 무장하는가(goal 상한 G16 유지), 아니면 무장 여부를 계약 필드로 받는가(A1 스키마에 해당 필드 없음 → A1 개정 필요).

**[NEEDS CLARIFICATION: NC-5 초안 contract.yaml 작성 주체]** A1 은 「manager-spec 이 plan phase 에 초안을 낸다」를 A3 로 넘겼다. 이 SPEC 은 에이전트 파일을 고치지 않으므로 지시를 `spec-assembly.md` 의 위임문에 싣는다(권장). manager-spec 에이전트 본문에도 소유 산출물로 추가해야 한다면 `make agents-emit` 을 동반한 에이전트 편집이 범위에 들어온다.

**[NEEDS CLARIFICATION: NC-6 always-loaded 증가 상한]** `design.md §4` 는 세 always-loaded 파일 합계 증가를 1,500자로 잡았다. t1175 리드가 이 증가를 허용하는가, 아니면 0 증가(always-loaded 파일에는 블록을 두지 않고 skill/path-scoped 파일에만 둠)를 요구하는가. 0 이면 CLAUDE.md §2·§7 과 askuser-protocol 의 contract 서술이 빠지며, 카드 문언(「CLAUDE.md 등 8곳+」)과 충돌한다.

## §C. 착수 전제와 기준 ref

run phase 진입 조건 (둘 다 필요):

1. **t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) 가 develop 에 병합됨.** 확인: `git log --oneline develop -- .moai/specs/SPEC-AUTONOMY-CONTRACT-001/contract.yaml internal/` 등 A1 run 산출물 존재, `moai contract verify --help` exit 0.
2. **t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) 가 develop 에 병합되고 이 브랜치에 흡수됨.** 확인: `git merge-base --is-ancestor <t1175 병합 커밋> HEAD` exit 0.

`BASE` 정의: 두 전제를 만족한 뒤 이 브랜치가 마지막으로 흡수한 develop 커밋. M0 에서 `git merge-base develop HEAD` 로 구해 `progress.md §E.2` 에 SHA 로 적는다. 이후 모든 바이트 보존·분기 불변 AC 는 이 SHA 를 쓴다(움직이는 ref 금지 — `verification-completeness.md §4`).

## §D. 제약

- 편집은 `design.md §2` 의 집합에 한정. 참조 지점 R 과 REQ-GR-090 목록은 바이트 불변.
- 로컬·템플릿은 같은 커밋에서 함께 편집한다.
- 블록과 SSOT 에 내부 토큰 금지(REQ-GR-072). 이 SPEC 의 ID·카드 id 는 커밋 메시지와 SPEC 문서에만.
- 템플릿 편집 후 `make build` 로 임베드 갱신. `.claude/agents/**` 를 건드리지 않으므로 `make agents-emit` 은 불필요하나 `make build` 의 선행 `agents-emit-check` 가 통과해야 한다.
- 검증은 범위 한정: `go test ./internal/template/...`, `moai constitution validate`, acceptance 스크립트. `go test ./...` 로컬 실행 금지.
- 커밋 메시지마다 카드 id `t1236`.

## §E. 자기 검증 (run phase 완료 보고에 포함)

acceptance.md 의 AC 전부를 PASS/FAIL 표로, 명령·원문 출력·exit 코드·트리 SHA 와 함께 보고한다. 특히: AC-GR-001(바이트 보존), AC-GR-002(블록 동등), AC-GR-004(Frozen 불변), AC-GR-005(발화 지점 전수), AC-GR-008(가드 테스트 RED 원문).

## §F. 마일스톤 (결정이 바뀔 가능성이 큰 것부터)

### M0 — 흡수와 재앵커 (Priority High, 선행)

- t1234·t1175 병합 확인 → develop 흡수 → `BASE` 기록.
- `research.md §1` 계수 명령을 `BASE` 에서 재실행. 파일 목록이 바뀌었으면 E/R/H 재분류를 `progress.md §E.2` 에 기록(research.md 는 plan 산출물이므로 run 에서 고치지 않는다 — 차이가 E 집합을 바꾸면 blocker 로 manager-spec 재위임).
- A1 감사 통과본의 `design.md § Contract Schema` 와 이 SPEC 의 **[A1 감사 통과본으로 재확인]** 표시 항목을 대조. 차이가 있으면 blocker 로 manager-spec 재위임.
- 절 제목 앵커(`design.md §2` 의 삽입 위치)가 t1175 이후에도 존재하는지 확인.

### M1 — SSOT 규칙 `contract-autonomy.md` (Priority High)

게이트 처분표·등가 조항·생명주기·에스컬레이션 라우팅이 모이는 곳. 가장 바뀌기 쉬운 결정(NC-1·2·4 의 답)이 여기에 들어간다. 로컬·템플릿 동시 생성.

### M2 — G1 발화 지점 (Priority High)

`orchestration-mode-selection.md`, `CLAUDE.md` §2, `moai.md`, `run.md`(서명 블록), `spec-assembly.md`(Step 2.3.3b·Phase 10 초안), `plan.md`, `SKILL.md`, `goal.md`, `goal-directive.md`.

### M3 — G2/G3/G4 (Priority Medium)

`CLAUDE.md` §7, `askuser-protocol.md`, (NC-2 가 (b) 면) `clarity-interview.md`.

### M4 — G7/G8 (Priority Medium)

`spec-assembly.md` Step 2.3.5·Phase 15, `moai.md` Step 11.3.

### M5 — G11 과 생명주기 (Priority Medium)

`run.md` 생명주기 블록, `sync.md`, `sync/doc-execution.md`, `sync/delivery.md`.

### M6 — 가드 테스트 (Priority Medium, TDD)

`internal/template/contract_mode_blocks_test.go`. RED: 불량 픽스처 3종(짝 불일치, evolvable 구간 안, SPEC ID 포함)과 「검사 대상 0건」에서 실패함을 먼저 관측. GREEN: M1~M5 의 실제 블록에서 통과.

### M7 — 기계적 마무리 (Priority Low)

`make build`, always-loaded 예산 측정, acceptance 스크립트 전체 실행, `moai constitution validate`, `go test ./internal/template/...`.

## §G. 반패턴

- guided 문장을 「조금 다듬기」 — 바이트 보존이 깨지고 AC-GR-001 이 FAIL.
- evolvable 구간 안에 블록 삽입 — `moai update` 가 사용자 구간으로 보존해 갱신되지 않음.
- 블록에 G 번호·SPEC ID·카드 id 사용 — 템플릿 중립성 위반.
- 참조 지점까지 블록을 뿌리기 — 편집 파일 수만 늘고 등가 조항이 이미 덮는다.
- A1 초안의 필드명을 감사 전 값 그대로 확정 — M0 재대조를 건너뛰는 것.

## §H. 교차 참조

- `spec.md` 요구사항, `acceptance.md` 판정, `design.md` 편집 집합·마커 규약, `research.md` 인벤토리·A1 기준선(`8f77d9a33`).
- `SPEC-AUTONOMY-CONTRACT-001` (A1) design.md § Contract Schema.
- `SPEC-ALWAYS-LOADED-DIET-002` (t1175) — 브랜치 `WT-rules-diet`.
- `.claude/rules/moai/development/verification-completeness.md` — RED-now 셀 요건.
