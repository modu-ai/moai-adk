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

(`<worktree>` abbreviates the absolute worktree path in this record.) Static checks on the RED tree: `gofmt -l internal/template/aside_skill_policy_test.go` printed nothing, exit 0; `go vet ./internal/template/` printed nothing, exit 0. The M2 RED commit (subject `test(SPEC-ASIDE-BROWSER-001): M2 RED - ...`) carries only the test file and this record, and precedes the GREEN commit in `git log`.

#### M2.2 (M2 GREEN) the skill, `internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md`

Authored per plan.md M2.2: frontmatter `name: moai-ref-aside-browser`, folded `description` and `when_to_use`, `user-invocable: false`, no `CLAUDE_SKILL_DIR` token; body sections: purpose, when to use and not, who operates it (orchestrator alone plus the subagent sentence), seven safe-use rules (no unrestricted permission, Guard by omission, repl read-only limit stated as discipline, write confirmation through the question channel, advise-only install, no secrets in outputs, bounded output), the optional registration command, and the relation to `/moai e2e`. Facts pinned are only those the orchestrator read from `aside --help`, `aside mcp --help`, `aside repl --help` and `aside skills --help` (see Pre-flight); the skill states that flag and tool names reflect one observed version and that a failed probe or call means Aside is treated as absent. The skill never spells the unrestricted permission value outside the two allow-listed literals. The skill is not added to `delegation.yaml` `domain_skills` (`grep -n 'aside' .moai/config/sections/delegation.yaml` printed nothing, exit 1) and has no command wrapper.

Observed mutant against the real file (not only the in-test controls): the line `never forget to run aside --permission full-access` inserted before the last section, `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1` exit 1, decisive lines `aside_skill_policy_test.go:234: line 57 mentions full-access outside the allow-list: "never forget to run aside --permission full-access"` and `--- FAIL: TestAsideSkillPolicyAnchors/full_access_prohibited`; the line was then removed (`grep -c 'never forget'` over the skill printed `0`) and the green run below ran after the revert.

Green run, `go test ./internal/template/ -run '^(TestAsideSkillPolicyAnchors|TestSkillTreeHasNoClaudeSkillDirToken|TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1`, exit 0, verbatim decisive lines:

```
--- PASS: TestAsideSkillPolicyAnchors (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/description_within_listing_cap (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/full_access_prohibited (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/guard_by_omission (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/repl_read_only_limit (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/write_confirmation_via_orchestrator (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/orchestrator_only_operation (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/subagent_never_invokes_skill (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/no_auto_install_advise_only (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/no_credentials_in_outputs (0.00s)
    --- PASS: TestAsideSkillPolicyAnchors/negative_controls (0.00s)
--- PASS: TestMCPNeutralityTemplateShape (0.00s)
--- PASS: TestMCPDefaultExcludesAside (0.00s)
--- PASS: TestSkillTreeHasNoClaudeSkillDirToken (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.255s
```

The `negative_controls` group ran 18 control subtests (counted in a separate full run: `grep -c` over its `--- PASS: .../negative_controls/` lines printed `18`): 12 anchor removals (every required literal of every rule), 4 bad lines (unqualified `aside --permission full-access`, `never forget to run aside --permission full-access`, `npm i -g aside`, the install command without the word `operator`), an oversize description, and frontmatter removal.

The `subagent_never_invokes_tester` subtest and the `checkAsideE2E` rows are left to M3 by design (the e2e-tester definition and `e2e.md` do not carry Aside yet).

#### M2.3 catalog, build, mirror

- Catalog entry added by hand under `catalog.core.skills` (`tier: core`, `path: templates/.claude/skills/moai-ref-aside-browser/`, `version: 1.0.0`, placeholder hash of 64 zeros), placed alphabetically before `moai-ref-git-workflow`. `make build` (agents-emit-check, commands-emit-check, tool-policy-drift-check, templ-generate, gen-catalog-hashes --all, go build) exit 0; its log ends `catalog.yaml updated successfully (14151 bytes)` followed by the `go build` line. Immediately after it, plain `git status --short` listed ` M internal/template/catalog.yaml` and the two untracked new-skill directories only: `make build` regenerated no unrelated file.
- Catalog diff reading, `git diff -U4 -- internal/template/catalog.yaml`: exactly one hunk, `@@ -52,8 +52,13 @@`, adding the entry `moai-ref-aside-browser` with `hash: 653940540ae35059faa4dcdd7d4f2bf4d7b4960de023f6da405e9632c0030b5f`. No other `hash:` line changed (`moai` and `e2e-tester` are untouched in M2, as expected: no file under the core `moai` skill directory and no e2e-tester definition changed), and no `generated_at` hunk appeared. The changed hash set is therefore exactly `{moai-ref-aside-browser}`.
- Mirror: `mkdir -p .claude/skills/moai-ref-aside-browser`, then `cp` from the template to `.claude/skills/moai-ref-aside-browser/SKILL.md`; `cmp internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md .claude/skills/moai-ref-aside-browser/SKILL.md` printed nothing, exit 0.
- `.agents/skills` is a derived, gitignored mirror and was not hand-edited: `git check-ignore -v .agents/skills/moai-ref-aside-browser/SKILL.md` printed `.gitignore:154:.agents/skills/moai*` (exit 0), and the two mirror guards are part of the 13-selector run below.
- The 13-selector command of AC-ASB-003 after the change, redirected to a file, exit 0, `grep -c '^--- PASS'` printed `13`, and the verbatim PASS lines were `TestGitignore_IgnoresSkillMirrorOnly`, `TestPublishedSkillsNamesMatchTree`, `TestSkillMirror_SetIsDerivedNotConstant`, `TestEmbeddedMoaiSkillNames`, `TestCatalogReferencesValid`, `TestCatalogNoDuplicateEntries`, `TestSlimFS_HidesNonCoreEntries`, `TestSlimFS_PreservesCoreEntries`, `TestCatalogTierValid`, `TestAllSkillsInCatalog`, `TestWorkflowTriggerCoverage`, `TestCatalogHashCoversSkillSubfiles`, `TestManifestHashFormat`; last line `ok  	github.com/modu-ai/moai-adk/internal/template	0.412s`.
- AC-ASB-003 reader checks: `grep -n 'user-invocable: false' <skill>` printed `14:user-invocable: false`; `grep -n -A3 'name: moai-ref-aside-browser' internal/template/catalog.yaml` printed the `tier: core` entry with the 64-hex hash above; `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user' <skill>` printed `1` (AC-ASB-002 skill-side check); `grep -c 'CLAUDE_SKILL_DIR' <skill>` printed `0`.
- Neutrality grep of AC-ASB-013 over the skill, `grep -n -E 'SPEC-[A-Z]|t1439|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{40}\b|\b[0-9a-f]{7,8}\b' <skill>`: no output, exit 1.

#### M2 static checks (tree HEAD `5b932b06c` plus the uncommitted M2 GREEN changes)

- `go vet ./internal/template/` printed nothing, exit 0. `gofmt -l internal/template/aside_skill_policy_test.go` printed nothing, exit 0.
- `golangci-lint run --timeout=5m --new-from-rev=1fa6e31ba ./internal/template/` printed `0 issues.`, exit 0 (linter `v2.1.6`, built with go1.26.8).
- `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` with a binary built by `go build -o <scratch>/moai ./cmd/moai` from this tree (exit 0; no commit stamp, built without the Makefile `LDFLAGS`, so its judging-build coordinate is "tree HEAD `5b932b06c` plus uncommitted M2 GREEN changes"): exit 0, output `✓ No findings — all SPEC documents are valid`. The installed `moai` on PATH was not used.

#### Gaps (M2)

- Not run: the full `internal/template` and `internal/cli` package suites (minutes-long); the repository-wide verdict is owned by the CI run on the project's integration branch and is PENDING at report time. Only named selectors were run, each with its `--- PASS:` line observed. In particular the neutrality and leak audits (`TestTemplate*Neutrality*`, internal content leak) were not run as whole tests; the skill was checked by the AC-ASB-013 grep above and by `TestSkillTreeHasNoClaudeSkillDirToken` only.
- AC-ASB-004 to AC-ASB-007 are carried by `TestAsideSkillPolicyAnchors` on the template copy; the root mirror is covered by `cmp` only.
- AC-ASB-006 is only half observed in M2 (the skill sentence); the e2e-tester sentence and the Codex TOML belong to M3. AC-ASB-008 to AC-ASB-012 (e2e wiring), AC-ASB-014 (docs-site) and the e2e-tester catalog hash are M3.
- No `aside` command or Aside MCP tool was run by this agent (the skill text rests on the orchestrator-supplied observation).
- Behavior of the MCP `exec` tool's own permission argument is unmeasured; the skill therefore treats every `exec` run as state-changing and requires the same operator confirmation (a policy choice, not a measured limit).

#### Residual risk (M2)

- The checker is lexical: it proves the text says what the anchors require and instructs nothing the forbidden patterns name; it cannot prove what a model does at run time, and no hook enforces the orchestrator-only boundary.
- The closed allow-list for the unrestricted permission value is line-level, so a line that carries an allow-listed literal and an additional bad mention would pass; the checker's stated rule (acceptance.md § Checker anchors) is line-level and was kept literal.
- `aside repl` has no permission flag, so the read-only limit is discipline only.
- The core tier puts the skill's `description` plus `when_to_use` into every install's listing; the measured cap is the `description_within_listing_cap` subtest (a character count, not a token count).

### M3 (e2e wiring, e2e-tester sentence, Codex emit; docs-site is a later milestone) — commit subjects `test(SPEC-ASIDE-BROWSER-001): M3 RED - ...` then `feat(SPEC-ASIDE-BROWSER-001): M3 GREEN - ...`

All measurements below were taken by manager-develop (cycle_type=tdd, run as a general-purpose agent because the manager-develop agent type auto-isolates into its own tree) in this run, branch `WT-aside-browser-cli`, worktree t1439, unless a line names another measurer. A commit cannot cite its own hash; commit SHAs of M3 are listed in the final report. No `aside` command or Aside MCP tool was run by this agent.

#### Pre-flight (plan.md § C)

- `git rev-parse --short HEAD` printed `1c023f7ba` (M2 GREEN); `git branch --show-current` printed `WT-aside-browser-cli`; `git status --short` printed nothing before any edit; `pwd` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1439`.
- `go build ./...` exit 0 (no output).
- E16 site enumeration on the unedited template `e2e.md`: `grep -c -E '<role pattern>' internal/template/templates/.claude/skills/moai/workflows/e2e.md` printed `15` (exit 0); the same pattern with `-n` printed lines 36, 54, 90, 108, 119, 197, 219, 277, 328, 331, 332, 334, 342, 344, 345 (re-read before editing: unchanged from the plan's list).
- `cmp internal/template/templates/.claude/skills/moai/workflows/e2e.md .claude/skills/moai/workflows/e2e.md` printed nothing, exit 0.
- `diff internal/template/templates/.claude/agents/moai/e2e-tester.md .claude/agents/moai/e2e-tester.md` printed hunks `2d1` (`< isolation: worktree`) and `144c143` (the task-tracking line), exit 1: 3 changed lines, the baseline, preserved untouched.
- `grep -c 'except Aside' internal/template/templates/.claude/skills/moai/workflows/e2e.md` printed `0`, exit 1.
- Size budget: `wc -c` printed `25391` for the template `e2e.md` and `10516` for the template `e2e-tester.md` (budget 40,000 characters per instruction file).

#### M3.0 orchestrator-measured input (measurer: the orchestrator, not manager-develop)

The orchestrator ran one operator-approved measurement on a blank tab, on tree HEAD `1c023f7ba`, with the skill `moai-ref-aside-browser` loaded first, no `--permission` flag, aside `1.26.916.1741`. Reported verbatim by the orchestrator:

- `aside repl "const page = await openTab('about:blank'); const r = await page.screenshot({path: '<scratch>/shot.png'}); ..."` printed `✔︎ Opened a new tab and set it active: tabs[0], page → (about:blank)` and then `Error: page.screenshot: browser CDP command timed out while capturing viewport screenshot before the default screenshot timeout completed` after about 35 s (`[error | 35475ms]`). The exit status of the shell command was 0. No file was written at the target path (the directory listing showed only the log).
- A second `aside repl` call using `getTabs()` printed `ReferenceError: 'getTabs' is not defined` (the help text names it; this version does not define it).
- Orchestrator's conclusion, adopted here: screenshot persistence to a path is UNCONFIRMED (one capture failure with a CDP timeout, not a "path unsupported" error). Per AC-ASB-011 the evidence wording stays at the hedged form: a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path, without asserting which side writes the bytes. No SPEC text was amended.

#### M3.1 (M3 RED) `internal/template/aside_skill_policy_test.go`

The test file gains `checkAsideE2E(text)` (every e2e-side anchor of acceptance.md § Checker anchors), `checkAsideTester(text)` (the e2e-tester half of the subagent anchor, subtest `subagent_never_invokes_tester`), eleven e2e rule subtests named exactly as acceptance.md names them (`e2e_explicit_only`, `e2e_ci_excluded`, `e2e_orchestrator_executes_aside`, `e2e_execution_owner_carveout`, `e2e_aside_output_bounded`, `e2e_silent_fallback_no_aside_message`, `e2e_missing_toolchain_carveout`, `e2e_tool_bypass_line_carveout`, `e2e_every_site_carved_out`, `e2e_no_install_command`, `e2e_repl_screenshot_evidence`), and an `e2e_negative_controls` group of 36 control subtests. Audit debts owned by M3 and implemented here:

- NF1: the site rule accepts `silently` only on the exact precedence-sentence line (a line equal to the sentence, optionally as a list item); every other matching line must carry `except Aside`. Controls: `add/silently_site_line`, `swap/silently_for_carve_out`, `append/precedence_sentence_to_site_line`.
- N3: the silence rule also forbids `tell the operator`, `inform the operator`, `let the operator know` (and the `user` forms) on any line that mentions Aside, beside `warn`, `fallback note`, `notify`, `report that Aside`. Controls: seven `add/note_line_N` mutants.
- N4: the CI clause (a line carrying `CI=true`, `Aside`, and `unavailable`) must sit inside the Aside paragraph (the run of non-blank lines around the `Skill("moai-ref-aside-browser")` load) or on the `--tool` bypass line, not only in the no-flag branch. Controls: `remove/ci_line`, `move/ci_line_to_no_flag_branch` (moves the clause under the existing `environment detected` line).
- N5: the evidence-ledger commands E11/E15 are run without the table-escape backslashes (the enumeration regexp in the test and in this record has plain `|`).
- Per-site strip mutants: `strip/execution_owner`, `strip/bounded_output`, `strip/missing_toolchain_header`, `strip/bypass_sentence_line`, the named `strip/chain_phase2_line`, `strip/chain_phase3_line`, `strip/summary_step4_line`, and `strip/every_enumerated_site`, which strips the appended carve-out from every enumerated line in turn and requires the sites rule to name that line. The minimum-count guard (at least 15 matched lines) is `delete/sites_below_minimum`.
- The bounded-output carve-out must itself carry `e2e/.runs/` (the original line already does, so the check reads only the text after `except Aside`): `mutate/bounded_tail_without_runs_dir`.

Observed RED, run on the unedited `e2e.md` and `e2e-tester.md` (tree HEAD `1c023f7ba` plus the uncommitted test change), `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1` redirected to a file, exit 1. Verbatim decisive lines:

```
--- FAIL: TestAsideSkillPolicyAnchors (0.04s)
    --- FAIL: TestAsideSkillPolicyAnchors/subagent_never_invokes_tester (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_explicit_only (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_ci_excluded (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_orchestrator_executes_aside (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_execution_owner_carveout (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_aside_output_bounded (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_silent_fallback_no_aside_message (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_missing_toolchain_carveout (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_tool_bypass_line_carveout (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_every_site_carved_out (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_repl_screenshot_evidence (0.00s)
    --- FAIL: TestAsideSkillPolicyAnchors/e2e_negative_controls (0.04s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.416s
```

Failure texts on the unedited files, verbatim: `required literal "never invokes Aside" is missing`; `required literal "only when `--tool aside` is passed explicitly" is missing`; `no line carries "CI=true", Aside, and "unavailable"`; `required literal "the ORCHESTRATOR runs every Aside step" is missing`; `line 36 (the execution-owner statement) lacks "except Aside": ...`; `line 54 (the bounded-output rule) lacks "except Aside": ...`; `required literal "no Aside-specific message" is missing` and `the precedence sentence is not present as a line of its own`; `line 108 (the missing-toolchain sequence header) lacks "except Aside": "Missing-toolchain sequence (per selected toolchain):"`; `line 119 (the --tool bypass sentence) lacks "except Aside": ...`; and one `line N matches the site pattern without "except Aside"` per site (N = 36, 54, 90, 108, 119, 197, 219, 277, 328, 331, 332, 334, 342, 344, 345).

Observed state of the 36 `e2e_negative_controls` subtests on the unedited files: 22 FAIL and 14 PASS. FAIL (control cannot run because the real text does not yet carry the anchor or the carve-out, reported by the control as vacuous): `move/ci_line_to_no_flag_branch`, `mutate/bounded_tail_without_runs_dir`, `remove/ci_line`, `remove/e2e_explicit_only/0..3`, `remove/e2e_orchestrator_executes_aside/0..1`, `remove/e2e_repl_screenshot_evidence/0`, `remove/e2e_silent_fallback_no_aside_message/1`, `remove/screenshot_evidence_lines`, `remove/tester_sentence`, `strip/bounded_output`, `strip/bypass_sentence_line`, `strip/chain_phase2_line`, `strip/chain_phase3_line`, `strip/every_enumerated_site`, `strip/execution_owner`, `strip/missing_toolchain_header`, `strip/summary_step4_line`, `swap/silently_for_carve_out`. PASS (the control adds a bad line or removes a literal the baseline already carries, so it does not depend on the e2e edit): `add/install_line_0..2`, `add/note_line_0..6`, `add/silently_site_line`, `append/precedence_sentence_to_site_line`, `delete/sites_below_minimum`, `remove/e2e_silent_fallback_no_aside_message/0` (the word `silently` already occurs elsewhere in the unedited workflow, in the retry rule; the precedence-line requirement of the silence rule is what the unedited text fails). The skill-side subtests (the nine M2 rules and the M2 `negative_controls` group) still PASS; `e2e_no_install_command` PASSes at birth because the unedited workflow has no Aside line (an invariant, like the M1 guards).

Static checks on the RED tree: `gofmt -l internal/template/aside_skill_policy_test.go` printed nothing, exit 0; `go vet ./internal/template/` printed nothing, exit 0. The M3 RED commit (subject `test(SPEC-ASIDE-BROWSER-001): M3 RED - ...`) carries only the test file and this record, and precedes the GREEN commit in `git log`.

#### M3.2 (M3 GREEN) `e2e.md` (template, then root by `cp`)

Every site is carved out by an appended parenthetical `(except Aside: ...)`, so stripping it restores the original line byte for byte (the `strip/*` controls assert this per site). Edited sites, found by the E16 command and re-read before editing (new line numbers after the edit): execution owner (36); bounded output (54); probe introduction (90); missing-toolchain sequence header (108); the `--tool` bypass sentence (119); Phase 2 delegation (207); Phase 3 delegation (229); Phase 4 recording delegation (287); Agent Chain Phase 0 (338), Phase 2 (341), Phase 3 (342), Phase 4 (344); Execution Summary step 4 (352), steps 6 and 7 (354, 355). New text: the flag list at line 40 gains `aside` with the explicit-only sentence; an Aside paragraph after the bypass sentence (a lead line plus six items: entry, execution with `Skill("moai-ref-aside-browser")` loaded first and the availability probe as a sentence rather than a probe-table row and "never installs", output to a file under `e2e/.runs/`, screenshot evidence at the hedged wording, the `CI=true` clause, and the precedence sentence of plan.md M3.2 verbatim); a Phase 0.5 option rule (never offered, never recommended, never auto-detected); a Tool Matrix row (web, CLI via `aside repl`, orchestrator-run, explicit-only, evidence = a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path). The evidence wording stays hedged because the orchestrator's measurement left screenshot persistence unconfirmed (M3.0 above).

After the edit the E16 pattern with `-n` prints 16 lines: the 15 baseline lines (all carry `except Aside`) plus the precedence sentence (line 127), which is the one line on which `silently` counts. `grep -c 'except Aside'` printed `15`. Lines outside the Aside path that stay unqualified (NF2), each confirmed by the same enumeration not selecting them and each read: detection delegation (line 60) and Execution Summary step 2 (350), journey-mapping delegation (175), its custom-journey item (190), Agent Chain Phase 1 (340) and Execution Summary step 5 (353) are Phases 0 and 1, which the e2e-tester keeps; the report rendering (300) is the orchestrator's own Phase 5; the Install step (112) is a step of the missing-toolchain sequence whose header (108) and the precedence sentence already exclude Aside. These correspond to the plan's `:60 :165 :330 :340 :343` and the three lines the plan omitted (`:180`, `:112`, `:290`).

Size: `wc -c` printed `28417` for `e2e.md` (budget 40,000) and `10660` for `e2e-tester.md`.

#### M3.3 `e2e-tester.md`, Codex emit, build

- One sentence, placed after the selection-is-out-of-scope paragraph, identically in the template and root files: `A subagent never invokes Aside (the `aside` CLI or the Aside MCP tools); a task that needs Aside returns a blocker report to the orchestrator.`
- `make agents-emit` exit 0 (the golden test regenerated `internal/template/templates/.codex/agents/moai/e2e-tester.toml`; its diff is that one added sentence). `make agents-emit-check` exit 0 (`ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.364s`). `make build` exit 0, its log ends with the `catalog.yaml updated successfully (14151 bytes)` line and the `go build` line. Immediately after it, plain `git status --short` listed exactly: the two `e2e.md` files, the two `e2e-tester.md` files, `internal/template/catalog.yaml`, and `e2e-tester.toml`. No root Codex mirror exists for that TOML, and `make build` regenerated no other file.
- `grep -c 'never invokes Aside'` printed `1` for the TOML, the template `e2e-tester.md`, and the root `e2e-tester.md`.

#### M3.4 mirror and catalog

- `cmp internal/template/templates/.claude/skills/moai/workflows/e2e.md .claude/skills/moai/workflows/e2e.md` printed nothing, exit 0.
- `diff` of the e2e-tester pair prints `2d1` (`< isolation: worktree`) and `146c145` (the task-tracking line): the same 3 baseline lines as before (the second hunk moved from 144/143 because the sentence added two lines above it), and nothing else.
- `git diff -U4 -- internal/template/catalog.yaml`: exactly two hunks, `moai` (`hash: ae8aa96c...` to `ea148045...`, because `workflows/e2e.md` is inside the core `moai` skill directory) and `e2e-tester` (`hash: c5bb26f2...` to `2bcd5d66...`); `shasum -a 256` of the template `e2e-tester.md` printed `2bcd5d661acbc37766d5e3c9b861097c025b7a6a586d4846cd5bc080d8cb10a2`, equal to the new hash. No other `hash:` hunk and no `generated_at` hunk; with the M2 entry the full changed set across the card is `{moai-ref-aside-browser, moai, e2e-tester}`.

#### M3 verification (tree HEAD `8159b4d3d` plus the uncommitted M3 GREEN changes)

- Guard test, `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, exit 0: `--- PASS:` printed for `description_within_listing_cap`, `full_access_prohibited`, `guard_by_omission`, `repl_read_only_limit`, `write_confirmation_via_orchestrator`, `orchestrator_only_operation`, `subagent_never_invokes_skill`, `no_auto_install_advise_only`, `no_credentials_in_outputs`, `negative_controls`, `subagent_never_invokes_tester`, `e2e_explicit_only`, `e2e_ci_excluded`, `e2e_orchestrator_executes_aside`, `e2e_execution_owner_carveout`, `e2e_aside_output_bounded`, `e2e_silent_fallback_no_aside_message`, `e2e_missing_toolchain_carveout`, `e2e_tool_bypass_line_carveout`, `e2e_every_site_carved_out`, `e2e_no_install_command`, `e2e_repl_screenshot_evidence`, and `e2e_negative_controls`; the last line `ok  	github.com/modu-ai/moai-adk/internal/template	0.334s`. `grep -c` over the `e2e_negative_controls/` PASS lines printed `36` and `grep -c '--- FAIL'` printed `0`.
- The 13-selector command of AC-ASB-003, exit 0, `grep -c '^--- PASS'` printed `13`, last line `ok  	github.com/modu-ai/moai-adk/internal/template	0.235s`.
- M1 tests: `go test ./internal/template/ -run '^(TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1` printed `--- PASS: TestMCPNeutralityTemplateShape` and `--- PASS: TestMCPDefaultExcludesAside`, exit 0; `go test ./internal/cli/ -run '^(TestMCP_Add_AsideDocumentedCommandLine|TestMCP_Add_IdempotentSkip)$' -v -count=1` printed `--- PASS` for both and for the `project` and `user` subtests, exit 0.
- Package-wide `go test ./internal/template/ -count=1` (once, under the `go-test-internal-template` slot lease, acquired and released): exit 0, `ok  	github.com/modu-ai/moai-adk/internal/template	118.293s`.
- `go vet ./internal/template/` printed nothing, exit 0; `gofmt -l internal/template/aside_skill_policy_test.go` printed nothing, exit 0; `golangci-lint run --timeout=5m --new-from-rev=1c023f7ba ./internal/template/` printed `0 issues.`, exit 0 (linter `v2.1.6`, built with go1.26.8).
- `moai spec lint SPEC-ASIDE-BROWSER-001 --strict` with a binary built by `go build -o <scratch>/moai-scratch ./cmd/moai` from this tree (exit 0; no commit stamp, built without the Makefile `LDFLAGS`, so its judging-build coordinate is "tree HEAD `8159b4d3d` plus uncommitted M3 GREEN changes"): exit 0, output `✓ No findings — all SPEC documents are valid`.
- Neutrality grep of AC-ASB-013 (`SPEC-[A-Z]|t1439|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{40}\b|\b[0-9a-f]{7,8}\b`): the skill printed nothing (exit 1); the template `e2e-tester.md` printed nothing (exit 1); the TOML's one added line (`git diff -U0`) contains none of the patterns; the template `e2e.md` printed exactly the two pre-existing lines (`13:  updated: "2026-07-14"` and `315:  - Log: e2e/.runs/20260101-120000-suite.log ...`, the example log path that was line 305 at baseline), exit 0, no new hit.

Checker-blindness probes (run on the GREEN tree, each reverted with `git checkout --` of the committed test file, after which `git diff --stat` of that file printed nothing): (a) disabling the carve-out tail check made `e2e_negative_controls/mutate/bounded_tail_without_runs_dir` print `--- FAIL`; (b) making the CI scope check always accept made `e2e_negative_controls/move/ci_line_to_no_flag_branch` print `--- FAIL`; (c) letting `silently` satisfy the site rule made `add/silently_site_line`, `swap/silently_for_carve_out`, and `append/precedence_sentence_to_site_line` print `--- FAIL`. Each control therefore bites the defect it names.

#### Gaps (M3)

- Whether `aside repl` can persist a screenshot to a path stays UNCONFIRMED: the orchestrator's single measurement ended in a CDP capture timeout (about 35 s) with no file written, which is neither a confirmation nor a "path unsupported" error. The workflow wording is the hedged form, so the residual claim is only "a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path"; whether Aside or the REPL script writes the bytes is not asserted. No `aside` command or Aside MCP tool was run by manager-develop.
- Guard-refused command: one `sed` over a log path held in a shell variable was refused by the worktree guard; the log was read with the Read tool instead. No verification claim rests on a refused command.
- Not run: the `internal/cli` package suite and `go test ./...` (by instruction); the repository-wide verdict is owned by the CI run on the project's integration branch and is PENDING at report time. The docs-site guides (AC-ASB-014) belong to the next milestone and were not touched.
- AC-ASB-006 to AC-ASB-012 are carried by `TestAsideSkillPolicyAnchors` on the template copies; the root mirrors are covered by `cmp` (e2e.md) and by the `grep -c` of the sentence plus the baseline-only `diff` (e2e-tester).

#### Residual risk (M3)

- The checkers are lexical: they prove the workflow and the e2e-tester definition say what the anchors require, carve out every enumerated site, and instruct nothing the forbidden patterns name. They cannot prove what a model emits at run time; no hook enforces the orchestrator-only boundary.
- The site enumeration is a role-based regexp, so a future delegating line phrased outside it is not caught; the minimum-count guard only stops an emptied enumeration, not a narrowed one.
- Orchestrator-run Aside steps enlarge the orchestrator's context; the workflow bounds that by redirecting output to `e2e/.runs/` and citing screenshots by path, which is text, not enforcement.
- `aside repl` has no permission flag, so its read-only limit is discipline only (stated in the skill).

### M3.5 (docs-site guides, four locales) — commit subject `docs(SPEC-ASIDE-BROWSER-001): document the optional Aside MCP server in the MCP guide (card t1439)`

Measured by manager-develop (run as a general-purpose agent), branch `WT-aside-browser-cli`, worktree t1439, unless a line names another measurer. No `aside` command was run. The `hugo` runs used `--source docs-site --destination <scratch>` (no `cd`, so nothing was written into the tree).

#### Pre-flight

- `git rev-parse --short HEAD` printed `2231bc22e` (M3 GREEN); `git branch --show-current` printed `WT-aside-browser-cli`; `git status --short` printed nothing before any edit.
- Baseline lines re-read before editing (verbatim; line 63 heading, line 65 sentence, same line numbers in all four locales):
  - ko: `### 네 가지 documented-but-disabled 엔트리` / `배포 기본값은 \`moai\` 서버 하나만 활성입니다. 네 개의 외부 서버는 ...`
  - en: `### Four documented-but-disabled entries` / `The distribution default activates only the \`moai\` server. Four external servers are ...`
  - ja: `### 4つの documented-but-disabled エントリ` / `配布デフォルトは \`moai\` サーバー1つだけがアクティブです。4つの外部サーバーは ...`
  - zh: `### 四个 documented-but-disabled 条目` / `部署默认值只有 \`moai\` 一个服务器处于活跃状态。四个外部服务器已 ...`
- Baseline `hugo --minify --gc --source docs-site --destination <scratch>/hugo-base` on the unedited tree: exit 0; `grep -c -i -E 'WARN|ERROR'` over its log printed `0`; page counts 189 / 187 / 187 / 187 (ko / en / ja / zh); `git status --short` afterwards printed nothing.

#### Edits

Per page, nothing else: the heading numeral and the sentence numeral changed to five (`다섯 가지` / `다섯 개의`, `Five` / `Five`, `5つの` / `5つの`, `五个` / `五个`), one table row for `aside` appended to the table, and one paragraph after the table carrying the exact registration command, the two tools, the read-only default, the orchestrator-only confirmation, the no-install and `aside skills install` operator sentence, and the explicit-only `/moai e2e --tool aside` (unavailable when `CI=true`, silent fallback) sentence. No heading added. Terminology follows each locale's existing pages (ko 운영자 / 서브에이전트 / MoAI 오케스트레이터; ja 運用者 / サブエージェント / オーケストレーター; zh 操作者 / 子智能体 / 编排器).

#### Verification (tree HEAD `2231bc22e` plus the uncommitted four-page edit)

- `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user'` over the four pages: `2` each (table row and paragraph), exit 0.
- Numeral greps (`Five documented-but-disabled`, `다섯 가지 documented-but-disabled`, `5つの documented-but-disabled`, `五个 documented-but-disabled`): `1` each. The old numerals (`Four external servers`, `네 개의 외부 서버`, `4つの外部サーバー`, `四个外部服务器`, and the four old heading forms): `0` in all four pages. The new sentence forms (`Five external servers`, `다섯 개의 외부 서버`, `5つの外部サーバー`, `五个外部服务器`) each print on line 65 of their page.
- `hugo --minify --gc --source docs-site --destination <scratch>/hugo-after`: exit 0; `grep -c -i -E 'WARN|ERROR'` over its log printed `0`; `<scratch>/hugo-after/sitemap.xml` exists (`sitemap OK`).
- Section-count parity of the four pages: `grep -c '^#\{2,\} '` printed `26` for each. Whole-tree ratchet (section 4 of the verify recipe, run without `cd`): 51 divergent pages now, 51 in `docs-site/.locale-parity-baseline`, and `comm -23` printed nothing.
- URL blacklist grep over `docs-site/content` and `README*.md`: no output, exit 1. Mermaid direction grep: no output, exit 1. Body-emoji scan over the four pages: no output, exit 1.
- `moai spec lint SPEC-ASIDE-BROWSER-001 --strict`, binary built by `go build -o <scratch>/moai-m35 ./cmd/moai` from this tree (exit 0; no commit stamp, built without the Makefile `LDFLAGS`; judging-build coordinate: tree HEAD `2231bc22e` plus uncommitted docs edits): exit 0, output `✓ No findings — all SPEC documents are valid`.

#### Gaps (M3.5)

- Not run: any Go test suite (by instruction); the version-string check (section 6) and the README parity check, which this change does not touch. The body-emoji scan covered the four edited pages, not the whole tree.
- The AC names `cd docs-site && hugo ...` and `test -f docs-site/public/sitemap.xml`; the `cd` form was avoided (worktree guard discipline), so the build wrote to a scratch destination and the sitemap was tested there.
- Native-idiom review of the ja and zh paragraphs is by the author only; no native-speaker or humanize pass was run.

#### Residual risk (M3.5)

- The paragraph repeats the facts of the skill in prose; if the skill or `e2e.md` wording changes, the four pages carry no mechanical guard against drifting from it (the only guards are the command and numeral greps above).

### Final acceptance matrix (all 14 ACs) and cross-cutting items (§E E1-E8)

Measured by manager-develop (run as a general-purpose agent), in this run, on tree HEAD `5629a0d6f` (branch `WT-aside-browser-cli`, worktree t1439, base develop `4bf547bca`), `git status --short` empty before and after every step. Every command below was run in this run against this tree; output lines are verbatim. No `aside` command or Aside MCP tool was run by this agent (REQ-ASB-007); the orchestrator's measurements (`aside --version` = `1.26.916.1741`; the M3.0 screenshot attempt = CDP timeout, persistence UNCONFIRMED) are recorded above and not re-measured. Evidence-ledger rows E11/E15 contain table-escaped pipes in acceptance.md; they were run with plain `|` and no backslashes (N5). Scratch outputs live under the session scratchpad (machine-local, never cited).

Pre-flight: `pwd` = `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1439`; `git branch --show-current` = `WT-aside-browser-cli`; `git rev-parse --short HEAD` = `5629a0d6f`; `git status --short` empty; `go build ./...` exit 0.

Judging-build provenance (verification-claim-integrity §2.2): Go tests, `go vet`, `gofmt`, `golangci-lint v2.1.6` read the tree directly. `moai spec lint` was run with `./bin/moai`, built by `make build` from this tree; `./bin/moai version` printed build id `archive/t1401-243-g5629a0d6f`, built `2026-10-02T15:40:27Z`, so its commit equals the tree HEAD `5629a0d6f`. The installed `moai` on PATH was not used.

Ordering witness for the RED-now cells: every RED-now cell is pinned at tree `4bf547bcad7c155b1e91485921569db709ec3ac2` (document-level pin of acceptance.md; `git rev-parse 4bf547bca` printed that full SHA). Commit graph, `git log --reverse --format='%h %s' 4bf547bca..HEAD` (verbatim, oldest first):

```
c9d4dd5d2 docs(spec): add SPEC-ASIDE-BROWSER-001 plan artifacts (card t1439)
2bba87fcd docs(spec): close plan-audit N1/N2/N6 for SPEC-ASIDE-BROWSER-001 (card t1439)
59d69f1dd docs(spec): record SPEC-ASIDE-BROWSER-001 mode selection and Kickoff (card t1439)
1fa6e31ba test(SPEC-ASIDE-BROWSER-001): M1 RED-record - Aside default-exclusion guard and documented-command test (card t1439)
5b932b06c test(SPEC-ASIDE-BROWSER-001): M2 RED - Aside skill policy checker and anchor guard (card t1439)
1c023f7ba feat(SPEC-ASIDE-BROWSER-001): M2 GREEN - optional Aside browser policy skill and core catalog entry (card t1439)
8159b4d3d test(SPEC-ASIDE-BROWSER-001): M3 RED - e2e wiring checker, tester anchor, and controls (card t1439)
2231bc22e feat(SPEC-ASIDE-BROWSER-001): M3 GREEN - e2e workflow Aside wiring and tester sentence (card t1439)
0de4cfabe docs(SPEC-ASIDE-BROWSER-001): document the optional Aside MCP server in the MCP guide (card t1439)
5629a0d6f docs(SPEC-ASIDE-BROWSER-001): correct the MoAI brand spelling in the Aside guide paragraph (card t1439)
```

`git merge-base --is-ancestor 5b932b06c 1c023f7ba` exit 0 (M2 RED precedes M2 GREEN); `git merge-base --is-ancestor 8159b4d3d 2231bc22e` exit 0 (M3 RED precedes M3 GREEN); `git merge-base --is-ancestor 4bf547bca 1fa6e31ba` exit 0 (the measurement pin precedes M1). M1 is a single `RED-record` commit by plan §F (its guards pass at birth, so it has no GREEN commit; the mutant runs recorded in M1 above are the observed-failure evidence). What the graph cannot witness: the order of the failing run inside a commit.

#### Per-AC results (each AC's own When command, as stated in acceptance.md)

- AC-ASB-001 PASS. `go test ./internal/template/ -run '^(TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1` exit 0: `--- PASS: TestMCPNeutralityTemplateShape (0.00s)`, `--- PASS: TestMCPDefaultExcludesAside (0.00s)`, `ok  	github.com/modu-ai/moai-adk/internal/template	0.255s`; no `[no tests to run]`. Cross-check `grep -c -i aside internal/template/templates/.mcp.json internal/template/templates/.claude/settings.json.tmpl` printed `internal/template/templates/.mcp.json:0` and `internal/template/templates/.claude/settings.json.tmpl:0`, exit 1. RED-now cell E1 pinned at `4bf547bca` (guard did not exist; the invariant was green by design). Mutants (a) and (b) were observed in M1 above (both `FAIL` for (a); only the new test `FAIL` for (b)).
- AC-ASB-002 PASS. `go test ./internal/cli/ -run '^TestMCP_Add_AsideDocumentedCommandLine$' -v -count=1` exit 0: `--- PASS: TestMCP_Add_AsideDocumentedCommandLine (0.00s)`, `--- PASS: TestMCP_Add_AsideDocumentedCommandLine/project (0.00s)`, `--- PASS: TestMCP_Add_AsideDocumentedCommandLine/user (0.00s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	0.847s`. `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md` printed `1`, exit 0. The user-scope subtest runs through the `userHomeDirFn` seam only (see Gaps). The full `TestMCP_Add_` prefix run, exit 0, printed `--- PASS` for `TestMCP_Add_AsideDocumentedCommandLine`, `TestMCP_Add_ConcurrentWriter`, `TestMCP_Add_RegistersEntry_PreservesUnrelated`, `TestMCP_Add_SecretRejection`, `TestMCP_Add_HTTPType`, `TestMCP_Add_IdempotentSkip`, `TestMCP_Add_SecretLiteralAccepted`. RED-now cells E2 and E3 pinned at `4bf547bca`; the `--arg` mutant was observed in M1 above.
- AC-ASB-003 PASS. The 13-selector command of the AC, redirected to a file, exit 0, `grep -c '^--- PASS'` printed `13`; PASS lines: `TestGitignore_IgnoresSkillMirrorOnly`, `TestPublishedSkillsNamesMatchTree`, `TestSkillMirror_SetIsDerivedNotConstant`, `TestEmbeddedMoaiSkillNames`, `TestSlimFS_HidesNonCoreEntries`, `TestCatalogNoDuplicateEntries`, `TestCatalogTierValid`, `TestCatalogReferencesValid`, `TestSlimFS_PreservesCoreEntries`, `TestAllSkillsInCatalog`, `TestWorkflowTriggerCoverage`, `TestCatalogHashCoversSkillSubfiles`, `TestManifestHashFormat`; last line `ok  	github.com/modu-ai/moai-adk/internal/template	0.267s`. `grep -n 'user-invocable: false' <skill>` printed `14:user-invocable: false`. `grep -n -A3 'name: moai-ref-aside-browser' internal/template/catalog.yaml` printed `56:            - name: moai-ref-aside-browser`, `57-              tier: core`, `58-              path: templates/.claude/skills/moai-ref-aside-browser/`, `59-              hash: 653940540ae35059faa4dcdd7d4f2bf4d7b4960de023f6da405e9632c0030b5f` (64 hex). `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1` exit 0 printed `--- PASS: TestAsideSkillPolicyAnchors/description_within_listing_cap (0.00s)`. RED-now cells E3 and E6 pinned at `4bf547bca`.
- AC-ASB-004 PASS. In the anchors run (exit 0): `--- PASS: TestAsideSkillPolicyAnchors/full_access_prohibited (0.00s)`, `.../guard_by_omission (0.00s)`, `.../repl_read_only_limit (0.00s)`. Reader check `grep -n 'full-access' <skill>` printed one line, `39:1. Permission. Never pass `--permission full-access`, and never request full-access by any other route (settings, prompts, or tool arguments).`, exit 0 (it carries both allow-listed literals). Mutants are the in-test controls (18 skill controls, all `--- PASS`, see below). RED-now cell E8 pinned at `4bf547bca`.
- AC-ASB-005 PASS. `--- PASS: TestAsideSkillPolicyAnchors/write_confirmation_via_orchestrator (0.00s)` and `.../orchestrator_only_operation (0.00s)`. RED-now cell E8.
- AC-ASB-006 PASS. `--- PASS: TestAsideSkillPolicyAnchors/subagent_never_invokes_skill (0.00s)` and `.../subagent_never_invokes_tester (0.00s)`. `grep -c 'never invokes Aside' <skill> <template e2e-tester.md> <root e2e-tester.md>` printed `internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md:1`, `.claude/agents/moai/e2e-tester.md:1`, `internal/template/templates/.claude/agents/moai/e2e-tester.md:1`, exit 0. RED-now cells E5 and E12 pinned at `4bf547bca`.
- AC-ASB-007 PASS. `--- PASS: TestAsideSkillPolicyAnchors/no_auto_install_advise_only (0.00s)` and `.../no_credentials_in_outputs (0.00s)`. `grep -c 'aside skills install' <skill>` printed `1`, exit 0; the one line is `43:5. Install. Never install Aside and never run install tooling for it. If the operator wants Aside's own browser skill added to their coding agents, advise the operator to run `aside skills install`; the orchestrator does not run it.` (contains `operator`). RED-now cells E8 and E3.
- AC-ASB-008 PASS. `--- PASS: TestAsideSkillPolicyAnchors/e2e_explicit_only (0.00s)` and `.../e2e_ci_excluded (0.00s)`. RED-now cell E4 pinned at `4bf547bca`.
- AC-ASB-009 PASS. `--- PASS: TestAsideSkillPolicyAnchors/e2e_orchestrator_executes_aside (0.00s)`, `.../e2e_execution_owner_carveout (0.00s)`, `.../e2e_aside_output_bounded (0.00s)`, `.../e2e_every_site_carved_out (0.00s)`. RED-now cells E4 and E13 pinned at `4bf547bca`.
- AC-ASB-010 PASS. `--- PASS: TestAsideSkillPolicyAnchors/e2e_silent_fallback_no_aside_message (0.00s)`, `.../e2e_missing_toolchain_carveout (0.00s)`, `.../e2e_tool_bypass_line_carveout (0.00s)`, `.../e2e_every_site_carved_out (0.00s)`, `.../e2e_no_install_command (0.00s)`. The AC's reader grep (`grep -n -E '[Dd]elegate .*(script creation|test execution|execution|recording)|Phase [234]: e2e-tester|[Mm]issing[- ]toolchain|Execution owner|Bounded output|toolchain probe/install|probes the DEFAULT toolchain' <template e2e.md>`, plain `|`), exit 0, listed 16 lines at 36, 54, 90, 108, 119, 127, 207, 229, 287, 338, 341, 342, 344, 352, 354, 355: 15 carry `except Aside` (the baseline 15 sites, at shifted line numbers) and line 127 is the precedence sentence (`... continue silently on the platform default, with no Aside-specific message, prompt, install attempt, or failure; ...`), the one line on which `silently` counts. None of the 16 is unqualified. RED-now cells E4, E13, E16 pinned at `4bf547bca` (the enumeration printed 15 lines and `except Aside` counted 0 there).
- AC-ASB-011 PASS. `--- PASS: TestAsideSkillPolicyAnchors/e2e_repl_screenshot_evidence (0.00s)`. The wording is the hedged form (see M3.2); the M3.0 measurement gate was run by the orchestrator and did not confirm persistence (recorded above; Gaps below). RED-now cell E4.
- AC-ASB-012 PASS. `cmp <template skill> .claude/skills/moai-ref-aside-browser/SKILL.md` exit 0 (no output); `cmp <template e2e.md> .claude/skills/moai/workflows/e2e.md` exit 0 (no output); `diff <template e2e-tester.md> .claude/agents/moai/e2e-tester.md` exit 1 and printed exactly `2d1` (`< isolation: worktree`) and `146c145` (the task-tracking line): 3 changed lines, the same baseline as the AC's `2d1`/`144c143` (the second hunk reads 146c145 because the added sentence sits two lines above it; the sentence is present identically in both files, `grep -c` printed `1` for each). `grep -c 'never invokes Aside' internal/template/templates/.codex/agents/moai/e2e-tester.toml` printed `1`, exit 0. `make agents-emit-check` exit 0 (`ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.384s`). `make build` exit 0 (log ends `catalog.yaml updated successfully (14151 bytes)` and the `go build` line); `git status --short` after it printed nothing (no tracked file changed, no regeneration drift). `git diff -U4 4bf547bca HEAD -- internal/template/catalog.yaml` printed exactly three hunks: `moai` (`hash: ae8aa96c...` to `ea148045...`), the new `moai-ref-aside-browser` entry (`tier: core`, hash `653940540ae35059faa4dcdd7d4f2bf4d7b4960de023f6da405e9632c0030b5f`), and `e2e-tester` (`hash: c5bb26f2...` to `2bcd5d661acbc37766d5e3c9b861097c025b7a6a586d4846cd5bc080d8cb10a2`); no other `hash:` hunk, no `generated_at` hunk. RED-now cells E7 and E10 pinned at `4bf547bca`. The one-byte-change `cmp` mutant was not re-run in this closing pass (see Gaps).
- AC-ASB-013 PASS. `go test ./internal/template/ -count=1` (once, under the `go-test-internal-template` slot lease: acquired `until 2026-10-02T15:56:54Z`, released, `moai slot release` printed `released`), exit 0: `ok  	github.com/modu-ai/moai-adk/internal/template	125.192s`; the three subpackages, `go test ./internal/template/agentemit ./internal/template/commandemit ./internal/template/scripts -count=1`, exit 0: `ok  	.../agentemit	0.232s`, `ok  	.../commandemit	0.087s`, `?   	.../scripts	[no test files]`. Neutrality grep (`SPEC-[A-Z]|t1439|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{40}\b|\b[0-9a-f]{7,8}\b`, plain `|`): skill exit 1 (no output); `e2e-tester.md` exit 1 (no output); `e2e-tester.toml` exit 1 (no output over the whole file); `e2e.md` exit 0 and printed exactly the two pre-existing baseline lines (E15), `13:  updated: "2026-07-14"` and `315:  - Log: e2e/.runs/20260101-120000-suite.log (bounded excerpt above; full log at path)`, no new line. RED-now: not applicable (regression-guard by the AC).
- AC-ASB-014 PASS. `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user'` over the four guides printed `docs-site/content/ko/guides/mcp-server.md:2`, `.../zh/...:2`, `.../en/...:2`, `.../ja/...:2` (table row plus paragraph), exit 0. Numeral greps printed `1` for each of `Five documented-but-disabled` (en), `다섯 가지 documented-but-disabled` (ko), `5つの documented-but-disabled` (ja), `五个 documented-but-disabled` (zh). Old numerals `Four external servers`, `네 개의 외부 서버`, `4つの外部サーバー`, `四个外部服务器` each printed `0`, exit 1. `hugo --minify --gc --source docs-site --destination <scratchpad>/hugo-final` exit 0, `grep -c -i -E 'WARN|ERROR'` over its log printed `0`, `test -f <scratchpad>/hugo-final/sitemap.xml` exit 0 (the build wrote to a scratch destination, never into the tree). Locale-parity ratchet (section 4 of the `hns-oss-docs-verify` recipe, run without `cd` by prefixing paths and stripping `docs-site/content/`): `comm -23 <now> <baseline>` printed nothing (exit 0); positive control: 51 divergent pages now and 51 in `docs-site/.locale-parity-baseline` (`wc -l`). RED-now cells E9 and E14 pinned at `4bf547bca`.

Negative controls and mutants that exist as named subtests: the anchors run, `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, exit 0, printed `--- PASS: TestAsideSkillPolicyAnchors (0.18s)`, 22 named rule subtests `--- PASS` (listed per AC above) including `.../negative_controls (0.00s)` and `.../e2e_negative_controls (0.17s)`; `grep -c` over `.../negative_controls/` PASS lines printed `18`, over `.../e2e_negative_controls/` PASS lines printed `36`, and `grep -c -- '--- FAIL'` over the whole run printed `0`. Every control produced its named violation or the run would not be green.

#### Cross-cutting items (E1-E8)

- E1 (AC matrix): the 14 results above, 14 PASS, 0 FAIL, each with a `--- PASS:` line or a grep count observed; no selector matched zero tests.
- E2 (builds): `go build ./...` exit 0. `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The Windows cross-build does not apply to this change in substance: `git diff --stat 4bf547bca HEAD -- '*.go'` lists only the three test files (`internal/cli/mcp_test.go` 114 insertions, `internal/template/aside_skill_policy_test.go` 912, `internal/template/mcp_template_neutrality_test.go` 43), and `grep -rn -E 'syscall|go:build'` over those three files printed nothing (exit 1): no `syscall` use and no build-tag file added.
- E3 (coverage): not applicable. The change is text assets (skill, workflow, agent definition, docs-site) plus test-only Go files; no non-test Go source was added or modified (`git diff --stat 4bf547bca HEAD -- '*.go'` above), so no coverage figure exists to measure for it. No coverage command was run.
- E4 (subagent boundary): `grep -n 'AskUserQuestion' internal/template/aside_skill_policy_test.go internal/cli/mcp_test.go internal/template/mcp_template_neutrality_test.go` printed one hit, `internal/cli/mcp_test.go:21://   - AC-TMC-010: TestMCP_NoAskUserQuestion (in mcp_boundary_test.go).`, a comment; `git grep -n -c 'AskUserQuestion' 4bf547bca -- internal/cli/mcp_test.go` printed `4bf547bca:internal/cli/mcp_test.go:1`, so the hit is pre-existing at baseline. No non-comment hit.
- E5 (lint): `golangci-lint run --timeout=5m --new-from-rev=4bf547bca ./internal/template/ ./internal/cli/` exit 0, output `0 issues.` (linter `v2.1.6`, built with go1.26.8). `go vet ./internal/template/ ./internal/cli/` exit 0, no output. `gofmt -l` over the three touched Go files printed nothing, exit 0. `./bin/moai spec lint SPEC-ASIDE-BROWSER-001 --strict` exit 0, output `✓ No findings — all SPEC documents are valid` (judging build commit `5629a0d6f` = tree HEAD, see provenance above). `make agents-emit-check` exit 0 and `make build` exit 0 with empty `git status --short` after it (AC-ASB-012).
- E6 (branch HEAD and push state): HEAD `5629a0d6f`, branch `WT-aside-browser-cli`; this card's commits are listed in §E.3; nothing was pushed (the factory leader batch-pushes `develop`).
- E7 (blockers): none. No command failed for an unexplained reason.
- E8 (RED output, TDD): the verbatim pre-GREEN failing runs are recorded above in M2.1 (RED, `skill file unreadable` and eleven `--- FAIL` lines, exit 1) and M3.1 (RED, twelve `--- FAIL` lines plus failure texts, exit 1); M1 is the guards-pass-at-birth `RED-record` commit with mutant failures recorded in M1.1 and M1.2. The ordering is witnessed by the commit graph above, not by this record.
- Mirror and default-config invariants: `grep -c -i aside` printed `0` for `internal/template/templates/.mcp.json`, `internal/template/templates/.claude/settings.json.tmpl` and `internal/template/templates/.moai/config/sections/delegation.yaml`; `git diff --stat 4bf547bca HEAD --` over those three plus root `.moai/config/sections/delegation.yaml`, `.mcp.json`, `.claude/settings.json` printed nothing (unchanged). No `moai mcp` Go source changed (Go diff above is test-only).
- AC-baseline guard: `go test ./internal/spec -run 'TestACCounterFullCorpusMatchesBaseline' -count=1 -v` exit 0, `--- PASS: TestACCounterFullCorpusMatchesBaseline (9.71s)`, `ok  	github.com/modu-ai/moai-adk/internal/spec	10.074s`; the new acceptance.md reports as `ac_count_clause_test.go:532:   absent-from-snapshot .moai/specs/SPEC-ASIDE-BROWSER-001/acceptance.md: COUNT 14` (report only; absent-from-snapshot does not fail the guard).

#### Guard-refused commands (verification-claim-integrity §3.1)

Two of my commands were refused by the worktree guard before running: a `go test` line carrying a shell variable assignment and `$S` expansion (two attempts, AC-ASB-001 and AC-ASB-002). I re-ran both as plain commands with literal paths, and the results above are from those re-runs. No claim rests on a refused command. The `cd`-based forms named in AC-ASB-014 and in the verify recipe were not attempted (instruction: no `cd`); the equivalent no-`cd` forms are recorded in the AC-ASB-014 result.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-10-02T15:44:53Z
- run_status: audit-ready
- final HEAD measured: `5629a0d6f` (branch `WT-aside-browser-cli`, worktree t1439, base develop `4bf547bca`); the acceptance matrix in §E.2 was measured against this HEAD. The commit that records this signal cannot cite its own hash and lands after `5629a0d6f` (it changes only this file).
- commits (oldest first): plan `c9d4dd5d2`, `2bba87fcd`; Kickoff record `59d69f1dd`; M1 `1fa6e31ba` (RED-record); M2 RED `5b932b06c`, M2 GREEN `1c023f7ba`; M3 RED `8159b4d3d`, M3 GREEN `2231bc22e`; docs `0de4cfabe`, brand-spelling fix `5629a0d6f`; then the run-phase record commit `chore(SPEC-ASIDE-BROWSER-001): record run-phase acceptance matrix and audit-ready signal (card t1439)`.
- result: 14 of 14 acceptance criteria PASS (§E.2 final matrix); changed hash set exactly `{moai-ref-aside-browser, moai, e2e-tester}`; package-wide `go test ./internal/template/ -count=1` exit 0 (`ok ... 125.192s`); no tracked file changed by `make build`.
- Gaps (explicitly not observed or not claimed):
  - Whether `aside repl` can persist a screenshot to a path is UNCONFIRMED: the orchestrator's single measurement ended in a CDP timeout (about 35 s), no file written, which is neither confirmation nor a "path unsupported" error. The workflow wording is the hedged form; the SPEC text was not amended.
  - The user-scope run of `moai mcp add aside ... --scope user` was measured only through the `userHomeDirFn` seam against a temporary home (HOME override is refused by the worktree guard); no run against a real user config file.
  - Every checker is lexical (static text anchors, forbidden patterns, per-site carve-out enumeration). None can prove what a model emits or does at run time, including the silent fallback and the read-only discipline of `aside repl`.
  - No hook enforces the orchestrator-only boundary (REQ-ASB-007/010); the guard is the sentences in the skill, in `e2e.md`, and in the e2e-tester definition, pinned by tests whose deletion mutants fail.
  - The ja and zh docs-site paragraphs were not run through the humanize pass (author review only).
  - The CI verdict on `origin/develop` is PENDING: only the package-wide `internal/template` suite, its three subpackages, named `internal/cli` selectors (the `TestMCP_Add_` prefix), the AC-baseline guard test and the build/lint targets above were run; `go test ./...` and the whole `internal/cli` suite were not.
  - The AC-ASB-012 one-byte `cmp` mutant, and the AC-ASB-001 mutants (a) and (b) and the AC-ASB-002 `--arg` mutant, were not re-run in this closing pass; their observed results are the M1 and M2 records above, from earlier runs on earlier trees.
  - Behavior of the Aside MCP `exec` tool's own permission argument is unmeasured (out of scope).
  - A Windows run-time check was not performed; only the cross-build `GOOS=windows GOARCH=amd64 go build ./...` (exit 0).
- residual risk: see the per-milestone residual-risk lists above (lexical checkers; role-based site enumeration can miss a future phrasing; orchestrator-run Aside steps enlarge the orchestrator context, bounded by text only; docs-site prose can drift from the skill with only grep guards).

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
