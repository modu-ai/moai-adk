# SPEC-CONFIG-ATOMIC-WRITE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: draft
plan_complete_at: pending
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: []
code_baseline: ed70e4354
```

- Artifacts authored at plan-phase: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
  Status `draft`, Tier M. Version 0.2.0 (iteration-2 re-author).
- Slice (b) of `SPEC-CONFIG-TIER-PERSIST-001`; owns the parent's §D.4 REQ-CTP-021/022/023/024/027
  as REQ-CAW-001..007. REQ-CTP-025/026 (mode-widening migration) DEFERRED to follow-up sibling
  SPEC `SPEC-CONFIG-MODE-MIGRATE-001` (not yet authored) per user pre-decision.
- Code baseline `ed70e4354` (worktree HEAD, clean, `0 0` vs `origin/main`). All `file:line`
  evidence in `spec.md` §A is attributable to this tree.
- **Iteration-2 re-author (version 0.2.0)** — addresses iteration-1 plan-auditor defects D1-D5
  that caused FAIL 0.74 (Tier M threshold 0.80):
  - D1 (self-contradiction): §A reframed as forward-looking invariant; §C `### Out of Scope —
    mode-widening migration` subsection added with explicit deferral.
  - D2 (incomplete site enumeration): REQ-CAW-007 target-path resolution algorithm added;
    `toolpolicy/tier_render.go:146` + `toolpolicy/codegen.go:244` added (REMEDIATE);
    `update_recovery_manifest.go:78` + `update_namespace_protect.go:242` classified EXEMPT
    (backup-dir artifacts, predicate-based).
  - D3 (placeholder helper name): locked to `atomicfile.Write` in
    `internal/config/atomicfile/` (sub-package isolation; importable by both `internal/config`
    and `internal/cli`).
  - D4 (uppercase SHALL in REQ-CAW-004a): corrected to lowercase `shall`; toolpolicy sites
    added to remediation list.
  - D5 (V3R5 prior-art qualifier): added "this SPEC does NOT supersede it" in §H.
- Live defects (manager.go:420-438 mode-narrowing; harness.go:390 non-atomic + hardcoded 0o644;
  update*.go bare `os.WriteFile`) re-verified at HEAD before authoring.
- Positive prior art confirmed: `yamlpatch.go:202` (Chmod-before-rename),
  `update_deny_migration.go:95-100` (stat-then-preserve), `update_disk_backup.go:39` (seam).
- Parent split memo at `SPEC-CONFIG-TIER-PERSIST-001/spec.md` `## §L Split Branches` — REQ
  mapping corrected to record REQ-CTP-025/026 deferral.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
