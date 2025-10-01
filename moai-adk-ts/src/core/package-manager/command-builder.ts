// @CODE:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Command builder for package managers
 * @author MoAI Team
 * @tags @CODE:COMMAND-BUILDER-001
 */

import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';

/**
 * Builds command strings for various package managers
 * @tags @CODE:COMMAND-BUILDER-001:FEATURE
 */
export class CommandBuilder {
  /**
   * Build install command for package manager
   * @param packages - Package names to install
   * @param options - Installation options
   * @returns Command string
   * @tags @CODE:BUILD-INSTALL-CMD-001:API
   */
  public buildInstallCommand(
    packages: string[],
    options: PackageInstallOptions
  ): string {
    const { packageManager, isDevelopment, isGlobal } = options;

    switch (packageManager) {
      case PackageManagerType.NPM:
        return this.buildNpmInstallCommand(packages, isDevelopment, isGlobal);
      case PackageManagerType.YARN:
        return this.buildYarnInstallCommand(packages, isDevelopment, isGlobal);
      case PackageManagerType.PNPM:
        return this.buildPnpmInstallCommand(packages, isDevelopment, isGlobal);
      default:
        throw new Error(`Unsupported package manager: ${packageManager}`);
    }
  }

  /**
   * Build run command for package manager
   * @param packageManagerType - Package manager type
   * @returns Run command string
   * @tags @CODE:BUILD-RUN-CMD-001:API
   */
  public buildRunCommand(packageManagerType: PackageManagerType): string {
    switch (packageManagerType) {
      case PackageManagerType.NPM:
        return 'npm run';
      case PackageManagerType.YARN:
        return 'yarn run';
      case PackageManagerType.PNPM:
        return 'pnpm run';
      default:
        return 'npm run';
    }
  }

  /**
   * Build test command
   * @param packageManagerType - Package manager type
   * @param testingFramework - Optional testing framework
   * @returns Test command string
   * @tags @CODE:BUILD-TEST-CMD-001:API
   */
  public buildTestCommand(
    packageManagerType: PackageManagerType,
    testingFramework?: string
  ): string {
    if (testingFramework === 'jest') {
      return 'jest';
    }

    switch (packageManagerType) {
      case PackageManagerType.NPM:
        return 'npm test';
      case PackageManagerType.YARN:
        return 'yarn test';
      case PackageManagerType.PNPM:
        return 'pnpm test';
      default:
        return 'npm test';
    }
  }

  /**
   * Build init command
   * @param packageManagerType - Package manager type
   * @returns Init command string
   * @tags @CODE:BUILD-INIT-CMD-001:API
   */
  public buildInitCommand(packageManagerType: PackageManagerType): string {
    switch (packageManagerType) {
      case PackageManagerType.NPM:
        return 'npm init -y';
      case PackageManagerType.YARN:
        return 'yarn init -y';
      case PackageManagerType.PNPM:
        return 'pnpm init';
      default:
        return 'npm init -y';
    }
  }

  /**
   * Get package manager engine requirement
   * @param packageManagerType - Package manager type
   * @returns Engine requirements
   * @tags @CODE:GET-ENGINE-001:API
   */
  public getPackageManagerEngine(
    packageManagerType: PackageManagerType
  ): Record<string, string> {
    switch (packageManagerType) {
      case PackageManagerType.NPM:
        return { npm: '>=9.0.0' };
      case PackageManagerType.YARN:
        return { yarn: '>=1.22.0' };
      case PackageManagerType.PNPM:
        return { pnpm: '>=8.0.0' };
      default:
        return {};
    }
  }

  /**
   * Build npm install command
   * @private
   */
  private buildNpmInstallCommand(
    packages: string[],
    isDevelopment?: boolean,
    isGlobal?: boolean
  ): string {
    const flags: string[] = [];
    if (isDevelopment) flags.push('--save-dev');
    if (isGlobal) flags.push('--global');

    return this.buildCommandWithFlags('npm install', packages, flags);
  }

  /**
   * Build yarn install command
   * @private
   */
  private buildYarnInstallCommand(
    packages: string[],
    isDevelopment?: boolean,
    isGlobal?: boolean
  ): string {
    if (isGlobal) {
      return this.buildCommandWithFlags('yarn global add', packages, []);
    }

    const flags: string[] = [];
    if (isDevelopment) flags.push('--dev');

    return this.buildCommandWithFlags('yarn add', packages, flags);
  }

  /**
   * Build pnpm install command
   * @private
   */
  private buildPnpmInstallCommand(
    packages: string[],
    isDevelopment?: boolean,
    isGlobal?: boolean
  ): string {
    const flags: string[] = [];
    if (isDevelopment) flags.push('--save-dev');
    if (isGlobal) flags.push('--global');

    return this.buildCommandWithFlags('pnpm add', packages, flags);
  }

  /**
   * Build command string with flags
   * @private
   */
  private buildCommandWithFlags(
    baseCommand: string,
    packages: string[],
    flags: string[]
  ): string {
    const commandParts = [baseCommand, ...flags, ...packages];
    return commandParts.join(' ');
  }
}
