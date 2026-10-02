// moai-board view layer: `$`-free. register.tsx resolves the element table of the
// surface being drawn and hands it in with the model and the press handlers, so this
// module draws with Box, Text, Button and Markdown only (present on every surface).
import type {
  MoaiBoardCard,
  MoaiBoardDoc,
  MoaiBoardFeed,
  MoaiBoardLanes,
  MoaiBoardQueue,
  MoaiBoardSpecs,
  MoaiBoardTab,
  MoaiBoardView,
} from '../types'
import {
  canPick,
  cutChars,
  feedStatus,
  firstLine,
  groupLanes,
  laneCardText,
  REDUCED_NOTICE,
  sessionText,
  summaryLine,
} from './data'
import { SPEC_FILES, chunkMarkdown, filterSpecs, pageOf, statusChips } from './specs'

/** The four constructors every surface's element table has (the typings' `Elements`). */
// The table is a per-surface union; only these four constructors are drawn, typed loosely on purpose.
export type Table = { Box: any; Text: any; Button: any; Markdown: any }

export type Model = {
  /** The pane's body width in character cells. */
  cols: number
  now: number
  view: MoaiBoardView
  queue: MoaiBoardFeed<MoaiBoardQueue>
  lanes: MoaiBoardFeed<MoaiBoardLanes>
  specs: MoaiBoardFeed<MoaiBoardSpecs>
  doc: MoaiBoardDoc | undefined
  notice: string
}

export type Actions = {
  tab: (tab: MoaiBoardTab) => void
  refresh: () => void
  back: () => void
  close: () => void
  openCard: (id: string) => void
  pick: (card: MoaiBoardCard) => void
  openSpec: (id: string) => void
  openFile: (file: string) => void
  setStatus: (status: string) => void
  setPage: (page: number) => void
}

const STATE_GROUPS = [
  { state: 'picked', label: 'Picked' },
  { state: 'queued', label: 'Queued' },
  { state: 'hold', label: 'Held' },
] as const

const rowLabel = (card: MoaiBoardCard, cols: number): string =>
  `${card.id}  ${cutChars(firstLine(card.text).replace(/\s+/g, ' '), Math.max(10, cols - card.id.length - 6))}`

export const drawBoard = (T: Table, m: Model, a: Actions) => {
  const { Box, Text, Button, Markdown } = T
  const { view } = m

  // Text takes no key, so every line a test or a reader must find sits in a keyed Box.
  const line = (key: string, text: string, props: Record<string, unknown> = {}) => (
    <Box key={key}>
      <Text {...props}>{text}</Text>
    </Box>
  )

  const tabButton = (tab: MoaiBoardTab, label: string, hotkey: string) => (
    <Button
      key={`tab-${tab}`}
      label={label}
      hotkey={hotkey}
      {...(view.tab === tab ? { variant: 'primary' } : {})}
      onPress={() => a.tab(tab)}
    />
  )

  // A failed read keeps the last good rows (dimmed) and names the cause on one line.
  const status = feedStatus(view.tab === 'lanes' ? m.lanes : view.tab === 'spec' ? m.specs : m.queue, m.now)
  const statusText = status.cause === '' ? '' : `${status.cause}${status.age === '' ? '' : ` (${status.age})`}`

  const queueBody = () => {
    const feed = m.queue
    const data = feed.data
    if (data === undefined)
      return line('queue-wait', feed.error === '' ? 'Reading the queue...' : 'No queue data to show.', { dimColor: true })
    const dim = feedStatus(feed, m.now).isDim
    if (view.card !== '') {
      const card = data.cards.find(c => c.id === view.card)
      if (card === undefined) return line('card-gone', `Card ${view.card} is no longer in the queue.`, { dimColor: true })
      const meta = [card.state, card.addedAt === '' ? '' : `added ${card.addedAt.slice(0, 10)}`, card.specId].filter(v => v !== '')
      return (
        <Box key="card-detail" flexDirection="column">
          {line('card-meta', `${card.id}  ${meta.join(' \u00b7 ')}`, { bold: true })}
          {chunkMarkdown(card.text, card.id).chunks.map((chunk, i) => (
            <Markdown key={`card-text-${i}`} dimColor={dim} text={chunk} />
          ))}
          <Box flexDirection="row" gap={1}>
            <Button key="back" label="back" hotkey="b" plain onPress={() => a.back()} />
            {canPick(card) && <Button key={`pick:${card.id}`} label="pick" hotkey="p" variant="primary" onPress={() => a.pick(card)} />}
          </Box>
        </Box>
      )
    }
    if (data.cards.length === 0) return line('queue-empty', 'The queue is empty.', { dimColor: true })
    return (
      <Box key="queue-list" flexDirection="column">
        {data.isReduced && line('reduced', REDUCED_NOTICE, { dimColor: true, wrap: 'wrap' })}
        {STATE_GROUPS.map(g => ({ ...g, cards: data.cards.filter(c => c.state === g.state) }))
          .filter(g => g.cards.length > 0)
          .map(g => (
            <Box key={`group-${g.state}`} flexDirection="column">
              {line(`head-${g.state}`, `${g.label} (${g.cards.length})`, { bold: true })}
              {g.cards.map(c => (
                <Button key={`open:${c.id}`} plain dimColor={dim} label={rowLabel(c, m.cols)} onPress={() => a.openCard(c.id)} />
              ))}
            </Box>
          ))}
      </Box>
    )
  }

  const lanesBody = () => {
    const data = m.lanes.data
    if (data === undefined)
      return line('lanes-wait', m.lanes.error === '' ? 'Reading the lanes...' : 'No lane data to show.', { dimColor: true })
    const dim = feedStatus(m.lanes, m.now).isDim
    const groups = groupLanes(data.cards)
    return (
      <Box key="lanes-list" flexDirection="column">
        {groups.length === 0 && line('lanes-none', 'No factory cards are running.', { dimColor: true })}
        {groups.map(g => (
          <Box key={`owner-${g.owner}`} flexDirection="column">
            {line(`owner:${g.owner}`, g.owner, { bold: true })}
            {g.cards.map(c => line(`lane:${c.id}`, `  ${laneCardText(c)}`, { dimColor: dim }))}
          </Box>
        ))}
        {line('sessions-head', 'Sessions with a heartbeat in the last 24 h', { bold: true })}
        {data.sessions.length === 0 && line('sessions-none', 'None.', { dimColor: true })}
        {data.sessions.map(s => line(`session:${s.sessionId}`, `  ${sessionText(s, m.now)}`, { dimColor: dim }))}
      </Box>
    )
  }

  const specsBody = () => {
    const data = m.specs.data
    if (view.spec !== '') {
      const doc = m.doc !== undefined && m.doc.spec === view.spec && m.doc.file === view.file ? m.doc : undefined
      return (
        <Box key="spec-detail" flexDirection="column">
          {line('spec-head', view.spec, { bold: true })}
          <Box flexDirection="row" gap={1}>
            {SPEC_FILES.map(f => (
              <Button key={`file:${f}`} plain label={f} {...(view.file === f ? { variant: 'primary' } : {})} onPress={() => a.openFile(f)} />
            ))}
          </Box>
          <Box flexDirection="row" gap={1}>
            <Button key="back" label="back" hotkey="b" plain onPress={() => a.back()} />
          </Box>
          {view.file === '' && line('spec-pick', 'Choose a file to read.', { dimColor: true })}
          {doc !== undefined && doc.notice !== '' && line('doc-notice', doc.notice, { wrap: 'wrap' })}
          {doc?.chunks.map((chunk, i) => <Markdown key={`doc-${i}`} text={chunk} />)}
        </Box>
      )
    }
    if (data === undefined)
      return line('specs-wait', m.specs.error === '' ? 'Reading the SPEC list...' : 'No SPEC data to show.', { dimColor: true })
    const dim = feedStatus(m.specs, m.now).isDim
    const filtered = filterSpecs(data.rows, view.status)
    const pg = pageOf(filtered, view.page)
    return (
      <Box key="specs-list" flexDirection="column">
        <Box flexDirection="row" gap={1}>
          {statusChips(data.rows).map(s => (
            <Button
              key={`filter:${s}`}
              plain
              label={s === 'active' ? 'active (draft, in-progress)' : s}
              {...(view.status === s ? { variant: 'primary' } : {})}
              onPress={() => a.setStatus(s)}
            />
          ))}
        </Box>
        {filtered.length === 0 && line('specs-none', 'No SPECs with this status.', { dimColor: true })}
        {pg.rows.map(r => (
          <Button key={`spec:${r.id}`} plain dimColor={dim} label={`${r.id}  ${r.status}`} onPress={() => a.openSpec(r.id)} />
        ))}
        <Box flexDirection="row" gap={1}>
          {pg.page > 0 && <Button key="page-prev" label="previous page" plain onPress={() => a.setPage(pg.page - 1)} />}
          {line('page', `page ${pg.page + 1} of ${pg.pages} \u00b7 ${filtered.length} SPECs`, { dimColor: true })}
          {pg.page < pg.pages - 1 && <Button key="page-next" label="next page" plain onPress={() => a.setPage(pg.page + 1)} />}
        </Box>
      </Box>
    )
  }

  const body = () => {
    switch (view.tab) {
      case 'queue':
        return queueBody()
      case 'lanes':
        return lanesBody()
      case 'spec':
        return specsBody()
    }
  }

  return (
    <Box flexDirection="column">
      <Box flexDirection="row" gap={1}>
        <Text bold>moai-board</Text>
        {view.root !== '' && <Text dimColor>{cutChars(view.root, Math.max(10, m.cols - 24))}</Text>}
        <Button key="close" label="close" hotkey="x" plain onPress={() => a.close()} />
      </Box>
      <Box flexDirection="row" gap={1}>
        {tabButton('queue', 'Queue', '1')}
        {tabButton('lanes', 'Lanes', '2')}
        {tabButton('spec', 'SPEC', '3')}
        <Button key="refresh" label="refresh" hotkey="r" plain onPress={() => a.refresh()} />
      </Box>
      {line('summary', m.queue.data === undefined ? 'Queue not read yet' : summaryLine(m.queue.data.cards))}
      {statusText !== '' && line('status', statusText, { bold: true, wrap: 'wrap' })}
      {m.notice !== '' && line('notice', m.notice, { wrap: 'wrap' })}
      {body()}
    </Box>
  )
}
