# SVG Animation Module

Comprehensive guide to CSS animations, SMIL, JavaScript animations, and interaction patterns.

---

## CSS Animations

### Basic CSS Animation

CSS animations work on SVG elements just like HTML elements.

```xml
<svg viewBox="0 0 100 100" width="100" height="100">
  <style>
    .pulse {
      animation: pulse 2s ease-in-out infinite;
    }
    @keyframes pulse {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.5; transform: scale(1.1); }
    }
  </style>

  <circle class="pulse" cx="50" cy="50" r="40" fill="blue"/>
</svg>
```

### Transform Origin

SVG elements have transform-origin at (0, 0) by default, unlike HTML elements.

```css
/* Center the transform origin */
.centered-transform {
  transform-origin: center;
  /* Or explicit coordinates */
  transform-origin: 50px 50px;
}
```

### Animatable Properties

Common animatable SVG properties:

- transform: translate, rotate, scale, skew
- opacity: Fade effects
- fill, stroke: Color transitions
- stroke-dasharray, stroke-dashoffset: Line drawing effects
- r, cx, cy: Circle properties (in some browsers)
- x, y, width, height: Rectangle properties

### Line Drawing Animation

Classic stroke animation technique:

```xml
<svg viewBox="0 0 200 100" width="200" height="100">
  <style>
    .draw {
      fill: none;
      stroke: #333;
      stroke-width: 2;
      stroke-dasharray: 300;
      stroke-dashoffset: 300;
      animation: draw 2s ease forwards;
    }
    @keyframes draw {
      to { stroke-dashoffset: 0; }
    }
  </style>

  <path class="draw" d="M10 50 Q50 10, 100 50 T190 50"/>
</svg>
```

Calculate total path length with JavaScript:

```javascript
const path = document.querySelector('path');
const length = path.getTotalLength();
console.log(length); // Use this value for dasharray
```

### Staggered Animations

```xml
<svg viewBox="0 0 200 100" width="200" height="100">
  <style>
    .bar {
      animation: grow 0.5s ease forwards;
      transform-origin: bottom;
      transform: scaleY(0);
    }
    .bar:nth-child(1) { animation-delay: 0s; }
    .bar:nth-child(2) { animation-delay: 0.1s; }
    .bar:nth-child(3) { animation-delay: 0.2s; }
    .bar:nth-child(4) { animation-delay: 0.3s; }

    @keyframes grow {
      to { transform: scaleY(1); }
    }
  </style>

  <rect class="bar" x="20" y="20" width="30" height="60" fill="#3498db"/>
  <rect class="bar" x="60" y="30" width="30" height="50" fill="#3498db"/>
  <rect class="bar" x="100" y="10" width="30" height="70" fill="#3498db"/>
  <rect class="bar" x="140" y="40" width="30" height="40" fill="#3498db"/>
</svg>
```

### Hover Animations

```xml
<svg viewBox="0 0 100 100" width="100" height="100">
  <style>
    .hover-circle {
      fill: #3498db;
      transition: fill 0.3s, transform 0.3s;
      transform-origin: center;
    }
    .hover-circle:hover {
      fill: #e74c3c;
      transform: scale(1.1);
    }
  </style>

  <circle class="hover-circle" cx="50" cy="50" r="40"/>
</svg>
```

---

## SMIL Animations

SMIL (Synchronized Multimedia Integration Language) provides native SVG animation without CSS or JavaScript. Note: SMIL is deprecated in Chrome but still works; CSS/JS preferred for new projects.

### animate Element

Animates a single attribute over time.

```xml
<circle cx="50" cy="50" r="40" fill="blue">
  <animate
    attributeName="r"
    from="40"
    to="20"
    dur="1s"
    repeatCount="indefinite"
  />
</circle>
```

### Attribute Reference

**attributeName**: The attribute to animate (r, cx, fill, opacity, etc.)

**from/to**: Start and end values

**values**: Multiple keyframe values (replaces from/to)

**dur**: Duration (e.g., "2s", "500ms")

**begin**: Start time or event (e.g., "0s", "click", "other.end")

**repeatCount**: Number of iterations ("indefinite" for infinite)

**fill**: End state ("freeze" to keep end value, "remove" to reset)

**calcMode**: Interpolation mode ("linear", "discrete", "paced", "spline")

**keyTimes**: Timing for values (0-1 scale, semicolon-separated)

**keySplines**: Bezier curves for spline calcMode

### Values and KeyTimes

```xml
<circle cx="50" cy="50" r="20" fill="blue">
  <animate
    attributeName="r"
    values="20; 40; 20"
    keyTimes="0; 0.5; 1"
    dur="2s"
    repeatCount="indefinite"
  />
</circle>
```

### Easing with KeySplines

```xml
<circle cx="50" cy="50" r="20" fill="blue">
  <animate
    attributeName="cx"
    from="50"
    to="150"
    dur="1s"
    calcMode="spline"
    keySplines="0.42 0 0.58 1"
    repeatCount="indefinite"
  />
</circle>
```

Common easing curves:
- Ease: 0.25 0.1 0.25 1
- Ease-in: 0.42 0 1 1
- Ease-out: 0 0 0.58 1
- Ease-in-out: 0.42 0 0.58 1

### animateTransform Element

Animates transform attribute.

```xml
<rect x="-25" y="-25" width="50" height="50" fill="blue" transform="translate(50, 50)">
  <animateTransform
    attributeName="transform"
    type="rotate"
    from="0 50 50"
    to="360 50 50"
    dur="2s"
    repeatCount="indefinite"
  />
</rect>
```

Transform types: translate, scale, rotate, skewX, skewY

### animateMotion Element

Moves element along a path.

```xml
<circle r="10" fill="red">
  <animateMotion
    path="M20,50 C20,-50 180,150 180,50 C180-50 20,150 20,50 z"
    dur="5s"
    repeatCount="indefinite"
    rotate="auto"
  />
</circle>
```

Alternatively, reference a path:

```xml
<defs>
  <path id="motion-path" d="M20,50 Q100,0 180,50 T340,50"/>
</defs>

<circle r="10" fill="red">
  <animateMotion dur="3s" repeatCount="indefinite">
    <mpath href="#motion-path"/>
  </animateMotion>
</circle>
```

### set Element

Discrete attribute change at specific time.

```xml
<rect width="100" height="100" fill="blue">
  <set attributeName="fill" to="red" begin="2s"/>
</rect>
```

### Chaining Animations

```xml
<circle cx="50" cy="50" r="20" fill="blue">
  <animate
    id="anim1"
    attributeName="r"
    from="20"
    to="40"
    dur="1s"
    fill="freeze"
  />
  <animate
    attributeName="fill"
    from="blue"
    to="red"
    dur="1s"
    begin="anim1.end"
    fill="freeze"
  />
</circle>
```

### Event-Triggered Animations

```xml
<circle cx="50" cy="50" r="20" fill="blue">
  <animate
    attributeName="r"
    from="20"
    to="40"
    dur="0.3s"
    begin="click"
    fill="freeze"
  />
</circle>
```

Begin event options: click, mouseover, mouseout, focusin, focusout, beginEvent, endEvent

---

## JavaScript Animations

### requestAnimationFrame

The preferred method for smooth JavaScript animations.

```javascript
const circle = document.querySelector('circle');
let radius = 20;
let growing = true;

function animate() {
  if (growing) {
    radius += 0.5;
    if (radius >= 40) growing = false;
  } else {
    radius -= 0.5;
    if (radius <= 20) growing = true;
  }

  circle.setAttribute('r', radius);
  requestAnimationFrame(animate);
}

animate();
```

### Web Animations API

Modern API for complex animations.

```javascript
const circle = document.querySelector('circle');

circle.animate([
  { transform: 'scale(1)', opacity: 1 },
  { transform: 'scale(1.5)', opacity: 0.5 },
  { transform: 'scale(1)', opacity: 1 }
], {
  duration: 2000,
  iterations: Infinity,
  easing: 'ease-in-out'
});
```

### Path Animation with JavaScript

Animate along a path using getTotalLength and getPointAtLength:

```javascript
const path = document.querySelector('#motion-path');
const circle = document.querySelector('#moving-circle');
const pathLength = path.getTotalLength();
let progress = 0;

function animate() {
  progress = (progress + 0.5) % pathLength;
  const point = path.getPointAtLength(progress);

  circle.setAttribute('cx', point.x);
  circle.setAttribute('cy', point.y);

  requestAnimationFrame(animate);
}

animate();
```

### Morphing Paths

Animate between two path shapes (requires same number of points):

```javascript
const path = document.querySelector('path');
const startPath = 'M50 10 A40 40 0 1 1 50 90 A40 40 0 1 1 50 10';
const endPath = 'M50 10 L90 50 L50 90 L10 50 Z';

// Using flubber library for path interpolation
import { interpolate } from 'flubber';
const interpolator = interpolate(startPath, endPath);

let t = 0;
function animate() {
  t = (t + 0.01) % 1;
  const morphedPath = interpolator(t < 0.5 ? t * 2 : (1 - t) * 2);
  path.setAttribute('d', morphedPath);
  requestAnimationFrame(animate);
}

animate();
```

### GSAP for Complex Animations

GreenSock Animation Platform for production-quality animations:

```javascript
import gsap from 'gsap';

// Simple tween
gsap.to('.circle', {
  duration: 2,
  attr: { cx: 200, r: 50 },
  fill: '#e74c3c',
  ease: 'elastic.out(1, 0.3)'
});

// Timeline for sequenced animations
const tl = gsap.timeline({ repeat: -1 });

tl.to('.circle', { duration: 0.5, attr: { r: 40 } })
  .to('.circle', { duration: 0.5, fill: 'red' })
  .to('.circle', { duration: 0.5, attr: { cx: 150 } })
  .to('.circle', { duration: 0.5, attr: { r: 20, cx: 50 }, fill: 'blue' });
```

---

## Interaction Patterns

### Hover Effects

```xml
<svg viewBox="0 0 200 100" width="200" height="100">
  <style>
    .interactive-rect {
      fill: #3498db;
      cursor: pointer;
      transition: fill 0.3s, transform 0.2s;
    }
    .interactive-rect:hover {
      fill: #2980b9;
      transform: translateY(-5px);
    }
    .interactive-rect:active {
      transform: translateY(0);
    }
  </style>

  <rect class="interactive-rect" x="50" y="25" width="100" height="50" rx="5"/>
</svg>
```

### Click Handlers

```javascript
document.querySelector('.clickable').addEventListener('click', (e) => {
  const target = e.target;
  const currentFill = target.getAttribute('fill');
  target.setAttribute('fill', currentFill === 'blue' ? 'red' : 'blue');
});
```

### Drag and Drop

```javascript
const draggable = document.querySelector('.draggable');
let isDragging = false;
let offset = { x: 0, y: 0 };

draggable.addEventListener('mousedown', (e) => {
  isDragging = true;
  const bbox = draggable.getBBox();
  const ctm = draggable.getScreenCTM();
  offset.x = (e.clientX - ctm.e) / ctm.a - bbox.x - bbox.width / 2;
  offset.y = (e.clientY - ctm.f) / ctm.d - bbox.y - bbox.height / 2;
});

document.addEventListener('mousemove', (e) => {
  if (!isDragging) return;
  const ctm = draggable.getScreenCTM();
  const x = (e.clientX - ctm.e) / ctm.a - offset.x;
  const y = (e.clientY - ctm.f) / ctm.d - offset.y;
  draggable.setAttribute('cx', x);
  draggable.setAttribute('cy', y);
});

document.addEventListener('mouseup', () => {
  isDragging = false;
});
```

### Tooltip on Hover

```xml
<svg viewBox="0 0 200 150" width="200" height="150">
  <style>
    .tooltip {
      opacity: 0;
      transition: opacity 0.3s;
      pointer-events: none;
    }
    .data-point:hover + .tooltip {
      opacity: 1;
    }
  </style>

  <circle class="data-point" cx="50" cy="50" r="10" fill="blue"/>
  <g class="tooltip" transform="translate(50, 30)">
    <rect x="-30" y="-20" width="60" height="25" fill="black" rx="3"/>
    <text x="0" y="-5" text-anchor="middle" fill="white" font-size="12">Value: 42</text>
  </g>
</svg>
```

---

## Performance Considerations

### GPU-Accelerated Properties

For smooth 60fps animations, prefer these properties:

- transform (translate, rotate, scale)
- opacity

Avoid animating these (triggers layout/paint):

- width, height
- x, y, cx, cy
- path d attribute
- fill, stroke (can be expensive)

### will-change Hint

```css
.animated-element {
  will-change: transform, opacity;
}
```

Use sparingly as it consumes memory.

### Reducing Repaints

Group animated elements together. Use CSS containment where supported:

```css
.animated-container {
  contain: layout paint;
}
```

### Throttling Event Handlers

```javascript
let ticking = false;

function handleMouseMove(e) {
  if (!ticking) {
    requestAnimationFrame(() => {
      updatePosition(e.clientX, e.clientY);
      ticking = false;
    });
    ticking = true;
  }
}
```

### Complexity Management

Limit simultaneous animations. Pause off-screen animations:

```javascript
const observer = new IntersectionObserver((entries) => {
  entries.forEach((entry) => {
    const animation = entry.target.querySelector('animate');
    if (entry.isIntersecting) {
      animation.beginElement();
    } else {
      animation.endElement();
    }
  });
});

observer.observe(document.querySelector('.animated-svg'));
```

---

## Cross-Browser Compatibility

### SMIL Support

SMIL is deprecated in Chrome/Edge but still works. For maximum compatibility, use CSS animations or JavaScript.

### CSS Animation Fallbacks

```css
/* Feature detection for CSS animations */
@supports (animation: name) {
  .animated { animation: pulse 2s infinite; }
}

/* Fallback for older browsers */
.no-cssanimations .animated {
  /* Static fallback styles */
}
```

### JavaScript Feature Detection

```javascript
// Check for Web Animations API
if (element.animate) {
  element.animate(keyframes, options);
} else {
  // Fallback to requestAnimationFrame or library
}
```

### Transform Origin Quirks

Firefox may handle transform-origin differently. Use explicit coordinates:

```css
.cross-browser-transform {
  transform-origin: 50px 50px; /* Explicit */
  /* Rather than */
  transform-origin: center; /* May vary */
}
```
