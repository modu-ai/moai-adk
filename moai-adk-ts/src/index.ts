/**
 * @file Main API entry point
 * @author MoAI Team
 * @tags @FEATURE:MAIN-API-001 @REQ:PACKAGE-CONFIG-012
 */

// Core system checker
export * from './core/system-checker';

// CLI components
export { CLIApp } from './cli';
export { DoctorCommand } from './cli/commands/doctor';
export { InitCommand } from './cli/commands/init';

// Utilities
export { Logger, logger, LogLevel, LogEntry } from './utils/logger';
export {
  getPackageInfo,
  getCurrentVersion,
  PackageInfo,
} from './utils/version';
