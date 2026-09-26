# SPEC-WORKTREE-STATE-ROOT-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`. Every
criterion names a **RED-now** cell (the input that fails on base `c630de892`, and
why) and a **GREEN** cell (the output required after the change), per
verification-completeness §2. Test names are suggestions; the run phase may rename
them but must keep one test per criterion traceable by the AC id in its name or doc
comment.

## Fixtures

**F — untracked `.moai` repository (the #1716 shape).** A git repository `P` with
`.moai/` listed in `.gitignore` and present on disk: `P/.moai/config/sections/workflow.yaml`
declares `workflow.audit.gates.codex: required` and `workflow.multi.review_gate.enabled: true`
and `workflow.codex.review_gate.enabled: true`; `P/.moai/specs/SPEC-PRI-001/spec.md`
exists. One commit. A linked worktree `W` made with `git worktree add`. `W` has no
`.moai` and is config-orphaned.

**F-noid — primary not identifiable.** Fixture F where the primary checkout cannot be
identified, in two variants reused from SPEC-MCP-WORKTREE-UNTRACKED-001 AC-MWU-015:
(a) a `PATH` from which git cannot be found; (b) the `HEAD` file inside `W`'s admin
directory `P/.git/worktrees/<W>/` deleted while `commondir` and `gitdir` stay, so the
scrubbed git inspection exits non-zero. As in that predecessor criterion, `W` carries
an empty `W/.moai/` directory (no `workflow.yaml`), so the validator accepts `W`
through its existing `.moai` branch without git while `W` stays config-orphaned;
without it the validator itself would reject `W` (REQ-MWU-005) and no state tool
would be reached.

**T — tracked `.moai` repository.** A repository that commits `.moai/config/sections/workflow.yaml`,
with a linked worktree `WT`. `WT` is not config-orphaned.

**B — bare non-git root.** A directory that is not a git repository and holds only
`.moai/` (the shape `newProbeProject` and the bare-`.moai` fixtures build).

Git-configuration isolation (every fixture): git runs with `GIT_CONFIG_GLOBAL` and
`GIT_CONFIG_SYSTEM` pointed at empty files under `t.TempDir()`. Tests that set
`CLAUDE_PROJECT_DIR` are non-parallel and use `t.Setenv`.

## §D AC Matrix

### State writes

- **AC-WSR-001 (receipt destination and identity; REQ-WSR-002, REQ-WSR-003).**
  Given fixture F and the `codex_audit` runner test seam returning verdict `pass`,
  When `codex_audit` is called with `project_root = W`,
  Then exactly one receipt file exists under `P/.moai/state/audit-receipts/`, its
  `tree_root` equals canonical `W`, and the path `W/.moai` does not exist.
  RED-now: the receipt lands under `W/.moai/state/audit-receipts/` and `W/.moai`
  is created, because `receiptTarget` writes to the validated `project_root`
  itself (`internal/cli/mcp_audit_receipt.go:30-33`).
  Command: `go test ./internal/cli/ -run TestWSR001 -count=1`.

- **AC-WSR-002 (convergence state destination; REQ-WSR-002).**
  Given fixture F and `audit_multi` stub backends (Claude `fail`, codex `pass`),
  When `audit_multi` is called with `project_root = W` and session id `S`,
  Then `P/.moai/state/audit-multi/S.json` exists with `overall_verdict: fail`, and
  `W/.moai` does not exist.
  RED-now: the file is written to `W/.moai/state/audit-multi/S.json`
  (`convergenceStateDirFor`, `internal/cli/mcp_convergence.go:922-927`).
  Command: `go test ./internal/cli/ -run TestWSR002 -count=1`.

- **AC-WSR-003 (primary not identifiable on write and read; REQ-WSR-004, REQ-WSR-006).**
  Given fixture F-noid (each variant),
  When `verify_snapshot` is called with `project_root = W`, `key = k1`, and a
  `command`, `verify_trend` is called with `project_root = W` and `key = k1`, and
  `codex_audit` (seam verdict `pass`) is called with `project_root = W`,
  Then `verify_snapshot` and `verify_trend` each return a tool error whose text
  names the primary checkout as unidentifiable; `codex_audit` returns the same verdict as the seam, no receipt
  file exists anywhere under `P` or `W`, and its result carries a notice naming the
  skipped receipt write; and `W/.moai` is still an empty directory.
  RED-now: `verify_snapshot` and `codex_audit` write under `W/.moai/state/` and
  report success, and `verify_trend` reads `W` and returns success
  (`internal/cli/mcp_server.go:820-866`, `mcp_audit_receipt.go:42-57`).
  Command: `go test ./internal/cli/ -run TestWSR003 -count=1`.

- **AC-WSR-004 (other roots unchanged; REQ-WSR-005).**
  Given fixture T's `WT` and fixture B,
  When `codex_audit`, `audit_multi` (session `S`), and `verify_snapshot` record are
  each called with that root as `project_root`, once normally and once with a `PATH`
  from which git cannot be found,
  Then the receipt, `audit-multi/S.json`, and snapshot files land under that root's
  own `.moai/state` in every run, exactly as on the base tree, and no primary
  identification is attempted (the PATH-without-git run behaves identically).
  RED cell (regression guard): a mutant that applies the store-root mapping to every
  root that is a linked worktree moves `WT`'s files to T's primary checkout, and the
  test fails on that mutant. The base tree passes this criterion; it is a
  preservation check, and the mutant is its discriminating input.
  Command: `go test ./internal/cli/ -run TestWSR004 -count=1`, plus the existing
  `go test ./internal/cli/ ./internal/auditreceipt/ ./internal/verify/ -count=1`
  passing unchanged except for the test REQ-WSR-015 updates (AC-WSR-013).

### Paired readers

- **AC-WSR-005 (verify tools and CLI agree; REQ-WSR-001, REQ-WSR-006).**
  Given fixture F with an uncommitted change in `W`,
  When `verify_snapshot` records check `c1` with `project_root = W` and key `k`
  (the `verify.Key` of `W`), then `verify_snapshot` load and `verify_trend` are
  called with `project_root = W`, and the `moai verify check` command runs in-process
  with `--project-root W` (the flag form, so the test needs no `chdir`),
  Then the snapshot file exists under `P/.moai/state/verify/snapshots/`, both MCP
  reads return check `c1`, and the CLI reports the snapshot as present.
  RED-now: the record lands under `W/.moai/state/verify/snapshots/`; the MCP reads
  on `W` find it, but the file is not under `P`, and the CLI reads `W` itself
  (`internal/cli/verify.go` `verifyResolveRoot` absolutizes the flag and uses it
  directly), so the GREEN predicate "file under `P`" fails.
  Command: `go test ./internal/cli/ -run TestWSR005 -count=1`.

- **AC-WSR-006 (multi-review-gate finds the result from either root; REQ-WSR-007, REQ-WSR-008).**
  Given the state AC-WSR-002 leaves (`overall_verdict: fail` for session `S` on `W`),
  When `runMultiReviewGate` is fed the stdin payload `{"session_id":"S","cwd":W}`,
  and separately `{"session_id":"S","cwd":P}`, with `CLAUDE_PROJECT_DIR` unset,
  Then both runs emit `"decision":"block"`.
  RED-now, both cells: with `cwd = W` the opt-in flag is read from `W`, where no
  `workflow.yaml` exists, so the gate is disabled and emits `{}`
  (`internal/cli/multi_review_gate.go:162, 211-212`); with `cwd = P` the flag is on
  but `loadConvergenceResult` looks under `P/.moai/state/audit-multi/` while the
  writer used `W` (`multi_review_gate.go:113-127`), so it emits `{}`.
  Command: `go test ./internal/cli/ -run TestWSR006 -count=1`.

- **AC-WSR-007 (review-gate opt-in flags; REQ-WSR-008).**
  Given fixture F with the codex review change-detector seam counting its calls, and
  fixture F-noid,
  When `runCodexReviewGate` is fed `{"cwd":W}` on F, and `runMultiReviewGate` and
  `runCodexReviewGate` are fed `{"session_id":"S","cwd":W}` on F-noid,
  Then on F the change detector is called once (the gate passed its opt-in check);
  on F-noid both gates emit `{}` and the change detector is called zero times.
  RED-now: on F the flag is read from `W`, absent, so the detector is called zero
  times (`internal/cli/codex_review_gate.go:191`, `readCodexReviewGateEnabled` in
  `internal/cli/mcp_codex.go:2293`).
  Command: `go test ./internal/cli/ -run TestWSR007 -count=1`.

### Receipt guard

- **AC-WSR-008 (guard active on a config-orphaned worktree; REQ-WSR-006, REQ-WSR-009).**
  Given fixture F,
  When a `plan-auditor` SubagentStart hook input with `cwd = W` is handled, then a
  SubagentStop input for the same agent whose final message carries a PASS verdict
  line citing **no** receipt is handled,
  Then the stop is refused with the `AUDIT_RECEIPT_VIOLATION` sentinel, and the start
  marker file exists under `P/.moai/state/audit-receipts/` with tree identity
  canonical `W`;
  and When the same sequence cites the receipt id AC-WSR-001's `codex_audit` call
  returned (created after the start marker), Then the stop is accepted.
  RED-now: `auditReceiptTree` returns `""` because `CodexGateRequired(W)` finds no
  `W/.moai/config/sections/workflow.yaml` (`internal/hook/audit_receipt_guard.go:52-60`,
  `internal/auditreceipt/store.go:155-166`), so the uncited PASS is accepted and no
  marker is written.
  Command: `go test ./internal/hook/ -run TestWSR008 -count=1`.

- **AC-WSR-009 (guard fail-closed with a named cause; REQ-WSR-010).**
  Given fixture F-noid (each variant), with `P`'s workflow config declaring **no**
  codex gate (so a refusal can only come from the fail-closed path),
  When the AC-WSR-008 uncited-PASS sequence is handled with `cwd = W`,
  Then the stop is refused with `AUDIT_RECEIPT_VIOLATION`, and the reason text states
  that the gate was assumed `required` because the primary checkout could not be
  identified;
  and Given fixture T's `WT` with the same no-gate config, Then the same sequence is
  accepted (no fail-closed outside config-orphaned roots).
  RED-now: the guard is a no-op on `W`, so the PASS is accepted.
  Command: `go test ./internal/hook/ -run TestWSR009 -count=1`.

### Catalogue

- **AC-WSR-010 (union and no silent vanish; REQ-WSR-011, REQ-WSR-013).**
  Given fixture F with `W/.moai/specs/SPEC-WTL-001/spec.md` written in the worktree,
  When `spec_progress`, `spec_drift`, and `spec_audit` are each called with
  `project_root = W`,
  Then `spec_progress` returns both `SPEC-WTL-001` (source `worktree`) and
  `SPEC-PRI-001` (source `primary`); every `spec_drift` / `spec_audit` finding for
  either SPEC names its source; and a mutant that reads the primary catalogue only
  fails this criterion because `SPEC-WTL-001` is absent.
  RED-now: `spec_progress` returns `SPEC-WTL-001` only, with no source field, and
  `SPEC-PRI-001` is absent (catalogue read from `W` only, `internal/cli/mcp_server.go:802`).
  Command: `go test ./internal/cli/ -run TestWSR010 -count=1`.

- **AC-WSR-011 (same SPEC in both; REQ-WSR-012).**
  Given fixture F plus `SPEC-DUP-001` present in both catalogues with
  `status: draft` in `W` and `status: completed` in `P`,
  When `spec_progress` is called with `project_root = W`,
  Then `SPEC-DUP-001` appears once, with status `draft` and source `worktree`; the
  response names the shadowed primary copy; and `count` equals the number of distinct
  SPEC IDs.
  RED-now: the response carries no shadowed-copy field (only `W`'s catalogue is read),
  so the "names the shadowed primary copy" predicate fails.
  Command: `go test ./internal/cli/ -run TestWSR011 -count=1`.

- **AC-WSR-012 (primary not identifiable on read; REQ-WSR-014).**
  Given fixture F-noid (each variant) with `SPEC-WTL-001` in `W`,
  When `spec_progress` is called with `project_root = W`,
  Then the call succeeds, returns `SPEC-WTL-001`, and the `_root` block states that
  the primary catalogue was not read and names the cause.
  RED-now: the `_root` block carries no statement about the primary catalogue.
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
  (`internal/cli/mcp_worktree_root.go:315-317`) and no sources list exists. The
  predecessor test for AC-MWU-016, which asserts the old text, is updated in the same
  change.
  Command: `go test ./internal/cli/ -run 'TestWSR013|WorktreeWarning' -count=1`.

### Shared mapping, documentation, end to end

- **AC-WSR-014 (one mapping; REQ-WSR-001).**
  Given fixtures F, F-noid, T, and B,
  When the store-root resolution is evaluated on `P`, `W`, `W` under F-noid, `WT`,
  T's primary, and B,
  Then it returns `P`, `P`, an error naming the unidentifiable primary, `WT`, T's
  primary, and B respectively; and the hook package and the CLI package reach it
  through the same exported function (a test in each package calls it and a
  `go vet ./internal/hook/ ./internal/cli/` run is clean).
  RED-now: no such resolution exists, so the test does not compile — the RED commit
  therefore carries the function signature with a stub body returning its input, and
  the RED run fails on the `W → P` row, not on a build error.
  Command: `go test ./internal/auditreceipt/ ./internal/hook/ ./internal/cli/ -run TestWSR014 -count=1`.

- **AC-WSR-015 (documentation; REQ-WSR-016).**
  Given the change,
  When `diff` compares `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` with its
  template mirror, the template neutrality guards run, and a doc test reads the
  catalogue section and the tool descriptions,
  Then `diff` exits 0; the guards pass; the section states the primary-checkout state
  store keyed by the worktree's tree identity, the union catalogue, and the guard's
  primary gate read; the `codex_audit` description no longer contains "that refusal is
  not guaranteed"; and the template copy contains no `SPEC-`, card id, or ISO date.
  RED-now: the section says "Other configuration, the SPEC catalogue, and state are
  still read from the accepted tree" (catalogue file line 176-177) and the
  description contains "that refusal is not guaranteed" (`internal/cli/mcp_server.go:317`).
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
  receipt as "another tree" (spec.md §5). Not an acceptance criterion: which value the
  host sends is unmeasured.

## §D.2 Definition of Done

- All sixteen criteria pass at the listed commands on the run-phase head.
- RED-first witnessed for AC-WSR-001, -006, -008, -010, -016: the RED commit is an
  ancestor of the fix, touches only `_test.go` files (or, for AC-WSR-014, a stub
  signature plus tests), and its test run exits non-zero on the named predicate rather
  than on a build error.
- `go vet` and `golangci-lint run` clean on the touched packages; `make build` run after
  the rule/template edit.
- Full-suite verdict comes from CI on the develop push, not a local `go test ./...`.
