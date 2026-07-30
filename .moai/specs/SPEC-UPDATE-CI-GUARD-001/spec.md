---
id: SPEC-UPDATE-CI-GUARD-001
title: "CI guard coverage: the checks that gate a merge must be able to see the failure classes this Epic found — config-only PRs, Windows filesystem behaviour, coverage regression, merge semantics, and unenumerated leaks"
version: "0.1.0"
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P1
phase: "v3.0.2"
module: ".github/workflows"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "ci, paths-filter, coverage-gate, windows, semantic-guard, neutrality, merge-semantics, update, config"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-KEY-HONESTY-001, SPEC-UPDATE-YAML-PRESERVE-001, SPEC-V3R6-CI-PR-SPEEDUP-001]
depends_on: []
---

# SPEC-UPDATE-CI-GUARD-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Epic SPEC 5 of 6 from the four-lens audit of `moai update` / `.moai/config`. Findings F1-F5 each re-verified while authoring against baseline HEAD `d5336214e`; F3 coverage independently re-measured; F4 per-function coverage independently re-measured; four drifts recorded (§A.6). |

## §A Problem / Motivation

The four sibling SPECs of this Epic each identify a real defect in `moai update` or `.moai/config`.
This SPEC asks a different question about the same defects: **could CI have seen any of them?**

The answer, measured, is no. A PR carrying only the artefacts SPEC 1-4 are about — shipped template
YAML, or `.moai/config/sections/*.yaml` — runs zero Go tests and is mergeable. The filesystem-heavy
update path is never exercised on Windows before release. Coverage is measured and uploaded but
gates nothing, and the one package most implicated sits below the project's own floor. The merge
function whose semantics SPEC-CONFIG-TIER-PERSIST-001 and SPEC-UPDATE-YAML-PRESERVE-001 record as
defective reports 100% statement coverage on every line. And the template-neutrality guard, being an
enumeration of regexes, is blind by construction to any leak class nobody thought to enumerate.

These are five independent gaps in one layer. The layer is not the defect — the defects are owned by
the siblings — but a defect the gate cannot see will recur after it is fixed. **This SPEC owns the
gate's ability to see; it does not own any of the defects the gate would have caught.**

### A.1 A config-only or template-only PR merges having run zero Go tests (F1)

`.github/workflows/ci.yml:39-70` defines the `dorny/paths-filter` filter set. Verified at HEAD
`d5336214e`:

```
$ grep -n 'go_code:' .github/workflows/ci.yml
65:            go_code:
```

The `go_code` filter is exactly six entries:

```yaml
go_code:
  - '**/*.go'
  - 'go.mod'
  - 'go.sum'
  - 'Makefile'
  - '.github/workflows/ci.yml'
  - '.github/workflows/codeql.yml'
```

It contains no entry for `internal/template/templates/**` and none for `.moai/config/**`. The test
job is gated on it (`ci.yml:86-91`, `if: needs.detect.outputs.go_code == 'true'`), and its
complement `test-skip-marker` (`ci.yml:184`, `if: needs.detect.outputs.go_code != 'true'`) emits a
SUCCESS under the identical required-check name `Test (ubuntu-latest)`.

Consequence: a PR that changes only shipped template YAML, or only `.moai/config/sections/*.yaml`,
evaluates `go_code == false`, skips the Go test job entirely, has its required check satisfied by
the skip-marker, and is mergeable. These are precisely the artefacts SPEC 1-4 of this Epic concern.

Two nuances bind any fix, and neither should be generalised past what is measured:

- **The `**.md` / repo-root asymmetry.** `docs_only` includes `'**.md'`, an unanchored pattern that
  matches markdown *inside* the template tree — `internal/template/templates/.claude/skills/x/SKILL.md`
  matches it. But `docs_only` also lists `'.claude/rules/**'` and `'.claude/skills/**'`, which are
  repo-root-anchored and therefore do **not** match their template-tree counterparts at
  `internal/template/templates/.claude/rules/**`. So the template tree is partially covered by one
  pattern and not at all by two others, for reasons of glob anchoring rather than intent.
- **A template-YAML-only PR is not entirely ungated.** `.github/workflows/template-neutrality-check.yaml`
  triggers on `internal/template/templates/**` and runs three isolated test targets (§A.5). That
  workflow is a *content-neutrality* gate, not a behavioural one — it cannot observe that a config
  key changed meaning. The gap this finding names is the absence of **behavioural** verification, not
  the absence of all verification.

### A.2 No Windows leg exercises the update path at PR time (F2)

`ci.yml:77-84` records the decision, verbatim:

```
# SPEC-V3R6-CI-PR-SPEEDUP-001 — PR-scope test matrix is ubuntu-only.
# GitHub Actions considers a workflow run unfinished until every non-cancelled
# job completes, so the macOS/Windows legs ran to completion even though they
# were never required checks — each PR's CI wall-time stretched toward an
# hour, blocking any caller holding a `gh pr checks --watch`.
# Cross-platform runtime coverage (macOS + Windows race + MX validator) moved
# to release time via .github/workflows/release-pr-multi-os.yml (full 3-OS
# matrix on `release/*` → main PRs).
```

`release-pr-multi-os.yml:33` confirms the trigger scope:
`if: startsWith(github.head_ref, 'release/') || github.event_name == 'workflow_dispatch'`.

A Windows *runner* does exist at PR time, but it does not exercise this code (see §A.6 drift 1):
`ci.yml:214-241` `test-integration` runs `os: [ubuntu-latest, macos-latest, windows-latest]` and its
only run step is

```
go test -tags=integration -race -timeout 180s ./test/integration/harness/...
```

— the harness integration suite, not `internal/cli/update/**`. The thirteen update-subpackage test
files listed by `ls internal/cli/update/*_test.go internal/cli/update/*/*_test.go` therefore execute
on Ubuntu only until a `release/*` PR opens.

This matters for this code specifically because the update path is filesystem-heavy: path
separators, permission bits, symlink and rename semantics, and case-insensitivity all differ on
Windows, and that is the class of defect that has historically surfaced only on the Windows leg.

**The PR-speedup decision is not reopened.** Reverting to a full 3-OS PR matrix would re-introduce
the hour-long wall-time this Epic's sibling explicitly removed. The requirement is a *targeted*
Windows leg scoped to the update subpackages — or an equivalent that preserves the wall-time goal.
The trade-off is stated explicitly in §B.2 and priced in plan.md §F.

### A.3 Coverage is measured, uploaded, and gates nothing (F3)

`ci.yml:163-172`:

```
      - name: Run tests with race detector and coverage
        shell: bash
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v7
        with:
          files: coverage.out
          fail_ci_if_error: false
          token: ${{ secrets.CODECOV_TOKEN }}
```

`ls -a` at repo root shows neither `codecov.yml` nor `.codecov.yml`. No step compares coverage
against a threshold, and `fail_ci_if_error: false` means even an upload failure is non-fatal. The
number is produced, shipped, and discarded.

Freshly measured while authoring (`go test -cover ./internal/cli/... ./internal/config/...`):

| package | coverage |
|---|---:|
| `internal/cli` | **75.7%** |
| `internal/config` | 80.5% |
| `internal/cli/update` | 88.9% |
| `internal/cli/update/backup` | 88.6% |
| `internal/cli/update/deploy` | 97.5% |
| `internal/cli/update/merge` | 90.3% |
| `internal/cli/update/plan` | 95.0% |
| `internal/cli/update/report` | 92.9% |
| `internal/cli/specid` | 58.3% |

Project policy (`CLAUDE.local.md` §6) sets an 85% package minimum and 90% for the critical
`cli` / `template` / `hook` packages. `internal/cli` at 75.7% is below both; `internal/cli/specid`
at 58.3% is below the minimum; `internal/config` at 80.5% is below the minimum. Nothing enforces any
of it.

**The gate's shape is a real decision, not a formality.** A hard 90% floor applied at once to
`internal/cli` would fail the very PR that introduces the gate, and a hard 85% floor would fail
three packages on day one. A no-regression delta gate (a package's coverage may not fall below its
recorded baseline) admits the current state, blocks the actual hazard — silent erosion — and can be
ratcheted upward later. §B.3 requires the delta form and records the floor form as the rejected
alternative, with the reason.

### A.4 One hundred percent statement coverage that proves nothing about merge semantics (F4)

`go tool cover -func` over `internal/cli/update/backup/`, measured while authoring:

```
merge.go:20:	MergeYAML3Way		100.0%
merge.go:43:	DeepMerge3Way		100.0%
merge.go:103:	ValuesEqual		100.0%
merge.go:116:	MergeYAMLDeep		100.0%
merge.go:134:	DeepMergeMaps		100.0%
```

Every function in the file is fully covered. Yet SPEC-CONFIG-TIER-PERSIST-001 and
SPEC-UPDATE-YAML-PRESERVE-001 record real semantic defects in this same code: a three-way merge that
degenerates to two-way because the base is drawn from the new template rather than a snapshot of the
old one, and zero-valued keys skipped so that `false` / `0` / `""` cannot turn anything off.

This is the load-bearing observation of the whole SPEC. **Statement coverage is structurally
incapable of detecting a merge that executes every line and produces the wrong value.** A test that
calls `DeepMerge3Way` and asserts only that it returns without error marks every statement covered;
the assertion that would catch the defect — *this input triple must produce that output map* — is a
different kind of assertion entirely.

§A.3 and this finding are therefore **complementary and non-substitutable**. Raising the number in
§A.3 would not have caught §A.4: `merge.go` was already at 100%. Adding the semantic guard of §A.4
would not have caught the erosion §A.3 names: a package can lose coverage in code the semantic guard
never touches. The requirement here is a **semantic** guard — property-based or table-driven
assertions over a fixture matrix of merge outcomes — NOT a coverage increase.

### A.5 The neutrality guard is an enumeration, so its blind spots are structural (F5)

`.github/workflows/template-neutrality-check.yaml` triggers on changes to
`internal/template/templates/**` and the two guard test files. Its header enumerates the classes the
neutrality audit detects:

```
#   C1 /Users/ macOS-bias path     — binary FAIL
#   C2 bare-narrative V3R[0-9]      — advisory WARN
#   C4 feedback_/memory.md ref      — advisory WARN
#   C5 CLAUDE.local.md ref          — binary FAIL
#   C6 PR #N ref                    — binary FAIL
#   C8 GOOS=<os> Go env var         — PRESERVE (never a violation)
```

Verified against the implementation (`internal/template/template_neutrality_audit_test.go:126-186`):
`C1-macos-bias-path` = `/Users/`, `C6-pr-number-ref` = `PR #[0-9]+`, and so on. Because detection is
a fixed regex enumeration, **any leak class nobody enumerated passes silently.**

SPEC-CONFIG-KEY-HONESTY-001 §A.9 records three concrete leaks the guard structurally cannot see.
Re-verified while authoring:

```
$ grep -rn 'SPEC-AGENT-ARCH-V2-001\|issue #' internal/template/templates/.moai/config/
.../sections/workflow.yaml:65:    # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
.../sections/workflow.yaml:85:    # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
.../sections/llm.yaml:179:    # (issue #653). Claude Code reports context_window_size based on the
```

Three distinct structural reasons, each verified against the pattern set in
`internal/template/internal_content_leak_test.go`:

1. `SPEC-AGENT-ARCH-V2-001` belongs to no registered family. The whole-tree class is
   `C1-spec-id-prefix` = `\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b` (line 171); its sibling
   `C1c-spec-id-non-v3r-known-families` = `\bSPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}\b`
   (line 282). `AGENT-ARCH-V2` is in neither enumeration.
2. `issue #653` is not `PR #N`. The neutrality audit's `C6-pr-number-ref` matches `PR #[0-9]+` only.
3. `plan.md §D D6` is an internal artifact citation, a shape no class in either file covers.

**The enumerated design is deliberate, and that is what makes this a design question rather than a
bug.** The comment at `internal_content_leak_test.go:271-281` states the reasoning verbatim:

```
// Deliberately NARROW (enumerated families, not a generic
// `SPEC-[A-Z-]+-[0-9]+` wildcard): a generic form would flag dozens
// of legitimate pedagogical placeholder SPEC IDs used throughout
// skill bodies (SPEC-BUG-042, SPEC-X-001, SPEC-PAY-001, etc.).
```

A generic wildcard is therefore not a free win — it trades a false-negative class for a
false-positive class across the whole skill corpus. §B.5 requires the pattern set to be extended
with the false-positive cost measured, not assumed.

**Boundary.** This SPEC owns *extending the detection pattern set* so the three shapes above become
detectable. **Removing those three specific leaks from the shipped content belongs to
SPEC-CONFIG-KEY-HONESTY-001 §A.9 (its REQ-CKH-012) and is out of scope here** (§C). The two SPECs
are the two halves of the same handoff: E4 removes the content, E5 makes the class detectable.

### A.6 Drift recorded while authoring

Four discrepancies between the audit's recorded evidence and the tree measured at authoring time.
None reverses a finding; each is recorded rather than silently corrected.

**Drift 1 — `ci.yml:222` job attribution.** The audit attributed
`os: [ubuntu-latest, macos-latest, windows-latest]` at line 222 to the *build* (cross-compilation)
job. Measured: line 222 belongs to the **`test-integration`** job (declared at line 214). The
`build` job is declared at line 287 and uses a `goos`/`goarch` `include` matrix on a single
`runs-on: ubuntu-latest`, not an `os:` matrix. Consequence: a Windows runner **is** present at PR
time (via `test-integration`), which the audit's phrasing "no Windows leg at PR time" understates.
F2's substance is unchanged — that job runs only
`go test -tags=integration ./test/integration/harness/...`, so no `internal/cli/update/**` test
executes on Windows — but the finding is restated in §A.2 as "no Windows leg *exercises the update
path*" rather than "no Windows leg".

**Drift 2 — RESOLVED: `internal/cli` coverage is 75.7% on the single baseline.** At authoring time
two figures existed for the same package: 75.7% recorded against the divergent local branch `main`
(`1d4e4f7da`) and 75.6% re-measured in the worktree, the 0.1pp delta accounted for by two files the
worktree carried and that branch did not (`internal/cli/update_clean_install.go` plus a new
`internal/cli/update_clean_install_merge_notice_test.go`). The two-tree framing is now obsolete: the
Epic branch has been merged with `origin/main` and there is one baseline, `d5336214e`. Re-measured
there (`go test -cover ./internal/cli/... ./internal/config/...`): **`internal/cli` 75.7%**. The
merge brought in `internal/cli/update_deny_migration_test.go` (+149 lines), which restores the 0.1pp.
The figure remains below both the 85% minimum and the 90% critical-package target, so the finding's
direction and consequence are unchanged. §B.3's baseline is nonetheless specified as *recorded at
gate introduction time*, not as a literal constant, so a future drift cannot invalidate the
requirement.

**Drift 3 — the neutrality workflow runs three test targets, not one.** The audit described it as
running "only `-run TestTemplateNeutralityAudit` in isolation". Measured
(`template-neutrality-check.yaml:58-88`): three isolated steps —
(a) `-run TestTemplateNeutralityAudit`,
(b) `-run TestTemplateNoInternalContentLeak` (narrow tier),
(c) the same test under `MOAI_TEMPLATE_LEAK_STRICT: '1'` (strict tier — the date class and the
commit-hash class). Consequence: the date and commit-hash classes **are** CI-enforced today, so
describing them as merely "owned by a disjoint test" understates the current coverage. F5's
substance — that an unenumerated class passes silently — is unchanged, because all three targets are
enumeration-based.

**Drift 4 — two independent `C<N>` numbering schemes.** The audit cited "C1", "C6" as if one
namespace. Measured: `template_neutrality_audit_test.go` numbers C1 = `/Users/`, C2 = bare-V3R,
C4 = `feedback_`/`memory.md`, C5 = `CLAUDE.local.md`, C6 = `PR #[0-9]+`, C9 = canonical-form;
`internal_content_leak_test.go` independently numbers C1 = SPEC-ID prefix, C2 = REQ/AC token,
C3 = audit citation, C4 = date/Finding marker, C5 = memory/archive path, plus skill-scoped
C1b / C1c / C2b / C2c / C6 / C7. **"C1" and "C6" mean different things in the two files.** Every
class citation in this SPEC and its acceptance criteria therefore names the owning file.

## §B Requirements (GEARS)

### B.1 Behavioural gating for config and template changes

**REQ-UCG-001** — **Where** a pull request modifies a file under
`internal/template/templates/**` or `.moai/config/**`, the CI system shall run the Go test suite
rather than satisfy the required check via the skip-marker path. **When** the `detect` job evaluates
its filters for such a pull request, the filter that gates the test job shall evaluate true.

**REQ-UCG-002** — The `detect` job's filter set shall remain mutually exhaustive with respect to the
required check name: for every pull request, exactly one of the real test job and the skip-marker
job shall run, so the required context is always reported. **When** REQ-UCG-001 widens the gating
filter, the skip-marker's complement condition shall be widened identically in the same change.

**REQ-UCG-003** — The repository shall carry a test that fails **when** a path known to be
behaviourally significant (a shipped template YAML path and a `.moai/config/sections` path) does not
match the gating filter. **When** a future edit narrows the filter, this test shall fail rather than
the narrowing passing silently.

### B.2 Windows coverage for the filesystem-heavy update path

**REQ-UCG-004** — **Where** a pull request modifies `internal/cli/update/**`, the CI system shall
execute the `internal/cli/update/**` test packages on a Windows runner before the pull request is
mergeable.

**REQ-UCG-005** — The mechanism satisfying REQ-UCG-004 shall preserve the PR wall-time goal recorded
at `ci.yml:77-84`. It shall be scoped to the update subpackages rather than restoring a full 3-OS
matrix over `./...`, and its added wall-time shall be measured and recorded rather than assumed.

**REQ-UCG-006** — **Where** the Windows leg of REQ-UCG-004 is conditional on a path filter, that leg
shall additionally emit its required check name on the complement path, so a pull request not
touching `internal/cli/update/**` is not blocked waiting for a status that will not arrive.

### B.3 Coverage as a gate, not a report

**REQ-UCG-007** — The CI system shall fail a pull request **when** a measured package's statement
coverage falls below that package's recorded baseline. The gate shall be a **no-regression delta**
against a recorded per-package baseline, not an absolute floor.

**REQ-UCG-008** — The per-package baselines shall be recorded in a tracked file at the time the gate
is introduced, from a measurement observed in that change, so the baseline is attributable rather
than assumed. **Where** a package's coverage legitimately decreases (code deleted, a package split),
the baseline update shall be an explicit edit to that file within the same change.

**REQ-UCG-009** — The gate shall state, at the point of failure, both the baseline and the measured
value for the offending package, so the failure is actionable without re-running the measurement
locally.

**REQ-UCG-010** — **Where** the recorded baseline for a package is below the `CLAUDE.local.md` §6
policy target (85% package minimum, 90% for `cli` / `template` / `hook`), the baseline file shall
record that gap explicitly as accepted debt with the package named, so the delta gate does not
silently normalise a sub-policy figure as if it were compliant.

> Rejected alternative, recorded so it is not silently re-proposed: an **absolute floor** gate at the
> §6 policy values. Rejected because `internal/cli` (75.7%), `internal/config` (80.5%), and
> `internal/cli/specid` (58.3%) are all below policy today (§A.3), so an absolute floor would fail
> the very pull request that introduces the gate and every unrelated pull request thereafter until
> three packages were separately remediated. The delta gate blocks the actual hazard — silent
> erosion — on day one and can be ratcheted toward the policy values incrementally.

### B.4 Semantic verification of merge outcomes

**REQ-UCG-011** — The repository shall carry a semantic guard over `internal/cli/update/backup`'s
merge functions that asserts **merge outcomes** — that a stated input triple produces a stated output
map — over a fixture matrix, rather than asserting only that the functions execute without error.

**REQ-UCG-012** — The REQ-UCG-011 fixture matrix shall include at minimum: a key present in all three
inputs with differing values; a key present only in the user's file; a key present only in the new
template; and a key whose intended new value is the zero value of its type (`false`, `0`, `""`).
**When** a merge implementation drops or ignores any of these cases, the guard shall fail naming the
key and the expected-versus-actual value.

**REQ-UCG-013** — The REQ-UCG-011 guard shall be demonstrated to FAIL against a deliberately broken
merge implementation before it is accepted, so that a guard which passes vacuously cannot be
recorded as satisfying this requirement.

**REQ-UCG-014** — This SPEC shall not modify the merge implementation. **Where** the REQ-UCG-011
guard reveals a defect, that defect shall be reported to its owning sibling SPEC rather than fixed
here; the guard may be committed in a state that documents a known-failing case as a skipped or
explicitly-expected-failure fixture with the owning SPEC ID named.

### B.5 Detection-pattern coverage for the neutrality guard

**REQ-UCG-015** — The template-neutrality detection pattern set shall detect the three leak shapes
measured in §A.5: a SPEC ID whose domain family is not in the current enumeration, an `issue #N`
reference, and an internal artifact citation of the `plan.md §D D6` shape.

**REQ-UCG-016** — **Where** REQ-UCG-015 is satisfied by broadening a pattern rather than extending an
enumeration, the false-positive cost shall be measured against the full shipped template tree and
recorded, because the narrow enumeration is a deliberate choice documented at
`internal/template/internal_content_leak_test.go:271-281` and reversing it trades a false-negative
class for a false-positive class.

**REQ-UCG-017** — Each pattern added under REQ-UCG-015 shall be accompanied by a fixture proving it
matches the intended shape **and** a fixture proving it does not match the pedagogical placeholder
shapes named in the existing comment (`SPEC-BUG-042`, `SPEC-X-001`, `SPEC-PAY-001`).

**REQ-UCG-018** — Every reference to a detection class in this repository's documentation and test
comments shall name the owning file alongside the class identifier, because the two guard files
carry independent `C<N>` numbering schemes (§A.6 drift 4).

### B.6 Non-functional

**NFR-UCG-001** — Every test added by this SPEC shall confine its filesystem writes to `t.TempDir()`.

**NFR-UCG-002** — The REQ-UCG-003 filter-coverage test and the REQ-UCG-011 semantic guard shall each
assert a non-empty, plausible input set, so a fixture list that silently becomes empty fails rather
than passes.

**NFR-UCG-003** — Go sources added by this SPEC shall use `snake_case.go` filenames, wrap errors with
`fmt.Errorf("...: %w", err)`, and carry English comments and godoc.

**NFR-UCG-004** — Workflow changes shall preserve the existing required-check name set exactly. A
change that renames or removes a required context makes every open pull request unmergeable, because
branch protection waits for a status that will not arrive.

## §C Exclusions

### Out of Scope — the merge implementation itself

- Anything inside `internal/cli/update/backup/merge.go` behaviour — merge-sequence atomicity,
  `ValuesEqual` string comparison, `systemFields` depth, nested old-only key dropping, the
  zero-value skip. Owned by `SPEC-UPDATE-YAML-PRESERVE-001` (revision requirements recorded in
  `SPEC-CONFIG-TIER-PERSIST-001` §I). This SPEC may require a guard that **detects** such defects
  (REQ-UCG-011); it may not restate, redesign, or repair the merge itself (REQ-UCG-014).
- The provenance of the three-way merge base (base drawn from the new template rather than a
  snapshot of the old one). Owned by the unwritten `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`; the
  impact table is preserved in `SPEC-CONFIG-TIER-PERSIST-001` §J.

### Out of Scope — sibling-owned config and update behaviour

- Tier precedence, `SrcLocal` ordering, zero-value merging semantics, atomic and mode-preserving
  writes, and the malformed-config contract. Owned by `SPEC-CONFIG-TIER-PERSIST-001`.
- Backup coverage, the update failure contract, and the tautology of
  `TestMoaiUpdate_PreservesUserArea`. Owned by `SPEC-UPDATE-DATA-SURVIVAL-001`.
- v2 detection, deprecated-path handling, and the clean-reinstall loop. Owned by
  `SPEC-UPDATE-REINSTALL-LOOP-002`.
- Dead config keys, the triage rule, the anti-rot key-reader guard, and the `system.yaml` /
  `quality.yaml` binding resolution. Owned by `SPEC-CONFIG-KEY-HONESTY-001`.

### Out of Scope — removal of the three neutrality leaks

- Rewriting `internal/template/templates/.moai/config/sections/workflow.yaml:65`, `:85`, and
  `llm.yaml:179` to remove `SPEC-AGENT-ARCH-V2-001`, `plan.md §D D6`, and `issue #653`. Owned by
  `SPEC-CONFIG-KEY-HONESTY-001` §A.9 / REQ-CKH-012. This SPEC makes those shapes **detectable**
  (REQ-UCG-015); it does not remove the content. The two halves may land in either order, but a
  pattern extension landing first will fail against the un-removed content — plan.md §F records the
  sequencing.

### Out of Scope — reopening the PR-speedup decision

- Restoring a full 3-OS test matrix over `./...` at PR time, or otherwise reverting
  `SPEC-V3R6-CI-PR-SPEEDUP-001`. REQ-UCG-005 explicitly constrains the Windows-coverage mechanism to
  preserve that decision's wall-time goal.
- The `release-pr-multi-os.yml` trigger condition and its 3-OS matrix. Unchanged by this SPEC.

### Out of Scope — CI infrastructure beyond the five named gaps

- The `lint`, `build`, `constitution-check`, and `test-integration` jobs' own contents, except where
  REQ-UCG-002 / REQ-UCG-006 require a complement-path or required-check-name adjustment.
- Codecov account configuration, tokens, or dashboard settings. REQ-UCG-007's gate is a repository
  check, not a Codecov-side setting, so it does not depend on an external service's configuration.
- Test execution speed, caching strategy, or runner sizing, except as REQ-UCG-005 requires the
  Windows leg's added wall-time to be measured.

### Out of Scope — code changes in this Epic

- No code changes. This Epic's six SPECs are plan-phase artefacts only; run-phase implementation is
  a separate, separately-approved step.
