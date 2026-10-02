# acceptance — SPEC-ASIDE-BROWSER-001

> Verification layer (Given-When-Then). Requirements are the GEARS `REQ-ASB-*` entries in `spec.md` § 2. Measurement tree for every RED-now cell: `4bf547bcad7c155b1e91485921569db709ec3ac2` (document-level pin; binds every criterion without its own pin). Every check below is binary; a check that selects zero tests is not a pass (`go test` prints `[no tests to run]` and exits 0 in that case, measured as E1b), so each check names the `--- PASS:` line it requires.

## §D AC Matrix

| AC | Axis | Covers | One-line verdict |
|----|------|--------|------------------|
| AC-ASB-001 | Default config | REQ-ASB-002 | `aside` absent from default `.mcp.json` and settings templates; the guard bites on constructed failing inputs |
| AC-ASB-002 | Documented command | REQ-ASB-001, REQ-ASB-003 | The documented line works in both scopes, is idempotent, and is pinned by a test |
| AC-ASB-003 | Skill + core catalog | REQ-ASB-004, REQ-ASB-014 | Skill exists, non-invocable, within the listing cap, catalogued in `core`, core-tier guards pass |
| AC-ASB-004 | Permission | REQ-ASB-005 | No full-access, Guard by omission, repl limited to read operations |
| AC-ASB-005 | Writes | REQ-ASB-006 | State-changing actions need orchestrator-obtained operator confirmation |
| AC-ASB-006 | Subagents | REQ-ASB-007 | No subagent invokes Aside; the sentence is in the skill and in the e2e-tester definition |
| AC-ASB-007 | Install and secrets | REQ-ASB-008 | Never auto-install (advise `aside skills install` only); no credentials in outputs |
| AC-ASB-008 | e2e gating | REQ-ASB-009, REQ-ASB-011 | Explicit-only; excluded when `CI=true` |
| AC-ASB-009 | Orchestrator execution | REQ-ASB-010 | Orchestrator runs every Aside step; owner and bounded-output carve-outs; skill loaded first |
| AC-ASB-010 | Silent fallback | REQ-ASB-012 | No Aside-specific output on absence or exclusion; carve-out at both missing-toolchain sites |
| AC-ASB-011 | Evidence | REQ-ASB-013 | Screenshot captured through repl, saved under `e2e/`, cited by path |
| AC-ASB-012 | Parity + build | REQ-ASB-014 | Mirrors identical, hash set exact, Codex TOML regenerated, build targets exit 0 |
| AC-ASB-013 | Neutrality | REQ-ASB-015 | No ids, SHAs, or dates in any authored template file; template package green |
| AC-ASB-014 | Docs-site | REQ-ASB-001 | Command and numeral correct in four locales; hugo build clean |

All fourteen are classified release-blocking except where a cell says "regression-guard". RED-now cells use a single read-only invocation; Check cells may be compound.

## Checker anchors (the literals `TestAsideSkillPolicyAnchors` pins)

The test calls pure checkers (`checkAsideSkill`, `checkAsideE2E`) on the real files and then on negative controls: the same text with one anchor removed, or one bad line added. Every control must produce its named violation; a control that passes the checker is a defect in the checker. Anchors are literal substrings, so two implementers cannot disagree about them.

| Subtest name | Text | Required literal(s) | Forbidden |
|--------------|------|---------------------|-----------|
| `description_within_listing_cap` | skill frontmatter | `description` plus `when_to_use` total at most 1536 characters | more than 1536 |
| `full_access_prohibited` | skill | `Never pass `--permission full-access`` | any other line containing `full-access` unless it also contains `never request full-access` (a closed two-literal allow-list, so `never forget to run aside --permission full-access` fails) |
| `guard_by_omission` | skill | ``omit `--permission` `` and `Guard` | none |
| `repl_read_only_limit` | skill | `aside repl`, `no permission flag`, `navigation, reading, and screenshot` | none |
| `write_confirmation_via_orchestrator` | skill | `explicit operator confirmation`, `question channel` | none |
| `orchestrator_only_operation` | skill | `operated by the orchestrator alone` | none |
| `subagent_never_invokes_skill` | skill | `never invokes Aside` | none |
| `subagent_never_invokes_tester` | e2e-tester definition | `never invokes Aside` | none |
| `no_auto_install_advise_only` | skill | ``advise the operator to run `aside skills install` ``; every line containing `aside skills install` also contains `operator` | any line mentioning Aside together with `npm`, `npx`, `curl`, `brew`, `cargo`, `pip`, or `go install` |
| `no_credentials_in_outputs` | skill | `never place credentials, cookies, tokens, or session data` | none |
| `e2e_explicit_only` | e2e.md | ``only when `--tool aside` is passed explicitly``, `never auto-detected`, `never offered`, `never recommended` | none |
| `e2e_ci_excluded` | e2e.md | one line containing `CI=true`, `Aside`, and `unavailable` | none |
| `e2e_orchestrator_executes_aside` | e2e.md | `the ORCHESTRATOR runs every Aside step`, ``Skill("moai-ref-aside-browser")`` | none |
| `e2e_execution_owner_carveout` | e2e.md | the `Execution owner` line contains `except Aside` | none |
| `e2e_aside_output_bounded` | e2e.md | the `Bounded output` line contains `Aside` and `e2e/.runs/` | none |
| `e2e_silent_fallback_no_aside_message` | e2e.md | `silently`, `no Aside-specific message` | any line mentioning Aside that contains `warn`, `fallback note`, `notify`, or `report that Aside` |
| `e2e_missing_toolchain_carveout` | e2e.md | the `Missing-toolchain sequence` header line contains `except Aside` | none |
| `e2e_tool_bypass_line_carveout` | e2e.md | the line containing `missing-toolchain sequence if absent` also contains `except Aside` | none |
| `e2e_no_install_command` | e2e.md | none | any line mentioning Aside together with the install verbs listed above |
| `e2e_repl_screenshot_evidence` | e2e.md | ``screenshot captured through `aside repl` ``, `e2e/` | none |

Stated limit: the silence check is two-sided but lexical. It proves the text instructs silence and instructs no note; it cannot prove what a model emits at run time (residual risk, `research.md` § 3).

---

### AC-ASB-001 — `aside` absent from defaults (maps REQ-ASB-002)

- **Given** the tree after M1
- **When** `go test ./internal/template/ -run '^(TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1`
- **Then** the output contains `--- PASS: TestMCPDefaultExcludesAside ` and `--- PASS: TestMCPNeutralityTemplateShape ` (each name followed by a space) and does not contain `[no tests to run]`, exit 0. Cross-check (regression-guard, green today by design): `grep -c -i aside internal/template/templates/.mcp.json internal/template/templates/.claude/settings.json.tmpl` reports `0` for both.
- **Why a named test when `TestMCPNeutralityTemplateShape` exists**: that test already fails on any key outside `moai` and `context7`, so the `.mcp.json` half of the new test is deliberately redundant — it makes the safety constraint greppable by name and gives a failure message that says Aside. The non-duplicate half is the settings-template token check; the mutant below separates the two.
- **RED-now** (E1): `grep -c 'func TestMCPDefaultExcludesAside' internal/template/mcp_template_neutrality_test.go` → `0`, exit 1. Red because the named guard does not exist; the invariant itself is green today, so the criterion is carried by the mutants.
- **Mutant probes**: (a) add an `aside` key to the template `.mcp.json`: both named tests print `--- FAIL`; (b) add the token `aside` to a comment-free line of `settings.json.tmpl`: only `TestMCPDefaultExcludesAside` prints `--- FAIL` (this proves it is not a duplicate); revert each, then a plain `git diff --stat` shows neither file changed.
- **Green path**: M1.1.

### AC-ASB-002 — documented command works and is pinned (maps REQ-ASB-001, REQ-ASB-003)

- **Given** the tree after M1 and M2
- **When** `go test ./internal/cli/ -run '^TestMCP_Add_AsideDocumentedCommandLine$' -v -count=1`, and `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md`
- **Then** the test prints `--- PASS: TestMCP_Add_AsideDocumentedCommandLine/project ` and `--- PASS: TestMCP_Add_AsideDocumentedCommandLine/user ` (one `aside` entry, `args` `["mcp"]`, file byte-identical after the second run, unrelated entries preserved); the grep prints `1` or more.
- **RED-now** (E2): `grep -c 'TestMCP_Add_AsideDocumentedCommandLine' internal/cli/mcp_test.go` → `0`, exit 1. The skill-side check is red too: E3 shows the skill file does not exist.
- **Mutant probe**: misspell `--args` as `--arg` in the test invocation; expect `unknown flag`; revert.
- **Plan-time observation (non-gating, project scope only)**: the documented line without `--scope user`, run twice in a scratch project against a binary built from this tree (`go build` exit 0), exited 0 both times; `.mcp.json` held exactly `{"mcpServers":{"aside":{"args":["mcp"],"command":"aside"}}}` and `moai mcp list --json` reported `count: 1`. The plan-audit re-ran this and reproduced it. User scope was not run: setting `HOME` is refused by the worktree guard, so the user-scope subtest goes through the `userHomeDirFn` seam.
- **Green path**: M1.2 (test) and M2.2 (skill line).

### AC-ASB-003 — skill present, non-invocable, core-catalogued (maps REQ-ASB-004, REQ-ASB-014)

- **Given** the tree after M2
- **When** `go test ./internal/template/ -run '^(TestAllSkillsInCatalog|TestCatalogHashCoversSkillSubfiles|TestManifestHashFormat|TestEmbeddedMoaiSkillNames|TestCatalogTierValid|TestCatalogNoDuplicateEntries|TestCatalogReferencesValid|TestWorkflowTriggerCoverage|TestSlimFS_PreservesCoreEntries|TestSlimFS_HidesNonCoreEntries|TestPublishedSkillsNamesMatchTree|TestSkillMirror_SetIsDerivedNotConstant|TestGitignore_IgnoresSkillMirrorOnly)$' -v -count=1`; `grep -n 'user-invocable: false' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md`; `grep -n -A3 'name: moai-ref-aside-browser' internal/template/catalog.yaml`; and `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** every named guard prints `--- PASS` (none selected zero); the skill carries `user-invocable: false`; the catalog entry sits in the `core` tier with `tier: core` and a 64-hex `hash:`; and `--- PASS: TestAsideSkillPolicyAnchors/description_within_listing_cap ` is printed (the cap is measured, not asserted).
- **Core-tier consequences, measured at baseline (E11)**: the 12 selectors that read the core tier or slim/published surfaces all printed `--- PASS` on the unmodified tree, exit 0. None hard-codes a core or total count: the totals in `catalog_loader_test.go` and `embed_catalog_test.go` are derived from `catalog.yaml`'s own `- name:` lines, and the slim core test's docstring mentions 40 while its body asserts no count. `TestSlimFS_PreservesCoreEntries` now also covers the new entry (the skill reaches slim installs, which is the reason for core). `TestSlimFS_HidesNonCoreEntries` is unaffected because the entry is core.
- **RED-now** (E3): `ls internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md` → `No such file or directory`, exit 1. (E6): `grep -c 'moai-ref-aside-browser' internal/template/catalog.yaml` → `0`, exit 1.
- **Green path**: M2.2, M2.3.

### AC-ASB-004 — permission prohibitions (maps REQ-ASB-005)

- **Given** the skill after M2
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/full_access_prohibited `, `.../guard_by_omission `, `.../repl_read_only_limit ` print (anchors: § Checker anchors). Reader check: `grep -n 'full-access' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md` prints only lines that contain ``Never pass `--permission full-access` `` or `never request full-access`.
- **RED-now** (E8): `grep -c 'func TestAsideSkillPolicyAnchors' internal/template/aside_skill_policy_test.go` → `grep: ... No such file or directory`, exit 2.
- **Mutant probes** (negative controls inside the test): the skill text with the prohibition removed; with an unqualified `aside --permission full-access` example appended; and with the line `never forget to run aside --permission full-access` appended — each must produce the `full_access_prohibited` violation.
- **Green path**: M2.1 then M2.2.

### AC-ASB-005 — write confirmation (maps REQ-ASB-006)

- **Given** the skill after M2
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/write_confirmation_via_orchestrator ` and `.../orchestrator_only_operation ` print.
- **RED-now**: E8.
- **Mutant probe**: remove the `explicit operator confirmation` sentence; remove `operated by the orchestrator alone`; each named subtest must fail.
- **Green path**: M2.1 then M2.2.

### AC-ASB-006 — no subagent invokes Aside (maps REQ-ASB-007)

- **Given** the skill and the e2e-tester definition after M2 and M3
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, and `grep -c 'never invokes Aside' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md internal/template/templates/.claude/agents/moai/e2e-tester.md .claude/agents/moai/e2e-tester.md`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/subagent_never_invokes_skill ` and `.../subagent_never_invokes_tester ` print; the grep reports `1` or more for each of the three files.
- **RED-now** (E5, E12): `grep -c 'never invokes Aside' internal/template/templates/.claude/agents/moai/e2e-tester.md` → `0`, exit 1.
- **Mutant probe**: delete the sentence from the e2e-tester text, and from the skill text; each named subtest must fail. (This is the strongest cheap mechanical anchor: the subagent reads its own prohibition, and deleting it turns a test red. No hook enforces the boundary; see residual risk.)
- **Green path**: M2.1/M2.2 (skill) and M3.3 (e2e-tester).

### AC-ASB-007 — no auto-install, no secrets in outputs (maps REQ-ASB-008)

- **Given** the skill after M2
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, and `grep -c 'aside skills install' internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/no_auto_install_advise_only ` and `.../no_credentials_in_outputs ` print; the grep prints `1` or more. "Advise-the-operator framing" is exactly the literal pair in § Checker anchors: the phrase ``advise the operator to run `aside skills install` `` plus the word `operator` on every line naming that command.
- **RED-now**: E8 and E3.
- **Mutant probe**: append `run: npm i -g aside`; delete the credentials sentence; reword the install line to omit `operator`; each fails its subtest.
- **Green path**: M2.1 then M2.2.

### AC-ASB-008 — explicit-only and CI exclusion (maps REQ-ASB-009, REQ-ASB-011)

- **Given** the e2e workflow after M3
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/e2e_explicit_only ` and `.../e2e_ci_excluded ` print.
- **RED-now** (E4): `grep -c -i aside internal/template/templates/.claude/skills/moai/workflows/e2e.md` → `0`, exit 1.
- **Mutant probe**: the workflow text with the CI line removed; with Aside inserted into the Phase 0.5 option rules (the `never offered` literal removed); each fails its subtest.
- **Green path**: M3.1, M3.2.

### AC-ASB-009 — orchestrator runs every Aside step (maps REQ-ASB-010)

- **Given** the e2e workflow after M3
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/e2e_orchestrator_executes_aside `, `.../e2e_execution_owner_carveout `, and `.../e2e_aside_output_bounded ` print. Bounded output for the orchestrator-run path: verbose `aside repl` output goes to a file under `e2e/.runs/`, the orchestrator's context receives the exit code and a bounded tail, and the screenshot is cited by path, never inlined — the same contract the e2e-tester has at `e2e.md:54`, stated for the orchestrator.
- **RED-now** (E4, E13): `grep -c 'except Aside' internal/template/templates/.claude/skills/moai/workflows/e2e.md` → `0`, exit 1.
- **Mutant probe**: restore the original `Execution owner` line (no carve-out); drop `e2e/.runs/` from the bounded-output line; drop the `Skill(...)` load instruction; each fails its named subtest.
- **Green path**: M3.1, M3.2.

### AC-ASB-010 — silent fallback and carve-out at both sites (maps REQ-ASB-012)

- **Given** the e2e workflow after M3
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/e2e_silent_fallback_no_aside_message `, `.../e2e_missing_toolchain_carveout `, `.../e2e_tool_bypass_line_carveout `, and `.../e2e_no_install_command ` print. What is checkable, exactly: the workflow text instructs silence (`silently`, `no Aside-specific message`), instructs no note, and carries `except Aside` at both the sequence header (`e2e.md:108`) and the `--tool` bypass sentence (`e2e.md:119`); what is not checkable here is what a model emits at run time.
- **RED-now** (E4, E13): the same `except Aside` count is `0`, exit 1. The unedited bypass sentence at `e2e.md:119` currently instructs the Surface and Install sequence for any `--tool` value, Aside included.
- **Mutant probes**: restore the original `:119` sentence (this is the case the first iteration's checker could not see); delete the header carve-out; add an Aside line containing `fallback note`; add an Aside install command; each fails its named subtest.
- **Green path**: M3.1, M3.2.

### AC-ASB-011 — screenshot evidence (maps REQ-ASB-013)

- **Given** the e2e workflow after M3
- **When** `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`
- **Then** `--- PASS: TestAsideSkillPolicyAnchors/e2e_repl_screenshot_evidence ` prints: the workflow names a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path, without asserting which side writes the bytes.
- **RED-now**: E4.
- **Measurement gate**: M3.0 observes how `aside repl` produces a screenshot file; the wording stays at the hedged form unless the observation supports more (unmeasured at plan time, see Gaps).
- **Mutant probe**: remove the screenshot sentence; the named subtest must fail.
- **Green path**: M3.0 to M3.2.

### AC-ASB-012 — parity, hash set, Codex emit, build (maps REQ-ASB-014)

- **Given** the tree after M3
- **When** run these in order: `cmp internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md .claude/skills/moai-ref-aside-browser/SKILL.md`; `cmp internal/template/templates/.claude/skills/moai/workflows/e2e.md .claude/skills/moai/workflows/e2e.md`; `diff internal/template/templates/.claude/agents/moai/e2e-tester.md .claude/agents/moai/e2e-tester.md` (count the `<` and `>` lines); `grep -c 'never invokes Aside' internal/template/templates/.codex/agents/moai/e2e-tester.toml`; `make agents-emit-check`; `make build`; then a plain `git diff -U4 -- internal/template/catalog.yaml`
- **Then** both `cmp` exit 0; the `diff` shows exactly 3 changed lines (the baseline, hunks `2d1` and `144c143`) and the new sentence appears in both files; the TOML grep prints `1` or more; `make agents-emit-check` and `make build` exit 0; and the changed `hash:` hunks, read against their `- name:` context lines, name **exactly** `moai-ref-aside-browser` (new entry), `moai`, and `e2e-tester`. The `moai` change comes from the `e2e.md` edit, the `e2e-tester` change from the e2e-tester definition (the committed `e2e-tester` hash equals the file's `shasum -a 256` at baseline, which is why it must move). Any other changed hash fails the criterion until explained.
- **RED-now** (E7): `cmp internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md .claude/skills/moai-ref-aside-browser/SKILL.md` → `cmp: ... No such file or directory`, exit 2. (E10): `grep -c -i 'aside' internal/template/templates/.codex/agents/moai/e2e-tester.toml` → `0`, exit 1.
- **Mutant probe**: change one byte in the root skill copy; `cmp` must exit 1; restore.
- **Why the e2e-tester steps stay after Q1**: the e2e-tester definition still changes by one sentence, so its core catalog hash still changes and `make agents-emit` still regenerates the Codex TOML (`make build` begins with `agents-emit-check`, which fails until it has).
- **Green path**: M2.3, M3.3, M3.4.

### AC-ASB-013 — neutrality (maps REQ-ASB-015)

- **Given** the tree after M3
- **When** `go test ./internal/template/... -count=1`, and `grep -n -E 'SPEC-[A-Z]|t1439|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{40}\b|\b[0-9a-f]{7,8}\b'` over each authored template file: the skill, `workflows/e2e.md`, `e2e-tester.md`, and `e2e-tester.toml`
- **Then** the package run exits 0 (neutrality, leak, and catalog audits included); the grep prints nothing for the skill, `e2e-tester.md`, and the TOML (exit 1 each), and for `e2e.md` prints exactly the two pre-existing lines (the frontmatter `updated:` date at line 13 and the example log path at line 305 — E12), no new line.
- **RED-now**: not applicable (invariant, regression-guard); the skill does not yet exist (E3), so its grep cannot run — a skill that fails the grep fails the criterion.
- **Mutant probe**: add a date string, then a 40-hex string, to the skill; the grep must print each line and the leak audit must fail; revert.
- **Green path**: M2.2, M3.2, M3.3.

### AC-ASB-014 — docs-site (maps REQ-ASB-001)

- **Given** the tree after M3
- **When** `grep -c -F 'moai mcp add aside --command aside --args mcp --scope user' docs-site/content/ko/guides/mcp-server.md docs-site/content/en/guides/mcp-server.md docs-site/content/ja/guides/mcp-server.md docs-site/content/zh/guides/mcp-server.md`; the numeral checks `grep -c -F 'Five documented-but-disabled' docs-site/content/en/guides/mcp-server.md`, `grep -c -F '다섯 가지 documented-but-disabled' docs-site/content/ko/guides/mcp-server.md`, `grep -c -F '5つの documented-but-disabled' docs-site/content/ja/guides/mcp-server.md`, `grep -c -F '五个 documented-but-disabled' docs-site/content/zh/guides/mcp-server.md`; then `cd docs-site && hugo --minify --gc` and `test -f docs-site/public/sitemap.xml`; then the locale-parity ratchet of the `hns-oss-docs-verify` recipe (its section 4: the `comm -23` against `docs-site/.locale-parity-baseline` must print nothing)
- **Then** each locale file reports `1` or more for the command; each numeral grep prints `1` (heading line 63); the old numerals are gone from line 65 in all four (`Four external servers`, `네 개의 외부 서버`, `4つの外部サーバー`, `四个外部服务器` each count `0`); `hugo` exits 0 and prints no `WARN` or `ERROR`; the sitemap exists; the ratchet prints nothing. The edit adds a table row and one paragraph, no heading, so per-page section counts stay equal across locales.
- **RED-now** (E9, E14): the same command grep with `-i aside` on each locale file reports `0`, exit 1 (en, ko, ja, zh all measured); the new-numeral greps report `0`, exit 1 for all four locales.
- **Green path**: M3.5.

---

## Edge cases

- Aside installed but the operator is not signed in: the probe (`aside --version`) can pass while a repl call fails; the orchestrator treats a failed repl step as a fallback condition with the same silence, never as a reason to ask for credentials.
- `CI=true` together with `--tool aside`: excluded, default toolchain, no Aside output (AC-ASB-008, AC-ASB-010).
- `--tool aside` on a mobile or desktop-native platform class: not applicable, default toolchain, silent.
- Aside absent entirely: identical to a failed probe; no install command is shown and nothing about Aside is printed.
- A subagent whose task would need Aside, for example the e2e-tester asked to run `aside` by a spawn prompt: it returns a blocker report; the orchestrator runs the step itself or drops it (AC-ASB-006).

## Quality gate and Definition of Done

- All fourteen ACs pass with the commands above; every guard test has an observed failing run recorded in `progress.md` §E.2 **in the commit that introduces the test**, and that commit precedes the implementation commit. What git can verify, and what it cannot: `git log --reverse --format=%s` over the card branch lists, within each milestone, the `RED` commit before the `GREEN` commit (subjects carry the tokens `M2 RED` before `M2 GREEN`, `M3 RED` before `M3 GREEN`, and a single `M1 RED-record` commit, because M1's guards pass at birth and have no implementation commit). The order of the failing run inside a commit is not claimed.
- `go test ./internal/template/... -count=1` and `go test ./internal/cli/ -run '^(TestMCP_Add_AsideDocumentedCommandLine|TestMCP_Add_IdempotentSkip)$' -v -count=1` exit 0 (both named tests print their `--- PASS` line); `golangci-lint` reports no new issues in the three touched test files; `gofmt` clean.
- Template and root mirrors verified byte-identical (apart from the stated baseline divergence); `make agents-emit-check` and `make build` exit 0.
- No change to the default `.mcp.json`, the settings templates, `delegation.yaml`, or any `moai mcp` Go source (plain `git diff --stat` confirms the file list in `spec.md` § 5).
- Gaps carried into the run report are named, not hidden.

## Evidence ledger (RED-now and baseline measurements)

Pin for every entry: tree `4bf547bcad7c155b1e91485921569db709ec3ac2` (worktree t1439, clean at measurement). `grep` here is the `ugrep` build on this machine (it prints a `warning:` line for a missing file).

| ID | Command | Stdout | Exit |
|----|---------|--------|------|
| E1 | `grep -c 'func TestMCPDefaultExcludesAside' internal/template/mcp_template_neutrality_test.go` | `0` | 1 |
| E1b | `go test ./internal/template/ -run '^TestMCPDefaultExcludesAside$' -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/template	0.431s [no tests to run]` | 0 |
| E2 | `grep -c 'TestMCP_Add_AsideDocumentedCommandLine' internal/cli/mcp_test.go` | `0` | 1 |
| E3 | `ls internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md` | `ls: internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md: No such file or directory` | 1 |
| E4 | `grep -c -i 'aside' internal/template/templates/.claude/skills/moai/workflows/e2e.md` | `0` | 1 |
| E5 | `grep -c -i 'aside' internal/template/templates/.claude/agents/moai/e2e-tester.md` | `0` | 1 |
| E6 | `grep -c 'moai-ref-aside-browser' internal/template/catalog.yaml` | `0` | 1 |
| E7 | `cmp internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md .claude/skills/moai-ref-aside-browser/SKILL.md` | `cmp: internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md: No such file or directory` | 2 |
| E8 | `grep -c 'func TestAsideSkillPolicyAnchors' internal/template/aside_skill_policy_test.go` | `ugrep: warning: internal/template/aside_skill_policy_test.go: No such file or directory` | 2 |
| E9 | `grep -c -i 'aside' docs-site/content/en/guides/mcp-server.md` (and ko, ja, zh) | `0` each | 1 each |
| E10 | `grep -c -i 'aside' internal/template/templates/.codex/agents/moai/e2e-tester.toml` | `0` | 1 |
| E11 | `go test ./internal/template/ -run '^(TestCatalogManifestPresent\|TestAllSkillsInCatalog\|TestCatalogTierValid\|TestCatalogNoDuplicateEntries\|TestCatalogReferencesValid\|TestWorkflowTriggerCoverage\|TestSlimFS_HidesNonCoreEntries\|TestSlimFS_PreservesCoreEntries\|TestManifestHashFormat\|TestCatalogHashCoversSkillSubfiles\|TestPublishedSkillsNamesMatchTree\|TestEmbeddedMoaiSkillNames)$' -v -count=1` | twelve `--- PASS:` lines, last line `ok  	github.com/modu-ai/moai-adk/internal/template	0.459s` | 0 |
| E12 | `grep -c 'never invokes Aside' internal/template/templates/.claude/agents/moai/e2e-tester.md` | `0` | 1 |
| E13 | `grep -c 'except Aside' internal/template/templates/.claude/skills/moai/workflows/e2e.md` | `0` | 1 |
| E14 | `grep -c -F 'Five documented-but-disabled' docs-site/content/en/guides/mcp-server.md` (and the ko, ja, zh numeral forms of AC-ASB-014) | `0` each | 1 each |
| E15 | `grep -n -E 'SPEC-[A-Z]\|t1439\|[0-9]{4}-[0-9]{2}-[0-9]{2}\|\b[0-9a-f]{40}\b\|\b[0-9a-f]{7,8}\b' internal/template/templates/.claude/skills/moai/workflows/e2e.md` | `13:  updated: "2026-07-14"` and `305:  - Log: e2e/.runs/20260101-120000-suite.log (bounded excerpt above; full log at path)` | 0 |

## Gaps (not observed)

- User-scope run of the documented command against a real user config file (HOME override refused by the guard; covered by the seam test at run time).
- Whether `aside repl` can persist a screenshot to a path (M3.0 measures it; launching a browser was not done at plan time).
- Behavior of the MCP `exec` tool's permission argument (out of scope, unmeasured).
- The full `go test ./internal/template/... -count=1`, `make build`, and the docs-site `hugo` build were not run at plan time (the first two write generated artifacts; the last belongs to run phase). Only the 12-selector baseline (E11) was run.
- Whether a model emits anything about Aside at run time despite the silent-fallback text (lexical check only).
