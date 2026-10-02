# plan — SPEC-ASIDE-BROWSER-001

> Tier M · development mode `tdd` · route per `repo-local-pr-policy.md` (git-flow: card branch, local merge into develop by the lane; no card PR) · all coordinates against tree `4bf547bcad7c155b1e91485921569db709ec3ac2` (card t1439, Epic Mods design 6/6). Revision 0.2.0: operator verdicts Q1 (orchestrator only), Q2 (core), Q3 (silent) applied.

## §A Context

### A.1 Background

Optional Aside integration, three surfaces (documented optional MCP, thin policy skill, explicit-only `/moai e2e --tool aside`). The safety constraints are the heart of the SPEC: each becomes an acceptance criterion with a runnable check. Requirements: `spec.md` § 2. Verification: `acceptance.md`. Facts and unmeasured items: `research.md`. Open and answered decisions: `decision-index.md`.

### A.2 Decision map by reversibility — review these first

Highest change-likelihood first (user-facing flows, then policy wording, then mechanical steps).

| # | Decision | Reversibility | Lives in |
|---|----------|---------------|----------|
| D1 | Orchestrator-only execution and the e2e phase split (e2e-tester keeps Phases 0-1, the orchestrator runs the Aside steps) | Low — changes who runs what in `/moai e2e` | M3, `e2e.md` (answered Q1; open Q8) |
| D2 | Completely silent fallback, carve-out at both missing-toolchain sites, CI=true exclusion form | Low — changes what `/moai e2e` prints | M3, `e2e.md` (answered Q3; open Q4) |
| D3 | Skill wording of the prohibitions, literal anchors, repl read-only limit | Medium — text, but guard-pinned | M2 |
| D4 | Core catalog tier; no `domain_skills` injection | Medium — one catalog entry, one config line | M2 (answered Q2; open Q6) |
| D5 | Docs-site scope (`mcp-server.md` x4, numeral) | High — additive prose | M3 (Q7) |
| D6 | Guard tests, mirrors, hashes | High — mechanical | M1, M2, M3 |

Execution order is M1, M2, M3 (guards first, so the observed-RED protocol runs on the cheapest surface); review order is D1 to D6, so a reviewer should open M3 and M2 before M1.

### A.3 Epic reference and dependency on Mods 1/6 (card t1434)

Standalone SPEC inside Epic "Mods design" (card t1439 is item 6 of 6). The card requires the dependency on Mods 1/6 (t1434) to be reported in the plan; this is that report.

- **What t1434 measured.** Its verdict (path `.moai/worktrees/t1434/.moai/reports/t1434/verdict.md`; stamp on line 21: `claude=2.1.287 codex=0.160.0`, date `2026-10-02`) reads, at line 27, `R01 skills`: `PLUGIN-OK` for Claude and for Codex; and at line 32, `R06 MCP servers`: `PLUGIN-OK` for both. These cells are capability only: t1434 decided nothing about moving any component.
- **Two measured consequences that matter here.** (1) Plugin skills take a namespaced name (the verdict's example is `p-skill:probe-skill`), so if this skill later ships through a plugin its invocation name changes. (2) Claude drops a plugin MCP server when a project server runs the same command (the verdict's `duplicates manually-configured`), so a project-scope registration of `aside mcp` would shadow a plugin copy.
- **Conclusion: no blocking dependency, because** this SPEC changes only template files and the project-scope MCP registration command; it requires no plugin delivery, the skill ships in the core init payload today, and t1434's cells are capability observations, not a decision. If a later card moves skills or MCP servers into a plugin, the only edits this SPEC's outputs would need are the skill's load name in `e2e.md` (`Skill("moai-ref-aside-browser")`) and the documented registration command; both are single-line follow-ups.
- **Reachability caveat.** The verdict was read at plan time from the sibling worktree t1434; it is local to that worktree and not merged into this branch. A reader on another machine, or on this branch, cannot fetch it; the cells above are quoted here for that reason. Re-read the verdict at its merged location, if it lands, before relying on the line numbers.

## §B Known Issues (filtered to what applies)

- **B-Template-First**: edit `internal/template/templates/` first; `make build` runs `agents-emit-check`, which fails after any `e2e-tester.md` edit until `make agents-emit` has regenerated the TOML (`internal/template/CLAUDE.md`). This still applies after Q1, because the e2e-tester definition still gains one sentence.
- **B-Pre-existing mirror divergence**: `e2e-tester.md` template and root already differ in 3 changed lines (baseline measured: `diff` reports hunks `2d1` and `144c143`); `settings-management.md` pair differs at byte 18093. Do not "fix" either; the parity AC is stated against the baseline.
- **B-Empty sweep**: a `go test` run whose name filter matches no test prints `ok ... [no tests to run]` with exit 0 (measured). Every check requires the `--- PASS:` line for its named test, never the exit code alone.
- **B-Refused commands in a worktree session** (observed): a command that sets `HOME=`, uses `env -C`, or composes `git` with other commands is refused by the isolation guard. The user-scope MCP test therefore uses the `userHomeDirFn` seam (`internal/cli/glm_tools.go:372`), not an environment override; verification commands stay plain and separate.
- **B-Catalog hash set**: `make build` regenerates catalog hashes with `--all`. The expected changed set is exactly `{moai-ref-aside-browser (new entry), moai, e2e-tester}`: `moai` because `workflows/e2e.md` lives inside the core `moai` skill, and `e2e-tester` because the core agent entry's committed hash equals the `shasum -a 256` of `e2e-tester.md` at baseline, so the one-sentence edit must move it. Read `git diff -U4 -- internal/template/catalog.yaml`: any other changed `hash:` hunk fails AC-ASB-012 until explained. `generated_at` may also change; it is not a hash.
- **B-Core-tier guards**: the 12 selectors that read the core tier or the slim and published surfaces all passed at baseline (E11). None hard-codes a core or total count (totals derive from `catalog.yaml`). Adding a core skill therefore needs no constant bump, but its description joins the skill listing of every install, so keep `description` plus `when_to_use` short (the measured cap is a subtest). The history comments in `catalog_loader_test.go`, `embed_catalog_test.go`, and `catalog_tier_audit_test.go` are not touched.
- **B-Neutrality**: no card ids, SPEC ids, SHAs, or dates in template content; `skill_dir_token_guard_test.go` forbids a `CLAUDE_SKILL_DIR` token; `published_skills_deploy_test.go` pins the published command skills — this skill must not become one (`user-invocable: false`, no command wrapper). `e2e.md` already has two pattern hits at baseline (E15); the edit must not add a third.
- **B-Version variance**: pin only the flags observed on Aside `1.26.916.1741`; word the skill so a different version degrades to "probe fails, silent fallback".

## §C Pre-flight (before M1)

```bash
git rev-parse --short HEAD
git branch --show-current
go build ./...
aside --version
```

Baseline attributions to record in `progress.md` §E.2: HEAD SHA, `go build ./...` exit, `aside --version` output (or its absence, which is a valid run condition), and the three baseline mirror facts (`cmp` e2e.md exit 0; e2e-tester diff changed-line count 3; settings-management differs).

## §D Constraints

- PRESERVE: default `.mcp.json`, settings templates, `delegation.yaml`, `settings-management.md`, every other skill, every Go source file outside the three test files.
- Do not touch `moai mcp` Go code. If a test needs a seam that does not exist, return a blocker report.
- Subagents never prompt the user; ambiguity returns a blocker report to the orchestrator.
- Conventional commits, card id in every commit message, no `--no-verify`, stage by explicit pathspec.

## §E Self-Verification (what the run phase reports)

Each `acceptance.md` AC is reported as command, verbatim output lines that decided it, exit code, tree HEAD, and the judging build's commit when a tool built from the tree is cited (`verification-claim-integrity.md` §2.2). Observed-RED output for each guard test goes verbatim into `progress.md` §E.2 in the commit that introduces the test.

## §F Milestones (priority order; no time estimates)

### Observed-RED-first protocol and commit order (applies to M1, M2, M3)

1. Write the check first. Run it against a constructed failing input and record the failing output (the mutant is a temporary edit or an in-test negative control; revert it and confirm with a plain `git diff` that the tree is back).
2. **Commit order** (chosen over rewriting the Definition of Done, so `git log` can witness it, per `verification-claim-integrity.md` §2.3): within M2 and M3 the guard test and its recorded RED run are committed **before** the implementation commit. Subjects carry the tokens `M2 RED` then `M2 GREEN`, `M3 RED` then `M3 GREEN`. M1's guards pass at birth (they assert invariants that already hold), so M1 is a single `M1 RED-record` commit carrying the tests and the mutant-run record; there is no M1 implementation commit.
3. Only then make the change that turns it GREEN and record the passing output.
4. A check whose swept set is empty (`[no tests to run]`) is not a pass.

### M1 (Priority High) — guards and the documented-command test

Files: `internal/template/mcp_template_neutrality_test.go`, `internal/cli/mcp_test.go`. Commit `M1 RED-record`.

- M1.1 Add `TestMCPDefaultExcludesAside`: asserts the template `.mcp.json` has no `aside` key and that no `aside` token appears in `settings.json.tmpl`. The `.mcp.json` half deliberately overlaps `TestMCPNeutralityTemplateShape` (it names the constraint and its failure message says Aside); the settings half is the non-duplicate part. The test reads files from disk, so a mutant needs no rebuild. Mutants: (a) add an `aside` key to the template `.mcp.json` — expect `--- FAIL: TestMCPDefaultExcludesAside ` and `--- FAIL: TestMCPNeutralityTemplateShape ` (each name followed by a space; observe the second, do not infer it); (b) add the token to `settings.json.tmpl` — expect only the new test to fail; revert both.
- M1.2 Add `TestMCP_Add_AsideDocumentedCommandLine` (subtests `project`, `user`): drives the cobra command with the documented arguments, twice, against `t.TempDir()` (the user scope through the `userHomeDirFn` seam), asserts one `aside` entry with `args` `["mcp"]`, a byte-identical file after the second run, and unrelated entries preserved. Mutant: misspell `--args` as `--arg` in the test invocation; expect `unknown flag`; revert.

### M2 (Priority High) — policy skill, core catalog entry, skill checker

Files: new skill (template + root mirror), `internal/template/catalog.yaml`, `internal/template/aside_skill_policy_test.go`. Commits `M2 RED` then `M2 GREEN`.

- M2.1 (`M2 RED`) Write `aside_skill_policy_test.go` first: the pure checker `checkAsideSkill(text)` over the literal anchors of `acceptance.md` § Checker anchors (skill rows), run on the real skill (expect no violations) and on negative controls — each anchor removed, and the bad lines added (unqualified `--permission full-access`, `never forget to run aside --permission full-access`, `npm i -g aside`). RED is natural: the skill does not exist yet, so the first run fails reading it; record that output.
- M2.2 (`M2 GREEN`) Author the skill (name `moai-ref-aside-browser`; frontmatter per `skill-authoring.md`: folded `description` plus `when_to_use` within the 1,536-character cap, `user-invocable: false`, no `CLAUDE_SKILL_DIR` token). Body sections: purpose; when to use and not; operated by the orchestrator alone and the subagent sentence; the safe-use rules (full-access, Guard by omission, repl limit, write confirmation, install advice, no credentials in outputs); the optional registration command; relation to `/moai e2e`. Target roughly 80 lines. English, language-neutral, no dates or ids; the anchors are literal substrings.
- M2.3 Catalog: add the entry under `catalog.core.skills` (`tier: core`), run `make build` so `gen-catalog-hashes` fills the hash, mirror the skill to `.claude/skills/moai-ref-aside-browser/SKILL.md` with `cp`, confirm `cmp` exit 0. `.agents/skills` needs no hand edit: its set is derived and gitignored (guards `TestSkillMirror_SetIsDerivedNotConstant`, `TestGitignore_IgnoresSkillMirrorOnly`, run to confirm). Do not add the skill to `delegation.yaml` (decision-index Q6).

### M3 (Priority High) — e2e wiring, e2e-tester sentence, Codex emit, docs-site

Files: `e2e.md` (template + root), `e2e-tester.md` (template + root), the generated TOML, `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md`. Commits `M3 RED`, `M3 GREEN`, then a docs commit.

- M3.0 Measurement first: observe, on a non-sensitive page (a blank tab or a public example page, with the operator's approval of any browser launch), how `aside repl` can produce a screenshot file. Pin the evidence wording in `e2e.md` to what is observed; if it cannot be saved to a path, record the observed mechanism and return a blocker report so the orchestrator can amend REQ-ASB-013 — do not guess.
- M3.1 (`M3 RED`) Extend `aside_skill_policy_test.go` with `checkAsideE2E(e2eText)` and the e2e-tester half of the subagent anchor, with negative controls (including the mutant that restores the original `e2e.md:119` bypass sentence). Record the failing run.
- M3.2 (`M3 GREEN`) `e2e.md` edits, each tied to a measured line: `:36` execution-owner sentence gains `except Aside` and assigns Aside to the orchestrator; `:40` flag list gains `aside` with the explicit-only sentence; `:54` bounded-output rule is restated for the orchestrator-run path (output redirected under `e2e/.runs/`, exit code and bounded tail, path cited); a probe-table row for `aside --version` whose install column says the workflow never installs it; Phase 0.5 clauses (never offered, never recommended, never auto-detected) and the `CI=true` clause (unavailable); the **`:108` sequence header** gains `except Aside`; the **`:119` bypass sentence** (`run the missing-toolchain sequence if absent`) gains `except Aside`, with this explicit precedence sentence placed at the carve-out: "An absent or excluded Aside is not a missing toolchain: continue silently on the platform default, with no Aside-specific message, prompt, install attempt, or failure; this rule takes precedence over the Surface and Install steps and over the `--tool` bypass." A Tool Matrix row (web, CLI via `aside repl`, explicit-only, evidence = screenshot captured through `aside repl`, saved under `e2e/` by path). The e2e-tester keeps Phases 0 and 1; the orchestrator loads `Skill("moai-ref-aside-browser")` before its first Aside step and runs every Aside step itself.
- M3.3 `e2e-tester.md`: one negative sentence — "A subagent never invokes Aside (the `aside` CLI or the Aside MCP tools); a task that needs Aside returns a blocker report to the orchestrator." — placed after the selection-is-out-of-scope paragraph, identically in the template and root files. Then `make agents-emit`, `make agents-emit-check`, `make build`. These steps stay after Q1 because the e2e-tester definition still changes: its core catalog hash moves and the Codex TOML regenerates.
- M3.4 Mirror `e2e.md` to root with `cp`, `cmp` exit 0; confirm the e2e-tester pair still differs only by its 3 baseline lines; read the catalog diff against the exact hash set of § B.
- M3.5 (docs commit) Docs-site, ko canonical then en, ja, zh (re-read at plan time: heading line 63 and sentence line 65 of each locale carry the numeral four): add one table row and one short paragraph for the optional Aside MCP in the "documented-but-disabled" section, carrying the exact registration command and the read-only and orchestrator-confirmation sentence; change the numeral in the heading and in the sentence to five (`Five`, `다섯 가지`, `5つ`, `五个`; sentence forms `Five external servers`, `다섯 개의 외부 서버`, `5つの外部サーバー`, `五个外部服务器`). Add no heading, so per-page section counts stay equal across locales. Follow `hns-oss-docs-i18n-rules`; run the `hns-oss-docs-verify` commands cited in AC-ASB-014. Native idiom per locale, no calques.

## §G Anti-patterns

- Adding Aside to any default-enabled config, or to `domain_skills`.
- Making the MCP registration a prerequisite of the e2e path.
- Any instruction that has a subagent invoke Aside, directly or through a spawn prompt.
- A fallback that asks the operator, runs an install command, or prints anything about Aside.
- Quoting `--permission full-access` as a usage example anywhere, even "for reference" (the checker's closed allow-list treats any other occurrence as a violation).
- "Fixing" the pre-existing mirror divergences in passing.
- Reading a green `go test` that selected zero tests as a pass.

## §H Cross-references

- `.claude/rules/moai/development/verification-completeness.md` § 1.1, § 2 (observed-failure completion, RED-now/green-path pairing)
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.2, § 2.3 (tool provenance; the commit graph as the ordering witness)
- `.claude/rules/moai/development/skill-authoring.md`; `internal/template/CLAUDE.md` (Template-First)
- `.claude/skills/moai/workflows/e2e.md:36, 40, 53-56, 108-133` (execution owner, flag line, hard rules, probe and Phase 0.5)
- `internal/cli/mcp.go`, `internal/cli/glm_tools.go:363-377` (scope resolution)
- `internal/template/mcp_template_neutrality_test.go` (`mcpAllowedActiveKeys`)

## File list after the revision (recount)

15 files, the Tier M ceiling: skill template and root (2), `catalog.yaml` (1), `e2e.md` template and root (2), `e2e-tester.md` template and root (2), generated TOML (1), three Go test files (3), four docs-site guides (4). Q1 removed the e2e-tester recipe block and the delegated-prompt subtests but not a file, because the one-sentence prohibition keeps the e2e-tester pair and the TOML. D13's README and skill-guide surfaces stay out of scope to hold the ceiling.

## MX tag plan (Phase 14, lightweight scan)

New Go code is test-only; no exported function reaches fan_in >= 3, no goroutines, no complexity hotspot. No MX targets. The skill and workflow edits are Markdown and carry no MX tags.

## Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Aside behavior varies by version | Wording drifts from the tool | Pin only observed flags; any probe failure degrades silently to the default toolchain |
| The e2e fallback carve-out contradicts the two existing missing-toolchain sites | Operators see a prompt they should not | Carve-out stated at the header and the bypass sentence, each pinned by a subtest with a mutant that restores the original line |
| `aside repl` cannot persist a screenshot to a path | REQ-ASB-013 wording unsupported | M3.0 measures first; blocker report if the observation differs |
| `aside repl` has no permission flag | "read-only" is discipline, not enforcement | The skill states the limit; the checker pins the literal; residual risk named in `research.md` |
| Orchestrator-run Aside steps enlarge the orchestrator's context | Cost, stall risk | Output redirected to `e2e/.runs/`, bounded tail only, screenshots cited by path (AC-ASB-009) |
| The subagent prohibition rests on prompt text | A model could still call Aside | One sentence in the e2e-tester definition and the skill, each pinned by a mutant; no hook enforces it |
| Core tier puts the skill description in every install's listing | Context bytes | Short `description` plus `when_to_use`; the cap is measured by a subtest |
| README ref-skill counts drift by one | Stale prose | Known and out of scope (`spec.md` § 7); no guard depends on it |
