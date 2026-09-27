# design.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.7)

A1 기준: 0.5.2 (`25283ebf8`) — 스키마(결정자 값 집합, 파생 기본값, 영수증 필드, 구조 검증, `card` 필드)는 A1, 결정 규칙은 A3. A2 기준: 0.4.3 (`status: implemented`, BASE `7fe658815`) — 기록 경로·YAML 머리·`revoke` 종류를 M0 에서 대조했다. A2 는 revoke 기록을 `status: resolved` 로 두고 needs-decision 으로 세지 않으므로 재개 차단은 A3 revoke 판독기가 맡는다(§10). A2b(t1245): push 직렬화·에이전트발 `sign` 차단·팩토리 에이전트 세션의 `decide` 거절(리드 결정 2026-09-26).

## §1. 설계의 뼈대

1. **블록 밖 텍스트 보존.** 기존 문장을 고치지 않고 contract 모드 지시를 마커 블록으로만 덧붙인다. 블록을 걷어내면 기준 파일과 바이트 동일하다(REQ-GR-002). 이것이 증명하는 것은 **블록 밖 텍스트의 불변**이지 guided 세션 문맥의 불변이 아니다 — guided 세션도 블록을 읽는다. 그 증가량은 §4 에서 수치로 묶는다. 예외는 Jev 원칙 개정(REQ-GR-013) 한 건이며, 원칙 문장 자체를 바꾸는 개정이라 블록 방식이 아니다(§11).
2. **always-loaded 예산.** 본문은 path-scoped SSOT 하나에 두고, always-loaded 파일에는 조건문 + 포인터만 둔다.
3. **판정은 코드가 한다.** 자율 Kickoff 의 통과 여부를 문서가 아니라 `moai contract kickoff-check` 가 판정한다(D3). 문서는 그 명령을 부르라고 지시할 뿐이다.

### §1.1 마커 규약

```text
<!-- moai:contract-mode-start id="<slug>" -->
Where `workflow.autonomy.mode: contract` — <지시>. See `.claude/rules/moai/workflow/contract-autonomy.md` § <절>.

<!-- moai:contract-mode-end -->
```

- 두 마커는 각각 한 줄 전체를 차지한다. 시작 마커는 기존 빈 줄 바로 뒤에, 끝 마커 다음 줄은 원래 있던 줄이다. 블록 안의 빈 줄은 블록에 속한다. 한 줄에서 닫히는 HTML 주석은 CommonMark 에서 그 줄로 끝나므로 다음 줄의 제목·문단은 정상 렌더된다.
- `<slug>` 는 `[a-z0-9-]+`, 파일 안에서 유일하다. 자율 Kickoff 를 활성화하는 블록은 **전용 id `contract-autonomous-kickoff`** 만 쓴다(§7.1, AC-GR-017).
- 첫 비어 있지 않은 줄은 `` Where `workflow.autonomy.mode: contract` `` 로 시작한다(REQ-GR-003).
- 블록은 `moai:evolvable-start/end` 구간 밖에 둔다. `internal/merge/evolvable_zone.go:33` 정규식은 `moai:evolvable-start` 만 잡으므로 새 마커와 충돌하지 않지만, 구간 안에 두면 사용자 진화 구간으로 보존되어 `moai update` 가 블록을 갱신하지 않는다.

걷어내기와 추출의 정의(Go 가드가 같은 규칙을 구현한다):

- 걷어내기: 시작 마커 줄부터 끝 마커 줄까지(양 끝 포함)를 지운다.
- 추출: 같은 범위만 남긴다.

## §2. 편집 허용 목록 (정본 — REQ-GR-024)

문서 행은 로컬 경로이며 같은 경로를 `internal/template/templates/` 아래에도 같은 커밋에서 편집한다(표시가 있는 행 제외).

| # | 경로 | 블록 id / 변경 | 요구 | always-loaded |
|---|---|---|---|---|
| 1 | `.claude/rules/moai/workflow/contract-autonomy.md` (신규) | 전체 | 전부 — SSOT | 아니오 |
| 2 | `CLAUDE.md` | `contract-signing-pipeline`(§2 ④ 뒤), `contract-safe-dev`(§7 Rule 5 뒤) | 004·014 | **예** |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | `contract-ambiguity`(`## Ambiguity Triggers and Exceptions` 스텁 문단 뒤 — BASE `7fe658815` 기준 제목 206행·문단 208행·빈 줄 209행 다음, `## Free-form Circumvention Prohibition`(210행) 앞. 「The Five Exceptions」 본문은 t1175 가 `askuser-protocol-reference.md:229` 로 옮겼고 그 파일은 편집 대상이 아니다) | 014 | **예** |
| 4 | `.claude/rules/moai/workflow/goal-directive.md` | `contract-signing-goal`(§ Goal-Presentation Timing 끝) | 004 | **예** |
| 5 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | `contract-signing`(머리의 `[ZONE:Frozen] [HARD]` 문단 뒤) — **사람 서명 등가만** | 005 | 아니오 |
| 6 | `.claude/skills/moai/SKILL.md` | `contract-signing-router` | 004 | 아니오 |
| 7 | `.claude/skills/moai/workflows/moai.md` | `contract-pipeline-gates`, `contract-merged-round` | 004·015 | 아니오 |
| 8 | `.claude/skills/moai/workflows/plan.md` | `contract-clarification` | 004·009(b) | 아니오 |
| 9 | `.claude/skills/moai/workflows/plan/spec-assembly.md` | `contract-draft`(Phase 10), `contract-signing-review`(Step 2.3.3b), `contract-audit-retry`(Step 2.3.5), `contract-quality-gate`(Phase 15) | 004·006·015 | 아니오 |
| 10 | `.claude/skills/moai/workflows/run.md` | `contract-signing-run`, `contract-lifecycle-run` | 004·019·020 | 아니오 |
| 11 | `.claude/skills/moai/workflows/goal.md` | `contract-progression` | 004 (D-4) | 아니오 |
| 12 | `.claude/skills/moai/workflows/sync.md` | `contract-sync-gates` | 016·019·021 | 아니오 |
| 13 | `.claude/skills/moai/workflows/sync/doc-execution.md` | `contract-doc-scope`(`gate-sync-2` 구간 밖) | 016 | 아니오 |
| 14 | `.claude/skills/moai/workflows/sync/delivery.md` | `contract-next-steps`, `contract-error-flow` | 016·021 | 아니오 |
| 15 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | 도구 표의 `mcp__moai__jev_ask` 행 문장 개정(BASE `7fe658815` 기준 138행) | 013·025 | 아니오(`paths:` 조건부) |
| 16 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | 계열 요약 표의 `Judgment (gated)` 행(`jev_ask`) 문장 개정(같은 기준 216행) | 013·025 | 아니오(`paths:` 조건부) |
| 17 | `.moai/config/sections/workflow.yaml` | `jev:` 주석 개정 | 013·025 | — |
| 18 | `.moai/specs/SPEC-JEV-CORE-001/spec.md` (템플릿 사본 없음) | REQ-JEVC-011·012 개정 표지, `### Out of Scope — authority` 두 항목, HISTORY 행, `version` — **manager-spec 이 작성**(§11.2) | 013·025 | — |
| 19 | `CLAUDE.local.md` (템플릿 사본 없음, 로컬 전용) | §29 3등급 문장 한 줄 예외 — **운영자가 레인 세션에서 확인한 뒤에만** | 013 | — |
| 20 | `internal/contract/receipt/` (신규) + 테스트 | 사건 저장소·체인·검증 | 012 | — |
| 21 | `internal/contract/kickoff/` (신규) + 테스트 | 전제조건·결정 규칙 R1~R5·decide·kickoff-check·활성 판정·연동 상수 | 007~011·025 | — |
| 22 | `internal/contract/revoke/` (신규) + 테스트 | revoke | 022 | — |
| 23 | `internal/contract/sign/` (A1 패키지) + 테스트, 그리고 `internal/contract/receipt.go`(A1 패키지 `contract` 의 `ReceiptOutcome` — BASE `7fe658815` 기준 주석 208행·함수 216~232행) + `internal/contract/receipt_test.go` + `internal/contract/doc.go`(주석만 — 47행 시그니처 목록과 135~141행 `# Kickoff receipt` 문단. 141행 이전의 「ReceiptOutcome then applies the interim A1 rule …」 서술은 조건화 뒤 조건부 서술로 고친다) | 서명·재봉인 직후 `events.jsonl` 사건 추가(주입 가능한 저장소 이음매), 실패 시 서명 파일 미기록 — 요구는 A1 0.5.2 (25283ebf8) §C.6 이 A3 에 넘김, 추가 지점은 A3 소관. **모든 서명 테스트는 `t.TempDir()` 를 가리키는 `MOAI_HOME` 을 쓰도록 바꾼다**(실제 홈에 쓰지 않음). 서명기 단계 (1)(임시 규칙)은 `ReceiptOutcome` 의 첫 분기(`EffectiveDecider == DeciderLLMJev` → `RefuseReceiptRequiresHuman`)이므로 **그 자리에서** doctrine 플래그를 인자(주입 가능한 이음매)로 받게 조건화하고 호출부 `sign/sign.go`(BASE 519행)가 그 값을 넘기며 CLI 가 상수 `jevDoctrineAmended` 를 넘긴다(M7, 상수 거짓이라 동작 불변). `sign/` 에 같은 규칙을 다시 두지 않는다. A1 의 AC-CONTRACT-016 [REF] (t) 테스트 행 — `sign/ac_contract_016_test.go` 의 `t_llm_jev_both_approve_interim_rule` 와 `receipt_test.go` 의 `TestValidateKickoffReceipt_AC016`·`TestReceiptOutcome` 의 (t) 경우 — 을 A3 의 대체 테스트로 교체(AC-GR-017) | 012·025 |
| 24 | `internal/cli/contract_decide.go`, `contract_kickoff_check.go`, `contract_revoke.go` + 테스트, 그리고 `internal/cli/contract.go`(A1 파일 — BASE `7fe658815` 기준 `newContractCmd` 의 하위 명령 등록 516행 `cmd.AddCommand(verifyCmd, showCmd, signCmd)`, 서명 옵션·이음매 조립 389~414행 `sign.Options{…}`·`sign.Seams{…}`·`sign.Sign(opts, seams)`) | A1 `contract` Cobra 명령에 하위 명령 추가. `contract.go` 의 편집은 **그 두 곳으로만** 한정한다 — 516행에 세 하위 명령 등록, 389~414행에 doctrine 플래그(`jevDoctrineAmended`)와 사건 저장소 이음매 전달. 그 밖의 `contract.go` 줄은 바꾸지 않는다 | 007·011·022·012·025 | — |
| 25 | `internal/template/contract_mode_blocks_test.go`, `contract_mode_guided_test.go` (신규, 템플릿 사본 없음) | 가드·보존·동등·변경 집합 | 002·023·024 | — |
| 26 | `.moai/specs/SPEC-AUTONOMY-GATE-REWIRE-001/**` | 진행 기록 | — | — |
| 27 | **sync 단계 문서 산출물 (sync-phase 행)** — `CHANGELOG.md`, `docs-site/content/ko/cli-reference/contract.md`, `docs-site/content/en/cli-reference/contract.md`, `docs-site/content/ja/cli-reference/contract.md`, `docs-site/content/zh/cli-reference/contract.md` (이 다섯 경로뿐, 템플릿 사본 없음) | sync 단계(manager-docs)에서만 편집. `CHANGELOG.md` 는 `[Unreleased]` 항목만, `contract.md` 네 로케일은 이 SPEC 이 더하는 `contract` 하위 명령(`kickoff-check`·`decide`·`revoke`) 문서만. 그 밖의 줄은 바꾸지 않는다 | 007·011·022 (사용자 문서) | — |

행 2~14 가 블록 방식(추가형)이고, 15~19 는 원칙 개정(비추가형, §11), 20~25 는 코드, 26 은 이 SPEC 디렉터리, 27 은 sync 단계 문서 산출물이다(v0.3.7 — sync 가 AC-GR-003 허용 목록에 막혀 추가, 레인 결정 선택지 a). A1 SPEC 문서는 편집 대상이 아니다 — A1 0.5.2 가 임시 규칙과 알림 문언을 「A3 가 착지하기 전/후」 조건부로 적어 두었다. 이 목록 밖의 파일 변경은 AC-GR-003 이 실패로 잡는다(D13).

## §3. SSOT — `contract-autonomy.md`

frontmatter:

```yaml
---
description: "Contract-mode gate rewiring — which gates a signed SPEC contract replaces, the autonomous Kickoff check, revocation, and the one-pass lifecycle"
paths: "**/.moai/specs/**/contract.yaml,**/.moai/specs/**/kickoff-receipt.json,**/contract-autonomy.md"
---
```

`paths:` 를 계약 파일로 좁혔다(D11) — guided 세션에서 `run.md`·`plan.md` 를 읽어도 SSOT 가 로드되지 않는다. contract 모드에서는 블록 포인터가 SSOT 를 명시적으로 읽게 하고, 계약 파일을 읽는 순간에도 로드된다.

절 구성(영어 — 규칙 본문 언어 정책). 각 절 제목은 AC-GR-005 가 절 단위로 검사하는 앵커다.

1. `## Scope and activation` — 적용 조건, 무효값 → guided, `MOAI_AUTONOMY_TIER` 와 다른 축, 판독은 `moai contract kickoff-check --json`(D-1 — `show --json` 에는 결정자 필드가 없다).
2. `## The signing gate` — kickoff-check 가 plan→run 게이트. 사람 결정자 절차(서명 명령 보고 → 턴 닫기 → 운영자 서명 → 다음 턴 kickoff-check). 사람의 개입은 서명 시점으로 옮겨진다.
3. `## Equivalence clause (human signature only)` — 사람 서명에 한정한 등가. 비인간 서명은 §7.1 조건으로만.
4. `## Gate disposition` — 게이트 이름별 처분표(guided / contract).
5. `## Gates a contract never touches` — 질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`, 헌법, sync-auditor must-pass, main/release, 카드 선택, goal 상한, 파괴적 명령 확인, Report-Before-Ask.
6. `## Escalation routing` — 보고이지 질문이 아님. 기록 형식은 A2. 라우팅 사유 목록.
7. `## One-pass lifecycle` — 일곱 단계 표(단계 · 입력 · 기록해야 할 증거 · 넘어가는 조건). 증거 열은 REQ-GR-020 의 목록을 그대로 담는다. Push 행은 A4 착지 전 비활성.
8. `## Guided mode` — 이 파일 전체가 guided 에서 비활성.
9. `## Autonomous Kickoff` — 결정자 설정(`llm` 기본, `llm+jev` 교차 확인, `jev` 거절), §7.1 활성 조건을 **이 design.md 를 가리키지 않고** 규칙 문장으로 요약(템플릿 중립성), 여섯 전제조건, 결정 규칙 R1~R5, Jev 실패의 `llm` 대체와 그 기록, 작성자 배제. **M1 에서는 이 절에 활성 문언을 쓰지 않는다** — M1 판은 「이 절은 활성 조건이 충족된 뒤 채워진다」 한 문장이며 `llm+jev` 리터럴을 쓰지 않는다. 활성 문언은 M8 에서 전용 블록 id `contract-autonomous-kickoff` 로 들어간다.
10. `## Revocation` — revoke 의 효과·정지 시점·하지 않는 일.

## §4. guided 세션의 문맥 증가 — 정량

| 표면 | guided 세션에서 로드되는가 | 상한 |
|---|---|---|
| always-loaded 블록 3개(`CLAUDE.md` 2, `askuser-protocol.md` 1, `goal-directive.md` 1) | 매 세션 | 블록당 500자, 합계 1,500자 |
| `moai-mcp-tools-catalogue.md` 두 `jev_ask` 행 개정(행 15·16) | always-loaded 아님 — `paths:`(`**/moai-mcp-tools.md`, `**/internal/cli/mcp_server.go`, `**/.claude/agents/moai/*.md`)에 맞는 파일을 읽을 때만 | 행당 증가 300자 이하, 사본당 합계 600자 이하 |
| 스킬 파일 블록(행 6~14) | 해당 스킬 파일을 읽을 때만 | 블록당 900자 |
| SSOT | 계약 파일을 읽을 때만 — guided 세션에서는 계약 파일이 없으므로 0 | — |

따라서 guided 세션의 상시 증가는 **최대 1,500자**이고, `/moai run` 등으로 스킬 파일을 읽을 때 그 파일의 블록 수 × 900자 이하가, 카탈로그가 로드될 때 600자 이하가 더해진다. t1175 이전 판(v0.3.3)은 `moai-mcp-tools.md` 의 `jev_ask` 행 증가(300자)를 상시 예산에 넣었으나, t1175 가 그 행을 카탈로그로 옮겨 always-loaded 표면에서 Jev 개정 증가는 0 이다. 측정은 `LC_ALL=en_US.UTF-8 wc -m` 기준, `BASE` 대비 증가분이다. 이 값은 t1175 와의 조율 한도이며(D-6) 초과하지 않는다.

## §5. 검증 기제 — 모두 Go 테스트

AC 명령은 워크트리 세션 가드에 막히지 않도록 `go test` 한 줄이다(D12). `git` 이 필요한 검사는 테스트 안에서 `exec` 로 부르고, 기준 ref 는 환경 변수 `MOAI_GR_BASE` 로 받는다.

| 테스트 | 파일 | `MOAI_GR_BASE` 필요 | 무엇을 보는가 |
|---|---|---|---|
| `TestContractModeBlocksWellFormed` | `internal/template/contract_mode_blocks_test.go` | 아니오 | 템플릿 트리 모든 `.md`: 짝 맞춤, 중첩 없음, evolvable 구간 밖, 금지 클래스 없음, 블록 문자 상한(§4), 검사 대상 블록 ≥ 1 |
| `TestContractModeLocalTemplateParity` | 같은 파일 | 아니오 | 행 1~14 의 로컬·템플릿 블록 추출 결과 동일, SSOT 통째 동일 |
| `TestContractModeSSOTSections` | 같은 파일 | 아니오 | §3 의 절 제목 10개 존재, 각 절 **안에서만** 필수 토큰 검사(§5.1) |
| `TestContractModeGuidedPreservation` | `contract_mode_guided_test.go` | 예 | 행 2~14 × 2 사본: 걷어낸 결과가 `git show $BASE:<path>` 와 바이트 동일 |
| `TestContractModeInheritedDivergence` | 같은 파일 | 예 | 걷어낸 로컬·템플릿 쌍의 차이 = 기준 ref 쌍의 차이 |
| `TestContractModeChangeSetAllowlist` | 같은 파일 | 예 | `git diff --name-only $BASE` ⊆ §2 허용 목록 |
| `TestContractModeConstitutionDriftNotIncreased` | 같은 파일 | 예 | `git archive $BASE` 를 `t.TempDir()` 에 풀고, 그 트리와 현재 트리에 `internal/constitution` 의 `Validate`(`moai constitution validate` 와 같은 구현)를 각각 실행해 상태가 OK 가 아닌 **모든 범주**의 `(sentinel, id)` 쌍을 모은다. 현재 트리의 집합 ⊆ BASE 의 집합이고 `drift_count`·`missing_count`·`unregistered_count` 가 각각 BASE 이하여야 통과(BASE 에 이미 있는 DRIFT 9건은 이 SPEC 범위 밖 — 리드가 별도 카드 발행). 전제 단언: `MOAI_CONSTITUTION_SKIP_VALIDATE` 를 지운 상태에서 두 결과 모두 `Skipped == false`, BASE 쪽 DRIFT id 집합 = EV-6 의 9개, BASE `missing_count`·`unregistered_count` = 0. 두 집합을 `t.Log` 로 출력. **`constitution.Validate` 는 `ZONE_UNREGISTERED`·`ANCHOR_NOT_FOUND` 를 내지 않으므로**(plan-audit-6 D56) 미등록 `[HARD]` 유입은 테스트 자체의 `[HARD]` 집합 비교가 판정한다 — always-loaded 대상 여섯 사본(`CLAUDE.md`·`askuser-protocol.md`·`goal-directive.md` 로컬·템플릿)마다 `[HARD]` 를 담은 줄의 중복 제거 집합이 현재 ⊆ BASE. 반증 하위 테스트 두 개: 등록 조항 하나를 지운 사본(DRIFT +1 → 범주 비교 FAIL), 대상 파일 사본에 미등록 `[HARD]` 줄 하나 삽입(→ `[HARD]` 집합 비교 FAIL) — 둘 다 FAIL 관측(AC-GR-003) |
| `TestContractModeAlwaysLoadedBudget` | 같은 파일 | 예 | §4 표의 상시 증가 상한과 카탈로그 조건부 상한 |
| `TestContractModeEmitterSites` | 같은 파일 | 예 | `BASE` 에서 Kickoff 를 담은 파일 집합 = E ∪ R ∪ H ∪ 로컬 전용, E 의 각 파일에 블록 ≥ 1 |

같은 `contract_mode_blocks_test.go` 에 BASE 없이 도는 내용 검사 여섯 개를 더 둔다: `TestContractModeLifecycleOrder`(절·블록 범위 안의 단계 순서), `TestContractModeLifecycleEvidence`(단계 표의 증거·진행 조건 열), `TestContractModeBlockCondition`(첫 문장 조건), `TestContractModeAuditRetryBlocks`, `TestContractModeSyncBlocks`, `TestContractModeSigningBlocks`, 그리고 `TestJevDoctrineAmendment`(§11 다섯 위치; `SPEC-JEV-CORE-001` 은 저장소 루트 기준 경로로 읽는다).

`MOAI_GR_BASE` 가 없으면 이 표에서 「예」인 테스트는 `t.Skip` 한다(CI 는 BASE 를 모름). AC 는 `--- PASS` 줄을 요구하며 `--- SKIP` 은 실패로 본다 — 빈 선택이 통과로 읽히지 않게 한다(`verification-completeness.md §1.1`).

### §5.1 절 단위 토큰 (D19)

| 절 | 필수 토큰 |
|---|---|
| The signing gate | `kickoff-check`, `AskUserQuestion`(「내지 않는다」 문맥), `interactive terminal`, `reject`, `human decision required` |
| Equivalence clause | `signer_kind: human`, `interactive-tty` |
| Gate disposition | `Socratic`, `approach`, `assumption`, `plan-audit`, `quality gate`, `gate-sync-2` |
| Gates a contract never touches | `question-channel`, `sync-auditor`, `/clear` |
| Escalation routing | `Report-Before-Ask`, `escalation/` |
| One-pass lifecycle | 일곱 단계명, `verdict.md`, `audit_multi`, `mutation` |
| Autonomous Kickoff | M8 이후에만: `decide`, `llm+jev`, `fallback`, `jev_min_confidence`, `author-decider-conflict` |
| Revocation | `moai contract revoke`, `next stage boundary` |

## §6. 결정과 잔여 위험

| 결정 | 근거 | 잔여 위험 |
|---|---|---|
| 기존 텍스트 무수정, 블록 추가 | 블록 밖 보존을 기계로 판정 | guided 세션도 블록을 읽는다(§4 상한). 모델이 조건문을 오독할 가능성은 텍스트 검사로 배제되지 않는다 |
| Frozen 태그 문단에는 사람 서명 등가만 | A-Q1 은 사람의 승인이 서명 시점으로 옮겨질 뿐이라 방어 가능. `llm`·`llm+jev` 결정자에는 사람이 없으므로 「user approval」과 등가로 선언할 수 없다(D4). zone-registry 미등록은 약한 증거로만 취급 | 비인간 서명 경로는 그 문단의 별도 개정 전까지 비활성 |
| 판정을 Go 로(kickoff-check) | A1 이 저장소 대조 거절을 A3 에 맡겼다(A1 0.5.2 §C.6) | 같은 권한의 행위자는 저장소 체인 전체를 다시 쓸 수 있다 |
| 모든 서명 사건을 저장소에 | A1 `supersedes` 는 재봉인 때 다시 쓸 수 있어 증거가 아니다(A1 0.5.2 §C.6·§H) | 저장소 추가가 A1 서명기 경로를 바꾼다 — A1 감사 통과본과 대조 필요 |
| 결정 규칙 R1~R5 는 A3 소유, 스키마는 A1 | 리드 결정 2026-09-26 — 규칙을 두 SPEC 에서 규범적으로 정의하면 어긋난다(D8). A1 0.5.2 는 결정자 값 집합·파생 기본값·영수증 필드·검증만 소유 | A1 은 기록된 `outcome` 이 답과 모순되지 않는지만 본다 — 규칙 위반 영수증을 A1 이 잡지 못하므로 규칙의 정확성은 A3 의 테스트(AC-GR-019·025)와 moai 발급(kickoff-check 의 저장소 대조)에 달렸다 |
| Jev 단독 거절, 교차 확인만 허용 | 운영자 재결정 2026-09-26 | 교차 확인은 두 신호가 모두 승인해야 시작하므로 자율 시작 빈도는 `llm` 단독보다 낮다 |
| `reject` 와 `human` 을 구분해 기록 | A1 0.5.2 의 `outcome` 집합, § Cross-check Rules(참고). 운영자 결정 「`llm` 거절 → 사람」은 두 값 모두 Kickoff 를 사람에게 보내는 효과로 충족 | 기록을 읽는 쪽이 두 값을 다르게 다루면 효과가 갈릴 수 있다 — SSOT 가 두 값의 효과가 같음을 명시한다 |
| 원칙 개정과 두 해제를 한 커밋으로, §29 포함 | 리드 결정 2026-09-26(D32·D33) — 어느 하나만 착지하면 원칙과 코드가 어긋난다(REQ-GR-025) | 운영자 확인이 늦으면 연동 전체가 늦어진다. 두 작성자가 차례로 쓰므로 한 커밋 조립 전 두 결과를 따로 검증해야 한다(§11.2) |
| plan-audit 판정을 산출물 해시에 묶음 | A1 §C.7 이 판정의 증거 결합을 A3 에 넘겼다 — 보고서 파일 해시만으로는 판정 뒤 산출물 수정을 막지 못한다(D30) | 카드 증거 경로 보고서에 `plan_artifact_hash:` 줄을 쓰는 쪽이 없으면 전제조건 (a) 가 늘 실패해 결과가 사람이다 — 안전한 방향의 공백(`plan.md §C`) |
| 작성자 배제는 트레일러 판독기 + 자기 신고 | 세션 식별자를 커밋에 남기는 장치가 없다(D28). git 트레일러 파서는 이 저장소 커밋의 값을 보지 못하고, `internal/spec` 의 정규식 판독기는 본다 | 판단 파일의 결정자 신원은 자기 신고라 거짓 신고를 막지 못한다 |
| Jev 실패는 `llm` 단독으로 대체 | 운영자 재결정 — 쓸 수 없는 신호 때문에 멈추지 않는다 | 대체 동안 두 번째 신호가 없다. 대체 사유가 영수증과 저장소에 남으므로 빈도는 사후에 잴 수 있다 |
| 두 번째 리뷰 정지는 A4 | A1 §C.1 | A4 착지 전 contract 모드 Push 단계 비활성 |
| `mission-governor` 미사용 | 그 정의가 승인을 배제(리드 L3) | 주 LLM 역할의 품질은 새 컨텍스트 판단 역할의 프롬프트에 달림 |
| 에이전트·헌법 파일 무수정 | 범위 밖 | manager-develop 전제가 「Kickoff」로 읽힌다 → 오케스트레이터가 위임문 Section A 에 kickoff-check 결과를 적는다 |

## §7. Kickoff 자율 승인

### §7.1 활성 조건 — 유일한 정본 (D18)

설정 결정자가 `llm` 또는 `llm+jev` 여도 아래 1~4·6행이 **모두** 참일 때만 활성이다. 5행은 전체 활성의 조건이 아니라 **Jev 가 답한 교차 확인 결정이 사람 대신 결정할 수 있게 되는** 조건이다(규칙 R3, REQ-GR-025). 다른 문서는 이 표를 가리킬 뿐 다시 적지 않는다.

| # | 조건 | 소유 | 기계 확인 |
|---|---|---|---|
| 1 | `moai contract revoke` 존재 | A3 | `internal/cli` 명령 트리 테스트 |
| 2 | moai 발급 영수증 — `decide` + 저장소(`receipts.jsonl`·`events.jsonl`), 모든 서명 사건 기록 | A3 (A1 0.5.2 §C.6 이 넘긴 요구) | `internal/contract/receipt` 테스트 |
| 3 | A1 영수증 서명 경로(`sign --signer <kind> --receipt`)가 `llm` 영수증과 대체 영수증(요청 `llm+jev` → `effective_decider: llm`)을 서명 | A1 — A1 0.5.2 (25283ebf8) REQ-CONTRACT-024(출처 `file`), design § Kickoff Receipt 필드 규칙 3·4 | A1 테스트 |
| 4 | push 직렬화 강제 + 에이전트발 대화형 `sign` 차단 + 팩토리 에이전트 세션의 `decide` 거절 | A2b (t1245) — 리드 결정 2026-09-26, A1 0.5.2 §C.1·§C.2 와 일치 **[A2b SPEC 으로 재확인]** | A2b 테스트 |
| 5 | Jev 원칙 개정(§29 포함) + 규칙 R3 해제 + A1 임시 규칙 해제 — **한 커밋**(REQ-GR-025), **전체 활성의 조건 아님** | A3 | 상수 `jevDoctrineAmended` 한 비트(R3 와 A1 서명기 단계 (1) 이 함께 읽음)와 개정 표지 전부가 같은 커밋에서 처음 참 |
| 6 | `orchestration-mode-selection.md` Frozen 태그 문단의 비인간 서명 개정 | 운영자의 명시적 개정 결정(이 SPEC 범위 밖) | 문단에 비인간 결정자 서명이 명시됨 |

**코드상 표현.** `internal/contract/kickoff` 에 컴파일 상수 `autonomousKickoffEnabled`(기본 `false`)와 `jevDoctrineAmended`(기본 `false`, 내보내기)를 둔다. 판정 함수는 두 값을 인자로 받고 CLI 가 상수를 넘긴다(테스트가 모든 상태를 주입할 수 있게). 비활성 동안 kickoff-check 는 `llm`·`llm+jev` 서명을 `autonomous-kickoff-inactive` 로 거절한다. A1 서명기 단계 (1) 은 `internal/contract/receipt.go` 의 `ReceiptOutcome`(패키지 `contract`, `sign/` 이 아니다) 첫 분기에 있다. M7 은 **그 자리에서** `ReceiptOutcome` 이 doctrine 플래그를 인자(주입 가능한 이음매)로 받아 그 값이 거짓일 때만 그 분기를 적용하도록 조건화하고, 호출부 `sign/sign.go` 가 받은 값을 넘기며, CLI 가 `jevDoctrineAmended` 를 넘긴다 — `sign/` 에 같은 규칙을 다시 두지 않고, 상수가 거짓이므로 A1 의 현재 동작은 그대로다. `contract` 패키지는 `kickoff` 를 가져오지 않는다(플래그는 인자로만 들어온다 — 가져오기 순환 없음).

**「A1 임시 규칙 해제」의 표지 (D34).** 문자열이 아니라 동작이다: `internal/contract/sign` 의 테스트 `TestSignInterimRuleFollowsDoctrine` 가 상수 거짓일 때 `effective_decider: llm+jev`·두 답 approve·`outcome: approve` 영수증이 `receipt_requires_human` 으로 거절되고, 참일 때 같은 영수증은 서명되며 `reject`·`human` 은 각각 `receipt_rejected`·`receipt_requires_human` 임을 본다. 이 테스트는 **매 head 에서** 거짓·참 두 상태를 모두 주입해 검사하므로, 임시 규칙 단계를 지운 변이는 어느 커밋이 head 가 되든(상수가 참이어도) 거짓 주입 경우에서 「거짓인데 서명됨」으로 FAIL 한다. 이 테스트는 서명기 경로(`sign/`)를 통해 `ReceiptOutcome` 의 조건화된 분기를 실행한다. 이 테스트가 A1 AC-CONTRACT-016 [REF] (t) 의 대체이며, 대체되는 A1 테스트 행은 §2 23행에 적은 두 파일의 (t) 경우다(A1 0.5.2 (25283ebf8) acceptance 「A3 owns the replacing test」).

AC-GR-017 은 두 가지를 본다. (i) 순서: `autonomousKickoffEnabled = true` 커밋(M8)이 조건 1·2 를 착지시킨 커밋들의 **엄격한 후손**(같은 커밋 아님). (ii) 연동: 트리에서 개정 표지 전부(REQ-GR-013 의 위치, §29 포함)와 `jevDoctrineAmended = true` 가 모두 있거나 모두 없고, 이력에서 둘이 **처음 나타나는 커밋이 같다**. 조건 3·4 는 A1·A2b 테스트가 develop 에 있는지로, 조건 6 은 해당 문단의 문언으로 확인하며, 셋 중 하나라도 불충족이면 M8 은 보류한다.

**개발 순서**: M6 revoke → M7 저장소·decide·kickoff-check·서명기 사건 추가 → M7b 연동 커밋(원칙 개정 + `jevDoctrineAmended = true`, §11.2 순서) → (조건 3·4·6 충족 대기) → M8 활성.

### §7.2 역할

| 역할 | 누구 | 하지 않는 일 |
|---|---|---|
| `llm` 신호 | 리드 세션 또는 새 컨텍스트 판단 역할, SPEC 작성자와 다른 모델 선호. 결정자 `llm` 과 `llm+jev` 모두에서 판단 파일을 낸다 | `manager-spec`·`mission-governor` 아님 |
| `jev` 신호 | 결정자 `llm+jev` 일 때만 `decide` 가 직접 호출하는 두 번째 신호 | 문장 생성 없음. 단독 결정자가 되지 않는다(R5) |
| 전제조건·결정 규칙 판정 | `decide` 의 결정적 코드(규칙 R1~R5, A3 소유) | LLM 에 맡기지 않음 |
| 서명 | A1 영수증 경로 | A3 가 구현하지 않음 |
| 통과 판정 | `kickoff-check` | 아무것도 쓰지 않음 |

### §7.3 Jev 질문 형태

`internal/jev` 의 `Question`(noul / choice / score)을 쓴다. 상태는 계약 요약과 `llm` 판단의 결론 줄, 질문은 「이 계약으로 시작해도 되는가」(noul, approve/reject/escalate 로 사상)와 「가장 약한 줄은」(choice, 사유 참조로만 기록). 두 번째 클라이언트를 만들지 않는다. 원시 요청·응답은 결과와 무관하게 저장소에 남는다.

## §8. `moai contract decide <card>` (가칭)

A3 가 만드는 동사다 — A1 의 동사는 `show`·`verify`·`sign` 뿐이다. 정상 경로는 리드 세션의 LLM 이 도구 호출로 부르는 것이고, `MOAI_FACTORY_ROLE=agent` 세션에서의 거절은 A2b 의 가드다 **[A2b SPEC 으로 재확인]**.

입력: `<card>`(카드 id — 계약의 `card` 필드와 같아야 한다, A1 0.5.2 (25283ebf8) design § Card Field), `--spec <SPEC-ID>`, `--judgement <file|->`(JSON: 결정자 신고 `agent`·`model`, `answer` ∈ approve/reject/escalate, 계약 줄 참조가 있는 `reasons[]`, `confidence`), `--json`.

출력: `$MOAI_HOME/db/<project-key>/contract/receipts.jsonl` 에 영수증 한 줄, 같은 디렉터리 `events.jsonl` 에 `decide` 사건 한 줄, 그리고 A1 서명 경로의 입력인 `.moai/specs/<SPEC-ID>/kickoff-receipt.json`(영수증 줄의 본문과 바이트 동일). 이 셋 밖에는 쓰지 않는다.

절차:

1. 저장소 체인 검증. 깨졌으면 exit 1, 쓰기 없음.
2. `<card>` 와 계약 `card` 필드 대조. 다르면 exit 2 `card-mismatch`, 쓰기 없음.
3. 설정 결정자 판독(A1 설정 판독기, D-1). `jev` 이면 **R5**: exit 2 `decider-jev-refused`, 쓰기 없음, Jev 미호출. `human` 이면 `outcome: human`, 사유 `decider-human` — **결정하지 않는 경로**라 `events.jsonl` 사건만 쓰고 exit 0.
4. 입력 해시 측정(`contract.yaml` 서명 대상 digest, `acceptance.md`, plan-audit 판정 파일, 판단 파일).
5. 전제조건 (a)~(f) 평가(REQ-GR-009). (a) 의 판독은 `internal/runtime` 의 plan-audit 판독(`parsePlanAuditSnapshot`)과 같은 키 줄 규칙이다 — `Verdict:` 값이 정확히 `PASS` 또는 `PASS-WITH-DEBT`, `Overall Score:` 는 순수 부동소수(괄호 주석이 붙어 파싱이 안 되면 **실패**, 보수적·리드 승인)이고 Tier 문턱 이상, `plan_artifact_hash:` 가 같은 패키지의 plan-artifact 해시(`ComputeHash`)로 잰 현재 값과 같아야 한다. 실패 → Jev 미호출, `outcome: human`, `reason: precondition:<x>`, 사건만 쓴다(결정하지 않는 경로).
6. 작성자 배제: SPEC 디렉터리를 건드린 커밋들(`git log --format=%B -- <SPEC dir>`)의 본문을 `internal/spec` 의 `parseAuthoredByAgent`(`lint_ownership.go`)로 읽어 작성 에이전트 집합을 만든다 — `git log --format=%(trailers)` 는 이 저장소 커밋에서 빈 값을 낸다(`research.md §10.5`). 판단 파일의 `agent` 가 `manager-spec` 이거나 그 집합에 들면 `author-decider-conflict`, 트레일러를 가진 커밋이 없으면 `author-check-unmeasured` — 둘 다 `outcome: human`, 사건만 쓴다(결정하지 않는 경로).
7. 설정 결정자 `llm` 이면 **R4**: `llm` 답이 approve → `approve`, reject → `reject`, escalate → `human`.
8. 설정 결정자 `llm+jev` 이면 Jev 호출. 쓸 수 없으면 — 도달 불가·호출 실패(`jev_call_failed`), 자격 증명 없음(`jev_key_missing`), `workflow.jev.enabled: false`(`jev_disabled`), 응답 형식 오류(`jev_malformed_response`), 신뢰도 < `jev_min_confidence`(`jev_low_confidence`) — **R2**: `effective_decider: llm`, `fallback{applied: true, reason}` 로 R4 를 적용. Jev 가 답했으면: `jevDoctrineAmended` 가 `false` 면 **R3** `outcome: human`(`jev-doctrine-not-amended`); 아니면 **R1** 두 답이 모두 approve → `approve`, 모두 reject → `reject`, 불일치나 escalate → `human`(`cross-check-disagree`).
9. **결정하는 경로(7·8)에서만**: 영수증(A1 필드만: `requested_decider`·`effective_decider`·`fallback`·`llm_answer`·`jev_answer`·`outcome`·`inputs` — A1 0.5.2 (25283ebf8) design § Kickoff Receipt; 사유 코드는 넣지 않는다 — 필드 규칙 1 의 strict decode 가 모르는 필드를 거절한다)을 `receipts.jsonl` 에 추가하고, 그 줄 해시와 사유 코드를 담은 `decide` 사건을 `events.jsonl` 에 추가한 뒤, 같은 본문을 `kickoff-receipt.json` 에 쓴다. exit 0.

`reject` 와 `human` 은 모두 kickoff-check 를 통과하지 못하고(A1 서명기가 둘 다 거절하고, 위조 서명이 있어도 kickoff-check 가 `receipt-not-approved` 로 거절) Kickoff 를 사람에게 보낸다 — 오케스트레이터는 둘을 구별하지 않고 사람 서명 절차로 간다(REQ-GR-004). 규칙 R1~R5 는 A1 0.5.2 (25283ebf8) design § Cross-check Rules(A3 소유, 참고)와 일치한다 — 정본은 이 SPEC 이다.

## §9. 저장소 (REQ-GR-012)

위치 `$MOAI_HOME/db/<project-key>/contract/` — 리드 결정 R10(2026-09-26): A2 의 무장 상태·감사 로그와 같은 디렉터리. `$MOAI_HOME` 은 `internal/homestate` 가 해석하는 moai 홈이고 `<project-key>` 는 `homestate.ProjectKey` 가 만든다(`research.md §9.5`). 테스트는 `MOAI_HOME` 을 `t.TempDir()` 로 주어 루트를 해석하므로 실제 홈에 쓰지 않는다. A1 0.5.2 (25283ebf8) §C.6 도 같은 위치를 적는다.

| 파일 | 담는 것 |
|---|---|
| `receipts.jsonl` | moai 가 발급한 영수증 — 한 줄에 A1 영수증 본문 하나, Jev 를 호출했으면 원시 요청·응답 |
| `events.jsonl` | 모든 서명 사건 — 아래 다섯 종류 |

| 기준 | 프라이머리 `.moai/state/…` | `$MOAI_HOME/db/<project-key>/contract/` |
|---|---|---|
| 에이전트가 늘 쓰는 트리 밖인가 | 아니오 | 예 |
| 모든 워크트리가 같은 곳을 보는가 | 별도 해석 필요 | `homestate.ProjectKey` 공유 키 |
| 선례 | — | `backlog.db`, factory db, A2 무장 상태 |

`events.jsonl` 사건 종류와 필드:

| 종류 | 누가 쓰는가 | 핵심 필드 |
|---|---|---|
| `sign-human` | A1 서명기(사람 경로) — A3 개정 | spec, card, `seal`, `contract_sha256`, `acceptance_sha256` |
| `sign-receipt` | A1 서명기(영수증 경로) — A3 개정 | 위 + `signer_kind`(`llm`·`llm+jev`), `receipt.sha256` |
| `reseal` | A1 `--resign` — A3 개정 | 위 + `supersedes` |
| `decide` | `moai contract decide` | 대응 `receipts.jsonl` 줄의 해시, 적용한 규칙(R1~R5), 결정자 신고 신원·SPEC 작성자 트레일러 집합, 전제조건 평가 |
| `revoke` | `moai contract revoke` | spec, card, 폐기된 서명의 `seal`, 쓴 에스컬레이션 기록 경로, 사유 |

두 파일의 모든 줄: `prev`(직전 줄 sha256), `hash`, `at`, `kind`. 서명기는 사건 추가가 실패하면 서명 파일을 쓰지 않는다(원자성의 방향: 기록 없는 서명이 생기지 않게).

**변조 흔적의 겹**: (1) 해시 체인 — 중간 줄 변경 검출, (2) Jev 원문 보관, (3) 저장소에 없는 영수증·서명은 kickoff-check 가 거절, (4) 서명의 `signer_kind` 가 영수증의 `effective_decider` 와 다르면 `decider-mismatch`.

**꼬리 절단 (D21)**: 체인은 마지막 줄 삭제를 스스로 검출하지 못한다. 서명 사건 삭제 → kickoff-check 가 `signature-not-recorded` 로 거절(기록 없는 서명이 되므로). revoke 사건 삭제 → revoke 가 쓴 `kind: revoke` 에스컬레이션 기록이 두 번째 증인으로 남고, A3 revoke 판독기(§10)가 그것을 현재 seal 의 차단으로 읽어 전제조건 (e)·kickoff-check(`revoked`)·단계 경계가 재개를 막는다. A2 는 그 기록을 `status: resolved` 로 두고 `NeedsDecision` 에 세지 않으므로 차단 효과는 전적으로 A3 판독기의 몫이다. 영수증 줄 삭제 → 그 영수증으로 한 서명이 `receipt-not-issued` 로 거절된다. 남는 경우: 저장소와 에스컬레이션 기록을 **함께** 지우는 행위자 — 막지 못한다.

## §10. `moai contract revoke <card> --spec <SPEC-ID>`

| 상황 | 동작 | exit |
|---|---|---|
| `<card>` ≠ 계약 `card` 필드 | 쓰기 없음 | 2 |
| 계약이 서명됨(서명 종류·출처 무관 — 사람 서명 포함, D9) | 저장소 `revoke` 사건 + `revoke` 종류 에스컬레이션 기록 1건 | 0 |
| 이미 revoke 됐고 그 뒤 새 서명 없음(판독기가 현재 seal 에 대해 차단을 보고) | 쓰기 없음, 「already revoked」 | 0 |
| 계약 없음 또는 미서명 | 쓰기 없음 | 1 |
| 사용법·I/O·체인 무결성 오류 | 쓰기 없음 | 2 |

**쓰는 기록(A2 형식, M0 대조 — A2 0.4.3 spec §I·§I.1·§I.2, `internal/escalation/record.go`, BASE `7fe658815`).** A2 의 내보낸 함수로 만든다 — `escalation.RecordPath(<워크트리 최상위>, <card>, "revoke-operator", fp, 0)` 경로(= `.moai/reports/<card>/escalation/revoke-operator-<fp>.md`)에 `escalation.Record{...}.Marshal()` 결과를 쓴다. `fp = escalation.Fingerprint("revoke-operator", <폐기한 서명의 seal>)`. YAML 머리 필드: `schema_version: 1`, `card`, `spec`, `kind: revoke`, `class: revoke-operator`, `fingerprint: <fp>`, `contract_ref: ""`, `escalate_on: ""`, `status: resolved`, `decider: human`(A2 §I.2 가 revoke 기록에 요구하는 비어 있지 않은 값 — revoke 는 운영자 행위), `occurrences: 1`, `head_sha`, `detected_at`, `updated_at`, `not_observed: []`. 본문 세 절(`## Observation` 에 폐기한 seal 과 저장소 사건 해시, `## Options` 에 둘 이상의 선택지, `## Not observed`). A2 코드와 SPEC 은 바꾸지 않는다.

**A3 revoke 판독기(B3 — 재개 차단의 소유자).** A2 는 revoke 기록을 해결된 기록으로 적는다(`status: resolved`, `NeedsDecision` 은 `contract`·`operational` 종류만 센다). 그래서 A3 는 자기 판독기로 그 기록을 차단 신호로 읽는다 — **A2 는 revoke 를 해결됨으로 기록하고, A3 의 재개 차단은 A3 판독기의 책임이다.**

| 항목 | 규칙 |
|---|---|
| 읽는 파일 | `<워크트리 최상위>/.moai/reports/<card>/escalation/*.md`(`escalation.RecordDir` 와 같은 디렉터리, D-11 의 루트) |
| 파서 | A2 의 `escalation.ParseRecord`(YAML 머리만 해석) — 두 번째 파서를 만들지 않는다 |
| 차단 판정 | 기록 하나라도 `kind == "revoke"` 이고 `card`·`spec` 이 대상과 같으며 `fingerprint == escalation.Fingerprint(<그 기록의 class>, <현재 서명의 seal>)` 이면 차단 |
| 보지 않는 필드 | `status`(A2 가 늘 `resolved` 로 쓰므로), `decider`, `occurrences` |
| 풀리는 조건 | 새 서명(사람 대화형 또는 새 영수증) — seal 이 달라져 지문이 맞지 않는다 |
| 디렉터리 부재 | 에스컬레이션 디렉터리가 없으면 기록 0건 — 해제이며 오류가 아니다(A2 `NeedsDecision` 의 `filepath.Glob` 과 같은 의미. `os.ReadDir` 로 구현하면 `fs.ErrNotExist` 를 0건으로 처리) |
| 오류 | 있는 디렉터리의 목록 실패, 있는 기록 파일의 읽기·파싱 실패는 차단으로 판정(조용한 「차단 없음」 금지) |
| 쓰는 곳 | 전제조건 (e)(A2 `NeedsDecision` 과 함께), kickoff-check 사유 `revoked`(저장소 revoke 사건과 함께), revoke 의 멱등 판정, 오케스트레이터의 단계 경계 확인(kickoff-check 경유) |

이 판독 규칙은 AC-GR-023 이 고정한다. 옛 판(v0.3.3 이전)이 인용한 A2 커밋 `d8926ff9a` 의 `escalations/<timestamp>.json` 은 리드가 철회한 형식이다(`research.md §10.1`, 이력).
- 진행 중인 run 은 다음 단계 경계에서 멈춘다(D-9). 프로세스를 죽이지 않는다.
- 하지 않는 일: 워크트리 삭제, 브랜치 삭제·개명, push, 큐 변경, 계약·서명·SPEC 수정, 프로세스 종료, git 쓰기.

## §11. Jev 원칙 개정 (REQ-GR-013·025, 운영자 재결정으로 좁힘)

| 위치 | 현재 문장(요지) | 개정 |
|---|---|---|
| `SPEC-JEV-CORE-001` REQ-JEVC-011 | 표시 전용 — 답이 파일 등을 바꾸지 않는다 | 개정 표지, 같은 한 곳 예외: 교차 확인 두 번째 신호의 답을 moai 가 Kickoff 영수증과 저장소에 기록하는 일 |
| `SPEC-JEV-CORE-001` `### Out of Scope — authority` 두 항목(현재 154~156행) | Jev 답이 운영자 게이트에 닿는 경로는 범위 밖, v0.2.0 개정 뒤에도 게이트가 「선택하는」 답은 제외 | 두 항목 모두에 같은 한 곳 예외를 명시한다 — 개정 뒤 그 SPEC 이 REQ-JEVC-012 와 모순되지 않게 |
| `SPEC-JEV-CORE-001` REQ-JEVC-012 | 운영자 게이트에 결정·입력으로 쓰지 않는다 | `[AMENDED 2026-09-26 — v0.3.0; see HISTORY]` 표지, 「contract 모드 Kickoff 의 `llm+jev` 교차 확인에서 두 번째 신호로 쓰일 때」 한 곳 예외, HISTORY 행에 운영자 결정(「Kickoff 는 LLM·Jev 도 할 수 있게」, 재결정 — Jev 단독 없음, 교차 확인만) 인용. `status: completed` 유지 — 이 SPEC 의 v0.2.0 「clarifying amendment」 선례와 같은 방식 |
| `moai-mcp-tools-catalogue.md` 도구 표 `mcp__moai__jev_ask` 행(BASE 138행) | display-only — … never a completion predicate, merge approval, queue mutation, or gate input | 「except as the second signal of the contract-mode Kickoff `llm+jev` cross-check, where it can only confirm an LLM approval or route to a human」 수준의 한 구절 추가(300자 이하) |
| `moai-mcp-tools-catalogue.md` 계열 요약 표 `Judgment (gated)` 행(BASE 216행) | 같음 | 같은 구절(300자 이하) |
| `workflow.yaml` `jev:` 주석 | display-only | 같은 예외 한 문장(템플릿 중립 — 날짜·ID 없음) |
| `CLAUDE.local.md §29` 3등급 | 운영자 게이트에 Jev 입력 금지 | 한 줄 예외(같은 한 곳) — **run phase 전제: 운영자가 레인 세션에서 확인한 뒤에만 편집**(지시 파일은 교차 세션 메시지만으로 고치지 않는다) |

다른 게이트(완료 판정, 병합 승인, 큐 변경, 사용자 표면 동작, CodeRabbit 판정)와 Jev 단독 결정에 대한 금지는 그대로다. 표의 모든 행은 R3 해제·A1 임시 규칙 해제(`jevDoctrineAmended = true`)와 **같은 커밋**에 싣는다(REQ-GR-025). `CLAUDE.local.md §29` 도 그 묶음 안이다(리드 결정 2026-09-26, D33) — 운영자 승인은 §11.1 에 기록했다.

### §11.1 `CLAUDE.local.md §29` 한 줄 개정 문안 (운영자 확인 대상)

위치: §29 「[HARD] 되돌릴 수 없는 판정은 모델 답을 입력으로도 쓰지 않는다」 절의 첫 문단(「3등급 항목에서는 Jev 를 호출하지 않는다. …」) 바로 뒤에 한 줄로 넣는다. 확인 경로: 문안이 확정되면 리드가 운영자에게 직접 확인을 받아 레인에 전달한다.

운영자 승인 2026-09-26 (리드 중계, 레인 재확인 전), 문안 그대로. M7b 착수 때 레인 창에서 운영자 확인을 다시 받는다(리드 판정 D50 — 교차 세션 메시지는 동의가 아니다). 실제 편집은 카드 워크트리의 develop 사본 `CLAUDE.local.md` 에서만 하고 develop 병합으로 착지한다.

<!-- §29-amendment-text-start -->
```text
예외는 한 곳뿐이다 — contract 모드 Kickoff 의 `llm+jev` 교차 확인에서는 `moai contract decide` 가 Jev 를 두 번째 신호로 직접 부를 수 있다. Jev 답은 LLM 의 승인을 확인하거나 사람에게 보낼 뿐 혼자서 시작시키지 못하며, 그 밖의 3등급 항목에 대한 금지는 그대로다.
```
<!-- §29-amendment-text-end -->

### §11.2 연동 커밋의 조립 (리드 결정 2026-09-26, D32)

1. **manager-spec** 이 SPEC 본문 개정(`SPEC-JEV-CORE-001` REQ-JEVC-011·012, `Out of Scope — authority`, HISTORY·`version`)을 작성하고 돌아온다. A1 SPEC 문서는 고치지 않는다(A1 0.5.2 가 조건부 문언).
2. **그 뒤에** manager-develop 이 코드(`jevDoctrineAmended = true`), 규칙·설정 문장(`moai-mcp-tools-catalogue.md` 의 두 행·`workflow.yaml`, 로컬·템플릿), `CLAUDE.local.md §29` 한 줄(§11.1 문안, 카드 워크트리의 develop 사본)을 작성하고 돌아온다. 두 작성자는 동시에 쓰지 않는다 — 한 번에 한 작성자.
3. **레인 오케스트레이터** 가 두 결과를 명시 경로(`git add <path> …`)로 스테이징해 **한 커밋**을 만든다. 소유권은 작성 단위마다 지켜지고, 커밋은 오케스트레이터의 git 작업이다. 요구사항의 예외나 병합 트리 검사 장치는 두지 않는다.
4. 그 커밋의 검증 범위: `go test ./internal/contract/... ./internal/template/ ./internal/spec/`, `moai spec lint SPEC-JEV-CORE-001` — `internal/spec` 는 다른 SPEC 본문이 바뀌므로 포함한다(D42).

## §12. 테스트 설계 (RED 먼저)

| 대상 | RED 로 관측할 실패 입력 | 녹색 조건 |
|---|---|---|
| 사건 저장소 | 중간 줄 변조, 마지막 서명 줄 삭제 | 변조 → 검증 실패; 삭제 → kickoff-check `signature-not-recorded` |
| 서명 사건 기록 | 사람·영수증 서명, 재봉인 각각 | 각각 정확히 한 줄 추가; 추가 실패 주입 시 서명 파일 미기록 |
| kickoff-check | (1)~(11) 기존 + (12) 저장소 `reject` 영수증 + 일치 서명 / (13) 저장소 `human` 영수증 + 일치 서명 / (14) 봉인 불일치 서명 / (15) 미서명 계약 / (16) `effective_decider: llm` 영수증 + `signer_kind: llm+jev` 서명 / (17) 저장소 revoke 사건 없이 `kind: revoke` 에스컬레이션 기록(현재 seal 지문)만 | (12)(13) `receipt-not-approved` / (14)(15) `not-signed-valid`(verify 상태 동봉) / (16) `decider-mismatch` / (17) `revoked` — 판독기를 부르지 않는 변이는 exit 0 으로 FAIL |
| decide 전제조건 | (a)~(f) 각각, (a) 점수 줄에 괄호 주석, (e) 의 두 변형(열린 `kind: contract` 기록 / `status: resolved` revoke 기록만) | `human`, `precondition:e` 는 두 변형 모두, Jev 생성 0회 — A2 `NeedsDecision` 만 보는 변이는 revoke 변형에서 FAIL |
| R1 교차 확인 일치 | `llm` approve + Jev approve, `jevDoctrineAmended = true` | `approve` |
| R1 교차 확인 모두 거절 | `llm` reject + Jev reject, `jevDoctrineAmended = true` | `reject` |
| R1 교차 확인 불일치 | `llm` approve + Jev reject / escalate, `llm` reject + Jev approve, `llm` escalate + Jev approve, `llm` reject + Jev escalate, `jevDoctrineAmended = true` | `human`, `cross-check-disagree` |
| R2 Jev 대체 | 원인별 하위 테스트 5개(도달 불가 / 자격 증명 없음 / `jev.enabled: false` / 형식 오류 / 신뢰도 0.49) × `llm` approve·reject | `llm` 단독 R4 결과(approve / reject), 영수증·`receipts.jsonl` 에 `requested_decider: llm+jev`·`effective_decider: llm`·`fallback.reason` |
| R3 연동 전 Jev 답 | `llm` approve + Jev approve, `jevDoctrineAmended = false` | `human`, `jev-doctrine-not-amended` |
| R4 `llm` 단독 | 설정 결정자 `llm`, `llm` approve / reject / escalate | `approve` / `reject` / `human`, Jev 생성 0회 |
| R5 Jev 단독 거절 | 설정 결정자 `jev` | exit 2 `decider-jev-refused`, 저장소 줄 수 불변, Jev 생성 0회 |
| 작성자 배제 | 결정자 `manager-spec` / 신고 `agent` 가 트레일러 집합에 듦(본문 끝에 서명 줄이 붙은 커밋 — git 트레일러 파서는 못 봄) / 트레일러 없는 커밋만 | `author-decider-conflict` / `author-decider-conflict` / `author-check-unmeasured`, 세 경우 모두 영수증 파일 없음 |
| decide 이름·입력·출력 | `moai contract decide --help`, `<card>` ≠ 계약 `card`, 정상 입력 | 도움말에 동사 `decide`·`<card>`·`--spec`·`--judgement`·`--json`; `card-mismatch` exit 2; `receipts.jsonl`·`events.jsonl` 각 +1줄, `kickoff-receipt.json` 이 영수증 줄 본문과 바이트 동일 |
| decide 부작용 | 임시 git 저장소 + `backlog.db` | `backlog.db`·ref·워크트리 목록·`contract.yaml` 바이트 동일, 작업 트리의 새 파일은 영수증 하나 |
| revoke | 사람 서명 계약 / `llm` 영수증 서명 계약 / 반복 / 미서명 | 0(+1 사건 +1 기록) / 0 / 0(쓰기 0) / 1 |
| revoke 판독기 | 디렉터리 부재 / revoke 가 쓴 기록 / 같은 기록인데 새 서명(다른 seal) / `kind: revoke` 이나 지문이 현재 seal 과 다른 기록 / `status: open` 으로 바꾼 revoke 기록 / 머리가 깨진 기록 / `kind: contract`·`status: open` 기록 | 해제(오류 없음) / 차단 / 해제 / 해제 / 차단(`status` 무시) / 차단(오류는 차단) / A3 판독기는 해제·A2 `NeedsDecision` 은 참 — 그리고 revoke 기록에 대해 A2 `NeedsDecision` 은 거짓(A2 불변) |
| revoke 금지 사항 | 워크트리·브랜치·원격·`backlog.db` 픽스처 | 전후 동일, git 실행 이음매에 push·branch -d·worktree remove 0회 |
| 활성 순서 | 상수 `true` 를 조건 커밋과 같은 커밋에 둔 픽스처 저장소 | 순서 테스트 FAIL(엄격한 후손 아님) |
| 연동 | 개정 표지만 있는 커밋 / `jevDoctrineAmended = true` 만 있는 커밋 / §29 만 빠진 커밋 / 임시 규칙 단계가 삭제된 head(상수 참) / 모두 한 커밋 | 앞의 넷 FAIL(넷째는 `TestSignInterimRuleFollowsDoctrine` 의 거짓 주입 경우가 FAIL), 마지막 PASS |

모든 테스트는 `t.TempDir()` 와 그것을 가리키는 `MOAI_HOME` 으로 저장소 루트를 해석한다. 실제 moai 홈에 쓰지 않는다. Jev 는 스텁 이음매로만 부른다.
