// @TEST:INIT-003:UI | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:UI, @SPEC:INIT-003

/**
 * @file Merge Report Generator Tests
 * @author MoAI Team
 * @tags @TEST:INIT-003:UI
 */

import { describe, it, expect } from 'vitest';
import { generateMergeReport } from '../../../../src/cli/commands/project/merge-report.js';
import type { MergeReport } from '../../../../src/cli/commands/project/backup-merger.js';

describe('Merge Report Generator', () => {
  describe('Report formatting', () => {
    it('should generate markdown report from merge data', () => {
      const report: MergeReport = {
        merged: ['config.json', '.claude/settings.json'],
        skipped: ['README.md'],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should include timestamp
      expect(markdown).toContain('2025-10-06');

      // Should include merged files
      expect(markdown).toContain('config.json');
      expect(markdown).toContain('.claude/settings.json');

      // Should include skipped files
      expect(markdown).toContain('README.md');

      // Should include summary counts
      expect(markdown).toContain('2'); // merged count
      expect(markdown).toContain('1'); // skipped count
    });

    it('should handle report with errors', () => {
      const report: MergeReport = {
        merged: ['file1.json'],
        skipped: [],
        errors: [
          { file: 'bad.json', error: 'Invalid JSON' },
          { file: 'bad2.txt', error: 'Permission denied' },
        ],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should include error section
      expect(markdown).toContain('Error');
      expect(markdown).toContain('bad.json');
      expect(markdown).toContain('Invalid JSON');
      expect(markdown).toContain('bad2.txt');
      expect(markdown).toContain('Permission denied');

      // Should include error count
      expect(markdown).toContain('2'); // error count
    });

    it('should handle empty report', () => {
      const report: MergeReport = {
        merged: [],
        skipped: [],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should still generate valid markdown
      expect(markdown).toContain('# MoAI-ADK Init Merge Report');
      expect(markdown).toContain('0'); // counts should be 0
    });
  });

  describe('Markdown structure', () => {
    it('should have proper markdown headings', () => {
      const report: MergeReport = {
        merged: ['file.json'],
        skipped: [],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should have main heading
      expect(markdown).toMatch(/^# MoAI-ADK Init Merge Report/m);

      // Should have section headings
      expect(markdown).toContain('## Summary');
      expect(markdown).toContain('## Merged Files');
    });

    it('should format file lists as markdown lists', () => {
      const report: MergeReport = {
        merged: ['file1.json', 'file2.md'],
        skipped: ['file3.txt'],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should use markdown list syntax
      expect(markdown).toMatch(/- .*file1\.json/);
      expect(markdown).toMatch(/- .*file2\.md/);
      expect(markdown).toMatch(/- .*file3\.txt/);
    });

    it('should include metadata section', () => {
      const report: MergeReport = {
        merged: [],
        skipped: [],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should include execution time
      expect(markdown).toContain('2025-10-06');
      expect(markdown).toContain('14:30');
    });
  });

  describe('Summary statistics', () => {
    it('should calculate total files processed', () => {
      const report: MergeReport = {
        merged: ['a.json', 'b.json', 'c.json'],
        skipped: ['d.txt', 'e.txt'],
        errors: [{ file: 'f.json', error: 'Error' }],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Total = merged + skipped + errors = 6
      expect(markdown).toContain('6');

      // Individual counts
      expect(markdown).toContain('3'); // merged
      expect(markdown).toContain('2'); // skipped
      expect(markdown).toContain('1'); // errors
    });

    it('should show success rate percentage', () => {
      const report: MergeReport = {
        merged: ['a.json', 'b.json', 'c.json'], // 3 success
        skipped: ['d.txt'], // 1 skipped (neutral)
        errors: [{ file: 'e.json', error: 'Error' }], // 1 error
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Success rate = merged / (merged + errors) = 3 / 4 = 75%
      expect(markdown).toMatch(/75%|75\.0%/);
    });

    it('should handle 100% success rate', () => {
      const report: MergeReport = {
        merged: ['a.json', 'b.json'],
        skipped: ['c.txt'],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      expect(markdown).toContain('100%');
    });

    it('should handle 0% success rate', () => {
      const report: MergeReport = {
        merged: [],
        skipped: [],
        errors: [
          { file: 'a.json', error: 'Error' },
          { file: 'b.json', error: 'Error' },
        ],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      expect(markdown).toContain('0%');
    });
  });

  describe('File categorization', () => {
    it('should group files by directory', () => {
      const report: MergeReport = {
        merged: ['.claude/settings.json', '.claude/hooks/test.cjs', '.moai/config.json'],
        skipped: [],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should show files organized by directory
      expect(markdown).toContain('.claude');
      expect(markdown).toContain('.moai');
    });

    it('should indicate merge strategy used', () => {
      const report: MergeReport = {
        merged: ['config.json', 'README.md', 'hook.cjs'],
        skipped: [],
        errors: [],
        timestamp: '2025-10-06T14:30:00.000Z',
      };

      const markdown = generateMergeReport(report);

      // Should indicate strategy for each file type
      // JSON → Deep merge
      expect(markdown).toMatch(/config\.json.*merge|merge.*config\.json/i);

      // Markdown → HISTORY accumulation
      expect(markdown).toMatch(/README\.md.*history|history.*README\.md/i);

      // Hooks → Version comparison
      expect(markdown).toMatch(/hook\.cjs.*version|version.*hook\.cjs/i);
    });
  });
});
