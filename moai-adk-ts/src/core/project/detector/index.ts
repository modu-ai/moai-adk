// @CODE:PROJ-003 | SPEC: project-detector
// Related: @CODE:PROJ-003:API, @CODE:PROJ-INFO-001

/**
 * @file Project type and structure detection orchestrator
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger.js';
import type {
  BuildToolIndicators,
  FrameworkIndicators,
  LanguageExtensions,
  PackageAnalysis,
  ProjectConfig,
  ProjectFileIndicators,
  ProjectInfo,
} from '../types';
import { FileAnalyzer } from './analyzers/file-analyzer';
import { PackageAnalyzer } from './analyzers/package-analyzer';
import { LanguageDetector } from './detectors/language-detector';
import { NodeJsDetector } from './detectors/nodejs-detector';
import { PythonDetector } from './detectors/python-detector';

/**
 * ProjectDetector class for analyzing project type, language, and frameworks
 */
export class ProjectDetector {
  private readonly projectFileIndicators: ProjectFileIndicators = {
    'package.json': { type: 'nodejs', language: 'javascript' },
    'requirements.txt': { type: 'python', language: 'python' },
    'pyproject.toml': { type: 'python', language: 'python' },
    'Cargo.toml': { type: 'rust', language: 'rust' },
    'go.mod': { type: 'go', language: 'go' },
    'pom.xml': { type: 'java', language: 'java' },
    'build.gradle': { type: 'java', language: 'java' },
    Gemfile: { type: 'ruby', language: 'ruby' },
    'composer.json': { type: 'php', language: 'php' },
    'pubspec.yaml': { type: 'flutter', language: 'dart' },
    Podfile: { type: 'ios', language: 'swift' },
    'Package.swift': { type: 'ios', language: 'swift' },
    'build.gradle.kts': { type: 'android', language: 'kotlin' },
  };

  private readonly frameworkIndicators: FrameworkIndicators = {
    react: ['react', '@types/react'],
    vue: ['vue', '@vue/cli'],
    angular: ['@angular/core', '@angular/cli'],
    svelte: ['svelte'],
    nextjs: ['next'],
    nuxtjs: ['nuxt'],
    express: ['express'],
    fastify: ['fastify'],
    'react-native': ['react-native', '@react-native'],
    expo: ['expo', 'expo-cli'],
    flutter: ['flutter'],
  };

  private readonly buildToolIndicators: BuildToolIndicators = {
    webpack: ['webpack'],
    vite: ['vite'],
    rollup: ['rollup'],
    parcel: ['parcel'],
    typescript: ['typescript', '@types/node'],
  };

  private readonly languageExtensions: LanguageExtensions = {
    python: ['.py', '.pyx', '.pyi'],
    javascript: ['.js', '.jsx', '.mjs'],
    typescript: ['.ts', '.tsx'],
    rust: ['.rs'],
    go: ['.go'],
    java: ['.java'],
    ruby: ['.rb'],
    php: ['.php'],
    cpp: ['.cpp', '.cxx', '.cc'],
    c: ['.c'],
    dart: ['.dart'],
    swift: ['.swift'],
    kotlin: ['.kt', '.kts'],
  };

  // Analyzers and detectors
  private readonly fileAnalyzer: FileAnalyzer;
  private readonly packageAnalyzer: PackageAnalyzer;
  private readonly languageDetector: LanguageDetector;
  private readonly nodeJsDetector: NodeJsDetector;
  private readonly pythonDetector: PythonDetector;

  constructor() {
    // Initialize analyzers
    this.fileAnalyzer = new FileAnalyzer();
    this.packageAnalyzer = new PackageAnalyzer(
      this.frameworkIndicators,
      this.buildToolIndicators
    );

    // Initialize detectors
    this.languageDetector = new LanguageDetector(
      this.fileAnalyzer,
      this.languageExtensions
    );
    this.nodeJsDetector = new NodeJsDetector(this.packageAnalyzer);
    this.pythonDetector = new PythonDetector();
  }

  /**
   * Detect project type based on existing files
   */
  public async detectProjectType(projectPath: string): Promise<ProjectInfo> {
    logger.info(
      `Detecting project type in: ${projectPath} (excluding MoAI framework files)`
    );

    // Try Node.js detector first
    const nodeResult = await this.nodeJsDetector.detect(projectPath);
    if (nodeResult) {
      logger.info(`Project detection completed`, nodeResult);
      return nodeResult;
    }

    // Try Python detector
    const pythonResult = await this.pythonDetector.detect(projectPath);
    if (pythonResult) {
      logger.info(`Project detection completed`, pythonResult);
      return pythonResult;
    }

    // Fallback: generic file-based detection
    const detected = await this.detectGenericProject(projectPath);
    logger.info(`Project detection completed`, detected);
    return detected;
  }

  /**
   * Generic project detection fallback
   */
  private async detectGenericProject(
    projectPath: string
  ): Promise<ProjectInfo> {
    const detected: ProjectInfo = {
      type: 'unknown',
      language: 'unknown',
      frameworks: [],
      buildTools: [],
      hasTests: false,
      hasDocker: false,
      hasCI: false,
      filesFound: [],
      hasScripts: false,
      packageManager: undefined,
      scripts: [],
    };

    // Check for various project files
    for (const [fileName, info] of Object.entries(this.projectFileIndicators)) {
      const filePath = path.join(projectPath, fileName);
      if (fs.existsSync(filePath)) {
        detected.filesFound.push(fileName);
        detected.type = info.type;
        detected.language = info.language;
        logger.info(`Found ${fileName}, detected as ${info.language} project`);
      }
    }

    return detected;
  }

  /**
   * Analyze package.json for frameworks and dependencies
   */
  public async analyzePackageJson(
    packageJsonPath: string
  ): Promise<PackageAnalysis> {
    return this.packageAnalyzer.analyzePackageJson(packageJsonPath);
  }

  /**
   * Check if package.json should be created based on project configuration
   */
  public shouldCreatePackageJson(config: ProjectConfig): boolean {
    const shouldCreate =
      config.runtime.name === 'node' ||
      config.runtime.name === 'tsx' ||
      config.techStack.some(tech =>
        ['nextjs', 'react', 'vue', 'angular', 'svelte'].includes(tech)
      );

    logger.info(
      `Should create package.json: ${shouldCreate} (runtime: ${config.runtime.name})`
    );
    return shouldCreate;
  }

  /**
   * Detect primary language based on file extensions in project
   */
  public async detectLanguageFromFiles(projectPath: string): Promise<string> {
    return this.languageDetector.detectLanguageFromFiles(projectPath);
  }
}
