# SPEC-HARNESS-EVOLVE-001 — Implementation Plan

> Tier M · 21 REQ · 27 AC · 4 Milestones · v0.1.1 (plan-audit iter-1 fix applied). Design SSOT:
> `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (§4/§5/§7).
> Epic: HARNESS-EVOLVE (1/5, entry SPEC, no `depends_on`).

## §A. Context

### A.1 Baseline (measured 2026-07-12, main @ working tree)

| Surface | Measured baseline | Meaning |
|---------|-------------------|---------|
| `internal/harness/routing/` | absent (`ls` → no dir) | new package, greenfield TDD |
| `grep -c "routing-ledger" .claude/skills/moai/SKILL.md` | 0 | router registration baseline |
| `grep -rl "routing-ledger" .claude/skills/moai/workflows/` | 0 files | workflow-body wiring baseline |
| `grep -c "harness/routing" internal/cli/hook.go` | 0 | hook finalizer registration baseline |
| `grep -rn "routing-ledger\|routingledger" internal/cli/` | 0 | CLI verb registration baseline |
| registration-surface `routing-ledger` (`grep -rn "routing-ledger" .claude/skills/ internal/`) | 0 | zero-collision token on all registration surfaces — safe discriminator. (NOTE: a broader `.claude/` sweep matches 1 hit in `.claude/agent-memory/` — agent memory, NOT a registration surface; excluded by scope) |
| `grep -c '"ledger"' internal/cli/harness_retirement_test.go` | 0 | `v3r5RequiredHarnessVerbs` registration baseline (AC-HEV-027) |
| `.gitignore` | `.moai/state/` at lines 198 & 265 | ledger path already ignored; NO gitignore edit needed |
| Stop hooks (settings.json) | `handle-stop.sh`, `sync-phase-quality-gate.sh`, `handle-harness-observe-stop.sh` (async, 5s) | existing transport for outcome capture |
| `internal/cli/hook.go` `runHarnessObserveStop` | records usage-log Stop event + auto-classify + auto-propose, fail-open | host function for the additive finalizer |
| `internal/cli/hook.go` `isHookOptInEnabled` (hook.go:576-604 truth table; first gate at hook.go:731) | fail-CLOSED, default OFF; local `.moai/config/sections/system.yaml:14-15` = `opt_in: enabled: false`; template `system.yaml.tmpl` = `false` | gate 0 to inherit (REQ-HEV-016) — Stop-path DORMANT under default config (spec.md §D.3) |
| `internal/cli/hook.go` `isHarnessLearningEnabled` (hook.go:549) | fail-open, default TRUE; `learning.enabled: false` → no-op; local harness.yaml = `true` | gate 1 to inherit (REQ-HEV-016) |
| `moai harness` CLI group — live tree = `newHarnessRouterCmd()` (`internal/cli/harness_route.go`, registered `root.go:183`) | live verbs: `route/validate/status/clusters/apply/rollback/disable/mute/mute-list/unmute/verify/propose/install` (+ execute, v4 list/edit/remove/doctor) — NO `ledger` verb; `newHarnessCmd()` in `harness.go` is a superseded deprecation-marker factory NEVER added to the tree | namespace free for `ledger` sub-group; register in `newHarnessRouterCmd()` ONLY |
| `internal/harness/observer.go` | usage-log.jsonl writer (PostToolUse family) | PRESERVE — separate subject |

### A.2 Design decisions (pinned)

- **D1 — Stop-hook transport = extend the EXISTING `harness-observe-stop` Go
  handler (KEPT per user decision 1, HOI precondition codified).** No new
  wrapper script, no settings.json(.tmpl) registration change, no separate new
  gate, no global default flip. Rationale: the wrapper + registration +
  template mirror already ship; the Frozen permission surface (design §4,
  delta A1) is untouched; the 5s/async envelope is inherited. Consequence
  accepted knowingly (audit D1 + user decision 1): the handler's FIRST gate is
  `isHookOptInEnabled` (fail-CLOSED, default OFF per
  SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001) — Stop-path outcome finalization is
  DORMANT under default shipped config and requires `hook.opt_in.enabled: true`
  (spec.md §D.3). Alternative (new `moai hook routing-outcome` event + wrapper
  + registration) rejected: 3 extra mirrored files + permission-surface churn
  for zero capability gain. The SEPARATION requirement (REQ-HEV-009) binds the
  ledger FILE and schema types, not the hook transport process.
- **D2 — Finalization authority = the hook, not the orchestrator.** The
  orchestrator only ever supplies machine evidence refs (`ledger evidence`);
  `DeriveOutcome` (REQ-HEV-013) runs inside the Stop-hook handler. This keeps
  `outcome` un-fakeable by prose (a bare `--outcome success` flag DOES NOT
  EXIST on any CLI verb).
- **D3 — Multi-turn pipelines stay pending.** Stop fires every turn-end but
  `/moai plan|run|sync` span many turns; finalize triggers only when evidence
  is terminal-derivable. Stale/foreign pending rows are aborted lazily by the
  next `ledger record` (REQ-HEV-014) — no SessionEnd extension in v1.
- **D4 — HOI dual-gate inheritance.** Routing capture inherits BOTH
  harness-observe family gates in their existing order and asymmetry: gate 0
  `isHookOptInEnabled` (fail-CLOSED, default OFF — governs the Stop-finalize
  path riding the HOI-gated handler) then gate 1 `isHarnessLearningEnabled`
  (fail-open, default true — governs ALL routing capture; observation IS the
  learning system's front end). Neither default is changed. Without HOI
  opt-in, rows still finalize via record-time reroute + lazy abort sweep;
  `success`/`fail` outcomes require the Stop path (spec.md §D.3).
- **D5 — Per-session pending isolation.** Pending state at
  `.moai/state/routing-pending-<session_id>.json` (one file per session)
  avoids cross-session read-modify-write races; the ledger itself is only ever
  O_APPEND'ed whole lines (REQ-HEV-007). Empty/unresolvable session_id falls
  back to a `nosession` key (single-session degraded mode, still functional).
- **D6 — `model_class` / `session_id` resolution is best-effort.** Sources:
  hook input JSON `session_id` (Stop path), `moai session current` /
  statusline state (record path), model class from env/statusline snapshot;
  fallback `unknown` / empty — never a blocking lookup (5s budget).

### A.3 Template-mirror decision (explicit, per Template-First HARD rule)

| Artifact | Ships to `internal/template/templates/`? | How |
|----------|------------------------------------------|-----|
| Go code (`internal/harness/routing/`, `internal/cli/*`) | n/a (compiled into binary) | `make build` re-embeds; users get the mechanism via the binary |
| Hook wrapper scripts | NO CHANGE | existing `handle-harness-observe-stop.sh(.tmpl)` reused verbatim |
| settings.json hook registration | NO CHANGE | existing Stop registration reused verbatim |
| `.claude/skills/moai/SKILL.md` router edit | YES — template FIRST, then live | recording obligation text; neutral prose (NO SPEC IDs, NO dates, NO SHAs in template copy) |
| `workflows/{plan,run,sync}.md` edits | YES — template FIRST, then live | same neutrality discipline |
| `routing-ledger.jsonl` / pending files (DATA) | **NEVER** | REQ-HEV-020; empty state — templates carry no `.moai/state/` seed; AC-HEV-022 pins absence |
| `.gitignore` | NO CHANGE | `.moai/state/` rule already covers the ledger path |

Mirror-class caveat (measured discipline): before editing each skill file,
re-verify its live↔template pair class (byte-parity vs sanitized-pair) with a
fresh `diff`; do NOT assume parity from memory. Where the pair is sanitized
(live copy carries local-only content), apply the SAME semantic edit to both
sides rather than copying bytes.

### A.4 Hook self-gate design (explicit, per brief)

```
runHarnessObserveStop (existing, internal/cli/hook.go:726)
  gate 0 [EXISTING, hook.go:731]: !isHookOptInEnabled(root)      → return nil
         (HOI master toggle — fail-CLOSED, default OFF; REQ-HOI-002.
          The routing finalizer below is NEVER reached without HOI opt-in.)
  gate 1 [EXISTING]: !isHarnessLearningEnabled(root)             → return nil
         (fail-open, default true)
  ├── [existing] usage-log Stop event + auto-classify + auto-propose   (UNCHANGED)
  └── [NEW, additive, last] routing.FinalizePendingOnStop(root, sessionID)
        gate 2: pending file for session absent          → return nil   (self-gate, REQ-HEV-012)
        gate 3: DeriveOutcome(evidence) non-terminal     → return nil   (multi-turn, D3)
        else  : append ledger row (O_APPEND) + remove pending file
        errors: fmt.Fprintf(stderr, ...) and return nil  (fail-open, REQ-HEV-015)
```

Gates 0-1 are the EXISTING handler entry gates (not re-implemented by this
SPEC); the finalizer adds only gates 2-3. The CLI `record`/`evidence` paths
apply gate 1 (learning) themselves; gate 0 (HOI) binds the hook transport.

Budget: 1 stat + 1 small JSON read + 1 O_APPEND write worst case — well inside
the 5s hook timeout; no subprocess, no network.

### A.5 PRESERVE list (do not touch)

- `internal/harness/observer.go`, `outcome.go`, `learner.go`, `retention.go`,
  all usage-log / learning-history paths and their tests
- `.claude/hooks/moai/*.sh` + `internal/template/templates/.claude/hooks/**`
- `.claude/settings.json` + `settings.json.tmpl` (hook blocks)
- `.moai/harness/usage-log.jsonl`, `.moai/state/*` runtime files of other
  subsystems (context-usage.json, active-sessions.json, swarm/, verify/)
- `internal/harness/router/` (harness-level routing — different concern),
  `internal/cli/harness_route.go` (SPEC→harness-level routing verb)
- Unrelated SPEC directories and parallel-session artifacts

## §B. Known Issues / Risk Injection (B1-B12 filtered to relevant)

- **B3/B11 — Subagent boundary**: no `AskUserQuestion` in `internal/harness/routing/`
  or the hook path; CI guard grep must stay 0 (`subagent_boundary_test.go`
  pattern already exists in `internal/harness/`).
- **B4 — Frontmatter schema**: canonical `created:`/`updated:`/`tags:` used.
- **B6 — spec-lint heading convention**: `### Out of Scope — <topic>` H3
  sub-headings with `-` bullets present in spec.md §E.
- **B7 — Path resolution**: use `resolveHookProjectRoot()` (env-first
  `CLAUDE_PROJECT_DIR`, then Getwd) for all hook-path file access — the
  `internal/hook/.moai/` leak incident pattern. The `ledger record` CLI path
  resolves root the same way as sibling harness verbs.
- **B8 — Working-tree hygiene**: never commit `.moai/state/**`; tests use
  `t.TempDir()` exclusively (never the repo's own `.moai/state/`).
- **B1 — Cross-platform**: no syscalls beyond `os` file APIs; O_APPEND
  single-line append is the portable idiom; verify
  `GOOS=windows GOARCH=amd64 go build ./...`.
- **B2 — Cross-SPEC conflict pre-scan**: `internal/harness` carries retired /
  superseded SPEC markers (harness retirement → un-retired lifecycle); the new
  package must not resurrect retired surfaces — it is additive-only. Run the §C
  pre-scan before M1.
- **Flaky-risk carve-out**: NO wall-clock assertions in tests (the 5s budget is
  verified by design review + fail-open test, not by timing tests) — per the
  worktree-masked-flaky lesson.

## §C. Pre-flight (before M1 code)

```bash
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5          # lint baseline (NEW vs pre-existing)
ls internal/harness/routing 2>/dev/null || echo "greenfield OK"
grep -rn "Retired\|superseded" internal/harness/*.go | head -5   # B2 pre-scan
grep -rn "routing-ledger" .claude/skills/ internal/ | wc -l   # expect 0 on registration surfaces
# (scope note: .claude/agent-memory/ is EXCLUDED — the planning memory entry
#  legitimately mentions the token; it is not a registration surface)
```

## §D. Constraints (DO NOT VIOLATE)

- Machine-signal-only outcome: NO CLI flag or API accepting a free-text/bare
  outcome value (REQ-HEV-006/013). Evidence kinds are a closed enum.
- NO verbatim user text on disk (REQ-HEV-005): digest computed in-process; the
  `record` verb receives request text via stdin and discards it after hashing.
- NO new hook wrapper / settings registration / SessionEnd change (§E spec.md).
- NO writes to CLAUDE.md / CLAUDE.local.md / managed blocks (EVOLVE-002).
- NO usage-log schema reuse or writes (REQ-HEV-009).
- Template-First ordering for skill docs; template copies stay neutral
  (no SPEC-HARNESS-EVOLVE-001 token, no internal dates/SHAs in
  `internal/template/templates/**`).
- Conventional Commits `feat(SPEC-HARNESS-EVOLVE-001): M{N} <subject>`; never
  `--no-verify`; specific-path `git add` only.
- Coverage: `internal/harness/routing` ≥ 90%; touched `internal/cli` paths keep
  package ≥ existing baseline.

## §E. Self-Verification (run-phase deliverables)

Per `manager-develop-prompt-template.md` §E (E1-E7) and
`verification-claim-integrity.md` §3 (5-section evidence format):

- E1: AC-HEV-001..027 binary PASS/FAIL matrix with verbatim command outputs
- E2: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- E3: `go test -cover ./internal/harness/routing/...` ≥ 90%
- E4: subagent-boundary grep = 0 matches in `internal/harness/routing/`
- E5: `golangci-lint run` — NEW issues 0 (baseline separately noted)
- E6: commit SHAs + push state
- E7: structured blocker report if any (never AskUserQuestion)

## §F. Milestones (priority-ordered, no time estimates)

### M1 — `internal/harness/routing/` core package (Priority: High)

Files: `internal/harness/routing/{types.go, digest.go, writer.go, reader.go,
outcome.go, pending.go}` + `_test.go` siblings (table-driven, `t.TempDir()`).

- Schema v1 types (spec.md §D.1) with `schema_version: 1`, nullable
  convergence fields, `delegations[]`, closed `evidence_refs` kind enum
- `request_digest` util: truncated SHA-256, raw-text-never-persisted contract
  (test: `TestRequestDigestNoVerbatim`)
- Append-only writer: `O_APPEND|O_CREATE|O_WRONLY`, one line per row,
  concurrent-append test
- Reader with filters: subcommand / outcome / time-window; malformed-line
  fail-open (test: `TestReaderSkipsMalformed`)
- `DeriveOutcome` fixed precedence table (test: `TestDeriveOutcome` —
  abort > fail > success > non-terminal; no override path)
- Pending-row store: per-session file, reroute-on-re-record
  (test: `TestReroute`), stale/foreign sweep → abort
  (test: `TestStaleSweepAbort`), self-gate no-op
  (test: `TestFinalize_SelfGate_NoPendingNoOp`), fail-open
  (test: `TestFinalize_FailOpen`)
- Boundary guard: `routing/subagent_boundary_test.go` (grep-style CI guard)

Exit: `go test ./internal/harness/routing/...` PASS, coverage ≥ 90%.

### M2 — CLI verbs + Stop-hook finalizer wiring (Priority: High)

Files: `internal/cli/harness_ledger.go` (+`_test.go`), `internal/cli/hook.go`
(additive edit in `runHarnessObserveStop`), `internal/cli/harness_route.go`
(**sub-command registration in `newHarnessRouterCmd()` — the LIVE tree wired
at `root.go:183`; do NOT touch the superseded `newHarnessCmd()` in
`harness.go`, which is a deprecation-marker factory never added to the
tree**), `internal/cli/harness_retirement_test.go` (verb-surface guard).

- `moai harness ledger record` — stdin request text → digest + class, flags
  for subcommand/mode/tier/level/clarify-rounds; creates pending row; runs the
  lazy stale sweep (REQ-HEV-011/014, cross-session or unresolvable+24h,
  same-session→reroute precedence); gated by `isHarnessLearningEnabled`
- `moai harness ledger evidence` — appends a machine evidence ref
  (`--kind gate_exit|audit_score|verify_path|abort`, `--value`, `--ref`,
  `--terminal`); NO `--outcome` flag exists on record OR evidence (the
  write surfaces — AC-HEV-011)
- `moai harness ledger list` — reader filters
  (`--subcommand`, `--outcome`, `--since`, `--until`), JSONL/table output
  (the read-side `--outcome` FILTER is legitimate — it selects rows, it never
  writes an outcome)
- `ledger` added to `v3r5RequiredHarnessVerbs`
  (`internal/cli/harness_retirement_test.go:31-50`) so the verb-surface CI
  guard pins the registration (AC-HEV-027)
- `runHarnessObserveStop` gains the §A.4 self-gated finalizer call (last,
  additive, fail-open, downstream of the EXISTING HOI + learning gates); gate
  test (`TestHarnessObserveStop_RoutingLedgerGated` — fixtures set
  `hook.opt_in.enabled: true` to reach the finalizer, then vary
  `learning.enabled`; plus an HOI-off dormancy case) + no-pending no-op test

Exit: `moai harness ledger --help` exits 0; `go test ./internal/cli/...` PASS;
existing hook tests untouched-green.

### M3 — Workflow skill wiring, Template-First (Priority: Medium)

Files (template FIRST, then live):
`internal/template/templates/.claude/skills/moai/SKILL.md` → `.claude/skills/moai/SKILL.md`;
same pairing for `workflows/plan.md`, `workflows/run.md`, `workflows/sync.md`.

- SKILL.md router: routing-ledger recording obligation at dispatch (short
  block; discriminating token `routing-ledger`)
- plan.md / run.md / sync.md: dispatch-record reference + ≥1 evidence-append
  point each, naming a machine signal (plan: plan-audit verdict; run: gate
  exits / verify paths; sync: sync-gate exit)
- Neutral prose in template copies (no internal SPEC IDs/dates/SHAs)
- `make build` (re-embed) + live↔template pair-class re-verification diff

Exit: baseline-0 → ≥1 greps on all 4 live files + 4 template mirrors;
neutrality CI guard green.

### M4 — Integration verification + evidence (Priority: Medium)

- Enable the HOI opt-in LOCALLY for verification: set
  `hook.opt_in.enabled: true` in this dev repo's
  `.moai/config/sections/system.yaml` (local runtime config, NOT template —
  the shipped default stays `false` per user decision 1 / spec.md §D.3), then
  exercise the live Stop-path finalization end-to-end
  - **Disposition (N-1)**: this local enable is a DELIBERATE, committed
    dogfood-enable — accumulating routing-observation data in this dev repo is
    the very purpose of the HARNESS-EVOLVE Epic, so the `system.yaml` change is
    committed (not reverted post-verification). NOTE the `hook.opt_in.enabled`
    master toggle activates ALL 3 observe wrappers (PostToolUse observe +
    SessionStart + Stop), not only the Stop finalizer this SPEC adds; this is
    local-repo config only and never changes the template default (`false`).
- Full suite `go test ./...` (+ `-race` for the concurrent-append path),
  `go vet ./...`, `golangci-lint run` NEW-issue check
- `moai spec lint` clean for this SPEC; template neutrality test green
- AC-HEV-001..027 sweep with verbatim outputs → progress.md §E.2
- @MX tags: `@MX:ANCHOR` on the writer append entry point (fan-in ≥3 expected:
  CLI record/finalize + hook path), `@MX:NOTE` on DeriveOutcome precedence

Exit: E1-E7 self-verification complete; ready for sync-phase handoff.

## §G. Anti-Patterns (this SPEC)

- AP-1: Adding an `--outcome` flag "for testing convenience" — breaks the
  un-fakeable contract (REQ-HEV-006/013). Test seams use evidence fixtures.
- AP-2: Finalizing on first Stop unconditionally — truncates multi-turn
  pipelines (D3). Terminal-derivable evidence is the only finalize trigger.
- AP-3: Reusing usage-log `Event` types "to save code" — couples the two
  observation subjects (REQ-HEV-009).
- AP-4: Editing the live skill file first and back-porting to the template —
  inverts Template-First; template is edited FIRST.
- AP-5: Wall-clock timing asserts for the 5s budget — flaky; verify by design
  (operation count) + fail-open test instead.
- AP-6: Writing tests against the repo's real `.moai/state/` — B8 violation;
  `t.TempDir()` only.

## §H. Resolved Clarifications (pinned — AskUserQuestion round 2026-07-12)

> All markers STRUCK per plan-audit iter-1 D4/F4. The three user decisions
> below are PINNED; no open clarifications remain (MP-7 gate clear).

1. **HOI transport (audit D1) — RESOLVED, decision 1**: KEEP the HOI-gated
   `runHarnessObserveStop` transport. The activation precondition is codified
   explicitly (spec.md §D.3 + REQ-HEV-016 dual-gate): Stop-path outcome
   finalization requires `hook.opt_in.enabled: true` (fail-closed, per
   SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001). NO separate gate added, NO global
   default flipped. Default-config Stop-path dormancy is accepted shipped
   behavior; this dev repo enables the opt-in locally during M4.
2. **request_class in schema v1 — RESOLVED, decision 2**: INCLUDED in v1 as a
   coarse keyword-derived enum (`feature|bugfix|refactor|docs|question|pipeline|other`,
   non-verbatim), alongside `request_digest` — enum aligned verbatim to the
   spec.md §D.1 SSOT (N-2).
3. **Ledger retention/rotation — RESOLVED, decision 3**: v1 no-rotation
   (append-only single file); retention revisited in EVOLVE-003 alongside the
   negative-evidence registry, with the option to reuse
   `internal/harness/retention.go` explicitly preserved (spec.md §E
   Out-of-Scope entry). Residual risk (unbounded growth) accepted for v1 —
   mitigated by ~300-500 B row size + the EVOLVE-003 revisit.

## §I. Cross-References

- spec.md §C (REQ SSOT) / §D (schema SSOT) / §E (exclusions)
- acceptance.md (AC matrix SSOT, AC-HEV-001..027)
- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` §5/§6/§7
- `internal/cli/hook.go` (`runHarnessObserveStop`, `isHookOptInEnabled`,
  `isHarnessLearningEnabled`, `resolveHookProjectRoot`)
- `internal/cli/harness_route.go` (`newHarnessRouterCmd()` — the LIVE CLI
  registration site, wired `root.go:183`) + `harness_retirement_test.go`
  (`v3r5RequiredHarnessVerbs`)
- SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 (HOI master toggle — gate 0 contract)
- `internal/harness/observer.go` (separation reference)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (§B/§E)
- CLAUDE.local.md §2 / §25 (Template-First + neutrality)
