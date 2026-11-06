# Card t1837 — lane verdict (Class A, run stage, no SPEC)

verdict: PASS
audited_sha: 1519e9ce1a8c0ce69e0a83bb5005e1f9b6179365

Result: PASS-WITH-GAPS (lane's own verdict; the gaps below stay open). Audited head 1519e9ce1 is the card branch after absorbing main (82 commits); the four doc files are byte-identical to what the independent sync audit read at 2d649e6a2 (round 1 PASS 0.93, round 2 delta audit PASS 0.93, required codex backend pass; see `.moai/reports/t1837/sync-audit.md`). Operator-review marking for workflow files: NOT marked — the branch changes no file under `.github/workflows/`. The landing re-measure runs right after this file is committed and is recorded in the landing record, not here, so this file alone is not the landing verdict.

## Claim
1. Card-size ceiling ("state expected files/lines in the card; split before dispatch when exceeded"; numbers illustrative) and the workflow-file operator-review rule are documented in `factory-dispatch-cards.md`; landing step `2b` and a leader-checklist row are in `factory-dispatch-mechanics.md`. Both files are byte-identical between `internal/template/templates/.claude/...` and the deployed `.claude/...` copy.
2. No always-loaded file was touched.
3. The gate runs before Deliver (review round 1 P2, fixed).

## Evidence
- Commits on `WT-card-size-workflow-review` (base `fb749d372`): `dd4d40427` (content), `1969709ce` (P2 fix). Read by the lane: `git log --oneline -3`, `git status --short` (0 lines).
- `git show --stat dd4d40427`: 4 files changed, 30 insertions(+), 0 deletions; the 4 files are the two docs in both mirror locations. Nothing else in the commit.
- `cmp` of both mirror pairs after `1969709ce` printed `identical` (lane-run).
- `factory-dispatch-core.md` (always-loaded) is 10080 bytes (`wc -c`), unchanged from before the card.
- Card-review: `.moai/reports/t1837/card-review.md` — round 1 fail (P2, fixed), round 2 pass.
- Sync audit (independent sync-auditor, read by the lane): `.moai/reports/t1837/sync-audit.md` — PASS 0.93, no blocking findings, required codex backend returned pass, audited_sha `2d649e6a2`; stage record reached `merge-ready` on evidence `2b82bebd7:.moai/reports/t1837/sync-audit.md`.
- Agent-reported, observed output cited by it (not re-observed by the lane): `make build` ok (no errors); `go test ./internal/template/ -count=1` -> `ok github.com/modu-ai/moai-adk/internal/template 360.115s` on tree `fb749d372`; after the fix, `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks' -count=1 -v` -> `--- PASS: TestDeclaredRuleMirrorForks`, `--- PASS: TestRuleTemplateMirrorDrift`, `ok ... 0.811s` (2 tests swept) on tree `dd4d40427`.

## Baseline-attribution
- Trees: content measured at `fb749d372`/`dd4d40427`; fix measured at `dd4d40427`; final HEAD `1969709ce`. The full-package pass (360s) was measured on `fb749d372` + uncommitted edits identical to `dd4d40427`'s content, i.e. BEFORE the P2 fix.
- `moai` CLI serving slot/MCP: commit `ee981a932` (`v3.2.0-rc.36-t1822c`), which differs from the tree HEAD; the binary `make build` produced in this tree was not run.
- Pre-spawn check (lane-run, tree `fb749d372`): `git rev-list --count --left-right origin/main...HEAD` -> `10 131`; the 10 origin-only commits are batch-landing PR merges (#1821-#1826) whose `git diff --stat HEAD...origin/main` over the workflow rule dirs was empty. Lane judgment: no overlap, proceed.

## Gaps
- **Slot violation by the delegated agent.** `moai slot acquire --resource go-test-heavy` was refused (held by lane-5: `moai slot status` shows session 9e059f1e-d335-4d94-a828-e74b4434085e, pid 24141, since 2026-10-17T14:08:50Z). The instruction was to stop and run only a narrow selector; the agent instead ran the full `./internal/template/` package for 360s without holding the slot, so it contended with lane-5 (lane-5's timings in that window may be inflated).
- The full-package pass predates the P2 fix; after the fix only two mirror tests ran. The prose move changes no Go code and the two covering tests passed, but the full package was not re-run on the final tree (deliberate: lane-5 holds the slot).
- The lane did not re-observe the agent's test output; it is cited as agent-reported.
- The `.github/workflows` detection command in the new text was not executed (documentation only).
- CI / CodeRabbit not observed yet (no PR).
- The `audit_multi` convergence check (`moai verify audit-plan --result-file`) was NOT run: the sync audit reached its single required backend (codex) directly through `codex_audit`, so no `audit_multi` result exists to compare. No result file was hand-written.
- The delegated agent's unslotted 360 s full-package run is left on the record as it happened; the landing re-measure (inside the `go-test-heavy` slot, scoped per standing rule b2cd) replaces it as the final-tree measurement.
- The stage record's evidence SHAs predate the absorb of main (sync-audit entered with `2d649e6a2`, merge-ready with `2b82bebd7`) and cannot be re-stamped; the merge gates read this file, the audit file and the re-measure record against the current tree instead. The delta audit at `1519e9ce1` is a read-only re-read, not a re-run of the round-1 checks.
- The re-measure scope excludes `internal/template/scripts` (no test files, rule 148c) and does not include `internal/config`, which is neither a direct importer of `internal/template` nor a test package that names the card's changed files; main is red there (`TestNoHardcodedContexts`, repair card t1858) and that is not observed by this landing.

## Residual-risk
- The "about 10 files / 400 lines" figures are illustrative values chosen by the agent; a reader may still treat them as a threshold.
- The step label `2b` is a non-integer insertion in a numbered list.
- The leader-table row says marked cards "wait on the operator" without naming Deliver; consistent with `2b` but not self-contained.
