---
description: "Detail companion for moai-constitution.md — the Opus 5 / 4.8 prompt-philosophy guidance and the full Lessons Protocol operational detail (topic-file store, harness edit discipline, auto-capture triggers, domain matching, integration points)"
paths: "**/moai-constitution.md,**/agent-authoring.md,**/.claude/skills/moai/workflows/*.md"
---

# MoAI Constitution — Detail Companion

> Detail companion of `moai-constitution.md` (the always-loaded stub). The stub owns every core
> principle and the six Agent Core Behaviors. This file owns the model-specific prompt guidance and
> the Lessons Protocol's operational machinery. Load it when authoring an agent prompt for a
> specific model tier, or when capturing, draining, or acting on a lesson.

## Opus 5 / 4.8 Prompt Philosophy

The binding bullet list lives in the stub. This section carries only what does not belong on the
always-loaded surface.

**Source.** Anthropic's model guidance at
`platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-8`.

**Why the 4.6-era scaffolding became counterproductive.** Opus 4.7+ follows instructions literally.
Phrases like "double-check X before returning" or "verify N times" were written for a model that
generalized loosely; under literal following they add turns without adding checking, and they crowd
out the specific instruction that would actually bind. The same shift is why scope has to be stated
rather than implied — "apply to every section" is now load-bearing text, not emphasis.

**Why fan-out and tool use became steerable rather than automatic.** The two default-reduction
behaviors (fewer subagents, fewer tool calls) trade breadth for reasoning depth. Neither is a
capability loss: an explicit instruction restores the behavior, and raising effort to `high` or
`xhigh` shifts the balance back toward tool use. The practical consequence is that a prompt relying
on implicit fan-out silently gets a single-agent answer.

**Effort routing by role.** Route effort by what the work demands rather than by which named agent
is running: coding and agentic work at `xhigh`, intelligence-sensitive work at `high` or above,
speed-critical or mechanical work stepped down. The per-agent default table lives in
`.claude/rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix, alongside the
archived-agent legacy reference.

## Lessons Protocol

Capture and reuse learnings from user corrections and agent failures across sessions.

Rules:
- When user corrects agent behavior, capture the pattern in auto-memory
- Store lessons as topic files in auto-memory — one fact per `feedback_*.md` file under `~/.claude/projects/{project-hash}/memory/`, indexed by `MEMORY.md`. This topic-file convention (`feedback_*.md` topic files + the `MEMORY.md` index) is the single designated lesson store; the legacy `lessons.md` is superseded (kept on disk marked `[SUPERSEDED]`, content not migrated)
- Each lesson entry: category, incorrect pattern, correct approach, date added
- Review relevant lessons before starting tasks in the same domain
- Lesson categories: architecture, testing, naming, workflow, security, performance, hardcoding
- Maximum 50 active topic files per project; archive older or superseded topic files into `memory/_archive/` (never delete — archive preserves the audit trail)
- Lessons are additive: never overwrite a lesson, append corrections as updates
- To supersede a lesson, add `[SUPERSEDED by #{new_lesson_number}]` prefix to the old entry
- Session start: scan lessons for patterns matching current task domain
- Repo-local lessons inbox (`.moai/lessons-inbox.jsonl`): tool failures and test failures append structured stubs (timestamp, event_key, summary, source) here as they occur — the two wired families are `tool_failure:<tool>:<sig>` and `test_fail:<pkg>:`. **Capture scope (capability + composition):** the inbox records failure-event stubs only; it is not a record of defect families that produce neither a tool failure nor a test failure (vacuous green checks, skip-before-verdict, empty-result-set pass conditions, sibling repair misses, moving-ref assertions, stale-value quotations). The human-mediated loop (lane discovery → lead judgment → auto-memory `feedback_*.md` + `MEMORY.md` record) is the learning channel for those tool-invisible families; the measured live composition and its dated baseline live in the learning-channel scope anchor document, not in prose. **Drain actor: the MoAI orchestrator. Drain trigger: when the inbox backlog grows large enough to obscure recurring patterns (a cluster of same-`event_key` stubs), the orchestrator drains these stubs into topic-file lesson entries as part of the Lessons Protocol — converting each recurring `event_key` cluster into one candidate `feedback_*.md` topic file before human review, and discarding one-off noise (single-occurrence stubs with no recurring pattern). Drained stubs are marked (the drain-marking mechanism is an implementation detail).**

Harness Edit Discipline (decision observability):
- Harness surface tag: each lesson entry SHOULD carry a `surface:` tag naming the harness component it binds to (rule / agent / skill / hook / config / template / workflow) — enables clustering recurring failures by component
- Prediction pairing: when a lesson motivates a harness edit (a change to a rule, agent, skill, hook, config, or template), the lesson entry SHOULD record `prediction:` — the falsifiable expected effect (which failure class stops recurring) — and later `verified: true|false` with the observed evidence
- Held-in / held-out acceptance: before accepting a harness edit, verify BOTH (a) held-in — the edit demonstrably addresses the motivating failure (reproduce or cite the failing case), and (b) held-out — existing guards still pass (lint, mirror-parity, neutrality checks, test suite). An edit failing either check is rejected, not merged
- Preserve rejected candidates: a rejected or reverted harness edit is recorded as a lesson entry with `verified: false` and the rejection reason — never silently discarded. This prevents re-attempting known-bad edits
- A falsified prediction (`verified: false`) is itself a signal: re-diagnose the root cause before authoring a second edit to the same surface

Auto-Capture Triggers:
- When a fix/refactor commit completes, check if the change matches a known anti-pattern category
- If match found, propose a lesson entry to the user via AskUserQuestion
- Auto-generated lesson entries include: category, incorrect pattern, correct approach, date, tags
- Duplicate detection: check existing lessons before proposing new entry

Domain Matching Algorithm:
- Extract domain keywords from current SPEC (title, scope, modified file paths)
- Match lesson categories against extracted keywords
- Match lesson tags against modified package names
- Relevance score: categories match (weight 2) + tags match (weight 1)
- Select top 5 lessons by relevance score, then by recency

Integration Points:
- run.md Phase 1: Load filtered lessons into agent context before implementation (see Lessons Loading section)
- /moai fix completion: Propose lesson capture after successful fix
- /moai loop completion: Propose lesson capture after successful iteration cycle

<!-- moai:evolvable-start id="agent-core-behaviors" -->

---

Classification: Lazy companion — model-tier guidance and lesson-handling procedure only. Every core
principle and every Agent Core Behavior stays in `moai-constitution.md`.
