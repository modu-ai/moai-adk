/**
 * @FEATURE-COMMIT-VALIDATOR-001: TypeScript CommitValidator 구현
 * @연결: @TASK:VALIDATION-001 → @FEATURE:COMMIT-VALIDATION-001 → @API:VALIDATE-COMMIT
 */

export interface ValidationResult {
  valid: boolean;
  reason: string;
}

export class CommitValidator {
  private static readonly MIN_MESSAGE_LENGTH = 10;
  private static readonly MAX_MESSAGE_LENGTH = 200;
  private static readonly EMOJI_PATTERN =
    /[\u{1F600}-\u{1F64F}\u{1F300}-\u{1F5FF}\u{1F680}-\u{1F6FF}\u{1F1E0}-\u{1F1FF}]/u;

  validateCommitMessage(message: string): ValidationResult {
    // Guard clause: Empty message check
    if (!message || !message.trim()) {
      return this.createValidationResult(false, '메시지가 비어있습니다');
    }

    // Guard clause: Length validation
    const lengthValidation = this.validateLength(message);
    if (!lengthValidation.valid) {
      return lengthValidation;
    }

    // Additional validations (emoji check for future use)
    this.checkEmojiPresence(message);

    return this.createValidationResult(true, '검증 통과');
  }

  validateFileList(files?: string[]): ValidationResult {
    if (!files) {
      return { valid: true, reason: '전체 파일 커밋' };
    }

    if (files.length === 0) {
      return { valid: false, reason: '커밋할 파일이 지정되지 않았습니다' };
    }

    // 파일 경로 유효성 검사
    for (const file of files) {
      if (!file.trim()) {
        return { valid: false, reason: '빈 파일 경로가 포함되어 있습니다' };
      }
    }

    return { valid: true, reason: '파일 목록 유효' };
  }

  validateChangeContext(changes: any): ValidationResult {
    if (!changes.success) {
      return { valid: false, reason: '변경사항 조회 실패' };
    }

    if (!changes.has_changes) {
      return { valid: false, reason: 'No changes to commit' };
    }

    return { valid: true, reason: '변경사항 존재' };
  }

  private validateLength(message: string): ValidationResult {
    const length = message.trim().length;

    if (length < CommitValidator.MIN_MESSAGE_LENGTH) {
      return this.createValidationResult(
        false,
        `메시지가 너무 짧습니다 (최소 ${CommitValidator.MIN_MESSAGE_LENGTH}자)`
      );
    }

    if (length > CommitValidator.MAX_MESSAGE_LENGTH) {
      return this.createValidationResult(
        false,
        `메시지가 너무 깁니다 (최대 ${CommitValidator.MAX_MESSAGE_LENGTH}자)`
      );
    }

    return this.createValidationResult(true, '길이 검증 통과');
  }

  private checkEmojiPresence(message: string): boolean {
    return CommitValidator.EMOJI_PATTERN.test(message);
  }

  private createValidationResult(
    valid: boolean,
    reason: string
  ): ValidationResult {
    return { valid, reason };
  }
}
