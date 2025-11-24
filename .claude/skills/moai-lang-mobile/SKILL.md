---
name: moai-lang-mobile
description: "Mobile development consolidated: Swift for iOS, Dart/Flutter for cross-platform"
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules: []
dependencies:
  - moai-foundation-trust
deprecated: false
successor: null
category_tier: 4
auto_trigger_keywords:
  - swift
  - ios
  - dart
  - flutter
  - mobile
  - app
  - cross-platform
  - async
  - providers
agent_coverage:
  - frontend-expert
  - tdd-implementer
  - quality-gate
context7_references:
  - swift
  - dart
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Mobile Languages Consolidated**

Unified mobile platform with Swift for iOS and Dart/Flutter for cross-platform development.

**Core Languages**:
- **Swift**: Modern, safe, performant iOS/macOS development with async/await
- **Dart**: Single language for iOS, Android, web via Flutter framework

**When to Use**:
- Native iOS apps with Swift
- Cross-platform apps with Flutter
- Mobile UI with reactive patterns
- High-performance mobile applications

---

## Core Languages

### Swift

Modern iOS/macOS development with safety and performance.

```swift
// Async/await for concurrency
async func fetchUser(id: Int) -> User {
    let response = try await URLSession.shared.data(from: URL(string: "/users/\(id)")!)
    return try JSONDecoder().decode(User.self, from: response.0)
}

// Structured concurrency
async {
    async let user = fetchUser(id: 1)
    async let posts = fetchPosts(userId: 1)

    let (userData, postsData) = await (user, posts)
}
```

---

### Dart/Flutter

Cross-platform mobile with reactive UI.

```dart
// Async/await
Future<User> fetchUser(int id) async {
  final response = await http.get(Uri.parse('/users/$id'));
  return User.fromJson(json.decode(response.body));
}

// Flutter widgets with Provider
Consumer<UserProvider>(
  builder: (context, provider, _) =>
    Text('User: ${provider.user.name}'),
)
```

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
