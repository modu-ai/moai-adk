# REJECT rows from the judgment layer (both surfaces)

Source of truth: `.moai/reports/t1226/judgments.tsv` (columns gov / plan / note). gov=a: a keyword line depends on the row (table, heading, enumeration, definition, loading/applicability scope); gov=b: the row narrows a keyword line's scope (exception, exemption, carve-out, boundary). citation-breaks rows: see `evidence/citations-claude.txt`. Rows flagged in design.md §4.3 keep their §4.3 classification.

| surf | file | section | gov | reason | why |
|---|---|---|---|---|---|
| * | CLAUDE.md | MoAI Execution Directive | a | governs-scope | H1 title; the numbered-section skeleton that every "CLAUDE.md §N" citation resolves against |
| * | CLAUDE.md | 0. Standing Contract (imported) | a | governs-scope | carries the @AGENTS.md import (design.md §4.3 rejected) |
| * | CLAUDE.md | 1. Core Identity | a | governs-scope | obligation summary framing delegation (design.md §4.3 rejected) |
| * | CLAUDE.md | HARD Rules (Mandatory) | a | governs-scope | obligation summary (design.md §4.3 rejected) |
| * | CLAUDE.md | 3. Command Reference | N | citation-breaks | heading-only row; removal renumbers the §N skeleton cited as "CLAUDE.md §N" (see evidence/citations-claude.txt) |
| * | CLAUDE.md | 4. Agent Catalog | a | governs-scope | parent heading of Selection Decision Tree, which carries a keyword line |
| * | CLAUDE.md | 11. Error Handling | N | citation-breaks | empty numbered heading kept as a citation anchor (session-handoff cross-refs cite "CLAUDE.md §11") |
| * | CLAUDE.md | 12. MCP Servers & Deep Analysis Modes | N | citation-breaks | empty numbered heading kept as a citation anchor for the §N skeleton |
| * | AGENTS.md | AGENTS.md — standing contract for agents in this repository | a | governs-scope | establishes that every clause binds regardless of harness and that the file is self-sufficient — the applicability of every keyword line (cf. kanban Scope, design.md §4.3) |
| * | core/agent-common-protocol.md | (서문) | a | governs-scope | frontmatter declares the loading scope ("no paths restriction") the file's keyword lines rely on |
| * | core/agent-common-protocol.md | Agent Common Protocol | a | governs-scope | applicability sentence + the companion pointer line REQ-ALD2-005a requires |
| * | core/agent-common-protocol.md | User Interaction Boundary | a | governs-scope | design.md §4.3 (a) |
| * | core/agent-common-protocol.md | Subagent Prohibitions ¶4 | b | narrows-scope | carves lane sessions out of the subagent class the MUST NOT binds |
| * | core/agent-common-protocol.md | Hook Invocation Surface ¶1 | a | governs-scope | names the hooks the section's MUST NOT binds |
| * | core/agent-common-protocol.md | Hook Invocation Surface ¶2 | a | governs-scope | heading of the MUST: procedure |
| * | core/agent-common-protocol.md | Hook Invocation Surface ¶4 | a | governs-scope | the list the "MUST:" line introduces |
| * | core/agent-common-protocol.md | Hook Invocation Surface ¶8 | b | narrows-scope | recovery-signal carve-out narrows when a hook block applies |
| * | core/agent-common-protocol.md | Blocker Report Format ¶2 | a | governs-scope | the format the "MUST return a structured blocker report:" line names |
| * | core/agent-common-protocol.md | Ledger Closure ¶1 | a | governs-scope | defines the invariant the MUST closes |
| * | core/agent-common-protocol.md | Ledger Closure ¶3 | a | governs-scope | the four clauses the MUST line enumerates |
| * | core/agent-common-protocol.md | Ledger Closure ¶4 | a | governs-scope | scope-boundary note fixing where the clause applies |
| * | core/agent-common-protocol.md | Language Handling ¶2 | b | narrows-scope | exceptions (code/identifiers stay English) to the [HARD] language line |
| * | core/agent-common-protocol.md | MCP Fallback Strategy ¶2 | b | narrows-scope | GLM routing replaces the fallback under a GLM backend |
| * | core/agent-common-protocol.md | Per-Spawn Model Injection ¶2 | a | governs-scope | how to resolve the value the [HARD] line requires |
| * | core/agent-common-protocol.md | Background Agent Execution ¶2 | a | governs-scope | "one writer per tree" — the concept the audit-window [HARD] line depends on |
| * | core/agent-common-protocol.md | Verbatim batch, output contracts, and CLI idioms ¶2 | a | governs-scope | introduces the restated binding list |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶2 | a | governs-scope | part of the [HARD] batch definition |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶4 | a | governs-scope | the batch the [HARD] line requires |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶6 | a | governs-scope | matrix label |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶8 | a | governs-scope | matrix label |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶9 | a | governs-scope | interpretation matrix |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶10 | a | governs-scope | matrix semantics ("no false positives") |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶11 | b | narrows-scope | read-only exemption |
| * | core/agent-common-protocol.md | Pre-Spawn Sync Check (Multi-Session Race Mitigation) ¶12 | b | narrows-scope | spawn-gate boundary |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶2 | a | governs-scope | part of the [HARD] rule body |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶3 | a | governs-scope | TRIGGER |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶4 | a | governs-scope | TRIGGER table |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶5 | a | governs-scope | CHECK |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶6 | a | governs-scope | DECIDE |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶7 | a | governs-scope | DECIDE table |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶8 | a | governs-scope | stale-registry caveat on the probe |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶9 | a | governs-scope | RE-CHECK |
| * | core/agent-common-protocol.md | Pre-Edit Sync Check (Direct-Edit Race Mitigation) ¶10 | a | governs-scope | heading of the sweep [HARD] |
| * | core/askuser-protocol.md | (서문) | a | governs-scope | frontmatter (loader contract) |
| * | core/askuser-protocol.md | AskUserQuestion Protocol — Canonical Reference | a | governs-scope | SSOT + loading-scope declaration + companion pointer |
| * | core/askuser-protocol.md | Channel Monopoly ¶2 | a | governs-scope | enumerates the turns the MUST covers |
| * | core/askuser-protocol.md | Channel Monopoly ¶3 | b | narrows-scope | exceptions |
| * | core/askuser-protocol.md | ToolSearch Preload Procedure | a | governs-scope | defines "deferred tool", used by the MUST in the child section |
| * | core/askuser-protocol.md | Mandatory Preload Step ¶2 | a | governs-scope | the call the MUST names |
| * | core/askuser-protocol.md | Socratic Interview Structure | a | governs-scope | design.md §4.3 (a) |
| * | core/askuser-protocol.md | Structural Constraints (all mandatory) ¶2 | a | governs-scope | defines the term the round MUST lines use |
| * | core/askuser-protocol.md | Recommendation Placement Principles | a | governs-scope | design.md §4.3 (a): Recommendation mode [HARD] names principle 5 |
| * | core/askuser-protocol.md | Recommendation mode ¶1 | a | governs-scope | push/pull definition the [HARD] line uses |
| * | core/askuser-protocol.md | Recommendation mode ¶2 | a | governs-scope | the config key the [HARD] line names |
| * | core/askuser-protocol.md | Recommendation mode ¶4 | a | governs-scope | when the mode resolves |
| * | core/askuser-protocol.md | Recommendation mode ¶5 | b | narrows-scope | pull withholds a recommendation and nothing else |
| * | core/askuser-protocol.md | The three adopted conditions ¶1 | a | governs-scope | intro of the binding list |
| * | core/askuser-protocol.md | Requested-Deliverable Primacy (user requirement analysis first) ¶2 | a | governs-scope | operationalizes the [HARD] line |
| * | core/askuser-protocol.md | Exceptions (gate does not apply) | b | narrows-scope | design.md §4.3 (b) |
| * | core/askuser-protocol.md | Orchestrator–Subagent Boundary | a | governs-scope | design.md §4.3 (a) |
| * | core/askuser-protocol.md | Blocker Report Format / Re-delegation Procedure | a | governs-scope | design.md §4.3 (a) ownership pointer |
| * | core/askuser-protocol.md | Ambiguity Triggers and Exceptions | a | governs-scope | design.md §4.3 (a) SSOT declaration |
| * | core/askuser-protocol.md | Free-form Circumvention Prohibition ¶3 | a | governs-scope | enumerates what the MUST NOT covers |
| * | core/askuser-protocol.md | Completion-Report Next-Step Discipline ¶2 | a | governs-scope | "exactly TWO valid closes" |
| * | core/askuser-protocol.md | Completion-Report Next-Step Discipline ¶3 | a | governs-scope | the two closes |
| * | core/moai-constitution.md | Opus 5.5 Prompt Philosophy ¶1 | a | governs-scope | "The binding points:" intro |
| * | core/moai-constitution.md | MX Tag Quality Gates ¶1 | a | governs-scope | frames the list that holds the MUST line |
| * | core/moai-constitution.md | Agent Core Behaviors | a | governs-scope | design.md §4.3 (a) scope of the six behaviors |
| * | core/moai-constitution.md | 1. Surface Assumptions [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 1. Surface Assumptions [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | required format |
| * | core/moai-constitution.md | 2. Manage Confusion Actively [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 2. Manage Confusion Actively [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | required steps |
| * | core/moai-constitution.md | 3. Push Back When Warranted [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 3. Push Back When Warranted [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | when/how of the binding behavior |
| * | core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | ladder intro |
| * | core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] ¶3 | a | governs-scope | the ladder the safety carve-out MUST NOT line refers to |
| * | core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] ¶4 | a | governs-scope | interpretation of the ladder |
| * | core/moai-constitution.md | 4. Enforce Simplicity [ZONE:Evolvable] [HARD] ¶6 | a | governs-scope | quantitative trigger of the binding behavior |
| * | core/moai-constitution.md | 5. Maintain Scope Discipline [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 5. Maintain Scope Discipline [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | the Do NOT list |
| * | core/moai-constitution.md | 6. Verify, Don't Assume [ZONE:Evolvable] [HARD] ¶1 | a | governs-scope | the behavior the [HARD] heading binds |
| * | core/moai-constitution.md | 6. Verify, Don't Assume [ZONE:Evolvable] [HARD] ¶2 | a | governs-scope | evidence requirements |
| * | core/moai-mcp-tools.md | moai-mcp Tool Catalogue | a | governs-scope | SSOT statement fixing what the file owns |
| * | core/moai-mcp-tools.md | MCP-over-CLI rule | a | governs-scope | design.md §4.3 (a): the file's only selection rule |
| * | core/moai-mcp-tools.md | The `project_root` input — name your own tree ¶1 | a | governs-scope | the tool list the MUST line applies to |
| * | core/moai-mcp-tools.md | The `project_root` input — name your own tree ¶4 | a | governs-scope | what to pass |
| * | core/native-idiom-and-register.md | Native-Idiom & Register Policy (Non-English Locales) | a | governs-scope | title + loading scope |
| * | core/native-idiom-and-register.md | The Invariant ¶2 | b | narrows-scope | conditional: zero overhead on English sessions |
| * | core/native-idiom-and-register.md | The Invariant ¶3 | a | governs-scope | defines "native prose" per surface |
| * | core/verification-claim-integrity.md | Verification-Claim Integrity | a | governs-scope | applicability ("loaded for the orchestrator and all agents") |
| * | core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶2 | a | governs-scope | headline of the invariant |
| * | core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶3 | a | governs-scope | elaborates the invariant |
| * | core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶4 | a | governs-scope | defect direction of the invariant |
| * | core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶7 | a | governs-scope | the norm binds independently of the mechanical layer |
| * | core/verification-claim-integrity.md | 1.1 Binding scope — ALL FOUR surfaces ¶1 | a | governs-scope | intro of the four surfaces |
| * | core/verification-claim-integrity.md | 1.1 Binding scope — ALL FOUR surfaces ¶5 | a | governs-scope | surface 4 of "ALL FOUR" |
| * | core/verification-claim-integrity.md | 2. Baseline-Integrity Attribution / baseline 무결성 귀속 ¶3 | a | governs-scope | defines an attributed claim |
| * | core/verification-claim-integrity.md | 2. Baseline-Integrity Attribution / baseline 무결성 귀속 ¶4 | a | governs-scope | defines an attributed claim |
| * | core/verification-claim-integrity.md | 2.1 Moving-ref attribution — the anchor-or-subject predicate ¶2 | a | governs-scope | frames the predicate the [HARD] lines apply |
| * | core/verification-claim-integrity.md | 2.1 Moving-ref attribution — the anchor-or-subject predicate ¶4 | a | governs-scope | heading within the binding body |
| * | core/verification-claim-integrity.md | 2.1 Moving-ref attribution — the anchor-or-subject predicate ¶6 | a | governs-scope | heading within the binding body |
| * | core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree ¶4 | a | governs-scope | the "either:" options of the MUST |
| * | core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree ¶5 | a | governs-scope | citation requirement of the obligation |
| * | core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree ¶6 | b | narrows-scope | where it does not bind |
| * | core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree ¶7 | b | narrows-scope | not a substitute for the tooling verdict |
| * | core/verification-claim-integrity.md | 3. The 5-Section Evidence-Bearing Report Format ¶2 | a | governs-scope | intro of the section table |
| * | core/verification-claim-integrity.md | 3. The 5-Section Evidence-Bearing Report Format ¶3 | a | governs-scope | the five sections the [HARD] line names |
| * | core/verification-claim-integrity.md | 3.1 Refused-tool degradation — a refusal is a Gap, never a silent substitution ¶2 | a | governs-scope | interprets the [HARD] line |
| * | core/verification-claim-integrity.md | 3.1 Refused-tool degradation — a refusal is a Gap, never a silent substitution ¶4 | a | governs-scope | intro of the consequences |
| * | core/verification-claim-integrity.md | 3.1 Refused-tool degradation — a refusal is a Gap, never a silent substitution ¶5 | b | narrows-scope | a refusal does not block the verdict |
| * | workflow/cache-aware-execution.md | Cache-Aware Execution | b | narrows-scope | design.md §4.3 (b): "change no gate semantics" + loading scope |
| * | workflow/cache-aware-execution.md | Non-goals | b | narrows-scope | bounds the directives (never justify skipping a gate) |
| * | workflow/context-window-management.md | Context Window Management | a | governs-scope | title + audience scope |
| * | workflow/context-window-management.md | Context Window Targets ¶2 | a | governs-scope | the table the [HARD] threshold line depends on |
| * | workflow/context-window-management.md | Context Window Targets ¶3 | a | governs-scope | tie-break rule for the table |
| * | workflow/context-window-management.md | User Responsibilities ¶1 | a | governs-scope | intro restating the thresholds the [HARD] lines use |
| * | workflow/context-window-management.md | Orchestrator Responsibilities ¶4 | a | governs-scope | part of the resume-format [HARD] body |
| * | workflow/context-window-management.md | Applies To | a | governs-scope | applicability of the whole rule |
| * | workflow/cross-session-messaging.md | Cross-Session Messaging | a | governs-scope | scope + loading scope |
| * | workflow/cross-session-messaging.md | Rules ¶6 | b | narrows-scope | role-boundary dispatch is permitted |
| * | workflow/cross-session-messaging.md | A send result has three shapes, and none of them says "read" ¶2 | a | governs-scope | the shapes table the [HARD] line relies on |
| * | workflow/cross-session-messaging.md | An idle notice is a scheduling hint ¶1 | a | governs-scope | defines notify_when_idle the [HARD] line refers to |
| * | workflow/cross-session-messaging.md | Codex broker path (session messaging tools) | a | governs-scope | extends every rule above to the Codex broker (scope) |
| * | workflow/goal-directive.md | Hard Preconditions for Every Recommendation | b | narrows-scope | design.md §4.3 (b) |
| * | workflow/kanban-dispatch.md | Kanban Dispatch Protocol ¶1 | a | governs-scope | what the rule governs |
| * | workflow/kanban-dispatch.md | Kanban Dispatch Protocol ¶2 | a | governs-scope | loading scope |
| * | workflow/kanban-dispatch.md | Scope — when this rule is live | b | narrows-scope | design.md §4.3 (b) remote governance |
| * | workflow/kanban-dispatch.md | Entry into the board is an operator act ¶1 | a | governs-scope | grounds the sole-producer [HARD] line |
| * | workflow/kanban-dispatch.md | Entry into the board is an operator act ¶5 | b | narrows-scope | what is not a silent promotion |
| * | workflow/kanban-dispatch.md | Card classes — not every card needs every column ¶1 | a | governs-scope | class definitions the Class A [HARD] line uses |
| * | workflow/kanban-dispatch.md | Card classes — not every card needs every column ¶4 | b | narrows-scope | Class B skips plan, not review |
| * | workflow/kanban-dispatch.md | The dispatch cycle | a | governs-scope | parent heading of binding subsections |
| * | workflow/kanban-dispatch.md | The delegation channel is the queue ¶2 | a | governs-scope | elaborates the [HARD] queue line |
| * | workflow/kanban-dispatch.md | Dispatch language ¶2 | b | narrows-scope | what stays verbatim |
| * | workflow/kanban-dispatch.md | Dispatch format ¶2 | a | governs-scope | the fields the [HARD] line names |
| * | workflow/kanban-dispatch.md | Completion is read, never trusted ¶2 | a | governs-scope | operationalizes the [HARD] line |
| * | workflow/kanban-dispatch.md | Completion is read, never trusted ¶3 | a | governs-scope | operationalizes the [HARD] line |
| * | workflow/kanban-dispatch.md | Completion is read, never trusted ¶4 | a | governs-scope | verdict home of the [HARD] line |
| * | workflow/kanban-dispatch.md | CodeRabbit is not read from `gh pr checks` ¶2 | a | governs-scope | condition 1 |
| * | workflow/kanban-dispatch.md | CodeRabbit is not read from `gh pr checks` ¶3 | a | governs-scope | condition 1 command |
| * | workflow/kanban-dispatch.md | CodeRabbit is not read from `gh pr checks` ¶4 | a | governs-scope | condition 2 |
| * | workflow/kanban-dispatch.md | CodeRabbit is not read from `gh pr checks` ¶5 | a | governs-scope | outcome of the conditions |
| * | workflow/kanban-dispatch.md | The `/clear` handoff between phases ¶2 | a | governs-scope | part of the [HARD] handoff |
| * | workflow/kanban-dispatch.md | The `/clear` handoff between phases ¶3 | a | governs-scope | extends the [HARD] handoff to the lead |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶2 | a | governs-scope | forms table the [HARD] launcher line uses |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶3 | a | governs-scope | creation verb of the [HARD] line |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶8 | a | governs-scope | slug rule intro |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶10 | b | narrows-scope | directory keeps the card id |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶12 | a | governs-scope | the three carriers |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶13 | a | governs-scope | consequence of the carriers rule |
| * | workflow/kanban-dispatch.md | Isolation is provisioned by MoAI, then entered through a launcher ¶15 | b | narrows-scope | binds card-delivering PRs only |
| * | workflow/kanban-dispatch.md | The env-isolated verification form ¶2 | a | governs-scope | the form the [HARD] line requires |
| * | workflow/kanban-dispatch.md | The env-isolated verification form ¶3 | a | governs-scope | load-bearing property of the form |
| * | workflow/kanban-dispatch.md | The env-isolated verification form ¶4 | a | governs-scope | defines which form is standard |
| * | workflow/kanban-dispatch.md | Integration into the release branch is self-served ¶3 | a | governs-scope | the integration rules under the [HARD] line |
| * | workflow/kanban-dispatch.md | Integration into the release branch is self-served ¶7 | a | governs-scope | completion signal of the [HARD] integration |
| * | workflow/kanban-dispatch.md | Factory Mode — the card travels whole | b | narrows-scope | design.md §4.3 (b): lane spawn authority |
| * | workflow/kanban-dispatch.md | Boundaries — what this protocol does not do | b | narrows-scope | design.md §4.3 (b) |
| * | workflow/main-checkout-branch-guard.md | Main-Checkout Branch Guard | a | governs-scope | title + loading scope |
| * | workflow/main-checkout-branch-guard.md | Rules ¶2 | a | governs-scope | the forbidden table |
| * | workflow/main-checkout-branch-guard.md | Rules ¶3 | b | narrows-scope | permitted list intro |
| * | workflow/main-checkout-branch-guard.md | Rules ¶4 | b | narrows-scope | permitted list |
| * | workflow/main-checkout-branch-guard.md | Staleness Rule ¶2 | a | governs-scope | the commands of the [HARD] line |
| * | workflow/main-checkout-branch-guard.md | Staleness Rule ¶3 | a | governs-scope | action of the [HARD] line |
| * | workflow/main-checkout-branch-guard.md | Detecting Concurrent Sessions ¶2 | a | governs-scope | operative consequence of the MUST NOT line |
| * | workflow/session-handoff.md | Session Handoff Protocol | a | governs-scope | loading scope + companion pointer |
| * | workflow/session-handoff.md | When To Generate (5 Triggers) ¶2 | a | governs-scope | the triggers the MUST line names |
| * | workflow/session-handoff.md | When To Generate (5 Triggers) ¶3 | b | narrows-scope | when none apply |
| * | workflow/session-handoff.md | Emission-Time Save Obligation (auto-resume wiring) ¶3 | a | governs-scope | record semantics of the save obligation |
| * | workflow/session-handoff.md | Canonical Format (Verbatim Spec) ¶2 | a | governs-scope | the verbatim format the MUST line names |
| * | workflow/session-handoff.md | Invariants (both modes) | b | narrows-scope | design.md §4.3 (b) |
| * | workflow/session-handoff.md | Auto-Memory Integration (Mandatory) ¶2 | a | governs-scope | the "MUST also:" list |
| * | workflow/session-handoff.md | Output Surface (User-Facing) ¶2 | a | governs-scope | defines non-emission for the [HARD] line |
| * | workflow/session-handoff.md | Output Surface (User-Facing) ¶3 | a | governs-scope | the obligation still binds without a banner |
| * | workflow/session-handoff.md | Diet Constraints ¶2 | a | governs-scope | the constraints the [HARD] line names |
| * | workflow/skill-routing.md | (서문) | a | governs-scope | frontmatter carries paths: (acceptance.md §D.3) |
| * | workflow/skill-routing.md | Skill Routing Protocol | a | governs-scope | title + scope |
| * | workflow/skill-routing.md | 1. Orchestrator Obligation ¶2 | a | governs-scope | the instruction the MUST line injects |
| * | workflow/skill-routing.md | 1. Orchestrator Obligation ¶4 | b | narrows-scope | zero matches is valid |
| * | workflow/skill-routing.md | §1.1 — Orchestrator-Direct Skill Routing (non-spawn) ¶2 | a | governs-scope | the mandatory-skill table |
