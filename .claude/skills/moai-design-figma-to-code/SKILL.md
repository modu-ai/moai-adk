---
name: moai-design-figma-to-code
description: Converting Figma designs to React/HTML/CSS code. Extract design tokens, generate components, and automate code generation. Use when converting Figma designs to code, extracting design systems, or building component libraries from design files.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: design
created: 2025-10-31
---

# Design: Figma to Code Conversion

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | design |
| Created | 2025-10-31 |
| Allowed tools | Read, Bash, WebFetch |
| Auto-load | When converting Figma designs to code |
| Trigger cues | "Convert Figma to code", "Extract design tokens", "Generate components from Figma" |

## What it does

Automates the conversion of Figma design files into production-ready React/HTML/CSS code by extracting design tokens (colors, spacing, typography), analyzing component structures, and generating clean, maintainable code with proper naming conventions and accessibility standards.

## When to use

- Converting Figma mockups to React components
- Extracting design system tokens from Figma files
- Building component libraries from design specifications
- Automating design-to-code workflows
- Maintaining design-code consistency

## How it works

### 1. Design Token Extraction

**Use Figma Variables as Native Token System (2025)**:

Figma Variables represent the native way to define tokens:
- **Color variables**: \`color.primary.base\`, \`color.bg.card\`
- **Number variables**: \`spacing.md = 16\`, \`radius.lg = 8\`
- **Boolean variables**: \`isDarkMode = true\`
- **String variables**: \`font.family.primary = Inter\`

**Token Organization**:
\`\`\`typescript
// Generated tokens from Figma Variables
export const tokens = {
  colors: {
    primary: {
      base: '#4F46E5',
      hover: '#4338CA',
      active: '#3730A3'
    },
    neutral: {
      50: '#F9FAFB',
      100: '#F3F4F6',
      // ...
    }
  },
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px'
  },
  typography: {
    fontFamily: {
      primary: 'Inter, system-ui, sans-serif',
      mono: 'JetBrains Mono, monospace'
    },
    fontSize: {
      xs: '12px',
      sm: '14px',
      base: '16px',
      lg: '18px',
      xl: '20px'
    }
  }
};
\`\`\`

### 2. Figma API Access (Enterprise Plan Required)

**Authentication**:
\`\`\`typescript
// Personal Access Token (read-only for variables)
const FIGMA_TOKEN = process.env.FIGMA_TOKEN;

const headers = {
  'X-Figma-Token': FIGMA_TOKEN
};
\`\`\`

**Fetching Design Data**:
\`\`\`typescript
// Get file data
const fileKey = 'abc123def456';
const response = await fetch(
  \`https://api.figma.com/v1/files/\${fileKey}\`,
  { headers }
);

// Extract components
const file = await response.json();
const components = file.document.children
  .filter(node => node.type === 'COMPONENT');
\`\`\`

### 3. Component Generation

**Pattern: Frame → React Component**:

Figma Frame Structure:
\`\`\`
Frame: "Button/Primary"
├─ Text: "Label"
├─ Rectangle: "Background"
└─ Vector: "Icon"
\`\`\`

Generated React Component:
\`\`\`tsx
interface ButtonProps {
  label: string;
  variant?: 'primary' | 'secondary';
  icon?: React.ReactNode;
  onClick?: () => void;
}

export const Button: React.FC<ButtonProps> = ({
  label,
  variant = 'primary',
  icon,
  onClick
}) => {
  return (
    <button
      className={\`btn btn-\${variant}\`}
      onClick={onClick}
      aria-label={label}
    >
      {icon && <span className="btn-icon">{icon}</span>}
      <span className="btn-label">{label}</span>
    </button>
  );
};
\`\`\`

### 4. CSS Generation from Styles

**Auto-Layout → Flexbox**:
\`\`\`css
/* Figma Auto-Layout Properties */
.container {
  display: flex;
  flex-direction: column; /* or row */
  gap: 16px; /* spacing between */
  padding: 24px; /* padding inside */
  align-items: center; /* alignment */
  justify-content: space-between; /* distribution */
}
\`\`\`

**Typography Styles**:
\`\`\`css
/* Figma Text Style → CSS */
.heading-1 {
  font-family: var(--font-primary);
  font-size: var(--text-2xl);
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.02em;
}
\`\`\`

### 5. Multi-Theme Support (Modes)

**Using Figma Modes for Theming**:
\`\`\`typescript
// Extract theme modes
const modes = {
  light: {
    bg: '#FFFFFF',
    text: '#1F2937',
    primary: '#4F46E5'
  },
  dark: {
    bg: '#1F2937',
    text: '#F9FAFB',
    primary: '#818CF8'
  }
};

// CSS Variables per theme
[data-theme="light"] {
  --bg: #FFFFFF;
  --text: #1F2937;
  --primary: #4F46E5;
}

[data-theme="dark"] {
  --bg: #1F2937;
  --text: #F9FAFB;
  --primary: #818CF8;
}
\`\`\`

## Best Practices

### Naming Conventions
- Use semantic token names: \`color.bg.card\` not \`gray.100\`
- Component names match Figma frame names: \`Button/Primary\` → \`ButtonPrimary\`
- Follow BEM or component-scoped naming for CSS

### Separation of Concerns
- Separate tokens into dedicated file and publish as Figma Library
- Keep design tokens in separate \`tokens.ts\` file
- Store components in \`components/\` directory
- Place styles in \`styles/\` directory

### Version Control
- Document Figma file version in component comments
- Track token changes with migration notes
- Use semantic versioning for design system updates

### Accessibility
- Extract alt text from Figma layer names
- Generate ARIA labels from component structure
- Ensure color contrast meets WCAG standards

### Automation
- Use Figma webhooks to trigger code regeneration
- Implement CI/CD pipeline for design token updates
- Validate generated code with linters and type checkers

## References
- Figma REST API Documentation. "Variables API (Enterprise)." https://www.figma.com/developers/api (accessed 2025-10-31).
- Figma Design. "Design Tokens in Figma." https://www.figma.com/community/plugin/888356646278934516 (accessed 2025-10-31).
- Diez Framework. "Figma Design Token Extraction." https://diez.org/getting-started/figma.html (accessed 2025-10-31).
- Designilo. "Token-Based Component Libraries." https://designilo.com/2025/07/10/token-based-component-library/ (accessed 2025-10-31).

## Changelog
- 2025-10-31: v1.0.0 - Initial release with Figma Variables support, multi-theme modes, React/HTML generation

## Works well with
- \`moai-design-figma-mcp\` (Figma MCP integration)
- \`moai-lang-typescript\` (TypeScript code generation)
- \`moai-foundation-trust\` (Code quality validation)
