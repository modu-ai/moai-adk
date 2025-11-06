---
name: moai-lang-swift
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Swift best practices with SwiftUI, iOS development, Swift Concurrency, and server-side Swift for 2025"
keywords: [swift, programming, ios, swiftui, server-side, concurrency, vapor, mobile]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Swift Development Mastery

**Modern Swift Development with 2025 Best Practices**

> Comprehensive Swift development guidance covering iOS/macOS applications with SwiftUI, Swift Concurrency, server-side development with Vapor, and cross-platform Swift applications using the latest tools and frameworks.

## What It Does

### iOS/macOS Development
- **Mobile App Development**: SwiftUI with modern declarative UI patterns, MVVM architecture
- **Platform Integration**: Core Data, Core Location, Camera, Push Notifications, Background Tasks
- **Performance Optimization**: Memory management, battery optimization, SwiftUI performance
- **Testing**: Unit tests, UI tests, performance tests with XCTest framework

### Server-Side Development
- **API Development**: Vapor 4, Hummingbird, or Perfect for backend services
- **Database Integration**: Fluent ORM, PostgreSQL, MongoDB with async/await
- **Real-time Communication**: WebSockets, Server-Sent Events with Swift Concurrency
- **Testing**: XCTVapor, Mocking libraries, integration testing

### Cross-Platform Development
- **SwiftUI for Multiple Platforms**: iOS, iPadOS, macOS, watchOS, visionOS
- **Shared Codebases**: Swift Package Manager, custom frameworks
- **Platform-Specific Optimizations**: Conditional compilation, platform APIs

## When to Use

### Perfect Scenarios
- **Building iOS and iPadOS applications with SwiftUI**
- **Developing macOS applications with modern Swift patterns**
- **Creating server-side APIs with Swift**
- **Implementing cross-platform Swift applications**
- **Building real-time applications with Swift Concurrency**
- **Developing watchOS and visionOS applications**
- **Creating Swift frameworks and libraries**

### Common Triggers
- "Create iOS app with Swift"
- "Build SwiftUI application"
- "Develop Swift backend API"
- "Implement Swift Concurrency"
- "Optimize Swift performance"
- "Test Swift application"
- "Swift best practices"

## Tool Version Matrix (2025-11-06)

### Core Swift
- **Swift**: 6.0 (current) / 5.10 (LTS)
- **Xcode**: 16.1 (current) / 15.4 (LTS)
- **Swift Package Manager**: Built-in with Swift 6.0
- **Platforms**: iOS 18+, iPadOS 18+, macOS 15+, watchOS 11+, visionOS 2+

### UI Frameworks
- **SwiftUI**: iOS 18.0, macOS 15.0
- **UIKit**: iOS 17.0+ (legacy support)
- **AppKit**: macOS 14.0+ (legacy support)
- **Combine**: iOS 13.0+, macOS 10.15+

### Server-Side Frameworks
- **Vapor**: 4.93.x - Web framework
- **Hummingbird**: 2.2.x - Lightweight web framework
- **Fluent**: 4.8.x - ORM for Vapor
- **PostgresNIO**: 2.12.x - PostgreSQL driver
- **MongoKitten**: 7.0.x - MongoDB driver

### Testing Tools
- **XCTest**: Built-in testing framework
- **XCUITest**: UI testing framework
- **Quick/Nimble**: BDD-style testing
- **XCTVapor**: Vapor testing utilities

### Development Tools
- **SwiftLint**: 0.54.x - Code style enforcement
- **SwiftFormat**: 0.53.x - Code formatting
- **Periphery**: 2.10.x - Unused code detection

## Ecosystem Overview

### Package Management

```bash
# Swift Package Manager initialization
swift package init --type executable
swift package init --type library

# Adding dependencies
swift package add dependency https://github.com/vapor/vapor.git --from 4.93.0
swift package add dependency https://github.com/Alamofire/Alamofire.git --from 5.9.0

# Building and testing
swift build
swift test
swift run

# Xcode project generation
swift package generate-xcodeproj
```

### Project Structure (2025 Best Practice)

```
SwiftProject/
├── Package.swift                # SPM configuration
├── Sources/
│   ├── App/                     # Main application
│   │   ├── App.swift           # Entry point
│   │   ├── ContentView.swift   # Main view
│   │   └── AppState.swift      # App state management
│   ├── Features/               # Feature modules
│   │   ├── Authentication/
│   │   │   ├── Models/
│   │   │   ├── Views/
│   │   │   ├── ViewModels/
│   │   │   └── Services/
│   │   └── UserProfile/
│   ├── Core/                   # Core utilities
│   │   ├── Extensions/
│   │   ├── Protocols/
│   │   └── Utilities/
│   └── Shared/                 # Shared code
│       ├── Models/
│       ├── Networking/
│       └── Database/
├── Tests/
│   ├── AppTests/
│   ├── FeaturesTests/
│   └── IntegrationTests/
├── Resources/                  # Resources
│   ├── Assets.xcassets
│   ├── Localizable.strings
│   └── Configuration.plist
├── .swiftlint.yml             # SwiftLint configuration
├── .swiftformat               # SwiftFormat configuration
└── Package.resolved           # Resolved dependencies
```

## Modern Development Patterns

### Swift 6.0 Concurrency Patterns

```swift
import Foundation
import SwiftConcurrency

// Async/await with structured concurrency
actor UserManager {
    private var users: [String: User] = [:]
    
    func addUser(_ user: User) async throws {
        // Actor ensures thread-safe access
        users[user.id] = user
    }
    
    func getUser(id: String) async -> User? {
        return users[id]
    }
    
    func updateProfile(id: String, profile: UserProfile) async throws -> User {
        guard var user = users[id] else {
            throw UserError.notFound
        }
        user.profile = profile
        users[id] = user
        return user
    }
    
    // Async sequence for real-time updates
    func userUpdates() -> AsyncStream<UserUpdate> {
        AsyncStream { continuation in
            let task = Task {
                for await update in userUpdateChannel {
                    continuation.yield(update)
                }
            }
            
            continuation.onTermination = { _ in
                task.cancel()
            }
        }
    }
}

// TaskGroup for concurrent operations
class DataProcessor {
    func processItems(_ items: [DataItem]) async throws -> [ProcessedItem] {
        return try await withThrowingTaskGroup(of: ProcessedItem.self) { group in
            for item in items {
                group.addTask {
                    return await self.processSingleItem(item)
                }
            }
            
            var results: [ProcessedItem] = []
            for try await result in group {
                results.append(result)
            }
            return results
        }
    }
    
    private func processSingleItem(_ item: DataItem) async -> ProcessedItem {
        // Processing logic
        return ProcessedItem(from: item)
    }
}

// AsyncImage with SwiftUI
struct UserAvatarView: View {
    let user: User
    
    var body: some View {
        AsyncImage(url: user.avatarURL) { image in
            image
                .resizable()
                .aspectRatio(contentMode: .fill)
        } placeholder: {
            ProgressView()
                .frame(width: 100, height: 100)
        }
        .frame(width: 100, height: 100)
        .clipShape(Circle())
        .overlay(
            Circle()
                .stroke(Color.blue, lineWidth: 2)
        )
    }
}

// AsyncStream for real-time data
class RealtimeDataService {
    private let continuation: AsyncStream<DataUpdate>.Continuation
    
    init() {
        let (stream, continuation) = AsyncStream<DataUpdate>.makeStream()
        self.continuation = continuation
    }
    
    var updates: AsyncStream<DataUpdate> {
        AsyncStream { continuation in
            let task = Task {
                for await update in dataChannel {
                    continuation.yield(update)
                }
            }
            
            continuation.onTermination = { _ in
                task.cancel()
            }
        }
    }
    
    func sendUpdate(_ update: DataUpdate) {
        continuation.yield(update)
    }
}
```

### Modern SwiftUI Patterns

```swift
import SwiftUI
import Combine

// MVVM with Observable macro
@Observable
class UserProfileViewModel {
    var user: User?
    var isLoading = false
    var errorMessage: String?
    
    private let userService: UserService
    private var cancellables = Set<AnyCancellable>()
    
    init(userService: UserService) {
        self.userService = userService
    }
    
    func loadUser(id: String) async {
        isLoading = true
        errorMessage = nil
        
        defer { isLoading = false }
        
        do {
            user = try await userService.getUser(id: id)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
    
    func updateUserProfile(_ profile: UserProfile) async {
        guard let user = user else { return }
        
        isLoading = true
        errorMessage = nil
        
        defer { isLoading = false }
        
        do {
            let updatedUser = try await userService.updateUser(id: user.id, profile: profile)
            self.user = updatedUser
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// Modern SwiftUI view with navigation
struct UserProfileView: View {
    @State private var viewModel: UserProfileViewModel
    @State private var showingEditSheet = false
    
    init(userId: String, userService: UserService = .shared) {
        _viewModel = State(initialValue: UserProfileViewModel(userService: userService))
    }
    
    var body: some View {
        NavigationStack {
            Group {
                if let user = viewModel.user {
                    UserProfileContent(user: user, viewModel: viewModel)
                } else if viewModel.isLoading {
                    LoadingView()
                } else if let errorMessage = viewModel.errorMessage {
                    ErrorView(message: errorMessage) {
                        Task {
                            await viewModel.loadUser(id: userId)
                        }
                    }
                } else {
                    EmptyStateView()
                }
            }
            .navigationTitle("Profile")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Edit") {
                        showingEditSheet = true
                    }
                    .disabled(viewModel.user == nil || viewModel.isLoading)
                }
            }
            .sheet(isPresented: $showingEditSheet) {
                if let user = viewModel.user {
                    EditProfileView(user: user) { profile in
                        Task {
                            await viewModel.updateUserProfile(profile)
                        }
                    }
                }
            }
            .task {
                if viewModel.user == nil {
                    await viewModel.loadUser(id: userId)
                }
            }
        }
    }
}

// Reusable view components
struct UserProfileContent: View {
    let user: User
    let viewModel: UserProfileViewModel
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Avatar section
                UserAvatarSection(user: user)
                
                // Profile information
                ProfileInfoSection(user: user)
                
                // Statistics
                UserStatsSection(user: user)
                
                // Actions
                UserActionsSection(user: user, viewModel: viewModel)
            }
            .padding()
        }
    }
}

struct UserAvatarSection: View {
    let user: User
    
    var body: some View {
        VStack(spacing: 16) {
            UserAvatarView(user: user)
                .frame(width: 120, height: 120)
            
            Text(user.name)
                .font(.title2)
                .fontWeight(.semibold)
            
            Text("@\(user.username)")
                .font(.body)
                .foregroundColor(.secondary)
        }
    }
}

// Custom modifiers
struct ProfileCardStyle: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding()
            .background(Color(.systemBackground))
            .cornerRadius(12)
            .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
    }
}

extension View {
    func profileCardStyle() -> some View {
        modifier(ProfileCardStyle())
    }
}

// State management with Environment
@main
struct MyApp: App {
    @State private var appState = AppState()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(appState)
        }
    }
}

// Observable app state
@Observable
class AppState {
    var currentUser: User?
    var isAuthenticated = false
    var theme: Theme = .system
    
    func login(user: User) {
        currentUser = user
        isAuthenticated = true
    }
    
    func logout() {
        currentUser = nil
        isAuthenticated = false
    }
}
```

### Server-Side Swift with Vapor

```swift
import Vapor
import Fluent
import PostgresNIO

// User model with Fluent
final class User: Model, Content {
    static let schema = "users"
    
    @ID(key: .id)
    var id: UUID?
    
    @Field(key: "username")
    var username: String
    
    @Field(key: "email")
    var email: String
    
    @Field(key: "password_hash")
    var passwordHash: String
    
    @Timestamp(key: "created_at", on: .create)
    var createdAt: Date?
    
    @Timestamp(key: "updated_at", on: .update)
    var updatedAt: Date?
    
    @Children(for: \.$user)
    var posts: [Post]
    
    init() { }
    
    init(id: UUID? = nil, username: String, email: String, passwordHash: String) {
        self.id = id
        self.username = username
        self.email = email
        self.passwordHash = passwordHash
    }
    
    // Public representation
    struct Public: Content {
        var id: UUID
        var username: String
        var email: String
        var createdAt: Date?
        var updatedAt: Date?
    }
    
    func convertToPublic() -> Public {
        return Public(
            id: id!,
            username: username,
            email: email,
            createdAt: createdAt,
            updatedAt: updatedAt
        )
    }
}

// Repository pattern with async/await
struct UserRepository {
    let database: Database
    
    func create(_ user: User) async throws -> User {
        try await user.save(on: database)
        return user
    }
    
    func findByID(_ id: UUID) async throws -> User? {
        try await User.query(on: database)
            .filter(\.$id == id)
            .first()
    }
    
    func findByUsername(_ username: String) async throws -> User? {
        try await User.query(on: database)
            .filter(\.$username == username)
            .first()
    }
    
    func findAll(page: Int, per: Int = 10) async throws -> [User] {
        try await User.query(on: database)
            .range(page * per ..< (page + 1) * per)
            .all()
    }
}

// Controllers with async routes
struct UserController: RouteCollection {
    func boot(routes: RoutesBuilder) throws {
        let users = routes.grouped("api/users")
        
        users.get(use: index)
        users.post(use: create)
        users.get(":userID", use: show)
        users.put(":userID", use: update)
        users.delete(":userID", use: delete)
    }
    
    @Sendable
    func index(req: Request) async throws -> [User.Public] {
        let users = try await User.query(on: req.db).all()
        return users.map { $0.convertToPublic() }
    }
    
    @Sendable
    func create(req: Request) async throws -> User.Public {
        let createUserData = try req.content.decode(CreateUser.self)
        
        guard let passwordHash = try? await req.password.hash(createUserData.password) else {
            throw Abort(.badRequest, reason: "Failed to hash password")
        }
        
        let user = User(
            username: createUserData.username,
            email: createUserData.email,
            passwordHash: passwordHash
        )
        
        try await user.save(on: req.db)
        return user.convertToPublic()
    }
    
    @Sendable
    func show(req: Request) async throws -> User.Public {
        guard let user = try await User.find(req.parameters.get("userID"), on: req.db) else {
            throw Abort(.notFound)
        }
        return user.convertToPublic()
    }
    
    @Sendable
    func update(req: Request) async throws -> User.Public {
        guard let user = try await User.find(req.parameters.get("userID"), on: req.db) else {
            throw Abort(.notFound)
        }
        
        let updateData = try req.content.decode(UpdateUser.self)
        user.username = updateData.username ?? user.username
        user.email = updateData.email ?? user.email
        
        try await user.save(on: req.db)
        return user.convertToPublic()
    }
    
    @Sendable
    func delete(req: Request) async throws -> HTTPStatus {
        guard let user = try await User.find(req.parameters.get("userID"), on: req.db) else {
            throw Abort(.notFound)
        }
        
        try await user.delete(on: req.db)
        return .noContent
    }
}

// DTOs for request/response
struct CreateUser: Content {
    var username: String
    var email: String
    var password: String
}

struct UpdateUser: Content {
    var username: String?
    var email: String?
}

// WebSocket support
struct WebSocketController: RouteCollection {
    func boot(routes: RoutesBuilder) throws {
        let websockets = routes.grouped("ws")
        websockets.webSocket("chat", onUpgrade: handleChat)
    }
    
    func handleChat(req: Request, ws: WebSocket) async {
        ws.onText { ws, text in
            // Handle incoming message
            do {
                let message = try JSONDecoder().decode(ChatMessage.self, from: text.data(using: .utf8)!)
                
                // Broadcast to all connected clients
                await broadcastMessage(message)
            } catch {
                ws.close(code: .invalidData)
            }
        }
        
        ws.onBinary { ws, data in
            // Handle binary data
        }
        
        ws.onClose.whenComplete { _ in
            // Handle connection close
        }
    }
    
    private func broadcastMessage(_ message: ChatMessage) async {
        // Implementation for broadcasting to all clients
    }
}
```

### Swift Package Manager Configuration

```swift
// Package.swift
// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "MySwiftApp",
    platforms: [
        .iOS(.v18),
        .macOS(.v15),
        .watchOS(.v11),
        .visionOS(.v2)
    ],
    products: [
        .library(name: "MySwiftApp", targets: ["MySwiftApp"]),
        .executable(name: "MyServer", targets: ["MyServer"])
    ],
    dependencies: [
        // Networking
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
        .package(url: "https://github.com/vapor/vapor.git", from: "4.93.0"),
        
        // Utilities
        .package(url: "https://github.com/apple/swift-async-algorithms", from: "1.0.0"),
        .package(url: "https://github.com/apple/swift-log.git", from: "1.6.0"),
        
        // Testing
        .package(url: "https://github.com/Quick/Quick.git", from: "7.3.0"),
        .package(url: "https://github.com/Quick/Nimble.git", from: "13.3.0")
    ],
    targets: [
        .target(
            name: "MySwiftApp",
            dependencies: [
                .product(name: "Alamofire", package: "Alamofire"),
                .product(name: "AsyncAlgorithms", package: "swift-async-algorithms"),
                .product(name: "Logging", package: "swift-log")
            ],
            path: "Sources/MySwiftApp"
        ),
        .target(
            name: "MyServer",
            dependencies: [
                .product(name: "Vapor", package: "vapor"),
                .product(name: "Fluent", package: "vapor"),
                .product(name: "FluentPostgresDriver", package: "vapor")
            ],
            path: "Sources/MyServer"
        ),
        .testTarget(
            name: "MySwiftAppTests",
            dependencies: [
                "MySwiftApp",
                "Quick",
                "Nimble"
            ],
            path: "Tests/MySwiftAppTests"
        ),
        .testTarget(
            name: "MyServerTests",
            dependencies: [
                "MyServer",
                .product(name: "XCTVapor", package: "vapor")
            ],
            path: "Tests/MyServerTests"
        )
    ]
)
```

## Performance Considerations

### Memory Management

```swift
// Efficient memory usage with weak references
class ImageCache {
    private var cache: [String: UIImage] = [:]
    private let queue = DispatchQueue(label: "com.app.imagecache", attributes: .concurrent)
    
    func setImage(_ image: UIImage, for key: String) {
        queue.async(flags: .barrier) {
            self.cache[key] = image
        }
    }
    
    func getImage(for key: String) -> UIImage? {
        return queue.sync {
            return cache[key]
        }
    }
    
    func clearCache() {
        queue.async(flags: .barrier) {
            self.cache.removeAll()
        }
    }
}

// Lazy loading and memory optimization
class DataViewController: UIViewController {
    private lazy var dataLoader: DataLoader = {
        return DataLoader(configuration: .default)
    }()
    
    private lazy var imageCache: NSCache<NSString, UIImage> = {
        let cache = NSCache<NSString, UIImage>()
        cache.countLimit = 100
        cache.totalCostLimit = 1024 * 1024 * 50 // 50MB
        return cache
    }()
    
    func loadImage(from url: URL) async -> UIImage? {
        let key = url.absoluteString as NSString
        
        if let cachedImage = imageCache.object(forKey: key) {
            return cachedImage
        }
        
        do {
            let (data, _) = try await dataLoader.data(from: url)
            if let image = UIImage(data: data) {
                imageCache.setObject(image, forKey: key)
                return image
            }
        } catch {
            print("Failed to load image: \(error)")
        }
        
        return nil
    }
}

// Efficient data structures
class EfficientDataProcessor {
    private var observations: [Observation] = []
    private var processedData: ProcessedData?
    
    // Use lazy evaluation for expensive computations
    private lazy var statistics: Statistics = {
        calculateStatistics(from: observations)
    }()
    
    func addObservation(_ observation: Observation) {
        observations.append(observation)
        // Invalidate cached data when underlying data changes
        processedData = nil
    }
    
    var data: ProcessedData {
        if let cached = processedData {
            return cached
        }
        
        let result = processData(observations)
        processedData = result
        return result
    }
    
    var stats: Statistics {
        return statistics
    }
    
    private func calculateStatistics(from observations: [Observation]) -> Statistics {
        // Expensive calculation
        return Statistics(observations: observations)
    }
    
    private func processData(_ observations: [Observation]) -> ProcessedData {
        // Data processing logic
        return ProcessedData(observations: observations)
    }
}
```

### SwiftUI Performance Optimization

```swift
// Performance optimizations for SwiftUI
struct OptimizedListView: View {
    @State private var items: [Item] = []
    
    var body: some View {
        // Use OnAppear for data loading instead of init
        List(items) { item in
            ItemRow(item: item)
                .id(item.id) // Stable identity for view recycling
        }
        .listStyle(.plain)
        .onAppear {
            loadItems()
        }
    }
    
    private func loadItems() {
        // Load data asynchronously
        Task {
            let loadedItems = await loadItemsFromAPI()
            
            // Update UI on main thread
            await MainActor.run {
                items = loadedItems
            }
        }
    }
}

struct ItemRow: View {
    let item: Item
    
    var body: some View {
        HStack {
            AsyncImage(url: item.imageURL) { image in
                image
                    .resizable()
                    .aspectRatio(contentMode: .fill)
            } placeholder: {
                Color.gray.opacity(0.3)
            }
            .frame(width: 50, height: 50)
            .clipped()
            
            VStack(alignment: .leading, spacing: 4) {
                Text(item.title)
                    .font(.headline)
                    .lineLimit(1)
                
                Text(item.description)
                    .font(.body)
                    .foregroundColor(.secondary)
                    .lineLimit(2)
            }
            
            Spacer()
        }
        .padding(.vertical, 8)
    }
}

// Efficient state management
@Observable
class PerformanceOptimizedViewModel {
    // Use @Published sparingly and only when UI needs to react
    @Published private(set) var items: [Item] = []
    @Published var isLoading = false
    @Published var error: Error?
    
    private let service: APIService
    private var currentPage = 0
    private var hasMorePages = true
    
    init(service: APIService) {
        self.service = service
    }
    
    func loadMoreItems() async {
        guard !isLoading, hasMorePages else { return }
        
        isLoading = true
        
        defer { isLoading = false }
        
        do {
            let newItems = try await service.fetchItems(page: currentPage + 1)
            
            await MainActor.run {
                items.append(contentsOf: newItems)
                currentPage += 1
                hasMorePages = !newItems.isEmpty
            }
        } catch {
            await MainActor.run {
                self.error = error
            }
        }
    }
    
    func refresh() async {
        currentPage = 0
        hasMorePages = true
        items = []
        error = nil
        
        await loadMoreItems()
    }
}
```

### Server-Side Performance

```swift
// Connection pooling and optimization
final class DatabaseService {
    private let pool: EventLoopGroupConnectionPool<PostgresConnectionSource>
    
    init(configuration: PostgresConfiguration) {
        let source = PostgresConnectionSource(configuration: configuration)
        pool = EventLoopGroupConnectionPool(
            source: source,
            maxConnectionsPerEventLoop: 10,
            on: MultiThreadedEventLoopGroup(numberOfThreads: 4)
        )
    }
    
    func execute<T>(_ query: SQLQuery, on eventLoop: EventLoop) async throws -> [T] {
        return try await withThrowingCheckedContinuation { continuation in
            pool.withConnection(on: eventLoop) { connection in
                connection.query(query)
                    .eachRow { row in
                        // Process each row
                    }
                    .flatMapThrowing { rows in
                        // Convert rows to result type
                        return rows.map { self.convertRow($0) }
                    }
                    .map { continuation.resume(returning: $0) }
                    .recover { continuation.resume(throwing: $0) }
            }
        }
    }
}

// Caching with async/await
actor CacheService {
    private var cache: [String: CacheEntry] = [:]
    private let cleanupTimer: Timer
    
    init() {
        cleanupTimer = Timer.scheduledTimer(withTimeInterval: 300, repeats: true) { [weak self] _ in
            Task {
                await self?.cleanup()
            }
        }
    }
    
    func get(_ key: String) async -> Data? {
        guard let entry = cache[key], !entry.isExpired else {
            return nil
        }
        return entry.data
    }
    
    func set(_ data: Data, forKey key: String, ttl: TimeInterval = 3600) async {
        cache[key] = CacheEntry(data: data, expiresAt: Date().addingTimeInterval(ttl))
    }
    
    private func cleanup() async {
        let expiredKeys = cache.compactMapValues { entry in
            entry.isExpired ? nil : entry
        }.keys
        
        for key in expiredKeys {
            cache.removeValue(forKey: key)
        }
    }
    
    struct CacheEntry {
        let data: Data
        let expiresAt: Date
        
        var isExpired: Bool {
            Date() > expiresAt
        }
    }
}
```

## Testing Strategy

### XCTest Configuration

```swift
// XCTest-based testing
class UserServiceTests: XCTestCase {
    var userService: UserService!
    var mockAPIClient: MockAPIClient!
    
    override func setUpWithError() throws {
        mockAPIClient = MockAPIClient()
        userService = UserService(apiClient: mockAPIClient)
    }
    
    override func tearDownWithError() throws {
        userService = nil
        mockAPIClient = nil
    }
    
    func testCreateUser_Success() async throws {
        // Given
        let userData = CreateUserData(username: "testuser", email: "test@example.com")
        let expectedUser = User(id: UUID(), username: "testuser", email: "test@example.com")
        
        mockAPIClient.createUserResult = .success(expectedUser)
        
        // When
        let result = try await userService.createUser(userData)
        
        // Then
        XCTAssertEqual(result.id, expectedUser.id)
        XCTAssertEqual(result.username, expectedUser.username)
        XCTAssertEqual(result.email, expectedUser.email)
        XCTAssertEqual(mockAPIClient.createUserCallCount, 1)
    }
    
    func testCreateUser_Failure() async {
        // Given
        let userData = CreateUserData(username: "testuser", email: "test@example.com")
        mockAPIClient.createUserResult = .failure(APIError.networkError)
        
        // When & Then
        do {
            _ = try await userService.createUser(userData)
            XCTFail("Expected error to be thrown")
        } catch {
            XCTAssertTrue(error is APIError)
        }
    }
    
    func testPerformanceLoadUsers() async throws {
        // Given
        let users = (0..<1000).map { User(id: UUID(), username: "user\($0)", email: "user\($0)@example.com") }
        mockAPIClient.loadUsersResult = .success(users)
        
        // Measure performance
        measure {
            Task {
                _ = try! await userService.loadUsers()
            }
        }
    }
}

// Mock objects for testing
class MockAPIClient: APIClientProtocol {
    var createUserResult: Result<User, APIError>!
    var loadUsersResult: Result<[User], APIError>!
    var createUserCallCount = 0
    
    func createUser(_ data: CreateUserData) async throws -> User {
        createUserCallCount += 1
        switch createUserResult {
        case .success(let user):
            return user
        case .failure(let error):
            throw error
        }
    }
    
    func loadUsers() async throws -> [User] {
        switch loadUsersResult {
        case .success(let users):
            return users
        case .failure(let error):
            throw error
        }
    }
}

// Quick/Nimble BDD-style testing
import Quick
import Nimble

class UserServiceSpec: QuickSpec {
    override func spec() {
        var userService: UserService!
        var mockAPIClient: MockAPIClient!
        
        beforeEach {
            mockAPIClient = MockAPIClient()
            userService = UserService(apiClient: mockAPIClient)
        }
        
        describe("createUser") {
            context("when API call succeeds") {
                it("should return the created user") {
                    let userData = CreateUserData(username: "testuser", email: "test@example.com")
                    let expectedUser = User(id: UUID(), username: "testuser", email: "test@example.com")
                    mockAPIClient.createUserResult = .success(expectedUser)
                    
                    waitUntil { done in
                        Task {
                            do {
                                let result = try await userService.createUser(userData)
                                expect(result.username).to(equal(expectedUser.username))
                                expect(result.email).to(equal(expectedUser.email))
                                done()
                            } catch {
                                fail("Unexpected error: \(error)")
                            }
                        }
                    }
                }
            }
            
            context("when API call fails") {
                it("should throw an error") {
                    let userData = CreateUserData(username: "testuser", email: "test@example.com")
                    mockAPIClient.createUserResult = .failure(APIError.networkError)
                    
                    waitUntil { done in
                        Task {
                            do {
                                _ = try await userService.createUser(userData)
                                fail("Expected error to be thrown")
                            } catch {
                                expect(error).to(beAKindOf(APIError.self))
                                done()
                            }
                        }
                    }
                }
            }
        }
    }
}
```

### SwiftUI Testing

```swift
import XCTest
import SwiftUI
@testable import MySwiftApp

class UserProfileViewTests: XCTestCase {
    
    func testUserProfileView_DisplaysUserInfo() {
        // Given
        let user = User(id: UUID(), username: "testuser", name: "Test User", email: "test@example.com")
        let userService = MockUserService()
        userService.mockUser = user
        
        // When
        let view = UserProfileView(userId: user.id.uuidString, userService: userService)
        
        // Then
        let hostingController = UIHostingController(rootView: view)
        hostingController.loadViewIfNeeded()
        
        // Test view content
        // Note: This is a simplified example - in practice, you'd use view inspection or snapshot testing
        XCTAssertNotNil(hostingController.view)
    }
}

// ViewInspector for SwiftUI testing (third-party library)
import ViewInspector

extension UserProfileView: Inspectable { }

class UserProfileViewInspectorTests: XCTestCase {
    
    func testUserProfileView_WhenLoaded_ShowsUserData() throws {
        // Given
        let user = User(id: UUID(), username: "testuser", name: "Test User", email: "test@example.com")
        let userService = MockUserService()
        userService.mockUser = user
        
        // When
        let view = UserProfileView(userId: user.id.uuidString, userService: userService)
        let inspectedView = try view.inspect()
        
        // Then
        let navigationStack = try inspectedView.navigationStack()
        let navigationTitle = try navigationStack.navigationBarTitleLabel().string()
        XCTAssertEqual(navigationTitle, "Profile")
    }
}
```

### Server-Side Testing

```swift
import XCTVapor
import XCTest
@testable import MyServer

final class UserControllerTests: XCTestCase {
    
    func testCreateUser() async throws {
        let app = Application(.testing)
        try configure(app)
        
        // Test data
        let userData = CreateUserData(username: "testuser", email: "test@example.com", password: "password123")
        
        try app.test(.POST, "/api/users", beforeRequest: { req in
            try req.content.encode(userData)
        }, afterResponse: { res in
            XCTAssertEqual(res.status, .ok)
            let user = try res.content.decode(User.Public.self)
            XCTAssertEqual(user.username, "testuser")
            XCTAssertEqual(user.email, "test@example.com")
        })
    }
    
    func testGetUser() async throws {
        let app = Application(.testing)
        try configure(app)
        
        // Create a user first
        let userData = CreateUserData(username: "testuser", email: "test@example.com", password: "password123")
        
        try app.test(.POST, "/api/users", beforeRequest: { req in
            try req.content.encode(userData)
        }, afterResponse: { createRes in
            XCTAssertEqual(createRes.status, .ok)
            let createdUser = try createRes.content.decode(User.Public.self)
            
            // Test getting the user
            try app.test(.GET, "/api/users/\(createdUser.id)", afterResponse: { getRes in
                XCTAssertEqual(getRes.status, .ok)
                let user = try getRes.content.decode(User.Public.self)
                XCTAssertEqual(user.id, createdUser.id)
                XCTAssertEqual(user.username, "testuser")
            })
        })
    }
    
    func testWebSocketConnection() async throws {
        let app = Application(.testing)
        try configure(app)
        
        try app.test(.websocket, "/ws/chat") { conn in
            try conn.send("Hello, server!")
            try conn.expectString(text: "Hello, client!")
            try conn.close()
        }
    }
}
```

## Security Best Practices

### Input Validation and Sanitization

```swift
// Input validation with property wrappers
@propertyWrapper
struct ValidatedEmail: Decodable {
    private let _value: String
    
    var wrappedValue: String {
        return _value
    }
    
    init(wrappedValue: String) throws {
        guard Self.isValidEmail(wrappedValue) else {
            throw ValidationError.invalidEmail
        }
        _value = wrappedValue.lowercased()
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let value = try container.decode(String.self)
        try self.init(wrappedValue: value)
    }
    
    private static func isValidEmail(_ email: String) -> Bool {
        let emailRegex = #"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"#
        return email.range(of: emailRegex, options: .regularExpression) != nil
    }
}

struct CreateUserRequest: Decodable {
    @ValidatedEmail var email: String
    @ValidatedUsername var username: String
    @SecurePassword var password: String
}

@propertyWrapper
struct SecurePassword: Decodable {
    private let _value: String
    
    var wrappedValue: String {
        return _value
    }
    
    init(wrappedValue: String) throws {
        guard Self.isValidPassword(wrappedValue) else {
            throw ValidationError.weakPassword
        }
        _value = wrappedValue
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        let value = try container.decode(String.self)
        try self.init(wrappedValue: value)
    }
    
    private static func isValidPassword(_ password: String) -> Bool {
        // At least 8 characters, one uppercase, one lowercase, one digit, one special character
        let passwordRegex = #"^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$"#
        return password.range(of: passwordRegex, options: .regularExpression) != nil
    }
}

enum ValidationError: Error, LocalizedError {
    case invalidEmail
    case weakPassword
    
    var errorDescription: String? {
        switch self {
        case .invalidEmail:
            return "Invalid email format"
        case .weakPassword:
            return "Password must be at least 8 characters and contain uppercase, lowercase, digit, and special character"
        }
    }
}
```

### Authentication and Authorization

```swift
// JWT authentication middleware
struct JWTMiddleware: AsyncMiddleware {
    let jwtSecret: String
    
    func respond(to request: Request, chainingTo next: AsyncResponder) async throws -> Response {
        guard let token = request.headers.bearerAuthorization?.token else {
            throw Abort(.unauthorized, reason: "Missing authorization token")
        }
        
        do {
            let payload = try JWT<AuthPayload>(from: token, verifiedUsing: .hs256(key: jwtSecret))
            request.auth.login(payload.user)
            return try await next.respond(to: request)
        } catch {
            throw Abort(.unauthorized, reason: "Invalid token")
        }
    }
}

struct AuthPayload: JWTPayload {
    var user: User
    var expirationTime: ExpirationClaim
    
    func verify(using algorithm: some JWTAlgorithm) throws {
        try expirationTime.verifyNotExpired()
    }
}

// Role-based access control
enum UserRole: String, Codable {
    case admin
    case moderator
    case user
    case guest
}

extension User {
    func hasRole(_ role: UserRole) -> Bool {
        return self.role == role
    }
    
    func canAccessResource(_ resource: Resource) -> Bool {
        switch resource.type {
        case .adminOnly:
            return hasRole(.admin)
        case .moderatorAndAbove:
            return hasRole(.admin) || hasRole(.moderator)
        case .userAndAbove:
            return hasRole(.admin) || hasRole(.moderator) || hasRole(.user)
        case .public:
            return true
        }
    }
}

// Request extension for authentication
extension Request {
    var authenticatedUser: User? {
        return auth.get(User.self)
    }
    
    func requireAuthentication() throws -> User {
        guard let user = authenticatedUser else {
            throw Abort(.unauthorized, reason: "Authentication required")
        }
        return user
    }
    
    func requireRole(_ role: UserRole) throws -> User {
        let user = try requireAuthentication()
        guard user.hasRole(role) else {
            throw Abort(.forbidden, reason: "Insufficient permissions")
        }
        return user
    }
}
```

### Security Headers and CORS

```swift
// Security headers middleware
struct SecurityHeadersMiddleware: AsyncMiddleware {
    func respond(to request: Request, chainingTo next: AsyncResponder) async throws -> Response {
        let response = try await next.respond(to: request)
        
        // Add security headers
        response.headers.add(name: .xContentTypeOptions, value: "nosniff")
        response.headers.add(name: .xFrameOptions, value: "DENY")
        response.headers.add(name: .xXSSProtection, value: "1; mode=block")
        response.headers.add(name: .strictTransportSecurity, value: "max-age=31536000; includeSubDomains")
        response.headers.add(name: .contentSecurityPolicy, value: "default-src 'self'")
        
        return response
    }
}

// CORS configuration
struct CORSMiddleware: AsyncMiddleware {
    private let allowedOrigins: [String]
    private let allowedMethods: [HTTPMethod]
    private let allowedHeaders: [String]
    
    init(allowedOrigins: [String] = ["*"], allowedMethods: [HTTPMethod] = [.GET, .POST, .PUT, .DELETE], allowedHeaders: [String] = ["Content-Type", "Authorization"]) {
        self.allowedOrigins = allowedOrigins
        self.allowedMethods = allowedMethods
        self.allowedHeaders = allowedHeaders
    }
    
    func respond(to request: Request, chainingTo next: AsyncResponder) async throws -> Response {
        let response = try await next.respond(to: request)
        
        // Add CORS headers
        if let origin = request.headers.first(name: .origin) {
            if allowedOrigins.contains("*") || allowedOrigins.contains(origin) {
                response.headers.add(name: .accessControlAllowOrigin, value: origin)
            }
        }
        
        response.headers.add(name: .accessControlAllowMethods, value: allowedMethods.map { $0.string }.joined(separator: ", "))
        response.headers.add(name: .accessControlAllowHeaders, value: allowedHeaders.joined(separator: ", "))
        
        return response
    }
}
```

## Integration Patterns

### Core Data Integration

```swift
import CoreData
import SwiftUI

// Core Data manager with async/await
actor CoreDataManager {
    static let shared = CoreDataManager()
    
    lazy var persistentContainer: NSPersistentContainer = {
        let container = NSPersistentContainer(name: "DataModel")
        container.loadPersistentStores { _, error in
            if let error = error {
                fatalError("Core Data error: \(error.localizedDescription)")
            }
        }
        container.viewContext.automaticallyMergesChangesFromParent = true
        return container
    }()
    
    var viewContext: NSManagedObjectContext {
        persistentContainer.viewContext
    }
    
    private var backgroundContext: NSManagedObjectContext {
        persistentContainer.newBackgroundContext()
    }
    
    func fetchUsers() async -> [User] {
        return await withCheckedContinuation { continuation in
            let request: NSFetchRequest<User> = User.fetchRequest()
            
            backgroundContext.perform {
                do {
                    let users = try self.backgroundContext.fetch(request)
                    continuation.resume(returning: users)
                } catch {
                    continuation.resume(returning: [])
                }
            }
        }
    }
    
    func saveUser(_ user: User) async {
        await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
            backgroundContext.perform {
                do {
                    try self.backgroundContext.save()
                    continuation.resume()
                } catch {
                    print("Failed to save user: \(error)")
                    continuation.resume()
                }
            }
        }
    }
    
    func deleteUser(_ user: User) async {
        await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
            backgroundContext.perform {
                self.backgroundContext.delete(user)
                do {
                    try self.backgroundContext.save()
                    continuation.resume()
                } catch {
                    print("Failed to delete user: \(error)")
                    continuation.resume()
                }
            }
        }
    }
}

// SwiftUI Core Data integration
struct CoreDataUserListView: View {
    @Environment(\.managedObjectContext) private var viewContext
    @FetchRequest(
        sortDescriptors: [NSSortDescriptor(keyPath: \User.createdAt, ascending: true)],
        animation: .default)
    private var users: FetchedResults<User>
    
    var body: some View {
        List {
            ForEach(users, id: \.objectID) { user in
                UserRow(user: user)
            }
            .onDelete(perform: deleteUsers)
        }
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button(action: addUser) {
                    Label("Add User", systemImage: "plus")
                }
            }
        }
    }
    
    private func addUser() {
        withAnimation {
            let newUser = User(context: viewContext)
            newUser.name = "New User"
            newUser.createdAt = Date()
            
            saveContext()
        }
    }
    
    private func deleteUsers(offsets: IndexSet) {
        withAnimation {
            offsets.map { users[$0] }.forEach(viewContext.delete)
            saveContext()
        }
    }
    
    private func saveContext() {
        do {
            try viewContext.save()
        } catch {
            let nsError = error as NSError
            fatalError("Unresolved error \(nsError), \(nsError.userInfo)")
        }
    }
}
```

### Networking with Async/Await

```swift
import Foundation

// Modern networking client
class APIClient: APIClientProtocol {
    private let session: URLSession
    private let baseURL: URL
    private let decoder: JSONDecoder
    
    init(baseURL: URL = URL(string: "https://api.example.com")!) {
        self.baseURL = baseURL
        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601
        self.decoder.keyDecodingStrategy = .convertFromSnakeCase
        
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        config.waitsForConnectivity = true
        
        self.session = URLSession(configuration: config)
    }
    
    func request<T: Decodable>(_ endpoint: APIEndpoint, responseType: T.Type) async throws -> T {
        let request = try buildRequest(for: endpoint)
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        
        guard 200...299 ~= httpResponse.statusCode else {
            throw APIError.serverError(statusCode: httpResponse.statusCode)
        }
        
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decodingError(error)
        }
    }
    
    func uploadData<T: Decodable>(_ data: Data, to endpoint: APIEndpoint, responseType: T.Type) async throws -> T {
        var request = try buildRequest(for: endpoint)
        request.httpMethod = "POST"
        request.setValue("application/octet-stream", forHTTPHeaderField: "Content-Type")
        request.httpBody = data
        
        let (responseData, response) = try await session.upload(for: request, from: data)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        
        guard 200...299 ~= httpResponse.statusCode else {
            throw APIError.serverError(statusCode: httpResponse.statusCode)
        }
        
        do {
            return try decoder.decode(T.self, from: responseData)
        } catch {
            throw APIError.decodingError(error)
        }
    }
    
    private func buildRequest(for endpoint: APIEndpoint) throws -> URLRequest {
        var components = URLComponents(url: baseURL.appendingPathComponent(endpoint.path), resolvingAgainstBaseURL: false)!
        
        if !endpoint.parameters.isEmpty {
            components.queryItems = endpoint.parameters.map { key, value in
                URLQueryItem(name: key, value: "\(value)")
            }
        }
        
        var request = URLRequest(url: components.url!)
        request.httpMethod = endpoint.method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let body = endpoint.body {
            request.httpBody = try JSONSerialization.data(withJSONObject: body)
        }
        
        return request
    }
}

// API endpoint configuration
enum APIEndpoint {
    case getUser(id: String)
    case getUsers(page: Int, limit: Int)
    case createUser(CreateUserData)
    case updateUser(id: String, UpdateUserData)
    case deleteUser(id: String)
    
    var path: String {
        switch self {
        case .getUser(let id):
            return "users/\(id)"
        case .getUsers:
            return "users"
        case .createUser:
            return "users"
        case .updateUser(let id, _):
            return "users/\(id)"
        case .deleteUser(let id):
            return "users/\(id)"
        }
    }
    
    var method: HTTPMethod {
        switch self {
        case .getUser, .getUsers:
            return .GET
        case .createUser:
            return .POST
        case .updateUser:
            return .PUT
        case .deleteUser:
            return .DELETE
        }
    }
    
    var parameters: [String: Any] {
        switch self {
        case .getUsers(let page, let limit):
            return ["page": page, "limit": limit]
        default:
            return [:]
        }
    }
    
    var body: [String: Any]? {
        switch self {
        case .createUser(let userData):
            return try? userData.asDictionary()
        case .updateUser(_, let userData):
            return try? userData.asDictionary()
        default:
            return nil
        }
    }
}

enum HTTPMethod: String {
    case GET = "GET"
    case POST = "POST"
    case PUT = "PUT"
    case DELETE = "DELETE"
}

enum APIError: Error, LocalizedError {
    case invalidURL
    case invalidResponse
    case serverError(statusCode: Int)
    case decodingError(Error)
    case networkError(Error)
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .invalidResponse:
            return "Invalid response"
        case .serverError(let statusCode):
            return "Server error with status code: \(statusCode)"
        case .decodingError(let error):
            return "Decoding error: \(error.localizedDescription)"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        }
    }
}
```

## Modern Development Workflow

### Xcode Configuration

```swift
// Project-level configuration for modern Swift development

// Info.plist additions for security and privacy
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <false/>
    <key>NSExceptionDomains</key>
    <dict>
        <key>api.example.com</key>
        <dict>
            <key>NSExceptionAllowsInsecureHTTPLoads</key>
            <true/>
            <key>NSExceptionMinimumTLSVersion</key>
            <string>TLSv1.2</string>
        </dict>
    </dict>
</dict>

<key>NSCameraUsageDescription</key>
<string>This app needs camera access for profile photos</string>

<key>NSLocationWhenInUseUsageDescription</key>
<string>This app needs location access for location-based features</string>

<key>NSUserTrackingUsageDescription</key>
<string>This app uses tracking for personalized advertising</string>
```

### SwiftLint Configuration

```yaml
# .swiftlint.yml
excluded:
  - Carthage
  - Pods
  - build
  - .build

opt_in_rules:
  - empty_count
  - force_unwrapping
  - implicitly_unwrapped_optional

disabled_rules:
  - trailing_whitespace
  - line_length

line_length:
  warning: 120
  error: 150

function_body_length:
  warning: 50
  error: 100

type_body_length:
  warning: 300
  error: 500

file_length:
  warning: 400
  error: 800

cyclomatic_complexity:
  warning: 10
  error: 20

custom_rules:
  no_console_log:
    name: "No Console Logging"
    regex: '(print|NSLog|debugPrint)\('
    message: "Console logging should be removed in production."
    severity: warning
```

### CI/CD Configuration

```yaml
# .github/workflows/swift.yml
name: Swift CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: macos-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Select Xcode
      uses: maxim-lobanov/setup-xcode@v1
      with:
        xcode-version: latest-stable
    
    - name: Cache Swift Package Manager
      uses: actions/cache@v3
      with:
        path: .build
        key: ${{ runner.os }}-spm-${{ hashFiles('**/Package.resolved') }}
    
    - name: Build
      run: swift build
    
    - name: Run tests
      run: swift test --enable-code-coverage
    
    - name: Generate code coverage report
      run: xcrun llvm-cov report -build-path .build -use-llvm --json > coverage.json
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.json
    
    - name: Run SwiftLint
      run: |
        mint run swiftlint/swiftlint swiftlint
        mint run swiftlint/swiftlint swiftlint --strict
    
    - name: Run SwiftFormat
      run: |
        mint run swiftformat/swiftformat swiftformat --lint --strict .

  ios_test:
    runs-on: macos-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Select Xcode
      uses: maxim-lobanov/setup-xcode@v1
      with:
        xcode-version: latest-stable
    
    - name: Build iOS App
      run: xcodebuild -scheme MyApp -destination 'platform=iOS Simulator,name=iPhone 15,OS=latest' clean build
    
    - name: Run iOS Tests
      run: xcodebuild -scheme MyApp -destination 'platform=iOS Simulator,name=iPhone 15,OS=latest' test
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Swift Target**: 6.0 with modern SwiftUI, Swift Concurrency, and server-side Swift  

This skill provides comprehensive Swift development guidance with 2025 best practices, covering everything from iOS/macOS applications with SwiftUI to server-side development with Vapor and modern concurrency patterns.
