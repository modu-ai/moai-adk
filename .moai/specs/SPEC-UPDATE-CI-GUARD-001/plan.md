# SPEC-UPDATE-CI-GUARD-001 — Implementation Plan

Milestones are ordered by **decision-reversibility**: the two gate-shape decisions come first
because they are the ones most likely to change under review, the filter and fixture design follow,
and the mechanical edits land last.

## §A Context

Baseline tree: HEAD `d5336214e`, branch `plan/epic-update-config-audit`, merged with `origin/main`
(`git rev-list --count HEAD..origin/main` → 0). Authored at `9426bf49b` on the same branch, before
the merge; the local branch `main` at `1d4e4f7da` was a stale divergent branch and is not an
ancestor of this baseline. All `file:line` references in `spec.md` §A were re-verified while
authoring; four drifts were found and recorded (spec.md §A.6), of which drift 2 is now resolved by
the merge.

Three measurements this plan depends on, all observed while authoring:

- The `go_code` paths-filter is six entries (`ci.yml:65-71`) and contains neither
  `internal/template/templates/**` nor `.moai/config/**`. The test job is gated on it (`ci.yml:90`);
  its complement `test-skip-marker` (`ci.yml:200`) satisfies the same required check name.
- Per-package coverage: `internal/cli` 75.7%, `internal/config` 80.5%, `internal/cli/specid` 58.3%,
  `internal/cli/update` 88.9%, `.../backup` 88.6%, `.../deploy` 97.5%, `.../merge` 90.3%,
  `.../plan` 95.0%, `.../report` 92.9%.
- `go tool cover -func` over `internal/cli/update/backup/`: all five `merge.go` functions at 100.0%,
  package total 88.6%.

## §B Known issues carried into this SPEC

1. **Required-check names are load-bearing.** Branch protection matches on the literal check name
   `Test (ubuntu-latest)`. The `test` / `test-skip-marker` pair exists precisely so that exactly one
   of them always reports that name. Any filter change must move both sides together, or a pull
   request waits forever for a status that will not arrive — the failure mode already documented in
   `ci.yml:186-198`.
2. **A pattern extension can land before the content it flags is removed.** REQ-UCG-015's patterns
   will match the three leaks that `SPEC-CONFIG-KEY-HONESTY-001` owns. Landing M5 before that SPEC's
   M6 turns `main` red. Sequencing is stated in §F M5.
3. **`main-fork/` is untracked and on disk.** Any inventory or scan step must walk git-tracked paths
   (`git ls-files`), not the filesystem, or it will include a local clone.
4. **This checkout is shared with concurrent sessions.** `git stash` is prohibited; falsification
   uses `go test -overlay` or a scratch worktree driven by `go -C`.

## §C Pre-flight

```bash
git rev-parse --short HEAD                       # expect d5336214e or a recorded successor
go build ./... && go vet ./...
go test -count=1 ./internal/template/... ./internal/cli/update/...
gh workflow list                                  # confirm the workflow set is as recorded
```

## §D Constraints

- **D1 — required-check name preservation.** No milestone may rename or remove a required check
  context. A new job that gates a merge must either reuse an existing name or be added to branch
  protection deliberately, and the complement path must be covered (NFR-UCG-004).
- **D2 — PR wall-time.** The Windows leg must not restore a full 3-OS matrix over `./...`. Its added
  wall-time is measured, not assumed (REQ-UCG-005).
- **D3 — no merge-implementation edits.** M4 adds a guard only. A defect the guard surfaces is
  reported to its owning sibling SPEC, never fixed here (REQ-UCG-014).
- **D4 — no leak-content edits.** M5 extends patterns only. The three leaks in
  `internal/template/templates/.moai/config/` belong to `SPEC-CONFIG-KEY-HONESTY-001`.
- **D5 — template neutrality.** No internal SPEC ID, REQ/AC token, internal date, commit SHA, or
  artifact citation may enter `internal/template/templates/**`. Fixtures that must contain such
  tokens live in `_test.go` files or `testdata/`, never in the shipped tree.
- **D6 — test isolation.** `t.TempDir()` only (NFR-UCG-001).

## §E Self-verification

Each milestone closes only when its acceptance.md criteria print the stated observable output. §E of
`progress.md` carries the run-phase evidence, and every claim there cites the command that produced
it per `.claude/rules/moai/core/verification-claim-integrity.md`.

## §F Milestones

### M1 — Coverage gate contract (REQ-UCG-007 … REQ-UCG-010)

The highest-change-likelihood decision in this SPEC: what shape the coverage gate takes, and what it
does to every subsequent pull request in the repository.

**Deliverable 1 — the baseline file**, `.moai/config/coverage-baselines.yaml` (tracked, project-local
— NOT under `internal/template/templates/`, so it ships to nobody). One entry per measured package:

```yaml
packages:
  github.com/modu-ai/moai-adk/internal/cli:
    baseline: 75.7
    policy_target: 90        # CLAUDE.local.md §6 critical-package target
    accepted_debt: true      # REQ-UCG-010 — below policy, recorded explicitly
  github.com/modu-ai/moai-adk/internal/cli/update/backup:
    baseline: 88.6
    policy_target: 85
    accepted_debt: false
```

Every `baseline` value is written from a measurement observed in the introducing change and cited in
`progress.md` §E.2 (REQ-UCG-008). The `accepted_debt` flag is what stops the delta gate from
normalising a sub-policy figure (REQ-UCG-010).

**Deliverable 2 — the gate.** A CI step that parses `coverage.out` (already produced at
`ci.yml:165`), computes per-package statement coverage, compares against the baseline file, and
exits non-zero on any regression, printing `<package>: baseline X.X%, measured Y.Y%` (REQ-UCG-009).
Implemented as a Go test (`internal/quality/coverage_baseline_test.go`) rather than a shell step, so
it is runnable locally with the same command CI runs.

**Decision to surface at review — delta vs floor.** The chosen form is a **no-regression delta**.
The rejected alternative and its reason are recorded in spec.md §B.3: an absolute floor at the §6
policy values fails three packages on day one, including the pull request that introduces the gate.
The reviewer should confirm this trade, because the delta gate's weakness is real — it admits and
freezes the current sub-policy state until someone ratchets it, and nothing in this SPEC forces the
ratchet.

**Second decision to surface — tolerance.** Statement coverage is deterministic for a fixed tree but
not across trees; adding a file with untested lines legitimately lowers a package's figure. The gate
therefore needs either an explicit zero tolerance (any decrease fails, forcing a baseline edit in the
same change) or a small epsilon. Zero tolerance is the proposal, because an epsilon silently permits
repeated sub-epsilon erosion — but it makes the baseline file a frequently-edited file, and the
reviewer should weigh that friction.

### M2 — Windows leg for the update path (REQ-UCG-004 … REQ-UCG-006)

The second gate-shape decision, and the one that spends wall-time.

**Chosen mechanism.** A new conditional job in `ci.yml`:

```yaml
  test-update-windows:
    name: Test Update (windows-latest)
    runs-on: windows-latest
    needs: detect
    if: needs.detect.outputs.update_path == 'true'
    # go test -race ./internal/cli/update/...
```

plus a `update_path` filter output listing `internal/cli/update/**`, plus a complement skip-marker
job emitting the same check name on `update_path != 'true'` (REQ-UCG-006, and D1's failure mode).

**The price, to be measured not assumed (REQ-UCG-005).** The six update packages take roughly 25s
wall-time on this macOS host (`update` 4.0s, `backup` 6.5s, `deploy` 1.8s, `merge` 4.5s, `plan` 5.3s,
`report` 2.6s); a Windows runner is slower and adds checkout plus Go setup. The measured figure goes
into `progress.md` §E.2. If it materially erodes the PR-speedup goal recorded at `ci.yml:77-84`, the
scope narrows further (e.g. `internal/cli/update/backup` + `.../merge` only — the two packages whose
defects the siblings actually record) rather than the requirement being dropped.

**Alternatives considered and rejected:**

- *Restore the full 3-OS matrix.* Rejected — directly reverts `SPEC-V3R6-CI-PR-SPEEDUP-001` and
  reinstates the hour-long PR wall-time (spec.md §C).
- *Rely on `release-pr-multi-os.yml`.* Rejected — that workflow triggers only on `release/*` → main
  pull requests, so a Windows-specific regression in the update path is discovered at release, which
  is exactly the "fix requires a hotfix" failure mode its own header argues against for the cron
  variant.
- *Add the update packages to the existing `test-integration` Windows leg.* Rejected — that job runs
  `-tags=integration ./test/integration/harness/...` and `needs: test`; widening it would couple the
  update path's gating to the harness suite's and to `go_code`, which M3 is separately changing.

### M3 — Widen the behavioural gating filter (REQ-UCG-001 … REQ-UCG-003)

**Chosen approach.** Add a **new** `detect` output rather than widening `go_code`:

```yaml
            behavioral:
              - '**/*.go'
              - 'go.mod'
              - 'go.sum'
              - 'Makefile'
              - '.github/workflows/ci.yml'
              - '.github/workflows/codeql.yml'
              - 'internal/template/templates/**'
              - '.moai/config/**'
```

and gate `test` on `behavioral == 'true'`, with `test-skip-marker` on `behavioral != 'true'`
(REQ-UCG-002). `go_code` is left in place for any other consumer.

**Why a new output rather than editing `go_code`.** The name `go_code` is descriptive and is
referenced in job comments; widening it to include YAML would make the name a lie — the same class
of defect this Epic keeps finding. A separate `behavioral` output states what the filter actually
governs. The reviewer may prefer the smaller diff of editing `go_code` in place; that is a real
alternative, and the cost is only the naming.

**Guard (REQ-UCG-003).** `internal/template/ci_filter_coverage_test.go` parses `ci.yml`, extracts the
filter that gates the test job, and asserts that a representative shipped-template YAML path and a
`.moai/config/sections` path both match it. NFR-UCG-002 requires the representative path list to be
non-empty and to be asserted non-empty, so a fixture list that silently empties fails.

**Note on the `**.md` anchoring asymmetry (spec.md §A.1).** Widening the gating filter makes the
asymmetry moot for the *test* gate — every template-tree path now gates behaviourally. The
`docs_only` output's own anchoring is left unchanged; it is not consumed by the test gate after this
change, and re-anchoring it would alter unrelated fast-track behaviour.

### M4 — Semantic merge-outcome guard (REQ-UCG-011 … REQ-UCG-014)

`internal/cli/update/backup/merge_semantics_test.go`, table-driven over a fixture matrix. Each case
is `(base, user, new) -> expected`, asserted on the whole output map rather than on absence of error.

Minimum matrix (REQ-UCG-012), each case named for the property it pins:

| case | base | user | new | property under test |
|---|---|---|---|---|
| `three_way_divergence` | `k: 1` | `k: 2` | `k: 3` | user's deliberate change survives a template change |
| `user_only_key` | absent | `k: v` | absent | a user-added key is not dropped |
| `template_only_key` | absent | absent | `k: v` | a new template key is introduced |
| `zero_value_false` | `k: true` | `k: true` | `k: false` | `false` is a value, not an absence |
| `zero_value_empty_string` | `k: "x"` | `k: "x"` | `k: ""` | `""` is a value, not an absence |
| `zero_value_int` | `k: 5` | `k: 5` | `k: 0` | `0` is a value, not an absence |
| `nested_old_only` | `a.b: 1` | `a.b: 1` | `a: {}` | nested user keys survive a template restructure |

**Falsification (REQ-UCG-013, acceptance.md §C).** The guard is demonstrated to FAIL against a
deliberately broken merge via `go test -overlay`, substituting a `merge.go` whose zero-value branch
is removed. A guard that passes against the broken implementation is inert and is not accepted.

**Expected-failure handling (REQ-UCG-014, D3).** The zero-value cases are expected to fail against
the current implementation — that is the defect `SPEC-UPDATE-YAML-PRESERVE-001` owns. Those cases are
committed as explicitly-expected failures naming that SPEC ID in the skip reason, so the guard is
green on `main` while the pending defect stays visible and the expectation flips to a hard assertion
when the sibling lands. **The alternative — withholding the whole guard until the sibling lands — is
rejected**, because it makes this SPEC's delivery depend on another SPEC's run-phase.

### M5 — Neutrality detection-pattern extension (REQ-UCG-015 … REQ-UCG-017)

Three shapes to detect, each with its own false-positive question:

1. **Unregistered SPEC-ID family.** Extend `internal_content_leak_test.go`'s `C1c` enumeration with
   `AGENT-ARCH-V2`, or replace the enumeration with a generic
   `SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}` plus an explicit allowlist of the pedagogical placeholders. The
   generic form is the durable fix; its cost must be measured (REQ-UCG-016) because the existing
   comment states a generic wildcard "would flag dozens of legitimate pedagogical placeholder SPEC
   IDs". **The measurement decides**, and it goes in `progress.md` §E.2 — if the false-positive count
   is small and enumerable, the generic form plus allowlist wins; if it is genuinely dozens across
   unrelated skills, the enumeration extension wins for now.
2. **`issue #N`.** Extend the neutrality audit's `C6-pr-number-ref` from `PR #[0-9]+` to
   `(PR|issue) #[0-9]+`. Low false-positive risk; the phrase "issue #N" in a shipped template is an
   internal tracker reference in every case measured.
3. **Internal artifact citation.** A new class matching `(plan|spec|acceptance|progress)\.md §`.
   Needs a fixture proving it does not match a legitimate instruction telling a *user* to read their
   own `plan.md` — that shape exists in skill bodies and is not a leak.

Each pattern ships with a positive fixture and a negative fixture (REQ-UCG-017), the negative one
covering `SPEC-BUG-042` / `SPEC-X-001` / `SPEC-PAY-001` verbatim from the existing comment.

**Sequencing (§B item 2, D4).** These patterns will match the three leaks currently in
`internal/template/templates/.moai/config/` that `SPEC-CONFIG-KEY-HONESTY-001` REQ-CKH-012 removes.
Land this milestone **after** that SPEC's leak removal, or land it with the new classes in the
advisory-WARN tier and promote them to binary-FAIL in a follow-up once the content is clean. The
second option is preferred because it does not couple this SPEC's completion to another SPEC's
run-phase; the promotion step is recorded as a one-line follow-up rather than left implicit.

### M6 — Class-naming hygiene (REQ-UCG-018)

Mechanical. Every citation of a detection class in a test comment, a workflow comment, or
`CLAUDE.local.md` §25 gains its owning filename: `internal_content_leak_test.go C1` versus
`template_neutrality_audit_test.go C1`. Motivated by spec.md §A.6 drift 4 — the two files number
their classes independently and both define a `C1` and a `C6` meaning different things, which made
the audit's own citations ambiguous.

## §G Anti-patterns

- **AP-1 — a coverage gate that fails on the change introducing it.** An absolute floor at the §6
  policy values fails three packages on day one (spec.md §A.3). The delta form exists to avoid this.
- **AP-2 — raising coverage instead of asserting semantics.** `merge.go` is at 100% and still wrong
  (spec.md §A.4). A milestone that satisfies M4 by adding coverage rather than outcome assertions has
  not satisfied it.
- **AP-3 — widening a filter without widening its complement.** The `test` / `test-skip-marker` pair
  must move together or the required check is never reported (§B item 1).
- **AP-4 — reverting the PR-speedup decision.** A full 3-OS matrix is out of scope (spec.md §C); the
  Windows leg is scoped to the update subpackages.
- **AP-5 — a semantic guard with no falsification.** A guard not demonstrated to FAIL against a
  broken implementation is indistinguishable from an inert one (REQ-UCG-013).
- **AP-6 — a generic leak pattern adopted without measuring false positives.** The narrow
  enumeration is deliberate and documented; reversing it on assumption trades one failure class for
  another (REQ-UCG-016).
- **AP-7 — fixing the merge, or removing the leaks, inside this SPEC.** Both are sibling-owned
  (D3, D4).
- **AP-8 — `git stash` for falsification.** Prohibited on this shared checkout (§B item 4).

## §H Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — every claim in `progress.md` §E must
  cite an observed command.
- `CLAUDE.local.md` §6 (coverage targets — the policy values M1 records as `policy_target`), §15
  (16-language template neutrality), §25 (template internal-content isolation).
- `.github/workflows/ci.yml` — `detect` (39), `test` (86), `test-skip-marker` (184),
  `test-integration` (214), `lint` (246), `build` (287).
- `.github/workflows/template-neutrality-check.yaml` — the three isolated test steps (58, 63, 73).
- `.github/workflows/release-pr-multi-os.yml` — the release-time 3-OS matrix M2 must not duplicate.
- `internal/template/internal_content_leak_test.go`, `internal/template/template_neutrality_audit_test.go`
  — the two independently-numbered class sets M5 extends and M6 disambiguates.
