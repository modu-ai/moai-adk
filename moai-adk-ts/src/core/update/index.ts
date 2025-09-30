/**
 * @file Update strategy module exports
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:UPDATE-INDEX-001 -> @TEST:UPDATE-INDEX-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

// Main orchestrator (only implemented module)
export { UpdateOrchestrator } from './update-orchestrator.js';
export type {
  UpdateConfiguration,
  UpdateOperationResult,
  UpdateSummary,
} from './update-orchestrator.js';
