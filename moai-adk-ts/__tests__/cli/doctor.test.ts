/**
 * @file Test for CLI doctor command
 * @author MoAI Team
 * @tags @TEST:CLI-DOCTOR-001 @REQ:CLI-FOUNDATION-012
 */

import { describe, test, expect, jest, beforeEach } from '@jest/globals';
import { DoctorCommand } from '@/cli/commands/doctor';
import { SystemDetector } from '@/core/system-checker/detector';

// Mock dependencies
jest.mock('@/core/system-checker/detector');
jest.mock('chalk', () => ({
  green: jest.fn(str => str),
  red: jest.fn(str => str),
  yellow: jest.fn(str => str),
  blue: jest.fn(str => str),
  bold: jest.fn(str => str)
}));

describe('DoctorCommand', () => {
  let doctorCommand: DoctorCommand;
  let mockDetector: jest.Mocked<SystemDetector>;

  beforeEach(() => {
    mockDetector = jest.mocked(new SystemDetector());
    doctorCommand = new DoctorCommand(mockDetector);
    jest.clearAllMocks();
  });

  describe('시스템 진단 실행', () => {
    test('should check all runtime requirements', async () => {
      // Given: 모든 런타임 요구사항이 만족됨
      mockDetector.checkMultipleRequirements.mockResolvedValueOnce([
        {
          requirement: {
            name: 'Git',
            category: 'runtime',
            installCommands: { darwin: 'brew install git' },
            checkCommand: 'git --version',
            minVersion: '2.30.0'
          },
          result: {
            isInstalled: true,
            detectedVersion: '2.35.1',
            versionSatisfied: true
          }
        },
        {
          requirement: {
            name: 'Node.js',
            category: 'runtime',
            installCommands: { darwin: 'brew install node' },
            checkCommand: 'node --version',
            minVersion: '18.0.0'
          },
          result: {
            isInstalled: true,
            detectedVersion: '18.17.0',
            versionSatisfied: true
          }
        }
      ]);

      // When: 진단 실행
      const result = await doctorCommand.run();

      // Then: 모든 요구사항이 검사되고 성공 결과 반환
      expect(mockDetector.checkMultipleRequirements).toHaveBeenCalledTimes(1);
      expect(result.allPassed).toBe(true);
      expect(result.results).toHaveLength(2);
    });

    test('should identify missing requirements', async () => {
      // Given: 일부 요구사항이 누락됨
      mockDetector.checkMultipleRequirements.mockResolvedValueOnce([
        {
          requirement: {
            name: 'Git',
            category: 'runtime',
            installCommands: { darwin: 'brew install git' },
            checkCommand: 'git --version',
            minVersion: '2.30.0'
          },
          result: {
            isInstalled: true,
            detectedVersion: '2.35.1',
            versionSatisfied: true
          }
        },
        {
          requirement: {
            name: 'Node.js',
            category: 'runtime',
            installCommands: { darwin: 'brew install node' },
            checkCommand: 'node --version',
            minVersion: '18.0.0'
          },
          result: {
            isInstalled: false,
            versionSatisfied: false,
            error: 'Command not found'
          }
        }
      ]);

      // When: 진단 실행
      const result = await doctorCommand.run();

      // Then: 실패 결과와 누락된 요구사항 식별
      expect(result.allPassed).toBe(false);
      expect(result.missingRequirements).toHaveLength(1);
      expect(result.missingRequirements[0]?.requirement.name).toBe('Node.js');
    });

    test('should identify version conflicts', async () => {
      // Given: 버전이 부족한 요구사항
      mockDetector.checkMultipleRequirements.mockResolvedValueOnce([
        {
          requirement: {
            name: 'Git',
            category: 'runtime',
            installCommands: { darwin: 'brew install git' },
            checkCommand: 'git --version',
            minVersion: '2.30.0'
          },
          result: {
            isInstalled: true,
            detectedVersion: '2.25.0',
            versionSatisfied: false
          }
        }
      ]);

      // When: 진단 실행
      const result = await doctorCommand.run();

      // Then: 버전 충돌 식별
      expect(result.allPassed).toBe(false);
      expect(result.versionConflicts).toHaveLength(1);
      expect(result.versionConflicts[0]?.requirement.name).toBe('Git');
    });
  });

  describe('진단 결과 포맷팅', () => {
    test('should format successful check results', () => {
      // Given: 성공한 검사 결과
      const checkResult = {
        requirement: {
          name: 'Git',
          category: 'runtime' as const,
          installCommands: { darwin: 'brew install git' },
          checkCommand: 'git --version',
          minVersion: '2.30.0'
        },
        result: {
          isInstalled: true,
          detectedVersion: '2.35.1',
          versionSatisfied: true
        }
      };

      // When: 결과 포맷팅
      const formatted = doctorCommand.formatCheckResult(checkResult);

      // Then: 성공 표시가 포함되어야 함
      expect(formatted).toContain('✅');
      expect(formatted).toContain('Git');
      expect(formatted).toContain('2.35.1');
    });

    test('should format failed check results with error', () => {
      // Given: 실패한 검사 결과
      const checkResult = {
        requirement: {
          name: 'Node.js',
          category: 'runtime' as const,
          installCommands: { darwin: 'brew install node' },
          checkCommand: 'node --version'
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
          error: 'Command not found'
        }
      };

      // When: 결과 포맷팅
      const formatted = doctorCommand.formatCheckResult(checkResult);

      // Then: 실패 표시와 에러 메시지가 포함되어야 함
      expect(formatted).toContain('❌');
      expect(formatted).toContain('Node.js');
      expect(formatted).toContain('Command not found');
    });

    test('should provide installation suggestions', () => {
      // Given: 누락된 요구사항
      const checkResult = {
        requirement: {
          name: 'SQLite3',
          category: 'development' as const,
          installCommands: {
            darwin: 'brew install sqlite3',
            linux: 'sudo apt-get install sqlite3',
            win32: 'winget install SQLite.SQLite'
          },
          checkCommand: 'sqlite3 --version'
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
          error: 'Command not found'
        }
      };

      mockDetector.getCurrentPlatform.mockReturnValue('darwin');
      mockDetector.getInstallCommandForCurrentPlatform.mockReturnValue('brew install sqlite3');

      // When: 설치 제안 생성
      const suggestion = doctorCommand.getInstallationSuggestion(checkResult);

      // Then: 플랫폼별 설치 명령어 제안
      expect(suggestion).toContain('brew install sqlite3');
      expect(suggestion).toContain('SQLite3');
    });
  });

  describe('출력 및 사용자 인터페이스', () => {
    test('should provide summary of all checks', async () => {
      // Given: 혼합된 검사 결과 (성공/실패)
      mockDetector.checkMultipleRequirements.mockResolvedValueOnce([
        {
          requirement: {
            name: 'Git',
            category: 'runtime',
            installCommands: { darwin: 'brew install git' },
            checkCommand: 'git --version'
          },
          result: { isInstalled: true, versionSatisfied: true }
        },
        {
          requirement: {
            name: 'Node.js',
            category: 'runtime',
            installCommands: { darwin: 'brew install node' },
            checkCommand: 'node --version'
          },
          result: { isInstalled: false, versionSatisfied: false, error: 'Not found' }
        }
      ]);

      // When: 진단 실행
      const result = await doctorCommand.run();

      // Then: 요약 정보가 포함되어야 함
      expect(result.summary).toBeDefined();
      expect(result.summary.total).toBe(2);
      expect(result.summary.passed).toBe(1);
      expect(result.summary.failed).toBe(1);
    });
  });
});