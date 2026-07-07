---
id: SPEC-HOOK-CONFIG-SAFETY-001
title: "moai update Configuration Safety /hooks Review Guidance"
version: "0.1.0"
status: in-progress
created: 2026-07-07
updated: 2026-07-07
author: manager-spec (P5 of hooks-hardening Epic)
priority: P3
tier: S
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hooks, cli, configuration-safety, ux, discoverability"
---

# SPEC-HOOK-CONFIG-SAFETY-001 — moai update Configuration Safety /hooks Review Guidance

> GEARS notation (canonical since v3.0.0). Tier S — see §C.1. P5 of the moai-adk hooks-hardening Epic (P1 = SPEC-HOOK-SESSIONSTART-PROBE-001 completed; P2/P3/P4 absorbed as stale-memory non-defects).

## §A Overview

### §A.1 Problem

Claude Code (CC) captures a snapshot of hooks at session startup and uses that snapshot throughout the session (CC official policy, "Configuration Safety", https://docs.claude.com/en/docs/claude-code/hooks § Configuration Safety, fetched via webReader this session). When `moai update` re-renders `.claude/settings.json` hooks section via template deployment + 3-way merge (per `internal/cli/update.go` "Template deployment uses a 3-way merge strategy to preserve local modifications"), any running CC session's snapshot becomes stale — new/changed hooks do NOT take effect until the user reviews the `/hooks` menu or restarts CC.

CC auto-warns on external modification (policy point 3), so this is NOT a functional defect. However, `moai update` completion messages currently give ZERO guidance about this, leaving the user to discover the stale-snapshot situation via the generic CC warning (which does not point at the `moai update` cause). This is a low-severity UX/discoverability gap, especially for the maintainer's frequent pattern of running `moai update` mid-CC-session.

### §A.2 Verified Defect Evidence (grep-authoritative, not memory-derived)

Per `feedback_hypothesis_as_defect` and `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3`, every "guidance is absent" claim in this SPEC is backed by a mechanical grep command and its observed output, not by memory or inference. The orchestrator re-ran all three greps independently during plan-phase; results below are the observed outputs.

**Grep 1 (broad — code + docs surfaces)**:
```
grep -rn -E '/hooks[^/]|hooks menu|configuration safety|startup snapshot|snapshot of hooks' internal/cli/ cmd/ README.md README.ko.md
```
Result: **0 hits** in user-facing output. The single indirect match is `internal/cli/update.go:2833` — a comment about removing the global `~/.claude/hooks/moai/` directory on uninstall, NOT a `/hooks` menu reference.

**Grep 2 (init.go + update.go narrow)**:
```
grep -n -E '/hooks|hook menu|hook review' internal/cli/init.go internal/cli/update.go
```
Result: **0 hits** in `init.go` and `update.go`. (Adjacent matches in `internal/cli/hook_install_precommit_test.go` are about git pre-commit hooks, NOT the CC `/hooks` menu — distinct concept.)

**Grep 3 (rules + docs canonical surfaces)**:
```
grep -rn -i 'configuration safety|/hooks menu|hooks review|hook snapshot|settings.json snapshot' .claude/rules/ .moai/docs/
```
Result: **0 hits** — no canonical rule or doc currently mirrors the CC Configuration Safety policy.

**Insertion-point verification**: `internal/cli/update.go` line ~874 carries the canonical "Template sync complete" branch (content token `Label: "Template sync complete"`). Distinct no-op branches exist at `Label: "Already up to date"` (~line 239), `Label: "Template version up-to-date · Skipping sync"` (~lines 534, 927), and `Label: "Binary updated (template sync skipped)"` (~line 228). The init.go success card at `init.go:434-444` renders `Directories N created` + `Files N created` + Warnings — NO hook snapshot guidance.

### §A.3 Why P5 is distinct from P2/P3/P4 (Epic context)

P2/P3/P4 of this Epic were ABSORBED without a SPEC because the memory defect-claim was REFUTED at code level (P2=already-shipped enum fix in 1169d2afb; P3=hash-keyed traversal-immune; P4=already-documented in tech.md). P5 is the OPPOSITE: the memory claim ("guidance absent") is CONFIRMED at code level (Greps 1-3 in §A.2), AND the CC official policy JUSTIFIES adding the guidance. This is a real low-severity UX gap, NOT a stale-memory phantom. The verify-before-author pipeline (`feedback_hypothesis_as_defect`) gate this SPEC's right to exist.

## §B Requirements (GEARS notation)

### §B.1 Primary requirement — guidance emission

**When** the `moai update` command completes the "Template sync complete" branch (`internal/cli/update.go`, located by content token `Label: "Template sync complete"`), the `moai update` completion output **shall** include a guidance line that mentions `/hooks` verbatim and conveys that running CC sessions require `/hooks` review (or a CC restart) for the new/changed hooks to take effect.

### §B.2 No-spurious-fire requirement — no-op paths

**When** the `moai update` command takes any no-op branch (`Label: "Already up to date"`, `Label: "Template version up-to-date · Skipping sync"`, OR `Label: "Binary updated (template sync skipped)"`), the `moai update` completion output **shall not** include the `/hooks` review guidance — because no template re-render occurred and no running CC session's snapshot could be stale.

### §B.3 Actionability requirement — literal `/hooks` token

**Where** the `/hooks` review guidance is emitted, the guidance string **shall** include the literal token `/hooks` (the CC menu name) so the user can act on it without consulting external documentation.

### §B.4 No subagent boundary violation

**Where** the guidance implementation lives in `internal/cli/update.go`, the implementation **shall not** invoke `AskUserQuestion`, `mcp__askuser__*`, or any other interactive prompt. Per `internal/cli/CLAUDE.md` (C-HRA-008 / REQ-PGN-012), CLI code runs in subagent context and the orchestrator owns user interaction — the guidance is a stdout message only.

### §B.5 Characterization test requirement

The implementation **shall** be covered by a Go characterization test (DDD per `moai-workflow-ddd`) that asserts:
- (a) the literal substring `/hooks` is present in completion output when the "Template sync complete" branch runs; AND
- (b) the literal substring `/hooks` is absent in completion output on each of the three no-op branches named in §B.2.

## §C Constraints (Non-functional)

### §C.1 Tier S — minimal scope
- Files affected: 2 (internal/cli/update.go edit + internal/cli/update_test.go characterization test). Optionally 3 if a separate static-guard test file is added for AC-CS-004.
- No new dependencies, no new packages, no architectural change.
- Single domain (CLI completion UX).
- Tier S justified per `moai-constitution.md § Agent Core Behaviors #4 Enforce Simplicity` — reach for the cheapest sufficient capability.

### §C.2 CC policy alignment, not re-implementation
- The guidance MUST cite the CC menu name (`/hooks`) verbatim.
- The guidance MUST NOT attempt to re-implement CC's snapshot mechanism or trigger snapshot refresh programmatically (Out of Scope — see §D).
- The guidance is ADVISORY; CC's auto-warn on external modification remains the functional safety net.

### §C.3 Simplicity ladder compliance
Per `moai-constitution.md § Agent Core Behaviors #4 Enforce Simplicity`, this SPEC MUST be implementable using the simplest trigger strategy that covers the primary use case (maintainer's frequent mid-session `moai update`). JSON-diff-based conditional emission (Option A in plan.md §A) is over-engineering for a UX hint and is rejected.

### §C.4 Test isolation (CLAUDE.local.md §6)
Any test temp directories MUST use `t.TempDir()`. The characterization tests MUST NOT modify the project root or `~/.claude/settings.json`. Path handling MUST use `filepath.Abs()` for user-supplied paths (not `filepath.Join(cwd, absPath)` which silently produces wrong paths on macOS `/var/folders/...` tempdirs).

## §D Out of Scope

### Out of Scope — init-time guidance
- Adding `/hooks` review guidance to `moai init` completion (`internal/cli/init.go:434-444` success card) is OUT OF SCOPE. Rationale: `moai init` typically runs in a terminal BEFORE a CC session starts, so CC captures the new snapshot at session start — no stale-snapshot risk in the common workflow. The edge case (running `moai init` from inside an existing CC session via Bash tool) is rarer and deferred to a follow-up SPEC if user feedback indicates confusion.

### Out of Scope — programmatic snapshot refresh
- Triggering CC's `/hooks` snapshot refresh programmatically (via IPC, signal, hook invocation, or any other mechanism) is OUT OF SCOPE. CC owns the snapshot lifecycle; this SPEC only emits advisory guidance pointing at the `/hooks` menu.

### Out of Scope — JSON-diff-based conditional emission (Option A)
- Implementing hook-section diff detection to emit guidance ONLY when the hooks JSON actually changed is OUT OF SCOPE. The unconditional-on-template-sync trigger (Option B, see plan.md §A) is sufficient for the UX goal; Option A's precision is gold-plating for a Tier S hint.

### Out of Scope — i18n / multi-locale guidance variants
- Translating the guidance string into the 16 supported project languages is OUT OF SCOPE. The guidance follows the existing `moai update` pill convention (English-only labels at the existing `Label:` sites). i18n can be added in a follow-up if and when the broader update-flow pills gain i18n support.

### Out of Scope — `--json` machine-readable output integration
- If `moai update --json` exists, deciding how the guidance interacts with JSON output is deferred to run-phase characterization (see acceptance.md §D.3). This SPEC does not pin a JSON schema for the guidance.

## §E Rationale

### §E.1 Why this is a SPEC (not just a one-line fix)

The defect claim ("guidance absent") was a hypothesis until verified. Per `feedback_hypothesis_as_defect` and `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3`, the hypothesis-to-SPEC pipeline required:
1. **Mechanical grep evidence** (Greps 1-3 in §A.2) — DONE, 0 hits confirmed by independent re-run during plan-phase, not by memory.
2. **CC official policy citation** (Configuration Safety, https://docs.claude.com/en/docs/claude-code/hooks) — DONE, fetched via webReader by the orchestrator this session.
3. **Distinct-from-P2/P3/P4 demonstration** (§A.3) — DONE, the memory claim is CONFIRMED (not refuted) at code level, the opposite of the P2/P3/P4 pattern.

Without this evidence trail, this SPEC would risk being a stale-memory phantom (the same failure mode as P2/P3/P4). The evidence trail is the rationale for SPEC authoring rather than direct absorption into P1.

### §E.2 Why Option B (unconditional on template-sync-complete) is the recommended trigger

See plan.md §A for the full Option A vs B vs C analysis. Short form: Option B is the cheapest sufficient capability on the simplicity ladder (`moai-constitution.md § Agent Core Behaviors #4`). It covers the primary maintainer use case (mid-session `moai update`), does not fire spuriously on no-op paths (satisfies AC-CS-002), and avoids the JSON-diff complexity of Option A. Init-time coverage (Option C) is deferred because the common `moai init` workflow has no stale-snapshot risk.

### §E.3 Why Tier S (not Tier M or L)

Tier S is appropriate because:
- Single domain (CLI completion UX).
- 2-3 files affected.
- No architectural change, no new dependencies, no new packages.
- No cross-cutting concerns (the change is localized to `moai update` output).

Tier M would require propagation across multiple commands or new infrastructure; Tier L would require coordinated changes across CLI + hooks + documentation subsystems. Neither fits this SPEC's actual scope.

### §E.4 Severity classification: P3 (Low)

Priority P3 because:
- CC auto-warns on external modification — the functional safety net already exists.
- This SPEC improves discoverability of the `/hooks` review action; it does NOT fix a functional defect.
- No data loss, no security implication, no production-incident risk.
- User can recover without the guidance (just slower — via the generic CC auto-warn).

## §F Cross-references

- **CC Configuration Safety policy**: https://docs.claude.com/en/docs/claude-code/hooks § Configuration Safety (fetched via webReader this session by the orchestrator; the policy states verbatim "1) Captures a snapshot of hooks at startup 2) Uses this snapshot throughout the session 3) Warns if hooks are modified externally 4) Requires review in `/hooks` menu for changes to apply")
- **Epic context**: P5 of moai-adk hooks-hardening Epic. P1 = SPEC-HOOK-SESSIONSTART-PROBE-001 (completed, sync 10cd484a0). P2/P3/P4 absorbed (3x stale-memory non-defects). P5 = this SPEC (real gap confirmed).
- **Verification doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 3` (defect/debt/drift identification claim — the binding surface that requires mechanical-tool verification, not text-pattern inference).
- **Hypothesis-as-defect lesson**: `feedback_hypothesis_as_defect` (P2/P3/P4 3x consecutive stale non-defects — verify-before-author is mandatory).
- **Defect-claim verification lesson**: `feedback_defect_claim_verification` (defect claim is a hypothesis until the domain's tool confirms it).
- **Simplicity ladder**: `.claude/rules/moai/core/moai-constitution.md § Agent Core Behaviors #4 Enforce Simplicity` (reach for the cheapest sufficient capability).
- **Subagent boundary**: `internal/cli/CLAUDE.md` (C-HRA-008 / REQ-PGN-012 — CLI code MUST NOT call AskUserQuestion).
- **Test isolation**: CLAUDE.local.md §6 (`t.TempDir()` mandatory; `filepath.Abs()` for user-supplied paths).
- **Insertion point**: `internal/cli/update.go` `Label: "Template sync complete"` branch (~line 874 at grep date 2026-07-07; run-phase must locate by content token per `feedback_line_number_drift_asymmetry`).

## §G HISTORY

- **2026-07-07** — Initial draft (manager-spec, plan-phase). P5 of hooks-hardening Epic. Verify-before-author COMPLETE: defect confirmed via independent re-run of Greps 1-3 (0 hits across `internal/cli/`, `cmd/`, `README.md`/`README.ko.md`, `.claude/rules/`, `.moai/docs/`). SPEC ID pre-write self-check PASS: `SPEC-HOOK-CONFIG-SAFETY-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`.
