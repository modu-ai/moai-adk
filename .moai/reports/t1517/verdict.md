# Card t1517 — web console Fable label 5 → 5.1 (verdict record)

## Dispatch record

- Card: t1517, class A/B (small change, no SPEC). Origin: follow-up of t1503, which moved the model alias to `claude-fable-5-1`.
- Worktree: `.claude/worktrees/t1517`, branch `WT-console-fable-51-label`, base `6643c7bba` (= local `develop` tip at dispatch; the tree was first created at `9e90e1252` with no commits and fast-forwarded to the base).
- **Lease: none** — run under leader direct dispatch, as on the preceding cards of this lane.

## Change

- `internal/web/assets/i18n.js`: `f.model.opt.fable` and `f.model.opt.fable[1m]` in all four locale blocks (lines 352-353, 1307-1308, 2142-2143, 2977-2978) from `Fable 5` to `Fable 5.1`.
- `internal/web/console_ux_fix_test.go`: `TestModelOptLabelsEnglishUnified` expectation `Fable 5` → `Fable 5.1`, and the comment above it. The two banned-old-label entries (`Fable 5 (200K)` / `Fable 5 (1M)`) are intentionally untouched.
- Commit `674bec513` (2 files, 10 insertions, 10 deletions).

## Evidence

| Claim | Command | Observed |
|---|---|---|
| New expectation is red on the base file | `git show HEAD:internal/web/assets/i18n.js` piped to `grep -c '"f.model.opt.fable\[1m\]": "Fable 5.1"'` (run before the commit, when `HEAD` was the base `6643c7bba`) | `0` — the updated test would fail on the old labels (RED) |
| New expectation is green on the edited file | `grep -c '"f.model.opt.fable\[1m\]": "Fable 5.1"' internal/web/assets/i18n.js` | `4` — one per locale, as the test requires |
| Package tests pass | `go test -count=1 -timeout 30m ./internal/web/...` | `ok  github.com/modu-ai/moai-adk/internal/web 36.405s` |
| CLI-side wording already agrees | `grep -n ModelFable internal/cli/profile_setup_translations.go` | `fable (Fable 5.1, deep reasoning)` / `fable[1m] (Fable 5.1 + 1M context)` (en), Korean equivalents |

## Gaps

- **Card review not run in the lane.** `codex_review` is not available in this session (tool search returned nothing). Per the dispatch, the leader runs the codex card review; `card-review.md` is therefore **absent by recorded reason**, not by omission.
- Only `./internal/web/...` was run (scoped to the change); the full suite is CI's. `internal/cli` was not re-run because it is not touched; its CLI-side labels already read 5.1.
- No browser rendering check of the console picker; the labels are plain dictionary strings pinned by the test.

## Residual risk

Low. Dictionary strings only. `i18n_governance_test.go` / `i18n_untranslated_allowlist_test.go` reference the same keys and passed in the package run.

## Status

merge-ready pending the leader's codex card review. Not pushed; not merged into `develop`.
