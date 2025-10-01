// @TEST:REPORTERS-001 | Chain: @SPEC:DOCTOR-001 -> @SPEC:DOCTOR-001 -> @CODE:DOCTOR-001
// Related: @CODE:RESULT-FORMATTER-001, @CODE:SUMMARY-REPORTER-001

/**
 * @file Result formatter and summary reporter tests
 * @author MoAI Team
 */

import { describe, expect, it } from 'vitest';
import type { RequirementCheckResult } from '@/core/system-checker';
import { ResultFormatter } from '../reporters/result-formatter.js';
import { SummaryReporter } from '../reporters/summary-reporter.js';

describe('ResultFormatter', () => {
  let formatter: ResultFormatter;

  beforeEach(() => {
    formatter = new ResultFormatter();
  });

  describe('formatCheckResult', () => {
    it('should format successful check with green checkmark', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'Node.js',
          category: 'runtime',
          minVersion: '18.0.0',
        },
        result: {
          isInstalled: true,
          versionSatisfied: true,
          detectedVersion: '20.10.0',
        },
      };

      const formatted = formatter.formatCheckResult(checkResult);

      expect(formatted).toContain('✅');
      expect(formatted).toContain('Node.js');
      expect(formatted).toContain('20.10.0');
    });

    it('should format version conflict with yellow warning', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'Python',
          category: 'runtime',
          minVersion: '3.10.0',
        },
        result: {
          isInstalled: true,
          versionSatisfied: false,
          detectedVersion: '3.9.0',
        },
      };

      const formatted = formatter.formatCheckResult(checkResult);

      expect(formatted).toContain('⚠️');
      expect(formatted).toContain('Python');
      expect(formatted).toContain('3.9.0');
      expect(formatted).toContain('>= 3.10.0');
    });

    it('should format missing requirement with red X', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'Git',
          category: 'development',
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
          error: 'Not found',
        },
      };

      const formatted = formatter.formatCheckResult(checkResult);

      expect(formatted).toContain('❌');
      expect(formatted).toContain('Git');
      expect(formatted).toContain('Not found');
    });

    it('should handle unknown version gracefully', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'Tool',
          category: 'runtime',
        },
        result: {
          isInstalled: true,
          versionSatisfied: true,
        },
      };

      const formatted = formatter.formatCheckResult(checkResult);

      expect(formatted).toContain('unknown');
    });
  });

  describe('getInstallationSuggestion', () => {
    it('should generate installation command suggestion', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'Git',
          category: 'development',
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
        },
      };

      const suggestion = formatter.getInstallationSuggestion(
        checkResult,
        'brew install git'
      );

      expect(suggestion).toContain('Install');
      expect(suggestion).toContain('Git');
      expect(suggestion).toContain('brew install git');
    });

    it('should show manual installation when no command available', () => {
      const checkResult: RequirementCheckResult = {
        requirement: {
          name: 'CustomTool',
          category: 'optional',
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
        },
      };

      const suggestion = formatter.getInstallationSuggestion(checkResult, null);

      expect(suggestion).toContain('Manual installation required');
      expect(suggestion).toContain('CustomTool');
    });
  });
});

describe('SummaryReporter', () => {
  let reporter: SummaryReporter;

  beforeEach(() => {
    reporter = new SummaryReporter();
  });

  describe('printEnhancedSummary', () => {
    it('should display success message when all checks pass', () => {
      const summary = {
        totalChecks: 10,
        passedChecks: 10,
        failedChecks: 0,
        detectedLanguages: ['TypeScript'],
        runtime: [],
        development: [],
        optional: [],
      };

      // Just ensure it doesn't throw
      expect(() => reporter.printEnhancedSummary(summary)).not.toThrow();
    });

    it('should display warning when some checks fail', () => {
      const summary = {
        totalChecks: 10,
        passedChecks: 7,
        failedChecks: 3,
        detectedLanguages: ['TypeScript', 'Python'],
        runtime: [],
        development: [],
        optional: [],
      };

      expect(() => reporter.printEnhancedSummary(summary)).not.toThrow();
    });

    it('should handle zero checks gracefully', () => {
      const summary = {
        totalChecks: 0,
        passedChecks: 0,
        failedChecks: 0,
        detectedLanguages: [],
        runtime: [],
        development: [],
        optional: [],
      };

      expect(() => reporter.printEnhancedSummary(summary)).not.toThrow();
    });
  });

  describe('printEnhancedResults', () => {
    it('should categorize and display runtime checks', () => {
      const results = {
        runtime: [
          {
            requirement: { name: 'Node.js', category: 'runtime' as const },
            result: { isInstalled: true, versionSatisfied: true },
          },
        ],
        development: [],
        optional: [],
        detectedLanguages: ['TypeScript'],
      };

      expect(() =>
        reporter.printEnhancedResults(results, new ResultFormatter())
      ).not.toThrow();
    });

    it('should display detected languages', () => {
      const results = {
        runtime: [],
        development: [],
        optional: [],
        detectedLanguages: ['TypeScript', 'Python', 'Go'],
      };

      expect(() =>
        reporter.printEnhancedResults(results, new ResultFormatter())
      ).not.toThrow();
    });

    it('should handle empty results gracefully', () => {
      const results = {
        runtime: [],
        development: [],
        optional: [],
        detectedLanguages: [],
      };

      expect(() =>
        reporter.printEnhancedResults(results, new ResultFormatter())
      ).not.toThrow();
    });
  });
});
