// Pure tests for hooks/data.ts, run with `bun test` (developer-local evidence,
// spec.md G-11). Named *.spec.ts so the engine runner's *.test.ts glob skips them.
import { expect, test } from 'bun:test'
import { ARGV } from '../../hooks/data'

test('argv: table carries the three diagnostic commands only', () => {
  expect(Object.keys(ARGV).sort()).toEqual(['binary', 'mcp', 'memory'])
  expect([...ARGV.binary]).toEqual(['moai', 'doctor', '--check', 'Binary Freshness'])
  expect([...ARGV.mcp]).toEqual(['moai', 'doctor', '--check', 'MCP Server Version'])
  expect([...ARGV.memory]).toEqual(['moai', 'memory', 'doctor', '--json'])
})
