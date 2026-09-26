# plan.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.2)

## §A. 맥락

- 카드 t1236 (AUTONOMY-A3), 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`. 이력: v0.1.0 `ca1e39f6f`, v0.2.0 `62a65f709`(plan-audit iter-1 FAIL 0.71, `.moai/reports/t1236/plan-audit-1.md`), v0.3.0 `781ddc355` 은 그 수리본, v0.3.1 은 운영자 재결정(`llm+jev` 교차 확인만, 결정 규칙 A3 소유)·A2b 소유·리드 결정 R10 반영본.
- v0.3.2: v0.3.1 `1b071a573` 의 plan-audit iter-2 FAIL 0.77(`.moai/reports/t1236/plan-audit-2.md`) 수리본, 리드 결정 2026-09-26(D32·D33) 반영, A1 0.5.2 `25283ebf8` 기준.
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L).
- 개발 방식: 문서 층(추가형 블록 + Jev 원칙 개정) + Go 코드(kickoff-check·decide·revoke·사건 저장소·A1 서명기 사건 추가) + Go 가드 테스트. 모두 TDD — RED 를 먼저 관측한다(`design.md §12`).
- A1 기준: 0.5.2 `25283ebf8`(이력 `65e0a9167`; 스키마: 결정자 값 집합·파생 기본값·영수증 필드·구조 검증·`card` 필드). A2b(t1245): push 직렬화·에이전트발 `sign` 차단·팩토리 에이전트 세션의 `decide` 거절. A2 기준: 리드가 전달한 최종 형식(A2 개정본 대기, 현재 커밋 `d8926ff9a` 는 철회된 형식 — `research.md §10.1`).

## §B. 결정 기록 (미해결 질문 0건)

v0.2.0 의 NC-1~NC-9 는 모두 닫혔다. 본문 정본은 `spec.md §B` 「본문 결정」 D-1~D-9 이고, 여기에는 근거만 적는다.

| 항목 | 처분 | 근거 |
|---|---|---|
| NC-1 모드 판독 | 결정 D-1: `moai contract kickoff-check --json` 의 `mode`·`decider` | 설계 결정(감사 D1 분류). A1 0.5.2 `show --json` 에 결정자 필드가 없어 A3 명령이 채운다 |
| NC-2 plan phase 인터뷰 | **기본값 D-2**: 서명 이후 작업에만 적용. 운영자가 바꿀 수 있음 | 리드 승인(2026-09-26). 계약이 plan-audit 뒤에 서명되므로 논리적으로 강제된다(감사 NC 분류) |
| NC-3 G4 원문 | 결정 D-3: 헌법 파일 무수정 | 헌법 개정은 범위 밖 — 바인딩 범위가 답을 강제(감사 D1) |
| NC-4 서명 후 goal 무장 | **기본값 D-4**: 자동 무장 없음. 운영자가 바꿀 수 있음 | 리드 승인(2026-09-26). 보수적 기본값 — run 은 이 기본값으로 진행할 수 있다 |
| NC-5 초안 작성 지시 | 결정 D-5: `spec-assembly.md` 위임문 | `.claude/agents/**` 제외가 답을 강제(감사 D1) |
| NC-6 always-loaded 예산 | 결정 D-6: 블록 합계 1,500자 + `jev_ask` 행 300자 | `design.md §4` 에 수치로 묶음. t1175 조율은 리드 몫 |
| NC-7 Jev 게이트 입력 | 결정 D-7: 리드 결정 L1 을 운영자 재결정으로 좁힘 — contract 모드 Kickoff 의 `llm+jev` 교차 확인 두 번째 신호 한정 개정 | REQ-GR-013, `design.md §11` |
| NC-8 주 LLM 결정자 | 결정 D-8: 리드 결정 L3 — 리드 세션 또는 새 컨텍스트 판단 역할 | `mission-governor` 정의가 승인을 배제 |
| NC-9 revoke 정지 시점 | 결정 D-9: 다음 단계 경계 | REQ-GR-020·022 가 이미 규정(감사 NC 분류) |
| (v0.3.1) 결정자와 결정 규칙 | 결정 D-10 + REQ-GR-010 R1~R5 | 운영자 재결정 2026-09-26, 결정 규칙(결과 `approve`·`reject`·`human` 도출)은 A3 소유·스키마는 A1 0.5.1(리드 소유 조정) |

**운영자에게 남은 질문**: 없음. D-2·D-4 는 운영자가 바꿀 수 있는 기본값으로 기록했고, run 은 기본값으로 진행한다.

## §C. 착수 전제와 기준 ref

run phase 진입 조건 (모두 필요):

1. **t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) develop 병합.** 확인: `moai contract verify --help` exit 0. 병합본이 A1 0.5.2(`25283ebf8`) 이상의 스키마(`human`·`llm`·`llm+jev` 값 집합, 대체 영수증 필드, `card` 필드)를 담아야 한다 — 아니면 M7 의 영수증 쓰기를 보류하고 리드에게 보고.
2. **t1235 (A2, `SPEC-AUTONOMY-ESCALATION-001`) develop 병합** (리드 결정). 에스컬레이션 기록의 **최종 형식**(`escalation/<class>-<fingerprint>.md` + YAML 머리 + `revoke` 종류)이 A2 개정본에 들어 있어야 한다 — 현재 커밋 `d8926ff9a` 의 JSON 형식이면 M6 을 보류하고 리드에게 보고.
3. **t1245 (A2b) develop 병합** (리드 결정 2026-09-26) — push 직렬화 강제, 에이전트발 대화형 `sign` 차단, `MOAI_FACTORY_ROLE=agent` 세션의 `decide` 거절.
4. **t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) develop 병합 후 이 브랜치에 흡수.**

Push 단계(REQ-GR-021)와 자율 Kickoff 활성(M8)에만 걸리는 추가 조건 — run 진입 전제는 아니다:

- **t1237 (A4)** 착지 — 두 번째 리뷰 미수행 시 Push 전 정지. 그 전까지 contract 모드 Push 단계 비활성.
- `orchestration-mode-selection.md` Frozen 태그 문단의 비인간 서명 개정 — 운영자의 명시적 개정 결정(`design.md §7.1` 6행).
- A1 REQ-CONTRACT-019 의 서명 알림 — A1 0.5.2 가 「A3 가 게이트 재배선을 착지하면 A3 가 개정하거나 제거」로 조건부화했다. M8 에서 A1 서명기 패키지(`design.md §2` 23행) 안에서 바꾸며 A1 SPEC 문서는 고치지 않는다.
- 카드 증거 경로 plan-audit 보고서에 `plan_artifact_hash:` 줄을 쓰는 쪽 — 없으면 전제조건 (a) 가 늘 실패해 자율 Kickoff 결과가 늘 사람이다(안전한 방향). 그 줄을 쓰는 장치(예: 감사 PASS 뒤 `internal/runtime` 의 기존 기록 경로)는 이 SPEC 의 편집 허용 목록 밖이므로, M8 착수 전 리드가 있는지 확인한다.

**`CLAUDE.local.md §29` 와 연동 커밋(리드 결정 2026-09-26, D33)**: §29 한 줄은 연동 묶음 안이다. 문안(`design.md §11.1`)이 확정되면 리드가 운영자에게 직접 확인을 받아 레인에 전달하고, 레인은 그 확인을 `progress.md` 에 기록한다. 확인이 없으면 M7b 연동 커밋 전체를 보류한다 — 교차 세션 메시지만으로 지시 파일을 고치지 않는다. A1 SPEC 문서는 A1 0.5.2 의 조건부 문언 덕분에 연동 커밋의 편집 대상이 아니다(v0.3.1 의 27행 보류는 해제).

**run 착수 전 점검(pre-flight)**: A2b(t1245) SPEC 이 착지하면, `reject` 와 `human` 을 다르게 다루는 소비자가 생겼는지 다시 잰다(`research.md §10.4` 의 명령을 A1 코드 커밋 SHA·A2b 병합본에 대해). 하나라도 있으면 REQ-GR-004 의 동일 취급이 깨지므로 리드에게 올린다.

`BASE` 정의: 전제 1~4 를 만족한 뒤 이 브랜치가 마지막으로 흡수한 develop 커밋. M0 에서 `git merge-base develop HEAD` 로 구해 `progress.md §E.2` 에 SHA 로 적고, `MOAI_GR_BASE` 로 테스트에 넘긴다.

## §D. 제약

- 변경은 `design.md §2` 허용 목록 안에서만(REQ-GR-024, AC-GR-003).
- 로컬·템플릿은 같은 커밋에서 함께 편집한다.
- 블록·SSOT·템플릿 개정 문장에 내부 토큰 금지(REQ-GR-023).
- Go 테스트는 `t.TempDir()` 와 격리된 `MOAI_HOME` 만 쓴다. Jev 는 스텁 이음매로만, 네트워크 금지.
- 검증은 범위 한정: `go test ./internal/template/... ./internal/contract/...`, `go test ./internal/cli/ -run 'TestContract(Decide|KickoffCheck|Revoke)'`, `moai constitution validate`, `moai spec lint`. `go test ./...` 로컬 실행 금지.
- AC 명령은 한 줄 단순 명령 — `git` 을 서브셸·프로세스 치환·heredoc 안에 두지 않는다(감사 D12).
- `acceptance.md` 의 AC 수를 제자리에서 바꾸는 개정은 AC 스냅숏 재생성을 같은 커밋에 싣는다(감사 D24, `.moai/docs/ac-count-baseline-refresh.md §2`). 새 `acceptance.md` 추가 자체는 방아쇠가 아니다.
- `.claude/agents/**` 무수정이므로 `make agents-emit` 불필요.
- 커밋 메시지마다 카드 id `t1236`.

## §E. 자기 검증 (run phase 완료 보고에 포함)

acceptance.md 의 AC 전부를 PASS/FAIL 표로, 명령·원문 출력·exit 코드·트리 SHA 와 함께. Go 마일스톤은 RED 원문(`--- FAIL`)을 GREEN 앞에 남긴다. `MOAI_GR_BASE` 를 쓰는 테스트는 `--- SKIP` 이 아니라 `--- PASS` 임을 보인다.

## §F. 마일스톤 (결정이 바뀔 가능성이 큰 것부터)

### M0 — 흡수와 재앵커 (Priority High, 선행)

- §C 전제 확인 → develop 흡수 → `BASE` 기록.
- `research.md §1` 계수를 `BASE` 에서 재실행. E/R/H 분류가 바뀌면 blocker 로 manager-spec 재위임.
- A1 병합본과 0.5.2(`25283ebf8`) 인용 항목을, A2 병합본과 **[A2 개정본으로 재확인]** 항목을, A2b 병합본과 **[A2b SPEC 으로 재확인]** 항목을 대조. 차이는 blocker.
- 절 제목 앵커가 t1175 이후에도 있는지 확인.

### M1 — SSOT `contract-autonomy.md` (Priority High)

절 10개(`design.md §3`). 「Autonomous Kickoff」 절은 한 문장 자리표시만 두고 `llm+jev` 리터럴을 쓰지 않는다(AC-GR-005·017 이 이 조건에 기댄다). 로컬·템플릿 동시 생성.

### M2 — G1 발화 지점 (Priority High)

행 2·4·5·6·7·8·9·10·11. `orchestration-mode-selection.md` 블록은 사람 서명 등가만 담는다(D4). 모든 블록은 결정자 `human` 경로만 서술한다.

### M3 — G2/G3/G4 (Priority Medium)

`CLAUDE.md` §7, `askuser-protocol.md`. plan phase 인터뷰는 건드리지 않는다(D-2).

### M4 — G7/G8 (Priority Medium)

`spec-assembly.md` Step 2.3.5·Phase 15, `moai.md` Step 11.3.

### M5 — G11 과 생명주기 (Priority Medium)

`run.md` 생명주기 블록, `sync.md`, `sync/doc-execution.md`, `sync/delivery.md`. Push 행은 A4 착지 전 비활성으로 적는다.

### M6 — `moai contract revoke` (Priority High, TDD, 자율 Kickoff 보다 먼저)

`internal/contract/receipt/`(체인의 최소 부분) + `internal/contract/revoke/` + `internal/cli/contract_revoke.go`. RED: `design.md §12` revoke 행. **M8 보다 먼저 커밋한다(AC-GR-017).**

### M7 — 사건 저장소 완성 + 서명기 사건 추가 + decide + kickoff-check (Priority High, TDD)

`internal/contract/receipt/` 완성(`$MOAI_HOME/db/<project-key>/contract/` 의 `receipts.jsonl`·`events.jsonl`), `internal/contract/sign/` 에 서명·재봉인 사건 추가(주입 가능한 저장소 이음매, 모든 서명 테스트를 `t.TempDir()` `MOAI_HOME` 으로 — D37)와 서명기 단계 (1) 의 `jevDoctrineAmended` 조건화(동작 불변), `internal/contract/kickoff/`(전제조건·결정 규칙 R1~R5·decide·kickoff-check, 상수 `autonomousKickoffEnabled`·`jevDoctrineAmended` 모두 `false`), `internal/cli/contract_decide.go`·`contract_kickoff_check.go`. RED: 저장소·서명 사건·kickoff-check·decide 행, 결정 규칙은 규칙마다 따로(AC-GR-019·025).

### M7b — Jev 원칙 개정 연동 커밋 (Priority High)

`design.md §11` 의 모든 위치와 `jevDoctrineAmended = true`(R3 와 A1 서명기 단계 (1) 의 해제 — 둘 다 M7 에서 이 상수로 조건화됨)를 **한 커밋**으로 싣는다(REQ-GR-025, AC-GR-017). 조립 순서는 리드 결정 2026-09-26(D32): ① manager-spec 이 `SPEC-JEV-CORE-001` 본문(REQ-JEVC-011·012, `Out of Scope — authority` 두 항목, HISTORY·`version`)을 작성하고 돌아온다 → ② 그 뒤 manager-develop 이 상수·규칙·설정 문장과(운영자 확인이 있으면) `CLAUDE.local.md §29` 한 줄을 작성하고 돌아온다 — 동시에 쓰지 않는다 → ③ 레인 오케스트레이터가 두 결과를 명시 경로로 스테이징해 한 커밋을 만든다. 운영자의 §29 확인이 없으면 이 커밋 전체를 보류한다(D33). 검증 범위: `go test ./internal/contract/... ./internal/template/ ./internal/spec/`, `moai spec lint SPEC-JEV-CORE-001`(D42).

### M8 — 자율 Kickoff 활성 (Priority Medium, 보류 가능)

`design.md §7.1` 1~6행 확인 → SSOT 「Autonomous Kickoff」 절과 발화 지점에 전용 블록 `contract-autonomous-kickoff` 추가, 활성 상수 `true`. 이 커밋은 M6·M7 커밋들의 엄격한 후손이어야 한다. 3·4·6행 중 하나라도 미충족이면 보류하고 리드에게 보고한다.

### M9 — 가드 테스트 (Priority Medium, TDD)

`contract_mode_blocks_test.go`, `contract_mode_guided_test.go`. RED: 불량 픽스처 다섯 종과 「검사 대상 0건」. M1~M5 와 함께 진행해도 되지만 RED 원문은 블록이 생기기 전에 남긴다.

### M10 — 기계적 마무리 (Priority Low)

`make build`, AC 전부, `moai constitution validate`, `moai spec lint`(이 SPEC 과 `SPEC-JEV-CORE-001`).

## §G. 반패턴

- guided 문장을 「조금 다듬기」 — AC-GR-001 FAIL.
- evolvable 구간 안에 블록 삽입.
- 블록·템플릿 문장에 G 번호·SPEC ID·카드 id·날짜.
- 참조 지점까지 블록을 뿌리기 — AC-GR-003 FAIL.
- M1 SSOT 에 `llm+jev` 리터럴을 쓰기 — AC-GR-005 FAIL.
- revoke 보다 먼저, 또는 조건 커밋과 같은 커밋에서 활성 상수를 켜기 — AC-GR-017 FAIL.
- 원칙 개정·R3 해제·A1 임시 규칙 해제·§29 를 나눠 착지 — AC-GR-017 FAIL.
- manager-spec 과 manager-develop 을 동시에 돌려 연동 커밋을 만들기 — 한 번에 한 작성자(D32).
- 결정하지 않는 decide 경로에서 A1 영수증 파일을 쓰기 — A1 필드 규칙 2·6 위반(D29).
- git 트레일러 파서(`%(trailers)`)로 `Authored-By-Agent:` 를 읽기 — 이 저장소 커밋에서 빈 값(D28).
- 저장소를 「변조 방지」로 서술 — 흔적만 남긴다.
- 에이전트가 쓴 영수증 파일을 권위 있는 영수증으로 취급 — kickoff-check 가 거절해야 한다.
- 결정 규칙 R1~R5 를 A1 validator 에 맡기거나 A1 문서에 다시 정의 — 규칙은 A3 소유, A1 은 스키마만.
- Jev 단독 결정자를 허용하거나, 원칙 개정 전에 Jev 답이 결정하게 두기.
- Jev 실패를 대체 기록 없이 `llm` 으로 넘기기(조용한 대체).

## §H. 교차 참조

- `spec.md` 요구사항·본문 결정, `acceptance.md` 판정, `design.md` 허용 목록·마커 규약·SSOT·활성 조건(§7.1 정본)·decide·저장소·revoke·Jev 개정, `research.md` 인벤토리·A1/A2 기준선·상충.
- `SPEC-AUTONOMY-CONTRACT-001` (A1 0.5.2 `25283ebf8`), `SPEC-AUTONOMY-ESCALATION-001` (A2 — 개정본 대기), A2b(t1245 — SPEC 미확인), `SPEC-JEV-CORE-001` (REQ-JEVC-007·011·012), `SPEC-ALWAYS-LOADED-DIET-002` (t1175).
- `.moai/reports/t1236/plan-audit-1.md` — iter-1 감사 보고서.

## §I. 감사 결함 처분 (plan-audit iter-1, `.moai/reports/t1236/plan-audit-1.md`)

| 결함 | 처분 | 어디서 |
|---|---|---|
| D1 미해결 질문 9건 | **수리** — 전부 결정·기본값으로 닫음, 표지 0건 | §B, `spec.md §B` |
| D2 REQ 번호 간격 | **수리** — 001~025 연속, 모든 참조 갱신 | 여섯 파일 |
| D3 영수증 출처 거절 소유자 | **수리** — A3 가 Go(kickoff-check)로 소유, 저장소에 없는 영수증 거절 | REQ-GR-007, AC-GR-016 |
| D4 Frozen 태그 문단 재해석 | **수리** — 등가를 사람 서명에 한정, 비인간 서명은 별도 개정을 활성 조건 6으로 | REQ-GR-005·008, `design.md §6`·`§7.1` |
| D5 AC 순서 검사 | **수리** — 전용 블록 id + 활성 상수 + 엄격한 후손 Go 테스트, 비활성 시 명시 출력 | AC-GR-017 |
| D6 Jev·LLM 단독 거절의 기계 근거 | **수리(v0.3.1 에서 방향 변경)** — 운영자 재결정으로 Jev 단독만 거절: kickoff-check `decider-not-permitted`(`signer_kind: jev`)와 decide 규칙 R5. v0.3.0 의 「`llm` 단독 거절」은 철회 — `llm` 은 기본 결정자다 | REQ-GR-007·010, AC-GR-016·019 |
| D7 두 번째 리뷰 정지 소유 | **수리** — A4 로 이동, Push 단계 활성 전제 | REQ-GR-021, §C |
| D8 합의 규칙 이중 정의 | **수리(v0.3.1 에서 소유 변경)** — 결정 규칙 R1~R5 는 이 SPEC 이 정본으로 소유(리드 2026-09-26), A1 은 스키마만. 규칙·갈래마다 RED 먼저의 Go 테스트 | REQ-GR-010, AC-GR-019·025 |
| D9 revoke 「시작된 run」 | **수리** — 서명된 계약이면 서명 종류 무관 | REQ-GR-022, AC-GR-023 |
| D10 A2 기록 경로·revoke 종류 | **수리** — 리드가 정한 A2 최종 형식 인용, `d8926ff9a` 는 철회 형식으로 기록, 종류 요청을 전제에 명시 | REQ-GR-018·022, §C, `research.md §10.1` |
| D11 guided 보존 범위 | **수리** — 주장을 블록 밖 텍스트로 좁힘, 증가량 수치화, SSOT `paths:` 를 계약 파일로 | REQ-GR-002, `design.md §3`·`§4`, AC-GR-010 |
| D12 AC 명령 실행 가능성 | **수리** — 모든 AC 를 `go test` 또는 단순 명령으로 | acceptance 전체 |
| D13 범위 밖 파일 가드 | **수리** — 변경 집합 허용 목록 테스트 | REQ-GR-024, AC-GR-003 |
| D14 통합이 가린 의무 | **수리** — AC-GR-020 에 `backlog.db`·LLM 호출 0 추가, 단계별 증거 AC-GR-007 신설. SPEC 분할은 **보류** — 요구사항 25개가 상한 안에 있고, Go 층과 문서 층이 활성 순서로 서로 묶여 있어 나누면 순서 검사가 두 SPEC 에 걸친다 | AC-GR-007·020 |
| D15 전제조건 (a)(f) 모호 | **수리** — (a) 파일 선택 규칙과 파싱 규칙, (f) A1 `frozen-files` 참조 | REQ-GR-009, `design.md §8` |
| D16 작성자 배제의 자기 신고 | **수리** — moai 측 세션 식별자와 트레일러 병행, 신고임을 명시 | REQ-GR-010 |
| D17 정의되지 않은 참조 | **수리** — 재번호와 함께 정리 | 전 파일 |
| D18 활성 순서 서술 불일치 | **수리** — `design.md §7.1` 단일 정본, 나머지는 참조 | REQ-GR-008, progress.md |
| D19 존재 검사의 사소한 통과 | **수리** — 절 단위 토큰 검사 | AC-GR-005, `design.md §5.1` |
| D20 SSOT 순서 검사 취약 | **수리** — 생명주기 절 안에서만 | AC-GR-006 |
| D21 꼬리 절단 | **수리** — 한계 명시, 두 번째 증인과 `signature-not-recorded` 로 보완, AC 행 추가 | REQ-GR-012, AC-GR-021 |
| D22 REQ-003 복합절 | **수리** — contract 모드 사람 경로를 REQ-GR-004 로 이동 | REQ-GR-003·004 |
| D23 REQ-040 실패 경로 명칭 | **수리** — 단계 이름으로 한정, CI autofix 루프 제외 명시 | REQ-GR-016 |
| D24 AC 스냅숏 | **수리** — 조건 한 줄 | §D |
| D25 A1 표지 누락 | **수리** — A1 0.5.1 이 확정한 항목은 인용(A1 재확인 표지 0건), 남은 표지는 A2·A2b·A4 항목 | 전 파일 |

**v0.3.1 요구사항 재배치**: revoke 의 두 요구(옛 REQ-GR-022·023)를 REQ-GR-022 로 합치고, 옛 023 은 022 에 흡수하고 옛 024·025 를 023·024 로 옮긴 뒤, 원칙 개정 연동 요구를 REQ-GR-025 로 신설했다. 요구사항 25개, AC 25개(AC-GR-025 신설).

## §J. 감사 결함 처분 (plan-audit iter-2, `.moai/reports/t1236/plan-audit-2.md`, 감사 대상 `1b071a573`)

| 결함 | 처분 | 어디서 |
|---|---|---|
| D26 AC-GR-016 생존 변이 3종 | **수리** — 사유 `receipt-not-approved`·`not-signed-valid`(verify 상태 동봉) 추가, `effective_decider` ≠ `signer_kind` 를 `decider-mismatch` 에 포함, 픽스처 (12)~(16) | REQ-GR-007, AC-GR-016, `design.md §12` |
| D27 reject/human 라우팅이 요구가 아님 | **수리** — REQ-GR-004 에 「`reject`·`human` 은 둘 다 서명 거절·사람 결정 필요, 구별하지 않음」 문장, §5.1 「The signing gate」 토큰 `reject`·`human decision required`, AC-GR-014 블록 검사, AC-GR-016 (12)(13). reject/human 소비자 실측을 기록하고 run pre-flight 재측정 항목 추가 | REQ-GR-004, AC-GR-014·016, `design.md §5.1`, `research.md §10.4`, §C |
| D28 작성자 배제 데이터 원천 | **수리** — 원천을 `internal/spec` `parseAuthoredByAgent`(정규식 판독기)로 지정, 세션 식별자 조건 삭제(원천 없음), 트레일러 없음 → `author-check-unmeasured` → 사람, 자기 신고의 잔여 위험 명시, 픽스처(서명 줄이 붙은 본문) 추가 | REQ-GR-010, AC-GR-019, `design.md §8`·`§6`, `research.md §10.5` |
| D29 비결정 경로의 영수증 스키마 | **수리** — 결정하지 않는 경로는 A1 영수증을 쓰지 않고 `events.jsonl` 사건만, 사유 코드는 사건·`--json` 에만(A1 필드 규칙 1·2·6 근거) | REQ-GR-011, AC-GR-018·019·020, `design.md §8` |
| D30 판정의 산출물 결합과 PASS-WITH-DEBT | **수리** — `plan_artifact_hash:` 가 현재 plan-artifact 해시와 같아야 (a) 성립(`internal/runtime` 재사용), `PASS-WITH-DEBT` 는 문턱 이상일 때만, 픽스처 4종 | REQ-GR-009, AC-GR-018, `design.md §8`, §C(쓰는 장치 확인) |
| D31 Jev 개정 위치 누락 | **수리** — REQ-JEVC-011 과 `Out of Scope — authority` 두 항목을 개정 위치·연동 묶음에 추가, AC-GR-022 에 자기모순 불량 픽스처 2개 | REQ-GR-013·025, AC-GR-022, `design.md §11` |
| D32 소유권 교차 | **수리(리드 결정 2026-09-26)** — 「한 커밋」 유지, manager-spec → manager-develop 차례 작성, 레인 오케스트레이터가 명시 경로로 한 커밋 | REQ-GR-025, M7b, `design.md §11.2` |
| D33 §29 분리 | **수리(리드 결정 2026-09-26, 선택지 a)** — §29 를 연동 묶음 안에, 운영자 확인 없으면 연동 커밋 전체 보류, 문안은 `design.md §11.1` | REQ-GR-013·025, §C, AC-GR-017·022 |
| D34 연동 표지 미정의 | **수리** — R3 와 A1 서명기 단계 (1) 을 한 상수 `jevDoctrineAmended` 로 조건화, 표지는 동작 테스트 `TestSignInterimRuleFollowsDoctrine`(A1 AC-CONTRACT-016 (t) 대체), 「앞 커밋에서 규칙 제거」 픽스처 **[A1 개정본으로 재확인 — `25283ebf8`]** | REQ-GR-025, AC-GR-017, `design.md §7.1` |
| D35 D-1 과 SSOT 절 1 모순 | **수리** — `kickoff-check --json` 으로 | `design.md §3` |
| D36 AC-GR-009 빈 선택 | **수리** — `TestTemplateNoInternalContentLeak`, `-list` 원문 첨부, 다른 AC 의 선택 확인 표 | AC-GR-009, `acceptance.md §B.1` |
| D37 A1 서명 테스트 격리 | **수리** — 23행에 테스트와 `MOAI_HOME` 격리·저장소 이음매 명시 | `design.md §2`, M7 |
| D38 R1 escalate 조합 | **수리** — 두 조합 추가 | AC-GR-019, `design.md §12` |
| D39 활성 여부 관측 불가 | **수리** — kickoff-check `--json` 에 `autonomous_kickoff_enabled`·`jev_doctrine_amended` | REQ-GR-007, AC-GR-016 |
| D40 동사의 루트 | **수리** — 본문 결정 D-11(호출한 작업 트리의 최상위) | `spec.md §B` |
| D41 `card-mismatch` exit 차이 | **수리** — 본문 결정 D-12(판정 명령 1, 쓰기 명령 2) | `spec.md §B`, REQ-GR-007 |
| D42 `./internal/spec` 재측정 | **수리** — M7b 검증 범위에 포함(다른 SPEC 본문 변경). A1 `acceptance.md` 는 더 이상 편집하지 않으므로 AC 스냅숏 재생성 의무는 생기지 않는다 | M7b, `design.md §11.2` |

요구사항 25개·AC 25개 그대로(새 id 없음, 기존 문장과 픽스처만 수정). v0.3.1 의 `design.md §2` 27행(A1 SPEC 문언 개정)은 A1 0.5.2 가 문언을 조건부화해 필요 없어졌으므로 삭제했고, 그에 걸려 있던 M7b 보류도 해제했다.
