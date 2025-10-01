// @CODE:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md | TEST: __tests__/claude/hooks/tag-enforcer/

/**
 * TAG Enforcer Hook (Refactored)
 * @CODE:REFACTOR-003:API - 코드 내 TAG 블록 불변성 보장 및 8-Core TAG 체계 검증
 *
 * SPEC-REFACTOR-003 요구사항:
 * - 554 LOC → 4개 모듈 분리
 * - @IMMUTABLE 보호 메커니즘 강화
 * - 테스트 커버리지 85% 이상
 * - 모든 기존 테스트 통과
 *
 * 리팩토링 구조:
 * - types.ts: 타입 정의
 * - tag-patterns.ts: 정규식 패턴
 * - tag-validator.ts: TAG 검증 로직
 * - tag-enforcer.ts: 메인 훅 클래스
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import type { HookInput, HookResult, MoAIHook } from '../types';
import { CODE_FIRST_PATTERNS } from './tag-enforcer/tag-patterns';
import { TagValidator } from './tag-enforcer/tag-validator';
import type { ImmutabilityCheck } from './tag-enforcer/types';

/**
 * Code-First TAG Enforcer Hook
 * @CODE:HOOK-004:API: TAG 블록 불변성 검증 API
 */
export class CodeFirstTAGEnforcer implements MoAIHook {
  name = 'tag-enforcer';
  private validator: TagValidator;

  constructor() {
    this.validator = new TagValidator();
  }

  /**
   * 새로운 Code-First TAG 불변성 검사 실행
   */
  async execute(input: HookInput): Promise<HookResult> {
    try {
      // 1. 파일 쓰기 작업인지 확인
      if (!this.isWriteOperation(input.tool_name)) {
        return { success: true };
      }

      const filePath = this.extractFilePath(input.tool_input || {});
      if (!filePath || !this.shouldEnforceTags(filePath)) {
        return { success: true };
      }

      // 2. 기존 파일 내용과 새 내용 추출
      const oldContent = await this.getOriginalFileContent(filePath);
      const newContent = this.extractFileContent(input.tool_input || {});

      // 3. @IMMUTABLE TAG 블록 수정 검사
      const immutabilityCheck = this.checkImmutability(
        oldContent,
        newContent,
        filePath
      );
      if (immutabilityCheck.violated) {
        return {
          success: false,
          blocked: true,
          message: `🚫 @IMMUTABLE TAG 수정 금지: ${immutabilityCheck.violationDetails}`,
          data: {
            suggestions: this.generateImmutabilityHelp(immutabilityCheck),
          },
          exitCode: 2,
        };
      }

      // 4. 새 TAG 블록 유효성 검증
      const validation = this.validator.validateCodeFirstTag(newContent);
      if (!validation.isValid) {
        return {
          success: false,
          blocked: true,
          message: `🏷️ Code-First TAG 검증 실패: ${validation.violations.join(', ')}`,
          data: {
            suggestions: this.generateTagSuggestions(filePath, newContent),
          },
          exitCode: 2,
        };
      }

      // 5. 경고 출력 (차단하지 않음)
      if (validation.warnings.length > 0) {
        console.error(`⚠️ TAG 개선 권장: ${validation.warnings.join(', ')}`);
      }

      return {
        success: true,
        message: validation.hasTag
          ? `✅ Code-First TAG 검증 완료`
          : `📝 TAG 블록이 없는 파일 (권장사항)`,
      };
    } catch (error) {
      // 오류 발생 시 블록하지 않고 경고만 출력
      console.error(
        `TAG Enforcer 경고: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
      return { success: true };
    }
  }

  /**
   * 파일 쓰기 작업 확인
   */
  private isWriteOperation(toolName?: string): boolean {
    return (
      !!toolName &&
      ['Write', 'Edit', 'MultiEdit', 'NotebookEdit'].includes(toolName)
    );
  }

  /**
   * 도구 입력에서 파일 경로 추출
   */
  private extractFilePath(toolInput: Record<string, any>): string | null {
    return (
      toolInput['file_path'] ||
      toolInput['filePath'] ||
      toolInput['notebook_path'] ||
      null
    );
  }

  /**
   * 도구 입력에서 파일 내용 추출
   */
  private extractFileContent(toolInput: Record<string, any>): string {
    if (toolInput['content']) return toolInput['content'];
    if (toolInput['new_string']) return toolInput['new_string'];
    if (toolInput['new_source']) return toolInput['new_source'];
    if (toolInput['edits'] && Array.isArray(toolInput['edits'])) {
      return toolInput['edits'].map((edit: any) => edit['new_string']).join('\n');
    }
    return '';
  }

  /**
   * TAG 검증 대상 파일인지 확인
   */
  private shouldEnforceTags(filePath: string): boolean {
    const enforceExtensions = [
      '.ts',
      '.tsx',
      '.js',
      '.jsx',
      '.py',
      '.md',
      '.go',
      '.rs',
      '.java',
      '.cpp',
      '.hpp',
    ];
    const ext = path.extname(filePath);

    // 테스트 파일은 제외 (다른 TAG 규칙 적용)
    if (
      filePath.includes('test') ||
      filePath.includes('spec') ||
      filePath.includes('__test__')
    ) {
      return false;
    }

    // node_modules, .git 등 제외
    if (
      filePath.includes('node_modules') ||
      filePath.includes('.git') ||
      filePath.includes('dist') ||
      filePath.includes('build')
    ) {
      return false;
    }

    return enforceExtensions.includes(ext);
  }

  /**
   * 기존 파일 내용 읽기
   */
  private async getOriginalFileContent(filePath: string): Promise<string> {
    try {
      return await fs.readFile(filePath, 'utf-8');
    } catch (_error) {
      // 새 파일인 경우 빈 문자열 반환
      return '';
    }
  }

  /**
   * @IMMUTABLE TAG 블록 수정 검사 (핵심 불변성 보장)
   */
  private checkImmutability(
    oldContent: string,
    newContent: string,
    _filePath: string
  ): ImmutabilityCheck {
    // 기존 파일이 없으면 새 파일이므로 통과
    if (!oldContent) {
      return { violated: false };
    }

    // 1. 기존 파일에서 @IMMUTABLE TAG 블록 찾기
    const oldTagBlock = this.validator.extractTagBlock(oldContent);
    const newTagBlock = this.validator.extractTagBlock(newContent);

    // 기존에 TAG 블록이 없었으면 통과
    if (!oldTagBlock) {
      return { violated: false };
    }

    // 2. @IMMUTABLE 마커 확인
    const wasImmutable = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(
      oldTagBlock.content
    );
    if (!wasImmutable) {
      return { violated: false };
    }

    // 3. @IMMUTABLE TAG 블록이 수정되었는지 확인
    if (!newTagBlock) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록이 삭제되었습니다',
      };
    }

    // 4. TAG 블록 내용 비교 (공백 및 주석 정규화 후)
    const oldNormalized = this.validator.normalizeTagBlock(oldTagBlock.content);
    const newNormalized = this.validator.normalizeTagBlock(newTagBlock.content);

    if (oldNormalized !== newNormalized) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록의 내용이 변경되었습니다',
      };
    }

    return { violated: false };
  }

  /**
   * @IMMUTABLE 위반 시 도움말 생성
   */
  private generateImmutabilityHelp(
    immutabilityCheck: ImmutabilityCheck
  ): string {
    const help = [
      '🚫 @IMMUTABLE TAG 수정이 감지되었습니다.',
      '',
      '📋 Code-First TAG 규칙:',
      '• @IMMUTABLE 마커가 있는 TAG 블록은 수정할 수 없습니다',
      '• TAG는 한번 작성되면 불변(immutable)입니다',
      '• 기능 변경 시에는 새로운 TAG를 생성하세요',
      '',
      '✅ 권장 해결 방법:',
      '1. 새로운 TAG ID로 새 기능을 구현하세요',
      '   예: @TAG:FEATURE:AUTH-002',
      '2. 기존 TAG에 @DEPRECATED 마커를 추가하세요',
      '3. 새 TAG에서 이전 TAG를 참조하세요',
      '   예: @REPLACES: FEATURE:AUTH-001',
      '',
      `🔍 수정 시도된 TAG: ${immutabilityCheck.modifiedTag || 'UNKNOWN'}`,
    ];

    return help.join('\n');
  }

  /**
   * TAG 제안 생성
   */
  private generateTagSuggestions(filePath: string, _content: string): string {
    const fileName = path.basename(filePath, path.extname(filePath));

    const suggestions = [
      '📝 Code-First TAG 블록 예시:',
      '',
      '```',
      '/**',
      ` * @TAG:FEATURE:${fileName.toUpperCase()}-001`,
      ` * @CHAIN: REQ:${fileName.toUpperCase()}-001 -> DESIGN:${fileName.toUpperCase()}-001 -> TASK:${fileName.toUpperCase()}-001 -> TEST:${fileName.toUpperCase()}-001`,
      ' * @DEPENDS: NONE',
      ' * @STATUS: active',
      ` * @CREATED: ${new Date().toISOString().split('T')[0]}`,
      ' * @IMMUTABLE',
      ' */',
      '```',
      '',
      '🎯 TAG 카테고리 가이드:',
      '• SPEC, REQ, DESIGN, TASK, TEST: 필수 생명주기',
      '• FEATURE, API, FIX: 구현 카테고리',
      '',
      '💡 추가 팁:',
      '• TAG 블록은 파일 최상단에 위치',
      '• @IMMUTABLE 마커로 불변성 보장',
      '• 체인으로 관련 TAG들 연결',
    ];

    return suggestions.join('\n');
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const enforcer = new CodeFirstTAGEnforcer();
    const result = await enforcer.execute(input);

    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      if (result.data?.['suggestions']) {
        console.error(
          `\n📝 Code-First TAG 가이드:\n${result.data['suggestions']}`
        );
      }
      process.exit(2);
    } else if (!result.success) {
      console.error(`ERROR: ${result.message}`);
      process.exit(result.exitCode || 1);
    } else if (result.message) {
      console.log(result.message);
    }

    process.exit(0);
  } catch (error) {
    console.error(
      `Code-First TAG Enforcer 오류: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(0); // 오류 시 블록하지 않음
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(error => {
    console.error(
      `Code-First TAG Enforcer 치명적 오류: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(0);
  });
}
