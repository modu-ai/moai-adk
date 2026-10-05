# t1290 integration CI: Jev doctrine linkage recheck

## Claim

The merged local-instruction migration moved the live Jev doctrine marker from `CLAUDE.local.md` to `.moai/docs/jev-local-operations.md`. The kickoff linkage test now checks that live document while retaining the original file path for the historical same-commit assertion. A migrated-fixture negative control rejects a missing marker.

## Evidence

- Remote CI job `108808433342` in run `36384957523`, head `c13cee6d58d0e8c7739b6f6fe22ac2960159f17d`: `TestJevAmendmentLinkage/tree` failed with `partial amendment: ... absent in CLAUDE.local.md`.
- `go test -count=1 -run '^TestJevAmendmentLinkage$' ./internal/contract/kickoff` on clean `c13cee6d5` before repair: `FAIL`, `TestJevAmendmentLinkage/tree` reported the same absent path.
- `rg -n 'contract decide|llm\+jev' .moai/docs/jev-local-operations.md` found the moved exception at line 32.
- After repair, `go test -count=1 -run '^TestJevAmendmentLinkage$' ./internal/contract/kickoff`: `ok  github.com/modu-ai/moai-adk/internal/contract/kickoff 1.806s`.
- `go test -count=1 ./internal/contract/kickoff`: `ok  github.com/modu-ai/moai-adk/internal/contract/kickoff 13.781s`.
- `go vet ./internal/contract/kickoff`: exit 0.
- `go test -count=1 -timeout 30m ./internal/template/...` on clean `c13cee6d5`: `ok` for `internal/template`, `agentemit`, and `commandemit`; `scripts` had no test files.

## Baseline-attribution

The failure and repair were measured in the MoAI-created `t1290-template-recheck` worktree, branch `WT-template-clean-verification`, based on local and remote `develop` head `c13cee6d58d0e8c7739b6f6fe22ac2960159f17d`. The remote CI failure is from the same head. The initial package run inside the shared `develop` worktree failed `TestEmbeddedTemplatesNoRuntimeLogs` because two ignored runtime logs dated 2026-09-25 remained under `internal/template/templates/.moai/logs/`; the clean worktree contained only `.gitkeep` and passed.

## Gaps

The patched commit has not yet run in remote CI. The t1290 SPEC remains `draft` with pending lifecycle fields; this test repair alone does not close the card. Other in-progress jobs in run `36384957523` were not treated as passed.

## Residual-risk

The next batched `origin/develop` push must receive its own CI verdict. The t1290 owner must complete post-fix sync audit and SPEC lifecycle evidence before card closure.
