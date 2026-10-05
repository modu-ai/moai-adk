# acceptance.md — SPEC-MODEL-OPUS55-001

Document-level pin: every RED-now cell below was measured on tree `6e75b74db` (branch `WT-opus-55-default`) on 2026-09-23 unless a cell names another tree. Iter-2 cells were re-measured at HEAD `366ad4ee0`, whose non-SPEC tree content is identical to `6e75b74db` (`git diff --name-only 6e75b74db 366ad4ee0` lists only this SPEC's files + `.moai/reports/t1089/`). Commands run from the worktree root. "Release-blocking" ACs carry command + verbatim output + exit code; "regression-guard" ACs are green today by design and exist to catch a regression (their RED is supplied by a named mutant).

## §0 Traceability (REQ → AC)

| REQ | AC |
|-----|----|
| REQ-OP55-001 (opus → claude-opus-5-5) | AC-OP55-001, AC-OP55-012 |
| REQ-OP55-002 (claude-opus-5 → opus) | AC-OP55-002, AC-OP55-012 |
| REQ-OP55-003 (no claude-opus-5 constant) | AC-OP55-003 |
| REQ-OP55-004 (labels name Opus 5.5) | AC-OP55-004, AC-OP55-007, AC-OP55-008 |
| REQ-OP55-005 (medium recommended) | AC-OP55-007, AC-OP55-008 |
| REQ-OP55-006 (Opus 5.5 recommended) | AC-OP55-007, AC-OP55-008 |
| REQ-OP55-007 (effort empty option wording matches launch; (b) max delivery + operator precedence) | AC-OP55-007, AC-OP55-007a, AC-OP55-007b |
| REQ-OP55-008 (label-drift guard, both halves) | AC-OP55-009 |
| REQ-OP55-009 (canonical fact line + consistency) | AC-OP55-004, AC-OP55-006, AC-OP55-010 |
| REQ-OP55-010 (every high-default statement rewritten) | AC-OP55-004, AC-OP55-006, AC-OP55-006e |
| REQ-OP55-011 (heading + anchors) | AC-OP55-005 |
| REQ-OP55-012 (measured-on-Opus-5 attribution) | AC-OP55-004 |
| REQ-OP55-013 (no template effort key) | AC-OP55-011 |
| REQ-OP55-014 (matrix cells unchanged) | AC-OP55-013 |
| REQ-OP55-015 (historical surfaces untouched) | AC-OP55-014 |
| REQ-OP55-016 (neutrality + rebuild) | AC-OP55-015, AC-OP55-016 |

## §1 Scenarios (Given-When-Then)

### AC-OP55-001 — opus alias targets claude-opus-5-5 (REQ-OP55-001) — release-blocking

- Given the model-alias table
- When the `opus` alias is resolved
- Then it yields `claude-opus-5-5`, carried by a constant named for Opus 5.5
- Command: `grep -nE 'ModelIDOpus55 = "claude-opus-5-5"|"opus":[[:space:]]+ModelIDOpus55' internal/template/model_policy.go`
- RED-now: stdout empty, exit 1
- Green: two lines (the const declaration and the alias row), exit 0; plus AC-OP55-012 PASS, where the alias test asserts `ModelAliasCanonicalID("opus") == "claude-opus-5-5"`.

### AC-OP55-002 — claude-opus-5 normalizes to opus (REQ-OP55-002) — release-blocking

- Given a stored preference carrying `claude-opus-5`
- When the canonical-id → alias normalization runs
- Then it returns `opus`
- Command: `grep -nE '"claude-opus-5":[[:space:]]+"opus",[[:space:]]*// superseded' internal/template/model_policy.go`
- RED-now: stdout empty, exit 1
- Green: exactly one line, exit 0; plus a test in `internal/template` asserting `ModelAliasFromCanonicalID("claude-opus-5") == "opus"`, whose RED output (before the row exists) is recorded verbatim in progress.md §E.2.

### AC-OP55-003 — no named constant for claude-opus-5 (REQ-OP55-003) — release-blocking

- Command: `grep -rnw 'ModelIDOpus5' internal --include=*.go`
- RED-now (per-file counts via `-c`): `model_policy.go:4`, `glm_slot_test.go:1`, `model_policy_test.go:4`, `glm_slot_effort_test.go:1`, `launcher_test.go:3`, exit 0
- Green: stdout empty, exit 1. This also catches the stale `ModelIDOpus48` doc comment (D15).

### AC-OP55-004 — no unattributed Opus 5 on current-behavior surfaces (REQ-OP55-004, -009, -010, -012) — release-blocking

- Given the current-behavior surfaces — Go production code incl. `internal/template/templates/**`, web assets, wizard goldens, local `.claude/{rules,skills,agents,commands,output-styles}`, `.moai/config`, `.moai/project`, root `CLAUDE.md` and `AGENTS.md` — with historical surfaces outside the `find` roots (CHANGELOG.md, `.moai/specs`, `.moai/reports`, `.moai/research`, `.moai/docs`, `docs-site`, `README*.md`) and `*_test.go` excluded
- When probe P4 v2 runs
- Then it prints nothing: every remaining Opus 5 mention carries `(superseded)`, `// superseded`, or `measured on Opus 5`
- Command (single invocation): `find internal pkg cmd .claude/rules .claude/skills .claude/agents .claude/commands .claude/output-styles .moai/config .moai/project CLAUDE.md AGENTS.md -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.tmpl' -o -name '*.js' -o -name '*.golden' \) ! -name '*_test.go' -exec awk 'tolower($0) ~ /opus[ -]?5([^.0-9-]|\.[^0-9]|\.$|$)/ && tolower($0) !~ /\(superseded\)|\/\/ superseded|measured on opus 5/ {print FILENAME ":" FNR}' {} +`
- RED-now: 153 lines, exit 0 — verbatim sorted list `.moai/reports/t1089/p4v2-rednow-6e75b74db.txt`
- Green: stdout empty, exit 0. Because an empty stdout is also what a probe matching nothing prints, the green is read together with the positive control.
- Positive control (proves the probe still fires on the green tree): `awk 'tolower($0) ~ /opus[ -]?5([^.0-9-]|\.[^0-9]|\.$|$)/ {n++} END {print n+0}' internal/template/model_policy.go` → ≥ 1 (the `// superseded` deprecated-id row itself matches the base pattern).
- Known limit: a phrase shaped "Opus 5-era" is not matched (the `-` continuation is excluded so `claude-opus-5-5` stays silent).

### AC-OP55-005 — constitution heading and zone-registry anchors move together (REQ-OP55-011) — release-blocking

- Commands:
  - `grep -c 'opus-55-prompt-philosophy' .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md` — RED-now `:0` / `:0`; Green `:2` / `:2`
  - `grep -c 'opus-5-48-prompt-philosophy' .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md` — RED-now `:2` / `:2`; Green `:0` / `:0`
  - `grep -c '^## Opus 5.5 Prompt Philosophy' .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — RED-now `:0` / `:0`; Green `:1` / `:1`
  - `go test -count=1 -run 'TestRegistrySyncGuard|TestRegistrySyncMirrorsIdentical' -v ./internal/constitution/` — Green: both test names print `--- PASS`, no `[no tests to run]`.

### AC-OP55-006 — current Opus facts and medium default stated; no high-default claim left (REQ-OP55-009, -010) — release-blocking

- (a) Canonical fact line, probe P7 (single invocation): `awk '/^- opus = Opus 5\.5/ && /claude-opus-5-5/ && /2\.1\.280/ && /1M/ && /128K/ && /\$4/ && /\$20/ && /medium/ && /always on/ {n++} END {print n+0}' .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` — RED-now `0`, exit 0; Green `2` (one line per copy).
- (b) Positive medium statement on every rewritten surface: `grep -cE 'Opus 5\.5[^|]*medium|medium[^|]*Opus 5\.5' .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md .moai/project/tech.md` — RED-now `:0` for all nine files; Green ≥ 1 for each of the nine.
- (c) No high-default claim, probe P6: `grep -rnoE '`high` \(default\)|high: the default|defaults? to `effort: high`|default effort: high' .claude/rules .claude/skills internal/template/templates/.claude internal/template/templates/CLAUDE.md CLAUDE.md .moai/project` — RED-now 11 lines, exit 0 (plan.md §B.7); Green: stdout empty, exit 1.
- **AC-OP55-006e — effort-calibration sentence no longer reserves `medium` for speed-critical/simple work (plan.md §C.6 row 9b, plan-audit iter-2 N1).** RED-now measured at HEAD `084722b91`.
  - Negative: `grep -cE '^- \*\*Effort calibration\*\*:.*only for speed-critical or simple tasks' .claude/rules/moai/development/prompting-best-practices.md internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md` — RED-now stdout `.claude/rules/moai/development/prompting-best-practices.md:1` / `internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md:1`, exit 0; Green `:0` / `:0`, exit 1.
  - Positive: `grep -cE '^- \*\*Effort calibration\*\*:.*(Opus 5\.5[^.]*medium|medium[^.]*Opus 5\.5)' .claude/rules/moai/development/prompting-best-practices.md internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md` — RED-now `:0` / `:0`, exit 1; Green `:1` / `:1`, exit 0. Mutant: deleting the "only for …" clause without naming the Opus 5.5 default passes the negative half and fails this one.
- (d) No wrong-direction rewrite, probe P5: `grep -rnE 'Opus 5\.5[^.|]*default[^.|]*high' .claude/rules .claude/skills internal/template/templates/.claude internal/template/templates/CLAUDE.md CLAUDE.md .moai/project` — now: no output, exit 1 (regression-guard); mutant "Opus 5.5 defaults to `effort: high`" prints one line, exit 0.

### AC-OP55-007 — web labels, recommendation, and honest empty-option wording (REQ-OP55-004, -005, -006, -007) — release-blocking

Each command runs on `internal/web/assets/i18n.js`; per-locale strings are pinned so one locale cannot satisfy another's check.

- Model label, English-unified (plan.md §C.8): `grep -c '"f.model.opt.opus\[1m\]": "Opus 5.5 (Recommended)"' internal/web/assets/i18n.js` — RED-now `0`; Green `4`. Companion: `grep -c '"f.model.opt.opus\[1m\]": "Opus 5"' internal/web/assets/i18n.js` — RED-now `4`; Green `0`.
- Medium recommended, one command per locale, each Green `1`, RED-now `0`: `grep -c '"f.effort_level.opt.medium": "Medium (Recommended)"' internal/web/assets/i18n.js`; `grep -c '"f.effort_level.opt.medium": "중간 (권장)"' internal/web/assets/i18n.js`; `grep -c '"f.effort_level.opt.medium": "中 (推奨)"' internal/web/assets/i18n.js`; `grep -c '"f.effort_level.opt.medium": "中 (推荐)"' internal/web/assets/i18n.js`.
- Empty-option wording names the fallback order (plan.md §C.7), one command per locale, each Green `1`, RED-now `0`: `grep -cE '"opt\.runtime_default": "[^"]*model policy[^"]*Opus 5\.5[^"]*"' internal/web/assets/i18n.js`; the same with `모델 정책`, `モデルポリシー`, and `模型策略` in place of `model policy`.
- Behavioral pin keeping the label true: `go test -count=1 -run 'TestResolveLaunchEffort' -v ./internal/cli/` — Green: `--- PASS: TestResolveLaunchEffort` with its five subtests (`model_policy fallback high` → `high` among them), unchanged by this SPEC.
- Web test updated in M2: `go test -count=1 -run 'TestModelOptLabelsEnglishUnified' -v ./internal/web/` — Green `--- PASS`.
- **AC-OP55-007a — `max` never in settings, travels as `--effort max` (REQ-OP55-007 (b)) — regression-guard.** Measured at HEAD `0512e6e5f` (amendment 0.1.2).
  - Command: `go test -count=1 -run 'TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection|TestLaunchEffortMaxTravelsAsArgvOnKanbanInjection|TestLaunchEffortXHighStaysOnSettingsPath' -v ./internal/cli/`
  - Now (GREEN, since `eb629efb5`): `--- PASS: TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection`, `--- PASS: TestLaunchEffortMaxTravelsAsArgvOnKanbanInjection`, `--- PASS: TestLaunchEffortXHighStaysOnSettingsPath` (verbatim in `.moai/reports/t1089/f2-probe-rednow-0512e6e5f.txt`).
  - Mutant that turns it RED: delete the `if effort == template.EffortLevelMax { … }` early return in `applyLaunchEffort` (internal/cli/launch_effort_settings.go) so `max` falls through to `payload[effortSettingsKey] = effort` → both `MaxTravelsAsArgv…` tests fail ("max must never be written to the settings payload"). Opposite-direction mutant: route every level to argv → `TestLaunchEffortXHighStaysOnSettingsPath` fails.
- **AC-OP55-007b — operator `--effort` anywhere in argv suppresses injection (REQ-OP55-007 (b)) — release-blocking.** Measured at HEAD `0512e6e5f` (amendment 0.1.2); the sync-audit F1 fix in run flips it.
  - Given a profile resolving `max`, When the operator argv is each of `[--effort low]`, `[--effort=low]`, `[-- --effort low]`, `[-- --effort=low]`, Then the general-path argv carries exactly one `--effort` token (the operator's) and the kanban injection adds zero `--effort` tokens.
  - RED-now command (tree untouched, probe supplied by overlay): `go test -overlay <scratch>/overlay.json -count=1 -run 'TestZZF2ProbeOperatorEffortAnywhere' -v ./internal/cli/`, probe source `.moai/reports/t1089/f2-probe-zz_f2_probe_test.go.txt`
  - RED-now output (exit 1; the two before-`--` cases pass, the two after-`--` cases fail — F1):
    ```
    general op=[-- --effort low] argv=[-- --effort low --effort max] effortFlags=2 want 1
    kanban op=[-- --effort low] injected=[--settings …/moai-kanban-….json --effort max] effortFlags=1 want 0
    general op=[-- --effort=low] argv=[-- --effort=low --effort max] effortFlags=2 want 1
    kanban op=[-- --effort=low] injected=[--settings …/moai-kanban-….json --effort max] effortFlags=1 want 0
    --- FAIL: TestZZF2ProbeOperatorEffortAnywhere
    ```
  - Green: the run-phase F1 fix commits an equivalent in-tree test covering the same four argv shapes on both paths; `go test -count=1 -run '<that test>' -v ./internal/cli/` prints `--- PASS`, and the overlay probe above also passes.

### AC-OP55-008 — TUI wizard labels and recommendation (REQ-OP55-004, -005, -006) — release-blocking

Anchored on the quoted field value (a comma inside the label is allowed); one command per locale marker so each locale is proved separately.

- `grep -cE 'EffortLevelMedium:[[:space:]]+"[^"]*\(Recommended\)"' internal/cli/profile_setup_translations.go` — RED-now `0`; Green `1`. Same form with `\(권장\)`, `\(推奨\)`, `\(推荐\)` — each RED-now `0`, Green `1`.
- `grep -cE 'ModelOpus1M:[[:space:]]+"[^"]*Opus 5\.5[^"]*Recommended[^"]*"' internal/cli/profile_setup_translations.go` — RED-now `0`; Green `1`. Same form with `권장`, `推奨`, `推荐` — each RED-now `0`, Green `1`.

### AC-OP55-009 — label-drift guards reject stale labels (REQ-OP55-008) — release-blocking (mutant-proved)

Three mutants, each applied to the green tree alone, each reverted before the next; evidence (mutant diff, verbatim failing output, reverted-tree PASS) recorded in progress.md §E.2.

- **AC-OP55-009a (guard A — profile-setup labels).** Mutant M-a: in `internal/cli/profile_setup_translations.go`, change one locale's `ModelOpus` label from "Opus 5.5" to "Opus 5 ". Command: `go test -count=1 -run 'TestGetProfileText_OpusAliasValues' ./internal/cli/` → FAIL naming that locale. Reverted tree → `ok`.
- **AC-OP55-009b (guard B — wizard model-policy descriptions).** Mutant M-b: in `internal/cli/wizard/translations.go`, change one locale's model-policy description from "Opus 5.5" to "Opus 5 ". Command: `go test -count=1 -run 'TestModelPolicyDescsAgreeWithProfileMatrix' ./internal/cli/wizard/` → FAIL naming that locale. Reverted tree → `ok`.
- **AC-OP55-009c (both guards follow an alias change).** Mutant M-c: point the `opus` alias at a hypothetical `claude-opus-6` (labels unchanged). Both commands above (`TestGetProfileText_OpusAliasValues`, `TestModelPolicyDescsAgreeWithProfileMatrix`) FAIL, and the derived version token printed in the failure is `6` (hyphen→dot derivation, plan.md §E K2). Reverted tree → both `ok`.

### AC-OP55-010 — local/template pair identity preserved (REQ-OP55-009) — regression-guard

- Commands (each single-invocation; exit 0 = identical): `cmp .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md`, and the same `cmp` for `rules/moai/development/model-policy.md`, `rules/moai/development/prompting-best-practices.md`, `rules/moai/workflow/context-window-management.md`, `rules/moai/workflow/dynamic-workflows.md`, `skills/moai-foundation-thinking/SKILL.md`
- Now: all six exit 0 (identical). Green: all six still exit 0 after edits. Mutant: editing only one side makes `cmp` exit 1.

### AC-OP55-011 — no effort key in the settings template (REQ-OP55-013) — regression-guard

- Command: `grep -n -i effort internal/template/templates/.claude/settings.json.tmpl`
- Now: stdout empty, exit 1. Green: stdout empty, exit 1. Mutant: adding `"effortLevel": "medium"` makes it print one line, exit 0.

### AC-OP55-012 — affected packages pass (all REQs) — release-blocking

- Commands (serial, one at a time):
  - `go test -count=1 ./internal/template/ ./internal/cli/wizard/ ./internal/web/ ./internal/settings/ ./internal/constitution/`
  - `go test -count=1 -timeout 25m ./internal/cli/` under `moai slot acquire --resource go-test-cli --max-duration 30m` … `moai slot release --resource go-test-cli`
- Green: every package line `ok`, none `[no test files]`/`[no tests to run]`, exit 0.

### AC-OP55-013 — per-agent profile matrix unchanged (REQ-OP55-014) — regression-guard

- Command: `git diff develop...HEAD -- internal/template/profile_matrix.go`
- Green: the diff touches comment lines only (no line inside the `defaultProfileMatrix` literal changes); read by the auditor. Mutant: changing any `{model, effort}` cell shows a non-comment hunk.

### AC-OP55-014 — historical surfaces untouched (REQ-OP55-015) — regression-guard

- Command: `git diff --name-only develop...HEAD -- CHANGELOG.md docs-site README.md README.ko.md README.ja.md README.zh.md .moai/research .moai/docs .moai/reports ':!.moai/reports/t1089'`
- Green: stdout empty. (Three-dot form measures from the merge-base with local `develop`, so absorbed develop commits do not read as this card's changes; this card's own evidence directory is excluded by the pathspec.)
- Companion: `git diff --name-only develop...HEAD -- .moai/specs` lists only `.moai/specs/SPEC-MODEL-OPUS55-001/*`.

### AC-OP55-015 — template neutrality holds (REQ-OP55-016) — release-blocking at run close

- Command: `go test -count=1 -run 'TestTemplateNoInternalContentLeak|TestLanguageNeutrality|TestLeakClassNoDateShaInDefaultTier' -v ./internal/template/`
- Green: the three test names print `--- PASS`, exit 0.

### AC-OP55-016 — embedded template rebuilt; agent emission consistent (REQ-OP55-016) — release-blocking at run close

- Commands: `make build` → exit 0; `git diff --name-only develop...HEAD -- internal/template/templates/.claude/agents` → empty (no agent file changed, so `make agents-emit` is not required). If that list is non-empty: `make agents-emit-check` → exit 0.

## §2 Edge cases

- "Opus 5." at a sentence end must still be caught (P4's `\.$` / `\.[^0-9]` alternatives).
- "claude-opus-5-5" and "Opus 5.5" must never be caught (continuation exclusion).
- "Opus5" is caught (P4 v2 `opus[ -]?5`).
- Wizard migration of `claude-opus-5[1m]`: the `[1m]` suffix is split off before normalization in the launcher path; run phase confirms the deprecated-row lookup sees the bare id.
- The three golden files are regenerated with `-update-golden`, not hand-edited.

## §3 Quality gate

- Tier M plan-auditor threshold 0.80.
- Run phase: zero new lint findings (`golangci-lint run ./internal/template/... ./internal/cli/... ./internal/web/...`), coverage of `internal/template` not lowered.

## §4 Definition of Done

All sixteen ACs green with evidence recorded in progress.md §E.2 (command + verbatim output + exit code + tree SHA); AC-OP55-009a/b/c mutant evidence present; follow-up docs card issued by the lead (plan.md §C.3).
