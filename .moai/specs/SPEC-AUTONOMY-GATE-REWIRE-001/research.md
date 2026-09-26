# research.md — SPEC-AUTONOMY-GATE-REWIRE-001

측정 트리: 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`, HEAD `ca1d5dc43` (develop 과 동일). 측정일 2026-09-26. 아래 수치는 모두 이 트리에서 이 실행에 잰 값이다. **t1175 흡수 후 run phase 진입 시 전부 재측정한다**(`plan.md §C` M0).

## §1. G1 (Implementation Kickoff Approval) 위치 전수

### §1.1 계수

명령:

```bash
grep -rlE "Kickoff" CLAUDE.md .claude/rules .claude/skills/moai .claude/output-styles .claude/agents | wc -l          # 33
grep -rhcE "Kickoff" CLAUDE.md .claude/rules .claude/skills/moai .claude/output-styles .claude/agents | awk '{s+=$1} END{print s}'   # 116
T=internal/template/templates
grep -rlE "Kickoff" $T/CLAUDE.md $T/.claude/rules $T/.claude/skills/moai $T/.claude/output-styles $T/.claude/agents | wc -l   # 32
grep -rhcE "Kickoff" $T/CLAUDE.md $T/.claude/rules $T/.claude/skills/moai $T/.claude/output-styles $T/.claude/agents | awk '{s+=$1} END{print s}'  # 111
```

| 사본 | 파일 수 | 줄 수 |
|---|---|---|
| 로컬 | **33** | **116** |
| 템플릿 | **32** | **111** |

차이 1파일·5줄은 `.claude/agents/harness/workflow-specialist.md`(로컬 전용 하네스, 템플릿 미러 없음)다. 카드 문언 「8곳+」는 발화 지점 수를 가리킨 것으로 읽힌다 — 문자열 출현 기준으로는 33/32 파일이다.

**로컬↔템플릿 Kickoff 줄 차이: 0.** 템플릿이 있는 32개 파일 각각에서 `grep Kickoff` 줄 집합을 비교한 결과 모두 동일했다. 파일 전체 차이는 여러 곳에 있으나(아래 표 `diff줄`) Kickoff 줄이 아닌 곳이다. 설계 문서가 지적한 `orchestration-mode-selection.md` 분기(20줄)는 `§C.1 Agent Teams` 절(107~152행)에 있고 Kickoff 헤더(16~18행)와 무관하다.

### §1.2 분류 — 발화 지점 / 참조 지점 / 동음이의

- **발화 지점(E)**: 게이트를 정의하거나 `AskUserQuestion` 발화를 지시한다 → contract-mode 블록 필수(REQ-GR-011).
- **참조 지점(R)**: 게이트가 「이미 통과했다」를 전제로 삼거나 「바뀌지 않는다」고 말한다 → SSOT 등가 조항(REQ-GR-011)이 덮는다, 편집 없음.
- **동음이의(H)**: 다른 게이트가 같은 이름을 쓴다 → 범위 밖.

| # | 파일 (로컬 경로; 템플릿은 `internal/template/templates/` 접두) | 줄 | 분류 | 앵커(절 제목) | 로컬↔템플릿 diff줄 |
|---|---|---|---|---|---|
| 1 | `CLAUDE.md` | 1 | **E** | § 2. Request Processing Pipeline — ④ | 0 |
| 2 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 14 | **E** (SSOT) | 파일 머리 (제목 아래 인용 블록 + `[ZONE:Frozen] [HARD]` 문단), § §E — Anti-Patterns | 20 (§C.1 만) |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | 2 | **E** | § Recommendation mode, § Report-Before-Ask Gate › Exceptions | 0 |
| 4 | `.claude/rules/moai/workflow/goal-directive.md` | 3 | **E** | § Goal-Presentation Timing, § Hard Preconditions | 0 |
| 5 | `.claude/skills/moai/SKILL.md` | 2 | **E** | goal 라우팅 항목 (Progression mode 줄), completion-condition 절 | 41 |
| 6 | `.claude/skills/moai/workflows/moai.md` | 8 | **E** | § Pipeline Gates (named, in order) #2, Step 11.3 | 8 |
| 7 | `.claude/skills/moai/workflows/plan.md` | 2 | **E** | § [NEEDS CLARIFICATION] Marker Usage | 9 |
| 8 | `.claude/skills/moai/workflows/plan/spec-assembly.md` | 5 | **E** | Step 2.3.3a, Step 2.3.3b | 0 |
| 9 | `.claude/skills/moai/workflows/run.md` | 13 | **E** | § Run-phase Autonomy (ac_converge) › 1. | 7 |
| 10 | `.claude/skills/moai/workflows/goal.md` | 6 | **E** | § Progression Mode (Autonomous / Semi-autonomous) | 0 |
| 11 | `.claude/skills/moai/workflows/project/doc-generation.md` | 4 | R | `pipeline` carry 선택지 | 0 |
| 12 | `.claude/skills/moai/workflows/design.md` | 5 | R | 전제 목록 | 0 |
| 13 | `.claude/skills/moai/workflows/factory.md` | 1 | R | 게이트 표 행 1 (inherited) | 2 |
| 14 | `.claude/skills/moai/workflows/run/mode-orchestration.md` | 1 | R | factory 계약 | 22 |
| 15 | `.claude/skills/moai/workflows/run/phase-execution.md` | 2 | R | sweep launch procedure | 28 |
| 16 | `.claude/skills/moai/workflows/run/task-decomposition.md` | 1 | R | workflow 발사 | 42 |
| 17 | `.claude/rules/moai/workflow/session-handoff.md` | 4 | R | § Invariants (both modes) | 0 |
| 18 | `.claude/rules/moai/workflow/session-handoff-examples.md` | 5 | R | (41,616자, 한도 초과 파일) | 0 |
| 19 | `.claude/rules/moai/workflow/spec-workflow.md` | 5 | R | skip-eligibility (40,799자, 한도 초과 파일) | 0 |
| 20 | `.claude/rules/moai/workflow/goal-directive-detail.md` | 5 | R | T1 트리거, 통합 노트 | 0 |
| 21 | `.claude/rules/moai/workflow/dynamic-workflows.md` | 3 | R | 「Kickoff Approval is unaffected」 | 0 |
| 22 | `.claude/rules/moai/workflow/cadence-bridge.md` | 2 | R | human-only 조항 | 0 |
| 23 | `.claude/rules/moai/workflow/kanban-dispatch.md` | 1 | R | § Boundaries — No gate bypass | 0 |
| 24 | `.claude/rules/moai/workflow/cache-aware-execution.md` | 1 | R | § Non-goals | 2 |
| 25 | `.claude/rules/moai/workflow/archived-agent-rejection.md` | 1 | R | paste-ready resume 반례 | 0 |
| 26 | `.claude/rules/moai/development/coding-standards.md` | 1 | R | Bash risk-amplifier 교차참조 | 1 |
| 27 | `.claude/output-styles/moai/moai.md` | 3 | R | §8 handoff 조항 | 0 |
| 28 | `.claude/agents/moai/manager-develop.md` | 2 | R | Preconditions, 상태 전이 표 | 5 |
| 29 | `.claude/agents/moai/manager-design.md` | 1 | R | 실행 시점 | 2 |
| 30 | `.claude/agents/moai/plan-auditor.md` | 1 | R | MP-7 | 6 |
| 31 | `.claude/skills/moai/workflows/e2e.md` | 5 | **H** | § Kickoff Approval (one-time gate) — `--autofix` 루프 게이트 | 0 |
| 32 | `.claude/skills/moai/workflows/harness-build-entry.md` | 1 | **H** | Builder 승인 게이트 | 10 |
| 33 | `.claude/agents/harness/workflow-specialist.md` (로컬 전용) | 5 | 범위 밖 | — | 템플릿 없음 |

집계: **E 10 파일**(로컬 10 + 템플릿 10), R 20, H 2, 로컬 전용 1. 합 33.

`cadence-bridge.md` 는 Kickoff 를 「human-only」라 한다. 계약 서명도 운영자가 하는 행위이므로 등가 조항 아래서 그 성질은 유지된다 — 사람의 개입이 게이트 시점에서 서명 시점으로 옮겨질 뿐이다. 이 해석은 SSOT 에 명시한다(`design.md §3`).

## §2. G2/G3/G4/G7/G8/G11 조항 위치

| 게이트 | 파일 · 앵커 | 로컬↔템플릿 |
|---|---|---|
| G2 소크라테스 인터뷰 | `CLAUDE.md` § 7 Rule 5 (96행); `askuser-protocol.md` § Socratic Interview Structure (56행), § Ambiguity Triggers and Exceptions (232행); plan 워크플로 `plan/clarity-interview.md` | CLAUDE.md·askuser 동일; clarity-interview 15줄 차 |
| G3 접근 승인 | `CLAUDE.md` § 7 Rule 1 (89행); `moai.md` § Phase 3: Plan Annotation Cycle | 동일 / 8줄 차 |
| G4 가정 확인 대기 | `moai-constitution.md` § Agent Core Behaviors › 1. Surface Assumptions (190행); `AGENTS.md` §5 1. | 2줄 차(GLM 문장, 무관) |
| G7 plan-audit FAIL 재시도 | `plan/spec-assembly.md` Step 2.3.4: FAIL Path — Retry Loop (max 3 iterations) (260행) | 동일 |
| G8 FAIL 후 사용자 선택 | 같은 파일 Step 2.3.5: Escalation after 3 FAIL Iterations (274행), § Phase 15 › HUMAN GATE: SPEC Quality Validation (456행, WARNING/FAIL 시 AskUserQuestion) | 동일 |
| G11 sync 확인 질문 | `sync.md` § HUMAN GATE Map (83행); `sync/doc-execution.md` HUMAN GATE: Documentation Scope (`gate-sync-2`, 101행 — **evolvable 구간 안**); `sync/delivery.md` § Context-Aware Next Steps (408행), Error Flow 「Sync changes on current branch?」(490행); 실패 경로 `sync/quality-gates-context.md`·`quality-gates-quality.md` | sync 계열 전부 차이 있음(18~76줄, 트레이스 주석 등) |

`gate-sync-2` 는 `<!-- moai:evolvable-start id="gate-sync-2" -->` 구간 안에 있다. `internal/merge/evolvable_zone.go:33` 의 정규식이 그 구간을 사용자 진화 구간으로 다루므로, contract 블록은 그 구간 **밖**(evolvable-end 뒤)에 둔다(REQ-GR-001).

## §3. zone-registry 대조

명령: `grep -c "Kickoff" .claude/rules/moai/core/zone-registry.md` → `0`. **G1 미등록 확인.** zone-registry 는 로컬·템플릿 동일.

| 게이트 | 항목 | zone | canary |
|---|---|---|---|
| G2 | `CONST-V3R2-013` (CLAUDE.md `#7-safe-development-protocol`, "When intent is unclear, conduct a Socratic interview before execution") | Evolvable | false |
| G3 | `CONST-V3R2-014` (같은 앵커, "Before non-trivial code, explain the approach + which files change + why; get user approval") | Evolvable | false |
| G4 | `CONST-V3R2-030` (moai-constitution.md `#agent-core-behaviors`, "…list assumptions explicitly and wait for user confirmation") | Evolvable | false |
| (참고) | `CONST-V3R2-027` (moai-constitution.md, askuser-protocol § Socratic Interview Structure 를 **정본으로 가리킴**) | **Frozen** | true |

귀결:

- G2/G3/G4 는 Evolvable 등록 조항이다. `moai constitution validate` 의 drift 검사는 조항 문자열이 원본 파일에 남아 있는지를 본다 — 추가형 개정은 조항 문자열을 지우지 않으므로 drift 를 만들지 않는다. 기준선: `moai constitution validate` → exit 0, `OK — no drift or violations detected (97 of 101 entries checked)`, 은퇴 4건.
- `CONST-V3R2-027` 은 Frozen 이며 askuser-protocol 의 `§ Socratic Interview Structure` 제목을 가리킨다. **그 제목은 바꾸지 않는다.** contract 블록은 `§ Ambiguity Triggers and Exceptions` 쪽에 둔다.
- G4 원본은 `moai-constitution.md` 에 있다. 이 파일을 고치면 헌법 개정 범위(G17/G18)와 겹치므로 편집 대상에서 제외하고, G4 처분은 SSOT 등가 조항으로 표현한다(`plan.md §B` NC-3).
- `orchestration-mode-selection.md` 머리 문단은 인라인 `[ZONE:Frozen] [HARD]` 태그를 달고 있으나 zone-registry 에 없다. 운영자 결정 A-Q1 이 이 게이트의 전환을 승인했다. 태그 문단은 바이트 그대로 두고 contract 블록이 등가를 선언한다 — 문단 본문의 개정이 아니다. 이 해석의 잔여 위험은 `design.md §6`.

## §4. t1175 (SPEC-ALWAYS-LOADED-DIET-002) 겹침

- 이 워크트리에는 `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/` 가 **없다.** SPEC 은 브랜치 `WT-rules-diet` 에 있다(`git show WT-rules-diet:.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md`). 최신 커밋 `c1727f93d`, frontmatter `status: in-progress`, `version: "0.9.0"`, `tier: L`.
- 범위(그 spec.md §A~§D 발췌): always-loaded 18파일(규칙 14 + `CLAUDE.md` + `AGENTS.md` + 설정 2) 합계 150,000자 미만, 규칙 파일당 40,000자 미만. 비구속 산문을 path-scoped companion 으로 옮기는 재배치(M1)와 제자리 압축(M2). 로컬·템플릿 사본 동시 미러. **출력 스타일은 범위 밖.** 이미 40,000자를 넘긴 4개(`worktree-integration.md`·`session-handoff-examples.md`·`kanban-dispatch-detail.md`·`spec-workflow.md`)는 수리하지 않고 악화 금지.
- 겹치는 파일(이 SPEC 편집 대상 중 always-loaded): `CLAUDE.md`, `askuser-protocol.md`, `goal-directive.md`. t1175 가 이 파일들의 문단을 companion 으로 옮기면 앵커 위치가 바뀐다 → **절 제목 기준으로 앵커를 잡고**, t1175 착지 후 develop 을 흡수해 다시 잡는다(`plan.md` M0).
- 현재 크기(`LC_ALL=en_US.UTF-8 wc -m`): `CLAUDE.md` 19,367 · `askuser-protocol.md` 24,466 · `goal-directive.md` 6,875 · 참고로 `kanban-dispatch.md` 39,077(편집하지 않음). 추가 문자는 t1175 의 감축 목표를 거스른다 → 예산 상한(REQ-GR-080, `design.md §4`).
- path-scoped 인 `orchestration-mode-selection.md`(37,140자)는 always-loaded 가 아니나 40,000 파일당 한도가 걸리는지 여부는 t1175 가 「규칙 파일당」을 always-loaded 로 한정하는지에 달렸다. 안전하게 40,000자 미만 유지를 요구한다.

## §5. 템플릿 중립성

`.moai/docs/template-internal-isolation-doctrine.md §25.1` 금지 클래스: 내부 SPEC ID, REQ 토큰, AC 토큰, 감사 인용, 내부 날짜, 내부 아카이브 경로, 내부 커밋 SHA, 내부 메모리 경로. CI 가드 `.github/workflows/template-neutrality-check.yaml` 존재 확인. contract 블록은 로컬·템플릿 바이트 동일이어야 하므로(REQ-GR-070) **로컬 블록도 중립이어야 한다.** 게이트는 G 번호가 아니라 이름으로 부른다(G 번호는 이 설계 문서의 내부 번호다).

## §6. 이웃 개념과의 이름 충돌

| 기존 용어 | 위치 | 이 SPEC 과의 관계 |
|---|---|---|
| 「pipeline contract」(`full-pipeline`/`single-phase`/`factory`) | `workflows/moai.md`, `sync/delivery.md`, `run/mode-orchestration.md` | 실행 형태 구분. 서명 계약과 다른 개념 → 블록에서는 항상 「signed contract (`contract.yaml`)」라고 쓴다 |
| 「Behavioral Contract (SEMAP)」 | `manager-develop.md`, `internal/template/contract_schema_test.go` | 에이전트 섹션 스키마. 무관 |
| `MOAI_AUTONOMY_TIER` (semi-auto/automatic/fully-autonomous) | `SPEC-AUTONOMY-TIERS-001` (completed) | 권한 번들 축. `workflow.autonomy.mode` 와 다른 축 → SSOT 에 두 축의 구분을 한 줄로 적는다 |
| `/moai goal --auto` 승인형 자율 미션 | `SKILL.md` goal 항목 | 목표 루프 축. 계약 서명이 goal 을 무장하는지 여부는 NC-4 |

## §7. A1 기준선 — 계약 스키마 초안 `8f77d9a33`

**기준선: 커밋 `8f77d9a33`** (`docs(SPEC-AUTONOMY-CONTRACT-001): plan-phase draft for contract schema and signing (t1234)`), 파일 `.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md` § Contract Schema. plan-audit **이전** 초안이다. 형제 워크트리 `.claude/worktrees/t1234` 의 파일을 읽기 전용으로 읽었고, `git show 8f77d9a33:<path> | cmp -s - <워크트리 파일>` 로 두 바이트가 같음을 확인했다(출력 `draft file == commit 8f77d9a33`).

이 SPEC 이 가져다 쓴 이름:

| 항목 | 초안의 값 | 이 SPEC 에서 쓰는 곳 |
|---|---|---|
| 파일 | `.moai/specs/<SPEC-ID>/contract.yaml` | REQ-GR-010·015 |
| 필드 | `acceptance{file,sha256,ac_count}`, `invariants`, `ownership{write,never}`, `approach`, `actions`, `reobserve`, `review{second_model,human}`, `budget{turns,operations,audit_retries}`, `escalate_on`, `plan_audit{verdict,score,report}`, `signature{…}` | REQ-GR-020·022·030·062 |
| 행동 어휘 | 허용 `commit`·`worktree`·`local-merge-develop`·`push-develop`(설정 `push_develop: true` 일 때, 창 요건 함의) / 금지 `push-main`·`merge-main`·`force-push`·`release-branch`·`release-pr` | REQ-GR-050·063 |
| `escalate_on` 6종 | `acceptance-change`, `invariant-violation`, `ownership-move`, `new-architecture-or-api`, `contradictory-evidence`, `irreversible-action` | REQ-GR-022 |
| CLI | `moai contract sign <ID…>` (대화형 터미널 필수, `--resign`, 배치), `show [--json]`, `verify [--json]`; verify exit 0 유효 / 1 무효(reasons) / 2 사용법·I/O | REQ-GR-010·013 |
| verify 상태 | `state ∈ unsigned \| signed-valid \| signed-invalid`, `reasons[]` 폐쇄 집합 18종 | REQ-GR-010 |
| 설정 | `workflow.autonomy.mode: guided\|contract` (무효값 → guided + 경고), `workflow.autonomy.contract.{batch_sign, second_review: required\|advisory\|off, push_develop}`, `workflow.autonomy.escalation.budget_default{turns, operations, audit_retries: 2}` | REQ-GR-003·030·050·063 |
| 서명 전제 | `plan_audit.verdict` 가 `PASS`/`PASS-WITH-DEBT` 여야 서명 가능(FAIL 이면 거절) | REQ-GR-010·017·030 (plan-audit → 서명 순서) |
| guided 중립 | A1 REQ-CONTRACT-019 — guided 에서는 서명해도 Kickoff 를 대체하지 않는다는 안내를 출력 | REQ-GR-003 과 정합 |
| A3 로 넘긴 일 | A1 spec.md § Out of Scope — Gate rewiring (A3): 「plan phase 에서 manager-spec 이 초안 `contract.yaml` 을 내는 것은 A3 의 문서 변경」 | REQ-GR-015 (신규) |

이 초안에서 드러나 배차문 요약을 고친 점:

1. **plan-audit 가 서명보다 먼저다.** 서명은 `plan_audit.verdict` FAIL 을 거절한다. 따라서 G7/G8 자동 수리 상한은 서명된 계약에서 올 수 없고, 초안 `contract.yaml` 의 `budget.audit_retries` 또는 설정 `budget_default` 에서 온다(REQ-GR-030).
2. **서명은 오케스트레이터가 할 수 없다.** `interactive-tty` 가 유일한 v1 방식이다. 오케스트레이터는 서명 명령을 보고하고 턴을 닫는다(REQ-GR-010). v0.2.0 에서 이 발견은 자율 Kickoff 와 충돌로 이어졌다 — §9.1.
3. **두 번째 리뷰 설정 키는** `workflow.autonomy.contract.second_review` 이고 값은 `required|advisory|off` 다.
4. **main/release 는 설정이 아니라 스키마가 금지한다** — 계약으로 허용될 수 없다(REQ-GR-050).

**후속 커밋 `98cb7879d`** (`docs(SPEC-AUTONOMY-CONTRACT-001): resolve plan-audit iter-1 defects D1-D18 (t1234)`, spec v0.2.0) 을 `git show 98cb7879d:<path>` 로 추가로 읽었다. v0.2.0 에서 이 SPEC 에 닿는 것:

- **서명은 사람 전용으로 강화됐다.** Signing Flow 1·2단계: 에이전트 환경 표지가 있으면 거절, 표준 입력이 터미널이 아니면 거절(`interactive-tty` 가 유일한 v1 방식). REQ-CONTRACT-010(대화형 서명), REQ-CONTRACT-021(에이전트 환경 거절).
- **§C.2 「A3 의 강한 전제」**: 서명이 Kickoff 를 대체하기 전에 에이전트 도구 호출이 낸 `moai contract sign` 을 막는 PreToolUse 거부가 (A2 또는 A3 에) 있어야 한다.
- **§C.1 구속 순서 제약**: A2 가 push 직렬화(A-Q2)와 두 번째 리뷰 정지(A-Q3)를 강제하기 전까지 A3 는 「서명이 Kickoff 를 대체」를 활성화하지 않는다.
- revoke 는 v0.2.0 본문에 없다(`grep -niE revoke` 0건) — 리드 전달대로 이 카드로 넘어왔다.
- Package Layout: `internal/contract/`(순수 코어, `os/exec`·`net` 없음), `internal/contract/sign/`, `internal/cli/contract.go`(Cobra `contract` — `sign`·`show`·`verify`).

## §8. 검증하지 못한 것 (Gaps)

- 운영자 승인 설계 문서의 원문은 디스크에서 찾지 못했다(`grep -rlE "A-Q1|audit_retries|escalate_on"` 가 `.moai/reports`·`.moai/docs`·primary 체크아웃 `.moai/reports` 에서 0건). 게이트 표는 리드 배차문의 요약을 따르고, 결정 A-Q1~A-Q4 는 A1 초안 spec.md §C.1 의 기록과 대조했다.
- A1 은 plan-audit 전 초안이다. 필드명·CLI 동사·설정 키·exit 코드는 감사 뒤 바뀔 수 있으므로, 그것에 기대는 REQ·AC 에 **[A1 감사 통과본으로 재확인]** 을 달았다(`plan.md` M0 가 재대조한다).
- A2 도 plan 초안이다(`ee9f57151`, v0.1.2). 에스컬레이션 기록의 경로·형식·종류는 **[A2 스키마 재확인]** 으로 표시했다.
- `moai contract decide`·`revoke` 는 아직 없다: `moai contract revoke --help` 가 `ERROR` 를 출력했고(`ca1e39f6f`), `internal/contract` 디렉터리가 없다(`ls: internal/contract: No such file or directory`).

## §9. v0.2.0 추가 조사 — 충돌과 의존

### §9.1 A1 과의 충돌: 서명 주체

v0.1.0 은 「서명은 대화형 터미널의 운영자 행위」로 적었고, A1 v0.2.0 은 그것을 더 굳혔다(§7 후속 커밋). 운영자 결정 「Kickoff 자율 승인」은 contract 모드의 서명을 **결정자 쌍**이 하게 한다. 둘은 같은 `sign` 경로로 공존할 수 없다:

| A1 v0.2.0 | 자율 Kickoff 가 요구하는 것 |
|---|---|
| 표준 입력이 터미널이 아니면 서명 거절 | 비대화형 서명 |
| 에이전트 환경 표지가 있으면 거절 | 리드 세션(에이전트)이 서명 절차를 개시 |
| §C.2: 에이전트발 `sign` 을 PreToolUse 로 거부해야 함 | 에이전트가 서명 절차에 관여 |

**해소 방향(A1 에 요청 — 리드가 전달):** 대화형 `sign` 은 그대로 사람 전용으로 두고, 그와 **다른** 영수증 기반 경로(예: `sign --receipt <id>`)를 A1 이 제공한다. 그 경로는 이 SPEC 의 저장소(§8.3)에 있는 `outcome: approve` 기록만 받고, 기록의 입력 해시가 현재 `contract.yaml`·`acceptance.md` 와 맞지 않으면(낡은 영수증) 거절하며, 같은 카드에 그 뒤 revoke 기록이 있으면 거절한다. A2 의 에이전트발 `sign` 차단은 대화형 경로에만 걸리고 영수증 경로는 막지 않아야 한다. 이 SPEC 은 A1·A2 의 내부를 정하지 않는다 **[A1 감사 통과본으로 재확인]** **[A2 스키마 재확인]**.

### §9.2 `SPEC-JEV-CORE-001` 과의 충돌: Jev 를 게이트 입력으로 쓰는 것

`SPEC-JEV-CORE-001` 은 `status: completed`, v0.2.0 이다.

- **REQ-JEVC-012**(96행): 「The system shall not consult Jev — not as a decision, and **not as an input** — for any of: a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, **an operator gate**, …」
- **REQ-JEVC-011**(80행): Jev 답은 파일·큐·카드 상태를 바꾸지 않는다, display-only.
- 같은 금지가 규칙 문서에도 있다: `moai-mcp-tools.md` 의 `jev_ask` 행 「never a completion predicate, merge approval, queue mutation, or gate input」, `internal/cli/mcp_jev.go` 머리 주석.
- 로컬 전용 `CLAUDE.local.md §29` 도 운영자 게이트를 3등급(Jev 를 입력으로도 쓰지 않음)으로 둔다.

Kickoff 는 운영자 게이트다. 운영자가 승인한 설계(Jev 가 두 번째 신호)는 완료된 SPEC 의 구속 조항과 정면으로 충돌한다. 이 SPEC 이 그 조항을 조용히 무시할 수 없으므로 NC-7 로 올리고, 개정 착지를 자율 Kickoff 활성 조건에 넣었다(REQ-GR-016 (iv)). 반면 재사용할 부분은 그대로 쓸 수 있다: `internal/jev.Client.Ask`, 사용 불가를 값으로 돌려주는 `Availability`(`disabled`·`no-credential`·`unreachable` 등 10종, `internal/jev/jev.go:104-129`), REQ-JEVC-007(소비자는 exit 0 으로 계속)은 decide 의 exit 설계와 맞는다.

### §9.3 `mission-governor` 재사용 충돌

`.claude/agents/moai/mission-governor.md` frontmatter: 「NOT for: writing files, shell or Git execution, queue mutation, dispatch, merge, **approval**, or PASS/FAIL audit verdicts」. 판단 객체를 돌려주고 적용은 결정적 실행기가 한다는 출력 모양은 decide 와 맞지만, 설명문이 승인을 명시적으로 배제한다. 에이전트 파일은 이 SPEC 의 편집 범위 밖이다 → NC-8.

### §9.4 A2 와 needs-decision

A2 초안(`ee9f57151`) spec.md §B.2-1: 큐 상태는 `queued`/`picked`/`dropped` 셋(`internal/kanban/backlog_store.go:61-65`)이고 `backlog_schema_freeze_test.go` 가 네 번째 상태를 금지한다. needs-decision 은 큐 옆의 기록이다(REQ-AE-015). 기록 위치는 A2 design.md §A 가 `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + 색인 줄로 제안하고 Q1 로 열어 두었다. 리드는 `.moai/reports/<card>/escalations/<ts>.json` 이라고 전달했다 — 두 형태가 다르다. 이 SPEC 은 A2 감사 통과본을 따른다 **[A2 스키마 재확인]**. 또 A2 의 9개 종류에 revoke 가 없다.

### §9.5 저장소 위치의 근거가 된 코드

`internal/homestate/paths.go`: `ProjectKey`(22행)는 모든 링크 워크트리가 공통 디렉터리를 통해 같은 키를 받는다고 명시하고, `BacklogDBPath`·`FactoryDBPath`(213·229행)가 같은 방식으로 `~/.moai` 아래 경로를 준다. `explicitMoaiHome`(192행)이 있어 테스트는 격리된 홈을 쓸 수 있다.
- Claude Code 가 `CLAUDE.md` 의 HTML 주석을 모델 컨텍스트에서 제거하는지 여부는 확인하지 않았다. 마커는 주석이고 블록 본문은 주석이 아니므로 어느 쪽이든 지시문은 전달되지만, 마커 문자 수가 예산에 잡히는지는 측정 대상이다.
- 계수는 t1175 착지 전 트리 기준이다.
