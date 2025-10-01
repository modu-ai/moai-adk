// @CODE:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Package manager installer (Refactored)
 * @author MoAI Team
 * @tags @CODE:PACKAGE-MANAGER-INSTALLER-001
 */

import { execa } from 'execa';
import type {
  PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';
import { CommandBuilder } from './command-builder';

/**
 * Result of package installation operation
 * @tags @CODE:INSTALL-RESULT-001:DATA
 */
export interface InstallResult {
  success: boolean;
  installedPackages: string[];
  error?: string;
  output?: string;
}

/**
 * Result of project initialization operation
 * @tags @CODE:INIT-RESULT-001:DATA
 */
export interface InitResult {
  success: boolean;
  packageJsonPath?: string;
  error?: string;
  output?: string;
}

/**
 * Orchestrates package installation and project initialization
 * @tags @CODE:PACKAGE-MANAGER-INSTALLER-001:FEATURE
 */
export class PackageManagerInstaller {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder?: CommandBuilder) {
    this.commandBuilder = commandBuilder ?? new CommandBuilder();
  }

  /**
   * Install packages using specified package manager
   * @param packages - Package names to install
   * @param options - Installation options
   * @returns Installation result
   * @tags @CODE:INSTALL-PACKAGES-001:API
   */
  public async installPackages(
    packages: string[],
    options: PackageInstallOptions
  ): Promise<InstallResult> {
    try {
      const command = this.commandBuilder.buildInstallCommand(
        packages,
        options
      );
      const [executable, ...args] = command.split(' ');

      const result = await execa(executable!, args, {
        cwd: options.workingDirectory || process.cwd(),
        timeout: 300000, // 5 minutes timeout
        reject: false,
      });

      if (result.exitCode === 0) {
        return {
          success: true,
          installedPackages: packages,
          output: result.stdout,
        };
      } else {
        return {
          success: false,
          installedPackages: [],
          error:
            result.stderr || `Command failed with exit code ${result.exitCode}`,
          output: result.stdout,
        };
      }
    } catch (error) {
      return {
        success: false,
        installedPackages: [],
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  /**
   * Initialize project with package manager
   * @param projectPath - Project directory path
   * @param packageManagerType - Package manager to use
   * @returns Initialization result
   * @tags @CODE:INIT-PROJECT-001:API
   */
  public async initializeProject(
    projectPath: string,
    packageManagerType: PackageManagerType
  ): Promise<InitResult> {
    try {
      const command = this.commandBuilder.buildInitCommand(packageManagerType);
      const [executable, ...args] = command.split(' ');

      const result = await execa(executable!, args, {
        cwd: projectPath,
        timeout: 30000,
        reject: false,
      });

      if (result.exitCode === 0) {
        return {
          success: true,
          packageJsonPath: `${projectPath}/package.json`,
          output: result.stdout,
        };
      } else {
        return {
          success: false,
          error:
            result.stderr || `Command failed with exit code ${result.exitCode}`,
          output: result.stdout,
        };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }
}
