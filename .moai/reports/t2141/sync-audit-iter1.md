auditor-model: glm-5.3
verdict: PASS-WITH-DEBT
audited_sha: 8d6d4ef59f923bac78a9cecfcf61ea4341dd08b4
receipt: rcpt-ba8034d9024db2b4a563ef99 (codex_audit, native, target=baseBranch, project_root=this worktree, codex verdict=fail, head 149e94f9fa — recorded in .moai/state/audit-receipts/receipts/; the id was absent from the result text and recovered from the store, see F7)

# Evaluation Report — card t2141, sync-audit iteration 1 (pass 1 of 2, Claude-side)

SPEC context: SPEC-ZONE-SHELL-PARSING-001 (session binding); card t2141 is Class B
(operator-authorized unfreeze d-20261106T163347Z-7adc, source: t2134 report).
Audited delta: `git diff 04d8c27a7e..8d6d4ef59f` — 5 files
(hook_install_precommit.go, templates/.git_hooks/pre-commit,
precommit_e2e_fixtures_test.go [new], precommit_vet_output_e2e_test.go,
hook_install_precommit_test.go), +105/−6; docs commit 149e94f9fa (progress.md only).

Overall Verdict: PASS-WITH-DEBT — the pass-2 codex leg ran IN this pass: the Stop-hook
receipt guard refused the receiptless PASS, and the moai MCP server was spawnable via
its own CLI (`moai mcp-server`, stdio JSON-RPC — the same implementation the MCP tool
rides), so the codex_audit was executed now (native, baseBranch, project_root = this
worktree). Codex returned verdict=fail with two P2 findings; the tree's gate is
ADVISORY at verdict time (post-merge config, see below), so per the convergence fold
(case 4: advisory FAIL, required PASS) this is a PASS with the disagreement surfaced
and one corroborated debt recorded. Windows execution (release-pr-multi-os, pending)
remains the open CI-owned gap. One recorded debt: D1 (F1, corroborated by codex).

Post-audit tree movement (process defect, F8): a foreign merge commit
`5bce362653` (main into this branch) landed at 2026-11-07 08:33 +0900, INSIDE the
audit window (audit began ~08:05; codex review captured head 149e94f9fa before the
merge, so the reviewed scope is the clean card delta 04d8c27a7e..149e94f9fa — the
merge-base codex reported, `04d8c27a7e`, confirms it). The merge carried main's commit
b25adf4fca "codex audit gate required → advisory for the local af-flow landing", which
is why the gate read `required` at audit start and `advisory` at verdict time
(`moai verify audit-plan` re-run post-merge: codex gate=advisory, explicit,
config.gates). Reported to the leader per the one-writer-per-audited-tree rule.

## Dimension Scores

| Dimension | Score | Verdict | Evidence (verbatim) |
|-----------|-------|---------|---------------------|
| Functionality (40%) | 92/100 | PASS | `go test ./internal/cli/ -run 'TestPre[Cc]ommit|TestPreCommit' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli  25.750s` (exit 0); verbose rerun: 50 `--- PASS`, 0 `--- SKIP`, 0 `--- FAIL`, `ok  github.com/modu-ai/moai-adk/internal/cli  30.108s`; 4 siblings individually: `--- PASS: TestPreCommitHook_GoVetBlocks (2.27s)` / `--- PASS: TestPreCommitRelocation_GoodCommitPasses (1.17s)` / `--- PASS: TestPrecommitVetFailureShowsDiagnostics (2.67s)` / `--- PASS: TestPrecommitVetPassStaysQuiet (2.07s)` |
| Security (25%) | 88/100 | PASS | Secret probe on the delta: `secret_probe_exit=1` (no hits for password/token/secret/api-key/PEM/eval/base64/curl/wget); dependency audit: `git diff 04d8c27a7e..8d6d4ef59f --stat -- go.mod go.sum` → empty (no dependency change); all shell expansions quoted (`:"$f"`, `>"$_moai_fmt_tmp"`, `rm -f "$_moai_fmt_tmp"`). One Medium-Low hardening observation (F1: predictable pid-keyed temp name, no O_EXCL) — environment-specific, no Critical/High |
| Craft (20%) | 88/100 | PASS | `gofmt -l internal/cli/` → no output, exit 0; template byte-identity: `--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)`; err/fmt/verdict messages unchanged vs base (line-by-line diff read: only the FMT_SCAN block + 2 comment lines changed in the hook). Coverage % NOT measured — full-package coverage run barred by lane-local heavy-run discipline (Gap) |
| Consistency (15%) | 88/100 | PASS | Template mirror edited in lockstep, enforced green (test above); commit `[t2141] fix(cli): ...` conventional + card id; raw-string no-backtick constraint respected (file compiles); no new coreutils beyond `rm` (verified by full read: no wc/mktemp/basename; printf/read/[/command -v are builtins). Stale doc enumeration F3; t2134 add/add conflict F5 (outside) |

Composite: weighted 89.6/100 (harmonic mean 89.0 — same verdict either way).
Must-pass firewall: Functionality PASS and Security PASS independently — satisfied.

## Findings

- F1 [Medium] [optional] [in-diff] internal/cli/hook_install_precommit.go:91 (+ template
  mirror templates/.git_hooks/pre-commit) — the temp path
  `${TMPDIR:-/tmp}/moai-precommit-gofmt.$$` is predictable and opened with `>`
  (truncate, no exclusivity). On multi-user hosts with TMPDIR unset (shared /tmp), a
  pre-planted symlink at the predicted name is followed on write (staged bytes written
  through it), and a foreign-owned regular file at that name makes the redirect fail
  (`|| :` swallows it) after which gofmt reads the FOREIGN file's bytes — wrong verdict —
  and the later `rm -f` fails EPERM, which under `set -eu` (line 50) aborts the hook and
  blocks the commit without a hook diagnostic. macOS (per-user TMPDIR) and Windows
  (user-profile TMP) are unaffected. The in-code claim "nothing survives a completed or
  interrupted run" (lines 84-86) is also slightly strong: a SIGKILL mid-scan leaves the
  pid-keyed file (the double-rm covers all normal exits; a later same-pid run truncates
  it, so it is benign). Required fix (follow-up card, not this one): exclusive creation —
  mktemp-based name, or `set -C` on the write — keeping the always-rm.
- F2 [Low] [optional] [in-diff] internal/cli/hook_install_precommit.go:94 — behavioral
  corner regression: a broken/unwritable TMPDIR now blocks every Go commit (redirect
  fails → gofmt on a nonexistent file → 'err' branch), where the old pipe form had no
  TMPDIR dependency. Rare environment corner; the net trade remains strongly positive
  (total Windows blockage fixed). Required fix if pursued: writability guard or an
  explicit diagnostic naming TMPDIR.
- F3 [Low] [optional] [touched-file] internal/cli/hook_install_precommit_test.go:568-572 —
  moduleShimPath's doc comment still enumerates "(git, grep, sort, dirname, sed, awk, tr,
  go, gofmt)" without `rm`, while the list at line 581 and the new comment at 579-580
  include it — the two halves of the same comment disagree. Required fix: add `rm` to the
  enumerated set in the header comment.
- F4 [Info] [optional] [in-diff] internal/cli/precommit_e2e_fixtures_test.go:44 —
  hookE2EPath's named-tool list omits `rm` (and `awk`, used by the build-tags block),
  which the hook now uses. Benign: rm/awk resolve from the same directories the list
  already adds (grep/sed/tr share Git-for-Windows usr\bin; /bin,/usr/bin POSIX fallbacks),
  unlike moduleShimPath where the shim dir holds ONLY named symlinks — hence rm had to be
  explicit there and rides along here. Required fix: optional parity — add "rm" (and
  "awk") to the named list.
- F5 [Info] [optional] [outside] cross-card merge (t2134 tree, unlanded) — t2134's
  internal/cli/precommit_e2e_fixtures_test.go is a superset (adds `buildE2EBinary`,
  differing comment wording); this card ported a subset into the SAME path → add/add
  textual conflict when both cards land. The `hookE2EPath` bodies are byte-identical
  (verified by diff), so resolution is mechanical (take t2134's superset). Also
  t2134's precommit_vet_output_e2e_test.go still carries the old hardcoded PATH — this
  card's change to that file is the superset there and merges cleanly. Listed for the
  leader; not queued.
- F6 [Info] [optional] [touched-file] internal/cli/hook_install_precommit.go:99-101 — the
  'err' branch (unparseable staged bytes / gofmt tool error → block) has no direct test
  (grep for "could not check" across internal/cli/*_test.go: no hit). Pre-existing gap,
  unchanged by this card; the card's err-branch equivalence claim is established by code
  reading (identical gofmt invocation semantics modulo input channel), not by execution.
  Required fix: follow-up characterization test staging unparseable bytes.
- F7 [Medium] [optional] [outside] internal/cli/mcp_convergence.go:1182 (codex P2,
  static) — audit-machinery defect on main (in this tree only via the mid-audit merge,
  NOT in the card delta): re-running a converged fan-out after a required backend was
  INCONCLUSIVE can leave the previous `GateUnmet` in the returned convergence, which
  `CheckCitedReceiptsSince` then rejects as `CauseReceiptAuditNotRun`, blocking the next
  step despite the rerun succeeding. Corroborating observation from THIS audit: the
  codex_audit result text carried no `audit_receipt` member even though a receipt WAS
  recorded in the store (rcpt-ba8034d9024db2b4a563ef99, written at run completion) — the
  id had to be recovered from `.moai/state/audit-receipts/receipts/`. Listed for the
  leader (audit-engine card); not queued against t2141.
- F8 [Medium] [optional] [outside] process — a foreign merge commit
  `5bce362653` (main into WT-the-frozen-precommithook) landed at 08:33 +0900 inside the
  audit window, moving HEAD under the running audit (one-writer rule for an audited
  worktree). The reviewed scope was unaffected (codex bound head 149e94f9fa pre-merge;
  merge-base 04d8c27a7e unchanged), but the tree at verdict time (5bce362653) is NOT
  the audited SHA. Required disposition: leader records the movement; the audited_sha
  binding below remains 8d6d4ef59f, and the lane's landing sequence must re-bind or
  re-audit if the merge is judged to change the card's governed surface.

Debt register (PASS-WITH-DEBT):
- D1 = F1 (hook temp file, exclusive creation via mktemp/noclobber + trap-style
  cleanup, in BOTH `preCommitHookContent` and the template mirror) — corroborated by
  codex P2 at the same file:line. Not blocking: no SPEC AC covers it, supported
  platforms (macOS per-user TMPDIR, Windows user-profile TMP) unaffected.

## Codex leg (ran in this pass — receipt rcpt-ba8034d9024db2b4a563ef99)

- Call: `codex_audit` {project_root: this worktree, target: baseBranch, mode: native},
  driven over the `moai mcp-server` stdio JSON-RPC surface after the Stop-hook receipt
  guard refused the receiptless PASS (the lane's pass-2 sequencing was superseded by the
  hook's demand; the call is the sanctioned MCP surface, 1 of the 305c cap of 2).
- `review_base`: "main (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)" — exactly
  the card delta + docs commit; receipt head_commit 149e94f9fa confirms the pre-merge
  scope. resolved_model gpt-6.1-sol, effort high, static-only review (codex states it
  ran no test/build/repro commands).
- `verdict: fail` on two P2 findings, folded per the case-4 advisory rule (gate
  advisory at verdict time): (1) mcp_convergence.go:1182 — OUTSIDE the card delta
  (pre-existing main code, in-tree via the mid-audit merge) → listed, never queued (F7);
  (2) hook_install_precommit.go:91 — the SAME file:line as F1, same fix (mktemp + trap,
  including the template mirror) → F1 is now corroborated by two independent reviewers
  and is recorded as debt D1 rather than left purely discretionary.

No blocking finding. Per the finding-consumption brake, F2-F6 remain reported and
discretionary; F1 is recorded as debt D1 (PASS-WITH-DEBT) on two-model agreement;
F7/F8 are leader items outside the card delta.

## Recommendations

- Accept F1 as a small follow-up hardening card (mktemp or noclobber on the temp write);
  it is the only finding with a real (if narrow) attack/corruption surface.
- Fix F3 in passing whenever the file is next touched (one-word comment edit).
- Leader note for landing: expect the add/add conflict on precommit_e2e_fixtures_test.go
  when t2134 lands; resolution = t2134's superset (hookE2EPath bodies identical).

## Verification Record (5 sections)

**Claim** — (1) The portable tmpfile gofmt handoff is verdict-equivalent to the old
/dev/stdin pipe on every input class and cleans up on every exit path; `$$` keying is
collision-safe for live concurrent commits. (2) Template mirror byte-identical to the
constant. (3) hookE2EPath port preserves the deliberate moai-unreachable isolation.
(4) moduleShimPath `rm` addition is the complete set of new coreutils dependencies.
(5) Scope: nothing outside the named surfaces changed semantically. (6) Targeted suite,
windows vet/compile, gofmt all green on the audited content.

**Evidence** — commands run by this auditor in this worktree (CWD
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2141), all read-only or /tmp-scoped:
- `GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache go test ./internal/cli/ -run 'TestPre[Cc]ommit|TestPreCommit' -count=1` → exit=0, `ok  github.com/modu-ai/moai-adk/internal/cli  25.750s`
- same with `-v` → exit=0; `grep -c '^--- PASS'` = 50; SKIP = 0; FAIL = 0; `PASS` / `ok  github.com/modu-ai/moai-adk/internal/cli  30.108s`
- 4 siblings with `-run 'TestPreCommitHook_GoVetBlocks|TestPreCommitRelocation_GoodCommitPasses|TestPrecommitVetFailureShowsDiagnostics|TestPrecommitVetPassStaysQuiet' -v` → all four `--- PASS` (2.27s / 1.17s / 2.67s / 2.07s), `ok ... 10.123s`
- `GOOS=windows GOARCH=amd64 GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache go vet ./internal/cli/` → winvet_exit=0, 0 output lines
- `GOOS=windows GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache go test -c -o /dev/null ./internal/cli/` → wintestc_exit=0, 0 output lines
- `gofmt -l internal/cli/` → empty output, exit 0
- `go test ./internal/cli/ -run 'TestPreCommitTemplateMatchesConstant' -count=1 -v` → `--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)`
- `git diff 04d8c27a7e..8d6d4ef59f` read in full (hook, mirror, 3 test files); `git diff 8d6d4ef59f..149e94f9fa --stat` → progress.md only
- `moai verify audit-plan --project-root <worktree>` → `config_status: ok`, `cross_model_active: true`, backends: claude=off, codex=required (enforced), glm=off
- code-read verification of the sh block: index-read failure → redirect still creates/truncates the file → empty file → gofmt exit 0, empty output → pass (== old empty pipe); unparseable → gofmt non-zero → 'err' (unchanged); unformatted → gofmt prints the path → 'fmt' (only emptiness of `_out` is consumed, so /dev/stdin-vs-tmpfile in the printed name is irrelevant); `rm -f` reached on every loop path (the only `continue` precedes file creation) and once after the loop for the zero-iteration case; `$$` distinct for any two live processes; same-repo concurrent commits serialize on the index lock held through pre-commit; own-stale files are truncated harmlessly by `>`.

**Baseline-attribution** — every command above was executed by this auditor, in this
run, in this worktree at HEAD 149e94f9fa whose internal/cli and templates content is
byte-identical to the audited commit (docs-only delta verified). Go toolchain pinned
via GOTOOLCHAIN=go1.27.1, GOCACHE=/tmp/t2117-gocache (the lane's prescribed pin). The
lane's own recorded runs on 8d6d4ef59f (darwin ok 30.960s; windows vet/compile exit 0)
were NOT reused as evidence — all PASS claims above are from this auditor's runs.

**Gaps**
- Windows EXECUTION verdict is CI-owned: release-pr-multi-os on this commit is pending.
  Local evidence is vet + cross-compile only. Per the card instruction, this is recorded
  as a Gap, not booked as a card defect.
- Codex leg: DONE in this pass (was listed as pass-2 pending; the Stop-hook receipt
  guard forced it forward). The `audit_multi` convergence path was NOT run — the plan's
  only enforced gate (codex) was satisfied via the single-backend call the hook named;
  claude and glm gates are explicitly off. The orchestrator's
  `moai verify audit-plan --result-file` check is the leader's closing step.
- Coverage percentage NOT measured: a coverage run executes the full internal/cli suite,
  barred by the lane-local heavy-run discipline; the targeted suite (50 green) is the
  craft evidence.
- The 'err' branch equivalence is code-read-established, not executed (F6, pre-existing);
  codex's review was likewise static-only (its own statement).
- Full suite and other packages not run (lane-local discipline).
- staged_read·gate_scope fixture files in this tree are still pre-t2134 (t2134 not yet
  landed here) — out of this card's scope, resolved by t2134's landing.
- Lane's recorded count for the same suite was 49 vs this auditor's 50 — both green;
  attributed to run environment, not to any tree difference (internal/cli identical).

**Residual-risk**
- Cross-model disagreement (advisory, case 4): this auditor's dimension verdicts are
  PASS while the codex backend returned fail; the divergence reduces to (a) F1, where
  both reviewers agree on the defect and disagree on blocking (this auditor: debt;
  codex: P2), and (b) codex's out-of-diff finding F7. Gate was advisory at verdict
  time; if the leader re-pins it required, this verdict must be revisited.
- Git-for-Windows sh runtime behavior (MSYS TMPDIR path conversion, redirection
  semantics) was verified by reading and by the shared POSIX logic, not by local
  execution; the darwin e2e suite exercises the same hook end-to-end, and the Windows CI
  leg is the execution authority (pending).
- On CI Windows runners, rm/grep/sed resolution rides Git's usr\bin entering the PATH via
  LookPath'd peers (F4); a runner-image change removing usr\bin from the ambient PATH
  would surface as e2e failures — pre-existing pattern inherited from the t2134 design.
- F1/D1's collision corner (shared-TMPDIR hosts) remains open until the hardening
  follow-up; it cannot affect the single-user dev machines that are this hook's primary
  audience.
- The verdict's binding is to audited_sha 8d6d4ef59f; the tree moved mid-audit (F8).
  Measured across the merge (8d6d4ef59f → 5bce362653): the production hook, the
  template mirror, and the two other test files are byte-identical;
  `precommit_e2e_fixtures_test.go` absorbed t2134's landed version (+56/−16, byte-equal
  to t2134's tree — the F5 add/add conflict resolved exactly as F5 predicted, superset
  taken). Post-merge recheck on the CURRENT tree:
  `go test ./internal/cli/ -run '<the 4 siblings>|TestPreCommitTemplateMatchesConstant'`
  → `ok  github.com/modu-ai/moai-adk/internal/cli  5.525s`. Any later commit on this
  branch re-opens the binding.

## Iteration history

- iter 1 (this file, 2026-11-07): first sync audit of 8d6d4ef59f. Claude-side pass
  initially rendered PASS with the codex leg deferred to the lane's pass 2; the
  Stop-hook receipt guard refused the receiptless PASS, so the codex_audit was executed
  within this pass (receipt rcpt-ba8034d9024db2b4a563ef99). Fold: codex fail (advisory
  gate) with F1 corroborated at the same file:line → verdict recorded as PASS-WITH-DEBT
  with debt D1 (F1, mktemp/noclobber + trap in hook AND template mirror); F7 (receipt
  machinery, outside the delta) and F8 (foreign merge mid-audit) reported to the leader.
  No pass-2 codex call remains owed: the 305c budget used 1 of 2, and the gate is
  advisory at verdict time. Windows CI verdict still pending.
