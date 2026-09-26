# progress.md — SPEC-AUTONOMY-DECIDER-MODE-001

> Plan-phase authoring record (2026-09-26): [PLAN] AC judgments are recorded in the §E.1 block
> below at authoring time as designed-for (grep counts verified during authoring); the formal
> re-execution belongs to the plan-audit gate.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.1.0 (initial plan-phase authoring)
- plan_audit: FAIL (iter-1, 2026-09-26, score 0.91 — Tier M threshold 0.80 met but 2 blocking
  findings: D1 research.md §3 π_d=0.15 row unreproducible from the stated formula [stated
  params yield n_d≈18/N≈118 vs claimed 29/191; 0.20/0.30 rows reproduce], D2 REQ-DM-002
  pinning-rule natural-vs-synthetic precedence reads two ways across spec.md/plan.md/research.md.
  MP-1..7 all PASS; D3 minor blocking: AC-DM-016 lacks the pool-fidelity Gaps-row requirement.
  Repair list + delta-scoped re-audit scope: .moai/reports/t1261/plan-audit-iter1.md)
- iter1_repair_adoption: D4 (AC-DM-002 command coverage over research.md §3 rows) adopted by lane
  per the card t1266 autonomy policy — closes the AC dead zone where a D1-class value error in
  research.md sat invisible to the AC layer; D5–D7 remain operator-discretion items in the
  iter-1 report, not adopted this iteration
- plan_audit: PASS (iter-2, 2026-09-26, score 0.97 — Tier M threshold 0.80 met; delta re-audit
  scope D1–D4, all four RESOLVED: D1 row ≈18/≈118 re-derived from the stated single parameter
  set + envelope 120–240 aligned with zero residue, D2 single-reading precedence
  [qualifying natural always outranks synthetic; lowest baseline ties among naturals] across
  spec/plan/research, D3 AC-DM-016 pool-fidelity Gaps row in the enforceable When/Then path,
  D4 AC-DM-002 extended greps re-executed green [3/1/1/1/1/2]. MP-1..7 unchanged PASS;
  REQ 15/AC 16; scope guard 0 product-code @ 681a081ca. Report:
  .moai/reports/t1261/plan-audit-iter2.md)
- tier: M (spec.md, plan.md, acceptance.md, research.md — measurement-design card, zero code
  deliverables; progress.md present at every tier)
- requirements: 15 (REQ-DM-001..015, continuous) / acceptance criteria: 16 (AC-DM-001..016;
  [PLAN] 11, [RUN] 5 — AC-DM-012..016 are run-record targets, `--- PENDING-RUN` until the run)
- card: t1261 (AUTONOMY-A5b — contract-mode kickoff decider default: llm 단독 vs llm+jev;
  t1244 split, operator decision 2026-09-26, Tier M, Class C)
- intent_clarity: ~9/4 quadrants — interview SKIPPED, rationale: operator-drained card (the
  operator's decision is already verbatim in the card text), factory-lane context (questions
  route through the lead, not the operator), and the card t1266 autonomy policy (card-specific
  condition choices take the recommended option with rationale recorded). The one genuinely
  open either/or — primary LLM-arm pool under an externally gated Opus quota — is closed by a
  pre-fixed conditional rule (GLM primary + pre-declared Opus supplement), not left open.
- measurement_target: the contract-mode kickoff decider DEFAULT decision evidence —
  `workflow.autonomy.kickoff.decider` derives `llm` under `mode: contract` today (A1 0.5.2
  REQ-CONTRACT-015); changing that default is A1-owned follow-up, NOT this card
- population_redefinition: FIRST-CLASS M0 — census per candidate (C-A kickoff rounds:
  sweep-exhausted, rejected; C-B other surfaces: census-gated, ≥ 759 non-kickoff AUQ rounds in
  the t1244 snapshot window; C-C synthetic control: recommended primary, baseline 50% by
  construction) → power arithmetic per candidate → pin ONE by the pre-fixed rule (or 판별 불가,
  a valid exit) → criteria_commit BEFORE any judge call
- power_arithmetic_seed: baseline > 90% kills baseline+10%p (t1244 structural lesson);
  balanced 50/50 → band (b) = 60% meaningful; two-arm Δ = 10%p at α = 0.05 / 80% power needs
  ≈ 120–240 items at plausible discordant rates (π_d 0.15–0.30) — target N = 200, minimum interpretable batch
  N = 120 (MDE ≈ 11–12%p there)
- predicate_preregistered: McNemar-style discordant-pair exact binomial (two-sided α = 0.05);
  per-arm accuracy always alongside both constant baselines; wrong-automation ≤ 10%; band shape
  (a) n ≥ 20 / (b) baseline+10%p / (c) wrong-automation, re-derived from the pinned baseline;
  confidence fields declared BEFORE measurement — Jev authoritative = response registry
  `probabilities` (t1244 sync-audit D3 divergence fix), instrument `confidence` auxiliary
- composite_and_mapping: llm+jev composite = agree → label, disagree → `hold`
  (confidence-gating deliberately outside the composite); mode mapping δ = 5%p — better →
  recommend `llm+jev`; worse or equal → recommend `llm 단독`; pool named on every
  recommendation, pools never merged
- llm_arm_pool: primary GLM (z.ai, `glm_task` — quota separate from the Anthropic lane pool);
  pre-declared Anthropic (Opus) supplement ONLY on a lead-granted quota window (timing
  externally gated — recorded dependency); Codex named fallback; criteria commit names the
  exact model identifier
- preregistration: criteria_commit pattern inherited from t1244 (criteria pinned by a branch
  commit, hash recorded and time-compared against the first judge call); three caps
  (turn 2/item, call 2 × (N + controls) per arm-batch, wall-clock PT8H/batch, batches = 1 per
  arm) declared with `declared_at` before the first call; no cap extension
- scrub_gate: inherited REQ-CALIB-011 shape — 4 deny classes + PII lower-bound set, positive
  control before FIRST transmission, fail-closed (hit ⇒ block, never strip-and-send), applied to
  EVERY external judge (Jev/TypeSafe, GLM, Anthropic, Codex); lower-bound limitation goes to the
  verdict Residual-risk
- korean_and_t943_discipline: no translation arm (§30 measured the ~10%p language effect,
  conclusion-invariant); t943 premise-triage NOT re-run; §30 rows cited only with the
  constant-baseline row; three-domain disjointness (t943 / t1244 / t1261) stated in the SPEC
  body and repeated in the verdict
- predecessor_anchors: A1 sync close `e4ea8eb05` reachable from this branch
  (`git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0, verified 2026-09-26 in this worktree);
  t1244 reference artifacts present at `.moai/reports/t1261/reference/` (verdict, run-record,
  extract summary); both depends_on SPECs `status: completed`; A3 (t1236) artifacts NOT in this
  tree — cited as card context only
- linked_pr: none (card worktree branch `WT-decider-mode-eval`; git-flow lane — local develop
  merge via the lead's window, lane does not push)
- run_preconditions: autonomous kickoff per §A.2 / REQ-DM-014 (card t1266 policy — proceeds
  after plan-audit, bound by bundled conditions); pool availability per REQ-DM-009 fallback
  order; measurement-impossible verdict as the total-unavailability exit; scrub failure blocks
  transmission (fail-closed), a different path from unavailability
- implementation_kickoff: AUTONOMOUS — executed 2026-09-26 by the lane orchestrator after
  plan-audit iter-2 PASS 0.97 (operator policy card t1266 + lead dispatch "킥오프 자율" +
  REQ-DM-014; no lane-window question asked). Bundled conditions stand: pool per REQ-DM-009
  fallback order, three caps + scrub positive control declared before the first judge call,
  criteria_commit precedes any judge call, 판별불가 is a valid documented exit
- open_clarifications: NONE at plan phase — NC-1 (δ = 5%p and +10%p are pre-registered judgment
  values, changeable only before the criteria commit) and NC-2 (natural-over-synthetic
  tie-break) are body decisions with defaults kept, recorded in research.md §7; no
  `[NEEDS CLARIFICATION: ...]` markers carried
- scope_guard: product code, `internal/`, `internal/template/` untouched — instruments live
  under `.moai/reports/t1261/` (untracked during work); AC-DM-011 checks the commit set
- plan_artifact_hash: 8549f3773b670a264ac47d076bfae8970d048d1f73c35c4806e10f0c1d0ba49c
  (sha256 over cat acceptance.md plan.md research.md spec.md, exact bytes; measured at
  plan-audit iter-2 PASS entry, 2026-09-26, tree @ 681a081ca; supersedes the iter-1 value
  3cabee24a6e5dc2f5b2c09202797f11d893d46bd617747b08c49ae5439b46387 @ e611a99a1)
- Implementation Kickoff Approval: not requested at plan phase (run entry follows the
  plan-audit gate and the autonomous-kickoff disposition)

## §E.2 Run-phase Evidence

### M0 — population census, power arithmetic, pin (2026-09-26, HEAD 681a081ca 기준 측정)

**Census snapshot**: 2026-09-26, 양쪽 코퍼스 루트 전체 스윕 (루트 A
`~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go*` + 루트 B
`~/.claude/projects/-Users-goos-MoAI-moai-adk-go*`): jsonl 1,053개, AskUserQuestion tool_use 925콜,
짝 922 / 미짝 3, 질문 행 1,429 중 응답 행 1,257. 계측기·원시 증거(미추적):
`.moai/reports/t1261/census/` (census.py, rounds.jsonl, excluded_records.json,
census_table.{md,json}, arithmetic.{md,json}, probe_*.py). 전사본 코퍼스는 읽기 전용 유지.

**표면별 센서스** (응답 행 N / always-approve 기준선 p_approve / always-hold p_hold / raw-modal
최다 원시 라벨 점유율 — 결정 클래스 정규화 술어는 census.py에 원문 기록):

| 표면 | N | p_approve | p_hold | p_modal(raw) |
|---|---|---|---|---|
| C-A kickoff-gate | 69 | 86.96% | 1.45% | 15.94% |
| B1 plan-audit-failure | 79 | 49.37% | 2.53% | 7.59% |
| B2 sync confirmation | 8 | 50.00% | 12.50% | 11.11% |
| B3 blocker re-delegation | 5 | 20.00% | 0.00% | 20.00% |
| B4 escalation/budget | 22 | 27.27% | 4.55% | 4.35% |
| B5 X1-shape (발행/범위·일정/렌즈) | 117 | 19.66% | 0.85% | 6.33% |
| B6 card pick/dispatch/queue | 132 | 12.88% | 5.30% | 2.17% |
| UNCLASSIFIED (비표면 기록) | 825 | 27.15% | 7.03% | 1.69% |
| MALFORMED_INPUT (제외) | 33 | — | — | — |
| NON_ANSWERED (제외) | 139 | — | — | — |

기록된 제외 (침묵 없음): MALFORMED_INPUT 33 = `__unparsedToolInput` 콜(하네스
InputValidationError — 비라운드); NON_ANSWERED 139 = "No response after 60s" 폴백 131(t1244 X2
동류) + 하네스 거절 8; tool_use 미짝 3. 원문 통째 보존: `census/excluded_records.json`.

**양성 대조 (핀 선행, 둘 다 발화)**:
- PC1 — t1244 기지 대조 항목 `toolu_01RMK35YKc8bLzoyCz6cCbCP` 본 스윕에서 발견(2행); 추출된
  질문·응답이 t1244 run-record 행과 원문 일치(t686 plan 감사 처분, 응답 「문면 1회 수정 후
  Kickoff」 = t1244의 modify 매핑 대조). 추출 술어(tool_use/tool_result 짝 + `"질문"="라벨"` 쌍)
  이 t1244 술어군과 호환 검증됨.
- PC2 — kickoff-gate 형상 검출: 응답 게이트 라운드 69, p_approve 86.96%. 다른 표면의 0이
  추출기 실패로 오독될 수 없다. (t1244는 자기 스냅숏·자기 2단계 술어에서 99.05% — 차이는 코퍼스
  드리프트 + 술어 세립도이며 동일시하지 않는다.)

**검정력 산술** (REQ-DM-003 공식, `census/arithmetic.py`로 재유도 가능): Δ=10%p·양측
α=0.05·검정력 80% — π_d 0.15→n_d≈17.6/N≈117.6, π_d 0.20→≈31.4/≈156.8, π_d 0.30→≈70.6/≈235.2
(research.md §3 재현). 달성 N이 완전 검정력을 내는 π_d 하한: π_d ≥ N/784.

**규칙 적용 흔적 (REQ-DM-002)**:
1. C-A: N=69 < 120 → 미달 (context-only; t1244 스윕 완전성은 자기 스냅숏 기준 유지).
2. B1 (79)·B2 (8)·B3 (5)·B4 (22)·B5 (117): 전부 최소 해석 가능 배치 120 미달 → 미달격.
3. B6 (카드 배차/선택/큐 승격, N=132): 센서스 두 조건(기준선 ≤ 90%, N ≥ 120)은 형식적으로
   통과하나, 고정된 계측 형태로는 측정 불가 — REQ-DM-008 이 Jev 단계 라벨 공간을 4값
   {approve, hold, modify, other}로 고정하고 REQ-DM-007 합성이 그 공간 위에서 두 팔 라벨을
   결합하는 반면, B6 의 ground truth 는 질문별 카드 선택지(identity) 공간이라 일치 정확도
   (REQ-DM-006 (2))가 두 공간 사이에서 정의되지 않는다. REQ-DM-003 의 커밋 대상(always-approve /
   always-hold 기준선, wrong-automation 밴드 (c))은 hold/approve ground-truth 라벨을 전제하는데
   B6 에는 없다(키워드 프록시 클래스는 카드명 위의 센서스 인공물로 별도 기록). 고정 계측 하
   달성 가능한 측정 N = 0 → B6 미달격. 이 발견과 수정 경로(선택지-선택형 모집단을 측정 가능하게
   하는 REQ 문구 개정은 후속 카드)를 여기에 기록 — 침묵 폐기 아님. 첫 판사 호출 전이라면 새
   criteria 커밋으로 재심 가능(t1244 재핀 선례); M2 첫 호출 뒤에는 고정.
4. C-C (합성 대조군): 구성에 의해 자격 — 기준선 50%, 밴드 (b)=60%; 커밋 배치 N=120 (clean 60 +
   결함 60, 6클래스 × 10; 69개 kickoff-gate 베이스 페이로드에서 서로 다른 베이스로 구성 —
   정보성 abs-path 적중 0; 스크럽 판정 자체는 M1 스캐너 몫). N=200 목표는 미커밋 — 베이스
   중복 사용을 요구해 항목 독립성을 깨므로; N=120 에서 MDE ≈ 11–12%p (π_d=0.20) — 귀무 판독 시
   인용.
5. 자격 있는 자연 표면 부재 → 동률 판정 불필요 → **핀: C-C 합성 대조군을 1차 모집단으로**
   (규칙의 폴백 1차).

**커밋된 기준 (본 M0 핀 커밋 = criteria_commit. SHA 는 M1 run record 가 기록하고 첫 판사 호출
전 `git cat-file -e` 검증 + 시간 비교(REQ-DM-014(a)); 자기참조 슬롯 pending-backfill-m0)**:
- 고정 모집단: C-C 결함 주입 kickoff 형상 항목; defect:clean 50:50; 클래스 균형 10×6;
  ground-truth 맵 사전 고정 — (i) 필수 섹션 누락→hold, (ii) 범위 위반→modify, (iii) 전제
  위반→hold, (iv) 증거 위조 주장→hold, (v) AC 수 불일치→modify, (vi) AC 상호 모순→hold →
  분포 approve 60 / hold 40 / modify 20 / 120 (4값 공간 실 변산; `other` 는 판사 측 매핑에서만).
  베이스 페이로드: 본 센서스의 69개 실제 kickoff-gate 질문 페이로드, scrub-passing 프로젝션,
  운영자 응답 제거.
- 상수 기준선: always-approve = 50%, always-hold = 33.33% (구성에 의함; M1 구성 기록에서 재측정).
- 술어 (REQ-DM-006): 항목별 정오/오답 짝 위 불일치쌍 정확 이항검정, 양측 α=0.05; 팔별 정확도는
  두 상수 기준선과 항상 병기; wrong-automation ≤ 10%; 밴드 재유도: (a) n ≥ 20, (b) 정확도 ≥
  60%, (c) wrong-automation ≤ 10%. 신뢰도 필드: Jev 본원 = 응답 레지스트리 `probabilities`
  (계측기 `confidence` 보조); llm 팔 응답 형식 = JSON {label, confidence, rationale} —
  `confidence` 본원.
- 합성 + 매핑 (REQ-DM-007): agree→label, disagree→hold; δ=5%p; better→`llm+jev` 추천,
  worse-or-equal→`llm 단독` 추천; 풀 명명, 풀 병합 금지.
- LLM 팔 풀 (REQ-DM-009): 1차 GLM (z.ai) via `glm_task`, 모델 식별자 `glm-5.3`, 할당 풀 z.ai
  (Anthropic 과 별개); run record 가 첫 호출 시점에 실제 해석 식별자를 재확인(드리프트 시
  계측기 수리 규칙); Opus 보충 배치는 리드 부여 창에서만 사전 선언 실행; Codex 이름 지정 폴백.
- 세 상한 + 스크럽 선언: M1 의 사전 호출 의무 (REQ-DM-013/012) — 본 커밋 대상 아님.

**Gaps (M0)**: t1244 스냅숏 대비 코퍼스 드리프트(907→925 콜; 본 수치는 이 단면의 것, 재현
불가); 결정 클래스 키워드 정규화의 boundary_both 9행(UNCLASSIFIED) 기록; B6 술어 우선순위로
B2 1행·B4 1행이 B6 으로 이동(경계 중복, rounds.jsonl `surface_match` 에 각 행 기록); 짝 수
불일치 콜 1; UNCLASSIFIED 825행은 이질적 꼬리 — 비표면으로 기록(모집단화 금지); B6 의 형식적
통과 + 계측 경계 판정은 양독 가능한 판단이므로 양쪽 독해를 모두 기록했다.

**Residual-risk (M0)**: 잘리는 코퍼스의 단면 비재현성; 분류기 경계 사례(B5/B6 의 p_approve
열은 프록시 성격); B6 측정 가능성 판정은 향후 새 criteria 커밋으로 개정될 수 있고 그때 핀이
바뀐다; C-C 의 합성-vs-자연 생태 타당성 격차는 §E 대로 판정서의 Residual-risk 로 이전된다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
