# Awesome Lists Verification Results — 2026-04-22

**Verifier**: MoAI orchestrator (manual cross-check via `gh search repos` + `gh api`)
**Purpose**: Resolve the 5 "UNVERIFIED" entries in `submission-plan.md` with live repo data and surface high-impact bonus targets.

---

## Previously Verified (unchanged)

| # | Repo | Stars | Last Push | Priority |
|---|------|------:|-----------|:--------:|
| 1 | `e2b-dev/awesome-ai-agents` | — | 2025-02-26 | **P1** |
| 2 | `avelino/awesome-go` | ~147k | 2026-04-22 | **P1** |

---

## Resolution of the 5 Unverified Candidates

### 1. `awesome-claude-code` — CONFIRMED (multiple instances)

Direct matches exist. Top three candidates ranked by reach × fit:

| Repo | Stars | Last Push | Fit | Priority |
|------|------:|-----------|-----|:--------:|
| `hesreallyhim/awesome-claude-code` | **40,283** | 2026-04-22 | Description explicitly lists "agent orchestrators" — direct fit | **P1 (flagship)** |
| `jqueryscript/awesome-claude-code` | 284 | 2026-04-22 | "tools, IDE integrations, frameworks" — fit | P2 |
| `ccplugins/awesome-claude-code-plugins` | — | 2026-04-22 | Plugin-specific, MoAI has plugins — partial fit | P3 (after plugin catalog) |

### 2. `awesome-llm-tools` — **NO MATCH, SKIP**

No repository with this exact name found. Closest adjacent (`awesome-llm-compression`, `awesome-gpt-prompt-engineering`, `awesome-llmops`) are specialization mismatches. **Recommendation**: Do not pursue.

### 3. `awesome-developer-tools` — **NO MATCH, SKIP**

Closest (`agamm/awesome-developer-first`) is a different niche (developer-targeted SaaS products). **Recommendation**: Skip.

### 4. `awesome-mcp` — **VERIFIED BUT WRONG FIT, SKIP**

Multiple strong repos exist (`punkpeye/awesome-mcp-servers`, `appcypher/awesome-mcp-servers`, etc.) but they curate **MCP servers**. MoAI-ADK is an MCP **client/consumer**, not a server. **Recommendation**: Skip unless we ship a dedicated MCP server package.

### 5. `awesome-ai-coding` — CONFIRMED

| Repo | Stars | Last Push | Fit | Priority |
|------|------:|-----------|-----|:--------:|
| `wsxiaoys/awesome-ai-coding` | 769 | 2026-03-02 | Maintained by ex-Tabby/TabNine founder — strong curator reputation | **P1** |
| `ai-for-developers/awesome-ai-coding-tools` | — | 2026-04-22 | Generic AI coding tools | P2 |
| `inmve/awesome-ai-coding-techniques` | — | 2026-04-20 | "Claude Code" listed explicitly in description — perfect fit | P2 |

---

## Bonus High-Impact Targets (newly discovered)

These were not in the original candidate list but surfaced during verification and rank above most original candidates:

### BONUS-1: `VoltAgent/awesome-claude-code-subagents` — **17,962 stars**, pushed 2026-04-20

- Description: "A collection of 100+ specialized Claude Code subagents covering a wide range of development use cases"
- **Direct fit**: MoAI has 24 specialized subagents (expert-*, manager-*, builder-*)
- **Priority**: P1 (second flagship)
- **Entry angle**: "MoAI-ADK — SPEC-first orchestration layer coordinating 24 subagents across 8 domains"

### BONUS-2: `travisvn/awesome-claude-skills` — **11,615 stars**, pushed 2026-03-16

- Description: "A curated list of awesome Claude Skills, resources, and tools for customizing Claude AI workflows — particularly Claude Code"
- **Direct fit**: MoAI has 52 skills (moai-*, domain skills, workflow skills)
- **Priority**: P1 (third flagship)
- **Caveat**: 347 open issues — slightly slower review cadence, expect longer PR tail

### BONUS-3: `rohitg00/awesome-claude-code-toolkit` — pushed 2026-04-22

- Description: "135 agents, 35 curated skills, 42 commands, 176+ plugins, 20 hooks, 15 rules"
- **Angle**: We're an example of a full toolkit (agents + skills + commands + hooks + rules)
- **Priority**: P2

### BONUS-4: `inmve/awesome-ai-coding-techniques` — pushed 2026-04-20

- Description: Mentions Claude Code, Codex CLI, Cursor, Copilot explicitly. Multi-language list (en/es/de/fr/ja).
- **Angle**: SPEC-First methodology entry under "Techniques"
- **Priority**: P2

---

## Updated P1 Submission Queue (7 targets, by expected impact)

Ordered by expected referral traffic (stars × topical fit):

| Rank | Repo | Stars | Primary Angle |
|:----:|------|------:|---------------|
| 1 | `hesreallyhim/awesome-claude-code` | 40,283 | agent orchestrators for Claude Code |
| 2 | `avelino/awesome-go` | ~147k | Go CLI tool under Command Line section |
| 3 | `VoltAgent/awesome-claude-code-subagents` | 17,962 | 24 specialized subagents |
| 4 | `travisvn/awesome-claude-skills` | 11,615 | 52 skills with Progressive Disclosure |
| 5 | `e2b-dev/awesome-ai-agents` | — | multi-agent orchestration framework |
| 6 | `wsxiaoys/awesome-ai-coding` | 769 | SPEC-first AI coding methodology |
| 7 | `jqueryscript/awesome-claude-code` | 284 | frameworks for Claude Code developers |

**P2 queue (follow-up after P1 merges, 4 targets)**:
- `ai-for-developers/awesome-ai-coding-tools`
- `inmve/awesome-ai-coding-techniques`
- `rohitg00/awesome-claude-code-toolkit`
- `ccplugins/awesome-claude-code-plugins`

**SKIP permanently (3 original candidates)**:
- `awesome-llm-tools` — no match
- `awesome-developer-tools` — no match
- `awesome-mcp-*` — wrong fit (we're not an MCP server)

---

## Submission Guidance Updates

For each new P1 target (ranks 1, 3, 4, 6, 7), before opening a PR:

1. **Read the target repo's CONTRIBUTING / README rules** — each has different sort order (alphabetical vs chronological), entry format (one-liner vs expanded), and required fields (tags, categories).
2. **Pick the correct section** — for `hesreallyhim/awesome-claude-code`, the most likely section is "Agent Orchestrators" or "Frameworks".
3. **Draft the one-liner** in the list's exact style. Below are first-pass drafts; refine by inspecting 3 adjacent entries before submitting:

```markdown
# For hesreallyhim/awesome-claude-code (Agent Orchestrators section):
- [MoAI-ADK](https://github.com/modu-ai/moai-adk) - SPEC-first orchestration framework with 24 specialized agents + 52 skills, enforced Plan→Run→Sync pipeline, TRUST 5 quality gates. Go CLI, 16 language projects, 4 documentation languages (en/ko/ja/zh).

# For VoltAgent/awesome-claude-code-subagents (Framework section):
- [MoAI-ADK](https://github.com/modu-ai/moai-adk) - SPEC-first framework orchestrating 24 specialized subagents (expert-backend, expert-frontend, expert-security, manager-spec, manager-ddd, manager-tdd, etc.) with automatic task decomposition and parallel execution via Agent Teams mode.

# For travisvn/awesome-claude-skills (Skills Collection section):
- [MoAI-ADK](https://github.com/modu-ai/moai-adk) - Production-grade skill library (52 skills) with 3-level Progressive Disclosure (metadata/body/bundled). Includes moai-domain-*, moai-workflow-*, moai-foundation-*, moai-tool-*, and moai-platform-* skill families. 4-language documentation.

# For wsxiaoys/awesome-ai-coding (Methodology section):
- [MoAI-ADK](https://github.com/modu-ai/moai-adk) - SPEC-first methodology + framework for Claude Code. EARS-format requirements, enforced Plan→Run→Sync pipeline, TDD/DDD auto-selection, TRUST 5 quality gates.
```

---

## Next Action

Update `submission-plan.md` to reflect these verification results (add 4 bonus targets, remove 3 skip-candidates, re-rank by impact), OR treat this file as the authoritative source and reference it from submission-plan.md.

**Recommended**: Keep this file as the verification audit trail. Revise `submission-plan.md` to index into it.

---

**Status**: 7 P1 targets confirmed (up from 2), 4 P2 targets identified, 3 original candidates skipped with rationale. Ready for PR drafting phase.
