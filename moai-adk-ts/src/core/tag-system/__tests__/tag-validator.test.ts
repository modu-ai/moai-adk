/**
 * @TEST:TAG-VALIDATOR-001 16-Core TAG Validator Tests
 *
 * Python tag_system/validator.py의 완전 포팅 테스트
 * @FEATURE:TAG-VALIDATOR-001 Primary Chain 검증 및 무결성 검사
 */

import { TagValidator } from '../tag-validator';
import type { TagMatch } from '../tag-parser';

describe('TagValidator - 16-Core TAG Chain Validation', () => {
  let tagValidator: TagValidator;

  beforeEach(() => {
    tagValidator = new TagValidator();
  });

  describe('@TEST:PRIMARY-CHAIN-VALIDATION-001 Primary Chain 검증', () => {
    it('should validate complete PRIMARY chain', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'USER-001',
          description: 'User requirement',
          references: [],
        },
        {
          category: 'DESIGN',
          identifier: 'AUTH-001',
          description: 'Auth design',
          references: [],
        },
        {
          category: 'TASK',
          identifier: 'IMPL-001',
          description: 'Implementation task',
          references: [],
        },
        {
          category: 'TEST',
          identifier: 'VERIFY-001',
          description: 'Test verification',
          references: [],
        },
      ];

      const result = tagValidator.validatePrimaryChain(tags);

      expect(result.isValid).toBe(true);
      expect(result.completenessScore).toBe(1.0);
      expect(result.missingLinks).toHaveLength(0);
      expect(result.chainType).toBe('PRIMARY');
    });

    it('should detect missing links in PRIMARY chain', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'USER-001',
          description: 'User requirement',
          references: [],
        },
        {
          category: 'TEST',
          identifier: 'VERIFY-001',
          description: 'Test verification',
          references: [],
        },
      ];

      const result = tagValidator.validatePrimaryChain(tags);

      expect(result.isValid).toBe(false);
      expect(result.completenessScore).toBe(0.5); // 2 out of 4
      expect(result.missingLinks).toEqual(['DESIGN', 'TASK']);
      expect(result.chainType).toBe('PRIMARY');
    });

    it('should handle empty tag list', () => {
      const result = tagValidator.validatePrimaryChain([]);

      expect(result.isValid).toBe(false);
      expect(result.completenessScore).toBe(0.0);
      expect(result.missingLinks).toEqual(['REQ', 'DESIGN', 'TASK', 'TEST']);
    });
  });

  describe('@TEST:CIRCULAR-REFERENCE-DETECTION-001 순환 참조 검출', () => {
    it('should detect circular references', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'A-001',
          description: '',
          references: ['DESIGN:B-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'B-001',
          description: '',
          references: ['TASK:C-001'],
        },
        {
          category: 'TASK',
          identifier: 'C-001',
          description: '',
          references: ['REQ:A-001'],
        }, // Circular!
      ];

      const circularRefs = tagValidator.detectCircularReferences(tags);

      expect(circularRefs).toHaveLength(1);
      expect(circularRefs[0]).toHaveLength(3);
    });

    it('should return empty array when no circular references exist', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'A-001',
          description: '',
          references: ['DESIGN:B-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'B-001',
          description: '',
          references: ['TASK:C-001'],
        },
        {
          category: 'TASK',
          identifier: 'C-001',
          description: '',
          references: [],
        },
      ];

      const circularRefs = tagValidator.detectCircularReferences(tags);

      expect(circularRefs).toHaveLength(0);
    });
  });

  describe('@TEST:ORPHANED-TAG-DETECTION-001 고아 TAG 검출', () => {
    it('should find orphaned tags', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'CONNECTED-001',
          description: '',
          references: ['DESIGN:AUTH-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'AUTH-001',
          description: '',
          references: [],
        },
        {
          category: 'TASK',
          identifier: 'ORPHAN-001',
          description: '',
          references: [],
        }, // Orphan: no refs, not referenced
        {
          category: 'TEST',
          identifier: 'ORPHAN-002',
          description: '',
          references: [],
        }, // Orphan: no refs, not referenced
      ];

      const orphans = tagValidator.findOrphanedTags(tags);

      expect(orphans).toHaveLength(2);
      expect(orphans.map(tag => tag.identifier)).toEqual([
        'ORPHAN-001',
        'ORPHAN-002',
      ]);
    });

    it('should return empty array when no orphaned tags exist', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'A-001',
          description: '',
          references: ['DESIGN:B-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'B-001',
          description: '',
          references: [],
        },
      ];

      const orphans = tagValidator.findOrphanedTags(tags);

      expect(orphans).toHaveLength(0);
    });
  });

  describe('@TEST:NAMING-CONSISTENCY-001 명명 일관성 검사', () => {
    it('should validate consistent naming patterns', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'USER-LOGIN-001',
          description: '',
          references: [],
        },
        {
          category: 'DESIGN',
          identifier: 'AUTH-SYSTEM-001',
          description: '',
          references: [],
        },
        {
          category: 'TASK',
          identifier: 'IMPL-001',
          description: '',
          references: [],
        },
      ];

      const violations = tagValidator.checkNamingConsistency(tags);

      expect(violations).toHaveLength(0);
    });

    it('should detect naming inconsistencies', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'user-login-001',
          description: '',
          references: [],
        }, // lowercase
        {
          category: 'DESIGN',
          identifier: 'Auth_System_001',
          description: '',
          references: [],
        }, // underscores
        {
          category: 'TASK',
          identifier: 'VALID-001',
          description: '',
          references: [],
        }, // valid
      ];

      const violations = tagValidator.checkNamingConsistency(tags);

      expect(violations).toHaveLength(2);
      expect(violations[0]!.identifier).toBe('user-login-001');
      expect(violations[0]!.issueType).toBe('naming_inconsistency');
      expect(violations[1]!.identifier).toBe('Auth_System_001');
    });
  });

  describe('@TEST:TAG-COVERAGE-001 TAG 커버리지 계산', () => {
    it('should calculate coverage for all 16-Core categories', () => {
      const tags: TagMatch[] = [
        // PRIMARY: 2/4 = 0.5
        {
          category: 'REQ',
          identifier: 'R-001',
          description: '',
          references: [],
        },
        {
          category: 'DESIGN',
          identifier: 'D-001',
          description: '',
          references: [],
        },

        // STEERING: 1/4 = 0.25
        {
          category: 'VISION',
          identifier: 'V-001',
          description: '',
          references: [],
        },

        // IMPLEMENTATION: 3/4 = 0.75
        {
          category: 'FEATURE',
          identifier: 'F-001',
          description: '',
          references: [],
        },
        {
          category: 'API',
          identifier: 'A-001',
          description: '',
          references: [],
        },
        {
          category: 'UI',
          identifier: 'U-001',
          description: '',
          references: [],
        },

        // QUALITY: 4/4 = 1.0
        {
          category: 'PERF',
          identifier: 'P-001',
          description: '',
          references: [],
        },
        {
          category: 'SEC',
          identifier: 'S-001',
          description: '',
          references: [],
        },
        {
          category: 'DOCS',
          identifier: 'DOC-001',
          description: '',
          references: [],
        },
        {
          category: 'TAG',
          identifier: 'T-001',
          description: '',
          references: [],
        },
      ];

      const coverage = tagValidator.calculateTagCoverage(tags);

      expect(coverage['PRIMARY']).toBe(0.5);
      expect(coverage['STEERING']).toBe(0.25);
      expect(coverage['IMPLEMENTATION']).toBe(0.75);
      expect(coverage['QUALITY']).toBe(1.0);
    });

    it('should return zero coverage for empty tag list', () => {
      const coverage = tagValidator.calculateTagCoverage([]);

      expect(coverage['PRIMARY']).toBe(0.0);
      expect(coverage['STEERING']).toBe(0.0);
      expect(coverage['IMPLEMENTATION']).toBe(0.0);
      expect(coverage['QUALITY']).toBe(0.0);
    });
  });

  describe('@TEST:REFERENCE-INTEGRITY-001 참조 무결성 검사', () => {
    it('should validate reference integrity', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'A-001',
          description: '',
          references: ['DESIGN:B-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'B-001',
          description: '',
          references: [],
        },
      ];

      const brokenRefs = tagValidator.validateReferenceIntegrity(tags);

      expect(brokenRefs).toHaveLength(0);
    });

    it('should detect broken references', () => {
      const tags: TagMatch[] = [
        {
          category: 'REQ',
          identifier: 'A-001',
          description: '',
          references: ['DESIGN:MISSING-001', 'TASK:ALSO-MISSING-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'B-001',
          description: '',
          references: ['NONEXISTENT:TAG-001'],
        },
      ];

      const brokenRefs = tagValidator.validateReferenceIntegrity(tags);

      expect(brokenRefs).toHaveLength(3);
      expect(brokenRefs[0]!.sourceIdentifier).toBe('A-001');
      expect(brokenRefs[0]!.brokenReference).toBe('DESIGN:MISSING-001');
      expect(brokenRefs[0]!.reason).toBe('Referenced tag does not exist');
    });
  });

  describe('@TEST:INTEGRATION-VALIDATION-001 통합 검증 시나리오', () => {
    it('should handle complex validation scenario', () => {
      const tags: TagMatch[] = [
        // Complete PRIMARY chain
        {
          category: 'REQ',
          identifier: 'USER-AUTH-001',
          description: 'User auth requirement',
          references: ['DESIGN:AUTH-FLOW-001'],
        },
        {
          category: 'DESIGN',
          identifier: 'AUTH-FLOW-001',
          description: 'Auth flow design',
          references: ['TASK:IMPL-AUTH-001'],
        },
        {
          category: 'TASK',
          identifier: 'IMPL-AUTH-001',
          description: 'Implement auth',
          references: ['TEST:AUTH-TEST-001'],
        },
        {
          category: 'TEST',
          identifier: 'AUTH-TEST-001',
          description: 'Test auth',
          references: [],
        },

        // IMPLEMENTATION tags (referenced by primary chain to avoid being orphans)
        {
          category: 'FEATURE',
          identifier: 'AUTH-FEATURE-001',
          description: '',
          references: [],
        },
        {
          category: 'API',
          identifier: 'AUTH-API-001',
          description: '',
          references: [],
        },

        // Referenced IMPLEMENTATION tags to avoid orphan status
        {
          category: 'UI',
          identifier: 'AUTH-UI-001',
          description: '',
          references: ['FEATURE:AUTH-FEATURE-001', 'API:AUTH-API-001'],
        },

        // Orphaned tag - truly orphaned (no refs, not referenced)
        {
          category: 'DOCS',
          identifier: 'ORPHAN-DOC-001',
          description: '',
          references: [],
        },

        // Broken reference
        {
          category: 'PERF',
          identifier: 'AUTH-PERF-001',
          description: '',
          references: ['TASK:MISSING-TASK-001'],
        },
      ];

      // Test primary chain validation
      const primaryValidation = tagValidator.validatePrimaryChain(tags);
      expect(primaryValidation.isValid).toBe(true);

      // Test circular reference detection
      const circularRefs = tagValidator.detectCircularReferences(tags);
      expect(circularRefs).toHaveLength(0);

      // Test orphaned tag detection
      const orphans = tagValidator.findOrphanedTags(tags);
      expect(orphans).toHaveLength(1);
      expect(orphans[0]!.identifier).toBe('ORPHAN-DOC-001');

      // Test reference integrity
      const brokenRefs = tagValidator.validateReferenceIntegrity(tags);
      expect(brokenRefs).toHaveLength(1);
      expect(brokenRefs[0]!.brokenReference).toBe('TASK:MISSING-TASK-001');

      // Test coverage calculation
      const coverage = tagValidator.calculateTagCoverage(tags);
      expect(coverage['PRIMARY']).toBe(1.0); // All 4 categories present
      expect(coverage['IMPLEMENTATION']).toBe(0.75); // 3 out of 4 categories (FEATURE, API, UI)
    });
  });
});
