#!/usr/bin/env node

/*
 * MDX Translation Script for MoAI-ADK Documentation
 *
 * Features:
 * - Preserves code blocks (does not translate content inside triple backtick fences)
 * - Preserves frontmatter structure
 * - Preserves technical terms (SPEC, DDD, TRUST 5)
 * - Generates _meta.ts files for each locale
 * - Supports batch translation of multiple files
 *
 * Usage:
 *   node scripts/translate.mjs [options]
 *
 * Options:
 *   --source-dir <dir>    Source directory (default: content/)
 *   --target-locale <loc> Target locale: en, ja, zh (default: all)
 *   --files <pattern>     File pattern (default: all .mdx files)
 *   --dry-run            Show what would be translated without writing
 *   --api-key <key>       Translation API key (if using external API)
 */

import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, statSync } from 'fs';
import { join, dirname, relative, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Configuration
const CONFIG = {
  // Technical terms that should not be translated
  preserveTerms: [
    'SPEC',
    'DDD',
    'TRUST 5',
    'MoAI',
    'MoAI-ADK',
    'Claude Code',
    'EARS',
    'TDD',
    'OWASP',
    'AST',
    'LSP',
    'MCP',
    'MDX',
    'JSON',
    'YAML',
    'API',
    'CLI',
    'URL',
    'HTTP',
    'JWT',
    'SQL',
    'NoSQL',
    'CI/CD',
    'README',
    'CHANGELOG',
  ],

  // Code block language identifiers to preserve
  codeLanguages: [
    'bash', 'sh', 'zsh',
    'javascript', 'js', 'typescript', 'ts', 'jsx', 'tsx',
    'python', 'py',
    'go', 'golang',
    'java', 'kotlin', 'scala',
    'rust', 'rs',
    'c', 'cpp', 'csharp', 'cs',
    'ruby', 'php', 'elixir', 'r',
    'swift', 'dart',
    'yaml', 'yml', 'json', 'toml',
    'markdown', 'md',
    'mermaid',
    'diff', 'patch',
  ],

  // Locale mappings
  locales: {
    en: { name: 'English', dir: 'en' },
    ja: { name: 'Japanese', dir: 'ja' },
    zh: { name: 'Chinese', dir: 'zh' },
    ko: { name: 'Korean', dir: 'ko' },
  },

  // Directories to skip
  skipDirs: ['node_modules', '.git', 'dist', 'build'],
};

/**
 * Parse command line arguments
 */
function parseArgs() {
  const args = process.argv.slice(2);
  const options = {
    sourceDir: 'content',
    targetLocale: null,
    files: '**/*.mdx',
    dryRun: false,
    apiKey: null,
  };

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--source-dir':
        options.sourceDir = args[++i];
        break;
      case '--target-locale':
        options.targetLocale = args[++i];
        break;
      case '--files':
        options.files = args[++i];
        break;
      case '--dry-run':
        options.dryRun = true;
        break;
      case '--api-key':
        options.apiKey = args[++i];
        break;
      case '--help':
        console.log(`
MDX Translation Script for MoAI-ADK Documentation

Usage:
  node scripts/translate.mjs [options]

Options:
  --source-dir <dir>    Source directory (default: content/)
  --target-locale <loc> Target locale: en, ja, zh (default: all)
  --files <pattern>     File pattern to translate (default: **/*.mdx)
  --dry-run            Show what would be translated without writing
  --api-key <key>      Translation API key (for external APIs)
  --help               Show this help message

Examples:
  # Translate all Korean MDX files to all languages
  node scripts/translate.mjs

  # Translate only to English
  node scripts/translate.mjs --target-locale en

  # Dry run to see what would be translated
  node scripts/translate.mjs --dry-run

  # Translate specific directory
  node scripts/translate.mjs --source-dir content/core-concepts --target-locale en
        `);
        process.exit(0);
    }
  }

  return options;
}

/**
 * Recursively find all MDX files in a directory
 */
function findMdxFiles(dir, baseDir = dir) {
  const files = [];

  if (!existsSync(dir)) {
    return files;
  }

  const entries = readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = join(dir, entry.name);

    if (entry.isDirectory()) {
      if (!CONFIG.skipDirs.includes(entry.name)) {
        files.push(...findMdxFiles(fullPath, baseDir));
      }
    } else if (entry.isFile() && entry.name.endsWith('.mdx')) {
      files.push(fullPath);
    }
  }

  return files;
}

/**
 * Parse MDX content into segments
 * Separates code blocks, frontmatter, and prose for selective translation
 */
function parseMdxContent(content) {
  const segments = [];
  let remaining = content;
  let position = 0;

  // Extract frontmatter (YAML between --- markers)
  const frontmatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n/;
  const frontmatterMatch = remaining.match(frontmatterRegex);

  if (frontmatterMatch) {
    segments.push({
      type: 'frontmatter',
      content: frontmatterMatch[0],
      translatable: false,
      start: position,
      end: position + frontmatterMatch[0].length,
    });
    remaining = remaining.slice(frontmatterMatch[0].length);
    position += frontmatterMatch[0].length;
  }

  // Extract import statements (preserve as-is)
  const importRegex = /^import\s+.*?;?\s*\n/gm;
  let importMatch;
  while ((importMatch = importRegex.exec(remaining)) !== null) {
    segments.push({
      type: 'import',
      content: importMatch[0],
      translatable: false,
      start: position + importMatch.index,
      end: position + importMatch.index + importMatch[0].length,
    });
  }

  // Remove imports from remaining content
  remaining = remaining.replace(importRegex, '');

  // Extract code blocks (preserve as-is)
  const codeBlockRegex = /```(\w*)\s*\n([\s\S]*?)```/g;
  let codeMatch;
  let codeOffset = 0;

  while ((codeMatch = codeBlockRegex.exec(remaining)) !== null) {
    const beforeCode = remaining.slice(codeOffset, codeMatch.index);

    if (beforeCode.trim()) {
      segments.push({
        type: 'prose',
        content: beforeCode,
        translatable: true,
        start: position + codeOffset,
        end: position + codeMatch.index,
      });
    }

    segments.push({
      type: 'code',
      language: codeMatch[1] || '',
      content: codeMatch[0],
      translatable: false,
      start: position + codeMatch.index,
      end: position + codeMatch.index + codeMatch[0].length,
    });

    codeOffset = codeMatch.index + codeMatch[0].length;
  }

  // Add remaining prose after last code block
  const afterCode = remaining.slice(codeOffset);
  if (afterCode.trim()) {
    segments.push({
      type: 'prose',
      content: afterCode,
      translatable: true,
      start: position + codeOffset,
      end: position + codeOffset + afterCode.length,
    });
  }

  return segments;
}

/**
 * Translate text content
 * This is a placeholder - replace with actual translation service
 */
async function translateText(text, sourceLocale, targetLocale, apiKey = null) {
  // Preserve technical terms
  let translated = text;

  for (const term of CONFIG.preserveTerms) {
    // Create a regex that matches the term with word boundaries
    const regex = new RegExp(`\\b${term}\\b`, 'g');
    // We'll use placeholders and restore them after translation
    translated = translated.replace(regex, `__PRESERVE_${term}__`);
  }

  // Placeholder translation logic
  // In production, replace this with actual translation API calls
  // Options: Google Translate, DeepL, Azure Translator, AWS Translate

  const translations = {
    ko: {
      en: {
        '소개': 'Introduction',
        '핵심 개념': 'Core Concepts',
        '시작하기': 'Getting Started',
        '고급 기능': 'Advanced',
        '도메인 주도 개발': 'Domain-Driven Development',
        'TRUST 5 품질': 'TRUST 5 Quality',
        'SPEC 기반 개발': 'SPEC-Based Development',
        'MoAI-ADK란?': 'What is MoAI-ADK?',
        // Add more translations as needed
      },
      ja: {
        '소개': 'イントロダクション',
        '핵심 개념': 'コアコンセプト',
        '시작하기': 'はじめに',
        '고급 기능': '高度な機能',
        '도메인 주도 개발': 'ドメイン駆動開発',
      },
      zh: {
        '소개': '介绍',
        '핵심 개념': '核心概念',
        '시작하기': '入门指南',
        '고급 기능': '高级功能',
        '도메인 주도 개발': '领域驱动开发',
      },
    },
  };

  // Simple word-by-word lookup for demonstration
  // Replace with actual API call in production
  if (sourceLocale === 'ko' && translations.ko[targetLocale]) {
    const dict = translations.ko[targetLocale];
    for (const [ko, target] of Object.entries(dict)) {
      translated = translated.replace(new RegExp(ko, 'g'), target);
    }
  }

  // Restore preserved terms
  for (const term of CONFIG.preserveTerms) {
    translated = translated.replace(new RegExp(`__PRESERVE_${term}__`, 'g'), term);
  }

  return translated;
}

/**
 * Translate frontmatter fields
 */
async function translateFrontmatter(frontmatterContent, sourceLocale, targetLocale, apiKey) {
  const lines = frontmatterContent.split('\n');
  const translated = [];

  for (const line of lines) {
    if (line.trim().startsWith('title:')) {
      const title = line.replace(/^title:\s*/, '').replace(/^["']|["']$/g, '');
      const translatedTitle = await translateText(title, sourceLocale, targetLocale, apiKey);
      translated.push(`title: "${translatedTitle}"`);
    } else if (line.trim().startsWith('description:')) {
      const desc = line.replace(/^description:\s*/, '').replace(/^["']|["']$/g, '');
      const translatedDesc = await translateText(desc, sourceLocale, targetLocale, apiKey);
      translated.push(`description: "${translatedDesc}"`);
    } else {
      translated.push(line);
    }
  }

  return translated.join('\n');
}

/**
 * Generate _meta.ts content for a locale
 */
async function generateMetaFile(sourceMetaPath, sourceLocale, targetLocale) {
  if (!existsSync(sourceMetaPath)) {
    return null;
  }

  const content = readFileSync(sourceMetaPath, 'utf-8');
  const lines = content.split('\n');
  const translated = [];

  // Translate comment
  for (const line of lines) {
    if (line.trim().startsWith('*') || line.trim().startsWith('/**')) {
      translated.push(await translateText(line, sourceLocale, targetLocale));
    } else if (line.includes(':')) {
      const [key, ...valueParts] = line.split(':');
      const value = valueParts.join(':').trim().replace(/^["']|["']$/g, '');
      if (value && value !== 'hidden') {
        const translatedValue = await translateText(value, sourceLocale, targetLocale);
        translated.push(`  ${key.trim()}: "${translatedValue}",`);
      } else {
        translated.push(line);
      }
    } else {
      translated.push(line);
    }
  }

  return translated.join('\n');
}

/**
 * Translate a single MDX file
 */
async function translateFile(sourcePath, targetPath, sourceLocale, targetLocale, options) {
  console.log(`Translating: ${sourcePath} -> ${targetPath}`);

  const content = readFileSync(sourcePath, 'utf-8');
  const segments = parseMdxContent(content);

  let translatedContent = '';

  for (const segment of segments) {
    if (!segment.translatable) {
      translatedContent += segment.content;
    } else if (segment.type === 'frontmatter') {
      const translated = await translateFrontmatter(
        segment.content,
        sourceLocale,
        targetLocale,
        options.apiKey
      );
      translatedContent += translated;
    } else {
      const translated = await translateText(
        segment.content,
        sourceLocale,
        targetLocale,
        options.apiKey
      );
      translatedContent += translated;
    }
  }

  if (!options.dryRun) {
    const targetDir = dirname(targetPath);
    if (!existsSync(targetDir)) {
      mkdirSync(targetDir, { recursive: true });
    }
    writeFileSync(targetPath, translatedContent, 'utf-8');
  }

  return { sourcePath, targetPath, segments: segments.length };
}

/**
 * Find and translate _meta.ts files
 */
async function translateMetaFiles(sourceDir, sourceLocale, targetLocale, options) {
  const metaFiles = [];

  function findMetaFiles(dir) {
    const entries = readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = join(dir, entry.name);
      if (entry.isDirectory() && !CONFIG.skipDirs.includes(entry.name)) {
        findMetaFiles(fullPath);
      } else if (entry.name === '_meta.ts') {
        metaFiles.push(fullPath);
      }
    }
  }

  findMetaFiles(sourceDir);

  const results = [];
  for (const metaPath of metaFiles) {
    let relativePath = relative(sourceDir, metaPath);
    // Strip source locale from path (e.g., 'ko/' from 'ko/advanced/_meta.ts')
    for (const [localeCode, localeData] of Object.entries(CONFIG.locales)) {
      const prefix = localeData.dir + '/';
      if (relativePath.startsWith(prefix)) {
        relativePath = relativePath.slice(prefix.length);
        break;
      }
    }
    const targetPath = join(process.cwd(), options.sourceDir, CONFIG.locales[targetLocale].dir, relativePath);

    console.log(`Processing meta: ${metaPath} -> ${targetPath}`);

    const translated = generateMetaFile(metaPath, sourceLocale, targetLocale);
    if (translated && !options.dryRun) {
      const targetDir = dirname(targetPath);
      if (!existsSync(targetDir)) {
        mkdirSync(targetDir, { recursive: true });
      }
      writeFileSync(targetPath, translated, 'utf-8');
      results.push({ source: metaPath, target: targetPath });
    }
  }

  return results;
}

/**
 * Main translation function
 */
async function main() {
  const options = parseArgs();
  const sourceDir = join(process.cwd(), options.sourceDir);

  console.log('='.repeat(60));
  console.log('MoAI-ADK MDX Translation Script');
  console.log('='.repeat(60));
  console.log(`Source directory: ${sourceDir}`);
  console.log(`Target locale(s): ${options.targetLocale || 'all (en, ja, zh)'}`);
  console.log(`Dry run: ${options.dryRun}`);
  console.log('='.repeat(60));

  // Find all MDX files
  console.log('\nScanning for MDX files...');
  const mdxFiles = findMdxFiles(sourceDir);
  console.log(`Found ${mdxFiles.length} MDX files`);

  // Determine target locales
  const targetLocales = options.targetLocale
    ? [options.targetLocale]
    : ['en', 'ja', 'zh'];

  // Process each locale
  for (const targetLocale of targetLocales) {
    console.log(`\nProcessing locale: ${targetLocale} (${CONFIG.locales[targetLocale].name})`);
    console.log('-'.repeat(60));

    const results = {
      mdx: [],
      meta: [],
    };

    // Translate MDX files
    for (const sourcePath of mdxFiles) {
      let relativePath = relative(sourceDir, sourcePath);
      // Strip source locale from path (e.g., 'ko/' from 'ko/advanced/...')
      for (const [localeCode, localeData] of Object.entries(CONFIG.locales)) {
        const prefix = localeData.dir + '/';
        if (relativePath.startsWith(prefix)) {
          relativePath = relativePath.slice(prefix.length);
          break;
        }
      }
      const targetPath = join(
        process.cwd(),
        options.sourceDir,
        CONFIG.locales[targetLocale].dir,
        relativePath
      );

      const result = await translateFile(
        sourcePath,
        targetPath,
        'ko',
        targetLocale,
        options
      );
      results.mdx.push(result);
    }

    // Translate _meta.ts files
    const metaResults = await translateMetaFiles(
      sourceDir,
      'ko',
      targetLocale,
      options
    );
    results.meta = metaResults;

    console.log(`\nResults for ${targetLocale}:`);
    console.log(`  MDX files processed: ${results.mdx.length}`);
    console.log(`  Meta files processed: ${results.meta.length}`);
  }

  console.log('\n' + '='.repeat(60));
  console.log('Translation complete!');
  console.log('='.repeat(60));

  if (options.dryRun) {
    console.log('\nThis was a dry run. No files were written.');
    console.log('Run without --dry-run to perform actual translation.');
  } else {
    console.log('\nTranslated files have been written to:');
    for (const locale of targetLocales) {
      console.log(`  ${join(options.sourceDir, CONFIG.locales[locale].dir)}`);
    }
  }
}

// Run the script
main().catch(console.error);
