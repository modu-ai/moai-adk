# plan.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §A. 맥락

- 카드 t1236 (AUTONOMY-A3), 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`. v0.1.0 작성 기준 `ca1d5dc43`, v0.2.0 개정 기준 `ca1e39f6f`.
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L).
- 개발 방식: 문서 층(추가형 블록) + Go 코드(`moai contract decide`·`revoke`, 영수증 저장소) + 가드 테스트. Go 는 TDD — RED 를 먼저 관측한다(`design.md §10`). 문서 블록은 acceptance 스크립트가 RED → GREEN 을 판정한다.

## §B. 미해결 질문 (Implementation Kickoff 전에 해소)

v0.2.0 에서 리드 결정으로 풀린 것: needs-decision 의 소유와 형태(A2 소유, 큐 상태 아님, 에스컬레이션 기록 — REQ-GR-052·091), 비대화형 서명 경로(A1 에 요청 — `research.md §9.1`), A2 선행 여부(선행 — §C). 영수증 저장소 위치는 선례를 따른 선택이라 NC 로 올리지 않았다(`design.md §8.3`).

**[NEEDS CLARIFICATION: NC-1 모드 판독 방식]** 오케스트레이터가 contract 모드와 Kickoff 결정자 설정을 무엇으로 아는가. (a) `.moai/config/sections/workflow.yaml` 을 직접 읽는다, (b) `moai contract show --json` 의 `mode` 필드, (c) SessionStart 주입(A1 초안에 없음). 권장 (b) — 무효값 → guided 규칙을 A1 코드 한 곳이 적용한다. 결정자 설정은 `show --json` 에 아직 없으므로 A1 에 필드 추가를 요청해야 할 수 있다.

**[NEEDS CLARIFICATION: NC-2 plan phase 소크라테스 인터뷰]** 계약은 plan-audit 뒤에 서명되므로 plan phase 초반 인터뷰 시점에는 계약이 없다. (a) 서명 이후의 인터뷰만 흡수(권장), (b) plan phase 인터뷰까지 생략. (b) 면 `clarity-interview.md` 가 편집 대상에 들어간다.

**[NEEDS CLARIFICATION: NC-3 G4 원문 위치]** 가정 확인 대기의 원문은 `moai-constitution.md` Agent Core Behaviors 1 (`CONST-V3R2-030`). 헌법 파일을 고치지 않고 SSOT 등가 조항으로만 처리하는 안(권장)이 충분한가.

**[NEEDS CLARIFICATION: NC-4 서명 후 goal 무장]** contract 모드에서 진행 방식 축은 묻지 않는다. 서명 검증 직후 `ac_converge` goal 을 자동으로 무장하는가(goal 상한 유지), 아니면 계약 필드로 받는가(A1 스키마 개정 필요).

**[NEEDS CLARIFICATION: NC-5 초안 contract.yaml 작성 주체]** 지시를 `spec-assembly.md` 위임문에 싣는 안(권장) vs manager-spec 에이전트 본문에도 추가(에이전트 편집 + `make agents-emit` 동반).

**[NEEDS CLARIFICATION: NC-6 always-loaded 증가 상한]** 세 always-loaded 파일 합계 증가 1,500자를 허용하는가, 0 을 요구하는가. v0.2.0 의 자율 Kickoff 서술은 SSOT(경로 한정)에 두고 always-loaded 포인터는 늘리지 않으므로 상한 값은 v0.1.0 과 같다.

**[NEEDS CLARIFICATION: NC-7 Jev 를 운영자 게이트 입력으로 쓰는 것]** 완료된 `SPEC-JEV-CORE-001` REQ-JEVC-012 가 「운영자 게이트에 Jev 를 결정으로도 **입력으로도** 쓰지 않는다」고 정하고, `moai-mcp-tools.md` 의 `jev_ask` 행도 「never … gate input」이라 한다(`research.md §9.2`). 운영자가 승인한 설계는 Jev 를 Kickoff 의 두 번째 신호로 쓴다. 선택지: (a) REQ-JEVC-012 를 개정하는 별도 카드를 열고, 개정 착지를 자율 Kickoff 활성 조건으로 둔다(권장 — 이 SPEC 은 이미 그렇게 적었다, REQ-GR-016 (iv)); (b) 개정하지 않고 자율 Kickoff 를 LLM 단독으로 설계를 바꾼다(운영자 결정의 합의 규칙과 어긋남); (c) Jev 를 합의가 아니라 「거절만 할 수 있는」 신호로 제한한다(여전히 입력이므로 개정 필요). 이 결정 없이는 M7 의 Jev 호출 부분을 구현하지 않는다.

**[NEEDS CLARIFICATION: NC-8 주 LLM 결정자 역할]** 설계는 기존 `mission-governor` 역할 재사용을 제안하지만 그 정의가 「NOT for: … approval」을 명시한다(`research.md §9.3`). (a) 리드 세션 자체 또는 새 컨텍스트의 `Agent(general-purpose)` 판단 역할을 쓴다(에이전트 파일 수정 없음 — 권장), (b) `mission-governor` 의 설명을 개정한다(에이전트 편집 + `make agents-emit`, 이 SPEC 범위 밖이라 별도 카드).

**[NEEDS CLARIFICATION: NC-9 revoke 의 정지 시점]** revoke 는 프로세스를 죽이지 않으므로 진행 중 run 은 다음 단계 경계에서 멈추고, 경계 사이 작업은 끝까지 간다. 이 지연을 허용하는가(권장 — 파괴적 중단보다 안전), 아니면 A2 감지기가 도구 호출 시점에 revoke 를 보고 즉시 거부하도록 A2 에 요청하는가.

## §C. 착수 전제와 기준 ref

run phase 진입 조건 (모두 필요):

1. **t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) 가 develop 에 병합됨.** 확인: `moai contract verify --help` exit 0. 비대화형 영수증 서명 경로가 A1 에 들어갔는지도 확인한다 — 없으면 M8(자율 Kickoff 블록)을 보류하고 M1~M7 만 진행한다.
2. **t1235 (A2, `SPEC-AUTONOMY-ESCALATION-001`) 가 develop 에 병합됨** (리드 결정). 에스컬레이션 기록 형식, push 직렬화 강제, 에이전트발 대화형 `sign` 차단이 여기서 온다.
3. **t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) 가 develop 에 병합되고 이 브랜치에 흡수됨.**

`depends_on` 에 세 SPEC 을 적었으므로 run 게이트의 Depends_on 사전 점검이 `status: completed` 로 판정한다.

`BASE` 정의: 전제를 만족한 뒤 이 브랜치가 마지막으로 흡수한 develop 커밋. M0 에서 `git merge-base develop HEAD` 로 구해 `progress.md §E.2` 에 SHA 로 적는다. 바이트 보존·분기 불변 AC 는 이 SHA 를 쓴다.

## §D. 제약

- 문서 편집은 `design.md §2` 1~15행에 한정. 참조 지점 R 과 REQ-GR-080 제외 목록은 바이트 불변.
- 로컬·템플릿은 같은 커밋에서 함께 편집한다.
- 블록과 SSOT 에 내부 토큰 금지(REQ-GR-072).
- Go 테스트는 `t.TempDir()` 와 격리된 `MOAI_HOME` 만 쓴다 — 실제 `~/.moai` 에 쓰지 않는다. Jev 는 스텁 이음매로만 테스트하고 네트워크를 쓰지 않는다.
- 검증은 범위 한정: `go test ./internal/contract/... ./internal/template/...`, `go test ./internal/cli/ -run 'TestContract(Decide|Revoke)'`, `moai constitution validate`, acceptance 스크립트. `go test ./...` 로컬 실행 금지.
- `.claude/agents/**` 를 건드리지 않으므로 `make agents-emit` 불필요.
- 커밋 메시지마다 카드 id `t1236`.

## §E. 자기 검증 (run phase 완료 보고에 포함)

acceptance.md 의 AC 전부를 PASS/FAIL 표로, 명령·원문 출력·exit 코드·트리 SHA 와 함께 보고한다. Go 마일스톤은 RED 원문(`--- FAIL`)을 GREEN 앞에 남긴다.

## §F. 마일스톤 (결정이 바뀔 가능성이 큰 것부터)

### M0 — 흡수와 재앵커 (Priority High, 선행)

- §C 세 전제 확인 → develop 흡수 → `BASE` 기록.
- `research.md §1` 계수를 `BASE` 에서 재실행. E/R/H 분류가 바뀌면 blocker 로 manager-spec 재위임.
- A1·A2 감사 통과본과 **[A1 감사 통과본으로 재확인]**·**[A2 스키마 재확인]** 표시 항목을 대조. 차이가 있으면 blocker.
- 절 제목 앵커가 t1175 이후에도 있는지 확인.

### M1 — SSOT 규칙 `contract-autonomy.md` (Priority High)

게이트 처분표·등가 조항·생명주기·에스컬레이션 라우팅·자율 Kickoff 절차·revoke 서술. NC-1·2·4·7·8 의 답이 여기에 들어간다. 로컬·템플릿 동시 생성. 자율 Kickoff 절(§3 9번)은 M8 까지 「비활성 — 활성 조건 미충족」으로 쓴다.

### M2 — G1 발화 지점 (Priority High)

`orchestration-mode-selection.md`, `CLAUDE.md` §2, `moai.md`, `run.md`, `spec-assembly.md`(Step 2.3.3b·Phase 10 초안), `plan.md`, `SKILL.md`, `goal.md`, `goal-directive.md`. 이 단계의 블록은 결정자 `human` 경로만 서술한다.

### M3 — G2/G3/G4 (Priority Medium)

`CLAUDE.md` §7, `askuser-protocol.md`, (NC-2 가 (b) 면) `clarity-interview.md`.

### M4 — G7/G8 (Priority Medium)

`spec-assembly.md` Step 2.3.5·Phase 15, `moai.md` Step 11.3.

### M5 — G11 과 생명주기 (Priority Medium)

`run.md` 생명주기 블록, `sync.md`, `sync/doc-execution.md`, `sync/delivery.md`.

### M6 — `moai contract revoke` (Priority High, TDD, 자율 Kickoff 보다 먼저)

`internal/contract/receipt/`(저장소·체인의 최소 부분) + `internal/contract/revoke/` + `internal/cli/contract_revoke.go`. RED: `design.md §10` 의 revoke 행 세 가지와 체인 변조 행을 먼저 실패로 관측. **이 마일스톤이 M8 보다 먼저 커밋돼야 한다(AC-GR-022).**

### M7 — 영수증 저장소 + `moai contract decide` (Priority High, TDD)

`internal/contract/receipt/` 완성 + `internal/contract/kickoff/` + `internal/cli/contract_decide.go`. RED: 전제조건 6종, Jev 불가 3종, 신뢰도, 합의 매트릭스, 작성자 배제. Jev 호출부는 NC-7 이 (a) 로 정해지고 REQ-JEVC-012 개정이 착지한 뒤에만 실제 클라이언트를 연결한다 — 그전에는 Jev 결정자가 항상 「측정되지 않음」을 돌려주는 상태로 둔다(결과는 늘 `human`).

### M8 — 자율 Kickoff 블록 활성 (Priority Medium, M6·M7 뒤)

SSOT §3 9번과 발화 지점 블록에 결정자 `llm+jev` 경로를 추가. 활성 조건(`design.md §7.1`)을 오케스트레이터가 실행 시점에 확인하는 절차로 적는다. A1 영수증 서명 경로 또는 A2 두 부분이 develop 에 없으면 이 마일스톤은 보류한다.

### M9 — 가드 테스트 (Priority Medium, TDD)

`internal/template/contract_mode_blocks_test.go`. RED: 불량 픽스처(짝 불일치, evolvable 구간 안, SPEC ID 포함)와 「검사 대상 0건」.

### M10 — 기계적 마무리 (Priority Low)

`make build`, always-loaded 예산 측정, acceptance 스크립트 전체, `moai constitution validate`, 범위 한정 테스트.

## §G. 반패턴

- guided 문장을 「조금 다듬기」 — 바이트 보존이 깨진다.
- evolvable 구간 안에 블록 삽입.
- 블록에 G 번호·SPEC ID·카드 id 사용.
- 참조 지점까지 블록을 뿌리기.
- A1·A2 초안의 이름을 감사 전 값 그대로 확정 — M0 재대조 생략.
- revoke 보다 먼저 자율 Kickoff 블록을 커밋 — AC-GR-022 FAIL.
- 영수증 저장소를 「변조 방지」로 서술 — 흔적만 남긴다.
- 에이전트가 쓴 `kickoff-receipt.json` 사본을 권위 있는 영수증으로 취급.

## §H. 교차 참조

- `spec.md` 요구사항, `acceptance.md` 판정, `design.md` 편집 집합·마커 규약·decide·revoke·저장소, `research.md` 인벤토리·A1 기준선(`8f77d9a33`, `98cb7879d`)·충돌(§9).
- `SPEC-AUTONOMY-CONTRACT-001` (A1), `SPEC-AUTONOMY-ESCALATION-001` (A2, 초안 `ee9f57151`), `SPEC-JEV-CORE-001` (REQ-JEVC-007·011·012).
- `SPEC-ALWAYS-LOADED-DIET-002` (t1175) — 브랜치 `WT-rules-diet`.
- `.claude/rules/moai/development/verification-completeness.md` — RED-now 셀 요건.
