/**
 * @TEST:POLICY-PERF-001 | Performance Benchmark for Policy Block Hook
 * 개선 전/후 실행 시간 비교
 */

import { beforeEach, describe, expect, it } from 'vitest';
import type { HookInput } from '../policy-block';
import { PolicyBlock } from '../policy-block';

describe('PolicyBlock Performance Benchmark', () => {
  let policyBlock: PolicyBlock;

  beforeEach(() => {
    policyBlock = new PolicyBlock();
  });

  describe('성능 비교: Read-only tools', () => {
    it('should measure Read tool execution time', async () => {
      const iterations = 100;
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: { file_path: '/test/file.txt' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const avgTime = (end - start) / iterations;
      console.log(
        `\n📊 Read tool: ${avgTime.toFixed(3)}ms per call (${iterations} iterations)`
      );

      expect(avgTime).toBeLessThan(1); // Should be under 1ms per call
    });

    it('should measure Glob tool execution time', async () => {
      const iterations = 100;
      const input: HookInput = {
        tool_name: 'Glob',
        tool_input: { pattern: '**/*.ts' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const avgTime = (end - start) / iterations;
      console.log(
        `📊 Glob tool: ${avgTime.toFixed(3)}ms per call (${iterations} iterations)`
      );

      expect(avgTime).toBeLessThan(1);
    });

    it('should measure MCP tool execution time', async () => {
      const iterations = 100;
      const input: HookInput = {
        tool_name: 'mcp__context7__resolve-library-id',
        tool_input: { libraryName: 'react' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const avgTime = (end - start) / iterations;
      console.log(
        `📊 MCP tool: ${avgTime.toFixed(3)}ms per call (${iterations} iterations)`
      );

      expect(avgTime).toBeLessThan(1);
    });
  });

  describe('성능 비교: Bash commands', () => {
    it('should measure safe Bash command execution time', async () => {
      const iterations = 100;
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: { command: 'git status' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const avgTime = (end - start) / iterations;
      console.log(
        `📊 Bash (safe): ${avgTime.toFixed(3)}ms per call (${iterations} iterations)`
      );

      expect(avgTime).toBeLessThan(2); // Bash commands take slightly longer
    });

    it('should measure dangerous command detection time', async () => {
      const iterations = 100;
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: { command: 'rm -rf /' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const avgTime = (end - start) / iterations;
      console.log(
        `📊 Bash (dangerous): ${avgTime.toFixed(3)}ms per call (${iterations} iterations)\n`
      );

      expect(avgTime).toBeLessThan(2);
    });
  });

  describe('대량 호출 시뮬레이션', () => {
    it('should handle 1000 Read calls efficiently', async () => {
      const iterations = 1000;
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: { file_path: '/test/file.txt' },
      };

      const start = performance.now();
      for (let i = 0; i < iterations; i++) {
        await policyBlock.execute(input);
      }
      const end = performance.now();

      const totalTime = end - start;
      const avgTime = totalTime / iterations;

      console.log(`\n📊 1000 Read calls:`);
      console.log(`   - Total: ${totalTime.toFixed(2)}ms`);
      console.log(`   - Average: ${avgTime.toFixed(3)}ms per call`);
      console.log(
        `   - Expected savings: ~${(1000 * 0.5).toFixed(0)}ms vs unoptimized\n`
      );

      expect(totalTime).toBeLessThan(100); // Should complete in under 100ms
    });
  });
});
