# moai-lang-java - Testing & Quality Assurance

_Last updated: 2025-11-21_

Comprehensive testing strategies for Java applications using JUnit 5, Mockito, and Spring Boot Test.

## JUnit 5 Testing with Mockito

```java
@ExtendWith(MockitoExtension.class)
public class UserServiceTest {
    private UserService userService;
    @Mock
    private UserRepository userRepository;

    @BeforeEach
    void setUp() {
        userService = new UserService(userRepository);
    }

    @Test
    void testFindUserById() {
        User user = new User(1L, "test@example.com");
        when(userRepository.findById(1L)).thenReturn(Optional.of(user));

        Optional<User> result = userService.findById(1L);

        assertTrue(result.isPresent());
        assertEquals("test@example.com", result.get().getEmail());
        verify(userRepository, times(1)).findById(1L);
    }

    @Test
    void testCreateUser() {
        User user = new User(null, "new@example.com");
        User savedUser = new User(2L, "new@example.com");
        when(userRepository.save(any(User.class))).thenReturn(savedUser);

        User result = userService.save(user);

        assertNotNull(result.getId());
        assertEquals("new@example.com", result.getEmail());
        verify(userRepository).save(any(User.class));
    }

    @Test
    void testDeleteUser() {
        Long userId = 1L;
        doNothing().when(userRepository).deleteById(userId);

        userService.delete(userId);

        verify(userRepository).deleteById(userId);
    }
}
```

## Integration Test with Spring Boot Test

```java
@SpringBootTest
@AutoConfigureMockMvc
public class UserControllerIntegrationTest {
    @Autowired
    private MockMvc mockMvc;
    @Autowired
    private UserRepository userRepository;

    @Test
    void testGetAllUsers() throws Exception {
        mockMvc.perform(get("/api/users"))
            .andExpect(status().isOk())
            .andExpect(jsonPath("$", hasSize(greaterThanOrEqualTo(0))));
    }

    @Test
    void testCreateUser() throws Exception {
        String userJson = "{\"email\":\"test@example.com\",\"name\":\"Test User\"}";

        mockMvc.perform(post("/api/users")
            .contentType("application/json")
            .content(userJson))
            .andExpect(status().isCreated())
            .andExpect(jsonPath("$.email").value("test@example.com"));
    }

    @Test
    void testGetUserNotFound() throws Exception {
        mockMvc.perform(get("/api/users/999"))
            .andExpect(status().isNotFound());
    }
}
```

## Exception Handling in Tests

```java
@Test
void testResourceNotFound() {
    when(userRepository.findById(999L)).thenReturn(Optional.empty());

    assertThrows(ResourceNotFoundException.class, () -> {
        userService.findById(999L);
    });
}

@Test
@DisplayName("Should return 404 when user not found")
void testGetUserNotFound() throws Exception {
    mockMvc.perform(get("/api/users/999"))
        .andExpect(status().isNotFound());
}

@RestControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(ResourceNotFoundException.class)
    public ErrorResponse handleResourceNotFound(ResourceNotFoundException ex, WebRequest request) {
        return ErrorResponse.builder()
            .timestamp(LocalDateTime.now())
            .status(HttpStatus.NOT_FOUND.value())
            .message(ex.getMessage())
            .path(request.getDescription(false).replace("uri=", ""))
            .build();
    }

    @ExceptionHandler(Exception.class)
    public ErrorResponse handleGlobalException(Exception ex, WebRequest request) {
        return ErrorResponse.builder()
            .timestamp(LocalDateTime.now())
            .status(HttpStatus.INTERNAL_SERVER_ERROR.value())
            .message("An error occurred: " + ex.getMessage())
            .path(request.getDescription(false).replace("uri=", ""))
            .build();
    }
}
```

## Microservices Testing with Mockito

```java
@Test
@DisplayName("Should handle service timeout gracefully")
void testServiceTimeout() {
    when(userServiceClient.getUser(anyLong()))
        .thenThrow(new FeignException.ServiceUnavailable("Service unavailable", null, null));

    assertThrows(ServiceException.class, () -> {
        orderService.getOrderWithUser(1L);
    });
}

@Test
@DisplayName("Should retry on transient failure")
void testRetryOnFailure() {
    when(userServiceClient.getUser(1L))
        .thenThrow(new FeignException.InternalServerError("Error", null, null))
        .thenReturn(new UserDTO(1L, "user@example.com"));

    UserDTO result = orderService.getOrderWithUser(1L);
    assertEquals("user@example.com", result.getEmail());
}
```

## Parameterized Tests

```java
@ParameterizedTest
@ValueSource(strings = { "valid@example.com", "test@domain.org" })
void testValidEmails(String email) {
    User user = new User();
    user.setEmail(email);

    assertTrue(emailValidator.isValid(email));
}

@ParameterizedTest
@CsvSource({
    "1, John, john@example.com",
    "2, Jane, jane@example.com",
    "3, Bob, bob@example.com"
})
void testMultipleUsers(Long id, String name, String email) {
    User user = new User(id, email, name);
    assertEquals(name, user.getName());
    assertEquals(email, user.getEmail());
}
```

---

_For basic Java programming, see SKILL.md and examples.md_
_For Spring Boot patterns, see spring-boot.md_
_For build tools and test execution, see reference.md_
