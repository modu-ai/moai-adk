---
description: "Manifest and readback contract for sync safety backups"
paths:
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/hooks/moai/verify-sync-backup.sh"
---

# Sync backup integrity

The sync safety backup is valid only when its manifest records every copied
file with a SHA-256 digest and explicitly records every requested path that was
missing. A non-empty directory, file count alone, or successful `cp` is not
proof of integrity.

The backup step must:

1. run `bash .claude/hooks/moai/verify-sync-backup.sh create <new-backup-dir> <project-root> <relative-path>...`
   for the approved path set, using a new `.moai/backups/sync-{timestamp}/` directory;
2. retain the generated `manifest.tsv` with file hashes and `MISSING` rows;
3. run `bash .claude/hooks/moai/verify-sync-backup.sh verify <backup-dir>` and retain its output;
4. read back the approved files (or perform a restore-to-isolated-fixture
   check) before any writer starts.

Any hash mismatch, malformed row, missing manifest, or missing requested input
blocks document synchronization. The report includes the manifest path, file
count, and verification result.
