# SPEC-HANDOFF-RESUME-ASK-001 — Implementation Plan

## A. Context

Card t1672 (Class C design change, P1, operator origin; operator approved the design 2026-10-10, board `d-20261010T113836Z-a35a`). Tree `8249bff8c3933d647b4f99e59048595f9808e83a`, branch `WT-clear-cc-codex`. Requirements and the re-read code facts are in `spec.md` (§C); the approved design document is `.moai/reports/moai-handoff-resume-plan-20261010.md` (primary checkout; the SSOT for the design). Line numbers below are coordinates at the tree above and drift with edits; the names are the contract.

Tier M (rationale in `spec.md` §A). Milestones ordered by decision-reversibility: the new behavior flows (ask matrix, capture, launcher) first, config keys before their consumers, mechanical documentation and yaml-value steps last.

Files this SPEC changes:

| File | Role |
|---|---|
| `internal/config/types.go` | `HandoffConfig`: add `CaptureOnClear`, `LauncherPrompt`; `Mode` doc gains `ask` |
| `internal/config/defaults.go` | `DefaultHandoffCaptureOnClear = false`, `DefaultHandoffLauncherPrompt = true`; capture-budget constants (CC 10s / Codex 3s); mode-validation helper |
| `internal/template/templates/.moai/config/sections/handoff.yaml` | generic defaults (manual/false/false/true) + ask/auto documentation |
| `.moai/config/sections/handoff.yaml` | this project's values: `mode: ask`, `capture_on_clear: true` |
| `internal/settings/schema_sections.go` | `handoffModeValues` gains `ask`; two new bool FieldDefs |
| `internal/settings/testdata/sections/handoff.yaml` | fixture kept consistent with the enum (update only if its assertions enumerate mode values) |
| `internal/hook/handoff_inject.go` | `ask` branch: notice-only additionalContext on startup+clear, silent resume+compact, ask-mode stale expiry, manual/auto cells byte-preserved |
| `internal/hook/handoff_inject_render.go` | new ask-notice string set, 4 locales (ko/en/ja/zh, en fallback), never containing the body |
| `internal/hook/session_end.go` | reads `input.Reason`; on `clear` calls the capture |
| `internal/hook/handoff/capture.go` (new) + `capture_test.go` | capture payload build + save with provenance kind; budget enforcement |
| `internal/homestate/factory.go` | `kind` provenance column migration (the `ALTER TABLE` block at `:396` is the precedent) |
| `internal/cli/handoff.go` | new `resume` verb (claim → print body → consumed; localized no-pending error) |
| `internal/cli/launcher_prompt.go` (new) + test | launcher question: gate seam, renderer, one-line parser, locale catalogue, decline event |
| `internal/cli/cc.go`, `internal/cli/glm.go` | call the launcher question before spawn; pass the body as the positional initial-prompt argument on resume |
| `internal/cli/codex_launcher.go` | same question; body through the existing positional PROMPT channel |
| `internal/cli/init_update_notice.go` (`isInteractiveStdin`, `:69`), `internal/cli/glamour_style.go` (`writerIsTerminal`, `:137`) | read only — reused seams, not duplicated |
| `.claude/rules/moai/workflow/session-handoff.md` (+ `-format.md`, `-examples.md` if they mention the store) and mirrors under `internal/template/templates/.claude/rules/moai/workflow/` | pending.json → `resume_handoffs` store reference; document the four keys + ask semantics |
| `.claude/output-styles/moai/moai.md` + its template mirror | auto-flow section amended with the ask-mode notice behavior |

Existing suites that must stay green (enumerated by `go test -list 'Handoff'` at the base tree): `internal/hook` — `TestHandoffConfig_Fallbacks`, `TestHandoffStale_ZeroTime`, `TestRenderHandoffContext_AllDirectives`, `TestHandoffInjectRender`, and the inject-suite tests of `handoff_inject_test.go`; `internal/hook/handoff` — all of `pending_test.go` and `persist_test.go`; `internal/cli` — `TestHandoffShow_*` (6), `TestHandoffSave_*` (4), `TestHandoffClear`, `TestHandoffCmdRegistered`.

## B. Decisions, ordered by likelihood of change (reversibility order)

1. **The ask branch's 36-cell behavior table** (`spec.md` §D, `acceptance.md` AC-HRA-001) — new flow, cheapest to change before M2 lands.
2. **The launcher question grammar and locale catalogue** — new UX, cheap before M5.
3. **The capture payload fact list and the `kind` provenance column** — a schema addition; reversible by one migration line before M3.
4. **Config key shapes and the validation point** (fail-fast in the config loader; hook coercion stays as defense-in-depth) — Q-gated by nothing; the key set is dispatch-fixed.
5. Mechanical last: yaml values, the docs/mirror drift correction, the output-style amendment.

## C. Design

### C.1 Config layer (M1)

`HandoffConfig` gains `CaptureOnClear bool \`yaml:"capture_on_clear"\`` and `LauncherPrompt bool \`yaml:"launcher_prompt"\``. `NewDefaultHandoffConfig` returns `{manual, false, false, true}` (compiled defaults = template defaults = today's behavior when keys are absent). Mode validation lives in the config loader's validation step: a `mode` outside `{manual, auto, ask}` produces a load error naming the key and value (REQ-HRA-013); `handoffConfig()` in the hook keeps its coerce-to-manual fallback untouched. `schema_sections.go` adds `"ask"` to `handoffModeValues` and two `TypeBool` seam fields (`capture_on_clear`, `launcher_prompt`) beside the existing `mode` select and `guide` bool.

### C.2 SessionStart ask branch (M2)

The `handoffInjectHandler` branch table extends from 2 mode columns to 3. For `ask` + live pending: `startup` and `clear` return an additionalContext ask-notice (rendered by a new `askLocaleStrings` set following the `handoffStrings` pattern — header, one-line directive to ask the user about resume on the first turn, two-choice guidance for question-tool-less harnesses; never the body); `resume` and `compact` return an empty `HookOutput`. For `ask` + expired: run the same `ExpirePending` stale cleanup `auto` runs, emit nothing. For `ask` + none: silent no-op. The `manual` and `auto` columns are untouched line-for-line — the matrix test asserts their current outputs so a regression is a test failure, not a review catch.

### C.3 Session-end capture (M3)

New `internal/hook/handoff/capture.go`: `CaptureOnClear(projectDir string, sessionID, branch, head, specID, lastPrompt string) error` — reads the existing `ReadPending`; on a live row it returns without writing (REQ-HRA-011); otherwise it saves a record whose provenance kind is `auto-capture` (new nullable `kind` TEXT column, `ALTER TABLE` migration beside `factory.go:396`; empty string = model-authored, the entire existing surface unchanged) with `saved_by_session` = the ending session and the light 5-fact body (each fact omitted-to-empty when unreadable). `session_end.go` reads `input.Reason` and calls the capture only on `reason == "clear"` when `cfg.Handoff.CaptureOnClear`. Budget: `DefaultHandoffCaptureBudgetCC = 10s`, `DefaultHandoffCaptureBudgetCodex = 3s` constants in `defaults.go`; the effective budget comes from an env override the Codex hook configuration sets (`MOAI_HANDOFF_CAPTURE_BUDGET`), defaulting to the CC value — the requirement is the ceiling per harness, and the capture path contains no network or model call, so the fixture measurement stays far under it. Every error is fail-open (log + return; session end never blocked). The capture never embeds a goal or directives.

### C.4 `moai handoff resume` (M4)

Mirrors the injector's claim sequence minus injection: `ReadPending` → stale check → `ClaimPending` (transactional CAS; a concurrent claimer wins) → print `claimed.Body` verbatim to stdout once → `FinishClaim(..., consumed)`; no live/expired pending → localized one-line error to stderr, exit 1, no body printed. Locale set follows `handoffShowLocaleStrings`.

### C.5 Launcher question (M5)

New `internal/cli/launcher_prompt.go`: `launcherPromptInteractive` seam (assignable, `isInteractiveStdin() && writerIsTerminal(os.Stderr)` — the `runSelectionInteractive` precedent), a pure row/choice model (`launcherChoice` ∈ resume/fresh/view), a renderer, and a one-line parser. Flow per launcher (`cc.go`, `glm.go`, `codex_launcher.go`), before spawning the child: resolve the project's pending row (read-only); when live and the seam passes, render the question in the project's `conversation_language` (unreadable → en) and read **one line byte-wise up to the newline** — the launcher then hands the terminal to the backend, so a buffered reader that over-reads would swallow the child's first keystrokes (the SPEC-LANE-JOIN-RUN-SELECT-001 lesson, applied before it can repeat). resume → claim, pass the body as the child's positional initial-prompt argument (Codex: its existing positional PROMPT channel); claim lost → start without the body, no error. fresh/empty/EOF → leave pending, append one `handoff_events` decline row, start. view → print the body verbatim, claim nothing, re-present. Non-TTY → no question, unchanged start.

### C.6 Documentation drift (M6)

`session-handoff.md:25` (and any sibling mention) re-pointed to the `resume_handoffs` store; the four keys and the manual/auto/ask semantics documented in the same edit; template mirrors updated with the identical generic text (no project values — the project's `mode: ask` / `capture_on_clear: true` live only in `.moai/config/sections/handoff.yaml`). The output-style auto-flow section gains the ask-mode notice sentence. After `make build`, `git status --porcelain` must list no modified tracked file (the build's regeneration, including any catalog/embed sync, leaves the committed tree consistent — the interpretation of the card's "make build 뒤 diff 0"; the live project yaml intentionally differs from the generic template mirror by design).

## D. Milestones (priority order; no time estimates)

Each milestone is RED-first: the named tests are written and observed failing for the stated reason before the code, then turned green.

### M1 — Config layer (Priority High; REQ-HRA-012, REQ-HRA-013)

- Files: `types.go`, `defaults.go`, both live/template `handoff.yaml`, `schema_sections.go`, testdata fixture.
- Tests: `TestHandoffModeValidationRejectsUnknown`, `TestHandoffModeValidationAcceptsAsk`, `TestHandoffConfigTemplateDefaults` (M1, config); `TestHandoffSchemaEnumCarriesAsk` (M1, settings).
- Done when: those pass; `GOOS=windows GOARCH=amd64 go build ./...` exits 0.

### M2 — SessionStart ask branch (Priority High; REQ-HRA-005, REQ-HRA-006, REQ-HRA-008, REQ-HRA-009)

- Files: `handoff_inject.go`, `handoff_inject_render.go` + tests.
- Tests: `TestHandoffInjectBranchMatrix` (the 36-cell table: mode × source × pending-state, asserting additionalContext, stderr hint, and record status per cell), `TestHandoffAskNoticeContentAndLocales` (never contains the body; 4 locales), `TestHandoffAskNeverConsumes` (record status unchanged after every ask cell).
- Done when: those pass and every existing `internal/hook` handoff guard stays green.

### M3 — /clear capture (Priority High; REQ-HRA-010, REQ-HRA-011)

- Files: `handoff/capture.go` (new), `session_end.go`, `homestate/factory.go` migration, `defaults.go` budgets.
- Tests: `TestSessionEndClearCaptureCreatesSingleAutoCaptureRow` (payload facts + kind), `TestSessionEndCaptureSkipsWhenLivePendingExists`, `TestSessionEndNonClearReasonCreatesNoRecord` (reason axis table), `TestCaptureWithinBudgetAndFailOpen` (error injection → no record, no panic).
- Done when: those pass; the `pending_test.go` / `persist_test.go` suites stay green.

### M4 — `moai handoff resume` (Priority High; REQ-HRA-007)

- Files: `internal/cli/handoff.go` + test.
- Tests: `TestHandoffResumeClaimsPrintsConsumes`, `TestHandoffResumeNoPendingExitsNonZero`, `TestHandoffResumeConcurrentClaimLoses`.
- Done when: those pass; `TestHandoffCmdRegistered` extended for the new verb.

### M5 — Launcher question (Priority High; REQ-HRA-001..004)

- Files: new `internal/cli/launcher_prompt.go` + test; edits in `cc.go`, `glm.go`, `codex_launcher.go`.
- Tests: `TestLauncherPromptGateOnTTYSeam`, `TestLauncherPromptRendererAndParser` (in-memory streams; empty/EOF → fresh), `TestLauncherResumePassesBodyAsFirstPrompt` (argv assertion for all three launchers), `TestLauncherFreshStartDeclineEvent`, `TestLauncherViewPrintsWithoutClaim`, `TestLauncherNonTTYSkipsSilently`.
- Done when: those pass; the launcher inventory grep (`grep -n "launcherPrompt" internal/cli/{cc,glm,codex_launcher}.go`) lists all three call sites.

### M6 — Documentation drift + build consistency (Priority Medium; REQ-HRA-014)

- Files: rules docs + mirrors, output style + mirror, project yaml values (if not landed in M1).
- Verification: the grep sweep of AC-HRA-015 (zero `pending.json` store references; zero project-specific values under `internal/template/templates/`); `make build` then `git status --porcelain` clean.

## E. Self-verification (manager-develop reports, Tier M template §E)

- Scoped runs only: `go test -cover` for `internal/config`, `internal/hook`, `internal/hook/handoff`, `internal/cli`, `internal/homestate`, limited to the new test names plus the named existing guards. Do not run the repo-wide suite or the full `internal/cli` package without the single-slot lease (`moai slot acquire --resource go-test-heavy`); CI runs the full suite.
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `golangci-lint run --new-from-rev=<base>` over the touched packages: 0 new issues.
- Locale-string hygiene: no `\u` escape in the new locale catalogues (native UTF-8); 4 locales × new strings present.
- Template generics: `grep -rn "ask" internal/template/templates/.moai/config/sections/handoff.yaml` shows the enum in comments only, `mode:` value stays `manual`; no `t1672`/project-path strings under `internal/template/templates/`.
- `make build` → `git status --porcelain` empty (tracked files).
- Manual measurements recorded verbatim in `progress.md` §E.2: launcher TTY question once (AC-HRA-008), resume-as-first-prompt once (AC-HRA-009), Codex SessionStart `source` value once (AC-HRA-014), Codex launcher question once (AC-HRA-014).

## F. Risks

- **The ask-notice can be missed by the session** (it is context, not a gate). Accepted: L1 is the guaranteed path; L2 is best-effort, and on Codex explicitly so (REQ-HRA-009).
- **Positional-prompt differences between the claude/glm/codex CLIs** (quoting, length limits on a long body). The argv-assembly test pins the mechanism; a body that a CLI refuses at spawn is a visible launcher error, not a silent loss — the record stays claimed, so a retry re-prints from `moai handoff show`. Residual accepted.
- **Codex SessionStart `source` vocabulary is unmeasured** (§G): the matrix asserts the Claude Code axis; the Codex measurement may require mapping, recorded in the report. The manual/auto cells are harness-agnostic and unaffected.
- **The one-pending unique index vs the capture**: a pending row from another session blocks the capture by design (REQ-HRA-011) — the capture is skipped, not queued.
- **Budget measurement on a slow disk** inflates the fixture timing; the budget is a code-path ceiling (no I/O beyond one SQLite write), and the failure mode is fail-open.
- **Three `handoff.yaml` copies** (template mirror, project, settings testdata) can drift; M1 edits all three in one commit and the schema-enum test pins the value set.

## G. Anti-patterns to avoid

- Consuming the record under `ask`, or on any SessionStart source other than the `auto`+`clear` cell and the explicit `resume` verb.
- Injecting the body under `ask` — the notice never carries the body (REQ-HRA-005).
- A capture path that calls the model or the network (budget and determinism both forbid it).
- Blocking session start or session end on any prompt/capture error (fail-open everywhere).
- A buffered stdin read in the launcher question that over-reads past the newline (the child's first keystrokes are lost).
- Project-specific values in any template-mirror file; a `\u` escape in a locale catalogue.
- Editing the model-authored 5-trigger doctrine or the paste-ready format while "fixing" the docs.
- Changing `manual`/`auto` cells while extending the branch table (the matrix test exists to catch exactly this).

## H. Cross-references

- SPEC-HANDOFF-AUTORESUME-001 — owner of the inject+consume cell and the pending store semantics this SPEC extends (never alters).
- SPEC-HANDOFF-NEUTRAL-001 — the harness-neutral `show` verb and its locale conventions, reused by `resume`.
- SPEC-INFINITE-GOAL-001 — embedded-goal re-arm semantics; the capture embeds no goal.
- SPEC-LANE-JOIN-RUN-SELECT-001 — the launcher seam precedents (`isInteractiveStdin`, `writerIsTerminal`, assignable gate vars) and the byte-wise stdin lesson.
- `.claude/rules/moai/workflow/session-handoff.md` — the doctrine this SPEC's M6 re-points to the real store.
