# SPEC-SONNET55-BUMP-001 — Implementation Plan

## §A Context

Card t1322, Class C (design change: a product-wide generation bump spanning code, templates, and
docs), Tier M. Tree measured at `a62a05764`. The bump is alias-preserving: users keep `sonnet`;
the canonical id behind it moves from `claude-sonnet-5` to `claude-sonnet-5-5`, and the old id
becomes a deprecated-prefs entry. Three card anchors were corrected against the tree (spec.md
§A.3) — most notably, the GLM slot mapping and the web validation layer are alias-driven and need
verification, not edits.

## §B Known Issues

- `expandModelString` currently snapshots `ModelAliasTable` at compile time; card t1315 (#1730,
  queued, NOT merged) will make it live. Same file (`internal/cli/launcher.go:1142`), different
  axis. Merge-order clause: REQ-SSB-011 / D-4 — second lander absorbs, no semantic conflict.
- `internal/web/assets/i18n.js` carries 44 case-insensitive sonnet mentions; most are
  generation-agnostic slot prose ("Sonnet slot"). Only picker labels and generation statements
  change — a blanket sed is wrong here.
- docs-site sonnet mentions span ~116 files across 4 locales; the spec scopes edits to pages
  stating the *current* generation, not every historical mention. M4 enumerates the page set
  before editing.

## §C Pre-flight

- [ ] `git rev-parse --short HEAD` == `a62a05764` (or absorb develop per gitflow §4.1 before starting)
- [ ] Confirm t1315 merge state: `gh pr view 1730 --json state,mergedAt` — record in progress.md; determines REQ-SSB-011 direction
- [ ] Confirm tree anchors: `grep -n '"sonnet"' internal/template/model_policy.go` shows `claude-sonnet-5`
- [ ] Baseline: `go test ./internal/template/... ./internal/cli/... ./internal/hook/... ./internal/web/...` green before first edit (affected packages only; no local full suite)

## §D Constraints

- Template-First: edit `internal/template/templates/**`, then `make build`. `make agents-emit`
  before build when `templates/.claude/agents/moai/*.md` changes (super-advisor.md qualifies).
  Never hand-edit the emitted `.codex/**/*.toml`.
- No local `go test ./...` (gitflow lane discipline); affected packages only, `-timeout 30m`.
- README ko canonical → en/ja/zh derived, same change; docs-site 4-locale parity in the same
  milestone; no emoji in docs bodies; Mermaid TD-only.
- Fixture-string principle (spec.md §D): served_model test fixtures unchanged.
- No commit/push by implementing agents inside the worktree without the lane's integration
  window (card worktree discipline).

## §E Self-Verification

Verification is AC-driven; every AC in acceptance.md names a command. Summary:

| Claim | Evidence command |
|---|---|
| Alias promoted + legacy map | `go test ./internal/template/...` + grep AC-SSB-002 |
| Slots resolve all three ids | full affected-package run from AC-SSB-003 (the AC-SSB-005 table case is inside it) (AC-SSB-005) |
| Labels updated | grep AC-SSB-006 (exact-match `"Sonnet 5"` vs `"Sonnet 5.5"` — substring trap: `grep '"Sonnet 5"'` matches the 5.5 string's prefix; use word-boundary or trailing-quote anchor) |
| Template build hygiene | `make agents-emit-check && make build` (AC-SSB-004) |
| Reference mirrors untouched | `git diff --stat -- internal/template/templates/.claude/skills/moai-foundation-cc/reference internal/template/templates/.claude/skills/moai-foundation-core` empty (AC-SSB-007) |
| Docs parity | README 4-file count parity + docs-site verify recipe (AC-SSB-008/009) |

## §F Milestones

Ordering follows decision-reversibility: the alias table (data-model, M1) first; then labels
(user-facing, M2); then template guidance (contains the research-gated context-window decision,
M3); then docs prose (derivative, M4).

### M1 — Alias core + slot verification (code)
- `internal/template/model_policy.go`: declare `ModelIDSonnet55 = "claude-sonnet-5-5"` beside
  `ModelIDOpus55`; promote `"sonnet"` row to reference it; add `"claude-sonnet-5": "sonnet"` to
  `ModelDeprecatedCanonicalIDs` (with a superseded-by comment, mirroring the opus row style).
  REQ-SSB-001/002/003.
- Tests: update current-canonical assertions in `glm_slot_test.go`, `glm_slot_effort_test.go`,
  `profile_setup_normalize_test.go`, `session_telemetry_cells_test.go` to pin
  `claude-sonnet-5-5`; ADD the three-way slot resolution case (REQ-SSB-005) proving
  `claude-sonnet-5` still resolves medium via the deprecated map. `launcher_test.go` and
  `served_model_*` untouched (REQ-SSB-010).
- Verify: `go test ./internal/template/... ./internal/cli/... ./internal/web/... ./internal/hook/...`
- **Type: code**

### M2 — User-facing labels (code, no logic)
- `internal/web/assets/i18n.js`: picker labels `f.model.opt.sonnet` / `f.model.opt.sonnet[1m]`
  → "Sonnet 5.5"; generation statements in opusplan desc and elsewhere where prose names the
  current Sonnet generation. Do NOT touch slot prose ("Sonnet slot") — it names a tier, not a
  generation.
- `internal/cli/profile_setup_translations.go`: "Sonnet 5" → "Sonnet 5.5" in model/effort strings
  for every locale block (en/ko/ja/zh as present); "Sonnet 5+" plan-gate prose → "Sonnet 5.5+".
  REQ-SSB-006.
- `internal/template/apply_harness.go`, `internal/web/validate.go`, `handlers.go`,
  `schemaform.go`: expected zero-diff; verify and record.
- Verify: affected-package tests + grep AC-SSB-006.
- **Type: code (user-facing strings)**

### M3 — Template guidance + context-window decision (template)
- Open with the research step (research.md §2): fetch the official model docs, record the
  context-window figure or its absence; apply D-5's gate. **This decision lands before any
  context-window table edit.**
- Edit moai-owned guidance (REQ-SSB-007): `model-policy.md` (generation statement + §E.4
  `between_tools` note), `context-window-management.md`(+detail) per D-5 outcome,
  `quality.yaml.tmpl`, `moai-constitution.md`, `prompting-best-practices.md`,
  `skill-authoring.md`, `settings-management.md`, `moai-mcp-tools-catalogue.md`,
  `agents/moai/super-advisor.md` — Sonnet 5 → Sonnet 5.5 in effort-support and generation lists
  only; alias mentions and slot prose untouched.
- Hygiene: `make agents-emit` (super-advisor.md changed) → `make agents-emit-check` →
  `make build`. REQ-SSB-008.
- Verify: AC-SSB-004/007.
- **Type: template**

### M4 — README + docs-site (docs)
- README.ko.md canonical (3 sonnet lines: benchmark row label, generation prose, GLM slot table
  row) → derive en/ja/zh; same change-set. REQ-SSB-009.
- docs-site: enumerate sonnet-mention pages per locale (measured 116 files; start from
  `advanced/profile-matrix.md`, `advanced/no-haiku-3tier.md`, `multi-llm/model-policy.md`,
  `tokenomics/*`, `cost-optimization/*`); edit only current-generation statements + add the
  `between_tools` migration note to the model-policy page; 4-locale parity in the same
  milestone. Verify with the hns-oss-docs-verify recipe (hugo build warning-free, 4-locale
  parity).
- Verify: AC-SSB-008/009/010.
- **Type: docs**

### Merge-order note (REQ-SSB-011)
Before merging to develop, re-check t1315 state. If t1315 landed in develop meanwhile: absorb it
(`git merge develop` in this worktree), confirm `expandModelString` still returns
`ModelIDSonnet55` for `"sonnet"` (launcher_test covers this dynamically), and record the
absorption in progress.md. If t1315 has not landed, nothing to do — its branch rebases onto this
SPEC's table contents.

## §G Anti-Patterns

- Blanket `sed s/sonnet-5/sonnet-5.5/` over the tree — corrupts `claude-sonnet-5` deprecated-map
  keys, test fixtures, and historical prose in one shot.
- Editing `moai-foundation-cc/reference/**` — verbatim upstream mirrors (D-3).
- Hand-editing `templates/.codex/agents/moai/*.toml` — machine-emitted (C3).
- Rewriting historical benchmark numbers to "predict" 5.5 (REQ-SSB-009 keeps them labeled).
- Assuming the context window (D-5 forbids silently writing 1M or 200K).

## §H Cross-References

- spec.md §E (research items) ↔ research.md; acceptance.md AC-SSB-001..010 ↔ REQ-SSB-001..013.
- t1315 / GitHub #1730 (mechanism), t1246 (absorption that shrank this SPEC's surface).
- `.claude/skills/hns-oss-docs-verify/SKILL.md` — docs-site exit gate recipe.
