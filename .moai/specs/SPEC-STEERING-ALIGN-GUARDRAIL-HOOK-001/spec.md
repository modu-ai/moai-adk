---
id: SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001
title: "GLM web-tooling routing — env-deterministic guardrail hook + always-load removal"
version: "0.1.0"
status: in-progress
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P3
phase: "v3.x steering-alignment"
module: "internal/hook"
lifecycle: spec-anchored
tags: "steering, guardrail-hook, glm, always-load, token-budget"
tier: M
depends_on: [SPEC-STEERING-ALIGN-RULE-SCOPING-001]
---

# SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001 — GLM web-tooling guardrail hook

## HISTORY

- 2026-06-23: Initial draft. Epic Steering-Align SPEC 3 of 5 (P3). Implements the
  remedy that the predecessor RULE-SCOPING SPEC deferred for env-triggered Class-B
  rules: convert the always-loaded `glm-web-tooling.md` rule into an on-demand
  guardrail delivered by a deterministic hook, and remove the rule from the
  always-load set to realize the token saving.
- 2026-06-23: iter-2 audit revision (plan-auditor iter-1 PASS-WITH-DEBT 0.83).
  D1 (MAJOR): cg-leader carve-out pinned to the PROCESS-env detector `hasGLMEnv`
  (NOT the SESSION-env `sessionEnvHasGLM` / `IsCGMode`, both of which return `true`
  on the leader); two-detector hazard named in §A. D2: REQ-SARS-010 reconciliation
  clause added to §A (this SPEC is the deferred (b2) on-demand remedy; `paths:` is
  paired with hook delivery, not bare scope). D3: removal glob corrected to the
  `**/`-prefixed live-path-only self-glob `**/glm-web-tooling.md`. D4: behavioral
  load-on-touch AC added. D5: `z.ai`-vs-`api.z.ai` substring-superset note. D6:
  coverage rescoped to the NEW code path (90%+) + package baseline (≥ 82.7%),
  not a package-wide 90% target. D7: `tier: M` frontmatter added. MP-2: REQ-GH-012
  modality corrected (Event-detected → Event-driven).

## A. Context / Why

Anthropic's official Claude Code "steering" guidance favors deterministic hooks
over probabilistic always-loaded prose. The predecessor SPEC RULE-SCOPING-001
path-scoped the Class-A always-loaded rules (those with a clean file-touch
trigger) by adding a `paths:` frontmatter so they load only when a matching file
is edited. RULE-SCOPING-001 §D explicitly deferred the **Class-B** rules — those
with *no* clean file-touch trigger — noting that a normal `paths:` glob is
infeasible for them, and naming the remedy as "(b2) convert to an on-demand
skill/reference."

`glm-web-tooling.md` is the cleanest Class-B candidate because its trigger is
**environment-deterministic**, not intent-heuristic: the rule applies if and only
if the session runs on the GLM backend, which is mechanically detectable from the
`ANTHROPIC_BASE_URL` environment variable. This SPEC implements remedy (b2) for
that single rule via a deterministic guardrail hook, and removes the rule from
the always-load set.

The hook operates in **advisory context-injection** mode: it detects the GLM
backend and injects a concise routing reminder into context. It does **not** block
or deny any tool. Enforcement (PreToolUse deny/allow) was considered and rejected
as too behavior-changing.

### RULE-SCOPING REQ-SARS-010 reconciliation (D2)

RULE-SCOPING-001 carries REQ-SARS-010 ([HARD], Unwanted behavior): "The change
SHALL NOT add `paths:` to any Class-B or Class-C rule … The SPEC MUST NOT pretend
Class-B is cleanly scopable." It names `glm-web-tooling.md` as Class-B with
"Path-scoping is INFEASIBLE" (§J:145). REQ-GH-007 of THIS SPEC adds `paths:` to
exactly that rule. These are NOT in conflict, for three reasons:

1. **REQ-SARS-010's prohibition is scoped to the bare-scope move it forbade** —
   adding `paths:` to a Class-B rule *while leaving its env trigger undelivered*,
   which would "silently suppress a rule when its env/intent trigger fires without
   a file touch." That is the failure mode REQ-SARS-010 prevents. This SPEC does
   NOT do that: the env trigger is now delivered on-demand by the SessionStart
   guardrail hook. The `paths:` self-glob here drops only the always-load
   *residency* — it does not pretend a meaningful file-touch trigger exists, and
   it does not leave the GLM rule undelivered when its env trigger fires.
2. **This SPEC IS the deferred (b2) remedy RULE-SCOPING named.** RULE-SCOPING-001
   §J:149 states: "the real remedy is (b1) keep always-loaded but TRIM, or (b2)
   convert to an on-demand skill/reference … Deferred to a follow-up SPEC." This
   SPEC is that follow-up. RULE-SCOPING phrased (b2) as a "skill/reference"; the
   SessionStart guardrail hook satisfies the **same on-demand-delivery principle**
   (the rule's content reaches the model only when relevant), implemented as a
   deterministic hook rather than a skill. The principle — "deliver the rule
   on-demand instead of always-loading it" — is identical; only the delivery
   vehicle differs.
3. **Therefore the `paths:` here is paired with a delivery mechanism, not bare.**
   REQ-SARS-010 forbids bare scoping (scope-without-delivery); this SPEC pairs the
   scoping with hook delivery. The reconciliation is: REQ-SARS-010 stands for
   Class-B rules that have no on-demand delivery; once a Class-B rule gains a
   deterministic on-demand delivery path, removing its always-load residency via a
   self-glob is the intended (b2) outcome, not the forbidden bare-scope.

### Grounding facts (tool-verified, see research.md for the trace)

- `glm-web-tooling.md` is MIRRORED: the live `.claude/rules/moai/core/` copy and
  the `internal/template/templates/.claude/rules/moai/core/` copy are byte-identical
  (7545 B) and currently carry NO frontmatter, so both are always-loaded.
- LIVE always-loaded rule count = 11; removing `glm-web-tooling.md` → 10.
- TEMPLATE always-loaded rule count = 10 (the template tree already lacks one
  workflow rule the live tree carries); removing `glm-web-tooling.md` → 9.
- The GLM-backend signal is the **PROCESS** env `ANTHROPIC_BASE_URL` containing
  the substring `z.ai` (canonical default `https://api.z.ai/api/anthropic`). The
  repo already uses this exact PROCESS-env check at `internal/tmux/cg_detect.go`
  `hasGLMEnv()` (lines 202-210, verified this session). The hook reuses this
  process-env signal; it does NOT invent a new one. **Detector hazard (D1):**
  `cg_detect.go` exposes TWO GLM detectors — `hasGLMEnv()` (PROCESS env, line 202)
  and `sessionEnvHasGLM()` (tmux SESSION env, line 179). The cg-leader pane strips
  its PROCESS GLM env but the tmux SESSION env still carries `z.ai`; the hook MUST
  gate on the PROCESS-env disjunct (`hasGLMEnv` / a direct
  `os.Getenv("ANTHROPIC_BASE_URL")` substring check), NOT on `sessionEnvHasGLM`
  and NOT on `IsCGMode()` (which returns `true` on the leader via the session-env
  path, lines 97-100, and would silently break the carve-out). See §C.2 + research
  §R4.
- **Signal-substring note (D5):** the Go code (`hasGLMEnv`) matches the substring
  `z.ai`, while the glm-web-tooling.md HARD body (line 13) defines the canonical
  signal as `api.z.ai`. `z.ai` is a strict superset of `api.z.ai` — every real GLM
  session (`https://api.z.ai/...`) matches both, so no GLM session is missed; a
  hypothetical non-GLM custom gateway whose host merely contains `z.ai` would
  falsely trigger an advisory injection (accepted edge — advisory-only, no tool is
  blocked). Reusing the code's existing `z.ai` check is correct; this note flags
  the SSOT-vs-code wording inconsistency for the record.
- The context-injection mechanism is already live: the UserPromptSubmit handler's
  `detectWorkflowContext` returns an `additionalContext` string emitted via
  `HookSpecificOutput.AdditionalContext`. The hook Go code is embedded in the
  binary (`go:embed`), NOT a template-mirrored file — only the rule frontmatter
  change requires Template-First handling.

## B. Definitions

- **Always-load set**: the set of `.claude/rules/moai/**/*.md` rules whose
  frontmatter has NO `paths:` key. Claude Code loads these into every session's
  context as project instructions. A rule that carries a `paths:` key is loaded
  only when an edited file matches one of its globs.
- **GLM-backend session**: a session where the **PROCESS** env `ANTHROPIC_BASE_URL`
  contains the substring `z.ai` (equivalently runtime mode `LLMModeGLM`). Covers
  `moai glm` (whole-session) and the GLM teammate panes of `moai cg`. The signal
  is read from the process env (`cg_detect.go` `hasGLMEnv`), NOT the tmux SESSION
  env (`sessionEnvHasGLM`) — see the Detector hazard in §A.
- **cg-leader pane**: the `moai cg` leader pane runs the Claude backend with its
  **PROCESS** GLM env stripped, so its process-env `ANTHROPIC_BASE_URL` does NOT
  contain `z.ai` (even though the tmux SESSION env still carries `z.ai`). The
  built-in web tools are the canonical path there. The carve-out depends on
  reading the PROCESS env, not the session env.
- **Concise routing reminder**: the small injected payload — the HARD routing
  table summary (WebSearch → `mcp__web_search_prime__webSearchPrime`,
  WebFetch → `mcp__web_reader__webReader`, Read-on-image →
  `mcp__zai-mcp-server__*` vision tools) plus a one-line ToolSearch-preload note
  and a pointer to the full rule — NOT the full 7.5 KB rule body.

## C. Requirements (GEARS)

### C.1 Guardrail detection and injection

**REQ-GH-001** (Ubiquitous) — The guardrail hook shall reuse the existing
**PROCESS-env** `ANTHROPIC_BASE_URL`-substring GLM-backend signal (`z.ai`, as read
by `cg_detect.go` `hasGLMEnv` / a direct `os.Getenv("ANTHROPIC_BASE_URL")`
substring check) as its sole trigger and shall not introduce a second, divergent
detection signal. The hook shall NOT gate on the tmux SESSION-env detector
(`sessionEnvHasGLM`) nor on `IsCGMode()`, because both report `true` on the
`moai cg` leader pane and would break the carve-out (REQ-GH-005).

**REQ-GH-002** (Event-driven) — **When** the session is a GLM-backend session,
the guardrail hook shall inject the concise routing reminder into context via the
`additionalContext` field of `HookSpecificOutput`.

**REQ-GH-003** (Event-driven) — **When** the session is NOT a GLM-backend session
(the `ANTHROPIC_BASE_URL` env var is absent or does not contain `z.ai`), the
guardrail hook shall inject nothing and shall return an allow result.

**REQ-GH-004** (Ubiquitous) — The injected reminder shall contain the three HARD
routing replacements (web search, web fetch, image read) and a ToolSearch-preload
note, and shall be substantially smaller than the full `glm-web-tooling.md` rule
body (the injected payload shall be a concise summary, not the 7.5 KB rule).

### C.2 cg-leader carve-out preservation

**REQ-GH-005** (State-driven) — **While** the current pane is the `moai cg` leader
pane (Claude backend, PROCESS GLM env stripped, so the PROCESS env
`ANTHROPIC_BASE_URL` does not contain `z.ai` — even though the tmux SESSION env
still carries `z.ai`), the guardrail hook shall not inject the routing reminder —
the built-in web tools remain the canonical path there.

**REQ-GH-006** (Where) — **Where** the **PROCESS-env** GLM signal is the trigger
(REQ-GH-001), the cg-leader carve-out (REQ-GH-005) shall be satisfied
automatically by the absence of `z.ai` in the leader pane's **PROCESS** env
`ANTHROPIC_BASE_URL` — no separate leader-pane special-case branch is required in
the hook. This holds ONLY because the trigger reads the process env; gating on the
tmux SESSION-env detector (`sessionEnvHasGLM`) or `IsCGMode()` would inject on the
leader pane and silently violate the carve-out (the two-detector hazard, §A).

### C.3 Always-load removal

**REQ-GH-007** (Ubiquitous) — The `glm-web-tooling.md` rule shall be removed from
the always-load set in BOTH the live tree and the template tree, by adding a
`paths: "**/glm-web-tooling.md"` frontmatter. The `**/`-prefixed self-glob
(matching the in-repo precedent `paths: "**/NOTICE.md"` and the RULE-SCOPING C-4
`**/`-prefix mandate) matches the rule's own file on a working-tree edit (so an
editor of the rule still loads it on a file-touch) and, by virtue of carrying a
`paths:` key at all, excludes the rule from the always-load set. The glob carries
NO `internal/template/...` path: the Claude Code rules loader matches **live
working-tree paths only** (verified this session — no in-repo rule scopes into the
template tree), so a template-mirror path in the glob would be dead weight
(research §R2 / D3).

**REQ-GH-008** (Ubiquitous) — After removal, the live always-loaded rule count
shall be exactly one fewer than before (11 → 10) and the template always-loaded
rule count shall be exactly one fewer than before (10 → 9), verifiable by the
reproducible count command.

**REQ-GH-009** (Event-driven) — **When** the template-tree copy of
`glm-web-tooling.md` is edited, the change shall be applied Template-First
(`internal/template/templates/...` first, then `make build` re-embed, then live
parity), and the rule frontmatter shall end byte-aligned between the two trees.

### C.4 Quality and neutrality

**REQ-GH-010** (Ubiquitous) — The guardrail hook Go package shall pass
`go test ./... + go vet ./... + golangci-lint run` with zero errors. The **NEW
code path** added by this SPEC (the GLM-detection branch + the reminder-injection
function) shall meet the critical-package coverage target (90%+). The
`internal/hook` package-wide baseline is 82.7% (verified this session); this SPEC
shall NOT regress the package baseline below 82.7%, but it does NOT impose a
package-wide 90% target (which would force coverage work on unrelated pre-existing
code outside this SPEC's scope — D6).

**REQ-GH-011** (Where) — **Where** the template tree is edited (the rule
frontmatter), the edit shall stay free of internal SPEC IDs, REQ tokens, internal
dates, and commit SHAs per the template-neutrality CI guard.

**REQ-GH-012** (Event-driven) — **When** the hook's detection or injection
errors mid-run (e.g., env read fails), the hook shall return an allow result and
inject nothing, so a detection failure never blocks the prompt or the session.

## D. Out of Scope / Exclusions

This SPEC builds the GLM guardrail and nothing more. The following are explicitly
NOT in scope. (out of scope)

### Out of Scope — other Class-B always-loaded rules

- `dynamic-workflows.md`, `goal-directive.md`, and `verification-batch-pattern.md`
  remain in the always-load set unchanged. Their triggers are intent-heuristic
  (not env-deterministic), carry higher false-positive risk, and have no clean
  file-touch trigger. They are deferred to a possible later SPEC and are not
  touched here.
- No other always-loaded rule (`agent-common-protocol.md`, `askuser-protocol.md`,
  `verification-claim-integrity.md`, `context-window-management.md`,
  `runtime-recovery-doctrine.md`, `session-handoff.md`, `sprint-round-naming.md`)
  is modified.

### Out of Scope — enforcement / blocking mode

- The hook is advisory context-injection ONLY. No PreToolUse deny/allow
  enforcement, no tool blocking, no permission-decision gating is added. The hook
  never prevents a tool call; it only injects a reminder.

### Out of Scope — generalized intent-trigger framework

- No general "intent-trigger framework", no plugin abstraction, and no new config
  surface beyond what the single GLM guardrail needs. The change is the minimal
  coherent extension of the existing keyword/context-injection hook path plus one
  rule frontmatter edit.

### Out of Scope — non-GLM web-tooling concerns

- Built-in `WebSearch` / `WebFetch` / `Read`-on-image behavior on the Claude
  backend (including the cg-leader pane) is unchanged. This SPEC neither alters
  nor re-routes any tool on a non-GLM session.

## E. Cross-references

- `.claude/rules/moai/core/glm-web-tooling.md` — the rule being converted to
  on-demand delivery (the SSOT routing table the hook summarizes).
- RULE-SCOPING-001 §J + REQ-SARS-010 — the deferred-remedy origin (Class-B "(b2)
  convert to on-demand") and the bare-scope prohibition this SPEC reconciles
  against (§A RULE-SCOPING REQ-SARS-010 reconciliation).
- `internal/hook/user_prompt_submit.go` `detectWorkflowContext` — the
  context-injection precedent the hook extends.
- `internal/tmux/cg_detect.go` `hasGLMEnv` (process-env, line 202) vs
  `sessionEnvHasGLM` (session-env, line 179) — the two-detector hazard; the hook
  reuses the PROCESS-env disjunct (§A Detector hazard / D1).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the 5-section
  evidence discipline the acceptance criteria follow.
