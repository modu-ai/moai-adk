/**
 * @file Development environment analyzer
 * @author MoAI Team
 * @tags @FEATURE:ENVIRONMENT-ANALYZER-001 @REQ:ADVANCED-DOCTOR-001
 */

import * as fs from 'fs/promises';
import * as path from 'path';
import { execSync } from 'child_process';
import type {
  EnvironmentConfig,
  OptimizationRecommendation,
  DiagnosticSeverity,
} from '@/types/diagnostics';

/**
 * Development environment analyzer for tool detection and configuration
 * @tags @FEATURE:ENVIRONMENT-ANALYZER-001
 */
export class EnvironmentAnalyzer {
  private readonly supportedEnvironments = [
    'Node.js',
    'TypeScript',
    'Python',
    'Git',
    'Docker',
    'Java',
    'Go',
    'Rust',
  ];

  /**
   * Analyze all detected development environments
   * @param projectPath - Path to analyze (defaults to current directory)
   * @returns Array of environment analysis results
   * @tags @API:ANALYZE-ENVIRONMENTS-001
   */
  public async analyzeEnvironments(
    projectPath: string = process.cwd()
  ): Promise<EnvironmentConfig[]> {
    const environments: EnvironmentConfig[] = [];

    for (const envName of this.supportedEnvironments) {
      try {
        const analysis = await this.analyzeEnvironment(envName, projectPath);
        if (analysis) {
          environments.push(analysis);
        }
      } catch (error) {
        // Environment not detected or analysis failed
        environments.push({
          name: envName,
          detected: false,
          configFiles: [],
          recommendations: [],
          status: 'problematic',
        });
      }
    }

    return environments.filter(env => env.detected);
  }

  /**
   * Analyze specific development environment
   * @param environmentName - Name of environment to analyze
   * @param projectPath - Project path
   * @returns Environment analysis result
   * @tags @API:ANALYZE-ENVIRONMENT-001
   */
  private async analyzeEnvironment(
    environmentName: string,
    projectPath: string
  ): Promise<EnvironmentConfig | null> {
    switch (environmentName) {
      case 'Node.js':
        return this.analyzeNodeJs(projectPath);
      case 'TypeScript':
        return this.analyzeTypeScript(projectPath);
      case 'Python':
        return this.analyzePython(projectPath);
      case 'Git':
        return this.analyzeGit(projectPath);
      case 'Docker':
        return this.analyzeDocker(projectPath);
      case 'Java':
        return this.analyzeJava(projectPath);
      case 'Go':
        return this.analyzeGo(projectPath);
      case 'Rust':
        return this.analyzeRust(projectPath);
      default:
        return null;
    }
  }

  /**
   * Analyze Node.js environment
   * @param projectPath - Project path
   * @returns Node.js environment analysis
   * @tags @UTIL:ANALYZE-NODEJS-001
   */
  private async analyzeNodeJs(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('node --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'package.json',
      '.nvmrc',
      '.node-version',
    ]);

    const recommendations: OptimizationRecommendation[] = [];

    // Check Node.js version
    if (version) {
      const majorVersion = parseInt(version.replace('v', '').split('.')[0] || '0');
      if (majorVersion < 18) {
        recommendations.push({
          category: 'compatibility',
          severity: DiagnosticSeverity.WARNING,
          title: 'Outdated Node.js Version',
          description: `Node.js ${version} is outdated. Consider upgrading to v18+ for better performance and security`,
          impact: 'medium',
          effort: 'medium',
          steps: [
            'Update to Node.js LTS version (v18 or v20)',
            'Test compatibility with existing code',
            'Update package.json engines field',
          ],
        });
      }
    }

    return {
      name: 'Node.js',
      detected: !!version,
      version,
      configFiles,
      recommendations,
      status: this.determineStatus(recommendations),
    };
  }

  /**
   * Analyze TypeScript environment
   * @param projectPath - Project path
   * @returns TypeScript environment analysis
   * @tags @UTIL:ANALYZE-TYPESCRIPT-001
   */
  private async analyzeTypeScript(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('tsc --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'tsconfig.json',
      'tsconfig.build.json',
      '.eslintrc.json',
      '.eslintrc.js',
    ]);

    const recommendations: OptimizationRecommendation[] = [];

    // Check if tsconfig.json exists
    if (!configFiles.includes('tsconfig.json')) {
      recommendations.push({
        category: 'compatibility',
        severity: DiagnosticSeverity.WARNING,
        title: 'Missing TypeScript Configuration',
        description: 'No tsconfig.json found. This is recommended for TypeScript projects',
        impact: 'medium',
        effort: 'easy',
        steps: [
          'Create tsconfig.json file',
          'Configure strict mode for better type safety',
          'Set appropriate target and module options',
        ],
      });
    }

    return {
      name: 'TypeScript',
      detected: !!version,
      version: version?.replace('Version ', ''),
      configFiles,
      recommendations,
      status: this.determineStatus(recommendations),
    };
  }

  /**
   * Analyze Python environment
   * @param projectPath - Project path
   * @returns Python environment analysis
   * @tags @UTIL:ANALYZE-PYTHON-001
   */
  private async analyzePython(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('python --version') ||
                   this.getToolVersion('python3 --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'requirements.txt',
      'pyproject.toml',
      'setup.py',
      'Pipfile',
      '.python-version',
    ]);

    const recommendations: OptimizationRecommendation[] = [];

    // Check Python version
    if (version) {
      const versionMatch = version.match(/(\d+)\.(\d+)/);
      if (versionMatch) {
        const major = parseInt(versionMatch[1] || '0');
        const minor = parseInt(versionMatch[2] || '0');

        if (major < 3 || (major === 3 && minor < 8)) {
          recommendations.push({
            category: 'security',
            severity: DiagnosticSeverity.ERROR,
            title: 'Unsupported Python Version',
            description: `Python ${version} is outdated and may have security vulnerabilities`,
            impact: 'high',
            effort: 'medium',
            steps: [
              'Upgrade to Python 3.8 or newer',
              'Update dependencies for compatibility',
              'Test application thoroughly after upgrade',
            ],
          });
        }
      }
    }

    return {
      name: 'Python',
      detected: !!version,
      version,
      configFiles,
      recommendations,
      status: this.determineStatus(recommendations),
    };
  }

  /**
   * Analyze Git environment
   * @param projectPath - Project path
   * @returns Git environment analysis
   * @tags @UTIL:ANALYZE-GIT-001
   */
  private async analyzeGit(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('git --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      '.gitconfig',
      '.gitignore',
      '.gitattributes',
    ]);

    const recommendations: OptimizationRecommendation[] = [];

    // Check if it's a git repository
    try {
      execSync('git rev-parse --git-dir', { cwd: projectPath, stdio: 'ignore' });
    } catch {
      recommendations.push({
        category: 'maintenance',
        severity: DiagnosticSeverity.INFO,
        title: 'Not a Git Repository',
        description: 'Consider initializing Git for version control',
        impact: 'low',
        effort: 'easy',
        steps: [
          'Run "git init" to initialize repository',
          'Create .gitignore file',
          'Make initial commit',
        ],
      });
    }

    return {
      name: 'Git',
      detected: !!version,
      version: version?.split(' ')[2],
      configFiles,
      recommendations,
      status: this.determineStatus(recommendations),
    };
  }

  /**
   * Analyze Docker environment
   * @param projectPath - Project path
   * @returns Docker environment analysis
   * @tags @UTIL:ANALYZE-DOCKER-001
   */
  private async analyzeDocker(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('docker --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'Dockerfile',
      'docker-compose.yml',
      'docker-compose.yaml',
      '.dockerignore',
    ]);

    return {
      name: 'Docker',
      detected: !!version,
      version: version?.split(' ')[2]?.replace(',', ''),
      configFiles,
      recommendations: [],
      status: 'optimal',
    };
  }

  /**
   * Analyze Java environment
   * @param projectPath - Project path
   * @returns Java environment analysis
   * @tags @UTIL:ANALYZE-JAVA-001
   */
  private async analyzeJava(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('java --version') ||
                   this.getToolVersion('java -version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'pom.xml',
      'build.gradle',
      'build.gradle.kts',
    ]);

    return {
      name: 'Java',
      detected: !!version,
      version,
      configFiles,
      recommendations: [],
      status: version ? 'good' : 'problematic',
    };
  }

  /**
   * Analyze Go environment
   * @param projectPath - Project path
   * @returns Go environment analysis
   * @tags @UTIL:ANALYZE-GO-001
   */
  private async analyzeGo(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('go version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'go.mod',
      'go.sum',
      'go.work',
    ]);

    return {
      name: 'Go',
      detected: !!version,
      version: version?.split(' ')[2],
      configFiles,
      recommendations: [],
      status: version ? 'good' : 'problematic',
    };
  }

  /**
   * Analyze Rust environment
   * @param projectPath - Project path
   * @returns Rust environment analysis
   * @tags @UTIL:ANALYZE-RUST-001
   */
  private async analyzeRust(projectPath: string): Promise<EnvironmentConfig> {
    const version = this.getToolVersion('rustc --version');
    const configFiles = await this.findConfigFiles(projectPath, [
      'Cargo.toml',
      'Cargo.lock',
      'rust-toolchain.toml',
    ]);

    return {
      name: 'Rust',
      detected: !!version,
      version: version?.split(' ')[1],
      configFiles,
      recommendations: [],
      status: version ? 'good' : 'problematic',
    };
  }

  /**
   * Get tool version by running version command
   * @param command - Version command to execute
   * @returns Version string or null
   * @tags @UTIL:GET-TOOL-VERSION-001
   */
  private getToolVersion(command: string): string | null {
    try {
      const output = execSync(command, {
        encoding: 'utf8',
        stdio: 'pipe',
        timeout: 5000
      });
      return output.trim();
    } catch {
      return null;
    }
  }

  /**
   * Find configuration files in project directory
   * @param projectPath - Project path
   * @param fileNames - List of file names to search for
   * @returns Array of found configuration files
   * @tags @UTIL:FIND-CONFIG-FILES-001
   */
  private async findConfigFiles(
    projectPath: string,
    fileNames: string[]
  ): Promise<string[]> {
    const foundFiles: string[] = [];

    for (const fileName of fileNames) {
      try {
        const filePath = path.join(projectPath, fileName);
        await fs.access(filePath);
        foundFiles.push(fileName);
      } catch {
        // File not found
      }
    }

    return foundFiles;
  }

  /**
   * Determine environment status based on recommendations
   * @param recommendations - List of recommendations
   * @returns Environment status
   * @tags @UTIL:DETERMINE-STATUS-001
   */
  private determineStatus(
    recommendations: OptimizationRecommendation[]
  ): 'optimal' | 'good' | 'needs_improvement' | 'problematic' {
    if (recommendations.length === 0) {
      return 'optimal';
    }

    const hasCritical = recommendations.some(r => r.severity === DiagnosticSeverity.CRITICAL);
    const hasError = recommendations.some(r => r.severity === DiagnosticSeverity.ERROR);
    const hasWarning = recommendations.some(r => r.severity === DiagnosticSeverity.WARNING);

    if (hasCritical) return 'problematic';
    if (hasError) return 'needs_improvement';
    if (hasWarning) return 'good';

    return 'optimal';
  }
}