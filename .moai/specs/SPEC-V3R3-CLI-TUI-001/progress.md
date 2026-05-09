## SPEC-V3R3-CLI-TUI-001 Progress

- Started: 2026-05-09 (run phase entry from paste-ready resume)
- Branch: feat/SPEC-V3R3-CLI-TUI-001 (worktree cli-tui-v2)
- Methodology: TDD (per quality.yaml development_mode)
- Harness: standard (default)
- Total Milestones: 7 (M1~M7 sequential)

### Phase 0.5: Plan Audit Gate (PASS)

- Cache check: MISS (no prior daily report)
- Auditor invocation: completed (plan-auditor iteration 1/3)
- Verdict: PASS (overall_score: 0.82)
- Must-pass: 4/4 PASS (REQ consistency, EARS format, YAML frontmatter, language neutrality)
- Category scores: Clarity 0.85 · Completeness 0.90 · Testability 0.78 · Traceability 0.95
- Should-fix defects: 9 (D1-D9)
  - D1/D2 (major): hardcoded `/Users/goos/.moai/worktrees/moai-adk/cli-tui-v2` (case-mismatch + user-specific path) at acceptance.md L22/266/339/376/424/449/485/520 + plan.md L60. Linux CI risk.
  - D3 (major): acceptance.md:L298 conflates programming-language neutrality with human-language i18n
  - D4 (major): no global hex-sweep AC for REQ-013 (only banner.go scope via AC-002)
  - D5 (major): no global emoji-sweep AC for REQ-014 (only init scope via AC-004)
  - D6 (minor): OQ2 (huh radio `◆/◇`) gates AC-004; resolve before M5
  - D7 (minor): REQ-013..016 SHALL NOT phrasing under Unwanted header (form/label mismatch)
  - D8 (minor): OQ3 glamour cache deferral conflicts with AC-003 ">=19 항목" floor
  - D9 (minor): AC-006 Windows cmd.exe YAML snippet not committed to `.github/workflows/test.yml`
- Plan artifact hash: 5d9eea3827f77a1e55a97866438536c54b9dc40d58dba0ae6cfc9966cff5f9f6
- Audit report: .moai/reports/plan-audit/SPEC-V3R3-CLI-TUI-001-review-1.md
- audit_at: 2026-05-09T08:18:37Z
- auditor_version: plan-auditor (iteration 1/3, total tokens 157584)

### Daily Audit Cache Persistence (Phase 0.5 Step 5)

- Daily report: `.moai/reports/plan-audit/SPEC-V3R3-CLI-TUI-001-2026-05-09.md` (created)
- audit_verdict: PASS
- audit_report: .moai/reports/plan-audit/SPEC-V3R3-CLI-TUI-001-review-1.md
- audit_at: 2026-05-09T08:18:37Z
- auditor_version: plan-auditor (iteration 1/3)

### D1-D5 SPEC 정정 (User decision: immediate fix)

- D1+D2: 9 hardcoded paths → `$(git rev-parse --show-toplevel)` (acceptance.md replace_all + plan.md L60)
- D3: AC-CLI-TUI-009 rewritten (i18n vs programming-language taxonomy)
- D4: AC-CLI-TUI-016 added (global hex sweep)
- D5: AC-CLI-TUI-017 added (global emoji codepoint sweep)
- DOD updated: 001~015 → 001~017 ACs
- New plan_artifact_hash: 39d853cd40b5cae85a0a1ad94bc89ea6371454def0c20ca6c1167691cd21e095
- Deferred OQs: D6 (M5 entry), D7 (sync), D8 (M4 step 4b), D9 (M7/sync)

### PR Strategy Decision (User)

- Per-milestone 7 PRs sequential (lessons #9 wave-split, Anthropic SSE stall avoidance)
- This session scope: SPEC 정정 (D1-D5) + M1 only
- Subsequent sessions: M2, M3, ..., M7 each in separate PR

---

## Phase 1: Analysis and Planning (M1 scope)

- Phase 1 Strategy: plan.md §2 (M1) is treated as the analysis output. plan-auditor PASS at 0.82 carries authority.
- Phase 1.5 Task Decomposition: M1 step list per plan.md §10:
  - M1-T-001: theme.go + theme_test.go (28 tokens light/dark, REQ-CLI-TUI-002)
  - M1-T-002: box.go + box_test.go (Box/ThickBox lipgloss API, 8 golden snapshots)
  - M1-T-003: pill.go + pill_test.go (6 variants × 2 themes × 2 solid = 24 snapshots)
  - M1-T-004: doc.go (godoc + design source attribution) + runeguard.go (한글 폭 헬퍼)
- Phase 1.6 AC Failing Checklist: M1 directly maps to AC-CLI-TUI-001, partial AC-007/011/014/016
- Phase 1.7 File Scaffolding: delegated to manager-tdd (RED phase creates stubs + failing tests)
- Phase 1.8 MX Context Scan: SKIP (greenfield — `internal/tui/` does not yet exist)

### Phase 2B: TDD Implementation (M1) — COMPLETE

- Methodology: TDD RED-GREEN-REFACTOR (per quality.yaml development_mode: tdd)
- Agent: manager-tdd subagent
- Worktree isolation: NOT applied (main session cwd is already SPEC worktree `cli-tui-v2`; nested isolation would create base mismatch per lessons #13)
- Status: COMPLETE — merged via PR #803 (commit `6df7d140d`, 2026-05-09T10:37:58Z)
- Artifacts shipped:
  - `internal/tui/theme.go` + `theme_test.go` (28 light/dark tokens, REQ-CLI-TUI-002)
  - `internal/tui/box.go` + `box_test.go` (Box / ThickBox + 8 golden snapshots)
  - `internal/tui/pill.go` + `pill_test.go` (6 variants × 2 themes × 2 solid = 24 snapshots)
  - `internal/tui/doc.go` (godoc + `source/project/tui.jsx` attribution)
  - `internal/tui/internal/runeguard.go` + `runeguard_test.go` (ko-width helper)
  - 35 testdata/*.golden files committed
- Follow-up shipped same wave: runeguard EastAsianWidth=true locale fix + Windows flaky test mitigations (PRs #802 / #803)
- Open follow-ups (deferred to later milestones, not blocking M2): fixture leak partial fix (PR #795 carry-over), docs-i18n-check warnings, Windows TestObserver_TickPersistence flaky

### Phase 0.5: Plan Audit Gate (M2 entry) — CACHE HIT

- Source cache: `.moai/reports/plan-audit/SPEC-V3R3-CLI-TUI-001-2026-05-09.md` (Run 1 of 1, verdict=PASS)
- audit_cache_hit: true
- cached_audit_at: 2026-05-09T08:18:37Z (~12h ago, within 24h window)
- plan_artifact_hash carry-forward: `39d853cd40b5cae85a0a1ad94bc89ea6371454def0c20ca6c1167691cd21e095` (post D1-D5)
- Hash match basis: `git rev-list --left-right --count origin/main...HEAD = 0/0` → SPEC files (`spec.md` / `plan.md` / `acceptance.md`) unchanged since M1 entry; recompute would yield same hash.
- Decision: skip Step 3 plan-auditor re-invocation, proceed directly to Phase 1 per run.md Step 2.
- Note: M1 implementation modified `internal/tui/*` only — outside SPEC artifact set; cache lookup hash basis (spec/plan/acceptance/tasks) intact.

---

## Phase 1: Analysis and Planning (M2 scope) — ENTRY

- Phase 1 Strategy: plan.md §3 (M2) is the analysis output. plan-auditor PASS 0.82 (M1 entry) + main↔branch diff=0 carry forward; Phase 0.5 cache HIT expected.
- Phase 0.95 Mode: Standard Mode (12 files: 6 source + 6 tests, single domain `internal/tui/`)
- Phase 1.5 Task Decomposition (6-step from plan.md §3.1, 10):
  - M2-S1: status.go + status_test.go (StatusIcon / Spinner / Progress / Stepper)
  - M2-S2: form.go + form_test.go (RadioRow / CheckRow huh helpers)
  - M2-S3: table.go + table_test.go (KV / CheckLine / Section)
  - M2-S4: prompt.go + prompt_test.go (Prompt / Cursor)
  - M2-S5: term.go + term_test.go (Term chrome, MOAI_SCREENSHOT=1 only)
  - M2-S6: help.go + help_test.go (HelpBar [KeyHint])
- Phase 1.6 AC Failing Checklist for M2:
  - AC-CLI-TUI-001 (extends to 6 of 9 component files; full coverage after M2 merge)
  - AC-CLI-TUI-007 mixed 18 ko-en cases — M2 enables, full validation per plan.md §2.3
  - AC-CLI-TUI-011 no hand-drawn box characters in production code (continuous)
  - AC-CLI-TUI-015 prefers-reduced-motion static fallback (M2 starting point — Spinner / Progress in S1)
- Phase 1.7 Scaffolding: delegated to manager-tdd (RED phase creates stubs)
- Phase 1.8 MX Context Scan: SKIP (greenfield additions to `internal/tui/`)

### Phase 2B: TDD Implementation (M2) — COMPLETE

- Methodology: TDD RED-GREEN-REFACTOR (committed per plan.md M2 row)
- Agent: manager-tdd subagent (single delegation, 6-step internal TodoList)
- Worktree isolation: NOT applied (lessons #13 — same rationale as M1)
- Status: COMPLETE — 2026-05-09T20:47:00Z (estimated)
- Artifacts shipped:
  - `internal/tui/status.go` + `status_test.go` (StatusIcon/Spinner/Progress/Stepper; AC-CLI-TUI-015 reduced-motion fallback; 13 golden snapshots)
  - `internal/tui/form.go` + `form_test.go` (RadioRow/CheckRow; ◆/◇ prefix; 5 golden snapshots)
  - `internal/tui/table.go` + `table_test.go` (KV/CheckLine/Section; 18 mixed ko-en × 2 themes = 36 snapshots; plus 7 component snapshots = 43 total)
  - `internal/tui/prompt.go` + `prompt_test.go` (Prompt/Cursor; pure functions; 7 golden snapshots)
  - `internal/tui/term.go` + `term_test.go` (Term; MOAI_SCREENSHOT=1 gate; 2 screenshot golden snapshots)
  - `internal/tui/help.go` + `help_test.go` (HelpBar/KeyHint; ko-en mixed; 4 golden snapshots)
- Total M2 golden snapshots: 74 new files (total testdata/ count: 106)
- AC coverage newly satisfied:
  - AC-CLI-TUI-007: 18 mixed ko-en cases × 2 themes = 36 snapshots GREEN
  - AC-CLI-TUI-015: MOAI_REDUCED_MOTION=1 → Spinner(●) + Progress(filled) static fallback GREEN
  - AC-CLI-TUI-011: zero hand-drawn box chars in all 6 M2 source files
  - AC-CLI-TUI-013: zero hex literals in all 6 M2 source files
  - AC-CLI-TUI-017: zero emoji codepoints in all 6 M2 source files
  - AC-CLI-TUI-001: 9/9 component files now exist (theme+box+pill+status+form+table+prompt+term+help)
- Implementation Divergence:
  - Section() uses ASCII dash "-" as rule instead of lipgloss.Border() — box chars (U+2500 ─) are forbidden by AC-CLI-TUI-011 in source files; plan implied a visual rule but the constraint takes precedence
  - Spinner uses Braille U+2808 (⠋) for animated frame — not "animated" in the goroutine sense; plan §3.3 confirmed stateless only; frame character is caller-re-rendered
  - StatusIcon returns single glyphs without lipgloss styling — caller applies styling via CheckLine
- Quality gate:
  - `go test ./internal/tui/... -count=1 -race` GREEN
  - `go vet ./internal/tui/...` CLEAN
  - `golangci-lint run ./internal/tui/...` CLEAN
  - `go mod tidy` produces zero diff
- Open follow-ups: none blocking M3

### Phase 0.5: Plan Audit Gate (M3 entry) — CACHE HIT

- Source cache: `.moai/reports/plan-audit/SPEC-V3R3-CLI-TUI-001-2026-05-09.md` (Run 1 of 1, verdict=PASS)
- audit_cache_hit: true
- cached_audit_at: 2026-05-09T08:18:37Z (~4h ago, within 24h window)
- plan_artifact_hash carry-forward: `39d853cd40b5cae85a0a1ad94bc89ea6371454def0c20ca6c1167691cd21e095`
- Hash match basis: M2 implementation (`34a92d4d8`) modified `internal/tui/*` only — outside SPEC artifact set; spec/plan/acceptance unchanged since cache write.
- Decision: skip Step 3 plan-auditor re-invocation, proceed directly to Phase 1 per run.md Step 2.
- M3 branch base: `feat/SPEC-V3R3-CLI-TUI-001-m3` (forked from `origin/main` after PR #806 admin squash merge at 2026-05-09T12:19:19Z).

---

## Phase 1: Analysis and Planning (M3 scope) — ENTRY

- Phase 1 Strategy: plan.md §4 (M3) is the analysis output. Phase 0.5 cache HIT carry forward.
- Phase 0.95 Mode: Standard Mode (2 source/test files: `internal/cli/banner.go` + `banner_test.go` plus testdata/*.golden, single domain `internal/cli/`)
- Methodology: DDD (per plan.md §1 milestone table, M3 row "DDD characterize")
- Phase 1.5 Task Decomposition (2-step from plan.md §4.2 / §4.3):
  - M3-S1: ANALYZE + PRESERVE — entry-point grep + characterization tests (TestBanner_Current_Light/Dark/NoColor + TestWelcome_Current_Light/Dark/NoColor)
  - M3-S2: IMPROVE — terra cotta hex (`#C45A3C/#DA7756`) 제거 + 보라 hex (`#5B21B6/#7C3AED`) 제거 + `tui.Theme().Accent` 사용 + `tui.Pill` 3개 추가 + new golden snapshots
- Phase 1.6 AC Failing Checklist for M3:
  - AC-CLI-TUI-013 (no hex literals outside `internal/tui/`) — `banner.go` currently violates with 3 hex literals (L25 terra cotta light/dark, L45 보라 light/dark)
  - AC-CLI-TUI-016 (global hex sweep) — M3 partial coverage (banner.go); full coverage at M7
  - AC-CLI-TUI-001 visual consistency (M3 introduces `tui.Pill` into banner per `screens.jsx:ScreenBanner`)
  - AC-CLI-TUI-011 zero hand-drawn box chars (preserved — banner.go does not use box chars; ASCII art exempt)
- Phase 1.7 Scaffolding: delegated to manager-ddd (Step 1 PRESERVE creates characterization tests; Step 2 IMPROVE replaces banner body)
- Phase 1.8 MX Context Scan: required (existing `internal/cli/banner.go` has callers — ANALYZE step performs grep `PrintBanner|PrintWelcomeMessage`)
- Worktree isolation: NOT applied (main session cwd is already SPEC worktree `cli-tui-v2`; nested isolation would create base mismatch per lessons #13)

### Surfaced Assumptions (Behavior 1)

1. `PrintBanner` / `PrintWelcomeMessage` exported signatures unchanged — only body output evolves; 4 entry points (init/update/version/doctor) call sites unaffected.
2. Characterization snapshots (Step 1 testdata/banner-current-*.golden + welcome-current-*.golden) are committed normally and **replaced** by new tui-derived snapshots in Step 2 (plan.md §4.2 "git stash" 표현은 폐기 의미로 해석).
3. NO_COLOR=1 환경에서 8-line ASCII art는 그대로 출력, 색만 제거 (terra cotta + 보라 모두 `lipgloss.NoColor{}` 또는 빈 Foreground 처리).
4. Scope boundary (Behavior 5): `internal/cli/banner.go` + `internal/cli/banner_test.go` + `internal/cli/testdata/banner-*.golden` + `internal/cli/testdata/welcome-*.golden` — outside files (init.go, update.go, version.go, doctor.go) untouched.
5. `internal/tui/theme.go`의 `Theme().Accent` (deep teal — tui.jsx light=`#144a46` / dark=`#7dd3c0`)는 banner 헤드라인 강조에 적합. caller는 `internal/tui/theme.go::Theme(env)`로 라이트/다크 자동 선택.
6. `tui.Pill` 3개는 design source `screens.jsx:ScreenBanner` 기준 — labels: "v" + version, "go " + go runtime version, "claude " + claude code version. 동적 fetch (runtime/debug.ReadBuildInfo for version, runtime.Version() for go) 또는 정적 placeholder 결정은 Step 2 시점에서 결정.
7. PrintBanner는 `io.Writer` 인자 형태 유지 (existing signature 추정 — Step 1 ANALYZE에서 확정).

### Phase 2B: DDD Implementation (M3) — COMPLETE

- Methodology: DDD ANALYZE-PRESERVE-IMPROVE (committed per plan.md M3 row "DDD characterize")
- Agent: manager-ddd subagent (2 sequential delegations: Step 1 PRESERVE → Step 2 IMPROVE)
- Worktree isolation: NOT applied (lessons #13)
- Status: COMPLETE — 2026-05-09T~21:40:00Z (estimated)

#### M3-S1 (PRESERVE) outputs:
- ANALYZE: 5 entry points confirmed — `root.go:25`, `init.go:244+245`, `update.go:1936+1942` (예상 ~4 → 실제 5: update.go에서 PrintBanner + PrintWelcomeMessage 모두 호출)
- PRESERVE: `internal/cli/banner_test.go` +152 lines (6 characterization tests: `TestBanner_Current_{Light,Dark}`, `TestBanner_NoColor`, `TestWelcome_Current_{Light,Dark}`, `TestWelcome_NoColor`)
- 6 initial golden snapshots committed under `internal/cli/testdata/banner-current-{light,dark,nocolor}.golden` + `welcome-current-{light,dark,nocolor}.golden`
- New finding: `os.Pipe()` 캡처 시 `lipgloss`는 비-TTY 감지 → ANSI escape 자동 strip → light/dark goldens byte-identical, NO_COLOR-mode goldens differ (Pill character form 차이)

#### M3-S2 (IMPROVE) outputs:
- `internal/cli/banner.go` rewrite (LOC delta -57 → +90, net +33):
  - 신규 헬퍼 3개: `resolveTheme()` (NO_COLOR/MOAI_THEME → tui.Theme), `goVersion()` (`runtime.Version()` → `1.x.y`), `claudeVersion()` (env CLAUDE_CODE_VERSION → fallback "claude")
  - `PrintBanner(version string)` — terra cotta `#C45A3C/#DA7756` 제거, `tui.Theme.Accent` 사용, 8-line ASCII art 보존, 3 `tui.Pill` row 추가 (L72-74)
  - `PrintWelcomeMessage()` — 보라 `#5B21B6/#7C3AED` 제거, `tui.Theme.Accent` 사용, body 텍스트 보존
- 6 characterization goldens replaced via `-update` flag (Step 1 → Step 2 출력 차이 = banner row + Pill 추가)
- @MX:NOTE [AUTO] 태그 추가 (L49) — code_comments=ko 정책 따라 한국어 본문
- AC coverage:
  - AC-CLI-TUI-013: banner.go hex literal 0건 (REQ-CLI-TUI-013 충족)
  - AC-CLI-TUI-016: M3 partial — banner.go global sweep PASS, M7에서 전 패키지 sweep
  - AC-CLI-TUI-001: 9/9 tui components + banner.go integration
- design source consulted: `.moai/design/SPEC-V3R3-CLI-TUI-001/source/project/screens.jsx:180-182` (ScreenBanner 3 pill 정의)
- Pill labels chosen: `PillPrimary, Solid=true, "v{version}"` / `PillOk, Solid=false, "go {goVer}"` / `PillInfo, Solid=false, claudeVer`
- Public signatures unchanged: `PrintBanner(version string)`, `PrintWelcomeMessage()` — 5 callers (root/init/update) 무영향
- Quality gate:
  - `grep -nE '#[0-9a-fA-F]{6}' internal/cli/banner.go` empty
  - `go test ./internal/cli/ -count=1 -race` GREEN (16.3s)
  - `go test ./... -count=1 -short` exit 0 (no cross-package regressions)
  - `go vet ./internal/cli/...` clean
  - `golangci-lint run ./internal/cli/...` 0 issues
- New assumptions surfaced (Step 2):
  - 8: `runtime.Version()`은 항상 `"go"` prefix → `strings.TrimPrefix` 안전
  - 9: light/dark goldens byte-identical은 의도 동작 (pipe ANSI strip); env별 별도 보존은 Step 2 이후 다양한 출력 시 즉시 catch
  - 10: claudeVersion() fallback "claude" (no version)는 `CLAUDE_CODE_VERSION` env 미설정 시 안전 fallback (테스트 deterministic)
- Open follow-ups (deferred to later milestones, not blocking M4):
  - banner.go L49 godoc/comments 영문 유지는 16-language neutrality 일관성 위함 (mx-tag-protocol L113 @MX 태그만 code_comments=ko 적용)



