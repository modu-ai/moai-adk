# design.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

System design for the behaviours in `spec.md`. Factory mode only; no kanban path is added or touched. Package placement is indicative and confirmed in the run phase (`plan.md` §D).

## 1. Components and data flow

```
Claude Code ──stdin rate_limits──► statusline writer ──► per-session record (context-usage/<sid>.json, schema 3)
                                                              │ read-only
                                                              ▼
                                         aggregator (freshest per window, rollover, fail-open)
                                                              │ one pressure evaluation (gate config + windows)
                    ┌───────────────────────┬────────────────┴───────────────┬──────────────────────┐
                    ▼                       ▼                                ▼                      ▼
        lane gate (factory next,     status quota block            --auto recommendation    acquire warning
        CLI + MCP; Claude lane)      (text + JSON)                 (before each accept)     (stderr, warn-only)
                    │                       ▲                                ▲
                    │ hold line (stderr)    └──── lane inventory ────────────┘
                    ▼                             (registry rows: liveness + backend written at claim)
              exit status 3
```

Every arrow after the record is a read. The only writer in the whole design is the statusline writer (already writing this file today).

## 2. Record (REQ-QAS-001..004)

- Added optional fields, omitted when empty: per window (five-hour, seven-day) the used percentage, the reset time in epoch seconds (the stdin unit), and the first-observed-exhausted capture time. Schema version 3.
- Reader: a version-1 or version-2 record decodes with no windows; a window-less record serializes as before, apart from the version value.
- Throttle payload gains, per window: presence, integer-rounded percentage, reset time. A window dropping out of stdin is a payload change. A record that carries a window is also rewritten when its on-disk capture time is older than 5 minutes (compiled constant); a window-less record keeps today's rule exactly.
- Exhausted time: set from the capture time when a window is first seen at or above 100%, kept while the window's reset time is unchanged, dropped when the window is no longer carried or the reset time changes.

## 3. Aggregator and pressure (REQ-QAS-005..007, -017)

1. List the record directory; consider only files whose modification time is within `max_age` (cost does not grow with the count of dead-session records).
2. Parse each; per window, keep the record with the newest capture time that carries that window.
3. Per window state: **unknown** (no fresh record, or any read failure), **reset** (reset time not after now), **fresh** (otherwise) with percentage and reset time.
4. **Pressure** = the gate is enabled and some **fresh** window is at or above its own hold percentage. **Hold** = pressure and the caller is a Claude lane.
5. No network, no process, no write; every failure maps to unknown.

Configuration (`workflow.quota_gate`, REQ-QAS-008): `enabled` false, `five_hour_hold_pct` 90, `seven_day_hold_pct` 95, `release_margin_pct` 5, `max_age` 30m — **unmeasured defaults** (DO-3), all four tunable without code. Out-of-range or unparseable values and an unreadable file yield the default. Compiled: heartbeat 5m, exhaustion 100%.

## 3b. Claude-lane predicate (REQ-QAS-012)

A lane is a Claude lane when its launch provider (the `MOAI_LAUNCH_PROVIDER` variable, falling back to `MOAI_KANBAN_BACKEND`) names Claude. Glm, gpt, empty, and unrecognised values are not Claude lanes (fail open). The accepted token set is fixed in the M3 pre-flight from what a `moai cc` lane carries.

## 4. Lane gate (REQ-QAS-009..012)

Flow of `moai factory next` and the `factory_next` MCP tool for a Claude lane with the gate enabled:

1. Evaluate pressure once.
2. No pressure → the unchanged path.
3. Pressure → run the selection pass with "no new cards" set: arm (a) (card assigned to this lane) may lease; arms (b), (b2), (c) are bypassed (so the queue is not promoted).
4. A card leased by arm (a) → returned exactly as without the gate. Nothing leased → print the hold line on stderr, print nothing on stdout, exit 3 (CLI) or return the hold line as text (MCP). No factory record, queue, or lane state changes.
5. With `--wait`: re-check every 5s until the bound. Latch: once held in this invocation, remain held until the held window's percentage is below `hold − release_margin` or its reset time has passed; a card assigned to the lane meanwhile is leased at the next check. After the bound, exit 3 with the hold line.

`stage`, `complete`, `factoryNextSkipForBackend`, the Codex loop, and the Claude relaunch loop are untouched. The hold line: `quota hold: <window> used=<pct>% resets_at=<RFC 3339 UTC>`, one segment per held window separated by `; `.

## 5. Lane inventory (REQ-QAS-018)

For each registered lane label (registry read), keep lanes whose registered pid is alive. The row's backend decides: `claude` → excluded; `glm` or `gpt` → candidate; empty (every row written before the claim write) or unrecognised → unknown. An unreadable registry → no candidates. Read-only; no session record is read (DO-12).

## 5b. Backend recorded at claim (REQ-QAS-023)

The claim engine gains a backend input and writes it in the row insert, inside the transaction that already selects the number and removes dead claims. Each launcher passes the value it already holds at its claim site: Claude `claude`, GLM `glm`, the Codex loop `gpt`, the Codex factory entry `codex` normalized to `gpt`. The existing exported claim entry points stay and pass the empty value, so the kanban companion and leader registries (separate files) and the existing tests are unaffected. The registry entry type and its reader carry the column; the registry rewrite round-trips it. The pid and heartbeat updates that follow (Codex re-stamp, card heartbeats) never touch it. No schema change, no migration: the column has existed since the table was created, and old rows read empty.

## 6. Steering output (REQ-QAS-019..022)

When pressure is on:

- Recommendation: `quota pressure: <window> used=<pct>% resets_at=<…Z>; recommend non-Claude lane(s): lane-2 (glm), lane-3 (gpt)`.
- No candidate: `quota pressure: <window> used=<pct>% resets_at=<…Z>; warning: no live non-Claude lane (unknown backend: <n>); nothing is re-dispatched`.
- `moai todo --auto`: the line immediately before each `accept <id> …` line, re-evaluated per card; the cycle's own queue mutations (its pick, its done or unpick) are exactly those of a pressure-off run; the directive still names an in-session worker.
- `moai factory status`: the same line under the quota block; JSON `quota` carries the per-window readings, `pressure`, `recommend` (label and backend per candidate), `warning` (`no-non-claude-lane` or absent), and `unknown_lanes`; the key is omitted when no window has data.
- Pressure off: byte-identical output to the pre-change output of the same state.
- Never: any card assignment, reassignment, lease, unpick, queue change, lane start or stop, message, process, or network call; the card's class is not read.

## 7. Integration-window warning (REQ-QAS-014)

At `moai integration acquire`, after the window is recorded, a Claude lane under pressure gets one stderr line naming the window and its reset time. The record, the exit status, and stdout (including `--json`) are unchanged; acquiring is never blocked, refused, or delayed (DO-7 final).

## 8. Failure modes

| Failure | Behaviour |
|---|---|
| No `rate_limits` on stdin (API-key user, GLM, Codex, before first response) | no window carried; windows unknown; never holds |
| Record directory missing or unreadable | unknown; fail open |
| A record unparseable | skipped; fail open |
| Only stale records | unknown |
| Reset time not after now | reset; never read as a stale high |
| Config unreadable or out of range | defaults |
| Registry unreadable | no candidates; the warning prints with unknown count when pressure is on |
| Registry row from an older binary (empty backend) | unknown; never a candidate |
| Backend token unrecognised | not a Claude lane; never held |
| Selection pass lost a lease race | existing retry behaviour unchanged |

## 9. Compatibility and ordering

- Baselines (M0, own commit): window-less record bytes, `moai factory status --json`, `moai todo --auto` output of a fixed fixture; every pressure-off AC compares against them.
- Template-First: template `workflow.yaml` (off) and local `workflow.yaml` (on) in the same change; config cache schema 10 → 11; shipped-key inventory rows.
- Kanban: the only file under `internal/kanban/` edited is the factory registry's claim cluster `factory_slots.go` (and its test); `internal/cli/kanban.go`, `kanban_settings.go`, and the kanban companion and leader registries are untouched (AC-QAS-016). The inventory reads the registry through `LoadFactoryRegistry` and `FactoryProcessAlive` only.
