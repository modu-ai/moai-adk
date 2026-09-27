# progress.md — SPEC-AUTONOMY-DECIDER-MODE-001

> Plan-phase authoring record (2026-09-26): [PLAN] AC judgments are recorded in the §E.1 block
> below at authoring time as designed-for (grep counts verified during authoring); the formal
> re-execution belongs to the plan-audit gate.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.2.0 (operator-approved criteria amendment 2026-09-27 — judge controls
  redefined as pipeline validation; 0.1.0 initial plan-phase authoring)
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

**Orchestrator disposition — B6 (2026-09-26, lane, t1266 자율 정책)**: M0 의 B6 미달격 판정을
유지하고 C-C 핀을 확정한다. 근거: 본 카드가 재는 구성은 「킥오프 판단자」의 결정 품질이며
B6(카드 배차·선택)은 4값 라벨 공간과 wrong-automation 이 정의되지 않는 별개 구성(선택지-선택)
이다 — REQ-DM-003/006/007/008 이 전제하는 계측 형태와 맞지 않아 이 카드의 질문에 대한 답을
생산하지 못한다. 재핀은 첫 판사 호출 전에만 가능하나 개정할 이유가 없다(개정은 다른 질문을
재는 것이지 이 질문을 더 잘 재는 것이 아니다). 선택지-선택형 모집단의 계측 가능화는 후속
카드 후보로 리드에게 상신한다. 본 판정은 첫 판사 호출 전 언제든 새 criteria 커밋으로 뒤집을
수 있다(되돌리기 비용 없음).

### M1 — 계측기·스크럽 게이트·상한 선언 (2026-09-26, HEAD 786e7a419 기준)

**criteria_commit backfill (REQ-DM-014(a))**: `criteria_commit: 786e7a419` — M0 핀 커밋(`fix(SPEC-AUTONOMY-DECIDER-MODE-001): M0 population census and pin`).
검증(본 워크트리, 2026-09-26 측정): `git cat-file -e 786e7a419^{commit}` → exit 0; `git merge-base --is-ancestor 786e7a419 HEAD` → exit 0. 첫 판사 호출(M2)은 본 기록과 이 검증보다 **뒤**여야 하고 그 사실이 M2 런 레코드에 시간 비교로 다시 기록된다(REQ-DM-013 선언 후 호출 순서).

**계측기 목록** (전부 `.moai/reports/t1261/` 아래 미추적 — REQ-DM-015 레이아웃, t1244 선례): `scrub_scan.py`(REQ-DM-012 스캐너), `scrub_positive_control.py`(양성 대조 러너), `construct.py`(REQ-DM-004 생성자), `inspect_bases.py`(베이스 프로파일링), `jev_judge.py`(REQ-DM-008 Jev 계측기 — M1 에서 dry 검증만), `llm_parse.py`(llm 팔 응답 파서), `batch.json`(120항목), `llm_payloads/*.txt`(120파일), `llm_manifest.json`, `controls_selected.json`(9건), `pilot_selected.json`(20건), `scan_log_m1.jsonl`(120행). M0 센서스 계측기(census/)는 원시 증거 사슬로 **수정 없이 보존**.

**스크럽 게이트 (REQ-DM-012)** — deny 4 클래스 + PII 하한: (1) 키 형태 sk-/ghp_/gho_/AKIA/PEM/.env KEY=VALUE, (2) settings.local 계열(tmux pane id·bearer·토큰형), (3) 절대 경로(/Users/, /home/, /private/, /tmp/, /var/folders/, /(usr|etc|opt|Library)/) **+ `~/` 확장 추가**(t1244 집합 대비 강화 — 패턴 추가는 스크럽 강화이지 criteria 변경 아님), (4) PII 하한(주민번호·전화·이메일·카드번호). fail-closed: 적중 시 exit 1로 전송 경로 차단 — 벗겨 보내기 없음, 호출부는 exit 0만 전송.

**양성 대조 (첫 전송에 선행 — 요구 8/8)**: `python3 scrub_positive_control.py` → 8종 더미 비밀(key_sk·key_ghp·key_aws·key_pem·path_users·pii_email·pii_phone_kr·pii_resident_kr) 전부 발화 — `scrub_positive_control: FIRED 8/8, clean_control: OK, passed: true`, exit 0. 더미 값은 파일에 리터럴로 두지 않고 런타임 조립(가짜 값, 로컬 전용, 전송 없음). 클린 대조군(게이트 질문 형태 표본) `scan=clean exit=0`. 관측 시각 **2026-09-26T20:51:27+09:00** — M2 첫 전송에 선행.

**생성자 (REQ-DM-004)** — 69개 C-A 베이스(센서스 `census/rounds.jsonl`, 읽기 전용)에서:
- clean 60 = 원래 approve 였던 60개 베이스를 그대로(운영자 응답 제거) — 지상진실 approve 가 원래 결정과 일치해 성립.
- defect 60 = 60개 **서로 다른** 베이스(비승인 9개 전부 + 승인 베이스 중 시드 샘플 51), 6클래스 × 10. 결함 베이스 상호 중복 0(기계 검사 통과). clean∩defect 베이스 겹침 51은 120 > 69 산술로 강제 — 항목 텍스트는 결함 주입으로 서로 달라 판정독립성 유지, 겹침 수치는 판정서 Residual-risk 로 이전.
- 지상진실 맵 사전 고정 준수: (i)필수 섹션 누락→hold, (ii)범위 위반→modify, (iii)전제 위반→hold, (iv)증거 위조 주장→hold, (v)AC 수 불일치→modify, (vi)AC 상호 모순→hold → **approve 60 / hold 40 / modify 20 / 120** (기계 검사: 항목 120, 클래스별 정확히 10, 라벨 분포 정확히 60/40/20, defect 베이스 중복 0 — 전부 통과, exit 0).
- 상수 기준선 재측정(구성 기록에서): always-approve 60/120 = **50.00%**, always-hold 40/120 = **33.33%** — M0 커밋값과 일치.
- 주입 변형 분포: (ii)3종, (iii)3종, (iv)3종, (v)n∈{3,4,5}, (vi)2종 시드 배분; 클래스 (i) 10건 전부 `i_omit_evidence`(증거 문장 제거형 — 10베이스 전부 다문장 적격으로 확인, 선택지-생략 변형 불필요).
- 스크럽 스캔: 생성 페이로드 120건 전부 `clean`(scan_log_m1.jsonl 120행, hit 0).
- 결정론: 고정 시드 20260926. `construct.py --verify` → 재빌드 sha256 `71147dbc1af2f0c0b7c5b5b6c170e26e26ca6e4aff1dc3e218a8b27170f173b3` **match=True ok=True, exit 0** — 배치 바이트 핀.
- 파일럿/대조 선정(첫 판사 호출 전, 고정 시드, 기록 완료): 대조 9건 = clean 시드 3 + 결함 클래스별 1씩(6) — 지상진실 라벨 전부 커버(controls_selected.json). 파일럿 20건 = clean 10 + defect 10, 대조 제외 집합에서 시드 선출(pilot_selected.json; REQ-DM-004 n=20, 난이도 진단 전용 — 본 배치 통계에 병합 금지).

**Jev 계측기 (REQ-DM-008)** — `jev_judge.py`: 직접 POST choice 질문 + **필수 criteria 맵 포함**(2026-09-26 ask.sh criteria-누락 422 결함의 수리 형태 — 엔드포인트 api.typesafe.ai/v1/systemone·키 ~/.moai/.env.typesafe·모델 jev-latest 동일), state/instructions 분리, 원시 응답 JSON 호출별 runs/ 보존, `probabilities` 본원·`confidence` 보조 파서(REQ-DM-006). **M1 에서 호출 없음**: `--dry-run`으로 파일럿 항목 t1261-027 요청을 구성·검증 → `errors: []`, criteria 4라벨+설명 존재, state/instructions 분리 확인, 네트워크 호출 0, exit 0. 전송 경로는 M2 전용으로 존재하며 M1 에서 미호출.

**llm 팔 아티팩트 (REQ-DM-009)** — 1차 풀 GLM(z.ai) `glm_task`용: `llm_payloads/` 항목당 완결 프롬프트 120파일(4값 라벨 정의 + [상태] 페이로드 + JSON 응답 형식 {label, confidence, rationale} — **confidence 본원**; 결함 분류학 미노출 중립 지시), `llm_manifest.json`(controls/pilot/main 배치 순서·파일·지상진실), `llm_parse.py` 파서 — `--selftest` 합성 표본 8건: 유효 3형(평문·펜스·산문 감싸짐) 전부 파싱, 불량 5형(잘못된 라벨·confidence 누락·범위 이탈·rationale 누락·비JSON) 전부 기각 — **PASS, exit 0**. glm_task 호출 자체는 M2 레인 오케스트레이터 몫.

**세 상한 선언 (REQ-DM-013 — 첫 판사 호출 전 선언)**:
- turn_cap: 2/항목(초기 1 + 형식 불량 재시도 1; llm+jev 팔의 Jev 단계는 같은 팔 상한 안에 산입)
- call_cap: **2 × (120 + 9 + 20) = 298회/팔-배치**(REQ-DM-013 공식의 커밋 배치 크기 구체화 — 본배치 120 + 대조 9 + 파일럿 20)
- wall_clock_cap: PT8H/배치
- batches: 대조 패스(9, 양팔) + 파일럿 배치(20, 양팔 동일 프로토콜) + 팔별 1차 본배치(llm 120 / llm+jev 120, payload_id 집합 동일 — AC-DM-015); 사전 선언 보충 배치(Opus 창)는 각자 자기 첫 호출 전 자기 상한 별도 선언
- declared_at: **2026-09-26T20:51:27+09:00** (declared_by: manager-develop, card t1261)
- 상한 도달 시: 측정된 것으로 판정서 작성 — 상한 연장·재측정 없음(REQ-DM-013; t1244 선례). 계측기 수리는 t1244 규칙(수리→재시작→이전 출력 무효→무효 행 별도 보존).

**M2 실행 순서(설계 — 레인 오케스트레이터 집행)**: ① criteria_commit 재검증(`git cat-file -e` + is-ancestor + 시간 비교) → ② 전송 페이로드 전부 재스캔(fail-closed) → ③ 판사 대조 9건 × 양팔(llm: GLM / llm+jev: GLM+Jev) — 어느 팔이든 실패 시 양팔 무효 규칙 → ④ 파일럿 20건 × 양팔(난이도 진단; 천장/바닥 퇴화 시 본배치 전 새 criteria 커밋으로 구성 개정, 아니면 진행) → ⑤ 본배치 120건 × 양팔(payload_id 집합 동일 기계 검증) → ⑥ 상한·스캔 행·호출 시각 실시간 기록 → ⑦ M3 분석(McNemar 불일치쌍 정확 이항, 기준선 병기, wrong-automation, 밴드 (a)(b)(c), REQ-DM-007 매핑). GLM 모델 식별자(`glm-5.3`)는 첫 호출 시점에 실측 재확인(드리프트 시 계측기 수리 규칙).

**Gaps (M1)**: 베이스 페이로드는 센서스가 기록한 프로젝션(question_head + 헤더 + 선택지 라벨)이 원문 전체인지 행별 단절 여부는 센서스 추출기가 결정 — 핀된 표현 자체가 이 프로젝션이므로 본 카드 기준선은 이것; clean/defect 베이스 겹침 51(산술 강제, 항목 텍스트 상이) → 판정서 Residual-risk; 스크럽 패턴 집합은 하한(패턴 밖 비밀 형태 통과 가능) → 판정서 Residual-risk; 배치 바이트 핀은 센서스 코퍼스 바이트 + 본 코드 + 시드의 함수(잘리는 코퍼스이므로 sha256 재현 조건 명시); 파일럿 20건의 라벨 스프레드(approve 10/hold 7/modify 3)는 층화 규칙의 산물.

**Residual-risk (M1)**: 주입 결함이 템플릿 형태라 판사가 템플릿을 패턴 매칭하면 인위적으로 잘 나올 수 있음 — 파일럿이 이 난이도를 먼저 진단(REQ-DM-004 파일럿 규칙이 존재한 이유); (i) 증거-제거형 10건이 전부 동일 변형으로 성립(적격 베이스가 충분해 변형 편중) — 클래스 내 다양성은 (ii)-(vi)의 다변형과 비대칭; judge-tool fail-open 시 미측정 기록과 measurement-impossible 경로는 M2 기록이 소유.

### 기준 개정 기록 — 판사 대조 규칙 재정의 (2026-09-27, 운영자 승인)

**무효된 실행 (voided run)**: 첫 M2 실행 — 첫 판사 호출 2026-09-26T21:08:15+09:00, 22 dispatches.
당시 규칙(plan.md M2 「어느 팔이든 대조 실패 시 양팔 무효」)이 발화: 대조 항목 t1261-005에서
llm+jev 합성 팔 오답. 전 9대조 진단 계측(원시 증거 `runs/`, `void_run_diagnostic` 태그 유지·보존):
**Jev 측이 상수 hold 응답자**(9/9 hold, clean 0/3), GLM 측은 산포 정답 5/9.

**왜 바꾸는가 (circular-gate 근거)**: 종전 규칙은 측정의 종속변수(판사 정확도)를 실행의 타당성
게이트에 걸어 있었다 — 판사가 틀리면 실행 자체가 무효가 되는 구조는 「판사가 잘 맞는 경우만
측정할 수 있다」는 자기순환이다. 이 카드의 종속변수가 곧 판사 품질인 이상, 대조의 원래 역할인
계측 경로 검증(전송·파싱·합성 가능성)만 게이트로 남기고 라벨 정확도는 데이터로 공표하는 쪽이
측정의 질문과 정합적이다.

**변경 내용 (v0.2.0, HISTORY 행 동일)**: (1) plan.md M2 — 판사 대조는 양팔 모두 항목별
데이터로 측정·기록·판정서 공표하되 라벨 정확도로 게이트하지 않는다; VOID 게이트는 파이프라인
검증(전송 실패 / 응답이 커밋 형식으로 파싱 불가 — 라벨이 4값 공간 밖·confidence 비수치 / 합성
산출 불가)에만 묶이며 이 경우에도 무효. 파이프라인 양성 대조(추출·라벨 일치, REQ-CALIB-004
형태)는 변경 없이 유지. (2) spec.md REQ-DM-004 — 파일럿 퇴화 판정은 구성 수준 지표(팔 전체의
인구 규모 천장/바닥 — 양팔 ≥90% 또는 ≤10%, 또는 전 결함 클래스가 양팔에서 자명하게 포착/누락)에만
묶인다; 개별 팔의 상수 응답(예: 진단에서 관측된 Jev 상수 hold)은 판정서에 보고되는 측정
결과이지 구성 퇴화도 개정 트리거도 아니다 — 구성이 신호를 갖는 것은 GLM 측 실측(clean 3/3,
결함 포착 2/6, 천장·바닥 아님)이 이미 보여준다. REQ-DM-006의 동결 4요소(비교 술어·팔별
지표·신뢰도 필드·밴드 술어)와 모집단·매핑·풀은 불변.

**운영자 승인**: 2026-09-27 레인 AskUserQuestion 라운드 — 선택지 「기준 재정의 후 재측정
(권장)」 채택.

**재시작 조건**: 재시작 실행은 첫 호출 전 새 `declared_at`으로 세 상한을 재선언하고, **본 개정
커밋을 `criteria_commit`으로** 한다(첫 호출 전 `git cat-file -e` 검증 + 시간 비교 —
REQ-DM-014(a)). 기존 runs/ 증거와 계측기는 무효 실행 기록으로 그대로 보존된다(수정 없음).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
