// @TEST:INIT-003:DATA | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:DATA, @SPEC:INIT-003

/**
 * @file Markdown Section Merge Strategy Tests
 * @author MoAI Team
 * @tags @TEST:INIT-003:DATA
 */

import { describe, it, expect } from 'vitest';
import { mergeMarkdown } from '../../../../../src/cli/commands/project/merge-strategies/markdown-merger.js';

describe('Markdown Merger', () => {
  describe('HISTORY section merging', () => {
    it('should merge HISTORY sections from both files', () => {
      const backup = `---
id: TEST-001
version: 0.1.0
---

# Test SPEC

## HISTORY

### v0.1.0 (2025-01-01)
- **INITIAL**: First version
- **AUTHOR**: @User
`;

      const current = `---
id: TEST-001
version: 0.2.0
---

# Test SPEC

## HISTORY

### v0.2.0 (2025-01-02)
- **CHANGED**: Updated implementation
- **AUTHOR**: @System
`;

      const result = mergeMarkdown(backup, current);

      // Should include both HISTORY entries
      expect(result).toContain('### v0.2.0 (2025-01-02)');
      expect(result).toContain('### v0.1.0 (2025-01-01)');

      // Should preserve order (newest first)
      const v02Index = result.indexOf('v0.2.0');
      const v01Index = result.indexOf('v0.1.0');
      expect(v02Index).toBeLessThan(v01Index);
    });

    it('should handle missing HISTORY section in backup', () => {
      const backup = `# Test SPEC

Some content without HISTORY.
`;

      const current = `# Test SPEC

## HISTORY

### v0.2.0 (2025-01-02)
- **CHANGED**: Updated
`;

      const result = mergeMarkdown(backup, current);

      // Should keep current HISTORY
      expect(result).toContain('## HISTORY');
      expect(result).toContain('### v0.2.0');
    });

    it('should handle missing HISTORY section in current', () => {
      const backup = `# Test SPEC

## HISTORY

### v0.1.0 (2025-01-01)
- **INITIAL**: First version
`;

      const current = `# Test SPEC

Some content without HISTORY.
`;

      const result = mergeMarkdown(backup, current);

      // Should add HISTORY from backup
      expect(result).toContain('## HISTORY');
      expect(result).toContain('### v0.1.0');
    });

    it('should handle missing HISTORY in both files', () => {
      const backup = `# Test SPEC

No HISTORY here.
`;

      const current = `# Test SPEC

No HISTORY here either.
`;

      const result = mergeMarkdown(backup, current);

      // Should not create HISTORY section
      expect(result).not.toContain('## HISTORY');
    });
  });

  describe('Front Matter merging', () => {
    it('should preserve backup front matter values', () => {
      const backup = `---
id: TEST-001
version: 0.1.0
author: "@User"
custom: true
---

# Content
`;

      const current = `---
id: TEST-001
version: 0.2.0
author: "@System"
status: active
---

# Content
`;

      const result = mergeMarkdown(backup, current);

      // Backup values should take priority
      expect(result).toContain('version: 0.1.0');
      expect(result).toContain('author: "@User"');
      expect(result).toContain('custom: true');

      // New fields from current should be added
      expect(result).toContain('status: active');
    });

    it('should handle missing front matter in backup', () => {
      const backup = `# Test SPEC

No front matter.
`;

      const current = `---
id: TEST-001
version: 0.2.0
---

# Content
`;

      const result = mergeMarkdown(backup, current);

      // Should keep current front matter
      expect(result).toContain('---');
      expect(result).toContain('id: TEST-001');
      expect(result).toContain('version: 0.2.0');
    });
  });

  describe('Edge cases', () => {
    it('should handle empty backup', () => {
      const backup = '';
      const current = `# Test SPEC

Content here.
`;

      const result = mergeMarkdown(backup, current);

      // Should return current as-is
      expect(result).toBe(current);
    });

    it('should handle empty current', () => {
      const backup = `# Test SPEC

Backup content.
`;
      const current = '';

      const result = mergeMarkdown(backup, current);

      // Should return backup as-is
      expect(result).toBe(backup);
    });

    it('should handle both empty', () => {
      const result = mergeMarkdown('', '');

      expect(result).toBe('');
    });

    it('should preserve other content sections', () => {
      const backup = `---
id: TEST-001
---

# Test SPEC

## Introduction

Backup intro.

## HISTORY

### v0.1.0
- Initial

## Details

Backup details.
`;

      const current = `---
id: TEST-001
---

# Test SPEC

## Introduction

Current intro.

## HISTORY

### v0.2.0
- Updated

## Details

Current details.
`;

      const result = mergeMarkdown(backup, current);

      // Should merge HISTORY
      expect(result).toContain('### v0.2.0');
      expect(result).toContain('### v0.1.0');

      // Other sections should come from current (template priority)
      expect(result).toContain('Current intro');
      expect(result).toContain('Current details');
    });
  });
});
