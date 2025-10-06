/**
 * @file Configuration system type definitions
 * @author MoAI Team
 * @tags @SPEC:CONFIG-TYPES-001 @SPEC:CORE-SYSTEM-013
 */

/**
 * Claude Code settings interface
 * @tags @SPEC:CLAUDE-SETTINGS-001
 */
export interface ClaudeSettings {
  mode: 'personal' | 'team';
  agents: {
    enabled: string[];
    disabled: string[];
  };
  commands: {
    enabled: string[];
    shortcuts: Record<string, string>;
  };
  hooks: {
    enabled: string[];
    configuration: Record<string, any>;
  };
  outputStyles: {
    default: string;
    available: string[];
  };
  features: {
    autoSync: boolean;
    gitIntegration: boolean;
    tagTracking: boolean;
  };
  security: {
    allowedCommands: string[];
    blockedPatterns: string[];
    requireApproval: string[];
  };
}

/**
 * MoAI configuration interface
 * Unified with template structure (templates/.moai/config.json)
 * @tags @SPEC:MOAI-CONFIG-001
 */
export interface MoAIConfig {
  _meta?: {
    '@CODE:CONFIG-STRUCTURE-001'?: string;
    '@SPEC:PROJECT-CONFIG-001'?: string;
  };

  project: {
    name: string;
    version: string;
    mode: 'personal' | 'team';
    description?: string;
    initialized: boolean;
    created_at: string;
    locale?: 'ko' | 'en';
  };

  constitution: {
    enforce_tdd: boolean;
    require_tags: boolean;
    test_coverage_target: number;
    simplicity_threshold: number;
    principles: {
      simplicity: {
        max_projects: number;
        notes: string;
      };
    };
  };

  git_strategy: {
    personal: {
      auto_checkpoint: boolean;
      auto_commit: boolean;
      branch_prefix: string;
      checkpoint_interval: number;
      cleanup_days: number;
      max_checkpoints: number;
    };
    team: {
      auto_pr: boolean;
      develop_branch: string;
      draft_pr: boolean;
      feature_prefix: string;
      main_branch: string;
      use_gitflow: boolean;
    };
  };

  tags: {
    auto_sync: boolean;
    storage_type: 'code_scan';
    categories: string[];
    code_scan_policy: {
      no_intermediate_cache: boolean;
      realtime_validation: boolean;
      scan_tools: string[];
      scan_command: string;
      philosophy: string;
    };
  };

  pipeline: {
    available_commands: string[];
    current_stage: string;
  };
}

/**
 * Package.json configuration interface
 * @tags @SPEC:PACKAGE-CONFIG-001
 */
export interface PackageConfig {
  name: string;
  version: string;
  description: string;
  main: string;
  scripts: Record<string, string>;
  dependencies: Record<string, string>;
  devDependencies: Record<string, string>;
  keywords: string[];
  author: string;
  license: string;
  engines: {
    node: string;
  };
}

/**
 * Configuration creation result
 * @tags @SPEC:CONFIG-RESULT-001
 */
export interface ConfigResult {
  success: boolean;
  filePath?: string;
  error?: string;
  backupCreated?: boolean;
  skipped?: boolean;
  reason?: string;
}

/**
 * Claude settings creation result
 * @tags @SPEC:CLAUDE-RESULT-001
 */
export interface ClaudeSettingsResult extends ConfigResult {
  settings?: ClaudeSettings;
}

/**
 * MoAI config creation result
 * @tags @SPEC:MOAI-RESULT-001
 */
export interface MoAIConfigResult extends ConfigResult {
  config?: MoAIConfig;
}

/**
 * Package.json creation result
 * @tags @SPEC:PACKAGE-RESULT-001
 */
export interface PackageJsonResult extends ConfigResult {
  packageConfig?: PackageConfig;
}

/**
 * Configuration validation result
 * @tags @SPEC:VALIDATION-RESULT-001
 */
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  suggestions: string[];
}

/**
 * Configuration backup result
 * @tags @SPEC:BACKUP-RESULT-001
 */
export interface BackupResult {
  success: boolean;
  backupPath?: string;
  error?: string;
  timestamp: Date;
}

/**
 * Full project configuration result
 * @tags @SPEC:FULL-CONFIG-RESULT-001
 */
export interface FullConfigResult {
  success: boolean;
  filesCreated: string[];
  errors: string[];
  claudeSettings?: ClaudeSettingsResult;
  moaiConfig?: MoAIConfigResult;
  packageJson?: PackageJsonResult;
  duration: number;
  timestamp: Date;
}

/**
 * Project configuration input
 * @tags @SPEC:PROJECT-CONFIG-INPUT-001
 */
export interface ProjectConfigInput {
  projectName: string;
  mode: 'personal' | 'team';
  runtime: {
    name: string;
    version?: string;
  };
  techStack: string[];
  shouldCreatePackageJson?: boolean;
  overwrite?: boolean;
  backup?: boolean;
}

/**
 * Configuration file info
 * @tags @SPEC:CONFIG-FILE-INFO-001
 */
export interface ConfigFileInfo {
  path: string;
  exists: boolean;
  size?: number;
  lastModified?: Date;
  isValid?: boolean;
}

/**
 * Configuration status summary
 * @tags @SPEC:CONFIG-STATUS-001
 */
export interface ConfigStatus {
  claudeSettings: ConfigFileInfo;
  moaiConfig: ConfigFileInfo;
  packageJson: ConfigFileInfo;
  overall: {
    configured: boolean;
    missingFiles: string[];
    invalidFiles: string[];
  };
}
