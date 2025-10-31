# moai-design-figma-mcp

Figma API integration for design file access, component library management, design tokens extraction, and webhooks.

## Quick Start

Figma API enables programmatic access to design files, components, and design tokens. Use this skill when syncing designs to code, extracting design tokens, managing design systems, or automating design workflows.

## Core Patterns

### Pattern 1: Reading Design Files & Components

**Pattern**: Access Figma files and extract component information via REST API.

```typescript
// lib/figma.ts - Figma API client
import axios from 'axios';

const FIGMA_API_KEY = process.env.FIGMA_API_KEY;
const FIGMA_FILE_ID = process.env.FIGMA_FILE_ID;

const figmaApi = axios.create({
  baseURL: 'https://api.figma.com/v1',
  headers: {
    'X-FIGMA-TOKEN': FIGMA_API_KEY,
  },
});

// Get design file structure
export async function getFileStructure() {
  try {
    const response = await figmaApi.get(`/files/${FIGMA_FILE_ID}`);

    const { document, components } = response.data;

    return {
      fileName: response.data.name,
      pages: document.children,
      components: components || {},
      version: response.data.version,
    };
  } catch (error) {
    console.error('Failed to fetch file structure:', error);
    throw error;
  }
}

// Get specific page content
export async function getPageContent(pageId: string) {
  try {
    const response = await figmaApi.get(
      `/files/${FIGMA_FILE_ID}/nodes`,
      {
        params: {
          ids: pageId,
        },
      }
    );

    return response.data.nodes[pageId].document;
  } catch (error) {
    console.error('Failed to fetch page content:', error);
    throw error;
  }
}

// Get component details
export async function getComponent(componentId: string) {
  try {
    const response = await figmaApi.get(
      `/files/${FIGMA_FILE_ID}/components/${componentId}`
    );

    return response.data;
  } catch (error) {
    console.error('Failed to fetch component:', error);
    throw error;
  }
}

// Extract component metadata
export async function listComponents() {
  const { components } = await getFileStructure();

  return Object.entries(components).map(([id, component]: any) => ({
    id,
    name: component.name,
    description: component.description,
    createdAt: component.created_at,
    updatedAt: component.updated_at,
  }));
}
```

**When to use**:
- Building design-to-code pipelines
- Extracting component libraries from Figma
- Syncing design tokens to code
- Generating component documentation

**Key benefits**:
- Access design data without exporting manually
- Programmatic component enumeration
- Version tracking and change detection
- Integration with CI/CD workflows

### Pattern 2: Design Tokens Extraction

**Pattern**: Extract colors, typography, spacing from Figma variables and styles.

```typescript
// lib/design-tokens.ts - Extract design tokens from Figma
interface DesignTokens {
  colors: Record<string, string>;
  typography: Record<string, any>;
  spacing: Record<string, string>;
  shadows: Record<string, string>;
}

export async function extractDesignTokens(): Promise<DesignTokens> {
  const fileStructure = await getFileStructure();

  const tokens: DesignTokens = {
    colors: {},
    typography: {},
    spacing: {},
    shadows: {},
  };

  // Parse styles from Figma document
  function parseStyles(node: any) {
    if (node.styles) {
      // Color styles
      if (node.fills) {
        node.fills.forEach((fill: any) => {
          if (fill.type === 'SOLID') {
            const color = fill.color;
            const hex = rgbToHex(color.r, color.g, color.b);
            tokens.colors[node.name] = hex;
          }
        });
      }

      // Typography styles
      if (node.typography) {
        tokens.typography[node.name] = {
          fontSize: node.fontSize,
          fontFamily: node.fontFamily,
          fontWeight: node.fontWeight,
          lineHeight: node.lineHeightPx,
          letterSpacing: node.letterSpacing,
        };
      }

      // Shadow styles
      if (node.effects) {
        node.effects.forEach((effect: any) => {
          if (effect.type === 'DROP_SHADOW') {
            tokens.shadows[node.name] = `
              ${effect.offset.x}px ${effect.offset.y}px
              ${effect.radius}px rgba(0, 0, 0, ${effect.color.a})
            `;
          }
        });
      }
    }

    // Recurse through children
    if (node.children) {
      node.children.forEach(parseStyles);
    }
  }

  parseStyles(fileStructure.document);

  return tokens;
}

// Helper to convert RGB to hex
function rgbToHex(r: number, g: number, b: number): string {
  return `#${[r, g, b]
    .map((x) => {
      const hex = Math.round(x * 255).toString(16);
      return hex.length === 1 ? '0' + hex : hex;
    })
    .join('')}`;
}

// Export tokens as TypeScript/CSS
export async function exportTokensAsTypeScript() {
  const tokens = await extractDesignTokens();

  const tsCode = `
// Generated design tokens from Figma
export const designTokens = {
  colors: ${JSON.stringify(tokens.colors, null, 2)},
  typography: ${JSON.stringify(tokens.typography, null, 2)},
  spacing: ${JSON.stringify(tokens.spacing, null, 2)},
  shadows: ${JSON.stringify(tokens.shadows, null, 2)},
} as const;

export type DesignTokens = typeof designTokens;
  `;

  return tsCode;
}
```

**When to use**:
- Keeping design tokens in sync with Figma
- Generating theme configuration files
- Creating design system documentation
- Maintaining single source of truth for design

**Key benefits**:
- Automatic token extraction reduces manual work
- Version control for design decisions
- Type-safe design tokens in code
- Easy to update across entire project

### Pattern 3: Webhooks for Design Change Detection

**Pattern**: Set up webhooks to detect Figma changes and trigger automation.

```typescript
// api/webhooks/figma.ts - Figma webhook handler
import { NextApiRequest, NextApiResponse } from 'next';
import crypto from 'crypto';

const FIGMA_WEBHOOK_SECRET = process.env.FIGMA_WEBHOOK_SECRET;

interface FigmaWebhookEvent {
  timestamp: number;
  trigger: {
    id: string;
    type: 'FILE' | 'LIBRARY';
  };
  event: {
    type: 'FILE_UPDATE' | 'FILE_DELETE' | 'LIBRARY_PUBLISH';
    file_name?: string;
    created_components?: Array<{
      id: string;
      name: string;
    }>;
    updated_components?: Array<{
      id: string;
      name: string;
    }>;
    deleted_components?: string[];
  };
}

// Verify webhook authenticity
function verifyWebhookSignature(
  request: NextApiRequest
): boolean {
  const signature = request.headers['x-figma-signature'] as string;
  const body = JSON.stringify(request.body);

  const hash = crypto
    .createHmac('sha256', FIGMA_WEBHOOK_SECRET)
    .update(body)
    .digest('base64');

  return signature === hash;
}

export default async function handler(
  request: NextApiRequest,
  response: NextApiResponse
) {
  if (request.method !== 'POST') {
    return response.status(405).json({ error: 'Method not allowed' });
  }

  // Verify webhook signature
  if (!verifyWebhookSignature(request)) {
    return response.status(401).json({ error: 'Unauthorized' });
  }

  const event: FigmaWebhookEvent = request.body;

  console.log('Figma webhook received:', event.event.type);

  try {
    // Handle different event types
    switch (event.event.type) {
      case 'FILE_UPDATE':
        // Trigger design-to-code conversion
        await triggerDesignSync(event.trigger.id);
        break;

      case 'LIBRARY_PUBLISH':
        // Update component library
        await updateComponentLibrary(event.trigger.id);
        break;

      case 'FILE_DELETE':
        // Clean up resources
        await cleanupFile(event.trigger.id);
        break;
    }

    response.status(200).json({
      success: true,
      processed: event.event.type,
    });
  } catch (error) {
    console.error('Webhook processing error:', error);
    response.status(500).json({ error: 'Processing failed' });
  }
}

async function triggerDesignSync(fileId: string) {
  console.log(`Syncing design changes for file ${fileId}`);
  // Trigger CI/CD pipeline or design-to-code generation
}

async function updateComponentLibrary(fileId: string) {
  console.log(`Updating component library from ${fileId}`);
  // Extract components and update library
}

async function cleanupFile(fileId: string) {
  console.log(`Cleaning up resources for deleted file ${fileId}`);
}
```

**When to use**:
- Setting up automated design-to-code workflows
- Triggering component library updates on design changes
- Creating real-time design sync pipelines
- Maintaining design and code in sync

**Key benefits**:
- Real-time notification of design changes
- Automated deployment of design updates
- Reduced manual synchronization effort
- Audit trail of design changes

## Progressive Disclosure

### Level 1: Basic File Access
- Read design file structure
- List components and pages
- Extract basic metadata
- Download design assets

### Level 2: Advanced Integration
- Extract design tokens and styles
- Programmatic component enumeration
- Asset export and optimization
- Design-to-code generation

### Level 3: Expert Automation
- Webhook configuration and handling
- Real-time design change detection
- Automated CI/CD integration
- Design system governance

## Works Well With

- **Figma**: Official design tool with REST API
- **Next.js**: Server-side API routes for webhook handling
- **TypeScript**: Type-safe Figma API responses
- **Tailwind CSS**: Token extraction for styling
- **shadcn/ui**: Component library synchronization
- **GitHub Actions**: Automated design-to-code workflows

## References

- **Figma REST API**: https://www.figma.com/developers/api
- **Access Tokens**: https://www.figma.com/developers/api#access-tokens
- **File Endpoints**: https://www.figma.com/developers/api#file-endpoints
- **Webhooks**: https://www.figma.com/developers/api#webhooks
- **Variables API**: https://www.figma.com/developers/api#variables-endpoints
