/**
 * @file Performance benchmark runner
 * @author MoAI Team
 * @tags @FEATURE:BENCHMARK-RUNNER-001 @REQ:ADVANCED-DOCTOR-001
 */

import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';
import type { BenchmarkResult, BenchmarkConfig } from '@/types/diagnostics';

/**
 * Performance benchmark runner for system testing
 * @tags @FEATURE:BENCHMARK-RUNNER-001
 */
export class BenchmarkRunner {
  private readonly benchmarks: BenchmarkConfig[];

  constructor() {
    this.benchmarks = this.getDefaultBenchmarks();
  }

  /**
   * Run all available benchmarks
   * @returns Array of benchmark results
   * @tags @API:RUN-ALL-BENCHMARKS-001
   */
  public async runAllBenchmarks(): Promise<BenchmarkResult[]> {
    const results: BenchmarkResult[] = [];

    for (const benchmark of this.benchmarks) {
      try {
        const result = await this.runSingleBenchmark(benchmark);
        results.push(result);
      } catch (error) {
        results.push({
          name: benchmark.name,
          duration: benchmark.timeout,
          status: 'fail',
          baseline: benchmark.baseline,
          score: 0,
          recommendations: [
            `Benchmark failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
          ],
        });
      }
    }

    return results;
  }

  /**
   * Run specific benchmark by name
   * @param benchmarkName - Name of benchmark to run
   * @returns Benchmark result
   * @tags @API:RUN-SPECIFIC-BENCHMARK-001
   */
  public async runSpecificBenchmark(
    benchmarkName: string
  ): Promise<BenchmarkResult | null> {
    const benchmark = this.benchmarks.find(b => b.name === benchmarkName);
    if (!benchmark) {
      return null;
    }

    return this.runSingleBenchmark(benchmark);
  }

  /**
   * Run single benchmark test
   * @param benchmark - Benchmark configuration
   * @returns Benchmark result
   * @tags @API:RUN-SINGLE-BENCHMARK-001
   */
  private async runSingleBenchmark(
    benchmark: BenchmarkConfig
  ): Promise<BenchmarkResult> {
    const startTime = Date.now();

    // Run the benchmark with timeout
    const duration = await Promise.race([
      benchmark.testFunction(),
      this.createTimeoutPromise(benchmark.timeout),
    ]);

    const actualDuration = Date.now() - startTime;

    // Calculate score based on performance vs baseline
    const score = this.calculateScore(duration, benchmark.baseline);
    const status = this.determineStatus(score, duration, benchmark.baseline);
    const recommendations = this.generateRecommendations(
      benchmark,
      score,
      duration
    );

    return {
      name: benchmark.name,
      duration: Math.round(duration),
      status,
      baseline: benchmark.baseline,
      score,
      recommendations,
    };
  }

  /**
   * Calculate performance score
   * @param duration - Actual duration
   * @param baseline - Expected baseline
   * @returns Score (0-100)
   * @tags @UTIL:CALCULATE-SCORE-001
   */
  private calculateScore(duration: number, baseline: number): number {
    if (duration <= baseline * 0.5) return 100; // Excellent
    if (duration <= baseline * 0.75) return 90; // Very good
    if (duration <= baseline) return 80; // Good
    if (duration <= baseline * 1.25) return 70; // Fair
    if (duration <= baseline * 1.5) return 60; // Below average
    if (duration <= baseline * 2) return 40; // Poor
    return 20; // Very poor
  }

  /**
   * Determine benchmark status
   * @param score - Performance score
   * @param duration - Actual duration
   * @param baseline - Expected baseline
   * @returns Status
   * @tags @UTIL:DETERMINE-STATUS-001
   */
  private determineStatus(
    score: number,
    duration: number,
    baseline: number
  ): 'pass' | 'warning' | 'fail' {
    if (score >= 80) return 'pass';
    if (score >= 60 || duration <= baseline * 1.5) return 'warning';
    return 'fail';
  }

  /**
   * Generate recommendations based on benchmark results
   * @param benchmark - Benchmark configuration
   * @param score - Performance score
   * @param duration - Actual duration
   * @returns Array of recommendations
   * @tags @UTIL:GENERATE-RECOMMENDATIONS-001
   */
  private generateRecommendations(
    benchmark: BenchmarkConfig,
    score: number,
    duration: number
  ): string[] {
    const recommendations: string[] = [];

    if (score < 60) {
      switch (benchmark.category) {
        case 'io':
          recommendations.push(
            'Consider using faster storage (SSD)',
            'Check for disk space availability',
            'Close unnecessary applications accessing disk'
          );
          break;
        case 'cpu':
          recommendations.push(
            'Close CPU-intensive background processes',
            'Consider upgrading CPU',
            'Check for thermal throttling'
          );
          break;
        case 'memory':
          recommendations.push(
            'Close memory-intensive applications',
            'Consider adding more RAM',
            'Check for memory leaks'
          );
          break;
        case 'network':
          recommendations.push(
            'Check network connection stability',
            'Consider using wired connection',
            'Check for bandwidth limitations'
          );
          break;
      }
    }

    if (score < 40) {
      recommendations.push('Performance is significantly below expectations');
    }

    return recommendations;
  }

  /**
   * Create timeout promise
   * @param timeout - Timeout in milliseconds
   * @returns Promise that rejects after timeout
   * @tags @UTIL:TIMEOUT-PROMISE-001
   */
  private createTimeoutPromise(timeout: number): Promise<number> {
    return new Promise((_, reject) => {
      setTimeout(() => reject(new Error('Benchmark timeout')), timeout);
    });
  }

  /**
   * Get default benchmark configurations
   * @returns Array of default benchmarks
   * @tags @UTIL:DEFAULT-BENCHMARKS-001
   */
  private getDefaultBenchmarks(): BenchmarkConfig[] {
    return [
      {
        name: 'File I/O Performance',
        description: 'Tests file read/write performance',
        category: 'io',
        timeout: 30000,
        baseline: 150,
        testFunction: this.fileIOBenchmark.bind(this),
      },
      {
        name: 'CPU Performance',
        description: 'Tests CPU computational performance',
        category: 'cpu',
        timeout: 15000,
        baseline: 100,
        testFunction: this.cpuBenchmark.bind(this),
      },
      {
        name: 'Memory Performance',
        description: 'Tests memory allocation/deallocation performance',
        category: 'memory',
        timeout: 10000,
        baseline: 75,
        testFunction: this.memoryBenchmark.bind(this),
      },
      {
        name: 'Network Connectivity',
        description: 'Tests network latency and connectivity',
        category: 'network',
        timeout: 20000,
        baseline: 200,
        testFunction: this.networkBenchmark.bind(this),
      },
    ];
  }

  /**
   * File I/O benchmark test
   * @returns Duration in milliseconds
   * @tags @BENCHMARK:FILE-IO-001
   */
  private async fileIOBenchmark(): Promise<number> {
    const start = Date.now();
    const tempDir = os.tmpdir();
    const testFile = path.join(tempDir, `moai-benchmark-${Date.now()}.txt`);
    const testData = 'x'.repeat(1024 * 1024); // 1MB of data

    try {
      // Write test
      await fs.writeFile(testFile, testData);

      // Read test
      const readData = await fs.readFile(testFile, 'utf8');

      // Verify data integrity
      if (readData.length !== testData.length) {
        throw new Error('Data integrity check failed');
      }

      // Cleanup
      await fs.unlink(testFile);

      return Date.now() - start;
    } catch (error) {
      // Cleanup on error
      try {
        await fs.unlink(testFile);
      } catch {
        // Ignore cleanup errors
      }
      throw error;
    }
  }

  /**
   * CPU benchmark test
   * @returns Duration in milliseconds
   * @tags @BENCHMARK:CPU-001
   */
  private async cpuBenchmark(): Promise<number> {
    const start = Date.now();

    // CPU-intensive calculation (fibonacci)
    const fibonacci = (n: number): number => {
      if (n <= 1) return n;
      return fibonacci(n - 1) + fibonacci(n - 2);
    };

    // Calculate fibonacci(35) - should take some time
    const result = fibonacci(35);

    // Verify result
    if (result !== 9227465) {
      throw new Error('CPU benchmark calculation failed');
    }

    return Date.now() - start;
  }

  /**
   * Memory benchmark test
   * @returns Duration in milliseconds
   * @tags @BENCHMARK:MEMORY-001
   */
  private async memoryBenchmark(): Promise<number> {
    const start = Date.now();

    // Memory allocation test
    const arrays: number[][] = [];
    const iterations = 1000;
    const arraySize = 1000;

    try {
      // Allocate arrays
      for (let i = 0; i < iterations; i++) {
        const arr = new Array(arraySize);
        for (let j = 0; j < arraySize; j++) {
          arr[j] = Math.random();
        }
        arrays.push(arr);
      }

      // Process arrays (memory access pattern test)
      let sum = 0;
      for (const arr of arrays) {
        for (const value of arr) {
          sum += value;
        }
      }

      // Verify we processed data
      if (sum === 0) {
        throw new Error('Memory benchmark processing failed');
      }

      return Date.now() - start;
    } finally {
      // Cleanup memory
      arrays.length = 0;
    }
  }

  /**
   * Network benchmark test
   * @returns Duration in milliseconds
   * @tags @BENCHMARK:NETWORK-001
   */
  private async networkBenchmark(): Promise<number> {
    const start = Date.now();

    try {
      // Test multiple network requests
      const urls = [
        'https://httpbin.org/delay/0',
        'https://httpbin.org/get',
        'https://httpbin.org/user-agent',
      ];

      const requests = urls.map(url =>
        fetch(url, { method: 'HEAD', cache: 'no-cache' })
      );

      const responses = await Promise.all(requests);

      // Verify all requests succeeded
      const allSuccessful = responses.every(response => response.ok);
      if (!allSuccessful) {
        throw new Error('Some network requests failed');
      }

      return Date.now() - start;
    } catch (error) {
      // Network might be unavailable
      throw new Error(
        `Network benchmark failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }
}
