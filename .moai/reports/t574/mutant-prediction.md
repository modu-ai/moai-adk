# t574 — mutant prediction (pinned before running)

Mutant M-574: `defaultTempRoots()` returns only `{os.TempDir()}` (the t536 sync-audit M-AUD-2 shape).
Tree: t574 worktree, base merge `f295fe698` (HEAD^2 = develop `a4461479e`).

Prediction:
- Every test named by an AC-THG-00x judging command stays PASS (no AC owns set membership).
- `TestDefaultTempRoots_Membership` FAILS (members `/tmp` and `/var/folders` missing, count 1 != 3).

Falsifier: any AC-named test going RED means an existing AC already pins membership and the card's
premise is wrong.
