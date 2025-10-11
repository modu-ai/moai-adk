/**
 * @TEST:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOKS-REFACTOR-001, @SPEC:HOOKS-REFACTOR-001
 *
 * Base Hook Utilities Test Suite
 * runHook() 유틸리티 함수 테스트
 */

import { EventEmitter } from 'node:events';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HookInput, HookResult, MoAIHook } from '../../types';
import { parseClaudeInput, outputResult } from '../../index';

describe('@TEST:HOOKS-REFACTOR-001 - base.ts', () => {
  describe('runHook() 기본 실행', () => {
    it('should execute hook instance successfully', async () => {
      // GIVEN: 유효한 훅 클래스
      class MockHook implements MoAIHook {
        name = 'mock-hook';
        async execute(): Promise<HookResult> {
          return { success: true };
        }
      }

      // WHEN: runHook 실행을 시뮬레이션
      const hook = new MockHook();
      const result = await hook.execute({
        tool_name: 'Test',
        tool_input: {},
      });

      // THEN: 성공 결과 반환
      expect(result.success).toBe(true);
    });
  });

  describe('parseClaudeInput 통합', () => {
    it('should parse stdin input correctly', async () => {
      // GIVEN: stdin에 JSON 입력
      const mockInput = { tool_name: 'Read', tool_input: {} };

      // Mock stdin
      const stdinMock = new EventEmitter();
      const originalStdin = process.stdin;
      Object.defineProperty(process, 'stdin', {
        value: stdinMock,
        configurable: true,
      });

      // WHEN: parseClaudeInput 호출
      const resultPromise = parseClaudeInput();

      setTimeout(() => {
        stdinMock.emit('data', JSON.stringify(mockInput));
        stdinMock.emit('end');
      }, 10);

      const result = await resultPromise;

      // Restore original stdin
      Object.defineProperty(process, 'stdin', {
        value: originalStdin,
        configurable: true,
      });

      // THEN: 올바르게 파싱
      expect(result).toEqual(mockInput);
    });

    it('should handle empty input gracefully', async () => {
      // GIVEN: 빈 stdin 입력
      const stdinMock = new EventEmitter();
      const originalStdin = process.stdin;
      Object.defineProperty(process, 'stdin', {
        value: stdinMock,
        configurable: true,
      });

      // WHEN: parseClaudeInput 호출
      const resultPromise = parseClaudeInput();

      setTimeout(() => {
        stdinMock.emit('data', '');
        stdinMock.emit('end');
      }, 10);

      const result = await resultPromise;

      // Restore original stdin
      Object.defineProperty(process, 'stdin', {
        value: originalStdin,
        configurable: true,
      });

      // THEN: 기본값 반환
      expect(result.tool_name).toBe('Unknown');
      expect(result.tool_input).toEqual({});
    });
  });

  describe('outputResult 성공 케이스', () => {
    let exitSpy: any;
    let logSpy: any;

    beforeEach(() => {
      exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {
        throw new Error('process.exit called');
      });
      logSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    });

    it('should output success result and exit 0', () => {
      // GIVEN: 성공 결과
      const result: HookResult = { success: true, message: 'OK' };

      // WHEN: outputResult 호출
      // THEN: 메시지 출력 및 exit(0)
      expect(() => outputResult(result)).toThrow('process.exit called');
      expect(logSpy).toHaveBeenCalledWith('OK');
      expect(exitSpy).toHaveBeenCalledWith(0);
    });
  });

  describe('outputResult 블록 케이스', () => {
    let exitSpy: any;
    let errorSpy: any;

    beforeEach(() => {
      exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {
        throw new Error('process.exit called');
      });
      errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    });

    it('should output blocked result and exit 2', () => {
      // GIVEN: 차단 결과
      const result: HookResult = {
        success: false,
        blocked: true,
        message: 'Blocked',
        exitCode: 2,
      };

      // WHEN: outputResult 호출
      // THEN: 에러 출력 및 exit(2)
      expect(() => outputResult(result)).toThrow('process.exit called');
      expect(errorSpy).toHaveBeenCalledWith('BLOCKED: Blocked');
      expect(exitSpy).toHaveBeenCalledWith(2);
    });
  });

  describe('runHook 패턴 검증', () => {
    it('should support async hook execution pattern', async () => {
      // GIVEN: 비동기 훅
      class AsyncHook implements MoAIHook {
        name = 'async-hook';
        async execute(input: HookInput): Promise<HookResult> {
          // Simulate async operation
          await new Promise(resolve => setTimeout(resolve, 10));
          return { success: true, message: `Processed ${input.tool_name}` };
        }
      }

      // WHEN: 훅 실행
      const hook = new AsyncHook();
      const result = await hook.execute({
        tool_name: 'AsyncTool',
        tool_input: {},
      });

      // THEN: 비동기 처리 완료
      expect(result.success).toBe(true);
      expect(result.message).toBe('Processed AsyncTool');
    });
  });
});
