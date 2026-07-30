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
| AC-UYP-020 | §D constraint | M6 | MUST | coverage not regressed |
| AC-UYP-021 | §E success | M6 | MUST | end-to-end `moai update` preserves |
| AC-UYP-022 | plan M3 | M3 | MUST | falsifiability — test RED on reverted impl |

## §D.1 Given-When-Then Scenarios

### AC-UYP-001 — comments preserved on merge

**Given** `internal/template/templates/.moai/config/sections/cache.yaml` (16 comment-bearing lines, measured this session)
**When** it is passed to `MergeYAML3Way(src, src, src)`
**Then** the output contains the same number of `#`-bearing lines as the input.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Comments -count=1 -v
```

Falsification baseline (must FAIL before the fix — see AC-UYP-022):
```bash
# with the map round-trip restored, this same test reports comments in=16 out=0
```

### AC-UYP-002 — key order preserved

**Given** a mapping whose source key order is `enabled, session_ttl, spec_ttl, min_cacheable_tokens`
**When** merged with no user edit
**Then** the output key order is identical and is NOT the alphabetized `enabled, min_cacheable_tokens, session_ttl, spec_ttl`.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_KeyOrder -count=1 -v
```

### AC-UYP-003 — scalar quoting preserved

**Given** `session_ttl: "1h"` in the source
**When** the value is carried through unchanged
**Then** the output line is `session_ttl: "1h"`, not `session_ttl: 1h`.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Quoting -count=1 -v
```

### AC-UYP-004 — indentation pinned to 2

**Given** the encoder configuration in the merge implementation
**When** any nested mapping is emitted
**Then** each nesting level adds exactly 2 spaces.

```bash
grep -n 'SetIndent(2)' internal/cli/update/backup/*.go
# expect: at least one match
go test ./internal/cli/update/backup/ -run TestPreserveGolden_Indent -count=1 -v
```

### AC-UYP-005 — byte identity on the no-edit case

**Given** any section template where new == old == base
**When** merged
**Then** the output bytes equal the input bytes exactly.

```bash
go test ./internal/cli/update/backup/ -run TestPreserveGolden_ByteIdentity -count=1 -v
```

### AC-UYP-006 — user-added key survives 3-way merge

**Given** `newData = "shared: new_val\n"`, `oldData = "shared: new_val\nuser_added: custom\n"`, `baseData = "shared: base_val\n"`
**When** `MergeYAML3Way` runs
**Then** the output contains `user_added`.

```bash
go test ./internal/cli/ -run 'TestMergeYAML3Way/user_added_key_not_in_new_template_is_preserved' -count=1 -v
```

### AC-UYP-007 — retained key is reported

**Given** the AC-UYP-006 inputs
**When** the merge retains `user_added`
**Then** an advisory naming `user_added` is written to stderr.

```bash
go test ./internal/cli/update/backup/ -run TestMergeYAML3Way_ReportsRetainedKey -count=1 -v
```

### AC-UYP-008 — 3-way is never more destructive than 2-way

**Given** any `(new, old)` pair
**When** both `MergeYAML3Way(new, old, base)` and `MergeYAMLDeep(new, old)` run
**Then** every top-level key present in the 2-way output is also present in the 3-way output.

```bash
go test ./internal/cli/update/backup/ -run TestMerge3WayNotMoreDestructiveThan2Way -count=1 -v
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
go test ./internal/cli/update/backup/ -run TestNodeMerge_AliasNotExpanded -count=1 -v
```

### AC-UYP-013 — merge key treated as an ordinary key

**Given** a document containing `<<: *defaults`
**When** merged
**Then** the `<<` key appears in the output as an ordinary mapping key and its referent is not resolved into sibling keys.

```bash
go test ./internal/cli/update/backup/ -run TestNodeMerge_MergeKeyNotResolved -count=1 -v
```

### AC-UYP-014 — sequences replaced wholesale

**Given** `new: [a, b]`, `old: [c]`, `base: [a, b]`
**When** merged
**Then** the output sequence is exactly `[c]` (user differs from base → user wins, whole), and never `[a, b, c]` or `[c, b]`.

```bash
go test ./internal/cli/update/backup/ -run TestNodeMerge_SequenceReplaced -count=1 -v
```

### AC-UYP-015 — multi-document input errors rather than silently truncating

**Given** input containing two documents separated by `---`
**When** passed to `MergeYAML3Way`
**Then** a non-nil error is returned, and its message names the multi-document condition.

```bash
go test ./internal/cli/update/backup/ -run TestMergeYAML3Way_MultiDocumentErrors -count=1 -v
```

### AC-UYP-016 — explicit null is distinct from empty mapping

**Given** `k:` (null) on one side and `k: {}` on the other
**When** compared by the node equality primitive
**Then** they compare unequal.

```bash
go test ./internal/cli/update/backup/ -run TestNodeValuesEqual_NullVsEmptyMap -count=1 -v
```

### AC-UYP-017 — golden test covers every section template

**Given** the 31 files under `internal/template/templates/.moai/config/sections/`
**When** the golden test runs
**Then** it reports one subtest per file, and the subtest count equals the file count.

```bash
ls internal/template/templates/.moai/config/sections/ | grep -cE '\.yaml(\.tmpl)?$'
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1 -v 2>&1 | grep -c '^    --- PASS'
# the two counts must match
```

### AC-UYP-018 — no test asserts a user key must be dropped

**Given** the full test tree
**When** searched for the retired expectation
**Then** zero matches.

```bash
matches=$(grep -rn 'expected user_added to be dropped' --include='*_test.go' . || true)
[ -z "$matches" ] && echo PASS || { echo "FAIL: $matches"; exit 1; }
```

### AC-UYP-019 — no fixture leaked into the template tree

```bash
git status --porcelain internal/template/templates/ | tee /dev/stderr | grep -q . && { echo "FAIL: template tree modified"; exit 1; } || echo PASS
grep -rn 'SPEC-UPDATE-YAML-PRESERVE\|REQ-UYP-' internal/template/templates/ && { echo FAIL; exit 1; } || echo PASS
```

### AC-UYP-020 — package coverage not regressed

**Given** the pre-flight baseline captured in plan §C step 3
**When** coverage is re-measured after the change
**Then** the post-change percentage is greater than or equal to the baseline.

```bash
go test -cover ./internal/cli/update/backup/
# compare against the verbatim pre-flight figure recorded in progress.md §E.2
```

### AC-UYP-021 — end-to-end `moai update` preserves a customized file

**Given** a scratch project in `t.TempDir()` whose `.moai/config/sections/cache.yaml` is the shipped template with `enabled` flipped to `false` and a hand-added `user_note: "keep me"` key
**When** the update restore path runs over it
**Then** the resulting file contains all 16 comments, `enabled: false`, `user_note: "keep me"`, `session_ttl: "1h"` still quoted, and 2-space indentation.

```bash
go test ./internal/cli/ -run TestUpdateEndToEnd_PreservesCustomizedSection -count=1 -v
```

### AC-UYP-022 — falsifiability: the golden test fails against the old implementation

**Given** `merge.go` temporarily reverted to the `map[string]any` round-trip
**When** the golden test runs
**Then** it FAILS.

```bash
git stash push internal/cli/update/backup/merge.go internal/cli/update/backup/node_merge.go
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1   # expect FAIL
git stash pop
go test ./internal/cli/update/backup/ -run TestPreserveGolden -count=1   # expect PASS
```

Both observations are recorded verbatim in `progress.md` §E.2. A golden test that passes in both states is vacuous and does not satisfy this AC.

## §D.2 Indirect Verification

- **AC-UYP-005 subsumes AC-UYP-001 through AC-UYP-004** on the no-edit path: byte identity implies comment, order, quoting, and indent preservation. The four narrower ACs are retained because they localize a failure — a byte-identity failure alone does not say which property broke.
- **AC-UYP-009/010/011** are verified by the survival of pre-existing tests rather than by new ones; that is deliberate, since the point is that nothing changed.

## §D.3 Closure Gate (Definition of Done)

- [ ] All MUST-severity ACs pass with verbatim command output recorded
- [ ] AC-UYP-022 falsifiability check shows RED-then-GREEN, both captured
- [ ] `go test ./... -count=1` fully green
- [ ] `golangci-lint run ./internal/cli/...` clean
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` succeeds
- [ ] `internal/template/templates/` unmodified (AC-UYP-019)
- [ ] Package coverage ≥ pre-flight baseline (AC-UYP-020)
- [ ] Decision D5 deferral carried into a follow-up SPEC stub at sync time

## §D.4 Forward-Looking Checks (Post-Merge)

- Re-run AC-UYP-021 against a real user project after the next release to confirm the #1243 report is closed in practice.
- Confirm the D5 follow-up SPEC (`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`) is filed, since this SPEC's fix makes the stale-base symptom the remaining visible one.

## §D.5 Quality Gate Criteria (TRUST 5)

- **Tested**: coverage at or above baseline; every requirement bound to a runnable command above.
- **Readable**: English godoc on every new exported symbol; the node-walk primitive documented with the yaml.v3 `Content` pairing convention it depends on.
- **Unified**: `gofmt` clean; `snake_case.go` filenames.
- **Secured**: no new external input surface — the merge reads files the update path already read. Error paths return wrapped errors and never panic.
- **Trackable**: single conventional commit referencing `#1243`.
