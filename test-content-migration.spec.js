// @TEST:TAG-MIGRATION-003
// Test suite for Content Migration functionality

import fs from 'fs/promises';
import path from 'path';

// Configuration
const DOCS_DIR = '/Users/goos/MoAI/MoAI-ADK/docs';
const PAGES_BAK_DIR = '/Users/goos/MoAI/MoAI-ADK/docs/pages.bak';
const CONTENT_DIR = '/Users/goos/MoAI/MoAI-ADK/docs/content';

/**
 * Test suite for content migration from pages.bak to content/
 */
export async function testContentMigration() {
  console.log('🧪 Testing Content Migration (TAG-MIGRATION-003)...\n');

  // Test 1: content/ directory structure exists
  console.log('📁 Test 1: Testing content/ directory structure...');
  try {
    await fs.access(CONTENT_DIR);
    console.log('✅ content/ directory exists');
  } catch (error) {
    console.log('❌ content/ directory does not exist');
    throw new Error('content/ directory must exist after migration');
  }

  // Test 2: Count files migrated correctly
  console.log('\n📊 Test 2: Testing all files migrated correctly...');
  const sourceFiles = await countFiles(PAGES_BAK_DIR, ['.md', '.mdx']);
  const targetFiles = await countFiles(CONTENT_DIR, ['.md', '.mdx']);

  if (sourceFiles !== targetFiles) {
    throw new Error(`File count mismatch: source=${sourceFiles}, target=${targetFiles}`);
  }
  console.log(`✅ All ${sourceFiles} files migrated successfully`);

  // Test 3: Language-specific directories preserved
  console.log('\n🌐 Test 3: Testing language-specific directories...');
  const expectedLangDirs = ['ko', 'en', 'ja', 'zh'];
  for (const lang of expectedLangDirs) {
    const langDir = path.join(CONTENT_DIR, lang);
    try {
      await fs.access(langDir);
      const langFiles = await countFiles(langDir, ['.md', '.mdx']);
      console.log(`✅ ${lang}/ directory exists with ${langFiles} files`);
    } catch (error) {
      throw new Error(`${lang}/ directory missing or empty`);
    }
  }

  // Test 4: _meta.json → meta.json conversion
  console.log('\n🔄 Test 4: Testing _meta.json → meta.json conversion...');
  const metaFiles = await findFiles(CONTENT_DIR, 'meta.json');
  const oldMetaFiles = await findFiles(CONTENT_DIR, '_meta.json');

  if (oldMetaFiles.length > 0) {
    throw new Error(`Found ${oldMetaFiles.length} remaining _meta.json files`);
  }
  console.log(`✅ All ${metaFiles.length} _meta.json files renamed to meta.json`);

  // Test 5: Internal links validation
  console.log('\n🔗 Test 5: Testing internal links...');
  const markdownFiles = await findFiles(CONTENT_DIR, ['.md', '.mdx']);
  let brokenLinks = 0;

  for (const file of markdownFiles) {
    const content = await fs.readFile(file, 'utf-8');
    const links = extractInternalLinks(content);

    for (const link of links) {
      if (!await validateInternalLink(CONTENT_DIR, link)) {
        console.log(`❌ Broken link found: ${file} → ${link}`);
        brokenLinks++;
      }
    }
  }

  if (brokenLinks > 0) {
    throw new Error(`Found ${brokenLinks} broken internal links`);
  }
  console.log('✅ All internal links are valid');

  // Test 6: Theme config update
  console.log('\n⚙️ Test 6: Testing theme.config.tsx update...');
  const themeConfigPath = path.join(DOCS_DIR, 'theme.config.tsx');
  const themeContent = await fs.readFile(themeConfigPath, 'utf-8');

  if (!themeContent.includes('content:')) {
    throw new Error('theme.config.tsx not updated to reference content/ directory');
  }
  console.log('✅ theme.config.tsx updated to reference content/ directory');

  console.log('\n🎉 All content migration tests passed!');
  return {
    totalFiles: sourceFiles,
    languageDirs: expectedLangDirs,
    metaFiles: metaFiles.length,
    brokenLinks: 0
  };
}

// Helper functions
async function countFiles(dir, extensions) {
  try {
    const files = await findFiles(dir, extensions);
    return files.length;
  } catch {
    return 0;
  }
}

async function findFiles(dir, extensions) {
  const files = [];
  await scanDirectory(dir, extensions, files);
  return files;
}

async function scanDirectory(dir, extensions, files) {
  const items = await fs.readdir(dir, { withFileTypes: true });

  for (const item of items) {
    const fullPath = path.join(dir, item.name);

    if (item.isDirectory()) {
      await scanDirectory(fullPath, extensions, files);
    } else if (item.isFile()) {
      const ext = path.extname(item.name);
      if (Array.isArray(extensions) ? extensions.includes(ext) : extensions === ext) {
        files.push(fullPath);
      }
    }
  }
}

function extractInternalLinks(content) {
  const linkPattern = /\[.*?\]\((.*?\.md(#[a-zA-Z0-9-_.]+)?|.*?\.mdx(#[a-zA-Z0-9-_.]+)?)\)/g;
  const links = [];
  let match;

  while ((match = linkPattern.exec(content)) !== null) {
    links.push(match[1]);
  }

  return links;
}

async function validateInternalLink(contentDir, link) {
  // Normalize link path
  let cleanLink = link;
  if (cleanLink.startsWith('./')) {
    cleanLink = cleanLink.substring(2);
  }

  const fullPath = path.join(contentDir, cleanLink);

  try {
    await fs.access(fullPath);
    return true;
  } catch {
    return false;
  }
}

// Run tests if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  testContentMigration().catch(console.error);
}