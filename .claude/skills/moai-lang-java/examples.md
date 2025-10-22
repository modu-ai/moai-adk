# moai-lang-java - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Maven

```bash
# Create new Maven project
mvn archetype:generate \
  -DgroupId=com.example \
  -DartifactId=my-project \
  -DarchetypeArtifactId=maven-archetype-quickstart \
  -DarchetypeVersion=1.4 \
  -DinteractiveMode=false

cd my-project

# Build project
mvn clean install

# Run tests
mvn test

# Run with coverage
mvn test jacoco:report
```

**pom.xml configuration**:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.example</groupId>
    <artifactId>my-project</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>23</maven.compiler.source>
        <maven.compiler.target>23</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <junit.version>5.11.0</junit.version>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>${junit.version}</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>3.2.5</version>
            </plugin>
            <plugin>
                <groupId>org.jacoco</groupId>
                <artifactId>jacoco-maven-plugin</artifactId>
                <version>0.8.11</version>
            </plugin>
        </plugins>
    </build>
</project>
```

## Example 2: TDD Workflow with JUnit 5

**RED: Write failing test**
```java
// src/test/java/com/example/CalculatorTest.java
package com.example;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

class CalculatorTest {
    
    @Test
    @DisplayName("Should add two positive numbers")
    void testAddPositive() {
        Calculator calculator = new Calculator();
        assertEquals(5, calculator.add(2, 3));
    }
    
    @Test
    @DisplayName("Should handle negative numbers")
    void testAddNegative() {
        Calculator calculator = new Calculator();
        assertEquals(-3, calculator.add(-1, -2));
    }
    
    @Test
    @DisplayName("Should handle zero")
    void testAddZero() {
        Calculator calculator = new Calculator();
        assertEquals(5, calculator.add(0, 5));
    }
}
```

**GREEN: Implement feature**
```java
// src/main/java/com/example/Calculator.java
package com.example;

public class Calculator {
    public int add(int a, int b) {
        return a + b;
    }
}
```

**REFACTOR: Add validation and documentation**
```java
// src/main/java/com/example/Calculator.java
package com.example;

/**
 * Calculator class providing basic arithmetic operations.
 */
public class Calculator {
    
    /**
     * Adds two integers.
     * 
     * @param a first operand
     * @param b second operand
     * @return sum of a and b
     * @throws ArithmeticException if overflow occurs
     */
    public int add(int a, int b) {
        int result = a + b;
        // Check for overflow
        if (((a ^ result) & (b ^ result)) < 0) {
            throw new ArithmeticException("Integer overflow");
        }
        return result;
    }
}
```

## Example 3: Spring Boot Integration Testing

```java
// src/test/java/com/example/UserServiceTest.java
package com.example;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import static org.mockito.Mockito.*;
import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
class UserServiceTest {
    
    @Autowired
    private UserService userService;
    
    @MockBean
    private UserRepository userRepository;
    
    @Test
    void shouldCreateUser() {
        User user = new User("john@example.com", "John Doe");
        when(userRepository.save(any(User.class))).thenReturn(user);
        
        User result = userService.createUser(user);
        
        assertNotNull(result);
        assertEquals("john@example.com", result.getEmail());
        verify(userRepository, times(1)).save(user);
    }
}
```

## Example 4: Parameterized Tests

```java
// src/test/java/com/example/ParameterizedTest.java
package com.example;

import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.*;
import static org.junit.jupiter.api.Assertions.*;

class ParameterizedCalculatorTest {
    
    @ParameterizedTest
    @CsvSource({
        "2, 3, 5",
        "-1, -2, -3",
        "0, 5, 5",
        "100, 200, 300"
    })
    void testAddWithMultipleInputs(int a, int b, int expected) {
        Calculator calc = new Calculator();
        assertEquals(expected, calc.add(a, b));
    }
    
    @ParameterizedTest
    @ValueSource(ints = {1, 2, 3, 5, 8, 13})
    void testPositiveNumbers(int num) {
        assertTrue(num > 0);
    }
}
```

## Example 5: Mocking with Mockito

```java
// src/test/java/com/example/ApiClientTest.java
package com.example;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ApiClientTest {
    
    @Mock
    private HttpClient httpClient;
    
    @InjectMocks
    private ApiClient apiClient;
    
    @Test
    void shouldFetchUserSuccessfully() {
        String mockResponse = "{\"id\": 1, \"name\": \"John\"}";
        when(httpClient.get("/users/1")).thenReturn(mockResponse);
        
        User user = apiClient.fetchUser(1);
        
        assertNotNull(user);
        assertEquals(1, user.getId());
        verify(httpClient).get("/users/1");
    }
}
```

## Example 6: Test Lifecycle and Setup

```java
// src/test/java/com/example/DatabaseTest.java
package com.example;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class DatabaseTest {
    
    private Database database;
    
    @BeforeAll
    void setupDatabase() {
        database = new Database("test_db");
        database.connect();
    }
    
    @BeforeEach
    void clearData() {
        database.clear();
    }
    
    @Test
    void shouldInsertUser() {
        User user = new User("test@example.com");
        int id = database.insert(user);
        assertTrue(id > 0);
    }
    
    @AfterEach
    void verifyCleanup() {
        // Verify no locks or open transactions
    }
    
    @AfterAll
    void closeDatabase() {
        database.disconnect();
    }
}
```

## Example 7: Exception Testing

```java
// src/test/java/com/example/ValidationTest.java
package com.example;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class ValidationTest {
    
    @Test
    void shouldThrowOnInvalidEmail() {
        UserValidator validator = new UserValidator();
        
        Exception exception = assertThrows(
            ValidationException.class,
            () -> validator.validateEmail("invalid-email")
        );
        
        assertTrue(exception.getMessage().contains("Invalid email"));
    }
    
    @Test
    void shouldNotThrowOnValidEmail() {
        UserValidator validator = new UserValidator();
        assertDoesNotThrow(() -> 
            validator.validateEmail("valid@example.com")
        );
    }
}
```

---

_For complete API reference and configuration options, see reference.md_
