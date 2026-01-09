import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { UIState } from '@/types'

interface UIStore extends UIState {
  setSidebarOpen: (open: boolean) => void
  toggleSidebar: () => void
  setActiveTab: (tab: UIState['activeTab']) => void
  setTheme: (theme: UIState['theme']) => void
}

export const useUIStore = create<UIStore>()(
  persist(
    (set) => ({
      sidebarOpen: true,
      activeTab: 'chat',
      theme: 'system',

      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
      setActiveTab: (tab) => set({ activeTab: tab }),
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'moai-ui-storage',
      partialize: (state) => ({
        sidebarOpen: state.sidebarOpen,
        theme: state.theme,
      }),
    }
  )
)
