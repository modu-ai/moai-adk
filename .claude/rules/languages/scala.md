---
paths:
  - "**/*.scala"
  - "**/*.sc"
  - "**/build.sbt"
---

# Scala Rules

Version: Scala 3.4+

## Tooling

- Build: sbt or Mill
- Testing: ScalaTest or specs2
- Formatting: scalafmt

## Preferred Patterns

- Prefer immutability
- Use case classes for data
- Use ZIO or Cats Effect for FP

## MoAI Integration

- Use Skill("moai-lang-scala") for detailed patterns
- Follow TRUST 5 quality gates
