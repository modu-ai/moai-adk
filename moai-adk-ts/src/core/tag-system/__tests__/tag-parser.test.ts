/**
 * @TEST:TAG-PARSER-001  TAG Parser Tests
 *
 * Python tag_system/parser.py의 완전 포팅 테스트
 * @FEATURE:TAG-SYSTEM-001  TAG 추적성 시스템
 */

import { TagParser, TagCategory } from '../tag-parser';

describe('TagParser -  TAG System', () => {
  let tagParser: TagParser;

  beforeEach(() => {
    tagParser = new TagParser();
  });

  describe('@TEST:-CATEGORIES-001  TAG 카테고리 체계', () => {
    it('should have correct  TAG categories', () => {
      const categories = tagParser.getTagCategories();

      // PRIMARY Chain: REQ → DESIGN → TASK → TEST
      expect(categories[TagCategory.PRIMARY]).toEqual([
        'REQ',
        'DESIGN',
        'TASK',
        'TEST',
      ]);

      // STEERING: Project direction
      expect(categories[TagCategory.STEERING]).toEqual([
        'VISION',
        'STRUCT',
        'TECH',
        'ADR',
      ]);

      // IMPLEMENTATION: Development
      expect(categories[TagCategory.IMPLEMENTATION]).toEqual([
        'FEATURE',
        'API',
        'UI',
        'DATA',
      ]);

      // QUALITY: Quality assurance
      expect(categories[TagCategory.QUALITY]).toEqual([
        'PERF',
        'SEC',
        'DOCS',
        'TAG',
      ]);

      // Total 16 categories
      const totalCategories = Object.values(categories).flat();
      expect(totalCategories).toHaveLength(16);
    });
  });

  describe('@TEST:TAG-EXTRACTION-001 TAG 추출 기능', () => {
    it('should extract valid  TAGs', () => {
      const content = `
        @REQ:USER-LOGIN-001 사용자 로그인 요구사항
        @DESIGN:AUTH-001 인증 설계
        @TASK:LOGIN-IMPL-001 로그인 구현 작업
        @TEST:LOGIN-001 로그인 테스트
        @FEATURE:AUTH-SYSTEM-001 인증 시스템 구현
        @PERF:LOGIN-SPEED-001 로그인 성능 최적화
      `;

      const tags = tagParser.extractTags(content);

      expect(tags).toHaveLength(6);
      expect(tags[0]).toEqual({
        category: 'REQ',
        identifier: 'USER-LOGIN-001',
        description: '사용자 로그인 요구사항',
        references: [],
      });

      expect(tags[1]).toEqual({
        category: 'DESIGN',
        identifier: 'AUTH-001',
        description: '인증 설계',
        references: [],
      });
    });

    it('should ignore invalid TAG categories', () => {
      const content = `
        @REQ:VALID-001 Valid tag
        @INVALID:BAD-001 Invalid tag
        @FEATURE:GOOD-001 Another valid tag
        @WRONG:CATEGORY-001 Wrong category
      `;

      const tags = tagParser.extractTags(content);

      expect(tags).toHaveLength(2);
      expect(tags[0]!.category).toBe('REQ');
      expect(tags[1]!.category).toBe('FEATURE');
    });

    it('should extract tags without descriptions', () => {
      const content = '@REQ:SIMPLE-001\n@TASK:BASIC-001';

      const tags = tagParser.extractTags(content);

      expect(tags).toHaveLength(2);
      expect(tags[0]!.description).toBeNull();
      expect(tags[1]!.description).toBeNull();
    });
  });

  describe('@TEST:TAG-CHAIN-PARSING-001 TAG 체인 파싱', () => {
    it('should parse PRIMARY chain correctly', () => {
      const content =
        '@REQ:LOGIN-001 → @DESIGN:AUTH-001 → @TASK:IMPL-001 → @TEST:VERIFY-001';

      const chains = tagParser.parseTagChains(content);

      expect(chains).toHaveLength(1);
      expect(chains[0]!.links).toHaveLength(4);
      expect(chains[0]!.links[0]!.category).toBe('REQ');
      expect(chains[0]!.links[1]!.category).toBe('DESIGN');
      expect(chains[0]!.links[2]!.category).toBe('TASK');
      expect(chains[0]!.links[3]!.category).toBe('TEST');
    });

    it('should parse multiple chains in content', () => {
      const content = `
        Primary: @REQ:A-001 → @DESIGN:A-001 → @TASK:A-001
        Feature: @FEATURE:B-001 → @API:B-001 → @UI:B-001
      `;

      const chains = tagParser.parseTagChains(content);

      expect(chains).toHaveLength(2);
      expect(chains[0]!.links).toHaveLength(3);
      expect(chains[1]!.links).toHaveLength(3);
    });

    it('should ignore chains with invalid categories', () => {
      const content = '@REQ:VALID-001 → @INVALID:BAD-001 → @TASK:VALID-001';

      const chains = tagParser.parseTagChains(content);

      expect(chains).toHaveLength(1);
      expect(chains[0]!.links).toHaveLength(2); // Only REQ and TASK
      expect(chains[0]!.links[0]!.category).toBe('REQ');
      expect(chains[0]!.links[1]!.category).toBe('TASK');
    });
  });

  describe('@TEST:TAG-FORMAT-VALIDATION-001 TAG 형식 검증', () => {
    it('should validate correct TAG formats', () => {
      const validTags = [
        '@REQ:USER-001',
        '@DESIGN:AUTH-SYSTEM-001',
        '@TASK:IMPL-LOGIN-001 Implementation task',
        '@FEATURE:NEW-FEATURE-001 Feature description',
      ];

      validTags.forEach(tag => {
        expect(tagParser.validateTagFormat(tag)).toBe(true);
      });
    });

    it('should reject invalid TAG formats', () => {
      const invalidTags = [
        'REQ:NO-AT-SYMBOL',
        '@req:lowercase-category',
        '@REQ:',
        '@:NO-CATEGORY',
        '@INVALID:WRONG-CATEGORY',
        'Not a tag at all',
      ];

      invalidTags.forEach(tag => {
        expect(tagParser.validateTagFormat(tag)).toBe(false);
      });
    });
  });

  describe('@TEST:TAG-POSITIONS-001 TAG 위치 추출', () => {
    it('should extract tags with line and column positions', () => {
      const content = `Line 1: @REQ:FIRST-001 First requirement
Line 2: Some content
Line 3: @DESIGN:SECOND-001 Design document`;

      const tagsWithPositions = tagParser.extractTagsWithPositions(content);

      expect(tagsWithPositions).toHaveLength(2);

      const [firstTag, firstPos] = tagsWithPositions[0]!;
      expect(firstTag.category).toBe('REQ');
      expect(firstTag.identifier).toBe('FIRST-001');
      expect(firstPos.lineNumber).toBe(1);
      expect(firstPos.column).toBe(8); // Position of '@' in "Line 1: @REQ..."

      const [secondTag, secondPos] = tagsWithPositions[1]!;
      expect(secondTag.category).toBe('DESIGN');
      expect(secondTag.identifier).toBe('SECOND-001');
      expect(secondPos.lineNumber).toBe(3);
      expect(secondPos.column).toBe(8); // Position of '@' in "Line 3: @DESIGN..."
    });
  });

  describe('@TEST:DUPLICATE-TAG-DETECTION-001 중복 TAG 검출', () => {
    it('should find duplicate tags', () => {
      const content = `
        @REQ:DUPLICATE-001 First occurrence
        Some content here
        @REQ:DUPLICATE-001 Second occurrence
        @TASK:UNIQUE-001 Unique tag
        @REQ:DUPLICATE-001 Third occurrence
      `;

      const duplicates = tagParser.findDuplicateTags(content);

      expect(duplicates).toHaveLength(1);
      expect(duplicates[0]!.category).toBe('REQ');
      expect(duplicates[0]!.identifier).toBe('DUPLICATE-001');
      expect(duplicates[0]!.positions).toHaveLength(3);
      expect(duplicates[0]!.positions[0]!.lineNumber).toBe(2);
      expect(duplicates[0]!.positions[1]!.lineNumber).toBe(4);
      expect(duplicates[0]!.positions[2]!.lineNumber).toBe(6);
    });

    it('should return empty array when no duplicates exist', () => {
      const content = `
        @REQ:UNIQUE-001 First tag
        @DESIGN:UNIQUE-002 Second tag
        @TASK:UNIQUE-003 Third tag
      `;

      const duplicates = tagParser.findDuplicateTags(content);

      expect(duplicates).toHaveLength(0);
    });
  });

  describe('@TEST:EDGE-CASES-001 경계 조건 처리', () => {
    it('should handle empty content', () => {
      const tags = tagParser.extractTags('');
      const chains = tagParser.parseTagChains('');
      const duplicates = tagParser.findDuplicateTags('');

      expect(tags).toHaveLength(0);
      expect(chains).toHaveLength(0);
      expect(duplicates).toHaveLength(0);
    });

    it('should handle content with no tags', () => {
      const content = 'This is just regular text without any tags';

      const tags = tagParser.extractTags(content);
      const chains = tagParser.parseTagChains(content);

      expect(tags).toHaveLength(0);
      expect(chains).toHaveLength(0);
    });

    it('should handle malformed tags gracefully', () => {
      const content = `
        @REQ:GOOD-001 Valid tag
        @@REQ:DOUBLE-AT-001 Double at
        @REQ::DOUBLE-COLON-001 Double colon
        @REQ:VALID-002 Another valid tag
      `;

      const tags = tagParser.extractTags(content);

      expect(tags).toHaveLength(2); // Only valid tags
      expect(tags[0]!.identifier).toBe('GOOD-001');
      expect(tags[1]!.identifier).toBe('VALID-002');
    });
  });

  describe('@TEST:REAL-WORLD-SCENARIOS-001 실제 사용 시나리오', () => {
    it('should handle complex document with mixed TAG usage', () => {
      const content = `
# Project Documentation

## Requirements
@REQ:USER-REGISTRATION-001 사용자는 이메일로 회원가입할 수 있어야 한다

## Design
@DESIGN:AUTH-FLOW-001 → @TASK:IMPL-AUTH-001 → @TEST:AUTH-VERIFICATION-001

## Implementation
@FEATURE:USER-MANAGEMENT-001 사용자 관리 시스템
@API:USER-ENDPOINT-001 사용자 API 엔드포인트
@UI:REGISTRATION-FORM-001 회원가입 폼

## Quality
@PERF:RESPONSE-TIME-001 API 응답 시간 1초 이내
@SEC:PASSWORD-POLICY-001 비밀번호 정책 적용
@DOCS:API-DOCUMENTATION-001 API 문서화
      `;

      const tags = tagParser.extractTags(content);
      const chains = tagParser.parseTagChains(content);

      expect(tags.length).toBeGreaterThan(5);
      expect(chains).toHaveLength(1);

      // Verify all 4 categories are represented
      const categories = new Set(tags.map(tag => tag.category));
      expect(categories.has('REQ')).toBe(true);
      expect(categories.has('DESIGN')).toBe(true);
      expect(categories.has('FEATURE')).toBe(true);
      expect(categories.has('PERF')).toBe(true);
    });
  });
});
