// @CODE:UTIL-005 | Refactored to validators/ module
// Related: @CODE:VALID-002:API

/**
 * @file Input validation utilities (now delegating to validators/ module)
 * @author MoAI Team
 * @deprecated Use @/utils/validators instead
 */

// Re-export everything from validators module for backward compatibility
export {
  InputValidator,
  type PathValidationOptions,
  type ProjectNameOptions,
  type ValidationResult,
  validateBranchName,
  validatePath,
  validateProjectName,
} from './validators/index';
