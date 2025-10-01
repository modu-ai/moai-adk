/**
 * @file Installation system type definitions
 * @author MoAI Team
 * @tags @SPEC:INSTALLER-TYPES-001 @SPEC:INSTALL-SYSTEM-012
 */

/**
 * Installation configuration interface
 * @tags @SPEC:INSTALL-CONFIG-001
 */
export interface InstallationConfig {
  readonly projectPath: string;
  readonly projectName: string;
  readonly mode: 'personal' | 'team';
  readonly backupEnabled: boolean;
  readonly overwriteExisting: boolean;
  readonly templatePath?: string | undefined;
  readonly additionalFeatures: readonly string[];
}

/**
 * Installation progress callback type
 * @tags @SPEC:PROGRESS-CALLBACK-001
 */
export type ProgressCallback = (
  message: string,
  current: number,
  total: number
) => void;

/**
 * Installation result interface
 * @tags @SPEC:INSTALL-RESULT-001
 */
export interface InstallationResult {
  readonly success: boolean;
  readonly projectPath: string;
  readonly filesCreated: readonly string[];
  readonly errors: readonly string[];
  readonly nextSteps: readonly string[];
  readonly config: InstallationConfig;
  readonly timestamp: Date;
  readonly duration: number;
}

/**
 * Installation phase status
 * @tags @SPEC:PHASE-STATUS-001
 */
export interface PhaseStatus {
  readonly name: string;
  readonly completed: boolean;
  readonly duration: number;
  readonly errors: readonly string[];
  readonly filesCreated: readonly string[];
}

/**
 * Installation context for tracking progress
 * @tags @SPEC:INSTALL-CONTEXT-001
 */
export interface InstallationContext {
  readonly config: InstallationConfig;
  readonly startTime: Date;
  readonly phases: PhaseStatus[];
  readonly allFilesCreated: string[];
  readonly allErrors: string[];
}

/**
 * Security validation result
 * @tags @SPEC:SECURITY-VALIDATION-001
 */
export interface SecurityValidationResult {
  readonly valid: boolean;
  readonly issues: readonly string[];
  readonly recommendations: readonly string[];
}

/**
 * Post-installation configuration options
 * @tags @SPEC:POST-INSTALL-CONFIG-001
 */
export interface PostInstallOptions {
  readonly projectPath: string;
  readonly setupGlobal: boolean;
  readonly validateResources: boolean;
  readonly force: boolean;
  readonly quiet: boolean;
}

/**
 * Post-installation result
 * @tags @SPEC:POST-INSTALL-RESULT-001
 */
export interface PostInstallResult {
  readonly success: boolean;
  readonly isFirstRun: boolean;
  readonly resourcesValidated: boolean;
  readonly globalSetupCompleted: boolean;
  readonly firstRunSetupCompleted: boolean;
  readonly errors: readonly string[];
  readonly warnings: readonly string[];
  readonly duration: number;
  readonly timestamp: Date;
}

/**
 * Resource validation result
 * @tags @SPEC:RESOURCE-VALIDATION-001
 */
export interface ValidationResult {
  readonly isValid: boolean;
  readonly missingTemplates: readonly string[];
  readonly errors: readonly string[];
}

/**
 * Template check result
 * @tags @SPEC:TEMPLATE-CHECK-001
 */
export interface TemplateCheckResult {
  readonly allPresent: boolean;
  readonly missingTemplates: readonly string[];
  readonly templateCount: number;
}

/**
 * Resource integrity result
 * @tags @SPEC:INTEGRITY-CHECK-001
 */
export interface IntegrityResult {
  readonly isValid: boolean;
  readonly corruptedFiles: readonly string[];
  readonly checksumMismatches: readonly string[];
}

/**
 * First run state information
 * @tags @SPEC:FIRST-RUN-STATE-001
 */
export interface FirstRunState {
  readonly isFirstRun: boolean;
  readonly hasMarkerFile: boolean;
  readonly setupCompleted: boolean;
}

/**
 * Technology stack enum for memory template mapping
 * @tags @SPEC:TECH-STACK-001
 */
export type TechStack =
  | 'python'
  | 'fastapi'
  | 'django'
  | 'flask'
  | 'java'
  | 'spring'
  | 'spring boot'
  | 'springboot'
  | 'spring-boot'
  | 'react'
  | 'nextjs'
  | 'vue'
  | 'nuxt'
  | 'angular'
  | 'typescript'
  | 'javascript'
  | 'nodejs'
  | 'express'
  | 'nestjs'
  | 'rust'
  | 'go'
  | 'kotlin'
  | 'scala'
  | 'cpp'
  | 'c'
  | 'csharp'
  | 'dotnet';

/**
 * Memory template names enum
 * @tags @SPEC:MEMORY-TEMPLATE-001
 */
export type MemoryTemplate =
  | 'development-guide'
  | 'backend-python'
  | 'backend-fastapi'
  | 'backend-spring'
  | 'backend-express'
  | 'frontend-react'
  | 'frontend-next'
  | 'frontend-vue'
  | 'frontend-angular'
  | 'fullstack-patterns'
  | 'microservice-patterns';
