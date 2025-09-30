// @CODE:REFACTOR-004 연결: @TEST:REFACTOR-004 -> @CODE:REFACTOR-004
/**
 * @file Git Commit Message Templates
 * @author MoAI Team
 * @tags @CODE:REFACTOR-004 @CODE:GIT-COMMIT-TEMPLATES-001:DATA
 * @description Git 커밋 메시지 템플릿 및 생성 유틸리티
 */

/**
 * Git 커밋 메시지 템플릿
 * @tags @CODE:GIT-COMMIT-TEMPLATES-001:DATA
 */
export const GitCommitTemplates = {
  FEATURE: '✨ feat: {message}',
  BUGFIX: '🐛 fix: {message}',
  DOCS: '📝 docs: {message}',
  REFACTOR: '♻️ refactor: {message}',
  TEST: '✅ test: {message}',
  CHORE: '🔧 chore: {message}',
  STYLE: '💄 style: {message}',
  PERF: '⚡ perf: {message}',
  BUILD: '👷 build: {message}',
  CI: '💚 ci: {message}',
  REVERT: '⏪ revert: {message}',

  /**
   * 템플릿에 메시지 적용
   */
  apply: (template: string, message: string): string => {
    return template.replace('{message}', message);
  },

  /**
   * 자동 커밋 메시지 생성
   */
  createAutoCommit: (type: string, scope?: string): string => {
    const emoji = GitCommitTemplates.getEmoji(type);
    const prefix = scope ? `${type}(${scope})` : type;
    return `${emoji} ${prefix}: Auto-generated commit`;
  },

  /**
   * 체크포인트 커밋 메시지 생성
   */
  createCheckpoint: (message: string): string => {
    return `🔖 checkpoint: ${message}`;
  },

  /**
   * 타입별 이모지 반환
   */
  getEmoji: (type: string): string => {
    const emojiMap: Record<string, string> = {
      feat: '✨',
      fix: '🐛',
      docs: '📝',
      refactor: '♻️',
      test: '✅',
      chore: '🔧',
      style: '💄',
      perf: '⚡',
      build: '👷',
      ci: '💚',
      revert: '⏪',
    };
    return emojiMap[type] || '📝';
  },
} as const;
