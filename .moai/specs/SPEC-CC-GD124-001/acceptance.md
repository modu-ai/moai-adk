# SPEC-CC-GD124-001 — Acceptance Criteria

Two-cell discipline: every AC pins its RED-now state to tree `615d18c1f` (measured 2026-09-07, verification record V7-V10, re-reproduced in this worktree at plan phase) and states the green-path observable. All commands are single-invocation (no pipes, `&&`, or `;` inside the pinned command) and use absolute worktree paths.

Path abbreviations (all under `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491/`):

- `CW` = `.claude/rules/moai/workflow/context-window-management.md`
- `CWT` = `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
- `CS` = `.claude/rules/moai/workflow/cross-session-messaging.md`
- `CST` = `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md`

Evidence obligation: judge printed `grep -c` counts, NOT exit codes — `grep -c` exits 1 on a zero count; a printed `0` with exit 1 is a PASS on zero-match assertions. Quote the verbatim printed output for every command.

## AC1 — Fable row corrected (cw, both copies)

- RED-now (615d18c1f): `grep -c 'Fable (256K)' CW` = 1; `grep -c '| Fable (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |' CW` = 0. Same on CWT.
- Given the cw pair is edited per M1 When `grep -c 'Fable (256K)' <file>` runs on CW and CWT Then each printed count is `0`. When `grep -c '| Fable (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |' <file>` runs on CW and CWT Then each printed count is `1` — full verbatim row pinned, so a mutant writing the 1M label with stale 256K/90% values fails.

## AC2 — 256K prose threshold gone (cw, both copies)

- RED-now (615d18c1f): `grep -c '200K/256K' CW` = 2 (L64 + L80). Same on CWT.
- Given M1 edit 4 is applied When `grep -c '200K/256K' <file>` runs on CW and CWT Then each printed count is `0`.

## AC3 — Sonnet 5 row separated from real-200K rows (cw, both copies)

- RED-now (615d18c1f): `grep -c '| Sonnet 5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |' CW` = 0; `grep -c '| Sonnet 4.x / earlier standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |' CW` = 0.
- Given M1 edits 2-3 are applied When `grep -c '| Sonnet 5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |' <file>` runs on CW and CWT Then each printed count is `1`. When `grep -c '| Sonnet 4.x / earlier standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |' <file>` runs on CW and CWT Then each printed count is `1` — both rows value-pinned, so a mutant moving the Sonnet 4.x row to 1M/50% fails (protecting REQ-003's error direction). Row separation is what keeps Haiku 4.5 and Sonnet 4.x-era models on 200K/90% rows.

## AC4 — native-Windows denial replaced (csm, both copies)

- RED-now (615d18c1f): `grep -c 'does not provide cross-session messaging on native Windows' CS` = 1.
- Given the M2 OS-bullet edit is applied When `grep -c 'does not provide cross-session messaging on native Windows' <file>` runs on CS and CST Then each printed count is `0`. When `grep -c 'v2.1.234' <file>` runs on CS and CST Then each printed count is at least `1`. When `grep -c 'named pipe' <file>` and `grep -c 'not documented' <file>` run on CS and CST Then each printed count is at least `1` — the mechanism parenthetical and the cross-machine gap statement are pinned, so a mutant dropping either fails.

## AC5 — Providers bullet two-axis rewrite (csm, both copies)

- RED-now (615d18c1f): `grep -c 'unavailable on Amazon Bedrock' CS` = 1; `grep -c 'every provider' CS` = 0.
- Given the M2 Providers-bullet edit is applied When `grep -c 'unavailable on Amazon Bedrock' <file>` runs on CS and CST Then each printed count is `0`. When `grep -c 'every provider' <file>` and `grep -c 'v2.1.248' <file>` run on CS and CST Then each printed count is at least `1`. When `grep -c 'Amazon Bedrock' <file>`, `grep -c 'Claude Platform on AWS' <file>`, `grep -c 'Agent Platform on Google Cloud' <file>`, and `grep -c 'Microsoft Foundry' <file>` run on CS and CST Then each printed count is at least `1` (the four names retained in the cross-machine sentence).

## AC6 — Flags bullet class-level fact (csm, both copies)

- RED-now (615d18c1f): `grep -c 'turning messaging off silently' CS` = 1; `grep -c 'feature-flag fetching off' CS` = 0.
- Given the M2 Flags-bullet edit is applied When `grep -c 'turning messaging off silently' <file>` runs on CS and CST Then each printed count is `0`. When `grep -c 'feature-flag fetching off' <file>` and `grep -c '/list-agents' <file>` run on CS and CST Then each printed count is at least `1` (class-level fact present; diagnostic sentence retained).

## AC7 — Mirror parity (control — MUST NOT flip)

- Pre-state (green today, measured 2026-09-07 in this tree): `diff -q CW CWT` exit 0, no output; `diff -u CST CS` shows exactly 1 hunk whose only changed line is the `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` blockquote (present in CS, absent from CST).
- Given both M1 and M2 are applied When `diff -q <CW> <CWT>` runs Then exit code is 0 with no output. When `diff -u <CST> <CS>` runs Then the output shows exactly one hunk, and the hunk's only change is the Origin-line blockquote block (the blockquote plus its separating blank line) — no new divergence hunks (REQ-010). A mutant that edits only one side of either pair flips this criterion red.

## AC8 — Scope control + five-bullet structure

- Given only M1/M2 edits are made When `grep -c 'tengu_harbor_kite' CS` and `grep -c 'v2.1.236' CS` run Then each printed count is at least `1` (untouched strings intact). When `grep -c 'Sonnet/Haiku=200K' CW` runs Then the printed count is at least `1` (cw L34 GLM parenthetical untouched). When `git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491 status --porcelain` runs Then there are no entries outside the 4 target files, the SPEC directory `.moai/specs/SPEC-CC-GD124-001/`, and `.moai/reports/t491/` (pre-existing untracked evidence files attributed; the observed set may legitimately shrink as run phase commits — the assertion is the absence of extras, not an exact set).
- Five-bullet structure (covers REQ-009): When `grep -c 'Five constraints' <file>` runs on CS and CST Then each printed count is `1`. When `grep -c '- **Operating system**' <file>`, `grep -c '- **Providers**' <file>`, `grep -c '- **Versions**' <file>`, `grep -c '- **Flag evaluation**' <file>`, and `grep -c '- **The shared flag slot**' <file>` run on CS and CST Then each printed count is at least `1` — the five bullets bounded without arithmetic; a bullet merge/split or an edit to the intro line flips this red.

## AC9 — Template embed sanity

- Given the template copies are edited When `make -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491 build` runs Then the exit code is 0. When `go test ./internal/template/...` runs from the worktree root Then the exit code is 0.

## Quality gates

- LSP gates: N/A — prose-only change to `.md` rule files (no Go symbols); recorded in plan.md §E.
- Harness level: minimal — the AC command battery above is the full gate.
- TRUST 5: Tested = AC1-AC9 command battery; Readable/Unified = existing prose register preserved (§D constraints); Secured = no secrets, no behavioral code paths; Trackable = card id t491 in commits + evidence path.

## Definition of Done

1. AC1-AC9 all green with verbatim outputs recorded.
2. AC7 still green (control invariant did not flip).
3. `git status` scope clean per AC8.
4. SPEC status transition `draft → in-progress` owned by manager-develop at first run-phase commit; this file is not edited by run phase without D-NEW-1 re-delegation.
