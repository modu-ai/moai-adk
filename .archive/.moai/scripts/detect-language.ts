#!/usr/bin/env tsx
// CODE-LANGUAGE-DETECT-001: 프로젝트 언어 자동 감지 스크립트
// 연결: SPEC-LANGUAGE-001 → SPEC-DETECT-001 → CODE-DETECT-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';

interface LanguageDetectOptions {
  path?: string;
  exclude?: string[];
  include?: string[];
  verbose?: boolean;
}

interface LanguageInfo {
  name: string;
  extensions: string[];
  configFiles: string[];
  testFrameworks: string[];
  linters: string[];
  formatters: string[];
  packageManagers: string[];
}

interface DetectionResult {
  primary: string;
  secondary: string[];
  confidence: number;
  files: Record<string, number>;
  frameworks: {
    test: string[];
    build: string[];
    lint: string[];
  };
  packageManager: string | null;
}

const LANGUAGE_MAP: Record<string, LanguageInfo> = {
  typescript: {
    name: 'TypeScript',
    extensions: ['.ts', '.tsx'],
    configFiles: ['tsconfig.json', 'tsup.config.ts', 'vite.config.ts'],
    testFrameworks: ['vitest', 'jest', 'mocha'],
    linters: ['biome', 'eslint', 'tslint'],
    formatters: ['biome', 'prettier'],
    packageManagers: ['bun', 'npm', 'yarn', 'pnpm']
  },
  javascript: {
    name: 'JavaScript',
    extensions: ['.js', '.jsx', '.mjs'],
    configFiles: ['package.json', 'webpack.config.js', 'rollup.config.js'],
    testFrameworks: ['jest', 'mocha', 'jasmine'],
    linters: ['eslint', 'jshint'],
    formatters: ['prettier'],
    packageManagers: ['npm', 'yarn', 'pnpm']
  },
  python: {
    name: 'Python',
    extensions: ['.py', '.pyw'],
    configFiles: ['pyproject.toml', 'setup.py', 'requirements.txt'],
    testFrameworks: ['pytest', 'unittest', 'nose'],
    linters: ['ruff', 'flake8', 'pylint'],
    formatters: ['black', 'autopep8'],
    packageManagers: ['pip', 'pipenv', 'poetry']
  },
  java: {
    name: 'Java',
    extensions: ['.java'],
    configFiles: ['pom.xml', 'build.gradle', 'gradle.properties'],
    testFrameworks: ['junit', 'testng'],
    linters: ['checkstyle', 'spotbugs'],
    formatters: ['google-java-format'],
    packageManagers: ['maven', 'gradle']
  },
  go: {
    name: 'Go',
    extensions: ['.go'],
    configFiles: ['go.mod', 'go.sum'],
    testFrameworks: ['testing'],
    linters: ['golint', 'golangci-lint'],
    formatters: ['gofmt', 'goimports'],
    packageManagers: ['go']
  },
  rust: {
    name: 'Rust',
    extensions: ['.rs'],
    configFiles: ['Cargo.toml', 'Cargo.lock'],
    testFrameworks: ['cargo test'],
    linters: ['clippy'],
    formatters: ['rustfmt'],
    packageManagers: ['cargo']
  },
  csharp: {
    name: 'C#',
    extensions: ['.cs'],
    configFiles: ['.csproj', '.sln'],
    testFrameworks: ['xunit', 'nunit', 'mstest'],
    linters: ['roslyn'],
    formatters: ['dotnet format'],
    packageManagers: ['nuget', 'dotnet']
  }
};

async function scanFiles(dirPath: string, excludePatterns: string[] = []): Promise<Record<string, number>> {
  const fileExtensions: Record<string, number> = {};

  async function scanDir(currentPath: string): Promise<void> {
    try {
      const entries = await fs.readdir(currentPath, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(currentPath, entry.name);

        // 제외 패턴 확인
        if (excludePatterns.some(pattern => fullPath.includes(pattern))) {
          continue;
        }

        if (entry.isDirectory()) {
          // node_modules, .git 등 제외
          if (!['node_modules', '.git', '.vscode', 'dist', 'build'].includes(entry.name)) {
            await scanDir(fullPath);
          }
        } else if (entry.isFile()) {
          const ext = path.extname(entry.name);
          if (ext) {
            fileExtensions[ext] = (fileExtensions[ext] || 0) + 1;
          }
        }
      }
    } catch (error) {
      // 접근 권한 없는 디렉토리는 무시
    }
  }

  await scanDir(dirPath);
  return fileExtensions;
}

async function detectConfigFiles(dirPath: string): Promise<string[]> {
  const configFiles: string[] = [];

  for (const [, langInfo] of Object.entries(LANGUAGE_MAP)) {
    for (const configFile of langInfo.configFiles) {
      try {
        await fs.access(path.join(dirPath, configFile));
        configFiles.push(configFile);
      } catch {
        // 파일이 없으면 무시
      }
    }
  }

  return configFiles;
}

async function detectLanguage(options: LanguageDetectOptions): Promise<DetectionResult> {
  const targetPath = options.path || process.cwd();
  const excludePatterns = options.exclude || [];

  // 파일 확장자 스캔
  const fileExtensions = await scanFiles(targetPath, excludePatterns);

  // 설정 파일 감지
  const configFiles = await detectConfigFiles(targetPath);

  // 언어별 점수 계산
  const languageScores: Record<string, number> = {};

  for (const [langKey, langInfo] of Object.entries(LANGUAGE_MAP)) {
    let score = 0;

    // 파일 확장자 점수 (가중치: 1)
    for (const ext of langInfo.extensions) {
      score += (fileExtensions[ext] || 0) * 1;
    }

    // 설정 파일 점수 (가중치: 10)
    for (const configFile of langInfo.configFiles) {
      if (configFiles.includes(configFile)) {
        score += 10;
      }
    }

    if (score > 0) {
      languageScores[langKey] = score;
    }
  }

  // 언어 순위 정렬
  const sortedLanguages = Object.entries(languageScores)
    .sort(([, a], [, b]) => b - a)
    .map(([lang]) => lang);

  const primary = sortedLanguages[0] || 'unknown';
  const secondary = sortedLanguages.slice(1, 4);

  // 신뢰도 계산 (0-100)
  const totalScore = Object.values(languageScores).reduce((sum, score) => sum + score, 0);
  const primaryScore = languageScores[primary] || 0;
  const confidence = totalScore > 0 ? Math.round((primaryScore / totalScore) * 100) : 0;

  // 프레임워크 감지
  const frameworks = {
    test: [] as string[],
    build: [] as string[],
    lint: [] as string[]
  };

  if (primary !== 'unknown') {
    const langInfo = LANGUAGE_MAP[primary];

    // package.json에서 프레임워크 감지 (Node.js 계열)
    if (['typescript', 'javascript'].includes(primary)) {
      try {
        const packageJson = JSON.parse(
          await fs.readFile(path.join(targetPath, 'package.json'), 'utf-8')
        );

        const allDeps = {
          ...packageJson.dependencies,
          ...packageJson.devDependencies
        };

        for (const dep of Object.keys(allDeps)) {
          if (langInfo.testFrameworks.includes(dep)) {
            frameworks.test.push(dep);
          }
          if (langInfo.linters.includes(dep)) {
            frameworks.lint.push(dep);
          }
        }
      } catch {
        // package.json 없으면 무시
      }
    }
  }

  // 패키지 매니저 감지
  let packageManager: string | null = null;
  const packageManagerFiles = {
    'bun.lock': 'bun',
    'package-lock.json': 'npm',
    'yarn.lock': 'yarn',
    'pnpm-lock.yaml': 'pnpm',
    'poetry.lock': 'poetry',
    'Cargo.lock': 'cargo',
    'pom.xml': 'maven',
    'build.gradle': 'gradle'
  };

  for (const [file, manager] of Object.entries(packageManagerFiles)) {
    try {
      await fs.access(path.join(targetPath, file));
      packageManager = manager;
      break;
    } catch {
      // 파일이 없으면 무시
    }
  }

  return {
    primary,
    secondary,
    confidence,
    files: fileExtensions,
    frameworks,
    packageManager
  };
}

program
  .name('detect-language')
  .description('프로젝트 언어 자동 감지')
  .option('-p, --path <path>', '스캔할 경로', process.cwd())
  .option('-e, --exclude <patterns...>', '제외할 패턴들')
  .option('-i, --include <patterns...>', '포함할 패턴들')
  .option('-v, --verbose', '상세 출력')
  .action(async (options: LanguageDetectOptions) => {
    try {
      if (options.verbose) {
        console.log(chalk.blue('🔍 언어 감지 시작...'));
        console.log(chalk.gray(`스캔 경로: ${options.path}`));
      }

      const result = await detectLanguage(options);

      if (options.verbose) {
        console.log(chalk.green('✅ 언어 감지 완료'));
        console.log(`주 언어: ${chalk.yellow(result.primary)} (신뢰도: ${result.confidence}%)`);
        if (result.secondary.length > 0) {
          console.log(`보조 언어: ${result.secondary.join(', ')}`);
        }
        if (result.packageManager) {
          console.log(`패키지 매니저: ${result.packageManager}`);
        }
      }

      console.log(JSON.stringify({
        success: true,
        language: {
          primary: result.primary,
          secondary: result.secondary,
          confidence: result.confidence
        },
        files: result.files,
        frameworks: result.frameworks,
        packageManager: result.packageManager,
        recommendation: {
          testFramework: LANGUAGE_MAP[result.primary]?.testFrameworks[0] || 'unknown',
          linter: LANGUAGE_MAP[result.primary]?.linters[0] || 'unknown',
          formatter: LANGUAGE_MAP[result.primary]?.formatters[0] || 'unknown'
        }
      }, null, 2));

      process.exit(0);
    } catch (error) {
      console.error(chalk.red('❌ 언어 감지 실패:'), error.message);
      console.log(JSON.stringify({
        success: false,
        error: error.message
      }, null, 2));
      process.exit(1);
    }
  });

program.parse();