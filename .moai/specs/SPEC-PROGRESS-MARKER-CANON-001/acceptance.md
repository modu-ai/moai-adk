# SPEC-PROGRESS-MARKER-CANON-001 — Acceptance Criteria

All AC are falsifiable and grep/test-verifiable. Commands containing a pipe are placed in fenced code blocks (never inline table cells) to avoid vacuous-pipe AC ambiguity.

## D. AC Matrix

| AC ID | REQ | Severity | Summary |
|-------|-----|----------|---------|
| AC-PMC-001 | REQ-PMC-001 | MUST | lifecycle-sync-gate.md worked example uses §E.4 for Sync heading (convention B) |
| AC-PMC-002 | REQ-PMC-002 | MUST | Worked-example trace stays H-4-valid: a literal §E.2 + §E.5 survive INSIDE the worked-example excerpt (region-scoped, not whole-file) |
| AC-PMC-003 | REQ-PMC-003 | MUST | Heuristic table + JSON audit excerpt unchanged |
| AC-PMC-004 | REQ-PMC-004 | MUST | era.go corrects the THREE comment framing sites (L33 + L91 + L120) to run-evidence/Mx-completion + string-presence framing (L122 is an executable `return` string → verbatim-stay) |
| AC-PMC-004b | REQ-PMC-002 | MUST | Worked-example §E.4 Sync heading co-exists with a §E.2 Run-phase Evidence heading after the flip (trace-narration consistency) |
| AC-PMC-005 | REQ-PMC-005 | MUST | era.go executable logic byte-identical (comment-only diff) |
| AC-PMC-006 | REQ-PMC-006 | MUST | `go test ./internal/spec/...` + `go build ./...` pass identically post-edit |
| AC-PMC-007 | REQ-PMC-007 | MUST | manager-spec.md contains progress.md skeleton-generation instruction |
| AC-PMC-008 | REQ-PMC-008 | MUST | Skeleton instruction enumerates all 5 §E headings in convention-B order, minimal placeholders |
| AC-PMC-009 | REQ-PMC-009 | MUST | Skeleton instruction does NOT authorize plan-phase §E.2-§E.5 population |
| AC-PMC-010 | REQ-PMC-010 | MUST | manager-spec.md mirror parity restored + embedded.go regenerated |
| AC-PMC-011 | REQ-PMC-011 | MUST | lifecycle-sync-gate.md edited single-copy (no mirror created) |
| AC-PMC-012 | REQ-PMC-012 | MUST | era.go edited single-copy (no mirror created) — promoted SHOULD→MUST per D5 (traces to REQ-PMC-012 "shall"; binary verifiable fact: a mirror either exists or it does not) |

## D.1 Given-When-Then Scenarios

### Scenario 1 — Worked-example convention-B alignment (AC-PMC-001, AC-PMC-002, AC-PMC-003)

**Given** `lifecycle-sync-gate.md` worked example currently shows `## §E.2 Sync-phase Audit-Ready Signal` (convention A, ~L297),
**When** M2 aligns the worked example to convention B,
**Then** the Sync heading reads `## §E.4 Sync-phase Audit-Ready Signal`, a `## §E.2 Run-phase Evidence` heading appears, `## §E.5 Mx-phase Audit-Ready Signal` is retained, and the H-1..H-6 heuristic table + JSON audit excerpt are unchanged.

> **Region-scoping note (D1)**: §E.2 and §E.5 also appear OUTSIDE the worked example — in the out-of-scope heuristic table (~L41-43), the prose at ~L51 and ~L165, and the Go-test-binding excerpt at ~L359-360. A whole-file `grep -c '§E.2'` can therefore NEVER reach 0 and would be a vacuous H-4 guard. The verifications below are **region-scoped** to the worked-example excerpt. Line numbers below are illustrative (~L294-318); the run-phase agent MUST re-locate the worked-example region by content (the block bounded by the `## §E.2`/`## §E.5` headings and the auto-detection trace), NOT by hardcoded line numbers, since prior edits may have drifted them.

Verification (run from project root):

```bash
# AC-PMC-001: convention-A Sync heading gone, convention-B Sync heading present (whole-file ok — these exact heading strings only ever lived in the worked example)
grep -c '§E.2 Sync-phase Audit-Ready Signal' .claude/rules/moai/workflow/lifecycle-sync-gate.md   # expect 0
grep -c '§E.4 Sync-phase Audit-Ready Signal' .claude/rules/moai/workflow/lifecycle-sync-gate.md   # expect >=1
```

```bash
# AC-PMC-002: H-4 detection still valid — a literal §E.2 AND §E.5 survive INSIDE the worked-example region.
# Region anchored by content: from the first '## §E.' heading of the excerpt to the trace's final ClassifyEra line.
# (Re-locate the start/end line numbers by content before running; the example range below is ~L294-318.)
sed -n '294,318p' .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -c '§E.2'   # expect >=1 (the '## §E.2 Run-phase Evidence' heading + trace refs)
sed -n '294,318p' .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -c '§E.5'   # expect >=1 (the '## §E.5 Mx-phase Audit-Ready Signal' heading + trace refs)
```

```bash
# AC-PMC-003: heuristic table rows + JSON excerpt unchanged. Inspect that no '+' (added) diff line touches a heuristic-table row or the JSON heuristic_matched string.
git diff .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -E '^\+' | grep -E 'H-1|H-2|H-3|H-4|H-5|H-6|heuristic_matched'
# expect: no output (no added line touches a heuristic table row or the JSON heuristic_matched string)
```

```bash
# AC-PMC-004b (D4 — trace-narration consistency): after the flip, the worked-example region MUST contain
# BOTH a '## §E.4 Sync-phase Audit-Ready Signal' heading AND a '## §E.2 Run-phase Evidence' heading,
# so the convention-A '§E.2 = Sync' framing is fully replaced (no half-flip where §E.2 still labels Sync).
# Re-locate the region by content before running; example range ~L294-318.
sed -n '294,318p' .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -c '## §E.4 Sync-phase Audit-Ready Signal'   # expect >=1
sed -n '294,318p' .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -c '## §E.2 Run-phase Evidence'            # expect >=1
# AND the convention-A heading must NOT survive anywhere in the region:
sed -n '294,318p' .claude/rules/moai/workflow/lifecycle-sync-gate.md | grep -c '## §E.2 Sync-phase Audit-Ready Signal' # expect 0
```

### Scenario 2 — era.go comment-only correction (AC-PMC-004, AC-PMC-005, AC-PMC-006)

**Given** `era.go` frames §E.2 as the "sync" gate at THREE comment sites — L33 const comment (`EraV3R5 — V3R5 SPECs with §E.2 sync but missing sync_commit_sha`), L91 ClassifyEra doc-comment (`H-3: §E.2 present but sync_commit_sha empty/missing → V3R5`), and L120 inline H-3 comment (`// H-3: §E.2 present but sync_commit_sha empty/missing → V3R5`),
**When** M1 corrects the comments at ALL three sites to convention-B semantics,
**Then** each corrected comment describes §E.2 as the §E-section run-evidence-start marker and §E.5 as the Mx-completion marker, clarifies that classification is string-presence-based (the `hasSyncSection`/`hasMxSection` names are misnomers), and the executable logic is byte-identical so `go test` passes unchanged.

> **D2 scope note (three comment sites only)**: The misleading "§E.2-means-sync gate" framing lives at exactly THREE `//` comment sites — L33, L91, L120. The fourth occurrence at L122 (`return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"`) is an **executable Go `return` statement, NOT a comment** — it MUST stay byte-identical (editing it would fail AC-PMC-005's comment-only diff gate) and belongs to the **verbatim-stay set**. The run-phase agent corrects only the three comment sites; the L122 return string, the literal field name `sync_commit_sha` everywhere, and the H-4 `§E.2 + §E.5` enumeration are all accurate and stay verbatim. The verification below is a single positive all-sites grep asserting the corrected canonical phrase appears at least once per corrected site (`>= 3`); no negative grep is used (a negative `grep -v 'sync_commit_sha'` would be self-stripping — every framing line contains `sync_commit_sha` — and could not distinguish a real three-site fix from a one-site no-op).

Verification:

```bash
# AC-PMC-005: the diff touches ONLY comment lines (every added/removed non-blank line starts with optional whitespace then //)
git diff internal/spec/era.go
# Manual gate: inspect that every '+'/'-' content line (excluding the @@ hunk headers) is a // comment line.
```

```bash
# AC-PMC-005 (mechanical): count added non-comment, non-blank lines — expect 0
git diff internal/spec/era.go | grep -E '^\+[^+]' | grep -vE '^\+\s*//' | grep -vE '^\+\s*$'
# expect: no output (no added line is a non-comment code line)
```

```bash
# AC-PMC-004 (positive, all-sites): the corrected canonical phrase appears once per corrected comment site.
# The run-phase agent writes the single canonical phrase "run-evidence start marker" into each of the
# THREE corrected comments (L33, L91, L120). >=3 proves all three sites were corrected (a one-site no-op
# yields 1, not 3). Single sound positive grep — no negative grep (a negative would be self-stripping
# since every framing line contains 'sync_commit_sha').
grep -c 'run-evidence start marker' internal/spec/era.go   # expect >=3 (one per corrected comment site L33/L91/L120)
```

```bash
# AC-PMC-006: tests + build pass post-edit (compare to pre-edit baseline captured in Pre-flight C.2/C.3)
go test ./internal/spec/... 2>&1 | tail -3   # expect: ok  .../internal/spec ... (PASS, same as baseline)
go build ./... ; echo "build exit: $?"        # expect: build exit: 0
```

### Scenario 3 — manager-spec.md skeleton-generation instruction (AC-PMC-007, AC-PMC-008, AC-PMC-009)

**Given** `manager-spec.md` has no progress.md skeleton-generation instruction (SPECs currently drift into §F.* markers),
**When** M3 adds the instruction,
**Then** the agent body directs manager-spec to emit a canonical progress.md skeleton with all 5 §E placeholder headings in convention-B order, the skeleton is minimal (heading + one-line placeholder), and the instruction does not authorize plan-phase population of §E.2-§E.5 evidence.

Verification:

```bash
# AC-PMC-007 + AC-PMC-008: all 5 convention-B §E headings present in manager-spec.md body
grep -c '§E.1 Plan-phase Audit-Ready Signal' .claude/agents/moai/manager-spec.md   # expect >=1
grep -c '§E.2 Run-phase Evidence' .claude/agents/moai/manager-spec.md              # expect >=1
grep -c '§E.3 Run-phase Audit-Ready Signal' .claude/agents/moai/manager-spec.md    # expect >=1
grep -c '§E.4 Sync-phase Audit-Ready Signal' .claude/agents/moai/manager-spec.md   # expect >=1
grep -c '§E.5 Mx-phase Audit-Ready Signal' .claude/agents/moai/manager-spec.md     # expect >=1
```

```bash
# AC-PMC-008: instruction references "skeleton" + plan-phase emission
grep -in 'progress.md skeleton\|skeleton-generation\|placeholder heading' .claude/agents/moai/manager-spec.md   # expect >=1
```

```bash
# AC-PMC-009: instruction preserves the existing Forbidden-modifications boundary (manager-spec does NOT populate §E.2-§E.5 evidence)
grep -n 'placeholder headings only\|do not populate\|belongs to manager-develop\|belongs to manager-docs' .claude/agents/moai/manager-spec.md   # expect >=1
```

### Scenario 4 — Mirror parity + single-copy edits (AC-PMC-010, AC-PMC-011, AC-PMC-012)

**Given** `manager-spec.md` had byte-parity with its template mirror at plan-phase,
**When** M3 edits both copies and runs `make build`,
**Then** parity is restored, `embedded.go` regenerates, and the no-mirror files (lifecycle-sync-gate.md, era.go) are edited single-copy.

Verification:

```bash
# AC-PMC-010: parity restored after edit
diff -q .claude/agents/moai/manager-spec.md internal/template/templates/.claude/agents/moai/manager-spec.md && echo "PARITY OK"
```

```bash
# AC-PMC-010: embedded.go regenerated (modified in working tree after make build)
git status --porcelain internal/template/embedded.go   # expect: a 'M' status line (regenerated)
```

```bash
# AC-PMC-011: lifecycle-sync-gate.md has no template mirror (none was created)
ls internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md 2>/dev/null && echo "UNEXPECTED MIRROR" || echo "NO MIRROR (correct)"
```

```bash
# AC-PMC-012: era.go has no template mirror
ls internal/template/templates/internal/spec/era.go 2>/dev/null && echo "UNEXPECTED" || echo "NO MIRROR (correct)"
```

## D.2 Edge Cases

- **EC-1 (H-4 regression guard)**: Because era.go greps the literal substring `§E.2`, the worked-example edit MUST keep a literal `§E.2` somewhere in the excerpt (now as `§E.2 Run-phase Evidence`). If the edit removed `§E.2` entirely, the worked-example trace narration (which asserts H-4) would become internally contradictory. AC-PMC-002 guards this.
- **EC-2 (grandfather false-edit)**: A run-phase agent might be tempted to "fix" the 16 §F.* SPECs while aligning conventions. CON-PMC-003 + the Out-of-Scope §B entry forbid this. Verification: `git diff --name-only .moai/specs/` shows ONLY `SPEC-PROGRESS-MARKER-CANON-001/` paths (no other SPEC dir touched).
- **EC-3 (variable-rename creep)**: The misnomer variables `hasSyncSection`/`hasMxSection` invite a rename. Out-of-scope §B forbids it; AC-PMC-005's comment-only-diff gate catches any rename (a rename touches non-comment lines).
- **EC-4 (template internal-content leak)**: The manager-spec.md body skeleton instruction must use generic §E heading prose, not embed this SPEC's ID. CON-PMC-005. Verification: `grep -c 'PROGRESS-MARKER-CANON' internal/template/templates/.claude/agents/moai/manager-spec.md` expects 0.

## D.3 Quality Gate Criteria

- `go test ./internal/spec/...` PASS (preservation — REQ-PMC-006).
- `go build ./...` exit 0.
- `make build` succeeds, embedded.go regenerated.
- spec-lint clean on this SPEC's 3 artifacts (frontmatter 12-field schema, MissingExclusions satisfied by §B H3 sub-headings).
- Template neutrality CI guard (`go test ./internal/template/... -run TestTemplateNeutralityAudit`) passes (manager-spec.md mirror carries no forbidden internal-content class).

## Definition of Done

- [ ] AC-PMC-001 through AC-PMC-012 all PASS with cited command output.
- [ ] era.go diff is comment-only (mechanical grep gate AC-PMC-005 returns empty).
- [ ] `go test ./internal/spec/...` identical to baseline (REQ-PMC-006).
- [ ] manager-spec.md + mirror byte-parity, embedded.go regenerated.
- [ ] No existing SPEC progress.md edited (EC-2 git diff scope check).
- [ ] This SPEC's own progress.md uses convention-B §E markers (dogfood).
- [ ] All commits scope the full SPEC-ID (close-subject full-ID mandate, CON-PMC-004).
