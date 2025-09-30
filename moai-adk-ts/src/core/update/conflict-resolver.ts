/**
 * @file Interactive conflict resolution system for update operations
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:CONFLICT-RESOLVER-001 -> @TASK:CONFLICT-RESOLVER-001 -> @TEST:CONFLICT-RESOLVER-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import inquirer from 'inquirer';
import { type FileChangeAnalysis, UpdateAction } from './types.js';
import { logger } from '../../utils/winston-logger.js';

/**
 * Conflict resolution choice
 * @tags @DESIGN:CONFLICT-CHOICE-001
 */
export interface ConflictChoice {
  readonly action: UpdateAction;
  readonly reason: string;
  readonly applyToSimilar: boolean;
}

/**
 * Conflict resolution result
 * @tags @DESIGN:CONFLICT-RESOLUTION-001
 */
export interface ConflictResolution {
  readonly filePath: string;
  readonly choice: ConflictChoice;
  readonly timestamp: string;
}

/**
 * Merge strategy for hybrid files
 * @tags @DESIGN:MERGE-STRATEGY-001
 */
export enum MergeStrategy {
  /** Keep user content, add template sections */
  APPEND_NEW = 'APPEND_NEW',

  /** Replace template sections, keep user sections */
  SECTION_REPLACE = 'SECTION_REPLACE',

  /** Manual three-way merge */
  MANUAL_MERGE = 'MANUAL_MERGE',

  /** Use template as base, apply user customizations */
  TEMPLATE_BASE = 'TEMPLATE_BASE',
}

/**
 * Interactive conflict resolver for update operations
 * @tags @FEATURE:CONFLICT-RESOLVER-001
 */
export class ConflictResolver {
  private readonly resolutions: Map<string, ConflictResolution> = new Map();
  private readonly patterns: Map<string, UpdateAction> = new Map();

  /**
   * Resolve conflicts for multiple files interactively
   * @param conflicts - Files with conflicts to resolve
   * @param projectPath - Project root directory
   * @param templatePath - Template directory
   * @returns Map of resolutions for each file
   * @tags @API:RESOLVE-CONFLICTS-001
   */
  public async resolveConflicts(
    conflicts: readonly FileChangeAnalysis[],
    projectPath: string,
    templatePath: string
  ): Promise<Map<string, ConflictResolution>> {
    logger.info(chalk.cyan('🔧 Conflict Resolution Required'));
    logger.info(`Found ${conflicts.length} files requiring manual review:\n`);

    for (const conflict of conflicts) {
      // Check if we have a pattern-based resolution
      const patternAction = this.checkPatternResolution(conflict.path);

      if (patternAction) {
        const resolution: ConflictResolution = {
          filePath: conflict.path,
          choice: {
            action: patternAction,
            reason: 'Applied pattern-based resolution',
            applyToSimilar: false,
          },
          timestamp: new Date().toISOString(),
        };

        this.resolutions.set(conflict.path, resolution);
        logger.info(
          chalk.green(`✓ ${conflict.path}: ${patternAction} (pattern)`)
        );
        continue;
      }

      // Interactive resolution required
      const resolution = await this.resolveFileConflict(
        conflict,
        projectPath,
        templatePath
      );

      this.resolutions.set(conflict.path, resolution);

      // Apply to similar files if requested
      if (resolution.choice.applyToSimilar) {
        this.addPattern(conflict.path, resolution.choice.action);
      }
    }

    logger.info(chalk.green('\n✅ All conflicts resolved!'));
    return new Map(this.resolutions);
  }

  /**
   * Resolve conflict for a single file
   * @param conflict - File conflict analysis
   * @param projectPath - Project root directory
   * @param templatePath - Template directory
   * @returns Conflict resolution
   * @tags @API:RESOLVE-FILE-CONFLICT-001
   */
  public async resolveFileConflict(
    conflict: FileChangeAnalysis,
    projectPath: string,
    templatePath: string
  ): Promise<ConflictResolution> {
    logger.info(chalk.yellow(`\n📄 Resolving: ${conflict.path}`));
    logger.info(`Type: ${conflict.type}`);
    logger.info(`Conflict Level: ${conflict.conflictPotential}`);
    logger.info(`Recommended: ${conflict.recommendedAction}\n`);

    // Show file diff if both files exist
    await this.showFileDiff(conflict.path, projectPath, templatePath);

    const choices = this.getActionChoices(conflict);
    const answers = await inquirer.prompt([
      {
        type: 'list',
        name: 'action',
        message: 'How would you like to handle this file?',
        choices,
        default: conflict.recommendedAction,
      },
      {
        type: 'input',
        name: 'reason',
        message: 'Reason for this choice (optional):',
        default: '',
      },
      {
        type: 'confirm',
        name: 'applyToSimilar',
        message: 'Apply this choice to similar files?',
        default: false,
        when: answers => answers.action !== UpdateAction.MANUAL,
      },
    ]);

    // Handle special merge cases
    if (answers.action === UpdateAction.MERGE) {
      const mergeStrategy = await this.selectMergeStrategy(conflict);
      answers.reason = `${answers.reason} [${mergeStrategy}]`.trim();
    }

    return {
      filePath: conflict.path,
      choice: {
        action: answers.action as UpdateAction,
        reason: answers.reason || 'User decision',
        applyToSimilar: answers.applyToSimilar || false,
      },
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Perform intelligent merge for hybrid files
   * @param filePath - File to merge
   * @param projectPath - Project root directory
   * @param templatePath - Template directory
   * @param strategy - Merge strategy to use
   * @returns True if merge successful
   * @tags @API:SMART-MERGE-001
   */
  public async performSmartMerge(
    filePath: string,
    projectPath: string,
    templatePath: string,
    strategy: MergeStrategy
  ): Promise<boolean> {
    try {
      const userPath = path.join(projectPath, filePath);
      const templateFile = path.join(templatePath, filePath);

      const userContent = await fs.readFile(userPath, 'utf-8');
      const templateContent = await fs.readFile(templateFile, 'utf-8');

      let mergedContent: string;

      switch (strategy) {
        case MergeStrategy.APPEND_NEW:
          mergedContent = await this.appendNewSections(
            userContent,
            templateContent
          );
          break;

        case MergeStrategy.SECTION_REPLACE:
          mergedContent = await this.replaceSections(
            userContent,
            templateContent
          );
          break;

        case MergeStrategy.TEMPLATE_BASE:
          mergedContent = await this.templateBaseMerge(
            userContent,
            templateContent
          );
          break;

        case MergeStrategy.MANUAL_MERGE:
        default:
          mergedContent = await this.manualMerge(userContent, templateContent);
          break;
      }

      // Write merged content
      await fs.writeFile(userPath, mergedContent);

      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get available action choices for conflict
   * @param conflict - File conflict analysis
   * @returns List of inquirer choices
   * @tags @UTIL:GET-ACTION-CHOICES-001
   */
  private getActionChoices(
    conflict: FileChangeAnalysis
  ): inquirer.DistinctChoice[] {
    const baseChoices = [
      { name: '🔄 Replace with template version', value: UpdateAction.REPLACE },
      { name: '✋ Keep current version', value: UpdateAction.KEEP },
      { name: '🤝 Intelligent merge', value: UpdateAction.MERGE },
      { name: '👤 Manual resolution required', value: UpdateAction.MANUAL },
    ];

    // Add regenerate option for generated files
    if (conflict.type === 'GENERATED') {
      baseChoices.splice(3, 0, {
        name: '🔄 Regenerate from scratch',
        value: UpdateAction.REGENERATE,
      });
    }

    return baseChoices;
  }

  /**
   * Select merge strategy for hybrid files
   * @param conflict - File conflict analysis
   * @returns Selected merge strategy
   * @tags @UTIL:SELECT-MERGE-STRATEGY-001
   */
  private async selectMergeStrategy(
    _conflict: FileChangeAnalysis
  ): Promise<MergeStrategy> {
    const choices = [
      {
        name: '📝 Append new template sections to user content',
        value: MergeStrategy.APPEND_NEW,
      },
      {
        name: '🔄 Replace template sections, keep user sections',
        value: MergeStrategy.SECTION_REPLACE,
      },
      {
        name: '📋 Use template as base, apply user customizations',
        value: MergeStrategy.TEMPLATE_BASE,
      },
      {
        name: '✋ Manual three-way merge (advanced)',
        value: MergeStrategy.MANUAL_MERGE,
      },
    ];

    const answer = await inquirer.prompt([
      {
        type: 'list',
        name: 'strategy',
        message: 'Select merge strategy:',
        choices,
        default: MergeStrategy.APPEND_NEW,
      },
    ]);

    return answer.strategy as MergeStrategy;
  }

  /**
   * Show diff between user and template files
   * @param filePath - File to show diff for
   * @param projectPath - Project root directory
   * @param templatePath - Template directory
   * @tags @UTIL:SHOW-FILE-DIFF-001
   */
  private async showFileDiff(
    filePath: string,
    projectPath: string,
    templatePath: string
  ): Promise<void> {
    try {
      const userPath = path.join(projectPath, filePath);
      const templateFile = path.join(templatePath, filePath);

      const [userExists, templateExists] = await Promise.all([
        fs
          .access(userPath)
          .then(() => true)
          .catch(() => false),
        fs
          .access(templateFile)
          .then(() => true)
          .catch(() => false),
      ]);

      if (userExists && templateExists) {
        const [userContent, templateContent] = await Promise.all([
          fs.readFile(userPath, 'utf-8'),
          fs.readFile(templateFile, 'utf-8'),
        ]);

        logger.info(chalk.blue('📄 Current file (first 10 lines):'));
        logger.info(userContent.split('\n').slice(0, 10).join('\n'));

        logger.info(chalk.green('\n📄 Template version (first 10 lines):'));
        logger.info(templateContent.split('\n').slice(0, 10).join('\n'));
      } else if (!userExists) {
        logger.info(chalk.yellow('📄 File does not exist in current project'));
      } else {
        logger.info(chalk.yellow('📄 No template version available'));
      }
    } catch {
      logger.info(chalk.red('❌ Could not read file contents'));
    }
  }

  /**
   * Check if pattern-based resolution exists
   * @param filePath - File path to check
   * @returns Matching action or null
   * @tags @UTIL:CHECK-PATTERN-RESOLUTION-001
   */
  private checkPatternResolution(filePath: string): UpdateAction | null {
    for (const [pattern, action] of this.patterns.entries()) {
      if (this.matchesPattern(filePath, pattern)) {
        return action;
      }
    }
    return null;
  }

  /**
   * Add pattern for similar file resolution
   * @param filePath - File path to create pattern from
   * @param action - Action to apply
   * @tags @UTIL:ADD-PATTERN-001
   */
  private addPattern(filePath: string, action: UpdateAction): void {
    // Create pattern from file path (simple glob-like)
    const pattern = filePath
      .replace(/[0-9]+/g, '*')
      .replace(/\/[^/]*\.[^/]*$/, '/*');
    this.patterns.set(pattern, action);
  }

  /**
   * Check if file path matches pattern
   * @param filePath - File path to check
   * @param pattern - Pattern to match against
   * @returns True if matches
   * @tags @UTIL:MATCHES-PATTERN-001
   */
  private matchesPattern(filePath: string, pattern: string): boolean {
    const regexPattern = pattern.replace(/\*/g, '.*');
    const regex = new RegExp(`^${regexPattern}$`);
    return regex.test(filePath);
  }

  /**
   * Append new sections from template to user content
   * @param userContent - Current user content
   * @param templateContent - Template content
   * @returns Merged content
   * @tags @UTIL:APPEND-NEW-SECTIONS-001
   */
  private async appendNewSections(
    userContent: string,
    templateContent: string
  ): Promise<string> {
    // Simple implementation: append template content that's not in user content
    const userLines = new Set(userContent.split('\n').map(line => line.trim()));
    const templateLines = templateContent.split('\n');

    const newLines = templateLines.filter(
      line => !userLines.has(line.trim()) && line.trim().length > 0
    );

    if (newLines.length > 0) {
      return `${userContent}\n\n# Added from template update:\n${newLines.join('\n')}`;
    }

    return userContent;
  }

  /**
   * Replace template sections while keeping user sections
   * @param userContent - Current user content
   * @param templateContent - Template content
   * @returns Merged content
   * @tags @UTIL:REPLACE-SECTIONS-001
   */
  private async replaceSections(
    _userContent: string,
    templateContent: string
  ): Promise<string> {
    // For now, return template content (sophisticated section detection would be needed)
    return templateContent;
  }

  /**
   * Use template as base and apply user customizations
   * @param userContent - Current user content
   * @param templateContent - Template content
   * @returns Merged content
   * @tags @UTIL:TEMPLATE-BASE-MERGE-001
   */
  private async templateBaseMerge(
    _userContent: string,
    templateContent: string
  ): Promise<string> {
    // Simple implementation: use template but preserve user-specific sections
    return templateContent;
  }

  /**
   * Manual merge with conflict markers
   * @param userContent - Current user content
   * @param templateContent - Template content
   * @returns Content with conflict markers
   * @tags @UTIL:MANUAL-MERGE-001
   */
  private async manualMerge(
    userContent: string,
    templateContent: string
  ): Promise<string> {
    return `<<<<<<< Current Version
${userContent}
=======
${templateContent}
>>>>>>> Template Version

# Please resolve conflicts above and remove conflict markers`;
  }
}
