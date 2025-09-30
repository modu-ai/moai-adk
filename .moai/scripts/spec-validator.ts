#!/usr/bin/env tsx
/**
 * @FEATURE:SPEC-VALIDATOR-001 | Chain: @REQ:VALIDATION-001 -> @DESIGN:SPEC-VALIDATOR-001 -> @TASK:VALIDATOR-001 -> @TEST:VALIDATOR-001
 * Related: @API:VALIDATION-001, @DATA:VALIDATION-001
 *
 * SPEC 문서 검증 스크립트
 * - SPEC 구조 검증
 * - TAG Catalog 일치성 확인
 * - EARS 요구사항 형식 검증
 */

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';

interface SpecValidatorOptions {
  spec?: string;
  all?: boolean;
  fix?: boolean;
  strict?: boolean;
}

interface ValidationRule {
  name: string;
  description: string;
  severity: 'error' | 'warning' | 'info';
  check: (content: string, metadata?: any) => boolean;
  message: string;
  fix?: (content: string) => string;
}

interface ValidationResult {
  rule: string;
  severity: 'error' | 'warning' | 'info';
  message: string;
  line?: number;
  column?: number;
  fixable: boolean;
}

interface SpecValidationReport {
  specId: string;
  filePath: string;
  valid: boolean;
  errors: ValidationResult[];
  warnings: ValidationResult[];
  infos: ValidationResult[];
  stats: {
    totalLines: number;
    wordCount: number;
    estimatedReadTime: number;
  };
}

const VALIDATION_RULES: ValidationRule[] = [
  {
    name: 'has-title',
    description: 'SPEC 제목이 있는지 확인',
    severity: 'error',
    check: (content: string) => /^# SPEC-\d+:.+/.test(content.trim()),
    message: 'SPEC 제목이 없거나 형식이 잘못되었습니다. (예: # SPEC-001: 제목)',
    fix: (content: string) => {
      if (!content.trim().startsWith('#')) {
        return `# SPEC-001: 제목을 입력하세요\n\n${content}`;
      }
      return content;
    }
  },
  {
    name: 'has-req-section',
    description: '@REQ 섹션이 있는지 확인',
    severity: 'error',
    check: (content: string) => /@REQ:[A-Z]+-\d+/.test(content),
    message: '@REQ 섹션이 없습니다. 요구사항 섹션은 필수입니다.',
    fix: (content: string) => {
      if (!/@REQ:[A-Z]+-\d+/.test(content)) {
        const insertPoint = content.indexOf('\n\n') + 2 || content.length;
        const reqSection = `\n## @REQ:FEATURE-001 요구사항\n\n### 배경\n여기에 배경을 작성하세요.\n\n### 문제\n여기에 문제를 작성하세요.\n\n`;
        return content.slice(0, insertPoint) + reqSection + content.slice(insertPoint);
      }
      return content;
    }
  },
  {
    name: 'has-design-section',
    description: '@DESIGN 섹션이 있는지 확인',
    severity: 'error',
    check: (content: string) => /@DESIGN:[A-Z]+-\d+/.test(content),
    message: '@DESIGN 섹션이 없습니다. 설계 섹션은 필수입니다.'
  },
  {
    name: 'has-task-section',
    description: '@TASK 섹션이 있는지 확인',
    severity: 'error',
    check: (content: string) => /@TASK:[A-Z]+-\d+/.test(content),
    message: '@TASK 섹션이 없습니다. 구현 섹션은 필수입니다.'
  },
  {
    name: 'has-test-section',
    description: '@TEST 섹션이 있는지 확인',
    severity: 'error',
    check: (content: string) => /@TEST:[A-Z]+-\d+/.test(content),
    message: '@TEST 섹션이 없습니다. 테스트 섹션은 필수입니다.'
  },
  {
    name: 'has-acceptance-criteria',
    description: '승인 기준이 있는지 확인',
    severity: 'warning',
    check: (content: string) => /승인 기준|Acceptance Criteria/i.test(content),
    message: '승인 기준(Acceptance Criteria)이 명시되지 않았습니다.'
  },
  {
    name: 'has-test-scenarios',
    description: '테스트 시나리오가 있는지 확인',
    severity: 'warning',
    check: (content: string) => /테스트 시나리오|Test Scenarios/i.test(content),
    message: '테스트 시나리오가 명시되지 않았습니다.'
  },
  {
    name: 'has-implementation-checklist',
    description: '구현 체크리스트가 있는지 확인',
    severity: 'warning',
    check: (content: string) => /- \[ \]/.test(content),
    message: '구현 체크리스트가 없습니다. TDD 단계별 체크리스트를 추가하세요.'
  },
  {
    name: 'proper-tag-format',
    description: 'TAG 형식이 올바른지 확인',
    severity: 'error',
    check: (content: string) => {
      const tags = content.match(/@[A-Z]+:[A-Z]+-\d+/g) || [];
      return tags.length >= 4; // REQ, DESIGN, TASK, TEST 최소 4개
    },
    message: '16-Core TAG 형식이 올바르지 않습니다. @TYPE:CATEGORY-NNN 형식을 사용하세요.'
  },
  {
    name: 'sufficient-content-length',
    description: '충분한 내용이 있는지 확인',
    severity: 'warning',
    check: (content: string) => content.trim().length > 500,
    message: 'SPEC 내용이 너무 짧습니다. 더 상세한 설명이 필요합니다.'
  },
  {
    name: 'no-placeholder-text',
    description: '플레이스홀더 텍스트가 남아있는지 확인',
    severity: 'info',
    check: (content: string) => !/여기에.*작성하세요|TODO|FIXME/i.test(content),
    message: '플레이스홀더 텍스트가 남아있습니다. 실제 내용으로 교체하세요.'
  },
  {
    name: 'has-estimated-time',
    description: '예상 구현 시간이 명시되었는지 확인',
    severity: 'info',
    check: (content: string) => /예상.*시간|estimated.*hour/i.test(content),
    message: '예상 구현 시간이 명시되지 않았습니다.'
  }
];

async function readSpecFile(specPath: string): Promise<{ content: string; metadata?: any }> {
  const content = await fs.readFile(specPath, 'utf-8');

  let metadata;
  try {
    const metadataPath = path.join(path.dirname(specPath), 'metadata.json');
    const metadataContent = await fs.readFile(metadataPath, 'utf-8');
    metadata = JSON.parse(metadataContent);
  } catch {
    // 메타데이터 파일이 없으면 무시
  }

  return { content, metadata };
}

function calculateStats(content: string): { totalLines: number; wordCount: number; estimatedReadTime: number } {
  const lines = content.split('\n').length;
  const words = content.split(/\s+/).filter(word => word.length > 0).length;
  const estimatedReadTime = Math.ceil(words / 200); // 분당 200단어 기준

  return {
    totalLines: lines,
    wordCount: words,
    estimatedReadTime
  };
}

function validateSpec(content: string, metadata?: any, strict: boolean = false): ValidationResult[] {
  const results: ValidationResult[] = [];

  for (const rule of VALIDATION_RULES) {
    // strict 모드가 아니면 info 레벨은 건너뛰기
    if (!strict && rule.severity === 'info') {
      continue;
    }

    const isValid = rule.check(content, metadata);

    if (!isValid) {
      results.push({
        rule: rule.name,
        severity: rule.severity,
        message: rule.message,
        fixable: !!rule.fix
      });
    }
  }

  return results;
}

function applyFixes(content: string, results: ValidationResult[]): string {
  let fixedContent = content;

  for (const result of results) {
    if (result.fixable) {
      const rule = VALIDATION_RULES.find(r => r.name === result.rule);
      if (rule?.fix) {
        fixedContent = rule.fix(fixedContent);
      }
    }
  }

  return fixedContent;
}

async function validateSingleSpec(specId: string, options: SpecValidatorOptions): Promise<SpecValidationReport> {
  const specDir = path.join('.moai/specs', specId);
  const specFile = path.join(specDir, 'spec.md');

  try {
    const { content, metadata } = await readSpecFile(specFile);
    const results = validateSpec(content, metadata, options.strict);

    const errors = results.filter(r => r.severity === 'error');
    const warnings = results.filter(r => r.severity === 'warning');
    const infos = results.filter(r => r.severity === 'info');

    const stats = calculateStats(content);
    const isValid = errors.length === 0;

    // 자동 수정 적용
    if (options.fix && results.some(r => r.fixable)) {
      const fixedContent = applyFixes(content, results);
      await fs.writeFile(specFile, fixedContent);
    }

    return {
      specId,
      filePath: specFile,
      valid: isValid,
      errors,
      warnings,
      infos,
      stats
    };

  } catch (error) {
    return {
      specId,
      filePath: specFile,
      valid: false,
      errors: [{
        rule: 'file-access',
        severity: 'error',
        message: `SPEC 파일을 읽을 수 없습니다: ${error.message}`,
        fixable: false
      }],
      warnings: [],
      infos: [],
      stats: { totalLines: 0, wordCount: 0, estimatedReadTime: 0 }
    };
  }
}

async function findAllSpecs(): Promise<string[]> {
  const specsDir = '.moai/specs';

  try {
    const entries = await fs.readdir(specsDir, { withFileTypes: true });
    return entries
      .filter(entry => entry.isDirectory())
      .map(entry => entry.name)
      .filter(name => name.startsWith('SPEC-'))
      .sort();
  } catch {
    return [];
  }
}

async function validateSpecs(options: SpecValidatorOptions): Promise<{ success: boolean; message: string; reports?: SpecValidationReport[] }> {
  try {
    let specsToValidate: string[] = [];

    if (options.all) {
      specsToValidate = await findAllSpecs();
      if (specsToValidate.length === 0) {
        return { success: false, message: 'SPEC 파일을 찾을 수 없습니다.' };
      }
    } else if (options.spec) {
      specsToValidate = [options.spec];
    } else {
      return { success: false, message: '--spec 또는 --all 옵션을 지정하세요.' };
    }

    const reports: SpecValidationReport[] = [];

    for (const specId of specsToValidate) {
      const report = await validateSingleSpec(specId, options);
      reports.push(report);
    }

    const totalErrors = reports.reduce((sum, r) => sum + r.errors.length, 0);
    const totalWarnings = reports.reduce((sum, r) => sum + r.warnings.length, 0);
    const validSpecs = reports.filter(r => r.valid).length;

    return {
      success: totalErrors === 0,
      message: `검증 완료: ${validSpecs}/${reports.length}개 SPEC 유효, ${totalErrors}개 오류, ${totalWarnings}개 경고`,
      reports
    };

  } catch (error) {
    return {
      success: false,
      message: `검증 실패: ${error.message}`
    };
  }
}

program
  .name('spec-validator')
  .description('MoAI SPEC 문서 검증')
  .option('-s, --spec <spec-id>', '검증할 SPEC ID (예: SPEC-001)')
  .option('-a, --all', '모든 SPEC 검증')
  .option('-f, --fix', '자동 수정 가능한 문제 해결')
  .option('--strict', '엄격 모드 (info 레벨 검사 포함)')
  .action(async (options: SpecValidatorOptions) => {
    try {
      console.log(chalk.blue('🔍 SPEC 검증 시작...'));

      const result = await validateSpecs(options);

      if (result.success) {
        console.log(chalk.green('✅'), result.message);
      } else {
        console.log(chalk.yellow('⚠️'), result.message);
      }

      if (result.reports) {
        // 상세 리포트 출력
        for (const report of result.reports) {
          if (!report.valid || report.warnings.length > 0) {
            console.log(chalk.cyan(`\n📄 ${report.specId}:`));

            for (const error of report.errors) {
              console.log(chalk.red(`  ❌ [${error.rule}] ${error.message}`));
            }

            for (const warning of report.warnings) {
              console.log(chalk.yellow(`  ⚠️  [${warning.rule}] ${warning.message}`));
            }

            if (options.strict) {
              for (const info of report.infos) {
                console.log(chalk.gray(`  ℹ️  [${info.rule}] ${info.message}`));
              }
            }
          }
        }
      }

      // JSON 출력
      console.log(JSON.stringify({
        success: result.success,
        summary: {
          total: result.reports?.length || 0,
          valid: result.reports?.filter(r => r.valid).length || 0,
          errors: result.reports?.reduce((sum, r) => sum + r.errors.length, 0) || 0,
          warnings: result.reports?.reduce((sum, r) => sum + r.warnings.length, 0) || 0
        },
        reports: result.reports
      }, null, 2));

      process.exit(result.success ? 0 : 1);
    } catch (error) {
      console.error(chalk.red('❌ 스크립트 실행 실패:'), error.message);
      console.log(JSON.stringify({
        success: false,
        error: error.message
      }, null, 2));
      process.exit(1);
    }
  });

program.parse();