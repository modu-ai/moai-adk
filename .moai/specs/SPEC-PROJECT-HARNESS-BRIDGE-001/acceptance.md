# Acceptance Criteria — SPEC-PROJECT-HARNESS-BRIDGE-001

> SSOT for the AC matrix. 14 ACs covering all 12 REQs (100% AC→REQ coverage).
> Verification conventions: (1) every grep is anchored to a content token, not a
> line number; (2) preservation / NO-WRITE ACs assert absence explicitly;
> (3) template↔local parity is a byte `diff` check on each touched file; (4) this
> is a doc-only SPEC, so ACs are grep / diff / `make build` / `moai init` based
> (no Go test / coverage ACs).

## §A. Given-When-Then Scenarios

### GWT-1 — Adaptive interview exits early on clear answers

- **Given** a `/moai project` run whose Phase 0.3 (or 1.5) interview answers
  reach `interview.yaml` `clarity_threshold` (4) after round 1,
- **When** the interview loop evaluates clarity at the end of round 1,
- **Then** the interview terminates early (does NOT run the remaining static
  rounds), and the workflow proceeds to doc generation.

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

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-PHB-001 | REQ-PHB-001/002 | clarity_threshold consumed in mode-detection.md (both phases) |
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
| AC-PHB-013 | REQ-PHB-004 | interview.yaml extended axes present (if config-declared) |
| AC-PHB-014 | REQ-PHB-012 | moai init deploys adaptive interview + harness-spec.yaml path |

## §C. Verification Commands (per AC)

### AC-PHB-001 (REQ-PHB-001/002)

```bash
grep -c "clarity_threshold" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1
# both interview phases reference the threshold (Phase 0.3 + Phase 1.5):
grep -n -i "Phase 0.3\|Phase 1.5" .claude/skills/moai/workflows/project/mode-detection.md | wc -l   # expect >= 2
```

Expected: `clarity_threshold` consumed; both phase anchors present.

### AC-PHB-002 (REQ-PHB-001/002)

```bash
grep -c -i "clarity\|max_rounds" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1
grep -c -i "early\|exit\|until\|additional round\|below threshold" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1 (adaptive-loop language, not fixed-3)
```

Expected: adaptive-loop language present (early-exit + additional-round), NOT a
hardcoded 3-round statement.

### AC-PHB-003 (REQ-PHB-003)

```bash
grep -c "clarity-interview" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1 (references the reference mechanism)
```

Expected: `mode-detection.md` cross-references / mirrors the
`plan/clarity-interview.md` mechanism (not a divergent reinvention).

### AC-PHB-004 (REQ-PHB-004)

```bash
grep -c -i "verification\|test command\|e2e" .claude/skills/moai/workflows/project/mode-detection.md      # expect >= 1
grep -c -i "ui.surface\|has-ui\|headless" .claude/skills/moai/workflows/project/mode-detection.md         # expect >= 1
grep -c -i "external.system\|database\|apis\|services" .claude/skills/moai/workflows/project/mode-detection.md   # expect >= 1
grep -c -i "team.shar\|team-shared\|solo" .claude/skills/moai/workflows/project/mode-detection.md          # expect >= 1
```

Expected: all four new axes elicited in the interview question set.

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
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/mode-detection.md \
  .claude/skills/moai/workflows/project/doc-generation.md | wc -l   # expect 0 (no specs write path in interview/harness-spec sections)
grep -c "\.moai/project/" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1 (harness-spec.yaml lives under project/)
```

Expected: 0 `.moai/specs/` references in the interview + harness-spec sections;
the artifact path is `.moai/project/`.

### AC-PHB-011 (REQ-PHB-011) — template↔local byte-parity

```bash
for f in project/mode-detection.md project/doc-generation.md project/meta-harness.md harness-build-entry.md; do
  diff -q ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f" \
    && echo "PARITY OK: $f" || echo "DRIFT: $f"
done
# expect: PARITY OK for every touched file (byte-identical mirror)
```

Expected: every touched file is byte-identical between the local and template
trees. (If `interview.yaml` is touched per the M1 clarification, add it to the
diff loop.)

### AC-PHB-012 (REQ-PHB-011) — neutrality guard

```bash
make build ; echo "exit=$?"                                  # expect 0
go test ./internal/template/... 2>&1 | tail -2               # expect ok (neutrality + internal-content-leak guards)
grep -rn "SPEC-PROJECT-HARNESS-BRIDGE-001" internal/template/templates/ | wc -l   # expect 0 (no internal SPEC ID leaked)
```

Expected: `make build` clean; template guards green; no internal SPEC ID / date /
SHA in the template tree.

### AC-PHB-013 (REQ-PHB-004) — interview.yaml axes (conditional on M1 clarification)

```bash
# Applies ONLY if the four new axes are config-declared in interview.yaml (per the
# M1 [NEEDS CLARIFICATION: interview.yaml schema surface] resolution). If the axes
# are prose-only in mode-detection.md, this AC is documented N/A with rationale.
grep -c -i "verification\|ui_surface\|external_systems\|team_sharing\|additional_axes" \
  .moai/config/sections/interview.yaml   # expect >= 1 (when config-declared)
diff -q .moai/config/sections/interview.yaml \
  internal/template/templates/.moai/config/sections/interview.yaml   # byte-parity when touched
```

Expected: when the axes are config-declared, `interview.yaml` carries them and
the mirror is byte-identical. Otherwise mark N/A with the clarification rationale.

### AC-PHB-014 (REQ-PHB-012) — moai init resurrection-positive

```bash
T=$(mktemp -d) && ./bin/moai init "$T/sandbox" >/dev/null 2>&1
grep -c "clarity_threshold" "$T/sandbox/.claude/skills/moai/workflows/project/mode-detection.md"   # expect >= 1
grep -c "harness-spec.yaml" "$T/sandbox/.claude/skills/moai/workflows/project/doc-generation.md"   # expect >= 1
```

Expected: the deployed tree carries the adaptive interview (clarity_threshold
consumed) AND the harness-spec.yaml write path.

## §D. Edge Cases

- **E1 clarity never reaches threshold**: when answers stay below
  `clarity_threshold` through `max_rounds` (3), the interview stops at
  `max_rounds` (the cap is a hard stop) and proceeds with the best-available
  answers — it does NOT loop indefinitely.
- **E2 harness-spec.yaml field with no interview answer**: a field the interview
  did not resolve is written as an explicit empty / null value (or omitted), and
  `harness-build-entry.md` Phase 1 treats it as ABSENT → eligible for re-ask
  (REQ-PHB-008 "absent or ambiguous").
- **E3 existing-project vs new-project axis differences**: Phase 1.5
  (existing-project) may auto-populate some axes from codebase analysis; those
  count as answered and must NOT be re-asked. Phase 0.3 (new-project) elicits all
  axes interactively.
- **E4 interview.yaml surface undecided at run-time**: if the M1 clarification is
  unresolved, apply the plan.md default (config-declared `additional_axes:`) and
  record the decision in progress.md §E.2; do NOT silently pick a divergent path.
- **E5 template mirror drift**: a touched local file whose template mirror was
  not updated fails AC-PHB-011; the fix is to mirror, not to exempt the file.

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

1. All 14 ACs PASS (or documented N/A / PASS-WITH-DEBT with rationale) with
   verbatim command output recorded in progress.md §E.2.
2. Both interview phases (0.3 + 1.5) carry the adaptive loop + extended axes.
3. `harness-spec.yaml` (8-field schema) written under `.moai/project/`, consumed
   by `meta-harness.md` Phase 5.1 and `harness-build-entry.md` Phase 1 (no
   duplicate re-ask).
4. Every touched file byte-identical between local and template trees;
   `make build` + neutrality guards green; `moai init` resurrection-positive.
5. The two `[NEEDS CLARIFICATION]` markers resolved (interview.yaml surface +
   re-run semantics) before Implementation Kickoff Approval.
6. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.
