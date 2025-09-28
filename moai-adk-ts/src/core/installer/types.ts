/**
 * @file Installation system type definitions
 * @author MoAI Team
 * @tags @DESIGN:INSTALLER-TYPES-001 @REQ:INSTALL-SYSTEM-012
 */

/**
 * Installation configuration interface
 * @tags @DESIGN:INSTALL-CONFIG-001
 */
export interface InstallationConfig {
  readonly projectPath: string;
  readonly projectName: string;
  readonly mode: 'personal' | 'team';
  readonly backupEnabled: boolean;
  readonly overwriteExisting: boolean;
  readonly templatePath?: string;
  readonly additionalFeatures: readonly string[];
}

/**
 * Installation progress callback type
 * @tags @DESIGN:PROGRESS-CALLBACK-001
 */
export type ProgressCallback = (message: string, current: number, total: number) => void;

/**
 * Installation result interface
 * @tags @DESIGN:INSTALL-RESULT-001
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
 * @tags @DESIGN:PHASE-STATUS-001
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
 * @tags @DESIGN:INSTALL-CONTEXT-001
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
 * @tags @DESIGN:SECURITY-VALIDATION-001
 */
export interface SecurityValidationResult {
  readonly valid: boolean;
  readonly issues: readonly string[];
  readonly recommendations: readonly string[];
}