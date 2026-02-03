---
paths:
  - "**/*.java"
  - "**/pom.xml"
  - "**/build.gradle"
  - "**/build.gradle.kts"
---

# Java Rules

Version: Java 21 LTS

## Tooling

- Build: Maven or Gradle
- Testing: JUnit 5 with coverage >= 85%
- Linting: SpotBugs, PMD

## Preferred Patterns

- Use virtual threads for concurrency
- Use records for data classes
- Use Spring Boot 3.3 for web apps

## MoAI Integration

- Use Skill("moai-lang-java") for detailed patterns
- Follow TRUST 5 quality gates
