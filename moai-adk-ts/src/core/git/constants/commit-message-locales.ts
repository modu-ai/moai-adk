// @CODE:GIT-LOCALE-001 | SPEC: SPEC-GIT-LOCALE-001.md
/**
 * @file Locale-based commit message templates
 * @author MoAI Team
 * @tags @CODE:GIT-LOCALE-001:DATA
 * @description TDD 단계별 커밋 메시지 다국어 지원
 */

/**
 * Supported locales for commit messages
 * Extends i18n.ts Locale type to include ja and zh
 */
export type CommitLocale = 'ko' | 'en' | 'ja' | 'zh';

/**
 * TDD Stage types
 */
export type TDDStage = 'RED' | 'GREEN' | 'REFACTOR' | 'DOCS';

/**
 * TDD stage commit message templates by locale
 */
export interface TDDCommitTemplates {
  RED: string;
  GREEN: string;
  REFACTOR: string;
  DOCS: string;
}

/**
 * Korean commit message templates
 */
const koTemplates: TDDCommitTemplates = {
  RED: '🔴 RED: {message}',
  GREEN: '🟢 GREEN: {message}',
  REFACTOR: '♻️ REFACTOR: {message}',
  DOCS: '📝 DOCS: {message}',
};

/**
 * English commit message templates
 */
const enTemplates: TDDCommitTemplates = {
  RED: '🔴 RED: {message}',
  GREEN: '🟢 GREEN: {message}',
  REFACTOR: '♻️ REFACTOR: {message}',
  DOCS: '📝 DOCS: {message}',
};

/**
 * Japanese commit message templates
 */
const jaTemplates: TDDCommitTemplates = {
  RED: '🔴 RED: {message}',
  GREEN: '🟢 GREEN: {message}',
  REFACTOR: '♻️ REFACTOR: {message}',
  DOCS: '📝 DOCS: {message}',
};

/**
 * Chinese commit message templates
 */
const zhTemplates: TDDCommitTemplates = {
  RED: '🔴 RED: {message}',
  GREEN: '🟢 GREEN: {message}',
  REFACTOR: '♻️ REFACTOR: {message}',
  DOCS: '📝 DOCS: {message}',
};

/**
 * All locale templates
 */
const localeTemplates: Record<CommitLocale, TDDCommitTemplates> = {
  ko: koTemplates,
  en: enTemplates,
  ja: jaTemplates,
  zh: zhTemplates,
};

/**
 * Get TDD commit message template for locale
 * @param locale - Target locale
 * @param stage - TDD stage
 * @param message - Commit message content
 * @returns Formatted commit message
 */
export function getTDDCommitMessage(
  locale: CommitLocale,
  stage: TDDStage,
  message: string
): string {
  const templates = localeTemplates[locale] || localeTemplates.en;
  const template = templates[stage];
  return template.replace('{message}', message);
}

/**
 * Get TDD commit message with @TAG
 * @param locale - Target locale
 * @param stage - TDD stage
 * @param message - Commit message content
 * @param specId - SPEC ID for @TAG
 * @returns Formatted commit message with @TAG
 */
export function getTDDCommitWithTag(
  locale: CommitLocale,
  stage: TDDStage,
  message: string,
  specId: string
): string {
  const commitMessage = getTDDCommitMessage(locale, stage, message);
  const tagSuffix = getTagSuffix(stage, specId);
  return `${commitMessage}\n\n${tagSuffix}`;
}

/**
 * Get @TAG suffix for commit message
 * @param stage - TDD stage
 * @param specId - SPEC ID
 * @returns @TAG suffix
 */
function getTagSuffix(stage: TDDStage, specId: string): string {
  switch (stage) {
    case 'RED':
      return `@TEST:${specId}-RED`;
    case 'GREEN':
      return `@CODE:${specId}-GREEN`;
    case 'REFACTOR':
      return `REFACTOR:${specId}-CLEAN`;
    case 'DOCS':
      return `@DOC:${specId}`;
    default:
      return `@TAG:${specId}`;
  }
}

/**
 * Validate locale
 * @param locale - Locale string to validate
 * @returns True if valid commit locale
 */
export function isValidCommitLocale(locale: string): locale is CommitLocale {
  return ['ko', 'en', 'ja', 'zh'].includes(locale);
}

/**
 * Get default locale (fallback to 'en')
 * @param locale - Requested locale
 * @returns Valid commit locale
 */
export function getValidatedLocale(locale?: string): CommitLocale {
  if (locale && isValidCommitLocale(locale)) {
    return locale;
  }
  return 'en';
}

/**
 * Export templates for testing
 */
export const CommitMessageTemplates = {
  ko: koTemplates,
  en: enTemplates,
  ja: jaTemplates,
  zh: zhTemplates,
  localeTemplates,
} as const;
