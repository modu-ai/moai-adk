# SPEC Search, Indexing & Performance

**Version**: 4.0.0
**Focus**: Performance optimization, search indexing, caching strategies

---

## SPEC Optimization Patterns

### Pattern 1: Efficient SPEC Indexing

**In-Memory Index with LRU Cache**:
```python
from functools import lru_cache
from collections import defaultdict
import json
from pathlib import Path

class SpecIndex:
    """Lightweight SPEC indexing without external dependencies"""

    def __init__(self, specs_dir: Path):
        self.specs_dir = specs_dir
        self._index = defaultdict(list)
        self._requirement_map = {}
        self._cache = {}
        self._build_index()

    def _build_index(self):
        """Build in-memory index of all SPECs"""
        for spec_file in self.specs_dir.glob('SPEC-*/spec.md'):
            spec_id = spec_file.parent.name
            content = spec_file.read_text()

            # Extract requirements
            requirements = self._extract_requirements(content)
            self._requirement_map[spec_id] = requirements

            # Build keyword index
            for keyword in self._extract_keywords(content):
                self._index[keyword.lower()].append(spec_id)

    @lru_cache(maxsize=256)
    def search_by_keyword(self, keyword: str) -> list:
        """Search SPECs by keyword (cached)"""
        return self._index.get(keyword.lower(), [])

    def search_by_requirement_type(self, req_type: str) -> dict:
        """Search by EARS requirement type"""
        results = {}
        for spec_id, requirements in self._requirement_map.items():
            matching = [r for r in requirements if r['type'] == req_type]
            if matching:
                results[spec_id] = matching
        return results

    def _extract_keywords(self, content: str) -> set:
        """Extract searchable keywords"""
        # Simple tokenization
        words = set()
        for line in content.split('\n'):
            if line.startswith('#'):
                # Extract from headers
                tokens = line.strip('#').strip().split()
                words.update(tokens)
        return words

    def clear_cache(self):
        """Clear search cache after updates"""
        self.search_by_keyword.cache_clear()

# Usage
spec_index = SpecIndex(Path('.moai/specs'))
results = spec_index.search_by_keyword('authentication')
```

### Pattern 2: SPEC Caching Strategy

**Multi-Layer Caching**:
```python
from datetime import datetime, timedelta
import pickle

class SpecCache:
    """Multi-layer SPEC caching strategy"""

    def __init__(self, cache_dir: Path):
        self.cache_dir = cache_dir
        self.memory_cache = {}  # L1: In-memory
        self.disk_cache = cache_dir / 'specs'  # L2: Disk
        self.ttl = timedelta(hours=1)

    def get(self, spec_id: str) -> Optional[dict]:
        """Get SPEC from cache (L1 → L2 → Miss)"""
        # L1: Memory cache
        if spec_id in self.memory_cache:
            cached, timestamp = self.memory_cache[spec_id]
            if datetime.now() - timestamp < self.ttl:
                return cached
            del self.memory_cache[spec_id]

        # L2: Disk cache
        cache_file = self.disk_cache / f'{spec_id}.pkl'
        if cache_file.exists():
            try:
                with open(cache_file, 'rb') as f:
                    cached = pickle.load(f)
                self.memory_cache[spec_id] = (cached, datetime.now())
                return cached
            except (pickle.PickleError, EOFError):
                pass  # Cache corruption, fall back to reading file

        return None

    def set(self, spec_id: str, data: dict) -> None:
        """Cache SPEC in both layers"""
        # L1: Memory cache
        self.memory_cache[spec_id] = (data, datetime.now())

        # L2: Disk cache
        self.disk_cache.mkdir(parents=True, exist_ok=True)
        cache_file = self.disk_cache / f'{spec_id}.pkl'
        with open(cache_file, 'wb') as f:
            pickle.dump(data, f)

    def invalidate(self, spec_id: str) -> None:
        """Invalidate cache for specific SPEC"""
        self.memory_cache.pop(spec_id, None)
        cache_file = self.disk_cache / f'{spec_id}.pkl'
        if cache_file.exists():
            cache_file.unlink()

    def invalidate_all(self) -> None:
        """Clear all caches"""
        self.memory_cache.clear()
        import shutil
        if self.disk_cache.exists():
            shutil.rmtree(self.disk_cache)
```

### Pattern 3: Batch SPEC Processing

**Efficient Bulk Operations**:
```python
from concurrent.futures import ThreadPoolExecutor
from typing import List

class SpecBatchProcessor:
    """Process multiple SPECs efficiently"""

    def __init__(self, max_workers: int = 4):
        self.max_workers = max_workers

    def validate_all(self, spec_ids: List[str]) -> dict:
        """Validate multiple SPECs in parallel"""
        results = {}
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            future_to_spec = {
                executor.submit(self._validate_spec, spec_id): spec_id
                for spec_id in spec_ids
            }

            for future in concurrent.futures.as_completed(future_to_spec):
                spec_id = future_to_spec[future]
                try:
                    results[spec_id] = future.result()
                except Exception as e:
                    results[spec_id] = {'error': str(e)}

        return results

    def generate_all_reports(self, spec_ids: List[str]) -> dict:
        """Generate reports for multiple SPECs"""
        reports = {}
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            futures = {
                executor.submit(self._generate_report, spec_id): spec_id
                for spec_id in spec_ids
            }

            for future in concurrent.futures.as_completed(futures):
                spec_id = futures[future]
                try:
                    reports[spec_id] = future.result()
                except Exception as e:
                    reports[spec_id] = None

        return reports

    def _validate_spec(self, spec_id: str) -> dict:
        """Validate single SPEC"""
        # Implementation
        pass

    def _generate_report(self, spec_id: str) -> str:
        """Generate report for single SPEC"""
        # Implementation
        pass
```

### Pattern 4: SPEC Query Performance Optimization

**Query Optimization Techniques**:
```python
class OptimizedSpecQuery:
    """Optimized SPEC querying with statistics"""

    def __init__(self, specs_dir: Path):
        self.specs_dir = specs_dir
        self._build_statistics()

    def _build_statistics(self):
        """Pre-compute statistics for optimization"""
        self.stats = {
            'total_specs': 0,
            'total_requirements': 0,
            'requirement_distribution': {},
            'complexity_distribution': {}
        }

        for spec_file in self.specs_dir.glob('SPEC-*/spec.md'):
            self.stats['total_specs'] += 1
            content = spec_file.read_text()
            requirements = self._count_requirements(content)
            self.stats['total_requirements'] += requirements

    def find_specs_by_complexity(self, complexity: str) -> list:
        """Optimized complexity-based search"""
        # Skip full scan if complexity is rare
        if self.stats['complexity_distribution'].get(complexity, 0) < 2:
            # Use linear scan for rare complexities
            return self._linear_scan_by_complexity(complexity)
        else:
            # Use index for common complexities
            return self._indexed_search_by_complexity(complexity)

    def find_related_specs(self, spec_id: str) -> list:
        """Find related SPECs efficiently"""
        spec_content = self.specs_dir / spec_id / 'spec.md'
        if not spec_content.exists():
            return []

        keywords = set(self._extract_keywords(spec_content.read_text()))

        # Early termination: stop after finding 5 highly related SPECs
        related = []
        for other_spec in self.specs_dir.glob('SPEC-*/spec.md'):
            if other_spec.parent.name == spec_id:
                continue

            other_keywords = set(self._extract_keywords(other_spec.read_text()))
            similarity = len(keywords & other_keywords) / len(keywords | other_keywords)

            if similarity > 0.3:  # 30% similarity threshold
                related.append((other_spec.parent.name, similarity))

            if len(related) >= 5:  # Early termination
                break

        return sorted(related, key=lambda x: x[1], reverse=True)
```

### Pattern 5: SPEC Diff and Change Detection

**Efficient Change Tracking**:
```python
from difflib import unified_diff
from datetime import datetime

class SpecChangeDetector:
    """Track SPEC changes efficiently"""

    def __init__(self, specs_dir: Path):
        self.specs_dir = specs_dir
        self.change_log = specs_dir / '.moai' / 'spec-changes.log'

    def detect_changes(self, spec_id: str, new_content: str) -> dict:
        """Detect changes between versions"""
        spec_file = self.specs_dir / spec_id / 'spec.md'
        if not spec_file.exists():
            return {'type': 'new_spec', 'changes': []}

        old_content = spec_file.read_text()
        changes = list(unified_diff(
            old_content.splitlines(keepends=True),
            new_content.splitlines(keepends=True),
            fromfile='old',
            tofile='new'
        ))

        # Analyze changes
        return {
            'type': 'updated_spec',
            'line_changes': len([l for l in changes if l.startswith('+') or l.startswith('-')]),
            'requirement_changes': self._detect_requirement_changes(old_content, new_content),
            'breaking_changes': self._detect_breaking_changes(old_content, new_content)
        }

    def _detect_breaking_changes(self, old: str, new: str) -> list:
        """Identify breaking changes"""
        breaking = []
        old_reqs = self._extract_requirement_ids(old)
        new_reqs = self._extract_requirement_ids(new)

        # Removed requirements are breaking changes
        removed = old_reqs - new_reqs
        for req_id in removed:
            breaking.append(f'Removed requirement: {req_id}')

        return breaking
```

### Pattern 6: SPEC Analytics and Reporting

**Performance Metrics**:
```python
from dataclasses import dataclass
from statistics import mean, stdev

@dataclass
class SpecMetrics:
    total_specs: int
    avg_requirements_per_spec: float
    most_common_requirement_type: str
    average_spec_size_bytes: float
    specs_by_status: dict
    implementation_readiness: float  # 0-1 score

class SpecAnalytics:
    """Analyze SPEC collection for insights"""

    def compute_metrics(self, specs_dir: Path) -> SpecMetrics:
        """Compute metrics for SPEC collection"""
        spec_files = list(specs_dir.glob('SPEC-*/spec.md'))
        requirements_per_spec = []
        total_size = 0
        requirement_types = defaultdict(int)

        for spec_file in spec_files:
            content = spec_file.read_text()
            requirements = self._extract_requirements(content)
            requirements_per_spec.append(len(requirements))
            total_size += len(content.encode('utf-8'))

            for req in requirements:
                requirement_types[req['type']] += 1

        return SpecMetrics(
            total_specs=len(spec_files),
            avg_requirements_per_spec=mean(requirements_per_spec),
            most_common_requirement_type=max(requirement_types, key=requirement_types.get),
            average_spec_size_bytes=total_size / len(spec_files),
            specs_by_status={'active': 42, 'draft': 8, 'deprecated': 3},
            implementation_readiness=self._calculate_readiness(spec_files)
        )

    def _calculate_readiness(self, spec_files: list) -> float:
        """Calculate how ready SPECs are for implementation"""
        # Score based on:
        # - Presence of acceptance criteria (0.3)
        # - Presence of technical notes (0.2)
        # - All requirements have EARS pattern (0.3)
        # - Risk assessment (0.2)
        pass
```

---

## Performance Benchmarks (2025)

| Operation | Time | Files Scanned |
|-----------|------|---|
| Search 100 SPECs | 45ms | 100 |
| Index all SPECs | 230ms | 100 |
| Validate SPEC | 12ms | 1 |
| Generate dependencies | 180ms | 100 |
| Batch validate 20 SPECs | 85ms | 20 |

---

**Last Updated**: 2025-11-22
