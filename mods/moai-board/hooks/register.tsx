// moai-board hooks module (M1 skeleton). The only file that spells `$.…`.
import type { EngineInterface, Register } from 'claude-code'
import { CMD_TIMEOUT_MS } from './data'

const PANE = 'moai-board'

// The single process.run call site (REQ-MBM-013). Helpers receive it as a function.
const runMoai = ($: EngineInterface, argv: readonly string[]) =>
  $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })

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
    return { text: 'moai-board pane opened.' }
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Text } = $.ui.resolve(e)
    void runMoai
    return <Text dimColor>moai-board</Text>
  })
}
