/**
 * @file Advanced diagnostics type definitions
 * @author MoAI Team
 * @tags @DATA:DIAGNOSTICS-001 @REQ:ADVANCED-DOCTOR-001
 */

/**
 * Diagnostic severity levels
 * @tags @DATA:SEVERITY-LEVELS-001
 */
export enum DiagnosticSeverity {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical'
}

/**
 * System performance metrics
 * @tags @DATA:PERFORMANCE-METRICS-001
 */
export interface SystemPerformanceMetrics {
  cpuUsage: number; // Percentage
  memoryUsage: {
    used: number; // MB
    total: number; // MB
    percentage: number;
  };
  diskSpace: {
    used: number; // GB
    available: number; // GB
    percentage: number;
  };
  networkLatency?: number; // ms
}

/**
 * Performance benchmark result
 * @tags @DATA:BENCHMARK-RESULT-001
 */
export interface BenchmarkResult {
  name: string;
  duration: number; // milliseconds
  status: 'pass' | 'fail' | 'warning';
  baseline?: number; // Expected duration
  score: number; // Performance score (0-100)
  recommendations?: string[];
}

/**
 * Optimization recommendation
 * @tags @DATA:OPTIMIZATION-001
 */
export interface OptimizationRecommendation {
  category: 'performance' | 'security' | 'compatibility' | 'maintenance';
  severity: DiagnosticSeverity;
  title: string;
  description: string;
  impact: 'low' | 'medium' | 'high' | 'critical';
  effort: 'easy' | 'medium' | 'hard';
  steps: string[];
  resources?: string[];
}

/**
 * Environment-specific configuration
 * @tags @DATA:ENV-CONFIG-001
 */
export interface EnvironmentConfig {
  name: string;
  detected: boolean;
  version?: string;
  configFiles: string[];
  recommendations: OptimizationRecommendation[];
  status: 'optimal' | 'good' | 'needs_improvement' | 'problematic';
}

/**
 * Extended doctor result with advanced features
 * @tags @DATA:ADVANCED-DOCTOR-RESULT-001
 */
export interface AdvancedDoctorResult {
  // Basic doctor result
  allPassed: boolean;
  basicChecks: {
    total: number;
    passed: number;
    failed: number;
  };

  // Performance metrics
  performanceMetrics: SystemPerformanceMetrics;

  // Benchmark results
  benchmarks: BenchmarkResult[];

  // Optimization recommendations
  recommendations: OptimizationRecommendation[];

  // Environment configurations
  environments: EnvironmentConfig[];

  // Overall health score
  healthScore: number; // 0-100

  // Summary
  summary: {
    status: 'excellent' | 'good' | 'fair' | 'poor';
    criticalIssues: number;
    warnings: number;
    suggestions: number;
  };
}

/**
 * Benchmark test configuration
 * @tags @DATA:BENCHMARK-CONFIG-001
 */
export interface BenchmarkConfig {
  name: string;
  description: string;
  category: 'io' | 'cpu' | 'memory' | 'network';
  timeout: number; // milliseconds
  baseline: number; // Expected performance baseline
  testFunction: () => Promise<number>; // Returns duration in ms
}

/**
 * Doctor command options
 * @tags @DATA:DOCTOR-OPTIONS-001
 */
export interface DoctorOptions {
  includeBenchmarks?: boolean;
  includeRecommendations?: boolean;
  includeEnvironmentAnalysis?: boolean;
  verbose?: boolean;
  outputFormat?: 'text' | 'json' | 'markdown';
}