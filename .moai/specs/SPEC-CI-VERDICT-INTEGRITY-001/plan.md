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

Allowed supporting files (only as needed): `scripts/ci-watch/lib/classify.sh` and `scripts/ci-watch/test/run_test.sh` (keep the M3 repair and its own test suite consistent); one new guard test file under `internal/template/` if the filter-correspondence check is authored as a separate test. Verified: none of the in-scope SSoT files has a mirror under `internal/template/templates/.github/` (only `branch-protection.json.gtmpl`, `labels.yml`, `actions/`, `workflows/` live there), so no Template-First regeneration is triggered; `.github/branch-protection.json.gtmpl` itself is shape-only (contexts injected at render time by `internal/cli/branch_protection.go:100`) and is NOT edited by this SPEC.

## §B Known Issues — verified RED evidence ledger

Document-level tree pin: **`a158b4b5f`** (verified via `git rev-parse --short HEAD` → `a158b4b5f`, exit 0, 2026-10-06). Every entry below was measured in this run, on this tree. Probe wrappers live under `/tmp/t1534-probe/` and are documented here; their content is a faithful, comment-marked expansion of the cited workflow step with GitHub template substitution applied by hand (values double-quoted as GHA renders them). `actionlint -color=false -shellcheck= .github/workflows/release-pr-multi-os.yml .github/workflows/auto-merge.yml .github/workflows/test-install.yml .github/workflows/ci.yml .github/workflows/codeql.yml` → no output, exit 0 (E18) — all five workflows are structurally valid; the defects are semantic.

| ID | Claim | Command (single invocation) | Verbatim stdout | Exit | Tree |
|----|-------|------------------------------|-----------------|------|------|
| E1 | Release gate never tests `cancelled` | `grep -c 'cancelled' .github/workflows/release-pr-multi-os.yml` | `0` | 1 | a158b4b5f |
| E2 | Gate with a cancelled matrix exits 0 + PASSED (release-pr-multi-os.yml:281-291 rejects only `failure`) | `sh /tmp/t1534-probe/gate-cancelled.sh` | `✓ Release PR multi-OS verification PASSED (or skipped for non-release / docs-only PR)` | 0 | a158b4b5f |
| E3 | Merge carries no head-commit pin | `grep -c 'match-head-commit' .github/workflows/auto-merge.yml` | `0` | 1 | a158b4b5f |
| E4 | Checks-lookup failure/empty output yields `should_merge=true` (auto-merge.yml:91-129 loop: `2>&1` capture + `|| true`, merge decision only from FAILED>0 / PENDING==0) | `sh /tmp/t1534-probe/checks-loop.sh` | `Check status: passed=0, pending=0, failed=0` / `All checks completed!` / `should_merge=true` | 0 | a158b4b5f |
| E5 | Job budget 20 min vs internal waits 10+15=25 min | `grep -n 'timeout-minutes\|MAX_WAIT' .github/workflows/auto-merge.yml` | `25:    timeout-minutes: 20` / `96:          MAX_WAIT=60` / `159:          MAX_WAIT=60   # 60 x 15s = 15 minutes` | 0 | a158b4b5f |
| E6 | Install summary never tests `cancelled` | `grep -c 'cancelled' .github/workflows/test-install.yml` | `0` | 1 | a158b4b5f |
| E7 | Summary with every result `cancelled` prints All tests passed, exit 0 (test-install.yml:352-356 reject only `failure`) | `sh /tmp/t1534-probe/install-summary-cancelled.sh` | `✅ All tests passed!` | 0 | a158b4b5f |
| E8 | Summary `needs` list (line 338) omits the `install-script-parity` job (defined at line 48) | `grep -n 'install-script-parity\|needs: \[' .github/workflows/test-install.yml` | `48:  install-script-parity:` / `338:    needs: [test-sh, test-ps1-pwsh, test-ps1-powershell, test-bat, compatibility-check]` | 0 | a158b4b5f |
| E9 | Required-feature probes print a missing marker but cannot fail (grep -q && echo ✓ \|\| echo ✗) — and the :309 install.sh probe's target is stale (E19) | `grep -n 'MOAI_INSTALL_DIR' .github/workflows/test-install.yml` | `309: … && echo "✓ MOAI_INSTALL_DIR support found" \|\| echo "✗ MOAI_INSTALL_DIR support missing"` / `317: …` (same shape) | 0 | a158b4b5f |
| E10 | Validator with yq absent from PATH: vacuous pass, exit 0 (get_yaml_list `2>/dev/null \|\| true` at :28 swallows the missing parser; empty lists skip every dimension) | `env PATH=/usr/bin:/bin sh scripts/ci-mirror/validate-required-checks.sh` | `=== Dimension A…` / (empty dimensions) / `✅ All validations passed` | 0 | a158b4b5f |
| E11 | Validator against a malformed-YAML SSoT: same vacuous pass, exit 0 (parser failure swallowed identically; wrapper `run-malformed.sh` cd's to `/tmp/t1534-probe/malformed` which carries a broken `.github/required-checks.yml`, then runs the repo's validator unmodified) | `sh /tmp/t1534-probe/run-malformed.sh` | `=== Dimension A…` / (empty dimensions) / `✅ All validations passed` | 0 | a158b4b5f |
| E12 | Installed gh CLI rejects the ci-watch field list — client-side, before any request | `gh pr checks 1748 --json name,status,conclusion,detailsUrl -R modu-ai/moai-adk` | `Unknown JSON field: "status"` + `Available fields:` `bucket, completedAt, description, event, link, name, startedAt, state, workflow` | 1 | a158b4b5f |
| E13 | Live published check names (PR 1748, read-only GET): `Analyze (Go) (go)` present, `Test (macos-latest)`/`Test (windows-latest)` absent, `Release PR Multi-OS Gate` present, `docs-site install-script parity` and `Test Summary` both publish | `gh pr checks 1748 --json name,state,bucket -R modu-ai/moai-adk` | JSON array (bucket=pass) containing `"name":"Analyze (Go) (go)"`, `"name":"Release PR Multi-OS Gate"`, `"name":"Test (ubuntu-latest)"`, `"name":"Test Summary"`, `"name":"docs-site install-script parity"`; no `Test (macos-latest)` / `Test (windows-latest)` / bare `CodeQL` | 0 | a158b4b5f |
| E14 | ci.yml detect filter has no branch-protection.json.gtmpl entry (parity test input at branch_protection_parity_test.go:41,57 never triggers the Test job when only that file changes) | `grep -c 'branch-protection.json.gtmpl' .github/workflows/ci.yml` | `0` | 1 | a158b4b5f |
| E15 | run.sh never calls `is_required` (SSoT whitelist classifier defined in lib/classify.sh:111 is unused; every non-auxiliary check is treated as required) | `grep -c 'is_required' scripts/ci-watch/run.sh` | `0` | 1 | a158b4b5f |
| E16 | Check names iterated via unquoted command substitution (whitespace split — `Test (ubuntu-latest)` fragments into 3 tokens) | `grep -n 'for check_name in' scripts/ci-watch/run.sh` | `134:    for check_name in $(_all_check_names "$TMP_JSON"); do` | 0 | a158b4b5f |
| E17 | Installed gh supports the merge-time head pin | `gh pr merge --help` | `--match-head-commit SHA   Commit SHA that the pull request head must match to allow merge` | 0 | a158b4b5f |
| E18 | actionlint baseline on the five in-scope workflows | `actionlint -color=false -shellcheck= .github/workflows/release-pr-multi-os.yml .github/workflows/auto-merge.yml .github/workflows/test-install.yml .github/workflows/ci.yml .github/workflows/codeql.yml` | (no output) | 0 | a158b4b5f |
| E19 | install.sh implements NO `MOAI_INSTALL_DIR` — the test-install.yml:309 probe targets a feature the script cannot express (stale probe; hard-checking it as-is would fail every install, the impossible direction). Real surface: `--install-dir` (:354), `-h\|--help` (:358-364), darwin/linux detection (:37-41) | `grep -c 'MOAI_INSTALL_DIR' install.sh` | `0` | 1 | a158b4b5f |
| E20 | install.ps1 DOES implement `MOAI_INSTALL_DIR` (env read at :307-308, set at :460) plus `IsWindows` (:287/:306) and `GetTempPath` (:221) — the :317 probe is live and stays | `grep -c 'MOAI_INSTALL_DIR' install.ps1` | `3` | 0 | a158b4b5f |

Codex-provenance items cited but NOT re-measured here (recorded as supports, not observations): the live branch-protection context values on `main` (codex used the leader-measured values and did not re-query the API; the protection GET needs admin scope and is packaged below as an operator-executable pre-apply diff step); real Actions-runner-level cancellation behavior (the E2/E7 probes are faithful template-substitution simulations of the gate scripts, not runner executions). Corroborating read: `internal/cli/branch_protection.go:100-105` renders any SSoT branch key into the protection payload, and `.github/required-checks.yml` lines 10-40 currently list `Test (macos-latest)`, `Test (windows-latest)`, `CodeQL` on both keys and omit `Release PR Multi-OS Gate` — E13 shows each of those four facts is wrong against live names.

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
- E4 gate surfaces: actionlint clean on touched workflows; validator E10/E11 probes now failing (exit non-zero) on the repaired tree; ci-watch E12 shape now succeeds with supported fields.
- E5 keep-set package: the apply payload + command + GET command as delivered (no execution by the agent).
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
| Single overall deadline < job timeout (AC-CI-005) | E5 | One deadline smaller than `timeout-minutes`; checks wait + review wait both bounded by it; E5's three-line read shows the new budget |
| Install summary success-only incl. parity (AC-CI-006) | E6 + E7 + E8 | `needs` gains `install-script-parity`; summary requires `success` from every need (cancelled/timed_out/skipped are non-success); E7's substitution on the repaired summary → non-zero |
| Feature-probe miss fails the step AND the probe list matches the scripts' real surface (AC-CI-007) | E9 + E19 + E20 | Required probes become hard checks (absent target → exit 1); the stale :309 install.sh probe (`MOAI_INSTALL_DIR`, unimplemented per E19) is realigned to the real surface (`--install-dir`); the :317 install.ps1 probe stays (E20); E9's shape no longer ends in `\|\| echo` for required probes |

### M2 — SSoT correction, parity wiring, validator hardening (Priority High; contains the keep-set gate)

Files: `.github/required-checks.yml`, `ci.yml` (detect filter), `validate-required-checks.sh`, `branch_protection_parity_test.go` (+ optional new guard test).

| Criterion | RED cell | Green path |
|---|---|---|
| SSoT lists only published names (AC-CI-008) | E13 vs current file values | Remove `Test (macos-latest)` / `Test (windows-latest)` / `CodeQL`; add `Analyze (Go) (go)`; add `Release PR Multi-OS Gate`; `release/*` mirrors the corrected `main` list (decision-index Q1, DEFAULT-APPLIED) |
| Protection apply is operator keep-set + GET re-verified (AC-CI-009) | n/a — process gate, not a tree-observable RED | Run phase packages payload+commands; operator applies and records the GET read-back in card evidence; run closes with the package delivered, not the apply done |
| Validator rejects parser failure / missing yq (AC-CI-010) | E10 + E11 | Parser absence/malformed YAML → non-zero with a named message; E10/E11 re-run → non-zero |
| Validator checks required-context publishability (AC-CI-011) | E10 (empty required list passed silently; no dimension inspects `branches.*.contexts` for publishability) | New dimension: every required context must appear among names the workflows publish (static name extraction from workflow `name:` fields + matrix), CI-wired |
| Detect filter covers parity-test input + correspondence guard (AC-CI-012) | E14 | Filter gains `.github/branch-protection.json.gtmpl`; a guard test fails CI when a repo-side test's read input is absent from the filter |

**M2 keep-set subsection (explicit, binding):** the apply step itself is NOT in any run-phase milestone. Deliverable = corrected SSoT + rendered payload + apply command + pre-apply live-diff command + post-apply GET command, handed to the operator/leader. Leader condition carried into the PR window: the develop-tip base `a158b4b5f` is sanctioned under current config; **merge-base must be re-verified at the PR window** (`CARD_BASE=$(git merge-base develop HEAD)` discipline, gitflow-lane-protocol §8) because the main-transition (t1453 #1751) is not yet landed.

### M3 — ci-watch verdict integrity (Priority Medium)

Files: `scripts/ci-watch/run.sh` (+ `lib/classify.sh`, `test/run_test.sh` as needed).

| Criterion | RED cell | Green path |
|---|---|---|
| Supported fields + JSON-array processing + bucket classification (AC-CI-013) | E12 (Unknown JSON field: "status") + E12's available-fields list | Request `name,state,bucket,link`; classify by `bucket`; gh failure → exit 1 (never all-pass); E12's field set now validates |
| Required completeness + `is_required` wired + no whitespace split (AC-CI-014) | E15 + E16 (+ structural: absent required checks never counted — the loop only iterates observed names) | Required set from SSoT via `is_required` for the PR's base branch; expected-but-absent = pending; name iteration without IFS splitting; run_test.sh extended and green |

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
- Related SPEC: SPEC-V3R3-CI-AUTONOMY-001 (the SSoT + watch-loop origin) · decision-index.md Q1 (release/* list, DEFAULT-APPLIED)
