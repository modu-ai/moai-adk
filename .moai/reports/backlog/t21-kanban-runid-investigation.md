# t21 — Kanban lead run-id vs session-name divergence

Read-only investigation. Repo `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99`.
No source file was edited.

---

## Claim

**C1 — Lifecycle.** The kanban run id is minted **per launcher process, at launch, from the
current Unix second**, published only into the **process environment**, and **never persisted
nor read back**. There is no state file, no settings key, and no registry field carrying it.

**C2 — Root cause of the observed divergence.** On the **lead** branch the launcher mints a
fresh id **unconditionally** and never reconciles it with an operator-supplied
`--name lead-<run-id>`; the SessionStart notice then composes companion launch commands from
the **environment** id, while the session's identity comes from the **`--name`** id. Nothing
joins the two. Directly evidenced on the very process this investigation runs inside
(PID 35136): launched at the second that encodes `tjr8zx`, named `lead-tjqim9`, env
`MOAI_KANBAN_ID=tjr8zx`.

**C3 — The card's stated measurement is partly wrong.** "All five live sessions carry the
`-tjqim9` suffix" does not hold. **Four** distinct run ids are live concurrently
(`tjqh1a`, `tjqim9`, `tjqk9a`, plus the env-only `tjr8zx`), across **two** live lead sessions
and **twelve** companions. The `-tjqim9` set is one of three companion sets, not the whole board.

**C4 — A second, independent divergence path exists** (lead restart / `--continue`): a
relaunched lead mints a *new* id with no memory of the old one, so companions launched from the
previous notice are orphaned even when name and env agree. Two live processes were launched
with `--continue` under a kanban settings file.

**C5 — Companions cannot diverge by construction.** `enterKanbanCompanionMode` derives the id
*from the label*, so a companion's env id and its name always agree. The defect is lead-only.

---

## Evidence

### E1 — The mint (verbatim source)

`internal/kanban/bootstrap.go:42-44`

```go
func NewRunID() string {
	return base36(time.Now().Unix())
}
```

Doc comment, same file, lines 30-41: *"Monotonic by construction … it needs no state file, no
counter, and no lock. The session registry is deliberately NOT consulted."*

### E2 — Publication: env only, unconditional, per launch

`internal/cli/kanban.go:89-106` (`enterKanbanMode`, the **lead** branch):

```go
	_ = os.Setenv(config.EnvMoaiKanban, "1")
	runID := kanban.NewRunID()
	_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	…
	_ = os.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-"+runID)
```

There is no read-back branch: the function takes only `specID` and never consults argv, a state
file, or a prior value. `captureEnvState` restores the *prior* value on return, which is a
process-teardown concern, not persistence.

Contrast — `internal/cli/kanban.go:158-170` (`enterKanbanCompanionMode`, the **companion** branch):

```go
	_ = os.Setenv(config.EnvMoaiKanbanLabel, label)
	if _, runID, ok := kanban.SplitCompanionLabel(label); ok {
		_ = os.Setenv(config.EnvMoaiKanbanID, runID)
	}
```

The companion **derives** the id from the label ("so the two can never disagree", per its doc
comment). The lead has no equivalent derivation. That asymmetry is the defect.

### E3 — The name injection does not close the gap (PR #1538)

`internal/cli/kanban.go:316-325`:

```go
func leadNameArgs(args []string) []string {
	if operatorSuppliedName(args) {
		return nil
	}
	runID := os.Getenv(config.EnvMoaiKanbanID)
	if runID == "" {
		return nil
	}
	return []string{nameFlagLong, kanban.LeadLabel(runID)}
}
```

When the operator supplies a name it returns `nil` — moai declines to *add* a name, but it also
does **not** adopt the run id embedded in the operator's name. Env keeps the fresh mint.
Confirmed by the shipped tests:

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
        MOAI_KANBAN_SETTINGS_INJECTED && \
  go test ./internal/cli/ -run 'TestLeadNameArgs|TestOperatorSuppliedName' -v
--- PASS: TestOperatorSuppliedName (0.00s)
    … 10 subtests PASS …
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.112s
```

`TestLeadNameArgs_NeverOverridesOperatorName` (kanban_lead_name_test.go:65-80) pins exactly the
behaviour above: operator name wins, run id untouched.

Commit provenance:

```
$ git log -1 --format="%H %ad %s" --date=iso 3b9b3bf99
3b9b3bf9959669c4bfc43da313e25bca61f910a2 2026-08-15 13:25:09 +0900 feat(kanban): name the lead session so it survives /clear (#1538)
```

The lead-name injection landed **2026-08-15 13:25**, i.e. *after* every live session in this
board was launched. The `--name lead-…` values observed below were therefore typed by the
operator, not injected.

### E4 — The notice reads the environment, and only the environment

`internal/hook/session_start_kanban.go:47`:

```go
	return kanbanLeadNotice(os.Getenv(config.EnvMoaiKanbanID), lang)
```

`session_start_kanban.go:117-120`:

```go
	for _, role := range kanban.CompanionRoles {
		launch = append(launch, "moai cc -k --name "+kanban.CompanionLabel(role, runID))
	}
```

So the printed commands are `plan|run|review|sync-<env id>` with no reference to the session's
own name. This reproduces the reported `moai cc -k --name plan-tjr8zx` exactly.

Emission is gated to `source == "startup"` (or empty) by `kanbanBootstrapNoticeForSource`
(lines 76-81), so `/clear`, resume, compact and fork do **not** re-announce — a genuine relaunch
does.

### E5 — Live process table (the decisive measurement)

```
$ ps -eo pid,lstart,command | grep -E "^ *[0-9]+ .*[c]laude " | grep -v node
35136 Fri Aug 14 19:26:21 2026  claude … -d --name lead-tjqim9  --settings /var/…/moai-kanban-35136-1786703181760280000.json
91135 Fri Aug 14 18:56:00 2026  claude …    --name lead-tjqk9a  --settings /var/…/moai-kanban-91135-1786701360982495000.json
96729 Fri Aug 14 20:16:44 2026  claude … --continue -d          --settings /var/…/moai-kanban-96716-1786706203949290000.json
37031 Fri Aug 14 09:22:22 2026  claude … --continue -d          --settings /var/…/moai-kanban-36921-1786666942755021000.json
28295 Fri Aug 14 20:18:05 2026  claude … -d                     --settings /var/…/moai-kanban-28295-1786706286022015000.json
  732 Sat Aug 15 13:06:52 2026  claude … --name plan-tjqim9
19962 Sat Aug 15 12:42:47 2026  claude … --name run-tjqim9
24492 Sat Aug 15 12:43:23 2026  claude … --name review-tjqim9
26158 Sat Aug 15 12:43:35 2026  claude … --name sync-tjqim9
27512 Sat Aug 15 12:43:45 2026  claude … --name plan-tjqim9
80099 Sat Aug 15 12:38:43 2026  claude … --name plan-tjqk9a
88949 Sat Aug 15 12:39:45 2026  claude … --name run-tjqk9a
91714 Sat Aug 15 12:40:06 2026  claude … --name review-tjqk9a
94929 Sat Aug 15 12:40:16 2026  claude … --name sync-tjqk9a
 4987 Sat Aug 15 12:41:19 2026  claude … --name run-tjqh1a
 9777 Sat Aug 15 12:41:49 2026  claude … --name review-tjqh1a
12260 Sat Aug 15 12:42:07 2026  claude … --name sync-tjqh1a
```

The transient settings filename carries `<pid>-<unix-nanos>`, which is the launch instant — an
independent clock for the mint.

### E6 — Decoding the ids (base36 seconds ⇄ wall clock)

```
$ python3 -c "…"      # base36 decode
tjqh1a 1786666942  2026-08-14 09:22:22
tjqim9 1786668993  2026-08-14 09:56:33
tjqk9a 1786671118  2026-08-14 10:31:58
tjr8zx 1786703181  2026-08-14 19:26:21

$ python3 -c "…"      # base36 re-encode of observed launch epochs
1786703181 -> tjr8zx     (PID 35136 settings nanotime)
1786701360 -> tjr7lc     (PID 91135 settings nanotime)
```

- **PID 35136** launched at the exact second encoding `tjr8zx`, yet is named `lead-tjqim9`
  (a mint from 09:56:33, ~9h30m earlier).
- **PID 91135** launched at the second encoding `tjr7lc`, yet is named `lead-tjqk9a`
  (a mint from 10:31:58).

Both live leads exhibit the same name-vs-env split. Neither is a one-off.

### E7 — This investigation's own environment

```
$ env | grep -i MOAI_KANBAN
MOAI_KANBAN=1
MOAI_KANBAN_ID=tjr8zx
MOAI_KANBAN_LEAD_ADDR=/tmp/moai-kanban-tjr8zx
MOAI_KANBAN_SETTINGS_INJECTED=1
```

This subagent runs inside PID 35136 — the session named `lead-tjqim9`. The reported symptom is
measured from inside the affected process.

### E8 — Nothing on disk carries a run id

```
$ grep -rn "tjr8zx\|tjqim9" .moai/state/ .claude/settings.local.json .moai/logs/
.moai/state/verify/4cd38714/pr1526-gotest-full.log:889:  … "Kanban Mode: joined run tjqim9."
.moai/state/kanban/backlog.json:77: … (the t21 card text itself)
```

Only a stale test log and the card. `.claude/settings.local.json`: no match.

`internal/kanban/record.go` `Record` fields — `session_id`, `spec_id`, `role`, `backend`,
`entered_at`, `deepscan_dir`, `verify_rung`, `verify_reentries`. **No run-id field.** Sample
record (`.moai/state/kanban/1358ceb0-….json`, written 2026-08-14 19:26 — the tjr8zx launch):

```json
{ "session_id": "1358ceb0-f8f0-45fd-b982-d6c68c0e2939", "spec_id": "",
  "backend": "claude", "entered_at": "2026-08-14T10:26:21Z",
  "deepscan_dir": "", "verify_reentries": 0 }
```

Note it also carries **no `role`** — `recordKanbanSession` was called before the role field was
populated for leads, so the records cannot even be joined back to a run by role.

```
$ moai session list --json
[ … 5 entries, fields: session_id, spec_id, phase, started_at, last_heartbeat, pid, host, cwd … ]
```

The registry carries **no run id and no session name**, so it cannot be used to reconcile either.

---

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99` (`git log -1` above).
- Source read at that HEAD: `internal/kanban/bootstrap.go`, `internal/cli/kanban.go`,
  `internal/cli/cc.go`, `internal/hook/session_start_kanban.go`, `internal/kanban/record.go`,
  `internal/cli/kanban_lead_name_test.go`.
- Runtime observations taken 2026-08-15 ~14:03 KST on host `goos.local` (`ps`, `env`,
  `moai session list --json`, `grep` over `.moai/state/`).
- Test run: `go test ./internal/cli/ -run 'TestLeadNameArgs|TestOperatorSuppliedName' -v`,
  observed `ok … 1.112s`, this tree, this run.
- Timestamp arithmetic: local `python3` base36 encode/decode, shown verbatim in E6.

---

## Root cause (stated plainly)

Two id-bearing surfaces exist for a lead session and nothing joins them:

| Surface | Value comes from | Survives restart? |
|---|---|---|
| Session identity (`--name lead-<id>`) | operator's argv, or `leadNameArgs` when absent | yes (claude keeps explicit names) |
| Run id (`MOAI_KANBAN_ID`, and everything the notice prints) | `NewRunID()` — a fresh mint every launch | no |

`resolveKanbanBranch` sends `-k --name lead-<id>` down the **lead** branch (`lead` is absent from
`CompanionRoles`, so `SplitCompanionLabel` rejects it by design — bootstrap.go:70-72), and the
lead branch mints unconditionally. The operator's name is preserved but its id is discarded.

The card's guess — *"a lead restart minted a fresh run-id while companions kept the old one"* —
is **half right**. A restart does mint a fresh id (there is no persistence to prevent it), and
that is real defect path #2. But the *observed* `tjr8zx` vs `tjqim9` split is not a restart
artifact: it is the operator-supplied `--name` carrying an id the launcher never reads.

---

## Reproduction

**R1 — Observed (already reproduced; no new run needed).** PID 35136 and PID 91135 are two live
instances, measured above.

**R2 — Minimal recipe (derived from source; NOT run — running it would `syscall.Exec` into a new
claude session):**

```
moai cc -k                      # note the run id X printed by the SessionStart notice; exit
moai cc -k --name lead-X        # relaunch, deliberately reusing X
```

Derived expectation: env `MOAI_KANBAN_ID` = a **new** id Y (fresh mint at relaunch), session
named `lead-X`, notice printing `moai cc -k --name plan-Y` … — the exact reported shape.
*Hypothesis until executed.*

**R3 — Restart path (derived, NOT run):** launch `moai cc -k` with no name, launch companions
from the notice, quit the lead, relaunch `moai cc -k`. Derived expectation: fresh id, new
companion commands, previously-launched companions orphaned. *Hypothesis until executed.*

**R4 — A runnable unit-level probe that needs no session** (derived, NOT run):
call `enterKanbanMode("")` then `leadNameArgs([]string{"--name","lead-abc123"})` and assert
`os.Getenv(MOAI_KANBAN_ID) != "abc123"`. This is the defect expressed as a failing test; the
existing `TestLeadNameArgs_NeverOverridesOperatorName` asserts the `nil` return but never
inspects the env, which is why the gap ships green.

---

## Gaps

- **R2/R3/R4 were not executed.** R2/R3 exec a real session; R4 would require writing a test
  file, which this investigation was told not to do. All three are labelled hypotheses.
- **Not established: whether claude honours a second `--name` on a resumed session.** The
  `--continue` processes (96729, 37031) carry no `--name` at all, so the interaction between
  `--continue` and an injected `lead-<id>` name post-#1538 is untested here.
- **Not established: which lead actually printed the `plan-tjr8zx` line.** The card reports it as
  directly observed; the code path (E4) accounts for it exactly, but no transcript was read.
- **The kanban records cannot be joined to runs at all** (no run id, and the older ones no role),
  so no on-disk history of past runs was reconstructable — only the live process table.
- **No check of the glm launcher path.** `internal/cli/glm.go:183` mirrors `cc.go:115`; the
  branch structure was read but glm-side behaviour was not separately measured.

---

## Residual risk

- The `ps` snapshot is one instant. A session started or died during the sweep would be
  mis-counted; the four-id conclusion rests on names visible at 14:03 KST.
- The settings-filename nanotime is treated as the launch instant. It is written by the launcher
  microseconds before `syscall.Exec`, so mint-second and file-second could in principle straddle
  a second boundary — irrelevant at the 9h30m / 8h24m deltas measured, but it would matter for a
  sub-second claim.
- `NewRunID` collides for two leads launched in the same second (documented and accepted at
  bootstrap.go:38-41). That is a separate, un-triggered residual, not this defect.
- Base36 decoding assumes `NewRunID`'s alphabet and unpadded form. Both were read from source;
  the round-trip in E6 (`1786703181 -> tjr8zx`) is the consistency check.

---

## Recommended fix (described, not applied)

**F1 — Adopt the id from an operator-supplied lead name (closes the observed defect).**
Make the lead branch symmetric with the companion branch. Introduce a `SplitLeadLabel` beside
`SplitCompanionLabel` (same shape check, `lead` prefix), and have `enterKanbanMode` take the
argv-derived label: when the operator's `--name` parses as `lead-<run-id-shape>`, **publish that
id** instead of minting. Mint only when no lead-shaped name is present. This makes the notice,
the socket path, and the session name agree by construction — the property the companion branch
already has and states in its own doc comment.

Cost: one new splitter, one parameter on `enterKanbanMode`, one call-site change in each of
`cc.go` / `glm.go`. Risk: an operator naming a lead `lead-notarunid` — guard with the existing
`isRunIDShape` and fall back to minting.

**F2 — Assert the env, not just the return value, in the lead-name tests.** The current suite
passes because it only checks that `leadNameArgs` returns `nil`. Add the R4 assertion so the
divergence cannot ship green again.

**F3 (optional, separate) — persist the active run id** so a restart can offer to rejoin rather
than silently orphan its companions. `Record` would gain a `run_id` field, and the launcher
would consult the newest live record for this project root. This addresses defect path #2 (C4),
not the observed symptom, and it reintroduces exactly the state file `NewRunID`'s doc comment
deliberately avoided — so it deserves its own decision rather than riding along with F1.

**F4 — Make the notice self-consistent regardless.** Have `kanbanLeadNotice` also print the
session's own name (or warn when `MOAI_KANBAN_ID` disagrees with a `lead-` name in argv), so the
failure is visible at the moment it happens instead of when a copied command opens an orphan.
