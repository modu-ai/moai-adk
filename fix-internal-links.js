#!/usr/bin/env node

// @CODE:TAG-MIGRATION-003
// Tool to fix internal links in migrated content

import fs from 'fs/promises';
import path from 'path';
import { promises as fsSync } from 'fs';

// Configuration
const CONTENT_DIR = '/Users/goos/MoAI/MoAI-ADK/docs/content';
const LINK_PATTERNS = [
  // Markdown links like [text](./path/file.md)
  /\[([^\]]+)\]\(([^)]+)\)/g,
  // MDX links like [text](./path/file.mdx)
  /\[([^\]]+)\]\(([^)]+)\)/g,
];

/**
 * Fix internal links in migrated content
 */
export async function fixInternalLinks() {
  console.log('🔧 Fixing internal links in migrated content...\n');

  const markdownFiles = await findMarkdownFiles(CONTENT_DIR);
  let fixedLinks = 0;

  for (const file of markdownFiles) {
    const content = await fs.readFile(file, 'utf-8');
    const updatedContent = await fixLinksInContent(content, file);

    if (updatedContent !== content) {
      await fs.writeFile(file, updatedContent, 'utf-8');
      fixedLinks++;
      console.log(`✅ Fixed internal links in: ${path.relative(CONTENT_DIR, file)}`);
    }
  }

  console.log(`\n🎉 Fixed internal links in ${fixedLinks} files!`);
  return fixedLinks;
}

/**
 * Find all markdown files in content directory
 */
async function findMarkdownFiles(dir) {
  const files = [];
  await scanDirectory(dir, ['.md', '.mdx'], files);
  return files;
}

async function scanDirectory(dir, extensions, files) {
  try {
    const items = await fs.readdir(dir, { withFileTypes: true });

    for (const item of items) {
      const fullPath = path.join(dir, item.name);

      if (item.isDirectory() && !item.name.startsWith('.')) {
        await scanDirectory(fullPath, extensions, files);
      } else if (item.isFile() && (item.name.endsWith('.md') || item.name.endsWith('.mdx'))) {
        files.push(fullPath);
      }
    }
  } catch (error) {
    console.error(`Error scanning directory ${dir}:`, error.message);
  }
}

/**
 * Fix links in content based on current file location
 */
async function fixLinksInContent(content, currentFilePath) {
  let updatedContent = content;

  // Find all markdown/MDX links
  const linkPattern = /\[([^\]]+)\]\(([^)]+)\)/g;
  let match;
  const links = [];

  while ((match = linkPattern.exec(content)) !== null) {
    const [fullMatch, text, url] = match;
    // Skip external links (http, https, mailto, etc.)
    if (!url.match(/^(https?|mailto|tel|data):/) && !url.startsWith('//')) {
      links.push({ fullMatch, text, url, startIndex: match.index, endIndex: match.index + fullMatch.length });
    }
  }

  // Process links from end to start to avoid index shifting
  for (let i = links.length - 1; i >= 0; i--) {
    const link = links[i];
    const fixedUrl = await fixUrl(link.url, currentFilePath);

    if (fixedUrl !== link.url) {
      const replacement = `[${link.text}](${fixedUrl})`;
      updatedContent = updatedContent.substring(0, link.startIndex) +
                      replacement +
                      updatedContent.substring(link.endIndex);
    }
  }

  return updatedContent;
}

/**
 * Fix URL based on current file location
 */
async function fixUrl(url, currentFilePath) {
  let fixedUrl = url.trim();

  // Remove ./ or ../ prefix for relative paths
  if (fixedUrl.startsWith('./')) {
    fixedUrl = fixedUrl.substring(2);
  }

  // Get current file directory
  const currentDir = path.dirname(currentFilePath);
  const contentRoot = CONTENT_DIR;

  // Calculate relative path from current file to content root
  const relativePath = path.relative(currentDir, contentRoot);

  // Handle different types of links
  if (fixedUrl.startsWith('../')) {
    // Parent directory links - adjust for new structure
    fixedUrl = handleParentLinks(fixedUrl, currentDir, contentRoot);
  } else if (!fixedUrl.startsWith('/')) {
    // Relative links within same directory
    fixedUrl = ensureLanguageContext(fixedUrl, currentDir);
  } else if (fixedUrl.startsWith('/')) {
    // Absolute links - remove leading slash
    fixedUrl = fixedUrl.substring(1);
  }

  return fixedUrl;
}

/**
 * Handle ../ navigation links
 */
function handleParentLinks(url, currentDir, contentRoot) {
  const parts = url.split('/').filter(p => p !== '');
  let currentPath = currentDir;

  // Process ../ navigation
  for (const part of parts) {
    if (part === '..') {
      currentPath = path.dirname(currentPath);
      // Don't go above content directory
      if (path.relative(currentPath, contentRoot).startsWith('..')) {
        currentPath = contentRoot;
      }
    } else {
      currentPath = path.join(currentPath, part);
    }
  }

  // Get relative path from content root
  const relativePath = path.relative(contentRoot, currentPath);
  return relativePath;
}

/**
 * Ensure language context is preserved for relative links
 */
async function ensureLanguageContext(url, currentDir) {
  const currentLang = getCurrentLanguage(currentDir);

  // If current file is in a language directory, ensure links stay within that context
  if (currentLang && ['ko', 'en', 'ja', 'zh'].includes(currentLang)) {
    // Check if target file exists in same language directory
    const targetPathInLang = path.join(currentDir, url);
    try {
      await fs.access(targetPathInLang);
      return url; // File exists in same language directory, keep as-is
    } catch {
      // File doesn't exist, continue to check language directory
    }

    // If not, try to find in current language
    const langDir = path.join(path.dirname(currentDir), currentLang);
    const targetInLang = path.join(langDir, url);
    try {
      await fs.access(targetInLang);
      return path.relative(currentDir, targetInLang);
    } catch {
      // File doesn't exist in language directory either
    }
  }

  return url;
}

/**
 * Get current language directory
 */
function getCurrentLanguage(currentDir) {
  const pathParts = path.relative(CONTENT_DIR, currentDir).split(path.sep);
  if (pathParts.length > 0 && ['ko', 'en', 'ja', 'zh'].includes(pathParts[0])) {
    return pathParts[0];
  }
  return null;
}

// Run if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  fixInternalLinks().catch(console.error);
}