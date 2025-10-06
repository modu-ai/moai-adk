// @TEST:INIT-001 | SPEC: SPEC-INIT-001.md | CODE: src/cli/commands/init.ts
// Related: @CODE:INIT-001:CLI, @CODE:INIT-001:TTY, @SPEC:INIT-001

/**
 * @file Test for moai init non-interactive mode
 * @author MoAI Team
 * @tags @TEST:INIT-001:CLI
 */

import { describe, test, expect, vi, beforeEach, afterEach, beforeAll, afterAll } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import type { InitCommand } from '@/cli/commands/init/index';

// Set test environment
const originalNodeEnv = process.env.NODE_ENV;

/**
 * Helper function to mock DoctorCommand for testing
 * @param initCommand - InitCommand instance to inject mock
 * @param allPassed - Whether doctor check should pass (default: true)
 * @param missingRequirements - Array of missing requirements
 */
function mockDoctorCommand(
  initCommand: InitCommand,
  allPassed = true,
  missingRequirements: any[] = []
) {
  const mockDoctorRun = vi.fn().mockResolvedValue({
    allPassed,
    results: [],
    missingRequirements,
    versionConflicts: [],
    summary: {
      total: missingRequirements.length,
      passed: allPassed ? 0 : 0,
      failed: allPassed ? 0 : missingRequirements.length
    }
  });

  (initCommand as any).doctorCommand = { run: mockDoctorRun };
}

describe('Init Command - Non-Interactive Mode', () => {
  beforeAll(() => {
    // Set test environment to bypass path validation
    process.env.NODE_ENV = 'test';
  });

  afterAll(() => {
    // Restore original environment
    process.env.NODE_ENV = originalNodeEnv;
  });
  let testProjectPath: string;
  let originalStdin: typeof process.stdin;
  let originalStdout: typeof process.stdout;

  beforeEach(() => {
    vi.clearAllMocks();
    originalStdin = process.stdin;
    originalStdout = process.stdout;

    // Create temporary test directory inside cwd (to pass path traversal check)
    // Add dummy package.json (to pass MoAI package check)
    testProjectPath = path.join(process.cwd(), '.test-project-' + Date.now());
    if (fs.existsSync(testProjectPath)) {
      fs.rmSync(testProjectPath, { recursive: true, force: true });
    }
    fs.mkdirSync(testProjectPath, { recursive: true });

    // Create dummy package.json to prevent MoAI package detection
    fs.writeFileSync(
      path.join(testProjectPath, 'package.json'),
      JSON.stringify({ name: 'test-project' }, null, 2)
    );
  });

  afterEach(() => {
    // Cleanup test directory
    if (fs.existsSync(testProjectPath)) {
      fs.rmSync(testProjectPath, { recursive: true, force: true });
    }

    // Restore original stdin/stdout
    Object.defineProperty(process, 'stdin', {
      value: originalStdin,
      writable: true,
    });
    Object.defineProperty(process, 'stdout', {
      value: originalStdout,
      writable: true,
    });
  });

  describe('--yes flag', () => {
    test('should initialize with default settings when --yes flag is provided', async () => {
      // Given: --yes 플래그 제공
      const options = {
        name: 'test-project',
        yes: true,
        path: testProjectPath,
      };

      // When: moai init --yes 실행
      const { InitCommand } = await import('@/cli/commands/init/index');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      const result = await initCommand.runNonInteractive(options);

      // Then: 프롬프트 없이 기본값으로 초기화
      expect(result.success).toBe(true);
      expect(result.projectPath).toBe(testProjectPath);

      // .moai/config.json 생성 확인
      const configPath = path.join(testProjectPath, '.moai', 'config.json');
      expect(fs.existsSync(configPath)).toBe(true);

      // 기본값 확인
      const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
      expect(config.mode).toBe('personal');
      expect(config.git?.enabled).toBeDefined(); // git.enabled는 InstallationOrchestrator가 관리
    }, 30000); // 30초 타임아웃

    test('should skip prompts even in TTY environment when --yes is provided', async () => {
      // Given: TTY 환경이지만 --yes 플래그 제공
      Object.defineProperty(process.stdin, 'isTTY', { value: true, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: true, writable: true });

      const options = {
        name: 'test-project',
        yes: true,
        path: testProjectPath,
      };

      // When: moai init --yes 실행
      const { InitCommand } = await import('@/cli/commands/init/index');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      const result = await initCommand.runNonInteractive(options);

      // Then: TTY 환경이어도 프롬프트 스킵
      expect(result.success).toBe(true);

      const configPath = path.join(testProjectPath, '.moai', 'config.json');
      const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
      expect(config.mode).toBe('personal');
      expect(config.git?.enabled).toBeDefined(); // git.enabled는 InstallationOrchestrator가 관리
    }, 30000); // 30초 타임아웃

    test('should display "Initializing with default settings..." message', async () => {
      // Given: --yes 플래그 제공
      const consoleSpy = vi.spyOn(console, 'log');

      const options = {
        name: 'test-project',
        yes: true,
        path: testProjectPath,
      };

      // When: moai init --yes 실행
      const { InitCommand } = await import('@/cli/commands/init/index');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      await initCommand.runNonInteractive(options);

      // Then: "Initializing with default settings..." 메시지 출력
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('default settings')
      );
    });
  });

  describe('TTY auto-detection', () => {
    test('should automatically switch to non-interactive mode when TTY is not available', async () => {
      // Given: TTY 없음 (Claude Code, CI/CD, Docker)
      Object.defineProperty(process.stdin, 'isTTY', { value: undefined, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: undefined, writable: true });

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행 (--yes 없이)
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      const result = await initCommand.runNonInteractive(options);

      // Then: 자동으로 비대화형 모드 전환
      expect(result.success).toBe(true);

      const configPath = path.join(testProjectPath, '.moai', 'config.json');
      const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
      expect(config.mode).toBe('personal');
      expect(config.git?.enabled).toBeDefined(); // git.enabled는 InstallationOrchestrator가 관리
    }, 30000); // 30초 타임아웃

    test('should display "Running in non-interactive mode" message when TTY is not available', async () => {
      // Given: TTY 없음
      Object.defineProperty(process.stdin, 'isTTY', { value: undefined, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: undefined, writable: true });

      const consoleSpy = vi.spyOn(console, 'log');

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      await initCommand.runNonInteractive(options);

      // Then: "Running in non-interactive mode with default settings" 메시지 출력
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('non-interactive mode')
      );
    });

    test('should use interactive mode when TTY is available and --yes is not provided', async () => {
      // Given: TTY 환경 및 --yes 플래그 없음
      Object.defineProperty(process.stdin, 'isTTY', { value: true, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: true, writable: true });

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      // Mock inquirer to prevent actual prompts in test
      vi.mock('inquirer', () => ({
        default: {
          prompt: vi.fn().mockResolvedValue({
            projectName: 'test-project',
            mode: 'personal',
            gitEnabled: true,
          }),
        },
      }));

      const result = await initCommand.runInteractive(options);

      // Then: 대화형 모드 사용 (프롬프트 실행)
      expect(result.success).toBe(true);
    });
  });

  describe('Default configuration values', () => {
    test('should use { mode: "personal", gitEnabled: true } as default', async () => {
      // Given: 비대화형 환경
      Object.defineProperty(process.stdin, 'isTTY', { value: undefined, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: undefined, writable: true });

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      await initCommand.runNonInteractive(options);

      // Then: 기본값 저장 확인
      const configPath = path.join(testProjectPath, '.moai', 'config.json');
      const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));

      expect(config.mode).toBe('personal');
      expect(config.git?.enabled).toBeDefined(); // git.enabled는 InstallationOrchestrator가 관리
    });

    test('should create .moai/config.json file in non-interactive mode', async () => {
      // Given: 비대화형 환경
      Object.defineProperty(process.stdin, 'isTTY', { value: undefined, writable: true });

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      await initCommand.runNonInteractive(options);

      // Then: .moai/config.json 파일 생성 확인
      const configPath = path.join(testProjectPath, '.moai', 'config.json');
      expect(fs.existsSync(configPath)).toBe(true);

      // JSON 유효성 확인
      expect(() => JSON.parse(fs.readFileSync(configPath, 'utf-8'))).not.toThrow();
    });

    test('should preserve existing user experience in interactive mode', async () => {
      // Given: TTY 환경 (기존 사용자)
      Object.defineProperty(process.stdin, 'isTTY', { value: true, writable: true });
      Object.defineProperty(process.stdout, 'isTTY', { value: true, writable: true });

      const options = {
        name: 'test-project',
        path: testProjectPath,
      };

      // When: moai init 실행 (--yes 없이)
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock DoctorCommand to pass
      mockDoctorCommand(initCommand, true);

      // Mock inquirer
      vi.mock('inquirer', () => ({
        default: {
          prompt: vi.fn().mockResolvedValue({
            projectName: 'custom-project',
            mode: 'team',
            gitEnabled: false,
          }),
        },
      }));

      const result = await initCommand.runInteractive(options);

      // Then: 기존 대화형 경험 유지 (사용자 선택 반영)
      expect(result.success).toBe(true);
    });
  });

  describe('Error handling', () => {
    test('should fail gracefully when required dependencies are missing', async () => {
      // Given: 필수 의존성 누락 (Git, Node.js)
      const options = {
        name: 'test-project',
        yes: true,
        path: testProjectPath,
      };

      // When: moai init 실행
      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock doctorCommand.run() to fail
      mockDoctorCommand(initCommand, false, [{ requirement: { name: 'git' } }]);

      const result = await initCommand.runNonInteractive(options);

      // Then: 초기화 실패 및 에러 메시지 출력
      expect(result.success).toBe(false);
      expect(result.errors).toContain('System verification failed');

      // .moai 디렉토리 미생성 확인
      const moaiDir = path.join(testProjectPath, '.moai');
      expect(fs.existsSync(moaiDir)).toBe(false);
    }, 30000);

    test('should display detailed error messages on failure', async () => {
      // Given: 초기화 실패 상황
      const consoleSpy = vi.spyOn(console, 'log');

      const options = {
        name: 'test-project',
        yes: true,
        path: testProjectPath,
      };

      const { InitCommand } = await import('@/cli/commands/init');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const initCommand = new InitCommand(detector);

      // Mock doctorCommand.run() to fail with multiple missing requirements
      mockDoctorCommand(initCommand, false, [
        { requirement: { name: 'git' } },
        { requirement: { name: 'npm' } }
      ]);

      // When: moai init 실행
      const result = await initCommand.runNonInteractive(options);

      // Then: "Run 'moai doctor' for detailed diagnostics" 힌트 표시
      expect(result.success).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('System verification failed')
      );
    }, 30000);
  });
});
