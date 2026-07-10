# Acceptance Criteria — SPEC-HARNESS-VERIFY-PROMOTE-001

> SSOT for the AC matrix. 11 ACs covering all 9 REQs (100% AC→REQ coverage).
> Verification conventions: (1) every grep is anchored to a content token, not a
> line number; (2) preservation / NO-WRITE ACs assert absence explicitly;
> (3) template↔local parity is a byte `diff` check on each touched file; (4) this
> is a doc-only SPEC, so ACs are grep / diff / `make build` / `moai init` based
> (no Go test / coverage ACs). Grep tokens target the run-phase INSERT tokens
> described in spec.md §C/§D and plan.md §F.

## §A. Given-When-Then Scenarios

### GWT-1 — Harness offer promoted to interview final question, Phase 4.2 fallback retained

- **Given** a `/moai project` run whose adaptive interview
  (`SPEC-PROJECT-HARNESS-BRIDGE-001`) has confirmed the project type,
- **When** the interview reaches its final question,
- **Then** the harness-generation proposal ("이 프로젝트에 <type> 개발 하네스를
  생성할까요?") is surfaced as that final question, AND the Phase 4.2 next-step
  "Generate harness" menu option remains reachable as a fallback.

### GWT-2 — Every generated harness ships a runnable verify skill + disciplined specialists

- **Given** `harness-builder.md` GENERATE produces a harness set,
- **When** the artifacts are emitted,
- **Then** the set includes a mandatory `harness-<name>-verify` companion skill
  (codifying the build / launch / test recipe), AND every generated specialist
  agent body carries the tool-priority decision-tree block + the Skill-First
  execution block, AND the PLAN phase enforces the 3-7-specialists guardrail.

### GWT-3 — Namespace + scope invariants hold

- **Given** the promoted offer + verify-skill + specialist-rule edits,
- **When** a full `/moai project` → harness-generation run completes,
- **Then** generated artifacts use the `harness-*` namespace only (FROZEN guard
  rejecting `.claude/agents/moai/`, `.claude/skills/moai-*/`, `.claude/rules/moai/`
  intact), AND no file is written under `.moai/specs/**`.

### GWT-4 — Template↔local parity + neutrality green

- **Given** every edit made Template-First,
- **When** the byte-parity sweep + `make build` + neutrality guards run,
- **Then** every touched file is byte-identical between the local and template
  trees, the neutrality guards are green, and no internal SPEC ID appears in the
  template tree.

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-HVP-001 | REQ-HVP-001 | promoted-offer language present in meta-harness.md (final question) |
| AC-HVP-002 | REQ-HVP-001 | Phase 4.2 "Generate harness" fallback menu retained in meta-harness.md |
| AC-HVP-003 | REQ-HVP-002 | harness-build-entry.md carries the interview final-round offer |
| AC-HVP-004 | REQ-HVP-003 | harness-builder.md mandatory `harness-<name>-verify` skill clause present |
| AC-HVP-005 | REQ-HVP-004 | tool-priority decision-tree rule block present in harness-builder.md |
| AC-HVP-006 | REQ-HVP-005 | Skill-First execution rule block present in harness-builder.md |
| AC-HVP-007 | REQ-HVP-006 | 3-7 agent-count guardrail present in harness-builder.md PLAN section |
| AC-HVP-008 | REQ-HVP-007 | harness-* namespace + FROZEN guard language intact |
| AC-HVP-009 | REQ-HVP-008 | NO-SPEC guard: 0 `.moai/specs/` write path in the project-flow files |
| AC-HVP-010 | REQ-HVP-009 | template↔local byte-parity on every touched file |
| AC-HVP-011 | REQ-HVP-009 | template neutrality + internal-content-leak guard green |

## §C. Verification Commands (per AC)

### AC-HVP-001 (REQ-HVP-001) — promoted offer

```bash
grep -c -i "생성할까요\|generate.*harness.*for this project\|final question\|promoted offer" \
  .claude/skills/moai/workflows/project/meta-harness.md   # expect >= 1
```

Expected: the harness-generation proposal is surfaced as the interview's final
question (promoted from the buried Phase 4.2 menu).

### AC-HVP-002 (REQ-HVP-001) — Phase 4.2 fallback retained

```bash
grep -c -i "Phase 4.2\|Generate harness" .claude/skills/moai/workflows/project/meta-harness.md   # expect >= 1 (menu option retained)
```

Expected: the Phase 4.2 next-step "Generate harness" menu option is RETAINED (both
entry points reachable).

### AC-HVP-003 (REQ-HVP-002) — harness-build-entry final-round offer

```bash
grep -c -i "final.round\|final question\|생성할까요\|harness.*proposal" \
  .claude/skills/moai/workflows/harness-build-entry.md   # expect >= 1
```

Expected: `harness-build-entry.md` surfaces the harness-generation proposal as its
interview final-round offer.

### AC-HVP-004 (REQ-HVP-003) — mandatory verify skill

```bash
grep -c "harness-.*-verify\|harness-<name>-verify" .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
grep -c -i "run-skill-generator\|build.*launch.*test\|verification loop\|runnable check" \
  .claude/skills/moai/workflows/harness-builder.md   # expect >= 1 (mirrors the /run-skill-generator pattern)
```

Expected: GENERATE mandates a `harness-<name>-verify` companion skill (6th
artifact) codifying the build / launch / test recipe.

### AC-HVP-005 (REQ-HVP-004) — tool-priority decision tree

```bash
grep -c -i "tool priority\|category fit\|category-fit\|style preference" \
  .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
```

Expected: the specialist-generation template injects a tool-priority decision-tree
block ("use a tool when it is the category fit, not a style preference").

### AC-HVP-006 (REQ-HVP-005) — Skill-First execution rule

```bash
grep -c -i "skill-first\|read the relevant.*SKILL.md\|before.*file.*code work" \
  .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
```

Expected: the specialist-generation template injects a Skill-First execution rule
block ("read the relevant companion SKILL.md before any file / code work").

### AC-HVP-007 (REQ-HVP-006) — 3-7 agent-count guardrail

```bash
grep -c -i "3-7\|3 to 7\|3–7\|recurring same-instruction\|emit.*skill instead\|over-generation" \
  .claude/skills/moai/workflows/harness-builder.md   # expect >= 1 (in PLAN section)
grep -n -i "PLAN" .claude/skills/moai/workflows/harness-builder.md | head   # context: guardrail lives in PLAN phase
```

Expected: the PLAN phase states a 3-7-specialists-maximum guardrail with a
justify-each-or-emit-skill rule.

### AC-HVP-008 (REQ-HVP-007) — namespace + FROZEN guard intact

```bash
grep -c -i "harness-\*\|harness- namespace\|harness-<name>" .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
grep -c "\.claude/agents/moai/\|\.claude/skills/moai-\|\.claude/rules/moai/" \
  .claude/skills/moai/workflows/harness-builder.md   # expect >= 1 (FROZEN guard names the rejected paths)
```

Expected: `harness-*` namespace-only invariant + the FROZEN guard rejecting the
`moai-*` reserved paths remain intact (not weakened by this SPEC's edits).

### AC-HVP-009 (REQ-HVP-008) — NO-SPEC scope guard

```bash
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/meta-harness.md \
  .claude/skills/moai/workflows/harness-build-entry.md | wc -l   # expect 0 (no specs write path)
```

Expected: 0 `.moai/specs/` references in the two project-flow files touched by this
SPEC (project flow never writes a SPEC).

### AC-HVP-010 (REQ-HVP-009) — template↔local byte-parity

```bash
for f in project/meta-harness.md harness-build-entry.md harness-builder.md; do
  diff -q ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f" \
    && echo "PARITY OK: $f" || echo "DRIFT: $f"
done
# expect: PARITY OK for every touched file (byte-identical mirror)
```

Expected: every touched file is byte-identical between the local and template trees.

### AC-HVP-011 (REQ-HVP-009) — neutrality guard

```bash
make build ; echo "exit=$?"                                  # expect 0
go test ./internal/template/... 2>&1 | tail -2               # expect ok (neutrality + internal-content-leak guards)
grep -rn "SPEC-HARNESS-VERIFY-PROMOTE-001" internal/template/templates/ | wc -l   # expect 0 (no internal SPEC ID leaked)
```

Expected: `make build` clean; template guards green; no internal SPEC ID / date /
SHA in the template tree.

## §D. Edge Cases

- **E1 no discoverable recipe**: when the project has no discoverable build /
  launch / test recipe, the mandatory `harness-<name>-verify` skill is STILL
  emitted — as a stub documenting "no recipe found" — rather than omitted (per the
  the plan.md verify-skill enforcement-surface default resolution).
- **E2 user declines the promoted offer**: a user who declines the interview's
  final-question harness offer must still reach the RETAINED Phase 4.2 "Generate
  harness" menu option later (both entry points reachable — AC-HVP-002).
- **E3 `<type>` token source**: the `<type>` in the promoted offer resolves from
  `harness-spec.yaml` `domain` (SPEC-1) when present; when `harness-spec.yaml` is
  absent, it falls back to the mode-detection confirmed project type.
- **E4 specialist count at boundary**: exactly 3 or exactly 7 specialists is WITHIN
  the guardrail; ≤2 or ≥8 triggers the justify-or-emit-skill path (AC-HVP-007).
- **E5 template mirror drift**: a touched local file whose template mirror was not
  updated fails AC-HVP-010; the fix is to mirror, not to exempt the file.
- **E6 depends_on unfulfilled**: if `SPEC-PROJECT-HARNESS-BRIDGE-001` is not
  `status: completed` at `/moai run` time, the Phase 0.5 Depends_on Pre-flight
  Check blocks with a 3-option AskUserQuestion (wait / override / abort); `override`
  requires the logged `--ignore-deps` rationale. This is a run-phase gate, not an
  AC of this SPEC, but is recorded here for run-phase awareness.

## §E. Quality Gates

- TRUST 5: Tested (this is a doc-only SPEC — verification is grep / diff /
  `make build` / `moai init`, not unit tests; state test ACs as N/A honestly);
  Readable (workflow prose stays consistent with surrounding sections; injected
  specialist blocks stay short); Unified (byte-identical template mirror); Secured
  (no new input surfaces; FROZEN + NO-SPEC guards preserved); Trackable
  (Conventional Commits per milestone, `🗿 MoAI` trailer).
- Neutrality: `template-neutrality-check.yaml` + `internal_content_leak_test.go`
  green on the final push.
- Non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code changed).

## §F. Definition of Done

1. All 11 ACs PASS (or documented N/A / PASS-WITH-DEBT with rationale) with
   verbatim command output recorded in progress.md §E.2.
2. The harness-generation offer is promoted to the interview's final question in
   both `meta-harness.md` and `harness-build-entry.md`, with the Phase 4.2 menu
   RETAINED as a fallback.
3. `harness-builder.md` GENERATE mandates the `harness-<name>-verify` skill (6th
   artifact) + injects the two short specialist rule blocks + states the 3-7 PLAN
   guardrail.
4. `harness-*` namespace-only + FROZEN guard + NO-SPEC guard all preserved; every
   touched file byte-identical between local and template trees; `make build` +
   neutrality guards green; `moai init` resurrection-positive.
5. The two plan.md open clarifications resolved (promoted-offer placement seam +
   verify-skill enforcement surface) before Implementation Kickoff Approval.
6. Depends_on gate honored: run-phase entry only after
   `SPEC-PROJECT-HARNESS-BRIDGE-001` is `completed`, or with a logged
   `--ignore-deps` override.
7. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.
