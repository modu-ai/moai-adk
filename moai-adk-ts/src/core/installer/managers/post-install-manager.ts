/**
 * @feature POST-INSTALL-001 Post-Installation Manager
 * @task POST-INSTALL-001 Handles post-installation setup and validation
 * @design TRUST-COMPLIANT Single responsibility for post-install operations
 */

import chalk from 'chalk';
import { logger } from '../../../utils/logger';
import { ResourceValidator } from './resource-validator';
import { FirstRunManager } from './first-run-manager';
import { GlobalSetupManager } from './global-setup-manager';
import type { PostInstallOptions, PostInstallResult } from '../types';

/**
 * Post-Installation Manager
 *
 * Manages the complete post-installation process including:
 * - Resource validation
 * - First run detection and setup
 * - Global configuration setup
 * - User environment initialization
 *
 * @tags @FEATURE:POST-INSTALL-001 @DESIGN:MANAGER-PATTERN-001
 */
export class PostInstallManager {
  private readonly resourceValidator: ResourceValidator;
  private readonly firstRunManager: FirstRunManager;
  private readonly globalSetupManager: GlobalSetupManager;

  constructor() {
    this.resourceValidator = new ResourceValidator();
    this.firstRunManager = new FirstRunManager();
    this.globalSetupManager = new GlobalSetupManager();
    logger.debug('PostInstallManager initialized');
  }

  /**
   * Execute complete post-installation process
   *
   * @param options - Post-installation configuration options
   * @returns Promise resolving to post-installation result
   * @tags @TASK:POST-INSTALL-EXECUTION-001
   */
  async runPostInstall(
    options: PostInstallOptions
  ): Promise<PostInstallResult> {
    const startTime = Date.now();
    const errors: string[] = [];
    const warnings: string[] = [];

    try {
      logger.info('Starting post-installation process', {
        projectPath: options.projectPath,
        setupGlobal: options.setupGlobal,
        validateResources: options.validateResources,
      });

      // Display banner unless quiet mode
      if (!options.quiet) {
        this.printPostInstallBanner('0.0.1');
      }

      // 1. Check if this is first run
      const isFirstRun = this.firstRunManager.isFirstRun();
      logger.debug('First run detection', { isFirstRun });

      // 2. Validate resources if requested
      let resourcesValidated = true;
      if (options.validateResources) {
        const validation =
          await this.resourceValidator.validatePackageResources();
        resourcesValidated = validation.isValid;

        if (!validation.isValid) {
          errors.push('Resource validation failed');
          errors.push(...validation.errors);
          logger.error('Resource validation failed', undefined, {
            missingTemplates: validation.missingTemplates,
            errors: validation.errors,
          });
        }
      }

      // 3. Setup global configuration if requested
      let globalSetupCompleted = true;
      if (options.setupGlobal && resourcesValidated) {
        try {
          await this.globalSetupManager.setupGlobalConfig();
          await this.globalSetupManager.registerGlobalCommands();
          logger.info('Global setup completed successfully');
        } catch (error) {
          globalSetupCompleted = false;
          const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
          errors.push(`Global setup failed: ${errorMessage}`);
          logger.error('Global setup failed', undefined, {
            error: errorMessage,
          });
        }
      }

      // 4. Initialize first run setup if needed
      let firstRunSetupCompleted = true;
      if (isFirstRun && resourcesValidated) {
        try {
          await this.firstRunManager.initializeFirstRun();
          await this.firstRunManager.setupUserEnvironment();
          await this.firstRunManager.createFirstRunMarker();
          logger.info('First run setup completed successfully');
        } catch (error) {
          firstRunSetupCompleted = false;
          const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
          errors.push(`First run setup failed: ${errorMessage}`);
          logger.error('First run setup failed', undefined, {
            error: errorMessage,
          });
        }
      }

      const duration = Date.now() - startTime;
      const success = errors.length === 0;

      logger.info('Post-installation process completed', {
        success,
        duration,
        errorCount: errors.length,
        warningCount: warnings.length,
      });

      return {
        success,
        isFirstRun,
        resourcesValidated,
        globalSetupCompleted,
        firstRunSetupCompleted,
        errors,
        warnings,
        duration,
        timestamp: new Date(),
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.error('Post-installation process failed', undefined, {
        error: errorMessage,
        duration,
      });

      return {
        success: false,
        isFirstRun: false,
        resourcesValidated: false,
        globalSetupCompleted: false,
        firstRunSetupCompleted: false,
        errors: [`Post-installation failed: ${errorMessage}`],
        warnings,
        duration,
        timestamp: new Date(),
      };
    }
  }

  /**
   * Validate package resources availability
   *
   * @returns Promise resolving to validation success status
   * @tags @TASK:RESOURCE-VALIDATION-001
   */
  async validateResources(): Promise<boolean> {
    try {
      const result = await this.resourceValidator.validatePackageResources();
      logger.debug('Resource validation completed', {
        isValid: result.isValid,
      });
      return result.isValid;
    } catch (error) {
      logger.error(
        'Resource validation error',
        error instanceof Error ? error : undefined,
        { error: error instanceof Error ? error.message : String(error) }
      );
      return false;
    }
  }

  /**
   * Check if auto-install can proceed on first run
   *
   * Implements the logic from Python's auto_install_on_first_run()
   *
   * @returns Promise resolving to auto-install feasibility
   * @tags @TASK:AUTO-INSTALL-CHECK-001
   */
  async autoInstallOnFirstRun(): Promise<boolean> {
    try {
      const result = await this.resourceValidator.checkRequiredTemplates();
      logger.debug('Auto-install check completed', {
        allPresent: result.allPresent,
        templateCount: result.templateCount,
      });
      return result.allPresent;
    } catch (error) {
      logger.error(
        'Auto-install check failed',
        error instanceof Error ? error : undefined,
        { error: error instanceof Error ? error.message : String(error) }
      );
      return false;
    }
  }

  /**
   * Setup global resources and configuration
   *
   * @returns Promise resolving when global setup is complete
   * @tags @TASK:GLOBAL-SETUP-001
   */
  async setupGlobalResources(): Promise<void> {
    try {
      await this.globalSetupManager.setupGlobalConfig();
      await this.globalSetupManager.registerGlobalCommands();
      logger.info('Global resources setup completed');
    } catch (error) {
      logger.error(
        'Global resources setup failed',
        error instanceof Error ? error : undefined,
        { error: error instanceof Error ? error.message : String(error) }
      );
      throw error;
    }
  }

  /**
   * Print post-installation banner
   *
   * Implements the functionality from Python's print_post_install_banner()
   *
   * @param version - MoAI-ADK version string
   * @param quiet - Whether to suppress output
   * @tags @TASK:BANNER-DISPLAY-001
   */
  printPostInstallBanner(version: string, quiet = false): void {
    if (quiet) {
      return;
    }

    const banner = `
${chalk.cyan(`🗿 MoAI-ADK v${version} Post-Installation`)}
${chalk.blue('━'.repeat(48))}

Setting up global resources for optimal MoAI-ADK experience...
`;

    console.log(banner);
    logger.info('Post-installation banner displayed', { version });
  }
}
