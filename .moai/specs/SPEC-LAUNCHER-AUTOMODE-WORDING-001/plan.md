# Plan — SPEC-LAUNCHER-AUTOMODE-WORDING-001

Card t1182 · Tier S · branch `WT-auto-mode-help` · base develop `a520187f1`.

## §A Context

Wording-only correction of the auto permission-mode eligibility claim and the acceptEdits "project default" claim across 6 files (2 Go, 4 locale mirrors of one docs line). Premise and repo evidence: spec.md §1. Tier S despite 6 files: four files are locale mirrors of a single sentence and total churn is well under 30 LOC.

## §B Known Issues / Decisions (highest change-likelihood first)

### B1 — Chosen wording per target (the decision most likely to be revised in review)

| Target | Replacement text |
|---|---|
| T1 `internal/cli/cc.go:103` | `  auto               Background classifier checks actions (requires a supported model and plan; see Claude Code permission-modes docs)` |
| T7 `internal/cli/cc.go:101` | `  acceptEdits        Auto-accept file edits, ask for commands (moai init default)` |
| T2 `internal/cli/glm.go:319` | `auto mode requires a supported Claude model; GLM models are not supported (see Claude Code permission-modes docs)` — the next line (`use 'moai cc --permission-mode auto' instead`) and the returned error stay byte-identical |
| T4 en `launchers.md:47` | ``The permission mode is one of `default`, `acceptEdits` (the `moai init` default), `plan`, `auto`, `bypassPermissions`, `dontAsk`. The `auto` mode runs a background classifier that inspects actions; the plans and models that support it are listed in the [Claude Code permission modes documentation](https://code.claude.com/docs/en/permission-modes).`` |
| T3 ko `launchers.md:47` | ``권한 모드는 `default`, `acceptEdits`(`moai init` 기본값), `plan`, `auto`, `bypassPermissions`, `dontAsk` 중 하나입니다. `auto` 모드에서는 백그라운드 분류기가 동작을 검사하며, 사용할 수 있는 플랜과 모델은 [Claude Code 권한 모드 문서](https://code.claude.com/docs/en/permission-modes)를 따릅니다.`` |
| T5 ja `launchers.md:47` | ``権限モードは `default`、`acceptEdits`(`moai init` の既定値)、`plan`、`auto`、`bypassPermissions`、`dontAsk` のいずれかです。`auto` モードはバックグラウンド分類器が動作を検査するもので、対応するプランとモデルは [Claude Code の権限モードのドキュメント](https://code.claude.com/docs/en/permission-modes) に従います。`` |
| T6 zh `launchers.md:47` | ``权限模式为 `default`、`acceptEdits`(`moai init` 的默认值)、`plan`、`auto`、`bypassPermissions`、`dontAsk` 之一。`auto` 模式由后台分类器检查动作,支持的方案与模型以 [Claude Code 权限模式文档](https://code.claude.com/docs/en/permission-modes) 为准。`` |

Parenthesis/punctuation style of each locale line is preserved from the current line (ASCII `(` `)` after the backticked token; zh keeps its existing ASCII comma). The run phase may adjust native idiom but MUST keep the AC-003/AC-004 tokens (`moai init`, the en URL, no removed phrase).

### B2 — Why defer to the doc instead of reusing constants (constraint resolution)

`internal/template/model_policy.go` holds model **IDs** for alias resolution (`ModelIDOpus55`, `ModelAliasTable`, effort constants) — not an auto-mode eligibility matrix. Reusing them would re-encode exactly the provider-dependent version list the dispatch forbids, and would still be wrong for Bedrock/Vertex/Foundry. No existing constant expresses "supported model and plan", so the Go strings name the doc page in prose ("Claude Code permission-modes docs") and carry **no URL**: CLAUDE.local.md §14 requires URLs in Go code to be extracted to constants, and adding one const for a help string is not worth it. The docs-site pages carry the URL (precedent: 121 `code.claude.com/docs/en/` links in non-en pages).

### B3 — Why "moai init default" (T7)

Evidence (spec.md §1.3): the project template has no `defaultMode`; `moai init` writes USER-scope `defaultMode="acceptEdits"` by default (`init.go:932` → `autonomy_bundle.go:71`). "moai init default" is true for both readers: a user reading `moai cc --help` and a docs reader. It does not claim Claude Code's built-in default (auto on Pro/Max/Team), which a user without the init record would get instead.

### B4 — Residual drift left unfixed (follow-up cards for the lead)

spec.md §4: `PermAuto` × 4 (`profile_setup_translations.go:192/288/384/480`), `acceptEditsConfirmationLine` + translations (`profile_setup.go:31`, pinned by `profile_setup_acceptEdits_test.go`), `launcher.go:1088/1107-1108` comments, `launcher_test.go:511` subtest name.

## §C Pre-flight (run phase)

1. `git branch --show-current` → `WT-auto-mode-help`; `git status --porcelain` → clean.
2. Re-read the six target lines and confirm the §1.2 baselines (`grep -c "4\.6"` = 1 per file) before editing — content-anchored replacements, not line numbers.
3. `git diff develop -- internal/cli/cc.go internal/cli/glm.go docs-site/content/*/cli-reference/launchers.md` empty (no sibling card touched them since base).

## §D Constraints

- Go code comments and strings in English; `gofmt` clean.
- Four locales land in ONE commit (docs-site i18n rule).
- No plan-tier name or model version literal added to Go strings (REQ-005 / AC-006).
- Scope verification: `go test ./internal/cli/ -run …` only; do not run `go test ./...` locally (CLAUDE.local.md §4). `internal/cli` full-package runs can hit the 10-minute timeout — use `-run`.
- Do not push; lane reports the local develop merge SHA to the lead.

## §E Self-Verification (run phase; results into progress.md §E.2)

Run AC-001..AC-007 (spec.md §3) as one batched read-only turn, plus `go vet ./internal/cli/` and `gofmt -l internal/cli/cc.go internal/cli/glm.go` (expect empty). Optional docs check: `hugo --gc --minify` warning-free in `docs-site/` if the toolchain is present; absence is recorded as a Gap, not a pass.

## §F Milestones (priority order)

- **M1 — Priority High — wording confirmation.** Confirm B1 texts against the official page (re-fetch allowed); adjust native idiom only.
- **M2 — Priority High — Go edits + wording tests.** Edit T1/T7/T2; extend `TestCharacterize_CC_HelpFlag` (or add a sibling test) to assert `moai init default` + `permission-modes docs` in help output; extend `TestCharacterize_GLM_AutoModeRejected` to capture stderr and assert `supported Claude model` present and `4.6` absent.
- **M3 — Priority Medium — docs 4-locale edit.** Replace line 47 in all four locales in the same commit.
- **M4 — Priority Medium — verification + evidence.** AC batch, progress §E.2, card verdict at `.moai/reports/t1182/verdict.md` (created at close, not in plan).

## §G Anti-Patterns

- Replacing the stale list with a newer hardcoded list (e.g. "Sonnet 5 / Opus 4.7") — reproduces the defect on the next release.
- Touching the out-of-scope `PermAuto` / confirmation-line surfaces "while in the file".
- Editing only en and leaving the other three locales for later.

## §H Cross-References

- spec.md §1.1 (official facts), §1.3 (repo evidence), §4 (exclusions).
- SPEC-AUT-PERMMODES-001 (acceptEdits init default decision).
- `.moai/docs/docs-site-i18n-rules.md`.
