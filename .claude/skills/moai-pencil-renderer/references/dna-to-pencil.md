# DNA to Pencil Property Mapping

Complete reference for translating aesthetic DNA codes to Pencil .pen file properties.

## Layout Axis

### centered
```javascript
{
  layout: "vertical",
  alignItems: "center",
  padding: [64, 48, 64, 48],  // symmetric
  gap: 24
}
```

### asymmetric
```javascript
{
  layout: "horizontal",
  alignItems: "flex-start",
  padding: [80, 120, 64, 48],  // varied
  gap: 32
}
// Child positioning: use x/y offsets, varied widths
```

### grid-breaking
```javascript
{
  layout: "none",  // absolute positioning
}
// Children use explicit x, y coordinates
// Overlapping elements, diagonal compositions
```

### full-bleed
```javascript
{
  width: "fill_container",
  height: "fill_container",
  padding: [0, 0, 0, 0]
}
```

### bento
```javascript
{
  layout: "grid",
  columns: 4,
  rows: 3,
  gap: 16
}
// Children use colspan, rowspan for varied cell sizes
```

### editorial
```javascript
{
  layout: "vertical",
  alignItems: "flex-start",
  padding: [96, 64, 96, 200],  // generous margins
  maxWidth: 720
}
```

## Color Axis

### dark
```javascript
// Frame fills
fill: "#0A0A0A"  // near-black
// or use variables
fill: "var(--color-background-dark)"

// Text colors
textColor: "#FAFAFA"  // near-white
```

### light
```javascript
fill: "#FEFEFE"  // near-white
textColor: "#1A1A1A"  // near-black
```

### monochrome
```javascript
// Limited palette: black, white, 2-3 grays
fill: "#FFFFFF"
// Accent via stroke weight, not color
stroke: "#000000"
strokeThickness: 2
```

### gradient
```javascript
fill: {
  type: "linear",
  angle: 135,
  stops: [
    { position: 0, color: "#1E3A5F" },
    { position: 1, color: "#0A1628" }
  ]
}
```

### high-contrast
```javascript
fill: "#000000"
textColor: "#FFFFFF"
// No mid-tones, stark opposites
```

### brand-tinted
```javascript
// Neutrals tinted with brand hue
fill: "oklch(0.98 0.01 250)"  // very light, slight blue
textColor: "oklch(0.15 0.02 250)"  // near-black, slight blue
```

## Typography Axis

### display-heavy
```javascript
// Headlines dominate
{
  type: "text",
  content: "Hero Title",
  fontSize: 72,
  fontWeight: "900",
  fontFamily: "Playfair Display"
}
// Body much smaller
{
  fontSize: 16,
  fontWeight: "400"
}
```

### text-forward
```javascript
// Body text prominence
{
  fontSize: 20,
  lineHeight: 1.7,
  fontFamily: "Georgia"
}
// Headlines restrained
{
  fontSize: 32,
  fontWeight: "600"
}
```

### minimal
```javascript
// Single family, tight scale
fontFamily: "Helvetica Neue"
// All weights from same family
// Scale: 14, 16, 20, 24, 32 (1.25 ratio)
```

### expressive
```javascript
// Mixed families, decorative
// Headline: Display serif
// Body: Geometric sans
// Accents: Handwritten or display
{
  fontFamily: "Fraunces",
  fontWeight: "400",
  fontStyle: "italic"
}
```

### editorial
```javascript
// Magazine-style hierarchy
// Large drop caps, pull quotes
// Serif headings, sans body or vice versa
{
  firstLetter: {
    fontSize: 96,
    float: "left"
  }
}
```

## Motion Axis

**Note:** Pencil is static. Motion DNA informs placeholder/documentation.

| Motion Value | Design Implication |
|-------------|-------------------|
| orchestrated | Placeholder for animation sequence notes |
| subtle | Minimal transition indicators |
| aggressive | Bold state changes (show both states) |
| none | No animation planned |
| scroll-triggered | Section reveals on scroll (show waypoints) |

## Density Axis

### spacious
```javascript
{
  gap: 48,
  padding: [80, 64, 80, 64]
}
// Generous whitespace
// Few elements per section
```

### compact
```javascript
{
  gap: 12,
  padding: [24, 20, 24, 20]
}
// Tight layout
// Information-dense
```

### mixed
```javascript
// Vary by section
// Hero: spacious
// Data: compact
// CTA: spacious
```

### full-bleed (density)
```javascript
{
  padding: [0, 0, 0, 0],
  gap: 0
}
// Edge-to-edge content
```

## Background Axis

### solid
```javascript
fill: "#FFFFFF"
// Single color, no complexity
```

### gradient
```javascript
fill: {
  type: "linear",
  angle: 180,
  stops: [
    { position: 0, color: "#F8FAFC" },
    { position: 1, color: "#E2E8F0" }
  ]
}
```

### textured
```javascript
// Use G() operation to apply texture
G("frameId", "stock", "subtle noise texture background")
// or
G("frameId", "ai", "subtle paper texture warm beige")
```

### patterned
```javascript
// Use G() for pattern
G("frameId", "ai", "geometric pattern minimal black white")
// or SVG pattern fill
```

### layered
```javascript
// Multiple overlapping frames
// Background layer + decorative shapes + content
bg=I(parent, {type: "frame", fill: "#0A0A0A"})
accent=I(parent, {type: "ellipse", fill: "gradient", opacity: 0.5})
content=I(parent, {type: "frame", fill: "transparent"})
```

## Complete DNA → Properties Example

DNA: `[asymmetric, dark, display-heavy, orchestrated, spacious, gradient]`

```javascript
// Container
container=I(document, {
  type: "frame",
  name: "Hero-Asymmetric-Dark",
  width: 1440,
  height: 900,
  layout: "horizontal",
  alignItems: "flex-start",
  padding: [120, 80, 80, 160],
  gap: 64,
  fill: {
    type: "linear",
    angle: 135,
    stops: [
      { position: 0, color: "#1E293B" },
      { position: 1, color: "#0F172A" }
    ]
  }
})

// Headline (display-heavy)
heading=I(container, {
  type: "text",
  content: "Headline Text",
  fontSize: 80,
  fontWeight: "900",
  fontFamily: "Playfair Display",
  textColor: "#F8FAFC",
  width: 600
})

// Body text (subordinate)
body=I(container, {
  type: "text",
  content: "Supporting copy...",
  fontSize: 18,
  fontWeight: "400",
  fontFamily: "Inter",
  textColor: "#94A3B8",
  width: 400,
  lineHeight: 1.6
})
```

## Style Guide Tag Mapping

Map DNA to Pencil style guide tags for `get_style_guide`:

| DNA | Tags |
|-----|------|
| dark | dark-mode, moody, dramatic |
| light | light, clean, airy |
| gradient | vibrant, colorful |
| spacious | whitespace, minimal, airy |
| compact | dense, data-heavy |
| display-heavy | bold, impactful |
| editorial | magazine, publication |
| brutalist | raw, industrial |
