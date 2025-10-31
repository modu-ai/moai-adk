# moai-content-image-generation

Generating AI images with prompt engineering for DALL-E, Midjourney, Stable Diffusion. Craft effective prompts, maintain style consistency, and optimize quality.

## Quick Start

AI image generation tools create unique, branded imagery for blog posts and marketing materials. Use this skill when creating cover images, illustrations, diagrams, or visual content for articles.

## Core Patterns

### Pattern 1: Prompt Engineering for Quality Images

**Pattern**: Write effective prompts that generate high-quality, consistent images.

```typescript
// lib/image-generation.ts
interface ImageGenerationPrompt {
  concept: string;
  style: string;
  medium: string;
  lighting: string;
  perspective: string;
  quality: string;
  aspectRatio: string;
}

function buildDetailedPrompt(params: ImageGenerationPrompt): string {
  return [
    // Concept and subject
    `A beautiful ${params.concept}`,

    // Style specification
    `in the style of ${params.style}`,

    // Medium and technique
    `${params.medium}`,

    // Lighting and atmosphere
    `with ${params.lighting}`,

    // Perspective and composition
    `${params.perspective}`,

    // Quality and detail
    `${params.quality}, highly detailed, professional, sharp focus`,

    // Negative prompts (what to avoid)
    `(no watermark, no text, no blurry, no artifacts)`,

    // Aspect ratio hint
    `${params.aspectRatio}`,
  ].join(', ');
}

// Example prompts for blog content
export const blogCoverPrompts = {
  nodejs_performance: {
    concept: 'server farm with glowing nodes and data streams',
    style: 'modern minimalist, geometric',
    medium: '3D render, digital art',
    lighting: 'neon blue and purple ambient lighting',
    perspective: 'isometric view, center composition',
    quality: '8k resolution, cinematic',
    aspectRatio: '16:9 for web banner',
  },

  cybersecurity: {
    concept: 'digital shield protecting data from threats',
    style: 'cyberpunk, tech noir',
    medium: 'digital illustration, concept art',
    lighting: 'dramatic red and blue neon lighting',
    perspective: 'dynamic angle, depth of field',
    quality: 'ultra detailed, professional quality',
    aspectRatio: '16:9',
  },

  frontend_design: {
    concept: 'colorful user interface components floating in space',
    style: 'glassmorphism, modern UI design',
    medium: '3D render, digital art',
    lighting: 'soft ambient with accent highlights',
    perspective: 'slightly angled, balanced composition',
    quality: 'vibrant colors, sharp details',
    aspectRatio: '16:9',
  },

  database_architecture: {
    concept: 'interconnected database nodes with flowing data',
    style: 'technical illustration, minimalist modern',
    medium: 'vector art style 3D render',
    lighting: 'cool blue lighting with tech aesthetics',
    perspective: 'orthographic projection, technical view',
    quality: 'clean lines, professional, technical precision',
    aspectRatio: '16:9',
  },
};

// Generate prompts for article series
export function generateSeriesPrompts(
  topic: string,
  partNumber: number,
  totalParts: number
): string {
  const progressionStyles = [
    'minimalist flat design',
    'geometric modern art',
    'digital illustration',
    '3D render',
    'technical schematic',
  ];

  const style = progressionStyles[partNumber % progressionStyles.length];

  return buildDetailedPrompt({
    concept: `${topic} concept visual - part ${partNumber} of ${totalParts} series`,
    style: `${style}, coherent visual series`,
    medium: 'digital art, professional illustration',
    lighting: 'consistent brand lighting',
    perspective: 'composition optimized for social media',
    quality: 'consistent with series style, high quality',
    aspectRatio: '16:9',
  });
}
```

**When to use**:
- Creating blog cover images
- Generating article illustrations
- Creating social media graphics
- Maintaining visual consistency

**Key benefits**:
- Unique, custom imagery
- Brand-consistent visuals
- Cost-effective compared to stock photos
- Fast iteration and variations

### Pattern 2: Integrating with DALL-E API

**Pattern**: Generate and manage images using OpenAI's DALL-E API.

```typescript
// lib/dalle-integration.ts
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY,
});

export async function generateBlogImage(
  prompt: string,
  options?: {
    size?: '1024x1024' | '1792x1024' | '1024x1792';
    quality?: 'standard' | 'hd';
    n?: number;
  }
) {
  const response = await openai.images.generate({
    model: 'dall-e-3',
    prompt: prompt,
    size: options?.size || '1792x1024', // Wide format for blog covers
    quality: options?.quality || 'hd', // High definition for professional use
    n: options?.n || 1,
  });

  return response.data[0].url;
}

// Generate multiple variations
export async function generateImageVariations(
  prompt: string,
  count: number = 3
) {
  const results = [];

  for (let i = 0; i < count; i++) {
    const imageUrl = await generateBlogImage(prompt, { n: 1 });
    results.push({
      variation: i + 1,
      url: imageUrl,
      prompt: prompt,
      generatedAt: new Date(),
    });
  }

  return results;
}

// Save generated images
export async function saveGeneratedImage(
  imageUrl: string,
  filename: string,
  outputDir: string = 'public/blog-images'
) {
  const response = await fetch(imageUrl);
  const buffer = await response.buffer();

  const fs = await import('fs/promises');
  const path = await import('path');

  await fs.mkdir(outputDir, { recursive: true });
  const filepath = path.join(outputDir, filename);

  await fs.writeFile(filepath, buffer);

  return filepath;
}

// Batch generate images for blog series
export async function generateBlogSeriesImages(
  seriesTopic: string,
  articles: Array<{ title: string; slug: string }>,
  styleOverride?: string
) {
  const results = [];

  for (let i = 0; i < articles.length; i++) {
    const article = articles[i];
    const prompt = generateSeriesPrompts(seriesTopic, i + 1, articles.length);

    try {
      const imageUrl = await generateBlogImage(prompt);
      const filename = `${article.slug}-cover.jpg`;
      const savedPath = await saveGeneratedImage(imageUrl, filename);

      results.push({
        article: article.title,
        imageUrl,
        savedPath,
        success: true,
      });

      console.log(`Generated image for: ${article.title}`);
    } catch (error) {
      results.push({
        article: article.title,
        success: false,
        error: String(error),
      });
    }
  }

  return results;
}
```

**When to use**:
- Batch generating images for multiple articles
- Creating variations for A/B testing
- Automating image generation workflow
- Maintaining visual consistency

**Key benefits**:
- Programmatic image generation
- Batch processing capabilities
- Consistent style application
- Easy integration in CI/CD

### Pattern 3: Image Optimization & Delivery

**Pattern**: Optimize generated images for web and different platforms.

```typescript
// lib/image-optimization.ts
import sharp from 'sharp';
import path from 'path';

export interface ImageOptimizationOptions {
  formats: ('webp' | 'jpg' | 'png')[];
  sizes: number[];
  quality: number;
}

export async function optimizeImageForWeb(
  inputPath: string,
  outputDir: string,
  options: ImageOptimizationOptions = {
    formats: ['webp', 'jpg'],
    sizes: [640, 1024, 1920],
    quality: 80,
  }
) {
  const results = [];

  for (const format of options.formats) {
    for (const size of options.sizes) {
      const filename = `image-${size}.${format}`;
      const outputPath = path.join(outputDir, filename);

      let pipeline = sharp(inputPath);

      // Resize
      pipeline = pipeline.resize(size, Math.round(size * 9 / 16), {
        fit: 'cover',
        position: 'center',
      });

      // Format-specific optimization
      switch (format) {
        case 'webp':
          pipeline = pipeline.webp({ quality: options.quality });
          break;
        case 'jpg':
          pipeline = pipeline.jpeg({ quality: options.quality, progressive: true });
          break;
        case 'png':
          pipeline = pipeline.png({ quality: options.quality });
          break;
      }

      await pipeline.toFile(outputPath);

      results.push({
        format,
        size,
        path: outputPath,
      });
    }
  }

  return results;
}

// Generate responsive image markup
export function generateResponsiveImageMarkup(
  baseFilename: string,
  altText: string,
  sizes: number[] = [640, 1024, 1920]
) {
  const sources = sizes.map(size => ({
    webp: `image-${size}.webp`,
    jpg: `image-${size}.jpg`,
  }));

  return `
<picture>
  ${sources.map(src => `<source srcset="/${src.webp}" type="image/webp" media="(min-width: ${src.webp.match(/\d+/)?.[0]}px)">`).join('\n  ')}
  <img src="/image-1024.jpg" alt="${altText}" loading="lazy">
</picture>
  `.trim();
}

// Generate Markdown image reference with optimization
export function generateMarkdownImage(
  filename: string,
  altText: string,
  title: string
) {
  return `![${altText}](/images/${filename} "${title}")`;
}
```

**When to use**:
- Optimizing images for web performance
- Creating responsive image markup
- Reducing page load times
- Supporting multiple formats

**Key benefits**:
- Faster page load times
- Better SEO through Core Web Vitals
- Responsive images for all devices
- Modern format support (WebP)

## Progressive Disclosure

### Level 1: Basic Image Generation
- Single image generation with DALL-E
- Basic prompt engineering
- Save generated images
- Manual image optimization

### Level 2: Advanced Generation
- Batch generation for multiple articles
- Style consistency across series
- Prompt variation and A/B testing
- Automated image optimization

### Level 3: Expert Scale
- Advanced prompt engineering
- Integration with design systems
- Performance optimization
- Brand-consistent automated generation

## Works Well With

- **OpenAI DALL-E**: Professional image generation
- **Midjourney**: Discord-based high-quality generation
- **Stable Diffusion**: Open-source local generation
- **Sharp.js**: Image optimization
- **Next.js**: Image component with optimization
- **Cloudinary**: Image delivery and transformation

## References

- **DALL-E API**: https://platform.openai.com/docs/guides/images
- **Midjourney**: https://www.midjourney.com/docs/
- **Stable Diffusion**: https://stability.ai/
- **Image Prompting Guide**: https://openai.com/blog/dall-e-api-updates/
- **Sharp.js**: https://sharp.pixelplumbing.com/
