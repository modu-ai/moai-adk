#!/usr/bin/env tsx
// CODE-TDD-RUNNER-001: TDD Red-Green-Refactor 실행 스크립트
// 연결: SPEC-TDD-001 → SPEC-TDD-RUNNER-001 → CODE-TDD-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import { execa } from 'execa';

interface TddRunnerOptions {
  spec?: string;
  phase?: 'red' | 'green' | 'refactor' | 'all';
  language?: string;
  watch?: boolean;
  coverage?: boolean;
  autofix?: boolean;
}

interface LanguageConfig {
  name: string;
  testRunner: string;
  testPattern: string;
  coverageCommand: string;
  lintCommand: string;
  formatCommand: string;
  buildCommand?: string;
}

interface TddPhaseResult {
  phase: 'red' | 'green' | 'refactor';
  success: boolean;
  duration: number;
  output: string;
  coverage?: number;
  testsPassed: number;
  testsTotal: number;
}

interface TddCycleResult {
  success: boolean;
  phases: TddPhaseResult[];
  totalDuration: number;
  finalCoverage?: number;
}

const LANGUAGE_CONFIGS: Record<string, LanguageConfig> = {
  typescript: {
    name: 'TypeScript',
    testRunner: 'vitest run',
    testPattern: '**/*.test.ts',
    coverageCommand: 'vitest run --coverage',
    lintCommand: 'biome check',
    formatCommand: 'biome format --write',
    buildCommand: 'tsc --noEmit'
  },
  javascript: {
    name: 'JavaScript',
    testRunner: 'npm test',
    testPattern: '**/*.test.js',
    coverageCommand: 'npm run test:coverage',
    lintCommand: 'eslint',
    formatCommand: 'prettier --write'
  },
  python: {
    name: 'Python',
    testRunner: 'pytest',
    testPattern: 'test_*.py',
    coverageCommand: 'pytest --cov=src',
    lintCommand: 'ruff check',
    formatCommand: 'black .'
  },
  java: {
    name: 'Java',
    testRunner: 'mvn test',
    testPattern: '**/*Test.java',
    coverageCommand: 'mvn test jacoco:report',
    lintCommand: 'mvn checkstyle:check',
    formatCommand: 'mvn fmt:format'
  },
  go: {
    name: 'Go',
    testRunner: 'go test ./...',
    testPattern: '*_test.go',
    coverageCommand: 'go test -cover ./...',
    lintCommand: 'golangci-lint run',
    formatCommand: 'gofmt -w .'
  },
  rust: {
    name: 'Rust',
    testRunner: 'cargo test',
    testPattern: '**/*_test.rs',
    coverageCommand: 'cargo tarpaulin',
    lintCommand: 'cargo clippy',
    formatCommand: 'cargo fmt'
  }
};

async function detectProjectLanguage(): Promise<string> {
  // package.json 확인 (TypeScript/JavaScript)
  try {
    await fs.access('package.json');
    const packageJson = JSON.parse(await fs.readFile('package.json', 'utf-8'));
    if (packageJson.devDependencies?.typescript || packageJson.dependencies?.typescript) {
      return 'typescript';
    }
    return 'javascript';
  } catch {}

  // pyproject.toml/setup.py 확인 (Python)
  try {
    await fs.access('pyproject.toml');
    return 'python';
  } catch {}

  try {
    await fs.access('setup.py');
    return 'python';
  } catch {}

  // pom.xml/build.gradle 확인 (Java)
  try {
    await fs.access('pom.xml');
    return 'java';
  } catch {}

  try {
    await fs.access('build.gradle');
    return 'java';
  } catch {}

  // go.mod 확인 (Go)
  try {
    await fs.access('go.mod');
    return 'go';
  } catch {}

  // Cargo.toml 확인 (Rust)
  try {
    await fs.access('Cargo.toml');
    return 'rust';
  } catch {}

  return 'typescript'; // 기본값
}

async function runCommand(command: string, args: string[] = []): Promise<{ success: boolean; output: string; duration: number }> {
  const startTime = Date.now();

  try {
    const result = await execa(command, args, {
      stdio: 'pipe',
      timeout: 300000 // 5분 타임아웃
    });

    return {
      success: true,
      output: result.stdout || result.stderr || '',
      duration: Date.now() - startTime
    };
  } catch (error) {
    return {
      success: false,
      output: error.stdout || error.stderr || error.message || '',
      duration: Date.now() - startTime
    };
  }
}

function parseTestResults(output: string, language: string): { passed: number; total: number; coverage?: number } {
  let passed = 0;
  let total = 0;
  let coverage: number | undefined;

  switch (language) {
    case 'typescript':
    case 'javascript':
      // Vitest 출력 파싱
      const vitestMatch = output.match(/Tests?\s+(\d+)\s+passed.*?(\d+)\s+total/i);
      if (vitestMatch) {
        passed = parseInt(vitestMatch[1]);
        total = parseInt(vitestMatch[2]);
      }
      // 커버리지 파싱
      const coverageMatch = output.match(/All files\s+\|\s+([0-9.]+)/);
      if (coverageMatch) {
        coverage = parseFloat(coverageMatch[1]);
      }
      break;

    case 'python':
      // pytest 출력 파싱
      const pytestMatch = output.match(/(\d+)\s+passed.*?(\d+)\s+total/i);
      if (pytestMatch) {
        passed = parseInt(pytestMatch[1]);
        total = parseInt(pytestMatch[2]);
      }
      // 커버리지 파싱
      const pyCoverageMatch = output.match(/TOTAL\s+\d+\s+\d+\s+(\d+)%/);
      if (pyCoverageMatch) {
        coverage = parseInt(pyCoverageMatch[1]);
      }
      break;

    case 'go':
      // go test 출력 파싱
      const goTestMatch = output.match(/PASS.*?(\d+)\s+tests/i);
      if (goTestMatch) {
        passed = parseInt(goTestMatch[1]);
        total = passed; // Go는 실패한 테스트만 따로 표시
      }
      break;

    case 'rust':
      // cargo test 출력 파싱
      const rustMatch = output.match(/test result: ok\. (\d+) passed/i);
      if (rustMatch) {
        passed = parseInt(rustMatch[1]);
        total = passed;
      }
      break;
  }

  return { passed, total, coverage };
}

async function runRedPhase(langConfig: LanguageConfig): Promise<TddPhaseResult> {
  console.log(chalk.red('🔴 RED Phase: 실패하는 테스트 작성 및 실행'));

  const result = await runCommand(langConfig.testRunner.split(' ')[0], langConfig.testRunner.split(' ').slice(1));
  const testResults = parseTestResults(result.output, langConfig.name.toLowerCase());

  return {
    phase: 'red',
    success: !result.success, // RED 단계에서는 테스트가 실패해야 성공
    duration: result.duration,
    output: result.output,
    testsPassed: testResults.passed,
    testsTotal: testResults.total
  };
}

async function runGreenPhase(langConfig: LanguageConfig): Promise<TddPhaseResult> {
  console.log(chalk.green('🟢 GREEN Phase: 최소 구현으로 테스트 통과'));

  const result = await runCommand(langConfig.testRunner.split(' ')[0], langConfig.testRunner.split(' ').slice(1));
  const testResults = parseTestResults(result.output, langConfig.name.toLowerCase());

  return {
    phase: 'green',
    success: result.success && testResults.passed === testResults.total,
    duration: result.duration,
    output: result.output,
    testsPassed: testResults.passed,
    testsTotal: testResults.total
  };
}

async function runRefactorPhase(langConfig: LanguageConfig, options: TddRunnerOptions): Promise<TddPhaseResult> {
  console.log(chalk.blue('🔄 REFACTOR Phase: 코드 품질 개선'));

  let success = true;
  let output = '';
  let totalDuration = 0;
  let coverage: number | undefined;

  // 1. 린팅
  if (langConfig.lintCommand) {
    console.log('  📏 린팅 실행...');
    const lintResult = await runCommand(langConfig.lintCommand.split(' ')[0], langConfig.lintCommand.split(' ').slice(1));
    output += `\n=== Lint Results ===\n${lintResult.output}`;
    totalDuration += lintResult.duration;

    if (!lintResult.success && !options.autofix) {
      success = false;
    } else if (!lintResult.success && options.autofix) {
      // 자동 수정 시도
      console.log('  🔧 자동 수정 시도...');
      const fixCommand = langConfig.lintCommand.includes('biome')
        ? 'biome check --write'
        : langConfig.formatCommand;

      if (fixCommand) {
        await runCommand(fixCommand.split(' ')[0], fixCommand.split(' ').slice(1));
      }
    }
  }

  // 2. 포맷팅
  if (langConfig.formatCommand) {
    console.log('  🎨 포맷팅 실행...');
    const formatResult = await runCommand(langConfig.formatCommand.split(' ')[0], langConfig.formatCommand.split(' ').slice(1));
    output += `\n=== Format Results ===\n${formatResult.output}`;
    totalDuration += formatResult.duration;
  }

  // 3. 빌드 확인 (TypeScript 등)
  if (langConfig.buildCommand) {
    console.log('  🔨 빌드 확인...');
    const buildResult = await runCommand(langConfig.buildCommand.split(' ')[0], langConfig.buildCommand.split(' ').slice(1));
    output += `\n=== Build Results ===\n${buildResult.output}`;
    totalDuration += buildResult.duration;
    if (!buildResult.success) success = false;
  }

  // 4. 최종 테스트 실행
  console.log('  ✅ 최종 테스트 실행...');
  const testCommand = options.coverage ? langConfig.coverageCommand : langConfig.testRunner;
  const testResult = await runCommand(testCommand.split(' ')[0], testCommand.split(' ').slice(1));
  const testResults = parseTestResults(testResult.output, langConfig.name.toLowerCase());

  output += `\n=== Final Test Results ===\n${testResult.output}`;
  totalDuration += testResult.duration;
  coverage = testResults.coverage;

  if (!testResult.success || testResults.passed !== testResults.total) {
    success = false;
  }

  return {
    phase: 'refactor',
    success,
    duration: totalDuration,
    output,
    coverage,
    testsPassed: testResults.passed,
    testsTotal: testResults.total
  };
}

async function runTddCycle(options: TddRunnerOptions): Promise<TddCycleResult> {
  const startTime = Date.now();
  const language = options.language || await detectProjectLanguage();
  const langConfig = LANGUAGE_CONFIGS[language];

  if (!langConfig) {
    throw new Error(`지원하지 않는 언어입니다: ${language}`);
  }

  console.log(chalk.cyan(`🗿 TDD 사이클 시작 - ${langConfig.name}`));

  const phases: TddPhaseResult[] = [];

  if (options.phase === 'all' || options.phase === 'red') {
    const redResult = await runRedPhase(langConfig);
    phases.push(redResult);

    if (!redResult.success) {
      console.log(chalk.yellow('⚠️  RED 단계 경고: 테스트가 통과했습니다. 실패하는 테스트를 작성하세요.'));
    }
  }

  if (options.phase === 'all' || options.phase === 'green') {
    const greenResult = await runGreenPhase(langConfig);
    phases.push(greenResult);

    if (!greenResult.success) {
      console.log(chalk.red('❌ GREEN 단계 실패: 테스트를 통과하도록 구현하세요.'));
      if (options.phase === 'all') {
        return {
          success: false,
          phases,
          totalDuration: Date.now() - startTime
        };
      }
    }
  }

  if (options.phase === 'all' || options.phase === 'refactor') {
    const refactorResult = await runRefactorPhase(langConfig, options);
    phases.push(refactorResult);

    if (!refactorResult.success) {
      console.log(chalk.red('❌ REFACTOR 단계 실패: 품질 검사를 통과하지 못했습니다.'));
    }
  }

  const totalDuration = Date.now() - startTime;
  const success = phases.every(p => p.success);
  const finalCoverage = phases.find(p => p.coverage)?.coverage;

  return {
    success,
    phases,
    totalDuration,
    finalCoverage
  };
}

program
  .name('tdd-runner')
  .description('MoAI TDD Red-Green-Refactor 실행')
  .option('-s, --spec <spec-id>', '대상 SPEC ID')
  .option('-p, --phase <phase>', 'TDD 단계 (red|green|refactor|all)', 'all')
  .option('-l, --language <language>', '프로젝트 언어')
  .option('-w, --watch', '감시 모드로 실행')
  .option('-c, --coverage', '커버리지 측정')
  .option('--autofix', '자동 수정 시도')
  .action(async (options: TddRunnerOptions) => {
    try {
      console.log(chalk.blue('🔄 TDD 사이클 시작...'));

      const result = await runTddCycle(options);

      if (result.success) {
        console.log(chalk.green('✅ TDD 사이클 완료'));
        if (result.finalCoverage) {
          console.log(chalk.cyan(`📊 커버리지: ${result.finalCoverage}%`));
        }
      } else {
        console.log(chalk.red('❌ TDD 사이클 실패'));
      }

      console.log(chalk.gray(`⏱️  총 소요시간: ${(result.totalDuration / 1000).toFixed(1)}초`));

      console.log(JSON.stringify({
        success: result.success,
        totalDuration: result.totalDuration,
        finalCoverage: result.finalCoverage,
        phases: result.phases.map(p => ({
          phase: p.phase,
          success: p.success,
          duration: p.duration,
          testsPassed: p.testsPassed,
          testsTotal: p.testsTotal,
          coverage: p.coverage
        })),
        nextSteps: result.success ? [
          'moai 3-sync로 문서 동기화',
          '다음 SPEC으로 진행'
        ] : [
          '실패한 단계를 다시 실행하세요',
          '테스트 또는 구현을 수정하세요'
        ]
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