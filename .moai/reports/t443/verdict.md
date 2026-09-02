# t443 — the two artifacts derived from `sync-auditor.md` catch up with their source

Card: t443 · Branch: `WT-sync-auditor-derived` · Base: local develop `5a8449859`
Class B (defect; `plan` skipped) — the cause was established mechanically before any repair.

## Claim

Commit `4244c4a06` (`docs(t386/t387): sync-auditor export-mandate clause`) added two lines to
`internal/template/templates/.claude/agents/moai/sync-auditor.md` and its local mirror, and
refreshed neither artifact derived from that file. Three guards failed on one root cause.
Regenerating exactly those two artifacts — the `catalog.yaml` entry hash and the emitted
`.codex` TOML — closes all three, and touches nothing else.

## Cause — established, not assumed

```
$ git show --stat 4244c4a06
 .claude/agents/moai/sync-auditor.md                             | 2 ++
 internal/template/templates/.claude/agents/moai/sync-auditor.md | 2 ++
 2 files changed, 4 insertions(+)
```

Exactly two files, both the source `.md` (C1 local + C2 template mirror per CLAUDE.local.md
§2.0). Neither `internal/template/catalog.yaml` nor
`internal/template/templates/.codex/agents/moai/sync-auditor.toml` appears — so the derived
layer was left behind by omission, not by a competing edit.

## Evidence

### RED, at base `5a8449859`

`.moai/reports/t443/red-spec.txt`:
```
$ go test ./internal/spec/ -run TestCatalogHashParity -count=1
exit=1
CATALOG_HASH_DRIFT: entry "sync-auditor" | stored=f1b4487f… | computed=545d03d9…
```

`.moai/reports/t443/red-template.txt`:
```
$ go test ./internal/template/ -run TestManifestHashFormat -count=1
exit=1
CATALOG_HASH_UNSTABLE: sync-auditor stored hash=f1b4487f…, computed hash=545d03d9…
    audited 45 catalog entries for hash validity
```

`.moai/reports/t443/red-agentemit.txt`:
```
$ go test ./internal/template/agentemit/ -run TestGoldenCommittedArtifactsMatchEmission -count=1
exit=1
.codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch)
```

### Repair 1 — narrowed catalog regeneration

`.moai/reports/t443/gen-dryrun.txt` then `.moai/reports/t443/gen-apply.txt`:
```
$ go run internal/template/scripts/gen-catalog-hashes.go --entry sync-auditor --dry-run
  [dry-run] sync-auditor: 545d03d9…  (source: …/templates/.claude/agents/moai/sync-auditor.md)
  [dry-run] catalog.yaml not modified
$ go run internal/template/scripts/gen-catalog-hashes.go --entry sync-auditor
  sync-auditor: 545d03d9…
  catalog.yaml updated successfully (12899 bytes)
```

The dry run's computed value equals the value both failing guards named as `computed`, so the
generator and the guards agree on the target before anything is written.

**`--all` was NOT used, and `catalog.yaml` was NOT wholesale-regenerated.** That is the
representative-mutant hazard: a full regeneration reaches GREEN whether or not this entry was
the only stale one, and silently absorbs any other entry's drift into the same commit. The
narrowing is verified by the diff, not asserted:

```
$ git diff --stat internal/template/catalog.yaml
 internal/template/catalog.yaml | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)
```

One line, the `sync-auditor` `hash:` field. No other entry moved. Whether any other entry is
independently stale is therefore still unmeasured here and stays visible to t444.

### Repair 2 — codex artifact re-emission

`.moai/reports/t443/agents-emit.txt`:
```
$ make agents-emit
AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.447s
```

```
$ git diff --stat internal/template/templates/.codex/agents/moai/sync-auditor.toml
 …/sync-auditor.toml | 2 ++
 1 file changed, 2 insertions(+)
```

Two lines added — the same export-mandate clause `4244c4a06` put in the `.md`, re-emitted into
the TOML. `make agents-emit` regenerates every agent, so the change set is the falsifiable part:
only this one artifact moved, meaning no other C2→C3 pair was concurrently stale.

### GREEN, same tree

`.moai/reports/t443/green-pkg-full.txt`:
```
$ go test ./internal/spec/... ./internal/template/... -count=1
exit=0
ok  	github.com/modu-ai/moai-adk/internal/spec              93.369s
ok  	github.com/modu-ai/moai-adk/internal/template          32.519s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit 1.082s
```

All three named guards pass, and no other test in the two touched package trees regressed.

## Baseline-attribution

Every figure above was measured in this worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t443`) at base `5a8449859` — the local develop
head after the t441 merge — which this branch fast-forwarded to before any change. No figure is
carried over from another tree or another card.

The three-guard failure set was first observed in card t441's window re-measurement and is
re-established here from scratch rather than cited: the RED logs above were produced in this
worktree, not copied.

## Gaps

- **Not measured: whether any catalog entry other than `sync-auditor` is stale.** Deliberate —
  narrowing to one entry is what keeps this card from masking that question. t444 owns it.
- **Not run: `make build`.** The binary was not rebuilt, so no claim is made about what an
  installed binary embeds. `go test` compiles the embed fresh from the committed tree, which is
  what the three guards read; the installed-binary axis (`make embed-check`) is untouched.
- **Not run: any package outside `internal/spec` and `internal/template`.** No Go source changed
  in this card — only a YAML hash field and a generated TOML — so no other package's compilation
  or behaviour is affected.
- **Not run: the full suite, and no CI verdict.** Lane-local scope.
- **Not re-measured: the ten inherited develop reds outside these packages**
  (`TestGateGraphFreshness_AllLayersFreshNotice` and the nine Doctor tests). Named, not measured.

## Residual-risk

The two repairs restore agreement at this instant; neither adds a mechanism that would have
caught the omission when `4244c4a06` landed. The same shape recurs whenever a
`templates/.claude/agents/**.md` edit ships without `gen-catalog-hashes` and `make agents-emit`
in the same commit — the guards fail only later, on whichever card next runs the suite, and are
then attributed to that card rather than to the edit that caused them. This card repairs the
instance; the recurrence surface is unchanged.

`make agents-emit` regenerates all agents, so a future run that emits more than the intended
artifact will look identical in kind to this one. The discriminator is the change-set size,
which is why it is recorded above rather than summarized.
