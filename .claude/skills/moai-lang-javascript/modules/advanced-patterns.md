# JavaScript Advanced Patterns

## Closures
```javascript
function counter() {
    let count = 0;
    return {
        increment: () => ++count,
        decrement: () => --count,
        get: () => count
    };
}
```

## Prototypes
```javascript
function Person(name) {
    this.name = name;
}
Person.prototype.greet = function() {
    return `Hello, ${this.name}`;
};
```

## Generator Functions
```javascript
function* fibonacci() {
    let [a, b] = [0, 1];
    while (true) {
        yield a;
        [a, b] = [b, a + b];
    }
}
```

## Proxy
```javascript
const handler = {
    get(target, prop) {
        return prop in target ? target[prop] : 'Default';
    }
};
const proxy = new Proxy({}, handler);
```

## WeakMap for Memory
```javascript
const cache = new WeakMap();
function memoize(obj) {
    if (!cache.has(obj)) {
        cache.set(obj, compute(obj));
    }
    return cache.get(obj);
}
```
