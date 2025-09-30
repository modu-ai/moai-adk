/**
 * @file Installation Result Builder
 * @author MoAI Team
 * @tags @CODE:RESULT-BUILDER-001 @SPEC:INSTALL-SYSTEM-012
 */

import { logger } from '@/utils/winston-logger';
import { InstallationError, getErrorMessage } from '@/utils/errors';
import type {
  InstallationConfig,
  InstallationContext,
  InstallationResult,
} from './types';

/**
 * Builds installation results and generates next steps
 * @tags @CODE:RESULT-BUILDER-001
 */
export class ResultBuilder {
  /**
   * Create success result
   * @param context Installation context
   * @param startTime Installation start time
   * @returns Success result
   * @tags @CODE:CREATE-SUCCESS-RESULT-001:API
   */
  createSuccessResult(
    context: InstallationContext,
    startTime: number
  ): InstallationResult {
    const result: InstallationResult = {
      success: true,
      projectPath: context.config.projectPath,
      filesCreated: [...context.allFilesCreated],
      errors: [...context.allErrors],
      nextSteps: this.generateNextSteps(context.config),
      config: context.config,
      timestamp: new Date(),
      duration: Date.now() - startTime,
    };

    logger.info('Installation succeeded', {
      projectPath: result.projectPath,
      filesCreated: result.filesCreated.length,
      duration: result.duration,
      tag: '@SUCCESS:INSTALL-001',
    });

    return result;
  }

  /**
   * Create failure result
   * @param context Installation context
   * @param startTime Installation start time
   * @param error Failure error
   * @returns Failure result
   * @tags @CODE:CREATE-FAILURE-RESULT-001:API
   */
  createFailureResult(
    context: InstallationContext,
    startTime: number,
    error: unknown
  ): InstallationResult {
    const errorMessage = error instanceof Error ? error.message : String(error);

    const result: InstallationResult = {
      success: false,
      projectPath: context.config.projectPath,
      filesCreated: [...context.allFilesCreated],
      errors: [
        ...context.allErrors,
        `Installation failed: ${errorMessage}`,
      ],
      nextSteps: ['Fix the errors above and retry installation'],
      config: context.config,
      timestamp: new Date(),
      duration: Date.now() - startTime,
    };

    logger.error('Installation failed', {
      projectPath: result.projectPath,
      error: errorMessage,
      duration: result.duration,
      tag: '@ERROR:INSTALL-001',
    });

    return result;
  }

  /**
   * Generate next steps for user
   * @param config Installation configuration
   * @returns Array of next steps
   * @tags @CODE:GENERATE-NEXT-STEPS-001:API
   */
  private generateNextSteps(config: InstallationConfig): string[] {
    const steps = [
      `cd ${config.projectName}`,
      `💡 Tip: Run "claude" to start development with Claude Code`,
    ];

    return steps;
  }

  /**
   * Build summary message from context
   * @param context Installation context
   * @returns Summary message
   * @tags @CODE:BUILD-SUMMARY-001:API
   */
  buildSummary(context: InstallationContext): string {
    const total = context.phases.length;
    const completed = context.phases.filter(p => p.completed).length;
    const failed = total - completed;
    const duration = Date.now() - context.startTime.getTime();

    return `Completed ${total} phases (${completed} passed, ${failed} failed) in ${duration}ms`;
  }
}