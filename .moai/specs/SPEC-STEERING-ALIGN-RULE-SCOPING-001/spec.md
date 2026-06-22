---
id: SPEC-STEERING-ALIGN-RULE-SCOPING-001
title: "Steering-Align: path-scope file-touch-triggered always-loaded rules + exclude legal-attribution rule from always-load"
version: "0.2.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "steering, rule-scoping, always-loaded, context-budget, template-first"
tier: S
era: V3R6
---

## HISTORY

- 2026-06-22 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier S, Section A-H, 4 artifacts). Entry SPEC of Epic Steering-Align. `status: draft`.
- 2026-06-22 — v0.1.1 — manager-spec — iter-2 audit revision (plan-auditor PASS-WITH-DEBT 0.82, Tier S thresh 0.75). D1 (BLOCKING): corrected the false "both trees carry 15 always-loaded" premise — the TEMPLATE tree carries only **13** always-loaded (59 total), because `workflow/lifecycle-sync-gate.md` AND `workflow/runtime-recovery-doctrine.md` are LIVE-ONLY (deliberately not template-mirror-enrolled). Edit targets split into MIRRORED (3: hook-independence, prompting-best-practices, NOTICE) vs LIVE-ONLY (1: lifecycle-sync-gate). D2: added `tier: S` frontmatter. D4: parity AC rewritten to assert both files exist before diffing, scoped to MIRRORED targets only. D3: progress.md §E.1 AC summary now lists AC-SARS-007. D5: normalized Class-A globs to the `**/` precedent prefix. All re-verified by command in §F.1.
- 2026-06-22 — v0.2.0 — manager-develop — Run-phase complete (cycle_type=tdd, command-based AC verification). M1-M5 done: 3 MIRRORED template edits + `make build` re-embed + 3 MIRRORED live edits (byte-identical parity) + 1 LIVE-ONLY edit (lifecycle-sync-gate). 8/8 AC PASS (LIVE 15→11, TEMPLATE 13→10, MIRRORED parity OK, frontmatter-only diff 28+/0-, NOTICE excluded+retained, byte-sum LIVE 211495→159761 / TEMPLATE 156308→128176, 0 Class-B/C scoped, lifecycle-sync-gate LIVE-ONLY confirmed). D6 fold-in added to §F.3. go build/go vet exit 0, internal/template test suite green (mirror-parity guard passing), spec-lint clean. `status: draft → in-progress`.

---

## A. Context / Background

Two Anthropic official documents were audited against moai-adk's rule layer:

1. **Blog — "Steering Claude Code: skills, hooks, rules, subagents and more"**. Canonical guidance: *"If a rule only applies to `src/api/**`, scoping it with `paths:` keeps it out of context during unrelated work."* The corollary anti-pattern: an UNSCOPED rule that constrains a directory- or context-specific concern wastes always-on context tokens on every turn, including turns that never touch the relevant files.

2. **best-practices — "Write an effective CLAUDE.md"**. Canonical guidance: *"Bloated [always-loaded instructions] cause Claude to ignore your actual instructions."* The per-line test: *"Would removing this cause Claude to make mistakes? If not, cut it."*

### A.1 Observed ground-truth (re-verified live; commands + output recorded in §F.1 / acceptance.md)

The LIVE tree and the TEMPLATE tree are NOT identical sets — the template tree is a deliberate SUBSET. This was the iter-1 BLOCKING defect (D1): an earlier draft falsely claimed "both trees carry 15 always-loaded rules". The corrected, re-verified counts:

| Tree | Total rule files | Always-loaded (no `paths:`) | Path-scoped |
|------|------------------|------------------------------|-------------|
| LIVE (`.claude/rules/moai/`) | **61** | **15** | 46 |
| TEMPLATE (`internal/template/templates/.claude/rules/moai/`) | **59** | **13** | 46 |

The 2-file difference (LIVE 15 always-loaded − TEMPLATE 13 always-loaded) is exactly: `workflow/lifecycle-sync-gate.md` AND `workflow/runtime-recovery-doctrine.md`. Both are present in the LIVE tree but ABSENT from the TEMPLATE tree (deliberately live-only, not template-mirror-enrolled). Verified per-file in §F.1.

- The exact `paths:` frontmatter syntax used in-repo is a comma-separated quoted glob string, e.g. `paths: "**/*.go,**/go.mod,**/go.sum"` (`.claude/rules/moai/languages/go.md`) and `paths: "**/.claude/agents/**"` (`.claude/rules/moai/development/agent-authoring.md`). Note the `**/` precedent prefix.
- `.claude/rules/` is embedded into the binary via `//go:embed all:templates` (`internal/template/embed.go:28`). The TEMPLATE SSOT is `internal/template/templates/.claude/rules/moai/...`; the live `.claude/rules/moai/...` is the deployed copy. The template tree is a subset of the live tree (live-only files like `lifecycle-sync-gate.md` exist in deployed projects but are not redistributed via the template).

### A.1b MIRRORED vs LIVE-ONLY edit-target split (D1 fix)

Because the template tree is a subset, the 4 edit targets split into two handling classes. Per-target template-presence verified in §F.1.

| Edit target | Class | Template mirror? | Handling |
|-------------|-------|------------------|----------|
| `development/hook-independence.md` | A | MIRRORED (template PRESENT) | Template-first + `make build` re-embed + live-tree parity |
| `development/prompting-best-practices.md` | A | MIRRORED (template PRESENT) | Template-first + `make build` re-embed + live-tree parity |
| `NOTICE.md` | D | MIRRORED (template PRESENT) | Template-first + `make build` re-embed + live-tree parity |
| `workflow/lifecycle-sync-gate.md` | A | **LIVE-ONLY (template ABSENT)** | **Live-file edit ONLY** — NO template edit, NO `make build` re-embed for it, NO parity check; AC asserts its live-only nature explicitly |

`runtime-recovery-doctrine.md` is the OTHER live-only file but it is a Class-C rule (NOT an edit target) — its live-only status is documented here only because it shares the LIVE-15/TEMPLATE-13 gap with `lifecycle-sync-gate.md` and corroborates that the template tree is a deliberate subset.

### A.2 The core analysis (the nuance)

Path-scoping in Claude Code triggers a rule to load when Claude TOUCHES a matching file. A rule is therefore cleanly path-scopable ONLY when its relevance has a **file-touch trigger**. Each of the 15 always-loaded rules was classified by reading its own header / "Loading scope" note:

| Class | Meaning | Rules | Disposition |
|-------|---------|-------|-------------|
| **A** | CLEAN file-touch trigger → add `paths:` | `hook-independence.md`, `prompting-best-practices.md`, `lifecycle-sync-gate.md` | IN SCOPE — add `paths:` |
| **B** | NO clean file-touch trigger (env- or intent-triggered) → scoping infeasible | `glm-web-tooling.md`, `dynamic-workflows.md`, `goal-directive.md`, `verification-batch-pattern.md` | OUT OF SCOPE (see §J) |
| **C** | GENUINELY cross-cutting → correctly always-loaded | `agent-common-protocol.md`, `askuser-protocol.md`, `verification-claim-integrity.md`, `context-window-management.md`, `runtime-recovery-doctrine.md`, `session-handoff.md`, `sprint-round-naming.md` | OUT OF SCOPE (see §J) |
| **D** | ZERO in-context behavioral value → exclude from always-load | `NOTICE.md` | IN SCOPE — exclude from always-load |

**Grounding of Class-A classifications** (each rule self-declares its trigger):

- `hook-independence.md` opens: *"Doctrine for the MoAI hook layer (`.claude/hooks/moai/`)."* → relevant exactly when a hook file is touched.
- `prompting-best-practices.md` Loading-scope line: *"read when authoring or tuning an agent prompt / skill body / system prompt."* → relevant when `.claude/agents/**` or `.claude/skills/**` is touched.
- `lifecycle-sync-gate.md` enforcement: *"`internal/spec/era.go` `ClassifyEra()`"* and SPEC-lifecycle concerns near `.moai/specs/`. → relevant when `internal/spec/**` or `.moai/specs/**` is touched.

**Grounding of the Class-D classification**: `NOTICE.md` is a third-party legal attribution file (Apache-2.0 / Karpathy / im-not-ai / Anthropic-2026 citations). Applying the best-practices per-line test — *"Would removing it cause Claude to make mistakes?"* — the answer is **No**. Legal attribution belongs in the distribution but provides zero behavioral steering value when loaded into the model context on every turn.

### A.3 Epic Steering-Align context

This is the **entry SPEC** of Epic Steering-Align — a 5-SPEC roadmap aligning moai-adk to Anthropic's official Claude Code "steering" guidance. The other 4 SPECs are FUTURE (named here for Epic context only; this SPEC authors artifacts ONLY for rule-scoping):

1. **SPEC-STEERING-ALIGN-RULE-SCOPING-001** (this SPEC) — path-scope file-touch-triggered always-loaded rules + exclude legal-attribution rule.
2. (FUTURE) CLAUDE.md diet — apply the best-practices per-line test to CLAUDE.md's always-loaded body.
3. (FUTURE) agent-orchestration guardrail determinization — convert intent-triggered orchestration prose into deterministic guardrails.
4. (FUTURE) output-style slimming — trim the always-loaded output-style body.
5. (FUTURE) CLAUDE.local.md diet — apply the per-line test to the maintainer-local always-loaded body.

---

## B. Requirements (GEARS notation)

### B.1 Path-scope the Class-A rules

- **REQ-SARS-001 (Ubiquitous)**: The rule-scoping change SHALL add a `paths:` frontmatter field to each of the three Class-A rules (`hook-independence.md`, `prompting-best-practices.md`, `lifecycle-sync-gate.md`), using the comma-separated quoted-glob syntax already established in-repo.

- **REQ-SARS-002 (Ubiquitous)**: The `hook-independence.md` rule SHALL carry `paths: "**/.claude/hooks/**"` matching the hook layer it governs, consistent with the existing `hooks-system.md` precedent (`paths: "**/.claude/hooks/**,..."`). The `**/` prefix matches the in-repo precedent form.

- **REQ-SARS-003 (Ubiquitous)**: The `prompting-best-practices.md` rule SHALL carry `paths: "**/.claude/agents/**,**/.claude/skills/**"` matching the agent-prompt / skill-body authoring surfaces it serves, consistent with the existing `agent-authoring.md` (`paths: "**/.claude/agents/**"`) and `skill-authoring.md` (`paths: "**/.claude/skills/**/SKILL.md"`) precedents. The `**/` prefix matches the in-repo precedent form.

- **REQ-SARS-004 (Ubiquitous)**: The `lifecycle-sync-gate.md` rule (LIVE-ONLY — see §A.1b) SHALL carry `paths: "**/internal/spec/**,**/.moai/specs/**"` matching the SPEC-lifecycle surfaces it governs. The `**/` prefix is normalized to the in-repo precedent form (D5 fix — the iter-1 draft omitted the `**/` prefix). Post-normalization value-context confirmation: `**/internal/spec/**` still matches the `era.go`/`audit.go` enforcement surface (`internal/spec/era.go`), and `**/.moai/specs/**` still matches the SPEC directory the lifecycle gate governs — the `**/` prefix only broadens matching (anchor-anywhere), never narrows it, so both contexts remain covered.

- **REQ-SARS-005 (Event-driven)**: **When** Claude touches a file matching a newly-scoped rule's `paths` glob, the rule SHALL load into context — i.e. behavior is PRESERVED for relevant work, only its always-on residency is removed.

### B.2 Exclude the Class-D rule from always-load

- **REQ-SARS-006 (Ubiquitous)**: The `NOTICE.md` legal-attribution rule SHALL no longer be always-loaded into the model context. The file itself SHALL be retained on disk (legal attribution is a distribution requirement); only its always-load residency is removed (via a trivially-rare `paths:` scope OR relocation out of the always-load set — the exact mechanism is decided in plan.md §D).

### B.3 Content-preservation and parity

- **REQ-SARS-007 (Ubiquitous)**: The change SHALL be frontmatter-only. No rule BODY content SHALL be modified — only the `paths:` frontmatter field is added (or, for `NOTICE.md`, the chosen exclusion mechanism is applied). This keeps the change low-risk and reviewable.

- **REQ-SARS-008 (Where MIRRORED)**: **Where** an edit target is MIRRORED (template mirror PRESENT — the 3 MIRRORED targets per §A.1b: hook-independence.md, prompting-best-practices.md, NOTICE.md), the change SHALL be applied to BOTH the template SSOT tree and the live deployed tree, with identical `paths:` frontmatter in both. The template tree is edited first per the CLAUDE.local.md §2 Template-First rule, then re-embedded via `make build`. The parity check (AC-SARS-003) applies ONLY to MIRRORED targets.

- **REQ-SARS-008b (Where LIVE-ONLY)**: **Where** an edit target is LIVE-ONLY (template mirror ABSENT — `lifecycle-sync-gate.md` per §A.1b), the change SHALL be applied ONLY to the live deployed tree. The change MUST NOT fabricate a template edit for a LIVE-ONLY target (there is no template file to edit), MUST NOT run `make build` re-embed on its behalf, and MUST NOT be subject to a template/live parity check. An AC asserts the LIVE-ONLY nature explicitly (acceptance.md AC-SARS-008).

- **REQ-SARS-009 (State-driven, per-tree delta)**: **While** the change is in effect, the always-loaded rule count SHALL drop per-tree as follows, each measurable by a re-runnable `find`/`grep` command (acceptance.md AC-SARS-001):
  - LIVE tree: **15 → 11** (−3 Class-A live edits: hook-independence, prompting-best-practices, lifecycle-sync-gate; −1 Class-D: NOTICE).
  - TEMPLATE tree: **13 → 10** (−2 Class-A mirrored edits: hook-independence, prompting-best-practices; −1 Class-D: NOTICE). `lifecycle-sync-gate.md` is NOT in the template tree, so it cannot be removed from the template always-loaded set — hence the template drops by 3, not 4.

### B.4 Honesty constraints (no over-claiming)

- **REQ-SARS-010 (Unwanted behavior)**: The change SHALL NOT add `paths:` to any Class-B or Class-C rule. Class-B rules have no clean file-touch trigger (env- or intent-triggered) and Class-C rules are genuinely cross-cutting; scoping either would either (b) silently suppress a rule when its env/intent trigger fires without a file touch, or (c) drop a cross-cutting rule that must apply to every turn. The SPEC MUST NOT pretend Class-B is cleanly scopable.

---

## C. Constraints

- **C-1** [HARD] Template-First (CLAUDE.local.md §2) applies to MIRRORED targets ONLY: edit `internal/template/templates/.claude/rules/moai/...` FIRST, then `make build` to re-embed, then verify live-tree parity. For the LIVE-ONLY target (`lifecycle-sync-gate.md`), there is NO template file — Template-First does not apply; edit the live file directly (REQ-SARS-008b).
- **C-2** [HARD] Frontmatter-only: no rule body edits (REQ-SARS-007). The diff for each Class-A rule is a 3-line `---\npaths: "..."\n---` insertion (none of the 4 targets currently has frontmatter, verified live).
- **C-3** [HARD] No new lint rule, no Go code change. This SPEC is mechanical frontmatter editing only.
- **C-4** Constitution alignment: the `paths:` glob syntax MUST match the in-repo precedent exactly (comma-separated quoted string, `**/` prefix as used by `agent-authoring.md` / `hooks-system.md`). All 4 globs in REQ-SARS-002/003/004 carry the `**/` prefix (D5).
- **C-5** `NOTICE.md` content DIFFERS between trees (template 4639 B vs live 9580 B — the live copy carries dev-local Karpathy/im-not-ai/Anthropic-2026 additions). The exclusion mechanism (REQ-SARS-006) MUST be applied to both trees regardless of body divergence; the mechanism touches frontmatter/residency only, not the divergent body.

---

## D. Out of Scope / Exclusions

The SPEC scope is exactly Class-A (3 rules) + Class-D (1 rule). The following are explicitly excluded.

### Out of Scope — Class-B rules (no clean file-touch trigger)

- `glm-web-tooling.md` — relevant only on the GLM backend (env `ANTHROPIC_BASE_URL` contains `api.z.ai`), NOT a file-path touch. Path-scoping is INFEASIBLE.
- `dynamic-workflows.md` — relevant on orchestration INTENT (fanning out a large task), NOT a file touch. Path-scoping is INFEASIBLE.
- `goal-directive.md` — relevant on orchestration INTENT (setting a `/goal`), NOT a file touch. Path-scoping is INFEASIBLE.
- `verification-batch-pattern.md` — relevant at run-phase completion INTENT (orchestrator verifying implementation), NOT a file touch. Path-scoping is INFEASIBLE.
- For these, the real remedy is (b1) keep always-loaded but TRIM, or (b2) convert to an on-demand skill/reference — a separate concern (skill conversion ≠ frontmatter scoping). Deferred to a follow-up SPEC; this SPEC MUST NOT pretend Class-B is path-scopable.

### Out of Scope — Class-C rules (genuinely cross-cutting, correctly always-loaded)

- `agent-common-protocol.md`, `askuser-protocol.md`, `verification-claim-integrity.md`, `context-window-management.md`, `runtime-recovery-doctrine.md` — cross-cutting doctrines that bind every turn regardless of which file is touched.
- `session-handoff.md` — self-justifies always-loaded: its Loading-scope note states *"Intentionally always-loaded (no `paths:` restriction) because Trigger #3 (user explicit session-end) can fire from any session context."* Scoping it would break Trigger #3.
- `sprint-round-naming.md` — borderline (authoring/banner concern); assessed as cross-cutting because the Epic/Milestone taxonomy applies to banner localization and paste-ready resume on any turn, neither of which has a file-touch trigger. KEEP always-loaded.

### Out of Scope — the other 4 Epic Steering-Align SPECs

- CLAUDE.md diet, agent-orchestration guardrail determinization, output-style slimming, CLAUDE.local.md diet — named in §A.3 for Epic context only. This SPEC authors artifacts ONLY for rule-scoping.

### Out of Scope — implementation mechanics deferred to run-phase

- Rule BODY edits, new lint rules, Go code changes, skill-conversion of Class-B rules, and any CLAUDE.md / output-style edits.

---

## E. Acceptance Criteria Reference

Concrete GEARS-format acceptance criteria with re-runnable verification commands live in `acceptance.md`. Summary of AC themes:

- AC-SARS-001 — per-tree always-loaded count drop: LIVE 15 → 11, TEMPLATE 13 → 10 (re-runnable `find`/`grep`, asserted separately per tree).
- AC-SARS-002 — each newly-scoped Class-A rule still LOADS on a matching file touch (behavior preserved; `**/`-prefixed globs present).
- AC-SARS-003 — template + live `paths:` frontmatter identical, asserted ONLY for the 3 MIRRORED targets, with both-files-exist guard before diffing (D4).
- AC-SARS-004 — no rule BODY content changed (frontmatter-only diff).
- AC-SARS-005 — `NOTICE.md` excluded from always-load, file retained on disk (both trees).
- AC-SARS-006 — always-on byte-sum reduced per tree (LIVE and TEMPLATE separately).
- AC-SARS-007 — honesty guardrail: no Class-B/Class-C rule scoped (stay-always-loaded set unchanged).
- AC-SARS-008 — `lifecycle-sync-gate.md` is LIVE-ONLY: scoped in live tree, ABSENT from template tree, NOT subject to parity (D1).

---

## F. Evidence (verification-claim-integrity)

### F.1 Re-verified ground-truth (command → observed output, this tree, 2026-06-22)

| Claim | Command | Observed |
|-------|---------|----------|
| LIVE: 61 total rule files | `find .claude/rules/moai -name '*.md' \| wc -l` | `61` |
| LIVE: 15 always-loaded | `for f in $(find .claude/rules/moai -name '*.md'); do grep -q '^paths:' "$f" \|\| echo "$f"; done \| wc -l` | `15` (the 15 enumerated in §A.2) |
| LIVE: 46 path-scoped | `for f in ...; do grep -q '^paths:' "$f" && echo "$f"; done \| wc -l` | `46` |
| TEMPLATE: 59 total rule files | `find internal/template/templates/.claude/rules -name '*.md' \| wc -l` | `59` |
| TEMPLATE: 13 always-loaded (D1) | `for f in $(find internal/template/templates/.claude/rules -name '*.md'); do grep -q '^paths:' "$f" \|\| echo "$f"; done \| wc -l` | `13` |
| Per-target template presence (D1) | `for r in development/hook-independence.md development/prompting-best-practices.md workflow/lifecycle-sync-gate.md NOTICE.md; do [ -f "internal/template/templates/.claude/rules/moai/$r" ] && echo "MIRRORED $r" \|\| echo "LIVE-ONLY $r"; done` | `MIRRORED hook-independence.md` / `MIRRORED prompting-best-practices.md` / `LIVE-ONLY lifecycle-sync-gate.md` / `MIRRORED NOTICE.md` |
| 2 live-only files (D1) | `for r in workflow/lifecycle-sync-gate.md workflow/runtime-recovery-doctrine.md; do [ -f "internal/template/templates/.claude/rules/moai/$r" ] && echo "template PRESENT" \|\| echo "template ABSENT"; done` | both `template ABSENT` (live PRESENT) |
| `paths:` syntax + `**/` prefix | `sed -n '1,3p' .claude/rules/moai/languages/go.md` | `paths: "**/*.go,**/go.mod,**/go.sum"` |
| agents/skills precedent | `sed -n '2p' .../agent-authoring.md`; `.../skill-authoring.md` | `paths: "**/.claude/agents/**"`; `paths: "**/.claude/skills/**/SKILL.md"` |
| hooks precedent | `sed -n '2p' .../core/hooks-system.md` | `paths: "**/.claude/hooks/**,**/.claude/settings.json,**/.claude/settings.local.json"` |
| rules embedded | `grep -rn 'go:embed' internal/template/*.go` | `internal/template/embed.go:28://go:embed all:templates` |
| LIVE always-loaded byte-sum (before) | byte-sum of 15 live always-loaded files | `211495` bytes |
| TEMPLATE always-loaded byte-sum (before) | byte-sum of 13 template always-loaded files | `156308` bytes |
| tier:S convention (D2) | `grep -rh '^tier:' .moai/specs/SPEC-*/spec.md \| sort \| uniq -c` | `58 tier: S` (dominant convention; e.g. `SPEC-CC2178-DOCS-ALIGN-001`) |

### F.2 Gaps (explicitly NOT observed at plan-phase)

- The token reduction is APPROXIMATED via byte-sum, not a real tokenizer count (acceptable per the per-line-test framing; exact token count deferred to run-phase if needed).
- The exact `NOTICE.md` exclusion mechanism (trivially-rare `paths:` vs relocation) is NOT decided here — it is a plan.md §D decision.
- The Claude Code runtime behavior "rule loads on matching file touch" is documented Anthropic behavior, NOT independently re-measured in this tree (it is the documented `paths:` semantics that the 46 existing scoped rules already rely on).

### F.3 Residual risk

- Globs that are too narrow could fail to load a Class-A rule when its relevant work happens via a path the glob misses (e.g. a hook authored outside `.claude/hooks/`). Mitigated by mirroring the EXACT precedent globs (`hooks-system.md`, `agent-authoring.md`) which are already battle-tested in-repo.
- `make build` re-embed could surface unrelated template drift; mitigated by scoping the run-phase diff to the 4 rules' frontmatter only.
- The LIVE-ONLY target (`lifecycle-sync-gate.md`) and its newly-added `paths:` scope do NOT survive a downstream `moai update` (whole-file deletion of any rule absent from the embedded template) — this is a PRE-EXISTING condition owned by the live-only mirror-enrollment policy, not introduced by this SPEC (D6 fold-in).

---

## G. SPEC ID Pre-Write Self-Check (recorded per protocol)

decomposition: SPEC ✓ | STEERING ✓ | ALIGN ✓ | RULE ✓ | SCOPING ✓ | 001 ✓ → PASS

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`: first segment literal `SPEC`; middle segments STEERING/ALIGN/RULE/SCOPING each match `[A-Z][A-Z0-9]*`; last segment `001` matches `\d{3}` (digit-only, no alpha suffix). PASS.

---

## H. Cross-References

- Anthropic blog "Steering Claude Code: skills, hooks, rules, subagents and more" (`paths:` scoping guidance).
- Anthropic best-practices "Write an effective CLAUDE.md" (per-line test, bloat warning).
- CLAUDE.local.md §2 — Template-First rule (`internal/template/templates/` SSOT → `make build` → live).
- `.claude/rules/moai/languages/go.md`, `.claude/rules/moai/development/agent-authoring.md`, `.claude/rules/moai/development/skill-authoring.md`, `.claude/rules/moai/core/hooks-system.md` — in-repo `paths:` precedents.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec.
