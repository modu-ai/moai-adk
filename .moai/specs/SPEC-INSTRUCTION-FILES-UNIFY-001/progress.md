# SPEC-INSTRUCTION-FILES-UNIFY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: L (5 artifacts + progress.md). Artifacts written: spec.md, plan.md, acceptance.md,
  design.md, research.md.
- SPEC ID regex check executed as Bash, output `PASS`.
- **v0.3.0 — B1/B2 carve applied (operator decision, 2026-09-26).** Requirements: **16**
  (`REQ-IFU-001~006`, `013~019`, `023~025`; Tier L ceiling 25). Acceptance criteria: **20**
  (ceiling 25). The nine requirements touching a user-owned file (`REQ-IFU-007~012`,
  `020~022`) and the seven criteria covering them transferred verbatim to
  **`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`** (card **t1259** owns its plan phase). Nothing was
  deleted. Ids were deliberately NOT renumbered — the gaps are the carve's footprint (spec.md
  HISTORY).
- **The carve's cause is arithmetic, not the recorded debt.** The plan-audit judged the debt on
  the two folded criteria acceptable at the 0.85 threshold. What forced the split is that its D2
  finding needs two new criteria while the SPEC stood at 25/25 with no tier above L.
- Frontmatter repaired to the canonical 12-field schema (audit MP-3): `tags` quoted as a
  comma-separated string, `module` and `lifecycle` added, `title` / `version` quoted. `moai spec
  lint` now exits `0`.
- Five criteria that were vacuously passable (audit D1) repaired against verified symbols, and
  every test-invoking criterion now asserts `--- PASS: <TestName>` under `-v` rather than exit
  `0` alone. `AC-IFU-016`'s guard does not exist in `internal/hook` (verified) and the criterion
  is written to require its creation.
- Two criteria added for the previously uncovered requirements (audit D2): `AC-IFU-026`
  (`REQ-IFU-001`, contract deployment) and `AC-IFU-027` (`REQ-IFU-005`, launcher invariance).
  §D.2's traceability table is now derived from criterion bodies and carries a re-runnable
  verification command instead of a coverage claim.
- `AC-IFU-022` gained `CONTRACT_HEAD` as its positive control (audit D5): its absence makes a
  run INCONCLUSIVE, so a render that captured nothing can no longer resolve to the benign
  branch.
- The `AGENTS.md` mirror-divergence figure re-measured (audit D4): 57 template-only / 17
  root-only lines against `553e224f3` on 2026-09-26. The t925-era "46 lines" is retired.
- Source locations converted to symbol anchors throughout (the audit's staleness warning:
  `origin/develop` is ~93 commits ahead and card t1224 already moved
  `frozenInstructionFiles`).
- `AC-IFU-025` gained the always-loaded-budget clause (audit item 6) — root `AGENTS.md` sits
  inside that surface, card t1175 is retuning it concurrently, and neither named ceiling would
  have caught an overrun.
- The §A.5 confirmed-branch transfer to card t1219 is now asserted as a conditional Definition
  of Done item (audit D7) rather than resting on prose in three files.
- Status: `draft`. Plan phase only — no implementation, no commits by this agent.
- Run phase is blocked on card t1175 landing on develop (plan.md §B).
- Two run-phase measurements remain mandatory and carry criteria whichever way they come out:
  AC-IFU-021 (real linked worktree, ancestor discovery) and AC-IFU-022 (Codex discovery of
  `AGENTS.local.md`, with head and tail sentinels so silent truncation is observable).
- AC-IFU-022's command verb was verified in the plan phase, not inherited: `codex debug --help`
  against **codex-cli 0.157.0** lists `prompt-input` ("Render the model-visible prompt input
  list as JSON"). A run under a different codex-cli version re-confirms before relying on it.
- AC-IFU-004 asserts the Codex-discovered **filename set** recursively over the template tree,
  not a path the criterion already expects — a `-maxdepth 1` or path-presence form would keep
  passing after a rename back to `AGENTS.md`.

## §F Phase 4 Mode Selection

- **Decision: serial.** Tier L, coding-heavy, and the three milestones are dependent (M1's
  measurements gate M2's start and feed M3's worktree paragraph and byte budget), so neither
  fan-out nor an agent team buys parallelism. Orchestrator decision, recorded at M1.

## §E.2 Run-phase Evidence

### M1a — AC-IFU-021 real-worktree ancestor discovery

Full transcript, fixture description, and raw outputs: `.moai/reports/t1243/m1/evidence.md`,
`.moai/reports/t1243/m1/m1a-{nested,sibling,primary}.{out,err}`. Tool: `claude 2.1.283`.

Fixture: throwaway repo in the session scratchpad; committed `CLAUDE.md` with
`WT_CONTROL = OSCAR4` (positive control); untracked, gitignored, primary-root-only
`AGENTS.local.md` (`LOCAL_AGENTS_TOKEN = PAPA6`) and `CLAUDE.local.md`
(`LOCAL_CLAUDE_TOKEN = QUEBEC2`). Two real linked worktrees via `git worktree add` — nested at
`.claude/worktrees/wt` (this repository's geometry) and a sibling `../sibling-wt` — each with
`.git` as a file (`-rw-r--r-- ... 139 ... primary/.claude/worktrees/wt/.git`,
`gitdir: .../primary/.git/worktrees/wt`).

Command (all three runs, cwd varies):
`unset MOAI_KANBAN ... MOAI_KANBAN_SETTINGS_INJECTED && cd <cwd> && timeout 180 claude -p "Answer only from your loaded instructions, run no tools. Three lines: (1) exact value of WT_CONTROL (write NOT_PRESENT if absent), (2) exact value of LOCAL_AGENTS_TOKEN (write NOT_PRESENT if absent), (3) exact value of LOCAL_CLAUDE_TOKEN (write NOT_PRESENT if absent)." --model claude-haiku-4-5-20251001`

| cwd | exit | verbatim output |
|---|---|---|
| nested worktree `primary/.claude/worktrees/wt` | 0 | `1. OSCAR4` / `2. NOT_PRESENT` / `3. QUEBEC2` |
| sibling worktree `sibling-wt` | 0 | `OSCAR4` / `NOT_PRESENT` / `NOT_PRESENT` |
| primary root (control) | 0 | `1. WT_CONTROL = OSCAR4` / `2. LOCAL_AGENTS_TOKEN = NOT_PRESENT` / `3. LOCAL_CLAUDE_TOKEN = QUEBEC2` |

stderr empty in all three.

**Outcome — design.md §A.5 row 2: discovery confirmed for `CLAUDE.local.md` but NOT for
`AGENTS.local.md`.** `CLAUDE.local.md` is reached from a real nested linked worktree by a
directory-ancestor walk (the sibling worktree does not get it, so the walk is not
git-worktree-aware). `AGENTS.local.md` is not discovered by Claude Code at all — not even at the
primary root. Per §A.5, Option 3 degrades to Option 1: **known limitation — a worktree session does
not receive `AGENTS.local.md` content** (the `@AGENTS.local.md` import is out-of-project there per
M0-3 and silently skipped per M0-1, and no filename discovery reaches it). Not retried.

**acceptance.md §D.3 conditional handoff: NOT triggered** — the condition is confirmation for
`AGENTS.local.md`, and that was not observed. Informational only (not a Definition-of-Done
transfer): the real-worktree `CLAUDE.local.md` result is direct evidence for the mechanism behind
card t1219 item (1); the lead may forward `.moai/reports/t1243/m1/evidence.md` to t1219.

### M1b — AC-IFU-022 Codex discovery of `AGENTS.local.md`

Full transcript: `.moai/reports/t1243/m1/evidence.md`; project-doc segment
`.moai/reports/t1243/m1/m1b-project-doc-excerpt.txt`; stderr `m1b-prompt-input.err` (empty). Full
render not committed (it embeds the operator's personal global `~/.codex/AGENTS.md`); sha256
`eea9b78fc070dc2b1a1d0a5cdcc1767685d519879aba847c90837a0d0c3f2a09`, 40,358 B.

Verb re-confirmed on this version: `codex debug --help` →
`prompt-input  Render the model-visible prompt input list as JSON`. `~/.codex/config.toml` has no
`fallback` / `project_doc` key; `CODEX_HOME` unset.

```
$ cd <fixture> && codex --version && timeout 120 codex debug prompt-input > ../prompt-input.json 2> ../prompt-input.err
codex-cli 0.157.0
exit=0
   40358 ../prompt-input.json
       0 ../prompt-input.err
CONTRACT_HEAD count=1
CONTRACT_TAIL count=1
LOCAL_HEAD count=0
LOCAL_TAIL count=0
```

**Outcome (positive control first):** `CONTRACT_HEAD` present → conclusive. `LOCAL_HEAD` absent →
**Codex does not discover `AGENTS.local.md`; the design's budget arithmetic holds.** Not a blocker.
`CONTRACT_TAIL` present → no truncation of the fixture `AGENTS.md`.

**Gate:** both measurements' evidence is on disk; M1b is negative, so M2 is gated open and M3's
byte budget is unchanged by M1.

### M2 — Codex read order, guard set, learner target

Base: `9f32f8077`. Raw outputs: `.moai/reports/t1243/m2/` (`red-*.txt` = RED before GREEN,
`ac-ifu-*.txt` = AC commands after GREEN, `pkg-*.txt` = wider package runs, `build-*.txt`,
`lint*.txt`). All test runs used `unset MOAI_KANBAN ... MOAI_KANBAN_SETTINGS_INJECTED && go test ... -count=1 -v`.

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-IFU-010 | PASS | `go test ./internal/cli/ -run '^TestCodexLocalInstructions_AgentsLocalReadFirst$\|^TestCodexLocalInstructions_DualFileMatrix$' -v` | `--- PASS: TestCodexLocalInstructions_AgentsLocalReadFirst (0.00s)` / `--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.01s)` / `ok ... internal/cli 1.017s`; `no tests to run` count 0 |
| AC-IFU-016 | PASS | `grep -n 'frozenInstructionFiles = ' internal/hook/pre_tool.go` + `go test ./internal/hook/ -run '^TestFrozenInstructionFiles$' -v` | `1405:var frozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "AGENTS.local.md"}`; `--- PASS: TestFrozenInstructionFiles (0.00s)` with four sub-cases `AGENTS.local.md`, `AGENTS.md`, `CLAUDE.local.md`, `CLAUDE.md`; `ok ... internal/hook 0.560s` |
| AC-IFU-017 | PASS | `go test ./internal/harness/curator/ -run '^TestSurfaceForTier_Tier3$\|^TestPrepareTierDispatch_Tier3$' -v` | `--- PASS: TestSurfaceForTier_Tier3 (0.00s)` / `--- PASS: TestPrepareTierDispatch_Tier3 (0.00s)`; `dispatch.go:44: 3: {Path: "AGENTS.local.md", ...}` |

RED (before GREEN, same commands): launcher — `ALPHA (AGENTS.local.md) at 72 must precede BETA
(CLAUDE.local.md) at 33` and `--- FAIL` for all three read-order tests; hook — `frozenInstructionFiles
= [CLAUDE.md CLAUDE.local.md], want exactly [AGENTS.local.md AGENTS.md CLAUDE.local.md CLAUDE.md]`;
curator — `Tier 3 Path = "CLAUDE.local.md", want AGENTS.local.md` (both tests); contract — `derived_test.go:265:
FrozenInstructionFiles = [CLAUDE.md CLAUDE.local.md]`.

Inverted in this milestone: `TestCodexLocalInstructions_DualFileMatrix` and
`TestCodexLocalInstructions_LargeBodySlicesAndFreshRead` (read order). **Not inverted:** the
`codex_contract_link_test.go` `@AGENTS.local.md` assertion (see Deviation 1).

**Cascade (outside the three named files, forced by a pin test).** `TestFrozenInstructionFilesPinnedInContract`
(`internal/hook`) requires `internal/contract.FrozenInstructionFiles` to equal the hook set, so the
contract copy gained the same two basenames, with its expectations in `internal/contract/derived_test.go`
and `internal/cli/contract_ac_test.go`. Observable effect: a signed contract's `frozen_files` now also
carries `**/AGENTS.md` and `**/AGENTS.local.md`, and the escalation detector's frozen-file class covers
them. This changes the literal set SPEC-AUTONOMY-CONTRACT-001 `AC-CONTRACT-018` states; that SPEC's
acceptance.md was not edited. The guard-set change and this cascade share one commit (they are
pinned to each other), separate from the launcher/curator commit, so the pair reverts as a unit.

**Deviation 1 — the link-test inversion is deferred to M3.** plan.md M2 says to invert the
`@AGENTS.local.md`-imports assertion here. Its `AGENTS.md` half (`= 0`) already holds and is unchanged;
its `CLAUDE.md` half (`= 1`) needs `codexCreatedClaudeBody` to emit `@AGENTS.local.md`, which plan.md
M3 owns, and acceptance.md `AC-IFU-012` (v0.3.2) assigns `TestCodexContractLink_LocalImportMatrix` to
M3. Inverting it at M2 would turn the tree red, the outcome the plan instruction exists to avoid.
`TestCodexLocalSeparation` ("no shared instruction file may point at AGENTS.local.md", `CLAUDE.md` entry)
will need the same inversion at M3.

Wider runs: `go test -count=1 -cover ./internal/harness/... ./internal/contract/... ./internal/escalation/...`
→ all `ok` except `internal/harness/rosterguard`; `go test ./internal/hook/ -run 'Frozen|Harness|Escalation|Contract' -v`
→ 24 `--- PASS`, `ok`; `go test ./internal/cli/ -run 'Codex|Contract' -v` → 547 `--- PASS`, one `--- FAIL:
TestCodexAuditMCPTool`. Both failures are pre-existing and not caused by M2: they assert text anchors in
files M2 does not touch (`CLAUDE.md` "consists of exactly **N retained agents**" → 0 matches;
`moai-mcp-tools.md` names `codex_role_audit*` → 0 matches), `git diff --quiet HEAD --` on those files
exits 0, and `go list -test -deps ./internal/harness/rosterguard/` includes none of the changed packages.

Build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (both 0 bytes output).
Lint: `golangci-lint v2.1.6 run ./internal/...` → `0 issues.` exit 0.
Coverage: `internal/harness/curator` 91.4%, `internal/contract` 96.3%, `internal/harness` 87.1%,
`internal/escalation` 88.7%; function level `codexLocalDeveloperInstructionArgs` 100.0%,
`checkHarnessFrozenZone` 58.3% (the unexercised part is the pre-existing prefix-zone branch).
Package-level `internal/cli` / `internal/hook` coverage not measured: slot `go-test-cli-hook` was held by
another session (`moai slot status`, ends 11:24Z) at load 26, so full suites were not run.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
