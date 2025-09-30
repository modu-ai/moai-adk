// @CODE:PKG-002 | 
// Related: @CODE:PKG-002:API, @CODE:PKG-CFG-001

/**
 * @file Package manager installer
 * @author MoAI Team
 */

import { execa } from 'execa';
import {
  type PackageInstallOptions,
  type PackageJsonConfig,
  PackageManagerType,
} from '@/types/package-manager';

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
 * Package manager installer for dependency management and project setup
 * @tags @CODE:PACKAGE-MANAGER-INSTALLER-001
 */
export class PackageManagerInstaller {
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
      const command = this.buildInstallCommand(packages, options);
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
   * Generate package.json configuration
   * @param projectConfig - Project configuration
   * @param packageManagerType - Package manager type
   * @param includeTypeScript - Whether to include TypeScript setup
   * @param testingFramework - Testing framework to include
   * @returns Generated package.json configuration
   * @tags @CODE:GENERATE-PACKAGE-JSON-001:API
   */
  public generatePackageJson(
    projectConfig: Partial<PackageJsonConfig>,
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean = false,
    testingFramework?: string
  ): PackageJsonConfig {
    const baseConfig: PackageJsonConfig = {
      name: projectConfig.name || 'unnamed-project',
      version: projectConfig.version || '1.0.0',
      description: projectConfig.description || '',
      main: projectConfig.main || 'index.js',
      type: projectConfig.type || 'commonjs',
      scripts: this.generateScripts(
        packageManagerType,
        includeTypeScript,
        testingFramework
      ),
      keywords: projectConfig.keywords || [],
      author: projectConfig.author || '',
      license: projectConfig.license || 'MIT',
      engines: {
        node: '>=18.0.0',
        ...this.getPackageManagerEngine(packageManagerType),
      },
      files: projectConfig.files || ['dist', 'lib'],
      dependencies: {},
      devDependencies: {},
    };

    // Add TypeScript dependencies if requested
    if (includeTypeScript) {
      baseConfig.devDependencies = {
        ...baseConfig.devDependencies,
        typescript: '^5.0.0',
        '@types/node': '^20.0.0',
      };
    }

    // Add testing framework dependencies
    if (testingFramework === 'jest') {
      baseConfig.devDependencies = {
        ...baseConfig.devDependencies,
        jest: '^29.0.0',
      };

      if (includeTypeScript) {
        baseConfig.devDependencies = {
          ...baseConfig.devDependencies,
          '@types/jest': '^29.0.0',
          'ts-jest': '^29.0.0',
        };
      }
    }

    return baseConfig;
  }

  /**
   * Add dependencies to existing package.json
   * @param existingPackageJson - Existing package.json content
   * @param newDependencies - New dependencies to add
   * @returns Updated package.json
   * @tags @CODE:ADD-DEPENDENCIES-001:API
   */
  public addDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDependencies: Record<string, string>
  ): PackageJsonConfig {
    return {
      ...existingPackageJson,
      dependencies: {
        ...existingPackageJson.dependencies,
        ...newDependencies,
      },
    } as PackageJsonConfig;
  }

  /**
   * Add dev dependencies to existing package.json
   * @param existingPackageJson - Existing package.json content
   * @param newDevDependencies - New dev dependencies to add
   * @returns Updated package.json
   * @tags @CODE:ADD-DEV-DEPENDENCIES-001:API
   */
  public addDevDependencies(
    existingPackageJson: Partial<PackageJsonConfig>,
    newDevDependencies: Record<string, string>
  ): PackageJsonConfig {
    return {
      ...existingPackageJson,
      devDependencies: {
        ...existingPackageJson.devDependencies,
        ...newDevDependencies,
      },
    } as PackageJsonConfig;
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
      const command = this.getInitCommand(packageManagerType);
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

  /**
   * Build install command for package manager
   * @param packages - Packages to install
   * @param options - Installation options
   * @returns Command string
   * @tags @UTIL:BUILD-COMMAND-001
   */
  private buildInstallCommand(
    packages: string[],
    options: PackageInstallOptions
  ): string {
    const { packageManager, isDevelopment, isGlobal } = options;

    switch (packageManager) {
      case PackageManagerType.NPM: {
        let npmCmd = 'npm install';
        if (isDevelopment) npmCmd += ' --save-dev';
        if (isGlobal) npmCmd += ' --global';
        return `${npmCmd} ${packages.join(' ')}`;
      }

      case PackageManagerType.YARN: {
        let yarnCmd = 'yarn add';
        if (isDevelopment) yarnCmd += ' --dev';
        if (isGlobal) yarnCmd = 'yarn global add';
        return `${yarnCmd} ${packages.join(' ')}`;
      }

      case PackageManagerType.PNPM: {
        let pnpmCmd = 'pnpm add';
        if (isDevelopment) pnpmCmd += ' --save-dev';
        if (isGlobal) pnpmCmd += ' --global';
        return `${pnpmCmd} ${packages.join(' ')}`;
      }

      default:
        throw new Error(`Unsupported package manager: ${packageManager}`);
    }
  }

  /**
   * Generate scripts section for package.json
   * @param packageManagerType - Package manager type
   * @param includeTypeScript - Include TypeScript scripts
   * @param testingFramework - Testing framework
   * @returns Scripts object
   * @tags @UTIL:GENERATE-SCRIPTS-001
   */
  private generateScripts(
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean,
    testingFramework?: string
  ): Record<string, string> {
    const testCommand = this.getTestCommand(
      packageManagerType,
      testingFramework
    );

    const baseScripts: Record<string, string> = {
      start: 'node index.js',
      build: includeTypeScript ? 'tsc' : 'echo "No build step configured"',
      test: testCommand,
    };

    if (includeTypeScript) {
      baseScripts['type-check'] = 'tsc --noEmit';
      baseScripts['dev'] = 'ts-node src/index.ts';
    }

    if (testingFramework === 'jest') {
      baseScripts['test:watch'] = `${testCommand} --watch`;
      baseScripts['test:coverage'] = `${testCommand} --coverage`;
    }

    return baseScripts;
  }

  /**
   * Get run command for package manager
   * @param packageManagerType - Package manager type
   * @returns Run command
   * @tags @UTIL:RUN-COMMAND-001
   */
  public getRunCommand(packageManagerType: PackageManagerType): string {
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
   * Get test command for package manager and framework
   * @param packageManagerType - Package manager type
   * @param testingFramework - Testing framework
   * @returns Test command
   * @tags @UTIL:TEST-COMMAND-001
   */
  private getTestCommand(
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
   * Get init command for package manager
   * @param packageManagerType - Package manager type
   * @returns Init command
   * @tags @UTIL:INIT-COMMAND-001
   */
  private getInitCommand(packageManagerType: PackageManagerType): string {
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
   * @tags @UTIL:ENGINE-REQUIREMENT-001
   */
  private getPackageManagerEngine(
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
}
