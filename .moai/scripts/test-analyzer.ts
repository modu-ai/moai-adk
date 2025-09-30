#!/usr/bin/env tsx
// @FEATURE-TEST-ANALYZER-001: 테스트 분석 및 품질 측정 스크립트
// 연결: @REQ-TEST-ANALYSIS-001 → @DESIGN-TEST-ANALYZER-001 → @TASK-TEST-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import { execa } from 'execa';

interface TestAnalyzerOptions {
  path?: string;
  format?: 'json' | 'table' | 'markdown';
  coverage?: boolean;
  performance?: boolean;
  trends?: boolean;
  save?: boolean;
}

interface TestFile {
  path: string;
  language: string;
  testCount: number;
  lineCount: number;
  complexity: number;
  coverage?: number;
}

interface TestSuite {
  name: string;
  files: TestFile[];
  totalTests: number;
  totalLines: number;
  avgComplexity: number;
  avgCoverage?: number;
  executionTime?: number;
  framework: string;
}

interface TestAnalysisResult {
  summary: {
    totalTestFiles: number;
    totalTests: number;
    totalLines: number;
    avgCoverage: number;
    codeToTestRatio: number;
    testDensity: number;
  };
  suites: TestSuite[];
  quality: {
    score: number;
    issues: string[];
    recommendations: string[];
  };
  trends?: {
    coverageChange: number;
    testCountChange: number;
    performanceChange: number;
  };
}

interface LanguageTestPattern {
  language: string;
  extensions: string[];
  testPatterns: string[];
  framework: string;
  coverageCommand?: string;
  complexityPattern?: RegExp;
}

const LANGUAGE_PATTERNS: LanguageTestPattern[] = [
  {
    language: 'TypeScript',
    extensions: ['.test.ts', '.spec.ts'],
    testPatterns: ['**/*.test.ts', '**/*.spec.ts', '__tests__/**/*.ts'],
    framework: 'vitest',
    coverageCommand: 'vitest run --coverage --reporter=json',
    complexityPattern: /function|class|if|for|while|switch/g
  },
  {
    language: 'JavaScript',
    extensions: ['.test.js', '.spec.js'],
    testPatterns: ['**/*.test.js', '**/*.spec.js', '__tests__/**/*.js'],
    framework: 'jest',
    coverageCommand: 'jest --coverage --outputFile=coverage/coverage.json',
    complexityPattern: /function|class|if|for|while|switch/g
  },
  {
    language: 'Python',
    extensions: ['.py'],
    testPatterns: ['test_*.py', '*_test.py', 'tests/**/*.py'],
    framework: 'pytest',
    coverageCommand: 'pytest --cov=src --cov-report=json',
    complexityPattern: /def |class |if |for |while |try:/g
  },
  {
    language: 'Java',
    extensions: ['.java'],
    testPatterns: ['**/*Test.java', '**/*Tests.java'],
    framework: 'junit',
    coverageCommand: 'mvn test jacoco:report',
    complexityPattern: /public|private|protected|if|for|while|switch/g
  },
  {
    language: 'Go',
    extensions: ['.go'],
    testPatterns: ['*_test.go'],
    framework: 'testing',
    coverageCommand: 'go test -cover -coverprofile=coverage.out ./...',
    complexityPattern: /func |if |for |switch |select/g
  },
  {
    language: 'Rust',
    extensions: ['.rs'],
    testPatterns: ['**/*_test.rs', '**/tests/*.rs'],
    framework: 'cargo',
    coverageCommand: 'cargo tarpaulin --out Json',
    complexityPattern: /fn |if |for |while |match |loop/g
  }
];

async function findTestFiles(searchPath: string): Promise<TestFile[]> {
  const testFiles: TestFile[] = [];

  async function scanDirectory(dirPath: string): Promise<void> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry.name);

        if (entry.isDirectory()) {
          // node_modules, .git 등 제외
          if (!['node_modules', '.git', '.vscode', 'dist', 'build', 'target'].includes(entry.name)) {
            await scanDirectory(fullPath);
          }
        } else if (entry.isFile()) {
          const testFile = await analyzeTestFile(fullPath);
          if (testFile) {
            testFiles.push(testFile);
          }
        }
      }
    } catch (error) {
      // 접근 권한 없는 디렉토리는 무시
    }
  }

  await scanDirectory(searchPath);
  return testFiles;
}

async function analyzeTestFile(filePath: string): Promise<TestFile | null> {
  try {
    const content = await fs.readFile(filePath, 'utf-8');
    const fileName = path.basename(filePath);

    // 테스트 파일인지 확인
    let languagePattern: LanguageTestPattern | null = null;
    for (const pattern of LANGUAGE_PATTERNS) {
      const isTestFile = pattern.extensions.some(ext => fileName.endsWith(ext)) ||
                        pattern.testPatterns.some(pat => {
                          // 간단한 glob 패턴 매칭
                          const regex = new RegExp(pat.replace(/\*\*/g, '.*').replace(/\*/g, '[^/]*'));
                          return regex.test(filePath);
                        });

      if (isTestFile) {
        languagePattern = pattern;
        break;
      }
    }

    if (!languagePattern) {
      return null;
    }

    // 테스트 개수 계산
    const testCount = countTests(content, languagePattern.language);

    // 라인 수 계산
    const lineCount = content.split('\n').length;

    // 복잡도 계산
    const complexity = calculateComplexity(content, languagePattern.complexityPattern);

    return {
      path: filePath,
      language: languagePattern.language,
      testCount,
      lineCount,
      complexity
    };

  } catch (error) {
    return null;
  }
}

function countTests(content: string, language: string): number {
  let count = 0;

  switch (language.toLowerCase()) {
    case 'typescript':
    case 'javascript':
      // it(), test(), describe() 블록 카운트
      count += (content.match(/\b(it|test|describe)\s*\(/g) || []).length;
      break;

    case 'python':
      // test_ 함수들 카운트
      count += (content.match(/def\s+test_\w+/g) || []).length;
      break;

    case 'java':
      // @Test 어노테이션 카운트
      count += (content.match(/@Test/g) || []).length;
      break;

    case 'go':
      // TestXxx 함수들 카운트
      count += (content.match(/func\s+Test\w+/g) || []).length;
      break;

    case 'rust':
      // #[test] 어노테이션 카운트
      count += (content.match(/#\[test\]/g) || []).length;
      break;
  }

  return count;
}

function calculateComplexity(content: string, pattern?: RegExp): number {
  if (!pattern) return 0;

  const matches = content.match(pattern) || [];
  return Math.ceil(matches.length / 10); // 10개 키워드당 복잡도 1
}

async function getCoverageData(language: string): Promise<{ coverage: number; details?: any } | null> {
  const pattern = LANGUAGE_PATTERNS.find(p => p.language.toLowerCase() === language.toLowerCase());
  if (!pattern?.coverageCommand) return null;

  try {
    const [command, ...args] = pattern.coverageCommand.split(' ');
    const result = await execa(command, args, { stdio: 'pipe', timeout: 60000 });

    // 언어별 커버리지 파싱
    switch (language.toLowerCase()) {
      case 'typescript':
      case 'javascript':
        // Vitest/Jest JSON 출력 파싱
        try {
          const coverageData = JSON.parse(result.stdout);
          return {
            coverage: coverageData.total?.lines?.pct || 0,
            details: coverageData
          };
        } catch {
          // 텍스트 출력에서 커버리지 추출
          const match = result.stdout.match(/All files\s+\|\s+([0-9.]+)/);
          return { coverage: match ? parseFloat(match[1]) : 0 };
        }

      case 'python':
        // pytest-cov JSON 출력 파싱
        try {
          const coverageData = JSON.parse(result.stdout);
          return {
            coverage: coverageData.totals?.percent_covered || 0,
            details: coverageData
          };
        } catch {
          const match = result.stdout.match(/TOTAL\s+\d+\s+\d+\s+(\d+)%/);
          return { coverage: match ? parseInt(match[1]) : 0 };
        }

      case 'go':
        // go test 커버리지 파싱
        const goMatch = result.stdout.match(/coverage:\s+([0-9.]+)%/);
        return { coverage: goMatch ? parseFloat(goMatch[1]) : 0 };

      default:
        return { coverage: 0 };
    }
  } catch (error) {
    return null;
  }
}

function groupTestsByFramework(testFiles: TestFile[]): TestSuite[] {
  const suites: Record<string, TestSuite> = {};

  for (const file of testFiles) {
    const pattern = LANGUAGE_PATTERNS.find(p => p.language === file.language);
    const framework = pattern?.framework || 'unknown';

    if (!suites[framework]) {
      suites[framework] = {
        name: framework,
        files: [],
        totalTests: 0,
        totalLines: 0,
        avgComplexity: 0,
        framework
      };
    }

    suites[framework].files.push(file);
    suites[framework].totalTests += file.testCount;
    suites[framework].totalLines += file.lineCount;
  }

  // 평균 복잡도 계산
  for (const suite of Object.values(suites)) {
    if (suite.files.length > 0) {
      suite.avgComplexity = suite.files.reduce((sum, f) => sum + f.complexity, 0) / suite.files.length;
    }
  }

  return Object.values(suites);
}

function calculateQualityScore(result: TestAnalysisResult): { score: number; issues: string[]; recommendations: string[] } {
  let score = 100;
  const issues: string[] = [];
  const recommendations: string[] = [];

  // 커버리지 평가 (40점)
  if (result.summary.avgCoverage < 50) {
    score -= 40;
    issues.push(`낮은 테스트 커버리지 (${result.summary.avgCoverage.toFixed(1)}%)`);
    recommendations.push('테스트 커버리지를 최소 85% 이상으로 향상시키세요');
  } else if (result.summary.avgCoverage < 70) {
    score -= 20;
    issues.push(`보통 테스트 커버리지 (${result.summary.avgCoverage.toFixed(1)}%)`);
    recommendations.push('테스트 커버리지를 85% 이상으로 향상시키세요');
  } else if (result.summary.avgCoverage < 85) {
    score -= 10;
    recommendations.push('테스트 커버리지를 85% 이상으로 향상시키세요');
  }

  // 테스트 밀도 평가 (30점)
  if (result.summary.testDensity < 0.1) {
    score -= 30;
    issues.push(`낮은 테스트 밀도 (${result.summary.testDensity.toFixed(3)})`);
    recommendations.push('더 많은 테스트를 작성하세요');
  } else if (result.summary.testDensity < 0.2) {
    score -= 15;
    recommendations.push('테스트 밀도를 향상시키세요');
  }

  // 코드 대비 테스트 비율 평가 (20점)
  if (result.summary.codeToTestRatio > 10) {
    score -= 20;
    issues.push(`높은 코드 대비 테스트 비율 (${result.summary.codeToTestRatio.toFixed(1)}:1)`);
    recommendations.push('테스트 코드의 비중을 늘리세요');
  } else if (result.summary.codeToTestRatio > 5) {
    score -= 10;
    recommendations.push('테스트 코드 비중을 늘리는 것을 고려하세요');
  }

  // 테스트 파일 수 평가 (10점)
  if (result.summary.totalTestFiles < 5) {
    score -= 10;
    issues.push(`적은 테스트 파일 수 (${result.summary.totalTestFiles}개)`);
    recommendations.push('더 많은 모듈에 대한 테스트를 작성하세요');
  }

  return {
    score: Math.max(0, Math.min(100, score)),
    issues,
    recommendations
  };
}

async function analyzeTests(options: TestAnalyzerOptions): Promise<TestAnalysisResult> {
  const searchPath = options.path || process.cwd();

  console.log(chalk.blue('🔍 테스트 파일 검색 중...'));
  const testFiles = await findTestFiles(searchPath);

  if (testFiles.length === 0) {
    throw new Error('테스트 파일을 찾을 수 없습니다.');
  }

  console.log(chalk.blue(`📊 ${testFiles.length}개 테스트 파일 분석 중...`));

  // 커버리지 데이터 수집
  if (options.coverage) {
    console.log(chalk.blue('📈 커버리지 데이터 수집 중...'));
    const languages = [...new Set(testFiles.map(f => f.language))];

    for (const language of languages) {
      const coverageData = await getCoverageData(language);
      if (coverageData) {
        // 해당 언어의 테스트 파일들에 커버리지 적용
        testFiles
          .filter(f => f.language === language)
          .forEach(f => f.coverage = coverageData.coverage);
      }
    }
  }

  // 테스트 스위트별 그룹화
  const suites = groupTestsByFramework(testFiles);

  // 전체 통계 계산
  const totalTests = testFiles.reduce((sum, f) => sum + f.testCount, 0);
  const totalLines = testFiles.reduce((sum, f) => sum + f.lineCount, 0);
  const avgCoverage = testFiles.length > 0
    ? testFiles.reduce((sum, f) => sum + (f.coverage || 0), 0) / testFiles.length
    : 0;

  // 소스 코드 라인 수 추정 (간단한 추정)
  const sourceLines = Math.round(totalLines * 3); // 테스트:소스 = 1:3 가정
  const codeToTestRatio = sourceLines / totalLines;
  const testDensity = totalTests / totalLines;

  const summary = {
    totalTestFiles: testFiles.length,
    totalTests,
    totalLines,
    avgCoverage,
    codeToTestRatio,
    testDensity
  };

  const result: TestAnalysisResult = {
    summary,
    suites,
    quality: { score: 0, issues: [], recommendations: [] }
  };

  // 품질 점수 계산
  result.quality = calculateQualityScore(result);

  return result;
}

function formatResults(result: TestAnalysisResult, format: string): string {
  switch (format) {
    case 'table':
      return formatAsTable(result);
    case 'markdown':
      return formatAsMarkdown(result);
    default:
      return JSON.stringify(result, null, 2);
  }
}

function formatAsTable(result: TestAnalysisResult): string {
  let output = '\n';
  output += chalk.cyan('📊 테스트 분석 결과\n');
  output += chalk.cyan('==================\n\n');

  // 요약
  output += chalk.yellow('📋 요약\n');
  output += `테스트 파일: ${result.summary.totalTestFiles}개\n`;
  output += `총 테스트: ${result.summary.totalTests}개\n`;
  output += `총 라인 수: ${result.summary.totalLines}줄\n`;
  output += `평균 커버리지: ${result.summary.avgCoverage.toFixed(1)}%\n`;
  output += `코드:테스트 비율: ${result.summary.codeToTestRatio.toFixed(1)}:1\n`;
  output += `테스트 밀도: ${result.summary.testDensity.toFixed(3)}\n\n`;

  // 품질 점수
  const scoreColor = result.quality.score >= 80 ? chalk.green :
                    result.quality.score >= 60 ? chalk.yellow : chalk.red;
  output += chalk.yellow('🎯 품질 점수\n');
  output += scoreColor(`점수: ${result.quality.score}/100\n\n`);

  // 테스트 스위트
  output += chalk.yellow('🧪 테스트 스위트\n');
  for (const suite of result.suites) {
    output += `${suite.framework}: ${suite.totalTests}개 테스트, ${suite.files.length}개 파일\n`;
  }

  return output;
}

function formatAsMarkdown(result: TestAnalysisResult): string {
  let md = '# 테스트 분석 리포트\n\n';

  // 요약
  md += '## 📋 요약\n\n';
  md += `- **테스트 파일**: ${result.summary.totalTestFiles}개\n`;
  md += `- **총 테스트**: ${result.summary.totalTests}개\n`;
  md += `- **총 라인 수**: ${result.summary.totalLines}줄\n`;
  md += `- **평균 커버리지**: ${result.summary.avgCoverage.toFixed(1)}%\n`;
  md += `- **코드:테스트 비율**: ${result.summary.codeToTestRatio.toFixed(1)}:1\n`;
  md += `- **테스트 밀도**: ${result.summary.testDensity.toFixed(3)}\n\n`;

  // 품질 점수
  md += '## 🎯 품질 점수\n\n';
  md += `**점수**: ${result.quality.score}/100\n\n`;

  if (result.quality.issues.length > 0) {
    md += '### ⚠️ 발견된 문제\n\n';
    for (const issue of result.quality.issues) {
      md += `- ${issue}\n`;
    }
    md += '\n';
  }

  if (result.quality.recommendations.length > 0) {
    md += '### 💡 개선 권장사항\n\n';
    for (const rec of result.quality.recommendations) {
      md += `- ${rec}\n`;
    }
    md += '\n';
  }

  // 테스트 스위트
  md += '## 🧪 테스트 스위트\n\n';
  md += '| 프레임워크 | 테스트 수 | 파일 수 | 평균 복잡도 |\n';
  md += '|-----------|----------|---------|------------|\n';
  for (const suite of result.suites) {
    md += `| ${suite.framework} | ${suite.totalTests} | ${suite.files.length} | ${suite.avgComplexity.toFixed(1)} |\n`;
  }

  return md;
}

program
  .name('test-analyzer')
  .description('MoAI 테스트 분석 및 품질 측정')
  .option('-p, --path <path>', '분석할 경로', process.cwd())
  .option('-f, --format <format>', '출력 형식 (json|table|markdown)', 'json')
  .option('-c, --coverage', '커버리지 데이터 수집')
  .option('--performance', '성능 데이터 수집')
  .option('--trends', '트렌드 분석')
  .option('--save', '결과를 파일로 저장')
  .action(async (options: TestAnalyzerOptions) => {
    try {
      console.log(chalk.blue('🧪 테스트 분석 시작...'));

      const result = await analyzeTests(options);
      const formattedOutput = formatResults(result, options.format);

      if (options.format === 'json') {
        console.log(formattedOutput);
      } else {
        console.log(formattedOutput);
        // JSON 결과도 함께 출력
        console.log('\n' + chalk.gray('=== JSON 출력 ==='));
        console.log(JSON.stringify({
          success: true,
          analysis: result
        }, null, 2));
      }

      // 파일 저장
      if (options.save) {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const filename = `.moai/reports/test-analysis-${timestamp}.${options.format === 'markdown' ? 'md' : 'json'}`;

        await fs.mkdir(path.dirname(filename), { recursive: true });
        await fs.writeFile(filename, formattedOutput);
        console.log(chalk.green(`📁 결과 저장: ${filename}`));
      }

      process.exit(0);
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