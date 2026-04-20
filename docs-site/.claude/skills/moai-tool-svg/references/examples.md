# SVG Working Examples

Complete, working code examples for common SVG patterns and use cases.

---

## Example 1: Responsive Icon System

Icon component with size and color props:

```xml
<svg xmlns="http://www.w3.org/2000/svg" style="display: none;">
  <defs>
    <symbol id="icon-home" viewBox="0 0 24 24">
      <path d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    </symbol>

    <symbol id="icon-settings" viewBox="0 0 24 24">
      <path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
            fill="none" stroke="currentColor" stroke-width="2"/>
      <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="2"/>
    </symbol>

    <symbol id="icon-user" viewBox="0 0 24 24">
      <path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    </symbol>
  </defs>
</svg>

<!-- Usage -->
<svg class="icon icon-small"><use href="#icon-home"/></svg>
<svg class="icon icon-medium" style="color: blue;"><use href="#icon-settings"/></svg>
<svg class="icon icon-large" style="color: green;"><use href="#icon-user"/></svg>
```

CSS for icon sizing:

```css
.icon { display: inline-block; }
.icon-small { width: 16px; height: 16px; }
.icon-medium { width: 24px; height: 24px; }
.icon-large { width: 32px; height: 32px; }
```

---

## Example 2: Logo with Gradients

Company logo with multiple gradients:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 60" width="200" height="60">
  <defs>
    <linearGradient id="logo-grad-1" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#667eea"/>
      <stop offset="100%" stop-color="#764ba2"/>
    </linearGradient>
    <linearGradient id="logo-grad-2" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="#f093fb"/>
      <stop offset="100%" stop-color="#f5576c"/>
    </linearGradient>
  </defs>

  <g transform="translate(10, 10)">
    <!-- Logo mark -->
    <circle cx="20" cy="20" r="18" fill="url(#logo-grad-1)"/>
    <path d="M12 20 L20 12 L28 20 L20 28 Z" fill="white"/>

    <!-- Logo text -->
    <text x="50" y="28" font-family="Arial, sans-serif" font-size="24" font-weight="bold" fill="url(#logo-grad-2)">
      BrandName
    </text>
  </g>
</svg>
```

---

## Example 3: Animated Loading Spinner

CSS-animated spinner:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 50 50" width="50" height="50">
  <style>
    .spinner-track { fill: none; stroke: #e0e0e0; stroke-width: 4; }
    .spinner-head {
      fill: none;
      stroke: #3498db;
      stroke-width: 4;
      stroke-linecap: round;
      stroke-dasharray: 80, 200;
      stroke-dashoffset: 0;
      animation: spinner-dash 1.5s ease-in-out infinite;
      transform-origin: center;
    }
    @keyframes spinner-dash {
      0% { stroke-dasharray: 1, 200; stroke-dashoffset: 0; }
      50% { stroke-dasharray: 89, 200; stroke-dashoffset: -35; }
      100% { stroke-dasharray: 89, 200; stroke-dashoffset: -124; }
    }
    .spinner-container { animation: spinner-rotate 2s linear infinite; transform-origin: center; }
    @keyframes spinner-rotate { 100% { transform: rotate(360deg); } }
  </style>

  <circle class="spinner-track" cx="25" cy="25" r="20"/>
  <g class="spinner-container">
    <circle class="spinner-head" cx="25" cy="25" r="20"/>
  </g>
</svg>
```

---

## Example 4: Data Visualization Bar Chart

Interactive bar chart with hover effects:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 250" width="400" height="250">
  <style>
    .bar { transition: fill 0.3s; cursor: pointer; }
    .bar:hover { fill: #e74c3c; }
    .axis { stroke: #333; stroke-width: 1; }
    .grid { stroke: #e0e0e0; stroke-width: 0.5; }
    .label { font-family: Arial, sans-serif; font-size: 12px; fill: #666; }
    .value { font-family: Arial, sans-serif; font-size: 10px; fill: #333; text-anchor: middle; }
  </style>

  <defs>
    <linearGradient id="bar-grad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#3498db"/>
      <stop offset="100%" stop-color="#2980b9"/>
    </linearGradient>
  </defs>

  <!-- Grid lines -->
  <g class="grid">
    <line x1="50" y1="50" x2="380" y2="50"/>
    <line x1="50" y1="100" x2="380" y2="100"/>
    <line x1="50" y1="150" x2="380" y2="150"/>
  </g>

  <!-- Axes -->
  <line class="axis" x1="50" y1="200" x2="380" y2="200"/>
  <line class="axis" x1="50" y1="50" x2="50" y2="200"/>

  <!-- Y-axis labels -->
  <text class="label" x="40" y="55" text-anchor="end">100</text>
  <text class="label" x="40" y="105" text-anchor="end">75</text>
  <text class="label" x="40" y="155" text-anchor="end">50</text>
  <text class="label" x="40" y="205" text-anchor="end">0</text>

  <!-- Bars -->
  <rect class="bar" x="70" y="80" width="50" height="120" fill="url(#bar-grad)" rx="3"/>
  <rect class="bar" x="140" y="110" width="50" height="90" fill="url(#bar-grad)" rx="3"/>
  <rect class="bar" x="210" y="60" width="50" height="140" fill="url(#bar-grad)" rx="3"/>
  <rect class="bar" x="280" y="130" width="50" height="70" fill="url(#bar-grad)" rx="3"/>

  <!-- Value labels -->
  <text class="value" x="95" y="75">80</text>
  <text class="value" x="165" y="105">60</text>
  <text class="value" x="235" y="55">93</text>
  <text class="value" x="305" y="125">47</text>

  <!-- X-axis labels -->
  <text class="label" x="95" y="220" text-anchor="middle">Q1</text>
  <text class="label" x="165" y="220" text-anchor="middle">Q2</text>
  <text class="label" x="235" y="220" text-anchor="middle">Q3</text>
  <text class="label" x="305" y="220" text-anchor="middle">Q4</text>
</svg>
```

---

## Example 5: Path Drawing Animation

Line drawing effect for illustration:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200" width="200" height="200">
  <style>
    .draw-line {
      fill: none;
      stroke: #2c3e50;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
      stroke-dasharray: 500;
      stroke-dashoffset: 500;
      animation: draw 2s ease forwards;
    }
    .draw-line:nth-child(2) { animation-delay: 0.5s; }
    .draw-line:nth-child(3) { animation-delay: 1s; }
    @keyframes draw {
      to { stroke-dashoffset: 0; }
    }
  </style>

  <!-- House outline -->
  <path class="draw-line" d="M30 100 L100 40 L170 100"/>
  <path class="draw-line" d="M50 100 L50 160 L150 160 L150 100"/>
  <path class="draw-line" d="M80 160 L80 120 L120 120 L120 160"/>
</svg>
```

---

## Example 6: Complex Filter Effect

Glass morphism card effect:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 200" width="300" height="200">
  <defs>
    <!-- Background blur filter -->
    <filter id="glass-blur" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur in="SourceGraphic" stdDeviation="10" result="blur"/>
      <feColorMatrix in="blur" type="matrix"
        values="1 0 0 0 0
                0 1 0 0 0
                0 0 1 0 0
                0 0 0 0.8 0"/>
    </filter>

    <!-- Drop shadow -->
    <filter id="card-shadow" x="-20%" y="-20%" width="140%" height="140%">
      <feDropShadow dx="0" dy="10" stdDeviation="20" flood-color="rgba(0,0,0,0.2)"/>
    </filter>

    <!-- Gradient overlay -->
    <linearGradient id="glass-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="rgba(255,255,255,0.4)"/>
      <stop offset="100%" stop-color="rgba(255,255,255,0.1)"/>
    </linearGradient>
  </defs>

  <!-- Background pattern -->
  <rect width="300" height="200" fill="#667eea"/>
  <circle cx="50" cy="50" r="60" fill="#764ba2" opacity="0.5"/>
  <circle cx="250" cy="150" r="80" fill="#f093fb" opacity="0.5"/>

  <!-- Glass card -->
  <g filter="url(#card-shadow)">
    <rect x="50" y="40" width="200" height="120" rx="15" fill="white" opacity="0.2" filter="url(#glass-blur)"/>
    <rect x="50" y="40" width="200" height="120" rx="15" fill="url(#glass-gradient)" stroke="rgba(255,255,255,0.3)" stroke-width="1"/>
  </g>

  <!-- Card content -->
  <text x="150" y="90" text-anchor="middle" font-family="Arial" font-size="18" font-weight="bold" fill="white">Glass Card</text>
  <text x="150" y="115" text-anchor="middle" font-family="Arial" font-size="12" fill="rgba(255,255,255,0.8)">Frosted glass effect</text>
</svg>
```

---

## Example 7: Responsive Illustration

Illustration that scales with container:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300" preserveAspectRatio="xMidYMid meet" style="width: 100%; max-width: 400px; height: auto;">
  <title>Mountain Landscape</title>
  <desc>A scenic mountain landscape with sun and trees</desc>

  <defs>
    <linearGradient id="sky-grad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#87CEEB"/>
      <stop offset="100%" stop-color="#E0F6FF"/>
    </linearGradient>
    <linearGradient id="mountain-grad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#5D6D7E"/>
      <stop offset="100%" stop-color="#85929E"/>
    </linearGradient>
    <radialGradient id="sun-grad" cx="50%" cy="50%" r="50%">
      <stop offset="0%" stop-color="#FFF5CC"/>
      <stop offset="100%" stop-color="#FFD700"/>
    </radialGradient>
  </defs>

  <!-- Sky -->
  <rect width="400" height="300" fill="url(#sky-grad)"/>

  <!-- Sun -->
  <circle cx="320" cy="60" r="35" fill="url(#sun-grad)"/>

  <!-- Mountains -->
  <polygon points="0,300 100,150 200,300" fill="url(#mountain-grad)"/>
  <polygon points="150,300 280,120 400,300" fill="#34495E"/>
  <polygon points="250,300 350,180 400,250 400,300" fill="#5D6D7E"/>

  <!-- Snow caps -->
  <polygon points="100,150 120,180 80,180" fill="white"/>
  <polygon points="280,120 300,150 260,150" fill="white"/>

  <!-- Ground -->
  <rect y="250" width="400" height="50" fill="#27AE60"/>

  <!-- Trees -->
  <g transform="translate(50, 220)">
    <polygon points="15,0 0,50 30,50" fill="#1E8449"/>
    <rect x="12" y="50" width="6" height="15" fill="#6E2C00"/>
  </g>
  <g transform="translate(350, 210)">
    <polygon points="20,0 0,60 40,60" fill="#1E8449"/>
    <rect x="16" y="60" width="8" height="20" fill="#6E2C00"/>
  </g>
</svg>
```

---

## Example 8: Interactive Button

Button with hover and click states:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 160 50" width="160" height="50" style="cursor: pointer;">
  <style>
    .btn-bg {
      fill: #3498db;
      transition: fill 0.2s, transform 0.1s;
    }
    .btn-bg:hover { fill: #2980b9; }
    .btn-bg:active { transform: translateY(2px); fill: #1a5276; }
    .btn-text {
      fill: white;
      font-family: Arial, sans-serif;
      font-size: 16px;
      font-weight: bold;
      pointer-events: none;
    }
    .btn-icon {
      fill: white;
      pointer-events: none;
    }
  </style>

  <defs>
    <filter id="btn-shadow">
      <feDropShadow dx="0" dy="3" stdDeviation="3" flood-color="rgba(0,0,0,0.3)"/>
    </filter>
  </defs>

  <rect class="btn-bg" x="5" y="5" width="150" height="40" rx="8" filter="url(#btn-shadow)"/>

  <g class="btn-icon" transform="translate(20, 15)">
    <path d="M10 0 L20 10 L10 20 L0 10 Z"/>
  </g>

  <text class="btn-text" x="90" y="30" text-anchor="middle">Click Me</text>
</svg>
```

---

## Example 9: Pie Chart

Data visualization pie chart:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 200" width="300" height="200">
  <style>
    .slice { transition: transform 0.2s; transform-origin: 100px 100px; }
    .slice:hover { transform: scale(1.05); }
    .legend-text { font-family: Arial, sans-serif; font-size: 12px; }
  </style>

  <!-- Pie slices using path arcs -->
  <!-- 40% slice (144 degrees) -->
  <path class="slice" d="M100 100 L100 20 A80 80 0 0 1 176.6 148.3 Z" fill="#3498db"/>

  <!-- 30% slice (108 degrees) -->
  <path class="slice" d="M100 100 L176.6 148.3 A80 80 0 0 1 38.4 164.7 Z" fill="#e74c3c"/>

  <!-- 20% slice (72 degrees) -->
  <path class="slice" d="M100 100 L38.4 164.7 A80 80 0 0 1 23.4 51.7 Z" fill="#f39c12"/>

  <!-- 10% slice (36 degrees) -->
  <path class="slice" d="M100 100 L23.4 51.7 A80 80 0 0 1 100 20 Z" fill="#27ae60"/>

  <!-- Legend -->
  <g transform="translate(210, 40)">
    <rect width="15" height="15" fill="#3498db"/>
    <text class="legend-text" x="20" y="12">Product A (40%)</text>

    <rect y="25" width="15" height="15" fill="#e74c3c"/>
    <text class="legend-text" x="20" y="37">Product B (30%)</text>

    <rect y="50" width="15" height="15" fill="#f39c12"/>
    <text class="legend-text" x="20" y="62">Product C (20%)</text>

    <rect y="75" width="15" height="15" fill="#27ae60"/>
    <text class="legend-text" x="20" y="87">Product D (10%)</text>
  </g>
</svg>
```

---

## Example 10: SMIL Animation

Native SVG animation without CSS:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200" width="200" height="200">
  <!-- Bouncing ball -->
  <circle cx="100" cy="50" r="20" fill="#e74c3c">
    <animate attributeName="cy" values="50;150;50" dur="1s" repeatCount="indefinite" calcMode="spline" keySplines="0.5 0 0.5 1; 0.5 0 0.5 1"/>
    <animate attributeName="rx" values="20;22;20" dur="1s" repeatCount="indefinite" calcMode="spline" keySplines="0.5 0 0.5 1; 0.5 0 0.5 1"/>
  </circle>

  <!-- Shadow -->
  <ellipse cx="100" cy="170" rx="20" ry="5" fill="rgba(0,0,0,0.3)">
    <animate attributeName="rx" values="20;30;20" dur="1s" repeatCount="indefinite" calcMode="spline" keySplines="0.5 0 0.5 1; 0.5 0 0.5 1"/>
    <animate attributeName="opacity" values="0.3;0.1;0.3" dur="1s" repeatCount="indefinite" calcMode="spline" keySplines="0.5 0 0.5 1; 0.5 0 0.5 1"/>
  </ellipse>

  <!-- Rotating star -->
  <g transform="translate(100, 100)">
    <polygon points="0,-30 7,-10 30,-10 12,5 18,30 0,15 -18,30 -12,5 -30,-10 -7,-10" fill="#f39c12">
      <animateTransform attributeName="transform" type="rotate" from="0" to="360" dur="4s" repeatCount="indefinite"/>
    </polygon>
  </g>

  <!-- Pulsing circle -->
  <circle cx="50" cy="50" r="10" fill="#3498db" opacity="0.7">
    <animate attributeName="r" values="10;15;10" dur="1.5s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.7;0.3;0.7" dur="1.5s" repeatCount="indefinite"/>
  </circle>
</svg>
```

---

## Example 11: Text Path Animation

Animated text along curved path:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 150" width="300" height="150">
  <defs>
    <path id="wave-path" d="M0 75 Q75 25 150 75 T300 75" fill="none"/>
  </defs>

  <style>
    .wave-text {
      font-family: Arial, sans-serif;
      font-size: 14px;
      fill: #3498db;
    }
  </style>

  <!-- Visible path for reference -->
  <use href="#wave-path" stroke="#e0e0e0" stroke-width="1"/>

  <!-- Animated text -->
  <text class="wave-text">
    <textPath href="#wave-path">
      This text follows a wavy path and animates smoothly
      <animate attributeName="startOffset" from="0%" to="100%" dur="8s" repeatCount="indefinite"/>
    </textPath>
  </text>
</svg>
```

---

## Example 12: Morphing Shape

Shape morphing between states:

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
  <style>
    .morph-shape {
      fill: #9b59b6;
      transition: d 0.5s ease-in-out;
    }
    .morph-shape:hover {
      d: path("M50 10 L90 50 L50 90 L10 50 Z");
    }
  </style>

  <!-- Circle that morphs to diamond on hover -->
  <path class="morph-shape" d="M50 10 A40 40 0 1 1 50 90 A40 40 0 1 1 50 10"/>

  <!-- Alternative using SMIL animation -->
  <path fill="#3498db" transform="translate(0, 0)">
    <animate attributeName="d"
             values="M50 10 A40 40 0 1 1 50 90 A40 40 0 1 1 50 10;
                     M50 10 L90 50 L50 90 L10 50 Z;
                     M50 10 A40 40 0 1 1 50 90 A40 40 0 1 1 50 10"
             dur="3s"
             repeatCount="indefinite"/>
  </path>
</svg>
```
