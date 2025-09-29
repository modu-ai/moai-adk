/**
 * @file Installation Process Orchestrator (Refactored)
 * @author MoAI Team
 * @tags @FEATURE:INSTALL-ORCHESTRATOR-001 @REQ:INSTALL-SYSTEM-012
 */

import { logger } from '@/utils/logger';
import { PhaseExecutor } from './phases/phase-executor';
import { ResourceInstaller } from './phases/resource-installer';
import type {
  InstallationConfig,
  InstallationContext,
  InstallationResult,
  PhaseStatus,
  ProgressCallback,
} from './types';

/**
 * Central coordinator for MoAI-ADK installation (Refactored)
 * @tags @FEATURE:INSTALL-ORCHESTRATOR-001
 */
export class InstallationOrchestrator {
  private readonly config: InstallationConfig;
  private context: InstallationContext;
  private readonly phaseExecutor: PhaseExecutor;
  private readonly resourceInstaller: ResourceInstaller;

  constructor(config: InstallationConfig) {
    this.config = config;
    this.context = this.createInitialContext(config);
    this.phaseExecutor = new PhaseExecutor(this.config, this.context);
    this.resourceInstaller = new ResourceInstaller(this.config, this.context);

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
   * @tags @API:EXECUTE-INSTALLATION-001
   */
  public async executeInstallation(
    progressCallback?: ProgressCallback
  ): Promise<InstallationResult> {
    const startTime = Date.now();

    try {
      await this.phaseExecutor.executePreparationPhase(progressCallback);
      await this.phaseExecutor.executeDirectoryPhase(progressCallback);
      await this.phaseExecutor.executeResourcePhase(progressCallback);
      await this.phaseExecutor.executeConfigurationPhase(progressCallback);
      await this.phaseExecutor.executeValidationPhase(progressCallback);

      this.updateProgress('Installation complete!', progressCallback);

      return this.createSuccessResult(startTime);
    } catch (error) {
      logger.error('Installation failed', {
        error,
        tag: '@ERROR:INSTALL-FAILED-001',
      });
      return this.createFailureResult(startTime, error);
    }
  }

  /**
   * Get current installation context
   * @returns Installation context
   * @tags @API:GET-CONTEXT-001
   */
  public getContext(): InstallationContext {
    return { ...this.context };
  }

  /**
   * Create initial installation context
   * @param config Installation configuration
   * @returns Initial context
   * @tags @UTIL:CREATE-CONTEXT-001
   */
  private createInitialContext(config: InstallationConfig): InstallationContext {
    return {
      projectPath: config.projectPath,
      mode: config.mode,
      startTime: Date.now(),
      phases: [],
      errors: [],
      filesCreated: [],
      metadata: {
        moaiVersion: '0.0.3',
        nodeVersion: process.version,
        platform: process.platform,
        architecture: process.arch,
      },
    };
  }

  /**
   * Update progress and notify callback
   * @param message Progress message
   * @param callback Progress callback
   * @tags @UTIL:UPDATE-PROGRESS-001
   */
  private updateProgress(message: string, callback?: ProgressCallback): void {
    logger.info(message, { tag: '@PROGRESS:ORCHESTRATOR-001' });

    if (callback) {
      const completedPhases = this.context.phases.filter(p => p.completed).length;
      const totalPhases = 5; // preparation, directories, resources, configuration, validation
      const percentage = Math.round((completedPhases / totalPhases) * 100);

      callback({
        phase: 'in-progress',
        message,
        percentage,
        details: {
          filesCreated: this.context.filesCreated,
          errors: this.context.errors,
          nextSteps: this.generateNextSteps(),
        },
      });
    }
  }

  /**
   * Create successful installation result
   * @param startTime Installation start time
   * @returns Success result
   * @tags @UTIL:CREATE-SUCCESS-RESULT-001
   */
  private createSuccessResult(startTime: number): InstallationResult {
    const duration = Date.now() - startTime;
    const totalFiles = this.context.filesCreated.length;

    return {
      success: true,
      message: `MoAI-ADK installation completed successfully in ${duration}ms`,
      duration,
      context: this.context,
      summary: {
        phases: this.context.phases.length,
        filesCreated: totalFiles,
        errors: this.context.errors.length,
        nextSteps: this.generateNextSteps(),
      },
    };
  }

  /**
   * Create failure installation result
   * @param startTime Installation start time
   * @param error Error that caused failure
   * @returns Failure result
   * @tags @UTIL:CREATE-FAILURE-RESULT-001
   */
  private createFailureResult(
    startTime: number,
    error: unknown
  ): InstallationResult {
    const duration = Date.now() - startTime;
    const errorMessage = error instanceof Error ? error.message : String(error);

    this.context.errors.push(errorMessage);

    return {
      success: false,
      message: `MoAI-ADK installation failed: ${errorMessage}`,
      duration,
      context: this.context,
      summary: {
        phases: this.context.phases.length,
        filesCreated: this.context.filesCreated.length,
        errors: this.context.errors.length,
        nextSteps: [
          'Check the error logs above',
          'Verify system requirements with: moai doctor',
          'Try running the installation again',
          'Report the issue if the problem persists',
        ],
      },
    };
  }

  /**
   * Generate next steps based on installation state
   * @returns Array of next step descriptions
   * @tags @UTIL:GENERATE-NEXT-STEPS-001
   */
  private generateNextSteps(): string[] {
    const steps: string[] = [];

    if (this.context.phases.some(p => p.name === 'validation' && p.completed)) {
      steps.push('Run "moai status" to verify installation');
      steps.push('Start developing with "moai init" or explore the documentation');
      steps.push('Use "/moai:1-spec" command to create your first specification');
    } else {
      steps.push('Installation is still in progress...');
    }

    if (this.config.mode === 'team') {
      steps.push('Configure GitHub integration for team collaboration');
    }

    return steps;
  }
}