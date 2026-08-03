---
id: SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001
title: "Rendered template snapshot as the 3-way merge BASE for moai update"
version: "0.1.0"
status: in-progress
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tags: "update, yaml, merge, provenance, snapshot"
era: V3R6
depends_on: ["SPEC-UPDATE-YAML-PRESERVE-001"]
related_specs: ["SPEC-UPDATE-YAML-PRESERVE-001", "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002"]
tier: L
---

# SPEC — Rendered template snapshot as the 3-way merge BASE

## HISTORY

- 2026-08-03 — plan-phase artifacts authored (Tier L, 6 artifacts: spec.md + plan.md + acceptance.md + design.md + research.md + progress.md). This SPEC is the deferred follow-up explicitly named by `SPEC-UPDATE-YAML-PRESERVE-001` plan.md §E "Decision D5" and audit.md finding D9.
- 2026-08-03 — iter-1 plan-audit revision (FAIL 0.68 → targeted PASS ≥ 0.85). D1 authored design.md + research.md; D2 repaired AC-TBS-010 inverted intermediate assertion; D3 corrected quality.yaml AC cross-ref (AC-TBS-006 → AC-TBS-013); D4 dispositioned the `runUpdateRestore` third restore-completion site; D5 corrected basePath line citation (restore.go:107-108 → restore.go:118); D6 rephrased REQ-TBS-001/002 subject to "the moai snapshot subsystem". The prior SPEC merged as commit `3ced5b152` (PR #1313) implementing node-tree YAML merge to preserve comments/order/quoting, and explicitly deferred this provenance defect because (1) it is a different defect class (provenance, not representation), (2) it requires a new persisted artifact, (3) it carries its own backward-compatibility design surface, and (4) the harm ordering made representation-defect-first strictly safer. The prior SPEC's audit finding D9 explicitly warned that the prior SPEC **enlarges D5's blast radius for `quality.yaml`** — making this follow-up the owner of that enlarged blast radius.

## §A Context and Blast Radius

### The defect — provenance, not representation

`SaveTemplateDefaults` at `internal/cli/update/backup/backup.go:147-192` derives the 3-way merge BASE from the **NEW embedded template** — raw and unrendered. The `.tmpl` suffix is stripped, but `{{.Version}}`-style Go-template placeholders are left intact in the bytes. It is called from `BackupMoaiConfig` (`backup.go:115`), which writes BASE into every per-backup `.template-defaults/sections/` directory. That BASE is consumed during restore by `MergeYAML3Way` (`internal/cli/update/backup/merge.go:25`) via the call site at `restore.go:121` (the `basePath` read at `restore.go:118`).

A correct 3-way merge needs BASE = **what the user's previously-installed template version deployed (rendered)**. The merge's decision logic (`merge.go:47-52`) is:

- `old == base` → user did not change this key → adopt the new template value
- `old != base` → user customized this key → preserve the user value

Because today's BASE = NEW raw template (with `{{.Version}}`-style placeholders) while LOCAL (old user file) = a rendered file whose `version: 3.0.1` was resolved at install time, every key whose template shipped a placeholder differs between BASE and LOCAL. `old != base` therefore fires for every such key, the merge misreads template-blessed values as user edits, and stale values stick across updates. The 3-way merge collapses toward 2-way behaviour for exactly the keys that most need a fresh default. The code comment at `backup.go:180-184` explicitly rationalizes this as "correct behavior" — **that rationalization IS the defect**: it conflates "the key is structurally present in the template" with "the value is the user's starting point", and only the latter is what `base` means in a 3-way merge.

### Blast radius

`SaveTemplateDefaults` walks `internal/template/templates/.moai/config/sections/` and writes each file (raw, `.tmpl` stripped) as BASE. Every one of the 30 shipped section templates (`SPEC-UPDATE-YAML-PRESERVE-001` plan.md §A re-measured the count as 30) that carries a Go-template placeholder is a victim. The placeholder-bearing templates include `system.yaml.tmpl` (`version: "{{.Version}}"`), `llm.yaml.tmpl`, `lsp.yaml.tmpl`, `harness.yaml.tmpl`, `cache.yaml.tmpl`, and `quality.yaml.tmpl` — the majority of the section surface.

**`quality.yaml` — the newly-enlarged victim (prior SPEC D4 / D9).** Prior SPEC `spec.md:88` (§A blast radius) and `audit.md` D4 established that `quality.yaml.tmpl` is the sole `.tmpl` that leaves its placeholders **unquoted** (`enforce_quality: {{.EnforceQuality}}`, `test_coverage_target: {{.TestCoverageTarget}}`). Pre-YAML-PRESERVE, the map decoder failed on this file, so `quality.yaml` was silently routed through the 2-way fallback at `restore.go:139` on every `moai update` — D5's wrong-base problem could not manifest because the 3-way path never ran. Post-YAML-PRESERVE, the node decoder parses the file successfully (placeholders land as scalar-nested-flow-mapping text rather than failing), so `quality.yaml` transitions from *always-2-way* to *real-3-way*. On the very first such run, every `{{.Version}}`-style placeholder in the wrong BASE differs from the user's rendered value, `old != base` fires, and every placeholder-bearing key is misread as a user edit. **This SPEC owns that enlarged blast radius** (prior SPEC `audit.md:183` D9). `quality.yaml` is therefore the primary correctness-test target — see acceptance.md AC-TBS-013.

### What this SPEC does NOT re-litigate

The merge *mechanism* is settled by the prior SPEC: `MergeYAML3Way` operates on `*yaml.Node` trees, preserves comments/order/quoting, and is byte-stable on the no-edit case for 8/30 templates and property-stable for all 30. This SPEC changes only **which bytes are passed as `baseData`** to `MergeYAML3Way` — the prior SPEC's plan.md:146 contract: "the follow-up changes only *which bytes* are passed, not the merge's shape". The 2-way fallback (`MergeYAMLDeep` at `restore.go:139`) remains available for unparseable bases; the node decoder's success on `quality.yaml.tmpl` is preserved; the harm ordering (a stale value is strictly less damaging than a destroyed file) is preserved.

## §B Goals

1. Capture the **rendered** section templates at deploy time and persist them across the update cycle, so a subsequent `moai update` can read the *previously-deployed rendered* values as the 3-way BASE.
2. Read the snapshot as BASE on every `moai update`; fall back to the current (embedded-raw) behaviour when the snapshot is absent so pre-existing installs are not broken.
3. Make `quality.yaml` — the file the prior SPEC incidentally promoted from 2-way to real-3-way — a first-class correctness target, asserting that template-blessed placeholder keys are NOT misread as user edits when a snapshot is present.
4. Keep the snapshot gitignored (rendered files carry machine-specific absolute paths from `{{.GoBinPath}}`) and outside the update clean step's deletion radius.

## §C Requirements (GEARS)

**REQ-TBS-001** — At the end of a successful `moai init`, the moai snapshot subsystem shall persist a rendered-template snapshot of the `.moai/config/sections/*.yaml` files.

**REQ-TBS-002** — At the end of a successful `moai update` restore phase (after NEW templates are deployed and the 3-way/2-way merge has run), the moai snapshot subsystem shall persist a rendered-template snapshot of the `.moai/config/sections/*.yaml` files. This obligation covers every restore-completion site: the template-sync path, the clean-install path, AND the user-invocable `runUpdateRestore` path (`update_restore.go`).

**REQ-TBS-003** — When persisting the snapshot, the `moai update` subsystem shall copy the on-disk rendered bytes from `.moai/config/sections/<name>` verbatim into the snapshot directory, preserving the values produced by Go-template rendering (e.g. `version: 3.0.1`, not `version: {{.Version}}`).

**REQ-TBS-004** — The snapshot directory shall be located under `.moai/cache/template-snapshot/sections/` so that it survives the update clean step (which deletes `.moai/config/` but not `.moai/cache/`).

**REQ-TBS-005** — The snapshot directory shall remain gitignored, because rendered section files carry machine-specific absolute paths (e.g. `{{.GoBinPath}}` resolved) that MUST NOT be committed.

**REQ-TBS-006** — While a snapshot is present at `.moai/cache/template-snapshot/sections/`, `BackupMoaiConfig` shall populate the per-backup `.template-defaults/sections/` BASE directory from the snapshot bytes, NOT from the embedded-raw template.

**REQ-TBS-007** — When the snapshot directory is absent (pre-existing install, first run after this feature ships, or any state where no snapshot was recorded), `BackupMoaiConfig` shall fall back to the current behaviour (BASE = embedded-raw template via `SaveTemplateDefaults`) so that existing user projects continue to update without breakage.

**REQ-TBS-008** — The `MergeYAML3Way` signature (`func(newData, oldData, baseData []byte) ([]byte, error)`) shall remain unchanged; only which bytes are passed as `baseData` is modified.

**REQ-TBS-009** — The 2-way fallback (`MergeYAMLDeep`) shall remain available at `restore.go:139` for any unparseable BASE (whether snapshot-sourced or embedded-sourced), unchanged by this SPEC.

**REQ-TBS-010** — When a snapshot is present, a template-blessed key whose value differs between the prior template version and the new template version (e.g. `version:` changing across releases) shall be adopted from the NEW template, NOT misread as a user edit, because the snapshot records the prior rendered value and `old == base` correctly fires.

**REQ-TBS-011** — When a snapshot is present, a key the user genuinely customized (LOCAL differs from the snapshot BASE) shall be preserved across `moai update`, because `old != base` fires correctly against the rendered snapshot.

**REQ-TBS-012** — `quality.yaml` (the file the prior SPEC promoted from always-2-way to real-3-way) shall, when a snapshot is present, complete a real-3-way merge whose BASE is the rendered `quality.yaml` (with `{{.EnforceQuality}}` / `{{.TestCoverageTarget}}` resolved), and the placeholder-bearing keys shall NOT be misread as user edits.

**REQ-TBS-013** — When the snapshot is absent on the first `moai update` after this feature ships, the update shall complete cleanly via the REQ-TBS-007 fallback, and SHALL write the snapshot at the end of that update so subsequent updates get the correct BASE.

**REQ-TBS-014** — The snapshot write SHALL be best-effort and non-blocking: a snapshot write failure (disk full, permission denied) SHALL NOT fail the `moai init` or `moai update` that triggered it; the next update falls back per REQ-TBS-007.

**REQ-TBS-015** — The snapshot SHALL be scoped to `.moai/config/sections/` only, matching the current `SaveTemplateDefaults` scope. This SPEC does NOT widen the snapshot to all embedded templates.

## §D Non-Functional Constraints

- **NFR-TBS-001 (cross-platform)**: All snapshot path handling uses `filepath.Join`; cross-platform build verified by `GOOS=windows GOARCH=amd64 go build ./...`.
- **NFR-TBS-002 (test isolation)**: All snapshot tests use `t.TempDir()`; no test writes to the project root's real `.moai/cache/`.
- **NFR-TBS-003 (no new dependency)**: `gopkg.in/yaml.v3 v3.0.1` only; no new module added. The snapshot copy uses `io/fs` + `os.ReadFile` / `os.WriteFile`.
- **NFR-TBS-004 (template neutrality)**: No file is added or modified under `internal/template/templates/`. The snapshot lives in the user project's `.moai/cache/`, never in the template source.
- **NFR-TBS-005 (no subagent boundary violation)**: Snapshot code in `internal/cli/...` MUST NOT call `AskUserQuestion`. All snapshot decisions resolve from on-disk state and exit codes.
- **NFR-TBS-006 (harm ordering preserved)**: The snapshot fix MUST NOT regress the prior SPEC's YAML-preservation correctness (comments / order / quoting / old-only keys). A regression on any of those axes blocks merge.
- **NFR-TBS-007 (snapshot survives clean step)**: The update clean step (`update_cleanup.go` / `update_clean_install.go`) MUST NOT delete `.moai/cache/`. The snapshot location is chosen specifically because `.moai/cache/` is already outside the clean radius and already gitignored.

## §E Scope Boundary — WHAT and WHY (not HOW)

This SPEC fixes the provenance of the 3-way merge BASE. It states WHAT the BASE shall be (the previously-deployed rendered template) and WHY (so the merge can distinguish user changes from template changes). It does NOT specify function names, file layout within the new snapshot package, or the copy mechanism — those are run-phase decisions, captured in plan.md §E as a Decision set but not binding at the requirement layer.

## §F Exclusions

### Out of Scope — representation / merge mechanism

- The node-tree merge, comment/order/quoting preservation, and old-only-key retention are OWNED by the merged `SPEC-UPDATE-YAML-PRESERVE-001`. This SPEC does not redesign the merge. Per the prior SPEC's plan.md:146 contract, only which bytes are passed as `baseData` changes; `MergeYAML3Way`'s shape is frozen.

### Out of Scope — snapshot widening

- Snapshotting `.claude/` templates, `.moai/project/`, or any non-section artifact is OUT OF SCOPE. The snapshot matches the current `SaveTemplateDefaults` scope (`.moai/config/sections/` only). Widening is a separate SPEC's decision.

### Out of Scope — historical snapshot migration

- Pre-existing installs (no snapshot on disk when this feature ships) are handled by REQ-TBS-007 + REQ-TBS-013 (graceful fallback → first update writes the snapshot). This SPEC does NOT attempt to retroactively reconstruct a "what would have been deployed" snapshot from git history or any other source; the first post-feature update creates the first real snapshot.

### Out of Scope — per-backup `.template-defaults/` removal

- The per-backup `.template-defaults/` directory (written by `BackupMoaiConfig` at `backup.go:114`) is RETAINED as the per-backup BASE carrier that `RestoreMoaiConfig` already reads (`restore.go:118`). This SPEC changes the *source* of those bytes (snapshot when present, embedded-raw as fallback) but does NOT remove the per-backup directory or change `RestoreMoaiConfig`'s read path.

### Out of Scope — sync_commit_sha backfill chore

- The prior SPEC's `progress.md §E.4 sync_commit_sha` backfill is a separate chore. This SPEC is the provenance fix ONLY. Bundling the chore would violate scope discipline (CLAUDE.local.md §16 Rule 5).

### Out of Scope — snapshot signing / integrity

- The snapshot is a rendered copy of the on-disk section files; no HMAC, signature, or tamper-evidence is added. The threat model is the same as `.moai/config/` itself (single-user offline machine, per `SPEC-SEC-HARDEN-003` §F.1).

## §G Acceptance Criteria Cross-Reference

Each requirement maps to one or more acceptance criteria in `acceptance.md` (§D AC matrix). The mandatory falsifiability AC (AC-TBS-010, per the prior SPEC's D5-repaired pattern at acceptance.md AC-UYP-022) reverts the base-derivation and asserts the test FAILS against the wrong-base implementation, proving the correctness ACs are non-vacuous.

## §H Cross-References

- `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/plan.md` §E Decision D5 (deferral rationale, 4-point argument, harm ordering)
- `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/audit.md` D9 (blast-radius enlargement for `quality.yaml`)
- `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/spec.md` §A (D4 finding: `quality.yaml.tmpl` unquoted placeholders, 2-way fallback today)
- `internal/cli/update/backup/backup.go:147-192` — `SaveTemplateDefaults` (defect site)
- `internal/cli/update/backup/backup.go:114-118` — per-backup `.template-defaults/` write site
- `internal/cli/update/backup/restore.go:118-121` — BASE read (`basePath` at `:118`) + `MergeYAML3Way` call site (`:121`)
- `internal/cli/update/backup/restore.go:139` — 2-way fallback (`MergeYAMLDeep`), unchanged
- `internal/cli/update_restore.go:53` — `runUpdateRestore` distinct restore-completion site (the lockout-escape entry); dispositioned in plan.md §B + Decision D4
- `internal/cli/update/backup/merge.go:25` — `MergeYAML3Way` signature (frozen by REQ-TBS-008)
- `internal/defs/dirs.go` — `MoAIDir`, `ConfigSubdir`, `SectionsSubdir` path constants
- `CLAUDE.local.md` §2 (`.moai/cache/` gitignored), §16 Rule 5 (scope discipline), §6 (test isolation)
