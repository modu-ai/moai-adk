/**
 * @file Type definitions for init prompts
 * @author MoAI Team
 * @tags @CODE:INSTALL-001 | Chain: @SPEC:INSTALL-001 -> @CODE:INSTALL-001 -> @TEST:INSTALL-001
 * Related: @CODE:INIT-TYPES-001, @DOC:INSTALL-001
 */

import type { Locale } from '@/utils/i18n';

/**
 * User answers from interactive prompts
 * Extended with SPEC-INSTALL-001 requirements
 */
export interface InitAnswers {
  locale?: Locale | undefined;
  projectName: string;
  mode: 'personal' | 'team';
  gitEnabled: boolean;
  githubEnabled?: boolean | undefined;
  githubUrl?: string | undefined;
  specWorkflow?: 'commit' | 'branch' | undefined;
  autoPush?: boolean | undefined;

  // SPEC-INSTALL-001: Developer information
  developerName?: string | undefined;

  // SPEC-INSTALL-001: SPEC workflow enforcement
  enforceSpec?: boolean | undefined;

  // SPEC-INSTALL-001: PR configuration
  autoPR?: boolean | undefined;
  draftPR?: boolean | undefined;
}

/**
 * Partial initialization answers for incremental collection
 */
export type PartialInitAnswers = Partial<InitAnswers>;
