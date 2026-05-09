# Swift Development Examples

Production-ready code examples for iOS/macOS development with Swift 6.

## Complete Feature: User Authentication

### Full Authentication Flow

```swift
// MARK: - Domain Layer

/// Authentication errors with typed throws
enum AuthError: Error, LocalizedError {
    case invalidCredentials
    case networkError(underlying: Error)
    case tokenExpired
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Invalid email or password"
        case .networkError(let error): return "Network error: \(error.localizedDescription)"
        case .tokenExpired: return "Session expired. Please login again"
        case .unauthorized: return "Unauthorized access"
        }
    }
}

struct User: Codable, Identifiable, Sendable {
    let id: String
    let email: String
    let name: String
    let avatarURL: URL?
}

struct AuthTokens: Codable, Sendable {
    let accessToken: String
    let refreshToken: String
    let expiresAt: Date
}

// MARK: - Data Layer

protocol AuthAPIProtocol: Sendable {
    func login(email: String, password: String) async throws(AuthError) -> AuthTokens
    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens
    func fetchUser(token: String) async throws(AuthError) -> User
    func logout(token: String) async throws(AuthError)
}

actor AuthAPI: AuthAPIProtocol {
    private let session: URLSession
    private let baseURL: URL

    init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    func login(email: String, password: String) async throws(AuthError) -> AuthTokens {
        let request = try makeRequest(
            path: "/auth/login",
            method: "POST",
            body: ["email": email, "password": password]
        )

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }

            switch httpResponse.statusCode {
            case 200:
                return try JSONDecoder().decode(AuthTokens.self, from: data)
            case 401:
                throw AuthError.invalidCredentials
            default:
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens {
        var request = try makeRequest(path: "/auth/refresh", method: "POST")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                throw AuthError.tokenExpired
            }
            return try JSONDecoder().decode(AuthTokens.self, from: data)
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func fetchUser(token: String) async throws(AuthError) -> User {
        var request = try makeRequest(path: "/auth/me", method: "GET")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }

            switch httpResponse.statusCode {
            case 200:
                return try JSONDecoder().decode(User.self, from: data)
            case 401:
                throw AuthError.unauthorized
            default:
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func logout(token: String) async throws(AuthError) {
        var request = try makeRequest(path: "/auth/logout", method: "POST")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (_, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    private func makeRequest(path: String, method: String, body: [String: Any]? = nil) throws -> URLRequest {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body = body {
            request.httpBody = try JSONSerialization.data(withJSONObject: body)
        }

        return request
    }
}

// MARK: - Presentation Layer

@Observable
@MainActor
final class AuthViewModel {
    private let api: AuthAPIProtocol
    private let keychain: KeychainProtocol

    var user: User?
    var isLoading = false
    var error: AuthError?
    var isAuthenticated: Bool { user != nil }

    init(api: AuthAPIProtocol, keychain: KeychainProtocol) {
        self.api = api
        self.keychain = keychain
    }

    func login(email: String, password: String) async {
        isLoading = true
        error = nil
        defer { isLoading = false }

        do {
            let tokens = try await api.login(email: email, password: password)
            try keychain.save(tokens: tokens)

            let user = try await api.fetchUser(token: tokens.accessToken)
            self.user = user
        } catch {
            self.error = error
        }
    }

    func restoreSession() async {
        guard let tokens = try? keychain.loadTokens(),
              tokens.expiresAt > Date() else {
            return
        }

        do {
            user = try await api.fetchUser(token: tokens.accessToken)
        } catch AuthError.unauthorized, AuthError.tokenExpired {
            do {
                let newTokens = try await api.refreshToken(tokens.refreshToken)
                try keychain.save(tokens: newTokens)
                user = try await api.fetchUser(token: newTokens.accessToken)
            } catch {
                try? keychain.deleteTokens()
            }
        } catch {
            self.error = error
        }
    }

    func logout() async {
        guard let tokens = try? keychain.loadTokens() else { return }

        do {
            try await api.logout(token: tokens.accessToken)
        } catch {
            // Log error but continue with local logout
        }

        try? keychain.deleteTokens()
        user = nil
    }
}

// MARK: - SwiftUI Views

struct LoginView: View {
    @Environment(AuthViewModel.self) private var viewModel
    @State private var email = ""
    @State private var password = ""

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Email", text: $email)
                        .textContentType(.emailAdddess)
                        .keyboardType(.emailAdddess)
                        .autocapitalization(.none)

                    SecureField("Password", text: $password)
                        .textContentType(.password)
                }

                if let error = viewModel.error {
                    Section {
                        Text(error.localizedDescription)
                            .foregroundColor(.red)
                    }
                }

                Section {
                    Button {
                        Task { await viewModel.login(email: email, password: password) }
                    } label: {
                        if viewModel.isLoading {
                            ProgressView()
                                .frame(maxWidth: .infinity)
                        } else {
                            Text("Sign In")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .disabled(email.isEmpty || password.isEmpty || viewModel.isLoading)
                }
            }
            .navigationTitle("Login")
        }
    }
}
```

## Network Layer Example

### URLSession with Async/Await

```swift
actor NetworkClient {
    private let session: URLSession
    private let baseURL: URL
    private let decoder: JSONDecoder

    init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
        self.decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
    }

    func get<T: Decodable>(_ path: String, query: [String: String] = [:]) async throws -> T {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: true)!
        components.queryItems = query.map { URLQueryItem(name: $0.key, value: $0.value) }

        let (data, response) = try await session.data(from: components.url!)
        try validateResponse(response)
        return try decoder.decode(T.self, from: data)
    }

    func post<T: Decodable, B: Encodable>(_ path: String, body: B) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await session.data(for: request)
        try validateResponse(response)
        return try decoder.decode(T.self, from: data)
    }

    private func validateResponse(_ response: URLResponse) throws {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.invalidResponse
        }
        guard 200..<300 ~= httpResponse.statusCode else {
            throw NetworkError.statusCode(httpResponse.statusCode)
        }
    }
}

enum NetworkError: Error {
    case invalidResponse
    case statusCode(Int)
}
```

## SwiftUI Components

### Paginated List View

```swift
@Observable
@MainActor
final class PaginatedListViewModel<Item: Identifiable & Decodable> {
    private(set) var items: [Item] = []
    private(set) var isLoading = false
    private(set) var error: Error?
    private(set) var hasMorePages = true

    private var currentPage = 1
    private let pageSize = 20
    private let fetchPage: (Int, Int) async throws -> (items: [Item], hasMore: Bool)

    init(fetchPage: @escaping (Int, Int) async throws -> (items: [Item], hasMore: Bool)) {
        self.fetchPage = fetchPage
    }

    func loadInitial() async {
        guard !isLoading else { return }

        isLoading = true
        error = nil
        currentPage = 1

        do {
            let result = try await fetchPage(currentPage, pageSize)
            items = result.items
            hasMorePages = result.hasMore
        } catch {
            self.error = error
        }

        isLoading = false
    }

    func loadMoreIfNeeded(currentItem: Item?) async {
        guard let item = currentItem,
              !isLoading,
              hasMorePages else { return }

        let thresholdIndex = items.index(items.endIndex, offsetBy: -5)
        guard items.firstIndex(where: { $0.id == item.id as? Item.ID }) ?? 0 >= thresholdIndex else {
            return
        }

        await loadNextPage()
    }

    private func loadNextPage() async {
        isLoading = true
        currentPage += 1

        do {
            let result = try await fetchPage(currentPage, pageSize)
            items.append(contentsOf: result.items)
            hasMorePages = result.hasMore
        } catch {
            currentPage -= 1
            self.error = error
        }

        isLoading = false
    }
}

struct PaginatedListView<Item: Identifiable & Decodable, RowContent: View>: View {
    @State private var viewModel: PaginatedListViewModel<Item>
    let rowContent: (Item) -> RowContent

    init(
        fetchPage: @escaping (Int, Int) async throws -> (items: [Item], hasMore: Bool),
        @ViewBuilder rowContent: @escaping (Item) -> RowContent
    ) {
        _viewModel = State(initialValue: PaginatedListViewModel(fetchPage: fetchPage))
        self.rowContent = rowContent
    }

    var body: some View {
        List {
            ForEach(viewModel.items) { item in
                rowContent(item)
                    .task {
                        await viewModel.loadMoreIfNeeded(currentItem: item)
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
            await viewModel.loadInitial()
        }
        .task {
            await viewModel.loadInitial()
        }
    }
}
```

### Async Image with Cache

```swift
actor ImageCacheManager {
    static let shared = ImageCacheManager()

    private var cache = NSCache<NSURL, UIImage>()
    private var inProgress: [URL: Task<UIImage, Error>] = [:]

    private init() {
        cache.countLimit = 100
        cache.totalCostLimit = 50 * 1024 * 1024 // 50MB
    }

    func image(for url: URL) async throws -> UIImage {
        // Check cache
        if let cached = cache.object(forKey: url as NSURL) {
            return cached
        }

        // Check in-progress
        if let task = inProgress[url] {
            return try await task.value
        }

        // Start download
        let task = Task<UIImage, Error> {
            let (data, _) = try await URLSession.shared.data(from: url)
            guard let image = UIImage(data: data) else {
                throw ImageError.invalidData
            }
            return image
        }

        inProgress[url] = task

        do {
            let image = try await task.value
            cache.setObject(image, forKey: url as NSURL, cost: image.pngData()?.count ?? 0)
            inProgress[url] = nil
            return image
        } catch {
            inProgress[url] = nil
            throw error
        }
    }

    func clearCache() {
        cache.removeAllObjects()
    }
}

enum ImageError: Error {
    case invalidData
}

struct CachedAsyncImage: View {
    let url: URL?
    var placeholder: Image = Image(systemName: "photo")

    @State private var image: UIImage?
    @State private var isLoading = false

    var body: some View {
        Group {
            if let image = image {
                Image(uiImage: image)
                    .resizable()
                    .aspectRatio(contentMode: .fill)
            } else if isLoading {
                ProgressView()
            } else {
                placeholder
                    .resizable()
                    .aspectRatio(contentMode: .fit)
                    .foregroundColor(.gray)
            }
        }
        .task(id: url) {
            guard let url = url else { return }
            isLoading = true
            defer { isLoading = false }

            do {
                image = try await ImageCacheManager.shared.image(for: url)
            } catch {
                image = nil
            }
        }
    }
}
```

## Testing Examples

### XCTest with async

```swift
@MainActor
final class AuthViewModelTests: XCTestCase {
    var sut: AuthViewModel!
    var mockAPI: MockAuthAPI!
    var mockKeychain: MockKeychain!

    override func setUp() {
        mockAPI = MockAuthAPI()
        mockKeychain = MockKeychain()
        sut = AuthViewModel(api: mockAPI, keychain: mockKeychain)
    }

    override func tearDown() {
        sut = nil
        mockAPI = nil
        mockKeychain = nil
    }

    func testLoginSuccess() async {
        // Given
        mockAPI.mockTokens = AuthTokens(
            accessToken: "test-token",
            refreshToken: "refresh-token",
            expiresAt: Date().addingTimeInterval(3600)
        )
        mockAPI.mockUser = User(id: "1", email: "test@example.com", name: "Test User", avatarURL: nil)

        // When
        await sut.login(email: "test@example.com", password: "password123")

        // Then
        XCTAssertNotNil(sut.user)
        XCTAssertEqual(sut.user?.email, "test@example.com")
        XCTAssertNil(sut.error)
        XCTAssertFalse(sut.isLoading)
    }

    func testLoginInvalidCredentials() async {
        // Given
        mockAPI.loginError = .invalidCredentials

        // When
        await sut.login(email: "test@example.com", password: "wrong")

        // Then
        XCTAssertNil(sut.user)
        XCTAssertEqual(sut.error, .invalidCredentials)
    }

    func testRestoreSessionWithValidToken() async {
        // Given
        let tokens = AuthTokens(
            accessToken: "valid-token",
            refreshToken: "refresh-token",
            expiresAt: Date().addingTimeInterval(3600)
        )
        mockKeychain.savedTokens = tokens
        mockAPI.mockUser = User(id: "1", email: "test@example.com", name: "Test User", avatarURL: nil)

        // When
        await sut.restoreSession()

        // Then
        XCTAssertNotNil(sut.user)
        XCTAssertEqual(sut.user?.email, "test@example.com")
    }

    func testLogoutClearsUser() async {
        // Given
        sut.user = User(id: "1", email: "test@example.com", name: "Test User", avatarURL: nil)
        mockKeychain.savedTokens = AuthTokens(
            accessToken: "token",
            refreshToken: "refresh",
            expiresAt: Date().addingTimeInterval(3600)
        )

        // When
        await sut.logout()

        // Then
        XCTAssertNil(sut.user)
        XCTAssertNil(mockKeychain.savedTokens)
    }
}

// MARK: - Mocks

class MockAuthAPI: AuthAPIProtocol {
    var mockTokens: AuthTokens?
    var mockUser: User?
    var loginError: AuthError?
    var fetchUserError: AuthError?

    func login(email: String, password: String) async throws(AuthError) -> AuthTokens {
        if let error = loginError { throw error }
        guard let tokens = mockTokens else {
            throw .networkError(underlying: URLError(.unknown))
        }
        return tokens
    }

    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens {
        guard let tokens = mockTokens else {
            throw .tokenExpired
        }
        return tokens
    }

    func fetchUser(token: String) async throws(AuthError) -> User {
        if let error = fetchUserError { throw error }
        guard let user = mockUser else {
            throw .unauthorized
        }
        return user
    }

    func logout(token: String) async throws(AuthError) {
        // No-op for testing
    }
}

protocol KeychainProtocol: Sendable {
    func save(tokens: AuthTokens) throws
    func loadTokens() throws -> AuthTokens?
    func deleteTokens() throws
}

class MockKeychain: KeychainProtocol {
    var savedTokens: AuthTokens?

    func save(tokens: AuthTokens) throws {
        savedTokens = tokens
    }

    func loadTokens() throws -> AuthTokens? {
        return savedTokens
    }

    func deleteTokens() throws {
        savedTokens = nil
    }
}
```

### Combine Testing

```swift
import Combine
import XCTest

final class SearchViewModelTests: XCTestCase {
    var sut: SearchViewModel!
    var mockService: MockSearchService!
    var cancellables: Set<AnyCancellable>!

    override func setUp() {
        mockService = MockSearchService()
        sut = SearchViewModel(searchService: mockService)
        cancellables = []
    }

    override func tearDown() {
        sut = nil
        mockService = nil
        cancellables = nil
    }

    func testSearchDebounces() {
        // Given
        let expectation = expectation(description: "Search debounce")
        var searchCount = 0

        mockService.searchHandler = { query in
            searchCount += 1
            return [SearchResult(id: "1", title: query)]
        }

        sut.$results
            .dropFirst()
            .sink { results in
                if !results.isEmpty {
                    expectation.fulfill()
                }
            }
            .store(in: &cancellables)

        // When - rapid input
        sut.searchText = "t"
        sut.searchText = "te"
        sut.searchText = "tes"
        sut.searchText = "test"

        // Then
        wait(for: [expectation], timeout: 1.0)
        XCTAssertEqual(searchCount, 1, "Should only search once after debounce")
    }
}

class MockSearchService: SearchServiceProtocol {
    var searchHandler: ((String) -> [SearchResult])?

    func search(_ query: String) -> AnyPublisher<[SearchResult], Error> {
        let results = searchHandler?(query) ?? []
        return Just(results)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
}
```

## SwiftData Example

### Complete CRUD Operations

```swift
import SwiftData
import SwiftUI

@Model
final class Task {
    var id: UUID
    var title: String
    var isCompleted: Bool
    var priority: Priority
    var dueDate: Date?
    var createdAt: Date

    @Relationship(deleteRule: .cascade, inverse: \Subtask.parentTask)
    var subtasks: [Subtask]

    init(title: String, priority: Priority = .medium, dueDate: Date? = nil) {
        self.id = UUID()
        self.title = title
        self.isCompleted = false
        self.priority = priority
        self.dueDate = dueDate
        self.createdAt = Date()
        self.subtasks = []
    }

    enum Priority: String, Codable, CaseIterable {
        case low, medium, high

        var color: Color {
            switch self {
            case .low: return .green
            case .medium: return .orange
            case .high: return .red
            }
        }
    }
}

@Model
final class Subtask {
    var id: UUID
    var title: String
    var isCompleted: Bool
    var parentTask: Task?

    init(title: String) {
        self.id = UUID()
        self.title = title
        self.isCompleted = false
    }
}

@MainActor
final class TaskRepository: ObservableObject {
    private let modelContext: ModelContext

    init(modelContext: ModelContext) {
        self.modelContext = modelContext
    }

    func fetchTasks(completed: Bool? = nil) throws -> [Task] {
        var descriptor = FetchDescriptor<Task>(
            sortBy: [
                SortDescriptor(\.priority, order: .reverse),
                SortDescriptor(\.createdAt, order: .reverse)
            ]
        )

        if let completed = completed {
            descriptor.predicate = #Predicate { $0.isCompleted == completed }
        }

        return try modelContext.fetch(descriptor)
    }

    func addTask(_ task: Task) throws {
        modelContext.insert(task)
        try modelContext.save()
    }

    func updateTask(_ task: Task) throws {
        try modelContext.save()
    }

    func deleteTask(_ task: Task) throws {
        modelContext.delete(task)
        try modelContext.save()
    }

    func toggleCompletion(_ task: Task) throws {
        task.isCompleted.toggle()
        try modelContext.save()
    }
}

struct TaskListView: View {
    @Environment(\.modelContext) private var modelContext
    @Query(sort: \Task.createdAt, order: .reverse) private var tasks: [Task]
    @State private var showAddTask = false

    var body: some View {
        NavigationStack {
            List {
                ForEach(tasks) { task in
                    TaskRow(task: task)
                }
                .onDelete(perform: deleteTasks)
            }
            .navigationTitle("Tasks")
            .toolbar {
                Button {
                    showAddTask = true
                } label: {
                    Image(systemName: "plus")
                }
            }
            .sheet(isPresented: $showAddTask) {
                AddTaskView()
            }
        }
    }

    private func deleteTasks(at offsets: IndexSet) {
        for index in offsets {
            modelContext.delete(tasks[index])
        }
    }
}

struct TaskRow: View {
    @Bindable var task: Task

    var body: some View {
        HStack {
            Button {
                task.isCompleted.toggle()
            } label: {
                Image(systemName: task.isCompleted ? "checkmark.circle.fill" : "circle")
                    .foregroundColor(task.isCompleted ? .green : .gray)
            }
            .buttonStyle(.plain)

            VStack(alignment: .leading) {
                Text(task.title)
                    .strikethrough(task.isCompleted)

                if let dueDate = task.dueDate {
                    Text(dueDate, style: .date)
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            Circle()
                .fill(task.priority.color)
                .frame(width: 10, height: 10)
        }
    }
}
```

---

Version: 1.0.0
Last Updated: 2025-12-07
