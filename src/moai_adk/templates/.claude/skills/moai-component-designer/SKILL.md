---
name: moai-component-designer
version: 4.0.0
description: Enterprise-grade component design and architecture with React 19, Vue
  3.5, Atomic Design patterns, accessibility standards (WCAG 2.1 AA), design systems,
  and production-ready component libraries. Use when building reusable component systems,
  designing component libraries, implementing Atomic Design, ensuring accessibility
  compliance, or creating scalable UI architectures.
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
status: stable
updated: '2025-11-18'
stability: stable
---


# Enterprise Component Designer & Accessible UI Architecture

## Enterprise Component Capabilities

**Modern Framework Integration**:
- React 19 with server components and async rendering
- Vue 3.5 with composition API and reactive systems
- Svelte 5 with rune-based reactivity
- Solid.js for fine-grained reactivity

**Atomic Design & Component Systems**:
- Atoms, Molecules, Organisms architecture
- Design tokens and theme systems
- Component composition patterns
- Storybook 8.0 integration

**Accessibility Standards**:
- WCAG 2.1 AA compliance
- ARIA patterns and semantics
- Keyboard navigation
- Screen reader optimization
- Focus management
- Color contrast validation

**Production Design Systems**:
- Material Design 3
- Chakra UI patterns
- Shadcn/UI architecture
- Radix UI primitives

## Skill Metadata
| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Frameworks** | React 19, Vue 3.5, Svelte 5, Solid.js |
| **Design Systems** | Material Design 3, Chakra, Shadcn, Radix |
| **Accessibility** | WCAG 2.1 AA, ARIA patterns |
| **Tools** | Storybook 8.0, Figma, Chromatic |
| **Tier** | **4 (Enterprise)** |

## Component Architecture Deep Dive

### Level 1: Atomic Foundation

**Atoms** (Basic UI elements):
- Button component with variants (solid, outline, ghost, link)
- Input field with validation states
- Label with accessibility attributes
- Icon wrapper with sizing
- Badge with semantic colors
- Text variants (heading, body, caption)

**Atomic Design Principles**:
1. Pure presentational (no business logic)
2. Single responsibility
3. Highly reusable
4. Type-safe with TypeScript
5. Accessibility first
6. Visual regression tested

### Level 2: Molecules (Component Combinations)

**Molecular Components**:
- Form group (label + input + validation feedback)
- Card with header/footer/actions
- Navigation menu item
- Search input with autocomplete
- Modal with focus trap
- Tooltip with positioning
- Dropdown with keyboard support
- Progress bar with labels
- Alert box with actions
- Tag input with removal

**Composition Strategies**:
- Props composition for flexibility
- Render props pattern
- Compound components (Dialog.Trigger, Dialog.Content)
- Slots/composition API (Vue)
- Context-based shared state

### Level 3: Organisms (Complex UI Sections)

**Organism Components**:
- Data table (sorting, filtering, pagination)
- Form with multi-step validation
- Navigation header with dropdown
- Sidebar layout with responsive collapse
- Dialog/Modal systems with overlay
- Breadcrumb navigation
- Pagination component
- Tabs interface
- Accordion/Disclosure
- Tree view with keyboard nav

**Advanced Patterns**:
- Virtual scrolling for large lists
- Infinite scroll implementation
- Lazy loading with suspense
- Keyboard navigation matrix
- Focus management
- Accessible overlays

## Accessibility & WCAG 2.1 AA

### Keyboard Navigation

**Standard patterns**:
```
- Tab: Move to next focusable element
- Shift+Tab: Move to previous
- Enter: Activate button/submit
- Space: Toggle checkbox/radio
- Arrow keys: Navigate lists/menus
- Escape: Close modals/dropdowns
- Home/End: Jump to start/end
```

**Implementation**:
- tabindex management (0, -1)
- Focus visible styles
- Skip links
- Focus traps in modals
- Logical tab order

### Screen Reader Support

**Semantic HTML**:
- Use `<button>` for buttons (not `<div>`)
- Use `<label>` for form inputs
- Use `<nav>`, `<main>`, `<article>`
- Proper heading hierarchy (`<h1>` to `<h6>`)
- List elements for lists

**ARIA Attributes**:
- `aria-label`: Accessible name
- `aria-describedby`: Description link
- `aria-invalid`: Form error state
- `aria-live`: Dynamic content updates
- `aria-expanded`: Toggle states
- `aria-hidden`: Hide decorative elements
- `role`: Custom element semantics

**Live Regions**:
- Alert notifications
- Status updates
- Loading states
- Form validation feedback
- Search results

### Visual Accessibility

**Color Contrast**:
- 4.5:1 for normal text (WCAG AA)
- 3:1 for large text (18pt+)
- Color + icon/pattern (not color alone)
- Test with tools: WebAIM, Stark

**Focus Indicators**:
- Minimum 2px visible focus ring
- Sufficient contrast
- Not removed with CSS reset
- Consistent styling

**Responsive Design**:
- Touch targets minimum 44x44px
- Text sizing support
- Zoom compatibility (200%)
- Mobile-first approach
- Orientation support

**Motion & Animation**:
- Respect `prefers-reduced-motion`
- Avoid flashing content
- Auto-play disabled
- Pause controls for video/animation

## Design System Architecture

### Token System

**Color Tokens**:
```
Primary, Secondary, Success, Warning, Error, Info
Light & Dark variants
Semantic naming (on-primary, background, surface)
WCAG contrast pre-validated
```

**Spacing Scale**:
```
Base unit: 4px
Scale: 4, 8, 12, 16, 24, 32, 48, 64, 96
Padding, margin, gaps
Responsive breakpoints
Touch-friendly (min 44px)
```

**Typography**:
```
Font families: system, serif, monospace
Sizes: 12px to 64px scale
Line height: 1.2 to 1.8 ratios
Font weights: 400, 500, 600, 700
Letter spacing adjustments
```

**Motion & Animation**:
```
Short: 150ms (hover, focus)
Medium: 300ms (transitions)
Long: 500ms+ (entrances, exits)
Easing: ease-out, ease-in-out, ease-in
Respect prefers-reduced-motion
```

### Component State Model

**Universal States**:
- Default/Rest: Initial state
- Hover: Mouse over
- Focus: Keyboard focus
- Active: Pressed/selected
- Disabled: Non-interactive
- Loading: Async operation
- Error: Validation failure
- Success: Action completed

**State Implementation**:
- Data attributes (`data-state`)
- CSS classes
- CSS variables
- Component state (React, Vue)
- Compound component pattern

## React 19 Examples

### 1. Button Component
- Variants: solid, outline, ghost, link
- Sizes: xs, sm, md, lg, xl
- Colors: primary, secondary, success, error
- States: default, hover, focus, disabled, loading
- Type-safe with TypeScript generics
- Accessibility: aria-label, aria-disabled, role

### 2. Form Input System
- Text, email, password, number, date, select
- Real-time validation feedback
- Error message display
- Helper text support
- Label integration with for/id
- Accessibility: aria-describedby, aria-invalid, aria-required

### 3. Modal Dialog
- Focus trap with FocusLock library
- Keyboard escape handling
- Scroll lock on body
- ARIA dialog pattern
- Smooth enter/exit animations
- Backdrop click handling

### 4. Data Table
- Column sorting with indicators
- Row filtering
- Pagination controls
- Responsive horizontal scroll
- Selection checkboxes
- Keyboard navigation (arrow keys)
- Accessibility: caption, thead/tbody, aria-sort, aria-label

### 5. Tabs Component
- Keyboard navigation (arrow keys, Home, End)
- Accessible names with aria-label
- Active state indication
- ARIA tabs pattern
- Lazy loading support
- Customizable indicators

## Vue 3.5 Composition API Examples

### 1. useButton Composable
- Variant and size handling
- State management (hover, focus, active)
- Event handling
- Type-safe props
- Return reactive button state

### 2. useForm Composable
- Form field registration
- Validation orchestration
- Error tracking
- Dirty/touched states
- Submit handling
- Reset functionality

### 3. useModal Composable
- Open/close state
- Focus trap management
- Keyboard escape handling
- Ref forwarding
- Animated enter/exit
- Custom close handlers

### 4. useDataTable Composable
- Sorting management
- Filtering logic
- Pagination state
- Selection handling
- Keyboard navigation
- Performance optimization

### 5. useAccessibility Composable
- Focus management
- ARIA attribute generation
- Keyboard event handling
- Live region updates
- Screen reader announcements

## Storybook 8.0 Integration

**Story Organization**:
```
- Atoms (basic components)
- Molecules (combinations)
- Organisms (complex UI)
- Templates (page layouts)
- Pages (full examples)
```

**Story Best Practices**:
1. One story per component state
2. Use controls for interactive props
3. Include accessibility notes
4. Link to design tokens
5. Document accessibility requirements
6. Visual regression setup
7. Performance monitoring
8. Component API documentation

**Accessibility Testing in Storybook**:
- Axe accessibility addon
- Color contrast checker
- Keyboard navigation testing
- Screen reader annotation

## Production Considerations

### Performance

**Component optimization**:
- React.memo for pure components
- Lazy loading with Suspense
- Virtual scrolling for lists (1000+ items)
- Code splitting by route
- Tree shaking unused code
- Bundle size monitoring

**Runtime performance**:
- Avoid inline objects/functions
- Batch state updates
- Use keys in lists
- Optimize re-renders
- Use web workers for heavy computation

### Testing Strategy

**Test coverage targets**:
- Unit tests: >90% (component logic)
- Integration tests: Interaction flows
- Visual regression: Storybook + Chromatic
- Accessibility: axe testing
- E2E: Critical user journeys
- Performance: Lighthouse + Web Vitals

### Documentation

**Component documentation**:
- Usage examples
- Props API table
- Accessibility requirements
- Design tokens used
- Related components
- Common patterns
- Migration guides
- Troubleshooting

### Deployment

**Release process**:
- Semantic versioning
- Changelog generation
- Breaking changes documentation
- Migration guides
- Design token versioning
- Component API stability

---

**Enterprise Checklist**:
- [ ] 100% TypeScript coverage
- [ ] WCAG 2.1 AA compliance verified (axe testing)
- [ ] All components keyboard navigable
- [ ] Screen reader tested
- [ ] <50KB gzipped library
- [ ] >95% Lighthouse accessibility score
- [ ] <100ms interaction response time
- [ ] >95% test coverage
- [ ] Design tokens documented
- [ ] Storybook stories complete
