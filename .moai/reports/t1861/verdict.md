# Card t1861 — lane verdict (Class B, no SPEC)

verdict: PASS
audited_sha: cff1f845eda2524f49d9c23f01c79f68f45a3035

Absorb update (lane-22, takeover from lane-37): main 629ba6f7b6 was absorbed by a no-ff merge, HEAD cff1f845e. The head-confirmation audit `.moai/reports/t1861/sync-audit-head-confirmation.md` is PASS 0.94 at that head (codex required pass, receipt `rcpt-ef2ec6e21d48739a3c112e48`): the card delta is still the five Go files, the merged `store.go` carries both this card's regex block and card t1855's clear helpers in disjoint regions, and no test was run by the audit (the landing re-measure replaces that). The paragraph below is the iter2 verdict at 081d2270c, kept as the history it was.

Final: sync-audit round 2 PASS 0.94 (all four dimensions PASS; F1 repair verified
adversarially, no gate loosening found — `.moai/reports/t1861/sync-audit-iter2.md`), receipt
`rcpt-38d0840a53536131174bf38a` (codex / required / pass, plan_source config, no fail-open,
no disagreement). Orchestrator convergence check: `moai verify audit-plan --project-root
<tree> --result-file <tree>/.moai/state/audit-plan-result.json` → `convergence_check.ok:
true`, exit 0 (judge build `ee981a932`, a strict ancestor of HEAD — F2 build lag, noted;
the check reads the recorded digest). Round 1 FAIL 0.78 preserved in `sync-audit.md`
(receipt `rcpt-ec22b4289519e9ee4fb310ca`); its blocking finding F1 was repaired in
0d5c1b40b (RED) + 081d2270c (fix) and re-audited above.

## Claim

1. Cause: `verdictLinePattern`'s spec= span accepts only SPEC-shaped ids
   (`specIDBody = SPEC(?:-[A-Z][A-Z0-9]*)+-[0-9]{3}`), so a card-flow audit's verdict line
   naming the card (`spec=t1854`, `spec=card-t1856` — the two observed spellings) failed to
   parse and SubagentStop recorded `unknown-spec` / `"verdict line missing"` refusals for
   provable PASSes. All three incident refusals (t1854 iter2, t1833 iter4, t1856 iter3) are
   this one parse miss; full evidence in `.moai/reports/t1861/cause-analysis.md`.
2. Fix (commit `061b8cf1f`): the verdict line's spec= alternation now also accepts
   `(?:card-)?t[0-9]+`. `specIDBody` and the free-text `specIDPattern` are untouched, receipt
   corroboration is unchanged, and near-miss shapes (`t`, `1854`, `card-`) stay refused.
3. RED (commit `6c11f1728`, precedes the fix in the commit graph): both observed spellings
   pinned at the parser table and at the SubagentStop guard, plus a no-verdict-line stop that
   must still block and three near-miss shapes that must stay refused.
4. No test was weakened, skipped or deleted; the fix widens one regex alternation only.

## Evidence (lane-observed in this run, tree `061b8cf1f`)

- RED observed at tests-only state (commit `6c11f1728` checked out content; fix not yet applied):
  - `go test ./internal/auditreceipt/ -run TestParseVerdictLine -count=1` → exit 1;
    `--- FAIL TestParseVerdictLine/card-spec` and `/card-spec-prefixed` (`ok = false, want true`).
  - `go test ./internal/hook/ -run TestSubagentStop_CardSpec -count=1` → exit 1;
    `decision = "block"`, reason `… reported no parseable verdict line … — verdict line missing`
    for both `t1861` and `card-t1861` — the incident reproduced verbatim.
  - The no-loosening case passed pre-fix (`CardSpecNoVerdictLineStillRefused`: block +
    unknown-spec record with cause `verdict line missing`).
- GREEN after the fix, same commands → exit 0 both (`ok … internal/auditreceipt 0.282s`,
  `ok … internal/hook 1.709s`).
- Scoped regression (changed-function consumers):
  - `go test ./internal/auditreceipt/ -count=1` → `ok 2.004s` (full package).
  - `go test ./internal/hook/ -run 'TestSubagentStop|TestServed|TestPreTool|WSR|Receipt' -count=1`
    → `ok 30.167s`.
  - `go test ./internal/closure/ ./internal/homestate/ -count=1` → `ok 18.352s` / `ok 249.716s`.
  - `go vet ./internal/auditreceipt/ ./internal/hook/` → exit 0;
    `golangci-lint run internal/auditreceipt/... internal/hook/...` → `0 issues.`;
    `gofmt -l internal/` → empty.

## Baseline-attribution

All measurements above were executed in this run on this tree
(`061b8cf1f`, branch `WT-10-18-9891-2`, worktree `.moai/worktrees/t1861`). The auditor
transcript quotes in `cause-analysis.md` were read from the incident sessions' subagent
transcripts under `~/.moai/claude-profiles/moai-adk/projects/…` in this run.

## Gaps

- The full `internal/hook` package and the CI matrix were not run locally (lane discipline:
  heavy suites go to the leader's slot rotation / batch CI), and the unchanged packages of the
  repo are unobserved.
- The verdict-line grammar widening is pinned only for the observed card-id spellings; other
  unobserved shapes (markdown-fenced lines, other prefixes) remain refused by design.
- Sync audit round 1 (see sync-audit.md, audited 061b8cf1f): FAIL 0.78 — codex(required) FAIL
  (P1 = F1) plus the must-pass firewall on Functionality/Security. The primary fix itself was
  verified by the auditor (TestParseVerdictLine 15/15, both guard tests, RED probe on the old
  pattern, gofmt/vet clean). Repaired and cleared in round 2 (PASS 0.94, receipt
  rcpt-38d0840a53536131174bf38a, convergence ok).
- Round-2 optional findings left open: F2 build lag (installed binary ee981a932 predates the
  fix — card-shaped verdict lines stay refused by the RUNNING guard until a rebuild), N1 no
  direct unit table for IsCardSpecID (covered indirectly; single production caller), N2 codex
  bare pass carries no independent finding list.

## Sync audit round 1 — F1 repair (commits 0d5c1b40b RED, 081d2270c fix)

F1 [high][blocking]: the widening changes served-refusal attribution from unknown-spec to the
card id, and `checkServedModelSpawn`'s skip condition matched only SPEC ids, so a
card-attributed served refusal stopped denying SPEC-named phase-entry spawns (the conservative
reach unknown-spec used to provide). Repair (direction (b) of the audit's options): a
card-attributed refusal now denies like unknown-spec — `auditreceipt.IsCardSpecID` + one added
skip-condition clause; the t1323 SPEC-attribution axis is unchanged (pinned by subtest).
- RED observed (0d5c1b40b, pre-fix): `TestServedModel_CardSpecRefusalStillDeniesPhaseEntry`
  FAIL for both `t1861` and `card-t1861` (`denied=false reason=""`); promptless and t1323
  subtests PASS.
- GREEN post-fix (081d2270c): full `TestServedModel` suite `ok 5.137s`; auditreceipt full
  `ok 2.153s`; hook targeted `ok 17.907s`; vet exit 0; `golangci-lint` `0 issues.` (first
  attempt refused: "parallel golangci-lint is running" — retried after the lock cleared);
  `gofmt -l` empty.
- F2 (build lag) and F3 (probe heredoc refusal) recorded in sync-audit.md; no code change.

## Residual-risk

- t1855 (lane-4, merge-ready) touches the same file `internal/auditreceipt/store.go` at a
  disjoint region (clear-by-cause helpers ~line 1156+ vs this card's pattern block ~121-150);
  landing both requires the later landing to absorb main and re-run the full auditreceipt
  package (leader instruction).
- The installed moai binary predates this fix, so until the next build any card-flow auditor
  must cite a SPEC-shaped id (`SPEC-NOSPEC-…`, t1799 precedent) or its verdict line will be
  refused by the running guard — the defect this card fixes.
- A card-attributed served refusal now denies every phase-entry spawn of its tree until
  cleared (conservative, pre-widening parity); a per-card discrimination would need a
  card-token extractor on prompts, which free-text card mentions make noisy — deliberately
  not attempted here.
