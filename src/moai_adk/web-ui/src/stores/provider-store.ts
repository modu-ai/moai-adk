import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Provider, ProviderConfig, Model } from '@/types'

interface ProviderState {
  providers: ProviderConfig[]
  activeProvider: Provider | null
  activeModel: string | null
  isLoading: boolean
  error: string | null
}

interface ProviderActions {
  setProviders: (providers: ProviderConfig[]) => void
  setActiveProvider: (provider: Provider) => void
  setActiveModel: (model: string) => void
  switchProvider: (provider: Provider, model: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

type ProviderStore = ProviderState & ProviderActions

export const useProviderStore = create<ProviderStore>()(
  persist(
    (set) => ({
      providers: [],
      activeProvider: 'anthropic',
      activeModel: 'claude-sonnet-4-20250514',
      isLoading: false,
      error: null,

      setProviders: (providers) => set({ providers }),

      setActiveProvider: (provider) => set({ activeProvider: provider }),

      setActiveModel: (model) => set({ activeModel: model }),

      switchProvider: (provider, model) =>
        set({ activeProvider: provider, activeModel: model }),

      setLoading: (loading) => set({ isLoading: loading }),
      setError: (error) => set({ error }),
    }),
    {
      name: 'moai-provider-storage',
      partialize: (state) => ({
        activeProvider: state.activeProvider,
        activeModel: state.activeModel,
      }),
    }
  )
)

// Selectors
export const useActiveProviderConfig = () => {
  const providers = useProviderStore((state) => state.providers)
  const activeProvider = useProviderStore((state) => state.activeProvider)
  return providers.find((p) => p.id === activeProvider) ?? null
}

export const useActiveModelConfig = (): Model | null => {
  const providerConfig = useActiveProviderConfig()
  const activeModel = useProviderStore((state) => state.activeModel)
  if (!providerConfig || !activeModel) return null
  return providerConfig.models.find((m) => m.id === activeModel) ?? null
}

export const useAvailableModels = () => {
  const providerConfig = useActiveProviderConfig()
  return providerConfig?.models ?? []
}
