# SVG Basics Module

Comprehensive guide to SVG document structure, coordinate system, shapes, paths, and text elements.

---

## Document Structure

### The SVG Element

The root svg element defines the canvas for vector graphics. Required attributes vary by context.

For standalone SVG files, xmlns is required:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
  <!-- content -->
</svg>
```

For inline HTML5, xmlns can be omitted:

```html
<svg viewBox="0 0 100 100" width="100" height="100">
  <!-- content -->
</svg>
```

Key attributes:

- xmlns: Namespace declaration, required for standalone files
- viewBox: Coordinate system definition "minX minY width height"
- width/height: Rendered dimensions, can use CSS units
- preserveAspectRatio: How content scales to fit viewport

### The g Element

Groups related elements together. All child elements inherit presentation attributes.

```xml
<g id="button-group" fill="blue" stroke="black" stroke-width="2">
  <rect x="10" y="10" width="80" height="30" rx="5"/>
  <text x="50" y="30" text-anchor="middle">Click</text>
</g>
```

Transform attribute moves, rotates, or scales the entire group:

```xml
<g transform="translate(50, 50) rotate(45) scale(1.5)">
  <!-- all children transformed together -->
</g>
```

### The defs Element

Container for reusable elements that are not rendered directly. Define gradients, patterns, filters, symbols, and markers here.

```xml
<defs>
  <linearGradient id="grad1">
    <stop offset="0%" stop-color="red"/>
    <stop offset="100%" stop-color="blue"/>
  </linearGradient>

  <filter id="blur1">
    <feGaussianBlur stdDeviation="5"/>
  </filter>

  <symbol id="icon1" viewBox="0 0 24 24">
    <path d="M12 2L2 22h20z"/>
  </symbol>
</defs>
```

### The symbol Element

Defines a reusable graphic template with its own viewBox. More powerful than plain g elements because symbols can have independent coordinate systems.

```xml
<defs>
  <symbol id="star" viewBox="0 0 100 100">
    <polygon points="50,5 61,40 98,40 68,62 79,97 50,75 21,97 32,62 2,40 39,40"/>
  </symbol>
</defs>

<!-- Use at different sizes -->
<use href="#star" x="0" y="0" width="20" height="20"/>
<use href="#star" x="30" y="0" width="50" height="50"/>
<use href="#star" x="100" y="0" width="100" height="100"/>
```

### The use Element

References and renders content defined elsewhere. Can reference symbols, groups, or any element with an id.

```xml
<use href="#my-shape" x="10" y="10" width="50" height="50"/>
```

Attributes:

- href: Reference to element id (use xlink:href for older browsers)
- x, y: Position offset for the referenced content
- width, height: Size constraints (only meaningful for symbols)

Styling limitations: Cannot directly style internal elements of referenced content. Use CSS custom properties or inherit from use element.

```xml
<style>
  .icon-use { --icon-color: blue; }
</style>

<symbol id="icon">
  <path fill="var(--icon-color, black)" d="..."/>
</symbol>

<use href="#icon" class="icon-use"/>
```

---

## Coordinate System

### ViewBox Explained

The viewBox attribute defines the internal coordinate system. Format: "min-x min-y width height"

```xml
<!-- 100x100 internal coordinates, rendered at 200x200 pixels -->
<svg viewBox="0 0 100 100" width="200" height="200">
  <!-- coordinates 0-100 map to 0-200 pixels -->
</svg>
```

ViewBox enables resolution-independent graphics. Content scales to fit the rendered width/height while maintaining the coordinate system.

Negative min values shift the origin:

```xml
<!-- Origin centered in the viewport -->
<svg viewBox="-50 -50 100 100" width="200" height="200">
  <circle cx="0" cy="0" r="40"/> <!-- centered circle -->
</svg>
```

### preserveAspectRatio

Controls how content fits when viewBox aspect ratio differs from viewport.

Format: "alignment meetOrSlice"

Alignment values combine X and Y positioning:

- xMin: Left edge
- xMid: Horizontal center
- xMax: Right edge
- YMin: Top edge
- YMid: Vertical center
- YMax: Bottom edge

Common combinations:

```xml
<!-- Center and fit entirely (default) -->
<svg viewBox="0 0 100 100" preserveAspectRatio="xMidYMid meet">

<!-- Center and fill, cropping excess -->
<svg viewBox="0 0 100 100" preserveAspectRatio="xMidYMid slice">

<!-- Stretch to fill, ignoring aspect ratio -->
<svg viewBox="0 0 100 100" preserveAspectRatio="none">

<!-- Align to top-left corner -->
<svg viewBox="0 0 100 100" preserveAspectRatio="xMinYMin meet">
```

### Nested SVG Elements

SVG elements can be nested, each with its own viewBox and coordinate system.

```xml
<svg viewBox="0 0 200 200" width="400" height="400">
  <!-- Main content in 200x200 space -->
  <rect x="0" y="0" width="200" height="200" fill="#eee"/>

  <!-- Nested SVG with different coordinate system -->
  <svg x="50" y="50" width="100" height="100" viewBox="0 0 50 50">
    <!-- This content uses 50x50 coordinates -->
    <circle cx="25" cy="25" r="20" fill="blue"/>
  </svg>
</svg>
```

---

## Basic Shapes

### Rectangle (rect)

```xml
<!-- Basic rectangle -->
<rect x="10" y="10" width="80" height="50"/>

<!-- Rounded corners -->
<rect x="10" y="10" width="80" height="50" rx="10" ry="5"/>

<!-- Square (equal width/height) -->
<rect x="10" y="10" width="50" height="50"/>

<!-- Rounded square (circle-like corners) -->
<rect x="10" y="10" width="50" height="50" rx="10"/>
```

Attributes:

- x, y: Top-left corner position (default: 0)
- width, height: Dimensions (required for visibility)
- rx: Horizontal corner radius
- ry: Vertical corner radius (defaults to rx if omitted)

### Circle

```xml
<!-- Basic circle -->
<circle cx="50" cy="50" r="40"/>

<!-- Positioned circle -->
<circle cx="100" cy="75" r="25" fill="red"/>
```

Attributes:

- cx, cy: Center position (default: 0)
- r: Radius (required for visibility)

### Ellipse

```xml
<!-- Basic ellipse -->
<ellipse cx="50" cy="50" rx="40" ry="25"/>

<!-- Vertical ellipse -->
<ellipse cx="50" cy="50" rx="20" ry="40"/>
```

Attributes:

- cx, cy: Center position (default: 0)
- rx: Horizontal radius
- ry: Vertical radius

### Line

```xml
<!-- Basic line -->
<line x1="10" y1="10" x2="90" y2="90" stroke="black" stroke-width="2"/>

<!-- Horizontal line -->
<line x1="0" y1="50" x2="100" y2="50" stroke="gray"/>

<!-- Vertical line -->
<line x1="50" y1="0" x2="50" y2="100" stroke="gray"/>
```

Lines require stroke attribute to be visible (no fill).

Attributes:

- x1, y1: Start point
- x2, y2: End point

### Polyline

Connected line segments that do not close automatically.

```xml
<!-- Open shape -->
<polyline points="10,10 50,50 90,10" fill="none" stroke="black"/>

<!-- Zigzag pattern -->
<polyline points="0,50 25,20 50,50 75,20 100,50" fill="none" stroke="blue"/>

<!-- With fill (creates closed appearance) -->
<polyline points="10,80 50,20 90,80" fill="lightblue" stroke="blue"/>
```

Points format: space or comma-separated coordinate pairs.

### Polygon

Connected line segments that close automatically.

```xml
<!-- Triangle -->
<polygon points="50,10 90,90 10,90"/>

<!-- Pentagon -->
<polygon points="50,5 95,35 80,90 20,90 5,35"/>

<!-- Star -->
<polygon points="50,5 61,40 98,40 68,62 79,97 50,75 21,97 32,62 2,40 39,40"/>
```

---

## Path Element

The most powerful and flexible shape element. Uses a mini-language of commands in the d attribute.

### Command Basics

Uppercase commands use absolute coordinates. Lowercase commands use relative offsets from current position.

```xml
<!-- Absolute: move to (10,10), line to (90,90) -->
<path d="M10 10 L90 90"/>

<!-- Relative: move to (10,10), line by (80,80) -->
<path d="M10 10 l80 80"/>
```

### Move Commands (M, m)

Start a new subpath at specified position.

```xml
<!-- Single path, multiple subpaths -->
<path d="M10 10 L50 10  M10 30 L50 30  M10 50 L50 50"/>
```

### Line Commands (L, l, H, h, V, v)

Draw straight lines.

```xml
<!-- Line to absolute position -->
<path d="M10 10 L90 50"/>

<!-- Horizontal line to x=80 -->
<path d="M10 50 H80"/>

<!-- Vertical line to y=90 -->
<path d="M50 10 V90"/>

<!-- Relative movements -->
<path d="M10 10 l40 40 h30 v-20"/>
```

### Close Command (Z, z)

Close current subpath by drawing line to starting point.

```xml
<!-- Closed triangle -->
<path d="M50 10 L90 90 L10 90 Z"/>

<!-- Multiple closed shapes -->
<path d="M10 10 L40 10 L40 40 L10 40 Z  M60 10 L90 10 L90 40 L60 40 Z"/>
```

### Cubic Bezier (C, c, S, s)

Smooth curves with two control points.

```xml
<!-- Basic cubic bezier -->
<path d="M10 80 C30 10, 70 10, 90 80"/>

<!-- S command continues smoothly from previous curve -->
<path d="M10 50 C20 20, 40 20, 50 50 S80 80, 90 50"/>
```

C command parameters: x1 y1 (first control), x2 y2 (second control), x y (end point)

S command parameters: x2 y2 (second control), x y (end point). First control is reflection of previous.

### Quadratic Bezier (Q, q, T, t)

Curves with single control point.

```xml
<!-- Basic quadratic bezier -->
<path d="M10 80 Q50 10, 90 80"/>

<!-- T command continues smoothly -->
<path d="M10 50 Q30 10, 50 50 T90 50"/>
```

Q command parameters: x1 y1 (control point), x y (end point)

T command parameters: x y (end point). Control point is reflection of previous.

### Arc Command (A, a)

Elliptical arc segments.

```xml
<!-- Basic arc -->
<path d="M10 50 A40 40 0 0 1 90 50"/>

<!-- Half circle -->
<path d="M10 50 A40 40 0 1 0 90 50"/>
```

Parameters: rx ry rotation large-arc-flag sweep-flag x y

- rx, ry: Ellipse radii
- rotation: X-axis rotation in degrees
- large-arc-flag: 0 for smaller arc, 1 for larger arc
- sweep-flag: 0 for counter-clockwise, 1 for clockwise
- x, y: End point

---

## Text Elements

### Basic Text

```xml
<text x="50" y="50">Hello World</text>
```

Positioning:

- x, y: Baseline position (y is baseline, not top)
- dx, dy: Offset from calculated position
- text-anchor: Horizontal alignment (start, middle, end)
- dominant-baseline: Vertical alignment (auto, middle, hanging)

```xml
<!-- Centered text -->
<text x="100" y="50" text-anchor="middle" dominant-baseline="middle">
  Centered
</text>
```

### Text Styling

```xml
<text x="10" y="50"
      font-family="Arial, sans-serif"
      font-size="24"
      font-weight="bold"
      font-style="italic"
      fill="#333"
      letter-spacing="2"
      text-decoration="underline">
  Styled Text
</text>
```

### tspan Element

Inline span for styling portions of text.

```xml
<text x="10" y="50" font-size="16">
  Normal text
  <tspan fill="red" font-weight="bold">highlighted</tspan>
  and more text
</text>
```

Position adjustments with tspan:

```xml
<text x="10" y="50">
  First line
  <tspan x="10" dy="20">Second line</tspan>
  <tspan x="10" dy="20">Third line</tspan>
</text>
```

### textPath Element

Render text along a path shape.

```xml
<defs>
  <path id="curve" d="M10 80 Q95 10 180 80" fill="none"/>
</defs>

<text font-size="14">
  <textPath href="#curve">
    Text following a curved path
  </textPath>
</text>
```

Attributes:

- href: Reference to path element
- startOffset: Starting position along path (percentage or length)
- method: How text conforms to path (align, stretch)
- spacing: Character spacing (auto, exact)

```xml
<!-- Centered on path -->
<text>
  <textPath href="#curve" startOffset="50%" text-anchor="middle">
    Centered text
  </textPath>
</text>
```
