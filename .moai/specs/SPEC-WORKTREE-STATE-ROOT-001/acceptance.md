# SPEC-WORKTREE-STATE-ROOT-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`. Every
criterion names a **RED-now** cell (the input that fails on base `c630de892`, and
why) and a **GREEN** cell (the output required after the change), per
verification-completeness §2. Where the defect a criterion guards against does not
exist on `c630de892` — it would be introduced by a shared store — the criterion
also names a **discriminating mutant** and states that the mutant outcome is derived
from the cited code, not executed at plan time; the run phase executes each named
mutant once and records the result in progress.md §E.2. Test names are suggestions;
the run phase may rename them but must keep one test per criterion traceable by the
AC id in its name or doc comment.

## Fixtures

**F — untracked `.moai` repository (the #1716 shape).** A git repository `P` with
`.moai/` listed in `.gitignore` and present on disk: `P/.moai/config/sections/workflow.yaml`
declares `workflow.audit.gates.codex: required` and `workflow.multi.review_gate.enabled: true`
and `workflow.codex.review_gate.enabled: true`; `P/.moai/specs/SPEC-PRI-001/spec.md`
exists. One commit. A linked worktree `W` made with `git worktree add`. `W` has no
`.moai` and is config-orphaned.

**F2 — two worktrees.** Fixture F plus a second linked worktree `W2` of `P`, also
without `.moai`.

**F-noid — primary not identifiable.** Fixture F where the primary checkout cannot be
identified, in two variants reused from SPEC-MCP-WORKTREE-UNTRACKED-001 AC-MWU-015 [REF]:
(a) a `PATH` from which git cannot be found; (b) the `HEAD` file inside `W`'s admin
directory `P/.git/worktrees/<W>/` deleted while `commondir` and `gitdir` stay, so the
scrubbed git inspection exits non-zero. As in that predecessor criterion, `W` carries
an empty `W/.moai/` directory (no `workflow.yaml`), so the validator accepts `W`
through its existing `.moai` branch without git while `W` stays config-orphaned;
without it the validator itself would reject `W` (REQ-MWU-005) and no state tool
would be reached.

**T — tracked `.moai` repository.** A repository that commits `.moai/config/sections/workflow.yaml`
declaring `workflow.audit.gates.codex: required`, with a linked worktree `WT`. `WT` is
not config-orphaned.

**B — bare non-git root.** A directory that is not a git repository and holds only
`.moai/` (the shape `newProbeProject` and the bare-`.moai` fixtures build); where a
criterion needs the receipt guard active, `B/.moai/config/sections/workflow.yaml`
declares `workflow.audit.gates.codex: required`.

Catalogue fixture SPECs carry parseable frontmatter with no `era:` field, so
`spec_audit` / `spec_drift` emit an `EraAutoDetected` INFO finding for each of them
(`internal/spec/audit.go:298-310`).

Git-configuration isolation (every fixture): git runs with `GIT_CONFIG_GLOBAL` and
`GIT_CONFIG_SYSTEM` pointed at empty files under `t.TempDir()`. Tests that set
`CLAUDE_PROJECT_DIR` are non-parallel and use `t.Setenv`.

Review-gate seam: the Stop review gates call `reviewGateChangeDetector(projectDir)`
after the opt-in check and ALLOW when it returns false
(`internal/cli/codex_review_gate.go:48, 74`, `internal/cli/multi_review_gate.go:73`).
Fixture F has no uncommitted change, so every criterion that drives a review gate
replaces that seam with a recorder that returns `true` and records each call's
argument; `codexLookPath` is stubbed to "not found" so the codex gate ALLOWs after
the detector (fail-open) without a real review.

## §D AC Matrix

### State writes

- **AC-WSR-001 (receipt destination and identity; REQ-WSR-002, REQ-WSR-003).**
  Given fixture F and the `codex_audit` runner test seam returning verdict `pass`,
  When `codex_audit` is called with `project_root = W`,
  Then exactly one receipt file exists under `P/.moai/state/audit-receipts/`, its
  `tree_root` equals canonical `W`, and the path `W/.moai` does not exist.
  RED-now: the receipt lands under `W/.moai/state/audit-receipts/` and `W/.moai`
  is created, because `receiptTarget` writes to the validated `project_root`
  itself (`internal/cli/mcp_audit_receipt.go:29-33`).
  Command: `go test ./internal/cli/ -run TestWSR001 -count=1`.

- **AC-WSR-002 (convergence destination, identity, and coexistence; REQ-WSR-002, REQ-WSR-003, REQ-WSR-007).**
  Cell 1 — Given fixture F and `audit_multi` stub backends (Claude `fail`, codex
  `pass`), When `audit_multi` is called with `project_root = W` and session id `S`,
  Then exactly one convergence result for `S` exists under `P/.moai/state/audit-multi/`,
  with `overall_verdict: fail` and tree identity canonical `W`, and `W/.moai` does
  not exist.
  Cell 2 — Given fixture F2, When `audit_multi` runs for session `S` with
  `project_root = W` (Claude `fail`, codex `pass`) and then with `project_root = W2`
  (Claude `pass`, codex `pass`), and `runMultiReviewGate` is then fed, in two
  separate calls with `CLAUDE_PROJECT_DIR` unset and the detector seam recording,
  (2a) `{"session_id":"S","cwd":W2}` and (2b) `{"session_id":"S","cwd":P}`, Then two
  convergence results for `S` exist in `P`'s store — tree identities `W` (`fail`)
  and `W2` (`pass`) — and the gate emits `"decision":"block"` in both 2a and 2b.
  Cell 2b is the reader half of REQ-WSR-007: a gate whose input resolves to `P`
  finds a `fail` that `audit_multi` wrote for `W`.
  RED-now: cell 1 — the file is written to `W/.moai/state/audit-multi/S.json` and
  carries no tree identity (`persistConvergenceResult` / `convergenceStateDirFor`,
  `internal/cli/mcp_convergence.go:903-927`); cell 2 — each result lands in its own
  worktree's store, and the gate on `W2` reads its opt-in flag from `W2`, where no
  `workflow.yaml` exists, so it emits `{}` (2a); the gate on `P` reads its flag from
  `P` (enabled) and calls the detector, but `P/.moai/state/audit-multi/S.json` does
  not exist, so it emits `{}` (2b).
  Discriminating mutant (the v0.1.0 design — shared store keyed by session id
  only): cell 2's second write replaces the first at one `S.json`, the store holds a
  single `pass` result, and the gate emits `{}` instead of blocking. Derived from
  the session-only path at `mcp_convergence.go:910`, not executed at plan time.
  Second discriminating mutant ("P gate reads only `S.json`" — `W`'s result is
  stored under the tree-qualified name of plan.md §F, but a gate resolved to a
  non-orphaned root still loads only `<store>/audit-multi/S.json`): cell 2b finds no
  `S.json` in `P`'s store and emits `{}` instead of blocking. Derived from
  `loadConvergenceResult` (`multi_review_gate.go:113-127`), not executed at plan time.
  Command: `go test ./internal/cli/ -run TestWSR002 -count=1`.

- **AC-WSR-003 (primary not identifiable on write and read; REQ-WSR-004, REQ-WSR-006).**
  Given fixture F-noid (each variant),
  When `verify_snapshot` is called with `project_root = W`, `key = k1`, and a
  `command`; `verify_trend` is called with `project_root = W` and `key = k1`;
  `codex_audit` (seam verdict `pass`) is called with `project_root = W`; and
  `audit_multi` (stub backends Claude `fail`, codex `pass`) is called with
  `project_root = W` and session `S`,
  Then `verify_snapshot` and `verify_trend` each return a tool error whose text
  names the primary checkout as unidentifiable; `codex_audit` returns the seam's
  verdict and `audit_multi` returns `overall_verdict: fail` (the value the stubs
  determine); no receipt file and no convergence result for `S` exists anywhere
  under `P` or `W`; each of the two audit results carries a notice naming the
  skipped write(s) and the cause; and `W/.moai` is still an empty directory.
  RED-now: `verify_snapshot`, `codex_audit`, and `audit_multi` write under
  `W/.moai/state/` and report no skip, and `verify_trend` reads `W` and returns
  success (`internal/cli/mcp_server.go:820-866`, `mcp_audit_receipt.go:42-57`,
  `mcp_convergence.go:768-770`).
  Command: `go test ./internal/cli/ -run TestWSR003 -count=1`.

- **AC-WSR-004 (other roots unchanged; REQ-WSR-005).**
  Given fixture F's `P`, fixture T's `WT`, and fixture B,
  When `codex_audit`, `audit_multi` (session `S`), and `verify_snapshot` record are
  each called with that root as `project_root`, once normally and once with a `PATH`
  from which git cannot be found,
  Then the receipt, the convergence result for `S`, and the snapshot land under that
  root's own `.moai/state` at the path the base tree uses, in every run, and no
  primary identification is attempted (the PATH-without-git run behaves
  identically).
  RED cell (regression guard): a mutant that applies the store-root mapping to every
  root that is a linked worktree moves `WT`'s files to T's primary checkout, and the
  test fails on that mutant. The base tree passes this criterion; it is a
  preservation check, and the mutant is its discriminating input.
  Predecessor tests: `go test ./internal/cli/ ./internal/hook/ ./internal/auditreceipt/ ./internal/verify/ -count=1`
  passes with exactly two existing tests edited, both named here —
  (i) `TestConfigOrphanedWorktree_PrimaryGateEnforced`
  (`internal/cli/mcp_project_root_worktree_gate_test.go:50`), whose premise at L61-64
  ("the first codex_audit call should have created W/.moai/state") becomes false by
  REQ-WSR-002: the premise is replaced by "`W/.moai` does not exist after the first
  call", and the second round creates `W/.moai/state` directly before repeating;
  (ii) `TestLinkedWorktree_DescriptionsStateTheGateSource`
  (`internal/cli/mcp_project_root_worktree_ac_test.go:296`), which asserts the phrases
  "still read from the accepted tree" (L300) and "that refusal is not guaranteed"
  (L315) that REQ-WSR-016 removes: those two assertions are replaced by the
  AC-WSR-015 phrases. `TestConfigOrphanedWorktree_CatalogueWarning` (AC-MWU-016 [REF])
  asserts only the presence of the `worktree_warning` key and the unchanged
  `_root.warning` text (L238, L256-L266), so it passes unedited. Any other existing
  test that fails is a blocker report, not an edit.
  Command: `go test ./internal/cli/ -run TestWSR004 -count=1`, plus the predecessor
  run above.

### Paired readers

- **AC-WSR-005 (verify tools and CLI agree; REQ-WSR-001, REQ-WSR-002, REQ-WSR-006).**
  Given fixture F with an uncommitted change in `W`,
  When `verify_snapshot` records check `c1` with `project_root = W` and key `k`
  (the `verify.Key` of `W`); the `moai verify record` command runs in-process with
  `--project-root W`, key `k`, and check `c2`; then `verify_snapshot` load and
  `verify_trend` are called with `project_root = W`, and the `moai verify check`
  command runs in-process with `--project-root W` (the flag form, so the test needs
  no `chdir`),
  Then exactly one snapshot file for `k` exists, under `P/.moai/state/verify/snapshots/`;
  both MCP reads return checks `c1` and `c2`; the CLI check reports the snapshot as
  present; and `W/.moai` does not exist.
  RED-now: both records land under `W/.moai/state/verify/snapshots/` (the MCP side
  through `resolveToolProjectRootWithSource`, the CLI through `verifyResolveRoot`,
  which absolutizes the flag and uses it directly — `internal/cli/verify.go:63-71`),
  so the predicates "file under `P`" and "`W/.moai` does not exist" fail.
  Command: `go test ./internal/cli/ -run TestWSR005 -count=1`.

- **AC-WSR-006 (review-gate root measurement matrix; REQ-WSR-007, REQ-WSR-008).**
  Given fixture F with a convergence result `overall_verdict: fail` for session `S`
  present in `P`'s store as `P`'s own record (the base-tree path
  `P/.moai/state/audit-multi/S.json`), a decoy result `overall_verdict: pass` for
  `S` at `W/.moai/state/audit-multi/S.json` (so `W/.moai` exists but holds no
  `workflow.yaml` — `W` stays config-orphaned), and the detector seam recording,
  When `runCodexReviewGate` and `runMultiReviewGate` are each fed, for every row of
  the matrix below, a stdin payload carrying `session_id: S` plus the row's
  `project_dir` and `cwd` values (a field marked — is omitted), with
  `CLAUDE_PROJECT_DIR` set to the row's value (— means unset),
  Then for every row each gate's observed outcome equals the row's **required**
  column, where an outcome is (i) whether the detector was called and with which
  argument — the detector runs only after the opt-in flag read, so a call proves the
  flag was read from a root whose config enables it, and only `P`'s does — and
  (ii) for the multi gate, the decision: `block` proves the result was read from
  `P`'s store (the `fail` record), and an allow after a detector call proves the
  decoy in `W`'s own store was read.

  Resolved root `R` = the first non-empty of stdin `project_dir`, stdin `cwd`,
  `CLAUDE_PROJECT_DIR` (`resolveProjectDirFromInput`, `internal/cli/codex_review_gate.go:236-246`).
  Outcome classes:
  - **N** — `R` empty: flag reader returns false (`readCodexReviewGateEnabled` /
    `readMultiReviewGateEnabled` on `""`); both gates `{}`, detector not called;
    no root read. Today and required.
  - **Wt** (today, `R = W`): flag read from `W` — absent — so both gates `{}` and the
    detector is not called; the state read is never reached. Config root `W`, state
    root none.
  - **Wr** (required, `R = W`): flags read from `P`; the detector is called once per
    gate with argument canonical `W` (the reviewed tree stays `R`); the multi gate
    reads store root `P` and emits `block`; the decoy in `W` is not read. Config root
    `P`, state root `P`.
  - **Pp** (`R = P`, today and required): flags read from `P`; the detector is called
    once per gate with `P`; the multi gate reads `P/.moai/state/audit-multi/` and
    emits `block`. Config root `P`, state root `P`.

  | Row | stdin `project_dir` | stdin `cwd` | `CLAUDE_PROJECT_DIR` | `R` | Today (`c630de892`) | Required |
  |---|---|---|---|---|---|---|
  | 1 | — | — | — | empty | N | N |
  | 2 | — | — | W | W | Wt | Wr |
  | 3 | — | — | P | P | Pp | Pp |
  | 4 | — | W | — | W | Wt | Wr |
  | 5 | — | W | W | W | Wt | Wr |
  | 6 | — | W | P | W | Wt | Wr |
  | 7 | — | P | — | P | Pp | Pp |
  | 8 | — | P | W | P | Pp | Pp |
  | 9 | — | P | P | P | Pp | Pp |
  | 10 | W | — | — | W | Wt | Wr |
  | 11 | W | — | W | W | Wt | Wr |
  | 12 | W | — | P | W | Wt | Wr |
  | 13 | W | W | — | W | Wt | Wr |
  | 14 | W | W | W | W | Wt | Wr |
  | 15 | W | W | P | W | Wt | Wr |
  | 16 | W | P | — | W | Wt | Wr |
  | 17 | W | P | W | W | Wt | Wr |
  | 18 | W | P | P | W | Wt | Wr |
  | 19 | P | — | — | P | Pp | Pp |
  | 20 | P | — | W | P | Pp | Pp |
  | 21 | P | — | P | P | Pp | Pp |
  | 22 | P | W | — | P | Pp | Pp |
  | 23 | P | W | W | P | Pp | Pp |
  | 24 | P | W | P | P | Pp | Pp |
  | 25 | P | P | — | P | Pp | Pp |
  | 26 | P | P | W | P | Pp | Pp |
  | 27 | P | P | P | P | Pp | Pp |

  F-noid rows (each variant, same seams): row 4's input (`cwd = W`, other two
  absent) yields **N′** today and required — both gates `{}`, detector not called
  (the flag stays disabled when the primary cannot be identified, REQ-WSR-008).
  RED-now: the 13 rows whose `R = W` (2, 4-6, 10-18) are red — today they produce
  **Wt** (`{}`, zero detector calls) where **Wr** is required. The "Today" column is
  code-derived from `codex_review_gate.go:191, 236-246` and
  `multi_review_gate.go:113-127, 162, 211-212`; the RED run observes it per row and
  records the observed matrix in progress.md §E.2 — that record is the measurement
  this criterion exists for. Rows 1, 3, 7-9, 19-27 and the F-noid rows are
  preservation cells (green today). Discriminating mutants: one that routes only the
  opt-in flag (state still read from `R`) turns every **Wr** row into an allow
  after a detector call (the decoy); one that routes the state read only turns every
  **Wr** row into **Wt**.
  Command: `go test ./internal/cli/ -run TestWSR006 -count=1` (non-parallel).

### Receipt guard

- **AC-WSR-007 (rejections are per tree in a shared store; REQ-WSR-003, REQ-WSR-006).**
  Given fixture F2,
  When, in order: a `plan-auditor` SubagentStart (agent `A1`, `cwd = W`) and its
  SubagentStop with a PASS verdict line for `SPEC-X` citing no receipt are handled;
  `codex_audit` (seam `pass`) runs with `project_root = P`; a `plan-auditor`
  SubagentStart (`A2`, `cwd = P`) and its SubagentStop with a PASS for `SPEC-Y`
  citing that receipt are handled; a `plan-auditor` start/stop pair (`A3`,
  `cwd = W2`) with an uncited PASS for `SPEC-X` is handled; and finally a PreToolUse
  `Agent` spawn of `manager-develop` is handled once with `cwd = W` and once with
  `cwd = P`,
  Then `A1`'s and `A3`'s stops are refused and `A2`'s is accepted; `P`'s store holds
  two rejection records for `plan-auditor` / `SPEC-X`, with tree identities
  canonical `W` and canonical `W2`; the spawn with `cwd = W` is denied with the
  `AUDIT_RECEIPT_VIOLATION` sentinel naming `SPEC-X`; and the spawn with `cwd = P` is
  allowed.
  RED-now: the guard is inactive on the config-orphaned `W` and `W2`
  (`auditReceiptTree` returns `""`, `internal/hook/audit_receipt_guard.go:52-60`),
  so `A1`'s and `A3`'s uncited PASSes are accepted, no rejection record exists, and
  the spawn with `cwd = W` is allowed.
  Discriminating mutant (the v0.1.0 design — shared store, rejections without tree
  identity): `A2`'s corroborated PASS clears every `plan-auditor` rejection in the
  store (`ClearRejectionsForRole`, `internal/auditreceipt/store.go:300`, called at
  `audit_receipt_guard.go:116`), deleting `A1`'s record — the silent cross-tree
  clearing; `A3` then writes a new record (the file name keys on agent type and SPEC
  ID only, `store.go:410-415`), so one record (`W2`'s) remains, and because
  `ListRejections` returns the whole store (`store.go:270`) the spawns with
  `cwd = W` and `cwd = P` are both denied — the "two records" and "`cwd = P`
  allowed" predicates fail. Derived from the cited code, not executed at plan time.
  Command: `go test ./internal/hook/ -run TestWSR007 -count=1`.

- **AC-WSR-008 (guard active on a config-orphaned worktree; REQ-WSR-003, REQ-WSR-006, REQ-WSR-009).**
  Given fixture F,
  When a `plan-auditor` SubagentStart hook input with `cwd = W` is handled, then a
  SubagentStop input for the same agent whose final message carries a PASS verdict
  line citing **no** receipt is handled,
  Then the stop is refused with the `AUDIT_RECEIPT_VIOLATION` sentinel; the start
  marker file exists under `P/.moai/state/audit-receipts/starts/` with tree identity
  canonical `W`; a rejection record exists under
  `P/.moai/state/audit-receipts/rejections/` with tree identity canonical `W`; and
  `W/.moai` does not exist;
  and When the same sequence cites the receipt id AC-WSR-001's `codex_audit` call
  returned (created after the start marker), Then the stop is accepted and the
  rejection with tree identity `W` is cleared.
  RED-now: `auditReceiptTree` returns `""` because `CodexGateRequired(W)` finds no
  `W/.moai/config/sections/workflow.yaml` (`internal/hook/audit_receipt_guard.go:52-60`,
  `internal/auditreceipt/store.go:155-166`), so the uncited PASS is accepted and no
  marker or rejection is written.
  Command: `go test ./internal/hook/ -run TestWSR008 -count=1`.

- **AC-WSR-009 (guard fail-closed without a store; REQ-WSR-010).**
  Given fixture F-noid (each variant), with `P`'s workflow config declaring **no**
  codex gate (so a refusal can only come from the fail-closed path),
  When the AC-WSR-008 uncited-PASS sequence is handled with `cwd = W`; then the same
  SubagentStop is handled again with `stop_hook_active: true`; then a PreToolUse
  `Agent` spawn of `manager-develop` is handled with `cwd = W`,
  Then the first stop is refused (`"decision":"block"`) with
  `AUDIT_RECEIPT_VIOLATION`; the re-entrant stop does not block and its system
  message carries `AUDIT_RECEIPT_VIOLATION`; the spawn is denied with
  `AUDIT_RECEIPT_VIOLATION`; each of the three texts states that the gate was
  assumed `required` because the primary checkout could not be identified; no
  `audit-receipts` directory exists under `P/.moai/state` and `W/.moai` is still an
  empty directory;
  and Given fixture T's `WT` with a workflow config declaring no codex gate, Then the
  same three inputs produce no refusal, no notice, and no denial (no fail-closed
  outside config-orphaned roots).
  RED-now: the guard is a no-op on `W`, so the first stop is accepted, the re-entrant
  stop emits nothing, and the spawn is allowed.
  Discriminating mutant (fail-closed at stop only, as v0.1.0 specified): the spawn
  is allowed, because with no store the re-entrant stop leaves no rejection for
  `checkAuditReceiptSpawn` (`audit_receipt_guard.go:198-222`) to find. Derived from
  the cited code, not executed at plan time.
  Command: `go test ./internal/hook/ -run TestWSR009 -count=1`.

### Catalogue

- **AC-WSR-010 (union and no silent vanish; REQ-WSR-011, REQ-WSR-013).**
  Given fixture F with `W/.moai/specs/SPEC-WTL-001/spec.md` written in the worktree,
  When `spec_progress`, `spec_drift`, and `spec_audit` are each called with
  `project_root = W`,
  Then `spec_progress` returns `count: 2` with `SPEC-WTL-001` (source `worktree`) and
  `SPEC-PRI-001` (source `primary`); `spec_drift` and `spec_audit` each report
  `total_specs: 2`; premise, asserted before the tag check — each of `spec_drift` and
  `spec_audit` returns at least one finding whose `spec_id` is `SPEC-WTL-001` and at
  least one whose `spec_id` is `SPEC-PRI-001` (the `EraAutoDetected` findings of the
  fixture); and every finding of both tools names its source, `worktree` for
  `SPEC-WTL-001` and `primary` for `SPEC-PRI-001`.
  Discriminating mutant: one that reads the primary catalogue only fails all three
  tools — `SPEC-WTL-001` is absent from `spec_progress`, `total_specs` is 1 for
  `spec_drift` and `spec_audit`, and the `SPEC-WTL-001` finding premise fails.
  RED-now: all three tools read `W` only (`internal/cli/mcp_server.go:802-907`):
  `spec_progress` returns `SPEC-WTL-001` only with no source field, `spec_drift` and
  `spec_audit` report `total_specs: 1`, and no `SPEC-PRI-001` finding exists.
  Command: `go test ./internal/cli/ -run TestWSR010 -count=1`.

- **AC-WSR-011 (same SPEC in both; REQ-WSR-012).**
  Given fixture F plus `SPEC-DUP-001` present in both catalogues with
  `status: draft` in `W` and `status: completed` in `P`,
  When `spec_progress`, `spec_drift`, and `spec_audit` are called with
  `project_root = W`,
  Then `spec_progress` lists `SPEC-DUP-001` once, with status `draft` and source
  `worktree`, and its `count` equals the number of distinct SPEC IDs (2);
  `spec_drift` and `spec_audit` report `total_specs: 2`, and their `SPEC-DUP-001`
  findings name source `worktree` only; and each of the three responses names the
  shadowed primary copy of `SPEC-DUP-001`.
  RED-now: every response carries no shadowed-copy field (only `W`'s catalogue is
  read), so the "names the shadowed primary copy" predicate fails for all three.
  Command: `go test ./internal/cli/ -run TestWSR011 -count=1`.

- **AC-WSR-012 (primary not identifiable on read; REQ-WSR-014).**
  Given fixture F-noid (each variant) with `SPEC-WTL-001` in `W`,
  When `spec_progress`, `spec_drift`, and `spec_audit` are called with
  `project_root = W`,
  Then each call succeeds; `spec_progress` returns `SPEC-WTL-001` and `spec_drift` /
  `spec_audit` report `total_specs: 1`; and each response's `_root` block states that
  the primary catalogue was not read and names the cause.
  RED-now: no `_root` block carries a statement about the primary catalogue.
  Command: `go test ./internal/cli/ -run TestWSR012 -count=1`.

- **AC-WSR-013 (`_root` provenance; REQ-WSR-015).**
  Given fixture F,
  When `spec_progress`, `spec_audit`, `spec_drift`, `verify_snapshot`, and
  `verify_trend` are called with `project_root = W`, and `spec_progress` is called
  again with no `project_root` and (non-parallel) `CLAUDE_PROJECT_DIR = W`,
  Then every response's `_root` block lists the sources read (catalogue tools: both
  `worktree` and `primary`; state tools: the store root `P`), its `worktree_warning`
  does not contain the phrase "read from the worktree tree", and in the fallback call
  the `_root.warning` key and text are byte-identical to the base tree's.
  RED-now: `worktreeWarning` contains "this answer was read from the worktree tree"
  (`internal/cli/mcp_worktree_root.go:315-317`) and no sources list exists. The new
  text is asserted by this criterion's own test; the predecessor
  `TestConfigOrphanedWorktree_CatalogueWarning` checks only that the key is present
  and the `_root.warning` text is unchanged, and runs here unedited as the
  preservation half.
  Command: `go test ./internal/cli/ -run 'TestWSR013|TestConfigOrphanedWorktree_CatalogueWarning' -count=1`
  (swept count asserted: the run reports both tests, not `[no tests to run]`).

### Shared mapping, documentation, end to end

- **AC-WSR-014 (one answer everywhere; REQ-WSR-001).**
  Given fixtures F, F-noid, T, and B (B carrying the `required` gate declaration),
  When, for each root `R` in {`P`, `W`, `W` under F-noid, `WT`, T's primary, `B`},
  three call sites act on `R`: the MCP writer (`codex_audit` with
  `project_root = R`, seam `pass`), the hook writer (a `plan-auditor` SubagentStart
  with `cwd = R`), and the CLI writer (`moai verify record --project-root R`, key
  `k`, check `c`),
  Then the three agree per row on where they wrote, or on refusing to: `P`, `P`,
  unresolved (codex_audit's skip notice, no start marker anywhere, and a CLI error,
  each naming the unidentifiable primary), `WT`, T's primary, and `B` respectively
  — the receipt, the start marker, and the snapshot all under that row's
  `.moai/state`.
  RED-now: on the `W` row the MCP receipt and the CLI snapshot land under `W` and the
  hook writes no start marker (guard inert on a config-orphaned `W`), so the three
  disagree and none is under `P`; on the F-noid row the MCP and CLI writes land
  under `W`.
  Command: `go test ./internal/auditreceipt/ ./internal/hook/ ./internal/cli/ -run TestWSR014 -count=1`.

- **AC-WSR-015 (documentation; REQ-WSR-016).**
  Given the change,
  When `diff` compares `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` with its
  template mirror, the template neutrality guards run, and a doc test reads the
  catalogue section, the shared `project_root` description, and the registered
  descriptions of `spec_progress`, `spec_drift`, `spec_audit`, `verify_snapshot`,
  `verify_trend`, `codex_audit`, and `audit_multi`,
  Then `diff` exits 0; the guards pass; the catalogue section states the
  primary-checkout state store keyed by the worktree's tree identity, the union
  catalogue, and the guard's primary gate read; the shared `project_root`
  description no longer contains "still read from the accepted tree"; each `spec_*`
  description contains "primary checkout" and "union"; each of the four state and
  audit descriptions contains "primary checkout" and ".moai/state"; the `codex_audit`
  description no longer contains "that refusal is not guaranteed"; and the template
  copy contains no `SPEC-`, card id, or ISO date.
  RED-now: the section says "Other configuration, the SPEC catalogue, and state are
  still read from the accepted tree" (catalogue file lines 176-177); the shared
  description contains "still read from the accepted tree"
  (`internal/cli/mcp_project_root.go:50`); and the `codex_audit` description contains
  "that refusal is not guaranteed" (`internal/cli/mcp_server.go:317`).
  Command: `go test ./internal/cli/ -run TestWSR015 -count=1` and
  `diff .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md`.

- **AC-WSR-016 (end to end in the untracked-`.moai` shape; REQ-WSR-002, -006, -009, -011).**
  Given fixture F with no `.moai` in `W`,
  When, in order, an auditor start is handled with `cwd = W`; `codex_audit` runs with
  `project_root = W` (seam `pass`); the auditor stop cites the returned receipt;
  `spec_progress` runs with `project_root = W`,
  Then the stop is accepted, `spec_progress` lists `SPEC-PRI-001`, and after the whole
  sequence the path `W/.moai` still does not exist.
  RED-now: the guard is inactive (no marker), the receipt and a `W/.moai` directory are
  created under `W`, and `spec_progress` does not list `SPEC-PRI-001`.
  Command: `go test ./internal/cli/ -run TestWSR016 -count=1`.

## §D.1 Edge cases

- `W` gains `W/.moai/specs` because a SPEC was written there: `W` stays config-orphaned
  (the predicate keys on `workflow.yaml`), so AC-WSR-010's union still applies.
- Verify keys are content hashes of a tree's state (`internal/verify/key.go:39-70`).
  Two trees in identical state share a key in the shared store; this is accepted
  (plan.md §B risk R3) and not tested as a defect.
- A receipt guard input whose `cwd` names `P` while the audit ran on `W` refuses the
  receipt as "another tree", and a spawn whose input names a different tree than
  the refused stop does not see that refusal (spec.md §5 cases 1-2). Not acceptance
  criteria: which value the host sends is unmeasured; AC-WSR-006 measures the Stop
  review gates' input-to-root mapping, and the optional M0 probe (plan.md §D)
  captures the real payload.
- A session that audited several trees must clear each tree's `fail` before its
  multi-review gate allows (REQ-WSR-007); AC-WSR-002 cell 2 pins the block, and its cell 2b pins that a gate resolved to
  `P` finds the `fail` written for `W`.

## §D.2 Definition of Done

- All sixteen criteria pass at the listed commands on the run-phase head.
- RED-first witnessed for AC-WSR-001, -002, -006, -007, -008, -010, -014, -016: the RED
  commit is an ancestor of the fix, touches only `_test.go` files, and its test run
  exits non-zero on the named predicate rather than on a build error.
- Every discriminating mutant named above (AC-WSR-002, -004, -006, -007, -009, -010) is
  executed once in the run phase and its failing predicate recorded in progress.md §E.2.
- AC-WSR-006's observed per-row matrix on the RED commit is recorded in progress.md §E.2.
- `go vet` and `golangci-lint run` clean on the touched packages; `make build` run after
  the rule/template edit.
- Full-suite verdict comes from CI on the develop push, not a local `go test ./...`.
