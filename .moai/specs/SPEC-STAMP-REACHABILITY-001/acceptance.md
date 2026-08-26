# acceptance.md — SPEC-STAMP-REACHABILITY-001

Verification layer. Every AC is Given-When-Then, binary-testable, with a concrete verification command. All shell commands run from the worktree root unless stated. Version 0.1.0 · 2026-08-27.

Baseline-attribution (all RED-now observations measured in THIS worktree at base `da791eb0a`, plan phase):

- `git merge-base --is-ancestor 0d15864ae90b origin/main` → rc≠0 (orphan negative fixture live)
- `.moai/project/codemaps/provenance.json` `.commit_sha` = `410da655f39d…`; ancestor-of-origin/main rc=0 (green baseline live)
- No `--commit` flag exists on `moai graph stamp codemaps` (graph_stamp.go:106 flags list: `--root` only); no reachability step exists in graph-freshness.yml (50-line file read: steps are checkout/setup-go/build/bootstrap/check only)

## §D. AC Matrix

### AC-SP-001 — Guard fails on non-ancestor stamp (RED-now)

**Given** a checkout whose history contains object `0d15864ae90b`
**When** the guard logic runs reading a provenance whose `commit_sha` names that object, testing ancestry versus `origin/main`
**Then** the guard exits non-zero with an error annotation naming both `0d15864ae90b` (or its resolved full sha) and the base ref.

Verify:

```bash
cat > /tmp/pv-guard-test.json <<'EOF'
{"schema_version":1,"commit_sha":"0d15864ae90b","dirty":false}
EOF
# [guard-script body equivalent executed locally; expect rc=1 and "::error" naming the sha]
```

Mutation/vacuity guard: the step must be shown executing its ancestry branch (count ≥1 annotated failure across the evidence); a selector matching zero executions is green-vacuous — count the branch hits.

### AC-SP-002 — Guard passes on correct ancestor-named stamps (no false positive; currently satisfied)

**Given** the tracked `.moai/project/codemaps/provenance.json` (`commit_sha` `410da655f…`)
**When** the guard logic runs
**Then** it exits 0, printing a pass line.

Verify now (plan-phase equivalent, literal):

```bash
SHA=$(jq -r '.commit_sha' .moai/project/codemaps/provenance.json) \
  && git merge-base --is-ancestor "$SHA" origin/main && echo GUARD_GREEN
```

Observed at baseline: `GUARD_GREEN`. Run-phase repeats with the actual step script. No-false-positive sweep additionally: both `c9eed8ac6` and the delivering PR's eventual final stamp evaluate green.

### AC-SP-003 — Guard fails on object-absence (environment-dependent shape closed)

**Given** a history-free context lacking the named object — produced by fetching ONLY `main` into a throwaway repo:

```bash
T=$(mktemp -d /tmp/stamp-guard.XXXXXX) && git init -q "$T" \
  && git -C "$T" fetch -q origin main && cd "$T"
# provenance naming 0d15864ae… committed as the fixture file
```

**When** the guard runs there
**Then** it exits non-zero naming the MISSING commit (distinct message from AC-SP-001's ancestry message), demonstrating the CI-side red shape that today manifests as generic exit 2.

Vacuity guard: assert the fetched repo truly lacks the object first (`git cat-file -e 0d15864ae90b^{commit}` rc≠0 inside `$T`) — otherwise the branch never fired.

### AC-SP-004 — Anchorless provenance: skip-with-reason

**Given** a provenance fixture `{…, "dirty": true, "content_fingerprint": "<64hex>", no commit_sha}`
**When** the guard runs
**Then** it exits 0 AND the log carries the printed skip-with-reason line naming the absence of a commit anchor (grep-able token defined in the step text, e.g. `no commit anchor`). Absence of BOTH a failure and a reason line is a FAIL.

### AC-SP-005 — Explicit-commit stamps the named sha verbatim

**Given** a clean described-source tree and resolvable revision `origin/main` (or merge-base)
**When** `go run ./cmd/moai graph stamp codemaps --root "$TREE" --commit "$(git merge-base HEAD origin/main)"` executes against a `t.TempDir()` fixture repo (or the real tree, with restoration)
**Then** written provenance has `dirty:false` and `commit_sha` equal-bytes to the resolved full sha (`jq -r '.commit_sha' pv == git rev-parse <rev>`), schema_version 1, described_roots `[internal cmd pkg]`.
RED-now: flag does not exist (baseline cited above).

### AC-SP-006 — Unresolvable revision rejected

**Given** a revision string resolving to nothing (`deadbeef42`)
**When** the stamp command runs with `--commit deadbeef42`
**Then** exit code ≠ 0, stderr names the resolution failure WITHOUT any absolute local path, and `.moai/project/codemaps/provenance.json` is byte-unchanged (cmp before/after).

### AC-SP-007 — Mixed-anchor combination rejected pre-write

**Given** a tree with ≥1 uncommitted change under internal/, cmd/, or pkg/
**When** stamp runs with a valid `--commit`
**Then** exit ≠ 0, path-free error, NO provenance file written (tmp neither renamed nor left behind), tree unchanged.
Unit-level: also asserts the underlying mx entry rejects without invoking the filesystem install path.

### AC-SP-008 — Default path characterized unchanged

**Given** the flagless invocation over (a) clean tree and (b) tree with described-root changes, snapshotted against pre-M1 behavior
**When** comparing outputs byte-for-byte (stdout shape, WriteFile+rename ordering, printed Describe line)
**Then** identical in all respects. Method: capture `pv.Describe()` strings and generated JSON key-set for both cases pre-edit and post-edit; equality required.

### AC-SP-009 — Freshness-contract regression lock

**Given** the existing suites `go test ./internal/cli/... ./internal/mx/... ./internal/graph/...`
**When** run after M1-M3 edits
**Then** zero failures; specifically-surfaced contracts (named selectors, all observed passing): codemaps changed-files endpoint diff (reverted churn counts zero), dirty-fingerprint anchoring, described_roots fidelity, threshold-from-gate.yaml loading incl. malformed-gate.yaml exit 2, not-comparable system-error path. If a contract's test cannot be located by selector in run phase, the gap is ADDED as a new pinning test before proceeding (record which).
Anti-regression: package tests green ALSO at baseline (captured in plan-phase pre-flight) — no inherited-red masking.

### AC-SP-010 — Operator guidance present, 4 locales

**Given** the docs-site cli-reference graph pages
**When** checking all four locales after M3
**Then** each carries the explicit-commit recipe (`merge-base HEAD origin/main` form) AND the branch-local-HEAD prohibition statement; `hugo -s docs-site --minify --gc` completes warning-free.

```bash
for l in en ja ko zh; do grep -l "merge-base HEAD origin/main" docs-site/content/$l/cli-reference/graph.md || echo "MISSING $l"; done
```

Locale bodies need NOT be transliteration-equal — presence of recipe + prohibition, in native idiom, satisfies (oss-docs i18n rules: derived locales, en canonical).

## §D.1 Severity & traceability

| AC | Severity | REQ covered | Direction |
|---|---|---|---|
| AC-SP-001 | Must | REQ-SR-001, REQ-SR-008 | guard red on bad input |
| AC-SP-002 | Must | REQ-SR-001, REQ-SR-008 | guard green on good input (mirror image) |
| AC-SP-003 | Must | REQ-SR-002 | guard red on missing object |
| AC-SP-004 | Must | REQ-SR-003 | honest skip path |
| AC-SP-005 | Must | REQ-SR-005 | explicit mode records named sha |
| AC-SP-006 | Must | REQ-SR-005 | invalid input rejected |
| AC-SP-007 | Must | REQ-SR-006 | mixed-anchor rejected |
| AC-SP-008 | Must | REQ-SR-007 | default preservation |
| AC-SP-009 | Must | REQ-SR-009, REQ-SR-004 | regression lock |
| AC-SP-010 | Should | REQ-SR-010 | docs parity |

Traceability both directions: every REQ-SR-001..010 maps to ≥1 AC above; every AC maps to ≥1 REQ. REQ-SR-004 rides AC-SP-009 (its ban lives inside the unchanged-check semantics) plus guard-code review in E4.

## §E. Edge Cases

| Case | Expected behavior |
|---|---|
| provenance.json unreadable/tracked-file corrupted | guard treats as failure-equivalent to anchorless-with-anomaly: fail the step (cannot verify what cannot be read — silent pass is the forbidden direction) — pinned by an extra unit row under AC-SP-001's evidence |
| Short-rev form passed to `--commit` (7-char) | accepted, resolves, FULL sha recorded |
| Ref-name form (`main`, `origin/main`) | accepted like any rev-parse-valid expression |
| `--commit` equal to current HEAD | legal no-op relative to default path (byte-identical output) |
| Push-to-main event with orphaned stamp landed anyway | guard fails on object-presence clause on main's own run; graph check still independently red (both signals may fire — independent layers, no suppression) |
| Fork PR where origin/<base_ref> differs across forks | guard resolves base_ref against the checkout's fetched refs; GitHub-hosted fork PRs carry base-repo refs with fetch-depth 0 — if ever unverifiable, the step fails safe (existence clause still binding), NEVER skips silently |

## §F. Quality Gates & Definition of Done

- Targeted package tests: zero failures on `./internal/{cli,mx,graph}/...` (gates at TRUST5 Tested floor for touched packages; coverage on NEW code paths ≥85%).
- `go vet ./internal/cli/... ./internal/mx/...` clean; `gofmt -l` clean on touched files.
- All ten ACs recorded PASS in progress.md §E.2 with command + verbatim output; five-section report format on the run-phase completion report.
- The delivering PR's own graph-freshness run observed green on CI (live witness of the guard's pass path AND premise C2) before sync close.
- Frontmatter transitions owned canonically: draft → in-progress (manager-develop, first run commit), implemented/completed (manager-docs, sync commit).
