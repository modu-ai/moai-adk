#!/usr/bin/env node

/**
 * Pagefind Build Script for MoAI-ADK Documentation
 *
 * This script builds Pagefind indexes for all supported languages:
 * - Korean (ko)
 * - English (en)
 * - Japanese (ja)
 * - Chinese (zh)
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const LANGUAGES = ['ko', 'en', 'ja', 'zh'];
const SITE_DIR = 'out';
const BASE_CONFIG = {
  rootSelector: '[data-pagefind-body]',
  excludeSelectors: [
    'nav',
    'header',
    'footer',
    '.sidebar',
    '.navigation',
    '[data-pagefind-ignore]',
    'script',
    'style',
    'noscript',
    '.pagefind-ui',
    '.search-container'
  ],
  includeCharacters: '-_.,:;!?()[]{}\"\'',
  verbose: false,
  silent: true,
  keepIndexUrl: false,
  glob: '**/*.html'
};

/**
 * Log with timestamp
 */
function log(message, type = 'info') {
  const timestamp = new Date().toISOString();
  const prefix = {
    info: 'ℹ️',
    error: '❌',
    success: '✅',
    warning: '⚠️'
  }[type] || 'ℹ️';

  console.log(`[${timestamp}] ${prefix} ${message}`);
}

/**
 * Verify that the site directory exists
 */
function verifySiteDirectory() {
  if (!fs.existsSync(SITE_DIR)) {
    throw new Error(`Site directory '${SITE_DIR}' does not exist. Run 'next build' first.`);
  }

  log(`Verified site directory: ${SITE_DIR}`);
}

/**
 * Build Pagefind index for a specific language
 */
async function buildLanguageIndex(locale) {
  log(`Building Pagefind index for locale: ${locale}`);

  const outputPath = path.join(SITE_DIR, 'pagefind', locale);
  const config = {
    ...BASE_CONFIG,
    site: SITE_DIR,
    outputSubdir: `pagefind/${locale}`,
    forceLanguage: locale
  };

  // Create locale-specific pagefind.yml
  const localeConfig = {
    site: SITE_DIR,
    output_subdir: `pagefind/${locale}`,
    root_selector: config.rootSelector,
    exclude_selectors: config.excludeSelectors,
    include_characters: config.includeCharacters,
    force_language: locale,
    verbose: config.verbose,
    silent: config.silent,
    keep_index_url: config.keepIndexUrl,
    glob: config.glob
  };

  const configPath = path.join(process.cwd(), `pagefind-${locale}.yml`);
  fs.writeFileSync(configPath, Object.entries(localeConfig)
    .map(([key, value]) => {
      if (Array.isArray(value)) {
        return `${key}:\n${value.map(item => `  - "${item}"`).join('\n')}`;
      }
      return `${key}: ${value}`;
    })
    .join('\n')
  );

  try {
    // Build the index
    execSync(`npx pagefind --source ${SITE_DIR} --output-subdir pagefind/${locale} --force-language ${locale}`, {
      stdio: 'inherit',
      cwd: process.cwd()
    });

    log(`Successfully built index for ${locale}`, 'success');

    // Verify the index was created
    const indexPath = path.join(outputPath, 'pagefind.js');
    if (fs.existsSync(indexPath)) {
      const stats = fs.statSync(indexPath);
      log(`Index size for ${locale}: ${(stats.size / 1024).toFixed(2)} KB`);
    } else {
      throw new Error(`Index file not found at ${indexPath}`);
    }

  } catch (error) {
    log(`Failed to build index for ${locale}: ${error.message}`, 'error');
    throw error;
  } finally {
    // Clean up temporary config file
    if (fs.existsSync(configPath)) {
      fs.unlinkSync(configPath);
    }
  }
}

/**
 * Create a unified search index configuration
 */
function createMultiIndexConfig() {
  const multiIndexPath = path.join(SITE_DIR, 'pagefind', 'multi-index.json');
  const multiIndex = {
    version: '1.0.0',
    languages: LANGUAGES,
    default: 'ko',
    indexes: LANGUAGES.reduce((acc, locale) => {
      acc[locale] = {
        path: `/pagefind/${locale}/`,
        language: locale,
        name: {
          ko: '한국어',
          en: 'English',
          ja: '日本語',
          zh: '中文'
        }[locale]
      };
      return acc;
    }, {})
  };

  fs.mkdirSync(path.dirname(multiIndexPath), { recursive: true });
  fs.writeFileSync(multiIndexPath, JSON.stringify(multiIndex, null, 2));

  log(`Created multi-index configuration at ${multiIndexPath}`);
}

/**
 * Generate search statistics
 */
function generateStats() {
  const statsPath = path.join(SITE_DIR, 'pagefind', 'stats.json');
  const stats = {
    generated: new Date().toISOString(),
    languages: LANGUAGES,
    totalFiles: 0,
    totalSize: 0
  };

  for (const locale of LANGUAGES) {
    const localePath = path.join(SITE_DIR, 'pagefind', locale);
    if (fs.existsSync(localePath)) {
      const files = fs.readdirSync(localePath);
      stats.totalFiles += files.length;

      for (const file of files) {
        const filePath = path.join(localePath, file);
        if (fs.statSync(filePath).isFile()) {
          stats.totalSize += fs.statSync(filePath).size;
        }
      }
    }
  }

  fs.writeFileSync(statsPath, JSON.stringify(stats, null, 2));
  log(`Generated search statistics: ${stats.totalFiles} files, ${(stats.totalSize / 1024).toFixed(2)} KB total`);
}

/**
 * Main build function
 */
async function buildAllIndexes() {
  const startTime = Date.now();

  try {
    log('Starting Pagefind build process for MoAI-ADK documentation');

    verifySiteDirectory();

    for (const locale of LANGUAGES) {
      await buildLanguageIndex(locale);
    }

    createMultiIndexConfig();
    generateStats();

    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    log(`Pagefind build completed successfully in ${duration}s`, 'success');

  } catch (error) {
    log(`Pagefind build failed: ${error.message}`, 'error');
    process.exit(1);
  }
}

// Run the build if called directly
if (require.main === module) {
  buildAllIndexes();
}

module.exports = { buildAllIndexes, buildLanguageIndex };