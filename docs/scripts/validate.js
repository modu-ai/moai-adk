#!/usr/bin/env node

/**
 * Build & Deployment Validation Script
 * Validates Nextra build output and multilingual structure
 */

const fs = require('fs');
const path = require('path');

const GREEN = '\x1b[32m';
const RED = '\x1b[31m';
const YELLOW = '\x1b[33m';
const RESET = '\x1b[0m';

let errors = 0;
let warnings = 0;

function log(type, message) {
  const symbols = {
    success: `${GREEN}✅${RESET}`,
    error: `${RED}❌${RESET}`,
    warning: `${YELLOW}⚠️${RESET}`,
  };

  console.log(`${symbols[type]} ${message}`);
}

function validateBuildOutput() {
  console.log('\n🔍 Validating Build Output...\n');

  // Check .next directory
  if (!fs.existsSync('.next')) {
    log('error', '.next directory not found');
    errors++;
    return false;
  }
  log('success', '.next directory exists');

  // Check .next/static directory
  if (!fs.existsSync('.next/static')) {
    log('error', '.next/static directory not found');
    errors++;
    return false;
  }
  log('success', '.next/static directory exists');

  // Count static files
  const staticFiles = countFiles('.next/static');
  log('success', `.next/static contains ${staticFiles} files`);

  return true;
}

function validateConfigurations() {
  console.log('\n⚙️ Validating Configuration Files...\n');

  const requiredFiles = [
    'next.config.cjs',
    'theme.config.tsx',
    'tsconfig.json',
    'package.json',
    'postcss.config.js',
    '.eslintrc.json',
    'vercel.json',
  ];

  let allExist = true;

  requiredFiles.forEach((file) => {
    if (fs.existsSync(file)) {
      log('success', `${file} present`);
    } else {
      log('error', `${file} missing`);
      errors++;
      allExist = false;
    }
  });

  return allExist;
}

function validateMultilingualStructure() {
  console.log('\n🌍 Validating Multilingual Structure...\n');

  const locales = ['ko', 'en', 'ja', 'zh'];
  let allValid = true;

  locales.forEach((locale) => {
    const localePath = path.join('pages', locale);

    if (!fs.existsSync(localePath)) {
      log('warning', `${locale} directory not found`);
      warnings++;
      return;
    }

    const mdFiles = countMarkdownFiles(localePath);
    log('success', `${locale}: ${mdFiles} markdown files`);

    // Check for _meta.json
    const metaPath = path.join(localePath, '_meta.json');
    if (fs.existsSync(metaPath)) {
      log('success', `${locale} has _meta.json`);
    } else {
      log('warning', `${locale} missing _meta.json`);
      warnings++;
    }
  });

  return allValid;
}

function validatePages() {
  console.log('\n📄 Validating Page Structure...\n');

  if (!fs.existsSync('pages')) {
    log('error', 'pages directory not found');
    errors++;
    return false;
  }

  const totalMdFiles = countMarkdownFiles('pages');
  log('success', `Total markdown files: ${totalMdFiles}`);

  // Check for required directories
  const requiredDirs = ['pages/ko', 'pages/en'];

  requiredDirs.forEach((dir) => {
    if (fs.existsSync(dir)) {
      log('success', `${dir} exists`);
    } else {
      log('error', `${dir} missing`);
      errors++;
    }
  });

  return true;
}

function validateSecurityHeaders() {
  console.log('\n🔒 Validating Security Configuration...\n');

  const vercelJsonPath = 'vercel.json';

  if (!fs.existsSync(vercelJsonPath)) {
    log('error', 'vercel.json not found');
    errors++;
    return false;
  }

  try {
    const vercelConfig = JSON.parse(fs.readFileSync(vercelJsonPath, 'utf-8'));

    const requiredHeaders = [
      'X-Content-Type-Options',
      'X-Frame-Options',
      'X-XSS-Protection',
      'Referrer-Policy',
    ];

    const headers = vercelConfig.headers?.[0]?.headers || [];
    const headerNames = headers.map((h) => h.key);

    requiredHeaders.forEach((header) => {
      if (headerNames.includes(header)) {
        log('success', `Security header ${header} configured`);
      } else {
        log('warning', `Security header ${header} not configured`);
        warnings++;
      }
    });

    return true;
  } catch (error) {
    log('error', `Failed to parse vercel.json: ${error.message}`);
    errors++;
    return false;
  }
}

function validateI18nConfig() {
  console.log('\n🌐 Validating i18n Configuration...\n');

  try {
    const nextConfigPath = 'next.config.cjs';

    if (!fs.existsSync(nextConfigPath)) {
      log('error', 'next.config.cjs not found');
      errors++;
      return false;
    }

    const configContent = fs.readFileSync(nextConfigPath, 'utf-8');

    const expectedLocales = ['ko', 'en', 'ja', 'zh'];
    let allLocalesFound = true;

    expectedLocales.forEach((locale) => {
      if (configContent.includes(`'${locale}'`)) {
        log('success', `Locale '${locale}' configured in next.config.cjs`);
      } else {
        log('warning', `Locale '${locale}' not found in next.config.cjs`);
        warnings++;
        allLocalesFound = false;
      }
    });

    return allLocalesFound;
  } catch (error) {
    log('error', `Failed to validate i18n config: ${error.message}`);
    errors++;
    return false;
  }
}

function countFiles(dir) {
  if (!fs.existsSync(dir)) return 0;

  let count = 0;

  function traverse(currentPath) {
    const files = fs.readdirSync(currentPath);

    files.forEach((file) => {
      const fullPath = path.join(currentPath, file);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        traverse(fullPath);
      } else {
        count++;
      }
    });
  }

  traverse(dir);
  return count;
}

function countMarkdownFiles(dir) {
  if (!fs.existsSync(dir)) return 0;

  let count = 0;

  function traverse(currentPath) {
    const files = fs.readdirSync(currentPath);

    files.forEach((file) => {
      const fullPath = path.join(currentPath, file);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        traverse(fullPath);
      } else if (file.endsWith('.md') || file.endsWith('.mdx')) {
        count++;
      }
    });
  }

  traverse(dir);
  return count;
}

// Main execution
async function main() {
  console.log('═'.repeat(60));
  console.log('  🚀 Nextra Build & Deployment Validation');
  console.log('═'.repeat(60));

  validateBuildOutput();
  validateConfigurations();
  validateMultilingualStructure();
  validatePages();
  validateSecurityHeaders();
  validateI18nConfig();

  console.log('\n' + '═'.repeat(60));
  console.log('  📋 Validation Summary');
  console.log('═'.repeat(60));

  const summary = `${GREEN}✅ Passed${RESET} | ${errors > 0 ? `${RED}❌ Errors: ${errors}${RESET}` : `${GREEN}❌ Errors: 0${RESET}`} | ${warnings > 0 ? `${YELLOW}⚠️ Warnings: ${warnings}${RESET}` : `${GREEN}⚠️ Warnings: 0${RESET}` }`;

  console.log(`\nResults: ${summary}\n`);

  if (errors > 0) {
    console.log(`${RED}Validation failed with ${errors} error(s)${RESET}`);
    process.exit(1);
  }

  if (warnings > 0) {
    console.log(`${YELLOW}Validation completed with ${warnings} warning(s)${RESET}`);
  } else {
    console.log(`${GREEN}All validations passed!${RESET}`);
  }
}

main().catch((error) => {
  console.error(`${RED}Validation error: ${error.message}${RESET}`);
  process.exit(1);
});
