// Provider types
export type Provider = 'anthropic' | 'openai' | 'google' | 'ollama' | 'bedrock' | 'vertex'

export interface ProviderConfig {
  id: Provider
  name: string
  models: Model[]
  isAvailable: boolean
}

export interface Model {
  id: string
  name: string
  provider: Provider
  contextWindow: number
  inputCostPer1k: number
  outputCostPer1k: number
}

// Message types
export type MessageRole = 'user' | 'assistant' | 'system' | 'tool'

export interface Message {
  id: string
  role: MessageRole
  content: string
  timestamp: string
  tokens?: {
    input: number
    output: number
  }
  cost?: number
  toolCalls?: ToolCall[]
  isStreaming?: boolean
}

export interface ToolCall {
  id: string
  name: string
  arguments: Record<string, unknown>
  result?: string
  status: 'pending' | 'running' | 'completed' | 'error'
  duration?: number
}

// Session types
export interface Session {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  provider: Provider
  model: string
  messages: Message[]
  totalCost: number
  totalTokens: {
    input: number
    output: number
  }
}

// SPEC types
export type SpecStatus = 'draft' | 'planned' | 'approved' | 'in_progress' | 'implementing' | 'testing' | 'completed' | 'blocked'

export interface Spec {
  id: string
  title: string
  description: string
  status: SpecStatus
  priority: 'low' | 'medium' | 'high'
  createdAt: string
  updatedAt: string
  tags: string[]
  progress: number
  tasks: SpecTask[]
  filePath?: string
  worktreePath?: string
}

export interface SpecListResponse {
  specs: Spec[]
  total: number
}

export interface SpecTask {
  id: string
  title: string
  status: 'pending' | 'in_progress' | 'completed'
  assignedTo?: string
}

// Cost tracking types
export interface CostRecord {
  timestamp: string
  provider: Provider
  model: string
  inputTokens: number
  outputTokens: number
  cost: number
  sessionId: string
}

export interface CostSummary {
  daily: number
  weekly: number
  monthly: number
  byProvider: Record<Provider, number>
  byModel: Record<string, number>
}

// Terminal types
export interface TerminalOutput {
  id: string
  type: 'stdin' | 'stdout' | 'stderr'
  content: string
  timestamp: string
}

// WebSocket types
export type WSMessageType =
  | 'session.start'
  | 'session.end'
  | 'message.create'
  | 'message.update'
  | 'message.complete'
  | 'tool.start'
  | 'tool.update'
  | 'tool.complete'
  | 'stream.chunk'
  | 'error'
  | 'ping'
  | 'pong'

export interface WSMessage<T = unknown> {
  type: WSMessageType
  payload: T
  timestamp: string
}

// API response types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

// UI state types
export interface UIState {
  sidebarOpen: boolean
  activeTab: 'chat' | 'specs' | 'terminal' | 'costs'
  theme: 'light' | 'dark' | 'system'
}
