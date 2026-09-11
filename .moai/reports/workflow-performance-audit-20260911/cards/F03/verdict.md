# F03 verdict

## Claim

F03 is fixed: plan-artifact cache identity now uses exact artifact bytes. A
whitespace or indentation change in executable shell/YAML examples produces a
different key instead of colliding after global whitespace normalization.

## Evidence

`InMemoryCache.ComputeHash` now binds each artifact path, byte length, exact
content, and a separator. The former `strings.Fields` normalization was
removed. Regression coverage includes ordinary whitespace changes and two YAML
block-scalar fixtures with different indentation.

Command:

```text
go test ./internal/runtime -count=1
ok  github.com/modu-ai/moai-adk/internal/runtime 1.501s
```

## Baseline-attribution

The command ran in `WT-workflow-audit-f03`, based on local `develop` merge
commit `459deba0c` (which includes F02) before the F03 commit.

## Gaps

No parser-based Markdown semantic equivalence is attempted. This card chooses
the safer byte-exact identity; a future parser optimization would require a
new collision test suite.

## Residual-risk

Byte-exact hashing may invalidate more audits after harmless prose formatting,
increasing audit calls. That trade-off is intentional until measured parser
semantics can prove the affected code blocks safe to normalize.
