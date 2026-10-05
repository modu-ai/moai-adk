# plan.md — SPEC-CI-VERDICT-INTEGRITY-001

Tier: M (8 primary files, < 300 LOC of edits, 3 milestones) · Card: t1534 · Plan-phase tree: `a158b4b5f` (branch `WT-ci-verdict-integrity`, clean at authoring)

## §A Context

Five CI verdict surfaces misjudge non-success conclusions as success, and the required-checks SSoT has drifted from the checks the workflows actually publish. The card (operator directive 2026-10-06, Class C) carries the finding list from the codex analysis at `.moai/reports/t1534/codex-ci-analysis-2026-10-06.md` (primary checkout, READ-ONLY; measured on `main@ec13872f3`). Per the re-observation directive, every claim below was independently re-verified as a RED-now observation on THIS tree before becoming a requirement — the analysis supports, it does not replace, the measurements.

**Decision surface first (highest change-likelihood, least reversible):** the M2 SSoT correction changes what branch protection requires on `main` and reaches GitHub through an irreversible external-shared apply. That apply is a keep-set gate — see §F/M2. The M1 gate repairs and M3 watch repair are ordinary repo-file changes, fully revertible.

Files in scope (8 primary):

| # | File | Milestone |
|---|------|-----------|
| 1 | `.github/workflows/release-pr-multi-os.yml` | M1 |
| 2 | `.github/workflows/auto-merge.yml` | M1 |
| 3 | `.github/workflows/test-install.yml` | M1 |
| 4 | `.github/required-checks.yml` (SSoT — note: at `.github/` root, NOT under `workflows/`) | M2 |
| 5 | `.github/workflows/ci.yml` (detect filter only) | M2 |
| 6 | `scripts/ci-mirror/validate-required-checks.sh` | M2 |
| 7 | `internal/template/branch_protection_parity_test.go` | M2 |
| 8 | `scripts/ci-watch/run.sh` | M3 |

Allowed supporting files (only as needed): `scripts/ci-watch/lib/classify.sh` and `scripts/ci-watch/test/run_test.sh` (keep the M3 repair and its own test suite consistent); one new guard test file under `internal/template/` if the filter-correspondence check is authored as a separate test — 11 files total, matching spec.md constraint 2. This SPEC directory's own `repro/` evidence scripts and fixtures are SPEC artifacts (committed, tracked), not implementation files. Verified: none of the in-scope SSoT files has a mirror under `internal/template/templates/.github/` (only `branch-protection.json.gtmpl`, `labels.yml`, `actions/`, `workflows/` live there), so no Template-First regeneration is triggered; `.github/branch-protection.json.gtmpl` itself is shape-only (contexts injected at render time by `internal/cli/branch_protection.go:100`) and is NOT edited by this SPEC.

## §B Known Issues — verified RED evidence ledger

Document-level tree pin: **`a158b4b5f`** (verified via `git rev-parse --short HEAD` → `a158b4b5f`, exit 0, 2026-10-06; `git diff --stat a158b4b5f..HEAD -- .github/workflows/ install.sh install.ps1 scripts/` → empty at amendment-2 time, so every entry measured on the later artifact trees is content-identical for the measured files). Every entry below was measured in this run, on this tree. **Probe scripts are committed canonical copies** under `repro/` in this SPEC directory (the first measurements ran from `/tmp/t1534-probe/` working copies — those were ephemeral; the committed copies are the citable ground). The three workflow probes EXTRACT the step body from the live workflow file at run time (awk slice by step-name anchor, then the `run: |` block, then template substitution), so a re-run after the M1 repair exercises the repaired logic — never a frozen copy; re-running them post-repair is the M1 flip evidence (plan §E). The table below is the index (id / claim / command / exit / tree); the **verbatim raw stdout of every multi-line output lives in the fenced ledger immediately after the table** — verification-completeness §2.1's evidence-ledger carrier, cited here by id (table cells mangle shell metacharacters). `actionlint -color=false -shellcheck= .github/workflows/release-pr-multi-os.yml .github/workflows/auto-merge.yml .github/workflows/test-install.yml .github/workflows/ci.yml .github/workflows/codeql.yml` → no output, exit 0 (E18) — all five workflows are structurally valid; the defects are semantic.

| ID | Claim | Command (single invocation) | Verbatim stdout | Exit | Tree |
|----|-------|------------------------------|-----------------|------|------|
| E1 | Release gate never tests `cancelled` | `grep -c 'cancelled' .github/workflows/release-pr-multi-os.yml` | `0` | 1 | a158b4b5f |
| E2 | Gate with a cancelled matrix exits 0 + PASSED (release-pr-multi-os.yml:281-291 rejects only `failure`) — body extracted from the LIVE workflow at run time by the committed script | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-gate-cancelled.sh` | verbatim block L-E2 | 0 | a158b4b5f |
| E3 | Merge carries no head-commit pin | `grep -c 'match-head-commit' .github/workflows/auto-merge.yml` | `0` | 1 | a158b4b5f |
| E4 | Checks-lookup failure/empty output yields `should_merge=true` (auto-merge.yml:91-129 loop: `2>&1` capture + `\|\| true`, merge decision only from FAILED>0 / PENDING==0) — body extracted from the LIVE workflow at run time; stub at `repro/stubbin/gh` | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-checks-loop.sh` | verbatim block L-E4 | 0 | a158b4b5f |
| E5 | Job budget 20 min vs internal waits 10+15=25 min | `grep -n 'timeout-minutes\|MAX_WAIT' .github/workflows/auto-merge.yml` | verbatim block L-E5 (7 lines, full) | 0 | a158b4b5f |
| E6 | Install summary never tests `cancelled` | `grep -c 'cancelled' .github/workflows/test-install.yml` | `0` | 1 | a158b4b5f |
| E7 | Summary rejects only `failure` — FULL per-dependency matrix (P2-I + leader addendum + t1543 P2-N): the control plus EVERY dependency in the needs set flipped one at a time to EACH of `failure`/`cancelled`/`timed_out`/`skipped` (6 × 4 = 24 single-failure cases + control = 25). Pre-repair TWO-CLASS: 5 `failure` variants exit 1 (test-sh, test-ps1-pwsh, test-ps1-powershell, test-bat, compatibility-check — the working failure-blocking, control class), 18 non-success variants exit 0 (defect class — non-success invisible), and `install-script-parity-failure` exits 0 (the 6th failure variant — parity not even a dependency, deeper defect). A partial-dependency mutant (checking only test-sh+parity) survives the all-non-success input but fails the failure cases; body extracted from the LIVE workflow at run time | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-install-summary-cancelled.sh` | verbatim block L-E7 (matrix, 25 case exits) | 0 | a158b4b5f |
| E8 | Summary `needs` list (line 338) omits the `install-script-parity` job (defined at line 48) | `grep -n 'install-script-parity\|needs: \[' .github/workflows/test-install.yml` | verbatim block L-E8 | 0 | a158b4b5f |
| E9 | Required-feature probes print a missing marker but cannot fail (grep -q && echo ✓ \|\| echo ✗) — and the :309 install.sh probe's target is stale (E19) | `grep -n 'MOAI_INSTALL_DIR' .github/workflows/test-install.yml` | verbatim block L-E9 | 0 | a158b4b5f |
| E10 | Validator with yq absent from PATH: vacuous pass, exit 0 (get_yaml_list `2>/dev/null \|\| true` at :28 swallows the missing parser; empty lists skip every dimension) | `env PATH=/usr/bin:/bin sh scripts/ci-mirror/validate-required-checks.sh` | verbatim block L-E10 | 0 | a158b4b5f |
| E11 | Validator against a malformed-YAML SSoT: same vacuous pass, exit 0 (parser failure swallowed identically; committed fixture `repro/malformed/.github/required-checks.yml`, runner `repro/run-malformed.sh`) | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh` | verbatim block L-E11 | 0 | a158b4b5f |
| E12 | Installed gh CLI rejects the ci-watch field list — client-side, before any request; the rejection goes to STDERR (stdout is empty) | `gh pr checks 1748 --json name,status,conclusion,detailsUrl -R modu-ai/moai-adk` | verbatim block L-E12 (stdout 0 bytes; stderr carries the message) | 1 | a158b4b5f |
| E13 | Live published check names (PR 1748, read-only GET), unique names via jq: `Analyze (Go) (go)` present; `Test (macos-latest)`/`Test (windows-latest)` absent; `Release PR Multi-OS Gate` present; `Test Summary` and `docs-site install-script parity` present. A `CodeQL` and a bare `Analyze (Go)` name ALSO appear — attributed to non-PR-trigger and skipped-matrix runs on the same head SHA (skipped matrix jobs publish their literal `${{ matrix }}` names); the deterministic PR-published analysis check is `Analyze (Go) (go)`. Re-confirmed during the operator GET re-verification | `gh pr checks 1748 --json name,state,bucket -R modu-ai/moai-adk --jq '[.[].name] \| unique'` | verbatim block L-E13 | 0 | a158b4b5f |
| E14 | ci.yml detect filter has no branch-protection.json.gtmpl entry (parity test input at branch_protection_parity_test.go:41,57 never triggers the Test job when only that file changes) | `grep -c 'branch-protection.json.gtmpl' .github/workflows/ci.yml` | `0` | 1 | a158b4b5f |
| E15 | run.sh never calls `is_required` (SSoT whitelist classifier defined in lib/classify.sh:111 is unused; every non-auxiliary check is treated as required) | `grep -c 'is_required' scripts/ci-watch/run.sh` | `0` | 1 | a158b4b5f |
| E16 | Check names iterated via unquoted command substitution (whitespace split — `Test (ubuntu-latest)` fragments into 3 tokens) | `grep -n 'for check_name in' scripts/ci-watch/run.sh` | `134:    for check_name in $(_all_check_names "$TMP_JSON"); do` | 0 | a158b4b5f |
| E17 | Installed gh supports the merge-time head pin (flag line at :23 of the full help stdout — verbatim block L-E17) | `gh pr merge --help` | verbatim block L-E17 (38 lines, quoted in full) | 0 | a158b4b5f |
| E18 | actionlint baseline on the five in-scope workflows | `actionlint -color=false -shellcheck= .github/workflows/release-pr-multi-os.yml .github/workflows/auto-merge.yml .github/workflows/test-install.yml .github/workflows/ci.yml .github/workflows/codeql.yml` | (no output) | 0 | a158b4b5f |
| E19 | install.sh implements NO `MOAI_INSTALL_DIR` — the test-install.yml:309 probe targets a feature the script cannot express (stale probe; hard-checking it as-is would fail every install, the impossible direction). Real surface: `--install-dir` (:354), `-h\|--help` (:358-364), darwin/linux detection (:37-41) | `grep -c 'MOAI_INSTALL_DIR' install.sh` | `0` | 1 | a158b4b5f |
| E20 | install.ps1 DOES implement `MOAI_INSTALL_DIR` (env read at :307-308, set at :460) plus `IsWindows` (:287/:306) and `GetTempPath` (:221) — the :317 probe is live and stays | `grep -c 'MOAI_INSTALL_DIR' install.ps1` | `3` | 0 | a158b4b5f |
| E21 | All three producing workflows restrict their pull_request triggers to `branches: [main]` — no corrected context (nor any Lint/Test/Build/Gate context) publishes on a release-targeting PR; grounds the decision-index Q1 resolution (release/* list = published-only → empty; trigger expansion out via t1536). Measurement environment (corrected at amendment 5): this shell's grep is **ugrep 7.8.4 aarch64-apple-macosx**, which printed 12 lines in codeql→release→ci order with NO `--` separator lines; the auditor's **BSD grep 2.6.0-FreeBSD** yields 14 lines (ci→codeql→release, with `--` separators) — both are real outputs of the same command; the anchor facts are identical in both | `grep -n -A3 'pull_request:' .github/workflows/ci.yml .github/workflows/codeql.yml .github/workflows/release-pr-multi-os.yml` | verbatim block L-E21 (12 lines, raw, ugrep 7.8.4) | 0 | a158b4b5f |
| E22 | Healthy-input positive control: validator with yq present against the repo's real SSoT — every dimension executes on real content and passes (baseline for AC-CI-010's two-directional form; an always-fail mutant fails this) | `sh scripts/ci-mirror/validate-required-checks.sh` | verbatim block L-E22 | 0 | a158b4b5f |
| E23 | AC-CI-009 RED-now: the keep-set apply package does not exist yet on this tree (flips when run-phase delivers `.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md`) | `test -e .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` | (no output — file absent) | 1 | a158b4b5f |
| E24 | Validator against a VALID-YAML SSoT carrying a phantom required context (`Test (windows-latest)`, committed fixture `repro/phantom/`): silent pass, exit 0 — no dimension inspects required-context publishability (grounds AC-CI-011; a parser-handling-only repair does not flip this). Fixture mirrors a real producing tree: its workflow set publishes every LEGIT required name (`Lint` AND `Test (ubuntu-latest)` — leader repair P1-1 added the second publisher), so post-M2 the phantom alone is the discriminator | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh` | verbatim block L-E24 | 0 | a158b4b5f |
| E26 | Deadline-absence probe, REAL-BODY OBSERVATION form (AC-CI-005 mutant record; micro-card t1543 P2-M supersedes all if:-term/string approaches — translating Actions `if:` into Bash cannot verify the deadline contract): extracts the merge step's REAL body (env map + run block) from the LIVE workflow, redirects `gh` at a call-recording stub, provides the DECLARED CLOCK INTERFACE (env `current_time` + `merge_deadline`, epoch seconds), and executes the body under two clock runs. RED: `gh pr merge` called in BOTH runs → exit 0. FLIP: called pre-deadline, withheld post-deadline → exit 2. Guard blocks the pre-deadline merge → exit 3. Harness/execution failure → exit 4 (never conflated with the repair signal). Extraction failure → exit 9. Current tree: called in both runs (1 pre / 1 post, merge_at=1141 > declared 1140) → exit 0 (RED) | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-deadline-mutant.sh` | verbatim block L-E26 | 0 | a158b4b5f |
| E27 | Positive control for E24: the SAME fixture tree with the phantom context replaced by the genuinely published `Test (ubuntu-latest)` (ledger E13) — vacuous pass, exit 0 pre-repair; post-M2 it must REMAIN exit 0 while E24 flips, which only a context-reading publishability dimension achieves | `sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh` | verbatim block L-E27 | 0 | a158b4b5f |
| E25 | `gh pr checks` documents exit code 8 = "Checks pending" (line 10 of the help stdout; verbatim block L-E25) — the ground for REQ-CI-013's pending classification: a watch that treats every non-zero gh exit as fatal terminates during normal CI runs, when checks are legitimately pending | `gh pr checks --help` | verbatim block L-E25 (39 lines, quoted in full) | 0 | a158b4b5f |

Codex-provenance items cited but NOT re-measured here (recorded as supports, not observations): the live branch-protection context values on `main` (codex used the leader-measured values and did not re-query the API; the protection GET needs admin scope and is packaged below as an operator-executable pre-apply diff step); real Actions-runner-level cancellation behavior (E2/E4/E7 now extract the step body from the LIVE workflow at run time, but the `cancelled`/failure values are still template substitutions, not runner executions). Corroborating read: `internal/cli/branch_protection.go:100-105` renders any SSoT branch key into the protection payload, and `.github/required-checks.yml` lines 10-40 currently list `Test (macos-latest)`, `Test (windows-latest)`, `CodeQL` on both keys and omit `Release PR Multi-OS Gate` — E13 shows each of those four facts is wrong against live names.

### Verbatim stdout ledger (carrier for the table cells above — verification-completeness §2.1)

```text
L-E2  | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-gate-cancelled.sh
✓ Release PR multi-OS verification PASSED (or skipped for non-release / docs-only PR)

L-E4  | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-checks-loop.sh
Checking required CI checks for PR #1...
Check status: passed=0, pending=0, failed=0
All checks completed!
--- GITHUB_OUTPUT ---
should_merge=true

L-E5  | grep -n 'timeout-minutes\|MAX_WAIT' .github/workflows/auto-merge.yml
25:    timeout-minutes: 20
96:          MAX_WAIT=60
99:          while [ $WAIT -lt $MAX_WAIT ]; do
122:            echo "Waiting for checks to complete... ($WAIT/$MAX_WAIT)"
159:          MAX_WAIT=60   # 60 x 15s = 15 minutes
163:          while [ "$WAIT" -lt "$MAX_WAIT" ]; do
171:              *)                echo "CodeRabbit status: '${STATE:-<absent>}' - waiting ($WAIT/$MAX_WAIT)" ;;

L-E7  | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-install-summary-cancelled.sh
E7 input matrix (Actions default shell: bash -e, no pipefail; 1 control + 24 single-failure cases):
variant A-full-success: exit 0
variant test-sh-failure: exit 1
variant test-sh-cancelled: exit 0
variant test-sh-timed_out: exit 0
variant test-sh-skipped: exit 0
variant test-ps1-pwsh-failure: exit 1
variant test-ps1-pwsh-cancelled: exit 0
variant test-ps1-pwsh-timed_out: exit 0
variant test-ps1-pwsh-skipped: exit 0
variant test-ps1-powershell-failure: exit 1
variant test-ps1-powershell-cancelled: exit 0
variant test-ps1-powershell-timed_out: exit 0
variant test-ps1-powershell-skipped: exit 0
variant test-bat-failure: exit 1
variant test-bat-cancelled: exit 0
variant test-bat-timed_out: exit 0
variant test-bat-skipped: exit 0
variant compatibility-check-failure: exit 1
variant compatibility-check-cancelled: exit 0
variant compatibility-check-timed_out: exit 0
variant compatibility-check-skipped: exit 0
variant install-script-parity-failure: exit 0
variant install-script-parity-cancelled: exit 0
variant install-script-parity-timed_out: exit 0
variant install-script-parity-skipped: exit 0

L-E8  | grep -n 'install-script-parity\|needs: \[' .github/workflows/test-install.yml
48:  install-script-parity:
338:    needs: [test-sh, test-ps1-pwsh, test-ps1-powershell, test-bat, compatibility-check]

L-E9  | grep -n 'MOAI_INSTALL_DIR' .github/workflows/test-install.yml
309:          grep -q "MOAI_INSTALL_DIR" install.sh && echo "✓ MOAI_INSTALL_DIR support found" || echo "✗ MOAI_INSTALL_DIR support missing"
317:          grep -q "MOAI_INSTALL_DIR" install.ps1 && echo "✓ MOAI_INSTALL_DIR support found" || echo "✗ MOAI_INSTALL_DIR support missing"

L-E10 | env PATH=/usr/bin:/bin sh scripts/ci-mirror/validate-required-checks.sh
=== Dimension A: Validating auxiliary → workflow name mapping ===

=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ===

=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ===

✅ All validations passed

L-E11 | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh
=== Dimension A: Validating auxiliary → workflow name mapping ===

=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ===

=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ===

✅ All validations passed

L-E12 | gh pr checks 1748 --json name,status,conclusion,detailsUrl -R modu-ai/moai-adk
stdout: (0 bytes)
stderr:
Unknown JSON field: "status"
Available fields:
  bucket
  completedAt
  description
  event
  link
  name
  startedAt
  state
  workflow

L-E13 | gh pr checks 1748 --json name,state,bucket -R modu-ai/moai-adk --jq '[.[].name] | unique'
["Analyze (Go)","Analyze (Go) (go)","Build (darwin/amd64)","Build (darwin/arm64)","Build (linux/amd64)","Build (linux/arm64)","Build (windows/amd64)","CodeQL","CodeRabbit","Constitution Check","Cross-platform compatibility","Detect (paths-filter)","Detect Release PR","Integration Tests (macos-latest)","Integration Tests (ubuntu-latest)","Integration Tests (windows-latest)","Judgment-First Consistency","Lint","Race Test 1","Race Test 2","Release PR Multi-OS Gate","Release Verify (${{ matrix.os }})","Test (${{ matrix.os }})","Test (browser fire guard)","Test (ubuntu-latest)","Test Summary","Test install.bat (windows-latest, cmd)","Test install.ps1 (macos-latest, pwsh)","Test install.ps1 (ubuntu-latest, pwsh)","Test install.ps1 (windows-latest, powershell)","Test install.ps1 (windows-latest, pwsh)","Test install.sh (macos-latest)","Test install.sh (ubuntu-latest)","Validate 4-locale parity","Vercel","Vercel Preview Comments","docs-site install-script parity","graph-freshness","graph-freshness-conflict-guard","labeler","lsel-leak-guard","spec-lint","spec-status-sync","stale","welcome","workflow-parse-guard"]

L-E17 | gh pr merge --help
Merge a pull request on GitHub.

Without an argument, the pull request that belongs to the current branch
is selected.

When targeting a branch that requires a merge queue, no merge strategy is required.
If required checks have not yet passed, auto-merge will be enabled.
If required checks have passed, the pull request will be added to the merge queue.
To bypass a merge queue and merge directly, pass the `--admin` flag.


USAGE
  gh pr merge [<number> | <url> | <branch>] [flags]

FLAGS
      --admin                   Use administrator privileges to merge a pull request that does not meet requirements
  -A, --author-email text       Email text for merge commit author
      --auto                    Automatically merge only after necessary requirements are met
  -b, --body text               Body text for the merge commit
  -F, --body-file file          Read body text from file (use "-" to read from standard input)
  -d, --delete-branch           Delete the local and remote branch after merge
      --disable-auto            Disable auto-merge for this pull request
      --match-head-commit SHA   Commit SHA that the pull request head must match to allow merge
  -m, --merge                   Merge the commits with the base branch
  -r, --rebase                  Rebase the commits onto the base branch
  -s, --squash                  Squash the commits into one commit and merge it into the base branch
  -t, --subject text            Subject text for the merge commit

INHERITED FLAGS
      --help                     Show help for command
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format

LEARN MORE
  Use `gh <command> <subcommand> --help` for more information about a command.
  Read the manual at https://cli.github.com/manual
  Learn about exit codes using `gh help exit-codes`
  Learn about accessibility experiences using `gh help accessibility`

L-E21 | grep -n -A3 'pull_request:' .github/workflows/ci.yml .github/workflows/codeql.yml .github/workflows/release-pr-multi-os.yml
.github/workflows/codeql.yml:6:  pull_request:
.github/workflows/codeql.yml-7-    branches: [main]
.github/workflows/codeql.yml-8-  schedule:
.github/workflows/codeql.yml-9-    # Weekly scan: Monday at 06:00 UTC
.github/workflows/release-pr-multi-os.yml:13:  pull_request:
.github/workflows/release-pr-multi-os.yml-14-    branches: [main]
.github/workflows/release-pr-multi-os.yml-15-    types: [opened, synchronize, reopened, ready_for_review]
.github/workflows/release-pr-multi-os.yml-16-  workflow_dispatch:
.github/workflows/ci.yml:19:  pull_request:
.github/workflows/ci.yml-20-    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
.github/workflows/ci.yml-21-  # Enabling change only. A workflow_dispatch trigger becomes dispatchable
.github/workflows/ci.yml-22-  # only once the file carrying it exists on the default branch, so this does

L-E22 | sh scripts/ci-mirror/validate-required-checks.sh
=== Dimension A: Validating auxiliary → workflow name mapping ===
✓ docs-i18n-check (file found)

=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.main.contexts

=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.release/*.contexts

✅ All validations passed

L-E23 | test -e .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md
(no output — file absent on this tree)

L-E24 | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh
=== Dimension A: Validating auxiliary → workflow name mapping ===
✓ docs-i18n-check (file found)

=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.main.contexts

=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.release/*.contexts

✅ All validations passed

L-E25 | gh pr checks --help
Show CI status for a single pull request.

Without an argument, the pull request that belongs to the current branch
is selected.

When the `--json` flag is used, it includes a `bucket` field, which categorizes
the `state` field into `pass`, `fail`, `pending`, `skipping`, or `cancel`.

Additional exit codes:
	8: Checks pending

For more information about output formatting flags, see `gh help formatting`.

USAGE
  gh pr checks [<number> | <url> | <branch>] [flags]

FLAGS
      --fail-fast         Exit watch mode on first check failure
  -i, --interval int      Refresh interval in seconds in watch mode (default 10)
  -q, --jq expression     Filter JSON output using a jq expression
      --json fields       Output JSON with the specified fields
      --required          Only show checks that are required
  -t, --template string   Format JSON output using a Go template; see "gh help formatting"
      --watch             Watch checks until they finish
  -w, --web               Open the web browser to show details about checks

INHERITED FLAGS
      --help                     Show help for command
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format

JSON FIELDS
  bucket, completedAt, description, event, link, name, startedAt, state, workflow

LEARN MORE
  Use `gh <command> <subcommand> --help` for more information about a command.
  Read the manual at https://cli.github.com/manual
  Learn about exit codes using `gh help exit-codes`
  Learn about accessibility experiences using `gh help accessibility`

L-E26 | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-deadline-mutant.sh
body_exit=0 gh_interactions=1 head_queries=0 merge_calls=1 late_calls=1 (served clock crossed 1100 -> 1200 at first gh contact; deadline 1140)
gh pr merge 1 --merge --delete-branch | served=1200 late=yes
DEFECT: merge call(s) served AFTER the deadline crossed (late=1 of 1 merge call(s)) — the clock was not re-evaluated before merging

L-E27 | sh .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh
=== Dimension A: Validating auxiliary → workflow name mapping ===
✓ docs-i18n-check (file found)

=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.main.contexts

=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ===
✓ docs-i18n-check not in branches.release/*.contexts

✅ All validations passed
```

### Audit-record notes (iteration-1, P3 disclosures)

1. **E13 audit-environment note**: the iteration-1 audit sandbox could not re-observe E13 (network-restricted); the recording stands on this session's own GET measurement (L-E13). A future re-observation rides the operator GET re-verification.
2. **OwnershipTransitionUnmeasured**: the plan-phase commit `2f06f34c9` carries no `Authored-By-Agent` trailer, so the `(none) → draft` transition is recorded as unmeasured (lint INFO, non-gating). History is not amended; from amendment 2 forward, every commit on this SPEC carries the trailer (`Authored-By-Agent: manager-spec`), so later transitions measure.
3. **Amendment-2 probe-script correction record**: the committed extraction wrappers' first run exposed two defects in my own first draft (indent threshold `<= 8` matched the 8-space `run: |` line and produced empty bodies; the E24 fixture initially lacked the Dimension A stand-in workflow file and failed for the wrong reason). Both were fixed and re-measured before this ledger was written — the committed repro/ copies are the fixed versions, which is itself a demonstration of why extraction-from-live-tree beats frozen copies (audit P2-1st).
4. **Iteration-2 audit sandbox limits** (pre-empting iteration 3): the audit sandbox could not generate heredoc files (so E22/E24 could not be re-run there), had no network (E13), and re-ran the E2/E4/E7 step bodies directly instead of through the committed wrappers. The lane's committed repro/ scripts and this §B ledger are the primary measurements. As of amendment 4 the wrappers fail loudly on extraction drift (anchor-match-count + non-empty-body guard, exit 9) and execute the extracted body under `bash -e -o pipefail` matching the runners' default for shell:-less steps, so a post-M1 re-run through the wrappers is the flip evidence and cannot silently measure an empty script.

## §C Pre-flight (run-phase entry checks)

1. `git rev-parse --short HEAD` + `git branch --show-current` — confirm the tree and re-verify branch state immediately before any commit (staleness rule).
2. `actionlint` over the touched workflows after each edit (E18 baseline must stay clean).
3. `go build ./...` after any Go test edit; `GOOS=windows GOARCH=amd64 go build ./...` once before close (M2 touches a Go test).
4. Affected-package tests only: `go test -timeout 30m ./internal/template/...` (M2), and `sh scripts/ci-watch/test/run_test.sh` (M3) — no local full-suite runs (repo test discipline).
5. Validator behavior checks after M2: repeat E10 (yq-absence must now FAIL) and E11 (malformed must now FAIL) — the two probes flip from RED to green.
6. ci-watch repair check after M3: `MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh <PR> <BRANCH>` with `MOAI_CIWATCH_NO_SLEEP=1` against a real PR read-only, plus the stub-gh failure path (must exit 1, never all-pass).

## §D Constraints

1. **Keep-set apply (REQ-CI-009, binding).** No run-phase agent executes the GitHub branch-protection apply or any protection write. Run phase produces: corrected `.github/required-checks.yml`, the rendered payload (`RenderBranchProtectionJSON` output or the equivalent `gh api` body), the exact apply command, and the exact post-apply GET command — packaged for the operator/leader. Pre-apply, the operator diffs live vs SSoT (`gh api repos/modu-ai/moai-adk/branches/main/protection` read-back); post-apply, the GET read-back is the only acceptance evidence (REQ-CI-009). Applying protection from the pre-correction file values is prohibited.
2. Scope discipline: the eight primary files above only; no drive-by fixes; `internal/template/templates/**` untouched (card t1537 boundary); spec-lint.yml / GitHub-Flow residues (t1535) and release-pipeline cost items (t1536) untouched.
3. The ci-watch repair may not add dependencies beyond the installed gh's published field set (E12); the watch loop stays read-only over the SSoT (ci-watch-protocol: the SSoT is read-only for the watch layer — M3 edits run.sh/classify.sh behavior, never the SSoT's role).
4. Local test rules: `t.TempDir()` for any new Go test fixtures; `/tmp` for probe fixtures; affected packages only; `-timeout 30m` on package runs.
5. Conventional Commits per milestone (`fix(SPEC-CI-VERDICT-INTEGRITY-001): M<N> …`), `card t1534` in the body, `🗿 MoAI` trailer, no `--no-verify`, no force-push.

## §E Self-Verification (run-phase deliverable shape)

Manager-develop reports per manager-develop-prompt-template §E with the attribution triple (command / verbatim output / tree SHA):

- E1 matrix: AC-CI-001..AC-CI-014 PASS/FAIL, each citing the flipped RED cell (plan §B) and the new green output.
- E2 build: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3 tests: `go test -timeout 30m ./internal/template/...` (parity + any new guard test) and `sh scripts/ci-watch/test/run_test.sh`, verbatim outputs.
- E4 gate surfaces: actionlint clean on touched workflows; the committed repro/ probes (E2/E4/E7/E11/E24/E26/E27) re-run against the REPAIRED tree and flip to their green expectations — E2 gate fails a cancelled matrix, E4 withholds on lookup failure, E7's control stays exit 0 while ALL 24 single-failure variants exit non-zero, E10/E11/E24 exit non-zero; E22/E27 (healthy + control) stay exit 0; E26 flips to exit 2 (gh called pre-deadline, withheld post-deadline); ci-watch E12's rejected field list now validates.
- E5 keep-set package: `.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` delivered (flips E23); the operator-executed apply commands + GET read-back pasted into `.moai/reports/t1534/` after the operator window.
- E6 scope proof: `git status --short` showing only in-scope files; `git diff --stat` confirming `internal/template/templates/**` untouched.
- E7 blockers, if any (structured, never user-prompted).

## §F Milestones

### M1 — Gate verdict repairs (Priority High)

Files: `release-pr-multi-os.yml`, `auto-merge.yml`, `test-install.yml`.

| Criterion | RED cell (plan §B) | Green path |
|---|---|---|
| Gate fails on cancelled/timed_out matrix (AC-CI-001) | E1 + E2 | Gate requires `success` from detect AND matrix (or a named intentional exclusion); re-run E2's substitution with the repaired script shape → non-zero + error annotation; actionlint clean |
| Intentional exclusion gets a distinct reason-bearing output (AC-CI-002) | E2 (unconditional PASSED line reused for every case) | Docs-only / non-release skip emits a dedicated output naming the reason; the unconditional PASSED line only fires on a real success |
| Merge pinned to verified head (AC-CI-003) | E3 + E17 | Merge call passes `--match-head-commit <verified-sha>` and re-reads the head immediately before; mismatch/empty withholds; same read-before-merge verified by reading the repaired step |
| Lookup failure/empty/never-completes withholds merge (AC-CI-004) | E4 | Lookup requires success + JSON/complete response + required-set coverage; failure → `should_merge=false`; re-run E4's stub against the repaired loop → withholds |
| Deadline green verification resists the merge-after-deadline mutant (AC-CI-005) | E5 + E26 (REAL-BODY behavioral probe: the merge body runs under two clock inputs and `gh pr merge` is called in BOTH — merge_at=1141 > declared 1140, exit 0) | One deadline smaller than `timeout-minutes`; the merge step re-evaluates the deadline and the head at merge time; the flip evidence is `repro/run-deadline-mutant.sh` re-run on the repaired workflow: `gh pr merge` CALLED pre-deadline and NOT called post-deadline → exit 2 (the auditor's iteration-2 mutant: declared_deadline=1140, merge_at=1141 → exit 0 today, must become withheld); the declared clock interface is env `current_time` + `merge_deadline` (plan §F M2 keep-set note) |
| Install summary success-only incl. parity (AC-CI-006) | E6 + E7 + E8 | `needs` gains `install-script-parity`; summary requires `success` from every need (failure/cancelled/timed_out/skipped all non-success); E7 re-run on the repaired summary: control exit 0, all 24 single-failure variants non-zero — the 5 `failure` variants keep failing (they already did), the `install-script-parity-failure` variant and the 18 non-success variants flip from exit 0 |
| Feature-probe miss fails the step AND the probe list matches the scripts' real surface (AC-CI-007) | E9 + E19 + E20 | Required probes become hard checks (absent target → exit 1); the stale :309 install.sh probe (`MOAI_INSTALL_DIR`, unimplemented per E19) is realigned to the real surface (`--install-dir`); the :317 install.ps1 probe stays (E20); E9's shape no longer ends in `\|\| echo` for required probes |

### M2 — SSoT correction, parity wiring, validator hardening (Priority High; contains the keep-set gate)

Files: `.github/required-checks.yml`, `ci.yml` (detect filter), `validate-required-checks.sh`, `branch_protection_parity_test.go` (+ optional new guard test).

| Criterion | RED cell | Green path |
|---|---|---|
| SSoT lists only published names (AC-CI-008) | E13 + E21 | `main` list: remove `Test (macos-latest)` / `Test (windows-latest)` / `CodeQL`; add `Analyze (Go) (go)`; add `Release PR Multi-OS Gate`. `release/*` per decision-index Q1 (amendment-2 resolution): only what release-targeting PRs publish — an EMPTY contexts list under the current main-only triggers (E21); trigger expansion explicitly out via the t1536 boundary. The `CodeQL`/`Analyze (Go)` names appearing in E13 come from non-PR-trigger and skipped-matrix runs — the operator GET re-verification re-confirms the deterministic PR-published set |
| Protection apply is operator keep-set + GET re-verified (AC-CI-009) | E23 — the apply package does not exist on this tree (`test -e .moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` → exit 1) | Run phase packages payload+commands at `apply-package.md` (flips E23); operator applies and records the GET read-back in card evidence; run closes with the package delivered, not the apply done |
| Validator rejects parser failure / missing yq AND still passes the healthy input (AC-CI-010, two-directional) | E10 + E11 (failure shapes) + E22 (healthy baseline, exit 0) | Parser absence/malformed YAML → non-zero with a named message (E10/E11 re-run flip); the healthy pass E22 must REMAIN exit 0 with every dimension executing — an always-fail mutant fails E22, an always-pass mutant fails E10/E11 |
| Validator checks required-context publishability (AC-CI-011) | E24 + E27 (two-directional pair: identical fixture trees whose workflow sets publish every legit required name — `Lint` — and differ ONLY in the SSoT context value; the phantom passes silently, the control passes) | New dimension: every required context must appear among names the pull_request-triggered workflows publish (static name extraction from workflow `name:` fields + matrix); E24 re-run → non-zero naming the phantom, E27 (control) stays exit 0. A context-blind mutant (one that reads no contexts — e.g. keying only on a fixture file's absence) cannot separate E24 from E27, because the two trees are identical except for the context value; and an always-fail/always-pass mutant fails one side of the E22/E24/E27 triple |
| Detect filter covers parity-test input + correspondence guard (AC-CI-012) | E14 | Filter gains `.github/branch-protection.json.gtmpl`; a guard test fails CI when a repo-side test's read input is absent from the filter |

**M2 keep-set subsection (explicit, binding):** the apply step itself is NOT in any run-phase milestone. Deliverable = corrected SSoT + rendered payload + apply command + pre-apply live-diff command + post-apply GET command, packaged at `.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/apply-package.md` and handed to the operator/leader. DECLARED CLOCK INTERFACE (micro-card t1543 P2-M contract): M1's deadline implementation in the auto-merge merge step MUST read the env vars `current_time` and `merge_deadline` (epoch seconds) — probe `repro/run-deadline-mutant.sh` injects exactly these and observes the real `gh pr merge` call; an implementation reading any other clock surface will not flip E26 and the green path cannot be recorded. Evidence for AC-CI-009 is two separated parts (audit iter-2 P2): **(a) live-state check** — the package's pre-apply GET (live-diff) command is executed BEFORE the operator applies and reads back the CURRENT (uncorrected) protection values; because a run-phase agent application would have changed them, this read-back is the non-execution proof, not `git status` and not the transcript alone (an agent-first apply followed by an operator re-apply leaves an identical transcript); **(b) actor attribution** — the package names the actor per command (run phase: packaging only; operator/leader: apply + post-apply GET), and the operator's evidence paste into `.moai/reports/t1534/` carries the executed commands with their actor and the GET read-back output. The package's existence flips E23. Leader condition carried into the PR window: the develop-tip base `a158b4b5f` is sanctioned under current config; **merge-base must be re-verified at the PR window** (`CARD_BASE=$(git merge-base develop HEAD)` discipline, gitflow-lane-protocol §8) because the main-transition (t1453 #1751) is not yet landed.

### M3 — ci-watch verdict integrity (Priority Medium)

Files: `scripts/ci-watch/run.sh` (+ `lib/classify.sh`, `test/run_test.sh` as needed).

| Criterion | RED cell | Green path |
|---|---|---|
| Supported fields + JSON-array processing + bucket classification + exit-8 pending semantics (AC-CI-013) | E12 (Unknown JSON field: "status") + E12's available-fields list + E25 (exit 8 = "Checks pending") | Request `name,state,bucket,link`; classify by `bucket`; exit 8 with a valid JSON body = pending → continue waiting (E25); any NON-8 non-zero gh exit is fatal — never read as all-pass; E12's rejected field list now validates |
| Required completeness + `is_required` wired + no whitespace split + exit-8 pending classification (AC-CI-014) | E15 + E16 (+ structural: absent required checks never counted — the loop only iterates observed names; run.sh:122-124 aborts on ANY gh failure, exit 8 included — E25 documents that 8 means "Checks pending") | Required set from SSoT via `is_required` for the PR's base branch; expected-but-absent = pending; name iteration without IFS splitting; exit 8 + valid JSON classified as pending-and-continue, non-8 failures fatal — run_test.sh gains the exit-8 fixture pair (exit-8 mock continues; non-8 mock exits); run_test.sh extended and green |

Sequencing: M1 → M2 → M3 per the card's priority. M3 can proceed in parallel with M2's file edits (disjoint trees) but the keep-set package lands only after M2's SSoT edit is final.

## §G Anti-Patterns

- Verifying a repair by grepping for the fixed form — run both the old probe (E2/E4/E7/E10/E11 shapes) and the new path and observe them diverge.
- Treating `gh pr checks` text-table lines as data — auto-merge's loop greps prose; the repair must classify from structured output, not substrings like "fail" (a 404 message containing "error" would currently block; one not containing it would merge).
- Applying branch protection from uncorrected file values, or letting a run-phase agent execute the apply (REQ-CI-009).
- Counting absent required checks as absent-and-fine — absent means pending, per REQ-CI-013.
- Splitting check names on whitespace, or classifying required checks as "everything not auxiliary" (E15/E16).
- Hardening the gate in a way that makes an ordinary non-release PR fail: the detect-SKIPPED path remains a legitimate intentional exclusion (REQ-CI-001/002 distinction).
- Touching the reporting layer from the autofix loop (ci-autofix-protocol) — this SPEC's reporting-layer edits are planned SPEC work, not an autofix patch.

## §H Cross-References

- Requirements: `spec.md` §B (REQ-CI-001..013) · Acceptance: `acceptance.md` (AC-CI-001..014)
- Codex source (read-only, primary checkout): `.moai/reports/t1534/codex-ci-analysis-2026-10-06.md` — measured on `main@ec13872f3`; supports, does not replace, §B's own measurements
- SSoT consumers: `internal/cli/branch_protection.go:100` (render), `scripts/ci-watch/lib/classify.sh` (classify), `scripts/ci-mirror/validate-required-checks.sh` (validate)
- Rules: `.claude/rules/local/ci-watch-protocol.md` · `.claude/rules/moai/workflow/ci-autofix-protocol.md` · `.claude/rules/moai/development/verification-completeness.md` (two-cell discipline + §2.1 four-element RED cells) · `.claude/rules/local/gitflow-lane-protocol.md` §8 (merge-base discipline)
- Related SPEC: SPEC-V3R3-CI-AUTONOMY-001 (the SSoT + watch-loop origin) · decision-index.md Q1 (release/* list — amendment-2 reasoned resolution: published-only, empty under current triggers; trigger expansion out via t1536)
