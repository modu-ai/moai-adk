/**
 * @FEATURE:TEMPLATE-PROCESSOR-001 Advanced Template Processor for MoAI-ADK
 *
 * Python의 고급 템플릿 처리 로직을 TypeScript로 완전 포팅
 * @TASK:TEMPLATE-PROCESSOR-001 복잡한 템플릿 워크플로우 지원
 * @DESIGN:MOAI-INTEGRATION-001 MoAI 프로젝트 특화 템플릿 기능
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import Mustache from 'mustache';
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
   * @API:EXPAND-VARIABLES-001 고급 변수 확장
   *
   * 다중 포맷 지원 및 중첩 변수 확장
   */
  expandTemplateVariables(content: string, context: TemplateContext): string {
    try {
      // 1단계: 중첩 변수 확장 먼저 처리 ({{PROJECT_{{ENV}}_CONFIG}} 형식)
      let result = expandNestedVariables(content, context);

      // 2단계: Mustache 템플릿 처리 ({{VAR}} 형식)
      result = Mustache.render(result, context);

      // 3단계: 다중 포맷 지원 ([VAR], ${VAR}, $VAR)
      result = processMultipleVariableFormats(result, context);

      return result;
    } catch (error) {
      logger.error(`Failed to expand template variables: ${error}`);
      return content;
    }
  }

  /**
   * @API:CONDITIONAL-LOGIC-001 조건부 로직 적용
   *
   * Mustache 기반 조건부 블록 및 반복 구조 처리
   */
  applyConditionalLogic(template: string, context: TemplateContext): string {
    try {
      return Mustache.render(template, context);
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
   * @TASK:PROCESS-DIRECTORY-RECURSIVE-001 재귀적 디렉토리 처리
   */
  private async processDirectoryRecursive(
    sourceDir: string,
    targetDir: string,
    context: TemplateContext,
    options: ProcessingOptions,
    state: ProcessingState
  ): Promise<void> {
    try {
      const entries = await fs.readdir(sourceDir, { withFileTypes: true });

      for (const entry of entries) {
        const sourcePath = path.join(sourceDir, entry.name);
        const targetPath = this.getTargetPath(
          targetDir,
          entry.name,
          context,
          options
        );

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
}

// Export MoAI template helpers from separate module
export { MoAITemplateHelpers } from './moai-template-helpers';
