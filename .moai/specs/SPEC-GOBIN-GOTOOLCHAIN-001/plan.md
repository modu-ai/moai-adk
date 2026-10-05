# Implementation Plan — SPEC-GOBIN-GOTOOLCHAIN-001

Card: t969. Tier S. Plan authored at HEAD `b1ec8602c`, branch `WT-gotoolchain-local`.

## A. Context

The defect, its firing condition, and the reason CI cannot see it are stated once, in
`spec.md` §1, and are not restated here. What this plan adds is the ordering: which
decisions are genuinely open, and in what sequence the run phase should settle them.

## B. Decisions, most-reversible-last

Ordered by change-likelihood: the decisions most likely to be revisited during review come
first; the mechanical edit comes last.

### B.1 — What the test asserts (highest change-likelihood)

The acceptance criterion fixed by the operator is: bind `HOME` to `t.TempDir()`, call
`gobin.Detect`, then assert `<HOME>/go/pkg/mod/golang.org/toolchain@*` is **empty**.

Two properties of that assertion are load-bearing and should be settled before code is
written:

- **Empty, not absent.** The glob may legitimately match nothing at all, and it may match a
  directory that exists but holds no entries. Both are passes. Asserting "the path does not
  exist" would be a narrower claim than the criterion, and would go red for reasons
  unrelated to the defect.
- **The assertion must be able to fail.** A glob that can never match — a typo in the path,
  a pattern anchored to the wrong root — passes vacuously in both the fixed and the unfixed
  tree, and would satisfy nobody. B.2 is what forces it to demonstrate otherwise.

The test lives with the package under test, `internal/runtime/gobin/`, which currently holds
`resolver.go` and `resolver_test.go` and has **zero** occurrences of `GOTOOLCHAIN` or
`toolchain` (measured in this tree in this run; the grep form was confirmed live by a
positive control against `exec.Command`, which returned 2 hits in `resolver.go`).

### B.2 — RED first, and only through the precompiled form

The assertion MUST be observed failing against unfixed `resolver.go` before the fix lands
(REQ-GGT-005). The observation form is not a preference: `go test` cannot produce the
condition, because the runtime prepends the resolved toolchain's `bin` to the test process's
PATH and the child `go` then has nothing to resolve (`spec.md` §1). RED is therefore
observed by building the package's test binary and running it directly:

```
go test ./internal/runtime/gobin -c -o <path>
<path>
```

A `go test` run that passes at this point is **not** evidence the assertion is wrong — it is
the documented behavior of that invocation form, and it is exactly why the defect survived
CI. Reading a green `go test` as a refutation of the RED requirement is the one
misinterpretation this SPEC most needs to prevent.

Card t964's control pair (precompiled FAIL / `go test` PASS) is the prior measurement this
expectation rests on; it is cited, not re-measured, until the run phase produces its own.

### B.3 — The fix

Set `cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")` on both `exec.Command` calls in
`internal/runtime/gobin/resolver.go` — the one in `goEnvGOBIN` and the one in
`goEnvGOPATHBin`. `os.Environ()` is what satisfies REQ-GGT-002: the child keeps `HOME`,
`PATH`, `GOPATH`, and `GOBIN`, and only `GOTOOLCHAIN` is overridden. A bare
`[]string{"GOTOOLCHAIN=local"}` would strip the rest of the environment and change what the
resolver returns — a fallback-chain behavior change, excluded by `spec.md` §4.

`os` is **not** currently in the file's import block (measured in this tree in this run: the
imports are `os/exec`, `path/filepath`, `strings`), so the import must be added.

`go env GOBIN` and `go env GOPATH` are queries that need no module-specific toolchain, which
is why `local` is safe here and would not be safe on an arbitrary `go` invocation.

### B.4 — GREEN in both forms, then the linters

After the fix: the precompiled binary and `go test` must both pass, turning card t964's
FAIL/PASS control pair into PASS/PASS. Then `go vet` and `golangci-lint` on the changed
package only.

When judging `gofmt`, read its **output**, not its exit code — `gofmt -l` exits 0 while
listing files.

## C. Milestones

| Priority | Milestone | Content |
|---|---|---|
| High | M1 | Write the characterization test (B.1); observe RED via the precompiled form (B.2) and record the verbatim failure |
| High | M2 | Apply the two-line fix plus the `os` import (B.3) |
| High | M3 | Observe GREEN in both invocation forms (B.4); record both verbatim |
| Medium | M4 | `go vet` + `golangci-lint` on `./internal/runtime/gobin/...`; record output |

M1 precedes M2 and the ordering is not a formality: an assertion first seen green proves
nothing about what it guards, and the commit graph is the only witness of that ordering
(`verification-claim-integrity.md` §2.3). M1's test and M2's fix therefore land in separate
commits, M1's first.

## D. Risks

- **R-1 — The RED is not reproducible on the run machine.** The defect needs the PATH `go`
  to be older than the module's `go` directive. On a machine where it is not, the
  precompiled binary passes before the fix and RED cannot be observed. This is an
  environmental precondition, not a refutation: the run phase reports it as a gap and does
  not record a RED it did not see.
- **R-2 — The assertion passes vacuously.** Mitigated by B.2: an assertion that has been
  seen red is, by construction, not vacuous.
- **R-3 — The download lands somewhere other than the asserted glob.** The glob is keyed to
  `HOME`; if the Go runtime resolves its module cache from `GOMODCACHE` or `GOPATH` instead,
  the bytes land outside it. The RED observation is what settles this — a red assertion has
  demonstrated it is watching the right path.
- **R-4 — Scope creep toward R2.** Removing the subprocess is excluded (`spec.md` §4) and
  stays excluded even if it looks strictly better while the file is open.

## E. Self-verification (plan phase)

- [ ] Every path cited in this plan verified present at HEAD `b1ec8602c`
- [ ] Every figure either measured in this tree in this run, or attributed to card t964
- [ ] No implementation file touched during the plan phase
- [ ] Out-of-scope R2 recorded as a follow-up candidate, with no requirement or AC written for it

## F. Follow-up candidates (not approved)

- **R2 — remove the subprocess entirely.** Derive GOBIN/GOPATH from the `GOBIN` / `GOPATH`
  environment variables plus the `$HOME/go` default, which the resolver already carries as
  its step-3 fallback. This would remove two process spawns from every `moai init` and every
  `moai update`, and would close the exposure at its root rather than pinning it. It is a
  behavior-surface change to the fallback chain and needs its own SPEC, its own
  characterization of what each step currently returns, and its own approval. Recorded here
  so it is not lost; explicitly **not** authorized by this SPEC.

## G. Cross-references

- `spec.md` — requirements, exclusions, constraints
- `acceptance.md` — AC-GGT-001 .. AC-GGT-006
- `.moai/reports/t964/verdict.md` — the prior measurement this SPEC rests on
