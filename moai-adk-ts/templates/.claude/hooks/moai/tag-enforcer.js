/**
 * @TAG:SECURITY:TAG-ENFORCER-001
 * @CHAIN: REQ:TAG-SYSTEM-001 -> DESIGN:CODE-FIRST-001 -> TASK:ENFORCER-001 -> @TAG:SECURITY:TAG-ENFORCER-001
 * @DEPENDS: NONE
 * @STATUS: active
 * @CREATED: 2025-09-29
 * @IMMUTABLE
 */

import fs from 'fs';
import path from 'path';

/**
 * Code-First TAG Enforcer Hook
 *
 * 새로운 Code-First TAG 시스템을 위한 불변성 보장 훅:
 * - @IMMUTABLE 마커가 있는 TAG 블록 수정 차단
 * - 새로운 @TAG:CATEGORY:DOMAIN-ID 형식 검증
 * - @AI-TAG 카테고리 지원
 * - 체인 및 의존성 검증
 */

/**
 * Code-First TAG 패턴
 */
const CODE_FIRST_PATTERNS = {
  // 전체 TAG 블록 매칭 (파일 최상단)
  TAG_BLOCK: /^\/\*\*\s*([\s\S]*?)\*\//m,

  // 핵심 TAG 라인들
  MAIN_TAG: /^\s*\*\s*@TAG:([A-Z]+):([A-Z0-9-]+)\s*$/m,
  CHAIN_LINE: /^\s*\*\s*@CHAIN:\s*(.+)\s*$/m,
  DEPENDS_LINE: /^\s*\*\s*@DEPENDS:\s*(.+)\s*$/m,
  STATUS_LINE: /^\s*\*\s*@STATUS:\s*(\w+)\s*$/m,
  CREATED_LINE: /^\s*\*\s*@CREATED:\s*(\d{4}-\d{2}-\d{2})\s*$/m,
  IMMUTABLE_MARKER: /^\s*\*\s*@IMMUTABLE\s*$/m,

  // TAG 참조
  TAG_REFERENCE: /@([A-Z]+):([A-Z0-9-]+)/g
};

/**
 * 8-Core TAG 카테고리 (16-Core에서 단순화)
 */
const VALID_CATEGORIES = {
  // Lifecycle (필수 체인)
  lifecycle: ['SPEC', 'REQ', 'DESIGN', 'TASK', 'TEST'],

  // Implementation (선택적)
  implementation: ['FEATURE', 'API', 'FIX']
};

class CodeFirstTAGEnforcer {
  name = 'tag-enforcer';

  /**
   * 새로운 Code-First TAG 불변성 검사 실행
   */
  async execute(input) {
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
      const immutabilityCheck = this.checkImmutability(oldContent, newContent, filePath);
      if (immutabilityCheck.violated) {
        return {
          success: false,
          blocked: true,
          message: `🚫 @IMMUTABLE TAG 수정 금지: ${immutabilityCheck.violationDetails}`,
          suggestions: this.generateImmutabilityHelp(immutabilityCheck),
          exitCode: 2
        };
      }

      // 4. 새 TAG 블록 유효성 검증
      const validation = this.validateCodeFirstTag(newContent);
      if (!validation.isValid) {
        return {
          success: false,
          blocked: true,
          message: `🏷️ Code-First TAG 검증 실패: ${validation.violations.join(', ')}`,
          suggestions: this.generateTagSuggestions(filePath, newContent),
          exitCode: 2
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
          : `📝 TAG 블록이 없는 파일 (권장사항)`
      };

    } catch (error) {
      // 오류 발생 시 블록하지 않고 경고만 출력
      console.error(`TAG Enforcer 경고: ${error.message}`);
      return { success: true };
    }
  }

  /**
   * 파일 쓰기 작업 확인
   */
  isWriteOperation(toolName) {
    return ['Write', 'Edit', 'MultiEdit', 'NotebookEdit'].includes(toolName);
  }

  /**
   * 도구 입력에서 파일 경로 추출
   */
  extractFilePath(toolInput) {
    return toolInput.file_path || toolInput.filePath || toolInput.notebook_path || null;
  }

  /**
   * 도구 입력에서 파일 내용 추출
   */
  extractFileContent(toolInput) {
    if (toolInput.content) return toolInput.content;
    if (toolInput.new_string) return toolInput.new_string;
    if (toolInput.new_source) return toolInput.new_source;
    if (toolInput.edits && Array.isArray(toolInput.edits)) {
      return toolInput.edits.map(edit => edit.new_string).join('\n');
    }
    return '';
  }

  /**
   * TAG 검증 대상 파일인지 확인
   */
  shouldEnforceTags(filePath) {
    const enforceExtensions = ['.ts', '.tsx', '.js', '.jsx', '.py', '.md', '.go', '.rs', '.java', '.cpp', '.hpp'];
    const ext = path.extname(filePath);

    // 테스트 파일은 제외 (다른 TAG 규칙 적용)
    if (filePath.includes('test') || filePath.includes('spec') || filePath.includes('__test__')) {
      return false;
    }

    // node_modules, .git 등 제외
    if (filePath.includes('node_modules') || filePath.includes('.git') || filePath.includes('dist') || filePath.includes('build')) {
      return false;
    }

    return enforceExtensions.includes(ext);
  }

  /**
   * 기존 파일 내용 읽기
   */
  async getOriginalFileContent(filePath) {
    try {
      return await fs.promises.readFile(filePath, 'utf-8');
    } catch (error) {
      // 새 파일인 경우 빈 문자열 반환
      return '';
    }
  }

  /**
   * @IMMUTABLE TAG 블록 수정 검사 (핵심 불변성 보장)
   */
  checkImmutability(oldContent, newContent, filePath) {
    // 기존 파일이 없으면 새 파일이므로 통과
    if (!oldContent) {
      return { violated: false };
    }

    // 1. 기존 파일에서 @IMMUTABLE TAG 블록 찾기
    const oldTagBlock = this.extractTagBlock(oldContent);
    const newTagBlock = this.extractTagBlock(newContent);

    // 기존에 TAG 블록이 없었으면 통과
    if (!oldTagBlock) {
      return { violated: false };
    }

    // 2. @IMMUTABLE 마커 확인
    const wasImmutable = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(oldTagBlock.content);
    if (!wasImmutable) {
      return { violated: false };
    }

    // 3. @IMMUTABLE TAG 블록이 수정되었는지 확인
    if (!newTagBlock) {
      return {
        violated: true,
        modifiedTag: this.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록이 삭제되었습니다'
      };
    }

    // 4. TAG 블록 내용 비교 (공백 및 주석 정규화 후)
    const oldNormalized = this.normalizeTagBlock(oldTagBlock.content);
    const newNormalized = this.normalizeTagBlock(newTagBlock.content);

    if (oldNormalized !== newNormalized) {
      return {
        violated: true,
        modifiedTag: this.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록의 내용이 변경되었습니다'
      };
    }

    return { violated: false };
  }

  /**
   * TAG 블록 추출 (파일 최상단에서만)
   */
  extractTagBlock(content) {
    const lines = content.split('\n');
    let inBlock = false;
    let blockLines = [];
    let startLineNumber = 0;

    for (let i = 0; i < Math.min(lines.length, 30); i++) {
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
          if (CODE_FIRST_PATTERNS.MAIN_TAG.test(blockContent)) {
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
   * TAG 블록에서 메인 TAG 추출
   */
  extractMainTag(blockContent) {
    const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    return match ? `@${match[1]}:${match[2]}` : 'UNKNOWN';
  }

  /**
   * TAG 블록 정규화 (비교용)
   */
  normalizeTagBlock(blockContent) {
    return blockContent
      .split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0)
      .join('\n');
  }

  /**
   * Code-First TAG 유효성 검증
   */
  validateCodeFirstTag(content) {
    const violations = [];
    const warnings = [];
    let hasTag = false;

    // 1. TAG 블록 추출
    const tagBlock = this.extractTagBlock(content);
    if (!tagBlock) {
      return {
        isValid: true, // TAG 블록이 없어도 차단하지 않음 (권장사항)
        violations: [],
        warnings: ['파일 최상단에 TAG 블록이 없습니다 (권장사항)'],
        hasTag: false
      };
    }

    hasTag = true;
    const blockContent = tagBlock.content;

    // 2. 메인 TAG 검증
    const tagMatch = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    if (!tagMatch) {
      violations.push('@TAG 라인이 발견되지 않았습니다');
    } else {
      const [, category, domainId] = tagMatch;

      // 카테고리 유효성 검사
      const allValidCategories = [...VALID_CATEGORIES.lifecycle, ...VALID_CATEGORIES.implementation];
      if (!allValidCategories.includes(category)) {
        violations.push(`유효하지 않은 TAG 카테고리: ${category}`);
      }

      // 도메인 ID 형식 검사
      if (!/^[A-Z0-9-]+-\d{3,}$/.test(domainId)) {
        warnings.push(`도메인 ID 형식 권장: ${domainId} -> DOMAIN-001`);
      }
    }

    // 3. 체인 검증
    const chainMatch = CODE_FIRST_PATTERNS.CHAIN_LINE.exec(blockContent);
    if (chainMatch) {
      const chainStr = chainMatch[1];
      const chainTags = chainStr.split(/\s*->\s*/);

      for (const chainTag of chainTags) {
        if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(chainTag.trim())) {
          warnings.push(`체인의 TAG 형식을 확인하세요: ${chainTag.trim()}`);
        }
      }
    }

    // 4. 의존성 검증
    const dependsMatch = CODE_FIRST_PATTERNS.DEPENDS_LINE.exec(blockContent);
    if (dependsMatch) {
      const dependsStr = dependsMatch[1];
      if (dependsStr.trim().toLowerCase() !== 'none') {
        const dependsTags = dependsStr.split(/,\s*/);

        for (const dependTag of dependsTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(dependTag.trim())) {
            warnings.push(`의존성 TAG 형식을 확인하세요: ${dependTag.trim()}`);
          }
        }
      }
    }

    // 5. 상태 검증
    const statusMatch = CODE_FIRST_PATTERNS.STATUS_LINE.exec(blockContent);
    if (statusMatch) {
      const status = statusMatch[1].toLowerCase();
      if (!['active', 'deprecated', 'completed'].includes(status)) {
        warnings.push(`알 수 없는 STATUS: ${status}`);
      }
    }

    // 6. 생성 날짜 검증
    const createdMatch = CODE_FIRST_PATTERNS.CREATED_LINE.exec(blockContent);
    if (createdMatch) {
      const created = createdMatch[1];
      if (!/^\d{4}-\d{2}-\d{2}$/.test(created)) {
        warnings.push(`생성 날짜 형식을 확인하세요: ${created} (YYYY-MM-DD)`);
      }
    }

    // 7. @IMMUTABLE 마커 권장
    if (!CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(blockContent)) {
      warnings.push('@IMMUTABLE 마커를 추가하여 TAG 불변성을 보장하는 것을 권장합니다');
    }

    return {
      isValid: violations.length === 0,
      violations,
      warnings,
      hasTag
    };
  }

  /**
   * @IMMUTABLE 위반 시 도움말 생성
   */
  generateImmutabilityHelp(immutabilityCheck) {
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
      `🔍 수정 시도된 TAG: ${immutabilityCheck.modifiedTag || 'UNKNOWN'}`
    ];

    return help.join('\n');
  }

  /**
   * TAG 제안 생성
   */
  generateTagSuggestions(filePath, content) {
    const fileName = path.basename(filePath, path.extname(filePath));
    const ext = path.extname(filePath);

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
      '• FEATURE, API, FIX: 구현 카테곦0리',
      '',
      '💡 추가 팁:',
      '• TAG 블록은 파일 최상단에 위치',
      '• @IMMUTABLE 마커로 불변성 보장',
      '• 체인으로 관련 TAG들 연결'
    ];

    return suggestions.join('\n');
  }

  /**
   * 파일 확장자에 따른 타입 반환
   */
  getFileType(filePath) {
    const ext = path.extname(filePath);
    switch (ext) {
      case '.ts':
      case '.tsx':
        return 'typescript';
      case '.js':
      case '.jsx':
        return 'javascript';
      case '.py':
        return 'python';
      case '.md':
        return 'markdown';
      case '.go':
        return 'go';
      case '.rs':
        return 'rust';
      case '.java':
        return 'java';
      case '.cpp':
      case '.hpp':
        return 'cpp';
      default:
        return 'unknown';
    }
  }
}

/**
 * 메인 실행 로직
 */
async function main() {
  try {
    let input = '';

    // stdin에서 입력 읽기
    process.stdin.setEncoding('utf8');

    for await (const chunk of process.stdin) {
      input += chunk;
    }

    const parsedInput = input.trim() ? JSON.parse(input) : {};
    const enforcer = new CodeFirstTAGEnforcer();
    const result = await enforcer.execute(parsedInput);

    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      if (result.suggestions) {
        console.error('\n📝 Code-First TAG 가이드:\n' + result.suggestions);
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
    console.error(`Code-First TAG Enforcer 오류: ${error.message}`);
    process.exit(0); // 오류 시 블록하지 않음
  }
}

// 테스트를 위한 export
export { CodeFirstTAGEnforcer, main };

// 직접 실행 시 (ES 모듈 대응)
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(error => {
    console.error(`Code-First TAG Enforcer 치명적 오류: ${error.message}`);
    process.exit(0);
  });
}
      '',
      '💡 추가 팁:',
      '• TAG 블록은 파일 최상단에 위치',
      '• @IMMUTABLE 마커로 불변성 보장',
      '• 체인으로 관련 TAG들 연결'
    ];

    return suggestions.join('\n');
  }

  /**
   * 파일 확장자에 따른 타입 반환
   */
  getFileType(filePath) {
    const ext = path.extname(filePath);
    switch (ext) {
      case '.ts':
      case '.tsx':
        return 'typescript';
      case '.js':
      case '.jsx':
        return 'javascript';
      case '.py':
        return 'python';
      case '.md':
        return 'markdown';
      case '.go':
        return 'go';
      case '.rs':
        return 'rust';
      case '.java':
        return 'java';
      case '.cpp':
      case '.hpp':
        return 'cpp';
      default:
        return 'unknown';
    }
  }
}

/**
 * Main execution
 */
async function main() {
  try {
    let input = '';

    // Read input from stdin
    process.stdin.setEncoding('utf8');

    for await (const chunk of process.stdin) {
      input += chunk;
    }

    const parsedInput = input.trim() ? JSON.parse(input) : {};
    const enforcer = new TAGEnforcer();
    const result = await enforcer.execute(parsedInput);

    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      if (result.suggestions) {
        console.error('\n📝 권장 @TAG 형식:\n' + result.suggestions);
      }
      process.exit(2);
    } else if (!result.success) {
      console.error(`ERROR: ${result.message}`);
      process.exit(result.exitCode || 1);
    } else if (result.message) {
