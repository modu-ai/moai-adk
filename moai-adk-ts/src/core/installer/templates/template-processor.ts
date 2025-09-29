/**
 * @FEATURE:TEMPLATE-PROCESSOR-001 Advanced Template Processor for MoAI-ADK
 *
 * Python의 고급 템플릿 처리 로직을 TypeScript로 완전 포팅
 * @TASK:TEMPLATE-PROCESSOR-001 복잡한 템플릿 워크플로우 지원
 * @DESIGN:MOAI-INTEGRATION-001 MoAI 프로젝트 특화 템플릿 기능
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { logger } from '../../../utils/logger';
import {
  processMultipleVariableFormats,
  expandNestedVariables,
  shouldProcessAsTemplate,
  fileExists,
  copyBinaryFile,
  mergeTemplateContexts,
} from './template-utils';
import { validateTemplateContext } from './template-validator';
import { renderTemplateSafely, validateTemplateContent } from './template-security';

/**
 * @DESIGN:INTERFACES-001 Template Processor Core Interfaces
 */

export interface TemplateContext {
  readonly [key: string]: string | number | boolean | object | Array<any>;
}

export interface ProcessResult {
  readonly success: boolean;
  readonly processedFiles: readonly string[];
  readonly skippedFiles: readonly string[];
  readonly errors: readonly ProcessError[];
}

export interface ProcessError {
  readonly file: string;
  readonly error: string;
  readonly line?: number;
  readonly column?: number;
}

export interface ContextSchema {
  readonly required: readonly string[];
  readonly optional: readonly string[];
  readonly defaults: Record<string, any>;
  readonly validation: Record<string, (value: any) => boolean>;
}

export interface ValidationResult {
  readonly valid: boolean;
  readonly errors: readonly string[];
  readonly warnings: readonly string[];
}

export interface ProcessingOptions {
  readonly overwrite: boolean;
  readonly preserveTimestamps: boolean;
  readonly createDirectories: boolean;
  readonly fileNameTemplating: boolean;
  readonly conditionalProcessing: boolean;
}

interface ProcessingState {
  readonly processedFiles: string[];
  readonly skippedFiles: string[];
  readonly errors: ProcessError[];
}

/**
 * @DESIGN:CLASS-001 Advanced Template Processor
 *
 * Python TemplateProcessor의 모든 고급 기능을 TypeScript로 포팅
 */
export class TemplateProcessor {
  private readonly defaultOptions: ProcessingOptions = {
    overwrite: true,
    preserveTimestamps: false,
    createDirectories: true,
    fileNameTemplating: true,
    conditionalProcessing: true,
  };

  /**
   * @API:PROCESS-DIRECTORY-001 디렉토리 단위 템플릿 처리
   *
   * @param sourceDir 소스 디렉토리 경로
   * @param targetDir 타겟 디렉토리 경로
   * @param context 템플릿 컨텍스트
   * @param options 처리 옵션
   * @returns 처리 결과
   */
  async processTemplateDirectory(
    sourceDir: string,
    targetDir: string,
    context: TemplateContext,
    options: Partial<ProcessingOptions> = {}
  ): Promise<ProcessResult> {
    const opts = { ...this.defaultOptions, ...options };
    const processedFiles: string[] = [];
    const skippedFiles: string[] = [];
    const errors: ProcessError[] = [];

    try {
      // 타겟 디렉토리 생성
      if (opts.createDirectories) {
        await fs.mkdir(targetDir, { recursive: true });
      }

      const state: ProcessingState = { processedFiles, skippedFiles, errors };
      await this.processDirectoryRecursive(
        sourceDir,
        targetDir,
        context,
        opts,
        state
      );

      return {
        success: errors.length === 0,
        processedFiles: Object.freeze(processedFiles),
        skippedFiles: Object.freeze(skippedFiles),
        errors: Object.freeze(errors),
      };
    } catch (error) {
      const processError: ProcessError = {
        file: sourceDir,
        error: error instanceof Error ? error.message : String(error),
      };

      return {
        success: false,
        processedFiles: Object.freeze(processedFiles),
        skippedFiles: Object.freeze(skippedFiles),
        errors: Object.freeze([processError]),
      };
    }
  }

  /**
   * @API:PROCESS-FILE-001 개별 파일 템플릿 처리
   *
   * @param templatePath 템플릿 파일 경로
   * @param outputPath 출력 파일 경로
   * @param context 템플릿 컨텍스트
   * @param options 처리 옵션
   */
  async processTemplateFile(
    templatePath: string,
    outputPath: string,
    context: TemplateContext,
    options: Partial<ProcessingOptions> = {}
  ): Promise<void> {
    const opts = { ...this.defaultOptions, ...options };

    try {
      await this.prepareOutputDirectory(outputPath, opts);

      if (!opts.overwrite && (await fileExists(outputPath))) {
        logger.debug(`File exists, skipping: ${outputPath}`);
        return;
      }

      const processedContent = await this.processTemplateContent(
        templatePath,
        context,
        opts
      );
      await this.writeProcessedFile(
        outputPath,
        processedContent,
        templatePath,
        opts
      );

      logger.debug(`Template file processed: ${templatePath} -> ${outputPath}`);
    } catch (error) {
      logger.error(`Failed to process template file ${templatePath}: ${error}`);
      throw error;
    }
  }

  /**
   * @TASK:PREPARE-OUTPUT-001 출력 디렉토리 준비
   */
  private async prepareOutputDirectory(
    outputPath: string,
    options: ProcessingOptions
  ): Promise<void> {
    if (options.createDirectories) {
      await fs.mkdir(path.dirname(outputPath), { recursive: true });
    }
  }

  /**
   * @TASK:PROCESS-CONTENT-001 템플릿 내용 처리
   */
  private async processTemplateContent(
    templatePath: string,
    context: TemplateContext,
    options: ProcessingOptions
  ): Promise<string> {
    const templateContent = await fs.readFile(templatePath, 'utf-8');

    let processedContent = templateContent;

    if (options.conditionalProcessing) {
      processedContent = this.safeApplyConditionalLogic(
        processedContent,
        context,
        templatePath
      );
    }

    return this.safeExpandTemplateVariables(
      processedContent,
      context,
      templatePath
    );
  }

  /**
   * @TASK:WRITE-FILE-001 처리된 파일 쓰기
   */
  private async writeProcessedFile(
    outputPath: string,
    content: string,
    templatePath: string,
    options: ProcessingOptions
  ): Promise<void> {
    await fs.writeFile(outputPath, content, 'utf-8');

    if (options.preserveTimestamps) {
      const stats = await fs.stat(templatePath);
      await fs.utimes(outputPath, stats.atime, stats.mtime);
    }
  }

  /**
   * @TASK:SAFE-CONDITIONAL-001 안전한 조건부 로직 적용
   */
  private safeApplyConditionalLogic(
    content: string,
    context: TemplateContext,
    templatePath: string
  ): string {
    try {
      return this.applyConditionalLogic(content, context);
    } catch (error) {
      logger.warn(
        `Conditional processing failed for ${templatePath}, using original content: ${error}`
      );
      return content;
    }
  }

  /**
   * @TASK:SAFE-EXPAND-001 안전한 변수 확장
   */
  private safeExpandTemplateVariables(
    content: string,
    context: TemplateContext,
    templatePath: string
  ): string {
    try {
      return this.expandTemplateVariables(content, context);
    } catch (error) {
      logger.warn(
        `Variable expansion failed for ${templatePath}, using processed content: ${error}`
      );
      return content;
    }
  }

  /**
   * @API:EXPAND-VARIABLES-001 고급 변수 확장 (보안 강화)
   *
   * 다중 포맷 지원 및 중첩 변수 확장 - Template Injection 방어
   */
  expandTemplateVariables(content: string, context: TemplateContext): string {
    try {
      // 보안 검증: 템플릿 내용 검사
      if (!validateTemplateContent(content)) {
        logger.error('Template content contains dangerous patterns');
        throw new Error('Template security validation failed');
      }

      // 1단계: 중첩 변수 확장 먼저 처리 ({{PROJECT_{{ENV}}_CONFIG}} 형식)
      let result = expandNestedVariables(content, context);

      // 2단계: 안전한 Mustache 템플릿 처리 ({{VAR}} 형식)
      result = renderTemplateSafely(result, context as Record<string, any>);

      // 3단계: 다중 포맷 지원 ([VAR], ${VAR}, $VAR)
      result = processMultipleVariableFormats(result, context);

      return result;
    } catch (error) {
      logger.error(`Failed to expand template variables: ${error}`);
      return content;
    }
  }

  /**
   * @API:CONDITIONAL-LOGIC-001 조건부 로직 적용 (보안 강화)
   *
   * 안전한 Mustache 기반 조건부 블록 및 반복 구조 처리
   */
  applyConditionalLogic(template: string, context: TemplateContext): string {
    try {
      // 보안 검증: 템플릿 내용 검사
      if (!validateTemplateContent(template)) {
        logger.error('Conditional template contains dangerous patterns');
        throw new Error('Conditional template security validation failed');
      }

      return renderTemplateSafely(template, context as Record<string, any>);
    } catch (error) {
      logger.error(`Failed to apply conditional logic: ${error}`);
      return template;
    }
  }

  /**
   * @API:VALIDATE-CONTEXT-001 컨텍스트 데이터 검증
   *
   * @param context 검증할 컨텍스트
   * @param schema 검증 스키마
   * @returns 검증 결과
   */
  validateTemplateContext(
    context: TemplateContext,
    schema: ContextSchema
  ): ValidationResult {
    return validateTemplateContext(context, schema);
  }

  /**
   * @API:MERGE-CONTEXTS-001 다중 컨텍스트 병합
   *
   * 우선순위 기반 오버라이드 (첫 번째 인자가 최고 우선순위)
   */
  mergeTemplateContexts(
    primary: TemplateContext,
    ...secondary: TemplateContext[]
  ): TemplateContext {
    return mergeTemplateContexts(primary, ...secondary);
  }

  /**
   * @TASK:PROCESS-DIRECTORY-RECURSIVE-001 재귀적 디렉토리 처리 (보안 강화)
   */
  private async processDirectoryRecursive(
    sourceDir: string,
    targetDir: string,
    context: TemplateContext,
    options: ProcessingOptions,
    state: ProcessingState
  ): Promise<void> {
    try {
      // 소스 디렉토리 보안 검증
      if (!this.validateDirectoryPath(sourceDir)) {
        throw new Error(`Invalid source directory path: ${sourceDir}`);
      }

      const entries = await fs.readdir(sourceDir, { withFileTypes: true });

      for (const entry of entries) {
        // 파일/디렉토리명 보안 검증
        if (!this.validateFileName(entry.name)) {
          logger.warn(`Skipping dangerous file name: ${entry.name}`);
          continue;
        }

        const sourcePath = path.join(sourceDir, entry.name);

        // 경로 정규화 및 검증
        const normalizedSourcePath = path.resolve(sourcePath);
        const normalizedSourceDir = path.resolve(sourceDir);

        // Path traversal 공격 방어
        if (!normalizedSourcePath.startsWith(normalizedSourceDir)) {
          logger.error(`Path traversal attempt detected: ${entry.name}`);
          this.addProcessError(state, sourcePath, new Error('Path traversal attack blocked'));
          continue;
        }

        const targetPath = this.getTargetPath(
          targetDir,
          entry.name,
          context,
          options
        );

        // 타겟 경로 보안 검증
        if (!this.validateTargetPath(targetPath, targetDir)) {
          logger.error(`Invalid target path: ${targetPath}`);
          this.addProcessError(state, sourcePath, new Error('Invalid target path blocked'));
          continue;
        }

        if (entry.isDirectory()) {
          await fs.mkdir(targetPath, { recursive: true });
          await this.processDirectoryRecursive(
            sourcePath,
            targetPath,
            context,
            options,
            state
          );
        } else {
          await this.processFileEntry(
            sourcePath,
            targetPath,
            context,
            options,
            state
          );
        }
      }
    } catch (error) {
      this.addProcessError(state, sourceDir, error);
    }
  }

  /**
   * @TASK:GET-TARGET-PATH-001 타겟 경로 계산
   */
  private getTargetPath(
    targetDir: string,
    fileName: string,
    context: TemplateContext,
    options: ProcessingOptions
  ): string {
    if (options.fileNameTemplating) {
      const processedName = this.expandTemplateVariables(fileName, context);
      return path.join(targetDir, processedName);
    }
    return path.join(targetDir, fileName);
  }

  /**
   * @TASK:PROCESS-FILE-ENTRY-001 파일 항목 처리
   */
  private async processFileEntry(
    sourcePath: string,
    targetPath: string,
    context: TemplateContext,
    options: ProcessingOptions,
    state: ProcessingState
  ): Promise<void> {
    try {
      if (shouldProcessAsTemplate(sourcePath)) {
        await this.processTemplateFile(
          sourcePath,
          targetPath,
          context,
          options
        );
        state.processedFiles.push(targetPath);
      } else {
        await copyBinaryFile(
          sourcePath,
          targetPath,
          options.preserveTimestamps,
          options.overwrite
        );
        state.skippedFiles.push(targetPath);
      }
    } catch (error) {
      this.addProcessError(state, sourcePath, error);
    }
  }

  /**
   * @TASK:ADD-ERROR-001 처리 오류 추가
   */
  private addProcessError(
    state: ProcessingState,
    filePath: string,
    error: unknown
  ): void {
    logger.error(`Error processing ${filePath}: ${error}`);
    const processError: ProcessError = {
      file: filePath,
      error: error instanceof Error ? error.message : String(error),
    };
    state.errors.push(processError);
  }

  /**
   * @SECURITY:PATH-VALIDATION-001 디렉토리 경로 보안 검증
   */
  private validateDirectoryPath(dirPath: string): boolean {
    try {
      // 절대 경로로 정규화
      const normalizedPath = path.resolve(dirPath);

      // 금지된 경로 패턴 검사
      const forbiddenPatterns = [
        /\.\./,           // 상위 디렉토리 접근
        /\/etc\//,        // 시스템 설정 디렉토리
        /\/bin\//,        // 실행 파일 디렉토리
        /\/usr\/bin\//,   // 시스템 바이너리
        /\/var\/log\//,   // 로그 디렉토리
        /\/home\/.*\/\.ssh\//,  // SSH 키 디렉토리
        /C:\\Windows\\/,  // Windows 시스템 디렉토리
        /C:\\Program Files\//,  // Windows 프로그램 디렉토리
      ];

      for (const pattern of forbiddenPatterns) {
        if (pattern.test(normalizedPath)) {
          logger.error(`Forbidden directory path detected: ${normalizedPath}`);
          return false;
        }
      }

      return true;
    } catch (error) {
      logger.error(`Path validation failed: ${error}`);
      return false;
    }
  }

  /**
   * @SECURITY:FILENAME-VALIDATION-001 파일명 보안 검증
   */
  private validateFileName(fileName: string): boolean {
    // 빈 파일명 거부
    if (!fileName || fileName.trim() === '') {
      return false;
    }

    // 위험한 파일명 패턴 검사
    const dangerousPatterns = [
      /\.\./,              // 상위 디렉토리 접근
      /^\.+$/,             // 점으로만 구성된 파일명
      /[\x00-\x1f]/,       // 제어 문자
      /[<>:"|?*]/,         // Windows 금지 문자
      /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$/i, // Windows 예약어
      /\.(exe|bat|cmd|scr|pif|com|vbs|js|jar|sh)$/i, // 실행 파일 확장자
    ];

    for (const pattern of dangerousPatterns) {
      if (pattern.test(fileName)) {
        logger.warn(`Dangerous file name pattern detected: ${fileName}`);
        return false;
      }
    }

    // 파일명 길이 제한 (255자)
    if (fileName.length > 255) {
      logger.warn(`File name too long: ${fileName}`);
      return false;
    }

    return true;
  }

  /**
   * @SECURITY:TARGET-PATH-VALIDATION-001 타겟 경로 보안 검증
   */
  private validateTargetPath(targetPath: string, baseDir: string): boolean {
    try {
      // 경로 정규화
      const normalizedTarget = path.resolve(targetPath);
      const normalizedBase = path.resolve(baseDir);

      // 타겟 경로가 기준 디렉토리 하위에 있는지 확인
      if (!normalizedTarget.startsWith(normalizedBase)) {
        logger.error(`Target path outside base directory: ${targetPath}`);
        return false;
      }

      // 디렉토리 경로 추가 검증
      return this.validateDirectoryPath(path.dirname(normalizedTarget));
    } catch (error) {
      logger.error(`Target path validation failed: ${error}`);
      return false;
    }
  }
}

// Export MoAI template helpers from separate module (임시 비활성화 - 모듈 누락)
// export { MoAITemplateHelpers } from './moai-template-helpers';
