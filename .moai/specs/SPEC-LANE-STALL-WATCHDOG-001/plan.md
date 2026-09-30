# plan.md — SPEC-LANE-STALL-WATCHDOG-001

## §A Context

- Worktree: `.moai/worktrees/t1370` (absolute:
  `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1370`), branch
  `WT-lane-stall-watchdog`; 0.1.0 authored at `3dd5adf2f`, iter-2 revision at
  the commit carrying v0.2.0.
- SPEC artifacts: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/{spec,plan,acceptance,progress}.md`
  (Tier M — 4 files).
- Audit lens: `--deep` (dispatch-mandated). Iter-2 carries four operator
  scope directives (record: `progress.md` §D).
- Card: t1370. Related landed work: SPEC-RELATION-PICKUP-FILTER-001 (t1343,
  completed — pickup filter, now OUT of this card's coverage), SPEC-MANAGER-TODO-001
  (the queue cycle — out of coverage, precedent only), SPEC-INFINITE-GOAL-001
  (goal boundedness — boundary surface), SPEC-CODEX-SESSION-MSG-001 (the
  broker tools the codex ⑤ substitute polls), SPEC-LSEL-DRAIN-STALL-001
  (adjacent prior stall-recovery card — LSEL drain trigger, different
  subsystem).
- Implementation route: worktree-local commits on `WT-lane-stall-watchdog`;
  integration via the serial develop window per the lane protocol. NO push
  from this lane.

## §B Known Issues (Tier M filtered set)

- **B2 — cross-SPEC policy pre-scan**: landed surfaces this card must not
  disturb — t1343's pickup filter (`todo_auto.go`), t1306's manager-todo
  cycle, t1309/t1343 relation findings, t1326's todo surface guard,
  SPEC-CODEX-SESSION-MSG-001's broker. The queue surface is now OUT of
  coverage entirely; the only queue-adjacent act is reading it as precedent.
- **B4 — frontmatter canonical schema**: canonical 12 fields; snake_case
  aliases produce `FrontmatterInvalid`; sibling artifacts carry no `status:`.
- **B6 — spec-lint heading convention**: exclusions under
  `### Out of Scope — <topic>` H3 with `-` bullets; the acceptance table's
  REQ column header must stay requirement-named (`REQ`) for the coverage
  sibling reader (`lint_coverage_sibling_table.go` header pattern — measured
  during iter-1 lint repair).
- **B8 — working-tree hygiene**: no `.moai/state/`, `.moai/harness/`, home
  queue DB, or runtime-managed writes; commit by explicit pathspec.
- **B10 — PRESERVE list**: everything outside spec.md §A.4's enumerated
  touch surface — including `internal/**`, `.moai/config/**`,
  `.claude/loop.md` (+ mirror), all other skills/rules, `factory decide`
  code paths, and template-mirrored doctrine files not named in §A.4/§A.7.
- **B11 — blocker reports**: any needed decision not in this plan returns a
  structured blocker; no AskUserQuestion, no prose questions.
- **B13 — always-loaded rule edits (cache directive 3)**: the §A.7 amendment
  list touches files loaded into every session prefix
  (`orchestration-mode-selection.md`, `askuser-protocol.md`, `spec-workflow.md`,
  `goal-directive.md`, `session-handoff.md`) — batch these edits at M3's END,
  immediately before the M4 verification boundary, never sprinkled.

## §C Pre-flight

```bash
git branch --show-current                       # WT-lane-stall-watchdog
git rev-parse --short HEAD                      # revision commit at plan close
go build ./...                                  # baseline green
golangci-lint run --timeout=2m 2>&1 | tail -5   # baseline (no Go changes expected)
ls internal/template/templates/.claude/skills/ | grep foreman    # mirror precedent present
go run ./cmd/moai spec lint SPEC-LANE-STALL-WATCHDOG-001          # exit 0
go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1  # neutrality baseline
grep -c "jev_ask" internal/cli/mcp_server.go    # > 0 — jev_ask tool present
ls scripts/jev/ 2>&1                            # absent — stale-surface guard
```

## §D Constraints

1. **Template-First**: every NEW/EDITED file under `.claude/` lands in
   `internal/template/templates/` first, then `make build`, then the local
   copy — one content, two trees, byte-identical (AC-LSW-010a/b). Local-only
   exception: `.claude/rules/local/gitflow-lane-protocol.md` is NEVER
   mirrored.
2. **Template neutrality**: template copies carry no card ids (`t\d+`), no
   internal dates, no commit SHAs; composition references to internal SPEC
   IDs stay in the local layer and `.moai/specs/` (local-only). Run
   `go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1`
   before commit; consult the §25.1 checklist when in doubt.
3. **No Go code changes** — `git diff --name-only develop...HEAD -- '*.go'`
   stays empty through close (AC-LSW-011); this holds THROUGH the autonomous
   transition because decision records go to the disk board, not to
   `factory decide` (spec.md §A.7 row E). A discovered Go need returns a
   blocker, never a silent expansion.
4. **No new config** — no `.moai/config/sections/` file or key; the
   watchdog's N is a per-invocation parameter; `workflow.jev.enabled` already
   exists and stays default-false.
5. **Lane discipline** — no push, no `moai integration` commands, no queue
   mutation, commits only on this branch by explicit pathspec, Conventional
   Commits with the `card t1370` marker and the `Authored-By-Agent` trailer,
   no backticks in commit messages (use `git commit -F <tmpfile>`).
6. **Keep-set semantics frozen** — environment-impossible, operator-held,
   and irreversible external-shared operations keep their current mechanics;
   the amendment list (§A.7) changes DOCUMENT wordings and default paths,
   never the three KEEP categories' protections.
7. **Stale-surface guard** — no artifact references `scripts/jev/` or an
   `ask.sh` Jev surface (measured absent); the only Jev surface named is the
   `jev_ask` MCP tool behind `workflow.jev.enabled`.

## §E Self-Verification Deliverables

- **E1** AC PASS/FAIL matrix over `acceptance.md` §D, each row naming command
  + verbatim output + baseline (this run, this tree, HEAD SHA).
- **E2** Cross-platform build: `go build ./...` and
  `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- **E3** Coverage: N/A — zero Go changes (explicit; substitute gates are
  E4+E5).
- **E4** Boundary: no AskUserQuestion in any touched file
  (`grep -rn 'AskUserQuestion' .claude/skills/moai-lane-watchdog/
  .claude/rules/moai/workflow/auto-semantics.md` → 0 matches).
- **E5** `make build` exit 0 after template edits; template neutrality scoped
  test green; spec lint exit 0.
- **E6** Branch HEAD + commit list; no push (report the unpushed state).
- **E7** Blocker report if any decision was missing.
- **E8** RED evidence: `acceptance.md` §E ledger cells (observed at
  `3dd5adf2f` for the 0.1.0 set, re-observed at the revision tree for the
  iter-2 set); each flipped criterion cites its cell.
- **E9** Amendment sweep: after M3, `grep -c "Implementation Kickoff
  Approval"` across the §A.7 row-A..D files reads consistent with the new
  default (each file's diff shown); `factory decide` help text unchanged
  (`go run ./cmd/moai factory decide --help` — human-only wording preserved).

## §F Milestones (ordered by decision-reversibility — changeable decisions first)

- **M1 (Priority High) — unified lane-centered `--auto` document.**
  Template-first author
  `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`,
  `make build`, local copy. Content: the common definition; the lane-surface
  definition of `--auto`; boundary inventories for `todo --auto` and
  `goal --auto` (precedent citation only); the decision ladder (§B.2) with
  per-step carriers; the harness-neutrality table (§B.6) incl. the codex ⑤
  broker substitute and `update_plan` view; the gate inventory + dispositions
  + keep-set + out-of-set justification duty (§B.5); the decision-record
  format (§B.8); the view–SSOT rule (§B.7); the Jev boundary (§B.2 step ④);
  the awaken rule + canonical awaken prompt (§B.4). Flips AC-LSW-001a,
  001b, 002, 013a, 014, 015, 016a.
- **M2 (Priority High) — `moai-lane-watchdog` skill.** Template-first author
  `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`,
  `make build`, local copy. Content: one watchdog iteration — progress
  measurement (REQ-LSW-001), cause classification (awaited-judgment /
  blocked-by / shell-error / accidental-stop), the remedy + ladder procedure
  (REQ-LSW-003..006) with per-runner paths, the decision-board read/write
  protocol with the view–SSOT rule (REQ-LSW-011), decision-record emission
  (REQ-LSW-012), the t1343 composition + queue-readonly note, escalate-or-record
  output shape. Flips AC-LSW-003, 004a-d, 005, 006, 013b, 016b, 017a, 017b.
- **M3 (Priority Medium) — doctrine integration + amendment sweep.** (a)
  `.claude/rules/moai/workflow/kanban-dispatch.md` (+ mirror): explicit-wait
  posture + watchdog pointer, batch-authorization promotion wording (§A.7
  row F). (b) `.claude/rules/local/gitflow-lane-protocol.md` §6: the
  open-ended wait becomes explicit wait + ladder recheck (local-only). (c)
  The §A.7 amendment list rows A-D, G, H: apply the worded amendments to
  `orchestration-mode-selection.md`, `askuser-protocol.md`,
  `spec-workflow.md`, `run.md`/`goal-directive.md` surfaces,
  `session-handoff.md`, `AGENTS.local.md` §29 pointer area,
  `contract-autonomy.md` — batched at the END of M3 (B13). Flips AC-LSW-007a,
  007b, 009.
- **M4 (Priority Medium) — mechanical verification sweep.** `make build`;
  mirror parity `diff -r` (AC-LSW-010a/b); template neutrality + spec lint
  re-run (AC-LSW-012); Go-scope guard (AC-LSW-011); E9 amendment sweep;
  codex-premise re-check of R-4 against current Codex docs; update
  progress.md §E.2/§E.3.

cycle_type=tdd — the acceptance commands ARE the RED probes (observed red at
`3dd5adf2f` and re-observed at the revision tree; ledger in acceptance.md §E);
each milestone flips its criteria green and re-runs them. No Go package
changes exist to coverage-measure, so the TDD loop is artifact-probe-driven —
the honest application of the repo's `development_mode: tdd` (quality.yaml:2)
to a doc+skill card.

## §G Anti-Patterns (refuse during run)

- Rewriting t1343's filter, the queue cycle, or any queue mechanic "while in
  the file" — the queue surface is out of coverage.
- Letting template copies drift from local copies (the diff -r gate exists
  for exactly this).
- Amending the three KEEP categories' protections to "help" automation.
- Naming `scripts/jev/` or `ask.sh` anywhere (stale surface — measured
  absent); the single Jev surface is `jev_ask` behind `workflow.jev.enabled`.
- Treating TaskList or `update_plan` as the decision SSOT (views only —
  §B.7); conflating ExecPlan with the decision board.
- Adding a config key or a Go verb because the skill "would be nicer with
  one"; claiming an "advance" verb exists.
- Leaving a §A.7 amendment row unapplied while its old wording still ships.
- Citing internal card ids inside template-mirrored content; committing with
  sweep staging; pushing the branch.

## §H Cross-References

- SPEC artifacts: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/`
- Composition: `.moai/specs/SPEC-RELATION-PICKUP-FILTER-001/spec.md` (§A.5)
- Broker: `.moai/specs/SPEC-CODEX-SESSION-MSG-001/` (codex ⑤ substitute)
- Surfaces: `internal/cli/todo_auto.go`, `internal/cli/todo.go`,
  `internal/cli/goal.go`, `internal/goal/evaluate.go`,
  `internal/cli/mcp_server.go` (jev_ask :650-658; codex_audit :316; glm_audit
  :493; session_msg_register :550; session_msg_send :566),
  `internal/cli/factory_card.go` (decide :1469-1485),
  `internal/config/types.go` (workflow.jev :522, :783), `.claude/loop.md`,
  `.claude/rules/moai/workflow/kanban-dispatch.md`,
  `.claude/rules/moai/workflow/session-handoff.md`,
  `.claude/rules/local/gitflow-lane-protocol.md`,
  `.claude/skills/moai-kanban-foreman/SKILL.md` (execution-vehicle precedent)
- Doctrine neighbors: `cross-session-messaging.md` (reply-independence),
  `agent-common-protocol.md` (retry ceiling), `runtime-recovery-doctrine.md`
  (ladder + checkpoint return), `contract-autonomy.md` (gate-free entry
  precedent)
