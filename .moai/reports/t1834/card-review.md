# Card review — card t1834, SPEC-INTEGRATION-STALE-BASE-001

- Tool: `codex_review`, `scope: card`, backend `codex` (advisory, no audit receipt)
- Base (merge base SHA): `28b727065bca2c2afd857a466600429fb87e3dbb` (recomputed at review time — current local main)
- Tree reviewed: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834` at HEAD `1a16da10f6`
- Verdict: **pass**
- Summary (codex): static review of the changed code found no definite defect; per its instructions it ran no tests, builds or reproductions.
- Findings: none → no dispositions owed for this round.
- `truncated: false`, `excluded_untracked: []`

## Prior round (10-18, HEAD ae092df3a, base fb749d372e)

Verdict pass, no findings. Superseded by this review because the diff basis moved (absorb merge 2439da8c90 + the iter12 repair commits 26dad7f9ef..1a16da10f6).

## Dispositions carried from the audit rounds (not review findings)

- D1 [routed, pre-existing, non-blocking] `internal/factory/integration_merge_step.go:950` reads the branch with `rev-parse --abbrev-ref HEAD` — same class as iter12 F2, on a non-card path (the card's plan-pin preserve check showed the file untouched). Live-observed during the F2 end-to-end work; iter13 judged it non-blocking for this HEAD; the required codex gate raised nothing on it. Needs its own repair routing (leader/operator).
- D2 [optional, cosmetic] The F2 fix comment says detached HEAD "stays `""`" like the former error path; measured, the former `--abbrev-ref` read returned the literal `HEAD` on detached HEAD. All callers refuse identically either way; correct the sentence when the routing owner next touches the file. Not adopted in this card (comment-only, out of the audited repair scope).

## Gaps

- Advisory static review only; the independent evidence for this card's sync is `sync-audit-iter13.md` (PASS-WITH-DEBT, receipt `rcpt-5fd438dcfae9026bf8fb0732`).
