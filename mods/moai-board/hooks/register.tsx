// moai-board hooks module. The only file that spells `$.…`: helper modules
// (data.ts, specs.ts, view.tsx) receive functions or resolved element tables,
// never `$` (the engine refuses `$` passed into an imported function, spec.md §4).
import type { EngineInterface, Register } from 'claude-code'
import type { MoaiBoardCard, MoaiBoardTab, MoaiBoardView } from '../types'
import {
  CANCEL_LABEL,
  CMD_TIMEOUT_MS,
  CONFIRM_LABEL,
  POLL_INTERVAL_MS,
  buildPickArgv,
  canPick,
  clampInterval,
  createFlightGate,
  createQueueReaders,
  cutChars,
  emptyFeed,
  executePick,
  firstLine,
  isConfirmed,
  isStillQueued,
  pickPrefix,
  readLanes,
  readQueue,
  sameFeed,
} from './data'
import { chunkMarkdown, isSpecFile, readSpecFile, readSpecList } from './specs'
import type { Actions, Model } from './view'
import { drawBoard } from './view'

const PANE = 'moai-board'

// Typed references: plugin and key are literals, used for nothing but $.state calls.
const viewRef = { plugin: 'moai-board', key: 'view' } as const
const queueRef = { plugin: 'moai-board', key: 'queue' } as const
const lanesRef = { plugin: 'moai-board', key: 'lanes' } as const
const specsRef = { plugin: 'moai-board', key: 'specs' } as const
const docRef = { plugin: 'moai-board', key: 'doc' } as const
const noticeRef = { plugin: 'moai-board', key: 'notice' } as const

const DEFAULT_VIEW: MoaiBoardView = { tab: 'queue', card: '', spec: '', file: '', status: 'active', page: 0, root: '' }

// The single process.run call site (REQ-MBM-013). Helpers receive it as a function.
const runMoai = ($: EngineInterface, argv: readonly string[]) =>
  $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })

const errorText = (err: unknown): string => (err instanceof Error ? err.message : String(err))

// ---- state helpers ---------------------------------------------------------------------
const readView = async ($: EngineInterface): Promise<MoaiBoardView> => (await $.state.get(viewRef)).value ?? DEFAULT_VIEW

const patchView = async ($: EngineInterface, patch: Partial<MoaiBoardView>): Promise<void> => {
  await $.state.set(viewRef, { ...(await readView($)), ...patch })
}

const say = async ($: EngineInterface, text: string): Promise<void> => {
  await $.state.set(noticeRef, text)
}

// Fail-soft (REQ-MBM-010): a handler that throws leaves a notice, never an exception in the hook chain.
const soft = async ($: EngineInterface, body: () => Promise<void>): Promise<void> => {
  try {
    await body()
  } catch (err) {
    try {
      await say($, `moai-board: ${errorText(err)}`)
    } catch {
      // nothing left to try; the session continues unaffected
    }
  }
}

// ---- polling (REQ-MBM-006, REQ-MBM-007) --------------------------------------------
// Module variables hold only what a hot reload may lose at the cost of one re-parse
// or one timer restart; view and data state live in $.state (REQ-MBM-015).
const readers = createQueueReaders()
const gate = createFlightGate()
let timer: { cancel: () => void } | undefined

type Wanted = { lanes: boolean; spec: boolean }

let pending: Wanted | undefined

const merge = (a: Wanted | undefined, b: Wanted): Wanted => ({
  lanes: (a?.lanes ?? false) || b.lanes,
  spec: (a?.spec ?? false) || b.spec,
})

const pollOnce = async ($: EngineInterface, want: Wanted): Promise<void> => {
  const run = (argv: readonly string[]) => runMoai($, argv)
  const now = Date.now()
  const view = await readView($)
  const prevQueue = (await $.state.get(queueRef)).value ?? emptyFeed()
  const q = await readQueue(run, readers, prevQueue, now)
  if (q.isChanged) await $.state.set(queueRef, q.feed)
  if (want.lanes || view.tab === 'lanes') {
    const prevLanes = (await $.state.get(lanesRef)).value ?? emptyFeed()
    const next = await readLanes(run, prevLanes, now)
    if (!sameFeed(prevLanes, next)) await $.state.set(lanesRef, next)
  }
  // The SPEC list is read on tab open and on refresh only, never by the timer.
  if (want.spec) {
    const prevSpecs = (await $.state.get(specsRef)).value ?? emptyFeed()
    const next = await readSpecList(run, prevSpecs, now)
    if (!sameFeed(prevSpecs, next)) await $.state.set(specsRef, next)
  }
}

// At most one poll in flight. A user action that arrives while one runs is remembered
// and served right after it; a timer tick that finds one running is dropped.
const refresh = async ($: EngineInterface, want: Wanted, isTick = false): Promise<void> => {
  if (!isTick) pending = merge(pending, want)
  if (!gate.tryStart()) return
  try {
    let next: Wanted | undefined = isTick ? merge(pending, want) : pending
    pending = undefined
    while (next !== undefined) {
      await pollOnce($, next)
      next = pending
      pending = undefined
    }
  } catch (err) {
    await soft($, async () => say($, `The board could not refresh: ${errorText(err)}`))
  } finally {
    gate.finish()
  }
}

const stopPolling = (): void => {
  timer?.cancel()
  timer = undefined
}

const startPolling = ($: EngineInterface): void => {
  stopPolling()
  timer = $.clock.every(clampInterval(POLL_INTERVAL_MS), () => {
    void refresh($, { lanes: false, spec: false }, true)
  })
}

// The pane's own close control. A plugin's own $.ui.close does not reach its own ui.close
// hook (measured with the test kit), so the timer is stopped here; the person's close mark
// and Esc do reach the hook below.
const closePane = async ($: EngineInterface): Promise<void> => {
  stopPolling()
  await $.ui.close({ id: PANE })
}

// ---- view navigation: local view state, and only argv of the fixed read-only table ------------
const switchTab = async ($: EngineInterface, tab: MoaiBoardTab): Promise<void> => {
  await patchView($, { tab, card: '', spec: '', file: '', page: 0 })
  await say($, '')
  await refresh($, { lanes: tab === 'lanes', spec: tab === 'spec' })
}

const refreshNow = async ($: EngineInterface): Promise<void> => {
  const view = await readView($)
  await say($, '')
  await refresh($, { lanes: view.tab === 'lanes', spec: view.tab === 'spec' })
}

// ---- the one write-capable action: pick (REQ-MBM-003 to REQ-MBM-005) ---------------------------
const onPickPress = async ($: EngineInterface, card: MoaiBoardCard): Promise<void> => {
  const argv = canPick(card) ? buildPickArgv(card.id, pickPrefix(card.text)) : undefined
  if (argv === undefined) return
  let answer: string
  try {
    answer = await $.ui.ask(`Pick card ${card.id} ("${cutChars(firstLine(card.text), 120)}")?`, {
      options: [CONFIRM_LABEL, CANCEL_LABEL],
      header: 'Pick card',
    })
  } catch {
    return // dismissed, or nobody to ask (-p run): no process, no change
  }
  if (!isConfirmed(answer)) return
  const result = await executePick(argv => runMoai($, argv), argv)
  await refresh($, { lanes: false, spec: false })
  const queued = (await $.state.get(queueRef)).value?.data
  const isUnconfirmed = result.kind === 'ok' && queued !== undefined && isStillQueued(queued.cards, card.id)
  const message =
    result.kind === 'failed'
      ? `Pick failed: ${result.text}`
      : isUnconfirmed
        ? 'The pick reported success but the card still reads queued (unconfirmed).'
        : `Picked ${card.id}.`
  await say($, message)
  if (result.kind === 'failed' || isUnconfirmed) $.ui.toast(message)
}

// ---- SPEC reading: the file calls are wired here, the guard itself is pure (specs.ts) -------------
const openSpecFile = async ($: EngineInterface, file: string): Promise<void> => {
  const view = await readView($)
  const root = await rootOf($)
  const io = {
    stat: (path: string) => $.fs.stat(path, { resolve: true }),
    read: (path: string) => $.fs.read(path),
  }
  const result = isSpecFile(file) ? await readSpecFile(io, root, view.spec, file) : undefined
  const label = `${view.spec}/${file}`
  const doc =
    result?.ok === true
      ? { spec: view.spec, file, ...chunkMarkdown(result.text, label) }
      : { spec: view.spec, file, chunks: [], notice: result?.ok === false ? result.reason : `"${file}" is not a SPEC file.` }
  await patchView($, { file, root })
  await $.state.set(docRef, doc)
}

const closeSpec = async ($: EngineInterface): Promise<void> => {
  await patchView($, { card: '', spec: '', file: '' })
  await $.state.set(docRef, undefined)
}

// ---- drawing ---------------------------------------------------------------------------------------
const readModel = async ($: EngineInterface, cols: number): Promise<Model> => ({
  cols,
  now: Date.now(),
  view: await readView($),
  queue: (await $.state.get(queueRef)).value ?? emptyFeed(),
  lanes: (await $.state.get(lanesRef)).value ?? emptyFeed(),
  specs: (await $.state.get(specsRef)).value ?? emptyFeed(),
  doc: (await $.state.get(docRef)).value,
  notice: (await $.state.get(noticeRef)).value ?? '',
})

const makeActions = ($: EngineInterface): Actions => ({
  tab: tab => void soft($, () => switchTab($, tab)),
  refresh: () => void soft($, () => refreshNow($)),
  back: () => void soft($, () => closeSpec($)),
  close: () => void soft($, () => closePane($)),
  openCard: id => void soft($, () => patchView($, { card: id })),
  pick: card => void soft($, () => onPickPress($, card)),
  openSpec: id => void soft($, () => patchView($, { spec: id, file: '' })),
  openFile: file => void soft($, () => openSpecFile($, file)),
  setStatus: status => void soft($, () => patchView($, { status, page: 0 })),
  setPage: page => void soft($, () => patchView($, { page })),
})

const rootOf = async ($: EngineInterface): Promise<string> => {
  try {
    return await $.session.root()
  } catch {
    return ''
  }
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    try {
      await $.command.register({
        name: 'moai-board',
        description: 'Open the moai board pane: queue, lanes, SPECs (read-only; pick asks first)',
      })
    } catch {
      // the command is simply absent; the session continues unaffected
    }
    return next(e)
  })

  on('command.run', { command: 'moai-board' }, async $ => {
    try {
      await $.ui.open({ id: PANE, title: 'moai-board' })
      await patchView($, { root: await rootOf($) })
      startPolling($)
      await refresh($, { lanes: false, spec: false })
      return { text: 'moai-board pane opened.' }
    } catch (err) {
      return { text: `moai-board could not open the pane: ${errorText(err)}` }
    }
  })

  on('ui.close', async ($, e, next) => {
    if (e.id === PANE) stopPolling()
    return next(e)
  })

  on('session.end', async ($, e, next) => {
    stopPolling()
    return next(e)
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const table = $.ui.resolve(e)
    try {
      const cols = typeof e.props.bodyColumns === 'number' ? e.props.bodyColumns : 80
      return drawBoard(table, await readModel($, cols), makeActions($))
    } catch (err) {
      const { Text } = table
      return <Text dimColor>{`moai-board could not draw: ${errorText(err)}`}</Text>
    }
  })
}
