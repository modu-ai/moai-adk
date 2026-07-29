---
id: SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001
title: "Pre-Edit Parallel-Session Detection & Auto-Isolation"
version: "0.1.0"
status: completed
created: 2026-07-28
updated: 2026-07-29
author: GOOS
priority: P1
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tags: "parallel-session, pre-edit-sync-check, hook, advisory"
era: V3R6
tier: M
---

# SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001 — Pre-Edit Parallel-Session Detection & Auto-Isolation

## A. Problem (observed incident)

The Pre-Spawn Sync Check (`.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check) detects parallel-session races **only at write-agent spawn boundaries**. Direct orchestrator edits to shared working-tree paths (Edit/Write/Bash in the main session — common under MoAI-Easy hands-on and direct main-session work) **bypass this gate**, so foreign active sessions go undetected mid-task.

Observed failure (2026-07-29, session c598f065): the orchestrator staged 12 language-policy files in the primary checkout and paused to prepare a commit. A concurrent session on the same checkout ran a broad `git add -A && commit`, sweeping the orchestrator's staged files into its own commit; that PR auto-merged, landing the orchestrator's uncommitted work on `origin/main` entangled inside an unrelated CI-flaky commit (`05117b6ba`). The orchestrator never spawned a write agent, so no Pre-Spawn Sync Check ran; and the auto-isolation procedure (`.claude/rules/moai/workflow/worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation) additionally requires "worktree entry is chosen", which is false for direct primary edits.

Root cause: **parallel-session detection is spawn-gated, but direct main-session editing bypasses the spawn gate.** On a repo with auto-merge, uncommitted ≠ undeployed.

## B. Scope

### B.1 In scope
- **REQ-PES-001 — Pre-Edit Sync Check (doctrine)**: a new section in `agent-common-protocol.md` specifying that the orchestrator runs a parallel-session detection check before **non-trivial direct edits** to shared working-tree paths (not only before write-agent spawns). The check mirrors the Pre-Spawn Sync Check (`moai session list --json` own-session-filtered + `git fetch` + `git rev-list --left-right origin/main...HEAD`).
- **REQ-PES-002 — auto-isolation trigger broadening**: `worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation trigger relaxes the "worktree entry is chosen" conjunct so that a foreign active session on the same checkout during **any write work (spawn OR direct edit)** triggers isolation/surfacing.
- **REQ-PES-003 — bypass failure-mode codification**: the "direct edit bypasses the spawn gate" failure mode is named explicitly in doctrine (cross-referenced from the Pre-Spawn Sync Check and the auto-isolation section).
- **REQ-PES-004 — PreToolUse-on-Edit hook evaluation**: a mechanical PreToolUse hook on Edit/Write that consults `active-sessions.json` is **evaluated** for cost/latency. If the per-edit overhead is unacceptable, the hook is **deferred** with recorded rationale and the procedural path (REQ-PES-001) remains the enforcement.
- **REQ-PES-005 — advisory ambient signal**: a cheap read-only ambient signal (e.g., session-start or statusline surfaces the count of foreign active sessions) so the orchestrator has ambient awareness without a per-edit cost.

### B.2 Out of scope
- A full blocking PreToolUse-on-Edit/Write hook if REQ-PES-004 evaluates it as too costly (the evaluation + rationale is in scope; the implementation may be deferred).
- Changes to the auto-merge bot policy or branch-protection settings (orthogonal).
- Rewriting the already-landed `05117b6ba` history (content is correct; attribution is cosmetic).

### B.3 Mechanism (design)
- **Primary enforcement**: procedural — the Pre-Edit Sync Check (orchestrator-run), mirroring the proven Pre-Spawn pattern. Lowest risk; coverage depends on the orchestrator following the procedure (same trust model as Pre-Spawn).
- **Advisory**: ambient foreign-session-count signal (read `active-sessions.json`, no fetch).
- **Mechanical (evaluated)**: PreToolUse-on-Edit hook — likely deferred on cost grounds; rationale recorded. This is the honest "is a hook worth it?" decision, not a foregone conclusion.

## C. Requirements (GEARS)

- **REQ-PES-001** [HARD]: The orchestrator MUST run a parallel-session detection check (`moai session list` + divergence) before non-trivial direct edits to shared working-tree paths, not only before write-agent spawns.
- **REQ-PES-002** [HARD]: When ≥1 foreign active session is detected during direct-edit work, the orchestrator SHALL isolate (worktree) OR surface via AskUserQuestion (isolate / wait / abort), reusing the Pre-Spawn Sync Check interpretation matrix.
- **REQ-PES-003** [HARD]: The auto-isolation procedure's trigger condition MUST include direct-edit work, not only "worktree entry chosen".
- **REQ-PES-004** [SHOULD]: The PreToolUse-on-Edit hook option MUST be evaluated with a recorded cost/latency finding and an explicit defer-or-implement decision.
- **REQ-PES-005** [SHOULD]: An ambient foreign-session signal SHOULD surface at session start / in the statusline.

## D. "Non-trivial direct edit" (REQ-PES-001 trigger definition)

An Edit/Write/Bash mutation touching shared working-tree paths under `.claude/`, `.moai/`, `internal/`, `pkg/`, `cmd/`, or repo-root config — paths another session could also mutate. Exempt: edits under an already-isolated worktree, `/tmp`, or a session-private scratch dir.

## E. Risks
- R1: Procedural enforcement depends on orchestrator compliance (not mechanical). Mitigation: ambient signal (REQ-PES-005) + explicit failure-mode naming (REQ-PES-003); the hook (REQ-PES-004) is the mechanical backstop.
- R2: A PreToolUse-on-Edit hook could add latency to every edit. Mitigation: evaluate first; defer if costly.
- R3: Stale-registry false positives (dead-PID entries) could trigger unnecessary isolation. Mitigation: reuse the existing conservative predicate + the stale-false-positive disposition already documented in the auto-isolation section.
