# Acceptance — SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001

## §A Verification Philosophy

Every AC in this matrix is binary-testable: it passes or fails against an observable, reproducible signal (test exit code, file existence, byte comparison, grep count). No subjective judgments. The GEARS obligation lives in `spec.md` §C (REQ-TBS-001 .. REQ-TBS-015); this file states the observable evidence for each.

The mandatory falsifiability AC (AC-TBS-010) mirrors the prior SPEC's AC-UYP-022 pattern: a RED step that reverts the implementation and asserts the correctness test FAILS against the wrong-base code, proving the AC is non-vacuous.

## §B Traceability Summary

| REQ | AC(s) |
|---|---|
| REQ-TBS-001 (snapshot at end of `moai init`) | AC-TBS-001 |
| REQ-TBS-002 (snapshot at end of `moai update` restore) | AC-TBS-002 |
| REQ-TBS-003 (snapshot = rendered bytes, not `{{.Version}}`) | AC-TBS-003 |
| REQ-TBS-004 (snapshot under `.moai/cache/`, survives clean step) | AC-TBS-004 |
| REQ-TBS-005 (snapshot gitignored) | AC-TBS-005 |
| REQ-TBS-006 (BackupMoaiConfig prefers snapshot when present) | AC-TBS-006a, AC-TBS-006b |
| REQ-TBS-007 (fallback when snapshot absent) | AC-TBS-007 |
| REQ-TBS-008 (MergeYAML3Way signature unchanged) | AC-TBS-008 |
| REQ-TBS-009 (2-way fallback remains available) | AC-TBS-009 |
| REQ-TBS-010 (template-blessed key adopted, not misread) | AC-TBS-010 (falsifiability), AC-TBS-011 |
| REQ-TBS-011 (user-customized key preserved) | AC-TBS-012 |
| REQ-TBS-012 (quality.yaml real-3-way correctness) | AC-TBS-013 |
| REQ-TBS-013 (first-update-after-feature completes cleanly + writes snapshot) | AC-TBS-014 |
| REQ-TBS-014 (snapshot write best-effort non-blocking) | AC-TBS-015 |
| REQ-TBS-015 (scope limited to `.moai/config/sections/`) | AC-TBS-016 |

Cross-cutting ACs: AC-TBS-017 (cross-platform build), AC-TBS-018 (subagent boundary grep), AC-TBS-019 (template-neutrality grep), AC-TBS-020 (no-new-dependency), AC-TBS-021 (coverage non-regression), AC-TBS-022 (harm-ordering: YAML-preservation non-regression).

## §C Severity Model

- **MUST-PASS**: AC-TBS-001, 002, 003, 004, 005, 006a, 007, 008, 009, 010 (falsifiability — itself a MUST-PASS), 011, 013, 014, 015, 016, 022.
- **SHOULD-PASS**: AC-TBS-006b, 012, 017, 018, 019, 020, 021.

## §D Acceptance Criteria Matrix (Given-When-Then)

### AC-TBS-001 — Snapshot written at end of `moai init`

**Given** a fresh `t.TempDir()` project with no `.moai/cache/template-snapshot/`
**When** `moai init` completes successfully (templates deployed)
**Then** `.moai/cache/template-snapshot/sections/` exists and contains one file per `.yaml`/`.yml` in `.moai/config/sections/`.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestInit_WritesSnapshot -count=1 -v
grep -q '^--- PASS: TestInit_WritesSnapshot' <(go test ./internal/cli/update/backup/ -run TestInit_WritesSnapshot -count=1 -v 2>&1)
```

### AC-TBS-002 — Snapshot written at end of every `moai update` restore-completion site

**Given** a `t.TempDir()` project with a populated `.moai/config/sections/` and no snapshot
**When** `moai update`'s restore phase completes successfully on ANY of the three restore-completion sites (template-sync path, clean-install path, OR the user-invocable `runUpdateRestore` lockout-escape path)
**Then** `.moai/cache/template-snapshot/sections/` exists and is non-empty after each.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run 'TestUpdateRestore_WritesSnapshot' -count=1 -v
grep -q '^--- PASS:' <(go test ./internal/cli/update/backup/ -run 'TestUpdateRestore_WritesSnapshot' -count=1 -v 2>&1)
# Dedicated subtests for each of the three restore-completion sites:
go test ./internal/cli/update/backup/ -run 'TestUpdateRestore_WritesSnapshot_TemplateSync' -count=1 -v
go test ./internal/cli/update/backup/ -run 'TestUpdateRestore_WritesSnapshot_CleanInstall' -count=1 -v
go test ./internal/cli/update/backup/ -run 'TestUpdateRestore_WritesSnapshot_RunUpdateRestore' -count=1 -v
```

### AC-TBS-003 — Snapshot carries rendered values, not placeholders

**Given** a deployed project whose `.moai/config/sections/system.yaml` contains `version: "3.0.1"` (rendered)
**When** the snapshot writer runs
**Then** the snapshot's `system.yaml` byte-equals the on-disk rendered file AND contains `version: "3.0.1"`, NOT `version: {{.Version}}`.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestSnapshot_CarriesRenderedValues -count=1 -v
grep -q '^--- PASS: TestSnapshot_CarriesRenderedValues' <(go test ./internal/cli/update/backup/ -run TestSnapshot_CarriesRenderedValues -count=1 -v 2>&1)
# Anti-placeholder grep: the snapshot must not contain unresolved Go-template tokens
! grep -q '{{\.Version}}\|{{\.GoBinPath}}' .moai/cache/template-snapshot/sections/*.yaml 2>/dev/null
```

### AC-TBS-004 — Snapshot survives the update clean step

**Given** a `t.TempDir()` project with `.moai/config/sections/` populated AND `.moai/cache/template-snapshot/sections/` populated
**When** the update clean step routine (the same walk+delete logic used by `update_cleanup.go` / `update_clean_install.go`) runs against the project
**Then** `.moai/cache/template-snapshot/sections/` is preserved (all files still present, byte-identical).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestSnapshot_SurvivesCleanStep -count=1 -v
grep -q '^--- PASS: TestSnapshot_SurvivesCleanStep' <(go test ./internal/cli/update/backup/ -run TestSnapshot_SurvivesCleanStep -count=1 -v 2>&1)
```

### AC-TBS-005 — Snapshot is gitignored

**Given** the repository `.gitignore`
**When** a snapshot is written to `.moai/cache/template-snapshot/`
**Then** `git check-ignore .moai/cache/template-snapshot/sections/system.yaml` exits 0 (the path is ignored).

**Verification command**:
```bash
grep -n '^\.moai/cache/' .gitignore   # existing line — load-bearing for this AC
git check-ignore .moai/cache/template-snapshot/sections/system.yaml && echo PASS || echo FAIL
```
(Note: the gitignore line already exists at `.gitignore:273`; this AC verifies it covers the new subdirectory, not that the line is added by this SPEC.)

### AC-TBS-006a — BackupMoaiConfig prefers snapshot when present (provenance)

**Given** a project with a snapshot containing `version: "3.0.1"` AND a NEW embedded template whose `system.yaml.tmpl` renders to `version: "3.1.0"`
**When** `BackupMoaiConfig(projectRoot)` runs
**Then** the per-backup `.template-defaults/sections/system.yaml` byte-equals the SNAPSHOT bytes (`version: "3.0.1"`), NOT the embedded-raw bytes (`version: "{{.Version}}"`).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestBackupMoaiConfig_PrefersSnapshot -count=1 -v
grep -q '^--- PASS: TestBackupMoaiConfig_PrefersSnapshot' <(go test ./internal/cli/update/backup/ -run TestBackupMoaiConfig_PrefersSnapshot -count=1 -v 2>&1)
```

### AC-TBS-006b — Snapshot bytes, not embedded bytes, reach MergeYAML3Way as baseData

**Given** the same setup as AC-TBS-006a
**When** `SaveTemplateBase(destDir, projectRoot)` is called and the resulting base bytes are fed to `MergeYAML3Way(newData, oldData, baseData)`
**Then** `baseData` is byte-identical to the snapshot's `system.yaml`, confirmed by a test that reads both and compares with `bytes.Equal`.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestSaveTemplateBase_BytesMatchSnapshot -count=1 -v
grep -q '^--- PASS: TestSaveTemplateBase_BytesMatchSnapshot' <(go test ./internal/cli/update/backup/ -run TestSaveTemplateBase_BytesMatchSnapshot -count=1 -v 2>&1)
```

### AC-TBS-007 — Fallback when snapshot absent (backward compat)

**Given** a project with NO `.moai/cache/template-snapshot/` directory
**When** `SaveTemplateBase(destDir, projectRoot)` is called
**Then** it delegates to `SaveTemplateDefaults(destDir)` and the resulting bytes byte-equal what `SaveTemplateDefaults` would produce today.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaults -count=1 -v
grep -q '^--- PASS: TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaults' <(go test ./internal/cli/update/backup/ -run TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaults -count=1 -v 2>&1)
# Existing SaveTemplateDefaults tests remain green (the fallback path is the same code)
go test ./internal/cli/update/backup/ -run 'TestSaveTemplateDefaults' -count=1 -v
```

### AC-TBS-008 — MergeYAML3Way signature unchanged

**Given** the merged source tree
**When** `grep -n 'func MergeYAML3Way' internal/cli/update/backup/merge.go` runs
**Then** the signature line is literally `func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error) {` — byte-identical to the pre-SPEC state.

**Verification command**:
```bash
grep -n 'func MergeYAML3Way' internal/cli/update/backup/merge.go
diff <(grep 'func MergeYAML3Way' internal/cli/update/backup/merge.go) \
     <(printf 'func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error) {\n') \
  && echo PASS || echo FAIL
```

### AC-TBS-009 — 2-way fallback remains available

**Given** an unparseable BASE (e.g. a deliberately corrupted YAML)
**When** `RestoreMoaiConfig` encounters a `MergeYAML3Way` error
**Then** it falls through to `MergeYAMLDeep` at `restore.go:139` and the restore completes.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestRestore_2WayFallback_UnparseableBase -count=1 -v
grep -q '^--- PASS: TestRestore_2WayFallback_UnparseableBase' <(go test ./internal/cli/update/backup/ -run TestRestore_2WayFallback_UnparseableBase -count=1 -v 2>&1)
# The fallback call site must still be present in restore.go:
grep -n 'MergeYAMLDeep' internal/cli/update/backup/restore.go
```

### AC-TBS-010 — MANDATORY falsifiability gate (RED-then-GREEN)

**Given** the implementation is stashed (reverted to the wrong-base behaviour)
**When** the provenance test runs against the stashed tree
**Then** it FAILS (the wrong base misreads the template-blessed key as a user edit, and the assertion `merged output carries version: "3.1.0"` fails because the stale `3.0.1` is preserved instead).

And when the implementation is restored, the same test PASSES.

**Verification procedure** (mirrors prior SPEC AC-UYP-022):
```bash
# 1. Stash implementation + the provenance test (so the test also reverts)
git stash push -u \
  internal/cli/update/backup/snapshot.go \
  internal/cli/update/backup/base_loader.go \
  internal/cli/update/backup/snapshot_provenance_test.go 2>&1

# 2. Intermediate-state assertion: the stash actually reverted the implementation.
#    For NEW untracked files (snapshot.go, base_loader.go, snapshot_provenance_test.go),
#    `git diff --exit-code HEAD -- <file>` returns 0 when the file is absent in BOTH
#    HEAD and the worktree — i.e. exactly when the stash SUCCEEDED — so a diff-based
#    check would print "FAIL" on success. Use the file-existence form instead,
#    matching the sibling base_loader.go and snapshot_provenance_test.go checks.
test ! -f internal/cli/update/backup/snapshot.go \
  && echo "PASS: snapshot.go absent" \
  || echo "FAIL: snapshot.go still present (untracked file not stashed)"
test ! -f internal/cli/update/backup/base_loader.go \
  && echo "PASS: base_loader.go absent" \
  || echo "FAIL: base_loader.go still present (untracked file not stashed)"
test ! -f internal/cli/update/backup/snapshot_provenance_test.go \
  && echo "PASS: provenance test absent" \
  || echo "FAIL: provenance test still present"

# 3. With backup.go's edit ALSO reverted, the wrong-base path (SaveTemplateDefaults) is active.
#    Re-apply ONLY the test (not the implementation) so we can run it against the wrong base:
git checkout stash@'{0}' -- internal/cli/update/backup/snapshot_provenance_test.go

# 4. Run the test against the wrong-base implementation — MUST FAIL
go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1
# Expected exit code: non-zero (FAIL), output contains '--- FAIL: TestMerge_BaseFromSnapshot_NotFromEmbedded'

# 5. Restore the full implementation
git stash pop

# 6. Re-run — MUST PASS
go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1
# Expected exit code: 0, output contains '--- PASS: TestMerge_BaseFromSnapshot_NotFromEmbedded'
```

Both observations (the FAIL output at step 4 and the PASS output at step 6) are recorded verbatim in `progress.md §E.2`. This AC is the single non-vacuity guard for the whole correctness surface — without it, every other correctness AC could be passing against a test that never exercises the wrong-base path.

### AC-TBS-011 — Template-blessed key adopted, not misread (REQ-TBS-010)

**Given** a snapshot whose `version: "3.0.1"` (rendered) matches the LOCAL file's `version: "3.0.1"`, AND a NEW template rendering to `version: "3.1.0"`
**When** the 3-way merge runs with snapshot-sourced BASE
**Then** the merged output carries `version: "3.1.0"` (NEW template value adopted, because `old == base`).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal -count=1 -v
grep -q '^--- PASS: TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal' <(go test ./internal/cli/update/backup/ -run TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal -count=1 -v 2>&1)
```

### AC-TBS-012 — User-customized key preserved (REQ-TBS-011)

**Given** a snapshot whose `test_coverage_target: 80` differs from the LOCAL file's `test_coverage_target: 95` (user customized), AND a NEW template with `test_coverage_target: 85`
**When** the 3-way merge runs with snapshot-sourced BASE
**Then** the merged output carries `test_coverage_target: 95` (user value preserved, because `old != base`).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal -count=1 -v
grep -q '^--- PASS: TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal' <(go test ./internal/cli/update/backup/ -run TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal -count=1 -v 2>&1)
```

### AC-TBS-013 — quality.yaml real-3-way correctness with snapshot (REQ-TBS-012)

**Given** a snapshot whose `quality.yaml` has RENDERED placeholder values (e.g. `enforce_quality: true`, `test_coverage_target: 80`) matching the LOCAL file, AND a NEW `quality.yaml` template with `test_coverage_target: 85`
**When** the 3-way merge runs with snapshot-sourced BASE
**Then** the merged output carries `test_coverage_target: 85` (NEW value adopted) AND `enforce_quality` is NOT misread as a user edit (it equals the snapshot, so `old == base`).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestMerge_QualityYaml_Real3Way_WithSnapshot -count=1 -v
grep -q '^--- PASS: TestMerge_QualityYaml_Real3Way_WithSnapshot' <(go test ./internal/cli/update/backup/ -run TestMerge_QualityYaml_Real3Way_WithSnapshot -count=1 -v 2>&1)
```

### AC-TBS-014 — First-update-after-feature completes cleanly + writes snapshot (REQ-TBS-013)

**Given** a `t.TempDir()` project with NO snapshot (simulating a pre-existing install)
**When** `moai update` runs end-to-end
**Then** (a) the update completes with exit code 0 (no breakage) AND (b) `.moai/cache/template-snapshot/sections/` exists after the update.

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestFirstUpdate_NoSnapshot_CompletesAndWritesSnapshot -count=1 -v
grep -q '^--- PASS: TestFirstUpdate_NoSnapshot_CompletesAndWritesSnapshot' <(go test ./internal/cli/update/backup/ -run TestFirstUpdate_NoSnapshot_CompletesAndWritesSnapshot -count=1 -v 2>&1)
```

### AC-TBS-015 — Snapshot write best-effort non-blocking (REQ-TBS-014)

**Given** a project where the snapshot directory is unwritable (e.g. `.moai/cache/` is a read-only file)
**When** `WriteSnapshot(projectRoot)` is called from within `moai init` / `moai update`
**Then** `WriteSnapshot` returns a non-nil error BUT the enclosing `moai init` / `moai update` still returns nil (the snapshot failure does not propagate).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run TestWriteSnapshot_FailureDoesNotBlockInit -count=1 -v
go test ./internal/cli/update/backup/ -run TestWriteSnapshot_FailureDoesNotBlockUpdate -count=1 -v
grep -q '^--- PASS: TestWriteSnapshot_FailureDoesNotBlockInit' <(go test ./internal/cli/update/backup/ -run TestWriteSnapshot_FailureDoesNotBlockInit -count=1 -v 2>&1)
grep -q '^--- PASS: TestWriteSnapshot_FailureDoesNotBlockUpdate' <(go test ./internal/cli/update/backup/ -run TestWriteSnapshot_FailureDoesNotBlockUpdate -count=1 -v 2>&1)
```

### AC-TBS-016 — Snapshot scope limited to `.moai/config/sections/` (REQ-TBS-015)

**Given** the snapshot writer source
**When** `grep -n 'config/sections\|ConfigSubdir\|SectionsSubdir' internal/cli/update/backup/snapshot.go` runs
**Then** every walk root references `.moai/config/sections/` (no `.claude/`, no `.moai/project/`, no other root).

**Verification command**:
```bash
grep -n 'filepath.Walk\|fs.WalkDir\|ReadDir' internal/cli/update/backup/snapshot.go
# The walk root MUST be filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir) — not broader.
grep -c 'defs.SectionsSubdir\|"config/sections"' internal/cli/update/backup/snapshot.go
# Expected: >= 1
```

### AC-TBS-017 — Cross-platform build (NFR-TBS-001)

**Given** the merged source tree
**When** `GOOS=windows GOARCH=amd64 go build ./...` runs
**Then** exit code is 0.

**Verification command**:
```bash
GOOS=windows GOARCH=amd64 go build ./... && echo PASS || echo FAIL
GOOS=linux GOARCH=amd64 go build ./... && echo PASS || echo FAIL
```

### AC-TBS-018 — Subagent boundary grep (NFR-TBS-005)

**Given** the merged source tree
**When** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/update/backup/ | grep -v _test.go | grep -v '//'` runs
**Then** 0 matches.

**Verification command**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/update/backup/ | grep -v _test.go | grep -v '//' | wc -l
# Expected: 0
```

### AC-TBS-019 — Template neutrality (NFR-TBS-004)

**Given** the merged source tree
**When** `grep -rn 'SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT' internal/template/templates/` runs
**Then** 0 matches.

**Verification command**:
```bash
grep -rn 'SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT' internal/template/templates/ | wc -l
# Expected: 0
```

### AC-TBS-020 — No new dependency (NFR-TBS-003)

**Given** the merged `go.mod`
**When** `git diff main -- go.mod go.sum` runs
**Then** no new `require` line is added (yaml.v3 was already present from the prior SPEC).

**Verification command**:
```bash
git diff main -- go.mod | grep -c '^\+require\|^\+.*yaml'
# Expected: 0 (no new require lines for yaml or anything else)
```

### AC-TBS-021 — Coverage non-regression

**Given** the M0 pre-flight coverage of `internal/cli/update/backup/`
**When** `go test -cover ./internal/cli/update/backup/` runs post-merge
**Then** the coverage percentage is >= the M0 baseline (recorded in `progress.md §E.2`).

**Verification command**:
```bash
go test -cover ./internal/cli/update/backup/ 2>&1 | tee /tmp/tbs-coverage.txt
# Compare against M0 baseline captured in §E.2; coverage MUST NOT decrease.
```

### AC-TBS-022 — Harm-ordering: YAML-preservation non-regression

**Given** the merged source tree
**When** the prior SPEC's golden preservation test suite runs
**Then** every test in `internal/cli/update/backup/preserve_golden_test.go` still passes (no regression on comment/order/quoting preservation introduced by the snapshot wiring).

**Verification command**:
```bash
go test ./internal/cli/update/backup/ -run 'TestPreserve' -count=1 -v 2>&1 | grep -c '^--- PASS:'
# Expected: same count as pre-merge (no PASS becomes FAIL)
go test ./internal/cli/update/backup/ -count=1 2>&1 | tail -5
# Expected: ok ... internal/cli/update/backup ...
```

## §E Edge Cases

- **Concurrent `moai update` invocations on the same project**: out of scope (single-user offline model per `SPEC-SEC-HARDEN-003` §F.1). The snapshot write is not locked; a concurrent update would race on the snapshot directory. This is the same trust model as `.moai/config/` itself.
- **User manually deletes `.moai/cache/`**: the next update falls back per REQ-TBS-007 (no snapshot → embedded-raw BASE → today's behaviour) and re-writes the snapshot at restore end. Self-healing.
- **Snapshot from a much older template version**: exactly the correct BASE — the merge's job is to diff against the user's previous install, however old. If the user skipped N releases, the snapshot captures the N-releases-ago rendered state, and the merge correctly distinguishes "user customized" from "template changed across N releases".
- **Snapshot file present but corrupted (invalid YAML)**: `SaveTemplateBase` copies the bytes verbatim into the per-backup BASE; `RestoreMoaiConfig` then attempts `MergeYAML3Way`, the BASE parse fails, and the 2-way fallback fires (REQ-TBS-009). The corrupted snapshot does not break the update; it merely degrades this one cycle to 2-way behaviour.
- **First-ever `moai init` on a brand-new project**: snapshot is written at init end (REQ-TBS-001). There is no "prior install" to snapshot from, but the freshly-deployed rendered files ARE the correct baseline for the NEXT update.

## §F Definition of Done

- All MUST-PASS ACs green.
- The mandatory falsifiability AC (AC-TBS-010) executed end-to-end with both RED and GREEN observations recorded verbatim in `progress.md §E.2`.
- `go test ./... -count=1` fully green.
- `golangci-lint run ./internal/cli/...` clean.
- `GOOS=windows GOARCH=amd64 go build ./...` clean.
- Coverage of `internal/cli/update/backup/` >= M0 baseline.
- The prior SPEC's preservation suite (`preserve_golden_test.go`) still green (AC-TBS-022).
- `progress.md §E.2` carries the run-phase evidence (commands + verbatim outputs).
- `progress.md §E.3` carries the run-phase audit-ready signal.
- Sync-phase (manager-docs) populates `progress.md §E.4` with `sync_commit_sha`.
