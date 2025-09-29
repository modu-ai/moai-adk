/**
 * @FEATURE:TEMPLATE-UTILS-001 Template Processing Utilities (Security Enhanced)
 *
 * 템플릿 처리를 위한 유틸리티 함수들 - ReDoS 방어 강화
 * @DESIGN:SEPARATE-CONCERNS-001 TemplateProcessor에서 분리하여 단일 책임 원칙 준수
 * @SECURITY:REDOS-PROTECTION-001 정규식 DoS 공격 방어
 */

import { createSafeRegex } from '../../../utils/regex-security';
import type { TemplateContext } from './template-processor';

/**
 * @TASK:MULTIPLE-FORMATS-001 다중 변수 포맷 처리 (ReDoS 방어)
 *
 * [VAR], ${VAR}, $VAR 형식 처리 - 안전한 정규식 사용
 */
export function processMultipleVariableFormats(
  content: string,
  context: TemplateContext
): string {
  // ReDoS 방어를 위한 안전한 정규식 생성
  const safePattern = createSafeRegex(
    '\\[([\\w]+)\\]|\\$\\{([\\w]+)\\}|\\$([\\w]+)(?![a-zA-Z0-9_])',
    'g'
  );

  if (!safePattern) {
    console.warn('Failed to create safe regex pattern for variable formats');
    return content;
  }

  // 안전성 검사 후 기본 replace 사용 (함수 replacement 지원)
  if (content.length > 50000) {
    console.warn('Content too long for variable replacement');
    return content;
  }

  try {
    return content.replace(
      safePattern,
      (match, bracket, dollarBrace, dollarSimple) => {
        const varName = bracket || dollarBrace || dollarSimple;
        if (varName && varName in context) {
          return String(context[varName]);
        }
        return match;
      }
    );
  } catch (error) {
    console.warn('Variable replacement failed:', error);
    return content;
  }
}

/**
 * @TASK:NESTED-VARIABLES-001 중첩 변수 확장
 *
 * {{PROJECT_{{ENVIRONMENT}}_CONFIG}} 형식 처리
 * Python 구현과 완전 동일한 로직
 */
export function expandNestedVariables(
  content: string,
  context: TemplateContext
): string {
  let result = content;
  let changed = true;
  let iterations = 0;
  const maxIterations = 5; // 무한 루프 방지

  while (changed && iterations < maxIterations) {
    changed = false;
    iterations++;

    // Python과 동일한 중첩 패턴: {{...{{내부변수}}...}} - ReDoS 방어
    const safeNestedPattern = createSafeRegex(
      '\\{\\{([^{}]*?)\\{\\{([^{}]+?)\\}\\}([^{}]*?)\\}\\}',
      'g'
    );

    if (!safeNestedPattern) {
      console.warn('Failed to create safe nested pattern regex');
      break;
    }
    let match;

    while ((match = safeNestedPattern.exec(result)) !== null) {
      const [fullMatch, prefix, innerVar, suffix] = match;

      if (!innerVar) continue;
      const innerVarTrimmed = innerVar.trim();

      // 내부 변수가 컨텍스트에 있는지 확인
      if (innerVarTrimmed in context) {
        changed = true;
        const expandedInner = String(context[innerVarTrimmed]);
        const fullVarName = (prefix + expandedInner + suffix).trim();

        // 확장된 변수명이 컨텍스트에 있으면 바로 치환
        if (fullVarName in context) {
          result = result.replace(fullMatch, String(context[fullVarName]));
        } else {
          // 없으면 일반 변수 형태로 변환
          result = result.replace(fullMatch, `{{${fullVarName}}}`);
        }

        // 패턴이 변경되었으므로 regex를 재시작
        safeNestedPattern.lastIndex = 0;
        break;
      }
    }

    // 더 이상 매치가 없으면 루프 종료
    if (!changed) {
      break;
    }
  }

  return result;
}

/**
 * @TASK:SHOULD-PROCESS-001 템플릿 처리 여부 판단
 */
export function shouldProcessAsTemplate(filePath: string): boolean {
  const binaryExtensions = new Set([
    '.exe',
    '.dll',
    '.so',
    '.dylib',
    '.bin',
    '.dat',
    '.db',
    '.sqlite',
    '.png',
    '.jpg',
    '.jpeg',
    '.gif',
    '.bmp',
    '.ico',
    '.svg',
    '.mp3',
    '.mp4',
    '.avi',
    '.mov',
    '.wav',
    '.pdf',
    '.zip',
    '.tar',
    '.gz',
    '.class',
    '.jar',
    '.pyc',
    '.pyo',
    '.whl',
    '.egg',
  ]);

  const textExtensions = new Set([
    '.md',
    '.json',
    '.yml',
    '.yaml',
    '.txt',
    '.py',
    '.js',
    '.ts',
    '.html',
    '.css',
    '.xml',
    '.ini',
    '.cfg',
    '.conf',
    '.sh',
    '.bat',
    '.ps1',
    '.sql',
    '.env',
    '.gitignore',
    '.dockerignore',
    '.editorconfig',
  ]);

  const ext = getFileExtension(filePath);

  // 바이너리 파일 확장자 제외
  if (binaryExtensions.has(ext)) {
    return false;
  }

  // 텍스트 파일 확장자 포함하거나 확장자가 없는 경우
  return textExtensions.has(ext) || !ext;
}

/**
 * @TASK:FILE-EXISTS-001 파일 존재 여부 확인
 */
export async function fileExists(filePath: string): Promise<boolean> {
  try {
    const { promises: fs } = await import('node:fs');
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

/**
 * @TASK:GET-EXTENSION-001 파일 확장자 추출
 */
function getFileExtension(filePath: string): string {
  const path = require('node:path');
  return path.extname(filePath).toLowerCase();
}

/**
 * @TASK:COPY-BINARY-001 바이너리 파일 복사
 */
export async function copyBinaryFile(
  sourcePath: string,
  targetPath: string,
  preserveTimestamps: boolean,
  overwrite: boolean
): Promise<void> {
  const { promises: fs } = await import('node:fs');

  if (!overwrite && (await fileExists(targetPath))) {
    return;
  }

  await fs.copyFile(sourcePath, targetPath);

  if (preserveTimestamps) {
    const stats = await fs.stat(sourcePath);
    await fs.utimes(targetPath, stats.atime, stats.mtime);
  }
}

/**
 * @TASK:MERGE-CONTEXTS-001 다중 컨텍스트 병합
 *
 * 우선순위 기반 오버라이드 (첫 번째 인자가 최고 우선순위)
 */
export function mergeTemplateContexts(
  primary: TemplateContext,
  ...secondary: TemplateContext[]
): TemplateContext {
  // 모든 컨텍스트를 순서대로 병합 (마지막에 primary 적용)
  const result: Record<string, any> = {};

  // secondary 컨텍스트들을 순서대로 병합 (나중 인덱스가 우선)
  for (const context of secondary) {
    if (context) {
      Object.assign(result, context);
    }
  }

  // primary는 최종적으로 덮어씀 (최고 우선순위)
  Object.assign(result, primary);

  return Object.freeze(result);
}

/**
 * @FEATURE:UNIFIED-SUBSTITUTE-001 통합 변수 치환 시스템
 *
 * Python의 unified_substitute_template_variables 완전 포팅
 * 다양한 템플릿 변수 포맷을 지원하는 통합 치환 시스템
 *
 * 지원하는 포맷:
 * - [VAR] format (square brackets)
 * - {{VAR}} format (double curly braces) - Mustache와 호환
 * - ${VAR} format (dollar curly braces)
 * - $VAR format (simple dollar)
 */
export function unifiedSubstituteTemplateVariables(
  content: string,
  projectContext: TemplateContext
): string {
  try {
    // Python regex pattern과 동일: r'\[(\w+)\]|\{\{(\w+)\}\}|\$\{(\w+)\}|\$(\w+)(?![a-zA-Z0-9_])'
    const pattern =
      /\[(\w+)\]|\{\{(\w+)\}\}|\$\{(\w+)\}|\$(\w+)(?![a-zA-Z0-9_])/g;

    const result = content.replace(
      pattern,
      (match, bracket, doubleCurly, dollarCurly, dollarSimple) => {
        // Python 구현과 동일: 매칭된 그룹에서 변수명 추출
        const varName = bracket || doubleCurly || dollarCurly || dollarSimple;

        if (varName && varName in projectContext) {
          return String(projectContext[varName]);
        } else {
          // Python 구현과 동일: 변수가 없으면 원본 매치 반환
          return match;
        }
      }
    );

    return result;
  } catch (error) {
    // Python 구현과 동일: 오류 시 원본 내용 반환
    console.error(
      `Failed to substitute template variables (unified): ${error}`
    );
    return content;
  }
}

/**
 * @FEATURE:APPLY-PROJECT-CONTEXT-001 프로젝트 컨텍스트 적용
 *
 * Python apply_project_context 메서드의 완전 포팅
 * 파일에 직접 프로젝트 컨텍스트를 적용하는 in-place 처리
 */
export async function applyProjectContext(
  templatePath: string,
  context: TemplateContext
): Promise<boolean> {
  try {
    const { promises: fs } = await import('node:fs');

    if (!(await fileExists(templatePath))) {
      console.warn(`Template file not found: ${templatePath}`);
      return false;
    }

    // Read original content
    const originalContent = await fs.readFile(templatePath, 'utf-8');

    // Apply context substitution using unified method
    const processedContent = unifiedSubstituteTemplateVariables(
      originalContent,
      context
    );

    // Write back if changed
    if (processedContent !== originalContent) {
      await fs.writeFile(templatePath, processedContent, 'utf-8');
      console.info(`Applied project context to: ${templatePath}`);
    }

    return true;
  } catch (error) {
    console.error(
      `Failed to apply project context to ${templatePath}: ${error}`
    );
    return false;
  }
}
