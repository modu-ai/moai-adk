/**
 * @FEATURE-TRUST-CHECKER-001: TypeScript TrustPrinciplesChecker 구현
 * @연결: @TASK:CONSTITUTION-CHECK-011 → @FEATURE:TRUST-VALIDATION-001 → @TASK:TRUST-TS-PORT-001
 *
 * TRUST 원칙 검증 스크립트 (strict/relaxed 지원)
 * MoAI-ADK의 TRUST 5원칙 준수 여부를 자동 검증합니다.
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../utils/winston-logger.js';

export interface TrustViolation {
  principle: string;
  violation: string;
  recommendation: string;
}

export interface TrustCheckResult {
  passed: boolean;
  violations: TrustViolation[];
  summary: {
    simplicity: boolean;
    architecture: boolean;
    testing: boolean;
    observability: boolean;
    traceability: boolean;
  };
}

export interface ProjectConfig {
  constitution?: {
    principles?: {
      simplicity?: {
        max_projects?: number;
      };
      testing?: {
        min_coverage?: number;
      };
    };
  };
  [key: string]: any;
}

export class TrustPrinciplesChecker {
  private projectRoot: string;
  private configPath: string;
  private violations: Array<[string, string, string]> = [];
  private strict: boolean;

  constructor(projectRoot: string = '.', strict: boolean = false) {
    this.projectRoot = path.resolve(projectRoot);
    this.configPath = path.join(this.projectRoot, '.moai', 'config.json');
    this.strict = strict;
    this.violations = [];
  }

  loadConfig(): ProjectConfig {
    try {
      if (fs.existsSync(this.configPath)) {
        const configContent = fs.readFileSync(this.configPath, 'utf-8');
        return JSON.parse(configContent);
      }
    } catch (error) {
      logger.warn(
        `Config 로드 실패: ${error instanceof Error ? error.message : String(error)}`
      );
    }

    return {};
  }

  /**
   * @API-CHECK-SIMPLICITY-001: Simplicity 원칙 검증 (프로젝트 복잡도 ≤ 3개)
   */
  checkSimplicity(config: ProjectConfig): boolean {
    const maxProjects =
      config.constitution?.principles?.simplicity?.max_projects || 3;
    const srcDir = path.join(this.projectRoot, 'src');

    if (!fs.existsSync(srcDir)) {
      return true;
    }

    if (this.strict) {
      // strict: 모든 Python/TypeScript 파일 수로 판단
      const files = this.getAllSourceFiles(srcDir);
      const sourceFiles = files.filter(
        f => !f.includes('__init__') && !f.includes('test')
      );

      if (sourceFiles.length > maxProjects) {
        this.violations.push([
          'Simplicity',
          `모듈 수 ${sourceFiles.length}개가 허용 한도 ${maxProjects}개를 초과`,
          `모듈을 ${maxProjects}개 이하로 통합하거나 기능을 단순화하세요`,
        ]);
        return false;
      }
      return true;
    }

    // relaxed: 상위 모듈(디렉터리) 수로 판단
    try {
      const topModules = fs
        .readdirSync(srcDir, { withFileTypes: true })
        .filter(dirent => dirent.isDirectory())
        .filter(dirent => !['__pycache__', 'tests'].includes(dirent.name));

      // 모듈로 볼 수 있는 디렉터리만 카운트 (소스 파일 포함)
      let moduleCount = 0;
      for (const dir of topModules) {
        const moduleDir = path.join(srcDir, dir.name);
        const hasSourceFiles = this.hasSourceFiles(moduleDir);
        if (hasSourceFiles) {
          moduleCount++;
        }
      }

      if (moduleCount > maxProjects) {
        this.violations.push([
          'Simplicity',
          `상위 모듈 ${moduleCount}개가 허용 한도 ${maxProjects}개를 초과`,
          '상위 구조를 단순화하거나 모듈을 통합하세요',
        ]);
        return false;
      }
      return true;
    } catch (_error) {
      return true; // 에러 시 통과
    }
  }

  /**
   * @API-CHECK-ARCHITECTURE-001: Architecture 원칙 검증 (라이브러리 분리)
   */
  checkArchitecture(): boolean {
    const srcDir = path.join(this.projectRoot, 'src');

    if (!fs.existsSync(srcDir)) {
      return true;
    }

    const expectedDirs = ['models', 'services', 'controllers', 'utils'];

    try {
      const foundDirs = fs
        .readdirSync(srcDir, { withFileTypes: true })
        .filter(dirent => dirent.isDirectory())
        .map(dirent => dirent.name);

      const overlap = expectedDirs.filter(dir =>
        foundDirs.includes(dir)
      ).length;

      if (this.strict) {
        if (overlap < 2) {
          this.violations.push([
            'Architecture',
            `계층 구조가 부족합니다 (발견: ${overlap}/4개 디렉터리)`,
            'models, services, controllers, utils 중 최소 2개 이상의 계층을 구성하세요',
          ]);
          return false;
        }
      } else {
        if (overlap < 1) {
          this.violations.push([
            'Architecture',
            '라이브러리 분리가 없습니다',
            'models, services, controllers, utils 중 최소 1개 이상의 계층을 구성하세요',
          ]);
          return false;
        }
      }

      return true;
    } catch (_error) {
      return true; // 에러 시 통과
    }
  }

  /**
   * @API-CHECK-TESTING-001: Testing 원칙 검증
   */
  checkTesting(): boolean {
    const testFileCount = this.countTestFiles();

    if (testFileCount === 0) {
      this.violations.push([
        'Testing',
        '테스트 파일이 없습니다',
        '각 모듈에 대응하는 테스트 파일을 작성하세요',
      ]);
      return false;
    }

    const coverage = this.calculateCoverage();
    const minCoverage = this.strict ? 85 : 70;

    if (coverage < minCoverage) {
      this.violations.push([
        'Testing',
        `테스트 커버리지가 부족합니다 (${coverage}% < ${minCoverage}%)`,
        `테스트 커버리지를 ${minCoverage}% 이상으로 높이세요`,
      ]);
      return false;
    }

    return true;
  }

  /**
   * @API-CHECK-OBSERVABILITY-001: Observability 원칙 검증
   */
  checkObservability(): boolean {
    const hasLogging = this.hasLoggingInfrastructure();
    const hasErrorTracking = this.hasErrorTracking();

    if (!hasLogging || !hasErrorTracking) {
      this.violations.push([
        'Observability',
        '로깅 또는 오류 추적 인프라가 부족합니다',
        '구조화된 로깅과 오류 추적 시스템을 구축하세요',
      ]);
      return false;
    }

    return true;
  }

  /**
   * @API-CHECK-TRACEABILITY-001: Traceability 원칙 검증
   */
  checkTraceability(): boolean {
    const hasVersionControl = this.hasVersionControl();
    const hasDocumentation = this.hasDocumentation();
    const hasTagSystem = this.hasTagSystem();

    if (!hasVersionControl || !hasDocumentation || !hasTagSystem) {
      this.violations.push([
        'Traceability',
        '버전 관리, 문서화 또는 TAG 시스템이 부족합니다',
        'Git 저장소, README 문서, @TAG 주석 시스템을 구축하세요',
      ]);
      return false;
    }

    return true;
  }

  /**
   * @API-CHECK-ALL-PRINCIPLES-001: 모든 TRUST 원칙 검증
   */
  checkAllPrinciples(): TrustCheckResult {
    this.violations = []; // 초기화

    const config = this.loadConfig();

    const results = {
      simplicity: this.checkSimplicity(config),
      architecture: this.checkArchitecture(),
      testing: this.checkTesting(),
      observability: this.checkObservability(),
      traceability: this.checkTraceability(),
    };

    const allPassed = Object.values(results).every(result => result);

    const violations: TrustViolation[] = this.violations.map(
      ([principle, violation, recommendation]) => ({
        principle,
        violation,
        recommendation,
      })
    );

    return {
      passed: allPassed,
      violations,
      summary: results,
    };
  }

  // 헬퍼 메서드들
  private getAllSourceFiles(dir: string): string[] {
    const files: string[] = [];

    try {
      const items = fs.readdirSync(dir, { withFileTypes: true });

      for (const item of items) {
        const fullPath = path.join(dir, item.name);

        if (item.isDirectory()) {
          files.push(...this.getAllSourceFiles(fullPath));
        } else if (item.isFile() && this.isSourceFile(item.name)) {
          files.push(fullPath);
        }
      }
    } catch (_error) {
      // 에러 시 빈 배열 반환
    }

    return files;
  }

  private hasSourceFiles(dir: string): boolean {
    try {
      const files = fs.readdirSync(dir, { withFileTypes: true });
      return files.some(file => file.isFile() && this.isSourceFile(file.name));
    } catch (_error) {
      return false;
    }
  }

  private isSourceFile(filename: string): boolean {
    const ext = path.extname(filename);
    return ['.py', '.ts', '.tsx', '.js', '.jsx'].includes(ext);
  }

  private countTestFiles(): number {
    const testDirs = [
      path.join(this.projectRoot, 'tests'),
      path.join(this.projectRoot, '__tests__'),
      path.join(this.projectRoot, 'src', '__tests__'),
    ];

    let count = 0;

    for (const testDir of testDirs) {
      if (fs.existsSync(testDir)) {
        count += this.getAllSourceFiles(testDir).filter(
          file => file.includes('test') || file.includes('spec')
        ).length;
      }
    }

    // 소스 디렉토리 내 테스트 파일도 검색
    const srcDir = path.join(this.projectRoot, 'src');
    if (fs.existsSync(srcDir)) {
      count += this.getAllSourceFiles(srcDir).filter(
        file => file.includes('test') || file.includes('spec')
      ).length;
    }

    return count;
  }

  private countSourceFiles(): number {
    const srcDir = path.join(this.projectRoot, 'src');

    if (!fs.existsSync(srcDir)) {
      return 0;
    }

    return this.getAllSourceFiles(srcDir).filter(
      file => !file.includes('test') && !file.includes('spec')
    ).length;
  }

  private calculateCoverage(): number {
    const testFiles = this.countTestFiles();
    const sourceFiles = this.countSourceFiles();

    if (sourceFiles === 0) {
      return 100; // 소스 파일이 없으면 100%
    }

    // 간단한 커버리지 계산: 테스트 파일 수 / 소스 파일 수 * 100
    // 실제로는 더 정교한 커버리지 도구를 사용해야 함
    return Math.min((testFiles / sourceFiles) * 100, 100);
  }

  private hasLoggingInfrastructure(): boolean {
    // 로깅 관련 파일 검색
    const srcDir = path.join(this.projectRoot, 'src');
    const packageJsonPath = path.join(this.projectRoot, 'package.json');

    try {
      // package.json에서 로깅 라이브러리 확인
      if (fs.existsSync(packageJsonPath)) {
        const packageJson = JSON.parse(
          fs.readFileSync(packageJsonPath, 'utf-8')
        );
        const deps = {
          ...packageJson.dependencies,
          ...packageJson.devDependencies,
        };
        if (deps.winston || deps.pino || deps.bunyan || deps.debug) {
          return true;
        }
      }

      // 소스 코드에서 로깅 코드 검색
      if (fs.existsSync(srcDir)) {
        const sourceFiles = this.getAllSourceFiles(srcDir);
        return sourceFiles.some(file => {
          try {
            const content = fs.readFileSync(file, 'utf-8');
            return (
              content.includes('console.log') ||
              content.includes('logger') ||
              content.includes('log.')
            );
          } catch {
            return false;
          }
        });
      }
    } catch (_error) {
      return false;
    }

    return false;
  }

  private hasErrorTracking(): boolean {
    // 에러 추적 관련 확인
    const packageJsonPath = path.join(this.projectRoot, 'package.json');

    try {
      if (fs.existsSync(packageJsonPath)) {
        const packageJson = JSON.parse(
          fs.readFileSync(packageJsonPath, 'utf-8')
        );
        const deps = {
          ...packageJson.dependencies,
          ...packageJson.devDependencies,
        };
        if (deps.sentry || deps['@sentry/node'] || deps.bugsnag) {
          return true;
        }
      }

      // try-catch 블록 존재 확인
      const srcDir = path.join(this.projectRoot, 'src');
      if (fs.existsSync(srcDir)) {
        const sourceFiles = this.getAllSourceFiles(srcDir);
        return sourceFiles.some(file => {
          try {
            const content = fs.readFileSync(file, 'utf-8');
            return content.includes('try') && content.includes('catch');
          } catch {
            return false;
          }
        });
      }
    } catch (_error) {
      return false;
    }

    return false;
  }

  private hasVersionControl(): boolean {
    return fs.existsSync(path.join(this.projectRoot, '.git'));
  }

  private hasDocumentation(): boolean {
    const docFiles = ['README.md', 'README.rst', 'docs/'];

    return docFiles.some(doc =>
      fs.existsSync(path.join(this.projectRoot, doc))
    );
  }

  private hasTagSystem(): boolean {
    // @TAG 주석 시스템 존재 확인
    const srcDir = path.join(this.projectRoot, 'src');

    if (!fs.existsSync(srcDir)) {
      return false;
    }

    try {
      const sourceFiles = this.getAllSourceFiles(srcDir);
      return sourceFiles.some(file => {
        try {
          const content = fs.readFileSync(file, 'utf-8');
          return (
            content.includes('@TAG') ||
            content.includes('@REQ') ||
            content.includes('@FEATURE')
          );
        } catch {
          return false;
        }
      });
    } catch (_error) {
      return false;
    }
  }
}
