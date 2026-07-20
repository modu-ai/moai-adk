# Plan — SPEC-PROJECT-HARNESS-BRIDGE-001

> Implementation plan for the project-interview adaptive clarity-scoring
> conversion + the `harness-spec.yaml` bridge into harness generation. Tier M
> (3 artifacts). Doc-only (markdown / yaml) — no Go code. Every touched file has
> a byte-identical template mirror; Template-First discipline governs every edit.
> Line-number citations are drift-prone — re-anchor by content token at run-phase.
>
> **v0.2.0 amendment in force.** M1-M4 are CLOSED (delivered at the v0.1.1 close,
> sync `0c9871b46`). The active work is **M5** (§F) — the two-stage interview that
> makes the extended-axes Round 4 REACHABLE. Read spec.md §C.7 + `## Amendments`
> before starting M5.

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not
  assume a plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/{spec,plan,acceptance,progress}.md`.
- **Epic position**: FOUNDATION SPEC of the 3-SPEC "Project-Harness Pipeline"
  Epic. NO `depends_on`. The two downstream SPECs `depend_on` this one (they
  consume the `harness-spec.yaml` contract §D).

### Amendment v0.2.0 state (READ FIRST)

- **Prior close**: run `9f1f0f53b`, sync `0c9871b46` (`prior_completed_sha:
  0c9871b46b6719325427dc0126e4eb65d7b0f2d8`), version `0.1.1`, `status: completed`.
- **Now**: `status: in-progress`, version `0.2.0`, `amendment_of:` self-referential
  (in-place amendment per `.claude/rules/moai/development/spec-frontmatter-schema.md`
  § Status Transition Ownership Matrix, `completed → in-progress (amendment)` row).
- **AC state**:
  - **AC-PHB-001 … AC-PHB-014 — PASS** as of the v0.1.1 close (`0c9871b46`).
    These are NOT re-litigated by M5; they must remain PASS (regression guard).
  - **AC-PHB-015 … AC-PHB-020 — OUTSTANDING**. These are the amendment scope
    (the REACHABILITY ACs). They are the M5 exit criteria.
- **Why the amendment exists**: a sync-auditor MUST-FIX found the extended-axes
  Round 4 UNREACHABLE on BOTH interview paths — `project.max_rounds: 3` caps the
  interview at 3 rounds AND the clarity early-exit "skips the remaining rounds",
  so the four axes were never collected, `harness-spec.yaml` wrote them
  always-empty, and `harness-build-entry.md` re-asked everything. The SPEC's
  headline feature was inert while all 14 ACs passed (AC-PHB-004 greps for axis
  TOKEN PRESENCE, never REACHABILITY).
- **Amendment file-pair set — 7 pairs** (the v0.1.x set of 6, plus `project.md`):
  1. `project/mode-detection.md` + template mirror
  2. `project/codebase-analysis.md` + template mirror
  3. `project/doc-generation.md` + template mirror
  4. `project/meta-harness.md` + template mirror (untouched by M5 — v0.1.x wiring stands)
  5. `harness-build-entry.md` + template mirror (untouched by M5 — v0.1.x wiring stands)
  6. `.moai/config/sections/interview.yaml` + template mirror
  7. **`project.md` + template mirror — NEW 7th pair (scope delta)**: the `/moai project`
     router's routing table carries two stale `3-round` claims ("3-round
     Vision/Technology/Scope interview", "3-round Ownership/Constraints/Priority
     interview"). Combined with the `codebase-analysis.md` frontmatter claim, the
     baseline is **6** `3-round` matches across both trees (3 per tree × 2 trees).
     AC-PHB-020 requires 0. This pair was NOT in the v0.1.x touch set.
- **Verified inventory** (research input — NOT to be re-investigated at plan
  authoring; re-verify by content token at run-phase):
  - `/moai project` router: `.claude/skills/moai/workflows/project.md` +
    `project/{mode-detection,codebase-analysis,doc-generation,meta-harness}.md`,
    each with a byte-identical mirror under
    `internal/template/templates/.claude/skills/moai/workflows/...`.
  - Interview: Phase 0.3 (new project) in `project/mode-detection.md` +
    Phase 1.5 (existing project) in `project/codebase-analysis.md` are STATIC
    3-round × 3-question interviews. `.moai/config/sections/interview.yaml`
    defines `clarity_threshold: 4` (the interview ENTRY floor) and
    `project.max_rounds: 3`, but the project flow does NOT consume
    `clarity_threshold` (always runs exactly 3 rounds, never adapting to answer
    clarity).
  - The plan flow's `.claude/skills/moai/workflows/plan/clarity-interview.md`
    DOES use clarity scoring on a 0-10 scale (sufficiency exit ≥ 8, abandon
    ≤ 3, entry floor 4) — the reference mechanism to mirror. Note:
    `clarity_threshold` (4) is the ENTRY floor, NOT the early-exit target (the
    exit target is the sufficiency bar, clarity ≥ 8).
  - Interview output `.moai/project/interview.md` is currently NOT passed into
    harness generation. `project/meta-harness.md` Phase 5.1 composes the
    harness-creation request from `product.md` / `structure.md` / `tech.md`
    only, then routes to `harness-build-entry.md`, whose Phase 1 (Context-First
    Discovery) re-extracts domain / goal / constraints / scope from scratch.
  - Anthropic verified pattern: "Let Claude interview you" (AskUserQuestion
    interview → machine-readable spec artifact → execute).

### Resolved decisions (D2 — clarifications settled at plan-phase fix pass)

Both plan-phase clarifications are RESOLVED (decision log recorded in
`progress.md` §E.2); no open clarification markers remain.

- **interview.yaml schema surface → CONFIG-DECLARED.** The four new elicitation
  fields (verification / ui_surface / external_systems / team_sharing) are
  declared in an `additional_axes:` block in
  `.moai/config/sections/interview.yaml` (both trees), NOT prose-only.
  Consequence: REQ-PHB-004 touches `interview.yaml` + its mirror, so AC-PHB-013
  APPLIES (unconditional — no longer N/A). Rationale: config-as-SSOT, consistent
  with the existing `clarity_threshold` / `max_rounds` living in `interview.yaml`.
- **harness-spec.yaml re-run semantics → OVERWRITE.** On a second `/moai project`
  invocation, `doc-generation.md` OVERWRITES `harness-spec.yaml` (regenerates
  from the latest interview answers), matching the existing `interview.md`
  regeneration behavior. Downstream Epic SPECs consume the fresh state — no merge
  / skip-if-present semantics.

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
- **B-INTERVIEW-YAML-DRIFT (pre-existing)**: `.moai/config/sections/interview.yaml`
  is ALREADY byte-divergent between the local tree (4-space indent + different
  key order) and the template tree. Do NOT incremental-edit both copies to
  convergence — instead edit the TEMPLATE `interview.yaml` (add the
  `additional_axes:` block), then OVERWRITE the local `interview.yaml` WHOLESALE
  from the edited template (copy template→local). This is the correct
  Template-First re-mirror and the only way AC-PHB-013 byte-parity passes despite
  the pre-existing drift (interview.yaml is template-managed config, not
  runtime-managed `settings.local.json`).
- **B-REFERENCE-READ**: the adaptive mechanism source `plan/clarity-interview.md`
  was NOT read in full at plan-authoring (token discipline; only its scoring
  header L14-45 was confirmed at the fix pass). Run-phase MUST read it and mirror
  its clarity-scoring loop faithfully (0-10 scale, sufficiency exit ≥ 8, abandon
  ≤ 3, entry floor 4) into BOTH project interviews rather than inventing a
  divergent scoring rubric.
- **B-DUAL-PHASE**: the interview lives in TWO phases across TWO files — Phase
  0.3 (new-project) in `project/mode-detection.md` and Phase 1.5
  (existing-project) in `project/codebase-analysis.md`. Both must receive the
  adaptive loop AND the extended axes (REQ-PHB-003). Editing only one file (only
  Phase 0.3 or only Phase 1.5) is an incomplete implementation.
- **B-STAGE-EXEMPTION (v0.2.0 — the amendment's central hazard)**: **a Round 4
  heading that merely EXISTS, without the explicit cap / early-exit exemption
  language, is the v0.1.x defect restated — the exemption prose IS the
  reachability contract.** Both hosts ALREADY have a `**Round 4: Verification,
  Surfaces, and Sharing (extended axes)**` heading today, and the feature is still
  inert, because nothing in either host says the round survives
  `project.max_rounds: 3` or the early-exit "skip the remaining rounds" clause.
  M5 is NOT "add a Round 4" (it is already there) — M5 is "state, in prose, that
  Round 4 is Stage B: it ALWAYS runs, it is EXEMPT from `max_rounds`, EXEMPT from
  the Stage A early-exit skip, EXEMPT from the abandon path, and EXEMPT from
  clarity scoring." AC-PHB-015 greps for exactly that exemption language, anchored
  to the Round 4 / Stage B section (REQ-PHB-014). An M5 that restructures the
  rounds but omits the exemption prose passes no AC and fixes nothing.
- **B-ASYMMETRIC-ROUNDS (v0.2.0)**: the two hosts' Stage A round topics are
  asymmetric and NEITHER covers all four REQUIRED base fields —
  `mode-detection.md` runs Vision / Technology / Scope (no constraints round, no
  domain round); `codebase-analysis.md` runs Ownership-Purpose / Constraints /
  Documentation-Priority (no scope round, no domain round). `doc-generation.md`'s
  field mapping assumes a uniform interview shape that does not exist. M5.3 must
  realign the round topics AND add an explicit base-field coverage mapping to each
  host (REQ-PHB-018 / AC-PHB-018) — do NOT just sprinkle the field names into
  existing prose.
- **B-STAGE-A-EARLY-EXIT (v0.2.0)**: the early-exit prose in both hosts currently
  says clarity ≥ 8 "terminate the interview early (skip the remaining rounds)".
  Two edits are required, not one: (a) gate the exit on all four REQUIRED base
  fields being answered (REQ-PHB-002 amended / REQ-PHB-017), and (b) rescope the
  exit to **Stage A** — it must not read as skipping "the interview" (which would
  swallow Stage B). AC-PHB-016 greps for BOTH.

## §C. Pre-flight Checklist (run before any change)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD

# 2. Locate BOTH interview hosts + confirm the static-3-round shape
#    (Phase 0.3 -> mode-detection.md; Phase 1.5 -> codebase-analysis.md)
grep -n "max_rounds\|clarity_threshold" .moai/config/sections/interview.yaml
grep -rn -i "round\|clarity" .claude/skills/moai/workflows/project/mode-detection.md | head
grep -rn -i "round\|clarity\|Phase 1.5" .claude/skills/moai/workflows/project/codebase-analysis.md | head

# 3. Locate the reference adaptive mechanism (READ this at run-phase)
ls .claude/skills/moai/workflows/plan/clarity-interview.md

# 4. Confirm the compose / discard sites
grep -n -i "product.md\|structure.md\|tech.md\|interview.md" .claude/skills/moai/workflows/project/meta-harness.md | head
grep -n -i "Context-First\|Discovery\|domain\|goal" .claude/skills/moai/workflows/harness-build-entry.md | head

# 5. Confirm both trees exist for every target file (Template-First)
for f in project/mode-detection.md project/codebase-analysis.md project/doc-generation.md project/meta-harness.md harness-build-entry.md; do
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
   (0-10 scale, sufficiency-exit at clarity ≥ 8, abandon at ≤ 3, entry floor 4;
   additional-round trigger).
2. Replace the static 3-round loop with the adaptive loop in BOTH interview
   hosts — Phase 0.3 (new-project) in `project/mode-detection.md` AND Phase 1.5
   (existing-project) in `project/codebase-analysis.md`: consume `interview.yaml`
   `clarity_threshold` (4) as the ENTRY floor; early-exit when clarity reaches the
   sufficiency target (≥ 8); abandon on a drop to ≤ 3; run additional rounds up
   to `project.max_rounds` (3). Use the SAME 0-10 scale semantics as
   `plan/clarity-interview.md` — do NOT treat `clarity_threshold` (4) as the exit
   target.
3. Extend the interview question set (both files) to elicit the four new fields:
   verification method, UI surface, external systems, team-sharing intent.
4. Add the config-declared `additional_axes:` block to `interview.yaml` (D2
   resolution): edit the TEMPLATE `interview.yaml`, then OVERWRITE the local
   `interview.yaml` WHOLESALE from the edited template (per B-INTERVIEW-YAML-DRIFT
   — the two trees are pre-existing byte-divergent; incremental-editing both
   would not converge).
5. Mirror every other edit (mode-detection.md, codebase-analysis.md) to the
   template tree; `make build`.
6. Exit: AC-PHB-001..004 + AC-PHB-013 grep-green on both trees; byte-parity diff
   clean (including interview.yaml + codebase-analysis.md).

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
5. Exit: **AC-PHB-001 … AC-PHB-014 PASS** (achieved — verified at the v0.1.1
   close, sync `0c9871b46`). **AC-PHB-015 … AC-PHB-020 are NOT in M4's scope** —
   they did not exist at the v0.1.1 close and are OUTSTANDING; they are the M5
   exit criteria (§A Amendment v0.2.0 state).

> **M1-M4 are CLOSED.** They were delivered at the v0.1.1 close (run `9f1f0f53b`,
> sync `0c9871b46`). They are retained above verbatim as the historical record and
> as the regression baseline: M5 MUST NOT break AC-PHB-001 … AC-PHB-014. Do NOT
> re-execute M1-M4.

### M5 — Two-stage interview: make Round 4 REACHABLE (v0.2.0 amendment; REQ-PHB-002-amended / 013-019)

**Goal**: the extended-axes Round 4 currently exists in prose but can never run
(`project.max_rounds: 3` caps the interview; the clarity early-exit skips the
remaining rounds). M5 restructures both interview hosts into **Stage A** (Rounds
1-3, clarity-scored, capped) + **Stage B** (Round 4, mandatory, exempt) so the four
axes are actually collected. Read spec.md §C.7 (the Stage A/B table + design
rationale) before starting. Read B-STAGE-EXEMPTION / B-ASYMMETRIC-ROUNDS /
B-STAGE-A-EARLY-EXIT (§B) — they name the three ways this milestone silently fails.

**M5.1 — Stage A / Stage B restructure in BOTH interview hosts** (REQ-PHB-013 /
REQ-PHB-014; AC-PHB-015)

1. In BOTH `project/mode-detection.md` (Phase 0.3) and
   `project/codebase-analysis.md` (Phase 1.5), name the two stages explicitly:
   **Stage A** = the clarity-scored adaptive discovery rounds (Rounds 1-3, governed
   by `clarity_threshold` entry floor 4 / sufficiency exit ≥ 8 / abandon ≤ 3 /
   `project.max_rounds` cap 3); **Stage B** = the mandatory extended-axes round
   (Round 4).
2. State the **Stage B exemption contract** explicitly in the Round 4 / Stage B
   section of BOTH hosts: Round 4 **ALWAYS runs**, **unconditionally**, AFTER
   Stage A terminates — by early exit, by abandon, OR by hitting the cap — and is
   **EXEMPT from `project.max_rounds`**, **EXEMPT from the Stage A early-exit
   skip**, **EXEMPT from the abandon path**, and **EXEMPT from clarity scoring**.
   Rationale to carry in the prose: the four axes are clarity-independent factual
   collection, not ambiguity resolution — there is nothing to "score" about whether
   the project has a UI or which test command it runs.
3. [HARD] Per B-STAGE-EXEMPTION: the Round 4 heading ALREADY exists in both hosts.
   Adding/keeping the heading is NOT the deliverable — the **exemption prose IS the
   reachability contract**. A Round 4 without the exemption language reproduces the
   v0.1.x defect exactly.
4. Exit: AC-PHB-015 green on both hosts (exemption language present AND anchored to
   the Round 4 / Stage B section; `Stage A` + `Stage B` both named).

**M5.2 — Stage A early-exit gated on REQUIRED base fields** (REQ-PHB-002 amended /
REQ-PHB-017; AC-PHB-016)

1. In BOTH hosts, amend the early-exit clause: it fires ONLY when the accumulated
   clarity score reaches the sufficiency target (≥ 8) **AND all four REQUIRED base
   fields (`domain`, `goal`, `constraints`, `scope`) have been answered**. While
   any required base field is unanswered, Stage A continues (up to
   `project.max_rounds`) regardless of clarity score.
2. Rescope the early-exit prose to **Stage A only** — it terminates the remaining
   **Stage A** rounds, NOT "the interview". The current wording ("terminate the
   interview early (skip the remaining rounds)") would swallow Stage B and must be
   corrected.
3. Cap behavior is unchanged and remains a hard stop: on reaching `max_rounds` with
   a required field still unanswered, record the field as absent (acceptance.md §D
   E6) and proceed to Stage B. Do NOT loop.
4. Exit: AC-PHB-016 green on both hosts (required-base-fields precondition present
   in the early-exit prose AND the exit scoped to Stage A).

**M5.3 — Base-field coverage mapping in each host** (REQ-PHB-018; AC-PHB-018)

1. Realign the Stage A round topics so each host's Rounds 1-3 elicit ALL FOUR
   required base fields. Current gaps (B-ASYMMETRIC-ROUNDS):
   - `mode-detection.md`: Vision / Technology / Scope → **no constraints round, no
     domain round**.
   - `codebase-analysis.md`: Ownership-Purpose / Constraints / Documentation-Priority
     → **no scope round, no domain round**.
2. Add an explicit **base-field coverage mapping** to each host — a short table or
   list naming, per required base field, which Stage A round elicits it. This is
   the deliverable AC-PHB-018 greps for; scattering the field names through
   existing prose does NOT satisfy it.
3. For `codebase-analysis.md` (existing projects): a base field confidently
   **auto-populated from the Phase 1 codebase analysis** (e.g. `domain` inferred
   from the detected stack) COUNTS as answered, MUST NOT be re-asked, and MUST be
   marked as auto-populated in the coverage mapping. Consistent with acceptance.md
   §D E3 / E8 — auto-population satisfies the REQ-PHB-017 gate, it does not bypass
   it (an un-inferable required field still keeps Stage A running).
4. Exit: AC-PHB-018 green on both hosts (coverage mapping present; all four
   required base fields named within it).

**M5.4 — `interview.yaml`: `max_rounds` is Stage-A-scoped** (REQ-PHB-015;
AC-PHB-017)

1. Document `project.max_rounds: 3` in-place as capping **ONLY the Stage A
   clarity-scored discovery rounds** — NOT the interview as a whole. Call out the
   mandatory extended-axes round (Stage B / Round 4) as EXEMPT from the cap.
2. Goal: a reader of `interview.yaml` ALONE must not be able to infer that the
   project interview terminates after 3 rounds.
3. Per B-INTERVIEW-YAML-DRIFT: edit the TEMPLATE `interview.yaml` first, then
   OVERWRITE the local copy WHOLESALE from the edited template (do NOT
   incremental-edit both trees).
4. Exit: AC-PHB-017 green (Stage-A scoping documented + exemption called out +
   byte-parity diff clean).

**M5.5 — REQUIRED vs EXTENDED partition in `doc-generation.md`** (REQ-PHB-016;
AC-PHB-019)

1. Annotate the 8-field `harness-spec.yaml` mapping in `project/doc-generation.md`
   with the two field classes: **REQUIRED base** (`domain`, `goal`, `constraints`,
   `scope` — collected in Stage A, gate the Stage A early exit) vs **EXTENDED**
   (`verification`, `ui_surface`, `external_systems`, `team_sharing` — collected in
   the mandatory Stage B round).
2. The write site must know which fields come from Stage A vs Stage B; the schema
   SSOT is spec.md §D (already annotated by the amendment).
3. Exit: AC-PHB-019 green (partition stated in `doc-generation.md`; the four
   EXTENDED fields named).

**M5.6 — F-4 sweep: remove all 6 `3-round` claims** (REQ-PHB-019; AC-PHB-020)

1. Baseline is **6** matches across both trees (verify at run-phase pre-flight):
   - `project/codebase-analysis.md` frontmatter `description:` — "3-round deep
     interview for existing projects" (×1 per tree)
   - `project.md` routing table — "3-round Vision/Technology/Scope interview" and
     "3-round Ownership/Constraints/Priority interview" (×2 per tree)
   - × 2 trees (local `.claude/` + `internal/template/templates/.claude/`) = 6
2. Reword each to describe the interview honestly: variable-length Stage A
   (clarity-scored, up to `max_rounds`) + a mandatory Stage B extended-axes round.
   Do NOT simply delete the descriptions.
3. `project.md` is the **7th file-pair** (§A) — it was not in the v0.1.x touch set.
   Mirror it like every other pair.
4. Exit: AC-PHB-020 green — `grep -rn "3-round" .claude/skills/moai/workflows/
   internal/template/templates/.claude/skills/moai/workflows/ | wc -l` → **0**.

**M5.7 — Mirror, build, verify all 20 ACs** (REQ-PHB-011/012; AC-PHB-011/012/014 +
full matrix)

1. Template-First mirror sweep across all **7** file-pairs: `diff -q` each local
   file against its template mirror — all byte-identical (AC-PHB-011). Remember
   `project.md` (the new 7th pair) and `interview.yaml` (wholesale
   template→local overwrite per B-INTERVIEW-YAML-DRIFT).
2. `make build`; template-neutrality guard + internal-content-leak test
   (`go test ./internal/template/...`) green (AC-PHB-012). Do NOT paste
   `SPEC-PROJECT-HARNESS-BRIDGE-001` or any internal date/SHA into the template
   tree.
3. `moai init` into a sandbox; confirm the deployed tree carries the two-stage
   interview + the `harness-spec.yaml` write path + the extended axes (AC-PHB-014).
4. Whole-repo non-regression: `go build ./...` + `go test ./...` exit 0.
5. Run the **full 20-AC matrix** (acceptance.md §C) with verbatim command output
   recorded in progress.md §E.2:
   - **AC-PHB-001 … AC-PHB-014 — regression guard**: these were PASS at the v0.1.1
     close; they MUST remain PASS. M5 must not weaken them.
   - **AC-PHB-015 … AC-PHB-020 — the amendment's exit criteria**: these are the
     REACHABILITY ACs and are OUTSTANDING until M5 lands.
6. Exit: all 20 ACs PASS (or documented PASS-WITH-DEBT with rationale).

## §G. Anti-Patterns (this SPEC)

- Editing only Phase 0.3 (`mode-detection.md`) or only Phase 1.5
  (`codebase-analysis.md`) — the adaptive loop + extended axes must land in BOTH
  interview files (REQ-PHB-003).
- Inventing a new clarity-scoring rubric instead of mirroring
  `plan/clarity-interview.md` — creates two divergent scoring mechanisms.
- Writing `harness-spec.yaml` under `.moai/specs/` — violates the NO-SPEC scope
  guard (REQ-PHB-010).
- Local-only edits without a template mirror (or vice versa) — fails byte-parity.
- Pasting `SPEC-PROJECT-HARNESS-BRIDGE-001` / this SPEC's dates into the template
  tree — fails neutrality.
- Making `harness-build-entry.md` Phase 1 STILL re-ask domain/goal after
  consuming harness-spec.yaml — defeats the whole point (REQ-PHB-009).

### v0.2.0 amendment anti-patterns (M5)

- **Leaving a Round 4 heading in place without the exemption prose** — the v0.1.x
  defect restated. The heading already exists today and the feature is still inert;
  the exemption language (ALWAYS runs / exempt from `max_rounds` / exempt from the
  early-exit skip) IS the reachability contract (B-STAGE-EXEMPTION, REQ-PHB-014).
- **Raising `project.max_rounds` from 3 to 4 as the "fix"** — this conflates the two
  stages and re-subjects the clarity-independent axes to the clarity loop. The
  extended-axes round is EXEMPT from the cap, not accommodated by a bigger cap
  (spec.md §C.7 design rationale). `max_rounds: 3` stays 3 (§D PRESERVE list).
- **Deleting the early-exit to "guarantee" Round 4 runs** — the early exit is a
  delivered v0.1.x feature (REQ-PHB-002, AC-PHB-002 PASS). It is AMENDED (gated on
  required base fields + rescoped to Stage A), never removed.
- **Scattering `domain` / `goal` / `constraints` / `scope` tokens through host prose
  instead of adding the explicit base-field coverage mapping** — AC-PHB-018 greps
  for the mapping, and token-sprinkling is precisely the token-presence-not-
  reachability failure mode the amendment exists to correct.
- **Re-asking a base field in `codebase-analysis.md` that Phase 1 codebase analysis
  already auto-populated** — auto-populated counts as answered (REQ-PHB-018,
  acceptance.md §D E3 / E8).
- **Fixing only the `codebase-analysis.md` frontmatter `3-round` claim and missing
  the `project.md` routing table** — the baseline is 6 matches across both trees,
  not 1. AC-PHB-020 requires 0 (M5.6).
- **Forgetting `project.md` in the mirror sweep** — it is the 7th file-pair, new in
  this amendment; it was not in the v0.1.x touch set (§A).

## §H. Cross-References

- `spec.md` §D — the canonical `harness-spec.yaml` 8-field schema (REQUIRED vs
  EXTENDED partition per the v0.2.0 amendment).
- `spec.md` §C.7 — the two-stage interview structure (Stage A / Stage B table +
  design rationale + REQ-PHB-013 … REQ-PHB-019). **Read before M5.**
- `spec.md` `## Amendments` — the v0.2.0 amendment record (prior_completed_sha,
  rationale, 7-pair scope table).
- `acceptance.md` — AC matrix (SSOT), GWT scenarios, quality gates, DoD.
- `acceptance.md` §C AC-PHB-015 … AC-PHB-020 — the REACHABILITY ACs (M5 exit
  criteria) + the REACHABILITY convention note ("token presence is not
  reachability").
- `.claude/skills/moai/workflows/plan/clarity-interview.md` — reference adaptive
  clarity-scoring mechanism (READ at run-phase M1).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M
  Section A-E delegation template applies to every milestone delegation.
- CLAUDE.local.md §2 (Template-First) + §15 (16-language neutrality) + §25
  (Template Internal-Content Isolation) — the neutrality / mirror discipline
  this SPEC's edits must respect.
