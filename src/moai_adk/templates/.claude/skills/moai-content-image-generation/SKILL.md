---
name: moai-content-image-generation
description: Generating AI images with prompt engineering for GPT-Image, DALL-E, Midjourney. Craft effective prompts, maintain style consistency, and optimize quality. Use when generating blog images, creating visual content, or automating image production.
allowed-tools: Read, Bash
version: 1.0.0
tier: content
created: 2025-10-31
---

# Content: AI Image Generation and Prompt Engineering

## What it does

Provides prompt engineering strategies for AI image generation tools (GPT-4o image generation, DALL-E 3, Midjourney v7, Stable Diffusion) to create high-quality, consistent visuals for blog posts, social media, and marketing content.

## When to use

- Generating featured images for blog posts
- Creating consistent visual branding
- Producing social media graphics
- Designing illustrations for documentation
- Automating image creation workflows

## Key Patterns

### 1. Prompt Structure (Universal Formula)

**Pattern**: Subject + Style + Details + Constraints

\`\`\`
[Subject] + [Style Modifier] + [Details] + [Composition] + [Negative Prompts]

Example:
"A modern developer workspace, minimalist style, MacBook Pro on wooden desk, 
natural window lighting, 16:9 aspect ratio, photorealistic, 
--no clutter, no people"
\`\`\`

### 2. GPT-4o Image Generation (2025)

**Pattern**: ChatGPT native image generation (replaced DALL-E 3)

\`\`\`typescript
// OpenAI API (GPT-4o with vision and generation)
const response = await openai.chat.completions.create({
  model: "gpt-4o",
  messages: [
    {
      role: "user",
      content: "Generate an image: A futuristic tech blog header featuring Next.js 15 logo with gradient background, modern minimalist style, 1200x630px"
    }
  ],
  max_tokens: 1024
});

// GPT-4o can now:
// - Render accurate text in images (Midjourney cannot)
// - Understand numbers and positioning
// - Track multiple objects precisely
\`\`\`

**GPT-4o Strengths (vs DALL-E 3)**:
- ✅ Accurate text rendering
- ✅ Better number understanding
- ✅ Precise spatial positioning
- ✅ Multi-object tracking
- ✅ Literal prompt following

### 3. Midjourney v7 Prompts (2025)

**Pattern**: Advanced prompt engineering with parameters

\`\`\`
/imagine prompt: [detailed description] --v 7 --ar 16:9 --style raw --q 2

Example:
/imagine prompt: modern tech startup office, glass walls, standing desks, 
monitors displaying code, natural sunlight, professional photography, 
vibrant colors --v 7 --ar 16:9 --style raw --quality 2 --no people, clutter

Parameters:
--v 7              # Midjourney version 7 (latest 2025)
--ar 16:9          # Aspect ratio (1:1, 4:3, 16:9, 9:16)
--style raw        # Less processed, more photographic
--quality 2        # Higher quality (1 = fast, 2 = best)
--no [elements]    # Negative prompts (exclude items)
--chaos 0-100      # Variation level (0 = consistent)
--seed 12345       # Reproducible results
\`\`\`

**Midjourney v7 Strengths**:
- ✅ Artistic, cinematic style
- ✅ Dreamy, detailed visuals
- ✅ Bold aesthetics
- ❌ Cannot render accurate text reliably

### 4. Style Consistency Formula

**Pattern**: Maintain brand visual consistency

\`\`\`
Brand Style Definition:
"Professional tech blog aesthetic, minimal modern design, 
gradient backgrounds [#4F46E5 to #818CF8], 
clean sans-serif typography, subtle shadows, 
high contrast, 8k resolution"

Template Prompt:
[Content Description] + [Brand Style Definition] + [Technical Specs]

Example:
"Illustration of API architecture diagram + [Brand Style] + 
vector art, flat design, 1920x1080px"
\`\`\`

### 5. Prompt Engineering Tricks (2025)

**Pattern**: Clarity + Style + Constraints + Iteration

\`\`\`markdown
## Basic → Advanced Progression

❌ Basic: "A computer"
🟡 Better: "A MacBook Pro on a desk"
✅ Best: "A silver MacBook Pro 16-inch on a minimalist wooden desk, 
natural window lighting from left, shallow depth of field, 
professional product photography, 4k resolution"

## Key Modifiers

**Artistic Style**:
- photorealistic, hyperrealistic, cinematic
- minimalist, modern, abstract
- watercolor, oil painting, digital art
- isometric, flat design, 3D render

**Lighting**:
- natural sunlight, golden hour, studio lighting
- dramatic shadows, soft diffused light
- neon glow, ambient occlusion

**Camera**:
- shallow depth of field, bokeh
- wide angle lens, macro shot
- drone view, bird's eye view
- 50mm lens, f/1.8 aperture

**Quality**:
- 4k, 8k resolution
- highly detailed, intricate
- professional photography
- award-winning

**Composition**:
- rule of thirds, centered composition
- symmetrical, asymmetrical balance
- negative space, minimalist
\`\`\`

### 6. Multi-Image Consistency

**Pattern**: Use seed values and style references

\`\`\`bash
# Midjourney: Use --seed for consistency
/imagine prompt: tech blog header, [description] --seed 12345 --v 7
/imagine prompt: tech blog thumbnail, [description] --seed 12345 --v 7

# DALL-E/GPT-4o: Reference previous images
"Generate similar to previous image but with [changes]"
\`\`\`

## Best Practices

### Prompt Writing
- Be specific and descriptive (avoid vague terms)
- Include composition details (framing, angle, perspective)
- Specify style explicitly (photorealistic vs artistic)
- Use negative prompts to exclude unwanted elements
- Iterate: generate → refine → regenerate

### Tool Selection (2025)
- **GPT-4o**: When text accuracy is critical (logos, diagrams, infographics)
- **Midjourney v7**: For artistic, cinematic, visually stunning images
- **DALL-E 3**: For literal prompt interpretation (integrated in ChatGPT)
- **Stable Diffusion**: For full control, local hosting, custom models

### Workflow Optimization
- Save successful prompts as templates
- Document seed values for consistency
- Use style guides for brand alignment
- Batch generate variations
- Post-process with image editors (brightness, crop, resize)

### Legal & Ethical
- Verify licensing terms for commercial use
- Disclose AI-generated content when required
- Avoid generating copyrighted characters/brands
- Respect image rights and attribution

## Resources

- Midjourney Documentation: https://docs.midjourney.com/
- OpenAI DALL-E Guide: https://platform.openai.com/docs/guides/images
- Prompt Engineering Guide: https://www.latestly.ai/p/prompt-engineering-tricks-for-image-generation-models-2025-guide
- CreateVision AI Analysis: https://createvision.ai/guides/gpt5-image-generation-analysis

## Examples

**Example 1: Blog Featured Image (GPT-4o)**

\`\`\`
Generate a professional blog header image:

Subject: Modern web development workspace
Details: MacBook Pro with Next.js code on screen, coffee cup, 
notebook with React logo sticker, succulent plant
Style: Clean, minimalist, natural lighting, shallow depth of field
Composition: Desk shot from 45-degree angle, rule of thirds
Colors: Warm tones, natural wood, green accents
Text: "Next.js 15 Guide" in modern sans-serif font (top-right)
Resolution: 1200x630px (OG image)
\`\`\`

**Example 2: Social Media Graphic (Midjourney v7)**

\`\`\`
/imagine prompt: abstract technology background with flowing data streams, 
purple and blue gradient, geometric shapes, modern digital art style, 
glowing particles, 4k resolution, professional design --v 7 --ar 1:1 --style raw --q 2 --no text, logos
\`\`\`

**Example 3: Icon Set (Consistent Style)**

\`\`\`
Base Prompt Template:
"[Icon subject], flat design, minimalist, single color #4F46E5, 
transparent background, vector art, 512x512px, centered composition"

Variations:
- "Gear icon for settings, [Base Template]"
- "User profile icon, [Base Template]"
- "Document icon, [Base Template]"

Use same seed value across all generations for consistency.
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with GPT-4o, Midjourney v7, prompt engineering strategies

## Works well with
- \`moai-content-seo-optimization\` (Optimize image alt text and file names)
- \`moai-saas-wordpress-publishing\` (Upload featured images)
- \`moai-content-blog-strategy\` (Plan visual content calendar)
