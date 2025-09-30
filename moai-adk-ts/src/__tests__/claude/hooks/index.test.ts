// @TEST:HOOK-001 | 
// Related: @CODE:HOOK-001

/**
 * @file Hook system integration tests
 * @author MoAI Team
 */

import { afterEach, beforeEach, describe, expect, vi } from 'vitest';

import {
  HookSystem,
  outputResult,
  parseClaudeInput,
} from '../../../claude/hooks/index';
import { PolicyBlock } from '../../../claude/hooks/security/policy-block';
import { SteeringGuard } from '../../../claude/hooks/security/steering-guard';
import type {
  HookInput,
  HookResult,
  MoAIHook,
} from '../../../claude/hooks/types';

// Mock console and process
const mockConsoleError = vi
  .spyOn(console, 'error')
  .mockImplementation(() => {});
const mockConsoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});
const mockProcessExit = vi.spyOn(process, 'exit').mockImplementation(() => {
  throw new Error('process.exit called');
});

describe('HookSystem', () => {
  let hookSystem: HookSystem;

  beforeEach(() => {
    hookSystem = new HookSystem();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('constructor', () => {
    it('should initialize with default config', () => {
      const system = new HookSystem();
      expect(system.getRegisteredHooks()).toEqual([]);
    });

    it('should accept custom config', () => {
      const customConfig = {
        enabled: false,
        timeout: 5000,
        disabledHooks: ['test-hook'],
      };

      const system = new HookSystem(customConfig);
      expect(system.getRegisteredHooks()).toEqual([]);
    });
  });

  describe('hook registration', () => {
    it('should register hooks successfully', () => {
      const mockHook: MoAIHook = {
        name: 'test-hook',
        execute: vi.fn().mockResolvedValue({ success: true }),
      };

      hookSystem.registerHook(mockHook);

      expect(hookSystem.hasHook('test-hook')).toBe(true);
      expect(hookSystem.getRegisteredHooks()).toContain('test-hook');
    });

    it('should not register disabled hooks', () => {
      const system = new HookSystem({
        disabledHooks: ['disabled-hook'],
      });

      const mockHook: MoAIHook = {
        name: 'disabled-hook',
        execute: vi.fn(),
      };

      system.registerHook(mockHook);

      expect(system.hasHook('disabled-hook')).toBe(false);
    });

    it('should register actual hook implementations', () => {
      const steeringGuard = new SteeringGuard();
      const policyBlock = new PolicyBlock();

      hookSystem.registerHook(steeringGuard);
      hookSystem.registerHook(policyBlock);

      expect(hookSystem.hasHook('steering-guard')).toBe(true);
      expect(hookSystem.hasHook('policy-block')).toBe(true);
    });
  });

  describe('hook execution', () => {
    it('should execute single hook successfully', async () => {
      const mockHook: MoAIHook = {
        name: 'test-hook',
        execute: vi.fn().mockResolvedValue({
          success: true,
          message: 'Hook executed',
        }),
      };

      hookSystem.registerHook(mockHook);

      const input: HookInput = { prompt: 'test prompt' };
      const result = await hookSystem.executeHook('test-hook', input);

      expect(result.success).toBe(true);
      expect(result.message).toBe('Hook executed');
      expect(mockHook.execute).toHaveBeenCalledWith(input);
    });

    it('should handle hook execution errors', async () => {
      const mockHook: MoAIHook = {
        name: 'failing-hook',
        execute: vi.fn().mockRejectedValue(new Error('Hook failed')),
      };

      hookSystem.registerHook(mockHook);

      const result = await hookSystem.executeHook('failing-hook', {});

      expect(result.success).toBe(false);
      expect(result.message).toContain('Hook failing-hook failed: Hook failed');
      expect(result.exitCode).toBe(1);
    });

    it('should handle non-existent hooks', async () => {
      const result = await hookSystem.executeHook('non-existent', {});

      expect(result.success).toBe(false);
      expect(result.message).toContain('Hook non-existent not found');
    });

    it('should respect timeout settings', async () => {
      const system = new HookSystem({ timeout: 100 });

      const slowHook: MoAIHook = {
        name: 'slow-hook',
        execute: vi
          .fn()
          .mockImplementation(
            () => new Promise(resolve => setTimeout(resolve, 200))
          ),
      };

      system.registerHook(slowHook);

      const result = await system.executeHook('slow-hook', {});

      expect(result.success).toBe(false);
      expect(result.message).toContain('timed out');
    });

    it('should return early when hooks disabled', async () => {
      const system = new HookSystem({ enabled: false });

      const result = await system.executeHook('any-hook', {});

      expect(result.success).toBe(true);
      expect(result.message).toBe('Hooks disabled');
    });
  });

  describe('executing all hooks', () => {
    it('should execute all registered hooks', async () => {
      const hook1: MoAIHook = {
        name: 'hook1',
        execute: vi
          .fn()
          .mockResolvedValue({ success: true, message: 'Hook 1' }),
      };

      const hook2: MoAIHook = {
        name: 'hook2',
        execute: vi
          .fn()
          .mockResolvedValue({ success: true, message: 'Hook 2' }),
      };

      hookSystem.registerHook(hook1);
      hookSystem.registerHook(hook2);

      const results = await hookSystem.executeAllHooks({ prompt: 'test' });

      expect(results).toHaveLength(2);
      expect(results[0].message).toBe('Hook 1');
      expect(results[1].message).toBe('Hook 2');
    });

    it('should stop execution when a hook blocks', async () => {
      const blockingHook: MoAIHook = {
        name: 'blocking-hook',
        execute: vi.fn().mockResolvedValue({
          success: false,
          blocked: true,
          message: 'Blocked',
        }),
      };

      const secondHook: MoAIHook = {
        name: 'second-hook',
        execute: vi.fn().mockResolvedValue({ success: true }),
      };

      hookSystem.registerHook(blockingHook);
      hookSystem.registerHook(secondHook);

      const results = await hookSystem.executeAllHooks({});

      expect(results).toHaveLength(1);
      expect(results[0].blocked).toBe(true);
      expect(secondHook.execute).not.toHaveBeenCalled();
    });

    it('should continue execution when hooks fail but dont block', async () => {
      const failingHook: MoAIHook = {
        name: 'failing-hook',
        execute: vi.fn().mockRejectedValue(new Error('Failed')),
      };

      const secondHook: MoAIHook = {
        name: 'second-hook',
        execute: vi.fn().mockResolvedValue({ success: true }),
      };

      hookSystem.registerHook(failingHook);
      hookSystem.registerHook(secondHook);

      const results = await hookSystem.executeAllHooks({});

      expect(results).toHaveLength(2);
      expect(results[0].success).toBe(false);
      expect(results[1].success).toBe(true);
    });
  });

  describe('config loading', () => {
    it('should load config from file', async () => {
      const mockConfig = {
        enabled: true,
        timeout: 15000,
        disabledHooks: ['test-hook'],
        security: {
          allowedCommands: ['git', 'npm'],
          blockedPatterns: ['rm -rf'],
          requireApproval: ['--force'],
        },
      };

      // Mock fs module
      const _fs = require('node:fs');
      vi.doMock('fs', () => ({
        existsSync: vi.fn().mockReturnValue(true),
        readFileSync: vi.fn().mockReturnValue(JSON.stringify(mockConfig)),
      }));

      const config = await HookSystem.loadConfig('/test/project');

      expect(config.timeout).toBe(15000);
      expect(config.disabledHooks).toContain('test-hook');
    });

    it('should return default config when file loading fails', async () => {
      vi.doMock('fs', () => ({
        existsSync: vi.fn().mockReturnValue(false),
      }));

      const config = await HookSystem.loadConfig('/test/project');

      expect(config.enabled).toBe(true);
      expect(config.timeout).toBe(10000);
      expect(config.security.allowedCommands).toContain('git');
    });
  });
});

describe('parseClaudeInput', () => {
  const originalStdin = process.stdin;

  afterEach(() => {
    process.stdin = originalStdin;
  });

  it('should parse valid JSON input', async () => {
    const mockStdin = {
      on: vi.fn(),
      setEncoding: vi.fn(),
    };

    process.stdin = mockStdin as any;

    const inputData = { prompt: 'test prompt', tool_name: 'Bash' };

    // Simulate stdin events
    setTimeout(() => {
      const dataCallback = mockStdin.on.mock.calls.find(
        call => call[0] === 'data'
      )?.[1];
      const endCallback = mockStdin.on.mock.calls.find(
        call => call[0] === 'end'
      )?.[1];

      if (dataCallback) dataCallback(JSON.stringify(inputData));
      if (endCallback) endCallback();
    }, 0);

    const result = await parseClaudeInput();

    expect(result).toEqual(inputData);
  });

  it('should handle empty input', async () => {
    const mockStdin = {
      on: vi.fn(),
      setEncoding: vi.fn(),
    };

    process.stdin = mockStdin as any;

    setTimeout(() => {
      const endCallback = mockStdin.on.mock.calls.find(
        call => call[0] === 'end'
      )[1];
      if (endCallback) endCallback();
    }, 0);

    const result = await parseClaudeInput();

    expect(result).toEqual({});
  });

  it('should reject invalid JSON', async () => {
    const mockStdin = {
      on: vi.fn(),
      setEncoding: vi.fn(),
    };

    process.stdin = mockStdin as any;

    setTimeout(() => {
      const dataCallback = mockStdin.on.mock.calls.find(
        call => call[0] === 'data'
      )?.[1];
      const endCallback = mockStdin.on.mock.calls.find(
        call => call[0] === 'end'
      )?.[1];

      if (dataCallback) dataCallback('invalid json');
      if (endCallback) endCallback();
    }, 0);

    await expect(parseClaudeInput()).rejects.toThrow('Invalid JSON input');
  });
});

describe('outputResult', () => {
  it('should exit with code 2 when blocked', () => {
    const result: HookResult = {
      success: false,
      blocked: true,
      message: 'Action blocked',
    };

    expect(() => outputResult(result)).toThrow('process.exit called');
    expect(mockConsoleError).toHaveBeenCalledWith('BLOCKED: Action blocked');
    expect(mockProcessExit).toHaveBeenCalledWith(2);
  });

  it('should exit with code 1 when failed', () => {
    const result: HookResult = {
      success: false,
      message: 'Hook failed',
      exitCode: 1,
    };

    expect(() => outputResult(result)).toThrow('process.exit called');
    expect(mockConsoleError).toHaveBeenCalledWith('ERROR: Hook failed');
    expect(mockProcessExit).toHaveBeenCalledWith(1);
  });

  it('should exit with code 0 when successful', () => {
    const result: HookResult = {
      success: true,
      message: 'Hook succeeded',
    };

    expect(() => outputResult(result)).toThrow('process.exit called');
    expect(mockConsoleLog).toHaveBeenCalledWith('Hook succeeded');
    expect(mockProcessExit).toHaveBeenCalledWith(0);
  });

  it('should exit with code 0 when successful without message', () => {
    const result: HookResult = {
      success: true,
    };

    expect(() => outputResult(result)).toThrow('process.exit called');
    expect(mockConsoleLog).not.toHaveBeenCalled();
    expect(mockProcessExit).toHaveBeenCalledWith(0);
  });
});
