# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-GATE-001 (revision 0.1.0)

## A. Pin and conventions

- **Document-level pin.** Tree commit `ae34f268319ce91f028ee692f25c1bd476f9e6e5` (branch `WT-docs-ci-gate`, equal to local `main` at the time). Every cell whose command reads a committed object names the commit in its command (`git grep … ae34f2683`, `git cat-file -e ae34f2683:<path>`), so the pin is in the cell. Cells that run a script, a build, or an environment probe against the working tree carry no commit in the command; each is attributed to the working tree at `ae34f2683`, whose `git status --short` printed nothing before and after the measurements. Measured on 2026-10-18.
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no `;`, `&&`, pipe, or redirection outside quotes, no subshell, and no `echo $?` wrapper. Its verbatim stdout, its exit code, and the tree it was taken on are separate fields of its ledger entry. Where the standard error carries the observation (a `git cat-file -e` failure prints only there), the entry says so; stdout is empty.
- **Exit codes.** The shell tool displays a non-zero exit code for most commands but displays nothing for a `git grep` with no match (exit 1). Every zero-match cell therefore carries an exit code captured with a wrapper script that prints `exit=$?` after the command (`<scratch>/recipe/zero-match-exits.sh`, run in the plan correction of iteration 1; the guard refuses `git` inside an inline shell string but accepts it inside a script file). A matching form run against the same file is shown as the positive control.
- **Supplemental cells.** A supplemental cell (labelled `supplemental`) is a multi-command simulation. It is not a RED-now cell and never carries the release-blocking label alone; it exists to show behavior that the single-invocation cells only imply.
- **Tool provenance (verification-claim-integrity §2.2).** Verb-surface observations (EP-10) come from a `moai` built from this tree: `go build -o <scratch>/moai-tree ./cmd/moai` at `ae34f2683`, a plain build without the Makefile stamp, so its version output reads `v3.1.3 none built unknown` and carries no commit; the build command and the tree are the attribution. The installed `moai` (v3.2.0-rc.36-t1822c, commit `ee981a932`) is a strict ancestor of `ae34f2683` (`git merge-base --is-ancestor` exit 0) and is cited only for `moai spec lint` in `progress.md`, where its commit is stated next to the tree. External tools are not built from the tree and have no lag to state: `actionlint v1.7.10` (installed), `yq v4.53.2`, `hugo v0.160.1+extended+withdeploy` (Homebrew), `python3 3.14.2`, node 22, htmltest v0.17.0 and linkinator 8.1.0 (installed into a scratch directory in this pass).
- **Scratch.** `<scratch>` is `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3b8aad6a-ce40-4e01-92a2-a04a97198934/scratchpad`, machine-local and untracked. It is never a citation target: the deciding command and output of every claim are in §C.
- **Local-first.** New workflows cannot run on a remote runner from this plan (no push; operator batch). Every criterion below is verified locally with the tools named in §A and §B.0; the first remote run of each job is a regression-guard observation, recorded with its UTC time and never counted as a pass.
- **Classes.** `release-blocking (gated)`: re-executable on the tree; counts toward the completion verdict once its gate rows carry a verdict. `regression-guard`: not counted; the reason is stated.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Gate rows (decision-index.md) | Class | RED-now cells | Green-path milestone |
|----|-----|-------------------------------|-------|---------------|----------------------|
| AC-PRC-002 | REQ-PRC-002 (+ contract C-1 to C-8) | Q5, Q18 are inputs and gate nothing; Q25 gates the trigger shape (M5 only); Q12 and Q22 only if the step becomes required | release-blocking (gated) | EG-01, EG-02, EG-03, EG-11 | M1 (checker, test, fixtures), M5 (job, registration) |
| AC-PRC-012 | REQ-PRC-012 | Q9 | release-blocking (gated); not applicable if Q9 records no blocking | EG-05, EG-06, EG-07 (supplemental: EG-08, EG-08b) | M3 |
| AC-PRC-013 | REQ-PRC-013 | Q10 (DEFAULT-APPLIED); Q25 (trigger shape) | release-blocking (gated) | EG-09, EG-12, EG-11 | M4 |
| AC-PRC-014 | REQ-PRC-014 | Q11, Q26 (DEFAULT-APPLIED), Q25 (trigger shape) | release-blocking (gated) | EG-10, EG-13, EG-14, EG-11 | M2 |

Every REQ has an AC and every AC maps to a REQ (4 of 4). The regression-guard half of each AC (the first remote run) is recorded in `progress.md`, not counted.

## B.0 Local verification recipes

These are the multi-command recipes the criteria below cite. They are verification aids with expected outputs; they are not RED-now cells.

**R-12, strictness matrix.** Simulates the "Determine strictness" step and the `continue-on-error` expression of the parity workflow for each event. Validated on this tree (EG-08), on a fixed variant and on a mutant variant (EP-12). Usage: `bash <script> .github/workflows/docs-i18n-check.yml`.

```bash
#!/usr/bin/env bash
set -u
f="$1"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
q="'"
script="$(yq '.jobs.parity.steps[] | select(.id == "mode") | .run' "$f")"
expr="$(yq '.jobs.parity.steps[] | select(.id == "check") | .["continue-on-error"]' "$f")"
expr="${expr#\$\{\{ }"; expr="${expr% \}\}}"
eval_coe() { # $1=event $2=strict-output -> prints the continue-on-error value (absent key = false)
  case "$expr" in null|false|"") echo false; return;; esac
  local e="${expr//steps.mode.outputs.strict/$q$2$q}"
  e="${e//github.event_name/$q$1$q}"
  node -e "console.log(eval(process.argv[1]))" "$e"
}
for ev in push pull_request workflow_dispatch; do
  : > "$tmp/out"
  sub="${script//\$\{\{ github.event_name \}\}/$ev}"
  sub="${sub//\$\{\{ inputs.strict \}\}/true}"
  GITHUB_OUTPUT="$tmp/out" bash -c "$sub"
  strict="$(sed -n 's/^strict=//p' "$tmp/out")"
  printf 'event=%s strict=%s continue-on-error=%s\n' "$ev" "$strict" "$(eval_coe "$ev" "$strict")"
done
printf 'event=pull_request strict=false(forced) continue-on-error=%s\n' "$(eval_coe pull_request false)"
```

**R-12b, job outcome from the real step bodies.** Runs the actual `run` text of the strictness step and of the check step of the parity job for one event, against a clean tree and a drifted tree, and prints the job outcome (`failure` only when the step exits non-zero and `continue-on-error` is false). Trees: `git archive --format=tar ae34f2683 docs-site/content scripts/docs-i18n-check.sh` extracted into `<clean>`, a copy of it into `<drift>`, and `docs-site/content/en/getting-started/faq.md` removed from `<drift>`. Usage: `bash <script> <workflow-file> <clean> <drift> [event]`. Validated on the baseline (EG-08b), on a fixed variant, on a line-77-only variant, and on a variant whose check step is `echo skipped` (EP-12).

```bash
#!/usr/bin/env bash
set -u
f="$1"; clean="$2"; drift="$3"; ev="${4:-pull_request}"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
q="'"
mode="$(yq '.jobs.parity.steps[] | select(.id == "mode") | .run' "$f")"
chk="$(yq '.jobs.parity.steps[] | select(.id == "check") | .run' "$f")"
expr="$(yq '.jobs.parity.steps[] | select(.id == "check") | .["continue-on-error"]' "$f")"
: > "$tmp/mode.out"
sub="${mode//\$\{\{ github.event_name \}\}/$ev}"; sub="${sub//\$\{\{ inputs.strict \}\}/true}"
GITHUB_OUTPUT="$tmp/mode.out" bash -c "$sub"
strict="$(sed -n 's/^strict=//p' "$tmp/mode.out")"
if [ "$expr" = "null" ] || [ "$expr" = "false" ] || [ -z "$expr" ]; then coe=false; else
  expr="${expr#\$\{\{ }"; expr="${expr% \}\}}"
  e="${expr//steps.mode.outputs.strict/$q$strict$q}"; e="${e//github.event_name/$q$ev$q}"
  coe="$(node -e "console.log(eval(process.argv[1]))" "$e")"
fi
for t in "clean:$clean" "drift:$drift"; do
  name="${t%%:*}"; dir="${t#*:}"
  body="${chk//\$\{\{ steps.mode.outputs.strict \}\}/$strict}"
  : > "$tmp/chk.out"
  ( cd "$dir" && GITHUB_OUTPUT="$tmp/chk.out" bash -c "$body" >/dev/null 2>&1 ); rc=$?
  if [ "$rc" -eq 0 ]; then job=success; elif [ "$coe" = true ]; then job=success-swallowed; else job=failure; fi
  printf 'event=%s strict=%s continue-on-error=%s tree=%s step-exit=%s job=%s\n' "$ev" "$strict" "$coe" "$name" "$rc" "$job"
done
```

**R-13, Hugo build and its known failing input.** Builds the working tree site, then a scratch copy carrying an unknown shortcode. `HUGO_ENABLEGITINFO=false` is needed only for the scratch copy, which has no `.git` (the site sets `enableGitInfo = true`); the in-repo build uses the real configuration.

```bash
S="$(mktemp -d)"
hugo -s docs-site --destination "$S/public" --cacheDir "$S/cache" --minify --gc ; echo "build exit=$?"
cp -R docs-site "$S/bad"
printf '\n{{< unclosed\n' >> "$S/bad/content/ko/getting-started/introduction.md"
env HUGO_ENABLEGITINFO=false hugo -s "$S/bad" --destination "$S/public-bad" --cacheDir "$S/cache" --minify --gc ; echo "bad exit=$?"
```

**R-14, link-check, with htmltest as the tool.** Builds the site, runs the tool with the committed configuration from a scratch working directory (the tool writes a `tmp/` directory in its working directory), then repeats with one injected broken internal link. If decision Q11 names another tool, the same three observations (baseline exit 0, injected exit non-zero naming the target, no file written in the tree) apply with that tool's invocation.

```bash
S="$(mktemp -d)"
root="$(pwd)"
env HUGO_ENABLEGITINFO=false hugo -s docs-site --destination "$S/public" --cacheDir "$S/cache" --minify --gc
( cd "$S" && htmltest -c "$root/docs-site/.htmltest.yml" "$S/public" ) ; echo "baseline exit=$?"
cp -R docs-site "$S/link"
printf '\n[injected broken link](/ko/does-not-exist/)\n' >> "$S/link/content/ko/getting-started/introduction.md"
env HUGO_ENABLEGITINFO=false hugo -s "$S/link" --destination "$S/public-link" --cacheDir "$S/cache" --minify --gc
( cd "$S" && htmltest -c "$root/docs-site/.htmltest.yml" "$S/public-link" ) ; echo "injected exit=$?"
git status --short
```

**R-14e, every ignore entry is necessary and tight.** Three checks per entry of `IgnoreURLs`. (1) Necessity: run R-14's baseline with that one entry removed from a scratch copy of the configuration; the run must exit non-zero. An entry whose removal leaves the run green is deleted. (2) Sibling negative control: build a scratch copy of the site into which one link has been injected that the entry's text must not cover, and run the tool with the committed configuration; the run must exit non-zero and its output must name that exact link. The siblings are `/changelog/not-a-baseline/` for the entry that covers `/changelog/` (injected as raw HTML `<a href="/changelog/not-a-baseline/">`, because Markdown links get a locale prefix), and `./not-a-baseline-link` for each relative-link entry. (3) Anchoring: `yq '.IgnoreURLs[]' docs-site/.htmltest.yml` prints only lines that begin with `^` and end with `$`; a line that does not is a defect. Measured on the draft configuration and on an over-broad mutant `^/changelog/.*$` (EP-16): the draft names both siblings; the mutant passes the necessity check (baseline exit 0) and does not name `/changelog/not-a-baseline/`, so check (2) rejects it.

**R-02m, mutant probe for the checker's test.** The test reads the checker path from the environment variable `CHECKER` (default: `scripts/docs-cited-verbs.py`). Three scratch stubs stand in for the checker: M-a prints `cited-verbs: 0` and exits 0 whatever the input; M-b prints one unresolved line plus `cited-verbs: 1` and exits 1 whatever the input; M-c exits 1 and prints no final line. Four more mutants are one-line edits of a scratch copy of the finished checker: M-d holds a hard-coded verb list that equals the first stub plus the real surface; M-e also scans prose outside code spans; M-f counts a flag as a citation; M-g has no exemption entries. `CHECKER=<mutant> sh scripts/test-docs-cited-verbs.sh` must exit non-zero for each and name the failing cases: M-a fails the fixture-a, fixture-c, inline-code-span, and same-text-elsewhere cases; M-b fails fixture-b and every pass-expecting case; M-c fails the format assertions of every case; M-d fails `stub-only-verb-resolves` and `second-stub-flips-fixture-b`; M-e fails `prose-only-citation-passes`; M-f fails `flag-citation-not-counted`; M-g fails `cg-mode-page-exempt` and `moai-todo-page-exempt`. At M1 the lane records the failing case names for each mutant.

## B. Acceptance criteria

### AC-PRC-002 — the cited-verbs check exists, fails on an unresolved citation, and was seen red (REQ-PRC-002)

- **Given** no checker, test, fixture, or job exists on the tree (EG-01 to EG-04), and no liveness-manifest entry names the job (EG-11).
- **When** M1 adds `scripts/docs-cited-verbs.py`, `scripts/test-docs-cited-verbs.sh`, and the three fixtures, and M5 adds `.github/workflows/docs-cited-verbs.yml` and its manifest entry,
- **Then**
  1. `sh scripts/test-docs-cited-verbs.sh` exits 0 and its last line is `RESULT pass=13 fail=0`. The thirteen cases print, in order, `PASS fixture-a-absent-subverb-fails`, `PASS fixture-b-listed-subverb-passes`, `PASS fixture-c-mermaid-label-fails`, `PASS unreadable-verb-source-exits-2`, `PASS empty-scan-set-exits-2`, `PASS prose-only-citation-passes` (C-1), `PASS inline-code-span-citation-counted` (C-1), `PASS flag-citation-not-counted` (C-6), `PASS stub-only-verb-resolves` (C-2), `PASS second-stub-flips-fixture-b` (C-2), `PASS cg-mode-page-exempt` (C-7), `PASS moai-todo-page-exempt` (C-7), `PASS same-text-elsewhere-flagged` (C-7 control). The test prints `RESULT pass=0 fail=0` and exits non-zero if it ran no case.
  2. The test builds its own stub `moai` (canned `--help` output with `spec status`, `plan`, `todo` with `add` and `list`, a verb `zz-stub-verb` that exists in no real CLI, and `codex` with the usage alternatives `[cli | status | app]`), a second stub whose help lacks `spec status`, and a stub commands directory, so it reads no installed binary. Cases that need inputs beyond the three committed fixtures write them into a temporary directory the test removes on exit; no file is added to the tree.
  3. Case a: the checker run on `scripts/docs-cited-verbs-testdata/red-absent-subverb.md` exits 1, prints exactly one citation line ending in `:moai codex job`, and its final line is `cited-verbs: 1`. Case b: on `green-listed-subverb.md` it exits 0 and prints only `cited-verbs: 0`. Case c: on `red-mermaid-label.md` it exits 1, prints one line ending in `:moai run`, and the final line is `cited-verbs: 1`. Case d: with `--moai` pointing at a path that does not exist it exits 2 and prints no `cited-verbs: 0` line. Case e: scanning a directory with no Markdown file exits 2. Case `prose-only-citation-passes`: a temporary file whose only citation of an absent verb (`moai cg`) sits in a prose sentence outside any code span exits 0 and prints `cited-verbs: 0`. Case `inline-code-span-citation-counted`: the same citation inside an inline code span exits 1 and prints one line ending in `:moai cg`. Case `flag-citation-not-counted`: a code block holding `moai spec status --factory-run` exits 0. Case `stub-only-verb-resolves`: a code block holding `moai zz-stub-verb` exits 0 (no hard-coded list can know that verb). Case `second-stub-flips-fixture-b`: fixture b run against the second stub exits 1 with one line ending in `:moai spec status`. Case `cg-mode-page-exempt`: under a temporary root, `docs-site/content/ko/multi-llm/cg-mode.md` holding a code block with `moai cg` exits 0; case `moai-todo-page-exempt`: `docs-site/content/ko/utility-commands/moai-todo.md` holding `moai todo rename` exits 0; case `same-text-elsewhere-flagged`: `docs-site/content/ko/other.md` holding `moai cg` exits 1 and names `:moai cg`.
  4. R-02m: each of the seven mutants M-a to M-g makes the test exit non-zero and fails the cases named for it there.
  5. `yq '.jobs.*.name' .github/workflows/docs-cited-verbs.yml` prints `Docs cited verbs`; `yq '.on | has("pull_request")'` prints `true`; the job builds the `moai` binary from the tree, runs the test, then runs the checker with that binary; no step carries `continue-on-error: true`; `actionlint .github/workflows/docs-cited-verbs.yml` exits 0 with no output.
  6. `.moai/guards/liveness.yaml` has an entry whose locator is `.github/workflows/docs-cited-verbs.yml` with all five fields (`kind`, `locator`, `events`, `window`, `measure`), and `go test ./internal/guardstate/ -run '^(TestCensus_SetDifferenceEmptyBothDirections|TestCensus_EveryEntryCarriesFiveFields|TestCensus_ReleaseOnlySubjectsDeclareTheirCondition)$' -count=1 -v` exits 0 with three `--- PASS` lines (EP-11 shape; the anchored selector keeps the run from silently selecting fewer or more tests).
  7. After all runs above, `git status --short` prints nothing (C-8).
- **RED-now cells:** EG-01 (the checker file is absent), EG-02 (running the test fails with exit 127), EG-03 (the job file is absent), EG-11 (no manifest entry). Each is red because the artifact it names does not exist, and each flips only when the work that this criterion measures adds it. EG-04 shows no producer exists anywhere else in the tracked tree.
- **Green path:** M1 flips EG-01 and EG-02 (`cat-file` exits 0; the test exits 0 with `RESULT pass=13 fail=0`). M5 flips EG-03 and EG-11. A file-existence cell alone can be satisfied by an empty file, so the behavioral outcomes in Then 1 to 4 are the criterion; the existence cells are the pre-state.
- **Mutant probe:** a checker that always exits 0 fails cases a and c; one that always exits 1 fails case b; one that prints no final line fails the format assertions; one that treats an empty scan as success fails case e; one that returns 0 on an unreadable verb source fails case d; one with a hard-coded verb list fails `stub-only-verb-resolves` and `second-stub-flips-fixture-b` (C-2); one that scans prose fails `prose-only-citation-passes` (C-1); one that counts flags fails `flag-citation-not-counted` (C-6); one without the exemption entries fails `cg-mode-page-exempt` and `moai-todo-page-exempt` (C-7). R-02m exercises all of these.
- **Gate:** none for M1. Q25 gates the trigger shape written at M5. Q5 and Q18 gate nothing (the checker scans neither `docs-site/scripts/` nor flags).
- **Done when:** Then 1 to 7 hold on the committed tree and the commit SHA is recorded in `progress.md` §E.2 for the parent's M6. The regression-guard half: the first remote run of "Docs cited verbs" is recorded with its UTC time.
- **Recorded, not a criterion.** At M1 the lane runs the finished checker once on the real tree with a binary built from the tree and records `cited-verbs: <N>` and the per-family comparison with the parent's E-01 to E-04 in `progress.md` §E.2. A mismatch is reported, not tuned away.

### AC-PRC-012 — the i18n parity job blocks pull requests (REQ-PRC-012)

- **Given** the strict run passes on this tree and on `origin/main` (EP-01, EP-02), the strict run goes red on a known failing input while the warn-only run on the same input exits 0 (EP-03), and the pull-request path of the job is advisory: it writes `strict=false` (EG-05), its `continue-on-error` expression can evaluate to true (EG-06), and its header says it is advisory (EG-07).
- **When** Q9 is applied and M3 edits the pull-request branch of the strictness step (lines 76 to 77), the expression at line 85, and the header comment (lines 3 to 16) of `docs-i18n-check.yml` (no wider edit is needed: with line 77 forcing strict mode the check step already runs `DOCS_I18N_STRICT=1`, EP-12),
- **Then**
  1. `bash <R-12 script> .github/workflows/docs-i18n-check.yml` prints exactly:
     ```
     event=push strict=false continue-on-error=true
     event=pull_request strict=true continue-on-error=false
     event=workflow_dispatch strict=true continue-on-error=false
     event=pull_request strict=false(forced) continue-on-error=false
     ```
     The first line (push) is unchanged from today, the second flips the pull-request path, the fourth shows that the expression does not let a pull-request run continue on error even if the strictness output were false (an absent `continue-on-error` key counts as false, so the fourth line does not pin one implementation; the push line still requires the expression to stay for non-pull-request events).
  2. `yq '.jobs.parity.steps[] | select(.id == "check") | has("if")' .github/workflows/docs-i18n-check.yml` prints `false` (the check step cannot be skipped on a pull request).
  3. `git grep -c -F "ADVISORY ONLY" -- .github/workflows/docs-i18n-check.yml` prints nothing.
  4. `actionlint .github/workflows/docs-i18n-check.yml` exits 0 with no output.
  5. The job's name is still `Validate 4-locale parity` (`yq '.jobs.parity.name'`), and the `on` triggers are unchanged: `yq -o=json -I=0 '.on' .github/workflows/docs-i18n-check.yml` prints exactly `{"pull_request":{"paths":["docs-site/content/**","scripts/docs-i18n-check.sh",".github/workflows/docs-i18n-check.yml"]},"push":{"branches":["main"],"paths":["docs-site/content/**","scripts/docs-i18n-check.sh"]},"workflow_dispatch":{"inputs":{"strict":{"description":"Run in strict mode (fail on any drift)","required":false,"default":"true","type":"choice","options":["true","false"]}}}}`. The parity trigger is out of scope of this child and is not governed by Q25 (spec.md § Out of Scope).
  6. The strict script still goes red on a known failing input: in a scratch tree extracted from the working tree (`docs-site/content` plus `scripts/docs-i18n-check.sh`), deleting one `en` page makes `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` exit 1 (EP-03 shape), and the unmodified tree exits 0.
  7. R-12b on the committed workflow for `pull_request` prints exactly:
     ```
     event=pull_request strict=true continue-on-error=false tree=clean step-exit=0 job=success
     event=pull_request strict=true continue-on-error=false tree=drift step-exit=1 job=failure
     ```
     This runs the real `run` text of both steps, so a check step that does not invoke the parity script cannot pass it.
- **RED-now cells (replacing the 0.2.2 cell that cited E-16d, defect B01):** EG-05 prints `strict=false` on line 77, the `else` branch that every event other than `workflow_dispatch` and `push` takes, that is `pull_request`; EG-06 prints the expression at line 85; EG-07 prints the header count 1. Each is red for the stated reason: the pull-request path is advisory. EG-08 (supplemental) shows the routing effect: `event=pull_request strict=false continue-on-error=true`; EG-08b (supplemental) shows the outcome: with strictness false the drifted tree gives `step-exit=0 job=success`, so the criterion fails on a baseline whose strictness is false. The retired cell, E-16d, is green by construction (the strict run passes today, EP-01) and could not distinguish this work from no work.
- **Green path:** M3 flips EG-05 (only line 74, the push branch, still prints), EG-06 (the expression no longer depends on the strictness output for a pull request), EG-07 (no output), EG-08 (Then 1), and EG-08b (Then 7).
- **Mutant probe (all measured, EP-12):** the variant that edits only line 77 prints `continue-on-error=true` on the fourth line, so Then 1 fails it (and passes Then 7, which is correct: the job does block). A variant that adds an `if:` to the check step fails Then 2. A variant whose check step `run` is `echo skipped` passes Then 1 to 5 and fails Then 7: its drifted tree gives `step-exit=0 job=success`. A variant whose strictness output stays false fails Then 1 and Then 7.
- **Gate:** Q9 (open, no Default; the row waits for an operator verdict). If Q9 records that pull requests are not blocked, REQ-PRC-012 and this criterion are not applicable, and the parent's M6 imports AC-PRC-012 as not applicable.
- **Done when:** Then 1 to 7 hold on the committed tree. Regression-guard half: a pull request that carries a parity drift gets a red "Validate 4-locale parity" run; recorded with its UTC time after the operator's batch.

### AC-PRC-013 — a Hugo build job runs on pull requests (REQ-PRC-013)

- **Given** no workflow contains a Hugo build step (EG-12) and none is named for it (EG-09); the deployed pin is `0.160.1` (EP-04); the local build of this tree is green (EP-05); an unknown shortcode turns the build red (EP-05); no manifest entry exists (EG-11).
- **When** Q10 is applied (DEFAULT-APPLIED) and M4 adds `.github/workflows/docs-hugo-build.yml` and its manifest entry,
- **Then**
  1. `yq '.jobs.*.name' .github/workflows/docs-hugo-build.yml` prints `Docs Hugo build`; `yq '.on | has("pull_request")'` prints `true`.
  2. The workflow file holds no literal `0.160.1` (`git grep -c -F "0.160.1" -- .github/workflows/docs-hugo-build.yml` prints nothing) and reads the version from `docs-site/vercel.json` (`git grep -c -F "vercel.json" -- .github/workflows/docs-hugo-build.yml` prints a count of 1 or more); the extraction expression the job uses, run locally on `docs-site/vercel.json`, prints `0.160.1`.
  3. The job verifies the downloaded archive against the release's published SHA-256 file before it runs the binary: `yq '[.jobs.*.steps[].run | select(. != null) | select(test("sha256sum -c"))] | length' .github/workflows/docs-hugo-build.yml` prints 1 or more (the assertion is on parsed step bodies, so a comment does not satisfy it); no step carries `continue-on-error: true` (`yq '[.jobs.*.steps[] | select(has("continue-on-error"))] | length'` prints `0`); and no step body swallows a failure (`yq '[.jobs.*.steps[].run | select(. != null) | select(test("[|][|] true"))] | length'` prints `0`).
  4. The job's build command is `hugo --minify --gc` run against `docs-site`, the command the deployed build uses (`docs-site/vercel.json` `buildCommand`).
  5. R-13: the first build exits 0; the second, with an unknown shortcode, exits 1 and its output names the shortcode (`template for shortcode "unclosed" not found`).
  6. `actionlint .github/workflows/docs-hugo-build.yml` exits 0 with no output.
  7. The manifest entry and census test hold as in AC-PRC-002 Then 6, for `docs-hugo-build.yml`; `git status --short` prints nothing after the runs.
- **RED-now cells:** EG-09 (job file absent), EG-12 (no Hugo build step in any workflow), EG-11 (no manifest entry). Red because no gate exists; they flip when M4 adds it. Preconditions that make the criterion satisfiable and non-vacuous: EP-04 (version and archive exist), EP-05 (green on the clean tree, red on the injected input, so the job's command can fail).
- **Green path:** M4 flips EG-09 and EG-11; EG-12's pattern is intentionally narrow (build step names) and is replaced as the check by Then 4 and 5.
- **Mutant probe:** a job that runs `hugo version` but never builds fails Then 4; one with a hard-coded other version fails Then 2; one with `continue-on-error: true` or a `|| true` after the build fails Then 3; one that skips the archive check fails Then 3.
- **Gate:** Q10 (DEFAULT-APPLIED); Q25 gates the trigger shape.
- **Done when:** Then 1 to 7 hold on the committed tree. Regression-guard half: the first remote run of "Docs Hugo build", recorded with its UTC time; an install failure there (the archive name, the network) is a defect of this criterion, not a pass.

### AC-PRC-014 — a link-check job blocks broken internal links (REQ-PRC-014)

- **Given** no workflow contains a link checker or a link-check job (EG-10, EG-13), no tool configuration exists (EG-14), no manifest entry exists (EG-11), and the measured candidate is red on arrival in its default configuration (EP-06).
- **When** Q11 and Q26 are applied and M2 adds `.github/workflows/docs-link-check.yml`, the tool configuration (`docs-site/.htmltest.yml` if the tool is htmltest), and the manifest entry,
- **Then**
  1. `yq '.jobs.*.name' .github/workflows/docs-link-check.yml` prints `Docs link check`; `yq '.on | has("pull_request")'` prints `true`; the job builds the site with the same version rule as AC-PRC-013 Then 2 and runs the tool from a working directory outside the tree.
  2. R-14, baseline: the tool exits 0 on the site built from the tree (htmltest prints `passed` and `tested <N> documents`).
  3. R-14, injected: with one broken internal link added to a content page, the tool exits non-zero and its output names the target (`/ko/does-not-exist/`).
  4. R-14e: every entry of the ignore list is necessary (removing any one makes the baseline run exit non-zero), is exact-anchored (`^` ... `$`), and is tight: with a sibling link injected that the entry must not cover, the run exits non-zero and names that exact sibling. Each entry carries a comment naming the baseline links it covers and the count measured in EP-07.
  5. After R-14, `git status --short` prints nothing (the tool's `tmp/` directory lands in the scratch directory).
  6. `actionlint .github/workflows/docs-link-check.yml` exits 0 with no output; no step carries `continue-on-error: true` or `|| true`.
  7. The manifest entry and census test hold as in AC-PRC-002 Then 6, for `docs-link-check.yml`.
- **RED-now cells:** EG-10, EG-13, EG-14, EG-11. Red because no job or configuration exists. The precondition cell EP-06 states why the criterion needs more than the job file: the tool's default configuration reports 73,692 findings on the baseline, so without Then 2 the criterion would be red forever (the impossible direction of verification-completeness §2).
- **Green path:** M2 flips EG-10, EG-14, EG-11; Then 2 to 4 are the behavioral criterion.
- **Mutant probe:** a job that runs the tool against an empty directory exits 0 on both runs and fails Then 3; a configuration that ignores `.*` fails Then 3; an over-broad entry such as `^/changelog/.*$` still passes the baseline run and the necessity check but fails the sibling control of Then 4 (measured, EP-16), as does an unanchored entry (the anchoring check). What Then 4 does not guarantee: a future link whose text is exactly an ignored text, on any page, is still ignored (Q26 residual risk).
- **Gate:** Q11 (open; the tool must be named), Q26 (DEFAULT-APPLIED), Q25 (trigger shape). If Q11 names a tool other than htmltest, row 11 of the file set changes name or disappears and Then 2 to 6 are re-executed with that tool.
- **Done when:** Then 1 to 7 hold on the committed tree. Regression-guard half: the first remote run of "Docs link check", recorded with its UTC time.

## C. Evidence ledger (verbatim)

Pin for every `ae34f2683` cell: tree commit `ae34f268319ce91f028ee692f25c1bd476f9e6e5`.

### EG-01 — the citation checker file is absent (RED-now; AC-PRC-002)

- Command: `git cat-file -e ae34f2683:scripts/docs-cited-verbs.py`
- Exit code: 128. Stdout: empty. Standard error (verbatim): `fatal: path 'scripts/docs-cited-verbs.py' does not exist in 'ae34f2683'`

### EG-02 — the checker's test cannot run (RED-now; AC-PRC-002)

- Command: `sh scripts/test-docs-cited-verbs.sh` (working tree at `ae34f2683`, clean)
- Exit code: 127. Stdout: empty. Standard error (verbatim): `sh: scripts/test-docs-cited-verbs.sh: No such file or directory`

### EG-03 — the cited-verbs job is absent (RED-now; AC-PRC-002)

- Command: `git cat-file -e ae34f2683:.github/workflows/docs-cited-verbs.yml`
- Exit code: 128. Stdout: empty. Standard error (verbatim): `fatal: path '.github/workflows/docs-cited-verbs.yml' does not exist in 'ae34f2683'`

### EG-04 — no producer by any name in the tracked tree (RED-now; AC-PRC-002)

- Command: `git grep -n -i -E "cited-verbs|cited_verbs|citedverbs|docs-cited|cited verbs" ae34f2683 -- . ":!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001"`
- Exit code: 1 (wrapper `exit=$?`, `<scratch>/recipe/zero-match-exits.sh`, observed on the worktree at `ec977eaf1`; the command addresses the baseline objects by hash). Stdout: no output. Positive control for the form: EG-06 (same `git grep -n -F … ae34f2683 -- <path>` shape prints a match).
- Reading: the parent directory is excluded because it quotes the pattern; at `ae34f2683` this child's directory is not yet tracked.

### EG-05 — the pull-request path writes strict=false (RED-now; AC-PRC-012)

- Command: `git grep -n -F "strict=false" ae34f2683 -- .github/workflows/docs-i18n-check.yml`
- Exit code: 0. Stdout (verbatim; 2 lines):

```
ae34f2683:.github/workflows/docs-i18n-check.yml:74:            echo "strict=false" >> "$GITHUB_OUTPUT"
ae34f2683:.github/workflows/docs-i18n-check.yml:77:            echo "strict=false" >> "$GITHUB_OUTPUT"
```

- Reading: line 74 is the `push` branch. Line 77 is the `else` branch (lines 75 to 77), which runs for every event other than `workflow_dispatch` and `push`, that is `pull_request`. The same file, read through `yq` (stdout verbatim), shows the whole step:

```
if [[ "${{ github.event_name }}" == "workflow_dispatch" ]]; then
  echo "strict=${{ inputs.strict }}" >> "$GITHUB_OUTPUT"
elif [[ "${{ github.event_name }}" == "push" ]]; then
  # push to main: warn-only during Phase 1 rollout (matches PR behavior).
  # Flip to strict after the baseline 35 drifts are cleared (Phase 2)
  # via SPEC-V3R3-DOCS-PARITY-001 or equivalent — see CLAUDE.local.md §17.3.
  echo "strict=false" >> "$GITHUB_OUTPUT"
else
  # PRs: warn-only during Phase 1 rollout
  echo "strict=false" >> "$GITHUB_OUTPUT"
fi
```

(command: `yq '.jobs.parity.steps[] | select(.id == "mode") | .run' .github/workflows/docs-i18n-check.yml`, working tree, exit 0).

### EG-06 — the continue-on-error expression (RED-now; AC-PRC-012)

- Command: `git grep -n -F "continue-on-error" ae34f2683 -- .github/workflows/docs-i18n-check.yml`
- Exit code: 0. Stdout (verbatim; 1 line):

```
ae34f2683:.github/workflows/docs-i18n-check.yml:85:        continue-on-error: ${{ steps.mode.outputs.strict != 'true' }}
```

- Reading: with `strict=false` from EG-05 the expression is true for a pull request, so a failing check step does not fail the job.

### EG-07 — the header calls the job advisory (RED-now; AC-PRC-012)

- Command: `git grep -c -F "ADVISORY ONLY" ae34f2683 -- .github/workflows/docs-i18n-check.yml`
- Exit code: 0. Stdout (verbatim): `ae34f2683:.github/workflows/docs-i18n-check.yml:1`

### EG-08 — strictness simulation on the baseline (supplemental; AC-PRC-012)

- Command: `bash <R-12 script> .github/workflows/docs-i18n-check.yml` (multi-command recipe, §B.0; not a RED-now cell). Working tree at `ae34f2683`. Exit code 0.
- Stdout (verbatim):

```
event=push strict=false continue-on-error=true
event=pull_request strict=false continue-on-error=true
event=workflow_dispatch strict=true continue-on-error=false
event=pull_request strict=false(forced) continue-on-error=true
```

### EG-08b — job outcome on the baseline (supplemental; AC-PRC-012)

- Command: `bash <R-12b script> .github/workflows/docs-i18n-check.yml <clean> <drift>` (multi-command recipe, §B.0; not a RED-now cell). Trees extracted from `ae34f2683` with `git archive` (the drifted tree lacks `docs-site/content/en/getting-started/faq.md`). Exit code 0.
- Stdout (verbatim):

```
event=pull_request strict=false continue-on-error=true tree=clean step-exit=0 job=success
event=pull_request strict=false continue-on-error=true tree=drift step-exit=0 job=success
```

- Reading: the drifted tree passes the job on a pull request today, which is the red state of AC-PRC-012 Then 7.

### EG-09 — the Hugo build job is absent (RED-now; AC-PRC-013)

- Command: `git cat-file -e ae34f2683:.github/workflows/docs-hugo-build.yml`
- Exit code: 128. Stdout: empty. Standard error (verbatim): `fatal: path '.github/workflows/docs-hugo-build.yml' does not exist in 'ae34f2683'`

### EG-10 — the link-check job is absent (RED-now; AC-PRC-014)

- Command: `git cat-file -e ae34f2683:.github/workflows/docs-link-check.yml`
- Exit code: 128. Stdout: empty. Standard error (verbatim): `fatal: path '.github/workflows/docs-link-check.yml' does not exist in 'ae34f2683'`

### EG-11 — no liveness-manifest entry names the new workflows (RED-now; AC-PRC-002, -013, -014)

- Command: `git grep -n -E "docs-cited-verbs|docs-hugo-build|docs-link-check" ae34f2683 -- .moai/guards/liveness.yaml`
- Exit code: 1 (same wrapper). Stdout: no output. The manifest does name the existing parity workflow (positive control: `.moai/guards/liveness.yaml` lines 68 to 73 hold its entry, read in this pass).

### EG-12 — no Hugo build step in any workflow (RED-now; AC-PRC-013)

- Command: `git grep -c -i -E "hugo +(--minify|build)|setup-hugo|hugo-action|peaceiris|cli/hugo|gohugoio" ae34f2683 -- .github/workflows`
- Exit code: 1 (same wrapper). Stdout: no output. Positive control for the directory scope: EG-06 matched in `.github/workflows` files with the same `git grep … ae34f2683 -- <path>` shape.

### EG-13 — no link checker in any workflow (RED-now; AC-PRC-014)

- Command: `git grep -c -i -E "lychee|linkinator|htmltest|link-check|linkcheck|markdown-link-check" ae34f2683 -- .github/workflows`
- Exit code: 1 (same wrapper). Stdout: no output.

### EG-14 — the link-check tool configuration is absent (RED-now; AC-PRC-014; conditional on Q11 naming htmltest)

- Command: `git cat-file -e ae34f2683:docs-site/.htmltest.yml`
- Exit code: 128. Stdout: empty. Standard error (verbatim): `fatal: path 'docs-site/.htmltest.yml' does not exist in 'ae34f2683'`

### EP-01 — strict i18n run on this tree (precondition; green)

- Command: `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (working tree at `ae34f2683`, run from the worktree root). Exit code: 0. Stdout (verbatim):

```
=== Preflight: locale directory presence ===

=== Check 1: File path parity (ko ↔ target locales) ===
ℹ️  INFO   ko: 153 .md files (canonical)
ℹ️  INFO   en: 153 .md files
ℹ️  INFO   ja: 153 .md files
ℹ️  INFO   zh: 153 .md files

=== Check 2: Frontmatter title presence ===

=== Check 3: H1 heading existence ===

=== Check 4: Glossary term preservation (canonical: ko) ===

=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

### EP-02 — strict i18n run on origin/main (precondition; green; different tree)

- Tree: `origin/main` at `6dde1c1c1bdc84bf82c1b8266c071c44756fe7c0` (`git rev-parse origin/main`; the local remote-tracking ref, not re-fetched, gap G-4). Extracted with `git archive --format=tar -o <scratch>/origin-main-6dde1c1c1.tar origin/main -- docs-site/content scripts/docs-i18n-check.sh`, then `tar -x` into `<scratch>/origin-main-tree` (the script resolves its root from its own location).
- Command: `DOCS_I18N_STRICT=1 <scratch>/origin-main-tree/scripts/docs-i18n-check.sh`. Exit code: 0. Stdout: the same twelve lines as EP-01, with `Errors:   0` and `Warnings: 0`.

### EP-03 — the strict run goes red on a known failing input; the warn-only run does not (precondition)

- Input: the EP-02 extracted tree with `docs-site/content/en/getting-started/faq.md` removed.
- Command 1: `DOCS_I18N_STRICT=1 <scratch>/origin-main-tree/scripts/docs-i18n-check.sh`. Exit code: 1. Output (verbatim, both channels):

```
=== Preflight: locale directory presence ===

=== Check 1: File path parity (ko ↔ target locales) ===
ℹ️  INFO   ko: 153 .md files (canonical)
ℹ️  INFO   en: 152 .md files
❌ ERROR   missing translation: en/getting-started/faq.md (exists in ko)
ℹ️  INFO   ja: 153 .md files
ℹ️  INFO   zh: 153 .md files

=== Check 2: Frontmatter title presence ===

=== Check 3: H1 heading existence ===

=== Check 4: Glossary term preservation (canonical: ko) ===

=== Summary ===
Errors:   1
Warnings: 0

FAIL: 1 error(s) detected. Fix them or re-run with DOCS_I18N_STRICT=0 for warn-only mode.
```

- Command 2: `DOCS_I18N_STRICT=0 <scratch>/origin-main-tree/scripts/docs-i18n-check.sh`. Exit code: 0. The same output up to `Errors:   1`, then `WARN: 1 error(s) detected, but DOCS_I18N_STRICT=0 — exiting 0.`

### EP-04 — Hugo version, pin, and release (precondition; Q10)

- `hugo version` (exit 0): `hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew`
- `git grep -n -F "HUGO_VERSION" ae34f2683 -- docs-site/vercel.json` (exit 0): `ae34f2683:docs-site/vercel.json:11:      "HUGO_VERSION": "0.160.1",`
- Release page `https://github.com/gohugoio/hugo/releases/tag/v0.160.1` (fetched through the web-fetch tool): the release exists and was published "08 Apr 14:23"; the asset list did not load, so no asset name is taken from it.
- Asset probe: `curl -sIL -o /dev/null -w '%{http_code}' https://github.com/gohugoio/hugo/releases/download/v0.160.1/hugo_extended_0.160.1_linux-amd64.tar.gz` printed `200` after redirects.
- Checksum file `hugo_0.160.1_checksums.txt` from the same release contains the line `133adf4932c1b626b6fe6aa28d56791555abe3ceff167e03be534e8324c9ed39  hugo_extended_0.160.1_linux-amd64.tar.gz`. The archive itself was not downloaded or hashed in this pass (gap G-5).

### EP-05 — local Hugo build, clean and with an unknown shortcode (precondition; AC-PRC-013)

- Clean build, in the repository: `hugo -s /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1717/docs-site --destination <scratch>/hugo-public --cacheDir <scratch>/hugo-cache --minify --gc`. Exit code: 0. Stdout (verbatim; the `Total in N ms` line varies between runs and is not a compared value):

```
Start building sites … 
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew


              │ KO  │ EN  │ JA  │ ZH  
──────────────┼─────┼─────┼─────┼─────
 Pages        │ 187 │ 185 │ 185 │ 185 
 Paginator    │   0 │   0 │   0 │   0 
 pages        │     │     │     │     
 Non-page     │  12 │  12 │  12 │  12 
 files        │     │     │     │     
 Static files │ 264 │ 264 │ 264 │ 264 
 Processed    │   0 │   0 │   0 │   0 
 images       │     │     │     │     
 Aliases      │   6 │   5 │   5 │   5 
 Cleaned      │   0 │   0 │   0 │   0 

Total in 15675 ms
```

- Control for the scratch environment, clean tree without `.git`: `env HUGO_ENABLEGITINFO=false hugo -s …/docs-site --destination <scratch>/hugo-public-nogit …` exited 0 with the same page counts (187, 185, 185, 185). Without that variable a scratch copy fails with `failed to load Git data: fatal: not a git repository`, which is why the variable is used for the scratch copies only.
- Known failing input: a scratch copy of `docs-site` with the line `{{< unclosed` appended to `content/ko/getting-started/introduction.md`. Command: `env HUGO_ENABLEGITINFO=false hugo -s <scratch>/docs-site-bad --destination <scratch>/hugo-public-bad --cacheDir <scratch>/hugo-cache --minify --gc`. Exit code: 1. Output (verbatim; path shortened to `…`):

```
Start building sites … 
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew

Total in 261 ms
ERROR error building site: assemble: failed to create page from pageMetaSource /getting-started/introduction: "…/docs-site-bad/content/ko/getting-started/introduction.md:291:1": failed to extract shortcode: template for shortcode "unclosed" not found
```

### EP-06 — link-check candidate htmltest on the site built from this tree (precondition; Q11, Q26)

- Tool: `go install github.com/wjdp/htmltest@latest` with `GOBIN=<scratch>/bin` resolved to v0.17.0 (`go: downloading github.com/wjdp/htmltest v0.17.0`); `go install github.com/wjdp/htmltest@v0.17.0` into a second scratch bin directory then exited 0 with no output. Binary not checksummed (gap G-10).
- Inputs: Run 1 read `<scratch>/hugo-public`, the in-repository build of EP-05 (real configuration). Runs 2 to 4 read `<scratch>/hugo-public-nogit` (built by `env HUGO_ENABLEGITINFO=false hugo -s …/docs-site --destination <scratch>/hugo-public-nogit …`, exit 0, 187/185/185/185 pages) and `<scratch>/hugo-public-link`.
- Run 1, default configuration, external checks skipped: `htmltest -s -l 1 <scratch>/hugo-public` (from a scratch working directory). Exit code: 1. Final lines: `✘✘✘ failed in 1.749395583s` / `73692 errors in 657 documents`. Derived counts from the output file (`grep -c` per message): `target does not exist` 668; `href lacks trailing slash` 72925; `hash does not exist` 7; `alt text empty` 92; sum 73692. The site sets `trailingSlash: false` and `cleanUrls: true` in `docs-site/vercel.json`, which explains the 72,925 trailing-slash findings as a false class for this deployment.
- Run 2, configuration A (`CheckExternal: false`, `IgnoreDirectoryMissingTrailingSlash: true`, `IgnoreAltEmpty: true`, `LogLevel: 1`): exit 1, `688 errors in 657 documents`: 668 `target does not exist` and 20 `hash does not exist`.
- Run 3, configuration B, verbatim file (the draft of `docs-site/.htmltest.yml`; `DirectoryPath` is absolute only because of the scratch run):

```
DirectoryPath: <scratch>/hugo-public-nogit
CheckExternal: false
CheckInternalHash: false
IgnoreDirectoryMissingTrailingSlash: true
IgnoreAltEmpty: true
LogLevel: 1
IgnoreURLs:
  - "^/changelog/$"
  - "^\\./(update|profile|status|inventory)$"
  - "^\\./(cli|init-wizard|quickstart)$"
```

  Command: `htmltest -c cfg-b.yml`. Exit code: 0. Output (verbatim): `✔✔✔ passed in 2.394227916s` / `tested 657 documents`.
- Run 4, injected broken internal link: a scratch copy of `docs-site` with `[injected broken link](/ko/does-not-exist/)` appended to `content/ko/getting-started/introduction.md`, built into `<scratch>/hugo-public-link`; command `htmltest -c cfg-b-link.yml` (configuration B with `DirectoryPath` pointed at that output). Exit code: 1. Output (verbatim):

```
ko/getting-started/introduction/index.html
  target does not exist --- ko/getting-started/introduction/index.html --> /ko/does-not-exist/
========================================================================
✘✘✘ failed in 8.000548541s
1 errors in 657 documents
```

- The tool writes a `tmp/` directory in its working directory, not under the site; `git check-ignore` shows neither `tmp/` nor `docs-site/tmp/` is ignored in this repository, so the job and R-14 run it from a scratch working directory.

### EP-07 — the baseline's unresolved targets (precondition; Q26)

- Distinct unresolved targets in Run 2 (derived from the output file): `/changelog/` 636; `./update` 9; `./profile` 8; `./status` 5; `./cli` 4; `./inventory` 3; `./init-wizard` 2; `./quickstart` 1; total 668. By source page (derived, the 32 relative links): `getting-started/cli` carries `./update` 6, `./status` 5, `./profile` 5, `./inventory` 3 (19); the `getting-started` section index carries `./update` 3 and `./profile` 3 (6); `cli-reference/status` carries `./init-wizard` 2 and `./cli` 2, `cli-reference/inventory` carries `./cli` 2, and `cli-reference/update` carries `./quickstart` 1 (7). The 636 + 19 + 6 + 7 sum to 668. The `ko` locale does not appear among the relative-link sources in the tool's output (its pages use other link forms); that was not investigated.
- `/changelog/` comes from `docs-site/layouts/partials/site-footer.html` line 9 (`<a href="/changelog/">{{ i18n "footer_changelog" }}</a> ·`; `grep -rn changelog docs-site/layouts docs-site/hugo.toml docs-site/data`). It resolves in production only through `docs-site/vercel.json` line 118: `"source": "/changelog"` with destination `https://github.com/modu-ai/moai-adk/releases` (`sed -n 108,124p`).
- On the `getting-started/cli` pages (URL without a trailing slash, `docs-site/vercel.json` sets `trailingSlash: false`), `./update`, `./profile`, `./status`, `./inventory` resolve to `/:locale/getting-started/<name>`, which `docs-site/vercel.json` redirects at lines 28 (`inventory`), 38 (`status`), 58 (`update`), and 68 (`profile`) (a `grep -n -E` over the `source` keys of `vercel.json` for the `getting-started/{update,profile,status,inventory}` forms and `/changelog` printed lines 28, 38, 58, 68, and 118). That reading is an inference from URL resolution rules, not an observed request (gap G-16).
- On the `getting-started` section index (URL `/<locale>/getting-started`), `./update` and `./profile` resolve by the same rules to `/<locale>/update` and `/<locale>/profile`. `grep -n -E '"source": "(/:locale\(ko\|en\|ja\|zh\))?/(update|profile|status|inventory|cli)"' docs-site/vercel.json` printed nothing, so no rule serves those two forms. By inference these 6 links are broken for readers.
- `./cli`, `./init-wizard`, `./quickstart` appear on `cli-reference/*` pages (for example `ja/cli-reference/status` → `./cli`, `en/cli-reference/update` → `./quickstart`). `grep -n -E 'quickstart|init-wizard|cli-reference/cli' docs-site/vercel.json` printed nothing: no redirect resolves them, so by inference these 7 links are broken for readers.

### EP-08 — link-check candidate linkinator (precondition; partial; Q11)

- Install: `npm install --prefix <scratch>/npm-linkinator --no-audit --no-fund --loglevel=error linkinator@8.1.0` exited 0 (`added 25 packages`).
- `linkinator hugo-public-nogit --recurse --skip '^https?://(?!localhost|127\.0\.0\.1)' --verbosity error` printed `→ crawling hugo-public-nogit` / `✓ Successfully scanned 1 links in 0.046 seconds.` and exited 0: the site root `index.html` is a 228-byte redirect page with no anchors, so the crawl never started.
- Same command with start path `ko/index.html` and `--server-root hugo-public-nogit`: exit 1; output `[404] changelog/` and `ERROR: Detected 1 broken links. Scanned 46 links in 0.084 seconds.` Recursion did not reach the full page set (46 links against 657 documents for htmltest); no baseline-green or injected-red result was obtained.

### EP-09 — tool availability (precondition)

- Found: `actionlint` at `/Users/goos/go/bin/actionlint` (`actionlint -version`: `v1.7.10`, built with go1.25.6); `yq` (v4.53.2); `hugo` (EP-04); `python3` (3.14.2); `node`; `go` (go1.26.8).
- Absent (`command -v <tool>` exits 1 for each): `lychee`, `linkinator`, `htmltest`, `markdown-link-check`, `cargo`, `shellcheck`.
- `actionlint .github/workflows/docs-i18n-check.yml` (baseline file): no output; exit code not displayed as non-zero, so 0.

### EP-10 — the verb surface of the tree build (precondition; C-2, C-3)

- Binary: `go build -o <scratch>/moai-tree ./cmd/moai` at `ae34f2683` (exit 0); `moai-tree version` printed `moai-adk v3.1.3` / ` v3.1.3   none   built unknown`.
- `moai-tree --help` lists verbs under the headings `COMMANDS`, `LAUNCH COMMANDS:`, `PROJECT COMMANDS:`, and `TOOLS:` (for example `spec [command]`, `codex [command] [cli | status | app] [--flags]`, `plan [command] [--flags]`).
- `moai-tree codex --help`: COMMANDS block lists exactly `audit <role> [--flags]` and `role-audit <role> [--flags]`; the usage line reads `moai codex [command] [cli | status | app] [--flags]`; the body names `moai codex status`, `moai codex cli`, `moai codex app`.
- `moai-tree codex job --help`: exit 0, prints the same `codex` help (no error shown): an absent sub-verb does not fail.
- `moai-tree run --help`: exit 1; output `ERROR` / `Unknown command "run" for "moai".` / `Try --help for usage.`
- Consequence for C-3: `moai codex status` is valid yet is not in the COMMANDS block, so the usage alternatives are a second input; an exit-code probe cannot tell `codex job` from `codex status`.

### EP-11 — the liveness census test on the baseline (precondition; coupling)

- Command: `go test ./internal/guardstate/ -run '^(TestCensus_SetDifferenceEmptyBothDirections|TestCensus_EveryEntryCarriesFiveFields|TestCensus_ReleaseOnlySubjectsDeclareTheirCondition)$' -count=1 -v` (working tree). Exit code: 0. Output (verbatim; swept count 3):

```
=== RUN   TestCensus_SetDifferenceEmptyBothDirections
--- PASS: TestCensus_SetDifferenceEmptyBothDirections (0.00s)
=== RUN   TestCensus_EveryEntryCarriesFiveFields
--- PASS: TestCensus_EveryEntryCarriesFiveFields (0.00s)
=== RUN   TestCensus_ReleaseOnlySubjectsDeclareTheirCondition
--- PASS: TestCensus_ReleaseOnlySubjectsDeclareTheirCondition (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/guardstate	0.278s
```

- Reading: `internal/guardstate/census_test.go` lines 41 to 83 globs `.github/workflows/*.yml` and `*.yaml` and fails when a file has no entry in `.moai/guards/liveness.yaml` (set difference in both directions). The coupling is established from the test source; a red run was not induced (gap G-12).

### EP-12 — the R-12 recipe on a fixed variant and on a mutant (validation of the green path)

- Fixed variant (scratch copy of the workflow with line 77 changed to `strict=true` and line 85 changed to `continue-on-error: ${{ github.event_name != 'pull_request' && steps.mode.outputs.strict != 'true' }}`): R-12 stdout (verbatim):

```
event=push strict=false continue-on-error=true
event=pull_request strict=true continue-on-error=false
event=workflow_dispatch strict=true continue-on-error=false
event=pull_request strict=false(forced) continue-on-error=false
```

  `actionlint` on that variant: no output (exit 0).
- Mutant variant (line 77 changed only): R-12 stdout (verbatim):

```
event=push strict=false continue-on-error=true
event=pull_request strict=true continue-on-error=false
event=workflow_dispatch strict=true continue-on-error=false
event=pull_request strict=false(forced) continue-on-error=true
```

  The fourth line differs from the fixed variant, so the recipe separates the two. These variants are scratch files, not the committed edit; the choice of expression in the committed edit is the run phase's.
- R-12b on the same variants (stdout verbatim, `pull_request`, clean tree then drifted tree; exit code 0 each):
  - fixed variant: `event=pull_request strict=true continue-on-error=false tree=clean step-exit=0 job=success` then `event=pull_request strict=true continue-on-error=false tree=drift step-exit=1 job=failure`.
  - line-77-only variant: the same two lines as the fixed variant (the job blocks; R-12's fourth line is what separates it).
  - check step replaced by `echo skipped` (fixed variant plus `yq -i` on the check step's `run`): `event=pull_request strict=true continue-on-error=false tree=clean step-exit=0 job=success` then `event=pull_request strict=true continue-on-error=false tree=drift step-exit=0 job=success`, so Then 7 rejects it.

### EP-13 — no callers of the two scripts and no flag citations in scripts or workflows (precondition; Q5, Q18)

- `git grep -n -E "translate\.mjs|migrate-frontmatter\.sh" ae34f2683 -- . ":!docs-site/scripts" ":!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001"`: exit code 1 (same wrapper); no output.
- `git grep -n -E "factory-run|clear-policy" ae34f2683 -- scripts .github`: exit code 1 (same wrapper); no output.

### EP-14 — the stub-job convention in ci.yml (precondition; Q25)

- Command: `git grep -n -F "required check satisfied" ae34f2683 -- .github/workflows/ci.yml`. Exit code: 0. Stdout (verbatim): `ae34f2683:.github/workflows/ci.yml:698:          echo "No Go-code changes per paths-filter — Go test matrix skipped; required check satisfied."`

### EP-15 — the registered slash set at this tree (precondition; C-2)

- Command: `git ls-tree --name-only ae34f2683 .claude/commands/moai/`. Exit code: 0. Stdout: 17 paths (`clean`, `codemaps`, `e2e`, `feedback`, `fix`, `gate`, `goal`, `gtd`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`, `todo`, each `.md`); there is no `design`.

### EP-16 — ignore-list tightness: sibling controls and an over-broad mutant (precondition; AC-PRC-014 Then 4)

- Inputs: two scratch copies of `docs-site`, each with one extra link in `content/ko/getting-started/introduction.md`, built with `HUGO_ENABLEGITINFO=false` into `<scratch>/public-sibA` (raw HTML `<a href="/changelog/not-a-baseline/">`; that copy also carries an earlier Markdown link `/changelog/not-a-baseline/`, which Hugo rewrote to `/ko/changelog/not-a-baseline/`) and `<scratch>/public-sibB` (`[sibling](./not-a-baseline-link)`).
- Draft configuration (EP-06 configuration B), run from a scratch working directory:
  - `htmltest -c cfg-b.yml ../public-sibB`: exit 1; output `target does not exist --- ko/getting-started/introduction/index.html --> ./not-a-baseline-link`, `1 errors in 657 documents`.
  - `htmltest -c cfg-b.yml ../public-sibA`: exit 1; output names both `/ko/changelog/not-a-baseline/` and `/changelog/not-a-baseline/`, `2 errors in 657 documents`.
- Over-broad mutant (configuration B with the first entry changed to `"^/changelog/.*$"`):
  - on the baseline site (`htmltest -c cfg-broad.yml ../hugo-public-nogit`): exit 0, so the necessity check alone does not reject it.
  - on `public-sibA`: exit 1 with `1 errors in 657 documents`; the output names `/ko/changelog/not-a-baseline/` only. The sibling `/changelog/not-a-baseline/` is not named, so the sibling control rejects the mutant.
- The path argument overrides `DirectoryPath` from the configuration while the other settings (including `IgnoreURLs`) still apply: `htmltest -c cfg-b.yml ../hugo-public-link` printed `1 errors in 657 documents`.

## D. Definition of done

This child closes when all four criteria hold on its committed tree and the parent can import them by reference (parent plan §B.3).

| AC | REQ | Class | Imported at the parent's M6 as | Note |
|----|-----|-------|--------------------------------|------|
| AC-PRC-002 | REQ-PRC-002 | release-blocking (gated) | this child's AC-PRC-002 at its commit SHA | the checker prints `cited-verbs: <N>` as its last line; the parent's closing run needs N = 0 and exit 0 |
| AC-PRC-012 | REQ-PRC-012 | release-blocking (gated on Q9) | this child's AC-PRC-012 at its commit SHA | the RED cell was corrected (defect B01), so it counts at the index |
| AC-PRC-013 | REQ-PRC-013 | release-blocking (gated) | this child's AC-PRC-013 at its commit SHA | — |
| AC-PRC-014 | REQ-PRC-014 | release-blocking (gated on Q11) | this child's AC-PRC-014 at its commit SHA | the ignore list is part of the criterion (Then 4) |

Common gates for every criterion: `moai spec lint --strict` on this directory reports no finding (recorded in `progress.md`); `git status --short` prints nothing after the verification runs; the census selector of AC-PRC-002 Then 6 exits 0 with three `--- PASS` lines; the verification output is carried into the lane's verdict file, not left in scratch.

## E. Edge cases

- The parity workflow's `on.pull_request.paths` filter is unchanged, so a pull request that touches no docs path gets no parity run; whether that is acceptable for a required check is Q25.
- A pull request from a fork has a read-only token; the parity workflow's comment step already writes comments, and a fork run may fail at that step. The AC does not cover fork behavior (gap G-13).
- The checker run in an empty or misconfigured checkout exits 2, never 0 (C-5, case e).
- The Q24 marker: until SWEEPE defines it, a marker-bearing line is reported as unresolved. The conservative direction can make the parent's closing run fail; it cannot hide a defect.
- The Hugo job and the link-check job both build the site; a change to `docs-site/vercel.json` `HUGO_VERSION` changes both without an edit to either workflow.

## F. Quality gates (TRUST 5 mapping)

- Tested: every criterion has a RED-now cell and a stated green output; the checker has fixtures in both directions plus fail-closed cases; the Hugo and link-check commands were observed red on a known failing input (EP-05, EP-06).
- Readable: file names state their purpose; the ignore list is commented entry by entry.
- Unified: scripts follow the `scripts/test-*.sh` and `scripts/<tool>` conventions; workflows follow the `actions/checkout@v7` and `permissions: contents: read` conventions of the existing files.
- Secured: the Hugo archive is verified against the published SHA-256 before it runs; workflows request read-only permissions; no step interpolates untrusted event data into a script.
- Trackable: commits carry the card id `t1717` and the SPEC id; the evidence ledger is carried into the verdict file.

## G. Gaps (not observed)

- G-1 — No remote run. The new workflows were not run on any GitHub runner; no local Actions runner is installed. Their behavior on a runner is a regression-guard observation after the operator's batch.
- G-2 — Branch protection. Whether the three jobs become required checks is REQ-PRC-015, outside this SPEC (parent ledger E-21, external).
- G-3 — The shell tool does not display the exit status of a zero-match `git grep`; those statuses (EG-04, EG-11, EG-12, EG-13, EP-13) were captured with a wrapper script file (exit 1 for each). The wrapper ran on the worktree at `ec977eaf1`, not at `ae34f2683`; the commands address `ae34f2683` objects by hash, so the result does not depend on the checkout.
- G-4 — `origin/main` (EP-02) is the local remote-tracking ref at `6dde1c1c1`; it was not re-fetched in this run, so it may trail the remote.
- G-5 — Not measured: the Hugo edition Vercel installs, the Vercel install step `hugo mod get -u`, the SHA-256 of the downloaded archive, and the behavior of `enableGitInfo` on a shallow runner checkout. Scratch builds used `HUGO_ENABLEGITINFO=false`; the in-repo build used the real configuration (EP-05). Local verification used the Homebrew `extended+withdeploy` edition while the job downloads the plain `extended` archive; their builds were not compared.
- G-6 — lychee and markdown-link-check were not run; linkinator was run only partially (EP-08).
- G-7 — The link-check measurements used a site built with `HUGO_ENABLEGITINFO=false`. Git information affects last-modified metadata; that it leaves link targets unchanged is expected and was not compared.
- G-8 — The 20 anchor findings are not classified (real, or tool false positives); anchor checking is off in the measured configuration.
- G-9 — REDIR's Q21 option (a) removes the `/changelog` redirect (`vercel.json` line 118), after which the ignore entry for `/changelog/` hides a real 404 on 636 pages. REDIR and this child must agree on that entry before either lands; this SPEC does not settle it.
- G-10 — htmltest and linkinator were fetched through the Go module proxy and npm; the binaries were not checksummed beyond the toolchain's own verification.
- G-11 — The parent's audit verdict files (including the finding B11 that the parent's REQ-PRC-012 pointer cites) are not in this tree; REQ-PRC-012 takes its conditional form from the parent's pointer text alone.
- G-12 — The census coupling (EP-11) is established from the test source; a red run (a workflow file with no manifest entry) was not induced, to keep the tree clean.
- G-13 — Fork pull requests and the parity workflow's comment step were not examined.
- G-14 — The real-tree `cited-verbs: <N>` figure does not exist yet; it is recorded at M1.
- G-15 — The line-count estimate in spec.md (600 to 900) is an estimate, not a measurement.
- G-16 — No live request was made to the deployed site. Which of the 32 relative links and the footer link work in production is inferred from URL resolution rules and the redirect table (EP-07), not observed.

## H. Residual risk

- A checker that reads inline code and fenced blocks cannot see a citation written in prose or in an image; the parent's sweep counted prose as noise, and the scope rule keeps that choice.
- The ignore list matches link text, not the page it appears on (Q26 residual risk).
- A required job whose path filter skips it can leave an unrelated pull request waiting; the choice is Q25 and the closure is the parent's.
- The Q24 marker is a cross-child interface; two different tokens defined independently would make the parent's closing run fail.
