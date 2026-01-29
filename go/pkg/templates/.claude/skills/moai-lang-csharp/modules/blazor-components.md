# Blazor Components Development

Comprehensive guide to Blazor development including Server, WASM, InteractiveServer modes, and component patterns.

---

## Blazor Render Modes

### Understanding Render Modes

```csharp
// Static Server Rendering (default) - No interactivity
@page "/static"

// Interactive Server - SignalR connection
@page "/server"
@rendermode InteractiveServer

// Interactive WebAssembly - Runs in browser
@page "/wasm"
@rendermode InteractiveWebAssembly

// Interactive Auto - Server first, then WASM
@page "/auto"
@rendermode InteractiveAuto
```

### Global Render Mode Configuration

```csharp
// App.razor
<!DOCTYPE html>
<html>
<head>
    <HeadOutlet @rendermode="InteractiveServer" />
</head>
<body>
    <Routes @rendermode="InteractiveServer" />
    <script src="_framework/blazor.web.js"></script>
</body>
</html>
```

---

## Component Basics

### Simple Component

```csharp
// Components/Counter.razor
@page "/counter"
@rendermode InteractiveServer

<h1>Counter</h1>

<p role="status">Current count: @currentCount</p>

<button class="btn btn-primary" @onclick="IncrementCount">Click me</button>

@code {
    private int currentCount = 0;

    private void IncrementCount()
    {
        currentCount++;
    }
}
```

### Component with Parameters

```csharp
// Components/UserCard.razor
<div class="card">
    <div class="card-body">
        <h5 class="card-title">@User.Name</h5>
        <p class="card-text">@User.Email</p>
        @if (ShowActions)
        {
            <button class="btn btn-primary" @onclick="() => OnEdit.InvokeAsync(User)">
                Edit
            </button>
            <button class="btn btn-danger" @onclick="() => OnDelete.InvokeAsync(User)">
                Delete
            </button>
        }
    </div>
</div>

@code {
    [Parameter, EditorRequired]
    public UserDto User { get; set; } = default!;

    [Parameter]
    public bool ShowActions { get; set; } = true;

    [Parameter]
    public EventCallback<UserDto> OnEdit { get; set; }

    [Parameter]
    public EventCallback<UserDto> OnDelete { get; set; }
}
```

### Component with Cascading Parameters

```csharp
// MainLayout.razor
@inherits LayoutComponentBase

<CascadingValue Value="@_theme">
    <CascadingValue Value="@_currentUser" Name="CurrentUser">
        <div class="page">
            @Body
        </div>
    </CascadingValue>
</CascadingValue>

@code {
    private Theme _theme = Theme.Light;
    private UserDto? _currentUser;
}

// ChildComponent.razor
@code {
    [CascadingParameter]
    public Theme Theme { get; set; }

    [CascadingParameter(Name = "CurrentUser")]
    public UserDto? CurrentUser { get; set; }
}
```

---

## Lifecycle Methods

### Component Lifecycle

```csharp
@page "/lifecycle"
@implements IAsyncDisposable

<h1>Lifecycle Demo</h1>

@code {
    [Parameter]
    public Guid? Id { get; set; }

    // Called when component is initialized (once)
    protected override void OnInitialized()
    {
        Console.WriteLine("OnInitialized");
    }

    // Async version for data loading
    protected override async Task OnInitializedAsync()
    {
        Console.WriteLine("OnInitializedAsync");
        await LoadDataAsync();
    }

    // Called when parameters are set/changed
    protected override void OnParametersSet()
    {
        Console.WriteLine("OnParametersSet");
    }

    // Called after each render
    protected override void OnAfterRender(bool firstRender)
    {
        if (firstRender)
        {
            Console.WriteLine("First render complete");
        }
    }

    // Async version
    protected override async Task OnAfterRenderAsync(bool firstRender)
    {
        if (firstRender)
        {
            await InitializeJsAsync();
        }
    }

    // Should the component re-render?
    protected override bool ShouldRender()
    {
        return true; // Custom logic here
    }

    // Cleanup
    public async ValueTask DisposeAsync()
    {
        Console.WriteLine("Disposing");
        await CleanupAsync();
    }

    private Task LoadDataAsync() => Task.CompletedTask;
    private Task InitializeJsAsync() => Task.CompletedTask;
    private Task CleanupAsync() => Task.CompletedTask;
}
```

---

## Data Binding

### Two-Way Binding

```csharp
@page "/binding"
@rendermode InteractiveServer

<h1>Data Binding</h1>

<!-- Simple binding -->
<input @bind="name" />
<p>Hello, @name!</p>

<!-- Bind with event -->
<input @bind="searchTerm" @bind:event="oninput" />

<!-- Bind with format -->
<input type="date" @bind="birthDate" @bind:format="yyyy-MM-dd" />

<!-- Bind with after -->
<input @bind="email" @bind:after="ValidateEmail" />

@code {
    private string name = "";
    private string searchTerm = "";
    private DateTime birthDate = DateTime.Today;
    private string email = "";
    private bool isEmailValid;

    private void ValidateEmail()
    {
        isEmailValid = email.Contains('@');
    }
}
```

### Custom Binding

```csharp
// PasswordInput.razor
<input type="password"
       value="@Value"
       @oninput="HandleInput"
       class="@CssClass" />

@code {
    [Parameter]
    public string Value { get; set; } = "";

    [Parameter]
    public EventCallback<string> ValueChanged { get; set; }

    [Parameter]
    public string CssClass { get; set; } = "form-control";

    private async Task HandleInput(ChangeEventArgs e)
    {
        var newValue = e.Value?.ToString() ?? "";
        await ValueChanged.InvokeAsync(newValue);
    }
}

// Usage
<PasswordInput @bind-Value="password" />
```

---

## Forms and Validation

### EditForm with DataAnnotations

```csharp
@page "/register"
@rendermode InteractiveServer
@inject IUserService UserService

<EditForm Model="@model" OnValidSubmit="HandleSubmit">
    <DataAnnotationsValidator />
    <ValidationSummary />

    <div class="mb-3">
        <label class="form-label">Name</label>
        <InputText @bind-Value="model.Name" class="form-control" />
        <ValidationMessage For="@(() => model.Name)" />
    </div>

    <div class="mb-3">
        <label class="form-label">Email</label>
        <InputText @bind-Value="model.Email" class="form-control" />
        <ValidationMessage For="@(() => model.Email)" />
    </div>

    <div class="mb-3">
        <label class="form-label">Password</label>
        <InputText type="password" @bind-Value="model.Password" class="form-control" />
        <ValidationMessage For="@(() => model.Password)" />
    </div>

    <button type="submit" class="btn btn-primary" disabled="@isSubmitting">
        @(isSubmitting ? "Submitting..." : "Register")
    </button>
</EditForm>

@code {
    private RegisterModel model = new();
    private bool isSubmitting;

    private async Task HandleSubmit()
    {
        isSubmitting = true;
        try
        {
            await UserService.RegisterAsync(model);
            // Navigate or show success
        }
        finally
        {
            isSubmitting = false;
        }
    }

    public class RegisterModel
    {
        [Required]
        [StringLength(100, MinimumLength = 2)]
        public string Name { get; set; } = "";

        [Required]
        [EmailAdddess]
        public string Email { get; set; } = "";

        [Required]
        [MinLength(8)]
        public string Password { get; set; } = "";
    }
}
```

### FluentValidation Integration

```csharp
// Install: dotnet add package Blazored.FluentValidation

@page "/register-fluent"
@using Blazored.FluentValidation

<EditForm Model="@model" OnValidSubmit="HandleSubmit">
    <FluentValidationValidator />
    <ValidationSummary />

    <!-- Form fields -->
</EditForm>

// RegisterModelValidator.cs
public class RegisterModelValidator : AbstractValidator<RegisterModel>
{
    public RegisterModelValidator()
    {
        RuleFor(x => x.Name)
            .NotEmpty()
            .MaximumLength(100);

        RuleFor(x => x.Email)
            .NotEmpty()
            .EmailAdddess();

        RuleFor(x => x.Password)
            .NotEmpty()
            .MinimumLength(8);
    }
}
```

---

## JavaScript Interop

### Calling JavaScript from C#

```csharp
@page "/js-interop"
@rendermode InteractiveServer
@inject IJSRuntime JS

<button @onclick="ShowAlert">Show Alert</button>
<button @onclick="GetWindowSize">Get Window Size</button>

<p>Window: @windowWidth x @windowHeight</p>

@code {
    private int windowWidth;
    private int windowHeight;

    private async Task ShowAlert()
    {
        await JS.InvokeVoidAsync("alert", "Hello from Blazor!");
    }

    private async Task GetWindowSize()
    {
        var dimensions = await JS.InvokeAsync<WindowDimensions>("getWindowDimensions");
        windowWidth = dimensions.Width;
        windowHeight = dimensions.Height;
    }

    public record WindowDimensions(int Width, int Height);
}
```

```javascript
// wwwroot/js/interop.js
window.getWindowDimensions = function() {
    return {
        width: window.innerWidth,
        height: window.innerHeight
    };
};

window.focusElement = function(elementId) {
    document.getElementById(elementId)?.focus();
};
```

### Calling C# from JavaScript

```csharp
@page "/js-callback"
@rendermode InteractiveServer
@implements IAsyncDisposable
@inject IJSRuntime JS

<div id="drop-zone" @ref="dropZone">Drop files here</div>

@code {
    private ElementReference dropZone;
    private DotNetObjectReference<FileDrop>? dotNetRef;

    protected override async Task OnAfterRenderAsync(bool firstRender)
    {
        if (firstRender)
        {
            dotNetRef = DotNetObjectReference.Create(this);
            await JS.InvokeVoidAsync("initializeDropZone", dropZone, dotNetRef);
        }
    }

    [JSInvokable]
    public void OnFileDropped(string fileName, long fileSize)
    {
        Console.WriteLine($"File dropped: {fileName} ({fileSize} bytes)");
        StateHasChanged();
    }

    public async ValueTask DisposeAsync()
    {
        if (dotNetRef is not null)
        {
            await JS.InvokeVoidAsync("cleanupDropZone", dropZone);
            dotNetRef.Dispose();
        }
    }
}
```

---

## State Management

### Component State

```csharp
@page "/state"
@rendermode InteractiveServer

@code {
    // Simple state
    private List<TodoItem> todos = [];
    
    private void AddTodo(string title)
    {
        todos.Add(new TodoItem { Title = title });
    }
}
```

### Cascading State

```csharp
// AppState.cs
public class AppState
{
    public UserDto? CurrentUser { get; private set; }
    public event Action? OnChange;

    public void SetUser(UserDto user)
    {
        CurrentUser = user;
        OnChange?.Invoke();
    }

    public void ClearUser()
    {
        CurrentUser = null;
        OnChange?.Invoke();
    }
}

// Program.cs
builder.Services.AddScoped<AppState>();

// MainLayout.razor
@inject AppState AppState
@implements IDisposable

<CascadingValue Value="@AppState">
    @Body
</CascadingValue>

@code {
    protected override void OnInitialized()
    {
        AppState.OnChange += StateHasChanged;
    }

    public void Dispose()
    {
        AppState.OnChange -= StateHasChanged;
    }
}

// Any child component
@code {
    [CascadingParameter]
    public AppState AppState { get; set; } = default!;

    private void Login()
    {
        AppState.SetUser(new UserDto { Name = "John" });
    }
}
```

---

## Common Patterns

### Loading States

```csharp
@page "/users"
@rendermode InteractiveServer
@inject IUserService UserService

@if (_loading)
{
    <div class="spinner-border" role="status">
        <span class="visually-hidden">Loading...</span>
    </div>
}
else if (_error is not null)
{
    <div class="alert alert-danger">@_error</div>
}
else if (_users is null || !_users.Any())
{
    <div class="alert alert-info">No users found.</div>
}
else
{
    <table class="table">
        @foreach (var user in _users)
        {
            <tr>
                <td>@user.Name</td>
                <td>@user.Email</td>
            </tr>
        }
    </table>
}

@code {
    private List<UserDto>? _users;
    private bool _loading = true;
    private string? _error;

    protected override async Task OnInitializedAsync()
    {
        try
        {
            _users = await UserService.GetAllAsync();
        }
        catch (Exception ex)
        {
            _error = ex.Message;
        }
        finally
        {
            _loading = false;
        }
    }
}
```

### Modal Dialog

```csharp
// Components/Modal.razor
<div class="modal @(IsVisible ? "show d-block" : "")" tabindex="-1">
    <div class="modal-dialog">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title">@Title</h5>
                <button type="button" class="btn-close" @onclick="Close"></button>
            </div>
            <div class="modal-body">
                @ChildContent
            </div>
            <div class="modal-footer">
                @if (FooterContent is not null)
                {
                    @FooterContent
                }
                else
                {
                    <button type="button" class="btn btn-secondary" @onclick="Close">
                        Close
                    </button>
                }
            </div>
        </div>
    </div>
</div>
@if (IsVisible)
{
    <div class="modal-backdrop fade show"></div>
}

@code {
    [Parameter]
    public bool IsVisible { get; set; }

    [Parameter]
    public EventCallback<bool> IsVisibleChanged { get; set; }

    [Parameter]
    public string Title { get; set; } = "Modal";

    [Parameter]
    public RenderFragment? ChildContent { get; set; }

    [Parameter]
    public RenderFragment? FooterContent { get; set; }

    private async Task Close()
    {
        await IsVisibleChanged.InvokeAsync(false);
    }
}

// Usage
<button @onclick="() => showModal = true">Open Modal</button>

<Modal @bind-IsVisible="showModal" Title="Confirm Delete">
    <p>Are you sure you want to delete this item?</p>
    <FooterContent>
        <button class="btn btn-secondary" @onclick="() => showModal = false">Cancel</button>
        <button class="btn btn-danger" @onclick="ConfirmDelete">Delete</button>
    </FooterContent>
</Modal>
```

### Virtualization for Large Lists

```csharp
@page "/large-list"
@rendermode InteractiveServer
@inject IUserService UserService

<Virtualize Items="@users" Context="user" ItemSize="50">
    <ItemContent>
        <div class="user-row" style="height: 50px;">
            @user.Name - @user.Email
        </div>
    </ItemContent>
    <Placeholder>
        <div class="user-row placeholder" style="height: 50px;">
            Loading...
        </div>
    </Placeholder>
</Virtualize>

@code {
    private List<UserDto> users = [];

    protected override async Task OnInitializedAsync()
    {
        users = await UserService.GetAllAsync();
    }
}

// With ItemsProvider for server-side paging
<Virtualize ItemsProvider="LoadUsers" Context="user">
    <div>@user.Name</div>
</Virtualize>

@code {
    private async ValueTask<ItemsProviderResult<UserDto>> LoadUsers(
        ItemsProviderRequest request)
    {
        var result = await UserService.GetPagedAsync(
            request.StartIndex,
            request.Count,
            request.CancellationToken);
        
        return new ItemsProviderResult<UserDto>(result.Items, result.TotalCount);
    }
}
```

---

## Error Boundaries

```csharp
// Components/ErrorBoundary.razor
<ErrorBoundary @ref="errorBoundary">
    <ChildContent>
        @ChildContent
    </ChildContent>
    <ErrorContent Context="exception">
        <div class="alert alert-danger">
            <h4>An error occurred</h4>
            <p>@exception.Message</p>
            <button class="btn btn-outline-danger" @onclick="Recover">
                Try Again
            </button>
        </div>
    </ErrorContent>
</ErrorBoundary>

@code {
    private ErrorBoundary? errorBoundary;

    [Parameter]
    public RenderFragment? ChildContent { get; set; }

    private void Recover()
    {
        errorBoundary?.Recover();
    }
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
