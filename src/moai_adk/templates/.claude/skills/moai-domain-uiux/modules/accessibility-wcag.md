name: moai-foundation-uiux-accessibility
description: WCAG 2.2 compliance, testing, and keyboard navigation

## WCAG 2.2 Accessibility Implementation

### Color Contrast Validation

```typescript
// utils/a11y/contrast.ts
/
 * Calculate relative luminance for WCAG compliance
 * @see https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html
 */
function getLuminance(rgb: [number, number, number]): number {
 const [r, g, b] = rgb.map(val => {
 const sRGB = val / 255;
 return sRGB <= 0.03928
 ? sRGB / 12.92
 : Math.pow((sRGB + 0.055) / 1.055, 2.4);
 });
 return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/
 * Calculate contrast ratio between two colors
 * WCAG AA: 4.5:1 (normal text), 3:1 (large text)
 * WCAG AAA: 7:1 (normal text), 4.5:1 (large text)
 */
export function getContrastRatio(
 foreground: string,
 background: string
): number {
 const fgLum = getLuminance(hexToRgb(foreground));
 const bgLum = getLuminance(hexToRgb(background));
 const lighter = Math.max(fgLum, bgLum);
 const darker = Math.min(fgLum, bgLum);
 return (lighter + 0.05) / (darker + 0.05);
}

/
 * Check if color pair meets WCAG AA/AAA requirements
 */
export function meetsWCAG(
 foreground: string,
 background: string,
 level: 'AA' | 'AAA' = 'AA',
 isLargeText: boolean = false
): boolean {
 const ratio = getContrastRatio(foreground, background);
 
 if (level === 'AAA') {
 return isLargeText ? ratio >= 4.5 : ratio >= 7;
 }
 
 // AA level
 return isLargeText ? ratio >= 3 : ratio >= 4.5;
}
```

### Keyboard Navigation

```typescript
// hooks/useKeyboardNavigation.ts
import { useEffect, useRef } from 'react';

export function useKeyboardNavigation<T extends HTMLElement>(
 options: {
 onEscape?: () => void;
 onEnter?: () => void;
 trapFocus?: boolean;
 } = {}
) {
 const elementRef = useRef<T>(null);

 useEffect(() => {
 const element = elementRef.current;
 if (!element) return;

 const handleKeyDown = (e: KeyboardEvent) => {
 if (e.key === 'Escape') {
 options.onEscape?.();
 } else if (e.key === 'Enter') {
 options.onEnter?.();
 } else if (e.key === 'Tab' && options.trapFocus) {
 // Focus trap implementation
 const focusableElements = element.querySelectorAll<HTMLElement>(
 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
 );
 const firstElement = focusableElements[0];
 const lastElement = focusableElements[focusableElements.length - 1];

 if (e.shiftKey && document.activeElement === firstElement) {
 lastElement.focus();
 e.preventDefault();
 } else if (!e.shiftKey && document.activeElement === lastElement) {
 firstElement.focus();
 e.preventDefault();
 }
 }
 };

 element.addEventListener('keydown', handleKeyDown);
 return () => element.removeEventListener('keydown', handleKeyDown);
 }, [options]);

 return elementRef;
}
```

### Motion Accessibility (Reduced Motion)

```css
/* styles/motion.css */
@media (prefers-reduced-motion: reduce) {
 *,
 *::before,
 *::after {
 animation-duration: 0.01ms !important;
 animation-iteration-count: 1 !important;
 transition-duration: 0.01ms !important;
 scroll-behavior: auto !important;
 }
}

/* Safe animations for reduced motion users */
.fade-enter {
 opacity: 0;
}

.fade-enter-active {
 opacity: 1;
 transition: opacity 200ms ease-in;
}

@media (prefers-reduced-motion: reduce) {
 .fade-enter-active {
 transition: none;
 opacity: 1;
 }
}
```

### Accessibility Testing Automation

Jest + jest-axe Configuration:

```typescript
// tests/setup.ts
import '@testing-library/jest-dom';
import { toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);
```

Component Accessibility Tests:

```typescript
// components/atoms/Button/Button.test.tsx
import { render } from '@testing-library/react';
import { axe } from 'jest-axe';
import { Button } from './Button';

describe('Button Accessibility', () => {
 it('should have no accessibility violations', async () => {
 const { container } = render(<Button>Click me</Button>);
 const results = await axe(container);
 expect(results).toHaveNoViolations();
 });

 it('should have correct ARIA attributes when disabled', () => {
 const { getByRole } = render(<Button disabled>Disabled</Button>);
 const button = getByRole('button');
 expect(button).toHaveAttribute('aria-disabled', 'true');
 });

 it('should indicate loading state to screen readers', () => {
 const { getByRole } = render(<Button isLoading>Loading</Button>);
 const button = getByRole('button');
 expect(button).toHaveAttribute('aria-busy', 'true');
 });
});
```

### Screen Reader Best Practices

ARIA Labels and Descriptions:

```typescript
// Example: Accessible modal
export function Modal({ title, children, onClose }) {
 const titleId = useId();
 const descriptionId = useId();

 return (
 <div
 role="dialog"
 aria-modal="true"
 aria-labelledby={titleId}
 aria-describedby={descriptionId}
 >
 <h2 id={titleId}>{title}</h2>
 <div id={descriptionId}>{children}</div>
 <button onClick={onClose} aria-label="Close dialog">
 ×
 </button>
 </div>
 );
}
```

Live Regions for Dynamic Content:

```typescript
// Announce status updates
export function StatusAnnouncer({ message }: { message: string }) {
 return (
 <div
 role="status"
 aria-live="polite"
 aria-atomic="true"
 className="sr-only"
 >
 {message}
 </div>
 );
}

// Alert for critical updates
export function Alert({ message }: { message: string }) {
 return (
 <div
 role="alert"
 aria-live="assertive"
 aria-atomic="true"
 >
 {message}
 </div>
 );
}
```

### Testing Checklist

Manual Testing:
- [ ] Keyboard navigation (Tab, Shift+Tab, Enter, Escape)
- [ ] Screen reader testing (NVDA, JAWS, VoiceOver)
- [ ] Color contrast verification (4.5:1 minimum)
- [ ] Focus indicators visible and clear
- [ ] Reduced motion preference respected

Automated Testing:
- [ ] jest-axe for component accessibility
- [ ] Storybook a11y addon for visual testing
- [ ] Chromatic for visual regression
- [ ] CI/CD integration for continuous testing

---

Last Updated: 2025-11-26
Related: [Main Skill](../SKILL.md), [Component Architecture](component-architecture.md)
