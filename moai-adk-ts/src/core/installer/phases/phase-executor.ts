/**
 * @file Installation Phase Execution Logic
 * @author MoAI Team
 * @tags @FEATURE:PHASE-EXECUTOR-001 @REQ:INSTALL-SYSTEM-012
 */

import * as path from 'node:path';
import { logger } from '@/utils/logger';
import type {
  InstallationConfig,
  InstallationContext,
  PhaseStatus,
  ProgressCallback,
} from '../types';

/**
 * Handles execution of installation phases
 * @tags @FEATURE:PHASE-EXECUTOR-001
 */
export class PhaseExecutor {
  constructor(
    private readonly config: InstallationConfig,
    private readonly context: InstallationContext
  ) {}

  /**
   * Execute preparation phase including backup creation
   * @param progressCallback Progress callback
   * @tags @PHASE:PREPARATION-001
   */
  async executePreparationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 1: Preparation and backup...', progressCallback);

    try {
      if (this.config.backupEnabled) {
        await this.createBackup();
      }

      await this.validateSystemRequirements();

      this.recordPhaseCompletion(
        'preparation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Preparation phase failed: ${error}`);
      this.recordPhaseCompletion(
        'preparation',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute directory creation phase
   * @param progressCallback Progress callback
   * @tags @PHASE:DIRECTORY-001
   */
  async executeDirectoryPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 2: Creating directories...', progressCallback);

    try {
      await this.createProjectDirectories();

      this.recordPhaseCompletion(
        'directories',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Directory phase failed: ${error}`);
      this.recordPhaseCompletion(
        'directories',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute resource installation phase
   * @param progressCallback Progress callback
   * @tags @PHASE:RESOURCE-001
   */
  async executeResourcePhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    let filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 3: Installing resources...', progressCallback);

    try {
      filesCreated = await this.installResources();

      this.recordPhaseCompletion(
        'resources',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Resource phase failed: ${error}`);
      this.recordPhaseCompletion(
        'resources',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute configuration generation phase
   * @param progressCallback Progress callback
   * @tags @PHASE:CONFIGURATION-001
   */
  async executeConfigurationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    let filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress(
      'Phase 4: Generating configurations...',
      progressCallback
    );

    try {
      filesCreated = await this.generateConfigurations();

      this.recordPhaseCompletion(
        'configuration',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Configuration phase failed: ${error}`);
      this.recordPhaseCompletion(
        'configuration',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute validation and finalization phase
   * @param progressCallback Progress callback
   * @tags @PHASE:VALIDATION-001
   */
  async executeValidationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 5: Validating installation...', progressCallback);

    try {
      await this.validateInstallation();
      await this.finalizeInstallation();

      this.recordPhaseCompletion(
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Validation phase failed: ${error}`);
      this.recordPhaseCompletion(
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  private updateProgress(message: string, callback?: ProgressCallback): void {
    logger.info(message, { tag: '@PROGRESS:PHASE-001' });

    if (callback) {
      callback({
        phase: 'in-progress',
        message,
        percentage: 0,
        details: {
          filesCreated: [],
          errors: [],
          nextSteps: [],
        },
      });
    }
  }

  private recordPhaseCompletion(
    phaseName: string,
    startTime: number,
    filesCreated: string[],
    errors: string[]
  ): void {
    const duration = Date.now() - startTime;
    const status: PhaseStatus = {
      name: phaseName,
      completed: errors.length === 0,
      duration,
      filesCreated,
      errors,
    };

    this.context.phases.push(status);

    logger.info(`Phase ${phaseName} completed`, {
      duration,
      success: status.completed,
      filesCount: filesCreated.length,
      errorsCount: errors.length,
      tag: '@PHASE:COMPLETED-001',
    });
  }

  // These methods would be implemented by delegating to other classes
  private async createBackup(): Promise<void> {
    // Implementation moved to BackupManager
    throw new Error('Not implemented - delegate to BackupManager');
  }

  private async validateSystemRequirements(): Promise<void> {
    // Implementation moved to SystemValidator
    throw new Error('Not implemented - delegate to SystemValidator');
  }

  private async createProjectDirectories(): Promise<void> {
    // Implementation moved to DirectoryManager
    throw new Error('Not implemented - delegate to DirectoryManager');
  }

  private async installResources(): Promise<string[]> {
    // Implementation moved to ResourceInstaller
    throw new Error('Not implemented - delegate to ResourceInstaller');
  }

  private async generateConfigurations(): Promise<string[]> {
    // Implementation moved to ConfigurationGenerator
    throw new Error('Not implemented - delegate to ConfigurationGenerator');
  }

  private async validateInstallation(): Promise<void> {
    // Implementation moved to InstallationValidator
    throw new Error('Not implemented - delegate to InstallationValidator');
  }

  private async finalizeInstallation(): Promise<void> {
    // Implementation moved to InstallationFinalizer
    throw new Error('Not implemented - delegate to InstallationFinalizer');
  }
}