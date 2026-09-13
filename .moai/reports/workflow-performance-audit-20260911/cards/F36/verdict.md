# F36 — Sync backup integrity

## Claim

Sync backups now use a SHA-256 manifest with explicit missing-path rows and a
readback verifier; a non-empty directory cannot satisfy the integrity gate.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-sync-backup-integrity.sh
```

Observed output:

```text
PASS: backup manifest created (.../manifest.tsv)
PASS: backup manifest verified (2 files)
PASS: sync backup uses a hash manifest and rejects tampering
```

## Baseline-attribution

The fixture test ran in `WT-workflow-audit-f36b` after fast-forwarding to the
F35 merge on local `develop`.

## Gaps

No live sync backup of the repository's actual docs/SPEC tree was created, and
no isolated restore of every supported path type was run.

## Residual-risk

The workflow caller must pass the approved path set and honor a `MISSING` row as
a block. The verifier proves copied-file integrity, not that the caller chose
the correct source set.
