# moai-design-figma-to-code

Design-to-code conversion pipelines, asset extraction, React component generation from Figma designs.

## Quick Start

Design-to-code bridges Figma designs and production code by automating component creation, CSS generation, and asset export. Use this skill when converting design files to React components, generating Tailwind CSS classes, or automating design implementation workflows.

## Core Patterns

### Pattern 1: Converting Design Frames to React Components

**Pattern**: Transform Figma frames into React components with Tailwind CSS.

```typescript
// lib/design-to-code.ts - Convert Figma frames to React
import { getFileStructure } from './figma';

interface FrameData {
  id: string;
  name: string;
  width: number;
  height: number;
  children: NodeData[];
  fills: Array<{
    color: { r: number; g: number; b: number; a: number };
  }>;
}

interface NodeData {
  id: string;
  name: string;
  type: string;
  x: number;
  y: number;
  width: number;
  height: number;
  fills?: Array<{ color: { r: number; g: number; b: number } }>;
  strokes?: Array<{ color: { r: number; g: number; b: number } }>;
  characters?: string;
  fontSize?: number;
  fontFamily?: string;
  fontWeight?: number;
  children?: NodeData[];
}

// Generate Tailwind CSS classes from design node
function generateTailwindClasses(node: NodeData): string {
  const classes: string[] = [];

  // Sizing
  classes.push(`w-[${node.width}px]`);
  classes.push(`h-[${node.height}px]`);

  // Positioning
  if (node.x !== 0) classes.push(`left-[${node.x}px]`);
  if (node.y !== 0) classes.push(`top-[${node.y}px]`);

  // Background color
  if (node.fills && node.fills.length > 0) {
    const fill = node.fills[0];
    const hex = rgbToHex(fill.color.r, fill.color.g, fill.color.b);
    classes.push(`bg-[${hex}]`);
  }

  // Border
  if (node.strokes && node.strokes.length > 0) {
    const stroke = node.strokes[0];
    const hex = rgbToHex(stroke.color.r, stroke.color.g, stroke.color.b);
    classes.push(`border-[${hex}]`);
    classes.push(`border`);
  }

  // Typography
  if (node.fontSize) {
    classes.push(`text-[${node.fontSize}px]`);
  }

  if (node.fontWeight) {
    const weight = node.fontWeight === 700 ? 'bold' : 'normal';
    classes.push(`font-${weight}`);
  }

  return classes.join(' ');
}

// Generate React component from frame
function generateReactComponent(frame: FrameData): string {
  const componentName = sanitizeName(frame.name);
  const tailwindClasses = generateTailwindClasses(frame);

  let childrenCode = '';

  if (frame.children) {
    childrenCode = frame.children
      .map((child) => generateElementCode(child))
      .join('\n  ');
  }

  return `
export function ${componentName}() {
  return (
    <div className="${tailwindClasses}">
      ${childrenCode}
    </div>
  );
}
  `.trim();
}

// Generate individual element code
function generateElementCode(node: NodeData, depth = 1): string {
  const indent = '  '.repeat(depth);
  const classes = generateTailwindClasses(node);

  switch (node.type) {
    case 'TEXT':
      return `${indent}<p className="${classes}">${node.characters || 'Text'}</p>`;

    case 'RECTANGLE':
      return `${indent}<div className="${classes}"></div>`;

    case 'IMAGE':
      return `${indent}<img className="${classes}" alt="" />`;

    case 'GROUP':
    case 'FRAME':
      const children = (node.children || [])
        .map((child) => generateElementCode(child, depth + 1))
        .join('\n');

      return `${indent}<div className="${classes}">\n${children}\n${indent}</div>`;

    default:
      return `${indent}<div className="${classes}"></div>`;
  }
}

// Main conversion function
export async function convertDesignToCode(fileId: string) {
  const fileStructure = await getFileStructure();

  // Find component frames
  function findFrames(node: any, frames: FrameData[] = []): FrameData[] {
    if (node.type === 'FRAME') {
      frames.push(node);
    }

    if (node.children) {
      node.children.forEach((child: any) => findFrames(child, frames));
    }

    return frames;
  }

  const frames = findFrames(fileStructure.document);

  // Generate components for each frame
  const components = frames.map((frame) => generateReactComponent(frame));

  return components;
}

function sanitizeName(name: string): string {
  return name
    .replace(/[^a-zA-Z0-9]/g, '')
    .replace(/^[0-9]/, '')
    .split('')
    .reduce((acc, char, idx) => {
      if (idx === 0) return char.toUpperCase();
      return acc + char;
    });
}

function rgbToHex(r: number, g: number, b: number): string {
  return `#${[r, g, b]
    .map((x) => Math.round(x * 255).toString(16).padStart(2, '0'))
    .join('')}`;
}
`;

/**
 * This exports components as individual files
 */
export async function exportComponentsAsFiles(fileId: string) {
  const components = await convertDesignToCode(fileId);

  const fileMap: Record<string, string> = {};

  components.forEach((code) => {
    const match = code.match(/export function (\w+)/);
    if (match) {
      const name = match[1];
      fileMap[`components/${name}.tsx`] = code;
    }
  });

  return fileMap;
}
```

**When to use**:
- Converting design mockups to production React components
- Automating repetitive component creation
- Maintaining design-code consistency
- Rapid prototyping from Figma

**Key benefits**:
- Reduces manual component creation time
- Ensures design accuracy in code
- Maintains design tokens from Figma
- Enables rapid iteration

### Pattern 2: Asset Extraction & Optimization

**Pattern**: Extract and optimize images, icons, and other assets from Figma.

```typescript
// lib/asset-extraction.ts - Extract assets from Figma
import fetch from 'node-fetch';
import fs from 'fs/promises';
import path from 'path';

export interface Asset {
  id: string;
  name: string;
  url: string;
  type: 'image' | 'icon' | 'illustration';
  format: 'png' | 'svg' | 'pdf';
}

// Get image URLs from Figma
export async function extractAssets(fileId: string): Promise<Asset[]> {
  const fileStructure = await getFileStructure();

  const assets: Asset[] = [];

  function findImages(node: any) {
    if (node.type === 'IMAGE') {
      assets.push({
        id: node.id,
        name: node.name,
        url: '', // Will be populated by image export endpoint
        type: 'image',
        format: 'png',
      });
    }

    if (node.type === 'COMPONENT' && node.name.includes('icon')) {
      assets.push({
        id: node.id,
        name: node.name,
        url: '',
        type: 'icon',
        format: 'svg',
      });
    }

    if (node.children) {
      node.children.forEach(findImages);
    }
  }

  findImages(fileStructure.document);

  // Get export URLs for each asset
  const assetIds = assets.map((a) => a.id).join(',');

  const response = await figmaApi.get(
    `/files/${fileId}/export`,
    {
      params: {
        ids: assetIds,
        format: 'png',
        scale: 2, // 2x resolution for retina
      },
    }
  );

  return assets.map((asset, idx) => ({
    ...asset,
    url: response.data.images[asset.id],
  }));
}

// Download and optimize assets
export async function downloadAssets(
  assets: Asset[],
  outputDir: string = 'public/assets'
) {
  await fs.mkdir(outputDir, { recursive: true });

  for (const asset of assets) {
    try {
      const response = await fetch(asset.url);
      const buffer = await response.buffer();

      const fileName = `${sanitizeName(asset.name)}.${asset.format}`;
      const filePath = path.join(outputDir, asset.type, fileName);

      // Create subdirectory
      await fs.mkdir(path.dirname(filePath), { recursive: true });

      // Write file
      await fs.writeFile(filePath, buffer);

      console.log(`Downloaded: ${filePath}`);
    } catch (error) {
      console.error(`Failed to download ${asset.name}:`, error);
    }
  }
}

// Generate TypeScript enum for assets
export function generateAssetEnum(assets: Asset[]): string {
  const iconAssets = assets
    .filter((a) => a.type === 'icon')
    .map((a) => `  ${sanitizeName(a.name).toUpperCase()} = '${a.id}',`);

  return `
// Generated from Figma assets
export enum IconAsset {
${iconAssets.join('\n')}
}

export const iconPaths: Record<IconAsset, string> = {
${assets
  .filter((a) => a.type === 'icon')
  .map((a) => {
    const name = sanitizeName(a.name).toUpperCase();
    const path = `/assets/icons/${sanitizeName(a.name)}.svg`;
    return `  [IconAsset.${name}]: '${path}',`;
  })
  .join('\n')}
};
  `.trim();
}
```

**When to use**:
- Exporting icons and illustrations from Figma
- Optimizing images for web
- Creating asset inventories
- Automating asset updates

**Key benefits**:
- Batch download and organize assets
- Automatic optimization
- Version control for design assets
- Single source of truth in Figma

### Pattern 3: Component Library Synchronization

**Pattern**: Keep Figma component library in sync with React component library.

```typescript
// lib/sync-components.ts - Sync Figma to React
export interface ComponentMapping {
  figmaId: string;
  figmaName: string;
  reactPath: string;
  lastSynced: Date;
  version: string;
}

export async function syncComponentLibrary(): Promise<ComponentMapping[]> {
  const figmaComponents = await listComponents();

  const mappings: ComponentMapping[] = [];

  for (const component of figmaComponents) {
    // Generate React component name from Figma name
    const reactName = sanitizeName(component.name);
    const reactPath = `components/ui/${reactName}.tsx`;

    // Check if React component exists
    const exists = await checkFileExists(reactPath);

    if (!exists) {
      // Generate new React component
      const code = await generateComponentCode(component);
      await writeFile(reactPath, code);

      console.log(`Created: ${reactPath}`);
    } else {
      // Component exists, mark as synced
      console.log(`Already exists: ${reactPath}`);
    }

    mappings.push({
      figmaId: component.id,
      figmaName: component.name,
      reactPath,
      lastSynced: new Date(),
      version: component.version,
    });
  }

  // Save mappings for future reference
  await writeFile(
    'component-mappings.json',
    JSON.stringify(mappings, null, 2)
  );

  return mappings;
}

async function checkFileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

async function generateComponentCode(component: any): Promise<string> {
  // Extract component properties from Figma
  const props = extractComponentProps(component);

  const propsInterface = props
    .map((p) => `  ${p.name}: ${p.type};`)
    .join('\n');

  return `
'use client';

interface ${component.name}Props {
${propsInterface}
}

export function ${component.name}({
${props.map((p) => p.name).join(', ')}
}: ${component.name}Props) {
  return (
    <div>
      {/* Component implementation */}
    </div>
  );
}
  `.trim();
}

function extractComponentProps(component: any): Array<{ name: string; type: string }> {
  // Extract props from component description or custom properties
  // This is a simplified example
  return [
    { name: 'children', type: 'React.ReactNode' },
    { name: 'className', type: 'string' },
  ];
}
```

**When to use**:
- Maintaining design-code consistency
- Automating component library updates
- Tracking design changes
- Ensuring all design components have code equivalents

**Key benefits**:
- Single source of truth for components
- Automated sync reduces manual updates
- Component versioning and tracking
- Faster design-to-implementation cycle

## Progressive Disclosure

### Level 1: Basic Conversion
- Export frames as React components
- Generate Tailwind CSS classes
- Basic styling and layout conversion
- Asset export and download

### Level 2: Advanced Patterns
- Component library synchronization
- Asset optimization and batching
- TypeScript prop generation
- Design token extraction

### Level 3: Expert Automation
- Multi-file component generation
- Complex nested component handling
- Animation and interaction conversion
- CI/CD integration for continuous sync

## Works Well With

- **Figma API**: Source of design data
- **React 19**: Target component framework
- **TypeScript**: Type-safe component generation
- **Tailwind CSS**: Styling framework
- **Next.js**: Build pipeline and API routes
- **shadcn/ui**: Component library base

## References

- **Figma REST API**: https://www.figma.com/developers/api
- **File Endpoints**: https://www.figma.com/developers/api#get-files-endpoint
- **Export Endpoint**: https://www.figma.com/developers/api#export-endpoint
- **Components**: https://www.figma.com/developers/api#components
- **Design Tokens**: https://www.figma.com/developers/api#variables-endpoints
