# iOS Native Development

Swift, SwiftUI, UIKit, and Xcode patterns for iOS native development.

## Quick Reference

Key Technologies:
- **Language**: Swift 5.9+ with Swift Concurrency (`async/await`, `actor`, `Sendable`)
- **UI framework**: SwiftUI (primary), UIKit (interop, complex custom views)
- **Data**: Core Data (iOS 13+), SwiftData (iOS 17+)
- **Dependency management**: Swift Package Manager (SPM)
- **Minimum target**: iOS 16+ recommended (iOS 17+ for SwiftData, `@Observable`)

---

## Core Patterns

### App Entry Point

```swift
@main
struct MyApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

### Navigation with NavigationStack (iOS 16+)

```swift
struct RootView: View {
    @State private var path = NavigationPath()

    var body: some View {
        NavigationStack(path: $path) {
            HomeView()
                .navigationDestination(for: UserRoute.self) { route in
                    switch route {
                    case .detail(let id): UserDetailView(id: id)
                    case .settings: SettingsView()
                    }
                }
        }
    }
}
```

### State Management

```swift
// iOS 17+: @Observable macro replaces ObservableObject
@Observable
final class UserViewModel {
    var users: [User] = []
    var isLoading = false
    private(set) var error: Error?

    func loadUsers() async {
        isLoading = true
        defer { isLoading = false }
        do {
            users = try await UserService.shared.fetchUsers()
        } catch {
            self.error = error
        }
    }
}

// Usage in view
struct UserListView: View {
    @State private var viewModel = UserViewModel()

    var body: some View {
        List(viewModel.users) { user in
            Text(user.name)
        }
        .task { await viewModel.loadUsers() }
    }
}
```

### Networking with async/await

```swift
actor UserService {
    static let shared = UserService()
    private let session = URLSession.shared

    func fetchUsers() async throws -> [User] {
        let url = URL(string: "https://api.example.com/users")!
        let (data, response) = try await session.data(from: url)
        guard let http = response as? HTTPURLResponse, http.statusCode == 200 else {
            throw URLError(.badServerResponse)
        }
        return try JSONDecoder().decode([User].self, from: data)
    }
}
```

---

## Works Well With

- expert-mobile (primary agent)
- moai-domain-mobile/modules/strategy-comparison.md (paradigm selection)
- expert-security (certificate pinning, Keychain storage, biometric auth)
