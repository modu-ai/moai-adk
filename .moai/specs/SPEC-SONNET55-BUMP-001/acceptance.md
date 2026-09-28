# SPEC-SONNET55-BUMP-001 — Acceptance Criteria

Every criterion is mechanically verifiable. Commands run from the worktree root.

## §D AC Matrix

| AC | Verifies | Given | When | Then (evidence) |
|---|---|---|---|---|
| AC-SSB-001 | REQ-SSB-001 | The tree carries the promoted table | `grep -n 'ModelIDSonnet55' internal/template/model_policy.go` runs | Output shows `const ModelIDSonnet55 = "claude-sonnet-5-5"` and the `"sonnet"` row referencing it; no `"sonnet": "claude-sonnet-5"` row remains (`grep -c '"sonnet":\s*"claude-sonnet-5"' internal/template/model_policy.go` returns 0) |
| AC-SSB-002 | REQ-SSB-002 | The deprecated-id map | `grep -n '"claude-sonnet-5":' internal/template/model_policy.go` runs | Output shows `"claude-sonnet-5": "sonnet"` inside `ModelDeprecatedCanonicalIDs` with a superseded-by comment |
| AC-SSB-003 | REQ-SSB-004, 010 | Affected packages | `go test -timeout 30m ./internal/template/... ./internal/cli/... ./internal/web/... ./internal/hook/...` runs | Exit 0; `launcher_test.go` and `served_model_test.go` / `served_model_stop_test.go` appear in the run unchanged (`git diff --stat -- internal/cli/launcher_test.go internal/hook/served_model_test.go internal/hook/served_model_stop_test.go` is empty) |
| AC-SSB-004 | REQ-SSB-008 | Template edits landed | `make agents-emit && make agents-emit-check && make build` runs (agents-emit needed only if a `templates/.claude/agents/moai/*.md` changed) | All three exit 0; `git status --porcelain -- internal/template/templates/.codex/` shows no hand-edit beyond the emission |
| AC-SSB-005 | REQ-SSB-005 | The GLM slot resolver | A table test exercises `GLMSlotForModel` with `claude-sonnet-5-5`, `claude-sonnet-5`, and `sonnet[1m]` | All three return `GLMSlotMedium`; test green in AC-SSB-003's run |
| AC-SSB-006 | REQ-SSB-006 | User-facing labels | `grep -rn '"Sonnet 5.5"' internal/web/assets/i18n.js internal/cli/profile_setup_translations.go` and `grep -rnE '"Sonnet 5"|Sonnet 5[,\) 、]' internal/web/assets/i18n.js internal/cli/profile_setup_translations.go` run | First grep shows the updated picker labels in every locale block present; second grep returns 0 hits (trailing-quote/punctuation anchor — `"Sonnet 5.5"` must not self-match) |
| AC-SSB-007 | REQ-SSB-007, D-3 | Reference mirrors | `git diff --stat -- internal/template/templates/.claude/skills/moai-foundation-cc/reference internal/template/templates/.claude/skills/moai-foundation-core` runs | Empty diff; meanwhile `git diff --name-only -- internal/template/templates | grep -c model-policy` shows the moai-owned rule changed |
| AC-SSB-008 | REQ-SSB-009 | README 4-locale set | `for f in README.ko.md README.md README.ja.md README.zh.md; do grep -c 'Sonnet 5.5' $f; done` runs | All four counts equal (parity), and each is ≥ the pre-change count for generation statements; benchmark rows retain their historical numbers (`grep -n '54%±4' README.ko.md` still resolves) |
| AC-SSB-009 | REQ-SSB-009 | docs-site | The hns-oss-docs-verify recipe runs: warning-free hugo build, 4-locale file-existence + section parity | Recipe exits clean; `grep -rlc 'Sonnet 5.5' docs-site/content | wc -l` ≥ 1 and each hit locale set covers en/ja/ko/zh |
| AC-SSB-010 | REQ-SSB-012, 013 | Research + migration note | (a) `grep -n 'between_tools' internal/template/templates/.claude/rules/moai/development/model-policy.md docs-site/content/*/multi-llm/model-policy.md` (b) research.md §2 | (a) shows the migration note in the template rule and in all 4 docs-site locales; (b) carries the official-docs citation (URL + figure or an explicit "not stated") and the D-5 decision applied or declined |

## §D.1 Severity

- **Blocker**: AC-SSB-001, AC-SSB-002, AC-SSB-003, AC-SSB-004 (behavior + build hygiene)
- **Major**: AC-SSB-005, AC-SSB-006, AC-SSB-010 (user-facing correctness)
- **Minor**: AC-SSB-007, AC-SSB-008, AC-SSB-009 (scope discipline + parity; still must pass)

## §D.2 Edge Cases

- `sonnet[1m]` resolves to `claude-sonnet-5-5[1m]` (suffix preserved) — covered by existing
  `launcher_test.go` alias-driven cases plus AC-SSB-005's `[1m]` case.
- Prefs file carrying `claude-sonnet-5` (pre-bump canonical) normalizes to alias `sonnet` via
  `ModelDeprecatedCanonicalIDs` — `profile_setup_normalize_test.go` gains the case.
- Unknown id passthrough unchanged (`expandModelString("claude-sonnet-4-6")` stays verbatim).
- i18n label grep must not self-match: "Sonnet 5.5" contains "Sonnet 5" — anchors required
  (AC-SSB-006).

## §D.3 Quality Gates

- TRUST 5 Tested: affected packages green (AC-SSB-003); no coverage regression in
  `internal/template` (alias table is data — the slot/normalize tests carry the behavior).
- Unified: `golangci-lint run ./internal/template/... ./internal/cli/...` clean.
- Trackable: commit messages carry card id t1322; template neutrality CI green.

## §D.4 Definition of Done

All Blocker + Major ACs pass with cited command output in progress.md §E.2; Minor ACs pass or
carry an explicit, evidence-cited waiver; the merge-order check (plan.md §F note) is recorded.
