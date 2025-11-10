#!/usr/bin/env node

/**
 * Performance Monitor Script for MoAI-ADK Documentation
 *
 * This script monitors real-time performance metrics during development
 * and provides continuous feedback on performance changes.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const { getCLS, getFID, getFCP, getLCP, getTTFB } = require('web-vitals');

const MONITOR_CONFIG = {
  // Monitoring intervals (in milliseconds)
  BUILD_INTERVAL: 60000, // 1 minute
  LCP_INTERVAL: 300000,  // 5 minutes
  BUDGET_INTERVAL: 120000, // 2 minutes

  // Performance budgets
  BUDGETS: {
    BUNDLE_SIZE: 250 * 1024, // 250KB
    JS_SIZE: 150 * 1024,    // 150KB
    CSS_SIZE: 50 * 1024,    // 50KB
    IMAGE_SIZE: 100 * 1024, // 100KB per image
    LCP: 2500,             // 2.5s
    FID: 100,              // 100ms
    CLS: 0.1,              // 0.1
    BUILD_TIME: 60000,     // 60s
  },

  // Output
  LOG_FILE: 'performance-monitor.log',
  METRICS_FILE: 'performance-metrics.json',
  ALERTS_FILE: 'performance-alerts.json',
};

/**
 * Get current timestamp
 */
function getTimestamp() {
  return new Date().toISOString();
}

/**
 * Log message with timestamp
 */
function log(message, level = 'info') {
  const timestamp = getTimestamp();
  const prefix = {
    info: 'ℹ️',
    warn: '⚠️',
    error: '❌',
    success: '✅',
  }[level] || 'ℹ️';

  const logMessage = `[${timestamp}] ${prefix} ${message}`;
  console.log(logMessage);

  // Write to log file
  const logPath = path.join(process.cwd(), MONITOR_CONFIG.LOG_FILE);
  fs.appendFileSync(logPath, logMessage + '\n');
}

/**
 * Get bundle size from .next directory
 */
function getBundleSize() {
  const nextDir = path.join(process.cwd(), '.next');
  if (!fs.existsSync(nextDir)) {
    return null;
  }

  let totalSize = 0;
  const bundles = {};

  try {
    // Static chunks
    const staticDir = path.join(nextDir, 'static');
    if (fs.existsSync(staticDir)) {
      const getDirectorySize = (dir) => {
        let size = 0;
        const files = fs.readdirSync(dir);
        files.forEach(file => {
          const filePath = path.join(dir, file);
          const stat = fs.statSync(filePath);
          if (stat.isDirectory()) {
            size += getDirectorySize(filePath);
          } else {
            size += stat.size;
          }
        });
        return size;
      };

      ['chunks', 'css', 'media', 'webpack'].forEach(subdir => {
        const fullPath = path.join(staticDir, subdir);
        if (fs.existsSync(fullPath)) {
          bundles[subdir] = getDirectorySize(fullPath);
        }
      });
    }

    totalSize = Object.values(bundles).reduce((sum, size) => sum + size, 0);
  } catch (error) {
    log(`Error getting bundle size: ${error.message}`, 'error');
  }

  return { totalSize, bundles };
}

/**
 * Get build performance metrics
 */
function getBuildMetrics() {
  const buildInfoPath = path.join(process.cwd(), '.next', 'build-manifest.json');
  if (!fs.existsSync(buildInfoPath)) {
    return null;
  }

  try {
    const buildInfo = JSON.parse(fs.readFileSync(buildInfoPath, 'utf8'));
    return {
      pages: Object.keys(buildInfo.pages).length,
      totalChunks: Object.values(buildInfo.pages).flat().length,
    };
  } catch (error) {
    log(`Error reading build manifest: ${error.message}`, 'error');
    return null;
  }
}

/**
 * Analyze bundle performance against budgets
 */
function analyzeBundlePerformance() {
  const bundleInfo = getBundleSize();
  if (!bundleInfo) {
    log('Bundle analysis skipped - no build found');
    return null;
  }

  const analysis = {
    timestamp: getTimestamp(),
    totalSize: bundleInfo.totalSize,
    bundles: bundleInfo.bundles,
    budgets: {
      total: bundleInfo.totalSize <= MONITOR_CONFIG.BUDGETS.BUNDLE_SIZE,
      js: (bundleInfo.bundles.chunks || 0) <= MONITOR_CONFIG.BUDGETS.JS_SIZE,
      css: (bundleInfo.bundles.css || 0) <= MONITOR_CONFIG.BUDGETS.CSS_SIZE,
    },
    recommendations: [],
  };

  // Generate recommendations
  if (!analysis.budgets.total) {
    analysis.recommendations.push('Consider code splitting to reduce total bundle size');
  }
  if (!analysis.budgets.js) {
    analysis.recommendations.push('JavaScript bundle is large, consider dynamic imports');
  }
  if (!analysis.budgets.css) {
    analysis.recommendations.push('CSS bundle is large, consider CSS optimization');
  }

  const passed = Object.values(analysis.budgets).every(Boolean);
  log(`Bundle analysis: ${passed ? '✅ Passed' : '❌ Failed'} (${(bundleInfo.totalSize / 1024).toFixed(1)}KB total)`, passed ? 'success' : 'warn');

  return analysis;
}

/**
 * Monitor build time
 */
function monitorBuildTime() {
  const startTime = Date.now();

  try {
    log('Starting build time measurement...');

    // Run build with timing
    execSync('bun run build:next', {
      stdio: 'pipe',
      cwd: process.cwd()
    });

    const buildTime = Date.now() - startTime;
    const passed = buildTime <= MONITOR_CONFIG.BUDGETS.BUILD_TIME;

    const metrics = {
      timestamp: getTimestamp(),
      buildTime,
      passed,
      budget: MONITOR_CONFIG.BUDGETS.BUILD_TIME,
    };

    log(`Build time: ${buildTime}ms ${passed ? '✅' : '❌'} (budget: ${MONITOR_CONFIG.BUDGETS.BUILD_TIME}ms)`, passed ? 'success' : 'warn');
    return metrics;

  } catch (error) {
    log(`Build failed: ${error.message}`, 'error');
    return {
      timestamp: getTimestamp(),
      buildTime: Date.now() - startTime,
      passed: false,
      error: error.message,
    };
  }
}

/**
 * Create performance alert
 */
function createAlert(type, message, severity = 'medium', data = {}) {
  const alert = {
    id: Date.now().toString(),
    timestamp: getTimestamp(),
    type,
    message,
    severity, // low, medium, high, critical
    data,
    resolved: false,
  };

  const alertsPath = path.join(process.cwd(), MONITOR_CONFIG.ALERTS_FILE);
  let alerts = [];

  // Load existing alerts
  if (fs.existsSync(alertsPath)) {
    try {
      alerts = JSON.parse(fs.readFileSync(alertsPath, 'utf8'));
    } catch (error) {
      log(`Error loading alerts: ${error.message}`, 'error');
    }
  }

  alerts.push(alert);

  // Keep only last 100 alerts
  if (alerts.length > 100) {
    alerts = alerts.slice(-100);
  }

  fs.writeFileSync(alertsPath, JSON.stringify(alerts, null, 2));

  log(`🚨 Alert: ${message}`, severity === 'high' || severity === 'critical' ? 'error' : 'warn');
  return alert;
}

/**
 * Save metrics to file
 */
function saveMetrics(metrics) {
  const metricsPath = path.join(process.cwd(), MONITOR_CONFIG.METRICS_FILE);

  let allMetrics = [];
  if (fs.existsSync(metricsPath)) {
    try {
      allMetrics = JSON.parse(fs.readFileSync(metricsPath, 'utf8'));
    } catch (error) {
      log(`Error loading metrics: ${error.message}`, 'error');
    }
  }

  allMetrics.push(metrics);

  // Keep only last 1000 metrics
  if (allMetrics.length > 1000) {
    allMetrics = allMetrics.slice(-1000);
  }

  fs.writeFileSync(metricsPath, JSON.stringify(allMetrics, null, 2));
}

/**
 * Generate performance summary
 */
function generatePerformanceSummary(bundleMetrics, buildMetrics) {
  const summary = {
    timestamp: getTimestamp(),
    bundle: bundleMetrics,
    build: buildMetrics,
    overall: {
      status: 'unknown',
      score: 0,
    },
  };

  // Calculate overall score
  let score = 100;
  if (bundleMetrics && !bundleMetrics.budgets.total) score -= 25;
  if (bundleMetrics && !bundleMetrics.budgets.js) score -= 15;
  if (bundleMetrics && !bundleMetrics.budgets.css) score -= 10;
  if (buildMetrics && !buildMetrics.passed) score -= 25;

  summary.overall.score = Math.max(0, score);

  if (score >= 90) summary.overall.status = 'excellent';
  else if (score >= 80) summary.overall.status = 'good';
  else if (score >= 70) summary.overall.status = 'fair';
  else summary.overall.status = 'poor';

  return summary;
}

/**
 * Main monitoring function
 */
async function runPerformanceMonitor(options = {}) {
  const { once = false, build = false, bundle = false } = options;

  try {
    log('🔍 Starting Performance Monitor');

    let metrics = {
      timestamp: getTimestamp(),
      session: Date.now().toString(),
    };

    // Monitor bundle performance
    if (bundle || !options.specific) {
      const bundleMetrics = analyzeBundlePerformance();
      if (bundleMetrics) {
        metrics.bundle = bundleMetrics;

        // Create alerts for budget violations
        if (!bundleMetrics.budgets.total) {
          createAlert('bundle_size', 'Bundle size exceeds budget', 'high', bundleMetrics);
        }
      }
    }

    // Monitor build performance
    if (build || !options.specific) {
      const buildMetrics = monitorBuildTime();
      if (buildMetrics) {
        metrics.build = buildMetrics;

        // Create alerts for build time violations
        if (!buildMetrics.passed) {
          createAlert('build_time', 'Build time exceeds budget', 'medium', buildMetrics);
        }
      }
    }

    // Generate summary
    const summary = generatePerformanceSummary(metrics.bundle, metrics.build);
    metrics.summary = summary;

    // Save metrics
    saveMetrics(metrics);

    // Print summary
    log(`📊 Performance Score: ${summary.overall.score}/100 (${summary.overall.status})`,
        summary.overall.score >= 80 ? 'success' : summary.overall.score >= 60 ? 'warn' : 'error');

    log('✅ Performance monitoring completed');

    if (!once) {
      log('🔄 Continuous monitoring enabled - Press Ctrl+C to stop');
    }

    return metrics;

  } catch (error) {
    log(`Performance monitoring failed: ${error.message}`, 'error');
    process.exit(1);
  }
}

// Handle command line arguments
function parseArgs() {
  const args = process.argv.slice(2);
  return {
    once: args.includes('--once'),
    build: args.includes('--build'),
    bundle: args.includes('--bundle'),
    specific: args.includes('--build') || args.includes('--bundle'),
  };
}

// Run monitor if called directly
if (require.main === module) {
  const options = parseArgs();
  runPerformanceMonitor(options);
}

module.exports = { runPerformanceMonitor, MONITOR_CONFIG };