# Frontend Architecture Patterns — Component Design & State Management

_Last updated: 2025-11-22_

## Component Architecture

### React Server Components Pattern
```typescript
// app/users/page.tsx - Server Component
export default async function UsersPage() {
  const users = await fetch('http://api/users').then(r => r.json());

  return (
    <div>
      <h1>Users</h1>
      <UserList users={users} />
    </div>
  );
}

// app/users/UserList.tsx - Client Component
'use client';

import { useState } from 'react';

export function UserList({ users }) {
  const [filter, setFilter] = useState('');

  const filtered = users.filter(u =>
    u.name.includes(filter)
  );

  return (
    <div>
      <input
        value={filter}
        onChange={e => setFilter(e.target.value)}
        placeholder="Filter users"
      />
      {filtered.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
}
```

### Atomic Design Structure
```
components/
├── atoms/
│   ├── Button.tsx
│   ├── Input.tsx
│   └── Label.tsx
├── molecules/
│   ├── FormField.tsx
│   ├── Card.tsx
│   └── Modal.tsx
├── organisms/
│   ├── UserForm.tsx
│   ├── Navigation.tsx
│   └── Hero.tsx
└── templates/
    ├── PageLayout.tsx
    └── AuthLayout.tsx
```

## State Management

### Zustand Pattern (Recommended for React)
```typescript
import { create } from 'zustand';

interface UserStore {
  users: User[];
  selectedId: string | null;
  addUser: (user: User) => void;
  selectUser: (id: string) => void;
  getSelected: () => User | undefined;
}

export const useUserStore = create<UserStore>((set, get) => ({
  users: [],
  selectedId: null,

  addUser: (user) => set((state) => ({
    users: [...state.users, user]
  })),

  selectUser: (id) => set({ selectedId: id }),

  getSelected: () => {
    const state = get();
    return state.users.find(u => u.id === state.selectedId);
  }
}));

// Usage in component
function UserList() {
  const { users, selectUser } = useUserStore();

  return (
    <div>
      {users.map(user => (
        <button key={user.id} onClick={() => selectUser(user.id)}>
          {user.name}
        </button>
      ))}
    </div>
  );
}
```

### Redux Toolkit Pattern
```typescript
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface User {
  id: string;
  name: string;
}

const userSlice = createSlice({
  name: 'users',
  initialState: [] as User[],
  reducers: {
    userAdded: (state, action: PayloadAction<User>) => {
      state.push(action.payload);
    },
    userRemoved: (state, action: PayloadAction<string>) => {
      return state.filter(u => u.id !== action.payload);
    }
  }
});

export const { userAdded, userRemoved } = userSlice.actions;
export default userSlice.reducer;
```

## Form Management

### React Hook Form with Validation
```typescript
import { useForm } from 'react-hook-form';

interface FormData {
  email: string;
  password: string;
}

export function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<FormData>({
    defaultValues: {
      email: '',
      password: ''
    }
  });

  const onSubmit = async (data: FormData) => {
    await loginUser(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div>
        <input
          {...register('email', {
            required: 'Email is required',
            pattern: {
              value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
              message: 'Invalid email'
            }
          })}
          placeholder="Email"
        />
        {errors.email && <span>{errors.email.message}</span>}
      </div>

      <button type="submit">Login</button>
    </form>
  );
}
```

## Data Fetching Patterns

### SWR (Stale-While-Revalidate)
```typescript
import useSWR from 'swr';

const fetcher = (url: string) => fetch(url).then(r => r.json());

export function UsersList() {
  const { data: users, isLoading, error } = useSWR(
    '/api/users',
    fetcher,
    { revalidateOnFocus: false }
  );

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <ul>
      {users?.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  );
}
```

### React Query (TanStack Query)
```typescript
import { useQuery, useMutation } from '@tanstack/react-query';

export function UsersList() {
  const { data: users, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: () => fetch('/api/users').then(r => r.json())
  });

  const createUserMutation = useMutation({
    mutationFn: (newUser: User) =>
      fetch('/api/users', {
        method: 'POST',
        body: JSON.stringify(newUser)
      }).then(r => r.json()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    }
  });

  return (
    <div>
      {isLoading ? <div>Loading...</div> : (
        <ul>
          {users?.map(user => (
            <li key={user.id}>{user.name}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

---

**Last Updated**: 2025-11-22

