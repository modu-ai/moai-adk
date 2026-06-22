# Research — SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001

Records the tool-verified evidence behind the two load-bearing plan.md decisions
(§D.1 removal mechanism, §D.2 injection point). All findings below were verified
against the repo state on 2026-06-23, not assumed.

## §R1. The always-load mechanism (loader behavior)

**Finding.** Always-load residency is governed by the Claude Code native rules
system, NOT by moai-adk Go code. A rule under `.claude/rules/moai/**/*.md` is
loaded into every session as project instructions IF its frontmatter has NO
`paths:` key. A rule WITH a `paths:` key is loaded only when an edited file
matches one of its globs.

**Evidence.**
- Reproducible count command (rules WITHOUT a `paths:` key in their first 8
  frontmatter lines):
  - Live tree → 11 rules. Members: `agent-common-protocol.md`,
    `askuser-protocol.md`, `glm-web-tooling.md`, `verification-claim-integrity.md`,
    `sprint-round-naming.md`, `context-window-management.md`,
    `dynamic-workflows.md`, `goal-directive.md`, `runtime-recovery-doctrine.md`,
    `session-handoff.md`, `verification-batch-pattern.md`.
  - Template tree → 10 rules (same set minus `runtime-recovery-doctrine.md`,
    which the template tree does not carry in its always-load set — a pre-existing
    composition difference, out of scope for this SPEC).
- Rules that ARE `paths:`-scoped (38+ in the live tree) confirm the lever:
  language rules (`go.md` → `**/*.go,...`), `moai-constitution.md` →
  `**/.claude/**`, `zone-registry.md` → `.claude/rules/**,.claude/agents/**`, etc.
  RULE-SCOPING-001 added these `paths:` keys to remove Class-A rules from
  always-load.

**Conclusion.** Adding a `paths:` key to `glm-web-tooling.md` removes it from the
always-load set. This is the loader's actual, verified behavior.

## §R2. Removal-mechanism alternatives (why the `**/glm-web-tooling.md` self-glob)

The chosen mechanism (plan.md §D.1) gives `glm-web-tooling.md` the self-glob
`paths: "**/glm-web-tooling.md"`. Two D3 corrections, both verified against source
this session, shape the exact glob form:

- **`**/` prefix is the in-repo precedent (RULE-SCOPING C-4 [HARD]).** Every
  path-anchoring live `paths:` uses the `**/` prefix; `NOTICE.md` carries the exact
  pure self-glob shape `paths: "**/NOTICE.md"`. RULE-SCOPING-001
  REQ-SARS-002/003/004 mandate `**/`, and its iter-1 D5 fix normalized exactly
  this. So the glob is `**/glm-web-tooling.md`, never a bare unprefixed path.
- **Loader matches LIVE working-tree paths only.** Verified this session: NO
  in-repo rule scopes its `paths:` into `internal/template/...`. The Claude Code
  rules loader operates on the live `.claude/rules/` tree; the template-mirror copy
  is never loaded as a project-instruction rule. So a template-tree path in the
  glob would be dead weight — dropped. The single self-glob `**/glm-web-tooling.md`
  is correct and complete.

Alternatives considered and rejected:

- **(R2-a) Relocate the file out of `.claude/rules/`.** Rejected. Six existing
  cross-references point at the canonical path
  `.claude/rules/moai/core/glm-web-tooling.md`
  (`agent-common-protocol.md`, `moai-constitution.md`, `settings-management.md`,
  `moai-domain-research/SKILL.md`, `einstein.md`, `CLAUDE.md` §10/§12). Moving the
  file breaks all six. Verified by grep: the path is referenced across rules,
  output-styles, and CHANGELOG entries.
- **(R2-b) Add a bespoke frontmatter flag (e.g. `always_load: false`).** Rejected.
  The loader keys off the *presence of the `paths:` key*, not a custom boolean.
  A flag the loader does not honor would leave the rule always-loaded.
- **(R2-c) Delete the rule.** Rejected. The rule is the SSOT the hook reminder
  summarizes and that six cross-refs target. It must remain at its canonical path;
  only its always-load residency changes.

**Precedent for the pure self-glob.** `NOTICE.md` carries `paths: "**/NOTICE.md"` —
a pure `**/`-prefixed self-glob, the exact shape adopted here. (Other rules scope
to a small specific file set — `orchestrator-templates.md` →
`.claude/rules/moai/core/moai-constitution.md,CLAUDE.md`; `archived-agent-rejection.md`
→ four specific workflow files — confirming that narrow self/specific-file scoping
is an accepted pattern.) A `paths:` that matches only the rule's own file removes
the rule from always-load while still surfacing it to an editor of the rule.

## §R3. Injection-point trade study (SessionStart vs UserPromptSubmit vs PreToolUse)

| Candidate | GLM signal available? | Cost | Coupling | Verdict |
|-----------|----------------------|------|----------|---------|
| SessionStart | Yes — `ANTHROPIC_BASE_URL` known at session start, stable for the session | Once per session (cheapest) | None (advisory) | **Chosen** |
| UserPromptSubmit | Yes | Every prompt (re-injects each turn) | None | Rejected — per-prompt token cost with no benefit (signal does not change mid-session) |
| PreToolUse | Yes | Per tool call | Couples advisory reminder to a tool-call gate | Rejected — the SPEC is advisory-only; PreToolUse implies enforcement semantics |

**Evidence the SessionStart channel is already live.** `internal/hook/session_start.go`
emits `HookSpecificOutput.AdditionalContext` to inject the orchestrator UUID
post-compaction. The guardrail reminder appends to the same channel — no new
emission plumbing needed.

**Evidence the injection pattern is proven.** `internal/hook/user_prompt_submit.go`
`detectWorkflowContext` returns a non-empty `additionalContext` string on a
keyword match and the handler emits it via `HookSpecificOutput`. The GLM guardrail
is the same shape with an env trigger replacing the keyword trigger.

## §R4. GLM signal source — TWO detectors, pin to the PROCESS-env one (D1 MAJOR)

The canonical GLM-backend signal is `ANTHROPIC_BASE_URL` containing the substring
`z.ai` (canonical default `https://api.z.ai/api/anthropic`, constant
`DefaultGLMBaseURL` in `internal/config/defaults.go`; env key constant
`EnvAnthropicBaseURL` in `internal/config/envkeys.go`).

**Two-detector hazard (verified against `internal/tmux/cg_detect.go` this session,
lines 60-210).** The file exposes TWO GLM detectors that read DIFFERENT env
sources:

| Detector | Line | Env source | On cg-leader pane |
|----------|------|-----------|-------------------|
| `hasGLMEnv()` | 202-210 | **PROCESS** env `os.Getenv("ANTHROPIC_BASE_URL")` | returns **false** (process env stripped) ✓ correct for carve-out |
| `sessionEnvHasGLM()` | 179-191 | **tmux SESSION** env via `tmux show-environment` | returns **true** (session env still carries `z.ai`) ✗ would break carve-out |
| `IsCGMode()` | 90-130 | layered OR — fires via `sessionEnvHasGLM` when `team_mode=="cg"` (lines 97-100) | returns **true** ✗ would break carve-out |

The PROCESS-env check is:

```go
// cg_detect.go:202 hasGLMEnv() — the disjunct the hook reuses
if strings.Contains(os.Getenv("ANTHROPIC_BASE_URL"), "z.ai") {
```

The `moai cg` leader pane strips its PROCESS GLM env but the tmux SESSION env still
carries `z.ai` — that is the documented reason `sessionEnvHasGLM` exists
(`cg_detect.go:73-75` comment: "This path detects CG mode even when the leader
PROCESS env is clean, which is the whole point of CG mode"). Consequently:

- **The hook MUST gate on the PROCESS-env signal** (`hasGLMEnv()` or an equivalent
  direct `os.Getenv("ANTHROPIC_BASE_URL")` substring test). This makes the
  cg-leader carve-out automatic — the leader's PROCESS env lacks `z.ai`, so the
  check returns false and nothing is injected (REQ-GH-005 / REQ-GH-006).
- **The hook MUST NOT gate on `sessionEnvHasGLM` or `IsCGMode()`.** Both return
  `true` on the cg-leader pane, so gating on either would inject the reminder on
  the leader pane and silently violate the carve-out. This is the D1 correctness
  hazard. acceptance.md AC-GH-009b asserts NO injection when the PROCESS env lacks
  `z.ai` even if a tmux GLM SESSION marker is present.

**Substring drift (D5).** The Go code matches the substring `z.ai`; the
glm-web-tooling.md HARD body (line 13) defines the canonical signal as `api.z.ai`.
`z.ai` is a strict superset of `api.z.ai`: every real GLM session
(`https://api.z.ai/...`) matches both, so no GLM session is missed. A hypothetical
non-GLM custom gateway whose host merely contains `z.ai` (but not `api.z.ai`)
would falsely trigger an advisory injection — an accepted edge, because the hook
is advisory-only (no tool is blocked, so a false-positive reminder is harmless).
Reusing the code's existing `z.ai` check is correct; this note flags the
SSOT-vs-code wording inconsistency for the record.

## §R5. Template-First scope

The hook Go code is embedded in the binary (`go:embed`), NOT a template-mirrored
file — the template tree carries only shell wrappers (`handle-*.sh.tmpl`) that
shell out to `moai hook <event>`. Therefore Template-First (CLAUDE.local.md §2)
applies ONLY to the `glm-web-tooling.md` frontmatter edit: edit the template copy
first, `make build` to re-embed, then verify live parity. The Go detection +
injection logic is single-tree (the moai-adk repo) and needs no mirror.
