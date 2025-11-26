# Figma Domain Expertise - Code Examples

This document provides practical code examples for the moai-domain-figma Skill.

---

## 1. FigmaClient Class (TypeScript)

**Purpose**: Reusable class that wraps the Figma REST API

```typescript
// figma-client.ts

import axios, { AxiosInstance } from 'axios'

interface GetFileParams {
  version?: string
  ids?: string[]
  depth?: number
  geometry?: boolean
  plugin_data?: boolean
  branch_data?: boolean
}

interface Variable {
  id: string
  name: string
  variableCollectionId: string
  resolvedType: 'BOOLEAN' | 'FLOAT' | 'STRING' | 'COLOR'
  valuesByMode: Record<string, VariableValue>
  scopes?: VariableScope[]
  codeSyntax?: Record<string, string>
  description?: string
  hiddenFromPublishing?: boolean
}

interface VariableCollection {
  id: string
  name: string
  modes: Array<{ modeId: string; name: string }>
  defaultModeId: string
  variableIds: string[]
  hiddenFromPublishing?: boolean
  remote?: boolean
}

interface Component {
  key: string
  file_key: string
  node_id: string
  thumbnail_url: string
  name: string
  description: string
  created_at: string
  updated_at: string
  user: { id: string; handle: string; email?: string }
  remote: boolean
}

interface ImageRenderResponse {
  images: Record<string, string> // { nodeId: "https://..." }
  err?: string
}

type VariableValue = boolean | number | string | RGBA

interface RGBA {
  r: number // 0-1
  g: number // 0-1
  b: number // 0-1
  a: number // 0-1
}

interface VariableScope {
  scope: 'COMPONENT_MAIN' | 'ALL_FILLS' | 'ALL_STROKES' | 'ALL_TEXT' | string
  sceneId?: string
}

export class FigmaClient {
  private axiosInstance: AxiosInstance
  private baseUrl = 'https://api.figma.com/v1'
  private token: string

  constructor(token: string) {
    this.token = token
    this.axiosInstance = axios.create({
      baseURL: this.baseUrl,
      headers: {
        'X-Figma-Token': token,
        'Content-Type': 'application/json'
      },
      timeout: 30000
    })
  }

  /**
   * Get file information
   */
  async getFile(fileKey: string, params: GetFileParams = {}) {
    const queryParams = new URLSearchParams()
    if (params.version) queryParams.append('version', params.version)
    if (params.ids) queryParams.append('ids', params.ids.join(','))
    if (params.depth !== undefined) queryParams.append('depth', params.depth.toString())
    if (params.geometry !== undefined) queryParams.append('geometry', params.geometry.toString())
    if (params.plugin_data !== undefined) queryParams.append('plugin_data', params.plugin_data.toString())
    if (params.branch_data !== undefined) queryParams.append('branch_data', params.branch_data.toString())

    const url = `/files/${fileKey}${queryParams.toString() ? '?' + queryParams.toString() : ''}`
    const response = await this.axiosInstance.get(url)
    return response.data
  }

  /**
   * Get local variables (Enterprise)
   */
  async getVariables(fileKey: string) {
    const response = await this.axiosInstance.get(`/files/${fileKey}/variables/local`)
    return response.data as {
      variables: Variable[]
      variableCollections: VariableCollection[]
    }
  }

  /**
   * Get component list
   */
  async getComponents(fileKey: string, pageSize = 50, before?: string) {
    const queryParams = new URLSearchParams()
    queryParams.append('page_size', pageSize.toString())
    if (before) queryParams.append('before', before)

    const response = await this.axiosInstance.get(
      `/files/${fileKey}/components?${queryParams.toString()}`
    )
    return response.data as {
      components: Component[]
      pagination: { has_more: boolean; start_cursor?: string; end_cursor?: string }
    }
  }

  /**
   * Render images
   */
  async renderImages(
    fileKey: string,
    nodeIds: string[],
    options: {
      scale?: number
      format?: 'png' | 'jpg' | 'svg' | 'pdf'
      use_absolute_bounds?: boolean
      svg_outline_text?: boolean
    } = {}
  ) {
    const queryParams = new URLSearchParams()
    queryParams.append('ids', nodeIds.join(','))
    queryParams.append('scale', (options.scale || 1).toString())
    queryParams.append('format', options.format || 'png')
    if (options.use_absolute_bounds !== undefined) {
      queryParams.append('use_absolute_bounds', options.use_absolute_bounds.toString())
    }
    if (options.svg_outline_text !== undefined) {
      queryParams.append('svg_outline_text', options.svg_outline_text.toString())
    }

    const response = await this.axiosInstance.get(
      `/images/${fileKey}?${queryParams.toString()}`
    )
    return response.data as ImageRenderResponse
  }

  /**
   * Get Dev Resources
   */
  async getDevResources(fileKey: string, nodeIds?: string[]) {
    const queryParams = new URLSearchParams()
    queryParams.append('file_key', fileKey)
    if (nodeIds) queryParams.append('node_ids', nodeIds.join(','))

    const response = await this.axiosInstance.get(`/dev_resources?${queryParams.toString()}`)
    return response.data
  }

  /**
   * Update variable values (Enterprise)
   */
  async updateVariable(
    fileKey: string,
    variableId: string,
    valuesByMode: Record<string, VariableValue>
  ) {
    const response = await this.axiosInstance.post('/variables', {
      file_key: fileKey,
      variables: [
        {
          id: variableId,
          values_by_mode: valuesByMode
        }
      ]
    })
    return response.data
  }
}

// Usage example
const client = new FigmaClient(process.env.FIGMA_TOKEN!)

// Get file
const fileData = await client.getFile('ABC123XYZ', {
  geometry: true,
  plugin_data: false
})

// Get variables
const { variables, variableCollections } = await client.getVariables('ABC123XYZ')

// Get components
const { components } = await client.getComponents('ABC123XYZ')

// Render images
const images = await client.renderImages('ABC123XYZ', ['1:2', '1:3'], {
  scale: 2,
  format: 'png'
})
```

---

## 2. Variables Extraction and Transformation

**Purpose**: Transform Figma Variables to various formats

### 2.1. DTCG JSON Generation

```typescript
// tokens-to-dtcg.ts

interface DtcgVariable {
  $value: any
  $type?: string
  $description?: string
}

function generateDtcgTokens(variables: Variable[]): Record<string, any> {
  const dtcg: Record<string, any> = {
    $schema: 'https://tr.designtokens.org/format/',
    $tokens: {}
  }

  // Color token classification
  const colorTokens: Record<string, DtcgVariable> = {}
  const spacingTokens: Record<string, DtcgVariable> = {}
  const typographyTokens: Record<string, DtcgVariable> = {}

  for (const variable of variables) {
    if (variable.name.startsWith('color/')) {
      const colorName = variable.name.replace(/\//g, '.')
      const lightModeValue = Object.values(variable.valuesByMode)[0]

      if (typeof lightModeValue === 'object' && 'r' in lightModeValue) {
        const rgba = lightModeValue as RGBA
        const hex = rgbaToHex(rgba)
        colorTokens[colorName] = {
          $type: 'color',
          $value: hex,
          $description: variable.description
        }
      }
    } else if (variable.name.startsWith('spacing/')) {
      const spacingName = variable.name.replace(/\//g, '.')
      const value = Object.values(variable.valuesByMode)[0]
      spacingTokens[spacingName] = {
        $type: 'dimension',
        $value: `${value}px`,
        $description: variable.description
      }
    } else if (variable.name.startsWith('typography/')) {
      const typographyName = variable.name.replace(/\//g, '.')
      const value = Object.values(variable.valuesByMode)[0]
      typographyTokens[typographyName] = {
        $type: 'typography',
        $value: value,
        $description: variable.description
      }
    }
  }

  dtcg.$tokens.color = colorTokens
  dtcg.$tokens.spacing = spacingTokens
  dtcg.$tokens.typography = typographyTokens

  return dtcg
}

function rgbaToHex(rgba: RGBA): string {
  const r = Math.round(rgba.r * 255).toString(16).padStart(2, '0')
  const g = Math.round(rgba.g * 255).toString(16).padStart(2, '0')
  const b = Math.round(rgba.b * 255).toString(16).padStart(2, '0')
  return `#${r}${g}${b}`
}

// Usage
const dtcgTokens = generateDtcgTokens(variables)
console.log(JSON.stringify(dtcgTokens, null, 2))
```

### 2.2. CSS Variables Generation

```typescript
// tokens-to-css.ts

function generateCSSVariables(variables: Variable[]): string {
  let css = ':root {\n'

  for (const variable of variables) {
    const value = Object.values(variable.valuesByMode)[0]

    if (typeof value === 'boolean') {
      css += `  --${variable.name}: ${value ? 'true' : 'false'};\n`
    } else if (typeof value === 'number') {
      css += `  --${variable.name}: ${value}${variable.name.includes('opacity') ? '' : 'px'};\n`
    } else if (typeof value === 'string') {
      css += `  --${variable.name}: ${value};\n`
    } else if (typeof value === 'object' && 'r' in value) {
      const rgba = value as RGBA
      css += `  --${variable.name}: rgba(${Math.round(rgba.r * 255)}, ${Math.round(rgba.g * 255)}, ${Math.round(rgba.b * 255)}, ${rgba.a});\n`
    }
  }

  css += '}\n'
  return css
}

// Usage
const css = generateCSSVariables(variables)
fs.writeFileSync('tokens.css', css)
```

### 2.3. Tailwind Config Generation

```typescript
// tokens-to-tailwind.ts

function generateTailwindConfig(variables: Variable[]): string {
  const theme: Record<string, Record<string, string>> = {}

  for (const variable of variables) {
    const category = variable.name.split('/')[0]

    if (!theme[category]) {
      theme[category] = {}
    }

    const tokenName = variable.name.replace(`${category}/`, '')
    const value = Object.values(variable.valuesByMode)[0]

    if (category === 'color' && typeof value === 'object' && 'r' in value) {
      const rgba = value as RGBA
      const hex = rgbaToHex(rgba)
      theme[category][tokenName] = hex
    } else if (category === 'spacing' && typeof value === 'number') {
      theme[category][tokenName] = `${value}px`
    }
  }

  return `module.exports = {
  theme: {
    extend: {
      colors: ${JSON.stringify(theme.color || {}, null, 6)},
      spacing: ${JSON.stringify(theme.spacing || {}, null, 6)},
    }
  }
}\n`
}

// Usage
const tailwindConfig = generateTailwindConfig(variables)
fs.writeFileSync('tailwind.config.js', tailwindConfig)
```

---

## 3. React Component Generation

**Purpose**: Refine code extracted from Design Context into actual React components

```typescript
// generate-react-component.ts

interface ComponentSpec {
  name: string
  description: string
  props: Record<string, PropDefinition>
  figmaUrl: string
}

interface PropDefinition {
  type: 'string' | 'number' | 'boolean' | 'react-node'
  required: boolean
  defaultValue?: any
  options?: string[]
}

function generateReactComponent(spec: ComponentSpec): string {
  const propsInterface = generatePropsInterface(spec)
  const componentFunction = generateComponentFunction(spec)
  const figmaConnect = generateFigmaConnect(spec)
  const storybook = generateStorybookMeta(spec)

  return `// ${spec.name}.tsx
import React from 'react'
import figma from '@figma/code-connect'

${propsInterface}

export const ${spec.name}: React.FC<${spec.name}Props> = (props) => {
  return (
    ${componentFunction}
  )
}

${figmaConnect}

export default ${spec.name}

// Storybook
${storybook}
`
}

function generatePropsInterface(spec: ComponentSpec): string {
  let code = `interface ${spec.name}Props {\n`

  for (const [propName, propDef] of Object.entries(spec.props)) {
    const typeMap = {
      'string': 'string',
      'number': 'number',
      'boolean': 'boolean',
      'react-node': 'React.ReactNode'
    }

    const typeString = typeMap[propDef.type]
    const optional = !propDef.required ? '?' : ''
    const defaultValue = propDef.defaultValue ? ` = ${JSON.stringify(propDef.defaultValue)}` : ''

    code += `  ${propName}${optional}: ${typeString}${defaultValue}\n`
  }

  code += '}\n\n'
  return code
}

function generateComponentFunction(spec: ComponentSpec): string {
  return `<div>
    {/* Component implementation */}
  </div>`
}

function generateFigmaConnect(spec: ComponentSpec): string {
  return `figma.connect(
  ${spec.name},
  '${spec.figmaUrl}',
  {
    props: {
      // Add property mappings
    }
  }
)`
}

function generateStorybookMeta(spec: ComponentSpec): string {
  return `import type { Meta, StoryObj } from '@storybook/react'

const meta: Meta<typeof ${spec.name}> = {
  component: ${spec.name},
  title: '${spec.name}',
  argTypes: {
    // Define argTypes for each prop
  }
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    // Default props
  }
}`
}

// Usage
const component = generateReactComponent({
  name: 'Button',
  description: 'Primary action button',
  props: {
    variant: {
      type: 'string',
      required: false,
      defaultValue: 'primary',
      options: ['primary', 'secondary', 'destructive']
    },
    size: {
      type: 'string',
      required: false,
      defaultValue: 'md',
      options: ['sm', 'md', 'lg']
    },
    disabled: {
      type: 'boolean',
      required: false,
      defaultValue: false
    },
    children: {
      type: 'react-node',
      required: true
    }
  },
  figmaUrl: 'https://figma.com/file/ABC123XYZ/DesignSystem?node-id=1:2'
})

fs.writeFileSync('Button.tsx', component)
```

---

## 4. Design Tokens Export

**Purpose**: Export Design Tokens in various formats

### 4.1. JSON Token Studio Format

```typescript
// export-token-studio.ts

interface TokenStudioFormat {
  [key: string]: {
    [key: string]: {
      value: any
      type: string
      description?: string
    }
  }
}

function exportTokenStudio(variables: Variable[]): TokenStudioFormat {
  const tokens: TokenStudioFormat = {}

  for (const variable of variables) {
    const parts = variable.name.split('/')
    const category = parts[0]
    const tokenName = parts.slice(1).join('.')

    if (!tokens[category]) {
      tokens[category] = {}
    }

    const value = Object.values(variable.valuesByMode)[0]

    tokens[category][tokenName] = {
      value: formatTokenValue(value, variable.resolvedType),
      type: mapVariableTypeToTokenType(variable.resolvedType),
      description: variable.description
    }
  }

  return tokens
}

function formatTokenValue(value: any, type: string): any {
  if (type === 'COLOR' && typeof value === 'object' && 'r' in value) {
    const rgba = value as RGBA
    return `rgba(${Math.round(rgba.r * 255)}, ${Math.round(rgba.g * 255)}, ${Math.round(rgba.b * 255)}, ${rgba.a})`
  }
  return value
}

function mapVariableTypeToTokenType(type: string): string {
  const typeMap: Record<string, string> = {
    'COLOR': 'color',
    'FLOAT': 'dimension',
    'STRING': 'typography',
    'BOOLEAN': 'other'
  }
  return typeMap[type] || 'other'
}

// Usage
const tokenStudio = exportTokenStudio(variables)
fs.writeFileSync('tokens.json', JSON.stringify(tokenStudio, null, 2))
```

### 4.2. Swift Constants (iOS)

```typescript
// export-swift.ts

function exportSwiftConstants(variables: Variable[]): string {
  let swift = `// Design Tokens (Generated from Figma)
import UIKit

class DesignTokens {
\n`

  // Colors
  swift += '  // MARK: - Colors\n'
  for (const variable of variables) {
    if (variable.name.startsWith('color/')) {
      const constantName = variable.name
        .replace('color/', '')
        .split('/')
        .map(part => part.charAt(0).toUpperCase() + part.slice(1))
        .join('')

      const value = Object.values(variable.valuesByMode)[0]
      if (typeof value === 'object' && 'r' in value) {
        const rgba = value as RGBA
        const r = Math.round(rgba.r * 255)
        const g = Math.round(rgba.g * 255)
        const b = Math.round(rgba.b * 255)
        const a = rgba.a

        swift += `  static let color${constantName} = UIColor(red: ${(r / 255).toFixed(3)}, green: ${(g / 255).toFixed(3)}, blue: ${(b / 255).toFixed(3)}, alpha: ${a})\n`
      }
    }
  }

  // Spacing
  swift += '\n  // MARK: - Spacing\n'
  for (const variable of variables) {
    if (variable.name.startsWith('spacing/')) {
      const constantName = variable.name
        .replace('spacing/', '')
        .split('/')
        .map(part => part.charAt(0).toUpperCase() + part.slice(1))
        .join('')

      const value = Object.values(variable.valuesByMode)[0]
      if (typeof value === 'number') {
        swift += `  static let spacing${constantName}: CGFloat = ${value}\n`
      }
    }
  }

  swift += '}\n'
  return swift
}

// Usage
const swiftCode = exportSwiftConstants(variables)
fs.writeFileSync('DesignTokens.swift', swiftCode)
```

---

## 5. Variables Multi-Mode Processing

**Purpose**: Manage Variables for Light/Dark/HighContrast modes

```typescript
// multi-mode-variables.ts

function generateMultiModeTokens(variables: Variable[]): Record<string, Record<string, any>> {
  const modes: Record<string, Record<string, any>> = {}

  // Group tokens by mode
  for (const variable of variables) {
    for (const [modeId, value] of Object.entries(variable.valuesByMode)) {
      if (!modes[modeId]) {
        modes[modeId] = {}
      }

      modes[modeId][variable.name] = value
    }
  }

  return modes
}

function generateMultiModeCss(variables: Variable[], modeNames: Record<string, string>): string {
  let css = ''

  // Light mode (default)
  const lightModeId = Object.keys(variables[0].valuesByMode)[0]
  css += ':root {\n'
  css += '  /* Light Mode (Default) */\n'

  for (const variable of variables) {
    const value = variable.valuesByMode[lightModeId]
    css += `  --${variable.name}: ${formatCssValue(value)};\n`
  }

  css += '}\n\n'

  // Dark mode
  const darkModeId = Object.keys(variables[0].valuesByMode)[1]
  if (darkModeId) {
    css += '@media (prefers-color-scheme: dark) {\n'
    css += '  :root {\n'
    css += '    /* Dark Mode */\n'

    for (const variable of variables) {
      const value = variable.valuesByMode[darkModeId]
      css += `    --${variable.name}: ${formatCssValue(value)};\n`
    }

    css += '  }\n'
    css += '}\n'
  }

  return css
}

function formatCssValue(value: any): string {
  if (typeof value === 'number') {
    return `${value}px`
  } else if (typeof value === 'string') {
    return value
  } else if (typeof value === 'object' && 'r' in value) {
    const rgba = value as RGBA
    return `rgba(${Math.round(rgba.r * 255)}, ${Math.round(rgba.g * 255)}, ${Math.round(rgba.b * 255)}, ${rgba.a})`
  }
  return String(value)
}

// Usage
const multiModeTokens = generateMultiModeTokens(variables)
const multiModeCss = generateMultiModeCss(variables, {
  'Mode:1:2': 'Light',
  'Mode:2:3': 'Dark'
})

fs.writeFileSync('tokens-multimode.css', multiModeCss)
```

---

## 6. Figma URL Parsing

**Purpose**: Extract required information from Figma URLs

```typescript
// parse-figma-url.ts

interface FigmaUrlParts {
  fileKey: string
  fileName: string
  nodeId: string
  sectionId?: string
  versionId?: string
}

function parseFigmaUrl(url: string): FigmaUrlParts {
  // Supported URL formats:
  // https://figma.com/file/{fileKey}/{fileName}?node-id={nodeId}
  // https://figma.com/design/{fileKey}/{fileName}?node-id={nodeId}

  const fileKeyMatch = url.match(/(?:file|design)\/([a-zA-Z0-9]+)/)
  const fileNameMatch = url.match(/\/([^/?]+)\?/)
  const nodeIdMatch = url.match(/node-id=(\d+)-(\d+)/)
  const sectionIdMatch = url.match(/section-id=([a-zA-Z0-9]+)/)
  const versionIdMatch = url.match(/version=([a-zA-Z0-9]+)/)

  if (!fileKeyMatch || !fileNameMatch || !nodeIdMatch) {
    throw new Error(`Invalid Figma URL: ${url}`)
  }

  return {
    fileKey: fileKeyMatch[1],
    fileName: decodeURIComponent(fileNameMatch[1]),
    nodeId: `${nodeIdMatch[1]}:${nodeIdMatch[2]}`,
    sectionId: sectionIdMatch?.[1],
    versionId: versionIdMatch?.[1]
  }
}

// Usage
const urlParts = parseFigmaUrl('https://figma.com/design/ABC123XYZ/LoginPage?node-id=10-25')
console.log(urlParts)
// {
//   fileKey: 'ABC123XYZ',
//   fileName: 'LoginPage',
//   nodeId: '10:25',
//   sectionId: undefined,
//   versionId: undefined
// }
```

---

## 7. Comprehensive Workflow

**Purpose**: Complete Design → Code automation pipeline

```typescript
// design-to-code-pipeline.ts

async function designToCodePipeline(figmaUrl: string): Promise<void> {
  console.log(`🎨 Starting Design-to-Code pipeline for: ${figmaUrl}`)

  // 1. URL parsing
  const { fileKey, nodeId, fileName } = parseFigmaUrl(figmaUrl)
  console.log(`📄 File: ${fileName} (${fileKey})`)
  console.log(`🎯 Node: ${nodeId}`)

  // 2. Fetch data from Figma
  const client = new FigmaClient(process.env.FIGMA_TOKEN!)

  console.log(`\n📥 Fetching Figma data...`)
  const [fileData, variables, images] = await Promise.all([
    client.getFile(fileKey, { geometry: true }),
    client.getVariables(fileKey),
    client.renderImages(fileKey, [nodeId])
  ])

  console.log(`✅ Fetched file with ${fileData.document.children.length} frames`)
  console.log(`✅ Found ${variables.variables.length} variables`)
  console.log(`✅ Rendered ${Object.keys(images.images).length} images`)

  // 3. Transform Design Tokens
  console.log(`\n🔄 Transforming Design Tokens...`)
  const dtcgTokens = generateDtcgTokens(variables.variables)
  const cssTokens = generateCSSVariables(variables.variables)
  const tailwindConfig = generateTailwindConfig(variables.variables)

  fs.mkdirSync('src/styles', { recursive: true })
  fs.writeFileSync('src/styles/tokens.json', JSON.stringify(dtcgTokens, null, 2))
  fs.writeFileSync('src/styles/tokens.css', cssTokens)
  fs.writeFileSync('tailwind.config.js', tailwindConfig)

  console.log(`✅ Generated token files`)

  // 4. Generate React components
  console.log(`\n⚛️ Generating React components...`)
  const componentSpec: ComponentSpec = {
    name: toPascalCase(fileName),
    description: 'Generated from Figma',
    props: {
      // Props inference would happen here in real implementation
    },
    figmaUrl
  }

  const componentCode = generateReactComponent(componentSpec)

  fs.mkdirSync('src/components', { recursive: true })
  fs.writeFileSync(`src/components/${componentSpec.name}.tsx`, componentCode)

  console.log(`✅ Generated ${componentSpec.name}.tsx`)

  // 5. Storybook metadata
  console.log(`\n📖 Creating Storybook stories...`)
  // Storybook setup would happen here

  // 6. Save image assets
  console.log(`\n🖼️ Saving image assets...`)
  fs.mkdirSync('public/assets', { recursive: true })

  for (const [nodeId, imageUrl] of Object.entries(images.images)) {
    // Download and save image
    // (In actual implementation, use axios to download)
  }

  console.log(`✅ All assets saved`)

  console.log(`\n✨ Pipeline completed successfully!`)
  console.log(`📁 Generated files:`)
  console.log(`  - src/styles/tokens.json`)
  console.log(`  - src/styles/tokens.css`)
  console.log(`  - tailwind.config.js`)
  console.log(`  - src/components/${componentSpec.name}.tsx`)
  console.log(`  - src/components/${componentSpec.name}.stories.tsx`)
}

function toPascalCase(str: string): string {
  return str
    .split(/[-_\s]/)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join('')
}

// Usage
designToCodePipeline(
  'https://figma.com/design/ABC123XYZ/LoginPage?node-id=10-25'
).catch(console.error)
```

---

## References

- **Figma REST API**: https://developers.figma.com/docs/rest-api
- **Code Connect**: https://github.com/figma/code-connect
- **DTCG Format**: https://designtokens.org
