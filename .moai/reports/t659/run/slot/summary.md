# t659 internal/cli compile slot — measurement summary

Tree: `WT-amend-apply` at `eec1c334c` (clean before the slot; only this directory added). Go toolchain: the worktree's `go`. Slot granted by the lead (lane-6). Pre-check: `ps` showed 0 running `go test` / `go build` processes. Every command ran serially.

## Steps

| Step | Command | Exit | Result | Evidence |
|---|---|---|---|---|
| 1a | `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_DRY_RUN && go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' -count=1 -v -timeout 600s` | 0 | PASS (subtests `two_occurrences`, `valid`) | `s1-ac015.txt` |
| 1b | `go test ./internal/cli/ -run '^TestResolveRegistryPath_MatchesExecute$' -count=1 -v -timeout 600s` | 0 | PASS | `s1-ac022.txt` |
| 1c | `go test ./internal/cli/ -run '^TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape$' -count=1 -v -timeout 600s` | 0 | PASS | `s1-ac024.txt` |
| 2 | `go test ./internal/cli/ -run '^TestConstitution' -count=1 -v -timeout 600s` | 0 | 13 top-level RUN, 13 PASS, 0 FAIL, 0 SKIP | `s2-constitution-all.txt` |
| 3 | `golangci-lint run ./internal/cli/...` | 0 | `0 issues.` | `s3-lint-cli.txt` |
| 4a | `CLAUDE_PROJECT_DIR=<worktree root> go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' …` (AC-CAA-017, CLI half) | 0 | PASS, same result as 1a | `s4-ac017-cli-sessionenv.txt` |
| 4b | `CLAUDE_PROJECT_DIR=<worktree root> go test ./internal/constitution/ -count=1` | 0 | ok | `s4-ac017-constitution-sessionenv.txt` |
| 4c | `unset … && go test ./internal/constitution/ -count=1 -cover` | 0 | ok, coverage 88.3% | `s4-ac017-constitution-scrubbed.txt` |
| 4d | `shasum -a 256 -c s4-ac017-sha-before.txt` (real registry `f7707b1d…`, real log `f5735051…`) | 0 | both OK after 4a–4c; `internal/constitution/.moai` and `internal/cli/.moai` absent | `s4-ac017-sha-before.txt` |

## CLI mutants (`mutants-cli.json`, runner `../mutate.py`)

Sources were backed up with `cp` before the run and compared with `cmp` after: `internal/cli/constitution.go` and `internal/constitution/registry_path.go` identical.

| Mutant | Change | Selector | Top-level RUN | Verdict | Failure reason |
|---|---|---|---|---|---|
| M-13 | CLI dry-run swallows the pipeline error and prints success | AC-CAA-015 `two_occurrences` | 1 | KILLED | `want an error on a two-occurrence fixture, got nil` |
| M-20 (i), CLI cell | registry-path containment check removed from the shared `LoadAmendRegistry` | AC-CAA-024 CLI case | 1 | KILLED | error does not name the offending registry; refusal did not come from the CLI's own validation |
| M-20 (ii) | CLI reads the registry with `LoadRegistry` (no check) while `Execute` keeps it | AC-CAA-024 CLI case | 1 | KILLED | same as above |

## Re-measure of the agent-reported mutant count (`remeasure/`)

The three runner files were copied with their `out` paths redirected to `remeasure/` so the earlier evidence is not overwritten.

| Set | Mutants | Killed | Summary |
|---|---|---|---|
| M1 | 9 | 9 | `remeasure/summary-m1.txt` |
| M2–M6 | 29 (M-20 (iii) runs against two selectors: `M-20-iii`, `M-20-iii-spec`) | 29 | `remeasure/summary-m2m6-a.txt`, `remeasure/summary-m2m6-b.txt` |
| M-14 | 1 run with `CLAUDE_PROJECT_DIR` exported | FAIL (killed) | `remeasure/summary-m14.txt` |
| M-14 control | same mutant, scrubbed environment | PASS (expected) — the two runs disagree, which is the kill | `remeasure/summary-m14-control.txt` |

Re-measured: 39 mutant runs killed, 0 survived, matching the agent's figure. With the three CLI mutants above: 42 runs killed, 0 survived, 0 pending. The guard mutants (M-GA, M-GB, M-GB-reg, M-GB-log, M-GB-always) are outside this count (`../guards/`).

After all runs, `internal/constitution/pipeline.go`, `registry_path.go`, and `apply_test.go` compared identical with `cmp` against pre-run copies; `git status --short` showed only this directory as new.

## Acceptance-criterion tally

Every cell that was `pending compile slot` is now measured: AC-CAA-015 (1a), AC-CAA-017 CLI half (4a with the scrubbed 1a; real files unchanged, 4d), AC-CAA-022 CLI (1b), AC-CAA-024 CLI case (1c). With the constitution-side cells recorded in progress §E.2 and the package re-run here (4b, 4c), all 25 acceptance criteria have every cell PASS. The constitution-side cells were not re-run one selector at a time in the slot; the package runs (4b, 4c) contain them.

## CLI baseline observation at `299bae37d` (tests committed, CLI change `38928086f` not yet applied)

The tree of `299bae37d` was exported with `git archive` into a scratch directory and the three CLI tests were run there with `go -C <scratch> test ./internal/cli/ -run '^(TestConstitutionAmend_DryRun_SurfacesValidation|TestResolveRegistryPath_MatchesExecute|TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape)$' -count=1 -v -timeout 600s` under a scrubbed environment (exit 1, `s5-cli-baseline-red-299bae37d.txt`):

| Test | At `299bae37d` | Prediction (progress §E.3) |
|---|---|---|
| AC-CAA-015 `two_occurrences` | FAIL — `want an error on a two-occurrence fixture, got nil` | success line printed — matches |
| AC-CAA-015 `valid` | PASS | — |
| AC-CAA-022 CLI `TestResolveRegistryPath_MatchesExecute` | PASS | not predicted RED: a preservation assertion (the CLI precedence already matched) |
| AC-CAA-024 CLI case | FAIL — `clause mismatch` carried; the refusal did not come from the CLI's own validation | `clause mismatch` — matches |

The commit graph already places the tests (`299bae37d`) before the CLI change (`38928086f`); this run observes the RED on that tree after the fact.
