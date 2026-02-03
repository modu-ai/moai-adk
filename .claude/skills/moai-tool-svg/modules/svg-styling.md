# SVG Styling Module

Comprehensive guide to fills, strokes, gradients, patterns, clipping, masking, and filter effects.

---

## Fill Properties

### Basic Fill

The fill attribute sets the interior color of shapes and text.

```xml
<!-- Solid color fill -->
<rect width="100" height="100" fill="blue"/>
<rect width="100" height="100" fill="#3498db"/>
<rect width="100" height="100" fill="rgb(52, 152, 219)"/>
<rect width="100" height="100" fill="rgba(52, 152, 219, 0.5)"/>

<!-- No fill (transparent interior) -->
<rect width="100" height="100" fill="none"/>

<!-- Current color (inherits from color property) -->
<g color="purple">
  <rect fill="currentColor" width="100" height="100"/>
</g>
```

### Fill Opacity

Separate from the element's overall opacity.

```xml
<!-- Semi-transparent fill -->
<rect width="100" height="100" fill="blue" fill-opacity="0.5"/>

<!-- Compare to overall opacity -->
<rect width="100" height="100" fill="blue" stroke="red" opacity="0.5"/>
```

### Fill Rule

Determines how complex paths with intersections are filled.

```xml
<!-- Default: nonzero (fills enclosed areas based on winding direction) -->
<path d="M50,0 L100,100 L0,100 Z M50,30 L80,80 L20,80 Z" fill-rule="nonzero"/>

<!-- evenodd (alternates fill based on crossing count) -->
<path d="M50,0 L100,100 L0,100 Z M50,30 L80,80 L20,80 Z" fill-rule="evenodd"/>
```

---

## Stroke Properties

### Basic Stroke

The stroke attribute sets the outline color.

```xml
<!-- Solid stroke -->
<rect width="100" height="100" fill="none" stroke="black"/>

<!-- Stroke width -->
<rect width="100" height="100" fill="none" stroke="black" stroke-width="5"/>

<!-- Stroke opacity -->
<rect width="100" height="100" fill="none" stroke="black" stroke-opacity="0.5"/>
```

### Stroke Linecap

Controls the appearance of line endings.

```xml
<!-- butt: Square end at exact endpoint (default) -->
<line x1="20" y1="20" x2="80" y2="20" stroke="black" stroke-width="10" stroke-linecap="butt"/>

<!-- round: Rounded end extending beyond endpoint -->
<line x1="20" y1="50" x2="80" y2="50" stroke="black" stroke-width="10" stroke-linecap="round"/>

<!-- square: Square end extending beyond endpoint -->
<line x1="20" y1="80" x2="80" y2="80" stroke="black" stroke-width="10" stroke-linecap="square"/>
```

### Stroke Linejoin

Controls the appearance of corners where lines meet.

```xml
<!-- miter: Sharp corners (default) -->
<polyline points="20,80 50,20 80,80" fill="none" stroke="black" stroke-width="10" stroke-linejoin="miter"/>

<!-- round: Rounded corners -->
<polyline points="20,80 50,20 80,80" fill="none" stroke="black" stroke-width="10" stroke-linejoin="round"/>

<!-- bevel: Beveled/flattened corners -->
<polyline points="20,80 50,20 80,80" fill="none" stroke="black" stroke-width="10" stroke-linejoin="bevel"/>
```

### Stroke Dash Pattern

Creates dashed or dotted lines.

```xml
<!-- Simple dashes: 10px dash, 5px gap -->
<line x1="10" y1="20" x2="190" y2="20" stroke="black" stroke-dasharray="10 5"/>

<!-- Dot pattern: 2px dash, 5px gap -->
<line x1="10" y1="40" x2="190" y2="40" stroke="black" stroke-width="2" stroke-linecap="round" stroke-dasharray="0.1 8"/>

<!-- Complex pattern: 15px dash, 5px gap, 5px dash, 5px gap -->
<line x1="10" y1="60" x2="190" y2="60" stroke="black" stroke-dasharray="15 5 5 5"/>

<!-- Offset starting position -->
<line x1="10" y1="80" x2="190" y2="80" stroke="black" stroke-dasharray="10 5" stroke-dashoffset="5"/>
```

### Stroke Animation with Dasharray

Classic line drawing effect:

```xml
<path d="M10 80 Q50 10, 90 80" fill="none" stroke="black" stroke-width="2"
      stroke-dasharray="200" stroke-dashoffset="200">
  <animate attributeName="stroke-dashoffset" from="200" to="0" dur="2s" fill="freeze"/>
</path>
```

---

## Linear Gradients

### Basic Linear Gradient

```xml
<defs>
  <linearGradient id="basic-grad">
    <stop offset="0%" stop-color="#3498db"/>
    <stop offset="100%" stop-color="#9b59b6"/>
  </linearGradient>
</defs>

<rect width="200" height="100" fill="url(#basic-grad)"/>
```

### Gradient Direction

Control direction with x1, y1, x2, y2 attributes (percentages or coordinates).

```xml
<!-- Horizontal (default: left to right) -->
<linearGradient id="h-grad" x1="0%" y1="0%" x2="100%" y2="0%">

<!-- Vertical (top to bottom) -->
<linearGradient id="v-grad" x1="0%" y1="0%" x2="0%" y2="100%">

<!-- Diagonal (top-left to bottom-right) -->
<linearGradient id="d-grad" x1="0%" y1="0%" x2="100%" y2="100%">

<!-- Reverse diagonal (bottom-left to top-right) -->
<linearGradient id="rd-grad" x1="0%" y1="100%" x2="100%" y2="0%">
```

### Multiple Color Stops

```xml
<linearGradient id="rainbow">
  <stop offset="0%" stop-color="#e74c3c"/>
  <stop offset="20%" stop-color="#f39c12"/>
  <stop offset="40%" stop-color="#f1c40f"/>
  <stop offset="60%" stop-color="#2ecc71"/>
  <stop offset="80%" stop-color="#3498db"/>
  <stop offset="100%" stop-color="#9b59b6"/>
</linearGradient>
```

### Stop Opacity

```xml
<linearGradient id="fade-grad">
  <stop offset="0%" stop-color="black" stop-opacity="1"/>
  <stop offset="100%" stop-color="black" stop-opacity="0"/>
</linearGradient>
```

### Spread Method

Controls behavior when gradient doesn't cover entire element.

```xml
<!-- pad: Extends edge colors (default) -->
<linearGradient id="pad-grad" spreadMethod="pad" x1="25%" x2="75%">

<!-- reflect: Mirrors gradient -->
<linearGradient id="reflect-grad" spreadMethod="reflect" x1="25%" x2="75%">

<!-- repeat: Tiles gradient -->
<linearGradient id="repeat-grad" spreadMethod="repeat" x1="25%" x2="75%">
```

### Gradient Units

```xml
<!-- objectBoundingBox (default): Percentages relative to element bounds -->
<linearGradient id="bbox-grad" gradientUnits="objectBoundingBox">

<!-- userSpaceOnUse: Absolute coordinates -->
<linearGradient id="user-grad" gradientUnits="userSpaceOnUse" x1="0" y1="0" x2="200" y2="0">
```

---

## Radial Gradients

### Basic Radial Gradient

```xml
<defs>
  <radialGradient id="basic-radial">
    <stop offset="0%" stop-color="white"/>
    <stop offset="100%" stop-color="blue"/>
  </radialGradient>
</defs>

<circle cx="100" cy="100" r="80" fill="url(#basic-radial)"/>
```

### Center and Radius

```xml
<!-- Centered gradient -->
<radialGradient id="centered" cx="50%" cy="50%" r="50%">

<!-- Off-center gradient -->
<radialGradient id="offset" cx="30%" cy="30%" r="70%">

<!-- Small radius (sharp gradient) -->
<radialGradient id="sharp" cx="50%" cy="50%" r="25%">
```

### Focal Point

Creates highlight effect when focal point differs from center.

```xml
<radialGradient id="sphere" cx="50%" cy="50%" r="50%" fx="30%" fy="30%">
  <stop offset="0%" stop-color="white"/>
  <stop offset="100%" stop-color="#3498db"/>
</radialGradient>
```

### Gradient Transform

Apply transformations to gradients.

```xml
<radialGradient id="ellipse-grad" gradientTransform="scale(1, 0.5)">
  <stop offset="0%" stop-color="white"/>
  <stop offset="100%" stop-color="green"/>
</radialGradient>
```

---

## Patterns

### Basic Pattern

```xml
<defs>
  <pattern id="dots" width="20" height="20" patternUnits="userSpaceOnUse">
    <circle cx="10" cy="10" r="5" fill="blue"/>
  </pattern>
</defs>

<rect width="200" height="200" fill="url(#dots)"/>
```

### Pattern Units

```xml
<!-- userSpaceOnUse: Pattern size in absolute coordinates -->
<pattern id="abs-pattern" width="20" height="20" patternUnits="userSpaceOnUse">

<!-- objectBoundingBox: Pattern size as percentage of element -->
<pattern id="rel-pattern" width="0.1" height="0.1" patternUnits="objectBoundingBox">
```

### Pattern Content Units

```xml
<!-- Content coordinates match patternUnits -->
<pattern id="pattern1" patternContentUnits="userSpaceOnUse">

<!-- Content uses objectBoundingBox coordinates -->
<pattern id="pattern2" patternContentUnits="objectBoundingBox">
```

### Stripe Pattern

```xml
<pattern id="stripes" width="10" height="10" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
  <rect width="5" height="10" fill="#3498db"/>
  <rect x="5" width="5" height="10" fill="white"/>
</pattern>
```

### Checkerboard Pattern

```xml
<pattern id="checkerboard" width="20" height="20" patternUnits="userSpaceOnUse">
  <rect width="10" height="10" fill="#333"/>
  <rect x="10" y="10" width="10" height="10" fill="#333"/>
  <rect x="10" width="10" height="10" fill="#fff"/>
  <rect y="10" width="10" height="10" fill="#fff"/>
</pattern>
```

---

## Clipping

### Basic Clip Path

Hard-edged masking that shows content inside the clip region.

```xml
<defs>
  <clipPath id="circle-clip">
    <circle cx="100" cy="100" r="80"/>
  </clipPath>
</defs>

<image href="photo.jpg" width="200" height="200" clip-path="url(#circle-clip)"/>
```

### Complex Clip Shapes

```xml
<clipPath id="star-clip">
  <polygon points="100,10 125,70 190,70 140,110 160,175 100,135 40,175 60,110 10,70 75,70"/>
</clipPath>
```

### Text Clip Path

```xml
<clipPath id="text-clip">
  <text x="100" y="100" font-size="80" font-weight="bold" text-anchor="middle">SVG</text>
</clipPath>

<image href="texture.jpg" width="200" height="200" clip-path="url(#text-clip)"/>
```

### Clip Path Units

```xml
<!-- userSpaceOnUse: Clip coordinates are absolute -->
<clipPath id="abs-clip" clipPathUnits="userSpaceOnUse">

<!-- objectBoundingBox: Clip coordinates relative to element (0-1) -->
<clipPath id="rel-clip" clipPathUnits="objectBoundingBox">
  <circle cx="0.5" cy="0.5" r="0.4"/>
</clipPath>
```

---

## Masking

### Basic Mask

Soft-edged masking using alpha channel. White areas reveal, black conceals, gray partially reveals.

```xml
<defs>
  <mask id="fade-mask">
    <rect width="200" height="200" fill="white"/>
    <rect width="200" height="200" fill="url(#fade-gradient)"/>
  </mask>
</defs>

<image href="photo.jpg" width="200" height="200" mask="url(#fade-mask)"/>
```

### Gradient Mask for Fade Effect

```xml
<defs>
  <linearGradient id="mask-gradient">
    <stop offset="0%" stop-color="white"/>
    <stop offset="100%" stop-color="black"/>
  </linearGradient>

  <mask id="horizontal-fade">
    <rect width="200" height="200" fill="url(#mask-gradient)"/>
  </mask>
</defs>

<rect width="200" height="200" fill="blue" mask="url(#horizontal-fade)"/>
```

### Radial Fade Mask

```xml
<defs>
  <radialGradient id="radial-mask-grad">
    <stop offset="0%" stop-color="white"/>
    <stop offset="100%" stop-color="black"/>
  </radialGradient>

  <mask id="spotlight">
    <rect width="200" height="200" fill="url(#radial-mask-grad)"/>
  </mask>
</defs>
```

### Mask vs Clip Path

Clip paths provide hard edges with binary visibility (inside or outside). Masks provide soft edges with gradient transparency. Use clip paths for geometric shapes, masks for fade effects.

---

## Filter Effects

### Filter Container

```xml
<filter id="my-filter" x="-20%" y="-20%" width="140%" height="140%">
  <!-- filter primitives -->
</filter>

<rect filter="url(#my-filter)" width="100" height="100"/>
```

Filter region (x, y, width, height) should extend beyond element bounds to accommodate effects like blur and shadow.

### Gaussian Blur

```xml
<filter id="blur">
  <feGaussianBlur in="SourceGraphic" stdDeviation="5"/>
</filter>

<!-- Different X and Y blur -->
<filter id="directional-blur">
  <feGaussianBlur stdDeviation="10 2"/>
</filter>
```

### Drop Shadow

```xml
<filter id="shadow">
  <feDropShadow dx="3" dy="3" stdDeviation="3" flood-color="rgba(0,0,0,0.5)"/>
</filter>
```

Manual drop shadow with more control:

```xml
<filter id="custom-shadow">
  <feGaussianBlur in="SourceAlpha" stdDeviation="3" result="blur"/>
  <feOffset in="blur" dx="5" dy="5" result="offsetBlur"/>
  <feFlood flood-color="rgba(0,0,0,0.5)" result="color"/>
  <feComposite in="color" in2="offsetBlur" operator="in" result="shadow"/>
  <feMerge>
    <feMergeNode in="shadow"/>
    <feMergeNode in="SourceGraphic"/>
  </feMerge>
</filter>
```

### Color Matrix

Transform colors using matrix multiplication.

```xml
<!-- Grayscale -->
<filter id="grayscale">
  <feColorMatrix type="matrix"
    values="0.33 0.33 0.33 0 0
            0.33 0.33 0.33 0 0
            0.33 0.33 0.33 0 0
            0    0    0    1 0"/>
</filter>

<!-- Saturate (0 = grayscale, 1 = normal, >1 = oversaturated) -->
<filter id="saturate">
  <feColorMatrix type="saturate" values="2"/>
</filter>

<!-- Hue rotation (degrees) -->
<filter id="hue-rotate">
  <feColorMatrix type="hueRotate" values="90"/>
</filter>
```

### Glow Effect

```xml
<filter id="glow">
  <feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur"/>
  <feMerge>
    <feMergeNode in="blur"/>
    <feMergeNode in="blur"/>
    <feMergeNode in="SourceGraphic"/>
  </feMerge>
</filter>
```

### Combining Filters

```xml
<filter id="combined">
  <!-- Step 1: Blur -->
  <feGaussianBlur in="SourceGraphic" stdDeviation="2" result="blurred"/>

  <!-- Step 2: Add noise -->
  <feTurbulence type="fractalNoise" baseFrequency="0.05" result="noise"/>

  <!-- Step 3: Combine -->
  <feBlend in="blurred" in2="noise" mode="multiply"/>
</filter>
```
