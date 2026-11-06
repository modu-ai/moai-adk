# t2144 progress — integration-name tag shadow (follow-up of t2121 audit debt G2)

Class B, no SPEC, cycle_type=tdd. Branch WT-merge-readiness-resolves, base f9d8866951.

## What changed

One helper, `integrationRevSHA(git, name)` in `internal/factorylane/merge.go`, resolves the integration name to a commit SHA:

1. Name not starting with `refs/`: `git show-ref --verify --hash refs/heads/<name>` (exact ref path, which a tag cannot shadow).
2. Otherwise the name as given, verified with `git rev-parse --verify --quiet <name>^{commit}` (keeps `origin/main` and `refs/...` working).
3. Neither resolves: fail-closed refusal naming the name and both candidate forms (`refs/heads/<name>` and `<name>`).

The SHA, not a spelling, is then the revision operand.

## Every site changed

| File | Function | Change |
|---|---|---|
| `internal/factorylane/merge.go` | new `integrationRevSHA` | the helper |
| `internal/factorylane/merge.go` | `checkConflictFree` | `merge-tree` now takes the resolved SHA; refusal on unresolvable |
| `internal/factorylane/merge.go` | `checkTreeIdentityPreMerge` | same |
| `internal/factorylane/merge_specless.go` | `absorbOnlyChange` | its inline `rev-parse develop^{commit}` replaced by the helper; the resolved SHA feeds the ancestor checks and `cardPatch` |

Sites examined and left unchanged:
- `checkSyncAuditVerdictFile` only forwards `develop` to `absorbOnlyChange` and prints it in messages (no revision use).
- `checkTreeIdentityOnMergeCommit` takes a merge commit, not the integration name.
- Other files in `internal/factorylane` have no use of the integration name (grep for develop/integration* found none).
- `internal/cli/factory_merge.go`: `developRef` is passed as `Develop` (covered) and also as `Branch:` and to `worktreeForBranch(developRef)` at lines 230/232 for the integration-window merge request. Those feed the `integration` package, not the merge-readiness checks, and were not audited here (see Gaps).

Message text for the existing checks is unchanged (display still shows the name as given).

## TDD evidence

- RED: `.moai/reports/t2144/red-output.txt` (commit 3b4cea05ca), tests in `internal/factorylane/merge_tag_shadow_t2144_test.go`. 5 of the new subtests failed on the old code: tree-identity and conflict-free measured the tag's old commit, a valid verdict was refused, the forged-by-tag absorb (probe N5) was accepted, the nonexistent name did not name `refs/heads/<name>`.
- GREEN: commit 41c61dc10b.

Cases covered: (a) tag `main` on an old commit with the branch ahead for conflict-free, tree-identity, a verdict valid only against the branch, and the forged absorb of an unaudited commit; (b) no tag, unchanged; (c) `origin/main`; (d) `refs/heads/main`; (e) nonexistent name refused naming both candidates.

## Verification (each run on its own, output in files under this directory)

| Command | Result tail |
|---|---|
| `go test ./internal/factorylane/ -run TestT2144 -count=1` | `ok  github.com/modu-ai/moai-adk/internal/factorylane 28.671s` (green-t2144.txt) |
| `go test ./internal/factorylane/ -count=1` | `ok  ... 200.963s`, exit=0 (v1-factorylane.txt) |
| `go test ./internal/cli/ -run MergeReady -count=1` | `ok  ... 98.277s`, exit=0 (v2-mergeready.txt) |
| `go test ./internal/cli/ -run SpecLess -count=1` | `ok  ... 67.891s`, exit=0 (v3-specless.txt) |
| `go test ./internal/cli/ -run FactoryMerge -count=1` | `ok  ... 34.241s`, exit=0 (v4-factorymerge.txt) |
| `go vet ./internal/factorylane/` | no output, exit=0 (v5-vet.txt) |
| `gofmt -l internal/factorylane` | empty, exit=0 (v6-gofmt.txt) |

No timeouts occurred. The repository-wide test verdict is PENDING with CI on the integration branch.

## Gaps

- Not run: the full `internal/cli` suite and other packages; only the three named patterns plus the whole `internal/factorylane` package.
- The `Branch:`/`worktreeForBranch` uses of `developRef` in `internal/cli/factory_merge.go` (lines 230, 232) and the `internal/integration` package were not examined for the same tag shadow.
- Tool provenance: tests ran through `go test` on this tree; git itself is the system git (version not recorded in this run).

## Residual risk

- A name containing `refs/` that is a tag path still resolves by git rule order; exact-path `refs/heads/...` wins first, so only a non-existent qualified ref could resolve to a tag literally named so. This fails toward a different, existing ref, not toward a silently broader one.
- A remote-tracking name (`origin/main`) is resolved bare and could in theory be shadowed by a tag of the same name; the task kept that behaviour.
- A local branch literally named `origin/main` would take precedence over the remote-tracking ref (step 1); this is the intended preference for local branches.
