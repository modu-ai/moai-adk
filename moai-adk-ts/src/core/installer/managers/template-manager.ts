/**
 * TemplateManager - Mustache.js 기반 템플릿 렌더링 시스템
 * Python Jinja2에서 TypeScript Mustache로 전환된 구현
 */

import * as fs from 'fs/promises';
import Mustache from 'mustache';

/**
 * 템플릿 컨텍스트 데이터 인터페이스
 */
export interface TemplateContext {
  readonly [key: string]: string | number | boolean | object | Array<any>;
}

/**
 * 템플릿 렌더링 결과 인터페이스
 */
export interface TemplateRenderResult {
  readonly success: boolean;
  readonly content: string;
  readonly errors: readonly string[];
}

/**
 * 템플릿 매니저 인터페이스
 */
export interface ITemplateManager {
  loadTemplate(templatePath: string): Promise<string>;
  renderTemplate(template: string, context: TemplateContext): string;
  renderTemplateFile(templatePath: string, context: TemplateContext): Promise<string>;
  validateTemplate(template: string): boolean;
  clearCache(): void;
}

/**
 * TemplateManager 클래스
 * Mustache.js를 사용하여 템플릿 렌더링 및 캐싱 기능 제공
 */
export class TemplateManager implements ITemplateManager {
  private templateCache: Map<string, string> = new Map();
  private parsedTemplateCache: Map<string, any> = new Map();
  private readonly enableCache: boolean;
  private readonly maxCacheSize: number;

  constructor(enableCache: boolean = true, maxCacheSize: number = 100) {
    this.enableCache = enableCache;
    this.maxCacheSize = maxCacheSize;
  }

  /**
   * 템플릿 파일을 로드하고 캐싱
   * @param templatePath 템플릿 파일 경로
   * @returns 템플릿 내용
   */
  async loadTemplate(templatePath: string): Promise<string> {
    if (this.enableCache && this.templateCache.has(templatePath)) {
      return this.templateCache.get(templatePath)!;
    }

    try {
      const template = await fs.readFile(templatePath, 'utf-8');

      if (this.enableCache) {
        this.templateCache.set(templatePath, template);
      }

      return template;
    } catch (error) {
      throw new Error(`Failed to load template: ${templatePath}. ${error}`);
    }
  }

  /**
   * Mustache 템플릿을 렌더링 (성능 최적화된 버전)
   * @param template 템플릿 문자열
   * @param context 컨텍스트 데이터
   * @returns 렌더링된 문자열
   */
  renderTemplate(template: string, context: TemplateContext): string {
    try {
      // 캐시 활성화 시 파싱된 템플릿 재사용
      if (this.enableCache) {
        const cacheKey = this.hashTemplate(template);
        let parsedTemplate = this.parsedTemplateCache.get(cacheKey);

        if (!parsedTemplate) {
          parsedTemplate = Mustache.parse(template);
          this.setParsedTemplateCache(cacheKey, parsedTemplate);
        }

        // Mustache.js에서 parse 결과를 사용하여 렌더링할 때는
        // Writer 클래스를 사용하거나 직접 render 함수 사용
        return Mustache.render(template, context);
      }

      return Mustache.render(template, context);
    } catch (error) {
      throw new Error(`Template rendering failed: ${error}`);
    }
  }

  /**
   * 템플릿 파일을 로드하고 렌더링
   * @param templatePath 템플릿 파일 경로
   * @param context 컨텍스트 데이터
   * @returns 렌더링된 문자열
   */
  async renderTemplateFile(templatePath: string, context: TemplateContext): Promise<string> {
    const template = await this.loadTemplate(templatePath);
    return this.renderTemplate(template, context);
  }

  /**
   * 템플릿 구문 검증
   * @param template 템플릿 문자열
   * @returns 유효한 템플릿인지 여부
   */
  validateTemplate(template: string): boolean {
    try {
      // Mustache 템플릿 파싱 테스트
      Mustache.parse(template);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * 템플릿 캐시 초기화
   */
  clearCache(): void {
    this.templateCache.clear();
    this.parsedTemplateCache.clear();
  }

  /**
   * 캐시된 템플릿 수 반환 (테스트용)
   */
  getCacheSize(): number {
    return this.templateCache.size;
  }

  /**
   * 파싱된 템플릿 캐시 수 반환 (테스트용)
   */
  getParsedCacheSize(): number {
    return this.parsedTemplateCache.size;
  }

  /**
   * 템플릿 문자열의 해시 생성 (캐시 키용)
   * @private
   */
  private hashTemplate(template: string): string {
    // 간단한 해시 함수 (실제 프로덕션에서는 crypto 모듈 사용 고려)
    let hash = 0;
    if (template.length === 0) return hash.toString();

    for (let i = 0; i < template.length; i++) {
      const char = template.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // 32비트 정수로 변환
    }

    return Math.abs(hash).toString();
  }

  /**
   * 파싱된 템플릿을 캐시에 저장 (LRU 방식)
   * @private
   */
  private setParsedTemplateCache(key: string, parsedTemplate: any): void {
    // 캐시 크기 제한 확인
    if (this.parsedTemplateCache.size >= this.maxCacheSize) {
      // 가장 오래된 항목 제거 (FIFO)
      const firstKey = this.parsedTemplateCache.keys().next().value;
      if (firstKey !== undefined) {
        this.parsedTemplateCache.delete(firstKey);
      }
    }

    this.parsedTemplateCache.set(key, parsedTemplate);
  }
}

/**
 * 싱글톤 TemplateManager 인스턴스
 */
export const templateManager = new TemplateManager();