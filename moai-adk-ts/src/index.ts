/**
 * @file Main API entry point
 * @author MoAI Team
 * @tags @FEATURE:MAIN-API-001 @REQ:PACKAGE-CONFIG-012
 */

// CLI components
export { CLIApp } from './cli';
export { DoctorCommand } from './cli/commands/doctor';
export { InitCommand } from './cli/commands/init';
// Core project components
export { TemplateManager } from './core/project/template-manager';
// Core system checker
export * from './core/system-checker';

// Utilities
export { LogEntry, LogLevel } from './types/logger';
export { MoaiLogger as Logger, logger } from './utils/winston-logger';
export {
  getCurrentVersion,
  getPackageInfo,
  PackageInfo,
} from './utils/version';
