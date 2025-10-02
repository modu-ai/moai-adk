// @CODE:GIT-003 |
// Related: @CODE:GIT-003:API, @CODE:GIT-GH-001

/**
 * @file GitHub API integration (backward compatibility barrel)
 * @author MoAI Team
 *
 * @fileoverview Re-export from refactored github/ module
 * @deprecated Import from './github/index' instead
 */

export type { RepositoryInfo } from './github/index';
export { GitHubIntegration } from './github/index';
