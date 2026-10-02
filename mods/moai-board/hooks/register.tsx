// moai-board hooks module. The only file that spells `$.…`: helper modules
// (data.ts, specs.ts, view.tsx) receive functions or resolved element tables,
// never `$` (the engine refuses `$` passed into an imported function, spec.md §4).
import type { EngineInterface, Register } from 'claude-code'
import {
  CMD_TIMEOUT_MS,
  POLL_INTERVAL_MS,
  clampInterval,
  createFlightGate,
  createQueueReaders,
  emptyFeed,
  readLanes,
  readQueue,
  sameFeed,
} from './data'

const PANE = 'moai-board'

// Typed references: plugin and key are literals, used for nothing but $.state calls.
const queueRef = { plugin: 'moai-board', key: 'queue' } as const
const lanesRef = { plugin: 'moai-board', key: 'lanes' } as const
const noticeRef = { plugin: 'moai-board', key: 'notice' } as const

// The single process.run call site (REQ-MBM-013). Helpers receive it as a function.
const runMoai = ($: EngineInterface, argv: readonly string[]) =>
  $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })

// ---- polling (REQ-MBM-006, REQ-MBM-007) --------------------------------------------
// Module variables hold only what a hot reload may lose at the cost of one re-parse
// or one timer restart; view and data state live in $.state (REQ-MBM-015).
const readers = createQueueReaders()
const gate = createFlightGate()
let timer: { cancel: () => void } | undefined

type Wanted = { lanes: boolean }

let pending: Wanted | undefined

const merge = (a: Wanted | undefined, b: Wanted): Wanted => ({ lanes: (a?.lanes ?? false) || b.lanes })

const pollOnce = async ($: EngineInterface, want: Wanted): Promise<void> => {
  const run = (argv: readonly string[]) => runMoai($, argv)
  const now = Date.now()
  const prevQueue = (await $.state.get(queueRef)).value ?? emptyFeed()
  const q = await readQueue(run, readers, prevQueue, now)
  if (q.isChanged) await $.state.set(queueRef, q.feed)
  if (want.lanes) {
    const prevLanes = (await $.state.get(lanesRef)).value ?? emptyFeed()
    const next = await readLanes(run, prevLanes, now)
    if (!sameFeed(prevLanes, next)) await $.state.set(lanesRef, next)
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
    await $.state.set(noticeRef, `The board could not refresh: ${err instanceof Error ? err.message : String(err)}`)
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
    void refresh($, { lanes: false }, true)
  })
}

// The pane's own close control. A plugin's own $.ui.close does not reach its own ui.close
// hook (measured with the test kit), so the timer is stopped here; the person's close mark
// and Esc do reach the hook below.
const closePane = async ($: EngineInterface): Promise<void> => {
  stopPolling()
  await $.ui.close({ id: PANE })
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'moai-board',
      description: 'Open the moai board pane: queue, lanes, SPECs (read-only; pick asks first)',
    })
    return next(e)
  })

  on('command.run', { command: 'moai-board' }, async $ => {
    await $.ui.open({ id: PANE, title: 'moai-board' })
    startPolling($)
    await refresh($, { lanes: false })
    return { text: 'moai-board pane opened.' }
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
    const { Box, Button, Text } = $.ui.resolve(e)
    return (
      <Box flexDirection="row" gap={1}>
        <Text dimColor>moai-board</Text>
        <Button key="close" label="close" hotkey="x" plain onPress={() => closePane($)} />
      </Box>
    )
  })
}
