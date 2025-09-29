/**
 * @TAG:API:CODE-FIRST-PARSER-001
 * @CHAIN: REQ:TAG-SYSTEM-001 -> DESIGN:CODE-FIRST-001 -> TASK:PARSER-001 -> @TAG:API:CODE-FIRST-PARSER-001
 * @DEPENDS: API:CODE-FIRST-TYPES-001
 * @STATUS: active
 * @CREATED: 2025-09-29
 * @IMMUTABLE
 */

import { promises as fs } from 'node:fs';
import { dirname, extname } from 'node:path';
import {
  TagBlock,
  TagCategory,
  TagStatus,
  TagParseResult,
  TagValidationResult,
  TagChain,
  TagDependencyGraph,
  PerformanceMetrics,
  CodeFirstTagConfig
} from './code-first-types.js';

/**
 * Code-First TAG 파서
 *
 * 코드 파일의 주석에서 TAG 블록을 파싱하고 분석하는 핵심 엔진
 * - 파일 최상단 TAG 블록 추출
 * - 체인 및 의존성 분석
 * - 불변성 검증
 * - 성능 최적화 (ripgrep 연동)
 */
export class CodeFirstTagParser {
  private readonly config: CodeFirstTagConfig;
  private performanceMetrics: PerformanceMetrics;

  /**
   * TAG 블록 패턴 (파일 최상단 주석)
   *
   * 매칭 예시:
   * ```
   * /**
   *  * @TAG:FEATURE:AUTH-001
   *  * @CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001 -> TEST:AUTH-001
   *  * @DEPENDS: FEATURE:USER-001, API:SESSION-001
   *  * @STATUS: active
   *  * @CREATED: 2025-09-29
   *  * @IMMUTABLE
   *  *\/
   * ```
   */
  private static readonly TAG_BLOCK_PATTERNS = {
    // 전체 TAG 블록 매칭
    FULL_BLOCK: /\/\*\*\s*([\s\S]*?)\*\//,

    // 개별 TAG 라인 매칭
    TAG_LINE: /^\s*\*\s*@TAG:([A-Z]+):([A-Z0-9-]+)\s*$/m,
    CHAIN_LINE: /^\s*\*\s*@CHAIN:\s*(.+)\s*$/m,
    DEPENDS_LINE: /^\s*\*\s*@DEPENDS:\s*(.+)\s*$/m,
    STATUS_LINE: /^\s*\*\s*@STATUS:\s*(\w+)\s*$/m,
    CREATED_LINE: /^\s*\*\s*@CREATED:\s*(\d{4}-\d{2}-\d{2})\s*$/m,
    IMMUTABLE_LINE: /^\s*\*\s*@IMMUTABLE\s*$/m,

    // 체인 파싱
    CHAIN_ARROW: /\s*->\s*/,
    TAG_REFERENCE: /([A-Z]+):([A-Z0-9-]+)/g,

    // 의존성 파싱
    DEPENDENCY_SEPARATOR: /,\s*/
  };

  constructor(config?: Partial<CodeFirstTagConfig>) {
    this.config = {
      scanExtensions: ['.ts', '.tsx', '.js', '.jsx', '.py', '.md'],
      excludeDirectories: ['node_modules', '.git', 'dist', 'build', '__pycache__'],
      tagBlockStartPattern: '/**',
      tagBlockEndPattern: '*/',
      enforceImmutability: true,
      enablePerformanceTracking: true,
      ...config
    };

    this.performanceMetrics = {
      searchTime: 0,
      parseTime: 0,
      memoryUsage: 0,
      filesScanned: 0,
      tagsFound: 0
    };
  }

  /**
   * 파일에서 TAG 블록 파싱
   */
  async parseFile(filePath: string): Promise<TagParseResult> {
    const startTime = performance.now();

    try {
      const content = await fs.readFile(filePath, 'utf-8');
      const result = this.parseContent(content, filePath);

      if (this.config.enablePerformanceTracking) {
        this.performanceMetrics.parseTime += performance.now() - startTime;
        this.performanceMetrics.filesScanned++;
        if (result.success) {
          this.performanceMetrics.tagsFound++;
        }
      }

      return result;
    } catch (error) {
      return {
        success: false,
        error: `파일 읽기 실패: ${error.message}`,
        warnings: []
      };
    }
  }

  /**
   * 파일 내용에서 TAG 블록 파싱
   */
  parseContent(content: string, filePath: string): TagParseResult {
    const warnings: string[] = [];

    // 1. TAG 블록 추출 (파일 최상단에서만)
    const blockMatch = this.extractTagBlock(content);
    if (!blockMatch) {
      return {
        success: false,
        error: 'TAG 블록이 발견되지 않았습니다',
        warnings
      };
    }

    const blockContent = blockMatch.content;
    const lineNumber = blockMatch.lineNumber;

    // 2. 메인 TAG 파싱
    const tagMatch = CodeFirstTagParser.TAG_BLOCK_PATTERNS.TAG_LINE.exec(blockContent);
    if (!tagMatch) {
      return {
        success: false,
        error: '@TAG 라인이 발견되지 않았습니다',
        warnings
      };
    }

    const [, categoryStr, domainId] = tagMatch;
    const category = categoryStr as TagCategory;

    // 3. 카테고리 유효성 검사
    if (!Object.values(TagCategory).includes(category)) {
      return {
        success: false,
        error: `유효하지 않은 TAG 카테고리: ${categoryStr}`,
        warnings
      };
    }

    const tag = `@${categoryStr}:${domainId}`;

    // 4. 선택적 필드 파싱
    const chain = this.parseChain(blockContent, warnings);
    const depends = this.parseDepends(blockContent, warnings);
    const status = this.parseStatus(blockContent, warnings);
    const created = this.parseCreated(blockContent, warnings);
    const immutable = this.parseImmutable(blockContent);

    // 5. TAG 블록 구성
    const tagBlock: TagBlock = {
      tag,
      category,
      domainId,
      chain,
      depends,
      status,
      created,
      immutable,
      filePath,
      lineNumber
    };

    // 6. 유효성 검증
    const validation = this.validateTagBlock(tagBlock);
    if (!validation.valid) {
      return {
        success: false,
        error: `TAG 검증 실패: ${validation.errors.join(', ')}`,
        warnings: [...warnings, ...validation.warnings]
      };
    }

    return {
      success: true,
      tagBlock,
      warnings: [...warnings, ...validation.warnings]
    };
  }

  /**
   * 파일 최상단에서 TAG 블록 추출
   */
  private extractTagBlock(content: string): { content: string; lineNumber: number } | null {
    // 파일 시작부분에서 첫 번째 주석 블록 찾기
    const lines = content.split('\n');
    let inBlock = false;
    let blockLines: string[] = [];
    let startLineNumber = 0;

    for (let i = 0; i < Math.min(lines.length, 50); i++) { // 첫 50줄만 검사
      const line = lines[i].trim();

      // 빈 줄이나 shebang 무시
      if (!line || line.startsWith('#!')) {
        continue;
      }

      // TAG 블록 시작
      if (line.startsWith('/**') && !inBlock) {
        inBlock = true;
        blockLines = [line];
        startLineNumber = i + 1;
        continue;
      }

      // TAG 블록 내부
      if (inBlock) {
        blockLines.push(line);

        // TAG 블록 종료
        if (line.endsWith('*/')) {
          const blockContent = blockLines.join('\n');

          // @TAG가 포함된 블록인지 확인
          if (blockContent.includes('@TAG:')) {
            return {
              content: blockContent,
              lineNumber: startLineNumber
            };
          }

          // @TAG가 없으면 리셋하고 계속
          inBlock = false;
          blockLines = [];
          continue;
        }
      }

      // TAG 블록이 아닌 코드 시작되면 중단
      if (!inBlock && line && !line.startsWith('//') && !line.startsWith('/*')) {
        break;
      }
    }

    return null;
  }

  /**
   * 체인 파싱 (REQ:AUTH-001 -> DESIGN:AUTH-001 -> ...)
   */
  private parseChain(blockContent: string, warnings: string[]): string[] | undefined {
    const chainMatch = CodeFirstTagParser.TAG_BLOCK_PATTERNS.CHAIN_LINE.exec(blockContent);
    if (!chainMatch) {
      return undefined;
    }

    const chainStr = chainMatch[1];
    const chainParts = chainStr.split(CodeFirstTagParser.TAG_BLOCK_PATTERNS.CHAIN_ARROW);

    const chain = chainParts
      .map(part => part.trim())
      .filter(part => part.length > 0)
      .map(part => part.startsWith('@') ? part : `@${part}`);

    // 체인 유효성 검사
    for (const chainTag of chain) {
      if (!CodeFirstTagParser.TAG_BLOCK_PATTERNS.TAG_REFERENCE.test(chainTag)) {
        warnings.push(`잘못된 체인 TAG 형식: ${chainTag}`);
      }
    }

    return chain.length > 0 ? chain : undefined;
  }

  /**
   * 의존성 파싱 (FEATURE:USER-001, API:SESSION-001)
   */
  private parseDepends(blockContent: string, warnings: string[]): string[] | undefined {
    const dependsMatch = CodeFirstTagParser.TAG_BLOCK_PATTERNS.DEPENDS_LINE.exec(blockContent);
    if (!dependsMatch) {
      return undefined;
    }

    if (dependsMatch[1].trim().toLowerCase() === 'none') {
      return [];
    }

    const dependsStr = dependsMatch[1];
    const dependsParts = dependsStr.split(CodeFirstTagParser.TAG_BLOCK_PATTERNS.DEPENDENCY_SEPARATOR);

    const depends = dependsParts
      .map(part => part.trim())
      .filter(part => part.length > 0)
      .map(part => part.startsWith('@') ? part : `@${part}`);

    // 의존성 유효성 검사
    for (const dependTag of depends) {
      if (!CodeFirstTagParser.TAG_BLOCK_PATTERNS.TAG_REFERENCE.test(dependTag)) {
        warnings.push(`잘못된 의존성 TAG 형식: ${dependTag}`);
      }
    }

    return depends.length > 0 ? depends : undefined;
  }

  /**
   * 상태 파싱
   */
  private parseStatus(blockContent: string, warnings: string[]): TagStatus {
    const statusMatch = CodeFirstTagParser.TAG_BLOCK_PATTERNS.STATUS_LINE.exec(blockContent);
    if (!statusMatch) {
      warnings.push('STATUS가 명시되지 않아 기본값(active) 사용');
      return TagStatus.ACTIVE;
    }

    const statusStr = statusMatch[1].toLowerCase();
    const status = Object.values(TagStatus).find(s => s === statusStr);

    if (!status) {
      warnings.push(`알 수 없는 STATUS: ${statusStr}, 기본값(active) 사용`);
      return TagStatus.ACTIVE;
    }

    return status;
  }

  /**
   * 생성 날짜 파싱
   */
  private parseCreated(blockContent: string, warnings: string[]): string {
    const createdMatch = CodeFirstTagParser.TAG_BLOCK_PATTERNS.CREATED_LINE.exec(blockContent);
    if (!createdMatch) {
      const today = new Date().toISOString().split('T')[0];
      warnings.push(`CREATED가 명시되지 않아 오늘 날짜(${today}) 사용`);
      return today;
    }

    return createdMatch[1];
  }

  /**
   * 불변성 마커 파싱
   */
  private parseImmutable(blockContent: string): boolean {
    return CodeFirstTagParser.TAG_BLOCK_PATTERNS.IMMUTABLE_LINE.test(blockContent);
  }

  /**
   * TAG 블록 유효성 검증
   */
  private validateTagBlock(tagBlock: TagBlock): TagValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];
    const suggestions: string[] = [];

    // 1. 필수 필드 검증
    if (!tagBlock.tag) {
      errors.push('TAG ID가 비어있습니다');
    }

    if (!tagBlock.category) {
      errors.push('TAG 카테고리가 비어있습니다');
    }

    if (!tagBlock.domainId) {
      errors.push('도메인 ID가 비어있습니다');
    }

    // 2. 도메인 ID 형식 검증 (DOMAIN-NNN)
    if (tagBlock.domainId && !/^[A-Z0-9-]+-\d{3,}$/.test(tagBlock.domainId)) {
      warnings.push(`도메인 ID 형식을 확인하세요: ${tagBlock.domainId} (권장: DOMAIN-001)`);
    }

    // 3. 체인 유효성 검증
    if (tagBlock.chain && tagBlock.chain.length > 0) {
      // Primary Chain 검증 (REQ -> DESIGN -> TASK -> TEST)
      const primaryCategories = [TagCategory.REQ, TagCategory.DESIGN, TagCategory.TASK, TagCategory.TEST];
      const chainCategories = tagBlock.chain
        .map(chainTag => chainTag.split(':')[1]?.split('-')[0])
        .filter(cat => primaryCategories.includes(cat as TagCategory));

      if (chainCategories.length > 1) {
        // 순서 검증
        for (let i = 0; i < chainCategories.length - 1; i++) {
          const currentIndex = primaryCategories.indexOf(chainCategories[i] as TagCategory);
          const nextIndex = primaryCategories.indexOf(chainCategories[i + 1] as TagCategory);

          if (currentIndex >= nextIndex) {
            warnings.push(`체인 순서가 올바르지 않습니다: ${chainCategories[i]} -> ${chainCategories[i + 1]}`);
          }
        }
      }
    }

    // 4. 불변성 검증
    if (this.config.enforceImmutability && !tagBlock.immutable) {
      suggestions.push('@IMMUTABLE 마커를 추가하여 TAG 불변성을 보장하는 것을 권장합니다');
    }

    // 5. 생성 날짜 검증
    if (tagBlock.created && !/^\d{4}-\d{2}-\d{2}$/.test(tagBlock.created)) {
      warnings.push(`생성 날짜 형식을 확인하세요: ${tagBlock.created} (YYYY-MM-DD)`);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings,
      suggestions
    };
  }

  /**
   * 디렉토리 전체 스캔
   */
  async scanDirectory(dirPath: string): Promise<TagBlock[]> {
    const startTime = performance.now();
    const tagBlocks: TagBlock[] = [];

    try {
      await this.scanDirectoryRecursive(dirPath, tagBlocks);

      if (this.config.enablePerformanceTracking) {
        this.performanceMetrics.searchTime = performance.now() - startTime;
      }

      return tagBlocks;
    } catch (error) {
      console.error(`디렉토리 스캔 실패: ${error.message}`);
      return [];
    }
  }

  /**
   * 재귀적 디렉토리 스캔
   */
  private async scanDirectoryRecursive(dirPath: string, results: TagBlock[]): Promise<void> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = `${dirPath}/${entry.name}`;

        if (entry.isDirectory()) {
          // 제외 디렉토리 검사
          if (!this.config.excludeDirectories.includes(entry.name)) {
            await this.scanDirectoryRecursive(fullPath, results);
          }
        } else if (entry.isFile()) {
          // 스캔 대상 파일 검사
          const ext = extname(entry.name).toLowerCase();
          if (this.config.scanExtensions.includes(ext)) {
            const parseResult = await this.parseFile(fullPath);
            if (parseResult.success && parseResult.tagBlock) {
              results.push(parseResult.tagBlock);
            }
          }
        }
      }
    } catch (error) {
      // 권한 없는 디렉토리는 무시
      if (error.code !== 'EACCES' && error.code !== 'EPERM') {
        console.warn(`디렉토리 스캔 경고: ${dirPath} - ${error.message}`);
      }
    }
  }

  /**
   * 성능 메트릭 반환
   */
  getPerformanceMetrics(): PerformanceMetrics {
    return { ...this.performanceMetrics };
  }

  /**
   * 성능 메트릭 리셋
   */
  resetPerformanceMetrics(): void {
    this.performanceMetrics = {
      searchTime: 0,
      parseTime: 0,
      memoryUsage: 0,
      filesScanned: 0,
      tagsFound: 0
    };
  }

  /**
   * 설정 업데이트
   */
  updateConfig(config: Partial<CodeFirstTagConfig>): void {
    Object.assign(this.config, config);
  }
}