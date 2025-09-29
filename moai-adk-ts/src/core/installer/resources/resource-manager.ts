/**
 * @FEATURE:RESOURCE-001 MoAI-ADK TypeScript Resource Manager
 *
 * @TASK:RESOURCE-TS-001 패키지 내장 리소스를 관리하는 TypeScript 모듈입니다.
 * @TASK:RESOURCE-TS-002 Python ResourceManager와 동일한 기능을 TypeScript로 포팅합니다.
 * @REFACTOR:TRUST-001 TRUST 원칙 적용으로 유틸리티 클래스 분리
 * @REFACTOR:RESOURCE-SPLIT-001 리소스 작업을 ResourceOperations로 분리
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { logger } from '../../../utils/logger';
import type { TemplateContext as TemplateContextType } from './utils';
import { ResourceOperations } from './resource-operations';

// Re-export for external use
export type TemplateContext = TemplateContextType;

/**
 * @DESIGN:INTERFACE-001 리소스 복사 결과 인터페이스
 */
export interface ResourceCopyResult {
  readonly success: boolean;
  readonly copiedFiles: readonly string[];
  readonly errors: readonly string[];
}

/**
 * @DESIGN:INTERFACE-002 리소스 검증 결과 인터페이스
 */
export interface ResourceValidationResult {
  readonly valid: boolean;
  readonly missingFiles: readonly string[];
  readonly invalidFiles: readonly string[];
}

// TemplateContext는 utils/template-processor.ts에서 import

/**
 * @DESIGN:CLASS-001 TypeScript Resource Manager
 *
 * Python ResourceManager의 TypeScript 포팅 버전
 * 패키지 내장 리소스 관리 및 프로젝트 설정 파일 복사를 담당합니다.
 */
export class ResourceManager {
  private readonly resourcesRoot: string;
  private readonly templatesRoot: string;
  private readonly resourceOperations: ResourceOperations;

  /**
   * @TASK:INIT-001 ResourceManager 초기화
   * @REFACTOR:COMPOSITION-001 ResourceOperations로 작업 위임
   */
  constructor() {
    // 리소스 경로 설정
    this.resourcesRoot = path.join(__dirname, '../../../../resources');
    this.templatesRoot = path.join(this.resourcesRoot, 'templates');

    // 리소스 작업 클래스 초기화
    this.resourceOperations = new ResourceOperations(this.templatesRoot);

    logger.info('ResourceManager initialized with ResourceOperations');
  }

  /**
   * @API:VERSION-001 패키지 버전 반환
   */
  async getVersion(): Promise<string> {
    try {
      const versionFile = path.join(this.resourcesRoot, 'VERSION');
      const version = await fs.readFile(versionFile, 'utf-8');
      return version.trim();
    } catch (error) {
      logger.warn(`Could not read version: ${error}`);
      return 'unknown';
    }
  }

  /**
   * @API:TEMPLATE-PATH-001 템플릿 경로 반환
   */
  getTemplatePath(templateName: string): string {
    return path.join(this.templatesRoot, templateName);
  }

  /**
   * @API:TEMPLATE-CONTENT-001 템플릿 내용 반환
   */
  async getTemplateContent(templateName: string): Promise<string | null> {
    try {
      const templatePath = this.getTemplatePath(templateName);
      const stat = await fs.stat(templatePath);

      if (stat.isFile()) {
        const ext = path.extname(templatePath);
        const allowedExts = ['.md', '.json', '.yml', '.yaml', '.txt'];

        if (allowedExts.includes(ext)) {
          return await fs.readFile(templatePath, 'utf-8');
        }
      }
      return null;
    } catch (error) {
      logger.warn(`Failed to read template content ${templateName}: ${error}`);
      return null;
    }
  }

  /**
   * @API:CLAUDE-RESOURCES-001 Claude Code 관련 리소스를 프로젝트에 복사
   */
  async copyClaudeResources(
    targetPath: string,
    overwrite: boolean = false
  ): Promise<string[]> {
    return await this.resourceOperations.copyClaudeResources(
      targetPath,
      overwrite
    );
  }

  /**
   * @API:MOAI-RESOURCES-001 MoAI 관련 리소스를 프로젝트에 복사
   */
  async copyMoaiResources(
    targetPath: string,
    overwrite: boolean = false,
    excludeTemplates: boolean = false,
    projectContext?: TemplateContextType
  ): Promise<string[]> {
    return await this.resourceOperations.copyMoaiResources(
      targetPath,
      overwrite,
      excludeTemplates,
      projectContext
    );
  }

  /**
   * @API:PROJECT-MEMORY-001 프로젝트 메모리 파일(CLAUDE.md) 생성
   */
  async copyProjectMemory(
    projectPath: string,
    overwrite: boolean = false,
    projectContext?: TemplateContextType
  ): Promise<boolean> {
    return await this.resourceOperations.copyProjectMemory(
      projectPath,
      overwrite,
      projectContext
    );
  }

  /**
   * @API:VALIDATE-001 프로젝트 리소스 검증
   */
  async validateProjectResources(projectPath: string): Promise<boolean> {
    const requiredPaths = [
      path.join(projectPath, '.claude'),
      path.join(projectPath, '.moai'),
      path.join(projectPath, 'CLAUDE.md'),
    ];

    const missingPaths: string[] = [];

    for (const requiredPath of requiredPaths) {
      try {
        await fs.access(requiredPath);
      } catch {
        missingPaths.push(requiredPath);
      }
    }

    if (missingPaths.length > 0) {
      logger.warn(`Missing required resources: ${missingPaths}`);
      return false;
    }

    logger.info('All required resources are present');
    return true;
  }

  // All complex operations are now delegated to ResourceOperations
  // This follows TRUST-U principle: Single responsibility and composition over inheritance
}
