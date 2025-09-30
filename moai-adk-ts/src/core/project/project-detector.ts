// @CODE:PROJ-003 | 
// Related: @CODE:PROJ-003:API, @CODE:PROJ-INFO-001

/**
 * @file Project type and structure detection
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../../utils/winston-logger.js';
import type {
  BuildToolIndicators,
  FileInfo,
  FrameworkIndicators,
  LanguageExtensions,
  PackageAnalysis,
  ProjectConfig,
  ProjectFileIndicators,
  ProjectInfo,
} from './types';

/**
 * ProjectDetector class for analyzing project type, language, and frameworks
 * @tags @CODE:PROJECT-DETECTOR-001
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
  };

  /**
   * Detect project type based on existing files
   * @param projectPath Path to project directory
   * @returns Project information
   * @tags @CODE:DETECT-PROJECT-001:API
   */
  public async detectProjectType(projectPath: string): Promise<ProjectInfo> {
    const detected: ProjectInfo = {
      type: 'unknown',
      language: 'unknown',
      frameworks: [],
      buildTools: [],
      filesFound: [],
      hasScripts: false,
      scripts: [],
    };

    logger.info(
      `Detecting project type in: ${projectPath} (excluding MoAI framework files)`
    );

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

    // Analyze package.json if present (but exclude MoAI-specific package.json)
    if (detected.filesFound.includes('package.json')) {
      const packageJsonPath = path.join(projectPath, 'package.json');

      // Check if this is a MoAI-ADK package.json (exclude it from user project analysis)
      try {
        const packageData = JSON.parse(
          fs.readFileSync(packageJsonPath, 'utf-8')
        );
        if (
          packageData.name === 'moai-adk' ||
          packageData.description?.includes('MoAI-ADK')
        ) {
          logger.info('Skipping MoAI-ADK package.json from project analysis');
          detected.filesFound = detected.filesFound.filter(
            f => f !== 'package.json'
          );
          detected.type = 'unknown';
          detected.language = 'unknown';
        } else {
          const packageAnalysis =
            await this.analyzePackageJson(packageJsonPath);
          return {
            ...detected,
            frameworks: packageAnalysis.frameworks,
            buildTools: packageAnalysis.buildTools,
            hasScripts: packageAnalysis.hasScripts,
            scripts: packageAnalysis.scripts,
          };
        }
      } catch (error) {
        logger.warn('Could not read package.json for MoAI check:', error);
      }
    }

    logger.info(`Project detection completed:`, detected);
    return detected;
  }

  /**
   * Analyze package.json for frameworks and dependencies
   * @param packageJsonPath Path to package.json file
   * @returns Package analysis result
   * @tags @CODE:ANALYZE-PACKAGE-001:API
   */
  public async analyzePackageJson(
    packageJsonPath: string
  ): Promise<PackageAnalysis> {
    try {
      const packageData = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

      const frameworks: string[] = [];
      const buildTools: string[] = [];

      // Combine dependencies and devDependencies
      const allDeps = {
        ...(packageData.dependencies || {}),
        ...(packageData.devDependencies || {}),
      };

      // Detect frameworks
      for (const [framework, indicators] of Object.entries(
        this.frameworkIndicators
      )) {
        if (indicators.some(indicator => indicator in allDeps)) {
          frameworks.push(framework);
          logger.info(`Detected framework: ${framework}`);
        }
      }

      // Detect build tools
      for (const [tool, indicators] of Object.entries(
        this.buildToolIndicators
      )) {
        if (indicators.some(indicator => indicator in allDeps)) {
          buildTools.push(tool);
          logger.info(`Detected build tool: ${tool}`);
        }
      }

      const hasScripts = Boolean(packageData.scripts);
      const scripts = Object.keys(packageData.scripts || {});

      logger.info(
        `Package.json analysis: frameworks=${frameworks}, buildTools=${buildTools}`
      );

      return {
        frameworks,
        buildTools,
        hasScripts,
        scripts,
      };
    } catch (error) {
      logger.error('Error analyzing package.json:', error);
      return {
        frameworks: [],
        buildTools: [],
        hasScripts: false,
        scripts: [],
      };
    }
  }

  /**
   * Check if package.json should be created based on project configuration
   * @param config Project configuration
   * @returns True if package.json should be created
   * @tags @CODE:SHOULD-CREATE-PACKAGE-001:API
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
   * @param projectPath Path to project directory
   * @returns Detected primary language
   * @tags @CODE:DETECT-LANGUAGE-001:API
   */
  public async detectLanguageFromFiles(projectPath: string): Promise<string> {
    if (!fs.existsSync(projectPath)) {
      logger.warn(`Project path does not exist: ${projectPath}`);
      return 'unknown';
    }

    const fileCounts: { [key: string]: number } = {};
    for (const lang of Object.keys(this.languageExtensions)) {
      fileCounts[lang] = 0;
    }

    try {
      const files = await this.scanDirectory(projectPath);

      for (const file of files) {
        if (file.isFile()) {
          const ext = path.extname(file.path).toLowerCase();
          for (const [lang, extensions] of Object.entries(
            this.languageExtensions
          )) {
            if (extensions.includes(ext)) {
              fileCounts[lang] = (fileCounts[lang] || 0) + 1;
            }
          }
        }
      }

      // Return language with most files
      const detectedLanguage = Object.keys(fileCounts).reduce((a, b) =>
        (fileCounts[a] || 0) > (fileCounts[b] || 0) ? a : b
      );

      if ((fileCounts[detectedLanguage] || 0) > 0) {
        logger.info(
          `Detected language: ${detectedLanguage} (${fileCounts[detectedLanguage]} files)`
        );
        return detectedLanguage;
      } else {
        logger.info('No specific language detected from file extensions');
        return 'unknown';
      }
    } catch (error) {
      logger.error(`Error scanning files: ${error}`);
      return 'unknown';
    }
  }

  /**
   * Scan directory recursively for files
   * @param dirPath Directory path to scan
   * @returns Array of file information
   * @tags @UTIL:SCAN-DIRECTORY-001
   */
  private async scanDirectory(dirPath: string): Promise<FileInfo[]> {
    const files: FileInfo[] = [];

    // MoAI framework files and directories to exclude from analysis
    const moaiExclusions = [
      '.claude',
      '.moai',
      'CLAUDE.md',
      'node_modules',
      '.git',
      'dist',
      'build',
      'coverage',
    ];

    const scanRecursive = (currentPath: string) => {
      try {
        const entries = fs.readdirSync(currentPath, { withFileTypes: true });

        for (const entry of entries) {
          const fullPath = path.join(currentPath, entry.name);

          // Skip MoAI framework files and directories
          if (moaiExclusions.includes(entry.name)) {
            logger.info(
              `Skipping MoAI framework file/directory: ${entry.name}`
            );
            continue;
          }

          if (entry.isDirectory() && !entry.name.startsWith('.')) {
            // Skip hidden directories but recurse into others
            scanRecursive(fullPath);
          } else if (entry.isFile()) {
            files.push({
              path: fullPath,
              isFile: () => true,
            });
          }
        }
      } catch (error) {
        logger.error(`Error scanning directory ${currentPath}:`, error);
      }
    };

    scanRecursive(dirPath);
    return files;
  }
}
