# t1921 — lane completion verdict

- card: t1921 (Class B, serial) · run tnbn7e · lane-7
- tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1921 · branch WT-spec-less-lane · **head 75e37d0f0b** (full 75e37d0f0bae19622b2214cfe5b54b62edc64a0b, parent c4784db4c2) · clean
- scope of change: 3 files — internal/cli/factory_card.go (+24/−1), internal/cli/factory_card_spec_propagation_test.go (new, +117), internal/homestate/card_picked.go (+6); no `.github/workflows` changes (verified: `git diff --name-only c4784db4c2...HEAD -- .github/workflows` → empty)

## Claim

The three lease-path record-creation sites now propagate a well-formed queue item's spec_id into the factory card row they create; malformed queue specs and existing rows keep today's behavior. Defect ② is landing-recorded, not repaired, per the leader dispatch.

## Evidence (command + verbatim output)

- RED (delegate, on @c4784db4c2 before any fix): `--- FAIL: TestFactoryNextLeasePropagatesQueueSpecID/b2-creation` / `/c-creation` / `/nominated-creation`, each `row SpecID = "", want "SPEC-X-001"`; guard cases passed pre-fix as designed.
- Lane re-observation (this run, env-scrubbed `unset GOFLAGS &&` compound): `go build ./...` OK · `go vet ./internal/cli/... ./internal/homestate/...` OK · `go test ./internal/cli/ -run 'TestFactoryNextLeasePropagatesQueueSpecID' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 16.507s` · `gofmt -l` (3 files) → empty · `golangci-lint run internal/cli/... internal/homestate/...` → `0 issues.` exit=0.
- Delegate regression set (exit 0 each): `go test ./internal/cli/ -run 'FactoryNext' -count=1 -timeout 45m` → ok 659.761s; `go test ./internal/homestate/ -count=1` → ok 180.847s.
- card-review (codex_review scope=card, base c4784db4c2): verdict pass, findings none → `.moai/reports/t1921/card-review.md`.

## Baseline-attribution

Every figure above was measured in this run against this tree: delegate runs on the same worktree at head 75e37d0f0b; lane re-observations immediately after the commit, HEAD re-read immediately before this verdict (`git rev-parse` → 75e37d0f0b). Independent audits: codex_audit verdict pass, receipt `rcpt-44b010c7dc0c64fdf1145ee2`, review_base main (merge base c4784db4c2), zero findings — bound to the diff whose head is 75e37d0f0b. claude_audit returned inconclusive (CLAUDE_PROVIDER_MISMATCH — GLM-session environment; fail-open) and was recorded, then replaced by the codex independent audit.

## Gaps

- No full-suite run (lane-local discipline; CI runs it on landing).
- The delegate's RED run recorded tee's exit status, not the test binary's; the RED verdict rests on the verbatim `--- FAIL` lines quoted in progress.md §7.
- The first FactoryNext attempt's timeout diagnosis (default 10m under load) rests on the stack trace + an isolation pass; the machine's load at that moment was not re-measured.
- claude_audit inconclusive — recorded above; the independent-audit duty is carried by codex_audit's receipted PASS.

## Residual-risk

- The malformed-spec guard propagates only `^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$`-shaped values; a queue spec that is well-formed but semantically wrong (e.g. a typo'd SPEC id) still propagates — the queue store is not a SPEC registry, unchanged by this fix.
- Existing pre-fix rows with empty spec_id keep it (creation-only propagation); the operator-side set-spec verb (t1840 REQ-CMV-005) remains the backfill path — noted in the card body as a complementary decision, left open by the dispatch.
- Defect ② (factory_card_pr.go:253-255 cwd-dependent specDir read) remains open by dispatch (landing-record only; progress.md §8).

## Lane verdict

**PASS** — run criteria met (regression ok + build/vet/lint 0, RED-first observed), card-review receipt clean, independent audit receipted PASS, workflow-file check clean. Proceeding to Deliver (`moai factory complete`, draft PR, title carries t1921). Merge is the leader's.
