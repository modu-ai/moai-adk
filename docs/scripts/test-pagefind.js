#!/usr/bin/env node

/**
 * Pagefind Integration Test Script
 *
 * Tests the Pagefind search functionality across all supported languages
 * and validates the build process.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const LANGUAGES = ['ko', 'en', 'ja', 'zh'];
const SITE_DIR = 'out';
const TEST_QUERIES = {
  ko: ['설치', '알프레드', 'TDD', '문서', '검색'],
  en: ['installation', 'alfred', 'TDD', 'documentation', 'search'],
  ja: ['インストール', 'アルフレッド', 'TDD', 'ドキュメント', '検索'],
  zh: ['安装', '阿尔弗雷德', 'TDD', '文档', '搜索']
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
 * Test if Pagefind files exist for a language
 */
function testPagefindFiles(locale) {
  const pagefindDir = path.join(SITE_DIR, 'pagefind', locale);
  const requiredFiles = [
    'pagefind.js',
    'pagefind-ui.js',
    'pagefind-ui.css'
  ];

  const missingFiles = [];
  for (const file of requiredFiles) {
    const filePath = path.join(pagefindDir, file);
    if (!fs.existsSync(filePath)) {
      missingFiles.push(file);
    }
  }

  if (missingFiles.length === 0) {
    log(`All Pagefind files present for ${locale}`, 'success');
    return true;
  } else {
    log(`Missing files for ${locale}: ${missingFiles.join(', ')}`, 'error');
    return false;
  }
}

/**
 * Test index size and content
 */
function testIndexQuality(locale) {
  const indexPath = path.join(SITE_DIR, 'pagefind', locale, 'pagefind.js');

  if (!fs.existsSync(indexPath)) {
    log(`Index file not found for ${locale}`, 'error');
    return false;
  }

  const stats = fs.statSync(indexPath);
  const sizeKB = (stats.size / 1024).toFixed(2);

  // Read index content to verify it contains search data
  const content = fs.readFileSync(indexPath, 'utf8');
  const hasData = content.includes('data') && content.includes('filters');

  if (hasData) {
    log(`Index quality check passed for ${locale} (${sizeKB} KB)`, 'success');
    return true;
  } else {
    log(`Index appears empty for ${locale} (${sizeKB} KB)`, 'warning');
    return false;
  }
}

/**
 * Test that HTML files contain the proper data-pagefind-body attribute
 */
function testHtmlAttributes() {
  const htmlDir = path.join(SITE_DIR);
  let filesChecked = 0;
  let filesWithAttribute = 0;

  for (const locale of LANGUAGES) {
    const localeDir = path.join(htmlDir, locale);
    if (!fs.existsSync(localeDir)) {
      log(`HTML directory not found for ${locale}`, 'warning');
      continue;
    }

    const htmlFiles = [];
    function scanDir(dir) {
      const items = fs.readdirSync(dir);
      for (const item of items) {
        const itemPath = path.join(dir, item);
        const stat = fs.statSync(itemPath);
        if (stat.isDirectory()) {
          scanDir(itemPath);
        } else if (item.endsWith('.html')) {
          htmlFiles.push(itemPath);
        }
      }
    }

    scanDir(localeDir);

    for (const htmlFile of htmlFiles) {
      const content = fs.readFileSync(htmlFile, 'utf8');
      filesChecked++;

      if (content.includes('data-pagefind-body')) {
        filesWithAttribute++;
      }
    }
  }

  const percentage = ((filesWithAttribute / filesChecked) * 100).toFixed(1);
  log(`HTML attribute test: ${filesWithAttribute}/${filesChecked} files have data-pagefind-body (${percentage}%)`,
      filesWithAttribute === filesChecked ? 'success' : 'warning');

  return filesWithAttribute === filesChecked;
}

/**
 * Test search functionality via API simulation
 */
function testSearchAPI(locale) {
  try {
    // This is a simplified test - in a real scenario, you might want to
    // spin up a server and test the actual search functionality
    const pagefindPath = path.join(SITE_DIR, 'pagefind', locale, 'pagefind.js');

    if (!fs.existsSync(pagefindPath)) {
      log(`Cannot test search API for ${locale} - index not found`, 'error');
      return false;
    }

    // Test that the index is loadable
    const indexContent = fs.readFileSync(pagefindPath, 'utf8');
    const isLoadable = indexContent.includes('version') && indexContent.includes('chunks');

    if (isLoadable) {
      log(`Search API test passed for ${locale}`, 'success');
      return true;
    } } else {
      log(`Search API test failed for ${locale}`, 'error');
      return false;
    }
  } catch (error) {
    log(`Search API test error for ${locale}: ${error.message}`, 'error');
    return false;
  }
}

/**
 * Generate test report
 */
function generateTestReport(results) {
  const reportPath = path.join(SITE_DIR, 'pagefind', 'test-report.json');
  const report = {
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    languages: LANGUAGES,
    results: results,
    summary: {
      totalTests: Object.values(results).flat().length,
      passedTests: Object.values(results).flat().filter(r => r.passed).length,
      failedTests: Object.values(results).flat().filter(r => !r.passed).length
    }
  };

  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
  log(`Test report generated at ${reportPath}`);
}

/**
 * Main test function
 */
async function runTests() {
  log('Starting Pagefind integration tests...');

  const results = {};
  let totalPassed = 0;
  let totalTests = 0;

  // Check if site is built
  if (!fs.existsSync(SITE_DIR)) {
    log(`Site directory '${SITE_DIR}' does not exist. Run 'bun run build' first.`, 'error');
    process.exit(1);
  }

  // Test 1: File presence
  log('\n📁 Testing file presence...');
  results.filePresence = {};
  for (const locale of LANGUAGES) {
    const passed = testPagefindFiles(locale);
    results.filePresence[locale] = { test: 'File Presence', passed };
    totalTests++;
    if (passed) totalPassed++;
  }

  // Test 2: Index quality
  log('\n📊 Testing index quality...');
  results.indexQuality = {};
  for (const locale of LANGUAGES) {
    const passed = testIndexQuality(locale);
    results.indexQuality[locale] = { test: 'Index Quality', passed };
    totalTests++;
    if (passed) totalPassed++;
  }

  // Test 3: HTML attributes
  log('\n🏷️  Testing HTML attributes...');
  const htmlTestPassed = testHtmlAttributes();
  results.htmlAttributes = { test: 'HTML Attributes', passed: htmlTestPassed };
  totalTests++;
  if (htmlTestPassed) totalPassed++;

  // Test 4: Search API
  log('\n🔍 Testing search API...');
  results.searchAPI = {};
  for (const locale of LANGUAGES) {
    const passed = testSearchAPI(locale);
    results.searchAPI[locale] = { test: 'Search API', passed };
    totalTests++;
    if (passed) totalPassed++;
  }

  // Generate report
  generateTestReport(results);

  // Summary
  log(`\n📋 Test Summary: ${totalPassed}/${totalTests} tests passed`);
  log(`Success Rate: ${((totalPassed / totalTests) * 100).toFixed(1)}%`);

  if (totalPassed === totalTests) {
    log('All tests passed! 🎉', 'success');
    process.exit(0);
  } else {
    log('Some tests failed. Check the logs above for details.', 'warning');
    process.exit(1);
  }
}

// Run tests if called directly
if (require.main === module) {
  runTests();
}

module.exports = { runTests, testPagefindFiles, testIndexQuality };