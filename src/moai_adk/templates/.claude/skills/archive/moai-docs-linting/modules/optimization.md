# Optimization for Documentation Linting

Performance optimization strategies for markdown validation.

---

## Caching Lint Results

Cache validation results to avoid redundant checks:

```python
from functools import lru_cache
import hashlib

class CachedLinter:
    @lru_cache(maxsize=1000)
    def lint_cached(self, file_hash: str) -> List[Issue]:
        """Cache lint results by file hash"""
        return self._lint_uncached(file_hash)

    def invalidate_cache(self, file_path: Path):
        """Clear cache for modified file"""
        self.lint_cached.cache_clear()
```

---

## Parallel Linting

Process multiple files simultaneously:

```python
from concurrent.futures import ThreadPoolExecutor

def lint_parallel(files: List[Path], max_workers: int = 4) -> Dict[Path, List[Issue]]:
    """Lint files in parallel"""
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {executor.submit(lint_file, f): f for f in files}
        return {futures[future]: future.result() for future in futures}
```

---

## Best Practices

- ✅ Cache lint results
- ✅ Process files in parallel
- ✅ Use incremental linting for large docs
- ❌ Don't re-lint unchanged files

---

**Last Updated**: 2025-11-22
