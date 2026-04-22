# Favicon Icons Required

The following favicon files should be added to the `public/` directory for optimal SEO and user experience:

## Required Files

1. **favicon.ico** - Legacy browser support (32x32 minimum)
2. **favicon.svg** - Modern scalable icon
3. **apple-touch-icon.png** - iOS home screen (180x180)
4. **icon-192.png** - Android/Manifest (192x192)
5. **icon-512.png** - Android/Manifest (512x512)

## Tools to Generate

### Option 1: Online Tools
- https://realfavicongenerator.net/ - Comprehensive favicon generator
- https://favicon.io/ - Simple favicon generator
- https://www.favicon-generator.org/ - Batch favicon creator

### Option 2: CLI Tool
```bash
# Install sharp for image processing
npm install --save-dev sharp

# Run the favicon generator script
node scripts/generate-favicons.js
```

### Option 3: Use Existing Logo
Use the existing `og.png` as source and resize it to the required dimensions.

## Current Status

The project currently has:
- ✅ `og.png` - Open Graph image (1200x630)
- ❌ `favicon.ico` - Missing
- ❌ `favicon.svg` - Missing
- ❌ `apple-touch-icon.png` - Missing
- ❌ `icon-192.png` - Missing
- ❌ `icon-512.png` - Missing

## Priority

For immediate SEO improvement, create at minimum:
1. `favicon.ico` (32x32)
2. `apple-touch-icon.png` (180x180)

The manifest.json references `icon-192.png` and `icon-512.png` for PWA support.
