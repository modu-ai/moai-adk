# Implementation Plan — SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001

> Milestones are ordered by decision-reversibility: the highest-change-likelihood decisions (Defect B's allowlist-derivation SOURCE — a data-model / SSOT choice, and Defect A's detection-predicate shape) lead, so human review focuses there. Mechanical golden-test regeneration and verification are deferred to the tail.

## §A Context

### §A.1 Affected files (run-phase scope)

| File | Change | Defect |
|------|--------|--------|
| `internal/cli/doctor_skills.go` | Replace `staticCoreAllowlist` static slice with a manifest-derived source (or a manifest-load helper feeding `classifySkill`) | B |
| `internal/cli/doctor_skills_test.go` | Add derivation-parity + genuine-unknown tests | B |
| `internal/cli/doctor_harness.go` | Adjust the "harness configured" detection predicate to exclude telemetry-only directories | A |
| `internal/cli/doctor_harness_test.go` | Add telemetry-only → OK repro + genuine-harness-preserved + partial-baseline edge tests | A |
| `internal/cli/doctor.go` | (Defect B) update the `checkSkillsAllowlist` remediation hint wording if the derivation changes the message | B |
| `internal/cli/doctor_golden_test.go` | Regenerate golden output for Skills Allowlist + Harness 5-Layer if the golden snapshot captures either signal | A + B |
| (manifest read helper) | A small helper to load the embedded template `moai-*` skill set from `internal/template` (catalog.yaml or embedded skills FS) — SSOT for the allowlist | B |

### §A.2 PRESERVE targets (do NOT modify)

- `internal/cli/update.go`, `internal/harness/applier.go` — CLEAN-REINSTALL-002 clean-reinstall path (REQ-DFS-008).
- `internal/cli/harness.go` `harnessDefaultLogPath`, `internal/cli/hook.go` observe-hook write paths — the DEFAULT Defect A fix is doctor-side; do NOT relocate the telemetry write path (§C Out of Scope).
- All `internal/cli/update_preserve_*_test.go`, `update_namespace_harness_v*_test.go` — must stay green unchanged.
- Runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/*`).

### §A.3 Baseline evidence (measured at plan-phase)

- Embedded template ships **27** `moai-*` skill directories under `internal/template/templates/.claude/skills/`.
- `staticCoreAllowlist` in `doctor_skills.go` lists **22** names; comment claims "23" (a pre-existing off-by-one comment defect — fix opportunistically only if the slice is being replaced anyway).
- **10** embedded skills are absent from the allowlist (exact match to issue #1088), and **5** allowlist entries (`moai-workflow-design-context`, `moai-workflow-design-import`, `moai-workflow-gan-loop`, `moai-domain-brand-design`, `moai-domain-copywriting`) are stale (no longer in the template).
- `internal/template/catalog.yaml` carries `catalog.<tier>.skills[].name` + `path` entries and is embedded via `//go:embed catalog.yaml` in `internal/template/embed.go` — a candidate derivation source.

### §A.4 Tier + development-mode

Tier M (see §D). `development_mode` per quality.yaml; brownfield with existing doctor test coverage → cycle_type=tdd (reproduction test first — the RED test IS the acceptance criterion).

## §B Milestones

### M1 — Defect B: allowlist derivation SOURCE (highest-reversibility decision)

The load-bearing decision is WHERE the known-skills set comes from. Two candidate SSOT sources — settle in run-phase, surface as a blocker only if neither is viable:

1. **`internal/template/catalog.yaml`** (embedded) — parse `catalog.*.skills[].name`, filter to `moai-`-prefixed. Pros: already embedded, already the generated manifest. Cons: catalog includes the bare `moai` skill + non-skill entries — filter carefully.
2. **Embedded `templates/.claude/skills/` FS** (`//go:embed all:templates`) — enumerate directory names with a `moai-` prefix. Pros: byte-identical to what `moai update` actually installs; zero drift by construction. Cons: requires exposing an embedded-FS reader from `internal/template`.

RECOMMENDED: source (2) — the embedded skills FS — because REQ-DFS-006 ("skills the binary just installed") is literally the embedded FS, making the derivation tautologically drift-free. Implement `classifySkill` to consult the derived set; retain a graceful fallback (empty derived set → do not spuriously WARN).

Deliverable: manifest-read helper + `classifySkill` derivation wiring + RED test AC-DFS-004 (fresh-install → 0 unknown) passing.

### M2 — Defect B: genuine-unknown + anti-drift invariant

Preserve the check's real purpose. Add AC-DFS-005 (every embedded manifest `moai-*` skill classifies non-WARN — the anti-drift invariant test that catches a future skill added to templates without the derivation picking it up) and AC-DFS-006 (a bogus `moai-nonexistent-xyz` dir still WARNs, count==1). Update the `checkSkillsAllowlist` remediation hint in `doctor.go` to be actionable for a genuine stale/third-party skill (REQ-DFS-007).

### M3 — Defect A: telemetry-exclusion detection predicate

Change the "harness configured" gate in `runHarnessCheck`. Replace the bare `os.Stat(harnessDir)` existence test with a predicate that treats a directory containing ONLY `usage-log.jsonl` (and other non-baseline runtime files) as "not configured". Concretely: the harness is "configured" iff at least one of the 7 `checkLayer5Files` baseline files exists in `.moai/harness/`. Deliverable: AC-DFS-001 (telemetry-only → CheckOK) RED test passing; REQ-DFS-002/004.

### M4 — Defect A: preserve genuine-harness evaluation + edge

Guard against over-suppression. AC-DFS-002 (all 7 baseline files present + usage-log.jsonl present → full L1-L6 battery runs, not short-circuited to OK) and AC-DFS-003 (partial baseline — some but not all 7 files — is a genuinely misconfigured harness and MUST still reach FAIL, NOT be masked as "not configured"). REQ-DFS-003.

### M5 — Golden-test regeneration + preservation guard (mechanical tail)

- Regenerate `doctor_golden_test.go` snapshots so Skills Allowlist + Harness 5-Layer golden output reflects the corrected signals (AC-DFS-007).
- Verify REQ-DFS-008: run the CLEAN-REINSTALL preservation test suites unchanged; confirm no diff to `update.go` / `applier.go` (AC-DFS-008).
- Full-suite verification: `go test ./...` green + `GOOS=windows GOARCH=amd64 go build ./...` (AC-DFS-009).

## §C Technical Approach

- **Defect A**: minimal predicate change — introduce a `harnessBaselinePresent(harnessDir) bool` helper (checks the 7 `checkLayer5Files` names) and gate the L1-L6 battery on it instead of on directory existence. This localizes the fix and reuses the existing baseline-file list as the single "configured" definition (no new magic list).
- **Defect B**: introduce a manifest-derivation helper (recommended: embedded skills FS enumeration exposed by `internal/template`) and make `classifySkill` a function of that derived set rather than a package-level static slice. Keep the classification rules (PASS / WARN / INFO) identical — only the "known" set's SOURCE changes.
- Both changes stay inside `internal/cli/doctor_*` + a read-only helper into `internal/template`. No write-path, no update-path, no cross-platform primitive.

## §D Tier Judgment

**Tier M** (validates the Epic memory estimate). Rationale:
- **Files affected: 5-7** (two doctor source files + their tests + a manifest helper + golden test + a one-line message tweak) — squarely in Tier M's 5-15 range, above Tier S's <5.
- **LOC: ~250-450** — the derivation helper + predicate change + reproduction tests + golden regeneration land in Tier M's 300-1000 band (low end), above Tier S's <300.
- **Cross-cutting**: two distinct defects, a new SSOT-derivation mechanism (allowlist ← embedded manifest), golden-snapshot regeneration, and a preservation-contract guard against a sibling SPEC — more coordination surface than a Tier S single-fix. Not Tier L: no constitutional change, no >15 files, no new architecture.
- plan-auditor PASS threshold for Tier M: **0.80**. Artifact set: 3 files (spec.md + plan.md + acceptance.md).

## §E Progress Signals

See `progress.md` for the §E lifecycle skeleton. Plan-phase audit-ready signal is recorded in `progress.md §E.1`.

## §F Risks & Anti-Patterns

- **R1 — Over-suppression (Defect A)**: a fix that makes the harness check "OK" whenever baseline files are incomplete would MASK a genuinely broken harness. Mitigated by AC-DFS-003 (partial baseline must still FAIL) — the "configured" gate is "≥1 baseline file present", and once configured the full L1-L6 battery (including L5's all-7-files requirement) still runs.
- **R2 — Derivation includes non-skill entries (Defect B)**: `catalog.yaml` carries the bare `moai` skill and potentially non-skill rows; an FS enumeration is cleaner. Filter to directory names with a literal `moai-` prefix (matching `classifySkill`'s own prefix test) to avoid classifying `moai` (no trailing dash) as a skill.
- **R3 — Golden-test brittleness**: the golden snapshot may encode the exact FAIL/warning strings; regenerate deliberately (M5), do NOT blind-accept — confirm each golden delta corresponds to a corrected signal, not an unrelated drift.
- **R4 — Scope leak into update path**: tempting to "also fix" the misleading `moai update` hint by touching update.go. Do NOT — REQ-DFS-008 confines the fix to the doctor + a read-only manifest helper. The hint reword lives in `doctor.go`, not update.
- **R5 — Anti-drift test coupling**: AC-DFS-005 asserts the derived set covers every embedded `moai-*` skill; if derivation source (1) catalog.yaml is chosen and catalog.yaml lags the FS, the invariant could false-fail. Source (2) FS enumeration avoids this coupling.

## §G @MX Tag Targets (identify only — create in run-phase)

| Location | Tag | Rationale |
|----------|-----|-----------|
| `doctor_harness.go` `runHarnessCheck` (the reworked "configured" gate) | `@MX:ANCHOR` + `@MX:REASON` | The telemetry-exclusion invariant is the contract that closes #1087; fan_in from doctor check registry ≥1 and it is the false-signal locus. |
| `doctor_harness.go` new `harnessBaselinePresent` helper | `@MX:NOTE` | Documents that "configured" == "≥1 baseline file present", not "directory exists". |
| `doctor_skills.go` `classifySkill` (manifest-derived) | `@MX:ANCHOR` + `@MX:REASON` | The known-set-from-manifest invariant is the contract that closes #1088; classification is the false-warning locus. |
| `doctor_skills.go` removed `staticCoreAllowlist` / new derivation helper | `@MX:NOTE` | Records the SSOT source (embedded template manifest) so a future editor does not reintroduce a hardcoded list. |
| `doctor.go` `checkSkillsAllowlist` remediation hint | `@MX:NOTE` | Notes the hint must stay actionable for genuine stale/third-party skills (REQ-DFS-007). |

## §H Cross-References

- Root theme + preservation constraint: `spec.md` §A.1 / §A.4.
- CLEAN-REINSTALL-002 preservation contract: `.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002/spec.md` §A.3.
- Harness 5-Layer origin: SPEC-V3R3-PROJECT-HARNESS-001; Skills Allowlist origin: SPEC-V3R3-HARNESS-001.
- Embedded template manifest: `internal/template/catalog.yaml` + `internal/template/embed.go` (`//go:embed`).
