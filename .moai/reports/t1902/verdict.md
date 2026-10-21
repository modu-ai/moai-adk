# Card t1902 — Verdict: sender lane label + harness in factory inbox metadata

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1902` · branch `WT-factory-inbox-stop`
Commits: `926c7899a9` (implementation, parent `e29a5c9be2`) → `27087e844a` (ruling mapping, parent `926c7899a9`) · status clean (evidence file gitignored per `.gitignore:235`)
Pushed: `origin/WT-factory-inbox-stop` at `27087e844a` (push tail: `926c7899a9..27087e844a`, exit 0 — the remote already carried `926c7899a9`, pushed by the lane)

## Leader ruling (applied)

The harness-vocabulary question raised in this verdict's first revision is RESOLVED by leader ruling: the render layer maps the recorded `gpt` to the display token `codex`; the registration record stays unchanged. Rationale (three points, per the ruling): (1) `internal/factory/factory_slots.go:184-186` records the Codex launcher's `gpt` deliberately (the registry's internal vocabulary); (2) the card's announced output vocabulary is `codex|claude|glm`; (3) `gpt` defeats instant recognition of a Codex lane.

## Claim

1. Each factory inbox metadata line additionally names the sender endpoint's registered lane label and harness: `lane=<slot> harness=<backend>`, values from the sender's `peers` registration record, with the single ruled display mapping gpt→codex (`factoryHarnessDisplay`).
2. The `from=` key is unchanged; the new fields are appended after it (consumer compatibility).
3. No registration record (unregistered endpoint), a failed lookup, or an empty field renders `unknown`; only record-present values print (plus the ruled mapping).
4. No existing consumer of the line format breaks; the REQ-SRL-009 env-literal guard stays at zero literals.

## Evidence

### Metadata line — before/after (verbatim from `internal/hook/factory_messages.go`)

Before (release head `8249bff8c3`):

```go
ids = append(ids, fmt.Sprintf("%s kind=%s from=%s generation=%d token=%s", c.ID, c.Kind, c.SenderSession, c.RecipientGeneration, c.ClaimToken))
```

Rendered: `<msg-id> kind=status_request from=<sender-uuid> generation=1 token=<token>`

After (`27087e844a`):

```go
ids = append(ids, fmt.Sprintf("%s kind=%s from=%s lane=%s harness=%s generation=%d token=%s", c.ID, c.Kind, c.SenderSession, meta.lane, meta.harness, c.RecipientGeneration, c.ClaimToken))
```

Rendered: `<msg-id> kind=status_request from=<sender-uuid> lane=lane-46 harness=codex generation=1 token=<token>`

### Mapping diff (verbatim, commit `27087e844a`)

```go
// factoryHarnessDisplay is the render-layer mapping from the registry's
// recorded backend vocabulary to the display token the inbox reader expects.
// The Codex launcher's token is deliberately recorded as gpt — the vocabulary
// every other surface speaks (internal/factory/factory_slots.go:184-186; the
// launcher exports config.EnvFactoryBackend=gpt,
// internal/cli/codex_launcher.go) — but the leader ruling on card t1902
// renders it back as codex, the card's announced output vocabulary, so a
// Codex lane reads as one at a glance. The registration record itself is
// unchanged: this is a display translation only, and every other recorded
// value passes through.
var factoryHarnessDisplay = map[string]string{
	factory.BackendGPT: "codex",
}
```

Applied inside `factoryHookSenderInfo` (which reads the sender's `peers` row via the sanctioned `Store.Peer(ctx, sessionUUID)` accessor, `internal/factorymsg/store.go:682`, memoized once per unique sender per batch):

```go
	lane, harness := p.Slot, p.Backend
	if shown, mapped := factoryHarnessDisplay[harness]; mapped {
		harness = shown
	}
```

Empty fields → `unknown`; `sql.ErrNoRows` → both `unknown`; any other lookup error → both `unknown` + `slog.Debug` (the claim proceeds — withholding a bound lane's inbox over a metadata decoration would repeat the false-unbound defect class named in the function's existing comment).

### Three-case test (verbatim, named run in this worktree, post-ruling tree)

Command: `TMPDIR=/tmp go test ./internal/hook -run 'TestNoNewEnvLiteralsInDiff|TestFactoryHookMetadataShowsSenderLaneAndHarness' -count=1 -p 1 -v`

```
=== RUN   TestNoNewEnvLiteralsInDiff
    env_literal_diff_test.go:52: env-literal sweep: 240 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=8249bff8c3933d647b4f99e59048595f9808e83a, distinct literals=0
--- PASS: TestNoNewEnvLiteralsInDiff (0.17s)
=== RUN   TestFactoryHookMetadataShowsSenderLaneAndHarness
=== RUN   TestFactoryHookMetadataShowsSenderLaneAndHarness/registered_codex_sender
=== RUN   TestFactoryHookMetadataShowsSenderLaneAndHarness/registered_claude_sender
=== RUN   TestFactoryHookMetadataShowsSenderLaneAndHarness/unregistered_sender
--- PASS: TestFactoryHookMetadataShowsSenderLaneAndHarness (0.74s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.517s
EXIT=0
```

The three subtests pin the full line format (dynamic `token=` excluded via prefix match):

| Case | Record backend | Expected line prefix |
|---|---|---|
| registered codex sender (slot `lane-46`) — the mapping's regression witness | `gpt` (`factory.BackendGPT`) | `<id> kind=status_request from=meta-sender lane=lane-46 harness=codex generation=1 token=` |
| registered claude sender (slot `lane-7`) — pass-through | `claude` | `<id> kind=status_request from=meta-sender lane=lane-7 harness=claude generation=1 token=` |
| unregistered sender (sent, then `DELETE FROM peers WHERE session_uuid='meta-sender'`) | n/a | `<id> kind=status_request from=meta-sender lane=unknown harness=unknown generation=1 token=` |

### Full package, lint, build (verbatim tails, post-ruling)

- `TMPDIR=/tmp go test ./internal/hook -count=1 -p 1` → `ok  	github.com/modu-ai/moai-adk/internal/hook	329.192s`, EXIT=0, 0 FAIL lines
- `golangci-lint run ./internal/hook` → `0 issues.`, LINT_EXIT=0
- `TMPDIR=/tmp go build ./...` → (no output), BUILD_EXIT=0

### Incident during re-verification (recorded)

The first post-mapping full-package run FAILED: `TestNoNewEnvLiteralsInDiff` (REQ-SRL-009 guard, AC-SRL-009) swept the committed branch diff (base `8249bff8c3`, 239 added lines) and found one new env-name literal — `MOAI_KANBAN_BACKEND`, which the first draft of the mapping-table comment named literally. The run-1 full pass had not caught this because the implementation was then still uncommitted (the guard diffs committed ranges). Fix: the comment now names the Go constant (`config.EnvFactoryBackend`), never the literal. Post-fix run: `distinct literals=0`, guard PASS (quoted above). No production behavior changed; the failure was comment-text only.

### Consumer-regression note

Consumers of the line format found by repo-wide grep for `Factory inbox run=`: `internal/hook/stale_run_gate_test.go:247`, `internal/hook/factory_rebind_test.go:430`, `internal/hook/factory_rebind_test.go:440` — all assert with `strings.Contains` on the run prefix and/or the message id, none asserts full-line equality. The change is purely additive to each line, so zero consumer tests required modification (the only test-file delta is the new three-case test). Both consumer tests are inside the full-package run above (exit 0). No non-test consumer of the per-line format exists in the repo (the line is hook-injected context read by the message recipient).

### Base refresh (leader directive mid-task)

Batch head `e29a5c9be2` merged before implementation. The first merge fast-forwarded (8249bff8c3 was an ancestor), so no separate merge commit exists; the `--no-ff` redo (`git reset --hard` + merge) was denied by the permission system and not retried (deny is final). Dependency SHA `e29a5c9be2` is recorded here and as `926c7899a9`'s parent. `factory_messages.go` had no batch divergence; the batch's only touched file in scope was `internal/factorymsg/store.go` (+14 lines, an unrelated test seam).

## Baseline-attribution

- All measurements were taken in this run, in this worktree, on the tree whose content is exactly commit `27087e844a` (working tree staged and committed without further edits after the runs).
- The named three-case + guard run, the full-package run, lint, and build are the verbatim outputs quoted above; none is carried over from a prior run or another tree.
- Tool provenance: `golangci-lint` and the Go toolchain are installed builds, not built from this tree (§2.2 tool-provenance). Stated per the clause rather than verified as non-ancestor.

## Gaps

- `go test ./...` (full repo) not run — scoped verification per the card; the full suite is CI's verdict surface.
- The non-ErrNoRows lookup-failure branch of `factoryHookSenderInfo` (debug log + unknown) has no dedicated test: `Store` exposes no fault seam on `Peer`. Its behavior is asserted here by code reading only.
- The >2048-char truncation interaction (a batch near `factoryHookContextLimit` can cut a now-longer line mid-field) is pre-existing behavior, unchanged by this card, and was not re-verified with the new fields.
- A record holding the literal `codex` (possible via a hook path launched with the raw token) also renders `harness=codex` — indistinguishable on the line from a mapped `gpt` record; only the record distinguishes them. Not separately tested.

## Residual-risk

- The display mapping is render-layer only; any other surface that prints `peers.backend` (roster, LaneStatus, MCP outputs) still shows `gpt`. Extending the mapping there would be a separate decision.
- Unknown-on-lookup-failure conflates "unregistered" with "momentarily unreadable" on the rendered line; the divergence is distinguishable only in the debug log.
- A legacy slot label, if present in the record, renders verbatim rather than normalized (record-verbatim rule).

## Commits and push

- `926c7899a9` — `feat(hook): show sender lane label and harness in factory inbox metadata (card t1902)` (2 files, +142/−1)
- `27087e844a` — `feat(hook): render the recorded gpt harness as codex in inbox metadata (card t1902 ruling)` (2 files, +33/−9)
- Both committed via `-F` message files, staged by explicit pathspec.
- Pushed: `origin/WT-factory-inbox-stop` → `27087e844a` (exit 0).
