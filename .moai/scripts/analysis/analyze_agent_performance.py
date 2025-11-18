#!/usr/bin/env python3
"""
Agent Performance Analysis Tool

Analyzes hook execution metrics from agent-performance.jsonl log file.
Provides statistics on:
- Hook execution times (mean, min, max, percentiles)
- Success rates by agent
- Cost calculations based on model usage
- Performance trends

Usage:
    uv run .moai/scripts/analysis/analyze_agent_performance.py
"""

import json
import sys
from pathlib import Path
from collections import defaultdict
from typing import Dict, List, Any
from datetime import datetime


def load_performance_logs(log_file: Path) -> List[Dict[str, Any]]:
    """Load JSONL performance log file"""
    records = []
    if not log_file.exists():
        print(f"⚠️ Log file not found: {log_file}")
        return records

    with open(log_file, 'r') as f:
        for line_num, line in enumerate(f, 1):
            if line.strip():
                try:
                    records.append(json.loads(line))
                except json.JSONDecodeError as e:
                    print(f"⚠️ Invalid JSON at line {line_num}: {e}")

    return records


def analyze_execution_times(records: List[Dict]) -> Dict[str, Any]:
    """Analyze execution time statistics"""
    if not records:
        return {}

    times = [r.get('execution_time_ms', 0) for r in records if 'execution_time_ms' in r]
    if not times:
        return {}

    times.sort()
    mean = sum(times) / len(times)

    return {
        'count': len(times),
        'mean_ms': round(mean, 2),
        'mean_s': round(mean / 1000, 2),
        'min_ms': min(times),
        'max_ms': max(times),
        'p50_ms': times[len(times) // 2],
        'p95_ms': times[int(len(times) * 0.95)],
        'p99_ms': times[int(len(times) * 0.99)],
    }


def analyze_by_agent(records: List[Dict]) -> Dict[str, Any]:
    """Analyze metrics by agent"""
    agents = defaultdict(list)

    for record in records:
        agent_name = record.get('agent_name', 'unknown')
        agents[agent_name].append(record)

    results = {}
    for agent_name, agent_records in sorted(agents.items()):
        times = [r.get('execution_time_ms', 0) for r in agent_records]
        success_count = sum(1 for r in agent_records if r.get('success', False))

        results[agent_name] = {
            'total_executions': len(agent_records),
            'successful': success_count,
            'failed': len(agent_records) - success_count,
            'success_rate': round(100 * success_count / len(agent_records), 1) if agent_records else 0,
            'avg_time_ms': round(sum(times) / len(times), 1) if times else 0,
            'min_time_ms': min(times) if times else 0,
            'max_time_ms': max(times) if times else 0,
        }

    return results


def analyze_success_rate(records: List[Dict]) -> Dict[str, Any]:
    """Analyze overall success rates"""
    if not records:
        return {}

    successful = sum(1 for r in records if r.get('success', False))
    total = len(records)

    return {
        'total_records': total,
        'successful': successful,
        'failed': total - successful,
        'success_rate': round(100 * successful / total, 1) if total else 0,
    }


def calculate_costs(records: List[Dict]) -> Dict[str, Any]:
    """Calculate estimated costs based on model usage"""
    # Cost per 1K tokens (typical hook: 2-5K tokens)
    haiku_cost_per_1k = 0.0008
    sonnet_cost_per_1k = 0.003

    # Assume average hook tokens based on execution time
    # ~100ms = ~2K tokens, ~200ms = ~4K tokens, etc
    # Linear approximation: tokens = execution_time_ms / 50

    total_cost_all_sonnet = 0
    total_cost_mixed = 0

    haiku_hooks = ['SessionStart', 'PreToolUse', 'SessionEnd', 'SubagentStart', 'SubagentStop']
    sonnet_hooks = ['UserPromptSubmit']

    # Estimate tokens from execution time
    for record in records:
        exec_time_ms = record.get('execution_time_ms', 0)
        estimated_tokens = max(2000, exec_time_ms / 50)  # At least 2K

        cost_sonnet = (estimated_tokens / 1000) * sonnet_cost_per_1k
        total_cost_all_sonnet += cost_sonnet
        total_cost_mixed += cost_sonnet  # Default to Sonnet

    # Recalculate for mixed model approach
    total_cost_mixed = 0
    for record in records:
        exec_time_ms = record.get('execution_time_ms', 0)
        estimated_tokens = max(2000, exec_time_ms / 50)

        # Cost depends on hook type (simplified estimation)
        # Assume 85% of hooks are Haiku, 15% are Sonnet
        cost = (estimated_tokens / 1000) * (0.85 * haiku_cost_per_1k + 0.15 * sonnet_cost_per_1k)
        total_cost_mixed += cost

    savings = total_cost_all_sonnet - total_cost_mixed
    savings_pct = (savings / total_cost_all_sonnet * 100) if total_cost_all_sonnet > 0 else 0

    return {
        'cost_if_all_sonnet': round(total_cost_all_sonnet, 4),
        'cost_with_haiku_optimization': round(total_cost_mixed, 4),
        'estimated_savings': round(savings, 4),
        'savings_percentage': round(savings_pct, 1),
    }


def print_analysis(records: List[Dict]):
    """Print formatted analysis report"""
    if not records:
        print("No performance records found")
        return

    print("\n" + "=" * 80)
    print("AGENT PERFORMANCE ANALYSIS REPORT")
    print("=" * 80)

    # Summary
    print(f"\nRecords analyzed: {len(records)}")

    # Success rate
    success_analysis = analyze_success_rate(records)
    print(f"\nSUCCESS RATE")
    print(f"  Total executions: {success_analysis['total_records']}")
    print(f"  Successful: {success_analysis['successful']} ({success_analysis['success_rate']}%)")
    print(f"  Failed: {success_analysis['failed']}")

    # Execution times
    time_analysis = analyze_execution_times(records)
    if time_analysis:
        print(f"\nEXECUTION TIME STATISTICS")
        print(f"  Mean: {time_analysis['mean_ms']}ms ({time_analysis['mean_s']}s)")
        print(f"  Min: {time_analysis['min_ms']}ms")
        print(f"  Max: {time_analysis['max_ms']}ms")
        print(f"  P50: {time_analysis['p50_ms']}ms")
        print(f"  P95: {time_analysis['p95_ms']}ms")
        print(f"  P99: {time_analysis['p99_ms']}ms")

    # By agent
    agent_analysis = analyze_by_agent(records)
    if agent_analysis:
        print(f"\nPER-AGENT STATISTICS")
        print(f"  Agent                          Executions  Success %  Avg Time")
        print(f"  {'─' * 76}")
        for agent_name, metrics in agent_analysis.items():
            print(f"  {agent_name:<28} {metrics['total_executions']:>10}  "
                  f"{metrics['success_rate']:>8.1f}%  {metrics['avg_time_ms']:>8.1f}ms")

    # Cost savings
    cost_analysis = calculate_costs(records)
    if cost_analysis:
        print(f"\nCOST ANALYSIS (ESTIMATED)")
        print(f"  All Sonnet: ${cost_analysis['cost_if_all_sonnet']}")
        print(f"  Mixed (Haiku+Sonnet): ${cost_analysis['cost_with_haiku_optimization']}")
        print(f"  Estimated savings: ${cost_analysis['estimated_savings']} "
              f"({cost_analysis['savings_percentage']}%)")

    print("\n" + "=" * 80)


def main():
    """Main entry point"""
    project_root = Path("/Users/goos/MoAI/MoAI-ADK")
    log_file = project_root / ".moai" / "logs" / "agent-performance.jsonl"

    records = load_performance_logs(log_file)
    print_analysis(records)


if __name__ == "__main__":
    main()
