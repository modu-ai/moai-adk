import { create } from 'zustand'
import type { Session, Message } from '@/types'

interface SessionState {
  sessions: Session[]
  activeSessionId: string | null
  isLoading: boolean
  error: string | null
}

interface SessionActions {
  setSessions: (sessions: Session[]) => void
  addSession: (session: Session) => void
  removeSession: (id: string) => void
  setActiveSession: (id: string | null) => void
  updateSession: (id: string, updates: Partial<Session>) => void
  addMessage: (sessionId: string, message: Message) => void
  updateMessage: (sessionId: string, messageId: string, updates: Partial<Message>) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
}

type SessionStore = SessionState & SessionActions

export const useSessionStore = create<SessionStore>((set) => ({
  sessions: [],
  activeSessionId: null,
  isLoading: false,
  error: null,

  setSessions: (sessions) => set({ sessions }),

  addSession: (session) =>
    set((state) => ({
      sessions: [session, ...state.sessions],
      activeSessionId: session.id,
    })),

  removeSession: (id) =>
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== id),
      activeSessionId: state.activeSessionId === id ? null : state.activeSessionId,
    })),

  setActiveSession: (id) => set({ activeSessionId: id }),

  updateSession: (id, updates) =>
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === id ? { ...s, ...updates, updatedAt: new Date().toISOString() } : s
      ),
    })),

  addMessage: (sessionId, message) =>
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === sessionId
          ? {
              ...s,
              messages: [...s.messages, message],
              updatedAt: new Date().toISOString(),
              totalCost: s.totalCost + (message.cost ?? 0),
              totalTokens: {
                input: s.totalTokens.input + (message.tokens?.input ?? 0),
                output: s.totalTokens.output + (message.tokens?.output ?? 0),
              },
            }
          : s
      ),
    })),

  updateMessage: (sessionId, messageId, updates) =>
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.id === sessionId
          ? {
              ...s,
              messages: s.messages.map((m) =>
                m.id === messageId ? { ...m, ...updates } : m
              ),
            }
          : s
      ),
    })),

  setLoading: (loading) => set({ isLoading: loading }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),
}))

// Selectors
export const useActiveSession = () => {
  const sessions = useSessionStore((state) => state.sessions)
  const activeId = useSessionStore((state) => state.activeSessionId)
  return sessions.find((s) => s.id === activeId) ?? null
}

export const useSessionMessages = (sessionId: string | null) => {
  const sessions = useSessionStore((state) => state.sessions)
  if (!sessionId) return []
  const session = sessions.find((s) => s.id === sessionId)
  return session?.messages ?? []
}
