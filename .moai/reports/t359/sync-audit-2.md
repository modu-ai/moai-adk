# Sync Audit (re-audit #2, delta-scoped) — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

> **PROVISIONAL — written early by budget discipline; revised in place as measurements land.**
> Superseded sections are replaced, never appended around.

- Auditor: sync-auditor (fresh agent; the first verdict `sync-audit.md` @ `d739cd051` is on disk and is treated as a record to check, not as a premise)
- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- Branch: `WT-landing-evidence`
- HEAD re-read by THIS auditor at audit start: `5350d0e63`, working tree clean
- Tier L · PASS threshold 0.85 · profile `default` · flat weighted-percentage mode

## Provisional Verdict

**FAIL (provisional) — pending the full-gate and coverage measurements still running.**

Delta confirmed so far: F1 closed by measurement (RED reproduced by this auditor at the pre-fix
commit, GREEN at HEAD, plus an end-to-end binary probe). F4 guarded at all three sites. One NEW
finding: the F4 test's positional assertion is a tautology.

(Full report follows below once the gates land.)
