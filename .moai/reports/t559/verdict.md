# t559 — GH #1680: recursive language-marker detection for the heavy gate

Worktree: `.claude/worktrees/agent-a412f91a4c4bcb93e` · branch `WT-heavy-gate-recursive`

## Claim

The heavy gate's language detection (`QualityGate.detectToolchain`,
`internal/hook/quality/gate.go` — the card's citation of `internal/cli/gate.go`
~524-527 was stale; the issue's location is the correct one) now finds module
markers below the project top: a bounded recursive scan
(`config.DefaultGateMarkerScanDepth` levels, dependency/build/cache dirs
excluded via the existing `sourceScanSkipDirs`) runs when the project
directory carries no marker, the detected module root is bound into step
execution (`detectedRoot` / `stepDir`), and a root-level marker still outranks
any nested one (existing precedence, byte-identical behavior for today's
flat projects).

## Detection-target count BEFORE / AFTER (card HARD requirement)

Fixture (built under `t.TempDir()`, same shape in both runs): no markers at
the top; `apps/api/go.mod` + `main.go` (depth 2, Go), `services/worker/pyproject.toml`
(depth 2, Python) — the two in-bound detection targets; `node_modules/leftpad/go.mod`,
`vendor/example.com/x/go.mod`, `.git/hooks/go.mod`, `dist/package.json`
(excluded classes); `deep/l2/l3/l4/l5/Cargo.toml` (depth 5, beyond the bound).

| Metric | BEFORE (base tree) | AFTER (patched tree) |
|---|---|---|
| Detection targets present in fixture | 2 | 2 |
| Gate detection count (`detectToolchain`) | **0** (nil — no language detected) | **1** (`[go.mod]`, root `apps/api`; Go wins by table order over Python) |
| Bound step dir (`stepDir`) | n/a (no binding existed) | `apps/api` |

BEFORE, verbatim (`go test ./internal/hook/quality/ -run TestT559MeasureDetectionBefore -v`):

```
=== RUN   TestT559MeasureDetectionBefore
    zz_t559_before_measure_test.go:89: MEASUREMENT detection-targets-present-in-fixture=2
    zz_t559_before_measure_test.go:94: MEASUREMENT gate-detection-count=0 (nil: no language detected)
--- PASS: TestT559MeasureDetectionBefore (0.00s)
```

AFTER, verbatim (`go test ./internal/hook/quality/ -run TestT559MeasureDetectionAfter -v`):

```
=== RUN   TestT559MeasureDetectionAfter
    zz_t559_after_measure_test.go:30: MEASUREMENT detection-targets-present-in-fixture=2 (apps/api go.mod, services/worker pyproject.toml)
    zz_t559_after_measure_test.go:36: MEASUREMENT gate-detection-count=1 language-markers=[go.mod] root-relative=apps/api
    zz_t559_after_measure_test.go:37: MEASUREMENT bound-step-dir-relative=apps/api
--- PASS: TestT559MeasureDetectionAfter (0.00s)
```

The measurement files were throwaway (the before-version targeted the old
signature and cannot compile against the new one); the behaviors they measured
are permanently asserted by `gate_detect_recursive_test.go`.

## Evidence

Commands run, in this worktree, against this tree:

```
$ go build ./internal/hook/quality/ ./internal/config/ && go vet ./internal/hook/quality/ ./internal/config/
VET-OK            # exit 0, no diagnostics

$ go test ./internal/hook/quality/ -run 'TestQualityGate_detectToolchain' -v
--- PASS: TestQualityGate_detectToolchain (0.01s)        # 14/14 subtests, incl. Unknown_project → nil
--- PASS: TestQualityGate_detectToolchain_Flutter (0.00s)
--- PASS: TestQualityGate_detectToolchain_NestedModule (0.02s)
--- PASS: TestQualityGate_detectToolchain_NestedExclusions (0.00s)   # node_modules, vendor, .git, dist, __pycache__
--- PASS: TestQualityGate_detectToolchain_NestedDepthBound (0.00s)   # at bound detected, bound+1 not
--- PASS: TestQualityGate_detectToolchain_NestedConfigScope (0.00s)
--- PASS: TestQualityGate_detectToolchain_NestedGlobMarker (0.02s)
--- PASS: TestQualityGate_detectToolchain_RootLevelPrecedence (0.02s)
--- PASS: TestQualityGate_detectToolchain_NestedTablePrecedence (0.00s)
--- PASS: TestQualityGate_detectToolchain_NestedPythonRunner (0.00s)
PASS
ok      github.com/modu-ai/moai-adk/internal/hook/quality    0.562s

$ go test ./internal/hook/quality/
ok      github.com/modu-ai/moai-adk/internal/hook/quality    14.555s   # full affected package

$ go test ./internal/config/
ok      github.com/modu-ai/moai-adk/internal/config          4.082s    # defaults.go touched

$ golangci-lint run internal/hook/quality/... internal/config/...
0 issues.
```

Changed files:
- `internal/hook/quality/gate.go` — `detectedToolchain{tc, root}`;
  `detectToolchain` = root-level first, then `detectNestedToolchain`
  (bounded `filepath.WalkDir`, `sourceScanSkipDirs` exclusions, no descent
  below a matched module root, candidates ranked by toolchains-table order);
  `matchToolchainAt` factors the former root loop (now also returning the
  table index); `detectedRoot` field + `stepDir` accessor; `Run`,
  `executeStep` (source scan, config-file check), and `runStep` (`cmd.Dir`)
  resolve through `stepDir`. Git-level concerns (staged-file query) and the
  project-level steps (ast-grep, graph-freshness) deliberately stay on the
  project directory.
- `internal/config/defaults.go` — `DefaultGateMarkerScanDepth = 4` (single
  source for the depth bound; comment records the rationale: covers
  apps/&lt;svc&gt; / packages/&lt;pkg&gt; / services/&lt;name&gt; at depth 2 and
  apps/services/&lt;svc&gt; at depth 3 with one spare level).
- `internal/hook/quality/gate_detect_recursive_test.go` — new: nested-module
  detection + root binding, exclusions (5 classes), depth bound (dynamic from
  the constant), table-order precedence across nested candidates, root-level
  precedence, nested glob marker, nested Python runner resolution, nested
  config-file scoping. All fixtures `t.TempDir()`; no OTEL env vars.
- `internal/hook/quality/gate_test.go`, `gate_python_test.go` — mechanical
  updates for the `detectToolchain` return-type change (`dt.tc.…`).

## Baseline-attribution

All measurements above were taken in this run, in this worktree, against this
tree: BEFORE on the unmodified checkout (`git status` clean apart from the
throwaway test file, later deleted), AFTER after the edits recorded in the
changed-files list. The before/after pairing shares one fixture shape and one
counting rule (in-bound, non-excluded module roots), so the 0→1 delta
attributes to the detection change alone. `go vet`, `golangci-lint`, and both
package suites were invoked from the worktree root; no figure is carried from
another tree or run.

## Reuse-first judgment (card item 3) — kept separate, deliberately

The landed SPEC-PRECOMMIT-VET-MONOREPO-001 fix (commit `45c60e1a3`) solves a
related but structurally different problem, and a shared resolution helper was
**not** warranted:

1. **Different language and surface.** That fix lives in the POSIX shell
   pre-commit template (`internal/template/templates/.git_hooks/pre-commit`,
   `_moai_module_root()`); this change is Go in `internal/hook/quality`. No
   helper can span the two without either duplicating the shell into Go
   strings or shipping Go logic to shell.
2. **Opposite walk directions, different predicates.** It ascends from a
   staged file's directory to the nearest enclosing `go.mod` (file-driven,
   Go-only, unbounded upward, fallback to repo root). This change descends
   from the project root across the 16-language marker table (project-scoped,
   glob markers, depth-bounded, exclusion-set pruned, table-order ranked).
   Forcing one shape onto both would weaken at least one.
3. **The issues are complements, not duplicates** — #1679 (that fix) names
   this defect as its "adjacent observation 1" and says so itself.

Per the card's instruction not to pre-emptively merge utilities, the two stay
separate; the only deliberate reuse in this change is `sourceScanSkipDirs`,
which already encoded the vendor/node_modules/.git/dist exclusion judgment in
the same package.

## Gaps

- **Issue layers 2 (multi-toolchain runs) not implemented**: one run still
  executes ONE toolchain (first match by table order, root level first). A
  repo with `package.json` at the top AND a nested `go.mod` runs only the
  Node toolchain — exactly as before this change. The card scoped the work to
  the detection path; detection now sees the nested module but singular
  execution does not fan out to it.
- **End-to-end A/B on an installed binary not run**: the issue's shim-based
  T0'/T1' sandbox verification (real `go vet` red under the patched gate) was
  not reproduced; the PROOF here is unit-level (root binding + stepDir +
  existing cwd tests), not a live `moai gate` run against a nested fixture.
- **Depth bound not config-exposed**: `DefaultGateMarkerScanDepth` is a
  compile-time default with no `gate.yaml` key (card asked for a single
  source, not a config surface). Projects needing deeper scans cannot tune it
  without a rebuild.
- **Fixture-external false positives unexamined**: an in-bound, non-excluded
  marker under an unrelated directory (e.g. `docs/samples/go.mod`) IS
  detected by design; no heuristic beyond depth + skip dirs distinguishes it.

## Residual-risk

- The `detectedRoot` binding follows the run-scoped-field pattern of
  `summary`; a hypothetical future caller executing steps outside `Run`
  without calling `detectToolchain` gets the old project-dir behavior
  (fallback), not the bound root — safe, but the two-step shape must be kept
  in mind when extending.
- A user-supplied `gate.typecheck.command` now executes at the module root
  for nested detections; for a root-level detection the directory is
  unchanged, but an operator whose override assumed the repo top would see
  the cwd move under nested detection.
- The walk runs on every gate invocation when the top level has no marker;
  the depth bound + skip dirs bound it, but a very wide tree (many
  in-bound-level directories) still pays a full shallow traversal per run.
- The issue's `quality gate steps (N configured)` display-delta note is
  inherited: this tree's behavior was verified via package tests, not against
  the issue's exact binary pair, so cosmetic output differences between
  main-line and develop-line builds are unobserved here.
