/**
 * @REFACTOR:TEMPLATE-PROCESSOR-001 템플릿 변수 치환 유틸리티
 *
 * @TASK:TEMPLATE-SUBSTITUTION-001 템플릿 변수 치환만 담당하는 전용 클래스
 * TRUST-U 원칙: 단일 책임 및 50 LOC 이하 메서드 준수
 */

import { promises as fs } from 'fs';
import { logger } from '../../../../utils/logger';

/**
 * @DESIGN:INTERFACE-001 템플릿 컨텍스트 인터페이스
 */
export interface TemplateContext {
  readonly [key: string]: string;
}

/**
 * @DESIGN:CLASS-001 템플릿 처리 전용 클래스
 *
 * 템플릿 변수 치환 로직을 캡슐화합니다.
 */
export class TemplateProcessor {
  /**
   * @API:SUBSTITUTE-VARIABLES-001 파일의 템플릿 변수 치환
   *
   * @param filePath 치환할 파일 경로
   * @param context 치환 변수들
   */
  async substituteFileVariables(
    filePath: string,
    context: TemplateContext
  ): Promise<void> {
    try {
      const content = await this.readFileContent(filePath);
      const substitutedContent = this.processTemplateVariables(content, context);
      await this.writeFileContent(filePath, substitutedContent);

      logger.debug(`Template variables substituted in ${filePath}`);
    } catch (error) {
      logger.warn(`Failed to substitute template variables in ${filePath}: ${error}`);
    }
  }

  /**
   * @API:SUBSTITUTE-BATCH-001 여러 파일의 템플릿 변수 일괄 치환
   *
   * @param filePaths 치환할 파일 경로들
   * @param context 치환 변수들
   */
  async substituteBatchVariables(
    filePaths: string[],
    context: TemplateContext
  ): Promise<void> {
    const promises = filePaths.map(filePath =>
      this.substituteFileVariables(filePath, context)
    );

    await Promise.allSettled(promises);
  }

  /**
   * @TASK:READ-CONTENT-001 파일 내용 읽기
   */
  private async readFileContent(filePath: string): Promise<string> {
    return await fs.readFile(filePath, 'utf-8');
  }

  /**
   * @TASK:WRITE-CONTENT-001 파일 내용 쓰기
   */
  private async writeFileContent(filePath: string, content: string): Promise<void> {
    await fs.writeFile(filePath, content, 'utf-8');
  }

  /**
   * @TASK:PROCESS-VARIABLES-001 템플릿 변수 처리
   *
   * 더 안전한 템플릿 치환을 위해 정확한 패턴만 치환
   */
  private processTemplateVariables(content: string, context: TemplateContext): string {
    let processedContent = content;

    for (const [key, value] of Object.entries(context)) {
      const pattern = `{{${key}}}`;
      processedContent = processedContent.replace(new RegExp(pattern, 'g'), value);
    }

    return processedContent;
  }
}