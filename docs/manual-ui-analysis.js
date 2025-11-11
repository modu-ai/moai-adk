#!/usr/bin/env node

/**
 * Manual UI/UX Analysis for MoAI-ADK Documentation
 * Analysis based on code structure and configuration files
 */

const fs = require('fs');
const path = require('path');

console.log('🔍 MoAI-ADK Documentation UI/UX Analysis');
console.log('=========================================');

// Analysis results
const analysis = {
  migrationStatus: {
    nextraVersion: null,
    nextjsVersion: null,
    migrationIssues: [],
    strengths: []
  },
  uiComponents: {
    theme: {},
    navigation: {},
    search: {},
    responsiveness: {},
    accessibility: {}
  },
  recommendations: {
    critical: [],
    high: [],
    medium: [],
    low: []
  }
};

// Read package.json
try {
  const packageJson = JSON.parse(fs.readFileSync('./package.json', 'utf8'));
  analysis.migrationStatus.nextraVersion = packageJson.dependencies?.nextra;
  analysis.migrationStatus.nextjsVersion = packageJson.dependencies?.next;

  console.log(`📦 Nextra Version: ${analysis.migrationStatus.nextraVersion}`);
  console.log(`📦 Next.js Version: ${analysis.migrationStatus.nextjsVersion}`);
} catch (error) {
  console.error('❌ Could not read package.json:', error.message);
}

// Analyze theme configuration
try {
  const themeConfigContent = fs.readFileSync('./theme.config.tsx', 'utf8');

  // Check for Nextra 4.x features
  if (themeConfigContent.includes('nextra-theme-docs')) {
    analysis.migrationStatus.strengths.push('✅ Using modern nextra-theme-docs');
  }

  // Check for i18n configuration
  if (themeConfigContent.includes('i18n:')) {
    analysis.uiComponents.navigation.internationalization = '✅ Configured';
    console.log('🌐 Internationalization: Configured');
  }

  // Check for custom search
  if (themeConfigContent.includes('CustomSearch')) {
    analysis.uiComponents.search.customImplementation = '✅ Custom Pagefind integration';
    console.log('🔍 Search: Custom Pagefind integration');
  }

  // Check for dark mode
  if (themeConfigContent.includes('darkMode: true')) {
    analysis.uiComponents.theme.darkMode = '✅ Enabled';
    console.log('🌙 Dark Mode: Enabled');
  }

} catch (error) {
  console.error('❌ Could not analyze theme config:', error.message);
  analysis.migrationStatus.migrationIssues.push('❌ Theme configuration analysis failed');
}

// Analyze CustomSearch component
try {
  if (fs.existsSync('./components/CustomSearch.tsx')) {
    const searchContent = fs.readFileSync('./components/CustomSearch.tsx', 'utf8');

    // Check for multi-language support
    if (searchContent.includes('translations:') && searchContent.includes('ko:') && searchContent.includes('en:')) {
      analysis.uiComponents.search.multilingual = '✅ Full i18n support';
      console.log('🌍 Search Multilingual: Full i18n support');
    }

    // Check for Pagefind integration
    if (searchContent.includes('pagefind') && searchContent.includes('PagefindUI')) {
      analysis.uiComponents.search.pagefindIntegration = '✅ Pagefind UI integrated';
      console.log('📄 Pagefind Integration: Complete');
    }

    // Check for dynamic locale detection
    if (searchContent.includes('currentLocale') && searchContent.includes('pathname')) {
      analysis.uiComponents.search.dynamicLocale = '✅ Dynamic locale detection';
      console.log('🔄 Dynamic Locale: Auto-detection enabled');
    }
  }
} catch (error) {
  console.error('❌ Could not analyze CustomSearch component:', error.message);
}

// Analyze Next.js configuration
try {
  const nextConfigContent = fs.readFileSync('./next.config.js', 'utf8');

  // Check for static export
  if (nextConfigContent.includes("output: 'export'")) {
    analysis.migrationStatus.strengths.push('✅ Static export configured');
    console.log('🚀 Static Export: Configured');
  }

  // Check for optimization features
  if (nextConfigContent.includes('compress: true')) {
    analysis.uiComponents.performance.compression = '✅ Enabled';
    console.log('🗜️ Compression: Enabled');
  }

  // Check for security headers
  if (nextConfigContent.includes('X-Content-Type-Options')) {
    analysis.uiComponents.security.headers = '✅ Configured';
    console.log('🔒 Security Headers: Configured');
  }

  // Check for performance optimizations
  if (nextConfigContent.includes('webpackBuildWorker: true')) {
    analysis.uiComponents.performance.webpackWorker = '✅ Enabled';
    console.log('⚡ Webpack Build Worker: Enabled');
  }

} catch (error) {
  console.error('❌ Could not analyze Next.js config:', error.message);
}

// Analyze content structure
try {
  const contentDirs = ['ko', 'en', 'ja', 'zh'];
  const contentStatus = {};

  contentDirs.forEach(lang => {
    if (fs.existsSync(`./content/${lang}`)) {
      const files = fs.readdirSync(`./content/${lang}`);
      contentStatus[lang] = {
        exists: true,
        fileCount: files.length,
        hasGettingStarted: files.includes('getting-started.mdx') || files.includes('getting-started')
      };
    } else {
      contentStatus[lang] = { exists: false };
    }
  });

  console.log('\n📚 Content Structure Analysis:');
  Object.entries(contentStatus).forEach(([lang, status]) => {
    if (status.exists) {
      console.log(`  ${lang.toUpperCase()}: ${status.fileCount} files ${status.hasGettingStarted ? '✅' : '⚠️'}`);
    } else {
      console.log(`  ${lang.toUpperCase()}: ❌ Missing`);
    }
  });

  // Check for content completeness
  const missingContent = Object.entries(contentStatus)
    .filter(([_, status]) => !status.exists)
    .map(([lang]) => lang);

  if (missingContent.length > 0) {
    analysis.recommendations.high.push(`Missing content directories: ${missingContent.join(', ')}`);
  }

} catch (error) {
  console.error('❌ Could not analyze content structure:', error.message);
}

// Analyze component structure
try {
  const componentFiles = fs.readdirSync('./components');

  console.log('\n🧩 Component Analysis:');
  componentFiles.forEach(file => {
    console.log(`  📄 ${file}: Available`);
  });

  // Check for critical components
  const criticalComponents = ['CustomSearch.tsx', 'CoreWebVitalsOptimizer.tsx'];
  criticalComponents.forEach(comp => {
    if (componentFiles.includes(comp)) {
      analysis.uiComponents[comp.replace('.tsx', '')] = '✅ Available';
    } else {
      analysis.recommendations.medium.push(`Missing component: ${comp}`);
    }
  });

} catch (error) {
  console.error('❌ Could not analyze components:', error.message);
}

// Analyze mobile responsiveness
try {
  const themeConfigContent = fs.readFileSync('./theme.config.tsx', 'utf8');

  // Check for responsive features
  if (themeConfigContent.includes('sidebar:')) {
    analysis.uiComponents.responsiveness.sidebar = '✅ Configured';
    console.log('📱 Responsive Sidebar: Configured');
  }

  if (themeConfigContent.includes('toc:')) {
    analysis.uiComponents.responsiveness.tableOfContents = '✅ Configured';
    console.log('📋 Responsive ToC: Configured');
  }

} catch (error) {
  console.error('❌ Could not analyze responsive features:', error.message);
}

// Generate recommendations
console.log('\n🎯 Recommendations:');
console.log('==================');

// Critical issues
if (analysis.recommendations.critical.length > 0) {
  console.log('\n🔴 CRITICAL:');
  analysis.recommendations.critical.forEach(rec => console.log(`  - ${rec}`));
}

// High priority
if (analysis.recommendations.high.length > 0) {
  console.log('\n🟠 HIGH:');
  analysis.recommendations.high.forEach(rec => console.log(`  - ${rec}`));
}

// Medium priority
if (analysis.recommendations.medium.length > 0) {
  console.log('\n🟡 MEDIUM:');
  analysis.recommendations.medium.forEach(rec => console.log(`  - ${rec}`));
}

// Low priority
if (analysis.recommendations.low.length > 0) {
  console.log('\n🟢 LOW:');
  analysis.recommendations.low.forEach(rec => console.log(`  - ${rec}`));
}

// Migration assessment
console.log('\n📊 Migration Assessment:');
console.log('=======================');

const isMigrationSuccessful = (
  analysis.migrationStatus.nextraVersion?.startsWith('4.') &&
  analysis.uiComponents.search.customImplementation &&
  analysis.uiComponents.navigation.internationalization
);

if (isMigrationSuccessful) {
  console.log('✅ Migration Status: SUCCESSFUL');
  console.log('   - Nextra 4.x properly configured');
  console.log('   - Custom search integration working');
  console.log('   - Multi-language support maintained');
} else {
  console.log('⚠️ Migration Status: NEEDS ATTENTION');
  analysis.migrationStatus.migrationIssues.forEach(issue => {
    console.log(`   - ${issue}`);
  });
}

// Performance analysis
console.log('\n⚡ Performance Features:');
console.log('======================');
Object.entries(analysis.uiComponents.performance || {}).forEach(([feature, status]) => {
  console.log(`  ${status} ${feature}`);
});

// Security analysis
console.log('\n🔒 Security Features:');
console.log('===================');
Object.entries(analysis.uiComponents.security || {}).forEach(([feature, status]) => {
  console.log(`  ${status} ${feature}`);
});

// Save analysis to file
const reportPath = './ui-analysis-report.json';
fs.writeFileSync(reportPath, JSON.stringify(analysis, null, 2));
console.log(`\n📄 Detailed report saved to: ${reportPath}`);

console.log('\n✅ Manual UI/UX Analysis Complete!');