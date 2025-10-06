/**
 * @file Configuration builder from interactive prompts
 * @author MoAI Team
 * @tags @CODE:INSTALL-001 | Chain: @SPEC:INSTALL-001 -> @CODE:INSTALL-001 -> @TEST:INSTALL-001
 * Related: @CODE:INTERACTIVE-INIT-019, @DOC:INSTALL-001
 */

import type { InitAnswers } from '../prompts/init';

/**
 * Enhanced MoAI configuration interface
 * Extended with SPEC-INSTALL-001 requirements
 */
export interface MoAIConfig {
  version: string;
  mode: 'personal' | 'team';
  projectName: string;
  features: string[];
  locale?: 'ko' | 'en'; // User's preferred CLI language

  // SPEC-INSTALL-001: Developer information
  developer?: {
    name: string;
    timestamp: string;
  };

  git: {
    enabled: boolean;
    autoCommit: boolean;
    branchPrefix: string;
    remote?: {
      enabled: boolean;
      url: string;
      autoPush: boolean;
      defaultBranch: string;
    };
  };

  // SPEC-INSTALL-001: Constitution with enforce_spec
  constitution?: {
    enforce_tdd?: boolean;
    enforce_spec?: boolean;
    require_tags?: boolean;
    test_coverage_target?: number;
  };

  // SPEC-INSTALL-001: Git strategy with auto_pr and draft_pr
  git_strategy?: {
    team?: {
      auto_pr?: boolean;
      draft_pr?: boolean;
    };
  };

  spec: {
    storage: 'local' | 'github';
    workflow: 'commit' | 'branch';
    localPath: string;
    github?: {
      issueLabels: string[];
      templatePath: string;
    };
  };

  backup: {
    enabled: boolean;
    retentionDays: number;
  };
}

/**
 * Configuration builder from interactive prompts
 */
export class ConfigBuilder {
  /**
   * Build MoAI configuration from prompt answers
   * @param answers User answers from interactive prompts
   * @returns Complete MoAI configuration
   */
  public buildConfig(answers: InitAnswers): MoAIConfig {
    const config: MoAIConfig = {
      version: '0.0.1',
      mode: answers.mode,
      projectName: answers.projectName,
      features: [],
      locale: answers.locale || 'ko', // Default to Korean if not specified

      git: this.buildGitConfig(answers),
      spec: this.buildSpecConfig(answers),
      backup: {
        enabled: true,
        retentionDays: 30,
      },
    };

    // SPEC-INSTALL-001: Add developer information if provided
    if (answers.developerName) {
      config.developer = {
        name: answers.developerName,
        timestamp: new Date().toISOString(),
      };
    }

    // SPEC-INSTALL-001: Add constitution configuration
    config.constitution = {
      enforce_tdd: true,
      enforce_spec: answers.enforceSpec ?? answers.mode === 'team',
      require_tags: true,
      test_coverage_target: 85,
    };

    // SPEC-INSTALL-001: Add git_strategy for Team mode
    if (answers.mode === 'team') {
      config.git_strategy = {
        team: {
          auto_pr: answers.autoPR ?? true,
          draft_pr: answers.draftPR ?? true,
        },
      };
    }

    return config;
  }

  /**
   * Build Git configuration
   */
  private buildGitConfig(answers: InitAnswers): MoAIConfig['git'] {
    const gitConfig: MoAIConfig['git'] = {
      enabled: answers.gitEnabled,
      autoCommit: true,
      branchPrefix: answers.mode === 'team' ? 'feature/' : '',
    };

    // Add remote configuration if GitHub is enabled
    if (answers.githubEnabled && answers.githubUrl) {
      gitConfig.remote = {
        enabled: true,
        url: answers.githubUrl,
        autoPush: answers.autoPush ?? true,
        defaultBranch: 'main',
      };
    }

    return gitConfig;
  }

  /**
   * Build SPEC configuration
   */
  private buildSpecConfig(answers: InitAnswers): MoAIConfig['spec'] {
    const specConfig: MoAIConfig['spec'] = {
      storage: answers.mode === 'team' ? 'github' : 'local',
      workflow: answers.specWorkflow || 'commit',
      localPath: '.moai/specs/',
    };

    // Add GitHub-specific configuration for team mode
    if (answers.mode === 'team' && answers.githubEnabled) {
      specConfig.github = {
        issueLabels: ['spec', 'requirements', 'moai-adk'],
        templatePath: '.github/ISSUE_TEMPLATE/spec.md',
      };
    }

    return specConfig;
  }

  /**
   * Get branch strategy from configuration
   * @param config MoAI configuration
   * @returns Branch strategy description
   */
  public getBranchStrategy(config: MoAIConfig): string {
    if (config.mode === 'personal') {
      return config.git.enabled
        ? 'Local Git with simple commits'
        : 'No version control';
    }

    if (config.spec.workflow === 'branch') {
      return 'GitHub PR workflow (feature branches)';
    }

    return 'Direct commits to main';
  }

  /**
   * Validate configuration
   * @param config MoAI configuration
   * @returns Validation result
   */
  public validateConfig(config: MoAIConfig): {
    valid: boolean;
    errors: string[];
  } {
    const errors: string[] = [];

    // Validate project name
    if (!config.projectName || config.projectName.trim().length === 0) {
      errors.push('Project name is required');
    }

    // Validate mode
    if (!['personal', 'team'].includes(config.mode)) {
      errors.push(`Invalid mode: ${config.mode}`);
    }

    // Validate GitHub URL if remote is enabled
    if (config.git.remote?.enabled && config.git.remote.url) {
      const githubRegex = /^https:\/\/github\.com\/[\w-]+\/[\w-]+$/;
      if (!githubRegex.test(config.git.remote.url)) {
        errors.push(`Invalid GitHub URL: ${config.git.remote.url}`);
      }
    }

    // Validate workflow consistency
    if (config.mode === 'team' && config.spec.workflow === 'branch') {
      if (!config.git.enabled) {
        errors.push('Branch workflow requires Git to be enabled');
      }
      if (!config.git.remote?.enabled) {
        errors.push('Branch workflow requires GitHub remote');
      }
    }

    return {
      valid: errors.length === 0,
      errors,
    };
  }

  /**
   * Get configuration file path
   * @param projectPath Project root path
   * @returns Path to config.json
   */
  public getConfigPath(projectPath: string): string {
    return `${projectPath}/.moai/config.json`;
  }
}

/**
 * Singleton instance
 */
export const configBuilder = new ConfigBuilder();

/**
 * Build MoAI config from interactive prompts (convenience function)
 * @param answers User answers from prompts
 * @returns Complete MoAI configuration
 */
export function buildMoAIConfig(answers: InitAnswers): MoAIConfig {
  return configBuilder.buildConfig(answers);
}
