# t248 run-phase trust-but-verify — orchestrator independent batch (2026-09-01)

Measured tree: `feb272f70` (branch `WT-audit-binary-sha`, base `64bba61aa`). Every row below was
executed in this worktree by the lane orchestrator — no value relayed from the implementer report.

| # | Check | Command (abbrev) | Result |
|---|---|---|---|
| V1 | Branch/commit state | `git log 64bba61aa..HEAD` | 5 commits, expected set exactly, unpushed |
| V2 | Diff scope | `git diff --stat 64bba61aa..HEAD` | 10 files — SPEC 5 + code 5 (`mcp_build_identity.go` +81, `_test.go` +623, `mcp_codex.go`, `mcp_convergence.go`, `mcp_glm.go`); touched packages only |
| V3 | Tests, touched packages | `unset MOAI_KANBAN* && go test -cover -count=1 ./internal/cli/ ./internal/binlag/` | exit 0 — cli 293.8s, binlag 4.1s |
| V4 | Coverage | same run | cli 80.0%, binlag 90.9% — implementer figures reproduced |
| V5 | Lint | `golangci-lint run ./internal/cli/... ./internal/binlag/...` | exit 0, "0 issues." |
| V6 | D-1 exact-set sweep | `grep -rn "merge-base\|is-ancestor" internal/cli --include='*.go' \| grep -v _test.go` | exactly 3 baseline hits (`graph_stamp.go:68`/`:131`, `mcp_review_material.go:95`) |
| V7 | Windows build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| V8 | Forbidden surfaces | diff greps | `build_version` added lines: 4, ALL markdown prose (prohibition/mutant naming) — 0 in Go. `AskUserQuestion` added lines: 2, both progress.md E4 quotes — 0 in Go. `performCodexAudit`/`performGLMAudit`: md references only, Go call-site args unchanged. `PerBackendVerdict`: struct body untouched (1 diff token = rationale comment on the added top-level fields) |
| V9 | Run evidence recorded | `progress.md` §E.2/§E.3 | committed in `feb272f70` |

## Gaps

- internal/cli package coverage pre-change baseline was not measured (implementer's own gap, kept as residual) — the 80.0% figure cannot be attributed as delta.
- Full suite (`go test ./...`) intentionally not run locally — CI's verdict per lane protocol; origin/develop CI judgment follows the lead's integration window.
- The 8 ACs were verified via the full-package run (exit 0) rather than re-running each AC's individual `-run` selector; per-test outputs in progress.md §E.2 are the implementer's measurements, accepted because the package-level rerun subsumes them.

## Residual-risk

- internal/cli 80.0% sits below the 85% package target — pre-existing package level, not introduced here (touched functions 96–100%).
- Worktree gopls diagnostics (DuplicateDecl / IncompatibleAssign / unused `auditBuildIdentity`) are contradicted by `go test` exit 0 and `golangci-lint` 0 issues — judged false per the worktree-LSP lesson (unloaded module mimics compile errors; the exit code decides both directions).
- Implementer-disclosed incident (repaired): first `TestPersistedConvergenceCarriesBuildCommit` draft missed the backend stub and spawned a real codex process until fail-open timeout (~60–77s) during the RED round; repaired with `withBackendCallStub`, post-repair 0.00s.

Logs: `orchestrator-test-run.log`, `orchestrator-lint.log`, `orchestrator-winbuild.log` (this directory) — **local-only**; `*.log` is gitignored (`.gitignore:106`), so this verdict file is the tracked evidence carrier and quotes the decision-relevant results inline.
