# Plan — SPEC-FACTORY-SELF-DISPATCH-001 (card t1240)

## §A Context

Base: `WT-factory-self-dispatch` at local develop `ed506740b` (includes F1 t1239, t1242, t1245).
Tier L: seven milestones, more than ten production files, six new MCP tools, three launchers.

## §B Decisions for the Kickoff gate

All decisions below carry a recommended default; Implementation Kickoff Approval confirms them.

- **B1 — Lane selection order** (design.md D3): own assigned → unowned picked → oldest queued with
  promotion. Basis: operator Q3 recorded in SPEC-ROLE-NAMING-CODE-001 plan.md §B.
- **B2 — Codex engine and integration** (D1, D2): interactive relaunch per card, stop at `merge-ready`.
  Basis: card text; overrides t1242 design.md §5 ("headless `codex exec` worker").
- **B3 — MCP lane check reads the lane label too** (D5), keeping the Codex allowlist frozen.
- **B4 — Integration branch = configured worktree base branch** (D6).
- **B5 — `--wait` bound**: a fixed default wait bound with an explicit flag to change it; exact values
  are a run-phase choice recorded in progress.md.

## §C Pre-flight (run-phase entry)

1. Record the local develop SHA (`git rev-parse develop`) in progress.md (REQ-SD-001).
2. Verify SPEC-ROLE-NAMING-CODE-001 landed on that SHA: `git show develop:internal/config/envkeys.go`
   shows the role-value constant equal to `lane`, and `git show develop:internal/cli/factory.go` shows
   the `-f` role token equal to `lane`. Either missing → halt, blocker report to the leader, no
   production edit.
3. Absorb develop into this branch (`git merge develop` inside the card worktree), then re-read every
   file:line citation in research.md against the absorbed tree and correct the ones t1256 moved.
4. Measure the Gap in research.md §9: does the moai MCP server of a Claude lane still resolve the same
   factory database after `EnterWorktree`? Record command and output.
5. Measure `codex --help` for the interactive working-directory flag (research.md §9).
6. Characterization run, scoped: `go test ./internal/homestate/... ./internal/kanban/... -count=1`,
   plus `go test ./internal/cli -run '^TestFactoryRoleTokenPinsGuardConstant$' -count=1`.

## §D Overlap with SPEC-ROLE-NAMING-CODE-001 (card t1256) and run order

### D.1 What t1256 has changed so far

`git diff --stat ed506740b...WT-role-naming-code` (this run):

```
 .../specs/SPEC-ROLE-NAMING-CODE-001/acceptance.md  | 121 ++++++++++++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py    | 140 +++++++++++++++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/design.md    |  81 ++++++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/plan.md      | 118 +++++++++++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/progress.md  |  54 ++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/research.md  |  81 ++++++++++++
 .moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md      | 104 +++++++++++++++
 7 files changed, 699 insertions(+)
```

No code has changed yet on that branch; the overlap is with its **planned** files, taken from the Go
paths its committed spec/plan/research/design/acceptance name.

### D.2 Files this run will touch, against t1256

| File | This SPEC's change | In t1256's plan? |
|---|---|---|
| `internal/cli/factory.go` | `-f lane` launch stamps the marker; clear-policy value; relaunch loop hook-in | **Yes** (M2 token/label rename; named 4×) |
| `internal/cli/codex_launcher.go` | accept `-f lane`, per-card relaunch loop, child env | **Yes** (named; M2 help text) |
| `internal/cli/todo.go` | lane refusal on queue mutators; promotion helper reused by `next` | **Yes** (REQ-RNC-021 pick help text; owner reader) |
| `internal/config/envkeys.go` | read-only use of the marker constants; possibly a clear-policy key | **Yes** (M4 flips the value constant) |
| `internal/hook/session_start_factory.go` | next-card rule on `clear` | **Yes** (M3 notices) |
| `internal/hook/session_start_factory_i18n.go` | rule text en/ko/ja/zh | **Yes** (M3 locales) |
| `internal/kanban/bootstrap.go` | read only (`FactoryLaneLabel`) | **Yes** (label prefix) — read dependency |
| `internal/kanban/factory_slots.go` | read only (lane claim/roster) | **Yes** — read dependency |
| `internal/cli/factory_role_pin_test.go` | extended to assert the stamp uses the constant | **Yes** (M4 pin) |
| `internal/cli/factory_card.go` | `next`, `stage`, `complete`; lane refusal on `decide` | No |
| `internal/cli/factory_handoff_recover.go` | register the three verbs | No |
| `internal/cli/mcp_server.go` + new `mcp_factory_card.go`, `mcp_todo.go` | six tools | No |
| `internal/hook/session_start.go` | wire the clear-source rule | No |
| `internal/cli/launch_exec_posix.go`, `launch_exec_windows.go` | supervising form for relaunch | No |
| `internal/homestate/card_*.go` | at most a selection query helper | No |

### D.3 Run order

1. t1256 run lands on local develop (leader's merge window).
2. This branch absorbs develop (`git merge develop` in the card worktree) and re-measures its citations
   (§C.3).
3. This SPEC's run starts. REQ-SD-001 halts it otherwise.

## §E Self-verification

Scoped remeasure after every develop absorption and before the merge-window request:
`go test ./internal/homestate/... ./internal/kanban/... ./internal/config/... ./internal/spec/... -count=1`
and, for `./internal/cli` and `./internal/hook`, only `-run`-filtered AC tests (acceptance.md §B).
Lint with the CI golangci-lint version.

## §F Milestones (ordered by change likelihood — highest first)

### M1 — Lane selection and the permission boundary (Priority High)
`next` selection order, queue promotion, `--wait`, parent-only check (REQ-SD-008..010); lane refusals on
queue mutators and `decide` (REQ-SD-015, -016). ACs: AC-SD-008..010, -015, -016.

### M2 — Worktree per card and the verbs `stage` / `complete` (Priority High)
REQ-SD-011..013. ACs: AC-SD-011..013.

### M3 — MCP tools (Priority High)
Six tools on the shared implementation (REQ-SD-014); MCP-path lane check (design.md §5). AC-SD-014.

### M4 — Launchers: cc/glm lane, marker stamp, git requirement, no remote (Priority Medium)
REQ-SD-002, -005, -006, -007, -017. ACs: AC-SD-002, -005, -006, -007, -017.

### M5 — Codex relaunch (Priority Medium)
REQ-SD-003, -004. ACs: AC-SD-003, -004.

### M6 — SessionStart next-card rule and clear policy (Priority Medium)
REQ-SD-019, -020. ACs: AC-SD-019, -020.

### M7 — Invariants, vocabulary, lifecycle records (Priority Low)
REQ-SD-018, -021, -022; run-gate REQ-SD-001. Sync phase adds `partially_superseded_by` on
SPEC-CODEX-FACTORY-RETIRE-001 (manager-spec by re-delegation). ACs: AC-SD-001, -018, -021, -022.

## §G Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | t1256 moves many of the cited lines | §C.3 re-reads every citation after absorption |
| R2 | MCP server in a Claude lane does not follow `EnterWorktree` | §C.4 measures it; the verbs resolve the factory database from the project key, not the cwd |
| R3 | A lane unsets the role marker | Residual (spec.md §E); CLI refusal also checks the lane label |
| R4 | Queue promotion by lanes contradicts the documented HARD clause until t1257 | Recorded in spec.md §E; leader informed at Kickoff |
| R5 | `merge-ready` lease expiry bounces Codex cards back to `assigned` | Residual; F3 owns integration of those cards |
| R6 | Interactive Codex flag spelling differs from `-C` | §C.5 measures it before M5 |

## §H Cross-references

SPEC-FACTORY-RECORD-001 (F1 record and transition API); SPEC-AUTONOMY-PRECONDITION-001 (marker
constants and guard); SPEC-CODEX-FACTORY-RETIRE-001 (refusal seam, frozen allowlist);
SPEC-ROLE-NAMING-CODE-001 (vocabulary; must land first).
