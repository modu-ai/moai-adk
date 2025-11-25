# Python Performance

## List Comprehensions vs Loops
```python
# Fast
result = [x*2 for x in range(1000)]

# Slow
result = []
for x in range(1000):
    result.append(x*2)
```

## Generator Expressions
```python
# Memory efficient
total = sum(x**2 for x in range(10000))
```

## Caching with functools
```python
from functools import lru_cache

@lru_cache(maxsize=128)
def fibonacci(n):
    if n < 2:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

## NumPy for Arrays
```python
import numpy as np

# Fast vectorized operations
arr = np.array([1, 2, 3, 4])
result = arr * 2
```
