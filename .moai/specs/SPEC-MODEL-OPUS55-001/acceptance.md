# acceptance.md — SPEC-MODEL-OPUS55-001

Document-level pin: every RED-now cell below was measured on tree `6e75b74db` (branch `WT-opus-55-default`) on 2026-09-23 unless a cell names another tree. Commands run from the worktree root. "Release-blocking" ACs carry command + verbatim output + exit code; "regression-guard" ACs are green today by design and exist to catch a regression (their RED is supplied by a named mutant).

## §0 Traceability (REQ → AC)

| REQ | AC |
|-----|----|
| REQ-OP55-001 (opus → claude-opus-5-5) | AC-OP55-001, AC-OP55-012 |
| REQ-OP55-002 (claude-opus-5 → opus) | AC-OP55-002, AC-OP55-012 |
| REQ-OP55-003 (no claude-opus-5 constant) | AC-OP55-003 |
| REQ-OP55-004 (labels name Opus 5.5) | AC-OP55-004, AC-OP55-007, AC-OP55-008 |
| REQ-OP55-005 (medium recommended) | AC-OP55-007, AC-OP55-008 |
| REQ-OP55-006 (Opus 5.5 recommended) | AC-OP55-007, AC-OP55-008 |
| REQ-OP55-007 (effort empty option wording) | AC-OP55-007 |
| REQ-OP55-008 (label-drift guard) | AC-OP55-009 |
| REQ-OP55-009 (rule prose facts, both copies) | AC-OP55-004, AC-OP55-006, AC-OP55-010 |
| REQ-OP55-010 (effort-default rewrite) | AC-OP55-004, AC-OP55-006 |
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
- Green: two lines (the const declaration and the alias row), exit 0; plus AC-OP55-012 test run PASS (behavioral proof: the alias test asserts `ModelAliasCanonicalID("opus") == "claude-opus-5-5"`).

### AC-OP55-002 — claude-opus-5 normalizes to opus (REQ-OP55-002) — release-blocking

- Given a stored preference carrying `claude-opus-5`
- When the canonical-id → alias normalization runs
- Then it returns `opus`
- Command: `grep -nE '"claude-opus-5":[[:space:]]+"opus"' internal/template/model_policy.go`
- RED-now: stdout empty, exit 1
- Green: exactly one line, exit 0; plus a test in `internal/template` asserting `ModelAliasFromCanonicalID("claude-opus-5") == "opus"`, whose RED output (before the row exists) is recorded verbatim in progress.md §E.2.

### AC-OP55-003 — no named constant for claude-opus-5 (REQ-OP55-003) — release-blocking

- Command: `grep -rnw 'ModelIDOpus5' internal --include=*.go`
- RED-now (counts per file via `-c`): `model_policy.go:4`, `glm_slot_test.go:1`, `model_policy_test.go:4`, `glm_slot_effort_test.go:1`, `launcher_test.go:3`, exit 0
- Green: stdout empty, exit 1

### AC-OP55-004 — no unattributed Opus 5 on current-behavior surfaces (REQ-OP55-004, -009, -012) — release-blocking

- Given the current-behavior surfaces (Go production code incl. `internal/template/templates/**`, web assets, wizard goldens, local `.claude/rules`, `.claude/skills`, `.moai/config`, root `CLAUDE.md`), with historical surfaces excluded by path (CHANGELOG.md, `.moai/specs`, `.moai/reports`, `.moai/research`, `.moai/docs`, `docs-site`, `README*.md` are outside the `find` roots) and `*_test.go` excluded
- When probe P4 runs
- Then it prints nothing: every remaining Opus 5 mention carries `superseded` or `measured on Opus 5`
- Command (single invocation): `find internal pkg cmd .claude/rules .claude/skills .moai/config CLAUDE.md -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.tmpl' -o -name '*.js' -o -name '*.golden' \) ! -name '*_test.go' -exec awk 'tolower($0) ~ /opus[ -]5([^.0-9-]|\.[^0-9]|\.$|$)/ && tolower($0) !~ /superseded|measured on opus 5/ {print FILENAME ":" FNR}' {} +`
- RED-now: 147 lines, exit 0 — verbatim list `.moai/reports/t1089/p4-rednow-6e75b74db.txt`
- Green: stdout empty, exit 0. Because an empty stdout is also what a probe matching nothing prints, the green is read together with the positive control below.
- Positive control (proves the probe still fires on the green tree): `awk 'tolower($0) ~ /opus[ -]5([^.0-9-]|\.[^0-9]|\.$|$)/ {n++} END {print n+0}' internal/template/model_policy.go` → ≥ 1 (the superseded deprecated-id row itself matches the base pattern).

### AC-OP55-005 — constitution heading and zone-registry anchors move together (REQ-OP55-011) — release-blocking

- Commands:
  - `grep -c 'opus-55-prompt-philosophy' .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md` — RED-now `:0` / `:0`; Green `:2` / `:2`
  - `grep -c 'opus-5-48-prompt-philosophy' .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md` — RED-now `:2` / `:2`; Green `:0` / `:0`
  - `grep -c '^## Opus 5.5 Prompt Philosophy' .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — RED-now `:0` / `:0`; Green `:1` / `:1`
  - `go test -count=1 -run 'TestRegistrySyncGuard|TestRegistrySyncMirrorsIdentical' -v ./internal/constitution/` — Green: both test names print `--- PASS`, no `[no tests to run]`.

### AC-OP55-006 — current Opus facts stated (REQ-OP55-009, -010) — release-blocking

- Commands:
  - `grep -c '2\.1\.280' .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` — RED-now `:0` / `:0`; Green ≥ 1 each
  - `grep -c 'claude-opus-5-5' .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` — RED-now `:0` / `:0`; Green ≥ 1 each
  - `grep -cE 'Opus 5\.5[^|]*medium' .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — RED-now `:0` / `:0`; Green ≥ 1 each

### AC-OP55-007 — web labels and recommendation (REQ-OP55-004, -005, -006, -007) — release-blocking

- Commands (each on `internal/web/assets/i18n.js`):
  - `grep -cE '"f\.model\.opt\.opus\[1m\]": "Opus 5\.5' internal/web/assets/i18n.js` — RED-now `0` (exit 1); Green `4`
  - `grep -cE '"f\.effort_level\.opt\.medium": "[^"]*(Recommended|권장|推奨|推荐)' internal/web/assets/i18n.js` — RED-now `0`; Green `4`
  - `grep -cE '"f\.model\.opt\.opus\[1m\]": "[^"]*(Recommended|권장|推奨|推荐)' internal/web/assets/i18n.js` — RED-now `0`; Green `4`
  - `grep -cE '"opt\.runtime_default": "[^"]*(medium|중간|中)' internal/web/assets/i18n.js` — RED-now `0`; Green `4`

### AC-OP55-008 — TUI wizard labels and recommendation (REQ-OP55-004, -005, -006) — release-blocking

- Commands:
  - `grep -cE 'EffortLevelMedium:[^,]*(Recommended|권장|推奨|推荐)' internal/cli/profile_setup_translations.go` — RED-now `0`; Green `4`
  - `grep -cE 'ModelOpus1M:[^,]*Opus 5\.5[^,]*(Recommended|권장|推奨|推荐)' internal/cli/profile_setup_translations.go` — RED-now `0`; Green `4`

### AC-OP55-009 — label-drift guard rejects bare "Opus 5" (REQ-OP55-008) — release-blocking (mutant-proved)

- Given the green tree
- When one locale's `ModelOpus` label (or one model-policy description) is mutated from "Opus 5.5" to "Opus 5 "
- Then `go test -count=1 -run 'TestGetProfileText_OpusAliasValues' ./internal/cli/` and `go test -count=1 ./internal/cli/wizard/` each fail naming the mutated label
- Evidence: the mutant diff, both failing outputs (verbatim), and the reverted-tree PASS, recorded in progress.md §E.2. A guard that still passes under the mutant fails this AC.

### AC-OP55-010 — local/template pair identity preserved (REQ-OP55-009) — regression-guard

- Commands (each single-invocation; exit 0 = identical): `cmp .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md`, and the same `cmp` for `rules/moai/development/model-policy.md`, `rules/moai/development/prompting-best-practices.md`, `rules/moai/workflow/context-window-management.md`, `skills/moai-foundation-thinking/SKILL.md`
- Now: all five exit 0 (identical). Green: all five still exit 0 after edits. Mutant: editing only one side makes `cmp` exit 1.

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

- Command: `git diff --name-only develop...HEAD -- CHANGELOG.md docs-site README.md README.ko.md README.ja.md README.zh.md .moai/research .moai/docs`
- Green: stdout empty. (Three-dot form measures from the merge-base with local `develop`, so absorbed develop commits do not read as this card's changes.)
- Companion: `git diff --name-only develop...HEAD -- .moai/specs` lists only `.moai/specs/SPEC-MODEL-OPUS55-001/*`.

### AC-OP55-015 — template neutrality holds (REQ-OP55-016) — release-blocking at run close

- Command: `go test -count=1 -run 'TestTemplateNoInternalContentLeak|TestLanguageNeutrality|TestLeakClassNoDateShaInDefaultTier' -v ./internal/template/`
- Green: the three test names print `--- PASS`, exit 0.

### AC-OP55-016 — embedded template rebuilt; agent emission consistent (REQ-OP55-016) — release-blocking at run close

- Commands: `make build` → exit 0; `git diff --name-only develop...HEAD -- internal/template/templates/.claude/agents` → empty (no agent file changed, so `make agents-emit` is not required). If that list is non-empty: `make agents-emit-check` → exit 0.

## §2 Edge cases

- "Opus 5." at a sentence end must still be caught (P4's `\.$` / `\.[^0-9]` alternatives).
- "claude-opus-5-5" and "Opus 5.5" must never be caught (continuation exclusion).
- Wizard migration of `claude-opus-5[1m]`: the `[1m]` suffix is split off before normalization in the launcher path; run phase confirms the deprecated-row lookup sees the bare id.
- The four golden files are regenerated, not hand-edited.

## §3 Quality gate

- Tier M plan-auditor threshold 0.80.
- Run phase: zero new lint findings (`golangci-lint run ./internal/template/... ./internal/cli/... ./internal/web/...`), coverage of `internal/template` not lowered.

## §4 Definition of Done

All sixteen ACs green with evidence recorded in progress.md §E.2 (command + verbatim output + exit code + tree SHA); AC-OP55-009 mutant evidence present; follow-up docs card issued by the lead (plan.md §C.3).
