# Progress — SPEC-CODEX-MIRROR-DOCTOR-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t498 · worktree `.claude/worktrees/t498` · branch `WT-codex-mirror-doctor` · base `ace1c5440`
- Tier: M (3-artifact set: spec.md + plan.md + acceptance.md) · plan-auditor PASS threshold 0.80
- Requirements: 11 (ceiling 16) · Acceptance criteria: 15 (ceiling 16)
- SPEC ID regex check executed as Bash: `PASS` (re-executed at iteration 2)
- Authoritative input: `.moai/reports/t498/root-cause.md` (cited, not re-derived)
- Status: `draft` — awaiting plan audit and Implementation Kickoff Approval

### Repair iteration 2 (spec.md 0.2.0 · acceptance.md 0.2.0 · plan.md unchanged)

Closes the six blocking findings of `.moai/reports/t498/plan-audit.md` (iteration 1, FAIL 0.775):

| Finding | Where closed |
|---|---|
| D1 — AC-CMD-008 fails as written (chmod 0o000 breaks `t.TempDir()` cleanup) | acceptance.md AC-CMD-008 rewritten onto the file's symlink-loop idiom (`doctor_codex_test.go:335-352`) with its three fixture obligations |
| D2 — nine ACs decidable by an empty sweep | acceptance.md gains a file-level **swept-count gate** section plus a per-AC clause; AC-CMD-001 gains RED-now + green-path cells pinned to tree `dcb3ba0c7` |
| D5 — REQ-CMD-009 tail-drop clause unjudged | acceptance.md AC-CMD-013 (new), including the lead-summary exception sub-case |
| D6 — REQ-CMD-001/011 uncovered, no AC states its REQ | acceptance.md AC-CMD-014 / AC-CMD-015 (new) + a `REQ:` line under all 15 AC headings; AC-CMD-001 states it **supplements** `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent` |
| D4 — REQ-CMD-006/007 wiring precondition | spec.md §2 — both requirements gain `Where the project declares Codex wiring`, matching plan.md §D M2; AC-CMD-007 gains sub-case (b) to judge it |
| D3 — spec.md §6 asserts an unmeasured condition | spec.md §6 cross-platform clause re-stated conditionally, matching §3 row 3 |

D7-D10 (optional) are deliberately not acted on, per the auditor's own recommendation. D8 (path-literal
drift between the doctor and the producer's unexported `mirrorSkillsRelDir`) is recorded as a
follow-up card candidate, not fixed here.

Mechanical re-verification at iteration 2 (tree `dcb3ba0c7`):

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md
✓ No findings — all SPEC documents are valid                                    EXIT=0
$ grep -c '^## AC-CMD-' acceptance.md → 15   $ grep -c '^REQ: ' acceptance.md → 15
$ REQ: lines cite REQ-CMD-001…011 → all 11 covered, none orphaned
```

## §E.2 Run-phase Evidence

Cycle: `tdd` (RED-GREEN-REFACTOR). Worktree `.claude/worktrees/t498`, branch `WT-codex-mirror-doctor`,
base `f842cb612`. All measurements below were taken in this run, against this tree.

### Pre-edit baseline (AC-CMD-011 baseline attribution)

`go test ./internal/cli/... -count=1` was NOT measured at plan time (deliberate, per CLAUDE.local.md
§4 load discipline); the run phase took its own baseline before the first edit, so a pre-existing
red would have been attributable. All three gates were green at baseline:

```
$ go test ./internal/cli/... -count=1 -timeout 600s      EXIT=0   (ok … internal/cli 353.875s, 17/17 packages ok)
$ go vet ./internal/cli/...                              EXIT=0   (no output)
$ golangci-lint run ./internal/cli/...                   EXIT=0   (0 issues.)
```

Measurement note: the first `go vet` invocation was written as `… 2>&1 | tail -5; echo "VET_EXIT=${PIPESTATUS[0]}"`
and printed an empty `VET_EXIT=` — zsh does not populate `PIPESTATUS` that way, so that read was
vacuous and was discarded. The exit code above is from a re-run using a direct `$?` read.

### RED evidence (E8 — verbatim pre-GREEN failing output)

The tests were authored first and run before any production change. They compiled cleanly (they
assert on `checkCodexWiring`, which already existed), so the RED is an assertion failure rather
than a build error:

```
$ go test ./internal/cli/ -run '<11 new test names>' -count=1 -v      EXIT=1
--- FAIL: TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy (0.00s)
    doctor_codex_test.go:1099: absent mirror on a wired project status = ok, want Warn: {…Message:wired and consistent…}
--- FAIL: TestCheckCodexWiring_DanglingMirrorEntriesCounted (0.00s)
--- FAIL: TestCheckCodexWiring_CopyModeDetailOnly (0.01s)
    doctor_codex_test.go:1155: Detail does not report the copy-mode count: ""
--- FAIL: TestCheckCodexWiring_UnmirroredSkillsDetailOnly (0.01s)
--- FAIL: TestCheckCodexWiring_MirrorUnreadableIndeterminate (0.01s)
--- FAIL: TestCheckCodexWiring_MirrorSummaryWidth (0.01s)
--- FAIL: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop (0.01s)
--- FAIL: TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters (0.01s)
```

Three of the eleven PASSED at RED — `ClaudeOnlyMachineNoMirrorRow`, `UnwiredNoMirrorNag`,
`MirrorCheckIsReadOnly`. This is expected and disclosed rather than hidden: all three are
**absence-guards** (they assert that certain text/state does NOT appear), and an absent feature
trivially satisfies them. RED-now therefore proves nothing about their discriminating power, so
each was probed with a mutant instead — see § Mutation probes.

### Mutation probes (non-vacuity of the three absence-guards)

| Guard | Mutant applied | Result |
|---|---|---|
| `ClaudeOnlyMachineNoMirrorRow` (AC-CMD-001) | mirror **finding** appended to the claude-only early-return Message | **FAIL** — `claude-only machine was nagged about the mirror (".agents" present)` |
| `UnwiredNoMirrorNag` (AC-CMD-007) | inspector call hoisted out of the `wired` branch | **FAIL** — both sub-cases: `unwired project was nagged about the mirror` + `unwired project got the re-deploy directive` |
| `MirrorCheckIsReadOnly` (AC-CMD-010) | `os.Remove(entryPath)` added on the dangling branch (repair-on-read) | **FAIL** — `the check mutated the skill trees` |

A first attempt at the AC-CMD-001 mutant leaked only *Detail* (empty for an absent mirror) and was
**not** caught; it is recorded because it shows the guard's exact boundary — it forbids mirror text
in either register, and an empty leak carries none. The mutant was sharpened to leak the finding
summary, which the guard caught. All three mutants were reverted; `grep -nE 'MUTANT|mutant'` over
both changed files returns no match, and the post-revert re-run restores the same 98 `--- PASS`
lines observed before the probes.

### §E.2 AC PASS/FAIL matrix

Every row cites the `--- PASS:` line that decided it (swept-count gate — an `ok` line or an exit
code alone satisfies none of them). No row is decided by a `--- SKIP:`; the indeterminate fixture
ran on darwin.

| AC | Status | Deciding evidence (actual output) |
|----|--------|-----------------------------------|
| AC-CMD-001 | PASS | `--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow (0.00s)` ×1 + mutant probe FAIL (above). Existing `--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent (0.00s)` unmodified and still passing. |
| AC-CMD-002 | PASS | `--- PASS: TestDoctorGolden_NoColor (0.00s)` (+ Light/Dark) and `git diff --name-only ace1c5440 -- internal/cli/testdata/` → no output. |
| AC-CMD-003 | PASS | `--- PASS: TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy (0.00s)` ×1 |
| AC-CMD-004 | PASS | `--- PASS: TestCheckCodexWiring_DanglingMirrorEntriesCounted (0.00s)` ×1 |
| AC-CMD-005 | PASS | `--- PASS: TestCheckCodexWiring_CopyModeDetailOnly (0.00s)` ×1 |
| AC-CMD-006 | PASS | `--- PASS: TestCheckCodexWiring_UnmirroredSkillsDetailOnly (0.00s)` ×1 |
| AC-CMD-007 | PASS | `--- PASS: TestCheckCodexWiring_UnwiredNoMirrorNag (0.00s)` ×1 + both sub-tests `--- PASS: …/no_mirror_directory`, `…/reportable_mirror_still_silent` + mutant probe FAIL |
| AC-CMD-008 | PASS | `--- PASS: TestCheckCodexWiring_MirrorUnreadableIndeterminate (0.00s)` ×1 — a PASS, not a SKIP; the symlink-loop control assertion held on darwin |
| AC-CMD-009 | PASS | `--- PASS: TestCheckCodexWiring_MirrorSummaryWidth (0.00s)` ×1 (+ `/absent`, `/dangling`) and `--- PASS: TestCheckCodexWiring_RenderedPanelStaysInBand (0.00s)` ×1 |
| AC-CMD-010 | PASS | `--- PASS: TestCheckCodexWiring_MirrorCheckIsReadOnly (0.00s)` ×1 + mutant probe FAIL |
| AC-CMD-011 | PASS | `go test ./internal/cli/... -count=1 -timeout 600s` EXIT=0 (17/17 ok); `go vet ./internal/cli/...` EXIT=0; `golangci-lint run ./internal/cli/...` EXIT=0 (`0 issues.`) |
| AC-CMD-012 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` EXIT=0 (no output) |
| AC-CMD-013 | PASS | `--- PASS: TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop (0.00s)` ×1 + both sub-tests. Residual risk settled by measurement — see below. |
| AC-CMD-014 | PASS | `--- PASS: TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters (0.00s)` ×1 and `--- PASS: TestDoctor_CheckCount (5.97s)` ×1 (no new top-level row) |
| AC-CMD-015 | PASS | subject `grep -nE 'os\.UserHomeDir\(\|exec\.LookPath\(\|os\.Getenv\("HOME"\)' internal/cli/doctor_codex_test.go` → no output (exit 1); control `grep -c … internal/cli/doctor.go` → `4` (pattern is live, so the empty subject result is not a broken pattern) |

### AC-CMD-013 residual risk — settled by measurement, not assumed

The plan audit judged the tail-drop main clause buildable on arithmetic (77 + 2 + 44 = 123) but
nobody had constructed the fixture. Measured on the emitted strings:

```
absent mirror summary   = 77 runes
stale-config summary    = 41 runes   (auditor estimated 44)
joined with "; "        = 120 runes  > 113 ceiling  → tail-drop branch reached
lead + overflow marker  = 102 runes  ≤ 113
```

The auditor's stale-summary estimate was 3 runes high; the conclusion is unaffected, and the branch
is reachable. The test does not rely on this arithmetic: it carries a **premise control** that
`t.Fatalf`s when `"see --verbose"` is absent, so a future string change that stops overflowing turns
the test red rather than letting it pass vacuously.

### Regression found and repaired during GREEN

The first full-package run after GREEN was **EXIT=1**: 16 pre-existing tests failed. This was a
genuine regression, not a fixture nuisance — `wireProjectForDoctor` builds a project that is wired
but carries no `.agents/skills`, which is precisely the state REQ-CMD-004 reports on, so every test
using that fixture inherited a mirror-absent finding.

Repair: `wireProjectForDoctor` now also creates an empty `.agents/skills` (a project with no skills
has nothing to mirror — it observes as neither a finding nor a detail count, and needs no symlink,
so no caller becomes windows-skipped). The former body is preserved verbatim as
`wireProjectWithoutMirror`, used by the five tests that need the mirror absent. No existing test
body or assertion was modified, and no guard was weakened: the 16 tests' subject is the home-layer
`skills.config` sub-check, which the fixture change does not touch.

Post-repair: `go test ./internal/cli/ -run 'TestCheckCodexWiring|TestCodexSkillPath|TestDoctor'
-count=1 -v` → EXIT=0, **98 `--- PASS`**, zero `--- FAIL`, zero `--- SKIP`.

### Gaps (explicitly NOT observed)

- **Windows runtime behaviour.** `GOOS=windows go build` passes, but a cross-build does not compile
  tests, so the copy-mode path and the symlink-skip branches were never *executed* on Windows.
  Consistent with root-cause.md Gaps and spec.md §6.
- **Full repository suite.** Only `./internal/cli/...` was run, per CLAUDE.local.md §4 load
  discipline. The full-suite verdict is CI's, on the pushed head.
- **Real Codex behaviour.** Nothing here observes how Codex CLI reacts to a dangling link or a
  partial mirror; every message is an action directive, never a claim about Codex.
- **The `.agents/skills` path literal.** The doctor hardcodes it (the producer's `mirrorSkillsRelDir`
  / `mirrorLinkTarget` are unexported in package `template`). The drift vulnerability is left
  intact per the scope fence and is recorded in a source comment as a candidate for its own card.
- **No mirror was created anywhere.** Confirmed by AC-CMD-010's tree listing and by the changed-file
  set; `.agents/` remains absent in this tree.

### Residual risk

- An empty `.agents/skills` directory is silent by design (no REQ covers "present but empty"), so a
  project whose mirror exists but holds nothing gets no signal. This follows the specified
  requirement set; whether it *should* warn is a scope question, not a defect in this implementation.
- A regular **file** occupying a mirror path is counted as neither copy-mode nor dangling. No
  requirement classifies that shape, and counting it would assert something the check cannot support.
- The `wireProjectForDoctor` fixture change alters the baseline every future test in this file
  inherits. A future test that wants a mirror-absent wired project must use
  `wireProjectWithoutMirror`; using the wrong helper produces a confusing extra finding rather than
  a build error.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: pending-backfill        # a commit cannot cite its own hash
run_status: audit-ready
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0         # internal/cli/update.go + internal/template/skill_mirror.go unmodified
l44_pre_commit_fetch: not-run           # lane does not push; no remote interaction this phase
l44_post_push_fetch: not-run            # push is the lead's batch act (gitflow-lane-protocol §4)
new_warnings_or_lints_introduced: 0     # golangci-lint "0 issues.", go vet exit 0
cross_platform_build:
  darwin_arm64: pass                    # go test ./internal/cli/... exit 0
  windows_amd64: pass                   # GOOS=windows GOARCH=amd64 go build ./... exit 0
  note: "cross-build does not compile tests; windows runtime path unexercised"
total_run_phase_files: 3                # doctor_codex.go, doctor_codex_test.go, progress.md (+ spec.md status transition)
m1_to_mN_commit_strategy: single-commit  # M1-M5 landed as one commit; no push (lane discipline)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: b921184b7c395cdc1831253e9b492712da19c3cc   # backfilled; the sync commit cannot cite its own hash
sync_status: audit-ready
b12_self_test_a: pass                    # grep -c 'SPEC-CODEX-MIRROR-DOCTOR-001' CHANGELOG.md → 0 before emission (no duplicate)
b12_self_test_b: pass                    # 15 distinct AC ids in acceptance.md == 15 rows in the §E.2 matrix == 15 referenced in the entry (non-zero, not a vacuous 0==0)
b12_self_test_c: pass                    # every path named in the CHANGELOG entry verified present with ls (8/8 OK)
changelog_entry_position: "[Unreleased] › Added, first bullet"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (status + updated only)"
  plan.md: "n/a — carries no status: field (grep -c '^status:' → 0)"
  acceptance.md: "n/a — carries no status: field (grep -c '^status:' → 0)"
  progress.md: "n/a — no frontmatter block"
canary_compliance_check: n/a             # this SPEC declares no forward-looking policy for its own sync to test
mx_tag_validation:
  doctor_codex_go_mx_count_at_base: 0    # grep -c '@MX' on git show ace1c5440:internal/cli/doctor_codex.go
  doctor_codex_go_mx_count_now: 0        # grep -c '@MX' on the working copy
  verdict: "no MX regression — the file has never carried an @MX tag; this card neither added nor removed one"
  indicated_but_not_applied: "an @MX:WARN on the path-literal drift (doctor_codex.go:62-69) is indicated; NOT applied — the sync dispatch fences Go source as out of scope, and the risk is already carried in a source comment and in the follow-up-card record"
  mandatory_anchors_owed: 0              # inspectSkillMirror and codexMirrorObservations each have exactly 1 non-test caller (fan_in 1 < 3)
docs_site_4_locale_sync: not-required    # judgment recorded below
files_changed_this_phase: 2              # CHANGELOG.md, progress.md (+ spec.md frontmatter transition)
```

### Sync-phase judgment — docs-site (adk.mo.ai.kr)

**Not required, and deliberately not started.** This card adds an advisory observation to a doctor
row that already exists (`Codex Wiring`); it introduces no command, no flag, no config key, and no
workflow the published guides describe. A user's interaction surface is unchanged — they run
`moai doctor` exactly as before and may now see one more line in it. A grep for the mirror's own
path across the published docs corpus returns nothing (`grep -rn '\.agents/' --include='*.md'
.claude/rules .moai/docs CLAUDE.md CLAUDE.local.md AGENTS.md` → no matches, root-cause.md Claim 3),
so there is no existing page whose text this change makes wrong. Were a page judged necessary, the
4-locale obligation (ko/en/ja/zh, same PR) makes it its own card rather than a tail on this one —
so nothing was started here.

### Carried forward from §E.2 — three facts the close does not smooth away

These are recorded verbatim in §E.2 above and are cited, not re-summarized, so that neither the
CHANGELOG entry nor this signal weakens them:

1. **A real regression was caused and repaired at the fixture** (§E.2 "Regression found and repaired
   during GREEN"). `wireProjectForDoctor` produced a wired project with no `.agents/skills` — the
   exact state REQ-CMD-004 reports — so 16 existing tests inherited a mirror-absent finding. The
   repair created an empty `.agents/skills` in the helper and preserved the former body verbatim as
   `wireProjectWithoutMirror`; no existing assertion was edited.
2. **RED-now was insufficient for the three absence guards, and mutation testing supplied what it
   could not** (§E.2 "Mutation probes"). AC-CMD-001 / 007 / 010 assert that text or state does *not*
   appear, which an absent feature satisfies trivially. The first AC-CMD-001 mutant leaked into
   `Detail` — empty in the mirror-absent state — and was **not** caught; that miss is a fact about
   the guard's boundary and stays in the record rather than being replaced by the sharpened mutant
   that did catch.
3. **The `.agents/skills` path literal is hardcoded in the doctor while the producer's
   `mirrorSkillsRelDir` is unexported** (`internal/cli/doctor_codex.go:62-69`). If the producer
   moves its layout, the doctor silently emits a *wrong* "mirror absent" finding. Per lead
   direction this vulnerability is stated explicitly rather than fixed here, and is a candidate for
   its own card.

### Unverified at close (carried forward, not narrowed)

- **Windows runtime behaviour.** `GOOS=windows GOARCH=amd64 go build ./...` exits 0, but a
  cross-build compiles no tests, so the copy-mode path and the symlink-skip branches were never
  *executed* on Windows.
- **The full repository suite.** Only `./internal/cli/...` was run, per the lane's load discipline.
  The full-suite verdict is CI's, on the integrated head — not this lane's to assert.
- **Codex CLI's actual reaction to a dangling or partial mirror.** Nothing here observes it; every
  message the check emits is an action directive, never a claim about Codex.

### Section-numbering note

No `## §E.5` section is authored. `spec-frontmatter-schema.md` § progress.md Section Map records
§E.5 as **RETIRED** (folded into §E.4; retained in `era.go` only for backward-compat classification
of pre-redesign SPECs), with an explicit "do NOT author new §E.5 sections". MX Tag validation is
therefore folded into the `mx_tag_validation` block above, as the 3-phase close prescribes.
