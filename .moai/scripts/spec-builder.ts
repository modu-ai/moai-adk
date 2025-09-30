#!/usr/bin/env tsx
/**
 * @FEATURE:SPEC-BUILDER-001 | Chain: @REQ:SPEC-001 -> @DESIGN:SPEC-BUILDER-001 -> @TASK:SPEC-001 -> @TEST:SPEC-001
 * Related: @API:SPEC-001, @DATA:SPEC-001
 *
 * SPEC 문서 생성 스크립트
 * - EARS 방식 요구사항 작성
 * - TAG Catalog 자동 생성
 * - SPEC 템플릿 기반 문서화
 */

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import inquirer from 'inquirer';

interface SpecBuilderOptions {
  id?: string;
  title?: string;
  type?: 'feature' | 'bug' | 'improvement' | 'research';
  priority?: 'critical' | 'high' | 'medium' | 'low';
  interactive?: boolean;
  template?: string;
}

interface SpecMetadata {
  id: string;
  title: string;
  type: string;
  priority: string;
  status: 'draft' | 'review' | 'approved' | 'implemented';
  author: string;
  createdAt: string;
  estimatedHours: number;
  tags: string[];
}

interface SpecContent {
  metadata: SpecMetadata;
  background: string;
  problem: string;
  goals: string[];
  nonGoals: string[];
  constraints: string[];
  acceptance: string[];
  implementation: {
    approach: string;
    risks: string[];
    alternatives: string[];
  };
  testing: {
    strategy: string;
    scenarios: string[];
    coverage: number;
  };
}

const SPEC_TEMPLATES = {
  feature: `# SPEC-{ID}: {TITLE}

## @REQ:{TYPE}-{ID} 요구사항

### 배경 (Background)
{BACKGROUND}

### 해결하려는 문제 (Problem Statement)
{PROBLEM}

### 목표 (Goals)
{GOALS}

### 비목표 (Non-Goals)
{NON_GOALS}

### 제약사항 (Constraints)
{CONSTRAINTS}

## @DESIGN:{TYPE}-{ID} 설계

### 접근 방법 (Approach)
{APPROACH}

### 위험 요소 (Risks)
{RISKS}

### 대안 검토 (Alternatives Considered)
{ALTERNATIVES}

## @TASK:{TYPE}-{ID} 구현

### 승인 기준 (Acceptance Criteria)
{ACCEPTANCE}

### 구현 체크리스트
- [ ] RED: 테스트 작성 및 실패 확인
- [ ] GREEN: 최소 구현으로 테스트 통과
- [ ] REFACTOR: 코드 품질 개선
- [ ] 문서화 완료

## @TEST:{TYPE}-{ID} 테스트

### 테스트 전략
{TEST_STRATEGY}

### 테스트 시나리오
{TEST_SCENARIOS}

### 커버리지 목표
- 단위 테스트: {COVERAGE}%
- 통합 테스트: 필요시
- E2E 테스트: 필요시

---

**메타데이터**
- 우선순위: {PRIORITY}
- 예상 시간: {ESTIMATED_HOURS}시간
- 작성자: {AUTHOR}
- 작성일: {CREATED_AT}
`,

  bug: `# SPEC-{ID}: {TITLE} (버그 수정)

## @REQ:BUG-{ID} 버그 정보

### 현상 (Symptoms)
{BACKGROUND}

### 원인 분석 (Root Cause)
{PROBLEM}

### 영향 범위 (Impact)
{GOALS}

### 재현 단계 (Reproduction Steps)
{CONSTRAINTS}

## @DESIGN:BUG-{ID} 수정 방안

### 수정 접근법 (Fix Approach)
{APPROACH}

### 회귀 위험 (Regression Risks)
{RISKS}

### 대안 방법 (Alternative Fixes)
{ALTERNATIVES}

## @TASK:BUG-{ID} 구현

### 수정 완료 기준 (Done Criteria)
{ACCEPTANCE}

### 수정 체크리스트
- [ ] 버그 재현 테스트 작성
- [ ] 수정 구현
- [ ] 회귀 테스트 실행
- [ ] 관련 문서 업데이트

## @TEST:BUG-{ID} 검증

### 회귀 테스트 (Regression Tests)
{TEST_STRATEGY}

### 검증 시나리오 (Verification Scenarios)
{TEST_SCENARIOS}

---

**메타데이터**
- 심각도: {PRIORITY}
- 예상 시간: {ESTIMATED_HOURS}시간
- 작성자: {AUTHOR}
- 작성일: {CREATED_AT}
`
};

async function generateSpecId(): Promise<string> {
  const specsDir = '.moai/specs';

  try {
    const entries = await fs.readdir(specsDir, { withFileTypes: true });
    const specDirs = entries
      .filter(entry => entry.isDirectory())
      .map(entry => entry.name)
      .filter(name => name.startsWith('SPEC-'))
      .map(name => parseInt(name.replace('SPEC-', ''), 10))
      .filter(num => !isNaN(num));

    const maxId = Math.max(0, ...specDirs);
    return `SPEC-${String(maxId + 1).padStart(3, '0')}`;
  } catch {
    return 'SPEC-001';
  }
}

async function interactiveSpecBuilder(): Promise<Partial<SpecContent>> {
  const answers = await inquirer.prompt([
    {
      type: 'input',
      name: 'title',
      message: 'SPEC 제목을 입력하세요:',
      validate: (input: string) => input.length > 0 || '제목은 필수입니다.'
    },
    {
      type: 'list',
      name: 'type',
      message: 'SPEC 유형을 선택하세요:',
      choices: [
        { name: '새 기능 (Feature)', value: 'feature' },
        { name: '버그 수정 (Bug Fix)', value: 'bug' },
        { name: '개선사항 (Improvement)', value: 'improvement' },
        { name: '연구/조사 (Research)', value: 'research' }
      ]
    },
    {
      type: 'list',
      name: 'priority',
      message: '우선순위를 선택하세요:',
      choices: [
        { name: '🔴 Critical - 즉시 처리 필요', value: 'critical' },
        { name: '🟠 High - 이번 스프린트', value: 'high' },
        { name: '🟡 Medium - 다음 스프린트', value: 'medium' },
        { name: '🟢 Low - 백로그', value: 'low' }
      ]
    },
    {
      type: 'input',
      name: 'background',
      message: '배경/현상을 설명하세요:',
      validate: (input: string) => input.length > 0 || '배경 설명은 필수입니다.'
    },
    {
      type: 'input',
      name: 'problem',
      message: '해결하려는 문제를 설명하세요:',
      validate: (input: string) => input.length > 0 || '문제 설명은 필수입니다.'
    },
    {
      type: 'input',
      name: 'goals',
      message: '목표를 입력하세요 (쉼표로 구분):',
      filter: (input: string) => input.split(',').map(s => s.trim()).filter(s => s.length > 0)
    },
    {
      type: 'input',
      name: 'estimatedHours',
      message: '예상 구현 시간(시간):',
      default: '4',
      validate: (input: string) => !isNaN(parseInt(input)) || '숫자를 입력하세요.'
    }
  ]);

  return answers;
}

function formatSpecContent(template: string, data: any): string {
  let content = template;

  // 기본 치환
  const replacements = {
    '{ID}': data.metadata.id.replace('SPEC-', ''),
    '{TITLE}': data.metadata.title,
    '{TYPE}': data.metadata.type.toUpperCase(),
    '{PRIORITY}': data.metadata.priority,
    '{AUTHOR}': data.metadata.author,
    '{CREATED_AT}': data.metadata.createdAt,
    '{ESTIMATED_HOURS}': data.metadata.estimatedHours,
    '{BACKGROUND}': data.background || '여기에 배경을 작성하세요.',
    '{PROBLEM}': data.problem || '여기에 문제를 작성하세요.',
    '{GOALS}': Array.isArray(data.goals) ? data.goals.map(g => `- ${g}`).join('\n') : '- 여기에 목표를 작성하세요.',
    '{NON_GOALS}': '- 여기에 비목표를 작성하세요.',
    '{CONSTRAINTS}': '- 여기에 제약사항을 작성하세요.',
    '{APPROACH}': '여기에 접근 방법을 작성하세요.',
    '{RISKS}': '- 여기에 위험 요소를 작성하세요.',
    '{ALTERNATIVES}': '- 여기에 대안을 작성하세요.',
    '{ACCEPTANCE}': '- 여기에 승인 기준을 작성하세요.',
    '{TEST_STRATEGY}': '여기에 테스트 전략을 작성하세요.',
    '{TEST_SCENARIOS}': '- 여기에 테스트 시나리오를 작성하세요.',
    '{COVERAGE}': '85'
  };

  for (const [key, value] of Object.entries(replacements)) {
    content = content.replace(new RegExp(key, 'g'), String(value));
  }

  return content;
}

async function buildSpec(options: SpecBuilderOptions): Promise<{ success: boolean; message: string; specId?: string; filePath?: string }> {
  try {
    let specData: Partial<SpecContent>;

    if (options.interactive) {
      specData = await interactiveSpecBuilder();
    } else {
      if (!options.title) {
        throw new Error('비대화형 모드에서는 --title이 필수입니다.');
      }
      specData = {
        title: options.title,
        type: options.type || 'feature',
        priority: options.priority || 'medium',
        background: '여기에 배경을 작성하세요.',
        problem: '여기에 문제를 작성하세요.',
        goals: ['여기에 목표를 작성하세요.'],
        estimatedHours: 4
      };
    }

    const specId = options.id || await generateSpecId();
    const specDir = path.join('.moai/specs', specId);
    const specFile = path.join(specDir, 'spec.md');

    // SPEC 디렉토리 생성
    await fs.mkdir(specDir, { recursive: true });

    // 메타데이터 구성
    const metadata: SpecMetadata = {
      id: specId,
      title: specData.title!,
      type: specData.type as string,
      priority: specData.priority as string,
      status: 'draft',
      author: process.env.USER || 'unknown',
      createdAt: new Date().toISOString(),
      estimatedHours: parseInt(String(specData.estimatedHours)) || 4,
      tags: [`@REQ-${specData.type?.toUpperCase()}-${specId.replace('SPEC-', '')}`]
    };

    const fullSpecData = {
      metadata,
      ...specData
    };

    // 템플릿 선택 및 컨텐츠 생성
    const templateType = (specData.type as keyof typeof SPEC_TEMPLATES) || 'feature';
    const template = SPEC_TEMPLATES[templateType] || SPEC_TEMPLATES.feature;
    const specContent = formatSpecContent(template, fullSpecData);

    // SPEC 파일 작성
    await fs.writeFile(specFile, specContent);

    // 메타데이터 파일 작성
    await fs.writeFile(
      path.join(specDir, 'metadata.json'),
      JSON.stringify(metadata, null, 2)
    );

    return {
      success: true,
      message: `SPEC ${specId} 생성 완료`,
      specId,
      filePath: specFile
    };

  } catch (error) {
    return {
      success: false,
      message: `SPEC 생성 실패: ${error.message}`
    };
  }
}

program
  .name('spec-builder')
  .description('MoAI SPEC 문서 생성')
  .option('-i, --id <id>', 'SPEC ID (예: SPEC-001)')
  .option('-t, --title <title>', 'SPEC 제목')
  .option('--type <type>', 'SPEC 유형 (feature|bug|improvement|research)', 'feature')
  .option('-p, --priority <priority>', '우선순위 (critical|high|medium|low)', 'medium')
  .option('--interactive', '대화형 모드로 실행', false)
  .option('--template <template>', '사용할 템플릿')
  .action(async (options: SpecBuilderOptions) => {
    try {
      console.log(chalk.blue('📝 SPEC 생성 시작...'));

      const result = await buildSpec(options);

      if (result.success) {
        console.log(chalk.green('✅'), result.message);
        console.log(chalk.gray(`파일 경로: ${result.filePath}`));

        console.log(JSON.stringify({
          success: true,
          specId: result.specId,
          filePath: result.filePath,
          nextSteps: [
            `에디터에서 ${result.filePath} 파일을 편집하세요`,
            'moai 2-build로 TDD 구현을 시작하세요',
            'moai 3-sync로 문서를 동기화하세요'
          ]
        }, null, 2));
        process.exit(0);
      } else {
        console.error(chalk.red('❌'), result.message);
        console.log(JSON.stringify({
          success: false,
          error: result.message
        }, null, 2));
        process.exit(1);
      }
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