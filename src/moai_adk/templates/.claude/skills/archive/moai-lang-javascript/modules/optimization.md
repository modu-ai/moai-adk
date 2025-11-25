# JavaScript Performance

## Debounce
```javascript
function debounce(func, wait) {
    let timeout;
    return function(...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), wait);
    };
}
```

## Throttle
```javascript
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}
```

## Lazy Loading
```javascript
const LazyComponent = React.lazy(() => import('./Component'));
```

## Memoization
```javascript
const memo = {};
function fibonacci(n) {
    if (n in memo) return memo[n];
    if (n <= 1) return n;
    return memo[n] = fibonacci(n-1) + fibonacci(n-2);
}
```
