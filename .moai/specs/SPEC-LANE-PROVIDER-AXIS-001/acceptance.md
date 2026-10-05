# Acceptance Criteria — SPEC-LANE-PROVIDER-AXIS-001

8 acceptance criteria across three milestones. Every criterion is mechanically checkable, and every criterion states what it would fail on.

Nothing in this SPEC has landed, so every criterion below is unmet at authoring — that is the expected starting state, not a defect.

---

## §D.0 The vacuity rule this file is written against

[HARD] A criterion that an unchanged tree already satisfies verifies nothing. Two shapes of criterion appear below, and they discharge that rule differently:

- **Additive criteria** (AC-LPA-001, 003, 004, 005, 006) assert something the current tree does not contain, so they fail today by construction. The **Fails on** line names what today's tree lacks.
- **Guard criteria** (AC-LPA-002, 007, 008) assert a property the current tree already has — a regression guard passes on a clean tree, always. Their evidence is therefore **not** the green run. It is a named mutation that makes the check fail, reverted immediately after being observed. A guard criterion whose mutation was not run is a Gap, not a pass: without it, a check that cannot fail for any reason is indistinguishable from one that holds.

---

## §D AC Matrix

### §D.1 M1 — The lane pool axis

**AC-LPA-001** (REQ-LPA-001)
- **Given** a factory registry with several registered lanes, some resolvable and at least one not
- **When** the lane pool aggregate is computed over the built lane rows
- **Then** it reports a count per provider present plus a count of lanes whose provider is unrecorded
- **Verify**: Go test in `internal/web/factory_lanes_test.go` constructing a fixture pool and asserting the aggregate's contents field by field
- **Fails on**: today's tree, where no aggregate exists — the test does not compile

**AC-LPA-002** (REQ-LPA-002) — guard criterion
- **Given** `internal/web/factory_lanes.go` and its test file
- **When** they are scanned for a reference to the profile axis or the agent axis
- **Then** there is no hit
- **Verify**:
  ```bash
  git grep -nE 'profile_matrix|ProfileMatrix|profileMatrix|ModelEffort|tierProfiles' -- internal/web/factory_lanes.go internal/web/factory_lanes_test.go
  # expect: no output
  git grep -cE 'Lane' -- internal/web/factory_lanes.go
  # positive control: non-zero — the instrument fires on this file
  ```
- **Mutation that must make it fail**: add a single reference to `template.ProfileMatrix` (or any identifier in the pattern set) inside `factory_lanes.go`; the scan must then report that line. Revert.
- **Fails on**: a tree where the pool axis was built by reading the profile matrix — which is the exact scope crossing `spec.md` §C forbids

### §D.2 M2 — The provider spread display

**AC-LPA-003** (REQ-LPA-003)
- **Given** a rendered `Factory lanes` panel over a non-empty lane pool
- **When** the panel head's markup is read
- **Then** it carries the provider spread beside the existing lane count
- **Verify**: Go test in `internal/web/factory_lane_section_test.go` rendering the section and asserting the spread markup is present, in the same style as the existing `TestLaneSectionRendersCompleteRow`
- **Fails on**: today's tree, whose panel head carries only `len(k.Lanes)`

**AC-LPA-004** (REQ-LPA-004)
- **Given** a lane pool of N lanes in which at least one carries an empty provider
- **When** the spread's buckets are summed
- **Then** the sum equals N, the empty-provider lanes appear in a distinct unrecorded bucket, and no named provider bucket contains them
- **Verify**: Go test asserting both the partition (sum equals the lane count) and the placement (the unrecorded bucket's count equals the number of empty-provider fixtures); table-driven over at least the cases N=1 all-unrecorded, N=3 mixed, and N=2 all-recorded
- **Fails on**: today's tree (no buckets), and on any implementation that lets an empty provider fall into a named bucket — the partition assertion still passes there, which is why the placement assertion is separate

**AC-LPA-005** (REQ-LPA-005)
- **Given** the rendered spread over a lane pool containing only reachable providers
- **When** the set of bucket labels in the output is enumerated
- **Then** it contains no bucket for a provider absent from the pool, and specifically no `gpt` bucket
- **Verify**:
  ```bash
  # after the render test writes its output fixture
  grep -c 'gpt' <rendered-spread-output>
  # expect: 0
  git grep -n 'BackendGPT' -- internal/web/
  # expect: no output — the bucket set is not derived from the constant block
  git grep -c 'BackendGLM' -- internal/web/factory_lanes_test.go
  # positive control: non-zero — the instrument fires on this tree
  ```
  plus a Go test asserting the bucket-label set equals exactly the set of providers present on the fixture rows, plus the unrecorded bucket when any lane carries an empty provider
- **Fails on**: an implementation that switches over `kanban.Backend*`, which yields a `gpt` bucket the fixture pool cannot fill

**AC-LPA-006** (REQ-LPA-006)
- **Given** every i18n key the spread introduces
- **When** each of the four locale blocks of `internal/web/assets/i18n.js` is looked up for that key
- **Then** all four resolve, with no key present in a subset
- **Verify**: extend `TestSPECIntroducedKeysResolveInEveryLocale` in `internal/web/factory_lane_section_test.go` with the new key set
- **Fails on**: a tree where the keys were added to `en` only — the existing test already demonstrates this shape of failure for the keys it covers

### §D.3 M3 — The regression guard

**AC-LPA-007** (REQ-LPA-007, shape direction) — guard criterion
- **Given** the guard, on an unmodified tree
- **When** `go test -count=1 -run '<guard>' ./internal/web/...` runs
- **Then** it passes, and the field is asserted to be a single scalar string
- **Mutation that must make it fail**: the **coherent** widening — change `LaneVM.Backend` from `string` to `[]string` **and**, in the same edit, change the assignment at `internal/web/factory_lanes.go:150` to construct a slice, so the package still builds. The guard must then fail. Revert.
- [HARD] The incoherent half-edit — widening the field and leaving the assignment — is a **compile error**, and a build failure is not the guard firing. Citing it as the mutation evidence proves nothing about the guard and is the specific failure mode `plan.md` §F R-1 names.
- **Fails on**: a tree whose guard asserts something the compiler already enforces; under the coherent mutation such a guard passes while the invariant is broken

**AC-LPA-008** (REQ-LPA-007, count direction) — guard criterion
- **Given** the guard, on an unmodified tree
- **When** the same guard run is executed
- **Then** it passes, and the number of assignment sites to the lane row's provider field within `internal/web/factory_lanes.go` is asserted to be exactly one
- **Mutation that must make it fail**: add a second assignment to that field in the file — for instance a fallback `row.Backend = kanban.BackendClaude` on an unresolved row. Nothing about the type changes and the package still builds, so **no compiler signal exists**; the guard must be what fails. Revert.
- **Fails on**: a tree with only the shape assertion of AC-LPA-007, which the second-assignment mutation passes untouched — that is why this criterion is separate rather than folded into AC-LPA-007

---

## §D.4 Traceability

| REQ | Criteria | Kind |
|---|---|---|
| REQ-LPA-001 | AC-LPA-001 | additive |
| REQ-LPA-002 | AC-LPA-002 | guard (1 mutation) |
| REQ-LPA-003 | AC-LPA-003 | additive |
| REQ-LPA-004 | AC-LPA-004 | additive |
| REQ-LPA-005 | AC-LPA-005 | additive |
| REQ-LPA-006 | AC-LPA-006 | additive |
| REQ-LPA-007 | AC-LPA-007, AC-LPA-008 | guard (2 mutations, one per direction) |

Every requirement has at least one criterion; no criterion is orphaned. 7 requirements, 8 criteria — both inside the Tier M ceilings of 16 and 16.

---

## §D.5 Quality gates

Package-scoped, per `plan.md` §E. The full-suite verdict belongs to CI.

| Gate | Command | Threshold |
|---|---|---|
| Tests | `go test -count=1 ./internal/web/...` | green |
| Coverage | `go test -cover ./internal/web/...` | at or above the project target of 85% for the package |
| Lint | `golangci-lint run ./internal/web/...` | 0 issues |
| Format | `gofmt -l internal/web` | no output |
| Vet | `go vet ./internal/web/...` | clean |
| Generated parity | `templ generate` then `git status --short -- internal/web/screens_templ.go` | no output (C-3) |

---

## §D.6 Definition of Done

- All 8 criteria met, with the three guard criteria each citing its observed mutation failure and its revert.
- Both `[NEEDS CLARIFICATION]` items in `plan.md` §D resolved before Implementation Kickoff Approval.
- Every changed file is on the `plan.md` §A scope-surface list. A file outside it means the SPEC was amended or the boundary was crossed; either way it is stated, not absorbed.
- No file under `internal/template/`, `internal/kanban/`, or `internal/cli/` is modified.
