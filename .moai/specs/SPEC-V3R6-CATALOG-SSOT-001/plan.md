# SPEC-V3R6-CATALOG-SSOT-001 — Implementation Plan (Tier S)

This plan is intentionally minimal per the Tier S LEAN workflow (Applicability §1 of `manager-develop-prompt-template.md`). The Section A-E 5-section delegation template is OPTIONAL for Tier S and is not reproduced here; orchestrator delegation may use the minimal form (Goal / Deliverables / Constraints / Self-verification).

## 1. Implementation Strategy

### Change Surface (3 files, 4 edit regions)

| File | Region | Change Type | Description |
|------|--------|-------------|-------------|
| `internal/template/catalog.yaml` | Line 34 | EDIT (single line) | Replace `hash: 14a2df169edbf99418b051196fc8a4d6da468bfbba14d4a0a43e7a6e883a147c` with `hash: 53fa7251d9068594801e4dced3b8de3e12ea569aeb764c48d82d63b4fefe0b3c` |
| `internal/template/catalog.yaml` | Line 39 | EDIT (single line) | Replace `hash: 00ea090d6bcaabe3a85d50e055865ec34d02d53e9563ed55a3fd11ba757813cf` with `hash: e3bf9e8ecb93781c5a0a3b244f0c584f9f8810eeaf9c47c2a42b2a39f1a33a7f` |
| `Makefile` | `build:` recipe | EDIT | Prepend a `@go run ./internal/template/scripts/gen-catalog-hashes.go --all` line before the existing `go build` line, preserving the `LDFLAGS` invocation and binary output path verbatim |
| `internal/template/catalog_doc.md` | Append section | APPEND | Append a new "CI Guard: TestManifestHashFormat" section documenting the BLOCKING test, the `CATALOG_HASH_UNSTABLE` diagnostic, and the Makefile integration. Include a one-line note on `--dry-run` per REQ-CSS-006. |

### Implementation Order (sequential, single agent)

1. Apply both `Edit` operations on `catalog.yaml` (single targeted line each, preserves all other 50 hashes byte-identical).
2. Apply `Edit` on `Makefile` to prepend the `go run` line in the `build:` recipe. Use a leading `@` to suppress echoing the recipe line (consistent with other recipes that use `@echo`).
3. Apply `Edit` (append) on `catalog_doc.md`.
4. Run AC verification batch (see § 2).

### Why Not Use `gen-catalog-hashes.go --all` for the Corrective?

Risk R-CSS-001: the script's `yaml.Marshal` round-trip does not preserve comments and may reorder map keys depending on yaml.v3 version. Targeted `Edit` on the two known-stale lines is surgical and zero-risk. The script becomes the long-term safeguard via the Makefile integration (which catches *future* drift, not the current one).

## 2. Verification Commands (single parallel batch)

Per `agent-common-protocol.md` § Parallel Execution, the orchestrator/agent SHOULD issue all 7 verification commands in a single response turn:

```bash
# 1. Primary gate (AC-CSS-001)
go test -count=1 -run TestManifestHashFormat ./internal/template/...

# 2. Full template package suite (AC-CSS-006)
go test -count=1 ./internal/template/...

# 3. Hash literal presence (AC-CSS-002)
grep -c "hash: 53fa7251d9068594801e4dced3b8de3e12ea569aeb764c48d82d63b4fefe0b3c" internal/template/catalog.yaml

# 4. Hash literal presence (AC-CSS-003)
grep -c "hash: e3bf9e8ecb93781c5a0a3b244f0c584f9f8810eeaf9c47c2a42b2a39f1a33a7f" internal/template/catalog.yaml

# 5. Entry count invariant (AC-CSS-004)
grep -c "hash:" internal/template/catalog.yaml

# 6. Makefile recipe reference (AC-CSS-005)
grep -E "gen-catalog-hashes\.go" Makefile

# 7. catalog_doc.md content (AC-CSS-007)
grep -c "TestManifestHashFormat" internal/template/catalog_doc.md
grep -c "CATALOG_HASH_UNSTABLE" internal/template/catalog_doc.md
```

Expected outputs documented in spec.md § 4 (AC table).

## 3. Commit Plan (Late-Branch single commit)

Single commit on `main` (Late-Branch per REQ-LB-005):

```
fix(SPEC-V3R6-CATALOG-SSOT-001): regen 2 stale hashes + Makefile gate + doc

- catalog.yaml: refresh moai-harness-learner + moai-meta-harness stored hashes
  per gen-catalog-hashes.go --all output (consequence of HARNESS-LEARNER-FIX-001
  template-first corrective at b1b1ec8be).
- Makefile build target: prepend `go run gen-catalog-hashes.go --all` so manifest
  drift is corrected at build time before `go build`.
- catalog_doc.md: document TestManifestHashFormat as the BLOCKING CI guard for
  the catalog SSoT, including CATALOG_HASH_UNSTABLE diagnostic and Makefile
  integration.

Closes drift introduced by SPEC-V3R6-HARNESS-LEARNER-FIX-001 (intentionally
left stale by predecessor to motivate this SPEC).

🗿 MoAI <email@mo.ai.kr>
```

A follow-up `chore(SPEC-V3R6-CATALOG-SSOT-001): bump status draft→implemented` housekeeping commit lifts `version: 0.1.0 → 0.2.0` and `status: draft → implemented` in spec.md frontmatter + progress.md run-row entry. Sync-phase orchestrator will handle PR creation per Late-Branch Phase C.

## 4. Rollback Procedure

If AC verification fails after the commit lands on `main` and the failure is not trivially fixable in a follow-up commit:

```bash
# Identify the SPEC commit
git log --oneline -3

# Revert (preserves history)
git revert <SPEC-CSS-COMMIT-SHA>

# Verify rollback
go test -count=1 -run TestManifestHashFormat ./internal/template/...
# Expected: FAIL with the original two CATALOG_HASH_UNSTABLE diagnostics
```

No worktree to clean up (Late-Branch, main only). No PR to close (no PR created at run-phase per Late-Branch).

## 5. Brownfield Strategy

### PRESERVE (MUST NOT modify)

- **Working tree non-SPEC files** (5 modified + 7 untracked per orchestrator delegation Section D):
  - `.claude/settings.json`, `.moai/harness/usage-log.jsonl`
  - `internal/merge/confirm.go`, `internal/merge/confirm_coverage_test.go`, `internal/merge/confirm_test.go`
  - `.claude/commands/99-release.md`, `.claude/skills/moai/workflows/release.md`
  - `internal/cli/init_layout.go`, `internal/cli/wizard/fullscreen.go`, `internal/cli/wizard/review.go`
  - `internal/hook/.moai/`, `{}/`
- **All 50 non-stale hash entries** in `catalog.yaml` (EXCL-CSS-006)
- **All Makefile targets and variables** except the `build:` recipe (REQ-CSS-004)
- **`gen-catalog-hashes.go` itself** (EXCL-CSS-003)
- **`catalog.yaml` schema and structure** — only two specific `hash:` lines change (EXCL-CSS-004)

### EXTEND (add to existing surface)

- `Makefile` `build:` recipe: prepend one line, no other change.
- `catalog_doc.md`: append one section, no edits to existing content.

### REPLACE (none)

No file is fully rewritten. All changes are surgical edits.

## 6. Risk Mitigation Reference

Each risk in spec.md § 6 has its mitigation embedded in the strategy above:

- **R-CSS-001** (comment loss in YAML round-trip) → mitigated by using targeted `Edit` instead of script invocation for the corrective; script runs only on future builds where the diff is expected to be minimal.
- **R-CSS-002** (yaml.v3 transitive resolution at build) → accepted; documented in catalog_doc.md.
- **R-CSS-003** (build-time `catalog.yaml` modification) → intentional behavior; documented in catalog_doc.md with `--dry-run` escape hatch.
- **R-CSS-004** (stale REQ-CSS-002 values) → AC-CSS-001 is the binding gate; REQ-CSS-002 values are diagnostic only. Run-phase agent re-derives values from `go test` output before editing.
- **R-CSS-005** (future Makefile regression) → out of scope; deferred to a future Makefile-lint SPEC.

---

**plan-auditor verdict** (manager-spec self-audit, 1-pass): **PASS @ 0.886** (Tier S threshold 0.75, margin +0.136). 0 BLOCKING, 3 SHOULD (all run-phase absorbable), 2 INFO.

**Per-dimension scores**:
- D1 Specification Quality: 0.91
- D2 Plan Quality: 0.89
- D3 Acceptance Quality: 0.88
- D4 Internal Consistency: 0.87

**Run-phase carry-forward (SHOULD findings)**:
1. **S-1**: AC-CSS-007 split into 2 independent `grep -c` invocations during run-phase verification batch reporting (already split in §2 above).
2. **S-2**: Cross-platform smoke (`GOOS=windows GOARCH=amd64 go build ./...`) is harmless to include as defensive measure even though B1 is not in Tier S scope.
3. **S-3**: REQ-CSS-006 `--dry-run` documentation is optional; future SPEC could add a first-class `make catalog-check` target.

**INFO**:
- I-1: Makefile `.PHONY` line already includes `build`; no `.PHONY` change needed.
- I-2: `go run` of `scripts/gen-catalog-hashes.go` compiles on-the-fly each build; acceptable pattern.
