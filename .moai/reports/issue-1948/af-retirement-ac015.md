# AC-AFL-015 — retirement gate test (lane-20)

Test: `TestRetirement_NoImportOfForgeMirrorGitref` in `internal/agenticfactory/retirement/retirement_test.go`.
It asserts, from `go list -e -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/...` (non-test imports) and the tree:
(a) nothing imports a retired package, (b) no retired directory remains (also when nothing imports it),
(c) no `internal/cli` source names `newFactoryResidentCommand` or is called `factory_resident*`,
(d) positive controls: the import scan sees edges (agenticfactory/core is imported), and ledger, mergeq, claim, core, runner remain.
`MOAI_RETIREMENT_ROOT=<dir>` points it at another checkout of the module.

## Claim 1 — RED on the pre-retirement tree (tree 8d0c47a7f3, parent of retirement 3fd9ab6456)
- Baseline: `git archive 8d0c47a7f3 go.mod go.sum cmd internal/{agenticfactory,cli,homestate,config,hook}` extracted to a scratch dir, test file copied in.
- Command: `MOAI_RETIREMENT_ROOT=<scratch> go test -p 4 -count=1 ./internal/agenticfactory/retirement/`
- Observed: `--- FAIL: TestRetirement_NoImportOfForgeMirrorGitref (1.38s)`, 32 failure lines, e.g.
  `effect/ghexec imports the retired package .../agenticfactory/forge`,
  `resident imports the retired package .../effect/ctxledger`,
  `the retired directory internal/agenticfactory/forge still exists` (also mirror, gitref, resident, effect/{ghintake,ctxledger,...}).
- Gap: only the first lines were read in full; the factory resident file case was exercised by the 32-line total, not read line by line.

## Claim 2 — GREEN on the retired tree
- Tree: HEAD of branch WT-agentic-monitor-effect-reconcile (integration 03490b9b8f merged as 43759b02fb, plus this commit).
- Command: `go test -p 4 -count=1 -v ./internal/agenticfactory/retirement/`
- Observed: `--- PASS: TestRetirement_NoImportOfForgeMirrorGitref (2.00s)`, `ok ... 2.232s`.

## Claim 3 — mutants die
- Empty retired directory (`internal/agenticfactory/mirror`, nothing imports it) → FAIL `the retired directory internal/agenticfactory/mirror still exists`.
- Tree without `ledger` (scratch copy of HEAD, ledger removed) → FAIL `the package .../agenticfactory/ledger must remain after the retirement`.
- Not exercised: a tree that registers `newFactoryResidentCommand` again (the check is a source scan; its RED is the pre-retirement run above, where the file existed).

## Residual risk
- The resident check is a name scan, not a walk of the cobra tree.
- `effect` itself, docstore.Store, identity's GitHub lookups and intake's GitHub search are outside this test (decided separately, SPEC §E).
