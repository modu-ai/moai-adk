# Progress — SPEC-UPDATE-MERGE-CONFLICT-BLIND-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M — `spec.md` + `plan.md` + `acceptance.md` (+ this `progress.md`).
- SPEC ID regex check executed as Bash; observed output: `PASS`.
- ID uniqueness: `ls .moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` reported
  `No such file or directory` before authoring — no collision.
- Frontmatter: canonical 12 fields present in `spec.md`, `status: draft`.
- Authoritative input: `.moai/reports/t576/verdict.md` (read in full).
- Not executed at plan-phase: no test, no build, no production-code change.
  `spec.md` §A is a hypothesis until M1 lands (`spec.md` §A.5).

## §E.2 Run-phase Evidence

### M1 — the reproduction (no production code change)

- **PKG** (fixed by M1, per `acceptance.md` preamble): `./internal/cli/update/merge/...`.
  The base derivation the measurement must exercise (`deriveTemplateBase` /
  `derive` / `pruneToShared`) is unexported in that package, so the harness lands
  there and reaches the engine through `internal/merge`'s exported `MergeFile`.
- **Harness**: `internal/cli/update/merge/conflict_blind_repro_test.go` (new file;
  no production file modified).
- **Baseline attribution**: worktree `.claude/worktrees/t576`, branch
  `WT-permissions-ask-empty`, tree `9cf415332` plus the one new untracked test
  file above. All readings below were observed in this run, against this tree.

#### The four control cells — one run, one fixture

Command:

```
go test ./internal/cli/update/merge/... -run 'TestSharedKeyControlCells' -v -count=1
```

Observed (verbatim):

```
=== RUN   TestSharedKeyControlCells
=== RUN   TestSharedKeyControlCells/untouched_shared
    conflict_blind_repro_test.go:165: cell untouched_shared: key=untouched_shared written=["template-old"] HasConflict=false len(Conflicts)=0 predicted=["template-old"] (shared key: derived base carries the template's value, so the template's change reads as no change and the user's value stands)
=== RUN   TestSharedKeyControlCells/emptied_shared
    conflict_blind_repro_test.go:165: cell emptied_shared: key=emptied_shared written=[] HasConflict=false len(Conflicts)=0 predicted=[] (shared key emptied by the user: the empty array is the user's change and is preserved)
=== RUN   TestSharedKeyControlCells/changed_shared
    conflict_blind_repro_test.go:165: cell changed_shared: key=changed_shared written=["user-choice"] HasConflict=false len(Conflicts)=0 predicted=["user-choice"] (shared key changed by the user: the user's value is preserved)
=== RUN   TestSharedKeyControlCells/omitted
    conflict_blind_repro_test.go:165: cell omitted: key=omitted_key written=["template-only"] HasConflict=false len(Conflicts)=0 predicted=["template-only"] (key absent from the user's side is not shared, so it stays out of the base and the template reads as introducing it)
--- PASS: TestSharedKeyControlCells (0.00s)
    --- PASS: TestSharedKeyControlCells/untouched_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/emptied_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/changed_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/omitted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.645s
```

| Cell | Prediction (`plan.md` §F) | Value written | `HasConflict` | `len(Conflicts)` | Held? |
|---|---|---|---|---|---|
| (i) `untouched_shared` | user's value stands | `["template-old"]` | `false` | `0` | yes |
| (ii) `emptied_shared` | `[]` preserved | `[]` | `false` | `0` | yes |
| (iii) `changed_shared` | user's value preserved | `["user-choice"]` | `false` | `0` | yes |
| (iv) `omitted` — discriminator | template's value lands | `["template-only"]` | `false` | `0` | yes |

Cell (iv) landing the template's value is what separates a merge that ran from a
file that was left alone: a harness that copied the user's document would have
written nothing for `omitted_key`, and the sub-test fails on an absent key rather
than reporting a pass. All four sub-tests appear in the single run — no cell is
absent, so no cell is a gap.

#### The conflict surface — measured, and measured against a control

Command:

```
go test ./internal/cli/update/merge/... -run 'TestConflictSurfaceReachability' -v -count=1
```

Observed (verbatim):

```
=== RUN   TestConflictSurfaceReachability
=== RUN   TestConflictSurfaceReachability/derived_base
    conflict_blind_repro_test.go:196: divergent shared key, base derived as the update path derives it: written="user-value" HasConflict=false len(Conflicts)=0 (prediction: false, 0)
=== RUN   TestConflictSurfaceReachability/genuine_base_control
    conflict_blind_repro_test.go:226: same key and same engine, base differing from both sides: HasConflict=true len(Conflicts)=1 (prediction: true, 1)
--- PASS: TestConflictSurfaceReachability (0.00s)
    --- PASS: TestConflictSurfaceReachability/derived_base (0.00s)
    --- PASS: TestConflictSurfaceReachability/genuine_base_control (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.543s
```

Both readings come from `MergeResult` as the engine returned it. The base is not
re-derived and compared to the value it was derived from, which would assert
nothing.

The control is the load-bearing half. "No conflict was reported" and "this
harness never reads a conflict" produce the same two readings, so a bare `false`
/ `0` would be uninterpreted output. The control runs the **same engine over the
same key** with a base that agrees with neither side, and the surface fires:
`HasConflict=true`, `len(Conflicts)=1`. The `false` / `0` above is therefore a
property of the derived base, not of the instrument.

#### What M1 establishes, and what it does not

- **Established.** `spec.md` §A.4's healing boundary is real and measured: an
  omitted key heals (cell iv), an emptied one does not (cell ii). The
  shared-key conflict surface is unreachable through the derived base while
  being reachable through a base that can differ from `updated` — the two
  sub-tests above differ in exactly that input. `spec.md` §A.5's hypothesis
  marker is discharged for §A.1-§A.4 at the shared-key level.
- **Not established.** Nothing here measures `permissions.ask` in any real
  checkout, nothing measures `moai update` end to end, and nothing measures
  nested-key or YAML behaviour (see Gaps in the M1 report). No production code
  changed, so no `REQ-UMC-008`/`009`/`011` obligation is discharged; M2's design
  choice stays open on this evidence.

#### Scoped verification

```
$ go test ./internal/merge/... ./internal/cli/update/merge/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/merge	0.395s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.371s
exit 0

$ go vet ./internal/merge/... ./internal/cli/update/merge/...
(no output)  exit 0

$ golangci-lint run --timeout=5m ./internal/merge/... ./internal/cli/update/merge/...
0 issues.
exit 0
```

Neither package reported `[no test files]`. The full-suite verdict is CI's, not
this run's (`CLAUDE.local.md` §4).

#### AC matrix — M1

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UMC-001 | PASS | `grep -n 't\.TempDir()' <file>` / `grep -nE '/Users/\|\.claude/settings\.json"\|/moai/moai-adk-go' <file>` | `37:	dir := t.TempDir()` / no match (rc 1) |
| AC-UMC-002 | PASS | `grep -nE 'exec\.Command\|RunUpdate\|runUpdate' <file>` | no match (rc 1) |
| AC-UMC-003 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/untouched_shared' -v -count=1` | `--- PASS`; wrote `["template-old"]` |
| AC-UMC-004 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/emptied_shared' -v -count=1` | `--- PASS`; wrote `[]` (not the template's non-empty value) |
| AC-UMC-005 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/changed_shared' -v -count=1` | `--- PASS`; wrote `["user-choice"]` |
| AC-UMC-006 (MUST) | PASS | `go test $PKG -run 'TestSharedKeyControlCells/omitted' -v -count=1` | `--- PASS`; template's value `["template-only"]` landed |
| AC-UMC-007 (MUST) | PASS | `go test $PKG -run 'TestSharedKeyControlCells' -v -count=1` | exactly 4 sub-tests, each with a `--- PASS` line |
| AC-UMC-008 (MUST) | PASS | `go test $PKG -run 'TestConflictSurfaceReachability' -v -count=1` | `HasConflict=false len(Conflicts)=0` on the derived base; control `true` / `1` |
| AC-UMC-015 (MUST) | PASS | `git status --porcelain -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` | no output (both untouched; `REQ-UMC-014` not reached) |
| AC-UMC-016 | N/A | `git diff --name-only` | no settings document changed in M1 — recorded N/A, not PASS |
| AC-UMC-017 (MUST) | PASS | `go test ./internal/merge/... ./internal/cli/update/merge/... -count=1` | two `ok` lines, exit 0, no `[no test files]` |
| AC-UMC-009 / 010 / 011 | not attempted | — | M2 scope; not run in this milestone |
| AC-UMC-012 / 013 / 014 | not attempted | — | M3 scope; not run in this milestone |

### M2.0 — the precondition measurement (no production code change)

`plan.md` §F M2.0 makes two breadths the precondition of any repair: the
recursive `pruneToShared` path (`base.go:124-126`) and the YAML path
(`base.go:75-78`). M1 entered neither, so `spec.md` §A.6 stood for top-level
JSON keys and for nothing else (`spec.md` §A.3). Both are now measured, each on
its own fixture, its own codec, and its own control. No production code changed;
the milestone's only artifact is
`internal/cli/update/merge/conflict_blind_breadth_test.go`.

#### Does the YAML path reach `deepMergeMap`? — read from the source

It does, and it is the SAME function the JSON path uses. Read at tree
`a79e8601c`:

- `internal/merge/strategies.go:79-80` — the strategy selector maps `.yaml` /
  `.yml` to `YAMLDeep`; `:81-82` maps `.json` to `JSONMerge`.
- `internal/merge/three_way.go:56-59` — `MergeFile` dispatches `YAMLDeep` to
  `mergeYAML` and `JSONMerge` to `mergeJSON`.
- `internal/merge/strategies.go:343` — `mergeYAML` calls
  `deepMergeMap(baseMap, currentMap, updatedMap, "")`;
  `internal/merge/strategies.go:314` — `mergeJSON` calls the same function.

So no separate YAML merge strategy exists: the two paths differ only in the
codec that produces the three maps (`yaml.Unmarshal` vs `json.Unmarshal`) and in
the codec that re-serializes the result. The arm structure of
`strategies.go:418-462` is shared. `deriveTemplateBase` likewise dispatches YAML
through `yaml.Unmarshal` / `yaml.Marshal` (`base.go:75-78`) into the same
`pruneToShared`.

This is a source reading, not a substitute for the measurement below — and the
measurement asserts `MergeResult.Strategy == YAMLDeep` on every YAML cell, so a
YAML reading cannot be a JSON reading taken through a `.yaml` file name.

#### Breadth (A) — the recursive path, JSON codec

Command (`TestBreadthRecursiveJSONCells`, container shared on both sides, four
cells one level down):

```
$ go test ./internal/cli/update/merge/ -run 'TestBreadthRecursiveJSONCells' -count=1 -v
```

Derived base, printed by the run (the recursion narrowed the container, dropping
`omitted_leaf` and `user_only`):

```
{
  "container": {
    "changed_shared": ["template-choice"],
    "emptied_shared": ["template-a", "template-b"],
    "untouched_shared": ["template-new"]
  }
}
```

| Cell (nested leaf) | Value written | `HasConflict` | `len(Conflicts)` | Prediction | Held? |
|---|---|---|---|---|---|
| (i) `untouched_shared` | `["template-old"]` | `false` | `0` | `["template-old"]` | yes |
| (ii) `emptied_shared` | `[]` | `false` | `0` | `[]` | yes |
| (iii) `changed_shared` | `["user-choice"]` | `false` | `0` | `["user-choice"]` | yes |
| (iv) `omitted_leaf` **(discriminator)** | `["template-only"]` | `false` | `0` | `["template-only"]` | yes |
| `user_only` | `["user-addition"]` | `false` | `0` | `["user-addition"]` | yes |

Cell (iv) is the case the prompt names as "container shared, leaf absent on the
user's side": the leaf's template value **does** land. It is also the
discriminator — it is the only cell of the five that ends with the template's
value, so a harness that copied the user's container would fail it.

Second additional case — the container itself absent from the user's side
(`TestBreadthRecursiveJSONOmittedContainer`):

```
$ go test ./internal/cli/update/merge/ -run 'TestBreadthRecursiveJSONOmittedContainer' -count=1 -v
    recursive-json cell omitted_container: written={"leaf_a":["template-a"],"leaf_b":["template-b"]} HasConflict=false len(Conflicts)=0 strategy=json_merge predicted={"leaf_a":["template-a"],"leaf_b":["template-b"]}
--- PASS: TestBreadthRecursiveJSONOmittedContainer (0.00s)
```

The whole container lands with its contents. Prediction held.

Conflict surface one level down (`TestBreadthRecursiveJSONConflictSurface`), a
nested shared leaf on which the two sides genuinely disagree, paired with a
control on the same engine and the same nested key:

| Sub-test | Base | `HasConflict` | `len(Conflicts)` | Prediction | Held? |
|---|---|---|---|---|---|
| `derived_base` | derived as the update path derives it | `false` | `0` | `false` / `0` | yes |
| `genuine_base_control` | differs from both sides at the leaf | `true` | `1` | `true` / `1` | yes |

The control fires, so the `false` / `0` reading is a property of the derived
base rather than of an instrument that never reads a conflict.

#### Breadth (B) — the YAML path

Flat four cells through the YAML strategy (`TestBreadthYAMLFlatCells`;
`MergeResult.Strategy == yaml_deep` asserted, not assumed):

```
$ go test ./internal/cli/update/merge/ -run 'TestBreadthYAMLFlatCells' -count=1 -v
```

| Cell | Value written | `HasConflict` | `len(Conflicts)` | Strategy | Prediction | Held? |
|---|---|---|---|---|---|---|
| (i) `untouched_shared` | `["template-old"]` | `false` | `0` | `yaml_deep` | `["template-old"]` | yes |
| (ii) `emptied_shared` | `[]` | `false` | `0` | `yaml_deep` | `[]` | yes |
| (iii) `changed_shared` | `["user-choice"]` | `false` | `0` | `yaml_deep` | `["user-choice"]` | yes |
| (iv) `omitted_key` **(discriminator)** | `["template-only"]` | `false` | `0` | `yaml_deep` | `["template-only"]` | yes |
| `user_only` | `["user-addition"]` | `false` | `0` | `yaml_deep` | `["user-addition"]` | yes |

The two breadths intersect in the shape `.moai/config/sections/*.yaml` actually
has — a nested leaf reached through the YAML codec — so that intersection is
measured too rather than inferred from either breadth
(`TestBreadthYAMLNestedCells`):

| Cell (nested leaf, YAML) | Value written | `HasConflict` | `len(Conflicts)` | Strategy | Prediction | Held? |
|---|---|---|---|---|---|---|
| (i) `untouched_shared` | `["template-old"]` | `false` | `0` | `yaml_deep` | `["template-old"]` | yes |
| (ii) `emptied_shared` | `[]` | `false` | `0` | `yaml_deep` | `[]` | yes |
| (iii) `changed_shared` | `["user-choice"]` | `false` | `0` | `yaml_deep` | `["user-choice"]` | yes |
| (iv) `omitted_leaf` **(discriminator)** | `["template-only"]` | `false` | `0` | `yaml_deep` | `["template-only"]` | yes |
| `user_only` | `["user-addition"]` | `false` | `0` | `yaml_deep` | `["user-addition"]` | yes |

YAML conflict surface, with its own control on its own codec — not inherited
from breadth (A), because a control inside one codec says nothing about the
other (`TestBreadthYAMLConflictSurface`):

| Sub-test | Base | `HasConflict` | `len(Conflicts)` | Strategy | Prediction | Held? |
|---|---|---|---|---|---|---|
| `derived_base` | derived as the update path derives it | `false` | `0` | `yaml_deep` | `false` / `0` | yes |
| `genuine_base_control` | differs from both sides | `true` | `1` | `yaml_deep` | `true` / `1` | yes |

#### What M2.0 establishes, and the one place §A.6 needs reading with care

- **Established.** `spec.md` §A.6's mechanism extends to **both** breadths **at
  leaf granularity**: for a shared scalar-or-sequence leaf — top-level or
  nested, JSON or YAML — the merge writes the user's value and the conflict
  surface reads `false` / `0`, while the same engine on the same key fires
  `true` / `1` under a base that can differ from `updated`. §A.4's healing
  boundary extends identically: an omitted key or container heals; an emptied
  one does not.
- **A granularity qualification, measured in the same runs.** For a shared
  **container** key, the value the merge writes is *not* the user's value. In
  both nested runs the container `container` was carried by both sides, and the
  written container reads
  `{"changed_shared":["user-choice"],"emptied_shared":[],"omitted_leaf":["template-only"],"untouched_shared":["template-old"],"user_only":["user-addition"]}`
  — the user's leaves plus the template's new leaf. So §A.6's sentence "the
  merge cannot change the VALUE of a key the user's file already carries" is
  true of a shared **leaf** and false of a shared **container**: a container's
  value changes whenever the template adds a leaf inside it. This is the
  mechanism working as `base.go:105-111` describes, not a defect, and it is why
  the finding should be stated at leaf granularity. Which arm of
  `strategies.go:422-462` fired for the container is **not** observable from
  `MergeResult` and is not claimed here.
- **Not established.** No end-to-end `moai update` run; no real checkout; no
  `permissions.ask` measurement anywhere; no non-map/non-scalar shape beyond
  arrays of strings; no cross-platform (`GOOS=windows`) reading; no coverage
  delta. No production code changed, so no `REQ-UMC-008` / `009` / `011`
  obligation is discharged and M2.1's design choice stays open — deliberately,
  since M2.0 was authorized and M2's repair was not.

#### Scoped verification — M2.0

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/update/merge/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.371s
exit 0

$ unset ... && go vet ./internal/cli/update/merge/
(no output)  exit 0

$ unset ... && golangci-lint run ./internal/cli/update/merge/...
0 issues.
exit 0

$ git status --short
 M .claude/settings.json                                        # operator's, untouched by this run
?? internal/cli/update/merge/conflict_blind_breadth_test.go     # the only file this run authored
```

The package did not report `[no test files]`, and the M1 tests in the same
package still pass in the same run. The full-suite verdict is CI's, not this
run's (`CLAUDE.local.md` §4).

#### AC matrix — M2.0

M2.0 is a `plan.md` §F precondition measurement, not an AC-bearing milestone:
`acceptance.md` maps `AC-UMC-009` / `010` / `011` to the M2.1 repair, which is
unauthorized here. Only the hygiene criteria that bind every run-phase milestone
are re-measured.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UMC-001 | PASS | `grep -n 't\.TempDir()' internal/cli/update/merge/conflict_blind_breadth_test.go` / `grep -nE '/Users/\|\.claude/settings\.json"\|/moai/moai-adk-go' <same file>` | `50:	dir := t.TempDir()` / no match (rc 1) |
| AC-UMC-002 | PASS | `grep -nE 'exec\.Command\|RunUpdate\|runUpdate' <same file>` | no match (rc 1) |
| AC-UMC-015 (MUST) | PASS | `git status --porcelain -- internal/template/templates/.claude/settings.json.tmpl` | no output (`REQ-UMC-014` not reached) |
| AC-UMC-017 (MUST) | PASS | `go test ./internal/cli/update/merge/ -count=1` | one `ok` line, exit 0, no `[no test files]` |
| AC-UMC-009 / 010 / 011 | not attempted | — | M2.1 scope; the repair is unauthorized in this run |
| AC-UMC-012 / 013 / 014 | not attempted | — | M3 scope; not run in this milestone |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: m2.0-complete          # NOT run-complete — M2.1 and M3 are unstarted
m1_complete_at: 2026-09-10
m1_commit_sha: pending-backfill    # M1 files authored; commit is the lead's
m2_0_complete_at: 2026-09-10
m2_0_commit_sha: pending-backfill  # M2.0 file authored; commit is the lead's
m2_0_measured_at_head: a79e8601c
pkg_selected: ./internal/cli/update/merge/...
ac_pass_count: 10                  # AC-UMC-001..008, 015, 017 (M1; M2.0 re-measured 001/002/015/017, no new AC)
ac_fail_count: 0
ac_na_count: 1                     # AC-UMC-016 (no settings document changed)
ac_not_attempted_count: 6          # AC-UMC-009..014 (M2.1/M3 scope)
production_files_changed: 0
new_test_files: 2                  # M1 repro + M2.0 breadth
new_warnings_or_lints_introduced: 0
preserve_list_post_run_count: 0    # no PRESERVE-listed file modified
breadth_recursive_path: measured   # plan.md §F M2.0 (1) — 6 cells + control, JSON codec
breadth_yaml_path: measured        # plan.md §F M2.0 (2) — 10 cells + control, yaml_deep asserted
breadth_predictions_held: 20       # 8 in breadth (A), 12 in breadth (B) — every cell and both control pairs
breadth_predictions_failed: 0
run_complete_at: pending           # awaits M2.1 + M3
```

Not measured, and therefore gaps rather than passes: cross-platform build
(`GOOS=windows`), coverage delta, `permissions.ask` in any real checkout, any
end-to-end `moai update` path, value shapes beyond scalars / string arrays /
nested maps, and which arm of `strategies.go:422-462` fires for a shared
container key (not observable from `MergeResult`).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
