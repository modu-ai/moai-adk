# t1416 — completion report (lane-6)

card: t1416 (class C) — ultracode wording correction, Claude Code 2.1.284
spec: SPEC-CC-ULTRACODE-TOGGLE-001 (Tier M, status `completed`)
branch: `WT-ultracode-toggle-wording` (worktree `.moai/worktrees/t1416`, base `c50da9c2f`)
tip at report time: the commit that adds this file on top of `763a6cf0c`
integration: NOT merged (see Gaps — no factory record, window designation requested from the leader)

## Claim

Everything the card named is corrected, in the rule source and its byte-identical template mirror and in the docs-site in four locales: `/effort ultracode` is described as an independent on/off toggle that leaves the effort level unchanged, off with `/effort ultracode off`, with the `--effort ultracode` launch flag (and the Agent SDK `effortLevel: "ultracode"`) as the one exception that also starts the session at `xhigh`, and persistence stated as: the command form lasts for the current session, the `"ultracode": true` settings key (v2.1.284 or later) starts every session with it on. Slider-toggle persistence (OQ-1) is not claimed.

Files changed against the base (`git diff --stat c50da9c2f HEAD`): 2 rule files, 16 docs-site files (workflows.md, multi-llm/_index.md, advanced/ultracode-workflows.md, foundations/commands.md × ko/en/ja/zh), the SPEC directory, and `.moai/reports/t1416/`.

## Evidence

Independent verdicts (committed, read by the lane, not just reported):
- plan-audit iteration 1 FAIL 0.75 (`plan-audit.md`), iteration 2 FAIL 0.86 (`plan-audit-iter2.md`), iteration 3 PASS-WITH-DEBT 0.89 (`plan-audit-iter3.md`, audited `3d44379b8`). Three is the audit ceiling; it was not extended.
- sync-audit PASS-WITH-DEBT 92/100, no blocking finding (`sync-audit.md`, audited `1f4b052b3`): AC-001..AC-010 re-run and hold, hugo exit 0 with 0 WARN/ERROR lines, rule file == mirror (`cmp` exit 0), 18-file diff read for a coupling worded without the token `xhigh`: none.
- Kickoff decision record: `decision-records.md` (autonomous plan→run transition, `auto-semantics.md` §9.1; PASS-WITH-DEBT read as the pass, audited hash unchanged, carried debt D1-D3 named).

Commands the lane ran itself in this session, output as observed:
- `cmp .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` → empty, `cmp_rc=0`.
- `grep -c -E 'xhigh.*xhigh'` over rule + mirror → `0` and `0`; `grep -c 'xhigh'` on the rule file → `1`; `grep -c -F 'leaves the effort level unchanged'` → `1`; `grep -c 'combines \`xhigh\` reasoning'` → `0`.
- `git diff --stat 3d44379b8 HEAD -- .moai/specs` taken before the run phase → empty (plan artifacts unchanged since the audited commit).
- `git status --short` after the sync commits → empty.

Commits (`git log --format='%h %s' c50da9c2f..HEAD`, before this file): plan `e7dbdeef3`, `d7e97e0f4`, `3d44379b8`; audit verdicts `0e7b6af5b`, `ff7b64f6e`, `2ffcbcf7e`, `763a6cf0c`; decision record `0e7087faa`; run `abf8ba627` (rule + mirror), `be5d56b8c` (docs-site), `90d5bac45` (run evidence, status draft→in-progress); sync `211905e37` (status →completed), `1f4b052b3` (sync_commit_sha backfill = `211905e37`). Every commit message names the card id; the `Authored-By-Agent` trailers read manager-spec, plan-auditor, manager-develop, manager-docs, sync-auditor.

## Baseline-attribution

All lane measurements above were taken in this session against the worktree `.moai/worktrees/t1416`. Divergence at the last pre-spawn check: `git rev-list --count --left-right origin/develop...HEAD` → `0 10` (origin/develop had already absorbed the leader's batch), local `develop` → `0 10`. The judging `moai` binary for lint/audit lags HEAD (strict ancestor), so the clean `moai spec lint` / `spec_audit` results are not cited as corroboration. The run-phase `make build` / `go test ./internal/template/...` results are the run agent's record in `progress.md` §E.2, not re-observed by the lane or the sync auditor; the sync auditor did confirm that a `bin/moai` built from this tree carries the new sentence once and the old one zero times.

## Gaps

- AC-011 (`make build`, `go test ./internal/template/...`) was observed only by the run agent: build exit 0; the first test run hit Go's default 10-minute timeout under machine load (~600) with 0 `--- FAIL` lines and was recorded as a tool failure, the rerun with `-timeout 45m` passed (`ok … 349.908s`). That is one more run than the "once" the lane asked for; disclosed here.
- Not integrated and not recorded in the factory: this card was dispatched directly by the leader and has no factory record, so `moai factory complete` cannot record it (documented lane limitation). Per the lane-integration lesson the lane did not improvise a merge; it asks the leader for the window designation. `moai integration status` read `release-integration window: free` when checked.
- OQ-1 (does the slider toggle persist across sessions) is unobserved; the edited text never mentions the slider.
- ja/zh/ko/en sentences were not read by a native speaker; the sync auditor flagged two readings as questionable (F2, F3 below).
- Cross-backend audit (`audit_multi`) was not requested and not run.

## Residual-risk and leader decisions

- Optional findings from the sync audit (none blocks): F1 the settings-key sentence says the key "requires v2.1.284" while upstream says the key existed before and ran at `xhigh` (the SPEC fixed that wording; follow-up card candidate, rule source and zh row); F2 the ko workflows row (`workflows.md:113`) "켜기와 끄기를 따로 정하는 토글입니다" can misdirect what the toggle is independent of (ja row has the same structure); F3 the ko `ultracode-workflows.md:83` callout is weakly at odds with the settings-key guidance on the same page; F5/D4 SPEC body mismatch (spec.md L50 ledger SHA, HISTORY order) left as is because the SPEC body is manager-spec's; F6 `progress.md` §E.1 still says "Final audit (iteration 3) pending"; F7 `draft→in-progress` landed on the fourth commit and the manager-develop role was played by a `general-purpose` spawn (documented isolation workaround), not a lint violation.
- CHANGELOG: no `[Unreleased]` line was written on purpose (the SPEC change map and AC-010 require `CHANGELOG.md` unchanged). Whether the wording fix deserves a line is the leader's decision.
- Rule note F9: `plan-auditor.md` L691-695 says an exhausted third iteration escalates to the user; the lane took the §9.1 autonomous path with PASS-WITH-DEBT instead. The sync auditor judged that defensible and did not rule on the two sentences' relationship.
- The pins are string-presence checks; they cannot prove a locale's sentence reads well. The known residual (coupling worded without the token `xhigh`) is left to diff review; the sync auditor's mutant run showed 6 of 7 rule mutants and 4 of 7 workflows-row mutants pass every release-blocking pin (wrong version, revived general coupling, flipped claim); the delivered text is not one of them.
