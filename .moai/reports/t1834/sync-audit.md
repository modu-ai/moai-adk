# Sync audit — card t1834, SPEC-INTEGRATION-STALE-BASE-001

Export relayed by the lane (lane-28): the `audit_multi` MCP tool returns the verdict as a result and writes no file, so the lane recorded the tool results below verbatim-in-substance. This artifact is a relay, not an auditor-authored file.

## Verdict

verdict: PASS
audited_sha: ae092df3adf82decccaa2e47a35a25ad9f963426


**PASS** — round 4, tree HEAD `ae092df3a`, backend codex (gate `required`), receipt `rcpt-8fdd95a2ca747abf709b8914`, `overall_verdict: pass`, `findings: []`, `participant_count: 1`. No numeric score is returned by this tool.

## Iteration history

| Round | Tree HEAD | Verdict | Receipt | Findings |
|---|---|---|---|---|
| 1 | `d8fef1617` | FAIL | `rcpt-c9bb6784d6ebb463b124e1d4` | P1 refusal message never printed (exit 16, empty streams) — AC-ISB-001/011; P2 record `Tree` not matched to current tree — AC-ISB-005; P2 config load failure read as gate-off — AC-ISB-009 |
| 2 | `85ef27e18` | FAIL | `rcpt-60312ee19c3179eea8af7e44` | P2 raw YAML parser error echoed config content |
| 3 | `6a6b9032f` | FAIL | `rcpt-56f80c4d7a80842575127f29` | P2 record `Command` leaked via `%v`; P2 raw git error text forwarded |
| 4 | `ae092df3a` | PASS | `rcpt-8fdd95a2ca747abf709b8914` | none |

Repairs between rounds (all by `manager-develop`, RED first): `187e58c1b`/`da0e3e840`/`619402f00`/`85ef27e18` (round 1), `d2f33e61d`/`8025a3253`/`6a6b9032f` (round 2), `b3e5df8ee`/`81054c29c`/`ae092df3a` (round 3: every cause in the gate output now comes from a fixed phrase set, with a table-driven non-disclosure test and two mutant probes). Detail: `progress.md` §E.2.7.

## Claim / Evidence / Baseline-attribution

- Claim: the sync-phase tree has no open required-gate finding. Evidence: the round-4 `audit_multi` result above, observed in this run against tree HEAD `ae092df3a`. Judging build: installed `moai` commit `ee981a932`, reported by the tool as a strict ancestor of the tree HEAD (the audit backend is codex, reading the tree; the installed binary only brokered the call).

## Gaps

1. The auditor returned `Verdict: pass` with no findings text; what codex re-ran or re-read in round 4 is not in the result. The lane did not independently re-run the findings' reproductions at `ae092df3a`.
2. Repository-wide `go build ./...`, the Windows cross-build, `make build` and whole-package `internal/factory` / `internal/cli` suites are not run (the `go-test-heavy` slot is the leader's rotation; `internal/cli` is `-run` only per rule 35e9). Package coverage is unmeasured.
3. The neighbouring `settings drift: UNDETERMINED` line in `moai integration acquire` forwards git stderr verbatim; it belongs to another feature, was named out of scope in the round-4 audit focus, and is carried to the leader.
4. The card branch is 3 first-parent landings behind local main (`f4bcd4f2c`) and has not absorbed them; the PASS is attributed to `ae092df3a`, not to a post-absorb tree.

## Residual risk

- A post-absorb tree is a different tree; the landing sequence's remeasure applies to it.
- The gate's K-slack behaviour with a live queue was not exercised (progress.md E.2.6 Gap 5).
