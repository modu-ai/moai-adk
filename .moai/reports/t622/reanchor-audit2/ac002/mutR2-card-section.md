### Pre-Spawn Sync Check (Multi-Session Race Mitigation)

[ZONE:Evolvable] [HARD] Before spawning any implementation `Agent()` (manager-develop / manager-docs / per-spawn `Agent(general-purpose)` with a domain whitelist) that will commit or modify shared working-tree files, the orchestrator MUST execute the following two-lane batch and surface any divergence to the user.

The lanes have one deliberate dependency boundary:

* **Lane A (ordered):** `git fetch origin main` MUST finish and its exit status be observed before `git rev-list --count --left-right origin/main...HEAD` starts. The divergence count is only attributable to the ref that the completed fetch installed; never run these two commands as one parallel batch or as a shell line whose completion cannot be distinguished.
* **Lane B (independent):** `moai session list --json --filter-spec=<SPEC-ID>` may run concurrently with Lane A's fetch because it does not read `origin/main`. Its result is joined with Lane A after both lanes complete.

```bash
# Lane A — ordered; wait for fetch completion before reading origin/main.
git fetch origin main 2>&1
fetch_status=$?
if [ "$fetch_status" -ne 0 ]; then
  printf 'pre-spawn sync blocked: fetch origin/main failed (status=%s)\n' "$fetch_status" >&2
  exit "$fetch_status"
fi
git rev-list --count --left-right origin/main...HEAD

# Lane B — can be started while Lane A is fetching, then joined before the
# divergence/session decision is surfaced (L1 of the canonical 4-layer policy).
moai session list --json --filter-spec=<SPEC-ID>
```

The orchestrator MUST retain the fetch completion status (and, where the
runner exposes it, the fetched-ref timestamp) beside the divergence output.
This prevents a delayed or failed fetch from being mistaken for a current
`origin/main` baseline.

Interpretation matrix (git divergence):

| Output | Meaning | Action |
|--------|---------|--------|
| `0 N` | Local ahead by N (clean — your commits not yet pushed) | Proceed normally |
| `0 0` | Synced (local == origin/main) | Proceed normally |
| `N 0` | Origin ahead by N | Proceed normally |
| `N M` | Diverged (both ahead) | STOP, MUST resolve before spawn |

Interpretation matrix (active-sessions query):

| Output | Meaning | Action |
|--------|---------|--------|
| `[]` | No other session on this SPEC | Proceed normally |
| `[{...}]` (≥1 entry from another session) | **Concurrent session race detected on same SPEC** | STOP, surface entries, AskUserQuestion: **wait** / **override** / **abort** |

The 3rd command is additive only (the original 2-command batch is preserved verbatim). Sessions predating the registry hook emit no entries — `[]` — no false positives. Rationale + the originating sync-phase race incident record: `agent-common-protocol-reference.md` § Pre-Spawn Sync Check rationale and incident record.

Exemption: read-only agents (`Explore`, or a per-spawn `Agent(general-purpose)` scoped to read-only investigation) do not require pre-spawn fetch — they cannot trigger race conflicts.

> **Spawn-gate boundary**: this check fires only at the write-agent spawn boundary. Direct main-session edits bypass this gate — see § Pre-Edit Sync Check below. Defense-in-depth policy: `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation Procedure; worktree-as-race-elimination: `session-handoff.md` § Worktree-Anchored Resume Pattern.

### Pre-Edit Sync Check (Direct-Edit Race Mitigation)
