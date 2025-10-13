/**
 * @CODE:SESSION-NOTICE-001:TYPES |
 * Related: @CODE:SESSION-NOTICE-001
 *
 * Session Notice Type Definitions
 */

/**
 * Project status information
 */
export interface ProjectStatus {
  projectName: string;
  moaiVersion: string; // moai-adk package version (internally from config.moai.version)
  initialized: boolean;
  constitutionStatus: ConstitutionStatus;
  pipelineStage: string;
  specProgress: SpecProgress;
}

/**
 * Constitution status information
 */
export interface ConstitutionStatus {
  status: 'ok' | 'violations_found' | 'not_initialized';
  violations: string[];
}

/**
 * SPEC progress information
 */
export interface SpecProgress {
  total: number;
  completed: number;
}

/**
 * Git information
 */
export interface GitInfo {
  branch: string;
  commit: string;
  message: string;
  changesCount: number;
}

/**
 * Version check result
 */
export interface VersionCheckResult {
  current: string;
  latest: string | null;
  hasUpdate: boolean;
}
