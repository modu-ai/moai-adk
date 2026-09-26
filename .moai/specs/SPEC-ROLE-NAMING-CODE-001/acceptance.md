---
id: SPEC-ROLE-NAMING-CODE-001
title: "Acceptance — role naming unification (code + CLI)"
version: "0.2.0"
created: 2026-09-26
---

# Acceptance — SPEC-ROLE-NAMING-CODE-001

Every criterion is binary. "Refused" means: no session launched, no registry / broker / role-declaration record written (row counts and file listings equal before and after), exactly one error line naming the canonical form, and a non-zero exit — the same status the launcher returns for any other invalid `-f` value. Test names are the run phase's to choose; each AC names the package that must hold its test.

## §A Acceptance criteria

### CLI tokens and labels (M2)

- **AC-RNC-001** (maps REQ-RNC-002) — **Given** a running factory with no live lane claims, **When** a lane joins with `-f lane`, **Then** the session launches as `lane-1`, the registry row's label is `lane-1`, and exit is 0. Package: `internal/cli`.
- **AC-RNC-002** (maps REQ-RNC-003, REQ-RNC-020) — **Given** the same factory, **When** a join runs with `-f worker`, and separately with `-f agent`, **Then** each is refused, its error line contains `-f lane`, and the registry row count is unchanged. Package: `internal/cli`.
- **AC-RNC-003** (maps REQ-RNC-004, REQ-RNC-022) — **Given** a live registry claim `worker-3` in the running factory (written by a pre-change binary), **When** `-f lane` runs, **Then** it is refused with a message naming `worker-3`, the run id, and the retire-and-relaunch step, and no `lane-<n>` claim is written; **and Given** the same claim whose pid is dead, **When** `-f lane` runs, **Then** it succeeds exactly as with no claim. Package: `internal/kanban` + `internal/cli`.
- **AC-RNC-004** (maps REQ-RNC-004) — **Given** the post-change tree, **When** the factory label composer is called with 3, **Then** it returns `lane-3`; **and When** the lane-label parser receives `lane-3`, `worker-3`, and `agent-3`, **Then** it accepts only `lane-3` and reports the other two as not-a-lane-label. Package: `internal/kanban`.
- **AC-RNC-005** (maps REQ-RNC-005, REQ-RNC-020) — **Given** no live claims, **When** a lane launch runs with `-f worker-2` and separately with `--name agent-5`, **Then** each is refused and the error lines contain `lane-2` and `lane-5` respectively. Package: `internal/cli`.
- **AC-RNC-016** (maps REQ-RNC-001) — **Given** the built binary, **When** `moai cc --help` and `moai glm --help` run, **Then** each output contains `-f lane` and `-f lane-<n>` and contains neither `-f worker` nor `-f agent`; and the `-f` usage error text contains `lane` and neither `worker` nor `lead`. Package: `internal/cli`.

### Leader label and persisted state (M1)

- **AC-RNC-006** (maps REQ-RNC-006) — **Given** no live leader, **When** a kanban leader and, in another project, a factory leader launch without `--name`, **Then** both sessions are named `leader`; a second concurrent leader in the same project is named `leader-2`. Package: `internal/cli` + `internal/kanban`.
- **AC-RNC-007** (maps REQ-RNC-007, REQ-RNC-020) — **Given** no live leader, **When** a leader launch runs with `--name lead`, and separately with `--name lead-7`, **Then** each is refused, the error lines contain `leader` and `leader-7` respectively, and `leads.json` is byte-identical before and after. Package: `internal/cli`.
- **AC-RNC-008** (maps REQ-RNC-009, REQ-RNC-010) — **Given** a fresh run, **When** a leader launches and a lane joins, **Then** the role declaration, broker `peers.role`, `peers.slot`, registry label, and a factory card owner written by each carry `leader` / `leader` for the leader and `lane` / `lane-1` / `lane-1` for the lane, and no written value equals `lead`, `worker`, or `agent`; **and When** the board write guard, the SessionStart role reader, and the run-retire owner lookup read a role declaration whose role is `leader` and a `peers` row with role `leader`, slot `leader`, **Then** each recognizes the leader. Package: `internal/kanban`, `internal/hook`, `internal/factorymsg`, `internal/cli`.
- **AC-RNC-009** (maps REQ-RNC-013) — **Given** broker peers `leader`/`leader` and `lane`/`lane-1`, **When** `factory_msg_send` targets `to_slot` = `leader` and `lane-1`, **Then** each is delivered to that peer; **and When** it targets `lead`, `worker-1`, `agent-1`, `worker`, and `agent` in turn, **Then** each returns an error naming the canonical slot (`leader` or `lane-1` / `lane`) and the broker message count is unchanged. Package: `internal/factorymsg` + `internal/cli`.
- **AC-RNC-010** (maps REQ-RNC-008) — **Given** develop at the merge-base, **When** `git diff <merge-base> -- internal/homestate internal/factorymsg` is filtered to changed lines containing `CREATE TABLE`, `ALTER TABLE`, or `CREATE INDEX`/`CREATE UNIQUE INDEX`, **Then** the filter prints 0 lines. Recorded in progress.md.
- **AC-RNC-011** (maps REQ-RNC-011) — **Given** the post-change tree, **When** the four environment variable name constants and the generated Codex `env_vars` value are read, **Then** they equal `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, and the develop-merge-base allowlist string byte for byte, and `grep -rnE '"MOAI_(KANBAN_LEADER|FACTORY_LANE)' internal --include='*.go'` returns 0 rows; **and When** a lane joins, **Then** the child environment's `MOAI_FACTORY_WORKER` value is `lane-<n>` and a leader's `MOAI_KANBAN_LEAD_NAME` value is `leader[-…]`; **and** the `glm task` factory-mode note describes the variable in lane vocabulary. Package: `internal/config`, `internal/codexwiring`, `internal/cli`.
- **AC-RNC-022** (maps REQ-RNC-009, REQ-RNC-022) — **Given** a live role declaration whose role is `lead` and a live `peers` row role `worker`, slot `worker-2` in the current run, **When** the board write guard, the SessionStart role reader, and a leader launch in that run read them, **Then** the board write is refused with `ErrNotSoleWriter` and a message naming `lead` and the relaunch step, the SessionStart output is the stale-run notice rather than a leader or lane notice, the leader launch exits non-zero naming the run, and both records are byte-identical afterwards; **and Given** a factory card whose recorded owner is `worker-3`, **When** its history is displayed, **Then** the owner reads `worker-3`. Package: `internal/kanban`, `internal/hook`, `internal/cli`.

### Notices and dashboard (M3)

- **AC-RNC-012** (maps REQ-RNC-015) — **Given** each of en, ko, ja, zh, **When** the factory leader notice, the factory lane notice, and the kanban leader notice render, **Then** each contains the design §3 leader and lane terms for that locale and contains none of: `worker-<n>`, `-f worker`, `-f agent`; the ko strings contain no `리드`; the en strings contain no standalone `lead` as a noun. Package: `internal/hook` (one table-driven test over all locales).
- **AC-RNC-013** (maps REQ-RNC-001) — **Given** a declared chain whose persisted leader role is `leader`, **When** the web dashboard view model and the doctor factory section render, **Then** the displayed role label is `leader` and the chain is recognized as present; **and Given** one whose persisted leader role is `lead`, **Then** it is shown as a legacy run that needs relaunch, not as a present leader. Package: `internal/web`, `internal/cli`.
- **AC-RNC-021** (maps REQ-RNC-021) — **Given** the built binary, **When** `moai todo next --help` runs, **Then** its output names both the operator's pick through the leader and a lane's self-dispatch as ways a queued card is promoted and contains no `lead session`; **and** `git diff <merge-base> -- internal/cli/todo.go` adds and removes no role check (no new or deleted call reading a role declaration or `MOAI_FACTORY_*` in the pick path). Package: `internal/cli`.
- **AC-RNC-023** (maps REQ-RNC-023) — **Given** the list of production user-facing strings naming the CG-mode leader or the Agent Teams lead, measured at pre-flight and recorded in progress.md, **When** each is read after the change, **Then** every entry carries its `CG` or `team` qualifier next to the noun (unqualified count 0), and `git diff --stat <merge-base>` shows no rename of the underlying identifiers or JSON fields. Recorded in progress.md.

### Role marker (M4, conditional)

- **AC-RNC-014** (maps REQ-RNC-012) — **Where** the role-marker variable constant exists: **Given** sessions whose marker is `lane`, `worker`, and `agent`, **When** the guarded commands of card t1245 run, **Then** only the `lane` session is denied with t1245's sentinel and the other two are allowed (t1245 REQ-AP-011's allow branch); the marker's value constant equals `lane`; t1245's equality assertion (value = `-f` role token = lane-label prefix) passes; and every production site that sets or compares the marker — listed with file:line in progress.md from a grep for the marker's name constant (t1245 REQ-AP-012) over production `.go` files under `internal/` — passes the value constant, with no site passing a string literal. **Where** it does not exist: progress.md records "value `lane` handed to t1245" and this AC is N/A.

### Boundaries, dependencies, mechanics (M5 and pre-flight)

- **AC-RNC-015** (maps REQ-RNC-014) — **Given** the run phase starts, **When** pre-flight completes, **Then** progress.md §E.2 records the develop SHA read, whether `internal/cli/codex_factory.go` is absent on develop, whether the `MOAI_FACTORY_ROLE` constants exist, and t1193's state; and no commit touching a t1242-deleted file predates a "t1242 landed" entry.
- **AC-RNC-017** (maps REQ-RNC-016, REQ-RNC-017) — **Given** the merge-base, **When** `git diff --stat <merge-base> -- internal/hook/subagent_start.go internal/sessionmsg internal/cli/agentlint internal/hook/session_end.go internal/tmux internal/cli/codex_role_fingerprint.go internal/cli/worktree/guard.go` runs, **Then** it prints nothing; **and** `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` equals its merge-base value.
- **AC-RNC-018** (maps REQ-RNC-018) — **Given** the post-change tree, **When** the vocabulary guard test runs over production `.go` string literals, **Then** it passes; its allowlist holds only the legacy-value literals the rejection and stale-record paths compare against and their error-message text (entry count recorded in progress.md); **and** a mutation that re-adds `"-f worker"` to a help string makes it fail (recorded red run).
- **AC-RNC-019** (maps REQ-RNC-019) — **Given** the post-change tree, **When** `grep -rnE 'factoryWorkerRoleToken|factoryLegacyAgentRoleToken|FactoryWorkerEntry|RoleLead\b|LeadLabel\(' internal --include='*.go'` runs, **Then** it returns 0 rows; each retained env-var name constant and each allowlisted legacy-value literal carries a comment stating it exists only to be refused or detected; and `go build ./...` and `go vet` on touched packages exit 0.
- **AC-RNC-020** (maps REQ-RNC-020) — **Given** the test tree, **When** each legacy spelling (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`) is searched in `_test.go` files of `internal/cli` and `internal/kanban`, **Then** each has at least one test asserting refusal — non-zero exit, the canonical form in the error, and no record written.

## §B Edge cases

- `-f lane-0`, `-f lane-`, `-f lane-a`, `-f lane-3-x` → rejected as not-a-lane, with the usage error.
- `-f Worker`, `-f WORKER-2` (case variants) → refused like the lowercase legacy spelling, not silently accepted.
- `--name leader-<run-id>` where run id is all digits → treated as a bump number, matching today's `lead-<digits>` rule.
- A registry holding a live `lead` and a live `leader` in one run (mid-upgrade) → the stale-record rule of REQ-RNC-022 fires on the next launch; neither record is rewritten.
- Broker slot `agent` supplied as a role input → refused like `worker` (AC-RNC-009).
- A marker value `worker` set by hand in a test environment → allowed by the guard (AC-RNC-014); no production stamp produces it.
- Locale fallback for an unknown language → the en table, which must itself pass AC-RNC-012.

## §C Traceability matrix

| REQ | AC |
|---|---|
| RNC-001 | 016, 013 |
| RNC-002 | 001 |
| RNC-003 | 002 |
| RNC-004 | 003, 004 |
| RNC-005 | 005 |
| RNC-006 | 006 |
| RNC-007 | 007 |
| RNC-008 | 010 |
| RNC-009 | 008, 022 |
| RNC-010 | 008 |
| RNC-011 | 011 |
| RNC-012 | 014 |
| RNC-013 | 009 |
| RNC-014 | 015 |
| RNC-015 | 012 |
| RNC-016 | 017 |
| RNC-017 | 017 |
| RNC-018 | 018 |
| RNC-019 | 019 |
| RNC-020 | 002, 005, 007, 020 |
| RNC-021 | 021 |
| RNC-022 | 003, 022 |
| RNC-023 | 023 |

## §D Quality gates

- Scoped `go test` on `internal/kanban`, `internal/cli`, `internal/hook`, `internal/factorymsg`, `internal/web`, `internal/config`, `internal/codexwiring` passes on the merge tree; full suite on CI after the leader's develop push.
- `golangci-lint run` on touched packages: 0 new findings.
- Coverage of touched packages not below the merge-base value; new rejection and stale-record paths covered.

## §E Definition of Done

- AC-RNC-001 … AC-RNC-023 PASS (AC-RNC-014 PASS or recorded N/A), each with command and verbatim output in progress.md §E.2.
- Open decisions in plan.md §B answered and recorded.
- Design §3 table handed to card t1257 unchanged or with recorded amendments.
