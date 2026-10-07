<p align="center">
  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
</p>

<h1 align="center">MoAI-ADK</h1>

<p align="center">
  <strong>A verification-driven agent orchestration harness — the structure that makes Claude Code's code trustworthy</strong>
</p>

<p align="center">
  English ·
  <a href="./README.ko.md">한국어</a> ·
  <a href="./README.ja.md">日本語</a> ·
  <a href="./README.zh.md">中文</a>
</p>

<p align="center">
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml"><img src="https://github.com/modu-ai/moai-adk/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
  <a href="https://codecov.io/gh/modu-ai/moai-adk"><img src="https://codecov.io/gh/modu-ai/moai-adk/branch/main/graph/badge.svg" alt="Codecov"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.3-blue.svg" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
</p>

<p align="center">
  <a href="https://adk.mo.ai.kr"><strong>Official Documentation</strong></a> ·
  <a href="https://adk.mo.ai.kr/book">Book: Practical Agentic Coding with Claude Code</a> ·
  <a href="https://discord.gg/Z7E7Mdc5aN">Discord</a>
</p>

---

> **"The model is a stochastic worker moving token by token. It cannot remember, turn to turn, what it used this turn and how much, whether the result is good, or how far the last session got. A harness enforces all three from the outside."**

---

## What's New in v3.2 — Factory Mode

A session spends one context window. A long SPEC fills it, and every later task carries everything before it. The plan is long finished yet stays in the window through the whole review, and the review stays through the whole write-up. The usual escape, `/clear`, throws away the context along with the baggage.

Factory Mode splits the work across **one leader session and several numbered lane sessions**. The leader watches the queue and hands cards to free lanes. A card does not change sessions at each stage; it goes **whole into one lane**, and that lane carries it through `plan → run → sync` in order, inside its own session. Each stage is spawned as an `Agent()` subagent, and the lane itself only orchestrates. Nothing is uncapped: each session's limit is still there. What changes is that one card's history accumulates only in the lane that owns it. The same budget therefore goes much further, and a lane empties its context after each card before taking the next.

### Getting started

```bash
moai cc -f                    # leader — opens the factory
moai cc -l                    # a lane, each in its own terminal — joins as the next free number
moai cc -l                    # lane-1, lane-2 … in order
moai glm -l                   # a lane on the GLM backend
moai codex -l                 # a Codex lane
```

`-f` (long form `--factory`) opens the leader, and `-l` (long form `--lane`) joins the running factory as a lane. Neither takes an argument. The operator never picks a lane number; it is assigned automatically. Lanes are launched **by hand, one fresh terminal each**, because a session cannot launch another session. `moai codex` has no leader entry and can only join as a lane.

A launch carries one entry token, so `-f` together with `-l` is an error. A value after either token (`moai cc -f <value>`, `moai cc -l lane-2`) is refused with a one-line error that names the correct form. The `-k` entry of the former multi-session board mode is removed, and `moai cg` exits with a migration notice (preview it with `moai migrate cg`). `moai gpt` does not exist: GPT models run through `moai codex`.

### Which backend to run

The backend is chosen per lane: `moai cc -l` is a Claude lane, `moai glm -l` a GLM lane, `moai codex -l` a Codex lane. The leader's seat is not for passing verdicts but for watching the queue and carrying cards, so GLM (`moai glm -f`), which is cheap to leave waiting, suits it. When one account starts hitting 429s, spreading lanes across accounts works. This mix is only one example — putting every session on a single backend is fine too.

### Many cards at once across N lanes

Grow the run one lane at a time by running `-l` once more. The number is assigned automatically, so passing `--name`/`-n` alongside is an error. A number is skipped only while a live session holds it — a dead lane's claim no longer blocks its number, but automatic assignment always takes the highest live number + 1, so it never backfills a gap in the middle. Lane ownership is recorded in `~/.moai/db/<project-key>/factory/factory.db`. When the base directory is a temporary directory (and no absolute `MOAI_HOME` override is set), it is recorded under the project-local `<base>/.moai/db/<project-key>/factory/` instead — the same exception the backlog queue makes. The old `.moai/state/factory/workers.json` is imported once and kept only as rollback evidence.

A lane runs at most 10 `Agent()` subagents at once, and write-capable spawns are isolated in their own worktrees. Don't light up the lanes all at once: bring up the first lane, confirm it is actually producing output, then start the rest. A card is never split across lanes.

A factory run records the process identity of the session that owns it. A run whose leader has died is retired automatically when the next lane joins, so the join does not stall on `AMBIGUOUS_FACTORY`. The reverse is covered too: when a lane joins and the run record is missing or retired while a live leader session exists, the join verifies that leader (pid plus process start fingerprint; the target is picked with the long form `--leader <name>`, default `leader`), restores the run record, and joins as usual — with more than one verified leader it fails closed, naming every candidate. `--leader` is a selector that composes with `-l`/`--lane` only; it is not a short form of any entry token. `moai factory runs` lists every run with its owner's liveness, and `moai factory runs --retire <run-id>` retires one named run by hand. It refuses if the owner is not actually dead.

After a run dies or is replaced, a lane session keeps pointing at the old run baked into its environment. Reattach it to a live run with `moai factory relaunch`. The command re-executes the provider's own lane-join line (`moai cc -l`, `moai glm -l`, and for Codex only `moai codex -l`), and its options are `--provider`, `--lane`, `--run`, `--from-run <run-id>` and `--dry-run`. Because `-l` takes no argument, a `--lane` you pass is not carried onto the join line: the command says so on stderr, and the lane joins as the next free number. `--from-run` retires that run first, but only when it is active and its owner is dead, and `--dry-run` prints the launch line and changes nothing. It is a one-shot command, separate from the `--clear-policy relaunch` loop that walks cards inside one run. The stale-run notice now shows this command filled in, instead of prose.

> Details: [Factory Mode](https://adk.mo.ai.kr/en/advanced/factory-mode)

Cards start in the queue. `backlog` has no owning session by design, so work enters the queue only when you put it there:

```text
/moai gtd "fix the stale rename hint"   # append a card
/moai gtd                               # list the queue
```

`/moai gtd` is the canonical task-management surface. `/moai todo` remains a compatibility name backed by the same SQLite queue, card IDs, order, and archive/restore behavior. `moai gtd capture|clarify|organize|reflect|engage` carries an item through Capture → Clarify → Organize → Reflect → Engage before approved work enters the unchanged `backlog → plan → run → sync → done` development flow. SQLite operation receipts and authoritative readback prevent a retry after a crash from silently duplicating publication, pick, or dispatch.

Two rules keep the factory honest. The leader advances a card **only on evidence it read** from the card's `progress.md` — never on a lane's reply, because a reply is a claim and inter-session delivery is not guaranteed. And when a card finishes, the lane asks for a `/clear`, since `/clear` is user-typed and cannot be sent as an instruction (`--clear-policy` changes this).

### Words the factory uses

The recurring vocabulary of the factory docs, gathered in one place. The **leader** is the session that watches the queue and carries cards; a **lane** is the pair of a session and its worktree that carries one card through to the end.

| Word | One-line definition |
|---|---|
| Card | One unit of work. Enters through `/moai gtd` and is called by a short id |
| Backlog | The entry queue. It has no owning session, so only a person can put cards in |
| Leader | The coordinating session. Advances cards on evidence it read and never writes code itself |
| Lane | A session + worktree pair that carries one card to the end. One parallel work stream, numbered `lane-1`, `lane-2` … automatically |
| Run id | The short identifier of one factory run. When several runs are alive, `--factory-run <run-id>` picks which one to join |
| Worktree | An isolated checkout just for the card. Its directory is the card id, its branch a `WT-<slug>` describing the work |
| Dispatch | The instruction the leader sends a lane — a pointer to the work, not a copy of it |

Cards differ in shape, and so do the stages a lane runs. When a card leaves the queue the leader sorts it into one of three classes and names it in the dispatch.

| Class | Shape | Shortcut |
|---|---|---|
| A — direct close | One file, one line, no design judgement, CI catches the regression | One lane takes it all the way to the PR (`plan` skipped) |
| B — defect, cause unknown | Clearly broken, cause not yet established | `run → sync` (no `plan`, no SPEC) |
| C — design change | Contains a decision, or spans subsystems | All three stages |

Class A is admitted on checked evidence, not assertion — a card that cannot cite a diff measured to one file and green CI on the head that will merge is not Class A. Class B skips only `plan`; the sync gate's review still runs, and the cause-establishing evidence (reproduction command and its output) is left in the card's progress record.

### Watching the factory

`moai web` serves a local console. The Factory screen shows the factory lanes alongside the SPEC pipeline, plus Overview, Specs, Monitor, Settings, and Todo screens.

<p align="center">
  <img src="./assets/images/moai-web-overview.png" alt="moai web console — Overview screen with SPEC counts, in-progress SPECs, and session registry" width="90%">
</p>

Full guide: [Factory Mode](https://adk.mo.ai.kr/en/advanced/factory-mode) · [manager-lead Leader Coordinator](https://adk.mo.ai.kr/en/advanced/manager-lead) · [`/moai todo`](https://adk.mo.ai.kr/en/utility-commands/moai-todo)

### What v3.1.1 adds

Earlier, v3.1.1 brought the following. Each one is covered in full in its own section further down.

**Home directory hygiene.** The longer you use it, the more leftovers from past runs pile up in `~/.moai`. `moai clean --home` clears them out, staying inside an allowlist — it is a dry run by default, so it shows you what would go before anything goes, and actual deletion needs `--force`. How old something has to be before it is swept is set by `state.home_retention_days` (30 days by default, `0` turns it off). To see how far the directory has grown right now, `moai doctor` reports it under Home Disk Usage. The home path itself can be moved with the `MOAI_HOME` environment variable — it takes absolute paths only. Only Go processes read it, though: move the path and the statusline and the shell hooks still look under `$HOME/.moai`. Shell-side credentials like `.env.glm` and the statusline's data stay behind, and your state quietly splits in two.

<p align="center">
  <img src="./assets/images/home-hygiene-infographic-en.png" alt="~/.moai home hygiene — MOAI_HOME keeps the path in one place, moai doctor reports usage, and moai clean --home deletes only inside the allowlist" width="85%">
</p>

**Cross-session messaging settings.** Whether a message from another Claude Code session arrives directly, waits for approval, or is refused outright is decided in `crosssession.yaml`. The switch that requires approval before a message leaves this machine lives there too.

<p align="center">
  <img src="./assets/images/cross-session-infographic-en.png" alt="Cross-session messaging — inbound, isolate_machines, and dialog_expiry control the receiving side. A message carries facts; approval stays with the user" width="85%">
</p>

**Statusline GitLab support.** `statusline.forge` picks whether open work is counted on GitHub or on GitLab. Left empty, it decides from the origin remote's host.

**A bare `/loop` becomes the factory foreman.** Typing `/loop` with no arguments starts a cycle that watches the backlog queue, dispatches the next card the operator has already marked `picked` to an isolated worker, confirms completion from evidence it read rather than from a claim, and reports. Nobody is watching that seat, so both putting cards in the queue and picking them stay the operator's job — the foreman never picks, it only carries.

---

## Why moai-adk?

The age of agents writing code has arrived, but you cannot take an agent's output on faith. Whether "the tests passed" is the result of actually running the tests or just the agent's guess has been the central problem from the start. moai-adk begins exactly there — it **bans unverified completion claims at the system level** and binds every completion claim to the command actually run and its output as evidence.

moai-adk is a harness that wraps Claude Code from the outside. It does not replace Claude Code; it takes over, in structure, the parts you used to manage by hand — which model to use, how deeply to reason, how to verify results, how to resume when a session breaks, how to keep parallel runs from stepping on each other. Verification integrity, the SPEC lifecycle, autonomous execution with real boundaries, a living codebase navigator, a self-improvement loop, and parallel-safe structure. These six form the identity of moai-adk.

<p align="center">
  <img src="./assets/images/why-harness-infographic-en.png" alt="An agentic development harness wrapping Claude Code" width="85%">
</p>

This identity organizes into three keys: **cost** (tokenomics — the same quality for fewer tokens), **self-improvement** (agentic loop engineering — turning observation into rules so the harness gets better as it runs), and **quality control** (the SPEC lifecycle, TRUST 5 gates, and isolation that prevents rework). No one of them suffices alone — below, why each needs the others.

### Eight differentiators

| Differentiator | What it means |
|---|---|
| **No false verification** | A claim that "tests pass" is always bound to the command actually run and its output. The system forbids presenting an unrun check as a success — verification-claim integrity is bound into every agent and orchestrator surface. |
| **Autonomy with real boundaries** | Declare a completion condition with `/moai goal` and the session works on its own until it holds. Four hard boundaries are attached — a turn limit (default 30), a stagnation guard, a wall-clock budget, and pre-approval gates — so it cannot fall into an infinite loop. |
| **Parallel-safe** | Every SPEC gets its own working tree, a branch-state guard blocks accidental branch switches in the primary checkout, and the gap against the remote is checked before spawning write agents. Two write-capable agents never run at the same time. |
| **Long-horizon continuity** | Work survives `/clear`. Progress stays in `progress.md`, handoff messages in memory, routing decisions in decision memory. The next session starts from what the last one learned, not from bare ground. |
| **Cost-efficient** | Pick the session's model and reasoning effort once, and every agent inherits it as-is. Prompt caches are reused and long output is spilled to disk to keep the context light. |
| **Equal support for 16 programming languages** | Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift — sixteen programming languages handled as one set via marker-based auto-detection. None receives preferential treatment. |
| **Self-improving** | Recurring failure patterns observed in the wild rise as proposed rule changes. Nothing is applied silently — approval comes first. Routing decisions and gate evidence accumulate in decision memory as material for the next run. |
| **Native-language friendly** | Korean, Japanese, Chinese, and English locales are maintained in the same PR, translationese is banned, and each language gets its own native prose. Users are never forced into English. |

### What's different

| | Claude Code alone | Typical harness | **moai-adk** |
|---|---|---|---|
| Evidence binding of completion claims | You check by hand | Usually absent | Enforced by the system (5-section evidence report format) |
| SPEC lifecycle | None | Limited | plan→run→sync 3-phase + Tier S/M/L |
| Hard boundaries on autonomous loops | N/A | Usually a turn cap only | Turn limit + stagnation guard + wall clock + approval gate |
| Parallel work isolation | Manual | Limited | worktree + branch guard + pre-spawn sync check |
| Session continuity | Broken by `/clear` | Limited | handoff + memory + progress files |
| Equal treatment of 16 programming languages | N/A | N/A | marker auto-detection + per-language toolchains |
| Self-improvement loop | None | Limited | failure observation → rule promotion (approval-gated) |

```mermaid
flowchart TD
    User["User request"] --> Analyze["Intent analysis<br/>Analyze-First routing"]
    Analyze --> Plan["plan — SPEC authoring"]
    Plan --> Audit["Independent audit<br/>plan-auditor"]
    Audit --> Run["run — TDD/DDD implementation"]
    Run --> Verify["trust-but-verify<br/>verification batch"]
    Verify --> Sync["sync — docs + PR"]
    Sync --> Learn["Decision memory + lessons"]
    Learn -.next session.-> Analyze
```

### The three keys hold each other up

Push the cost key alone and quality silently erodes — rework and debug loops follow, and rework is the most expensive token spend of all. Build quality gates with no learning loop and the same mistakes recur every session. Run an autonomous loop with no cost ceiling and a single runaway task drains the quota. The three keys hold each other up — **cost stays economical because quality prevents rework, quality stays enforceable because the loop captures what worked, and the loop stays affordable because cost gates stop it before overage.**

Every design decision serves one of these three keys. Which model to use, how deeply to reason, how to spend context — none of it is left to chance turn by turn. The system decides, and records the decision so the next run is smarter.

<p align="center">
  <img src="./assets/images/three-axes-infographic-en.png" alt="The three keys of moai-adk — Tokenomics · Agentic Loop · Agentic Harness" width="90%">
</p>

### Cost is determined by assignment, not unit price

Token prices fell **98% over three years** (Linux Foundation), yet enterprise AI spend rose **320%** in the same window. Volume growth overwhelmed the price drop. Agents spin through dozens to hundreds of steps to solve a single task, burning tokens proportionally. In usage-based pricing this becomes the invoice; in subscription, it eats the weekly quota shared by every model.

Uber deployed Claude Code to 5,000 engineers and **burned through a year of coding budget in four months**, then imposed monthly token limits. Meta, Amazon, and Microsoft each walked back unlimited-AI policies. **Tokenomics** — matching the model to the task to raise token efficiency — became the tech industry's new baseline.

Traditional cost control was built for rising unit prices, so it is helpless against this paradox: prices falling while total spend climbs. The bottleneck is not unit price but volume — more precisely, the step count an agent spins before finishing.

The DeepSWE leaderboard (113 tasks, per-effort view) demonstrates this. Within the same Claude family, per-task cost tracks how efficiently a model *finishes* — not what a token costs.

| Model [effort] | Score | Per-task cost | Note |
|---|---|---|---|
| opus-5 [low] | 58%±2 | **$1.66** | |
| opus-5 [medium] | **69%±1** | **$3.29** | **value-for-money knee** |
| opus-5 [high] | 73%±2 | $6.08 | +4pt score, 1.8× cost |
| opus-5 [xhigh] | 73%±3 | $9.07 | **net loss** — ties high, +49% cost only |
| opus-5 [max] | 74%±4 | $11.84 | |
| glm-5.2 [max] | 44%±2 | $3.92 | API-metered disadvantage · valuable under z.ai flat-fee |
| sonnet-5 [max] (Sonnet 5) | 54%±4 | $26.40 | Pareto-dominated by opus-5 [low] |

Opus 5 at its **lowest** effort scores higher than Sonnet 5 at its **highest** (58% vs 54%) while costing one-sixteenth as much per task ($1.66 vs $26.40) — even though Sonnet's per-token price is lower. The cause is 268 steps against 36: retry loops, not token rates, write the invoice. Cost is determined by **assigning the right model and reasoning depth to each task**, not by unit price.

The table above was measured on Opus 5. MoAI's `opus` alias now points to Opus 5.5 (requires Claude Code v2.1.280 or later; default effort `medium`), and the `sonnet` alias now points to Sonnet 5.5 (1M context per the official docs); neither has been re-measured yet.

<p align="center">
  <img src="./assets/images/why-tokenomics-infographic-en.png" alt="The Tokenomics Paradox — price down 98%, spend up 320%. The response: measure → route → diet → stop" width="80%">
</p>

![DeepSWE benchmark — model×effort score and per-task cost](./assets/images/deepswe-benchmark-2.png)

> Source: [DeepSWE v1.1 leaderboard](https://deepswe.datacurve.ai) (datacurve.ai, 113 tasks, 2026-07-25)

---

## Quick Start

### Install

#### macOS / Linux / WSL

```bash
curl -fsSL https://adk.mo.ai.kr/install.sh | bash
```

#### Windows (PowerShell 7.x+)

```powershell
irm https://adk.mo.ai.kr/install.ps1 | iex
```

#### Build from source (Go 1.26+)

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk && make build
```

Already installed? Run `moai update` to move to the latest version. From v3.2.0, `moai update` updates an existing project preservation-first — it no longer wipes the template-managed directories and redeploys. Files you added yourself survive in place, your edits to template files are 3-way merged (on a conflict your file stays and the new version lands next to it as a `<path>.moai-new.N` sidecar), and files the template no longer carries are moved to `.moai/archive/files/` before removal — the summary reports every refreshed, merged, conflicted, preserved, and archived path.

> 💡 **To cut costs — z.ai GLM recommended**: signing up via [this link](https://z.ai/subscribe?ic=1NDV03BGWU) grants bonus tokens. The link is also a way to sponsor moai-adk open-source development. Free models (GLM-4.7-Flash, GLM-4.5-Flash) exist too — see the [z.ai pricing](https://docs.z.ai/guides/overview/pricing).

### Project initialization

```bash
moai init my-project
cd my-project
```

The interactive wizard auto-detects language, framework, and methodology, and generates the Claude Code integration files.

#### Choosing the agent harness

The wizard asks which agent harness to deploy and wire; `--llm` gives the same choice non-interactively:

| Selection | What lands at the project root |
|---|---|
| `claude` (default) | The full `.claude/` surface plus `AGENTS.md` — today's default behavior |
| `gpt` | Codex only deployment: `AGENTS.md` and Codex surfaces (`.codex/`, `.agents/skills/`, `.moai/`) only. No `.claude/` tree, no `CLAUDE.md`, no `.mcp.json`. Claude-only runtime features (AskUserQuestion, sub-agent spawning, output styles, slash commands, Workflow scripts) are not available |
| `both` | Same `claude` deployment plus `.codex/` wiring; `.mcp.json` provisioning forced on |

#### Deploy mode: plugin default and full local deploy

On the default path (plugin mode), skills and commands are not copied into the project — the moai plugin carries them. The rest of the `.claude/` surface (agents, rules, hook registration, settings) deploys as today. To keep skills and commands as local files use `--no-plugin` (a full local deploy — including the `.mcp.json` moai entry and the Codex mirror); `--all` deploys every catalog tier locally as well. The deploy mode is recorded as `deployment_mode` in `.moai/config/sections/llm.yaml`, and `moai update` keeps the same scope per that record. A project whose plugin install could not be demonstrated is recorded on the safe side, as `local`.

```bash
moai init my-project --llm gpt   # Codex-only project
```

A project initialized before this choice existed has no `llm.harness` key and keeps the `claude` behavior on update — nothing to migrate.

> **GPT gateway withdrawn (2026-09-16).** The former `moai gpt` launcher — Claude Code driven by GPT
> models through the built-in translation gateway — has been removed. GPT models are reached through
> their native harness instead: `moai codex` (Codex CLI). The `--llm gpt` init value above is
> unaffected; it selects the Codex-only deployment, not the withdrawn launcher.

### First workflow

```bash
claude        # or moai cc — run Claude Code inside the project
```

```text
/moai plan "Add JWT login"      # author a SPEC
/moai run SPEC-AUTH-001         # TDD/DDD implementation
/moai sync SPEC-AUTH-001        # sync docs + create PR
```

Natural language works too. `/moai "fix the login bug"` triggers intent analysis (Analyze-First routing) to read the request and route to the appropriate workflow.

### Requirements

| Platform | Supported environments | Notes |
|---|---|---|
| macOS | Terminal, iTerm2 | Full support |
| Linux | Bash, Zsh | Full support |
| Windows | **WSL (recommended)**, PowerShell 7.x+ | Native cmd.exe unsupported |

- **Git** — required on all platforms
- **Claude Code** — moai-adk is a harness for Claude Code
- **Recommended**: `gh` CLI (PR automation), `tmux` (worktree windows), your language's lint/test toolchain (e.g. `golangci-lint`)

---

## Core Capabilities

### One entry point: `/moai`

Natural language and 16 subcommands feed the same pipeline. `/moai plan`, `/moai run`, `/moai sync` are the backbone of the SPEC pipeline; `goal`, `loop`, `fix`, `review`, `gate`, `clean`, `codemaps`, `e2e`, `mx`, `feedback`, `project`, `harness`, and `todo` fill out the surroundings.

> Four retired subcommands — `design` · `brain` · `coverage` · `security`. What `security` did is now covered by the `moai-ref-owasp-checklist` + `moai-ref-llm-security` skills.

### MCP server

`moai init` provisions exactly **one** active MCP entry by default — the self-hosted `moai mcp-server` (a local stdio server). It exposes the MoAI tools to Claude Code. The table below lists the main groups only — the full list is in the [MCP server guide](https://adk.mo.ai.kr/en/guides/mcp-server), and the authoritative tool count and list are what the installed binary returns from `tools/list`. Four documented-but-disabled entries (`context7`, `chrome-devtools`, `playwright`, `ast-grep`) are activated via `moai mcp add <name>`. The `moai mcp add|remove|list` CLI manages entries via an atomic-RWM seam — users never hand-edit `.mcp.json`.

| Group | Tools | Purpose |
|-------|-------|---------|
| SPEC lifecycle | `spec_progress`, `spec_audit`, `spec_drift` | Era classification + drift detection |
| Verification | `verify_snapshot`, `verify_trend` | Per-key evidence snapshots |
| Goal + session | `goal_arm`, `goal_status`, `session_list` | Autonomous loop + multi-session coordination |
| Cross-model audit | `audit_multi`, `claude_audit`, `codex_audit`, `glm_audit`, `audit_cache` | Multi-auditor convergence |
| Codex delegation | `codex_task`, `codex_setup`, `codex_job_*` | Background cross-model jobs |
| GLM delegation | `glm_task`, `glm_job_status`, `glm_job_result`, `glm_job_cancel` | GLM (z.ai) background job delegation |

All backends are fail-open — GLM (`~/.moai/.env.glm`) and codex (`~/.codex/auth.json`) are optional; an unavailable backend returns `inconclusive`, never a hard error.

In the Codex-enabled harness (`moai init --llm gpt|both`), Codex supports only built-in identifier arrays for its status line (`tui.status_line`), so MoAI-specific items (goal, todo, SPEC state) cannot be displayed — a limitation until openai/codex#17827 lands command-backed status lines.

> Details: [MCP Server Guide](https://adk.mo.ai.kr/en/guides/mcp-server) · [Claude Code MCP](https://adk.mo.ai.kr/en/claude-code/extensibility/mcp)

### Goal engine — an autonomous loop with real boundaries

Declare a completion condition and the session works on its own until it holds. A turn limit, a stagnation guard, a wall-clock budget, and pre-approval gates are attached, so it cannot fall into an infinite loop. Mechanical conditions (a command's exit code) and model conditions (a claim in the transcript) are both supported. `--max-turns 0` arms an auto-compact-driven infinite goal — in that case `--max-duration` and the stagnation guard provide the boundary.

`moai goal --auto "<mission>"` creates a separate `mission_mode=auto` draft; `approve` seals scope, actions, evidence, and limits once, while `run`, `status`, `revoke`, and policy-bounded `resume` operate on that persisted contract. Mission text is data, never shell or a goal condition. `super-advisor` remains non-binding, the read-only judgment sub-role of `manager-todo` proposes a structured decision, and deterministic owner adapters perform receipt-backed queue/dispatch effects, explicit-path commits, and leased local develop `--no-ff` merges. Without a provider proven to support durable start, reconnect, replacement, credentials, and process identity, the workflow falls back to `active-session-only`; remote push, PR, and merge completion remain unproven. [GTD and auto-mission guide](https://adk.mo.ai.kr/en/utility-commands/moai-gtd)

The final execution boundary is stricter: `run --supervise` follows the bounded publish→pick→leased disk dispatch→commit→local develop `--no-ff` plan. Supervised Git effects require separate `--card-worktree` and `--develop-worktree` paths; legacy `--repo` produces zero effects. Completion additionally requires a sealed `0600` `--completion-receipt` with true typed evidence and merged ancestry—exhausting the action list alone is not completion—and replay after completion has zero effects. `--recommend` grants no authority; every effect needs contained `0600` governor and independent-audit receipts. Unconfigured remote/release providers return `provider_unsupported` instead of simulating success.

### Parallel worktrees

Every SPEC gets its own working tree. Enter with `moai cc -w <name>`; add `--spawn` to open it in a new window while keeping the current session. A branch-state guard blocks accidental branch switches in the primary checkout.

### Factory Mode

`-f` (`--factory`) opens the factory leader and `-l` (`--lane`) opens a lane. Neither takes an argument. One leader carries the queue's cards to numbered lanes, and each lane takes a card through `plan → run → sync` to the end. The launch sequence is in the "What's New in v3.2 — Factory Mode" section above.

Session lineage is handled separately by the **Origin-Trail Chain**. It is independent of Factory Mode — no leader or lane is needed, and every session launched with a named worktree, as in `moai cc -w <name>`, lands on the chain. An append-only JSONL lineage tree tracks worktree ancestry, solves depth amnesia (root-to-leaf chain recovery after `/clear`), and marks a session whose heartbeat has gone quiet as `stale`.

| Concept | What it does |
|---------|-------------|
| Origin-Trail Chain | Append-only JSONL event stream at `.moai/state/chain/events.jsonl` |
| WorktreeNode (13 fields) | Per-session state: ID, parent, depth, origin chain, milestone, resume target |
| CWD-collision resolution | `(worktree_path, session_id)` pair disambiguates reused paths |
| Depth ceiling | Caps nesting complexity |

> **Available now**: `moai chain <status|lineage|back|list|prune>` reads the lineage, and `moai todo` (bare invocation lists the queue; subcommands `add` · `list` · `next` · `done` · `unpick` · `drop` · `undrop` · `edit` · `move` · `analyze`; two or more words become a new card) operates the `backlog` queue.

> Details: [Factory Mode](https://adk.mo.ai.kr/en/advanced/factory-mode) · [Session lineage chain](https://adk.mo.ai.kr/en/advanced/origin-trail-chain)

### CG retirement and migration

`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

This writes `llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, and `llm.gateway.teammate_provider: inherit`. It removes the old hybrid role assignment; it does not preserve a Claude leader with GLM teammate panes.

The `claude-glm` target describes a Claude leader with GLM teammates in tmux. Its apply and launch paths are currently unavailable because the TEAMMATE integration gate has not passed. Preview is available. Installing tmux or setting `verified: true` does not open this gate.

### Equal support for 16 programming languages

Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift. Marker-based auto-detection runs each language's standard lint/format/test toolchain.

### Automated quality gates

TRUST 5 (Tested · Readable · Unified · Secured · Trackable) applies to every change. `/moai gate` runs lint + format + type + tests in one pass, and sync-auditor scores across four dimensions: functionality, security, craft, and consistency.

### @MX tags

Inline code annotations that let AI agents exchange context, invariants, and danger zones. Only high-fan-in, complex, or dangerous code gets marked.

### Navigator — a living codebase map

`@NAV:DEC`, `@NAV:SYM`, and `@MX:SPEC` bind into one addressable graph (`nav-graph.json`). Design decisions, SPECs, and code symbols link in both directions — fix the code and the decision's context follows.

### Session handoff

Work survives `/clear`. A 6-block paste-ready resume message carries progress into the next session; in auto-inject mode, one message resumes the session.

### loop / fix — error-driven development

`/moai loop` sweeps LSP diagnostics, AST-grep, and linters in parallel, buckets issues by level, and runs until the queue drains. `/moai fix` is the single-pass variant.

### review --deep

`/moai review --deep` runs a multi-agent adversarial vulnerability scan, backed by OWASP · LLM-security · supply-chain · DevSecOps reference skills.

### 4-locale documentation

Korean, Japanese, Chinese, and English docs are maintained in the same PR. Translationese is banned, each language gets native prose, and a 4-locale parity check is bound into the build gate.

### moai web console

<p align="center">
  <img src="./assets/images/moai-web-settings.png" alt="moai web console — Settings screen with profile bar and setting tabs" width="90%">
</p>

`moai web` opens a console bound to localhost. Six screens — Overview, Factory, Specs, Monitor, Settings, Todo; the settings screen splits into these tabs: Identity, Language, Claude settings, GLM Settings, Codex settings, Workflow, Git & Worktree, Audit, Report, MCP, Cross-Session, Feedback, Quality Gate. The Codex settings tab is a read-only screen that gathers the scattered codex settings in one place — each value is still edited on its owning tab. Profile create/rename/delete lives on the same screen.

### ref / domain skills

Eleven ref skills (`moai-ref-api-patterns`, `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-react-patterns`, `moai-ref-testing-pyramid`, `moai-ref-ui-polish`, `moai-ref-secops`, `moai-ref-supply-chain`, `moai-ref-seo`, `moai-ref-git-workflow`, `moai-ref-cross-model-audit`) and seven domain skills (`moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`, `moai-domain-design-dna`, `moai-domain-html-report`, `moai-domain-humanize`, `moai-domain-svg-infographic`) inject field knowledge into agents.

### SVG technical infographics

The `moai-domain-svg-infographic` skill produces editable SVG technical infographics. Coordinates are computed numerically before any markup is written, and the finished file passes a deterministic source lint plus a dimension-verified 2x PNG render. An external-catalog benchmark measured nine forms — approval-gate flow, before-after comparison, KPI card grid, decision matrix, layer stack, nested scope, process flow, roadmap timeline, and component topology — and all nine proved reproducible (per-form artifacts and verdict: `.moai/reports/t272/verdict.md`).

### Cross-platform

A single Go binary with no extra dependencies, running on macOS, Linux, and Windows. The hook system enforces gates mechanically, and the statusline surfaces cost and context in real time.

---

## How It Works

### The SPEC 3-phase lifecycle

All work flows through plan → run → sync. Tier S/M/L size classification determines verification depth and PR routing. GEARS-format requirements and acceptance criteria judge completion by evidence.

```mermaid
flowchart TD
    P["plan — SPEC authoring<br/>GEARS requirements + acceptance criteria"] --> PA["plan-auditor<br/>independent audit (bias prevention)"]
    PA -->|PASS| R["run — TDD / DDD implementation<br/>cycle_type auto-selected"]
    PA -->|DEBT| P
    R --> SA["sync-auditor<br/>4-dimension quality scoring"]
    SA -->|PASS| S["sync — doc sync + PR"]
    SA -->|DEBT| R
    S --> MX["@MX tags + Navigator update"]
```

<p align="center">
  <img src="./assets/images/spec-3phase-infographic-en.png" alt="SPEC 3-Phase Workflow — plan → run → sync" width="80%">
</p>

Project state picks the methodology. `moai init` reads coverage and chooses automatically.

```mermaid
flowchart TD
    A["Project analysis"] --> B{"New project or<br/>10%+ coverage?"}
    B -->|"Yes"| C["TDD (default)"]
    B -->|"No"| D["DDD"]
    C --> F["RED → GREEN → REFACTOR"]
    D --> G["ANALYZE → PRESERVE → IMPROVE"]
```

| Methodology | Cycle | Target |
|-------------|-------|-----|
| **TDD** (default) | RED → GREEN → REFACTOR | New projects and feature work |
| **DDD** | ANALYZE → PRESERVE → IMPROVE | Existing code under 10% coverage |

### The 13-agent catalog

| Category | Agent | Role |
|----------|-------|------|
| **Manager** | manager-spec | Plan-phase SPEC authoring |
| | manager-develop | Run-phase TDD/DDD/autofix implementation |
| | manager-docs | Sync-phase documentation |
| | manager-git | PR creation and routing |
| | manager-design | Design-phase collaboration (Claude Design) |
| | manager-lead | Hierarchical-team Tier L coordination + factory leader-session dispatch (sole Agent-carrier, depth-2 sealed) |
| **Evaluator** | plan-auditor | Independent plan audit (bias prevention) |
| | sync-auditor | 4-dimensional quality scoring (Functionality 40 · Security 25 · Craft 20 · Consistency 15) |
| **Builder** | builder-harness | Project-specific agents, skills, commands, hooks scaffolding |
| **Advisor** | super-advisor | On-demand high-reasoning consultation (E1-E4 escalation) |
| **Specialist** | e2e-tester | Web/mobile/desktop E2E test execution (CLI-first) |
| | manager-todo | Todo-queue management (queue lifecycle, `/moai:todo --auto` serial cycle, dispatch guidance) — its read-only sealed-snapshot judgment sub-role returns one bounded decision and never applies it (dispatched by the GTD workflow, so it carries no selection-tree row) |
| **Built-in** | Explore | Read-only codebase exploration |

Every agent inherits the session's model and reasoning effort — the model and effort the session starts with are the assignment for all of them. Authoring and auditing are separated from the start, so the writing side never grades its own work.

Twelve of the thirteen are agents moai-adk built; `Explore` is a built-in that already ships with Claude Code.

### trust-but-verify — binding evidence to completion claims

When an agent reports "tests passed", the orchestrator does not take the claim at face value; it runs its own verification batch. Seven read-only verifications (tests, coverage, subagent boundaries, sentinel scans, CLI smoke, benchmarks, lint) run in parallel in a single turn, leaving each one's exit code and output as evidence.

The verification-claim integrity rule backs this flow — you must not present an unrun check as a success, must not pass off a previously measured value as a fresh measurement, and must not wave through what was never observed. The 5-section report format (Claim · Evidence · Baseline attribution · Gaps · Residual risk) binds every completion report from every agent and orchestrator.

### Trim verification cost, stop before overage

Verification is necessary; verification output sitting in context is not. Verbose output spills to disk files, leaving only the exit code and a bounded tail (max 50 lines) in context. Prompt-cache reuse (cached reads cost 0.1×) keeps the window light, and a context-diet `/clear` strategy issues recommendations at the thresholds (1M 50% / 200K 90%).

On the budget side, a token circuit breaker stands guard — it aborts at the hard limit (default 90%), saves progress to `progress.md`, and issues a paste-ready resume message. The statusline keeps context usage, cache hit rate, and rate-limit depletion visible at all times, so an overage never passes unnoticed.

### Reading the statusline

```
🤖 Opus | 🧠 xhigh·t | ♻️ 87% | 🔅 v2.1.212 | 🗿 v3.1.3 | ⏳ 2h 34m | 💬 MoAI
🪫 CW: ████████░░ 88% (⚠️/clear) | 🔋 5H: ████░░░░░░ 45% (4h 30m) | 🪫 7D: ████████░░ 82% (Jan 21)
📁 moai-adk-go | 📡 modu-ai/moai-adk, 7/3 | 🅱️ [WT] release/v3.1.3 +3 | 💾 +1 M2 ?0 | 📋 [run SPEC-AUTH-001-run] | 💌 PR #1042 (⌥approved)
🏷️ run | 👤 manager-develop | 🔄 TODO: 1/3
```

| Element | Meaning |
|------|------|
| 🤖 Model | Current active model |
| 🧠 effort | Reasoning effort — `·t` suffix when extended reasoning is active |
| ♻️ Cache hit rate | Prompt cache hit rate |
| CW: Context | Context-window usage + 2-stage `/clear` markers (⚠️ soft, 🛑 hard) |
| 5H / 7D | Plan usage rate + reset time |
| 📁 Directory | Project directory name |
| 📡 Repo | GitHub repo `owner/name` + open issues/PRs pair (`, 7/3`, or `, -/-` when unreadable) |
| 🅱️ Branch | Current branch — `[WT]` marks a worktree, `+` is the dirty count (modified+staged+untracked) |
| 💾 git status | Staged `+` · modified `M` · untracked `?` counts — one 💾 icon in every state |
| 📋 Task | Active SPEC workflow `[command SPEC-ID-phase]` |
| 💌 PR | Active GitHub PR number + review state (`⌥state`) |
| 🏷️ Session line | Conditional last line — session name · 👤 agent · 🔄 `TODO: in progress/queued` backlog |

> Details: [Statusline Guide](https://adk.mo.ai.kr/en/advanced/statusline)

---

## Workflow Examples

### Build a new feature (TDD)

```text
/moai plan "Add user profile image upload"
/moai run SPEC-PROFILE-001
/moai sync SPEC-PROFILE-001
```

New code, or code with sufficient coverage, gets TDD (RED → GREEN → REFACTOR). `moai init` detects project state and picks between TDD and DDD.

### Run long jobs (goal)

```text
/moai plan "Refactor the payment module"
/moai run SPEC-PAY-001
/moai goal "go test ./... exits 0 && lint clean, or stop after 20 turns"
```

Declare the completion condition and the session works on its own until it holds. The turn limit defaults to 30 and the stagnation guard is attached. When context reaches the threshold (1M 50% / 200K 90%), it recommends `/clear` and saves progress to `progress.md`.

### Run in parallel (worktree)

```bash
moai cc -w feature-auth        # open the auth working tree
moai cc -w feature-billing --spawn   # billing in a new window, current session kept
```

```text
# inside the auth tree
/moai run SPEC-AUTH-001

# inside the billing tree
/moai run SPEC-BILL-001
```

Each SPEC gets its own working tree so two agents never step on each other. The branch-state guard blocks accidental branch switches in the primary checkout.

### CG retirement and migration

`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.

```bash
moai migrate cg
moai migrate cg --target claude-only --apply --accept-role-change
```

This writes `llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, and `llm.gateway.teammate_provider: inherit`. It removes the old hybrid role assignment; it does not preserve a Claude leader with GLM teammate panes.

The `claude-glm` target describes a Claude leader with GLM teammates in tmux. Its apply and launch paths are currently unavailable because the TEAMMATE integration gate has not passed. Preview is available. Installing tmux or setting `verified: true` does not open this gate.

### Auto-fix bugs (loop)

```text
/moai loop
```

Sweeps LSP diagnostics, AST-grep, and linters in parallel, buckets issues by level, and runs until the queue drains. Single issues end in one `/moai fix` pass.

---

## Configuration and Profiles

### `.moai/config/sections/`

Project configuration splits into YAML section files. `moai init` lays down 30 section files in all; the six below are the ones you end up editing.

| Section | Role |
|---|---|
| `language.yaml` | User name · conversation language · code-comment language · commit-message language |
| `quality.yaml` | Quality gates · dev mode (TDD/DDD) · coverage |
| `harness.yaml` | Harness depth (minimal · standard · thorough) · auto-detection |
| `workflow.yaml` | Workflow behavior |
| `lsp.yaml` | LSP gate thresholds (SSOT) |
| `user.yaml` | User information |

v3.1.1 adds four more sections worth touching.

| Section | Role |
|---|---|
| `crosssession.yaml` | How cross-session messages are handled. `inbound` (empty, `accept`, `hold`, `refuse`), `isolate_machines` (whether a message leaving this machine needs approval), `dialog_expiry` (the deadline on the approval dialog for a held message) |
| `cache.yaml` | The prompt cache settings file. It holds `session_ttl` (`1h`, `5m`, `off`), `spec_ttl`, and the smallest chunk worth caching, and round-trips through the `moai web` settings editor. Nothing currently reads these values, so changing them does not change behavior |
| `state.yaml` | `home_retention_days` — how old something has to be before `moai clean --home` sweeps it. Read from the HOME tier only (`~/.moai/config/sections/state.yaml`); 30 days by default, and `0` turns home cleanup off |
| `statusline.yaml` | A `forge` key joins the existing theme and segment toggles. One of `github`, `gitlab`, or `none`, it decides which host the statusline counts open work on. Left empty it decides from the origin remote's host, so on a self-hosted instance write it in yourself |

`gate.yaml`'s `ast_grep_gate.rules_dir` is not a new key — it is a key whose **default changed**. What used to default to an empty string now defaults to `.moai/config/astgrep-rules`, and `moai init` / `moai update` lay the bundled ruleset down at that path. The fallback path on the code side is gone, so this key is now the only source of truth for where the ruleset lives — move the rules elsewhere and this value has to move with them, or the gate will not find them.

Environment variables override file values. For precedence details and the full section list, see the [CLI reference](https://adk.mo.ai.kr/en/cli-reference).

### settings.json / settings.local.json separation

| File | Role | Template |
|---|---|---|
| `.claude/settings.json` | Rendered from template — project-shared settings | Included |
| `.claude/settings.local.json` | Runtime-managed — per-machine values (tmux pane IDs · API tokens · absolute paths) | **Never included** |

`settings.local.json` is modified at runtime by `moai glm` and `moai cc`, and the SessionStart hook fills the environment. If accidentally committed, remove it with `git rm --cached .claude/settings.local.json`.

---

## Runs Anywhere

### Equal support for 16 programming languages

| | | | |
|---|---|---|---|
| Go | Python | TypeScript | JavaScript |
| Rust | Java | Kotlin | C# |
| Ruby | PHP | Elixir | C++ |
| Scala | R | Flutter | Swift |

Each language is auto-detected via project markers, and its standard lint/format/test toolchain runs. Missing tools are skipped quietly. The canonical Dart/Flutter name is "flutter". None receives preferential treatment.

### 4-locale documentation

| Locale | Site |
|---|---|
| 한국어 | adk.mo.ai.kr/ko |
| English | adk.mo.ai.kr/en |
| 日本語 | adk.mo.ai.kr/ja |
| 中文 | adk.mo.ai.kr/zh |

All four locales are maintained in the same PR, with a 4-locale parity check bound into the build gate. Translationese is banned; each language gets native prose.

### Operating systems

| Platform | Status |
|---|---|
| macOS | Full support (Terminal, iTerm2) |
| Linux | Full support (Bash, Zsh) |
| Windows | WSL recommended, PowerShell 7.x+ supported, native cmd.exe unsupported |

### Claude + GLM

z.ai GLM serves as an alternative backend for Claude Code. Switching is environment-variable only — the code stays the same.

| Command | Leader | Workers | tmux | Cost saving |
|---|---|---|---|---|
| `moai cc` | Claude | Claude | not required | — |
| `moai glm` | GLM | GLM | recommended | ~70% |

The GLM Coding Plan starts at $10/month. glm-5.3-flash (the default), glm-5.3, glm-4.7, glm-4.5-air, and free models (GLM-4.7-Flash, GLM-4.5-Flash) are available.

Each Claude tier maps to a GLM model through the `ANTHROPIC_DEFAULT_*_MODEL` environment variables:

| Claude tier | GLM model | Context |
|---|---|---|
| Opus | glm-5.3-flash | 1M |
| Sonnet | glm-5.3-flash | 1M |
| Haiku | glm-5.3-flash | 1M |
| Fable | glm-5.3-flash | 1M |

> glm-5.3 stays selectable in any tier slot (`llm.glm.models.*` in `llm.yaml`); switching a slot back is a one-line config change.

> Details: [Multi-LLM guide](https://adk.mo.ai.kr/en/multi-llm) · [z.ai pricing](https://docs.z.ai/guides/overview/pricing)

---

## Documentation and Learning

### Official documentation — adk.mo.ai.kr

The [adk.mo.ai.kr](https://adk.mo.ai.kr) online documentation is organized into 12 sections.

| Section | Description |
|---|---|
| [Getting Started](https://adk.mo.ai.kr/en/getting-started) | Introduction, installation, Windows guide, init wizard, quickstart, CLI overview, FAQ |
| [Core Concepts](https://adk.mo.ai.kr/en/core-concepts) | moai-adk identity, constitution, harness engineering, SPEC-based development, DDD, TRUST 5 |
| [Workflow Commands](https://adk.mo.ai.kr/en/workflow-commands) | `plan` · `run` · `sync` — SPEC pipeline backbone |
| [Utility Commands](https://adk.mo.ai.kr/en/utility-commands) | `fix` · `loop` · `gate` · `review` · `clean` · `codemaps` · `e2e` · `feedback` · `goal` · `gtd` (`todo` compatibility) |
| [CLI Reference](https://adk.mo.ai.kr/en/cli-reference) | Every `moai` binary command (49 total) |
| [Claude Code Guide](https://adk.mo.ai.kr/en/claude-code) | Claude Code integration — basics, context·memory, agentic, extensibility |
| [Multi-LLM](https://adk.mo.ai.kr/en/multi-llm) | CG migration and model policy |
| [Cost Optimization](https://adk.mo.ai.kr/en/cost-optimization) | Prompt caching strategies and token cost reduction |
| [Guides](https://adk.mo.ai.kr/en/guides) | CI automation, multi-LLM CI, and other operational recipes |
| [Git Worktree](https://adk.mo.ai.kr/en/worktree) | Worktree guide for parallel SPEC development |
| [Advanced](https://adk.mo.ai.kr/en/advanced) | Tokenomics, token budget, statusline, settings.json, hooks, @MX tags, skill guide, Harness v4 Builder, self-evolution, decision memory |
| [Contributing](https://adk.mo.ai.kr/en/contributing) | Open-source contribution guide |

### Book

[**Practical Agentic Coding with Claude Code**](https://adk.mo.ai.kr/book) — a hands-on harness engineering guide by the moai-adk author. [book.mo.ai.kr](https://book.mo.ai.kr)

### CLI command table (17 frequently used)

| Command | Description |
|---|---|
| `moai init` | Interactive project setup (auto-detects language/framework/methodology) |
| `moai doctor` | System state diagnosis and environment verification — the Home Disk Usage check reports, as advice, how far `~/.moai` has grown |
| `moai status` | Project status summary (Git branch, quality metrics) |
| `moai update` | Update to latest version (preserves local files · 3-way merge with conflict sidecars · archived removals) |
| `moai graph <build\|query>` | Build/query the codebase graph (edges.jsonl) — caller lookup, blast radius, milestone cross-checks |
| `moai cc` / `moai glm` | Claude-only / GLM-only sessions |
| `moai codex [cli\|status\|app]` | Codex launcher — called with no verb it launches the Codex CLI; `status` prints the readiness readout and starts nothing |
| `moai worktree <sync\|done\|sweep\|hoist\|remove\|clean\|recover\|snapshot\|verify\|restore>` | Git worktree maintenance (entering a worktree is the launchers' job) |
| `moai session <list\|register\|current>` | Multi-session coordination |
| `moai spec <audit\|archive\|lint\|list\|new>` | SPEC lifecycle tools |
| `moai goal <arm\|status\|clear>` | Goal engine CLI |
| `moai harness <status\|apply\|rollback\|disable>` | Harness learning lifecycle |
| `moai handoff <save\|show\|clear>` | Session handoff records |
| `moai preference <list\|decay-scan\|toggle>` | Decision memory management |
| `moai memory <doctor\|archive>` | Agent memory checks and archiving of stale entries |
| `moai tokens record` | Per-pool token usage ledger records |
| `moai clean [--home] [--codex-skills] [--reports-archive]` | Clear leftovers from past runs. With `--home` it sweeps `~/.moai` inside the allowlist; with `--codex-skills` it removes the `[[skills.config]]` registrations in `~/.codex/config.toml` whose declared path is provably absent. With `--reports-archive` it moves aging evidence directories from `.moai/reports/` into `archive/<YYYY-MM>/` (move-only, never deletes; default retention 90 days via `--reports-archive-days`). Exactly one scope per invocation. Dry run by default; `--force` to actually delete |
| `moai web` | Web console — 6 screens (Overview · Factory · Specs · Monitor · Settings · Todo), settings tabs |

> All 49 commands: [CLI reference](https://adk.mo.ai.kr/en/cli-reference)

### ref / domain skills

**ref (field knowledge) — 11**: `moai-ref-api-patterns`, `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-react-patterns`, `moai-ref-testing-pyramid`, `moai-ref-ui-polish`, `moai-ref-secops`, `moai-ref-supply-chain`, `moai-ref-seo`, `moai-ref-git-workflow`, `moai-ref-cross-model-audit`

**domain (specialist domains) — 7**: `moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`, `moai-domain-design-dna`, `moai-domain-html-report`, `moai-domain-humanize`, `moai-domain-svg-infographic`

`moai-domain-design-dna` is new in v3.1.1. Hand it one design to work from — a screenshot, a set of images, or a live URL — and it reverse-engineers a single Design DNA JSON covering the measurable values (color, spacing, corners, typography), the feel of that design, and its special rendering effects. Feed the JSON back in and it produces a new artifact carrying the same feel — a route that moves "make it look like this screen" across as values instead of as words. Diagram profiles are supported too: the active profile marker persists under the project root's `.design-dna/` so it survives `moai update`, and the opt-in mermaid and drawio importers treat their sources as untrusted input — coordinates, colors, fonts, and layout never carry over.

### CHANGELOG

Recent changes live in [CHANGELOG.md](./CHANGELOG.md).

### Code quality requirements

Every contribution passes the TRUST 5 gate — 85%+ coverage · lint errors 0 · type errors 0 · Conventional commits. Existing code is fixed in behavior by characterization tests then improved incrementally (DDD); new code follows RED → GREEN → REFACTOR (TDD).

---

## FAQ

### Why doesn't every function have an @MX tag?

That's normal. Tags mark high-fan-in, complex, or dangerous code only. In any project, most code never crosses a tag threshold — a file without tags is not a defect.

### What does the statusline version display mean?

```
🗿 v3.1.2 -> 🗿 v3.1.3
```

The first value is the currently installed moai-adk version; the arrow indicates an available update. It disappears after `moai update`.

### Can I use Claude only, without GLM?

Yes. `moai cc` starts a Claude session without requiring GLM. Only projects retaining legacy CG configuration require migration first.

### Does it work on existing projects?

Yes. `moai init` detects project state and selects the methodology — DDD (characterization tests fix behavior, then incremental improvement) for existing code under 10% coverage, TDD for new or well-tested code.

---

## Contribute

### Contributing

Contributions are welcome anytime. Detailed procedures live in [CONTRIBUTING.md](CONTRIBUTING.md).

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Write tests — TDD for new code, characterization tests for existing code
4. Verify tests, lint, format pass: `make test` · `make lint` · `make fmt`
5. Commit with a Conventional commit message and open a pull request

**Code quality requirements**: 85%+ coverage · lint errors 0 · type errors 0 · Conventional commits

### Feedback

Inside Claude Code, `/moai feedback` files bug reports and feature requests straight to GitHub issues. From the terminal, use [GitHub Issues](https://github.com/modu-ai/moai-adk/issues).

### Community

- [Discord](https://discord.gg/Z7E7Mdc5aN) — live discussion and tips
- [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — bug reports · feature requests

### License

[Apache License 2.0](./LICENSE) — see the LICENSE file for details.

---

## Star History

<a href="https://www.star-history.com/?type=date&repos=modu-ai%2Fmoai-adk">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&theme=dark&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=modu-ai/moai-adk&type=date&legend=top-left&sealed_token=9wFuBO5GMKxHZsaknxlIW3oypXLJlyW1qqq8T--aTRyfp6j9EK9KTR2vJvyAG8AKSs3Lindw7LUt-m-I6ysz9BoV6kdtrKlJYTViQAYR56A_3ie4ZVOqIw" />
 </picture>
</a>

<p align="center">
  <sub>Built by the MoAI-ADK team · <a href="https://adk.mo.ai.kr">adk.mo.ai.kr</a></sub>
</p>
