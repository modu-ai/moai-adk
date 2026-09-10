# t366 — plan-phase discovery (binary lag on tree-reading CLI)

Tree: worktree `.claude/worktrees/t366`, branch `WT-lint-binary-lag`, HEAD `d7010f86a` (= origin/develop at fetch time).
Installed binary: `/Users/goos/go/bin/moai` — `v3.1.2  343399d2f  built 2026-08-27T14:07:38Z`.

## Q1 — does t326's landing cover this path?

Claim: it does not, and the reason is not "spec lint does not look" but "the surface that
would speak is compiled into the stale binary itself".

Evidence:

    $ which -a moai
    /Users/goos/go/bin/moai

    $ git log --oneline --diff-filter=A -- internal/binlag/binlag.go
    c70c6aed9 feat(t326): make the binary-lag verdict reach an observer unprompted

    $ git merge-base --is-ancestor 343399d2f c70c6aed9; echo $?
    0                       # installed commit is a strict ancestor of t326's landing

    $ strings ~/go/bin/moai | grep -ci binlag
    0                       # the installed binary carries no binlag package

    $ grep -rn checkBinaryFreshness --include='*.go' internal/ | grep -v _test.go
    internal/cli/doctor.go:201:  {"Binary Freshness", checkBinaryFreshness}
    internal/cli/doctor.go:518:  func checkBinaryFreshness(verbose bool) DiagnosticCheck
                            # exactly ONE caller: the doctor registry

    $ ~/go/bin/moai doctor | grep Freshness
    warn  Binary Freshness  binary is behind source tree (binary: 343399d2f, HEAD: d7010f86a)

So: the pre-t326 doctor item exists in the installed binary and reports the lag correctly —
but only when typed by hand. t326's contribution was REACH (the unprompted session-start
advisory), and that code is absent from the installed binary.

Correction (raised by manager-spec, verified here): the sentence above describes the INSTALLED
binary, where the doctor item predates `binlag`. In the TREE the two are already unified —

    $ grep -rln 'binlag' --include='*.go' internal/ | grep -v _test.go
    internal/cli/doctor.go
    internal/binlag/binlag.go
    internal/hook/session_start_binary_lag.go

t326 rewired `doctor.go` onto the shared seam as well, so a new consumer only has to attach to
it. This strengthens the conclusion rather than changing it. The SessionStart wrapper
resolves `moai` from PATH first (`.claude/hooks/moai/handle-session-start.sh`: `command -v moai`
then `exec moai hook session-start`), so the advisory that would announce the lag is itself
gated behind the lagging binary. Bootstrap paradox.

Caveat that lowers severity: the paradox is one-shot. `make build && make install` once, and
the advisory starts firing on every later session. It is not a standing hole in t326's design;
it is the cost of t326 not having been installed.

## Q2 — scope

`moai spec lint` is not special. The lag reaches any subcommand whose answer is derived from
compiled-in rules read against the tree.

    $ grep -rln "findProjectRootFn\|os.Getwd()" --include='*.go' internal/cli/ | grep -v _test.go | wc -l
    68
    $ ls internal/cli/*.go | grep -v _test.go | wc -l
    199

68 of 199 non-test CLI files read the project root or cwd. Coarse (a file-count proxy, not a
command count), but it establishes the direction: the class is broad, which argues for a
generic remedy at the invocation seam rather than a per-command patch.

Gap: no exact per-subcommand census was run. A run-phase repair keyed on "which commands"
needs one.

## Q3 — CI is unaffected

    $ grep -rn 'moai ' .github/workflows/
    spec-lint.yml:40           go run ./cmd/moai spec lint --strict
    ci.yml:367,464             go build -o ./bin/moai ./cmd/moai/ ; ./bin/moai ...
    graph-freshness.yml:78     go build -o ./bin/moai ./cmd/moai ; ./bin/moai ...
    spec-status-auto-sync.yml:85   moai spec status ...   (bare PATH call)

    $ grep -n 'install' .github/workflows/spec-status-auto-sync.yml
    35:        run: make build && make install

Every workflow builds from the checked-out source; the one bare-PATH call is preceded by
`make build && make install` in the same job. CI cannot exhibit this defect. It is
local-only, which puts the repair on the warning/visibility axis rather than the correctness
axis.

## The green itself, observed

The card's premise — a green whose output says nothing about which rules produced it — is
directly observable:

    $ ~/go/bin/moai spec lint            # ran to completion, exit 0
    ... 64 WARNING rows ...
    0 error(s), 64 warning(s)
    rc=0

    $ grep -ci 'lag|behind|freshness' <that output>
    0

Exit 0, and nothing anywhere in the output identifies the rule set as the one compiled at
`343399d2f`. A reader has no signal distinguishing "every rule ran and passed" from "the
rules that would have failed were not in this binary".

## Incidental observation (not this card)

The same invocation exceeded a 120s foreground bound before finishing in the background;
it did complete with exit 0. CI runs the same linter via `go run ... --strict`. Duration
was not measured precisely — recorded so it is not lost, not asserted as a defect.

## Gaps

- No per-subcommand census (Q2 above).
- The plant-a-rule reproduction (warning fires when the sets differ, stays silent when they
  match) is DESIGNED in the plan, not yet executed — plan phase only.
- No measurement of how many lanes are currently running a stale binary; the three cited
  encounters (t343, t362, t357) come from the lead's dispatch, not from a measurement here.
