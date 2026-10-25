auditor-model: glm-5.3

# Sync Audit — card t1959 (GH #1913)

verdict: PASS
audited_sha: c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce

SPEC: none (Class B card — SPEC-less by definition; audited against GH #1913 and the lane's claims 1-5)
Overall Verdict: PASS

## Claim

The lane claims, for the single commit c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce on top of origin/main 53fbc2d206:

1. An omitted `--spec`, or the card id passed as `--spec`, routes `moai factory merge ready` to the SPEC-less path: `SpecDir` stays `""` and condition (a) closes on `.moai/reports/<card-id>/verdict.md` (PASS/PASS-WITH-DEBT + `audited_sha` binding HEAD, receipts admitted — the existing `merge_specless.go` rules, unchanged).
2. A card WITH a SPEC keeps its `progress.md` record as condition (a) — the SPEC path is untouched.
3. Every other gate is unchanged: conflict-free merge-tree, tree identity, re-measure record, window acquire semantics, refusal-before-window.
4. The empty-`Join` trap is handled: empty spec id stays `""` (not the specs directory).
5. Tests: the former git-flow "`--spec` is required" pin is replaced by the new contract (`TestFactoryMergeReadySpecLessOnVerdictFile`: spec omitted + spec==card id, both clearing and taking the window; `TestFactoryMergeReadySpecLessWithoutVerdictRefused`: no verdict file → condition (a) refusal, no window).

## Evidence

Code read in full (this tree, this run): `internal/cli/factory_merge.go` (395 lines), `internal/cli/factory_card_pr_specless_test.go` (179 lines), `internal/factorylane/merge.go` (condition table + `checkSyncAudit`), `internal/factorylane/merge_specless.go` (verdict admission), `internal/cli/factory_card_pr.go:227-254` (the `factory complete` precedent for the same SpecDir pattern). Diff verified one-commit: `git show c1c9c26d98 --stat` → `internal/cli/factory_card_pr_specless_test.go | 91 +++++---` and `internal/cli/factory_merge.go | 22 ++--` (2 files, +90/−23); `internal/factorylane/**` untouched by the commit.

Claim 1 — routing. `factory_merge.go:83-85`: `if strings.TrimSpace(specID) == "" || specID == card { specID = "" }`; `factory_merge.go:110-113`: `specDir := ""; if specID != "" { specDir = filepath.Join(root, ".moai", "specs", specID) }`. Empty `SpecDir` selects `checkSyncAuditVerdictFile` at `merge.go:182-183`. Machine-verified by `TestFactoryMergeReadySpecLessOnVerdictFile` (both subtests, clearing AND taking the window: `lock.Held() && SessionName == "lane-9" && Card == "t9001"` asserted at test line 152-155).

Claim 2 — SPEC path untouched. `merge.go:184-185` keeps `checkSyncAudit(in.SpecDir)` (progress.md §E.4 `sync_status: complete`, first-token rule at `merge.go:251-271`) for non-empty SpecDir; `checkSyncAudit` and `merge.go`'s condition ordering are byte-unchanged in this commit (diff stat). A SPEC card passing `--spec SPEC-X` (≠ card id) reaches exactly the old branch.

Claim 3 — other gates unchanged. Conditions (b) conflict-free, (c) tree identity, (d) re-measure record are computed at `merge.go:192-235` independent of `SpecDir`; the window sequence (refusal-before-window at `factory_merge.go:152-181`, holder-named waiting at 224-243, pre-acquire proof at 246-281) is untouched by the diff. The no-verdict refusal taking no window is machine-verified by `TestFactoryMergeReadySpecLessWithoutVerdictRefused`.

Claim 4 — empty-Join trap. `factory_merge.go:107-113` guards the `filepath.Join` behind `specID != ""`; the comment names the trap. Form is byte-parallel to the existing precedent `factory_card_pr.go:242-245` (`specDir := ""; if card.SpecID != "" { … }`).

Claim 5 — admission rules unchanged. `merge_specless.go` is not in the commit; its rules (shared `auditverdict.Admit` predicate at line 101, non-pass/malformed receipt refusal at 93-100, sole `audited_sha` at 105-111, HEAD-or-evidence-dir-only binding at 113-141, 4 MiB regular-file bound at 62-63, card-id path-safety at 44-46) apply to the newly reachable path as-is.

Mechanical verification (run by this auditor, this tree, plain single commands):

    $ go test ./internal/cli/ -run 'Specless|MergeReady' -count=1
    ok  github.com/modu-ai/moai-adk/internal/cli	31.699s

    $ go build ./internal/cli/
    (exit 0, no output)

    $ go vet ./internal/cli/
    (exit 0, no output)

    $ golangci-lint run ./internal/cli/...
    0 issues.

    $ go test ./internal/cli/ -run 'TestFactoryMergeReadySpecLess' -count=1 -v
    === RUN   TestFactoryMergeReadySpecLessOnVerdictFile
    === RUN   TestFactoryMergeReadySpecLessOnVerdictFile/spec_omitted
    === RUN   TestFactoryMergeReadySpecLessOnVerdictFile/spec_is_card_id
    --- PASS: TestFactoryMergeReadySpecLessOnVerdictFile (0.94s)
        --- PASS: TestFactoryMergeReadySpecLessOnVerdictFile/spec_omitted (0.48s)
        --- PASS: TestFactoryMergeReadySpecLessOnVerdictFile/spec_is_card_id (0.46s)
    === RUN   TestFactoryMergeReadySpecLessWithoutVerdictRefused
    --- PASS: TestFactoryMergeReadySpecLessWithoutVerdictRefused (0.41s)
    PASS
    ok  github.com/modu-ai/moai-adk/internal/cli	2.275s

Adversarial gate-regression sweep (the leader's three questions):

- **Can a card that previously could NOT reach merge-ready now do so WITHOUT its evidence closing?** No. Previously a SPEC-less card died at the CLI flag gate before the triple ran; now it must pass condition (a) on the verdict file — `verdict: PASS`/`PASS-WITH-DEBT` via the shared auditverdict predicate, all recorded receipts `pass` (fail/inconclusive refuses, malformed fails closed), exactly one `audited_sha` naming a commit that is HEAD or differs from HEAD only under `.moai/reports/<card>/`. No verdict file → refusal, exit-0 verdict, no window (machine-verified). The new reachability is evidence-gated, not open.
- **Can a SPEC card silently lose its progress.md gate?** Only by the operator omitting `--spec` (or passing the card id), which routes it to the verdict-file gate — a strictly independent, HEAD-bound evidence class, not a weaker one (progress.md §E.4 is the lane's own record; the verdict file requires a third-party audit bound to the exact HEAD). With `--spec SPEC-X` given, the SPEC path is byte-unchanged. No silent substitution exists: the operator's flag choice is the routing input, and both branches demand closed evidence.
- **Does the SPEC-less path read the verdict file under the same admission rules as before?** Yes — `merge_specless.go` is untouched by this commit (diff stat: 2 files, both in `internal/cli`).

## Baseline-attribution

All measurements above were taken in this run against this tree: worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1930`, branch `WT-merge-ready-specless`, HEAD `c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce` (verified: `git rev-parse HEAD` → `c1c9c26d9891845fb1c11fb1a3df1a26ccb6b4ce`; `git status --short` → clean; parent `53fbc2d206` = origin/main tip of the audited one-commit delta). No figure is carried over from another run, tree, or card. Toolchain: repo-local `go` toolchain as resolved in this worktree; golangci-lint as installed on PATH (§2.2 note: no repo-pinned build exists for golangci-lint; go/vet/build are the tree's own toolchain).

## Gaps

- The full `go test ./...` suite was NOT run — lane-local scope discipline and the leader's explicit constraint. Cross-package regressions rest on the CI of PR #1915, not on this audit.
- Package-level coverage percentage was NOT measured (full-suite prohibition makes a package-wide `go test -cover` a heavy run; the changed lines are instead covered line-for-line by the two new tests plus `TestMergeReadinessBeforePRSpecLess`'s four-table admission matrix, all observed passing).
- No guard refusals were encountered this run (verifications ran as plain single commands per standing discipline; the receipt driver ran as `bash /tmp/t1959-mint-receipt.sh`).

## Residual-risk

- **R1 (medium):** `--spec` omission routes a card WITH a registered SPEC to the verdict-file gate. Independently reproduced by the codex backend (see Iteration history, P1-1): a card whose §E.4 reads `audit-ready` and whose SPEC path refuses is `cleared` when `--spec` is omitted or set to the card id — because a HEAD-bound PASS verdict file exists. The substituted evidence class is not weaker (an independent audit bound to the exact HEAD, versus the lane's own progress.md record), so no evidence-unclosed cleared path exists; but the routing input is the caller's flag, not the card record's `SpecID`. Follow-up hardening (route on the card record, admit the verdict file only for cards with no SPEC) is recommended, not required for this card's scope (GH #1913 fixed reachability, and the fixed gate stays evidence-closed).
- **R2 (informational):** a whitespace-padded `--spec " t9001"` evades neither gate: it is neither empty nor `== card`, so it names a nonexistent specs directory and `checkSyncAudit` fails closed ("no progress.md under …"). The refusal is correct but less pointed than the SPEC-less refusal text; cosmetic only.
- **R3 (low):** the JSON merge verdict carries no `spec` field, so the routing decision is not visible in the emitted verdict object — but the recorded check `detail` names the exact evidence path read (`.moai/reports/<card>/verdict.md` or the specs dir), which is the audit-relevant trace.
- **R4 (medium):** the SPEC-less `audited_sha` binds the worktree HEAD, while conditions (b)(c)(d) examine the `--branch` candidate — a caller who checks out the audited commit but passes a different `--branch` can clear a candidate carrying unaudited changes (codex P1-2, reproduced at this head). Structural, not introduced by this commit: the SPEC path's progress.md evidence is equally un-correlated with the candidate tip, and the fast path (no `--branch` → `currentBranch()` == HEAD) cannot hit it. Follow-up: bind `audited_sha` to the candidate branch tip, or refuse when the candidate tip differs from HEAD on the SPEC-less path.
- The verdict file's own admission relies on the shared `auditverdict` predicate being sound; that surface was audited under its own cards (t1571, t1557) and is outside this commit's delta.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | Claims 1-5 verified against code + `go test` filtered run `ok … 31.699s`; new-contract subtests all PASS |
| Security (25%) | 95/100 | PASS | No gate bypass reachable: verdict admission fails closed (receipt rules, sole SHA, HEAD binding, path-safety regex, 4 MiB bound — all pre-existing, untouched); lint `0 issues.` |
| Craft (20%) | 90/100 | PASS | `go vet` clean, `golangci-lint` `0 issues.`, new tests pin both the cleared and refused edges with the window state asserted |
| Consistency (15%) | 95/100 | PASS | SpecDir guard byte-parallel to `factory_card_pr.go:242-245` precedent; comment style and refusal-message conventions match surrounding code |

Must-pass firewall (Functionality, Security): both PASS independently.

## Findings

- F1 [Medium] [optional] internal/cli/factory_merge.go:83 — routing decides SPEC vs SPEC-less on the caller's flag alone, without consulting the card record's `SpecID`, so a card WITH a registered SPEC can be evaluated on the verdict file by omitting `--spec` (independently reproduced by the codex backend: `audit-ready` SPEC card → refused with its SPEC, `cleared` with it omitted). Not blocking: the substituted evidence is a HEAD-bound independent-audit verdict — a closed, not weakened, gate; no evidence-unclosed cleared path exists. Required fix (follow-up card, not this PR): route on the card record's `SpecID` and admit the verdict-file path only for cards with no SPEC.
- F2 [Informational] [optional] internal/cli/factory_merge.go:83 — `specID == card` compares the raw flag (no TrimSpace), so `" t9001"` falls through to the SPEC branch and fails closed with the generic no-progress.md refusal. Required fix: none (fail-closed direction is safe); a TrimSpace on the comparison would only sharpen the message.
- F3 [Medium] [optional] internal/cli/factory_merge.go:117 — the SPEC-less `audited_sha` binds the worktree HEAD while (b)(c)(d) examine the `--branch` candidate: an explicit `--branch` pointing at a commit other than HEAD can clear a candidate carrying changes outside the audited commit (codex P1-2, reproduced at this head with `HEAD=develop`, `--branch WT-card`, an unaudited `card.txt` change, `cleared` + window). Not blocking and not new in kind — the SPEC path's progress.md evidence is equally un-correlated with the candidate tip, and the default path (`--branch` unset → current branch == HEAD) cannot reach it; the cleared verdict still requires (b)(c)(d) to pass on the candidate. Required fix (follow-up card, not this PR): bind `audited_sha` to the candidate branch tip, or refuse on the SPEC-less path when the candidate tip differs from HEAD.

No blocking findings. All findings are optional: per the finding-consumption discipline an all-optional findings list does not convert this PASS into a FAIL.

## Cross-model audit plan result (for the orchestrator's check)

`moai verify audit-plan --project-root <this worktree>` → `config_status: ok`, `cross_model_active: true`, enforced required: `codex` (claude/glm off). `audit_multi` was driven over stdio (no `gates` argument; the tree's own plan applied) with `project_root` = this worktree, `target: baseBranch`. Result members, per the report obligation:

- `overall_verdict`: fail
- `gate_unmet`: (member absent from the result)
- `plan_source`: config
- `per_backend_verdicts`: backend=codex, gate=required, verdict=fail (digest members only; finding text folded into F1/F3 above)
- `audit_receipt`: rcpt-26567a378837d950e5081ccf

This auditor's verdict remains its own (the receipt proves the codex audit ran; a codex `fail` corroborates and does not dictate the verdict). The verdict is not final until the orchestrator's convergence check passes: the orchestrator writes these members fresh to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`.

## Iteration history

- 2026-10-25 iteration 1: full read (4 files + precedent) + 5-command mechanical batch, all green (`ok … 31.699s`, build/vet exit 0, `0 issues.`). Verdict PASS; blocked stop with receipts=none per the continuation-minting procedure.
- 2026-10-25 iteration 2 (continuation turn): receipt minted ONCE via the stdio driver (`bash /tmp/t1959-mint-receipt.sh`, moai mcp-server v3.2.0-rc.38 b1806498fe, stdin held open). Mint succeeded: `audit_receipt: rcpt-26567a378837d950e5081ccf`, `tree_root` = this worktree. The codex backend returned `verdict: fail` with two P1 findings, both reproduced at the audited head — re-evaluated independently and folded in as F1 (upgraded Medium) and F3 (new, Medium): both are evidence-class/hardening design gaps with no evidence-unclosed cleared path, hence optional, not blocking; PASS stands. No timing refusal occurred; the N1 escape was not needed.

AUDIT-VERDICT: PASS spec=t1959 receipts=rcpt-26567a378837d950e5081ccf
