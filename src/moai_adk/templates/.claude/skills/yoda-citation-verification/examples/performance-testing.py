#!/usr/bin/env python3
"""
Performance Testing for YodA Citation Verification

This example demonstrates performance testing capabilities of the citation verification
system, including batch processing, concurrent verification, and caching efficiency.
"""

import asyncio
import time
import random
from typing import List, Dict, Any
from dataclasses import dataclass, asdict
import json

@dataclass
class PerformanceTestResult:
    """Results from performance testing"""
    test_name: str
    total_citations: int
    successful_verifications: int
    failed_verifications: int
    processing_time: float
    average_time_per_citation: float
    throughput: float  # citations per second
    cache_hit_rate: float = 0.0
    memory_usage_mb: float = 0.0

@dataclass
class CitationTestConfig:
    """Configuration for citation testing"""
    batch_size: int
    max_concurrent: int
    enable_cache: bool
    test_urls: List[str]

class CitationPerformanceTester:
    """Performance testing suite for citation verification"""

    def __init__(self):
        self.test_urls = self._generate_test_urls()
        self.results: List[PerformanceTestResult] = []

    def _generate_test_urls(self) -> List[str]:
        """Generate test URLs for performance testing"""
        return [
            "https://docs.anthropic.com/claude-code",
            "https://docs.anthropic.com/claude-code/getting-started",
            "https://docs.anthropic.com/claude-code/features",
            "https://docs.anthropic.com/claude-code/agents",
            "https://docs.anthropic.com/claude-code/mcp",
            "https://docs.python.org/3/",
            "https://docs.python.org/3/tutorial/index.html",
            "https://docs.python.org/3/library/index.html",
            "https://peps.python.org/",
            "https://react.dev/",
            "https://react.dev/learn",
            "https://react.dev/reference/react",
            "https://nextjs.org/docs",
            "https://nextjs.org/learn",
            "https://nodejs.org/docs",
            "https://nodejs.org/api",
            "https://www.typescriptlang.org/docs/",
            "https://www.typescriptlang.org/docs/handbook/intro.html",
            "https://developer.mozilla.org/",
            "https://docs.github.com/",
            "https://vercel.com/docs"
        ]

    async def run_performance_tests(self) -> Dict[str, Any]:
        """Run comprehensive performance tests"""
        print("🚀 Starting Citation Verification Performance Tests")
        print("=" * 60)

        test_configs = [
            # Single citation tests
            CitationTestConfig(batch_size=1, max_concurrent=1, enable_cache=False, test_urls=self.test_urls[:1]),
            CitationTestConfig(batch_size=1, max_concurrent=1, enable_cache=True, test_urls=self.test_urls[:1]),

            # Small batch tests
            CitationTestConfig(batch_size=5, max_concurrent=3, enable_cache=False, test_urls=self.test_urls[:5]),
            CitationTestConfig(batch_size=5, max_concurrent=3, enable_cache=True, test_urls=self.test_urls[:5]),

            # Medium batch tests
            CitationTestConfig(batch_size=10, max_concurrent=5, enable_cache=False, test_urls=self.test_urls[:10]),
            CitationTestConfig(batch_size=10, max_concurrent=5, enable_cache=True, test_urls=self.test_urls[:10]),

            # Large batch tests
            CitationTestConfig(batch_size=20, max_concurrent=10, enable_cache=False, test_urls=self.test_urls[:20]),
            CitationTestConfig(batch_size=20, max_concurrent=10, enable_cache=True, test_urls=self.test_urls),
        ]

        for i, config in enumerate(test_configs, 1):
            print(f"\n📊 Test {i}/{len(test_configs)}: Batch Size {config.batch_size}, "
                  f"Max Concurrent {config.max_concurrent}, Cache {'ON' if config.enable_cache else 'OFF'}")

            result = await self._run_single_test(config)
            self.results.append(result)

            self._print_test_result(result)

        # Generate comprehensive report
        report = self._generate_performance_report()
        self._save_results_to_file(report)

        return report

    async def _run_single_test(self, config: CitationTestConfig) -> PerformanceTestResult:
        """Run a single performance test"""
        start_time = time.time()

        # Simulate citation verification with concurrency
        semaphore = asyncio.Semaphore(config.max_concurrent)
        verification_tasks = []

        async def verify_single_citation(url: str) -> bool:
            async with semaphore:
                # Simulate verification time (0.1-1.0 seconds)
                verification_time = random.uniform(0.1, 1.0)
                await asyncio.sleep(verification_time)

                # Simulate success rate (85-95%)
                return random.random() > 0.1

        # Create verification tasks
        for url in config.test_urls[:config.batch_size]:
            task = asyncio.create_task(verify_single_citation(url))
            verification_tasks.append(task)

        # Wait for all verifications to complete
        verification_results = await asyncio.gather(*verification_tasks, return_exceptions=True)

        end_time = time.time()
        processing_time = end_time - start_time

        # Calculate metrics
        successful_count = sum(1 for result in verification_results if result is True)
        failed_count = len(verification_results) - successful_count

        # Simulate cache hit rate for cached tests
        cache_hit_rate = random.uniform(0.7, 0.9) if config.enable_cache else 0.0

        # Simulate memory usage
        memory_usage = config.batch_size * random.uniform(0.5, 2.0)  # MB per citation

        return PerformanceTestResult(
            test_name=f"Batch_{config.batch_size}_Concurrent_{config.max_concurrent}_Cache_{config.enable_cache}",
            total_citations=config.batch_size,
            successful_verifications=successful_count,
            failed_verifications=failed_count,
            processing_time=processing_time,
            average_time_per_citation=processing_time / config.batch_size,
            throughput=config.batch_size / processing_time,
            cache_hit_rate=cache_hit_rate,
            memory_usage_mb=memory_usage
        )

    def _print_test_result(self, result: PerformanceTestResult):
        """Print formatted test result"""
        print(f"✅ {result.test_name}")
        print(f"   📋 Total Citations: {result.total_citations}")
        print(f"   ✅ Successful: {result.successful_verifications}")
        print(f"   ❌ Failed: {result.failed_verifications}")
        print(f"   🕐 Processing Time: {result.processing_time:.2f}s")
        print(f"   ⚡ Average per Citation: {result.average_time_per_citation:.3f}s")
        print(f"   🚀 Throughput: {result.throughput:.2f} citations/sec")
        if result.cache_hit_rate > 0:
            print(f"   💾 Cache Hit Rate: {result.cache_hit_rate:.1%}")
        print(f"   🧠 Memory Usage: {result.memory_usage_mb:.1f} MB")

    def _generate_performance_report(self) -> Dict[str, Any]:
        """Generate comprehensive performance report"""
        if not self.results:
            return {"error": "No test results available"}

        # Calculate aggregate metrics
        total_citations = sum(r.total_citations for r in self.results)
        total_successful = sum(r.successful_verifications for r in self.results)
        total_failed = sum(r.failed_verifications for r in self.results)
        total_processing_time = sum(r.processing_time for r in self.results)

        # Find best and worst performance
        best_throughput = max(self.results, key=lambda r: r.throughput)
        worst_throughput = min(self.results, key=lambda r: r.throughput)
        fastest_single = min(self.results, key=lambda r: r.average_time_per_citation)
        slowest_single = max(self.results, key=lambda r: r.average_time_per_citation)

        # Cache performance analysis
        cached_results = [r for r in self.results if r.cache_hit_rate > 0]
        non_cached_results = [r for r in self.results if r.cache_hit_rate == 0]

        avg_cached_throughput = sum(r.throughput for r in cached_results) / len(cached_results) if cached_results else 0
        avg_non_cached_throughput = sum(r.throughput for r in non_cached_results) / len(non_cached_results) if non_cached_results else 0

        cache_performance_improvement = ((avg_cached_throughput - avg_non_cached_throughput) / avg_non_cached_throughput * 100) if avg_non_cached_throughput > 0 else 0

        # Scalability analysis
        batch_sizes = sorted(set(r.total_citations for r in self.results))
        scalability_data = []

        for batch_size in batch_sizes:
            batch_results = [r for r in self.results if r.total_citations == batch_size]
            if batch_results:
                avg_throughput = sum(r.throughput for r in batch_results) / len(batch_results)
                scalability_data.append({
                    "batch_size": batch_size,
                    "avg_throughput": avg_throughput
                })

        report = {
            "summary": {
                "total_tests_run": len(self.results),
                "total_citations_processed": total_citations,
                "overall_success_rate": (total_successful / total_citations * 100) if total_citations > 0 else 0,
                "total_processing_time": total_processing_time,
                "overall_throughput": total_citations / total_processing_time if total_processing_time > 0 else 0
            },
            "performance_highlights": {
                "best_throughput": {
                    "test": best_throughput.test_name,
                    "throughput": best_throughput.throughput,
                    "config": f"Batch: {best_throughput.total_citations}, Concurrent: {best_throughput.throughput:.2f}"
                },
                "worst_throughput": {
                    "test": worst_throughput.test_name,
                    "throughput": worst_throughput.throughput,
                    "config": f"Batch: {worst_throughput.total_citations}"
                },
                "fastest_single_citation": {
                    "test": fastest_single.test_name,
                    "time": fastest_single.average_time_per_citation
                },
                "slowest_single_citation": {
                    "test": slowest_single.test_name,
                    "time": slowest_single.average_time_per_citation
                }
            },
            "cache_analysis": {
                "cached_tests_count": len(cached_results),
                "non_cached_tests_count": len(non_cached_results),
                "avg_cached_throughput": avg_cached_throughput,
                "avg_non_cached_throughput": avg_non_cached_throughput,
                "cache_performance_improvement_percent": cache_performance_improvement
            },
            "scalability_analysis": scalability_data,
            "detailed_results": [asdict(result) for result in self.results],
            "recommendations": self._generate_recommendations()
        }

        return report

    def _generate_recommendations(self) -> List[str]:
        """Generate performance recommendations based on test results"""
        recommendations = []

        # Analyze cache performance
        cached_results = [r for r in self.results if r.cache_hit_rate > 0]
        non_cached_results = [r for r in self.results if r.cache_hit_rate == 0]

        if cached_results and non_cached_results:
            avg_cached_throughput = sum(r.throughput for r in cached_results) / len(cached_results)
            avg_non_cached_throughput = sum(r.throughput for r in non_cached_results) / len(non_cached_results)

            if avg_cached_throughput > avg_non_cached_throughput * 1.2:
                recommendations.append("🚀 캐싱을 활성화하여 20% 이상의 성능 향상을 얻을 수 있습니다")

        # Analyze batch size performance
        small_batches = [r for r in self.results if r.total_citations <= 5]
        large_batches = [r for r in self.results if r.total_citations >= 10]

        if small_batches and large_batches:
            avg_small_throughput = sum(r.throughput for r in small_batches) / len(small_batches)
            avg_large_throughput = sum(r.throughput for r in large_batches) / len(large_batches)

            if avg_large_throughput > avg_small_throughput:
                recommendations.append("📦 배치 처리를 사용하여 처리량을 최적화하세요")

        # Analyze concurrency performance
        concurrent_results = [r for r in self.results if "Concurrent_10" in r.test_name]
        if concurrent_results:
            avg_concurrent_throughput = sum(r.throughput for r in concurrent_results) / len(concurrent_results)
            if avg_concurrent_throughput > 5:  # citations per second
                recommendations.append("⚡ 높은 동시성 수준으로 처리량을 극대화하세요")

        # Memory optimization
        high_memory_results = [r for r in self.results if r.memory_usage_mb > 20]
        if high_memory_results:
            recommendations.append("🧠 메모리 사용량을 최적화하기 위해 더 작은 배치 크기를 고려하세요")

        if not recommendations:
            recommendations.append("✅ 현재 구성이 최적으로 조정되어 있습니다")

        return recommendations

    def _save_results_to_file(self, report: Dict[str, Any]):
        """Save performance results to file"""
        timestamp = time.strftime("%Y%m%d_%H%M%S")
        filename = f"citation_verification_performance_{timestamp}.json"

        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(report, f, indent=2, ensure_ascii=False)
            print(f"\n💾 Detailed results saved to: {filename}")
        except Exception as e:
            print(f"❌ Failed to save results: {e}")

    async def run_stress_test(self, duration_seconds: int = 60) -> Dict[str, Any]:
        """Run stress test for specified duration"""
        print(f"\n🔥 Running Stress Test for {duration_seconds} seconds")

        start_time = time.time()
        end_time = start_time + duration_seconds

        citation_count = 0
        successful_count = 0
        failed_count = 0
        peak_memory = 0

        while time.time() < end_time:
            # Simulate concurrent verification bursts
            batch_size = random.randint(1, 10)
            tasks = []

            for _ in range(batch_size):
                url = random.choice(self.test_urls)
                task = asyncio.create_task(self._stress_verify_single(url))
                tasks.append(task)

            results = await asyncio.gather(*tasks, return_exceptions=True)

            # Update counters
            for result in results:
                citation_count += 1
                if result is True:
                    successful_count += 1
                else:
                    failed_count += 1

            # Simulate memory usage tracking
            current_memory = random.uniform(5, 25)  # MB
            peak_memory = max(peak_memory, current_memory)

            # Small delay between bursts
            await asyncio.sleep(0.1)

        actual_duration = time.time() - start_time

        stress_report = {
            "test_type": "stress_test",
            "duration_seconds": actual_duration,
            "total_citations_processed": citation_count,
            "successful_verifications": successful_count,
            "failed_verifications": failed_count,
            "success_rate": (successful_count / citation_count * 100) if citation_count > 0 else 0,
            "throughput": citation_count / actual_duration,
            "peak_memory_usage_mb": peak_memory,
            "average_memory_usage_mb": peak_memory / 2  # Estimate
        }

        print(f"✅ Stress Test Complete:")
        print(f"   📊 Total Citations: {citation_count}")
        print(f"   ✅ Success Rate: {stress_report['success_rate']:.1f}%")
        print(f"   🚀 Throughput: {stress_report['throughput']:.2f} citations/sec")
        print(f"   🧠 Peak Memory: {peak_memory:.1f} MB")

        return stress_report

    async def _stress_verify_single(self, url: str) -> bool:
        """Single citation verification for stress testing"""
        # Simulate verification time
        await asyncio.sleep(random.uniform(0.05, 0.5))
        return random.random() > 0.05  # 95% success rate


async def main():
    """Main performance testing runner"""
    tester = CitationPerformanceTester()

    # Run comprehensive performance tests
    performance_report = await tester.run_performance_tests()

    # Run stress test
    stress_report = await tester.run_stress_test(duration_seconds=30)

    # Print summary
    print("\n" + "=" * 60)
    print("📊 PERFORMANCE TESTING SUMMARY")
    print("=" * 60)

    summary = performance_report["summary"]
    print(f"🧪 Total Tests: {summary['total_tests_run']}")
    print(f"📋 Citations Processed: {summary['total_citations_processed']}")
    print(f"✅ Overall Success Rate: {summary['overall_success_rate']:.1f}%")
    print(f"🚀 Overall Throughput: {summary['overall_throughput']:.2f} citations/sec")

    print("\n💡 RECOMMENDATIONS:")
    for rec in performance_report["recommendations"]:
        print(f"   {rec}")

    print(f"\n🔥 STRESS TEST RESULTS:")
    print(f"   📊 Citations: {stress_report['total_citations_processed']}")
    print(f"   🚀 Throughput: {stress_report['throughput']:.2f} citations/sec")
    print(f"   🧠 Peak Memory: {stress_report['peak_memory_usage_mb']:.1f} MB")


if __name__ == "__main__":
    asyncio.run(main())