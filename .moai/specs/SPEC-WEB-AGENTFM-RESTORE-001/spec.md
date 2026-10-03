---
id: SPEC-WEB-AGENTFM-RESTORE-001
title: "moai web 서브 에이전트 설정 표면 복원 — agentfm 오버라이드 · llm.profile · 저장 경로 · UI 노출"
version: "0.3.0"
status: in-progress
created: 2026-10-02
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/web, internal/settings/agentfm, internal/config, internal/template, internal/template/templates/.moai/config/sections"
lifecycle: spec-anchored
tags: "web-console, agentfm, agent-overrides, llm-profile, restore, i18n, t1246-reversal"
tier: M
related_specs: [SPEC-AGENT-MODEL-INHERIT-001, SPEC-AGENT-TIER-001, SPEC-WEB-CONSOLE-011, SPEC-MODEL-PROFILE-MATRIX-002]
amendment_of: SPEC-WEB-AGENTFM-RESTORE-001
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-02 | manager-spec | 최초 작성 (카드 t1411, Class C, 운영자 지시 2026-10-02). 삭제 커밋 `384eb3460`(card t1246, SPEC-AGENT-MODEL-INHERIT-001 M2/REQ-AMI-011)의 역방향 복원 SPEC. 블루프린트: `.moai/reports/t1411/removed-commit-384eb3460.diff` (7,922행). |
| 0.2.0 | 2026-10-03 | manager-spec | **In-place amendment of the `completed` SPEC (card t1446 — t1411 sync-audit round-2 N1; manager-spec 재위임).** REQ-AFR-007 축소: 영속-오류 원자 복원을 llm.yaml agent-overrides write pair(`llm.profile` + `llm.agent_overrides`)로 한정. 검증-오류 원자 거절은 의미 불변. pair 밖 선행 단계는 문서화된 best-effort 동작 유지 + 한계를 본문에 명시. §D 원자성 제약 행 동기화. 요구사항 삭제 없음, id 재번호 없음, AC 매핑 무변경 (14 REQ / 13 AC); progress.md §E.4 `sync_commit_sha` 불변. 상태 축은 SSOT 수정 전이 `completed → in-progress`를 따르며 구조 기록은 `## Amendments`. |
| 0.3.0 | 2026-10-03 | manager-spec | **In-place amendment (card t1421 — decision-index Q2 후속 카드, 운영자 결정 2026-10-03 옵트인 고정).** 스폰-소비 계약 신설: REQ-AFR-015..020 / AC-AFR-014..019 — 옵트인 키 `llm.agent_overrides_consume`(기본 false)·소비 리졸버·교리 수비 의미론(룰 텍스트는 sync-phase)·훅 advise/audit 확장·console·doctor 계약 상태 가시성·AMI-001 수정안 링크. 기존 14 REQ·13 AC 무변경 — 삭제·재번호·AC 매핑 변경 없음. Out of Scope 런타임 소비 절 재기술(공백이 본 수정안으로 경계). REQ 20 / AC 19 — Tier M 상한 초과분의 예산 처분은 `## Amendments` v0.3.0 블록과 plan.md §I.3. 재close는 0.2.0 + 0.3.0 공동 close 커밋(카드 t1446 재진입 동행). |

## Amendments

**2026-10-03 — v0.2.0 — in-place amendment of the prior `completed` SPEC (card t1446).**

- Transition: `completed → in-progress` per the SSOT amendment contract (`.claude/rules/moai/development/spec-frontmatter-schema.md` § completed → in-progress (amendment)); `amendment_of: SPEC-WEB-AGENTFM-RESTORE-001` (self-referential). Authorized by the orchestrator's re-delegation to manager-spec (card t1446, N1 — the t1411 M6-repair F3 disposition had already scoped this body edit to manager-spec).
- Prior completed version: **0.1.0** — closed 2026-10-02 (card t1411 sync lane).
- `prior_completed_sha: bd51d75a17a39c7fd0d4437cf62a28ce4be677a3` — the prior close's `sync_commit_sha` (this SPEC's progress.md §E.4); that field is left unmodified. The last commit that touched this spec.md body before the amendment is `422524f5d` (2026-10-02 — card t1411 sync-audit F5/F7 dispositions).
- Rationale: t1411 sync-audit round-2 finding N1 — the prior REQ-AFR-007 wording ("any validation or persistence error … all section files remain byte-identical") was broader than the implemented repair. The implemented guarantee (internal/web/handlers.go, the llm.yaml write pair) is exactly the llm.yaml agent-overrides write pair: `SnapshotLLMYAML` before step 7 (`applyPerfTierEdits`, writes llm.profile), restored on a step-8 failure (`patchAgentFM`, writes llm.agent_overrides). Validation errors reject before any write (the merge gate precedes persistence), so validation-error atomicity is real and unchanged. Out-of-pair earlier-step partial persistence (e.g. a step-6 schema edit persisting after a step-8 rollback) is real and predates the repaired card — stated as a limitation rather than claimed as a guarantee.
- Scope — one requirement's wording narrowed; no requirement deleted, no id renumbered, no AC mapping changed:

  | # | Surface | Disposition |
  |---|---|---|
  | 1 | REQ-AFR-007 body | Amended here — persistence-error atomic restore scoped to the llm.yaml agent-overrides write pair (`llm.profile` + `llm.agent_overrides`); validation-error rejection unchanged; out-of-pair best-effort behavior + limitation stated |
  | 2 | §D 비기능 제약 원자성 행 | Amended here — mirrors the narrowed wording (the same claim stated at constraint level) |
  | 3 | acceptance.md AC-AFR-004 (REQ-AFR-006/007) | Unchanged — its scenarios are validation-error rejections, which keep full atomicity under the narrowed wording; no criterion pins the broad persistence-error claim |
  | 4 | progress.md | Dated amendment note appended at the tail; §E.2/§E.3/§E.4 evidence untouched |

**2026-10-03 — v0.3.0 — in-place amendment (card t1421 — decision-index Q2 후속 카드).**

- Transition: `in-progress` 유지 — v0.2.0 수정안이 이미 SSOT 수정 진행형이라 재전이 불요; `amendment_of: SPEC-WEB-AGENTFM-RESTORE-001`(자기참조) 불변. Authorized by the operator decision 2026-10-03(lane AskUserQuestion 라운드, 카드 t1421 — decision-index Q4): contract = **옵트인 고정** — `llm` 섹션의 단일 명시 키(기본 off)를 켠 세션만 서브에이전트 스폰에 `llm.agent_overrides`를 소비하고, 오케스트레이터가 Agent() 호출에 설정된 model/effort를 실어 보낸다. 키 없는 세션은 오늘의 상속 기본을 유지한다.
- Prior completed version: **0.1.0** — closed 2026-10-02 (card t1411 sync lane). `prior_completed_sha: bd51d75a17a39c7fd0d4437cf62a28ce4be677a3` 불변 — 0.2.0도 아직 재close 전이며, **0.2.0 + 0.3.0 두 수정안의 재close는 하나의 joint close 커밋**으로 수행한다(카드 t1446 재진입이 본 카드와 동행). 수정 전 spec.md 마지막 본문 커밋: `9c6056164` (2026-10-03 — card t1446 v0.2.0 amendment).
- Rationale: v0.1.0이 복원한 저장 표면의 스폰-경로 공백(decision-index Q2)을 메우는 후속 계약. 메커니즘 제약 실측: Claude Code의 서브에이전트 모델 해상은 스폰 시점 `model` → 프론트매터 `model` → `CLAUDE_CODE_SUBAGENT_MODEL` → 세션 모델 순서(`.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention)라, 오케스트레이터가 Agent() 호출에 model/effort를 전달하는 것이 유일한 승인 소비 경로다 — 훅 입력 재작성과 차단 레이어 부활은 없다(관측층 확장만).
- Scope — six new requirements + six new acceptance criteria; no requirement deleted, no id renumbered, no existing AC remapped:

  | # | Surface | Disposition |
  |---|---|---|
  | 1 | REQ-AFR-015 신설 | 옵트인 키 `llm.agent_overrides_consume`(bool, 기본 false) — 비-불리언 값은 기존 원자 거절 합류, 템플릿 `llm.yaml` 수재 + shipped-key 인벤토리 등록(REQ-AFR-008의 seam 연장) |
  | 2 | REQ-AFR-016 신설 | 소비 리졸버 — consume on 시 override 항목 승리·부재 항목 상속·`inherit` no-op; off 시 저장 전용 (블루프린트 형태 `3fa8bd2ab^`, 현 트리 M1 재포트 해상 기계 재사용) |
  | 3 | REQ-AFR-017 신설 | 교리 수비 소비 의미론 — 스폰 전 조회·Agent() 호출 파라미터 전달; 룰 텍스트 소비 조항 자체는 sync-phase(REQ-AMD-001 계열) |
  | 4 | REQ-AFR-018 신설 | 훅 advise/audit 확장 — override 적중/미적중 필드, observe-무차단 보존 (비교 축은 모델 — effort는 스폰 페이로드 미운반) |
  | 5 | REQ-AFR-019 신설 | 계약 상태 가시성 — agentfm 콘솔 표면 + doctor served-model 표면을 하나의 REQ로(동일 요구의 두 검증 면) — 저장 항목이 live로 침묵-독해되는 것 금지 |
  | 6 | REQ-AFR-020 신설 | AMI-001 REQ-AMI-011 운영자-지시 예외의 소비 계약 확장 — REQ 텍스트는 본 수정안, 동 SPEC 파일 편집은 run-phase M10(M5b 선례 재위임) |
  | 7 | Out of Scope — 런타임 소비 절 | 재기술 — 스폰-소비 공백이 본 수정안으로 경계; Codex/GLM task 모델 파라미터·런처 세션 표면·프론트매터 쓰기·`moai model profile` CLI·차단 레이어는 계속 밖 |
  | 8 | §D 비기능 제약 | 스폰 소비 제약 1행 추가(옵트인 기본 off·observe/advise 한정·[1m] 잔여·캐시 비용 독트라 인용) |
  | 9 | 예산 | REQ 20 / AC 19 — Tier M 상한(16/16) 초과. 분할 기각(운영자가 단일 계약을 본 SPEC 수정안으로 지정), 티어 상향 기각(Tier L 아티팩트 집합 design.md·research.md를 수정안이 저술하지 않음 — `TierArtifactMissingRule` 경고 유발, `internal/spec/lint.go:227`). 초과 판정은 본 수정안의 plan 감사(plan-auditor) 소관 — 상세 plan.md §I.3 |

## §A 배경과 복원 대상

2026-09-27 커밋 `384eb3460`은 SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011(서브에이전트가 메인 세션의 model/effort를 상속한다)에 따라 `moai web` 콘솔에서 에이전트 설정 탭과 저장 API를 삭제했다. 운영자는 2026-10-02 카드 t1411로 이 설정 표면(스키마·저장 경로·UI 노출)의 복원을 지시했다. 상속은 기본 동작으로 유지하되, 에이전트별 오버라이드를 다시 저장·표시하는 콘솔 표면을 되살린다.

블루프린트는 삭제 커밋의 전체 diff(`.moai/reports/t1411/removed-commit-384eb3460.diff`)이며, "revert가 아니라 re-port"가 원칙이다(§C 결정 4). 삭제 이후 5일간의 드리프트 — SPEC-AGENT-TIER-001 착지(workflow 패널 안 agent-tier 서브섹션, `agenttierpanel.go` / `jevkey.go` 배치 패턴), 모델 매트릭스 갱신(t1368), i18n 거버넌스 강화 — 위에서 현재 트리의 정형 형태로 재작성한다.

측정으로 확인된 현재 트리 상태:

- 콘솔 쪽 부재 락: `internal/web/agent_settings_removed_test.go`(렌더 마커 5종 부재 + 위조 POST 무효), `internal/web/agent_settings_test.go`(role_profiles·workflow_agents 무시 — 복원과 무관, 유지).
- 설정·템플릿 쪽 부재: `config.LLMConfig`에 `Profile`/`AgentOverrides` 필드 없음, `internal/template/profile_matrix.go` 소멸(삭제 커밋 `3fa8bd2ab`, AMI-001 M5), `internal/settings/agentfm` 패키지 소멸, 템플릿 `llm.yaml`에서 키 삭제 + 은퇴 주석 블록. 같은 커밋 `3fa8bd2ab`가 `internal/template/glm_effort_overlay.go`의 `ResolveGLMReasoningForModel`도 소멸시켰다(측정: HEAD에서 정의 0건, 본문은 제거 주석만 생존 — glm_effort_overlay.go:195) — **재포트 대상**(plan §A.1/§D.4, M1).
- 생존 기계: `internal/harness/v4manifest` closed-set 상수(`ModelInherit`/`ModelHaiku`/`EffortLow`..`EffortMax` — 실측 `schema.go:44-63`), `template.IsGLMBackend`, 키 스크립 기계 `template.ShippedRetiredModelKeys`(템플릿이 여전히 수재하는 키는 strip 제외 — 자동 해장 seam), i18n `hint.effort.go_unbound`/`hint.effort.haiku_na`(en·ko 생존 실측).
- 현재 에이전트 로스터(`.claude/agents/moai/`, 12종): manager-todo가 있고 mission-governor는 없다 — 삭제당시 agentdesc 키 집합과 다르다(re-port 근거).

## §B 요구사항 (GEARS)

### B.1 UI 노출

- **REQ-AFR-001** (Ubiquitous): The `moai web` console shall render an agent-settings sub-section inside the llm (3rd Party LLM) panel that lists every parseable agent definition under `.claude/agents/moai/`, each row carrying the agent name, the profile-matrix-resolved model/effort, and a select pair wired as `agentfm.<agent>.model` / `agentfm.<agent>.effort`, with harness-directory rows scanned but not rendered.
- **REQ-AFR-009** (State-driven): **While** a row's resolved model is `haiku`, the console shall render that row's effort select disabled with the `hint.effort.haiku_na` note, and the client wiring shall re-apply the lock on model change and on tier repopulation.
- **REQ-AFR-010** (Capability-gate): **Where** the live llm.yaml marks a GLM backend (`team_mode: glm`), the console shall additionally render each row's z.ai reasoning state, re-porting and reusing `template.ResolveGLMReasoningForModel`(원천 `3fa8bd2ab^`, plan §A.1) — no second derivation — and shall render nothing extra under a Claude backend.
- **REQ-AFR-011** (Ubiquitous): The sub-section's strings shall resolve through 4-locale i18n — new `agentfm.*` / `sec.agentfm.*` / `fieldDesc.agentfm.*` keys plus per-agent `agentdesc.<name>` ko/ja/zh entries whose en baseline is the agent frontmatter description (`data-i18n-baseline`) — and the `agentdesc.` prefix shall re-enter `i18nEnExemptPrefixes` with a recorded justification.
- **REQ-AFR-013** (Ubiquitous): The rendered sub-section shall be test-sliceable through one marker attribute (`data-section="agent-overrides"`), and the llm panel's field count shall count the rendered agent rows, so the 13-tab settings contract (`tab_layout_test.go` `wantTabOrder`, `primary_surface_test.go` `>13<`) stays byte-for-byte untouched.

### B.2 저장 경로

- **REQ-AFR-003** (Event-driven): **When** the operator submits the performance-tier selector (wire field `performance_tier`, closed set {max, medium, low} plus the client-only pseudo-state `custom`), the console shall persist a non-custom selection to `llm.profile`, and shall treat `custom` and an empty submission as preserve.
- **REQ-AFR-004** (Event-driven): **When** the operator submits per-agent edits, the console shall persist to `llm.agent_overrides` exactly the submitted matrix-member agents whose (model, effort) differs from the profile default under the target tier, shall clear submitted overrides that equal the profile default, shall skip non-matrix agents, shall backfill an unsubmitted effort with the resolved value, and shall leave llm.yaml byte-identical when nothing changes.
- **REQ-AFR-005** (Ubiquitous): The console shall not write agent frontmatter — `.claude/agents/**/*.md` is a read-only scan surface, and every save path shall leave it byte-identical (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-040 계약 승계).
- **REQ-AFR-006** (Ubiquitous): Validation closed sets shall be model ∈ {inherit, haiku, sonnet, opus, fable} and effort ∈ {low, medium, high, xhigh, max}, reusing the surviving v4manifest closed-set constants (`inherit`/`haiku`/`sonnet`/`opus` — `schema.go:62-73`) plus the config-level `Fable` constant (`internal/config/types.go:367` — v4manifest에는 fable 상수가 없다); the out-of-set rejection shall join the existing atomic-reject flow.
- **REQ-AFR-007** (Event-detected): **When** any validation error exists in a save request, the console shall reject before any write — llm.yaml, agent frontmatter, and all section files remain byte-identical — and shall re-render with per-field errors. **When** a persistence error occurs inside the llm.yaml agent-overrides write pair (llm.profile + llm.agent_overrides), the console shall restore llm.yaml byte-identical to its pre-pair snapshot, and a best-effort rollback failure shall surface in the rendered error. Persistence failures outside that pair (earlier steps — profile config, project config, nested config, section schema edits) keep the documented best-effort behavior: earlier successful writes may persist, and the failure names the failed step. Limitation (t1411 sync-audit round-2 N1): the atomic restore guarantee is scoped to exactly this write pair — partial persistence from an earlier step (e.g. section schema edits, which also edit llm.yaml sections) is a stated limitation, not a covered guarantee.
- **REQ-AFR-012** (Ubiquitous): The restored save path shall expose injectable seams (list / parse / persist) so tests substitute each step without touching default wiring, matching the console's existing seam convention (`glmcredSave` / `jevcredSave` pattern).

### B.3 설정 스키마

- **REQ-AFR-002** (Ubiquitous): Sub-agent model/effort inheritance shall remain the default behavior — an agent without an `llm.agent_overrides` entry resolves to the session model/effort — and this SPEC shall change no spawn-path behavior for sessions without the v0.3.0 opt-in key (REQ-AFR-015); the opt-in consumption contract itself is REQ-AFR-015..017. (v0.3.0 in-place qualification — REQ-AFR-007의 v0.2.0 축소 선례 형식; id 불변·재번호 없음. plan-audit iter1 D1)
- **REQ-AFR-008** (Ubiquitous): The shipped template `llm.yaml` shall carry the `llm.profile` and `llm.agent_overrides` keys again with an updated comment block, and `moai update` shall not strip a retired key while the embedded template still ships it (`template.ShippedRetiredModelKeys` seam), so `llm.performance_tier`, `llm.harness_agents`, and the `workflow.yaml` retired keys remain retired-and-stripped.

### B.4 상위 SPEC 수정 관계

- **REQ-AFR-014** (Ubiquitous): The restoration shall be recorded as a narrow amendment of SPEC-AGENT-MODEL-INHERIT-001 — REQ-AMI-011 gains the operator-directed console-surface exception, REQ-AMI-013 narrows to the still-retired key set (`performance_tier`, `harness_agents`, workflow keys) — following the repo's completed-SPEC amendment contract (spec-frontmatter-schema.md § completed → in-progress (amendment): HISTORY 행, 요구사항 삭제 없음, id 재번호 없음, `amendment_of` + `## Amendments` + `prior_completed_sha` = 기존 close의 `sync_commit_sha` 기록; `sync_commit_sha` 불변, 상태 축은 SSOT 수정 전이 `completed → in-progress`를 따른다 — sync-audit F7 재기술). 수정안 SPEC의 재완료는 수정 범위의 독립 검증 후 본 카드와 별도 close 커밋으로 수행하며(§E.4 갱신 동반), 그 전까지 드리프트-면제 형상(in-progress + amendment_of)을 유지한다.

### B.5 스폰 소비 계약 (v0.3.0 — card t1421, 옵트인 고정)

- **REQ-AFR-015** (Capability-gate): **Where** the `llm` config section declares `llm.agent_overrides_consume`, the key shall be a boolean defaulting to `false` — a session that sets it `true` opts its subagent spawns into consuming `llm.agent_overrides`, and a session without the key keeps today's storage-only behaviour (REQ-AFR-002 불변) — and a non-boolean value shall join the existing atomic-reject flow (REQ-AFR-006/007). The shipped template `llm.yaml` shall carry the key (default `false`) with an updated comment block, and `moai update` shall not strip it — the shipped-key inventory registration contract (`TestShippedConfigKeysHaveReaders`, REQ-AFR-008의 seam 연장) extends to the new key.
- **REQ-AFR-016** (State-driven): **While** `llm.agent_overrides_consume` is `true`, the read path shall resolve each agent's (model, effort) as — the agent's `llm.agent_overrides` entry wins; an absent entry resolves to plain inheritance (the session model/effort); an override value of `inherit` is an explicit inheritance no-op — and **While** the key is `false` or absent, the same map stays console-stored-only. 해상 형태는 삭제 커밋 `3fa8bd2ab^`의 블루프린트(`EffectiveProfile`·`validateAgentOverrides`)를 따르되, 현 트리에 M1이 재포트한 해상 기계(`ResolveAgentModelEffort` 등)를 소비 게이트 뒤에 연결한다 — 재유도 금지, 단일 유도.
- **REQ-AFR-017** (Ubiquitous): When the opt-in is on, the orchestrator shall consult the resolved overrides before spawning and pass the configured model/effort on the Agent() call — 유일한 승인 소비 메커니즘이다(Claude Code의 서브에이전트 모델 해상 순서: 스폰 시점 `model` → 프론트매터 → `CLAUDE_CODE_SUBAGENT_MODEL` → 세션 모델 — `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention); opt-in 없는 세션의 상속 기본 문장은 변경되지 않는다(`.claude/rules/moai/core/agent-common-protocol.md` § Subagent Model and Effort). 룰 텍스트의 소비 조항 자체는 sync-phase 규칙 개정으로 착지한다(REQ-AMD-001 계열 — run 코드 확정 후 정산, 본 SPEC은 본문에 그 계약만 규정).
- **REQ-AFR-018** (Event-driven): **When** the opt-in is on, the spawn guard hook (`internal/hook/agent_model_guard.go`) shall extend its `advise` layer to compare the spawn's declared parameters against the agent's override expectation, and the audit record (`.moai/logs/agent-model-audit.jsonl`) shall gain an override hit/miss field — observe-never-blocks는 보존된다(차단 레이어 부활·훅 입력 재작성 없음). 비교 가능한 축은 모델이다 — effort는 스폰 페이로드가 운반하지 않는다(기존 파일 헤더 계약 "the Agent tool exposes no effort parameter" 승계); effort의 전달은 교리 수비(REQ-AFR-017)가 담당한다.
- **REQ-AFR-019** (Ubiquitous): The contract state — consuming vs stored-only — shall be visible wherever a stored override can be mistaken for a live one: the agentfm console surface shall mark the sub-section's consumption state, and the served-model doctor surface (`internal/cli/doctor_served_model.go`) shall report the switch state — 하나의 요구에 두 검증 면(console·doctor)이며, 저장 항목이 소비 상태를 침묵시키는 표현은 어느 면에서도 금지된다.
- **REQ-AFR-020** (Ubiquitous): The opt-in consumption contract (REQ-AFR-015..019) shall be recorded as the growth of the operator-directed exception in SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011 — REQ 텍스트는 본 수정안(v0.3.0)이 규정하고, 동 SPEC 파일 편집(HISTORY 신규 행 + REQ-AMI-011 예외 문단의 소비 계약 확장)은 run-phase 마일스톤 M10이 수행한다(M5b 선례 — manager-spec 재위임, spec-frontmatter-schema.md § Forbidden ownership crossings). 형식 불변: 요구사항 삭제 없음, id 재번호 없음, 동 SPEC 기존 close의 `sync_commit_sha` 불변, `TestJevAmendmentLinkage` 스타일의 본문·HISTORY·인용 3자 정합.

REQ 총수 20 (v0.1.0 14 + v0.3.0 신설 6) · AC 총수 19 (13 + 6) — Tier M 상한(16/16) 초과. 예산 신호의 처분: 분할·티어 상향 모두 기각, 근거는 `## Amendments` v0.3.0 블록 9행과 plan.md §I.3; 초과 판정은 본 수정안의 plan 감사 소관이다.

## §C 설계 결정 (운영자 지시 5항목의 명시적 처분)

| # | 질문 | 처분 | 요지 |
|---|------|------|------|
| 1 | REQ-AMI-011 처리 | **형식적 좁은 수정안(narrow amendment)** — Q1 운영자 비준 2026-10-02 (decision-index) | 상속은 기본값으로 유지, 콘솔 표면만 예외화. 단면별 처분표는 plan.md §D.1. 정합성 필연으로 REQ-AMI-013(템플릿 키)·REQ-AMI-014(스크립)도 함께 정리한다 — 복원 저장 경로가 쓰는 키를 update가 벗겨내면 표면이 자기모순이 되기 때문(`ShippedRetiredModelKeys` seam 실측: 템플릿 수재 키는 자동 strip 제외). |
| 2 | 같은 커밋이 지운 rosterguard/registry.go·v4manifest 표시 헬퍼 | **OUT** | 블루프린트에서 웹 저장 경로가 v4manifest에서 가져오는 것은 closed-set 상수뿐이고(전부 생존 실측), `ModelColor`/`Tier` 표시 헬퍼는 현재 티어 기계와 중복된다. rosterguard 행은 콘솔 저장 경로와 무관. 단, 복원 파일에 대응하는 rosterguard/numeral 등록은 M5에서 재검토한다(고아 방지). |
| 3 | 착지된 티어 패널과의 공존 | **llm 패널 서브섹션 + 기계 공유** — Q3 운영자 확정 2026-10-02 (decision-index) | `agenttierpanel.go`/`jevkey.go`의 배치 패턴(마커 상수 + `.subsection` 블록 + 커스텀 parse/persist)을 공유한다. 배치는 workflow 패널이 아니라 **llm(3rd Party LLM) 패널** — 저장 대상이 llm.yaml이므로 파일 정합성이 있다. 티어 차트(workflow 패널)는 '왜 티어가 있는가'의 표시 근거, 본 서브섹션은 에이전트별 제어 — 상호 참조하는 두 표면. 탭 14개 복원안은 기각(탭 계약·i18n·chrome 락 전면 재협상 비용). |
| 4 | revert가 아니라 re-port | **채택 (구체 차이 명시)** | 로스터 차이(mission-governor→manager-todo), 매트릭스 셀은 9-27 스냅숏이 아니라 현재 모델 매트릭스(t1368 착지본)에서 재유도, `llm.performance_tier` 레거시 alias 미부활(마지막 소비자 소멸), 프론트매터 쓰기 없음(REQ-MPM-040 승계), 14번째 탭 아님. 포트 선례: t1385/t1388. |
| 5 | 락 테스트 반전 (observational-RED) | **열거 처분표 (표면-결합 집합)** | plan.md §D.5에 표면-결합 테스트 파일 전부(실측 스캔 기준 28파일)를 파일별 invert/replace/keep-and-why 행으로 **명시 열거**한다 — 전수 서술이 아니라 열거이며, 신규 파일은 run-phase 착수 시 재판정한다. 현재 부재 락이 TDD의 RED 기선이다. |

## §D 비기능 제약

- 방법론: TDD(quality.yaml `constitution.development_mode: tdd`) — 각 마일스톤 RED→GREEN→REFACTOR, 최소 커버리지 커밋당 80%.
- 원자성: 검증 오류는 쓰기 전 원자 거절 — llm.yaml·agent frontmatter·전체 섹션 파일 byte-identical (기존 atomic-reject 흐름 재사용). 영속 오류 복원은 llm.yaml agent-overrides write pair(`llm.profile` + `llm.agent_overrides`)로 한정 — 사전 스냅숏 복원, 롤백 실패는 오류에 병기. pair 밖 선행 단계는 기존 best-effort 동작 유지 — 선행 성공 쓰기는 잔존할 수 있고 실패는 실패 단계를 명명한다 (REQ-AFR-007 한계 명시, 카드 t1446).
- i18n: 신규 사용자 노출 문자열 4-locale(ko/en/ja/zh) 필수, i18n 거버넌스·미번역 허용목록 테스트 통과.
- TRUST 5 전 영역 + LSP 게이트(quality.yaml `lsp_quality_gates`: plan 기준선 필수, run 오류 0).
- 탭 계약: 13탭 유지 — `wantTabOrder`·`>13<` 어설션 무변경 통과.
- 스폰 소비 (v0.3.0): 옵트인 기본 off — 소비 계약은 관측(observe/advise)만 강화하고 어떤 스폰도 차단하지 않는다(차단 레이어·훅 입력 재작성 부활 금지). override 값은 REQ-AFR-006 폐쇄집합 별칭만 허용한다(전체 모델 ID 형태 금지 — model-policy.md 의도적 불허). 스폰 시점 모델 핀은 부모의 [1m] 컨텍스트 자격을 상속하지 않는다("Usage credits required for 1M" 즉시 스폰 실패 — upstream #36670/#51060; STILL-ACTIVE 기제, 현재 기본 라인업 실영향 0 — model-policy.md; 별칭 폐쇄집합이 완화하며 잔여는 문서화로 닫는다). 프롬프트 캐시 비용은 기존 독트라 인용만 한다(cache-aware-execution.md 지시 5 — 스폰별 모델 오버라이드는 세션이 쌓은 모든 캐시에서 스폰을 갈라놓는다; 지시 10 — effort 전환은 캐시를 무효화한다); 미측정 토큰·금액 수치를 본문에 넣지 않는다.

## §E 관련 SPEC

- 수정 대상: SPEC-AGENT-MODEL-INHERIT-001 (completed → 좁은 수정안, REQ-AFR-014)
- 기계 공유: SPEC-AGENT-TIER-001 (agenttierpanel 배치 패턴), SPEC-WEB-CONSOLE-011 (원래 agentfm 표면의 모 SPEC)
- 셀 데이터 계보: SPEC-MODEL-PROFILE-MATRIX-002 (33셀 매트릭스 원형; superseded — 현재 계보: SPEC-MODEL-MATRIX-UPDATE-001), SPEC-MODEL-MATRIX-UPDATE-001 (t1368 현재 모델 매트릭스)

## Out of Scope

### Out of Scope — 런타임 소비(spawn 경로)

- (v0.3.0 재기술) 복원된 `llm.agent_overrides`의 스폰-소비는 **본 수정안(v0.3.0)이 옵트인 계약으로 범위에 넣었다** — REQ-AFR-015..020. decision-index Q2의 후속 카드 t1421이 이 공백을 소진한다(운영자 결정 2026-10-03, 옵트인 고정). 구 서술 "스폰 시점 강제는 별도 후속 SPEC이다"는 소진됨 — 후속이 발행된 것이 아니라 본 SPEC 안에 흡수되었다.
- 계속 밖: (a) Codex/GLM 작업 위임 표면(`codex_task`·`glm_task` 등)의 모델 파라미터 — Claude 서브에이전트 Agent() 스폰 소비와 별개 위임 경로다. (b) 런처 세션 표면(`moai cc`/`moai glm`/`moai gpt`)의 세션 시작 모델 선택 — 세션 모델은 프로파일·환경 축 소관이다. (c) 에이전트 프론트매터 쓰기(REQ-AFR-005 승계 — 소비는 Agent() 호출 파라미터 전달이지 `.claude/agents` 파일 편집이 아니다). (d) `moai model profile` CLI 부활. (e) 차단 레이어(`workflow.agent_model_guard.enabled`) 부활과 훅 입력 재작성 — 소비 계약은 observe/advise 관측만 강화한다.

### Out of Scope — t1246 커밋의 비(非)콘솔 삭제분

- `internal/harness/rosterguard/registry.go` 행 복원(삭제된 5개 사이트·1 exemption) — 콘솔 저장 경로가 임포트하지 않음(§C 결정 2). 단 복원 파일 자체의 가드 등록은 M5 범위.
- `internal/harness/v4manifest` Tier/ModelColor 표시 헬퍼 복원 — 현재 티어 기계와 중복.
- `internal/web/webux_followup_test.go`(-402)·`webux_haiku_effort_test.go`(-214) 통째 재생성 — 해당 AC는 새 형태의 테스트로 흡수.

### Out of Scope — 프론트매터 쓰기·레거시 alias·문서

- 에이전트 `.md` frontmatter `model:`/`effort:` 쓰기 (REQ-AFR-005 — 읽기 전용 스캔만 복원, `internal/settings/agentfm`의 `Patch` 쓰기 레이어는 부활하지 않는다).
- `llm.performance_tier` 레거시 alias 키와 `llm.harness_agents` — 마지막 소비자가 소멸한 죽은 키를 되살리지 않는다 (단순성 사다리).
- docs-site 페이지(삭제당시 i18n 안내 문구 포함)와 `model-policy.md` 등 규칙 문서 개정 — run-phase 코드 변화 확정 후 sync-phase에서 REQ-AMD-001 계열과 정산.
- `workflow.workflow_agents`·`model_routing*` 키 — Agent Teams 실험 재허용과 무관하게 유지.
