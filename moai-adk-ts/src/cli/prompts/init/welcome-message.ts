// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/welcome-message.test.ts
/**
 * @file Alfred welcome message display
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:WELCOME-MESSAGE
 */

/**
 * Welcome message configuration
 */
export interface WelcomeConfig {
  developerName: string;
}

/**
 * Display Alfred welcome message
 * @param config Welcome configuration with developer name
 */
export function displayWelcomeMessage(config: WelcomeConfig): void {
  console.log(`
✅ MoAI-ADK 설치가 완료되었습니다!

🤖 AI-Agent Alfred가 ${config.developerName}님의 개발을 도와드리겠습니다.

다음 명령어로 시작하세요:
/alfred:0-project  # 프로젝트 초기화
/alfred:1-spec     # 첫 SPEC 작성

질문이 있으시면 언제든 @agent-debug-helper를 호출하세요.
`);
}
