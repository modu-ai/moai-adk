# SPEC-UPDATE-LEGACY-SKILL-LIST-001 — Acceptance Criteria

## §A Verification discipline

Every command below was **executed at authoring time** against `main` @ `9a6b6c854` and its output recorded as the pre-fix baseline. An AC whose baseline was not observed is not an AC.

Four rules bind every criterion here:

1. **No vacuous `-run` selectors.** A `go test -run <pattern>` naming a nonexistent test exits 0 and passes silently. Every test-based AC therefore (a) greps that the test name exists in the tree, (b) runs with `-count=1`, and (c) asserts on the observed `--- PASS` line, not merely on the exit code.
2. **Token presence is not reachability.** An AC asserting that a guard exists must also show the guard runs and that reverting the fix turns it red. §C carries the falsification procedure.
3. **Greps are scoped precisely.** Each grep names the exact path set and, where a window is used, a window that provably contains its own target line. The `legacySkillIDs` block is addressed by a content-anchored `awk` range rather than by line numbers, which drift.

3a. **[HARD] Anchoring rule for count-based criteria over `go test` output.** Go's `-v` output emits a result line for the parent test at column 0 and an **indented** result line (four leading spaces) for every subtest. An unanchored `grep -c -- '--- PASS: TestX'` therefore counts parent **plus** every passing subtest of `TestX`. Any criterion whose EXPECT is a **specific non-zero count** MUST anchor its pattern so the count is determined by the pattern, not by how many subtests the implementation happens to contain:

   | Counting | Anchor | Rationale |
   |----------|--------|-----------|
   | A parent result line, EXPECT ≥ 1 | `grep -cE '^--- PASS: TestX'` | `^` excludes indented subtest lines; the count is 1 regardless of subtest count |
   | A specific subtest result line, EXPECT ≥ 1 | `grep -cE '^    --- PASS: TestX/subname'` | Four-space indent pins it to a subtest and excludes the parent |
   | Any failure at all, **EXPECT 0** | `grep -c -- '--- FAIL'` — deliberately UNANCHORED | For a zero-assertion, unanchored is strictly **more sensitive**: it catches a failure at parent or subtest depth. Anchoring here would narrow detection for no benefit. |

   The rule is therefore **anchor when counting a specific non-zero number; leave unanchored when asserting zero** — not "anchor everything". A blanket anchor would weaken the EXPECT-0 forms.

   Any new AC that counts `go test` output MUST state which row of the table above it falls under, and MUST have had its EXPECT produced by an actual run against the test shape the SPEC pins (plan.md §E M2 / §E M4). This rule exists because the defect it prevents recurred once: iteration 1 fixed the unanchored parent-count in the guard ACs, and iteration 2 found the identical trap re-created between AC-LSL-010(b) and the newly-added AC-LSL-016(b).
4. **Baselines are stated, not assumed.** Each AC records the value observed pre-fix so the run-phase delta is checkable rather than asserted.
5. **Exit codes beat greps for whole-suite checks.** A pattern-based suite check cannot see a test-file compile error (`go test` emits a TAB-separated `FAIL\t<pkg>\t[build failed]` and no `--- ` line), so any AC asserting "the suite is green" asserts on the process exit code, with the grep kept only as a readability aid.
6. **Falsification never writes to the working tree.** See §C.0 — the mechanism is `go test -overlay`, and every falsification carries a fix-present post-condition, because a clean tree is also what a destroyed fix produces.

> **A measurement note that cost two contradictory readings during authoring.** This shell aliases `ls` to long format, so `ls <dir> | grep -E '^name'` matches nothing — every line begins with `-rw-r--r--@`. Use a glob or `find … -exec basename` when checking for a file's existence by name prefix. Two independent reviews of this SPEC reached opposite conclusions about whether a set of report files existed for exactly this reason.

Notation: `BASELINE (2026-07-31, 9a6b6c854)` marks an observed pre-fix value. The block extractor used repeatedly below is:

```bash
LIST_BLOCK='awk "/^var legacySkillIDs = \\[\\]string\\{/{f=1;next} f&&/^\\}/{f=0} f" internal/cli/update_archive.go'
```

which prints only the lines between the `var legacySkillIDs = []string{` opener and its closing brace.

---

## §B Acceptance criteria

### AC-LSL-001 — The list holds exactly the 13 genuinely-removed IDs (REQ-LSL-001, REQ-LSL-002)

```bash
# (a) entry count
awk '/^var legacySkillIDs = \[\]string\{/{f=1;next} f&&/^\}/{f=0} f' \
  internal/cli/update_archive.go | grep -c '"moai-'
# EXPECT: 13
# BASELINE (2026-07-31, 9a6b6c854): 16

# (b) the 13 expected IDs are all present, in order
awk '/^var legacySkillIDs = \[\]string\{/{f=1;next} f&&/^\}/{f=0} f' \
  internal/cli/update_archive.go | sed -E 's/.*"(moai-[^"]*)".*/\1/' | tr -d '\t'
# EXPECT, exactly and in this order:
#   moai-domain-db-docs
#   moai-domain-mobile
#   moai-framework-electron
#   moai-library-shadcn
#   moai-library-mermaid
#   moai-library-nextra
#   moai-tool-ast-grep
#   moai-platform-auth
#   moai-platform-deployment
#   moai-platform-chrome-extension
#   moai-workflow-research
#   moai-workflow-pencil-integration
#   moai-formats-data
# BASELINE: the same 13, preceded by moai-domain-backend, moai-domain-frontend,
#           moai-domain-database (16 lines total).
```

Why the block extractor and not a whole-file grep: the file's header comment and the `@MX:NOTE` blocks also mention skill IDs, so a file-wide `grep -c '"moai-'` would over-count. Scoping to the literal slice body is the claim's actual subject.

---

### AC-LSL-002 — The three live skills are absent from the list (REQ-LSL-003)

```bash
awk '/^var legacySkillIDs = \[\]string\{/{f=1;next} f&&/^\}/{f=0} f' \
  internal/cli/update_archive.go \
  | grep -cE '"(moai-domain-backend|moai-domain-frontend|moai-domain-database)"'
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 3
```

Companion assertion — the three must still exist as live template skills, otherwise this AC could be satisfied by the wrong fix (deleting the skills instead of correcting the list):

```bash
for s in moai-domain-backend moai-domain-frontend moai-domain-database; do
  test -f "internal/template/templates/.claude/skills/$s/SKILL.md" || echo "MISSING $s"
done
# EXPECT: no output
# BASELINE: no output (all three present)
```

---

### AC-LSL-003 — The doc comment states the corrected count and the reason (REQ-LSL-004)

```bash
# (a) the stale "16 skill IDs" wording is gone
grep -c 'lists the 16 skill IDs removed in BC-V3R3-007' internal/cli/update_archive.go
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 1

# (b) the comment above the slice names the corrected count and the revival commit
awk '/^\/\/ legacySkillIDs/,/^var legacySkillIDs/' internal/cli/update_archive.go
# EXPECT: text containing "13" and the commit id 697a6e2c7
# BASELINE: "legacySkillIDs lists the 16 skill IDs removed in BC-V3R3-007." — no
#           mention of 13, no mention of the revival.
```

---

### AC-LSL-004 — The cross-check guard exists, runs, and passes (REQ-LSL-005, REQ-LSL-009)

```bash
# (a) the test name exists — without this the -run below is vacuous
grep -rn 'func TestLegacySkillIDsNotEmbedded' internal/cli/ | wc -l | tr -d ' '
# EXPECT: 1
# BASELINE (2026-07-31, 9a6b6c854): 0   ← the -run in (b) would have matched
#                                          nothing and exited 0 vacuously

# (b) it passes, asserted on the PARENT PASS line rather than on the exit code.
#     The ^ anchor is load-bearing: Go's -v output indents subtest result lines
#     ("    --- PASS: Test.../production"), so an unanchored grep counts the
#     parent AND every passing subtest. Measured on a probe carrying one passing
#     subtest plus two skips: unanchored = 2, anchored = 1.
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 \
  | grep -cE '^--- PASS: TestLegacySkillIDsNotEmbedded'
# EXPECT: 1
# BASELINE: 0

# (c) the guard derives its live set from the embedded manifest, not a literal
grep -c 'template.EmbeddedMoaiSkillNames' internal/cli/update_archive_guard_test.go
# EXPECT: >= 1
# BASELINE: file does not exist
```

Negative half of (c) — **[NON-BINARY — reviewer judgment]** the guard must not smuggle in a second inventory copy. The grep is mechanical but its EXPECT carries a reviewer clause, so it is labelled non-binary rather than presented as a gate:

```bash
grep -cE '"moai-(domain|library|platform|framework|tool|workflow|formats)-' \
  internal/cli/update_archive_guard_test.go
# EXPECT: 0 for the production assertion path. A synthetic ID inside a
#         degradation subtest fixture is permitted only if it is NOT a real
#         template skill name; if any literal appears, the reviewer must
#         confirm it is fixture-only.
# BASELINE: file does not exist
```

---

### AC-LSL-005 — The guard is demonstrated to catch this exact defect (REQ-LSL-006, REQ-LSL-008)

Falsification, not a pass observation. Full procedure in §C.1; the criterion is:

```bash
# Under the §C.1 overlay that re-adds "moai-domain-backend" to legacySkillIDs
# (the working tree is NOT modified — see §C.1):
go test -overlay="$OVERLAY_JSON" ./internal/cli/ \
  -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 | tee "$FALSIFY_LOG" >/dev/null

grep -cE '^--- FAIL: TestLegacySkillIDsNotEmbedded' "$FALSIFY_LOG"
# EXPECT: 1
# (anchored for the same reason as AC-LSL-004(b): measured on a probe,
#  unanchored = 2 when a subtest also fails, anchored = 1)

# and the failure message names the offending ID (REQ-LSL-006)
grep -c 'moai-domain-backend' "$FALSIFY_LOG"
# EXPECT: >= 1
```

BASELINE (2026-07-31, `9a6b6c854`): the guard does not exist, so neither command can produce a `--- FAIL` line. The measured intersection on today's tree was obtained instead with a temporary in-package probe, which reported `legacySkillIDs=16 embedded=30 overlap=3 [moai-domain-backend moai-domain-frontend moai-domain-database]`; the probe file was removed and `git status --porcelain` confirmed the tree returned to its prior state. The `-overlay` falsification mechanism itself was separately proven on a throwaway module: with the overlay applied the guard printed `--- FAIL` and named the injected ID, and `grep -c '<injected>' <source>` on the real file returned `0` — the working tree was provably untouched.

An AC-LSL-004 pass without an AC-LSL-005 pass is not acceptance. A guard observed only green has not been shown to be capable of red.

---

### AC-LSL-006 — The six pre-existing test files still pass, and the `All16` naming decision is recorded (parts a-b → REQ-LSL-018; part c → REQ-LSL-019)

```bash
# (a) all six files' tests. TWO assertions, because a zero-FAIL count alone is
#     vacuous if the -run regex selects nothing (it then exits 0 with no output).
#     §A rule 1 mandates an existence guard; this is the whole-suite AC, so the
#     omission would be the most expensive place to leave one.
go test ./internal/cli/ -count=1 -v \
  -run 'TestArchiveSkill_|TestArchiveLegacySkills_|TestArchiveForce|TestArchiveIdempotency|TestRestoreSkill_|TestSkipSyncNoArchive' \
  > /tmp/lsl-ac006.log 2>&1; echo "exit=$?"

#   (a-i) EXISTENCE GUARD — the selector actually selected the expected tests.
#         §A rule 3a row 1 (parent lines, non-zero EXPECT → ^-anchored). The
#         top-level test count does NOT vary with list length (only subtest
#         counts do), so this number is stable across M1's 16→13 shrink.
grep -cE '^--- PASS' /tmp/lsl-ac006.log
# EXPECT: 19
# BASELINE (2026-07-31, 9a6b6c854): 19 (exit=0; 61 total PASS+SKIP lines incl. subtests)
# CONSTRAINT: AC-LSL-006(c) permits renaming the `…_All16…` tests, but any rename
#   MUST preserve the `TestArchiveSkill_` / `TestRestoreSkill_` prefixes the
#   selector matches on. A rename that drops the prefix reduces this count and
#   fails here — which is the intended detection, not a false positive.

#   (a-ii) zero failures. §A rule 3a row 3 (EXPECT 0 → deliberately UNANCHORED,
#          so a subtest-depth failure is caught as well as a parent-depth one).
grep -c -- '--- FAIL' /tmp/lsl-ac006.log
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 0

# (b) positional accesses remain in range at 13 entries
grep -nE 'legacySkillIDs\[[0-9]+\]|legacySkillIDs\[:[0-9]+\]' internal/cli/*_test.go
# EXPECT: every index < 13 and every slice bound <= 13
# BASELINE: [0] (force, skip_sync), [1], [2], [3] (force), [:5] (flow), [:8] (idempotency)
#           → max index 3, max bound 8. All in range at 13.

# (c) the All16 naming decision is observable either way (REQ-LSL-019)
grep -c 'All16' internal/cli/update_archive_test.go internal/cli/migrate_restore_skill_test.go
# EXPECT: either 0 for both files (renamed) or unchanged at 2 and 2 (kept).
#         A mixed result — one file renamed, the other not — FAILS this AC.
# BASELINE (2026-07-31, 9a6b6c854):
#           internal/cli/update_archive_test.go:2
#           internal/cli/migrate_restore_skill_test.go:2
# A rename MUST preserve the TestArchiveSkill_ / TestRestoreSkill_ prefixes
# (REQ-LSL-019) — part (a-i)'s count of 19 detects a prefix-dropping rename.
```

Part (c) is deliberately permissive on the outcome and strict on consistency: the SPEC does not mandate the rename, but it does forbid a half-applied one.

---

### AC-LSL-007 — The guard skips, rather than passing or failing, when the manifest is unavailable (REQ-LSL-007)

```bash
# (a) both degradation subtests exist
grep -cE 'manifest_error|manifest_empty' internal/cli/update_archive_guard_test.go
# EXPECT: >= 2
# BASELINE: file does not exist

# (b) both SKIP — asserted on the SKIP marker, since a skipped test also
#     produces a green package result and would be invisible to exit code alone.
#     §A rule 3a row 2 (specific subtest lines, non-zero EXPECT → indent-anchored),
#     which pins the count to this test's own subtests and cannot drift if another
#     selected test ever skips. Measured under the pinned shape: unanchored 2,
#     anchored 2 — the anchor changes no value today and makes the 2 deterministic.
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 \
  | grep -cE '^    --- SKIP: TestLegacySkillIDsNotEmbedded/'
# EXPECT: 2
# BASELINE: 0

# (c) neither degradation path is a silent pass
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 \
  | grep -E -- '--- (PASS|SKIP|FAIL)'
# EXPECT: the two degradation subtests appear as SKIP, never as PASS
# BASELINE: no output
```

Rationale for asserting on `--- SKIP` and not on the package result: a subtest that silently returned without asserting would also leave the package green. Only the explicit marker distinguishes "declined to judge" from "judged and found fine".

---

### AC-LSL-008 — The three wrong archive files are no longer tracked (REQ-LSL-010)

```bash
git ls-files .moai/archive/skills/v2.16/moai-domain-backend/ \
             .moai/archive/skills/v2.16/moai-domain-frontend/ \
             .moai/archive/skills/v2.16/moai-domain-database/ | wc -l | tr -d ' '
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 3
```

---

### AC-LSL-009 — The four genuine archive entries are preserved, byte-unchanged (REQ-LSL-011, REQ-LSL-012)

```bash
# (a) still tracked, still four
git ls-files .moai/archive/skills/v2.16/ | wc -l | tr -d ' '
# EXPECT: 4
# BASELINE (2026-07-31, 9a6b6c854): 7

# (b) they are the four expected ones
git ls-files .moai/archive/skills/v2.16/ | cut -d/ -f5 | sort
# EXPECT exactly:
#   moai-framework-electron
#   moai-platform-auth
#   moai-platform-chrome-extension
#   moai-platform-deployment
# BASELINE: the same four, plus moai-domain-backend, moai-domain-database,
#           moai-domain-frontend (7 lines).

# (c) byte-unchanged — no content edit smuggled in with the deletions
git diff --stat 9a6b6c854 -- \
  .moai/archive/skills/v2.16/moai-framework-electron/ \
  .moai/archive/skills/v2.16/moai-platform-auth/ \
  .moai/archive/skills/v2.16/moai-platform-chrome-extension/ \
  .moai/archive/skills/v2.16/moai-platform-deployment/ | wc -l | tr -d ' '
# EXPECT: 0
# BASELINE: 0 (no diff against itself)

# (d) no .gitignore rule was added — §A D5 keeps that out of scope.
#     Pattern scoped to a path-anchored archive rule so an unrelated future line
#     containing the substring "archive" (e.g. "*.tar.archive") does not trip it.
#     NOTE: grep -c exits 1 on zero matches, so this must NOT sit in a `set -e`
#     script or a `&&` chain — capture the count, then compare.
n=$(grep -cE '^\.?moai/archive' .gitignore); echo "$n"
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 0 (grep exit 1, no match)
```

Part (d) is a negative AC guarding a deliberate omission: without it, an implementer "helpfully" adding `.moai/archive/` to `.gitignore` would leave the four genuine files tracked-but-ignored, a state §A D5 explicitly declines.

---

### AC-LSL-010 — One failing entry no longer aborts the remaining entries (REQ-LSL-013, REQ-LSL-016)

```bash
# (a) the test name exists
grep -rn 'func TestArchiveLegacySkills_ContinuesAfterFailure' internal/cli/ | wc -l | tr -d ' '
# EXPECT: 1
# BASELINE (2026-07-31, 9a6b6c854): 0

# (b) it passes — §A rule 3a row 1 (parent result line, non-zero EXPECT → ^-anchored).
#     AC-LSL-016(b) mandates a `success_count_excludes_failures` subtest, so the
#     unanchored form counts parent + subtest. Measured under that exact shape:
#     unanchored = 2, anchored = 1.
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v 2>&1 \
  | grep -cE '^--- PASS: TestArchiveLegacySkills_ContinuesAfterFailure'
# EXPECT: 1
# BASELINE: 0
```

The success-only count obligation (REQ-LSL-016) is NOT left to a reviewer read — it has its own mechanical criterion, AC-LSL-016.

The remaining obligations are structural properties of the fixture, checkable by reading the test. They are explicitly labelled non-binary so they are not mistaken for criteria:

- the fixture seeds at least two archivable skills and makes the *first* one fail, so "continued" is distinguishable from "the failure happened to be last";
- it asserts the later skill's archive directory exists after the call;
- the failure is injected portably. Preferred: seed the destination path of the failing ID as a regular file so `MkdirAll` fails with an `ENOTDIR`-class error on every OS. A `chmod 0o000` fixture is acceptable only with an explicit `runtime.GOOS == "windows"` skip and a root-user caveat, because `chmod` is a no-op for root.

The first bullet is partially mechanized by AC-LSL-016(b): a success-count subtest that passes on a one-failure fixture can only do so if the loop actually continued past the failure.

---

### AC-LSL-011 — The `total:` summary is emitted even when an entry failed (REQ-LSL-014)

Asserted inside the same test, on captured output:

```bash
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v 2>&1 \
  | grep -c -- '--- FAIL'
# EXPECT: 0

# and the assertion itself is present in the test source.
# The pattern is UNQUOTED on purpose: a `"total:"` literal would require a
# closing quote immediately after the colon and would reject the natural
# assertion forms `strings.Contains(output, "total: ")` or `"total: 2 skills"`.
grep -cF 'total:' internal/cli/update_archive_continue_test.go
# EXPECT: >= 1
# BASELINE: file does not exist
```

Pre-fix behaviour for contrast, from reading `archiveLegacySkills`: the three in-loop `return archived, fmt.Errorf(...)` sites all return **before** the `fmt.Fprintln(out, tui.Pill(... "total: %d skills archived" ...))` call that follows the loop, so a failure suppresses the summary entirely.

---

### AC-LSL-012 — The aggregate error names every failure, and is `nil` on a clean pass (REQ-LSL-015)

```bash
# (a) accumulation replaces the in-loop returns — all three of them
awk '/^func archiveLegacySkills/{f=1} f&&/^\}/{exit} f&&/return archived, fmt\.Errorf/{c++} END{print c+0}' \
  internal/cli/update_archive.go
# EXPECT: 0
# BASELINE (2026-07-31, 9a6b6c854): 3
#   (lines 302 "create drift backup parent for %s", 305 "backup drift archive
#    for %s", 320 "archive %s" — all inside the loop, all abortive)

# (b) errors.Join is used, preserving each %w chain
grep -c 'errors.Join' internal/cli/update_archive.go
# EXPECT: >= 1
# BASELINE: 0

# (c) the multi-failure case names both IDs — asserted in the test
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v 2>&1 \
  | grep -c -- '--- FAIL'
# EXPECT: 0, with the test itself asserting the returned error string contains
#         each failing ID, and asserting a nil return on the all-success fixture.
```

Part (a)'s `awk` window is anchored on the function opener and terminates at the first column-0 `}`, which is the function's own closing brace — so the window provably contains the three target lines and nothing from neighbouring functions. The `END{print c+0}` form prints `0` rather than an empty line when there are no matches, so a zero result is distinguishable from a broken command.

---

### AC-LSL-013 — The output keyword contract is preserved (REQ-LSL-017)

```bash
grep -c '"archive: "' internal/cli/update_archive.go
# EXPECT: 1
# BASELINE (2026-07-31, 9a6b6c854): 1

grep -c 'total: %d skills archived' internal/cli/update_archive.go
# EXPECT: 2
# BASELINE (2026-07-31, 9a6b6c854): 2
#   (one in archiveLegacySkills, one in dryRunArchiveLegacySkills)
```

The existing contract test remains the binding check — `update_archive_flow_test.go` asserts `strings.Contains(output, "archive: "+id)` per skill and `strings.Contains(output, "total:")` — and is covered by AC-LSL-006(a). The greps above are the cheap early-warning, not the authority.

---

### AC-LSL-014 — Build, suite, formatting, and isolation (NFR-LSL-001 … NFR-LSL-005)

```bash
# (a) build + vet
go build ./... && go vet ./internal/cli/ ./internal/template/
# EXPECT: exit 0, no output
# BASELINE (2026-07-31, 9a6b6c854): exit 0, no output

# (b) full package suites — asserted on the EXIT CODE, plus an explicit
#     build-failure marker check. A grep for '^(FAIL|---) ' is blind to a
#     test-file compile error: Go emits "FAIL\t<pkg>\t[build failed]" (TAB, not
#     space) and no "--- " line at all, so the grep returns 0 and passes
#     vacuously. Measured on a throwaway module with one broken _test.go:
#       go build ./...  → exit 0   (go build does not compile _test.go)
#       go test  ./...  → exit 1, "FAIL\tvac/pkg [build failed]"
#       grep -c -E '^(FAIL|---) ' → 0   ← the retired form certified nothing
#     This SPEC adds two NEW test files to internal/cli, so this is a live hazard.
#     NOTE: both greps below expect zero matches, and `grep -c` exits 1 on zero
#     matches even though it prints `0`. The counts are captured into variables
#     so this AC cannot abort a `set -e` script or a `&&` chain — the same hazard
#     AC-LSL-009(d) documents.
go test ./internal/cli/... ./internal/template/... -count=1 \
  > /tmp/lsl-suite.log 2>&1; echo "exit=$?"
bf=$(grep -c 'build failed' /tmp/lsl-suite.log); echo "build_failed=$bf"
tf=$(grep -c -- '--- FAIL' /tmp/lsl-suite.log); echo "test_failed=$tf"
# EXPECT: exit=0, build_failed=0, test_failed=0
# BASELINE (2026-07-31, 9a6b6c854): exit=0, 0, 0
# (test_failed is §A rule 3a row 3 — EXPECT 0, deliberately unanchored.)
#   (all packages "ok"; internal/template/scripts has no test files)
# NOTE: AC-LSL-014(a)'s `go build ./...` does NOT cover this — it skips _test.go.

# (c) gofmt on the files this SPEC touches (NFR-LSL-004)
gofmt -l internal/cli/update_archive.go internal/cli/update_archive_guard_test.go \
         internal/cli/update_archive_continue_test.go
# EXPECT: no output
# BASELINE: internal/cli/update_archive.go IS listed — a pre-existing single-hunk
#           import-order deviation (internal/tui sorted before
#           internal/cli/update/backup). Since M1 and M4 both edit this file,
#           gofmt-cleaning it is in scope and this AC requires the file to be
#           clean afterwards. The two new test files do not exist at baseline.

# (d) no template mutation (NFR-LSL-002)
git status --porcelain internal/template/templates/ | wc -l | tr -d ' '
# EXPECT: 0
# BASELINE: 0

# (e) test isolation — every new fixture uses t.TempDir() (NFR-LSL-001)
grep -c 'os.MkdirAll("/\|os.WriteFile("/\|filepath.Join("/' \
  internal/cli/update_archive_guard_test.go internal/cli/update_archive_continue_test.go
# EXPECT: 0 for both (no absolute-path fixture roots)
# BASELINE: files do not exist

grep -c 't.TempDir()' internal/cli/update_archive_continue_test.go
# EXPECT: >= 1
# BASELINE: file does not exist
```

---

### AC-LSL-015 — `archiveVersion` and the archive scheme are unchanged (NFR-LSL-003, spec.md §B Goal 5)

```bash
grep -c 'archiveVersion = "v2.16"' internal/cli/update_archive.go
# EXPECT: 1
# BASELINE (2026-07-31, 9a6b6c854): 1

# (b) no alternative archive root was introduced — the directory scheme is
#     pinned by count, not by prose.
grep -cE '"\.moai", "archive", "skills"' internal/cli/update_archive.go
# EXPECT: 6
# BASELINE (2026-07-31, 9a6b6c854): 6 — lines 73, 287, 298, 311, 328, 347:
#   73  archiveSkill               dstDir
#   287 archiveLegacySkills        dstDir
#   298 archiveLegacySkills        backupDir      (force drift-backup path)
#   311 archiveLegacySkills        archiveBackupRel (force drift-backup path)
#   328 archiveLegacySkills        archiveDst
#   347 dryRunArchiveLegacySkills  archiveDst
# The v0.2.0 text read "EXPECT: unchanged from baseline" with a prose baseline
# naming three regions. That violated §A rule 4 (no number) and undercounted
# (six lines, four regions — dryRunArchiveLegacySkills was unnamed). Both are
# corrected here against a measured value.
```

M4 refactors `archiveLegacySkills`, which owns four of the six sites (287, 298, 311, 328). A refactor that consolidates two joins into one, or introduces a helper, changes this count legitimately — in that case the implementer updates the EXPECT **and records the new measured value with its line list**, exactly as this baseline does. What the AC forbids is the count changing silently.

Why this AC exists separately from AC-LSL-008/009: those two exercise `.moai/archive/skills/v2.16/` as **git paths**, which a change to the Go constant would not disturb — the old directory would simply stop being written to while the tracked files sat unchanged. Only a source-level pin catches that. M4 is the only milestone that edits the file the constant lives in, which is where the risk sits.

---

### AC-LSL-016 — The reported count counts successes only (REQ-LSL-016)

`REQ-LSL-016` was previously carried only by a reviewer-read obligation under AC-LSL-010 ("checkable by reading it"). A reviewer-read obligation is not a binary criterion, so it gets its own mechanical check:

```bash
# (a) [NON-BINARY — reviewer judgment] the count assertion exists in the test
#     source and is not a bare len(legacySkillIDs) comparison.
grep -nE 'archived[^=]*==|want.*archived|archived.*want' \
  internal/cli/update_archive_continue_test.go
# EXPECT: at least one matching line (mechanical), AND a reviewer confirms it
#         compares against the number of SUCCESSFUL entries (e.g.
#         len(legacySkillIDs)-1 for a one-failure fixture), NOT against
#         len(legacySkillIDs). The grep cannot make that distinction.
# BASELINE: file does not exist
# The binary half of REQ-LSL-016 is part (b) below; (a) is the supplementary
# source-shape read and is labelled non-binary so it is not mistaken for a gate.

# (b) the assertion is exercised — the subtest carrying it reports PASS
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' \
  -count=1 -v 2>&1 | grep -cE '^    --- PASS: TestArchiveLegacySkills_ContinuesAfterFailure/success_count_excludes_failures'
# EXPECT: 1
# BASELINE: 0
```

Part (b) requires the success-count check to live in a **named subtest** (`success_count_excludes_failures`) so that its execution is independently observable rather than buried in the parent body. The four-space indent in the anchor is the subtest indentation Go's `-v` output emits.

---

## §C Falsification procedure

An AC that has only been observed green proves the command runs, not that it discriminates. Each item below turns a green check red on purpose.

### C.0 The falsification mechanism: `go test -overlay` (working tree is never modified)

[HARD] **No falsification step in this SPEC may modify a tracked file.** The obvious form — edit the source, observe red, `git checkout -- <path>` — is prohibited on two independent grounds:

1. **It can destroy the fix it is validating.** `internal/cli/update_archive.go` carries **both** the M1 list correction and the M4 loop refactor. `git checkout -- <path>` restores from the index, so if either milestone's work is unstaged when the revert runs, the whole fix is silently discarded. A `git status --porcelain` post-condition cannot detect this: a full revert produces exactly the clean tree the check expects. The verification step would report success in the one case where it most needs to report failure.
2. **It is on the repository's forbidden-command list.** `.claude/rules/moai/workflow/main-checkout-branch-guard.md` names `git checkout -- <path>` in its forbidden table for the primary checkout ("discards work the orchestrator cannot see the provenance of"), and this checkout is shared (progress.md §E.1 records concurrent-session discipline). `git restore --source=<sha> -- <path>` is the same class of working-tree write and is not an adequate substitute.

The mechanism used instead is Go's `-overlay` flag, which substitutes file content **at compile time only**:

```bash
# Setup — run once per falsification. Nothing under the repo is written.
FIX_SRC=internal/cli/update_archive.go
WORK=$(mktemp -d)
OVERLAY_JSON="$WORK/overlay.json"
FALSIFY_LOG="$WORK/falsify.log"

# Produce the MUTATED copy outside the repo (edit as each §C item directs).
sed 's|<the mutation>|<the replacement>|' "$FIX_SRC" > "$WORK/mutated.go"

printf '{"Replace": {"%s": "%s"}}\n' \
  "$(cd "$(dirname "$FIX_SRC")" && pwd)/$(basename "$FIX_SRC")" \
  "$WORK/mutated.go" > "$OVERLAY_JSON"

# Run against the mutation.
go test -overlay="$OVERLAY_JSON" ./internal/cli/ -run '<TestName>' -count=1 -v \
  2>&1 | tee "$FALSIFY_LOG"
```

**Mandatory post-conditions after every §C item** — these are what make the procedure discriminating. A clean tree alone is NOT evidence, since a destroyed fix also yields a clean tree:

```bash
# P1. The working tree file was never touched.
git diff --quiet -- internal/cli/update_archive.go; echo "unmodified_exit=$?"
# EXPECT: unmodified_exit=0 relative to whatever M1/M4 committed or staged.
#         (Under overlay this is guaranteed by construction; the check is the
#          proof, not the mechanism.)

# P2. The fix is still present.
awk '/^var legacySkillIDs = \[\]string\{/{f=1;next} f&&/^\}/{f=0} f' \
  internal/cli/update_archive.go | grep -c '"moai-'
# EXPECT: 13

# P3. The mutation never reached the repo.
grep -c 'moai-domain-backend' internal/cli/update_archive.go
# EXPECT: 0

rm -r "$WORK"
```

**What P2 does and does not do.** Under the overlay mechanism the source is untouched *by construction*, so P2 cannot fail as a consequence of *this* procedure — it is a **belt-and-braces check against unrelated damage** (a concurrent session, a mistaken edit elsewhere in the turn), not a discriminator of the overlay procedure's own failure mode. It is retained because it is nearly free and because the procedure it replaced *did* have a failure mode P2 would have caught: a `git checkout -- <path>` revert destroys the fix while leaving a clean tree, so a `git status` post-condition reports success precisely when it should report failure. P2 is what that retired procedure lacked; under the overlay it is insurance, not the load-bearing check. Stating this plainly matters because a run-phase reader relying on "P2 is the discriminating check" would over-trust a green P2.

Mechanism verified on a throwaway module before this SPEC was written: with the overlay applied the test printed `--- FAIL` and named the injected ID, while `grep -c '<injected>' <source>` on the real file returned `0`.

**No commit precondition.** Because the source is never written, M1 does not need to be committed before falsification begins — plan.md §E M2 records this as the reason the overlay form was chosen over a `git restore --source=<sha>` form.

### C.1 Guard falsification (binds AC-LSL-004 and AC-LSL-005 → REQ-LSL-008)

```bash
# 1. Confirm green on the corrected list (parent line, anchored).
go test ./internal/cli/ -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 \
  | grep -E '^--- PASS: TestLegacySkillIDsNotEmbedded'

# 2. Build the mutation per §C.0: re-add ONE live skill as the first element
#    of legacySkillIDs in "$WORK/mutated.go".
sed 's|^var legacySkillIDs = \[\]string{|&\n\t"moai-domain-backend",|' \
  "$FIX_SRC" > "$WORK/mutated.go"

# 3. The guard must go red under the overlay, and must name the ID.
go test -overlay="$OVERLAY_JSON" ./internal/cli/ \
  -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 | tee "$FALSIFY_LOG" >/dev/null
grep -cE '^--- FAIL: TestLegacySkillIDsNotEmbedded' "$FALSIFY_LOG"   # EXPECT: 1
grep -c 'moai-domain-backend' "$FALSIFY_LOG"                          # EXPECT: >= 1

# 4. Run §C.0 post-conditions P1, P2, P3.
```

Record steps 3 and 4's verbatim output in progress.md §E.2. A run-phase report claiming this procedure passed without that output is an unobserved-verification claim.

### C.2 Archive-loop falsification (binds AC-LSL-010 … AC-LSL-012, AC-LSL-016)

```bash
# 1. Confirm green.
go test ./internal/cli/ -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v

# 2. Build the mutation per §C.0. FIX_SRC is already internal/cli/update_archive.go,
#    so $OVERLAY_JSON from the §C.0 setup is reused unchanged. Restore the abort
#    behaviour by turning the accumulation of the "archive %s" error back into an
#    in-loop return. The accumulation line's exact text depends on the M4
#    implementation, so the sed pattern is written against the ACCUMULATOR call
#    the implementer introduced — substitute its literal for <ACC_CALL>:
sed 's|<ACC_CALL>|return archived, fmt.Errorf("archive %s: %w", id, err)|' \
  "$FIX_SRC" > "$WORK/mutated.go"
# Verify the mutation actually applied before trusting step 3 — a sed that
# matched nothing yields a byte-identical copy and a falsely-green result:
cmp -s "$FIX_SRC" "$WORK/mutated.go" && echo "MUTATION DID NOT APPLY — fix the pattern"
# EXPECT: no output

# 3. The test must go red under the overlay — both because the later skill is
#    not archived and because the `total:` line is absent.
go test -overlay="$OVERLAY_JSON" ./internal/cli/ \
  -run 'TestArchiveLegacySkills_ContinuesAfterFailure' -count=1 -v 2>&1 \
  | grep -cE '^--- FAIL: TestArchiveLegacySkills_ContinuesAfterFailure'
# EXPECT: 1

# 4. Run §C.0 post-conditions P1, P2, P3.
```

If step 3 stays green, the test's fixture does not actually reach the failure branch — fix the fixture, not the assertion.

### C.3 Degradation falsification (binds AC-LSL-007)

This procedure mutates a **different file** from §C.0's default, so `FIX_SRC` and `$OVERLAY_JSON` must be re-pointed. `go test -overlay` accepts a `_test.go` file as a Replace target — verified.

```bash
# 0. RE-POINT the overlay at the guard test file and rebuild the JSON.
FIX_SRC=internal/cli/update_archive_guard_test.go
sed 's|t\.Skip(|_ = fmt.Sprint(|' "$FIX_SRC" > "$WORK/mutated.go"   # skip → silent no-op
cmp -s "$FIX_SRC" "$WORK/mutated.go" && echo "MUTATION DID NOT APPLY — fix the pattern"
printf '{"Replace": {"%s": "%s"}}\n' \
  "$(cd "$(dirname "$FIX_SRC")" && pwd)/$(basename "$FIX_SRC")" \
  "$WORK/mutated.go" > "$OVERLAY_JSON"

# 1. Under the overlay, run with -v.
go test -overlay="$OVERLAY_JSON" ./internal/cli/ \
  -run 'TestLegacySkillIDsNotEmbedded' -count=1 -v 2>&1 | tee "$FALSIFY_LOG" >/dev/null

# 2. The degradation subtests now report PASS where they previously reported SKIP.
#    That transition is the proof the SKIP marker is load-bearing, not incidental.
grep -cE '^    --- SKIP: TestLegacySkillIDsNotEmbedded/' "$FALSIFY_LOG"   # EXPECT: 0 (was 2)
grep -cE '^    --- PASS: TestLegacySkillIDsNotEmbedded/' "$FALSIFY_LOG"   # EXPECT: 2 (was 0)

# 3. Post-conditions: P1 and P3 applied to the guard-test file.
git diff --quiet -- internal/cli/update_archive_guard_test.go; echo "unmodified_exit=$?"
# EXPECT: unmodified_exit=0
# P2 does not apply — that file carries no list.
# NOTE: the exact sed pattern depends on how the implementer wrote the skip
#       branch; the cmp guard in step 0 is what makes a non-matching pattern
#       visible instead of silently producing a byte-identical "mutation".

# 4. Re-point FIX_SRC back to internal/cli/update_archive.go before any further
#    §C procedure reuses the §C.0 setup.
```

### C.4 Archive-removal falsification (binds AC-LSL-008, AC-LSL-009)

This one cannot use an overlay (it is a git-index property, not a compile-time one), so it is constructed to need no destructive step at all:

```bash
# The AC-LSL-008 check reads the INDEX, not the filesystem, so it already
# discriminates a staged removal from a bare file deletion. Prove that without
# deleting anything, using git's index-only staging:
git rm --cached -q .moai/archive/skills/v2.16/moai-domain-backend/SKILL.md
git ls-files .moai/archive/skills/v2.16/moai-domain-backend/ | wc -l   # EXPECT: 0
git reset -q -- .moai/archive/skills/v2.16/moai-domain-backend/SKILL.md
git ls-files .moai/archive/skills/v2.16/moai-domain-backend/ | wc -l   # EXPECT: 1

# Post-condition: the file itself was never removed from disk.
test -f .moai/archive/skills/v2.16/moai-domain-backend/SKILL.md && echo "present"
# EXPECT: present
```

`git rm --cached` stages the un-tracking without touching the working file, and `git reset -- <path>` restores the index entry. Both are index operations; neither appears on the primary-checkout forbidden list, and no file content is at risk. Run this BEFORE M3's real `git rm`, so the demonstration and the fix do not interfere.

---

## §D Definition of Done

- [ ] AC-LSL-001 … AC-LSL-016 all pass, each with its command output recorded in progress.md §E.2.
- [ ] §C.1 through §C.4 executed, with the deliberate-failure output recorded verbatim AND the §C.0 post-conditions P1/P2/P3 recorded after each.
- [ ] No falsification step modified a tracked file (§C.0 [HARD]).
- [ ] `legacySkillIDs` holds 13 entries; the intersection with the embedded manifest is empty.
- [ ] `git ls-files .moai/archive/skills/v2.16` lists exactly 4 files.
- [ ] `go build ./...`, `go vet`, and the `internal/cli` + `internal/template` suites are green.
- [ ] `internal/template/templates/` is untouched.
- [ ] The out-of-scope follow-ups named in spec.md §C are carried forward, in particular the `moai migrate restore-skill` guard.
