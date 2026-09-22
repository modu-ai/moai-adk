# t242 — Verdict: chain node-creation gap

> Card t242 (G3 gate batch, from the t216 hook-audit split). Worktree
> `.claude/worktrees/t242`, branch `WT-chain-node-spawn`, HEAD `b7462203a`
> (local develop), measured 2026-09-02.
>
> First task per dispatch: call-graph trace before any dead/alive call. The
> t216 audit's "unwired = unreachable" form was overturned once already (t243);
> this verdict re-traces every reference independently.

## Claim

**The gap is a DEFECT, not an intended-incomplete.** `CreateNodeAtSpawn` has no
production caller, nothing sets `MOAI_CHAIN_NODE_ID`, and `events.jsonl` has
never existed — yet the SPEC that delivered the chain (SPEC-CHAIN-CORE-001)
explicitly planned a launch-path caller as an M1 deliverable with a matching
exit criterion, and never deferred it. Disposition per dispatch: **establish
the spawn-time node-creation caller** (done in this card, see § Disposition).

## Evidence

### E1 — call graph: `CreateNodeAtSpawn`

`git grep -n "CreateNodeAtSpawn"` over the full tracked tree @ `b7462203a`:

- Definition: `internal/chain/populate.go:53`
  `func (p *Populator) CreateNodeAtSpawn(worktreePath, specID, milestone string)`
- Production callers: **zero**. The only call sites are
  `internal/chain/populate_test.go` (6 test invocations).
- No `moai chain enter` subcommand exists: `internal/cli/chain.go:52-55`
  registers exactly `status / lineage / back / list / prune` — all read/prune
  paths.
- The plan-named integration file `internal/cli/worktree/new.go` **does not
  exist**; `git grep -n "chain" -- internal/cli/worktree/*.go` (non-test)
  returns one unrelated comment (`guard.go:51`, `exec.ExitError`).

### E2 — call graph: `MOAI_CHAIN_NODE_ID` / `EnvChainNodeID`

`git grep -n` for both the literal and the constant:

| Site | Direction |
|---|---|
| `internal/config/envkeys.go:279` | constant definition |
| `internal/chain/populate.go:54` | **read** (`os.Getenv`) — parent lookup |
| `internal/chain/populate.go:155` | **read** — fast-path resolve |
| `internal/hook/chain_banner.go:78` | **read** — SessionStart backfill/re-inject |
| `internal/chain/populate.go:51` (comment) | obligation note: "the caller MUST set as MOAI_CHAIN_NODE_ID on the child process environment" |

**SET sites: zero.** Every reference reads; nothing writes. The comment at
`populate.go:51` names the obligation and the missing caller in one line.

### E3 — the SPEC planned the caller explicitly

- REQ-CHAIN-005 (`spec.md:148-154`): "**When** a session enters a worktree via
  `moai cc -w`, the `EnterWorktree` path, or an `Agent(isolation: "worktree")`
  spawn, the chain population path **shall create** a new `WorktreeNode`…"
- plan.md M1 deliverables (`plan.md:106-110`): "Population at `moai cc -w`
  (spawn path in `internal/cli/worktree/new.go`)", "Population at
  `EnterWorktree` path", "`MOAI_CHAIN_NODE_ID` set on the child process".
- plan.md M1 exit criteria (`plan.md:116-118`): "entering a worktree via
  `moai cc -w` creates a ledger entry; `MOAI_CHAIN_NODE_ID` is set in the
  child environment; nested entry produces depth 2 then 3."
- design.md §3 Path A (`design.md:95-110`): five-step algorithm ending
  "Set `MOAI_CHAIN_NODE_ID=<node_id>` on the child process environment before
  `exec`'ing into Claude Code."

None of spec.md / plan.md / progress.md marks this deferred, phase-2, or
out-of-scope. The only forward-looking items (`acceptance.md` §D.5) defer
*real-environment validation* — not the population itself.

### E4 — how the AC matrix absorbed the gap

AC-CHAIN-005 (`acceptance.md:64-74`): the Given-When prose describes the spawn
path ("a new worktree session is entered via the spawn path"), but the
verification Command is `go test ./internal/chain/ -run
TestSpawnBoundaryNodeCreation -v` — a **package-internal unit test** of
`CreateNodeAtSpawn` called from the test itself. The M1 exit criterion
(ledger entry on `moai cc -w`, env set on the child) was **never carried into
any AC**. Result: progress.md §E.2 records 24/24 AC PASS and the SPEC closed
(completed at sync commit `435bc2bbd`) with the launch-path deliverable absent.
The gap was invisible because the AC measured the function, not the integration.

### E5 — ledger state re-measured today

`ls /Users/goos/MoAI/moai-adk-go/.moai/state/chain/` (primary checkout,
2026-09-02): `.gitkeep` only. No `events.jsonl`, matching t216's E3 finding
independently re-confirmed on a different tree and date.

### E6 — wiring state (context, not this card's fix)

`chain-event.sh` (SubagentStop completion-edge hook) IS wired in the
distributed template (`internal/template/templates/.claude/settings.json.tmpl`
line 213, fail-open wrapper) and NOT wired in this project's local
`settings.json` (t216 d-1 territory). Even where wired, the hook resolves the
node via `ResolveCurrentNode` and logs "no matching chain node" when the ledger
is empty — so with E1-E2 as measured, the wiring is a permanent no-op. The
caller is the upstream dependency of every downstream edge.

## Baseline-attribution

- All greps/reads in this verdict: this run, worktree t242 @ `b7462203a`
  (local develop tip), 2026-09-02.
- t216 d-1 findings were cross-checked (its own baseline: HEAD `a9eb896ce`,
  2026-08-24) and are consistent; the population gap is baseline to
  SPEC-CHAIN-CORE-001 Phase 1 as merged (`435bc2bbd`), nothing regressed it.

## Intended-incomplete test — why this is a defect

An "intended-incomplete" reading requires at least one of:
(a) a defer/Phase-2 marker in spec/plan/progress — **absent** (full read);
(b) an Out-of-Scope declaration covering the launch path — **absent**
    (`spec.md` § Out of Scope covers other topics; REQ-CHAIN-005 claims it);
(c) progress.md recording the M1 exit criterion as unmet — **absent** (§E.3
    records run complete, `ac_fail_count: 0`).

Against that: the M1 deliverable names the integration file explicitly and the
exit criterion states observable launch-path behavior. A planned deliverable
that shipped as an uncalled function, with the acceptance matrix reduced to a
unit test that never exercises the integration, is a delivery gap — not a
phase boundary.

## Disposition (this card continues into implementation)

Per dispatch: defect → **establish the spawn-time node-creation caller**.
Implemented in this card as Path A only:

- `internal/cli/launch_chain.go` — `injectChainNodeForLaunch`: at the
  `launchClaudeDefault` env-build site (launcher.go, before
  `execOrSpawnClaude`), when the launch carries `--worktree <name|path>`,
  resolve the worktree path (short name → `<root>/.claude/worktrees/<name>`,
  per the launcher's own resolution rule), append a node via
  `Populator.CreateNodeAtSpawn`, and set `MOAI_CHAIN_NODE_ID` on the child
  env (replace-or-append, `buildEnvForLaunch` pattern). Fail-open: any error
  warns on stderr and launches without a node — the chain is auxiliary and
  MUST NOT block a launch.
- `--continue`/resume launches are NOT a spawn boundary — no node.
- Fail-open limits (recorded, deliberate): `--worktree` with no name
  (claude auto-generates the name; the launcher cannot know it) creates no
  node; `moai cc -k/-f` labels and SPEC context are not propagated to
  `spec_id` (launch has no SPEC binding at this seam) — nodes record
  `spec_id=""` per `CreateNodeAtSpawn`'s optional-context contract.

Path B (`EnterWorktree`, a Claude Code session tool) and Path C
(`Agent(isolation: "worktree")`, a runtime spawn) are outside Go's reach at
this layer; their population path remains unimplemented here — recorded as a
known boundary, not silently claimed.

## Gaps

1. The "intended-incomplete" refutation is an absence-of-document argument
   (three full-file reads, zero defer markers). An operator decision made
   outside the SPEC artifacts would not appear there; none is recorded in the
   repo.
2. The end-to-end behavior (launch → ledger entry → SessionStart banner) was
   verified by unit tests at the new seam, not by a live `moai cc -w` launch
   in this session — a live launch would create a real tmux/session on this
   machine and is outside lane scope.
3. Path B/C population remains unimplemented (session-tool layer, above).
4. Whether the local `settings.json` should gain the chain-event wiring once
   nodes exist is t216 d-1's separate axis; this verdict does not change it.
5. **Depth-base documentation mismatch found during implementation.**
   design.md:103 says a parentless spawn "creates a depth-0 root node", but
   `CreateNodeAtSpawn` computes `depth = parentDepth + 1` with `parentDepth`
   zero-initialized — a parentless node is depth **1** — and the existing
   `TestCreateNodeAtSpawnDepth0` (whose name says 0) asserts
   `Depth != 1 → "want 1 (first node)"` (populate_test.go:53-55). Code and
   its test agree with each other and disagree with the design prose. The new
   launcher tests follow the code's behavior. Which base is canonical is a
   separate one-line decision (fix the prose or fix the +1) and is left here
   recorded, not fixed — outside this card's disposition.

## Residual-risk

- Short-name path prediction assumes claude resolves `-w <name>` to
  `.claude/worktrees/<name>` (the launcher's own documented rule,
  launcher.go:912-914). If claude's resolution diverges, ledger
  `worktree_path` values drift from reality — the env fast-path in
  `ResolveCurrentNode` still binds child sessions correctly, so the drift
  degrades only the `(worktree_path, session_id)` re-resolution fallback.
- Nodes now accumulate for every worktree launch (including disposable ones);
  `moai chain prune` (REQ-CHAIN-011, shipped) is the retention valve.
- The completion-edge hook remains unwired locally (t216 d-1) — the ledger
  will gain node-enter/backfill events but no completion edges until that
  separate decision lands.
