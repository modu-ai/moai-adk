---
id: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
title: "moai update User-Owned Namespace Protection + Backup Standardization — Progress"
version: "0.1.0"
status: in-progress
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/cli/update.go, internal/cli/update_archive.go, internal/defs/dirs.go"
lifecycle: spec-anchored
tags: "update, namespace, backup, harness, protection, contract, tier-m, progress"
tier: M
---

# Run-Phase Progress — SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001

## M1 — Baseline Measurement (COMPLETE)

**Branch**: main
**Baseline HEAD**: `cf6a1c45f` (note: plan.md cites `bac893173` — plan written at earlier commit; line anchors verified unchanged)

### Baseline LOC (verified by direct read)

| Path | LOC | Anchors |
|------|----:|---------|
| `internal/cli/update.go` | 2723 | `isUserAreaPath`@1113, `isMoaiManaged`@1140 (harness@1178), `backupMoaiConfig`@1335, `cleanMoaiManagedPaths`@1518, `cleanup_old_backups`@1678 |
| `internal/cli/update_archive.go` | 354 | `copyDirAll`@193, `copyFile`@331, `driftStamp` "20060102T150405Z"@251 |
| `internal/cli/update_test.go` | 2757 | `TestIsMoaiManaged`@634, `TestIsMoaiManaged_OutputStyles`@737, `TestIsMoaiManaged_MoaiConfig`@775, `TestCleanMoaiManagedPaths`@2147 |
| `internal/defs/dirs.go` | 108 | `BackupsDir = ".moai-backups"`@12 |

### Baseline Build / Test / Lint

```
$ go build ./...                          → exit 0
$ GOOS=linux GOARCH=amd64 go build ./...  → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ go test ./internal/cli/... -run "TestIsMoaiManaged|TestCleanMoaiManagedPaths|TestBackupMoaiConfig" → PASS
$ golangci-lint run --timeout=2m ./internal/cli/... → 17 pre-existing issues (errcheck 6, ineffassign 1, unused 10)
$ grep -r "Retired\|TestHarnessRetirement\|superseded" internal/cli/ → only plan_audit_d7_d8_test.go (unrelated SPEC)
```

### Decision Lock-In

| Decision | Resolution |
|----------|------------|
| Keep or fold `isUserAreaPath`? | **Keep** (NFR-UNP-005 additivity). `isUserOwnedNamespace` becomes strict superset. |
| Parallel or sequential backup? | **Sequential**. `backupUserOwnedNamespace` called immediately AFTER `backupMoaiConfig` in `Backup` step. |
| New file or append? | **New file** `internal/cli/update_namespace_protect.go` to keep update.go boundary stable. |
| Test file location? | `internal/cli/update_namespace_protect_test.go` |

### Critical Finding Status

`update.go:1178` line `case "core", "expert", "meta", "harness":` confirmed present. M2 will remove "harness" from this switch.

Existing `TestIsMoaiManaged` (line 634) does NOT include a `harness` case → removing `harness` from `isMoaiManaged` will not break existing tests.

## M2 — Namespace Protection Logic (PENDING)

## M3 — Backup Mechanism Standardization (PENDING)

## M4 — Sentinel + Tests (PENDING)

## M5 — Cross-Platform Validation + Status Update (PENDING)
