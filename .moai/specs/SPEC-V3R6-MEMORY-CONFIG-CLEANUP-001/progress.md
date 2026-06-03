# Progress — SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001

## §E.2 Sync-phase Audit-Ready Signal

**sync_commit_sha**: 0dcd29647

**era**: V3R6

**ac_summary**: 10/10 Blocking AC PASS (AC-MCC-001 MemoryConfig removal grep 0 | AC-MCC-002 build + config tests exit 0 | AC-MCC-003 YAML + embedded mirror drift green | AC-MCC-004 const block + subsystem green | AC-MCC-005 template neutrality + mirror-drift green | AC-MCC-006a Gap C comment + warn present | AC-MCC-006b Gap C doctrine note + no logic change | AC-MCC-007 Gap B recorded + no code change | AC-MCC-008 no byte-cap wiring | AC-MCC-009 full suite green).

**sync_status**: `in-progress → implemented` transition applied; CHANGELOG.md [Unreleased] entry added; spec.md frontmatter status updated.

**deliverables**: 
- CHANGELOG.md entry (Removed section, SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001 bullet point added)
- spec.md frontmatter (status: in-progress → implemented; updated: 2026-06-03)
- progress.md §E.2 (this file, §E.2 section created)
- Git commits: will execute one sync commit per B12 discipline

**decision_notes**: Gap A (MemoryConfig removal) is a real defect (config theater). Gap B (byte-cap fabricated requirement) is evaluated-and-deferred per spec.md §B.2 with zero code change. Gap C (per-cwd memory divergence) is documented-only per `.moai/docs/memory-dir-resolution-doctrine.md` + comment additions to hook files; intentional alignment with Claude Code per-cwd model.

**known_limitations_deferred**: none.
