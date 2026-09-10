#!/bin/bash
# t518 M-B1 mutant harness (SPEC-SPEC-LINT-ID-ARG-001).
#
# Each mutant reverses one piece of the repair and re-runs the acceptance
# criterion that is supposed to notice. No git commands: the originals are
# copied aside and restored with cp, so nothing depends on git state, and no
# background load is spawned.
#
# Two things are proved rather than assumed on every run, because each of them
# produced a false reading on the first pass of this harness:
#   - the mutation ACTUALLY LANDED (an unapplied mutation reads exactly like an
#     escaped one: the tests pass and nothing says why);
#   - the selector matched at least one test (a selector that matches nothing
#     is green and looks identical to a pass).
set -u
WT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518
cd "$WT" || exit 1
BK=/tmp/t518-mutant-bk
OUT="$WT/.moai/reports/t518/mb1-mutants.txt"
mkdir -p "$BK"
cp internal/cli/spec_lint.go "$BK/spec_lint.go"
cp internal/cli/spec_status.go "$BK/spec_status.go"
cp internal/cli/specid/specid.go "$BK/specid.go"

restore() {
  cp "$BK/spec_lint.go" internal/cli/spec_lint.go
  cp "$BK/spec_status.go" internal/cli/spec_status.go
  cp "$BK/specid.go" internal/cli/specid/specid.go
}

: > "$OUT"

# applied? reports whether the working tree differs from the pristine copies.
applied() {
  if diff -q "$BK/spec_lint.go" internal/cli/spec_lint.go > /dev/null &&
     diff -q "$BK/spec_status.go" internal/cli/spec_status.go > /dev/null &&
     diff -q "$BK/specid.go" internal/cli/specid/specid.go > /dev/null; then
    echo no
  else
    echo yes
  fi
}

run_mutant() {
  local id="$1" desc="$2" pkg="$3" sel="$4"
  local tmp="$BK/run.txt"
  local landed
  landed=$(applied)
  go test "$pkg" -run "$sel" -v > "$tmp" 2>&1
  local rc=$?
  local ran
  ran=$(grep -c '^=== RUN' "$tmp")
  local verdict="NOT CAUGHT"
  [ "$rc" -ne 0 ] && verdict="CAUGHT"
  [ "$ran" -eq 0 ] && verdict="$verdict / VOID: selector matched no test (rc reflects a build failure, not an assertion)"
  [ "$landed" = "no" ] && verdict="VOID: mutation never applied"
  {
    echo "=============================================================="
    echo "MUTANT $id — $desc"
    echo "command: go test $pkg -run '$sel'"
    echo "mutation landed in source: $landed"
    echo "tests matched by selector: $ran"
    echo "rc=$rc  => $verdict"
    echo "--- diff of the mutation ---"
    diff -u "$BK/spec_lint.go" internal/cli/spec_lint.go
    diff -u "$BK/spec_status.go" internal/cli/spec_status.go
    diff -u "$BK/specid.go" internal/cli/specid/specid.go
    echo "--- output ---"
    cat "$tmp"
  } >> "$OUT"
  echo "MUTANT $id landed=$landed rc=$rc ran=$ran  $verdict"
  restore
}

SEL_ALL='TestSpecLint|TestSpecStatusIDPattern|TestShapeFunction'

# M1 — remove the ID resolver: every argument is treated as a path.
perl -0pi -e 's/if !specid\.HasCanonicalSpecIDShape\(arg\) \{/if true {/' internal/cli/spec_lint.go
run_mutant M1 "ID resolver removed (AC-SLI-001a/001b/001c)" ./internal/cli/ "$SEL_ALL"

# M2 — remove the directory branch.
perl -0pi -e 's/if info, err := os\.Stat\(arg\); err == nil && info\.IsDir\(\) \{/if false {/' internal/cli/spec_lint.go
run_mutant M2 "directory branch removed (AC-SLI-002)" ./internal/cli/ 'TestSpecLint_DirectoryArg'

# M3 — remove the exit-code-3 diagnostic.
perl -0pi -e 's/if _, err := os\.Stat\(specPath\); err != nil \{/if false {/' internal/cli/spec_lint.go
run_mutant M3 "rc=3 diagnostic removed (AC-SLI-003)" ./internal/cli/ 'TestSpecLint_UnresolvableID_IsArgumentError'

# M4 — remove the path-signal pre-exclusion (plan.md D1's worse misreading).
perl -0pi -e 's/if hasPathSignal\(arg\) \{/if false {/' internal/cli/spec_lint.go
run_mutant M4 "path-signal pre-exclusion removed, anchored shape intact (AC-SLI-004a)" ./internal/cli/ 'TestSpecLint_PathForm_Unchanged'

# M4b — the same removal, with the shape loosened at the same time. This is the
# state under which the pre-exclusion is load-bearing.
perl -0pi -e 's/if hasPathSignal\(arg\) \{/if false {/' internal/cli/spec_lint.go
perl -0pi -e 's/if !specid\.HasCanonicalSpecIDShape\(arg\) \{/if !specIDPattern.MatchString(arg) {/' internal/cli/spec_lint.go
run_mutant M4b "path-signal pre-exclusion removed AND shape loosened (AC-SLI-004a)" ./internal/cli/ 'TestSpecLint_PathForm_Unchanged'

# M5 — swallow every unresolvable argument into exit code 3.
perl -0pi -e 's/\tfor _, arg := range args \{/\tfor _, arg := range args {\n\t\tif _, serr := os.Stat(arg); serr != nil \&\& !specid.HasCanonicalSpecIDShape(arg) {\n\t\t\treturn nil, \&exitCodeError{code: 3, msg: "spec lint: unresolved argument"}\n\t\t}/' internal/cli/spec_lint.go
run_mutant M5 "all unresolvable arguments routed to exit 3 (AC-SLI-004b)" ./internal/cli/ 'TestSpecLint_MissingPathArg_StillParseFailure'

# M6 — swap the discriminator for the unanchored same-named symbol. This is
# the mutant AC-SLI-005 exists for.
perl -0pi -e 's/if !specid\.HasCanonicalSpecIDShape\(arg\) \{/if !specIDPattern.MatchString(arg) {/' internal/cli/spec_lint.go
run_mutant M6 "discriminator swapped for spec_status.go:18 unanchored specIDPattern (AC-SLI-005)" ./internal/cli/ 'TestSpecLint_AnchorInvariant_SPEC_A_1'

# M7 — revert the help text.
perl -0pi -e 's/Use:   "lint \[SPEC-ID \| path\/to\/spec\.md \| SPEC directory \.\.\.\]"/Use:   "lint [spec.md...]"/' internal/cli/spec_lint.go
perl -0pi -e 's/Long: `Accepted arguments.*?Validate SPEC documents against:/Long: `Validate SPEC documents against:/s' internal/cli/spec_lint.go
run_mutant M7 "help text reverted to 'lint [spec.md...]' (AC-SLI-006)" ./internal/cli/ 'TestSpecLintHelp_NamesThreeArgumentShapes'

# M8 — resolve only the first argument.
perl -0pi -e 's/\tfor _, arg := range args \{/\tfor argIdx, arg := range args {\n\t\tif argIdx > 0 {\n\t\t\ttargets = append(targets, arg)\n\t\t\tcontinue\n\t\t}/' internal/cli/spec_lint.go
run_mutant M8 "only the first argument resolved (AC-SLI-007)" ./internal/cli/ 'TestSpecLint_MixedArgs'

# M9 — resolve IDs against cwd instead of the project root.
perl -0pi -e 's/root, err := findProjectRootFn\(\)/root, err := os.Getwd()/' internal/cli/spec_lint.go
run_mutant M9 "ID base swapped from findProjectRootFn to cwd (AC-SLI-001c)" ./internal/cli/ 'TestSpecLint_IDArg_FromSubdirectory'

# M10 — option alpha: a private copy of the shape check declared inside
# spec_lint.go. The copy must COMPILE, otherwise the test never runs and the
# red is the build failing rather than the criterion noticing.
perl -0pi -e 's/\/\/ hasPathSignal reports/func HasCanonicalSpecIDShape(s string) bool {\n\treturn specidShapeLocal.MatchString(s)\n}\n\nvar specidShapeLocal = regexp.MustCompile(`^SPEC(-[A-Z][A-Z0-9]*)+-\\d{3}\$`)\n\n\/\/ hasPathSignal reports/' internal/cli/spec_lint.go
perl -0pi -e 's/\t"path\/filepath"/\t"path\/filepath"\n\t"regexp"/' internal/cli/spec_lint.go
run_mutant M10 "option-alpha local copy of the shape check in spec_lint.go (AC-SLI-009 branch 1)" ./internal/cli/ 'TestShapeFunction_HomeAndName'

# M14 — keep the diagnostic inside the error object and stop writing it where
# the user can read it. Exit code 3 still fires, so a criterion asserting only
# on the code would stay green; this is what makes AC-SLI-003 non-vacuous.
perl -0pi -e 's/\t_, _ = fmt\.Fprintln\(stderr, msg\)\n//' internal/cli/spec_lint.go
perl -0pi -e 's/func argumentError\(stderr io\.Writer, format string, a \.\.\.any\) error \{/func argumentError(stderr io.Writer, format string, a ...any) error {\n\t_ = stderr/' internal/cli/spec_lint.go
run_mutant M14 "diagnostic no longer written to stderr, rc=3 unchanged (AC-SLI-003 non-vacuity)" ./internal/cli/ 'TestSpecLint_UnresolvableID_IsArgumentError'

# M11 — change ValidateSpecID's rejection set.
perl -0pi -e 's/strings\.Contains\(specID, "\.\."\)/strings.Contains(specID, "....")/' internal/cli/specid/specid.go
run_mutant M11 "ValidateSpecID stops rejecting '..' (AC-SLI-009 branch 2 non-vacuity)" ./internal/cli/specid/ 'TestValidateSpecID'

# M12 — anchor the frozen spec_status literal.
perl -0pi -e 's/regexp\.MustCompile\(`SPEC-\[A-Z0-9-\]\+-\[0-9\]\+`\)/regexp.MustCompile(`^SPEC-[A-Z0-9-]+-[0-9]+\$`)/' internal/cli/spec_status.go
run_mutant M12 "spec_status.go:18 literal given anchors (AC-SLI-010 branch 2)" ./internal/cli/ 'TestSpecStatusIDPattern_LiteralFrozen'

# M13 — widen the copied shape literal so it drifts from internal/spec.
perl -0pi -e 's/\\d\{3\}\$`/\\d+\$`/' internal/cli/specid/specid.go
run_mutant M13 "copied shape literal widened to \\d+ (REQ-SLI-006 drift guard)" ./internal/cli/specid/ 'TestHasCanonicalSpecIDShape|TestCanonicalShapeLiteral'

restore
echo "ALL DONE"
