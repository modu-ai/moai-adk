---
id: SPEC-ROLE-NAMING-CODE-001
title: "Acceptance — role naming unification (code + CLI)"
version: "0.3.1"
created: 2026-09-26
---

# Acceptance — SPEC-ROLE-NAMING-CODE-001

Every criterion is binary. "Refused" means: no session launched, no registry / broker / role-declaration record written (row counts and file listings equal before and after), exactly one error line naming the canonical form, and a non-zero exit — the same status the launcher returns for any other invalid `-f` value. Test names are the run phase's to choose; each AC names the package that must hold its test.

**Word-boundary rule (applies to every AC below that forbids `lead`).** "Contains no `lead`" means no match of the case-insensitive regular expression `(?i)\blead\b`: `lead`, `Lead`, and `LEAD` match; `leader`, `leading`, and `mislead` do not; and an occurrence inside a `MOAI_[A-Z0-9_]+` token is not counted. Under word-boundary matching `_` is a word character, so `LEAD` inside `MOAI_KANBAN_LEAD_ADDR` is not a separate word and never matches; the exclusion states that invariant, and AC-RNC-018's paired control proves a token does not mask a real hit in the same string.

## §A Acceptance criteria

### CLI tokens and labels (M2)

- **AC-RNC-001** (maps REQ-RNC-002) — **Given** a running factory with no live lane claims, **When** a lane joins with `-f lane`, **Then** the session launches as `lane-1`, the registry row's label is `lane-1`, and exit is 0. Package: `internal/cli`.
- **AC-RNC-002** (maps REQ-RNC-003, REQ-RNC-020) — **Given** the same factory, **When** a join runs with `-f worker`, separately with `-f agent`, and separately with `-f WORKER`, **Then** each is refused, its error line contains `-f lane`, and the registry row count is unchanged. Package: `internal/cli`.
- **AC-RNC-003** (maps REQ-RNC-004, REQ-RNC-022) — **Given** a live registry claim `worker-3` in the running factory (written by a pre-change binary), **When** `-f lane` runs, **Then** it is refused with a message naming `worker-3`, the run id, and `moai factory runs --retire`, and no `lane-<n>` claim is written; **and Given** the same claim whose pid is dead, **When** `-f lane` runs, **Then** it succeeds exactly as with no claim. Package: `internal/kanban` + `internal/cli`.
- **AC-RNC-004** (maps REQ-RNC-004) — **Given** the post-change tree, **When** the factory label composer is called with 3, **Then** it returns `lane-3`; **and When** the lane-label parser receives `lane-3`, `worker-3`, `agent-3`, `lane-0`, `lane-`, `lane-a`, and `lane-3-x`, **Then** it accepts only `lane-3` and reports every other input as not-a-lane-label. Package: `internal/kanban`.
- **AC-RNC-005** (maps REQ-RNC-005, REQ-RNC-020) — **Given** no live claims, **When** a lane launch runs with `-f worker-2`, separately with `--name agent-5`, and separately with `-f Worker-4`, **Then** each is refused and the error lines contain `lane-2`, `lane-5`, and `lane-4` respectively. Package: `internal/cli`.
- **AC-RNC-016** (maps REQ-RNC-001) — **Given** the built binary, **When** `moai cc --help` and `moai glm --help` run, **Then** each output contains `-f lane` and `-f lane-<n>` and contains neither `-f worker` nor `-f agent`; and the `-f` usage error text contains `lane` and `leader`, contains no `worker`, and contains no `\blead\b` match (per the word-boundary rule above). Package: `internal/cli`.

### Leader label and persisted state (M1)

- **AC-RNC-006** (maps REQ-RNC-006) — **Given** no live leader, **When** a kanban leader and, in another project, a factory leader launch without `--name`, **Then** both sessions are named `leader`; a second concurrent leader in the same project is named `leader-2`; and a factory leader label composed for run id `r7` reads `leader-r7`. Package: `internal/cli` + `internal/kanban`.
- **AC-RNC-007** (maps REQ-RNC-007, REQ-RNC-020) — **Given** no live leader, **When** a leader launch runs with `--name lead`, and separately with `--name lead-7`, **Then** each is refused, the error lines contain `leader` and `leader-7` respectively, and `leads.json` is byte-identical before and after. Package: `internal/cli`.
- **AC-RNC-008** (maps REQ-RNC-009, REQ-RNC-010) — **Given** a fresh factory run, **When** a leader launches, a lane joins, and the lane claims a card, **Then** the written records carry exactly these values, and no written value equals `lead`, `worker`, or `agent`:

  | Writer | Record | Expected value |
  |---|---|---|
  | leader | broker `peers.role` | `leader` |
  | leader | broker `peers.slot` | `leader` |
  | lane | broker `peers.role` | `lane` |
  | lane | broker `peers.slot` | `lane-1` |
  | lane | factory registry label | `lane-1` |
  | lane | factory card owner (`cards.owner_label`) | `lane-1` |

  (A board role declaration is not in this table: no launch path writes one — `DeclareRole` has no production caller.) **And When** the board write guard reads a board role declaration whose role is `leader`, written by the test itself through `DeclareRole`, the run-retire owner lookup reads a `peers` row with role `leader`, slot `leader`, and SessionStart runs in a leader launch environment, **Then** the board write is admitted, the owner lookup returns that peer's identity, and the session record SessionStart writes carries role `leader`. Package: `internal/kanban`, `internal/hook`, `internal/factorymsg`, `internal/cli`.
- **AC-RNC-009** (maps REQ-RNC-013) — **Given** broker peers `leader`/`leader` and `lane`/`lane-1`, **When** `factory_msg_send` targets `to_slot` = `leader` and `lane-1`, **Then** each is delivered to that peer; **and When** it targets `lead`, `worker-1`, `agent-1`, `worker`, and `agent` in turn, **Then** each returns an error naming the canonical slot (`leader` or `lane-1` / `lane`) and the broker message count is unchanged. Package: `internal/factorymsg` + `internal/cli`.
- **AC-RNC-010** (maps REQ-RNC-008) — **Given** develop at the merge-base, **When** `git diff <merge-base> -- internal/homestate internal/factorymsg` is filtered to changed lines containing `CREATE TABLE`, `ALTER TABLE`, or `CREATE INDEX`/`CREATE UNIQUE INDEX`, **Then** the filter prints 0 lines. Recorded in progress.md.
- **AC-RNC-011** (maps REQ-RNC-011) — **Given** the post-change tree, **When** the four environment variable name constants and the generated Codex `env_vars` value are read, **Then** they equal `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, and the develop-merge-base allowlist string byte for byte, and `grep -rnE '"MOAI_(KANBAN_LEADER|FACTORY_LANE)' internal --include='*.go'` returns 0 rows; **and When** a lane joins, **Then** the child environment's `MOAI_FACTORY_WORKER` value is `lane-<n>` and a leader's `MOAI_KANBAN_LEAD_NAME` value is `leader` or `leader-<suffix>`; **and** the `glm task` factory-mode note describes the variable in lane vocabulary. Package: `internal/config`, `internal/codexwiring`, `internal/cli`.
- **AC-RNC-022** (maps REQ-RNC-009, REQ-RNC-022) — **Given** a factory run `R` whose broker holds a live `peers` row role `worker`, slot `worker-2`, **When** a factory leader re-enters with `--factory-run R`, **Then** it exits non-zero with one message naming `worker-2`, `R`, and `moai factory runs --retire`, and the peer row and the run row are byte-identical afterwards; **and When** the factory SessionStart hook runs for run `R` with the lane-label variable set to `worker-2`, **Then** its output is the stale-run message naming `worker-2` and `R`, not a leader or lane notice, and it writes no session record; **and Given** the same row whose pid is dead, and separately a dead legacy leader peer row role `lead`, slot `lead`, **When** the leader re-enters with `--factory-run R`, **Then** it proceeds exactly as with no such row; **and Given** a factory card whose recorded owner is `worker-3`, **When** its history is displayed, **Then** the owner reads `worker-3`. Package: `internal/factorymsg`, `internal/hook`, `internal/cli`.
- **AC-RNC-024** (maps REQ-RNC-024) — **Given** a factory run `R` whose run row has `lead_pid = 0` and whose broker holds one peer role `lead` with a recorded pid and process start, **When** that pid is dead and `moai factory runs --retire R` runs, **Then** it prints `retired R (owner dead)`, the retirement event records basis `peer`, and a subsequent factory leader launch succeeds; **and When** that pid is live (the test's own process identity), **Then** the retire exits non-zero with `ErrRunOwnerNotDead` naming owner `live` and the run row is byte-identical; **and** no reader other than the run-retire owner lookup treats that peer as the leader (the SessionStart role reader and the broker slot matcher still refuse it per AC-RNC-022 and AC-RNC-009). Package: `internal/homestate` + `internal/factorymsg` + `internal/cli`.
- **AC-RNC-025** (maps REQ-RNC-009, REQ-RNC-025) — **Given** a `leads.json` holding a live entry `lead`, **When** a kanban leader launches without `--name`, **Then** it launches as `leader` with exit 0, prints one notice naming `lead` and the relaunch step, and the `lead` entry is byte-identical afterwards; **and Given** a board role declaration whose role is `lead`, **When** that session attempts a board write, **Then** it is refused with `ErrNotSoleWriter` and a message naming `lead`, and the declaration file (written by the test through `DeclareRole`) is byte-identical afterwards; **and** SessionStart is checked for each trigger separately: **(a) label trigger — Given** no session record for the session, **When** SessionStart runs with `MOAI_KANBAN` set and `MOAI_KANBAN_LEAD_NAME=lead`, **Then** its output is the stale-run notice rather than a leader notice and no session record file is created; **(b) record trigger — Given** an existing session record whose role is `lead`, **When** SessionStart runs with `MOAI_KANBAN` set and `MOAI_KANBAN_LEAD_NAME=leader`, **Then** its output is the stale-run notice rather than a leader notice and the session record file is byte-identical afterwards (the writer did not re-derive `leader` and overwrite it). Package: `internal/kanban`, `internal/hook`, `internal/cli`.

### Notices and dashboard (M3)

- **AC-RNC-012** (maps REQ-RNC-015) — **Given** each of en, ko, ja, zh, **When** the factory leader notice, the factory lane notice, and the kanban leader notice render, **Then** each contains the design §3 leader and lane terms for that locale and contains none of: `worker-<n>`, `-f worker`, `-f agent`; the ko strings contain no `리드`; the en strings contain no `\blead\b` match (per the word-boundary rule above). Package: `internal/hook` (one table-driven test over all locales).
- **AC-RNC-013** (maps REQ-RNC-001) — **Given** a declared chain whose persisted leader role is `leader`, **When** the web dashboard view model and the doctor factory section render, **Then** the displayed role label is `leader` and the chain is recognized as present; **and Given** one whose persisted leader role is `lead`, **Then** the view model reports no present leader and its role label reads exactly `legacy run: relaunch required`, and the doctor factory section contains the literal `legacy run: relaunch required`. Package: `internal/web`, `internal/cli`.
- **AC-RNC-021** (maps REQ-RNC-021) — **Given** the built binary, **When** `moai todo next --help` runs, **Then** its output names both the operator's pick through the leader and a lane's self-dispatch as ways a queued card is promoted and contains no `lead session`; **and** `git diff <merge-base> -- internal/cli/todo.go` adds and removes no role check (no new or deleted call reading a role declaration or `MOAI_FACTORY_*` in the pick path). Package: `internal/cli`.
- **AC-RNC-023** (maps REQ-RNC-023) — **Given** the population produced at pre-flight by `grep -rnE '"[^"]*\b[Ll]ead(er)?\b[^"]*"' internal cmd pkg --include='*.go' | grep -v '_test\.go:' | sed -E 's#[[:space:]]+//.*$##' | grep -E '"[^"]*\b[Ll]ead(er)?\b[^"]*"'` (the `sed` step drops Go comments, whole-line and trailing, so a quoted example inside a comment is not in the population), filtered to rows that also match `grep -iE 'moai cg|\bcg\b|agent teams|teammate|team'`, and recorded verbatim in progress.md, **When** each row is read after the change, **Then** every entry carries its `CG` or `team` qualifier next to the noun (unqualified count 0), and `git diff --stat <merge-base>` shows no rename of the underlying identifiers or JSON fields. Recorded in progress.md.

### Role marker (M4, conditional)

- **AC-RNC-014** (maps REQ-RNC-012) — **Where** the role-marker variable constant exists: **Given** sessions whose marker is `lane`, `worker`, and `agent`, **When** the guarded commands of card t1245 run, **Then** only the `lane` session is denied with t1245's sentinel and the other two are allowed (t1245 REQ-AP-011's allow branch); the marker's value constant equals `lane`; t1245's REQ-AP-013 equality assertion (marker value = `-f` role token = lane-label prefix) passes with all three equal to `lane`, and a mutation that sets any one of the three to `worker` makes it fail (recorded red run); and every production site that sets or compares the marker — listed with file:line in progress.md from a grep for the marker's name constant (t1245 REQ-AP-012) over production `.go` files under `internal/` — passes the value constant, with no site passing a string literal. **Where** it does not exist: progress.md records "value `lane` handed to t1245" and this AC is N/A.

### Boundaries, dependencies, mechanics (M5 and pre-flight)

- **AC-RNC-015** (maps REQ-RNC-014) — **Given** the run phase starts, **When** pre-flight completes, **Then** progress.md §E.2 records the develop SHA read, whether `internal/cli/codex_factory.go` is absent on develop, whether the `MOAI_FACTORY_ROLE` constants exist, and t1193's state; and no commit touching a t1242-deleted file predates a "t1242 landed" entry.
- **AC-RNC-017** (maps REQ-RNC-016, REQ-RNC-017) — **Given** the merge-base, **When** `git diff --stat <merge-base> -- internal/hook/subagent_start.go internal/sessionmsg internal/cli/agentlint internal/hook/session_end.go internal/tmux internal/cli/codex_role_fingerprint.go internal/cli/worktree/guard.go` runs, **Then** it prints nothing; **and** `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` equals its merge-base value.
- **AC-RNC-018** (maps REQ-RNC-018) — **Given** the post-change tree, **When** the vocabulary guard test runs over production `.go` string literals in `internal/cli`, `internal/kanban`, `internal/hook`, `internal/factorymsg`, and `internal/web`, matching `-f worker`, `-f agent`, `worker-<digits or n>`, `agent-<digits or n>`, and `(?i)\blead\b` (not counted inside `MOAI_[A-Z0-9_]+` tokens, not counted when directly preceded by `team ` in any letter case), **Then** it passes; every allowlist entry names the file and the exact allowlisted literal of the comparison site, error-message site, or non-role-sense site it covers, the guard fails when the named file no longer contains that literal, and an edit that only moves the literal to another line does not fail it; the entry count is recorded in progress.md; **and** mutations that re-add `"-f worker"` to a help string, that change `leader` to `lead` in a notice, and that change `leader` to `Lead` at the start of a notice sentence each make it fail (recorded red runs); **and** the env-token controls are paired: a string containing only `MOAI_KANBAN_LEAD_ADDR` passes, while the string `MOAI_KANBAN_LEAD_ADDR names the lead address` fails on its free-standing `lead` (a recorded hit in a string that contains the token); a string containing `leader` passes.
- **AC-RNC-019** (maps REQ-RNC-019) — **Given** the post-change tree, **When** `grep -rnE 'factoryWorkerRoleToken|factoryLegacyAgentRoleToken|FactoryWorkerEntry|RoleLead\b|LeadLabel\(' internal --include='*.go'` runs, **Then** it returns 0 rows; each of the four retained env-var name constants carries a comment stating the name is kept under REQ-RNC-011 and its value follows the leader/lane vocabulary (and none carries a "refused or detected" comment); each allowlisted legacy-value literal carries a comment stating it exists only to be refused or detected; and `go build ./...` and `go vet` on touched packages exit 0; **and**, because REQ-RNC-019 covers identifiers and comments beyond these five patterns, M5 closes by re-running the tracked `.moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py` (`--out` pointing at an untracked path) on the post-change tree and recording in progress.md the identifier-internal rows that still carry role-sense `lead`, `worker`, or `agent`, each tagged with the exclusion that keeps it (REQ-RNC-016, REQ-RNC-017, a persisted key under REQ-RNC-008, or an env-var name under REQ-RNC-011), and the count of untagged rows, which must be 0.
- **AC-RNC-020** (maps REQ-RNC-020) — **Given** the test tree, **When** each legacy spelling (`-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`) is searched in `_test.go` files of `internal/cli` and `internal/kanban`, **Then** each has at least one test asserting refusal — non-zero exit, the canonical form in the error, and no record written.

## §B Edge cases

Each edge case below is exercised by the AC named next to it.

- `-f lane-0`, `-f lane-`, `-f lane-a`, `-f lane-3-x` → rejected as not-a-lane, with the usage error (AC-RNC-004).
- `-f WORKER`, `-f Worker-4` (case variants) → refused like the lowercase legacy spelling, not silently accepted (AC-RNC-002, AC-RNC-005).
- `--name leader-<run-id>` where run id is all digits → treated as a bump number, matching today's `lead-<digits>` rule (AC-RNC-006 covers the composed run-id form; the digits case is part of the same test).
- A factory run holding a live `lead` peer and a live `leader` peer (mid-upgrade) → the stale-record rule of REQ-RNC-022 fires on the next launch; neither record is rewritten (AC-RNC-022).
- A dead legacy leader record → treated as stale, not refused (AC-RNC-022).
- A legacy-only run whose leader session has exited → retirable through `moai factory runs --retire` (AC-RNC-024).
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
| RNC-009 | 008, 022, 025 |
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
| RNC-024 | 024 |
| RNC-025 | 025 |

## §D Quality gates

- Scoped `go test` on `internal/kanban`, `internal/cli`, `internal/hook`, `internal/factorymsg`, `internal/homestate`, `internal/web`, `internal/config`, `internal/codexwiring`, and `./internal/spec` passes on the merge tree; full suite on CI after the leader's develop push.
- `golangci-lint run` on touched packages: 0 new findings.
- Coverage of touched packages not below the merge-base value; new rejection, stale-record, and legacy-retire paths covered.

## §E Definition of Done

- AC-RNC-001 … AC-RNC-025 PASS (AC-RNC-014 PASS or recorded N/A), each with command and verbatim output in progress.md §E.2.
- Decisions in plan.md §B recorded as resolved (no open item).
- Design §3 table handed to card t1257 unchanged or with recorded amendments.
