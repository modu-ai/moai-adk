# Combine Framework

Publishers, Subscribers, Operators, and integration with async/await.

## Publishers

### Built-in Publishers

Just and Future:
```swift
import Combine

let justPublisher = Just("Hello, Combine!")

let futurePublisher = Future<User, Error> { promise in
    Task {
        do {
            let user = try await api.fetchUser("123")
            promise(.success(user))
        } catch {
            promise(.failure(error))
        }
    }
}
```

### PassthroughSubject

Event Broadcasting:
```swift
class EventBus {
    private let eventSubject = PassthroughSubject<AppEvent, Never>()
    
    var eventPublisher: AnyPublisher<AppEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }
    
    func emit(_ event: AppEvent) { eventSubject.send(event) }
}
```

### CurrentValueSubject

State Container:
```swift
class ThemeManager {
    private let themeSubject = CurrentValueSubject<Theme, Never>(.system)
    
    var themePublisher: AnyPublisher<Theme, Never> { themeSubject.eraseToAnyPublisher() }
    var currentTheme: Theme { themeSubject.value }
    
    func setTheme(_ theme: Theme) { themeSubject.send(theme) }
}
```

## Operators

### Transformation

Map and FlatMap:
```swift
userPublisher
    .map { userId in URLRequest(url: URL(string: "https://api.example.com/users/\(userId)")!) }
    .flatMap { request -> AnyPublisher<User, Error> in
        URLSession.shared.dataTaskPublisher(for: request)
            .map(\.data)
            .decode(type: User.self, decoder: JSONDecoder())
            .eraseToAnyPublisher()
    }
    .sink(receiveCompletion: { _ in }, receiveValue: { user in print(user.name) })
```

### Filtering

Debounce and RemoveDuplicates:
```swift
class SearchViewModel: ObservableObject {
    @Published var searchText = ""
    @Published private(set) var results: [SearchResult] = []
    private var cancellables = Set<AnyCancellable>()
    
    init(searchService: SearchServiceProtocol) {
        $searchText
            .debounce(for: .milliseconds(300), scheduler: DispatchQueue.main)
            .removeDuplicates()
            .filter { $0.count >= 2 }
            .flatMap { query in searchService.search(query).catch { _ in Just([]) } }
            .receive(on: DispatchQueue.main)
            .assign(to: &$results)
    }
}
```

### Combining

CombineLatest:
```swift
class FormViewModel: ObservableObject {
    @Published var email = ""
    @Published var password = ""
    @Published private(set) var isValid = false
    
    init() {
        Publishers.CombineLatest($email, $password)
            .map { email, password in
                email.contains("@") && password.count >= 8
            }
            .assign(to: &$isValid)
    }
}
```

### Error Handling

Catch and Retry:
```swift
func fetchWithRetry<T: Decodable>(_ url: URL) -> AnyPublisher<T, Error> {
    URLSession.shared.dataTaskPublisher(for: url)
        .map(\.data)
        .decode(type: T.self, decoder: JSONDecoder())
        .retry(3)
        .eraseToAnyPublisher()
}
```

## Subscribers

### Sink

Basic Sink:
```swift
let subscription = [1, 2, 3].publisher
    .sink(
        receiveCompletion: { print($0) },
        receiveValue: { print($0) }
    )
```

### Assign

Property Assignment:
```swift
class ViewModel: ObservableObject {
    @Published var count = 0
    @Published var displayText = ""
    
    init() {
        $count.map { "Count: \($0)" }.assign(to: &$displayText)
    }
}
```

## Async/Await Bridge

### Publisher to Async

Converting Publishers:
```swift
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
                    receiveValue: { continuation.resume(returning: $0) }
                )
        }
    }
}
```

### Async to Publisher

Converting Async:
```swift
func asyncToPublisher<T>(_ operation: @escaping () async throws -> T) -> AnyPublisher<T, Error> {
    Future { promise in
        Task {
            do { promise(.success(try await operation())) }
            catch { promise(.failure(error)) }
        }
    }.eraseToAnyPublisher()
}
```

## SwiftUI Integration

### @Published with Combine

ViewModel Pattern:
```swift
class ContentViewModel: ObservableObject {
    @Published var items: [Item] = []
    @Published var searchQuery = ""
    @Published private(set) var filteredItems: [Item] = []
    private var cancellables = Set<AnyCancellable>()
    
    init() {
        Publishers.CombineLatest($items, $searchQuery)
            .map { items, query in
                guard !query.isEmpty else { return items }
                return items.filter { $0.name.localizedCaseInsensitiveContains(query) }
            }
            .assign(to: &$filteredItems)
    }
}
```

### onReceive Modifier

Responding to Publishers:
```swift
struct TimerView: View {
    @State private var currentTime = Date()
    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()
    
    var body: some View {
        Text(currentTime.formatted(date: .omitted, time: .standard))
            .onReceive(timer) { currentTime = $0 }
    }
}
```

### Notification Center

System Notifications:
```swift
struct KeyboardAdaptiveView: View {
    @State private var keyboardHeight: CGFloat = 0
    
    var body: some View {
        content
            .padding(.bottom, keyboardHeight)
            .onReceive(NotificationCenter.default.publisher(for: UIResponder.keyboardWillShowNotification)) { notification in
                if let frame = notification.userInfo?[UIResponder.keyboardFrameEndUserInfoKey] as? CGRect {
                    withAnimation { keyboardHeight = frame.height }
                }
            }
            .onReceive(NotificationCenter.default.publisher(for: UIResponder.keyboardWillHideNotification)) { _ in
                withAnimation { keyboardHeight = 0 }
            }
    }
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
Context7: /CombineCommunity/CombineExt
