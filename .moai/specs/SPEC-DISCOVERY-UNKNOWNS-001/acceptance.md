---
id: SPEC-DISCOVERY-UNKNOWNS-001
title: "Unknowns framework Tier-1 for Context-First Discovery — Blind Spot Pass + decision-reversibility ordering + 4-quadrant lens"
version: "0.1.0"
status: in-progress
created: 2026-07-05
updated: 2026-07-12
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tier: M
tags: "discovery, unknowns, blind-spot, gears, askuser, planning, context-first, doc"
---

# SPEC-DISCOVERY-UNKNOWNS-001 — Acceptance Criteria

## §D. AC Matrix

Each AC is independently testable. Commands run from the project root (`/Users/goos/MoAI/moai-adk-go`) after run-phase edits + `make build`. `<local>` / `<mirror>` shorthand: the local file and its `internal/template/templates/` mirror per spec.md §A.4.

> **Command-format note (audit D1 fix)**: every verification command is placed in a fenced `bash` block, NOT a markdown table cell. Under ERE (`grep -E`/`-iE`) a `\|` is a LITERAL pipe, not alternation (verified on this platform: `printf 'optional' | grep -iE 'optional\|x'` → NO MATCH), so alternation must be written with a raw `|`. A raw `|` breaks a markdown table cell (column separator), so the commands live in fenced blocks where `|` is preserved literally and every command is copy-paste-correct.

### AC-DU-001 (REQ-DU-001) — named Blind Spot Pass subsection exists
```bash
grep -cE '^#{2,3} .*Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md
```
PASS: ≥ 1 (a named H2/H3 "Blind Spot Pass" subsection is present).

### AC-DU-016 (REQ-DU-002) — Blind Spot Pass trigger text present (unfamiliar-domain + before-plan-phase-entry)
```bash
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'unfamiliar'
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'before plan-phase|before the plan phase|pre-plan'
```
PASS: both match — the subsection states the trigger is an unfamiliar domain AND that the pass runs before plan-phase entry (covers REQ-DU-002, whose trigger/timing was previously unverified).

### AC-DU-002 (REQ-DU-003) — mechanism = Agent(Explore) read-only scan
```bash
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -E 'Agent\(Explore\)'
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'read-only'
```
PASS: both match.

### AC-DU-003 (REQ-DU-004) — optionality stated
```bash
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'optional|not a mandatory gate'
```
PASS: ≥ 1 match.

### AC-DU-004 (REQ-DU-005) — subagent boundary reaffirmed
```bash
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'AskUserQuestion'
grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md | grep -iE 'not prompt|never prompt|does not prompt'
```
PASS: both match (surface via AskUserQuestion; Explore does not prompt the user directly).

### AC-DU-005 (REQ-DU-006) — wired from Plan Phase + §7 Rule 5 (audit D4 fix — section-scoped)
```bash
# D4 fix: REQ-DU-006 requires the ref specifically in the Plan Phase section of spec-workflow.md
# AND in §7 Rule 5 of CLAUDE.md. A whole-file `grep -c` vacuously passes on a match ANYWHERE in
# the file, so both greps are section-scoped. The awk range extracts exactly the "## Plan Phase"
# section (from its heading until the next `## ` H2 heading); the `grep -A3` window scopes to the
# §7 Rule 5 bullet (heading line included — a Blind Spot Pass ref inline on the Rule 5 line is
# intentionally matched here, unlike the AC-DU-007 heading-strip case).
awk '/^## Plan Phase$/{f=1;next} /^## /{f=0} f' .claude/rules/moai/workflow/spec-workflow.md | grep 'Blind Spot Pass'
grep -A3 'Rule 5 — Context-First' CLAUDE.md | grep 'Blind Spot Pass'
```
PASS: both match — the Blind Spot Pass ref is present INSIDE the Plan Phase section (not merely somewhere in spec-workflow.md) AND within the §7 Rule 5 window.

### AC-DU-006 (REQ-DU-007) — plan.md decision-reversibility ordering guidance in manager-spec.md
```bash
grep -iE 'most likely to change|decision-reversibility' .claude/agents/moai/manager-spec.md
grep -iE 'mechanical|refactor' .claude/agents/moai/manager-spec.md | grep -iE 'bottom|defer|last'
```
PASS: both match (lead-with-change-likelihood + defer-mechanical).

### AC-DU-007 (REQ-DU-008) — CLAUDE.md §7 Rule 1 leads with high-change decisions (audit D1 fix — non-vacuous)
```bash
# D1 vacuity fix: the prior `...|first` alternation matched the "Rule 1 — Approach-First"
# HEADING (which literally contains "First"), so it vacuously PASSED on the UNMODIFIED
# CLAUDE.md and REQ-DU-008 would ship unverified. Fix = (a) drop the bare `first` term;
# (b) strip the heading line via `tail -n +2` so the "Approach-First" heading cannot match;
# (c) anchor to a genuinely-NEW token that only appears once the run-phase adds
# decision-reversibility content.
# [RUN-PHASE PLACEMENT CONTRACT — see plan.md M2]: the T2 Rule 1 content MUST be added on a
# line AFTER the Rule 1 heading (within the -A2 window), NOT inline on the heading line —
# `tail -n +2` strips the heading line, so inline-on-heading content would not be detected.
grep -A2 'Rule 1 — Approach-First' CLAUDE.md | tail -n +2 | grep -iE 'most likely to change|highest change[ -]likelihood|change-likelihood'
```
PASS: ≥ 1 match AFTER the run-phase T2 edit. Non-vacuity proof (verified during D1 remediation against the UNMODIFIED CLAUDE.md): the prior `...|first` form MATCHED the Approach-First heading (vacuous PASS); this fixed form returns NO match on the unmodified tree (the change-likelihood tokens are absent from Rule 2/Rule 3), so it fails-closed until the run-phase edit lands.

### AC-DU-008 (REQ-DU-009, REQ-DU-013) — no Go change in THIS SPEC's commit set (audit D4 fix)
```bash
# Diff-based, scoped to this SPEC's commits — NOT a blanket `-newer` mtime find
# (which false-fails against any parallel-session .go edit newer than spec.md).
# Run after run-phase commits land:
git log --grep='SPEC-DISCOVERY-UNKNOWNS-001' --name-only --format='' -- '*.go' | grep -c '\.go$'
```
PASS: 0 (this SPEC's commits touch zero `.go` files — ordering rule only, no machinery). Sound companion: AC-DU-015.

### AC-DU-009 (REQ-DU-010) — all four quadrant terms present in § Ambiguity Triggers
```bash
for t in 'Known-Knowns' 'Known-Unknowns' 'Unknown-Knowns' 'Unknown-Unknowns'; do
  printf '%s=' "$t"; grep -c "$t" .claude/rules/moai/core/askuser-protocol.md
done
```
PASS: each of the four counts ≥ 1.

### AC-DU-010 (REQ-DU-011) — suspected Unknown-Unknowns routes to a Blind Spot Pass
```bash
grep -A6 'Unknown-Unknowns' .claude/rules/moai/core/askuser-protocol.md | grep 'Blind Spot Pass'
```
PASS: ≥ 1 (the taxonomy paragraph routes Unknown-Unknowns to the Blind Spot Pass).

### AC-DU-011 (REQ-DU-012) — CLAUDE.md §7 Rule 5 lens framing + SSOT pointer (diet-respecting)
```bash
grep -A3 'Rule 5 — Context-First' CLAUDE.md | grep -iE 'Known-Unknowns|4-quadrant|unknown-unknown'
```
PASS: ≥ 1 (lens framing + SSOT pointer present) AND a diet check: the T3 net addition to §7 Rule 5 is ≤ 1 line + pointer, verified against the run-phase diff of CLAUDE.md §7 Rule 5.

### AC-DU-012 (REQ-DU-014) — template-mirror parity, per-file token (audit D6 fix)
Each file receives DIFFERENT inserted text, so the parity token is enumerated per file (do not reuse `Blind Spot Pass` for the manager-spec mirror, which receives the T2 phrase):
```bash
grep -c 'Blind Spot Pass'        internal/template/templates/.claude/rules/moai/core/askuser-protocol.md    # askuser mirror (T1 + T3)
grep -ci 'most likely to change' internal/template/templates/.claude/agents/moai/manager-spec.md           # manager-spec mirror (T2)
grep -c 'Blind Spot Pass'        internal/template/templates/CLAUDE.md                                      # CLAUDE.md mirror (T1 Rule 5 ref)
grep -c 'Blind Spot Pass'        internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md   # spec-workflow mirror (T1 wire)
```
PASS: each mirror count ≥ 1, AND the same file-specific token is present in the corresponding LOCAL file (parity in both directions).

### AC-DU-013 (REQ-DU-014) — make build succeeds
```bash
make build
```
PASS: exit 0 (embedded template FS regenerated from the mirror edits).

### AC-DU-014 (REQ-DU-015) — §25 template neutrality (audit D1 fix — raw `|` alternation)
```bash
grep -rn 'SPEC-DISCOVERY-UNKNOWNS' internal/template/templates/
grep -rnE '2026-07-05|Finding [0-9]|Audit [0-9]' \
  internal/template/templates/.claude/rules/moai/core/askuser-protocol.md \
  internal/template/templates/.claude/agents/moai/manager-spec.md \
  internal/template/templates/CLAUDE.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
```
PASS: both commands return zero matches introduced by this SPEC (no internal SPEC ID / date / audit-citation leak). The raw `|` here is genuine ERE alternation — the prior `\|` form vacuously passed (matched a literal `2026-07-05|Finding...` string that never occurs), masking real leakage.

### AC-DU-015 (REQ-DU-013) — no `.go` in the working-tree change set
```bash
git status --porcelain | grep -E '\.go$'
```
PASS: empty (no `.go` file modified/added by this SPEC — doc/rule/agent-body only).

### AC-DU-017 (REQ-DU-014) — CI parity guards pass after mirror edits (audit D2 fix)
Two of the four edit targets are enrolled in ENFORCED parity tests that this SPEC's `make build`
(AC-DU-013) + grep-token (AC-DU-012) checks do NOT run: `spec-workflow.md` is in
`workflowOptMirroredPaths` (BYTE-IDENTICAL mirror) and `askuser-protocol.md` is in
`sanitizedPairPaths` (sanitized-pair doctrine-propagation). A whitespace-only mirror drift on
`spec-workflow.md`, or a partial/absent propagation of the T1+T3 doctrine into the
`askuser-protocol.md` mirror, PASSES AC-DU-012/013 but REDS main on the Route A push. Run BOTH
enforced parity tests AFTER the mirror edits land:
```bash
go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestSanitizedPairParity'
```
PASS: exit 0. `TestRuleTemplateMirrorDrift` (sentinel `RULE_TEMPLATE_MIRROR_DRIFT`) enforces byte-identity for `spec-workflow.md`; `TestSanitizedPairParity` (sentinel `SANITIZED_PAIR_PARITY_DRIFT`) enforces doctrine parity for `askuser-protocol.md` — the T1 Blind Spot Pass + T3 4-quadrant lens doctrine must land in BOTH trees. `manager-spec.md` (sanitized mirror) + `CLAUDE.md` are in NEITHER parity registry — their mirror parity is grep-token only (AC-DU-012).

## §D.1 Given-When-Then scenarios

### Scenario 1 — unfamiliar-domain plan entry triggers an optional Blind Spot Pass (T1)

- **Given** the user requests plan-phase work in a subsystem they have not worked in before, AND the orchestrator suspects unknown-unknowns (unfamiliar design/library territory),
- **When** the orchestrator reaches the plan-phase entry,
- **Then** the orchestrator runs a Blind Spot Pass — spawns `Agent(Explore)` in read-only mode to scan the relevant domain, then surfaces the user's likely unknown-unknowns through an `AskUserQuestion` round — before authoring the SPEC,
- **And** the surfacing goes only through the orchestrator's AskUserQuestion channel (Explore never prompts the user directly).

### Scenario 2 — plan.md leads with the decisions most likely to change (T2)

- **Given** a plan.md is authored for a SPEC that changes a data model, adds a new type interface, and also performs some mechanical refactors,
- **When** manager-spec orders the plan.md milestones/sections,
- **Then** the data-model, type-interface, and user-facing/UX-flow decisions appear before the mechanical refactoring steps,
- **And** the mechanical/refactoring steps are deferred to the bottom so human review focuses on the high-change-likelihood decisions.

### Scenario 3 — the 4-quadrant lens routes suspected Unknown-Unknowns to a Blind Spot Pass (T3 → T1)

- **Given** the orchestrator applies the § Ambiguity Triggers with the Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns lens and suspects Unknown-Unknowns,
- **When** it classifies the ambiguity,
- **Then** the lens directs it to initiate a Blind Spot Pass (per REQ-DU-002) rather than proceeding directly to plan.

## §D.2 Edge cases

- **Familiar domain / no suspected unknown-unknowns** → the Blind Spot Pass is NOT run; optionality is preserved and no forced overhead is incurred (guards REQ-DU-004). Verifiable: the subsection text explicitly gates on "suspected unknown-unknowns", not on every plan entry.
- **Template-mirror divergence** → `askuser-protocol.md` diverges local-vs-template by ~1 hunk and `manager-spec.md` by ~5 hunks (all §25-neutrality strips of internal dates / SPEC-IDs / REQ tokens); neither divergence touches this SPEC's T1/T2 edit anchors. The paired edit targets the semantically-equivalent anchor in each surface; the new phrase must still appear in BOTH (AC-DU-012).
- **CLAUDE.md active diet** → §7 Rule 5 gains ≤ ~2 lines total (T1 ref + T3 lens framing + pointer); §7 Rule 1 gains 1 line (T2). No section-block expansion (AC-DU-011 diet check).
- **Parallel-session working tree** → ~52 uncommitted files from other SPECs are present; edits touch only the 8 enumerated targets + this SPEC dir; AC-DU-015 asserts no unrelated `.go` change is attributed to this SPEC.

## §D.3 Quality gate criteria / Definition of Done

- [ ] All 17 ACs (AC-DU-001 .. AC-DU-017) PASS with observed command output recorded in the run-phase E1 matrix.
- [ ] `make build` exits 0 (AC-DU-013).
- [ ] Template parity holds for all 4 surfaces — every new phrase present in local AND mirror (AC-DU-012).
- [ ] §25 neutrality: zero internal SPEC ID / date / SHA / audit-citation leakage into template content (AC-DU-014).
- [ ] No Go source, no new agent, no new skill, no new subsystem introduced (AC-DU-008 + AC-DU-015).
- [ ] Blind Spot Pass is optional (AC-DU-003) and preserves the subagent boundary (AC-DU-004).
- [ ] CLAUDE.md diet respected — §7 Rule 5 addition ≤ ~2 lines + pointer (AC-DU-011).
- [ ] `spec-lint` clean on the new SPEC: 12 canonical frontmatter fields present; `### Out of Scope — <topic>` H3 sub-headings with `-` bullets satisfy `OutOfScopeRule`.
- [ ] plan-auditor verdict PASS at the Tier M threshold (0.80).
