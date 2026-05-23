---
title: Migration Guide
description: Convert existing .agency/ projects to new design system
weight: 60
draft: false
---

# Migration Guide

Per SPEC-AGENCY-ABSORB-001, the `/agency` command has merged into `/moai design`. Projects with existing `.agency/` directories can **migrate** to the new system.

## When to Migrate

**Migrate if any of these apply:**

- `.agency/` directory exists
- Using existing agency learnings/observations
- Using old agents (agency-copywriter, agency-designer, etc.)

**After Migration:**

- `.agency/` → `.agency.archived/` (backup)
- `.moai/project/brand/` (newly created)
- `.moai/config/sections/design.yaml` (newly created)
- Previous learnings mergeable

## Run Migration

### Step 1: Pre-flight Check

```bash
# Verify .agency/ exists
ls -la .agency/
# brand-voice.md
# visual-identity.md
# learnings/
# observations/
```

### Step 2: Dry Run (Optional)

Preview migration results:

```bash
moai migrate agency --dry-run
```

**Output Example:**
```
[Preview] Migrate .agency/ to .moai/project/brand/
3 files to transfer:
  ✓ brand-voice.md
  ✓ visual-identity.md
  ✓ target-audience.md
Config merge:
  ✓ 5 learnings collected
  ✓ 12 observations collected
Backup:
  ✓ Will save to .agency.archived/
```

### Step 3: Execute Migration

```bash
moai migrate agency
```

**Execution (6 Phases):**

1. **Validate** — Check `.agency/` exists, disk space
2. **Staging** — Copy to temp directory
3. **Context Transfer** — Copy brand files to `.moai/project/brand/`
4. **Config Merge** — Merge learnings/observations to `.moai/config/`
5. **Learning Transfer** — Convert heuristics to new structure
6. **Atomic Swap** — Back up to `.agency.archived/`, complete

**On Success:**
```
Migration complete [TX-abc123def456]

Files transferred: 47
  ✓ .moai/project/brand/ 3 files
  ✓ .moai/config/sections/design.yaml created
  ✓ .moai/research/ config merged

Backup: .agency.archived/

Next:
  /moai design
```

## Migration Options

### --force Option

Overwrite existing target directories:

```bash
moai migrate agency --force
```

**Warning:** Overwrites `.moai/project/brand/` if exists. Back up first.

### --resume Option

Resume interrupted migration:

```bash
# After SIGINT stopped migration
moai migrate agency --resume TX-abc123def456
```

Checkpoint file: `~/.moai/.migrate-tx-<txID>.json`

## Error Codes

Migration failures:

| Error Code | Cause | Solution |
|---|---|---|
| `MIGRATE_NO_SOURCE` | `.agency/` missing | Check for existing agency directory |
| `MIGRATE_TARGET_EXISTS` | `.moai/project/brand/` exists | Use `--force` |
| `MIGRATE_ARCHIVE_EXISTS` | `.agency.archived/` exists | Delete or move old backup |
| `MIGRATE_DISK_FULL` | Insufficient disk space | Free space (min 100MB) |
| `MIGRATE_MERGE_CONFLICT` | tech-preferences.md conflict | Back up `.moai/project/tech.md`, retry |
| `MIGRATE_INTERRUPT` | SIGINT/SIGTERM received | Use `--resume` to continue |
| `MIGRATE_CHECKPOINT_CORRUPT` | Checkpoint file damaged | Delete `~/.moai/.migrate-tx-*.json`, retry |

## Migration Results

### Generated File Structure

```
.moai/
├── project/
│   └── brand/
│       ├── brand-voice.md        (from .agency/)
│       ├── visual-identity.md    (from .agency/)
│       └── target-audience.md    (from .agency/)
├── config/
│   └── sections/
│       └── design.yaml           (newly created)
└── research/
    ├── learnings/                (merged)
    └── observations/             (merged)

.agency.archived/                  (original backup)
├── brand-voice.md
├── visual-identity.md
├── learnings/
└── observations/
```

### Learning Merge

Existing `.agency/learnings/` entries:
- Converted to new structure
- Merged into `.moai/research/learnings/`
- Tagged with MIGRATED

**Conversion Example:**

```yaml
# Original
id: LEARN-20260401-001
category: copy
observation: "Hero paragraph limited to 15 words"
confidence: 0.85

# After migration
id: LEARN-20260420-001-MIGRATED-FROM-20260401-001
status: graduated
category: copy
observation: "Hero paragraph limited to 15 words"
confidence: 0.85
migrated_from_agency: true
```

## Rollback

To restore pre-migration state:

### Option 1: Restore from Backup

```bash
# Restore from backup
mv .agency.archived .agency

# Remove newly created files
rm -rf .moai/project/brand
rm .moai/config/sections/design.yaml
```

### Option 2: Git Revert

If migration created a commit:

```bash
git log --oneline | grep migrate
# abc1234 chore: migrate agency to moai design system

git revert abc1234
```

## Post-Migration Steps

After successful migration:

1. **Verify Brand Context**
   ```bash
   cat .moai/project/brand/brand-voice.md
   cat .moai/project/brand/visual-identity.md
   ```

2. **Start New Design Workflow**
   ```
   /moai design
   ```

3. **Review Existing Learnings**
   ```bash
   ls .moai/research/learnings/
   ```

4. **Optional: Delete Old Backup**
   ```bash
   rm -rf .agency.archived
   ```

## Check Migration Status

Query migration results:

```bash
# View migration log
cat ~/.moai/.migrate-tx-abc123def456.json

# Or check status
moai status design
# Design System Status: MIGRATED (2026-04-20)
# Brand Files: 3/3 ✓
# Design Config: ✓
```

## SIGINT/SIGTERM Handling

If migration interrupted:

**On Ctrl+C:**
```
Migration interrupted [TX-abc123def456]
Completed phases: validation, staging, context-transfer
Incomplete phases: config-merge, learning-transfer, atomic-swap

To resume:
  moai migrate agency --resume TX-abc123def456
```

**Checkpoint file:** `~/.moai/.migrate-tx-abc123def456.json`

Resumes from incomplete phase.

## FAQ

### Q: Can I use /agency command after migration?

**A:** No. `/agency` no longer supported. Use `/moai design` instead.

### Q: Can I run migration multiple times?

**A:** First completion moves `.agency/` to `.agency.archived/`. Second run fails with source-not-found. Use `--force` to override.

### Q: Are learnings lost?

**A:** No. All learnings/observations merged into `.moai/research/`. Backup also preserved in `.agency.archived/`.

### Q: What if network drops during migration?

**A:** Completed phases saved. Resume with `--resume` after network restored.
