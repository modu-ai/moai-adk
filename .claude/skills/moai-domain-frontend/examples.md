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

## Example 4: Next.js E-commerce Product Page

### Full-Stack Product Page with SEO

```typescript
// app/products/[slug]/page.tsx
import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import Image from 'next/image';
import AddToCartButton from './AddToCartButton';

interface Product {
  id: string;
  slug: string;
  name: string;
  price: number;
  description: string;
  images: string[];
  inStock: boolean;
  rating: number;
  reviews: number;
}

async function getProduct(slug: string): Promise<Product | null> {
  const res = await fetch(`https://api.example.com/products/${slug}`, {
    next: { revalidate: 60 }, // Revalidate every 60 seconds
  });

  if (!res.ok) return null;
  return res.json();
}

export async function generateMetadata({
  params,
}: {
  params: { slug: string };
}): Promise<Metadata> {
  const product = await getProduct(params.slug);

  if (!product) {
    return {
      title: 'Product Not Found',
    };
  }

  return {
    title: `${product.name} | My Store`,
    description: product.description,
    openGraph: {
      title: product.name,
      description: product.description,
      images: [product.images[0]],
      type: 'product',
    },
  };
}

export default async function ProductPage({
  params,
}: {
  params: { slug: string };
}) {
  const product = await getProduct(params.slug);

  if (!product) {
    notFound();
  }

  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: product.name,
    image: product.images,
    description: product.description,
    offers: {
      '@type': 'Offer',
      price: product.price,
      priceCurrency: 'USD',
      availability: product.inStock
        ? 'https://schema.org/InStock'
        : 'https://schema.org/OutOfStock',
    },
    aggregateRating: {
      '@type': 'AggregateRating',
      ratingValue: product.rating,
      reviewCount: product.reviews,
    },
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* Image Gallery */}
        <div className="space-y-4">
          <Image
            src={product.images[0]}
            alt={product.name}
            width={600}
            height={600}
            priority
            className="rounded-lg"
          />
          <div className="grid grid-cols-4 gap-2">
            {product.images.slice(1, 5).map((img, i) => (
              <Image
                key={i}
                src={img}
                alt={`${product.name} view ${i + 2}`}
                width={150}
                height={150}
                className="rounded cursor-pointer hover:opacity-75"
              />
            ))}
          </div>
        </div>

        {/* Product Info */}
        <div className="space-y-6">
          <h1 className="text-3xl font-bold">{product.name}</h1>

          <div className="flex items-center gap-2">
            <div className="flex">
              {Array.from({ length: 5 }).map((_, i) => (
                <span key={i} className={i < product.rating ? 'text-yellow-400' : 'text-gray-300'}>
                  ★
                </span>
              ))}
            </div>
            <span className="text-sm text-gray-600">({product.reviews} reviews)</span>
          </div>

          <div className="text-4xl font-bold">${product.price.toFixed(2)}</div>

          <p className="text-gray-700">{product.description}</p>

          <div className="space-y-3">
            <div className="flex items-center gap-2">
              {product.inStock ? (
                <>
                  <span className="w-3 h-3 bg-green-500 rounded-full"></span>
                  <span className="text-green-700">In Stock</span>
                </>
              ) : (
                <>
                  <span className="w-3 h-3 bg-red-500 rounded-full"></span>
                  <span className="text-red-700">Out of Stock</span>
                </>
              )}
            </div>

            <AddToCartButton product={product} />
          </div>
        </div>
      </div>
    </div>
  );
}
```

### Client Component for Add to Cart

```typescript
// app/products/[slug]/AddToCartButton.tsx
'use client';

import { useState } from 'react';
import { useCartStore } from '@/store/cart';

export default function AddToCartButton({ product }) {
  const [quantity, setQuantity] = useState(1);
  const addItem = useCartStore((state) => state.addItem);
  const [added, setAdded] = useState(false);

  const handleAddToCart = () => {
    addItem({ ...product, quantity });
    setAdded(true);
    setTimeout(() => setAdded(false), 2000);
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3">
        <label htmlFor="quantity" className="font-medium">
          Quantity:
        </label>
        <select
          id="quantity"
          value={quantity}
          onChange={(e) => setQuantity(Number(e.target.value))}
          className="border rounded px-3 py-2"
        >
          {[1, 2, 3, 4, 5].map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
      </div>

      <button
        onClick={handleAddToCart}
        disabled={!product.inStock}
        className={`w-full py-3 px-6 rounded-lg font-semibold transition ${
          product.inStock
            ? 'bg-blue-600 text-white hover:bg-blue-700'
            : 'bg-gray-300 text-gray-500 cursor-not-allowed'
        }`}
      >
        {added ? '✓ Added to Cart!' : 'Add to Cart'}
      </button>
    </div>
  );
}
```

---

## Example 5: Vue 3 Real-time Dashboard

### Dashboard with WebSocket Updates

```vue
<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useWebSocket } from '@/composables/useWebSocket';

interface Metric {
  label: string;
  value: number;
  change: number;
  trend: 'up' | 'down';
}

interface Activity {
  id: string;
  user: string;
  action: string;
  timestamp: Date;
}

const metrics = ref<Metric[]>([
  { label: 'Revenue', value: 0, change: 0, trend: 'up' },
  { label: 'Users', value: 0, change: 0, trend: 'up' },
  { label: 'Orders', value: 0, change: 0, trend: 'down' },
  { label: 'Conversion', value: 0, change: 0, trend: 'up' },
]);

const activities = ref<Activity[]>([]);
const loading = ref(true);

// WebSocket connection
const { data: wsData, connect, disconnect } = useWebSocket('wss://api.example.com/dashboard');

// Update metrics when WebSocket data arrives
watch(wsData, (newData) => {
  if (newData?.type === 'metrics') {
    metrics.value = newData.metrics;
  } else if (newData?.type === 'activity') {
    activities.value.unshift(newData.activity);
    if (activities.value.length > 10) {
      activities.value.pop();
    }
  }
});

const totalRevenue = computed(() => {
  return metrics.value.find((m) => m.label === 'Revenue')?.value || 0;
});

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value);
};

const formatRelativeTime = (date: Date) => {
  const now = new Date();
  const diff = now.getTime() - new Date(date).getTime();
  const minutes = Math.floor(diff / 60000);

  if (minutes < 1) return 'Just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
};

onMounted(async () => {
  try {
    const res = await fetch('/api/dashboard/initial');
    const data = await res.json();
    metrics.value = data.metrics;
    activities.value = data.activities;
    connect();
  } finally {
    loading.value = false;
  }
});

onUnmounted(() => {
  disconnect();
});
</script>

<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-7xl mx-auto">
      <h1 class="text-3xl font-bold mb-8">Dashboard</h1>

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>

      <template v-else>
        <!-- Metrics Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div
            v-for="metric in metrics"
            :key="metric.label"
            class="bg-white rounded-lg shadow p-6"
          >
            <div class="flex justify-between items-start">
              <div>
                <p class="text-sm text-gray-600 mb-1">{{ metric.label }}</p>
                <p class="text-2xl font-bold">
                  {{ metric.label === 'Revenue' ? formatCurrency(metric.value) : metric.value }}
                </p>
              </div>
              <span
                :class="[
                  'text-xs font-semibold px-2 py-1 rounded',
                  metric.trend === 'up' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                ]"
              >
                {{ metric.trend === 'up' ? '↑' : '↓' }} {{ Math.abs(metric.change) }}%
              </span>
            </div>
          </div>
        </div>

        <!-- Activity Feed -->
        <div class="bg-white rounded-lg shadow">
          <div class="px-6 py-4 border-b">
            <h2 class="text-xl font-semibold">Recent Activity</h2>
          </div>
          <div class="divide-y">
            <div
              v-for="activity in activities"
              :key="activity.id"
              class="px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition"
            >
              <div class="flex items-center gap-4">
                <div class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white font-semibold">
                  {{ activity.user.charAt(0).toUpperCase() }}
                </div>
                <div>
                  <p class="font-medium">{{ activity.user }}</p>
                  <p class="text-sm text-gray-600">{{ activity.action }}</p>
                </div>
              </div>
              <span class="text-xs text-gray-500">
                {{ formatRelativeTime(activity.timestamp) }}
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
```

---

## Example 6: Angular Shopping Cart

### Cart Service with RxJS

```typescript
// cart.service.ts
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { map } from 'rxjs/operators';

export interface CartItem {
  id: string;
  name: string;
  price: number;
  quantity: number;
  image: string;
}

@Injectable({ providedIn: 'root' })
export class CartService {
  private itemsSubject = new BehaviorSubject<CartItem[]>([]);
  public items$ = this.itemsSubject.asObservable();

  public total$: Observable<number> = this.items$.pipe(
    map((items) => items.reduce((sum, item) => sum + item.price * item.quantity, 0))
  );

  public itemCount$: Observable<number> = this.items$.pipe(
    map((items) => items.reduce((sum, item) => sum + item.quantity, 0))
  );

  constructor() {
    // Load cart from localStorage
    const savedCart = localStorage.getItem('cart');
    if (savedCart) {
      this.itemsSubject.next(JSON.parse(savedCart));
    }

    // Save cart on changes
    this.items$.subscribe((items) => {
      localStorage.setItem('cart', JSON.stringify(items));
    });
  }

  addItem(item: Omit<CartItem, 'quantity'>): void {
    const currentItems = this.itemsSubject.value;
    const existingItem = currentItems.find((i) => i.id === item.id);

    if (existingItem) {
      this.updateQuantity(item.id, existingItem.quantity + 1);
    } else {
      this.itemsSubject.next([...currentItems, { ...item, quantity: 1 }]);
    }
  }

  removeItem(id: string): void {
    const currentItems = this.itemsSubject.value;
    this.itemsSubject.next(currentItems.filter((item) => item.id !== id));
  }

  updateQuantity(id: string, quantity: number): void {
    if (quantity <= 0) {
      this.removeItem(id);
      return;
    }

    const currentItems = this.itemsSubject.value;
    this.itemsSubject.next(
      currentItems.map((item) => (item.id === id ? { ...item, quantity } : item))
    );
  }

  clearCart(): void {
    this.itemsSubject.next([]);
  }
}
```

### Cart Component

```typescript
// cart.component.ts
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CartService, CartItem } from './cart.service';
import { Observable } from 'rxjs';

@Component({
  selector: 'app-cart',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="cart-container">
      <h2>Shopping Cart</h2>

      <div *ngIf="(itemCount$ | async) === 0" class="empty-cart">
        <p>Your cart is empty</p>
      </div>

      <div *ngIf="(itemCount$ | async)! > 0">
        <div class="cart-items">
          <div *ngFor="let item of items$ | async" class="cart-item">
            <img [src]="item.image" [alt]="item.name" class="item-image" />

            <div class="item-details">
              <h3>{{ item.name }}</h3>
              <p class="item-price">\${{ item.price.toFixed(2) }}</p>
            </div>

            <div class="item-quantity">
              <button
                (click)="decrementQuantity(item)"
                class="quantity-btn"
              >
                -
              </button>
              <span>{{ item.quantity }}</span>
              <button
                (click)="incrementQuantity(item)"
                class="quantity-btn"
              >
                +
              </button>
            </div>

            <div class="item-subtotal">
              \${{ (item.price * item.quantity).toFixed(2) }}
            </div>

            <button (click)="removeItem(item.id)" class="remove-btn">
              Remove
            </button>
          </div>
        </div>

        <div class="cart-summary">
          <div class="summary-row">
            <span>Subtotal ({{ itemCount$ | async }} items):</span>
            <span class="total-price">\${{ (total$ | async)!.toFixed(2) }}</span>
          </div>
          <button (click)="checkout()" class="checkout-btn">
            Proceed to Checkout
          </button>
          <button (click)="clearCart()" class="clear-btn">
            Clear Cart
          </button>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .cart-container {
      max-width: 800px;
      margin: 0 auto;
      padding: 20px;
    }

    .cart-item {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 16px;
      border-bottom: 1px solid #eee;
    }

    .item-image {
      width: 80px;
      height: 80px;
      object-fit: cover;
      border-radius: 8px;
    }

    .item-details {
      flex: 1;
    }

    .item-quantity {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .quantity-btn {
      width: 32px;
      height: 32px;
      border: 1px solid #ddd;
      border-radius: 4px;
      background: white;
      cursor: pointer;
    }

    .quantity-btn:hover {
      background: #f5f5f5;
    }

    .cart-summary {
      margin-top: 24px;
      padding: 20px;
      background: #f9f9f9;
      border-radius: 8px;
    }

    .checkout-btn {
      width: 100%;
      padding: 12px;
      background: #007bff;
      color: white;
      border: none;
      border-radius: 6px;
      font-weight: 600;
      cursor: pointer;
      margin-top: 16px;
    }

    .checkout-btn:hover {
      background: #0056b3;
    }
  `]
})
export class CartComponent {
  items$: Observable<CartItem[]>;
  total$: Observable<number>;
  itemCount$: Observable<number>;

  constructor(private cartService: CartService) {
    this.items$ = this.cartService.items$;
    this.total$ = this.cartService.total$;
    this.itemCount$ = this.cartService.itemCount$;
  }

  incrementQuantity(item: CartItem): void {
    this.cartService.updateQuantity(item.id, item.quantity + 1);
  }

  decrementQuantity(item: CartItem): void {
    this.cartService.updateQuantity(item.id, item.quantity - 1);
  }

  removeItem(id: string): void {
    this.cartService.removeItem(id);
  }

  clearCart(): void {
    if (confirm('Are you sure you want to clear your cart?')) {
      this.cartService.clearCart();
    }
  }

  checkout(): void {
    // Navigate to checkout page
    console.log('Proceeding to checkout...');
  }
}
```

---

## Example 7: React Infinite Scroll with Intersection Observer

```typescript
import { useState, useEffect, useRef, useCallback } from 'react';

interface Post {
  id: number;
  title: string;
  body: string;
  author: string;
}

function useInfiniteScroll(
  fetchMore: (page: number) => Promise<Post[]>,
  hasMore: boolean
) {
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const observerRef = useRef<IntersectionObserver | null>(null);

  const lastElementRef = useCallback(
    (node: HTMLDivElement | null) => {
      if (loading) return;
      if (observerRef.current) observerRef.current.disconnect();

      observerRef.current = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting && hasMore) {
          setPage((prev) => prev + 1);
        }
      });

      if (node) observerRef.current.observe(node);
    },
    [loading, hasMore]
  );

  useEffect(() => {
    setLoading(true);
    fetchMore(page).finally(() => setLoading(false));
  }, [page]);

  return { loading, lastElementRef };
}

export default function InfiniteScrollFeed() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [hasMore, setHasMore] = useState(true);

  const fetchPosts = async (page: number): Promise<Post[]> => {
    const res = await fetch(`/api/posts?page=${page}&limit=10`);
    const data = await res.json();

    if (data.length === 0) {
      setHasMore(false);
      return [];
    }

    setPosts((prev) => [...prev, ...data]);
    return data;
  };

  const { loading, lastElementRef } = useInfiniteScroll(fetchPosts, hasMore);

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-8">Feed</h1>

      <div className="space-y-4">
        {posts.map((post, index) => {
          const isLastElement = posts.length === index + 1;

          return (
            <div
              key={post.id}
              ref={isLastElement ? lastElementRef : null}
              className="border rounded-lg p-6 bg-white shadow-sm hover:shadow-md transition"
            >
              <h2 className="text-xl font-semibold mb-2">{post.title}</h2>
              <p className="text-gray-600 mb-3">{post.body}</p>
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-sm font-semibold">
                  {post.author.charAt(0).toUpperCase()}
                </div>
                <span className="text-sm text-gray-700">{post.author}</span>
              </div>
            </div>
          );
        })}
      </div>

      {loading && (
        <div className="flex justify-center my-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )}

      {!hasMore && (
        <p className="text-center text-gray-500 my-8">No more posts to load</p>
      )}
    </div>
  );
}
```

---

## Example 8: Vue Multi-Step Form

```vue
<script setup lang="ts">
import { ref, computed } from 'vue';
import { useForm } from 'vee-validate';
import * as yup from 'yup';

interface FormData {
  // Step 1
  firstName: string;
  lastName: string;
  email: string;
  // Step 2
  address: string;
  city: string;
  zipCode: string;
  // Step 3
  cardNumber: string;
  expiryDate: string;
  cvv: string;
}

const currentStep = ref(1);
const totalSteps = 3;

const step1Schema = yup.object({
  firstName: yup.string().required('First name is required'),
  lastName: yup.string().required('Last name is required'),
  email: yup.string().required('Email is required').email('Invalid email'),
});

const step2Schema = yup.object({
  address: yup.string().required('Address is required'),
  city: yup.string().required('City is required'),
  zipCode: yup.string().required('ZIP code is required').matches(/^\d{5}$/, 'Invalid ZIP'),
});

const step3Schema = yup.object({
  cardNumber: yup.string().required('Card number is required').matches(/^\d{16}$/, 'Invalid card number'),
  expiryDate: yup.string().required('Expiry date is required').matches(/^\d{2}\/\d{2}$/, 'Format: MM/YY'),
  cvv: yup.string().required('CVV is required').matches(/^\d{3}$/, 'Invalid CVV'),
});

const currentSchema = computed(() => {
  switch (currentStep.value) {
    case 1: return step1Schema;
    case 2: return step2Schema;
    case 3: return step3Schema;
    default: return step1Schema;
  }
});

const { handleSubmit, errors, values } = useForm({
  validationSchema: currentSchema,
});

const progress = computed(() => (currentStep.value / totalSteps) * 100);

const nextStep = handleSubmit(() => {
  if (currentStep.value < totalSteps) {
    currentStep.value++;
  }
});

const prevStep = () => {
  if (currentStep.value > 1) {
    currentStep.value--;
  }
};

const submitForm = handleSubmit(async (formData) => {
  console.log('Submitting form:', formData);
  // Submit to API
  await fetch('/api/checkout', {
    method: 'POST',
    body: JSON.stringify(formData),
  });
});
</script>

<template>
  <div class="max-w-2xl mx-auto p-6">
    <h1 class="text-3xl font-bold mb-8">Checkout</h1>

    <!-- Progress Bar -->
    <div class="mb-8">
      <div class="flex justify-between mb-2">
        <span class="text-sm font-medium">Step {{ currentStep }} of {{ totalSteps }}</span>
        <span class="text-sm text-gray-600">{{ Math.round(progress) }}% Complete</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2">
        <div
          class="bg-blue-600 h-2 rounded-full transition-all duration-300"
          :style="{ width: `${progress}%` }"
        ></div>
      </div>
    </div>

    <form @submit.prevent="currentStep === totalSteps ? submitForm() : nextStep()">
      <!-- Step 1: Personal Info -->
      <div v-show="currentStep === 1" class="space-y-4">
        <h2 class="text-xl font-semibold mb-4">Personal Information</h2>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">First Name</label>
            <Field name="firstName" type="text" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="firstName" class="text-red-500 text-sm" />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">Last Name</label>
            <Field name="lastName" type="text" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="lastName" class="text-red-500 text-sm" />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Email</label>
          <Field name="email" type="email" class="w-full border rounded px-3 py-2" />
          <ErrorMessage name="email" class="text-red-500 text-sm" />
        </div>
      </div>

      <!-- Step 2: Address -->
      <div v-show="currentStep === 2" class="space-y-4">
        <h2 class="text-xl font-semibold mb-4">Shipping Address</h2>

        <div>
          <label class="block text-sm font-medium mb-1">Address</label>
          <Field name="address" type="text" class="w-full border rounded px-3 py-2" />
          <ErrorMessage name="address" class="text-red-500 text-sm" />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">City</label>
            <Field name="city" type="text" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="city" class="text-red-500 text-sm" />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">ZIP Code</label>
            <Field name="zipCode" type="text" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="zipCode" class="text-red-500 text-sm" />
          </div>
        </div>
      </div>

      <!-- Step 3: Payment -->
      <div v-show="currentStep === 3" class="space-y-4">
        <h2 class="text-xl font-semibold mb-4">Payment Information</h2>

        <div>
          <label class="block text-sm font-medium mb-1">Card Number</label>
          <Field name="cardNumber" type="text" placeholder="1234 5678 9012 3456" class="w-full border rounded px-3 py-2" />
          <ErrorMessage name="cardNumber" class="text-red-500 text-sm" />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">Expiry Date</label>
            <Field name="expiryDate" type="text" placeholder="MM/YY" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="expiryDate" class="text-red-500 text-sm" />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1">CVV</label>
            <Field name="cvv" type="text" placeholder="123" class="w-full border rounded px-3 py-2" />
            <ErrorMessage name="cvv" class="text-red-500 text-sm" />
          </div>
        </div>
      </div>

      <!-- Navigation Buttons -->
      <div class="flex justify-between mt-8">
        <button
          type="button"
          @click="prevStep"
          v-show="currentStep > 1"
          class="px-6 py-2 border rounded-lg hover:bg-gray-50"
        >
          Back
        </button>

        <button
          type="submit"
          class="ml-auto px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          {{ currentStep === totalSteps ? 'Complete Order' : 'Next' }}
        </button>
      </div>
    </form>
  </div>
</template>
```

---

**Last Updated**: 2025-10-22
**Frameworks**: React 19.0, Vue 3.5, Angular 19.0
