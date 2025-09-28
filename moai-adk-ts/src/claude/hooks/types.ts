/**
 * @file types.ts
 * @description Core type definitions for MoAI-ADK Claude Code hooks
 * @version 1.0.0
 * @tag @API:HOOK-TYPES-013
 */

/**
 * Input data structure for Claude Code hooks
 */
export interface HookInput {
  /** The user's prompt text */
  prompt?: string;
  /** List of files being processed */
  files?: string[];
  /** Action being performed */
  action?: string;
  /** Tool name being executed */
  tool_name?: string;
  /** Tool input parameters */
  tool_input?: Record<string, any>;
  /** Additional context data */
  context?: Record<string, any>;
}

/**
 * Output result structure for Claude Code hooks
 */
export interface HookResult {
  /** Whether the hook executed successfully */
  success: boolean;
  /** Whether the action should be blocked */
  blocked?: boolean;
  /** Human-readable message */
  message?: string;
  /** Additional data to return */
  data?: any;
  /** Exit code for Claude Code */
  exitCode?: number;
}

/**
 * Interface for all MoAI hooks
 */
export interface MoAIHook {
  /** Unique name of the hook */
  name: string;
  /** Execute the hook with given input */
  execute(input: HookInput): Promise<HookResult>;
}

/**
 * Security pattern for blocking dangerous content
 */
export interface SecurityPattern {
  /** Regular expression pattern */
  pattern: RegExp;
  /** Message to show when blocked */
  message: string;
  /** Severity level */
  severity?: 'low' | 'medium' | 'high' | 'critical';
}

/**
 * File monitoring event data
 */
export interface FileEvent {
  /** Path of the changed file */
  path: string;
  /** Type of change */
  type: 'created' | 'modified' | 'deleted';
  /** Timestamp of the event */
  timestamp: Date;
}

/**
 * Language detection result
 */
export interface LanguageInfo {
  /** Detected programming language */
  language: string;
  /** Confidence score (0-1) */
  confidence: number;
  /** Recommended test runner */
  testRunner?: string;
  /** Recommended linter */
  linter?: string;
  /** Recommended formatter */
  formatter?: string;
}

/**
 * Project status information
 */
export interface ProjectStatus {
  /** Project name */
  projectName: string;
  /** MoAI version */
  moaiVersion: string;
  /** Whether project is initialized */
  initialized: boolean;
  /** Development guide status */
  constitutionStatus: ConstitutionStatus;
  /** Current pipeline stage */
  pipelineStage: string;
  /** SPEC progress information */
  specProgress: SpecProgress;
}

/**
 * Development guide compliance status
 */
export interface ConstitutionStatus {
  /** Overall status */
  status: 'ok' | 'violations_found' | 'not_initialized';
  /** List of violations */
  violations: string[];
}

/**
 * SPEC progress tracking
 */
export interface SpecProgress {
  /** Total number of SPECs */
  total: number;
  /** Number of completed SPECs */
  completed: number;
}

/**
 * Test execution result
 */
export interface TestResult {
  /** Test runner name */
  runner: string;
  /** Exit code */
  exitCode: number;
  /** Standard output */
  stdout: string;
  /** Standard error */
  stderr: string;
  /** Execution time in milliseconds */
  duration: number;
}

/**
 * Configuration for hook system
 */
export interface HookConfig {
  /** Whether hooks are enabled */
  enabled: boolean;
  /** Timeout for hook execution in milliseconds */
  timeout: number;
  /** List of disabled hooks */
  disabledHooks: string[];
  /** Security settings */
  security: SecurityConfig;
}

/**
 * Security configuration
 */
export interface SecurityConfig {
  /** List of allowed commands */
  allowedCommands: string[];
  /** List of blocked patterns */
  blockedPatterns: string[];
  /** Commands requiring approval */
  requireApproval: string[];
}
