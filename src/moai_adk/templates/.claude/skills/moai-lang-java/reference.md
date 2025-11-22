# moai-lang-java - CLI Reference & Build Tools

_Last updated: 2025-11-21_

## Java Installation & Setup

### Install JDK

```bash
# macOS (Homebrew)
brew install openjdk@21

# Ubuntu/Debian
sudo apt-get install openjdk-21-jdk

# Verify installation
java --version
javac --version
```

### Set JAVA_HOME

```bash
# macOS
export JAVA_HOME=$(/usr/libexec/java_home -v 21)

# Linux
export JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64

# Add to ~/.bashrc or ~/.zshrc for persistence
```

## Maven Build System

### Maven Project Structure

```
myproject/
├── pom.xml
├── src/
│   ├── main/
│   │   ├── java/
│   │   └── resources/
│   └── test/
│       ├── java/
│       └── resources/
└── target/
```

### Maven Commands

```bash
# Create new project
mvn archetype:generate \
  -DgroupId=com.example \
  -DartifactId=my-app \
  -DarchetypeArtifactId=maven-archetype-quickstart

# Build project
mvn clean install

# Compile only
mvn compile

# Run tests
mvn test

# Skip tests during build
mvn clean install -DskipTests

# Run specific test
mvn test -Dtest=UserServiceTest

# Package as JAR
mvn clean package

# Run application
mvn spring-boot:run

# Generate Javadoc
mvn javadoc:javadoc
```

### Spring Boot with Maven

```xml
<!-- pom.xml snippet -->
<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.2.0</version>
</parent>

<dependencies>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
    </dependency>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-data-jpa</artifactId>
    </dependency>
    <dependency>
        <groupId>com.h2database</groupId>
        <artifactId>h2</artifactId>
        <scope>runtime</scope>
    </dependency>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-test</artifactId>
        <scope>test</scope>
    </dependency>
</dependencies>
```

## Gradle Build System

### Gradle Project Structure

```
myproject/
├── build.gradle
├── settings.gradle
├── gradle/
│   └── wrapper/
└── src/
    ├── main/
    └── test/
```

### Gradle Commands

```bash
# Create new project
gradle init --type java-application

# Build project
./gradlew build

# Compile only
./gradlew compileJava

# Run tests
./gradlew test

# Build without tests
./gradlew build -x test

# Run specific test
./gradlew test --tests UserServiceTest

# Run application
./gradlew bootRun

# Clean build artifacts
./gradlew clean
```

### Spring Boot with Gradle

```gradle
// build.gradle
plugins {
    id 'java'
    id 'org.springframework.boot' version '3.2.0'
    id 'io.spring.dependency-management' version '1.1.4'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'
    runtimeOnly 'com.h2database:h2'
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
}
```

## Testing with JUnit 5 & Mockito

```bash
# Run all tests
mvn test

# Run with coverage (Jacoco)
mvn clean test jacoco:report

# View coverage report
open target/site/jacoco/index.html  # macOS
firefox target/site/jacoco/index.html  # Linux
```

## Code Quality Tools

### Checkstyle

```bash
# Run checkstyle
mvn checkstyle:check

# Generate report
mvn checkstyle:checkstyle
```

### SonarQube

```bash
# Analyze project
mvn clean verify sonar:sonar \
  -Dsonar.projectKey=my-project \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.login=your-token
```

### Spotbugs (Bug Detection)

```bash
# Run spotbugs
mvn spotbugs:check

# Generate report
mvn spotbugs:spotbugs
```

## Application Properties (Spring Boot)

```properties
# application.properties
spring.application.name=my-app
server.port=8080
server.servlet.context-path=/api

# Database Configuration
spring.datasource.url=jdbc:mysql://localhost:3306/mydb
spring.datasource.username=root
spring.datasource.password=password
spring.datasource.driver-class-name=com.mysql.cj.jdbc.Driver

# JPA/Hibernate
spring.jpa.hibernate.ddl-auto=update
spring.jpa.show-sql=true
spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.MySQL8Dialect

# Logging
logging.level.root=INFO
logging.level.com.example=DEBUG
logging.pattern.console=%d{yyyy-MM-dd HH:mm:ss} - %logger{36} - %msg%n
```

## Docker Support

### Dockerfile for Spring Boot

```dockerfile
FROM openjdk:21-slim

WORKDIR /app

COPY target/my-app.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
```

### Docker Commands

```bash
# Build Docker image
docker build -t my-app:latest .

# Run Docker container
docker run -p 8080:8080 my-app:latest

# Push to registry
docker push myregistry/my-app:latest

# Docker Compose
docker-compose up -d
docker-compose down
```

## Debugging Java Applications

### Remote Debugging

```bash
# Run with debug flags
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005 -jar app.jar

# Connect with IDE or jdb (Java Debugger)
jdb -attach localhost:5005
```

### Profiling with JFR (Java Flight Recorder)

```bash
# Run with JFR enabled
java -XX:StartFlightRecording=duration=60s,filename=recording.jfr -jar app.jar

# Analyze recording
jcmd <pid> JFR.dump name=1 filename=recording.jfr
```

## Maven Dependency Management

```bash
# Display dependency tree
mvn dependency:tree

# Find dependency conflicts
mvn dependency:analyze

# Update dependencies
mvn versions:use-latest-versions

# Check for security vulnerabilities
mvn dependency-check:check
```

## Useful Maven Plugins

```xml
<!-- pom.xml -->
<build>
    <plugins>
        <!-- Spring Boot Maven Plugin -->
        <plugin>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-maven-plugin</artifactId>
        </plugin>

        <!-- Jacoco for Code Coverage -->
        <plugin>
            <groupId>org.jacoco</groupId>
            <artifactId>jacoco-maven-plugin</artifactId>
            <version>0.8.10</version>
        </plugin>

        <!-- Surefire for Running Tests -->
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-surefire-plugin</artifactId>
            <version>3.1.2</version>
        </plugin>
    </plugins>
</build>
```

## Tool Versions (2025-11-21)

- **Java**: 21 LTS (23.0.0 latest)
- **Maven**: 3.9.9
- **Gradle**: 8.12.0
- **Spring Boot**: 3.2.0
- **JUnit**: 5.11.0
- **Mockito**: 5.6.0
- **Hibernate**: 6.4.0
- **Jackson**: 2.16.0

---

_For examples and advanced patterns, see SKILL.md and examples.md_
