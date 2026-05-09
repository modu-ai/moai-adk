import assert from 'node:assert/strict';
import {
  formatResetDurationForTest,
} from './src/codex-quota.ts';
import {
  composeMoaiNativeFooterLinesForTest,
  normalizeClaudeStatusLinesForTest,
} from './src/statusline.ts';

const raw = [
  'MoAI v0.0.0 │ gpt-5.5',
  'CW: 🔋 █████░░░░░ 45% │ 5H: 🔋 ░░░░░░░░░░ 0% │ 7D: 🔋 ░░░░░░░░░░ 0%',
  '~/repo (main)',
].join('\n');

const normalized = normalizeClaudeStatusLinesForTest(raw);
assert.deepEqual(normalized, [
  'MoAI v0.0.0 │ gpt-5.5',
  'CW: 🔋 █████░░░░░ 45%',
  '~/repo (main)',
]);

const nativeQuota = '5H: 🔋 █████████░ 9% (2h 15m) │ 7D: 🔋 ███████░░░ 28% (2d 16h 7m)';
const lines = composeMoaiNativeFooterLinesForTest(normalized, nativeQuota, 220);
const barLine = lines.find((line) => line.includes('CW:')) ?? '';
const linesWithoutClaudeCw = composeMoaiNativeFooterLinesForTest(
  ['MoAI v0.0.0 │ gpt-5.5', '~/repo (main)'],
  nativeQuota,
  220,
  'CW: 🔋 ░░░░░░░░░░ 0%',
);
const synthesizedBarLine = linesWithoutClaudeCw.find((line) => line.includes('CW:')) ?? '';

assert(barLine.includes('CW: 🔋 █████░░░░░ 45%'), 'context window state should remain visible');
assert(synthesizedBarLine.includes('CW: 🔋 ░░░░░░░░░░ 0%'), 'context window state should be synthesized when moai statusline omits CW');
assert(synthesizedBarLine.includes('5H: 🔋'), 'quota should share the synthesized CW line');
assert(barLine.includes('5H: 🔋'), '5H quota should use MoAI native 5H segment style');
assert(barLine.includes('7D: 🔋'), '7D quota should use MoAI native 7D segment style');
assert(barLine.includes('(2h 15m)'), '5H reset should use compact relative format');
assert(barLine.includes('(2d 16h 7m)'), '7D reset should use compact day/time format');
assert(!lines.some((line) => line.includes('Codex')), 'footer must not use @kmiyh/pi-codex-plan-limits Codex prefix style');
assert(!lines.some((line) => /5H:.*0%|7D:.*0%/.test(line)), 'fallback 0% quota from moai statusline should be removed');
assert(!lines.some((line) => /Resets|reset/.test(line)), 'reset labels should stay compact');

const now = Date.UTC(2026, 0, 1, 0, 0, 0);
assert.equal(formatResetDurationForTest(now + (3 * 60 + 11) * 60_000, now), '3h 11m');
assert.equal(formatResetDurationForTest(now + (2 * 1_440 + 16 * 60 + 7) * 60_000, now), '2d 16h 7m');

console.log('statusline footer regression ok');
