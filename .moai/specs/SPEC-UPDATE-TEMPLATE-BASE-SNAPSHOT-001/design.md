# Design — SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001

> Tier L design artifact. Bounds the snapshot data model, the lifecycle state machine, the first-run migration flow, and the non-regression boundary vs the YAML-preservation correctness shipped by `SPEC-UPDATE-YAML-PRESERVE-001`. Empirical claims (clean-step walk radius, restore-entry call graph) are owned by `research.md`; this file states the design that those measurements make safe.

## §A Scope of this document

This is the **design** layer: WHAT the snapshot IS, HOW it behaves across its lifecycle, and WHERE its boundaries sit. It is deliberately silent on function names and file layout within the new package (run-phase decisions, captured in `plan.md` §E as a Decision set) and on empirical claims about the existing clean-step radius (measured in `research.md` §A).

## §B Snapshot data model

### B.1 Location and naming

The snapshot is a directory tree at `.moai/cache/template-snapshot/`, relative to the project root:

```
<projectRoot>/
└── .moai/
    └── cache/
        └── template-snapshot/
            └── sections/
                ├── system.yaml         (rendered; version: "3.0.1", not {{.Version}})
                ├── quality.yaml        (rendered; enforce_quality: true, not {{.EnforceQuality}})
                ├── llm.yaml
                ├── … (one file per .moai/config/sections/*.yaml at deploy time)
                └── <user-created-sections>.yaml
```

**Layout invariants**:
- The `sections/` subtree mirrors `.moai/config/sections/` exactly — same relative paths, same filenames. A section at `.moai/config/sections/foo/bar.yaml` has its snapshot at `.moai/cache/template-snapshot/sections/foo/bar.yaml`.
- Filenames are the on-disk rendered names (`.yaml` / `.yml`), NOT the `.tmpl` source names. The `.tmpl` suffix is already stripped by the deploy step before the on-disk file exists, so the snapshot copies the on-disk name verbatim.

### B.2 Content semantics

Each snapshot file is a **byte-identical copy** of the corresponding on-disk rendered section file at the moment the snapshot is written. Specifically:
- Go-template placeholders are RESOLVED (`version: "3.0.1"`, not `version: {{.Version}}`).
- Machine-specific absolute paths from `{{.GoBinPath}}` are baked in — this is why the snapshot MUST be gitignored (REQ-TBS-005).
- Comments, key order, quoting, and whitespace match the on-disk file exactly. The snapshot is NOT re-marshalled or re-rendered; it is a file copy.

**Why byte-copy and not re-render** (Decision D3 in plan.md): re-rendering would require reconstructing the `TemplateContext` (GoBinPath, HomeDir) at snapshot-write time, introduces a second rendering site that could drift from the deploy-time render, and re-renders `system.yaml.tmpl`'s `version: {{.Version}}` against the binary's version — which may differ from the version recorded in the on-disk `system.yaml` if the user is mid-update. The on-disk file IS the rendered truth; copying it is both simpler and more correct.

### B.3 Scope boundary

The snapshot is scoped to `.moai/config/sections/` ONLY (REQ-TBS-015). This matches the current `SaveTemplateDefaults` walk scope and the current `RestoreMoaiConfig` walk scope. Specifically NOT snapshotted:
- `.claude/` templates (different deploy path, different merge machinery)
- `.moai/project/` (user-authored documentation, not template-sourced)
- `.moai/config/*.yaml` at the config root (outside sections/; not part of the merge surface)
- Any non-section artifact

### B.4 User-created sections

Files in `.moai/config/sections/` that the user created (no template counterpart) are copied into the snapshot like any other file. This is harmless:
- The 3-way merge at restore time walks the BACKUP, not the snapshot directly. A user-created section in the backup routes through the "target file does not exist" branch at `restore.go:89-95` (user's custom config not in new template → restore as-is). The snapshot's copy of that user-created section is never read by the merge for that file.
- The snapshot's copy DOES serve as the BASE for the next cycle's 3-way merge of that user-created section — and since NEW has no counterpart, the `old == base` test is moot. The merge correctly preserves the user's value.

No special handling for user-created sections is required.

## §C Lifecycle state machine

The snapshot transitions through four states:

```
   (project created)        moai init          moai update          moai update
         │                     │                   │                    │
         ▼                     ▼                   ▼                    ▼
   ┌─────────────┐       ┌──────────┐        ┌──────────┐         ┌──────────┐
   │   ABSENT    │──────▶│ PRESENT  │◀──────▶│ PRESENT  │◀───────▶│ PRESENT  │
   │ (no snapshot)│  init │ (fresh   │ update │ (rewritten│ update  │ (rewritten│
   └─────────────┘  writes│  install │ rewrites│ with post-│ rewrites │ with post-│
        │                     │  baseline)│      │  restore   │         │  restore   │
        │                     └──────────┘      │  state)    │         │  state)    │
        │                            ▲          └──────────┘         └──────────┘
        │              moai update    │
        │              (no snapshot)  │
        │              falls back +   │
        └─────────────writes at end───┘
```

**States**:
1. **ABSENT** — no `.moai/cache/template-snapshot/sections/` directory exists. Entered: brand-new project before any `moai init`; OR user manually deleted `.moai/cache/`; OR the snapshot write failed best-effort on a prior cycle.
2. **PRESENT (fresh-install)** — written by `moai init` (REQ-TBS-001). Captures the install-time rendered baseline.
3. **PRESENT (post-update)** — written/overwritten by `moai update` restore-completion (REQ-TBS-002). Captures "what this update just produced".
4. **PRESENT (post-restore)** — written/overwritten by `runUpdateRestore` (the lockout-escape path; REQ-TBS-002 third site). Captures the chosen-backup merge result.

**Transitions**:
- `ABSENT → PRESENT`: any of the four write triggers fires (init end, three restore-completion sites).
- `PRESENT → PRESENT`: any update / restore cycle rewrites the snapshot in place. The snapshot is NOT append-only and NOT versioned; it always reflects the most recent deploy. (Versioning is OUT OF SCOPE — see spec.md §F.)
- `PRESENT → ABSENT`: only by user action (deleting `.moai/cache/`) or filesystem failure. No code path deletes the snapshot. The next update falls back per REQ-TBS-007 and re-creates it at restore end (REQ-TBS-013).

**Idempotency**: writing the snapshot twice from the same on-disk state produces the same bytes (file copy is deterministic). Running `moai update` twice in succession leaves the snapshot at the same state.

## §D First-run migration flow (REQ-TBS-007 + REQ-TBS-013)

The migration problem: every existing user project has NO snapshot when this feature ships. The first `moai update` after the feature ships must not break.

**Flow**:

```
1. User runs `moai update` on a pre-existing install (snapshot ABSENT).
2. BackupMoaiConfig runs.
   └─ SaveTemplateBase(destDir, projectRoot) is called.
      └─ HasSnapshot(projectRoot) == false → delegates to SaveTemplateDefaults(destDir)
         (today's embedded-raw BASE; the wrong-base behaviour persists for THIS cycle).
3. Clean + deploy + RestoreMoaiConfig run, using the embedded-raw BASE.
   └─ Merge behaves exactly as it does today (wrong base, but no worse).
4. Update completes successfully (exit 0). REQ-TBS-013 (a) satisfied.
5. WriteSnapshot(projectRoot) fires at the restore-completion site.
   └─ Snapshot is CREATED at .moai/cache/template-snapshot/sections/.
6. NEXT `moai update`:
   └─ SaveTemplateBase finds the snapshot → reads rendered BASE.
      └─ Merge now correctly distinguishes user changes from template changes.
```

**Migration cost**: exactly one update cycle of "still wrong base" for pre-existing installs. Acceptable because:
- Strictly better than today (today is always wrong base; post-feature is wrong base only for the first cycle).
- Harm ordering (prior SPEC plan.md §E D5): a stale value is strictly less damaging than a destroyed file. The representation defect (destroyed files) was fixed first; this provenance defect (stale values) fixes the second-order harm.
- No way to reconstruct a historically-correct snapshot from the current on-disk state without making the very assumption this SPEC invalidates (that BASE = NEW embedded template).

**Rejected — seed the snapshot from embedded-raw BEFORE the first update's backup read**: would make the first update's BASE the NEW embedded template (today's wrong behaviour) AND would overwrite any future-correct snapshot on every run. No benefit over the fallback.

**Rejected — block the first update on snapshot creation**: violates REQ-TBS-013 (update must complete cleanly). Updates MUST NOT depend on a snapshot existing.

## §E Non-regression boundary vs SPEC-UPDATE-YAML-PRESERVE-001

The prior SPEC (`SPEC-UPDATE-YAML-PRESERVE-001`, merged as `3ced5b152` / PR #1313) shipped the representation fix: `MergeYAML3Way` operates on `*yaml.Node` trees, preserves comments / key order / scalar quoting / old-only keys. Its correctness ACs (`preserve_golden_test.go`) MUST NOT regress.

### E.1 What this SPEC changes (and what it does NOT)

**Changes**:
- The *source* of the `baseData` bytes that `BackupMoaiConfig` writes into the per-backup `.template-defaults/sections/` directory. Today: embedded-raw. After: snapshot (rendered) when present, embedded-raw as fallback.

**Does NOT change**:
- `MergeYAML3Way` signature (`func(newData, oldData, baseData []byte) ([]byte, error)` — frozen, REQ-TBS-008).
- `RestoreMoaiConfig`'s read path (continues to read BASE from `backupDir/.template-defaults/sections/<name>` at `restore.go:118`).
- The 2-way fallback at `restore.go:139` (`MergeYAMLDeep` — frozen, REQ-TBS-009).
- The merge decision semantics (`old == base` → adopt new; `old != base` → preserve old).
- The old-only-key preservation + stderr advisory.

### E.2 Boundary invariant

The snapshot fix is a **strict subset** of the merge's input space: it changes only which bytes feed the `baseData` parameter. The merge's internal behaviour on those bytes is unchanged. Therefore:

- If the snapshot bytes parse as YAML (they will — they are copies of on-disk rendered section files, which the deploy step already validated), the merge behaves identically to today, just with a different (correct) base.
- If the snapshot bytes are unparseable (corruption, disk error), the merge errors out and the 2-way fallback fires — exactly the existing contract for any unparseable base. The snapshot does not introduce a new failure mode.

### E.3 Non-regression test (AC-TBS-022)

The prior SPEC's `preserve_golden_test.go` runs against the merged tree. Every test in that file MUST still pass. This is the load-bearing non-regression gate: if any preservation test fails after the snapshot wiring, the snapshot fix has invaded the merge's representation contract and the SPEC is not done.

Additionally, the snapshot wiring MUST NOT change the merge's output for the no-user-edit case: with `new == old == base` (snapshot matches LOCAL matches NEW), the merged output must be byte-identical to the prior SPEC's output for the same input. This is covered by the prior SPEC's existing golden tests.

### E.4 Harm ordering preserved

The prior SPEC's harm ordering (a stale value is strictly less damaging than a destroyed file) is preserved by this SPEC in two ways:
1. The snapshot fix does not touch the representation layer, so it cannot introduce file destruction.
2. The snapshot fix's failure mode (wrong base → stale value preserved) is strictly less harmful than the representation layer's failure mode (comments/order/quoting lost), and the representation layer is the one this SPEC depends on, not the one it changes.

## §F Concurrency and threat model

**Single-user offline model**: this SPEC inherits the threat model of `SPEC-SEC-HARDEN-003` §F.1 (single-user offline machine). Specifically:
- The snapshot write is NOT locked. A concurrent `moai update` on the same project would race on the snapshot directory. This is the same trust model as `.moai/config/` itself; the SPEC does not add locking.
- The snapshot is NOT signed or integrity-protected. A tampered snapshot produces a wrong BASE, the merge misreads tampered values as user edits, and stale values stick — but this is the same failure mode as a tampered `.moai/config/`, which is already in the threat model.
- Path containment: the snapshot writer's walk root is `<projectRoot>/.moai/config/sections/` (a fixed, cleaned path). It does NOT follow symlinks (best-effort copy uses `os.ReadFile` which follows links — but the source is a template-deployed directory the deploy step already sanitized; the restore-side symlink guards at `restore.go` are unchanged).

## §G Open design questions (deferred to run-phase, NOT blocking)

- **Snapshot directory file mode**: `.moai/cache/` is typically `0o755`; individual snapshot files inherit `0o644` from `os.WriteFile`. No special mode is required; this matches the rest of `.moai/`.
- **Snapshot pruning**: the snapshot is overwritten on every cycle; it does NOT accumulate. No pruning logic is needed. If a future SPEC adds versioned snapshots, pruning becomes relevant — out of scope here.
- **Snapshot size**: bounded by the size of `.moai/config/sections/` (the 30 shipped section templates total < 100KB). No compression or size guard is warranted.

These are run-phase mechanical decisions, not plan-phase design questions. They do not change the data model, the lifecycle, or the non-regression boundary.
