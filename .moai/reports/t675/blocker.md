# t675 Blocker — inference gateway 5xx

## Claim

Card t675 is paused before plan repair and iter2 audit. The repair delegations cannot proceed because the inference gateway returned `502 Bad Gateway`. No further retry is authorized while the card is paused.

## Evidence

### Existing gateway record

Source: `.moai/reports/t675/gateway-502-20260913.md`

```text
2026-09-13 t675 manager-spec HTTP 502: 21:14:01+0900, 21:28:39+0900, 21:45:17+0900; family=da7062b1-7b2b-4f3c-8f08-049654a8e378; baseline=WT-doctor-red@74d872aafbd90235e67163a5bc233f7c8a934491; files_created=0.
```

### Latest repair delegation failure

Task: `a2e146a86a28af5a2`

Observed error:

```text
API Error: 502 Bad Gateway. This is a server-side issue, usually temporary — try again in a moment. If it persists, check your inference gateway (127.0.0.1:49298).
error type: server_error
HTTP: 502
model sent to API: gpt-5.6-sol
```

The earlier fresh repair delegation `a3b5df6f0e52c1d82` ended with HTTP 400 conversation-history/reasoning-chain mismatch. The initial plan delegation `a5d71cfd88143aaa9` ended with the same HTTP 400 class after writing the initial artifacts. These are recorded here for resume diagnosis; the three measured 502 timestamps and family id remain the canonical gateway comparison points above.
- Fourth 502: `2026-09-14 00:30:34+0900` (`2026-09-13T15:30:34.424Z`), session family `332fd0c3-895a-4e42-948a-5cc8edbe02d6`, agent `a4dc79d08ad972d06`, attempt `11/11`.

### Local debug-log search

Commands run:

```bash
find /Users/goos/.claude -maxdepth 4 -type f \( -name '*.log' -o -name '*.jsonl' \) -mmin -30 -print 2>/dev/null | sort | tail -20
find /tmp -maxdepth 3 -type f \( -iname '*debug*' -o -iname '*gateway*' -o -iname '*.log' \) -mmin -30 -print 2>/dev/null | sort | tail -30
```

Both commands produced no output. This is a bounded path search, not evidence that no gateway logs exist elsewhere.

## Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675`
- Branch: `WT-doctor-red`
- HEAD: `74d872aafbd90235e67163a5bc233f7c8a934491`
- Baseline: `WT-doctor-red@74d872aafbd90235e67163a5bc233f7c8a934491`
- Gateway family: `da7062b1-7b2b-4f3c-8f08-049654a8e378`

## Current partial state

- `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/spec.md` has partial `version: "0.2.0"` metadata, but the D1–D5 audit findings remain unresolved.
- `.moai/specs/SPEC-DOCTOR-TEST-CWD-ISOLATION-001/plan.md` remains the pre-repair artifact.
- `.moai/reports/t675/plan-audit.md` preserves the iter1 `FAIL / 0.67` verdict and D1–D5 findings.
- The artifacts were untracked at the time of this record; preservation is handled by the card evidence commit.

## Gaps

- No gateway process log was found in the two bounded search locations above.
- The API error does not expose a request trace beyond the gateway address, family id, task id, model, and HTTP status.
- No plan repair or iter2 audit was run after the final 502.

## Requested action

Lead to keep t675 paused, preserve this blocker and the partial SPEC/audit artifacts in a local card commit, and resume only after lane-5 card t707 (receipt-chain / upstream 5xx repair) lands on `develop` and the inference gateway is restarted. No push, worktree disposal, plan repair retry, or iter2 audit is performed before the lead's resume signal.
