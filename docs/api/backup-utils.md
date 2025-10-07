# backup-utils.ts - API Reference

**Module**: `moai-adk-ts/src/core/installer/backup-utils.ts`
**Version**: v0.2.1
**SPEC**: SPEC-INIT-003
**TAG**: @CODE:INIT-003:DATA

---

## Overview

Common utilities for backup creation across Phase A (`moai init`) and Phase B (`/alfred:8-project` emergency backup).

This module provides reusable functions for selective backup logic, ensuring data loss prevention when upgrading MoAI-ADK installations.

---

## API Reference

### hasAnyMoAIFiles

```typescript
function hasAnyMoAIFiles(projectPath: string): boolean
```

Check if any MoAI-ADK files exist in a directory (OR condition).

**Parameters**:
- `projectPath: string` - Project directory path

**Returns**:
- `boolean` - True if at least one MoAI-ADK file exists

**Example**:
```typescript
import { hasAnyMoAIFiles } from './backup-utils';

if (hasAnyMoAIFiles('/path/to/project')) {
  console.log('Existing MoAI-ADK installation detected');
}
```

**Details**:
- Checks for `.claude/`, `.moai/`, and `CLAUDE.md`
- Uses OR condition (returns true if any file exists)
- **@CODE:INIT-003:DATA** - v0.2.1 selective backup check

---

### generateBackupDirName

```typescript
function generateBackupDirName(): string
```

Generate backup directory name with timestamp.

**Returns**:
- `string` - Backup directory name (e.g., `.moai-backup-2025-10-07T05-04-03`)

**Example**:
```typescript
import { generateBackupDirName } from './backup-utils';

const backupDir = generateBackupDirName();
// Output: ".moai-backup-2025-10-07T05-04-03"
```

**Details**:
- Uses ISO 8601 timestamp format
- Replaces `:` and `.` with `-` for filesystem compatibility
- **@CODE:INIT-003:DATA**

---

### getBackupTargets

```typescript
function getBackupTargets(projectPath: string): string[]
```

Get list of files that need to be backed up.

**Parameters**:
- `projectPath: string` - Project directory path

**Returns**:
- `string[]` - Array of file/directory paths that should be backed up

**Example**:
```typescript
import { getBackupTargets } from './backup-utils';

const targets = getBackupTargets('/path/to/project');
// Output: ['.claude', '.moai', 'CLAUDE.md']
```

**Details**:
- Only returns existing files/directories
- Checks directories: `.claude`, `.moai`
- Checks files: `CLAUDE.md`
- **@CODE:INIT-003:DATA** - v0.2.1 selective backup

---

### copyDirectoryRecursive

```typescript
async function copyDirectoryRecursive(
  src: string,
  dest: string
): Promise<void>
```

Copy directory recursively (async version for phase-executor).

**Parameters**:
- `src: string` - Source directory
- `dest: string` - Destination directory

**Returns**:
- `Promise<void>` - Resolves when copy is complete

**Example**:
```typescript
import { copyDirectoryRecursive } from './backup-utils';

await copyDirectoryRecursive('.claude', '.moai-backup-2025-10-07/.claude');
```

**Details**:
- Creates destination directory if not exists
- Recursively copies subdirectories
- Preserves file structure
- **@CODE:INIT-003:DATA**

---

### isValidBackupMetadata

```typescript
function isValidBackupMetadata(metadata: unknown): boolean
```

Validate backup metadata structure.

**Parameters**:
- `metadata: unknown` - Backup metadata object

**Returns**:
- `boolean` - True if metadata is valid

**Example**:
```typescript
import { isValidBackupMetadata } from './backup-utils';

const metadata = JSON.parse(fs.readFileSync('.moai/backups/latest.json', 'utf-8'));

if (isValidBackupMetadata(metadata)) {
  console.log('Valid backup metadata');
}
```

**Details**:
- Validates required fields: `timestamp`, `backup_path`, `backed_up_files`, `status`, `created_by`
- Checks `status` enum: `'pending' | 'merged' | 'ignored'`
- Returns `false` for null/undefined or malformed data
- **@CODE:INIT-003:DATA**

**Metadata Structure**:
```typescript
interface BackupMetadata {
  timestamp: string;              // ISO 8601 format
  backup_path: string;            // Backup directory path
  backed_up_files: string[];      // Actually backed up files (v0.2.1)
  status: 'pending' | 'merged' | 'ignored';
  created_by: string;             // 'moai init' or '/alfred:8-project (emergency backup)'
}
```

---

## Usage Scenarios

### Phase A: moai init

```typescript
import {
  hasAnyMoAIFiles,
  generateBackupDirName,
  getBackupTargets,
  copyDirectoryRecursive
} from './backup-utils';

// Check for existing installation
if (!hasAnyMoAIFiles(projectPath)) {
  console.log('New installation - skip backup');
  return;
}

// Generate backup directory
const backupDir = generateBackupDirName();

// Get selective backup targets
const targets = getBackupTargets(projectPath);

// Copy each target
for (const target of targets) {
  await copyDirectoryRecursive(
    path.join(projectPath, target),
    path.join(backupDir, target)
  );
}
```

### Phase B: /alfred:8-project Emergency Backup

```typescript
import {
  hasAnyMoAIFiles,
  generateBackupDirName,
  getBackupTargets,
  copyDirectoryRecursive,
  isValidBackupMetadata
} from './backup-utils';

// Check for existing metadata
const metadataPath = '.moai/backups/latest.json';
if (fs.existsSync(metadataPath)) {
  const metadata = JSON.parse(fs.readFileSync(metadataPath, 'utf-8'));

  if (isValidBackupMetadata(metadata) && metadata.status === 'pending') {
    // Use existing backup
    await handleBackupMerge(metadata);
    return;
  }
}

// Emergency backup if files exist but no metadata
if (hasAnyMoAIFiles(projectPath)) {
  console.log('Creating emergency backup...');

  const backupDir = generateBackupDirName();
  const targets = getBackupTargets(projectPath);

  for (const target of targets) {
    await copyDirectoryRecursive(
      path.join(projectPath, target),
      path.join(backupDir, target)
    );
  }

  // Create metadata
  const metadata = {
    timestamp: new Date().toISOString(),
    backup_path: backupDir,
    backed_up_files: targets,
    status: 'pending',
    created_by: '/alfred:8-project (emergency backup)'
  };

  fs.writeFileSync(metadataPath, JSON.stringify(metadata, null, 2));
}
```

---

## Related Files

- **Implementation**: `moai-adk-ts/src/core/installer/backup-utils.ts`
- **Tests**: `moai-adk-ts/__tests__/core/installer/phase-executor.test.ts` (v0.2.1 scenarios)
- **Phase A Usage**: `moai-adk-ts/src/core/installer/phase-executor.ts`
- **Phase B Usage**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts`
- **SPEC**: `.moai/specs/SPEC-INIT-003/spec.md`

---

## TAG Traceability

- **@SPEC:INIT-003** - Init 백업 및 병합 옵션
- **@CODE:INIT-003:DATA** - 백업 유틸리티 (backup-utils.ts)
- **@CODE:INIT-003:BACKUP** - Phase A 백업 로직 (phase-executor.ts)
- **@CODE:INIT-003:MERGE** - Phase B 병합 로직 (backup-merger.ts)
- **@TEST:INIT-003:BACKUP** - Phase A 테스트

---

**Last Updated**: 2025-10-07 (v0.2.1)
**Author**: MoAI Team
**Related SPEC**: SPEC-INIT-003 v0.2.1
