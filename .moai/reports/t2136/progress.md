# t2136 progress — GitVerifier.Lookup inherited GIT_DIR (t2126 follow-up)

Class B, no SPEC. Base 04d8c27a7e, branch WT-internal-ctxmodel-ctxmod. AF-Flow: local only, no push. Codex not approved (Claude audit only).

## Reproduction (RED)
`TestGitVerifierLookupIgnoresInheritedGitDir` (internal/agenticfactory/ctxmodel/gitverifier_env_test.go): two real repos (clone without the commit, other with it); `GIT_DIR=other/.git` inherited; `GitVerifier{Dir: clone}.Lookup(commit, "f.md")`.
Command: `go test -count=1 -run TestGitVerifierLookupIgnoresInheritedGitDir ./internal/agenticfactory/ctxmodel`
Observed (EXPECTED_RED, assertion failure, exit 1): `Lookup found the blob (sha256 f3552474…) in another repository through an inherited GIT_DIR`.
The defect reproduces, so the fix applies.

## Fix
`GitVerifier.Lookup` sets `cmd.Env = gitenv.Env()` (+ import). Same one-line pattern as t2126.

## GREEN
- `go test -count=1 ./internal/agenticfactory/ctxmodel ./internal/gitenv` -> both `ok`.
- `go vet ./internal/agenticfactory/ctxmodel` -> exit 0.

## Gaps
- No separate mutant of the fix line (the RED run is the fix-absent evidence).
- Only `./internal/agenticfactory/ctxmodel` and `./internal/gitenv` re-run; not the whole agenticfactory tree.
- Independent audit iter1 (sync-audit-iter1.md): verdict FAIL @5e02a7a780, code score 0.94, no code finding. Sole cause: the required Codex gate (workflow.audit.gates.codex: required) is unmet — Codex is not approved for this card (leader dispatch), so no codex_audit was called; the receipt guard refused a receipt-less PASS and the auditor wrote FAIL (fail-closed). No verdict.md was created and merge-ready was NOT recorded. Reported to the leader.
- Audit-side checks executed: mutant (fix line reverted in a scratch export) fails the new test, HEAD passes; ctxmodel+gitenv `ok`; go vet exit 0; gofmt clean.
- The auditor's F1 (checkpoint GitRemote.git lacks Env) is already fixed on card t2126 (954a2e42a3), not on this card's base.
