# Java Advanced Patterns

## Design Patterns (Gang of Four)

### Singleton Pattern
```java
public enum DatabaseConnection {
    INSTANCE;
    
    private Connection connection;
    
    public Connection getConnection() {
        if (connection == null) {
            connection = DriverManager.getConnection(url);
        }
        return connection;
    }
}
```

### Factory Pattern
```java
public interface PaymentProcessor {
    void process(Payment payment);
}

public class PaymentFactory {
    public static PaymentProcessor create(PaymentType type) {
        return switch (type) {
            case CREDIT_CARD -> new CreditCardProcessor();
            case PAYPAL -> new PayPalProcessor();
            case CRYPTO -> new CryptoProcessor();
        };
    }
}
```

### Builder Pattern
```java
public class User {
    private final String name;
    private final String email;
    private final int age;
    
    private User(Builder builder) {
        this.name = builder.name;
        this.email = builder.email;
        this.age = builder.age;
    }
    
    public static class Builder {
        private String name;
        private String email;
        private int age;
        
        public Builder name(String name) {
            this.name = name;
            return this;
        }
        
        public Builder email(String email) {
            this.email = email;
            return this;
        }
        
        public User build() {
            return new User(this);
        }
    }
}
```

---

## Generics

### Generic Methods
```java
public class Utils {
    public static <T> List<T> filter(List<T> list, Predicate<T> pred) {
        return list.stream().filter(pred).collect(Collectors.toList());
    }
    
    public static <T extends Comparable<T>> T max(List<T> list) {
        return list.stream().max(Comparable::compareTo).orElse(null);
    }
}
```

### Bounded Type Parameters
```java
public class Repository<T extends Entity> {
    public T save(T entity) {
        // Implementation
        return entity;
    }
    
    public Optional<T> findById(Long id) {
        // Implementation
        return Optional.empty();
    }
}
```

---

## Annotations and Reflection

### Custom Annotations
```java
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.FIELD)
public @interface Validate {
    String message() default "Validation failed";
    int min() default 0;
    int max() default Integer.MAX_VALUE;
}

public class Validator {
    public static void validate(Object obj) throws Exception {
        for (Field field : obj.getClass().getDeclaredFields()) {
            if (field.isAnnotationPresent(Validate.class)) {
                Validate annotation = field.getAnnotation(Validate.class);
                field.setAccessible(true);
                Object value = field.get(obj);
                
                if (value instanceof Integer i) {
                    if (i < annotation.min() || i > annotation.max()) {
                        throw new ValidationException(annotation.message());
                    }
                }
            }
        }
    }
}
```

---

## Concurrency Patterns

### Executor Framework
```java
ExecutorService executor = Executors.newFixedThreadPool(10);

List<Future<String>> futures = new ArrayList<>();
for (Task task : tasks) {
    Future<String> future = executor.submit(() -> processTask(task));
    futures.add(future);
}

for (Future<String> future : futures) {
    String result = future.get(); // Blocking
}

executor.shutdown();
```

### CompletableFuture Chaining
```java
CompletableFuture<User> userFuture = fetchUser(userId);

userFuture
    .thenApply(user -> user.getName())
    .thenApply(String::toUpperCase)
    .thenAccept(System.out::println)
    .exceptionally(ex -> {
        log.error("Error: " + ex);
        return null;
    });
```

---

## Reactive Streams

### Flux and Mono
```java
@Service
public class ReactiveService {
    
    public Flux<Data> streamData() {
        return Flux.range(1, 10)
            .map(i -> new Data(i))
            .delayElements(Duration.ofSeconds(1));
    }
    
    public Mono<User> getUser(Long id) {
        return webClient.get()
            .uri("/users/{id}", id)
            .retrieve()
            .bodyToMono(User.class)
            .timeout(Duration.ofSeconds(5))
            .retry(3);
    }
}
```

---

**Total Lines**: ~300
