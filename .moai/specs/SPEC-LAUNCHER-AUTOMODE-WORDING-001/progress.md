# Progress — SPEC-LAUNCHER-AUTOMODE-WORDING-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts (Tier S: spec.md + plan.md + progress.md; AC inline in spec.md §3) authored 2026-09-25 by manager-spec for card t1182 on branch `WT-auto-mode-help`, base develop `a520187f1`.
- Premise measured on this tree in this run: six target lines read verbatim; baselines recorded in spec.md §1.2; `defaultMode` absent from `settings.json.tmpl` (grep exit 1); init USER-scope acceptEdits write traced (`init.go:932` → `autonomy_bundle.go:71`).
- Spec lint: see §E.1.1.

### §E.1.1 Spec lint result

Command: `moai spec lint SPEC-LAUNCHER-AUTOMODE-WORDING-001` (2026-09-25, pre-commit, this tree at base `a520187f1`, installed moai v3.2.0-rc.15). Verbatim output: `✓ No findings — all SPEC documents are valid` (exit 0). Lint ownership checks are blind before the commit — the close path re-measures post-commit.

## §E.2 Run-phase Evidence

- go test ./internal/cli/ -run 'TestCharacterize_CC_HelpFlag|TestCharacterize_GLM_AutoMode' -count=1 -timeout=90s → ok github.com/modu-ai/moai-adk/internal/cli 0.843s (exit 0). Help test checks the new init default and documentation reference; both GLM flag spellings check the version-free reason.
- Inline Python AC-001..004 check over the two Go files and four locale pages → AC-001..004 PASS; locale link/default line: [47, 47, 47, 47] (exit 0).
- go vet ./internal/cli/ → no output, exit 0.
- gofmt -l internal/cli/cc.go internal/cli/glm.go internal/cli/cc_test.go internal/cli/glm_new_test.go → no output, exit 0.
- git diff --check → no output, exit 0.
- git diff --name-only a520187f1..origin/develop -- the eight implementation/test/doc paths → no output, exit 0. The current upstream develop did not change these paths since the SPEC baseline.
- Revised AC-006 command: git diff origin/develop...HEAD -- internal/cli/cc.go internal/cli/glm.go | grep '^+' | grep -v '^+++' | grep -cE 'Sonnet|Opus|Fable|[0-9]\.[0-9]|\b(Pro|Max|Team|Enterprise)\b' → output 0; grep pipeline exit 1 means no matching added line. Revised AC-007 command, git diff --name-only origin/develop...HEAD → 12 paths, all within the SPEC allowlist (the card report, three SPEC files, four locale pages, two Go sources, two Go tests).
- After the main-thread merge, HEAD 0d8fa76fc: git rev-list --count --left-right origin/develop...HEAD → 0 7 (exit 0). The revised AC-006 command again prints 0 (grep exit 1, no matches). The revised AC-007 command again prints the same 12 permitted paths (exit 0). Their upstream-absorption precondition is now met: AC-006 PASS, AC-007 PASS.
- Post-merge go test ./internal/cli/ -run 'TestCharacterize_CC_HelpFlag|TestCharacterize_GLM_AutoMode' -count=1 -timeout=90s → ok github.com/modu-ai/moai-adk/internal/cli 0.674s (exit 0). Post-merge inline Python AC-001..004 → AC-001..004 PASS; locale link/default line: [47, 47, 47, 47] (exit 0). git diff --check origin/develop...HEAD → no output, exit 0.

## §E.3 Run-phase Audit-Ready Signal

Six requested surfaces and two focused tests were committed together in 7b69ab3ca. The main thread absorbed origin/develop in merge commit 0d8fa76fc. Post-merge AC-001..007 checks pass against this branch and the focused GLM rejection tests remain green. Sync-phase review and remote PR checks remain open.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: pending-backfill  # the sync commit itself cannot carry its own SHA; reported to lead
sync_status: completed
frontmatter_status_transitions:
  spec.md: in-progress -> completed (updated 2026-09-26)
  plan.md: no frontmatter (unchanged)
  progress.md: no frontmatter (unchanged)
changelog_entry_position: none (no CHANGELOG/README/docs-site changes in this sync, per lead instruction)
b12_self_test_a: not applicable (no CHANGELOG emission)
b12_self_test_b: not applicable (no CHANGELOG emission)
b12_self_test_c: not applicable (no CHANGELOG emission)
```

Landing: implementation landed on origin/develop as bf0b34df2 (PR #1727). Lane-reported (2026-09-26): card files byte-identical to lane tree 63254d565. Sync-tree measurement (this run, worktree `t1182-sync`, HEAD e5d6030f1): `git diff --stat HEAD origin/develop -- <12 card files> | wc -l` → `0`. Note: origin/develop had advanced to 85d99c6b4 at measurement time; the only HEAD↔origin/develop difference under `internal/cli docs-site/content .moai/reports/t1182 .moai/specs/SPEC-LAUNCHER-AUTOMODE-WORDING-001` is `internal/cli/mcp_project_root_doc_test.go | 2 +-`, outside the card's file set.

Re-measurement in this tree (HEAD e5d6030f1), verbatim:

- `go test ./internal/cli/ -run TestCharacterize_CC_HelpFlag -count=1 -timeout=180s` → `ok  	github.com/modu-ai/moai-adk/internal/cli	0.808s` (exit 0)
- `go test ./internal/cli/ -run TestCharacterize_GLM_AutoMode -count=1 -v -timeout=180s` → `--- PASS: TestCharacterize_GLM_AutoModeRejected (0.00s)`, `--- PASS: TestCharacterize_GLM_AutoModeEqualsSyntaxRejected (0.00s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	0.590s` (exit 0)
- `go vet ./internal/cli/` → no output, exit 0
- `gofmt -l internal/cli/cc.go internal/cli/glm.go internal/cli/cc_test.go internal/cli/glm_new_test.go` → no output, exit 0
- Removed-phrase grep (AC-001/002/003 patterns): cc.go `0`, glm.go `0`, launchers.md en/ja/ko/zh `0` each — 0 across the 6 targets
- `moai spec lint SPEC-LAUNCHER-AUTOMODE-WORDING-001` (pre-edit) → `✓ No findings — all SPEC documents are valid` (exit 0)

Lead decisions (2026-09-26):

- Codex-authored commits accepted as evidence verified.
- D1 resolved by substitute measurement: `origin/develop...HEAD` range = 12 allowed paths; the pre-absorption measurement is no longer possible.
- Remote-develop absorption concern moot after landing.

Carried-over debt (not fixed):

- D4 — AC-007 permits more than REQ-006 (all `internal/cli/*_test.go` and `.moai/reports/t1182/`).
- D5 — "requires a supported model and plan" is loose for API users, for whom only the model condition applies.

Out-of-scope follow-ups (not in this SPEC):

- `profile_setup_translations.go` PermAuto ×4
- `profile_setup.go:31` `acceptEditsConfirmationLine` and its test
- `launcher.go:1088` comment
- `launcher_test.go:511` subtest name
