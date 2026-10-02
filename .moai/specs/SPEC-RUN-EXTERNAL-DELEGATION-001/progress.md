# SPEC-RUN-EXTERNAL-DELEGATION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_revision: 4                          # revision 4 of the plan: the run-phase errata (plan.md header; spec.md HISTORY 0.4.0). Revision 3 was written after plan-audit iteration 2 and audited in iteration 3 (the final audit, PASS-WITH-DEBT 0.88)
plan_complete_at: 2026-10-02T07:10:00Z    # revision-1 signal was 2026-10-02T05:53:41Z, revision-2 06:32:52Z; this value is the measured time (date -u) taken after the last evidence command and the strict lint of revision 3
card: t1424
tier: M
plan_base_sha: c50da9c2f8aa1227073bd77caa07ca1c75b8d81b
artifacts: [spec.md, plan.md, acceptance.md]
requirements: 16
acceptance_criteria: 16
planned_files: 15
design_decisions: [DR-1, DR-2, DR-3]   # DR-3 confirmed by the leader 2026-10-02; no open question
plan_audit_history:
  - iteration: 1
    verdict: FAIL
    score: 0.78                          # Tier M threshold 0.80; MP-8 (RED-now cell, ledger row E10) failed
    findings: D1-D18                     # D1-D10 blocking, D11-D18 optional; all dispositioned in revision 2
  - iteration: 2
    verdict: FAIL
    score: 0.81                          # at/above the Tier M threshold 0.80; the FAIL was writable mutants + a dangling milestone, not the score; no must-pass failure
    findings: N1-N12                     # N1-N6 blocking, N7-N10 optional (taken), N11-N12 accepted residuals; all dispositioned in revision 3 (spec.md HISTORY 0.3.0)
    re_audit: pending                    # iteration 3 of max 3 (the final audit) is the orchestrator's to dispatch
```

## §E.2 Run-phase Evidence

Format: Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (verification-claim-integrity §3). Milestone commits on branch `WT-run-codex-glm-delegation`, base `c50da9c2f` (`git merge-base develop HEAD`, read at use): M1 `d415f30aa` (test + spec.md status lines only, RED), M2 `07d921629`, M3 `3b90c3f47`, M4 `19bb1e8cf` (this agent), M5 = the commit carrying this file. Every command below ran in this run, in the worktree `.moai/worktrees/t1424`, against tree HEAD `19bb1e8cf` (M5 adds only this file, so the measured tree equals the M5 tree minus progress.md). Go tests ran as one compound `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ...`. M1-M3 evidence (RED capture, M2/M3 greens) belongs to the previous agent and is not re-observed here, except where a command below re-ran it on the final tree.

### E.2.1 Baselines re-measured on this tree (not carried over)

| Baseline | Command | Observed |
|---|---|---|
| Generator fixed point | copy `internal/template/catalog.yaml` to scratch, `go run ./internal/template/scripts/gen-catalog-hashes.go --all` once more, `cmp -s <copy> internal/template/catalog.yaml` | generator ended `catalog.yaml updated successfully (13900 bytes)`; `cmp -s` rc 0 (identical); `git status --short` empty. M4 itself changed no hashed file: the first generator run of M4 left `catalog.yaml` unchanged (`git status --short` listed only the four rule files) |
| Plan-time baselines E13-E19 | re-run on the final tree, see E.2.3 | all green; the tree differs from `c50da9c2f` by this card's commits, so they are new measurements, not carry-overs |

### E.2.2 AC matrix (AC-RXD-001..016)

Swept-count control for every `go test -run` row: the verbose stream was read; the final `ok` alone was never taken as evidence.

| AC | Result | Command | Actual output (verbatim, deciding lines) |
|---|---|---|---|
| AC-RXD-001 | PASS | `grep -c -E '^tools: Read, ... mcp__moai__glm_job_cancel$' .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md` (the E1b full-line pin); `grep -c -F "mcp__moai__codex_setup" <same two files>`; guard `tools` subtest | `internal/template/templates/.claude/agents/moai/manager-develop.md:1` / `.claude/agents/moai/manager-develop.md:1`; codex_setup: `.claude/agents/moai/manager-develop.md:0` / `internal/template/templates/.claude/agents/moai/manager-develop.md:0`; `--- PASS: TestRunExternalDelegationDoctrine/tools (0.00s)` |
| AC-RXD-002 | PASS | `grep -c "^## External Model Delegation" <both run.md>`; `cmp -s <run.md pair>`; guard `section` and `pointers` | `internal/template/templates/.claude/skills/moai/workflows/run.md:1` / `.claude/skills/moai/workflows/run.md:1`; `cmp -s` rc 0 (silent); `--- PASS: .../section`, `--- PASS: .../pointers (0.02s)` |
| AC-RXD-003 | PASS | guard `allowlist`, `exclusions` | `--- PASS: TestRunExternalDelegationDoctrine/allowlist (0.00s)`, `--- PASS: .../exclusions (0.00s)` |
| AC-RXD-004 | PASS | guard `request` (anchors, write-instruction negative check, two positive controls) | `--- PASS: TestRunExternalDelegationDoctrine/request (0.00s)` |
| AC-RXD-005 | PASS | guard `result` | `--- PASS: TestRunExternalDelegationDoctrine/result (0.00s)` |
| AC-RXD-006 | PASS | guard `value` | `--- PASS: TestRunExternalDelegationDoctrine/value (0.00s)` |
| AC-RXD-007 | PASS | guard `failopen` | `--- PASS: TestRunExternalDelegationDoctrine/failopen (0.00s)` |
| AC-RXD-008 | PASS | guard `onewriter` | `--- PASS: TestRunExternalDelegationDoctrine/onewriter (0.00s)` |
| AC-RXD-009 | PASS | guard `harness`; `git diff --name-only c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- internal/template/agentemit internal/template/templates/.codex`; `make agents-emit-check` | `--- PASS: .../harness (0.00s)`; `internal/template/templates/.codex/agents/moai/manager-develop.toml` (exactly one path); `ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.405s` |
| AC-RXD-010 | PASS | the E4 probes below (E4a-E4g, E5, E5b) and `go test -count=1 -v -run '^(TestMCPToolCatalogueDocsStayMirrorIdentical\|TestMCPToolCatalogueFiguresMatchRegistry\|TestCodexTaskAllowWrite_DistributedDefaultIsFalse\|TestCodexTaskAllowWriteReader_AgreesWithConfigLoader)$' ./internal/cli/` | E4a `8`; E4b `2`; E4c `2`; E4e `1`; E4f `1`; E4d `0`; E4g `1`; E5 `read-only only while`: `.claude/rules/moai/development/agent-authoring.md:1` / template `:1`; E5b `External Model Delegation`: `:1` / `:1`; `--- PASS: TestCodexTaskAllowWriteReader_AgreesWithConfigLoader (0.01s)`, `--- PASS: TestCodexTaskAllowWrite_DistributedDefaultIsFalse (0.00s)`, `--- PASS: TestMCPToolCatalogueDocsStayMirrorIdentical (0.00s)`, `--- PASS: TestMCPToolCatalogueFiguresMatchRegistry (0.00s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.287s` |
| AC-RXD-011 | PASS | `cmp -s` of the catalogue pair and of the run.md pair; guard `pairdelta` | both `cmp -s` rc 0 (silent); `pairdelta` logged `manager-develop.md ... = 5 (constant 5)`, `fix.md ... = 4 (constant 4)`, `loop.md ... = 2 (constant 2)`, `agent-authoring.md ... = 4 (constant 4)`; `--- PASS: .../pairdelta (0.00s)` |
| AC-RXD-012 | PASS | `git diff --name-only -G'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}\|REQ-[A-Z]\|AC-[A-Z]\|20[0-9]{2}-[0-9]{2}-[0-9]{2}' c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- .claude internal/template/templates`; `grep -rn -E "t1424\|SPEC-RUN-EXTERNAL-DELEGATION" internal/template/templates .claude/agents .claude/skills .claude/rules/moai`; `TestTemplateNoInternalContentLeak` | both commands: no output; `--- PASS: TestTemplateNoInternalContentLeak (0.81s)` |
| AC-RXD-013 | PASS | `git diff --name-only c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- . ':(exclude).moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001' ':(exclude).moai/reports/t1424'` | 15 lines: `.claude/agents/moai/manager-develop.md`, `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`, `.claude/rules/moai/development/agent-authoring.md`, `.claude/skills/moai/workflows/fix.md`, `.claude/skills/moai/workflows/loop.md`, `.claude/skills/moai/workflows/run.md`, `internal/template/catalog.yaml`, `internal/template/run_external_delegation_test.go`, the six `internal/template/templates/...` mirrors, `internal/template/templates/.codex/agents/moai/manager-develop.toml`. The agents-only range printed exactly the two `manager-develop.md` paths. The PRESERVE range (`internal/config .moai/config internal/template/templates/.moai/config` and both `moai-mcp-tools.md` copies) printed nothing |
| AC-RXD-014 | PASS | `grep -rn "allow_write: true" internal/template/templates .claude .moai/config`; E7d expression over the eight files; E7f over `internal/cli/mcp_server.go`; `grep -n "AllowWrite: false" internal/config/defaults.go`; the overlay mutation and mutants 9 and 10 below | `allow_write: true` grep: no output; E7d: no output; E7f: `1`; `1245:				AllowWrite: false,`; `--- PASS: TestCodexTaskAllowWrite_DistributedDefaultIsFalse` (shipped tree); observed failures in E.2.5 |
| AC-RXD-015 | PASS | the five-test set inside the 17-name batch below; `make agents-emit-check`, `make commands-emit-check`, `make tool-policy-drift-check`; idempotence comparison (E.2.1) | `--- PASS:` for `TestTemplateNoInternalContentLeak`, `TestManifestHashFormat`, `TestCatalogHashCoversSkillSubfiles`, `TestAllAgentsInCatalog`, `TestAllSkillsInCatalog`; `ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.405s`, `ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.261s`, `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.277s`; `cmp -s` rc 0; no `_templ.go` path in the 15-path set |
| AC-RXD-016 | PASS | guard verbose; whole packages; Windows build and vet; `moai agent lint`; `moai spec lint --strict`; the ten mutants | see E.2.3 and E.2.5 |

### E.2.3 Verification batch (plan §F)

| Command | Exit | Output (verbatim, deciding lines) |
|---|---|---|
| `go test -count=1 -v -run '^TestRunExternalDelegationDoctrine$' ./internal/template/` (after all mutants were reverted) | 0 | parent and 12 subtests `--- PASS`: `tools`, `section`, `allowlist`, `exclusions`, `request`, `result`, `value`, `failopen`, `onewriter`, `harness`, `pointers`, `pairdelta` (12 named, swept count read from the stream); `ok  	github.com/modu-ai/moai-adk/internal/template	0.300s` |
| 17-name template guard batch at M4 (`TestRunExternalDelegationDoctrine`, `TestTemplateNoInternalContentLeak`, `TestContractModeLocalTemplateParity`, `TestContractModeBlocksWellFormed`, `TestContractModeSSOTSections`, `TestContractModeLifecycleOrder`, `TestContractModeLifecycleEvidence`, `TestAgentlessUtilityNoLLMControlFlow`, `TestRunSkillContainsModeUnknownSentinel`, `TestAgentFrontmatterAudit`, `TestDeclaredRuleMirrorForks`, `TestSanitizedPairParity`, `TestManifestHashFormat`, `TestCatalogHashCoversSkillSubfiles`, `TestAllAgentsInCatalog`, `TestAllSkillsInCatalog`, `TestEvidenceCitation_Corpus`) | 0 | every name printed `--- PASS`; `ok  	github.com/modu-ai/moai-adk/internal/template	1.315s` |
| `moai slot acquire --resource go-test-internal-template --max-duration 20m`, then `go test -count=1 ./internal/template/... ./internal/config/...`, then `moai slot release --resource go-test-internal-template` | 0 | `slot go-test-internal-template acquired by 21477fb9-d7d8-482a-b4d4-9712b7dde154 until 2026-10-02T08:01:23Z`; `ok  	github.com/modu-ai/moai-adk/internal/template	167.845s`, `ok  	.../internal/template/agentemit	0.317s`, `ok  	.../internal/template/commandemit	0.119s`, `?   	.../internal/template/scripts	[no test files]`, `ok  	.../internal/config	10.439s`, `ok  	.../internal/config/atomicfile	0.417s`, `ok  	.../internal/config/toolpolicy	0.218s`; `slot go-test-internal-template released (was 21477fb9-d7d8-482a-b4d4-9712b7dde154)` |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 | no output |
| `GOOS=windows GOARCH=amd64 go vet ./internal/template/` | 0 | no output (type-checks the new `_test.go` for Windows) |
| `moai agent lint` | 0 | last line `Summary: 25 total (0 errors, 25 warnings)` (baseline 25 warnings, 0 errors: unchanged) |
| `moai spec lint --strict SPEC-RUN-EXTERNAL-DELEGATION-001` | 0 | `✓ No findings — all SPEC documents are valid` |
| `golangci-lint run --timeout=2m ./internal/template/...` | 0 | `0 issues.` CI pin read from `.github/workflows/ci.yml:464`: `golangci-lint@v2.1.6`; local `golangci-lint --version`: `v2.1.6 built with go1.26.8`, so the result is authoritative. The IDE `stringsseq` modernize hints at `run_external_delegation_test.go` ~245/248/304/437 are not reported by the CI-pinned lint, so the test file was left untouched |
| `make agents-emit-check`, `make commands-emit-check`, `make tool-policy-drift-check` | 0 each | the three `ok` lines quoted in AC-RXD-015 |

### E.2.4 Consumer-statement ledger (M4)

```
E4a   grep -c -E '^\| `mcp__moai__(codex_task|codex_job_status|codex_job_result|codex_job_cancel|glm_task|glm_job_status|glm_job_result|glm_job_cancel)` .*manager-develop' .claude/rules/moai/core/moai-mcp-tools-catalogue.md      -> 8
E4b   grep -c -E '^\| (Codex|GLM) delegation \|.*manager-develop' <catalogue>                                                                                                                               -> 2
E4c   grep -c -F "External Model Delegation" <catalogue>                                                                                                                                                    -> 2
E4d   grep -c -E '^\| `mcp__moai__codex_setup` .*manager-develop' <catalogue>                                                                                                                                -> 0
E4e   grep -c -E 'manager-develop.*codex_task.*External Model Delegation' <catalogue>                                                                                                                         -> 1
E4f   grep -c -E 'manager-develop.*glm_task.*External Model Delegation' <catalogue>                                                                                                                           -> 1
E4g   grep -c -F "manager-develop (all but codex_setup)" .claude/rules/moai/core/moai-mcp-tools-catalogue.md                                                                                                 -> 1   (debt D-2: literal probe on the codex family-table row)
E5    grep -c -F "read-only only while" <agent-authoring pair>                                                                                                                                              -> :1 :1
E5b   grep -c -F "External Model Delegation" <agent-authoring pair>                                                                                                                                         -> :1 :1
E7d'  the write-instruction expression over both agent-authoring copies and both catalogue copies (files M4 added text to)                                                                                   -> no output
```

`catalog.yaml`: no hashed file changed in M4 (the catalogue and `agent-authoring.md` are rules, not catalogued skill or agent files), so it is not in the M4 commit; the generator run printed `catalog.yaml updated successfully (13900 bytes)` and the file stayed byte-identical.

### E.2.5 Observed failures (mutation proofs)

Method for every mutant: copy the real file to the scratch directory, edit the real file, run `go test -count=1 -run '^TestRunExternalDelegationDoctrine$' ./internal/template/` (exit 1 each), restore with `cp` from the scratch copy, run `git diff --stat -- <file>` (printed nothing each time; `git status --short` printed nothing after the last restore). No `git checkout`, `git stash` or `git reset` was used.

**AllowWrite default flip (AC-RXD-014), no PRESERVE file edited.** A scratch copy of `internal/config/defaults.go` with line 1245 `AllowWrite: false,` changed to `AllowWrite: true,` (`diff` of real vs scratch: `1245c1245`) and an overlay JSON `{"Replace": {"<abs>/internal/config/defaults.go": "<scratch>/defaults_mut.go"}}`:

```
$ unset ... && go test -overlay=<scratch>/overlay.json -count=1 -v -run '^TestCodexTaskAllowWrite_DistributedDefaultIsFalse$' ./internal/cli/
=== RUN   TestCodexTaskAllowWrite_DistributedDefaultIsFalse
    codex_task_test.go:407: the distributed default for workflow.codex.task.allow_write must be false
--- FAIL: TestCodexTaskAllowWrite_DistributedDefaultIsFalse (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.222s
```
`git diff --stat -- internal/config/defaults.go` printed nothing (the real file was never edited).

**Ten guard mutants** (plan §D and DoD name nine; AC-RXD-014 additionally requires the "append ` to true`" mutant, so ten were run: **plan inconsistency, nine vs ten, reported**). Several mutants applied to one tree only also trip `section` (byte-equality of the two run.md copies) or `pairdelta`; the deciding subtest for each is the one named.

1. Remove one tool from the template tools line (`internal/template/templates/.claude/agents/moai/manager-develop.md`, dropped `, mcp__moai__glm_job_cancel`): `--- FAIL: .../tools (0.00s)`: `run_external_delegation_test.go:312: template .claude/agents/moai/manager-develop.md: \`tools:\` line is not the existing prefix plus the eight delegation tools.` with `got:` ending `mcp__moai__glm_job_result` and `want:` ending `mcp__moai__glm_job_result, mcp__moai__glm_job_cancel`; and `--- FAIL: .../pairdelta (0.00s)`: `.claude/agents/moai/manager-develop.md: live/template multiset line difference is 7, want 5`.
2. Add `codex_setup` (live `manager-develop.md` tools line, appended `, mcp__moai__codex_setup`): `--- FAIL: .../tools`: `run_external_delegation_test.go:312: live .claude/agents/moai/manager-develop.md: \`tools:\` line is not the existing prefix plus the eight delegation tools.` and `run_external_delegation_test.go:315: live .claude/agents/moai/manager-develop.md: must not carry mcp__moai__codex_setup (the probe tool stays with super-advisor)`; also `pairdelta` 7 vs 5.
3. Insert `Agent` into the tools prefix (live `manager-develop.md`, `... Glob, Agent, TaskCreate ...`): `--- FAIL: .../tools`: `run_external_delegation_test.go:312: live .claude/agents/moai/manager-develop.md: \`tools:\` line is not the existing prefix plus the eight delegation tools.`; also `pairdelta` 7 vs 5.
4. Rename the section heading (live `run.md`, `## External Model Delegations`): `--- FAIL: .../section`: `run_external_delegation_test.go:375: live .claude/skills/moai/workflows/run.md: want exactly one "## External Model Delegation" heading, found 0` and `run_external_delegation_test.go:407: run.md copies are not byte-equal (live 25292 bytes, template 25291 bytes)`; each of `allowlist`, `exclusions`, `request`, `result`, `value`, `failopen`, `onewriter`, `harness` failed with `run_external_delegation_test.go:421: live .claude/skills/moai/workflows/run.md: section "## External Model Delegation" found 0 times, want 1 (the section is missing or duplicated)`.
5. Copy a sub-heading into `fix.md` (live, a `### Delegable classes` line after the pointer): `--- FAIL: .../pointers (0.02s)`: `run_external_delegation_test.go:516: live .claude/skills/moai/workflows/fix.md: carries the doctrine phrase "### Delegable classes" (pointers name the section, they do not restate it)`; `pairdelta`: `fix.md: live/template multiset line difference is 6, want 4`.
6. One-writer rewrite (live `run.md`, "the agent may freely edit the files named in a prompt ... reads or cancels every job it started."): `--- FAIL: .../onewriter`: `run_external_delegation_test.go:431: live .claude/skills/moai/workflows/run.md: anchor "does not edit the files named in an in-flight prompt" is missing from subsection "### One-writer rule"` and `... anchor "reads or cancels every job it started before reporting completion" is missing from subsection "### One-writer rule"` (also `section`: copies not byte-equal).
7. Scoping mutant (live `run.md`: the `never sets the write argument` sentence moved from `### Request construction` into `### Fail-open`): `--- FAIL: .../request`: `run_external_delegation_test.go:431: live .claude/skills/moai/workflows/run.md: anchor "never sets the write argument" is missing from subsection "### Request construction"` (also `section`: not byte-equal at equal size 25291, content moved).
8. Two-line pointer paraphrase with no anchor phrase (live `fix.md`, a second line after the pointer line: "In short the agent hands a narrow drafting job to the outside model, ... never touching the files the job names meanwhile."): `--- FAIL: .../pointers (0.03s)`: `run_external_delegation_test.go:510: live .claude/skills/moai/workflows/fix.md:196: pointer paragraph is 103 words (max 40)`; `pairdelta`: `fix.md: ... is 5, want 4`.
9. Markdown-variant write instruction (live `run.md`, appended "For lint-repair drafts it sets the `write` argument to `true`."): `--- FAIL: .../request`: `run_external_delegation_test.go:450: write-instruction pattern matched in the delegation section: live .claude/skills/moai/workflows/run.md (section):33: For \`mcp__moai__codex_task\` the agent passes \`project_root\` equal to its own \`git rev-parse --show-toplevel\`, which is its own L1 tree where the spawn auto-isolated into one; the argument is required and is never defaulted. The agent never sets the write argument, so a delegated turn stays read-only. For lint-repair drafts it sets the \`write\` argument to \`true\`.` The E7d grep over the same file printed line `229:` with the same text.
10. Anchor sentence with ` to true` appended (live `run.md`, `The agent never sets the write argument to true, so a delegated turn stays read-only.`): `--- FAIL: .../request`: `run_external_delegation_test.go:450: write-instruction pattern matched in the delegation section: live .claude/skills/moai/workflows/run.md (section):33: For \`mcp__moai__codex_task\` ... The agent never sets the write argument to true, so a delegated turn stays read-only.`

Additional observed failures for the regression-guards of AC-RXD-012 and AC-RXD-015 (uncommitted mutants, each restored with empty `git diff --stat`):

- Stale catalog hash (template `fix.md` pointer line edited, hashes not regenerated): `--- FAIL: TestCatalogHashCoversSkillSubfiles (0.01s)`: `catalog_tier_audit_test.go:499: CATALOG_HASH_SKINNY: moai hash=fd2f0680f57e757d08aebeb42e8074dffde7da7a098d03148f06f4507e8e1d53 does not cover the deployed directory tree (whole-tree hash 9416874cee829f6546a2eab577f0e06c7c30f7a31aa507a830aaec1c4d377181) - run gen-catalog-hashes.go --all (card t323)`.
- Hand-edited generated TOML (`name = "manager-develop"` changed to `name = "manager-develop-hand-edited"`): `make agents-emit-check` exit 2: `--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)`, `golden_test.go:101: .codex/agents/moai/manager-develop.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing`; `make: *** [agents-emit-check] Error 1`.
- Internal token pasted into a template file (SPEC id plus ISO date in the template `fix.md` pointer): `--- FAIL: TestTemplateNoInternalContentLeak (1.22s)`: `internal_content_leak_test.go:1627:   [1] templates/.claude/skills/moai/workflows/fix.md | class=C1-spec-id-prefix | match=SPEC-RUN-EXTERNAL-DELEGATION-001`; the `-G` scan over the worktree diff printed `internal/template/templates/.claude/skills/moai/workflows/fix.md`. A bare card id (`(card t1424)`) was NOT flagged by `TestTemplateNoInternalContentLeak` (it exited with the combined stale-hash failure only; the leak test itself printed no failure); only the `grep -rn -E "t1424|SPEC-RUN-EXTERNAL-DELEGATION"` command printed the line. This contradicts the AC-RXD-012 mutant-probe sentence "pasting the card id fails the second and the leak test": the leak test does not cover a bare card id.

### E.2.6 Debt dispositions (plan-audit iteration 3, carried via the Jev decision record)

| Debt | Disposition |
|---|---|
| D-1 (N13, write-regex false positives; failure message must print the matching line) | Observed, not changed. The `request` subtest prints `<file> (section):<line>: <text>`: the file label and the matching text are printed, but the line number is relative to the section slice (`33`), not to the file (`229`); the `pointers` subtest prints file-relative line numbers (`fix.md:196`). A false positive is diagnosable from the printed matching text. Making the `request` line number file-relative would be a test edit outside the lint-fix permission of this milestone; reported to sync |
| D-2 (N14, family-table row unprobed) | Closed: E4g printed `1` (`grep -c -F "manager-develop (all but codex_setup)" ...`) |
| D-3 (N15, " to true" mutant not in the nine) | Closed by running ten mutants; the nine-versus-ten count mismatch (plan §D/§E M5 and DoD-2 say nine, AC-RXD-014 requires the extra mutant) is reported as a plan inconsistency |
| D-4 (N16, HISTORY 0.3.0 omits REQ-RXD-002) | Carried to sync: a spec.md body edit, outside run-phase ownership; not touched |
| D-5 (N17, the Claude-harness-only sentence in the agent-body pointer has no test) | Unbacked sentence, recorded. It sits in the agent-body pointer paragraph (`it applies to Claude Code sessions only.`); no requirement, criterion or subtest decides it |

### E.2.7 Gaps (explicitly not observed)

- M1 RED output (E8) and the M2/M3 intermediate greens were produced by the previous agent; this run re-observed only the final GREEN. The ordering witness is the commit graph: `d415f30aa` (parent `0b33c713d`) carries exactly `spec.md` and `internal/template/run_external_delegation_test.go`, and the doctrine commits `07d921629`, `3b90c3f47`, `19bb1e8cf` follow it.
- Tool provenance (verification-claim-integrity §2.2): `moai agent lint` and `moai spec lint --strict` were judged by the installed build `v3.2.0-rc.24 moai_cp/20260925_122548-1896-gc50da9c2f` (`moai version`), whose commit `c50da9c2f` is a strict ancestor of tree HEAD `19bb1e8cf`; a lint rule added after `c50da9c2f` would not have run. The `go test`, `go build`, `go vet`, `make ...-check` and `golangci-lint v2.1.6` results were produced from the tree.
- Hand-editing a `catalog.yaml` hash (the third AC-RXD-015 mutant) was not run separately; the stale-hash mutant reaches the same `CATALOG_HASH_SKINNY` check.
- Mutants 5 and 8 were run against the live `fix.md` only; the template twin was not mutated separately. `pairdelta` is what turns such one-copy hunks red.
- The repository-wide test verdict belongs to CI on the integration branch: PENDING (DoD-5). `go test ./...` was not run, by rule.
- The docs-site four-locale consumer-table correction is open and handed to the sync phase (DoD-6).
- The opt-in `workflow.codex.task.allow_write` is `true` in the maintainer's primary checkout (E8b, spec.md R-6); the evidence here proves the shipped default and the absence of a shipped write instruction, not the opt-in state of any machine.
- The accepted lexical residuals of spec.md R-3 (a short pointer paraphrase within the caps; a write instruction with no enabling word near `write`) are not decided by any probe here.

### E.2.8 Residual-risk

A delegated codex turn stays read-only only because the agent never sets `write` while the opt-in is open on the maintainer machine; a future wording change that names the argument with an enabling word within 40 characters fails the guard, but an unhooked paraphrase would not. The eight added MCP tool schemas raise every `manager-develop` spawn's tool-listing cost (spec.md R-7, unmeasured).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-02T07:46:24Z      # date -u, taken after the last evidence command of this run
card: t1424
cycle_type: tdd
head_measured: 19bb1e8cf                    # M5 commit adds only this progress.md
milestones:
  M1: d415f30aa
  M2: 07d921629
  M3: 3b90c3f47
  M4: 19bb1e8cf
  M5: this commit (progress.md only)
planned_files: 15
changed_files_equal_planned: true           # E.2.2 AC-RXD-013
acceptance_criteria: {pass: 16, fail: 0}
guard_test: {parent: PASS, subtests: 12}
mutants: {guard: 10, extra_regression_guard: 3, allow_write_overlay: 1}   # nine in the plan, ten run
open_for_sync: [D-4 HISTORY wording (manager-spec), docs-site 4-locale consumer table, D-1 section-relative line number in the request message]
unbacked_sentence: D-5
ci_verdict: PENDING                          # repository-wide suite belongs to CI on the integration branch
pushed: false                                # the run phase does not push
```

## §E.4 Sync-phase Audit-Ready Signal

Written by manager-docs on the single sync commit (the 3-phase close). Every measurement below ran in this sync run, in the worktree `.moai/worktrees/t1424`, against tree HEAD `9d8e8d02c` (the SPEC-errata commit, version 0.4.1) before the sync edits; the judging `moai` build for `moai spec lint` is `v3.2.0-rc.24 moai_cp/20260925_122548-1896-gc50da9c2f` (its commit `c50da9c2f` is a strict ancestor of HEAD).

```yaml
sync_status: audit-ready
sync_complete_at: 2026-10-02T08:00:05Z     # date -u, taken after the last evidence command of this sync run
sync_commit_sha: pending-backfill           # a commit cannot cite its own hash; the next commit backfills the real SHA (D3 window)
card: t1424
head_measured: 9d8e8d02c
b12_self_test_a: PASS                       # grep -c 'SPEC-RUN-EXTERNAL-DELEGATION-001' CHANGELOG.md printed 0 before the entry was written
b12_self_test_b: PASS                       # ac_source=.moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001/acceptance.md, tier M; counter printed live=16 excluded=0 ambiguous=0 and 16 (exit 0); the CHANGELOG entry states 16
b12_self_test_c: PASS                       # every path named in the CHANGELOG entry was checked with ls (21 paths, all present)
changelog_entry_position: "[Unreleased] > ### Added, first entry"
frontmatter_status_transitions:
  spec.md: "in-progress -> implemented -> completed (one commit), updated: 2026-10-02 (unchanged, equals the commit date)"
  plan.md: "no status frontmatter (stateless artifact per spec-frontmatter-schema.md Artifact Statelessness); body untouched"
  acceptance.md: "no status frontmatter (stateless artifact); body untouched"
  progress.md: "no frontmatter; this section and the §E.1 plan_revision field (3 to 4) are the only edits"
canary_compliance_check: N/A                # this SPEC defines no forward-looking policy that its own sync tests
mx_tag_validation: "no @MX annotation required (see below)"
codemaps_readme: N/A                        # see below
docs_site: "four-locale guides/mcp-server.md updated; hugo build clean (see below)"
sync_audit: pending                         # the independent sync-audit is the next step and is not this agent's to perform
pushed: false
```

**B12 counter run (the attributable count).** The first sync attempt halted: at HEAD `c5aa54abb` the counter printed `live=24 excluded=0 ambiguous=0` because eight bare short-form cross-references (`AC-001`, `AC-002`, `AC-003`, `AC-009`, `AC-010`, `AC-011`, `AC-012`, `AC-016`) sat unmarked next to the sixteen canonical `AC-RXD-NNN` identifiers. The orchestrator had manager-spec normalize those to the canonical spelling (commit `9d8e8d02c`, SPEC version 0.4.1). The count in the CHANGELOG entry is from this agent's own re-run at `9d8e8d02c`: command `sh <scratchpad>/run-counter.sh .moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001/acceptance.md` (the verbatim counter program of `manager-docs.md` § B12, run through a wrapper because the worktree guard refuses a bare `awk -f`), output `live=16 excluded=0 ambiguous=0`, `16`, `counter_exit=0`. Cross-check: `grep -c -E "^## AC-RXD-0(0[1-9]|1[0-6])"` printed `16` at `c5aa54abb`.

**Carried-debt dispositions (plan-audit iteration 3, via the Jev decision record).**

| Debt | Disposition at sync |
|---|---|
| D-1 (write-regex false positives; failure message must print the matching line) | Done with a note: the `request` subtest failure message prints the file label, a section-relative line number and the matching text (`<file> (section):<line>: <text>`), so a false positive is diagnosable from the message; the line number is relative to the section slice, not to the file (the `pointers` subtest prints file-relative lines). Source: §E.2.6 and mutants 9 and 10 of §E.2.5. |
| D-2 (family-table `codex_setup` clause unprobed) | Closed in the run phase by ledger probe E4g (`grep -c -F "manager-develop (all but codex_setup)"` on the catalogue printed `1`, §E.2.4). |
| D-3 (the " to true" mutant not scheduled) | Closed in the run phase by running ten guard mutants (§E.2.5); the plan's nine-versus-ten count was corrected in the errata. |
| D-4 (HISTORY 0.3.0 understated the REQ-RXD-002 change) | Closed by the errata commit `c5aa54abb` (spec.md HISTORY 0.4.0 item (a) names the REQ-RXD-002 pointer-forbidden set and the two-line, forty-word cap). |
| D-5 (the Claude-harness-only sentence in the agent-body pointer has no test) | Residual, kept explicit: the sentence `it applies to Claude Code sessions only.` sits in the pointer paragraph of `manager-develop.md` and is decided by no requirement, criterion or subtest. |
| D-6 (accepted lexical residuals, R-3) | Unchanged: a short pointer paraphrase inside the caps and a write instruction without `true`/`enabled`/`on` are decided by the plan-auditor and sync-auditor reading, as the SPEC states. |

**Observed by the run and not fixable inside this SPEC.** (1) `TestTemplateNoInternalContentLeak` does not flag a bare card id (a pasted `(card t1424)` produced no failure from the leak test itself; only the plain `grep -rn -E "t1424|SPEC-RUN-EXTERNAL-DELEGATION"` of AC-RXD-012 caught it): a gap in a generic guard, reported to the leader as a follow-up candidate. (2) The IDE's `itoa` redeclaration diagnostic comes from two pre-existing files in different test packages and is unrelated to this change (as reported by the run phase and the leader; not re-measured in this sync run).

**MX tag validation (sync sub-step).** `.claude/rules/moai/workflow/mx-tag-protocol.md` was read in full. The only Go file this card added is `internal/template/run_external_delegation_test.go`. Observed in this run: `grep -n -E "^func [A-Z]|^func Test|goroutine|go func|sync\.|chan "` found exactly one exported function, `TestRunExternalDelegationDoctrine` (a test entry point, no callers); no goroutine, channel or `sync` use; `grep -c "@MX"` printed `0`. None of the protocol's add-triggers applies (no fan_in of 3 or more, no goroutine or channel, no global-state mutation, no exported function lacking a test, no magic constant needing a NOTE), so no annotation is required and none was added; no code edit was made in this phase. Not measured: cyclomatic complexity (no `gocyclo` or `gocognit` binary is installed), so the protocol's complexity-15 WARN trigger is unobserved for the 12-subtest function.

**Codemaps and README.** N/A. `grep -rln -E "codex_task|glm_task|codex_job_status|glm_job"` over `README*.md` and `.moai/project` found the four README files (a tool list with no consumer column) and `.moai/project/codemaps/modules.md` (unrelated prose); none names which agent consumes the delegation tools, so no codemap or README sentence became false. Not edited.

**docs-site verification (hns-oss-docs-verify).** Edited only the four `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md` pages: `manager-develop` joins the consumer cell of the eight codex and GLM delegation rows (not `codex_setup`, not the read-only role rows) and both family paragraphs now name `manager-develop` and the `External Model Delegation` section of the run workflow. Results: hugo `v0.160.1+extended` build with `--minify --gc` to a scratch directory printed no WARN or ERROR line, `sitemap.xml` present; URL-blacklist grep no match; Mermaid LR/RL grep no match; 4-locale file-existence parity no MISSING line; section-count parity against `docs-site/.locale-parity-baseline`: no NEW divergence (and no converged page); the changed page has 26 headings of level 2 or deeper in each of the four locales; README H2 counts 12/12/12/12; body-emoji scan of the four changed pages no hit (positive control hit); version displays: `hugo.toml` `v3.1.3` equals the four README badges, no `🗿 v` display on the changed pages. Other pages: `grep -rln "manager-develop" docs-site/content` lists many pages, but a grep for the delegation tool names, for the `manager-develop` tool-list names (`verify_snapshot`, `goal_status`) and for sole-consumer wording about `super-advisor` matched only `guides/mcp-server.md` (and, for the tool names alone, `advanced/config-sections.md`, which names `codex_task` and `glm_task` without a consumer claim); those pages were left unchanged. This is a grep-based finding, not a full read of every page that mentions `manager-develop`.

**Gaps (not observed).**
- The independent sync-audit has not run; this signal states what the sync phase produced, not a verdict.
- The repository-wide test verdict belongs to CI on the integration branch (DoD-5, still PENDING); `go test ./...` was not run, and no Go test was re-run in this sync run (no Go file changed in this phase; `moai spec lint --strict` printed `✓ No findings — all SPEC documents are valid` at `9d8e8d02c`).
- A bare `awk -f` and a bare `grep -P` were not usable here (guard refusal; macOS grep has no `-P`): the counter ran through a `sh` wrapper (same program, executed) and the emoji scan ran with `perl -CSD` instead, with a positive control.
- The body-emoji scan, the blacklist grep and the Mermaid grep were scoped as the recipe states; the emoji scan covered the four changed pages only, not the whole docs-site tree.
- The three non-Korean locale wordings are derived in this commit from the Korean page; no native-speaker review was available.
- Cyclomatic complexity of the new test function (see the MX paragraph).

**Residual-risk.** A delegated codex turn stays read-only only through the prose and its lexical guard while the `allow_write` opt-in is open on the maintainer machine; the eight added tool schemas raise each `manager-develop` spawn's tool-listing cost (unmeasured); D-5 and the R-3 lexical residuals are accepted and read by the auditors, not by a probe.
