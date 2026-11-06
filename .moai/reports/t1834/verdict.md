# Verdict — card t1834, SPEC-INTEGRATION-STALE-BASE-001

verdict: PASS-WITH-DEBT
audited_sha: 3cd1c782bdd7203c5598ed1ef210599e932d546f

The lane's evidence-bearing completion verdict (the card record carries no SPEC link, so the factory delivery reads this file). Written by lane-6 on 2026-10-21; the independent evidence beneath it is the required codex sync audit, CI (pending at report time, owned by the integration branch), and the mechanical merge-readiness triple.

## Claim

1. Both iter12 P2 blocking findings are closed and test-pinned: F1 `record.Tree == tipTree` inside `factoryAdoptedTip` (internal/cli/factory_card.go:2889), F2 `git symbolic-ref --quiet HEAD` with only `refs/heads/` stripped (internal/cli/factory_card.go:3047).
2. The card's non-.moai diff against origin/main is exactly 19 paths (iter12's 18 + the F2 test's home `internal/cli/factory_card_pr_gitenv_test.go`); the factory-dispatch-gates.md pair is byte-identical.
3. The tree is fresh: 0 first-parent landings behind origin/main after two absorbs (2439da8c90 local main, 3cd1c782bd origin/main #1862).
4. The post-iter13 delta is non-code except the absorbed #1862 landing, which is already-landed main work (rules/worktree hygiene + an mcp catalog test gofmt), none of it on the card's 19 paths.

## Evidence

- Required codex audit (iter13): **PASS-WITH-DEBT** at 1a16da10f6, receipt `rcpt-674fe3203ae611fbda1573e7` (audit_multi, codex required, re-run and corroborated by the stop-hook receipt guard; first receipt rcpt-5fd438dcfae9026bf8fb0732 refused as predating the auditor start). Convergence check `moai verify audit-plan --result-file` → `convergence_check {ok: true, unmet: []}`. Full detail: sync-audit-iter13.md (ten scoped `-p 1 -count=1 -json` runs green with real exit codes, build/vet green, both F1/F2 regression tests green from the JSONL).
- Post-audit delta, enumerated per commit:
  - 0b6086d491 — receipt-corroboration records only (.moai/reports/t1834/*).
  - 1daede4abd — progress.md §G delivery-blocker wait record (docs).
  - d5a905f341 — progress.md §E.4 closed sync_status backfill (docs; the close was already real — spec.md frontmatter `completed` — the line shape just hid it from the readiness checker).
  - 3cd1c782bd — absorb of origin/main #1862 (AGENTS.md, worktree-integration rule + mirror, mcp catalog_test gofmt; merge clean, no card path touched).
- Post-absorb re-verification observed this run: `git rev-list --count HEAD..origin/main` → 0; 19-path diff vs origin/main unchanged; `diff -q` gates pair identical; `go test ./internal/factory/ -run Freshness -count=1` → ok 9.731s; `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestWorkflowRulePathsPinned' -count=1` → ok 0.388s. Earlier scoped confirmation at the repair HEAD (manager-develop §E): the three families (factory Freshness|Remeasure ok 12.654s; cli Freshness|…TestIntegration ok 109.710s; config BaseFreshness|ShippedConfigKeys ok 3.655s), gofmt clean, vet clean.
- merge-readiness triple (moai factory merge ready, this run): conflict-free PASS, tree-identity PASS, re-measure record valid at the then-HEAD tree.

## Baseline-attribution

Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834 at 3cd1c782bd (clean) in this run. The codex audit read 1a16da10f6 — an ancestor of HEAD; the commits between are the four enumerated above (three docs-only on the card, one absorbed main landing), and the card's own 19-path diff was re-verified byte-level after the absorb. moai CLI/MCP calls were served by build 8249bff8c3, a strict ancestor of HEAD (stated per verification-claim-integrity §2.2); the Go toolchain commands ran directly in this tree.

## Gaps

- The repository-wide build/test verdict belongs to CI on the integration branch and is pending at report time.
- The base-freshness gate ships OFF (advisory) in this deployment; its refusal path is test-pinned but not live-exercised on a real window entry.
- One process deviation disclosed by the iter13 auditor: the pre-command load reading before the cli Remeasure scope was 36.38 (above the 32 ceiling, transient compile tail); slot free, mem 84%, the run completed rc=0.
- The re-measure record is keyed to the tree before this verdict file's own commit; the PR-edge triple does not read it (re-measure reader not wired on that path) — recorded, not silently skipped.

## Residual-risk (debts riding with the pass)

- D1 [routed, pre-existing, non-blocking] `internal/factory/integration_merge_step.go:950` reads the branch with `rev-parse --abbrev-ref HEAD` — the same same-name-tag defect class as F2, on a non-card path; live-observed during the F2 end-to-end work; needs its own repair routing.
- D2 [optional, cosmetic] The F2 fix comment's detached-HEAD wording is imprecise (the former `--abbrev-ref` read returned the literal `HEAD`, not `""`); all callers refuse identically either way. Correct when the routing owner next touches the file.

AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-INTEGRATION-STALE-BASE-001 receipts=rcpt-674fe3203ae611fbda1573e7
