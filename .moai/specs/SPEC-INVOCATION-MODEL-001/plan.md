# Implementation Plan — SPEC-INVOCATION-MODEL-001

> Tier M. Derived from spec.md. Priority-ordered milestones (no time estimates).
> development_mode: **tdd** — the config-ization (M2) is testable Go code with a unit test; the doctrine rule (M1) and workflow edits (M3) are documentation/policy (no test).

---

## §A. Context

Two coupled deliverables from a subcommand-audit reframing (see spec.md §A):
1. A policy-layer doctrine rule that names the invocation-model axis (PROGRAMMATIC bundled skill vs HUMAN-ONLY built-in command) as the true justification for a MoAI subcommand's existence.
2. A hardening of `/moai feedback` (the doctrine's first concrete application), covering diagnostic attachment, duplicate detection, gh-failure fallback, and target-repo config-ization.

The two are coupled because the feedback hardening is the worked example that proves the doctrine: `/moai feedback` has no native equivalent (it targets the remote tool repo, not the user's repo), so it is a legitimate first-class MoAI workflow whose quality is worth investing in.

---

## §B. Known Constraints (HARD — carried from spec.md §E and CLAUDE.local.md)

- **[HARD] Template-First mirror** (CLAUDE.local.md §2). Every edit to a `.claude/` or `.moai/config/` asset MUST be mirrored to `internal/template/templates/<path>`. `make build` recompiles the embedded templates (there is NO generated `embedded.go`; `//go:embed all:templates` in `internal/template/embed.go` embeds the tree directly). Mirror pairs are enumerated per-milestone below.
- **[HARD] 16-language neutrality** (CLAUDE.local.md §15 / §25). The doctrine rule + feedback edits are template-distributed to all 16 languages — keep them language-neutral. The `modu-ai/moai-adk` default repo is a MoAI-ADK system identifier (explicitly ALLOWED per §25 allowed-class), NOT internal-dev leakage. The CI guards `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` are the safety net.
- **[HARD] Dedupe interaction shape (orchestrator-direct — D2)**. `/moai feedback` runs orchestrator-direct (`user-invocable: false`, "the orchestrator executes directly", no subagent spawn) and ALREADY uses orchestrator-level AskUserQuestion legitimately for feedback collection. The constraint is NOT a subagent boundary: the duplicate-detection step MUST emit a structured "possible duplicates found" candidate-report that the orchestrator turns into an AskUserQuestion round, rather than hard-coding a blind inline dedupe prompt.
- **[HARD] Doctrine is policy-layer only**. No runtime enforcement mechanism (hook / lint) for the invocation-model classification.
- **[HARD] SPEC scope boundary**. Axis-A bundled-skill reuse refactoring is OUT OF SCOPE (follow-up SPEC). This SPEC applies the doctrine to feedback ONLY.

---

## §C. Pre-flight (run-phase entry checks)

1. Confirm the local `.claude/skills/moai/workflows/feedback.md` and its template mirror are byte-identical at run-phase start (verified during plan: **IDENTICAL** as of 2026-07-01).
2. Confirm no `SPEC-INVOCATION-MODEL-*` collision (verified during plan: **unique**).
3. Config-section registration is RESOLVED (LIVE FACT — D4): `internal/config/loader.go` uses explicit per-section loaders + per-file wrapper structs (NOT a directory scan). A new `feedback.yaml` therefore requires a `feedbackFileWrapper` + `loadFeedbackSection` + `Loader.Load()` wiring — these are DEFINITE M2 deliverables (see §E M2). Run-phase re-confirms the exact wrapper/loader naming against the live loader before implementing.

---

## §D. Technical Approach and Design Decisions

### D1. Doctrine rule (M1)

Resolved path (fixed): the doctrine rule is created at `.claude/rules/moai/workflow/native-invocation-model.md` (template mirror `internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md`) — `workflow/` subdirectory, sibling to `dynamic-workflows.md` and `goal-directive.md`.

Follow the existing workflow-rule convention observed in `.claude/rules/moai/workflow/{dynamic-workflows,goal-directive}.md`: a `> **Loading scope**` note near the top, a body of taxonomy + matrix + thesis + per-subcommand notes + caveat, a `## Cross-references` section, and (since this is a canonical doctrine reference) an optional `Version` / `Classification` footer per `coding-standards.md` § Footer Convention.

Provisional classification matrix — each row carries a per-command citation anchor to be authored verbatim in the doctrine rule (REQ-IM-002). The tags below are the coordinator-supplied provisional classification; run-phase MUST re-confirm each against the official `commands.md`/`skills.md` via WebFetch before committing (AC-IM-022):

| Native command | Category | Citation anchor (per-command) |
|----------------|----------|-------------------------------|
| `/code-review` | PROGRAMMATIC | `[Skill]` bundled marker (commands.md) |
| `/simplify` | PROGRAMMATIC | `[Skill]` bundled marker (commands.md, v2.1.154+) |
| `/loop` | PROGRAMMATIC | `[Skill]` bundled marker (commands.md ~L94) — self-paces when interval omitted; NOT merely a time-interval scheduler |
| `/deep-research` | PROGRAMMATIC | `[Workflow]` bundled marker (commands.md ~L68; in-repo dynamic-workflows.md ~L73) |
| `/security-review` | HUMAN-ONLY | built-in command, NO `[Skill]`/`[Workflow]` marker |
| `/goal` | HUMAN-ONLY | built-in command, NO marker |
| `/review` | HUMAN-ONLY | built-in command |
| `/clear` | HUMAN-ONLY | built-in command |
| `/compact` | HUMAN-ONLY | built-in command |

> **`/loop` framing correction (REQ-IM-002)**: the sibling rule `.claude/rules/moai/workflow/goal-directive.md` describes native `/loop` as "a fixed time interval scheduler" — this is INCOMPLETE. The official `commands.md` shows native `/loop` is a bundled `[Skill]` that self-paces when the interval is omitted (not merely a scheduler). The doctrine rule MUST carry a cross-reference that ACKNOWLEDGES + CORRECTS this `goal-directive.md` framing. The doctrine does NOT edit `goal-directive.md` itself (out of scope — that would be scope creep).

> Run-phase MUST verify each tag + citation anchor against the current official docs before committing — the classification is the load-bearing content and must not be asserted without observing the source (verification-claim-integrity §1.1). This official-docs verification MUST be RECORDED as evidence (AC-IM-022).

### D2. Feedback target-repo config-ization (M2 — the testable Go part)

Config-section pattern (observed in `internal/config/types.go`): each `.moai/config/sections/<name>.yaml` maps to a top-level YAML key merged into `Config`, with a Go struct tagged `yaml:"<name>"` (e.g. `ResearchConfig` / `yaml:"research"`).

Fixed shape (resolved decision — section file `.moai/config/sections/feedback.yaml`, key `feedback.repository`, default `modu-ai/moai-adk`):
- New section file: `.moai/config/sections/feedback.yaml`
  ```yaml
  feedback:
      repository: modu-ai/moai-adk   # default tool feedback channel; fork maintainers override
  ```
- New Go struct `FeedbackConfig { Repository string yaml:"repository" }` added to `internal/config/types.go`, wired into `Config` as `Feedback FeedbackConfig yaml:"feedback"`.
- Default `modu-ai/moai-adk` set in `internal/config/defaults.go` so an absent section still resolves to the tool channel.
- An accessor (e.g. `FeedbackRepository()` returning the resolved value with default fallback) — location: `internal/config/workflow_accessors.go` or a new `feedback_accessors.go`.
- **Definite loader wiring (LIVE FACT — D4)**: `internal/config/loader.go` uses explicit per-section loaders (`loadUserSection`, `loadQualitySection`, … `loadResearchSection` via `loadYAMLFile`) + per-file wrapper structs (e.g. `researchFileWrapper`, types.go ~L1100). A new `feedback.yaml` therefore DEFINITELY needs a `feedbackFileWrapper` struct + a `loadFeedbackSection` loader + wiring into `Loader.Load()`. This is a required M2 deliverable, not conditional.
- Test (TDD RED→GREEN): the resolution test is an **integration/file-load test** — it writes a real `feedback.yaml` (with and without an override) into a temp dir and loads it through `Loader.Load()`, so it exercises the loader registration (a hand-constructed struct would NOT verify wiring). Default/absent → `modu-ai/moai-adk`; override → override value.

Design note: `internal/config/**` Go code is the TOOL implementation and is **NOT** template-mirrored. Only the `.moai/config/sections/feedback.yaml` section file is a mirror pair.

### D3. Feedback workflow enhancement (M3)

Edit `.claude/skills/moai/workflows/feedback.md` (+ template mirror) for the 4 items:
1. Diagnostic attachment (REQ-IM-006/007) — **Resolved (HYBRID)**: the skill ALWAYS collects the 2 guaranteed items — `moai version`, `uname` (OS) — the skill-guaranteed diagnostic set (this is the testable contract of REQ-IM-006), and additionally attempts `go version` (tool build-provenance) on a best-effort basis (its absence is NOT an AC failure, per the §15 neutrality note). Beyond that, the orchestrator MAY pass last-failed-command / error context into the feedback invocation prompt; the skill appends it if present (best-effort additive, NOT required — its absence is not an AC failure). The skill never reads session error history itself and never attaches arbitrary user file contents (REQ-IM-007 boundary).
2. Duplicate detection (REQ-IM-008/014): before issue creation, run `gh issue list --repo <resolved-target> --search "<title keywords>"`; emit a structured "possible duplicates" report for the orchestrator (NOT an in-skill AskUserQuestion).
3. gh-failure fallback (REQ-IM-009): detect `gh auth status` failure / rate-limit; guide the user to resolve; offer to save the drafted issue body locally (e.g. under `.moai/state/` or a user-visible path) so no draft is lost.
4. Config reference (REQ-IM-010/011): replace the hardcoded `--repo modu-ai/moai-adk` (line 104, both local + template) with the resolved config value, default preserved as `modu-ai/moai-adk`.

### D4. Build + verification (M4)

`make build` to regenerate embedded templates; then the verification batch (parity diff, spec-lint, `go test ./internal/config/...`, full `go test ./...`).

---

## §E. Milestones (priority-ordered)

### M1 — Author the invocation-model doctrine rule (Priority High)
- Covers REQ-IM-001..005.
- Files (mirror pair):
  - `.claude/rules/moai/workflow/native-invocation-model.md` (new)
  - `internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md` (new, mirror)
- No test (policy-layer doc). Verified by AC presence checks (acceptance.md).

### M2 — Feedback target-repo config-ization (TDD) (Priority High)
- Covers REQ-IM-010, REQ-IM-011 (and unblocks M3 item 4).
- Files (Go — NOT mirrored):
  - `internal/config/types.go` (add `FeedbackConfig` + wire into `Config`)
  - `internal/config/defaults.go` (default `modu-ai/moai-adk`)
  - `internal/config/workflow_accessors.go` OR new `internal/config/feedback_accessors.go` (resolver/accessor)
  - `internal/config/<feedback_config>_test.go` (new TDD test — RED first; MUST be an integration/file-load test that loads a real `feedback.yaml`, per AC-IM-012)
  - `internal/config/types.go` — add a `feedbackFileWrapper` per-file wrapper struct (parallel to the existing `researchFileWrapper` pattern, types.go ~L1100)
  - `internal/config/loader.go` — add a `loadFeedbackSection` per-section loader (parallel to `loadResearchSection` via `loadYAMLFile`) and wire it into `Loader.Load()` (LIVE FACT D4: loader uses explicit per-section loaders + per-file wrappers, NOT a directory scan — this wiring is a DEFINITE deliverable)
- Files (mirror pair):
  - `.moai/config/sections/feedback.yaml` (new)
  - `internal/template/templates/.moai/config/sections/feedback.yaml` (new, mirror)
- Test: `go test ./internal/config/...` GREEN with the new resolution test.

### M3 — Feedback workflow enhancement (Priority Medium)
- Covers REQ-IM-006, 007, 008, 009, 014 (+ item 4 uses M2's config).
- Files (mirror pair):
  - `.claude/skills/moai/workflows/feedback.md`
  - `internal/template/templates/.claude/skills/moai/workflows/feedback.md`
- No Go test (workflow doc). Verified by AC content checks + boundary grep (no AskUserQuestion inside the skill for duplicate resolution).

### M4 — Build, parity, and verification (Priority High — closes run-phase)
- `make build` regenerates embedded templates.
- Verification batch: local↔template `diff` parity (both feedback.md and the new rule + config section), `moai spec lint` on the SPEC, `go test ./internal/config/...`, full `go test ./...`, and the template-neutrality guard (`go test ./internal/template/...`).

---

## §F. Risks

- **R1 — Classification drift**: the PROGRAMMATIC/HUMAN-ONLY tags could be asserted without observing the live `skills.md`/`commands.md`. Mitigation: run-phase MUST verify each tag against the official docs before committing (verification-claim-integrity §1.1). WebFetch the official docs at run-phase.
- **R2 — Config-section registration (RESOLVED — D4)**: the loader uses explicit per-section loaders + per-file wrapper structs (confirmed LIVE FACT), so a new `feedback.yaml` is silently ignored UNLESS the `feedbackFileWrapper` + `loadFeedbackSection` + `Loader.Load()` wiring is added. This is now a DEFINITE M2 deliverable (§E M2), not a conditional risk; the integration/file-load test (AC-IM-012) verifies the wiring took effect.
- **R3 — Diagnostic-attachment overreach**: attaching "recent workflow state" risks leaking user file contents. Mitigation: REQ-IM-007 hard boundary — tool-diagnostic info only; error-context is orchestrator-mediated best-effort, not a filesystem scrape.
- **R4 — Template-neutrality false-positive**: the doctrine rule cites native command names and the `modu-ai/moai-adk` default. Mitigation: `modu-ai/moai-adk` is an ALLOWED system identifier per CLAUDE.local.md §25; native command names are public Claude Code identifiers (allowed). No internal SPEC IDs / REQ tokens / dates / SHAs in the template-bound files.

---

## §G. Anti-Patterns to Avoid

- Asserting a PROGRAMMATIC/HUMAN-ONLY tag without observing the official doc (unobserved-claim — verification-claim-integrity).
- Editing the local feedback.md without mirroring to the template (Template-First violation).
- Adding an AskUserQuestion call inside feedback.md for duplicate resolution (boundary violation — REQ-IM-014).
- Scope-creeping into Axis-A reuse refactoring (clean→/simplify, review→/code-review) — explicitly out of scope.
- Hardcoding the target repo in a second place instead of resolving the config value.

---

## §H. Resolved Design Decisions

All three formerly-open design questions were resolved by the orchestrator/user prior to run-phase. No open questions remain.

1. **Doctrine rule file path — RESOLVED**: `.claude/rules/moai/workflow/native-invocation-model.md` (template mirror `internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md`). The `workflow/` subdirectory was chosen, sibling to `dynamic-workflows.md` / `goal-directive.md`. Reflected in §D1, §E M1, and acceptance.md AC-IM-015.
2. **Feedback config location + key — RESOLVED**: new section file `.moai/config/sections/feedback.yaml` (template mirror `internal/template/templates/.moai/config/sections/feedback.yaml`), key `feedback.repository`, default value `modu-ai/moai-adk`, resolved by a `FeedbackConfig` Go struct (M2 TDD). `internal/config/**` Go code is NOT a mirror target; the `.moai/config/sections/feedback.yaml` asset IS a mirror target. Reflected in §D2, §E M2, and acceptance.md AC-IM-011/012/016.
3. **Diagnostic error-context scope — RESOLVED (HYBRID)**: the feedback skill ALWAYS collects the 2 guaranteed items (`moai version`, `uname`) — the skill-guaranteed set and the testable contract of REQ-IM-006 — and additionally attempts `go version` (tool build-provenance) best-effort (its absence is NOT an AC failure, per the §15 neutrality note). The orchestrator MAY additionally pass last-failed-command / error context into the invocation prompt; the skill appends it if present (best-effort additive, absence is NOT an AC failure). The skill never reads session error history and never attaches arbitrary user file contents (REQ-IM-007). Reflected in REQ-IM-006, §D3 item 1, and acceptance.md AC-IM-007.
