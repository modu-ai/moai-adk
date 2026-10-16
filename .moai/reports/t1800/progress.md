# t1800 progress (lane-40) — handoff, paused by leader

## Where it stopped (tree ec5a4f392, branch WT-t1794-f1-class)
- Lease taken via `moai factory next --card t1800`; entered `.moai/worktrees/t1800`.
- Investigated ec193e477: in `.claude/rules/moai/core/hooks-system.md` and its template twin it deleted exactly one hunk (line 128: the 50K clause lost its "— this 50K cap is on a hook's TOTAL stdout." tail, and the per-`additionalContext` 10,000-char row vanished). The other change in that commit (UserPromptSubmit 5s→15s row) is an addition, not a deletion — t1775 territory. No other deleted lines under .claude/rules, template, or .moai/policies.
- DONE (uncommitted): template `internal/template/templates/.claude/rules/moai/core/hooks-system.md` — restored both rows, without SPEC/card ids or dates (replaced the "decision-index Q4 / lane session" citation with a neutral "measured live: a 23,166-character assembled session-start context was replaced by a file path and preview").
- NOT DONE: nothing committed; local mirror `.claude/rules/moai/core/hooks-system.md` not yet updated.

## Next steps
1. `make build` (or copy the two rows to the local mirror) so `.claude/rules/moai/core/hooks-system.md` matches the template byte-for-byte; `diff` the pair.
2. Re-run AC-RIB-007 grep: `grep -c '10,000' .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md` (expect >=1 each).
3. Sync audit: `go test ./internal/template/...` and `go test ./internal/cli -run 'Mirror|Parity|Doc'` (go-test-heavy slot, last in queue).
4. Commit by explicit pathspec, card id t1800 in message; then the standard landing.
5. Watch t1775: it edits the UserPromptSubmit row of the same file — absorb after it lands.

## t1808 live-acceptance (first check)
- Session model: claude-sonnet-5-5 (from the system prompt). Session effort: not observable from inside the session — no env var or setting visible; unconfirmed.
- First spawn's declared_model: no spawn has happened yet, and `.moai/logs/agent-model-audit.jsonl` does not exist in this tree (ls: No such file). Gap, not a pass.

## lane-39 continuation
- Local mirror copied byte-identical from the template; committed d1481b321 (explicit pathspec).
- Stages recorded with evidence d1481b321: run v7, sync v8, sync-audit v9 (artifact: the template hooks-system.md at that commit).
- Sync audit: PASS-WITH-DEBT, `.moai/reports/t1800/sync-audit.md` (audited_sha d1481b3215796fa62cace8dcebdbb0d8a9959d55), codex gate pass, receipts rcpt-062951309100c8ee65d64e08 / rcpt-20f6ff7c281765f0af5929e8. This file doubles as the card review; no separate card-review.md was written.
- Debt (not caused by this card): `internal/cli` hook-parity test fails because ec193e477 raised the project UserPromptSubmit timeout to 15 while the template settings.json.tmpl still says 5 (needs its own card, t1775 territory); six TestCheckCodexWiring_Mirror* tests fail, inferred to depend on the local Codex home and not run against a clean base (Gap).
- Model: auditor self-reported claude-sonnet-5-5.

## Leader supplement (reported by leader, not observed here)
- effortLevel=medium per the injected --settings file of lane-40 (leader measurement); profile settings say high, injected wins.
- agent-model-audit.jsonl lives in the primary checkout (.moai/logs/), not this tree; after the first spawn, check the lane-40 session_id row there.
