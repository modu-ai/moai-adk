# Swift Development Reference

## Platform Version Matrix

### Swift 6.0 (iOS 18+, macOS 15+)
- Release: September 2025
- Xcode: 16.0+
- Minimum Deployment: iOS 15.0+ (recommended iOS 17.0+)
- Key Features:
  - Complete data-race safety by default
  - Typed throws for precise error handling
  - Custom actor executors for concurrency control
  - Embedded Swift for IoT and embedded systems
  - Improved C++ interoperability

### Swift 5.10 (iOS 17+, macOS 14+)
- Release: March 2024
- Xcode: 15.3+
- Key Features:
  - Complete strict concurrency checking
  - @Observable macro
  - NavigationStack improvements
  - Swift macros

## Context7 Library Mappings

### Core Swift Libraries
```
/apple/swift                    - Swift language and standard library
/apple/swift-package-manager    - SwiftPM package management
/apple/swift-nio                - Non-blocking I/O framework
/apple/swift-async-algorithms   - Async sequence algorithms
/apple/swift-collections        - Additional collection types
/apple/swift-algorithms         - Sequence/collection algorithms
```

### Networking Libraries
```
/Alamofire/Alamofire            - HTTP networking library
/Moya/Moya                      - Network abstraction layer
/kean/Nuke                      - Image loading and caching
/onevcat/Kingfisher             - Image downloading and caching
```

### Database Libraries
```
/realm/realm-swift              - Mobile database
/groue/GRDB.swift               - SQLite toolkit
/stephencelis/SQLite.swift      - Type-safe SQLite wrapper
```

### Architecture Libraries
```
/pointfreeco/swift-composable-architecture - TCA architecture
/ReactiveX/RxSwift              - Reactive programming
/CombineCommunity/CombineExt    - Combine extensions
```

### Testing Libraries
```
/Quick/Quick                    - BDD testing framework
/Quick/Nimble                   - Matcher framework
/pointfreeco/swift-snapshot-testing - Snapshot testing
/nalexn/ViewInspector           - SwiftUI view testing
```

### UI Libraries
```
/SnapKit/SnapKit                - Auto Layout DSL
/airbnb/lottie-ios              - Animation library
/danielgindi/Charts             - Charting library
/SwiftUIX/SwiftUIX              - SwiftUI extensions
```

## Architecture Patterns

### SwiftUI + TCA (The Composable Architecture)

Feature Definition:
```swift
import ComposableArchitecture

@Reducer
struct UserFeature {
    @ObservableState
    struct State: Equatable {
        var user: User?
        var isLoading = false
        var error: String?
        @Presents var alert: AlertState<Action.Alert>?
    }

    enum Action: BindableAction {
        case binding(BindingAction<State>)
        case loadUser(String)
        case userLoaded(Result<User, Error>)
        case logout
        case alert(PresentationAction<Alert>)

        enum Alert: Equatable {
            case confirmLogout
        }
    }

    @Dependency(\.userClient) var userClient
    @Dependency(\.mainQueue) var mainQueue

    var body: some ReducerOf<Self> {
        BindingReducer()
        Reduce { state, action in
            switch action {
            case .binding:
                return .none

            case let .loadUser(id):
                state.isLoading = true
                state.error = nil
                return .run { send in
                    await send(.userLoaded(
                        Result { try await userClient.fetch(id) }
                    ))
                }

            case let .userLoaded(.success(user)):
                state.isLoading = false
                state.user = user
                return .none

            case let .userLoaded(.failure(error)):
                state.isLoading = false
                state.error = error.localizedDescription
                return .none

            case .logout:
                state.alert = AlertState {
                    TextState("Confirm Logout")
                } actions: {
                    ButtonState(role: .destructive, action: .confirmLogout) {
                        TextState("Logout")
                    }
                    ButtonState(role: .cancel) {
                        TextState("Cancel")
                    }
                } message: {
                    TextState("Are you sure you want to logout?")
                }
                return .none

            case .alert(.presented(.confirmLogout)):
                state.user = nil
                return .none

            case .alert:
                return .none
            }
        }
        .ifLet(\.$alert, action: \.alert)
    }
}
```

View Implementation:
```swift
struct UserView: View {
    @Bindable var store: StoreOf<UserFeature>

    var body: some View {
        WithPerceptionTracking {
            content
                .task { store.send(.loadUser("current")) }
                .alert($store.scope(state: \.alert, action: \.alert))
        }
    }

    @ViewBuilder
    private var content: some View {
        if store.isLoading {
            ProgressView()
        } else if let error = store.error {
            ErrorView(message: error) {
                store.send(.loadUser("current"))
            }
        } else if let user = store.user {
            UserProfileContent(user: user) {
                store.send(.logout)
            }
        }
    }
}
```

### MVVM with @Observable

ViewModel Pattern:
```swift
@Observable
@MainActor
final class PostListViewModel {
    private(set) var posts: [Post] = []
    private(set) var isLoading = false
    private(set) var error: Error?
    private(set) var hasMorePages = true

    private var currentPage = 1
    private let pageSize = 20
    private let postService: PostServiceProtocol

    init(postService: PostServiceProtocol) {
        self.postService = postService
    }

    func loadPosts() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil
        currentPage = 1

        do {
            let result = try await postService.fetchPosts(page: currentPage, limit: pageSize)
            posts = result.posts
            hasMorePages = result.hasMore
        } catch {
            self.error = error
        }

        isLoading = false
    }

    func loadMoreIfNeeded(currentItem: Post?) async {
        guard let item = currentItem,
              !isLoading,
              hasMorePages else { return }

        let thresholdIndex = posts.index(posts.endIndex, offsetBy: -5)
        guard posts.firstIndex(where: { $0.id == item.id }) ?? 0 >= thresholdIndex else {
            return
        }

        await loadNextPage()
    }

    private func loadNextPage() async {
        isLoading = true
        currentPage += 1

        do {
            let result = try await postService.fetchPosts(page: currentPage, limit: pageSize)
            posts.append(contentsOf: result.posts)
            hasMorePages = result.hasMore
        } catch {
            currentPage -= 1
            self.error = error
        }

        isLoading = false
    }
}
```

View with Pagination:
```swift
struct PostListView: View {
    @State private var viewModel: PostListViewModel

    init(postService: PostServiceProtocol) {
        _viewModel = State(initialValue: PostListViewModel(postService: postService))
    }

    var body: some View {
        NavigationStack {
            List {
                ForEach(viewModel.posts) { post in
                    PostRow(post: post)
                        .task {
                            await viewModel.loadMoreIfNeeded(currentItem: post)
                        }
                }

                if viewModel.isLoading {
                    HStack {
                        Spacer()
                        ProgressView()
                        Spacer()
                    }
                }
            }
            .refreshable {
                await viewModel.loadPosts()
            }
            .task {
                await viewModel.loadPosts()
            }
            .navigationTitle("Posts")
        }
    }
}
```

## Concurrency Deep Dive

### Custom Actor Executors

Serial Executor for Background Processing:
```swift
actor BackgroundProcessor {
    private let executor: SerialDispatchQueueExecutor

    nonisolated var unownedExecutor: UnownedSerialExecutor {
        executor.asUnownedSerialExecutor()
    }

    init(label: String) {
        executor = SerialDispatchQueueExecutor(
            queue: DispatchQueue(label: label, qos: .background)
        )
    }

    func processHeavyTask(_ data: Data) async throws -> ProcessedResult {
        // This runs on the background queue
        let result = try JSONDecoder().decode(RawData.self, from: data)
        return ProcessedResult(from: result)
    }
}

final class SerialDispatchQueueExecutor: SerialExecutor {
    let queue: DispatchQueue

    init(queue: DispatchQueue) {
        self.queue = queue
    }

    func enqueue(_ job: consuming ExecutorJob) {
        let unownedJob = UnownedJob(job)
        queue.async {
            unownedJob.runSynchronously(on: self.asUnownedSerialExecutor())
        }
    }
}
```

### AsyncSequence Patterns

Custom AsyncSequence:
```swift
struct AsyncTimerSequence: AsyncSequence {
    typealias Element = Date

    let interval: TimeInterval

    struct AsyncIterator: AsyncIteratorProtocol {
        let interval: TimeInterval
        var lastTick: Date?

        mutating func next() async -> Date? {
            if Task.isCancelled { return nil }

            if let last = lastTick {
                let nextTick = last.addingTimeInterval(interval)
                let delay = nextTick.timeIntervalSinceNow
                if delay > 0 {
                    try? await Task.sleep(for: .seconds(delay))
                }
            }

            let now = Date()
            lastTick = now
            return now
        }
    }

    func makeAsyncIterator() -> AsyncIterator {
        AsyncIterator(interval: interval)
    }
}

// Usage
for await tick in AsyncTimerSequence(interval: 1.0) {
    print("Tick: \(tick)")
}
```

### Structured Concurrency with TaskGroups

Controlled Parallelism:
```swift
func processImages(_ urls: [URL], maxConcurrency: Int = 4) async throws -> [ProcessedImage] {
    try await withThrowingTaskGroup(of: ProcessedImage.self) { group in
        var results: [ProcessedImage] = []
        var urlIterator = urls.makeIterator()

        // Start initial batch
        for _ in 0..<min(maxConcurrency, urls.count) {
            if let url = urlIterator.next() {
                group.addTask { try await self.processImage(url) }
            }
        }

        // Process results and add new tasks
        for try await result in group {
            results.append(result)

            if let url = urlIterator.next() {
                group.addTask { try await self.processImage(url) }
            }
        }

        return results
    }
}
```

## Network Layer

### Modern URLSession with Async/Await

Type-Safe API Client:
```swift
actor APIClient {
    private let session: URLSession
    private let baseURL: URL
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(
        baseURL: URL,
        session: URLSession = .shared,
        decoder: JSONDecoder = JSONDecoder(),
        encoder: JSONEncoder = JSONEncoder()
    ) {
        self.baseURL = baseURL
        self.session = session
        self.decoder = decoder
        self.encoder = encoder

        decoder.keyDecodingStrategy = .convertFromSnakeCase
        encoder.keyEncodingStrategy = .convertToSnakeCase
    }

    func get<T: Decodable>(
        _ path: String,
        queryItems: [URLQueryItem] = []
    ) async throws(APIError) -> T {
        let request = try buildRequest(path: path, method: "GET", queryItems: queryItems)
        return try await execute(request)
    }

    func post<T: Decodable, B: Encodable>(
        _ path: String,
        body: B
    ) async throws(APIError) -> T {
        var request = try buildRequest(path: path, method: "POST")
        request.httpBody = try encoder.encode(body)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        return try await execute(request)
    }

    func put<T: Decodable, B: Encodable>(
        _ path: String,
        body: B
    ) async throws(APIError) -> T {
        var request = try buildRequest(path: path, method: "PUT")
        request.httpBody = try encoder.encode(body)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        return try await execute(request)
    }

    func delete(_ path: String) async throws(APIError) {
        let request = try buildRequest(path: path, method: "DELETE")
        let (_, response) = try await performRequest(request)
        try validateResponse(response)
    }

    private func buildRequest(
        path: String,
        method: String,
        queryItems: [URLQueryItem] = []
    ) throws(APIError) -> URLRequest {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: true)
        if !queryItems.isEmpty {
            components?.queryItems = queryItems
        }

        guard let url = components?.url else {
            throw .invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        return request
    }

    private func execute<T: Decodable>(_ request: URLRequest) async throws(APIError) -> T {
        let (data, response) = try await performRequest(request)
        try validateResponse(response)

        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw .decodingError(error)
        }
    }

    private func performRequest(_ request: URLRequest) async throws(APIError) -> (Data, URLResponse) {
        do {
            return try await session.data(for: request)
        } catch {
            throw .networkError(error)
        }
    }

    private func validateResponse(_ response: URLResponse) throws(APIError) {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw .invalidResponse
        }

        guard 200..<300 ~= httpResponse.statusCode else {
            throw .httpError(statusCode: httpResponse.statusCode)
        }
    }
}

enum APIError: Error, LocalizedError {
    case invalidURL
    case networkError(Error)
    case invalidResponse
    case httpError(statusCode: Int)
    case decodingError(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .invalidResponse:
            return "Invalid server response"
        case .httpError(let code):
            return "HTTP error: \(code)"
        case .decodingError(let error):
            return "Decoding error: \(error.localizedDescription)"
        }
    }
}
```

### Request Interceptor Pattern

Authentication Interceptor:
```swift
protocol RequestInterceptor: Sendable {
    func intercept(_ request: inout URLRequest) async throws
}

actor AuthInterceptor: RequestInterceptor {
    private let tokenProvider: TokenProviderProtocol

    init(tokenProvider: TokenProviderProtocol) {
        self.tokenProvider = tokenProvider
    }

    func intercept(_ request: inout URLRequest) async throws {
        let token = try await tokenProvider.getValidToken()
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
    }
}

actor LoggingInterceptor: RequestInterceptor {
    func intercept(_ request: inout URLRequest) async throws {
        print("[\(request.httpMethod ?? "?")] \(request.url?.absoluteString ?? "")")
    }
}
```

## SwiftData Integration

Modern Persistence (iOS 17+):
```swift
import SwiftData

@Model
final class Post {
    var id: UUID
    var title: String
    var content: String
    var createdAt: Date
    var author: Author?

    @Relationship(deleteRule: .cascade, inverse: \Comment.post)
    var comments: [Comment]

    init(title: String, content: String) {
        self.id = UUID()
        self.title = title
        self.content = content
        self.createdAt = Date()
        self.comments = []
    }
}

@Model
final class Comment {
    var id: UUID
    var text: String
    var createdAt: Date
    var post: Post?

    init(text: String) {
        self.id = UUID()
        self.text = text
        self.createdAt = Date()
    }
}

// Repository
@MainActor
final class PostRepository {
    private let modelContext: ModelContext

    init(modelContext: ModelContext) {
        self.modelContext = modelContext
    }

    func fetchPosts() throws -> [Post] {
        let descriptor = FetchDescriptor<Post>(
            sortBy: [SortDescriptor(\.createdAt, order: .reverse)]
        )
        return try modelContext.fetch(descriptor)
    }

    func save(_ post: Post) throws {
        modelContext.insert(post)
        try modelContext.save()
    }

    func delete(_ post: Post) throws {
        modelContext.delete(post)
        try modelContext.save()
    }
}
```

## Performance Optimization

### SwiftUI Performance Tips

Minimize View Updates:
- Use `@Observable` instead of `@ObservedObject` for better performance
- Implement `Equatable` on views when appropriate
- Use `LazyVStack`/`LazyHStack` for lists
- Apply `task(id:)` for cancellable async operations
- Use `@State` for view-local state, avoid unnecessary bindings

Memory Management:
- Use weak references in closures
- Implement proper cancellation in async tasks
- Profile with Instruments: Allocations, Leaks
- Use `nonisolated` where actor isolation is not needed

### Profiling with Instruments

Key Instruments:
- Time Profiler: CPU usage and hot paths
- Allocations: Memory allocation patterns
- Leaks: Memory leak detection
- SwiftUI: View body evaluations and updates
- Network: HTTP request analysis

---

Version: 1.0.0
Last Updated: 2025-12-07
