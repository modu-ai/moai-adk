# t1527 investigation map — moai init/update user-visible output (card t1527, lane-24)

Measured against tree `.moai/worktrees/t1453` @ `d0e54b322` (WT-github-flow-default;
init/update sources untouched by the recent develop landings t1520/t1521/t1524 — attribution
noted, run-phase will re-verify on a fresh develop-based tree). Full-surface read performed by
a read-only Explore sweep (84 tool calls, 2026-10-05).

## Defect → producing call sites

1. **Version banner ×3, shuffled**: `update.go:176-177` KV "Current version" (top of every
   runUpdate) → binary-update step repeats New+Current (`update.go:718-720`) → `syscall.Exec`
   re-exec replays runUpdate → emitter #1 again → `update_template_sync.go:166-171`
   `renderIdentityBand` (◆, `update_tux.go:76-81`) + a THIRD KV "Current version". Even without
   a binary update: 1×KV + 1×◆ band + 1×KV = 3. The shuffle is the process boundary.
2. **Changed-files table, conflicts undifferentiated**: TUI rows `update/preview_tui.go:124-203`
   (conflicts sort first since t694, `displayRank` :225-236); fallback dump ALL rows
   `update/preview_fallback.go:92-123` (no gate). Data for summary+conflict-detail ALREADY
   EXISTS: `classifyUpdateCounts` + `renderClassificationSummary` pill (`update_tux.go:90-125`,
   printed pre-confirm at `update_template_sync.go:256-259`); `merge.MergeAnalysis.Files` carries
   `RiskLevel low|medium|high` (`internal/merge/types.go:97-103`). Gap: post-deploy
   `renderUpdateOutcome` (`update_tux.go:172-217`) consumes only `len(analysis.Files)` — class
   counts never carried to the end-of-run summary.
3. **Progress double-render + mismatched 5-stage**: per-step `tui.ProgressLine` ✓ lines
   (`update_template_sync.go:343-644` region) + `renderDeployProgress(done,total)` gauge after
   EVERY step (`update_template_sync.go:737-740`, glyph hardcoded ✓ `update_tux.go:135-139`) +
   a THIRD surface: `printerReporter` phase lines to STDERR (`reporter.go:46-100`,
   early phases only — the deploy-step double-render is already consciously suppressed at
   `update_template_sync.go:534-540`). Stage labels come from the 5-entry steps slice (:330-504)
   but the loop intercepts "Backup" and "Restore Settings" before `step.execute` (:541-735) —
   the labeled Backup step's closure is dead; stage 5 actually does merge/retrack work.
4. **Plugin-install failure buried**: NO plugin installer exists in this repo ("Install it
   yourself" string: 0 hits in .go/.sh/templates). Closest analogue: MCP provision failure →
   stderr warning mid-noise, not aggregated (`init.go:254-264`). The rc.25 observed emitter is
   NOT identifiable from this tree's code — run-phase sandbox capture must pin it (likely from
   the installed main-line binary or a shell-invoked plugin flow).
5. **Prefix zoo, no severity scheme**: ✓ (`ProgressLine.Done`, `uikit.SymSuccess`,
   `report.RenderOutcome`, `StatusIcon("ok")`), · (`update_tux.go:249,264`, shell-env list
   `update.go:794`), note:/Note:/Tip: (`init.go:206`, `worktree_advisory.go:39/48`,
   hook installers), advisory:/Advisory: (`update_model_key_strip.go:134`, `backup/node_merge.go:60`,
   `merge/merge.go:338`, `migrate_local_instructions.go:180-184`), Warning:/warning:/hint:
   (three casings — `printer.go:304`, hook installers, `init.go:197/260`,
   `update_codex_wiring.go:42-81`, `update_mirror_heal.go:101`, `update_noise.go:113/119`).
   `tui.CheckLine` states (ok/warn/err/run/info/skip, `table.go:60`) EXIST but most emitters
   bypass with raw Fprintf + uikit symbols.
6. **End-of-run information wall** (15 tail emitters, pure call-site order, stdout/stderr
   interleave): outcome pill+notes (:759) → undeployed-codex warning :762 → hooks guidance :763
   → `-c` hint :765 → global-env :770 → pre-push :775 → pre-commit :778 → worktree advisory :783
   → codex-wiring warnings (update.go:521) → skill-mirror repair (:533) → "Post-sync steps"
   section (:556) → memory-profile migration advisory (:564-568, the "48 files" line, raw
   Fprintln) → evolution scaffold (:572) → profile-sync warnings (:577-591). `moai init` HAS an
   action block (`buildInitSuccessCard`, `init_warnings.go:99-123`) but it prints BEFORE the
   tail (:919-921), so tail warnings land after it; `moai update` has no action block at all.

## Existing mechanisms the design can reuse

- Classification SSOT + counts: `class.go`, `classifyUpdateCounts`, `renderClassificationSummary`.
- Outcome card infra: `internal/cli/update/report/outcome.go` (`✓ Up to date / ✓ Updated N files`).
- CheckLine severity states; printer modes (ModePlain when piped = de-facto quiet; `--verbose`
  expands merge-fallback + retained-key lists; ModeJSON exists but update never overrides mode).
- Env suppression pattern: `MOAI_SKIP_BINARY_UPDATE` (re-exec flag precedent, envkeys.go).
- init's success card as the ACTION-REQUIRED block precedent.

## Open items for run-phase

1. Sandbox init→update live capture (before) to pin the rc.25 "Install it yourself" emitter and
   record the baseline transcript for after-comparison.
2. ko/en locale verification of these surfaces (card requirement; the sweep found no locale
   layer on init/update output — currently English-only; confirm and record the decision).
3. Re-verify line numbers on the fresh develop-based tree (this map is measured on
   WT-github-flow-default @ d0e54b322).
