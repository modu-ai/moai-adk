# t201 — constitution validate honours a retirement marker (#1595)

Class B (defect, cause established during run). Branch `WT-constitution-retire`, base `28bde4022`.

## Claim

`moai constitution validate` now skips drift, canary-gate, and source-file checks for a registry entry whose `clause` begins with `[SUPERSEDED …]`, counts it as retired, and keeps it visible in the report. `--strict` restores verbatim checking.

## Cause

`internal/constitution/validator.go` `Validate` checked every entry's clause verbatim against its source file. A retired clause's text is gone from source by construction, so the DRIFT check could only ever fail — deleting the entry was the only escape, which is what destroyed the audit trail the issue describes.

The issue's second symptom (`canary_gate: false`) was investigated and is **not** a second marker. The four entries named in the issue (`CONST-V3R2-021..024`) are `zone: Evolvable`, for which `canary_gate: false` is the documented default; measured, they raise DRIFT only, never `FROZEN_WITHOUT_CANARY`. Honouring `canary_gate: false` as a retirement marker would have disabled the drift check for most of the registry. Retirement is the `[SUPERSEDED …]` prefix alone; the canary invariant is relaxed for a retired Frozen entry as a *consequence* of retirement.

The marker is a prefix rather than a substring for a measured reason: `CONST-V3R2-152`'s live clause instructs that superseded memory entries be marked `[SUPERSEDED by <new-file>]`, and a substring test would silently retire it.

## Evidence

Baseline reproduction, this tree, before the change (65 errors was the primary checkout at an older HEAD; this tree measures 67):

```
$ go run ./cmd/moai constitution validate 2>&1 | grep -E "CONST-V3R2-02[1-4]"
  [DRIFT] CONST-V3R2-021 @ CLAUDE.md #14-parallel-execution-safeguards
  [DRIFT] CONST-V3R2-022 @ CLAUDE.md #14-parallel-execution-safeguards
  [DRIFT] CONST-V3R2-023 @ CLAUDE.md #14-parallel-execution-safeguards
  [DRIFT] CONST-V3R2-024 @ CLAUDE.md #14-parallel-execution-safeguards
```

After, against the same registry:

```
$ go run ./cmd/moai constitution validate --format json | grep -E '"(drift|retired)_count"'
  "drift_count": 63,
  "retired_count": 4,

$ go run ./cmd/moai constitution validate --strict --format json | grep -E '"(drift|retired)_count"'
  "drift_count": 67,
  "retired_count": 0,
```

63 + 4 = 67: the four skipped entries are exactly the retired ones, and `--strict` reproduces the pre-change count byte for byte. No other entry changed state.

Regression tests (`internal/constitution/retirement_test.go`, fixture-based, written before the fix and observed failing to compile on `RetiredCount` / `IsRetiredClause`):

- retired entry raises no DRIFT and reports `RetiredCount: 1`
- retired Frozen entry with `canary_gate: false` raises no `FROZEN_WITHOUT_CANARY`
- retired entry whose source file is deleted raises no fatal `SOURCE_FILE_MISSING`
- `--strict` reports the retired entry as DRIFT
- control: a non-retired absent clause still reports DRIFT
- marker-boundary table, including the mid-clause mention that must NOT retire

```
$ go test ./internal/constitution/... ./internal/cli/ -count=1
ok  github.com/modu-ai/moai-adk/internal/constitution  0.244s
ok  github.com/modu-ai/moai-adk/internal/cli           406.930s

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/template            23.443s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit   0.473s

$ golangci-lint run ./internal/constitution/... ./internal/cli/
0 issues.

$ go build ./...   # clean
$ make build       # templates re-embedded; catalog.yaml unchanged
```

## Defect found in this card's own work (caught by CI, fixed)

The first push broke `Constitution Check`. The documentation section added to `zone-registry.md` carried a yaml-tagged example fence, and the loader (`extractYAMLFence`) takes the **first** yaml-tagged fence in the document as the entry list — so the example became the registry:

```
Registry load error ".claude/rules/moai/core/zone-registry.md":
  YAML parsing error: yaml: unmarshal errors: line 2: cannot unmarshal !!map into []constitution.rawEntry
```

Retagging the fence `text` was not sufficient: the explanatory sentence spelled the fence marker out literally in prose, and `strings.Index(content, "```yaml")` matched that too, producing a second, different parse error (`line 5: mapping values are not allowed in this context`). Both are now avoided by never writing the literal marker in this file.

`TestShippedRegistriesLoad` (`internal/constitution/shipped_registry_test.go`) was added so this cannot recur silently: it loads both shipped registries — local and template mirror — and fails on a parse error. Nothing previously covered the real files; the CI job that noticed is `continue-on-error: true`, i.e. advisory, so a red run there does not block a merge.

After the fix, both CI steps reproduce locally:

```
$ ./bin/moai constitution list --format json | ... len(entries)
entries: 101          # identical to the base registry at 28bde4022
$ ./bin/moai constitution list --zone frozen | tail -1
Total 57 entries
```

## Baseline attribution

All figures measured in `.claude/worktrees/t201` at base `28bde4022`, this run. The 65-error figure quoted in the issue's repro is from the downstream mo.ai.kr checkout and is not this tree's baseline.

## Gaps

- Not measured: the downstream mo.ai.kr registry itself. The fix is verified against this repository's registry, which carries the same four entries verbatim.
- Not run locally: the full `go test ./...` suite and the darwin/windows matrix — deferred to CI on the PR head, per the repository's local test policy.
- Not addressed: the 63 remaining DRIFT entries in this repository's registry. Pre-existing, unrelated to retirement, out of this card's scope.
- Not added: a way to retire an entry from the CLI (`moai constitution retire <id>`). The marker is hand-edited into the registry today.

## Residual risk

- An entry retired by mistake goes unchecked silently. Mitigated by the reported retired count and by `--strict`, but a maintainer who reads neither would not notice.
- `[SUPERSEDED` is matched case-sensitively. A registry using `[Superseded` would not be recognised; no such entry exists in this repository.
