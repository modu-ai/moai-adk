// State contract of the moai-board plugin (Claude Code function hooks, 2.1.287).
// Self-contained on purpose: the engine requires a contract with no import, its
// exported names led by the plugin's PascalCase name, and `PluginState` declared
// for the plugin's own name. `plugin.json` names this file under "types".

export type MoaiBoardTab = 'queue' | 'lanes' | 'spec'

export type MoaiBoardCardState = 'picked' | 'queued' | 'hold'

/** One non-dropped queue item. `addedAt` and `specId` are '' when the source did not carry them. */
export type MoaiBoardCard = {
  id: string
  state: MoaiBoardCardState
  text: string
  addedAt: string
  specId: string
}

/** The last good data of one source, when it was read, and the cause of the latest failure ('' when it succeeded). */
export type MoaiBoardFeed<T> = {
  data: T | undefined
  at: number
  error: string
}

export type MoaiBoardQueue = {
  cards: MoaiBoardCard[]
  /** True when the rows came from the text list: detail columns (spec id, added date) are missing. */
  isReduced: boolean
}

export type MoaiBoardLaneCard = {
  id: string
  owner: string
  state: string
  stage: string
  specId: string
  isLeaseExpired: boolean
}

export type MoaiBoardSession = {
  sessionId: string
  specId: string
  phase: string
  heartbeatMs: number
}

export type MoaiBoardLanes = {
  cards: MoaiBoardLaneCard[]
  sessions: MoaiBoardSession[]
}

export type MoaiBoardSpecRow = { id: string; status: string }

export type MoaiBoardSpecs = { rows: MoaiBoardSpecRow[] }

/** The SPEC file shown in the SPEC tab: at most 12 chunks of at most 9,000 characters, plus a notice. */
export type MoaiBoardDoc = {
  spec: string
  file: string
  chunks: string[]
  notice: string
}

export type MoaiBoardView = {
  tab: MoaiBoardTab
  /** Card id whose detail is open, '' for the list. */
  card: string
  /** SPEC id whose files are open, '' for the list. */
  spec: string
  /** The SPEC file shown, '' while none is chosen. */
  file: string
  /** 'active' (draft and in-progress) or one status word. */
  status: string
  page: number
  /** The session's project root, shown in the header. */
  root: string
}

declare module 'claude-code' {
  interface PluginState {
    'moai-board': {
      view: MoaiBoardView
      queue: MoaiBoardFeed<MoaiBoardQueue>
      lanes: MoaiBoardFeed<MoaiBoardLanes>
      specs: MoaiBoardFeed<MoaiBoardSpecs>
      doc: MoaiBoardDoc | undefined
      notice: string
    }
  }
}
