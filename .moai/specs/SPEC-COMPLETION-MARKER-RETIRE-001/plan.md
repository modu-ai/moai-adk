# Implementation Plan — SPEC-COMPLETION-MARKER-RETIRE-001

## §A. Context

This is a retirement SPEC: it removes a coupled dead/dormant unit (completion markers + persistent-mode subsystem) rather than adding behavior. The dominant risk is not functional regression of the retired feature (it is dead/dormant in production) but **collateral breakage** in the surrounding live code (`stop.go` telemetry/reflection paths, the config loader symmetry guards, the embedded-template parity checks, and the 4-locale docs).

Ground-truth was confirmed by direct file reads (line numbers cited below are current as of authoring). One blast-radius item beyond the original brief was discovered during plan-phase verification: `internal/hook/compact.go` is a **second live consumer** of `persistent-mode.json` (via `readPersistentMode`), so the PreCompact handler must also be edited — captured as REQ-CMR-003 / M1.

## §B. Tier Classification

**Tier: M (standard harness).** Justification:

- Multi-file Go change across 3 packages (`internal/hook`, `internal/hook/lifecycle`, `internal/config`) plus `pkg/models`.
- Test retirement across ≥5 test files (Class A breakers).
- Template-First obligation (`make build` + embedded regeneration) with template↔mirror parity.
- 4-locale docs-site sync (HARD per CLAUDE.local.md §15/§17) spanning ≥6 EN files and their ko/ja/zh siblings.
- 2 SPEC traceability supersessions.

This exceeds Tier S (single-file, no docs, no template) but does not reach Tier L (no new public API, no architectural decision, no PR-mandatory cross-cutting design). The output-style design choice (below) is a documented decision, not a new subsystem — consistent with Tier M.

## §C. Design Decision (Central) — Prompt-Layer Marker Contract + `/moai loop` Exit Signal

The completion markers serve **three** roles, not two. Plan-phase re-verification (auditor D8) corrected the earlier "no current consumer" claim:

1. **Runtime role (DEAD/dormant)** — marker detection in `stop.go` + `compact.go`. Unambiguously removed; nothing consumes it in production.
2. **Prompt-layer terminal-token role (UX)** — the orchestrator emits `<moai>DONE</moai>` / `<moai>COMPLETE</moai>` as the terminal token of the Completion Report / Session Handoff. The actual `moai.md` rule (line 48) carves the marker out as an **EXCEPTION** to "No XML tags in user-facing output" (a second exception clause repeats at line 729) — it is NOT the "sole permitted XML" (D9 correction).
3. **`/moai loop` (Ralph Engine) LIVE loop-exit signal** — `loop.md:64-66` Step 1 ("Completion Check") reads the marker from the previous iteration's response and exits the loop on detection (`:100`, `:157` reinforce). **This is a live prompt-layer consumer** — dropping the marker without a replacement removes `/moai loop`'s explicit success-exit path, leaving only its other 4 exit conditions (zero-errors+tests+coverage / max-iterations / memory-pressure / user-interruption). This is the auditor's D8 finding and the reason "no current consumer" was wrong.

Two options for roles 2+3:

- **Option (a) — Drop the marker; replace the loop-exit signal with a natural-language completion sentence (RECOMMENDED).** Completion Report / Session Handoff signal completion via the banner / prose alone (role 2). For role 3, `loop.md` Step 1 Completion Check and the §Completion Conditions list are rewritten to key on an **explicit natural-language completion sentence** (e.g., "All loop completion conditions satisfied; exiting loop.") that the Ralph Engine loop evaluator detects, replacing the `<moai>DONE</moai>` / `<moai>COMPLETE</moai>` check. Both `moai.md` XML-exception clauses (lines 48, 729) are removed so the "No XML in user-facing output" rule has no carve-out.
  - Trade-off: the loop-exit signal becomes a prose sentence rather than a rigid greppable token. Mitigation: the loop evaluator already reads the previous iteration's full response (not a regex on a fixed token); a stable natural-language sentence is a reliable trigger and is consistent with the `/goal`-style "model judges the transcript" pattern. The other 4 exit conditions remain unchanged as a backstop.
- **Option (b) — Replace with a non-marker terminal token.** Preserves a rigid greppable end token (under a non-XML spelling) that doubles as the loop-exit signal.
  - Trade-off: re-introduces a token to maintain across output-style + docs + 4 locales; contradicts the "no XML in user-facing output" tightening; a future retirement would have to remove it again. Its only benefit (rigid greppability) is marginal because the loop evaluator reads the full response anyway.

**Recommendation: Option (a).** Simplest end-state; removes the misleading "configurable-but-unwired" surface fully; aligns `moai.md` with the broader "No XML in user-facing output" rule; and gives `/moai loop` a concrete, equally-reliable natural-language exit signal (REQ-CMR-015). This is surfaced for orchestrator/user confirmation at GATE-2 (§G Q1) — it is the single open design decision and the only item that materially changes user-visible behavior + loop-control semantics. If the user selects Option (b), the replacement token satisfies REQ-CMR-015 directly.

## §D. Milestones

Milestones are priority-ordered work steps (no time estimates). M1→M3 are the runtime core; M4→M5 are the surface (template + docs); M6 is traceability + close.

### M1 — Go runtime removal (Priority High)

- `internal/hook/stop.go`: remove `defaultCompletionMarkers` (16-19), the marker-detection loop (109-121), the `hasCompletionMarker` method (158-171), and the persistent-mode branches (74-98). Preserve `stop_hook_active` short-circuit, telemetry pruning (100-107, 132-135), reflective-learning (123-130), and the empty-`HookOutput{}` allow return. (REQ-CMR-001, REQ-CMR-004, REQ-CMR-005)
- `internal/hook/lifecycle/persistent_mode.go`: retire the whole file. (REQ-CMR-002)
- `internal/hook/compact.go`: remove `readPersistentMode` (142-159) and the "Execution Mode" section block (97-107); preserve `readWorktrees` and the P1/P2 worktrees section. (REQ-CMR-003)
- `internal/cli/deps.go`: verify `NewStopHandler()` registration (159) still compiles after the constructor reduction; no marker injection to remove (it never injected). (REQ-CMR-005)
- Verify no residual coupling in `internal/cli/team_spawn.go` and `internal/hook/router.go`.

### M2 — Config struct + YAML removal (Priority High)

- `internal/config/types.go`: remove `CompletionConfig` (344-348), `MarkersConfig` (350-354), and the `Completion` field from `WorkflowConfig` (311). (REQ-CMR-006)
- `internal/config/defaults.go`: remove the `Completion:` block from `NewDefaultWorkflowConfig` (355-361). (REQ-CMR-006)
- `.moai/config/sections/workflow.yaml`: remove `completion:` block (7-11). (REQ-CMR-007)
- `internal/template/templates/.moai/config/sections/workflow.yaml`: remove `completion:` block (12-16). (REQ-CMR-007)
- `pkg/models/config.go`: remove `LogCompletionMarkers` from `LSPStateLogging` (187). (REQ-CMR-008)
- Verify struct↔YAML symmetry: the symmetry test (`audit_struct_yaml_symmetry_test.go`) does NOT enumerate `WorkflowConfig`, so removing the struct does not break a symmetry case directly; the obligation is to keep YAML and struct removal in the same commit so the loader-completeness guard (`audit_loader_completeness_test.go`) and any top-level workflow-key check stay green. (REQ-CMR-007)

### M3 — Test retirement (Priority High)

The complete coupled test set (8 files) was re-derived by live grep and classified (REQ-CMR-009). The 4 false-positive files (`cli/coverage_improvement_test.go`, `cli/github_test.go`, `cli/oauth_token_preservation_test.go`, `hook/handoff/persist_test.go`) MUST NOT be touched.

**DELETE whole file:**
- `internal/hook/stop_completion_test.go` (156 LOC, 4 marker-detection tests). (REQ-CMR-009)
- `internal/hook/lifecycle/persistent_mode_test.go` (227 LOC, the retired subsystem's tests). (REQ-CMR-009)

**DELETE specific functions (PRESERVE the non-coupled siblings):**
- `internal/hook/stop_test.go`: remove 6 persistent-mode functions (`TestStopHandler_PersistentMode` 128, `_CompletionMarker` 161, `_Expired` 205, `_Inactive` 253, `_NoFile` 288, `_StopHookActive_OverridesPersistentMode` 316) + all `defaultCompletionMarkers` / `NewStopHandlerWithMarkers` usages. PRESERVE the 4 non-marker tests (`_EventType` 18, `_Handle` 28, `_Handle_StopHookActive` 76, `_Handle_StopHookNotActive` 101). (REQ-CMR-009)
- `internal/hook/compact_test.go`: remove `TestCompactHandler_Handle_ReadsPersistentMode` (179); PRESERVE worktrees tests. (REQ-CMR-009)
- `internal/hook/compact_coverage_test.go` **(auditor D2 — was omitted; WILL compile-break)**: remove 3 functions (`TestReadPersistentMode_MissingFile` 63, `_ValidJSON` 73, `_MalformedJSON` 90); PRESERVE `TestReadWorktrees_*` (13/24/46). (REQ-CMR-009)

**EDIT specific assertions (file survives):**
- `internal/config/defaults_test.go` **(auditor D3 — undercounted; 3 lines not 2)**: remove `Completion.DetectInOutput` (394), `Completion.Markers.Done` (437), `Completion.Markers.Complete` (438); reduce the "AC-WSE-007 36-assertion oracle" comment count by 3. (REQ-CMR-009)
- `internal/config/workflow_nested_test.go`: remove the `Completion.Markers.{Done,Complete}` assertion block (62-66) from the AC-WSE-003 oracle. (REQ-CMR-009)
- `internal/config/types_test.go` **(auditor D3)**: remove the 3 field-reachability rows (321 `Completion/DetectInOutput`, 322 `Completion/Markers/Complete`, 323 `Completion/Markers/Done`) that break once the struct fields drop. (REQ-CMR-009)

### M4 — Output-style + prompt-layer + `/moai loop` exit signal + template (Priority Medium)

Complete consumer set (8 live-tree files, re-derived by `grep -rln '<moai>\|moai>DONE\|moai>COMPLETE' .claude/ internal/template/templates/.claude/` — auditor D1/D6). The `agent-memory/plan-auditor/feedback_*.md` hit is a memory artifact, NOT a consumer (excluded).

**Live `.claude/` tree — 8 consumers:**
- `.claude/output-styles/moai/moai.md` — remove BOTH XML-exception clauses (line 48 + line 729) per REQ-CMR-011; remove §6/§7/§8 + line 181/198/602/640 handoff marker usages. (REQ-CMR-010, REQ-CMR-011)
- `.claude/output-styles/moai/einstein.md` — remove literal-list mention (line 317) + handoff usage (line 423). (REQ-CMR-010)
- `.claude/skills/moai/SKILL.md` — remove "add the appropriate completion marker" instruction (line 323). (REQ-CMR-010)
- `.claude/skills/moai/workflows/loop.md` **(role 3 — the LIVE loop-exit consumer)** — rewrite Step 1 Completion Check (64-66), the "Prompt user to add completion marker" line (100), and the §Completion Conditions list (157) to key on the natural-language completion sentence instead of `<moai>DONE</moai>`/`<moai>COMPLETE</moai>`. (REQ-CMR-010, **REQ-CMR-015**)
- `.claude/skills/moai/workflows/release.md` — remove terminal token (line 731). (REQ-CMR-010)
- `.claude/rules/moai/workflow/spec-workflow.md` — remove the marker glossary (lines 281-282 + the §Completion Markers heading). (REQ-CMR-010)
- `.claude/agents/local/release-update-specialist.md` — remove `Emit: <moai>DONE</moai>` (line 334). (REQ-CMR-010)

**Template mirror scope (REQ-CMR-012):** mirror the edits for the **template-managed subset** into `internal/template/templates/.claude/...`: `output-styles/moai/{moai,einstein}.md`, `skills/moai/SKILL.md`, `skills/moai/workflows/loop.md`, `rules/moai/workflow/spec-workflow.md` (5 mirrors — confirmed present by the D1 grep). **`release.md` + `release-update-specialist.md` are dev-only (CLAUDE.local.md §21 dev-only commands / §24 `.claude/agents/local/`) and have NO template mirror** — edit them in the live tree only; do NOT create template copies (would violate template-internal-isolation §25). Then run `make build`; verify template↔mirror byte-parity for SSOT-mirrored files. (REQ-CMR-012)

### M5 — docs-site 4-locale + README sync (Priority Medium)

Complete docs marker surface re-derived by `grep -rln '<moai>' docs-site/ README.md README.ko.md` (auditor D4 — adds `docs-site/README.md`; root `README.md`/`README.ko.md` have NO `<moai>` hits).

- **4-locale `content/`** (remove / rewrite in all 4 locales in parallel — no locale lags):
  - `utility-commands/moai.md` (en:200/252/263/278/279, ko:204/282/283, ja:198/276/277, zh:198/276/277)
  - `advanced/hooks-guide.md` (en:501, ko:537, ja:501, zh:501)
  - EN auxiliaries (these are EN-only — no ko/ja/zh siblings contain the marker per the D4 grep, so no missing-locale gap): `utility-commands/moai-loop.md` (en:80), `getting-started/introduction.md` (en:288), `core-concepts/what-is-moai-adk.md` (en:414)
- **`docs-site/README.md`** (auditor D4 addition — EN landing doc, 7 hits: 517/563/745/778/782/783/810). Rewrite the marker-based completion narrative to the natural-language completion signal.
- Enforce parity: AC-CMR-007 greps `docs-site/` (incl. README) + verifies per-locale residual count is identically zero. (REQ-CMR-013)

### M6 — SPEC traceability + Mx close (Priority Low)

- Record supersession notes in `SPEC-PERSIST-001` and `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001` frontmatter / body (partial supersession of the marker/persistent-mode surface; oracle inversion). (REQ-CMR-014)
- Standard Mx-phase close.

## §E. Technical Approach

- **Surgical removal, not refactor.** Each edit deletes marker/persistent-mode code and leaves adjacent live code byte-identical. No opportunistic cleanup of `stop.go` telemetry/reflection.
- **Same-commit YAML+struct symmetry.** Config struct and YAML edits land together (M2) to never leave the loader-completeness / symmetry guards in a transient-broken state.
- **Template-First discipline.** Any `.claude/...` edit is made in `internal/template/templates/.claude/...` and the local copy together, then `make build` regenerates `embedded.go` (do not edit `embedded.go` directly).
- **4-locale atomicity.** docs-site edits to all four locales are made in one logical unit so no PR/commit leaves a locale lagging.
- **Verification batching.** Run the read-only verification batch (go test, grep zero-residual, make build, locale-parity) as a single multi-Bash turn per the verification-batch pattern.

## §F. Risks

- **R1 — Loader-completeness / symmetry guard breakage.** Removing the `completion:` YAML key without removing the struct field (or vice versa) trips `CONFIG_STRUCT_YAML_MISMATCH` / `YAML_SECTION_NO_LOADER`. Mitigation: M2 removes both in the same commit; AC-CMR-005 asserts the config test package is green.
- **R2 — `defaults_test.go` assertion-count drift.** The AC-WSE-007 "36-assertion oracle" comment counts assertions; removing **3** (not 2) marker assertions requires reducing the count by 3 (auditor D3). `types_test.go` field-reachability rows (321-323) and `workflow_nested_test.go` (62-66) also break. Mitigation: M3 enumerates all three config-test edits; AC-CMR-006 verifies the suite green.
- **R3 — `compact_coverage_test.go` compile break (auditor D2).** `compact_coverage_test.go` calls `readPersistentMode` (3 `TestReadPersistentMode_*` functions) and was omitted from the original test list — it WILL fail to compile after `persistent_mode.go`/`compact.go` removal. Mitigation: M3 now lists it explicitly; AC-CMR-006 (`go test ./...` green) is the catch-all.
- **R3b — Hidden further consumer.** Plan-phase grep found two runtime consumers (`stop.go`, `compact.go`) + the live prompt-layer consumer (`loop.md`); a deeper indirection is unlikely but covered. Mitigation: AC-CMR-004 is an exhaustive zero-residual grep across `internal/`, `pkg/`, `cmd/` (production, non-test).
- **R4 — Embedded template drift.** Forgetting `make build` after template edits leaves `embedded.go` stale; CI golden-file check fails. Mitigation: AC-CMR-008 runs `make build` then asserts `git diff` on `embedded.go` is empty (already regenerated).
- **R5 — docs-site locale lag.** Editing EN but not ko/ja/zh violates the 4-locale parity rule. Mitigation: AC-CMR-007 greps all four locales for residual markers.
- **R6 — Output-style behavior change surprises users.** Dropping the marker changes the user-visible session-end format. Mitigation: §C design decision is surfaced at GATE-2 for explicit user confirmation before run-phase.

## §G. Open Questions (require orchestrator/user resolution before GATE-2)

- **Q1 (design — central):** Confirm Option (a) — drop the marker emit instruction entirely AND replace the `/moai loop` exit signal with a natural-language completion sentence (recommended, REQ-CMR-015) — vs Option (b) — replace with a non-marker terminal token that doubles as the loop-exit signal. This is the only material user-visible behavior change AND it alters `/moai loop` control semantics (the marker is a LIVE loop-exit consumer per auditor D8). Default proceeds with (a). Both options keep `/moai loop`'s other 4 exit conditions unchanged.

## §H. Cross-References

- spec.md §B (problem statement, 3-layer dead config + dormant persistent-mode), §C (REQ-CMR-*), §F (Exclusions).
- acceptance.md (AC-CMR-* verifiable commands).
- Precedent: `SPEC-V3R6-SEQ-THINKING-RETIRE-001`, `SPEC-LSPMCP-RETIRE-001` (retirement + 4-locale docs sync convention).
