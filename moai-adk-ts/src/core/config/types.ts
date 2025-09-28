/**
 * @file Configuration system type definitions
 * @author MoAI Team
 * @tags @DESIGN:CONFIG-TYPES-001 @REQ:CORE-SYSTEM-013
 */

/**
 * Claude Code settings interface
 * @tags @DESIGN:CLAUDE-SETTINGS-001
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
 * @tags @DESIGN:MOAI-CONFIG-001
 */
export interface MoAIConfig {
  projectName: string;
  version: string;
  mode: 'personal' | 'team';
  runtime: {
    name: string;
    version?: string;
  };
  techStack: string[];
  features: {
    tdd: boolean;
    tagSystem: boolean;
    gitAutomation: boolean;
    documentSync: boolean;
  };
  directories: {
    moai: string;
    claude: string;
    specs: string;
    templates: string;
  };
  createdAt: Date;
  updatedAt: Date;
}

/**
 * Package.json configuration interface
 * @tags @DESIGN:PACKAGE-CONFIG-001
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
 * @tags @DESIGN:CONFIG-RESULT-001
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
 * @tags @DESIGN:CLAUDE-RESULT-001
 */
export interface ClaudeSettingsResult extends ConfigResult {
  settings?: ClaudeSettings;
}

/**
 * MoAI config creation result
 * @tags @DESIGN:MOAI-RESULT-001
 */
export interface MoAIConfigResult extends ConfigResult {
  config?: MoAIConfig;
}

/**
 * Package.json creation result
 * @tags @DESIGN:PACKAGE-RESULT-001
 */
export interface PackageJsonResult extends ConfigResult {
  packageConfig?: PackageConfig;
}

/**
 * Configuration validation result
 * @tags @DESIGN:VALIDATION-RESULT-001
 */
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  suggestions: string[];
}

/**
 * Configuration backup result
 * @tags @DESIGN:BACKUP-RESULT-001
 */
export interface BackupResult {
  success: boolean;
  backupPath?: string;
  error?: string;
  timestamp: Date;
}

/**
 * Full project configuration result
 * @tags @DESIGN:FULL-CONFIG-RESULT-001
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
 * @tags @DESIGN:PROJECT-CONFIG-INPUT-001
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
 * @tags @DESIGN:CONFIG-FILE-INFO-001
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
 * @tags @DESIGN:CONFIG-STATUS-001
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
