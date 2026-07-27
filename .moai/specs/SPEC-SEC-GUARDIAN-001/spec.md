---
id: SPEC-SEC-GUARDIAN-001
title: "In-session always-on 3-layer security guardian (pattern warnings + turn-diff review + commit-time cross-file review) — Epic SECURITY-ABSORB SPEC-2"
version: "0.1.0"
status: completed
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/hook/security (Go scanner + handlers) + internal/cli/hook.go (subcommand wiring) + internal/template/templates/.claude/hooks/moai (3 shell wrappers) + settings.json.tmpl + local .claude siblings"
lifecycle: spec-anchored
tags: "security, guardian, security-absorb, hooks, posttooluse, stop, pattern-warnings, owasp, 16-language, template, go"
tier: L
era: V3R6
related_specs: [SPEC-SEC-DEEPSCAN-001, SPEC-OBSERVE-HYGIENE-001, SPEC-HOOK-OFFICIAL-COMPLIANCE-001]
---

# SPEC-SEC-GUARDIAN-001 — In-session always-on 3-layer security guardian

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Plan-phase artifact set authored (Tier L, 5 artifacts: spec / plan / acceptance / design / research + progress skeleton). Epic SECURITY-ABSORB, **SPEC-2** of the cohort. **Absorption source**: Anthropic's official `security-guidance` Claude Code plugin architecture — reimplemented NATIVELY into MoAI's own harness (NOT installed as a third-party plugin). **Candidate B** = the IN-SESSION, always-on 3-layer security guardian that reviews code AS it is written. **Complement to SPEC-1**: SPEC-1 (SPEC-SEC-DEEPSCAN-001) is the heavy, explicitly-invoked, on-demand deep scan (`/moai review --deep`); THIS SPEC is the light, always-on, in-session guardian wired into MoAI's shell-hook system. The two compose without overlap (§C Out of Scope enumerates the SPEC-1 boundary). **Distribution shape**: Go pattern-engine + 3 hook handlers compiled into the binary (tested), plus template-first shell wrappers + settings.json hook wiring (16-language-neutral). |

---

## §A Context & Problem

### A.1 Motivation

Anthropic ships an official `security-guidance` Claude Code plugin whose most-used surface is an **in-session, always-on** security guardian: it reviews code AS the assistant writes it, complementary to any explicit deep scan. That guardian is three layers:

1. **Instant pattern warnings** on every Edit/Write (PostToolUse) — a fast, regex-based flag on ~25 known-dangerous patterns, no LLM in the loop, so it never stalls routine editing.
2. **Turn-end diff review** (Stop) — when the assistant finishes a turn, the turn's diff is reviewed and high-severity findings are surfaced back into the session so they can be fixed before the user ever sees the result.
3. **Commit-time cross-file review** — on git commit, a reviewer reads related files (Read/Grep/Glob) to trace data flow ACROSS files, catching multi-file vulnerabilities (IDOR, auth bypass, cross-file SSRF) that single-file pattern-matching misses.

MoAI already owns the substrate to reproduce this natively: a mature shell-hook system (`.claude/hooks/moai/*.sh` wrappers forwarding to `moai hook <event>`), two integration-precedent gate hooks (`sync-phase-quality-gate.sh` Stop gate + `status-transition-ownership.sh` PostToolUse gate), settings.json hook-wiring conventions, an opt-in blocking pattern (`MOAI_SYNC_GATE_BLOCKING`), a proven in-wrapper regex-heuristic precedent (the Bash Risk-Amplifier warn signal in `handle-pre-tool.sh`), and four production security reference skills for pattern vocabulary. What MoAI does NOT yet have is a native, always-on in-session security guardian composing these into the three layers above.

### A.2 The gap this SPEC closes

Today a user who edits code inside a MoAI session gets NO in-session security feedback as they write. The only MoAI-native security surface is the on-demand `/moai review --security` lens (single-pass, explicit) and — after SPEC-1 lands — `/moai review --deep` (heavy, explicit). Both are user-initiated and after-the-fact. There is no lightweight, continuous, always-on guardian that flags an obvious `yaml.load` on untrusted input, a raw `innerHTML` assignment, or a hardcoded secret at the moment the Edit lands. This SPEC adds that guardian.

### A.3 What is absorbed (source architecture, restated in MoAI terms)

The absorbed source architecture, reproduced natively as three layers wired into MoAI's shell-hook system:

| Layer | Absorbed behavior | MoAI realization | Default posture |
|-------|-------------------|------------------|-----------------|
| **L1 — pattern warnings** | Instant regex warnings on Edit/Write for ~25 known-dangerous patterns | PostToolUse hook (`Write\|Edit\|MultiEdit`) → `moai hook` Go handler → single-source pattern table organized by vulnerability CLASS across the 16 supported languages | ON (lightweight, regex-only, async) |
| **L2 — turn-end diff review** | Review the finished turn's diff; surface high-severity findings back into the session | Stop hook → Go handler runs the L1 pattern engine over the turn's working-tree diff; advisory `systemMessage` | ON (regex-only default; LLM escalation opt-in) |
| **L3 — commit-time cross-file review** | On git commit, a reviewer reads related files to trace data flow across files | A sibling commit-time Stop hook extending the sync-gate model with a security axis (surface L3-A), emitting a structured escalation signal the orchestrator translates into an `Agent()` cross-file data-flow review | OPT-IN (off by default; LLM per-commit is expensive) |

### A.4 MoAI assets reused (NOT reinvented)

| Absorbed concept | MoAI asset reused |
|------------------|-------------------|
| Hook plumbing | The existing `.claude/hooks/moai/handle-*.sh` wrapper pattern (3-tier `moai` binary resolution + silent fail-open) — `hook-independence.md` §5 |
| PostToolUse regex heuristic in a wrapper | The Bash Risk-Amplifier warn-signal precedent in `handle-pre-tool.sh` (inline regex, WARN-only, fail-open) |
| Commit-time / turn-end gate model | `sync-phase-quality-gate.sh` (Stop gate, commit-subject-aware, once-per-commit sentinel, language-neutral detection) |
| Advisory-by-default + opt-in blocking | The `MOAI_SYNC_GATE_BLOCKING` opt-out env pattern + the exit-code + stdout-JSON hook contract (`agent-common-protocol.md` § Hook Invocation Surface) |
| Hook → orchestrator translation of a block | The orchestrator-translation-responsibility boundary (hooks emit JSON, orchestrator runs AskUserQuestion) — `agent-common-protocol.md` § Hook Invocation Surface |
| Pattern vocabulary | Existing security reference skills: `moai-ref-owasp-checklist`, `moai-ref-llm-security`, `moai-ref-secops`, `moai-ref-supply-chain` (source of the vulnerability-class taxonomy; NOT shipped patterns) |
| Go subcommand plumbing | `moai hook <event>` dispatch in `internal/cli/hook.go` (precedent: `db-schema-sync`, `spec-status`, `harness-classify` subcommands) |
| Template distribution | Template-First shell wrappers + `settings.json.tmpl` wiring (16-language-neutral), byte-lockstep local sync |

### A.5 Non-goals (see §C for the formal exclusions)

This SPEC does NOT re-implement SPEC-1's on-demand multi-agent deep scan, does NOT add a new agent, does NOT ship a `.js` workflow script, and does NOT block routine editing by default. Layers 1-2 are advisory; Layer 3 is opt-in. The heavy path stays in SPEC-1.

---

## §B Requirements (GEARS)

### Surface & layering

**REQ-SG-001 — Native 3-layer in-session guardian (Ubiquitous)**
The security guardian **shall** be reimplemented natively inside MoAI's own harness (shell-hook system + compiled Go handlers) as three always-available layers — pattern warnings (L1), turn-end diff review (L2), and commit-time cross-file review (L3) — and **shall not** be installed as, or depend on, a third-party Claude Code plugin.

**REQ-SG-002 — No overlap with the on-demand deep scan (Unwanted behavior)**
The guardian **shall not** re-implement, duplicate, or subsume SPEC-SEC-DEEPSCAN-001's on-demand multi-agent deep scan (six-phase pipeline, adversarial 3-voter panel, reviewer-vouched patches, timestamped results directory); the guardian is the LIGHT, always-on, in-session layer and defers all heavy on-demand rigor to SPEC-1.

**REQ-SG-003 — Explicit layering with the deep scan (Where — capability gate)**
**Where** the guardian and the deep scan both exist, the guardian **shall** document the layering explicitly so the two compose without overlap: the guardian = light + always-on + inline (reviews code as written); the deep scan = heavy + on-demand + explicit (`/moai review --deep`).

### Layer 1 — pattern warnings

**REQ-SG-010 — Instant pattern scan on Edit/Write (Event-driven)**
**When** an Edit, Write, or MultiEdit tool call completes (PostToolUse), the guardian **shall** scan the written content for known-dangerous patterns and surface a finding for each match.

**REQ-SG-011 — Patterns organized by vulnerability class, 16-language-neutral (Ubiquitous)**
The Layer-1 pattern set **shall** be organized by vulnerability CLASS (not by any single programming language), and each class **shall** apply equally across the 16 supported languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) — NO single language treated as PRIMARY.

**REQ-SG-012 — Coverage of the known-dangerous pattern classes (Ubiquitous)**
The Layer-1 pattern set **shall** cover the absorbed known-dangerous classes, including at minimum: unsafe deserialization (e.g. `yaml.load` on untrusted input, `pickle.load`, `torch.load(weights_only=False)`), unsafe DOM injection / XSS (raw `innerHTML`, `dangerouslySetInnerHTML`), hardcoded secrets, code injection / dynamic evaluation (`eval` on external input), SQL string concatenation, command injection (subprocess/child-process spawning with unsanitized input), and weak / insecure randomness or cryptography — approximately 25 patterns total across the classes.

**REQ-SG-013 — Layer 1 is regex-only, no LLM (Unwanted behavior)**
Layer 1 **shall not** invoke any language model, network call, or subprocess-per-pattern; it **shall** be a deterministic in-process pattern match so it never stalls an Edit/Write.

**REQ-SG-014 — Layer 1 is advisory, never blocks the edit (State-driven)**
**While** Layer 1 detects one or more dangerous patterns, the guardian **shall** surface an advisory warning back into the session and **shall not** block, reject, or revert the Edit/Write that triggered it.

**REQ-SG-015 — Layer 1 respects the hook timeout budget (Where — capability gate)**
**Where** Layer 1 runs on the PostToolUse event, it **shall** complete within the hook timeout budget (MoAI 5s policy default) and **shall** run asynchronously so it never adds latency to the Edit/Write it observes.

### Layer 2 — turn-end diff review

**REQ-SG-020 — Turn-end diff review (Event-driven)**
**When** the assistant finishes a turn (Stop hook), the guardian **shall** review the turn's working-tree diff for high-severity security findings.

**REQ-SG-021 — Surface high-severity findings back into the session (State-driven)**
**While** the turn-end review finds one or more high-severity issues, the guardian **shall** surface them back into the session (advisory `systemMessage`) so they can be addressed before the user acts on the result.

**REQ-SG-022 — Layer 2 advisory-by-default, opt-in escalation (State-driven)**
**While** no opt-in escalation flag is set, the Layer-2 review **shall** default to the regex pattern engine (fast, no LLM) and **shall** be advisory (non-blocking); a fast-model / blocking escalation **shall** occur only when explicitly enabled via an opt-in environment flag aligned with the existing `MOAI_SYNC_GATE_BLOCKING` model, and when enabled the model-backed review **shall** be delivered by the orchestrator spawning an `Agent()` review from the hook's structured signal — the hook itself **shall not** invoke a `type: agent` / `type: prompt` sub-model.

**REQ-SG-023 — Layer 2 default review is model-free (Where — capability gate)**
**Where** Layer 2 runs without the opt-in escalation flag, it **shall** reuse the Layer-1 pattern engine over the diff and **shall not** invoke a language model, keeping the default turn-end path fast.

### Layer 3 — commit-time cross-file review

**REQ-SG-030 — Commit-time cross-file review (Event-driven)**
**When** a git commit is made within the session and Layer 3 is enabled, the guardian **shall** run a review over the commit's changed files that traces data flow ACROSS files.

**REQ-SG-031 — Cross-file vulnerability classes (Ubiquitous)**
The Layer-3 review **shall** target multi-file vulnerability classes that single-file pattern-matching misses — including IDOR, authorization bypass, and cross-file SSRF — by reading related files (read-only reconnaissance) rather than matching a single buffer.

**REQ-SG-032 — Layer 3 opt-in and advisory-by-default (State-driven)**
**While** Layer 3 is not explicitly enabled, it **shall** be dormant (a no-op that never runs and never blocks); **while** Layer 3 IS enabled, it **shall** default to advisory (surface findings, do not block) and **shall** enter a blocking posture only when a further opt-in blocking flag is set.

**REQ-SG-033 — Layer 3 extends the commit-aware gate model (Where — capability gate)**
**Where** Layer 3 is realized, it **shall** extend MoAI's existing commit-aware gate model — a sibling commit-time Stop hook adding a security axis to the `sync-phase-quality-gate.sh` model — rather than introducing an unrelated new gate mechanism, and its cross-file findings **shall** be surfaced as a structured escalation signal the orchestrator translates into an `Agent()` review (settled surface L3-A; design.md §L3).

### Hook boundary & orchestrator translation

**REQ-SG-040 — Guardian hooks never prompt the user (Unwanted behavior)**
No guardian hook script **shall** invoke `AskUserQuestion` or `mcp__askuser`, and no guardian hook **shall** emit a free-form prose question directed at the user; hooks emit exit codes + structured JSON only.

**REQ-SG-041 — Orchestrator translates a block into AskUserQuestion (Event-driven)**
**When** a guardian hook signals a block (structured JSON `decision`/`continue:false` on the exit-0 stdout channel per the hook event's schema), the orchestrator **shall** translate that block into an `AskUserQuestion` round (accept / override with `--skip-hook` / abort), preserving the single-point-of-contact boundary.

**REQ-SG-042 — Blocking is opt-in; advisory is the default (State-driven)**
**While** no opt-in blocking flag is set, every guardian layer **shall** be advisory (emit findings, never block routine work); a blocking posture **shall** be reachable only through an explicit opt-in environment flag, aligned with the `MOAI_SYNC_GATE_BLOCKING` precedent.

**REQ-SG-043 — Documented, audit-logged `--skip-hook` override (Where — capability gate)**
**Where** a guardian gate supports a bypass, it **shall** accept the `--skip-hook` first-argument override and **shall** append the bypass to `.moai/logs/hook-skip.log`, matching the existing governance-gate audit convention.

### Distribution, neutrality & quality

**REQ-SG-050 — Template-First byte-lockstep distribution (Ubiquitous)**
Every shipped shell wrapper and every settings.json hook-wiring change **shall** be authored in `internal/template/templates/` FIRST, regenerated via `make build`, then synced to the local `.claude/` copy, with the template `.sh.tmpl` / rendered `.sh` and template `settings.json.tmpl` / rendered `settings.json` kept lockstep.

**REQ-SG-051 — 16-language neutrality of shipped template content (Unwanted behavior)**
The shipped shell wrappers, settings.json wiring, and any shipped pattern/prompt content **shall not** elevate a single language as PRIMARY, **shall not** hardcode a single-language toolchain as the only path, and **shall not** embed internal SPEC IDs, internal dates, or commit SHAs (CI guard: `internal_content_leak_test.go` + `template-neutrality-check.yaml`).

**REQ-SG-052 — Go code is compiled, tested, and NOT template content (Ubiquitous)**
The pattern engine and hook handlers **shall** live in Go under `internal/` (compiled into the binary, NOT shipped as template content), and **shall** carry tests using `t.TempDir()` isolation with no OTEL environment variables set in parallel tests (per CLAUDE.local.md §6).

**REQ-SG-053 — Single-source pattern configuration (Where — capability gate)**
**Where** the Layer-1 pattern set is defined, it **shall** live in a single-source configuration (one Go table / config file), NOT scattered across hook scripts or duplicated per language, per the hardcoding-prevention rules (CLAUDE.local.md §14 / §16).

**REQ-SG-060 — Graceful fail-open on missing dependencies (Unwanted behavior)**
No guardian hook **shall** crash, block, or break the session when the `moai` binary is unresolvable in all three resolution tiers or when a required tool (`jq`, `git`) is absent; every guardian hook **shall** degrade to a silent no-op (exit 0), inheriting the wrapper fail-open convention.

**REQ-SG-061 — Hook-event-schema-compliant output (Ubiquitous)**
Each guardian hook **shall** emit output compliant with its hook event's stdout schema — PostToolUse (Layer 1) uses `additionalContext` / `systemMessage` (advisory; async delivers `additionalContext` only), Stop (Layer 2) uses `systemMessage` or the `hookSpecificOutput.decision` block, and the Layer-3 surface uses the schema of whichever event it binds to — and **shall not** emit unknown fields that fail Claude Code JSON-schema validation.

---

## §C Out of Scope

This section prevents scope creep and records the layering boundaries. Items below are explicitly **out of scope** for SPEC-SEC-GUARDIAN-001.

### Out of Scope — On-demand multi-agent deep scan (Epic SECURITY-ABSORB SPEC-1)
- The heavy, explicitly-invoked deep scan (`/moai review --deep`) — its six-phase pipeline, 3-voter adversarial panel, reviewer-vouched patch drafting, and timestamped `.moai/reports/security-deepscan-*/` results directory — is owned entirely by SPEC-SEC-DEEPSCAN-001. This SPEC-2 covers ONLY the light, always-on, in-session guardian. The two layers compose without overlap: SPEC-1 = deep + on-demand + report/patch; SPEC-2 = light + always-on + inline warnings.

### Out of Scope — Patch generation / auto-fix / auto-apply
- The guardian identifies and surfaces findings; it **shall not** draft patches, apply fixes, stage, commit, or revert code. Patch drafting is SPEC-1's `--patch` flow. Auto-fix of a flagged pattern is deferred (a future SPEC could add a `/moai fix --security` path).

### Out of Scope — Blocking routine work by default
- No guardian layer blocks a routine Edit/Write/commit by default. Blocking is reachable only via explicit opt-in flags (§B REQ-SG-042). A default-blocking security gate is out of scope — it would stall every edit and contradict the advisory-first posture.

### Out of Scope — A new `/moai security` subcommand or new agent
- Reviving the retired `/moai security` subcommand, adding a new intent-router row, or adding a new agent to the catalog is out of scope (`catalog.yaml` agent/skill counts stay unchanged). The guardian is realized through hooks + compiled Go handlers, not a new subcommand or agent.

### Out of Scope — Shipping a workflow `.js` script or new skill
- No `.js` dynamic-workflow script is shipped into the template tree, and no new skill directory is added. The guardian is hook-driven; the security reference skills it draws vocabulary from already exist.

### Out of Scope — Offensive security / exploit execution
- The guardian is defensive-only: it flags and explains dangerous patterns and cross-file risks. Authoring working exploits, running attack payloads, or any offensive tooling is out of scope (consistent with the defensive posture of the reused reference skills).

### Out of Scope — Runtime / production security monitoring
- The guardian operates only inside a MoAI Claude Code session against the working tree (edit-time / turn-end / commit-time). Runtime application security, production intrusion detection, or CI/CD pipeline scanning (owned by `moai-ref-secops` operational guidance) is out of scope.

---

## §D Acceptance Criteria

Full Given-When-Then scenarios, the AC-to-REQ matrix, edge cases, and the Definition of Done live in `acceptance.md`. Summary of the must-pass gates:

- Layer 1 PostToolUse hook wired (template + local, byte-lockstep) and forwards to a compiled Go handler that flags the enumerated vulnerability classes across ≥16 languages from a single-source pattern table.
- Layer 1 is regex-only (no LLM), async, advisory (never blocks the edit); pattern count ≈ 25 across classes.
- Layer 2 Stop hook reviews the turn's diff; advisory-by-default (regex engine); LLM/blocking escalation opt-in via env flag.
- Layer 3 commit-time cross-file review is opt-in (dormant by default); reuses the commit-aware gate model; advisory-by-default when enabled.
- Hook boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/` → no matches; blocks emit structured JSON translated by the orchestrator.
- Fail-open: every guardian hook degrades to a silent no-op when `moai`/`jq`/`git` is absent.
- Template-First + 16-language neutrality: no PRIMARY-language elevation, no internal SPEC-ID/date/SHA leak in template content; `internal_content_leak_test.go` PASS.
- Go handlers + pattern engine carry tests (`t.TempDir`, no OTEL env in parallel tests); `go build ./...` (host + `GOOS=windows`) exit 0; catalog agent/skill counts unchanged.
