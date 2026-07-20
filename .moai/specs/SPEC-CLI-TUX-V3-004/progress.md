# SPEC-CLI-TUX-V3-004 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~10-15 files (internal/cli + uikit), domains=1 (Go CLI rendering), language=Go, concurrency benefit LOW (coding-heavy).
- Evaluation: trivial no (multi-milestone) / background no (write-capable) / agent-team RETIRED / parallel no (single domain, coding-heavy) / workflow no (<30 files, non-uniform transforms) / sub-agent selected.
- Decision: sub-agent
- Justification: Coding-heavy single-domain Tier M SPEC per Anthropic coding-task parallelism caveat — sequential manager-develop (cycle_type=tdd) per milestone is the default and correct envelope. Plan-audit iter-1 PASS-WITH-DEBT 0.87 (2026-07-20); Implementation Kickoff Approval obtained (user selected run-phase entry).

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C #1-#7, executed 2026-07-20)

- **#1 baseline**: branch `main`, spawn-time HEAD `db70597a2f07cb83e3188413d6dcbc5d3eea5e94`; re-based mid-run to `319c3e93e`/`dd0a6ce72` as the parallel session (SPEC-MODEL-PROFILE-MATRIX-001 M2/M3 + README restructure) landed commits. Divergence re-checked before every commit (`git rev-list --count --left-right origin/main...HEAD` = `0 0`).
- **#2 depends_on (strict completed)**: SPEC-CLI-TUX-V3-002 `status: completed`, SPEC-CLI-TUX-V3-003 `status: completed` — PASS.
- **#3 cross-platform build + lint baseline**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=5m` → `0 issues.` (lint baseline = 0; any finding at M4e is NEW). Logs: `.moai/state/verify/tux4/preflight-{build-native,build-win,lint}.log`.
- **#4 golden green anchor**: `go test ./internal/cli/ -run 'Golden' -count=1` → exit 0 (`[no tests to run]` — no existing test name contains "Golden"; existing characterization tests are `TestDoctor_Current_*`/`TestStatus_Current_*`, verified green in the full-package baseline). `go test ./internal/cli/uikit/ -count=1` → ok (exit 0).
- **#5 ratchet inventory re-measure (audit-debt D1/D2 resolution)**: raw grep = **14** (plan.md "전체 40" was stale). Breakdown: `internal/cli/uikit/banner.go` ×12 (in-scope, REQ-TUX4-006), `internal/cli/branch_protection.go:44` ×1 (out-of-scope, see D2), `internal/cli/worktree/new.go:433` ×1 (commented-out line, see D1). Inventory files: `.moai/state/verify/tux4/preflight-ratchet-{raw,nocomment}.txt`.
  - **D1 resolution**: the series-final ratchet grep (AC-TUX4-010) appends the comment-exclusion idiom `| grep -v '^[^:]*:[0-9]*:[[:space:]]*//'` (same as AC-TUX4-005). The out-of-scope comment in `worktree/new.go:433` is NOT deleted to satisfy a naive grep. Comment-excluded count at pre-flight = **13**.
  - **D2 resolution**: `internal/cli/branch_protection.go:44` (interactive y/N `fmt.Printf`) is owned by deferred SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 and is OUTSIDE this SPEC's §E commit scope. The final ratchet-0 gate is scoped to the §E surface set (doctor/status/spec/root/uikit + pre-flight #5 inventory minus branch_protection.go). Residual: **PASS-WITH-DEBT** — 1 remaining call site, deferred to its owning SPEC. branch_protection.go is not modified by this SPEC.
- **#6 glamour version + dependency graph**: latest glamour = `v1.0.0`. Its go.mod requires `github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834` (**lipgloss v1 line**). The repo graph ALREADY carries the v1/v2 coexistence (`github.com/charmbracelet/lipgloss v1.1.0` direct + `charm.land/lipgloss/v2 v2.0.5`), so glamour introduces NO new major — it bumps the existing v1 entry to v1.1.1-pre. No graph re-expansion; pinned `glamour@v1.0.0`.
- **#7 coverage baseline (REQ-TUX4-011 downgrade decision)**: `internal/cli` **74.3%**, `internal/template` **85.2%**, `internal/hook` **83.5%** (sub-packages 79.5-100%). **All three below 90%** → per REQ-TUX4-011's built-in relief, the gate degrades to **strict non-regression** for all three packages, gap recorded here: cli −15.7pt, template −4.8pt, hook −6.5pt vs the 90% target. Log: `.moai/state/verify/tux4/preflight-cover.log`.

### M4a — glamour introduction + style decision (REQ-TUX4-004, AC-TUX4-005)

- **Dependency**: `go get github.com/charmbracelet/glamour@v1.0.0` + import in `internal/cli/glamour_style.go` + `go mod tidy` → direct requirement in go.mod. go.mod/go.sum diff verified pure (glamour-induced entries only) before staging — shared-file discipline with the parallel session.
- **Style file path (AC-TUX4-005 convention)**: `internal/cli/glamour_style.go` — matches the acceptance.md path convention exactly; no substituted verification command needed.
- **Token mapping decision** (no hex — all colours reference `tui.Theme` fields): Document/Text/Item/Enumeration/CodeBlock/Table → `Body`; Heading/H1-H3 → `Accent` (bold); Strong → `Fg` (bold); Emph → `Body` (italic); Link/LinkText/inline Code → `Info` (link underlined); BlockQuote → `Dim`; HorizontalRule → `Rule`.
- **Render routing (REQ-TUX4-005)**: rich glamour path only when NO_COLOR unset AND destination writer is a character-device `*os.File`; all other cases (pipes, CI, golden-test buffers) are byte-stable plain markdown passthrough. Fixed word-wrap width 100 (`glamourWordWrap`) — no live terminal-width probing, keeping golden line-wrap stable (acceptance §C width edge).
- **Shared render symbols (AC-TUX4-004 wiring record)**: `renderMarkdown(w, md)` is the glamour-mediated gateway; `glamourRender(md, theme)` is the always-rich body; `markdownRichEnabled(noColor, tty)` is the routing predicate. M4b wires `status.go` and `spec_view.go` through these symbols.
- **TDD**: RED `internal/cli/glamour_style_test.go` (token-mapping, light/dark divergence, routing matrix, passthrough byte-stability, rich-path ANSI presence) → GREEN `internal/cli/glamour_style.go`.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
