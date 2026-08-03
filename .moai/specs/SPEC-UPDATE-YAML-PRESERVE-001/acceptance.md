# Acceptance — SPEC-UPDATE-YAML-PRESERVE-001

All commands are run from the repository root `/Users/goos/MoAI/moai-adk-go`.

## §D AC Matrix

| AC | REQ | Milestone | Severity | Verification |
|----|-----|-----------|----------|--------------|
| AC-UYP-001 | REQ-UYP-001 | M3 | MUST | golden test — comment count preserved |
| AC-UYP-002 | REQ-UYP-002 | M3 | MUST | golden test — key order preserved |
| AC-UYP-003 | REQ-UYP-003 | M3 | MUST | golden test — quoting preserved |
| AC-UYP-004 | REQ-UYP-004 | M2 | MUST | `SetIndent(2)` present + indent-width assertion |
| AC-UYP-005 | REQ-UYP-005 | M3 | MUST | byte-identity on no-edit merge |
| AC-UYP-006 | REQ-UYP-006 | M4 | MUST | user-added key survives |
| AC-UYP-007 | REQ-UYP-007 | M4 | SHOULD | stderr advisory names retained key |
| AC-UYP-008 | REQ-UYP-008 | M4 | MUST | 3-way retention ⊇ 2-way retention |
| AC-UYP-009 | REQ-UYP-009 | M2 | MUST | existing decision tests still green |
| AC-UYP-010 | REQ-UYP-010 | M2 | MUST | system-field tests still green |
| AC-UYP-011 | REQ-UYP-011 | M2 | MUST | invalid-YAML error tests still green |
| AC-UYP-012 | REQ-UYP-012 | M1 | MUST | alias node not expanded |
| AC-UYP-013 | REQ-UYP-013 | M1 | SHOULD | merge key treated as ordinary key |
| AC-UYP-014 | REQ-UYP-014 | M1 | MUST | sequence replaced wholesale |
| AC-UYP-015 | REQ-UYP-015 | M2 | MUST | multi-document input errors |
| AC-UYP-016 | REQ-UYP-016 | M1 | MUST | null ≠ empty mapping |
| AC-UYP-017 | REQ-UYP-017 | M3 | MUST | golden test enumerates all templates |
| AC-UYP-018 | REQ-UYP-018 | M4 | MUST | no drop-asserting test remains |
| AC-UYP-019 | REQ-UYP-019 | M6 | MUST | no fixture under templates/ |
| AC-UYP-020 | §D constraint | M6 | MUST | coverage ≥ 88.9% baseline |
| AC-UYP-021 | §E success | M6 | MUST | end-to-end `moai update` preserves (cache.yaml + workflow.yaml) |
| AC-UYP-022 | plan M3 | M3 | MUST | falsifiability — test RED on reverted impl |
| AC-UYP-023 | REQ-UYP-011 | M3 | MUST | `quality.yaml.tmpl` parses via node decoder (nil error) |

## §D.1 Given-When-Then Scenarios

### AC-UYP-001 — comments preserved on merge

**Given** `internal/template/templates/.moai/config/sections/cache.yaml` (16 comment-bearing lines, measured this session)
**When** it is passed to `MergeYAML3Way(src, src, src)`
**Then** the output contains the same number of `#`-bearing lines as the input.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_Comments'
```
The trailing `grep -q '^--- PASS: …'` is the existence assertion (plan-audit D6): a missing or mis-named test fails the AC instead of passing vacuously with `no tests to run`.

Falsification baseline (must FAIL before the fix — see AC-UYP-022):
```bash
# with the map round-trip restored, this same test reports comments in=16 out=0
```

### AC-UYP-002 — key order preserved

**Given** a mapping whose source key order is `enabled, session_ttl, spec_ttl, min_cacheable_tokens`
**When** merged with no user edit
**Then** the output key order is identical and is NOT the alphabetized `enabled, min_cacheable_tokens, session_ttl, spec_ttl`.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_KeyOrder -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_KeyOrder'
```

### AC-UYP-003 — scalar quoting preserved

**Given** `session_ttl: "1h"` in the source
**When** the value is carried through unchanged
**Then** the output line is `session_ttl: "1h"`, not `session_ttl: 1h`.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Quoting -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_Quoting'
```

### AC-UYP-004 — indentation pinned to 2

**Given** the encoder configuration in the merge implementation
**When** any nested mapping is emitted
**Then** each nesting level adds exactly 2 spaces.

```bash
grep -n 'SetIndent(2)' internal/cli/update/backup/*.go
# expect: at least one match
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Indent -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_Indent'
```

### AC-UYP-005 — property set on the no-edit case (plan-audit D1 demoted byte-identity)

**Given** any section template where new == old == base
**When** merged through the node-tree encoder
**Then** the output preserves (a) comment count, (b) key order at every mapping level, (c) scalar quoting style, and (d) 2-space block-mapping indent. Byte-equality is recorded as a per-template boolean diagnostic, **not** the gate: only 8 of the 30 shipped section templates produce byte-identical no-edit output through the node encoder, and the 22 that differ (incl. `llm.yaml`, `workflow.yaml`, `lsp.yaml.tmpl`, `harness.yaml`, `quality.yaml.tmpl`, `delegation.yaml`) reindent and reflow to the canonical 2-space shape. A byte-equality failure must localize to one of (a)-(d); a byte-equality failure that does not is itself a finding.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_PropertySet -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestPreserveGolden_PropertySet'
```
Per-template `bytes_equal` diagnostic for the 22 non-identical files is asserted in the same test as a recorded field (not a pass/fail axis); the 8 byte-identical files (`cache.yaml` is one of them — see AC-UYP-021 caveat) ARE asserted as byte-equal to keep that diagnostic meaningful.

### AC-UYP-006 — user-added key survives 3-way merge

**Given** `newData = "shared: new_val\n"`, `oldData = "shared: new_val\nuser_added: custom\n"`, `baseData = "shared: base_val\n"`
**When** `MergeYAML3Way` runs
**Then** the output contains `user_added`.

```bash
# Go rewrites the subtest name "user added key not in new template is preserved" (spaces)
# to the underscored form below. Pin the normalized name so a typo or a future rename
# that breaks the matching fails the AC rather than passing vacuously (plan-audit D6).
# NB: an interim commit on main currently names the subtest "..._not_in_base_preserved";
# M4 renames it to "..._not_in_new_template_is_preserved" to match the SPEC's wording.
go test ./internal/cli/ -run 'TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved' -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved'
```

### AC-UYP-007 — retained key is reported

**Given** the AC-UYP-006 inputs
**When** the merge retains `user_added`
**Then** an advisory naming `user_added` is written to stderr.

```bash
go test ./internal/cli/update/backup/ -run TestMergeYAML3Way_ReportsRetainedKey -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestMergeYAML3Way_ReportsRetainedKey'
```

### AC-UYP-008 — 3-way is never more destructive than 2-way

**Given** any `(new, old)` pair
**When** both `MergeYAML3Way(new, old, base)` and `MergeYAMLDeep(new, old)` run
**Then** every top-level key present in the 2-way output is also present in the 3-way output.

```bash
go test ./internal/cli/update/backup/ -run TestMerge3WayNotMoreDestructiveThan2Way -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestMerge3WayNotMoreDestructiveThan2Way'
```

### AC-UYP-009 / AC-UYP-010 / AC-UYP-011 — decision semantics, system fields, error wrapping unchanged

**Given** the pre-existing decision, system-field, and invalid-input test bodies
**When** run against the node-tree implementation
**Then** all pass with their semantic assertions intact (whitespace-coupled substrings may be updated per plan M5; decision outcomes may not).

```bash
go test ./internal/cli/update/backup/ -run 'TestMergeYAML3Way|TestDeepMerge3Way|TestMergeYAMLDeep|TestDeepMergeMaps' -count=1 -v
go test ./internal/cli/ -run 'TestMergeYAML3Way|TestDeepMerge3Way|TestMergeYAMLDeep|TestDeepMergeMaps|TestValuesEqual' -count=1 -v
```

### AC-UYP-012 — anchors and aliases not expanded

**Given** a document containing `base: &a {k: v}` and `derived: *a`
**When** merged
**Then** the output still contains the literal `*a` alias and does not inline `{k: v}` a second time.

```bash
go test ./internal/cli/update/backup/ -run TestNodeMerge_AliasNotExpanded -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestNodeMerge_AliasNotExpanded'
```

### AC-UYP-013 — merge key treated as an ordinary key

**Given** a document containing `<<: *defaults`
**When** merged
**Then** the `<<` key appears in the output as an ordinary mapping key and its referent is not resolved into sibling keys.

```bash
go test ./internal/cli/update/backup/ -run TestNodeMerge_MergeKeyNotResolved -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestNodeMerge_MergeKeyNotResolved'
```

### AC-UYP-014 — sequences replaced wholesale

**Given** `new: [a, b]`, `old: [c]`, `base: [a, b]`
**When** merged
**Then** the output sequence is exactly `[c]` (user differs from base → user wins, whole), and never `[a, b, c]` or `[c, b]`.

```bash
go test ./internal/cli/update/backup/ -run TestNodeMerge_SequenceReplaced -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestNodeMerge_SequenceReplaced'
```

### AC-UYP-015 — multi-document input errors rather than silently truncating

**Given** input containing two documents separated by `---`
**When** passed to `MergeYAML3Way`
**Then** a non-nil error is returned, and its message names the multi-document condition.

```bash
go test ./internal/cli/update/backup/ -run TestMergeYAML3Way_MultiDocumentErrors -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestMergeYAML3Way_MultiDocumentErrors'
```

### AC-UYP-016 — explicit null is distinct from empty mapping

**Given** `k:` (null) on one side and `k: {}` on the other
**When** compared by the node equality primitive
**Then** they compare unequal.

```bash
go test ./internal/cli/update/backup/ -run TestNodeValuesEqual_NullVsEmptyMap -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestNodeValuesEqual_NullVsEmptyMap'
```

### AC-UYP-017 — golden test covers every section template

**Given** the 30 files under `internal/template/templates/.moai/config/sections/`
**When** the golden test runs
**Then** it reports one subtest per file, and the subtest count equals the file count.

```bash
ls internal/template/templates/.moai/config/sections/ | grep -cE '\.yaml(\.tmpl)?$'
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1 -v 2>&1 | grep -c '^    --- PASS'
# the two counts must match (currently 30 — re-derive at run time; do not hardcode)
```

### AC-UYP-018 — no test asserts a user key must be dropped

**Given** the full test tree
**When** searched for the retired expectation
**Then** zero matches.

```bash
matches=$(grep -rn 'expected user_added to be dropped' --include='*_test.go' . || true)
[ -z "$matches" ] && echo PASS || { echo "FAIL: $matches"; exit 1; }
```

### AC-UYP-019 — no fixture leaked into the template tree (plan-audit D13 simplified)

```bash
[ -z "$(git status --porcelain internal/template/templates/)" ] && echo PASS || { git status --porcelain internal/template/templates/; exit 1; }
grep -rn 'SPEC-UPDATE-YAML-PRESERVE\|REQ-UYP-' internal/template/templates/ && { echo FAIL; exit 1; } || echo PASS
```
The first command replaces the brittle `tee /dev/stderr` form (which mixed a diagnostic into the exit-status path); the bracketed emptiness test separates "show me the porcelain if non-empty" from the exit decision.

### AC-UYP-020 — package coverage ≥ plan-time baseline (plan-audit D11 captured the figure)

**Given** the plan-time baseline measured 2026-08-03 against the worktree tree
**When** coverage is re-measured after the change
**Then** the post-change percentage is greater than or equal to the baseline.

**Baseline (verbatim from `go test -cover ./internal/cli/update/backup/` on 2026-08-03):**
```
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.568s	coverage: 88.9% of statements
```
The post-change coverage MUST be ≥ **88.9%**. The figure is captured here in the SPEC body (not deferred to `progress.md` §E.2) so the AC cannot pass vacuously while the baseline is missing.

```bash
go test -cover ./internal/cli/update/backup/ 2>&1 | tee /dev/stderr | grep -E 'coverage: ([89][0-9]|[1-9][0-9]\.|\d{3})%'
# parse the percentage; assert ≥ 88.9
```

### AC-UYP-021 — end-to-end `moai update` preserves a customized file (plan-audit D12 added the DIFFER-class fixture)

**Given** a scratch project in `t.TempDir()` with TWO customized section files:
- `cache.yaml` — the shipped template with `enabled` flipped to `false` and a hand-added `user_note: "keep me"` key. `cache.yaml` is one of the 8 byte-identical templates, so it tests comment + value preservation in the easy case.
- `workflow.yaml` — the shipped template with one top-level value changed. `workflow.yaml` is the strongest single discriminator for the DIFFER class (no-edit round-trip measures 6797 → 6195 bytes, −602 B, the largest blank-line-collapse delta in the set). It tests that comment count survives even when byte-equality does not hold.

**When** the update restore path runs over the project
**Then** both resulting files preserve their full comment count (cache.yaml: 16; workflow.yaml: as-measured at run time), the user edits survive (`enabled: false`, `user_note: "keep me"` in cache.yaml; the edited value in workflow.yaml), `session_ttl: "1h"` is still quoted, and both files use 2-space indentation. Byte-equality is asserted for cache.yaml only; for workflow.yaml the property set (comment count + key order + quoting + indent) is asserted, NOT byte-equality.

```bash
go test ./internal/cli/ -run TestUpdateEndToEnd_PreservesCustomizedSection -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestUpdateEndToEnd_PreservesCustomizedSection'
```
The single-test form covers both fixtures via subtests; the existence grep prevents the previous vacuous-pass failure mode (D6).

### AC-UYP-022 — falsifiability: the golden test fails against the old implementation (plan-audit D5 repaired)

**Given** `merge.go` + `node_merge.go` temporarily reverted to the `map[string]any` round-trip
**When** the golden test runs
**Then** it FAILS.

The original command was non-executable: `git stash push` without `-u` rejects untracked pathspecs, and `node_merge.go` is a new untracked file at the moment this AC runs, so the stash aborted atomically — nothing was stashed, the operator saw the golden test still GREEN, and could mis-read that as "the check ran and the tree was fine". This is the load-bearing defect: AC-UYP-022 is the only mechanism in the SPEC that distinguishes a real preservation test from a vacuous one, and the original command did not run.

Repaired sequence (the `-u` is load-bearing; the intermediate-state assertion is load-bearing):

```bash
# 1. Stash the new implementation, INCLUDING the untracked node_merge.go (-u).
git stash push -u internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go

# 2. Intermediate-state assertion (plan-audit D5): the tree is now actually reverted.
#    merge.go MUST be restored to its pre-fix map-round-trip form, and node_merge.go
#    MUST be absent. If either check fails, the stash did not take and the RED test
#    below would run against the green tree — a false vacuous-pass.
git diff --exit-code internal/cli/update/backup/merge.go \
  && { echo "FAIL: merge.go unchanged after stash — stash did not take"; exit 1; }
test ! -e internal/cli/update/backup/node_merge.go \
  || { echo "FAIL: node_merge.go still present after stash"; exit 1; }

# 3. RED: the golden test MUST fail against the reverted (map round-trip) implementation.
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1
# expect: FAIL (non-zero exit). Capture verbatim output to progress.md §E.2.

# 4. Restore the new implementation.
git stash pop

# 5. GREEN: the golden test MUST pass against the node-tree implementation.
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1
# expect: PASS (zero exit). Capture verbatim output to progress.md §E.2.
```

Both observations (RED verbatim, GREEN verbatim) are recorded in `progress.md` §E.2. A golden test that passes in both states is vacuous and does not satisfy this AC. The intermediate-state assertion at step 2 is what makes the other preservation ACs non-vacuous: without it, a stash that aborts silently leaves the operator looking at a green tree and believing the RED step ran.

### AC-UYP-023 — `quality.yaml.tmpl` parses via node decoder (plan-audit D4)

**Given** the `quality.yaml.tmpl` template, whose unquoted placeholders `{{.EnforceQuality}}` / `{{.TestCoverageTarget}}` cause the **map** decoder (`map[string]any`) to fail with `yaml: invalid map key`
**When** `MergeYAML3Way` runs against the rendered-or-raw `quality.yaml` body using the **node** decoder
**Then** the merge returns a nil error (the placeholder lands as scalar-nested-flow-mapping text rather than failing), AND the file's transition from always-2-way (pre-fix, silent `restore.go:139` fallback) to real-3-way (post-fix) is observable.

This AC closes the D4 finding: the original SPEC never named `quality.yaml.tmpl`'s unparseability, and its fix-side effect (the file gets promoted from 2-way to 3-way) was entirely untested. Combined with the D5 deferral (whose wrong-base problem newly manifests for this file post-fix — see plan.md Decision D5 safety-argument point 4), this AC is what makes the `quality.yaml` path auditable rather than silent.

```bash
go test ./internal/cli/update/backup/ -run TestMergeYAML3Way_QualityTemplateParses -count=1 -v 2>&1 \
  | tee /dev/stderr | grep -q '^--- PASS: TestMergeYAML3Way_QualityTemplateParses'
```
The test asserts (a) `MergeYAML3Way(qualityBody, qualityBody, qualityBody)` returns `(non-nil output, nil error)`, and (b) the same body fed to a `map[string]any` decode in the same test (as a baseline) DOES error — so the test would fail if a future refactor regressed the node decoder to a map decoder.

## §D.2 Indirect Verification

- **AC-UYP-005 no longer subsumes AC-UYP-001 through AC-UYP-004** (plan-audit D1 corrected the original subsumption claim). The original argument was "byte identity implies comment, order, quoting, and indent preservation" — but byte identity holds for only 8 of 30 templates, so it cannot be the subsuming predicate. The corrected relationship: AC-UYP-005's **property set** (comment count + key order + quoting + 2-space indent) is the gate; AC-UYP-001 through AC-UYP-004 are the per-axis localizers that name which property broke when the gate fails. AC-UYP-005 records `bytes_equal` as a per-template diagnostic, asserted as byte-equality only for the 8 templates where it is achievable (cache.yaml among them).
- **AC-UYP-009/010/011** are verified by the survival of pre-existing tests rather than by new ones; that is deliberate, since the point is that nothing changed.
- **AC-UYP-023 indirectly verifies the permanent 2-way fallback contract** (REQ-UYP-011): by asserting the node decoder succeeds where the map decoder fails, it confirms the fix does not need to special-case `quality.yaml.tmpl` — the node decoder's success IS the fix, and the 2-way fallback remains as the safety net for any future unparseable template.

## §D.3 Closure Gate (Definition of Done)

- [ ] All MUST-severity ACs pass with verbatim command output recorded
- [ ] AC-UYP-022 falsifiability check shows RED-then-GREEN, both captured, with the intermediate-state assertion (step 2) observed
- [ ] `go test ./... -count=1` fully green
- [ ] `golangci-lint run ./internal/cli/...` clean
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` succeeds
- [ ] `internal/template/templates/` unmodified (AC-UYP-019)
- [ ] Package coverage ≥ 88.9% plan-time baseline (AC-UYP-020)
- [ ] Decision D5 deferral carried into a follow-up SPEC stub at sync time

## §D.4 Forward-Looking Checks (Post-Merge)

- Re-run AC-UYP-021 against a real user project after the next release to confirm the #1243 report is closed in practice.
- Confirm the D5 follow-up SPEC (`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`) is filed, since this SPEC's fix makes the stale-base symptom the remaining visible one — and (per plan.md Decision D5 safety-argument point 4) `quality.yaml` becomes a new victim of the wrong-base problem the moment this SPEC ships.
- After one release cycle, audit whether any user has reported `quality.yaml` key-flip symptoms (the D5-enlarged blast radius); if so, prioritize the D5 follow-up SPEC.

## §D.5 Quality Gate Criteria (TRUST 5)

- **Tested**: coverage at or above baseline; every requirement bound to a runnable command above.
- **Readable**: English godoc on every new exported symbol; the node-walk primitive documented with the yaml.v3 `Content` pairing convention it depends on.
- **Unified**: `gofmt` clean; `snake_case.go` filenames.
- **Secured**: no new external input surface — the merge reads files the update path already read. Error paths return wrapped errors and never panic.
- **Trackable**: single conventional commit referencing `#1243`.
