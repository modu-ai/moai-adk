// @CODE:REFACTOR-007 | Chain: @SPEC:REFACTOR-007 -> @SPEC:REFACTOR-007 -> @CODE:REFACTOR-007 -> @TEST:REFACTOR-007
// Related: @CODE:PKG-002

/**
 * @file Package.json configuration builder
 * @author MoAI Team
 * @tags @CODE:PACKAGE-JSON-BUILDER-001
 */

import type {
  PackageJsonConfig,
  PackageManagerType,
} from '@/types/package-manager';
import type { CommandBuilder } from './command-builder';

/**
 * Builds and manages package.json configurations
 * @tags @CODE:PACKAGE-JSON-BUILDER-001:FEATURE
 */
export class PackageJsonBuilder {
  private commandBuilder: CommandBuilder;

  constructor(commandBuilder: CommandBuilder) {
    this.commandBuilder = commandBuilder;
  }

  /**
   * Generate package.json configuration
   * @param projectConfig - Partial project configuration
   * @param packageManagerType - Package manager type
   * @param includeTypeScript - Whether to include TypeScript setup
   * @param testingFramework - Testing framework to include
   * @returns Generated package.json configuration
   * @tags @CODE:GENERATE-PKG-JSON-001:API
   */
  public generatePackageJson(
    projectConfig: Partial<PackageJsonConfig>,
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean = false,
    testingFramework?: string
  ): PackageJsonConfig {
    const baseConfig = this.createBaseConfig(projectConfig);
    const scripts = this.generateScripts(
      packageManagerType,
      includeTypeScript,
      testingFramework
    );
    const engines = this.buildEngines(packageManagerType);

    let config: PackageJsonConfig = {
      ...baseConfig,
      scripts,
      engines,
      dependencies: {},
      devDependencies: {},
    };

    if (includeTypeScript) {
      config = this.addTypeScriptDependencies(config);
    }

    if (testingFramework) {
      config = this.addTestingDependencies(
        config,
        testingFramework,
        includeTypeScript
      );
    }

    return config;
  }

  /**
   * Generate scripts section for package.json
   * @param packageManagerType - Package manager type
   * @param includeTypeScript - Include TypeScript scripts
   * @param testingFramework - Testing framework
   * @returns Scripts object
   * @tags @CODE:GENERATE-SCRIPTS-001:API
   */
  public generateScripts(
    packageManagerType: PackageManagerType,
    includeTypeScript: boolean,
    testingFramework?: string
  ): Record<string, string> {
    const testCommand = this.commandBuilder.buildTestCommand(
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
   * Add dependencies to existing package.json
   * @param existingPackageJson - Existing package.json content
   * @param newDependencies - New dependencies to add
   * @returns Updated package.json
   * @tags @CODE:ADD-DEPS-001:API
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
   * @tags @CODE:ADD-DEV-DEPS-001:API
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
   * Create base package.json configuration
   * @private
   */
  private createBaseConfig(
    config: Partial<PackageJsonConfig>
  ): Omit<
    PackageJsonConfig,
    'scripts' | 'engines' | 'dependencies' | 'devDependencies'
  > {
    return {
      name: config.name || 'unnamed-project',
      version: config.version || '1.0.0',
      description: config.description || '',
      main: config.main || 'index.js',
      type: config.type || 'commonjs',
      keywords: config.keywords || [],
      author: config.author || '',
      license: config.license || 'MIT',
      files: config.files || ['dist', 'lib'],
    };
  }

  /**
   * Build engines object with Node.js and package manager requirements
   * @private
   */
  private buildEngines(
    packageManagerType: PackageManagerType
  ): Record<string, string> {
    return {
      node: '>=18.0.0',
      ...this.commandBuilder.getPackageManagerEngine(packageManagerType),
    };
  }

  /**
   * Add TypeScript dependencies to configuration
   * @private
   */
  private addTypeScriptDependencies(
    config: PackageJsonConfig
  ): PackageJsonConfig {
    return {
      ...config,
      devDependencies: {
        ...config.devDependencies,
        typescript: '^5.0.0',
        '@types/node': '^20.0.0',
      },
    };
  }

  /**
   * Add testing framework dependencies to configuration
   * @private
   */
  private addTestingDependencies(
    config: PackageJsonConfig,
    framework: string,
    includeTypeScript: boolean
  ): PackageJsonConfig {
    if (framework !== 'jest') {
      return config;
    }

    const jestDependencies: Record<string, string> = {
      jest: '^29.0.0',
    };

    if (includeTypeScript) {
      jestDependencies['@types/jest'] = '^29.0.0';
      jestDependencies['ts-jest'] = '^29.0.0';
    }

    return {
      ...config,
      devDependencies: {
        ...config.devDependencies,
        ...jestDependencies,
      },
    };
  }
}
