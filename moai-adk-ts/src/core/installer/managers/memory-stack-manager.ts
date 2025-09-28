/**
 * @FEATURE:MEMORY-STACK-001 Memory Stack Template Manager for MoAI-ADK
 *
 * Python의 MEMORY_STACK_TEMPLATE_MAP 기능을 TypeScript로 완전 포팅
 * 기술 스택에 따라 적절한 메모리 템플릿을 자동 선택하는 핵심 시스템
 *
 * @TASK:MEMORY-STACK-001 기술 스택별 메모리 템플릿 매핑 관리
 * @DESIGN:PYTHON-PORTING-001 Python template_manager.py의 완전 호환 구현
 */

import { logger } from '../../../utils/logger';
import type { TechStack, MemoryTemplate } from '../types';

/**
 * 메모리 스택 템플릿 매핑 테이블
 * Python의 MEMORY_STACK_TEMPLATE_MAP과 완전 동일한 매핑
 */
const MEMORY_STACK_TEMPLATE_MAP: Record<string, MemoryTemplate[]> = {
  // Python 백엔드 스택
  python: ['backend-python'],
  fastapi: ['backend-python', 'backend-fastapi'],
  django: ['backend-python'],
  flask: ['backend-python'],

  // Java/Spring 백엔드 스택
  java: ['backend-spring'],
  spring: ['backend-spring'],
  'spring boot': ['backend-spring'],
  springboot: ['backend-spring'],
  'spring-boot': ['backend-spring'],

  // 프론트엔드 스택
  react: ['frontend-react'],
  nextjs: ['frontend-react', 'frontend-next'],
  vue: ['frontend-vue'],
  nuxt: ['frontend-vue'],
  angular: ['frontend-angular'],
  typescript: ['frontend-react'],
  javascript: ['frontend-react'],
};

/**
 * @DESIGN:CLASS-001 Memory Stack Manager
 *
 * Python TemplateManager.get_memory_templates_for_stack() 메서드의 완전 포팅
 * 기술 스택 배열을 받아서 해당하는 메모리 템플릿 목록을 반환
 */
export class MemoryStackManager {
  /**
   * @API:GET-MEMORY-TEMPLATES-001 기술 스택별 메모리 템플릿 목록 반환
   *
   * Python 구현과 동일한 로직:
   * 1. 기본적으로 'development-guide' 포함
   * 2. 각 기술 스택에 대해 매핑된 템플릿 추가
   * 3. 중복 제거하면서 순서 보존
   *
   * @param techStack 기술 스택 배열
   * @returns 메모리 템플릿 이름 배열
   */
  getMemoryTemplatesForStack(techStack: TechStack[]): string[] {
    // Python 구현: templates_to_copy = ["development-guide"]
    const templatesToCopy: string[] = ['development-guide'];

    // Python 구현: for tech in tech_stack:
    for (const tech of techStack) {
      // Python 구현: tech_key = tech.lower()
      const techKey = tech.toLowerCase();

      // Python 구현: if tech_key in MEMORY_STACK_TEMPLATE_MAP:
      if (techKey in MEMORY_STACK_TEMPLATE_MAP) {
        // Python 구현: templates_to_copy.extend(MEMORY_STACK_TEMPLATE_MAP[tech_key])
        templatesToCopy.push(...MEMORY_STACK_TEMPLATE_MAP[techKey]!);
      }
    }

    // Python 구현: 중복 제거하면서 순서 보존
    // Remove duplicates while preserving order
    // seen = set()
    // unique_templates = []
    // for template in templates_to_copy:
    //     if template not in seen:
    //         seen.add(template)
    //         unique_templates.append(template)
    const seen = new Set<string>();
    const uniqueTemplates: string[] = [];

    for (const template of templatesToCopy) {
      if (!seen.has(template)) {
        seen.add(template);
        uniqueTemplates.push(template);
      }
    }

    logger.debug(
      `Memory templates for stack [${techStack.join(', ')}]: ${uniqueTemplates.join(', ')}`
    );

    return uniqueTemplates;
  }

  /**
   * @API:GET-AVAILABLE-STACKS-001 지원되는 기술 스택 목록 반환
   *
   * @returns 지원되는 기술 스택 목록
   */
  getAvailableTechStacks(): string[] {
    return Object.keys(MEMORY_STACK_TEMPLATE_MAP);
  }

  /**
   * @API:GET-TEMPLATES-FOR-STACK-001 특정 기술 스택의 템플릿 반환
   *
   * @param techStack 단일 기술 스택
   * @returns 해당 스택의 템플릿 목록 (development-guide 제외)
   */
  getTemplatesForStack(techStack: TechStack): MemoryTemplate[] {
    const techKey = techStack.toLowerCase();
    return MEMORY_STACK_TEMPLATE_MAP[techKey] || [];
  }

  /**
   * @API:IS-STACK-SUPPORTED-001 기술 스택 지원 여부 확인
   *
   * @param techStack 확인할 기술 스택
   * @returns 지원 여부
   */
  isStackSupported(techStack: TechStack): boolean {
    const techKey = techStack.toLowerCase();
    return techKey in MEMORY_STACK_TEMPLATE_MAP;
  }

  /**
   * @API:GET-MAPPING-INFO-001 전체 매핑 정보 반환 (디버깅용)
   *
   * @returns 전체 매핑 테이블
   */
  getMappingInfo(): Record<string, MemoryTemplate[]> {
    return { ...MEMORY_STACK_TEMPLATE_MAP };
  }

  /**
   * @API:ADD-CUSTOM-MAPPING-001 사용자 정의 매핑 추가
   *
   * @param techStack 기술 스택명
   * @param templates 매핑할 템플릿 목록
   */
  addCustomMapping(techStack: string, templates: MemoryTemplate[]): void {
    const techKey = techStack.toLowerCase();
    MEMORY_STACK_TEMPLATE_MAP[techKey] = templates;
    logger.info(
      `Added custom mapping: ${techStack} -> [${templates.join(', ')}]`
    );
  }

  /**
   * @API:VALIDATE-MAPPING-001 매핑 설정 검증
   *
   * @returns 검증 결과
   */
  validateMapping(): { valid: boolean; issues: string[] } {
    const issues: string[] = [];

    // 각 매핑이 유효한 템플릿을 참조하는지 확인
    for (const [stack, templates] of Object.entries(
      MEMORY_STACK_TEMPLATE_MAP
    )) {
      if (templates.length === 0) {
        issues.push(`Empty template list for stack: ${stack}`);
      }

      for (const template of templates) {
        if (!this.isValidTemplate(template)) {
          issues.push(`Invalid template '${template}' for stack '${stack}'`);
        }
      }
    }

    return {
      valid: issues.length === 0,
      issues,
    };
  }

  /**
   * @HELPER:IS-VALID-TEMPLATE-001 템플릿 이름 유효성 검사
   *
   * @param template 검사할 템플릿 이름
   * @returns 유효성 여부
   */
  private isValidTemplate(template: string): boolean {
    // 기본 템플릿 목록과 비교 (실제 구현에서는 파일 시스템 확인 가능)
    const validTemplates = [
      'development-guide',
      'backend-python',
      'backend-fastapi',
      'backend-spring',
      'backend-express',
      'frontend-react',
      'frontend-next',
      'frontend-vue',
      'frontend-angular',
      'fullstack-patterns',
      'microservice-patterns',
    ];

    return validTemplates.includes(template);
  }
}

/**
 * 싱글톤 MemoryStackManager 인스턴스
 */
export const memoryStackManager = new MemoryStackManager();
