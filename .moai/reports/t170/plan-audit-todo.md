# SPEC Review Report: SPEC-TODO-ENABLE-FLAG-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.78** (harmonic mean) — graded against the **Tier M PASS threshold 0.80** (`spec-workflow.md` § SPEC Complexity Tier)

Reasoning context ignored per M1 Context Isolation. The task message's measurements were treated as claims to re-verify, not as inputs.

Tree: `git rev-parse --short HEAD` → `e7fb0e1d2` — matches the pinned commit.

Counts re-measured independently: `grep -c '^### REQ-' spec.md` → **6**; `grep -c '^### AC-T-' acceptance.md` → **11**. Both within Tier M ceilings (16/16).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `spec.md:84,92,102,108,117,123` yield REQ-1…REQ-6, sequential, no gaps, no duplicates, consistent formatting.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in `spec.md`), not the AC layer. Each REQ carries a condition clause + `shall`/`shall not` modality: REQ-1 `spec.md:86` "키가 설정에 없는 경우 … 해석돼야 한다" + `:90` "만들어서는 안 된다(shall not)" (unwanted form); REQ-2 `:94` "…`false`인 경우, 아래 표면은 … 출력해서는 안 된다(shall not)" (state-driven + negative); REQ-3 `:104`; REQ-4 `:110` "`moai init`이 대화형으로 실행되는 경우" (event-driven); REQ-5 `:119`; REQ-6 `:125`. The 11 Given-When-Then entries in `acceptance.md` §D.2 are verification-layer `AC-XXX` and are correctly NOT graded here (M3 § Scope) — they are graded under Group 4 below.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names, verified field-by-field at `spec.md:2-15`: `id`, `title`, `version`, `status: draft`, `created: 2026-08-22`, `updated: 2026-08-22`, `author`, `priority: High`, `phase: "v3.1.3"`, `module: "internal/config"`, `lifecycle: spec-anchored`, `tags`. No rejected snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`) present. `phase` is a release target, not a prohibited lifecycle token. Optional `tier: M`, `era: V3R6` present and valid. (One minor deviation noted as D8 below.)
- **[PASS] MP-4 language neutrality** — the SPEC touches template-bound content (`internal/template/templates/.claude/skills/moai/SKILL.md`, optionally the template `workflow.yaml`) but names **no** language-specific tool. Measured: `grep -rn 'SPEC-TODO-ENABLE-FLAG\|REQ-' internal/template/templates/.moai/config/sections/workflow.yaml internal/template/templates/.claude/skills/moai/SKILL.md | wc -l` → `0` today, and AC-T-010 (`acceptance.md:150`) pins that expectation. Neutrality obligation is explicitly carried at `spec.md:127`.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — extracted refs: `SPEC-FEEDBACK-AUTO-SUBMIT-001`, `SPEC-KANBAN-TODO-CLI-001`. Both exist under `.moai/specs/`. Measured statuses: `status: draft` (sibling, `tier: L`) and `status: in-progress`. Neither is `retired`/`superseded`/`archived` ⇒ no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. D8 auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-ENABLE-FLAG-001/` → no match. No `research.md` (Tier M, correctly 3 artifacts).

**No must-pass failure.** The FAIL verdict below rests entirely on the aggregate score falling under the Tier M threshold.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are individually unambiguous; two carry implementation detail into §C (`spec.md:114` names `saveBoolAnswer`/`WritePhase1Configs`/`yamlpatch.PatchFile`; `:119` embeds the verbatim Go call). One genuine ambiguity: REQ-2 #3 (`:98`) vs REQ-3 (`:104`) leave explicit `/moai todo` invocation-under-flag-off undefined (D2). |
| Completeness | 0.80 | 0.75–1.0 | All sections present: HISTORY `§G:191`, WHY `§A:21`, WHAT `§B:51`, REQUIREMENTS `§C:82`, AC in `acceptance.md` (Tier M), Out of Scope 4× H3 with bullets at `spec.md:64,70,76,79`. Frontmatter complete. Gaps: malformed-value semantics (D3), sibling-conflict resolution procedure (D4). |
| Testability | 0.65 | 0.50–0.75 | 9 of 11 ACs name one command and one binary expectation. Two do not: AC-T-009 names a test that does not exist (D1, vacuous half) and asserts a `Then` no named command observes; AC-T-004's third command is a `git diff` that returns empty post-commit (D5). |
| Traceability | 1.00 | 1.0 | Every REQ has ≥1 AC and every AC names a REQ that exists: REQ-1→AC-T-001; REQ-2→002/003/004; REQ-3→005; REQ-4→006/007/008; REQ-5→009; REQ-6→010; AC-T-011→전체. Matrix at `acceptance.md:12-24`. Zero orphans, zero uncovered REQs. |

Aggregate = harmonic mean(0.75, 0.80, 0.65, 1.00) = 4 / (1.3333 + 1.2500 + 1.5385 + 1.0000) = **0.781**.

---

## Verification of the five directed checks

**Check 1 — `internal/settings/schema_sections.go:334` is the one-line exemplar. CONFIRMED.**
`grep -n 'branch_guard", "enabled"'` → `334:  s(SectionWorkflow, "workflow", TypeBool, "workflow", "branch_guard", "enabled"),`. The adjacent comment `:330-333` independently documents the exact situation the SPEC relies on: *"the distributed template ships without a branch_guard block, so this key is absent until the user opts in… The web console renders it from this FieldDef via schemaform.go; the seam writer (yamlpatch) upserts the nested mapping on first edit."* `workflow` is already both seam-writable and web-rendered (lens §B.5, `sectionroute.go:55-99` × `consoleTabs()`), so no handler or template edit is required. The SPEC's claim at `:119` and `:125` is accurate.

**Check 2 — `readMCPToolEnablement` `*bool` default-ON pattern. CONFIRMED, with one divergence the SPEC does not name.**
The function exists at `internal/cli/mcp_server.go`; it seeds a map to `true`, unmarshals into an anonymous struct carrying `Enabled *bool \`yaml:"enabled"\``, and applies only non-nil pointers. Its doc comment states the fail-OPEN truth table verbatim, and the `@MX:NOTE` records the fail-open choice deliberately — the SPEC's "태도까지 계승한다" (`plan.md:51`) is grounded.

A plain `bool` genuinely cannot express the requirement: `yaml.Unmarshal` leaves the zero value `false` for both "key absent" and "explicitly false", so absent and false become indistinguishable. The `*bool` is a real necessity, not a stylistic preference. The shipped template will not carry the block — measured `grep -c 'todo' internal/template/templates/.moai/config/sections/workflow.yaml` → `0`.

**Divergence (not a defect, but unstated):** `readMCPToolEnablement`'s own comment says it *deliberately avoids* `config.Loader.Load` ("stays a small hand-rolled read… should touch exactly one file with no defaults/env-override machinery"). `plan.md` M1 (`:68-70`) routes the new key through the typed loader instead (`WorkflowConfig` + `NewDefaultWorkflowConfig`). The SPEC inherits the *shape* but not the *mechanism*, and does not say so. This matters for D3 below.

**Check 3 — the two suppressible surfaces are already conditional. CONFIRMED at the exact cited lines.**
`internal/hook/session_start_kanban.go:180` → `context = append(context, fmt.Sprintf(m.backlogSummary, queuedBacklogCount(root)))`, reached only inside the kanban-env + `source == "startup"` path. `internal/statusline/renderer.go:188` → `if r.isSegmentEnabled(SegmentBacklog) && data.Backlog.Available {`. Both are one-line guard sites exactly as REQ-2 describes.

**Check 4 — the two NON-suppressible surfaces. CONFIRMED. No missed mechanism; this is NOT a MUST-FIX.**
`.claude/rules/moai/workflow/kanban-dispatch.md:5` states verbatim: *"**Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach it."* The only rule-scoping mechanism in this repo is the frontmatter `paths:` restriction, which is a path predicate, not a config predicate; a search of `internal/config` and `internal/settings` for any rule- or skill-disabling key returned nothing. `.claude/skills/moai/SKILL.md` carries `/moai todo` in listing metadata at `:6`, `:81`, `:105`, `:166-172` — measured and matching the SPEC's citation exactly.

**I searched specifically for the mechanism the SPEC claims does not exist, and did not find one.** The scope boundary at `spec.md:35-42` is therefore sound, and its [HARD] prohibition on writing an unachievable AC (`:42`) is correct.

**Check 5 — the surface count. The enumeration is complete for in-session surfaces; the SPEC does not say "in-session".**
The lens enumerates nine (`lens-web-todo.md` §B.4) and I confirmed each citation. My independent sweep found live guidance the nine do not cover: `README.md` + 3 locale variants, and `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md` plus `advanced/kanban-mode.md`, `core-concepts/kanban-board-terms.md`, `_index.md` — 16 docs-site pages. These are published documentation, not session-context surfaces, so excluding them is defensible; but `spec.md:35` says flatly "todo 표면 9개" with no qualifier. See D7 (optional).

---

## Defects Found (structured defect-list)

**D1. AC-T-009 names a test that does not exist, and asserts a `Then` no named command observes** — `.moai/specs/SPEC-TODO-ENABLE-FLAG-001/acceptance.md:130-138` — Severity: **critical** — Class: **blocking** — The AC's command is `go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestSchemaLabel'`. Measured: `grep -n '^func Test' internal/web/schema_label_test.go` → `TestSchemaEmptyLabelParity` (`:16`), `TestI18nKeySetParity` (`:74`), `TestI18nSegmentKeysRemovedFromWebDictionary` (`:133`). **No function name contains the substring `TestSchemaLabel`**, so the `internal/web` half matches zero tests and exits 0 with "no tests to run" — a vacuous pass. Separately, the AC's `Then` ("존재하고 Type이 TypeBool, Persist.Kind가 PersistSeam이다") is not asserted by either named test: `TestSchemaCurrentValuesReadsAllSections` (`internal/settings/schema_sections_test.go:441`) checks a hardcoded `cases` map of 13 unrelated keys and is entirely unaffected by a new field. The i18n half of the claim IS mechanically guarded, but by `TestI18nKeySetParity`, which iterates `settings.AllFields()` and requires `.title`/`.desc` in all 4 locales — the SPEC's prose at `spec.md:121` cites `internal/web/schema_label_test.go:96`, a line that falls inside `TestI18nKeySetParity` (74–132), so the SPEC located the right guard and then named the wrong function in the AC. **Required fix:** replace `TestSchemaLabel` with `TestI18nKeySetParity`, and add a named deliverable test (e.g. `TestWorkflowTodoEnabledFieldRegistered`) that asserts the field's presence, `TypeBool`, and `PersistSeam` — otherwise the `Then` clause has no observation.

**D2. REQ-2 #3 and REQ-3 leave the user-visible behavior of an explicit `/moai todo` invocation undefined** — `spec.md:98` vs `spec.md:104` — Severity: **major** — Class: **blocking** — REQ-3 keeps the `moai todo` CLI registered and working when the flag is `false`. REQ-2 #3 says the flag suppresses routing to `workflows/todo.md`. The SPEC does not say what happens when a user — who, per the SPEC's own residual-risk statement at `:178`, *still sees `/moai todo` in the skill listing* — types `/moai todo` explicitly. Is the slash surface refused, or honored? The two requirements point opposite ways, and the asymmetry (CLI works, slash may not) is exactly the state the user will hit. AC-T-004 does not resolve it: it greps for the *presence of a condition sentence*, never its semantics. This is the concrete form of the lead's question "does the SPEC say what that user actually experiences?" — for surfaces 5/6/9 it does (§E.3), for the slash entry point it does not. **Required fix:** state in REQ-2 #3 whether an explicit `/moai todo` invocation is honored (recommended, consistent with REQ-3's "the flag suppresses guidance, not the feature") or refused, and extend AC-T-004 with an observation of that behavior rather than of the sentence's existence.

**D3. Malformed-value behavior is unspecified and untested — and its blast radius is the whole `workflow` section** — `spec.md:86-88`, `acceptance.md:30-44` — Severity: **major** — Class: **optional** — I traced all five states the lead named. Absent → `Todo.Enabled` nil → `true` (AC-T-001 case 1). `false` → `false`. `true` → `true`. Existing project upgrading with no block → template ships nothing (measured `grep -c 'todo' …/templates/…/workflow.yaml` → `0`) → absent → `true`. **Malformed → also `true`, but by a path the SPEC never names:** `internal/config/loader.go:226-237` `loadWorkflowSection` seeds the wrapper with `cfg.Workflow` defaults, and on unmarshal error emits `slog.Warn("failed to load workflow config, using defaults")` and returns — so a bad value anywhere in the block silently discards **every** user-set `workflow.*` key (branch_guard, worktree, loop_prevention, audit…) and reverts them to construction-time defaults. **No path reads OFF for existing users, so the regression the AC set had to catch does not exist** — but the SPEC asserts it inherits `readMCPToolEnablement`'s fail-open posture (`plan.md:51`) while actually routing through the typed loader, whose fail-open has a far larger blast radius. **Required fix:** add one line to REQ-1 stating the malformed-value outcome (falls back to defaults ⇒ active) and note the section-wide fallback; optionally add a 4th case to AC-T-001.

**D4. The sibling merge discipline asserts an outcome it cannot guarantee and names no resolution owner** — `spec.md:164` — Severity: **major** — Class: **blocking** — The [HARD] clause requires "어느 쪽이 먼저 착지하든 나중 것이 텍스트 충돌 없이 얹혀야 한다". Both SPECs append a new entry to the *same struct literal at the same location* in `internal/cli/wizard/questions.go` (Page3 "Quality & Workflow" group), `types.go`, `wizard.go` `saveBoolAnswer`, and three locale blocks in `translations.go`. Adjacent-line insertions conflict in git regardless of whether the inserted items differ — "add different items only" prevents *semantic* clash, not *textual* conflict. So the clause states an expectation as if it were a mechanism. AC-T-011 (`acceptance.md:155-167`) verifies the post-merge state but specifies no procedure for reaching it. Compounding this: `.claude/rules/local/repo-local-pr-policy.md` (verified present) puts **all tiers** on Route B in this repo, so the two SPECs land as two PRs against the same nine files — the conflict is the expected case, not the exception. **Required fix:** replace "충돌 없이 얹혀야 한다" with a resolution rule — name the second lander as the conflict owner, require keeping both entries, forbid reordering during resolution, and re-run AC-T-011 after resolution.

**D5. AC-T-004's third command is vacuous after commit and requires judgment to evaluate** — `acceptance.md:87` — Severity: **major** — Class: **blocking** — `git diff --unified=0 .claude/skills/moai/SKILL.md | grep '^[-+].*todo'` compares the working tree against the index/HEAD, so once the M3 change is committed it returns empty and the AC passes without observing anything. Its stated expectation ("라우팅 절 외 목록 줄 변경 0건") also requires a human to classify each diff line as routing-clause vs listing-line — not a single-observation binary decision. **Required fix:** pin the expectation to the actual line content instead of to a diff, e.g. assert the three listing lines are byte-identical to their pre-change values (`sed -n '6p;81p;105p' … | shasum`, expected value recorded in the AC), or diff against the merge-base rather than the working tree.

**D6. `plan.md` states the wrong PR route for this repository** — `plan.md:11` — Severity: **major** — Class: **blocking** — "Tier M이므로 Route A(main-direct)가 기본이나…". `.claude/rules/local/repo-local-pr-policy.md` is [HARD] and explicit: *"In THIS repository, the `spec-workflow.md` Route A … is DISABLED"* because `main` carries `enforce_admins: true` + required PR; **all** tiers use Route B. A run-phase agent following `plan.md:11` would attempt a direct push to `main` and be rejected by branch protection. The trailing hedge about batch-lane discipline does not correct the primary statement. **Required fix:** replace with "Route B (PR) — repo-local policy disables Route A for all tiers".

**D7. "todo 표면 9개" is unqualified** — `spec.md:35` — Severity: **minor** — Class: **optional** — The nine are in-session surfaces (verified against `lens-web-todo.md` §B.4, every citation checked). Published documentation also describes todo and is not suppressed by the flag: `README.md` + 3 locales, and 16 `docs-site/content/{en,ko,ja,zh}/…` pages including a dedicated `utility-commands/moai-todo.md` per locale. The lens itself flags this scope choice in its Gaps section. **Required fix:** qualify as "런타임/세션 표면 9개", and optionally add published docs to §B Out of Scope so the boundary reads complete.

**D8. `version` frontmatter value is unquoted** — `spec.md:4` — Severity: **minor** — Class: **optional** — `version: 0.1.0`; `spec-frontmatter-schema.md` § Field Reference specifies "semver `X.Y.Z`, quoted". It decodes as a string regardless (two dots make it non-numeric), so MP-3 is not failed and no lint finding results. **Required fix:** quote it (`version: "0.1.0"`) for schema conformance.

**D9. Two `file:line` citations in the SPEC have drifted** — `spec.md:49`, `spec.md:112` — Severity: **minor** — Class: **optional** — `spec.md:49` cites `translations_completeness_test.go:89`; `TestWizardQuestionTranslationCompleteness` is at `:95`. `spec.md:112` (and `plan.md:18`) cite `questions_test.go:101` for `TestQuestionOrder`; it is at `:87`. The claims themselves are correct — I verified `DefaultQuestions` asserts exactly 5 IDs (`questions_test.go:98-106`) and `ReconfigureQuestionsOrder` asserts exactly 12 (`:187-199`). Note the AC's `-run 'TestReconfigureQuestions'` is an unanchored regex and **does** match `TestReconfigureQuestionsOrder`, so that pattern is not vacuous. **Required fix:** correct the two line numbers.

---

## What the SPEC gets right (cited, so a revision does not regress it)

- **The scope boundary is stated where a reader will hit it, repeatedly and early.** `spec.md:33` (§A.1 correction P4, before Scope), `:42` ([HARD] no-unachievable-AC), `:64-68` (first Out of Scope H3), `:178` (residual risk), `acceptance.md:5` (document header), `:174` (§D.3), `:184` (closure gate forbids "todo 안내를 전부 껐다"), `plan.md:21` + AP-6/7/8. It is not buried.
- **No AC promises the unachievable version.** I read all 11. AC-T-002's suppression assertion is scoped to SessionStart hook output; AC-T-004's second assertion only verifies the listing was *not touched*. `acceptance.md:174` states explicitly that out-of-scope surfaces are neither required to disappear nor required to remain.
- **Control cases defeat degenerate passes.** AC-T-002 and AC-T-003 each carry a "remove the key ⇒ it reappears" clause (`acceptance.md:53-56`, `:66-68`), which is what stops a test from passing because the output was never produced. AC-T-003's third case (`:69-71`) proves the pre-existing `statusline.yaml backlog: false` path survives — the OR semantics of `plan.md` T4 are correct (AND of two enable-predicates = OR of two off-switches).
- **AC-T-008 defends against the dead-code trap.** The SPEC identifies that the card's own cited wiring precedent (`applyAutonomyTierFromWizard`) is dead (`spec.md:71`) and refuses to follow it, then makes the live path falsifiable by observing the file write rather than the assignment (`acceptance.md:119-128`).

---

## On the safety of the split (flagged as requested)

The split is **thematically sound but operationally under-protected**. The seam is todo-vs-feedback, which is a clean requirement boundary, and each SPEC is independently completable — so the recorded `depends_on` rationale (`spec.md:166`, "기능 의존이 없다") is *literally* correct.

But that rationale evaluates only the functional axis. Declaring `depends_on` would have had a second effect the SPEC does not weigh: the Phase 1 Depends_on Pre-flight blocks run-phase entry until the dependency reaches `status: completed`, which would have **serialized** the two SPECs and thereby dissolved the shared-file hazard entirely. Omitting it buys concurrency and pays for it in merge conflicts across nine files (D4). Under this repo's all-tiers-PR policy those are two concurrent PRs on the same files. The sibling is `tier: L` and this one `tier: M`, so they also carry different artifact sets and different audit thresholds while sharing a wizard surface whose count-fixed tests (5 / 12) bind both.

That is a defensible trade — but it is currently recorded as "no dependency exists" rather than as "we chose concurrency over serialization and here is how conflicts are resolved". Fixing D4 fixes the split's safety; the split itself does not need undoing.

---

## Recommendation

FAIL at 0.78 against the Tier M threshold of 0.80. All seven must-pass criteria pass; the shortfall is Testability (0.65), driven by one vacuous AC and one AC that cannot fail after commit. Fix in this order:

1. **D1** — `acceptance.md:130-138`: replace `TestSchemaLabel` with `TestI18nKeySetParity`, and add a named deliverable test asserting the new field's presence + `TypeBool` + `PersistSeam`. Without this, AC-T-009 passes while observing nothing on the `internal/web` side.
2. **D5** — `acceptance.md:87`: make the "listing unchanged" expectation a content assertion (hash or exact-line compare against recorded pre-change values), not a working-tree `git diff`.
3. **D6** — `plan.md:11`: Route B, all tiers, per `.claude/rules/local/repo-local-pr-policy.md`.
4. **D2** — `spec.md:98`/`:104`: define the explicit `/moai todo` invocation behavior under flag-off, and give AC-T-004 an observation of that behavior.
5. **D4** — `spec.md:164`: replace the "충돌 없이" assertion with a conflict-resolution rule naming the second lander as owner.
6. **D3** (optional) — `spec.md:86-88`: one line on malformed-value outcome + the section-wide default fallback.
7. **D7 / D8 / D9** (optional) — qualify the surface count; quote `version`; correct the two drifted line citations.

Items 1–5 are the blocking set. Items 6–8 are surfaced for the orchestrator's discretion and do not, on their own, justify holding the SPEC — a long optional list must not be used to manufacture a FAIL (M6).

Iteration 2 should be scoped to this enumerated defect delta plus a regression check over it, not a from-scratch re-audit.
