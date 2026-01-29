# SwiftUI 6 Patterns

Modern SwiftUI development with @Observable, NavigationStack, state management, and view lifecycle.

## @Observable Macro

The @Observable macro (iOS 17+) replaces @ObservableObject with simpler, more performant observation.

### Basic Usage

Simple ViewModel:
```swift
import Observation

@Observable
class CounterViewModel {
    var count = 0
    
    func increment() { count += 1 }
    func decrement() { count -= 1 }
}

struct CounterView: View {
    @State private var viewModel = CounterViewModel()
    
    var body: some View {
        VStack {
            Text("Count: \(viewModel.count)")
            HStack {
                Button("−") { viewModel.decrement() }
                Button("+") { viewModel.increment() }
            }
        }
    }
}
```

### Complex State Management

Full-Featured ViewModel:
```swift
@Observable
@MainActor
final class ProductListViewModel {
    private(set) var products: [Product] = []
    private(set) var isLoading = false
    var searchQuery = ""
    var selectedCategory: Category?
    
    var filteredProducts: [Product] {
        var result = products
        if !searchQuery.isEmpty {
            result = result.filter { $0.name.localizedCaseInsensitiveContains(searchQuery) }
        }
        if let category = selectedCategory {
            result = result.filter { $0.category == category }
        }
        return result
    }
    
    private let api: ProductAPIProtocol
    
    init(api: ProductAPIProtocol) { self.api = api }
    
    func loadProducts() async {
        isLoading = true
        defer { isLoading = false }
        products = (try? await api.fetchProducts()) ?? []
    }
}
```

### Environment Integration

Dependency Injection:
```swift
@Observable
class AuthService {
    var currentUser: User?
    var isAuthenticated: Bool { currentUser != nil }
}

@main
struct MyApp: App {
    @State private var authService = AuthService()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authService)
        }
    }
}

struct ProfileView: View {
    @Environment(AuthService.self) private var authService
    
    var body: some View {
        if let user = authService.currentUser {
            UserProfileContent(user: user)
        }
    }
}
```

## NavigationStack

### Programmatic Navigation

Navigation Router Pattern:
```swift
@Observable
class NavigationRouter {
    var path = NavigationPath()
    
    func push<D: Hashable>(_ destination: D) { path.append(destination) }
    func pop() { guard !path.isEmpty else { return }; path.removeLast() }
    func popToRoot() { path.removeLast(path.count) }
}

enum AppDestination: Hashable {
    case productDetail(Product)
    case settings
    case userProfile(User)
}

struct RootView: View {
    @State private var router = NavigationRouter()
    
    var body: some View {
        NavigationStack(path: $router.path) {
            HomeView()
                .navigationDestination(for: AppDestination.self) { destination in
                    switch destination {
                    case .productDetail(let product): ProductDetailView(product: product)
                    case .settings: SettingsView()
                    case .userProfile(let user): UserProfileView(user: user)
                    }
                }
        }
        .environment(router)
    }
}

struct HomeView: View {
    @Environment(NavigationRouter.self) private var router
    
    var body: some View {
        Button("View Settings") { router.push(AppDestination.settings) }
    }
}
```

## State Management

### @State vs @Binding

Local State:
```swift
struct ToggleCard: View {
    @State private var isExpanded = false
    
    var body: some View {
        VStack {
            Button { withAnimation { isExpanded.toggle() } } label: {
                HStack {
                    Text("Details")
                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                }
            }
            if isExpanded { Text("Expanded content") }
        }
    }
}
```

Binding for Two-Way Communication:
```swift
struct FilterSheet: View {
    @Binding var selectedCategory: Category?
    
    var body: some View {
        Picker("Category", selection: $selectedCategory) {
            Text("All").tag(nil as Category?)
            ForEach(Category.allCases) { category in
                Text(category.name).tag(category as Category?)
            }
        }
    }
}
```

### @Bindable for Observable

Using @Bindable:
```swift
@Observable
class FormViewModel {
    var name = ""
    var email = ""
    var isValid: Bool { !name.isEmpty && email.contains("@") }
}

struct RegistrationForm: View {
    @Bindable var viewModel: FormViewModel
    
    var body: some View {
        Form {
            TextField("Name", text: $viewModel.name)
            TextField("Email", text: $viewModel.email)
            Button("Submit") { }.disabled(!viewModel.isValid)
        }
    }
}
```

## View Lifecycle

### Task Modifier

Async Operations:
```swift
struct UserProfileView: View {
    let userId: String
    @State private var user: User?
    
    var body: some View {
        Group {
            if let user { UserContent(user: user) }
            else { ProgressView() }
        }
        .task { user = try? await api.fetchUser(userId) }
    }
}
```

Task with ID for Refresh:
```swift
struct SearchResultsView: View {
    let query: String
    @State private var results: [SearchResult] = []
    
    var body: some View {
        List(results) { SearchResultRow(result: $0) }
            .task(id: query) {
                results = try? await searchService.search(query) ?? []
            }
    }
}
```

### onChange Modifier

Responding to State Changes:
```swift
struct SearchView: View {
    @State private var searchText = ""
    @State private var debouncedText = ""
    
    var body: some View {
        TextField("Search", text: $searchText)
            .onChange(of: searchText) { _, newValue in
                Task {
                    try? await Task.sleep(for: .milliseconds(300))
                    if searchText == newValue { debouncedText = newValue }
                }
            }
    }
}
```

## Advanced Patterns

### Preference Key

Size Preference:
```swift
struct SizePreferenceKey: PreferenceKey {
    static var defaultValue: CGSize = .zero
    static func reduce(value: inout CGSize, nextValue: () -> CGSize) { value = nextValue() }
}

extension View {
    func readSize(_ onChange: @escaping (CGSize) -> Void) -> some View {
        background(GeometryReader { Color.clear.preference(key: SizePreferenceKey.self, value: $0.size) })
            .onPreferenceChange(SizePreferenceKey.self, perform: onChange)
    }
}
```

### Custom View Modifiers

Reusable Card Style:
```swift
struct CardModifier: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding()
            .background(Color(.systemBackground))
            .cornerRadius(12)
            .shadow(radius: 4)
    }
}

extension View {
    func cardStyle() -> some View { modifier(CardModifier()) }
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
Context7: /apple/swift, /SwiftUIX/SwiftUIX
