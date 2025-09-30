/**
 * @file Installation Context Manager
 * @author MoAI Team
 * @tags @FEATURE:CONTEXT-MANAGER-001 @REQ:INSTALL-SYSTEM-012
 */

import { logger } from '@/utils/logger';
import type {
  InstallationConfig,
  InstallationContext,
  PhaseStatus,
  ProgressCallback,
} from './types';

/**
 * Manages installation context and progress tracking
 * @tags @FEATURE:CONTEXT-MANAGER-001
 */
export class ContextManager {
  /**
   * Create initial installation context
   * @param config Installation configuration
   * @returns Initial context
   * @tags @API:CREATE-CONTEXT-001
   */
  createInitialContext(config: InstallationConfig): InstallationContext {
    return {
      config,
      startTime: new Date(),
      phases: [],
      allFilesCreated: [],
      allErrors: [],
    };
  }

  /**
   * Record phase completion
   * @param context Current context
   * @param phaseName Phase name
   * @param startTime Phase start time
   * @param filesCreated Files created in phase
   * @param errors Errors in phase
   * @tags @API:RECORD-PHASE-001
   */
  recordPhaseCompletion(
    context: InstallationContext,
    phaseName: string,
    startTime: number,
    filesCreated: string[],
    errors: string[]
  ): void {
    const phase: PhaseStatus = {
      name: phaseName,
      completed: errors.length === 0,
      duration: Date.now() - startTime,
      errors: [...errors],
      filesCreated: [...filesCreated],
    };

    context.phases.push(phase);
    context.allFilesCreated.push(...filesCreated);
    context.allErrors.push(...errors);

    logger.debug('Phase recorded', {
      phase: phaseName,
      completed: phase.completed,
      duration: phase.duration,
      tag: '@DEBUG:PHASE-RECORD-001',
    });
  }

  /**
   * Update progress and notify callback
   * @param context Current context
   * @param message Progress message
   * @param totalPhases Total number of phases
   * @param callback Progress callback
   * @tags @API:UPDATE-PROGRESS-001
   */
  updateProgress(
    context: InstallationContext,
    message: string,
    totalPhases: number,
    callback?: ProgressCallback
  ): void {
    const current = context.phases.length;

    logger.debug(message, {
      current,
      total: totalPhases,
      tag: '@PROGRESS:UPDATE-001',
    });

    if (callback) {
      callback(message, current, totalPhases);
    }
  }

  /**
   * Get current phase count
   * @param context Current context
   * @returns Number of completed phases
   * @tags @API:GET-PHASE-COUNT-001
   */
  getPhaseCount(context: InstallationContext): number {
    return context.phases.length;
  }

  /**
   * Check if context has errors
   * @param context Current context
   * @returns True if context has errors
   * @tags @API:HAS-ERRORS-001
   */
  hasErrors(context: InstallationContext): boolean {
    return context.allErrors.length > 0;
  }

  /**
   * Get total duration from context
   * @param context Current context
   * @returns Duration in milliseconds
   * @tags @API:GET-DURATION-001
   */
  getDuration(context: InstallationContext): number {
    return Date.now() - context.startTime.getTime();
  }
}