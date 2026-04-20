# SVG Optimization Module

Comprehensive guide to SVGO configuration, compression techniques, icon sprites, and performance best practices.

---

## SVGO Installation

### Global Installation

```bash
npm install -g svgo
```

### Project Installation

```bash
npm install --save-dev svgo
```

### Verify Installation

```bash
svgo --version
```

---

## SVGO CLI Usage

### Basic Optimization

```bash
# Optimize single file
svgo input.svg -o output.svg

# Optimize in place
svgo input.svg

# Optimize directory
svgo -f ./src/icons -o ./dist/icons

# Recursive directory optimization
svgo -f ./src/icons -o ./dist/icons --recursive
```

### Common Options

```bash
# Show optimization statistics
svgo input.svg --pretty

# Pretty print with indentation
svgo input.svg --pretty --indent=2

# Use custom config file
svgo input.svg --config svgo.config.mjs

# Disable specific plugin
svgo input.svg --disable=removeViewBox

# Enable specific plugin
svgo input.svg --enable=removeDimensions
```

### Output Formats

```bash
# Output to stdout
svgo input.svg -o -

# String input
svgo --string='<svg>...</svg>'

# Multiple files
svgo file1.svg file2.svg file3.svg
```

---

## SVGO Configuration

### Basic Configuration File

Create svgo.config.mjs in project root:

```javascript
export default {
  multipass: true,
  plugins: [
    'preset-default'
  ]
};
```

### Customizing preset-default

Override specific plugins within the default preset:

```javascript
export default {
  multipass: true,
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          // Disable plugins
          removeViewBox: false,
          removeTitle: false,
          removeDesc: false,

          // Customize plugins
          cleanupIds: {
            minify: true,
            preserve: ['logo', 'icon-']
          },
          convertPathData: {
            floatPrecision: 2
          }
        }
      }
    }
  ]
};
```

### Adding Custom Plugins

```javascript
export default {
  multipass: true,
  plugins: [
    'preset-default',

    // Add prefix to all IDs
    {
      name: 'prefixIds',
      params: {
        prefix: 'app-',
        delim: '-'
      }
    },

    // Remove specific attributes
    {
      name: 'removeAttrs',
      params: {
        attrs: ['data-name', 'class', 'style']
      }
    },

    // Add attributes to root SVG
    {
      name: 'addAttributesToSVGElement',
      params: {
        attributes: [
          { 'aria-hidden': 'true' },
          { focusable: 'false' }
        ]
      }
    }
  ]
};
```

### Configuration for Different Use Cases

Icon system configuration:

```javascript
export default {
  multipass: true,
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          removeViewBox: false,
          cleanupIds: false
        }
      }
    },
    'removeXMLNS',
    'removeDimensions',
    {
      name: 'removeAttrs',
      params: {
        attrs: ['class', 'data-name', 'fill', 'stroke']
      }
    },
    {
      name: 'addAttributesToSVGElement',
      params: {
        attributes: [{ fill: 'currentColor' }]
      }
    }
  ]
};
```

Logo/illustration configuration:

```javascript
export default {
  multipass: true,
  plugins: [
    {
      name: 'preset-default',
      params: {
        overrides: {
          removeViewBox: false,
          removeTitle: false,
          removeDesc: false,
          cleanupIds: {
            preserve: ['logo', 'brand']
          }
        }
      }
    },
    {
      name: 'prefixIds',
      params: {
        prefix: 'logo'
      }
    }
  ]
};
```

---

## Plugin Reference

### Cleanup Plugins

**cleanupAttrs**: Removes newlines, trailing spaces from attributes.

**cleanupIds**: Minifies IDs and removes unused ones.

```javascript
{
  name: 'cleanupIds',
  params: {
    remove: true,
    minify: true,
    preserve: ['important-id'],
    preservePrefixes: ['keep-']
  }
}
```

**cleanupNumericValues**: Rounds numeric values, removes units from zero values.

```javascript
{
  name: 'cleanupNumericValues',
  params: {
    floatPrecision: 3
  }
}
```

### Removal Plugins

**removeDoctype**: Removes DOCTYPE declaration.

**removeXMLProcInst**: Removes XML processing instructions.

**removeComments**: Removes comments.

**removeMetadata**: Removes metadata elements.

**removeTitle**: Removes title element (disable for accessibility).

**removeDesc**: Removes desc element (disable for accessibility).

**removeUselessDefs**: Removes unused defs.

**removeEditorsNSData**: Removes editor namespaces and data.

**removeEmptyAttrs**: Removes empty attributes.

**removeHiddenElems**: Removes hidden elements.

**removeEmptyText**: Removes empty text elements.

**removeEmptyContainers**: Removes empty container elements.

**removeViewBox**: Removes viewBox when it matches width/height (often disable).

**removeUnknownsAndDefaults**: Removes unknown elements and default values.

**removeNonInheritableGroupAttrs**: Removes non-inheritable group attributes.

**removeUselessStrokeAndFill**: Removes useless stroke and fill attributes.

**removeUnusedNS**: Removes unused namespaces.

**removeDimensions**: Removes width/height, preserves viewBox.

### Conversion Plugins

**convertStyleToAttrs**: Converts style to presentation attributes.

**convertColors**: Optimizes color values.

```javascript
{
  name: 'convertColors',
  params: {
    currentColor: true,
    names2hex: true,
    rgb2hex: true,
    shorthex: true,
    shortname: false
  }
}
```

**convertPathData**: Optimizes path data.

```javascript
{
  name: 'convertPathData',
  params: {
    applyTransforms: true,
    applyTransformsStroked: true,
    floatPrecision: 3,
    transformPrecision: 5,
    removeUseless: true,
    collapseRepeated: true,
    utilizeAbsolute: true,
    leadingZero: true,
    negativeExtraSpace: true
  }
}
```

**convertTransform**: Optimizes transform values.

**convertShapeToPath**: Converts basic shapes to paths.

### Merge/Collapse Plugins

**collapseGroups**: Collapses useless groups.

**mergePaths**: Merges multiple paths into one.

**moveElemsAttrsToGroup**: Moves common element attributes to groups.

**moveGroupAttrsToElems**: Moves group attributes to elements.

### Addition Plugins

**prefixIds**: Adds prefix to IDs.

```javascript
{
  name: 'prefixIds',
  params: {
    prefix: 'svg-',
    delim: '-',
    prefixIds: true,
    prefixClassNames: true
  }
}
```

**addAttributesToSVGElement**: Adds attributes to root SVG.

**addClassesToSVGElement**: Adds classes to root SVG.

---

## Icon Sprite Techniques

### External Sprite File

Create sprites.svg with symbols:

```xml
<svg xmlns="http://www.w3.org/2000/svg" style="display: none;">
  <symbol id="icon-home" viewBox="0 0 24 24">
    <path d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
  </symbol>

  <symbol id="icon-settings" viewBox="0 0 24 24">
    <path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35"/>
  </symbol>
</svg>
```

Usage in HTML:

```html
<svg class="icon" width="24" height="24">
  <use href="/sprites.svg#icon-home"></use>
</svg>
```

### Inline Sprite

Include sprite in HTML body:

```html
<body>
  <!-- Hidden sprite at top of body -->
  <svg xmlns="http://www.w3.org/2000/svg" style="display: none;">
    <symbol id="icon-home" viewBox="0 0 24 24">...</symbol>
    <symbol id="icon-settings" viewBox="0 0 24 24">...</symbol>
  </svg>

  <!-- Usage anywhere in document -->
  <svg class="icon"><use href="#icon-home"></use></svg>
</body>
```

### Build Tool Integration

Using svg-sprite with npm script:

```json
{
  "scripts": {
    "icons": "svg-sprite --symbol --symbol-dest=dist --symbol-sprite=sprites.svg src/icons/*.svg"
  }
}
```

---

## SVGZ Compression

### Creating SVGZ Files

SVGZ is gzip-compressed SVG. Typical 20-50% size reduction.

```bash
# Using gzip
gzip -c input.svg > output.svgz

# Using Node.js
node -e "require('fs').writeFileSync('output.svgz', require('zlib').gzipSync(require('fs').readFileSync('input.svg')))"
```

### Server Configuration

Apache (.htaccess):

```apache
AddType image/svg+xml svg svgz
AddEncoding gzip svgz
```

Nginx:

```nginx
location ~ \.svgz$ {
    add_header Content-Encoding gzip;
    types { image/svg+xml svgz; }
}
```

---

## Performance Best Practices

### File Size Optimization

Reduce decimal precision in paths from default 6 decimals to 2:

Before: M 10.123456 20.654321 L 30.987654 40.123456

After: M 10.12 20.65 L 30.99 40.12

Use relative path commands when they result in shorter strings. Combine adjacent paths when possible. Remove invisible elements and unused definitions.

### Rendering Performance

Minimize filter complexity. Filters are expensive, especially blur. Limit nested groups to reduce DOM depth. Use symbols for repeated elements instead of duplicating paths.

Avoid animating expensive properties. Transform and opacity are GPU-accelerated. Avoid animating path d attribute for complex shapes.

### Loading Performance

Inline critical SVGs to avoid additional HTTP requests. Use external sprites for non-critical icons. Lazy load below-fold SVG content.

Cache external SVG files with appropriate headers:

```nginx
location ~ \.svg$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### Accessibility Performance

Keep title and desc for meaningful graphics. Screen readers announce these elements. Decorative SVGs should have aria-hidden="true".

```xml
<!-- Meaningful graphic -->
<svg role="img" aria-labelledby="title">
  <title id="title">Sales Chart</title>
  <desc>Bar chart showing quarterly sales increasing from Q1 to Q4</desc>
</svg>

<!-- Decorative graphic -->
<svg aria-hidden="true" focusable="false">
  <!-- decorative content -->
</svg>
```

### Mobile Optimization

Use viewBox for responsive sizing. Avoid fixed width/height. Test touch target sizes (minimum 44x44 CSS pixels). Reduce complexity for mobile devices.

```css
/* Responsive SVG */
.responsive-svg {
  width: 100%;
  max-width: 400px;
  height: auto;
}

/* Touch-friendly icon */
.touch-icon {
  min-width: 44px;
  min-height: 44px;
  padding: 10px;
}
```

---

## Optimization Workflow

### Development Workflow

Keep source SVGs unoptimized for editing. Run optimization as build step. Version control both source and optimized versions.

```
src/
  icons/           # Source SVGs (unoptimized)
  illustrations/   # Source SVGs (unoptimized)

dist/
  icons/           # Optimized SVGs
  sprites.svg      # Combined sprite
```

### Build Integration

Package.json scripts:

```json
{
  "scripts": {
    "svg:optimize": "svgo -f src/icons -o dist/icons --config svgo.config.mjs",
    "svg:sprite": "svg-sprite --symbol --symbol-sprite=sprites.svg dist/icons/*.svg",
    "svg:build": "npm run svg:optimize && npm run svg:sprite"
  }
}
```

### Continuous Integration

Add SVG optimization to CI pipeline:

```yaml
- name: Optimize SVGs
  run: |
    npm install -g svgo
    svgo -f src/icons -o dist/icons --config svgo.config.mjs

- name: Check for uncommitted changes
  run: |
    git diff --exit-code dist/icons
```

### Quality Assurance

Compare before/after visually. Automated visual regression testing recommended for critical graphics. Check file sizes in CI to prevent regression.

```bash
# Check total icon size
du -sh dist/icons/
# Expected: < 100KB for typical icon set
```
