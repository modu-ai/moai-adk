# MoAI-ADK v3.0.0 — Master Design Document (Ground-Up Redesign)

> Status: DRAFT (Wave 3 — fresh architect synthesis from 48K words of research)
> Supersedes: `.moai/design/v3-legacy/docs/major-v3-master-v1.md` (first v3 attempt)
> Authored: 2026-04-23
> Review target: Wave 4 SPEC generation (round 2, prefix `SPEC-V3R2-`)

---

## 0. Revision History

This document is the **second** ground-up architectural synthesis for v3. It supersedes the v3-legacy master drafted under the v3 prefix, which conflated CC-architecture mirroring with moai's unique SPEC-governed posture.

- **v1 context** — `major-v3-master-v1.md` in the v3-legacy tree focused on enumerating gaps against Claude Code's architecture and proposed direct adoption of most CC subsystems (bridge, feature flags, 52-subcommand CLI). Wave 1 re-read of CC (R3) surfaced that most of those subsystems are implementation details specific to CC's product shape, not structural patterns. v1 also predates the agency absorption and the sprint-contract evaluator memory problem (P-Z01) surfaced in this round.
- **Why v2 ground-up** — Round 2 research (R1 papers + R2 opensource + R3 CC re-read + R4-R6 moai audits) converged on 12 principles that v1 either missed or mis-ordered (notably Principle 4 "Evaluator Judgments Fresh, Contract State Durable" is new). Wave 2 synthesis produced 72 specific problems in moai v2.13.2 state and 37 pattern candidates with explicit ADOPT/CONSIDER/NOT-NOW disposition. This document is the first artifact to work from that corpus in full.
- **Scope** — This master document defines the seven-layer architecture, identifies the target SPEC set (~35 SPECs), catalogs breaking changes (BC-V3R2-*), and proposes the phase plan. It does not generate SPEC bodies (Wave 4) and does not implement any code.

---

## 1. Executive Summary

### 1.1 v3 North Star

MoAI-ADK v3 is a **SPEC-governed, language-neutral, harness-routed development orchestrator that rides on Claude Code's turn runtime**. It treats the EARS-formatted SPEC document — with its Given/When/Then acceptance criteria and associated @MX tags — as the constitutional contract that every phase (plan, run, sync, design, review, fix, loop, db, mx) reads from and writes to. v3 composes seven structural ideas: (a) Ralph-style file-first persistence with fresh-context iteration for run/loop phases; (b) MetaGPT-style role-specialized multi-agent orchestration for open-ended plan/design work; (c) Agent-as-a-Judge independent adversarial evaluation with **fresh judgments per iteration** and durable Sprint Contract state; (d) Constitutional AI-style declarative governance with FROZEN/EVOLVABLE zones; (e) SWE-agent-style Agent-Computer Interface layered over LSP + @MX + structured hook JSON; (f) LangGraph-style typed state checkpointed at phase boundaries with `interrupt()`-equivalent blocker reports surfacing to AskUserQuestion; (g) TRUST 5 quality gates routed by a 3-level harness system that scales evaluation depth with SPEC complexity.

Moai v3's unique position: the only agentic development orchestrator that combines a constitutional SPEC contract with adversarial skeptical evaluation, typed phase transitions, sandboxed execution by default, and language-neutral harness routing — betting that explicit, externalized intent plus adversarial verification produces more correct software than any autonomy-maximizing agent, at a complexity cost bounded by harness routing and file-first primitives. *Source: synthesis/design-principles.md §v3 North Star.*

### 1.2 What changes vs v2.13.2

| Axis | v2 state | v3 state | Source |
|------|----------|----------|--------|
| Hook protocol | Exit-code-only | JSON-OR-ExitCode dual protocol | R3 §2 Dec 5, P-H05 |
| Permission model | Flat `permissions.allow` | 8-source stack with provenance + `bubble` mode | R3 §4 Adopt 2, P-C01 |
| Execution sandbox | None (approval-only) | Bubblewrap/Seatbelt/Docker default for implementers | R2 §A#5, P-C03 |
| Skills | 48 (platform triplets, thinking triplet, kitchen-sink domains) | ~24 (merge clusters 1-5) | R4 verdict, P-S01..S14 |
| Agents | 22 (9 protocol violations, 3 near-identical builders, 4 advisor-only) | ~17 (common-protocol scrub, `manager-cycle`, `builder-platform`) | R5 verdict, P-A01..A23 |
| Evaluator memory | Cascading across iterations | Fresh per-iteration; Sprint Contract state durable | R1 §9 anti-pattern, P-Z01 |
| Acceptance criteria | Flat Given/When/Then list | Hierarchical (Agent-as-a-Judge 365-sub-req shape) | R1 §9, E-1 |
| Workflow surface | Subcommands only (15) | Subcommands × `--mode {autopilot,loop,team}` | R2 §3 OMC, O-4 |
| Utility subcommands | Multi-agent default | Fixed Agentless pipeline (`/moai fix|coverage|mx|codemaps|clean`) | R1 §25, O-6, P-Z02 |
| Config loaders | 5 yaml sections template-only (constitution/context/interview/design/harness) | All 23 sections typed + Go loaders | R6 §5.2, P-H06 |
| Memory model | Informal MEMORY.md | Typed taxonomy (user/feedback/project/reference) + staleness caveat | R3 §4 Adopt 7, M-1, M-5 |
| Tool layer | Raw `Bash` exposed to agents | 6 canonical ACI commands + LSP + ast-grep, Bash as escape hatch | R1 §11, T-1 |
| Migration | Explicit `moai migrate agency` | Versioned silent auto-apply at session-start | R3 §2 Dec 10, X-5 |
| Hardcoded path | `/Users/goos/go/bin/moai` in 26 shell wrappers | `$HOME/go/bin/moai` primary fallback | R6 §2.1, P-H04 |
| Rule tree | 34 files, duplicate workflow-modes/spec-workflow, misfiled lsp-client.md | 31 files, consolidated | R6 §4, P-H10..H14 |

### 1.3 What stays FROZEN

v3 preserves moai's invariants. Each of these is load-bearing and cannot change without constitutional amendment (see Layer 1).

- **SPEC system + EARS format** — REQ-*, modality vocabulary (ubiquitous/event/state/optional/unwanted), Given/When/Then acceptance. *Source: `.claude/rules/moai/workflow/spec-workflow.md`, design-principles Principle 1.*
- **TRUST 5 framework** — Tested, Readable, Unified, Secured, Trackable. *Source: `.claude/rules/moai/core/moai-constitution.md`, design-principles Principle 12.*
- **@MX TAG protocol** — NOTE / WARN (+REASON) / ANCHOR / TODO / LEGACY sub-lines, autonomous agent add/update/remove without human approval. *Source: `.claude/rules/moai/workflow/mx-tag-protocol.md`.*
- **16-language neutrality** — go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift — equal citizens in `internal/template/templates/`. *Source: CLAUDE.local.md §15.*
- **Template-First discipline** — every `.claude/`, `.moai/`, `.agency/` file has a twin under `internal/template/templates/`. Local drift is a CI error. *Source: CLAUDE.local.md §2.*
- **AskUserQuestion-Only-for-Orchestrator** — subagents MUST NOT prompt; missing inputs must surface as structured blocker reports. *Source: `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary.*
- **Claude Code as execution substrate** — moai rides on CC's turn runtime; never replaces it. *Source: R3 §6 "moai's big idea".*

### 1.4 Release shape

Nine phases, tagged alpha.1 through GA. Phase labels are ordering, not durations.

| Phase | Tag | Theme |
|-------|-----|-------|
| 1 | v3.0.0-alpha.1 | Constitution & Foundation (FROZEN zones, settings layering, migration framework) |
| 2 | v3.0.0-alpha.2 | Runtime Hardening (hook JSON, 27-event coverage, sandbox layer) |
| 3 | v3.0.0-alpha.3 | Agent Cleanup (22 → 17, common-protocol scrub, effort calibration) |
| 4 | v3.0.0-alpha.4 | Skill Consolidation (48 → 24, merge clusters 1-5, retirements) |
| 5 | v3.0.0-beta.1 | Harness + Evaluator (Sprint Contract fresh judgment, 3-level routing) |
| 6 | v3.0.0-beta.2 | Multi-Mode Workflow (/moai loop Ralph, /moai design, team mode) |
| 7 | v3.0.0-rc.1 | Extension (output styles, ACI tools, memdir) |
| 8 | v3.0.0-rc.2 | Migration Tool + Docs (v2 → v3 migrator, release docs) |
| 9 | v3.0.0 | GA |

---

## 2. Vision & Positioning

### 2.1 Mission Statement

Deliver a development orchestrator for Claude Code that makes *SPEC-driven, adversarially-verified, sandboxed, language-neutral* the default posture for any codebase — without asking the user to learn a framework. The SPEC is the contract. TRUST 5 is the quality floor. Claude Code is the runtime. moai is the workflow.

*Source: synthesis/design-principles.md Preamble + Principle 1.*

### 2.2 Differentiation Map

| | moai v3 | Claude Code | Ralph (iannuttall) | oh-my-claudecode | Aider | SWE-agent / Agentless |
|---|---|---|---|---|---|---|
| Core abstraction | SPEC contract + phases | Turn runtime | Outer loop + prd.json | Multi-mode router + 6-tier memory | Git-commit-native chat | Fixed 3-phase pipeline |
| Intent source of truth | EARS SPEC + hierarchical acceptance | User prompt | prd.json | User prompt | User prompt + repo-map | GitHub issue text |
| Evaluation | evaluator-active + Sprint Contract + TRUST 5 | None (runtime only) | None | None | `/undo` via git | Fail-to-pass test delta |
| Memory | Typed taxonomy (user/feedback/project/reference) + staleness | memdir (user/feedback/project/reference) | `progress.md` + `.ralph/` files | 6-tier + session JSONL | Repo-map + chat | History processor compression |
| Parallelism | DAG from plan phase + team-mode file ownership | Subagents + teams | Sequential-only | Tmux panes | None | None |
| Sandboxing | Bubblewrap/Seatbelt/Docker default | None (permissions.allow) | None | None | Git undo | Docker for benchmark only |
| Language neutrality | 16-language parity FROZEN | Any (via tool names) | Delegates | Primarily English | 100+ via Tree-sitter | Python-only |
| Surface | 15 subcommands × 3 modes | 52 subcommands | 3 skills | 6 modes + 10+ shortcuts | 9 slash commands | ACI (open/scroll/edit/search) |
| Extension axis | Skills + agents + hooks + plugins (deferred) | Plugins (3 origins) + skills + agents | Skills (install-time) | OpenClaw hook bridge + tmux | MCP servers | ACI iteration |
| Quality gate | TRUST 5 harness (minimal/standard/thorough) | Permission approvals | None | Stop-triggers | Review each atomic commit | Linter + test runner |

*Sources: r2-opensource-tools.md §C Design-space taxonomy; r3-cc-architecture-reread.md §6.*

### 2.3 FROZEN Zones (what v3 MUST preserve)

Listed verbatim in §1.3. In architectural terms these manifest as:

- Layer 1 constitution codifies all 7 as FROZEN clauses; amendment requires the 5-layer safety gate (S-5) and human approval.
- Layer 2 enforces SPEC + EARS schema via `moai spec lint`.
- Layer 3 enforces TRUST 5 via harness routing and agent common-protocol lint.
- Layer 6 preserves AskUserQuestion monopoly via CI check that scans agent bodies.
- Layer 7 preserves Template-First via `diff -rq` CI check (already exists per CLAUDE.local.md §2).
- 16-language neutrality flows through Layer 3 (ACI commands are LSP-backed, powernap-abstracted), Layer 5 (evaluator rubrics per-language), and Layer 6 (language rules in `.claude/rules/moai/languages/`, not skill tree).

---

## 3. Design Principles

Twelve principles govern v3 design decisions. Each is grounded in ≥2 Wave 1 sources. They form a partial order: Tier 1 principles constrain lower tiers when conflicts arise. *Source: synthesis/design-principles.md.*

**Tier 1 — FROZEN invariants**

1. **SPEC as Constitutional Contract** — SPEC with EARS requirements, Given/When/Then acceptance, and typed artifacts is the single source of truth across phases. Context and memory derive from SPEC; they do not override it. *(MetaGPT R1 §7; Constitutional AI R1 §18; R3 §6.)*
2. **Constitutional Governance with FROZEN/EVOLVABLE Zones** — rules live in declarative constitution files with explicit zones; every amendment is auditable, reversible, canary-tested. *(Constitutional AI R1 §18; ADAS R1 §16 anti-pattern flag; moai design-constitution v3.3.0 §2, §5.)*
3. **Sandboxed Execution as Default Safety Layer** — implementer agents execute in ephemeral isolated sandbox; approval prompts alone are empirically exploitable. *(R2 §A Pattern 5; OWASP Top 10 for Agentic Apps 2025; Cline npm-token exfiltration R2 §5.)*

**Tier 2 — Structural invariants**

4. **Interface Design Over Tool Count (ACI)** — ACI matters more than number of tools or model capability; design LM-optimized interfaces, not human IDE passthroughs. *(SWE-agent R1 §11, R2 §8 — 8× empirical improvement.)*
5. **Evaluator Judgments Fresh, Contract State Durable** — Agent-as-a-Judge evaluators start each iteration with no memory of previous judgments; Sprint Contract carries criteria state forward, not reasoning traces. *(Agent-as-a-Judge R1 §9 anti-pattern "errors cascade"; Reflexion R1 §2.)*
6. **Typed State + Durable Checkpoint at Phase Boundaries** — cross-phase state is a typed schema with immutable updates; `interrupt()`-equivalent surfaces to AskUserQuestion. *(LangGraph R2 §16; DSPy R1 §17; MS Agent Framework R2 §12.)*
7. **Permission Bubble Over Bypass** — permissions resolve through 8-source stack with provenance; escalate to parent terminal on ambiguity rather than default-allow. *(R3 §4 Adopt 2; R3 §2 Dec 15 bubble mode.)*
8. **Hook Output = JSON Protocol (Exit Codes Remain as Fallback)** — hook handlers communicate via structured JSON capable of carrying `additionalContext`, `permissionDecision`, `updatedInput`, `systemMessage`, `continue`. Exit codes preserved for backward compat. *(R3 §2 Dec 5, §4 Adopt 4.)*

**Tier 3 — Optimization principles**

9. **Fresh-Context Iteration Over Session Accumulation** — long-horizon tasks decompose into iterations each starting from clean LM context; state persists on disk (SPEC, MX tags, checkpoint files, git). *(Ralph R2 §1-2; Reflexion R1 §2; CC BigQuery 250K wasted calls/day R3 §3.2.)*
10. **Parallelism via Explicit Dependency DAG** — concurrent tasks represented as explicit DAG with `reads: [paths]` / `writes: [paths]` per agent spawn; not implicit ordering. *(LLMCompiler R1 §15 — 3.7× latency; R3 §2 Dec 9 worktree orthogonality.)*
11. **Agent Count Matches Task Structure** — well-structured tasks (lint, coverage gap, mx) use Agentless fixed pipelines; open-ended tasks (plan, design) use multi-agent role specialization. *(Agentless R1 §25; R1 §Y divergence matrix.)*
12. **File-First Primitives Over Framework Abstractions** — prefer git + disk + markdown + YAML substrate; reject frameworks requiring >~500 LOC scaffolding. *(Ralph R2 §1; R2 §B anti-pattern 9 "framework-over-primitives"; CC memdir markdown R3 §1.1.)*

**Conflict resolution**: Tier 1 always wins over Tier 2/3. Same-tier conflicts require SPEC-documented trade-off decisions. Principle 3 (Fresh context) ↔ Principle 6 (Durable checkpoint) resolve via "disk durable, LM context ephemeral." Principle 6/7 (typed state, sandbox) ↔ Principle 12 (file-first) resolve via "keep frameworks thin; total framework budget ≤ 500 LOC." *Source: synthesis/design-principles.md §Principle interaction matrix.*

---

## 4. v3 Architecture — 7 Layers

The architecture is a bottom-up stack. Each layer depends only on layers below it. Cross-cutting concerns (§5) thread across layers but do not introduce new horizontal dependencies.

```
 ┌─────────────────────────────────────────────────────────────────┐
 │ Layer 7: Extension   (plugins, output styles, memdir, migrations)│
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 6: Workflow    (plan/run/sync/design/review/fix/loop/...) │
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 5: Harness     (minimal/standard/thorough routing)         │
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 4: Orchestration (agents, teams, tasklist, mailbox)        │
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 3: Runtime     (hooks, permissions, session, sandbox)      │
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 2: SPEC & TAG  (EARS contract + code annotation)           │
 ├─────────────────────────────────────────────────────────────────┤
 │ Layer 1: Constitution (FROZEN / EVOLVABLE zone model)            │
 └─────────────────────────────────────────────────────────────────┘
```

### Layer 1: Constitution

**Purpose.** Define the invariants that no automation may change. Provide the vocabulary and safety gates for evolving everything else.

**Owned SPECs.** `SPEC-V3R2-CON-001`, `-CON-002`, `-CON-003`.

**Core types.**

```go
// internal/constitution/zone.go
type Zone uint8
const (
    ZoneFrozen Zone = iota  // immutable by automation
    ZoneEvolvable           // amendable via graduation protocol
)

type Rule struct {
    ID         string   // CONST-V3R2-NNN
    Zone       Zone
    File       string   // .claude/rules/moai/core/moai-constitution.md#anchor
    Clause     string   // the HARD-rule text
    CanaryGate bool     // does amendment require shadow eval?
}

type AmendmentProposal struct {
    RuleID       string
    Before       string
    After        string
    Evidence     []Observation
    CanaryResult CanaryVerdict
    Contradicts  []string   // rule IDs in conflict
    Approved     bool       // human approval required
}
```

**Key interfaces.**
- `FrozenGuard.Check(file, clause) bool` — reject writes to FROZEN zone.
- `Canary.Evaluate(proposal) CanaryVerdict` — shadow-eval against last 3 completed SPECs.
- `ContradictionDetector.Scan(proposal) []Conflict` — flag rule-level contradictions.
- `RateLimiter.Admit() bool` — enforce 3/week, 24h cooldown, 50 active learnings cap.
- `HumanOversight.Approve(proposal) bool` — via AskUserQuestion at the orchestrator.

**Mapped moai subsystem.** `.claude/rules/moai/core/moai-constitution.md` (core invariants) + `.claude/rules/moai/design/constitution.md` v3.3.0 (design subsystem, already FROZEN/EVOLVABLE codified per SPEC-AGENCY-ABSORB-001). v3 generalizes the zone model from design subsystem to core. *Source: pattern-library.md S-4, S-5.*

```
 ┌─────────────────────────────────────────┐
 │ ConstitutionFile ▶ parses HARD clauses  │
 │        │                                │
 │        ▼                                │
 │    Zone {Frozen, Evolvable}             │
 │        │                                │
 │        ▼                                │
 │  5-Layer Safety Gate                    │
 │    1 FrozenGuard                        │
 │    2 Canary (shadow 3 SPECs)            │
 │    3 ContradictionDetector              │
 │    4 RateLimiter (3/week)               │
 │    5 HumanOversight (AskUserQuestion)   │
 │        │                                │
 │        ▼                                │
 │  ApplyAmendment → commit + evolution-log│
 └─────────────────────────────────────────┘
```

### Layer 2: SPEC & TAG

**Purpose.** Encode intent (SPEC) and in-code invariants (TAG) as durable, parseable, agent-consumable artifacts. This is what makes moai SPEC-governed rather than prompt-governed.

**Owned SPECs.** `SPEC-V3R2-SPC-001`, `-SPC-002`, `-SPC-003`, `-SPC-004`.

**Core types.**

```go
// internal/spec/ears.go
type Modality uint8
const (
    ModUbiquitous Modality = iota  // "The system SHALL ..."
    ModEvent                        // "WHEN X, the system SHALL ..."
    ModState                        // "WHILE Y, the system SHALL ..."
    ModOptional                     // "WHERE Z, the system SHALL ..."
    ModUnwanted                     // "IF X, THEN the system SHALL NOT ..."
)

type EARSRequirement struct {
    ID         string           // REQ-V3R2-xxx
    Modality   Modality
    Condition  string
    Response   string
    Acceptance []Acceptance     // Agent-as-a-Judge hierarchical shape (R1 §9)
}

type Acceptance struct {
    ID       string
    Given    string
    When     string
    Then     string
    Children []Acceptance       // nested sub-requirements
}

// internal/mx/tag.go
type TagKind uint8
const (
    MXNote TagKind = iota
    MXWarn                        // requires MXReason sub-line
    MXAnchor                      // high fan_in (>=3 callers)
    MXTodo                        // incomplete; resolved in GREEN
    MXLegacy                      // no SPEC traceability
)

type Tag struct {
    Kind      TagKind
    File      string
    Line      int
    Body      string
    Reason    string              // MXWarn only
    AnchorID  string              // stable ID for cross-reference
    CreatedBy string              // agent name / "human"
}
```

**Key interfaces.**
- `SPECParser.Parse(path) (*SPEC, error)` — EARS syntax + acceptance hierarchy.
- `SPECLinter.Check(spec) []Violation` — missing modality, untestable Then, orphan acceptance.
- `TagScanner.Scan(file) []Tag` — locate @MX in source across 16 languages.
- `TagResolver.Resolve(anchorID) []Callsite` — fan-in analysis for ANCHOR promotion.
- `AcceptanceRunner.Exec(acceptance, artifact) Verdict` — run Given/When/Then against built code.

**Mapped moai subsystem.** `.moai/specs/SPEC-*/spec.md` + `.claude/rules/moai/workflow/mx-tag-protocol.md` + `/moai mx` subcommand. v3 upgrades acceptance from flat to hierarchical (enables E-1 Agent-as-a-Judge scoring per sub-criterion).

### Layer 3: Runtime

**Purpose.** Safely execute model turns on the user's machine. This is the dense layer — hooks, permissions, session state, sandbox all live here.

**Owned SPECs.** `SPEC-V3R2-RT-001`, `-RT-002`, `-RT-003`, `-RT-004`, `-RT-005`, `-RT-006`, `-RT-007`.

**Core types.**

```go
// internal/hook/response.go
type HookResponse struct {
    AdditionalContext  string              `json:"additionalContext,omitempty"`
    PermissionDecision PermissionDecision  `json:"permissionDecision,omitempty"`
    UpdatedInput       map[string]any      `json:"updatedInput,omitempty"`
    SystemMessage      string              `json:"systemMessage,omitempty"`
    Continue           *bool               `json:"continue,omitempty"`
}

type PermissionDecision string
const (
    PermAllow PermissionDecision = "allow"
    PermAsk   PermissionDecision = "ask"
    PermDeny  PermissionDecision = "deny"
)

// internal/permission/stack.go
type Source uint8
const (
    SrcPolicy Source = iota  // managed-settings.json
    SrcUser                  // ~/.claude/settings.json
    SrcProject               // .claude/settings.json
    SrcLocal                 // .claude/settings.local.json
    SrcPlugin
    SrcSkill
    SrcSession
    SrcBuiltin
)

type PermissionRule struct {
    Pattern string  // e.g. "Bash(go test:*)"
    Action  PermissionDecision
    Source  Source
    Origin  string  // file path for provenance
}

type PermissionMode string
const (
    ModeDefault        PermissionMode = "default"
    ModeAcceptEdits    PermissionMode = "acceptEdits"
    ModeBypass         PermissionMode = "bypassPermissions"
    ModePlan           PermissionMode = "plan"
    ModeBubble         PermissionMode = "bubble"   // escalate to parent terminal
)

// internal/sandbox/context.go
type Sandbox string
const (
    SandboxNone        Sandbox = "none"
    SandboxBubblewrap  Sandbox = "bubblewrap"   // Linux
    SandboxSeatbelt    Sandbox = "seatbelt"     // macOS
    SandboxDocker      Sandbox = "docker"       // CI
)

// internal/session/state.go
type PhaseState struct {
    Phase       Phase       // plan | run | sync | design
    SPECID      string
    Checkpoint  Checkpoint  // typed snapshot
    BlockerRpt  *BlockerReport  // interrupt()-equivalent
    UpdatedAt   time.Time
}
```

**Key interfaces.**
- `HookDispatcher.Fire(event HookEvent, payload) HookResponse` — JSON-first, exit-code fallback.
- `PermissionResolver.Resolve(tool, input, ctx) HookResponse` — walk 8-source stack.
- `SandboxLauncher.Exec(cmd []string, sandbox Sandbox) (output, error)` — ephemeral isolated process.
- `SessionStore.Checkpoint(phase, state)` / `Hydrate(phase) State` — file-first at `.moai/state/`.

**Mapped moai subsystem.** `internal/hook/*.go` (27 handlers), `internal/cli/deps.go` (registration), `.claude/hooks/moai/*.sh` (26 shell wrappers), `.claude/settings.json` (25 event registrations). v3 addresses R6 audit findings: 10 logging-only stubs upgraded or retired; hardcoded `/Users/goos/go/bin/moai` replaced with `$HOME/go/bin/moai`; JSON protocol added; subagent-stop tmux pane cleanup bug fixed (P-H02). *Source: r6-commands-hooks-style-rules.md §2, problem-catalog.md Cluster 3.*

```
 ┌──────────────────────────────────────────────────────┐
 │ Event → ShellWrapper → moai hook <event>             │
 │                              │                        │
 │                              ▼                        │
 │                   HookDispatcher (Go handler)         │
 │                              │                        │
 │         ┌────────────────────┼────────────────────┐  │
 │         ▼                    ▼                    ▼  │
 │   PermissionStack     SandboxLauncher      SessionStore│
 │   (8-source, prov.)   (bwrap/seatbelt)    (.moai/state)│
 │         │                    │                    │  │
 │         └────────────────────┼────────────────────┘  │
 │                              ▼                        │
 │              HookResponse (JSON or exit code)         │
 └──────────────────────────────────────────────────────┘
```

### Layer 4: Orchestration

**Purpose.** Launch and coordinate agents; manage teams, mailboxes, and task ledgers. This is where CC's `Agent()` tool is wrapped and moai's team primitives (`TeamCreate`, `SendMessage`, `TaskCreate/Update/List/Get`) live.

**Owned SPECs.** `SPEC-V3R2-ORC-001`, `-ORC-002`, `-ORC-003`, `-ORC-004`, `-ORC-005`.

**Core types.**

```go
// internal/agent/frontmatter.go
type AgentFrontmatter struct {
    Name           string
    Description    string
    Model          string            // sonnet | opus
    Effort         Effort            // low | medium | high | xhigh | max
    PermissionMode PermissionMode
    Isolation      Isolation         // none | worktree
    Sandbox        Sandbox           // inherits Layer 3 enum
    Memory         MemoryScope       // none | project | user
    Tools          []string          // CSV allowed-tools
    Skills         []string          // preloaded skills (YAML list)
    Hooks          []HookBinding     // agent-scoped PreToolUse/PostToolUse
}

// internal/team/team.go
type Team struct {
    ID          string
    Config      TeamConfig           // ~/.claude/teams/{id}/config.json
    Teammates   map[string]Teammate  // name → config entry
    Mailboxes   map[string]Mailbox
    TaskList    *TaskList
}
```

**Key interfaces.**
- `AgentSpawner.Spawn(name, prompt, overrides) AgentHandle` — wraps CC `Agent()` tool.
- `CommonProtocolLinter.Check(agent) []Violation` — reject `AskUserQuestion` in body, `Agent` in tools, missing `effort`.
- `TeamCoordinator.Create(roster []string) Team` — static or dynamic roster.
- `Mailbox.Send(target, msg)` / `Recv() Message` — SendMessage abstraction.
- `TaskList.Claim() *Task` — FIFO with dependency check.

**Mapped moai subsystem.** `.claude/agents/moai/*.md` (22 → 17 in v3). v3 merges builders (3 → 1 `builder-platform`), merges manager-ddd/tdd (→ `manager-cycle`), retires advisor-only experts (expert-debug → manager-quality sub-mode; expert-testing folded into manager-cycle + expert-performance). *Source: r5-agent-audit.md §Recommended v3 agent inventory, problem-catalog.md Cluster 6.*

Dynamic team generation (SPEC-TEAM-001 in v2) preserved: teammates spawned via `Agent(subagent_type: "general-purpose")` with runtime overrides from `workflow.yaml` role profiles (researcher / analyst / architect / implementer / tester / designer / reviewer). No static team agent files. *Source: CLAUDE.md §4 Dynamic Team Generation.*

### Layer 5: Harness

**Purpose.** Route quality depth (minimal / standard / thorough) based on SPEC complexity and user opt-in. Own the evaluator-active flow, Sprint Contract negotiation, and GAN loop contract.

**Owned SPECs.** `SPEC-V3R2-HRN-001`, `-HRN-002`, `-HRN-003`.

**Core types.**

```go
// internal/harness/level.go
type Level string
const (
    LevelMinimal  Level = "minimal"
    LevelStandard Level = "standard"
    LevelThorough Level = "thorough"
)

// internal/harness/sprint.go
type SprintContract struct {
    Iteration int
    Checklist []CriterionState        // per acceptance criterion
    Priority  EvalDimension           // DesignQuality | Originality | Completeness | Functionality
    Tests     []TestScenario          // Playwright or language-specific
    // FRESH JUDGMENT: evaluator sees no prior judgment reasoning
    EvaluatorMemoryScope MemoryScope  // always PerIteration per Principle 4
}

type CriterionState struct {
    CriterionID string
    Status      Status                // Passed | Failed | Refined | New
    Evidence    string                // test output, artifact paths
    Score       float64               // 0.0-1.0, rubric-anchored
}

// internal/harness/evaluator.go
type EvaluatorProfile struct {
    Name          string              // default | strict | lenient | frontend
    PassThreshold float64             // floor 0.60 FROZEN (design-constitution §5)
    MaxIterations int                 // default 5
    Escalation    int                 // default 3
    StrictMode    bool
    Rubrics       map[string]Rubric   // per-criterion 0.25/0.50/0.75/1.0 anchors
}
```

**Key interfaces.**
- `HarnessRouter.Route(spec) Level` — complexity estimator.
- `SprintContractNegotiator.Propose(spec, iter) SprintContract` — evaluator-active proposes; Actor may request adjustment.
- `EvaluatorRunner.Score(contract, artifact) ScoreCard` — fresh context per iteration.
- `GanLoop.Execute(spec) Result` — Builder/Evaluator with max_iterations cap.

**Mapped moai subsystem.** `.moai/config/sections/harness.yaml` + `.moai/config/evaluator-profiles/` + `evaluator-active` agent + `moai-workflow-gan-loop` skill + `.claude/rules/moai/design/constitution.md` §11 (Sprint Contract Protocol). v3 amends constitution §11.4 to explicitly scope evaluator memory per-iteration, closing P-Z01. v3 adds Go loader for harness.yaml (currently template-only per R6 §5.2). *Source: problem-catalog.md Cluster 4.*

### Layer 6: Workflow

**Purpose.** The user-visible surface. Subcommands, skills, modes, the plan/run/sync pipeline, and all utility subcommands.

**Owned SPECs.** `SPEC-V3R2-WF-001`, `-WF-002`, `-WF-003`, `-WF-004`, `-WF-005`, `-WF-006`.

**Core types.**

```go
// internal/workflow/mode.go
type Mode string
const (
    ModeAutopilot Mode = "autopilot"  // single-lead
    ModeLoop      Mode = "loop"       // Ralph fresh-context
    ModeTeam      Mode = "team"       // multi-agent DAG
)

type Command struct {
    Name       string     // plan | run | sync | ...
    SkillRoute string     // moai:plan
    MultiAgent bool       // vs Agentless fixed pipeline
    AllowedModes []Mode
}
```

**Subcommand classification.**

| Subcommand | Multi-agent? | Fixed pipeline? | Default mode |
|---|---|---|---|
| `/moai plan` | Yes | No | autopilot |
| `/moai run` | Yes | No | auto (by harness) |
| `/moai sync` | Yes | No | autopilot |
| `/moai design` | Yes | No | autopilot |
| `/moai db` | No | Yes | — |
| `/moai project` | No (interview + doc-gen) | Partial | — |
| `/moai fix` | **No (Agentless)** | Yes | — |
| `/moai coverage` | **No (Agentless)** | Yes | — |
| `/moai codemaps` | **No (Agentless)** | Yes | — |
| `/moai mx` | **No (Agentless)** | Yes | — |
| `/moai clean` | **No (Agentless)** | Yes | — |
| `/moai loop` | **No (outer Ralph)** | Yes | loop |
| `/moai review` | Partial (expert-security agent + checklist) | Partial | autopilot |
| `/moai e2e` | No | Yes | — |
| `/moai feedback` | No | Yes | — |

*Source: pattern-library.md O-6 Agentless, design-principles.md Principle 11.*

**Skill consolidation (48 → 24).** Five merge clusters from r4-skill-audit.md:

1. Thinking triplet (foundation-thinking + foundation-philosopher + workflow-thinking) → `moai-foundation-thinking`.
2. Design cluster (design-craft + domain-uiux) → `moai-design-system`; retain brand-design and copywriting (agency absorption contracts FROZEN).
3. Database cluster (domain-database + platform-database-cloud) → `moai-domain-database`.
4. Templates/docs (workflow-templates + docs-generation) → `moai-workflow-project` + `moai-workflow-jit-docs`.
5. Design tools (moai-design-tools) → split to `moai-tool-figma` + absorb Pencil side into `moai-workflow-pencil-integration`.

**Retired.** moai-tool-svg, moai-docs-generation, moai-foundation-philosopher (merged), moai-workflow-thinking (merged), moai-workflow-templates (merged), moai-platform-database-cloud (merged).

**Mapped moai subsystem.** `.claude/commands/moai/*.md` (15 thin wrappers) + `.claude/skills/` (48 → 24). `/98-github.md` and `/99-release.md` (698 + 890 LOC dev-only commands violating thin-wrapper pattern) extracted to `moai-workflow-github` and `moai-workflow-release` skills. *Source: r6-commands-hooks-style-rules.md §1.2, problem-catalog.md P-H08, P-H09.*

### Layer 7: Extension

**Purpose.** Controlled extension points — output styles, memdir, migrations. Plugins deferred.

**Owned SPECs.** `SPEC-V3R2-EXT-001`, `-EXT-002`, `-EXT-003`, `-EXT-004`.

**Core types.**

```go
// internal/memdir/taxonomy.go
type MemoryType string
const (
    MemUser      MemoryType = "user"       // about the developer
    MemFeedback  MemoryType = "feedback"   // correction patterns
    MemProject   MemoryType = "project"    // current work state
    MemReference MemoryType = "reference"  // external pointers
)

type MemoryEntry struct {
    Type      MemoryType
    Body      string
    CreatedAt time.Time
    UpdatedAt time.Time
    Stale     bool   // age > staleness_window_days
}

// internal/outputstyle/style.go
type OutputStyle struct {
    Name                    string
    Description             string
    KeepCodingInstructions  bool
    ForceForPlugin          string  // plugin name auto-applies this style
    Body                    string
}

// internal/migration/runner.go
type Migration struct {
    Version int
    Name    string
    Apply   func(projectRoot string) error  // idempotent
}
```

**Key interfaces.**
- `MemdirStore.Retrieve(query) []MemoryEntry` — LLM-selected relevance (not grep).
- `MemdirStore.MarkStale(entry)` — freshness caveat wrapping.
- `OutputStyleLoader.Load(dir) []OutputStyle` — CC-compatible frontmatter.
- `MigrationRunner.Apply(current int) error` — preAction auto-apply at session-start.

**Plugin system (X-4): deferred to v4.** Rationale: moai's extensibility via skills + agents is already sufficient; a second plugin layer would confuse the surface. Revisit if v4 adds a marketplace. *Source: pattern-library.md X-4.*

---

## 5. Cross-Layer Concerns

### 5.1 Typed Memory Taxonomy (pattern M-1)

Four memory kinds with distinct lifetimes, each stored in `~/.claude/projects/{hash}/memory/`:

| Type | Lifetime | Example |
|------|----------|---------|
| `user` | Persistent, rarely changes | "GOOS행님; local dev on macOS; avoid time estimates in reports" |
| `feedback` | Persistent, correction patterns | "Team tmux pane cleanup — kill-pane before TeamDelete" |
| `project` | Per-project, current work state | "Active SPEC-V3R2-RT-003; Phase 2; sandbox layer incomplete" |
| `reference` | Persistent, external pointers | "design-constitution v3.3.0 §11.4 FROZEN" |

Retrieval uses LLM-selected relevance (M-5) with a dedicated Sonnet sideQuery; stale entries (>24h) wrapped in `<system-reminder>` with explicit caveat to prevent over-trust. Cross-cuts: Layer 1 (constitution constant pointers), Layer 4 (per-agent memory directories), Layer 7 (memdir loader). *Source: pattern-library.md M-1, M-5; r3-cc-architecture-reread.md §1.1 memdir.*

### 5.2 Multi-Source Permission Resolution (pattern S-1)

Permissions resolve via an 8-source ordered stack (priority high → low):

```
 policy > user > project > local > plugin > skill > session > builtin
```

Every rule carries a `Source` tag for `/moai doctor` provenance. Resolution returns `allow | ask | deny + updatedInput?`. Provenance enables: "which file set this rule?" diagnostics, targeted migrations (touch userSettings without disturbing projectSettings), and plugin-hook dedup. Cross-cuts: Layer 3 (runtime enforcement), Layer 4 (agent `permissionMode`), Layer 7 (plugin-contributed rules). *Source: r3-cc-architecture-reread.md §1.3, §4 Adopt 2; pattern-library.md S-1.*

### 5.3 ACI — Agent-Computer Interface (pattern T-1)

Six canonical ACI commands form the preferred tool layer for moai agents. Raw `Bash` remains as escape hatch but is discouraged for common moves. Each command returns structured, paginated, LM-optimized responses with a linter guardrail at write time.

```
 moai_spec_read          — load SPEC + acceptance criteria, truncated by token budget
 moai_locate_mx_anchor   — find @MX:ANCHOR by ID, return callsite fan-in
 moai_run_tests_for_spec — run acceptance tests with fail-to-pass classification
 moai_lsp_find_references — LSP-backed via powernap, 16-language neutral
 moai_lsp_workspace_symbols — symbol search across project
 moai_linter_gated_write — block syntactically-invalid edits before commit
```

Cross-cuts: Layer 2 (SPEC read, MX resolution), Layer 3 (hook PostToolUse validates linter), Layer 6 (workflow uses). *Source: r1-ai-harness-papers.md §11 SWE-agent; pattern-library.md T-1; r2-opensource-tools.md §8.*

### 5.4 Hook JSON-OR-ExitCode Dual Protocol (pattern T-5)

Hook handlers emit structured JSON on stdout when they have rich output; empty JSON or missing output falls back to exit code parsing. Fields:

```
 additionalContext   — inject text into next model turn
 permissionDecision  — allow | ask | deny
 updatedInput        — mutate the tool input mid-turn
 systemMessage       — user-visible notification
 continue            — false blocks teammate from idling
```

Migration: shell wrappers unchanged (they forward stdin/stdout); Go handlers in `internal/hook/*.go` gain typed `HookResponse` return. v3 upgrades 5 critical handlers first (subagent-stop tmux fix, config-change reload, setup, instructions-loaded CLAUDE.md length check, file-changed MX re-scan), then broader migration. Cross-cuts: Layer 3 (protocol definition), Layer 4 (agent-scoped hooks), Layer 5 (Sprint Contract injection via PostToolUse). *Source: r3-cc-architecture-reread.md §2 Dec 5; pattern-library.md T-5.*

### 5.5 Multi-Layer Settings with Provenance (pattern X-2)

Every configuration value that enters the merged runtime representation carries a `Source` tag. The merge is deterministic (ordered by the 8-source priority). Provenance is exposed via `moai doctor config dump` and via permission-denial diagnostics. This is the prerequisite for S-1 (permission stack), S-2 (bubble mode), and sandbox routing by source. Cross-cuts: Layer 1 (config is part of constitutional runtime surface), Layer 3 (runtime config), Layer 7 (plugin-contributed settings). *Source: r3-cc-architecture-reread.md §2 Dec 11, §4 Adopt 1; pattern-library.md X-2.*

### 5.6 File-First State + Fresh-Context Iteration (Principle 3 + 12)

Cross-iteration state is a file on disk; LM context that consumes the state is ephemeral. Canonical layout at `.moai/state/`:

```
 .moai/state/
   task-ledger.md      # append-only, Magentic pattern (O-3)
   progress.md         # Ralph-shape per iteration
   activity.log
   errors.log
   runs/{iter-id}/
     prompt.md
     response.md
     artifacts/
   sprint-contract.yaml
   checkpoint-{phase}.json
```

Cross-iteration reset: each Ralph iteration starts fresh; the prompt rebuilds from the SPEC + state files, not from accumulated transcripts. `STALE_SECONDS` primitive (from Ralph) governs crash-resume. Cross-cuts: Layer 3 (session state), Layer 5 (Sprint Contract state), Layer 6 (loop mode execution). *Source: pattern-library.md R-6; design-principles.md P3, P12.*

### 5.7 Agent-as-Judge Without Memory (pattern E-1 + Principle 4)

**This is the single most important amendment to moai's design subsystem.** Current design-constitution §11.4 (Sprint Contract) retains evaluator memory across iterations; Agent-as-a-Judge paper (R1 §9) explicitly flags this as an anti-pattern: "any errors in previous judgments could lead to a chain of errors."

v3 amends the design constitution:

- **Sprint Contract state is durable** — passed criteria carry forward (no regression allowed); failed criteria get refined based on evaluator feedback; new criteria may be added if previous sprint revealed gaps.
- **Evaluator judgment memory is ephemeral** — each iteration spawns evaluator-active with a fresh context that sees only: (a) the BRIEF / SPEC, (b) the Sprint Contract state, (c) the artifact to evaluate. It MUST NOT see prior scoring rationale.
- Constitutional amendment: `evaluator.memory_scope: per_iteration` added to `.moai/config/sections/design.yaml` and harness.yaml.

Cross-cuts: Layer 5 (evaluator-active flow), Layer 4 (evaluator agent frontmatter). *Source: r1-ai-harness-papers.md §9 anti-pattern flag; design-principles.md P4; problem-catalog.md P-Z01.*

---

## 6. Problem Resolution Matrix

72 problems from `synthesis/problem-catalog.md`. Severity: C=Critical, H=High, M=Medium, L=Low. Resolution type: Fix (direct fix), Redesign (architectural change), Retire (removed), Defer (v3.1+).

| Problem | Sev | Addressed by SPEC | Layer | Type |
|---------|-----|-------------------|-------|------|
| P-A01 AskUserQuestion in 9 agents | C | SPEC-V3R2-ORC-001, -ORC-002 | 4 | Fix |
| P-A02 19 agents missing `effort` | H | SPEC-V3R2-ORC-003 | 4 | Fix |
| P-A03 3 wrong `effort` | H | SPEC-V3R2-ORC-003 | 4 | Fix |
| P-A04 4 agents dead `Agent` tool | H | SPEC-V3R2-ORC-002 | 4 | Fix |
| P-A05 3 builders near-identical | H | SPEC-V3R2-ORC-001 | 4 | Redesign (merge) |
| P-A06 expert-debug router | H | SPEC-V3R2-ORC-001 | 4 | Retire (→ manager-quality sub-mode) |
| P-A07 expert-testing strategy-only | H | SPEC-V3R2-ORC-001 | 4 | Retire (→ manager-cycle) |
| P-A08 expert-performance advisor | H | SPEC-V3R2-ORC-001 | 4 | Redesign (grant Write or retire) |
| P-A09 manager-ddd/tdd 60% overlap | H | SPEC-V3R2-ORC-001 | 4 | Redesign (→ manager-cycle) |
| P-A10 manager-project 6 modes | H | SPEC-V3R2-ORC-001 | 4 | Redesign (scope shrink) |
| P-A11 6 agents missing worktree | H | SPEC-V3R2-ORC-004 | 4 | Fix |
| P-A12 manager-project scope overreach | M | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-A13 duplicate Skeptical Mandate | M | SPEC-V3R2-ORC-002 | 4 | Fix (extract to common) |
| P-A14 plan-auditor no memory field | M | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-A15 Context7 over-included | M | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-A16 manager-git 14 triggers | L | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-A17 expert-backend 24 triggers w/ dup | L | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-A18 dead hook config | M | SPEC-V3R2-ORC-002 | 4 | Fix |
| P-A19 --deepthink boilerplate × 22 | L | SPEC-V3R2-ORC-002 | 4 | Fix |
| P-A20 expert-frontend mixed scope | M | SPEC-V3R2-ORC-001 | 4 | Redesign (split Pencil) |
| P-A21 missing manager-design | M | SPEC-V3R2-ORC-001 | 4 | Defer (evaluate after WF-003) |
| P-A22 researcher no worktree | H | SPEC-V3R2-ORC-004 | 4 | Fix |
| P-A23 skills injection parity | L | SPEC-V3R2-ORC-001 | 4 | Fix |
| P-S01 Thinking triplet overlap | H | SPEC-V3R2-WF-001 | 6 | Redesign (merge) |
| P-S02 Platform triplets | H | SPEC-V3R2-WF-001 | 6 | Redesign (split/narrow) |
| P-S03 Kitchen-sink domain skills | H | SPEC-V3R2-WF-001 | 6 | Redesign (router pattern) |
| P-S04 `moai` root skill 300KB | H | SPEC-V3R2-WF-001 | 6 | Redesign (promote workflows to skills) |
| P-S05 4 skills over Level 2 budget | M | SPEC-V3R2-WF-001 | 6 | Fix |
| P-S06 43 bundled files in testing | M | SPEC-V3R2-WF-001 | 6 | Fix |
| P-S07 reference/ vs references/ | L | SPEC-V3R2-WF-001 | 6 | Fix (lint) |
| P-S08 moai-lang-* referenced but absent | H | SPEC-V3R2-WF-005 | 6 | Redesign (boundary decision) |
| P-S09 moai-ref-* zero static refs | M | SPEC-V3R2-WF-001 | 6 | Fix (document activation) |
| P-S10 moai-tool-svg niche | M | SPEC-V3R2-WF-001 | 6 | Retire |
| P-S11 moai-docs-generation superseded | M | SPEC-V3R2-WF-001 | 6 | Retire |
| P-S12 moai-design-tools stapled | M | SPEC-V3R2-WF-001 | 6 | Redesign (split) |
| P-S13 Design cluster 4-way overlap | M | SPEC-V3R2-WF-001 | 6 | Redesign (merge) |
| P-S14 moai-workflow-templates nested aggregate | M | SPEC-V3R2-WF-001 | 6 | Retire (→ workflow-project) |
| P-S15 PD declaration missing in 40% | L | SPEC-V3R2-WF-001 | 6 | Fix |
| P-S16 moai-foundation-context zero refs | L | SPEC-V3R2-WF-001 | 6 | Fix (merge into core) |
| P-S17 Level 2 budget inconsistent | L | SPEC-V3R2-WF-001 | 6 | Fix (document ladder) |
| P-S18 Stale sibling refs | L | SPEC-V3R2-WF-001 | 6 | Fix |
| P-S19 lang rules vs skills boundary | H | SPEC-V3R2-WF-005 | 6 | Redesign |
| P-H01 10 logging-only handlers | H | SPEC-V3R2-RT-006 | 3 | Redesign (upgrade or retire each) |
| P-H02 subagentStop no tmux kill (BUG) | C | SPEC-V3R2-RT-006 | 3 | Fix |
| P-H03 setupHandler orphan | H | SPEC-V3R2-RT-006 | 3 | Fix (implement or remove) |
| P-H04 hardcoded /Users/goos path | C | SPEC-V3R2-RT-007 | 3 | Fix |
| P-H05 exit-code-only hook protocol | H | SPEC-V3R2-RT-001 | 3 | Redesign |
| P-H06 5 yaml sections no Go loader | C | SPEC-V3R2-MIG-003 | 3,5 | Fix |
| P-H07 sunset.yaml dormant | M | SPEC-V3R2-MIG-003 | 7 | Fix (activate or retire) |
| P-H08 /98-github.md 698 LOC | H | SPEC-V3R2-WF-002 | 6 | Redesign (extract skill) |
| P-H09 /99-release.md 890 LOC | H | SPEC-V3R2-WF-002 | 6 | Redesign (extract skill) |
| P-H10 lsp-client.md misfiled | L | SPEC-V3R2-CON-003 | 1 | Fix (move) |
| P-H11 workflow-modes/spec-workflow overlap | M | SPEC-V3R2-CON-003 | 1 | Fix (merge) |
| P-H12 team-protocol/worktree-integration overlap | M | SPEC-V3R2-CON-003 | 1 | Fix (merge) |
| P-H13 file-reading-optimization is heuristic | L | SPEC-V3R2-CON-003 | 1 | Fix (move to skill) |
| P-H14 frontmatter inconsistency | L | SPEC-V3R2-CON-003 | 1 | Fix (migrate to paths:) |
| P-H15 configChange reload no-op | M | SPEC-V3R2-RT-006 | 3 | Fix |
| P-H16 instructionsLoaded validation no-op | L | SPEC-V3R2-RT-006 | 3 | Fix |
| P-H17 fileChanged MX re-scan no-op | L | SPEC-V3R2-RT-006 | 3 | Fix |
| P-H18 design.md/db.md extension drift | L | SPEC-V3R2-WF-002 | 6 | Fix |
| P-H19 59% events partial coverage | H | SPEC-V3R2-RT-006 | 3 | Redesign |
| P-H20 workflow.yaml partial schema | M | SPEC-V3R2-MIG-003 | 3 | Fix |
| P-R01 handler count drift | M | SPEC-V3R2-MIG-002 | 3 | Fix |
| P-R02 Constitutional sprawl | M | SPEC-V3R2-CON-003 | 1 | Redesign |
| P-R03 CLAUDE.md/common-protocol dup | M | SPEC-V3R2-CON-003 | 1 | Fix |
| P-C01 No permission bubble | C | SPEC-V3R2-RT-002 | 3 | Redesign |
| P-C02 No sub-agent context isolation | H | SPEC-V3R2-RT-004 | 3 | Redesign |
| P-C03 No sandbox default | C | SPEC-V3R2-RT-003 | 3 | Redesign |
| P-C04 No config provenance | H | SPEC-V3R2-RT-005 | 3 | Redesign |
| P-C05 No cache-prefix discipline | M | SPEC-V3R2-RT-004 | 3 | Fix (audit prompt assembly) |
| P-C06 Explicit migrate command | L | SPEC-V3R2-EXT-004 | 7 | Redesign (auto-apply) |
| P-Z01 Evaluator memory cascade | H | SPEC-V3R2-HRN-002 | 5 | Redesign (constitution amendment) |
| P-Z02 Utility multi-agent over-use | M | SPEC-V3R2-WF-004 | 6 | Redesign (classify as Agentless) |
| P-X01 /98-/99- template drift | L | SPEC-V3R2-WF-002 | 6 | Fix |

### Problems intentionally deferred to v3.1+

- **M-4 Workflow Memory Induction (AWM)** — auto-distillation from successful `/moai run` trajectories. Premature; needs ≥100 completed SPECs + telemetry corpus.
- **T-2 Deeper Tree-sitter Repo-Map** — `/moai codemaps` already provides partial coverage; deeper integration is optimization, not critical.
- **O-2 Pub-Sub Shared Message Pool** — current team sizes don't warrant; direct `SendMessage` works. Revisit past ~10 teammates.
- **X-4 Plugin Marketplace** — skills + agents already adequate; second plugin axis would confuse surface.
- **E-4 Pass@N default** — N× cost; thorough-harness opt-in only.
- **R-4 Tree-of-Thoughts default** — 5-100× token cost; thorough-harness opt-in only.
- **P-A21 manager-design agent** — evaluate after WF-003 multi-mode router ships and /moai design flow stabilizes.

*Source: pattern-library.md §Patterns deliberately NOT adopted.*

---

## 7. Component Inventory (Target State)

### 7.1 Skills — target ~24 (from 48)

| Skill | Verdict | Notes |
|-------|---------|-------|
| moai (root) | REFACTOR | Thin router; 20 bundled workflows promote to `moai:*` system skills |
| moai-foundation-core | KEEP | TRUST 5, SPEC-DDD authoritative |
| moai-foundation-cc | KEEP | CC authoring kit |
| moai-foundation-quality | KEEP | TRUST 5 glue |
| moai-foundation-thinking | MERGE | absorbs philosopher + workflow-thinking |
| moai-foundation-context | RETIRE | fold into foundation-core |
| moai-workflow-spec | KEEP | EARS authority |
| moai-workflow-tdd | KEEP | RED-GREEN-REFACTOR canonical |
| moai-workflow-ddd | KEEP | ANALYZE-PRESERVE-IMPROVE canonical |
| moai-workflow-testing | REFACTOR | slim 22.5KB → ~12KB, split bundled 43 files |
| moai-workflow-project | KEEP | absorbs templates + docs-generation |
| moai-workflow-worktree | KEEP | worktree mechanics |
| moai-workflow-loop | KEEP | Ralph Engine (LSP + ast-grep) |
| moai-workflow-gan-loop | KEEP | Builder-Evaluator contract (amend §11.4 per P-Z01) |
| moai-workflow-design-context | KEEP | brief loader |
| moai-workflow-design-import | KEEP | handoff bundle parser |
| moai-workflow-research | KEEP | experimental loop |
| moai-workflow-jit-docs | KEEP | JIT docs loader |
| moai-domain-backend | REFACTOR | decision matrix, not kitchen sink |
| moai-domain-frontend | REFACTOR | router to ref-react / library-nextra |
| moai-domain-database | MERGE | absorbs platform-database-cloud |
| moai-domain-db-docs | KEEP | migration parser |
| moai-domain-copywriting | KEEP (FROZEN) | agency absorption contract |
| moai-domain-brand-design | KEEP (FROZEN) | agency absorption contract |
| moai-design-system | NEW | merge design-craft + domain-uiux |
| moai-tool-ast-grep | KEEP | canonical structural search |
| moai-tool-figma | NEW | split from moai-design-tools |
| moai-tool-svg | RETIRE | niche, zero refs |
| moai-library-mermaid | KEEP | |
| moai-library-nextra | KEEP | |
| moai-library-shadcn | KEEP | |
| moai-ref-api-patterns | KEEP | agent-extending reference |
| moai-ref-git-workflow | KEEP | |
| moai-ref-owasp-checklist | KEEP | |
| moai-ref-react-patterns | KEEP | |
| moai-ref-testing-pyramid | KEEP | |
| moai-docs-generation | RETIRE | superseded by jit-docs + nextra |
| moai-design-craft | RETIRE | merged into design-system |
| moai-design-tools | RETIRE | split to tool-figma + pencil-integration |
| moai-workflow-templates | RETIRE | merged into workflow-project |
| moai-workflow-thinking | RETIRE | merged into foundation-thinking |
| moai-workflow-pencil-integration | KEEP (absorbs Pencil side of design-tools) | |
| moai-domain-uiux | RETIRE | merged into design-system |
| moai-foundation-philosopher | RETIRE | merged into foundation-thinking |
| moai-platform-auth | REFACTOR | vendor-narrow contract |
| moai-platform-deployment | REFACTOR | vendor-narrow contract |
| moai-platform-database-cloud | RETIRE | merged into domain-database |
| moai-platform-chrome-extension | KEEP (monitor) | niche but well-scoped |
| moai-framework-electron | KEEP (monitor) | niche, retire if unused |
| moai-formats-data | KEEP (monitor) | TOON + JSON optimization |

*Source: r4-skill-audit.md §Recommended v3 skill inventory.*

### 7.2 Agents — target 17 (from 22)

| Agent | Category | Verdict | Notes |
|-------|----------|---------|-------|
| manager-spec | manager | KEEP | scrub AskUserQuestion per P-A01 |
| manager-strategy | manager | KEEP | scrub AskUserQuestion |
| manager-cycle | manager | NEW | merges manager-ddd + manager-tdd with `cycle_type:` |
| manager-quality | manager | KEEP | absorbs expert-debug diagnostic sub-mode |
| manager-docs | manager | KEEP | |
| manager-git | manager | KEEP | trim triggers per P-A16 |
| manager-project | manager | REFACTOR | scope to `.moai/project/` only; other modes → CLI |
| expert-backend | expert | KEEP | add `isolation: worktree`; drop duplicate Oracle trigger |
| expert-frontend | expert | KEEP | add `isolation: worktree`; consider split of Pencil scope |
| expert-security | expert | KEEP | `effort: xhigh`; drop dead `Agent` tool |
| expert-devops | expert | KEEP | scrub AskUserQuestion |
| expert-performance | expert | REFACTOR | grant `Write` scoped to `.moai/docs/` or retire |
| expert-refactoring | expert | KEEP | `effort: xhigh`; add `isolation: worktree`; document boundary vs manager-cycle IMPROVE |
| builder-platform | builder | NEW | merges builder-agent + builder-skill + builder-plugin |
| evaluator-active | evaluator | KEEP | `effort: xhigh`; evaluator memory per-iteration (P-Z01) |
| plan-auditor | evaluator | KEEP | `effort: xhigh`; add `memory: project` |
| researcher | meta | REFACTOR | `effort: xhigh`; `isolation: worktree`; or retire to skill runbook |
| manager-ddd | — | RETIRED (→ manager-cycle) | |
| manager-tdd | — | RETIRED (→ manager-cycle) | |
| expert-debug | — | RETIRED (→ manager-quality sub-mode) | |
| expert-testing | — | RETIRED (→ manager-cycle strategy + expert-performance load) | |
| builder-agent | — | RETIRED (→ builder-platform) | |
| builder-skill | — | RETIRED (→ builder-platform) | |
| builder-plugin | — | RETIRED (→ builder-platform) | |

*Source: r5-agent-audit.md §Recommended v3 agent inventory.*

### 7.3 Hooks — 25 native events (business-logic audit)

| Event | Current Logic | v3 Target | Resolution |
|-------|--------------|-----------|------------|
| SessionStart | Full (GLM, skill, memory) | Full | KEEP |
| SessionEnd | Full (memo, MX) | Full | KEEP |
| PreToolUse | Full (security) | Full + JSON injection | UPGRADE |
| PostToolUse | Full (MX validate) | Full + MX tag injection via JSON | UPGRADE |
| PostToolUseFailure | Partial | Full (error classification) | UPGRADE |
| PreCompact | Full (memo save) | Full | KEEP |
| PostCompact | Full (memo restore) | Full | KEEP |
| Stop | Full | Full | KEEP |
| StopFailure | Full | Full | KEEP |
| SubagentStart | Full | Full | KEEP |
| SubagentStop | **Partial (no-op on tmux bug)** | Full (kill tmux pane) | **FIX P-H02** |
| Notification | Partial | Retire from settings.json | RETIRE |
| UserPromptSubmit | Full | Full | KEEP |
| PermissionRequest | Full | Full | KEEP |
| PermissionDenied | Full | Full | KEEP |
| TeammateIdle | Full | Full | KEEP |
| TaskCompleted | Full | Full | KEEP |
| TaskCreated | Partial | Retire from settings.json | RETIRE |
| WorktreeCreate | Full | Full | KEEP |
| WorktreeRemove | Full | Full | KEEP |
| ConfigChange | Partial | Full (reload + revalidate) | UPGRADE |
| CwdChanged | Full | Full | KEEP |
| FileChanged | Partial | Full (MX re-scan) | UPGRADE |
| InstructionsLoaded | Partial | Full (CLAUDE.md length check) | UPGRADE |
| Elicitation | Partial | Retire | RETIRE |
| ElicitationResult | Partial | Retire | RETIRE |
| Setup (special) | Orphan | Implement or remove Go handler | FIX P-H03 |

*Source: r6-commands-hooks-style-rules.md §A Hook Coverage Matrix.*

### 7.4 Commands — target 15 thin + 2 refactored

| Command | LOC | Verdict | Route |
|---------|-----|---------|-------|
| /moai plan|run|sync|project|design|db|fix|loop|clean|mx|feedback|review|coverage|e2e|codemaps (×15) | 8 each | KEEP | `Skill("moai")` |
| /98-github.md | 698 | REFACTOR | extract to `moai-workflow-github` skill |
| /99-release.md | 890 | REFACTOR | extract to `moai-workflow-release` skill |

Template extension drift (design.md, db.md use `.md` instead of `.md.tmpl`) fixed under WF-002. *Source: r6-commands-hooks-style-rules.md §1.*

### 7.5 Rules — target 31 files (from 34)

| Action | Count | Notes |
|--------|-------|-------|
| KEEP | 28 | core/, design/, development/, languages/ (all 16), most workflow/ |
| MOVE | 1 | `core/lsp-client.md` → `.moai/decisions/` (SPEC decision record, not agent rule) |
| MERGE | 2 | `workflow/workflow-modes.md` → `workflow/spec-workflow.md`; `workflow/team-protocol.md` → `workflow/worktree-integration.md` |
| MOVE | 1 | `workflow/file-reading-optimization.md` → `moai-foundation-context` skill references |
| FRONTMATTER MIGRATION | 4 | `moai-constitution.md`, `coding-standards.md`, `team-protocol.md`, `worktree-integration.md` migrate `description + globs` → `paths:` CSV |

*Source: r6-commands-hooks-style-rules.md §4.5.*

### 7.6 Config sections — target 23 active (all Go-loaded)

| Section | Go loader? | Verdict |
|---------|-----------|---------|
| language, llm, quality, workflow, lsp, mx, security, statusline, system, user, project, git-convention, git-strategy, ralph, research, state | YES | KEEP |
| **harness.yaml** | **NO (currently template-only)** | **ADD LOADER (P-H06 CRITICAL)** |
| **constitution.yaml** | NO | ADD LOADER |
| **context.yaml** | NO | ADD LOADER |
| **interview.yaml** | NO | ADD LOADER |
| **design.yaml** | Partial (migrate_agency only) | ADD RUNTIME LOADER |
| sunset.yaml | Struct exists but dormant | ACTIVATE or RETIRE |

Every v3 yaml section requires a Go struct in `internal/config/types.go`, a loader in `internal/config/loader.go`, and a test in `loader_test.go` — new CI rule. *Source: r6-commands-hooks-style-rules.md §5.2.*

---

## 8. Breaking Changes Catalog

Each BC identifies what breaks, migration automation level, and the deprecation window. The `v2→v3 migrator` (SPEC-V3R2-MIG-001) handles all **AUTO** migrations; **MANUAL** migrations require user action with documentation.

| ID | Breaking Change | Migration | Deprecation |
|----|-----------------|-----------|-------------|
| BC-V3R2-001 | Hook handlers migrate to JSON-OR-ExitCode protocol | AUTO (wrappers unchanged; handlers rewritten) | v2.x hooks continue via exit code fallback |
| BC-V3R2-002 | Agent frontmatter requires `effort` field | AUTO (migrator populates per matrix; default medium) | v2.x agents missing field fall through to session default |
| BC-V3R2-003 | Sandbox-by-default for implementer agents | AUTO (frontmatter adds `sandbox: seatbelt` on macOS, `bubblewrap` on Linux) | Opt-out via `sandbox: none` with SPEC-documented justification |
| BC-V3R2-004 | 4 dead `Agent` tool declarations removed from subagent definitions | AUTO (migrator scrubs) | N/A (dead config) |
| BC-V3R2-005 | 9 agent bodies lose `AskUserQuestion` lines; replaced by blocker-report pattern | AUTO (migrator rewrites) + CI lint | CI lint fails v2-style bodies after v3.0.0-alpha.3 |
| BC-V3R2-006 | 48 skills → 24 via merge clusters | AUTO (migrator rewrites `related-skills`, `skills:` fields) | Old skill names return a stub with "merged into X" pointer for one v3.x cycle |
| BC-V3R2-007 | `/moai fix | coverage | mx | codemaps | clean` become Agentless fixed pipelines (no subagents by default) | AUTO (flag flip) | Opt-in `--mode agent` preserves v2 behavior during v3.x |
| BC-V3R2-008 | Hardcoded `/Users/goos/go/bin/moai` fallback removed from 26 shell wrappers | AUTO (`make build` regenerates via updated GoBinPath resolver) | N/A (bug fix) |
| BC-V3R2-009 | 22 → 17 agents: manager-ddd + tdd → manager-cycle; 3 builders → builder-platform; expert-debug + testing retired | AUTO (migrator rewrites SPEC agent references; stub agents map to new) | Stubs removed after v3.1.0 |
| BC-V3R2-010 | Evaluator memory scope per-iteration (design-constitution §11.4 amendment) | AUTO (config flag added; evaluator-active respawn per iteration) | Old evaluator sessions (memory-retentive) retired on first upgrade |
| BC-V3R2-011 | SPEC acceptance criteria become hierarchical (nested Given/When/Then); flat criteria promoted to 1-level tree | AUTO (migrator wraps flat criteria as single-level children) | Old flat SPECs remain parseable indefinitely |
| BC-V3R2-012 | `/98-github.md` + `/99-release.md` extracted to `moai-workflow-github` + `moai-workflow-release` skills | AUTO (for dev tree) | Dev-local only; no user impact |
| BC-V3R2-013 | Config sections gain Go loaders (constitution, context, interview, design, harness) | AUTO (loaders added; existing YAML files unchanged) | Template-only era deprecated at v3.0.0-alpha.1 |
| BC-V3R2-014 | `lsp-client.md` rule moved to `.moai/decisions/lsp-client-choice.md` | AUTO (move + update references) | Rule-tree grep returns empty; decisions/ grep finds new location |
| BC-V3R2-015 | Multi-layer settings resolution with `Source` tags replaces flat merge | AUTO (reader layer; settings.json files unchanged) | Flat-merge consumers are internal; no user API change |
| BC-V3R2-016 | `manager-ddd` + `manager-tdd` agent names deprecated in SPEC references | AUTO (migrator rewrites to `manager-cycle` with `cycle_type:` argument) | Stubs for one v3.x cycle |
| BC-V3R2-017 | Workflow-modes + team-protocol rules merged into spec-workflow + worktree-integration | AUTO (file move + ref update) | Old rule paths return 404 after v3.0.0-alpha.1 |
| BC-V3R2-018 | `notification`, `elicitation`, `elicitationResult`, `taskCreated` hook events removed from settings.json | AUTO | Events still fire at CC level but no moai handler; Go handlers retained as observability tap with explicit disable option |
| BC-V3R2-019 | Migration framework auto-runs on `init`/`update`/`doctor`/`migrate` entry points (idempotent, ordered, rollback-aware) | AUTO (framework ships with v3.0.0-alpha.1) | v2.x users hit migration on first v3 invocation; idempotent re-runs are no-ops |

Top 5 for reviewer summary: BC-V3R2-001 (hook JSON), BC-V3R2-003 (sandbox default), BC-V3R2-005 (AskUserQuestion scrub), BC-V3R2-006 (skill 48→24), BC-V3R2-010 (evaluator memory amendment).

---

## 9. Release Plan

| Phase | Tag | Scope | Primary SPECs |
|-------|-----|-------|---------------|
| 1 Constitution & Foundation | v3.0.0-alpha.1 | FROZEN/EVOLVABLE codification; settings layering with provenance; migration framework auto-apply | SPEC-V3R2-CON-001, -CON-002, -CON-003, SPEC-V3R2-RT-005, SPEC-V3R2-EXT-004 |
| 2 Runtime Hardening | v3.0.0-alpha.2 | Hook JSON protocol + 27-event coverage + sandbox layer + hardcoded-path fix | SPEC-V3R2-RT-001, -RT-002, -RT-003, -RT-004, -RT-006, -RT-007 |
| 3 Agent Cleanup | v3.0.0-alpha.3 | 22→17 agents, Common-Protocol scrub, effort calibration, worktree MUST | SPEC-V3R2-ORC-001, -ORC-002, -ORC-003, -ORC-004 |
| 4 Skill Consolidation | v3.0.0-alpha.4 | 48→24 skills; merge clusters 1-5; retirements | SPEC-V3R2-WF-001 |
| 5 Harness + Evaluator | v3.0.0-beta.1 | Sprint Contract fresh-judgment fix, harness routing, hierarchical acceptance | SPEC-V3R2-HRN-001, -HRN-002, -HRN-003, SPEC-V3R2-SPC-001 |
| 6 Multi-Mode Workflow | v3.0.0-beta.2 | `/moai loop` Ralph, `/moai design` path B, `--mode team`, Agentless classification, dynamic team generation, language rules/skills boundary | SPEC-V3R2-WF-002, -WF-003, -WF-004, -WF-005, -WF-006, SPEC-V3R2-ORC-005 |
| 7 Extension | v3.0.0-rc.1 | ACI commands, output styles CC-alignment, memdir typed taxonomy | SPEC-V3R2-SPC-002, -SPC-003, -SPC-004, SPEC-V3R2-EXT-001, -EXT-002 |
| 8 Migration Tool + Docs | v3.0.0-rc.2 | v2→v3 migrator, release docs, doctor diagnostics, MIG-002/003 closeout | SPEC-V3R2-MIG-001, -MIG-002, -MIG-003 |
| 9 Stable | v3.0.0 | GA — no new SPECs; bug fixes + docs | — |

Phase ordering is dependency-driven: Constitution (1) enables all others; Runtime (2) enables Agent Cleanup (3) which enables Skill (4) which enables Harness (5); Multi-Mode (6) builds on 3 and 4; Extension (7) polish; Migration (8) unblocks user upgrade; GA (9). No phase assumes a successor; each is releasable alpha/beta.

---

## 10. Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| R1. Skill merge breaks agent keyword auto-activation | HIGH | MEDIUM | Keep union of trigger keywords at merge; CI regression tests; alpha.4 gating |
| R2. Hook JSON migration breaks 26 shell wrappers | MEDIUM | MEDIUM | Wrappers unchanged; Go emits JSON stdout; exit-code fallback preserved |
| R3. Sandbox per-OS divergence (bwrap vs seatbelt vs docker) | HIGH | MEDIUM | Per-OS backends with feature detection; `sandbox: none` fallback; CI matrix across 3 OS |
| R4. Design-constitution §11.4 amendment triggers FROZEN-zone gate | MEDIUM | LOW | Use graduation protocol; canary-eval on last 3 design projects; human approval via AskUserQuestion |
| R5. 22→17 agent reduction breaks SPEC workflows with hardcoded agent names | MEDIUM | HIGH | `moai migrate` rewrites SPEC agent refs; stub agents for one v3.x cycle per BC-V3R2-009, -016 |
| R6. Adding 5 config loaders introduces schema churn | LOW | MEDIUM | Strict versioning per yaml; migration v1→v2 schema handled by MIG-003 |
| R7. Permission bubble mode UX fatigue | HIGH | MEDIUM | Pre-allowlist for common dev ops; bubble only for genuinely novel risks; telemetry post-beta.1 |
| R8. Fresh-context per iteration loses within-task reasoning | MEDIUM | LOW | @MX tags + SPEC + `.moai/state/` as persistent substrate; explicit `progress.md` append-only log |
| R9. Agentless classification regresses users expecting agent delegation | MEDIUM | LOW | Opt-in `--mode agent` preserves v2 behavior; empirical measurement in alpha.4 telemetry |
| R10. Go binary absent hot-reload — changes require rebuild | HIGH | LOW | Document in CLAUDE.local.md; `make build && make install` + Claude Code restart standard; `moai doctor` Binary Freshness check |
| R11. Claude Code version drift past 2.1.111 target | MEDIUM | MEDIUM | Version-pin check in `moai doctor`; migration path in CHANGELOG; max 2 minor-version window |
| R12. AskUserQuestion scrub + blocker-report pattern subtle behavior change | LOW | HIGH | CI lint rejects literal `AskUserQuestion`; comprehensive e2e tests at alpha.3 |
| R13. 16-language neutrality strained if ACI commands favor specific LSP | LOW | MEDIUM | Language-agnostic contract via powernap abstraction; per-language adapters tested in CI |
| R14. Template vs local drift during multi-phase v3 migration | MEDIUM | MEDIUM | `diff -rq` CI check already present per CLAUDE.local.md §2; hardened at each alpha |
| R15. v3 redesign scope bloat (v3-legacy already bloated) | HIGH | MEDIUM | Explicit NOT-NOW list (§6 deferred); phase gating; each alpha releasable; no Phase 9 "catch-up" pattern allowed |

---

## 11. SPEC Index (Wave 4 input)

35 target SPECs grouped by layer. All prefix `SPEC-V3R2-` (round 2) to distinguish from v3-legacy. Each cites related principles (P#), problems (P-XYZ), and patterns (letter-number).

### 11.1 Constitution (Layer 1) — 3 SPECs

- **SPEC-V3R2-CON-001: FROZEN/EVOLVABLE zone codification** — generalize design-constitution zone model to core constitution; list the 7 FROZEN invariants. *Principles P1, P2, P12. Pattern S-4.*
- **SPEC-V3R2-CON-002: Constitutional amendment protocol** — 5-layer safety gate (FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight); evolution-log format. *Pattern S-5.*
- **SPEC-V3R2-CON-003: Constitution consolidation pass** — rule tree cleanup (P-H10..H14, P-R02, P-R03); move lsp-client.md; merge workflow-modes/spec-workflow, team-protocol/worktree-integration; frontmatter `paths:` migration.

### 11.2 SPEC & TAG (Layer 2) — 4 SPECs

- **SPEC-V3R2-SPC-001: EARS + hierarchical acceptance criteria** — Agent-as-a-Judge hierarchical shape (R1 §9 365-sub-req pattern); migration from flat Given/When/Then. *Principle 1. Pattern E-1.*
- **SPEC-V3R2-SPC-002: @MX TAG protocol v2 with hook JSON injection** — autonomous add/update/remove via PostToolUse JSON `additionalContext`; cross-language TagScanner; sub-line parsers for NOTE/WARN/ANCHOR/TODO/LEGACY. *Principle 1, 8.*
- **SPEC-V3R2-SPC-003: SPEC linter** — `moai spec lint` enforces EARS modality, Given/When/Then testability, orphan acceptance detection; CI integration.
- **SPEC-V3R2-SPC-004: SPEC-to-MX-anchor resolver** — `moai_locate_mx_anchor` ACI command; fan-in analysis for ANCHOR promotion. *Pattern T-1.*

### 11.3 Runtime (Layer 3) — 7 SPECs

- **SPEC-V3R2-RT-001: Hook JSON-OR-ExitCode protocol** — HookResponse schema, 27-event compliance, 5 critical handler upgrades first. *Principle 8. Pattern T-5.*
- **SPEC-V3R2-RT-002: Permission stack with bubble mode** — 8-source resolution, `bubble` as first-class PermissionMode, hook PreToolUse `permissionDecision` wiring. *Principle 7. Patterns S-1, S-2.*
- **SPEC-V3R2-RT-003: Sandbox execution layer** — Bubblewrap/Seatbelt/Docker backends; network egress denylist; file-write scope; `security.yaml` integration. *Principle 3. Pattern S-3.*
- **SPEC-V3R2-RT-004: Typed session state + checkpoint** — `.moai/state/` file-first layout, `PhaseState`/`Checkpoint` types, `interrupt()`-equivalent `BlockerReport`. *Principle 6. Patterns M-1 partial.*
- **SPEC-V3R2-RT-005: Multi-layer settings with provenance tags** — Source enum, merger with origin tracking, `moai doctor config dump` diagnostic. *Principle 7. Pattern X-2.*
- **SPEC-V3R2-RT-006: Hook handler completeness** — upgrade 5 handlers (subagent-stop tmux fix P-H02, config-change reload, setup, instructions-loaded, file-changed); retire 4 (notification, elicitation×2, task-created); document 27-event coverage. *Pattern T-5.*
- **SPEC-V3R2-RT-007: Hardcoded path fix + versioned migration** — replace `/Users/goos/go/bin/moai` with `$HOME/go/bin/moai`; migration framework with `CURRENT_MIGRATION_VERSION` + idempotent guards; auto-apply at session-start. *Problem P-H04. Pattern X-5.*

### 11.4 Orchestration (Layer 4) — 5 SPECs

- **SPEC-V3R2-ORC-001: Agent roster reduction 22→17** — merge manager-ddd+tdd into manager-cycle; merge 3 builders into builder-platform; retire expert-debug, expert-testing; split expert-frontend Pencil scope; scope-shrink manager-project. *Problems Cluster 6.*
- **SPEC-V3R2-ORC-002: Common Protocol CI lint** — reject literal `AskUserQuestion` / `Agent` tool / missing `effort` / `--deepthink` boilerplate; extract duplicate Skeptical Mandate to common-protocol. *Problem P-A01, P-A04, P-A13, P-A19.*
- **SPEC-V3R2-ORC-003: Effort-level matrix population** — publish table in agent-authoring.md; auto-populate all 17 agents per constitution guidance. *Problem P-A02, P-A03.*
- **SPEC-V3R2-ORC-004: Worktree isolation MUST for implementers** — upgrade SHOULD → MUST for agents touching ≥3 files; add `isolation: worktree` to manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher. *Problem P-A11, P-A22.*
- **SPEC-V3R2-ORC-005: Dynamic team generation formalization** — preserve general-purpose spawn + workflow.yaml role_profiles; document role → isolation/mode mapping; remove any static team-* agent definitions if present. *Existing SPEC-TEAM-001.*

### 11.5 Harness (Layer 5) — 3 SPECs

- **SPEC-V3R2-HRN-001: Harness routing + harness.yaml loader** — complexity estimator (domains × files × score) → minimal/standard/thorough; Go loader for harness.yaml (currently template-only); evaluator-profiles per level. *Problem P-H06.*
- **SPEC-V3R2-HRN-002: Evaluator fresh-memory amendment** — constitutional amendment to design-constitution §11.4: `evaluator.memory_scope: per_iteration` FROZEN; Sprint Contract state durable; judgment transcripts ephemeral. *Principle 4. Pattern E-1. Problem P-Z01.*
- **SPEC-V3R2-HRN-003: Hierarchical acceptance scoring** — evaluator-active scores per-criterion per sub-criterion (Agent-as-a-Judge §9 shape); rubric files in `.moai/config/evaluator-profiles/` per harness level. *Pattern E-1, E-3.*

### 11.6 Workflow (Layer 6) — 6 SPECs

- **SPEC-V3R2-WF-001: Skill consolidation 48→24** — execute merge waves 1-4 per r4-audit §v3 migration sequencing; retire moai-tool-svg, moai-docs-generation, moai-design-tools (split), moai-workflow-templates, moai-platform-database-cloud, moai-foundation-philosopher, moai-workflow-thinking, moai-design-craft, moai-domain-uiux, moai-foundation-context. *Problems Cluster 2.*
- **SPEC-V3R2-WF-002: Command thin-wrapper enforcement + /98-/99- extraction** — all command files ≤20 LOC; extract /98-github to `moai-workflow-github` skill; extract /99-release to `moai-workflow-release` skill; unify `.md.tmpl` extension. *Problem P-H08, P-H09, P-H18.*
- **SPEC-V3R2-WF-003: Multi-mode router** — `/moai run --mode {autopilot,loop,team}`; auto-select by harness; surface Ralph loop + team orchestration as first-class modes. *Pattern O-4. Principle 11.*
- **SPEC-V3R2-WF-004: Agentless vs multi-agent classification** — `/moai fix | coverage | codemaps | mx | clean` as fixed pipelines; `/moai plan | run | sync | design` as multi-agent; empirical validation at alpha.4 telemetry. *Pattern O-6. Principle 11. Problem P-Z02.*
- **SPEC-V3R2-WF-005: Language rules vs skills boundary decision** — codify `.claude/rules/moai/languages/*.md` as authoritative (rules) and remove stale `moai-lang-*` skill references; document rule-vs-skill boundary in agent-authoring. *Problem P-S08, P-S19.*
- **SPEC-V3R2-WF-006: Output styles CC schema alignment** — match CC's `{name, description, keep-coding-instructions, force-for-plugin}` frontmatter exactly; retain MoAI + Einstein; test serialization identity. *Pattern X-3.*

### 11.7 Extension (Layer 7) — 4 SPECs

- **SPEC-V3R2-EXT-001: Typed memory taxonomy** — formalize MemoryType {user, feedback, project, reference} in `~/.claude/projects/{hash}/memory/`; update `moai-memory.md` rule; add MemdirStore loader. *Patterns M-1, M-5. Principle 12.*
- **SPEC-V3R2-EXT-002: Output style + memdir loader** — share loader across `.claude/output-styles/` + memdir; CC-compatible. Pair with WF-006.
- **SPEC-V3R2-EXT-003: Plugin system design-only** — document 3-origin design (builtin/installed/inline) for v4 planning; no implementation in v3. *Pattern X-4. Deferred.*
- **SPEC-V3R2-EXT-004: Versioned migration auto-apply** — `MigrationRunner` with CURRENT_MIGRATION_VERSION; session-start hook trigger; migration list in `internal/migration/`; silent apply with log. *Pattern X-5. Problem P-C06.*

### 11.8 Cleanup / Migration — 3 SPECs

- **SPEC-V3R2-MIG-001: v2→v3 user migrator tool** — `moai migrate v3` one-shot upgrade: SPEC hierarchical acceptance wrapping, agent renames (manager-ddd/tdd→manager-cycle, builder-*→builder-platform), skill rename mapping, settings provenance-annotate pass. Executes all AUTO BCs (§8).
- **SPEC-V3R2-MIG-002: Hook registration cleanup** — document setup orphan + autoUpdate composite; align settings.json / Go handler / shell wrapper counts. *Problem P-R01.*
- **SPEC-V3R2-MIG-003: Config loader addition for 5 sections** — add Go struct + loader + test for constitution.yaml, context.yaml, interview.yaml, design.yaml, harness.yaml; decide sunset.yaml fate (activate or retire); complete workflow.yaml typed schema. *Problem P-H06, P-H07, P-H20.*

**Total: 35 SPECs.**

- Total REQs: 695
- Total ACs: 549 (402 Given/When/Then + 147 AC-NNN labels)
- Breaking SPECs: 16

---

## 12. Open Questions (for implementation phase)

From Wave 2 synthesis return message + new questions surfaced while authoring this master:

1. **Should `/moai plan` output become a formal task-DAG artifact?** LLMCompiler (R1 §15) and MLE-bench (R1 §21) suggest DAG reduces latency and reveals genuine dependencies. Open: does Go's concurrency model plus moai's existing team-mode file-ownership already capture this implicitly, or does the plan phase need an explicit `task-dag.yaml`? Needs empirical measurement at beta.1.
2. **Where does skill distillation happen?** Voyager auto-distills skills from successful code; AWM auto-induces workflows. Moai currently relies on human authorship via builder-platform. Open: pilot in v3.1+ with 1-2 workflow categories (e.g., "add REST endpoint"); gate via graduation protocol. Do not auto-distill in v3.0.
3. **Is the evaluator itself an agent or a pipeline?** Agent-as-a-Judge says agent (matches human reliability); Agentless says pipeline (simpler, cheaper). Open: which wins for moai's specific workloads? Likely both — Agent-judge for thorough harness (SPEC-V3R2-HRN-003), pipeline-scorer for minimal/standard. Empirical via harness telemetry.
4. **How persistent is episodic memory exactly?** Reflexion scopes per trial; Generative Agents span days; AWM spans workloads. Open: for moai, the right unit is per-SPEC (episodic), per-project (procedural workflow memory), per-machine (global lessons.md). v3 codifies this in SPEC-V3R2-EXT-001 but exact retention TTLs need telemetry.
5. **When does moai invoke `xhigh` vs `max` effort?** Anthropic guidance: xhigh is starting point; max only when evals show headroom. Open: per-agent telemetry `(effort_level, duration, success)` aggregated weekly, auto-recommend. SPEC-V3R2-ORC-003 populates per matrix; dynamic tuning deferred.
6. **Does `manager-design` become a first-class agent?** P-A21 flagged the gap; deferred to post WF-003 validation. Decision point: if `/moai design` stabilizes without it, retire the question; if orchestration pain surfaces, add in v3.1.
7. **Sandbox compatibility with LSP server processes?** LSP servers need filesystem access + local sockets; running them inside a sandbox may require carve-outs. Open: validate at alpha.2 that `moai_lsp_*` ACI commands work with `sandbox: seatbelt` on macOS.
8. **Which Ralph primitives become first-class in `/moai loop`?** `STALE_SECONDS`, `progress.md`, `activity.log`, `errors.log`, `runs/` directory — which must be present per constitution? Decision at beta.2 in SPEC-V3R2-WF-003 refinement.

---

## 13. Non-Goals

v3 explicitly does NOT pursue these, despite Wave 1 evidence for some of them. Each is tagged with the Wave 1 source so future planners can re-open if the ecosystem shifts.

- **Codex integration** — removed from v3 scope at Wave 1 planning time. moai supports Claude + GLM (via CG mode); Codex/Gemini cross-provider teams are OMC territory, not moai's. *Source: user directive prior to Wave 1.*
- **Cursor-first UX** — moai is Claude Code-native. Cursor parallel-subagent-for-exploration and Composer multi-file agent are interesting but belong in a separate product. *Source: r2-opensource-tools.md honourable mentions.*
- **CC Bridge / cloud-relay architecture** — 33 files, OAuth, trusted-device-token. moai uses local tmux; bridge would require Anthropic partnership. *Source: r3-cc-architecture-reread.md §3.9, §5 divergence 3.*
- **CC Ink TUI fork** — 750KB vendored UI; moai stays headless and rides CC's renderer. *Source: r3-cc-architecture-reread.md §5 divergence 1.*
- **CC GrowthBook feature-flag gating** — OSS binary cannot hide commands behind analytics-backed flags. moai uses explicit `.moai/config/sections/*.yaml` keys. *Source: r3-cc-architecture-reread.md §5 divergence 4.*
- **Centralized MCP registry + marketplace** — moai treats MCP as per-project opt-in (`.mcp.json` user-owned). No registry API. *Source: r3-cc-architecture-reread.md §5 divergence 5.*
- **52-subcommand CLI** — moai's ~15 subcommands × 3 modes is the right surface. *Source: r3-cc-architecture-reread.md §5 divergence 6.*
- **ADAS / Meta-Agent harness synthesis** — research-grade; far-horizon; requires S-4/S-5 maturity. Revisit in v5+. *Source: r1-ai-harness-papers.md §16.*
- **DSPy prompt compiler / optimizer** — research framework with high production effort. moai lessons.md + evaluator-profiles are the manual equivalent for now. Revisit v5+. *Source: r1-ai-harness-papers.md §17.*
- **Tree-of-Thoughts as default reasoning** — 5-100× token cost. Available via `harness: thorough` opt-in only. *Source: r1-ai-harness-papers.md §4.*
- **Pass@N as default evaluation** — N× cost. Available via `--attempts N` flag gated by thorough harness. *Source: r1-ai-harness-papers.md §21.*
- **LangGraph-style graph framework** — P11 file-first primitive preference; frameworks requiring >500 LOC scaffolding rejected. *Source: pattern-library.md §Patterns deliberately NOT adopted; design-principles.md P12.*
- **Plugin marketplace** — X-4 3-Origin Plugin System deferred to v4+. *Source: pattern-library.md X-4.*
- **Publish-Subscribe shared message pool** — O-2; current team sizes don't warrant. Revisit past ~10 teammates. *Source: pattern-library.md O-2.*
- **Workflow memory auto-induction (AWM)** — M-4; requires ≥100-SPEC trajectory corpus. Revisit after telemetry accumulates. *Source: pattern-library.md M-4.*

---

## Appendix A: Design Principle → SPEC cross-reference

| Principle | Primary SPECs | Secondary SPECs |
|-----------|---------------|-----------------|
| P1 SPEC as Constitutional Contract | CON-001, SPC-001 | SPC-002, HRN-003 |
| P2 Interface Design Over Tool Count (ACI) | SPC-004, RT-006 | EXT-002 |
| P3 Fresh-Context Iteration | RT-004, WF-003 | HRN-002 |
| P4 Evaluator Judgments Fresh, Contract State Durable | HRN-002 | HRN-003 |
| P5 Typed State + Durable Checkpoint | RT-004 | RT-005 |
| P6 Permission Bubble Over Bypass | RT-002 | RT-005 |
| P7 Sandboxed Execution Default | RT-003 | — |
| P8 Hook JSON Protocol | RT-001 | SPC-002, RT-006 |
| P9 Parallelism via Explicit DAG | ORC-005, WF-003 | — |
| P10 Agent Count Matches Task Structure | WF-004 | ORC-001 |
| P11 File-First Primitives | RT-004, EXT-001 | CON-003 |
| P12 Constitutional Governance | CON-001, CON-002 | CON-003 |

## Appendix B: Problem → SPEC cross-reference

See §6 for the full 72-problem matrix. Key critical mappings:

- **P-A01 (9 agents with AskUserQuestion, CRITICAL)** → SPEC-V3R2-ORC-001 (roster) + SPEC-V3R2-ORC-002 (CI lint)
- **P-H02 (subagent-stop tmux bug, CRITICAL)** → SPEC-V3R2-RT-006 (handler upgrade)
- **P-H04 (hardcoded path, CRITICAL)** → SPEC-V3R2-RT-007 (path + migration)
- **P-H06 (5 yaml sections no loader, CRITICAL)** → SPEC-V3R2-MIG-003 (loader addition)
- **P-C01 (no permission bubble, CRITICAL)** → SPEC-V3R2-RT-002 (permission stack)
- **P-C03 (no sandbox default, CRITICAL)** → SPEC-V3R2-RT-003 (sandbox layer)
- **P-Z01 (evaluator memory cascade, HIGH)** → SPEC-V3R2-HRN-002 (amendment)

## Appendix C: Pattern → SPEC cross-reference

| Pattern ID | Name | Disposition | SPEC |
|-----------|------|-------------|------|
| R-1 ReAct | Reason-Act-Observe | ADOPT | implicit across ORC-001, HRN-003 |
| R-2 Self-Refine | Feedback-Refine | ADOPT | WF-003 (--mode loop) |
| R-3 Reflexion | Actor/Evaluator/Reflector | ADOPT | HRN-002, HRN-003 |
| R-6 Ralph Fresh-Context | Outer loop with fresh context | ADOPT | WF-003, RT-004 |
| O-1 LLMCompiler DAG | Parallel task DAG | ADOPT | WF-003 |
| O-3 Magentic Ledger | Dynamic task ledger | ADOPT | ORC-005 |
| O-4 Multi-Mode Router | Mode surface | CONSIDER (2-3 modes) | WF-003 |
| O-5 Plan/Act Mode | plan | acceptEdits | ADOPT | RT-002 |
| O-6 Agentless | Fixed pipeline for well-structured | ADOPT | WF-004 |
| M-1 Typed Memory | user/feedback/project/reference | ADOPT | EXT-001 |
| M-5 LLM-Selected + Staleness | Freshness caveats | ADOPT | EXT-001 |
| T-1 ACI | LM-centric command set | ADOPT (priority 1) | SPC-004, RT-006 |
| T-5 Hook JSON-OR-ExitCode | Dual protocol | ADOPT (priority 2) | RT-001 |
| S-1 Multi-Source Permission | 8-source stack | ADOPT (priority 3) | RT-002, RT-005 |
| S-2 Permission Bubble | bubble mode | ADOPT | RT-002 |
| S-3 Ephemeral Sandbox | Bubblewrap/Seatbelt/Docker | ADOPT | RT-003 |
| S-4 FROZEN + Graduation | Zone model | ADOPT | CON-001, CON-002 |
| S-5 5-Layer Safety | FrozenGuard/Canary/Contr./RateLim./Human | ADOPT | CON-002 |
| E-1 Agent-as-a-Judge | Intermediate scoring + fresh memory | ADOPT (priority 9) | HRN-002, HRN-003 |
| E-2 Sprint Contract | Criteria-state negotiation | ADOPT (priority 6) | HRN-002 |
| E-3 Rubric-Anchored + Independent Re-eval | 0.25/0.5/0.75/1.0 rubrics | ADOPT | HRN-003 |
| X-1 Markdown + YAML Frontmatter | One file = one artifact | ADOPT | ORC-002 (lint), WF-006 |
| X-2 Multi-Layer Settings Provenance | Source tags | ADOPT (priority 4) | RT-005 |
| X-3 Output Style Override | CC-compatible frontmatter | ADOPT | WF-006, EXT-002 |
| X-5 Versioned Migration preAction | Silent auto-apply | ADOPT | EXT-004, RT-007 |

## Appendix D: Wave 1 + Wave 2 file manifest

Wave 1 research (input, 33.5K words):
- `.moai/design/v3-redesign/research/r1-ai-harness-papers.md` (25 papers)
- `.moai/design/v3-redesign/research/r2-opensource-tools.md` (16+ tools)
- `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md`
- `.moai/design/v3-redesign/research/r4-skill-audit.md` (48 skills)
- `.moai/design/v3-redesign/research/r5-agent-audit.md` (22 agents)
- `.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md`

Wave 2 synthesis (input, 15K words):
- `.moai/design/v3-redesign/synthesis/design-principles.md` (12 principles + North Star)
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` (72 problems in 6 clusters)
- `.moai/design/v3-redesign/synthesis/pattern-library.md` (37 patterns in 7 categories)

Wave 3 output (this document):
- `docs/design/major-v3-master.md`

Wave 4 input (not in this document; for SPEC generation):
- Target: 35 SPECs per §11, prefix `SPEC-V3R2-`
- Output directory: `.moai/specs/SPEC-V3R2-*/spec.md`
- Generation method: one SPEC per file, EARS format, hierarchical acceptance criteria

---

**End of v3.0.0 Master Design Document (DRAFT).**
**Next action**: Wave 4 SPEC regeneration from §11 index.
