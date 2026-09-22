# dr0912 — verdict

Class B regression repair, cycle_type=tdd. Worktree
`.claude/worktrees/dr0912`, branch `WT-destructive-registry-row`, base local
develop `30cf7f422`.

## Claim

1. `TestDestructiveTargetRegistry_CoversAllSites` failed at `30cf7f422` with
   `unregistered destructive site: internal/cli/update/backup/settings_snapshot.go
   promoteSettingsSnapshot has 1 os.RemoveAll/os.Rename call site(s) but no
   registry row`, scan 13 vs registry 12.
2. The repair is exactly one registry row (plus the header accounting that row
   invalidates) in `internal/cli/update_destructive_registry.go`. No product
   file changed; `settings_snapshot.go` is untouched.
3. The row carries an **Exemption**, not a Protection.
4. After the repair the selector passes, and the counts are 13 pairs / 23 sites
   on both sides.
5. The guard actually reads the new row: deleting it reproduces the original
   failure verbatim; restoring it restores PASS.
6. `go vet ./internal/cli/` and `golangci-lint run ./internal/cli/` both exit 0.

## Evidence

All runs used `go test ./internal/cli/ -run TestDestructiveTargetRegistry
-count=1 -v -timeout 600s`, one at a time, no full suite, no background load.

| # | Command | Exit | Artifact |
|---|---|---|---|
| 1 | the selector, at `30cf7f422` | 1 | `red.txt` |
| 2 | the selector, after the row | 0 | `green.txt` |
| 3 | the selector, row deleted then restored | 1 then 0 | `mutant.txt` |
| 4 | `go vet ./internal/cli/` | 0 | `vet.txt` (empty) |
| 5 | `golangci-lint run ./internal/cli/` | 0 | `lint.txt` (`0 issues.`) |

RED (`red.txt`, verbatim):

```
=== RUN   TestDestructiveTargetRegistry_CoversAllSites
    update_destructive_registry_test.go:57: unregistered destructive site: internal/cli/update/backup/settings_snapshot.go promoteSettingsSnapshot has 1 os.RemoveAll/os.Rename call site(s) but no registry row
    update_destructive_registry_test.go:87: scanned 13 (file, function) pair(s) across internal/cli/update and internal/cli/update*.go; registry has 12 row(s)
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.935s
FAIL
```

GREEN (`green.txt`, verbatim):

```
=== RUN   TestDestructiveTargetRegistry_CoversAllSites
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.036s
```

Mutant (`mutant.txt`, both directions, verbatim): with the new row deleted the
run fails with the identical `unregistered destructive site: …
promoteSettingsSnapshot …` line and the identical `scanned 13 … registry has
12` tally; with the row restored the run passes. The restored file was verified
byte-identical to the committed state (`git diff --stat HEAD --
internal/cli/update_destructive_registry.go` — empty) before the second run, so
the PASS is the committed artifact's, not a near-miss variant's.

Counts (`counts.txt`): before 13 pairs vs 12 rows (observed in the failure log);
after 13 rows (`grep -c 'File: "'`) against 23 call sites enumerated by grep
across both scan scopes — 14 under `internal/cli/update/`, 9 under
`internal/cli/update*.go` — i.e. 13/13.

**Classification and its cited basis.** The row is an **Exemption**.
`spec.md:278-280` REQ-UDS-006 requires every destructive operation to be
enumerated "naming, per target, the protection set that covers it **or the
recorded reason it is exempt**", and `spec.md:208-209` (§B Goals) frames the
alternative as a target "explicitly and deliberately exempt, with the exemption
recorded"; the registry's own field doc (`update_destructive_registry.go:34-39`)
splits the two as Protection = "what stands between this site and user data
loss" versus Exemption = "why this site destroys nothing that needs
protecting". Neither operand of this rename is user data. The source is the
staging copy `.moai/cache/template-snapshot/claude/settings.json.pending`
written by the same flow's deploy, and its content survives the move. The
destination is the canonical merge base left by a previous flow — moai's own
record of a template render under `.moai/cache/`, not the user's
`.claude/settings.json`, which `promoteSettingsSnapshot` never touches. Its loss
is bounded: `SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001` REQ-USB-009
(`spec.md:175`) has the merge fall back to the derived base when the canonical
base is absent or unreadable, without failing the update. So there is no user
data for a protection to stand in front of, which is the Exemption side of the
REQ-UDS-006 disjunction rather than the Protection side.

Two neighbouring rows were checked against this reading rather than assumed.
`archiveLegacySkills` is also an `os.Rename` and carries a *Protection* — but
its operand is a legacy skill directory in the user's `.claude/` tree, i.e.
user-area content, and the recorded protection is that the move preserves it.
`CleanupOldBackups` is the closest exempt precedent (a moai-authored artifact
from a PREVIOUS run) and is recorded as exempt-but-not-harmless. This row is a
third ground, so the registry's "two materially different grounds" note was
widened to three rather than folding cache-base replacement into either of
them.

## Baseline-attribution

Every figure above was measured in this run, in this worktree, against this
tree. The RED run was taken at `30cf7f422` with a clean working tree
(`git status --short` empty, re-read immediately before the run). The GREEN,
mutant, vet and lint runs were taken on the repaired tree, whose only delta from
`30cf7f422` in tracked product code is `internal/cli/update_destructive_registry.go`
(the rest of the diff is `.moai/reports/dr0912/`). Go toolchain: the worktree's
configured toolchain, unchanged across all five runs. No figure is carried over
from another tree, package, or point in time; the `scanned 13` term appears in
two independently observed runs (RED at `30cf7f422` and the mutant on the
repaired tree), which is what licenses treating it as unchanged by the fix.

## Gaps

- **The full suite was not run**, by instruction. Only
  `-run TestDestructiveTargetRegistry` in `./internal/cli/` was executed. Other
  tests in that package, and every other package, are unobserved here; CI at the
  integration head is the verdict for them.
- **The Test and Race CI jobs were not re-run.** The lead's report attributes
  both failures to this test; that attribution is consumed, not re-verified — a
  second, unrelated failure in either job would not have been visible to any
  measurement taken here.
- **`go test -race` was not run** for this selector. The test is a pure
  single-goroutine AST scan, but that is an argument, not an observation.
- **The Exemption's substantive claim is a code-and-spec reading, not a runtime
  observation.** That `promoteSettingsSnapshot` never touches
  `.claude/settings.json` was established by reading `settings_snapshot.go` and
  the REQ text; no test was run that would fail if a future edit made it touch
  user data.
- **Cross-platform behaviour was not measured.** The claim that the rename
  replaces an existing file on every platform is the existing doc comment's,
  carried forward unverified.

## Residual-risk

- The registry is a human-maintained assertion. The guard proves a row *exists*
  with the right count; nothing mechanically checks that the Exemption's
  *reason* is true. A wrong reason passes exactly like a right one — which is
  why the reason is written to be falsifiable by reading, and cites the REQ it
  rests on.
- The header comment's "13 rows covering 23 call sites" is prose, not an
  assertion the guard evaluates. It was correct at this commit (23 sites
  enumerated in `counts.txt`) but can silently rot: a future row that updates
  the table and not the comment fails no test.
- Classifying this site as Exempt removes it from the user-data protection set
  permanently. If `.moai/cache/template-snapshot/` ever becomes a location users
  are expected to edit or rely on, this row's ground disappears and the
  classification must be revisited — nothing signals that transition.
- The third exemption ground widens the taxonomy. Widening lowers the bar for a
  future site to be argued exempt by analogy rather than on its own merits; the
  row's wording names what distinguishes it from the other two grounds
  specifically to make such an analogy visible.
