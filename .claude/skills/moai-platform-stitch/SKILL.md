---
name: moai-platform-stitch
description: >
  Google Stitch MCP integration for AI-powered UI/UX design generation. Use when generating
  UI designs from text, extracting design context from screens, exporting screen code and
  images, managing Stitch projects and screens, or implementing autonomous build loops
  with the baton system for continuous site development.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash Grep Glob mcp__stitch__extract_design_context mcp__stitch__fetch_screen_code mcp__stitch__fetch_screen_image mcp__stitch__generate_screen_from_text mcp__stitch__create_project mcp__stitch__list_projects mcp__stitch__list_screens mcp__stitch__get_project mcp__stitch__get_screen
user-invocable: false
metadata:
  version: "2.0.0"
  category: "platform"
  modularized: "false"
  status: "active"
  updated: "2026-01-29"
  tags: "stitch, google, ui, ux, design, code-generation, ai-design, build-loop, autonomous-frontend, ci-cd"
  context7-libraries: "/stitch/stitch-mcp"
  related-skills: "moai-domain-uiux, moai-domain-frontend, moai-foundation-claude"

# MoAI Extension: Triggers
triggers:
  keywords:
    - UI design
    - screen generation
    - design system
    - Stitch
    - UI/UX
    - design extraction
    - code generation from design
    - visual prototyping
    - design consistency
    - design DNA
    - Google Stitch
    - AI design generation
    - design to code
    - design tokens extraction
    - build loop
    - autonomous frontend
    - baton system
    - continuous site generation
    - iterative design
    - automated web development
---

# Google Stitch MCP Integration

Comprehensive Google Stitch MCP integration for AI-powered UI/UX design generation, design context extraction, screen code export, and design system management with advanced prompt engineering capabilities.

---

## Quick Reference

Stitch MCP Core Features:

- extract_design_context: Extract "Design DNA" (fonts, colors, layouts, components) from existing screens
- fetch_screen_code: Download production-ready HTML/CSS/JavaScript code from generated screens
- fetch_screen_image: Download high-resolution screenshots of designs
- generate_screen_from_text: Generate new screens from text descriptions using AI
- create_project: Create new Stitch workspace/project folders
- list_projects: List all available Stitch projects
- list_screens: List all screens within a project
- get_project: Retrieve project details and metadata
- get_screen: Get screen metadata and information

Quick Decision Tree:

- Need UI design from description? Use generate_screen_from_text
- Have existing design to extract from? Use extract_design_context
- Need production code from design? Use fetch_screen_code
- Need screenshot of design? Use fetch_screen_image
- Managing multiple designs? Use create_project, list_projects, list_screens
- Autonomous site development? Use Build Loop pattern with baton system

Common UI/UX Keywords:

Layout: single-column, grid, sidebar, centered, stacked, scrollable, flex, card, masonry, dashboard

Components: button, input, dropdown, checkbox, radio, toggle, slider, modal, dialog, drawer, tooltip, popover, tabs, breadcrumb, pagination, navigation, footer, header, card, badge, avatar, progress, spinner, alert, toast, notification, accordion, carousel

Style: minimal, modern, flat, material, skeuomorphic, brutalist, glassmorphism, neomorphism, dark-mode, light-mode, responsive, mobile-first, desktop-first

Interactions: hover, focus, active, disabled, loading, success, error, warning, info, tooltip, transition, animation, slide, fade, scale, rotate, bounce, shake, pulse

Colors: primary, secondary, accent, neutral, success, warning, error, info, light, dark, background, foreground, border, shadow

Typography: heading, title, subtitle, body, caption, label, button, link, code, monospace, sans-serif, serif

Spacing: tight, normal, relaxed, compact, spacious, consistent, uniform

Alignment: left, center, right, justify, start, end, stretch, baseline

---

## Installation Guide

### Prerequisites

Google Cloud Setup:

Step 1: Create a Google Cloud project at console.cloud.google.com

Step 2: Enable the Stitch API by running:
```bash
gcloud beta services mcp enable stitch.googleapis.com
```

Step 3: Authenticate with Google Cloud:
```bash
gcloud auth application-default login
```

Step 4: Set environment variable in your shell profile (.bashrc, .zshrc):
```bash
export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"
```

### MCP Configuration

Add to .claude/.mcp.json:
```json
{
  "mcpServers": {
    "stitch": {
      "command": "npx",
      "args": ["-y", "stitch-mcp"],
      "env": {
        "GOOGLE_CLOUD_PROJECT": "YOUR_PROJECT_ID"
      }
    }
  }
}
```

### Verification

Verify Stitch MCP is working:
```bash
# List available projects
npx -y stitch-mcp list_projects
```

---

## Implementation Guide

### The "Designer Flow" Pattern

This is the recommended workflow for consistent UI design generation:

Phase 1: Extract Design Context

If you have existing screens, extract their design context first to maintain consistency.

Call extract_design_context with an existing screen_id to retrieve:
- Color palette (primary, secondary, accent, neutral colors)
- Typography (font families, sizes, weights, line heights)
- Spacing (margin, padding, gap scales)
- Component styles (buttons, inputs, cards, modals)
- Layout patterns (grid systems, breakpoints)

Phase 2: Generate New Screens

Use generate_screen_from_text with the extracted design context to create new screens that maintain visual coherence.

Provide a detailed prompt describing:
- Screen purpose and user goals
- Components needed (header, navigation, cards, forms)
- Layout type (single column, grid, sidebar)
- Content hierarchy (headings, body, calls-to-action)
- Interactive elements (hover states, focus states)

Phase 3: Export Deliverables

Export both code and images for complete design handoff:
- fetch_screen_code for HTML/CSS/JavaScript implementation
- fetch_screen_image for high-resolution PNG screenshots

### Prompt Engineering Guide

Official Google Stitch Prompt Guidelines:

Golden Rule: One screen/component at a time, one or two adjustments per prompt

5-Part Prompt Structure:

1. Context: What is this screen for? Who is it for?
2. Design: Overall visual style and aesthetic
3. Components: Complete list of UI elements needed
4. Layout: How components are arranged and spaced
5. Style: Specific colors, fonts, and visual attributes

Prompt Quality Checklist:

Before calling generate_screen_from_text, verify:

- Component List: All UI components explicitly listed
- Layout Type: Grid, flex, single column, sidebar specified
- Content Hierarchy: Headings, body, CTAs in order of importance
- Interactions: Hover, focus, active states described
- Responsive: Mobile vs desktop behavior specified
- Accessibility: ARIA labels, keyboard navigation included
- Design Context: Extracted tokens passed (if available)
- Style Keywords: Colors, fonts, visual attributes specified

Effective Prompt Pattern:

"Create a [screen type] with [component list]. Use [layout type] with [content hierarchy]. Include [interactive elements] with [responsive behavior]. Apply [design context/style keywords]."

Example:
"Create a login screen with email input, password input with show/hide toggle, remember me checkbox, login button, forgot password link, and social login options. Use centered card layout with single column. Include page title 'Welcome Back' and subtitle. Primary button with hover state, input fields with focus ring. Mobile: stacked vertically. Desktop: 400px max-width card. Use blue primary color, white background, Inter font family."

### Design Context Extraction

When to Extract Design Context:

- Existing design system needs to be maintained
- Multiple screens require visual consistency
- Design tokens need to be documented
- Migrating from Figma or other design tools

Design Context Structure:

The extracted design context includes:

Colors:
- Primary palette (main brand colors)
- Secondary palette (supporting colors)
- Accent colors (highlights, CTAs)
- Neutral colors (grays for text, borders)
- Semantic colors (success, warning, error, info)

Typography:
- Font families (headings, body, code)
- Font sizes (display, h1-h6, body, small, caption)
- Font weights (light, regular, medium, semibold, bold)
- Line heights (tight, normal, relaxed)

Spacing:
- Base spacing unit (4px, 8px, etc.)
- Spacing scale (xs, sm, md, lg, xl)
- Component spacing (padding, margins)

Components:
- Button styles (primary, secondary, ghost, danger)
- Input styles (default, error, disabled)
- Card styles (elevation, border, padding)
- Modal styles (overlay, content, animation)

Layout:
- Grid system (columns, gaps)
- Breakpoints (mobile, tablet, desktop)
- Container max-widths

### Code Export

Exported Code Includes:

HTML Structure:
- Semantic HTML5 elements (header, main, section, nav, footer)
- Accessibility attributes (aria-label, role, tabindex)
- Meta tags for responsive design

CSS Styling:
- CSS variables for design tokens
- Responsive media queries
- Hover and focus states
- Dark mode support (if applicable)

JavaScript Functionality:
- Form validation
- Show/hide password toggle
- Mobile menu toggle
- Smooth scrolling

Code Quality Features:

- Semantic HTML for accessibility and SEO
- CSS variables for easy theming
- Responsive breakpoints
- ARIA attributes for screen readers
- Cross-browser compatible

### Project and Screen Management

Project Organization:

Use create_project to organize screens by feature or product:
```bash
create_project("E-commerce App")
```

Project Naming Best Practices:

- Use descriptive names: "E-commerce Checkout Flow" not "Project 1"
- Group related screens in same project
- Separate staging and production projects

Screen Listing:

Use list_screens to view all screens in a project:
- View screen names and IDs
- Check screen status (generated, exported)
- Identify screens needing updates

Screen Metadata:

Use get_screen to retrieve:
- Screen creation date
- Last modification date
- Generation parameters (prompt used)
- Export status

---

## Anti-Patterns to Avoid

Common Prompt Mistakes:

Over-Specification:

Don't specify pixel-perfect dimensions like "375px wide" or "button 48px height". Instead use relative terms like "mobile width" or "large touch targets". Stitch handles responsive scaling automatically.

Under-Specification:

Don't use vague prompts like "create a nice login page". Always include component list, layout type, and content hierarchy. Vague prompts produce unpredictable results.

Missing Context:

Don't ignore design_context parameter when existing screens exist. Passing extracted design tokens ensures consistency across all screens. Always extract context from at least one screen before generating new ones in a project.

Mixing Concerns:

Don't combine layout changes with component additions in one prompt. For example, don't say "add a sidebar and also make the header sticky". Generate base layout first, then refine with follow-up prompts.

Long Prompts:

Keep prompts under 500 characters for best results. Long prompts confuse the AI model. Focus on essential elements first, then refine incrementally.

Ignoring Accessibility:

Don't skip accessibility requirements. Always specify ARIA labels for icon buttons, keyboard navigation patterns, and focus indicators. Accessible design is not optional.

Assuming Mobile-First:

Don't assume Stitch will automatically optimize for mobile. Always specify responsive behavior: "Mobile: stack elements vertically. Desktop: horizontal layout with sidebar."

Forgetting Brand Guidelines:

Don't ignore brand colors and fonts. Extract design context from brand assets first, then reference those tokens in prompts: "Use primary brand color #3B82F6, Inter font family."

---

## Advanced Patterns

### The Build Loop Pattern (Autonomous Frontend Development)

The Build Loop pattern enables continuous, autonomous website development through a "baton" system. Each iteration automatically:
1. Reads the current task from a baton file (next-prompt.md)
2. Generates a page using Stitch MCP tools
3. Integrates the page into the site structure
4. Writes the next task to the baton file for the next iteration

Use Case: Continuous website development without manual intervention between iterations.

Required Files:
- DESIGN.md: Visual design system specification
- SITE.md: Site vision, sitemap, and roadmap
- next-prompt.md: Baton file containing next task
- stitch.json: Stitch project identifier

#### Baton System

The next-prompt.md file acts as a relay baton between iterations:

```yaml
---
page: about
---
A page describing how the tracking system works.

**DESIGN SYSTEM (REQUIRED):**
[Copy from DESIGN.md Section 6]

**Page Structure:**
1. Header with navigation
2. Explanation of tracking methodology
3. Footer with links
```

#### Execution Protocol

Step 1: Read the Baton
Parse next-prompt.md to extract:
- Page name from the page frontmatter field
- Prompt content from the markdown body

Step 2: Consult Context Files
Read SITE.md and DESIGN.md before generating to maintain consistency.

Step 3: Generate with Stitch
1. Discover namespace by running list_tools
2. Get or create project (check for stitch.json)
3. Generate screen with generate_screen_from_text
4. Retrieve assets using get_screen

Step 4: Integrate into Site
1. Move generated HTML to site/public/{page}.html
2. Fix asset paths to be relative to public folder
3. Update navigation and wire real links
4. Ensure consistent headers/footers

Step 5: Visual Verification (Optional)
If Chrome DevTools MCP is available:
1. Start dev server with npx serve site/public
2. Navigate to page
3. Capture screenshot
4. Compare against Stitch screenshot
5. Stop server

Step 6: Update Site Documentation
Modify SITE.md:
- Add new page to Section 4 (Sitemap)
- Remove idea from Section 6 (Creative Freedom)
- Update Section 5 (Roadmap) if completed

Step 7: Prepare the Next Baton

Critical: You MUST update next-prompt.md before completing.

Decide the next page and write with proper YAML frontmatter including design system block.

#### File Structure Reference

```
project/
├── next-prompt.md      # The baton
├── stitch.json         # Stitch project ID
├── DESIGN.md           # Visual design system
├── SITE.md             # Site vision, sitemap, roadmap
├── queue/              # Staging area
│   ├── {page}.html
│   └── {page}.png
└── site/public/        # Production pages
    ├── index.html
    └── {page}.html
```

#### Orchestration Options

CI/CD: GitHub Actions triggers on next-prompt.md changes
Human-in-loop: Developer reviews before continuing
Agent chains: One agent dispatches to another
Manual: Developer runs repeatedly

#### Common Pitfalls

- Forgetting to update next-prompt.md (breaks the loop)
- Recreating existing pages
- Not including the design system block
- Leaving placeholder links
- Forgetting to persist stitch.json

### Design System Migration

Migrating from Figma to Stitch:

Step 1: Create representative screens in Stitch from Figma designs
Step 2: Extract design context from multiple screens
Step 3: Consolidate design tokens into unified system
Step 4: Use extracted context for all new screen generation

### Consistency Verification

Verifying Design Consistency:

Step 1: Extract design context from multiple screens
Step 2: Compare color palettes and typography scales
Step 3: Identify inconsistencies (spacing variations, conflicting colors)
Step 4: Generate new screens with corrected context

### Rapid Prototyping Workflow

For MVP and rapid prototyping:

Step 1: Generate basic screens with minimal context
Step 2: Extract unified design context from generated screens
Step 3: Regenerate screens with consistent context
Step 4: Export code for immediate implementation

### Incremental Refinement Strategy

Start Simple, Complexify Gradually:

Iteration 1: Generate basic layout with core components
Iteration 2: Add interactive elements (hover, focus, active states)
Iteration 3: Refine spacing and alignment
Iteration 4: Add polish (animations, transitions, micro-interactions)

This incremental approach prevents prompt complexity from overwhelming the AI model.

---

## Error Handling

### Common Issues

Authentication Errors:

If you receive authentication errors:
1. Verify gcloud auth application-default login was successful
2. Check GOOGLE_CLOUD_PROJECT environment variable is set
3. Ensure Stitch API is enabled in your Google Cloud project

Resolution:
```bash
# Re-authenticate
gcloud auth application-default login

# Enable API
gcloud beta services mcp enable stitch.googleapis.com

# Verify project ID
echo $GOOGLE_CLOUD_PROJECT
```

API Not Enabled Errors:

If API is not enabled:
```bash
gcloud beta services mcp enable stitch.googleapis.com
```

Quota Errors:

If you exceed quota:
- Wait for quota reset (typically daily)
- Consider upgrading to paid Google Cloud tier
- Optimize prompt specificity to reduce regeneration

### Generation Failures

Prompt Too Vague:

If generation fails or produces poor results:
- Add specific component list to prompt
- Specify layout type explicitly
- Include content hierarchy details
- Mention interactive elements and behaviors

Prompt Too Complex:

If generation times out or fails:
- Simplify prompt to core components
- Generate screen in multiple iterations
- Start with basic layout, then add components

---

## Best Practices

### Design First Principles

Consistency Over Speed:

Always extract design context from existing screens before generating new ones. Consistency creates professional, polished designs. Skipping context creates disjointed user experience.

Accessibility First:

Always include accessibility requirements in prompts. Specify ARIA labels needed, keyboard navigation requirements, and color contrast needs. Accessible design serves all users and ensures legal compliance.

Responsive Default:

Always specify responsive behavior in prompts. Describe mobile layout, tablet behavior, and desktop presentation. Responsive design ensures usability across all devices.

Semantic Structure:

Always request semantic HTML in prompts. Specify header, main, section, nav, footer elements. Semantic HTML enables accessibility, SEO, and maintainability.

### Prompt Engineering Best Practices

Effective Prompt Structure:

Component List: List all UI components needed (buttons, inputs, cards)
Layout Type: Specify layout (single column, grid, sidebar navigation)
Content Hierarchy: Describe headings, body text, CTAs in order of importance
Interactions: Describe hover states, focus states, animations
Responsive: Describe mobile vs desktop behavior
Context: Pass design context from existing screens for consistency

Incremental Approach:

Generate screens in iterations. Start with basic layout, then add complexity. This approach produces better results than trying to specify everything in one prompt.

Design Token Integration:

Always pass extracted design_context when available. This ensures color, typography, and spacing consistency across all screens in a project.

---

## Resources

### Official Resources

- Google Stitch: https://cloud.google.com/stitch
- Stitch MCP GitHub: https://github.com/Kargatharaakash/stitch-mcp
- Google Cloud Console: https://console.cloud.google.com

### Prompt Template Library

For comprehensive prompt templates organized by category (authentication, data screens, navigation, forms, modals, accessibility patterns), see reference.md which includes 20+ ready-to-use prompt examples with progressive complexity levels.

### Works Well With

- moai-domain-uiux: Design system architecture and component design
- moai-domain-frontend: Frontend implementation with exported code
- moai-foundation-claude: Claude Code integration and agent patterns
- moai-library-shadcn: Component library for React implementations

### Related Skills

- moai-domain-uiux: For design system architecture beyond Stitch
- moai-domain-frontend: For component implementation after design export

---

## Examples

### Example 1: E-Commerce Product Screen

Generate product listing screen with consistent design context:

Step 1: Extract context from existing home screen:
```
extract_design_context(screen_id="home-screen-001")
```

Step 2: Generate product screen:
```
generate_screen_from_text(
  prompt="E-commerce product listing screen with 12 product cards in 3-column grid layout, filter sidebar on left with category links, sort dropdown at top right. Each product card has image, title in bold, price in accent color, add to cart button as primary CTA. Card hover effect with shadow increase. Mobile: single column, no sidebar. Desktop: sidebar + grid.",
  design_context={colors, typography, spacing from step 1},
  project_id="ecommerce-app"
)
```

Step 3: Export code and image:
```
fetch_screen_code(screen_id="product-listing-001")
fetch_screen_image(screen_id="product-listing-001")
```

### Example 2: Dashboard Analytics Screen

Generate analytics dashboard with charts:

```
generate_screen_from_text(
  prompt="Analytics dashboard with header containing page title 'Dashboard' and date range picker dropdown. Main content has three metric cards (Total Revenue, Active Users, Conversion Rate) at top in row. Line chart showing revenue trend over time below metrics. Data table with recent transactions at bottom. Use sidebar navigation layout with logo, nav links, user profile. Mobile: hide sidebar, stack metric cards vertically.",
  design_context={existing dashboard context},
  project_id="analytics-app"
)
```

### Example 3: Mobile App Authentication

Generate mobile-first authentication screens:

```
generate_screen_from_text(
  prompt="Mobile login screen with full-width primary button, large touch targets (44px min height), email input with email icon prefix, password input with show/hide eye icon, social login buttons (Google, Apple) with brand colors. Stack all elements vertically with 16px spacing between. Center on screen. Use primary brand color for main button, gray for social buttons.",
  design_context={mobile design tokens},
  project_id="mobile-app"
)
```

### Example 4: Multi-Step Form Wizard

Generate complex form with progressive disclosure:

```
generate_screen_from_text(
  prompt="Multi-step registration wizard with progress indicator at top showing 3 steps (Account, Profile, Complete). Step 1: Account form with email, password, confirm password inputs. Step 2: Profile form with name, phone, bio textarea. Step 3: Success message with continue button. Use Next/Back navigation buttons at bottom right. Center content in 600px card. Highlight current step in progress indicator.",
  design_context={form design tokens},
  project_id="saas-app"
)
```

### Example 5: Data Table with Actions

Generate complex data table with inline actions:

```
generate_screen_from_text(
  prompt="Data table with columns: Name (avatar + text), Email, Role (badge), Status (colored dot), Actions (kebab menu). Include sort indicators on column headers. Row hover effect with background color change. Kebab menu opens dropdown with Edit, Delete options. Pagination controls at bottom right. Sticky header when scrolling. Use compact spacing (32px row height).",
  design_context={data table tokens},
  project_id="admin-panel"
)
```

---

Status: Production Ready
Version: 2.0.0
Last Updated: 2026-01-29
Platform Coverage: Google Stitch MCP Only
