# plan.md — SPEC-LANE-STALL-WATCHDOG-001

## §A Context

- Worktree: `.moai/worktrees/t1370` (absolute:
  `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1370`), branch
  `WT-lane-stall-watchdog`, plan-phase HEAD `3dd5adf2f`.
- SPEC artifacts: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/{spec,plan,acceptance,progress}.md`
  (Tier M — 4 files).
- Audit lens: `--deep` (dispatch-mandated).
- Card: t1370 (operator directive 2026-09-30, worker-70 배차). Related landed
  work: SPEC-RELATION-PICKUP-FILTER-001 (t1343, completed — pickup filter in
  `internal/cli/todo_auto.go:162-187`), SPEC-MANAGER-TODO-001 (the `--auto`
  cycle), SPEC-INFINITE-GOAL-001 (goal boundedness), SPEC-LSEL-DRAIN-STALL-001
  (adjacent prior stall-recovery card — LSEL drain trigger, a different
  subsystem; no shared surface).
- Implementation route: worktree-local commits on `WT-lane-stall-watchdog`;
  integration via the serial develop window per the lane protocol. NO push
  from this lane.

## §B Known Issues (Tier M filtered set)

- **B2 — cross-SPEC policy pre-scan**: the affected surfaces carry landed
  SPEC state — t1343's pickup filter (`todo_auto.go`), t1306's manager-todo
  cycle, t1309/t1343 relation findings, t1326's todo surface guard. The run
  MUST NOT modify any of those behaviors; the only `todo_auto.go`-adjacent
  act is reading it as evidence. The unified doc states the inventory; it
  does not restate or rewrite the queue mechanics.
- **B4 — frontmatter canonical schema**: spec.md uses the canonical 12 fields
  (`created:`/`updated:`/`tags:`); snake_case aliases produce
  `FrontmatterInvalid`. Sibling artifacts carry no `status:` field.
- **B6 — spec-lint heading convention**: exclusions live under
  `### Out of Scope — <topic>` H3 headings with `-` bullets (spec.md §D
  complies; keep it that way through any revision).
- **B8 — working-tree hygiene**: do not touch `.moai/state/`,
  `.moai/harness/`, the home queue DB, or runtime-managed files; commit by
  explicit pathspec; no sweep staging.
- **B10 — PRESERVE list**: everything outside §A.4's enumerated touch surface
  is PRESERVED — including `internal/**`, `.moai/config/**`, `.claude/loop.md`
  (+ its template mirror), all other skills/rules, and both template-mirrored
  doctrine files not named in §A.4.
- **B11 — blocker reports**: any needed decision not in this plan returns a
  structured blocker; no AskUserQuestion, no prose questions.

## §C Pre-flight

```bash
git -C <worktree> branch --show-current        # WT-lane-stall-watchdog
git -C <worktree> rev-parse --short HEAD       # 3dd5adf2f at plan close
go build ./...                                 # baseline green
golangci-lint run --timeout=2m 2>&1 | tail -5  # baseline (no Go changes expected)
ls internal/template/templates/.claude/skills/ | grep foreman   # mirror precedent present
go run ./cmd/moai spec lint SPEC-LANE-STALL-WATCHDOG-001         # plan-phase exit 0
go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1  # neutrality baseline
```

## §D Constraints

1. **Template-First**: every NEW/EDITED file under `.claude/` lands in
   `internal/template/templates/` first, then `make build`, then the local
   copy — one content, two trees, byte-identical (AC-LSW-010). Local-only
   exception: `.claude/rules/local/gitflow-lane-protocol.md` is NEVER
   mirrored (`AGENTS.local.md` §2a).
2. **Template neutrality**: template copies carry no card ids (`t\d+`), no
   internal dates, no commit SHAs; composition references to internal SPEC
   IDs stay in the local layer and `.moai/specs/` (local-only). Run the
   neutrality guard (`go test ./internal/template/ -run '^TestTemplateNeutralityAudit$' -count=1`
   or the CI-equivalent) before commit; consult the §25.1 checklist when in
   doubt.
3. **No Go code changes** — `git diff --name-only develop...HEAD -- '*.go'`
   stays empty through close (AC-LSW-011); a discovered need returns a
   blocker, never a silent scope expansion.
4. **No new config** — no `.moai/config/sections/` file or key; the
   watchdog's N is a per-invocation parameter documented in the skill.
5. **Lane discipline** — no push, no `moai integration` commands, no queue
   mutation, commits only on this branch by explicit pathspec, Conventional
   Commits with the `card t1370` marker and the `Authored-By-Agent` trailer,
   no backticks in commit messages (use `git commit -F <tmpfile>`).
6. **Gate semantics frozen** — the doctrine edits ADD an explicit-wait
   posture; they must not weaken, reword, or re-route any gate's decision
   semantics (Kickoff, sync blocking, operator queue gates).

## §E Self-Verification Deliverables

- **E1** AC PASS/FAIL matrix over `acceptance.md` §D, each row naming command
  + verbatim output + baseline (this run, this tree, HEAD SHA).
- **E2** Cross-platform build: `go build ./...` and
  `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (no Go changes; the
  build proves the tree still compiles with the template edits embedded).
- **E3** Coverage: N/A — zero Go changes (state this explicitly; the
  substitute quality gate is E4+E5: template neutrality tests + spec lint).
- **E4** Boundary: no AskUserQuestion in any touched file
  (`grep -rn 'AskUserQuestion' .claude/skills/moai-lane-watchdog/
  .claude/rules/moai/workflow/auto-semantics.md` → 0 matches).
- **E5** `make build` exit 0 after template edits; `go test
  ./internal/template/... -count=1` scoped run green; spec lint exit 0.
- **E6** Branch HEAD + commit list; no push (report the unpushed state).
- **E7** Blocker report if any decision was missing.
- **E8** RED evidence: the pre-implementation observations of
  `acceptance.md` §E (ledger) serve as the RED outputs — each flipped
  criterion cites its ledger cell.

## §F Milestones (ordered by decision-reversibility — changeable decisions first)

- **M1 (Priority High) — unified `--auto` semantics document.** Template-first
  author `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`,
  `make build`, local copy at `.claude/rules/moai/workflow/auto-semantics.md`.
  Content: the common definition ("minimize human intervention through
  autonomous adjudication"), the three-surface inventory (spec.md §A.3),
  the common invariants (gates remain gates; queue production stays the
  operator's; Jev display-only, REQ-LSW-009), the awaken rule + canonical
  lane-awaken prompt (§B.4), and the explicit-wait protocol (§B.5).
  Flips AC-LSW-001, AC-LSW-002.
- **M2 (Priority High) — `moai-lane-watchdog` skill.** Template-first author
  `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`,
  `make build`, local copy. Content: one watchdog iteration — progress
  measurement on the three channels (REQ-LSW-001), cause classification into
  the four labels (awaited-actor / blocked-by / shell-error / accidental-stop),
  the per-cause remedy table (spec.md §B.2), the t1343 composition note +
  queue-readonly prohibition (REQ-LSW-004), escalate-or-record output shape.
  Flips AC-LSW-003, AC-LSW-004a-d, AC-LSW-005, AC-LSW-006.
- **M3 (Priority Medium) — doctrine integration edits.** (a)
  `.claude/rules/moai/workflow/kanban-dispatch.md` (+ template mirror): add
  the explicit-wait posture and the watchdog pointer to the lane-completion
  section, phrased to add a waiting protocol without touching the completion
  or verdict rules. (b) `.claude/rules/local/gitflow-lane-protocol.md` §6:
  amend the open-ended "리더가 다음 카드를 dispatch 할 때까지 기다린다"
  posture to the explicit-wait + re-check protocol (local-only, no mirror).
  Flips AC-LSW-007, AC-LSW-008, AC-LSW-009.
- **M4 (Priority Medium) — mechanical verification sweep.** `make build`;
  mirror parity `diff -r` on both new artifacts (AC-LSW-010a/b); template
  neutrality + scoped template tests (E5); spec lint re-run (AC-LSW-012);
  Go-scope guard re-measure (AC-LSW-011); update progress.md §E.2/§E.3.

cycle_type=tdd — the acceptance commands ARE the RED probes (all observed red
on `3dd5adf2f`, ledger in acceptance.md §E); each milestone flips its criteria
green and re-runs them. No Go package changes exist to coverage-measure, so
the TDD loop is artifact-probe-driven; this is the honest application of the
repo's `development_mode: tdd` (quality.yaml:2) to a doc+skill card.

## §G Anti-Patterns (refuse during run)

- Rewriting t1343's filter or any queue mechanic "while in the file".
- Letting template copies drift from local copies (the diff -r gate exists
  for exactly this).
- Weakening a gate's wording to "help" the explicit-wait protocol.
- Adding a config key or a Go verb because the skill "would be nicer with
  one".
- Citing internal card ids inside template-mirrored content.
- Committing with sweep staging or pushing the branch.

## §H Cross-References

- SPEC artifacts: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/`
- Composition: `.moai/specs/SPEC-RELATION-PICKUP-FILTER-001/spec.md` (§A.5)
- Surfaces: `internal/cli/todo_auto.go`, `internal/cli/todo.go`,
  `internal/cli/goal.go`, `internal/goal/evaluate.go`,
  `internal/goal/dashboard.go`, `.claude/loop.md`,
  `.claude/rules/moai/workflow/kanban-dispatch.md`,
  `.claude/rules/moai/workflow/session-handoff.md`,
  `.claude/rules/local/gitflow-lane-protocol.md`,
  `.claude/skills/moai-kanban-foreman/SKILL.md` (execution-vehicle precedent)
- Doctrine neighbors: `cross-session-messaging.md` (reply-independence),
  `agent-common-protocol.md` (retry ceiling), `runtime-recovery-doctrine.md`
  (ladder + checkpoint return)
