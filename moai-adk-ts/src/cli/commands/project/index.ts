// @CODE:INIT-003:MERGE | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:DATA, @CODE:INIT-003:UI, @SPEC:INIT-003

/**
 * @file Project Command - Backup merge and project management
 * @author MoAI Team
 * @tags @CODE:INIT-003:MERGE
 */

// Export backup merger
export { BackupMerger } from './backup-merger.js';
export type { MergeReport } from './backup-merger.js';

// Export merge report generator
export { generateMergeReport } from './merge-report.js';

// Export merge strategies
export { mergeJSON } from './merge-strategies/json-merger.js';
export { mergeMarkdown } from './merge-strategies/markdown-merger.js';
export { mergeHooks } from './merge-strategies/hooks-merger.js';
