// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file Alfred welcome message tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:WELCOME-MESSAGE
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('@TEST:INSTALL-001 - Alfred Welcome Message', () => {
  let consoleLogSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleLogSpy.mockRestore();
  });

  describe('ER-006: 환영 메시지 출력', () => {
    it('should display welcome message after installation', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toContain('✅ MoAI-ADK 설치가 완료되었습니다!');
    });

    it('should include Alfred persona greeting', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toContain('🤖 AI-Agent Alfred가');
      expect(output).toContain('홍길동님의 개발을 도와드리겠습니다');
    });

    it('should provide next steps guidance', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert - AC6 요구사항
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toContain('다음 명령어로 시작하세요');
      expect(output).toContain('/alfred:0-project');
      expect(output).toContain('/alfred:1-spec');
    });

    it('should suggest debug-helper for questions', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toContain('@agent-debug-helper');
    });
  });

  describe('SR-003: Alfred 페르소나 톤 유지', () => {
    it('should use polite Korean language', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toContain('도와드리겠습니다');
      expect(output).toContain('시작하세요');
      expect(output).toContain('호출하세요');
    });

    it('should use appropriate emojis', async () => {
      // Act
      const { displayWelcomeMessage } = await import(
        '@/cli/prompts/init/welcome-message'
      );
      displayWelcomeMessage({ developerName: '홍길동' });

      // Assert
      const output = consoleLogSpy.mock.calls.join('\n');
      expect(output).toMatch(/✅/);
      expect(output).toMatch(/🤖/);
    });
  });
});
