# SVG Reference Documentation

Extended API reference for SVG elements, attributes, and SVGO configuration.

---

## SVG Element Reference

### Container Elements

**svg**: Root container element. Attributes include xmlns (required for standalone), viewBox, width, height, preserveAspectRatio, and x/y for nested SVGs.

**g**: Group element for organizing content. Supports transform, opacity, and presentation attributes that inherit to children.

**defs**: Container for reusable definitions. Content inside defs is not rendered directly. Used for gradients, patterns, filters, symbols, and markers.

**symbol**: Defines reusable graphic with its own viewBox. Never rendered directly. Instantiated via use element. Supports viewBox and preserveAspectRatio.

**use**: References and renders content defined elsewhere. Attributes include href (or xlink:href for legacy), x, y, width, height. Inherits styles but cannot modify internal elements directly.

**a**: Hyperlink wrapper for SVG content. Attributes include href, target, rel. Wraps any SVG element to make it clickable.

**switch**: Conditional rendering based on feature support. Children evaluated in order, first match rendered. Uses requiredFeatures, requiredExtensions, systemLanguage.

**foreignObject**: Embeds non-SVG content like HTML. Attributes include x, y, width, height. Useful for complex text layouts or HTML form elements within SVG.

---

## Shape Element Reference

### rect

Rectangle with optional rounded corners.

Required attributes: width, height

Optional attributes: x, y (default 0), rx, ry (corner radii)

Special cases: If only rx specified, ry equals rx. Maximum rx is half of width.

### circle

Circle defined by center and radius.

Required attributes: cx, cy, r

All attributes default to 0. Radius of 0 produces no visible circle.

### ellipse

Ellipse defined by center and two radii.

Required attributes: cx, cy, rx, ry

rx for horizontal radius, ry for vertical radius.

### line

Straight line between two points.

Required attributes: x1, y1, x2, y2

Must have stroke attribute to be visible (no fill).

### polyline

Connected series of line segments.

Required attribute: points (space or comma-separated coordinate pairs)

Example: points="0,0 10,20 30,10"

Does not automatically close. Typically used with fill="none".

### polygon

Closed shape from connected line segments.

Required attribute: points (same format as polyline)

Automatically closes path from last point to first.

### path

Most powerful shape element using path commands.

Required attribute: d (path data string)

Supports all drawing operations through command letters.

---

## Path Command Reference

### Move Commands

M x y: Absolute move to coordinates (x, y). Starts new subpath.

m dx dy: Relative move by offset (dx, dy) from current position.

### Line Commands

L x y: Absolute line to coordinates (x, y).

l dx dy: Relative line by offset from current position.

H x: Horizontal line to absolute x coordinate. Y unchanged.

h dx: Horizontal line by relative offset. Y unchanged.

V y: Vertical line to absolute y coordinate. X unchanged.

v dy: Vertical line by relative offset. X unchanged.

### Close Command

Z or z: Closes current subpath by drawing line to starting point.

### Cubic Bezier Commands

C x1 y1 x2 y2 x y: Cubic bezier curve.
- (x1, y1): First control point
- (x2, y2): Second control point
- (x, y): End point

c: Relative version with all offsets.

S x2 y2 x y: Smooth cubic bezier.
- First control point reflected from previous curve
- (x2, y2): Second control point
- (x, y): End point

s: Relative version.

### Quadratic Bezier Commands

Q x1 y1 x y: Quadratic bezier curve.
- (x1, y1): Control point
- (x, y): End point

q: Relative version.

T x y: Smooth quadratic bezier.
- Control point reflected from previous curve
- (x, y): End point

t: Relative version.

### Arc Command

A rx ry rotation large-arc sweep x y: Elliptical arc.
- rx: X-axis radius
- ry: Y-axis radius
- rotation: X-axis rotation in degrees
- large-arc: 0 for smaller arc, 1 for larger arc
- sweep: 0 for counter-clockwise, 1 for clockwise
- (x, y): End point

a: Relative version (end point relative).

---

## Text Element Reference

### text

Primary text container element.

Positioning: x, y for starting position. Multiple values create multiple text positions.

Styling: font-family, font-size, font-weight, font-style, text-decoration.

Alignment: text-anchor (start, middle, end), dominant-baseline (auto, middle, hanging, text-top).

### tspan

Inline text span for styling portions of text.

Inherits from parent text element. Can override any text attribute.

Supports dx, dy for relative positioning from previous character.

### textPath

Text rendered along a path shape.

Required: href pointing to path element.

Attributes: startOffset (percentage or length along path), method (align, stretch), spacing (auto, exact).

---

## Gradient Reference

### linearGradient

Linear color transition along a line.

Direction attributes: x1, y1 (start), x2, y2 (end). Values 0% to 100% or absolute units.

Common directions:
- Horizontal: x1="0%" y1="0%" x2="100%" y2="0%"
- Vertical: x1="0%" y1="0%" x2="0%" y2="100%"
- Diagonal: x1="0%" y1="0%" x2="100%" y2="100%"

gradientUnits: userSpaceOnUse (absolute) or objectBoundingBox (percentage, default).

spreadMethod: pad (default), reflect, repeat for colors beyond gradient bounds.

### radialGradient

Radial color transition from center outward.

Position attributes: cx, cy (center), r (radius), fx, fy (focal point).

Focal point creates offset highlight effect when different from center.

gradientUnits and spreadMethod same as linearGradient.

### stop

Color stop within gradient.

Required: offset (0% to 100% or 0 to 1).

Color: stop-color (color value), stop-opacity (0 to 1).

---

## Filter Reference

### filter

Container for filter operations.

Size attributes: x, y, width, height define filter region (default: -10%, -10%, 120%, 120%).

filterUnits: userSpaceOnUse or objectBoundingBox.

primitiveUnits: userSpaceOnUse or objectBoundingBox for child primitives.

### Filter Primitives

**feGaussianBlur**: Blur effect.
- in: Input (SourceGraphic, SourceAlpha, or previous result)
- stdDeviation: Blur amount (single value or "x y")
- result: Output name

**feOffset**: Offset transformation.
- in: Input
- dx, dy: Offset amounts
- result: Output name

**feColorMatrix**: Color transformation.
- in: Input
- type: matrix, saturate, hueRotate, luminanceToAlpha
- values: Matrix values or single value for non-matrix types

**feBlend**: Blend two inputs.
- in, in2: Two inputs
- mode: normal, multiply, screen, darken, lighten, overlay

**feComposite**: Combine two inputs using Porter-Duff operations.
- in, in2: Two inputs
- operator: over, in, out, atop, xor, arithmetic

**feMerge**: Layer multiple inputs.
- Contains feMergeNode children, each with in attribute
- Renders in order (first at bottom)

**feDropShadow**: Shorthand for drop shadow effect.
- dx, dy: Shadow offset
- stdDeviation: Blur amount
- flood-color: Shadow color
- flood-opacity: Shadow opacity

**feFlood**: Fill with solid color.
- flood-color: Color value
- flood-opacity: Opacity value

**feImage**: Insert external image or SVG reference.
- href: Image source
- preserveAspectRatio: Aspect ratio handling

**feMorphology**: Erode or dilate shapes.
- operator: erode, dilate
- radius: Effect amount

**feTurbulence**: Generate noise pattern.
- type: turbulence, fractalNoise
- baseFrequency: Noise frequency
- numOctaves: Detail level
- seed: Random seed

**feDisplacementMap**: Distort using another image.
- in: Input to distort
- in2: Displacement map
- scale: Distortion amount
- xChannelSelector, yChannelSelector: R, G, B, or A

---

## Clipping and Masking Reference

### clipPath

Hard-edged clipping region.

clipPathUnits: userSpaceOnUse or objectBoundingBox.

Content: Any shape elements define clip region. Union of all shapes.

Usage: clip-path="url(#clip-id)" on target element.

### mask

Soft-edged masking with alpha channel.

maskUnits: userSpaceOnUse or objectBoundingBox.

maskContentUnits: userSpaceOnUse or objectBoundingBox for child elements.

Content: Shapes where white reveals, black conceals, gray partially reveals.

Usage: mask="url(#mask-id)" on target element.

---

## SVGO Plugin Reference

### preset-default

Collection of safe default optimizations. Includes:
- cleanupAttrs: Remove newlines, trailing spaces
- removeDoctype: Remove DOCTYPE
- removeXMLProcInst: Remove XML declaration
- removeComments: Remove comments
- removeMetadata: Remove metadata
- removeTitle: Remove title (disable for accessibility)
- removeDesc: Remove desc (disable for accessibility)
- removeUselessDefs: Remove empty defs
- removeEditorsNSData: Remove editor metadata
- removeEmptyAttrs: Remove empty attributes
- removeHiddenElems: Remove hidden elements
- removeEmptyText: Remove empty text
- removeEmptyContainers: Remove empty containers
- removeViewBox: Remove viewBox when possible (often disable)
- cleanupEnableBackground: Remove enable-background
- convertStyleToAttrs: Convert style to attributes
- convertColors: Optimize color values
- convertPathData: Optimize path data
- convertTransform: Optimize transforms
- removeUnknownsAndDefaults: Remove unknown elements
- removeNonInheritableGroupAttrs: Clean group attributes
- removeUselessStrokeAndFill: Remove useless stroke/fill
- removeUnusedNS: Remove unused namespaces
- cleanupIds: Minify IDs
- cleanupNumericValues: Round numeric values
- moveElemsAttrsToGroup: Consolidate attributes
- moveGroupAttrsToElems: Distribute attributes
- collapseGroups: Merge nested groups
- mergePaths: Merge adjacent paths
- sortAttrs: Sort attributes
- sortDefsChildren: Sort defs children

### Individual Plugins

**prefixIds**: Add prefix to IDs to avoid conflicts.
- prefix: Custom prefix string or function
- delim: Delimiter between prefix and ID

**removeAttrs**: Remove specific attributes.
- attrs: Array of attribute patterns to remove

**addAttributesToSVGElement**: Add attributes to root svg.
- attributes: Object of attributes to add

**removeAttributesBySelector**: Remove attributes matching selector.
- selector: CSS selector
- attributes: Attributes to remove

**removeDimensions**: Remove width/height, keep viewBox.

**removeScriptElement**: Remove script elements.

**removeStyleElement**: Remove style elements.

**removeXMLNS**: Remove xmlns from root (for inline SVG).

**reusePaths**: Convert repeated paths to use elements.

**convertShapeToPath**: Convert shapes to path elements.

---

## Presentation Attributes Reference

### Fill Attributes

fill: Paint color (color value, none, url(#gradient-id))

fill-opacity: Opacity (0 to 1)

fill-rule: Winding rule (nonzero, evenodd)

### Stroke Attributes

stroke: Stroke color (color value, none, url(#gradient-id))

stroke-width: Line width (length)

stroke-opacity: Opacity (0 to 1)

stroke-linecap: Line end style (butt, round, square)

stroke-linejoin: Corner style (miter, round, bevel)

stroke-dasharray: Dash pattern (length list, none)

stroke-dashoffset: Dash offset (length)

stroke-miterlimit: Miter limit ratio (number)

### Opacity and Visibility

opacity: Element opacity (0 to 1)

visibility: Visibility (visible, hidden, collapse)

display: Display (inline, block, none)

### Transform

transform: Transformation functions

Functions:
- translate(x, y): Move by offset
- scale(x, y): Scale by factor
- rotate(angle, cx, cy): Rotate around point
- skewX(angle): Skew along X axis
- skewY(angle): Skew along Y axis
- matrix(a, b, c, d, e, f): Transformation matrix

transform-origin: Origin point for transformations

### Color and Painting

color: Current color (inheritable)

color-interpolation: Color space (auto, sRGB, linearRGB)

color-interpolation-filters: Filter color space

paint-order: Paint order (fill, stroke, markers)

### Markers

marker-start: Marker at path start

marker-mid: Marker at path vertices

marker-end: Marker at path end

---

## Coordinate System Reference

### viewBox

Format: "min-x min-y width height"

Defines the coordinate system for SVG content. Content scales to fit width/height while preserving aspect ratio (by default).

### preserveAspectRatio

Format: "align meetOrSlice"

Align values:
- none: No uniform scaling
- xMinYMin: Align to top-left
- xMidYMin: Align to top-center
- xMaxYMin: Align to top-right
- xMinYMid: Align to center-left
- xMidYMid: Align to center (default)
- xMaxYMid: Align to center-right
- xMinYMax: Align to bottom-left
- xMidYMax: Align to bottom-center
- xMaxYMax: Align to bottom-right

meetOrSlice:
- meet: Entire viewBox visible (letterboxing)
- slice: Fill viewport (cropping)

### Units

Absolute units: px, pt, pc, mm, cm, in

Relative units: em, ex, % (percentage of viewBox)

User units: Numbers without unit (same as px)
