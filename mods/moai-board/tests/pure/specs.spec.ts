// Pure tests for hooks/specs.ts (SPEC list, read guards, chunker), run with `bun test`
// (developer-local evidence, spec.md G-13). Named *.spec.ts so the engine glob skips them.
import { expect, test } from 'bun:test'
import {
  MAX_CHUNKS,
  PAGE_SIZE,
  SPEC_FILES,
  chunkAll,
  chunkMarkdown,
  filterSpecs,
  isSpecId,
  pageOf,
  parseSpecList,
  readSpecFile,
  statusChips,
  type SpecIo,
} from '../../hooks/specs'

// ---- SPEC list (shaped like spec.md M-14) ------------------------------------------------------

const row = (id: string, status: string) => `${id.padEnd(30)} ${status.padEnd(15)} 2026-10-02 17:28`

const specListText = () => {
  const lines = ['SPEC-ID                        Status          Modified', '-'.repeat(80)]
  let n = 1
  const add = (status: string, count: number) => {
    for (let i = 0; i < count; i++, n++) lines.push(row(`SPEC-X${n}-001`, status))
  }
  add('draft', 25)
  add('in-progress', 15)
  add('completed', 970)
  lines.push(row('SPEC-GITHUB-WORKFLOW', 'implemented'))
  lines.push(row('SPEC-I18N-001-ARCHIVED', 'archived'))
  lines.push(row('_archive', 'unknown'))
  return lines.join('\n') + '\n'
}

test('specs: list rows, header rule and non-ids dropped', () => {
  const rows = parseSpecList(specListText())
  expect(rows.length).toBe(1010)
  const ids = rows.map(r => r.id)
  for (const bad of ['SPEC-GITHUB-WORKFLOW', 'SPEC-I18N-001-ARCHIVED', '_archive', 'SPEC-ID']) expect(ids).not.toContain(bad)
  expect(rows[0]).toEqual({ id: 'SPEC-X1-001', status: 'draft' })
  expect(isSpecId('SPEC-GITHUB-WORKFLOW')).toBe(false)
  expect(isSpecId('SPEC-I18N-001-ARCHIVED')).toBe(false)
  expect(isSpecId('SPEC-MOAI-BOARD-MOD-001')).toBe(true)
})

test('specs: default filter', () => {
  const rows = parseSpecList(specListText())
  expect(filterSpecs(rows, 'active').length).toBe(40)
  expect(filterSpecs(rows, 'completed').length).toBe(970)
  expect(filterSpecs(rows, 'archived').length).toBe(0)
  expect(statusChips(rows)).toEqual(['active', 'completed'])
  expect(statusChips([...rows, { id: 'SPEC-Z-001', status: 'rejected' }, { id: 'SPEC-Y-001', status: 'planned' }])).toEqual([
    'active', 'planned', 'completed', 'rejected',
  ])
})

test('specs: pages of 15', () => {
  const active = filterSpecs(parseSpecList(specListText()), 'active')
  expect(PAGE_SIZE).toBe(15)
  const pages = [0, 1, 2].map(p => pageOf(active, p))
  expect(pages.map(p => p.pages)).toEqual([3, 3, 3])
  expect(pages.map(p => p.rows.length)).toEqual([15, 15, 10])
  expect(pageOf(active, 99).page).toBe(2)
  expect(pageOf(active, -3).page).toBe(0)
  expect(pageOf([], 0)).toEqual({ rows: [], page: 0, pages: 1 })
})

// ---- read guards -----------------------------------------------------------------------------------

const BASE = '/work/.moai/specs'
const REAL_BASE = '/real/work/.moai/specs'

const fakeIo = (stats: Record<string, { kind: 'file' | 'dir' | 'other'; realPath?: string }>) => {
  const statCalls: string[] = []
  const readCalls: string[] = []
  const io: SpecIo = {
    stat: async path => {
      statCalls.push(path)
      const found = stats[path]
      if (found === undefined) throw new Error('ENOENT')
      return found
    },
    read: async path => {
      readCalls.push(path)
      return 'TEXT'
    },
  }
  return { io, statCalls, readCalls }
}

const goodStats = (id = 'SPEC-MOAI-BOARD-MOD-001', file = 'spec.md') => ({
  [BASE]: { kind: 'dir' as const, realPath: REAL_BASE },
  [`${BASE}/${id}/${file}`]: { kind: 'file' as const, realPath: `${REAL_BASE}/${id}/${file}` },
})

test('specs-guard: id pattern', async () => {
  for (const id of ['../x', '/etc/passwd', 'spec-lower-001', 'SPEC-A-001/../../x', 'SPEC-A-1', '']) {
    const { io, statCalls, readCalls } = fakeIo(goodStats())
    const r = await readSpecFile(io, '/work', id, 'spec.md')
    expect(r.ok).toBe(false)
    expect(statCalls.length).toBe(0)
    expect(readCalls.length).toBe(0)
  }
  const { io, readCalls } = fakeIo(goodStats())
  expect(await readSpecFile(io, '/work', 'SPEC-MOAI-BOARD-MOD-001', 'spec.md')).toEqual({
    ok: true,
    text: 'TEXT',
    realPath: `${REAL_BASE}/SPEC-MOAI-BOARD-MOD-001/spec.md`,
  })
  // the read target is the resolved real path, never the spelling that was asked for
  expect(readCalls).toEqual([`${REAL_BASE}/SPEC-MOAI-BOARD-MOD-001/spec.md`])
})

test('specs-guard: file allow-list', async () => {
  expect([...SPEC_FILES]).toEqual(['spec.md', 'plan.md', 'acceptance.md', 'design.md', 'research.md', 'progress.md'])
  for (const file of SPEC_FILES) {
    const { io } = fakeIo(goodStats('SPEC-A-001', file))
    expect((await readSpecFile(io, '/work', 'SPEC-A-001', file)).ok).toBe(true)
  }
  for (const file of ['../spec.md', 'spec.md.bak', 'tasks.md', 'SPEC.md', 'a/b.md', '', '/etc/passwd']) {
    const { io, statCalls, readCalls } = fakeIo(goodStats('SPEC-A-001', file))
    expect((await readSpecFile(io, '/work', 'SPEC-A-001', file)).ok).toBe(false)
    expect(statCalls.length + readCalls.length).toBe(0)
  }
})

test('specs-guard: real path outside base denied', async () => {
  const id = 'SPEC-A-001'
  const outside = [
    '/elsewhere/spec.md', // plainly outside
    `${REAL_BASE}-other/${id}/spec.md`, // shares the base as a string prefix, not as a directory
    `${REAL_BASE}`, // the base itself is not a file below it
  ]
  for (const realPath of outside) {
    const { io, readCalls } = fakeIo({
      [BASE]: { kind: 'dir', realPath: REAL_BASE },
      [`${BASE}/${id}/spec.md`]: { kind: 'file', realPath },
    })
    const r = await readSpecFile(io, '/work', id, 'spec.md')
    expect(r.ok).toBe(false)
    expect(readCalls.length).toBe(0)
  }
})

test('specs-guard: nothing read before both stats', async () => {
  const id = 'SPEC-A-001'
  const cases: Record<string, { kind: 'file' | 'dir' | 'other'; realPath?: string }>[] = [
    { [`${BASE}/${id}/spec.md`]: { kind: 'file', realPath: `${REAL_BASE}/${id}/spec.md` } }, // base stat missing
    { [BASE]: { kind: 'dir', realPath: REAL_BASE } }, // file stat missing
    { [BASE]: { kind: 'dir' }, [`${BASE}/${id}/spec.md`]: { kind: 'file', realPath: `${REAL_BASE}/${id}/spec.md` } }, // base unresolved
    { [BASE]: { kind: 'dir', realPath: REAL_BASE }, [`${BASE}/${id}/spec.md`]: { kind: 'file' } }, // file unresolved
    { [BASE]: { kind: 'dir', realPath: REAL_BASE }, [`${BASE}/${id}/spec.md`]: { kind: 'dir', realPath: `${REAL_BASE}/${id}/spec.md` } }, // not a file
  ]
  for (const stats of cases) {
    const { io, readCalls } = fakeIo(stats)
    expect((await readSpecFile(io, '/work', id, 'spec.md')).ok).toBe(false)
    expect(readCalls.length).toBe(0)
  }
})

// ---- chunker -----------------------------------------------------------------------------------------

const hasLoneSurrogate = (s: string) => /[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/.test(s)

const para = (n: number, size: number) => `P${n} `.padEnd(size, 'x')

// A document over 133,000 characters with fenced code (blank lines inside), one fenced block
// larger than a chunk, one very long line carrying surrogate pairs, and plain paragraphs.
const bigDoc = () => {
  const out: string[] = []
  for (let i = 0; i < 40; i++) {
    out.push(para(i, 900))
    out.push('```ts\n' + Array.from({ length: 30 }, (_, k) => `const v${i}_${k} = ${k}\n\nconst w${i}_${k} = ${k}`).join('\n') + '\n```')
  }
  out.push('```json\n' + Array.from({ length: 1500 }, (_, k) => `  "key${k}": "value ${k}",`).join('\n') + '\n```')
  out.push('a' + '\u{1F600}'.repeat(12_000))
  out.push(para(99, 20_000))
  while (out.join('\n\n').length < 133_596) out.push(para(100 + out.length, 1000))
  return out.join('\n\n')
}

test('specs-chunk: each chunk at most 9000', () => {
  const doc = bigDoc()
  expect(doc.length).toBeGreaterThanOrEqual(133_596)
  const all = chunkAll(doc)
  expect(all.length).toBeGreaterThan(MAX_CHUNKS)
  for (const chunk of all) {
    expect(chunk.length).toBeLessThanOrEqual(9000)
    expect(hasLoneSurrogate(chunk)).toBe(false)
  }
  expect(chunkAll('').length).toBe(0)
  expect(chunkAll('short').length).toBe(1)
})

test('specs-chunk: splits at blank lines', () => {
  const blocks = [para(1, 3000), para(2, 3000), para(3, 3000)]
  const text = blocks.join('\n\n')
  const chunks = chunkAll(text)
  expect(chunks).toEqual([blocks[0] + '\n\n' + blocks[1], blocks[2]])
  expect(chunks.join('\n\n')).toBe(text)
  // a blank line inside a fence is not a boundary
  const fenced = '```\nline one\n\nline two\n```'
  expect(chunkAll(`${para(4, 100)}\n\n${fenced}`)).toEqual([`${para(4, 100)}\n\n${fenced}`])
})

test('specs-chunk: fences closed and reopened', () => {
  const body = Array.from({ length: 900 }, (_, k) => `    row ${k} of a block larger than a chunk`)
  const text = '```ts\n' + body.join('\n') + '\n```'
  expect(text.length).toBeGreaterThan(9000)
  const chunks = chunkAll(text)
  expect(chunks.length).toBeGreaterThan(2)
  for (const [i, chunk] of chunks.entries()) {
    const fences = chunk.split('\n').filter(l => /^```/.test(l))
    expect(fences.length % 2).toBe(0) // balanced inside every chunk
    if (i > 0) expect(chunk.startsWith('```ts\n')).toBe(true) // reopened with the same info string
    expect(chunk.length).toBeLessThanOrEqual(9000)
  }
  // nothing is lost: dropping the added fence lines gives the original body back
  const rebuilt = chunks.flatMap(c => c.split('\n')).filter(l => !/^```/.test(l))
  expect(rebuilt).toEqual(body)
})

test('specs-chunk: 12-chunk cap with notice', () => {
  const doc = bigDoc()
  const { chunks, notice } = chunkMarkdown(doc, 'SPEC-A-001/spec.md')
  expect(chunks.length).toBe(MAX_CHUNKS)
  expect(notice).toContain('SPEC-A-001/spec.md')
  expect(notice).toContain('12')
  for (const chunk of chunks) expect(chunk.length).toBeLessThanOrEqual(9000)
  const small = chunkMarkdown('hello', 'SPEC-A-001/plan.md')
  expect(small).toEqual({ chunks: ['hello'], notice: '' })
})
