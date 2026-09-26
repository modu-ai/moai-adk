# research.md — SPEC-AUTONOMY-GATE-REWIRE-001 (v0.3.0)

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

- **발화 지점(E)**: 게이트를 정의하거나 `AskUserQuestion` 발화를 지시한다 → contract-mode 블록 필수(REQ-GR-005).
- **참조 지점(R)**: 게이트가 「이미 통과했다」를 전제로 삼거나 「바뀌지 않는다」고 말한다 → SSOT 등가 조항(REQ-GR-005, 사람 서명 한정)이 덮는다, 편집 없음.
- **동음이의(H)**: 다른 게이트가 같은 이름을 쓴다 → 범위 밖.

| # | 파일 (로컬 경로; 템플릿은 `internal/template/templates/` 접두) | 줄 | 분류 | 앵커(절 제목) | 로컬↔템플릿 diff줄 |
|---|---|---|---|---|---|
| 1 | `CLAUDE.md` | 1 | **E** | § 2. Request Processing Pipeline — ④ | 0 |
| 2 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 14 | **E** (SSOT) | 파일 머리 (제목 아래 인용 블록 + `[ZONE:Frozen] [HARD]` 문단), § §E — Anti-Patterns | 20 (§C.1 만) |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | 2 | **E** | § Recommendation mode, § Report-Before-Ask Gate › Exceptions | 0 |
| 4 | `.claude/rules/moai/workflow/goal-directive.md` | 3 | **E** | § Goal-Presentation Timing, § Hard Preconditions | 0 |
| 5 | `.claude/skills/moai/SKILL.md` | 2 | **E** | goal 라우팅 항목 (Progression mode 줄), completion-condition 절 | 41 |
| 6 | `.claude/skills/moai/workflows/moai.md` | 8 | **E** | § Pipeline Gates (named, in order) #2, Step 11.3 | 8 |
| 7 | `.claude/skills/moai/workflows/plan.md` | 2 | **E** | § NEEDS-CLARIFICATION Marker Usage 절 | 9 |
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
- G4 원본은 `moai-constitution.md` 에 있다. 이 파일을 고치면 헌법 개정 범위(G17/G18)와 겹치므로 편집 대상에서 제외하고, G4 처분은 SSOT 등가 조항으로 표현한다(본문 결정 D-3).
- `orchestration-mode-selection.md` 머리 문단은 인라인 `[ZONE:Frozen] [HARD]` 태그를 달고 있으나 zone-registry 에 없다. 운영자 결정 A-Q1 이 이 게이트의 전환을 승인했다. 태그 문단은 바이트 그대로 두고 contract 블록이 등가를 선언한다 — 문단 본문의 개정이 아니다. 이 해석의 잔여 위험은 `design.md §6`.

## §4. t1175 (SPEC-ALWAYS-LOADED-DIET-002) 겹침

- 이 워크트리에는 `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/` 가 **없다.** SPEC 은 브랜치 `WT-rules-diet` 에 있다(`git show WT-rules-diet:.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md`). 최신 커밋 `c1727f93d`, frontmatter `status: in-progress`, `version: "0.9.0"`, `tier: L`.
- 범위(그 spec.md §A~§D 발췌): always-loaded 18파일(규칙 14 + `CLAUDE.md` + `AGENTS.md` + 설정 2) 합계 150,000자 미만, 규칙 파일당 40,000자 미만. 비구속 산문을 path-scoped companion 으로 옮기는 재배치(M1)와 제자리 압축(M2). 로컬·템플릿 사본 동시 미러. **출력 스타일은 범위 밖.** 이미 40,000자를 넘긴 4개(`worktree-integration.md`·`session-handoff-examples.md`·`kanban-dispatch-detail.md`·`spec-workflow.md`)는 수리하지 않고 악화 금지.
- 겹치는 파일(이 SPEC 편집 대상 중 always-loaded): `CLAUDE.md`, `askuser-protocol.md`, `goal-directive.md`. t1175 가 이 파일들의 문단을 companion 으로 옮기면 앵커 위치가 바뀐다 → **절 제목 기준으로 앵커를 잡고**, t1175 착지 후 develop 을 흡수해 다시 잡는다(`plan.md` M0).
- 현재 크기(`LC_ALL=en_US.UTF-8 wc -m`): `CLAUDE.md` 19,367 · `askuser-protocol.md` 24,466 · `goal-directive.md` 6,875 · 참고로 `kanban-dispatch.md` 39,077(편집하지 않음). 추가 문자는 t1175 의 감축 목표를 거스른다 → 예산 상한(REQ-GR-002·025, `design.md §4`).
- path-scoped 인 `orchestration-mode-selection.md`(37,140자)는 always-loaded 가 아니나 40,000 파일당 한도가 걸리는지 여부는 t1175 가 「규칙 파일당」을 always-loaded 로 한정하는지에 달렸다. 안전하게 40,000자 미만 유지를 요구한다.

## §5. 템플릿 중립성

`.moai/docs/template-internal-isolation-doctrine.md §25.1` 금지 클래스: 내부 SPEC ID, REQ 토큰, AC 토큰, 감사 인용, 내부 날짜, 내부 아카이브 경로, 내부 커밋 SHA, 내부 메모리 경로. CI 가드 `.github/workflows/template-neutrality-check.yaml` 존재 확인. contract 블록은 로컬·템플릿 바이트 동일이어야 하므로(REQ-GR-024) **로컬 블록도 중립이어야 한다.** 게이트는 G 번호가 아니라 이름으로 부른다(G 번호는 이 설계 문서의 내부 번호다).

## §6. 이웃 개념과의 이름 충돌

| 기존 용어 | 위치 | 이 SPEC 과의 관계 |
|---|---|---|
| 「pipeline contract」(`full-pipeline`/`single-phase`/`factory`) | `workflows/moai.md`, `sync/delivery.md`, `run/mode-orchestration.md` | 실행 형태 구분. 서명 계약과 다른 개념 → 블록에서는 항상 「signed contract (`contract.yaml`)」라고 쓴다 |
| 「Behavioral Contract (SEMAP)」 | `manager-develop.md`, `internal/template/contract_schema_test.go` | 에이전트 섹션 스키마. 무관 |
| `MOAI_AUTONOMY_TIER` (semi-auto/automatic/fully-autonomous) | `SPEC-AUTONOMY-TIERS-001` (completed) | 권한 번들 축. `workflow.autonomy.mode` 와 다른 축 → SSOT 에 두 축의 구분을 한 줄로 적는다 |
| `/moai goal --auto` 승인형 자율 미션 | `SKILL.md` goal 항목 | 목표 루프 축. 계약 서명이 goal 을 무장하는지 여부는 본문 결정 D-4(자동 무장 없음) |

## §7. A1 기준선 — 0.4.1 `6d98ca466`

**기준선: A1 0.4.1, 커밋 `6d98ca466`** (`docs(SPEC-AUTONOMY-CONTRACT-001): iter-4 text fixes V1-V3, lane-verified by operator decision (t1234)`). `git show 6d98ca466:<path>` 로 스크래치에 뽑아 형제 워크트리 `.claude/worktrees/t1234` 의 파일과 `cmp` 했고 spec·design 모두 같았다(출력 `wt spec == 6d98ca466`, `wt design == 6d98ca466`). design.md 는 `652243c72` 와 바이트 동일하고, spec.md 는 0.4.1 HISTORY 행과 §C.6·§H·REQ-015 문장만 다르다(`diff` 34줄).

**이력(참고만)**: `8f77d9a33`(0.1.0 초안, v0.1.0 이 처음 맞춘 기준), `98cb7879d`(0.2.0), `4208a3a3b`(0.3.0, 감사 iter-1 이 읽은 판), `652243c72`(0.4.0). 판단의 근거는 0.4.1 이다.

0.4.1 이 확정한 항목 — 이 SPEC 은 재확인 표지 대신 「A1 0.4.1 (6d98ca466)」으로 인용한다:

| 항목 | A1 0.4.1 의 값 | 이 SPEC |
|---|---|---|
| 계약 파일·필드 | design § Contract Schema — `acceptance`, `invariants`, `ownership{write,never,scratch}`, `approach`, `actions`, `reobserve`, `review`, `budget{turns,operations,audit_retries}`, `escalate_on`(6종), `plan_audit.verdict`, `signature{signer_kind, operator, signed_at, head_sha, contract_sha256, acceptance_sha256, method, receipt{path,sha256,provenance}, batch_id, supersedes, seal}` | REQ-GR-005·006·014·015·020 |
| 서명 경로 | 사람: 대화형 터미널, 에이전트 환경 표지 거절(REQ-CONTRACT-010·021). 영수증: `sign --signer <llm\|jev\|llm+jev> --receipt <path>`, 비대화형, `mode: contract` 에서만, 출처 `file` 기록(REQ-CONTRACT-024) | REQ-GR-004·007 |
| 영수증 형식 | `.moai/specs/<SPEC-ID>/kickoff-receipt.json` 고정 경로, `inputs{contract_sha256, acceptance_sha256, plan_audit_report{path,sha256}}`, 결정자별 `decisions[]`, Jev 항목의 `request_sha256`·`raw_response` (design § Kickoff Receipt) | REQ-GR-009·011 |
| 합의 규칙 | design § Agreement Rule 0~4 — 0: Jev 결정은 A3 가 원칙을 개정할 때까지 측정되지 않음, 1: 에스컬레이션 → 사람, 2: Jev 신뢰도 미달 → 사람, 3: 거절 → `on_disagree`, 4: 전원 승인 | REQ-GR-010 |
| 결정자 설정 | `workflow.autonomy.kickoff.{decider: human\|llm\|jev\|llm+jev, jev_min_confidence: 0.50, on_disagree: human\|reject}`; `jev`/`llm+jev` + `workflow.jev.enabled: false` 는 설정 오류이며 영수증 서명만 거절(REQ-CONTRACT-015) | REQ-GR-003·004·010 |
| `frozen-files` | design § Frozen Files — zone-registry Frozen 항목의 파일 + `**/CLAUDE.md`·`**/CLAUDE.local.md` + 계약 `ownership.never`, 모두 glob | REQ-GR-009 (f) |
| verify | 상태 `unsigned \| signed-valid \| signed-invalid`, 폐쇄 사유 코드, exit 0/1/2 | REQ-GR-007 |
| push 리스 | `push-develop` 은 `moai slot` 리스 `push-develop` 을 잡은 상태의 push, `push_requires_lease: true` 파생 필드(REQ-CONTRACT-018) | REQ-GR-021 |
| `show --json` | `mode` 는 있으나 결정자 필드는 **없음** | D-1: kickoff-check 가 채움 |

0.4.1 이 A3 에 **남긴** 항목 — 이 SPEC 이 소유한다:

- 저장소 대조로 영수증을 거절하는 일(§C.6 「so that A3 can refuse it」) → REQ-GR-007.
- moai 가 Jev 를 직접 부르고 저장소에 발급하는 일(§B Out of Scope — Revocation and receipt issuance) → REQ-GR-011·012.
- **모든 서명 사건 기록**(§C.6·§H 「that store shall record every signing event, including human signatures」) → REQ-GR-012.
- 합의 규칙 0번 해제와 표시 전용 원칙 개정(§C.8, 세 위치: 템플릿 `workflow.yaml:171-173`, `moai-mcp-tools.md:75`, `moai-mcp-tools-catalogue.md:138`) → REQ-GR-013.
- 비인간 서명의 자율 Kickoff 활성 조건(§C.6: revoke 존재 + moai 발급 영수증) → REQ-GR-008.
- plan-audit 판정의 증거 결합(§C.7, design § Forward Note — Verdict Binding) → REQ-GR-009 (a).

## §8. 검증하지 못한 것 (Gaps)

- 운영자 승인 설계 문서 원문은 디스크에서 찾지 못했다. 게이트 표는 리드 배차문 요약을 따르고 결정 A-Q1~A-Q4 는 A1 0.4.1 §C.1 과 대조했다.
- A2 의 **개정본**은 아직 없다. 에스컬레이션 기록 형식은 리드가 전달한 최종 형식이며 **[A2 개정본으로 재확인]** 으로 표시했다(§10.1).
- A4(t1237)의 SPEC 은 읽지 않았다. 두 번째 리뷰 정지 형태는 **[A4 감사 통과본으로 재확인]** 으로 표시했다.
- `moai contract decide`·`kickoff-check`·`revoke` 는 아직 없다: `moai contract revoke --help` → `ERROR Unknown command "contract"`, exit 1(`ca1e39f6f`); `ls internal/contract` → 디렉터리 없음, exit 1.
- Claude Code 가 `CLAUDE.md` 의 HTML 주석을 모델 문맥에서 제거하는지는 확인하지 않았다 — 마커 문자가 AC-GR-010 예산에 잡히는지는 run phase 측정 대상이다.
- 계수는 t1175 착지 전 트리 기준이다.

## §9. 충돌과 그 처분

### §9.1 A1 과의 충돌: 서명 주체 — 해소됨

v0.1.0 은 서명을 대화형 터미널의 운영자 행위로 적었다. 자율 Kickoff 는 결정자 쌍의 서명을 요구한다. A1 0.4.1 은 두 경로를 분리해 해소했다: 사람 경로는 대화형·에이전트 거절 그대로, 영수증 경로는 비대화형이고 에이전트 표지를 보지 않는다(A1 design § Agent-Environment Markers). A2b 의 에이전트발 `sign` 차단도 A1 문언상 사람 경로에 대한 것이다(§C.2). 남은 일은 영수증의 **출처**를 믿을 수 없다는 점이며, A1 이 그 거절을 A3 에 맡겼으므로 kickoff-check 가 저장소 대조로 판정한다(REQ-GR-007).

### §9.2 `SPEC-JEV-CORE-001` 과의 충돌 — 리드 결정 L1 로 범위 편입

`SPEC-JEV-CORE-001`(`status: completed`, v0.2.0) REQ-JEVC-012(96행) 「not as a decision, and **not as an input** — … **an operator gate**」, REQ-JEVC-011(80행) 표시 전용, `moai-mcp-tools.md:75` 「never … gate input」, `moai-mcp-tools-catalogue.md:138`, 템플릿 `workflow.yaml:171-173` 주석, 로컬 `CLAUDE.local.md §29` 3등급. Kickoff 는 운영자 게이트다. 리드 결정 L1(2026-09-26): contract 모드 Kickoff 교차 확인 한 곳에 한정한 개정을 이 카드에서 한다(REQ-GR-013, `design.md §11`). 개정 방식은 그 SPEC 의 v0.2.0 선례 — `completed` 상태를 유지한 채 개정 표지와 HISTORY 행을 다는 「clarifying amendment to a completed SPEC」 — 를 따른다. 재사용: `internal/jev.Client.Ask`, `Availability`(`internal/jev/jev.go:104-129`), REQ-JEVC-007(소비자는 exit 0 으로 계속) — decide 의 exit 설계와 맞는다.

### §9.3 `mission-governor` — 리드 결정 L3 로 해소

`.claude/agents/moai/mission-governor.md` 「NOT for: … **approval** …」. 결정자로 쓰지 않는다(D-8).

### §9.4 needs-decision — 리드 결정으로 해소

큐 상태는 `queued`/`picked`/`dropped` 셋이고 `internal/kanban/backlog_schema_freeze_test.go` 가 네 번째 상태를 금지한다(A2 초안 §B.2-1). needs-decision 은 A2 형식의 열린 에스컬레이션 기록이다.

### §9.5 저장소 위치의 근거가 된 코드

`internal/homestate/paths.go`: `ProjectKey`(22행)는 링크 워크트리가 같은 키를 받게 하고, `BacklogDBPath`·`FactoryDBPath`(213·229행)가 같은 방식으로 `~/.moai` 아래 경로를 준다. `explicitMoaiHome`(192행)이 테스트용 격리 홈을 허용한다.

### §9.6 plan-audit 판정 파일의 실제 모양

`internal/runtime/audit_review.go` `ResolveLatestPlanAudit` 는 `<reportDir>/<SPEC-ID>-review-<N>.md` 를 찾고 `Verdict:`·`Overall Score:` 키 줄을 읽으며 점수를 `strconv.ParseFloat` 로 파싱한다. 그러나 이 카드의 감사 보고서는 `.moai/reports/t1236/plan-audit-1.md` 이고 점수 줄은 `Overall Score: 0.71 (Tier L PASS threshold 0.85)` 처럼 괄호 주석이 붙는다 — 그대로면 파싱 오류다. 그래서 REQ-GR-009 (a) 는 카드 증거 경로의 `plan-audit-<N>.md`·`plan-audit-iter<N>.md` 를 고르고, 파싱이 안 되면 전제조건 실패(→ 사람)로 처리하는 보수적 규칙을 택했다(`design.md §8`). 두 보고서 명명 규칙의 통일은 이 SPEC 범위 밖이다.

## §10. v0.3.0 추가 조사

### §10.1 A2 에스컬레이션 기록 형식 — 리드가 정한 최종 형식을 인용

| 출처 | 형식 | 지위 |
|---|---|---|
| A2 `ee9f57151` design.md §A | `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + 색인 줄 | 옛 초안 |
| A2 `d8926ff9a`(현재 커밋) design.md:11 | `.moai/reports/<card-id>/escalations/<timestamp>.json` | **리드가 지시했다가 철회한 형식** — 알려진 낡은 초안 |
| 리드 전달 최종 형식(2026-09-26) | `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` + **기계 판독 YAML 머리** + `revoke` 종류 | A2 가 감사 창 뒤 다음 개정에서 반영 예정 |

이 SPEC 은 최종 형식을 인용하고 **[A2 개정본으로 재확인]** 을 단다. `d8926ff9a` 에는 `revoke` 종류가 없다(감사 D10 확인).

### §10.2 상충 보고 — push 직렬화와 에이전트발 `sign` 차단의 소유 카드

| 출처 | 소유 카드 |
|---|---|
| A1 0.4.1 `6d98ca466` §B Out of Scope, §C.1, §C.2 | **A2b (t1245)** — 「A2b (t1245) enforces serialization」, 「A2b (t1245) owns it」 |
| A2 `d8926ff9a` spec.md REQ-AE-024·025, design §G.1·G.2 | **A2 (t1235)** — 「two A3 preconditions assigned by lead」 |
| 리드 4번째 지시(2026-09-26) | **A2 (t1235)** — run 전제 |

세 출처가 일치하지 않는다. 이 SPEC 은 어느 쪽도 고르지 않았다: 활성 조건(`design.md §7.1` 4행)은 **기능**(두 가드의 착지)으로 적고 소유 카드는 이 상충을 가리키며, run 전제는 리드 지시대로 t1234·t1235 병합으로 둔다. 리드가 소유 카드를 확정하면 §7.1 4행과 `plan.md §C` 를 그에 맞춘다.

### §10.3 A1 0.4.1 과 A3 요구 사이의 다른 상충 — 없음

대조한 항목: 서명 경로 분리, 영수증 경로와 형식, 결정자 설정 키와 기본값, `frozen-files`, 합의 규칙(0번 해제는 A3 소관으로 남김), verify 상태와 exit 코드, 모든 서명 사건 기록 요구. A3 가 `llm`·`jev` 단독 서명을 Kickoff 대체에서 거절하는 것은 A1 이 그 서명을 허용하는 것과 모순이 아니다 — A1 은 서명을 만들고, A3 는 그것이 Kickoff 를 대체하는지를 판정한다(A1 §C.6 이 그 판정을 A3 에 둔다). A1 이 「In A1 only an `llm` receipt can be accepted」라 하므로 `llm` 단독 서명은 실제로 생길 수 있고, kickoff-check 의 `decider-not-permitted` 가 그것을 막는다.
