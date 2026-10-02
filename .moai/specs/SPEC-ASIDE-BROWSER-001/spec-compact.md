# spec-compact — SPEC-ASIDE-BROWSER-001

> Run-phase digest: requirements, acceptance criteria, files, exclusions. Full text: `spec.md`, `acceptance.md`. Tier M, status draft, version 0.2.0, tree `4bf547bcad7c155b1e91485921569db709ec3ac2`. Operator verdicts (2026-10-02): orchestrator-only execution, core skill, completely silent fallback.

## Requirements (GEARS)

- REQ-ASB-001 (Ubiquitous): The policy skill and the docs-site MCP guide (ko, en, ja, zh) SHALL document `moai mcp add aside --command aside --args mcp --scope user` as the optional way to register Aside's MCP server.
- REQ-ASB-002 (Unwanted): The distributed default MCP configuration and the distributed settings templates SHALL NOT contain an `aside` entry or token.
- REQ-ASB-003 (Event-driven): When the documented registration command is run any number of times in project or user scope, `moai mcp add` SHALL leave exactly one `aside` entry with `command` `aside` and `args` `["mcp"]` and every unrelated entry unchanged.
- REQ-ASB-004 (Ubiquitous): The skill `moai-ref-aside-browser` SHALL be a non-user-invocable reference skill stating when Aside is appropriate, when it is not, and that Aside is operated by the orchestrator alone.
- REQ-ASB-005 (Unwanted): The skill SHALL prohibit `--permission full-access` in every form, direct that `--permission` be omitted (Guard), and limit `aside repl` code to navigation, reading, and screenshot capture unless the operator confirmed a state-changing action.
- REQ-ASB-006 (Event-driven): When an Aside action would change state in the operator's browser session, the skill SHALL require explicit operator confirmation, obtained by the orchestrator through its question channel, before the action runs.
- REQ-ASB-007 (Unwanted): A subagent SHALL NOT invoke Aside (the `aside` CLI in any form, or an Aside MCP tool) under any circumstance; the skill and the e2e-tester definition SHALL each state this, and a subagent whose task needs Aside SHALL return a blocker report.
- REQ-ASB-008 (Unwanted): The skill SHALL prohibit installing Aside or its skills automatically (only advise the operator to run `aside skills install`) and SHALL prohibit placing credentials, cookies, tokens, or session data in any output, report, commit, or memory entry.
- REQ-ASB-009 (Capability gate): Where the operator passes `--tool aside` explicitly, the e2e workflow SHALL use Aside through the `aside repl` CLI; Aside SHALL NOT be auto-detected, offered among Phase 0.5 options, or recommended.
- REQ-ASB-010 (State-driven): While Aside is the active e2e toolchain, the workflow SHALL assign every Aside step to the orchestrator (carved out of the e2e-tester's execution ownership), SHALL instruct it to load the policy skill first, and SHALL keep Aside output bounded in the orchestrator's context.
- REQ-ASB-011 (State-driven): While `CI=true`, the e2e workflow SHALL treat Aside as unavailable even when requested and continue on the platform default toolchain.
- REQ-ASB-012 (Event-driven): When the Aside probe fails or Aside is excluded (`CI=true`, non-web platform), the workflow SHALL continue on the default toolchain with no Aside-specific message, prompt, install attempt, or failure, and the missing-toolchain Surface and Install steps SHALL NOT apply to Aside at either site that instructs them.
- REQ-ASB-013 (State-driven): While Aside is active, journey evidence SHALL be a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path.
- REQ-ASB-014 (Ubiquitous): The skill SHALL have a core-tier catalog entry whose hash matches its content; skill and `workflows/e2e.md` SHALL be byte-identical between template and root; the e2e-tester definition SHALL be identical between template and root except its pre-existing divergence; the committed Codex e2e-tester definition SHALL match its emission from source; and the build SHALL succeed.
- REQ-ASB-015 (Unwanted): Authored template content SHALL NOT carry card ids, SPEC ids, commit SHAs, or dates, and SHALL NOT position any one programming language as primary.

## Acceptance criteria (Given / When / Then)

- AC-ASB-001 — Given M1, When `go test ./internal/template/ -run '^(TestMCPDefaultExcludesAside|TestMCPNeutralityTemplateShape)$' -v -count=1`, Then both print `--- PASS` (none selected zero); an `aside` key in `.mcp.json` fails both, a settings-template token fails only the new test. (REQ-ASB-002)
- AC-ASB-002 — Given M1 and M2, When `go test ./internal/cli/ -run '^TestMCP_Add_AsideDocumentedCommandLine$' -v -count=1` and the skill contains the documented line, Then `project` and `user` subtests pass with one idempotent `aside` entry. (REQ-ASB-001, REQ-ASB-003)
- AC-ASB-003 — Given M2, When the 13 catalog, slim, published, and mirror guards and the description-cap subtest run, Then all pass; the skill is `user-invocable: false` and `tier: core`. (REQ-ASB-004, REQ-ASB-014)
- AC-ASB-004 — Given the skill, When `go test ./internal/template/ -run '^TestAsideSkillPolicyAnchors$' -v -count=1`, Then `full_access_prohibited`, `guard_by_omission`, `repl_read_only_limit` pass and the negative controls (including `never forget to run aside --permission full-access`) fail. (REQ-ASB-005)
- AC-ASB-005 — Same test: `write_confirmation_via_orchestrator`, `orchestrator_only_operation` pass. (REQ-ASB-006)
- AC-ASB-006 — Same test: `subagent_never_invokes_skill`, `subagent_never_invokes_tester` pass; deleting either sentence fails the subtest. (REQ-ASB-007)
- AC-ASB-007 — Same test: `no_auto_install_advise_only`, `no_credentials_in_outputs` pass. (REQ-ASB-008)
- AC-ASB-008 — Same test: `e2e_explicit_only`, `e2e_ci_excluded` pass against `e2e.md`. (REQ-ASB-009, REQ-ASB-011)
- AC-ASB-009 — Same test: `e2e_orchestrator_executes_aside`, `e2e_execution_owner_carveout`, `e2e_aside_output_bounded` pass. (REQ-ASB-010)
- AC-ASB-010 — Same test: `e2e_silent_fallback_no_aside_message`, `e2e_missing_toolchain_carveout`, `e2e_tool_bypass_line_carveout`, `e2e_no_install_command` pass; restoring the original `:119` sentence fails. (REQ-ASB-012)
- AC-ASB-011 — Same test: `e2e_repl_screenshot_evidence` passes; wording pinned to the M3.0 measurement. (REQ-ASB-013)
- AC-ASB-012 — Given M3, When the `cmp`, diff-count, TOML grep, `make agents-emit-check`, `make build`, and `git diff -U4 -- internal/template/catalog.yaml` commands run, Then mirrors are identical, the e2e-tester pair differs only by its 3 baseline lines, both make targets exit 0, and the changed hash set is exactly `{moai-ref-aside-browser, moai, e2e-tester}`. (REQ-ASB-014)
- AC-ASB-013 — Given M3, When `go test ./internal/template/... -count=1` and the neutrality grep (with a 40-hex pattern) over each authored template file, Then exit 0, no hits in the skill, e2e-tester, and TOML, and only the two pre-existing hits in `e2e.md`. (REQ-ASB-015)
- AC-ASB-014 — Given M3, When the command and numeral greps run on the four `guides/mcp-server.md` files and `cd docs-site && hugo --minify --gc` and the parity ratchet run, Then all match and the build is warning-free. (REQ-ASB-001)

## Files

Skill (template + root mirror), `internal/template/catalog.yaml`, `workflows/e2e.md` (template + root), `e2e-tester.md` (template + root, one sentence), generated `e2e-tester.toml`, `mcp_template_neutrality_test.go`, `internal/cli/mcp_test.go`, `internal/template/aside_skill_policy_test.go`, `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md` — 15 files, the Tier M ceiling.

## Exclusions (What NOT to Build)

### Out of Scope — Aside behavior beyond what was observed

- MCP `exec` permission semantics; `aside exec` as an e2e driver; account, host, model options.

### Out of Scope — Other MoAI surfaces

- Go changes to `moai mcp`; the default `.mcp.json`, settings templates, `delegation.yaml`; `settings-management.md`; the docs-site skill-guide table and README ref-skill lists; docs-site `moai-e2e.md` and router entries; any Aside-specific message in the e2e report.

### Out of Scope — Operating Aside

- Running Aside in CI, installing Aside or its skills, moving the skill into a plugin.
