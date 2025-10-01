// @CODE:PHASE-RUNNER-001 |
// Related: @CODE:INST-002

/**
 * @file Phase execution template pattern
 * @author MoAI Team
 */

import type { ContextManager } from '../context-manager';
import type { InstallationContext, ProgressCallback } from '../types';

/**
 * Phase execution result
 */
export interface PhaseExecutionResult {
  filesCreated: string[];
  errors: string[];
}

/**
 * Phase execution handler
 */
export type PhaseHandler = () => Promise<string[]>;

/**
 * Execute a phase with common pattern (template method)
 */
export async function executePhase(
  phaseName: string,
  phaseLabel: string,
  context: InstallationContext,
  contextManager: ContextManager,
  handler: PhaseHandler,
  progressCallback?: ProgressCallback
): Promise<void> {
  const phaseStartTime = Date.now();
  const filesCreated: string[] = [];
  const errors: string[] = [];

  // Update progress
  contextManager.updateProgress(context, phaseLabel, 5, progressCallback);

  try {
    // Execute phase-specific logic
    const files = await handler();
    filesCreated.push(...files);

    // Record success
    contextManager.recordPhaseCompletion(
      context,
      phaseName,
      phaseStartTime,
      filesCreated,
      errors
    );
  } catch (error) {
    // Record failure
    errors.push(`${phaseName} phase failed: ${error}`);
    contextManager.recordPhaseCompletion(
      context,
      phaseName,
      phaseStartTime,
      filesCreated,
      errors
    );
    throw error;
  }
}
