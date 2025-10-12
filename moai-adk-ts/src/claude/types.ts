/**
 * @CODE:HOOK-TYPES-001 |
 * Related: @CODE:HOOK-001, @CODE:HOOK-002, @CODE:HOOK-003, @CODE:HOOK-004
 *
 * Claude Code Hook System Type Definitions
 * MoAI 훅 시스템의 공통 타입 정의
 */

/**
 * Hook input from Claude Code
 * Claude Code에서 전달되는 훅 입력 데이터
 */
export interface HookInput {
  tool_name: string;
  tool_input?: Record<string, any>;
  context?: {
    working_directory?: string;
    project_root?: string;
    user?: string;
    [key: string]: any;
  };
}

/**
 * Hook execution result
 * 훅 실행 결과
 */
export interface HookResult {
  success: boolean;
  blocked?: boolean;
  message?: string;
  exitCode?: number;
  data?: Record<string, any>;
  warnings?: string[];
  suggestions?: string[];
}

/**
 * MoAI Hook interface
 * 모든 훅이 구현해야 하는 인터페이스
 */
export interface MoAIHook {
  name: string;
  execute(input?: HookInput): Promise<HookResult>;
}

/**
 * Hook configuration
 * 훅 설정 인터페이스
 */
export interface HookConfig {
  enabled: boolean;
  priority?: number;
  options?: Record<string, any>;
}
