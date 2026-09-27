# Plan — SPEC-FACTORY-SELF-DISPATCH-001 (card t1240)

## §A Context

Base: `WT-factory-self-dispatch` at local develop `ed506740b` (includes F1 t1239, t1242, t1245).
Tier L: seven milestones, more than ten production files, six new MCP tools, three launchers.
v0.2.0 resolves plan-audit iteration 1 (FAIL 0.66, `.moai/reports/t1240/plan-audit-iter1.md`).

## §B Decisions for the Kickoff gate

Each decision carries the chosen default and its rationale; Implementation Kickoff Approval confirms
them. None is open.

- **B1 — Lane selection order** (design.md D3): own assigned → unowned picked → oldest queued with
  promotion; other lanes' cards never. Rationale: operator Q3 in SPEC-ROLE-NAMING-CODE-001 plan.md §B.
- **B2 — Codex engine, integration, and livelock** (D1, D2; plan-audit D1): interactive relaunch per
  card; stop at `merge-ready`; a Codex lane's `next` skips any card at `merge-ready` or later, including
  one returned to `assigned` by lease expiry (REQ-SD-025, AC-SD-023). Rationale: card text, and a
  re-lease would livelock the launcher on a card it cannot advance.
- **B3 — Lane predicate** (D5; plan-audit D5): role marker equals the value constant, on every path; a
  label alone is not a lane. Default keeps the Codex MCP allowlist frozen (REQ-SD-022), accepting that
  `todo_add`/`factory_decide` are not refused on the Codex MCP path (spec.md §E.1). Rationale: one rule
  everywhere; widening the allowlist is a REQ-CFR-020 reversal with user-visible drift and is left to a
  later explicit decision.
- **B4 — Integration surface** (D6; plan-audit D2, O10): a Claude lane acquires `moai integration
  acquire`, merges in the worktree that has the integration branch checked out, records `complete`, and
  releases; the parent checkout never changes branch; `complete` refuses when that worktree is not
  provisioned. Rationale: the repository's existing single integration surface; merging in the parent
  would contradict REQ-SD-018. Local override noted in spec.md §E.1.
- **B5 — Harness signal** (D7; plan-audit D6): `MOAI_KANBAN_BACKEND` with the existing
  `kanban.Backend*` constants; `gpt` identifies Codex. Rationale: existing name (REQ-SD-022 forbids new
  names) and it is in the Codex MCP allowlist.
- **B6 — Queue guard as allowlist** (D4; plan-audit D7, D8): lanes may run bare `todo`, `list`,
  `history`, `why`, `pr`, `triage`; everything else is refused, `todo next <n>` included. Rationale: card
  text "에이전트 권한에서 큐 변경 제외"; t1256 Q3's "the pick path gains none" is a statement about t1256's
  own scope, and REQ-RNC-021's help text is handed to t1257 (spec.md §E.2).
- **B7 — MCP tree resolution** (D10; plan-audit D4): caller-supplied `project_root`, required on
  `factory_next`/`stage`/`complete`; `factory_next`'s parent-checkout check evaluates that argument.
  Branch if the pre-flight measurement (§C.4) shows the server does follow the session tree: nothing
  changes — the argument remains required, because the rule must hold on every harness.
- **B8 — Next-card rule on `startup`** (plan-audit D3): injected on every lane `startup` regardless of
  clear policy, and on `clear`.
- **B9 — Pre-dispatch cross-check** (plan-audit D12, O9): `next` prints the leased card's PR and landed
  state through the `moai todo pr` reader; the next-card rule states that lane promotion is
  operator-authorized.
- **B10 — `--wait` bound and no-card status**: exit status 3 for no card; the wait interval and bound are
  fixed defaults with a flag to change the bound; exact values recorded in progress.md at run.

## §C Pre-flight (run-phase entry)

1. Record the local develop SHA (`git rev-parse develop`) in progress.md §E.2 as `develop_sha:`
   (REQ-SD-001, AC-SD-001).
2. Verify SPEC-ROLE-NAMING-CODE-001 landed on that SHA: `git show develop:internal/config/envkeys.go`
   shows the role-value constant equal to `lane`, and `git show develop:internal/cli/factory.go` shows the
   `-f` role token equal to `lane`. Record `t1256_landed: yes|no`. `no` → halt, blocker report to the
   leader, no production commit.
3. Absorb develop into this branch (`git merge develop` inside the card worktree), then re-read every
   file:line citation in research.md and design.md against the absorbed tree and correct the ones t1256
   moved.
4. Measure research.md §9's open item: does the moai MCP server of a Claude lane follow `EnterWorktree`?
   Record command and output (B7 branch).
5. Measure `codex --help` for the interactive working-directory flag.
6. Characterization run, scoped: `go test ./internal/homestate/... ./internal/kanban/... -count=1`, plus
   `go test ./internal/cli -run '^TestFactoryRoleTokenPinsGuardConstant$' -count=1`.

## §D Overlap with SPEC-ROLE-NAMING-CODE-001 (card t1256) and run order

### D.1 What t1256 has changed so far

`git diff --stat ed506740b...WT-role-naming-code` (measured at plan time):

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

No code has changed on that branch yet; the overlap is with its **planned** files, taken from the Go
paths its committed spec/plan/research/design/acceptance name.

### D.2 Files this run will touch, against t1256

| File | This SPEC's change | In t1256's plan? |
|---|---|---|
| `internal/cli/factory.go` | `-f lane` stamps marker and backend; clear-policy value; relaunch hook-in | **Yes** (M2; named 4×) |
| `internal/cli/codex_launcher.go` | accept `-f lane`, per-card relaunch, child env, REQ-SD-004 line | **Yes** (named; M2 help text) |
| `internal/cli/todo.go` | lane allowlist guard; promotion helper reused by `next` | **Yes** (REQ-RNC-021 pick help; owner reader) |
| `internal/config/envkeys.go` | read-only use of the marker and backend constants | **Yes** (M4 flips the value constant) |
| `internal/hook/session_start_factory.go` | next-card rule on `startup`/`clear` | **Yes** (M3 notices) |
| `internal/hook/session_start_factory_i18n.go` | rule text en/ko/ja/zh | **Yes** (M3 locales) |
| `internal/kanban/bootstrap.go`, `internal/kanban/factory_slots.go` | read only (label producer, roster) | **Yes** — read dependency |
| `internal/cli/factory_role_pin_test.go` | extended to assert the stamp uses the constant | **Yes** (M4 pin) |
| `internal/cli/cc.go` | export backend on the factory-lane path | No |
| `internal/cli/factory_card.go` | `next`, `stage`, `complete`; lane refusal on `decide` | No |
| `internal/cli/factory_handoff_recover.go` | register the three verbs | No |
| `internal/cli/integration.go` | read only (target resolution reused by `complete`) | No |
| `internal/cli/mcp_server.go` + new `mcp_factory_card.go`, `mcp_todo.go` | six tools with `project_root` | No |
| `internal/hook/session_start.go` | wire the rule | No |
| `internal/cli/launch_exec_posix.go`, `launch_exec_windows.go` | supervising form for relaunch | No |
| `internal/homestate/card_*.go` | at most a selection query helper | No |

### D.3 Run order and the landing criterion

1. t1256 run lands on local develop (leader's merge window).
2. This branch absorbs develop and re-measures its citations (§C.3).
3. This SPEC's run starts.

**Landing criterion (plan-audit O3).** The gate for t1256 is REQ-SD-001's tree check — the role-value
constant and the `-f` role token both equal `lane` on the develop SHA read at run start — not
SPEC-ROLE-NAMING-CODE-001's `status:`. That is why `depends_on` lists only SPEC-FACTORY-RECORD-001 (already
`completed`): a t1256 that has merged its run but not yet synced would otherwise block the Depends_on
pre-flight although the code this SPEC needs is present.

## §E Self-verification

Scoped remeasure after every develop absorption and before the merge-window request:
`go test ./internal/homestate/... ./internal/kanban/... ./internal/config/... ./internal/spec/... -count=1`
and, for `./internal/cli`, `./internal/hook`, and `./internal/codexwiring`, only the anchored AC tests
(acceptance.md §B). Lint with the CI golangci-lint version.

## §F Milestones (ordered by change likelihood — highest first)

### M1 — Lane predicate, selection, permission boundary (Priority High)
REQ-SD-008..010, -015, -016, -025 (selection half). ACs: AC-SD-008, -009, -010, -015, -016, -023.

### M2 — Integration surface and Codex merge-edge refusal (Priority High)
REQ-SD-013, -023, -025 (edge half). ACs: AC-SD-013, -024, -025.

### M3 — Worktree per card, `stage`, MCP tools (Priority High)
REQ-SD-011, -012, -014, -024. ACs: AC-SD-011, -012, -014.

### M4 — Launchers: cc/glm lane, marker and backend stamp, git requirement, no remote (Priority Medium)
REQ-SD-002, -005, -006, -007, -017. ACs: AC-SD-002, -005, -006, -007, -017.

### M5 — Codex relaunch and refusal wording (Priority Medium)
REQ-SD-003, -004. ACs: AC-SD-003, -004.

### M6 — SessionStart next-card rule and clear policy (Priority Medium)
REQ-SD-019, -020. ACs: AC-SD-019, -020.

### M7 — Invariants, vocabulary, lifecycle records (Priority Low)
REQ-SD-001, -018, -021, -022. Sync phase adds `partially_superseded_by` on
SPEC-CODEX-FACTORY-RETIRE-001 (manager-spec by re-delegation). ACs: AC-SD-001, -018, -021, -022.

## §G Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | t1256 moves many cited lines | §C.3 re-reads every citation after absorption |
| R2 | MCP server in a Claude lane does not follow `EnterWorktree` | REQ-SD-024 makes the caller name its tree; §C.4 records the measurement |
| R3 | A lane unsets the role marker | Residual (spec.md §E.1); the lane then loses `next`/`stage`/`complete` |
| R4 | Queue promotion and the pre-dispatch cross-check contradict documented HARD clauses until t1257 | spec.md §E.2 hand-off list; `next` prints PR/landed state |
| R5 | Codex `merge-ready` cards need an integrator | REQ-SD-025 stops the livelock; integration is F3's |
| R6 | Interactive Codex flag spelling differs from `-C` | §C.5 measures it before M5 |
| R7 | Local-repo overrides (`CLAUDE.local.md` §4.1, `gitflow-lane-protocol.md` §6) contradict REQ-SD-008/-023 here | spec.md §E.1; operator decision before F2 runs in this repository |

## §H Cross-references

SPEC-FACTORY-RECORD-001 (F1 record and transition API); SPEC-AUTONOMY-PRECONDITION-001 (marker
constants and guard); SPEC-CODEX-FACTORY-RETIRE-001 (refusal seam, frozen allowlist);
SPEC-ROLE-NAMING-CODE-001 (vocabulary; must land first).
