# acceptance.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.3)

모든 AC 는 Given-When-Then 이며 명령과 기대 출력으로 판정한다. 명령은 워크트리 세션 가드를 통과하도록 **한 줄짜리 단순 명령**만 쓴다 — `git` 을 `$( )`·`<( )`·heredoc 안에 두지 않는다. `git` 이 필요한 검사는 Go 테스트 안에서 실행하고, 기준 ref 는 환경 변수 `MOAI_GR_BASE`(= `progress.md §E.2` 에 기록된 `BASE` SHA)로 넘긴다. 브랜치 이름을 기준으로 쓰지 않는다.

**빈 선택은 통과가 아니다.** 모든 `go test -run` AC 는 출력에 `[no tests to run]` 이 없고, 명시된 테스트 이름마다 `--- PASS` 줄이 있어야 한다. `MOAI_GR_BASE` 를 요구하는 테스트가 `--- SKIP` 을 내면 FAIL 이다.

## §A. 수용 기준

### AC-GR-001 — 블록 밖 텍스트 보존 (REQ-GR-001·002)

- **Given** `design.md §2` 1~14행의 기존 파일 13개와 템플릿 사본(26개)
- **When** contract-mode 블록을 걷어낸다
- **Then** 결과가 `BASE` 의 같은 파일과 바이트 동일하고, 걷어낸 블록은 26개 사본 합계 1개 이상이다

```bash
MOAI_GR_BASE=<BASE> go test ./internal/template/ -run 'TestContractModeGuidedPreservation' -count=1 -v
```

기대: `--- PASS: TestContractModeGuidedPreservation`. 이 AC 는 **블록 밖 텍스트의 보존**만 판정한다 — guided 세션이 블록을 읽는 증가는 AC-GR-010 이 묶는다.

### AC-GR-002 — 로컬↔템플릿 동등과 승계 분기 불변 (REQ-GR-023)

- **Given** 편집 대상 파일 쌍과 신규 SSOT
- **When** 두 사본의 블록을 추출하고, 걷어낸 두 사본의 차이를 기준 ref 두 사본의 차이와 비교한다
- **Then** 블록이 바이트 동일하고 SSOT 가 통째로 동일하며, 승계 분기가 변하지 않는다

```bash
MOAI_GR_BASE=<BASE> go test ./internal/template/ -run 'TestContractModeLocalTemplateParity|TestContractModeInheritedDivergence' -count=1 -v
```

기대: 두 테스트 `--- PASS`.

### AC-GR-003 — 변경 집합이 허용 목록 안 (REQ-GR-024)

- **Given** `design.md §2` 의 허용 목록
- **When** `BASE` 대비 변경된 파일 목록을 읽는다
- **Then** 목록 밖 파일이 0개다 — 특히 `moai-constitution.md`·`zone-registry.md`·`.claude/agents/**`·`.claude/output-styles/**`·`ci-autofix-protocol.md`·`context-window-management.md`·`agent-common-protocol.md`·`spec-workflow.md`·`session-handoff-examples.md`·`internal/kanban/**` 가 포함되지 않는다

```bash
MOAI_GR_BASE=<BASE> go test ./internal/template/ -run 'TestContractModeChangeSetAllowlist' -count=1 -v
moai constitution validate
```

기대: 테스트 `--- PASS`; 둘째 명령 exit 0, 첫 줄이 `constitution validate: OK — no drift or violations detected` 로 시작하고 검사 항목 수가 `BASE` 에서 잰 값과 같다(`ca1d5dc43` 에서 `97 of 101`).

### AC-GR-004 — G1 발화 지점 전수 (REQ-GR-005)

- **Given** `BASE` 에서 `Kickoff` 문자열을 담은 파일 집합
- **When** 그 집합을 `research.md §1.2` 의 E ∪ R ∪ H ∪ 로컬 전용 분류(M0 재분류 반영)와 비교하고, E 의 각 파일(로컬·템플릿)에서 블록 수를 센다
- **Then** 분류되지 않은 파일이 0개이고, E 의 20개 사본 모두 블록이 1개 이상이다

```bash
MOAI_GR_BASE=<BASE> go test ./internal/template/ -run 'TestContractModeEmitterSites' -count=1 -v
```

기대: `--- PASS`.

### AC-GR-005 — SSOT 절 구조와 절 안의 필수 토큰 (REQ-GR-005·008·010·014·016·017·018·021·022)

- **Given** 신규 SSOT
- **When** `design.md §3` 의 절 제목 10개를 찾고, `§5.1` 표의 토큰을 **그 절 안에서만** 찾는다(파일 전체 검색이 아니다)
- **Then** 절 10개가 있고 각 절의 토큰이 모두 그 절에 있다. 「Equivalence clause」 절에 `llm+jev`·`signer_kind: llm` 이 없다(사람 서명 한정, D4). 「Autonomous Kickoff」 절의 토큰 검사는 전용 블록 `contract-autonomous-kickoff` 가 있을 때만 수행하고, 없으면 그 절에 `llm+jev` 리터럴이 없음을 검사한다

```bash
go test ./internal/template/ -run 'TestContractModeSSOTSections' -count=1 -v
```

기대: `--- PASS`. 존재 검사이며, 문장이 옳게 지시하는지는 plan-auditor·sync-auditor 검토가 판정한다.

### AC-GR-006 — 생명주기 순서 (REQ-GR-019)

- **Given** `run.md`·`sync.md` 의 contract 블록과 SSOT 의 「One-pass lifecycle」 절
- **When** 각 범위 **안에서만** 단계명의 첫 등장 순서를 뽑는다
- **Then** run 블록은 `Discovery RED GREEN Qualification`, sync 블록은 `Closure Integration Push`, SSOT 생명주기 절은 일곱 단계 전부가 이 순서다

```bash
go test ./internal/template/ -run 'TestContractModeLifecycleOrder' -count=1 -v
```

기대: `--- PASS`.

### AC-GR-007 — 단계별 증거 의무 (REQ-GR-020)

- **Given** SSOT 의 「One-pass lifecycle」 절의 단계 표
- **When** 각 단계 행의 증거 열을 읽는다
- **Then** Discovery 행에 `reobserve`, RED 행에 `commit` 과 `failing`, GREEN 행에 `passing`, Qualification 행에 `lint`·`coverage`·`mutation`·`audit_multi`, Closure 행에 `verdict.md`, Integration 행에 `merged tree` 가 있고, 모든 행의 진행 조건 열에 `open escalation record` 가 있다

```bash
go test ./internal/template/ -run 'TestContractModeLifecycleEvidence' -count=1 -v
```

기대: `--- PASS`.

### AC-GR-008 — 블록 가드 (REQ-GR-001·023)

- **Given** `internal/template/contract_mode_blocks_test.go`
- **When** 불량 픽스처(짝 불일치 / evolvable 구간 안의 블록 / SPEC ID·카드 id·날짜를 담은 블록 / 문자 상한 초과 블록 / 검사 대상 블록 0개)로 먼저 실행하고, 이어 실제 템플릿 트리로 실행한다
- **Then** 다섯 불량 픽스처 모두 실패가 관측되고, 실제 트리에서는 통과한다

```bash
go test ./internal/template/ -run 'TestContractModeBlocksWellFormed' -count=1 -v
```

기대: `--- PASS`. 불량 픽스처의 `--- FAIL` 원문은 RED 증거로 `progress.md §E.2` 에 남긴다.

### AC-GR-009 — 템플릿 중립성 (REQ-GR-023)

- **Given** 템플릿 트리의 모든 블록, 템플릿 SSOT, 템플릿 `moai-mcp-tools*.md`·`workflow.yaml` 의 개정 문장
- **When** 금지 클래스를 찾고 기존 중립성 테스트를 실행한다
- **Then** 적중 0, 테스트 통과

```bash
go test ./internal/template/ -run 'TestContractModeBlocksWellFormed|TestTemplateNeutrality|TestTemplateNoInternalContentLeak' -count=1 -v
```

기대: `TestContractModeBlocksWellFormed`·`TestTemplateNeutralityAudit`·`TestTemplateNeutralityAuditC8Preserve`·`TestTemplateNoInternalContentLeak` 각각 `--- PASS`. 빈 선택은 통과가 아니다.

선택 확인(실행 없이 목록만, 트리 `1b071a573`, `go test -list '<위 정규식 그대로>' ./internal/template/`, exit 0) — 이미 있는 테스트 3개가 선택되고 `TestContractModeBlocksWellFormed` 는 아직 없다(run 에서 RED):

```text
TestTemplateNoInternalContentLeak
TestTemplateNeutralityAudit
TestTemplateNeutralityAuditC8Preserve
ok  	github.com/modu-ai/moai-adk/internal/template	0.382s
```

### AC-GR-010 — guided 문맥 증가 상한 (REQ-GR-002)

- **Given** always-loaded 편집 파일 4개(`CLAUDE.md`, `askuser-protocol.md`, `goal-directive.md`, `moai-mcp-tools.md`)와 스킬 파일 블록
- **When** `BASE` 대비 문자 수 증가(`LC_ALL=en_US.UTF-8` 문자 단위)와 블록별 길이를 잰다
- **Then** always-loaded 블록은 블록당 500자·세 파일 합계 1,500자 이하, `moai-mcp-tools.md` 증가 300자 이하, 스킬 파일 블록은 블록당 900자 이하, 각 always-loaded 파일 40,000자 미만

```bash
MOAI_GR_BASE=<BASE> go test ./internal/template/ -run 'TestContractModeAlwaysLoadedBudget' -count=1 -v
```

기대: `--- PASS`, 테스트가 파일별 증가량을 `t.Log` 로 출력한다.

### AC-GR-011 — 블록 첫 문장이 적용 조건 (REQ-GR-003)

- **Given** 모든 블록(로컬·템플릿)
- **When** 시작 마커 다음의 첫 비어 있지 않은 줄을 본다
- **Then** 모두 `` Where `workflow.autonomy.mode: contract` `` 로 시작한다

```bash
go test ./internal/template/ -run 'TestContractModeBlockCondition' -count=1 -v
```

기대: `--- PASS`. 설정 키는 A1 0.5.2 (25283ebf8) REQ-CONTRACT-015.

### AC-GR-012 — plan-audit FAIL 자동 수리 상한 (REQ-GR-015)

- **Given** `spec-assembly.md` 의 `contract-audit-retry`·`contract-quality-gate` 블록
- **When** 블록 내용을 본다
- **Then** 두 블록 모두 `audit_retries` 와 `budget_default` 를 담고, `AskUserQuestion` 을 담지 않는다

```bash
go test ./internal/template/ -run 'TestContractModeAuditRetryBlocks' -count=1 -v
```

기대: `--- PASS`. 키 이름은 A1 0.5.2 (25283ebf8) § Configuration.

### AC-GR-013 — sync 확인 질문 제거와 실패 결정점 (REQ-GR-016)

- **Given** `sync.md`·`sync/doc-execution.md`·`sync/delivery.md` 의 블록
- **When** 블록 내용을 본다
- **Then** 블록들이 `gate-sync-2`, Phase 1·3·6·7·8·13 을 이름으로 가리키고, `escalat` 경로를 명시하며, `AskUserQuestion` 과 `ci-autofix` 루프 변경을 담지 않는다

```bash
go test ./internal/template/ -run 'TestContractModeSyncBlocks' -count=1 -v
```

기대: `--- PASS`.

### AC-GR-014 — 서명 게이트와 초안 지시 문언 (REQ-GR-004·006)

- **Given** `run.md` 의 `contract-signing-run`, `spec-assembly.md` 의 `contract-signing-review`·`contract-draft` 블록
- **When** 명령·파일 이름을 찾는다
- **Then** run 블록에 `moai contract kickoff-check` 와 「decide 결과 `reject`·`human` 은 둘 다 사람 서명 절차」 문장(토큰 `reject`·`human` 이 같은 문장에 있음), 서명 검토 블록에 `moai contract sign`, 초안 블록에 `contract.yaml` 이 있다

```bash
go test ./internal/template/ -run 'TestContractModeSigningBlocks' -count=1 -v
```

기대: `--- PASS`. 명령 이름은 A1 0.5.2 (25283ebf8) REQ-CONTRACT-010 과 이 SPEC 의 kickoff-check.

### AC-GR-015 — 빌드와 범위 테스트 (전 REQ 공통)

- **Given** 편집이 끝난 트리
- **When** 빌드와 변경 패키지 테스트를 실행한다
- **Then** 모두 성공

```bash
make build
go test ./internal/template/... ./internal/contract/... -count=1
go test ./internal/cli/ -run 'TestContract(Decide|KickoffCheck|Revoke)' -count=1 -v
```

기대: `make build` exit 0(선행 `agents-emit-check`·`commands-emit-check` 포함), 둘째 명령 `ok`, 셋째 명령에 세 테스트 `--- PASS`. 전체 스위트는 CI 가 판정한다.

### AC-GR-016 — kickoff-check 판정 (REQ-GR-004·007)

- **Given** 임시 저장소 픽스처 16종: (1) 사람 서명 + `events.jsonl` `sign-human` 사건, (2) 에이전트가 쓴 영수증 파일로 한 `llm` 서명(`receipts.jsonl` 에 없음), (3) `receipts.jsonl` 의 `outcome: approve` 영수증(`effective_decider: llm`)과 일치하는 `llm` 서명 + 활성, (4) 같은 조건의 `llm+jev` 영수증(R1)과 `llm+jev` 서명 + 활성, (5) 대체 영수증(`requested_decider: llm+jev` → `effective_decider: llm`)과 일치하는 `llm` 서명 + 활성, (6) `signer_kind: jev` 서명, (7) (3) + 비활성, (8) (3) + 그 뒤 `revoke` 사건, (9) `events.jsonl` 에 서명 사건이 없는 사람 서명, (10) 설정 결정자가 `llm` 인데 영수증의 `requested_decider` 가 `llm+jev`, (11) `--card` 가 계약 `card` 필드와 다름, (12) 저장소의 `outcome: reject` 영수증과 해시가 일치하도록 위조한 `llm` 서명 + 그 서명과 일치하는 체인이 맞는 위조 `sign-receipt` 사건 줄 + 활성, (13) 저장소의 `outcome: human` 영수증과 일치하도록 위조한 `llm+jev` 서명 + 체인이 맞는 위조 `sign-receipt` 사건 줄 + 활성, (14) 봉인 불일치 서명(`signed-invalid`), (15) 미서명 계약(`unsigned`), (16) `effective_decider: llm` 영수증과 `signer_kind: llm+jev` 서명 + 활성
- **When** `moai contract kickoff-check <SPEC-ID> --card <card> --json` 을 실행한다(활성 여부는 판정 함수의 인자로 주입)
- **Then** (1)·(3)·(4)·(5) exit 0; (2) exit 1 `receipt-not-issued`; (6) exit 1 `decider-not-permitted`; (7) exit 1 `autonomous-kickoff-inactive`; (8) exit 1 `revoked`; (9) exit 1 `signature-not-recorded`; (10)·(16) exit 1 `decider-mismatch`; (11) exit 1 `card-mismatch`; (12)·(13) exit 1 `receipt-not-approved` — `reject` 와 `human` 이 같은 사유로 거절돼 둘 다 사람 결정으로 간다; (14)·(15) exit 1 `not-signed-valid` 이고 `--json` 에 A1 verify 상태(`signed-invalid`·`unsigned`)가 있다. 기대 사유는 모두 REQ-GR-007 의 우선순위로 정한 대표 사유이고, `--json` 의 사유 목록은 성립하는 사유를 모두 담는다 — (6) 의 목록에는 `decider-not-permitted` 와 함께 `not-signed-valid` 도 있고(A1 일관성 표상 `method: receipt` 에 `signer_kind: jev` 는 `signature_inconsistent`), (12)(13) 은 위조 사건 줄 덕분에 `signature-not-recorded` 가 성립하지 않아 `receipt-not-approved` 가 사유를 가른다. `--json` 에는 늘 `autonomous_kickoff_enabled`·`jev_doctrine_amended` 가 있다. 어느 경우에도 파일이 생기거나 바뀌지 않는다

```bash
go test ./internal/contract/kickoff/ -run 'TestKickoffCheck' -count=1 -v
```

기대: 하위 테스트 16개 `--- PASS`. 아직 없는 테스트다(`internal/contract` 패키지 부재, EV-8) — run 에서 RED 를 먼저 관측한다. 서명 필드와 `signer_kind` 값 집합(`human | llm | llm+jev`), `card` 필드는 A1 0.5.2 (25283ebf8) design § Contract Schema·§ Card Field.

### AC-GR-017 — 활성 순서, 원칙 개정 연동, A1 임시 규칙의 대체 테스트 (REQ-GR-008·013·025)

- **Given** 이 브랜치의 이력, 상수 `autonomousKickoffEnabled`·`jevDoctrineAmended`, REQ-GR-013 의 개정 위치 전부(`CLAUDE.local.md §29` 포함), A1 서명기 단계 (1)(임시 규칙, 상수로 조건화됨)
- **When** 순서 테스트, 연동 테스트, 서명기 대체 테스트를 실행한다
- **Then** 순서: `autonomousKickoffEnabled = false` 이면 그 상태를 `t.Log` 로 출력하고 `llm`·`llm+jev` 서명이 거절됨을 확인해 통과하고(공허 통과 아님), `true` 이면 그 값을 넣은 커밋이 (i) `internal/contract/revoke` 를 처음 추가한 커밋과 (ii) `internal/contract/receipt` 를 처음 추가한 커밋의 **엄격한 후손**(같은 커밋 아님)이어야 통과한다. 연동: 트리에서 개정 표지 전부와 `jevDoctrineAmended = true` 가 모두 있거나 모두 없어야 하고, 이력에서 둘이 **처음 나타나는 커밋이 같아야** 한다. 서명기 대체 테스트(A1 AC-CONTRACT-016 [REF] (t) 의 대체, 매 head 에서 doctrine 플래그의 거짓·참을 모두 주입): 거짓 주입이면 `effective_decider: llm+jev`·두 답 approve·`outcome: approve` 영수증이 `receipt_requires_human` 으로 거절되고, 참 주입이면 같은 영수증은 서명되며 `outcome: reject`·`human` 인 `llm+jev` 영수증은 각각 `receipt_rejected`·`receipt_requires_human` 이다. 픽스처 저장소 다섯 개 — 상수를 조건 커밋과 같은 커밋에 둔 것, 개정 표지만 있는 커밋, `jevDoctrineAmended = true` 만 있는 커밋, `CLAUDE.local.md §29` 만 빠진 커밋, 임시 규칙 단계가 삭제된 head(상수 참) — 에서 FAIL 을 관측한다(마지막은 그 head 에서 서명기 대체 테스트의 거짓 주입 경우가 「거짓인데 서명됨」으로 FAIL)

```bash
go test ./internal/contract/kickoff/ ./internal/contract/sign/ -run 'TestAutonomousKickoffActivationOrder|TestJevAmendmentLinkage|TestSignInterimRuleFollowsDoctrine' -count=1 -v
```

기대: 세 테스트 `--- PASS`. 모두 아직 없는 테스트다(run 에서 RED). A1 영수증 경로·A2b 가드·Frozen 문단 개정(`design.md §7.1` 3·4·6행)은 이 테스트가 보지 않는다 — 그 셋은 M8 착수 전 리드가 develop 에서 확인하고 `progress.md` 에 근거를 적는다. 임시 규칙과 대체 테스트 소유는 A1 0.5.2 (25283ebf8) spec §C.8, REQ-CONTRACT-024, acceptance AC-CONTRACT-016 [REF] (t)(「A3 owns the replacing test」).

### AC-GR-018 — decide 전제조건과 판정의 산출물 결합 (REQ-GR-009·011)

- **Given** (a)~(f) 를 하나씩 깬 픽스처 6종, (a) 의 변형 5종 — `Overall Score:` 줄에 괄호 주석, `plan_artifact_hash:` 줄 없음, `plan_artifact_hash:` 가 있으나 보고서 뒤 `spec.md` 를 고쳐 현재 해시와 다름, `Verdict: PASS-WITH-DEBT` + 점수가 Tier 문턱 미만, `Verdict: PASS-WITH-DEBT` + 점수가 문턱 이상 — 모두 성립하는 픽스처, Jev 생성 횟수를 세는 이음매, `t.TempDir()` 를 가리키는 `MOAI_HOME`
- **When** decide 를 실행한다
- **Then** 깬 10종(6 + 앞의 네 변형)은 `outcome: human`·`reason: precondition:<x>`·Jev 생성 0회이고 **`kickoff-receipt.json` 이 생기지 않으며 `receipts.jsonl` 줄 수가 그대로이고 `events.jsonl` 만 한 줄 는다**; `PASS-WITH-DEBT` + 문턱 이상과 성립 픽스처(결정자 `llm+jev`, `llm` approve)는 Jev 를 1회 생성한다. (a) 는 `.moai/reports/<card>/` 의 가장 높은 N 파일을 고르고 그 경로·해시를 영수증 `inputs.plan_audit_report` 에 기록한다

```bash
go test ./internal/contract/kickoff/ -run 'TestDecidePreconditions' -count=1 -v
```

기대: 하위 테스트 12개 `--- PASS`. 아직 없는 테스트다(run 에서 RED). 영수증 형식·`frozen-files` 정의는 A1 0.5.2 (25283ebf8); `plan_artifact_hash:` 판독과 해시는 `internal/runtime`(`parsePlanAuditSnapshot`, `ComputeHash`) 재사용; 열린 에스컬레이션 판독은 **[A2 개정본으로 재확인]**.

### AC-GR-019 — 결과 도출 규칙 R1·R3·R4·R5 와 작성자 배제 (REQ-GR-010)

- **Given** 규칙·갈래마다 하나씩인 테스트: R1 모두 승인(`llm` approve + Jev approve), R1 모두 거절(`llm` reject + Jev reject), R1 불일치(`llm` approve + Jev reject·escalate, `llm` reject + Jev approve, `llm` escalate + Jev approve, `llm` reject + Jev escalate) — 셋 모두 `jevDoctrineAmended = true`; R3(`llm` approve + Jev approve, `jevDoctrineAmended = false`); R4(설정 결정자 `llm`, `llm` approve·reject·escalate); R5(설정 결정자 `jev`); 작성자 배제(판단 `agent: manager-spec` / 판단 `agent` 가 SPEC 커밋의 `Authored-By-Agent:` 값과 같음 — 그 커밋 본문은 트레일러 뒤에 빈 줄과 서명 줄이 붙어 git 트레일러 파서로는 빈 값 / SPEC 커밋 어디에도 트레일러 없음). Jev 는 스텁 이음매이고 생성 횟수를 센다
- **When** decide 를 실행한다
- **Then** R1 모두 승인 → `outcome: approve`; R1 모두 거절 → `reject`; R1 불일치 → `human`·`cross-check-disagree`; R3 → `human`·`jev-doctrine-not-amended`; R4 → `approve`·`reject`·`human` 이고 Jev 생성 0회; R5 → exit 2·`decider-jev-refused`·저장소 줄 수 불변·Jev 생성 0회; 작성자 배제 → 앞의 둘은 `human`·`author-decider-conflict`, 셋째는 `human`·`author-check-unmeasured` 이고 세 경우 모두 영수증 파일이 생기지 않는다. 영수증의 `outcome`·`requested_decider`·`effective_decider`·`fallback.applied: false` 가 결과와 일치하고, `events.jsonl` 의 `decide` 사건에 적용한 규칙, 신고 신원, 판독한 트레일러 집합이 함께 있다

```bash
go test ./internal/contract/kickoff/ -run 'TestDecideRuleCrossCheckAgree|TestDecideRuleCrossCheckBothReject|TestDecideRuleCrossCheckDisagree|TestDecideRuleJevBeforeAmendment|TestDecideRuleLLMAlone|TestDecideRuleJevAloneRefused|TestDecideAuthorExclusion' -count=1 -v
```

기대: 일곱 테스트 각각 `--- PASS`. RED 는 decide 구현 전 일곱 테스트 각각의 `--- FAIL` 원문이다 — 규칙·갈래마다 따로 관측한다. 규칙은 이 SPEC 이 소유하며 A1 validator 를 대조 기준으로 쓰지 않는다(A1 은 기록된 `outcome` 의 일관성만 본다 — A1 0.5.2 (25283ebf8) design § Kickoff Receipt 필드 규칙 9).

### AC-GR-020 — decide 의 이름·입력·출력 위치와 부작용 (REQ-GR-011)

- **Given** 워크트리·브랜치·원격 ref·`backlog.db`·`contract.yaml`(`card: <card>`)·SPEC 문서가 있는 임시 git 저장소와 `t.TempDir()` 를 가리키는 `MOAI_HOME`
- **When** `moai contract decide --help` 를 읽고, 정상 입력 / 설정 결정자 `human` / `<card>` 가 계약 `card` 와 다른 입력 / 형식 오류 판단 파일 / 변조된 저장소로 `moai contract decide <card> --spec <SPEC-ID> --judgement <file>` 을 실행한다
- **Then** 도움말에 동사 `decide`, 위치 인자 `<card>`, `--spec`·`--judgement`·`--json` 이 있다. exit 0 / 0 / 2(`card-mismatch`) / 2 / 1. 설정 결정자 `human` 에서는 `events.jsonl` 만 한 줄 늘고 `receipts.jsonl`·`kickoff-receipt.json` 은 생기지 않는다(A1 필드 규칙 2·6 을 어기는 영수증을 쓰지 않음). 정상 입력에서 `MOAI_HOME/db/<project-key>/contract/receipts.jsonl` 과 `events.jsonl` 이 각각 정확히 한 줄 늘고, `.moai/specs/<SPEC-ID>/kickoff-receipt.json` 이 새 영수증 줄의 본문과 바이트 동일하며 A1 영수증 필드 밖의 키(사유 코드 등)를 담지 않는다. exit 1·2 에서 두 파일의 줄 수 불변. 모든 경우 `backlog.db` 바이트, `git for-each-ref` 출력, 워크트리 목록, `contract.yaml`, SPEC 문서가 전후 동일하고, 작업 트리의 새 파일은 정상 입력의 `kickoff-receipt.json` 하나뿐이며, LLM 호출 이음매 기록이 0이다

```bash
go test ./internal/cli/ -run 'TestContractDecide' -count=1 -v
```

기대: `--- PASS`. 동사 이름은 가칭이며 이름이 바뀌면 이 AC 의 명령과 도움말 검사도 함께 바뀐다. 영수증 고정 경로와 `card` 필드는 A1 0.5.2 (25283ebf8) design § Kickoff Receipt·§ Card Field, 저장소 디렉터리는 리드 결정 R10.

### AC-GR-021 — 저장소: 모든 서명 사건과 변조 흔적 (REQ-GR-012)

- **Given** `t.TempDir()` 를 가리키는 `MOAI_HOME` 과 사건 종류 다섯(`sign-human`, `sign-receipt`, `reseal`, `decide`, `revoke`)을 각각 일으키는 픽스처
- **When** 각 사건을 일으키고, (i) 두 파일의 체인을 검증하고, (ii) `events.jsonl` 가운데 줄 한 바이트를 바꿔 검증하고, (iii) 마지막 `sign-human` 줄을 지우고 kickoff-check 를 실행하고, (iv) 마지막 `revoke` 줄을 지우고 decide 를 다시 실행하고, (v) 저장소 추가를 실패시키는 이음매로 사람 서명을 시도하고, (vi) 규칙 R2 대체로 decide 를 실행한다
- **Then** 사건마다 `events.jsonl` 에 정확히 한 줄이 추가되고(`decide` 는 `receipts.jsonl` 에도 한 줄); (i) 통과; (ii) 깨진 줄 번호와 함께 실패; (iii) `signature-not-recorded`; (iv) revoke 가 쓴 에스컬레이션 기록이 남아 전제조건 (e)로 `outcome: human`; (v) 서명 파일이 쓰이지 않는다; (vi) 그 영수증 줄에 `requested_decider: llm+jev`·`effective_decider: llm`·`fallback{applied: true, reason}` 이 있다. 두 파일은 `MOAI_HOME` 아래 `db/<project-key>/contract/`(리드 결정 R10)에 있고 작업 트리 밖이다

```bash
go test ./internal/contract/receipt/ ./internal/contract/sign/ -run 'TestEventStore|TestSignRecordsEvent' -count=1 -v
```

기대: `--- PASS`. 이 AC 는 변조 **흔적**만 판정한다 — 저장소와 에스컬레이션 기록을 함께 지우는 행위는 막지 못하며 판정하지 않는다. 서명기 개정 지점은 A1 0.5.2 §C.6 이 A3 에 넘긴 요구다.

### AC-GR-022 — Jev 원칙 개정과 자기모순 부재 (REQ-GR-013)

- **Given** REQ-GR-013 의 개정 위치 전부와, 불량 픽스처 두 개 — REQ-JEVC-012 만 개정되고 `### Out of Scope — authority` 두 항목은 옛 문장 그대로인 `SPEC-JEV-CORE-001` 사본, REQ-JEVC-011 의 표시 전용 문장이 예외 없이 남은 사본
- **When** 불량 픽스처로 먼저, 이어 실제 트리로 개정 문장과 표지를 확인한다
- **Then** 두 불량 픽스처는 FAIL 이 관측된다. 실제 트리에서는 `SPEC-JEV-CORE-001/spec.md` 의 REQ-JEVC-011·REQ-JEVC-012 에 `[AMENDED 2026-09-26` 표지가 있고, `Out of Scope — authority` 의 두 항목이 모두 「contract 모드 Kickoff 의 `llm+jev` 교차 확인에서 두 번째 신호로 쓰일 때」 예외를 명시하며(예외 없는 게이트 금지 문장이 남아 있지 않음), HISTORY 에 운영자 결정 「Kickoff 는 LLM·Jev 도 할 수 있게」와 2026-09-26 재결정(Jev 단독 없음, 교차 확인만)을 인용한 행이 있고 `status: completed` 가 유지되며; `moai-mcp-tools.md`·`moai-mcp-tools-catalogue.md`(로컬·템플릿)의 `jev_ask` 행과 `workflow.yaml`(로컬·템플릿) `jev:` 주석이 같은 예외 한 곳만 명시하고 Jev 단독 결정과 다른 금지 대상(완료 판정·병합 승인·큐 변경)을 그대로 담으며; `CLAUDE.local.md §29` 에 `design.md §11.1` 문안이 있고 운영자 확인 기록이 `progress.md` 에 있다

```bash
go test ./internal/template/ -run 'TestJevDoctrineAmendment' -count=1 -v
moai spec lint SPEC-JEV-CORE-001
```

기대: 테스트 `--- PASS`(불량 픽스처의 `--- FAIL` 원문은 RED 증거로 남긴다), lint exit 0. `TestJevDoctrineAmendment` 는 아직 없다(`go test -list 'TestContractMode|TestJevDoctrineAmendment' ./internal/template/` → 선택 0, `ok` — 아래 §B.1).

### AC-GR-023 — revoke 동작과 멱등 (REQ-GR-022·018)

- **Given** 사람 서명 계약 / `llm` 영수증 서명 계약 / 이미 revoke 된 계약 / 미서명 계약 / 변조된 저장소 / `<card>` 가 계약 `card` 필드와 다른 호출
- **When** `moai contract revoke <card> --spec <SPEC-ID>` 를 실행한다
- **Then** exit 0(저장소 +1, 에스컬레이션 기록 정확히 1건, 종류 `revoke`) / 0(같음) / 0(쓰기 0) / 1(쓰기 0) / 2(쓰기 0) / 2(쓰기 0). 테스트는 A2 가 제공하는 판독 함수(또는 형식 검증기)로 그 기록을 읽어 열린 기록 1건으로 판정한다

```bash
go test ./internal/contract/revoke/ ./internal/cli/ -run 'TestRevoke|TestContractRevoke' -count=1 -v
```

기대: `--- PASS`. 기록 경로·YAML 머리·`revoke` 종류는 **[A2 개정본으로 재확인]**.

### AC-GR-024 — revoke 가 하지 않는 일 (REQ-GR-022)

- **Given** 워크트리 1개·브랜치 2개·원격 ref·`backlog.db` 가 있는 임시 git 저장소와 git 실행 이음매
- **When** revoke 를 실행한다
- **Then** `git worktree list`·`git for-each-ref`·원격 ref·`backlog.db`·`contract.yaml`·SPEC 문서가 전후 바이트 동일하고, git 실행 이음매에 기록된 `push`·`branch -d`·`branch -m`·`worktree remove` 호출이 0회이며, 프로세스 종료 호출이 0회다

```bash
go test ./internal/contract/revoke/ -run 'TestRevokeLeavesRepositoryUntouched' -count=1 -v
```

기대: `--- PASS`.

### AC-GR-025 — 규칙 R2: Jev 쪽 실패의 `llm` 단독 대체와 기록 (REQ-GR-010·012)

- **Given** 설정 결정자 `llm+jev`, `jevDoctrineAmended = true`, `t.TempDir()` 를 가리키는 `MOAI_HOME`, 그리고 Jev 실패 원인별 픽스처 5종 — 도달 불가(호출 실패), 자격 증명 없음, `workflow.jev.enabled: false`, 형식이 깨진 응답, 신뢰도 0.49(`jev_min_confidence` 0.50) — 각각을 `llm` approve 판단과 `llm` reject 판단으로
- **When** decide 를 실행한다
- **Then** 열 경우 모두 결과가 `llm` 단독 R4 로 도출되어(approve → `approve`, reject → `reject`) 두 번째 신호 없이 결정되고, 영수증 파일과 `receipts.jsonl` 줄 양쪽에 `requested_decider: llm+jev`·`effective_decider: llm`·`fallback.applied: true`·원인에 맞는 `fallback.reason`(`jev_call_failed`·`jev_key_missing`·`jev_disabled`·`jev_malformed_response`·`jev_low_confidence`)이 있고 `jev_answer` 가 없다. 신뢰도 0.49 경우의 원시 응답은 `receipts.jsonl` 줄에 남는다

```bash
go test ./internal/contract/kickoff/ -run 'TestDecideJevFallback' -count=1 -v
```

기대: 하위 테스트 10개(원인 5 × `llm` approve·reject) `--- PASS`. RED 는 원인별 하위 테스트 각각의 `--- FAIL` 원문. 대체 사유의 닫힌 집합과 영수증 필드는 A1 0.5.2 (25283ebf8) design § Kickoff Receipt 필드 규칙 4·5·6.

## §B. 증거 원장 — RED-now 셀

문서 기준 핀: `ca1d5dc43`(EV-1~6), `ca1e39f6f`(EV-7·8). M0 에서 `BASE` 가 정해지면 같은 명령을 `BASE` 에서 다시 재고 `progress.md §E.2` 에 갱신한다.

| id | 대상 AC | 명령 (단일 호출) | stdout | exit | 빨간 이유 |
|---|---|---|---|---|---|
| EV-1 | 004, 011 | `grep -rlc 'moai:contract-mode-start' CLAUDE.md .claude internal/template/templates` | (빈 출력) | 1 | 블록이 아직 없다 |
| EV-2 | 002, 005, 008 | `ls .claude/rules/moai/workflow/contract-autonomy.md internal/template/templates/.claude/rules/moai/workflow/contract-autonomy.md internal/template/contract_mode_blocks_test.go` | `ls: … No such file or directory` ×3 (stderr) | 1 | SSOT 와 가드 테스트가 없다 |
| EV-3 | 006 | `grep -c 'Discovery' .claude/skills/moai/workflows/run.md` | `0` | 1 | 생명주기 서술이 없다 |
| EV-4 | 006, 007 | `grep -c 'Qualification' .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/sync.md` | `…/run.md:0` `…/sync.md:0` | 1 | 같은 이유 |
| EV-5 | 001 (판정식 반증) | 스크래치 파일에 걷어내기 적용 후 `cmp` — 복원 사례와 블록 밖 한 단어 변이 사례 | `STRIP-RESTORES`; `mutant cmp exit=1` | 0 / 1 | AC-GR-001 은 불변 가드라 RED-now 대신 변이로 반증 능력을 확인했다 |
| EV-6 | 003 | `moai constitution validate` | `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)` | 0 | 불변 가드 기준선 |
| EV-7 | 015, 016, 020, 023 | `moai contract revoke --help` | `ERROR Unknown command "contract" for "moai". Try --help for usage.` (공백 정리) | 1 | `contract` 명령 자체가 없다 — A1 이 만들고 이 SPEC 이 하위 명령을 더한다 |
| EV-8 | 016~021, 023~025 | `ls internal/contract` | `ls: internal/contract: No such file or directory` (stderr) | 1 | 패키지가 없다. 녹색 경로는 A1 병합 후 M6·M7 |

AC-GR-001·002·003·009·010 은 불변 가드 또는 상한 검사라 녹색이 기본이다 — 반증 능력은 각 Go 테스트의 불량 픽스처로 run phase 에서 관측한다(AC-GR-008 과 같은 방식).

### §B.1 테스트 선택 확인 (실행 없이 목록만, 트리 `1b071a573`)

Go 테스트를 이름으로 고르는 AC 마다 `go test -list '<AC 의 정규식>' <패키지>` 로 선택 집합을 쟀다. 이미 있는 테스트는 AC-GR-009 의 세 개뿐이다.

| AC | 정규식 · 패키지 | 관측(원문) | 판정 |
|---|---|---|---|
| 009 | `TestContractModeBlocksWellFormed\|TestTemplateNeutrality\|TestTemplateNoInternalContentLeak` · `./internal/template/` | `TestTemplateNoInternalContentLeak` / `TestTemplateNeutralityAudit` / `TestTemplateNeutralityAuditC8Preserve` / `ok  github.com/modu-ai/moai-adk/internal/template 0.382s`, exit 0 | 기존 3개 선택. `TestContractModeBlocksWellFormed` 는 아직 없음 — RED at run |
| 001·002·003·004·006·007·008·010·011·012·013·014·022 | `TestContractMode…`·`TestJevDoctrineAmendment` · `./internal/template/` | `go test -list 'TestContractMode\|TestJevDoctrineAmendment' ./internal/template/` → 목록 없이 `ok  github.com/modu-ai/moai-adk/internal/template 0.199s`, exit 0 | 아직 없음 — RED at run |
| 015·020·023 | `TestContract(Decide\|KickoffCheck\|Revoke)` · `./internal/cli/` | `go test -list 'TestContract(Decide\|KickoffCheck\|Revoke)' ./internal/cli/` → 목록 없이 `ok  github.com/modu-ai/moai-adk/internal/cli 0.969s`, exit 0 | 아직 없음 — RED at run |
| 016·017·018·019·021·023·024·025 | `./internal/contract/...` 의 테스트 | `ls internal/contract` → `ls: internal/contract: No such file or directory`, exit 1 | 패키지 자체가 없음 — RED at run |

run 단계에서 각 테스트가 생긴 뒤 같은 `-list` 로 정규식이 의도한 이름을 모두 고르는지 다시 잰다 — 빈 선택은 통과가 아니다.

## §C. 품질 게이트와 완료 정의

### §C.1 MUST-PASS

AC-GR-001, 002, 003, 004, 006, 008, 009, 015, 016, 017, 018, 019, 020, 021, 023, 024, 025.

### §C.2 추적성

| REQ | AC |
|---|---|
| REQ-GR-001 | AC-GR-001, AC-GR-008 |
| REQ-GR-002 | AC-GR-001, AC-GR-010 |
| REQ-GR-003 | AC-GR-011 |
| REQ-GR-004 | AC-GR-014, AC-GR-016 |
| REQ-GR-005 | AC-GR-004, AC-GR-005 |
| REQ-GR-006 | AC-GR-014 |
| REQ-GR-007 | AC-GR-016 |
| REQ-GR-008 | AC-GR-017, AC-GR-005 |
| REQ-GR-009 | AC-GR-018 |
| REQ-GR-010 | AC-GR-019, AC-GR-025, AC-GR-005 |
| REQ-GR-011 | AC-GR-020 |
| REQ-GR-012 | AC-GR-021, AC-GR-025 |
| REQ-GR-013 | AC-GR-022, AC-GR-017 |
| REQ-GR-014 | AC-GR-005 (+ 검토) |
| REQ-GR-015 | AC-GR-012 |
| REQ-GR-016 | AC-GR-013, AC-GR-005 |
| REQ-GR-017 | AC-GR-003, AC-GR-005 |
| REQ-GR-018 | AC-GR-023, AC-GR-005 |
| REQ-GR-019 | AC-GR-006 |
| REQ-GR-020 | AC-GR-007 |
| REQ-GR-021 | AC-GR-005 (+ 검토) |
| REQ-GR-022 | AC-GR-023, AC-GR-024 |
| REQ-GR-023 | AC-GR-002, AC-GR-008, AC-GR-009 |
| REQ-GR-024 | AC-GR-003, AC-GR-010 |
| REQ-GR-025 | AC-GR-017 |

요구사항 25개 전부가 하나 이상의 AC 에 매핑된다.

### §C.3 기계로 판정하지 않는 것

- 블록 문장이 정확히 무엇을 지시하는지는 절 단위 토큰 존재로만 부분 판정한다(REQ-GR-014·021). 의미는 plan-auditor·sync-auditor 가 판정한다.
- guided 세션의 모델이 contract 블록을 무시한다는 것 — 텍스트 보존과 문자 상한만 판정한다.
- `design.md §7.1` 3·4·6행(A1 영수증 경로, A2b 두 가드, Frozen 문단 개정)의 착지 — M8 착수 전 리드가 확인한다(AC-GR-017 기대 참조).

### §C.4 완료 정의

- MUST-PASS 전부 PASS, 나머지 AC PASS 또는 사유가 적힌 PASS-WITH-DEBT.
- `plan.md §B` 의 기본값 두 건(D-2·D-4)에 운영자의 변경 지시가 없거나, 있으면 반영됨.
- **[A2 개정본으로 재확인]**·**[A2b SPEC 으로 재확인]** 항목과 A1 0.5.2 인용이 M0 에서 각 병합본과 대조됨.
- `progress.md §E.2` 에 `BASE` SHA, 재측정 원장, AC 표가 원문 증거와 함께 있음.
- AC 수를 제자리에서 바꾸는 개정은 AC 스냅숏 재생성을 같은 커밋에 싣는다(`plan.md §D`).
