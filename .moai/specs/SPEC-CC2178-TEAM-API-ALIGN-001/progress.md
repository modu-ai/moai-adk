# Progress — SPEC-CC2178-TEAM-API-ALIGN-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set authored: spec.md + plan.md + acceptance.md + design.md + progress.md.
- SPEC ID self-check: `decomposition: SPEC ✓ | CC2178 ✓ | TEAM ✓ | API ✓ | ALIGN ✓ | 001 ✓ → PASS`.
- Frontmatter: 12 canonical fields present + validated; `status: draft`.
- Blast radius measured live at authoring (not assumed): 16 doctrine files (15 `.claude/**` + CLAUDE.md) symmetric with 16 template mirrors; 9 docs-site files (4×2 + ko/advanced/hooks-guide); 4 README locales; 4 Go references assessed (all comment/`t.Log`/`fmt.Printf`/mock-name — none live-behavior).
- Out of Scope section present (spec.md §J), with `### Out of Scope —` H3 sub-headings.
- Tier M justification recorded (design.md §E).
- plan-auditor verdict: PASS 0.89 (1 major + 3 minor). Pre-run-phase fixes applied (draft retained, spec-lint clean): D1 AC-LOC-003 + REQ-LOC-003 (README orphan-class MUST-FIX); D2 `.moai/config/`-path baseline capture (plan.md §C step 3 — to be recorded here at run-phase entry as CFG_CONFIG_PATH_BASELINE); D3 AC-CFG-001 non-vacuous 3-anchor tightening; D4 OQ-1 type unconfirmed → M5 in-instance-confirmation-conditional; D5 OQ-2/OQ-3 defaults encoded as Implementation-Kickoff-Approval-overridable. Counts after fixes: 21 REQs, 18 ACs (added REQ-LOC-003 + AC-LOC-003).
- _Audit-ready: pending Implementation Kickoff Approval (plan→run human gate)._

## §E.2 Run-phase Evidence

### Run-phase entry (pre-flight, recorded before edits)

- Frontmatter transition `draft → in-progress` applied to spec.md (`updated:` unchanged — same-day run).
- Git baseline: synced (`git rev-list --count --left-right origin/main...HEAD` → `0  0`); HEAD at run-phase entry verified clean.
- Cross-platform build baseline: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Lint baseline: `golangci-lint run --timeout=2m` → `0 issues`.
- Blast-radius re-grep (16 in-scope doctrine files confirmed; `agent-memory/` excluded as out-of-scope per spec.md §J transient/runtime).
- **CFG_CONFIG_PATH_BASELINE = 152** (D2 / AC-CFG-002 invariant). Command:
  `grep -rc '\.moai/config/' .claude/ CLAUDE.md | grep -v '\.claude/worktrees/' | grep -v '/backups/' | grep -v 'agent-memory' | awk -F: '{s+=$2} END{print s}'` → `152`.
  AC-CFG-002 PASS condition: identical post-M4 count.
- AC-CFG-001 baseline (Facts 1/2/3): all 0 before edit — any post-edit ≥1 is genuinely introduced by M4.

### OQ-1 RESOLUTION (sessionUrl type CONFIRMED — no deferral)

The installed `claude` binary is **v2.1.183** (`/Users/goos/.local/share/claude/versions/2.1.183`) — the exact CC version that introduced `attribution.sessionUrl`. Searching the bundled (minified-JS) zod schema in the binary via `strings | grep` yielded the authoritative machine-readable schema definition verbatim:

```
sessionUrl: E.boolean().optional().describe("Whether to append the claude.ai session link
to commits and PRs created from web or Remote Control sessions (default: true).
Set to false to omit the Claude-Session trailer and PR-body link.")
```

(`E` is the zod alias; the sibling `commit`/`pr` keys use `E.string().optional()`.) **Type CONFIRMED = boolean, default true, `false` omits the link.** Per the maintainer's OQ-1 instruction ("if you discover a local authoritative schema bundled with the installed `claude` that confirms the type, you MAY pin it"), the deferral is NOT triggered. M5 pins `"sessionUrl": false` (omit the claude.ai link, consistent with MoAI's own `🗿 MoAI` attribution trailers — the omit-behavior the setting was introduced for).

### REQ-GO-001 ASSESSMENT (recorded — no Go code change)

All 4 named Go files reference `TeamCreate`/`TeamDelete` ONLY in non-live-behavior contexts (verified via `grep -n`):
- `internal/template/skills_audit_test.go:50` — code comment inside a `requiredPatterns` literal; the actual patterns checked for `team/run.md` are gate markers (`Phase 0.5: Plan Audit Gate`, `plan-auditor`, `--skip-audit`, `INCONCLUSIVE`, `.moai/reports/plan-audit/`), NOT the tool name.
- `internal/cli/team_run_audit_gate_test.go` — all 13 matches are comments + `mockTeamCreator` mock type name + `t.Error` assertion strings modeling the gate-before-spawn ordering (MoAI-internal Plan Audit Gate contract, survives the upstream tool removal conceptually).
- `internal/runtime/audit_gate.go:301,306` — one code comment + one `fmt.Printf("[plan-audit] team mode detected, gate applies before TeamCreate\n")` log string.
- `internal/cli/team_spawn_test.go:648` — `t.Log` string.

None invoke a removed Claude Code tool. Per design.md §G AP-4, the gate-before-spawn ordering is a MoAI-internal contract; the comment/log/mock names are preserved. **Default applied: no Go behavior change.**

### M2 hooks-guide ko reconciliation (OQ-3 default applied)

OQ-3 default applied per maintainer instruction: REWRITE (not delete) the ko `hooks-guide.md` 3-step block — preserve the tmux-pane-lifecycle warning, drop the `TeamDelete` cleanup step (cleanup is automatic on session exit).

### OQ-2 (Mermaid nodes) default applied

OQ-2 default applied per maintainer instruction: REWRITE `TeamCreate → SendMessage` Mermaid nodes to `Agent(name=…) → SendMessage` (do NOT remove the nodes).

## §E.3 Run-phase Audit-Ready Signal

### AC PASS/FAIL matrix (verification-claim-integrity — each PASS attributed to an actually-run command)

| AC | Sev | Status | Evidence (command → observed output) |
|----|-----|--------|--------------------------------------|
| AC-TAA-001 | MUST | PASS | `grep -rn 'TeamCreate\|TeamDelete' .claude/ CLAUDE.md` (excl worktrees/backups/agent-memory) filtered for non-migration → `no live-action instructions remain`; 11 residual mentions all migration/historical notes ("removed in v2.1.178" / "no explicit ... call") |
| AC-TAA-002 | MUST | PASS | `grep -rln 'implicit team\|name parameter\|spawn teammates directly\|no setup step'` → 12 files |
| AC-TAA-003 | MUST | PASS | `grep -rn 'accepted but ignored\|team_name.*deprecated\|session-derived'` → 14 matches (Agent-tool param + hook-payload field both covered) |
| AC-TAA-004 | MUST | PASS | `grep -rn 'Call TeamDelete only after'` → 0; `grep -rln 'automatic.*session exit\|cleanup.*automatic'` → 13 |
| AC-TAA-005 | SHOULD | PASS | teammateMode file set unchanged (worktree-integration.md + moai-workflow-worktree/SKILL.md, no new files); `grep -c 'in-process'` → 1 each. NOTE: the 2 files discuss MoAI's `settings.local.json` launcher-field `teammateMode` (tmux/glm/claude), NOT the CC-runtime `teammateMode` (auto→in-process); a clarifying note was added to BOTH distinguishing the two same-named fields and recording the v2.1.179 auto→in-process + v2.1.181 idle-row-hide facts (verification-claim-integrity: avoided conflating two distinct settings) |
| AC-TAA-006 | MUST | PASS | `grep -n 'TeamCreate\|TeamDelete' CLAUDE.md internal/template/templates/CLAUDE.md` filtered for non-migration → 0 live-API listings; both files byte-identical (AC-MIR-001) |
| AC-TAA-007 | SHOULD | PASS | CLAUDE.md §4 "Watch (Claude Code 2.1.172)" note untouched — factually consistent with implicit-team model (nested-delegation topic, not team-lifecycle); no scope expansion |
| AC-CFG-001 | SHOULD | PASS | Fact1 `/config key=value`→1, Fact2 `/config --help`→1, Fact3 `Enter.*Space.*Esc`→1 (all in settings-management.md § /config command subsection; baseline was 0/0/0 pre-edit) |
| AC-CFG-002 | MUST | PASS | `.moai/config/` path count: baseline 152 → post-M4 152 (invariant held; an initial 153 from a `.moai/config/` token in my M4 prose was reworded to `.moai`-prefixed paths) |
| AC-ATT-001 | MUST | PASS | `grep -n sessionUrl settings.json.tmpl` → `"sessionUrl": false`; type CONFIRMED boolean from bundled CC v2.1.183 schema (OQ-1 resolved, NOT deferred) |
| AC-ATT-002 | SHOULD | PASS | settings-management.md § Claude Code Settings attribution bullet documents commit/pr/sessionUrl; local+mirror parity |
| AC-MIR-001 | MUST | PASS | parity diff loop over all 19 edited `.claude`/CLAUDE.md files → ALL PARITY OK; `make build` regenerated catalog.yaml (moai-workflow-worktree skill hash) + compiled binary (embed mechanism is `//go:embed all:templates` in embed.go — NO separate embedded.go snapshot file; direct directory embed at compile time) |
| AC-NEU-001 | MUST | PASS | `go test ./internal/template/... -run TestTemplateNeutralityAudit -count=1` → ok |
| AC-LOC-001 | MUST | PASS | per-locale loop: en/ko/ja/zh × {what-is, intro} all `grep -c 'TeamCreate\|TeamDelete'` → 0 (8/8 files) |
| AC-LOC-002 | MUST | PASS | per-locale loop: en/ko/ja/zh hooks-guide all → 0 (ko reconciled per OQ-3: rewrote block, preserved tmux-pane warning, dropped TeamDelete step, no literal TeamDelete token) |
| AC-LOC-003 | MUST | PASS | Part A: README.{md,ko,ja,zh}.md all `grep -c 'TeamCreate\|TeamDelete'` → 0; Part B: all `grep -c 'Agent(name'` → 2 (OQ-2 default applied: Mermaid node + bullet) |
| AC-QG-001 | MUST | PASS | `go test ./...` → all ok (initial flaky FAIL did not reproduce on `-count=1`); `golangci-lint run` → 0 issues; TestTemplateNeutralityAudit → ok; skills audit (TestSkills) → ok |
| AC-GO-001 | MUST | PASS | `git diff --name-only HEAD -- internal/**/*.go` → 0 Go source/test files changed; the 4 named files' TeamCreate/TeamDelete references confirmed comment/t.Log/fmt.Printf/mock-name only (default applied: no Go change) |

### §E self-verification (E1-E7)

- **E1 (AC matrix)**: 18/18 ACs PASS (13 MUST-FIX + 5 SHOULD). Matrix above.
- **E2 (cross-platform build)**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3 (coverage)**: N/A — documentation/template-only SPEC; zero Go behavior change (AC-GO-001), no new Go code to cover. Existing package tests all pass.
- **E4 (subagent-boundary grep)**: N/A — no `internal/harness/` or `internal/hook/` edits; this SPEC is doctrine/template/docs only.
- **E5 (lint status)**: `golangci-lint run --timeout=2m` → `0 issues` (== baseline 0).
- **E6 (push state)**: recorded at commit (see §E.3 commit signal below after push).
- **E7 (blocker report)**: none. OQ-1 (sessionUrl type) was RESOLVED in-instance (bundled CC v2.1.183 schema confirmed boolean) — the deferral did NOT trigger.

### Files edited (actual count)

- Local `.claude/**` + CLAUDE.md: **19** (8 rules + 9 skills + CLAUDE.md + settings-management.md). NOTE: plan blast-radius listed 16 (15 + CLAUDE.md); actual is 19 because (a) design.md §C.2 required the 2 teammateMode files (worktree-integration.md + moai-workflow-worktree/SKILL.md) which ARE template-mirrored, and (b) settings-management.md was added for M4 `/config` + M5 attribution doctrine.
- Template mirrors: **19** (byte-identical to the 19 above).
- Build artifacts: catalog.yaml (skill-hash manifest regen via make build).
- docs-site: **9** (en/ko/ja/zh × {what-is, intro} = 8 + ko/advanced/hooks-guide = 1).
- README: **4** (README.{md,ko,ja,zh}.md — repo-root project-owned, NOT mirrored).
- settings.json.tmpl: 1 (attribution.sessionUrl added — template-side only, no local counterpart).
- Total: 19 + 19 + 1(catalog) + 9 + 4 + 1(tmpl) = 53 files.

### Residual risk (verification-claim-integrity §3.5)

- The single initial `go test ./...` FAIL did not reproduce on `-count=1` (flaky); all changes are markdown/template content with zero Go behavior change, so a test regression from this SPEC is implausible. Residual: the flaky test (unidentified) could re-fail intermittently in CI — unrelated to this SPEC's edits.
- The template-side `sprint-round-naming.md` mirror carries an older (Sprint/Round v2.0.0) taxonomy than the local Epic-taxonomy version — this is a PRE-EXISTING template-vs-local drift UNRELATED to this SPEC and was NOT touched (scope discipline).

## §E.4 Sync-phase Audit-Ready Signal

### Sync-phase completion (post-implementation)

sync_commit_sha: 4533fc093

- **Sync commit SHA**: `4533fc093` (Authored-By-Agent: manager-docs) — human-readable reference

- Frontmatter transition applied: `in-progress → implemented → completed` (spec.md §A frontmatter `status:` field updated, `updated:` field refreshed to sync-phase date 2026-06-20).
- CHANGELOG.md entry added to `[Unreleased]` section (18 ACs documented, run-phase REQ-GO-001 assessment recorded, OQ-1 resolved, 21 REQs total, 3-axis scope confirmed).
- README and docs-site locale parity verified (no stale TeamCreate/TeamDelete live-action instructions remain).
- Template-mirror parity confirmed (all 19 doctrine files + 19 template mirrors byte-identical per make build catalog regeneration).
- Commit hygiene: specific-path git add (CHANGELOG + progress.md + spec.md frontmatter transition); NO `git add -A` / `git add .`.
- Cross-platform build verified: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Lint clean: `golangci-lint run --timeout=2m` → 0 issues (no NEW issues vs baseline 0).
- Quality gate: TestTemplateNeutralityAudit + spec-lint 0 errors.

**Sync commit deliverable**: CHANGELOG.md entry + progress.md §E.4 + spec.md frontmatter (status + updated)
