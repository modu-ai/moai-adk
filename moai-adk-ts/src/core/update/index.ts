/**
 * @file Update strategy module exports
 * @author MoAI Team
 * @tags @CODE:UPDATE-STRATEGY-001 | Chain: @SPEC:UPDATE-REAL-001 -> @SPEC:UPDATE-STRATEGY-001 -> @CODE:UPDATE-INDEX-001 -> @TEST:UPDATE-INDEX-001
 * Related: @CODE:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

// Main orchestrator (only implemented module)
export { UpdateOrchestrator } from './update-orchestrator.js';
export type {
  UpdateConfiguration,
  UpdateOperationResult,
  UpdateSummary,
} from './update-orchestrator.js';
