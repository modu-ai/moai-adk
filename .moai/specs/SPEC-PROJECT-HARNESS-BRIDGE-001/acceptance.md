# Acceptance Criteria — SPEC-PROJECT-HARNESS-BRIDGE-001

> SSOT for the AC matrix. **20 ACs covering all 19 REQs** (100% AC→REQ coverage).
> AC-PHB-001 … AC-PHB-014 are the original v0.1.x matrix (unchanged);
> AC-PHB-015 … AC-PHB-020 are the v0.2.0 amendment's **REACHABILITY** ACs.
>
> Verification conventions: (1) every grep is anchored to a content token, not a
> line number; (2) preservation / NO-WRITE ACs assert absence explicitly;
> (3) template↔local parity is a byte `diff` check on each touched file; (4) this
> is a doc-only SPEC, so ACs are grep / diff / `make build` / `moai init` based
> (no Go test / coverage ACs).
>
> **REACHABILITY convention (v0.2.0 — the lesson of the missed MUST-FIX).** The
> original matrix verified that the four extended axes were *mentioned* in the
> interview hosts (AC-PHB-004 greps for axis tokens) but never that the round
> carrying them could actually *run*. It could not: `project.max_rounds: 3` capped
> the interview at 3 rounds and the clarity early-exit "skipped the remaining
> rounds", so Round 4 was dead on both paths and the headline feature was inert
> while every AC passed. **Token presence is not reachability.** An AC that asserts
> a feature's text exists MUST be paired with an AC that asserts the control flow
> reaching that text is not blocked by a cap, a gate, or an early exit elsewhere in
> the system. AC-PHB-015 … AC-PHB-018 are that pairing.

## §A. Given-When-Then Scenarios

### GWT-1 — Adaptive interview exits early on clear answers

- **Given** a `/moai project` run whose Phase 0.3 (mode-detection.md) or
  Phase 1.5 (codebase-analysis.md) interview answers reach the sufficiency target
  (clarity ≥ 8 on the 0-10 scale, mirroring `plan/clarity-interview.md`) after
  round 1,
- **When** the interview loop evaluates clarity at the end of round 1,
- **Then** the interview terminates early (does NOT run the remaining static
  rounds), and the workflow proceeds to doc generation. (`clarity_threshold` (4)
  is the interview ENTRY floor, NOT this early-exit target.)

### GWT-2 — Interview data survives into harness generation

- **Given** a completed project interview that produced
  `.moai/project/harness-spec.yaml` (domain / goal / constraints / scope +
  verification / external_systems / ui_surface / team_sharing),
- **When** `harness-build-entry.md` Phase 1 Context-First Discovery runs,
- **Then** the domain / goal / constraints / scope fields are pre-satisfied from
  `harness-spec.yaml` and Discovery re-asks ONLY fields absent or ambiguous in
  that file — no already-answered field is re-asked.

### GWT-3 — Project flow never writes a SPEC

- **Given** the new adaptive interview + `harness-spec.yaml` write path,
- **When** a full `/moai project` run completes,
- **Then** no file is written under `.moai/specs/**`, and `harness-spec.yaml`
  is written under `.moai/project/`.

### GWT-4 — Fresh deploy carries the new behavior

- **Given** the post-change binary built via `make build`,
- **When** `moai init <sandbox>` deploys a fresh project,
- **Then** the deployed tree carries the adaptive clarity-scored interview, the
  `harness-spec.yaml` write path, and the extended interview axes.

### GWT-5 — The extended-axes round runs even when Stage A exits at round 1 (v0.2.0)

- **Given** a `/moai project` run whose Stage A answers reach the sufficiency
  target (clarity ≥ 8) after round 1 AND have answered all four REQUIRED base
  fields (`domain`, `goal`, `constraints`, `scope`),
- **When** Stage A terminates early (GWT-1),
- **Then** the workflow STILL runs the mandatory Stage B round (Round 4) and
  collects all four extended axes — the early exit skips the remaining **Stage A**
  rounds only, never Stage B; `harness-spec.yaml` is therefore written with its
  EXTENDED fields populated, not empty.

### GWT-6 — Stage A does not exit early while a REQUIRED base field is unanswered (v0.2.0)

- **Given** a `/moai project` run whose round-1 answers score clarity ≥ 8 but leave
  a REQUIRED base field unanswered (e.g. `mode-detection.md` has elicited vision
  and technology but not `constraints`),
- **When** the clarity loop evaluates the early-exit condition at the end of round 1,
- **Then** Stage A does NOT exit — it continues (up to `project.max_rounds`) until
  every REQUIRED base field is answered or the cap is reached; a high clarity score
  alone never satisfies the early exit.

### GWT-7 — Stage B runs after the abandon path and after the cap (v0.2.0)

- **Given** a `/moai project` run whose Stage A ends by abandon (clarity ≤ 3) or by
  exhausting `project.max_rounds` (3) without reaching the sufficiency target,
- **When** Stage A terminates,
- **Then** the workflow still runs the mandatory Stage B round — Stage B is exempt
  from the cap, the early-exit skip, AND the abandon path; the extended axes are
  collected on every Stage A exit route.

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-PHB-001 | REQ-PHB-001/002 | clarity_threshold consumed in both interview hosts (mode-detection.md Phase 0.3 + codebase-analysis.md Phase 1.5) |
| AC-PHB-002 | REQ-PHB-001/002 | adaptive round logic present (early-exit + additional-round) |
| AC-PHB-003 | REQ-PHB-003 | clarity-interview.md mechanism referenced/mirrored |
| AC-PHB-004 | REQ-PHB-004 | four new axes present in interview question set |
| AC-PHB-005 | REQ-PHB-005 | harness-spec.yaml 8-field schema present in doc-generation.md |
| AC-PHB-006 | REQ-PHB-005 | 8-field schema stated in spec.md §D |
| AC-PHB-007 | REQ-PHB-006 | doc-generation writes harness-spec.yaml under .moai/project/ |
| AC-PHB-008 | REQ-PHB-007 | meta-harness Phase 5.1 composes from harness-spec.yaml |
| AC-PHB-009 | REQ-PHB-008/009 | harness-build-entry Phase 1 consumes + pre-satisfies (no re-ask) |
| AC-PHB-010 | REQ-PHB-010 | NO-SPEC guard: no .moai/specs/ write path in project flow |
| AC-PHB-011 | REQ-PHB-011 | template↔local byte-parity on every touched file |
| AC-PHB-012 | REQ-PHB-011 | template neutrality + internal-content-leak guard green |
| AC-PHB-013 | REQ-PHB-004 | interview.yaml extended axes present (config-declared per Decision 1) |
| AC-PHB-014 | REQ-PHB-012 | moai init deploys adaptive interview + harness-spec.yaml path |
| AC-PHB-015 | REQ-PHB-013/014 | **REACHABILITY** — Round 4 documented as ALWAYS-RUNS + exempt from `max_rounds` / early-exit / clarity, in BOTH hosts |
| AC-PHB-016 | REQ-PHB-002/017 | **REACHABILITY** — the early-exit prose in BOTH hosts carries the required-base-fields precondition |
| AC-PHB-017 | REQ-PHB-015 | **REACHABILITY** — `interview.yaml` documents `max_rounds` as scoped to the Stage A clarity loop, not the whole interview |
| AC-PHB-018 | REQ-PHB-018 | **REACHABILITY** — each host's Stage A rounds cover all four REQUIRED base fields (explicit coverage mapping) |
| AC-PHB-019 | REQ-PHB-016 | REQUIRED vs EXTENDED field partition stated in spec.md §D AND in doc-generation.md |
| AC-PHB-020 | REQ-PHB-019 | F-4 — no `3-round` fixed-length claim in either host's frontmatter or the `project.md` routing table (both trees) |

## §C. Verification Commands (per AC)

### AC-PHB-001 (REQ-PHB-001/002)

```bash
# Phase 0.3 (new-project) interview host = mode-detection.md
grep -c "clarity_threshold" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1
grep -c -i "Phase 0.3" .claude/skills/moai/workflows/project/mode-detection.md        # expect >= 1
# Phase 1.5 (existing-project) interview host = codebase-analysis.md (NOT mode-detection.md)
grep -c "clarity_threshold" .claude/skills/moai/workflows/project/codebase-analysis.md   # expect >= 1
grep -c -i "Phase 1.5" .claude/skills/moai/workflows/project/codebase-analysis.md        # expect >= 1
```

Expected: `clarity_threshold` consumed in BOTH interview hosts; the Phase 0.3
anchor lives in mode-detection.md and the Phase 1.5 anchor in codebase-analysis.md.

### AC-PHB-002 (REQ-PHB-001/002)

```bash
# adaptive-loop + reconciled 0-10 scale language present in BOTH interview hosts:
for f in project/mode-detection.md project/codebase-analysis.md; do
  grep -c -i "clarity\|max_rounds" ".claude/skills/moai/workflows/$f"   # expect >= 1
  grep -c -i "sufficiency\|early.exit\|additional round\|abandon\|entry floor" ".claude/skills/moai/workflows/$f"   # expect >= 1 (reconciled scale language)
done
```

Expected: reconciled 0-10 scale adaptive-loop language present in both files
(sufficiency target ≥ 8 + entry floor + additional-round), NOT a hardcoded
3-round statement and NOT treating `clarity_threshold` (4) as the exit target.

### AC-PHB-003 (REQ-PHB-003)

```bash
for f in project/mode-detection.md project/codebase-analysis.md; do
  grep -c "clarity-interview" ".claude/skills/moai/workflows/$f"   # expect >= 1 (references the reference mechanism)
done
```

Expected: BOTH interview hosts cross-reference / mirror the
`plan/clarity-interview.md` mechanism (not a divergent reinvention).

### AC-PHB-004 (REQ-PHB-004)

```bash
# all four new axes elicited in BOTH interview hosts:
for f in project/mode-detection.md project/codebase-analysis.md; do
  grep -c -i "verification\|test command\|e2e" ".claude/skills/moai/workflows/$f"      # expect >= 1
  grep -c -i "ui.surface\|has-ui\|headless" ".claude/skills/moai/workflows/$f"         # expect >= 1
  grep -c -i "external.system\|database\|apis\|services" ".claude/skills/moai/workflows/$f"   # expect >= 1
  grep -c -i "team.shar\|team-shared\|solo" ".claude/skills/moai/workflows/$f"          # expect >= 1
done
```

Expected: all four new axes elicited in the interview question set of both files.

### AC-PHB-005 (REQ-PHB-005)

```bash
grep -c -E "domain|goal|constraints|scope|verification|external_systems|ui_surface|team_sharing" \
  .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 8 (all 8 schema fields named)
```

Expected: the 8-field `harness-spec.yaml` schema appears in the doc-generation
write step.

### AC-PHB-006 (REQ-PHB-005)

```bash
grep -c -E "domain|goal|constraints|scope|verification|external_systems|ui_surface|team_sharing" \
  .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md   # expect >= 8 (schema stated in §D)
grep -c "harness-spec.yaml" .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md   # expect >= 1
```

### AC-PHB-007 (REQ-PHB-006)

```bash
grep -c "\.moai/project/harness-spec.yaml" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c "\.moai/specs/.*harness-spec" .claude/skills/moai/workflows/project/doc-generation.md        # expect 0 (never under specs)
```

Expected: write target is `.moai/project/harness-spec.yaml`, never
`.moai/specs/`.

### AC-PHB-008 (REQ-PHB-007)

```bash
grep -c "harness-spec.yaml" .claude/skills/moai/workflows/project/meta-harness.md   # expect >= 1 (Phase 5.1 composes from it)
grep -n -i "Phase 5.1\|product.md\|structure.md\|tech.md" .claude/skills/moai/workflows/project/meta-harness.md | head   # context: compose site references all sources
```

Expected: Phase 5.1 composes from `harness-spec.yaml` PLUS product/structure/tech
(harness-spec.yaml no longer discarded).

### AC-PHB-009 (REQ-PHB-008/009)

```bash
grep -c "harness-spec.yaml" .claude/skills/moai/workflows/harness-build-entry.md   # expect >= 1 (Phase 1 consumes it)
grep -c -i "pre-satisf\|already-answered\|already answered\|absent or ambiguous\|only.*absent\|skip.*answered" \
  .claude/skills/moai/workflows/harness-build-entry.md   # expect >= 1 (no-duplicate-question language)
```

Expected: Phase 1 Discovery consumes `harness-spec.yaml` AND carries explicit
pre-satisfy / re-ask-only-absent language (REQ-PHB-009 no-duplicate).

### AC-PHB-010 (REQ-PHB-010) — NO-SPEC scope guard

```bash
# (a) NO-SPEC guard prose PRESERVED (must still exist — do NOT delete to satisfy a grep)
grep -cE "MUST NOT.*\.moai/specs|Never write to.*\.moai/specs" \
  .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1 (guard prose intact)
# (b) no NEW write path targeting .moai/specs/ (exclude the guard prose itself)
grep -rnE "\.moai/specs/" \
  .claude/skills/moai/workflows/project/mode-detection.md \
  .claude/skills/moai/workflows/project/codebase-analysis.md \
  .claude/skills/moai/workflows/project/doc-generation.md \
  | grep -viE "MUST NOT|Never write|forbidden|guard|scope guard|NO-SPEC" | wc -l   # expect 0 (no new .moai/specs write path)
# (c) harness-spec.yaml lives under .moai/project/
grep -c "\.moai/project/" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
```

Expected: the NO-SPEC guard prose is PRESERVED (>= 1 — the guard is never
deleted to satisfy a grep); no NEW `.moai/specs/` write path is introduced
(0 after excluding the guard prose itself); the artifact path is `.moai/project/`.

### AC-PHB-011 (REQ-PHB-011) — template↔local byte-parity

```bash
for f in project/mode-detection.md project/codebase-analysis.md project/doc-generation.md project/meta-harness.md harness-build-entry.md; do
  diff -q ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f" \
    && echo "PARITY OK: $f" || echo "DRIFT: $f"
done
# interview.yaml is ALSO touched (config-declared additional_axes) and is
# pre-existing byte-divergent — the fix is a wholesale template->local overwrite:
diff -q .moai/config/sections/interview.yaml \
  internal/template/templates/.moai/config/sections/interview.yaml \
  && echo "PARITY OK: interview.yaml" || echo "DRIFT: interview.yaml"
# expect: PARITY OK for every touched file (byte-identical mirror)
```

Expected: every touched file is byte-identical between the local and template
trees, including codebase-analysis.md and interview.yaml (the latter achieved by
wholesale template->local overwrite, per plan.md B-INTERVIEW-YAML-DRIFT).

### AC-PHB-012 (REQ-PHB-011) — neutrality guard

```bash
make build ; echo "exit=$?"                                  # expect 0
go test ./internal/template/... 2>&1 | tail -2               # expect ok (neutrality + internal-content-leak guards)
grep -rn "SPEC-PROJECT-HARNESS-BRIDGE-001" internal/template/templates/ | wc -l   # expect 0 (no internal SPEC ID leaked)
```

Expected: `make build` clean; template guards green; no internal SPEC ID / date /
SHA in the template tree.

### AC-PHB-013 (REQ-PHB-004) — interview.yaml axes (config-declared per resolved Decision 1)

```bash
# AC-PHB-013 APPLIES: the four new axes are config-declared in interview.yaml
# per resolved Decision 1 (additional_axes: block added to both trees).
grep -c -i "verification\|ui_surface\|external_systems\|team_sharing\|additional_axes" \
  .moai/config/sections/interview.yaml   # expect >= 1
diff -q .moai/config/sections/interview.yaml \
  internal/template/templates/.moai/config/sections/interview.yaml   # byte-parity
```

Expected: `interview.yaml` carries the four config-declared axes and the mirror
is byte-identical.

### AC-PHB-014 (REQ-PHB-012) — moai init resurrection-positive

```bash
T=$(mktemp -d) && ./bin/moai init "$T/sandbox" >/dev/null 2>&1
grep -c "clarity_threshold" "$T/sandbox/.claude/skills/moai/workflows/project/mode-detection.md"   # expect >= 1
grep -c "harness-spec.yaml" "$T/sandbox/.claude/skills/moai/workflows/project/doc-generation.md"   # expect >= 1
```

Expected: the deployed tree carries the adaptive interview (clarity_threshold
consumed) AND the harness-spec.yaml write path.

---

> **AC-PHB-015 … AC-PHB-020 — v0.2.0 amendment (REACHABILITY).** These ACs verify
> that the extended-axes round can actually RUN, not that its tokens are present.
> AC-PHB-004 (token presence) already passed while Round 4 was dead code; these ACs
> are the pairing that would have caught it.

### AC-PHB-015 (REQ-PHB-013/014) — REACHABILITY: Round 4 always runs, exempt from the cap

```bash
for f in project/mode-detection.md project/codebase-analysis.md; do
  P=".claude/skills/moai/workflows/$f"
  # (a) the exemption contract is stated somewhere in the host
  grep -c -i -E "always runs|unconditional|exempt from .*max_rounds|not counted against .*max_rounds|regardless of clarity|not skipped by the early.exit" "$P"   # expect >= 1
  # (b) the exemption is anchored to the Round 4 / Stage B section (not a stray mention elsewhere)
  grep -A 12 -i -E "^\*\*Round 4|Stage B" "$P" \
    | grep -c -i -E "always runs|unconditional|exempt|regardless of clarity"   # expect >= 1
  # (c) Stage A / Stage B structure is named (the two stages are distinguishable)
  grep -c -E "Stage A|Stage B" "$P"   # expect >= 2
done
```

Expected (BOTH hosts): the Round 4 / Stage B section explicitly states that the
round ALWAYS runs and is EXEMPT from `project.max_rounds`, from the Stage A
early-exit skip, and from clarity scoring. A host that carries a Round 4 heading
WITHOUT the exemption language FAILS this AC — that silent state is exactly the
v0.1.x defect (Round 4 present in prose, unreachable in control flow).

### AC-PHB-016 (REQ-PHB-002/017) — REACHABILITY: early exit is gated on required base fields

```bash
for f in project/mode-detection.md project/codebase-analysis.md; do
  P=".claude/skills/moai/workflows/$f"
  # (a) the early-exit prose carries the required-base-fields precondition
  grep -A 6 -i -E "Early exit|early-exit" "$P" \
    | grep -c -i -E "required base field|all four required|domain.*goal.*constraints.*scope"   # expect >= 1
  # (b) the early exit is scoped to Stage A only (it must NOT claim to skip "the interview")
  grep -A 6 -i -E "Early exit|early-exit" "$P" \
    | grep -c -i -E "Stage A|remaining Stage A rounds"   # expect >= 1
done
```

Expected (BOTH hosts): the early-exit clause states BOTH (i) that it fires only
once all four REQUIRED base fields are answered, and (ii) that it terminates
Stage A only — never the mandatory Stage B round.

### AC-PHB-017 (REQ-PHB-015) — REACHABILITY: `max_rounds` documented as Stage-A-scoped

```bash
# (a) interview.yaml scopes max_rounds to the Stage A clarity loop, in-place
grep -c -i -E "Stage A|clarity-scored discovery|caps only|does not cap the|not the whole interview" \
  .moai/config/sections/interview.yaml   # expect >= 1
# (b) the extended-axes round is named as NOT capped by max_rounds
grep -A 4 -i -E "max_rounds" .moai/config/sections/interview.yaml \
  | grep -c -i -E "Stage A|extended.axes round|Round 4|not capped|exempt"   # expect >= 1
# (c) byte-parity with the template mirror
diff -q .moai/config/sections/interview.yaml \
  internal/template/templates/.moai/config/sections/interview.yaml \
  && echo "PARITY OK: interview.yaml" || echo "DRIFT: interview.yaml"
```

Expected: a reader of `interview.yaml` ALONE cannot infer that the project
interview terminates after 3 rounds — the `project.max_rounds: 3` key is documented
in-place as capping only the Stage A clarity-scored discovery rounds, with the
mandatory extended-axes round called out as exempt.

### AC-PHB-018 (REQ-PHB-018) — REACHABILITY: Stage A covers all four required base fields

```bash
for f in project/mode-detection.md project/codebase-analysis.md; do
  P=".claude/skills/moai/workflows/$f"
  # (a) an explicit base-field coverage mapping exists
  grep -c -i -E "base.field coverage|Stage A coverage|required base fields" "$P"   # expect >= 1
  # (b) the mapping names all four required base fields
  grep -A 10 -i -E "base.field coverage|Stage A coverage" "$P" \
    | grep -c -i -E "domain|goal|constraints|scope"   # expect >= 4 (one line per field)
done
```

Expected: each host carries an explicit mapping from each REQUIRED base field
(`domain` / `goal` / `constraints` / `scope`) to the Stage A round that elicits it.
Per REQ-PHB-018, the existing-project host (`codebase-analysis.md`) MAY satisfy a
field via confident auto-population from the Phase 1 codebase analysis — such a
field is marked as auto-populated in the coverage mapping and is NOT re-asked
(consistent with §D E3). A host whose Stage A rounds leave a required base field
with no eliciting round and no auto-population source FAILS this AC (the v0.1.x
state: `mode-detection.md` had no constraints round; `codebase-analysis.md` had no
scope round).

### AC-PHB-019 (REQ-PHB-016) — REQUIRED vs EXTENDED field partition

```bash
# (a) spec.md §D states the partition
grep -c -E "REQUIRED|EXTENDED" .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md   # expect >= 2
# (b) doc-generation.md carries the same partition (the write site must know which
#     fields come from Stage A vs the mandatory Stage B round)
grep -c -i -E "REQUIRED|EXTENDED|Stage A|Stage B" \
  .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c -i -E "verification|ui_surface|external_systems|team_sharing" \
  .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1 (extended fields named)
```

Expected: both the SPEC §D schema and the `doc-generation.md` write site partition
the 8 fields into REQUIRED (Stage A) vs EXTENDED (mandatory Stage B).

### AC-PHB-020 (REQ-PHB-019) — F-4: no fixed `3-round` claim survives

```bash
# Whole-surface sweep across BOTH trees (local + template mirror):
grep -rn "3-round" \
  .claude/skills/moai/workflows/ \
  internal/template/templates/.claude/skills/moai/workflows/ \
  | wc -l   # expect 0
```

Expected: **0 matches.** The baseline (pre-amendment) count is **6**:
`project/codebase-analysis.md` frontmatter `description:` (×1) and the `project.md`
routing table (×2) — each mirrored into the template tree (×2 trees = 6). All six
must be reworded to describe the interview honestly (variable-length Stage A +
mandatory Stage B round). Note the `project.md` routing-table matches make
`project.md` a **7th file-pair** in the amendment touch set (spec.md `## Amendments`
records this scope delta).

## §D. Edge Cases

- **E1 clarity never reaches threshold**: when answers stay below
  `clarity_threshold` through `max_rounds` (3), **Stage A** stops at
  `max_rounds` (the cap is a hard stop) and proceeds with the best-available
  answers — it does NOT loop indefinitely. (v0.2.0 clarification: the cap ends
  **Stage A only**; the mandatory Stage B round still runs afterward — see E6.)
- **E2 harness-spec.yaml field with no interview answer**: a field the interview
  did not resolve is written as an explicit empty / null value (or omitted), and
  `harness-build-entry.md` Phase 1 treats it as ABSENT → eligible for re-ask
  (REQ-PHB-008 "absent or ambiguous").
- **E3 existing-project vs new-project axis differences**: Phase 1.5
  (existing-project) may auto-populate some axes from codebase analysis; those
  count as answered and must NOT be re-asked. Phase 0.3 (new-project) elicits all
  axes interactively.
- **E4 interview.yaml surface resolved at run-time**: the M1 clarification is
  resolved — config-declared `additional_axes:` block added per Decision 1;
  AC-PHB-013 applies. Do NOT silently pick a divergent path.
- **E5 template mirror drift**: a touched local file whose template mirror was
  not updated fails AC-PHB-011; the fix is to mirror, not to exempt the file.
- **E6 Stage A ends by cap or abandon (v0.2.0)**: on EVERY Stage A exit route —
  early exit, abandon (clarity ≤ 3), or `max_rounds` cap — the mandatory Stage B
  round still runs (REQ-PHB-013). A required base field left unanswered when the
  cap is reached is recorded as absent in `harness-spec.yaml` (per E2) and becomes
  eligible for re-ask in `harness-build-entry.md` Phase 1; it does NOT block
  Stage B and does NOT loop Stage A.
- **E7 Stage B answer refused / unknown (v0.2.0)**: an extended axis the user
  declines or cannot answer is written as an explicit empty / null value (per E2)
  — the round still RAN, so this is a legitimate empty, distinct from the v0.1.x
  defect where the round never ran at all. Both cases surface as ABSENT to
  `harness-build-entry.md`; only the latter is a bug.
- **E8 base field auto-populated from codebase analysis (v0.2.0)**: in the
  existing-project host, a REQUIRED base field confidently inferred from the
  Phase 1 analysis (e.g. `domain` from the detected stack) counts as ANSWERED for
  the REQ-PHB-002 / REQ-PHB-017 early-exit gate and is NOT re-asked. Auto-population
  satisfies the gate; it does not bypass it — an un-inferable required field still
  keeps Stage A running.

## §E. Quality Gates

- TRUST 5: Tested (this is a doc-only SPEC — verification is grep / diff /
  `make build` / `moai init`, not unit tests; state test ACs as N/A honestly);
  Readable (workflow prose stays consistent with surrounding sections); Unified
  (byte-identical template mirror); Secured (no new input surfaces; NO-SPEC guard
  preserved); Trackable (Conventional Commits per milestone, `🗿 MoAI` trailer).
- Neutrality: `template-neutrality-check.yaml` + `internal_content_leak_test.go`
  green on the final push.
- Non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code changed).

## §F. Definition of Done

1. All **20** ACs PASS (or documented N/A / PASS-WITH-DEBT with rationale) with
   verbatim command output recorded in progress.md §E.2.
2. Both interview phases (0.3 + 1.5) carry the adaptive loop + extended axes.
3. `harness-spec.yaml` (8-field schema) written under `.moai/project/`, consumed
   by `meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1 (no
   duplicate re-ask).
4. Every touched file byte-identical between local and template trees;
   `make build` + neutrality guards green; `moai init` resurrection-positive.
5. The two clarifications resolved before Implementation Kickoff Approval:
   (1) interview.yaml `additional_axes:` block added (config-declared) —
   AC-PHB-013 applies; (2) OVERWRITE re-run semantics confirmed.
6. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.

### v0.2.0 amendment — additional Definition-of-Done items

7. **Round 4 is REACHABLE** on both interview paths: the mandatory Stage B round
   runs after every Stage A exit route (early exit / abandon / cap), and both hosts
   state the exemption from `project.max_rounds`, from the early-exit skip, and
   from clarity scoring (AC-PHB-015).
8. The Stage A early exit is gated on completion of the four REQUIRED base fields
   in BOTH hosts (AC-PHB-016), and each host's Stage A rounds carry an explicit
   coverage mapping for all four (AC-PHB-018).
9. `interview.yaml` documents `project.max_rounds` as Stage-A-scoped (AC-PHB-017);
   the REQUIRED / EXTENDED field partition is stated in §D and in
   `doc-generation.md` (AC-PHB-019).
10. Zero `3-round` fixed-length claims remain across both trees — baseline 6 → 0
    (AC-PHB-020); the `project.md` file-pair is included in the mirror/parity sweep
    as the 7th touched pair.
11. The amendment closes via the `completed → in-progress → implemented → completed`
    path: `status` returns to `completed` on the sync commit, and the
    `## Amendments` HISTORY sub-section records prior version `0.1.1` +
    `prior_completed_sha: 0c9871b46b6719325427dc0126e4eb65d7b0f2d8`.
