/**
 * @TEST:MEMORY-STACK-001 Memory Stack Template Manager Tests
 *
 * Python의 MEMORY_STACK_TEMPLATE_MAP 기능을 TypeScript로 포팅한 테스트
 * @FEATURE:MEMORY-STACK-001 기술 스택별 메모리 템플릿 자동 선택
 */

import { MemoryStackManager } from '../memory-stack-manager';
import type { TechStack } from '../../types';

describe('MemoryStackManager', () => {
  let memoryStackManager: MemoryStackManager;

  beforeEach(() => {
    memoryStackManager = new MemoryStackManager();
  });

  describe('@TEST:STACK-MAPPING-001 기술 스택별 템플릿 매핑', () => {
    it('should map Python stack to backend-python template', () => {
      const techStack: TechStack[] = ['python'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toContain('backend-python');
      expect(templates).toContain('development-guide'); // 기본 포함
    });

    it('should map FastAPI stack to backend-python and backend-fastapi templates', () => {
      const techStack: TechStack[] = ['fastapi'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toContain('backend-python');
      expect(templates).toContain('backend-fastapi');
      expect(templates).toContain('development-guide');
    });

    it('should map React stack to frontend-react template', () => {
      const techStack: TechStack[] = ['react'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toContain('frontend-react');
      expect(templates).toContain('development-guide');
    });

    it('should map NextJS stack to frontend-react and frontend-next templates', () => {
      const techStack: TechStack[] = ['nextjs'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toContain('frontend-react');
      expect(templates).toContain('frontend-next');
      expect(templates).toContain('development-guide');
    });

    it('should map Spring Boot stack correctly (multiple naming variants)', () => {
      const variants: TechStack[][] = [
        ['spring'],
        ['spring boot'],
        ['springboot'],
        ['spring-boot'],
      ];

      variants.forEach(stack => {
        const templates = memoryStackManager.getMemoryTemplatesForStack(stack);
        expect(templates).toContain('backend-spring');
        expect(templates).toContain('development-guide');
      });
    });
  });

  describe('@TEST:STACK-COMBINATION-001 기술 스택 조합 처리', () => {
    it('should handle multiple tech stacks without duplication', () => {
      const techStack: TechStack[] = ['python', 'fastapi', 'react'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      // 중복 제거 확인
      const uniqueTemplates = Array.from(new Set(templates));
      expect(templates.length).toBe(uniqueTemplates.length);

      // 모든 관련 템플릿 포함 확인
      expect(templates).toContain('development-guide');
      expect(templates).toContain('backend-python');
      expect(templates).toContain('backend-fastapi');
      expect(templates).toContain('frontend-react');
    });

    it('should preserve order while removing duplicates', () => {
      const techStack: TechStack[] = ['python', 'python', 'react'];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      const expectedOrder = [
        'development-guide',
        'backend-python',
        'frontend-react',
      ];
      expect(templates).toEqual(expectedOrder);
    });
  });

  describe('@TEST:UNKNOWN-STACK-001 알려지지 않은 기술 스택 처리', () => {
    it('should return only development-guide for unknown tech stack', () => {
      const techStack: TechStack[] = ['unknown-framework' as TechStack];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toEqual(['development-guide']);
    });

    it('should handle mix of known and unknown stacks', () => {
      const techStack: TechStack[] = [
        'python',
        'unknown-framework' as TechStack,
        'react',
      ];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toContain('development-guide');
      expect(templates).toContain('backend-python');
      expect(templates).toContain('frontend-react');
      expect(templates).not.toContain('unknown-framework');
    });
  });

  describe('@TEST:EMPTY-STACK-001 빈 기술 스택 처리', () => {
    it('should return development-guide for empty tech stack', () => {
      const techStack: TechStack[] = [];
      const templates =
        memoryStackManager.getMemoryTemplatesForStack(techStack);

      expect(templates).toEqual(['development-guide']);
    });
  });

  describe('@TEST:CASE-INSENSITIVE-001 대소문자 무관 처리', () => {
    it('should handle case-insensitive tech stack names', () => {
      const testCases = [
        ['Python', 'PYTHON', 'python'],
        ['React', 'REACT', 'react'],
        ['FastAPI', 'FASTAPI', 'fastapi'],
      ];

      testCases.forEach(variants => {
        const results = variants.map(variant =>
          memoryStackManager.getMemoryTemplatesForStack([variant as TechStack])
        );

        // 모든 변형이 같은 결과를 반환해야 함
        const firstResult = results[0];
        results.forEach(result => {
          expect(result).toEqual(firstResult);
        });
      });
    });
  });

  describe('@TEST:TEMPLATE-MAP-001 템플릿 매핑 정확성', () => {
    it('should have correct complete mapping as per Python implementation', () => {
      const expectedMappings = {
        python: ['backend-python'],
        fastapi: ['backend-python', 'backend-fastapi'],
        django: ['backend-python'],
        flask: ['backend-python'],
        java: ['backend-spring'],
        spring: ['backend-spring'],
        'spring boot': ['backend-spring'],
        springboot: ['backend-spring'],
        'spring-boot': ['backend-spring'],
        react: ['frontend-react'],
        nextjs: ['frontend-react', 'frontend-next'],
        vue: ['frontend-vue'],
        nuxt: ['frontend-vue'],
        angular: ['frontend-angular'],
        typescript: ['frontend-react'],
        javascript: ['frontend-react'],
      };

      Object.entries(expectedMappings).forEach(([tech, expectedTemplates]) => {
        const templates = memoryStackManager.getMemoryTemplatesForStack([
          tech as TechStack,
        ]);

        // development-guide 제외하고 비교
        const actualTemplates = templates.filter(
          (t: string) => t !== 'development-guide'
        );
        expect(actualTemplates).toEqual(expectedTemplates);
      });
    });
  });
});
