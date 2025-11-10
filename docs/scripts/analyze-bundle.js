#!/usr/bin/env node

/**
 * Bundle Analyzer Script for MoAI-ADK Documentation
 *
 * This script analyzes bundle composition, identifies optimization opportunities,
 * and provides detailed insights into bundle structure.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const { gzipSync } = require('zlib');

const ANALYZER_CONFIG = {
  // Analysis thresholds
  THRESHOLDS: {
    LARGE_CHUNK: 50 * 1024,      // 50KB
    LARGE_DEPENDENCY: 20 * 1024, // 20KB
    DUPLICATE_THRESHOLD: 10 * 1024, // 10KB
    COMPRESSION_RATIO: 0.3,      // 30% compression expected
  },

  // Directories to analyze
  DIRECTORIES: [
    '.next/static/chunks',
    '.next/static/css',
    '.next/static/media',
  ],

  // Output
  REPORT_FILE: 'bundle-analysis-report.json',
  VISUALIZATION_FILE: 'bundle-visualization.html',
};

/**
 * Get all files in directory recursively
 */
function getAllFiles(dir, extensions = ['.js', '.css', '.woff', '.woff2', '.png', '.jpg', '.jpeg', '.gif', '.svg']) {
  const files = [];

  if (!fs.existsSync(dir)) {
    return files;
  }

  const scan = (currentDir) => {
    const items = fs.readdirSync(currentDir);

    items.forEach(item => {
      const fullPath = path.join(currentDir, item);
      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        scan(fullPath);
      } else if (extensions.some(ext => item.endsWith(ext))) {
        files.push(fullPath);
      }
    });
  };

  scan(dir);
  return files;
}

/**
 * Analyze file details
 */
function analyzeFile(filePath) {
  const stat = fs.statSync(filePath);
  const content = fs.readFileSync(filePath);

  const analysis = {
    path: filePath,
    relativePath: path.relative(process.cwd(), filePath),
    name: path.basename(filePath),
    extension: path.extname(filePath),
    size: stat.size,
    sizeGzipped: gzipSync(content).length,
    compressionRatio: 1 - (gzipSync(content).length / stat.size),
    lastModified: stat.mtime.toISOString(),
  };

  // Try to extract more information for JavaScript files
  if (analysis.extension === '.js') {
    try {
      const contentStr = content.toString();

      // Check for React components
      analysis.isReactComponent = /React\.Component|useState|useEffect|export default/.test(contentStr);

      // Check for third-party imports
      const imports = contentStr.match(/require\(['"]([^'"]+)['"]\)|from ['"]([^'"]+)['"]/g) || [];
      analysis.dependencies = imports.map(imp => {
        const match = imp.match(/['"]([^'"]+)['"]/);
        return match ? match[1] : null;
      }).filter(Boolean);

      // Estimate module count (very rough approximation)
      analysis.estimatedModules = contentStr.split(/export|import|require/).length;

    } catch (error) {
      // Ignore analysis errors
    }
  }

  return analysis;
}

/**
 * Analyze bundle composition
 */
function analyzeBundleComposition() {
  const analysis = {
    timestamp: new Date().toISOString(),
    files: [],
    summary: {
      totalFiles: 0,
      totalSize: 0,
      totalSizeGzipped: 0,
      averageCompressionRatio: 0,
    },
    categories: {
      javascript: { count: 0, size: 0, files: [] },
      css: { count: 0, size: 0, files: [] },
      fonts: { count: 0, size: 0, files: [] },
      images: { count: 0, size: 0, files: [] },
      other: { count: 0, size: 0, files: [] },
    },
    issues: [],
    recommendations: [],
  };

  // Analyze all directories
  ANALYZER_CONFIG.DIRECTORIES.forEach(dir => {
    const fullPath = path.join(process.cwd(), dir);
    const files = getAllFiles(fullPath);

    files.forEach(filePath => {
      const fileAnalysis = analyzeFile(filePath);
      analysis.files.push(fileAnalysis);
      analysis.summary.totalFiles++;
      analysis.summary.totalSize += fileAnalysis.size;
      analysis.summary.totalSizeGzipped += fileAnalysis.sizeGzipped;
      analysis.summary.averageCompressionRatio += fileAnalysis.compressionRatio;

      // Categorize files
      const category = getCategory(fileAnalysis);
      analysis.categories[category].count++;
      analysis.categories[category].size += fileAnalysis.size;
      analysis.categories[category].files.push(fileAnalysis);

      // Check for issues
      checkForIssues(fileAnalysis, analysis);
    });
  });

  // Calculate average compression ratio
  if (analysis.summary.totalFiles > 0) {
    analysis.summary.averageCompressionRatio /= analysis.summary.totalFiles;
  }

  // Generate recommendations
  generateRecommendations(analysis);

  return analysis;
}

/**
 * Get file category
 */
function getCategory(fileAnalysis) {
  const { extension } = fileAnalysis;

  if (['.js', '.mjs'].includes(extension)) return 'javascript';
  if (['.css'].includes(extension)) return 'css';
  if (['.woff', '.woff2', '.ttf', '.eot'].includes(extension)) return 'fonts';
  if (['.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp', '.avif'].includes(extension)) return 'images';

  return 'other';
}

/**
 * Check for performance issues
 */
function checkForIssues(fileAnalysis, analysis) {
  const { size, sizeGzipped, compressionRatio } = fileAnalysis;

  // Large files
  if (size > ANALYZER_CONFIG.THRESHOLDS.LARGE_CHUNK) {
    analysis.issues.push({
      type: 'large_file',
      severity: 'high',
      file: fileAnalysis.relativePath,
      size,
      message: `Large file detected: ${(size / 1024).toFixed(1)}KB`,
    });
  }

  // Poor compression
  if (compressionRatio < ANALYZER_CONFIG.THRESHOLDS.COMPRESSION_RATIO && size > 1024) {
    analysis.issues.push({
      type: 'poor_compression',
      severity: 'medium',
      file: fileAnalysis.relativePath,
      ratio: compressionRatio,
      message: `Poor compression ratio: ${(compressionRatio * 100).toFixed(1)}%`,
    });
  }

  // Large dependencies (for JS files)
  if (fileAnalysis.dependencies && fileAnalysis.dependencies.length > 20) {
    analysis.issues.push({
      type: 'many_dependencies',
      severity: 'medium',
      file: fileAnalysis.relativePath,
      count: fileAnalysis.dependencies.length,
      message: `Many dependencies: ${fileAnalysis.dependencies.length}`,
    });
  }
}

/**
 * Generate optimization recommendations
 */
function generateRecommendations(analysis) {
  const recommendations = [];

  // Bundle size recommendations
  if (analysis.summary.totalSize > 300 * 1024) {
    recommendations.push({
      priority: 'high',
      type: 'bundle_splitting',
      message: 'Consider implementing code splitting to reduce initial bundle size',
      impact: 'High',
      effort: 'Medium',
    });
  }

  // JavaScript-specific recommendations
  const jsCategory = analysis.categories.javascript;
  if (jsCategory.size > 150 * 1024) {
    recommendations.push({
      priority: 'high',
      type: 'js_optimization',
      message: 'JavaScript bundle is large, consider dynamic imports and tree shaking',
      impact: 'High',
      effort: 'Medium',
    });
  }

  // CSS recommendations
  const cssCategory = analysis.categories.css;
  if (cssCategory.size > 50 * 1024) {
    recommendations.push({
      priority: 'medium',
      type: 'css_optimization',
      message: 'CSS bundle is large, consider critical CSS extraction and CSS-in-JS',
      impact: 'Medium',
      effort: 'Low',
    });
  }

  // Compression recommendations
  if (analysis.summary.averageCompressionRatio < 0.3) {
    recommendations.push({
      priority: 'medium',
      type: 'compression',
      message: 'Consider enabling gzip or Brotli compression for better performance',
      impact: 'Medium',
      effort: 'Low',
    });
  }

  // Large file recommendations
  const largeFiles = analysis.files.filter(f => f.size > ANALYZER_CONFIG.THRESHOLDS.LARGE_CHUNK);
  if (largeFiles.length > 0) {
    recommendations.push({
      priority: 'high',
      type: 'large_files',
      message: `${largeFiles.length} large files detected, consider lazy loading or optimization`,
      impact: 'High',
      effort: 'Medium',
      details: largeFiles.map(f => `${f.relativePath}: ${(f.size / 1024).toFixed(1)}KB`),
    });
  }

  analysis.recommendations = recommendations.sort((a, b) => {
    const priorityOrder = { high: 3, medium: 2, low: 1 };
    return priorityOrder[b.priority] - priorityOrder[a.priority];
  });
}

/**
 * Generate HTML visualization
 */
function generateVisualization(analysis) {
  const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bundle Analysis - MoAI-ADK Documentation</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; background: #f8f9fa; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric { display: inline-block; margin: 10px 20px 10px 0; }
        .metric-value { font-size: 24px; font-weight: bold; color: #2563eb; }
        .metric-label { font-size: 14px; color: #64748b; }
        .chart-container { position: relative; height: 400px; margin: 20px 0; }
        .issue { background: #fef2f2; border-left: 4px solid #ef4444; padding: 12px; margin: 8px 0; }
        .recommendation { background: #f0f9ff; border-left: 4px solid #3b82f6; padding: 12px; margin: 8px 0; }
        .high { border-left-color: #ef4444; }
        .medium { border-left-color: #f59e0b; }
        .low { border-left-color: #10b981; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #e5e7eb; }
        th { background: #f9fafb; font-weight: 600; }
        .file-list { max-height: 400px; overflow-y: auto; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>📦 Bundle Analysis Report</h1>
            <p>Generated: ${analysis.timestamp}</p>

            <div class="metric">
                <div class="metric-value">${(analysis.summary.totalSize / 1024).toFixed(1)}KB</div>
                <div class="metric-label">Total Size</div>
            </div>
            <div class="metric">
                <div class="metric-value">${(analysis.summary.totalSizeGzipped / 1024).toFixed(1)}KB</div>
                <div class="metric-label">Gzipped Size</div>
            </div>
            <div class="metric">
                <div class="metric-value">${analysis.summary.totalFiles}</div>
                <div class="metric-label">Total Files</div>
            </div>
            <div class="metric">
                <div class="metric-value">${(analysis.summary.averageCompressionRatio * 100).toFixed(1)}%</div>
                <div class="metric-label">Avg Compression</div>
            </div>
        </div>

        <div class="card">
            <h2>📊 Bundle Composition</h2>
            <div class="chart-container">
                <canvas id="compositionChart"></canvas>
            </div>
        </div>

        <div class="card">
            <h2>⚠️ Issues Found (${analysis.issues.length})</h2>
            ${analysis.issues.map(issue => `
                <div class="issue ${issue.severity}">
                    <strong>${issue.message}</strong><br>
                    <small>File: ${issue.file} | Size: ${(issue.size / 1024).toFixed(1)}KB</small>
                </div>
            `).join('')}
        </div>

        <div class="card">
            <h2>💡 Recommendations</h2>
            ${analysis.recommendations.map(rec => `
                <div class="recommendation ${rec.priority}">
                    <strong>${rec.message}</strong><br>
                    <small>Priority: ${rec.priority} | Impact: ${rec.impact} | Effort: ${rec.effort}</small>
                </div>
            `).join('')}
        </div>

        <div class="card">
            <h2>📁 File Breakdown</h2>
            <table>
                <thead>
                    <tr>
                        <th>Category</th>
                        <th>Files</th>
                        <th>Size</th>
                        <th>Gzipped</th>
                    </tr>
                </thead>
                <tbody>
                    ${Object.entries(analysis.categories).map(([cat, data]) => `
                        <tr>
                            <td>${cat.charAt(0).toUpperCase() + cat.slice(1)}</td>
                            <td>${data.count}</td>
                            <td>${(data.size / 1024).toFixed(1)}KB</td>
                            <td>${(data.files.reduce((sum, f) => sum + f.sizeGzipped, 0) / 1024).toFixed(1)}KB</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>

        <div class="card">
            <h2>📋 Largest Files</h2>
            <div class="file-list">
                <table>
                    <thead>
                        <tr>
                            <th>File</th>
                            <th>Size</th>
                            <th>Gzipped</th>
                            <th>Compression</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${analysis.files
                            .sort((a, b) => b.size - a.size)
                            .slice(0, 20)
                            .map(file => `
                                <tr>
                                    <td>${file.relativePath}</td>
                                    <td>${(file.size / 1024).toFixed(1)}KB</td>
                                    <td>${(file.sizeGzipped / 1024).toFixed(1)}KB</td>
                                    <td>${(file.compressionRatio * 100).toFixed(1)}%</td>
                                </tr>
                            `).join('')}
                    </tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        // Composition Chart
        const ctx = document.getElementById('compositionChart').getContext('2d');
        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ${JSON.stringify(Object.keys(analysis.categories))},
                datasets: [{
                    data: ${JSON.stringify(Object.values(analysis.categories).map(c => c.size))},
                    backgroundColor: [
                        '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    </script>
</body>
</html>`;

  const outputPath = path.join(process.cwd(), ANALYZER_CONFIG.VISUALIZATION_FILE);
  fs.writeFileSync(outputPath, html);

  return outputPath;
}

/**
 * Main analyzer function
 */
function runBundleAnalyzer() {
  try {
    console.log('🔍 Starting Bundle Analysis');

    // Check if .next directory exists
    if (!fs.existsSync(path.join(process.cwd(), '.next'))) {
      console.error('❌ Build directory not found. Please run "bun run build" first.');
      process.exit(1);
    }

    // Perform analysis
    const analysis = analyzeBundleComposition();

    // Save JSON report
    const reportPath = path.join(process.cwd(), ANALYZER_CONFIG.REPORT_FILE);
    fs.writeFileSync(reportPath, JSON.stringify(analysis, null, 2));

    // Generate HTML visualization
    const visualizationPath = generateVisualization(analysis);

    // Print summary
    console.log('\n📊 Bundle Analysis Summary');
    console.log('='.repeat(50));
    console.log(`Total Files: ${analysis.summary.totalFiles}`);
    console.log(`Total Size: ${(analysis.summary.totalSize / 1024).toFixed(1)}KB`);
    console.log(`Gzipped Size: ${(analysis.summary.totalSizeGzipped / 1024).toFixed(1)}KB`);
    console.log(`Compression Ratio: ${(analysis.summary.averageCompressionRatio * 100).toFixed(1)}%`);
    console.log(`Issues Found: ${analysis.issues.length}`);
    console.log(`Recommendations: ${analysis.recommendations.length}`);

    if (analysis.recommendations.length > 0) {
      console.log('\n💡 Top Recommendations:');
      analysis.recommendations.slice(0, 3).forEach(rec => {
        console.log(`  • ${rec.message} (${rec.priority} priority)`);
      });
    }

    console.log(`\n📄 Detailed report: ${reportPath}`);
    console.log(`📈 Visualization: ${visualizationPath}`);

    // Exit with appropriate code
    const hasHighPriorityIssues = analysis.recommendations.some(r => r.priority === 'high') ||
                                analysis.issues.some(i => i.severity === 'high');
    process.exit(hasHighPriorityIssues ? 1 : 0);

  } catch (error) {
    console.error('❌ Bundle analysis failed:', error.message);
    process.exit(1);
  }
}

// Run analyzer if called directly
if (require.main === module) {
  runBundleAnalyzer();
}

module.exports = { runBundleAnalyzer, ANALYZER_CONFIG };