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

### M3 — AGENTS.md body rewrite, CLAUDE.md thinning, contract link shape

Base: `dbec65279`. Raw outputs: `.moai/reports/t1243/m3/` (`red-*.txt` = RED, `ac-ifu-*.txt` = AC commands,
`pkg-*.txt` / `cli-subset-*.txt` = wider runs, `headless/` = fixture transcripts, `build-*.txt`, `lint.txt`).
The three `internal/cli` logs over 50 KB (`cli-subset-after-docs{,-2}.txt`, `green-cli-codex-contract.txt`) stay local;
their decisive lines are committed as `cli-runs-excerpt-{run1,run2,green}.txt`.
Test runs used `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ... -count=1 -v`.
Headless fixtures lived under the session scratchpad (`…/scratchpad/ifu/`), never this project; the tree was
`git status --short` clean after every launcher run. Binary: `make build` at `f89d4b650`, copied to the
scratchpad and called by path (`moai version` → `moai_cp/20260925_122548-728-gf89d4b650`).

**Start-of-milestone re-measure** (develop had moved 20 commits; none touched an M3 file —
`git diff --stat HEAD...develop -- AGENTS.md CLAUDE.md …AGENTS.md.tmpl …CLAUDE.md internal/cli/codex_contract.go …` empty):
root `AGENTS.md` 15,786 B / 8 `## `; `AGENTS.md.tmpl` 19,171 B / 12; `CLAUDE.md` 14,221 B both, byte-identical
(`cmp` exit 0) — equal to `.moai/reports/t1243/baseline-post-t1175.md`.

**plan.md §C.2 (SPEC A) resolved by measurement:** SPEC A landed first as `SPEC-CODEX-FACTORY-RETIRE-001`
(card t1242, `b3b60d82d` "M4 drop retired codex lane wording"). At M3 start `git show dbec65279:AGENTS.md | grep -c -- '-f\b\|Codex lanes'`
→ `0`, and `Codex lanes` → `0` in the template too. The `-f` drop from §8 and the §3 "Codex lanes" fix were
therefore already absorbed; M3 had nothing left to remove there.

#### AC matrix (M3-owned, plus the criteria no milestone named)

| AC | Status | Command | Decisive output |
|---|---|---|---|
| AC-IFU-002 | PASS | `grep -n '^@AGENTS\.md$\|^@AGENTS\.local\.md$' CLAUDE.md` / same on `internal/template/templates/CLAUDE.md` (two commands) | each: `9:@AGENTS.md` / `163:@AGENTS.local.md`, exit 0. Supplementary (not the criterion): `grep -n '^@' CLAUDE.md` → `9:@AGENTS.md`, `107:@.moai/config/sections/user.yaml`, `108:…language.yaml`, `163:@AGENTS.local.md` — the local import is the final import |
| AC-IFU-003 | PASS (declared proxy) | `grep -c '^@' AGENTS.md` / `… AGENTS.md.tmpl` | `0` / `0` |
| AC-IFU-004 | PASS | `find internal/template/templates -type f \( -name 'AGENTS.md' -o -name 'AGENTS.override.md' -o -name 'CLAUDE.local.md' \) -print`; `ls …/AGENTS.md.tmpl`; `go test ./internal/config/ -run '^TestCodexContractByteCeiling$' -v` | empty, exit 0; `…AGENTS.md.tmpl` exit 0; `--- PASS: TestCodexContractByteCeiling (0.00s)` |
| AC-IFU-005 | PASS | `wc -c < AGENTS.md` / `wc -c < internal/template/templates/AGENTS.md.tmpl` | `18106` / `19572` (≤ 24576); guard log `AGENTS.md = 18106 bytes (ceiling 24576, headroom 6470)`, `AGENTS.md.tmpl = 19572 bytes (… headroom 5004)` |
| AC-IFU-006 | PASS | `go test ./internal/config/ -run '^TestCodexNestedTemplateDiscoveryBudget$' -v` | `--- PASS: TestCodexNestedTemplateDiscoveryBudget (0.08s)`, no `no tests to run`. The guard reports no chain sums (it asserts the whole-repo total ≤ 24,576, stricter than 32,768), so the sum was measured independently: `find . -path ./.git -prune -o -path ./.claude/worktrees -prune -o -type f \( -name AGENTS.md -o -name AGENTS.override.md \) -print -exec wc -c {} \;` → `./AGENTS.md` / `18106` — one file, chain sum 18,106 ≤ 32,768 |
| AC-IFU-008 | PASS | `diff <(grep '^## ' AGENTS.md) <(grep '^## ' internal/template/templates/AGENTS.md.tmpl)` | empty, exit 0; 12 / 12 sections |
| AC-IFU-009 | PASS | clause 1 `grep -in 'personal.*~/\.codex/AGENTS\.md.*consumed\|narrowing what the project' AGENTS.md …AGENTS.md.tmpl`; clause 2 `grep -c 'truncat' AGENTS.md`; clause 3 `grep -c 'project instruction files only'` per mirror | no matches, exit 1; `1`; `1` / `1` (template `truncat` also `1`). Was RED at plan time (`0`/`0`, t1270) |
| AC-IFU-012 | PASS | `go test ./internal/cli/ -run '^TestCodexContractLinkCreation$\|^TestCodexContractLink_LocalImportMatrix$' -v` | `--- PASS: TestCodexContractLinkCreation (0.04s)` / `--- PASS: TestCodexContractLink_LocalImportMatrix (0.13s)` (8 sub-cases), `ok`; no `no tests to run` |
| AC-IFU-018 | PASS (see noun table: proxy) | `go test ./internal/template/... -run '^TestTemplateNeutralityC5_LocalFileNames$\|^TestTemplateNeutralityAudit$\|^TestNeutralityByInheritance$\|^TestTemplateNoInternalContentLeak$\|^TestAgentsDisclosureCompleteness$' -v`; strict tier `MOAI_TEMPLATE_LEAK_STRICT=1 … -run '^TestTemplateNoInternalContentLeak$'` | `--- PASS: TestNeutralityByInheritance (0.02s)`, `--- PASS: TestTemplateNeutralityAudit (0.00s)`, `--- PASS: TestTemplateNeutralityC5_LocalFileNames (0.00s)` (sub-cases `agents_local_allowed`, `claude_local_forbidden`), strict `--- PASS: TestTemplateNoInternalContentLeak (0.82s)` |
| AC-IFU-019 | PASS | (a) durable: `go test ./internal/cli/ -run '^TestClaudeImportResolution_AgentsLocalSentinel$' -v`; (b) headless on a `moai init` fixture: `claude -p "…(1) CONTRACT_SENTINEL, (2) LOCAL_SENTINEL…" --model claude-haiku-4-5-20251001` | (a) `sentinel IFU019-D2F92053B2E1 resolved; response: IFU019-D2F92053B2E1`, control arm `response: NOT_PRESENT`, `--- PASS: TestClaudeImportResolution_AgentsLocalSentinel (13.13s)`; (b) exit 0, stdout `CON-4H8MZ3` / `LOC-7K2QX9` (`LOC-…` is the `AGENTS.local.md` sentinel) |
| AC-IFU-020 (init leg) | PASS | fixture with no `AGENTS.local.md`: `moai init ac020 --non-interactive --llm claude`; `grep -n '^@AGENTS\.local\.md$' ac020/CLAUDE.md`; `claude -p "Run no tools. Reply with exactly: SESSION_OK" --model claude-haiku-4-5-20251001` | `moai init exit=0`; `ls ac020/AGENTS.local.md` → `No such file or directory`; `163:@AGENTS.local.md` (exit 0); `claude exit=0`, stdout `SESSION_OK` |
| AC-IFU-020 (update leg) | PASS | same fixture: `moai update --templates-only --yes < /dev/null`; re-grep | `moai update exit=0` (`Clean reinstall complete (4 files preserved, 4 deprecated removed)`); still no `AGENTS.local.md`; `163:@AGENTS.local.md` (exit 0) |
| AC-IFU-025 (local read) | PASS locally; CI clause open | `go test ./internal/config/ -run '^TestAlwaysLoadedTokenBudget$' -v` | `always-loaded surface = 66310 tokens (budget 77600, headroom 11290, 16 entries)` / `--- PASS: TestAlwaysLoadedTokenBudget (0.01s)` (baseline 65,591 → +719) |
| AC-IFU-026 | PASS | fixture: `moai init ac026 …`; `shasum -a 256 ac026/AGENTS.md`; overwrite with `SENTINEL_OVERWRITE_026`; `moai update --templates-only --yes`; re-hash; `grep -c` | `init exit=0`; `af31cc40…321562` (= `shasum` of `internal/template/templates/AGENTS.md.tmpl`); `update exit=0`; `af31cc40…321562`; `0` |
| AC-IFU-027 | PASS (6/6) | same `moai init` fixture (`CLAUDE.md` unmodified; `AGENTS.md` prefixed `CONTRACT_SENTINEL = CON-4H8MZ3`; `AGENTS.local.md` = `LOCAL_SENTINEL = LOC-7K2QX9`); one prompt per launcher asking both values | bare `claude -p … --model haiku` → exit 0, `CON-4H8MZ3` / `LOC-7K2QX9`; `moai cc --print … --model haiku` → exit 0, `CON-4H8MZ3` / `LOC-7K2QX9` (stderr `Launching Claude Code...`); `moai glm --print …` → exit 0, `CONTRACT_SENTINEL = CON-4H8MZ3` / `LOCAL_SENTINEL = LOC-7K2QX9` (stderr `Launching Claude Code with GLM backend...`, model `glm-5.3-flash[1m]`). Launcher sources: `git diff --stat c501bd1da..HEAD -- 'internal/cli/cc*.go' 'internal/cli/glm*.go'` → empty, exit 0; positive control `git diff --name-only c501bd1da..HEAD \| wc -l` → `44` (`c501bd1da` = `git merge-base develop HEAD`) |

Attribution control for AC-IFU-027 (not a criterion clause): a copy of the fixture with only the `@AGENTS.md`
line removed from `CLAUDE.md`, bare `claude -p` → `(1) NOT_PRESENT` / `(2) LOC-7K2QX9` — the contract sentinel
arrives through the import alone. `AGENTS.local.md` has the matching control inside the durable test.

Regression reads of M2 criteria after M3: AC-IFU-010 `--- PASS: TestCodexLocalInstructions_AgentsLocalReadFirst (0.00s)` /
`--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.00s)`; AC-IFU-016 `--- PASS: TestFrozenInstructionFiles (0.00s)`;
AC-IFU-017 `--- PASS: TestSurfaceForTier_Tier3 (0.00s)` / `--- PASS: TestPrepareTierDispatch_Tier3 (0.00s)`.
AC-IFU-001: `ls internal/template/templates/AGENTS.local.md` → `No such file or directory`; `grep -c '^/AGENTS\.local\.md$'` → `1` (`.gitignore`), `1` (template `.gitignore`).

#### RED (before GREEN)

- Contract (`red-contract.txt`, same four-test run before `codex_contract.go` changed): `--- FAIL: TestCodexContractLinkCreation`,
  `--- FAIL: TestCodexContractIdempotent`, `--- FAIL: TestCodexLocalSeparation`, `--- FAIL: TestCodexContractLink_LocalImportMatrix`;
  e.g. `snapshot 1: executing @AGENTS.local.md imports in CLAUDE.md = 0, want 1`, `entry CLAUDE.md: directives pointing at AGENTS.local.md = 0, want 1`,
  `executing directive order = [AGENTS.md], want AGENTS.md before AGENTS.local.md and AGENTS.local.md last`.
- Live import test: `git grep -c 'func TestClaudeImportResolution_AgentsLocalSentinel' HEAD -- '*_test.go'` at `c56407fcd` → no output, exit 1 (0 declarations).
  Its control arm is the in-test RED: without the import the response is `NOT_PRESENT`, which the resolved-arm assertion rejects.
- C5 test (`red-c5-mutant.txt`, pattern temporarily widened to `[A-Z]+\.local\.md`, then reverted):
  `C5 hit on "# CLAUDE.md\n\n@AGENTS.md\n\n@AGENTS.local.md\n" = true, want false` / `--- FAIL: TestTemplateNeutralityC5_LocalFileNames`.
- Signed-contract compatibility test: characterization of existing behaviour, GREEN on first run (no code change was expected); its
  non-vacuity is the exact-delta assertion below.

#### Named debt item 1 — the §3 worktree-entry paragraph, before and after (both mirrors identical)

Before (`dbec65279`, `AGENTS.md` and `AGENTS.md.tmpl`):

> **Work inside a worktree, entered through the launcher** (`moai cc -w <name>`,
> `moai cc -w <name> --spawn` for a new window, `EnterWorktree(<path>)` to re-enter); never create one
> with a bare `git worktree add`. Leave with `ExitWorktree`. Drive a worktree with `git -C <path>`,
> not `cd`.

After (`742c994f5`, both mirrors):

> **Work inside a worktree, entered through your harness's launcher** — Claude Code:
> `moai cc -w <name>` (`--spawn` for a new window), `EnterWorktree(<path>)` to re-enter,
> `ExitWorktree` to leave; Codex: `moai codex -w <worktree>`, which enters an existing tree and never
> creates one. Create a tree with `moai worktree new <name>`, never with a bare `git worktree add`.
> Drive a worktree with `git -C <path>`, not `cd`.

The Claude launcher is kept and the Codex one named beside it; `moai worktree new` (verified present:
`internal/cli/worktree/new.go` `Use: "new <name>"`) replaces the implicit "the launcher creates it". This is a
recorded paragraph, not a criterion's output — AC-IFU-003's `0`/`0` says nothing about clause wording.

#### Decisions taken inside M3 (reviewer attention)

1. **Q3 — the four template-only sections are folded into root, verbatim.** Dropping them from the template was
   not available: `TestAgentsDisclosureCompleteness` requires the shipped `AGENTS.md` to carry `Hook Event Coverage`.
   Cost measured: root +2,320 B (18,106 ≤ 24,576), always-loaded +719 tokens (66,310 ≤ 77,600).
2. **Codex contract "two-import shape" (`codex_contract.go`).** `codexLinkAgentsDirective` keeps its value
   `@AGENTS.md` (it is also the import counter's key); a new `codexLinkLocalDirective = "@AGENTS.local.md"` joins it.
   `codexCreatedClaudeBody` emits both; `codexCreatedAgentsBody` is unchanged (`# AGENTS.md\n` — the neutral contract
   imports nothing, AC-IFU-003). An existing `CLAUDE.md` gets only its missing link line(s) appended, so a file
   carrying the pre-unification `@AGENTS.md` alone gains `@AGENTS.local.md`. Append-only cannot reorder: a file
   carrying only `@AGENTS.local.md` gets `@AGENTS.md` after it (fixture `i9_local_link_only` pins that honestly).
   The `i5` "already linked → untouched" fixtures now carry both links, keeping their meaning.
3. **`TestCodexLocalSeparation` inverted** in the same commit as the constant (`e3476aecd`): the local file is reachable
   from `CLAUDE.md` through exactly one directive and never from `AGENTS.md`.
4. **Template §8 names no local filename.** `TestCodexLocalInstructions_TemplateDescribesCommonAndSpecificInputs`
   (SPEC-CODEX-LOCALMD-001's guard) failed on the first draft that named `AGENTS.local.md`
   (`template must describe common and Codex-specific inputs without enumerating local filenames`); the template §8 was
   reworded to satisfy it rather than editing another SPEC's test. Root §8 names both files and the new order.
5. **`@AGENTS.local.md` sits in a new `## 18. Local Instructions (imported)`, before `## MOAI:LEARNED-WORKFLOW`.**
   `internal/merge/strategies.go` `managedSectionHeadings` preserves that section from the local side, so an import
   placed inside it would not propagate on `moai update`.
6. **The live test skips on no binary, no credentials, or `MOAI_SKIP_LIVE_CLAUDE`** (measured: both skip paths print
   `--- SKIP: TestClaudeImportResolution_AgentsLocalSentinel`). With a working `claude` it runs by default in
   `go test ./internal/cli/` — two haiku calls, about 13 s.

#### M2 follow-up: signed-contract backward compatibility

`go test ./internal/contract/ -run '^TestVerify_SignedUnderPriorFrozenSetStillVerifies$' -v` (committed, `f89d4b650`):
signs a fixture contract while `FrozenInstructionFiles` holds the prior two entries, verifies the same bytes under the grown set.

```
prior-set frozen_files = [**/CLAUDE.local.md **/CLAUDE.md .claude/rules/moai/core/moai-constitution.md internal/x/**]
grown-set frozen_files = [**/AGENTS.local.md **/AGENTS.md **/CLAUDE.local.md **/CLAUDE.md .claude/rules/moai/core/moai-constitution.md internal/x/**]
--- PASS: TestVerify_SignedUnderPriorFrozenSetStillVerifies (0.00s)
```

Both verifications are `signed-valid` with no reasons and the same `contract_sha256` — **verification does not break**;
the set feeds only the derived `frozen_files` report, never the signed digest, and `internal/contract/sign/sign.go` reads
no frozen-instruction value.

Escalation side (one-off probe, not committed; source was a temporary `internal/escalation` test, output
`m2-followup-escalation-probe.txt`): a card **armed while the prior set was in force keeps its arming snapshot** —

```
armed snapshot frozen_files (prior set) = [**/CLAUDE.local.md **/CLAUDE.md internal/x/**]
after set growth, write AGENTS.md: invariant-violation records = 0
armed snapshot frozen_files after the write = [**/CLAUDE.local.md **/CLAUDE.md internal/x/**]
control, write CLAUDE.md: invariant-violation records = 1
```

So an already-armed card does not start tripping class 2b on `AGENTS.md` / `AGENTS.local.md` until it re-arms
(`classFrozenFile` reads `r.st.Armed.FrozenFiles`, "cached at arming"). Not a break; a lag bounded by the arming.

#### Builds, lint, coverage

- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=linux GOARCH=amd64 go build ./...` exit 0 (0 bytes output each).
- `golangci-lint v2.1.6 run ./...` → `0 issues.` exit 0.
- `make build` (three times; last at `f89d4b650`) exit 0; `catalog.yaml updated successfully (13408 bytes)` with no tracked change.
- `moai constitution validate` (built binary, by path) → `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`.
- Coverage: `internal/contract` 96.3%; `internal/template` 82.0% (package-level, test-only change there);
  `codex_contract.go` function level from `-run 'CodexContract|CodexLocal|CodexInit|CodexCount'`: `secureCodexInstructionContract` 80.0%,
  `codexAppendLine` 75.0%, `codexCountExecutingImports` 100.0%. Package-level `internal/cli` not measured (slot `go-test-cli-hook` held by
  another session, `moai slot status` → `held … ends 2026-09-27T14:26:09Z`).
- Wider runs: `go test ./internal/constitution/... ./internal/harness/rosterguard/ ./internal/config/... ./internal/contract/...` → all `ok`;
  `go test ./internal/template/... ./internal/merge/... ./internal/core/project/...` → all `ok`;
  `go test ./internal/cli/ -run 'Codex|Contract|MCP|Claude' -v` → 1,763 `--- PASS`, `ok` (M2's `TestCodexAuditMCPTool` failure no longer reproduces);
  `go test ./internal/cli/ -run 'Init|Update|Doctor|Template|Agents|Neutral'` → `ok`; hook `-run '^TestFrozenInstructionFiles$'` `ok`.

#### Close pass — requirement × criterion noun comparison

Re-derived from each criterion's own citation line (acceptance.md §D), not from the §D.2 table. Declared pair total:
**19**. `AC-IFU-025` cites `(all)` rather than an id, so it forms no pair. Every one of the 15 requirement ids appears below.

| # | REQ | AC | Noun the requirement constrains | Noun the criterion measures | Verdict |
|---|---|---|---|---|---|
| 1 | REQ-IFU-001 | AC-IFU-026 | `AGENTS.md` deployed as a harness-neutral contract, overwritten by update, not user-edited | `AGENTS.md` exists after init; update restores its hash and drops a sentinel | proxy — "the file is deployed and is replaced rather than preserved, which is what makes it a non-user-edited surface". The harness-neutral half is not measured here (debt item 1) |
| 2 | REQ-IFU-002 | AC-IFU-002 | `CLAUDE.md` carrying, in order, `@AGENTS.md`, mechanism layer, `@AGENTS.local.md` as final import | the two import lines, exactly two matches, `@AGENTS.md` line number below `@AGENTS.local.md` | proxy — the criterion states no reason; "final import" and "mechanism layer between" are unmeasured by it (supplementary `grep -n '^@'` above shows the local import is the last `@` line) |
| 3 | REQ-IFU-002 | AC-IFU-012 | same | executing-import counts in the codex contract's output: `AGENTS.local.md` in `CLAUDE.md` = 1, in `AGENTS.md` = 0 | proxy — "Dropping the `AGENTS.md` direction would permit the neutral contract to pull a local file into Codex's discovered chain; dropping the `CLAUDE.md` direction would leave the import unreached." Measures the codex contract producer, not the `moai init` template |
| 4 | REQ-IFU-003 | AC-IFU-001 | `AGENTS.local.md` not deployed, absent from templates, listed in deployed `.gitignore` | `ls` of the template path fails; `^/AGENTS\.local\.md$` count in both `.gitignore` files | match |
| 5 | REQ-IFU-004 | AC-IFU-020 | init and update succeed with no `AGENTS.local.md` and `CLAUDE.md` keeps the unresolved import line | init / update exit codes and the literal line after each | match |
| 6 | REQ-IFU-005 | AC-IFU-027 | Claude reaches `AGENTS.md` and `AGENTS.local.md` through `CLAUDE.md` imports alone, no launcher change in `moai cc` / `claude` / `moai glm` | two sentinels per launcher, plus launcher-source diff | match |
| 7 | REQ-IFU-006 | AC-IFU-010 | Codex launcher reads `AGENTS.local.md` ahead of `CLAUDE.local.md` | ALPHA before BETA in assembled `developer_instructions` and preamble order | match |
| 8 | REQ-IFU-013 | AC-IFU-016 | frozen-instruction set includes the four basenames | the `frozenInstructionFiles` line and one sub-case per entry | match |
| 9 | REQ-IFU-014 | AC-IFU-017 | curator Tier-3 target is `AGENTS.local.md` | two dispatch tests and the Tier-3 `Path` | match |
| 10 | REQ-IFU-015 | AC-IFU-018 | neutrality scan: `AGENTS.local.md` allowed, `CLAUDE.local.md` forbidden | `TestNeutralityByInheritance` (emitter-introduced SPEC-id/date/hash tokens) plus the neutrality workflow over fixtures | proxy — the criterion states no reason, and its named test measures a different noun (it never reads C5). The fixture half is what matches; the CI workflow runs only over the live tree, so M3 supplied the fixture measurement as `TestTemplateNeutralityC5_LocalFileNames` |
| 11 | REQ-IFU-017 | AC-IFU-009 | preamble states the budget is charged against project instruction files only, and keeps the truncation statement | old sentence absent (both mirrors), `truncat` ≥ 1 (root), replacement phrase ≥ 1 per mirror | match |
| 12 | REQ-IFU-018 | AC-IFU-008 | same `## ` section set across the two mirrors | diff of `## ` lines | match |
| 13 | REQ-IFU-018 | AC-IFU-003 | no clause in §1-§7 restricted to one harness | count of `^@` lines | proxy — "This is a declared proxy on the neutrality clause, not its assertion." Before/after paragraph recorded above (debt item 1) |
| 14 | REQ-IFU-019 | AC-IFU-005 | each deployed contract document ≤ 24,576 B | `wc -c` per mirror | match |
| 15 | REQ-IFU-023 | AC-IFU-019 | a mechanical check that the `@AGENTS.local.md` import resolved | sentinel in a headless response; durable test `--- PASS:` | match (CI records the test skipped — debt item 3) |
| 16 | REQ-IFU-023 | AC-IFU-021 | same | real-worktree ancestor discovery of the local files | proxy — "This criterion passes on a recorded result, not on a particular result — it exists to settle the question" (cited as "REQ-IFU-023; gates the worktree leg") |
| 17 | REQ-IFU-024 | AC-IFU-004 | template mirror keeps a non-discovered filename; no `AGENTS.md` under templates | recursive find of the discovered filename set; `.tmpl` exists; per-file guard | match |
| 18 | REQ-IFU-025 | AC-IFU-006 | sum of discovered instruction files in any nested chain ≤ 32,768 | nested-discovery guard `--- PASS:` (asserts whole-repo total ≤ 24,576) | match — the guard is stricter than the requirement; the criterion's "every reported chain sum" clause has no reported sums, so the sum was measured separately (18,106) |
| 19 | REQ-IFU-025 | AC-IFU-022 | same | Codex prompt-input render: whether `AGENTS.local.md` enters the discovered chain, and tail truncation | proxy — "A completed run proves nothing here: the documented failure is a silent tail cut with exit `0` and empty stderr, so the tail sentinels are the only positive indicator." |

Rows: 19 = declared total. Mismatches needing a SPEC change: none; proxies without a stated reason: rows 2 and 10
(reported below as findings, not repaired — acceptance.md bodies are out of this agent's scope).

§D.2 set-membership command, run at close from the SPEC directory:

```
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
10d9
< REQ-IFU-016
exit 1
```

**Not empty.** `acceptance.md` still names the retired `REQ-IFU-016` in its v0.3.2 repair notes (e.g. "The citation also
dropped `REQ-IFU-016`"), and `spec.md` no longer declares it. No criterion cites it; the diff is a prose token, the exact
phantom-row shape acceptance.md §D.5's own note warns about. Blocker for manager-spec (body edit outside this agent's scope).

`moai spec lint SPEC-INSTRUCTION-FILES-UNIFY-001` (built binary at `f89d4b650`) → `0 error(s), 5 warning(s)`, all
`VacuousTestAssertion` — so the t1269 rule the acceptance.md head block calls "not landed" is live in this tree. The five
hits are acceptance.md lines 23, 222, 223, 301 and spec.md line 186: the illustrative `TestFoo` in the anchoring rule
and the quoted pre-repair patterns in the v0.3.1 repair notes, not a live criterion's command. Recorded for manager-spec;
debt item 2's "enforced by review until t1269 lands" wording is now stale.

#### Gaps

- `AC-IFU-025`'s own clause (CI full suite on the PR head) is not observed — only the local guard read. CI owns it.
- Package-level `internal/cli` and `internal/hook` coverage and full suites not run (slot held by another session).
- The sibling `TestCodexLocalInstructions_*` launcher behaviour is unchanged by M3 and was only regression-read.
- `moai cc` / `moai glm` legs ran `--print` headless; interactive launch was not exercised.
- The escalation probe measured one fixture with the in-process set swap; a real binary upgrade across an armed card was not run.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-27T14:15:26Z
run_commit_sha: f89d4b650            # last implementation commit; the evidence commit follows it
run_status: complete-pending-ci      # AC-IFU-025's CI clause is the only open criterion
ac_pass_count: 19                    # of 20; AC-IFU-021 / AC-IFU-022 pass on a recorded result (M1)
ac_fail_count: 0
ac_open: [AC-IFU-025]                # local guard PASS; CI full suite on the PR head not yet observed
preserve_list_post_run_count: n/a    # tdd cycle, no DDD preserve list
l44_pre_commit_fetch: "fetched tip c7c3f3592 vs HEAD f89d4b650 -> rev-list left/right 38 12 (absorbed at the integration window, not here)"
l44_post_push_fetch: n/a             # lane does not push
new_warnings_or_lints_introduced: 0  # golangci-lint v2.1.6 run ./... -> 0 issues
cross_platform_build:
  darwin: pass                       # go build ./... exit 0
  linux_amd64: pass
  windows_amd64: pass
total_run_phase_files: 44            # git diff --name-only c501bd1da..f89d4b650 | wc -l (M1-M3, incl. evidence)
m1_to_mN_commit_strategy: "per logical unit: M3 = e3476aecd contract+link tests, 742c994f5 AGENTS.md mirrors, c56407fcd CLAUDE.md mirrors, 8066cc69d live import test, 22229c12d C5 test, f89d4b650 signed-contract compat test"
blockers_for_manager_spec:
  - "acceptance.md §D.2 set-diff prints '< REQ-IFU-016' (retired id still named in v0.3.2 repair prose)"
  - "AC-IFU-002 and AC-IFU-018 are proxies with no stated reason (noun table rows 2 and 10)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
