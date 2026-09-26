---
id: SPEC-ROLE-NAMING-CODE-001
title: "Acceptance — role naming unification (code + CLI)"
version: "0.1.0"
created: 2026-09-26
---

# Acceptance — SPEC-ROLE-NAMING-CODE-001

Every criterion is binary. "Hint" means exactly one line on stderr naming the canonical form. Test names are the run phase's to choose; each AC names the package that must hold its test.

## §A Acceptance criteria

### CLI tokens and labels (M2)

- **AC-RNC-001** (maps REQ-RNC-002) — **Given** a running factory with no live lane claims, **When** a lane joins with `-f lane`, **Then** the session launches as `lane-1`, the registry row's label is `lane-1`, and exit is 0. Package: `internal/cli`.
- **AC-RNC-002** (maps REQ-RNC-003, REQ-RNC-020) — **Given** the same factory, **When** a lane joins with `-f worker`, and separately with `-f agent`, **Then** each launches as the next free `lane-<n>`, exits 0, and prints a hint containing `-f lane`; no other stderr line mentions a deprecation. Package: `internal/cli`.
- **AC-RNC-003** (maps REQ-RNC-004) — **Given** a live registry claim `worker-3` and a live claim `agent-4`, **When** `-f lane` runs twice, **Then** the two launches take numbers other than 3 and 4, and `-f lane-3` is refused with a message naming `lane-3` and the legacy holder. Package: `internal/kanban` + `internal/cli`.
- **AC-RNC-004** (maps REQ-RNC-004) — **Given** the post-change tree, **When** the factory label composer is called with 3, **Then** it returns `lane-3`; and the canonical-label function maps `worker-3`, `agent-3`, and `lane-3` all to `lane-3`. Package: `internal/kanban`.
- **AC-RNC-005** (maps REQ-RNC-005, REQ-RNC-020) — **Given** no live claims, **When** a lane launches with `-f worker-2` and separately with `--name agent-5`, **Then** the sessions are named `lane-2` and `lane-5` and each prints one hint naming the `lane-<n>` form. Package: `internal/cli`.
- **AC-RNC-016** (maps REQ-RNC-001) — **Given** the built binary, **When** `moai cc --help` and `moai glm --help` run, **Then** each output contains `-f lane` and `-f lane-<n>` and contains neither `-f worker` nor `-f agent`; and the `-f` usage error text contains `lane` and not `worker`. Package: `internal/cli`.

### Leader label and persisted state (M1)

- **AC-RNC-006** (maps REQ-RNC-006) — **Given** no live leader, **When** a kanban leader and, in another project, a factory leader launch without `--name`, **Then** both sessions are named `leader`; a second concurrent leader in the same project is named `leader-2`. Package: `internal/cli` + `internal/kanban`.
- **AC-RNC-007** (maps REQ-RNC-007, REQ-RNC-020) — **Given** a live leader registry entry `lead` (written by the current release), **When** a new leader launches, **Then** it is named `leader-2`, not `leader`; **and When** a leader launches with `--name lead-7`, **Then** it is named `leader-7` and prints one hint. Package: `internal/cli`.
- **AC-RNC-008** (maps REQ-RNC-009, REQ-RNC-010) — **Given** a role declaration file whose role is `lead` and a broker `peers` row with role `lead`, slot `lead`, **When** the board write guard, the SessionStart role reader, and the run-retire owner lookup read them, **Then** each recognizes the leader (board write allowed for that session, notice rendered as leader, owner found); **and** after a new leader launches, the role and slot values it writes are `lead`. Package: `internal/kanban`, `internal/hook`, `internal/factorymsg`.
- **AC-RNC-009** (maps REQ-RNC-009, REQ-RNC-013) — **Given** broker peers `lead`/`lead`, `worker`/`worker-1`, and `worker`/`lane-2`, **When** `factory_msg_send` targets `to_slot` = `leader`, `lead`, `lane-1`, `worker-1`, `agent-1`, and `lane-2` in turn, **Then** each message is delivered to the same endpoint the current-release spelling reaches, and none returns an unknown-slot error. Package: `internal/factorymsg` + `internal/cli`.
- **AC-RNC-010** (maps REQ-RNC-008) — **Given** develop at the merge-base, **When** `git diff <merge-base> -- internal/homestate internal/factorymsg` is filtered to changed lines containing `CREATE TABLE`, `ALTER TABLE`, or `CREATE INDEX`/`CREATE UNIQUE INDEX`, **Then** the filter prints 0 lines. Recorded in progress.md.
- **AC-RNC-011** (maps REQ-RNC-011) — **Given** the post-change tree, **When** the four environment variable name constants and the generated Codex `env_vars` value are read, **Then** they equal `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, and the develop-merge-base allowlist string byte for byte; **and** the `glm task` factory-mode note describes the variable as the factory signal in lane vocabulary. Package: `internal/config`, `internal/codexwiring`, `internal/cli`.

### Notices and dashboard (M3)

- **AC-RNC-012** (maps REQ-RNC-015) — **Given** each of en, ko, ja, zh, **When** the factory leader notice, the factory lane notice, and the kanban leader notice render, **Then** each contains the design §3 leader and lane terms for that locale and contains none of: `worker-<n>`, `-f worker`, `-f agent`; the ko strings contain no `리드`; the en strings contain no standalone `lead` as a noun. Package: `internal/hook` (one table-driven test over all locales).
- **AC-RNC-013** (maps REQ-RNC-001) — **Given** a declared chain whose persisted leader role is `lead`, **When** the web dashboard view model and the doctor factory section render, **Then** the displayed role label is `leader` and the chain is recognized as present. Package: `internal/web`, `internal/cli`.

### Role marker (M4, conditional)

- **AC-RNC-014** (maps REQ-RNC-012) — **Where** the role-marker variable constant exists: **Given** sessions whose marker is `lane`, `worker`, and `agent`, **When** the guarded commands of card t1245 run, **Then** all three are denied with t1245's sentinel; the marker's value constant equals `lane`; and the launcher stamp site and the guard both reference that constant (a grep for the quoted value in the hook and launcher packages returns 0 outside the constant definition and tests). **Where** it does not exist: progress.md records "value `lane` handed to t1245" and this AC is N/A.

### Boundaries, preconditions, mechanics (M5 and pre-flight)

- **AC-RNC-015** (maps REQ-RNC-014) — **Given** the run phase starts, **When** pre-flight completes, **Then** progress.md §E.2 records the develop SHA read, whether `internal/cli/codex_factory.go` is absent on develop, and t1193's state; and no commit touching the six t1193-overlap files predates a "resolved" entry.
- **AC-RNC-017** (maps REQ-RNC-016, REQ-RNC-017) — **Given** the merge-base, **When** `git diff --stat <merge-base> -- internal/hook/subagent_start.go internal/sessionmsg internal/cli/agentlint internal/hook/session_end.go internal/tmux internal/cli/codex_role_fingerprint.go internal/cli/worktree/guard.go` runs, **Then** it prints nothing; **and** `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` equals its merge-base value.
- **AC-RNC-018** (maps REQ-RNC-018) — **Given** the post-change tree, **When** the vocabulary guard test runs over production `.go` string literals, **Then** it passes; its allowlist holds only legacy parse values and hint text (entry count recorded in progress.md); **and** a mutation that re-adds `"-f worker"` to a help string makes it fail (recorded red run).
- **AC-RNC-019** (maps REQ-RNC-019) — **Given** the post-change tree, **When** `grep -rnE 'factoryWorkerRoleToken|FactoryWorkerEntry|EnvMoaiFactoryWorker[^s]' internal --include='*.go'` runs, **Then** the only matches are the env-var constant name (kept by REQ-RNC-011) and its references; and each retained persisted-key definition carries a comment stating it is an on-disk key. `go build ./...` and `go vet` on touched packages exit 0.
- **AC-RNC-020** (maps REQ-RNC-020) — **Given** the test tree, **When** each legacy spelling (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`) is searched in `_test.go` files of `internal/cli` and `internal/kanban`, **Then** each has at least one test asserting it parses and hints.

## §B Edge cases

- `-f lane-0`, `-f lane-`, `-f lane-a`, `-f lane-3-x` → rejected as not-a-lane exactly as the worker shapes are today.
- `--name leader-<run-id>` where run id is all digits → treated as a bump number, matching today's `lead-<digits>` rule.
- A registry holding both `lead` and `leader` live (mid-upgrade) → next leader takes the next free number; neither is displaced.
- Broker slot `agent` supplied as a role token (not a label) → resolves like `lane`.
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
| RNC-009 | 008, 009 |
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

## §D Quality gates

- Scoped `go test` on `internal/kanban`, `internal/cli`, `internal/hook`, `internal/factorymsg`, `internal/web`, `internal/config`, `internal/codexwiring` passes on the merge tree; full suite on CI after the leader's develop push.
- `golangci-lint run` on touched packages: 0 new findings.
- Coverage of touched packages not below the merge-base value; new alias code paths covered.

## §E Definition of Done

- AC-RNC-001 … AC-RNC-020 PASS (AC-RNC-014 PASS or recorded N/A), each with command and verbatim output in progress.md §E.2.
- Open decisions O0-O7 answered and recorded.
- Design §3 table handed to card t1257 unchanged or with recorded amendments.
