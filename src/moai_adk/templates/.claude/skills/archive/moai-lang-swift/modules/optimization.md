# Swift 6.0 Performance Optimization

**Focus**: Runtime performance, memory efficiency, compilation optimization
**Standards**: Swift 6.0 with Xcode 16+
**Last Updated**: 2025-11-22

---

## Optimization Level 1: View Hierarchy Optimization

```swift
// Bad: Large view hierarchy causes re-renders
struct UserListView: View {
    @State private var users: [User] = []

    var body: some View {
        VStack {
            ForEach(users) { user in
                HStack {
                    Image(systemName: "person.fill")
                    VStack(alignment: .leading) {
                        Text(user.name)
                        Text(user.email)
                    }
                    Spacer()
                }
            }
        }
    }
}

// Good: Extract row component to optimize re-rendering
struct UserListView: View {
    @State private var users: [User] = []

    var body: some View {
        List(users) { user in
            UserRowView(user: user)  // Extracted component
        }
    }
}

struct UserRowView: View {
    let user: User
    var body: some View {
        HStack {
            Image(systemName: "person.fill")
            VStack(alignment: .leading) {
                Text(user.name)
                Text(user.email)
            }
            Spacer()
        }
    }
}
```

---

## Optimization Level 2: Memory Management

```swift
// Bad: Strong reference cycles
class APIClient {
    var onDataReceived: ((Data) -> Void)?

    func fetch() {
        Task {
            // Captures self strongly - potential cycle
            let data = try await fetchData()
            onDataReceived?(data)
        }
    }
}

// Good: Break reference cycles with weak self
class APIClient {
    var onDataReceived: ((Data) -> Void)?

    func fetch() {
        Task { [weak self] in
            guard let self else { return }
            let data = try await fetchData()
            self.onDataReceived?(data)
        }
    }
}

// Better: Use @escaping closure with unowned self
final class APIClient {
    var onDataReceived: ((Data) -> Void)?

    func fetch(completion: @escaping (Data) -> Void) {
        Task { [unowned self] in  // Safe with final class
            let data = try await fetchData()
            completion(data)
        }
    }
}
```

---

## Optimization Level 3: Lazy Loading and Pagination

```swift
// Lazy-load image data
@Observable
class ImageViewModel {
    @ObservationIgnored
    private var imageCache: [String: UIImage] = [:]

    func loadImage(from url: URL) async -> UIImage? {
        // Check cache first
        if let cached = imageCache[url.absoluteString] {
            return cached
        }

        // Load from network
        let image = try? await URLSession.shared.image(from: url)
        if let image {
            imageCache[url.absoluteString] = image
        }
        return image
    }
}

// Paginated list loading
@Observable
class PaginatedListViewModel {
    private let pageSize = 20
    private var currentPage = 0
    private var hasMore = true

    var items: [Item] = []

    func loadMore() async {
        guard hasMore else { return }

        do {
            let newItems = try await fetchItems(page: currentPage, size: pageSize)
            items.append(contentsOf: newItems)

            hasMore = newItems.count == pageSize
            currentPage += 1
        } catch {
            // Handle error
        }
    }
}
```

---

## Optimization Level 4: Value Types Over Reference Types

```swift
// Good: Use struct (value type) instead of class
struct User: Hashable {
    let id: Int
    let name: String
    let email: String
}

// Avoid unnecessary copies with @inlinable
@inlinable
func process(user: User) -> String {
    user.name
}

// Copy-on-write optimization
struct CoWCollection {
    private var storage: [Int]

    init(_ values: [Int] = []) {
        self.storage = values
    }

    mutating func append(_ value: Int) {
        // Only copies when needed
        storage.append(value)
    }
}
```

---

## Optimization Level 5: Async/Await Efficiency

```swift
// Bad: Sequential async calls
async func fetchUserAndPosts(id: Int) async throws -> (User, [Post]) {
    let user = try await fetchUser(id: id)
    let posts = try await fetchPosts(userId: id)
    return (user, posts)
}

// Good: Concurrent async calls
async func fetchUserAndPosts(id: Int) async throws -> (User, [Post]) {
    async let user = fetchUser(id: id)
    async let posts = fetchPosts(userId: id)
    return (try await user, try await posts)
}

// Efficient task grouping
func fetchMultipleUsers(ids: [Int]) async throws -> [User] {
    return try await withThrowingTaskGroup(of: User.self) { group in
        for id in ids {
            group.addTask {
                try await fetchUser(id: id)
            }
        }

        var results: [User] = []
        for try await user in group {
            results.append(user)
        }
        return results
    }
}
```

---

## Optimization Level 6: String and Collection Performance

```swift
// Bad: String concatenation in loops (O(n²))
var result = ""
for i in 0..<1000 {
    result += String(i)  // Creates new string each iteration
}

// Good: Use StringBuilder pattern
var result = ""
result.reserveCapacity(3000)  // Pre-allocate
for i in 0..<1000 {
    result.append(contentsOf: String(i))
}

// Best: Use collection directly
let result = (0..<1000).map(String.init).joined()

// Collection allocation
var items: [User] = []
items.reserveCapacity(expectedCount)  // Pre-allocate
for user in fetchedUsers {
    items.append(user)
}
```

---

## Optimization Level 7: Compiler Optimization Flags

```bash
# Build for release with optimizations
swift build -c release

# Enable whole module optimization (slower build, faster runtime)
swiftc -whole-module-optimization -O myfile.swift

# Enable optimization and enable assertions
swift build -c release -Xswiftc -Osize

# Profile-guided optimization
swiftc -profile-generate -profile-use=profile.profraw myfile.swift
```

---

## Optimization Level 8: Generics and Type Erasure

```swift
// Avoid unnecessary type erasure - prefer generics
struct GenericContainer<T> {
    let value: T
}

let container = GenericContainer(value: 42)

// When needed, use type erasure with specific erased type
struct AnyPublisher<Output, Failure: Error> {
    // Implementation
}

// Specialized generics for performance
@inlinable
func process<T: Numeric>(_ values: [T]) -> T {
    values.reduce(0, +)
}
```

---

## Optimization Level 9: Profiling with Instruments

```swift
// Add time tracking to performance-critical code
import os.log

let logger = Logger(subsystem: "com.myapp", category: "performance")

func expensiveOperation() {
    let startTime = Date()
    defer {
        let elapsed = Date().timeIntervalSince(startTime)
        logger.debug("Operation took \(elapsed)s")
    }

    // Do work
}

// Use os_signpost for Instruments visibility
import os

let pointsOfInterest = OSLog(subsystem: "com.myapp", category: .pointsOfInterest)

func trackOperation<T>(_ name: StaticString, block: () throws -> T) rethrows -> T {
    let signpostID = OSSignpostID(log: pointsOfInterest)
    os_signpost(.begin, log: pointsOfInterest, name: name, signpostID: signpostID)
    defer {
        os_signpost(.end, log: pointsOfInterest, name: name, signpostID: signpostID)
    }
    return try block()
}
```

---

## Optimization Level 10: Sendable Protocol for Concurrency

```swift
// Mark types as Sendable for safe concurrency
@Observable
final class UserViewModel: @unchecked Sendable {
    nonisolated let id: Int
    var name: String = ""

    init(id: Int) {
        self.id = id
    }
}

// Use Sendable for actors
actor UserRepository: Sendable {
    private var users: [Int: User] = [:]

    func fetch(id: Int) -> User? {
        users[id]
    }

    func store(_ user: User) {
        users[user.id] = user
    }
}
```

---

## Build Optimization Settings

```swift
// In Package.swift
let package = Package(
    name: "MyPackage",
    products: [
        .library(name: "MyLibrary", targets: ["MyLibrary"])
    ],
    targets: [
        .target(
            name: "MyLibrary",
            swiftSettings: [
                .unsafeFlags(["-O", "-whole-module-optimization"]),
                .define("NDEBUG")
            ]
        )
    ]
)
```

---

**Last Updated**: 2025-11-22 | **Production Ready**
