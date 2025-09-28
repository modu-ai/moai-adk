/**
 * @file Test for system requirement detector
 * @author MoAI Team
 * @tags @TEST:SYSTEM-DETECTOR-001 @REQ:AUTO-VERIFY-012
 */

import { describe, test, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { SystemDetector } from '@/core/system-checker/detector';
import { SystemRequirement } from '@/core/system-checker/requirements';

// Mock execa for command execution
jest.mock('execa', () => ({
  execa: jest.fn()
}));

describe('SystemDetector', async () => {
  let detector: SystemDetector;
  const mockExeca = jest.mocked((await import('execa')).default);

  beforeEach(() => {
    detector = new SystemDetector();
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe('단일 요구사항 검증', () => {
    const gitRequirement: SystemRequirement = {
      name: 'Git',
      category: 'runtime',
      minVersion: '2.30.0',
      installCommands: {
        darwin: 'brew install git',
        linux: 'sudo apt-get install git',
        win32: 'winget install Git.Git'
      },
      checkCommand: 'git --version',
      versionCommand: 'git --version'
    };

    test('should detect installed requirement with valid version', async () => {
      // Given: Git이 설치되어 있고 유효한 버전
      mockExeca.mockResolvedValueOnce({
        stdout: 'git version 2.35.1',
        stderr: '',
        exitCode: 0,
        command: 'git --version',
        escapedCommand: 'git --version',
        failed: false,
        timedOut: false,
        isCanceled: false,
        killed: false
      });

      // When: Git 검증 수행
      const result = await detector.checkRequirement(gitRequirement);

      // Then: 설치됨으로 감지되어야 함
      expect(result.isInstalled).toBe(true);
      expect(result.detectedVersion).toBe('2.35.1');
      expect(result.versionSatisfied).toBe(true);
      expect(result.error).toBeUndefined();
    });

    test('should detect installed requirement with insufficient version', async () => {
      // Given: Git이 설치되어 있지만 버전이 부족함
      mockExeca.mockResolvedValueOnce({
        stdout: 'git version 2.25.0',
        stderr: '',
        exitCode: 0,
        command: 'git --version',
        escapedCommand: 'git --version',
        failed: false,
        timedOut: false,
        isCanceled: false,
        killed: false
      });

      // When: Git 검증 수행
      const result = await detector.checkRequirement(gitRequirement);

      // Then: 설치되어 있지만 버전이 부족함으로 감지
      expect(result.isInstalled).toBe(true);
      expect(result.detectedVersion).toBe('2.25.0');
      expect(result.versionSatisfied).toBe(false);
      expect(result.error).toBeUndefined();
    });

    test('should detect missing requirement', async () => {
      // Given: Git이 설치되어 있지 않음
      const error = new Error('Command failed');
      (error as any).exitCode = 1;
      mockExeca.mockRejectedValueOnce(error);

      // When: Git 검증 수행
      const result = await detector.checkRequirement(gitRequirement);

      // Then: 설치되지 않음으로 감지
      expect(result.isInstalled).toBe(false);
      expect(result.detectedVersion).toBeUndefined();
      expect(result.versionSatisfied).toBe(false);
      expect(result.error).toBeDefined();
    });

    test('should handle version parsing errors gracefully', async () => {
      // Given: 버전 파싱이 불가능한 출력
      mockExeca.mockResolvedValueOnce({
        stdout: 'invalid version output',
        stderr: '',
        exitCode: 0,
        command: 'git --version',
        escapedCommand: 'git --version',
        failed: false,
        timedOut: false,
        isCanceled: false,
        killed: false
      });

      // When: Git 검증 수행
      const result = await detector.checkRequirement(gitRequirement);

      // Then: 설치되어 있지만 버전 검증 실패
      expect(result.isInstalled).toBe(true);
      expect(result.detectedVersion).toBeUndefined();
      expect(result.versionSatisfied).toBe(false);
    });
  });

  describe('여러 요구사항 일괄 검증', () => {
    const requirements: SystemRequirement[] = [
      {
        name: 'Git',
        category: 'runtime',
        minVersion: '2.30.0',
        installCommands: { darwin: 'brew install git' },
        checkCommand: 'git --version',
        versionCommand: 'git --version'
      },
      {
        name: 'Node.js',
        category: 'runtime',
        minVersion: '18.0.0',
        installCommands: { darwin: 'brew install node' },
        checkCommand: 'node --version',
        versionCommand: 'node --version'
      }
    ];

    test('should check multiple requirements concurrently', async () => {
      // Given: Git은 설치됨, Node.js는 설치되지 않음
      mockExeca
        .mockResolvedValueOnce({
          stdout: 'git version 2.35.1',
          stderr: '',
          exitCode: 0,
          command: 'git --version',
          escapedCommand: 'git --version',
          failed: false,
          timedOut: false,
          isCanceled: false,
          killed: false
        })
        .mockRejectedValueOnce(new Error('node: command not found'));

      // When: 여러 요구사항 검증
      const results = await detector.checkMultipleRequirements(requirements);

      // Then: 각각 올바르게 검증되어야 함
      expect(results).toHaveLength(2);

      const gitResult = results.find(r => r.requirement.name === 'Git');
      const nodeResult = results.find(r => r.requirement.name === 'Node.js');

      expect(gitResult?.result.isInstalled).toBe(true);
      expect(nodeResult?.result.isInstalled).toBe(false);
    });

    test('should return empty array for empty requirements', async () => {
      // Given: 빈 요구사항 배열
      // When: 빈 배열로 검증
      const results = await detector.checkMultipleRequirements([]);

      // Then: 빈 결과 배열 반환
      expect(results).toEqual([]);
    });
  });

  describe('플랫폼 감지', () => {
    test('should detect current platform', () => {
      // Given: 현재 실행 환경
      // When: 플랫폼 감지
      const platform = detector.getCurrentPlatform();

      // Then: 유효한 플랫폼이 반환되어야 함
      expect(['darwin', 'linux', 'win32']).toContain(platform);
    });

    test('should return install command for current platform', () => {
      // Given: 플랫폼별 설치 명령어가 정의된 요구사항
      const requirement: SystemRequirement = {
        name: 'TestTool',
        category: 'runtime',
        installCommands: {
          darwin: 'brew install testtool',
          linux: 'apt-get install testtool',
          win32: 'winget install testtool'
        },
        checkCommand: 'testtool --version'
      };

      // When: 현재 플랫폼용 설치 명령어 조회
      const installCommand = detector.getInstallCommandForCurrentPlatform(requirement);

      // Then: 유효한 설치 명령어가 반환되어야 함
      expect(installCommand).toBeDefined();
      expect(typeof installCommand).toBe('string');
      expect(installCommand!.length).toBeGreaterThan(0);
    });

    test('should return undefined for unsupported platform', () => {
      // Given: 현재 플랫폼이 지원되지 않는 요구사항
      const requirement: SystemRequirement = {
        name: 'TestTool',
        category: 'runtime',
        installCommands: {
          freebsd: 'pkg install testtool' // 지원되지 않는 플랫폼
        } as any,
        checkCommand: 'testtool --version'
      };

      // When: 현재 플랫폼용 설치 명령어 조회
      const installCommand = detector.getInstallCommandForCurrentPlatform(requirement);

      // Then: undefined가 반환되어야 함
      expect(installCommand).toBeUndefined();
    });
  });
});