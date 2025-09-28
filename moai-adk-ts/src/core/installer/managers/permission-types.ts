/**
 * @file Permission management type definitions
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-TYPES-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

/**
 * File permission structure interface
 * @tags @DESIGN:FILE-PERMISSIONS-001
 */
export interface FilePermissions {
  readonly owner: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
  readonly group: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
  readonly others: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
}

/**
 * Permission status for a file or directory
 * @tags @DESIGN:PERMISSION-STATUS-001
 */
export interface PermissionStatus {
  readonly path: string;
  readonly readable: boolean;
  readonly writable: boolean;
  readonly executable: boolean;
  readonly octalMode?: string; // Unix systems only
  readonly isOwner: boolean;
}

/**
 * Result of permission fix operation
 * @tags @DESIGN:PERMISSION-FIX-RESULT-001
 */
export interface PermissionFixResult {
  readonly success: boolean;
  readonly fixedFiles: readonly string[];
  readonly failures: readonly string[];
  readonly warnings: readonly string[];
}

/**
 * Platform type detection
 * @tags @DESIGN:PLATFORM-TYPE-001
 */
export type PlatformType = 'windows' | 'unix';

/**
 * Standard MoAI permission policies
 * @tags @DESIGN:MOAI-PERMISSIONS-001
 */
export const MoAIPermissions = {
  SCRIPT_FILES: {
    // .py, .sh, .js files
    owner: { read: true, write: true, execute: true },
    group: { read: true, write: false, execute: true },
    others: { read: true, write: false, execute: false },
  },
  CONFIG_FILES: {
    // .json, .yaml files
    owner: { read: true, write: true, execute: false },
    group: { read: true, write: false, execute: false },
    others: { read: false, write: false, execute: false },
  },
  DIRECTORIES: {
    owner: { read: true, write: true, execute: true },
    group: { read: true, write: false, execute: true },
    others: { read: true, write: false, execute: true },
  },
  SENSITIVE_FILES: {
    // secret files
    owner: { read: true, write: true, execute: false },
    group: { read: false, write: false, execute: false },
    others: { read: false, write: false, execute: false },
  },
} as const;
