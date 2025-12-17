---
name: moai-lang-swift
description: Swift 6+ development specialist covering SwiftUI, Combine, Swift Concurrency, and iOS patterns. Use when building iOS apps, macOS apps, or Apple platform applications.
version: 1.0.0
category: language
tags: [swift, swiftui, ios, macos, combine, concurrency]
context7-libraries: [/apple/swift, /apple/swift-evolution]
related-skills: [moai-lang-kotlin, moai-lang-flutter]
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Swift 6+ Development Expert - iOS/macOS with SwiftUI, Combine, and Swift Concurrency.

Auto-Triggers: Swift files (`.swift`), iOS/macOS projects, Xcode workspaces

Core Capabilities:
- Swift 6.0: Typed throws, actors, data-race safety by default
- SwiftUI 6: @Observable, NavigationStack, modern declarative UI
- Combine: Reactive programming with publishers and subscribers
- Swift Concurrency: async/await, actors, TaskGroup
- XCTest: Unit testing, UI testing, async test support
- Swift Package Manager: Dependency management

Version Requirements:
- Swift: 6.0+
- Xcode: 16.0+
- iOS: 17.0+ (recommended), minimum 15.0
- macOS: 14.0+ (recommended)

## Implementation Guide (5 minutes)

### Swift 6.0 Core Features

Typed Throws (Error Type Specification):
```swift
enum NetworkError: Error {
    case invalidURL
    case requestFailed(statusCode: Int)
    case decodingFailed
}

func fetchData() throws(NetworkError) -> Data {
    guard let url = URL(string: "https://api.example.com") else {
        throw .invalidURL
    }
    // Implementation
}

// Caller knows exact error types
do {
    let data = try fetchData()
} catch .invalidURL {
    print("Invalid URL")
} catch .requestFailed(let code) {
    print("Request failed: \(code)")
} catch .decodingFailed {
    print("Decoding failed")
}
```

Complete Concurrency (Data-Race Safety):
```swift
// Swift 6 enforces data-race safety by default
actor UserCache {
    private var cache: [String: User] = [:]

    func get(_ id: String) -> User? { cache[id] }
    func set(_ id: String, user: User) { cache[id] = user }
    func clear() { cache.removeAll() }
}

// Sendable conformance required for cross-actor data
struct User: Codable, Identifiable, Sendable {
    let id: String
    let name: String
    let email: String
}

// MainActor for UI-related code
@MainActor
final class UserViewModel: ObservableObject {
    @Published private(set) var user: User?
    private let cache = UserCache()

    func loadUser(_ id: String) async throws {
        if let cached = await cache.get(id) {
            self.user = cached
            return
        }
        let user = try await api.fetchUser(id)
        await cache.set(id, user: user)
        self.user = user
    }
}
```

### SwiftUI 6 Patterns

@Observable Macro (iOS 17+):
```swift
import Observation

@Observable
class ProfileViewModel {
    var user: User?
    var isLoading = false
    var errorMessage: String?

    func loadProfile() async {
        isLoading = true
        defer { isLoading = false }

        do {
            user = try await api.fetchCurrentUser()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct ProfileView: View {
    @State private var viewModel = ProfileViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else if let user = viewModel.user {
                    UserDetailView(user: user)
                } else if let error = viewModel.errorMessage {
                    ErrorView(message: error)
                }
            }
            .task { await viewModel.loadProfile() }
            .navigationTitle("Profile")
        }
    }
}
```

Modern Navigation (NavigationStack):
```swift
@Observable
class NavigationRouter {
    var path = NavigationPath()

    func push<D: Hashable>(_ destination: D) {
        path.append(destination)
    }

    func pop() {
        guard !path.isEmpty else { return }
        path.removeLast()
    }

    func popToRoot() {
        path.removeLast(path.count)
    }
}

struct ContentView: View {
    @State private var router = NavigationRouter()

    var body: some View {
        NavigationStack(path: $router.path) {
            HomeView()
                .navigationDestination(for: User.self) { user in
                    UserDetailView(user: user)
                }
                .navigationDestination(for: Post.self) { post in
                    PostDetailView(post: post)
                }
        }
        .environment(router)
    }
}
```

### Swift Concurrency Patterns

Async/Await with Error Handling:
```swift
protocol UserServiceProtocol: Sendable {
    func fetchUser(_ id: String) async throws(NetworkError) -> User
    func updateUser(_ user: User) async throws(NetworkError) -> User
}

actor UserService: UserServiceProtocol {
    private let session: URLSession
    private let decoder: JSONDecoder

    init(session: URLSession = .shared) {
        self.session = session
        self.decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
    }

    func fetchUser(_ id: String) async throws(NetworkError) -> User {
        let url = URL(string: "https://api.example.com/users/\(id)")!

        do {
            let (data, response) = try await session.data(from: url)
            guard let httpResponse = response as? HTTPURLResponse,
                  200..<300 ~= httpResponse.statusCode else {
                throw NetworkError.requestFailed(statusCode: (response as? HTTPURLResponse)?.statusCode ?? 0)
            }
            return try decoder.decode(User.self, from: data)
        } catch is DecodingError {
            throw NetworkError.decodingFailed
        } catch let error as NetworkError {
            throw error
        } catch {
            throw NetworkError.requestFailed(statusCode: 0)
        }
    }
}
```

TaskGroup for Parallel Execution:
```swift
func loadDashboard() async throws -> Dashboard {
    // Parallel async let for fixed number of tasks
    async let user = api.fetchUser()
    async let posts = api.fetchPosts()
    async let notifications = api.fetchNotifications()

    return try await Dashboard(
        user: user,
        posts: posts,
        notifications: notifications
    )
}

// TaskGroup for dynamic parallelism
func loadAllUsers(_ ids: [String]) async throws -> [User] {
    try await withThrowingTaskGroup(of: User.self) { group in
        for id in ids {
            group.addTask { try await api.fetchUser(id) }
        }
        return try await group.reduce(into: []) { $0.append($1) }
    }
}
```

Actor Isolation for Thread-Safe Caching:
```swift
actor ImageCache {
    private var cache: [URL: UIImage] = [:]
    private var inProgress: [URL: Task<UIImage, Error>] = [:]

    func image(for url: URL) async throws -> UIImage {
        // Return cached image if available
        if let cached = cache[url] { return cached }

        // Return in-progress task if already downloading
        if let task = inProgress[url] {
            return try await task.value
        }

        // Start new download task
        let task = Task { try await downloadImage(url) }
        inProgress[url] = task

        do {
            let image = try await task.value
            cache[url] = image
            inProgress[url] = nil
            return image
        } catch {
            inProgress[url] = nil
            throw error
        }
    }

    private func downloadImage(_ url: URL) async throws -> UIImage {
        let (data, _) = try await URLSession.shared.data(from: url)
        guard let image = UIImage(data: data) else {
            throw ImageError.invalidData
        }
        return image
    }
}
```

### Combine Framework

Publisher and Subscriber Patterns:
```swift
import Combine

class SearchViewModel: ObservableObject {
    @Published var searchText = ""
    @Published private(set) var results: [SearchResult] = []
    @Published private(set) var isSearching = false

    private var cancellables = Set<AnyCancellable>()
    private let searchService: SearchServiceProtocol

    init(searchService: SearchServiceProtocol) {
        self.searchService = searchService
        setupSearchPipeline()
    }

    private func setupSearchPipeline() {
        $searchText
            .debounce(for: .milliseconds(300), scheduler: DispatchQueue.main)
            .removeDuplicates()
            .filter { $0.count >= 2 }
            .handleEvents(receiveOutput: { [weak self] _ in
                self?.isSearching = true
            })
            .flatMap { [searchService] query in
                searchService.search(query)
                    .catch { _ in Just([]) }
            }
            .receive(on: DispatchQueue.main)
            .sink { [weak self] results in
                self?.results = results
                self?.isSearching = false
            }
            .store(in: &cancellables)
    }
}

// Publisher extension for async/await bridge
extension Publisher {
    func async() async throws -> Output where Failure == Error {
        try await withCheckedThrowingContinuation { continuation in
            var cancellable: AnyCancellable?
            cancellable = first()
                .sink(
                    receiveCompletion: { completion in
                        if case .failure(let error) = completion {
                            continuation.resume(throwing: error)
                        }
                        cancellable?.cancel()
                    },
                    receiveValue: { value in
                        continuation.resume(returning: value)
                    }
                )
        }
    }
}
```

### XCTest Unit Testing

Async Test with MainActor:
```swift
@MainActor
final class UserViewModelTests: XCTestCase {
    var sut: UserViewModel!
    var mockAPI: MockUserAPI!

    override func setUp() {
        mockAPI = MockUserAPI()
        sut = UserViewModel(api: mockAPI)
    }

    override func tearDown() {
        sut = nil
        mockAPI = nil
    }

    func testLoadUserSuccess() async throws {
        // Given
        let expectedUser = User(id: "1", name: "Test User", email: "test@example.com")
        mockAPI.mockUser = expectedUser

        // When
        try await sut.loadUser("1")

        // Then
        XCTAssertEqual(sut.user?.id, expectedUser.id)
        XCTAssertEqual(sut.user?.name, expectedUser.name)
        XCTAssertFalse(sut.isLoading)
        XCTAssertNil(sut.errorMessage)
    }

    func testLoadUserNetworkError() async {
        // Given
        mockAPI.error = NetworkError.requestFailed(statusCode: 500)

        // When
        do {
            try await sut.loadUser("1")
            XCTFail("Expected error to be thrown")
        } catch {
            // Then
            XCTAssertNil(sut.user)
            XCTAssertNotNil(sut.errorMessage)
        }
    }
}

// Mock implementation
class MockUserAPI: UserAPIProtocol {
    var mockUser: User?
    var error: Error?

    func fetchUser(_ id: String) async throws -> User {
        if let error = error { throw error }
        guard let user = mockUser else {
            throw NetworkError.requestFailed(statusCode: 404)
        }
        return user
    }
}
```

### Swift Package Manager

Package.swift Configuration:
```swift
// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "MyApp",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(name: "MyAppCore", targets: ["MyAppCore"]),
        .executable(name: "MyAppCLI", targets: ["MyAppCLI"])
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
        .package(url: "https://github.com/onevcat/Kingfisher.git", from: "7.12.0"),
        .package(url: "https://github.com/pointfreeco/swift-composable-architecture", from: "1.15.0")
    ],
    targets: [
        .target(
            name: "MyAppCore",
            dependencies: [
                "Alamofire",
                "Kingfisher",
                .product(name: "ComposableArchitecture", package: "swift-composable-architecture")
            ],
            swiftSettings: [
                .enableExperimentalFeature("StrictConcurrency")
            ]
        ),
        .testTarget(
            name: "MyAppCoreTests",
            dependencies: ["MyAppCore"]
        )
    ]
)
```

## Advanced Patterns

For comprehensive coverage including:
- The Composable Architecture (TCA) patterns
- Advanced actor patterns and custom executors
- Keychain and secure storage
- Core Data and SwiftData integration
- Network layer with retry and caching
- CI/CD with Xcode Cloud and Fastlane

See: [reference.md](reference.md) and [examples.md](examples.md)

## Context7 Library Mappings

Core Swift:
- `/apple/swift` - Swift language and standard library
- `/apple/swift-evolution` - Swift evolution proposals
- `/apple/swift-package-manager` - SwiftPM documentation

Popular Libraries:
- `/Alamofire/Alamofire` - HTTP networking
- `/onevcat/Kingfisher` - Image downloading and caching
- `/realm/realm-swift` - Mobile database
- `/pointfreeco/swift-composable-architecture` - TCA architecture
- `/Quick/Quick` - BDD testing framework
- `/Quick/Nimble` - Matcher framework

## Works Well With

- `moai-lang-kotlin` - Android counterpart for cross-platform projects
- `moai-lang-flutter` - Flutter/Dart for cross-platform mobile
- `moai-domain-backend` - API integration and backend communication
- `moai-quality-security` - iOS security best practices
- `moai-essentials-debug` - Xcode debugging and profiling

---

Version: 1.0.0
Last Updated: 2025-12-07
Status: Production Ready
