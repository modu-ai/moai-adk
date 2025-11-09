#!/usr/bin/env node

/**
 * Build Performance & Bundle Analysis Script
 * Analyzes Nextra build output size, structure, and performance metrics
 */

const fs = require('fs');
const path = require('path');

const GREEN = '\x1b[32m';
const RED = '\x1b[31m';
const YELLOW = '\x1b[33m';
const BLUE = '\x1b[34m';
const RESET = '\x1b[0m';

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}

function getDirectorySize(dir) {
  if (!fs.existsSync(dir)) return 0;

  let size = 0;

  function traverse(currentPath) {
    const files = fs.readdirSync(currentPath);

    files.forEach((file) => {
      const fullPath = path.join(currentPath, file);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        traverse(fullPath);
      } else {
        size += stat.size;
      }
    });
  }

  traverse(dir);
  return size;
}

function analyzeNextBuild() {
  console.log('\n📦 Next.js Build Analysis\n');

  if (!fs.existsSync('.next')) {
    console.log(`${RED}❌ .next directory not found${RESET}`);
    return {};
  }

  const nextSize = getDirectorySize('.next');
  const nextStatic = getDirectorySize('.next/static');
  const nextServer = getDirectorySize('.next/server');

  console.log(`${BLUE}Total .next size:${RESET} ${formatBytes(nextSize)}`);
  console.log(`  ├─ Static files: ${formatBytes(nextStatic)}`);
  console.log(`  └─ Server files: ${formatBytes(nextServer)}`);

  // Analyze file types
  console.log('\n📊 File Type Distribution:\n');

  const fileTypes = {};

  function countFileTypes(dir) {
    const files = fs.readdirSync(dir);

    files.forEach((file) => {
      const fullPath = path.join(dir, file);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        countFileTypes(fullPath);
      } else {
        const ext = path.extname(file) || 'no-extension';
        const stat = fs.statSync(fullPath);

        if (!fileTypes[ext]) {
          fileTypes[ext] = { count: 0, size: 0 };
        }

        fileTypes[ext].count++;
        fileTypes[ext].size += stat.size;
      }
    });
  }

  countFileTypes('.next');

  // Sort by size
  const sorted = Object.entries(fileTypes).sort(([, a], [, b]) => b.size - a.size);

  sorted.slice(0, 10).forEach(([ext, data]) => {
    const percentage = ((data.size / nextSize) * 100).toFixed(1);
    console.log(`  ${ext.padEnd(12)} ${data.count.toString().padEnd(6)} files ${formatBytes(data.size).padEnd(12)} (${percentage}%)`);
  });

  return { nextSize, nextStatic, nextServer };
}

function analyzeSourceFiles() {
  console.log('\n📄 Source Files Analysis\n');

  if (!fs.existsSync('pages')) {
    console.log(`${RED}❌ pages directory not found${RESET}`);
    return {};
  }

  const pagesSize = getDirectorySize('pages');
  const locales = ['ko', 'en', 'ja', 'zh'];

  console.log(`${BLUE}Total pages size:${RESET} ${formatBytes(pagesSize)}\n`);
  console.log('Breakdown by locale:\n');

  const localeStats = {};

  locales.forEach((locale) => {
    const localePath = path.join('pages', locale);

    if (fs.existsSync(localePath)) {
      let fileCount = 0;
      let totalSize = 0;

      function countFiles(dir) {
        const files = fs.readdirSync(dir);

        files.forEach((file) => {
          const fullPath = path.join(dir, file);
          const stat = fs.statSync(fullPath);

          if (stat.isDirectory()) {
            countFiles(fullPath);
          } else if (file.endsWith('.md') || file.endsWith('.mdx')) {
            fileCount++;
            totalSize += stat.size;
          }
        });
      }

      countFiles(localePath);

      localeStats[locale] = { fileCount, totalSize };
      console.log(`  ${locale.toUpperCase()} ${fileCount.toString().padEnd(3)} files ${formatBytes(totalSize).padEnd(12)}`);
    }
  });

  return { pagesSize, localeStats };
}

function analyzeStyles() {
  console.log('\n🎨 Style Configuration Analysis\n');

  const configFiles = [
    { path: 'styles/globals.css', name: 'Global Styles' },
    { path: 'tailwind.config.ts', name: 'Tailwind Config' },
    { path: 'postcss.config.js', name: 'PostCSS Config' },
  ];

  configFiles.forEach(({ path: filePath, name }) => {
    if (fs.existsSync(filePath)) {
      const stat = fs.statSync(filePath);
      const lines = fs.readFileSync(filePath, 'utf-8').split('\n').length;

      console.log(`  ${name.padEnd(20)} ${formatBytes(stat.size).padEnd(12)} (${lines} lines)`);
    } else {
      console.log(`  ${name.padEnd(20)} ${RED}NOT FOUND${RESET}`);
    }
  });
}

function generateReport() {
  console.log('\n' + '═'.repeat(70));
  console.log('  📊 Build Performance & Bundle Analysis Report');
  console.log('═'.repeat(70));

  const buildStats = analyzeNextBuild();
  const sourceStats = analyzeSourceFiles();
  analyzeStyles();

  // Summary metrics
  console.log('\n' + '═'.repeat(70));
  console.log('  📈 Summary Metrics');
  console.log('═'.repeat(70) + '\n');

  if (buildStats.nextSize) {
    const ratio = (buildStats.nextSize / sourceStats.pagesSize).toFixed(2);

    console.log(`${BLUE}Build Efficiency:${RESET}`);
    console.log(`  Source Size (pages/):   ${formatBytes(sourceStats.pagesSize)}`);
    console.log(`  Build Size (.next/):    ${formatBytes(buildStats.nextSize)}`);
    console.log(`  Compression Ratio:      ${ratio}x`);

    // Performance assessment
    console.log('\n${BLUE}Performance Assessment:${RESET}');

    if (buildStats.nextSize < 10 * 1024 * 1024) {
      console.log(`  Bundle Size: ${GREEN}✅ EXCELLENT${RESET} (< 10 MB)`);
    } else if (buildStats.nextSize < 20 * 1024 * 1024) {
      console.log(`  Bundle Size: ${GREEN}✅ GOOD${RESET} (< 20 MB)`);
    } else if (buildStats.nextSize < 50 * 1024 * 1024) {
      console.log(`  Bundle Size: ${YELLOW}⚠️ WARNING${RESET} (< 50 MB)`);
    } else {
      console.log(`  Bundle Size: ${RED}❌ CRITICAL${RESET} (> 50 MB)`);
    }
  }

  console.log(`\n${BLUE}Internationalization:${RESET}`);
  console.log(`  Supported Languages: ${Object.keys(sourceStats.localeStats).length}`);

  Object.entries(sourceStats.localeStats).forEach(([locale, stats]) => {
    console.log(`    • ${locale.toUpperCase()}: ${stats.fileCount} files (${formatBytes(stats.totalSize)})`);
  });

  // Recommendations
  console.log('\n' + '═'.repeat(70));
  console.log('  💡 Recommendations');
  console.log('═'.repeat(70) + '\n');

  if (buildStats.nextSize && buildStats.nextSize > 20 * 1024 * 1024) {
    console.log(`${YELLOW}⚠️ Bundle Size Optimization:${RESET}`);
    console.log('   • Consider enabling compression');
    console.log('   • Review large dependencies');
    console.log('   • Enable incremental static regeneration');
  } else {
    console.log(`${GREEN}✅ Bundle size is optimized${RESET}`);
  }

  if (!fs.existsSync('.next/static')) {
    console.log(`${RED}❌ Static optimization not enabled${RESET}`);
  } else {
    console.log(`${GREEN}✅ Static files properly optimized${RESET}`);
  }

  console.log(`\n${GREEN}✅ Analysis complete!${RESET}\n`);
}

// Main execution
try {
  generateReport();
} catch (error) {
  console.error(`${RED}Analysis error: ${error.message}${RESET}`);
  process.exit(1);
}
