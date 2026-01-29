# State Management

Comprehensive patterns for modern state management covering Zustand, Redux Toolkit, Pinia, and React Context patterns.

---

## Zustand

### Basic Store Setup

```typescript
// stores/counterStore.ts
import { create } from 'zustand'

interface CounterState {
  count: number
  increment: () => void
  decrement: () => void
  reset: () => void
  incrementBy: (amount: number) => void
}

export const useCounterStore = create<CounterState>((set, get) => ({
  count: 0,

  increment: () => set((state) => ({ count: state.count + 1 })),

  decrement: () => set((state) => ({ count: state.count - 1 })),

  reset: () => set({ count: 0 }),

  incrementBy: (amount) => set((state) => ({ count: state.count + amount })),
}))

// Usage in component
function Counter() {
  const count = useCounterStore((state) => state.count)
  const increment = useCounterStore((state) => state.increment)

  return (
    <button onClick={increment}>Count: {count}</button>
  )
}
```

### Advanced Store with Middleware

```typescript
// stores/userStore.ts
import { create } from 'zustand'
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

interface User {
  id: string
  name: string
  email: string
  preferences: {
    theme: 'light' | 'dark'
    notifications: boolean
  }
}

interface UserState {
  user: User | null
  isLoading: boolean
  error: string | null

  // Actions
  setUser: (user: User | null) => void
  updatePreferences: (prefs: Partial<User['preferences']>) => void
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

export const useUserStore = create<UserState>()(
  devtools(
    persist(
      subscribeWithSelector(
        immer((set, get) => ({
          user: null,
          isLoading: false,
          error: null,

          setUser: (user) => {
            set({ user })
          },

          updatePreferences: (prefs) => {
            set((state) => {
              if (state.user) {
                state.user.preferences = {
                  ...state.user.preferences,
                  ...prefs
                }
              }
            })
          },

          login: async (email, password) => {
            set({ isLoading: true, error: null })

            try {
              const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password }),
              })

              if (!response.ok) {
                throw new Error('Login failed')
              }

              const user = await response.json()
              set({ user, isLoading: false })
            } catch (error) {
              set({
                error: error instanceof Error ? error.message : 'Unknown error',
                isLoading: false
              })
            }
          },

          logout: () => {
            set({ user: null })
          },
        }))
      ),
      {
        name: 'user-storage',
        partialize: (state) => ({
          user: state.user
        }),
      }
    ),
    { name: 'UserStore' }
  )
)

// Selectors for optimized re-renders
export const selectUser = (state: UserState) => state.user
export const selectIsLoggedIn = (state: UserState) => !!state.user
export const selectTheme = (state: UserState) => state.user?.preferences.theme ?? 'light'

// Subscribe to changes outside React
useUserStore.subscribe(
  (state) => state.user,
  (user) => {
    console.log('User changed:', user)
  }
)
```

### Slices Pattern for Large Stores

```typescript
// stores/slices/cartSlice.ts
import { StateCreator } from 'zustand'

interface CartItem {
  id: string
  name: string
  price: number
  quantity: number
}

export interface CartSlice {
  items: CartItem[]
  addItem: (item: Omit<CartItem, 'quantity'>) => void
  removeItem: (id: string) => void
  updateQuantity: (id: string, quantity: number) => void
  clearCart: () => void
  totalItems: () => number
  totalPrice: () => number
}

export const createCartSlice: StateCreator<CartSlice> = (set, get) => ({
  items: [],

  addItem: (item) => {
    set((state) => {
      const existingItem = state.items.find((i) => i.id === item.id)
      if (existingItem) {
        return {
          items: state.items.map((i) =>
            i.id === item.id ? { ...i, quantity: i.quantity + 1 } : i
          ),
        }
      }
      return { items: [...state.items, { ...item, quantity: 1 }] }
    })
  },

  removeItem: (id) => {
    set((state) => ({
      items: state.items.filter((item) => item.id !== id),
    }))
  },

  updateQuantity: (id, quantity) => {
    set((state) => ({
      items: state.items.map((item) =>
        item.id === id ? { ...item, quantity } : item
      ),
    }))
  },

  clearCart: () => set({ items: [] }),

  totalItems: () => get().items.reduce((sum, item) => sum + item.quantity, 0),

  totalPrice: () =>
    get().items.reduce((sum, item) => sum + item.price * item.quantity, 0),
})

// stores/slices/uiSlice.ts
export interface UISlice {
  sidebarOpen: boolean
  modalOpen: boolean
  toggleSidebar: () => void
  toggleModal: () => void
}

export const createUISlice: StateCreator<UISlice> = (set) => ({
  sidebarOpen: true,
  modalOpen: false,
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  toggleModal: () => set((state) => ({ modalOpen: !state.modalOpen })),
})

// stores/index.ts - Combine slices
import { create } from 'zustand'
import { createCartSlice, type CartSlice } from './slices/cartSlice'
import { createUISlice, type UISlice } from './slices/uiSlice'

type StoreState = CartSlice & UISlice

export const useStore = create<StoreState>()((...args) => ({
  ...createCartSlice(...args),
  ...createUISlice(...args),
}))
```

---

## Redux Toolkit

### Store Configuration

```typescript
// store/store.ts
import { configureStore } from '@reduxjs/toolkit'
import { setupListeners } from '@reduxjs/toolkit/query'
import userReducer from './slices/userSlice'
import cartReducer from './slices/cartSlice'
import { api } from './api'

export const store = configureStore({
  reducer: {
    user: userReducer,
    cart: cartReducer,
    [api.reducerPath]: api.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(api.middleware),
  devTools: process.env.NODE_ENV !== 'production',
})

setupListeners(store.dispatch)

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
```

### Slice with Async Thunks

```typescript
// store/slices/userSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'

interface User {
  id: string
  name: string
  email: string
}

interface UserState {
  currentUser: User | null
  users: User[]
  status: 'idle' | 'loading' | 'succeeded' | 'failed'
  error: string | null
}

const initialState: UserState = {
  currentUser: null,
  users: [],
  status: 'idle',
  error: null,
}

// Async thunks
export const fetchUsers = createAsyncThunk(
  'users/fetchUsers',
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/users')
      if (!response.ok) {
        throw new Error('Failed to fetch users')
      }
      return await response.json()
    } catch (error) {
      return rejectWithValue(
        error instanceof Error ? error.message : 'Unknown error'
      )
    }
  }
)

export const createUser = createAsyncThunk(
  'users/createUser',
  async (userData: Omit<User, 'id'>, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(userData),
      })
      if (!response.ok) {
        throw new Error('Failed to create user')
      }
      return await response.json()
    } catch (error) {
      return rejectWithValue(
        error instanceof Error ? error.message : 'Unknown error'
      )
    }
  }
)

const userSlice = createSlice({
  name: 'users',
  initialState,
  reducers: {
    setCurrentUser: (state, action: PayloadAction<User | null>) => {
      state.currentUser = action.payload
    },
    updateUser: (state, action: PayloadAction<Partial<User> & { id: string }>) => {
      const index = state.users.findIndex((u) => u.id === action.payload.id)
      if (index !== -1) {
        state.users[index] = { ...state.users[index], ...action.payload }
      }
    },
    clearError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch users
      .addCase(fetchUsers.pending, (state) => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(fetchUsers.fulfilled, (state, action) => {
        state.status = 'succeeded'
        state.users = action.payload
      })
      .addCase(fetchUsers.rejected, (state, action) => {
        state.status = 'failed'
        state.error = action.payload as string
      })
      // Create user
      .addCase(createUser.fulfilled, (state, action) => {
        state.users.push(action.payload)
      })
  },
})

export const { setCurrentUser, updateUser, clearError } = userSlice.actions
export default userSlice.reducer

// Selectors
export const selectAllUsers = (state: RootState) => state.user.users
export const selectUserById = (state: RootState, userId: string) =>
  state.user.users.find((u) => u.id === userId)
export const selectUserStatus = (state: RootState) => state.user.status
```

### RTK Query for API Calls

```typescript
// store/api.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

interface User {
  id: string
  name: string
  email: string
}

interface Post {
  id: string
  title: string
  content: string
  authorId: string
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: '/api',
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).auth.token
      if (token) {
        headers.set('Authorization', `Bearer ${token}`)
      }
      return headers
    },
  }),
  tagTypes: ['User', 'Post'],
  endpoints: (builder) => ({
    // Users
    getUsers: builder.query<User[], void>({
      query: () => '/users',
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'User' as const, id })),
              { type: 'User', id: 'LIST' },
            ]
          : [{ type: 'User', id: 'LIST' }],
    }),

    getUser: builder.query<User, string>({
      query: (id) => `/users/${id}`,
      providesTags: (result, error, id) => [{ type: 'User', id }],
    }),

    createUser: builder.mutation<User, Omit<User, 'id'>>({
      query: (body) => ({
        url: '/users',
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'User', id: 'LIST' }],
    }),

    updateUser: builder.mutation<User, Partial<User> & Pick<User, 'id'>>({
      query: ({ id, ...body }) => ({
        url: `/users/${id}`,
        method: 'PATCH',
        body,
      }),
      invalidatesTags: (result, error, { id }) => [{ type: 'User', id }],
    }),

    deleteUser: builder.mutation<void, string>({
      query: (id) => ({
        url: `/users/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (result, error, id) => [
        { type: 'User', id },
        { type: 'User', id: 'LIST' },
      ],
    }),

    // Posts with optimistic updates
    addPost: builder.mutation<Post, Omit<Post, 'id'>>({
      query: (body) => ({
        url: '/posts',
        method: 'POST',
        body,
      }),
      async onQueryStarted(newPost, { dispatch, queryFulfilled }) {
        // Optimistic update
        const patchResult = dispatch(
          api.util.updateQueryData('getPosts', undefined, (draft) => {
            draft.push({ ...newPost, id: 'temp-id' })
          })
        )

        try {
          await queryFulfilled
        } catch {
          patchResult.undo()
        }
      },
      invalidatesTags: [{ type: 'Post', id: 'LIST' }],
    }),
  }),
})

export const {
  useGetUsersQuery,
  useGetUserQuery,
  useCreateUserMutation,
  useUpdateUserMutation,
  useDeleteUserMutation,
  useAddPostMutation,
} = api
```

---

## React Context Pattern

### Type-Safe Context with Reducer

```typescript
// context/AuthContext.tsx
import { createContext, useContext, useReducer, useCallback, ReactNode } from 'react'

// Types
interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user'
}

interface AuthState {
  user: User | null
  isLoading: boolean
  error: string | null
}

type AuthAction =
  | { type: 'LOGIN_START' }
  | { type: 'LOGIN_SUCCESS'; payload: User }
  | { type: 'LOGIN_FAILURE'; payload: string }
  | { type: 'LOGOUT' }
  | { type: 'UPDATE_USER'; payload: Partial<User> }

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  updateUser: (updates: Partial<User>) => void
}

// Reducer
function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN_START':
      return { ...state, isLoading: true, error: null }
    case 'LOGIN_SUCCESS':
      return { ...state, isLoading: false, user: action.payload }
    case 'LOGIN_FAILURE':
      return { ...state, isLoading: false, error: action.payload }
    case 'LOGOUT':
      return { ...state, user: null }
    case 'UPDATE_USER':
      return {
        ...state,
        user: state.user ? { ...state.user, ...action.payload } : null,
      }
    default:
      return state
  }
}

// Context
const AuthContext = createContext<AuthContextValue | null>(null)

// Provider
interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(authReducer, {
    user: null,
    isLoading: false,
    error: null,
  })

  const login = useCallback(async (email: string, password: string) => {
    dispatch({ type: 'LOGIN_START' })

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })

      if (!response.ok) {
        throw new Error('Login failed')
      }

      const user = await response.json()
      dispatch({ type: 'LOGIN_SUCCESS', payload: user })
    } catch (error) {
      dispatch({
        type: 'LOGIN_FAILURE',
        payload: error instanceof Error ? error.message : 'Unknown error',
      })
      throw error
    }
  }, [])

  const logout = useCallback(() => {
    dispatch({ type: 'LOGOUT' })
  }, [])

  const updateUser = useCallback((updates: Partial<User>) => {
    dispatch({ type: 'UPDATE_USER', payload: updates })
  }, [])

  const value: AuthContextValue = {
    ...state,
    login,
    logout,
    updateUser,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// Hook
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
```

---

## Comparison and Selection Guide

### When to Use Each Solution

Zustand:
- Small to medium applications
- Simple global state needs
- Minimal boilerplate preference
- TypeScript-first approach

Redux Toolkit:
- Large applications with complex state
- Team already familiar with Redux
- Need for middleware ecosystem
- Complex async workflows

React Context:
- Theme/locale preferences
- Authentication state
- Component-level shared state
- Avoiding external dependencies

Pinia (Vue):
- Vue 3 applications
- Composition API integration
- DevTools support needed
- Type-safe store modules

### Performance Considerations

Zustand Performance Tips:
- Use selectors to minimize re-renders
- Split stores by domain
- Use shallow comparison when needed

Redux Performance Tips:
- Use createSelector for memoization
- Normalize state shape
- Split reducers appropriately

Context Performance Tips:
- Split contexts by update frequency
- Memoize context values
- Use useMemo for complex computations

---

Version: 2.0.0
Last Updated: 2026-01-06
