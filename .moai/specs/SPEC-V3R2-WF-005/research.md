# SPEC-V3R2-WF-005 Deep Research (Phase 0.5)

> Research artifact for `Language Rules vs Skills Boundary Codification`.
> Companion to `spec.md` (v0.2.0). Authored against branch `feature/SPEC-V3R2-WF-005-language-rules-boundary` from `/Users/goos/MoAI/moai-adk-go` (solo mode, no worktree).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (run-phase research) | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

Substantiate `spec.md` §1 (Goal), §1.1 (배경), §2.1 (In Scope), and §4 (Assumptions) with concrete file:line evidence so that the run phase can implement REQ-WF005-001..015 against a known-good baseline. The research answers four questions:

1. Where do the 16 language rules currently live, and do they all use `paths:` frontmatter (REQ-WF005-004)?
2. Which skill files reference non-existent `moai-lang-*` skills, and how many references exist in total (REQ-WF005-005)?
3. Which skill/rule files reference the 4 other dead skill IDs (`moai-infra-docker`, `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`) called out in spec.md §2.1 (REQ-WF005-015)?
4. Which insertion anchor in `.claude/rules/moai/development/skill-authoring.md` is the appropriate target for the "language guidance lives in rules" principle (REQ-WF005-003)?

---

## 2. Inventory of `.claude/rules/moai/languages/` (16 files)

`ls .claude/rules/moai/languages/` returns 16 `.md` files matching the canonical list declared in `CLAUDE.local.md` §15 (16-language neutrality):

| # | File | Size (bytes) | `paths:` frontmatter declared? |
|---|------|--------------|---------------------------------|
| 1 | `cpp.md` | 7642 | YES — `**/*.cpp,**/*.hpp,**/*.h,**/*.cc,**/CMakeLists.txt` |
| 2 | `csharp.md` | 5427 | YES — `**/*.cs,**/*.csproj,**/*.sln` |
| 3 | `elixir.md` | 6655 | YES — `**/*.ex,**/*.exs,**/mix.exs` |
| 4 | `flutter.md` | 5953 | YES — `**/*.dart,**/pubspec.yaml,**/pubspec.lock` |
| 5 | `go.md` | 3652 | YES — `**/*.go,**/go.mod,**/go.sum` |
| 6 | `java.md` | 7601 | YES (per Bash sweep) |
| 7 | `javascript.md` | 8056 | YES |
| 8 | `kotlin.md` | 6505 | YES |
| 9 | `php.md` | 2314 | YES |
| 10 | `python.md` | 3399 | YES — `**/*.py,**/pyproject.toml,**/requirements*.txt` |
| 11 | `r.md` | 7700 | YES |
| 12 | `ruby.md` | 8090 | YES |
| 13 | `rust.md` | 6395 | YES |
| 14 | `scala.md` | 4717 | YES |
| 15 | `swift.md` | 4301 | YES |
| 16 | `typescript.md` | 7728 | YES |

Verification: `grep -l "^paths:" .claude/rules/moai/languages/*.md | wc -l` returns 16. **All 16 language rules already use `paths:` frontmatter — REQ-WF005-004 is satisfied at baseline.**

Canonical name confirmation per CLAUDE.local.md §15 + `spec.md` REQ-WF005-010: file `flutter.md` is named "flutter" (not "dart"). The `paths:` glob declares `**/*.dart` because the source files use `.dart` extension, but the rule filename and `project_markers` detection key are both "flutter".

---

## 3. Dead `moai-lang-*` skill references — existing codebase scan

A `Grep` sweep with pattern `moai-lang-` over `.claude/skills/` and `.claude/rules/` produced the following reference counts (top 12 paths):

| File | Count of `moai-lang-*` mentions | Reference type |
|------|---------------------------------|----------------|
| `.claude/skills/moai/workflows/run.md` | 16 | language detection mapping (lines 338-352) + Phase 0.9 example (line 979) |
| `.claude/skills/moai-foundation-core/modules/agents-reference.md` | 7 | catalog table rows (lines 268, 278-282, 292) |
| `.claude/skills/moai-framework-electron/SKILL.md` | 3 | `related-skills` (line 19) + body (lines 228, 230) |
| `.claude/skills/moai-platform-chrome-extension/SKILL.md` | 3 | `related-skills` (line 19) + body (lines 278, 279) |
| `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` | 3 | sub-agent definition examples (lines 30, 439, 936) |
| `.claude/skills/moai-platform-deployment/SKILL.md` | 2 | body (lines 398, 399) |
| `.claude/skills/moai-platform-auth/SKILL.md` | 2 | `related-skills` (line 20) + body (line 225) |
| `.claude/skills/moai-workflow-loop/SKILL.md` | 2 | body (lines 148, 149) |
| `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md` | 2 | example sub-agent definitions (lines 201, 392) |
| `.claude/skills/moai-foundation-cc/reference/skill-examples.md` | 2 | example sub-agent skill lists (lines 237, 427) |
| `.claude/skills/moai-domain-frontend/SKILL.md` | 1 | body (line 119) |
| `.claude/skills/moai-foundation-cc/references/examples.md` | 1 | example reference (line 175) |
| `.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md` | 1 | example body (line 113) |
| `.claude/rules/moai/languages/scala.md` | 1 | body (line 131) — *language rule cross-references another lang* |
| `.claude/rules/moai/languages/flutter.md` | 2 | body (lines 94, 95) |
| `.claude/rules/moai/languages/cpp.md` | 1 | body (line 100) |
| `.claude/rules/moai/development/agent-authoring.md` | 1 | YAML schema example (line 165, mentions `skills:` array exception convention) |

Aggregate: **17 distinct files** across `.claude/skills/` (15 files) and `.claude/rules/` (3 files: 3 language rules + 1 development authoring doc) reference the non-existent `moai-lang-*` skill IDs. Total inline mentions ≈ 50 (run.md alone accounts for 16).

Reference categories:
- `related-skills` frontmatter field (3 SKILL.md files: `moai-platform-auth`, `moai-framework-electron`, `moai-platform-chrome-extension`) — REQ-WF005-005 frontmatter target.
- Body prose listing "X for Y patterns" (most other files) — REQ-WF005-005 + REQ-WF005-013 body target.
- Documentation examples / catalog tables (`moai-foundation-core/modules/agents-reference.md`, `moai-foundation-cc/reference/...`) — REQ-WF005-008 substitute path target.
- Cross-language references inside the language rules themselves (e.g., `scala.md` mentioning `moai-lang-java`, `flutter.md` mentioning `moai-lang-swift` and `moai-lang-kotlin`, `cpp.md` mentioning `moai-lang-rust`) — REQ-WF005-005 + REQ-WF005-008 cleanup target. These are arguably the most surprising findings: the language rule cross-references its sibling language rules via the *skill* ID rather than a path reference. Run phase must convert these to `.claude/rules/moai/languages/<name>.md` path references.

Note on `agent-authoring.md:165`: the line documents the YAML format exception ("`skills:` field uses YAML array, not CSV") and uses `moai-lang-go` as the example. The reference is illustrative-only; run phase decides whether to keep the example with a real skill ID or substitute a placeholder. Recommendation: substitute the example with an existing skill ID such as `moai-domain-backend`.

---

## 4. Other dead skill ID references (per spec.md §2.1)

A second `Grep` sweep with pattern `moai-infra-docker|moai-essentials-debug|moai-quality-testing|moai-quality-security` produced 15 distinct line matches across 13 files:

### 4.1 `moai-essentials-debug` references

| File | Line | Context |
|------|------|---------|
| `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` | 621 | sub-agent skill list example |
| `.claude/rules/moai/languages/r.md` | 133 | "Related Skills" section |
| `.claude/rules/moai/languages/typescript.md` | 126 | "Related Skills" section |
| `.claude/rules/moai/languages/flutter.md` | 98 | "Related Skills" section |
| `.claude/rules/moai/languages/cpp.md` | 103 | "Related Skills" section |
| `.claude/rules/moai/languages/csharp.md` | 110 | "Related Skills" section |
| `.claude/rules/moai/languages/elixir.md` | 95 | "Related Skills" section |
| `.claude/rules/moai/languages/javascript.md` | 130 | "Related Skills" section |
| `.claude/rules/moai/languages/ruby.md` | 131 | "Related Skills" section |

Substitute per REQ-WF005-015: replace with "delegate to `expert-debug` agent" note.

### 4.2 `moai-quality-testing` references

| File | Line | Context |
|------|------|---------|
| `.claude/rules/moai/languages/kotlin.md` | 109 | "Related Skills" section |

Substitute per REQ-WF005-015: replace with `moai-foundation-quality` + `moai-ref-testing-pyramid`.

### 4.3 `moai-quality-security` references

| File | Line | Context |
|------|------|---------|
| `.claude/skills/moai-domain-backend/SKILL.md` | 110 | body — "Security validation and compliance" |
| `.claude/skills/moai-workflow-project/references/overview.md` | 120 | overview body |
| `.claude/rules/moai/languages/flutter.md` | 97 | "Related Skills" section |

Substitute per REQ-WF005-015: replace with `moai-foundation-quality` + `moai-ref-owasp-checklist`.

### 4.4 `moai-infra-docker` references

| File | Line | Context |
|------|------|---------|
| `.claude/rules/moai/languages/kotlin.md` | 110 | "Related Skills" section |
| `.claude/rules/moai/languages/java.md` | 121 | "Related Skills" section |

Substitute per REQ-WF005-015: **remove** (no substitute; platform infra deferred to a future `moai-platform-*` skill not yet specified).

Aggregate: 15 dead-skill-ID line references across 13 distinct files. Combined with the `moai-lang-*` references (17 files / ~50 mentions), the run-phase total cleanup surface is **~28 distinct files / ~65 mentions**.

---

## 5. Insertion anchor for skill-authoring.md (REQ-WF005-003)

`.claude/rules/moai/development/skill-authoring.md` exists at the project root (9536 bytes per `ls`). The file documents the YAML frontmatter schema, progressive disclosure, tool permissions by category, and best practices.

The natural insertion point for the "language guidance lives in rules" principle is a new section, appended after the existing "## Tool Permissions by Category" section. Recommended anchor: a new top-level section titled `## Language Guidance Lives in Rules, Not Skills` containing:

- The principle declaration (REQ-WF005-001 + REQ-WF005-002 verbatim).
- A pointer to `.claude/rules/moai/languages/` as the canonical location.
- The `paths:` frontmatter convention (REQ-WF005-004 reference).
- The `LANG_AS_SKILL_FORBIDDEN` CI key (REQ-WF005-007).
- A cross-link to `CLAUDE.local.md` §15 (16-language neutrality).
- A short rationale: paths-based loading is the structurally correct primary mechanism for language-scoped guidance; keyword-based skill activation is the wrong abstraction for files-on-disk language detection.

The "## Tool Permissions by Category" section already lists "Language Skills" as a permitted category — this is the strongest pre-existing source of confusion that REQ-WF005-003 must resolve. The new section MUST clarify that the "Language Skills" tool-permission row is reserved for hypothetical future skill-class language support; today, all 16 supported languages are rules.

---

## 6. CI Guard Architecture (REQ-WF005-007 / REQ-WF005-013 / REQ-WF005-014)

### 6.1 Existing audit-test scaffold

`internal/template/commands_audit_test.go:1-60` is the canonical pattern for static audit tests against the embedded skill/rule corpus (also referenced in WF-004 plan). The same pattern applies here:

1. Walk `.claude/skills/moai-lang-*/` directory tree in the embedded FS (`fs.WalkDir`).
2. If any directory matches the prefix `moai-lang-`, fail with the sentinel `LANG_AS_SKILL_FORBIDDEN`.
3. Walk `.claude/skills/**/SKILL.md` files. Read frontmatter. If `related-skills:` field contains a `moai-lang-*` token, emit `DEAD_LANG_SKILL_REFERENCE` warning with the rule-path substitute (`.claude/rules/moai/languages/<name>.md`).

### 6.2 Forbidden patterns (test will fail if matched)

- Directory pattern in embedded FS: `^\.claude/skills/moai-lang-[a-z-]+/`. Any match triggers `LANG_AS_SKILL_FORBIDDEN` (REQ-WF005-002, REQ-WF005-007, REQ-WF005-009).
- Frontmatter `related-skills:` token regex: `\bmoai-lang-[a-z]+\b`. Any match emits `DEAD_LANG_SKILL_REFERENCE` warning (REQ-WF005-013).
- Skill body 16-language neutrality regex: phrases like `(go|python|typescript|...)\s+is\s+(primary|the\s+main)` or `only\s+(go|python|...)\s+is\s+supported`. Any match triggers `LANG_NEUTRALITY_VIOLATION` (REQ-WF005-014).

### 6.3 Recommended test location

New file: `internal/template/lang_boundary_audit_test.go`. Mirror structure of `commands_audit_test.go:1-60`. Three test functions:

- `TestNoLangSkillDirectory` — fails if any `.claude/skills/moai-lang-*/` directory is present in the embedded FS. Sentinel: `LANG_AS_SKILL_FORBIDDEN`.
- `TestRelatedSkillsNoLangReference` — walks all `.claude/skills/**/SKILL.md` files in the embedded FS, parses frontmatter, asserts the `related-skills:` value (when present) does NOT contain `moai-lang-`. Sentinel: `DEAD_LANG_SKILL_REFERENCE`.
- `TestLanguageNeutrality` — walks all `.claude/skills/**/SKILL.md` and `.claude/rules/**/*.md` files in the embedded FS, scans bodies (excluding code blocks), asserts no language-primacy phrases match. Sentinel: `LANG_NEUTRALITY_VIOLATION`.

### 6.4 Cross-SPEC integration

- WF-005's CI guards are independent of WF-003 (`--mode` flag) and WF-004 (`AGENTLESS_CONTROL_FLOW_VIOLATION`) but share the same audit-test scaffold convention. No content collision risk.
- The "language guidance lives in rules" principle in skill-authoring.md is a single-source-of-truth document; downstream `.claude/skills/**/SKILL.md` and `.claude/rules/**/*.md` cleanup is a one-time run-phase migration plus the standing CI guard.

---

## 7. Reference Patterns to Preserve

The run phase MUST NOT touch the following — they encode the rhythm WF-005 is formalizing:

- **Reference**: `.claude/rules/moai/languages/*.md` (16 files) — `paths:` frontmatter is verified correct (R6 §4.2). Body content stays unchanged per `spec.md` §1.2 / §2.2 Out of Scope. Only the "Related Skills" sections inside language rules get edited (per §3 above) to remove dead-skill-ID lines or substitute rule-path references for cross-language pointers.
- **Reference**: `.claude/rules/moai/development/skill-authoring.md:1-269` — append-only; the new "Language Guidance Lives in Rules" section is added at the bottom.
- **Reference**: `internal/template/commands_audit_test.go:1-60` — audit-test scaffold pattern (preserve as scaffold reference, not modified).
- **Reference**: `CLAUDE.local.md` §15 — 16-language neutrality (preserve verbatim; the new principle in skill-authoring.md cross-links to it).
- **Reference**: `.claude/skills/moai/workflows/run.md:338-352` — the language detection mapping. This must be **edited** to replace `moai-lang-<name>` with `.claude/rules/moai/languages/<name>.md` (or a `Skill("moai-lang-<name>")` reference removed entirely, since the targets are rules, not skills). The mapping itself is preserved — only the right-hand-side pointer changes.

---

## 8. BC scope analysis (does this SPEC introduce a behavioral break?)

`spec.md` §10 implies no BC line item. Verifying:

| Surface | Today's behavior | Post-WF-005 behavior | BC? |
|---------|------------------|----------------------|-----|
| Language rule auto-loading via `paths:` | Loads on file match | Loads on file match | NO — runtime unchanged |
| Skill auto-activation by keyword | `moai-lang-*` keywords match no skill (dead reference, silent miss) | Keyword no longer documented; rule-path documented instead | NO — dead reference doesn't activate anything today |
| Agent prompts referencing `moai-lang-*` | Dead reference, silently ignored | Replaced with rule-path reference | NO — agent receives no skill body either way; the change is documentation accuracy |
| `skill-authoring.md` schema | Documents skill creation | Documents skill creation + forbids language-scoped skills | YES (declaration-level) — new principle constrains future PRs but doesn't break existing surfaces |
| CI guard | None for lang-as-skill | New `LANG_AS_SKILL_FORBIDDEN` test | NEW — forward-looking guard, no existing PR breaks |

**Conclusion**: like WF-004, this is a **declaration-level** change. The `moai-lang-*` skills don't exist today, so removing references doesn't break runtime behavior. The CI guards are forward-looking. `breaking: false` in spec.md frontmatter is correct (no `bc_id` entries).

---

## 9. Recommendations for the Run Phase

1. **Methodology**: Per `.moai/config/sections/quality.yaml` `development_mode`, this project uses TDD. The CI guard tests (REQ-WF005-007, REQ-WF005-013, REQ-WF005-014) MUST be written first (RED), confirmed failing, and then the cleanup edits (REQ-WF005-005, REQ-WF005-008, REQ-WF005-015) and the skill-authoring.md principle insertion (REQ-WF005-003) added (GREEN). Refactor pass at the end consolidates any duplication across substituted reference patterns.
2. **Solo mode discipline**: Run phase work happens directly on `feature/SPEC-V3R2-WF-005-language-rules-boundary` from `/Users/goos/MoAI/moai-adk-go`. No worktree (per user directive in this session).
3. **Multi-file decomposition**: The cleanup touches ~28 distinct files (per §3 + §4 aggregate). Per CLAUDE.md §1 HARD rule "Multi-File Decomposition: Split work when modifying 3+ files," tasks.md decomposes the cleanup into per-file or per-category sub-tasks.
4. **MX targets**: Pre-mark `internal/template/lang_boundary_audit_test.go` with `@MX:ANCHOR` (the test becomes the contract enforcer — high fan_in for future regressions). Pre-mark the new "Language Guidance Lives in Rules" section in `skill-authoring.md` with `@MX:NOTE` documenting why language-as-rules is the canonical decision.
5. **Quality gate**: Per `.claude/rules/moai/workflow/spec-workflow.md:172-204` Phase 0.5 Plan Audit Gate, plan-auditor will audit research + plan + acceptance + tasks before the implementation phase begins.
6. **No new lang skill creation under any circumstance**: The SPEC's content (forbid `moai-lang-*` skill creation) is itself a fitting test of the principle. Run phase MUST NOT create any `moai-lang-<name>/SKILL.md` file even as a placeholder.
7. **Cross-SPEC sync**: WF-005 is independent of WF-003 / WF-004 and may land in any order. Once WF-005 lands, future SPECs that propose new language support MUST follow REQ-WF005-009 (rule file under `.claude/rules/moai/languages/`, never a new skill).

---

## 10. Citation Summary (file:line anchors used)

1. `spec.md:1-229` (entire SPEC v0.2.0, contract source)
2. `.claude/rules/moai/languages/cpp.md:1-2,100,103` (paths frontmatter + dead refs in body)
3. `.claude/rules/moai/languages/csharp.md:1-2,110`
4. `.claude/rules/moai/languages/elixir.md:1-2,95`
5. `.claude/rules/moai/languages/flutter.md:1-2,94,95,97,98`
6. `.claude/rules/moai/languages/go.md:1-2`
7. `.claude/rules/moai/languages/java.md:1-2,121`
8. `.claude/rules/moai/languages/javascript.md:1-2,130`
9. `.claude/rules/moai/languages/kotlin.md:1-2,109,110`
10. `.claude/rules/moai/languages/python.md:1-2`
11. `.claude/rules/moai/languages/r.md:1-2,133`
12. `.claude/rules/moai/languages/ruby.md:1-2,131`
13. `.claude/rules/moai/languages/scala.md:1-2,131`
14. `.claude/rules/moai/languages/typescript.md:1-2,126`
15. `.claude/rules/moai/languages/php.md:1-2`
16. `.claude/rules/moai/languages/rust.md:1-2`
17. `.claude/rules/moai/languages/swift.md:1-2`
18. `.claude/rules/moai/development/skill-authoring.md:1-269`
19. `.claude/rules/moai/development/agent-authoring.md:165`
20. `.claude/skills/moai/workflows/run.md:338-352,979`
21. `.claude/skills/moai-foundation-core/modules/agents-reference.md:268,278-282,292`
22. `.claude/skills/moai-framework-electron/SKILL.md:19,228,230`
23. `.claude/skills/moai-platform-chrome-extension/SKILL.md:19,278,279`
24. `.claude/skills/moai-platform-auth/SKILL.md:20,225`
25. `.claude/skills/moai-platform-deployment/SKILL.md:398,399`
26. `.claude/skills/moai-workflow-loop/SKILL.md:148,149`
27. `.claude/skills/moai-domain-frontend/SKILL.md:119`
28. `.claude/skills/moai-domain-backend/SKILL.md:110`
29. `.claude/skills/moai-workflow-project/references/overview.md:120`
30. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:30,439,621,936`
31. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md:201,392`
32. `.claude/skills/moai-foundation-cc/reference/skill-examples.md:237,427`
33. `.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md:113`
34. `.claude/skills/moai-foundation-cc/references/examples.md:175`
35. `internal/template/commands_audit_test.go:1-60` (audit-test scaffold pattern)
36. `CLAUDE.local.md:§15` (16-language neutrality)
37. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)
38. `.moai/specs/SPEC-V3R2-WF-001/spec.md:1-30` (blocked-by SPEC, status: completed)

Total distinct file:line citations: **38** (exceeds the §Hard-Constraints minimum of 10 for research.md).

---

End of research.md.
