# SPEC-HARNESS-EVOLVE-001 — Implementation Plan

> Tier M · 21 REQ · 26 AC · 4 Milestones. Design SSOT:
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
| repo-wide `routing-ledger` (internal/ + .claude/, excl. reports) | 0 | zero-collision token — safe discriminator |
| `.gitignore` | `.moai/state/` at lines 198 & 265 | ledger path already ignored; NO gitignore edit needed |
| Stop hooks (settings.json) | `handle-stop.sh`, `sync-phase-quality-gate.sh`, `handle-harness-observe-stop.sh` (async, 5s) | existing transport for outcome capture |
| `internal/cli/hook.go` `runHarnessObserveStop` | records usage-log Stop event + auto-classify + auto-propose, fail-open | host function for the additive finalizer |
| `internal/cli/hook.go` `isHarnessLearningEnabled` | default TRUE; `learning.enabled: false` → no-op | gate to inherit (REQ-HEV-016) |
| `moai harness` CLI group | `route/validate/status/apply/rollback/disable/mute/list/edit/remove/doctor` — NO `ledger` verb | namespace free for `ledger` sub-group |
| `internal/harness/observer.go` | usage-log.jsonl writer (PostToolUse family) | PRESERVE — separate subject |

### A.2 Design decisions (pinned)

- **D1 — Stop-hook transport = extend the EXISTING `harness-observe-stop` Go
  handler.** No new wrapper script, no settings.json(.tmpl) registration change.
  Rationale: the wrapper + registration + template mirror already ship; the
  Frozen permission surface (design §4, delta A1) is untouched; the 5s/async
  envelope is inherited. Alternative (new `moai hook routing-outcome` event +
  wrapper + registration) rejected: 3 extra mirrored files + permission-surface
  churn for zero capability gain. The SEPARATION requirement (REQ-HEV-009)
  binds the ledger FILE and schema types, not the hook transport process.
- **D2 — Finalization authority = the hook, not the orchestrator.** The
  orchestrator only ever supplies machine evidence refs (`ledger evidence`);
  `DeriveOutcome` (REQ-HEV-013) runs inside the Stop-hook handler. This keeps
  `outcome` un-fakeable by prose (a bare `--outcome success` flag DOES NOT
  EXIST on any CLI verb).
- **D3 — Multi-turn pipelines stay pending.** Stop fires every turn-end but
  `/moai plan|run|sync` span many turns; finalize triggers only when evidence
  is terminal-derivable. Stale/foreign pending rows are aborted lazily by the
  next `ledger record` (REQ-HEV-014) — no SessionEnd extension in v1.
- **D4 — `learning.enabled` gate inheritance.** Routing capture no-ops when
  `learning.enabled: false`, consistent with the whole harness-observe family
  (observation IS the learning system's front end; if learning is off,
  observing for it is off). Default remains enabled.
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
runHarnessObserveStop (existing, internal/cli/hook.go)
  ├── [existing] usage-log Stop event + auto-classify + auto-propose   (UNCHANGED)
  └── [NEW, additive, last] routing.FinalizePendingOnStop(root, sessionID)
        gate 0: isHarnessLearningEnabled(root) == false → return nil   (REQ-HEV-016)
        gate 1: pending file for session absent          → return nil   (self-gate, REQ-HEV-012)
        gate 2: DeriveOutcome(evidence) non-terminal     → return nil   (multi-turn, D3)
        else  : append ledger row (O_APPEND) + remove pending file
        errors: fmt.Fprintf(stderr, ...) and return nil  (fail-open, REQ-HEV-015)
```

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
grep -rn "routing-ledger" .claude/ internal/ | wc -l    # expect 0 (token still free)
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

- E1: AC-HEV-001..026 binary PASS/FAIL matrix with verbatim command outputs
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
(additive edit in `runHarnessObserveStop`), `internal/cli/harness.go`
(sub-command registration).

- `moai harness ledger record` — stdin request text → digest + class, flags
  for subcommand/mode/tier/level/clarify-rounds; creates pending row; runs the
  lazy stale sweep (REQ-HEV-011/014); gated by `isHarnessLearningEnabled`
- `moai harness ledger evidence` — appends a machine evidence ref
  (`--kind gate_exit|audit_score|verify_path`, `--value`, `--ref`,
  `--terminal`); NO outcome flag exists
- `moai harness ledger list` — reader filters
  (`--subcommand`, `--outcome`, `--since`, `--until`), JSONL/table output
- `runHarnessObserveStop` gains the §A.4 self-gated finalizer call (last,
  additive, fail-open); gate test
  (`TestHarnessObserveStop_RoutingLedgerGated`) + no-pending no-op test

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

- Full suite `go test ./...` (+ `-race` for the concurrent-append path),
  `go vet ./...`, `golangci-lint run` NEW-issue check
- `moai spec lint` clean for this SPEC; template neutrality test green
- AC-HEV-001..026 sweep with verbatim outputs → progress.md §E.2
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

## §H. Open Clarifications

> To be resolved via orchestrator AskUserQuestion before Implementation
> Kickoff Approval (plan→run gate). Both have a pinned default so run-phase is
> not blocked if the defaults are confirmed.

1. **[NEEDS CLARIFICATION: request_class field in schema v1]** — The design
   SSOT §5 ledger schema line lists `request_digest` only, but the Loop 1
   pattern key (design §5 Loop 1: `request_class + subcommand + mode +
   outcome`) requires a class token. Proposed default (as drafted in spec.md
   §D.1): include `request_class` as a coarse keyword-derived enum
   (`feature|bugfix|refactor|docs|question|pipeline|other`) in v1 — non-verbatim,
   privacy-preserving. Alternative: omit in v1 and let EVOLVE-002 add schema v2.
2. **[NEEDS CLARIFICATION: ledger retention/rotation]** — The ledger is
   append-only and unbounded in v1. usage-log.jsonl has a `Retention`
   (archive + prune) component. Proposed default: NO rotation in v1 (rows are
   ~300-500 B; thousands of routings ≪ 1 MB; premature rotation complicates the
   Loop 1 consumer contract); revisit in EVOLVE-003 alongside the negative-
   evidence registry. Alternative: reuse `Retention` at 10 MB single-level
   rotation now.

## §I. Cross-References

- spec.md §C (REQ SSOT) / §D (schema SSOT) / §E (exclusions)
- acceptance.md (AC matrix SSOT, AC-HEV-001..024)
- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` §5/§6/§7
- `internal/cli/hook.go` (`runHarnessObserveStop`, `isHarnessLearningEnabled`,
  `resolveHookProjectRoot`)
- `internal/harness/observer.go` (separation reference)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (§B/§E)
- CLAUDE.local.md §2 / §25 (Template-First + neutrality)
