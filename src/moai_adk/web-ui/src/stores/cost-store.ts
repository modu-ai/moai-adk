import { create } from 'zustand'
import type { CostRecord, CostSummary, Provider } from '@/types'

interface CostState {
  records: CostRecord[]
  summary: CostSummary | null
  isLoading: boolean
  error: string | null
}

interface CostActions {
  setRecords: (records: CostRecord[]) => void
  addRecord: (record: CostRecord) => void
  setSummary: (summary: CostSummary) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

type CostStore = CostState & CostActions

const initialSummary: CostSummary = {
  daily: 0,
  weekly: 0,
  monthly: 0,
  byProvider: {} as Record<Provider, number>,
  byModel: {},
}

export const useCostStore = create<CostStore>((set) => ({
  records: [],
  summary: initialSummary,
  isLoading: false,
  error: null,

  setRecords: (records) => set({ records }),

  addRecord: (record) =>
    set((state) => {
      const newRecords = [...state.records, record]
      const summary = state.summary ?? initialSummary

      return {
        records: newRecords,
        summary: {
          ...summary,
          daily: summary.daily + record.cost,
          weekly: summary.weekly + record.cost,
          monthly: summary.monthly + record.cost,
          byProvider: {
            ...summary.byProvider,
            [record.provider]:
              (summary.byProvider[record.provider] ?? 0) + record.cost,
          },
          byModel: {
            ...summary.byModel,
            [record.model]: (summary.byModel[record.model] ?? 0) + record.cost,
          },
        },
      }
    }),

  setSummary: (summary) => set({ summary }),
  setLoading: (loading) => set({ isLoading: loading }),
  setError: (error) => set({ error }),
}))

// Selectors
export const useTotalCost = () => {
  return useCostStore((state) => state.summary?.monthly ?? 0)
}

export const useCostByProvider = (provider: Provider) => {
  return useCostStore((state) => state.summary?.byProvider[provider] ?? 0)
}

export const useRecentRecords = (limit = 10) => {
  return useCostStore((state) => state.records.slice(-limit))
}
