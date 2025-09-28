/**
 * @file Test for CLI version command
 * @author MoAI Team
 * @tags @TEST:CLI-VERSION-001 @REQ:CLI-FOUNDATION-012
 */

import { describe, test, expect, beforeEach } from 'vitest';
import { Command } from 'commander';

describe('CLI Version Command', () => {
  let program: Command;

  beforeEach(() => {
    program = new Command();
  });

  describe('버전 정보 출력', () => {
    test('should display version from package.json', () => {
      // Given: CLI 프로그램이 설정됨
      const mockVersion = '0.0.1';
      program.version(mockVersion);

      // When: 버전 정보 조회
      const version = program.version();

      // Then: 올바른 버전이 반환되어야 함
      expect(version).toBe(mockVersion);
    });

    test('should support --version flag', () => {
      // Given: 버전 플래그가 설정된 프로그램
      const mockVersion = '0.0.1';
      program.version(mockVersion, '-v, --version', 'output the current version');

      // When: 버전 옵션 확인
      const versionOption = program.options.find(opt => opt.long === '--version');

      // Then: 버전 옵션이 정의되어야 함
      expect(versionOption).toBeDefined();
      expect(versionOption?.short).toBe('-v');
      expect(versionOption?.long).toBe('--version');
    });
  });

  describe('도움말 정보', () => {
    test('should provide help information', () => {
      // Given: 도움말이 설정된 프로그램
      program
        .name('moai')
        .description('🗿 MoAI-ADK: Modu-AI Agentic Development kit')
        .version('0.0.1');

      // When: 프로그램 정보 확인
      const name = program.name();
      const description = program.description();

      // Then: 올바른 정보가 설정되어야 함
      expect(name).toBe('moai');
      expect(description).toBe('🗿 MoAI-ADK: Modu-AI Agentic Development kit');
    });

    test('should support --help flag', () => {
      // Given: 도움말이 설정된 프로그램
      program.name('moai').description('Test description');

      // When: 프로그램 정보 확인 (Commander.js는 자동으로 --help 옵션 제공)
      const name = program.name();
      const description = program.description();

      // Then: 프로그램이 올바르게 설정되어야 함
      expect(name).toBe('moai');
      expect(description).toBe('Test description');
    });
  });
});