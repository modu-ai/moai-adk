# F05 verdict

## Claim

F05 is resolved in the current contract: the Stop hook's manifest diff is
explicitly an informational observation, while dependency vulnerability and
supply-chain review remains a separate security-review step. The hook no longer
claims that a manifest diff is a vulnerability scan.

## Evidence

The regression fixture checks the source contract and runs the real hook in a
temporary Go repository whose sync commit changes `go.mod` and a source file.
It observes `deps_modified=1` in the hook output and log.

Command:

```text
.claude/hooks/tests/test-dependency-observation-contract.sh
PASS: manifest observation is explicit and reports changed dependency files
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f05`, fast-forwarded to local `develop`
commit `aaf64c411` before this card's commit.

## Gaps

No vulnerability database or language-specific scanner was invoked by this
card; that is intentionally a separate security-review responsibility and is
reported as a gap when the selected scanner is unavailable.

## Residual-risk

The manifest list is language-specific and can miss a new ecosystem marker
until the hook's detector matrix is updated. A future scanner integration must
record scanner identity, lockfile hash, and advisory freshness before reusing a
result.
