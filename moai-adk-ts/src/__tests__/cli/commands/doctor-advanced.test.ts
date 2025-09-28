/**
 * @file Advanced doctor command test suite
 * @author MoAI Team
 * @tags @TEST:ADVANCED-DOCTOR-001 @REQ:ADVANCED-DOCTOR-001
 */

import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import '@/__tests__/setup';
import { AdvancedDoctorCommand } from '@/cli/commands/doctor-advanced';
import { SystemPerformanceAnalyzer } from '@/core/diagnostics/performance-analyzer';
import { BenchmarkRunner } from '@/core/diagnostics/benchmark-runner';
import { OptimizationRecommender } from '@/core/diagnostics/optimization-recommender';
import { EnvironmentAnalyzer } from '@/core/diagnostics/environment-analyzer';
import { SystemDetector } from '@/core/system-checker/detector';
import { DiagnosticSeverity, DoctorOptions } from '@/types/diagnostics';

// Mock modules
jest.mock('@/core/diagnostics/performance-analyzer');
jest.mock('@/core/diagnostics/benchmark-runner');
jest.mock('@/core/diagnostics/optimization-recommender');
jest.mock('@/core/diagnostics/environment-analyzer');
jest.mock('@/core/system-checker/detector');

describe('AdvancedDoctorCommand', () => {
  let doctorCommand: AdvancedDoctorCommand;
  let mockSystemDetector: jest.Mocked<SystemDetector>;
  let mockPerformanceAnalyzer: jest.Mocked<SystemPerformanceAnalyzer>;
  let mockBenchmarkRunner: jest.Mocked<BenchmarkRunner>;
  let mockOptimizationRecommender: jest.Mocked<OptimizationRecommender>;
  let mockEnvironmentAnalyzer: jest.Mocked<EnvironmentAnalyzer>;

  beforeEach(() => {
    jest.clearAllMocks();

    mockSystemDetector = new SystemDetector() as jest.Mocked<SystemDetector>;
    mockPerformanceAnalyzer = new SystemPerformanceAnalyzer() as jest.Mocked<SystemPerformanceAnalyzer>;
    mockBenchmarkRunner = new BenchmarkRunner() as jest.Mocked<BenchmarkRunner>;
    mockOptimizationRecommender = new OptimizationRecommender() as jest.Mocked<OptimizationRecommender>;
    mockEnvironmentAnalyzer = new EnvironmentAnalyzer() as jest.Mocked<EnvironmentAnalyzer>;

    doctorCommand = new AdvancedDoctorCommand(
      mockSystemDetector,
      mockPerformanceAnalyzer,
      mockBenchmarkRunner,
      mockOptimizationRecommender,
      mockEnvironmentAnalyzer
    );
  });

  describe('Comprehensive System Diagnostics', () => {
    test('should run complete diagnostics with all features enabled', async () => {
      // Arrange
      const options: DoctorOptions = {
        includeBenchmarks: true,
        includeRecommendations: true,
        includeEnvironmentAnalysis: true,
        verbose: true
      };

      // Mock system performance metrics
      mockPerformanceAnalyzer.analyzeSystem.mockResolvedValue({
        cpuUsage: 45.2,
        memoryUsage: {
          used: 8192,
          total: 16384,
          percentage: 50.0
        },
        diskSpace: {
          used: 256,
          available: 768,
          percentage: 25.0
        },
        networkLatency: 15
      });

      // Mock benchmark results
      mockBenchmarkRunner.runAllBenchmarks.mockResolvedValue([
        {
          name: 'File I/O Performance',
          duration: 125,
          status: 'pass',
          baseline: 150,
          score: 85,
          recommendations: ['Use SSD for better I/O performance']
        },
        {
          name: 'CPU Performance',
          duration: 89,
          status: 'pass',
          baseline: 100,
          score: 92
        }
      ]);

      // Mock optimization recommendations
      mockOptimizationRecommender.generateRecommendations.mockResolvedValue([
        {
          category: 'performance',
          severity: DiagnosticSeverity.WARNING,
          title: 'Memory Usage Optimization',
          description: 'Memory usage is above 50%, consider optimization',
          impact: 'medium',
          effort: 'easy',
          steps: [
            'Close unnecessary applications',
            'Consider adding more RAM'
          ]
        }
      ]);

      // Mock environment analysis
      mockEnvironmentAnalyzer.analyzeEnvironments.mockResolvedValue([
        {
          name: 'Node.js',
          detected: true,
          version: '18.17.0',
          configFiles: ['package.json', '.nvmrc'],
          recommendations: [],
          status: 'optimal'
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.performanceMetrics.cpuUsage).toBe(45.2);
      expect(result.performanceMetrics.memoryUsage.percentage).toBe(50.0);
      expect(result.benchmarks).toHaveLength(2);
      expect(result.recommendations).toHaveLength(1);
      expect(result.environments).toHaveLength(1);
      expect(result.healthScore).toBeGreaterThan(0);
      expect(result.healthScore).toBeLessThanOrEqual(100);

      // Verify all analyzers were called
      expect(mockPerformanceAnalyzer.analyzeSystem).toHaveBeenCalled();
      expect(mockBenchmarkRunner.runAllBenchmarks).toHaveBeenCalled();
      expect(mockOptimizationRecommender.generateRecommendations).toHaveBeenCalled();
      expect(mockEnvironmentAnalyzer.analyzeEnvironments).toHaveBeenCalled();
    });

    test('should calculate health score based on various metrics', async () => {
      // Arrange
      const options: DoctorOptions = {
        includeBenchmarks: true,
        includeRecommendations: true
      };

      // Mock good performance metrics
      mockPerformanceAnalyzer.analyzeSystem.mockResolvedValue({
        cpuUsage: 25.0,
        memoryUsage: { used: 4096, total: 16384, percentage: 25.0 },
        diskSpace: { used: 128, available: 896, percentage: 12.5 }
      });

      // Mock excellent benchmark results
      mockBenchmarkRunner.runAllBenchmarks.mockResolvedValue([
        {
          name: 'Fast Benchmark',
          duration: 50,
          status: 'pass',
          baseline: 100,
          score: 95
        }
      ]);

      // Mock no critical recommendations
      mockOptimizationRecommender.generateRecommendations.mockResolvedValue([]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.healthScore).toBeGreaterThan(85); // Good health score
      expect(result.summary.status).toBe('excellent');
    });

    test('should identify critical issues and lower health score', async () => {
      // Arrange
      const options: DoctorOptions = {
        includeRecommendations: true
      };

      // Mock poor performance metrics
      mockPerformanceAnalyzer.analyzeSystem.mockResolvedValue({
        cpuUsage: 95.0,
        memoryUsage: { used: 15360, total: 16384, percentage: 93.75 },
        diskSpace: { used: 950, available: 50, percentage: 95.0 }
      });

      // Mock critical recommendations
      mockOptimizationRecommender.generateRecommendations.mockResolvedValue([
        {
          category: 'performance',
          severity: DiagnosticSeverity.CRITICAL,
          title: 'Critical Memory Usage',
          description: 'System is running out of memory',
          impact: 'critical',
          effort: 'hard',
          steps: ['Immediate action required']
        },
        {
          category: 'security',
          severity: DiagnosticSeverity.ERROR,
          title: 'Security Vulnerability',
          description: 'Outdated dependencies detected',
          impact: 'high',
          effort: 'medium',
          steps: ['Update dependencies']
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.healthScore).toBeLessThan(50); // Poor health score
      expect(result.summary.status).toBe('poor');
      expect(result.summary.criticalIssues).toBeGreaterThan(0);
    });
  });

  describe('Performance Benchmarking', () => {
    test('should run performance benchmarks and provide scores', async () => {
      // Arrange
      const options: DoctorOptions = { includeBenchmarks: true };

      mockBenchmarkRunner.runAllBenchmarks.mockResolvedValue([
        {
          name: 'File System I/O',
          duration: 200,
          status: 'warning',
          baseline: 150,
          score: 75,
          recommendations: ['Consider using faster storage']
        },
        {
          name: 'Network Connectivity',
          duration: 50,
          status: 'pass',
          baseline: 100,
          score: 90
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.benchmarks).toHaveLength(2);
      expect(result.benchmarks[0]?.name).toBe('File System I/O');
      expect(result.benchmarks[0]?.score).toBe(75);
      expect(result.benchmarks[1]?.score).toBe(90);
    });

    test('should handle benchmark failures gracefully', async () => {
      // Arrange
      const options: DoctorOptions = { includeBenchmarks: true };

      mockBenchmarkRunner.runAllBenchmarks.mockRejectedValue(
        new Error('Benchmark execution failed')
      );

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.benchmarks).toEqual([]);
      expect(result.summary.warnings).toBeGreaterThan(0);
    });
  });

  describe('Optimization Recommendations', () => {
    test('should provide categorized optimization recommendations', async () => {
      // Arrange
      const options: DoctorOptions = { includeRecommendations: true };

      mockOptimizationRecommender.generateRecommendations.mockResolvedValue([
        {
          category: 'performance',
          severity: DiagnosticSeverity.WARNING,
          title: 'CPU Optimization',
          description: 'CPU usage can be optimized',
          impact: 'medium',
          effort: 'easy',
          steps: ['Close background applications'],
          resources: ['https://example.com/cpu-optimization']
        },
        {
          category: 'security',
          severity: DiagnosticSeverity.ERROR,
          title: 'Security Update Required',
          description: 'Critical security patches available',
          impact: 'high',
          effort: 'medium',
          steps: ['Update system packages', 'Restart services']
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.recommendations).toHaveLength(2);
      expect(result.recommendations[0]?.category).toBe('performance');
      expect(result.recommendations[1]?.category).toBe('security');
      expect(result.recommendations[1]?.severity).toBe(DiagnosticSeverity.ERROR);
    });

    test('should prioritize recommendations by severity and impact', async () => {
      // Arrange
      const options: DoctorOptions = { includeRecommendations: true };

      mockOptimizationRecommender.generateRecommendations.mockResolvedValue([
        {
          category: 'maintenance',
          severity: DiagnosticSeverity.INFO,
          title: 'Low Priority Task',
          description: 'Minor optimization available',
          impact: 'low',
          effort: 'easy',
          steps: ['Optional step']
        },
        {
          category: 'security',
          severity: DiagnosticSeverity.CRITICAL,
          title: 'Critical Security Issue',
          description: 'Immediate action required',
          impact: 'critical',
          effort: 'hard',
          steps: ['Emergency patch']
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      // Critical recommendations should be prioritized
      const criticalRecs = result.recommendations.filter(
        r => r.severity === DiagnosticSeverity.CRITICAL
      );
      expect(criticalRecs).toHaveLength(1);
      expect(result.summary.criticalIssues).toBe(1);
    });
  });

  describe('Environment Analysis', () => {
    test('should analyze development environment configurations', async () => {
      // Arrange
      const options: DoctorOptions = { includeEnvironmentAnalysis: true };

      mockEnvironmentAnalyzer.analyzeEnvironments.mockResolvedValue([
        {
          name: 'TypeScript',
          detected: true,
          version: '5.2.2',
          configFiles: ['tsconfig.json', '.eslintrc.json'],
          recommendations: [
            {
              category: 'compatibility',
              severity: DiagnosticSeverity.WARNING,
              title: 'TypeScript Configuration',
              description: 'Consider updating compiler options',
              impact: 'medium',
              effort: 'easy',
              steps: ['Update tsconfig.json']
            }
          ],
          status: 'good'
        },
        {
          name: 'Git',
          detected: true,
          version: '2.40.0',
          configFiles: ['.gitconfig', '.gitignore'],
          recommendations: [],
          status: 'optimal'
        }
      ]);

      // Act
      const result = await doctorCommand.runAdvanced(options);

      // Assert
      expect(result.environments).toHaveLength(2);
      expect(result.environments[0]?.name).toBe('TypeScript');
      expect(result.environments[0]?.status).toBe('good');
      expect(result.environments[1]?.status).toBe('optimal');
    });
  });

  describe('Output Formatting', () => {
    test('should support different output formats', async () => {
      // Arrange
      const textOptions: DoctorOptions = { outputFormat: 'text' };
      const jsonOptions: DoctorOptions = { outputFormat: 'json' };

      // Act
      const textResult = await doctorCommand.runAdvanced(textOptions);
      const jsonResult = await doctorCommand.runAdvanced(jsonOptions);

      // Assert
      expect(textResult).toBeDefined();
      expect(jsonResult).toBeDefined();
      // Both should have the same data structure
      expect(textResult.healthScore).toEqual(jsonResult.healthScore);
    });

    test('should provide verbose output when requested', async () => {
      // Arrange
      const options: DoctorOptions = { verbose: true };

      // Mock console.log to capture verbose output
      const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

      // Act
      await doctorCommand.runAdvanced(options);

      // Assert
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('🔍 Advanced MoAI-ADK System Diagnostics')
      );
      consoleSpy.mockRestore();
    });
  });
});