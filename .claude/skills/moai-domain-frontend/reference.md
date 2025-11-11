# moai-domain-frontend — Reference Documentation

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-12  
**Status**: Production Ready

---

## 📚 Framework & Library Versions (2025)

### **React Ecosystem**
- **React**: 19.2.0
  - Server Components (stable)
  - Server Actions (stable)
  - React Compiler (stable)
  - useFormState, useOptimistic hooks
- **Next.js**: 15.1.8
  - App Router (stable)
  - Turbopack (stable)
  - Server Actions (stable)
  - Parallel routes, intercepting routes

### **Vue Ecosystem**
- **Vue**: 3.5+
  - Composition API (recommended)
  - TypeScript integration
  - Suspense (stable)
  - Teleport (stable)
- **Pinia**: 3.x
  - Official Vue state management
  - TypeScript support
  - DevTools integration

### **State Management**
- **Zustand**: 5.x
  - Lightweight (< 1KB)
  - Middleware: devtools, persist
- **Jotai**: 3.x
  - Atomic state model
  - Fine-grained reactivity
- **Pinia**: 3.x (Vue)
  - Composition API style
  - TypeScript support

### **Build Tools**
- **Vite**: 6.x
- **Turbopack**: (Next.js 15 default)
- **esbuild**: 0.24+

### **Testing**
- **Vitest**: Latest
- **Testing Library**: Latest
- **Playwright**: Latest

---

## 🔍 Official Documentation Links

### **React 19**
- [React Official Docs](https://react.dev/)
- [Server Components Guide](https://react.dev/reference/rsc/server-components)
- [React Compiler](https://react.dev/learn/react-compiler)
- [useFormState Hook](https://react.dev/reference/react-dom/hooks/useFormState)
- [useOptimistic Hook](https://react.dev/reference/react/useOptimistic)

### **Next.js 15**
- [Next.js Documentation](https://nextjs.org/docs)
- [App Router Guide](https://nextjs.org/docs/app)
- [Server Actions & Mutations](https://nextjs.org/docs/app/building-your-application/data-fetching/server-actions-and-mutations)
- [Parallel Routes](https://nextjs.org/docs/app/building-your-application/routing/parallel-routes)
- [Intercepting Routes](https://nextjs.org/docs/app/building-your-application/routing/intercepting-routes)
- [Streaming & Suspense](https://nextjs.org/docs/app/building-your-application/routing/loading-ui-and-streaming)

### **Vue 3.5**
- [Vue 3 Documentation](https://vuejs.org/)
- [Composition API](https://vuejs.org/guide/extras/composition-api-faq)
- [TypeScript with Composition API](https://vuejs.org/guide/typescript/composition-api)
- [Suspense](https://vuejs.org/guide/built-ins/suspense)
- [Teleport](https://vuejs.org/guide/built-ins/teleport)

### **State Management**
- [Zustand Documentation](https://zustand-demo.pmnd.rs/)
- [Jotai Documentation](https://jotai.org/)
- [Pinia Documentation](https://pinia.vuejs.org/)

### **Accessibility**
- [WCAG 2.2 Guidelines](https://www.w3.org/WAI/WCAG22/quickref/)
- [ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM Resources](https://webaim.org/)

---

## 🎯 Key Concepts

### **React 19 Server Components**

**What They Are**:
- Components that run exclusively on the server
- No client-side JavaScript shipped
- Direct access to databases, APIs, environment variables

**When to Use**:
- Data fetching from databases/APIs
- Content-heavy pages (blogs, dashboards)
- SEO-critical pages
- Pages with sensitive logic (auth, payments)

**When NOT to Use**:
- Interactive components (useState, useEffect)
- Browser APIs (localStorage, window)
- Event handlers (onClick, onChange)

### **Server Actions**

**What They Are**:
- Server-side functions callable from client components
- Replace traditional API routes
- Progressive enhancement support

**Benefits**:
- Zero API route boilerplate
- Type-safe server calls
- Built-in validation & error handling
- Works without JavaScript (forms)

**Best Practices**:
- Use for form submissions
- Implement server-side validation
- Call `revalidatePath()` after mutations
- Return structured error responses

### **State Management Decision Matrix**

| Scenario | Recommended Tool | Why |
|----------|------------------|-----|
| Simple global state | Zustand | Lightweight, easy to use |
| Complex state relationships | Jotai | Atomic model, fine-grained reactivity |
| Vue application | Pinia | Official, Composition API style |
| Rare updates, small state | Context API | Built-in, no dependencies |
| Form state | React Hook Form | Optimized for forms |

---

## 🚀 Performance Benchmarks

### **Core Web Vitals Targets (2025)**

| Metric | Target | Threshold |
|--------|--------|-----------|
| **LCP** (Largest Contentful Paint) | < 2.5s | Good |
| **INP** (Interaction to Next Paint) | < 200ms | Good |
| **CLS** (Cumulative Layout Shift) | < 0.1 | Good |

### **Bundle Size Targets**

| Resource | Target | Max |
|----------|--------|-----|
| Initial JS | < 150KB | 200KB |
| Total JS | < 350KB | 500KB |
| CSS | < 50KB | 100KB |
| Images (per page) | < 500KB | 1MB |

### **Optimization Strategies**

1. **Code Splitting**
   - Lazy load routes
   - Dynamic imports for heavy components
   - Split vendor bundles

2. **Image Optimization**
   - Use Next.js `<Image />` or `next/image`
   - WebP/AVIF formats
   - Responsive images
   - Lazy loading

3. **Caching**
   - Static assets: 1 year cache
   - API responses: Appropriate cache headers
   - Service Workers (PWA)

---

## ♿ Accessibility Quick Reference

### **Semantic HTML Elements**

```html
<header>   <!-- Page/section header -->
<nav>      <!-- Navigation links -->
<main>     <!-- Main content -->
<article>  <!-- Self-contained content -->
<section>  <!-- Thematic grouping -->
<aside>    <!-- Complementary content -->
<footer>   <!-- Footer content -->
```

### **ARIA Roles (Common)**

| Role | Use Case |
|------|----------|
| `button` | Interactive button (non-semantic) |
| `dialog` | Modal dialog |
| `listbox` | Dropdown menu |
| `menu` | Menu widget |
| `tab`, `tabpanel` | Tab interface |
| `alert` | Important message |
| `status` | Status update (polite) |

### **ARIA States & Properties**

| Attribute | Purpose |
|-----------|---------|
| `aria-label` | Accessible name |
| `aria-describedby` | Additional description |
| `aria-expanded` | Expanded/collapsed state |
| `aria-selected` | Selected state |
| `aria-disabled` | Disabled state |
| `aria-invalid` | Validation error |
| `aria-required` | Required field |
| `aria-live` | Live region updates |

### **Keyboard Navigation Standards**

| Key | Action |
|-----|--------|
| `Tab` | Focus next element |
| `Shift + Tab` | Focus previous element |
| `Enter` | Activate button/link |
| `Space` | Activate button |
| `Escape` | Close modal/menu |
| `Arrow keys` | Navigate lists/menus |

---

## 🧪 Testing Strategies

### **Test Types**

1. **Unit Tests** (Vitest)
   - Test individual components
   - Test custom hooks
   - Test utility functions

2. **Integration Tests** (Testing Library)
   - Test component interactions
   - Test form submissions
   - Test API integrations

3. **E2E Tests** (Playwright)
   - Test complete user flows
   - Test across browsers
   - Test mobile experiences

### **Testing Best Practices**

```typescript
// ✅ Good: Test behavior, not implementation
test('adds product to cart', async () => {
  render(<ShoppingCart />)
  
  await userEvent.click(screen.getByRole('button', { name: /add to cart/i }))
  
  expect(screen.getByText(/1 item in cart/i)).toBeInTheDocument()
})

// ❌ Bad: Testing internal state
test('updates state', () => {
  const { result } = renderHook(() => useCart())
  
  act(() => {
    result.current.addItem(mockProduct)
  })
  
  expect(result.current.items).toHaveLength(1)
})
```

---

## 🔒 Security Checklist

### **Input Validation**
- ✅ Validate all user input server-side
- ✅ Use schema validation (Zod, Yup)
- ✅ Sanitize HTML content (DOMPurify)
- ✅ Escape user-generated content

### **Authentication & Authorization**
- ✅ Use secure session management
- ✅ Implement CSRF protection
- ✅ Use httpOnly cookies for tokens
- ✅ Validate permissions server-side

### **API Security**
- ✅ Never expose API keys in client code
- ✅ Use Server Actions for sensitive operations
- ✅ Implement rate limiting
- ✅ Validate request origins (CORS)

### **Content Security Policy (CSP)**
```typescript
// next.config.js
const ContentSecurityPolicy = `
  default-src 'self';
  script-src 'self' 'unsafe-eval' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' blob: data: https:;
  font-src 'self';
  connect-src 'self' https://api.example.com;
`

export default {
  async headers() {
    return [{
      source: '/:path*',
      headers: [
        {
          key: 'Content-Security-Policy',
          value: ContentSecurityPolicy.replace(/\s{2,}/g, ' ').trim()
        }
      ]
    }]
  }
}
```

---

## 📊 Monitoring & Analytics

### **Key Metrics to Track**

1. **Performance**
   - Core Web Vitals (LCP, INP, CLS)
   - Time to First Byte (TTFB)
   - First Contentful Paint (FCP)

2. **User Behavior**
   - Page views
   - User flows
   - Conversion rates
   - Bounce rates

3. **Errors**
   - JavaScript errors
   - Network failures
   - Failed API calls
   - Render errors

### **Recommended Tools**

- **Performance**: Lighthouse, WebPageTest, Chrome DevTools
- **Analytics**: Google Analytics 4, Plausible, PostHog
- **Error Tracking**: Sentry, Rollbar, Bugsnag
- **Real User Monitoring**: Vercel Analytics, Cloudflare Web Analytics

---

## 🔄 Migration Guides

### **React 18 → React 19**

**Key Changes**:
1. Server Components are now stable
2. React Compiler replaces manual memoization
3. New hooks: `useFormState`, `useOptimistic`
4. `<form>` action prop supports Server Actions

**Migration Steps**:
1. Update dependencies: `npm install react@19 react-dom@19`
2. Identify components that can be Server Components
3. Remove unnecessary `useMemo`, `useCallback`, `React.memo`
4. Refactor forms to use Server Actions
5. Test thoroughly

### **Next.js Pages Router → App Router**

**Key Changes**:
1. File-based routing with `app/` directory
2. `page.tsx` replaces `pages/*.tsx`
3. `layout.tsx` for shared layouts
4. Server Components by default
5. Server Actions replace API routes

**Migration Steps**:
1. Create `app/` directory alongside `pages/`
2. Move routes incrementally
3. Convert data fetching to Server Components
4. Replace API routes with Server Actions
5. Update navigation to use `next/navigation`

### **Vue Options API → Composition API**

**Key Changes**:
1. `<script setup>` replaces `export default {}`
2. `ref()` and `reactive()` replace `data()`
3. Lifecycle hooks: `onMounted()` replaces `mounted()`
4. Props: `defineProps()` replaces `props` option
5. Emits: `defineEmits()` replaces `emits` option

**Migration Steps**:
1. Convert one component at a time
2. Use `<script setup lang="ts">`
3. Replace reactive data with `ref()` or `reactive()`
4. Convert methods to regular functions
5. Convert computed properties to `computed()`
6. Update lifecycle hooks

---

## 🎓 Learning Resources

### **Courses**
- [React Official Tutorial](https://react.dev/learn)
- [Next.js Learn](https://nextjs.org/learn)
- [Vue Mastery](https://www.vuemastery.com/)
- [Epic React](https://epicreact.dev/) by Kent C. Dodds

### **Books**
- "Learning React" by Alex Banks & Eve Porcello
- "Full Stack React" by Accomazzo et al.
- "Vue.js 3 Design Patterns" by Carlos Rodrigues

### **Blogs & Newsletters**
- [React Newsletter](https://reactnewsletter.com/)
- [This Week in React](https://thisweekinreact.com/)
- [Vue.js News](https://news.vuejs.org/)
- [Josh Comeau's Blog](https://www.joshwcomeau.com/)

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (frontend-expert)
