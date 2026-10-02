// Engine tests that mount the Pane through the plugin on two surfaces (never an assumed
// one). They exercise the hooks and the tree they return, not any surface's paint.
import { expect, test } from 'claude-code/testing'
import { PANE, queueFixture, setup } from './support'

const SURFACES = ['terminal', 'desktop'] as const
const SEP = ' · '

test('board: pane draws on terminal and desktop', async ($, on) => {
  setup(on, queueFixture({ picked: 17, queued: 44, hold: 2, dropped: 5 }))
  await $.command.run({ command: 'moai-board' })
  for (const surface of SURFACES) {
    const ui = await $.ui.mount({ plugin: 'moai-board', surface, component: 'Pane', requestId: 'moai-board', props: PANE })
    for (const key of ['tab-queue', 'tab-lanes', 'tab-spec', 'refresh'])
      expect(await ui.find({ key })).toBeDefined()
    const summary = await ui.find({ key: 'summary' })
    expect(summary?.text).toBe(`Picked 17${SEP}Queued 44${SEP}Held 2`)
    await ui.unmount()
  }
})

test('board: tab buttons switch the view', async ($, on) => {
  setup(on)
  await $.command.run({ command: 'moai-board' })
  for (const surface of SURFACES) {
    const ui = await $.ui.mount({ plugin: 'moai-board', surface, component: 'Pane', requestId: 'moai-board', props: PANE })
    expect(await ui.find({ key: 'summary' })).toBeDefined()
    await ui.press({ key: 'tab-lanes' })
    expect(await ui.find({ key: 'lane:t1399' })).toBeDefined()
    expect((await ui.find({ key: 'lane:t1399' }))?.text).toContain('lease expired')
    expect(await ui.find({ key: 'summary' })).toBeDefined()
    await ui.press({ key: 'tab-queue' })
    expect(await ui.find({ key: 'lane:t1399' })).toBeUndefined()
    expect(await ui.find({ key: 'open:t2' })).toBeDefined()
    await ui.unmount()
  }
})

test('board: refresh button re-runs the read argv', async ($, on) => {
  const { calls } = setup(on)
  await $.command.run({ command: 'moai-board' })
  for (const surface of SURFACES) {
    const ui = await $.ui.mount({ plugin: 'moai-board', surface, component: 'Pane', requestId: 'moai-board', props: PANE })
    const before = calls.length
    await ui.press({ key: 'refresh' })
    const added = calls.slice(before).map(argv => argv.join(' '))
    expect(added).toContain('moai gtd list --json')
    for (const argv of added) expect(argv.includes('next')).toBe(false)
    await ui.unmount()
  }
})
