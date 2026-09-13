# t672 Matrix — real-request cells (isolated gateway children, card t672)

Every cell is a REAL request through a real gateway child started from this
worktree's build (`internal-gateway` entry, ephemeral loopback port, per-run
session token, isolated receipt store under `/private/tmp/t672/receipts-*`,
real credential store reused read-only in place). Upstream replies were real
model replies, not synthesized. Driver: `probe_driver_main.go.txt`.

Two runs, same driver, same cell sequence:

- BEFORE = binary built from `git archive HEAD` (pre-t672, includes t695)
  → `/tmp/t672/moai-before`, sha256 `bf305362e97152df0ab427de2ff62b66e48cfa36`
- AFTER = binary built from this worktree (t672 classification)
  → `/tmp/t672/moai-after`, sha256 `c61a2af50bfa304df4f1e45140eff5b0b1514caf`

| cell | shape | BEFORE | AFTER |
|---|---|---|---|
| C1 | parent fresh turn, sol+low | 200 `{"content":[{"text":"ok",...}]}` | 200 (same shape) |
| C1b | terra reasoning-bearing turn (backs C5/C6) | 200, real `moai_opaque_v2_…` carrier in content | 200 (same shape) |
| C2 | spawn-shape WITH model arg (Agent tool, `model:"opus"` arg requested) | 200 real `tool_use` `{"name":"Agent","input":{"model":"opus",...}}` | 200 (same shape) |
| C3 | spawn-shape no model arg | 200 real `tool_use` | 200 (same shape) |
| C4a | parent second turn | 200 | 200 |
| C4b | fork-shaped TRUNCATED replay (turn 1 only + new user msg) | **200 `"done"` — tail truncation already tolerated** | 200 (unchanged) |
| C5 | wedge replay: assistant turn the gateway never published | 400 base message only (verbatim below) | 400 **classified: chain** (verbatim below) |
| C6 | stripped-reasoning replay of a published boundary | 400 base message only | 400 **classified: reasoning** (verbatim below) |

Verbatim rejection bodies (AFTER run):

- C5:
  `conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation (reason: replayed history does not match the recorded receipt chain; start a new conversation)`
- C6:
  `conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation (reason: replayed history omits gateway-issued reasoning recorded at this position; replay the history unmodified or start a new conversation)`

BEFORE run returned both bodies with the historical sentence only (no
`(reason: …` clause) — the before/after contrast is the classification, and
functional 200/400 behavior is byte-for-byte the same on every cell.

Cell verdicts:

1. Model-arg and no-arg spawn-shaped first turns pass with a live receipt
   root, before and after — a spawn's first turn cannot produce
   HistoryReplayError (mechanism §1.1 of investigation.md).
2. Fork-shaped tail truncation is legitimately accepted, before and after —
   locked by `TestReceiptHistoryAcceptsTruncatedForkReplay` as well.
3. The two rejection classes most likely to explain the production
   spawn/fork/fork-summary incidents (stripped-reasoning replay; unpublished-
   turn wedge) now self-identify in the client-visible body.

## Scoped verification (D5)

- `go vet ./internal/gateway/...` — clean.
- `golangci-lint run --new-from-rev=HEAD internal/gateway/translate/...` —
  `0 issues.`
- `go test ./internal/gateway/... -count=1` — all packages pass EXCEPT the
  known pre-existing environmental failure
  `TestAppServerSubprocessHTTPToolContinuation` (named by the card as
  to-ignore; reproduced at pristine HEAD by t695 §5). `internal/cli` was not
  touched; no full-suite run, no background load.

New/updated tests (`internal/gateway/translate/receipt_history_cause_test.go`,
all passing):

- truncated-fork passes with pairing kept (3 shapes incl. compatible-domain)
- mid-chain truncation rejected with the chain guidance
- mid-pair truncation rejected before translation, chain-classified
- foreign-item (forged envelope) replay rejected, chain-classified
- foreign-root replay rejected, lineage-classified (no seeding exists)
- spawn-shaped fresh history passes on sol and astra with no lineage seeding
- stripped-reasoning replay still rejected, now reasoning-classified
  (security stance locked, not loosened)
- desync wedge shape rejected chain-classified while an exact retry passes
