#!/usr/bin/env tsx
// CODE-TRUST-CHECKER-001: TRUST 5원칙 검증 스크립트
// 연결: SPEC-TRUST-001 → SPEC-TRUST-CHECKER-001 → CODE-TRUST-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import { execa } from 'execa';

interface TrustCheckerOptions {
  path?: string;
  principle?: 'test' | 'readable' | 'unified' | 'secured' | 'trackable' | 'all';
  fix?: boolean;
  strict?: boolean;
  report?: boolean;
}

interface TrustViolation {
  principle: string;
  severity: 'error' | 'warning' | 'info';
  file: string;
  line?: number;
  message: string;
  suggestion?: string;
  fixable: boolean;
}

interface TrustReport {
  principle: string;
  score: number;
  violations: TrustViolation[];
  metrics: Record<string, number>;
  recommendations: string[];
}

interface TrustAssessment {
  overallScore: number;
  reports: TrustReport[];
  summary: {
    totalViolations: number;
    criticalIssues: number;
    fixableIssues: number;
  };
  projectMetrics: {
    totalFiles: number;
    totalLines: number;
    testCoverage?: number;
    complexity: number;
  };
}

interface FileMetrics {
  path: string;
  lines: number;
  functions: number;
  complexity: number;
  testCoverage?: number;
}

async function scanProjectFiles(basePath: string): Promise<FileMetrics[]> {
  const files: FileMetrics[] = [];

  async function scanDirectory(dirPath: string): Promise<void> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry.name);

        if (entry.isDirectory()) {
          // 제외할 디렉토리
          if (!['node_modules', '.git', '.vscode', 'dist', 'build', 'target', 'coverage'].includes(entry.name)) {
            await scanDirectory(fullPath);
          }
        } else if (entry.isFile()) {
          const metrics = await analyzeFile(fullPath);
          if (metrics) {
            files.push(metrics);
          }
        }
      }
    } catch (error) {
      // 접근 권한 없는 디렉토리는 무시
    }
  }

  await scanDirectory(basePath);
  return files;
}

async function analyzeFile(filePath: string): Promise<FileMetrics | null> {
  try {
    const ext = path.extname(filePath).toLowerCase();
    const relevantExtensions = ['.ts', '.js', '.py', '.java', '.go', '.rs', '.cpp', '.c', '.h'];

    if (!relevantExtensions.includes(ext)) {
      return null;
    }

    const content = await fs.readFile(filePath, 'utf-8');
    const lines = content.split('\n').filter(line => line.trim().length > 0).length;

    // 함수 개수 계산 (언어별)
    const functions = countFunctions(content, ext);

    // 복잡도 계산 (간단한 순환 복잡도)
    const complexity = calculateComplexity(content, ext);

    return {
      path: filePath,
      lines,
      functions,
      complexity
    };

  } catch (error) {
    return null;
  }
}

function countFunctions(content: string, extension: string): number {
  let pattern: RegExp;

  switch (extension) {
    case '.ts':
    case '.js':
      pattern = /function\s+\w+|const\s+\w+\s*=\s*\([^)]*\)\s*=>|class\s+\w+/g;
      break;
    case '.py':
      pattern = /def\s+\w+|class\s+\w+/g;
      break;
    case '.java':
      pattern = /public\s+\w+\s+\w+\s*\(|private\s+\w+\s+\w+\s*\(|protected\s+\w+\s+\w+\s*\(/g;
      break;
    case '.go':
      pattern = /func\s+\w+/g;
      break;
    case '.rs':
      pattern = /fn\s+\w+/g;
      break;
    case '.cpp':
    case '.c':
      pattern = /\w+\s+\w+\s*\([^)]*\)\s*\{/g;
      break;
    default:
      return 0;
  }

  return (content.match(pattern) || []).length;
}

function calculateComplexity(content: string, extension: string): number {
  let complexityKeywords: string[];

  switch (extension) {
    case '.ts':
    case '.js':
      complexityKeywords = ['if', 'else', 'for', 'while', 'switch', 'case', 'catch', '&&', '||', '?'];
      break;
    case '.py':
      complexityKeywords = ['if', 'elif', 'else', 'for', 'while', 'try', 'except', 'and', 'or'];
      break;
    case '.java':
      complexityKeywords = ['if', 'else', 'for', 'while', 'switch', 'case', 'catch', '&&', '||', '?'];
      break;
    case '.go':
      complexityKeywords = ['if', 'else', 'for', 'switch', 'case', 'select', '&&', '||'];
      break;
    case '.rs':
      complexityKeywords = ['if', 'else', 'for', 'while', 'match', 'loop', '&&', '||'];
      break;
    default:
      return 1;
  }

  let complexity = 1; // 기본 복잡도

  for (const keyword of complexityKeywords) {
    const regex = new RegExp(`\\b${keyword}\\b`, 'g');
    const matches = content.match(regex) || [];
    complexity += matches.length;
  }

  return complexity;
}

async function checkTestPrinciple(files: FileMetrics[]): Promise<TrustReport> {
  const violations: TrustViolation[] = [];
  const metrics: Record<string, number> = {};

  // 테스트 파일 감지
  const testFiles = files.filter(f =>
    f.path.includes('test') ||
    f.path.includes('spec') ||
    f.path.includes('__tests__') ||
    /\.(test|spec)\.(ts|js|py|java|go|rs)$/.test(f.path)
  );

  const sourceFiles = files.filter(f => !testFiles.some(tf => tf.path === f.path));

  metrics.totalFiles = files.length;
  metrics.testFiles = testFiles.length;
  metrics.sourceFiles = sourceFiles.length;
  metrics.testToSourceRatio = sourceFiles.length > 0 ? testFiles.length / sourceFiles.length : 0;

  // TDD 원칙 위반 체크
  if (metrics.testToSourceRatio < 0.5) {
    violations.push({
      principle: 'Test',
      severity: 'error',
      file: 'project',
      message: `테스트 파일 비율이 낮습니다 (${(metrics.testToSourceRatio * 100).toFixed(1)}%)`,
      suggestion: '최소 50% 이상의 테스트 파일 비율을 유지하세요',
      fixable: false
    });
  }

  // 커버리지 체크 (가능한 경우)
  try {
    const coverage = await getCoverageData();
    if (coverage !== null) {
      metrics.coverage = coverage;
      if (coverage < 85) {
        violations.push({
          principle: 'Test',
          severity: coverage < 50 ? 'error' : 'warning',
          file: 'project',
          message: `테스트 커버리지가 낮습니다 (${coverage}%)`,
          suggestion: '테스트 커버리지를 85% 이상으로 향상시키세요',
          fixable: false
        });
      }
    }
  } catch {
    // 커버리지 데이터를 가져올 수 없으면 무시
  }

  const score = calculateScore(violations, 100);

  return {
    principle: 'Test-Driven Development',
    score,
    violations,
    metrics,
    recommendations: generateTestRecommendations(metrics, violations)
  };
}

async function checkReadablePrinciple(files: FileMetrics[]): Promise<TrustReport> {
  const violations: TrustViolation[] = [];
  const metrics: Record<string, number> = {};

  let totalLines = 0;
  let totalFunctions = 0;
  let filesWithLongFunctions = 0;
  let filesWithManyFunctions = 0;

  for (const file of files) {
    totalLines += file.lines;
    totalFunctions += file.functions;

    // 파일 크기 체크 (300 LOC 이하)
    if (file.lines > 300) {
      violations.push({
        principle: 'Readable',
        severity: 'warning',
        file: file.path,
        message: `파일이 너무 큽니다 (${file.lines} LOC > 300 LOC)`,
        suggestion: '파일을 더 작은 모듈로 분할하세요',
        fixable: false
      });
    }

    // 함수당 평균 라인 수 체크 (50 LOC 이하)
    if (file.functions > 0) {
      const avgLinesPerFunction = file.lines / file.functions;
      if (avgLinesPerFunction > 50) {
        filesWithLongFunctions++;
        violations.push({
          principle: 'Readable',
          severity: 'warning',
          file: file.path,
          message: `함수가 너무 깁니다 (평균 ${avgLinesPerFunction.toFixed(1)} LOC > 50 LOC)`,
          suggestion: '함수를 더 작은 단위로 분할하세요',
          fixable: false
        });
      }
    }

    // 파일당 함수 수 체크 (너무 많으면 단일 책임 원칙 위반)
    if (file.functions > 20) {
      filesWithManyFunctions++;
      violations.push({
        principle: 'Readable',
        severity: 'info',
        file: file.path,
        message: `파일에 함수가 너무 많습니다 (${file.functions}개)`,
        suggestion: '관련 함수들을 별도 모듈로 그룹화하세요',
        fixable: false
      });
    }
  }

  metrics.totalLines = totalLines;
  metrics.totalFunctions = totalFunctions;
  metrics.avgLinesPerFile = files.length > 0 ? totalLines / files.length : 0;
  metrics.avgFunctionsPerFile = files.length > 0 ? totalFunctions / files.length : 0;
  metrics.filesWithLongFunctions = filesWithLongFunctions;
  metrics.filesWithManyFunctions = filesWithManyFunctions;

  const score = calculateScore(violations, 100);

  return {
    principle: 'Readable Code',
    score,
    violations,
    metrics,
    recommendations: generateReadableRecommendations(metrics, violations)
  };
}

async function checkUnifiedPrinciple(files: FileMetrics[]): Promise<TrustReport> {
  const violations: TrustViolation[] = [];
  const metrics: Record<string, number> = {};

  // 복잡도 분석
  const complexities = files.map(f => f.complexity);
  const totalComplexity = complexities.reduce((sum, c) => sum + c, 0);
  const avgComplexity = files.length > 0 ? totalComplexity / files.length : 0;
  const maxComplexity = Math.max(...complexities, 0);

  metrics.avgComplexity = avgComplexity;
  metrics.maxComplexity = maxComplexity;
  metrics.highComplexityFiles = complexities.filter(c => c > 10).length;

  // 복잡도 위반 체크
  for (const file of files) {
    if (file.complexity > 10) {
      violations.push({
        principle: 'Unified',
        severity: file.complexity > 20 ? 'error' : 'warning',
        file: file.path,
        message: `순환 복잡도가 높습니다 (${file.complexity} > 10)`,
        suggestion: '함수를 분할하거나 조건문을 단순화하세요',
        fixable: false
      });
    }
  }

  // 아키텍처 일관성 체크 (디렉토리 구조)
  const directories = new Set(files.map(f => path.dirname(f.path)));
  const deepNesting = Array.from(directories).filter(dir => dir.split(path.sep).length > 5);

  if (deepNesting.length > 0) {
    violations.push({
      principle: 'Unified',
      severity: 'warning',
      file: 'project',
      message: `디렉토리 중첩이 너무 깊습니다 (최대 ${Math.max(...Array.from(directories).map(d => d.split(path.sep).length))}단계)`,
      suggestion: '디렉토리 구조를 평면화하세요',
      fixable: false
    });
  }

  const score = calculateScore(violations, 100);

  return {
    principle: 'Unified Architecture',
    score,
    violations,
    metrics,
    recommendations: generateUnifiedRecommendations(metrics, violations)
  };
}

async function checkSecuredPrinciple(files: FileMetrics[]): Promise<TrustReport> {
  const violations: TrustViolation[] = [];
  const metrics: Record<string, number> = {};

  let securityIssues = 0;
  let hardcodedSecrets = 0;

  // 보안 패턴 검사
  const dangerousPatterns = [
    { pattern: /password\s*=\s*["'][^"']+["']/gi, message: '하드코딩된 패스워드 발견' },
    { pattern: /api[_-]?key\s*=\s*["'][^"']+["']/gi, message: '하드코딩된 API 키 발견' },
    { pattern: /secret\s*=\s*["'][^"']+["']/gi, message: '하드코딩된 시크릿 발견' },
    { pattern: /token\s*=\s*["'][^"']+["']/gi, message: '하드코딩된 토큰 발견' },
    { pattern: /eval\s*\(/gi, message: '위험한 eval() 함수 사용' },
    { pattern: /exec\s*\(/gi, message: '위험한 exec() 함수 사용' },
    { pattern: /innerHTML\s*=/gi, message: 'XSS 위험: innerHTML 사용' },
    { pattern: /document\.write/gi, message: 'XSS 위험: document.write 사용' }
  ];

  for (const file of files) {
    try {
      const content = await fs.readFile(file.path, 'utf-8');

      for (const { pattern, message } of dangerousPatterns) {
        const matches = content.match(pattern);
        if (matches) {
          securityIssues++;
          if (pattern.source.includes('password|api|secret|token')) {
            hardcodedSecrets++;
          }

          violations.push({
            principle: 'Secured',
            severity: 'error',
            file: file.path,
            message: `${message}: ${matches[0]}`,
            suggestion: '환경변수나 설정 파일을 사용하세요',
            fixable: false
          });
        }
      }

      // TODO, FIXME 주석 확인 (보안 관련)
      const securityTodos = content.match(/(?:TODO|FIXME).*(?:security|auth|permission|access)/gi);
      if (securityTodos) {
        violations.push({
          principle: 'Secured',
          severity: 'warning',
          file: file.path,
          message: '미완성된 보안 구현이 있습니다',
          suggestion: '보안 관련 TODO를 즉시 해결하세요',
          fixable: false
        });
      }

    } catch (error) {
      // 파일 읽기 실패는 무시
    }
  }

  metrics.securityIssues = securityIssues;
  metrics.hardcodedSecrets = hardcodedSecrets;
  metrics.securityScore = Math.max(0, 100 - (securityIssues * 10));

  const score = calculateScore(violations, 100);

  return {
    principle: 'Secured',
    score,
    violations,
    metrics,
    recommendations: generateSecuredRecommendations(metrics, violations)
  };
}

async function checkTrackablePrinciple(files: FileMetrics[]): Promise<TrustReport> {
  const violations: TrustViolation[] = [];
  const metrics: Record<string, number> = {};

  let filesWithTags = 0;
  let totalTags = 0;

  // TAG 패턴 검사
  const tagPattern = /@([A-Z]+)(?:[:|-]([A-Z0-9-]+))?/g;

  for (const file of files) {
    try {
      const content = await fs.readFile(file.path, 'utf-8');
      const tags = content.match(tagPattern) || [];

      if (tags.length > 0) {
        filesWithTags++;
        totalTags += tags.length;
      } else if (file.path.includes('src/') || file.path.includes('lib/')) {
        // 소스 파일인데 TAG가 없는 경우
        violations.push({
          principle: 'Trackable',
          severity: 'warning',
          file: file.path,
          message: 'TAG가 없는 소스 파일입니다',
          suggestion: '적절한 @TAG를 추가하여 추적성을 확보하세요',
          fixable: true
        });
      }

    } catch (error) {
      // 파일 읽기 실패는 무시
    }
  }

  metrics.filesWithTags = filesWithTags;
  metrics.totalTags = totalTags;
  metrics.tagCoverage = files.length > 0 ? (filesWithTags / files.length) * 100 : 0;

  // TAG 데이터베이스 확인
  try {
    const tagDbPath = '.moai/indexes/tags.json';
    await fs.access(tagDbPath);
    const tagDb = JSON.parse(await fs.readFile(tagDbPath, 'utf-8'));
    metrics.tagsInDatabase = Object.keys(tagDb.tags || {}).length;
  } catch {
    violations.push({
      principle: 'Trackable',
      severity: 'error',
      file: 'project',
      message: 'TAG 데이터베이스가 없습니다',
      suggestion: 'tag-updater를 실행하여 TAG 시스템을 초기화하세요',
      fixable: true
    });
  }

  // Git 히스토리 확인
  try {
    const result = await execa('git', ['log', '--oneline', '-10'], { stdio: 'pipe' });
    const commits = result.stdout.split('\n').filter(line => line.trim().length > 0);
    const commitsWithTags = commits.filter(commit => /@[A-Z]+-\d+/.test(commit));

    metrics.recentCommits = commits.length;
    metrics.commitsWithTags = commitsWithTags.length;

    if (commitsWithTags.length / commits.length < 0.5) {
      violations.push({
        principle: 'Trackable',
        severity: 'warning',
        file: 'project',
        message: 'Git 커밋 메시지에 TAG가 부족합니다',
        suggestion: '커밋 메시지에 관련 TAG를 포함하세요',
        fixable: false
      });
    }
  } catch {
    // Git이 없거나 저장소가 아닌 경우 무시
  }

  const score = calculateScore(violations, 100);

  return {
    principle: 'Trackable',
    score,
    violations,
    metrics,
    recommendations: generateTrackableRecommendations(metrics, violations)
  };
}

async function getCoverageData(): Promise<number | null> {
  try {
    // package.json에서 test:coverage 스크립트 확인
    const packageJson = JSON.parse(await fs.readFile('package.json', 'utf-8'));
    if (packageJson.scripts?.['test:coverage']) {
      const result = await execa('npm', ['run', 'test:coverage'], {
        stdio: 'pipe',
        timeout: 30000
      });

      // Vitest 커버리지 파싱
      const coverageMatch = result.stdout.match(/All files\s+\|\s+([0-9.]+)/);
      if (coverageMatch) {
        return parseFloat(coverageMatch[1]);
      }
    }
  } catch {
    // 커버리지 실행 실패는 무시
  }

  // Python pytest-cov 시도
  try {
    const result = await execa('pytest', ['--cov=src', '--cov-report=term'], {
      stdio: 'pipe',
      timeout: 30000
    });

    const coverageMatch = result.stdout.match(/TOTAL\s+\d+\s+\d+\s+(\d+)%/);
    if (coverageMatch) {
      return parseInt(coverageMatch[1]);
    }
  } catch {
    // Python 프로젝트가 아니거나 실행 실패
  }

  return null;
}

function calculateScore(violations: TrustViolation[], maxScore: number): number {
  let score = maxScore;

  for (const violation of violations) {
    switch (violation.severity) {
      case 'error':
        score -= 20;
        break;
      case 'warning':
        score -= 10;
        break;
      case 'info':
        score -= 5;
        break;
    }
  }

  return Math.max(0, score);
}

function generateTestRecommendations(metrics: Record<string, number>, violations: TrustViolation[]): string[] {
  const recommendations: string[] = [];

  if (metrics.testToSourceRatio < 0.5) {
    recommendations.push('더 많은 테스트 파일을 작성하세요 (권장: 소스파일:테스트파일 = 1:0.5 이상)');
  }

  if (metrics.coverage && metrics.coverage < 85) {
    recommendations.push('테스트 커버리지를 85% 이상으로 향상시키세요');
  }

  if (violations.length === 0) {
    recommendations.push('훌륭합니다! TDD 원칙을 잘 따르고 있습니다');
  }

  return recommendations;
}

function generateReadableRecommendations(metrics: Record<string, number>, violations: TrustViolation[]): string[] {
  const recommendations: string[] = [];

  if (metrics.avgLinesPerFile > 200) {
    recommendations.push('파일 크기를 줄이세요 (권장: 300 LOC 이하)');
  }

  if (metrics.filesWithLongFunctions > 0) {
    recommendations.push('긴 함수들을 더 작은 단위로 분할하세요 (권장: 50 LOC 이하)');
  }

  if (metrics.filesWithManyFunctions > 0) {
    recommendations.push('관련 함수들을 별도 모듈로 그룹화하세요');
  }

  return recommendations;
}

function generateUnifiedRecommendations(metrics: Record<string, number>, violations: TrustViolation[]): string[] {
  const recommendations: string[] = [];

  if (metrics.avgComplexity > 5) {
    recommendations.push('코드 복잡도를 줄이세요 (권장: 순환복잡도 10 이하)');
  }

  if (metrics.highComplexityFiles > 0) {
    recommendations.push('복잡한 함수들을 리팩토링하세요');
  }

  recommendations.push('단일 책임 원칙을 준수하세요');
  recommendations.push('일관된 아키텍처 패턴을 유지하세요');

  return recommendations;
}

function generateSecuredRecommendations(metrics: Record<string, number>, violations: TrustViolation[]): string[] {
  const recommendations: string[] = [];

  if (metrics.hardcodedSecrets > 0) {
    recommendations.push('하드코딩된 시크릿을 환경변수로 이동하세요');
  }

  if (metrics.securityIssues > 0) {
    recommendations.push('보안 이슈를 즉시 해결하세요');
  }

  recommendations.push('입력 검증을 철저히 하세요');
  recommendations.push('보안 관련 라이브러리를 최신 버전으로 유지하세요');

  return recommendations;
}

function generateTrackableRecommendations(metrics: Record<string, number>, violations: TrustViolation[]): string[] {
  const recommendations: string[] = [];

  if (metrics.tagCoverage < 50) {
    recommendations.push('더 많은 파일에 @TAG를 추가하세요');
  }

  if (!metrics.tagsInDatabase) {
    recommendations.push('TAG 데이터베이스를 초기화하세요 (tag-updater 실행)');
  }

  recommendations.push('Git 커밋 메시지에 관련 TAG를 포함하세요');
  recommendations.push('정기적으로 TAG 시스템을 업데이트하세요');

  return recommendations;
}

async function runTrustAssessment(options: TrustCheckerOptions): Promise<TrustAssessment> {
  const basePath = options.path || process.cwd();

  console.log(chalk.blue('📁 프로젝트 파일 스캔 중...'));
  const files = await scanProjectFiles(basePath);

  console.log(chalk.blue(`📊 ${files.length}개 파일 분석 중...`));

  const reports: TrustReport[] = [];
  const principles = options.principle === 'all' ? ['test', 'readable', 'unified', 'secured', 'trackable'] : [options.principle || 'all'];

  if (principles.includes('all') || principles.includes('test')) {
    console.log(chalk.blue('🧪 Test 원칙 검사 중...'));
    const testReport = await checkTestPrinciple(files);
    reports.push(testReport);
  }

  if (principles.includes('all') || principles.includes('readable')) {
    console.log(chalk.blue('📖 Readable 원칙 검사 중...'));
    const readableReport = await checkReadablePrinciple(files);
    reports.push(readableReport);
  }

  if (principles.includes('all') || principles.includes('unified')) {
    console.log(chalk.blue('🏗️  Unified 원칙 검사 중...'));
    const unifiedReport = await checkUnifiedPrinciple(files);
    reports.push(unifiedReport);
  }

  if (principles.includes('all') || principles.includes('secured')) {
    console.log(chalk.blue('🔒 Secured 원칙 검사 중...'));
    const securedReport = await checkSecuredPrinciple(files);
    reports.push(securedReport);
  }

  if (principles.includes('all') || principles.includes('trackable')) {
    console.log(chalk.blue('🏷️  Trackable 원칙 검사 중...'));
    const trackableReport = await checkTrackablePrinciple(files);
    reports.push(trackableReport);
  }

  // 전체 점수 계산
  const overallScore = reports.length > 0 ? reports.reduce((sum, r) => sum + r.score, 0) / reports.length : 0;

  // 요약 통계
  const totalViolations = reports.reduce((sum, r) => sum + r.violations.length, 0);
  const criticalIssues = reports.reduce((sum, r) => sum + r.violations.filter(v => v.severity === 'error').length, 0);
  const fixableIssues = reports.reduce((sum, r) => sum + r.violations.filter(v => v.fixable).length, 0);

  // 프로젝트 메트릭
  const totalLines = files.reduce((sum, f) => sum + f.lines, 0);
  const avgComplexity = files.length > 0 ? files.reduce((sum, f) => sum + f.complexity, 0) / files.length : 0;

  return {
    overallScore,
    reports,
    summary: {
      totalViolations,
      criticalIssues,
      fixableIssues
    },
    projectMetrics: {
      totalFiles: files.length,
      totalLines,
      complexity: avgComplexity
    }
  };
}

async function saveReport(assessment: TrustAssessment): Promise<string> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const reportPath = `.moai/reports/trust-assessment-${timestamp}.json`;

  await fs.mkdir(path.dirname(reportPath), { recursive: true });
  await fs.writeFile(reportPath, JSON.stringify(assessment, null, 2));

  return reportPath;
}

program
  .name('trust-checker')
  .description('MoAI TRUST 5원칙 검증')
  .option('-p, --path <path>', '검사할 경로', process.cwd())
  .option('--principle <principle>', 'TRUST 원칙 (test|readable|unified|secured|trackable|all)', 'all')
  .option('-f, --fix', '자동 수정 가능한 문제 해결')
  .option('-s, --strict', '엄격 모드')
  .option('-r, --report', '상세 리포트 저장')
  .action(async (options: TrustCheckerOptions) => {
    try {
      console.log(chalk.blue('🛡️  TRUST 원칙 검증 시작...'));

      const assessment = await runTrustAssessment(options);

      // 전체 점수에 따른 색상 결정
      const scoreColor = assessment.overallScore >= 80 ? chalk.green :
                        assessment.overallScore >= 60 ? chalk.yellow : chalk.red;

      console.log(scoreColor(`\n🎯 전체 TRUST 점수: ${assessment.overallScore.toFixed(1)}/100`));

      // 원칙별 점수 출력
      for (const report of assessment.reports) {
        const color = report.score >= 80 ? chalk.green :
                     report.score >= 60 ? chalk.yellow : chalk.red;
        console.log(color(`  ${report.principle}: ${report.score}/100`));

        if (report.violations.length > 0 && options.strict) {
          for (const violation of report.violations.slice(0, 3)) { // 상위 3개만 표시
            console.log(chalk.gray(`    ${violation.severity}: ${violation.message}`));
          }
          if (report.violations.length > 3) {
            console.log(chalk.gray(`    ... ${report.violations.length - 3}개 추가 위반사항`));
          }
        }
      }

      // 요약 통계
      console.log(chalk.cyan('\n📊 요약 통계:'));
      console.log(`  총 위반사항: ${assessment.summary.totalViolations}개`);
      console.log(`  심각한 문제: ${assessment.summary.criticalIssues}개`);
      console.log(`  수정 가능: ${assessment.summary.fixableIssues}개`);

      // 리포트 저장
      if (options.report) {
        const reportPath = await saveReport(assessment);
        console.log(chalk.blue(`📁 상세 리포트 저장: ${reportPath}`));
      }

      // JSON 출력
      console.log(JSON.stringify({
        success: assessment.summary.criticalIssues === 0,
        overallScore: assessment.overallScore,
        summary: assessment.summary,
        principleScores: assessment.reports.map(r => ({
          principle: r.principle,
          score: r.score,
          violations: r.violations.length
        })),
        nextSteps: assessment.summary.criticalIssues > 0 ? [
          '심각한 TRUST 원칙 위반을 즉시 해결하세요',
          '자동 수정 가능한 문제는 --fix 옵션을 사용하세요'
        ] : [
          'TRUST 원칙을 잘 준수하고 있습니다',
          '지속적인 품질 관리를 유지하세요'
        ]
      }, null, 2));

      process.exit(assessment.summary.criticalIssues === 0 ? 0 : 1);
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