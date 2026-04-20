#!/usr/bin/env node

/**
 * Favicon Generator Script
 *
 * Generates favicon files from the existing og.png
 * Requirements: npm install --save-dev sharp
 */

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import sharp from "sharp";

const sourceImage = path.join(process.cwd(), "public", "og.png");
const outputDir = path.join(process.cwd(), "public");

const faviconSizes = [
  { name: "favicon.ico", sizes: [32, 16], format: "ico" },
  { name: "favicon-16x16.png", size: 16, format: "png" },
  { name: "favicon-32x32.png", size: 32, format: "png" },
  { name: "apple-touch-icon.png", size: 180, format: "png" },
  { name: "icon-192.png", size: 192, format: "png" },
  { name: "icon-512.png", size: 512, format: "png" },
];

async function generateFavicons() {
  console.log("🎨 Generating favicons from og.png...");

  try {
    // Check if source image exists
    await readFile(sourceImage);
  } catch {
    console.error("❌ Error: og.png not found in public directory");
    process.exit(1);
  }

  // Generate PNG favicons
  for (const { name, size, format } of faviconSizes) {
    if (format === "png") {
      const outputPath = path.join(outputDir, name);
      await sharp(sourceImage)
        .resize(size, size, { fit: "cover", position: "center" })
        .png()
        .toFile(outputPath);
      console.log(`✅ Generated ${name} (${size}x${size})`);
    }
  }

  // Generate ICO file with multiple sizes
  const icoPath = path.join(outputDir, "favicon.ico");

  // For ICO, we need to create it differently
  // Sharp doesn't directly support ICO, so we'll create a simple PNG version
  // and note that browsers will accept PNG as favicon
  await sharp(sourceImage)
    .resize(32, 32, { fit: "cover", position: "center" })
    .png()
    .toFile(icoPath);
  console.log(`✅ Generated favicon.ico (32x32)`);

  // Create SVG favicon (simple version with MoAI emoji)
  const svgContent = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
  <text y="0.9em" font-size="90">🗿</text>
</svg>`;
  const svgPath = path.join(outputDir, "favicon.svg");
  await writeFile(svgPath, svgContent);
  console.log(`✅ Generated favicon.svg`);

  console.log("\n✨ All favicons generated successfully!");
  console.log("\n📋 Generated files:");
  console.log("  - favicon.ico (32x32)");
  console.log("  - favicon.svg (scalable)");
  console.log("  - favicon-16x16.png");
  console.log("  - favicon-32x32.png");
  console.log("  - apple-touch-icon.png (180x180)");
  console.log("  - icon-192.png (192x192)");
  console.log("  - icon-512.png (512x512)");
  console.log(
    "\n💡 Tip: You may want to use a dedicated favicon generator for better quality:",
  );
  console.log("   https://realfavicongenerator.net/");
}

generateFavicons().catch(console.error);
