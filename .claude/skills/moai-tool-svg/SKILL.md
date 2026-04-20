---
name: moai-tool-svg
description: >
  SVG creation, optimization, and transformation specialist. Use when creating vector
  graphics, optimizing SVG files with SVGO, implementing icon systems, building data
  visualizations, or adding SVG animations.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(svgo:*) Bash(npx:*) WebFetch
user-invocable: false
metadata:
  version: "1.0.0"
  category: "tool"
  modularized: "true"
  status: "active"
  updated: "2026-01-26"
  tags: "svg, vector, graphics, svgo, optimization, animation, icons"
  related-skills: "moai-domain-frontend, moai-docs-generation"
  context7-libraries: "/nicolo-ribaudo/svgo"
---

# SVG Creation and Optimization Specialist

Comprehensive SVG development covering vector graphics creation, SVGO optimization, icon systems, data visualizations, and animations. This skill provides patterns for all SVG workflows from basic shapes to complex animated graphics.

---

## Quick Reference (30 seconds)

### Basic SVG Template

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
  <title>Accessible Title</title>
  <desc>Description for screen readers</desc>
  <!-- Content here -->
</svg>
```

### Common Shapes Cheatsheet

Rectangle: `<rect x="10" y="10" width="80" height="60" rx="5" />`

Circle: `<circle cx="50" cy="50" r="40" />`

Ellipse: `<ellipse cx="50" cy="50" rx="40" ry="25" />`

Line: `<line x1="10" y1="10" x2="90" y2="90" stroke="black" />`

Polyline: `<polyline points="10,10 50,50 90,10" fill="none" stroke="black" />`

Polygon: `<polygon points="50,10 90,90 10,90" />`

### Path Commands Quick Reference

Movement Commands:
- M x y: Move to absolute position
- m dx dy: Move relative
- L x y: Line to absolute
- l dx dy: Line relative
- H x: Horizontal line absolute
- h dx: Horizontal line relative
- V y: Vertical line absolute
- v dy: Vertical line relative
- Z: Close path

Curve Commands:
- C x1 y1 x2 y2 x y: Cubic bezier (two control points)
- S x2 y2 x y: Smooth cubic (reflects previous control)
- Q x1 y1 x y: Quadratic bezier (one control point)
- T x y: Smooth quadratic (reflects previous control)
- A rx ry rotation large-arc sweep x y: Arc

### SVGO CLI Commands

Install globally: `npm install -g svgo`

Optimize single file: `svgo input.svg -o output.svg`

Optimize directory: `svgo -f ./src/icons -o ./dist/icons`

Show optimization stats: `svgo input.svg --pretty --indent=2`

Use config file: `svgo input.svg --config svgo.config.mjs`

### Fill and Stroke Quick Reference

Fill properties: fill, fill-opacity, fill-rule (nonzero, evenodd)

Stroke properties: stroke, stroke-width, stroke-opacity, stroke-linecap (butt, round, square), stroke-linejoin (miter, round, bevel), stroke-dasharray, stroke-dashoffset

---

## Implementation Guide (5 minutes)

### SVG Document Structure

The SVG element requires the xmlns attribute for standalone files. The viewBox defines the coordinate system as "minX minY width height". Width and height set the rendered size.

```xml
<svg xmlns="http://www.w3.org/2000/svg"
     viewBox="0 0 200 200"
     width="200" height="200"
     preserveAspectRatio="xMidYMid meet">

  <!-- Reusable definitions -->
  <defs>
    <linearGradient id="grad1">
      <stop offset="0%" stop-color="#ff0000" />
      <stop offset="100%" stop-color="#0000ff" />
    </linearGradient>
  </defs>

  <!-- Grouped content -->
  <g id="main-group" transform="translate(10, 10)">
    <rect width="100" height="100" fill="url(#grad1)" />
  </g>
</svg>
```

### Creating Reusable Symbols

Symbols define reusable graphics that can be instantiated with use elements. They support their own viewBox for scaling.

```xml
<svg xmlns="http://www.w3.org/2000/svg">
  <defs>
    <symbol id="icon-star" viewBox="0 0 24 24">
      <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
    </symbol>
  </defs>

  <!-- Use the symbol multiple times -->
  <use href="#icon-star" x="0" y="0" width="24" height="24" />
  <use href="#icon-star" x="30" y="0" width="24" height="24" fill="gold" />
  <use href="#icon-star" x="60" y="0" width="48" height="48" />
</svg>
```

### Path Creation Patterns

Simple icon path combining moves, lines, and curves:

```xml
<path d="M10 20 L20 10 L30 20 L20 30 Z" />
```

Rounded rectangle using arcs:

```xml
<path d="M15 5 H85 A10 10 0 0 1 95 15 V85 A10 10 0 0 1 85 95 H15 A10 10 0 0 1 5 85 V15 A10 10 0 0 1 15 5 Z" />
```

Heart shape using cubic beziers:

```xml
<path d="M50 88 C20 65 5 45 5 30 A15 15 0 0 1 35 30 Q50 45 50 45 Q50 45 65 30 A15 15 0 0 1 95 30 C95 45 80 65 50 88 Z" />
```

### Gradient Implementation

Linear gradient from left to right:

```xml
<defs>
  <linearGradient id="horizontal-grad" x1="0%" y1="0%" x2="100%" y2="0%">
    <stop offset="0%" stop-color="#3498db" />
    <stop offset="50%" stop-color="#9b59b6" />
    <stop offset="100%" stop-color="#e74c3c" />
  </linearGradient>
</defs>
<rect fill="url(#horizontal-grad)" width="200" height="100" />
```

Radial gradient with focal point:

```xml
<defs>
  <radialGradient id="sphere-grad" cx="50%" cy="50%" r="50%" fx="30%" fy="30%">
    <stop offset="0%" stop-color="#ffffff" />
    <stop offset="100%" stop-color="#3498db" />
  </radialGradient>
</defs>
<circle fill="url(#sphere-grad)" cx="50" cy="50" r="40" />
```

### SVGO Configuration

Create svgo.config.mjs in project root:

```javascript
export default {
  multipass: true,
  plugins: [
    'preset-default',
    'prefixIds',
    {
      name: 'sortAttrs',
      params: {
        xmlnsOrder: 'alphabetical'
      }
    },
    {
      name: 'removeAttrs',
      params: {
        attrs: ['data-name', 'class']
      }
    }
  ]
};
```

Configuration preserving specific elements:

```javascript
export default {
  multipass: true,
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          removeViewBox: false,
          cleanupIds: {
            preserve: ['icon-', 'logo-']
          }
        }
      }
    }
  ]
};
```

### Embedding SVG in React

Inline SVG component:

```tsx
const Icon = ({ size = 24, color = 'currentColor' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
    <path d="M12 2L2 7l10 5 10-5-10-5z" stroke={color} strokeWidth="2" />
  </svg>
);
```

SVG sprite with use element:

```tsx
const SpriteIcon = ({ name, size = 24 }) => (
  <svg width={size} height={size}>
    <use href={`/sprites.svg#${name}`} />
  </svg>
);
```

### Text Elements

Basic text positioning:

```xml
<text x="50" y="50" text-anchor="middle" dominant-baseline="middle"
      font-family="Arial" font-size="16" fill="#333">
  Centered Text
</text>
```

Text on a path:

```xml
<defs>
  <path id="text-curve" d="M10 80 Q95 10 180 80" fill="none" />
</defs>
<text font-size="14">
  <textPath href="#text-curve">Text following a curved path</textPath>
</text>
```

---

## Advanced Implementation (10+ minutes)

### Complex Filter Effects

Drop shadow with blur:

```xml
<defs>
  <filter id="drop-shadow" x="-20%" y="-20%" width="140%" height="140%">
    <feGaussianBlur in="SourceAlpha" stdDeviation="3" result="blur" />
    <feOffset in="blur" dx="3" dy="3" result="offsetBlur" />
    <feMerge>
      <feMergeNode in="offsetBlur" />
      <feMergeNode in="SourceGraphic" />
    </feMerge>
  </filter>
</defs>
```

Glow effect:

```xml
<filter id="glow">
  <feGaussianBlur stdDeviation="4" result="coloredBlur" />
  <feMerge>
    <feMergeNode in="coloredBlur" />
    <feMergeNode in="SourceGraphic" />
  </feMerge>
</filter>
```

### Clipping and Masking

Clip path for cropping:

```xml
<defs>
  <clipPath id="circle-clip">
    <circle cx="50" cy="50" r="40" />
  </clipPath>
</defs>
<image href="photo.jpg" width="100" height="100" clip-path="url(#circle-clip)" />
```

Gradient mask for fade effect:

```xml
<defs>
  <linearGradient id="fade-grad">
    <stop offset="0%" stop-color="white" />
    <stop offset="100%" stop-color="black" />
  </linearGradient>
  <mask id="fade-mask">
    <rect width="100" height="100" fill="url(#fade-grad)" />
  </mask>
</defs>
<rect width="100" height="100" fill="blue" mask="url(#fade-mask)" />
```

### CSS Animation Integration

Keyframe animation for SVG elements:

```css
@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.1); opacity: 0.8; }
}

.animated-circle {
  animation: pulse 2s ease-in-out infinite;
  transform-origin: center;
}
```

Stroke drawing animation:

```css
.draw-path {
  stroke-dasharray: 1000;
  stroke-dashoffset: 1000;
  animation: draw 2s ease forwards;
}

@keyframes draw {
  to { stroke-dashoffset: 0; }
}
```

### Accessibility Best Practices

Always include title and desc for meaningful graphics:

```xml
<svg role="img" aria-labelledby="title desc">
  <title id="title">Company Logo</title>
  <desc id="desc">A blue mountain with snow-capped peak</desc>
  <!-- graphic content -->
</svg>
```

For decorative SVGs, hide from screen readers:

```xml
<svg aria-hidden="true" focusable="false">
  <!-- decorative content -->
</svg>
```

### Performance Optimization

Reduce precision in path data from 6 decimals to 2:

Before: `M10.123456 20.654321 L30.987654 40.123456`

After: `M10.12 20.65 L30.99 40.12`

Convert shapes to paths for smaller file size when appropriate. Use symbols for repeated elements. Apply SVGZ compression for 20-50% size reduction.

For detailed patterns on each topic, see the modules directory.

---

## Module Index

- modules/svg-basics.md: Document structure, coordinate system, shapes, paths, text
- modules/svg-styling.md: Fills, strokes, gradients, patterns, filters, clipping, masking
- modules/svg-optimization.md: SVGO configuration, compression, sprites, performance
- modules/svg-animation.md: CSS animations, SMIL, JavaScript, interaction patterns

---

## Works Well With

- moai-domain-frontend: React/Vue SVG component integration
- moai-docs-generation: SVG diagram generation for documentation
- moai-domain-uiux: Icon systems and design system integration
