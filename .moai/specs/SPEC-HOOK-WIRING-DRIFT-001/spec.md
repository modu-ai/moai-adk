---
id: SPEC-HOOK-WIRING-DRIFT-001
title: "Hook wiring drift — close the local drift, make it detectable, record the dispositions, stop the MX dead work"
version: "0.3.0"
status: completed
created: 2026-08-24
updated: 2026-09-03
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/settings.json, internal/cli/doctor.go, internal/cli/mx_query.go, internal/hook/session_start.go, .claude/rules/moai/development/hook-independence.md"
lifecycle: spec-anchored
tags: "hooks, settings, wiring-drift, doctor, diagnostics, mx-index, dead-work, card-t216"
tier: M
---

# SPEC-HOOK-WIRING-DRIFT-001 — Hook wiring drift + dead-work cleanup

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.3.0 | 2026-08-24 | manager-spec | Audit iteration 2 amendment. plan-audit returned **PASS 0.862** (Clarity 0.80 / Completeness 0.92 / Testability 0.80 / Traceability 0.95), all 7 must-pass PASS, no iteration-1 defect unresolved. Two process facts recorded first. **(1) The iteration-1 threshold was misapplied**: 0.85 is the Tier L value, and this SPEC is `tier: M`, whose SSOT threshold (`spec-workflow.md:138-142`) is **0.80** — so v0.1.0's 0.807 was a PASS, and the FAIL that produced v0.2.0 rested on the wrong number. The findings were real and the revision improved the document; the verdict was not. **(2) PASS did not make the remaining findings optional** — three blocking-class findings were absorbed into the Testability score rather than the must-pass firewall, and all are fixed here. **N1 (critical, and a self-inflicted regression from the D7 fix)**: widening AC-HWD-007 to close mutant M-3's false-pass hole made the criterion **unsatisfiable** — `moai doctor` writes `.moai/state/config-cache.json` on first run in a project, so snapshot (iii) changes against the freshly-built fixture no matter how the check is implemented. Reproduced independently here. Fixed by warming the fixture before snapshotting and naming the excluded path with its reason. **N2**: REQ-HWD-012 named six surfaces while AC-HWD-018 observed four, leaving `hook-independence.md`'s five "dormant" occurrences unbound — and AC-018's fresh-mutant note claimed AC-009/AC-010 closed that gap, which is false. Given extended to six; the false closure retracted; a line-specific caution added, since three of those five occurrences are correct and must survive. **N3**: AC-HWD-009's "on a line carrying one of the five classes" was defeated by a blanket line carrying all 11 names. Rewritten to bind one sole-name row per script **carrying that script's expected class** — a strengthening beyond the audit's prescription, because the prescribed wording alone is defeated by a second mutant constructed here. **N4/N5**: two measured-evidence corrections under the [HARD] §C-5 constraint. **N6/N7**: two rewordings. Both mutants for N3 and the N1 unsatisfiability were **constructed and executed**, not reasoned. |
| 0.2.0 | 2026-08-24 | manager-spec | Audit iteration 1 amendment. plan-audit returned **FAIL 0.807** (Clarity 0.75 / Completeness 0.90 / Testability 0.75 / Traceability 0.85; threshold 0.85) with all 7 must-pass criteria PASS — score-driven, not gate-driven. The auditor independently re-ran nine `Pre-impl observed:` values and all matched exactly; what did not hold was **mutation adequacy**. It constructed four passing mutants, including one against the criterion v0.1.0 declared mutant-proof. Fixes: **M-1** (relocate-and-rename defeated AC-HWD-013's three single-file greps) forced package-scoped greps, the `spawnDeferredAdvisoryScans` arity check promoted from prose into the Then clause, and a behavioural no-index-written assertion; **M-2** (a hardcoded expected-entry list passed AC-HWD-005 + AC-HWD-006 without ever rendering the template) forced the new mechanism-binding AC-HWD-017 and an injectable template source; **M-3** (a diagnostic writing `.moai/logs/doctor-drift.log` passed AC-HWD-007 while violating REQ-HWD-004's "any other file on disk") forced REQ-HWD-004 to narrow to the project root and AC-HWD-007 to widen to `git status --porcelain` + gitignored-path snapshots; **M-4** (`git rm --cached`, or truncation to 0 bytes, passed AC-HWD-011's `ls` inventory) forced git-state assertions and retracted a false "none constructible" claim. Also: the three deferred baselines were measured (`0`, `0`, exit `0`) restoring §C-5; the M3 correction was extended to the two remaining "dormant" surfaces including an always-loaded one; the two constraint-only criteria were promoted to REQ-HWD-013/014; and AC-HWD-001/002 were retired into AC-HWD-003 to stay at the Tier M 16-criterion ceiling. |
| 0.1.0 | 2026-08-24 | manager-spec | Initial draft, Tier M / Class C. Authored from the three plan-phase investigation reports (`.moai/reports/t216/d1-chain-event.md`, `d2-unwired-scripts.md`, `d3-mx-cold-start.md`), NOT from card t216 — see §A.2 for the five card premises the measurement disproved and how the scope was reframed as a result. |

---

## §A Context

### §A.1 What this SPEC is

Card t216 was opened on a hypothesis: a hook script (`chain-event.sh`) is wired
in the distributed template but not in this project, the local `settings.json`
is a stale rendered output, and fixing the wiring restores the chain ledger's
completion edge. A read-only investigation was run first, on three axes, and
persisted in this tree:

| Report | Axis | Path |
|---|---|---|
| D-1 | the `chain-event.sh` wiring drift and the `moai update` path | `.moai/reports/t216/d1-chain-event.md` |
| D-2 | the 11 present-but-unwired hook scripts | `.moai/reports/t216/d2-unwired-scripts.md` |
| D-3 | the MX cold-start scan | `.moai/reports/t216/d3-mx-cold-start.md` |
| audit | plan-audit iteration 1, whose four constructed mutants drove v0.2.0 | `.moai/reports/t216/plan-audit.md` |

Those three reports are the source of every claim below; each carries the
file:line citations. This SPEC delivers four milestones drawn from what was
measured.

### §A.2 [HARD] The card's premises that measurement disproved

This section is itself a deliverable. Five card premises did not survive
measurement; carrying any of them into the body would put a disproved claim in
the requirement set.

| # | Card premise | What was measured | Where |
|---|---|---|---|
| 1 | `.claude/settings.json` is **a stale rendered output** | It is a **git-tracked, hand-maintained, committed file** — clean in both this worktree and the primary checkout, with its own commit history (`34a94cc80`, `e7acc30a7`, `3a2dfbaf0`, …). The local symptom is therefore fixed by **a commit**, not by any update-path change | d1 §E2 |
| 2 | There is **one** drift | There are **two**. (a) `chain-event.sh` missing from `SubagentStop` (template `settings.json.tmpl:198`, added `435bc2bbd` 2026-08-13). (b) `status-transition-ownership.sh`: the template has **three** `if`-scoped `async: true` entries, the project has **one** unscoped **synchronous** entry. Drift (b) is itself two divergences folded together — the `if`-scoping split (`796181198`, 2026-08-01) and the `async: true` addition (`ea4c6736f`, 2026-08-24, **card t214, merged the same day as this card was opened**) | d1 §E1 |
| 3 | Wiring `chain-event.sh` **records the completion edge** | **A no-op.** `CreateNodeAtSpawn` (`internal/chain/populate.go:53`) — the only writer of `node-enter` events — has **no production caller**, and nothing anywhere *sets* `MOAI_CHAIN_NODE_ID` (every reference reads it). `.moai/state/chain/events.jsonl` **has never existed** in any tree. With the hook wired, `ResolveCurrentNode` finds no node and `chain_event.go:86` logs `no matching chain node (non-blocking)` and returns nil | d1 §E3 |
| 4 | The audit's **34 hook entries** | **33 hook entries** across 20 events. `grep -c '"type": "command"'` returns 34 because the **`statusLine` block is also `"type": "command"`** and is not a hook | d2 § entry-count correction |
| 5 | The MX cold-start scan is dead work **every session** | Overstated. The scan is gated behind `mxIndexNeedsRebuild` (`session_start.go:1499`), false whenever an `mx-index.json` younger than 7 days exists — so in the primary checkout it **is not spawned at all**. 2 of 153 worktrees hold an index. The scan's cost is a goroutine killed within single-digit milliseconds; the loss is **the capability, not the cycles** | d3 §§1, 5 |

Two further corrections belong here:

- **`handle-agent-hook.sh` is NOT unwired.** It is registered **10 times via
  agent frontmatter `hooks:` blocks** (5 local, 5 template-mirrored) — a real
  Claude Code invocation surface the settings-only view cannot see. The prior
  audit classified those agent-`.md` hits as prose (d2 § per-script table).
- **`handle-session-start-navigator.sh` is reclassified** from "unreachable by
  design" to an **apparent template regression**: commit `7171880a9` removed it
  and `handle-codex-review-gate.sh` in one diff; the latter was restored two
  commits later by `e8e46396f`, the navigator entry was not (d2 § why it is
  ambiguous).

### §A.3 The structural finding that shapes M2

A hook entry added to the template after a project was initialized **can never
reach that project via `moai update`**. The deciding line is
`internal/merge/strategies.go:427-429`: `pruneToShared`
(`internal/cli/update/merge/base.go:118-128`) recurses **only through maps**, and
a hook entry is an **array element**, so the derived merge base equals the
template at the `hooks` key, the merge concludes only the user changed `hooks`,
and the user's entire `hooks` block is preserved wholesale (d1 §E4).

The asymmetry that makes this silent: `moai update` wipes `.claude/hooks/moai`
wholesale before redeploying, so **scripts are force-synced while registrations
are never synced** — `chain-event.sh` is present on disk in this project and has
never been registered. And there is no detection surface: `moai doctor`'s
`checkHooksConfig` (`internal/cli/doctor.go:678-692`) only `os.Stat`s the
directory; all 20 tests in `internal/template/settings_test.go` assert against
the template **only**; and the wrapper's `hook-missing.log` fallback catches the
inverse case (a registered entry whose script is absent), never this one
(d1 §E5).

**How would anyone notice? They would not.** That is what M2 exists to change.

### §A.4 What the MX consumer analysis found

The cold-start scan's output has exactly **one** reader in the tree —
`moai mx query` (`internal/cli/mx_query.go:95-104`) — and that reader already
fails loud with a self-service remedy (`run 'moai mx scan'`). The consumer is
live (`.claude/skills/moai/workflows/review.md:289` invokes it in the `/moai
review` lean audit), so this is not dead code; it is a **convenience**. The
capability at stake is narrow: *"`moai mx query` works in a fresh worktree
without the operator typing `moai mx scan`."* No automation anywhere runs that
remedy (d3 §5).

Raising the scan's time budget would change nothing: the 2 s box is never
reached (`ScanDir` measures ~260-350 ms at load 67 against a 0.05 s binary
floor). The scan dies because `spawnDeferredAdvisoryScans`
(`session_start.go:568-573`) dispatches it **after** the channel send, `Handle`
returns on receipt, and the CLI process exits, tearing down the goroutine
(d3 §3). M4 therefore moves the capability to the consumer and removes the
hook-side scan.

---

## §B Requirements (GEARS)

### M1 — close the local drift

- **REQ-HWD-001** (Ubiquitous) — The project's `.claude/settings.json` hook-entry
  set shall equal the rendered template's hook-entry set, compared as a set keyed
  on `(event, matcher, script, if, timeout, async)`, in both directions.

- **REQ-HWD-002** (Ubiquitous) — The SPEC record shall state that the
  `chain-event.sh` entry delivered by REQ-HWD-001 is **inert** — it produces no
  ledger event, because no chain node exists for a completion edge to attach to
  — and shall not claim that the entry makes the chain ledger work.

### M2 — make the drift detectable

- **REQ-HWD-003** (Event-driven) — When an operator runs `moai doctor`, the
  hook-wiring diagnostic shall render the settings template in memory, compare
  its hook entries against the project's `.claude/settings.json`, and report
  drift in **both** directions, naming the affected script for each divergent
  entry.

- **REQ-HWD-004** (Unwanted) — The hook-wiring diagnostic shall not create,
  modify, or delete any file **under the project root**: after it runs, the
  worktree's `git status --porcelain` output and the name/size/mtime listing of
  the gitignored `.moai/logs/` and `.moai/state/` paths shall both be unchanged.

  > **Why "under the project root" and not "on disk"** (audit D7 / mutant M-3):
  > v0.1.0 said "any other file on disk" while its criterion observed only
  > `.claude/`, so a diagnostic writing `.moai/logs/doctor-drift.log` — a
  > realistic implementation choice, not a contrived one — satisfied every
  > assertion while violating the requirement as written. The requirement is now
  > scoped to what a criterion can actually observe, and the residual gap it
  > leaves (a write outside the project root, e.g. under `~/.moai/`) is recorded
  > in acceptance.md §D.6 rather than papered over by an unobservable clause.
  >
  > Why this is a requirement and not a description of one: a diagnostic that
  > silently repaired the drift would become the same class of defect this card
  > exists to close — a change nobody asked for and nobody can see. The
  > repair-vs-report question is settled in §C-2 and is not the diagnostic's to
  > reopen at runtime.

- **REQ-HWD-005** (Event-driven) — When the settings template cannot be
  rendered, or the project's `.claude/settings.json` is absent or unparseable,
  the diagnostic shall report a warn status naming the cause and the `moai
  doctor` run shall complete with its baseline exit status.

### M3 — record the disposition of the 11

- **REQ-HWD-006** (Ubiquitous) — The hook-wiring rule surface shall carry a
  disposition for each of the 11 present-but-unwired scripts, drawn from exactly
  five classes: `reachable-via-template-settings`,
  `reachable-via-agent-frontmatter`, `reachable-via-in-binary-registry`,
  `dead-by-decision`, `open-question`.

- **REQ-HWD-007** (Ubiquitous) — The rule surface shall state that
  `.claude/settings.json` carries **33 hook entries** across 20 events, that the
  34th `"type": "command"` occurrence is `statusLine` and is not a hook, and that
  `handle-agent-hook.sh` is registered via **agent frontmatter**, not via
  settings.

- **REQ-HWD-008** (Unwanted) — This SPEC shall not delete any hook wrapper
  script, in the project or in the template. All 11 have template twins, so
  removal would be a distributed act (d2 § template twins).

### M4 — stop the MX dead work

- **REQ-HWD-009** (Event-driven) — When `moai mx query` detects that the
  sidecar index is absent, empty, corrupt, or older than the freshness threshold,
  it shall build the index in-process and then serve the query.

- **REQ-HWD-010** (Ubiquitous) — The SessionStart handler shall carry no
  MX cold-start scan: the scan function, its freshness gate call site, and the
  parameter threading it through the deferred-advisory path shall be absent.

- **REQ-HWD-011** (Unwanted) — No comment in `internal/hook/session_start.go`
  shall assert that a fire-and-forget goroutine spawned in the hook process
  survives that process's exit.

### M3 (continued) — cross-surface consistency and the two [HARD] constraints

- **REQ-HWD-012** (Ubiquitous) — Every rule surface that describes
  `team-ac-verify.sh`'s registration state shall carry the same corrected
  reading: that it is not registered in any settings surface, so no configuration
  flag activates it. The surfaces are `hook-independence.md`,
  `agent-common-protocol.md` (always-loaded),
  `agent-common-protocol-reference.md`, and their template twins.

  > Added in audit iteration 1 (D6). Correcting one surface while two others —
  > one of them always-loaded — keep the "dormant" wording leaves three surfaces
  > disagreeing, and an unscheduled contradiction is worse than the original
  > error. This is a **factual correction, not a decision**: whether team mode
  > should fire the hook remains §G-3.

- **REQ-HWD-013** (Ubiquitous) — Every template-managed file this SPEC edits
  shall be edited at its `internal/template/templates/` source first and mirrored
  afterwards, and the two copies shall be byte-identical at closure.

- **REQ-HWD-014** (Unwanted) — No file this SPEC writes under
  `internal/template/templates/` shall contain a SPEC ID, a REQ token, an
  internal card number, an internal date, or a commit SHA.

  > REQ-HWD-013 and REQ-HWD-014 promote §C-1 and §C-4 into the requirement layer
  > (audit D12). They were already [HARD] constraints with criteria attached
  > (AC-HWD-015, AC-HWD-016); promoting them removes the two requirement-orphaned
  > criteria the auditor deducted Traceability for, and changes no obligation.

---

## §C Constraints

1. **[HARD] Template-First.** Files under `.claude/`, `.moai/`, `.agency/` that
   are template-managed are edited at `internal/template/templates/<path>`
   **first**, then `make build`, then the local mirror. Per-file classification
   is in §D.

2. **[HARD] Report, never repair — M2 is a diagnostic only.** The structurally
   correct fix (snapshot the deployed template at deploy time so a genuine merge
   base exists — d1 §E6 option B) is **deferred**, see §G-5. The tempting
   one-liner (extend `pruneToShared` to arrays — option A) is **rejected**: with
   no snapshot base it cannot distinguish "the template gained this entry" from
   "the user deliberately deleted it", so it would silently reinstate deliberate
   removals on every update, with no diagnostic and no opt-out. That failure is
   worse than today's silence because it is *active*.

3. **[HARD] No requirement may be phrased as "run `moai update` to fix it".**
   `moai update` wipes `.claude/hooks/moai` wholesale (`CleanMoaiManagedPaths`,
   `internal/cli/update/deploy/deploy.go`) and, per §A.3, cannot deliver a
   missing hook entry anyway. Should any acceptance step nonetheless invoke
   `moai update`, it must carry the CLAUDE.local.md §2.3 post-update deletion
   check (`git status --porcelain | grep '^ D'` → must be empty).

4. **[HARD] Template neutrality.** The M3 rule-surface edit lands in
   `internal/template/templates/**` and must therefore be free of SPEC IDs, REQ
   tokens, internal card numbers (`t242`/`t243`/`t244`), internal dates, and
   commit SHAs. The `open-question` disposition rows in the **template** name the
   pending decision without naming the card; card numbers appear in this SPEC's
   own artifacts (§G, frontmatter `tags`, plan.md, progress.md) and **never in
   `internal/template/templates/**`**.

5. **Every criterion is falsifiable and pre-measured.** Each entry in
   `acceptance.md` carries a `Pre-impl observed:` value measured on this tree —
   tagged `@950cb4399` where unchanged since authoring, `@4842760a7` where
   re-measured in audit iteration 1 — plus a mutant note. A criterion whose
   command already passes before implementation observes nothing and is rejected.

   > **No deferral class exists, and none is needed.** v0.1.0 deferred three
   > baselines to run time in violation of this clause (audit D4). All three are
   > now measured: AC-HWD-010's `statusLine` grep → `0`; AC-HWD-016's neutrality
   > grep → `0`; AC-HWD-008's doctor baseline → exit `0`, obtained by running the
   > `make build` that §C-6 and plan.md §C already required. The clause stays
   > absolute.

6. **Verification is scoped to the change.** Affected packages only
   (`internal/cli`, `internal/hook`, `internal/template`), then push and read CI
   for the full-suite verdict. `internal/cli` needs a `-timeout` floor of 600s.

---

## §D File classification — template-managed vs local-only

| File | Class | Edit path |
|---|---|---|
| `.claude/settings.json` | **local, git-tracked** (its template counterpart `settings.json.tmpl` is already correct and is **not** edited by this SPEC) | edit + commit the local file directly |
| `internal/template/templates/.claude/rules/moai/development/hook-independence.md` | **template source** | edit here **first** |
| `.claude/rules/moai/development/hook-independence.md` | **template-managed mirror** (currently byte-identical to its twin; lives inside a `CleanMoaiManagedPaths` root, so an unmirrored local edit is transient) | mirror after `make build` |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | **template source** (REQ-HWD-012; the mirror is **always-loaded**, which is why this surface cannot be left uncorrected) | edit here **first** |
| `.claude/rules/moai/core/agent-common-protocol.md` | **template-managed mirror**, always-loaded | mirror after `make build` |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md` | **template source** (REQ-HWD-012) | edit here **first** |
| `.claude/rules/moai/core/agent-common-protocol-reference.md` | **template-managed mirror** | mirror after `make build` |
| `internal/cli/doctor.go` | Go source | direct |
| `internal/cli/mx_query.go` | Go source | direct |
| `internal/hook/session_start.go` | Go source | direct |
| new `*_test.go` files | Go source | direct |

---

## §E Non-functional constraints

- The M2 diagnostic runs inside `moai doctor`'s existing check list; it performs
  one in-memory template render plus one file read, and must not add a
  user-perceptible delay to the doctor run.
- The M4 auto-build path pays the scan cost (~260-350 ms measured) **on the
  consumer's own invocation**, where the wait is expected and attributable —
  never on the SessionStart path. Cold-session SessionStart latency must not
  increase.
- Both M2 and M4 fail open: neither may turn a recoverable condition into a
  non-zero exit that did not exist before.

---

## §F Exclusions

### Out of Scope — the chain ledger itself

- Calling `CreateNodeAtSpawn`, setting `MOAI_CHAIN_NODE_ID`, adding a `moai chain
  enter` subcommand, or otherwise making `events.jsonl` come into existence.
- Any claim, in code comment or documentation, that this SPEC makes chain
  completion edges record.

### Out of Scope — the `moai update` merge path

- Extending `pruneToShared` to arrays; adding an array-aware merge strategy;
  snapshotting the deployed template at deploy time; a `moai update --hooks`
  partial re-render; a full re-render of `settings.json`.
- Any change to `internal/merge/strategies.go` or
  `internal/cli/update/merge/base.go`.

### Out of Scope — deletions

- Deleting, retiring, or unregistering any hook wrapper script, in the project or
  the template.
- Retiring `navigator_hook_test.go` or `hook_official_compliance_test.go`, or
  changing the acceptance criteria they enforce.

### Out of Scope — the three owner decisions and the fourth defect

- Whether `handle-session-start-navigator.sh` should be restored or retired.
- Whether team mode is meant to fire `team-ac-verify.sh`.
- Whether SPEC-CHAIN-CORE-001 Phase 1 intended to leave spawn-side population
  unimplemented.
- Fixing the identical die-at-exit bug on the incremental index path
  (`internal/hook/file_changed.go:110-118`).

Each is carried, with its evidence, in §G.

### Out of Scope — MX index staleness semantics

- Changing `mxIndexFreshnessThreshold`, or addressing the untested
  stale-but-fresh-enough case (a 6-day-old index is served as authoritative; the
  last measurement recorded 764 missing tags on a fresh worktree).
- Raising `mxIndexScanTimeout`. Explicitly ruled out: the 2 s box is never
  reached, so the change would deliver nothing (d3 §6 option F).

---

## §G Named follow-ups

**G-1 — chain node-population gap → card t242.** `CreateNodeAtSpawn`
(`internal/chain/populate.go:53`) has no production caller; nothing sets
`MOAI_CHAIN_NODE_ID` (`internal/config/envkeys.go:238` is read-only at every one
of its 4 call sites); `internal/cli/chain.go` exposes only `status`, `lineage`,
`back`, `list`, `prune`. `.moai/state/chain/events.jsonl` has never existed.
`internal/chain/node.go:27` documents `EventCompletionEdge` as produced by the
SubagentStop hook. **SPEC-CHAIN-CORE-001's owner decides** whether Phase 1
intended this. That SPEC carries `status: completed`, so the decision lands as an
**amendment** to it (`completed → in-progress` per the frontmatter schema),
carried by card t242 — not as work inside a live phase. Until then the entry M1
delivers is parity, not function.

**G-2 — `handle-session-start-navigator.sh` regression → card t243.** Added to
the template only, by `2c87d195f` (SessionStart, timeout 5); removed by
`7171880a9` in a diff whose commit body describes Go-side build recovery and
never mentions a hook, alongside `handle-codex-review-gate.sh`, which was
restored two commits later by `e8e46396f`. `internal/template/navigator_hook_test.go`
still executes the script directly and asserts AC-PN-009/010/012 — green
criteria for a delivery mechanism that is gone. Note also that
`.moai/project/navigator/navigator.md` does not exist in this tree, so a restored
wiring would fail open at the script's first `[ ! -r "$NAV_FILE" ]` guard.
**SPEC-PROJECT-NAVIGATOR-001's owner decides** restore-or-retire. That SPEC is
also `status: completed`, so the same amendment path applies, carried by card
t243.

**G-3 — `team-ac-verify.sh` wiring → card t244.** It has **never** appeared in
any settings surface in the entire history (pickaxe empty across
`.claude/settings.json`, `settings.local.json`, the `.tmpl`, and the three
home-directory settings paths). `TaskCompleted` is wired to
`handle-task-completed.sh`, which does not call it. So flipping `team.enabled:
true` would not activate it — the script's runtime self-gate is unreachable
because the invocation is. Meanwhile `hook_official_compliance_test.go:62,73`
keeps its reject-shape contract green for a path nothing can enter, and
`agent-common-protocol-reference.md:291` describes it as "dormant", which reads
as *registered but gated* and is not what the wiring says. **Decide whether team
mode is meant to fire it**; if yes it needs a `TaskCompleted` entry in the
template (the script already self-gates, so an unconditional entry is safe).

**G-4 — the die-at-exit twin on the incremental index path (no card yet).**
`internal/hook/file_changed.go:110-118` spawns the MX sidecar `UpdateFile` in a
`context.Background()` goroutine with a 5 s deadline **inside the same
short-lived CLI process**, and dies the same way M4's scan does. Found in D-3,
**not flagged by the hook audit**. Asserted from code, not measured — no
FileChanged event was fired. Fixing it fixes index *maintenance*, not cold
start: `UpdateFile` merges one file at a time and never produces the initial
451-file set.

**G-5 — deploy-time template snapshot (deferred by §C-2).** The structurally
correct fix for §A.3: write the rendered `settings.json` to `.moai/state/` at
deploy time so the next update has a genuine merge base. It is the only option
that gets the hard case right by construction — a deliberately removed entry
lands on `inBase && !inCurrent && inUpdated` → "respect user deletion"
(`strategies.go:409-411`). Deferred here because it needs a new on-disk artifact,
a migration for projects with no snapshot (they fall back to today's derived
base), and **array-aware merge is still needed on top**. `base.go`'s own header
already records the gap. Sequencing: M2 first (cheap, safe, makes the class
visible and lets the affected population be measured), then this.

**G-6 — record-integrity note.** The prior audit reports
(`.moai/reports/hook-audit/*.md`) live in the primary checkout, are untracked,
and are **point-in-time records**. The two corrections in §A.2 (33-not-34;
`handle-agent-hook.sh` reachable) are recorded **here and in the durable rule
surface (M3)** — the audit reports are not rewritten, because editing a
landed-at-the-time record to match current understanding makes the record false.

---

## §H Cross-references

- `.moai/reports/t216/d1-chain-event.md` — wiring drift, the `moai update` merge
  path, the five fix options and their behaviour on a deliberately-removed entry
- `.moai/reports/t216/d2-unwired-scripts.md` — the per-script disposition table,
  the search space covered for every `dead` verdict, the template-twin matrix
- `.moai/reports/t216/d3-mx-cold-start.md` — the cold-start mechanism, the
  consumer analysis, the six options and why F is ruled out
- `CLAUDE.local.md` §2.3 — `moai update` deletes local-only files inside managed
  roots; the post-update deletion check
- `.claude/rules/moai/development/hook-independence.md` — the M3 target surface
