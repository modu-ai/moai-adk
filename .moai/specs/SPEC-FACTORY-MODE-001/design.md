# SPEC-FACTORY-MODE-001 — Design

Version: 0.4.0 | Tier: L

## 1. The chain, and what Factory adds to it

```
  moai cc --factory [SPEC-ID]
        │
        │  launcher: strip flag, set MOAI_FACTORY / MOAI_FACTORY_SPEC
        │            (defer-restored to prior state on return),
        │            write .moai/state/factory/<session>.json,
        │            inject CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200
        ▼
  ┌──────────┐
  │   plan   │  ← factory delta: the chain head
  └────┬─────┘
       │  ▲ GATE 1 — Implementation Kickoff Approval  (existing, mandatory)
       ▼  │        + factory_chain goal armed here, alongside the work
  ┌──────────┐
  │   run    │
  └────┬─────┘
       │
       ▼
  ┌──────────┐
  │  verify  │  ← factory delta: /moai review --security --deep --repo
  └────┬─────┘     the EXIT GATE of run, not a stage of sync
       │
       ├── confirmed CRITICAL / HIGH ──► GATE 2 (human) ──► re-enter run,
       │                                                    scoped to changed
       │                                                    surface (max ×2)
       │
       ├── DEGRADED rung ──────────────► proceeds (rung is an attribute of S1/S2,
       │   (attribute, not a 4th case)    not a peer case); verify_rung recorded
       │                                  → Phase 8 suppression forced OFF even
       │                                  on a predicate match (REQ-FM-013 (b))
       │
       ├── NO readable result ─────────► HALT + 5-section verdict + escalate
       │   (errored / aborted /          (does NOT count against the ×2 ceiling;
       │    findings.jsonl absent)        never classified as "no findings")
       │
       └── readable, only MEDIUM / LOW,─┐
           or readable with none        │  inherited full-pipeline auto-chain
                                        ▼  (no approval at this boundary)
                                   ┌──────────┐
                                   │   sync   │
                                   └──────────┘
                                     GATE 3 = gate-sync-1 (inherited)
                                     Phase 8 Step 0.55.1 skipped iff
                                       revision_match AND
                                       verify_rung recorded ∈ {PRIMARY, FALLBACK}
                                       (absent / empty / DEGRADED → no skip)
                                     dependency manifest audit ALWAYS runs
                                     GATE 4 = gate-sync-2
```

The verify outcome structure is **two axes, not one list** (corrected in v0.3.0):

- The **severity partition** is disjoint and jointly exhaustive over three cases — S1 readable + confirmed CRITICAL/HIGH, S2 readable + MEDIUM/LOW-or-none, S3 no readable result. Readability is what separates S3 from S2, which is the property the v0.1.0 draft lacked: it had only the first and last branches plus a DEGRADED label with no consequence, so a verify that produced nothing at all fell into the last branch and reached sync looking scanned. A gate whose failure mode is "proceed" is not a gate.
- The **rung** (`PRIMARY` / `FALLBACK` / `DEGRADED`) is an **attribute of** an S1 or S2 result, not a fourth peer case. A readable DEGRADED result with no confirmed findings is S2 *and* DEGRADED simultaneously. The v0.2.0 draft listed the two axes as four peer outcomes and asserted they were disjoint; they overlap by construction, and the overlap left REQ-FM-011 ("proceed") and REQ-FM-013 ("not an automatic pass") both binding with no precedence — while REQ-FM-013's stated surfacing point, the CRITICAL/HIGH gate, was unreachable for exactly the non-critical result where the rung matters most.

**Precedence, which the design now states rather than implies:** the severity case governs routing and the rung never changes it; the rung governs sync-phase suppression and the severity case never relaxes it. So S2 × DEGRADED proceeds to sync with Phase 8 Step 0.55.1 suppression forced OFF — the chain keeps moving (it is an unattended four-hour chain by REQ-FM-027) while the independent adversarial analysis of the same surface still runs. The alternative — widening the CRITICAL/HIGH gate to fire on DEGRADED as well — was rejected in REQ-FM-013 because it halts an unattended chain for a human decision at a moment when no critical finding exists, spending the operator's attention where nothing is at stake.

The run→sync arrow is **not new**. It is the `full-pipeline` contract already documented in `moai.md`, including its clause that the sync-internal gates keep firing. Factory contributes the plan head and the verify gate, and nothing else about how phases chain.

## 2. Contract relationship

Three contracts now exist, and they are ordered by inclusion rather than by alternative:

| Contract | Trigger | run→sync boundary | verify stage |
|---|---|---|---|
| `single-phase` | explicit `/moai run` or `/moai sync` | surfaced as the "(Recommended)" first option of a next-step question; never silent | absent |
| `full-pipeline` | default `/moai` route | auto-chains, announced in the transcript | absent |
| `factory` | `--factory` on `moai cc` / `moai glm` | **inherited verbatim from `full-pipeline`** | present, at run exit |

Stating `factory` as an extension rather than a peer is the load-bearing choice. A peer contract would have to restate the auto-chain semantics and the gate-preservation clause, and those two statements would then drift against `moai.md` independently. An extension has exactly two deltas to keep in sync.

## 3. Why verify is the run exit gate

Placing verify after the run→sync boundary would mean a confirmed CRITICAL finding has already entered a phase whose job is to write documentation, update CHANGELOG, and (on Tier L) open a PR. Undoing that is more expensive than never entering it.

Placing it at run exit means the failure mode is a loop back into the phase that owns code changes — which is where a security fix belongs — and the re-entry is naturally scoped to the surface that changed, because that surface is what the scan flagged.

The 2-re-entry ceiling exists because a third identical re-entry is evidence that the loop is not converging, and the correct response to a non-converging security loop is a human with the 5-section verdict in front of them, not a fourth attempt.

## 4. The dedup predicate

Two inputs are derived at sync entry, never supplied by judgement:

```
headSHA   := stdout of `git rev-parse HEAD`
treeDirty := `git status --porcelain` has ≥1 line after excluding
             .moai/state/ .moai/reports/ .moai/cache/ .moai/logs/
             and the results directory itself
```

```
                    ┌─────────────────────────────┐
   sync entry ─────►│ read .moai/state/factory/   │
                    │ <session>.json → deepscan_dir│
                    └──────────────┬──────────────┘
                                   │ absent / unreadable
                                   ├────────────────────────► FALSE
                                   ▼
              findings.jsonl present AND every line parses ?
                                   ├── no ─────────────────► FALSE
                                   ▼
                    ┌─────────────────────────────┐
                    │ LoadRevision(dir/revision.json)
                    └──────────────┬──────────────┘
                       error / malformed ├────────────────────► FALSE
                                   ▼
              scanned_commit == headSHA ? ── no ─────────────► FALSE
                                   │ yes
              scope == "repo" ?        ── no ──────────────► FALSE
                                   │ yes
              treeDirty ?
                  ├── no  ─────────────────────────────────► TRUE
                  └── yes ─► working_tree_included == true ?
                                   ├── yes ────────────────► TRUE
                                   └── no  ────────────────► FALSE
```

Every edge that is not the single TRUE path leads to FALSE, and FALSE means Phase 8 Step 0.55.1 runs. The predicate has one accepting path and seven rejecting ones; that asymmetry is the design, not an accident of enumeration.

`findings.jsonl` sits at the top of the chain because `revision.json` cannot answer the question it appears to answer. Its schema — `scanned_commit`, `effort_tier`, `working_tree_included`, `scope`, `generated_at` — has no completion or status field, so an aborted run that stamped it is indistinguishable from a completed one. `findings.jsonl` is the completion signal: a clean scan writes it with zero lines (which the predicate accepts), while an aborted scan characteristically never writes it at all.

**The predicate is only half the suppression gate.** The DEGRADED rung is not recoverable from `revision.json` — `effort_tier` is an effort level, not a rung, and treating a `max` effort tier as evidence of a full 3-voter panel would be a category error. So the rung is carried on the factory state record (`verify_rung`), written by the orchestrator from the label the verify stage surfaces per REQ-FM-013, and the composed decision is:

```
suppress_step_0551 := revision_match AND verify_rung ∈ {"PRIMARY", "FALLBACK"}
```

The membership test is an **allow-list**: an absent, empty, or unrecognized `verify_rung` is outside the set and therefore yields no suppression. The `!= "DEGRADED"` phrasing this replaces was the v0.2.0 deny-list that failed open on absence — see §7.

The alternative — adding a rigor field to `revision.json` — would mean extending an artifact schema this SPEC does not own, and is listed out of scope.

The predicate is a **reader**. Nothing in this SPEC writes `revision.json` or `findings.jsonl`, and the design does not assume anything about who does beyond "the `--deep` workflow agents, per doctrine". A reader that returns FALSE when the artifact is missing is correct whether or not a writer ever ships — which is precisely why the contract can be kept in v1 despite R1.

**A reader nothing calls is dead code.** Every unit test over a pure function passes regardless of whether production ever invokes it, so the sync Phase 8 dedup gate doctrine names the call and both derived inputs explicitly. That is what makes the §B.3 criteria falsifiable rather than merely green.

## 5. Launcher integration

`runCC` currently sequences: `--help` scan → `stripSpawnFlag` → `parseProfileFlag` → `resolveWorktreeL2Path` → `normalizeWorktreeFlag` → `unifiedLaunch`.

`parseFactoryFlag` inserts after `parseProfileFlag` and before the worktree handling. Rationale: it must run after `--spawn` is stripped (a spawned session re-issues the same command and must carry `--factory` through), and before worktree normalization so that a `--factory` token can never be mistaken for a `-w` value by `normalizeWorktreeFlag`'s "consume the next non-flag token" branch.

All four parsers share the same `--` discipline: iterate, break on `--`, append the remainder verbatim. `parseFactoryFlag` copies that shape rather than inventing a new one.

## 6. Block-cap inject

The existing `injectStopHookBlockCapForGoal` is goal-conditional by design: it exists so that "arm an infinite goal, then launch and walk away" works. Factory arms its goal *after* launch, so that predicate is structurally unable to see it.

It has exactly one production caller, `internal/cli/launcher.go:790` inside `launchClaudeDefault`, reached from `runCC` through five hops:

```
runCC → unifiedLaunch → unifiedLaunchFunc(=unifiedLaunchDefault)
      → launchClaude → launchClaudeFunc(=launchClaudeDefault) → :790
```

None of those hops carries a factory parameter, and `unifiedLaunch(profile, mode, args)` receives `args` with the `--factory` token already stripped by REQ-FM-001. So the flag is simply not in scope at the call site, and the extension needs a transport.

**The transport is the process environment**, not a threaded parameter:

```
runCC / runGLM:
    prev, had := os.LookupEnv(config.EnvMoaiFactory)      # capture prior STATE, not just value
    os.Setenv(config.EnvMoaiFactory, "1")
    defer restore(config.EnvMoaiFactory, prev, had)       # Unsetenv when !had — not Setenv("")
    ... same capture/set/defer pair for config.EnvMoaiFactorySpec, when supplied ...

inject(ctx, base, projectRoot, sessionID):          # signature UNCHANGED
    if os.Getenv(config.EnvMoaiFactory) != "":      # NEW — unconditional
        return setCap(base, DefaultRaisedStopHookBlockCap)
    ... existing armed-infinite-goal branch, verbatim ...
```

**The `defer` restore is not hygiene, it is a correctness requirement** (added v0.3.0). `os.Setenv` is process-global. Inside the `internal/cli` test binary, AC-FM-001 calls `runCC` with `--factory`; without the restore, `MOAI_FACTORY=1` persists for every later test in that process, so AC-FM-022a's stated negative control — "the identical call with `MOAI_FACTORY` unset returns `base` unchanged" — passes or fails by test-execution order. That is the ordering-dependent flake this project's test-isolation rules exist to prevent, and it would have been *caused by the SPEC's own criterion*. In production the same leak lets a later re-exec or spawn inherit factory semantics it was never given. The restore must cover the error return path as well, and must restore *absence* rather than an empty string, because `os.Getenv` cannot distinguish unset from empty but `os.LookupEnv` (and therefore the assertion in AC-FM-023d) can. Setting the variables on the launch env slice instead was rejected: the inject reads `os.Getenv` at `launcher.go:790` and would not see a slice-local value without the signature change §6 exists to avoid.

This works because `launchEnv` is built from `os.Environ()` at `launcher.go:783` (GLM branch) and `:786` (Claude branch), immediately above the call — so the variable is already inside the slice the inject mutates, and it also reaches the child process, which the orchestrator needs for `MOAI_FACTORY_SPEC` anyway.

**Rejected: threading a parameter.** It would change `unifiedLaunch`, the `unifiedLaunchFunc` variable type, `launchClaude`, the `launchClaudeFunc` variable type, and both test seams — five signature changes to carry a signal the environment already delivers to the same line.

**Consequence, made explicit.** REQ-FM-024's environment variables become load-bearing for REQ-FM-023. Rename or drop them and the cap raise silently stops working: the session just ends after 8 Stop-hook blocks, with no error and no obvious cause. AC-FM-023c asserts the variables survive into the child environment for exactly this reason — a unit test of the inject alone would stay green through that failure.

One constant, one env key pair, one function, one call site. A second constant or a parallel inject site would be the failure this design exists to avoid.

**Blast radius.** The raised cap is session-wide. It is applied at launch, before Implementation Kickoff Approval has been asked, so a user who declines the gate — or who arms an unrelated goal later in the same session — still carries a 200-block ceiling. This is not a gate bypass: all four human gates are orchestrator-issued `AskUserQuestion` rounds, not Stop-hook blocks, so a raised block cap cannot skip any of them. It is a longer unattended leash on an unrelated loop, and `factory.md` documents it so an operator can recognize the cause. Scoping the cap to a single goal would require a per-goal runtime mechanism Claude Code does not expose.

## 7. State record

```json
{
  "session_id": "<uuid>",
  "spec_id": "<spec identifier, or empty when the chain heads at plan>",
  "backend": "claude | glm",
  "entered_at": "<RFC3339>",
  "deepscan_dir": ".moai/reports/security-deepscan-<timestamp>",
  "verify_rung": "PRIMARY | FALLBACK | DEGRADED",
  "verify_reentries": 0
}
```

Written best-effort at launch; `deepscan_dir`, `verify_rung`, and `verify_reentries` are filled in by the orchestrator as the chain progresses. A write failure never blocks a launch — the chain degrades to a session with no record, which is exactly a non-factory session plus an unusable dedup predicate (which then returns FALSE and runs Phase 8 Step 0.55.1, the safe direction).

**The three orchestrator-written fields are written independently, which is why the rung check is an allow-list** (v0.3.0). `deepscan_dir`, `verify_rung`, and `verify_reentries` are filled in at different points in the chain on a best-effort record. A record carrying `deepscan_dir` but not `verify_rung` is therefore reachable — an orchestrator omission, a `/clear` mid-chain, a partial write — and under the v0.2.0 `!= "DEGRADED"` phrasing that record cleared the rung exclusion and permitted suppression. Empty is not `DEGRADED`, so the deny-list failed open on exactly the state a best-effort writer produces. REQ-FM-014 now requires `verify_rung` to be *recorded and equal to* `PRIMARY` or `FALLBACK`; every other value, absent and empty included, yields no suppression. This is the same fail-safe direction as the revision-match predicate: absence means run the check.

`verify_rung` exists because the rung is not recoverable from any artifact this SPEC can read: `review.md` specifies that the DEGRADED rung self-labels in the report, but `revision.json` carries no rigor field. Recording it here — on a record this SPEC owns — is what makes the REQ-FM-014 exclusion mechanically checkable instead of a prose hope.

## 8. Rejected alternatives

- **A `moai factory` subcommand.** Rejected: it would need its own profile, worktree, and spawn handling, duplicating three parsers for no gain. An entry switch on the two existing launchers reuses all of it.
- **A `/moai verify` subcommand.** Rejected by the settled decisions, and independently by the retirement of `/moai security` — the deep scan is a *mode* of review, and adding a second entry point to it would reintroduce exactly the surface that was retired.
- **Verify as sync Phase 6.5.** Rejected: see §3.
- **Dropping the dedup contract until a writer exists.** Rejected: the predicate is safe under absence, so keeping it costs nothing today and pays off the moment a writer lands.
- **Supporting `moai cg`.** Rejected: the leader runs Claude and the teammates run GLM, so "one session, one backend, one chain" does not hold and the verify stage would run under an indeterminate backend.
- **Adding `effort_tier` to the dedup predicate with a named floor.** Rejected: `effort_tier` is an effort level (`low`…`max`), and the rung (`PRIMARY` / `FALLBACK` / `DEGRADED`) is what determines whether a 3-voter panel ran. A `max`-effort single-pass scan would clear any effort floor while still being rigor-reduced, so the floor would read as a safeguard and function as none. The rung is carried on the factory state record instead.
- **Adding a rigor field to `revision.json`.** Rejected: it is an artifact owned by the `--deep` workflow agents, and extending its schema would put this SPEC in the producer's business — which §C explicitly excludes.
- **Threading a factory parameter through the launch chain.** Rejected — see §6 for the five signature changes it would require and the environment route chosen instead.
