# Plan — SPEC-PROJECT-HARNESS-BRIDGE-001

> Implementation plan for the project-interview adaptive clarity-scoring
> conversion + the `harness-spec.yaml` bridge into harness generation. Tier M
> (3 artifacts). Doc-only (markdown / yaml) — no Go code. Every touched file has
> a byte-identical template mirror; Template-First discipline governs every edit.
> Line-number citations are drift-prone — re-anchor by content token at run-phase.

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not
  assume a plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/{spec,plan,acceptance,progress}.md`.
- **Epic position**: FOUNDATION SPEC of the 3-SPEC "Project-Harness Pipeline"
  Epic. NO `depends_on`. The two downstream SPECs `depend_on` this one (they
  consume the `harness-spec.yaml` contract §D).
- **Verified inventory** (research input — NOT to be re-investigated at plan
  authoring; re-verify by content token at run-phase):
  - `/moai project` router: `.claude/skills/moai/workflows/project.md` +
    `project/{mode-detection,codebase-analysis,doc-generation,meta-harness}.md`,
    each with a byte-identical mirror under
    `internal/template/templates/.claude/skills/moai/workflows/...`.
  - Interview: Phase 0.3 (new project) + Phase 1.5 (existing project) in
    `project/mode-detection.md` are STATIC 3-round × 3-question interviews.
    `.moai/config/sections/interview.yaml` defines `clarity_threshold: 4` and
    `project.max_rounds: 3`, but the project flow does NOT consume
    `clarity_threshold` (always runs exactly 3 rounds).
  - The plan flow's `.claude/skills/moai/workflows/plan/clarity-interview.md`
    DOES use clarity scoring — the reference mechanism to mirror.
  - Interview output `.moai/project/interview.md` is currently NOT passed into
    harness generation. `project/meta-harness.md` Phase 5.1 composes the
    harness-creation request from `product.md` / `structure.md` / `tech.md`
    only, then routes to `harness-build-entry.md`, whose Phase 1 (Context-First
    Discovery) re-extracts domain / goal / constraints / scope from scratch.
  - Anthropic verified pattern: "Let Claude interview you" (AskUserQuestion
    interview → machine-readable spec artifact → execute).

### Open clarifications (resolve before Implementation Kickoff Approval)

- **[NEEDS CLARIFICATION: interview.yaml schema surface]** — Do the four new
  elicitation fields (verification / ui_surface / external_systems /
  team_sharing) require a new question-axis schema block ADDED to
  `interview.yaml` (both trees), or are they added purely as interview prose in
  `project/mode-detection.md`? This determines whether REQ-PHB-004 touches
  `interview.yaml` + its mirror (→ AC-PHB-013 applies) or only
  `mode-detection.md`. Default assumption if unresolved: add a minimal
  `additional_axes:` list to `interview.yaml` so the axes are config-declared
  (config-as-SSOT, consistent with the existing `clarity_threshold` /
  `max_rounds` living in `interview.yaml`).
- **[NEEDS CLARIFICATION: harness-spec.yaml re-run semantics]** — On a second
  `/moai project` invocation, does `doc-generation.md` overwrite
  `harness-spec.yaml`, merge with the prior file, or skip-if-present? Default
  assumption if unresolved: OVERWRITE (regenerate from the latest interview),
  matching the existing `interview.md` regeneration behavior. This is a
  low-blast-radius default; confirm only if the Epic's downstream SPECs need
  merge semantics.

## §B. Known Issues (filtered, Tier M — doc-only)

- **B-TF (Template-First)**: every touched file exists in BOTH trees (local
  `.claude/...` + `internal/template/templates/.claude/...`). Edit the template
  FIRST, mirror to local, then `make build`. A local-only edit that is not
  mirrored fails AC-PHB-011 (byte-parity diff).
- **B-NEUTRAL (template neutrality)**: the template tree MUST NOT gain internal
  SPEC IDs, internal dates, or commit SHAs. Do NOT paste `SPEC-PROJECT-HARNESS-BRIDGE-001`
  or this SPEC's dates into any `internal/template/templates/**` file. The
  neutrality CI guard (`template-neutrality-check.yaml` +
  `internal_content_leak_test.go`) triggers on the path change and must stay
  green (AC-PHB-012).
- **B-NOSPEC (scope guard)**: the project workflow must never write to
  `.moai/specs/**`. `harness-spec.yaml` lives under `.moai/project/`. When
  editing the interview / doc-generation sections, do NOT introduce any
  `.moai/specs/` write path (AC-PHB-010).
- **B-MIRROR-PARITY (byte diff)**: after each edit, `diff` the local file
  against its template mirror; they must be byte-identical (AC-PHB-011). The
  ONLY intentional divergence class is neutrality stripping — but these files
  carry no internal-content tokens, so the mirror should be a clean byte copy.
- **B-REFERENCE-READ**: the adaptive mechanism source `plan/clarity-interview.md`
  was NOT read at plan-authoring (token discipline). Run-phase MUST read it and
  mirror its clarity-scoring loop faithfully into the project interview rather
  than inventing a divergent scoring rubric.
- **B-DUAL-PHASE**: the interview lives in TWO phases (0.3 new-project + 1.5
  existing-project). Both must receive the adaptive loop AND the extended axes
  (REQ-PHB-003). A single-phase edit is an incomplete implementation.

## §C. Pre-flight Checklist (run before any change)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD

# 2. Locate the interview host + confirm the static-3-round shape
grep -n "max_rounds\|clarity_threshold" .moai/config/sections/interview.yaml
grep -rn -i "round\|clarity" .claude/skills/moai/workflows/project/mode-detection.md | head

# 3. Locate the reference adaptive mechanism (READ this at run-phase)
ls .claude/skills/moai/workflows/plan/clarity-interview.md

# 4. Confirm the compose / discard sites
grep -n -i "product.md\|structure.md\|tech.md\|interview.md" .claude/skills/moai/workflows/project/meta-harness.md | head
grep -n -i "Context-First\|Discovery\|domain\|goal" .claude/skills/moai/workflows/harness-build-entry.md | head

# 5. Confirm both trees exist for every target file (Template-First)
for f in project/mode-detection.md project/doc-generation.md project/meta-harness.md harness-build-entry.md; do
  ls ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f"
done
ls .moai/config/sections/interview.yaml internal/template/templates/.moai/config/sections/interview.yaml

# 6. NO-SPEC scope baseline (must remain 0 in touched sections after edit)
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/ || echo "no .moai/specs write path (good)"
```

## §D. Constraints (DO NOT VIOLATE)

**PRESERVE list (assert STILL EXISTS at exit)**:
- The `/moai project` NO-SPEC scope guard — no `.moai/specs/**` write path.
- The existing `.moai/project/interview.md` human-readable output (harness-spec.yaml
  is additive).
- The builder-harness specialist-generation internals (unchanged).
- `interview.yaml` `clarity_threshold: 4` + `project.max_rounds: 3` values
  (consumed, not changed).

**Forbidden**:
- Writing internal SPEC IDs / dates / SHAs into `internal/template/templates/**`
  (neutrality).
- Local-only edits that are not mirrored to the template tree (or vice versa).
- Any `.moai/specs/` write path in the project workflow.
- Re-asking a field in `harness-build-entry.md` Phase 1 that is already answered
  in `harness-spec.yaml` (REQ-PHB-009).
- Editing only one of the two interview phases (0.3 / 1.5).
- Inventing a clarity-scoring rubric that diverges from
  `plan/clarity-interview.md`.

**Required**: Conventional Commits (`feat(SPEC-PROJECT-HARNESS-BRIDGE-001): M{N} …`
for run-phase; the plan-phase artifact commit uses the `feat(` prefix per the
plan-commit-subject lesson), `🗿 MoAI` trailer, specific-path `git add`.

## §E. Self-Verification Deliverables

Per the manager-develop prompt template §E (E1-E7), each milestone completion
report carries: E1 AC PASS/FAIL matrix (verbatim command output), E2 build
result (`make build` exit 0 — this is a doc-only SPEC, so the cross-platform Go
build is a non-regression check, not a feature check), E5 lint / neutrality
(template-neutrality guard + internal-content-leak test green), E6 commit SHAs +
push state, E7 blocker reports (never AskUserQuestion). E3 coverage and E4
subagent-boundary grep are N/A (no Go code touched); state them as N/A rather
than fabricating output.

## §F. Milestones

### M1 — Adaptive clarity-scored interview + extended axes (REQ-PHB-001/002/003/004)

1. READ `plan/clarity-interview.md` to extract the exact clarity-scoring loop
   (scoring rubric, early-exit condition, additional-round trigger).
2. In `project/mode-detection.md` Phase 0.3 (new-project) AND Phase 1.5
   (existing-project), replace the static 3-round loop with the adaptive loop:
   consume `interview.yaml` `clarity_threshold` (4); early-exit when clarity ≥
   threshold; run additional rounds up to `project.max_rounds` (3).
3. Extend the interview question set (both phases) to elicit the four new fields:
   verification method, UI surface, external systems, team-sharing intent.
4. Resolve `[NEEDS CLARIFICATION: interview.yaml schema surface]` — if the axes
   are config-declared, add the minimal axis block to `interview.yaml`.
5. Mirror every edit to the template tree; `make build`.
6. Exit: AC-PHB-001..004 (+ AC-PHB-013 if interview.yaml touched) grep-green on
   both trees; byte-parity diff clean.

### M2 — harness-spec.yaml write (REQ-PHB-005/006)

1. In `project/doc-generation.md`, add the `harness-spec.yaml` write step:
   emit `.moai/project/harness-spec.yaml` alongside `interview.md`, populated
   from the interview answers, using the §D schema (8 fields: domain, goal,
   constraints, scope, verification, external_systems, ui_surface, team_sharing).
2. Confirm the write target path is `.moai/project/` (NOT `.moai/specs/`).
3. Mirror to the template tree; `make build`.
4. Exit: AC-PHB-005/006/007-write grep-green (schema fields present in
   doc-generation.md, write path `.moai/project/harness-spec.yaml`); AC-PHB-010
   NO-SPEC guard clean.

### M3 — Harness generation consumption (REQ-PHB-007/008/009)

1. In `project/meta-harness.md` Phase 5.1, change the compose step to draw from
   `harness-spec.yaml` PLUS `product.md` / `structure.md` / `tech.md` (no longer
   product/structure/tech only).
2. In `harness-build-entry.md` Phase 1 (Context-First Discovery), consume
   `harness-spec.yaml` to pre-satisfy domain / goal / constraints / scope; add
   explicit prose that Discovery re-asks ONLY absent / ambiguous fields and does
   NOT re-ask an already-answered field.
3. Mirror both to the template tree; `make build`.
4. Exit: AC-PHB-008 (meta-harness references harness-spec.yaml), AC-PHB-009
   (harness-build-entry consumes + pre-satisfy / no-duplicate language)
   grep-green on both trees.

### M4 — Template mirror parity + neutrality + init verification (REQ-PHB-011/012)

1. Full byte-parity sweep: `diff` every touched local file against its template
   mirror — all must be byte-identical (AC-PHB-011).
2. `make build`; run the template-neutrality guard + internal-content-leak test
   (`go test ./internal/template/...`) — must be green (AC-PHB-012).
3. `moai init` into a sandbox; confirm the deployed tree carries the adaptive
   interview, the `harness-spec.yaml` write path, and the extended axes
   (AC-PHB-014).
4. Whole-repo non-regression: `go build ./...` + `go test ./...` exit 0 (no Go
   code changed, so this is a non-regression check).
5. Exit: all 14 ACs PASS or documented PASS-WITH-DEBT.

## §G. Anti-Patterns (this SPEC)

- Editing only Phase 0.3 (or only Phase 1.5) — the adaptive loop + extended axes
  must land in BOTH interview phases (REQ-PHB-003).
- Inventing a new clarity-scoring rubric instead of mirroring
  `plan/clarity-interview.md` — creates two divergent scoring mechanisms.
- Writing `harness-spec.yaml` under `.moai/specs/` — violates the NO-SPEC scope
  guard (REQ-PHB-010).
- Local-only edits without a template mirror (or vice versa) — fails byte-parity.
- Pasting `SPEC-PROJECT-HARNESS-BRIDGE-001` / this SPEC's dates into the template
  tree — fails neutrality.
- Making `harness-build-entry.md` Phase 1 STILL re-ask domain/goal after
  consuming harness-spec.yaml — defeats the whole point (REQ-PHB-009).

## §H. Cross-References

- `spec.md` §D — the canonical `harness-spec.yaml` 8-field schema.
- `acceptance.md` — AC matrix (SSOT), GWT scenarios, quality gates, DoD.
- `.claude/skills/moai/workflows/plan/clarity-interview.md` — reference adaptive
  clarity-scoring mechanism (READ at run-phase M1).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M
  Section A-E delegation template applies to every milestone delegation.
- CLAUDE.local.md §2 (Template-First) + §15 (16-language neutrality) + §25
  (Template Internal-Content Isolation) — the neutrality / mirror discipline
  this SPEC's edits must respect.
