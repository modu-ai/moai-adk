# Batch Design Patterns

Common operation sequences for Pencil MCP `batch_design` tool.

## Operation Syntax Quick Reference

```javascript
// Insert node with binding
binding=I("parentId", {type: "frame", ...props})

// Copy node (creates ref to reusable, or duplicate)
copy=C("sourceId", "parentId", {positionDirection: "right"})

// Update existing node properties
U("nodeId", {fill: "#000"})

// Update nested node in component instance
U("instanceId/childId", {content: "New text"})

// Replace node entirely
newNode=R("path/to/node", {type: "text", content: "Replaced"})

// Move node to new parent
M("nodeId", "newParentId", 0)

// Delete node
D("nodeId")

// Generate/get image for frame
G("frameId", "ai", "prompt describing image")
G("frameId", "stock", "search keywords")
```

## Hero Section Pattern

```javascript
// Hero container
hero=I(document, {
  type: "frame",
  name: "Hero",
  width: 1440,
  height: 800,
  layout: "vertical",
  alignItems: "center",
  padding: [120, 80, 120, 80],
  gap: 32,
  fill: "#0A0A0A"
})

// Headline
headline=I(hero, {
  type: "text",
  content: "Hero Headline",
  fontSize: 64,
  fontWeight: "700",
  textColor: "#FAFAFA",
  textAlign: "center"
})

// Subheadline
subhead=I(hero, {
  type: "text",
  content: "Supporting description text",
  fontSize: 20,
  fontWeight: "400",
  textColor: "#A0A0A0",
  textAlign: "center",
  width: 600
})

// CTA container
ctas=I(hero, {
  type: "frame",
  layout: "horizontal",
  gap: 16
})

// Primary button
primary=I(ctas, {
  type: "frame",
  layout: "horizontal",
  alignItems: "center",
  padding: [16, 32, 16, 32],
  cornerRadius: [8, 8, 8, 8],
  fill: "#FFFFFF"
})
primaryText=I(primary, {
  type: "text",
  content: "Get Started",
  fontSize: 16,
  fontWeight: "600",
  textColor: "#0A0A0A"
})

// Secondary button
secondary=I(ctas, {
  type: "frame",
  layout: "horizontal",
  alignItems: "center",
  padding: [16, 32, 16, 32],
  cornerRadius: [8, 8, 8, 8],
  fill: "transparent",
  stroke: "#FFFFFF",
  strokeThickness: 1
})
secondaryText=I(secondary, {
  type: "text",
  content: "Learn More",
  fontSize: 16,
  fontWeight: "500",
  textColor: "#FFFFFF"
})
```

## Card Pattern

```javascript
// Card container
card=I(parent, {
  type: "frame",
  name: "Card",
  width: 360,
  height: 420,
  layout: "vertical",
  gap: 0,
  cornerRadius: [12, 12, 12, 12],
  fill: "#FFFFFF",
  stroke: "#E5E5E5",
  strokeThickness: 1
})

// Image area
imageArea=I(card, {
  type: "frame",
  width: "fill_container",
  height: 200,
  fill: "#F5F5F5"
})
G(imageArea, "stock", "abstract gradient background")

// Content area
content=I(card, {
  type: "frame",
  layout: "vertical",
  padding: [24, 24, 24, 24],
  gap: 12,
  width: "fill_container"
})

// Title
title=I(content, {
  type: "text",
  content: "Card Title",
  fontSize: 20,
  fontWeight: "600",
  textColor: "#1A1A1A"
})

// Description
desc=I(content, {
  type: "text",
  content: "Card description text goes here with more details.",
  fontSize: 14,
  fontWeight: "400",
  textColor: "#6B7280",
  lineHeight: 1.5
})

// Actions
actions=I(card, {
  type: "frame",
  layout: "horizontal",
  padding: [16, 24, 24, 24],
  gap: 12
})
```

## Form Pattern

```javascript
// Form container
form=I(parent, {
  type: "frame",
  name: "Form",
  width: 400,
  layout: "vertical",
  padding: [32, 32, 32, 32],
  gap: 24,
  cornerRadius: [16, 16, 16, 16],
  fill: "#FFFFFF"
})

// Form title
formTitle=I(form, {
  type: "text",
  content: "Sign Up",
  fontSize: 28,
  fontWeight: "700",
  textColor: "#1A1A1A"
})

// Input group factory (repeat for each field)
field1=I(form, {
  type: "frame",
  layout: "vertical",
  gap: 8,
  width: "fill_container"
})
label1=I(field1, {
  type: "text",
  content: "Email",
  fontSize: 14,
  fontWeight: "500",
  textColor: "#374151"
})
input1=I(field1, {
  type: "frame",
  width: "fill_container",
  height: 48,
  padding: [12, 16, 12, 16],
  cornerRadius: [8, 8, 8, 8],
  fill: "#F9FAFB",
  stroke: "#D1D5DB",
  strokeThickness: 1
})
placeholder1=I(input1, {
  type: "text",
  content: "you@example.com",
  fontSize: 16,
  textColor: "#9CA3AF"
})

// Submit button
submit=I(form, {
  type: "frame",
  width: "fill_container",
  height: 48,
  layout: "horizontal",
  alignItems: "center",
  justifyContent: "center",
  cornerRadius: [8, 8, 8, 8],
  fill: "#2563EB"
})
submitText=I(submit, {
  type: "text",
  content: "Continue",
  fontSize: 16,
  fontWeight: "600",
  textColor: "#FFFFFF"
})
```

## Navigation Pattern

```javascript
// Nav container
nav=I(document, {
  type: "frame",
  name: "Navigation",
  width: "fill_container",
  height: 72,
  layout: "horizontal",
  alignItems: "center",
  justifyContent: "space-between",
  padding: [0, 48, 0, 48],
  fill: "#FFFFFF"
})

// Logo area
logo=I(nav, {
  type: "frame",
  layout: "horizontal",
  alignItems: "center",
  gap: 8
})
logoText=I(logo, {
  type: "text",
  content: "Brand",
  fontSize: 20,
  fontWeight: "700",
  textColor: "#1A1A1A"
})

// Nav links
links=I(nav, {
  type: "frame",
  layout: "horizontal",
  gap: 32
})
link1=I(links, {type: "text", content: "Features", fontSize: 15, textColor: "#4B5563"})
link2=I(links, {type: "text", content: "Pricing", fontSize: 15, textColor: "#4B5563"})
link3=I(links, {type: "text", content: "About", fontSize: 15, textColor: "#4B5563"})

// CTA
navCta=I(nav, {
  type: "frame",
  padding: [10, 20, 10, 20],
  cornerRadius: [6, 6, 6, 6],
  fill: "#1A1A1A"
})
navCtaText=I(navCta, {
  type: "text",
  content: "Sign Up",
  fontSize: 14,
  fontWeight: "500",
  textColor: "#FFFFFF"
})
```

## Bento Grid Pattern

```javascript
// Bento container
bento=I(document, {
  type: "frame",
  name: "Bento Grid",
  width: 1200,
  layout: "grid",
  gap: 16,
  padding: [48, 48, 48, 48]
})

// Large feature cell (spans 2 cols, 2 rows)
feature=I(bento, {
  type: "frame",
  width: 584,
  height: 400,
  cornerRadius: [16, 16, 16, 16],
  fill: "#1A1A1A"
})

// Small cells
cell1=I(bento, {
  type: "frame",
  width: 284,
  height: 192,
  cornerRadius: [12, 12, 12, 12],
  fill: "#F5F5F5"
})
cell2=I(bento, {
  type: "frame",
  width: 284,
  height: 192,
  cornerRadius: [12, 12, 12, 12],
  fill: "#E8E8E8"
})
```

## Using Design System Components

When design system components are available:

```javascript
// First, search for available components
// mcp__pencil__batch_get({ patterns: [{ reusable: true }] })

// Insert component instance
card=I(parent, {type: "ref", ref: "CardComponent"})

// Customize instance properties
U(card+"/title", {content: "Custom Title"})
U(card+"/description", {content: "Custom description"})
U(card, {width: 400})

// Replace slot content
newContent=R(card+"/imageSlot", {
  type: "frame",
  fill: {type: "linear", stops: [...]}
})
```

## Error Prevention

1. **Always create bindings** for Insert/Copy/Replace
2. **Never reference copied descendant IDs** - use `descendants` in Copy instead
3. **Keep batches under 25 operations**
4. **Validate parent exists** before inserting
5. **Use document for top-level frames**, frame IDs for nested content
