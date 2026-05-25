---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "Design — v2-to-v3 Clean Reinstall Architecture with v.2.x FLAT layout restoration"
version: "0.5.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, design, architecture, algorithm, layout-restoration"
tier: M
---

# Design — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001

## §A — Architectural Overview

The clean-reinstall architecture is a **single linear pipeline** with explicit step boundaries and a defer-based cleanup contract. It composes existing infrastructure rather than introducing new abstractions.

### §A.1 High-Level Pipeline

```
                       ┌────────────────────────────────────────────┐
                       │ moai update (existing CLI entrypoint)      │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 0: v2-fingerprint detection           │
                       │   detectV2Fingerprint(projectRoot)         │
                       │   Returns: V2Fingerprint{ IsV2, Signals }  │
                       └────────────────┬───────────────────────────┘
                                        │
                          ┌─────────────┴─────────────┐
                          │                           │
                          ▼                           ▼
                  ┌─────────────────┐         ┌──────────────────────┐
                  │ IsV2 = false    │         │ IsV2 = true          │
                  │ → existing file-│         │ → clean-reinstall    │
                  │   level sync    │         │   code path begins   │
                  │   (idempotent)  │         │                      │
                  └─────────────────┘         └──────┬───────────────┘
                                                     │
                                                     ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 1 (conditional): runMigrateAgency     │
                       │   if .agency/ present (REQ-VVCR-025)       │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 2: buildPreserveInventory             │
                       │   - enumerate user-owned paths             │
                       │   - detect user-modified configs (hash)    │
                       │   - return PreserveInventory struct        │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 3: backupAll                          │
                       │   path: .moai/backups/v2-to-v3-{stamp}/    │
                       │   snapshots .moai/ + .claude/ verbatim     │
                       │   (HARD-5 — before any removal)            │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 4: REMOVE phase                       │
                       │   - scanDeprecatedPaths (extended table)   │
                       │   - remove template-managed namespaces     │
                       │   - remove design assets                   │
                       │   - remove TBD + 0-ref yaml + db dir       │
                       │   - remove cache + agency dirs             │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 5: Reinstall phase                    │
                       │   - deploy v3 embedded template baseline   │
                       │   - skip PRESERVE-inventory paths          │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 6: MERGE-back phase                   │
                       │   - restore PRESERVE inventory paths       │
                       │   - restore user-modified configs          │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Step 7: Integrity verification             │
                       │   - assert PRESERVE existence              │
                       │   - assert REMOVE absence                  │
                       │   - assert v3 surface presence             │
                       │   - assert system.yaml version updated     │
                       └────────────────┬───────────────────────────┘
                                        │
                                        ▼
                       ┌────────────────────────────────────────────┐
                       │ Telemetry event emission                   │
                       │   tag: update.clean_reinstall.v2_to_v3     │
                       │   (graceful failure via defer recover)     │
                       └────────────────────────────────────────────┘
```

### §A.2 Component Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                       internal/cli/update.go                          │
│                       (runUpdate entrypoint)                          │
│                                                                       │
│   if IsV2: → runCleanReinstall                                        │
│   else:    → existing file-level sync                                 │
└──────────────┬───────────────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────────────┐
│             internal/cli/v2_detection.go (NEW)                        │
│   detectV2Fingerprint(projectRoot)                                    │
│   readSystemYAMLVersion / probeAgencyDir / probeDeprecatedPath        │
└──────────────────────────────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────────────┐
│         internal/cli/update_clean_install.go (NEW)                    │
│   runCleanReinstall(ctx, projectRoot, dryRun, out)                    │
│   ┌──────────────────────────┐  ┌──────────────────────────┐         │
│   │ Calls:                   │  │ Reuses verbatim:         │         │
│   │  - buildPreserveInventory│  │  - runMigrateAgency      │         │
│   │  - backupAll             │  │  - scanDeprecatedPaths   │         │
│   │  - removePhase           │  │  - backupDeprecatedPaths │         │
│   │  - reinstallPhase        │  │  - newNamespaceBackupStamp│        │
│   │  - mergeBackPhase        │  │  - resolveNamespaceBackupDir│      │
│   │  - integrityVerify       │  │  - collectUserOwnedFiles │         │
│   │  - emitTelemetry         │  │  - isUserOwnedNamespace  │         │
│   └──────────────────────────┘  └──────────────────────────┘         │
└──────────────────────────────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────────────┐
│       internal/cli/update_preserve_inventory.go (NEW)                 │
│   buildPreserveInventory(projectRoot) → PreserveInventory             │
│   detectUserModifiedConfigs(projectRoot) → []string                   │
│   snapshotPreserveInventory(inv, backupDir) → error                   │
└──────────────────────────────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────────────┐
│             internal/defs/dirs.go (MODIFIED)                          │
│   DeprecatedPaths slice extended with 34 NEW entries (§A.4)           │
│   DeprecatedPathEntry struct unchanged                                │
└──────────────────────────────────────────────────────────────────────┘
```

## §B — Core Algorithm Pseudocode

### §B.1 v2-Fingerprint Detection

```go
type V2Fingerprint struct {
    IsV2                          bool
    V2DetectedViaVersion          bool   // moai.version matches Signal 1 of 3 sub-states: (a) "v2.*" prefix, (b) empty string, (c) system.yaml missing (errors.Is(err, fs.ErrNotExist)) — see §B.1 body
    V2DetectedViaAgencyDir        bool   // .agency/ exists
    V2DetectedViaDeprecatedPath   bool   // any DeprecatedPaths entry exists
    SignalDetails                 map[string]string  // per-signal diagnostic info
}

func detectV2Fingerprint(projectRoot string) (V2Fingerprint, error) {
    var fp V2Fingerprint
    fp.SignalDetails = make(map[string]string)

    // Signal 1: read .moai/config/sections/system.yaml version field
    sysYAMLPath := filepath.Join(projectRoot, ".moai", "config", "sections", "system.yaml")
    if version, err := readMoAIVersion(sysYAMLPath); err == nil {
        if strings.HasPrefix(version, "v2.") || version == "" {
            fp.V2DetectedViaVersion = true
            fp.SignalDetails["version_signal"] = "moai.version=" + version
        }
    } else if errors.Is(err, fs.ErrNotExist) {
        // Missing system.yaml is itself a v2 signal (v3 always carries it)
        fp.V2DetectedViaVersion = true
        fp.SignalDetails["version_signal"] = "system.yaml missing"
    }

    // Signal 2: probe .agency/ legacy directory
    agencyPath := filepath.Join(projectRoot, ".agency")
    if _, err := os.Stat(agencyPath); err == nil {
        fp.V2DetectedViaAgencyDir = true
        fp.SignalDetails["agency_signal"] = ".agency/ present"
    }

    // Signal 3: probe any DeprecatedPaths entry
    for _, entry := range defs.DeprecatedPaths {
        p := filepath.Join(projectRoot, filepath.FromSlash(entry.Path))
        if _, err := os.Stat(p); err == nil {
            fp.V2DetectedViaDeprecatedPath = true
            fp.SignalDetails["deprecated_signal_first_hit"] = entry.Path
            break
        }
    }

    fp.IsV2 = fp.V2DetectedViaVersion || fp.V2DetectedViaAgencyDir || fp.V2DetectedViaDeprecatedPath
    return fp, nil
}
```

### §B.2 PRESERVE Inventory Construction

```go
type PreserveInventory struct {
    Paths                []string          // absolute paths
    UserModifiedConfigs  map[string][]byte // section filename → current user content
    TemplateBaselines    map[string][]byte // section filename → template baseline (for conflict report)
    Manifest             map[string]string // path → SHA-256 hash (for post-update integrity check)
}

func buildPreserveInventory(projectRoot string) (PreserveInventory, error) {
    inv := PreserveInventory{
        UserModifiedConfigs: make(map[string][]byte),
        TemplateBaselines:   make(map[string][]byte),
        Manifest:            make(map[string]string),
    }

    // 1. Enumerate .moai/specs/ recursively (HARD-1)
    addRecursive(&inv, projectRoot, ".moai/specs")

    // 2. .moai/project/{product,structure,tech}.md (HARD-2)
    for _, p := range []string{"product.md", "structure.md", "tech.md"} {
        addIfExists(&inv, projectRoot, filepath.Join(".moai/project", p))
    }

    // 3. .moai/harness/, .moai/state/, .moai/backups/, .moai/logs/
    for _, d := range []string{"harness", "state", "backups", "logs"} {
        addRecursive(&inv, projectRoot, filepath.Join(".moai", d))
    }

    // 4. .claude/skills/<dirs not prefixed with moai-> (includes my-harness-* per HARD-3)
    addSkillsExceptMoaiPrefix(&inv, projectRoot)

    // 5. .claude/agents/{harness,local}/ + .claude/agents/<custom>.md at root
    for _, d := range []string{"harness", "local"} {
        addRecursive(&inv, projectRoot, filepath.Join(".claude/agents", d))
    }
    addAgentRootCustomFiles(&inv, projectRoot)  // .md files at .claude/agents/ root that don't match {core,expert,meta}

    // 6. .claude/commands/ root files + non-moai/ subdirs (HARD-4)
    addCommandsRootAndNonMoai(&inv, projectRoot)

    // 7. .claude/rules/ non-moai subdirs
    addRulesNonMoai(&inv, projectRoot)

    // 8. .claude/hooks/ non-moai subdirs
    addHooksNonMoai(&inv, projectRoot)

    // 9. .claude/output-styles/ non-moai subdirs
    addOutputStylesNonMoai(&inv, projectRoot)

    // 10. .claude/settings.local.json + .claude/agent-memory/
    addIfExists(&inv, projectRoot, ".claude/settings.local.json")
    addRecursive(&inv, projectRoot, ".claude/agent-memory")

    // 11. User-modified .moai/config/sections/*.yaml (REQ-VVCR-007)
    modified, err := detectUserModifiedConfigs(projectRoot)
    if err != nil {
        return inv, fmt.Errorf("detect user-modified configs: %w", err)
    }
    for _, filename := range modified {
        sectionPath := filepath.Join(projectRoot, ".moai/config/sections", filename)
        content, err := os.ReadFile(sectionPath)
        if err != nil {
            return inv, fmt.Errorf("read user-modified config %s: %w", filename, err)
        }
        inv.UserModifiedConfigs[filename] = content
        baseline, err := readEmbeddedTemplateBaseline(filepath.Join(".moai/config/sections", filename))
        if err == nil {
            inv.TemplateBaselines[filename] = baseline
        }
        inv.Paths = append(inv.Paths, sectionPath)
    }

    return inv, nil
}
```

### §B.3 User-Modified Config Detection

```go
func detectUserModifiedConfigs(projectRoot string) ([]string, error) {
    var modified []string
    sectionsDir := filepath.Join(projectRoot, ".moai/config/sections")
    entries, err := os.ReadDir(sectionsDir)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            return modified, nil
        }
        return nil, err
    }
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
            continue
        }
        userPath := filepath.Join(sectionsDir, e.Name())
        userContent, err := os.ReadFile(userPath)
        if err != nil {
            continue
        }
        userHash := sha256Sum(userContent)
        baseline, err := readEmbeddedTemplateBaseline(filepath.Join(".moai/config/sections", e.Name()))
        if err != nil {
            // No template baseline → file is user-only → preserve
            modified = append(modified, e.Name())
            continue
        }
        baselineHash := sha256Sum(baseline)
        if userHash != baselineHash {
            modified = append(modified, e.Name())
        }
    }
    return modified, nil
}
```

### §B.4 REMOVE Phase

```go
func removePhase(ctx context.Context, projectRoot string, inv PreserveInventory, dryRun bool) error {
    // 4.1 Remove all DeprecatedPaths entries (extended table from M2)
    scanResult, err := scanDeprecatedPaths(projectRoot)  // reused from update_cleanup.go
    if err != nil {
        return fmt.Errorf("scan deprecated paths: %w", err)
    }
    for _, hit := range scanResult.Hits {
        if dryRun {
            fmt.Fprintf(os.Stdout, "[dry-run] would remove deprecated: %s\n", hit.Path)
            continue
        }
        if err := removeAll(hit.Path); err != nil {
            return fmt.Errorf("remove deprecated %s: %w", hit.Path, err)
        }
    }

    // 4.2 Remove template-managed namespace surfaces (REQ-VVCR-010)
    templateSurfaces := []string{
        ".claude/skills/moai",         // bare moai/ subdir
        ".claude/agents/core",
        ".claude/agents/expert",
        ".claude/agents/meta",
        ".claude/rules/moai",
        ".claude/commands/moai",
        ".claude/hooks/moai",
        ".claude/output-styles/moai",
    }
    for _, s := range templateSurfaces {
        p := filepath.Join(projectRoot, s)
        if dryRun {
            fmt.Fprintf(os.Stdout, "[dry-run] would remove template surface: %s\n", s)
            continue
        }
        if err := removeAll(p); err != nil {
            return fmt.Errorf("remove template surface %s: %w", s, err)
        }
    }

    // 4.3 Remove moai-* prefixed skills (REQ-VVCR-010)
    skillsDir := filepath.Join(projectRoot, ".claude/skills")
    entries, _ := os.ReadDir(skillsDir)
    for _, e := range entries {
        if !e.IsDir() {
            continue
        }
        if strings.HasPrefix(e.Name(), "moai-") {
            target := filepath.Join(skillsDir, e.Name())
            if dryRun {
                fmt.Fprintf(os.Stdout, "[dry-run] would remove moai-* skill: %s\n", target)
                continue
            }
            if err := removeAll(target); err != nil {
                return fmt.Errorf("remove moai-* skill %s: %w", e.Name(), err)
            }
        }
    }

    return nil
}
```

### §B.5 MERGE-Back Phase

```go
func mergeBackPhase(projectRoot string, inv PreserveInventory) error {
    // Restore every PRESERVE inventory path from the in-memory snapshot
    for _, p := range inv.Paths {
        relPath, _ := filepath.Rel(projectRoot, p)
        snapshot, ok := inv.Manifest[relPath]
        if !ok {
            return fmt.Errorf("preserve manifest missing entry: %s", relPath)
        }
        // Snapshot already captures content + permissions; we re-write if path is absent
        if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
            if err := restoreFromBackup(projectRoot, relPath, snapshot); err != nil {
                return fmt.Errorf("restore preserve path %s: %w", relPath, err)
            }
        }
    }

    // Restore user-modified configs (REQ-VVCR-022)
    for filename, userContent := range inv.UserModifiedConfigs {
        targetPath := filepath.Join(projectRoot, ".moai/config/sections", filename)
        if err := os.WriteFile(targetPath, userContent, 0o644); err != nil {
            return fmt.Errorf("restore user-modified config %s: %w", filename, err)
        }
    }

    return nil
}
```

### §B.6 Integrity Verification

```go
type IntegrityViolation struct {
    Condition string  // "preserve_missing", "remove_present", "v3_surface_missing", "version_not_updated"
    Path      string
    Detail    string
}

func integrityVerify(projectRoot string, inv PreserveInventory) ([]IntegrityViolation, error) {
    var violations []IntegrityViolation

    // (a) PRESERVE inventory paths exist
    for _, p := range inv.Paths {
        if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
            violations = append(violations, IntegrityViolation{
                Condition: "preserve_missing", Path: p,
                Detail:    "PRESERVE inventory path absent after merge-back",
            })
        }
    }

    // (b) REMOVE target paths are absent
    for _, entry := range defs.DeprecatedPaths {
        p := filepath.Join(projectRoot, filepath.FromSlash(entry.Path))
        if _, err := os.Stat(p); err == nil {
            violations = append(violations, IntegrityViolation{
                Condition: "remove_present", Path: p,
                Detail:    "DeprecatedPaths entry still exists after REMOVE",
            })
        }
    }

    // (c) v3 template-managed surfaces present (FLAT layout per REQ-VVCR-LR-001)
    for _, s := range []string{
        ".claude/skills/moai",
        ".claude/agents/moai",                    // directory exists
        ".claude/agents/moai/manager-spec.md",    // sample retained agent (FLAT)
        ".claude/agents/moai/builder-harness.md", // sample retained agent (FLAT)
        ".claude/agents/moai/plan-auditor.md",    // sample retained agent (FLAT)
        ".claude/rules/moai",
        ".claude/commands/moai",
        ".claude/hooks/moai",
    } {
        p := filepath.Join(projectRoot, s)
        if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
            violations = append(violations, IntegrityViolation{
                Condition: "v3_surface_missing", Path: p,
                Detail:    "v3 template-managed surface absent after reinstall",
            })
        }
    }

    // (c.1) FLAT layout enforcement: rc1-stage subdirectories MUST be absent
    // (per REQ-VVCR-LR-006 unwanted-behavior + HARD-7 baseline)
    for _, deviant := range []string{
        ".claude/agents/core",
        ".claude/agents/expert",
        ".claude/agents/meta",
        ".claude/agents/moai/core",  // FLAT means no subdirs even under moai/
        ".claude/agents/moai/meta",
        ".claude/agents/moai/expert",
    } {
        p := filepath.Join(projectRoot, deviant)
        if _, err := os.Stat(p); err == nil {
            violations = append(violations, IntegrityViolation{
                Condition: "rc1_split_layout_present", Path: p,
                Detail:    "rc1-stage subdirectory split present after reinstall — violates FLAT v.2.x baseline",
            })
        }
    }

    // (d) version field updated
    sysYAMLPath := filepath.Join(projectRoot, ".moai/config/sections/system.yaml")
    if v, err := readMoAIVersion(sysYAMLPath); err == nil {
        if !strings.HasPrefix(v, "v3.") {
            violations = append(violations, IntegrityViolation{
                Condition: "version_not_updated", Path: sysYAMLPath,
                Detail:    fmt.Sprintf("version field reads %q, expected v3.*", v),
            })
        }
    }

    return violations, nil
}
```

## §C — Data Model

### §C.1 `V2Fingerprint` struct

Defined inline in §B.1 above. Lives in `internal/cli/v2_detection.go`.

### §C.2 `PreserveInventory` struct

Defined inline in §B.2 above. Lives in `internal/cli/update_preserve_inventory.go`.

### §C.3 `IntegrityViolation` struct

Defined inline in §B.6 above. Lives in `internal/cli/update_clean_install.go`.

### §C.4 Extended `DeprecatedPathEntry` table (M2 deliverable)

The 34 NEW entries (31 Category B v.2.x-era + 3 Category C rc1-stage per spec.md §A.4) follow the existing struct schema verbatim — no schema changes required. Sample entries:

```go
{
    Path:            ".agency",
    DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    RemovalSchedule: "v3.0.0",
},
{
    Path:            ".claude/agents/expert/expert-backend.md",
    DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    RemovalSchedule: "v3.0.0",
},
{
    Path:            ".moai/config/sections/design.yaml",
    DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    DeprecatedBy:    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    RemovalSchedule: "v3.0.0",
},
// ... continues for all 34 NEW entries per spec.md §A.4
```

## §D — Sequence Diagram (Scenario A: Full v2 cleanup)

```
User              moai update              v2_detection          clean_install         migrate_agency
 │                     │                        │                      │                      │
 │── invoke ──────────▶│                        │                      │                      │
 │                     │── detectV2─────────────▶                      │                      │
 │                     │◀── IsV2=true ──────────                       │                      │
 │                     │                                                │                      │
 │                     │── runCleanReinstall ──────────────────────────▶                      │
 │                     │                                                │                      │
 │                     │                                                │── .agency/ exists? ─▶│
 │                     │                                                │◀── yes ──────────────│
 │                     │                                                │── runMigrateAgency ─▶│
 │                     │                                                │◀── migrated ─────────│
 │                     │                                                │                      │
 │                     │                                                │── buildPreserveInv ──│
 │                     │                                                │   (HARD-1,2,3,4)     │
 │                     │                                                │                      │
 │                     │                                                │── backupAll ─────────│
 │                     │                                                │   .moai/backups/...  │
 │                     │                                                │                      │
 │                     │                                                │── removePhase ───────│
 │                     │                                                │   - DeprecatedPaths  │
 │                     │                                                │   - template surfs   │
 │                     │                                                │   - design assets    │
 │                     │                                                │                      │
 │                     │                                                │── reinstallPhase ────│
 │                     │                                                │   v3 baseline deploy │
 │                     │                                                │                      │
 │                     │                                                │── mergeBackPhase ────│
 │                     │                                                │   restore PRESERVE   │
 │                     │                                                │                      │
 │                     │                                                │── integrityVerify ───│
 │                     │                                                │   all 4 conditions   │
 │                     │                                                │                      │
 │                     │                                                │── emitTelemetry ─────│
 │                     │                                                │   (defer recover)    │
 │                     │                                                │                      │
 │                     │◀── nil error ──────────────────────────────────                      │
 │◀── done ────────────│                                                                       │
```

## §E — Error Handling and Recovery

### §E.1 Error Boundaries

Each of the 7 canonical steps has an explicit error boundary. If any step returns a non-nil error:

1. The remaining steps are NOT executed
2. A structured error report is emitted to stderr with the failed step name and underlying error
3. The user is pointed to the backup directory `.moai/backups/v2-to-v3-{stamp}/` for recovery (HARD-5 guarantee — backup always exists by step 4)
4. Exit code is non-zero

### §E.2 Telemetry Failure Handling

Telemetry emission is wrapped in `defer recover()`:

```go
func emitTelemetry(event TelemetryEvent) {
    defer func() {
        if r := recover(); r != nil {
            // Log but never propagate
            log.Printf("telemetry emission failed (non-fatal): %v", r)
        }
    }()
    telemetryClient.Send(event)
}
```

### §E.3 SIGTERM Handling

The clean-reinstall code path registers a context-aware signal handler (consistent with existing `internal/cli/migrate_agency.go` pattern). On SIGTERM:

1. The current step completes if already past the no-return point (e.g., REMOVE phase mid-iteration)
2. A structured "interrupted" error is emitted referencing the backup directory
3. Exit code is `130` (standard SIGINT/SIGTERM convention)

## §F — Cross-Platform Considerations

### §F.1 Path Normalization

All paths in `DeprecatedPaths` use forward slashes per `internal/defs/dirs.go` convention. Filesystem operations use `filepath.FromSlash(entry.Path)` before any `os.Stat` / `os.Remove` call.

### §F.2 Windows-Specific Concerns

- `os.RemoveAll` on Windows handles directory removal correctly when no open handles remain.
- The backup directory permission mode `0o600` is mapped to `Read+Write owner` on Windows (no group/other concept).
- The cross-platform CI matrix verifies AC-VVCR-014.

### §F.3 macOS / Linux Case-Sensitivity

`.claude/skills/moai-foundation-core` (lowercase) MUST not be confused with `.claude/skills/Moai-Foundation-Core` (case variant). The fingerprint heuristic uses literal byte-for-byte comparison; case-variant filesystems (default macOS) would treat them as identical but the heuristic still returns correct results because all template paths are lowercase by convention.

## §G — Testing Strategy

### §G.1 Unit Test Coverage

- `internal/defs/dirs_test.go` — `DeprecatedPaths` enumeration (AC-VVCR-005)
- `internal/cli/v2_detection_test.go` — table-driven 8-permutation signal coverage (AC-VVCR-001)
- `internal/cli/update_preserve_inventory_test.go` — PRESERVE inventory enumeration + hash diff detection (AC-VVCR-003, AC-VVCR-004)

### §G.2 Integration Test Coverage

- `internal/cli/update_clean_install_test.go` — 3 canonical scenarios (A: full v2 / B: partial v2 / C: clean v3)
- Per-scenario assertions:
  - PRESERVE inventory hash invariance (HARD-1/2/3/4)
  - REMOVE target absence (AC-VVCR-006/007/008/009)
  - Reinstall surface presence (AC-VVCR-010)
  - MERGE-back restoration (AC-VVCR-011)
  - Integrity verification pass (AC-VVCR-012)

### §G.3 Edge Case Coverage

5 edge cases (EC-1 through EC-5 in `acceptance.md`) require dedicated tests:
- EC-1: Concurrent invocation lock test
- EC-2: Disk-full simulation via fixture limited filesystem
- EC-3: SIGTERM injection mid-REMOVE
- EC-4: Permission-denied PRESERVE path
- EC-5: Missing `system.yaml` v2-signal detection

## §I — Layout Restoration Algorithm (M2a — added v0.3.0)

The M2a milestone is mechanically a 5-step revert of SPEC-V3R6-AGENT-FOLDER-SPLIT-001 commit `1bd083725`, scoped to the 7 retained agents (the 12 archived agents are NOT moved — they stay in `.moai/backups/agent-archive-2026-05-25/`).

### §I.1 Step-by-Step Algorithm

```
Step 1 — Verify v.2.x baseline (read-only)
  $ git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/' | wc -l
    Expected: ~19 files at v.2.x baseline (pre-split)

Step 2 — git mv template tree (7 operations)
  $ git mv internal/template/templates/.claude/agents/core/manager-develop.md \
            internal/template/templates/.claude/agents/moai/manager-develop.md
  $ git mv internal/template/templates/.claude/agents/core/manager-docs.md   \
            internal/template/templates/.claude/agents/moai/manager-docs.md
  $ git mv internal/template/templates/.claude/agents/core/manager-git.md    \
            internal/template/templates/.claude/agents/moai/manager-git.md
  $ git mv internal/template/templates/.claude/agents/core/manager-spec.md   \
            internal/template/templates/.claude/agents/moai/manager-spec.md
  $ git mv internal/template/templates/.claude/agents/meta/builder-harness.md   \
            internal/template/templates/.claude/agents/moai/builder-harness.md
  $ git mv internal/template/templates/.claude/agents/meta/evaluator-active.md  \
            internal/template/templates/.claude/agents/moai/evaluator-active.md
  $ git mv internal/template/templates/.claude/agents/meta/plan-auditor.md      \
            internal/template/templates/.claude/agents/moai/plan-auditor.md

Step 3 — git mv local tree (7 operations)
  Same 7 git mv on .claude/agents/{core,meta}/*.md → .claude/agents/moai/*.md

Step 4 — Remove empty directories (6 operations)
  $ rmdir internal/template/templates/.claude/agents/{core,expert,meta}
  $ rmdir .claude/agents/{core,expert,meta}
    Note: `expert/` is empty from the SPEC-V3R6-AGENT-TEAM-REBUILD-001 M3 archive

Step 5 — Cross-reference grep + sed substitution
  $ grep -rln '\.claude/agents/core/\|\.claude/agents/meta/' \
      .claude/skills/ .claude/rules/ .claude/agents/moai/ \
      CLAUDE.md CLAUDE.local.md \
      .moai/specs/*/spec.md .moai/specs/*/plan.md \
      .moai/specs/*/acceptance.md .moai/specs/*/design.md \
      .moai/specs/*/research.md 2>/dev/null \
    | xargs -I {} sed -i.bak \
        -e 's|\.claude/agents/core/|.claude/agents/moai/|g' \
        -e 's|\.claude/agents/meta/|.claude/agents/moai/|g' \
        -e 's|\.claude/agents/expert/|.claude/agents/moai/|g' \
        {}
  $ find . -name '*.bak' -delete
  # Note: this SPEC's own body and SPEC-V3R6-AGENT-FOLDER-SPLIT-001 are
  # exempted by sed scope (the grep is bounded to other namespaces, the
  # AC-VVCR-LR-004 grep verification permits 3 documented exceptions).
```

### §I.2 Idempotency

Step 2-4 are idempotent: re-running on an already-restored tree fails fast because `git mv` errors on a non-existent source. The orchestrator can detect this and skip the milestone:

```go
func isLayoutAlreadyRestored(repoRoot string) bool {
    coreDir := filepath.Join(repoRoot, "internal/template/templates/.claude/agents/core")
    moaiDir := filepath.Join(repoRoot, "internal/template/templates/.claude/agents/moai")
    coreExists := dirExists(coreDir)
    moaiHasAgents := countMDFiles(moaiDir) >= 7
    return !coreExists && moaiHasAgents
}
```

When `isLayoutAlreadyRestored` returns true, M2a is a no-op (logged as "M2a SKIPPED — layout already restored").

### §I.3 Catalog Regeneration Sequencing

M5 runs `gen-catalog-hashes.go --all` AFTER M2a completes. The sequencing is critical because:

1. `gen-catalog-hashes.go` walks the on-disk filesystem and records (path, SHA-256) tuples in `catalog.yaml`
2. If M5 runs before M2a, the catalog records the rc1-stage `{core,meta}/<agent>.md` paths
3. If M5 runs after M2a, the catalog records the FLAT `moai/<agent>.md` paths
4. The catalog is the authoritative source for `moai update` `scanForDrift` operations, so a stale catalog would cause spurious drift warnings

Solution: plan.md M5 explicitly lists M2a as a `Dependencies:` entry, and the orchestrator MUST verify M2a completion via `isLayoutAlreadyRestored` before invoking `gen-catalog-hashes.go`.

## §H — Cross-References

| Reference | Purpose |
|-----------|---------|
| `spec.md` §A.0 | Baseline principle (v.2.x is canonical) |
| `spec.md` §B | REQ-VVCR enumeration (29 entries + 6 LR entries) |
| `spec.md` §B.10 | Layout Restoration REQ-VVCR-LR-NNN enumeration |
| `acceptance.md` | AC-VVCR enumeration (17 base + 5 LR) + edge cases |
| `plan.md` §F.M2a | Layout Restoration milestone deliverables |
| `research.md` §F.6 | v.2.x baseline archaeology (git ls-tree verification) |
| `internal/cli/migrate_agency.go` | Transaction log + checkpoint pattern (reused for `runMigrateAgency`) |
| `internal/cli/update_cleanup.go` | `scanDeprecatedPaths` + `backupDeprecatedPaths` API |
| `internal/cli/update_namespace_protect.go` | `collectUserOwnedFiles` + namespace stamp API |
| `internal/template/scripts/gen-catalog-hashes.go` | catalog.yaml regeneration (REQ-VVCR-LR-005) |
| `CLAUDE.local.md` §24 | PRESERVE namespace contract |
| Commit `1bd083725` (SPEC-V3R6-AGENT-FOLDER-SPLIT-001) | Source of the layout deviation being reverted |
