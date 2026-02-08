# Pencil MCP Rendering Guide

Pencil MCP integration for rendering DNA codes into visual .pen frames and creating design proposals.

## Overview

Pencil MCP enables text-to-design conversion through DNA codes, a declarative format for describing UI designs that can be version controlled and rendered into visual .pen frames.

## DNA Code Format

### Basic Structure

```dna
// Button component DNA code
component Button {
  variant: primary
  size: medium
  content: "Click me"
  onClick: handleSubmit
}
```

### Layout Structure

```dna
// Form layout with multiple components
layout LoginForm {
  direction: column
  spacing: 16
  children: [
    Input {
      placeholder: "Email"
      type: email
    }
    Input {
      placeholder: "Password"
      type: password
    }
    Button {
      variant: primary
      content: "Sign In"
    }
  ]
}
```

## MCP Configuration

### Server Setup

Add to `.mcp.json`:

```json
{
  "mcpServers": {
    "pencil": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-pencil"],
      "env": {
        "PENCIL_API_KEY": "your-key-here"
      }
    }
  }
}
```

### Authentication

1. Obtain Pencil API key:
   - Sign up at https://pencil.dev
   - Generate API key in settings
   - Add to MCP configuration

2. Environment Variables:
   - Set `PENCIL_API_KEY` in environment
   - Required for rendering operations

## Text-to-Design Generation

### Natural Language Prompts

```
// Simple component prompt
"Create a primary button with medium size"

// Complex layout prompt
"Design a login form with email input, password input,
and submit button arranged vertically with 16px spacing"

// Styled component prompt
"Create a card component with white background, rounded
corners, subtle shadow, and padding of 24px"
```

### DNA Code Generation

```typescript
// Generate DNA from natural language
const dna = await pencil.generate_dna({
  prompt: "Create a navigation bar with logo and links",
  style: "minimalist"
});

// Returns DNA code structure
```

## .pen Frame Rendering

### Render DNA to Frame

```typescript
// Render DNA code to visual frame
const frame = await pencil.render_dna({
  dna: dnaCode,
  format: "pen",
  options: {
    width: 1200,
    height: 800,
    theme: "light"
  }
});

// Returns .pen file URL or base64 data
```

### Frame Configuration

```typescript
// Configure rendering options
const options = {
  width: 1200,           // Frame width in pixels
  height: 800,           // Frame height in pixels
  theme: "light",        // light, dark, or auto
  scale: 1,              // Resolution scale
  interactive: true,     // Enable interactive elements
  annotations: true      // Show design annotations
};
```

## Design Iteration Workflow

### Iterative Refinement

1. **Initial Design**
   ```dna
   component Button {
     content: "Click me"
   }
   ```

2. **Add Styling**
   ```dna
   component Button {
     content: "Click me"
     background: primary.500
     color: white.100
     padding: md
     radius: md
   }
   ```

3. **Add States**
   ```dna
   component Button {
     content: "Click me"
     states: [hover, active, disabled]
     hover: {
       background: primary.600
     }
     active: {
       transform: scale(0.98)
     }
   }
   ```

4. **Final Polish**
   ```dna
   component Button {
     content: "Click me"
     variant: primary
     size: medium
     states: [hover, active, disabled]
     animation: {
       duration: 200ms
       easing: ease-in-out
     }
   }
   ```

### Version Control Patterns

```bash
# Commit DNA code changes
git add designs/button.dna
git commit -m "design: add hover states to button"

# Track design iterations
git log --follow designs/button.dna

# Compare design versions
git diff HEAD~1 designs/button.dna
```

## Collaborative Workflows

### Design Review Process

1. **Create Design Proposal**
   - Generate DNA code from requirements
   - Render to .pen frame
   - Share with team for review

2. **Collect Feedback**
   - Add comments to .pen frame
   - Document requested changes
   - Create feedback issue

3. **Iterate on Design**
   - Update DNA code based on feedback
   - Re-render to .pen frame
   - Share updated version

4. **Final Approval**
   - Mark design as approved
   - Export to code for implementation
   - Archive design iterations

### Team Collaboration

```typescript
// Share design for review
const shareUrl = await pencil.share_design({
  frame: penFrameUrl,
  permissions: "comment",
  expiresAt: "2026-02-16"
});

// Collect feedback
const feedback = await pencil.get_feedback({
  frame: penFrameUrl
});

// Apply feedback to DNA
const updatedDna = apply_feedback(dnaCode, feedback);
```

## DNA Code Reference

### Component Syntax

```dna
// Component definition
component ComponentName {
  // Properties
  property: value

  // Children
  children: [
    ChildComponent {}
  ]

  // States
  states: [state1, state2]

  // Styling
  style: {
    property: value
  }
}
```

### Design Tokens

```dna
// Token references
color: primary.500
spacing: md
radius: lg

// Token definitions
tokens {
  primary.500 = #3B82F6
  md = 16px
  lg = 8px
}
```

### Layout Syntax

```dna
// Layout containers
layout Container {
  direction: row | column
  justify: start | center | end | space-between
  align: start | center | end | stretch
  gap: 16
  wrap: true | false
}
```

## Best Practices

### DNA Code Organization

```
designs/
  tokens/          # Design token definitions
  components/      # Reusable component DNA
  layouts/         # Layout compositions
  pages/           # Full page designs
```

### Naming Conventions

```dna
// Use clear, descriptive names
component PrimaryButton {}  // Good
component Btn {}            // Bad

// Use kebab-case for files
primary-button.dna          // Good
primaryButton.dna           // Bad
```

### Documentation

```dna
/**
 * Primary Button Component
 *
 * Usage: Main call-to-action buttons
 * Variants: primary, secondary, tertiary
 * Sizes: small, medium, large
 * States: hover, active, disabled
 */
component Button {}
```

## Advanced Patterns

### Responsive Design

```dna
component Responsive {
  breakpoints: {
    mobile: 640px
    tablet: 768px
    desktop: 1024px
  }
  mobile: {
    direction: column
  }
  desktop: {
    direction: row
  }
}
```

### Theme Support

```dna
component Themed {
  theme: {
    light: {
      background: white.100
      color: gray.900
    }
    dark: {
      background: gray.900
      color: white.100
    }
  }
}
```

### Animation

```dna
component Animated {
  animation: {
    property: transform
    duration: 300ms
    easing: ease-in-out
    delay: 0ms
  }
  states: {
    hover: {
      transform: scale(1.05)
    }
  }
}
```

## Error Handling

### Invalid DNA Syntax
```
Error: Failed to parse DNA code (line 15)
Solution: Validate DNA syntax, check for missing brackets or commas
```

### Render Failure
```
Error: Failed to render .pen frame
Solution: Check DNA code for unsupported properties or invalid tokens
```

### Token Resolution
```
Error: Token 'primary.500' not found
Solution: Define missing token in tokens section or reference existing token
```

## Performance Optimization

### Caching Rendered Frames

```typescript
// Cache rendered .pen frames
const cacheKey = `pencil:${dnaHash}`;
const cached = await cache.get(cacheKey);

if (cached) {
  return cached;
}

const frame = await pencil.render_dna({ dna });
await cache.set(cacheKey, frame, 3600);
```

### Batch Rendering

```typescript
// Render multiple frames in parallel
const frames = await Promise.all([
  pencil.render_dna({ dna: buttonDna }),
  pencil.render_dna({ dna: inputDna }),
  pencil.render_dna({ dna: cardDna })
]);
```

## Resources

- Pencil Documentation: https://docs.pencil.dev
- DNA Code Reference: https://docs.pencil.dev/dna
- Pencil MCP Server: https://github.com/modelcontextprotocol/servers/tree/main/src/pencil
- Design Examples: https://pencil.dev/examples

---

Last Updated: 2026-02-09
Tool Version: Pencil MCP 1.0.0
