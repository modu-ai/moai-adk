# design.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.0)

A1 기준: 0.4.1 (`6d98ca466`). A2 기준: 리드가 전달한 최종 형식(A2 개정본 대기).

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

## §2. 편집 허용 목록 (정본 — REQ-GR-025)

문서 행은 로컬 경로이며 같은 경로를 `internal/template/templates/` 아래에도 같은 커밋에서 편집한다(표시가 있는 행 제외).

| # | 경로 | 블록 id / 변경 | 요구 | always-loaded |
|---|---|---|---|---|
| 1 | `.claude/rules/moai/workflow/contract-autonomy.md` (신규) | 전체 | 전부 — SSOT | 아니오 |
| 2 | `CLAUDE.md` | `contract-signing-pipeline`(§2 ④ 뒤), `contract-safe-dev`(§7 Rule 5 뒤) | 004·014 | **예** |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | `contract-ambiguity`(§ Ambiguity Triggers and Exceptions › The Five Exceptions 뒤) | 014 | **예** |
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
| 15 | `.claude/rules/moai/core/moai-mcp-tools.md` | `jev_ask` 행 문장 개정 | 013 | **예** |
| 16 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | `jev_ask` 행 문장 개정 | 013 | 아니오 |
| 17 | `.moai/config/sections/workflow.yaml` | `jev:` 주석 개정 | 013 | — |
| 18 | `.moai/specs/SPEC-JEV-CORE-001/spec.md` (템플릿 사본 없음) | REQ-JEVC-012 개정 표지 + HISTORY 행 + `version` | 013 | — |
| 19 | `CLAUDE.local.md` (템플릿 사본 없음, 로컬 전용) | §29 3등급 문장 한 줄 예외 — **운영자가 레인 세션에서 확인한 뒤에만** | 013 | — |
| 20 | `internal/contract/receipt/` (신규) + 테스트 | 사건 저장소·체인·검증 | 012 | — |
| 21 | `internal/contract/kickoff/` (신규) + 테스트 | 전제조건·합의·decide·kickoff-check·활성 판정 | 007~011 | — |
| 22 | `internal/contract/revoke/` (신규) + 테스트 | revoke | 022·023 | — |
| 23 | `internal/contract/sign/` (A1 패키지) | 서명·재봉인 직후 사건 추가, 실패 시 서명 파일 미기록 — 요구는 A1 0.4.1 (6d98ca466) §C.6 이 A3 에 넘김, 추가 지점은 A3 소관 | 012 | — |
| 24 | `internal/contract/`(A1 코어) 합의 규칙 0번 | 개정 착지 뒤에만 해제 — A1 0.4.1 (6d98ca466) design § Agreement Rule 0번, §C.8 | 013 | — |
| 25 | `internal/cli/contract_decide.go`, `contract_kickoff_check.go`, `contract_revoke.go` + 테스트 | A1 `contract` Cobra 명령에 하위 명령 추가 | 007·011·022 | — |
| 26 | `internal/template/contract_mode_blocks_test.go`, `contract_mode_guided_test.go` (신규, 템플릿 사본 없음) | 가드·보존·동등·변경 집합 | 002·024·025 | — |
| 27 | `.moai/specs/SPEC-AUTONOMY-GATE-REWIRE-001/**` | 진행 기록 | — | — |

행 2~14 가 블록 방식(추가형)이고, 15~19 는 원칙 개정(비추가형, §11), 20~26 은 코드다. 이 목록 밖의 파일 변경은 AC-GR-003 이 실패로 잡는다(D13).

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

1. `## Scope and activation` — 적용 조건, 무효값 → guided, `MOAI_AUTONOMY_TIER` 와 다른 축, 판독은 `moai contract show --json`(D-1).
2. `## The signing gate` — kickoff-check 가 plan→run 게이트. 사람 결정자 절차(서명 명령 보고 → 턴 닫기 → 운영자 서명 → 다음 턴 kickoff-check). 사람의 개입은 서명 시점으로 옮겨진다.
3. `## Equivalence clause (human signature only)` — 사람 서명에 한정한 등가. 비인간 서명은 §7.1 조건으로만.
4. `## Gate disposition` — 게이트 이름별 처분표(guided / contract).
5. `## Gates a contract never touches` — 질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`, 헌법, sync-auditor must-pass, main/release, 카드 선택, goal 상한, 파괴적 명령 확인, Report-Before-Ask.
6. `## Escalation routing` — 보고이지 질문이 아님. 기록 형식은 A2. 라우팅 사유 목록.
7. `## One-pass lifecycle` — 일곱 단계 표(단계 · 입력 · 기록해야 할 증거 · 넘어가는 조건). 증거 열은 REQ-GR-020 의 목록을 그대로 담는다. Push 행은 A4 착지 전 비활성.
8. `## Guided mode` — 이 파일 전체가 guided 에서 비활성.
9. `## Autonomous Kickoff` — 결정자 설정, §7.1 활성 조건을 **이 design.md 를 가리키지 않고** 규칙 문장으로 요약(템플릿 중립성), 여섯 전제조건, 합의는 A1 규칙 참조, 작성자 배제, 「측정되지 않음 → 사람」, `jev`·`llm` 단독 불허. **M1 에서는 이 절에 활성 문언을 쓰지 않는다** — M1 판은 「이 절은 활성 조건이 충족된 뒤 채워진다」 한 문장이며 결정자 쌍 값 리터럴을 쓰지 않는다. 활성 문언은 M8 에서 전용 블록 id `contract-autonomous-kickoff` 로 들어간다.
10. `## Revocation` — revoke 의 효과·정지 시점·하지 않는 일.

## §4. guided 세션의 문맥 증가 — 정량

| 표면 | guided 세션에서 로드되는가 | 상한 |
|---|---|---|
| always-loaded 블록 3개(`CLAUDE.md` 2, `askuser-protocol.md` 1, `goal-directive.md` 1) | 매 세션 | 블록당 500자, 합계 1,500자 |
| `moai-mcp-tools.md` `jev_ask` 행 개정 | 매 세션 | 증가 300자 이하 |
| 스킬 파일 블록(행 6~14) | 해당 스킬 파일을 읽을 때만 | 블록당 900자 |
| SSOT | 계약 파일을 읽을 때만 — guided 세션에서는 계약 파일이 없으므로 0 | — |

따라서 guided 세션의 상시 증가는 **최대 1,800자**(1,500 + 300)이고, `/moai run` 등으로 스킬 파일을 읽을 때 그 파일의 블록 수 × 900자 이하가 더해진다. 측정은 `LC_ALL=en_US.UTF-8 wc -m` 기준, `BASE` 대비 증가분이다. 이 값은 t1175 와의 조율 한도이며(D-6) 초과하지 않는다.

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
| `TestContractModeAlwaysLoadedBudget` | 같은 파일 | 예 | §4 표의 상시 증가 상한 |
| `TestContractModeEmitterSites` | 같은 파일 | 예 | `BASE` 에서 Kickoff 를 담은 파일 집합 = E ∪ R ∪ H ∪ 로컬 전용, E 의 각 파일에 블록 ≥ 1 |

같은 `contract_mode_blocks_test.go` 에 BASE 없이 도는 내용 검사 여섯 개를 더 둔다: `TestContractModeLifecycleOrder`(절·블록 범위 안의 단계 순서), `TestContractModeLifecycleEvidence`(단계 표의 증거·진행 조건 열), `TestContractModeBlockCondition`(첫 문장 조건), `TestContractModeAuditRetryBlocks`, `TestContractModeSyncBlocks`, `TestContractModeSigningBlocks`, 그리고 `TestJevDoctrineAmendment`(§11 다섯 위치; `SPEC-JEV-CORE-001` 은 저장소 루트 기준 경로로 읽는다).

`MOAI_GR_BASE` 가 없으면 네 테스트는 `t.Skip` 한다(CI 는 BASE 를 모름). AC 는 `--- PASS` 줄을 요구하며 `--- SKIP` 은 실패로 본다 — 빈 선택이 통과로 읽히지 않게 한다(`verification-completeness.md §1.1`).

### §5.1 절 단위 토큰 (D19)

| 절 | 필수 토큰 |
|---|---|
| The signing gate | `kickoff-check`, `AskUserQuestion`(「내지 않는다」 문맥), `interactive terminal` |
| Equivalence clause | `signer_kind: human`, `interactive-tty` |
| Gate disposition | `Socratic`, `approach`, `assumption`, `plan-audit`, `quality gate`, `gate-sync-2` |
| Gates a contract never touches | `question-channel`, `sync-auditor`, `/clear` |
| Escalation routing | `Report-Before-Ask`, `escalation/` |
| One-pass lifecycle | 일곱 단계명, `verdict.md`, `audit_multi`, `mutation` |
| Autonomous Kickoff | M8 이후에만: `decide`, `jev_min_confidence`, `author-decider-conflict`, `not measured` |
| Revocation | `moai contract revoke`, `next stage boundary` |

## §6. 결정과 잔여 위험

| 결정 | 근거 | 잔여 위험 |
|---|---|---|
| 기존 텍스트 무수정, 블록 추가 | 블록 밖 보존을 기계로 판정 | guided 세션도 블록을 읽는다(§4 상한). 모델이 조건문을 오독할 가능성은 텍스트 검사로 배제되지 않는다 |
| Frozen 태그 문단에는 사람 서명 등가만 | A-Q1 은 사람의 승인이 서명 시점으로 옮겨질 뿐이라 방어 가능. 결정자 쌍은 사람이 없으므로 「user approval」과 등가로 선언할 수 없다(D4). zone-registry 미등록은 약한 증거로만 취급 | 비인간 서명 경로는 그 문단의 별도 개정 전까지 비활성 |
| 판정을 Go 로(kickoff-check) | A1 이 저장소 대조 거절을 A3 에 맡겼다(A1 0.4.1 §C.6) | 같은 권한의 행위자는 저장소 체인 전체를 다시 쓸 수 있다 |
| 모든 서명 사건을 저장소에 | A1 `supersedes` 는 재봉인 때 다시 쓸 수 있어 증거가 아니다(A1 0.4.1 §C.6·§H) | 저장소 추가가 A1 서명기 경로를 바꾼다 — A1 감사 통과본과 대조 필요 |
| 합의 규칙은 A1 이 정본 | 두 곳에서 규범적으로 정의하면 어긋난다(D8) | A1 이 바뀌면 이 SPEC 의 AC 도 따라 바뀐다 |
| 두 번째 리뷰 정지는 A4 | A1 §C.1 | A4 착지 전 contract 모드 Push 단계 비활성 |
| `mission-governor` 미사용 | 그 정의가 승인을 배제(리드 L3) | 주 LLM 역할의 품질은 새 컨텍스트 판단 역할의 프롬프트에 달림 |
| 에이전트·헌법 파일 무수정 | 범위 밖 | manager-develop 전제가 「Kickoff」로 읽힌다 → 오케스트레이터가 위임문 Section A 에 kickoff-check 결과를 적는다 |

## §7. Kickoff 자율 승인

### §7.1 활성 조건 — 유일한 정본 (D18)

결정자 설정이 `llm+jev` 여도 아래가 **모두** 참일 때만 활성이다. 다른 문서는 이 표를 가리킬 뿐 다시 적지 않는다.

| # | 조건 | 소유 | 기계 확인 |
|---|---|---|---|
| 1 | `moai contract revoke` 존재 | A3 | `internal/cli` 명령 트리 테스트 |
| 2 | moai 발급 영수증 — `decide` + 사건 저장소, 모든 서명 사건 기록 | A3 (A1 0.4.1 §C.6 이 넘긴 요구) | `internal/contract/receipt` 테스트 |
| 3 | A1 영수증 서명 경로(`sign --signer <kind> --receipt`) | A1 — A1 0.4.1 (6d98ca466) REQ-CONTRACT-024, 출처 `file` | A1 테스트 |
| 4 | push 직렬화 강제 + 에이전트발 대화형 `sign` 차단 | A2 또는 A2b — 소유 카드 상충(A1 0.4.1 은 A2b t1245, 리드·A2 `d8926ff9a` 는 A2 t1235) — `research.md §10.2` **[A2 개정본으로 재확인]** | A2 테스트 |
| 5 | Jev 원칙 개정(REQ-GR-013) + A1 합의 규칙 0번 해제 | A3 | 개정 표지 존재, 규칙 0번 테스트 반전 |
| 6 | `orchestration-mode-selection.md` Frozen 태그 문단의 비인간 서명 개정 | 운영자의 명시적 개정 결정(이 SPEC 범위 밖) | 문단에 결정자 쌍 서명이 명시됨 |

**코드상 표현.** `internal/contract/kickoff` 에 컴파일 상수 `autonomousKickoffEnabled`(기본 `false`)를 둔다. 비활성 동안 kickoff-check 는 결정자 쌍 서명을 `autonomous-kickoff-inactive` 로 거절한다. 상수를 `true` 로 바꾸는 커밋은 M8 이며, AC-GR-017 의 순서 테스트가 그 커밋이 조건 1·2·5 를 착지시킨 커밋들의 **엄격한 후손**(같은 커밋 아님)인지 확인한다. 조건 3·4 는 A1·A2 테스트가 develop 에 있는지로, 조건 6 은 해당 문단의 문언으로 확인하며, 셋 중 하나라도 불충족이면 M8 은 보류한다.

**개발 순서**: M6 revoke → M7 저장소·decide·kickoff-check·서명기 사건 추가 → M7b Jev 원칙 개정 → (조건 3·4·6 충족 대기) → M8 활성.

### §7.2 역할

| 역할 | 누구 | 하지 않는 일 |
|---|---|---|
| 주 LLM 결정자 | 리드 세션 또는 새 컨텍스트 판단 역할, SPEC 작성자와 다른 모델 선호 | `manager-spec`·`mission-governor` 아님 |
| Jev 결정자 | `decide` 가 직접 호출 | 문장 생성 없음 |
| 전제조건·합의 판정 | `decide` 의 결정적 코드(합의는 A1 validator 재사용) | LLM 에 맡기지 않음 |
| 서명 | A1 영수증 경로 | A3 가 구현하지 않음 |
| 통과 판정 | `kickoff-check` | 아무것도 쓰지 않음 |

### §7.3 Jev 질문 형태

`internal/jev` 의 `Question`(noul / choice / score)을 쓴다. 상태는 계약 요약과 주 LLM 판단의 결론 줄, 질문은 「이 계약으로 시작해도 되는가」(noul)와 「가장 약한 줄은」(choice). 두 번째 클라이언트를 만들지 않는다. 개정 착지 전에는 호출은 하되 A1 규칙 0번이 결과를 「측정되지 않음」으로 처리한다 — 원문은 저장소에 남는다.

## §8. `moai contract decide <card>` (가칭)

입력: `<card>`, `--spec <SPEC-ID>`, `--judgement <file|->`(JSON: 결정자 신고 `agent`·`model`, `verdict` ∈ approve/reject/escalate, 계약 줄 참조가 있는 `reasons[]`), `--json`.

절차:

1. 저장소 체인 검증. 깨졌으면 exit 1, 쓰기 없음.
2. 입력 해시 측정(`contract.yaml` 서명 대상 digest, `acceptance.md`, plan-audit 판정 파일, 판단 파일).
3. 전제조건 (a)~(f) 평가(REQ-GR-009). (a) 의 파싱 규칙은 `internal/runtime` 의 plan-audit 판독과 같다 — `Verdict:`·`Overall Score:` 키 줄, 점수는 순수 부동소수여야 하며 뒤에 괄호 주석이 붙어 파싱이 안 되면 **실패로** 처리(보수적). 실패 → Jev 미호출, `outcome: human`, `reason: precondition:<x>`.
4. 작성자 배제: `moai session current` 의 세션 식별자와 판단 신고값, plan-phase 커밋 `Authored-By-Agent:` 트레일러를 비교(REQ-GR-010).
5. Jev 호출. `available` 이 아니면 원시 상태를 기록하고 「측정되지 않음」.
6. A1 합의 규칙으로 결과 결정.
7. A1 형식의 영수증 파일 `.moai/specs/<SPEC-ID>/kickoff-receipt.json` 을 쓰고, 그 해시를 담은 사건을 저장소에 추가. exit 0.

## §9. 사건 저장소 (REQ-GR-012)

위치 `~/.moai/db/<project-key>/contract/receipts.jsonl`(`internal/homestate` 경로 헬퍼).

| 기준 | 프라이머리 `.moai/state/…` | `~/.moai/db/<project-key>/contract/` |
|---|---|---|
| 에이전트가 늘 쓰는 트리 밖인가 | 아니오 | 예 |
| 모든 워크트리가 같은 곳을 보는가 | 별도 해석 필요 | `homestate.ProjectKey` 공유 키 |
| 선례 | — | `backlog.db`, factory db |

사건 종류와 줄 필드:

| 종류 | 누가 쓰는가 | 핵심 필드 |
|---|---|---|
| `sign-human` | A1 서명기(사람 경로) — A3 개정 | spec, card(있으면), `seal`, `contract_sha256`, `acceptance_sha256` |
| `sign-receipt` | A1 서명기(영수증 경로) — A3 개정 | 위 + `signer_kind`, `receipt.sha256` |
| `reseal` | A1 `--resign` — A3 개정 | 위 + `supersedes` |
| `decide` | `moai contract decide` | §8 의 전 필드, Jev 원문, 발급 영수증 해시 |
| `revoke` | `moai contract revoke` | spec, card, 폐기된 서명의 `seal`, 쓴 에스컬레이션 기록 경로, 사유 |

모든 줄: `prev`(직전 줄 sha256), `hash`, `at`, `kind`. 서명기는 사건 추가가 실패하면 서명 파일을 쓰지 않는다(원자성의 방향: 기록 없는 서명이 생기지 않게).

**변조 흔적의 겹**: (1) 해시 체인 — 중간 줄 변경 검출, (2) Jev 원문 보관, (3) 저장소에 없는 영수증·서명은 kickoff-check 가 거절.

**꼬리 절단 (D21)**: 체인은 마지막 줄 삭제를 스스로 검출하지 못한다. 서명 사건 삭제 → kickoff-check 가 `signature-not-recorded` 로 거절(기록 없는 서명이 되므로). revoke 사건 삭제 → revoke 가 쓴 에스컬레이션 기록이 두 번째 증인으로 남아 전제조건 (e)와 단계 경계가 재개를 막는다. decide 사건 삭제 → 그 영수증으로 한 서명이 `receipt-not-issued` 로 거절된다. 남는 경우: 저장소와 에스컬레이션 기록을 **함께** 지우는 행위자 — 막지 못한다.

## §10. `moai contract revoke <card> --spec <SPEC-ID>`

| 상황 | 동작 | exit |
|---|---|---|
| 계약이 서명됨(서명 종류·출처 무관 — 사람 서명 포함, D9) | 저장소 `revoke` 사건 + `revoke` 종류 에스컬레이션 기록 1건 | 0 |
| 이미 revoke 됐고 그 뒤 새 서명 사건 없음 | 쓰기 없음, 「already revoked」 | 0 |
| 계약 없음 또는 미서명 | 쓰기 없음 | 1 |
| 사용법·I/O·체인 무결성 오류 | 쓰기 없음 | 2 |

- 에스컬레이션 기록 형식은 A2 최종 형식 `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + 기계 판독 YAML 머리, 종류 `revoke`(리드가 A2 에 요청) **[A2 개정본으로 재확인]**. 이 경로는 A2 현재 커밋 `d8926ff9a` 의 `escalations/<timestamp>.json` 과 다르다 — 그것은 리드가 철회한 형식이다(`research.md §10.1`).
- 진행 중인 run 은 다음 단계 경계에서 멈춘다(D-9). 프로세스를 죽이지 않는다.
- 하지 않는 일: 워크트리 삭제, 브랜치 삭제·개명, push, 큐 변경, 계약·서명·SPEC 수정, 프로세스 종료, git 쓰기.

## §11. Jev 원칙 개정 (REQ-GR-013, 리드 L1)

| 위치 | 현재 문장(요지) | 개정 |
|---|---|---|
| `SPEC-JEV-CORE-001` REQ-JEVC-012 | 운영자 게이트에 결정·입력으로 쓰지 않는다 | `[AMENDED 2026-09-26 — v0.3.0; see HISTORY]` 표지, contract 모드 Kickoff 교차 확인 한 곳 예외, HISTORY 행에 운영자 결정 「Kickoff 는 LLM·Jev 도 할 수 있게」 인용. `status: completed` 유지 — 이 SPEC 의 v0.2.0 「clarifying amendment」 선례와 같은 방식 |
| `moai-mcp-tools.md` `jev_ask` 행 | never … gate input | 「except the contract-mode Kickoff cross-check, where it is one of two required signals and can only route to a human or approve jointly」 수준의 한 구절 추가(300자 이하) |
| `moai-mcp-tools-catalogue.md` 같은 행 | 같음 | 같은 구절 |
| `workflow.yaml` `jev:` 주석 | display-only | 같은 예외 한 문장(템플릿 중립 — 날짜·ID 없음) |
| `CLAUDE.local.md §29` 3등급 | 운영자 게이트에 Jev 입력 금지 | 한 줄 예외 — **run phase 전제: 운영자가 레인 세션에서 확인한 뒤에만 편집**(지시 파일은 교차 세션 메시지만으로 고치지 않는다) |

다른 게이트(완료 판정, 병합 승인, 큐 변경, 사용자 표면 동작, CodeRabbit 판정)의 금지는 그대로다.

## §12. 테스트 설계 (RED 먼저)

| 대상 | RED 로 관측할 실패 입력 | 녹색 조건 |
|---|---|---|
| 사건 저장소 | 중간 줄 변조, 마지막 서명 줄 삭제 | 변조 → 검증 실패; 삭제 → kickoff-check `signature-not-recorded` |
| 서명 사건 기록 | 사람·영수증 서명, 재봉인 각각 | 각각 정확히 한 줄 추가; 추가 실패 주입 시 서명 파일 미기록 |
| kickoff-check | 사람 서명(기록 있음) / 저장소에 없는 영수증 파일 / `signer_kind: jev` / `llm` / 결정자 쌍 + 비활성 / 결정자 쌍 + revoke 뒤 | 0 / `receipt-not-issued` / `decider-not-permitted` ×2 / `autonomous-kickoff-inactive` / 거절 |
| decide 전제조건 | (a)~(f) 각각, (a) 점수 줄에 괄호 주석 | `human`, Jev 생성 0회 |
| decide Jev | disabled / no-credential / unreachable / 신뢰도 0.49 | exit 0, `human`, 원시 상태 기록 |
| decide 합의 | A1 규칙 매트릭스 | A1 validator 결과와 같음(에스컬레이션은 늘 사람) |
| 작성자 배제 | 결정자 `manager-spec` / 세션 식별자 일치 / 트레일러 일치 | `author-decider-conflict` |
| decide 부작용 | 임시 git 저장소 + `backlog.db` | `backlog.db`·ref·워크트리 목록·`contract.yaml` 바이트 동일, 새 파일은 영수증 하나 |
| revoke | 사람 서명 계약 / 결정자 쌍 서명 계약 / 반복 / 미서명 | 0(+1 사건 +1 기록) / 0 / 0(쓰기 0) / 1 |
| revoke 금지 사항 | 워크트리·브랜치·원격·`backlog.db` 픽스처 | 전후 동일, git 실행 이음매에 push·branch -d·worktree remove 0회 |
| 활성 순서 | 상수 `true` 를 조건 커밋과 같은 커밋에 둔 픽스처 저장소 | 순서 테스트 FAIL(엄격한 후손 아님) |

모든 테스트는 `t.TempDir()` 와 격리된 `MOAI_HOME` 을 쓴다. 실제 `~/.moai` 에 쓰지 않는다.
