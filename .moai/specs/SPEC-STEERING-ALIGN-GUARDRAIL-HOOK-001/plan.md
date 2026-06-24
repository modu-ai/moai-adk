# Implementation Plan — SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001

## §A. Context

Tier: **M (standard)**. Rationale: this SPEC touches Go (a new guardrail detection
+ context-injection path in a critical hook package, plus tests at the 90%+
coverage target) AND a mirrored rule frontmatter change AND a cross-tree
always-load count verification. That is more than a doc-only Tier S. It is a
single coherent feature (one hook extension + one rule frontmatter edit), not a
multi-component architecture or breaking change, so it does not warrant Tier L
(no design.md). Artifact set: spec.md + plan.md + acceptance.md + progress.md,
plus a short research.md recording the two load-bearing design decisions.

The work realizes the remedy RULE-SCOPING-001 §D deferred for env-triggered
Class-B rules — and it realizes the token saving in the same SPEC (the rule
removal is NOT deferred).

## §B. Known issues / preconditions discovered

1. **Trees are NOT byte-identical in always-load composition.** The template tree
   currently has 10 always-loaded rules; the live tree has 11 (the live tree
   carries `runtime-recovery-doctrine.md` in the always-load set, the template
   tree does not). This is pre-existing and out of scope. The relevant invariant
   for THIS SPEC is per-tree delta: live 11 → 10, template 10 → 9. Do NOT attempt
   to reconcile the unrelated 1-rule composition difference.
2. **Hook Go code is embedded, not template-mirrored.** `internal/hook/*.go` is
   compiled into the binary; the template tree only carries shell wrappers
   (`handle-*.sh.tmpl`) that shell out to `moai hook <event>`. So the Go detection
   + injection logic lives in ONE tree. Template-First applies ONLY to the
   `glm-web-tooling.md` frontmatter edit.
3. **The injection precedent already exists.** `detectWorkflowContext` in
   `internal/hook/user_prompt_submit.go` returns an `additionalContext` string on
   a keyword match. The GLM guardrail is the same shape with an env trigger
   instead of a keyword trigger.

## §C. Pre-flight checklist

- [ ] Confirm live always-load count = 11 and template = 10 with the reproducible
      count command (baseline capture before any edit).
- [ ] Confirm `glm-web-tooling.md` is byte-identical across both trees (7545 B,
      no frontmatter) before editing.
- [ ] Confirm the PROCESS-env `ANTHROPIC_BASE_URL` check exists at
      `internal/tmux/cg_detect.go` `hasGLMEnv()` (line 202) — the reuse target.
      Confirm `sessionEnvHasGLM()` (line 179) and `IsCGMode()` (line 90) are the
      session-env detectors that MUST NOT be used as the hook trigger (D1 hazard).
- [ ] Confirm the chosen injection-point handler exists and already emits
      `HookSpecificOutput.AdditionalContext`.

## §D. Load-bearing decisions (research-then-decide)

### §D.1 Removal mechanism — how `glm-web-tooling.md` leaves the always-load set

**Decision: add a `paths: "**/glm-web-tooling.md"` self-glob frontmatter.**

The Claude Code rules loader treats a rule with NO `paths:` key as always-loaded,
and a rule WITH a `paths:` key as load-on-file-touch (loaded only when an edited
file matches a glob). RULE-SCOPING-001 used exactly this lever for Class-A rules.
For an env-triggered Class-B rule there is no natural file-touch glob, so the
established repo pattern is to give the rule a `paths:` whose glob matches the
rule's own file. Precedent in the repo: `NOTICE.md` carries `paths: "**/NOTICE.md"`
(a pure self-glob — the exact shape used here). The mere *presence* of a `paths:`
key removes the rule from the always-load set; the glob value then governs the
(now-rare) on-touch load.

**D3 corrections applied (two, both verified against source this session):**

1. **`**/` prefix (RULE-SCOPING C-4 [HARD]).** Every live `paths:` in the repo
   that anchors a path uses the `**/` prefix (RULE-SCOPING REQ-SARS-002/003/004
   mandate it; its iter-1 D5 fix normalized exactly this). The glob is therefore
   `**/glm-web-tooling.md`, NOT a bare `.claude/rules/.../glm-web-tooling.md`.
2. **No template-tree path in the glob.** The loader matches **live working-tree
   paths only** — verified this session: NO in-repo rule scopes its `paths:` into
   `internal/template/...`. A template-mirror path in the glob would be dead weight
   (the loader never matches it). The glob is the single live-path self-glob
   `**/glm-web-tooling.md`, which also matches the template copy's basename — but
   the template copy is never loaded as a project-instruction rule anyway (only the
   live `.claude/rules/` tree is the loader's domain).

Concretely, the frontmatter added to BOTH copies of `glm-web-tooling.md`
(byte-identical for mirror parity; the template copy carries the same frontmatter
so the mirror-drift guard stays green, even though only the live copy is
loader-relevant):

```yaml
---
description: "GLM-backend web-tooling routing SSOT. Delivered on-demand by the GLM guardrail hook (SessionStart) when a GLM backend is detected; loaded into context here only when this rule file is edited."
paths: "**/glm-web-tooling.md"
---
```

Why this glob: editing the rule is the only file-touch where a human needs the
full rule in context. Every *other* time the rule's content is needed at runtime,
the hook delivers the concise reminder. This is the cleanest honoring of the
loader's actual behavior — verified against the loader's no-`paths`-means-always-load
semantics + the live-path-only matching domain, not assumed.

Rejected alternatives (recorded in research.md §R2): (a) relocating the file out
of `.claude/rules/` — breaks the six existing cross-references that point at the
canonical path; (b) a bespoke frontmatter flag the loader does not honor — the
loader keys off `paths:` presence, not a custom flag; (c) deleting the rule — the
rule is the SSOT the hook summarizes and the six cross-refs target, so it must
remain at its canonical path.

### §D.2 Injection point — SessionStart vs UserPromptSubmit vs PreToolUse

**Decision: SessionStart** (once-per-session injection).

The GLM signal (PROCESS env `ANTHROPIC_BASE_URL`) is known at session start and
does not change within a session. Injecting the reminder once at SessionStart is
strictly cheaper than re-injecting on every UserPromptSubmit, and avoids per-prompt
token cost. PreToolUse is rejected outright because it would couple the advisory
reminder to a tool-call gate (the SPEC is explicitly advisory-only, no PreToolUse
semantics). The SessionStart handler already emits
`HookSpecificOutput.AdditionalContext` (it injects the orchestrator UUID), so the
guardrail reminder is appended to the same channel — no new emission plumbing. The
acceptance criteria pin SessionStart so the choice is mechanically verifiable.

**D1 — detector pinning (MAJOR, correctness hazard, verified against source).**
`internal/tmux/cg_detect.go` exposes TWO GLM detectors:
- `hasGLMEnv()` (line 202) reads the **PROCESS** env `os.Getenv("ANTHROPIC_BASE_URL")`.
- `sessionEnvHasGLM()` (line 179) reads the **tmux SESSION** env via
  `tmux show-environment`.

On the `moai cg` leader pane the PROCESS GLM env is stripped but the tmux SESSION
env still carries `z.ai` (that is the whole reason `sessionEnvHasGLM` exists — see
the comment at `cg_detect.go:73-75`). Critically, `IsCGMode()` (line 90) returns
`true` on the leader pane via the session-env path (lines 97-100,
`readLLMTeamMode == "cg"` + `sessionEnvHasGLM()`), by design.

Therefore the guardrail hook MUST gate on the **PROCESS-env** check
(`hasGLMEnv()`, or an equivalent direct `os.Getenv("ANTHROPIC_BASE_URL")`
substring test), NOT on `sessionEnvHasGLM` and NOT on `IsCGMode()`. If the
implementer reuses `sessionEnvHasGLM` or `IsCGMode`, the hook WILL inject the
reminder on the cg-leader pane and silently violate the carve-out (REQ-GH-005).
The implementation MUST add a tiny process-env helper (e.g.
`hookProcessEnvHasGLM()` reading `ANTHROPIC_BASE_URL`) or call `hasGLMEnv` — and
acceptance.md AC-GH-009b asserts NO injection when the PROCESS env lacks `z.ai`
even if a tmux GLM SESSION marker is present.

> Note: if a future need arises to refresh the reminder after a backend switch
> mid-session, UserPromptSubmit is the fallback. That is out of scope here —
> backend does not switch within a session under `moai glm` / `moai cg`.

### §D.3 Payload content (concise reminder)

The injected string is a compact summary, not the 7.5 KB rule:

```
[GLM backend detected] Route web tooling to z.ai MCP (not built-in):
WebSearch → mcp__web_search_prime__webSearchPrime
WebFetch  → mcp__web_reader__webReader
Read-on-image → mcp__zai-mcp-server__* (pass a local file path, not base64)
Preload deferred MCP schemas first: ToolSearch(query: "select:<tool>").
Full rule: .claude/rules/moai/core/glm-web-tooling.md
```

Net token effect: non-GLM sessions load nothing (rule no longer always-loaded +
no injection); GLM sessions get this small reminder once. Positive net saving on
every Claude-backend session, near-neutral-to-positive on GLM sessions (small
reminder vs the previously always-loaded 7.5 KB rule).

## §E. Self-verification deliverables

The run phase must produce, in progress.md §E.2:
- Before/after always-load counts for BOTH trees (reproducible command output).
- `go test ./internal/hook/... -cover` output: NEW code path (GLM-detection
  branch + injection function) at 90%+, AND package baseline not regressed below
  82.7% (the verified-this-session baseline). NOT a package-wide 90% target (D6).
- `go vet ./...` + `golangci-lint run` clean output.
- A grep proving the SessionStart handler emits the GLM reminder only when
  `ANTHROPIC_BASE_URL` contains `z.ai`.
- Template-neutrality test pass for the rule frontmatter edit.
- Byte-parity confirmation between the two `glm-web-tooling.md` copies after edit.

## §F. Milestones (priority-ordered, no time estimates)

- **M1 — Baseline capture.** Record live=11 / template=10 always-load counts and
  the byte-identical state of both `glm-web-tooling.md` copies. (Evidence basis
  for the before/after AC.)
- **M2 — RED: guardrail detection test.** Write a failing test asserting the
  SessionStart handler injects the concise reminder when `ANTHROPIC_BASE_URL`
  contains `z.ai`, and injects nothing when it does not (covers REQ-GH-002 /
  REQ-GH-003 / REQ-GH-005). Confirm it fails.
- **M3 — GREEN: implement detection + injection.** Add the env-signal helper
  (reusing the `z.ai` substring check) and append the concise reminder to the
  SessionStart `additionalContext` only on a GLM backend. Minimal implementation
  to pass M2. (REQ-GH-001/002/003/004/006/012.)
- **M4 — Always-load removal (Template-First).** Add the `paths:` frontmatter from
  §D.1 to the template copy FIRST, `make build` re-embed, then verify live parity;
  confirm live 11→10 and template 10→9. (REQ-GH-007/008/009/011.)
- **M5 — REFACTOR + coverage.** Tidy the detection helper, raise hook-package
  coverage to the 90%+ target, run the full verification batch
  (`go test ./... + go vet + golangci-lint`). (REQ-GH-010.)
- **M6 — Acceptance sweep.** Run every acceptance.md AC command, capture verbatim
  output into progress.md §E.2, confirm all PASS including the cg-leader carve-out
  and template-neutrality.

## §G. Anti-patterns to avoid

- **AP-GH-1 — Inventing a second GLM signal.** Reuse the `z.ai` substring check;
  do not add a divergent detection path (REQ-GH-001).
- **AP-GH-2 — Injecting the full 7.5 KB rule.** Inject only the concise reminder
  (REQ-GH-004); injecting the full rule defeats the token saving.
- **AP-GH-3 — Special-casing the cg-leader pane.** The env signal already excludes
  it; do not add a leader-pane branch (REQ-GH-006).
- **AP-GH-4 — Deferring the rule removal.** Build the hook AND remove the rule from
  always-load in this SPEC; do not split the removal into a follow-up.
- **AP-GH-5 — Adding a config surface.** No new config keys, no intent-trigger
  framework; the change is minimal and coherent.
- **AP-GH-6 — Editing the live rule first.** The rule is mirrored — Template-First
  (template → make build → live parity), per CLAUDE.local.md §2.

## §H. Cross-references

- spec.md §C (GEARS requirements) / §D (Out of Scope).
- research.md §R1 (loader mechanism), §R2 (rejected removal alternatives),
  §R3 (injection-point trade study).
- acceptance.md §D (AC matrix).
- `.claude/rules/moai/core/glm-web-tooling.md` (rule under conversion).
- `internal/hook/session_start.go` (injection-point handler).
- `internal/tmux/cg_detect.go` (reused env signal).
