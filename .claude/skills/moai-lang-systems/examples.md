# Systems Programming Examples

Production-ready code examples for Go 1.23 and Rust 1.91.

---

## REST API Examples

### Go Fiber Complete API

```go
// main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/limiter"
    "github.com/gofiber/fiber/v3/middleware/logger"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
    Port        string
    DatabaseURL string
}

type App struct {
    fiber  *fiber.App
    db     *pgxpool.Pool
    config Config
}

func main() {
    config := Config{
        Port:        getEnv("PORT", "3000"),
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/myapp"),
    }

    app, err := NewApp(config)
    if err != nil {
        log.Fatal(err)
    }

    // Graceful shutdown
    go func() {
        if err := app.Start(); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")
    app.Shutdown()
}

func NewApp(config Config) (*App, error) {
    // Database pool
    pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
    if err != nil {
        return nil, err
    }

    // Fiber app
    f := fiber.New(fiber.Config{
        ErrorHandler: errorHandler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    })

    app := &App{fiber: f, db: pool, config: config}
    app.setupMiddleware()
    app.setupRoutes()

    return app, nil
}

func (a *App) setupMiddleware() {
    a.fiber.Use(recover.New())
    a.fiber.Use(logger.New())
    a.fiber.Use(cors.New())
    a.fiber.Use(limiter.New(limiter.Config{
        Max:        100,
        Expiration: time.Minute,
    }))
}

func (a *App) setupRoutes() {
    // Health check
    a.fiber.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // API v1
    api := a.fiber.Group("/api/v1")

    // Users
    users := api.Group("/users")
    users.Get("/", a.listUsers)
    users.Get("/:id", a.getUser)
    users.Post("/", a.createUser)
    users.Put("/:id", a.updateUser)
    users.Delete("/:id", a.deleteUser)
}

func (a *App) Start() error {
    return a.fiber.Listen(":" + a.config.Port)
}

func (a *App) Shutdown() {
    a.db.Close()
    a.fiber.Shutdown()
}

// Handlers
type User struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func (a *App) listUsers(c fiber.Ctx) error {
    limit := c.QueryInt("limit", 10)
    offset := c.QueryInt("offset", 0)

    rows, err := a.db.Query(c.Context(),
        "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
        limit, offset)
    if err != nil {
        return err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return err
        }
        users = append(users, u)
    }

    return c.JSON(users)
}

func (a *App) getUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
    }

    var u User
    err = a.db.QueryRow(c.Context(),
        "SELECT id, name, email, created_at FROM users WHERE id = $1", id).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    return c.JSON(u)
}

func (a *App) createUser(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    var u User
    err := a.db.QueryRow(c.Context(),
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
        req.Name, req.Email).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        return err
    }

    return c.Status(fiber.StatusCreated).JSON(u)
}

func (a *App) updateUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
    }

    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
    }

    var u User
    err = a.db.QueryRow(c.Context(),
        "UPDATE users SET name = $2, email = $3 WHERE id = $1 RETURNING id, name, email, created_at",
        id, req.Name, req.Email).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    return c.JSON(u)
}

func (a *App) deleteUser(c fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
    }

    result, err := a.db.Exec(c.Context(), "DELETE FROM users WHERE id = $1", id)
    if err != nil {
        return err
    }

    if result.RowsAffected() == 0 {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    return c.SendStatus(fiber.StatusNoContent)
}

func errorHandler(c fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    message := "Internal Server Error"

    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
        message = e.Message
    }

    return c.Status(code).JSON(fiber.Map{"error": message})
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Rust Axum Complete API

```rust
// src/main.rs
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{delete, get, post, put},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::net::SocketAddr;
use thiserror::Error;
use tokio::signal;
use tower_http::{cors::CorsLayer, trace::TraceLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Clone)]
struct AppState {
    db: PgPool,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Tracing
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(
            std::env::var("RUST_LOG").unwrap_or_else(|_| "info".into()),
        ))
        .with(tracing_subscriber::fmt::layer())
        .init();

    // Database
    let database_url = std::env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://localhost/myapp".into());

    let pool = PgPoolOptions::new()
        .max_connections(25)
        .connect(&database_url)
        .await?;

    let state = AppState { db: pool };

    // Router
    let app = Router::new()
        .route("/health", get(health_check))
        .nest("/api/v1", api_routes())
        .layer(TraceLayer::new_for_http())
        .layer(CorsLayer::permissive())
        .with_state(state);

    // Server
    let addr = SocketAddr::from(([0, 0, 0, 0], 3000));
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    Ok(())
}

fn api_routes() -> Router<AppState> {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route(
            "/users/:id",
            get(get_user).put(update_user).delete(delete_user),
        )
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c().await.expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    tracing::info!("signal received, starting graceful shutdown");
}

// Models
#[derive(Debug, Serialize, Deserialize, sqlx::FromRow)]
struct User {
    id: i64,
    name: String,
    email: String,
    created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Deserialize)]
struct CreateUserRequest {
    name: String,
    email: String,
}

#[derive(Debug, Deserialize)]
struct ListParams {
    limit: Option<i64>,
    offset: Option<i64>,
}

// Error handling
#[derive(Error, Debug)]
enum AppError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("not found: {0}")]
    NotFound(String),
    #[error("bad request: {0}")]
    BadRequest(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::NotFound(msg) => (StatusCode::NOT_FOUND, msg.clone()),
            AppError::BadRequest(msg) => (StatusCode::BAD_REQUEST, msg.clone()),
            AppError::Database(_) => {
                tracing::error!("Database error: {:?}", self);
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal error".into())
            }
        };

        (status, Json(serde_json::json!({"error": message}))).into_response()
    }
}

// Handlers
async fn health_check() -> Json<serde_json::Value> {
    Json(serde_json::json!({"status": "ok"}))
}

async fn list_users(
    State(state): State<AppState>,
    Query(params): Query<ListParams>,
) -> Result<Json<Vec<User>>, AppError> {
    let limit = params.limit.unwrap_or(10);
    let offset = params.offset.unwrap_or(0);

    let users = sqlx::query_as!(
        User,
        r#"
        SELECT id, name, email, created_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
        "#,
        limit,
        offset
    )
    .fetch_all(&state.db)
    .await?;

    Ok(Json(users))
}

async fn get_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<Json<User>, AppError> {
    let user = sqlx::query_as!(
        User,
        "SELECT id, name, email, created_at FROM users WHERE id = $1",
        id
    )
    .fetch_optional(&state.db)
    .await?
    .ok_or_else(|| AppError::NotFound(format!("User {} not found", id)))?;

    Ok(Json(user))
}

async fn create_user(
    State(state): State<AppState>,
    Json(req): Json<CreateUserRequest>,
) -> Result<(StatusCode, Json<User>), AppError> {
    if req.name.len() < 2 {
        return Err(AppError::BadRequest("Name must be at least 2 characters".into()));
    }

    let user = sqlx::query_as!(
        User,
        r#"
        INSERT INTO users (name, email)
        VALUES ($1, $2)
        RETURNING id, name, email, created_at
        "#,
        req.name,
        req.email
    )
    .fetch_one(&state.db)
    .await?;

    Ok((StatusCode::CREATED, Json(user)))
}

async fn update_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<CreateUserRequest>,
) -> Result<Json<User>, AppError> {
    let user = sqlx::query_as!(
        User,
        r#"
        UPDATE users
        SET name = $2, email = $3
        WHERE id = $1
        RETURNING id, name, email, created_at
        "#,
        id,
        req.name,
        req.email
    )
    .fetch_optional(&state.db)
    .await?
    .ok_or_else(|| AppError::NotFound(format!("User {} not found", id)))?;

    Ok(Json(user))
}

async fn delete_user(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> Result<StatusCode, AppError> {
    let result = sqlx::query!("DELETE FROM users WHERE id = $1", id)
        .execute(&state.db)
        .await?;

    if result.rows_affected() == 0 {
        return Err(AppError::NotFound(format!("User {} not found", id)));
    }

    Ok(StatusCode::NO_CONTENT)
}
```

---

## CLI Tool Examples

### Go CLI with Cobra

```go
// main.go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    verbose bool
)

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use:   "myctl",
    Short: "A CLI tool for managing resources",
    Long:  `myctl is a comprehensive CLI tool for managing cloud resources.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initConfig()
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.myctl.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(migrateCmd)
    rootCmd.AddCommand(userCmd)
}

// Serve command
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the API server",
    RunE: func(cmd *cobra.Command, args []string) error {
        port, _ := cmd.Flags().GetInt("port")
        fmt.Printf("Starting server on port %d...\n", port)
        return nil
    },
}

func init() {
    serveCmd.Flags().IntP("port", "p", 3000, "Port to listen on")
    viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
}

// Migrate command
var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations",
}

var migrateUpCmd = &cobra.Command{
    Use:   "up",
    Short: "Run all pending migrations",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("Running migrations...")
        return nil
    },
}

var migrateDownCmd = &cobra.Command{
    Use:   "down",
    Short: "Rollback last migration",
    RunE: func(cmd *cobra.Command, args []string) error {
        steps, _ := cmd.Flags().GetInt("steps")
        fmt.Printf("Rolling back %d migrations...\n", steps)
        return nil
    },
}

func init() {
    migrateDownCmd.Flags().IntP("steps", "n", 1, "Number of migrations to rollback")
    migrateCmd.AddCommand(migrateUpCmd)
    migrateCmd.AddCommand(migrateDownCmd)
}

// User command
var userCmd = &cobra.Command{
    Use:   "user",
    Short: "Manage users",
}

var userListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all users",
    RunE: func(cmd *cobra.Command, args []string) error {
        limit, _ := cmd.Flags().GetInt("limit")
        fmt.Printf("Listing %d users...\n", limit)
        return nil
    },
}

var userCreateCmd = &cobra.Command{
    Use:   "create [name] [email]",
    Short: "Create a new user",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        name, email := args[0], args[1]
        fmt.Printf("Creating user: %s <%s>\n", name, email)
        return nil
    },
}

func init() {
    userListCmd.Flags().IntP("limit", "l", 10, "Limit results")
    userCmd.AddCommand(userListCmd)
    userCmd.AddCommand(userCreateCmd)
}

func initConfig() error {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            return err
        }
        viper.AddConfigPath(home)
        viper.SetConfigName(".myctl")
    }

    viper.SetEnvPrefix("MYCTL")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err == nil {
        if verbose {
            fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
        }
    }

    return nil
}
```

### Rust CLI with clap

```rust
// src/main.rs
use clap::{Args, Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser)]
#[command(name = "myctl")]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, global = true)]
    config: Option<PathBuf>,

    #[arg(short, long, global = true)]
    verbose: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Start the API server
    Serve(ServeArgs),
    /// Run database migrations
    Migrate(MigrateArgs),
    /// Manage users
    User(UserArgs),
}

#[derive(Args)]
struct ServeArgs {
    #[arg(short, long, default_value = "3000")]
    port: u16,

    #[arg(long)]
    workers: Option<usize>,
}

#[derive(Args)]
struct MigrateArgs {
    #[command(subcommand)]
    command: MigrateCommands,
}

#[derive(Subcommand)]
enum MigrateCommands {
    /// Run all pending migrations
    Up,
    /// Rollback migrations
    Down {
        #[arg(short = 'n', long, default_value = "1")]
        steps: u32,
    },
}

#[derive(Args)]
struct UserArgs {
    #[command(subcommand)]
    command: UserCommands,
}

#[derive(Subcommand)]
enum UserCommands {
    /// List all users
    List {
        #[arg(short, long, default_value = "10")]
        limit: u32,
    },
    /// Create a new user
    Create {
        name: String,
        email: String,
    },
    /// Delete a user
    Delete {
        id: i64,
        #[arg(short, long)]
        force: bool,
    },
}

fn main() {
    let cli = Cli::parse();

    if cli.verbose {
        if let Some(config) = &cli.config {
            eprintln!("Using config: {:?}", config);
        }
    }

    match cli.command {
        Commands::Serve(args) => {
            println!("Starting server on port {}...", args.port);
            if let Some(workers) = args.workers {
                println!("Using {} workers", workers);
            }
        }
        Commands::Migrate(args) => match args.command {
            MigrateCommands::Up => {
                println!("Running migrations...");
            }
            MigrateCommands::Down { steps } => {
                println!("Rolling back {} migrations...", steps);
            }
        },
        Commands::User(args) => match args.command {
            UserCommands::List { limit } => {
                println!("Listing {} users...", limit);
            }
            UserCommands::Create { name, email } => {
                println!("Creating user: {} <{}>", name, email);
            }
            UserCommands::Delete { id, force } => {
                if force {
                    println!("Force deleting user {}...", id);
                } else {
                    println!("Deleting user {}...", id);
                }
            }
        },
    }
}
```

---

## Concurrency Examples

### Go Worker Pool

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "golang.org/x/sync/errgroup"
    "golang.org/x/sync/semaphore"
)

type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID int
    Data  string
    Error error
}

// Worker pool with fixed number of workers
func workerPool(ctx context.Context, jobs <-chan Job, numWorkers int) <-chan Result {
    results := make(chan Result, len(jobs))

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    result := processJob(job)
                    results <- result
                }
            }
        }(i)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}

func processJob(job Job) Result {
    time.Sleep(100 * time.Millisecond) // Simulate work
    return Result{JobID: job.ID, Data: fmt.Sprintf("Processed: %s", job.Data)}
}

// Rate-limited concurrent operations with semaphore
func rateLimitedOperations(ctx context.Context, items []string, maxConcurrent int64) error {
    sem := semaphore.NewWeighted(maxConcurrent)
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item
        g.Go(func() error {
            if err := sem.Acquire(ctx, 1); err != nil {
                return err
            }
            defer sem.Release(1)

            return processItem(ctx, item)
        })
    }

    return g.Wait()
}

func processItem(ctx context.Context, item string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(100 * time.Millisecond):
        fmt.Printf("Processed: %s\n", item)
        return nil
    }
}

// Fan-out/fan-in pattern
func fanOutFanIn(ctx context.Context, input <-chan int, workers int) <-chan int {
    // Fan-out: distribute work
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = worker(ctx, input)
    }

    // Fan-in: merge results
    return merge(ctx, channels...)
}

func worker(ctx context.Context, input <-chan int) <-chan int {
    output := make(chan int)
    go func() {
        defer close(output)
        for n := range input {
            select {
            case <-ctx.Done():
                return
            case output <- n * 2:
            }
        }
    }()
    return output
}

func merge(ctx context.Context, channels ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for n := range c {
                select {
                case <-ctx.Done():
                    return
                case output <- n:
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}
```

### Rust Async Concurrency

```rust
use std::time::Duration;
use tokio::sync::{mpsc, Semaphore};
use tokio::time::sleep;

#[derive(Debug)]
struct Job {
    id: u32,
    data: String,
}

#[derive(Debug)]
struct Result {
    job_id: u32,
    data: String,
}

// Worker pool with channels
async fn worker_pool(mut rx: mpsc::Receiver<Job>, num_workers: usize) -> Vec<Result> {
    let (result_tx, mut result_rx) = mpsc::channel::<Result>(100);

    // Spawn workers
    for _ in 0..num_workers {
        let result_tx = result_tx.clone();
        let job_rx = rx.clone();
        tokio::spawn(async move {
            while let Some(job) = job_rx.recv().await {
                let result = process_job(job).await;
                let _ = result_tx.send(result).await;
            }
        });
    }
    drop(result_tx);

    // Collect results
    let mut results = Vec::new();
    while let Some(result) = result_rx.recv().await {
        results.push(result);
    }
    results
}

async fn process_job(job: Job) -> Result {
    sleep(Duration::from_millis(100)).await;
    Result {
        job_id: job.id,
        data: format!("Processed: {}", job.data),
    }
}

// Rate-limited operations with semaphore
async fn rate_limited_operations(items: Vec<String>, max_concurrent: usize) -> Vec<String> {
    let semaphore = std::sync::Arc::new(Semaphore::new(max_concurrent));
    let mut handles = Vec::new();

    for item in items {
        let sem = semaphore.clone();
        handles.push(tokio::spawn(async move {
            let _permit = sem.acquire().await.unwrap();
            process_item(item).await
        }));
    }

    let mut results = Vec::new();
    for handle in handles {
        if let Ok(result) = handle.await {
            results.push(result);
        }
    }
    results
}

async fn process_item(item: String) -> String {
    sleep(Duration::from_millis(100)).await;
    format!("Processed: {}", item)
}

// Select for multiple futures
async fn timeout_or_result() -> Option<String> {
    tokio::select! {
        result = fetch_data() => Some(result),
        _ = sleep(Duration::from_secs(5)) => {
            eprintln!("Timeout");
            None
        }
    }
}

async fn fetch_data() -> String {
    sleep(Duration::from_secs(1)).await;
    "Data fetched".to_string()
}

// Stream processing
use tokio_stream::{self as stream, StreamExt};

async fn process_stream() {
    let mut stream = stream::iter(1..=10)
        .throttle(Duration::from_millis(100))
        .map(|n| n * 2);

    while let Some(value) = stream.next().await {
        println!("Value: {}", value);
    }
}
```

---

## Deployment Configurations

### Docker Compose

```yaml
# docker-compose.yml
version: "3.9"

services:
  go-api:
    build:
      context: ./go-service
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/myapp
      - REDIS_URL=redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: "0.5"
          memory: 256M

  rust-api:
    build:
      context: ./rust-service
      dockerfile: Dockerfile
    ports:
      - "3001:3000"
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/myapp
      - RUST_LOG=info
    depends_on:
      db:
        condition: service_healthy
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: "0.5"
          memory: 128M

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-api
  labels:
    app: go-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-api
  template:
    metadata:
      labels:
        app: go-api
    spec:
      containers:
        - name: go-api
          image: registry.example.com/go-api:latest
          ports:
            - containerPort: 3000
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: go-api
spec:
  selector:
    app: go-api
  ports:
    - port: 80
      targetPort: 3000
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-api
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: go-api
                port:
                  number: 80
```

---

## Testing Examples

### Go Integration Tests

```go
// integration_test.go
package main

import (
    "context"
    "encoding/json"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gofiber/fiber/v3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
    ctx := context.Background()

    container, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    pool, err := pgxpool.New(ctx, connStr)
    require.NoError(t, err)

    // Run migrations
    _, err = pool.Exec(ctx, `
        CREATE TABLE users (
            id BIGSERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
    require.NoError(t, err)

    cleanup := func() {
        pool.Close()
        container.Terminate(ctx)
    }

    return pool, cleanup
}

func TestUserAPI(t *testing.T) {
    pool, cleanup := setupTestDB(t)
    defer cleanup()

    app, _ := NewApp(Config{Port: "3000", DatabaseURL: ""})
    app.db = pool

    t.Run("create and get user", func(t *testing.T) {
        // Create user
        body := `{"name": "John Doe", "email": "john@example.com"}`
        req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")

        resp, err := app.fiber.Test(req)
        require.NoError(t, err)
        assert.Equal(t, 201, resp.StatusCode)

        var created User
        json.NewDecoder(resp.Body).Decode(&created)
        assert.Equal(t, "John Doe", created.Name)
        assert.NotZero(t, created.ID)

        // Get user
        req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", created.ID), nil)
        resp, err = app.fiber.Test(req)
        require.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)

        var fetched User
        json.NewDecoder(resp.Body).Decode(&fetched)
        assert.Equal(t, created.ID, fetched.ID)
    })

    t.Run("get non-existent user", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/users/99999", nil)
        resp, err := app.fiber.Test(req)
        require.NoError(t, err)
        assert.Equal(t, 404, resp.StatusCode)
    })
}
```

### Rust Integration Tests

```rust
// tests/integration_test.rs
use axum::{body::Body, http::Request};
use sqlx::PgPool;
use tower::ServiceExt;

async fn setup_test_db() -> PgPool {
    let database_url = std::env::var("TEST_DATABASE_URL")
        .unwrap_or_else(|_| "postgres://test:test@localhost/test".into());

    let pool = PgPool::connect(&database_url).await.unwrap();

    // Run migrations
    sqlx::query(
        r#"
        CREATE TABLE IF NOT EXISTS users (
            id BIGSERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )
        "#,
    )
    .execute(&pool)
    .await
    .unwrap();

    // Clean up
    sqlx::query("DELETE FROM users").execute(&pool).await.unwrap();

    pool
}

#[tokio::test]
async fn test_create_and_get_user() {
    let pool = setup_test_db().await;
    let app = create_app(pool.clone());

    // Create user
    let response = app
        .clone()
        .oneshot(
            Request::builder()
                .method("POST")
                .uri("/api/v1/users")
                .header("Content-Type", "application/json")
                .body(Body::from(r#"{"name": "John Doe", "email": "john@example.com"}"#))
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::CREATED);

    let body = hyper::body::to_bytes(response.into_body()).await.unwrap();
    let created: User = serde_json::from_slice(&body).unwrap();
    assert_eq!(created.name, "John Doe");

    // Get user
    let response = app
        .oneshot(
            Request::builder()
                .uri(format!("/api/v1/users/{}", created.id))
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::OK);

    let body = hyper::body::to_bytes(response.into_body()).await.unwrap();
    let fetched: User = serde_json::from_slice(&body).unwrap();
    assert_eq!(fetched.id, created.id);
}

#[tokio::test]
async fn test_get_nonexistent_user() {
    let pool = setup_test_db().await;
    let app = create_app(pool);

    let response = app
        .oneshot(
            Request::builder()
                .uri("/api/v1/users/99999")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(response.status(), StatusCode::NOT_FOUND);
}
```

---

Last Updated: 2025-12-07
Version: 1.0.0
