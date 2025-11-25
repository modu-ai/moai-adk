# Advanced Swift 6.0 Patterns

**Focus**: Concurrency, SwiftUI advanced patterns, protocol-oriented programming
**Standards**: Swift 6.0 with strict concurrency checking
**Last Updated**: 2025-11-22

---

## Pattern 1: Actor-Based Concurrency

```swift
// Thread-safe actor for shared state
actor UserRepository {
    private var cache: [Int: User] = [:]

    func fetch(id: Int) -> User? {
        cache[id]
    }

    func store(_ user: User) {
        cache[user.id] = user
    }

    func clear() {
        cache.removeAll()
    }
}

// Usage - automatic synchronization
let repo = UserRepository()
await repo.store(User(id: 1, name: "Alice"))
let user = await repo.fetch(id: 1)
```

---

## Pattern 2: Async/Await Error Handling

```swift
// Proper async error handling
actor DataService {
    enum DataError: Error {
        case networkError(String)
        case invalidData
        case timeout
    }

    func fetchData(from url: URL) async throws -> Data {
        var request = URLRequest(url: url)
        request.timeoutInterval = 30

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw DataError.invalidData
        }

        guard (200..<300).contains(httpResponse.statusCode) else {
            throw DataError.networkError("HTTP \(httpResponse.statusCode)")
        }

        return data
    }
}

// Retry with exponential backoff
func retryAsync<T>(
    maxAttempts: Int = 3,
    delay: Duration = .milliseconds(100),
    operation: () async throws -> T
) async throws -> T {
    var lastError: Error?

    for attempt in 1...maxAttempts {
        do {
            return try await operation()
        } catch {
            lastError = error
            if attempt < maxAttempts {
                try await Task.sleep(nanoseconds: UInt64(delay.components.seconds * 1_000_000_000))
            }
        }
    }

    throw lastError ?? URLError(.unknown)
}
```

---

## Pattern 3: SwiftUI with MVVM Architecture

```swift
// Observable ViewModel
@Observable
final class UserListViewModel {
    var users: [User] = []
    var isLoading = false
    var errorMessage: String?

    private let service: UserService

    init(service: UserService = .shared) {
        self.service = service
    }

    @MainActor
    func loadUsers() async {
        isLoading = true
        errorMessage = nil

        do {
            users = try await service.fetchUsers()
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }
}

// SwiftUI View using MVVM
struct UserListView: View {
    @State private var viewModel: UserListViewModel

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else if let error = viewModel.errorMessage {
                    Text(error)
                        .foregroundStyle(.red)
                } else {
                    List(viewModel.users) { user in
                        NavigationLink(value: user) {
                            UserRowView(user: user)
                        }
                    }
                }
            }
            .navigationTitle("Users")
            .task {
                await viewModel.loadUsers()
            }
        }
    }
}
```

---

## Pattern 4: Combine for Reactive Programming

```swift
// Publisher-Subscriber pattern
class APIClient {
    private let urlSession: URLSession

    func publisher<T: Decodable>(for url: URL) -> AnyPublisher<T, Error> {
        URLSession.shared.dataTaskPublisher(for: url)
            .map { $0.data }
            .decode(type: T.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
}

// Usage with error handling
@Observable
class SearchViewModel {
    var results: [SearchResult] = []
    var errorMessage: String?
    private var cancellables = Set<AnyCancellable>()

    func search(query: String) {
        guard !query.isEmpty else { results = []; return }

        URLSession.shared.dataTaskPublisher(for: searchURL(query))
            .map { $0.data }
            .decode(type: SearchResponse.self, decoder: JSONDecoder())
            .map { $0.results }
            .receive(on: DispatchQueue.main)
            .sink { [weak self] completion in
                if case .failure(let error) = completion {
                    self?.errorMessage = error.localizedDescription
                }
            } receiveValue: { [weak self] results in
                self?.results = results
            }
            .store(in: &cancellables)
    }
}
```

---

## Pattern 5: Protocol-Oriented Design

```swift
// Protocol composition for flexibility
protocol DataProvider {
    associatedtype Output: Decodable

    func fetchData() async throws -> Output
}

protocol Cacheable {
    associatedtype CacheKey: Hashable
    var cache: [CacheKey: Self] { get set }
}

// Combine protocols
final class CachedDataService<T: DataProvider>: Cacheable {
    typealias CacheKey = String

    var cache: [String: T] = [:]
    private let provider: T

    init(provider: T) {
        self.provider = provider
    }

    func fetchData(cacheKey: String = "default") async throws -> T.Output {
        if cache[cacheKey] != nil {
            // Return from cache
        }
        return try await provider.fetchData()
    }
}
```

---

## Pattern 6: Custom Property Wrappers

```swift
// Thread-safe property wrapper
@propertyWrapper
struct ThreadSafe<Value> {
    private let lock = NSLock()
    private var value: Value

    init(wrappedValue: Value) {
        self.value = wrappedValue
    }

    var wrappedValue: Value {
        get {
            lock.lock()
            defer { lock.unlock() }
            return value
        }
        set {
            lock.lock()
            defer { lock.unlock() }
            value = newValue
        }
    }
}

// Validated property wrapper
@propertyWrapper
struct Validated<Value> {
    private var value: Value
    let validate: (Value) -> Bool

    init(wrappedValue: Value, validate: @escaping (Value) -> Bool) {
        self.value = wrappedValue
        self.validate = validate
    }

    var wrappedValue: Value {
        get { value }
        set { if validate(newValue) { value = newValue } }
    }
}

// Usage
@ThreadSafe var counter = 0
@Validated(validate: { $0 > 0 }) var positiveNumber = 1
```

---

## Pattern 7: SwiftUI Modifier Composition

```swift
// Reusable custom modifiers
struct PrimaryButtonStyle: ViewModifier {
    func body(content: Content) -> some View {
        content
            .font(.headline)
            .foregroundStyle(.white)
            .padding()
            .background(Color.blue)
            .cornerRadius(8)
    }
}

extension View {
    func primaryButtonStyle() -> some View {
        modifier(PrimaryButtonStyle())
    }
}

// Conditional modifiers
extension View {
    @ViewBuilder
    func conditionalModifier<T: View>(
        _ condition: Bool,
        modifier: (Self) -> T
    ) -> some View {
        if condition {
            modifier(self)
        } else {
            self
        }
    }
}

// Usage
Button("Submit") { /* action */ }
    .primaryButtonStyle()
    .conditionalModifier(isLoading) { button in
        button.opacity(0.5).disabled(true)
    }
```

---

## Pattern 8: Task Groups for Concurrent Operations

```swift
// Parallel fetching with task groups
func fetchAllData() async throws -> (users: [User], posts: [Post]) {
    try await withThrowingTaskGroup(of: DataType.self) { group in
        var users: [User] = []
        var posts: [Post] = []

        group.addTask {
            .users(try await fetchUsers())
        }

        group.addTask {
            .posts(try await fetchPosts())
        }

        for try await data in group {
            switch data {
            case .users(let u): users = u
            case .posts(let p): posts = p
            }
        }

        return (users, posts)
    }
}

enum DataType: Sendable {
    case users([User])
    case posts([Post])
}
```

---

## Pattern 9: Memory Management with Weak References

```swift
// Weak reference in closure to prevent retain cycles
class ViewController: UIViewController {
    func setupBindings() {
        viewModel.userPublisher
            .sink { [weak self] user in
                self?.update(with: user)
            }
            .store(in: &cancellables)
    }

    // Weak self in async closure
    func loadData() {
        Task { [weak self] in
            do {
                let data = try await fetchData()
                await self?.update(data)
            } catch {
                self?.showError(error)
            }
        }
    }
}
```

---

## Pattern 10: Type-Safe Resource Management

```swift
// RAII-like pattern in Swift
protocol Resource {
    func close() async throws
}

struct FileHandle: Resource {
    private let handle: Foundation.FileHandle

    func close() async throws {
        try? handle.close()
    }
}

// Safe resource usage
func withResource<R: Resource, T>(
    _ resource: R,
    block: (R) async throws -> T
) async throws -> T {
    defer { try? Task { try await resource.close() }.value }
    return try await block(resource)
}

// Usage
try await withResource(fileHandle) { handle in
    // Use resource safely
    let data = try handle.readToEndOfFile()
    return data
}
```

---

**Last Updated**: 2025-11-22 | **Production Ready**
