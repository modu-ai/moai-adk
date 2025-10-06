// @TEST:INIT-001 | SPEC: SPEC-INIT-001.md | CODE: src/core/installer/dependency-installer.ts
// Related: @CODE:INIT-001:INSTALLER, @SPEC:INIT-001

/**
 * @file Test for Dependency Installer
 * @author MoAI Team
 * @tags @TEST:INIT-001:INSTALLER
 */

import { describe, test, expect } from 'vitest';
import { DependencyInstaller } from '@/core/installer/dependency-installer';

describe('Dependency Installer', () => {
  describe('installGit', () => {
    test('should provide manual installation guide on Windows', async () => {
      // Given: Windows 환경
      const platform = 'win32';

      // When: Git 자동 설치 실행
      const installer = new DependencyInstaller();
      const result = await installer.installGit(platform);

      // Then: 수동 설치 가이드 제공 (자동 설치 미지원)
      expect(result).toBe(false);
    });

    test('should handle unsupported platform gracefully', async () => {
      // Given: 지원하지 않는 플랫폼
      const platform = 'unknown';

      // When: Git 자동 설치 실행
      const installer = new DependencyInstaller();
      const result = await installer.installGit(platform);

      // Then: 실패 처리
      expect(result).toBe(false);
    });
  });

  describe('installNodeJS', () => {
    test('should provide manual installation guide on Windows', async () => {
      // Given: Windows 환경
      const platform = 'win32';

      // When: Node.js 자동 설치 실행
      const installer = new DependencyInstaller();
      const result = await installer.installNodeJS(platform);

      // Then: 수동 설치 가이드 제공
      expect(result).toBe(false);
    });

    test('should handle unsupported platform gracefully', async () => {
      // Given: 지원하지 않는 플랫폼
      const platform = 'unknown';

      // When: Node.js 자동 설치 실행
      const installer = new DependencyInstaller();
      const result = await installer.installNodeJS(platform);

      // Then: 실패 처리
      expect(result).toBe(false);
    });
  });

  describe('DependencyInstaller class', () => {
    test('should be instantiable', () => {
      // When: DependencyInstaller 인스턴스 생성
      const installer = new DependencyInstaller();

      // Then: 정상 생성
      expect(installer).toBeInstanceOf(DependencyInstaller);
    });

    test('should have installGit method', () => {
      // Given: DependencyInstaller 인스턴스
      const installer = new DependencyInstaller();

      // Then: installGit 메서드 존재
      expect(typeof installer.installGit).toBe('function');
    });

    test('should have installNodeJS method', () => {
      // Given: DependencyInstaller 인스턴스
      const installer = new DependencyInstaller();

      // Then: installNodeJS 메서드 존재
      expect(typeof installer.installNodeJS).toBe('function');
    });
  });
});
