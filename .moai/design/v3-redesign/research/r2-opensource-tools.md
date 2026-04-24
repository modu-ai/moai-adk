# R2 — Opensource Agentic Coding Tools Survey

*Research Team R2 deliverable for MoAI-ADK v3 redesign*
*Date: 2026-04-23 · Target: 12+ tools (surveyed: 16) · Length: ~6,500 words*

---

## Executive summary

The opensource agentic-coding ecosystem has bifurcated into two architectural camps. Camp A — **persistent, file-first loops** — is typified by `iannuttall/ralph`, `snarktank/ralph`, `Pythagora/gpt-pilot`, and `Yeachan-Heo/oh-my-claudecode`. Camp A treats git + disk as the primary memory, uses fresh agent context per iteration, encodes progress in PRD/story/task files, and relies on naive persistence to beat sophisticated context management. Camp B — **graph/state orchestration** — is typified by `langchain-ai/langgraph`, `microsoft/autogen` (now Microsoft Agent Framework 1.0 GA April 2026), `stanfordnlp/dspy`, and `crewAIInc/crewAI`. Camp B treats agents as nodes with typed state transitions, optimises via compilers (DSPy) or checkpointers (LangGraph), and exposes first-class human-in-the-loop interrupt semantics. The third, hybrid camp — **IDE-embedded agents** — `cline/cline`, `sst/opencode`, `Aider-AI/aider`, and `continuedev/continue` — integrates deeply with the local environment (terminal, LSP, git, browser) and layers plan/act permission gates over raw tool calls.

The most widely copied architectural idea of 2025-2026 is the **Ralph Loop** — an infinite outer loop where each inner iteration gets a fresh agent context and state persists only on disk. `iannuttall/ralph` implements this in ~TypeScript with a 3-skill primitive (commit / dev-browser / prd), and `oh-my-claudecode` fold it in as one of six operating modes. The second widely copied idea is the **Agent-Computer Interface (ACI)** from `princeton-nlp/SWE-agent` (NeurIPS 2024): instead of exposing raw shell to the LM, design a small set of LM-centric commands (open, scroll, edit, search) with concise structured feedback and linter guardrails. Nearly every coding agent built after SWE-agent — including Claude Code itself — echoes this principle.

Safety architecture has become a first-class concern in 2026. Post-incidents (Claude Code's `rm -rf ~/`, Cline prompt-injection → npm token exfiltration, Alibaba LLM spontaneous cryptomining), the consensus is that approval prompts alone are insufficient — defense-in-depth requires ephemeral sandboxes (E2B, Modal, Cloudflare V8 isolates, Bubblewrap/Seatbelt/Landlock), network egress blocking, file-write scope controls, reversibility layers, and deterministic policy engines (OWASP Top 10 for Agentic Apps + Microsoft Agent Governance Toolkit, both released late-2025/Q1-2026). MoAI-ADK v3 should treat sandboxing as a required architectural layer, not a nice-to-have.

---

## Tool-by-tool analysis

### 1. iannuttall/ralph

- **Repo**: [github.com/iannuttall/ralph](https://github.com/iannuttall/ralph)
- **Stars**: 895 · **Forks**: 81 · **Language**: TypeScript · **Last push**: 2026-02-04
- **Core proposition**: Minimal CLI wrapper around "Ralph Wiggum Loop" — an infinite loop with fresh agent context per iteration, persistence purely via git + on-disk files.
- **Architecture**: Single-agent outer-loop dispatcher. No multi-agent; no graph. One story per iteration (`ralph build 1`), one commit per iteration.
- **Tool exposure**: Ralph delegates to a runtime agent (`codex` / `claude` / `droid` / `opencode`) chosen at `ralph install` time. Ralph itself does not expose tools to an LM; the underlying agent does.
- **Memory/context**: Entirely file-based. `.agents/ralph/config.sh` (config), `.agents/tasks/*.json` (PRDs), `.ralph/progress.md` (append-only log), `.ralph/activity.log`, `.ralph/errors.log`, `.ralph/runs/` (per-iteration logs). Context window is thrown away every iteration.
- **Hook/extension**: Skills system — installs `commit`, `dev-browser`, `prd` as Claude Code skills via `ralph install --skills`. Extension = add more skills, not more hooks.
- **Configuration**: `config.sh` (bash), `tasks/*.json`, `Guardrails.md` (Signs / lessons).
- **Safety/sandboxing**: Story-level locking via `in_progress` status; `STALE_SECONDS` auto-reopens stalled stories; `Guardrails.md` captures lessons from failures. No process-level sandbox — safety delegated to underlying agent.
- **Adopt lessons**: (a) File-as-memory pattern beats elaborate context compression; (b) explicit Signs/Guardrails document as an append-only learning channel (parallels moai `lessons.md`); (c) `STALE_SECONDS` crash recovery is a simple, proven primitive.
- **Avoid lessons**: No parallelism — Ralph is sequential-only. No quality gates between iterations (unlike moai's TRUST 5). No type-safe state schema.

### 2. snarktank/ralph & PageAI-Pro/ralph-loop

- **Repo**: [github.com/snarktank/ralph](https://github.com/snarktank/ralph), [github.com/PageAI-Pro/ralph-loop](https://github.com/PageAI-Pro/ralph-loop)
- **Stars**: 12k+ (snarktank, cited by 2026 blog posts)
- **Core proposition**: Same loop concept as iannuttall/ralph but with Docker-sandboxed execution and longer-running (multi-day) iterations.
- **Architecture**: Outer loop + Docker container per iteration. Memory via `progress.txt`, `prd.json`, git history.
- **Adopt lessons**: Docker-by-default execution — proves you can make the Ralph loop multi-day-safe with ephemeral containers.
- **Avoid lessons**: Docker startup tax per iteration; ties implementation to Docker (E2B/Modal/Cloudflare V8 isolates are now preferred in 2026 for 100× faster starts).

### 3. Yeachan-Heo/oh-my-claudecode (OMC)

- **Repo**: [github.com/Yeachan-Heo/oh-my-claudecode](https://github.com/Yeachan-Heo/oh-my-claudecode) · website [yeachan-heo.github.io/oh-my-claudecode-website](https://yeachan-heo.github.io/oh-my-claudecode-website/)
- **Stars**: 30,898 · **Forks**: 2,869 · **Language**: TypeScript · **License**: MIT · **Last push**: 2026-04-23 (actively developed)
- **Core proposition**: "Teams-first multi-agent orchestration for Claude Code" — a plugin that ships 19 specialized agents, 36+ skills, tmux-integrated real agent panes, and six operating modes.
- **Architecture**: Plugin over Claude Code's native plugin system (`.claude-plugin/`). Native teams API (requires `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`). Gateway hooks via "OpenClaw" bridge. Real tmux worker panes for `omc team` CLI.
- **Tool exposure**: Inherits Claude Code's tools; layers its own skill system on top (`.omc/skills/`, `~/.omc/skills/`).
- **Memory/context**: Multi-tier. `.omc/sessions/*.json` (session summaries), `.omc/state/agent-replay-*.jsonl` (execution traces), `.omc/artifacts/ask/` (advisor outputs). 6-tier memory claim (inspired by Claudify).
- **Hook/extension**: Six OpenClaw hook events: `session-start`, `stop`, `keyword-detector`, `ask-user-question`, `pre-tool-use`, `post-tool-use`. Template variables `{{sessionId}}`, `{{projectName}}`, `{{projectPath}}`, `{{prompt}}`, `{{toolName}}`, `{{question}}`.
- **Configuration**: `~/.claude/omc_config.openclaw.json` (gateway/hook definitions), `~/.claude/settings.json` (teams toggle), env vars (`OMC_ASK_ADVISOR_SCRIPT`, `OMC_PLUGIN_ROOT`, `OMC_OPENCLAW`, `OMC_OPENCLAW_DEBUG`).
- **Safety/sandboxing**: Inherits Claude Code's settings.json permissions; no additional sandbox layer. Prompt triggers `stopomc`, `cancelomc` act as kill switches.
- **Operating modes**: `/team` (staged pipeline: team-plan → team-prd → team-exec → team-verify → team-fix), `/autopilot` (single lead, autonomous end-to-end), `/ultrawork` / `ulw` (non-team burst parallelism), `/ralph` (Ralph loop adapter — persistence with verify/fix), `/ralplan` (iterative planning consensus), `/deep-interview` (requirements clarification), `omc team` CLI (real tmux workers: claude/codex/gemini panes).
- **Prompt triggers**: `deepsearch`, `ultrathink`, `cancelomc`, `stopomc`.
- **Adopt lessons**: (a) Explicit multi-mode router (team/autopilot/ultrawork/ralph) — matches moai's subcommand pattern but richer; (b) OpenClaw hook bridge is a well-designed shape — six events with template variables is cleaner than moai's current ad-hoc hooks; (c) cross-provider tmux teams (claude/codex/gemini in separate panes) is a proven pattern for CG-mode-style cost optimization; (d) `/ralph` as a first-class mode inside a broader orchestrator — don't silo Ralph as its own product.
- **Avoid lessons**: (a) 19 agents + 36 skills is likely on the edge of discoverability — curate ruthlessly; (b) npm package named differently from the plugin (`oh-my-claude-sisyphus` vs `oh-my-claudecode`) creates documentation drift; (c) lack of documented sandboxing — moai should not copy this gap.

### 4. anthropics/claude-code-action & anthropics/claude-code

- **Repo**: [github.com/anthropics/claude-code-action](https://github.com/anthropics/claude-code-action), [github.com/anthropics/claude-code](https://github.com/anthropics/claude-code)
- **Stars**: 7,247 (action), 117,205 (main) · **Language**: TypeScript / Shell · **License**: MIT / (Anthropic proprietary)
- **Core proposition**: Official Anthropic bindings. `claude-code` is the CLI home; `claude-code-action` is the GitHub Action wrapper that runs Claude Code inside a workflow.
- **Architecture**: Thin wrapper. Action dispatches via mode detection (tag/`@claude`, review, agent/automation). Base-action (`claude-code-base-action`) is the low-level runner. Uses `actions/create-github-app-token` for GitHub App-based permissions.
- **Tool exposure**: `allowed_tools` YAML input — e.g. `Bash(git:*),View,GlobTool,GrepTool,BatchTool`. Tools defined by the CLI itself.
- **Memory/context**: Session-scoped per workflow run. No cross-run persistence; each PR/issue event gets fresh context. Does integrate with repo state via git checkout.
- **Hook/extension**: Via plugins (in `anthropics/claude-code` repo `/plugins/`). Main repo ships `.claude/`, `.claude-plugin/`, `.devcontainer/`. Skills/agents/commands distinction comes from the CLI, not from these repos.
- **Configuration**: `claude_env` (YAML multiline key=value), `claude_args` (e.g. `--model opus`), OAuth token (`CLAUDE_CODE_OAUTH_TOKEN`) or API key. Multi-cloud: Anthropic direct, AWS Bedrock, Google Vertex AI, Microsoft Foundry.
- **Safety/sandboxing**: GitHub workflow isolation (ephemeral runner); granular GitHub App permissions (Contents R/W, PR R/W). No custom sandbox — relies on runner isolation.
- **Adopt lessons**: (a) `allowed_tools` parameter shape — CSV + parameterized (`Bash(git:*)`) is expressive and has become the de-facto standard; (b) Mode auto-detection from event context removes configuration burden; (c) Cloud-provider abstraction (single `ANTHROPIC_BEDROCK_BASE_URL` override) is elegant.
- **Avoid lessons**: (a) Two-repo split (action + base-action mirror) is a maintenance footgun; (b) sparse documentation inside the main repo — forces users to rely on external `code.claude.com/docs`.

### 5. cline/cline (formerly Claude Dev)

- **Repo**: [github.com/cline/cline](https://github.com/cline/cline)
- **Stars**: 60,682 · **Forks**: 6,241 · **Language**: TypeScript · **License**: Apache-2.0 · **Last push**: 2026-04-23 (very active)
- **Core proposition**: VS Code extension (+ JetBrains port + CLI 2.0) that runs an autonomous coding agent with human-in-the-loop approval for every destructive action. 5M+ installs.
- **Architecture**: Plan-then-Act mode. Analyzes file structure + source ASTs + regex searches + reads relevant files before planning. Core agent loop with tool-use iterations; each file change + terminal command requires user approval through GUI.
- **Tool exposure**: Native VS Code APIs for file ops; shell integration API (VS Code 1.93+) for terminal command + output capture; browser control; MCP servers for user-defined extensibility. 2026 addition: native subagents (v3.58) + CLI 2.0 headless mode.
- **Memory/context**: Session state within a task; token + cost tracking per task. No claimed long-term memory — tasks are the unit.
- **Hook/extension**: MCP servers. Cline can itself generate and install MCP servers on request ("add a tool" → creates new MCP server).
- **Configuration**: VS Code settings + `.clinerules` / `.cline/` directory for workflows and rules. Multi-provider API support (OpenRouter, Anthropic, OpenAI, Gemini, Bedrock, Azure, Vertex, Cerebras, Groq, LM Studio, Ollama).
- **Safety/sandboxing**: **Approval-per-action** as safety model. Plan mode prevents writes. However, Cline was the victim of a 2026 prompt-injection chain that exfiltrated npm tokens — showing approval-based safety is insufficient against supply-chain attacks.
- **Adopt lessons**: (a) Plan/Act mode separation — planned approach approved before write actions; (b) MCP-first extensibility enables user-programmable tool expansion without forking; (c) Token + cost tracking surfaced per task.
- **Avoid lessons**: (a) Approval-only safety — the 2026 npm-token exfil proved approval fatigue is exploitable; moai must not repeat; (b) no first-class sandboxing; (c) VS Code coupling limits headless/CI scenarios (addressed by CLI 2.0 in 2026 but late).

### 6. Aider-AI/aider

- **Repo**: [github.com/Aider-AI/aider](https://github.com/Aider-AI/aider)
- **Stars**: 43,799 · **Forks**: 4,274 · **Language**: Python · **License**: Apache-2.0 · **Last push**: 2026-04-09
- **Core proposition**: Terminal-native AI pair programming. Git-first: every edit is an LLM-authored commit. "Best git integration among terminal coding agents" per 2026 reviews.
- **Architecture**: Pipeline — User prompt → Repo-map + active files → LLM → AST/regex edits → auto-commit. Single-agent chat with mode switches.
- **Tool exposure**: "Edit formats" — `diff`, `whole-file`, `udiff`, plus prefill-based infinite-output handling for long generations. Uses Tree-sitter across 100+ languages for repo-map.
- **Memory/context**: Repo-map as context primer; active files tracked via `/add`, `/read-only`; mid-session model switch via `/model`; `/tokens` surfaces budget.
- **Hook/extension**: Slash-command chat protocol: `/add`, `/read-only`, `/ask`, `/architect`, `/diff`, `/run`, `/undo`, `/tokens`, `/model`. "AI?" inline code comments let you ask questions from inside the source file.
- **Configuration**: `.aider.conf.yml`, `.aiderignore`, env vars for API keys. Broad provider support.
- **Safety/sandboxing**: Git is the safety net — `/undo` = `git reset HEAD~1`. No process sandbox; relies on user reviewing each atomic commit.
- **Adopt lessons**: (a) **Tree-sitter repo-map** is the single most-adopted technique in the whole ecosystem — moai already has codemaps but could deepen Tree-sitter integration; (b) Atomic LLM-authored commits with descriptive messages = trivial rollback; (c) `/architect` + `/ask` + `/run` slash-command shape is clean, 9-command set that covers 95% of workflows; (d) inline `AI?` comment as secondary input channel.
- **Avoid lessons**: (a) No agent parallelism — Aider is strictly one-agent-one-user; (b) no SPEC/PRD layer — you must externally manage intent; (c) no TRUST-5-style quality gates.

### 7. continuedev/continue

- **Repo**: [github.com/continuedev/continue](https://github.com/continuedev/continue)
- **Stars**: 32,743 · **Forks**: 4,409 · **Language**: TypeScript · **License**: Apache-2.0 · **Last push**: 2026-04-23
- **Core proposition**: 2026 pivot from "IDE chat" to **"Continuous AI"** — source-controlled AI checks that run as GitHub status checks on every PR. `.continue/checks/*.md` defines markdown-authored agents.
- **Architecture**: Three-surface system sharing a common Core: (1) GitHub PR checks (flagship), (2) CLI `cn` (headless + TUI), (3) IDE extensions (VS Code in-process, JetBrains out-of-process via TCP/stdio CoreMessenger). Type-safe message-passing protocol (`ToCoreProtocol` / `FromCoreProtocol`).
- **Tool exposure**: Hub-distributed agents with pre-configured MCP tools; custom tools via `.continue/tools/`.
- **Memory/context**: Rules system (`.continue/rules/*.md`) — markdown conventions that become part of AI reasoning; checks (`.continue/checks/*.md`) — per-PR AI validators.
- **Hook/extension**: GitHub-native integration (PR status checks). Headless + TUI modes. Hub agents.
- **Configuration**: `.continue/` directory convention (`checks/`, `rules/`, `agents/`). VS Code `.vsix`, IntelliJ `.jar`, npm `cn` CLI.
- **Safety/sandboxing**: GitHub Action runner isolation; check-level scope (read-only PR diff). Runs in CI, not locally.
- **Adopt lessons**: (a) **Markdown-authored agents** (`.continue/checks/*.md`) — one file = one agent — is the cleanest config shape in the survey, aligns perfectly with moai's existing `.claude/agents/*.md`; (b) Source-controlled checks as CI status is a natural fit for moai's `/moai sync` phase; (c) three-surface Core sharing means moai could similarly abstract for future IDE/CI surfaces.
- **Avoid lessons**: (a) The 2025 pivot invalidated prior users' investment in IDE features — architectural churn has cost; (b) no local sandboxing since it runs in CI.

### 8. princeton-nlp/SWE-agent

- **Repo**: [github.com/princeton-nlp/SWE-agent](https://github.com/princeton-nlp/SWE-agent) (redirects to SWE-agent/SWE-agent)
- **Stars**: 19,044 · **Forks**: 2,057 · **Language**: Python · **License**: MIT · **Last push**: 2026-04-20 · NeurIPS 2024 paper
- **Core proposition**: Coined **Agent-Computer Interface (ACI)** — a purpose-built command set for LMs (vs raw shell). 12.5% pass@1 on SWE-bench (SoTA at time); 1.0 (Feb 2026) + Claude 3.7 is current SoTA.
- **Architecture**: LM + ACI + history processor. Two-step pipeline: inference (issue → PR) + evaluation (SWE-bench runner).
- **Tool exposure**: **LM-centric commands** — `open`, `scroll_up`, `scroll_down`, `find_file`, search-in-file, `edit` (line-range), plus integrated linter. Not raw bash.
- **Memory/context**: History processors collapse old observations; observations beyond last 5 become single-line summaries; error messages suppressed after first; "command ran successfully and did not produce any output" filler message for empty outputs.
- **Hook/extension**: ACI spec is itself the extension surface — you iterate on which commands exist and how they respond.
- **Configuration**: YAML agent + environment configs; demo files for in-context learning.
- **Safety/sandboxing**: Linter blocks syntactically-invalid edits from committing. Scientific sandbox (Docker per task instance) for reproducibility rather than safety.
- **Adopt lessons**: (a) **ACI is the single most-important architectural insight** — don't expose raw shell; design a compact command set optimized for LM cognition. moai agents currently still use generic Bash — v3 should consider an ACI layer; (b) Linter-as-guardrail at write time — block on syntax errors before commit; (c) History compression as a first-class concept; (d) Informative but concise feedback — every action has a structured response.
- **Avoid lessons**: (a) Single-agent — no multi-agent orchestration; (b) Python-only implementation limits embedding in Go/Rust hosts.

### 9. gpt-engineer-org/gpt-engineer

- **Repo**: [github.com/AntonOsika/gpt-engineer](https://github.com/AntonOsika/gpt-engineer) (org: gpt-engineer-org)
- **Stars**: 55,216 · **Forks**: 7,317 · **Language**: Python · **License**: MIT · **Last push**: 2025-05-14 (no 2026 activity — essentially archival)
- **Core proposition**: Original "spec → whole codebase" CLI. Precursor to Lovable (gptengineer.app). Community mission is "tools coding agent builders can use."
- **Architecture**: Modular `steps.py` pipeline — customizable step sequence. Clarification step (asks follow-ups on ambiguous requirements) before generating.
- **Tool exposure**: Writes files + executes commands; vision-input support (images + text for architecture diagrams).
- **Memory/context**: Logs execution steps for resumability (`logs/`). Preprompts (`preprompts/`) customize coding conventions, preferred frameworks.
- **Hook/extension**: Custom AI workflows via step definitions.
- **Configuration**: Preprompts dir; natural-language spec file.
- **Safety/sandboxing**: None documented; runs locally with full permissions.
- **Adopt lessons**: (a) **Clarification step before generation** — moai does this via AskUserQuestion in plan phase, worth reinforcing as explicit architectural requirement; (b) Preprompts = per-project conventions baked in — mirrors `.claude/rules/` purpose; (c) Vision-input for wireframes → code.
- **Avoid lessons**: (a) No multi-agent specialisation (single monolithic generator); (b) archival status shows greenfield-scaffold-only is a narrow niche that doesn't sustain.

### 10. Pythagora-io/gpt-pilot

- **Repo**: [github.com/Pythagora-io/gpt-pilot](https://github.com/Pythagora-io/gpt-pilot)
- **Stars**: 33,776 · **Forks**: 3,499 · **Language**: Python · **License**: Fair Source · **Last push**: 2026-04-17
- **Core proposition**: Simulated software company — 6 specialized agents (Product Manager, Architect, DevOps, Tech Team Lead, Developer, Code Monkey) decompose high-level idea into architecture → tasks → code.
- **Architecture**: Assembly-line pipeline. Each agent has a narrow, well-defined role. Tasks include (a) description, (b) automated test spec, (c) human verification step — enforcing TDD + review.
- **Tool exposure**: Workspace/ directory as output root; step-level task execution.
- **Memory/context**: Built-in SQLite or Postgres. State survives crashes (resumable).
- **Hook/extension**: Agent definitions are the extension surface.
- **Configuration**: `~/.gpt-pilot/` config; `app_type` hint (Web App / Script / Mobile App / Chrome Extension).
- **Safety/sandboxing**: Developer-in-the-loop at every task boundary. No process sandbox.
- **Adopt lessons**: (a) **Per-task TDD spec + human-verification description** — moai's TRUST 5 `Tested` gate could adopt this explicit "how the human checks" artifact; (b) SQLite-as-state backing is a low-overhead alternative to file-based state; (c) Task-size tuning observation — "too broad" and "too narrow" both degrade quality — matches moai's harness minimal/standard/thorough routing rationale.
- **Avoid lessons**: (a) Fair Source licence creates commercial friction; (b) opinionated stack (Node+Mongo+Express backend, Vite+React+Shadcn+Tailwind frontend) limits language neutrality — moai explicitly aims to be language-neutral.

### 11. smol-ai/developer

- **Repo**: [github.com/smol-ai/developer](https://github.com/smol-ai/developer)
- **Stars**: 12,192 · **Forks**: 1,083 · **Language**: Python · **License**: MIT · **Last push**: 2024-04-07 (dormant)
- **Core proposition**: "Junior developer" agent library — <200 lines of Python + prompts. Create-anything-app generator tuned via `prompt.md`.
- **Architecture**: Monolithic single-agent. Markdown-all-you-need philosophy (markdown prompt → whole-program synthesis).
- **Adopt lessons**: **Markdown is the correct authoring format for agent/skill prompts** — this insight is now universal; moai already embodies it. The "v1 rewrite to be importable" shape is a good meta-lesson — authoring shape evolves toward library-embeddable artifacts.
- **Avoid lessons**: Dormant since 2024; proves that without persistence/iteration, a prompt-only approach plateaus.

### 12. microsoft/autogen → Microsoft Agent Framework 1.0 (April 2026)

- **Repo**: [github.com/microsoft/autogen](https://github.com/microsoft/autogen)
- **Stars**: 57,358 · **Forks**: 8,644 · **Language**: Python · **License**: CC-BY-4.0 · **Last push**: 2026-04-15 (maintenance mode)
- **Core proposition**: v0.2 was a chat-based multi-agent framework. v0.4 (2024) redesigned as asynchronous event-driven. **In 2026 frozen → Microsoft Agent Framework** (unifies AutoGen + Semantic Kernel, GA 2026-04-03).
- **Architecture**: Three layers: Core API (message passing, event-driven, local + distributed runtime, Python + .NET), AgentChat API (opinionated prototyping), Extensions API (LLM clients, tools).
- **Tool exposure**: Async message-passing. Pluggable tools; OpenTelemetry observability native.
- **Memory/context**: Pluggable memory (Foundry, Mem0, Redis, Neo4j, custom). Checkpointing + hydration for long-running.
- **Hook/extension**: Orchestration patterns — sequential, concurrent, handoff, group chat, **Magentic** (manager agent builds dynamic task ledger, coordinates specialists + humans). Graph-based workflow engine for deterministic flows.
- **Configuration**: Python/.NET SDK; model-provider swap via one-line change (Azure OpenAI / OpenAI / Anthropic / Bedrock / Gemini / Ollama).
- **Safety/sandboxing**: Checkpointing + pause/resume; human-in-the-loop approval gates first-class in 1.0.
- **Adopt lessons**: (a) **Magentic pattern** (manager with dynamic task ledger + delegation) is very close to moai's manager-strategy; worth studying the formal structure; (b) Checkpointing + hydration for long-running — moai's state machine could gain this; (c) One-line model swap as a 1.0 promise — aligns with moai's CC/GLM/CG mode switching.
- **Avoid lessons**: (a) Breakneck architectural churn (v0.2 → v0.4 → consolidation into Agent Framework) has burned users; (b) Framework-over-primitives complexity — moai should stay closer to file-first primitives.

### 13. geekan/MetaGPT (now FoundationAgents/MetaGPT)

- **Repo**: [github.com/geekan/MetaGPT](https://github.com/geekan/MetaGPT)
- **Stars**: 67,365 · **Forks**: 8,547 · **Language**: Python · **License**: MIT · **Last push**: 2026-01-21
- **Core proposition**: **"Code = SOP(Team)"** — encode Standard Operating Procedures as prompt sequences, assign to role-specialized LMs. Product Manager → Architect → Project Manager → Engineer → QA.
- **Architecture**: Publish-subscribe shared message pool (not direct agent-to-agent calls). Structured outputs between phases (PRD docs, design artifacts, flowcharts, interface specs).
- **Tool exposure**: Per-role tool sets; structured output schemas.
- **Memory/context**: Message pool is the memory. Executable feedback (self-correction on code that fails to run).
- **Hook/extension**: Role definitions as primary extension.
- **Configuration**: `~/.metagpt/config2.yaml`; Python 3.9+.
- **Safety/sandboxing**: No process-level sandbox.
- **Adopt lessons**: (a) **Publish-subscribe message pool > direct agent calls** — avoids N×N coupling; scales better; moai's TeamCreate + SendMessage is closer to direct pairs + should consider a shared pool pattern; (b) Structured intermediate outputs between roles (PRD → design → tasks → code) — matches moai's SPEC → plan → run → sync; (c) Executable feedback loop (run → inspect error → revise) = Ralph-loop-lite at the task level.
- **Avoid lessons**: (a) Costly — each phase spends multiple LLM calls; (b) rigid SOP — hard to adapt outside greenfield web-app generation.

### 14. OpenInterpreter/open-interpreter

- **Repo**: [github.com/OpenInterpreter/open-interpreter](https://github.com/OpenInterpreter/open-interpreter)
- **Stars**: 63,280 · **Forks**: 5,507 · **Language**: Python · **License**: AGPL-3.0 · **Last push**: 2026-04-22
- **Core proposition**: "Natural-language interface for computers" — equip a function-calling LM with `exec(language, code)`, stream outputs to terminal as Markdown.
- **Architecture**: Single-agent REPL. Conversation history retained within session; explicit reset.
- **Tool exposure**: Single `exec()` tool (Python/JS/Shell/etc.); new `--os` mode (Anthropic-powered) for screen-interaction.
- **Memory/context**: Conversation history in memory; no persistent state.
- **Hook/extension**: LiteLLM abstraction layer (any OpenAI-compatible provider, local via LM Studio/Ollama).
- **Configuration**: `interpreter.offline`, `interpreter.llm.model`, `interpreter.llm.api_base`; `interpreter.safe_mode` (experimental).
- **Safety/sandboxing**: Recommended runbook is Google Colab or Replit sandbox (user responsibility); GitHub Codespaces path (press `,` on repo). `safe_mode` is experimental and not default.
- **Adopt lessons**: (a) Raw `exec()` is a very thin tool surface — simpler than ACI but works for interactive REPL; (b) LiteLLM abstraction = moai-friendly model-provider neutrality; (c) Markdown-streaming output is a universal UX pattern.
- **Avoid lessons**: (a) No sandboxing by default is the single biggest risk in the survey; Open Interpreter has been cited in multiple 2026 security reviews as "do not run on host"; moai should not copy this posture.

### 15. crewAIInc/crewAI

- **Repo**: [github.com/crewAIInc/crewAI](https://github.com/crewAIInc/crewAI)
- **Stars**: 49,660 · **Forks**: 6,810 · **Language**: Python · **License**: MIT · **Last push**: 2026-04-23
- **Core proposition**: Role-playing autonomous agent crews. Two complementary layers: **Crews** (autonomous agents collaborating) + **Flows** (deterministic event-driven production orchestration).
- **Architecture**: `crew.py` assembles agents + tasks. `config/agents.yaml` + `config/tasks.yaml` as YAML-defined specs. Independent of LangChain.
- **Tool exposure**: Per-agent tool lists. Short-term, long-term, shared memory tiers.
- **Memory/context**: Three memory tiers per agent (short-term, long-term, shared). Qdrant Edge storage backend (March 2026).
- **Hook/extension**: Event listeners (`HumanFeedbackRequestedEvent`, `LLMCallCompletedEvent`). Agent skills added in 2026.
- **Configuration**: `agents.yaml` + `tasks.yaml` + `crew.py`. Multi-provider (OpenRouter, DeepSeek, Ollama, vLLM, Cerebras, Dashscope).
- **Safety/sandboxing**: **Guardrails as first-class feature** — "handle errors, hallucinations, and infinite loops for reliable task execution." Root_scope for hierarchical memory isolation.
- **Adopt lessons**: (a) **YAML agent/task config** — very clean authoring experience. moai already uses YAML in `.moai/config/sections/`, could extend per-SPEC task YAML; (b) Crews (autonomous) + Flows (deterministic) duality matches moai's `/moai run` autonomy vs `/moai sync` deterministic flow; (c) Three-tier memory per agent (short/long/shared) is richer than moai's current memory model.
- **Avoid lessons**: (a) Closed commercial AMP suite creates feature gap between OSS and paid; (b) two-framework overlap (Crews + Flows) increases learning curve.

### 16. langchain-ai/langgraph

- **Repo**: [github.com/langchain-ai/langgraph](https://github.com/langchain-ai/langgraph)
- **Stars**: 30,139 · **Forks**: 5,153 · **Language**: Python · **License**: MIT · **Last push**: 2026-04-23
- **Core proposition**: Low-level stateful orchestration engine. LangChain 1.0 (2026) consolidated agents onto LangGraph — now the execution core.
- **Architecture**: Directed-cyclic graph. Nodes = agents/functions/decisions. Edges = data flow. Centralised `StateGraph` with typed state schema + immutable updates.
- **Tool exposure**: Tools attached to nodes; ToolNode primitive.
- **Memory/context**: Checkpointers save state at every execution step; short-term working memory + long-term cross-session; pluggable backends.
- **Hook/extension**: `interrupt()` primitive for human-in-the-loop; LangSmith observability integration.
- **Configuration**: Python code-first (no DSL); Pregel-inspired runtime.
- **Safety/sandboxing**: Durable execution survives failures; explicit interrupt semantics.
- **Adopt lessons**: (a) **Typed state schema with immutable updates** — prevents race conditions in parallel agent work, solves a problem moai currently papers over with file-ownership conventions; (b) Checkpointer-per-step for durable execution; (c) `interrupt()` is the cleanest human-in-the-loop primitive across all surveyed tools.
- **Avoid lessons**: (a) Python-only (Go port is a hypothetical); (b) Framework lock-in — 1.0 consolidation shows it will keep absorbing surface area.

### 17. stanfordnlp/dspy

- **Repo**: [github.com/stanfordnlp/dspy](https://github.com/stanfordnlp/dspy)
- **Stars**: 33,950 · **Forks**: 2,831 · **Language**: Python · **License**: MIT · **Last push**: 2026-04-21
- **Core proposition**: **Programming, not prompting.** Compositional modules with learnable parameters; optimizer compiles the pipeline to maximize a metric.
- **Architecture**: Module-based (inspired by PyTorch NN modules). `dspy.Predict` is the fundamental module; all others (ChainOfThought, ReAct, RLM) build on it. Signatures define I/O; optimizers (MIPROv2, BetterTogether, LeReT, GEPA) produce optimised prompts + few-shot demos.
- **Tool exposure**: Tools via `dspy.ReAct`; `dspy.RLM` (Recursive Language Model) for sandboxed Python REPL + recursive sub-LLM calls.
- **Memory/context**: Stateless per call; context managed through signatures + optimised demos.
- **Hook/extension**: Optimizers; new `dspy.Reasoning` (April 2026) for native reasoning-model output capture.
- **Configuration**: `dspy.configure(lm=...)`; LiteLLM-based provider support.
- **Safety/sandboxing**: `dspy.RLM` uses sandboxed Python REPL; `guard` against pkl-file loading; `ContextWindowExceededError` retry cap.
- **Adopt lessons**: (a) **Signature-first agent authoring** (I/O schema before prompt) — huge lesson for moai agent quality; could drive a v3 agent schema validator; (b) Compiler/optimizer concept — moai's learner (Section 7 Rule 5 lesson capture) could evolve toward offline optimizer; (c) Module composability (ChainOfThought wraps Predict wraps Signature) — clean layering.
- **Avoid lessons**: (a) Research framework — production wiring is still effortful; (b) learning curve high; (c) optimisation runs are expensive.

### 18. sst/opencode (bonus, for completeness)

- **Repo**: [github.com/sst/opencode](https://github.com/sst/opencode)
- **Stars**: 148,121 · **Forks**: 16,937 · **Language**: TypeScript · **License**: MIT · **Last push**: 2026-04-23
- **Core proposition**: Terminal-native agent (Go original, TS active) — **client/server architecture**, session persistence across SSH drops, multi-client, 75+ provider support.
- **Architecture**: Background server holds session; TUI/desktop/mobile clients connect. Two built-in agents (build = full access, plan = read-only); Tab-key mode switch. **LSP auto-loaded per language** — understands type system, dependencies.
- **Tool exposure**: File R/W, shell, browser, LSP-derived structural tools.
- **Memory/context**: Persistent server session. No training-data ingestion claim ("privacy-sensitive").
- **Hook/extension**: GitHub Action (`opencode.yml`), `.opencode/` config dir, MCP.
- **Configuration**: Provider/model pair string (`provider/model`), agent selector.
- **Safety/sandboxing**: Plan mode denies writes by default; build mode full-access. No process sandbox by default.
- **Adopt lessons**: (a) **Client/server architecture** — server survives client disconnect. moai's current model is process-per-session; for long-running v3 agents this becomes a limitation; (b) **LSP-auto-load** matches moai's existing SPEC-LSP-CORE-002 / powernap integration — validates the direction; (c) Tab-key mode switch is a clean affordance for plan↔act.

---

## Cross-cutting synthesis

### A. Top-5 patterns to adopt

**1. Agent-Computer Interface (ACI) over raw shell** — *(from SWE-agent, echoed in Claude Code tool namespaces, OpenCode LSP integration)*

Why: Exposing raw `Bash` to an LM wastes tokens on shell arcana and invites mistakes. An ACI is a curated, LM-optimised command set with structured, concise feedback and linter guardrails at write time. Empirically, ACI-based agents beat shell-only agents by 8× on SWE-bench.

How to apply to moai: Introduce a layered tool namespace in v3. Keep generic `Bash` for edge cases but push common operations (symbol search, file navigation, atomic edit, linter-gated write) through named commands. Hook into existing moai LSP (SPEC-LSP-CORE-002) for structural navigation commands. Block syntactically-invalid edits before they commit.

**2. File-first memory + fresh-context iteration (Ralph loop)** — *(from iannuttall/ralph, snarktank/ralph, oh-my-claudecode `/ralph` mode)*

Why: LM context is noisy, costly, and bounded. Git + on-disk state is free, bounded only by disk, and inherently auditable. Fresh-context iteration prevents context rot and matches the reality that humans also work in bursts.

How to apply to moai: Make `/moai loop` a first-class execution mode on par with `/moai run`, not an add-on. Formalise `.moai/state/` as the authoritative source of cross-iteration state, define append-only artifacts (`progress.md`, `activity.log`, `errors.log`, `runs/*`), and guarantee each iteration starts from a clean LM context. Ralph's `STALE_SECONDS` recovery primitive becomes the moai equivalent of crash-resume.

**3. Markdown-authored agents/skills with YAML frontmatter** — *(from Continue `.continue/checks/*.md`, Claude Code skills, moai already, crewAI `agents.yaml`)*

Why: Markdown is the universal authoring format; YAML frontmatter gives typed config without schema files. One file = one agent/skill/check is the cleanest discoverability pattern.

How to apply to moai: moai already uses this pattern. V3 should formalize a single schema (shared between agents, skills, commands, rules) and a validator. Consider Continue's pattern of treating PR checks as markdown agents — could become moai's `/moai review` primitives.

**4. Typed state with durable checkpointing + `interrupt()`** — *(from LangGraph, Microsoft Agent Framework 1.0, partly crewAI flows)*

Why: Parallel agent work without typed immutable state is a race-condition farm. Checkpoint-per-step lets long-running workflows survive failures. `interrupt()` is the cleanest human-in-the-loop primitive in the entire survey.

How to apply to moai: Define a typed `SessionState` schema with immutable updates, checkpoint at each phase boundary (plan→run→sync), and expose an `interrupt()`-equivalent that surfaces to AskUserQuestion instead of text I/O. Resumable agents (moai already claims this) need formal checkpoint/hydrate semantics.

**5. Ephemeral sandboxed execution as the default safety layer** — *(from 2026 security consensus, absent in most tools, present only in snarktank/ralph via Docker, partly in Codex via Landlock/seccomp)*

Why: Approval-per-action has been empirically proven exploitable (Cline npm-token exfil, Ona sandbox-escape). Defense-in-depth requires ephemeral isolated execution (E2B / Modal / Cloudflare V8 / Bubblewrap / Seatbelt / Landlock), network egress blocking, file-write scope enforcement. OWASP Top 10 for Agentic Apps (Dec 2025) codifies this as mandatory.

How to apply to moai: Add a **sandbox layer** to v3 agent execution. Opt-in per agent role (implementer roles default-sandboxed; reviewer roles read-only). Default network egress denylist for implementation teammates. Document compatibility with Bubblewrap (Linux) / Seatbelt (macOS) / Docker (CI). Do not rely solely on `permissions.allow` in settings.json.

### B. Anti-patterns observed

**1. Approval-fatigue-only safety** (Cline, Open Interpreter, default Claude Code)
Users click Approve reflexively after ~5 prompts. 2026 incidents prove this is exploitable. *moai currently has `permissions.allow` lists but no sandbox — matches the anti-pattern.*

**2. Direct agent-to-agent calling (N² coupling)** (crewAI direct crews, early AutoGen v0.2)
Scales poorly past 3-4 agents. MetaGPT's publish-subscribe message pool solves it. *moai's team mode uses SendMessage direct channel — scales to current team size but will not scale past ~10 teammates.*

**3. Raw shell as the agent's primary tool** (Open Interpreter `exec()`, older Claude Code setups without `allowed_tools`)
SWE-agent ACI is 8× better. *moai agents still have wide `Bash` scope in many agent definitions — worth auditing.*

**4. Framework churn without migration paths** (Continue 2025 pivot, AutoGen v0.2→v0.4→Agent Framework)
Users lose work when architectural direction shifts. *moai is about to do v3 — the lesson is: commit to backward-compat layers and migration tools (moai already has `moai migrate` command — keep investing).*

**5. Opinionated stacks limiting language neutrality** (gpt-pilot's Node+Mongo+React+Shadcn, MetaGPT's web-app-shaped SOP)
Locks out 80% of codebases. *moai explicitly aims for 16-language neutrality (CLAUDE.local.md §15) — maintain this as a first-class constraint.*

**6. Two-repo splits for same artefact** (claude-code-action + claude-code-base-action mirror)
Maintenance footgun; documentation drift inevitable. *moai-adk-go should not split templates into multiple repos.*

**7. npm package name ≠ plugin/CLI name** (oh-my-claudecode package = `oh-my-claude-sisyphus`)
Documentation friction; users type wrong install command. *moai: keep `moai`, `moai-adk`, and `moai-adk-go` aligned.*

**8. No task-size tuning / cliff at both ends** (gpt-pilot observation)
Tasks too broad → bugs; too narrow → LM loses context. *moai's harness minimal/standard/thorough routing addresses this — ensure the auto-estimator is good.*

**9. Framework-over-primitives complexity** (LangGraph, AutoGen, Agent Framework 1.0 — powerful but steep)
Users reach for Ralph instead because 200 LOC > 200 KLOC. *moai risk: v3 could over-engineer; Ralph's minimalism is a competitive benchmark. moai CLAUDE.md §16 self-check rule is the right countermeasure.*

**10. Static agent rosters in multi-agent frameworks** (early AutoGen pre-Magentic)
Users must pre-plan roster before knowing task; real tasks need dynamic teammate spawn. *moai already dynamic via workflow.yaml role_profiles + general-purpose spawn — this is correct.*

### C. Design space taxonomy (matrix)

| Axis | ralph | oh-my-cc | claude-code | cline | aider | continue | swe-agent | gpt-pilot | autogen/AF | metagpt | crewai | langgraph | dspy | opencode | moai-current |
|------|-------|----------|-------------|-------|-------|----------|-----------|-----------|------------|---------|--------|-----------|------|----------|--------------|
| **Session model** | outer-loop fresh | multi-mode | single-agent | task-scoped | chat-session | per-PR/CLI | per-task | resumable | async event-driven | pipeline | crew+flow | durable graph | stateless call | server-persistent | session + state files |
| **Memory model** | files+git | 6-tier + session | conv history | per-task | repo-map + chat | `.continue/rules` | history processor | SQLite/PG | pluggable (Mem0/Redis/...) | msg pool | 3-tier per-agent | checkpointer | signatures+demos | server session | auto-memory + lessons |
| **Tool exposure** | delegates | native CC tools | `allowed_tools` CSV | VS Code + MCP | slash cmds | hub+MCP | **ACI** (open/scroll/edit) | role-scoped | message passing | role-scoped | per-agent list | node tools | `dspy.ReAct` | LSP+shell | tool set per agent |
| **Parallelism** | sequential | tmux panes + teams | subagents | subagents (v3.58) | none | CI matrix | none | sequential | concurrent+group chat | message pool | crews parallel | fan-out nodes | per call | multi-client | sub-agents + teams |
| **Extension model** | skills (3) | plugin + hooks | plugins | MCP | slash cmds | `.continue/checks` | ACI redesign | role defs | orchestration patterns | role defs | agent+task YAML | node types | modules | plugins+MCP | agents+skills+rules |
| **Config format** | bash + json | json (openclaw) + env | YAML | VS Code settings | yml | `.continue/` md | YAML | `~/.gpt-pilot/` | Python/.NET SDK | `config2.yaml` | `agents.yaml`+`tasks.yaml` | Python | Python | json | YAML sections |
| **Determinism gates** | none | hooks | hooks | approval GUI | commit/undo | CI checks | linter | tests per task | checkpoint | structured outputs | guardrails | typed state | metric | plan mode | TRUST 5 + harness |
| **Safety model** | story lock | env stop triggers | permissions | approval | git undo | PR review | linter | HITL | HITL + checkpoint | structured IO | guardrails + root_scope | interrupt() | RLM sandbox | plan mode | permissions + bg deny |
| **Sandboxing** | none | none | OS-level off-by-default | none | none | CI runner | Docker (bench) | none | none | none | none | none | none | none | none (templates only) |
| **User interaction** | CLI | slash + CLI | CLI/API | GUI approval | slash chat | GitHub UI | ACI turn | GUI + CLI | SDK | CLI | SDK | Python | Python | TUI/desktop | AskUserQuestion only |
| **Multi-provider** | via agent | cross-provider team | Bedrock/Vertex/Foundry | 10+ | 10+ | hub | any | 3 | 6 | any | 6+ | LiteLLM | LiteLLM | 75+ | CC/GLM/CG |
| **SPEC / PRD layer** | PRD JSON | /ralplan | manual | manual | manual | checks | issue text | PRD doc | task ledger | PRD artifact | tasks.yaml | state typed | signature | manual | SPEC-First EARS |
| **Learning channel** | Signs+Guardrails | /learner → `.omc/skills/` | manual | manual | manual | `.continue/rules` | ACI iteration | human review | optimizer | executable fb | event listeners | metric tuning | optimiser compile | none | lessons.md + auto-memory |

### D. oh-my-claudecode deep-dive

**Located successfully** — primary: [Yeachan-Heo/oh-my-claudecode](https://github.com/Yeachan-Heo/oh-my-claudecode) (30.9k ⭐, actively developed as of 2026-04-23).

Also found closest matches:
- [TechDufus/oh-my-claude](https://github.com/TechDufus/oh-my-claude) — "ultrawork" keyword trigger, background hooks.
- [zephyrpersonal/oh-my-claude-code](https://github.com/zephyrpersonal/oh-my-claude-code) — inspired-by-OMC variant (0 stars, minimal).
- [huangdijia/oh-my-claude-code-plugins](https://github.com/huangdijia/oh-my-claude-code-plugins) — small marketplace (9 stars).

**Full manifest** (from README + website):

*Operating modes* (6):
1. `/team` — staged pipeline: team-plan → team-prd → team-exec → team-verify → team-fix (close analogue to moai's plan→run→sync)
2. `/autopilot` — single lead, autonomous E2E
3. `/ultrawork` (`ulw`) — non-team burst parallelism
4. `/ralph` — Ralph loop adapter with verify/fix
5. `/ralplan` — iterative planning consensus
6. `omc team` CLI — tmux panes (real `claude`, `codex`, `gemini` workers)

*Shortcuts* (10+): `/ccg` (codex+gemini synthesis), `/ask {provider}` (provider advisor routing), `/skill` (skill CRUD), `/learner` (extract reusable patterns), `/deep-interview` (requirements clarification), plus prompt keywords `deepsearch`, `ultrathink`, `cancelomc`, `stopomc`.

*Claimed counts* (README): 19 specialised agents (architecture, research, design, testing, data science — not enumerated), 36+ skills.

*Hook system* (OpenClaw bridge, 6 events): `session-start`, `stop`, `keyword-detector`, `ask-user-question`, `pre-tool-use`, `post-tool-use`. Template variables: `{{sessionId}}`, `{{projectName}}`, `{{projectPath}}`, `{{prompt}}`, `{{toolName}}`, `{{question}}`.

*Config locations*:
- Project: `.omc/skills/`, `.omc/sessions/*.json`, `.omc/state/agent-replay-*.jsonl`, `.omc/artifacts/ask/`
- User: `~/.omc/skills/`
- Global: `~/.claude/omc_config.openclaw.json`, `~/.claude/settings.json`

**Patterns unique to oh-my-claudecode vs official Claude Code**:
1. Explicit multi-mode router (team/autopilot/ultrawork/ralph/ralplan/pipeline) — CC has no equivalent "mode" axis.
2. OpenClaw hook bridge — a unified gateway over CC's hook surface with template variables.
3. Cross-provider tmux teams (`claude`, `codex`, `gemini` in separate panes) — CC does not natively coordinate alternative agents.
4. `/learner` slash command that extracts reusable patterns into `.omc/skills/` — self-improving skill library.
5. `/ralph` mode integration — first-class Ralph loop adapter inside a broader orchestrator.

**Gap vs moai-adk**:

| Concept | OMC | moai |
|---------|-----|------|
| Mode router | 6 modes, Tab-switchable in shortcuts | plan/run/sync + fix/loop/review/... |
| SPEC-First | No EARS/acceptance criteria layer | Yes (SPEC + acceptance criteria + EARS) |
| TRUST 5 | No equivalent quality framework | Yes (Tested/Readable/Unified/Secured/Trackable) |
| Memory tiers | 6-tier claim | auto-memory + lessons.md + sessions |
| Hook surface | 6 OpenClaw events | Many (SessionStart/End, PreToolUse, PostToolUse, UserPromptSubmit, SubagentStop, Team events…) |
| Languages neutral | Primarily English | 16-language neutrality mandate |
| Sandboxing | None | None (matching anti-pattern — v3 should diverge) |
| Self-improving skills | `/learner` → `.omc/skills/` | lessons auto-capture (SPEC-SLQG-001) — comparable |
| Cross-provider panes | claude/codex/gemini tmux | CG mode (claude+GLM); could add codex/gemini |
| Real tmux workers | `omc team` CLI | moai CG mode uses tmux session-level env |
| Plugin packaging | CC-native plugin + npm | embedded templates + `moai init` |

**Strategic takeaway for moai v3**: OMC is moai's closest philosophical cousin in the Claude Code ecosystem. The main axes where moai is *ahead*: SPEC-First, TRUST 5, EARS, 16-language neutrality, SPEC-LSP integration, auto-memory semantics. The main axes where OMC is *ahead*: explicit mode router with Tab-switchable shortcuts, multi-provider tmux team orchestration out of the box, documented hook event template variables, `/learner`-driven skill auto-generation. V3 could adopt: (a) a clearer mode surface (beyond subcommands), (b) multi-provider tmux panes beyond CG mode, (c) codify hook template variables. V3 should *not* adopt OMC's lack of sandboxing or its sprawling 19-agent/36-skill surface without curation.

---

## Honourable mentions (not full-surveyed, noted for completeness)

- **huggingface/smolagents** (26.8k⭐, 2026-04-17) — 1,000-LOC minimalist framework. CodeAgent writes actions as Python code rather than JSON. 30% fewer steps than tool-calling agents. Sandboxes: Blaxel/E2B/Modal/Docker/Pyodide+Deno. *Lesson for moai*: code-as-action is a credible alternative to JSON tool calls.
- **Cursor / Cursor 3** (closed source but ubiquitous) — subagents parallel codebase exploration, Composer multi-file agent, Design Mode (April 2026). *Lesson*: parallel-subagent-for-exploration is now table stakes.
- **Claudify by hugo662** — 1,727 skills × 31 categories, 9 agents, 21 slash commands, 9 hooks, 6-tier memory. *Lesson*: sheer skill volume isn't a moat; curation + discoverability matters more.

---

## Sources

- [iannuttall/ralph](https://github.com/iannuttall/ralph) · [README](https://github.com/iannuttall/ralph/blob/main/README.md) · [Ralph Wiggum (article)](https://ralph-wiggum.ai) · [2026 Year of Ralph Loop](https://dev.to/alexandergekov/2026-the-year-of-the-ralph-loop-agent-1gkj) · [Mule AI overview](https://muleai.io/blog/2026-03-07-ralph-autonomous-ai-agent-loop/)
- [snarktank/ralph](https://github.com/snarktank/ralph) · [PageAI-Pro/ralph-loop](https://github.com/PageAI-Pro/ralph-loop)
- [Yeachan-Heo/oh-my-claudecode](https://github.com/Yeachan-Heo/oh-my-claudecode) · [OMC website](https://yeachan-heo.github.io/oh-my-claudecode-website/) · [Releases v4.12.0](https://github.com/Yeachan-Heo/oh-my-claudecode/releases)
- [hesreallyhim/awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code) · [awesomeclaude.ai](https://awesomeclaude.ai/)
- [TechDufus/oh-my-claude](https://github.com/TechDufus/oh-my-claude) · [zephyrpersonal/oh-my-claude-code](https://github.com/zephyrpersonal/oh-my-claude-code) · [huangdijia/oh-my-claude-code-plugins](https://github.com/huangdijia/oh-my-claude-code-plugins)
- [Claudify](https://claudify.tech/) · [hugo662/claudify submission](https://github.com/hesreallyhim/awesome-claude-code/issues/1089)
- [anthropics/claude-code](https://github.com/anthropics/claude-code) · [anthropics/claude-code-action](https://github.com/anthropics/claude-code-action) · [Claude Code GitHub Actions docs](https://code.claude.com/docs/en/github-actions) · [Claude Code product page](https://www.anthropic.com/product/claude-code)
- [cline/cline](https://github.com/cline/cline) · [Cline marketplace](https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev) · [Cline 2026 review](https://visionstack.visionsparksolutions.com/reviews/cline/) · [cline.bot](https://cline.bot)
- [Aider-AI/aider](https://github.com/Aider-AI/aider) · [aider.chat docs](https://aider.chat/docs/) · [Best CLI Coding Agents 2026](https://toolradar.com/guides/best-ai-terminal-coding-agents) · [2026 review](https://vibecodinghub.org/tools/aider)
- [continuedev/continue](https://github.com/continuedev/continue) · [docs.continue.dev](https://docs.continue.dev/) · [DeepWiki: continue](https://deepwiki.com/continuedev/continue) · [Continue 2026 review](https://vibecoding.app/blog/continue-dev-review) · [Beyond Code Generation blog](https://blog.continue.dev/beyond-code-generation-how-continue-enables-ai-code-review-at-scale)
- [princeton-nlp/SWE-agent](https://github.com/princeton-nlp/SWE-agent) · [SWE-agent paper (arXiv)](https://arxiv.org/pdf/2405.15793) · [Princeton NLP publication](https://collaborate.princeton.edu/en/publications/swe-agent-agent-computer-interfaces-enable-automated-software-eng/) · [docs/background](https://github.com/princeton-nlp/SWE-agent/blob/main/docs/background/index.md)
- [AntonOsika/gpt-engineer](https://github.com/AntonOsika/gpt-engineer) · [gpt-engineer-org](https://github.com/gpt-engineer-org) · [GPT Engineer 2026 review](https://vibecoding.app/blog/gpt-engineer-review) · [Spec-kit walkthrough](https://matsen.fredhutch.org/general/2026/02/10/spec-kit-walkthrough.html)
- [Pythagora-io/gpt-pilot](https://github.com/Pythagora-io/gpt-pilot) · [gpt-pilot wiki](https://github.com/Pythagora-io/gpt-pilot/wiki/) · [Pythagora.ai](https://www.pythagora.ai) · [Pythagora blog](https://blog.pythagora.ai/430/)
- [smol-ai/developer](https://github.com/smol-ai/developer) · [huggingface/smolagents](https://github.com/huggingface/smolagents) · [smolagents 26k stars analysis](https://www.decisioncrafters.com/smolagents-build-powerful-ai-agents-in-1-000-lines-of-code-with-26-3k-github-stars/)
- [microsoft/autogen](https://github.com/microsoft/autogen) · [Microsoft Agent Framework overview](https://learn.microsoft.com/en-us/agent-framework/overview/agent-framework-overview) · [Agent Framework 1.0 GA](https://devblogs.microsoft.com/agent-framework/microsoft-agent-framework-version-1-0/) · [Migration guide](https://learn.microsoft.com/en-us/agent-framework/migration-guide/from-autogen/) · [AutoGen 2026 guide](https://sanj.dev/post/autogen-microsoft-multi-agent-framework)
- [geekan/MetaGPT](https://github.com/geekan/MetaGPT) · [MetaGPT paper (arXiv)](https://arxiv.org/html/2308.00352v6) · [MetaGPT intro docs](https://github.com/geekan/MetaGPT-docs/blob/main/src/en/guide/get_started/introduction.md) · [MetaGPT 2026 explainer](https://aiinovationhub.com/metagpt-multi-agent-framework-explained/)
- [OpenInterpreter/open-interpreter](https://github.com/OpenInterpreter/open-interpreter) · [OpenCodeInterpreter](https://github.com/OpenCodeInterpreter/OpenCodeInterpreter) · [Open Interpreter README](https://github.com/openinterpreter/open-interpreter/blob/main/README.md)
- [crewAIInc/crewAI](https://github.com/crewAIInc/crewAI) · [CrewAI changelog](https://docs.crewai.com/en/changelog) · [Releases](https://github.com/crewAIInc/crewAI/releases) · [game-builder-crew example](https://github.com/crewAIInc/crewAI-examples/tree/main/crews/game-builder-crew)
- [langchain-ai/langgraph](https://github.com/langchain-ai/langgraph) · [langchain.com/langgraph](https://www.langchain.com/langgraph) · [LangGraph overview docs](https://docs.langchain.com/oss/python/langgraph/overview) · [LangGraph vs LangChain 2026](https://docs.bswen.com/blog/2026-04-16-langgraph-vs-langchain/) · [IBM LangGraph primer](https://www.ibm.com/think/topics/langgraph)
- [stanfordnlp/dspy](https://github.com/stanfordnlp/dspy) · [dspy.ai](https://dspy.ai/) · [Stanford HAI DSPy](https://hai.stanford.edu/research/dspy-compiling-declarative-language-model-calls-into-state-of-the-art-pipelines) · [dspy v1 branch](https://github.com/stanfordnlp/dspy/tree/v1)
- [sst/opencode](https://github.com/sst/opencode) · [opencode.ai](https://opencode.ai/) · [OpenCode docs for GitHub](https://opencode.ai/docs/github/) · [2026 review](https://www.openaitoolshub.org/en/blog/opencode-review-terminal-ai-coding)
- [NVIDIA sandboxing guidance](https://developer.nvidia.com/blog/practical-security-guidance-for-sandboxing-agentic-workflows-and-managing-execution-risk/) · [Bunnyshell coding agent sandbox](https://www.bunnyshell.com/guides/coding-agent-sandbox/) · [Cloudflare V8 isolate sandboxing](https://blog.cloudflare.com/dynamic-workers/) · [Microsoft Agent Governance Toolkit](https://opensource.microsoft.com/blog/2026/04/02/introducing-the-agent-governance-toolkit-open-source-runtime-security-for-ai-agents/) · [Coding Agents Supply Chain blog](https://securityboulevard.com/2026/03/coding-agents-widen-your-supply-chain-attack-surface/) · [Firecrawl AI Agent Sandbox](https://www.firecrawl.dev/blog/ai-agent-sandbox) · [Northflank sandbox comparison](https://northflank.com/blog/best-code-execution-sandbox-for-ai-agents)
- [Cursor product page](https://cursor.com/product) · [Cursor 3 launch post](https://dev.to/devtoolpicks/cursor-3-just-launched-with-an-ai-agents-window-what-changed-and-is-it-still-worth-it-496f)
