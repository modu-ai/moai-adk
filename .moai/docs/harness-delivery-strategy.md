# 하네스 제공(Harness Delivery) 전략 — 종합 분석 + 로드맵

> **작성**: 2026-06-03 · ultracode Dynamic Workflow 오케스트레이션 산출 (12 에이전트 / 5 외부연구 / 3 내부진단 / 3 모델 브레인스토밍)
> **트리거**: "다른 프로젝트(MINK)에서 하네스 설치했는데 자동으로 안 됨" + Claude Code Dynamic Workflows 심층 분석 요청 + `https://claude.com/blog/a-harness-for-every-task-dynamic-workflows-in-claude-code` 통합 분석
> **유지보수자 결정 (2026-06-03)**: 모델 C 단계적 채택 — **Lane 1+3 먼저, Lane 2 deferred**. #1 NAMESPACE-V2(BLOCKING) → #2 ACTIVATION-WIRING → #3 MANIFEST-REFRESH 순.
> **분류**: 설계/의사결정 기록(decision record). 본 문서는 §15 언어 중립성·§25 template isolation 대상이 아닌 maintainer-local 전략 문서다.

---

## 핵심 요약

가장 중요한 결론: **MoAI 하네스가 "생성은 되지만 자동 활성화되지 않는" 근본 원인은 디스크립션 품질이나 스킬 부재가 아니라, 단 두 개의 시스템적 배선 단절** — (1) `<!-- moai:harness-start -->` CLAUDE.md 라우팅 마커가 5개 표본 프로젝트 + 템플릿 전부에서 0개(설치 SPEC은 `completed`이나 dead-code), (2) `harness-*` 정규 prefix vs `my-harness-*` 실제 생성 prefix의 doctrine-code drift로 정규 lookup이 항상 빈 결과를 반환 — 에서 발생한다. 사용자의 4개 가설 중 **"빈 디스크립션"은 반증**됨(MINK 9개 에이전트 모두 충실한 트리거 텍스트 보유)을 확인했다.

권장 모델: **C-hybrid-layered("Harness Router: 3 실행 레인")의 단계적 채택** — 단, **Lane 1(대화형 auto-load)+Lane 3(learning loop)을 먼저 출고하고 Lane 2(workflow)는 좁은 read-only 감사용으로만 후순위 가산**한다. Anthropic 자체 원칙("대부분의 coding task는 진짜 병렬화 가능한 subtask가 적다", "subagent definition이 도메인 전문성의 유일한 영속·배포 가능 패키징 단위")이 coding-heavy 하네스 작업을 workflow로 옮기는 것을 명시적으로 반대하기 때문이다. **단, 어떤 모델도 `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`(prefix 통일 catch-up SPEC)이 먼저 착지하기 전에는 출고 불가** — 지금 생성하면 `moai update`가 산출물을 삭제한다.

---

## 1. Claude Code Dynamic Workflows 심층 분석

### 1.1 무엇인가

Dynamic Workflow는 **subagent를 대규모로 오케스트레이션하는 JavaScript 스크립트**다. Claude가 기술된 task에 맞춰 그 스크립트를 직접 작성하고, 런타임이 백그라운드 격리 환경에서 실행하는 동안 세션은 응답성을 유지한다. Research preview이며 **Claude Code v2.1.154+ 필요**. Pro에서는 `/config`의 "Dynamic workflows" 행으로 활성화.

**핵심 차별점 — 누가 plan을 쥐는가**: subagent/skill/agent-team에서는 **Claude가 turn-by-turn 오케스트레이터**이며 모든 중간 결과가 컨텍스트 윈도우에 누적된다. Workflow는 **plan을 코드로 옮긴다** — 스크립트가 루프·분기·중간 결과를 보유하므로 Claude의 컨텍스트에는 **최종 답만** 돌아온다.

### 1.2 "Claude가 workflow를 작성한다"는 어떻게 작동하나

1. **프롬프트로 요청**: 키워드 `ultracode` 포함 → Claude가 turn-by-turn 대신 스크립트를 작성.
2. **Claude가 결정하게 함**: `/effort ultracode` — Claude가 모든 substantive task마다 workflow를 계획.

### 1.3 스크립트 primitive

| Primitive | 동작 |
|-----------|------|
| `agent(prompt, opts)` | leaf subagent 1개 spawn. `schema`(구조화출력 검증→자동 재시도), `label`, `phase`, `agentType`, `model`. leaf만 토큰 소비, 각자 clean 컨텍스트 |
| `parallel(...)` | fan-out 배리어 — thunk(`() => agent(...)`) 받음 |
| `pipeline(...)` | 출력→입력 스트림, 배리어 없음. 공식: "pipeline을 디폴트, 한 stage가 모든 선행 결과 필요할 때만 parallel 배리어" |
| `phase` / `log` / `args` / `budget` | 진행그룹 / narrator / JSON입력 / 토큰타겟(깊이 스케일) |

**META 블록**: pure literal 필수(resume journaling cache key 보호).

### 1.4 결정성 제약 (verbatim)

`Date.now()`, `Math.random()`, argless `new Date()`는 workflow 내에서 **THROW**. 우회: 타임스탬프는 `args` 주입, 다양성은 index 변주. 스크립트는 fs/shell 직접 접근 없음 — 읽기/쓰기/명령은 `agent()` 프롬프트 **내부**.

### 1.5 실행 모델 / 권한 / resumability

- **한도**: 동시 16 / 총 1000 에이전트. mid-run user input 불가.
- **권한**: subagent는 항상 acceptEdits + allowlist 상속. allowlist 밖 shell/web/MCP는 mid-run prompt 가능.
- **resume**: 같은 세션 내에서만 — 완료 에이전트 캐시, 나머지 live. 세션 종료 시 fresh 재시작.
- **저장**: `/workflows`→`s`→`.claude/workflows/`(프로젝트, 충돌 시 개인 이김) — `/<name>` 슬래시 커맨드화.
- **`/deep-research`**: 번들 workflow(fan-out search→교차검증→투표→cited 리포트), WebSearch 필요.
- **`/effort ultracode`**: xhigh + 자동 오케스트레이션. 토큰↑. 세션만 지속, 새 세션 reset.

---

## 2. 두 가지 harness 의미의 재조정 (Anthropic per-task vs MoAI project)

> `ext:harness-blog`("A harness for every task", Thariq Shihipar & Sid Bidasaria, Anthropic)는 fetch 성공으로 verbatim 인용 확보.

### 2.1 정의 충돌

| 축 | **MoAI "harness"** (revfactory) | **Anthropic "harness"** (블로그) |
|----|--------------------------------------|--------------------------------------|
| 정의 | `/moai project` 시 1회 생성되는 영속 프로젝트 agent team + skills | Claude가 task마다 즉석 작성하는 per-task 오케스트레이션 스크립트 |
| 산출물 | `.claude/agents/harness/*` + `harness-*` skills + `.moai/harness/main.md` | `.claude/workflows/<name>.js` (선택적 저장) |
| 수명 | 영속(auto-trigger 의도) | ephemeral, resumable, token-budgeted |
| verbatim | — | *"Claude can now write its own harness on the fly, custom-built for the task at hand."* |

### 2.2 수렴 — 둘은 같은 적을 공격

블로그 thesis: 정적/사전구축 harness는 generic하지만, dynamic workflow는 use case 맞춤 custom harness를 작성해 **single-context 3대 실패 모드**를 격파 — (1) agentic laziness("50개 중 35개에서 멈춤"), (2) self-preferential bias(자기검증 시 자기 findings 신뢰), (3) goal drift(edge-case 규칙 유실). **결정적 수렴점**: 블로그 패턴(adversarial verification, fan-out→synthesize, generate-and-filter, loop-until-done) = MoAI `plan-auditor`/`sync-auditor` 모델 그 자체. `/deep-research`는 MoAI가 이미 번들 출고 중인 블로그의 canonical 예시.

### 2.3 충돌 — MoAI harness는 블로그가 반대하는 쪽에 가깝다

MoAI harness는 정적·영속·auto-trigger team으로, 블로그가 대비시키는 "generic, pre-built" 쪽이다. 외부 findings: **"도메인 전문성의 영속·배포 가능 패키징 단위는 오직 subagent definition"** — workflow의 재사용 단위는 "오케스트레이션 스크립트"이지 "도메인 지식"이 아니다.

### 2.4 재조정 결론

블로그 reuse 가이드(verbatim): 저장 harness는 *"verbatim 실행 스크립트가 아니라 재사용 가능한 TEMPLATE"* 로 다뤄라.

> **재조정 명제(layering thesis)**: MoAI의 정적 project harness는 dynamic workflow를 **방출/등록**하되, workflow가 정적 team을 **대체하지 않고 그 위에 task-shaped·savable template으로 적층**되어야 한다.

→ 모델 B(치환)는 원칙 위반으로 배제, 모델 A(workflow layer 없음)는 견고하나 미완, 모델 C(적층)만 명제를 정확히 구현. 비용 caveat(verbatim): *"Dynamic workflows often use more tokens", "best suited for complex, high value tasks"*.

---

## 3. 3 오케스트레이션 primitive와 하네스의 위치

| 축 | **Sub-agents** | **Agent Teams** | **Dynamic Workflows** |
|----|---------------|------------------|------------------------|
| plan 소유 | Claude turn-by-turn | 고정 lead + 공유 task list peer | 스크립트 |
| 중간결과 | fresh isolated 컨텍스트→summary | 각 teammate 컨텍스트 + 공유 list | 스크립트 변수→최종답만 |
| 규모 | turn당 소수 | 3-5 teammate 권장 | 수십~수백/run (동시16/총1000) |
| 재사용 단위 | **subagent definition 파일** | composition(ephemeral) | 오케스트레이션 스크립트 |
| 핵심 제약 | subagent는 subagent spawn 불가; AskUserQuestion 불가 | 중첩 불가, 고정 lead, experimental | mid-run input 불가, 결정적 body 필수 |
| 영속/배포성 | **높음**(버전관리 파일) | 낮음 | 중간 |

**라우팅**: workflow=수십~수백 read-only 독립 fan-out / Agent Teams=소수(3-5) 장기 peer 조율 / sequential subagent=coding-heavy 디폴트. **하네스 본체는 subagent definition(+skills+memory)에 두고, 대규모 sweep만 workflow, 병렬 cross-layer만 Agent Teams.**

---

## 4. 현재 MoAI 하네스 제공 방식과 자동 트리거 실패 진단

### 4.1 meta-harness 7-Phase 생성

`moai-meta-harness/SKILL.md` 정의. Phase 4-5(skeleton+customization)만 방출 소유. **trigger 전제조건(SKILL.md:60-61)**: (a) `.moai/harness/main.md` 존재 AND (b) CLAUDE.md `<!-- moai:harness-start -->` 마커 — 마커는 `SPEC-V3R3-PROJECT-HARNESS-001` 소유.

### 4.2 broken wiring chain (ground-truth 검증)

```
Step 2 (trigger 전제): 런타임이 CLAUDE.md 마커 탐지 → ★ 단절, SYSTEMIC ★
Step 3 (entry): .moai/harness/main.md 읽기 → MINK 단절(부재)
Step 4 (agent 발견): 디스크립션 매칭 → INTACT(충실)
```

### 4.3 4개 가설 vs ground truth

| 가설 | 판정 | 근거 |
|------|------|------|
| (a) 빈 agent 디스크립션 | **반증** | MINK 9개 모두 충실. 근본원인 아님 |
| (b) harness-* skills 부재 | **확인(nuance)** | 정규 prefix 0/5, 단 `my-harness-*`로 존재 — drift 결과 |
| (c) CLAUDE.md 마커 부재 | **확인, SYSTEMIC** | 5개 프로젝트 + 템플릿 = 전부 0개. SPEC `completed`인데 0 전파 = dead-code. **PRIMARY killer** |
| (d) main.md 부재 | **MINK만, non-systemic** | 4개 peer는 보유. MINK만 Phase-4 불완전 |

### 4.4~4.6 추가 systemic 근본원인

- **doctrine-code drift**(SKILL.md:168): `harness-*` 선언 vs `my-harness-*` 방출 → 정규 lookup 항상 빈 결과. **`SPEC-V3R6-HARNESS-NAMESPACE-V2-001`(Tier M) 착지 전 생성 PROHIBITED**.
- **cross-project 오염**: MINK+fortune-table에 유지보수자 자신의 `moai-harness-*-specialist`(moai-adk-go Go CLI 기술) 누출 → MINK에서 부재 skill 참조하는 dangling ref. `moai update`가 user-owned namespace 침범 증거.
- **결론**: gap은 systemic. MINK가 main.md를 갖춰도 Step 2(마커)에서 여전히 죽음 → 마커 수정이 최우선 leverage.

---

## 5. 브레인스토밍: 3개 제공 모델 비교

| 차원 | **A — trigger-fix** | **B — Workflow-native** | **C — Hybrid Router** |
|------|---------------------|-------------------------|------------------------|
| 메커니즘 | 기존 모델 유지 + 5개 배선 수리 | meta-harness가 `.claude/workflows/<h>.js` 방출, `/<h>` 호출 | main.md→live ROUTER, task-shape별 3레인 |
| auto-trigger 해법 | 마커를 실제 Go 코드로 설치 | trigger=슬래시 커맨드 이름(매칭 불필요화) | 마커 설치를 Phase 4로 이전 + skills preload + smoke gate |
| 견고성 | 의미적 매칭 의존(잔존 risk) | 구조적(가장 견고) | Lane1 의미적/Lane2 결정적 |
| Anthropic 정합 | ⚠️ standing roster 충돌 | ❌ run-phase 위반 / ✅ read-only만 | ✅ 레인별 best primitive 매칭 |
| 템플릿 충돌 | 없음 | 심각(`.claude/workflows/` user-owned) | 중간(Lane2만 새 계약) |
| effort | Tier L(+M 선행) | Tier L | Tier L(단계 분해) |

**A**: 최저 disruption, PRIMARY killer 직격. 약점: standing roster가 최적인지 회피 + installer dead-code 재발 방지 smoke test 필수. **B**: auto-trigger를 가장 우아하게 해소하나 wrong-shape(interactive 생성 + coding-heavy 사용) + 채널 충돌 → 본체 불가. **C**: 재조정 명제 정확 구현, 마커 설치를 생성 시점으로 이전 = 최고 leverage. 약점: 부품 多 + Lane2가 최대 리스크/최소 payoff → Lane1+3 우선, Lane2 지연.

---

## 6. 권장 모델과 근거

### 권장: 모델 C 단계적 — "Lane 1+3 먼저, Lane 2는 좁은 read-only 감사로만 후순위"

사실상 A의 견고한 trigger-fix 본체 + C의 라우터 추상화 + B의 workflow를 좁은 슬롯에만.

1. **§2.4 재조정 명제 정합**: project harness가 workflow를 방출/등록하되 대체 않고 적층 = C 구조. B는 "영속 패키징=subagent definition" 원칙 위반.
2. **Anthropic 원칙 준수**: coding=병렬 subtask 적음→Lane1 sequential / 15x 토큰비용→Lane2는 high-value read-only만 / 3-5 ceiling+option-flooding→specialist 수 적게 / "repetition not anticipation"→Lane3 관찰 후 정련.
3. **최고-leverage 수정 정조준**: 마커 설치를 phantom SPEC→Phase 4 이전 + Phase 6 smoke gate.
4. **정직한 trade-off**: Lane2는 진짜 doctrine 리스크+좁은 payoff→deferred. Lane1+3만으로 문제 완전 해소. 공통: NAMESPACE-V2 선행 없으면 출고 불가(sequencing dependency, optional 아님).

---

## 6.5 Ground-truth 정정 (2026-06-03, SPEC 위임 전 검증)

> 본 절은 §6/§7의 sequencing을 **부분 정정**한다. 보고서 synthesis는 doctrine 텍스트(`meta-harness SKILL.md:168`)에 근거해 #1 NAMESPACE-V2를 "출고 BLOCKING 선행"으로 제시했으나, SPEC 위임 전 기존 SPEC + Go 코드 실측에서 다음이 확인됨:

1. **`moai update` 보호는 이미 구현·테스트 완료** — `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`(PR #1048, commit `767bc04a4`) + `SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001`(completed). `internal/cli/update_namespace_protect.go` + `update.go isUserOwnedNamespace` + 14 테스트. 제안 #1 scope의 "보호 확립"은 **중복**.
2. **현 generator는 보호되는 `my-harness-*` 방출** — 코드가 `my-harness-*` 생성·보호 → **현재 삭제 위험 없음**. "지금 생성하면 moai update가 삭제" 전제는 거짓(generation이 `my-harness-*`에 머무는 한).
3. **진짜 drift는 latent 명명 split** — commit `66a3d53be`("Phase 1 doctrine-only")가 doctrine만 `harness-*`로 변경, 코드+generator는 `my-harness-*`. `SKILL.md:168`: *"새 harness-* prefix로 실제 generation 금지 — protection 없음."* → active blocker 아님.

### 정정된 sequencing

- **#2 ACTIVATION-WIRING이 1순위** (사용자 실제 고충 auto-trigger 해결). 기존 보호 `my-harness-*` prefix로 즉시 안전 출고, namespace 결정과 독립. **BLOCKING 의존 없음.**
- **#1 NAMESPACE-V2는 BLOCKING 아님** — doctrine/code 명명 split 해소(별도 정리) + 방향 결정(코드→`harness-*` 완성 vs doctrine→`my-harness-*` 롤백) 필요. 보고서 scope의 "보호 확립"은 이미 완료이므로 제외.
- §7 표의 Priority/순서는 본 정정으로 갱신됨 (#2 우선, #1은 방향 결정 대상 별도).

---

## 7. 구체적 다음 단계 (SPEC 후보)

| 순서 | SPEC 후보 | Tier | scope | Priority |
|------|-----------|------|-------|----------|
| 1 (BLOCKING) | **SPEC-V3R6-HARNESS-NAMESPACE-V2-001** | M | `harness-*` vs `my-harness-*` 통일(~39 Go 파일+30 테스트) + `moai update` preserve+backup 보호 + maintainer harness 누출 차단 | High — 없으면 생성물 삭제됨 |
| 2 | **SPEC-HARNESS-ACTIVATION-WIRING-001** (Lane1+3) | S | Phase 4 마커 설치 + trigger 디스크립션 + `skills:` preload 의무 + Phase 6 smoke gate | High — 최고 leverage, 저비용 |
| 3 | **SPEC-HARNESS-MANIFEST-REFRESH-001** | S | main.md를 stale 기술문서→task-shape→lane ROUTER로 재작성 + 아카이브 agent 참조 purge | Medium — #2와 병합 가능 |
| 4 (DEFERRED) | **SPEC-HARNESS-WORKFLOW-LANE-001** (Lane2) | M | `builder-harness` artifact_type=workflow + `.claude/workflows/` namespace 계약 + `/deep-research`-스타일 PoC | Low — 최대 리스크/최소 payoff |

**핵심 sequencing**: #1 → (#2+#3 병합 가능) → #4.

---

## 8. Context-Governance Axis — eager-vs-on-demand weight drift alarm

> **도입 배경**: SPEC-V3R6-CONTEXT-GOV-AXIS-001 (Epic 15 P2b, harness-books 적용 에픽).
> **외부 근거**: book2 ch8.3 (signal dilution) + diag-05 (three context-governance paths).
> **Tier-2 추가 신호**: 본 절은 기존 harness-learning Tier system (1-4)을 **변경하지 않고** Tier-2 제안으로 feed되는 **additive drift SIGNAL**이다 (REQ-CGA-006).

### 8.1 문제 — signal dilution (book2 ch8.3)

> *"the model sees more, but is not necessarily clearer about which working semantics matter next."* — book2 ch8.3

하네스가 성숙할수록 저항이 가장 적은 경로는 eagerly-loaded 컨텍스트(`CLAUDE.md` + 자동 로드 `.claude/rules/moai/*.md` ~60개 + output-style `moai.md` + auto-memory index `MEMORY.md`)에 더 많은 부트스트랩/스킬/식별/규칙 텍스트를 끼워넣고 **사후 truncation**에 기대는 것이다 ("inject first, rescue later"). 깊은 비용은 **signal dilution**이다 — 모델은 더 많이 보지만, 다음에 어떤 작업 의미가 중요한지 더 명확해지지 않는다.

diag-05는 세 context-governance path를 구분한다: (a) **budgeted working memory** (건강), (b) **structured context units**, (c) **prompt-stacking** (OpenClaw foil). moai-adk는 skills에 token budget가 있지만 (`skillListingBudgetFraction` ~1%, compaction ~25K, progressive disclosure L1/L2/L3) **eagerly-loaded 컨텍스트에 대한 동등한 budget나 audit이 없다**. 기존 `session-handoff.md` thresholds (1M=50% / 200K=90%)은 **REACTIVE** (언제 `/clear`할지)이지 **PREVENTIVE** (얼마나 주입해야 하는지)가 아니다.

### 8.2 기계적 drift 게이트 — 증상 테스트 (book2 ch8.3)

> *tokens burn fast, quality doesn't climb as context fattens.* — book2 ch8.3 symptom test

게이트: **eagerly-loaded weight가 N consecutive session에 걸쳐 단조 증가** AND **대응하는 skill-budget 감소가 없는 경우** (즉, 규칙은 추가됐으나 on-demand skill로 강등되지 않음), **Tier-2 harness-learning proposal**을 surface한다 ("eager context X% 증가 — on-demand skill로 강등 후보 규칙").

### 8.3 세션 카운트 임계값 N

**N = 3** (문서화된 범위 **N ∈ [3, 5]** 중 최솟값 — 가장 빠른 정직한 신호; 3-session 단조 상승은 이미 실제 패턴이지 noise가 아니다). N을 3 이상으로 bound하면 alarm이 N=1에서 자명하게 발화하지 않고 (단일 session 성장은 noise), 5 이하로 bound하면 반응성을 유지한다 (5 session 이상의 silent growth는 이미 실제 토큰 비용을 지불한 상태).

> N은 본 절 상단에 named constant로 고정되어 있어 reviewer에게 가시적이며, re-spec 없이 조정 가능하다 (REQ-CGA-004).

### 8.4 recording — eager-vs-on-demand weight (REQ-CGA-001/002/003)

Layer A Observer hook pipeline (`internal/harness/observer.go` + `internal/cli/hook.go` `runHarnessObserve*`)가 매 turn마다 `.moai/harness/usage-log.jsonl`에 eager-vs-on-demand weight를 기록한다. schema_version이 `"v2"` → `"v2.1"`로 bump되었으므로 reader는 두 case를 구분할 수 있다:

- **case (a)** — `schema_version ∈ {"v1", "v2"}`: 본 SPEC 이전 line (weight 측정 안 됨 — legacy weight-absent).
- **case (b)** — `schema_version == "v2.1"` + weight fields = sentinel (`0`/`""`): new binary가 line을 썼으나 estimation skip (fail-open path).

fail-open (REQ-CGA-003): weight 측정 실패 시 hook은 stderr에 warning을 쓰고 exit 0한다 (절대 exit 2 아님). 이는 P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001의 death-spiral avoidance 속성을 보존한다.

### 8.5 Tier system 비변경 (REQ-CGA-006)

본 alarm은 기존 harness-learning Tier system (Tiers 1-4)에 **additive drift SIGNAL**로서 Tier-2 제안을 feed한다. Tier 정의를 추가·삭제·renumber하지 않는다. **자동 강등은 out-of-scope** (§X.1) — alarm은 제안을 surface할 뿐, 강등은 human-in-the-loop harness-learning 결정으로 남는다 (Tier-2 proposal → human 승인 → 별도 future SPEC이 강등 실행).

---

## Sources

- https://claude.com/blog/a-harness-for-every-task-dynamic-workflows-in-claude-code (fetch 성공, verbatim 인용)
- https://code.claude.com/docs/en/workflows
- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/agent-teams
- https://www.anthropic.com/engineering/built-multi-agent-research-system
- (보조 커뮤니티 소스) alexop.dev workflow primitive 시그니처, ray-amjad `workflow-creator` 스킬 — primitive 시그니처 교차참조용
