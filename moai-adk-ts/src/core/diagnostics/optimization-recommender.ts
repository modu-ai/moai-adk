/**
 * @file Optimization recommendation engine
 * @author MoAI Team
 * @tags @FEATURE:OPTIMIZATION-RECOMMENDER-001 @REQ:ADVANCED-DOCTOR-001
 */

import type {
  BenchmarkResult,
  OptimizationRecommendation,
  SystemPerformanceMetrics,
} from '@/types/diagnostics';

/**
 * Optimization recommendation engine for system analysis
 * @tags @FEATURE:OPTIMIZATION-RECOMMENDER-001
 */
export class OptimizationRecommender {
  /**
   * Generate optimization recommendations based on system metrics
   * @param performanceMetrics - System performance data
   * @param benchmarkResults - Benchmark test results
   * @returns Array of optimization recommendations
   * @tags @API:GENERATE-RECOMMENDATIONS-001
   */
  public async generateRecommendations(
    performanceMetrics?: SystemPerformanceMetrics,
    benchmarkResults?: BenchmarkResult[]
  ): Promise<OptimizationRecommendation[]> {
    const recommendations: OptimizationRecommendation[] = [];

    // Performance-based recommendations
    if (performanceMetrics) {
      recommendations.push(
        ...this.analyzePerformanceMetrics(performanceMetrics)
      );
    }

    // Benchmark-based recommendations
    if (benchmarkResults) {
      recommendations.push(...this.analyzeBenchmarkResults(benchmarkResults));
    }

    // System-specific recommendations
    recommendations.push(...this.getSystemSpecificRecommendations());

    // Sort by severity and impact
    return recommendations.sort((a, b) => {
      const severityOrder = { critical: 4, error: 3, warning: 2, info: 1 };
      const impactOrder = { critical: 4, high: 3, medium: 2, low: 1 };

      const severityDiff =
        severityOrder[b.severity] - severityOrder[a.severity];
      if (severityDiff !== 0) return severityDiff;

      return impactOrder[b.impact] - impactOrder[a.impact];
    });
  }

  /**
   * Analyze performance metrics for recommendations
   * @param metrics - System performance metrics
   * @returns Performance-based recommendations
   * @tags @UTIL:ANALYZE-PERFORMANCE-001
   */
  private analyzePerformanceMetrics(
    metrics: SystemPerformanceMetrics
  ): OptimizationRecommendation[] {
    const recommendations: OptimizationRecommendation[] = [];

    // CPU usage analysis
    if (metrics.cpuUsage > 80) {
      recommendations.push({
        category: 'performance',
        severity: DiagnosticSeverity.CRITICAL,
        title: 'High CPU Usage Detected',
        description: `CPU usage is at ${metrics.cpuUsage.toFixed(1)}%, which may cause system slowdowns`,
        impact: 'critical',
        effort: 'medium',
        steps: [
          'Close unnecessary applications and processes',
          'Check for CPU-intensive background tasks',
          'Consider upgrading to a faster CPU if persistent',
          'Monitor for malware or runaway processes',
        ],
        resources: [
          'https://docs.microsoft.com/en-us/windows/win32/perfctrs/performance-counters-portal',
        ],
      });
    } else if (metrics.cpuUsage > 60) {
      recommendations.push({
        category: 'performance',
        severity: DiagnosticSeverity.WARNING,
        title: 'Moderate CPU Usage',
        description: `CPU usage is at ${metrics.cpuUsage.toFixed(1)}%, consider optimization`,
        impact: 'medium',
        effort: 'easy',
        steps: [
          'Review running applications',
          'Close browser tabs and unused programs',
          'Consider process optimization',
        ],
      });
    }

    // Memory usage analysis
    if (metrics.memoryUsage.percentage > 85) {
      recommendations.push({
        category: 'performance',
        severity: DiagnosticSeverity.CRITICAL,
        title: 'Critical Memory Usage',
        description: `Memory usage is at ${metrics.memoryUsage.percentage}% (${metrics.memoryUsage.used}MB/${metrics.memoryUsage.total}MB)`,
        impact: 'critical',
        effort: 'medium',
        steps: [
          'Close memory-intensive applications',
          'Consider adding more RAM',
          'Check for memory leaks in applications',
          'Restart the system if necessary',
        ],
      });
    } else if (metrics.memoryUsage.percentage > 70) {
      recommendations.push({
        category: 'performance',
        severity: DiagnosticSeverity.WARNING,
        title: 'High Memory Usage',
        description: `Memory usage is at ${metrics.memoryUsage.percentage}%, optimization recommended`,
        impact: 'medium',
        effort: 'easy',
        steps: [
          'Close unnecessary applications',
          'Clear browser cache and temporary files',
          'Consider increasing virtual memory',
        ],
      });
    }

    // Disk space analysis
    if (metrics.diskSpace.percentage > 90) {
      recommendations.push({
        category: 'maintenance',
        severity: DiagnosticSeverity.CRITICAL,
        title: 'Critical Disk Space Warning',
        description: `Disk usage is at ${metrics.diskSpace.percentage}% (${metrics.diskSpace.used}GB used)`,
        impact: 'critical',
        effort: 'medium',
        steps: [
          'Delete unnecessary files and folders',
          'Empty trash/recycle bin',
          'Use disk cleanup utilities',
          'Move large files to external storage',
          'Consider upgrading to a larger drive',
        ],
      });
    } else if (metrics.diskSpace.percentage > 80) {
      recommendations.push({
        category: 'maintenance',
        severity: DiagnosticSeverity.WARNING,
        title: 'Low Disk Space',
        description: `Disk usage is at ${metrics.diskSpace.percentage}%, cleanup recommended`,
        impact: 'medium',
        effort: 'easy',
        steps: [
          'Run disk cleanup',
          'Delete temporary files',
          'Uninstall unused programs',
        ],
      });
    }

    // Network latency analysis
    if (metrics.networkLatency && metrics.networkLatency > 500) {
      recommendations.push({
        category: 'performance',
        severity: DiagnosticSeverity.WARNING,
        title: 'High Network Latency',
        description: `Network latency is ${metrics.networkLatency}ms, which may affect productivity`,
        impact: 'medium',
        effort: 'medium',
        steps: [
          'Check network connection quality',
          'Consider switching to a wired connection',
          'Contact network administrator or ISP',
          'Use network optimization tools',
        ],
      });
    }

    return recommendations;
  }

  /**
   * Analyze benchmark results for recommendations
   * @param results - Benchmark test results
   * @returns Benchmark-based recommendations
   * @tags @UTIL:ANALYZE-BENCHMARKS-001
   */
  private analyzeBenchmarkResults(
    results: BenchmarkResult[]
  ): OptimizationRecommendation[] {
    const recommendations: OptimizationRecommendation[] = [];

    for (const result of results) {
      if (result.status === 'fail') {
        recommendations.push({
          category: 'performance',
          severity: DiagnosticSeverity.ERROR,
          title: `${result.name} Benchmark Failed`,
          description: `The ${result.name} benchmark failed, indicating potential performance issues`,
          impact: 'high',
          effort: 'hard',
          steps: [
            'Review system requirements',
            'Check for hardware issues',
            'Update system drivers',
            ...(result.recommendations || []),
          ],
        });
      } else if (result.status === 'warning' || result.score < 70) {
        recommendations.push({
          category: 'performance',
          severity: DiagnosticSeverity.WARNING,
          title: `${result.name} Performance Below Optimal`,
          description: `${result.name} scored ${result.score}/100, indicating room for improvement`,
          impact: 'medium',
          effort: 'medium',
          steps: [
            'Consider hardware upgrades',
            'Optimize system settings',
            'Close background applications during intensive tasks',
            ...(result.recommendations || []),
          ],
        });
      }
    }

    return recommendations;
  }

  /**
   * Get general system-specific recommendations
   * @returns System maintenance recommendations
   * @tags @UTIL:SYSTEM-RECOMMENDATIONS-001
   */
  private getSystemSpecificRecommendations(): OptimizationRecommendation[] {
    const recommendations: OptimizationRecommendation[] = [];

    // General maintenance recommendations
    recommendations.push({
      category: 'maintenance',
      severity: DiagnosticSeverity.INFO,
      title: 'Regular System Maintenance',
      description:
        'Perform regular system maintenance to ensure optimal performance',
      impact: 'medium',
      effort: 'easy',
      steps: [
        'Update operating system regularly',
        'Run antivirus scans weekly',
        'Defragment hard drives monthly (if using HDD)',
        'Clean registry periodically',
        'Update device drivers',
      ],
    });

    // Security recommendations
    recommendations.push({
      category: 'security',
      severity: DiagnosticSeverity.INFO,
      title: 'Security Best Practices',
      description: 'Follow security best practices to protect your system',
      impact: 'high',
      effort: 'easy',
      steps: [
        'Keep operating system and software updated',
        'Use strong, unique passwords',
        'Enable automatic security updates',
        'Use reputable antivirus software',
        'Be cautious with email attachments and downloads',
      ],
    });

    return recommendations;
  }
}
