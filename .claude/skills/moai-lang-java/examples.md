# Java Language - Practical Examples (10 Examples)

## Example 1: Spring Boot REST API

**Problem**: Build production REST API with validation and exception handling.

**Solution**:
```java
@RestController
@RequestMapping("/api/users")
public class UserController {
    
    @Autowired
    private UserService userService;
    
    @GetMapping
    public List<User> getAllUsers() {
        return userService.findAll();
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }
    
    @PostMapping
    public ResponseEntity<User> createUser(@Valid @RequestBody UserDTO userDTO) {
        User user = userService.create(userDTO);
        return ResponseEntity.created(URI.create("/api/users/" + user.getId()))
            .body(user);
    }
}

@Service
public class UserService {
    
    @Autowired
    private UserRepository repository;
    
    @Transactional(readOnly = true)
    public List<User> findAll() {
        return repository.findAll();
    }
    
    @Transactional
    public User create(UserDTO dto) {
        User user = new User();
        user.setName(dto.getName());
        user.setEmail(dto.getEmail());
        return repository.save(user);
    }
}
```

---

## Example 2: Streams and Functional Programming

**Problem**: Process collections functionally with Java Streams API.

**Solution**:
```java
List<String> names = Arrays.asList("Alice", "Bob", "Charlie", "David");

// Filter and map
List<String> result = names.stream()
    .filter(name -> name.length() > 3)
    .map(String::toUpperCase)
    .collect(Collectors.toList());

// Reduce
int total = IntStream.range(1, 11)
    .reduce(0, Integer::sum);

// Grouping
Map<Integer, List<String>> grouped = names.stream()
    .collect(Collectors.groupingBy(String::length));

// Parallel processing
long count = names.parallelStream()
    .filter(name -> name.startsWith("A"))
    .count();
```

---

## Example 3: CompletableFuture for Async

**Problem**: Handle asynchronous operations with CompletableFuture.

**Solution**:
```java
public class AsyncService {
    
    public CompletableFuture<String> fetchData(String url) {
        return CompletableFuture.supplyAsync(() -> {
            // Simulate API call
            return httpClient.get(url);
        });
    }
    
    public CompletableFuture<UserProfile> getUserProfile(Long userId) {
        CompletableFuture<User> userFuture = fetchUser(userId);
        CompletableFuture<List<Order>> ordersFuture = fetchOrders(userId);
        
        return userFuture.thenCombine(ordersFuture, (user, orders) -> {
            UserProfile profile = new UserProfile();
            profile.setUser(user);
            profile.setOrders(orders);
            return profile;
        });
    }
    
    public void handleAsyncWithTimeout() {
        CompletableFuture<String> future = fetchData("https://api.example.com")
            .orTimeout(5, TimeUnit.SECONDS)
            .exceptionally(ex -> "Fallback data");
    }
}
```

---

## Example 4: JPA and Hibernate

**Problem**: Type-safe database access with JPA.

**Solution**:
```java
@Entity
@Table(name = "users")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String name;
    
    @Column(unique = true)
    private String email;
    
    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL)
    private List<Order> orders = new ArrayList<>();
}

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);
    
    @Query("SELECT u FROM User u WHERE u.name LIKE %:name%")
    List<User> searchByName(@Param("name") String name);
}
```

---

## Example 5: Exception Handling

**Problem**: Centralized exception handling with @ControllerAdvice.

**Solution**:
```java
@ControllerAdvice
public class GlobalExceptionHandler {
    
    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(ResourceNotFoundException ex) {
        ErrorResponse error = new ErrorResponse(
            HttpStatus.NOT_FOUND.value(),
            ex.getMessage(),
            LocalDateTime.now()
        );
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
    }
    
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ErrorResponse> handleValidation(MethodArgumentNotValidException ex) {
        Map<String, String> errors = new HashMap<>();
        ex.getBindingResult().getFieldErrors().forEach(error ->
            errors.put(error.getField(), error.getDefaultMessage())
        );
        
        ErrorResponse response = new ErrorResponse(
            HttpStatus.BAD_REQUEST.value(),
            "Validation failed",
            errors
        );
        return ResponseEntity.badRequest().body(response);
    }
}
```

---

## Example 6: Testing with JUnit 5

**Problem**: Comprehensive unit and integration tests.

**Solution**:
```java
@SpringBootTest
@AutoConfigureMockMvc
class UserControllerTest {
    
    @Autowired
    private MockMvc mockMvc;
    
    @MockBean
    private UserService userService;
    
    @Test
    void shouldGetUser() throws Exception {
        User user = new User(1L, "John", "john@example.com");
        when(userService.findById(1L)).thenReturn(Optional.of(user));
        
        mockMvc.perform(get("/api/users/1"))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$.name").value("John"));
    }
    
    @Test
    void shouldReturn404WhenUserNotFound() throws Exception {
        when(userService.findById(999L)).thenReturn(Optional.empty());
        
        mockMvc.perform(get("/api/users/999"))
            .andExpect(status().isNotFound());
    }
}
```

---

## Example 7: Records (Java 14+)

**Problem**: Immutable data classes with minimal boilerplate.

**Solution**:
```java
// Before (verbose)
public final class User {
    private final Long id;
    private final String name;
    
    public User(Long id, String name) {
        this.id = id;
        this.name = name;
    }
    
    // Getters, equals, hashCode, toString...
}

// After (concise)
public record User(Long id, String name) {
    // Auto-generated: constructor, getters, equals, hashCode, toString
}

// Custom validation in compact constructor
public record User(Long id, String name) {
    public User {
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("Name cannot be blank");
        }
    }
}
```

---

## Example 8: Pattern Matching (Java 17+)

**Problem**: Type-safe instanceof with pattern matching.

**Solution**:
```java
// Before
if (obj instanceof String) {
    String str = (String) obj;
    System.out.println(str.toUpperCase());
}

// After (Java 17+)
if (obj instanceof String str) {
    System.out.println(str.toUpperCase());
}

// Switch expressions with pattern matching
String result = switch (obj) {
    case Integer i -> "Integer: " + i;
    case String s -> "String: " + s.toUpperCase();
    case null -> "Null value";
    default -> "Unknown type";
};
```

---

## Example 9: Optional for Null Safety

**Problem**: Avoid NullPointerException with Optional.

**Solution**:
```java
public class UserService {
    
    public Optional<User> findByEmail(String email) {
        return repository.findByEmail(email);
    }
    
    public String getUserName(Long id) {
        return findById(id)
            .map(User::getName)
            .orElse("Unknown");
    }
    
    public void processUser(Long id) {
        findById(id).ifPresent(user -> {
            System.out.println("Processing: " + user.getName());
        });
    }
    
    public User getOrCreate(Long id) {
        return findById(id).orElseGet(() -> createDefaultUser());
    }
}
```

---

## Example 10: Reactive with Project Reactor

**Problem**: Non-blocking reactive streams with Spring WebFlux.

**Solution**:
```java
@RestController
@RequestMapping("/api/reactive")
public class ReactiveController {
    
    @Autowired
    private ReactiveUserRepository repository;
    
    @GetMapping("/users")
    public Flux<User> getAllUsers() {
        return repository.findAll();
    }
    
    @GetMapping("/users/{id}")
    public Mono<ResponseEntity<User>> getUser(@PathVariable Long id) {
        return repository.findById(id)
            .map(ResponseEntity::ok)
            .defaultIfEmpty(ResponseEntity.notFound().build());
    }
    
    @PostMapping("/users")
    public Mono<User> createUser(@RequestBody Mono<User> userMono) {
        return userMono.flatMap(repository::save);
    }
}
```

---

**Total Examples**: 10  
**Lines**: ~600
