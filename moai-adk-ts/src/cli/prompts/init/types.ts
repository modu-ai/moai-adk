/**
 * @file Type definitions for init prompts
 * @author MoAI Team
 * @tags @CODE:INIT-TYPES-001 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

import type { Locale } from '@/utils/i18n';

/**
 * User answers from interactive prompts
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
}

/**
 * Partial initialization answers for incremental collection
 */
export type PartialInitAnswers = Partial<InitAnswers>;
