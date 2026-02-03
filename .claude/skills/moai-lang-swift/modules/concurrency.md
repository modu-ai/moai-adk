# Swift Concurrency

Swift Concurrency including async/await, actors, TaskGroup, and structured concurrency patterns.

## Async/Await Fundamentals

### Basic Async Functions

Simple Async Function:
```swift
func fetchUser(_ id: String) async throws -> User {
    let url = URL(string: "https://api.example.com/users/\(id)")!
    let (data, response) = try await URLSession.shared.data(from: url)
    
    guard let httpResponse = response as? HTTPURLResponse,
          200..<300 ~= httpResponse.statusCode else {
        throw NetworkError.requestFailed
    }
    return try JSONDecoder().decode(User.self, from: data)
}
```

### Async Sequences

Custom AsyncSequence:
```swift
struct CountdownSequence: AsyncSequence {
    typealias Element = Int
    let start: Int
    
    struct AsyncIterator: AsyncIteratorProtocol {
        var current: Int
        
        mutating func next() async -> Int? {
            guard current >= 0 else { return nil }
            defer { current -= 1 }
            try? await Task.sleep(for: .seconds(1))
            return current
        }
    }
    
    func makeAsyncIterator() -> AsyncIterator {
        AsyncIterator(current: start)
    }
}
```

## Actor Isolation

### Basic Actor Pattern

Thread-Safe State:
```swift
actor ImageCache {
    private var cache: [URL: UIImage] = [:]
    private var inProgress: [URL: Task<UIImage, Error>] = [:]
    
    func image(for url: URL) async throws -> UIImage {
        if let cached = cache[url] { return cached }
        if let task = inProgress[url] { return try await task.value }
        
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
        guard let image = UIImage(data: data) else { throw ImageError.invalidData }
        return image
    }
}
```

### MainActor

UI Updates:
```swift
@MainActor
@Observable
final class ContentViewModel {
    private(set) var items: [Item] = []
    private(set) var isLoading = false
    
    func loadItems() async {
        isLoading = true
        defer { isLoading = false }
        items = (try? await api.fetchItems()) ?? []
    }
}
```

## TaskGroup

### Parallel Execution

Basic TaskGroup:
```swift
func fetchAllUsers(_ ids: [String]) async throws -> [User] {
    try await withThrowingTaskGroup(of: User.self) { group in
        for id in ids { group.addTask { try await api.fetchUser(id) } }
        
        var users: [User] = []
        for try await user in group { users.append(user) }
        return users
    }
}
```

### Controlled Concurrency

Limiting Parallel Tasks:
```swift
func processImages(_ urls: [URL], maxConcurrency: Int = 4) async throws -> [ProcessedImage] {
    try await withThrowingTaskGroup(of: ProcessedImage.self) { group in
        var results: [ProcessedImage] = []
        var urlIterator = urls.makeIterator()
        
        for _ in 0..<min(maxConcurrency, urls.count) {
            if let url = urlIterator.next() {
                group.addTask { try await self.processImage(url) }
            }
        }
        
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

## Structured Concurrency

### Async Let

Parallel Independent Operations:
```swift
func loadDashboard() async throws -> Dashboard {
    async let user = api.fetchCurrentUser()
    async let posts = api.fetchRecentPosts()
    async let notifications = api.fetchNotifications()
    
    return try await Dashboard(user: user, posts: posts, notifications: notifications)
}
```

### Task Cancellation

Checking Cancellation:
```swift
func processLargeDataset(_ items: [DataItem]) async throws -> [ProcessedItem] {
    var results: [ProcessedItem] = []
    
    for item in items {
        try Task.checkCancellation()
        let processed = try await process(item)
        results.append(processed)
    }
    return results
}
```

### Task Priority

Setting Priority:
```swift
func loadContent() async {
    let imageTask = Task(priority: .userInitiated) { try await loadVisibleImages() }
    let prefetchTask = Task(priority: .background) { try await prefetchNextPage() }
    
    _ = try? await imageTask.value
    _ = try? await prefetchTask.value
}
```

## AsyncStream

### Creating AsyncStream

Event-Based:
```swift
actor EventBus {
    private var continuations: [UUID: AsyncStream<AppEvent>.Continuation] = [:]
    
    func subscribe() -> AsyncStream<AppEvent> {
        let id = UUID()
        return AsyncStream { continuation in
            continuations[id] = continuation
            continuation.onTermination = { [weak self] _ in
                Task { await self?.removeContinuation(id) }
            }
        }
    }
    
    private func removeContinuation(_ id: UUID) { continuations.removeValue(forKey: id) }
    
    func emit(_ event: AppEvent) {
        for continuation in continuations.values { continuation.yield(event) }
    }
}
```

### Timer AsyncStream

```swift
func makeTimer(interval: Duration) -> AsyncStream<Date> {
    AsyncStream { continuation in
        let task = Task {
            while !Task.isCancelled {
                continuation.yield(Date())
                try? await Task.sleep(for: interval)
            }
            continuation.finish()
        }
        continuation.onTermination = { _ in task.cancel() }
    }
}
```

## Continuation Bridges

### Callback to Async

Converting Completion Handlers:
```swift
func fetchDataAsync() async throws -> Data {
    try await withCheckedThrowingContinuation { continuation in
        fetchData { result in
            switch result {
            case .success(let data): continuation.resume(returning: data)
            case .failure(let error): continuation.resume(throwing: error)
            }
        }
    }
}
```

### Delegate to Async

```swift
actor LocationProvider {
    func requestLocation() async throws -> CLLocation {
        try await withCheckedThrowingContinuation { continuation in
            let delegate = SingleLocationDelegate(continuation: continuation)
            let manager = CLLocationManager()
            manager.delegate = delegate
            manager.requestLocation()
        }
    }
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
Context7: /apple/swift, /apple/swift-async-algorithms
