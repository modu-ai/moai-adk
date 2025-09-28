/**
 * @REFACTOR:PATH-VALIDATOR-001 경로 안전성 검증 유틸리티
 *
 * @TASK:SINGLE-RESPONSIBILITY-001 경로 검증만 담당하는 전용 클래스
 * TRUST-U 원칙: 단일 책임 및 50 LOC 이하 메서드 준수
 */

import { tmpdir } from 'os';
import * as path from 'path';
import { logger } from '../../../../utils/logger';

/**
 * @DESIGN:CLASS-001 경로 안전성 검증 전용 클래스
 *
 * 보안을 위한 경로 검증 로직을 캡슐화합니다.
 */
export class PathValidator {
  private readonly dangerousPaths: readonly string[];
  private readonly tmpDirectory: string;

  constructor() {
    // 시스템 중요 디렉토리 목록 (Unix 계열)
    this.dangerousPaths = [
      '/etc',
      '/usr/bin',
      '/usr/sbin',
      '/boot',
      '/sys',
      '/proc',
    ];
    this.tmpDirectory = tmpdir();
  }

  /**
   * @API:VALIDATE-SAFE-PATH-001 경로 안전성 검증
   *
   * @param targetPath 검증할 경로
   * @returns 안전한 경로 여부
   */
  validateSafePath(targetPath: string): boolean {
    try {
      if (this.hasPathTraversal(targetPath)) {
        logger.warn(`Path traversal detected in: ${targetPath}`);
        return false;
      }

      if (this.isTempDirectory(targetPath)) {
        return true; // 임시 디렉토리는 항상 허용
      }

      if (this.isDangerousSystemPath(targetPath)) {
        logger.warn(`Attempt to write to dangerous system path: ${targetPath}`);
        return false;
      }

      return true;
    } catch (error) {
      logger.warn(`Path validation failed for ${targetPath}: ${error}`);
      return false;
    }
  }

  /**
   * @TASK:PATH-TRAVERSAL-001 경로 순회 공격 탐지
   */
  private hasPathTraversal(targetPath: string): boolean {
    const resolvedPath = path.resolve(targetPath);
    return targetPath.includes('..') || resolvedPath.includes('..');
  }

  /**
   * @TASK:TEMP-DIR-001 임시 디렉토리 확인
   */
  private isTempDirectory(targetPath: string): boolean {
    const resolvedPath = path.resolve(targetPath);
    return resolvedPath.startsWith(this.tmpDirectory);
  }

  /**
   * @TASK:DANGEROUS-PATH-001 위험한 시스템 경로 확인
   */
  private isDangerousSystemPath(targetPath: string): boolean {
    const resolvedPath = path.resolve(targetPath);
    return this.dangerousPaths.some(dangerous =>
      resolvedPath.startsWith(dangerous)
    );
  }
}
