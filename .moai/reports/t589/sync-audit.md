# Sync-Audit — SPEC-UPDATE-ADD-CODEX-001 (card t589)

- Auditor: sync-auditor (independent, fresh-judgment — re-executed, not trusted)
- Tree: `.claude/worktrees/t589` @ `55b5d5b8e` (branch `WT-add-codex-verb`, base `5caddeb2d` = origin/develop)
- Date: 2026-09-09
- Audit binary: `/tmp/t589-audit/moai` — built in this audit run via `go build -o /tmp/t589-audit/moai ./cmd/moai` from this tree (exit 0, `moai-adk v3.1.3`)

## Overall Verdict: PASS

14/14 acceptance criteria independently reproduced. F1 mutant closure confirmed at both layers. Deviation 1 (clean-reinstall second wiring seat) confirmed by code reading AND dynamic reproduction through the genuine v2 clean-reinstall branch. No blocking findings; 2 optional findings recorded.

---

## Evaluation Report

SPEC: SPEC-UPDATE-ADD-CODEX-001
Overall Verdict: PASS (flat weighted mode, default profile, Tier M threshold 0.80)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 100 | PASS | 14/14 AC re-executed this run — all reproduced (matrix below); F1 refusal test + wire-level refusal test both green: `--- PASS: TestUpdateAddCodex_ValidationRefusalFailsLoud` / `--- PASS: TestWireValidationRefusalWritesNothing`; AC-014: `go test ./internal/cli/... ./internal/codexwiring/... ./internal/config/...` → zero FAIL lines, exit 0 |
| Security (25%) | 100 | PASS | Write surface limited to 3 relative constants (`HooksRelPath = ".codex/hooks.json"`, `ConfigRelPath = ".codex/config.toml"`, `SidecarPath = ".moai/state/codex-wiring.json"` — `internal/codexwiring/codexwiring.go:29,31,40`), always under `filepath.Join(projectRoot, …)`; dry-run preview takes only an `io.Writer` and measured writes nothing; `.mcp.json` sha256 identical across the verb (measured); whitelist pre-validation refuses before bytes |
| Craft (20%) | 90 | PASS | `go vet` exit 0; `golangci-lint run ./internal/cli/... ./internal/codexwiring/... ./internal/config/...` → `0 issues.`; coverage: `ok github.com/modu-ai/moai-adk/internal/codexwiring 0.716s coverage: 89.6% of statements` (≥85); 13 new tests + 5 sub-cases all PASS this run. Deduction: 3 source-based guards cover only the FIRST `addCodexWiringAt` call site (F-A1) |
| Consistency (15%) | 95 | PASS | Flag registration/comment/godoc style matches sibling SPEC-UPDATE-VERSION-FLAG-001 conventions; help text, CHANGELOG entry, and docs-site 4-locale pages carry identical "sanctioned additive path" language; measured: docs-site H2=7/H3=21 in all four locales. Deduction: dry-run preview prints plain `fmt.Fprintln` lines beside the TUI-styled existing dry-run output (F-A2 note, cosmetic) |

**Harmonic mean: 96.1** — above the Tier M threshold (0.80). Must-pass firewall: Functionality + Security both 100 — firewall holds.

### AC Matrix — independent re-execution (this run, this tree, HEAD 55b5d5b8e)

| AC | My re-run (verbatim key output) | Verdict |
|----|--------------------------------|---------|
| AC-UAC-001 | `./moai update --help \| grep -c -- --add-codex` → `1` | PASS |
| AC-UAC-002 | scratch s2: `hooks: exists / config: exists / sidecar: exists` · `mcp_servers.moai count: 1` · FirstTrustGuidance line printed | PASS |
| AC-UAC-003 | scratch s3 flag-absent: `flag-absent-exit=0` · `.codex/hooks.json: absent` · `.codex/config.toml: absent` · `.moai/state/codex-wiring.json: absent` (practical contract; literal `test -e .codex` premise is known accepted debt — init deploys `.codex/agents`) | PASS |
| AC-UAC-004 | `.mcp.json` before/after: `AC-004 .mcp.json: SHA-IDENTICAL` | PASS |
| AC-UAC-005 | re-run: `sidecar: SHA-IDENTICAL` · `guidance-count-in-run2: 0` | PASS |
| AC-UAC-006 | `user-key: 1 · mcp: 1 · statusline: 1` | PASS |
| AC-UAC-007 | dry-run: preview lines printed, `hooks/config/sidecar: absent` after | PASS |
| AC-UAC-008 | `$B update --check --add-codex` → `exit=1` (direct, no pipe) + `--Check and --add-codex are mutually exclusive (--check is informational; --add-codex mutates project wiring).` | PASS |
| AC-UAC-009 | titles: `5`; anchors (a)2 (b)3 (c)1 (d)7 (e)3 — all ≥1 | PASS |
| AC-UAC-010 | `wc -c` template AGENTS.md = `18582` ≤ 24576 · `--- PASS: TestCodexContractByteCeiling` | PASS |
| AC-UAC-011 | CLAUDE.md `^## [0-9]` = `18` · `@AGENTS.md` = `1` · quality pointer in AGENTS.md = `1` | PASS |
| AC-UAC-012 | marker in template AGENTS.md = `0` · `Path: "CLAUDE.md"` in `internal/harness/curator/dispatch.go` = `1` (all other curator files 0) | PASS |
| AC-UAC-013 | `init --force --agent both --non-interactive` → `exit=0` · line 2: `note: this project is already initialized — the sanctioned additive path ... is \`moai update --add-codex\` ... Proceeding with the requested reinit.` · line 3: `· Initializing MoAI project...` (guidance precedes reinit) | PASS |
| AC-UAC-014 | 3 affected packages ok (background run exit 0) · `grep -c '^func Test' internal/cli/init_agent_flag_test.go` = `8` (non-decreasing, EV-20 recorded 8) | PASS |

### F1 mutant closure — verified both layers

1. `TestUpdateAddCodex_ValidationRefusalFailsLoud` exists (`internal/cli/update_add_codex_test.go:157`) and PASSES: asserts non-nil error, `errors.Is(err, codexwiring.ErrValidationRefused)`, diagnostic naming the violating key (`version`), hooks.json bytes unchanged, and config.toml NOT written despite the refusal.
2. `TestWireValidationRefusalWritesNothing` still PASSES (`--- PASS` in this run's codexwiring package execution).
3. Code reading of `addCodexWiringAt` (`internal/cli/update_codex_wiring.go`): `errors.Is(err, ErrValidationRefused) → return err`; all other errors warn-and-continue (`warning: Codex wiring failed: %v`) and return nil — exactly the documented posture (refusal hard, IO best-effort). The swallowing-sibling pattern (`refreshCodexWiringBestEffortAt`, which discards every error silently) was NOT copied. Two source-based guards additionally pin reachability and placement for the first call site; `TestUpdateAddCodex_CallSitsInsideFlagGuard` pins the flag guard (M-3 mutant), `TestUpdateAddCodex_CallSitsBeforeSyncSkippedReturn` pins the up-to-date-branch seat.

### Deviation 1 — second wiring seat (clean-reinstall success block): verified

- **Code shape** (diff + Read of `internal/cli/update.go:472-482`): identical flag-guard + refusal-propagation form as the primary seat — `if getBoolFlag(cmd, "add-codex") { if err := addCodexWiringAt(cwd, out, cmd.ErrOrStderr()); err != nil { return err } }` inside the `fingerprint.IsV2 && isMoAIProject(cwd)` clean-reinstall success block, immediately before its `return nil`.
- **Dynamic reproduction**: scratch s7 → `init --agent claude` → `system.yaml` `moai.version` set to `"v2.9.0"` (Signal 1 positive per `internal/cli/v2_detection.go` — normalized major 2 with leading "v") → `$B update --add-codex` → observed verbatim:
  ```
  ·  v2 detected  running clean reinstall · signals: version=true agency=false deprecated=true
   Clean reinstall complete (4 files preserved, 1 deprecated removed)
  ```
  followed by `.codex/hooks.json: exists · .codex/config.toml: exists · .moai/state/codex-wiring.json: exists · mcp count: 1`. The executor's measured defect (v2-era project silently never wired without this second seat) is real and the fix works end-to-end. Two intermediate attempts that did NOT trigger the branch (root `moai.version` file — the signal lives in `system.yaml`; `.agency/` alone on a v3 project — v3-version negative-override) confirmed the branch discriminator is specific and the primary seat correctly serves the up-to-date path (both covered in scenarios s2/s6).

### Sync commit integrity — 55b5d5b8e

- File set: exactly the 7 reported files (`progress.md`, `spec.md`, `CHANGELOG.md`, docs-site `update.md` × ko/en/ja/zh). No template or Go file touched.
- `spec.md` diff: one hunk, `status: in-progress → completed` only — body untouched. Verified byte-level via `git show`.
- `progress.md` §E.4: sync signal + `pending-backfill-sync` placeholder (D3 self-reference convention; backfill owed in a follow-up commit by the lead's window).
- docs-site 4-locale: re-measured per file — `add-codex` count 4, `^## ` 7, `^### ` 21, total headings 28 in ALL four locales (identical structure). Flag-table row + dedicated 4-bullet section, matching the reporter's measurement.
- CHANGELOG: single entry prepended to `### Added` at the top (above the t572 entry), SPEC-linked, consistent with the file's entry conventions.
- Progress.md §E.2 deviation records (1: AC-UAC-003 literal premise; 2: clean-reinstall second seat) present and accurate — they match my independent measurements.

### Findings (structured defect-list)

| ID | Severity | Blocking | Location | Description | Required fix |
|----|----------|----------|----------|-------------|--------------|
| F-A1 | low | optional | `internal/cli/update_add_codex_test.go:195-252` | The three source-based guards (`RunUpdatePropagatesRefusal`, `CallSitsInsideFlagGuard`, `CallSitsBeforeSyncSkippedReturn`) use `strings.Index` and therefore inspect only the FIRST `addCodexWiringAt` occurrence in update.go — the clean-reinstall second seat has no source-level regression guard. Current behavior verified correct by my dynamic reproduction, but a regression that drops or unguards the second seat would pass this family. Confidence: high (index-based scanning is the whole mechanism). | Add second-occurrence coverage (e.g. scan after the first index, or use `strings.LastIndex`) asserting the same `getBoolFlag(cmd, "add-codex")` + `return err` shape for the clean-reinstall seat. Discretionary — the dynamic path is exercised by nothing in CI, so the guard has real value. |
| F-A2 | info | optional | `internal/cli/update_codex_wiring.go` `emitAddCodexDryRunPreview` | The preview prints plain text lines while the surrounding dry-run output uses the TUI pill/check-line style. Cosmetic inconsistency; content is complete and correct. Confidence: high. | Optional: render via the same TUI helpers the neighboring dry-run plan uses. |

No blocking findings. F3/F4/F6 plan-audit optionals, §D.3:137 "간접" label, AC-UAC-003 literal premise, and EV-11/12 exit-code pipe slippage are accepted debt per the dispatch and were not re-raised.

### Recommendations

- F-A1 second-seat source guard: cheapest possible fix (one additional index-based assertion in the existing test file) and it closes the only coverage asymmetry this audit found.
- Backfill `sync_commit_sha` in the owed follow-up commit (placeholder `pending-backfill-sync` recorded in 55b5d5b8e).

---

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim.** SPEC-UPDATE-ADD-CODEX-001 satisfies all 14 acceptance criteria on the audited tree; the F1 refusal contract holds at both wrapper and wiring layers; the clean-reinstall second wiring seat exists, has the mandated shape, and serves wiring on the genuine v2 path; the sync commit is scoped exactly as reported; the change is fit to PASS.

**Evidence.** Every claim above cites a command executed in this audit run and its verbatim output (AC matrix rows, dimension Evidence cells, Deviation-1 console block). Build: `go build -o /tmp/t589-audit/moai ./cmd/moai` → exit 0. Scratch scenarios: s2 (add-codex happy path), s3 (flag-absent + dry-run), s4 (user config + check rejection), s5 (init --force guidance), s6/s7 (branch-discriminator isolation + genuine v2 clean-reinstall). All scratch projects under `/tmp/t589-audit/` with `MOAI_SKIP_BINARY_UPDATE=1` (the codebase's own isolation guard, as in §D.5 EV-14).

**Baseline-attribution.** This run, this tree (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t589` @ HEAD `55b5d5b8e`), audit binary built from this tree. §D.5 EV-13~20 values were NOT carried over — each was re-derived by my own invocation; where my measurement instruments differed from the ledger's (e.g. anchor counts (b) 3 vs 2, (d) 7 vs 5 — different but stricter grep patterns; all minimums exceeded), the difference is instrument-shape, not verdict-shape.

**Gaps.** Explicitly NOT observed in this audit:
1. CI verdict on origin/develop — the branch is unpushed by lane discipline; remote CI judgment is the lead's batch-push surface, not observable from here.
2. `make build` / `make embed-check` (binary-embed axis, §D.4 #2) — not run; the embed axis (committed templates vs built binary bytes) was not measured in this audit. The audited Go binary embeds the same committed templates it was built from, so template-vs-tree consistency is measured; template-vs-distributed-binary is not.
3. Interactive-flow interaction with the verb (update prompts that appear when a newer binary exists, wizard-driven paths) — all scenarios ran non-interactively; prompt-path behavior of `--add-codex` is unmeasured.
4. `go test -race` on the touched packages — not run (change introduces no goroutines/channels; race discipline follows the sibling refresh path already shipped).
5. internal/cli package-wide coverage — the package suite passed (background run, exit 0, zero FAIL lines) but per-package coverage percentage for internal/cli was not measured; only codexwiring (89.6%) and config (80.7%) were.

**Residual-risk.** What could still be wrong despite the above:
1. `internal/cli/doctor_codex.go` `initCodexAdvice` still prints `init --agent codex` as the declared path — the docs (doctor.md 4 locales) faithfully document the Go string, so docs match code, but the user-facing advice now names the destructive path while the sanctioned one exists. Known follow-up (out of this audit's scope per dispatch); risk: users following doctor advice still land on the destructive reinit.
2. Root `AGENTS.md` (15,415 B) remains the pre-M2 version while the template (18,582 B) carries the five sections — the recorded §7 gap (no byte-twin enforcement guard). Harmless until the next `moai update` absorbs the template, but until then a codex session in THIS repo reads the thinner contract. Byte ceiling is measured on both (guard green), so no ceiling risk.
3. F-A1's unguarded second seat: any future refactor of the clean-reinstall block could silently drop or unguard the wiring call without a test failing — the dynamic path is not exercised in CI.
4. `internal/config` coverage 80.7% is below the 85% package target — pre-existing state untouched by this SPEC (no Go changes there), noted here so it is not mistaken for a regression.

---

Verdict recap: **PASS** — harmonic 96.1 against Tier M 0.80, must-pass firewall holds, zero blocking findings.
