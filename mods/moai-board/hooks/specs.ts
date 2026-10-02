// moai-board SPEC layer: `$`-free. The list parser, the id and file allow-lists, the
// real-path guard (it takes its stat and read as an injected `io`, so bun can test it)
// and the Markdown chunker. register.tsx wires `io` to the engine's file calls.
import type { MoaiBoardFeed, MoaiBoardSpecRow, MoaiBoardSpecs } from '../types'
import { ARGV, SRC_SPECS, describeFailure, feedFail, feedOk, runOutcome, type Run } from './data'

// ---- ids and files (REQ-MBM-011, REQ-MBM-012) ----------------------------------------------------
/** The SPEC-id rule (spec.md §3), stated once and used by the list and by the reader. */
const SPEC_ID_RE = /^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$/

export const SPEC_FILES = ['spec.md', 'plan.md', 'acceptance.md', 'design.md', 'research.md', 'progress.md'] as const

export const isSpecId = (id: string): boolean => SPEC_ID_RE.test(id)

export const isSpecFile = (file: string): boolean => (SPEC_FILES as readonly string[]).includes(file)

const STATUS_RE = /^[a-z][a-z-]*$/

/** `moai spec status --list`: only rows whose first column is a SPEC id and whose second is a status word. */
export const parseSpecList = (text: string): MoaiBoardSpecRow[] => {
  const rows: MoaiBoardSpecRow[] = []
  for (const line of text.split('\n')) {
    const [id = '', status = ''] = line.trim().split(/\s+/)
    if (isSpecId(id) && STATUS_RE.test(status)) rows.push({ id, status })
  }
  return rows
}

/** One SPEC list poll (on tab open and on refresh only). */
export const readSpecList = async (
  run: Run,
  prev: MoaiBoardFeed<MoaiBoardSpecs>,
  now: number,
): Promise<MoaiBoardFeed<MoaiBoardSpecs>> => {
  const out = await runOutcome(run, ARGV.specList)
  return out.kind === 'ok' ? feedOk({ rows: parseSpecList(out.stdout) }, now) : feedFail(prev, describeFailure(SRC_SPECS, out))
}

// ---- filter and pages ---------------------------------------------------------------------------------
export const PAGE_SIZE = 15
const ACTIVE_STATUSES = ['draft', 'in-progress']
const STATUS_ORDER = ['planned', 'implemented', 'completed', 'superseded', 'archived', 'rejected', 'unknown']

/** 'active' is draft plus in-progress (the default); any other value is one status word. */
export const filterSpecs = (rows: readonly MoaiBoardSpecRow[], status: string): MoaiBoardSpecRow[] =>
  rows.filter(r => (status === 'active' ? ACTIVE_STATUSES.includes(r.status) : r.status === status))

/** The filter chips: `active` first, then every other status present, in lifecycle order. */
export const statusChips = (rows: readonly MoaiBoardSpecRow[]): string[] => {
  const present = [...new Set(rows.map(r => r.status))].filter(s => !ACTIVE_STATUSES.includes(s))
  const rank = (s: string) => (STATUS_ORDER.includes(s) ? STATUS_ORDER.indexOf(s) : STATUS_ORDER.length)
  return ['active', ...present.sort((a, b) => rank(a) - rank(b) || a.localeCompare(b))]
}

export const pageOf = <T>(rows: readonly T[], page: number): { rows: T[]; page: number; pages: number } => {
  const pages = Math.max(1, Math.ceil(rows.length / PAGE_SIZE))
  const at = Math.min(Math.max(0, Math.trunc(page) || 0), pages - 1)
  return { rows: rows.slice(at * PAGE_SIZE, (at + 1) * PAGE_SIZE), page: at, pages }
}

// ---- reading one SPEC file (REQ-MBM-012) -----------------------------------------------------------------
export type SpecIo = {
  stat: (path: string) => Promise<{ kind: 'file' | 'dir' | 'other'; realPath?: string | undefined }>
  read: (path: string) => Promise<string>
}

export type SpecRead = { ok: true; text: string; realPath: string } | { ok: false; reason: string }

const refuse = (reason: string): SpecRead => ({ ok: false, reason })

/**
 * Reads `<root>/.moai/specs/<id>/<file>` only when the id and the file name pass their allow-lists and the
 * file's resolved real path lies below the resolved real path of `<root>/.moai/specs`. Both paths are
 * resolved before anything is read, and what is read is the real path, never the spelling asked for.
 */
export const readSpecFile = async (io: SpecIo, root: string, id: string, file: string): Promise<SpecRead> => {
  if (!isSpecId(id)) return refuse(`"${id}" is not a SPEC id, so nothing was read.`)
  if (!isSpecFile(file)) return refuse(`"${file}" is not one of the SPEC files, so nothing was read.`)
  const base = `${root.replace(/[\\/]+$/, '')}/.moai/specs`
  let baseReal: string | undefined
  let fileStat: Awaited<ReturnType<SpecIo['stat']>>
  try {
    baseReal = (await io.stat(base)).realPath
    fileStat = await io.stat(`${base}/${id}/${file}`)
  } catch {
    return refuse(`${id}/${file} was not found under ${base}.`)
  }
  const fileReal = fileStat.realPath
  if (baseReal === undefined || fileReal === undefined || fileStat.kind !== 'file') return refuse(`${id}/${file} could not be resolved to a file.`)
  const sep = baseReal.includes('\\') && !baseReal.includes('/') ? '\\' : '/'
  const below = `${baseReal.replace(/[\\/]+$/, '')}${sep}`
  if (!fileReal.startsWith(below)) return refuse(`${id}/${file} resolves outside the SPEC folder, so it was not read.`)
  try {
    return { ok: true, text: await io.read(fileReal), realPath: fileReal }
  } catch (err) {
    return refuse(`${id}/${file} could not be read: ${err instanceof Error ? err.message : String(err)}`)
  }
}

// ---- Markdown chunker (REQ-MBM-012) -------------------------------------------------------------------------
export const MAX_CHUNK = 9000
export const MAX_CHUNKS = 12

const FENCE_RE = /^ {0,3}(`{3,}|~{3,})(.*)$/

/** Cuts a line to parts of at most `limit` UTF-16 units, never between the halves of a surrogate pair. */
const cutLine = (line: string, limit: number): string[] => {
  const parts: string[] = []
  let rest = line
  while (rest.length > limit) {
    const last = rest.charCodeAt(limit - 1)
    const at = last >= 0xd800 && last <= 0xdbff ? limit - 1 : limit
    parts.push(rest.slice(0, at))
    rest = rest.slice(at)
  }
  parts.push(rest)
  return parts
}

/** Blocks split at blank lines outside fenced code. */
const splitBlocks = (text: string): string[] => {
  const blocks: string[] = []
  let cur: string[] = []
  let marker = ''
  const flush = () => {
    if (cur.length > 0) blocks.push(cur.join('\n'))
    cur = []
  }
  for (const line of text.replace(/\r\n/g, '\n').split('\n')) {
    const fence = FENCE_RE.exec(line)
    if (marker === '') {
      if (line.trim() === '') {
        flush()
        continue
      }
      if (fence !== null) marker = fence[1] ?? ''
    } else if (fence !== null && (fence[1] ?? '')[0] === marker[0] && (fence[1] ?? '').length >= marker.length && (fence[2] ?? '').trim() === '') {
      marker = ''
    }
    cur.push(line)
  }
  flush()
  return blocks
}

const LINE_LIMIT = MAX_CHUNK - 200

/** One block larger than a chunk, split at line boundaries; a fence it is inside is closed and reopened. */
const splitBlock = (block: string, max: number): string[] => {
  const pieces: string[] = []
  let lines: string[] = []
  let length = 0
  let opener = '' // the opening fence line to repeat, '' outside a fence
  let marker = ''
  const reopened = () => (opener === '' ? [] : [opener])
  const lengthOf = (ls: readonly string[]) => ls.reduce((n, l) => n + l.length, 0) + Math.max(0, ls.length - 1)
  lines = reopened()
  length = lengthOf(lines)
  for (const raw of block.split('\n')) {
    const parts = cutLine(raw, LINE_LIMIT)
    const fence = parts.length === 1 ? FENCE_RE.exec(raw) : null
    for (const part of parts) {
      const room = opener === '' ? max : max - 1 - marker.length
      if (lines.length > reopened().length && (lines.length === 0 ? 0 : length + 1) + part.length > room) {
        pieces.push((opener === '' ? lines : [...lines, marker]).join('\n'))
        lines = reopened()
        length = lengthOf(lines)
      }
      length += (lines.length === 0 ? 0 : 1) + part.length
      lines.push(part)
    }
    if (fence !== null) {
      const m = fence[1] ?? ''
      if (opener === '') {
        marker = m
        opener = `${m}${(fence[2] ?? '').slice(0, 80)}`
      } else if (m[0] === marker[0] && m.length >= marker.length && (fence[2] ?? '').trim() === '') {
        opener = ''
        marker = ''
      }
    }
  }
  pieces.push(lines.join('\n'))
  return pieces
}

/** Every chunk of the text: at most `max` characters, split at blank lines, fenced code kept whole where it fits. */
export const chunkAll = (text: string, max = MAX_CHUNK): string[] => {
  const chunks: string[] = []
  let cur = ''
  for (const block of splitBlocks(text)) {
    if (block.length > max) {
      if (cur !== '') chunks.push(cur)
      cur = ''
      chunks.push(...splitBlock(block, max))
    } else if (cur === '') {
      cur = block
    } else if (cur.length + 2 + block.length <= max) {
      cur += `\n\n${block}`
    } else {
      chunks.push(cur)
      cur = block
    }
  }
  if (cur !== '') chunks.push(cur)
  return chunks
}

/** At most 12 chunks of at most 9,000 characters, then a notice that names the file. */
export const chunkMarkdown = (text: string, file: string): { chunks: string[]; notice: string } => {
  const all = chunkAll(text)
  return all.length <= MAX_CHUNKS
    ? { chunks: all, notice: '' }
    : {
        chunks: all.slice(0, MAX_CHUNKS),
        notice: `Showing the first ${MAX_CHUNKS} of ${all.length} parts of ${file}; the rest is not shown.`,
      }
}
