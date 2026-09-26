# plan.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.0)

## §A. 맥락

- 카드 t1236 (AUTONOMY-A3), 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`. 이력: v0.1.0 `ca1e39f6f`, v0.2.0 `62a65f709`(plan-audit iter-1 FAIL 0.71, `.moai/reports/t1236/plan-audit-1.md`), v0.3.0 은 그 수리본.
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L).
- 개발 방식: 문서 층(추가형 블록 + Jev 원칙 개정) + Go 코드(kickoff-check·decide·revoke·사건 저장소·A1 서명기 사건 추가) + Go 가드 테스트. 모두 TDD — RED 를 먼저 관측한다(`design.md §12`).
- A1 기준: 0.4.1 `6d98ca466`. A2 기준: 리드가 전달한 최종 형식(A2 개정본 대기, 현재 커밋 `d8926ff9a` 는 철회된 형식 — `research.md §10.1`).

## §B. 결정 기록 (미해결 질문 0건)

v0.2.0 의 NC-1~NC-9 는 모두 닫혔다. 본문 정본은 `spec.md §B` 「본문 결정」 D-1~D-9 이고, 여기에는 근거만 적는다.

| 항목 | 처분 | 근거 |
|---|---|---|
| NC-1 모드 판독 | 결정 D-1: `moai contract kickoff-check --json` 의 `mode`·`decider` | 설계 결정(감사 D1 분류). A1 0.4.1 `show --json` 에 결정자 필드가 없어 A3 명령이 채운다 |
| NC-2 plan phase 인터뷰 | **기본값 D-2**: 서명 이후 작업에만 적용. 운영자가 바꿀 수 있음 | 리드 승인(2026-09-26). 계약이 plan-audit 뒤에 서명되므로 논리적으로 강제된다(감사 NC 분류) |
| NC-3 G4 원문 | 결정 D-3: 헌법 파일 무수정 | 헌법 개정은 범위 밖 — 바인딩 범위가 답을 강제(감사 D1) |
| NC-4 서명 후 goal 무장 | **기본값 D-4**: 자동 무장 없음. 운영자가 바꿀 수 있음 | 리드 승인(2026-09-26). 보수적 기본값 — run 은 이 기본값으로 진행할 수 있다 |
| NC-5 초안 작성 지시 | 결정 D-5: `spec-assembly.md` 위임문 | `.claude/agents/**` 제외가 답을 강제(감사 D1) |
| NC-6 always-loaded 예산 | 결정 D-6: 블록 합계 1,500자 + `jev_ask` 행 300자 | `design.md §4` 에 수치로 묶음. t1175 조율은 리드 몫 |
| NC-7 Jev 게이트 입력 | 결정 D-7: 리드 결정 L1 — contract 모드 Kickoff 교차 확인 한정 개정을 이 카드 범위로 | REQ-GR-013, `design.md §11` |
| NC-8 주 LLM 결정자 | 결정 D-8: 리드 결정 L3 — 리드 세션 또는 새 컨텍스트 판단 역할 | `mission-governor` 정의가 승인을 배제 |
| NC-9 revoke 정지 시점 | 결정 D-9: 다음 단계 경계 | REQ-GR-020·022 가 이미 규정(감사 NC 분류) |

**운영자에게 남은 질문**: 없음. D-2·D-4 는 운영자가 바꿀 수 있는 기본값으로 기록했고, run 은 기본값으로 진행한다.

## §C. 착수 전제와 기준 ref

run phase 진입 조건 (모두 필요):

1. **t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) develop 병합.** 확인: `moai contract verify --help` exit 0.
2. **t1235 (A2, `SPEC-AUTONOMY-ESCALATION-001`) develop 병합** (리드 결정). 에스컬레이션 기록의 **최종 형식**(`escalation/<class>-<fingerprint>.md` + YAML 머리 + `revoke` 종류)이 A2 개정본에 들어 있어야 한다 — 현재 커밋 `d8926ff9a` 의 JSON 형식이면 M6 을 보류하고 리드에게 보고.
3. **t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) develop 병합 후 이 브랜치에 흡수.**

Push 단계(REQ-GR-021)와 자율 Kickoff 활성(M8)에만 걸리는 추가 조건 — run 진입 전제는 아니다:

- **t1237 (A4)** 착지 — 두 번째 리뷰 미수행 시 Push 전 정지. 그 전까지 contract 모드 Push 단계 비활성.
- push 직렬화 강제와 에이전트발 대화형 `sign` 차단 — 소유 카드가 A1 0.4.1 에서는 A2b t1245, 리드·A2 초안에서는 A2 t1235 로 상충한다(`research.md §10.2`). M8 착수 전 리드가 어느 카드로 착지했는지 확인한다.
- `orchestration-mode-selection.md` Frozen 태그 문단의 비인간 서명 개정 — 운영자의 명시적 개정 결정(`design.md §7.1` 6행).

**`CLAUDE.local.md §29` 편집 전제**: 운영자가 레인 세션에서 그 편집을 확인한다. 교차 세션 메시지만으로는 지시 파일을 고치지 않는다. 확인이 없으면 M7b 에서 그 한 줄만 보류하고 나머지 네 위치는 진행한다.

`BASE` 정의: 전제 1~3 을 만족한 뒤 이 브랜치가 마지막으로 흡수한 develop 커밋. M0 에서 `git merge-base develop HEAD` 로 구해 `progress.md §E.2` 에 SHA 로 적고, `MOAI_GR_BASE` 로 테스트에 넘긴다.

## §D. 제약

- 변경은 `design.md §2` 허용 목록 안에서만(REQ-GR-025, AC-GR-003).
- 로컬·템플릿은 같은 커밋에서 함께 편집한다.
- 블록·SSOT·템플릿 개정 문장에 내부 토큰 금지(REQ-GR-024).
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
- A1 병합본과 0.4.1(`6d98ca466`) 인용 항목을 대조, A2 병합본과 **[A2 개정본으로 재확인]** 항목을 대조. 차이는 blocker.
- 절 제목 앵커가 t1175 이후에도 있는지 확인.

### M1 — SSOT `contract-autonomy.md` (Priority High)

절 10개(`design.md §3`). 「Autonomous Kickoff」 절은 한 문장 자리표시만 두고 결정자 쌍 값 리터럴을 쓰지 않는다(AC-GR-005·017 이 이 조건에 기댄다). 로컬·템플릿 동시 생성.

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

`internal/contract/receipt/` 완성, `internal/contract/sign/` 에 서명·재봉인 사건 추가, `internal/contract/kickoff/`(전제조건·합의·decide·kickoff-check·활성 상수 `false`), `internal/cli/contract_decide.go`·`contract_kickoff_check.go`. RED: 저장소·서명 사건·kickoff-check·decide 행.

### M7b — Jev 원칙 개정 (Priority High)

`design.md §11` 의 다섯 위치. `CLAUDE.local.md §29` 는 운영자 확인 뒤에만. 개정 착지 뒤 A1 합의 규칙 0번 해제 커밋(별도 커밋 — AC-GR-017 의 조건 커밋).

### M8 — 자율 Kickoff 활성 (Priority Medium, 보류 가능)

`design.md §7.1` 1~6행 확인 → SSOT 「Autonomous Kickoff」 절과 발화 지점에 전용 블록 `contract-autonomous-kickoff` 추가, 활성 상수 `true`. 이 커밋은 M6·M7·M7b 커밋들의 엄격한 후손이어야 한다. 3·4·6행 중 하나라도 미충족이면 보류하고 리드에게 보고한다.

### M9 — 가드 테스트 (Priority Medium, TDD)

`contract_mode_blocks_test.go`, `contract_mode_guided_test.go`. RED: 불량 픽스처 다섯 종과 「검사 대상 0건」. M1~M5 와 함께 진행해도 되지만 RED 원문은 블록이 생기기 전에 남긴다.

### M10 — 기계적 마무리 (Priority Low)

`make build`, AC 전부, `moai constitution validate`, `moai spec lint`(이 SPEC 과 `SPEC-JEV-CORE-001`).

## §G. 반패턴

- guided 문장을 「조금 다듬기」 — AC-GR-001 FAIL.
- evolvable 구간 안에 블록 삽입.
- 블록·템플릿 문장에 G 번호·SPEC ID·카드 id·날짜.
- 참조 지점까지 블록을 뿌리기 — AC-GR-003 FAIL.
- M1 SSOT 에 결정자 쌍 값 리터럴을 쓰기 — AC-GR-005 FAIL.
- revoke 보다 먼저, 또는 조건 커밋과 같은 커밋에서 활성 상수를 켜기 — AC-GR-017 FAIL.
- 저장소를 「변조 방지」로 서술 — 흔적만 남긴다.
- 에이전트가 쓴 영수증 파일을 권위 있는 영수증으로 취급 — kickoff-check 가 거절해야 한다.
- 합의 규칙을 A1 과 다르게 다시 정의.

## §H. 교차 참조

- `spec.md` 요구사항·본문 결정, `acceptance.md` 판정, `design.md` 허용 목록·마커 규약·SSOT·활성 조건(§7.1 정본)·decide·저장소·revoke·Jev 개정, `research.md` 인벤토리·A1/A2 기준선·상충.
- `SPEC-AUTONOMY-CONTRACT-001` (A1 0.4.1 `6d98ca466`), `SPEC-AUTONOMY-ESCALATION-001` (A2 — 개정본 대기), `SPEC-JEV-CORE-001` (REQ-JEVC-007·011·012), `SPEC-ALWAYS-LOADED-DIET-002` (t1175).
- `.moai/reports/t1236/plan-audit-1.md` — iter-1 감사 보고서.

## §I. 감사 결함 처분 (plan-audit iter-1, `.moai/reports/t1236/plan-audit-1.md`)

| 결함 | 처분 | 어디서 |
|---|---|---|
| D1 미해결 질문 9건 | **수리** — 전부 결정·기본값으로 닫음, 표지 0건 | §B, `spec.md §B` |
| D2 REQ 번호 간격 | **수리** — 001~025 연속, 모든 참조 갱신 | 여섯 파일 |
| D3 영수증 출처 거절 소유자 | **수리** — A3 가 Go(kickoff-check)로 소유, 저장소에 없는 영수증 거절 | REQ-GR-007, AC-GR-016 |
| D4 Frozen 태그 문단 재해석 | **수리** — 등가를 사람 서명에 한정, 비인간 서명은 별도 개정을 활성 조건 6으로 | REQ-GR-005·008, `design.md §6`·`§7.1` |
| D5 AC 순서 검사 | **수리** — 전용 블록 id + 활성 상수 + 엄격한 후손 Go 테스트, 비활성 시 명시 출력 | AC-GR-017 |
| D6 Jev·LLM 단독 거절의 기계 근거 | **수리** — kickoff-check 가 `decider-not-permitted` | REQ-GR-007, AC-GR-016 |
| D7 두 번째 리뷰 정지 소유 | **수리** — A4 로 이동, Push 단계 활성 전제 | REQ-GR-021, §C |
| D8 합의 규칙 이중 정의 | **수리** — A1 정본 참조, A1 validator 대조 테스트 | REQ-GR-010, AC-GR-019 |
| D9 revoke 「시작된 run」 | **수리** — 서명된 계약이면 서명 종류 무관 | REQ-GR-022, AC-GR-023 |
| D10 A2 기록 경로·revoke 종류 | **수리** — 리드가 정한 A2 최종 형식 인용, `d8926ff9a` 는 철회 형식으로 기록, 종류 요청을 전제에 명시 | REQ-GR-018·022, §C, `research.md §10.1` |
| D11 guided 보존 범위 | **수리** — 주장을 블록 밖 텍스트로 좁힘, 증가량 수치화, SSOT `paths:` 를 계약 파일로 | REQ-GR-002, `design.md §3`·`§4`, AC-GR-010 |
| D12 AC 명령 실행 가능성 | **수리** — 모든 AC 를 `go test` 또는 단순 명령으로 | acceptance 전체 |
| D13 범위 밖 파일 가드 | **수리** — 변경 집합 허용 목록 테스트 | REQ-GR-025, AC-GR-003 |
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
| D25 A1 표지 누락 | **수리** — A1 0.4.1 이 확정한 항목은 인용으로 교체, 남은 표지는 A2·A4 항목 | 전 파일 |
