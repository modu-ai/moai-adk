#!/usr/bin/env node

/**
 * Performance Audit Script for MoAI-ADK Documentation
 *
 * This script performs comprehensive performance audits using Lighthouse
 * and provides actionable insights for optimization.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');

const PERFORMANCE_CONFIG = {
  // Lighthouse performance thresholds
  THRESHOLDS: {
    PERFORMANCE: 85,
    FIRST_CONTENTFUL_PAINT: 2000,
    LARGEST_CONTENTFUL_PAINT: 2500,
    SPEED_INDEX: 3400,
    INTERACTIVE: 5000,
    CUMULATIVE_LAYOUT_SHIFT: 0.1,
    TOTAL_BLOCKING_TIME: 200,
  },

  // URLs to audit
  URLS: [
    'http://localhost:3000',
    'http://localhost:3000/en/getting-started',
    'http://localhost:3000/ko/getting-started',
  ],

  // Output directories
  OUTPUT_DIR: 'performance-reports',
  AUDIT_DIR: 'lighthouse-audits',
};

/**
 * Ensure output directories exist
 */
function ensureDirectories() {
  const dirs = [
    PERFORMANCE_CONFIG.OUTPUT_DIR,
    path.join(PERFORMANCE_CONFIG.OUTPUT_DIR, PERFORMANCE_CONFIG.AUDIT_DIR),
  ];

  dirs.forEach(dir => {
    const fullPath = path.join(process.cwd(), dir);
    if (!fs.existsSync(fullPath)) {
      fs.mkdirSync(fullPath, { recursive: true });
    }
  });
}

/**
 * Check if local server is running
 */
function checkServerRunning(url) {
  return new Promise((resolve) => {
    const request = https.get(url, (res) => {
      resolve(res.statusCode === 200);
    });

    request.on('error', () => resolve(false));
    request.setTimeout(5000, () => {
      request.destroy();
      resolve(false);
    });
  });
}

/**
 * Run Lighthouse audit for a single URL
 */
async function runLighthouseAudit(url, outputPath) {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const reportPath = path.join(outputPath, `lighthouse-${timestamp}.html`);
  const jsonPath = path.join(outputPath, `lighthouse-${timestamp}.json`);

  try {
    console.log(`🚀 Running Lighthouse audit for: ${url}`);

    // Use npx to run lighthouse
    const lighthouseCmd = [
      'npx lighthouse',
      url,
      '--output=html,json',
      `--output-path=${reportPath.replace('.html', '')}`,
      '--chrome-flags="--headless --no-sandbox"',
      '--quiet',
      '--only-categories=performance,accessibility,best-practices,seo',
      '--view'
    ].join(' ');

    execSync(lighthouseCmd, { stdio: 'pipe', cwd: process.cwd() });

    // Read and parse the JSON results
    const jsonContent = fs.readFileSync(jsonPath, 'utf8');
    const results = JSON.parse(jsonContent);

    return {
      url,
      timestamp,
      reportPath,
      jsonPath,
      results,
    };

  } catch (error) {
    console.error(`❌ Lighthouse audit failed for ${url}:`, error.message);
    return null;
  }
}

/**
 * Analyze Lighthouse results and extract key metrics
 */
function analyzeLighthouseResults(auditResult) {
  if (!auditResult || !auditResult.results) return null;

  const { results } = auditResult;
  const { lhr } = results;

  const metrics = {
    url: auditResult.url,
    timestamp: auditResult.timestamp,
    performance: Math.round(lhr.categories.performance.score * 100),
    accessibility: Math.round(lhr.categories.accessibility.score * 100),
    bestPractices: Math.round(lhr.categories['best-practices'].score * 100),
    seo: Math.round(lhr.categories.seo.score * 100),

    // Core Web Vitals
    fcp: lhr.audits['first-contentful-paint'].numericValue,
    lcp: lhr.audits['largest-contentful-paint'].numericValue,
    cls: lhr.audits['cumulative-layout-shift'].numericValue,
    fid: lhr.audits['max-potential-fid']?.numericValue || 0,
    tbt: lhr.audits['total-blocking-time'].numericValue,
    si: lhr.audits['speed-index'].numericValue,
    tti: lhr.audits['interactive'].numericValue,
  };

  // Check against thresholds
  const status = {
    performance: metrics.performance >= PERFORMANCE_CONFIG.THRESHOLDS.PERFORMANCE,
    fcp: metrics.fcp <= PERFORMANCE_CONFIG.THRESHOLDS.FIRST_CONTENTFUL_PAINT,
    lcp: metrics.lcp <= PERFORMANCE_CONFIG.THRESHOLDS.LARGEST_CONTENTFUL_PAINT,
    cls: metrics.cls <= PERFORMANCE_CONFIG.THRESHOLDS.CUMULATIVE_LAYOUT_SHIFT,
    tbt: metrics.tbt <= PERFORMANCE_CONFIG.THRESHOLDS.TOTAL_BLOCKING_TIME,
    interactive: metrics.tti <= PERFORMANCE_CONFIG.THRESHOLDS.INTERACTIVE,
  };

  return {
    ...metrics,
    status,
    passed: Object.values(status).every(Boolean),
  };
}

/**
 * Generate performance report
 */
function generatePerformanceReport(auditResults) {
  const reportPath = path.join(
    process.cwd(),
    PERFORMANCE_CONFIG.OUTPUT_DIR,
    `performance-report-${new Date().toISOString().split('T')[0]}.json`
  );

  const report = {
    generated: new Date().toISOString(),
    config: PERFORMANCE_CONFIG,
    results: auditResults,
    summary: {
      totalAudits: auditResults.length,
      passedAudits: auditResults.filter(r => r && r.passed).length,
      averagePerformance: auditResults.reduce((sum, r) => sum + (r?.performance || 0), 0) / auditResults.length,
      issues: auditResults.flatMap(r => r ? Object.entries(r.status)
        .filter(([_, passed]) => !passed)
        .map(([metric]) => `${metric} failed`) : []),
    },
  };

  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

  return report;
}

/**
 * Print audit summary to console
 */
function printAuditSummary(report) {
  console.log('\n📊 Performance Audit Summary');
  console.log('='.repeat(50));
  console.log(`Generated: ${report.generated}`);
  console.log(`Total Audits: ${report.summary.totalAudits}`);
  console.log(`Passed Audits: ${report.summary.passedAudits}`);
  console.log(`Average Performance Score: ${report.summary.averagePerformance.toFixed(1)}`);

  if (report.summary.issues.length > 0) {
    console.log('\n⚠️ Issues Found:');
    report.summary.issues.forEach(issue => console.log(`  - ${issue}`));
  }

  console.log('\n📁 Detailed Reports:');
  report.results.forEach(result => {
    if (result) {
      const status = result.passed ? '✅' : '❌';
      console.log(`  ${status} ${result.url} - Performance: ${result.performance}`);
    }
  });

  console.log(`\n📄 Full report saved to: ${report.results[0]?.reportPath}`);
}

/**
 * Main audit function
 */
async function runPerformanceAudit() {
  const startTime = Date.now();

  try {
    console.log('🎯 Starting MoAI-ADK Documentation Performance Audit');

    ensureDirectories();

    const auditResults = [];

    for (const url of PERFORMANCE_CONFIG.URLS) {
      // Check if server is running
      const serverRunning = await checkServerRunning(url);
      if (!serverRunning) {
        console.warn(`⚠️ Server not running at ${url}, skipping audit`);
        continue;
      }

      // Run Lighthouse audit
      const auditDir = path.join(
        process.cwd(),
        PERFORMANCE_CONFIG.OUTPUT_DIR,
        PERFORMANCE_CONFIG.AUDIT_DIR
      );

      const auditResult = await runLighthouseAudit(url, auditDir);

      if (auditResult) {
        const analysis = analyzeLighthouseResults(auditResult);
        if (analysis) {
          auditResults.push(analysis);
        }
      }
    }

    if (auditResults.length === 0) {
      console.error('❌ No audits completed. Please ensure the development server is running.');
      process.exit(1);
    }

    // Generate and display report
    const report = generatePerformanceReport(auditResults);
    printAuditSummary(report);

    const duration = ((Date.now() - startTime) / 1000).toFixed(2);
    console.log(`\n✅ Performance audit completed in ${duration}s`);

    // Exit with appropriate code
    const allPassed = auditResults.every(r => r.passed);
    process.exit(allPassed ? 0 : 1);

  } catch (error) {
    console.error('❌ Performance audit failed:', error.message);
    process.exit(1);
  }
}

// Run audit if called directly
if (require.main === module) {
  runPerformanceAudit();
}

module.exports = { runPerformanceAudit, PERFORMANCE_CONFIG };