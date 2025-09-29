/**
 * @REFACTOR:RESOURCE-OPERATIONS-001 리소스 복사 작업 전용 클래스
 *
 * @TASK:COPY-OPERATIONS-001 리소스 복사 로직만 담당하는 전용 클래스
 * TRUST-U 원칙: 단일 책임 및 300 LOC 이하 파일 크기
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { logger } from '../../../utils/logger';
import {
  PathValidator,
  TemplateProcessor,
  FileOperations,
  type TemplateContext,
} from './utils';

/**
 * @DESIGN:CLASS-001 리소스 복사 작업 전용 클래스
 *
 * 템플릿 복사, 디렉토리 생성, 권한 설정 등의 실제 작업을 담당합니다.
 */
export class ResourceOperations {
  private readonly templatesRoot: string;
  private readonly pathValidator: PathValidator;
  private readonly templateProcessor: TemplateProcessor;
  private readonly fileOperations: FileOperations;

  constructor(templatesRoot: string) {
    this.templatesRoot = templatesRoot;
    this.pathValidator = new PathValidator();
    this.templateProcessor = new TemplateProcessor();
    this.fileOperations = new FileOperations();
  }

  /**
   * @API:COPY-CLAUDE-001 Claude 리소스 복사
   */
  async copyClaudeResources(
    targetPath: string,
    overwrite: boolean = false
  ): Promise<string[]> {
    const copiedFiles: string[] = [];
    const claudeResources = ['.claude'];

    for (const resource of claudeResources) {
      const resourceTargetPath = path.join(targetPath, resource);
      const success = await this.copyTemplate(
        resource,
        resourceTargetPath,
        overwrite
      );

      if (success) {
        // 실행 권한 보장 (Unix 계열 시스템에서)
        try {
          await this.fileOperations.ensureHookPermissions(resourceTargetPath);
        } catch (error) {
          logger.warn(`Hook permission ensure skipped: ${error}`);
        }
        copiedFiles.push(resourceTargetPath);
      }
    }

    return copiedFiles;
  }

  /**
   * @API:COPY-MOAI-001 MoAI 리소스 복사
   */
  async copyMoaiResources(
    targetPath: string,
    overwrite: boolean = false,
    excludeTemplates: boolean = false,
    projectContext?: TemplateContext
  ): Promise<string[]> {
    const copiedFiles: string[] = [];
    const moaiResources = ['.moai'];

    logger.info('Starting MoAI resources installation...');

    for (const resource of moaiResources) {
      const resourceTargetPath = path.join(targetPath, resource);
      logger.info(`Installing ${resource} to ${resourceTargetPath}`);

      const excludeSubdirs = excludeTemplates ? ['_templates'] : undefined;
      const success = await this.copyTemplate(
        resource,
        resourceTargetPath,
        overwrite,
        excludeSubdirs
      );

      if (success) {
        copiedFiles.push(resourceTargetPath);
        await this.fileOperations.validateCleanInstallation(resourceTargetPath);
        logger.info(`Successfully installed ${resource}`);
      } else {
        logger.error(`Failed to install ${resource}`);
      }
    }

    // 템플릿 변수 치환 수행
    if (projectContext && copiedFiles.length > 0) {
      await this.applyProjectContextBatch(copiedFiles, projectContext);
    }

    logger.info(
      `MoAI resources installation completed. ${copiedFiles.length} resources installed.`
    );
    return copiedFiles;
  }

  /**
   * @API:COPY-PROJECT-MEMORY-001 프로젝트 메모리 파일 복사
   */
  async copyProjectMemory(
    projectPath: string,
    overwrite: boolean = false,
    projectContext?: TemplateContext
  ): Promise<boolean> {
    try {
      const targetPath = path.join(projectPath, 'CLAUDE.md');
      const success = await this.copyTemplate(
        'CLAUDE.md',
        targetPath,
        overwrite
      );

      if (success && projectContext) {
        await this.templateProcessor.substituteFileVariables(
          targetPath,
          projectContext
        );
      }

      return success;
    } catch (error) {
      logger.error(`Failed to copy project memory: ${error}`);
      return false;
    }
  }

  /**
   * @TASK:COPY-TEMPLATE-001 템플릿을 대상 경로로 복사
   *
   * 단일 책임: 템플릿 복사 로직만 담당 (TRUST-U)
   */
  async copyTemplate(
    templateName: string,
    targetPath: string,
    overwrite: boolean = false,
    excludeSubdirs?: string[]
  ): Promise<boolean> {
    try {
      // 경로 안전성 검증
      if (!this.pathValidator.validateSafePath(targetPath)) {
        throw new Error(`Unsafe target path detected: ${targetPath}`);
      }

      // 절대 경로로 변환
      const absoluteTargetPath = path.resolve(targetPath);
      const templatePath = this.getTemplatePath(templateName);

      // 대상 경로 존재 여부 확인
      try {
        const stat = await fs.stat(absoluteTargetPath);
        if (stat.isFile() && !overwrite) {
          logger.info(
            `Target file already exists, skipping: ${absoluteTargetPath}`
          );
          return true;
        } else if (stat.isDirectory()) {
          logger.info(
            `Target directory exists, will merge contents: ${absoluteTargetPath}`
          );
        }
      } catch {
        // 파일이 존재하지 않음 - 정상적인 케이스
      }

      // 부모 디렉토리 생성
      await fs.mkdir(path.dirname(absoluteTargetPath), { recursive: true });

      // 템플릿 소스 확인 및 복사
      const sourceStat = await fs.stat(templatePath);

      if (sourceStat.isDirectory()) {
        await this.fileOperations.copyDirectory(
          templatePath,
          absoluteTargetPath,
          overwrite,
          excludeSubdirs
        );
      } else {
        await fs.copyFile(templatePath, absoluteTargetPath);
      }

      logger.info(
        `Successfully copied ${templateName} to ${absoluteTargetPath}`
      );
      return true;
    } catch (error) {
      logger.error(`Failed to copy template ${templateName}: ${error}`);
      return false;
    }
  }

  /**
   * @TASK:GET-TEMPLATE-PATH-001 템플릿 경로 반환
   */
  private getTemplatePath(templateName: string): string {
    return path.join(this.templatesRoot, templateName);
  }

  /**
   * @TASK:APPLY-CONTEXT-BATCH-001 프로젝트 컨텍스트 일괄 적용
   */
  private async applyProjectContextBatch(
    copiedPaths: string[],
    projectContext: TemplateContext
  ): Promise<void> {
    const templateFiles = [
      'config.json',
      'project/product.md',
      'project/structure.md',
      'project/tech.md',
    ];

    const filesToProcess: string[] = [];

    // 처리할 파일 목록 수집
    for (const basePath of copiedPaths) {
      try {
        const stat = await fs.stat(basePath);
        if (!stat.isDirectory()) continue;

        for (const templateFile of templateFiles) {
          const filePath = path.join(basePath, templateFile);
          try {
            await fs.access(filePath);
            filesToProcess.push(filePath);
          } catch {
            // 파일이 존재하지 않음 - 정상적인 케이스
          }
        }
      } catch {
      }
    }

    // 일괄 템플릿 변수 치환
    await this.templateProcessor.substituteBatchVariables(
      filesToProcess,
      projectContext
    );
  }
}
