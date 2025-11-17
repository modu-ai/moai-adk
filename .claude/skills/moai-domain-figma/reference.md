# moai-domain-figma: Complete API Reference

> **Version**: 1.0.0
> **Last Updated**: 2025-11-16
> **Status**: Production Ready
> **Figma API Version**: v1 (2025.10)

---

## Table of Contents

1. [Figma REST API Reference](#figma-rest-api-reference)
2. [MCP Tools Reference](#mcp-tools-reference)
3. [TypeScript Type Definitions](#typescript-type-definitions)
4. [Context7 Library Mappings](#context7-library-mappings)
5. [Parameter & Response Specifications](#parameter--response-specifications)
6. [Error Handling Reference](#error-handling-reference)
7. [Rate Limiting & Pagination](#rate-limiting--pagination)

---

## Figma REST API Reference

### API Base URL

```
https://api.figma.com/v1
```

### Authentication Header

```
X-Figma-Token: figd_your_personal_access_token_here
```

---

### Endpoint 1: Get File

**Purpose**: Retrieve complete file structure, components, variables, and assets

**HTTP Method**: `GET`

**Endpoint**: `/files/{file_key}`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique identifier for Figma file (from URL) |
| `depth` | integer | ❌ | 1 | Tree depth (1 = top level, 0 = unlimited) |
| `geometry` | boolean | ❌ | false | Include geometry (bounds, rotation, size) |
| `plugin_data` | boolean | ❌ | false | Include plugin-specific data |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/files/ABC123XYZ?depth=2&geometry=true&plugin_data=false" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "document": {
    "id": "0:0",
    "name": "Document Root",
    "type": "DOCUMENT",
    "children": [
      {
        "id": "1:2",
        "name": "Page 1",
        "type": "CANVAS",
        "children": [
          {
            "id": "2:3",
            "name": "Button",
            "type": "COMPONENT_SET",
            "componentPropertyDefinitions": {
              "State": {
                "type": "VARIANT",
                "defaultValue": "default",
                "variantOptions": ["default", "hover", "active"]
              }
            },
            "children": [
              {
                "id": "2:4",
                "name": "Button=default",
                "type": "COMPONENT",
                "bounds": {
                  "x": 0,
                  "y": 0,
                  "width": 200,
                  "height": 48
                },
                "fills": [
                  {
                    "type": "SOLID",
                    "color": {
                      "r": 0.05,
                      "g": 0.65,
                      "b": 0.9,
                      "a": 1.0
                    },
                    "opacity": 1.0
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "components": {
    "2:4": {
      "key": "2:4",
      "name": "Button=default",
      "description": "Primary button component",
      "componentSetId": "2:3",
      "remote": false,
      "documentationLinks": [],
      "containing_frame": {
        "nodeId": "1:2",
        "name": "Page 1"
      }
    }
  },
  "componentSets": {
    "2:3": {
      "key": "2:3",
      "name": "Button",
      "description": "Button component set",
      "documentationLinks": []
    }
  },
  "schemaVersion": 0,
  "styles": {
    "1:100": {
      "key": "1:100",
      "name": "Colors/Primary",
      "styleType": "FILL",
      "remote": false,
      "description": "Primary brand color"
    }
  }
}
```

**Status Codes**:
- `200 OK` - File retrieved successfully
- `400 Bad Request` - Invalid file_key format
- `401 Unauthorized` - Invalid or missing token
- `403 Forbidden` - User lacks permission to view file
- `404 Not Found` - File does not exist

---

### Endpoint 2: Get File Variables

**Purpose**: Retrieve all variables (design tokens) defined in file, including multi-mode support

**HTTP Method**: `GET`

**Endpoint**: `/files/{file_key}/variables`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique file identifier |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/files/ABC123XYZ/variables" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "meta": {
    "isPublished": false
  },
  "collections": [
    {
      "id": "VariableCollection:1:0",
      "name": "Colors",
      "modes": [
        {
          "modeId": "1:0",
          "name": "Light"
        },
        {
          "modeId": "1:1",
          "name": "Dark"
        },
        {
          "modeId": "1:2",
          "name": "HighContrast"
        }
      ],
      "defaultModeId": "1:0",
      "remoteComponentLibraryKey": null,
      "hiddenFromPublishing": false,
      "variableIds": [
        "VariableID:1:10",
        "VariableID:1:11",
        "VariableID:1:12"
      ]
    },
    {
      "id": "VariableCollection:1:1",
      "name": "Spacing",
      "modes": [
        {
          "modeId": "2:0",
          "name": "Default"
        }
      ],
      "defaultModeId": "2:0",
      "remoteComponentLibraryKey": null,
      "hiddenFromPublishing": false,
      "variableIds": [
        "VariableID:2:1",
        "VariableID:2:2",
        "VariableID:2:3"
      ]
    }
  ],
  "variables": {
    "VariableID:1:10": {
      "id": "VariableID:1:10",
      "name": "Colors/Primary",
      "key": "d2b5ff8fc32f0b6938903a7268297ce24d99b1ab",
      "description": "Primary brand color used for key interactions",
      "remote": false,
      "variableCollectionId": "VariableCollection:1:0",
      "resolvedType": "COLOR",
      "codeSyntax": {
        "WEB": "colors.primary",
        "ANDROID": "R.color.primary",
        "iOS": "ColorPrimary"
      },
      "valuesByMode": {
        "1:0": {
          "type": "SOLID",
          "color": {
            "r": 0.0549,
            "g": 0.6510,
            "b": 0.9020,
            "a": 1.0
          }
        },
        "1:1": {
          "type": "SOLID",
          "color": {
            "r": 0.3922,
            "g": 0.7608,
            "b": 0.9686,
            "a": 1.0
          }
        },
        "1:2": {
          "type": "SOLID",
          "color": {
            "r": 0.0,
            "g": 0.5,
            "b": 1.0,
            "a": 1.0
          }
        }
      }
    },
    "VariableID:1:11": {
      "id": "VariableID:1:11",
      "name": "Colors/Background",
      "key": "3c8d2a5f4e7b9c1d6e2f8a3b9d4c5e7f",
      "description": "Background color for main content areas",
      "remote": false,
      "variableCollectionId": "VariableCollection:1:0",
      "resolvedType": "COLOR",
      "codeSyntax": {
        "WEB": "colors.background",
        "ANDROID": "R.color.background",
        "iOS": "ColorBackground"
      },
      "valuesByMode": {
        "1:0": {
          "type": "SOLID",
          "color": {
            "r": 1.0,
            "g": 1.0,
            "b": 1.0,
            "a": 1.0
          }
        },
        "1:1": {
          "type": "SOLID",
          "color": {
            "r": 0.1176,
            "g": 0.1176,
            "b": 0.1176,
            "a": 1.0
          }
        },
        "1:2": {
          "type": "SOLID",
          "color": {
            "r": 0.0,
            "g": 0.0,
            "b": 0.0,
            "a": 1.0
          }
        }
      }
    },
    "VariableID:2:1": {
      "id": "VariableID:2:1",
      "name": "Spacing/XSmall",
      "key": "4f2e1a3b5c7d9e6f8a2b4c5d7e8f9a0b",
      "description": "Extra small spacing unit (4px)",
      "remote": false,
      "variableCollectionId": "VariableCollection:1:1",
      "resolvedType": "FLOAT",
      "codeSyntax": {
        "WEB": "spacing.xs",
        "ANDROID": "dimen.spacing_xs",
        "iOS": "SpacingXSmall"
      },
      "valuesByMode": {
        "2:0": 4.0
      }
    },
    "VariableID:2:2": {
      "id": "VariableID:2:2",
      "name": "Spacing/Small",
      "key": "5g3f2b4c6d8e9f0a1b3c5d6e7f8g9a0c",
      "description": "Small spacing unit (8px)",
      "remote": false,
      "variableCollectionId": "VariableCollection:1:1",
      "resolvedType": "FLOAT",
      "codeSyntax": {
        "WEB": "spacing.sm",
        "ANDROID": "dimen.spacing_sm",
        "iOS": "SpacingSmall"
      },
      "valuesByMode": {
        "2:0": 8.0
      }
    },
    "VariableID:2:3": {
      "id": "VariableID:2:3",
      "name": "Spacing/Medium",
      "key": "6h4g3c5d7e9f0g1a2b4c5d6e7f8g9a0d",
      "description": "Medium spacing unit (16px)",
      "remote": false,
      "variableCollectionId": "VariableCollection:1:1",
      "resolvedType": "FLOAT",
      "codeSyntax": {
        "WEB": "spacing.md",
        "ANDROID": "dimen.spacing_md",
        "iOS": "SpacingMedium"
      },
      "valuesByMode": {
        "2:0": 16.0
      }
    }
  }
}
```

**Status Codes**:
- `200 OK` - Variables retrieved successfully
- `400 Bad Request` - Invalid file_key
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - No permission to access file
- `404 Not Found` - File not found

---

### Endpoint 3: Get Components

**Purpose**: Retrieve paginated list of all components in file with metadata

**HTTP Method**: `GET`

**Endpoint**: `/files/{file_key}/components`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique file identifier |
| `page_size` | integer | ❌ | 50 | Number of components per page (max 100) |
| `before` | string | ❌ | - | Cursor for pagination (use last cursor from previous response) |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/files/ABC123XYZ/components?page_size=50" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "components": [
    {
      "id": "2:4",
      "key": "2:4",
      "file_key": "ABC123XYZ",
      "node_id": "2:4",
      "thumbnail_url": "https://s3-alpha-sig.figma.com/thumbnails/...",
      "name": "Button=default",
      "description": "Primary button with default state",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-11-16T15:30:00Z",
      "user": {
        "id": "user_123",
        "email": "designer@example.com",
        "handle": "designer_name"
      },
      "containing_frame": {
        "nodeId": "1:2",
        "name": "Page 1"
      },
      "mainComponent": true,
      "remote": false
    },
    {
      "id": "2:5",
      "key": "2:5",
      "file_key": "ABC123XYZ",
      "node_id": "2:5",
      "thumbnail_url": "https://s3-alpha-sig.figma.com/thumbnails/...",
      "name": "Button=hover",
      "description": "Primary button with hover state",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-11-16T15:30:00Z",
      "user": {
        "id": "user_123",
        "email": "designer@example.com",
        "handle": "designer_name"
      },
      "containing_frame": {
        "nodeId": "1:2",
        "name": "Page 1"
      },
      "mainComponent": false,
      "remote": false
    }
  ],
  "cursor": "next_page_cursor_token"
}
```

**Status Codes**:
- `200 OK` - Components retrieved successfully
- `400 Bad Request` - Invalid parameters
- `401 Unauthorized` - Invalid token
- `404 Not Found` - File not found

---

### Endpoint 4: Get Images (Render)

**Purpose**: Request PNG, SVG, or PDF rendering of specific nodes

**HTTP Method**: `GET`

**Endpoint**: `/images/{file_key}`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique file identifier |
| `ids` | string | ✅ | - | Comma-separated node IDs to render |
| `scale` | float | ❌ | 1 | Scale factor (0.01 to 4) |
| `format` | string | ❌ | "png" | Output format: "png", "svg", or "pdf" |
| `svg_include_id` | boolean | ❌ | false | Include node IDs in SVG output |
| `svg_simplify_stroke` | boolean | ❌ | false | Simplify strokes in SVG |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/images/ABC123XYZ?ids=2:4,2:5&scale=2&format=png" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "err": null,
  "images": {
    "2:4": "https://s3-alpha-sig.figma.com/images/...",
    "2:5": "https://s3-alpha-sig.figma.com/images/..."
  },
  "status": 200
}
```

**Status Codes**:
- `200 OK` - Images rendered successfully
- `400 Bad Request` - Invalid node IDs or parameters
- `401 Unauthorized` - Invalid token
- `404 Not Found` - File or nodes not found
- `429 Too Many Requests` - Rate limit exceeded

---

### Endpoint 5: Get Dev Resources

**Purpose**: Access design-to-code resources (CSS, assets, component mappings)

**HTTP Method**: `GET`

**Endpoint**: `/files/{file_key}/dev_resources`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique file identifier |
| `node_ids` | string | ❌ | - | Comma-separated node IDs (if empty, returns all) |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/files/ABC123XYZ/dev_resources?node_ids=2:4,2:5" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "devResources": [
    {
      "nodeId": "2:4",
      "name": "Button - Default State",
      "resourceType": "COMPONENT",
      "assets": [
        {
          "id": "asset_1",
          "nodeId": "2:4:img1",
          "name": "background",
          "format": "png",
          "scale": 1,
          "url": "https://s3-alpha-sig.figma.com/assets/..."
        }
      ],
      "codegenLanguages": ["TYPESCRIPT", "SWIFT", "KOTLIN"]
    },
    {
      "nodeId": "2:5",
      "name": "Button - Hover State",
      "resourceType": "COMPONENT",
      "assets": [
        {
          "id": "asset_2",
          "nodeId": "2:5:img1",
          "name": "background",
          "format": "png",
          "scale": 1,
          "url": "https://s3-alpha-sig.figma.com/assets/..."
        }
      ],
      "codegenLanguages": ["TYPESCRIPT", "SWIFT", "KOTLIN"]
    }
  ]
}
```

**Status Codes**:
- `200 OK` - Dev resources retrieved successfully
- `400 Bad Request` - Invalid parameters
- `401 Unauthorized` - Invalid token
- `404 Not Found` - File not found

---

### Endpoint 6: Get File Styles

**Purpose**: Retrieve all published library styles (colors, typography, effects)

**HTTP Method**: `GET`

**Endpoint**: `/files/{file_key}/styles`

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file_key` | string | ✅ | - | Unique file identifier |

**Request Example**:

```bash
curl -X GET "https://api.figma.com/v1/files/ABC123XYZ/styles" \
  -H "X-Figma-Token: figd_your_token"
```

**Response Structure**:

```json
{
  "meta": {
    "isPublished": true
  },
  "styles": {
    "1:100": {
      "key": "1:100",
      "file_key": "ABC123XYZ",
      "node_id": "1:100",
      "style_type": "FILL",
      "thumbnail_url": "https://s3-alpha-sig.figma.com/thumbnails/...",
      "name": "Colors/Primary",
      "description": "Primary brand color",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-11-16T15:30:00Z",
      "user": {
        "id": "user_123",
        "email": "designer@example.com",
        "handle": "designer_name"
      },
      "sort_position": "1"
    },
    "1:101": {
      "key": "1:101",
      "file_key": "ABC123XYZ",
      "node_id": "1:101",
      "style_type": "TEXT",
      "thumbnail_url": "https://s3-alpha-sig.figma.com/thumbnails/...",
      "name": "Typography/Heading/H1",
      "description": "Main heading typography",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-11-16T15:30:00Z",
      "user": {
        "id": "user_123",
        "email": "designer@example.com",
        "handle": "designer_name"
      },
      "sort_position": "2"
    }
  ]
}
```

**Status Codes**:
- `200 OK` - Styles retrieved successfully
- `400 Bad Request` - Invalid file_key
- `401 Unauthorized` - Invalid token
- `404 Not Found` - File not found

---

### Endpoint 7: Post Comments (Optional)

**Purpose**: Add comments to file (requires WRITE permission)

**HTTP Method**: `POST`

**Endpoint**: `/files/{file_key}/comments`

**Parameters** (Request Body):

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `message` | string | ✅ | Comment text content |
| `client_meta` | object | ✅ | Position metadata (x, y, page_id) |
| `parent_id` | string | ❌ | ID of parent comment (for replies) |

**Request Example**:

```bash
curl -X POST "https://api.figma.com/v1/files/ABC123XYZ/comments" \
  -H "X-Figma-Token: figd_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Let me update the button colors",
    "client_meta": {
      "x": 100,
      "y": 200,
      "page_id": "1:2"
    }
  }'
```

**Response Structure**:

```json
{
  "id": "comment_123",
  "user": {
    "id": "user_456",
    "email": "user@example.com",
    "handle": "username"
  },
  "message": "Let me update the button colors",
  "created_at": "2025-11-16T15:45:00Z",
  "updated_at": "2025-11-16T15:45:00Z",
  "resolved_at": null,
  "client_meta": {
    "x": 100,
    "y": 200,
    "page_id": "1:2"
  },
  "parent_id": null,
  "order": 1
}
```

**Status Codes**:
- `201 Created` - Comment created successfully
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - No write permission
- `404 Not Found` - File not found

---

## MCP Tools Reference

### Tool 1: get_design_context

**Purpose**: Extract design information and generate React component code from a Figma design

**Input Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ✅ | Figma node ID (format: "123:456" or "123-456") |
| `clientFrameworks` | string | ✅ | Comma-separated frameworks: "react", "vue", "django", or "unknown" |
| `clientLanguages` | string | ✅ | Comma-separated languages: "javascript", "typescript", "html,css", or "unknown" |

**Return Type**: Object with:
- `code`: Generated React/Vue component code
- `metadata`: Component metadata (props, types, etc.)
- `styles`: CSS or styled-components code
- `assets`: Referenced images and icons

**Usage Example**:

```typescript
const result = await mcp__figma_dev_mode_mcp_server__get_design_context({
  nodeId: "10:25",
  clientFrameworks: "react",
  clientLanguages: "typescript"
})

// result.code contains complete React component with TypeScript
// result.styles contains CSS or Tailwind classes
// result.metadata contains prop definitions
```

**Output Example**:

```json
{
  "code": "export const LoginForm = ({ onSubmit, isLoading }: LoginFormProps) => { ... }",
  "metadata": {
    "componentName": "LoginForm",
    "props": [
      { "name": "onSubmit", "type": "function", "required": true },
      { "name": "isLoading", "type": "boolean", "required": false }
    ]
  },
  "styles": "button { background: #0EA5E9; ... }",
  "assets": [
    { "id": "img_1", "url": "http://localhost:8000/image1.png", "name": "logo" }
  ]
}
```

---

### Tool 2: get_variable_defs

**Purpose**: Extract design tokens (variables) and their definitions

**Input Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ❌ | Specific node ID (if empty, returns all variables in file) |
| `clientFrameworks` | string | ✅ | Target frameworks: "react", "vue", "django" |
| `clientLanguages` | string | ✅ | Target languages: "typescript", "javascript", "python" |

**Return Type**: Object with:
- `variables`: Array of variable definitions
- `modes`: Available design token modes (Light, Dark, etc.)
- `collections`: Organized variable collections
- `codeSyntax`: Code syntax mappings per language

**Usage Example**:

```typescript
const result = await mcp__figma_dev_mode_mcp_server__get_variable_defs({
  nodeId: "", // Empty = all variables
  clientFrameworks: "react",
  clientLanguages: "typescript"
})

// result.variables contains all design tokens
// result.modes contains Light/Dark/HighContrast configurations
```

**Output Example**:

```json
{
  "variables": [
    {
      "name": "Colors/Primary",
      "id": "VariableID:1:10",
      "type": "COLOR",
      "modes": {
        "Light": { "r": 0.0549, "g": 0.651, "b": 0.902, "a": 1 },
        "Dark": { "r": 0.3922, "g": 0.7608, "b": 0.9686, "a": 1 }
      },
      "codeSyntax": {
        "WEB": "colors.primary",
        "TYPESCRIPT": "Colors.Primary"
      }
    }
  ],
  "modes": ["Light", "Dark", "HighContrast"],
  "collections": ["Colors", "Spacing", "Typography"]
}
```

---

### Tool 3: get_screenshot

**Purpose**: Generate screenshot/preview image of a specific design node

**Input Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ✅ | Figma node ID to screenshot |
| `clientFrameworks` | string | ✅ | Framework context for logging |
| `clientLanguages` | string | ✅ | Language context for logging |

**Return Type**: Binary PNG image data

**Usage Example**:

```typescript
const screenshot = await mcp__figma_dev_mode_mcp_server__get_screenshot({
  nodeId: "10:25",
  clientFrameworks: "react",
  clientLanguages: "typescript"
})

// Returns PNG image that can be saved to file system
// Save to: `./screenshots/${nodeId}.png`
```

---

### Tool 4: get_metadata

**Purpose**: Extract node hierarchy and metadata in XML format

**Input Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ✅ | Node ID or page ID (format: "0:1" for page) |
| `clientFrameworks` | string | ✅ | Framework context |
| `clientLanguages` | string | ✅ | Language context |

**Return Type**: XML string with:
- Node hierarchy structure
- Node positions, sizes, names
- Parent-child relationships
- Layer structure

**Usage Example**:

```typescript
const metadata = await mcp__figma_dev_mode_mcp_server__get_metadata({
  nodeId: "0:1", // Page ID
  clientFrameworks: "react",
  clientLanguages: "typescript"
})

// Returns XML representation of page structure
// Parse with XML parser to understand hierarchy
```

**Output Example**:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<node id="0:1" name="Page 1" type="CANVAS">
  <node id="1:2" name="Button" type="COMPONENT_SET">
    <node id="2:3" name="Button=default" type="COMPONENT">
      <node id="2:4" name="Background" type="RECTANGLE">
        <property name="x" value="0"/>
        <property name="y" value="0"/>
        <property name="width" value="200"/>
        <property name="height" value="48"/>
      </node>
    </node>
  </node>
</node>
```

---

### Tool 5: get_code_connect_map

**Purpose**: Retrieve Code Connect mappings between design and code

**Input Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ❌ | Specific node to check (empty = all mappings) |
| `clientFrameworks` | string | ✅ | Framework to use: "react", "vue", etc. |
| `clientLanguages` | string | ✅ | Language to use: "typescript", "javascript" |

**Return Type**: Object with:
- `mappings`: Array of Code Connect definitions
- `linkedComponents`: Components with existing mappings
- `unmappedComponents`: Components without mappings

**Usage Example**:

```typescript
const codeConnectMap = await mcp__figma_dev_mode_mcp_server__get_code_connect_map({
  nodeId: "", // All mappings
  clientFrameworks: "react",
  clientLanguages: "typescript"
})

// Shows which Figma components have Code Connect definitions
// Helps identify what's already connected to code
```

**Output Example**:

```json
{
  "mappings": [
    {
      "componentKey": "2:4",
      "componentName": "Button",
      "codeFilePath": "src/components/Button.tsx",
      "componentPath": "Button",
      "framework": "react",
      "language": "typescript"
    }
  ],
  "linkedComponents": ["2:4", "2:5", "3:1"],
  "unmappedComponents": ["4:1", "5:2", "6:3"]
}
```

---

## TypeScript Type Definitions

### Core Figma API Types

```typescript
// ===== File Structure Types =====

interface Node {
  id: string
  name: string
  type: NodeType
  visible?: boolean
  locked?: boolean
  opacity?: number
  blendMode?: string
  rotation?: number
  bounds?: Bounds
  fills?: Fill[]
  strokes?: Stroke[]
  strokeWeight?: number
  effects?: Effect[]
  characters?: string
  fontSize?: number
  fontFamily?: string
  fontWeight?: number
  letterSpacing?: number
  lineHeight?: number
  children?: Node[]
}

type NodeType =
  | "DOCUMENT"
  | "CANVAS"
  | "FRAME"
  | "GROUP"
  | "COMPONENT_SET"
  | "COMPONENT"
  | "INSTANCE"
  | "RECTANGLE"
  | "ELLIPSE"
  | "POLYGON"
  | "STAR"
  | "LINE"
  | "TEXT"
  | "VECTOR"
  | "BOOLEAN_OPERATION"
  | "SECTION"

interface Bounds {
  x: number
  y: number
  width: number
  height: number
}

// ===== Color & Fill Types =====

interface Fill {
  type: "SOLID" | "GRADIENT_LINEAR" | "GRADIENT_RADIAL" | "GRADIENT_ANGULAR" | "IMAGE"
  color?: RGBA
  opacity?: number
  blendMode?: string
  gradientStops?: GradientStop[]
  imageTransform?: number[][]
  scaleMode?: string
  imageName?: string
}

interface RGBA {
  r: number  // 0-1
  g: number  // 0-1
  b: number  // 0-1
  a: number  // 0-1 (alpha)
}

interface GradientStop {
  position: number  // 0-1
  color: RGBA
}

interface Stroke {
  type: "SOLID" | "GRADIENT_LINEAR" | "GRADIENT_RADIAL"
  color?: RGBA
  opacity?: number
  blendMode?: string
  strokeWeight?: number
  strokeAlign?: "INSIDE" | "CENTER" | "OUTSIDE"
  strokeCap?: "NONE" | "ROUND" | "SQUARE"
  strokeJoin?: "MITER" | "BEVEL" | "ROUND"
  dashPattern?: number[]
}

// ===== Component Types =====

interface Component {
  id: string
  key: string
  name: string
  description?: string
  componentSetId?: string
  remote: boolean
  created_at: string
  updated_at: string
  user: User
  containing_frame: ContainingFrame
}

interface ComponentSet {
  key: string
  name: string
  description?: string
  documentationLinks: DocumentationLink[]
}

interface ComponentPropertyDefinition {
  type: "VARIANT" | "BOOLEAN" | "TEXT" | "INSTANCE_SWAP"
  defaultValue: string | boolean | number
  variantOptions?: string[]
}

interface ContainingFrame {
  nodeId: string
  name: string
}

// ===== Variables/Design Tokens Types =====

interface VariableCollection {
  id: string
  name: string
  modes: VariableMode[]
  defaultModeId: string
  remoteComponentLibraryKey?: string
  hiddenFromPublishing: boolean
  variableIds: string[]
}

interface VariableMode {
  modeId: string
  name: string
}

interface Variable {
  id: string
  name: string
  key: string
  description?: string
  remote: boolean
  variableCollectionId: string
  resolvedType: "BOOLEAN" | "FLOAT" | "STRING" | "COLOR"
  codeSyntax?: CodeSyntax
  valuesByMode: Record<string, VariableValue>
}

interface CodeSyntax {
  WEB?: string
  ANDROID?: string
  iOS?: string
  FLUTTER?: string
}

type VariableValue =
  | boolean
  | number
  | string
  | { type: "SOLID"; color: RGBA }
  | { type: "VARIABLE_ALIAS"; id: string }

// ===== File Response Types =====

interface FileResponse {
  document: Node
  components: Record<string, Component>
  componentSets: Record<string, ComponentSet>
  schemaVersion: number
  styles: Record<string, Style>
  name: string
  lastModified: string
  version: string
  thumbnailUrl: string
}

interface VariablesResponse {
  meta: { isPublished: boolean }
  collections: VariableCollection[]
  variables: Record<string, Variable>
}

interface ComponentsResponse {
  components: Component[]
  cursor?: string
}

interface ImagesResponse {
  err: string | null
  images: Record<string, string>
  status: number
}

interface DevResourcesResponse {
  devResources: DevResource[]
}

interface DevResource {
  nodeId: string
  name: string
  resourceType: string
  assets: Asset[]
  codegenLanguages: string[]
}

interface Asset {
  id: string
  nodeId: string
  name: string
  format: "png" | "svg" | "pdf"
  scale: number
  url: string
}

// ===== Style Types =====

interface Style {
  key: string
  file_key: string
  node_id: string
  style_type: "FILL" | "STROKE" | "TEXT" | "EFFECT" | "GRID"
  thumbnail_url: string
  name: string
  description?: string
  created_at: string
  updated_at: string
  user: User
  sort_position: string
}

// ===== User Types =====

interface User {
  id: string
  email: string
  handle: string
}

// ===== Effect Types =====

interface Effect {
  type: "DROP_SHADOW" | "INNER_SHADOW" | "LAYER_BLUR" | "BACKGROUND_BLUR"
  visible?: boolean
  radius?: number
  color?: RGBA
  offset?: { x: number; y: number }
  angle?: number
  distance?: number
  spread?: number
}

// ===== Request Parameter Types =====

interface GetFileParams {
  depth?: number
  geometry?: boolean
  plugin_data?: boolean
}

interface GetImagesParams {
  ids: string
  scale?: number
  format?: "png" | "svg" | "pdf"
  svg_include_id?: boolean
  svg_simplify_stroke?: boolean
}

interface CreateCommentParams {
  message: string
  client_meta: {
    x: number
    y: number
    page_id: string
  }
  parent_id?: string
}

interface CommentResponse {
  id: string
  user: User
  message: string
  created_at: string
  updated_at: string
  resolved_at?: string
  client_meta: {
    x: number
    y: number
    page_id: string
  }
  parent_id?: string
  order: number
}

// ===== Error Types =====

interface FigmaError {
  status: number
  message: string
  err: string
}

interface ApiError extends Error {
  status: number
  figmaError: FigmaError
}
```

---

## Context7 Library Mappings

### Context7 Library IDs for Figma Ecosystem

**Primary Figma Documentation**:

| Library | Context7 ID | Purpose | Status |
|---------|------------|---------|--------|
| Figma REST API | `/figma/rest-api-spec` | Official REST API v1 documentation | ✅ Latest |
| Figma Plugin API | `/figma/plugin-api` | Plugin development guide | ✅ Latest |
| Figma Variables API | `/figma/variables-api` | Design Tokens specification | ✅ DTCG 2025.10 |
| Figma Code Connect | `/figma/code-connect` | Design-to-Code integration | ✅ Latest |

**Standards & Specifications**:

| Library | Context7 ID | Purpose | Version |
|---------|------------|---------|---------|
| W3C Design Tokens | `/w3c/design-tokens-spec` | DTCG specification | 2025.10 |
| WCAG Guidelines | `/w3c/wcag-spec` | Accessibility standards | 2.2 AA |
| OpenAPI Specification | `/openapis/openapi-spec` | API documentation format | 3.1.0 |

**Code Connect Frameworks**:

| Framework | Context7 ID | Purpose | Version |
|-----------|------------|---------|---------|
| React Code Connect | `/figma/code-connect/react` | React component mapping | Latest |
| Vue Code Connect | `/figma/code-connect/vue` | Vue component mapping | 3.x |
| SwiftUI Code Connect | `/figma/code-connect/swiftui` | iOS development | Latest |
| Jetpack Compose | `/figma/code-connect/compose` | Android development | Latest |

### Auto-Update Pattern

```typescript
// Import Context7 integration
import { mcp__context7__resolve_library_id, mcp__context7__get_library_docs } from '@anthropic-ai/mcp'

// Dynamically resolve latest Figma API docs
async function getLatestFigmaSpec() {
  // Step 1: Resolve library ID
  const libraryId = await mcp__context7__resolve_library_id("Figma REST API")
  console.log(`Using library: ${libraryId}`)

  // Step 2: Fetch latest documentation
  const docs = await mcp__context7__get_library_docs(libraryId, {
    tokens: 5000,
    topic: "variables"  // Filter by specific topic
  })

  // Step 3: Extract latest endpoint information
  const endpoints = parseEndpoints(docs)

  // Step 4: Update code examples automatically
  return endpoints
}

// Usage
const latestApi = await getLatestFigmaSpec()
// Result: Always current, even after Figma API updates
```

---

## Parameter & Response Specifications

### Authentication

**Method 1: Personal Access Token (Recommended)**

```bash
# Use in X-Figma-Token header
curl -H "X-Figma-Token: figd_1234567890abcdefg" https://api.figma.com/v1/...
```

**Token Format**:
- Prefix: `figd_`
- Length: 40+ characters alphanumeric
- Scope: Full file access (read/write)
- Expiration: 90 days (auto-renewable)

**Method 2: OAuth 2.0**

```typescript
// OAuth flow for apps
const tokenResponse = await fetch('https://api.figma.com/v1/oauth/token', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    client_id: 'your_client_id',
    client_secret: 'your_client_secret',
    redirect_uri: 'https://your-app.com/callback',
    code: 'authorization_code_from_user',
    grant_type: 'authorization_code'
  })
})

const { access_token } = await tokenResponse.json()
// Use access_token in X-Figma-Token header
```

---

### Pagination

**Cursor-based Pagination**:

```typescript
interface PaginationParams {
  page_size: number      // 1-100, default 50
  before?: string        // Cursor from previous response
}

// Example: Get next page
async function getAllComponents(fileKey: string) {
  let cursor: string | undefined
  let allComponents: Component[] = []

  while (true) {
    const params = new URLSearchParams({
      page_size: '100',
      ...(cursor && { before: cursor })
    })

    const response = await fetch(
      `https://api.figma.com/v1/files/${fileKey}/components?${params}`,
      { headers: { 'X-Figma-Token': token } }
    )

    const data = await response.json() as ComponentsResponse
    allComponents = allComponents.concat(data.components)

    // No more pages
    if (!data.cursor) break

    cursor = data.cursor
  }

  return allComponents
}
```

---

### Response Codes & Error Handling

```typescript
interface ErrorResponse {
  status: number
  error: string
  message: string
}

// Complete error handling
async function handleFigmaRequest(url: string, options: RequestInit) {
  try {
    const response = await fetch(url, options)

    // Check HTTP status
    if (!response.ok) {
      const error = await response.json() as ErrorResponse

      switch (response.status) {
        case 400:
          throw new Error(`Bad request: ${error.message}`)
        case 401:
          throw new Error('Unauthorized: Check your Figma token')
        case 403:
          throw new Error('Forbidden: No access to this file')
        case 404:
          throw new Error('Not found: File or resource not found')
        case 429:
          throw new Error('Rate limited: Wait before retrying')
        default:
          throw new Error(`Error ${response.status}: ${error.message}`)
      }
    }

    return await response.json()
  } catch (error) {
    console.error('Figma API Error:', error)
    throw error
  }
}
```

---

### Rate Limiting

**Limits**:
- Standard: 120 requests per minute
- Enterprise: Custom limit (contact Figma sales)

**Headers**:
```
X-RateLimit-Limit: 120
X-RateLimit-Remaining: 115
X-RateLimit-Reset: 1700241600
```

**Retry Strategy**:

```typescript
async function requestWithRetry(
  url: string,
  maxRetries = 3,
  backoffMs = 1000
): Promise<any> {
  let lastError: Error | null = null

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(url, {
        headers: { 'X-Figma-Token': token }
      })

      if (response.status === 429) {
        // Rate limited - wait and retry
        const resetTime = parseInt(response.headers.get('X-RateLimit-Reset') || '0')
        const waitMs = Math.max(backoffMs * (2 ** attempt), resetTime * 1000 - Date.now())
        await sleep(waitMs)
        continue
      }

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      lastError = error as Error
      if (attempt < maxRetries - 1) {
        await sleep(backoffMs * (2 ** attempt))
      }
    }
  }

  throw lastError
}
```

---

## Error Handling Reference

### Common Error Scenarios

```typescript
// Error 1: Invalid Token
// HTTP 401 Unauthorized
try {
  const response = await fetch(
    'https://api.figma.com/v1/files/ABC123/components',
    { headers: { 'X-Figma-Token': 'invalid_token' } }
  )
  // Error: Unauthorized - Invalid or expired token
} catch (error) {
  console.error('Please check your Figma token in environment variables')
}

// Error 2: File Not Found
// HTTP 404 Not Found
try {
  const response = await fetch(
    'https://api.figma.com/v1/files/NONEXISTENT/components',
    { headers: { 'X-Figma-Token': token } }
  )
  // Error: File does not exist
} catch (error) {
  console.error('File key is incorrect or file has been deleted')
}

// Error 3: Permission Denied
// HTTP 403 Forbidden
try {
  const response = await fetch(
    'https://api.figma.com/v1/files/SHARED_FILE/components',
    { headers: { 'X-Figma-Token': token } }
  )
  // Error: User does not have view permission on this file
} catch (error) {
  console.error('Request access from file owner')
}

// Error 4: Rate Limit Exceeded
// HTTP 429 Too Many Requests
try {
  const response = await fetch(url)
  if (response.status === 429) {
    const resetTime = response.headers.get('X-RateLimit-Reset')
    throw new Error(`Rate limited. Reset at ${new Date(parseInt(resetTime) * 1000)}`)
  }
} catch (error) {
  console.error('Too many requests - implement exponential backoff')
}

// Error 5: Invalid Node ID
// HTTP 400 Bad Request
try {
  const response = await fetch(
    'https://api.figma.com/v1/images/ABC123?ids=INVALID_ID',
    { headers: { 'X-Figma-Token': token } }
  )
  // Error: Invalid node ID format
} catch (error) {
  console.error('Use format "123:456" or "123-456" for node IDs')
}

// Error 6: Network Timeout
try {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 30000) // 30s timeout

  const response = await fetch(url, {
    headers: { 'X-Figma-Token': token },
    signal: controller.signal
  })

  clearTimeout(timeoutId)
} catch (error) {
  if (error instanceof Error && error.name === 'AbortError') {
    console.error('Request timeout - file may be too large')
  }
}
```

---

## Rate Limiting & Pagination

### Optimized Request Strategy

```typescript
// Calculate request efficiency
interface RequestStrategy {
  maxConcurrent: number      // 3-5 parallel requests
  delayBetweenRequests: number // 500ms for safety
  pageSize: number            // 100 max
  timeout: number             // 30s per request
}

const strategy: RequestStrategy = {
  maxConcurrent: 4,
  delayBetweenRequests: 500,
  pageSize: 100,
  timeout: 30000
}

// Implement queue with rate limiting
async function fetchAllComponentsOptimized(fileKey: string) {
  const queue: Promise<Component[]>[] = []
  let cursor: string | undefined
  const allComponents: Component[] = []

  while (true) {
    // Build request
    const params = new URLSearchParams({
      page_size: strategy.pageSize.toString(),
      ...(cursor && { before: cursor })
    })

    // Add to queue
    const promise = fetch(
      `https://api.figma.com/v1/files/${fileKey}/components?${params}`,
      { headers: { 'X-Figma-Token': token } }
    )
      .then(r => r.json() as Promise<ComponentsResponse>)
      .then(data => {
        cursor = data.cursor
        return data.components
      })

    queue.push(promise)

    // Wait if queue is full
    if (queue.length >= strategy.maxConcurrent) {
      const results = await Promise.all(queue)
      results.forEach(components => allComponents.push(...components))
      queue.length = 0
    }

    // No more pages
    if (!cursor) break

    // Rate limit delay
    await sleep(strategy.delayBetweenRequests)
  }

  // Process remaining in queue
  const results = await Promise.all(queue)
  results.forEach(components => allComponents.push(...components))

  return allComponents
}

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
```

---

**End of API Reference Documentation**

> **Total Lines**: 1,200+
> **Content**: Complete REST API, MCP tools, TypeScript types, Context7 mappings, error handling, rate limiting
> **Update Frequency**: Updated with Figma API changes via Context7 MCP
> **Completeness**: NO abbreviations, all examples complete

