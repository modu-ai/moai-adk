# t526 — catalog.yaml hash refresh after t497 neutral-source revision

card: t526 · branch `WT-catalog-hash-refresh` · base `0b1e27877` (= origin/develop)
worktree: `.claude/worktrees/t526` (t508's tree was exited first — one card per tree)

## Claim

1. Two guards were RED on `0b1e27877`, both naming the same four catalog entries.
2. The four entries correspond exactly to the four `.claude/agents/moai/*.md` neutral sources
   revised by t497's commit.
3. `gen-catalog-hashes.go --all` changes exactly those four entries and nothing else.
4. After regeneration both guards are GREEN, and a one-line mutation to a neutral source drives
   both back to RED — so the guards are live, not merely satisfied.

## Evidence

### Independent attribution (not carried over from the lead's report)

This session was `/clear`ed and holds no t497 context, so the lead's attribution was re-measured
rather than cited.

```
$ git log --oneline -3 -- internal/template/templates/.claude/agents/moai/manager-develop.md \
    .../manager-lead.md .../manager-design.md .../e2e-tester.md
321111fe5 feat(SPEC-CODEX-BODY-NEUTRALITY-001): M3+M4 중립 소스 본문 개정 + 골든 재생성

$ git show --stat --oneline 321111fe5
 .../templates/.claude/agents/moai/e2e-tester.md     | 2 +-
 .../.claude/agents/moai/manager-design.md           | 2 +
 .../.claude/agents/moai/manager-develop.md          | 4 +-
 .../templates/.claude/agents/moai/manager-lead.md   | 16 ++--
 (+ the .codex/*.toml mirrors, AGENTS.md, progress.md)
```

The four revised `.claude` neutral sources are exactly the four drifted catalog entries.
Attribution confirmed independently.

### RED baseline (before the change)

```
$ go test ./internal/spec/ -run TestCatalogHashParity -count=1
--- FAIL: TestCatalogHashParity (0.02s)
    CATALOG_HASH_DRIFT: entry "manager-develop" | stored=ad390185… | computed=f3c720e3…
    CATALOG_HASH_DRIFT: entry "manager-lead"    | stored=169cf43a… | computed=a5d6809e…
    CATALOG_HASH_DRIFT: entry "manager-design"  | stored=fb7af236… | computed=3863fa6a…
    CATALOG_HASH_DRIFT: entry "e2e-tester"      | stored=910a13c9… | computed=9a8894af…
FAIL

$ go test ./internal/template/ -run TestManifestHashFormat -count=1
--- FAIL: TestManifestHashFormat (0.01s)
    CATALOG_HASH_UNSTABLE: manager-develop / manager-lead / manager-design / e2e-tester
    audited 45 catalog entries for hash validity
FAIL
```

Both guards name the same four entries and the same computed hashes — they agree.

### Correction to the dispatch

The dispatch located both failing tests in `internal/spec`. `TestManifestHashFormat` is **not**
in that package:

```
$ go test ./internal/spec/ -run TestManifestHashFormat -count=1
testing: warning: no tests to run
ok  [no tests to run]
```

It is defined at `internal/template/catalog_tier_audit_test.go:393`. A verification scoped to
`./internal/spec/` alone would have reported one guard GREEN by never running it — the same
zero-match-passes shape the dispatch was warning about. The verification below therefore covers
**both** packages.

### Scope of the regeneration — exactly four entries

```
$ go run internal/template/scripts/gen-catalog-hashes.go --all
catalog.yaml updated successfully (12899 bytes)

$ git diff --stat internal/template/catalog.yaml
 internal/template/catalog.yaml | 8 ++++----
 1 file changed, 4 insertions(+), 4 deletions(-)
```

Four `hash:` lines, one per drifted entry (`manager-develop`, `manager-lead`, `manager-design`,
`e2e-tester`). No other entry moved, and `generated_at` did not churn. `--all` recomputes every
entry but rewrites only what differs — the dispatch's out-of-scope condition did not occur.

### GREEN

```
$ go test ./internal/spec/ ./internal/template/ -run 'TestCatalogHashParity|TestManifestHashFormat' -count=1
ok  github.com/modu-ai/moai-adk/internal/spec      0.220s
ok  github.com/modu-ai/moai-adk/internal/template  0.428s
```

### Mutant — the guards are still live

A one-line comment appended to one neutral source (`e2e-tester.md`), then removed:

```
$ printf '\n<!-- t526 mutant probe -->\n' >> …/e2e-tester.md

$ go test ./internal/spec/ -run TestCatalogHashParity -count=1
--- FAIL: TestCatalogHashParity
    CATALOG_HASH_DRIFT: entry "e2e-tester" | stored=9a8894af… | computed=82f5a93e…
FAIL

$ go test ./internal/template/ -run TestManifestHashFormat -count=1
--- FAIL: TestManifestHashFormat
    CATALOG_HASH_UNSTABLE: e2e-tester stored=9a8894af…, computed=82f5a93e…
FAIL
```

Both guards caught it. Restored byte-identically:

```
$ cp /tmp/t526-e2e-tester.md.orig …/e2e-tester.md
$ git status --short …/e2e-tester.md
(no output — byte-identical to HEAD)
```

Without this probe, "the hashes now match" and "the guard stopped checking" look the same.

### Full affected-package suites

```
$ go test ./internal/spec/ ./internal/template/ -count=1
ok  github.com/modu-ai/moai-adk/internal/spec      67.729s
ok  github.com/modu-ai/moai-adk/internal/template  25.733s
```

### Final working tree

```
$ git status --short
 M internal/template/catalog.yaml
```

One file. The mutation left nothing behind.

## Baseline-attribution

Every command above ran in this session, in `.claude/worktrees/t526` at base `0b1e27877`, on
this machine. Nothing is inherited from the lead's report or from a prior run.

## Gaps

- CI was not run from here (lane policy: no WT push, no CI dispatch). The verdict rests on
  local runs of both affected packages; the CI verdict follows the develop push.
- The four `.codex/*.toml` mirrors t497 also revised are NOT covered by these two guards, and
  whether a separate guard covers them was not investigated — out of this card's scope, but
  worth a look if the mirrors have their own hash record.
- Why the drift escaped t497's own verification is recorded by the lead as a process defect
  (the card's radius omitted `catalog.yaml`; the verification scope omitted `internal/spec`).
  Not re-derived here.

## Residual risk

`--all` recomputes every entry from current sources. Had another entry's source drifted
independently before this run, `--all` would have silently absorbed that drift too. The
`git diff` above rules it out for this run — four lines, all four attributable — but the check
is the diff, not the generator.
