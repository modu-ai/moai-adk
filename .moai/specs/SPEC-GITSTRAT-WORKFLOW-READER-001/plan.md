# plan.md — SPEC-GITSTRAT-WORKFLOW-READER-001

Tier M · cycle_type: tdd · Card t656 (re-scoped) · Branch WT-git-flow-reader

## §A Context

`git_strategy.<mode>.workflow` is read in production by `LoadGitFlowIntegrationConfig` (`internal/config/loader_integration_branch.go`, cards t449/t637) but interpreted as a binary git-flow discriminator. The re-scoped goal (operator, card t656): validate against 4 allowed values, interpret an integration target per flow, preserve git-flow behavior via characterization-first TDD. Design decisions D1-D4 are owned and resolved in `spec.md` §C — nothing in this plan re-opens them.

## §B Known issues / gaps this plan closes

- G1: Invalid workflow values (typos, `trunk-based`) are indistinguishable from deliberate non-git-flow choices — no diagnostic surface exists.
- G2: Only git-flow has a target-branch interpretation; github-flow (the shipped default!) resolves its target through the caller fallback even though `main` is well-defined.
- G3: `shipped_key_inventory.yaml` evidence for the three workflow keys names a generic "reader" — after extension it should name the validated surface.

## §C Pre-flight

- [ ] Confirm base: local develop head `b1bd81b23` present in history (`git merge-base` sanity).
- [ ] Run existing suite scoped to the change: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR && go test ./internal/config/ -run 'TestLoadGitFlow|TestIsGitFlow|TestGitFlowIntegration' -count=1` — record baseline PASS before M1.
- [ ] Confirm template default: `grep -c 'workflow: github-flow' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` → 3 (measured 2026-09-14).
- [ ] Coverage baseline: `go test -cover ./internal/config/` snapshot recorded in progress.md §E.1.

## §D Constraints

- Characterization FIRST: M1 commits before M2 (ordering is witnessed by the commit graph, per verification-claim-integrity §2.3).
- No production code change in M1.
- Fail-open reader contract preserved (no error returns; structured disposition).
- Scope discipline: touch only `internal/config/{types.go?,loader_integration_branch.go,+_test}`, `internal/cli/{integration.go warning, doctor check, +_test}`, `internal/config/testdata/shipped_key_inventory.yaml`. No wizard changes (D3 deferred). No template value change.
- Verification load lane-local: affected packages only (`go test ./internal/config/... ./internal/cli/ -run <scoped>`), never the full suite locally (CLAUDE.local.md §4).

## §E Self-verification plan (run-phase §E.2 evidence obligations)

1. AC PASS/FAIL matrix from `acceptance.md` with verbatim `go test` output per AC.
2. Coverage: `go test -cover ./internal/config/ ./internal/cli/` ≥ 85% observed output.
3. Characterization ordering: `git log --oneline` shows the M1 (characterization) commit preceding the M2 (extension) commit.
4. Behavior-preservation proof: characterization tests from M1 pass UNMODIFIED after M2/M3 land (run with `-count=1`).
5. Lint: `golangci-lint run internal/config/... internal/cli/...` clean.
6. Doctor item: `go run ./cmd/moai doctor --check <new-item-name>` output captured for the invalid-value fixture.

## §F Milestones (priority-ordered, no time estimates)

### M1 — Characterization suite (Priority High, RED-neutral)
- Write table-driven characterization tests in `internal/config/loader_integration_branch_test.go` (extend the existing t637 file) pinning:
  - `LoadGitFlowIntegrationConfig`: manual+git-flow+develop_branch set → full struct; each failure path (missing file, unparseable YAML, mode=personal/team, workflow≠git-flow, no active profile, develop_branch empty/whitespace) → expected zero/false halves.
  - `IsGitFlow()` truth table over all four halves combinations.
  - `LoadGitFlowDevelopBranch` delegation + trimming behavior.
- All tests pass against the UNMODIFIED production tree. This is the PRESERVE baseline.
- Commit: `test(config): characterize git-strategy workflow reader behavior (t656)` — its own commit, before M2.

### M2 — Multi-flow validation + interpretation table (Priority High, RED→GREEN)
- `internal/config`: add the allowed-set constant (4 values), the three-way disposition type (`WorkflowDisposition`: git-flow / valid-non-git-flow / invalid), extend `GitFlowIntegrationConfig` with `Workflow string` + disposition fields; add the per-flow integration-target resolver implementing the D2 table (github-flow→main fixed; git-flow→develop_branch gated as today; gitlab-flow→environment; release-flow→release_branch_prefix; empty key → empty string).
- Unit tests RED first (invalid value, each allowed value, empty target keys), then GREEN.
- Existing characterization tests (M1) MUST pass unmodified.
- Commit: `feat(config): validate git_strategy workflow against 4 flows + per-flow target table (t656)`

### M3 — Consumer diagnosability (Priority Medium)
- `internal/cli`: invalid-disposition warning in `moai integration acquire` (alongside the t637 git-flow fallback warning, same stderr fail-open pattern); extend/parallel in the auto-merge gate path only if it has a warning surface (gate skip stays silent as today — the diagnostic lives in doctor).
- New `moai doctor` check item (precedent `doctor_worktree_base.go`): reports invalid workflow value + the 4 allowed entries + repair next-step (edit git-strategy.yaml). OK for git-flow, OK-silent for valid-non-git-flow, WARN for invalid.
- Tests: doctor check four-state table; acquire warning emission on invalid fixture (stderr capture).
- Commit: `feat(cli): diagnose invalid git_strategy workflow values in acquire + doctor (t656)`

### M4 — Inventory evidence + gates (Priority Medium)
- Update `internal/config/testdata/shipped_key_inventory.yaml`: the three `git_strategy.<mode>.workflow` entries' evidence names the validated reader (path/function), keeping `class: W`.
- `make build` (Template-First loop is a no-op here — no template edit — run for embed-check hygiene only if any template-adjacent file moved).
- Gates: scoped tests + coverage ≥ 85% + `golangci-lint run` on touched packages; record verbatim outputs in §E.2.
- Commit: `docs(config): point shipped_key_inventory workflow keys at the validated reader (t656)`

## §G Anti-patterns (explicitly forbidden in run phase)

- Making the reader return an error / refuse load on invalid values (breaks t449 fail-open contract).
- Reordering: writing M2 production code before the M1 characterization commit exists.
- Changing template defaults or the wizard interview (D3 is deferred, not silently dropped).
- Sweeping the full `go test ./...` locally (CLAUDE.local.md §4).

## §H Cross-references

- spec.md §C (D1-D4 design decisions) · acceptance.md (AC matrix) · research.md (provenance) · progress.md (§E.1 signal)
