# Component Designer Reference

## Official Resources

### React 19 & Component Libraries
- React 19 Docs: https://react.dev (Server Components, Actions, Hooks)
- Shadcn/UI: https://ui.shadcn.com (Accessible component library)
- Chakra UI: https://chakra-ui.com (Component system with accessibility)
- Radix UI: https://www.radix-ui.com (Unstyled accessible primitives)

### Vue 3 & Composition API
- Vue 3 Docs: https://vuejs.org (Composition API, Reactivity)
- HeadlessUI Vue: https://headlessui.com (Unstyled, accessible components)
- VueUse: https://vueuse.org (Composition utilities)
- Pinia: https://pinia.vuejs.org (State management)

### Atomic Design
- Brad Frost's Atomic Design: https://atomicdesign.bradfrost.com
- Component-Driven Development: https://www.componentdriven.org

### Accessibility Standards
- WCAG 2.1 Official: https://www.w3.org/WAI/WCAG21/quickref/
- ARIA Authoring Guide: https://www.w3.org/WAI/ARIA/apg/
- WebAIM Resources: https://webaim.org/
- MDN Accessibility: https://developer.mozilla.org/en-US/docs/Web/Accessibility

### Design System Tools
- Storybook 8: https://storybook.js.org (Component documentation)
- Chromatic: https://chromatic.com (Visual regression testing)
- Figma: https://figma.com (Design tool)
- Tokens Studio: https://tokens.studio (Design tokens)

## Key Concepts & Patterns

### Component Composition Patterns

**Prop Composition**:
```
Primary method for component flexibility
Pass data and callbacks as props
Type-safe with TypeScript
Shallow prop drilling issues
```

**Render Props Pattern**:
```
Pass functions as children
Access component internals
Function as children pattern
Supports complex state sharing
```

**Compound Components**:
```
Dialog.Root > Dialog.Trigger + Dialog.Content
Tab.Root > Tab.List > Tab.Item + Tab.Panel
Menu.Root > Menu.Trigger > Menu.Content > Menu.Item
Provides implicit coupling through context
```

**Context-Based State**:
```
Theme context
Layout context
Form context
Accessibility context
Performance optimized with memo
```

### Keyboard Navigation Patterns

**Tab Navigation**:
```
Move focus to next interactive element
Shift+Tab for reverse direction
tabindex: 0 for custom elements
tabindex: -1 for programmatic focus
Logical tab order (left→right, top→bottom)
```

**Menu/Dropdown Navigation**:
```
Arrow Up/Down: Move within menu
Home/End: Jump to start/end
Enter: Select item
Escape: Close menu
Type to search: First letter jump
```

**Data Table Navigation**:
```
Arrow keys: Navigate cells
Page Up/Down: Navigate rows
Home/End: Jump to columns
Ctrl+Home/End: Jump to edges
Tab: Skip to interactive elements
```

### ARIA Implementation Patterns

**Buttons with Icons**:
```html
<!-- Text button (no aria-label needed) -->
<button>Send Message</button>

<!-- Icon-only button (aria-label required) -->
<button aria-label="Send message">
  <IconSend />
</button>

<!-- Button with icon and text -->
<button>
  <IconSend />
  Send
</button>
```

**Dropdown Menu**:
```html
<button aria-haspopup="true" aria-expanded="false">
  Menu
</button>
<ul role="menu">
  <li role="menuitem">Option 1</li>
  <li role="menuitem" aria-disabled="true">Option 2</li>
</ul>
```

**Tabs**:
```html
<div role="tablist">
  <button role="tab" aria-selected="true" aria-controls="panel1">
    Tab 1
  </button>
  <button role="tab" aria-selected="false" aria-controls="panel2">
    Tab 2
  </button>
</div>
<div id="panel1" role="tabpanel">Content 1</div>
<div id="panel2" role="tabpanel">Content 2</div>
```

## Performance Optimization Techniques

### React Optimization
- React.memo for pure components
- useMemo for expensive calculations
- useCallback for stable function refs
- Lazy loading with React.lazy + Suspense
- Code splitting by route
- Image optimization (webp, responsive sizes)
- Bundle analysis (webpack-bundle-analyzer)

### Vue Optimization
- v-once for static content
- v-memo for static subtrees
- Computed properties for reactive caching
- Lazy components with defineAsyncComponent
- Virtual scrolling for long lists
- Batch updates with nextTick
- Bundle analysis (rollup-plugin-visualizer)

### CSS Optimization
- CSS-in-JS: Emotion, Styled Components
- Utility-first CSS: Tailwind (PurgeCSS)
- CSS Modules: Scoped styles
- Critical CSS extraction
- Unused CSS removal
- Media query optimization
- Animation performance (transform, opacity)

## Design System Tokens

### Color System
```
Primary: #007AFF (iOS blue)
Secondary: #5AC8FA (light blue)
Success: #4CD964 (green)
Warning: #FF9500 (orange)
Error: #FF3B30 (red)
Info: #00C7BE (teal)

Light mode: #FFFFFF background
Dark mode: #1A1A1A background

Contrast: 4.5:1 (normal), 3:1 (large)
```

### Spacing Scale
```
0: 0
1: 4px
2: 8px
3: 12px
4: 16px
6: 24px
8: 32px
12: 48px
16: 64px
20: 80px
24: 96px
```

### Typography
```
Display Large: 56px, 1.2 line height
Display: 45px, 1.2 line height
Headline Large: 32px, 1.3 line height
Headline: 28px, 1.4 line height
Title Large: 22px, 1.3 line height
Title: 16px, 1.5 line height
Body Large: 16px, 1.5 line height
Body: 14px, 1.5 line height
Label Large: 12px, 1.2 line height
```

### Breakpoints
```
Mobile: 0px - 639px
Tablet: 640px - 1023px
Desktop: 1024px+

Orientation: portrait, landscape
Touch: hover disabled
```

## Testing Strategies

### Unit Testing
- Jest for test framework
- React Testing Library (behavior-focused)
- Vue Test Utils (component testing)
- Component interaction testing
- Props validation
- Event handler testing

### Integration Testing
- Multi-component flows
- Form submission with validation
- Data fetching and display
- State synchronization
- Context providers
- Navigation flows

### Accessibility Testing
- axe DevTools (automated)
- Manual keyboard navigation
- Screen reader testing (NVDA, JAWS, VoiceOver)
- Color contrast checking
- Focus order validation
- ARIA attribute verification

### Visual Regression Testing
- Chromatic (visual testing)
- Percy (visual regression)
- Storybook visual tests
- Screenshot comparison
- Responsive design verification
- Component state variations

## Common Pitfalls & Solutions

### Accessibility
- Pitfall: Using `<div>` as `<button>` without ARIA
- Solution: Use semantic HTML elements (button, input, label)

- Pitfall: Color as only indicator
- Solution: Add text, icons, or patterns

- Pitfall: Ignoring keyboard users
- Solution: Test with Tab, Arrow keys, Enter, Escape

### Performance
- Pitfall: Inline objects/functions in render
- Solution: Use useMemo/useCallback or move outside

- Pitfall: Large bundle sizes
- Solution: Code splitting, tree shaking, minification

- Pitfall: Unnecessary re-renders
- Solution: Use memo, useMemo, shouldComponentUpdate

### Developer Experience
- Pitfall: Complex component props
- Solution: Use compound components or context

- Pitfall: No TypeScript types
- Solution: Full TS coverage, exported types

- Pitfall: Poor documentation
- Solution: Storybook stories, inline JSDoc, examples

---

**Last Updated**: November 2025
**Components**: 50+ enterprise-grade components
**Test Coverage**: >95%
**Bundle Size**: <50KB gzipped
