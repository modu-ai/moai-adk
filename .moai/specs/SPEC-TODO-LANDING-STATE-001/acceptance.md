# SPEC-TODO-LANDING-STATE-001 — Acceptance Criteria

Every AC below names the **observed RED** — the command and the wrong output it produced **before**
the change, measured at tree `3de2f85a2` — and the **expected GREEN**, so a later reader can re-run
both. An AC whose RED was not actually observed says so in its RED cell rather than asserting one.

Commands run from the card worktree root (`.claude/worktrees/t331`) unless stated.

---

## §A Standing preconditions

### A.1 Baseline tree [HARD]

Every RED cell below was measured at `3de2f85a2`. Re-measuring a RED at a later tree requires
re-citing the tree; a RED carried forward without re-measurement is a carry-over, not a baseline
(`verification-claim-integrity.md` §2).

### A.2 Exit-code reading [HARD]

`| wc -l` reports `0` for a command that failed. Every count-shaped assertion reads the exit code
separately or runs under `set -o pipefail`.

### A.3 Verification isolation [HARD]

Suites run scoped to the affected packages. An environment-scrubbed run inside this worktree is one
compound invocation (`unset … && go test …`), never a separate `unset`.

### A.4 The three misjudgement inputs

These are test inputs, not illustrations. Each AC that repairs a discriminator names which one it is
established against.

| Input | Shape | Measured fact @ `3de2f85a2` |
|---|---|---|
| **squash-blind** | t200 | `git merge-base --is-ancestor 7fc161b36 origin/develop` → rc 1; content landed as `294b4b6ab` (PR #1612), rc 0 |
| **false not-started** | t338 | `git log origin/develop --perl-regexp --grep='\bt338\b' --oneline` → 0 commits; work existed, a dead session's stale `index.lock` blocked committing, 43 untracked artifacts alive |
| **silent pass** | F6 | `todo done --require-landed` on an unanswerable query exits 0 and prints `done <id>`, byte-identical to a satisfied guard |

---

## §B AC Matrix

| AC | Verifies | Scenario (Given / When / Then) | Observed RED @ `3de2f85a2` | Expected GREEN |
|---|---|---|---|---|
| **AC-TLS-001** | REQ-TLS-001, REQ-TLS-002 | **Given** a project whose `git_strategy.worktree_base_branch` is `develop` **when** the landed check runs for a card named only by `origin/develop` **then** it answers `landed`. | `git log origin/main --perl-regexp --grep='\bt293\b' --oneline` → **no output (0 commits)**, while `origin/develop` → 9 commits. The check answers `not landed`, and `todo pr` renders `no-link`. Same for t310 (0 / 6) and t322 (0 / 5). | Unit test over the resolver: configured `develop` ⇒ ref `origin/develop`; t293-shaped fixture answers `landed`. |
| **AC-TLS-002** | REQ-TLS-002, REQ-TLS-004 | **Given** a project whose `worktree_base_branch` is empty **when** the landed check runs **then** it uses `origin/main`, byte-identically to `3de2f85a2`. | Not a defect — this is the behaviour being preserved. RED cell: none; the criterion is a **no-change** assertion. | Table test: empty ⇒ `origin/main`; the existing `todo pr` golden output over an unconfigured fixture is unchanged. |
| **AC-TLS-003** | REQ-TLS-003 | **Given** a project configured on `develop` **when** `moai todo done tX --require-landed` refuses **then** the refusal names `origin/develop`. | The refusal names `origin/main` unconditionally — the string is built from the constant (`internal/cli/todo.go:428` @ `3de2f85a2`), and `internal/cli/todo_undone_test.go:277` asserts it against `kanban.LandedRef`. | Refusal text asserted against the **resolved** ref; the existing test updated to resolve rather than to read a constant. |
| **AC-TLS-004** | REQ-TLS-005..007 | **Given** a resolved ref that does not exist **when** the landed check runs **then** it answers `unknown`, and the answer is a field a caller must read. | `git log origin/no-such-ref --perl-regexp --grep='\bt331\b' --oneline ; echo rc=$?` → `fatal: ambiguous argument 'origin/no-such-ref'`, `rc=128`. `Landed` maps this to `(false, err)` (`internal/kanban/prlink_landed.go:76-80` @ `3de2f85a2`) — a caller that drops the error reads `not landed`. | Three-valued answer; a caller that ignores the answer field does not compile. Test asserts `unknown` ≠ `not-landed` for the unresolvable ref. |
| **AC-TLS-005** | REQ-TLS-008 | **Given** an unresolvable landed ref **when** `moai todo pr` renders **then** the outcome kind is `unknown`, distinct from `no-link`, and names the ref. | `ResolveCardPRLink` returns `Kind: no-link` with the error alongside (`internal/kanban/prlink.go:170-178` @ `3de2f85a2`); `runTodoPR` prints the same `no-link` cell and appends a stderr note (`internal/cli/todo_pr.go:142-147`). Column output for a degraded card is identical to a genuinely unstarted one. | Fifth kind rendered; golden output distinguishes the two rows. |
| **AC-TLS-006** | REQ-TLS-009..011 | **Given** `--require-landed` **when** the query is unanswerable **then** stdout differs from the satisfied case. | Two PASSing tests @ `3de2f85a2` establish the indistinguishability: `TestTodoDone_RequireLandedProceedsWhenInconclusive` asserts exit 0 + stdout prefix `done t1` on an unanswerable query (`internal/cli/todo_undone_test.go:287-302`), and `TestTodoDone_NoLandingQueryWithoutTheFlag` asserts exit 0 with a stub reporting landed (`:326-331`). Verbatim: `--- PASS: TestTodoDone_RequireLandedProceedsWhenInconclusive (0.14s)` and `--- PASS: TestTodoDone_NoLandingQueryWithoutTheFlag (0.19s)`. | A new test captures stdout for all three cases (`landed` / `not-landed` / `unknown`) and asserts three distinct verdict tokens. The two existing tests are updated to assert the new stdout, deliberately. |
| **AC-TLS-007** | REQ-TLS-012 | **Given** `--require-landed` and an `unknown` answer **when** `done` runs **then** the card is archived and the exit code is 0. | Current behaviour, and it is being **preserved**: `TestTodoDone_RequireLandedProceedsWhenInconclusive` PASS @ `3de2f85a2`. RED cell: none — this is a policy-preservation criterion. | That test still passes on its policy assertions after M3 rewrites its stdout expectation. |
| **AC-TLS-008** | REQ-TLS-013, 014, 016 | **Given** a card with two recorded landings **when** the record is reloaded **then** both survive in observation order, and `BacklogItem` still has five fields. | No landing storage exists @ `3de2f85a2` (`grep -n 'landing' internal/kanban/*.go` → no match). RED is absence. | Round-trip test: two observations in, two out, in order. A reflection or field-count test pins `BacklogItem` at five fields (REQ-TODO-013). |
| **AC-TLS-009** | REQ-TLS-015 | **Given** the t200 shape — card-branch commit `7fc161b36`, landed squash commit `294b4b6ab` **when** a landing is recorded **then** the recorded SHA is on the resolved ref. | `git merge-base --is-ancestor 7fc161b36 origin/develop ; echo $?` → `1`. `git merge-base --is-ancestor 294b4b6ab origin/develop ; echo $?` → `0`. A recorder that took the card-branch HEAD would store a SHA absent from the ref. | Test asserts the recorded SHA satisfies `merge-base --is-ancestor <sha> <ref>` rc 0. This is the **squash-blind** input's discriminator. |
| **AC-TLS-010** | REQ-TLS-017, 018 | **Given** a queue database created by a binary predating this change **when** a new binary opens it **then** the `landings` table appears with no migration, and the old binary still opens and mutates it. | `internal/kanban/backlog_sqlite.go:232-235` @ `3de2f85a2` runs the whole DDL on every open, all `IF NOT EXISTS` — the mechanism exists; the table does not. RED is absence. | Open-old-db test asserts the table appears and `schema_version` stays `"1"`; a grep asserts zero `ALTER TABLE` in the diff. |
| **AC-TLS-011** | REQ-TLS-019, 020 | **Given** a picked card with a `spec_id` **when** `todo pr` renders **then** the live SPEC status is shown with its source tag; an unresolvable status renders as unresolved. | `todo pr` renders five columns and none is SPEC status (`internal/cli/todo_pr.go:167-172` @ `3de2f85a2`). Fully-landed t293/t310/t322 render as `no-link` with no status signal at all. | Status column populated via `kanban.ReadCardStatus`; a fixture with no SPEC file renders `unresolved`, never a status value. |
| **AC-TLS-012** | REQ-TLS-021, 023 | **Given** any card **when** `moai todo land <id>` runs **then** the card's `state`, position, text, and `spec_id` are byte-identical afterwards. | Verb does not exist @ `3de2f85a2`. RED is absence. | Byte-identity assertion over the item row across the verb, matching the `todo relate` precedent. |
| **AC-TLS-013** | REQ-TLS-022 | **Given** a card whose landing is detected and whose SPEC reads `completed` **when** every read and record path runs **then** no card state changes anywhere. | Cannot go red today — no detection path writes. The criterion guards a **future** slip, so it is asserted mechanically: a source sweep proving no `state` write is reachable from the detection or recording paths. | Sweep test: no assignment to `Items[i].State`, no `ArchiveCard`, no `Drop` call reachable from the landing paths; plus a behavioural test over a landed-and-completed fixture asserting the queue file is byte-identical. |
| **AC-TLS-014** | REQ-TLS-024 | **When** `moai todo pr` runs over a seeded queue **then** the queue file, the lock, and the landings table are byte-identical afterwards. | Holds today by ruling (`internal/cli/todo_pr.go:1-16` @ `3de2f85a2`); the criterion prevents M6's status read from becoming a write. | Byte-identity assertion extended to cover the new table. |
| **AC-TLS-015** | REQ-TLS-025, 026 | **When** both `todo.md` copies are read **then** they carry the new verb, the `unknown` outcome, the stdout verdict, the unweakened [HARD] operator-only rule, and the stated remaining limit. | Both copies @ `3de2f85a2` describe four outcomes and a two-valued landed check (`.claude/skills/moai/workflows/todo.md:51`, `:59-68`; mirror at the same line numbers). | Both files updated in one change; `make build` exit 0; a grep proves the embedded bundle carries the edit. |
| **AC-TLS-016** | spec.md §A.6 C2 row | **Given** a card that is `picked` with zero commits (the t338 shape) **when** `todo pr` renders **then** the row is distinguishable from a `queued`, never-started card. | `git log origin/develop --perl-regexp --grep='\bt338\b' --oneline` → **no output**. The card reads exactly like C1. Absence of commits is not absence of work — t338's work existed and was blocked by a stale `index.lock`. | The rendered row carries the queue `state` alongside the link outcome, so `queued`+`no-link` and `picked`+`no-link` are different rows. This is the **false not-started** input's discriminator. Note the honest limit, asserted as part of the AC: the tooling distinguishes *picked with no commits* from *never started*; it does **not** claim to detect work that produced no commit. |

---

## §C Severity classification

- **Release-blocking**: AC-TLS-001, 003, 004, 005, 006, 008, 009, 010, 012, 013, 014, 015, 016.
- **Regression-guard (must stay green, a failure is a contract break)**: AC-TLS-002, AC-TLS-007.
- **Quality-bar (a FAIL requires a written debt note)**: AC-TLS-011 — the status column is a read-path
  enrichment; a degraded status read must not block the verb.

## §D Indirect verification accepted

- The three misjudgement inputs are **historical**: t200's squash, t293/t310/t322's develop landings,
  and t338's blocked session cannot be re-created live. Each AC therefore asserts against a
  **fixture built from the measured shape**, with the measurement quoted in the RED cell so the
  fixture's fidelity is auditable.
- AC-TLS-013's forward-looking half (no future auto-transition) is verified by a source sweep, not by
  execution. A sweep proves reachability, not intent; the intent is carried by spec.md §B.4.
- Windows behaviour arrives via CI job read-back; the local obligation stops at
  `GOOS=windows go vet ./internal/...` compile evidence.

## §E Closure gates

1. Every release-blocking AC PASSes with an evidence path resolvable from `progress.md` §E.2.
2. Both `todo.md` copies updated and `make build` run in the same change (Template-First).
3. The subprocess census tests pass, or their budget change is stated and justified in `progress.md`.
4. `moai todo pr` byte-identity assertion (AC-TLS-014) green — the read-only ruling survived.

## §F Definition of Done

- The landed question is asked about the configured integration branch, and an unconfigured project
  is byte-identical to `3de2f85a2`.
- `landed`, `not-landed`, and `unknown` are three distinct answers, and `unknown` is visible on both
  the `todo pr` surface and the `todo done` stdout.
- A card can carry landing observations, each naming a commit that exists on the ref it names, and
  more than one observation per card survives a round trip.
- No path in the change closes, moves, drops, or re-states a card on the tooling's own authority.
- The doctrine and its template mirror say all of the above, including what the check still cannot
  answer.
