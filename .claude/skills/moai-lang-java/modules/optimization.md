# Java Performance Optimization

## JVM Tuning

### Heap Configuration
```bash
# Set heap size
java -Xms2g -Xmx4g -jar app.jar

# GC tuning (G1GC)
java -XX:+UseG1GC \
     -XX:MaxGCPauseMillis=200 \
     -XX:G1HeapRegionSize=16m \
     -jar app.jar
```

### GC Monitoring
```bash
# Enable GC logs
java -Xlog:gc*:file=gc.log \
     -XX:+PrintGCDetails \
     -jar app.jar
```

---

## Collections Optimization

### ArrayList vs LinkedList
```java
// Use ArrayList for random access
List<String> list = new ArrayList<>(1000);

// Use LinkedList for frequent insertions
List<String> list = new LinkedList<>();
```

### HashMap Sizing
```java
// Pre-size HashMap
Map<String, Integer> map = new HashMap<>(1000);

// Load factor optimization
Map<String, Integer> map = new HashMap<>(1000, 0.75f);
```

---

## Stream Optimization

### Parallel Streams
```java
// Use parallel for large datasets
List<Integer> result = numbers.parallelStream()
    .filter(n -> n % 2 == 0)
    .map(n -> n * 2)
    .collect(Collectors.toList());
```

---

## Database Optimization

### JPA Batch Operations
```java
@Transactional
public void batchInsert(List<User> users) {
    int batchSize = 50;
    for (int i = 0; i < users.size(); i++) {
        entityManager.persist(users.get(i));
        if (i % batchSize == 0 && i > 0) {
            entityManager.flush();
            entityManager.clear();
        }
    }
}
```

---

## Caching

### Caffeine Cache
```java
LoadingCache<Long, User> cache = Caffeine.newBuilder()
    .maximumSize(1000)
    .expireAfterWrite(10, TimeUnit.MINUTES)
    .build(key -> repository.findById(key).orElse(null));
```

---

**Total Lines**: ~250
