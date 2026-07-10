# Plan — SPEC-HARNESS-VERIFY-PROMOTE-001

> Implementation plan for promoting the harness-generation offer + mandating a
> `harness-<name>-verify` companion skill + injecting two specialist-agent rule
> blocks + a 3-7 PLAN guardrail. Tier S (doc-only; markdown). Every touched file
> has a byte-identical template mirror; Template-First discipline governs every
> edit. Line-number citations are drift-prone — re-anchor by content token at
> run-phase.

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not
  assume a plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-HARNESS-VERIFY-PROMOTE-001/{spec,plan,acceptance,progress}.md`.
- **Epic position**: THIRD (Tier S, doc-only) SPEC of the 3-SPEC "Project-Harness
  Pipeline" Epic. `depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]` — the foundation
  SPEC that introduced the adaptive interview + `.moai/project/harness-spec.yaml`
  + the confirmed project type this SPEC's promoted offer consumes.
- **[HARD] Depends_on gate**: `SPEC-PROJECT-HARNESS-BRIDGE-001` is currently
  `status: draft`. The Phase 0.5 **Depends_on Pre-flight Check** (strict
  fulfillment = dependency `status: completed`) will BLOCK `/moai run` on this SPEC
  until BRIDGE is `completed`, unless the orchestrator surfaces the 3-option
  blocker and the user selects `override` (`--ignore-deps`, logged to
  `.moai/logs/depends-on-override.log` with rationale). Run-phase entry MUST NOT
  proceed on an unfulfilled dependency without that logged override.
- **Verified inventory** (research input — NOT to be re-investigated at plan
  authoring; re-verify by content token at run-phase):
  - Harness-generation offer today is buried at project workflow Phase 4.2 as ONE
    menu option among ~5 next-step choices (DB-sync / Create SPEC / Review /
    Generate harness / Done), AFTER all docs are generated. `project/meta-harness.md`
    handles the redirect to `harness-build-entry.md` → `harness-builder.md`.
  - `harness-builder.md` GENERATE emits 5 artifact types (thin command / Runner JS /
    specialist agents / companion skills / manifest.json). It does NOT currently
    mandate a verification / run skill.
  - Official Claude Code ships `/run-skill-generator`: discovers build / launch /
    test from a clean environment once, then commits the recipe to
    `.claude/skills/run-<name>/`. Precedent for a per-project verify skill.
  - Anthropic guidance (verified this session): ship a runnable verification loop
    ("give Claude a check it can run"); generate FEW trigger-rich agents (3-7 max) —
    over-generation degrades automatic sub-agent delegation.
  - claude.ai consumer system-prompt leak (reviewed this session): absorb ONLY (a)
    tool-priority decision tree ("category fit, not style preference") + (b)
    Skill-First execution rule ("read the relevant SKILL.md before file/code work").
    The rest is consumer-app-specific and out of scope.
  - Namespace: generated artifacts use `harness-*` prefix only; FROZEN guard
    rejects writes to `.claude/agents/moai/`, `.claude/skills/moai-*/`,
    `.claude/rules/moai/`. Template-First + neutrality preserved.

### Resolved clarifications (settled before Implementation Kickoff Approval)

Both plan-phase clarifications are RESOLVED (also recorded in progress.md §E.1). No
open markers remain.

- **Promoted-offer placement seam — RESOLVED.** The promoted offer lands in
  `project/meta-harness.md` as a **post-project-type-confirmation harness proposal**
  (meta-harness.md is the Phase 5.1 handoff module, NOT the interview's final
  question) + in `harness-build-entry.md` as that entry's final-round offer. Scope is
  NOT extended into `project/mode-detection.md` or `project/codebase-analysis.md`
  (those files are `SPEC-PROJECT-HARNESS-BRIDGE-001`'s scope). The Phase 4.2 next-step
  "Generate harness" menu — which lives in `project/doc-generation.md` (read-only for
  THIS SPEC) — is RETAINED as a fallback (both entry points reachable). The `<type>`
  token source: `harness-spec.yaml` `domain` field (`SPEC-PROJECT-HARNESS-BRIDGE-001`)
  when present, else the mode-detection confirmed project type.
- **Verify-skill enforcement surface — RESOLVED.** The `harness-<name>-verify` skill
  is ALWAYS mandatory for EVERY generated harness (mirroring `/run-skill-generator`'s
  "always ship a runnable check" theme) — it is NOT gated on any `harness-spec.yaml`
  field. When no build / launch / test recipe is discoverable, the verify skill is
  STILL emitted as a documented stub ("no recipe found") rather than omitted (edge E1).

## §B. Known Issues (filtered, Tier S — doc-only)

- **B-TF (Template-First)**: every touched file exists in BOTH trees (local
  `.claude/...` + `internal/template/templates/.claude/...`). Edit the template
  FIRST, mirror to local, then `make build`. A local-only edit that is not
  mirrored fails AC-HVP-010 (byte-parity diff).
- **B-NEUTRAL (template neutrality)**: the template tree MUST NOT gain internal
  SPEC IDs, internal dates, or commit SHAs. Do NOT paste
  `SPEC-HARNESS-VERIFY-PROMOTE-001` or this SPEC's dates into any
  `internal/template/templates/**` file. The neutrality CI guard
  (`template-neutrality-check.yaml` + `internal_content_leak_test.go`) triggers on
  the path change and must stay green (AC-HVP-011).
- **B-FROZEN (namespace guard)**: generated artifacts use `harness-*` only; the
  FROZEN guard rejecting `.claude/agents/moai/`, `.claude/skills/moai-*/`,
  `.claude/rules/moai/` MUST remain intact. When editing `harness-builder.md`, do
  NOT weaken or remove the guard language (AC-HVP-008).
- **B-NOSPEC (scope guard)**: the project + harness-generation flow must never
  write to `.moai/specs/**`. When editing `meta-harness.md` /
  `harness-build-entry.md`, do NOT introduce any `.moai/specs/` write path
  (AC-HVP-009).
- **B-MIRROR-PARITY (byte diff)**: after each edit, `diff` the local file against
  its template mirror; they must be byte-identical (AC-HVP-010). These files carry
  no internal-content tokens, so the mirror should be a clean byte copy.
- **B-DEPENDS (foundation gate)**: this SPEC's promoted offer references the
  confirmed project type + `harness-spec.yaml` from `SPEC-PROJECT-HARNESS-BRIDGE-001`.
  Run-phase entry is gated on BRIDGE being `status: completed` (§A Depends_on gate).
  Do NOT begin run-phase M1 without dependency fulfillment OR a logged override.
- **B-FEWAGENTS (over-generation)**: the 3-7 guardrail exists because
  over-generation degrades automatic sub-agent delegation. When editing the PLAN
  section, state the guardrail as a HARD cap with a justify-or-emit-skill rule; do
  NOT phrase it as an unbounded suggestion.
- **B-SHORT-BLOCKS (agent bloat)**: the two injected specialist rule blocks (§D.2
  of spec.md) MUST stay short (a few lines each). Do NOT expand them into long
  prose — over-long injected blocks bloat every generated agent.

## §C. Pre-flight Checklist (run before any change)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD

# 2. Depends_on fulfillment (foundation SPEC must be completed)
grep -n "^status:" .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md   # expect: status: completed (else Phase 0.5 blocks / requires --ignore-deps)

# 3. Locate the buried Phase 4.2 offer + the redirect
grep -n -i "Phase 4.2\|Generate harness\|harness-build-entry" .claude/skills/moai/workflows/project/meta-harness.md | head

# 4. Locate the harness-builder GENERATE artifact set + specialist-generation template + PLAN section
grep -n -i "GENERATE\|specialist\|manifest\|PLAN\|companion skill" .claude/skills/moai/workflows/harness-builder.md | head
grep -n -i "harness-\*\|FROZEN\|\.claude/agents/moai/\|\.claude/skills/moai-" .claude/skills/moai/workflows/harness-builder.md | head

# 5. Confirm both trees exist for every target file (Template-First)
for f in project/meta-harness.md harness-build-entry.md harness-builder.md; do
  ls ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f"
done

# 6. NO-SPEC scope baseline (must remain 0 in the two project-flow files after edit)
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/meta-harness.md \
  .claude/skills/moai/workflows/harness-build-entry.md || echo "no .moai/specs write path (good)"
```

## §D. Constraints (DO NOT VIOLATE)

**PRESERVE list (assert STILL EXISTS at exit)**:
- The Phase 4.2 next-step menu "Generate harness" option (RETAINED as fallback).
- The `harness-*` namespace-only invariant + FROZEN guard language in
  `harness-builder.md`.
- The `/moai project` NO-SPEC scope guard — no `.moai/specs/**` write path.
- The existing 5 generated artifact types (the verify skill is a 6th, additive).
- The `builder-harness` specialist-generation internals (except the two injected
  rule blocks + the PLAN guardrail statement).

**Forbidden**:
- Writing internal SPEC IDs / dates / SHAs into `internal/template/templates/**`
  (neutrality).
- Local-only edits that are not mirrored to the template tree (or vice versa).
- Any `.moai/specs/` write path in the project / harness-generation flow.
- Weakening or removing the FROZEN namespace guard.
- Removing the Phase 4.2 menu option (it is RETAINED, not replaced).
- Importing any of the claude.ai leak beyond the two absorbed patterns.
- Editing `project/mode-detection.md` or `project/codebase-analysis.md` (foundation
  SPEC scope). The promoted-offer placement seam is RESOLVED to stay in
  `meta-harness.md` + `harness-build-entry.md`; scope is NOT extended into those files.
- Expanding the two injected specialist rule blocks into long prose (agent bloat).

**Required**: Conventional Commits (`feat(SPEC-HARNESS-VERIFY-PROMOTE-001): M{N} …`
for run-phase; the plan-phase artifact commit uses the `feat(` prefix per the
plan-commit-subject lesson), `🗿 MoAI` trailer, specific-path `git add`.

## §E. Self-Verification Deliverables

Per the manager-develop prompt template §E (E1-E7), each milestone completion
report carries: E1 AC PASS/FAIL matrix (verbatim command output), E2 build result
(`make build` exit 0 — this is a doc-only SPEC, so the cross-platform Go build is a
non-regression check, not a feature check), E5 lint / neutrality (template-neutrality
guard + internal-content-leak test green), E6 commit SHAs + push state, E7 blocker
reports (never AskUserQuestion). E3 coverage and E4 subagent-boundary grep are N/A
(no Go code touched); state them as N/A rather than fabricating output.

## §F. Milestones

### M1 — Promote the harness-generation offer (REQ-HVP-001/002)

1. In `project/meta-harness.md`: surface the harness-generation proposal ("이
   프로젝트에 <type> 개발 하네스를 생성할까요?") as a post-project-type-confirmation
   harness proposal (meta-harness.md is the Phase 5.1 handoff module, NOT the
   interview) AFTER project-type confirmation. The Phase 4.2 next-step "Generate
   harness" menu option — hosted in `project/doc-generation.md` (read-only for THIS
   SPEC; retained by non-modification) — remains reachable as a fallback (both entry
   points reachable). Per the resolved placement seam: stay in `meta-harness.md`, do
   NOT edit `mode-detection.md` / `codebase-analysis.md`; `<type>` from
   `harness-spec.yaml` `domain`, else confirmed project type.
2. In `harness-build-entry.md`: surface the same proposal as the interview's
   final-round offer.
3. Mirror every edit to the template tree; `make build`.
4. Exit: AC-HVP-001/002/003 grep-green on both trees; byte-parity clean.

### M2 — harness-builder GENERATE: verify skill + specialist rule blocks + guardrail (REQ-HVP-003/004/005/006)

1. In `harness-builder.md` GENERATE: add the mandatory `harness-<name>-verify`
   companion skill as **artifact 6** (mirroring `/run-skill-generator`: discover +
   codify build / launch / test recipe from a clean environment). Per the resolved
   enforcement surface: ALWAYS mandatory for every generated harness (not gated on
   any `harness-spec.yaml` field); emit a documented stub ("no recipe found") when no
   recipe is discoverable (edge E1). (The optional MCP fragment is artifact 7, owned
   by `SPEC-HARNESS-MCP-PROVISION-001` — OUT OF SCOPE for THIS SPEC.)
2. In the specialist-agent generation template: inject the two short mandatory rule
   blocks (§D.2 of spec.md) — tool-priority decision tree + Skill-First execution —
   into every generated specialist agent body. Keep each block short.
3. In the PLAN section: state the 3-7-specialists-maximum guardrail + the
   justify-each-specialist-or-emit-skill rule.
4. Mirror to the template tree; `make build`.
5. Exit: AC-HVP-004/005/006/007/012 grep-green on both trees (AC-HVP-012 asserts the
   two injected specialist blocks stay bounded — ≤8 lines / verbatim §D.2 shape).

### M3 — Invariants + parity + neutrality + init smoke (REQ-HVP-007/008/009)

1. Confirm the `harness-*` namespace-only + FROZEN guard language intact
   (AC-HVP-008); confirm 0 `.moai/specs/` write paths in the two project-flow
   files (AC-HVP-009).
2. Full byte-parity sweep: `diff` every touched local file against its template
   mirror — all byte-identical (AC-HVP-010).
3. `make build`; run the template-neutrality guard + internal-content-leak test
   (`go test ./internal/template/...`) — must be green; no internal SPEC ID in the
   template tree (AC-HVP-011).
4. `moai init` into a sandbox; confirm the deployed tree carries the promoted offer
   + mandatory verify-skill clause (non-regression resurrection check).
5. Whole-repo non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code
   changed).
6. Exit: all 12 ACs PASS or documented PASS-WITH-DEBT.

## §G. Anti-Patterns (this SPEC)

- Removing the Phase 4.2 "Generate harness" menu option instead of RETAINING it as
  a fallback — REQ-HVP-001 requires both entry points reachable.
- Editing only `meta-harness.md` OR only `harness-build-entry.md` — the promoted
  offer must land in both (REQ-HVP-001 + REQ-HVP-002).
- Making the verify skill optional / omitting it when no recipe is discoverable —
  default is always-mandatory with a stub (edge E1).
- Expanding the two specialist rule blocks into long prose — bloats every generated
  agent (B-SHORT-BLOCKS).
- Weakening the FROZEN namespace guard while editing `harness-builder.md`.
- Writing `harness-spec` or any artifact under `.moai/specs/` — violates the
  NO-SPEC scope guard (REQ-HVP-008).
- Local-only edits without a template mirror (or vice versa) — fails byte-parity.
- Pasting `SPEC-HARNESS-VERIFY-PROMOTE-001` / this SPEC's dates into the template
  tree — fails neutrality.
- Beginning run-phase M1 while `SPEC-PROJECT-HARNESS-BRIDGE-001` is not `completed`
  without a logged `--ignore-deps` override — violates the Depends_on gate.

## §H. Cross-References

- `spec.md` §D — the 6-artifact generated-harness set + the two specialist rule
  blocks (SSOT shape).
- `acceptance.md` — AC matrix (SSOT), GWT scenarios, edge cases, quality gates, DoD.
- `SPEC-PROJECT-HARNESS-BRIDGE-001` — foundation SPEC (`depends_on`); confirmed
  project type + `harness-spec.yaml` consumed by the promoted offer.
- Official `/run-skill-generator` bundled skill — the runnable-verification pattern
  the mandatory `harness-<name>-verify` skill mirrors.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability
  — Tier S minimal delegation form applies (Section B may be filtered).
- CLAUDE.local.md §24 (Harness Namespace) + §25 (Template Internal-Content
  Isolation) + §15 (16-language neutrality) — the namespace / neutrality / mirror
  discipline this SPEC's edits must respect.
