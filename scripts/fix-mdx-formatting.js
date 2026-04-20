#!/usr/bin/env node

/**
 * Fix MDX formatting issues in Nextra documentation
 *
 * This script fixes the following MDX rendering errors:
 * 1. **text(value) -> **text** (value) - Add space before parenthesis
 * 2. **[text](url) -> **text** [text](url) - Separate links from bold
 * 3. **text**value - **text** value - Fix bold syntax issues
 *
 * Usage: node scripts/fix-mdx-formatting.js
 */

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const CONTENT_DIR = path.join(__dirname, "../content");

// Counters for reporting
let totalFiles = 0;
let modifiedFiles = 0;
let totalReplacements = 0;

/**
 * Fix MDX formatting issues in a single file
 * @param {string} filePath - Path to the file
 * @returns {number} - Number of replacements made
 */
function fixFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, "utf8");
    const originalContent = content;
    let replacements = 0;

    // Pattern 1: **[text](url) -> **text** [text](url)
    // This handles bold links which need to be separated
    const boldLinkPattern = /\*\*([^\]]+)\]\([^)]+\)/g;
    content = content.replace(boldLinkPattern, (match) => {
      replacements++;
      return match.replace(/\]\(/g, "**] **(");
    });

    // Pattern 2: **text(value) -> **text** (value)
    // Handles bold text immediately followed by parenthesis
    const boldParenPattern = /\*\*([^*]+?)\*\*\([^)]*\)/g;
    content = content.replace(boldParenPattern, (_match, p1) => {
      replacements++;
      return `**${p1}** **(`;
    });

    // Pattern 3: **LSP(Language -> **LSP** **(Language
    // More aggressive pattern for nested parentheses
    const pattern3 = /\*\*([A-Z]+)\(([^)]+)\)\*\*/g;
    content = content.replace(pattern3, (_match, p1, p2) => {
      replacements++;
      return `**${p1}** **(${p2})**`;
    });

    // Write back if changed
    if (content !== originalContent) {
      fs.writeFileSync(filePath, content, "utf8");
      return replacements;
    }
    return 0;
  } catch (error) {
    console.error(`Error processing ${filePath}:`, error.message);
    return 0;
  }
}

/**
 * Recursively find and fix all MDX files
 * @param {string} dir - Directory to scan
 */
function scanDirectory(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      scanDirectory(fullPath);
    } else if (entry.isFile() && entry.name.endsWith(".mdx")) {
      totalFiles++;
      const replacements = fixFile(fullPath);
      if (replacements > 0) {
        modifiedFiles++;
        totalReplacements += replacements;
        console.log(
          `✓ Fixed ${fullPath.replace(CONTENT_DIR, "")} (${replacements} replacements)`,
        );
      }
    }
  }
}

// Main execution
console.log("🔧 Fixing MDX formatting issues in Nextra documentation...\n");

scanDirectory(CONTENT_DIR);

console.log("\n📊 Summary:");
console.log(`   Total files scanned: ${totalFiles}`);
console.log(`   Files modified: ${modifiedFiles}`);
console.log(`   Total replacements: ${totalReplacements}`);

if (modifiedFiles > 0) {
  console.log("\n✅ MDX formatting fixes completed successfully!");
  console.log("💡 Tip: Run `npm run build` to verify the changes.");
} else {
  console.log("\n✨ No MDX formatting issues found!");
}
