// @CODE:CLI-003 |
// Related: @CODE:GIT-001:API, @CODE:GIT-STATUS-001

/**
 * @file CLI status command for project status (Refactored)
 * @author MoAI Team
 * @description 리팩토링된 StatusCommand - 책임 분리 및 모듈화
 */

import chalk from 'chalk';
import { logger } from '../../utils/winston-logger.js';
import { FileCounter, type FileCount } from './status/collectors/file-counter.js';
import { StatusCollector, type ProjectStatus } from './status/collectors/status-collector.js';
import { VersionCollector, type VersionInfo } from './status/collectors/version-collector.js';
import { StatusFormatter } from './status/formatters/status-formatter.js';

/**
 * Status command options
 * @tags @SPEC:STATUS-OPTIONS-001
 */
export interface StatusOptions {
  readonly verbose: boolean;
  readonly projectPath?: string | undefined;
}

/**
 * Complete project status with all information
 * @tags @SPEC:COMPLETE-STATUS-001
 */
export interface CompleteProjectStatus extends ProjectStatus {
  readonly versions?: VersionInfo;
  readonly fileCounts?: FileCount | undefined;
}

/**
 * Status result
 * @tags @SPEC:STATUS-RESULT-001
 */
export interface StatusResult {
  readonly success: boolean;
  readonly status?: CompleteProjectStatus;
  readonly recommendations?: string[];
  readonly error?: string;
}

/**
 * Status command for project status display
 * @tags @CODE:CLI-STATUS-001
 *
 * 리팩토링 완료:
 * - 이전: 368 LOC 단일 파일
 * - 이후: 120 LOC 명령어 + 4개 전문 모듈
 * - 책임 분리: 상태 수집, 버전 수집, 파일 카운팅, 출력 포맷팅
 */
export class StatusCommand {
  private readonly statusCollector: StatusCollector;
  private readonly versionCollector: VersionCollector;
  private readonly fileCounter: FileCounter;
  private readonly formatter: StatusFormatter;

  constructor() {
    this.statusCollector = new StatusCollector();
    this.versionCollector = new VersionCollector();
    this.fileCounter = new FileCounter();
    this.formatter = new StatusFormatter();
  }

  /**
   * Run status command
   * @param options - Status options
   * @returns Status result
   * @tags @CODE:STATUS-RUN-001:API
   */
  public async run(options: StatusOptions): Promise<StatusResult> {
    try {
      const projectPath = options.projectPath || process.cwd();

      // Step 1: Collect project status
      const status = await this.statusCollector.collectStatus(projectPath);
      this.formatter.displayStatus(status);

      // Step 2: Collect and display version information
      const versions = await this.versionCollector.collectVersionInfo(projectPath);
      this.formatter.displayVersions(versions);

      // Step 3: Collect and display file counts if verbose
      let fileCounts: FileCount | undefined;
      if (options.verbose) {
        fileCounts = await this.fileCounter.countProjectFiles(projectPath);
        this.formatter.displayFileCounts(fileCounts);
      }

      // Step 4: Generate and display recommendations
      const recommendations = this.formatter.generateRecommendations(status);
      this.formatter.displayRecommendations(recommendations);

      // Build complete status
      const completeStatus: CompleteProjectStatus = {
        ...status,
        versions,
        fileCounts,
      };

      return {
        success: true,
        status: completeStatus,
        recommendations,
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.info(
        chalk.red(`❌ Failed to get project status: ${errorMessage}`)
      );

      return {
        success: false,
        error: errorMessage,
      };
    }
  }
}

// Re-export types for backward compatibility
export type { VersionInfo, FileCount, ProjectStatus };
