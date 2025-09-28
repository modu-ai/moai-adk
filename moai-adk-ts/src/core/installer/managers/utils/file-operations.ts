/**
 * @REFACTOR:FILE-OPERATIONS-001 파일 작업 유틸리티
 *
 * @TASK:FILE-HANDLING-001 파일 복사 및 권한 관리만 담당하는 전용 클래스
 * TRUST-U 원칙: 단일 책임 및 50 LOC 이하 메서드 준수
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { logger } from '../../../../utils/logger';

/**
 * @DESIGN:CLASS-001 파일 작업 전용 클래스
 *
 * 파일 복사, 디렉토리 생성, 권한 설정 등의 파일 시스템 작업을 캡슐화합니다.
 */
export class FileOperations {
  /**
   * @API:COPY-DIRECTORY-001 디렉토리 재귀 복사
   *
   * @param sourcePath 소스 디렉토리 경로
   * @param targetPath 대상 디렉토리 경로
   * @param overwrite 덮어쓰기 여부
   * @param excludeSubdirs 제외할 하위 디렉토리들
   */
  async copyDirectory(
    sourcePath: string,
    targetPath: string,
    overwrite: boolean,
    excludeSubdirs?: string[]
  ): Promise<void> {
    const entries = await fs.readdir(sourcePath, { withFileTypes: true });
    await fs.mkdir(targetPath, { recursive: true });

    for (const entry of entries) {
      if (this.shouldExcludeEntry(entry.name, excludeSubdirs)) {
        continue;
      }

      await this.copyEntry(
        entry,
        sourcePath,
        targetPath,
        overwrite,
        excludeSubdirs
      );
    }
  }

  /**
   * @API:ENSURE-PERMISSIONS-001 훅 파일 실행 권한 보장
   *
   * @param claudeRoot Claude 루트 디렉토리
   */
  async ensureHookPermissions(claudeRoot: string): Promise<void> {
    try {
      const hooksDir = path.join(claudeRoot, 'hooks', 'moai');
      const pythonFiles = await this.findPythonFiles(hooksDir);

      for (const pyFile of pythonFiles) {
        await this.setExecutablePermission(pyFile);
      }
    } catch (error) {
      logger.warn(
        `Failed to ensure hook permissions under ${claudeRoot}: ${error}`
      );
    }
  }

  /**
   * @API:VALIDATE-CLEAN-001 깨끗한 설치 검증
   *
   * @param targetPath 검증할 대상 경로
   * @returns 깨끗한 설치 여부
   */
  async validateCleanInstallation(targetPath: string): Promise<boolean> {
    try {
      const checks = [
        this.validateEmptySpecs(targetPath),
        this.validateInitialTags(targetPath),
        this.validateEmptyReports(targetPath),
      ];

      const results = await Promise.all(checks);
      const isClean = results.every(result => result);

      if (isClean) {
        logger.debug(
          `Clean installation validated successfully: ${targetPath}`
        );
      }

      return isClean;
    } catch (error) {
      logger.error(
        `Failed to validate clean installation at ${targetPath}: ${error}`
      );
      return false;
    }
  }

  // Private helper methods

  /**
   * @TASK:EXCLUDE-ENTRY-001 엔트리 제외 여부 확인
   */
  private shouldExcludeEntry(
    entryName: string,
    excludeSubdirs?: string[]
  ): boolean {
    return excludeSubdirs?.includes(entryName) ?? false;
  }

  /**
   * @TASK:COPY-ENTRY-001 개별 엔트리 복사
   */
  private async copyEntry(
    entry: any,
    sourcePath: string,
    targetPath: string,
    overwrite: boolean,
    excludeSubdirs?: string[]
  ): Promise<void> {
    const sourceEntryPath = path.join(sourcePath, entry.name);
    const targetEntryPath = path.join(targetPath, entry.name);

    if (entry.isDirectory()) {
      await this.copyDirectory(
        sourceEntryPath,
        targetEntryPath,
        overwrite,
        excludeSubdirs
      );
    } else {
      await this.copyFileWithPolicy(
        sourceEntryPath,
        targetEntryPath,
        overwrite
      );
    }
  }

  /**
   * @TASK:COPY-FILE-POLICY-001 파일 복사 정책 적용
   */
  private async copyFileWithPolicy(
    sourcePath: string,
    targetPath: string,
    overwrite: boolean
  ): Promise<void> {
    try {
      await fs.access(targetPath);
      if (!overwrite) {
        logger.debug(`File exists, skipping: ${targetPath}`);
        return;
      }
    } catch {
      // 파일이 존재하지 않음 - 복사 진행
    }

    await fs.copyFile(sourcePath, targetPath);
  }

  /**
   * @TASK:FIND-PYTHON-001 Python 파일 찾기
   */
  private async findPythonFiles(hooksDir: string): Promise<string[]> {
    try {
      const stat = await fs.stat(hooksDir);
      if (!stat.isDirectory()) {
        return [];
      }

      const entries = await fs.readdir(hooksDir);
      return entries
        .filter(entry => entry.endsWith('.py'))
        .map(entry => path.join(hooksDir, entry));
    } catch {
      return [];
    }
  }

  /**
   * @TASK:SET-EXECUTABLE-001 실행 권한 설정
   */
  private async setExecutablePermission(filePath: string): Promise<void> {
    try {
      // 0o755: 소유자 실행/읽기/쓰기, 그룹/기타 실행/읽기
      await fs.chmod(filePath, 0o755);
      logger.debug(`Set executable permission: ${filePath}`);
    } catch (error) {
      logger.warn(`Failed to chmod +x ${filePath}: ${error}`);
    }
  }

  /**
   * @TASK:VALIDATE-SPECS-001 specs 디렉토리 검증
   */
  private async validateEmptySpecs(targetPath: string): Promise<boolean> {
    const specsDir = path.join(targetPath, 'specs');
    return this.validateEmptyDirectory(specsDir, 'spec files');
  }

  /**
   * @TASK:VALIDATE-TAGS-001 tags.db 초기 구조 검증
   */
  private async validateInitialTags(targetPath: string): Promise<boolean> {
    try {
      const tagsFile = path.join(targetPath, 'indexes', 'tags.db');
      const stats = await fs.stat(tagsFile);
      const fileSize = stats.size;

      // SQLite3 파일이 너무 크면 개발 데이터가 포함되었을 가능성
      if (fileSize > 50 * 1024) {
        // 50KB 제한
        logger.warn(
          `tags.db seems to contain development data: ${fileSize} bytes (expected < 50KB)`
        );
        return false;
      }
      return true;
    } catch {
      return true; // tags.db가 없음 - 정상 (처음 설치)
    }
  }

  /**
   * @TASK:VALIDATE-REPORTS-001 reports 디렉토리 검증
   */
  private async validateEmptyReports(targetPath: string): Promise<boolean> {
    const reportsDir = path.join(targetPath, 'reports');
    return this.validateEmptyDirectory(reportsDir, 'report files');
  }

  /**
   * @TASK:VALIDATE-EMPTY-DIR-001 빈 디렉토리 검증
   */
  private async validateEmptyDirectory(
    dirPath: string,
    fileType: string
  ): Promise<boolean> {
    try {
      const entries = await fs.readdir(dirPath);
      const nonGitkeepFiles = entries.filter(f => f !== '.gitkeep');

      if (nonGitkeepFiles.length > 0) {
        logger.warn(
          `Found unexpected ${fileType} in clean installation: ${nonGitkeepFiles}`
        );
        return false;
      }
      return true;
    } catch {
      return true; // 디렉토리가 없음 - 정상
    }
  }
}
