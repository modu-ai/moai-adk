# moai-lang-swift - Working Examples

_Last updated: 2025-11-12_

## Quick Start (Swift 6.0 Minimal Setup)

```swift
import Foundation

// Modern Swift 6.0 with async/await
actor HelloWorld {
    func greet(_ name: String) -> String {
        "Hello, \(name)!"
    }
}

// Usage
let greeter = HelloWorld()
let greeting = await greeter.greet("Swift")
print(greeting)  // Output: Hello, Swift!
```

## Basic Usage Examples

### Example 1: Async/Await Function - Basic

```swift
import Foundation

// Simple async function that fetches data
func fetchUserData(userId: String) async throws -> String {
    let urlString = "https://jsonplaceholder.typicode.com/users/1"
    guard let url = URL(string: urlString) else {
        throw URLError(.badURL)
    }

    do {
        let (data, _) = try await URLSession.shared.data(from: url)
        let jsonString = String(data: data, encoding: .utf8) ?? "Unable to decode"
        return jsonString
    } catch {
        throw error
    }
}

// Using the async function
Task {
    do {
        let userData = try await fetchUserData(userId: "1")
        print("User data fetched: \(userData)")
    } catch {
        print("Error: \(error.localizedDescription)")
    }
}
```

### Example 2: SwiftUI Basics - Counter App

```swift
import SwiftUI

struct ContentView: View {
    @State private var count = 0

    var body: some View {
        VStack(spacing: 20) {
            Text("Counter: \(count)")
                .font(.title)
                .fontWeight(.bold)

            HStack(spacing: 10) {
                Button(action: { count -= 1 }) {
                    Text("Decrease")
                        .padding()
                        .background(Color.red)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                }

                Button(action: { count += 1 }) {
                    Text("Increase")
                        .padding()
                        .background(Color.green)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                }

                Button(action: { count = 0 }) {
                    Text("Reset")
                        .padding()
                        .background(Color.gray)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                }
            }
        }
        .padding()
    }
}

#Preview {
    ContentView()
}
```

### Example 3: Combine Publisher - Event Stream

```swift
import Combine
import Foundation

// Create a simple publisher
let numbers = [1, 2, 3, 4, 5]
let subject = PassthroughSubject<Int, Never>()
var cancellables = Set<AnyCancellable>()

// Define pipeline
subject
    .filter { $0 % 2 == 0 }  // Keep only even numbers
    .map { $0 * 2 }           // Double each number
    .sink { value in
        print("Processed: \(value)")  // Output: 4, 8
    }
    .store(in: &cancellables)

// Emit values
for number in numbers {
    subject.send(number)
}
```

## Intermediate Patterns

### Example 4: Complex Async/Await with Error Handling

```swift
import Foundation

enum NetworkError: Error {
    case invalidURL
    case httpError(Int)
    case decodingError(String)
    case timeout
}

struct User: Decodable {
    let id: Int
    let name: String
    let email: String
}

// Advanced async function with error handling and timeouts
func fetchUserProfile(userId: Int, timeout: TimeInterval = 10) async throws -> User {
    guard let url = URL(string: "https://jsonplaceholder.typicode.com/users/\(userId)") else {
        throw NetworkError.invalidURL
    }

    do {
        // Set timeout
        var request = URLRequest(url: url)
        request.timeoutInterval = timeout

        let (data, response) = try await URLSession.shared.data(for: request)

        // Validate HTTP response
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.httpError(0)
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            throw NetworkError.httpError(httpResponse.statusCode)
        }

        // Decode JSON
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let user = try decoder.decode(User.self, from: data)

        return user
    } catch let error as DecodingError {
        throw NetworkError.decodingError(error.localizedDescription)
    } catch {
        throw error
    }
}

// Usage with try-catch
Task {
    do {
        let user = try await fetchUserProfile(userId: 1)
        print("User: \(user.name) (\(user.email))")
    } catch NetworkError.invalidURL {
        print("Invalid URL provided")
    } catch NetworkError.httpError(let statusCode) {
        print("HTTP Error: \(statusCode)")
    } catch NetworkError.decodingError(let message) {
        print("Decoding Error: \(message)")
    } catch {
        print("Unknown error: \(error)")
    }
}
```

### Example 5: SwiftUI State Management - Todo List

```swift
import SwiftUI

struct Todo: Identifiable {
    let id = UUID()
    var title: String
    var isCompleted: Bool = false
}

@MainActor
class TodoViewModel: ObservableObject {
    @Published var todos: [Todo] = []

    func addTodo(title: String) {
        let newTodo = Todo(title: title)
        todos.append(newTodo)
    }

    func toggleTodo(_ todo: Todo) {
        if let index = todos.firstIndex(where: { $0.id == todo.id }) {
            todos[index].isCompleted.toggle()
        }
    }

    func deleteTodo(_ todo: Todo) {
        todos.removeAll { $0.id == todo.id }
    }
}

struct TodoListView: View {
    @StateObject private var viewModel = TodoViewModel()
    @State private var newTodoTitle = ""

    var body: some View {
        NavigationView {
            VStack {
                // Input field for new todo
                HStack {
                    TextField("New todo...", text: $newTodoTitle)
                        .textFieldStyle(.roundedBorder)

                    Button(action: {
                        if !newTodoTitle.isEmpty {
                            viewModel.addTodo(title: newTodoTitle)
                            newTodoTitle = ""
                        }
                    }) {
                        Text("Add")
                    }
                }
                .padding()

                // Todo list
                List {
                    ForEach(viewModel.todos) { todo in
                        HStack {
                            Image(systemName: todo.isCompleted ? "checkmark.circle.fill" : "circle")
                                .foregroundColor(todo.isCompleted ? .green : .gray)

                            Text(todo.title)
                                .strikethrough(todo.isCompleted)

                            Spacer()

                            Button(role: .destructive, action: {
                                viewModel.deleteTodo(todo)
                            }) {
                                Image(systemName: "trash")
                            }
                        }
                        .contentShape(Rectangle())
                        .onTapGesture {
                            viewModel.toggleTodo(todo)
                        }
                    }
                }
            }
            .navigationTitle("My Todos")
        }
    }
}

#Preview {
    TodoListView()
}
```

### Example 6: Combine Reactive Filtering

```swift
import Combine

// Search filter with Combine
class SearchViewModel: ObservableObject {
    @Published var searchText = ""
    @Published var results: [String] = []

    private var cancellables = Set<AnyCancellable>()

    private let allItems = ["Apple", "Banana", "Cherry", "Date", "Elderberry"]

    init() {
        $searchText
            .debounce(for: .milliseconds(500), scheduler: RunLoop.main)
            .removeDuplicates()
            .sink { [weak self] query in
                self?.updateResults(for: query)
            }
            .store(in: &cancellables)
    }

    private func updateResults(for query: String) {
        if query.isEmpty {
            results = allItems
        } else {
            results = allItems.filter {
                $0.localizedCaseInsensitiveContains(query)
            }
        }
    }
}

// Usage in SwiftUI
struct SearchView: View {
    @StateObject private var viewModel = SearchViewModel()

    var body: some View {
        VStack {
            TextField("Search...", text: $viewModel.searchText)
                .textFieldStyle(.roundedBorder)
                .padding()

            List(viewModel.results, id: \.self) { item in
                Text(item)
            }
        }
    }
}
```

## Advanced Patterns

### Example 7: Concurrent Task Execution with TaskGroup

```swift
import Foundation

// Fetch multiple resources concurrently
actor DataAggregator {
    func fetchMultipleResources() async throws -> (users: [String], posts: [String], comments: [String]) {
        return try await withThrowingTaskGroup(of: (String, [String]).self) { group in
            // Add tasks to fetch different resources
            group.addTask { ("users", try await self.fetchUsers()) }
            group.addTask { ("posts", try await self.fetchPosts()) }
            group.addTask { ("comments", try await self.fetchComments()) }

            // Collect results
            var users: [String] = []
            var posts: [String] = []
            var comments: [String] = []

            for try await (type, data) in group {
                switch type {
                case "users": users = data
                case "posts": posts = data
                case "comments": comments = data
                default: break
                }
            }

            return (users, posts, comments)
        }
    }

    private func fetchUsers() async throws -> [String] {
        try await Task.sleep(nanoseconds: 1_000_000_000)  // Simulate 1 second delay
        return ["User 1", "User 2", "User 3"]
    }

    private func fetchPosts() async throws -> [String] {
        try await Task.sleep(nanoseconds: 2_000_000_000)  // Simulate 2 second delay
        return ["Post 1", "Post 2"]
    }

    private func fetchComments() async throws -> [String] {
        try await Task.sleep(nanoseconds: 1_500_000_000)  // Simulate 1.5 second delay
        return ["Comment 1", "Comment 2", "Comment 3", "Comment 4"]
    }
}

// Usage
Task {
    let aggregator = DataAggregator()
    do {
        let result = try await aggregator.fetchMultipleResources()
        print("Users: \(result.users)")
        print("Posts: \(result.posts)")
        print("Comments: \(result.comments)")
    } catch {
        print("Error aggregating data: \(error)")
    }
}
```

### Example 8: Actor Isolation - Thread-Safe Counter

```swift
// Thread-safe counter using Actor
actor CounterService {
    private var count: Int = 0

    func increment() {
        count += 1
    }

    func decrement() {
        count -= 1
    }

    func getCount() -> Int {
        count
    }

    func reset() {
        count = 0
    }
}

// Usage with concurrent access
Task {
    let counter = CounterService()

    // Concurrent increments (thread-safe)
    await withTaskGroup(of: Void.self) { group in
        for _ in 0..<100 {
            group.addTask {
                await counter.increment()
            }
        }
    }

    let finalCount = await counter.getCount()
    print("Final count: \(finalCount)")  // Output: 100
}
```

### Example 9: Vapor Web Server - RESTful API

```swift
import Vapor

func routes(_ app: Application) throws {
    // Health check endpoint
    app.get("api", "health") { req async in
        return ["status": "healthy", "timestamp": Date().ISO8601Format()]
    }

    // List all users
    app.get("api", "users") { req async -> [String: String] in
        return ["users": "List of users would be returned here"]
    }

    // Get single user by ID
    app.get("api", "users", ":id") { req async -> [String: String] in
        guard let id = req.parameters.get("id") else {
            throw Abort(.badRequest)
        }
        return ["user_id": id, "name": "User \(id)"]
    }

    // Create new user
    app.post("api", "users") { req async -> [String: String] in
        // In a real app, decode request body
        let id = UUID().uuidString
        return ["user_id": id, "created": "true"]
    }

    // Update user
    app.put("api", "users", ":id") { req async -> [String: String] in
        guard let id = req.parameters.get("id") else {
            throw Abort(.badRequest)
        }
        return ["user_id": id, "updated": "true"]
    }

    // Delete user
    app.delete("api", "users", ":id") { req async -> HTTPStatus in
        return .noContent
    }
}
```

## Context7 Integration Examples

### Example 10: Getting Latest Swift Documentation

Use Context7 to access real-time Swift documentation:

```python
# Python example for accessing Context7
from context7_client import Context7

# Resolve library ID for Swift
library_id = Context7.resolve_library_id("swift")
print(f"Swift library ID: {library_id}")

# Get latest documentation
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="structured-concurrency",
    tokens=5000
)
print(f"Latest Swift docs:\n{docs}")
```

### Example 11: Getting SwiftUI Documentation

```python
# Get SwiftUI specific documentation
from context7_client import Context7

library_id = Context7.resolve_library_id("swiftui")
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="state-management",
    tokens=3000
)
print(f"SwiftUI patterns:\n{docs}")
```

### Example 12: Getting Vapor Framework Documentation

```python
# Get Vapor framework documentation
from context7_client import Context7

library_id = Context7.resolve_library_id("vapor")
docs = Context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="routing",
    tokens=4000
)
print(f"Vapor routing:\n{docs}")
```

## Migration Guides

### From Swift 5.x to Swift 6.0

```swift
// OLD: Combine-based async
func fetchDataOld() -> AnyPublisher<String, Error> {
    URLSession.shared.dataTaskPublisher(for: URL(string: "https://...")!)
        .map { String(data: $0.data, encoding: .utf8) ?? "" }
        .eraseToAnyPublisher()
}

// NEW: Swift 6.0 async/await
func fetchDataNew() async throws -> String {
    let (data, _) = try await URLSession.shared.data(from: URL(string: "https://...")!)
    return String(data: data, encoding: .utf8) ?? ""
}
```

### From Callbacks to Async/Await

```swift
// OLD: Callback-based
func fetchUserOld(id: Int, completion: @escaping (String) -> Void) {
    URLSession.shared.dataTask(with: URL(string: "https://...")!) { data, _, _ in
        if let data = data {
            let userName = String(data: data, encoding: .utf8) ?? "Unknown"
            completion(userName)
        }
    }.resume()
}

// NEW: Async/await
func fetchUserNew(id: Int) async throws -> String {
    let (data, _) = try await URLSession.shared.data(from: URL(string: "https://...")!)
    return String(data: data, encoding: .utf8) ?? "Unknown"
}
```

---

_For advanced patterns and comprehensive reference, see SKILL.md_
