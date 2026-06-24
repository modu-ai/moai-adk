# SPEC-WEB-CONSOLE-008 — Progress

> statusline config redesign (honest hybrid). Tier M, cycle_type=tdd.
> Ground-truth: `.moai/reports/web-console-statusline-gitconvention-audit.md` (statusline 17-defect inventory + Redesign Proposal).

## §A Plan-phase

- plan_complete_at: 2026-06-07
- plan_status: audit-ready
- tier: M
- cycle_type: tdd
- plan-auditor: iter-1 0.84 PASS-WITH-DEBT → 6 orchestrator-direct patches (D1/D2/D3/D5/D6/D7) → iter-2 0.87 PASS-WITH-DEBT (monotonic +0.03, MP-1..4 PASS); D-r1 MINOR (AC-WC8-010 decorative clause) cleaned orchestrator-direct.
- GATE-2: user-approved run-phase entry (2026-06-07).

## §E — Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope (file count): ~12-16 (internal/cli, internal/profile, internal/statusline, internal/web, internal/config, pkg/models, internal/template/templates)
- domain count: 1 (statusline config — single domain, multi-package)
- file language mix: Go + Templ + YAML
- concurrency benefit: LOW (coding-heavy; sequential milestone dependencies M1→M9, Finding A4 caveat)
- Agent Teams prereqs status: NOT met (harness level not `thorough` / team env unconfirmed)

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-package semantic change, not a typo/single-line |
| 2 background | no | Write/Edit required (CONST-V3R2-020 — background auto-denies writes) |
| 3 agent-team | no | REQ-ATR-013 prereqs not all met AND single-domain (<3 domains) |
| 4 parallel | no | coding-heavy, not research-heavy (Finding A4 caveat) |
| 5 sub-agent | YES | coding-heavy default; sequential M1→M9 dependency chain |
| 6 workflow | no | <30 files AND not a single uniform mechanical transform (multi-rule semantic) |

Decision: sub-agent (Mode 5)

Justification: SPEC-WEB-CONSOLE-008 is coding-heavy, single-domain (statusline config), Tier M, ~12-16 files with a sequential milestone dependency chain (M1 template schema → M3 struct Mode removal → M4 preset write-effective → M9 symmetry guard LAST). Per Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. GATE-2 (plan→run HUMAN GATE) passed before this selection — Mode 5 is strictly downstream of GATE-2.

## §E.2 — Run-phase Evidence / Audit-Ready Signal

- run_complete_at: 2026-06-07
- cycle_type: tdd (RED→GREEN→REFACTOR)
- mode: sub-agent (Mode 5, sequential M1→M9), single implementer manager-develop
- run_branch: feat/SPEC-WEB-CONSOLE-008 (run baseline 808bc3fb0; L1 worktree reset --keep onto plan HEAD; M1-M9 stacked)

### Per-milestone commits

| Mn | SHA | Subject |
|----|-----|---------|
| M1 | 1b37dc2fa | template schema correction (drop mode:/refresh_interval:, theme→catppuccin-mocha) — flips status draft→in-progress |
| M2 | 34cedba62 | seed 15-key completion (defaultStatuslineSegments via statusline.Segment* constants) |
| M3 | 3a8a0ab51 | 3-struct consolidation (Mode removal; Builder API PRESERVE; canonical no Mode) |
| M4 | 09c3736f4 | preset write-effective (NEW statusline.PresetToSegments SSOT; cli delegation; profile write expansion) |
| M5 | 08133b16d | remove dead loadSegmentConfig + 6 test sites |
| M6 | a82b0a039 | doc/comment correction (builder ThemeName 2 themes; SegmentRepo exclusion comment) |
| M7 | ed82d13e6 | remove web statusline_mode control + count label 3→2 fields |
| M8 | 2b0dab9a9 | web conditional segment exposure (disabled when preset≠custom) |
| M9 | dec809a91 | StatuslineConfig symmetry CI guard (LAST) |

### AC PASS/FAIL matrix (22 AC, all PASS)

| AC | Status | Verification | Actual |
|----|--------|--------------|--------|
| AC-WC8-001 | PASS | `grep -cE '^[[:space:]]*(mode\|refresh_interval):' template statusline.yaml` | 0 |
| AC-WC8-002 | PASS | theme seed catppuccin-mocha + no `default` on theme lines | grep PASS |
| AC-WC8-003 | PASS | profile/cli Mode count 0 + canonical Preset + NormalizeMode + Render present | PASS |
| AC-WC8-004 | PASS | web statusline_mode/StatuslineModes/statuslineModeCanonical non-comment refs | 0 |
| AC-WC8-005 | PASS | loadSegmentConfig non-comment refs in internal/cli | 0 |
| AC-WC8-006 | PASS | renderFullV3 production call sites | 0 |
| AC-WC8-007 | PASS | `go test ./internal/profile -run TestDefaultStatuslineSegments\|TestSyncStatusline` | ok (15 keys) |
| AC-WC8-008 | PASS | `go test -run TestSyncStatuslinePresetExpand\|TestStatuslinePresetEffective` | ok (compact/minimal/full expand) |
| AC-WC8-009 | PASS (must-remain-green) | `go test ./internal/web -run TestRealServerRoundTrip\|TestGoldenPath` | ok (theme-only preserve, HARD-7) |
| AC-WC8-010 | PASS | builder ThemeName catppuccin + no nerd + SegmentRepo exclusion comment | PASS |
| AC-WC8-011 | PASS | fieldsets.templ "2 fields" present + "3 fields · segments" absent | PASS |
| AC-WC8-012 | PASS | `go test ./internal/web -run TestFieldsetStatuslineConditionalSegments` | ok (custom editable; non-custom disabled) |
| AC-WC8-013 | PASS | `templ generate -path ./internal/web && git diff --exit-code *_templ.go` | clean (canonical -path form; see note) |
| AC-WC8-014 | PASS | `grep StatuslineConfig{}` + `go test -run TestStructYAMLSymmetry` | ok (5/5 symmetry) |
| AC-WC8-015 | PASS | git_convention added-line tokens (added lines) | 0 |
| AC-WC8-016 | PASS | `grep -cE 'func validate(Workflow\|GitStrategy\|Harness\|Llm\|Statusline)Config'` | 0 |
| AC-WC8-017 | PASS (must-remain-green) | 006 sentinel integration_test.go:197-205 added workflow/harness/git-strategy lines | 0 (byte-identical, HARD-3) |
| AC-WC8-018 | PASS (must-remain-green) | offline CDN guard (unpkg/jsdelivr/cdn/googleapis/gstatic) | 0 (HARD-8) |
| AC-WC8-019 | PASS | web layer direct yaml.Marshal/os.WriteFile (non-comment) | 0 |
| AC-WC8-020 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 |
| AC-WC8-021 | PASS | template-is-embed (//go:embed all:templates; embedded.go retired) + neutrality test | ok |
| AC-WC8-022 | PASS | full suite (web/cli/config/profile/statusline/pkg/models) `-count=1` | 0 FAIL |

### HARD invariants (all upheld)

- HARD-1 Builder API PRESERVE: StatuslineMode enum / NormalizeMode / Render(data, mode) / Builder Config Mode field / SetMode unchanged; builder_test.go mode-collapse + full-retired tests PASS unmodified.
- HARD-2 renderFullV3 NOT resurrected: 0 production call sites.
- HARD-3 006 sentinel (integration_test.go:197-205): byte-identical (0 diff vs origin/main).
- HARD-4 git_convention CRITICAL SCOPE: `git diff origin/main -- internal/git/convention/` = 0 lines.
- HARD-5 canonical models.StatuslineConfig = {Preset, Segments, Theme}; no Mode added.
- HARD-6 runtime precedence (segments map wins over preset) unchanged; M4 is write-time materialization only.
- HARD-7 theme-only round-trip GREEN (AC-WC8-009).
- HARD-8 offline zero-network preserved (AC-WC8-018).
- HARD-9 template-first + neutrality: template edit + neutrality test GREEN; comments generic.

### Notes

- AC-WC8-013 codegen: the canonical drift-free command is `templ generate -path ./internal/web` (Makefile `templ-generate` target + `//go:generate` in internal/web/templ.go). Running bare `templ generate` from repo root path-qualifies the embedded `FileName` (`internal/web/page.templ` vs the committed basename `page.templ`) and produces spurious page_templ.go/root_templ.go diffs; the project SSOT is the `-path` form, which regenerates ONLY fieldsets_templ.go for the M7/M8 source edits.
- Coverage (`go test -cover`, total incl. codegen dilution): internal/web 72.1%, internal/statusline 83.0%, internal/profile 80.5% — consistent with the 006/007 total-coverage baseline; no regression.
- M4 SSOT relocation: new internal/statusline/preset.go (PresetToSegments + CanonicalSegments); cli presetToSegments/allStatuslineSegments became thin delegations (cli tests + call sites byte-stable). An expander, not a validator (AC-WC8-016 unaffected).

## §E.3 — Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-07
run_commit_sha: dec809a91
run_status: implemented
ac_pass_count: 22
ac_fail_count: 0
preserve_list_post_run_count: 0   # HARD-1..9 PRESERVE targets all intact (Builder API, 006 sentinel, git_convention, canonical model)
l44_pre_commit_fetch: not-applicable   # L1 isolated agent worktree; run baseline = plan HEAD 808bc3fb0 (origin/main 54c09a61b ancestor, FF-clean)
l44_post_push_fetch: pending           # push performed at run-phase close (HEAD:main FF)
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues across all changed packages
cross_platform_build:
  host: pass        # go build ./... exit 0
  windows: pass     # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 22   # 11 production (cli/profile/statusline x3 / template yaml / web x3 / preset.go new) + generated _templ.go x1 + test files + progress.md
m1_to_mN_commit_strategy: per-milestone   # M1..M9 each a separate feat(SPEC-WEB-CONSOLE-008): M{N} commit with Authored-By-Agent: manager-develop trailer; M1 flipped status draft->in-progress
```

## §E.4 — Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-07
sync_commit_sha: 591a310de   # orchestrator-direct sync (trailer omitted = OwnershipTransitionRule silent SKIP per cohort 006/007 pattern)
sync_status: implemented   # spec.md frontmatter in-progress -> implemented
deliverables:
  - CHANGELOG.md   # [Unreleased] ### Changed: SPEC-WEB-CONSOLE-008 entry (no duplicate — prior match was 007's forward-reference)
  - .moai/specs/SPEC-WEB-CONSOLE-008/spec.md   # status implemented
readme_docs_site: not-applicable   # internal config redesign; no user-facing API surface beyond the moai web console UI
independent_verification: orchestrator Trust-but-verify 7/7 (V1-V7) + precise V2/V6 false-positive clearance + full focused suite GREEN + cross-platform exit 0
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-06-07
mx_commit_sha: 45fdd9a5b   # Mx close commit (implemented -> completed, orchestrator-direct canonical owner)
mx_status: completed     # spec.md frontmatter implemented -> completed (4-phase close)
four_phase_close: plan(808bc3fb0) + run(M1-M9 1b37dc2fa..dec809a91) + sync + Mx
drift_expected: completed/completed aligned   # era H-4 V3R6 (progress.md §E.2 + sync_commit_sha + §E.5 + mx_commit_sha)
followups:
  - SPEC-WEB-CONSOLE-009   # git_convention redesign (25 deferred defects)
  - GCR-5 dormant-engine wiring   # maintainer-decision (pre-push hook never invokes moai hook pre-push) — separate follow-up
```
