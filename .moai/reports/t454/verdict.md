# t454 — deferred hook side-effects: three judgments

Card: t454 (Class B — plan skipped, run → sync)
Tree: `.claude/worktrees/t454`, branch `WT-hook-goroutine-truth`, base `400f37eb9` (local develop tip)
Re-measured against: this tree. lane-8's observations were taken at `1b78bed0e` (`WT-deferred-side-effects`).

---

## Claim

1. **⑴ `file_changed` gate alignment — lane-8's "disjoint" premise does NOT hold under the
   regex reading of the matcher.** The registered matcher `.env|.envrc|.gitignore` is a regex with
   unescaped dots. Under a regex reading it matches Go source paths containing `env`, and those
   paths also clear the 21-extension gate — so the goroutine DOES spawn in the field for at least
   two real paths in this repository. The judgment lane-8 asked for (widen the matcher vs. retire
   handler+registration) is therefore being asked on a premise that needs settling first.
2. **⑵ `config_change`** — lane-8's premise still holds in this tree, and is stronger than
   reported: not only does the hook-path slog sink discard, but `h.mgr` is never wired by any
   production constructor, so even the non-discarded path reloads a throwaway manager.
3. **⑶a comment fact** — the `file_changed.go` claim *"it runs to completion or the asyncDeadline
   expires"* is false. **Corrected in this card.**
4. **⑶b comment fact** — the *"JSONL append"* claim lane-8 attributed to the `notification` header
   **does not exist** in this tree. Misattribution; nothing corrected.

## Evidence

### ⑴ matcher ↔ extension gate

Registration (identical in template and local settings):

```
internal/template/templates/.claude/settings.json.tmpl:367   "matcher": ".env|.envrc|.gitignore"
.claude/settings.json:343                                    "matcher": ".env|.envrc|.gitignore"
```

Gate (`internal/hook/file_changed.go:20-42`): `supportedExtensions` — 21 source extensions
(`.go .py .ts .js .rs .java .kt .cs .rb .php .ex .exs .cpp .cc .cxx .h .hpp .scala .r .dart .swift`).

Probe (`go run` against this tree, Go `regexp` + `path/filepath`):

```
literal .env         ext=".env"
literal .envrc       ext=".envrc"
literal .gitignore   ext=".gitignore"
regex  internal/config/envkeys.go             match=true  ext=".go"
regex  internal/hook/environment_test.go      match=true  ext=".go"
regex  internal/cli/deps.go                   match=false ext=".go"
regex  internal/hook/file_changed.go          match=false ext=".go"
regex  cmd/moai/main.go                       match=false ext=".go"
regex-and-ext-gate both pass: 2
```

Two readings, two different answers:

| Reading of the matcher | Intersection with the extension gate | Field exposure |
|---|---|---|
| Literal filenames | empty — none of the three literals has a listed extension | 0 (lane-8's finding) |
| Regex over the path | non-empty — `internal/config/envkeys.go` is a real file that clears both | > 0 |

`internal/config/envkeys.go` is not a constructed example; it exists in this tree and is edited
during ordinary work on this repository.

### ⑵ config_change

- `internal/cli/logging.go:59-63` — `resolveLoggingDecision` returns `io.Discard` for the `hook`
  subcommand, unconditionally (`MOAI_LOG_LEVEL` does not re-open it).
- `internal/hook/config_change.go:36-38` — `NewConfigChangeHandler()` leaves `mgr` nil, and it is
  the only production constructor (`internal/cli/deps.go:274`). So `runReload` never takes the
  RT-005 branch; it takes the fallback, which builds a fresh `config.NewConfigManager()`, calls
  `Reload()` on it, and drops it on return.
- Net: the reload mutates nothing that outlives the goroutine, and the two `slog.Info` lines that
  would record it go to `io.Discard`.

### ⑶a comment correction (the one change this card makes)

`internal/hook/file_changed.go` — the `context.Background()` rationale block previously asserted
two outcomes (completion, or deadline expiry). Process exit is the third and, in the field, the
dominant one. Replaced with the accurate statement plus lane-8's measured `0/10`.

### ⑶b JSONL misattribution

```
$ grep -rn 'JSONL' internal/hook/notification.go internal/hook/config_change.go
(no output)
```

`JSONL` appears in the hook package only in `agent_model_guard.go`, `askuser_observer.go`,
`agent_stop_guard.go`, `failure_observer.go`, `navigator_detect.go` — none of them the
`notification` header.

### Sibling sweep

`grep -rn 'runs to completion' internal/ --include='*.go'` → one hit, the one corrected. Four
handlers share the `context.Background(), asyncDeadline` pattern (`task_created.go`,
`file_changed.go`, `config_change.go`, `notification.go`); only `file_changed.go` carried the
false completion promise.

### Verification of the change

```
$ go vet ./internal/hook/          → exit 0
$ go test ./internal/hook/ -run 'FileChanged|ConfigChange|Notification' -count=1
ok  github.com/modu-ai/moai-adk/internal/hook  1.047s
```

## Baseline-attribution

Every figure above was produced in this run, in this tree, at `400f37eb9`. The `0/10` file_changed
loss figure is lane-8's, cited as theirs (`.moai/reports/t448/probe-results.md` @ `1b78bed0e`) and
NOT re-measured here — it is quoted in the corrected comment as a measurement, with its owner
recorded in this report.

## Gaps

- **How Claude Code applies a `FileChanged` matcher is NOT verified.** The probe measures Go's
  `regexp` semantics on the matcher string; it does not establish that the runtime treats the
  matcher as a regex, nor what it matches against (full path, relative path, or basename). Under a
  basename reading, `envkeys.go` would still match `.env` — but this is inference, not measurement.
  Settling ⑴ requires measuring the runtime's matcher behavior, which this card did not do.
- **The `edges` axis was not touched** — t448 (lane-8) owns it, per the dispatch's hard exclusion.
  `session_start.go` is unmodified in this branch.
- **No behavior change was made.** ⑴ and ⑵ are reported, not implemented.
- Coverage was not re-measured; a comment-only edit moves no statement count.

## Residual-risk

- If the runtime does NOT apply the matcher as a regex, ⑴'s premise reverts to lane-8's and the
  correct action is the one they framed. The report deliberately does not pick between the two.
- The corrected comment now cites a measured `0/10` from another card's tree. Should the process
  lifetime change (a hook daemon, a batched runner), the comment becomes stale in the opposite
  direction — it would then understate what completes.
- ⑵ left as-is means the `config_change` goroutine keeps running per event, paying a spawn plus a
  20 ms debounce for an effect nothing observes.

---

## What the lead must decide

**D1 — ⑴, blocked on a measurement this card did not own.** Before choosing between "widen the
matcher to the source extensions" and "retire handler + registration", someone measures how the
runtime applies the matcher. Widening on the literal reading and retiring on the regex reading are
opposite mistakes.

**D2 — ⑵, a landed-SPEC contract change.** Removing the async wrapping is right on the merits
(nothing survives the goroutine), but AC-HAE-003 of SPEC-V3R6-HOOK-ASYNC-EXPAND-001 asserts the
async return path. That is an operator call, not a lane call.
