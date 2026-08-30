# D4 probe — `set -e` defeats the naive `rc=$?` capture

Card t358 · SPEC-CI-TEST-OBSERVABILITY-001 · measured 2026-08-29 in worktree
`.claude/worktrees/t358` at HEAD `c6aa61346`.

## Claim

The plan's original prescription `go test … > f; rc=$?; <census>; exit $rc` does
not print the census on a red run, because GitHub Actions runs `run:` bodies
under `-e`. The corrected form `rc=0; go test … > f || rc=$?` does.

## Baseline-attribution — where the shell string comes from

Not assumed. Read from a real CI run of this repository's own `ci.yml`:

- run `33173944485` (push, base commit `c6aa61346`), job `98857764037` (`Race Test`)
- fetched with `gh run view --job 98857764037 --log-failed`
- step header, line 3 of that step's log, verbatim:

```
shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
```

`-e` is present. The SPEC's earlier draft cited `pipefail` from this same string
and missed `-e`.

## Evidence

Both scripts in this directory were run with the flags above. The command inside
each is `go test -count=1 -json ./internal/version/...` — a path that is not a
package in this tree, so it exits non-zero deterministically (a red run without
needing to plant a failing test).

WRONG form (`t358_wrong.sh`):

```
$ bash --noprofile --norc -e -o pipefail /tmp/t358_wrong.sh; echo "WRONG form step rc=$?"
WRONG form step rc=1
```

The step's exit status is correct — and `CENSUS RAN` never printed. `-e`
terminated the script at `go test`; `rc=$?`, the census, and `exit $rc` were
never reached.

CORRECT form (`t358_correct.sh`):

```
$ bash --noprofile --norc -e -o pipefail /tmp/t358_correct.sh; echo "CORRECT form step rc=$?"
CENSUS RAN (rc=1)
CORRECT form step rc=1
```

Census printed, and the non-zero status still propagated.

## Why this defect would have shipped

On a **green** run the two forms are indistinguishable: `go test` succeeds, `-e`
never fires, both print the census and both exit 0. The divergence appears only
on a red run — the one case where the census is most wanted. A reviewer checking
that "the census works" against a passing run would have seen it working.

This is the same shape as the defect the SPEC exists to close: a check that
exists, reports success, and observes nothing in the case that matters.

## Gaps — explicitly not observed

- Not run on a GitHub-hosted runner. The shell string is quoted from a real run
  on this repository, but these two probes executed locally on darwin. The
  behaviour under test is `bash` option semantics, not runner infrastructure.
- The bash builds differ (local vs `/usr/bin/bash` on ubuntu-latest); `-e`
  semantics for a simple command followed by `;` versus `||` are POSIX and not
  version-dependent, but that generality is reasoning, not measurement here.
- The probe uses a non-existent package path to force rc≠0. It therefore
  exercises the `[setup failed]` path, not an in-suite `--- FAIL` path. Both
  yield a non-zero `go test` status, which is what `-e` reacts to.

## Residual risk

A later editor may "simplify" `|| rc=$?` back to `; rc=$?` on the grounds that
the latter reads more naturally, and no green run will contradict them. The plan
carries the prohibition and the reason for exactly this reason.
