# Frontend Development - Working Examples

> Real-world React, Vue, Angular examples with modern patterns

---

## Example 1: React Todo App with TypeScript

### Project Structure
```
todo-app/
├── src/
│   ├── components/
│   │   ├── TodoList.tsx
│   │   ├── TodoItem.tsx
│   │   └── AddTodo.tsx
│   ├── hooks/
│   │   └── useTodos.ts
│   ├── types/
│   │   └── Todo.ts
│   └── App.tsx
├── package.json
└── vite.config.ts
```

### types/Todo.ts
```typescript
export interface Todo {
  id: string;
  text: string;
  completed: boolean;
  createdAt: Date;
}
```

### hooks/useTodos.ts
```typescript
import { useState, useCallback } from 'react';
import type { Todo } from '../types/Todo';

export function useTodos() {
  const [todos, setTodos] = useState<Todo[]>([]);

  const addTodo = useCallback((text: string) => {
    const newTodo: Todo = {
      id: crypto.randomUUID(),
      text,
      completed: false,
      createdAt: new Date()
    };
    setTodos(prev => [...prev, newTodo]);
  }, []);

  const toggleTodo = useCallback((id: string) => {
    setTodos(prev => prev.map(todo =>
      todo.id === id ? { ...todo, completed: !todo.completed } : todo
    ));
  }, []);

  const deleteTodo = useCallback((id: string) => {
    setTodos(prev => prev.filter(todo => todo.id !== id));
  }, []);

  return { todos, addTodo, toggleTodo, deleteTodo };
}
```

### components/TodoList.tsx
```typescript
import { Todo } from '../types/Todo';
import { TodoItem } from './TodoItem';

interface TodoListProps {
  todos: Todo[];
  onToggle: (id: string) => void;
  onDelete: (id: string) => void;
}

export function TodoList({ todos, onToggle, onDelete }: TodoListProps) {
  if (todos.length === 0) {
    return <p className="text-gray-500">No todos yet. Add one above!</p>;
  }

  return (
    <ul className="space-y-2">
      {todos.map(todo => (
        <TodoItem
          key={todo.id}
          todo={todo}
          onToggle={onToggle}
          onDelete={onDelete}
        />
      ))}
    </ul>
  );
}
```

---

## Example 2: Vue 3 Product Catalog

### ProductList.vue
```vue
<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Product } from '@/types';

interface Props {
  category?: string;
}

const props = defineProps<Props>();
const products = ref<Product[]>([]);
const searchQuery = ref('');

const filteredProducts = computed(() => {
  let result = products.value;
  
  if (props.category) {
    result = result.filter(p => p.category === props.category);
  }
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(p => 
      p.name.toLowerCase().includes(query)
    );
  }
  
  return result;
});

async function loadProducts() {
  const response = await fetch('/api/products');
  products.value = await response.json();
}

onMounted(loadProducts);
</script>

<template>
  <div>
    <input
      v-model="searchQuery"
      type="search"
      placeholder="Search products..."
      class="mb-4 px-4 py-2 border rounded"
    />
    
    <div class="grid grid-cols-3 gap-4">
      <div
        v-for="product in filteredProducts"
        :key="product.id"
        class="border rounded p-4"
      >
        <h3 class="font-bold">{{ product.name }}</h3>
        <p class="text-gray-600">\${{ product.price }}</p>
        <button class="mt-2 bg-blue-600 text-white px-4 py-2 rounded">
          Add to Cart
        </button>
      </div>
    </div>
  </div>
</template>
```

---

## Example 3: Angular User Management

### user.service.ts
```typescript
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface User {
  id: number;
  name: string;
  email: string;
}

@Injectable({ providedIn: 'root' })
export class UserService {
  private apiUrl = '/api/users';

  constructor(private http: HttpClient) {}

  getUsers(): Observable<User[]> {
    return this.http.get<User[]>(this.apiUrl);
  }

  getUser(id: number): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/${id}`);
  }

  createUser(user: Omit<User, 'id'>): Observable<User> {
    return this.http.post<User>(this.apiUrl, user);
  }
}
```

### user-list.component.ts
```typescript
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserService, User } from './user.service';

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div>
      <h2>Users</h2>
      <ul *ngIf="users.length > 0">
        <li *ngFor="let user of users">
          {{ user.name }} ({{ user.email }})
        </li>
      </ul>
      <p *ngIf="loading">Loading...</p>
      <p *ngIf="error" class="error">{{ error }}</p>
    </div>
  `
})
export class UserListComponent implements OnInit {
  users: User[] = [];
  loading = true;
  error: string | null = null;

  constructor(private userService: UserService) {}

  ngOnInit() {
    this.userService.getUsers().subscribe({
      next: (users) => {
        this.users = users;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load users';
        this.loading = false;
      }
    });
  }
}
```

---

**Last Updated**: 2025-10-22
**Frameworks**: React 19.0, Vue 3.5, Angular 19.0
