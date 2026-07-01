# Codex-maxxing 백서 기반 MoAI-ADK Long-Running Work 구조 분석

> **메타**
> - 작성일: 2026-06-24
> - 기준 HEAD: `7d27ca427`
> - 성격: 탐색적 분석 보고서 (SPEC 아님 — 본 문서는 작업 분해·SPEC 수립을 의도하지 않음)
> - 입력 출처: Codex-maxxing 백서 요약 + Claude Code 공식문서(B) + Anthropic 엔지니어링 블로그(C) + 3rd-party 프레임워크(D) 교차 조사 + MoAI-ADK 룰 인벤토리
> - 정합 원칙: 본 문서는 verification-claim-integrity.md에 따라 입력 JSON에 근거한 사실만을 서술하며, 출처가 불확실한 항목은 "— (출처 미확보)"로 명시

---

## 1. 백서 핵심 원칙 요약

Codex-maxxing 백서의 핵심 전제는 **"Codex는 persistent workspace(시간 단위로 측정)이며, 단일 프롬프트를 operating loop로 전환한다"**는 것이다. 작업은 일회성 질의가 아니라 지속적으로 실행되는 루프로 재구성되며, 이 루프를 장시간 안정적으로 유지하기 위한 3개의 전략 레버가 제시된다.

### 3개 전략 레버 (three levers)

1. **verifiable-steps (검증 가능한 단계로의 분해)** — 작업을 검증 가능한 단계로 분해한다. 핵심 명제는 "검증은 프롬프트가 아니라 실행 사슬의 아키텍처 결정"이라는 점이다. 즉 모델에게 "잘 검증해"라고 지시하는 것으로는 부족하며, 실행 구조 자체가 각 단계의 pass/fail을 기계적으로 판정할 수 있어야 한다.
2. **cross-session continuity (세션 간 연속성)** — summaries, handoff artifacts, checkpointing을 1급 운영 관심사(first-class operational concern)로 격상시킨다. 세션은 한 번에 끝나는 단위가 아니라, 다음 세션이 이어받을 수 있는 상태를 남기는 단위로 설계된다.
3. **delegate-vs-oversee (위임 vs 감독의 분리)** — 복잡도와 위험도에 기반한 라우팅을 수행한다. 위임(delegate)과 감독(oversee)을 의식적으로 분리하여, 고비용·고위험 작업은 감독하고, 단순·반복 작업은 위임한다.

### 인프라 정합 (infra)

- **token budget = session-level problem**: 토큰 예산은 단일 호출이 아닌 세션 단위 문제이며, correlation ID로 추적된다.
- **context degradation = routing signal**: 컨텍스트 저하는 라우팅 신호로 사용된다(75% 임계값 headroom routing).
- **parallel delegation = parallel provider dispatch**: 병렬 위임은 곧 병렬 provider 디스패치다.
- **verification gates = human-in-the-loop policy**: 검증 게이트는 human-in-the-loop 정책으로 모델링된다(가벼운 모델로 검증하는 verification-tier model routing 포함).
- **routing policies**: session-scoped cost caps, context-window headroom routing(75%), verification-tier model routing.

### AGENTS.md (persistent context file)

백서는 모델에게 배운 것과 진척을 기록하도록 지시하는 persistent context 파일 `AGENTS.md`를 강조한다. 두 계층으로 운영된다:
- **전역 AGENTS.md** = global memory / toolchain / preferences
- **프로젝트 루트 AGENTS.md** = project rules / testing / protected dirs

### Session compaction

긴 대화를 compaction으로 살려두어 thread crash를 방지한다. 이는 연속성 메커니즘의 한 축이다.

### 출처

- 공식 랜딩: https://openai.com/index/codex-maxxing-long-running-work/
- TheRouter.ai 분석: https://therouter.ai/news/openai-codex-maxxing-long-running-work-operator-routing/
- Jason Liu 원본: https://jxnl.co/writing/2026/05/10/codex-maxxing/
- PDF(직접 패치 불가 — Gaps 명시): https://cdn.openai.com/pdf/8a9f00cf-d379-4e20-b06f-dd7ba5196a11/OAI_WhitePaper_Codex-maxxing26.pdf

---

## 2. 베스트 프랙티스 교차 조사 결과

4개 그룹(B1: CC orchestration, B2: CC verification/hooks, C1: Anthropic agents, C2: Anthropic context-eng, D1: 3rd-party frameworks)에서 확보한 패턴을 정리한다. 각 패턴은 출처 URL과 증명 인용문(evidence_quote)을 포함한다.

### 2.1 B1: Claude Code orchestration (sub-agents / workflows / agent-teams / goal)

| 패턴 | 출처 | 핵심 증명 인용 |
|------|------|----------------|
| Subagent context isolation | sub-agents | "the subagent does that work in its own context and returns only the summary" |
| Subagent resumption by agent ID | sub-agents | "Resumed subagents retain their full conversation history... If a stopped subagent receives a `SendMessage`, it auto-resumes in the background" |
| Subagent persistent memory (`memory` frontmatter) | sub-agents | "The `memory` field gives the subagent a persistent directory that survives across conversations" |
| Subagent worktree isolation (`isolation: worktree`) | sub-agents | "Set `isolation: worktree` to run the subagent in a temporary git worktree... branched by default from your default branch" |
| Nested subagents (v2.1.172+) | sub-agents | "a subagent can spawn its own subagents... so the intermediate output never reaches your main conversation" |
| Agent Teams shared TaskList (file-lock claim) | agent-teams | "Task claiming uses file locking to prevent race conditions when multiple teammates try to claim the same task simultaneously" |
| Teammate gating hooks (TeammateIdle / TaskCreated / TaskCompleted exit-2) | agent-teams | "`TeammateIdle`... Exit with code 2 to send feedback and keep the teammate working" |
| Team/task state persistence (session resume) | agent-teams | "The task list directory persists locally and is never uploaded, so resumed sessions keep their tasks" |
| Plan-approval gate (lead-approves-teammate) | agent-teams | "the teammate works in read-only plan mode until the lead approves their approach" |
| Adversarial fan-out (defeat anchoring bias) | agent-teams | "Sequential investigation suffers from anchoring... With multiple independent investigators actively trying to disprove each other, the theory that survives is much more likely to be the actual root cause" |
| Dynamic Workflow (plan in script variables) | workflows | "Intermediate results stay in script variables instead of landing in Claude's context" |
| Dynamic Workflow (resumable runs, cached + live) | workflows | "agents that already completed return their cached results, and the rest run live" |
| Dynamic Workflow (adversarial cross-checking in code) | workflows | "it can have independent agents adversarially review each other's findings before they're reported" |
| Dynamic Workflow (resource caps: 16 concurrent / 1000 total) | workflows | "Up to 16 concurrent agents... 1,000 agents total per run [Prevents runaway loops]" |
| /goal (model-evaluated completion) | goal | "a small fast model checks whether the condition holds. If not, Claude starts another turn instead of returning control" |
| /goal resume restoration (--resume/--continue) | goal | "A goal that was still active when a session ended is restored when you resume that session" |
| /goal + auto-mode (unattended long runs) | goal | "auto mode removes per-tool prompts, and `/goal` removes per-turn prompts" |

### 2.2 B2: Claude Code verification / hooks

| 패턴 | 출처 | 핵심 증명 인용 |
|------|------|----------------|
| Deterministic Stop hook (8-block cap + stop_hook_active) | best-practices | "a Stop hook runs your check as a script and blocks the turn from ending until it passes. Claude Code overrides the hook and ends the turn after 8 consecutive blocks" |
| Prompt-based Stop hook (LLM evaluation) | hooks | "the reason is fed back to Claude as its next instruction and the turn continues" |
| Agent-based hook (multi-turn tool-access verification) | hooks | "An agent hook spawns a subagent that can use tools like Read, Grep, and Glob to verify conditions before returning a decision" |
| Aggressively manage context window (compaction + /clear) | best-practices | "After two failed corrections, /clear and write a better initial prompt... A clean session with a better prompt almost always outperforms a long session with accumulated corrections" |
| Subagents to preserve context (investigation + verification) | best-practices | "Subagents run in separate context windows and report back summaries" |
| Adversarial review via fresh-context subagent (Writer/Reviewer) | best-practices | "A reviewer running in a fresh subagent context sees only the diff and the criteria you give it, not the reasoning that produced the change" |
| TaskCompleted / TeammateIdle hooks (exit-2 block) | hooks | "When a TaskCompleted hook exits with code 2, the task is not marked as completed and the stderr message is fed back to the model" |
| Explore first, then plan, then code (4-phase) | best-practices | "Use plan mode to separate exploration from execution" |
| Provide verification criteria (pass/fail check) | best-practices | "Give Claude something that produces a pass or fail, and the loop closes on its own" |
| Surgical / concise CLAUDE.md (prune or convert to hook) | best-practices | "If your CLAUDE.md is too long, Claude ignores half of it... delete it or convert it to a hook" |

### 2.3 C1: Anthropic engineering (agents)

| 패턴 | 출처 | 핵심 증명 인용 |
|------|------|----------------|
| Orchestrator-Workers (동적 서브태스크 분할) | multi-agent-research-system + building-effective-agents | "a central LLM dynamically breaks down tasks, delegates them to worker LLMs, and synthesizes their results" |
| 두 종류의 병렬화 (subagent spawn + 도구 호출) | multi-agent-research-system | "the lead agent spins up 3-5 subagents in parallel... subagents use 3+ tools in parallel. These changes cut research time by up to 90%" |
| Subagent 분리로 컨텍스트 압축 | multi-agent-research-system | "The essence of search is compression... Subagents facilitate compression by operating in parallel with their own context windows" |
| Initializer Agent + Coding Agent (세션 경계) | effective-harnesses-for-long-running-agents | "an initializer agent that sets up the environment on the first run, and a coding agent... leaving clear artifacts for the next session" |
| 구조화된 Feature List + increment-only 원칙 | effective-harnesses-for-long-running-agents | "It is unacceptable to remove or edit tests because this could lead to missing or buggy functionality" |
| 세션 시작 'get bearings' 루틴 + git 복구 | effective-harnesses-for-long-running-agents | "1. Run `pwd`... 2. Read the git logs and progress files... 3. Read the features list file and choose the highest-priority feature that's not yet done" |
| End-state 평가 (다중 턴 상태 변이 검증) | multi-agent-research-system | "focus on end-state evaluation rather than turn-by-turn analysis... evaluate whether it achieved the correct final state" |
| Subagent 산출물 파일시스템 직접 기록 | multi-agent-research-system | "implement artifact systems where specialized agents can create outputs that persist independently... pass lightweight references back" |
| Brain-Hands-Session 3-way 분리 (cattle화) | managed-agents | "We virtualized the components of an agent: a session... a harness... and a sandbox... implementation of each to be swapped without disturbing the others" |
| 세션 = 컨텍스트 창 외부 객체 (getEvents 슬라이스) | managed-agents | "the session provides this same benefit, serving as a context object that lives outside Claude's context window" |
| Rainbow 배포 (실행 중 에이전트 보호) | multi-agent-research-system | "we use rainbow deployments to avoid disrupting running agents, by gradually shifting traffic" |
| Errors compound + resume-from-where-failed | multi-agent-research-system | "we built systems that can resume from where the agent was when the errors occurred... deterministic safeguards like retry logic and regular checkpoints" |
| Effort scaling to query complexity | multi-agent-research-system | "Simple fact-finding requires just 1 agent with 3-10 tool calls... complex research might use more than 10 subagents" |

### 2.4 C2: Anthropic context engineering

| 패턴 | 출처 | 핵심 증명 인용 |
|------|------|----------------|
| Context rot / attention budget | effective-context-engineering-for-ai-agents | "as the number of tokens in the context window increases, the model's ability to accurately recall information from that context decreases... Context, therefore, must be treated as a finite resource with diminishing marginal returns" |
| Compaction (in-place summarization) | effective-context-engineering-for-ai-agents | "Compaction is the practice of taking a conversation nearing the context window limit, summarizing its contents, and reinitiating a new context window... Start by maximizing recall... then iterate to improve precision" |
| Context reset (clean-slate handoff vs. compaction) | harness-design-long-running-apps | "Context resets—clearing the context window entirely and starting a fresh agent, combined with a structured handoff... This differs from compaction... A reset provides a clean slate, at the cost of the handoff artifact having enough state" |
| Structured note-taking / agentic memory | effective-context-engineering-for-ai-agents | "the agent regularly writes notes persisted to memory outside of the context window... After context resets, the agent reads its own notes and continues multi-hour training sequences" |
| Sub-agent architecture (distilled handback) | effective-context-engineering-for-ai-agents | "Each subagent might explore extensively... but returns only a condensed, distilled summary (often 1,000-2,000 tokens)" |
| Just-in-time context retrieval | effective-context-engineering-for-ai-agents | "agents built with the \"just in time\" approach maintain lightweight identifiers... and use these references to dynamically load data into context at runtime" |

### 2.5 D1: 3rd-party frameworks (LangGraph / CrewAI / AutoGen)

| 프레임워크 | 패턴 | 출처 | 핵심 증명 인용 |
|-----------|------|------|----------------|
| LangGraph | Checkpointer (per-superstep thread-scoped) | persistence | "Checkpointers persist a thread's graph state as checkpoints. Use them for short-term, thread-scoped memory" |
| LangGraph | Store (cross-thread long-term) | persistence | "Stores persist application-defined data outside the graph state. Use them for long-term, cross-thread memory" |
| LangGraph | HITL interrupt + resume (indefinite pause) | human-in-the-loop | "Interrupts use LangGraph's persistence layer... to indefinitely pause graph execution until you resume" |
| CrewAI | Crew-level Checkpointing (from_checkpoint) | crews (v1.14.7) | "Checkpointing lets a crew automatically save its state after key events... resumed exactly where they left off without re-executing completed tasks" |
| CrewAI | Memory subsystem (short/long/entity) | crews | "Crews can utilize memory (short-term, long-term, and entity memory) to enhance their execution and learning over time" |
| CrewAI | Replay-from-task (CLI) | crews | "`crewai replay -t <task_id>`... retaining context from previously executed tasks" |
| AutoGen | Termination conditions (bounded run) | termination | "a run can go on forever... This is the role of the termination condition... They are stateful but reset automatically after each run" |
| AutoGen | Termination auto-reset (continuation) | termination | "termination conditions automatically reset after each run or run_stream call, allowing the team to resume its conversation from where it left off" |
| AutoGen | save_state/load_state (explicit serialization) | state | "save_state on a team, it saves the state of all the agents in the team... State is a dictionary that can be serialized to a file or written to a database" |

---

## 3. MoAI-ADK 현재 long-running 구조 인벤토리

MoAI-ADK가 이미 보유한 long-running 관련 메커니즘을 레버 정합도 기준으로 정리한다.

### 3.1 continuity (세션 간 연속성) 정합 메커니즘

| 메커니즘 | 파일 | 기능 | 정합도 |
|----------|------|------|--------|
| Graduated-compaction 소비 정책 | `.claude/rules/moai/workflow/context-window-management.md` | CC 5-layer compaction 위에서 /clear 규율(1M=50%, 200K=90%, 95%=hard stop)로 상호작용, progress.md 저장 + resume emit 의무화 | 강(strong) |
| 모델별 컨텍스트 임계값 SSOT | `context-window-management.md:38-44` | 1M=50% / 200K=90% 비대칭 stall-risk 테이블, session-handoff Trigger #1과 공유 | 강 |
| Paste-ready resume 6-block 구조 | `.claude/rules/moai/workflow/session-handoff.md` | ✂ 복사 마커 포함 6-block 표준 resume 메시지, 5개 트리거 의무 emit | 강 |
| Block 0 worktree-anchored resume | `session-handoff.md` | L3 `--worktree` 시 L2 worktree cwd에서 시작하도록 Block 0 prepend | 강 |
| Auto-memory integration | `session-handoff.md` | resume 메시지 memory 파일 저장, MEMORY.md 갱신, `[SUPERSEDED]` 추적 | 강 |
| Resume Diet constraints (AP-D-001..005) | `session-handoff.md` | paste-ready 본문 부풀림 방지(Block 2 ≤4, Block 4 ≤200char, Block 5 single, Block 6 ≤2) | 강 |
| V0 Abort Gate Doctrine | `session-handoff.md` | lsof + cwd 교차검증으로 다중 세션 race false-abort 방지 | 강 |
| 4-rung cheapest-first recovery ladder | `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` | PTL/max_output/media/compact-failure 시 (1) in-turn 자가수정 → (2) paste-ready+/clear → (3) worktree 재시작 → (4) abort+preserve | 강 |
| 5 circuit-breaker invariants | `runtime-recovery-doctrine.md` | 동일 rung 3회 실패 승격, 동일 turn 재시도 금지, abort 시 progress.md로 ledger 닫기 등 | 강 |
| Recovery-Signal Carve-Out | `runtime-recovery-doctrine.md` | recovery turn의 Stop/PostToolUse exit 0 policy(단 현재 훅은 stopReason 미파싱 → mechanical 강제 불가) | 부분(partial) |
| Worktree base branch + .worktreeinclude | `.claude/rules/moai/workflow/worktree-integration.md` | fresh/head base, gitignored 파일 복사, mid-session 진입/이탈 | 부분 |
| /goal autonomous continuation | `.claude/rules/moai/workflow/goal-directive.md` | 매 turn 후 fast model이 완료 조건 판정해 다음 turn 자동 시작 | 부분 |
| Ledger Closure (REQ-LEDGER-001..006) | `.claude/rules/moai/core/agent-common-protocol.md` | 중단된 delegation의 dangling tool_use 정리, abort 시 progress.md로 ledger 닫기 | 부분 |
| Lessons Protocol (auto-memory 영속 학습) | `.claude/rules/moai/core/moai-constitution.md` | 사용자 정정/실패 시 lessons.md 캡처, 도메인 매칭 주입, additive | 강 |
| Context Search Protocol | `CLAUDE.md §16` | 이전 세션 인덱스 grep(30일, 5000 token 한도)으로 컨텍스트 발견·주입 | 부분 |

### 3.2 verifiable-steps (검증 가능한 단계) 정합 메커니즘

| 메커니즘 | 파일 | 기능 | 정합도 |
|----------|------|------|--------|
| SPEC plan→run→sync 3-phase lifecycle | `.claude/rules/moai/workflow/spec-workflow.md` | plan→run→sync 분해 + 각 phase 경계 게이트(plan-auditor / sync-auditor / 구현 착수 승인) | 강 |
| Verification-claim integrity + 5-section report | `.claude/rules/moai/core/verification-claim-integrity.md` | 관측하지 않은 verification/defect 주장 금지(양방향), Claim/Evidence/Baseline/Gaps/Residual 구조화 | 강 |
| Read-only verification batch (7-item 병렬) | `.claude/rules/moai/core/agent-common-protocol.md` | run-phase 완료 검증 7종 단일 turn multi-Bash 병렬(직렬 W3 ~11% wall-time 해소) | 강 |
| Agent Core Behavior #6 Verify, Don't Assume | `.claude/rules/moai/core/moai-constitution.md` | 완료 증거 의무, SPEC 없는 작업은 테스트 가능한 주장으로 목표 정의 | 강 |
| 5 circuit-breaker invariants | `runtime-recovery-doctrine.md` | 복구 절차 자체의 death-spiral 방지(검증 루프 안정성) | 강 |

### 3.3 delegate-vs-oversee (위임 vs 감독) 정합 메커니즘

| 메커니즘 | 파일 | 기능 | 정합도 |
|----------|------|------|--------|
| Dynamic workflow 원시 (script-orchestrated fan-out) | `.claude/rules/moai/workflow/dynamic-workflows.md` | JS script가 수십~수백 subagent 조정, 중간 결과 script 변수 저장, 16 동시/1000 총 제한 | 강 |
| 3 orchestration 원시 비교 | `dynamic-workflows.md` | '누가 plan을 잡는가'로 원시 선택(sub-agents / Agent Teams / workflows) 휴리스틱 | 강 |
| Workflow agent() 목적별 model/effort taxonomy | `dynamic-workflows.md` | read-only-extract(haiku/low) ~ design-architecture(opus/xhigh) 7목적 테이블, silent effort 상속 leak 방지 | 강 |
| L1/L2/L3 worktree 격리 계층 | `.claude/rules/moai/workflow/worktree-integration.md` | L1(CC-native) / L2(MoAI persistent) / L3(사용자 opt-in), 쓰기-heavy는 isolation:worktree 의무 | 강 |
| 3 orchestrator template (Sub/Team/Hybrid) | `.claude/rules/moai/development/orchestrator-templates.md` | 순차/병렬/혼합 패턴 + decision matrix + 전환 신호 | 강 |
| Team Pattern Cookbook (5 구성) | `.claude/rules/moai/workflow/team-pattern-cookbook.md` | Research/Implementation/Review/Design/Debug 팀 + file ownership/통신/shutdown | 강 |
| Pre-Spawn Sync Check (4중 방어) | `.claude/rules/moai/core/agent-common-protocol.md` | git fetch + rev-list + session list로 병렬 세션 race spawn 전 차단 | 강 |
| Team 자동 선택 임계값 + role_profiles | `.moai/config/sections/workflow.yaml` | 복잡도≥7 OR 도메인≥3 OR 파일≥10 시 Teams, 7개 role profile 선언적 정의 | 강 |
| CG Mode (Claude leader + GLM teammates) | `CLAUDE.md §15` | tmux Agent Teams로 구현 무거운 작업 60-70% 비용 절감, planning/보안은 제외 | 부분 |
| Implementation Kickoff Approval | `CLAUDE.local.md §19.1` | run-phase 진입 직전 AskUserQuestion 명시적 승인(Phase 0.5 skip과 별개) | 강 |
| /goal autonomous continuation | `goal-directive.md` | model-evaluated completion(감독의 자동화 측면) | 부분 |

---

## 4. 매핑 분석 (백서 원칙 ↔ MoAI-ADK)

백서 3개 레버를 행으로, MoAI-ADK 메커니즘을 열로 하는 정합도 매핑 표. 정합도는 강(strong) / 부분(partial) / 갭(gap) 3단계.

| 레버 | 백서 원칙 | MoAI-ADK 메커니즘 | 정합도 | 비고 |
|------|-----------|-------------------|--------|------|
| **verifiable-steps** | 작업을 검증 가능한 단계로 분해, 검증은 아키텍처 결정 | SPEC plan→run→sync 3-phase + 각 phase 게이트(plan-auditor / sync-auditor / 구현 착수 승인) | 강 | 백서의 "검증 = 아키텍처 결정"을 phase 경계 게이트로 구현 |
| verifiable-steps | pass/fail 기계적 신호 | Read-only verification batch(7-item 단일 turn 병렬) + Agent Core Behavior #6 | 강 | CC best-practices "provide verification criteria"와 정합 |
| verifiable-steps | 관측 기반 주장만 허용 | Verification-claim integrity(5-section evidence report) | 강 | 백서에 명시적 대응 항목은 없으나 "검증 사슬" 원칙의 MoAI 구현 |
| verifiable-steps | End-state 평가(비결정적 경로) | sync-auditor 4-dimension 결과 평가 + end-state AC | 강 | Anthropic C1 "end-state evaluation rather than turn-by-turn"과 정합 |
| **cross-session continuity** | summaries / handoff / checkpoint 1급 관심사 | Paste-ready resume 6-block + auto-memory + `[SUPERSEDED]` 추적 | 강 | 백서 "1급 운영 관심사"를 5개 트리거 의무 emit으로 구현 |
| cross-session continuity | Initializer + Coding Agent 교대 | Block 0 worktree-anchored resume + session-handoff 전제 검증 | 강 | Anthropic C1 "initializer + coding agent"와 동형 |
| cross-session continuity | 컨텍스트 한계 도달 시 상태 보존 | Graduated-compaction 소비 정책 + progress.md 의무 저장 | 강 | 백서 session compaction + CC 5-layer와 정합 |
| cross-session continuity | 저하 시 routing 신호 | 컨텍스트 임계값 SSOT(1M=50% / 200K=90%) → /clear 또는 worktree 재시작 | 강 | Anthropic C2 "context rot" + 백서 "degradation = routing signal" 정합 |
| cross-session continuity | compaction vs reset 구분 | 4-rung recovery ladder(rung 2=/clear+resume, rung 3=worktree reset) | 강 | Anthropic C2 "compaction(reset과 구분)" + 백서 compaction 정합 |
| cross-session continuity | AGENTS.md (global vs project persistent context) | CLAUDE.md + memory/ + .moai/config/ + agent-memory/ | 부분 | 계층 구조는 존재하나, 백서의 "모델에게 진척을 기록하라" 능동 지시와 auto-memory의 발현 방식 차이 존재(갭 참조) |
| cross-session continuity | 무인 autonomous continuation | /goal(model-evaluated completion) + auto-mode 페어링 | 부분 | 원시는 존재하나 MoAI 내 명시적 auto-mode 페어링 가이드 부재(갭 참조) |
| cross-session continuity | 다중 세션 race 완화 | Pre-Spawn Sync Check(4중 방어) + V0 Abort Gate Doctrine | 강 | 백서에 명시적 항목은 없으나 continuity의 신뢰성 근간 |
| cross-session continuity | abort 시 ledger 정리 | Ledger Closure(REQ-LEDGER-001..006) + abort-closes-ledger invariant | 부분 | team-ac-verify.sh exit-2 ledger_note는 stub(전체 AC 검증은 follow-up SPEC 연기) |
| **delegate-vs-oversee** | 복잡도/위험 기반 라우팅 | Team 자동 선택 임계값(복잡도≥7/도메인≥3/파일≥10) + 3 orchestrator template | 강 | 백서 "라우팅 정책"의 MoAI 구현 |
| delegate-vs-oversee | 위임 vs 감독 의식적 분리 | Workflow agent() 목적별 model/effort taxonomy(7목적) + role_profiles | 강 | 백서 "verification-tier model routing(가벼운 모델로 검증)"과 정합 |
| delegate-vs-oversee | 병렬 위임 = 병렬 provider dispatch | Dynamic workflow fan-out(16 동시/1000 총) + Agent Teams | 강 | 단 CG Mode(tmux env 격리)는 partial(특정 워크플로우 한정) |
| delegate-vs-oversee | cost cap (session-scoped) | CG Mode(60-70% 절감) + workflow effort 명시 의무 | 부분 | 백서 "session-scoped cost caps"의 부분 구현, 정량 cost cap 정책 자체는 부재(갭 참조) |
| delegate-vs-oversee | human-in-the-loop verification gate | Implementation Kickoff Approval(AskUserQuestion 명시 승인) | 강 | Anthropic Ctrl+G plan editor mandate + 백서 "HITL policy" 정합 |
| delegate-vs-oversee | fresh-context 독립 검증 | Team Pattern Cookbook(Review/Debug 팀) + Writer/Reviewer 분리 | 강 | Anthropic C1 "adversarial review via fresh-context subagent"와 정합 |
| delegate-vs-oversee | 작업 산출물 파일시스템 직접 기록 | Subagent가 SPEC 산출물을 `.moai/specs/`에 직접 기록 + 경로 참조 전달 | 강 | Anthropic C1 "artifact systems... lightweight references"와 정합 |

---

## 5. 갭 식별 + 적용 권고안

백서 원칙 대비 MoAI-ADK의 보완점을 최소 변경 지향으로 정리. 과잉 추상화·미래 대비 훅은 지양한다.

### 갭 G1: AGENTS.md 능동 진척 기록 지시의 부재

- **백서 원칙**: "모델에게 배운 것/진척을 기록하라"는 능동적 지시(AGENTS.md). global(project rules/testing/protected dirs) vs project 계층.
- **MoAI-ADK 현상**: `memory/` + `.moai/config/` + `agent-memory/` 계층은 존재하며, Lessons Protocol이 자동 캡처를 수행한다. 그러나 백서의 "모델이 매 세션 진척을 능동적으로 기록하라"는 지시와 비교하면, MoAI는 주로 trigger 기반(사용자 정정, 세션 종단, 실패 시)으로 기록한다.
- **권고안(최소 변경)**: 새로운 메커니즘을 추가하지 말고, 기존 `session-handoff.md`의 5개 트리거와 `progress.md` 의무 저장 규칙이 이미 "진척 기록"을 담당함을 문서에서 명시하라. 추가 제어가 필요하면 SPEC plan/run 산출물의 `progress.md` 갱신을 run-phase 종료 조건에 포함하는 것으로 충분하다. AGENTS.md라는 별도 파일 도입은 불필요(CLAUDE.md + memory/가 동일 역할).

### 갭 G2: Session-level cost cap 가시성 부재

- **백서 원칙**: "token budget = session-level problem (correlation ID)", "session-scoped cost caps".
- **MoAI-ADK 현상**: `context-window-management.md`가 토큰 임계값(1M=50% / 200K=90%)을 관리하나, 이는 context window 기준이지 달러/크레딧 cost cap이 아니다. CG Mode가 비용 절감(60-70%)을 제공하지만 session-scoped 정량 cap 정책은 부재.
- **권고안(최소 변경)**: Claude Code 자체의 `/cost` 표시와 statusline에 의존하라. MoAI가 별도 cost cap 게이트를 구현하는 것은 over-engineering이다. 다만 `session-handoff.md` Trigger #1(컨텍스트 임계)가 이미 cost 압력의 간접 신호로 기능함을 문서에 명시하는 것은 가치가 있다.

### 갭 G3: Verification-tier model routing의 명시적 가이드 부재

- **백서 원칙**: "verification-tier model routing (가벼운 모델로 검증)".
- **MoAI-ADK 현상**: `dynamic-workflows.md`의 workflow `agent()` 목적별 taxonomy에 `verify-judge` 항목(sonnet 또는 opus, xhigh)이 존재한다. 그러나 이는 "검증에 더 비싼 모델"을 지시하며, 백서의 "가벼운 모델로 검증"과는 방향이 다르다. 백서의 verification-tier는 "싼 모델로 빠르게 1차 검증"을 의미한다.
- **권고안(최소 변경)**: `dynamic-workflows.md`의 taxonomy에 "lightweight verification pass" 개념을 보완하라. CC의 prompt-based Stop hook(Haiku 기본 평가)과 `/goal` evaluator(Haiku)가 이미 이 역할을 수행하므로, MoAI는 이를 "verification-tier"로 명명하여 문서화하는 것으로 충분하다. 새 메커니즘 불필요.

### 갭 G4: Recovery-Signal Carve-Out의 mechanical 강제 부재

- **백서 원칙**: 직접적 대응 항목은 아니나, continuity 안정성 근간.
- **MoAI-ADK 현상**: `runtime-recovery-doctrine.md` §4는 recovery turn에서 Stop/PostToolUse 훅이 exit 0해야 함을 policy로 선언했으나, 현재 훅은 `stopReason`을 parse하지 못해 기계적 강제가 불가하다. 미래 runtime-layer SPEC(`SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`)으로 연기됨.
- **권고안(최소 변경)**: 현 상태(partial, documentation-only)를 유지하라. `stopReason` 파싱은 CC 런타임 영역이며, MoAI가 섣불리 기계화 시도하면 over-creep(AP-RR-001)이다. 연기된 SPEC 후보로만 추적.

### 갭 G5: Ledger Closure의 stub 상태

- **백서 원칙**: "errors compound + resume-from-where-failed" (처음부터 재시작 금지).
- **MoAI-ADK 현상**: `team-ac-verify.sh` exit-2의 `ledger_note` 주입은 stub이며, 전체 AC 검증 로직은 follow-up SPEC으로 연기됨.
- **권고안(최소 변경)**: REQ-LEDGER-002의 stub 상태를 명시적으로 추적하되, abort 시 `progress.md`로 ledger를 닫는 REQ-LEDGER-001(이미 강제)이 핵심 불변량임을 유지하라. 전체 AC 검증은 별도 SPEC 후보.

### 갭 G6: CG Mode의 부분적 적용 범위

- **백서 원칙**: "parallel delegation = parallel provider dispatch".
- **MoAI-ADK 현상**: CG Mode는 tmux 세션 env 격리 기반으로 구현 무거운 작업에 한정되며, planning/보안/debugging은 Claude reasoning 필요로 제외된다. 이는 의도적 제한이지 결함이 아니다.
- **권고안(최소 변경)**: 변경 불필요. CG Mode의 제외 범위(planning/보안/debugging)는 백서의 "delegate-vs-oversee(위험 기반 라우팅)"와 정합한다. 문서에 이 제한이 의도적임을 명시.

---

## 6. 후속 SPEC 후보 (사용자 결정용)

본 분석은 SPEC 수립이 아니며, 아래는 갭 해결을 위한 SPEC ID 후보만 나열한다. 사용자가 우선순위를 결정한다.

1. **SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001** — Recovery-Signal Carve-Out의 mechanical 강제(`stopReason` 파싱). runtime-layer. (갭 G4)
2. **SPEC-V3R6-TEAM-AC-VERIFY-FULL-001** — `team-ac-verify.sh`의 stub AC 검증 로직 완성. (갭 G5)
3. **SPEC-V3R6-VERIFICATION-TIER-DOC-001** — verification-tier model routing(가벼운 모델 1차 검증) 문서 보완. 문서 전용, 최소 변경. (갭 G3)
4. **(권장 안 함)** session-scoped 정량 cost cap 게이트 — over-engineering风险, `/cost` + statusline 의존 권장. (갭 G2)
5. **(권장 안 함)** AGENTS.md 별도 파일 도입 — CLAUDE.md + memory/가 동일 역할, 불필요. (갭 G1)

> 후속 SPEC 후보 수: 5 (그중 2개는 "권장 안 함" — 최소 변경 원칙에 따라 신규 메커니즘 도입 지양)

---

## Gaps (미검증)

본 보고서의 한계를 명시한다.

1. **PDF 직접 패치 불가**: 백서 PDF(https://cdn.openai.com/pdf/8a9f00cf-d379-4e20-b06f-dd7ba5196a11/OAI_WhitePaper_Codex-maxxing26.pdf)는 webReader로 직접 패치되지 않았다. 본 보고서의 백서 요약은 `paperSummary` 입력(공식 랜딩 + TheRouter.ai 분석 + Jason Liu 원본에 근거한 2차 요약)에 의존한다. 백서 원문의 정확한 문구와 추가 세부 정책은 미검증이다.

2. **Anthropic "Interactive context engineering" 미발견**: 사용자가 지정할 수 있었던 세 번째 Anthropic 타겟(정확 문구 "Interactive context engineering" 또는 별도 "Context engineering" research/engineering 블로그)은 Anthropic 도메인에서 발견되지 않았다. 해당 정확 문구는 유일하게 제3자 GitHub 리포(deanpeters/Product-Manager-Skills)의 context-engineering-advisor 스킬 설명에만 등장하며, 이는 Anthropic 공식 출처가 아니므로 verification-claim-integrity에 따라 본 보고서에서 제외했다.

3. **Anthropic managed-agents 본문 미패치**: https://www.anthropic.com/engineering/managed-agents URL은 사전 탐색에서 snippet으로 확인되었으나, 본 패치 단계에서 webReader 호출을 수행하지 않아 본문 전체를 패치하지 못했다. 단 C1 패턴 표의 managed-agents 관련 2개 패턴(Brain-Hands-Session 3-way 분리, 세션 = 컨텍스트 창 외부 객체)은 해당 URL에서 인용되었으나, 이는 snippet 기반이며 본문 전체 검증은 미수행이다. 위 두 출처(effective-context-engineering + harness-design)가 백서 비교에 필요한 핵심 패턴을 충분히 포함하므로 영향은 제한적이다.

4. **LangGraph 레거시 도메인 미패치**: 요청된 정확 도메인 `langchain-ai.github.io/langgraph/`는 마이그레이션되어 직접 패치되지 않았다. 현재 canonical인 `docs.langchain.com`(persistence / human-in-the-loop)에서 패치했다.

5. **CrewAI Workflows 문서 미패치**: CrewAI Workflows(start/listen/router, persist execution, resume)는 검색 결과 snippet으로만 확인되었고, 전용 문서 페이지는 패치되지 않았다. crews 개념 페이지만 패치되었다. Checkpointing 섹션은 v1.14.7 버전 페이지에서 확보했으며, 최신 crews 페이지에는 아직 체크포인팅 섹션이 없다.

6. **AutoGen 자동 per-step 체크포인팅 미확인**: AutoGen의 save_state/load_state는 명시적·응용 프로그램 호출 기반이다. LangGraph의 superstep checkpointer에 해당하는 AutoGen 자동 per-step 체크포인팅은 공식 microsoft.github.io/autogen 문서에서 발견되지 않았다(learn.microsoft.com의 마이그레이션 가이드는 후속 Microsoft Agent Framework 설명이므로 제외).

7. **제3자 출처 전부 제외**: Medium 블로그, GitHub issues/discussions, CSDN/HuggingFace mirrors 등 비공식 출처는 official-docs-first 원칙에 따라 패턴에서 전부 제외했다.

8. **MoAI-ADK 인벤토리 인벤토리 파일 경로 미독립 검증**: 인벤토리에 나열된 파일 경로(예: `orchestrator-templates.md`, `team-pattern-cookbook.md`, `worktree-integration.md`)는 입력 JSON에 기재된 것이며, 본 보고서 작성 시점에 각 파일의 존재와 최신 내용을 독립적으로 Read 검증하지 않았다. 해당 파일이 존재하지 않거나 내용이 변경되었을 가능성은 미검증이다.

---

**문서 끝.** 본 보고서는 탐색적 분석이며, SPEC plan/run/sync 워크플로우를 개시하지 않는다. 후속 작업이 필요한 경우 섹션 6의 SPEC 후보 중 사용자가 선택하여 `/moai plan`으로 진입한다.
