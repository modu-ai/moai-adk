# t441 — catalog hash: the third consumer joins the shared implementation

Card: t441 · Branch: `WT-catalog-hash-consumers` · Base: local develop `9b89b8c5b`

## Claim

`internal/spec/catalog_hash_test.go` (`TestCatalogHashParity`) computed catalog hashes on the
v1 boundary — resolve a directory entry to its root `SKILL.md` and hash that one file — while
the generator and `TestManifestHashFormat` had both moved to whole-tree hashing in card t323.
Directing the third consumer at the same exported `template.ComputeDirTreeHash` removes the
disagreement on every directory entry.

## Evidence

RED baseline, at base `9b89b8c5b`, `.moai/reports/t441/red-baseline.txt`:

```
$ go test ./internal/spec/ -run TestCatalogHashParity -count=1
exit=1
CATALOG_HASH_DRIFT lines: 35   (35 unique entry names, no duplicates)
  34 with source=.../<skill-dir>/SKILL.md   (directory entries)
   1 with source=.../.claude/agents/moai/sync-auditor.md   (file entry)
```

GREEN, same tree, `.moai/reports/t441/green-spec.txt`:

```
$ go test ./internal/spec/ -run TestCatalogHashParity -count=1 -v
exit=1
CATALOG_HASH_DRIFT lines: 1
  entry "sync-auditor" | stored=f1b4487f… | computed=545d03d9…
--- PASS: TestCatalogHashParity_MissingTemplates
```

Both-directions regression, `.moai/reports/t441/green-template.txt`:

```
$ go test ./internal/template/ -run 'TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles' -count=1
--- FAIL: TestManifestHashFormat
    CATALOG_HASH_UNSTABLE: sync-auditor stored hash=f1b4487f…, computed hash=545d03d9…
(TestCatalogHashCoversSkillSubfiles: PASS — 34 directory entries audited)
```

The two guards now name the same single entry with byte-identical stored and computed values,
so they reach the same verdict on the same corpus. The directory-entry population is unchanged
at 34, so the consumer set did not grow from three to four.

Toolchain, same tree: `gofmt -l` on both touched files printed nothing;
`go vet ./internal/spec/ ./internal/template/` exited 0.

Touched packages in full, `.moai/reports/t441/pkg-full.txt`:

```
$ go test ./internal/spec/... ./internal/template/... -count=1
FAIL internal/spec              — TestCatalogHashParity (sync-auditor)
FAIL internal/template          — TestManifestHashFormat (sync-auditor)
FAIL internal/template/agentemit — TestGoldenCommittedArtifactsMatchEmission
```

## Baseline-attribution

Every figure above was measured in this worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t441`) at HEAD `9b89b8c5b`, which is the local
develop head after fast-forward absorption. No figure is carried over from an earlier tree.

The "generator produces diff 0" premise carried in the dispatch was measured against an earlier
develop and does NOT hold at this SHA — see Gaps.

## The duplication premise, re-measured

`catalog_hash_test.go:65` justified its local `resolveHashSourcePath` copy as
"a deliberate duplicate kept here because the script lives in `package main` and cannot be
imported". Judged in two halves:

- **The stated fact is true.** `internal/template/scripts/gen-catalog-hashes.go:30` is
  `package main`, so it cannot be imported.
- **The safety conclusion attached to it is refuted.** The same comment claimed that sharing
  `template.NormalizeForHash` had eliminated the highest-risk drift surface, leaving "only path
  resolution" local. After t323 the highest-risk surface was no longer normalization but the
  hash BOUNDARY, and the test duplicated that boundary implicitly by always hashing one file.
  The 34 directory-entry failures are that refutation: every shared normalization call agreed
  while the two sides disagreed on what to normalize.

Direction taken: (가). `ComputeDirTreeHash(fsys fs.FS, dir string)` is FS-agnostic, so the parity
guard passes `os.DirFS(templatesDir)` exactly as the generator does and keeps its distinct
purpose — reading on-disk source rather than the embed. The local helper survives as a
file-only resolver (its now-unreachable SKILL.md branch is removed) because
`TestCatalogHashParity_MissingTemplates` exercises its error propagation for REQ-CHR-007.

`internal/template/catalog_tree_hash.go` said "Both consumers"; it now says three and names them.

## Gaps

- **Not verified: whether the generator would close the residual entry.** `gen-catalog-hashes.go`
  was NOT executed in this card, in any mode, including `--dry-run`. The dispatch's "generator
  diff 0" figure was measured against an earlier develop and is not re-measured here.
- **Not attempted: the residual `sync-auditor` red.** Left in place deliberately, see below.
- **Not run: any package outside `internal/spec` and `internal/template`.** The change touches a
  test file and a doc comment; no other package imports either symbol path.
- **Not run: the full suite, and no CI verdict.** Lane-local scope only.

## Residual-risk

The repair reads a directory entry's tree from on-disk source while the sibling guard reads it
from the embed. That asymmetry is the point of having both guards, but it means a tree edited
without `make build` now shows as agreement here and disagreement there — the reverse of the
pre-t323 situation is not reintroduced, but the two guards remain non-interchangeable and a
future reader may mistake one for the other.

`resolveHashSourcePath` still double-stats: the caller stats the entry to choose a branch, and
the helper stats it again. Kept so the file branch continues to exercise the production path the
REQ-CHR-007 sub-test asserts against.

## Out of scope — the residual `sync-auditor` red (NOT this card)

Two failures survive this card and share one root cause. Commit `4244c4a06`
(`docs(t386/t387): sync-auditor export-mandate clause`) edited
`internal/template/templates/.claude/agents/moai/sync-auditor.md` without refreshing the two
artifacts derived from it:

| Stale artifact | Failing guard | Remedy |
|---|---|---|
| `catalog.yaml` `sync-auditor` hash | `TestCatalogHashParity`, `TestManifestHashFormat` | `go run internal/template/scripts/gen-catalog-hashes.go --entry sync-auditor` |
| `.codex/agents/moai/sync-auditor.toml` | `TestGoldenCommittedArtifactsMatchEmission` | `make agents-emit` |

This is a file entry, not a directory entry, and it fails identically on both the on-disk and
the embedded reading — so it is not the v1/v2 boundary defect this card closes. It is a separate
source-body drift with its own regression surface, and it is left for the lead to card.

Note for whoever takes it: the dispatch's instruction not to reach for `gen-catalog-hashes`
was correct for the 34 boundary failures and does not extend to this entry, whose catalog value
is genuinely behind its source.
