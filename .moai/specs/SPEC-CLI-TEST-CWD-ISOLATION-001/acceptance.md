# Acceptance — SPEC-CLI-TEST-CWD-ISOLATION-001

## §D AC Matrix

All commands run from the worktree root
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t334`), Bash timeout ≥ 600 s for any
full-package run, serial (no other lane-local test process in parallel).

### The frozen reproducer (verbatim, incl. env scrub — plan.md B9/D7)

```bash
rm -rf internal/cli/.moai
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR \
  && go test ./internal/cli -run 'Kanban|Factory' -count=1
```

### The baseline-delta tree scan (D1 instrument — audit iter1)

The base tree `d34a789a4` carries **4 COMMITTED `.moai` directories** (verified:
`git ls-files` lists tracked content under all four). An emptiness predicate is therefore
unsatisfiable by construction; the instrument is a **pre-run-baseline delta**:

```bash
# Pre-run (recorded at pre-flight, AFTER the probe residue is removed):
find . -mindepth 2 -name .moai -type d -not -path './.git/*' | sort > /tmp/t334-moai-pre.txt
# Expected baseline content (exactly these 4 lines):
#   ./internal/harness/router/testdata/keyword-force/.moai
#   ./internal/harness/router/testdata/normal/.moai
#   ./internal/harness/router/testdata/spec-overrides/.moai
#   ./internal/template/templates/.moai

# Post-run judgment:
find . -mindepth 2 -name .moai -type d -not -path './.git/*' | sort > /tmp/t334-moai-post.txt
diff /tmp/t334-moai-pre.txt /tmp/t334-moai-post.txt
# PASS (delta clean) = empty diff — no NEW .moai entries (and none removed either).
```

The package-locus binary check (primary, O(1)): `test ! -e internal/cli/.moai`.

| AC | REQ | Given / When / Then (judgment) | RED-now cell | Green path |
|----|-----|-------------------------------|--------------|------------|
| AC-001 | REQ-1 | Given the unfixed tree with `internal/cli/.moai` removed and the recorded 4-line baseline; When the frozen reproducer runs; Then `internal/cli/.moai/` exists post-run with ≥1 file (any state file — name-agnostic per R5). Judgment: `find internal/cli/.moai -type f` lists ≥1; command + verbatim output + exit code + HEAD SHA recorded in `progress.md` §E.2; the post-run scan differs from baseline by `> ./internal/cli/.moai` | **Observed at plan phase** (spec.md §A (1), lead session, base `d34a789a4`, probe record `.moai/reports/t334/red-probe.md`); re-established by the run-phase session with its own attribution triple before any fix commit | M3 (AC-003 = same command, delta clean) |
| AC-002 | REQ-3 | Given the guard present (M2), unfixed tree; When `go test ./internal/cli -run 'Kanban\|Factory\|<guard>' -count=1` runs with the same env scrub; Then the run FAILS with a failure message naming the detected `internal/cli/.moai` path; verbatim output recorded | Populated at M2 (unfixed tree) | M3 |
| AC-003 | REQ-1, REQ-2 | Given the fixed tree, `internal/cli/.moai` absent and the recorded baseline unchanged; When the frozen reproducer runs verbatim; Then rc=0 AND `test ! -e internal/cli/.moai` succeeds AND the post-run scan diffs empty against the baseline. (REQ-4: repeat per individual producing-test selector when M1 names more than one — each run leaves the delta clean) | (RED = AC-001) | M3 |
| AC-004 | REQ-1, REQ-3 | Given the fixed tree, recorded baseline; When `go test ./internal/cli -count=1` runs (timeout ≥ 600 s, serial); Then exit 0 (guard green) AND the post-run scan diffs empty against the baseline | (RED shape = AC-002) | M4 |
| AC-005 | REQ-2, REQ-5 | Given the branch diff vs base `d34a789a4`; When every touched package's tests run (`go test ./internal/<pkg>/...` per `git diff --name-only`; expected: `internal/cli` only) and the diff's non-`*_test.go` files are enumerated; Then all runs exit 0 AND the enumeration is empty of production default-path behavior changes (a test-only diff is the expected shape; an env-guarded default-off seam is admissible only enumerated with its guard named) | N/A (preservation AC) | M4 |

## §D.1 Severity / Traceability

| AC | Severity | REQ | Milestone |
|----|----------|-----|-----------|
| AC-001 | MUST | REQ-1 | M1 |
| AC-002 | MUST | REQ-3 | M2 |
| AC-003 | MUST | REQ-1, REQ-2 (+ REQ-4 via per-selector repeats) | M3 |
| AC-004 | MUST | REQ-1, REQ-3 | M4 |
| AC-005 | MUST | REQ-2, REQ-5 | M4 |

REQ-4 (per-test isolation) is judged by running the AC-003 command once per frozen M1
per-test selector *individually* (not only the combined `Kanban|Factory` string) when M1
identifies more than one producing test; each individual run must leave the delta clean. If
M1 freezes exactly one producing test, REQ-4 and REQ-1 collapse onto the same command and no
extra run is needed.

## §D.2 Preconditions (adoption discipline)

- **Two-cell adoption** (verification-completeness §2): an AC with a RED-now cell is
  *unadopted* until its RED is observed on the stated tree and pinned to its SHA. GREEN without
  the recorded RED counterpart is reported as a Gap, never a PASS. AC-001's plan-phase
  observation satisfies the plan-phase cell; the run-phase session still re-establishes it
  with its own attribution triple before any fix commit.
- **RED for the right reason**: residue created by the run, from the recorded clean baseline
  (probe residue removed; baseline = exactly the 4 committed dirs). A RED caused by
  pre-existing files (baseline not recorded / residue not removed) invalidates the pair —
  restart from plan.md §C pre-flight.
- **Baseline-delta rule** (D1): every GREEN judgment begins by recording the pre-run scan and
  asserts post-run == pre-run. NEVER assert scan emptiness — the base carries 4 committed
  `.moai` roots and emptiness is unsatisfiable (audit iter1 D1).
- **Selector freeze**: the frozen reproducer (command + env scrub + selector) is evidence.
  M1's per-test refinement ADDS child selectors under `Kanban|Factory`; changing the frozen
  command itself re-establishes RED first (plan.md D7).
- **Evidence pinning**: every claim cites the tree SHA it was measured on
  (verification-completeness §4).
- **Mutant probe** (verification-completeness §2): before adopting the guard as AC-002's
  instrument, verify no trivial mutant satisfies it — a guard checking a path that can never
  exist passes vacuously. The guard's RED failure naming the actually-detected residue path
  (AC-002) is the adopted counter-evidence.

## §D.3 Given-When-Then Scenarios

**AC-001 — RED reproduction**
- Given: the unfixed tree at HEAD `<M1 SHA>`, probe residue removed, 4-line baseline recorded
- When: the frozen reproducer runs (env scrub verbatim)
- Then: `internal/cli/.moai/` exists; `find internal/cli/.moai -type f` lists ≥1 file; the
  post-run scan differs from baseline by the added `./internal/cli/.moai` line; command +
  verbatim output + exit code + SHA recorded in `progress.md` §E.2.

**AC-002 — guard RED**
- Given: the guard test present (M2), tree otherwise unfixed, recorded baseline
- When: `go test ./internal/cli -run 'Kanban|Factory|<guard>' -count=1` runs (same env scrub)
- Then: the run FAILS; the failure output names the detected `internal/cli/.moai` path;
  verbatim output recorded.

**AC-003 — targeted GREEN**
- Given: the fixed tree, `internal/cli/.moai` absent, baseline recorded
- When: the frozen reproducer runs verbatim
- Then: rc=0; `test ! -e internal/cli/.moai` succeeds; post-run scan diffs empty vs baseline.
  (REQ-4: repeat per individual M1 selector — each run leaves the delta clean.)

**AC-004 — package-wide GREEN**
- Given: the fixed tree, baseline recorded
- When: `go test ./internal/cli -count=1` runs (timeout ≥ 600 s, serial)
- Then: exit 0 (guard green); post-run scan diffs empty vs baseline.

**AC-005 — producer preservation**
- Given: the branch diff against base `d34a789a4`
- When: touched packages' tests run per `git diff --name-only`, and the non-test diff files
  are enumerated
- Then: all runs exit 0; the enumeration shows no production default-path behavior change.

## §D.4 Indirect Verification

- REQ-5's "identical default behavior" is verified indirectly: (a) touched packages' existing
  suites green (AC-005), (b) the non-test diff enumeration (AC-005), (c) any escape hatch used
  is an existing, documented, default-off env var — no new always-on behavioral branch.
- REQ-2's sandbox-landing property is verified through its observable consequence (baseline
  delta clean — AC-003/AC-004) plus the M1 mechanism record in `progress.md` §E.2 (which
  root-resolution each producing test hit and which ladder rung fixed it).

## §D.5 Closure Gate (Definition of Done)

- [ ] AC-001..AC-005 PASS with recorded evidence (RED cells populated at their milestones).
- [ ] `progress.md` §E.2 populated with attribution triples (command / verbatim output / HEAD
      SHA); §E.3 audit-ready signal present.
- [ ] Branch `WT-cli-test-cwd` integrated into local `develop` per the git-flow lane protocol;
      CI green on `origin/develop` (the full-suite verdict surface).
- [ ] No consumer-side file (`glm.go` walk, `state_dir.go` semantics, hooks) in the diff.
- [ ] Card id `t334` present in every commit message on the branch.

## §D.6 Forward-Looking Checks (post-merge)

- **PRIMARY checkout cleanup (lead-owned)**: the stale residue
  `/Users/goos/MoAI/moai-adk-go/internal/cli/.moai/state/{config-cache.json, kanban/leads.json,
  factory/workers.json}` survives this SPEC (different tree). Once the fix lands on develop,
  the lead removes it so marker walks from the primary's `internal/cli` stop hitting the stale
  `.moai/`.
- **One CI observation cycle**: watch the guard across one develop CI cycle (darwin × windows)
  for flakiness — the guard is order-dependent by design (after `m.Run()`), which is stable,
  but the observation is cheap insurance.

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: RED/GREEN pair per AC; durable guard test; per-test isolation (REQ-4).
- **Readable**: English comments; guard comment explains WHY (residue class + drift history),
  citing this SPEC.
- **Unified**: gofmt; golangci-lint delta vs pre-flight baseline = new findings 0.
- **Secured**: no new input surface, no env var read outside `envkeys.go` constants; no secret
  material in evidence.
- **Trackable**: Conventional commits + card id `t334`; SPEC-ID in scope; related_specs wired.
