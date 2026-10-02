// PURE (bun) — the fail-soft guard shapes, read from the deployed source.
// The engine sandbox has no `process` global, so the escaped-rejection surface
// of a discarded timer promise is observed at the host level (the sync-audit's
// fault injection is the observed failure on record). This file pins the
// structure that keeps that surface closed: every catch that calls setNotice
// nests a guard, exactly as soft() does.
import { expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const source = (): string => readFileSync(join(import.meta.dir, '../../hooks/register.ts'), 'utf8')

test('guard: the health cycle catch wraps its notice write (F-1)', () => {
  const src = source()
  const start = src.indexOf('const runHealthCycle')
  const end = src.indexOf('const startHealth', start)
  expect(start).toBeGreaterThan(-1)
  expect(end).toBeGreaterThan(start)
  const body = src.slice(start, end)
  const catchAt = body.indexOf('} catch (err) {')
  const finallyAt = body.indexOf('} finally {')
  expect(catchAt).toBeGreaterThan(-1)
  expect(finallyAt).toBeGreaterThan(catchAt)
  // soft()'s pattern: the notice write inside the catch sits in a nested
  // try/catch — a state failure during the catch itself must not reject the
  // promise the timer discards with `void` (sync-audit F-1).
  const guard = body.slice(catchAt, finallyAt)
  expect(guard).toContain('try {')
  expect(guard).toContain('} catch {')
})

test('guard: every setNotice call site sits inside a nested guard or soft()', () => {
  const src = source()
  // setNotice is safe to await in exactly two places: inside soft()'s guarded
  // bodies, and inside a catch's nested try. A bare top-level `await setNotice(`
  // directly in a catch block (no nested try between) is the F-1 shape.
  const softGuarded = src.slice(src.indexOf('const soft ='), src.indexOf('// The health timer'))
  const guarded = src.slice(src.indexOf('const runHealthCycle'), src.indexOf('const startHealth'))
  for (const region of [softGuarded, guarded]) {
    const catches = [...region.matchAll(/\} catch(?: \([^)]*\))? \{/g)]
    for (const hit of catches) {
      const after = region.slice(hit.index + hit[0].length, region.indexOf('}', hit.index) + 1)
      if (after.includes('setNotice(')) expect(after).toContain('try {')
    }
  }
  expect(softGuarded).toContain('try {')
})
