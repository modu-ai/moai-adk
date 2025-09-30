// @CODE:INST-001 | 
// Related: @CODE:INST-001:API, @CODE:INST-CTX-001

/**
 * @file Installation orchestration coordinator
 * @author MoAI Team
 */

import { logger } from '@/utils/winston-logger';
import { InstallationError } from '@/utils/errors';
import type {
  InstallationConfig,
  InstallationContext,
  InstallationResult,
  ProgressCallback,
} from './types';
import { ContextManager } from './context-manager';
import { PhaseExecutor } from './phase-executor';
import { ResultBuilder } from './result-builder';

/**
 * Central coordinator for MoAI-ADK installation (Refactored)
 *
 * Responsibilities:
 * - Coordinate installation phases
 * - Delegate execution to PhaseExecutor
 * - Manage context via ContextManager
 * - Build results via ResultBuilder
 *
 * @tags @CODE:INSTALL-ORCHESTRATOR-001
 */
export class InstallationOrchestrator {
  private readonly config: InstallationConfig;
  private context: InstallationContext;

  // Dependency Injection
  private readonly contextManager: ContextManager;
  private readonly phaseExecutor: PhaseExecutor;
  private readonly resultBuilder: ResultBuilder;

  constructor(config: InstallationConfig) {
    this.config = config;

    // Initialize dependencies
    this.contextManager = new ContextManager();
    this.phaseExecutor = new PhaseExecutor(this.contextManager);
    this.resultBuilder = new ResultBuilder();

    // Create initial context
    this.context = this.contextManager.createInitialContext(config);

    logger.debug('InstallationOrchestrator initialized', {
      projectPath: config.projectPath,
      mode: config.mode,
      tag: '@INIT:ORCHESTRATOR-001',
    });
  }

  /**
   * Execute complete MoAI-ADK project installation
   * @param progressCallback Progress callback function
   * @returns Complete installation result
   * @tags @CODE:EXECUTE-INSTALLATION-001:API
   */
  public async executeInstallation(
    progressCallback?: ProgressCallback
  ): Promise<InstallationResult> {
    const startTime = Date.now();

    try {
      // Execute all phases sequentially
      await this.phaseExecutor.executePreparationPhase(
        this.context,
        progressCallback
      );

      await this.phaseExecutor.executeDirectoryPhase(
        this.context,
        progressCallback
      );

      await this.phaseExecutor.executeResourcePhase(
        this.context,
        progressCallback
      );

      await this.phaseExecutor.executeConfigurationPhase(
        this.context,
        progressCallback
      );

      await this.phaseExecutor.executeValidationPhase(
        this.context,
        progressCallback
      );

      // Update final progress
      this.contextManager.updateProgress(
        this.context,
        'Installation complete!',
        5,
        progressCallback
      );

      // Build success result
      return this.resultBuilder.createSuccessResult(this.context, startTime);
    } catch (error) {
      const installError = error instanceof InstallationError
        ? error
        : new InstallationError('Installation failed', {
            error: error instanceof Error ? error : undefined,
            errorMessage: error instanceof Error ? error.message : String(error),
          });

      logger.error('Installation failed', {
        error: installError.message,
        tag: '@ERROR:INSTALL-FAILED-001',
      });

      // Build failure result
      return this.resultBuilder.createFailureResult(
        this.context,
        startTime,
        installError
      );
    }
  }

  /**
   * Get current installation context (for testing/debugging)
   * @returns Current context
   * @tags @CODE:GET-CONTEXT-001:API
   */
  public getContext(): InstallationContext {
    return this.context;
  }

  /**
   * Get installation configuration (for testing/debugging)
   * @returns Installation configuration
   * @tags @CODE:GET-CONFIG-001:API
   */
  public getConfig(): InstallationConfig {
    return this.config;
  }
}