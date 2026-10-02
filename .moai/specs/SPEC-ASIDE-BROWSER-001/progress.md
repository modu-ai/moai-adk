# progress — SPEC-ASIDE-BROWSER-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-02 (iteration 2 revision; the iteration 1 signal carried 2026-10-02T11:52:52Z from the `spec_audit` tool, and this revision's validator run is recorded below)
- plan-audit: iteration 1 returned FAIL 0.75 (report `.moai/reports/t1439/plan-audit-iter1.md`, defects D1-D13); iteration 2 audit not yet run. This signal states that the revised artifacts pass the plan-phase validators below, not that an independent audit passed.
- tier: M · artifacts: spec.md (0.2.0), plan.md, acceptance.md, progress.md, plus spec-compact.md, research.md, decision-index.md (the last two beyond the Tier M set, by card request and by `interview.decision_gate: on`)
- requirements: 15 (REQ-ASB-001..015, budget 16) · acceptance criteria: 14 (AC-ASB-001..014, budget 16) · files: 15 (Tier M ceiling)
- measurement tree: `4bf547bcad7c155b1e91485921569db709ec3ac2` (worktree t1439, base develop); `git rev-parse HEAD` re-read at the start of the revision printed the same SHA
- SPEC ID check (iteration 1, Bash, verbatim): `[[ "SPEC-ASIDE-BROWSER-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`

### Iteration 2 revision — what changed

- Operator verdicts recorded 2026-10-02 (decision-index Q1, Q2, Q3; spec.md § 3 R-1 to R-3): orchestrator-only execution, core skill tier, completely silent fallback.
- REQ changes: REQ-006 split into REQ-006 (write confirmation) and REQ-007 (no subagent invokes Aside); the old REQ-012 (e2e-tester delegated prompt) inverted into REQ-007 and REQ-010 (orchestrator runs every Aside step, bounded output); the fallback-note requirement removed from REQ-012; REQ-003 and REQ-014 stripped of process text (moved to plan.md); REQ-013 hedged (no claim about which side writes the screenshot bytes). 14 to 15 REQs, 12 to 14 ACs.
- Removed: the e2e-tester Aside recipe block, its template/root edit, and the delegated-prompt subtests. Kept: one negative sentence in the e2e-tester definition (template and root) and the same boundary in the skill, pinned by `subagent_never_invokes_*` subtests with deletion mutants. The e2e-tester core catalog hash and the Codex TOML therefore still change, so those steps stay.
- Audit defects closed: D1 (hash set now `{moai-ref-aside-browser, moai, e2e-tester}`), D2 (t1434 dependency paragraph in plan § A.3, lines 21/27/32 of the sibling verdict re-read), D3 (core tier; measured slim fact recorded in Q2 and research), D4 (`description_within_listing_cap` subtest), D5 (`:108` and `:119` carve-outs with subtests and a restore mutant), D6 (operator verdicts applied), D7 (commit order chosen: RED commit before GREEN commit), D8 (REQ-013 hedged), D9 (exact docs commands and the four-to-five numeral edit), D10 (neutrality grep over every authored template file with a 40-hex pattern; `e2e.md` baseline has two pre-existing hits), D11 (closed two-literal allow-list for `full-access`, `never forget` mutant, concrete advise-the-operator phrase), D12 (REQ split, process text moved), D13 (README and skill-guide listed out of scope; the new test's duplication justified and its settings-token half named as the non-duplicate part).
- New open decision-index row: Q8 (which e2e phases stay with the e2e-tester). New assumption: A-8.
- validator evidence for this revision (tree HEAD `4bf547bca`): `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` → exit 0, last output line `✓ No findings — all SPEC documents are valid`; the orchestrator independently re-ran it after the revision and observed exit 0 with the same line (REQ rows = 15, AC headings = 14).

### Delta revision (N1, N2, N6)

Scope: the three findings of `.moai/reports/t1439/plan-audit-iter2.md` (PASS-WITH-DEBT 0.81) that the leader approved for one extra iteration; N3, N4, N5, N7 and the P3 items were not touched. spec.md is now version 0.2.1; counts are unchanged (REQ rows 15, AC headings 14, files 15).

- N1: the Q1 carve-out now reaches every `e2e.md` site by role. The sites are enumerated by command (E16: 15 lines at baseline, current numbers `:36, :54, :90, :108, :119, :197, :219, :277, :328, :331, :332, :334, :342, :344, :345`) and listed with their roles in plan.md M3.2; the carve-out is an appended `(except Aside: ...)` parenthetical so stripping it restores each original line. REQ-ASB-010 and REQ-ASB-012 and spec.md § 1 now say "every site ... that delegates script creation, execution, or recording to the e2e-tester or runs the missing-toolchain sequence" instead of an enumerated count. A new subtest `e2e_every_site_carved_out` (AC-ASB-009 and AC-ASB-010; no new AC) fails when any matching line lacks `except Aside` or `silently`, has a minimum-count guard of 15 lines, and has per-site strip mutants (including `:331`, `:332`, `:342`, `:119`). The probe-table row was dropped: the Aside probe is a sentence in the new Aside paragraph so it sits outside the e2e-tester probe introduction at `:90`. `e2e-tester.md:90` (a generic missing-toolchain sentence in the agent definition) is deliberately not edited; the negative sentence already removes Aside from that agent.
- N2: plan.md § C now splits pre-flight: `aside --version` and the M3.0 screenshot-persistence measurement are run by the orchestrator (main session), with the operator's approval obtained through its question channel, and handed to manager-develop as run-prompt input; manager-develop records them in §E.2 and returns a blocker report if they are absent. spec.md § 3 (R-1, A-6), acceptance.md (AC-ASB-011 gate, Gaps), decision-index Q5, research.md § 5 and spec-compact.md carry the same wording.
- N6: plan.md and acceptance.md AC-ASB-003 now say 13 selectors, matching the command; the baseline run of that exact command printed 13 `--- PASS` lines, `ok  	github.com/modu-ai/moai-adk/internal/template	0.708s`, exit 0 (E11b).
- validator evidence: `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` (binary built at plan time from the base tree; SPEC-only changes since) → exit 0, last output line `0 error(s), 0 warning(s)`. Tree HEAD at the run was `c9d4dd5d2` (the leader's commit of the SPEC artifacts on `4bf547bca`; `git diff --stat 4bf547bca HEAD` lists only files under `.moai/specs/SPEC-ASIDE-BROWSER-001/`). The run printed one `INFO OwnershipTransitionUnmeasured` row for spec.md (the commit carries no `Authored-By-Agent` trailer), which does not affect the exit status.

## §E.2 Run-phase Evidence

### M1 (guards and the documented-command test) — commit subject `test(SPEC-ASIDE-BROWSER-001): M1 RED-record`

All measurements below were taken by manager-develop (cycle_type=tdd) in this run on tree HEAD `59d69f1dd` (branch `WT-aside-browser-cli`, worktree t1439, base develop `4bf547bca`), unless a line names another measurer. The M1 commit's own SHA cannot appear here (a commit does not know its own hash); it is the tip of the branch after this edit lands.

#### Pre-flight (plan.md § C)

- `git rev-parse --short HEAD` printed `59d69f1dd`; `git branch --show-current` printed `WT-aside-browser-cli`.
- `go build ./...` exit 0 (no output).
- `aside --version`: measured by the orchestrator (main session), not by manager-develop (REQ-ASB-007: no subagent ever invokes `aside`). The orchestrator observed exactly `1.26.916.1741` (exit 0), supplied in the run prompt and recorded here verbatim on its attribution.
- Baseline mirror facts: `cmp internal/template/templates/.claude/skills/moai/workflows/e2e.md .claude/skills/moai/workflows/e2e.md` exit 0 (identical). `diff` of the e2e-tester template against root printed hunks `2d1` and `144c143` (changed-line count 3: one line only in the template, one line replaced by one line), exit 1. `cmp` of the `settings-management.md` template/root pair printed `differ: char 18093, line 167`, exit 1. These three are the pre-existing baselines of plan.md § B; none is touched by M1.
- Provenance of tool measurements (verification-claim-integrity §2.2): the Go tests and `go vet` compile from the tree under measurement. `moai spec lint` was run with a binary built from this tree's HEAD by `go build -o <scratch>/moai-t1439 ./cmd/moai` (exit 0); it carries no embedded commit stamp (built without the Makefile `LDFLAGS`; its `version` row printed `v3.1.3   none   built unknown`), so its judging-build commit is stated as "built from tree HEAD `59d69f1dd`, uncommitted M1 working changes only" and not as a commit stamp. The installed `moai` on PATH (`v3.2.0-rc.26`) was not used for the cited measurements.

#### M1.1 `TestMCPDefaultExcludesAside` (file `internal/template/mcp_template_neutrality_test.go`)

Guard passes at birth by design (it asserts invariants that already hold): baseline `grep -c -i aside` over the template `.mcp.json` and `settings.json.tmpl` printed `0` for both before the test was written.

Green run, `go test ./internal/template/ -run '^(TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1`, exit 0:

```
--- PASS: TestMCPNeutralityTemplateShape (0.00s)
--- PASS: TestMCPDefaultExcludesAside (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.172s
```

Mutant (a) — an `aside` key (`"command": "aside"`, `"args": ["mcp"]`) added to the template `.mcp.json`; same command, exit 1, verbatim:

```
--- FAIL: TestMCPNeutralityTemplateShape (0.00s)
--- FAIL: TestMCPDefaultExcludesAside (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.375s
```

with the new test's own message `template .mcp.json registers the optional Aside server "aside" in the default; Aside is activated only by an explicit operator command, never by the distributed default` and the sibling's `template .mcp.json carries non-default-on entry "aside" (only moai and context7 are permitted in the distributed default)`. Both failures were observed, not inferred.

Mutant (b) — the token `?aside` appended to the `$schema` value on line 2 of `settings.json.tmpl`; same command, exit 1, verbatim:

```
--- PASS: TestMCPNeutralityTemplateShape (0.00s)
--- FAIL: TestMCPDefaultExcludesAside (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.401s
```

with the message `settings.json.tmpl mentions Aside; the distributed settings must not enable, permit, or wire the optional Aside server`. Only the new test failed: the settings half is the non-duplicate part.

Both mutants were reverted by reversing the edit; the revert is proven by plain `git status --short`, which then listed only the two intended test files (` M internal/cli/mcp_test.go`, ` M internal/template/mcp_template_neutrality_test.go`), and `git diff --stat -- internal/template/templates` printed nothing.

#### M1.2 `TestMCP_Add_AsideDocumentedCommandLine` (file `internal/cli/mcp_test.go`)

The test drives the cobra command from `newMCPCmd()` with `add aside --command aside --args mcp` (plus `--scope user` for the user subtest), twice against a seeded config holding one unrelated server and one unrelated top-level key. Project scope runs under `t.Chdir(t.TempDir())` (the project path resolves from the working directory); user scope runs through the `userHomeDirFn` seam against a `t.TempDir()` home. No environment override; `moai mcp` Go code is untouched. It asserts exactly one `aside` entry, `command` `aside`, `args` `["mcp"]`, unrelated server and key preserved, and a byte-identical file after the second run. It is deliberately not parallel (it swaps a package global and the working directory).

Green run, `go test ./internal/cli/ -run '^TestMCP_Add_AsideDocumentedCommandLine$' -v -count=1`, exit 0:

```
--- PASS: TestMCP_Add_AsideDocumentedCommandLine (0.00s)
    --- PASS: TestMCP_Add_AsideDocumentedCommandLine/project (0.00s)
    --- PASS: TestMCP_Add_AsideDocumentedCommandLine/user (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.733s
```

Mutant — `--args` misspelled as `--arg` in the test invocation; same command, exit 1, verbatim:

```
    mcp_test.go:223: documented command line failed: unknown flag: --arg
--- FAIL: TestMCP_Add_AsideDocumentedCommandLine (0.00s)
    --- FAIL: TestMCP_Add_AsideDocumentedCommandLine/project (0.00s)
    --- FAIL: TestMCP_Add_AsideDocumentedCommandLine/user (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.830s
```

The mutant was reverted by reversing the edit; the final green run above ran after the revert.

#### M1 static checks (tree HEAD `59d69f1dd` plus the uncommitted M1 changes)

- `gofmt -l internal/cli/mcp_test.go internal/template/mcp_template_neutrality_test.go` printed nothing, exit 0.
- `go vet ./internal/template/ ./internal/cli/` printed nothing, exit 0.
- `golangci-lint run --timeout=5m --new-from-rev=2bba87fcd ./internal/template/ ./internal/cli/` printed `0 issues.`, exit 0 (linter `v2.1.6`, built with go1.26.8; `--new-from-rev` bounds the report to changes since the plan base).
- `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` (tree-built binary, see provenance above) printed `0 error(s), 0 warning(s)`, exit 0, with the one `INFO OwnershipTransitionUnmeasured` row for the plan commit.

#### Gaps (M1)

- Not run: the full `internal/template` and `internal/cli` package suites (minutes-long; the repository-wide verdict is owned by the CI run on the project's integration branch and is PENDING at report time). Only the named selectors were run, each with its `--- PASS:` or `--- FAIL:` line observed.
- Not run in M1 by design: AC-ASB-003 to AC-ASB-014 (they belong to M2 and M3; no skill, catalog entry, or workflow edit exists yet).
- The ordering claim "the failing run was observed before the commit" rests on this record plus the single `M1 RED-record` commit: the guards pass at birth, so the observed-failure evidence is the mutant runs above, which exist only as this progress record and were not committed separately. The commit graph can witness that M1 carries no implementation commit; it cannot witness the order of the mutant runs inside the commit.
- The mutant runs used the Go test cache bypass (`-count=1`) and file edits reverted by hand; no mutation tool was used.

#### Residual risk (M1)

- The settings half of `TestMCPDefaultExcludesAside` is a case-insensitive substring match for `aside`. A future legitimate use of the English word in a settings string (a comment-free JSON file makes that unlikely) would fail it and need a deliberate edit; this is the intended direction (fail loudly, then decide).
- `TestMCP_Add_AsideDocumentedCommandLine` shares the process working directory and a package global with other sequential tests; Go runs parallel tests only after sequential ones finish, so no overlap exists today, but a future test that parallelizes around it would need the same discipline.

### M2 (policy skill, core catalog entry, skill checker) — commit subjects `test(SPEC-ASIDE-BROWSER-001): M2 RED - ...` then `feat(SPEC-ASIDE-BROWSER-001): M2 GREEN - ...`

All measurements below were taken by manager-develop (cycle_type=tdd, run as a general-purpose agent because the manager-develop agent type auto-isolates into its own tree) in this run, branch `WT-aside-browser-cli`, worktree t1439, unless a line names another measurer. A commit cannot cite its own hash; commit SHAs of M2 are listed in the final report.

#### Pre-flight (plan.md § C)

- `git rev-parse --short HEAD` printed `1fa6e31ba` (the M1 commit); `git branch --show-current` printed `WT-aside-browser-cli`; `git status --short` printed nothing before any edit.
- `go build ./...` exit 0 (no output).
- `aside --version`: measured by the orchestrator, not by manager-develop (REQ-ASB-007); the orchestrator-supplied value is `1.26.916.1741`. The skill pins only the flags the orchestrator read from `aside --help`, `aside mcp --help`, `aside repl --help` and `aside skills --help`; no `aside` command was run by this agent.
- Baseline of the 13-selector command of AC-ASB-003, run on unmodified HEAD `1fa6e31ba` before any M2 edit, redirected to a file: exit 0, `grep -c '^--- PASS'` over that file printed `13`, last lines `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/template	0.250s`.

#### M2.1 (M2 RED) `internal/template/aside_skill_policy_test.go`

The test file holds the pure checker `checkAsideSkill(text)` over the literal anchors of acceptance.md § Checker anchors (the nine skill rows; the e2e-tester row and the `checkAsideE2E` rows are left to M3) and `TestAsideSkillPolicyAnchors`: one subtest per rule name, plus a `negative_controls` group that mutates the real skill text (each required literal removed in turn; bad lines added: unqualified `aside --permission full-access`, `never forget to run aside --permission full-access`, `npm i -g aside`, and the install command without the word `operator`; an oversize description; frontmatter removed). Every control must make the checker report its named rule; a control that does not is a failure.

Observed RED, run before the skill exists, `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, exit 1, verbatim decisive lines:

```
    aside_skill_policy_test.go:231: skill file unreadable: open <worktree>/internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md: no such file or directory
--- FAIL: TestAsideSkillPolicyAnchors (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/description_within_listing_cap (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/full_access_prohibited (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/guard_by_omission (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/repl_read_only_limit (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/write_confirmation_via_orchestrator (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/orchestrator_only_operation (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/subagent_never_invokes_skill (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/no_auto_install_advise_only (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/no_credentials_in_outputs (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/negative_controls (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.382s
```

(`<worktree>` abbreviates the absolute worktree path in this record.) Static checks on the RED tree: `gofmt -l internal/template/aside_skill_policy_test.go` printed nothing, exit 0; `go vet ./internal/template/` printed nothing, exit 0.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_

## §F Phase 4 Mode Selection and Kickoff record

### Mode Selection (orchestrator, 2026-10-02)

- Input parameters: tier M; scope about 15 files (template skill, catalog.yaml, e2e.md template+root, e2e-tester.md template+root, Codex TOML, Go guard tests, docs-site guides in 4 locales); domains 5 (skill, workflow doc, agent, catalog/Go tests, docs-site); file mix markdown + Go + TOML; concurrency benefit LOW (coding-heavy, ordered RED-then-GREEN commits per plan §F).
- Mode evaluation: direct not selected (non-trivial); serial selected (default for coding-heavy work, one writer per tree); fanout not selected (coding-heavy; at the 10-file threshold the tie-breaker defaults to the simpler mode); sweep not selected (not a single uniform mechanical transform).
- Decision: serial
- Justification: the work is coding-heavy and has a mandated commit order (guard test and recorded RED run before each implementation commit), so one manager-develop spawn per milestone M1, M2, M3 in order is the safe path. The orchestrator itself runs `aside --version` and the M3.0 screenshot-persistence measurement before the M3 spawn (operator verdict Q1: no subagent invokes aside).

### Kickoff record (plan→run, autonomous transition, auto-semantics §9.1)

decision record: decided_by=claude-code lane-10 orchestrator (factory lane, Kickoff autonomous transition; leader-approved delta iteration + leader instruction "PASS family and no P2+ blocker -> autonomous Kickoff") evidence_refs=.moai/reports/t1439/plan-audit-iter3-delta.md(verdict=PASS-WITH-DEBT score=0.86 open_P2_or_higher=NO audited_sha=2bba87fcd),.moai/reports/t1439/plan-audit-iter2.md(PASS-WITH-DEBT 0.81),.moai/reports/t1439/plan-audit-iter1.md(FAIL 0.75),.moai/specs/SPEC-ASIDE-BROWSER-001/progress.md#E.1,commit 2bba87fcd,sha256 spec=3ba0b82e plan=50ed9d06 acceptance=ee8b4971 research=2f134afe ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)

The verdict is PASS-WITH-DEBT where §9.1 names PASS; the entry rests on the factory leader's explicit instruction and the leader-extended Tier M audit ceiling. Debts carried to the run delegation: NF1, NF2, NF3, N3, N4, N5, N7 (details in `.moai/reports/t1439/kickoff-20261002.md`, local).
