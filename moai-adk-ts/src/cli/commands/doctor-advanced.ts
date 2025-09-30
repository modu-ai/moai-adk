/**
 * @file Advanced CLI doctor command implementation
 * @author MoAI Team
 * @tags @FEATURE:ADVANCED-DOCTOR-001 @REQ:ADVANCED-DOCTOR-001
 */

import chalk from 'chalk';
import type { BenchmarkRunner } from '@/core/diagnostics/benchmark-runner';
import type { EnvironmentAnalyzer } from '@/core/diagnostics/environment-analyzer';
import type { OptimizationRecommender } from '@/core/diagnostics/optimization-recommender';
import type { SystemPerformanceAnalyzer } from '@/core/diagnostics/performance-analyzer';
import type { SystemDetector } from '@/core/system-checker/detector';
import { logger } from '../../utils/winston-logger.js';
import {
  DiagnosticSeverity,
  type AdvancedDoctorResult,
  type BenchmarkResult,
  type DoctorOptions,
  type EnvironmentAnalysis,
  type OptimizationRecommendation,
  type SystemPerformanceMetrics,
} from '@/types/diagnostics';

/**
 * Advanced doctor command for comprehensive system diagnostics
 * @tags @FEATURE:ADVANCED-DOCTOR-001
 */
export class AdvancedDoctorCommand {
  constructor(
    readonly _systemDetector: SystemDetector,
    private readonly performanceAnalyzer: SystemPerformanceAnalyzer,
    private readonly benchmarkRunner: BenchmarkRunner,
    private readonly optimizationRecommender: OptimizationRecommender,
    private readonly environmentAnalyzer: EnvironmentAnalyzer
  ) {}

  /**
   * Run advanced system diagnostics
   * @param options - Doctor command options
   * @returns Advanced doctor result
   * @tags @API:RUN-ADVANCED-001
   */
  public async runAdvanced(
    options: DoctorOptions = {}
  ): Promise<AdvancedDoctorResult> {
    if (options.verbose) {
      this.printAdvancedHeader();
    }

    const results = await this.gatherAdvancedMetrics(options);
    const healthScore = this.calculateHealthScore(results);
    const summary = this.generateAdvancedSummary(results, healthScore);

    if (options.verbose) {
      this.printAdvancedResults(results, healthScore, summary);
    }

    return {
      allPassed: summary.criticalIssues === 0 && summary.warnings === 0,
      basicChecks: {
        total: 0, // Basic checks would be run separately
        passed: 0,
        failed: 0,
      },
      performanceMetrics: results.performanceMetrics,
      benchmarks: results.benchmarks,
      recommendations: results.recommendations,
      environments: results.environments,
      healthScore,
      summary,
    };
  }

  /**
   * Print advanced diagnostic header
   * @tags @UTIL:PRINT-ADVANCED-HEADER-001
   */
  private printAdvancedHeader(): void {
    logger.info(chalk.blue.bold('🔍 Advanced MoAI-ADK System Diagnostics'));
    logger.info(chalk.blue('Running comprehensive system analysis...\n'));
  }

  /**
   * Gather all advanced metrics and analysis
   * @param options - Doctor command options
   * @returns Collected metrics and analysis results
   * @tags @UTIL:GATHER-ADVANCED-METRICS-001
   */
  private async gatherAdvancedMetrics(options: DoctorOptions): Promise<{
    performanceMetrics: SystemPerformanceMetrics;
    benchmarks: BenchmarkResult[];
    recommendations: OptimizationRecommendation[];
    environments: EnvironmentAnalysis[];
  }> {
    // Gather performance metrics
    const performanceMetrics = await this.performanceAnalyzer.analyzeSystem();

    // Run benchmarks if requested
    let benchmarks: BenchmarkResult[] = [];
    if (options.includeBenchmarks) {
      try {
        benchmarks = await this.benchmarkRunner.runAllBenchmarks();
      } catch (error) {
        logger.warn(chalk.yellow('⚠️ Benchmark execution failed:'),
          error instanceof Error ? { error: error.message } : { error: String(error) }
        );
        benchmarks = [];
      }
    }

    // Generate recommendations if requested
    let recommendations: OptimizationRecommendation[] = [];
    if (options.includeRecommendations) {
      recommendations =
        await this.optimizationRecommender.generateRecommendations(
          performanceMetrics,
          benchmarks
        );
    }

    // Analyze environments if requested
    let environments: EnvironmentAnalysis[] = [];
    if (options.includeEnvironmentAnalysis) {
      const envConfigs = await this.environmentAnalyzer.analyzeEnvironments();
      environments = envConfigs.map(env => ({
        ...env,
        status: this.mapEnvironmentStatus(env.status),
      }));
    }

    return {
      performanceMetrics,
      benchmarks,
      recommendations,
      environments,
    };
  }

  /**
   * Map environment status from EnvironmentConfig to EnvironmentAnalysis
   * @param status - Original status
   * @returns Mapped status
   * @tags @UTIL:MAP-ENV-STATUS-001
   */
  private mapEnvironmentStatus(
    status: 'optimal' | 'good' | 'needs_improvement' | 'problematic'
  ): 'optimal' | 'good' | 'warning' | 'poor' {
    switch (status) {
      case 'optimal':
        return 'optimal';
      case 'good':
        return 'good';
      case 'needs_improvement':
        return 'warning';
      case 'problematic':
        return 'poor';
      default:
        return 'poor';
    }
  }

  /**
   * Calculate overall system health score
   * @param results - Diagnostic results
   * @returns Health score (0-100)
   * @tags @UTIL:CALCULATE-HEALTH-SCORE-001
   */
  private calculateHealthScore(results: {
    performanceMetrics: SystemPerformanceMetrics;
    benchmarks: BenchmarkResult[];
    recommendations: OptimizationRecommendation[];
    environments: EnvironmentAnalysis[];
  }): number {
    let score = 100;

    // Performance metrics impact (40% of score)
    const { performanceMetrics } = results;

    // CPU usage penalty
    if (performanceMetrics.cpuUsage > 80) score -= 15;
    else if (performanceMetrics.cpuUsage > 60) score -= 8;
    else if (performanceMetrics.cpuUsage > 40) score -= 3;

    // Memory usage penalty
    if (performanceMetrics.memoryUsage.percentage > 85) score -= 15;
    else if (performanceMetrics.memoryUsage.percentage > 70) score -= 8;
    else if (performanceMetrics.memoryUsage.percentage > 50) score -= 3;

    // Disk space penalty
    if (performanceMetrics.diskSpace.percentage > 90) score -= 10;
    else if (performanceMetrics.diskSpace.percentage > 80) score -= 5;

    // Benchmark results impact (30% of score)
    if (results.benchmarks.length > 0) {
      const avgBenchmarkScore =
        results.benchmarks.reduce((sum, b) => sum + b.score, 0) /
        results.benchmarks.length;
      const failedBenchmarks = results.benchmarks.filter(
        b => b.status === 'fail'
      ).length;

      score -= (100 - avgBenchmarkScore) * 0.3; // Scale benchmark impact
      score -= failedBenchmarks * 5; // Penalty for failed benchmarks
    }

    // Recommendations impact (20% of score)
    for (const rec of results.recommendations) {
      switch (rec.severity) {
        case DiagnosticSeverity.CRITICAL:
          score -= 10;
          break;
        case DiagnosticSeverity.ERROR:
          score -= 7;
          break;
        case DiagnosticSeverity.WARNING:
          score -= 3;
          break;
        case DiagnosticSeverity.INFO:
          score -= 1;
          break;
      }
    }

    // Environment status impact (10% of score)
    for (const env of results.environments) {
      switch (env.status) {
        case 'poor':
          score -= 5;
          break;
        case 'warning':
          score -= 2;
          break;
        case 'good':
          score += 1;
          break;
        case 'optimal':
          score += 2;
          break;
      }
    }

    return Math.max(0, Math.min(100, Math.round(score)));
  }

  /**
   * Generate advanced summary
   * @param results - Diagnostic results
   * @param healthScore - Calculated health score
   * @returns Summary object
   * @tags @UTIL:GENERATE-ADVANCED-SUMMARY-001
   */
  private generateAdvancedSummary(
    results: {
      recommendations: OptimizationRecommendation[];
    },
    healthScore: number
  ): AdvancedDoctorResult['summary'] {
    const criticalIssues = results.recommendations.filter(
      r => r.severity === DiagnosticSeverity.CRITICAL
    ).length;

    const warnings = results.recommendations.filter(
      r =>
        r.severity === DiagnosticSeverity.WARNING ||
        r.severity === DiagnosticSeverity.ERROR
    ).length;

    const suggestions = results.recommendations.filter(
      r => r.severity === DiagnosticSeverity.INFO
    ).length;

    let status: 'excellent' | 'good' | 'fair' | 'poor';
    if (healthScore >= 90) status = 'excellent';
    else if (healthScore >= 70) status = 'good';
    else if (healthScore >= 50) status = 'fair';
    else status = 'poor';

    return {
      status,
      criticalIssues,
      warnings,
      suggestions,
    };
  }

  /**
   * Print advanced diagnostic results
   * @param results - Diagnostic results
   * @param healthScore - Health score
   * @param summary - Result summary
   * @tags @UTIL:PRINT-ADVANCED-RESULTS-001
   */
  private printAdvancedResults(
    results: {
      performanceMetrics: SystemPerformanceMetrics;
      benchmarks: BenchmarkResult[];
      recommendations: OptimizationRecommendation[];
      environments: EnvironmentAnalysis[];
    },
    healthScore: number,
    summary: AdvancedDoctorResult['summary']
  ): void {
    // Performance metrics
    logger.info(chalk.bold('📊 Performance Metrics:'));
    logger.info(
      `  CPU Usage: ${results.performanceMetrics.cpuUsage.toFixed(1)}%`
    );
    logger.info(
      `  Memory Usage: ${results.performanceMetrics.memoryUsage.percentage}% (${results.performanceMetrics.memoryUsage.used}MB/${results.performanceMetrics.memoryUsage.total}MB)`
    );
    logger.info(
      `  Disk Usage: ${results.performanceMetrics.diskSpace.percentage}% (${results.performanceMetrics.diskSpace.used}GB/${results.performanceMetrics.diskSpace.available + results.performanceMetrics.diskSpace.used}GB)`
    );
    if (results.performanceMetrics.networkLatency) {
      logger.info(
        `  Network Latency: ${results.performanceMetrics.networkLatency}ms`
      );
    }
    logger.info('');

    // Benchmarks
    if (results.benchmarks.length > 0) {
      logger.info(chalk.bold('🏃 Benchmark Results:'));
      results.benchmarks.forEach(benchmark => {
        const statusIcon =
          benchmark.status === 'pass'
            ? '✅'
            : benchmark.status === 'warning'
              ? '⚠️'
              : '❌';
        logger.info(
          `  ${statusIcon} ${benchmark.name}: ${benchmark.score}/100 (${benchmark.duration}ms)`
        );
      });
      logger.info('');
    }

    // Environments
    if (results.environments.length > 0) {
      logger.info(chalk.bold('🛠️ Development Environments:'));
      results.environments.forEach(env => {
        const statusIcon =
          env.status === 'optimal'
            ? '✅'
            : env.status === 'good'
              ? '👍'
              : env.status === 'warning'
                ? '⚠️'
                : '❌';
        logger.info(
          `  ${statusIcon} ${env.name} ${env.version || 'unknown'} - ${env.status}`
        );
      });
      logger.info('');
    }

    // Health score
    const scoreColor =
      healthScore >= 90
        ? chalk.green
        : healthScore >= 70
          ? chalk.blue
          : healthScore >= 50
            ? chalk.yellow
            : chalk.red;

    logger.info(chalk.bold('🎯 System Health Score:'));
    logger.info(
      `  ${scoreColor(healthScore.toString())}/100 - ${chalk.bold(summary.status.toUpperCase())}`
    );
    logger.info('');

    // Summary
    logger.info(chalk.bold('📋 Summary:'));
    logger.info(`  Critical Issues: ${summary.criticalIssues}`);
    logger.info(`  Warnings: ${summary.warnings}`);
    logger.info(`  Suggestions: ${summary.suggestions}`);
    logger.info('');

    // Recommendations
    if (results.recommendations.length > 0) {
      logger.info(chalk.bold('💡 Top Recommendations:'));
      results.recommendations.slice(0, 5).forEach((rec, index) => {
        const severityIcon =
          rec.severity === DiagnosticSeverity.CRITICAL
            ? '🚨'
            : rec.severity === DiagnosticSeverity.ERROR
              ? '❌'
              : rec.severity === DiagnosticSeverity.WARNING
                ? '⚠️'
                : 'ℹ️';
        logger.info(`  ${index + 1}. ${severityIcon} ${rec.title}`);
        logger.info(`     ${rec.description}`);
      });
    }
  }
}
