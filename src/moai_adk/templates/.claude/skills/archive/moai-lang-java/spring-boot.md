# moai-lang-java - Spring Boot & Enterprise Development

_Last updated: 2025-11-21_

Spring Boot framework patterns for building REST APIs, microservices, and enterprise applications with clean architecture.

## Spring Boot REST API with Validation

```java
@RestController
@RequestMapping("/api/users")
public class UserController {
    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserDTO> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<UserDTO> createUser(@Valid @RequestBody UserDTO dto) {
        UserDTO created = userService.save(dto);
        return ResponseEntity.status(201).body(created);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserDTO> updateUser(@PathVariable Long id, @Valid @RequestBody UserDTO dto) {
        return userService.update(id, dto)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        userService.delete(id);
        return ResponseEntity.noContent().build();
    }
}
```

## JPA Entity with Relationships

```java
@Entity
@Table(name = "users")
@Data
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String email;

    private String name;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL)
    private List<Post> posts = new ArrayList<>();

    @ManyToMany
    @JoinTable(
        name = "user_roles",
        joinColumns = @JoinColumn(name = "user_id"),
        inverseJoinColumns = @JoinColumn(name = "role_id")
    )
    private List<Role> roles = new ArrayList<>();

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}

@Entity
@Table(name = "posts")
@Data
public class Post {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private String title;
    private String content;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;
}
```

## Service Layer with Transactions

```java
@Service
@Transactional
public class UserService {
    private final UserRepository userRepository;

    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    @Transactional(readOnly = true)
    public Optional<UserDTO> findById(Long id) {
        return userRepository.findById(id).map(this::toDTO);
    }

    public UserDTO save(UserDTO dto) {
        User user = new User();
        user.setEmail(dto.getEmail());
        user.setName(dto.getName());
        User saved = userRepository.save(user);
        return toDTO(saved);
    }

    public Optional<UserDTO> update(Long id, UserDTO dto) {
        return userRepository.findById(id)
            .map(user -> {
                user.setName(dto.getName());
                user.setEmail(dto.getEmail());
                return toDTO(userRepository.save(user));
            });
    }

    public void delete(Long id) {
        userRepository.deleteById(id);
    }

    private UserDTO toDTO(User user) {
        UserDTO dto = new UserDTO();
        dto.setEmail(user.getEmail());
        dto.setName(user.getName());
        return dto;
    }
}

public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);
}
```

## Microservices Pattern with Feign Client

```java
@FeignClient(name = "user-service", url = "http://localhost:8081")
public interface UserServiceClient {
    @GetMapping("/api/users/{id}")
    UserDTO getUser(@PathVariable("id") Long id);
}

@RestController
@RequestMapping("/api/orders")
public class OrderController {
    private final UserServiceClient userServiceClient;

    public OrderController(UserServiceClient userServiceClient) {
        this.userServiceClient = userServiceClient;
    }

    @GetMapping("/{id}/user")
    public UserDTO getOrderUser(@PathVariable Long id) {
        return userServiceClient.getUser(id);
    }
}
```

## Stream API & Functional Programming

```java
public class UserStreamExamples {
    public List<String> getAdminEmails(List<User> users) {
        return users.stream()
            .filter(user -> user.getRoles().stream()
                .anyMatch(role -> "ADMIN".equals(role.getName())))
            .map(User::getEmail)
            .sorted()
            .collect(Collectors.toList());
    }

    public Optional<User> findUserByEmail(List<User> users, String email) {
        return users.stream()
            .filter(user -> email.equals(user.getEmail()))
            .findFirst();
    }

    public Map<String, List<User>> groupUsersByRole(List<User> users) {
        return users.stream()
            .flatMap(user -> user.getRoles().stream()
                .map(role -> new AbstractMap.SimpleEntry<>(role.getName(), user)))
            .collect(Collectors.groupingBy(
                AbstractMap.SimpleEntry::getKey,
                Collectors.mapping(AbstractMap.SimpleEntry::getValue, Collectors.toList())
            ));
    }
}
```

---

_For basic Java programming, see SKILL.md and examples.md_
_For testing patterns, see testing.md_
_For build tools and deployment, see reference.md_
