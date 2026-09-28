# REJECT bind>0 — surface init

Each partition row below holds keyword lines (`\[HARD\]|MUST|shall `); condition 1 fails, so the row is not relocatable or compressible under the freeze. Its non-binding paragraphs are separate `¶n` rows.

## CLAUDE.md | Selection Decision Tree (gross 1806, bind 1)

- `**Retained agents (13)**: 'manager-spec', 'manager-develop', 'manager-docs', 'manager-git', 'plan-auditor', 'sync-auditor', 'builder-harness', 'super-advisor', 'manager-design', 'e2e-tester', 'manager`

## CLAUDE.md | 8. User Interaction Architecture (gross 1091, bind 2)

- `[ZONE:Frozen] [HARD] Every question directed at the user MUST be asked via AskUserQuestion. Free-form prose questions in response text are prohibited.`
- `[ZONE:Frozen] [HARD] 'AskUserQuestion', 'TaskCreate', 'TaskUpdate', 'TaskList', 'TaskGet' are **deferred tools** — schemas NOT loaded at session start; call 'ToolSearch(query: "select:AskUserQuestion,`

## CLAUDE.md | 14. Parallel Execution Safeguards (gross 1503, bind 1)

- `For core principles, see '.claude/rules/moai/core/moai-constitution.md'. Operational safeguards: file-write-conflict prevention (dependency graphs before parallel execution), agent tool requirements (`

## AGENTS.md | 1. Evidence and verification claims (gross 1312, bind 2)

- `**No unobserved claim.** An actor MUST NOT assert a verification, a completion, **a defect / debt /`
- `**Baseline-integrity attribution.** Every verification claim MUST be attributed to an`

## AGENTS.md | 5. Core behaviors (gross 2607, bind 1)

- `minimum code that works. The ladder is language-neutral. **Never simplify away safety**: it MUST`

## AGENTS.md | 6. Output, language, and format (gross 1589, bind 2)

- `sites, question text — MUST read as natural native prose. Translation-style calques (carry-over of`
- `'conversation_language' text — command strings, file content, question text — MUST be native UTF-8;`

## .claude/rules/moai/core/agent-common-protocol.md | Subagent Prohibitions (gross 1429, bind 1)

- `[ZONE:Frozen] [HARD] Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.`

## .claude/rules/moai/core/agent-common-protocol.md | Hook Invocation Surface (gross 2103, bind 1)

- `Hooks return exit codes and structured JSON; they MUST NOT invoke 'AskUserQuestion' directly. When a hook signals a block (stdout JSON '"decision":"block"' on exit 0, or a legacy exit-2), the orchestr`

## .claude/rules/moai/core/agent-common-protocol.md | Blocker Report Format (gross 514, bind 1)

- `When a subagent requires user input not provided in the spawn prompt, it MUST return a structured blocker report:`

## .claude/rules/moai/core/agent-common-protocol.md | Ledger Closure (gross 1500, bind 1)

- `[ZONE:Evolvable] [HARD] The orchestrator MUST close the ledger on any aborted delegation. Four`

## .claude/rules/moai/core/agent-common-protocol.md | Language Handling (gross 679, bind 1)

- `[ZONE:Evolvable] [HARD] All agents receive and respond in user's configured conversation_language.`

## .claude/rules/moai/core/agent-common-protocol.md | Output Format (gross 350, bind 2)

- `[ZONE:Evolvable] [HARD] User-Facing: Always use Markdown formatting. Never display XML tags to users.`
- `[ZONE:Evolvable] [HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only. Use semantic XML sections for structured data exchange between agents; never surface XML struc`

## .claude/rules/moai/core/agent-common-protocol.md | MCP Fallback Strategy (gross 507, bind 1)

- `[ZONE:Evolvable] [HARD] Maintain effectiveness without MCP servers. Where one is unavailable, use`

## .claude/rules/moai/core/agent-common-protocol.md | Agent Invocation Pattern (gross 272, bind 1)

- `[ZONE:Evolvable] [HARD] Agents are invoked through MoAI's natural language delegation pattern ("Use the {agent-name} subagent to {task description}") — natural language conveys full context including `

## .claude/rules/moai/core/agent-common-protocol.md | Per-Spawn Model Injection (gross 1162, bind 1)

- `[ZONE:Evolvable] [HARD] When spawning a subagent, pass the model the active profile resolves for that agent as an explicit 'model' argument on the spawn. (Why omitting is not neutral, and the full pro`

## .claude/rules/moai/core/agent-common-protocol.md | Background Agent Execution (gross 1695, bind 2)

- `[ZONE:Evolvable] [HARD] Since Claude Code v2.1.198 subagents run in the background by **default**;`
- `[ZONE:Evolvable] [HARD] **While a worktree is being actively audited, it has exactly one writer.**`

## .claude/rules/moai/core/agent-common-protocol.md | Tool Usage Guidelines (gross 129, bind 1)

- `[ZONE:Evolvable] [HARD] Agents must follow tool usage patterns optimized for accuracy and efficiency.`

## .claude/rules/moai/core/agent-common-protocol.md | Parallel Execution (gross 339, bind 1)

- `[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification batch as a single-turn multi-Bash call. Serial verification across turns wastes wall-time and is the single largest s`

## .claude/rules/moai/core/agent-common-protocol.md | Verbatim batch, output contracts, and CLI idioms (gross 2192, bind 1)

- `- **Evidence export.** The cited path must still resolve at audit time, and surviving '/tmp' clearance is not the same thing. '.moai/state/verify/<session>/' is **machine-local scratch**: it outlives `

## .claude/rules/moai/core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) (gross 3250, bind 4)

- `[ZONE:Evolvable] [HARD] Before spawning any implementation 'Agent()' (manager-develop / manager-docs / per-spawn 'Agent(general-purpose)' with a domain whitelist) that will commit or modify shared wor`
- `* **Lane A (ordered):** 'git fetch origin main' MUST finish and its exit status be observed before 'git rev-list --count --left-right origin/main...HEAD' starts. The divergence count is only attributa`
- `The orchestrator MUST retain the fetch completion status (and, where the`
- `| 'N M' | Diverged (both ahead) | STOP, MUST resolve before spawn |`

## .claude/rules/moai/core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) (gross 2929, bind 2)

- `[ZONE:Evolvable] [HARD] Direct main-session edits to shared working-tree paths (Edit/Write/Bash — any direct edit) bypass the spawn gate above, so the orchestrator MUST run the parallel-session detect`
- `[ZONE:Evolvable] [HARD] In the primary checkout, NEVER 'git add -A', 'git add .', or 'git commit -a'. Stage by explicit pathspec ('git add <path> …'), and re-read 'git status --short' immediately befo`

## .claude/rules/moai/core/agent-common-protocol.md | Time Estimation (gross 251, bind 1)

- `[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.`

## .claude/rules/moai/core/askuser-protocol.md | Channel Monopoly (gross 1091, bind 1)

- `**AskUserQuestion is the only user-facing question channel.** The MoAI orchestrator MUST route every user-facing question through an 'AskUserQuestion' tool invocation. Free-form interrogative prose in`

## .claude/rules/moai/core/askuser-protocol.md | Mandatory Preload Step (gross 165, bind 1)

- `Immediately before **every** 'AskUserQuestion' call, the orchestrator MUST invoke:`

## .claude/rules/moai/core/askuser-protocol.md | Structural Constraints (all mandatory) (gross 1389, bind 4)

- `3. **First option label**: MUST carry the '(권장)' (Korean) or '(Recommended)' (English) suffix to signal the recommended choice — this is the 'push'-mode branch; while 'interview.recommendation_mode' i`
- `4. **Language**: All question text, option labels, and option descriptions MUST be in the user's 'conversation_language' (read from '.moai/config/sections/language.yaml')`
- `5. **Round progression**: Each subsequent round MUST narrow ambiguity by building on previous answers — repeating the same question is prohibited`
- `6. **Termination condition**: Rounds continue until intent clarity reaches 100%; the interview MUST NOT end prematurely`

## .claude/rules/moai/core/askuser-protocol.md | Option Description Standards (gross 1025, bind 3)

- `Every option in an 'AskUserQuestion' call MUST have a 'description' field populated with sufficient detail for the user to evaluate implications and trade-offs **without consulting external context**.`
- `Each option description MUST include:`
- `**Bias prevention**: Option descriptions MUST use neutral, factual language — no persuasive or deprecating tone. The recommendation signal is conveyed **exclusively** through the '(권장)' / '(Recommende`

## .claude/rules/moai/core/askuser-protocol.md | Recommendation mode (gross 1434, bind 2)

- `[ZONE:Evolvable] [HARD] While 'recommendation_mode' is 'pull', the orchestrator MUST omit the`
- `and MUST NOT re-encode the same preference through option ordering, description wording, or`

## .claude/rules/moai/core/askuser-protocol.md | On-request emission (gross 437, bind 2)

- `[ZONE:Evolvable] [HARD] When the user explicitly asks for a recommendation, a preference, or an`
- `analysis, the orchestrator MUST emit the withheld recommendation on the requested surface, in the`

## .claude/rules/moai/core/askuser-protocol.md | The three adopted conditions (gross 524, bind 1)

- `2. **An LLM 'best practice' is not a policy.** A model-inferred default MUST NOT be presented as an`

## .claude/rules/moai/core/askuser-protocol.md | Preview Field Standards (gross 518, bind 1)

- `'description' and never replaces it. [HARD] Single-select only — it is silently dropped when`

## .claude/rules/moai/core/askuser-protocol.md | Report-Before-Ask Gate (gross 665, bind 1)

- `[ZONE:Evolvable] [HARD] A decision-type 'AskUserQuestion' whose options derive from investigation results MUST be preceded — in the same turn's response body — by a substantive findings report. Invest`

## .claude/rules/moai/core/askuser-protocol.md | Requested-Deliverable Primacy (user requirement analysis first) (gross 1011, bind 1)

- `[ZONE:Evolvable] [HARD] When the user's latest message explicitly requests a report, analysis, or explanation ("report on X", "explain why", "analyze this first"), that requested deliverable IS the tu`

## .claude/rules/moai/core/askuser-protocol.md | Report Completeness Criteria (all mandatory) (gross 755, bind 1)

- `2. **Option-to-report traceability**: every codename, identifier, or finding referenced in the question's option labels / descriptions / previews MUST have been introduced and explained in the precedi`

## .claude/rules/moai/core/askuser-protocol.md | Preview-as-Report Substitution (named anti-pattern) (gross 413, bind 1)

- `[HARD] Option 'preview' / 'description' fields MUST NOT be the sole carrier of investigation findings. The preview compresses a comparison; the report explains the evidence. Compressing all findings i`

## .claude/rules/moai/core/askuser-protocol.md | Report-Promise Fulfillment (gross 396, bind 1)

- `[HARD] When prior narration in the same task promised a consolidated report ("I will consolidate and report", or its equivalent in any locale), the report MUST be rendered before any subsequent decisi`

## .claude/rules/moai/core/askuser-protocol.md | Orchestrator Obligations (gross 493, bind 1)

- `The MoAI orchestrator (main session) MUST:`

## .claude/rules/moai/core/askuser-protocol.md | Subagent Prohibitions (gross 403, bind 3)

- `- [ZONE:Frozen] [HARD] Subagents MUST NOT invoke 'AskUserQuestion'`
- `- [ZONE:Frozen] [HARD] Subagents MUST NOT output free-form prose questions directed at the user`
- `- [ZONE:Frozen] [HARD] Subagents MUST NOT embed AskUserQuestion call syntax in their response body`

## .claude/rules/moai/core/askuser-protocol.md | Free-form Circumvention Prohibition (gross 1117, bind 1)

- `Free-form interrogative prose in the response body MUST NOT be used as a substitute for 'AskUserQuestion' — always use AskUserQuestion.`

## .claude/rules/moai/core/askuser-protocol.md | Completion-Report Next-Step Discipline (gross 1513, bind 1)

- `[ZONE:Evolvable] [HARD] A completion report (a "done" / "All Done" summary) MUST NOT end with a free-form prose next-step question — "What would you like to do next?", "무엇을 도와드릴까요? (예: A / B / C)", or`

## .claude/rules/moai/core/askuser-protocol.md | Non-ASCII Tool-Call Encoding (gross 860, bind 1)

- `The 'AskUserQuestion' payload — 'question', 'header', and every option 'label' / 'description' / 'preview' — routinely carries text in the user's 'conversation_language'. For Korean, Japanese, Chinese`

## .claude/rules/moai/core/moai-constitution.md | MoAI Constitution (gross 90, bind 1)

- `Core principles that MUST always be followed. These are HARD rules.`

## .claude/rules/moai/core/moai-constitution.md | MoAI Orchestrator (gross 1083, bind 1)

- `- [ZONE:Frozen] [HARD] AskUserQuestion is the sole user-facing question channel, used ONLY by the MoAI orchestrator (subagents must never prompt users); all preload ('ToolSearch(query: "select:AskUser`

## .claude/rules/moai/core/moai-constitution.md | Response Language (gross 660, bind 2)

- `All user-facing responses MUST be in the user's conversation_language.`
- `- [ZONE:Evolvable] [HARD] For non-English 'conversation_language', output MUST be native idiom, not English mapped word-for-word (no translation-style calques — e.g. figurative "축(axis)" / "기둥(pillar)`

## .claude/rules/moai/core/moai-constitution.md | Opus 5.5 Prompt Philosophy (gross 1811, bind 2)

- `- [ZONE:Evolvable] [HARD] **Principle 4 — fewer subagents by default**: 4.7+ does not auto-spawn.`
- `- [ZONE:Evolvable] [HARD] **Principle 5 — fewer tool calls by default**: specify when and why each`

## .claude/rules/moai/core/moai-constitution.md | MX Tag Quality Gates (gross 589, bind 1)

- `- High fan_in functions (>=3 callers): MUST have @MX:ANCHOR`

## .claude/rules/moai/core/moai-constitution.md | Lessons Protocol (gross 2609, bind 2)

- `- [ZONE:Evolvable] [HARD] **A long index never justifies dropping a lesson.** Write the topic file`
- `- [ZONE:Evolvable] [HARD] **Which index holds the line is decided when it is written, not by`

## .claude/rules/moai/core/moai-constitution.md | 1. Surface Assumptions [ZONE:Evolvable] [HARD] (gross 518, bind 1)

- `### 1. Surface Assumptions [ZONE:Evolvable] [HARD]`

## .claude/rules/moai/core/moai-constitution.md | 2. Manage Confusion Actively [ZONE:Evolvable] [HARD] (gross 461, bind 1)

- `### 2. Manage Confusion Actively [ZONE:Evolvable] [HARD]`

## .claude/rules/moai/core/moai-constitution.md | 3. Push Back When Warranted [ZONE:Evolvable] [HARD] (gross 549, bind 1)

- `### 3. Push Back When Warranted [ZONE:Evolvable] [HARD]`

## .claude/rules/moai/core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] (gross 1894, bind 2)

- `### 4. Enforce Simplicity [ZONE:Evolvable] [HARD]`
- `Never simplify away (safety carve-out): the ladder is a code-economy aid, NOT a license to cut safety. It MUST NOT be used to drop input validation at trust boundaries, error handling that prevents da`

## .claude/rules/moai/core/moai-constitution.md | 5. Maintain Scope Discipline [ZONE:Evolvable] [HARD] (gross 672, bind 1)

- `### 5. Maintain Scope Discipline [ZONE:Evolvable] [HARD]`

## .claude/rules/moai/core/moai-constitution.md | 6. Verify, Don't Assume [ZONE:Evolvable] [HARD] (gross 613, bind 1)

- `### 6. Verify, Don't Assume [ZONE:Evolvable] [HARD]`

## .claude/rules/moai/core/moai-mcp-tools.md | The `project_root` input — name your own tree (gross 2402, bind 1)

- `[HARD] **An agent working inside a worktree MUST pass it**, and the value is its`

## .claude/rules/moai/core/native-idiom-and-register.md | The Invariant (gross 856, bind 1)

- `[ZONE:Evolvable] [HARD] When 'conversation_language ≠ en', every user-facing surface — chat replies, reports, README, docs-site, generated sites, 'AskUserQuestion' text — MUST read as natural native p`

## .claude/rules/moai/core/native-idiom-and-register.md | Mechanism — when to invoke humanize (gross 464, bind 1)

- `[ZONE:Evolvable] [HARD] Heavy non-English artifacts (multi-paragraph reports, README rewrites, docs-site pages, generated sites) MUST pass through the 'moai-domain-humanize' skill as a final phase bef`

## .claude/rules/moai/core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) (gross 2110, bind 2)

- `[ZONE:Evolvable] [HARD] An actor MUST NOT assert a verification, a completion, **a defect / debt / drift, OR the premise underlying a recommendation** it did not actually verify with the domain's mech`
- `The binding extends to the premise beneath a recommendation. A recommendation to KEEP, retain, or preserve something rests on a premise — that the thing is still live, still reachable, still depended `

## .claude/rules/moai/core/verification-claim-integrity.md | 1.1 Binding scope — ALL FOUR surfaces (gross 2548, bind 3)

- `1. **Orchestrator self-report** — the orchestrator's own Completion Report and Verification Matrix banners, and its trust-but-verify batches, as defined in '.claude/output-styles/moai/moai.md' (Respon`
- `2. **Manager-agent completion report** — the self-verification deliverables of 'manager-develop' and 'manager-docs'. When a manager agent reports an acceptance-criteria PASS/FAIL matrix, a build resul`
- `3. **Defect / debt / drift identification claim** — any actor's assertion that a defect, technical-debt item, drift, or anomalous state EXISTS and warrants action. A claim that "module X is broken", "`

## .claude/rules/moai/core/verification-claim-integrity.md | 2. Baseline-Integrity Attribution / baseline 무결성 귀속 (gross 1028, bind 3)

- `[ZONE:Evolvable] [HARD] Every verification claim MUST be attributed to an actually-measured baseline — the command that was run plus the output that was observed.`
- `A claim MUST NOT be assumed, and MUST NOT be carried over from a prior unrelated measurement. "Coverage is at threshold" attributed to a baseline means: the actor ran the coverage command and observed`
- `Anything else (an inferred value, a stale figure, a "should be" estimate) is unattributed and MUST be reported as a Gap (§3.4), not as a Claim.`

## .claude/rules/moai/core/verification-claim-integrity.md | 2.1 Moving-ref attribution — the anchor-or-subject predicate (gross 2942, bind 4)

- `[ZONE:Evolvable] [HARD] A claim decided against a **moving ref** — 'origin/main', 'origin/develop', 'origin/HEAD', or any other name that resolves to a different commit as work lands — carries no base`
- `[HARD] The predicate is applied, not recalled. Before remediating any moving-ref or moving-coordinate claim, read 'verification-claim-integrity-detail.md' § Moving-ref predicate and run its four tests`
- `[HARD] The tests return a **class**; the class does not name the remedy. There are **two classes and four remedies** — ANCHOR selects between R1 and R2, SUBJECT between R3 and R4. Collapsing the two s`
- `The branch table, its cost table, and the exemption-marker syntax live in 'verification-claim-integrity-detail.md' § The four remediation branches and § The exemption marker — the same section the [HA`

## .claude/rules/moai/core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree (gross 2099, bind 2)

- `[ZONE:Evolvable] [HARD] A measurement produced by the project's own tooling is attributed to **two** coordinates, not one: the tree it read, and the build that judged it. §2 binds the first. This clau`
- `**The obligation.** A tool measurement cited as evidence MUST have been produced by a build made from the tree under measurement. Concretely, either:`

## .claude/rules/moai/core/verification-claim-integrity.md | 2.3 Ordering attribution — the commit graph is the only sequencing witness (gross 1245, bind 1)

- `[ZONE:Evolvable] [HARD] When a claim's validity depends on a measurement having been taken BEFORE the change it measures (a baseline-first acceptance criterion), the baseline artifact MUST land in its`

## .claude/rules/moai/core/verification-claim-integrity.md | 3. The 5-Section Evidence-Bearing Report Format (gross 1506, bind 1)

- `[ZONE:Evolvable] [HARD] Verification and completion reports — on either binding surface (§1.1) — SHOULD be structured as the following five sections. The format is the operational mechanism that enfor`

## .claude/rules/moai/core/verification-claim-integrity.md | 3.1 Refused-tool degradation — a refusal is a Gap, never a silent substitution (gross 1761, bind 1)

- `[ZONE:Evolvable] [HARD] Where a command a verification rested on was **refused rather than executed** — the worktree-isolation guard, a permission deny, a policy gate — the report names that refusal i`

## .claude/rules/moai/workflow/cache-aware-execution.md | Directives (gross 4040, bind 5)

- `6. **Pass files by '@'-mention, not by name** [ZONE:Evolvable] [HARD] When a prompt needs a file's content, pass it with an '@'-mention or a Read call rather than citing the filename for the model to `
- `7. **Keep command output bounded** [ZONE:Evolvable] [HARD] Every command must bound what it returns: quiet flags, targeted queries, or redirect-to-file with the exit code and a bounded tail. 'BASH_MAX`
- `8. **Prefer the quiet form of routine commands** [ZONE:Evolvable] [HARD] Call everyday commands in their quiet form — '--no-progress', '-q', machine-readable output with a targeted filter — not forms `
- `9. **Weigh session length as a cost axis** [ZONE:Evolvable] [HARD] One long session is cheaper than several short ones for the same work — every fresh session re-pays the always-loaded prefix at write`
- `10. **A mid-session model or effort switch busts the cache** [ZONE:Evolvable] [HARD] Changing model or effort mid-session (thinking budget included — 'MAX_THINKING_TOKENS') discards the prompt cache; `

## .claude/rules/moai/workflow/context-window-management.md | Context Window Targets (gross 1612, bind 1)

- `[ZONE:Evolvable] [HARD] Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling late`

## .claude/rules/moai/workflow/context-window-management.md | User Responsibilities (gross 746, bind 3)

- `[ZONE:Evolvable] [HARD] When usage crosses the model-specific threshold:`
- `[ZONE:Evolvable] [HARD] When usage crosses 95% on any model:`
- `- The next action MUST be '/clear' — no further large work in the current session`

## .claude/rules/moai/workflow/context-window-management.md | Orchestrator Responsibilities (gross 1770, bind 3)

- `The orchestrator MUST proactively recognize the model-specific boundary and prepare the user for a clean handoff.`
- `[ZONE:Evolvable] [HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (50% on 1M / GLM-5.3, 90% on 200K), it MUST:`
- `[ZONE:Evolvable] [HARD] Resume message format: include all of the following so the next session is self-sufficient (locale renderings per 'session-handoff.md' § Localization Table — do not redefine a `

## .claude/rules/moai/workflow/cross-session-messaging.md | Rules (gross 4424, bind 4)

- `[ZONE:Evolvable] [HARD] **Never route a user decision through a peer.** The user-facing question channel is unchanged: questions go to the user through the orchestrator's question tool. A peer session`
- `[ZONE:Evolvable] [HARD] **Never ask a peer to do what this session may not do.** Work blocked or denied here does not become permissible by delegation. When a needed action is outside this session's p`
- `[ZONE:Evolvable] [HARD] **Send facts, not instructions to mutate shared state.** A message may report what landed, what broke, what a decision was, or ask a question. It must not direct a peer to edit`
- `[ZONE:Evolvable] [HARD] **Never address a stopped teammate by name.** A teammate stopped with`

## .claude/rules/moai/workflow/cross-session-messaging.md | A send result has three shapes, and none of them says "read" (gross 2055, bind 1)

- `[ZONE:Evolvable] [HARD] **A successful send means the message reached the session, not that its Claude read it.** The result answers where the text went; it never answers whether a model consumed it. `

## .claude/rules/moai/workflow/cross-session-messaging.md | An idle notice is a scheduling hint (gross 1087, bind 1)

- `[ZONE:Evolvable] [HARD] **An idle notice is not completion evidence.** A session goes idle when it finishes, when it stops at a permission prompt, and when it dies, and the notice cannot tell those th`

## .claude/rules/moai/workflow/kanban-dispatch.md | Kanban Dispatch Protocol (gross 1308, bind 1)

- `> **Detail companion**: 'kanban-dispatch-detail.md' owns the long tables, dispatch-cycle walkthrough, incident narratives, and rationale — now also per-card fan-out, Factory in-lane 3-stage, and the '`

## .claude/rules/moai/workflow/kanban-dispatch.md | Entry into the board is an operator act (gross 4526, bind 7)

- `[HARD] **The lead is the queue's sole producer.** The operator asks; the lead turns the request into a card with 'moai gtd add "<description>"' ('moai gtd' alone lists the queue). Production is the on`
- `[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** '/moai project' issues one card when it completes, prefixed '[PROJECT] '; the codemaps-debt`
- `[HARD] **Promotion is the operator's act, always.** After a '/clear', the lead presents the queued cards through 'AskUserQuestion' and the operator picks; only then does the lead dispatch according to`
- `[HARD] **The lead may attach a finding; it may not act on one.** Analysis runs automatically and records a relation between two cards — a near-duplicate the machine measured on 'add' or 'analyze', or `
- `[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of 'backlog', the lead reads that card's pull-request and landed state and reports what it read in the same turn. 'moai gtd pr`
- `[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id answers "did THIS card land"; it cannot answer "has someone else already done this", because the delive`
- `[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the lead surfaces that and the operator **confirms or withdraws** it. The lead ne`

## .claude/rules/moai/workflow/kanban-dispatch.md | Report milestones ↔ queue cards (gross 579, bind 1)

- `[HARD] **A milestone-bearing report under '.moai/reports/' carries a '## Card Cross-Check' section** — one table row per milestone, a 'card' column holding the delivering card id or an explicit new-ca`

## .claude/rules/moai/workflow/kanban-dispatch.md | Card classes — not every card needs every column (gross 1990, bind 1)

- `[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checkable, and are checked and cited: the diff is measured ('git diff --stat' agai`

## .claude/rules/moai/workflow/kanban-dispatch.md | The delegation channel is the queue (gross 1694, bind 2)

- `[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitt`
- `Two properties of that nudge channel bear on dispatch. **A lane can be asked to report when it next goes idle** ('SendMessage' 'notify_when_idle', opt-in and one-shot), which spares the lead a polling`

## .claude/rules/moai/workflow/kanban-dispatch.md | Dispatch language (gross 607, bind 1)

- `[HARD] A dispatch is written in the operator's 'conversation_language' — the operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The boundary is **wh`

## .claude/rules/moai/workflow/kanban-dispatch.md | Dispatch format (gross 2051, bind 2)

- `[HARD] A dispatch is a fixed-field address block, not prose. The fields:`
- `- **[HARD] The send is read, not assumed.** The result has three shapes and only one of them delivered:`

## .claude/rules/moai/workflow/kanban-dispatch.md | Deputy dispatch surface (gross 1869, bind 5)

- `[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the '-k'/'-f' lead session spawns exactly one UNNAMED background 'Agent()' running manager-lead as its **coordi`
- `[HARD] **A completion report reaches the lead as a 'RECOMMEND:' summary, not as raw reading batches** — the deputy does the raw-tree read and names the paths it read. Turn occupancy moves; the lead's `
- `[HARD] **Round-report measurement and drafting are the deputy's; the asserted figures are the lead's** — every figure names its measurer, and an unattributed one is a defect. The report is per-round f`
- `[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval ('LEAD-MERGE-APPROVED'), operator gates, card issuance and 'done' ('moai gtd' mutations), CodeR`
- `[HARD] **Nothing structural moves with the delegation.** The queue stays the channel, completion stays evidence the lead read, and the verdict's home stays the lead. The deputy reads and reports; the `

## .claude/rules/moai/workflow/kanban-dispatch.md | Completion is read, never trusted (gross 1205, bind 1)

- `[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Reply routing is not guaranteed to arrive, and a reply is a claim rather than an observation.`

## .claude/rules/moai/workflow/kanban-dispatch.md | CodeRabbit is not read from `gh pr checks` (gross 989, bind 1)

- `[HARD] A 'gh pr checks' row naming CodeRabbit is not evidence that a review happened: the status is 'success' **even when no review ran**, and the row prints 'pass' byte-identically in both cases — on`

## .claude/rules/moai/workflow/kanban-dispatch.md | The `/clear` handoff between phases (gross 874, bind 1)

- `[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to '/clear' that session** — '/`

## .claude/rules/moai/workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher (gross 6381, bind 9)

- `[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare 'git worktree add'.`
- `[HARD] **'moai worktree done' closes L2 trees only.** A worktree entered by short name ('moai cc -w <name>' → '.claude/worktrees/<name>/') is L1 and is never in 'moai worktree''s registry — 'done' on `
- `[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the lead has integrated the branch and the remote merge has landed; dis`
- `[HARD] **A new card starts in a new worktree — exit any previous one first.** 'EnterWorktree(<card-id>)' cannot run from inside a worktree session: a lane still anchored in the previous card's tree MU`
- `[HARD] **Card worktree branches carry the 'WT-' prefix and a descriptive slug — never the card id.** 'EnterWorktree(<name>)' auto-names its branch 'worktree-<name>', which is unwieldy and invisible to`
- `| Card id | MUST NOT appear — not as a prefix, a suffix, or a token |`
- `[HARD] **Dropping the id from the branch moves traceability onto three other carriers, and all three are mandatory.** The branch name no longer answers "which card was this?", so nothing may rely on r`
- `[HARD] **A card-delivering pull request's PR title MUST carry the delivering card id** — and this does not contradict the branch-name rule above: the branch name is read by a human scanning 'git branc`
- `[HARD] **Inside a worktree session that '<path>' is the worktree's own absolute path**, and the dispatch writes it that way. Measured on Claude Code 2.1.275: the guard refuses '-C .', a relative path,`

## .claude/rules/moai/workflow/kanban-dispatch.md | Verification load is lane-local (gross 1100, bind 2)

- `[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite — the better evidence: the full suite, in a clean `
- `[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a 'timeout' w`

## .claude/rules/moai/workflow/kanban-dispatch.md | The env-isolated verification form (gross 1912, bind 1)

- `[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound 'unset … && <command>' invocation:`

## .claude/rules/moai/workflow/kanban-dispatch.md | Integration into the release branch is self-served (gross 3939, bind 4)

- `[HARD] A lane whose card has passed verification does not wait for the lead to integrate it: the lane merges its own branch into the batch's release branch ('release/vX.Y.Z') itself. The lead provisio`
- `[HARD] **Serialize by the recorded hold and the announcement, not by probe.** A lane takes the window BEFORE entering the release worktree: 'moai integration acquire --name <lane> --card <card-id>' re`
- `[HARD] **'acquire' asserts the caller's tree first.** It checks tracked '.claude/settings.json', preserves a modified copy outside that tree, and reports — every time; only the refusal is opt-in, and `
- `[HARD] **Re-read 'HEAD' immediately before the commit and again before the push.** 'AGENTS.md' §2 binds this everywhere; the release worktree is where it has already earned its keep. A repair commit l`

## .claude/rules/moai/workflow/main-checkout-branch-guard.md | Rules (gross 1296, bind 1)

- `[ZONE:Evolvable] [HARD] The orchestrator MUST NOT change branch state in the primary project checkout. Specifically forbidden there:`

## .claude/rules/moai/workflow/main-checkout-branch-guard.md | Staleness Rule (gross 528, bind 1)

- `[ZONE:Evolvable] [HARD] Re-read branch and commit state **immediately before** any commit or push — never rely on a value read earlier in the turn, and never on the branch reported in session-start co`

## .claude/rules/moai/workflow/main-checkout-branch-guard.md | Detecting Concurrent Sessions (gross 521, bind 1)

- `Process-registry lookups are not a reliable emptiness signal — a registry can hold entries whose recorded PIDs no longer match live processes, including the querying session's own. An empty or all-sta`

## .claude/rules/moai/workflow/session-handoff.md | When To Generate (5 Triggers) (gross 1208, bind 1)

- `[ZONE:Evolvable] [HARD] The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate:`

## .claude/rules/moai/workflow/session-handoff.md | Emission-Time Save Obligation (auto-resume wiring) (gross 1858, bind 2)

- `[ZONE:Evolvable] [HARD] When the orchestrator emits a paste-ready resume message (any of the 5 triggers above), it MUST also persist the cut-line-bounded main block verbatim as the pending handoff rec`
- `[ZONE:Evolvable] [HARD] **Fail-open invariant**: when the 'moai' CLI is absent from PATH or 'moai handoff save' exits non-zero, the orchestrator emits the paste-ready surface UNCHANGED — a save failur`

## .claude/rules/moai/workflow/session-handoff.md | Canonical Format (Verbatim Spec) (gross 1359, bind 1)

- `[ZONE:Evolvable] [HARD] Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers** (literal format: § Cut-line Marker Specification below). Cut-line markers sit **inside*`

## .claude/rules/moai/workflow/session-handoff.md | Field-by-Field Specification (gross 2859, bind 2)

- `- **Block 1** — 'ultrathink.' opener (sets 'effort: xhigh'; Adaptive Thinking is a separate axis it does not toggle). '<phase>' ∈ 'plan | run | sync | mx'. [HARD] Fixed line order: opener (plus any ap`
- `- **Block 5** — separator + 'Run:' carrying a **single primary action**, which is always the work-starting command. [HARD] '/moai goal' is arm-only and starts no work, so it never occupies this line a`

## .claude/rules/moai/workflow/session-handoff.md | Auto-Memory Integration (Mandatory) (gross 1384, bind 1)

- `[ZONE:Evolvable] [HARD] When generating a resume message, the orchestrator MUST also:`

## .claude/rules/moai/workflow/session-handoff.md | Output Surface (User-Facing) (gross 1631, bind 1)

- `[ZONE:Evolvable] [HARD] Emitting a resume message means **rendering it in the response body of the turn that generates it** — not storing it. At session end the orchestrator displays all three of: (1)`

## .claude/rules/moai/workflow/session-handoff.md | Worktree-Anchored Resume Pattern (gross 721, bind 1)

- `> [ZONE:Evolvable] [HARD] When the work happened inside a worktree, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure, and Block 4 gains precondition '0`

## .claude/rules/moai/workflow/session-handoff.md | Diet Constraints (gross 701, bind 1)

- `[ZONE:Evolvable] [HARD] A paste-ready resume message is "next session minimum executable context" — NOT an audit trail, history record, or ceremonial commitment record. Two concrete anti-patterns illu`

## .claude/rules/moai/workflow/session-handoff.md | V0 Abort Gate Doctrine (gross 339, bind 1)

- `> [ZONE:Evolvable] [HARD] The paste-ready Block 4 V0 precondition uses **lsof + cwd cross-validation** (NOT a raw 'ps aux' count). When V0-b ≥ 1 OR V0-c ≥ 3, spawning implementation agents is prohibit`

## .claude/rules/moai/workflow/skill-routing.md | 1. Orchestrator Obligation (gross 892, bind 1)

- `[ZONE:Evolvable] [HARD] Before spawning an implementation or review 'Agent()', the orchestrator MUST match the mission's domain against the available 'moai-ref-*' / 'moai-domain-*' skill descriptions `

## .claude/rules/moai/workflow/skill-routing.md | §1.1 — Orchestrator-Direct Skill Routing (non-spawn) (gross 3001, bind 4)

- `[ZONE:Evolvable] [HARD] The §1 obligation binds agent **spawns**. When the orchestrator performs a task **directly** (no 'Agent()' spawn) whose output shape matches a 'moai-domain-*' skill, the orches`
- `Config coupling (report): the orchestrator reads 'report.format' from the settings chain ('.moai/config/sections/report.yaml', persisted via 'internal/settings' — values 'html+md' \| 'md') before rend`
- `[ZONE:Evolvable] [HARD] **Routing is intent-based, not keyword-based**: the orchestrator LLM reads skill descriptions semantically and matches the intent of a request against the described capability,`
- `[ZONE:Evolvable] [HARD] **Anti-pattern (named)**: reaching for 'artifact-design' when the task is a markdown→HTML **report render**. 'artifact-design' calibrates visual identity for claude.ai-hosted w`

