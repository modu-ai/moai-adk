#!/usr/bin/env python3
"""
Claude Code Memory Optimizer

Optimizes memory usage through three-layer architecture:
- Working memory compression
- Long-term memory consolidation
- Cache optimization and cleanup

Usage: python memory-optimizer.py [--analyze] [--optimize] [--cleanup]
"""

import os
import json
import shutil
import argparse
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from datetime import datetime, timedelta


class MemoryOptimizer:
    """Optimize Claude Code memory usage and management."""

    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.moai_dir = self.project_root / ".moai"
        self.cache_dir = self.moai_dir / "cache"
        self.logs_dir = self.moai_dir / "logs"
        self.temp_dir = self.moai_dir / "temp"

    def analyze_memory_usage(self) -> Dict:
        """Analyze current memory usage across three layers."""
        analysis = {
            'timestamp': datetime.now().isoformat(),
            'working_memory': self.analyze_working_memory(),
            'long_term_memory': self.analyze_long_term_memory(),
            'cache_memory': self.analyze_cache_memory(),
            'recommendations': []
        }

        # Generate recommendations
        analysis['recommendations'] = self.generate_recommendations(analysis)

        return analysis

    def analyze_working_memory(self) -> Dict:
        """Analyze working memory usage (active context)."""
        working_dir = self.moai_dir / "working"

        if not working_dir.exists():
            return {'size_bytes': 0, 'file_count': 0, 'status': 'not_initialized'}

        total_size = 0
        file_count = 0
        file_types = {}

        for file_path in working_dir.rglob("*"):
            if file_path.is_file():
                size = file_path.stat().st_size
                total_size += size
                file_count += 1

                ext = file_path.suffix or 'no_ext'
                file_types[ext] = file_types.get(ext, 0) + 1

        return {
            'size_bytes': total_size,
            'size_mb': round(total_size / (1024 * 1024), 2),
            'file_count': file_count,
            'file_types': file_types,
            'status': 'active'
        }

    def analyze_long_term_memory(self) -> Dict:
        """Analyze long-term memory usage (persistent storage)."""
        memory_dirs = ['specs', 'docs', 'memory', 'backups']
        total_size = 0
        dir_analysis = {}

        for dir_name in memory_dirs:
            dir_path = self.moai_dir / dir_name
            if dir_path.exists():
                dir_size = self._get_directory_size(dir_path)
                total_size += dir_size
                dir_analysis[dir_name] = {
                    'size_bytes': dir_size,
                    'size_mb': round(dir_size / (1024 * 1024), 2),
                    'file_count': len(list(dir_path.rglob("*")))
                }

        return {
            'total_size_bytes': total_size,
            'total_size_mb': round(total_size / (1024 * 1024), 2),
            'directories': dir_analysis,
            'status': 'persistent'
        }

    def analyze_cache_memory(self) -> Dict:
        """Analyze cache memory usage."""
        if not self.cache_dir.exists():
            return {'size_bytes': 0, 'file_count': 0, 'status': 'no_cache'}

        total_size = 0
        file_count = 0
        old_files = []
        cache_types = {}

        cutoff_time = datetime.now() - timedelta(days=7)

        for file_path in self.cache_dir.rglob("*"):
            if file_path.is_file():
                size = file_path.stat().st_size
                total_size += size
                file_count += 1

                # Check for old files
                mtime = datetime.fromtimestamp(file_path.stat().st_mtime)
                if mtime < cutoff_time:
                    old_files.append({
                        'path': str(file_path.relative_to(self.cache_dir)),
                        'size_bytes': size,
                        'age_days': (datetime.now() - mtime).days
                    })

                # Classify by type
                parent_dir = file_path.parent.name
                cache_types[parent_dir] = cache_types.get(parent_dir, 0) + 1

        return {
            'size_bytes': total_size,
            'size_mb': round(total_size / (1024 * 1024), 2),
            'file_count': file_count,
            'old_files_count': len(old_files),
            'old_files_size_mb': round(sum(f['size_bytes'] for f in old_files) / (1024 * 1024), 2),
            'cache_types': cache_types,
            'status': 'active'
        }

    def optimize_memory(self, aggressive: bool = False) -> Dict:
        """Optimize memory usage across all layers."""
        optimization_result = {
            'timestamp': datetime.now().isoformat(),
            'actions_taken': [],
            'space_freed_mb': 0,
            'errors': []
        }

        try:
            # Clean up old cache files
            cache_result = self.cleanup_cache(aggressive)
            optimization_result['actions_taken'].extend(cache_result['actions'])
            optimization_result['space_freed_mb'] += cache_result['space_freed_mb']

            # Consolidate long-term memory
            ltm_result = self.consolidate_long_term_memory()
            optimization_result['actions_taken'].extend(ltm_result['actions'])
            optimization_result['space_freed_mb'] += ltm_result['space_freed_mb']

            # Optimize working memory
            wm_result = self.optimize_working_memory()
            optimization_result['actions_taken'].extend(wm_result['actions'])
            optimization_result['space_freed_mb'] += wm_result['space_freed_mb']

        except Exception as e:
            optimization_result['errors'].append(f"Optimization failed: {e}")

        return optimization_result

    def cleanup_cache(self, aggressive: bool = False) -> Dict:
        """Clean up cache files."""
        result = {
            'actions': [],
            'space_freed_mb': 0,
            'files_removed': 0
        }

        if not self.cache_dir.exists():
            return result

        cutoff_days = 1 if aggressive else 7
        cutoff_time = datetime.now() - timedelta(days=cutoff_days)

        for file_path in self.cache_dir.rglob("*"):
            if file_path.is_file():
                mtime = datetime.fromtimestamp(file_path.stat().st_mtime)

                if mtime < cutoff_time or aggressive:
                    try:
                        size = file_path.stat().st_size
                        file_path.unlink()
                        result['space_freed_mb'] += size / (1024 * 1024)
                        result['files_removed'] += 1
                        result['actions'].append(f"Removed cache file: {file_path.name}")
                    except Exception as e:
                        result['actions'].append(f"Failed to remove {file_path.name}: {e}")

        return result

    def consolidate_long_term_memory(self) -> Dict:
        """Consolidate and compress long-term memory."""
        result = {
            'actions': [],
            'space_freed_mb': 0,
            'files_processed': 0
        }

        # Archive old specs
        specs_dir = self.moai_dir / "specs"
        archive_dir = self.moai_dir / "backups" / "specs"

        if specs_dir.exists():
            cutoff_time = datetime.now() - timedelta(days=90)

            for spec_dir in specs_dir.iterdir():
                if spec_dir.is_dir():
                    # Check if spec is old (based on file modification times)
                    spec_files = list(spec_dir.rglob("*"))
                    if spec_files and all(
                        datetime.fromtimestamp(f.stat().st_mtime) < cutoff_time
                        for f in spec_files if f.is_file()
                    ):
                        try:
                            archive_dir.mkdir(parents=True, exist_ok=True)
                            archive_path = archive_dir / spec_dir.name

                            if not archive_path.exists():
                                shutil.move(str(spec_dir), str(archive_path))
                                size = sum(f.stat().st_size for f in spec_files.rglob("*") if f.is_file())
                                result['space_freed_mb'] += size / (1024 * 1024)
                                result['files_processed'] += len(spec_files)
                                result['actions'].append(f"Archived old spec: {spec_dir.name}")
                        except Exception as e:
                            result['actions'].append(f"Failed to archive {spec_dir.name}: {e}")

        return result

    def optimize_working_memory(self) -> Dict:
        """Optimize working memory (compression and cleanup)."""
        result = {
            'actions': [],
            'space_freed_mb': 0,
            'files_processed': 0
        }

        working_dir = self.moai_dir / "working"
        if not working_dir.exists():
            return result

        # Remove temporary working files
        for file_path in working_dir.rglob("*"):
            if file_path.is_file():
                # Remove temporary files
                if (file_path.name.startswith('.') or
                    file_path.name.endswith('.tmp') or
                    file_path.name.endswith('.temp')):
                    try:
                        size = file_path.stat().st_size
                        file_path.unlink()
                        result['space_freed_mb'] += size / (1024 * 1024)
                        result['files_processed'] += 1
                        result['actions'].append(f"Removed temp file: {file_path.name}")
                    except Exception as e:
                        result['actions'].append(f"Failed to remove {file_path.name}: {e}")

        return result

    def generate_recommendations(self, analysis: Dict) -> List[str]:
        """Generate optimization recommendations."""
        recommendations = []

        # Working memory recommendations
        wm = analysis['working_memory']
        if wm.get('size_mb', 0) > 50:  # 50MB threshold
            recommendations.append(f"Working memory is large ({wm['size_mb']}MB). Consider compressing session data.")

        # Cache recommendations
        cache = analysis['cache_memory']
        if cache.get('old_files_size_mb', 0) > 10:
            recommendations.append(f"Cache has {cache['old_files_size_mb']}MB of old files. Run cleanup.")

        # Long-term memory recommendations
        ltm = analysis['long_term_memory']
        if ltm.get('total_size_mb', 0) > 500:
            recommendations.append(f"Long-term memory is large ({ltm['total_size_mb']}MB). Consider archiving old data.")

        return recommendations

    def _get_directory_size(self, directory: Path) -> int:
        """Get total size of directory in bytes."""
        total_size = 0
        for file_path in directory.rglob("*"):
            if file_path.is_file():
                total_size += file_path.stat().st_size
        return total_size

    def print_analysis(self, analysis: Dict) -> None:
        """Print memory analysis results."""
        print("\n" + "="*50)
        print("CLAUDE CODE MEMORY ANALYSIS")
        print("="*50)

        # Working Memory
        wm = analysis['working_memory']
        print(f"\n📱 Working Memory: {wm.get('size_mb', 0)}MB ({wm.get('file_count', 0)} files)")
        if wm.get('file_types'):
            print("  File types:")
            for ext, count in wm['file_types'].items():
                print(f"    {ext}: {count}")

        # Long-term Memory
        ltm = analysis['long_term_memory']
        print(f"\n💾 Long-term Memory: {ltm.get('total_size_mb', 0)}MB")
        if ltm.get('directories'):
            print("  Directories:")
            for name, info in ltm['directories'].items():
                print(f"    {name}/: {info['size_mb']}MB ({info['file_count']} files)")

        # Cache Memory
        cache = analysis['cache_memory']
        print(f"\n⚡ Cache Memory: {cache.get('size_mb', 0)}MB ({cache.get('file_count', 0)} files)")
        if cache.get('old_files_count', 0) > 0:
            print(f"  Old files: {cache['old_files_count']} ({cache['old_files_size_mb']}MB)")

        # Recommendations
        if analysis['recommendations']:
            print(f"\n💡 Recommendations:")
            for rec in analysis['recommendations']:
                print(f"  • {rec}")

    def print_optimization_result(self, result: Dict) -> None:
        """Print optimization results."""
        print("\n" + "="*50)
        print("MEMORY OPTIMIZATION RESULTS")
        print("="*50)

        print(f"\n📊 Summary:")
        print(f"  Actions taken: {len(result['actions_taken'])}")
        print(f"  Space freed: {result['space_freed_mb']:.2f}MB")

        if result['actions_taken']:
            print(f"\n✅ Actions completed:")
            for action in result['actions_taken'][:10]:  # Show first 10
                print(f"  • {action}")
            if len(result['actions_taken']) > 10:
                print(f"  ... and {len(result['actions_taken']) - 10} more actions")

        if result['errors']:
            print(f"\n❌ Errors:")
            for error in result['errors']:
                print(f"  • {error}")


def main():
    parser = argparse.ArgumentParser(description="Optimize Claude Code memory usage")
    parser.add_argument("--analyze", action="store_true", help="Analyze current memory usage")
    parser.add_argument("--optimize", action="store_true", help="Optimize memory usage")
    parser.add_argument("--aggressive", action="store_true", help="Aggressive cleanup (1-day cache retention)")
    parser.add_argument("--cleanup", action="store_true", help="Clean up old files only")
    parser.add_argument("--project-root", default=".", help="Project root directory")

    args = parser.parse_args()

    optimizer = MemoryOptimizer(args.project_root)

    if args.analyze:
        analysis = optimizer.analyze_memory_usage()
        optimizer.print_analysis(analysis)

        # Save analysis to file
        analysis_file = Path(args.project_root) / ".moai" / "logs" / f"memory-analysis-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        analysis_file.parent.mkdir(exist_ok=True)
        with open(analysis_file, 'w') as f:
            json.dump(analysis, f, indent=2)
        print(f"\n📄 Analysis saved to: {analysis_file}")

    elif args.optimize or args.cleanup:
        result = optimizer.optimize_memory(aggressive=args.aggressive)
        optimizer.print_optimization_result(result)

        # Save result to file
        result_file = Path(args.project_root) / ".moai" / "logs" / f"memory-optimization-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        result_file.parent.mkdir(exist_ok=True)
        with open(result_file, 'w') as f:
            json.dump(result, f, indent=2)
        print(f"\n📄 Optimization result saved to: {result_file}")

    else:
        # Default: analyze and show recommendations
        analysis = optimizer.analyze_memory_usage()
        optimizer.print_analysis(analysis)


if __name__ == "__main__":
    main()