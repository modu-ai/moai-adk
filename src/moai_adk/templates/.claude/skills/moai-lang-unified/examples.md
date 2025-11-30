# Multi-Language Code Examples

## Quick Start Examples

### REST API Implementation by Language

#### Python (FastAPI)
```python
# main.py
from fastapi import FastAPI, HTTPException, Depends
from pydantic import BaseModel
from typing import List, Optional
import asyncpg

app = FastAPI(title="User API", version="1.0.0")

# Models
class User(BaseModel):
 id: int
 name: str
 email: str
 created_at: Optional[str] = None

class CreateUserRequest(BaseModel):
 name: str
 email: str

# Database
DATABASE_URL = "postgresql+asyncpg://user:pass@localhost/db"

async def get_db():
 return await asyncpg.connect(DATABASE_URL)

# Routes
@app.get("/users", response_model=List[User])
async def list_users():
 conn = await get_db()
 rows = await conn.fetch("SELECT id, name, email, created_at FROM users")
 await conn.close()
 return [User(dict(row)) for row in rows]

@app.get("/users/{user_id}", response_model=User)
async def get_user(user_id: int):
 conn = await get_db()
 row = await conn.fetchrow("SELECT id, name, email, created_at FROM users WHERE id = $1", user_id)
 await conn.close()
 if not row:
 raise HTTPException(status_code=404, detail="User not found")
 return User(dict(row))

@app.post("/users", response_model=User, status_code=201)
async def create_user(user: CreateUserRequest):
 conn = await get_db()
 row = await conn.fetchrow(
 "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
 user.name, user.email
 )
 await conn.close()
 return User(dict(row))

# Run: uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

#### TypeScript (Express.js)
```typescript
// src/app.ts
import express, { Request, Response } from 'express';
import { Pool } from 'pg';
import { z } from 'zod';

const app = express();
app.use(express.json());

// Database
const pool = new Pool({
 connectionString: 'postgresql://user:pass@localhost/db'
});

// Schemas
const CreateUserSchema = z.object({
 name: z.string().min(2),
 email: z.string().email()
});

const UserSchema = z.object({
 id: z.number(),
 name: z.string(),
 email: z.string(),
 created_at: z.string().nullable()
});

type User = z.infer<typeof UserSchema>;
type CreateUserRequest = z.infer<typeof CreateUserSchema>;

// Routes
app.get('/users', async (req: Request, res: Response) => {
 const result = await pool.query('SELECT id, name, email, created_at FROM users');
 const users = result.rows.map(row => UserSchema.parse(row));
 res.json(users);
});

app.get('/users/:id', async (req: Request, res: Response) => {
 const { id } = req.params;
 const result = await pool.query(
 'SELECT id, name, email, created_at FROM users WHERE id = $1',
 [parseInt(id)]
 );
 
 if (result.rows.length === 0) {
 return res.status(404).json({ error: 'User not found' });
 }
 
 const user = UserSchema.parse(result.rows[0]);
 res.json(user);
});

app.post('/users', async (req: Request, res: Response) => {
 try {
 const createUser = CreateUserSchema.parse(req.body);
 const result = await pool.query(
 'INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at',
 [createUser.name, createUser.email]
 );
 
 const user = UserSchema.parse(result.rows[0]);
 res.status(201).json(user);
 } catch (error) {
 if (error instanceof z.ZodError) {
 res.status(400).json({ error: error.errors });
 } else {
 res.status(500).json({ error: 'Internal server error' });
 }
 }
});

const PORT = 3000;
app.listen(PORT, () => {
 console.log(`Server running on port ${PORT}`);
});

// Run: npm install express pg zod && npm run dev
```

#### Go (Fiber)
```go
// main.go
package main

import (
 "context"
 "database/sql"
 "encoding/json"
 "fmt"
 "log"
 "net/http"
 "time"

 "github.com/gofiber/fiber/v3"
 "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
 ID int `json:"id"`
 Name string `json:"name"`
 Email string `json:"email"`
 CreatedAt time.Time `json:"created_at,omitempty"`
}

type CreateUserRequest struct {
 Name string `json:"name"`
 Email string `json:"email"`
}

var pool *pgxpool.Pool

func main() {
 // Database
 var err error
 pool, err = pgxpool.New(context.Background(), "postgres://user:pass@localhost/db")
 if err != nil {
 log.Fatal(err)
 }
 defer pool.Close()

 // Fiber app
 app := fiber.New()

 // Routes
 app.Get("/users", listUsers)
 app.Get("/users/:id", getUser)
 app.Post("/users", createUser)

 log.Fatal(app.Listen(":3000"))
}

func listUsers(c fiber.Ctx) error {
 rows, err := pool.Query(context.Background(), "SELECT id, name, email, created_at FROM users")
 if err != nil {
 return c.Status(500).JSON(fiber.Map{"error": err.Error()})
 }
 defer rows.Close()

 var users []User
 for rows.Next() {
 var user User
 err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
 if err != nil {
 return c.Status(500).JSON(fiber.Map{"error": err.Error()})
 }
 users = append(users, user)
 }

 return c.JSON(users)
}

func getUser(c fiber.Ctx) error {
 id, err := c.ParamsInt("id")
 if err != nil {
 return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
 }

 var user User
 err = pool.QueryRow(context.Background(),
 "SELECT id, name, email, created_at FROM users WHERE id = $1", id).
 Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
 
 if err == sql.ErrNoRows {
 return c.Status(404).JSON(fiber.Map{"error": "User not found"})
 }
 if err != nil {
 return c.Status(500).JSON(fiber.Map{"error": err.Error()})
 }

 return c.JSON(user)
}

func createUser(c fiber.Ctx) error {
 var req CreateUserRequest
 if err := c.BodyParser(&req); err != nil {
 return c.Status(400).JSON(fiber.Map{"error": err.Error()})
 }

 var user User
 err := pool.QueryRow(context.Background(),
 "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
 req.Name, req.Email).
 Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
 
 if err != nil {
 return c.Status(500).JSON(fiber.Map{"error": err.Error()})
 }

 return c.Status(201).JSON(user)
}

// Run: go mod init user-api && go get github.com/gofiber/fiber/v3 github.com/jackc/pgx/v5 && go run main.go
```

#### Rust (Axum)
```rust
// src/main.rs
use axum::{
 extract::{Path, State},
 http::StatusCode,
 response::Json,
 routing::get,
 Router,
};
use serde::{Deserialize, Serialize};
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::net::SocketAddr;

#[derive(Debug, Serialize, Deserialize)]
struct User {
 id: i32,
 name: String,
 email: String,
 created_at: Option<String>,
}

#[derive(Debug, Deserialize)]
struct CreateUserRequest {
 name: String,
 email: String,
}

#[derive(Clone)]
struct AppState {
 db: PgPool,
}

#[tokio::main]
async fn main() {
 // Database
 let pool = PgPoolOptions::new()
 .max_connections(5)
 .connect("postgres://user:pass@localhost/db")
 .await
 .expect("Failed to connect to database");

 let state = AppState { db: pool };

 // Routes
 let app = Router::new()
 .route("/users", get(list_users).post(create_user))
 .route("/users/:id", get(get_user))
 .with_state(state);

 // Server
 let addr = SocketAddr::from(([0, 0, 0, 0], 3000));
 println!("Server listening on {}", addr);

 axum::Server::bind(&addr)
 .serve(app.into_make_service())
 .await
 .unwrap();
}

async fn list_users(State(state): State<AppState>) -> Json<Vec<User>> {
 let users = sqlx::query_as!(
 User,
 "SELECT id, name, email, TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at FROM users"
 )
 .fetch_all(&state.db)
 .await
 .expect("Failed to fetch users");

 Json(users)
}

async fn get_user(
 State(state): State<AppState>,
 Path(user_id): Path<i32>,
) -> Result<Json<User>, StatusCode> {
 let user = sqlx::query_as!(
 User,
 "SELECT id, name, email, TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at FROM users WHERE id = $1",
 user_id
 )
 .fetch_optional(&state.db)
 .await
 .expect("Failed to fetch user");

 match user {
 Some(user) => Ok(Json(user)),
 None => Err(StatusCode::NOT_FOUND),
 }
}

async fn create_user(
 State(state): State<AppState>,
 Json(payload): Json<CreateUserRequest>,
) -> Result<Json<User>, StatusCode> {
 let user = sqlx::query_as!(
 User,
 r#"
 INSERT INTO users (name, email) 
 VALUES ($1, $2) 
 RETURNING id, name, email, TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at
 "#,
 payload.name,
 payload.email
 )
 .fetch_one(&state.db)
 .await
 .expect("Failed to create user");

 Ok(Json(user))
}

// Cargo.toml dependencies:
// axum = "0.8"
// sqlx = { version = "0.8", features = ["runtime-tokio-rustls", "postgres", "chrono"] }
// serde = { version = "1.0", features = ["derive"] }
// tokio = { version = "1.0", features = ["full"] }
```

## Cross-Language Communication

### gRPC Service Implementation

#### Go Server
```go
// server.go
package main

import (
 "context"
 "log"
 "net"

 pb "github.com/yourorg/user-service/proto"
 "google.golang.org/grpc"
)

type server struct {
 pb.UnimplementedUserServiceServer
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
 // Business logic
 return &pb.UserResponse{
 Id: req.Id,
 Name: "John Doe",
 Email: "john@example.com",
 }, nil
}

func main() {
 lis, err := net.Listen("tcp", ":50051")
 if err != nil {
 log.Fatal(err)
 }

 s := grpc.NewServer()
 pb.RegisterUserServiceServer(s, &server{})

 log.Println("gRPC server listening on :50051")
 if err := s.Serve(lis); err != nil {
 log.Fatal(err)
 }
}
```

#### Python Client
```python
# client.py
import grpc
from generated import user_pb2, user_pb2_grpc

def get_user(user_id: int):
 with grpc.insecure_channel('localhost:50051') as channel:
 stub = user_pb2_grpc.UserServiceStub(channel)
 request = user_pb2.GetUserRequest(id=user_id)
 response = stub.GetUser(request)
 return response

# Usage
user = get_user(123)
print(f"User: {user.name}, Email: {user.email}")
```

#### Rust Client
```rust
// client.rs
use tonic::transport::Channel;
use proto::user_service_client::UserServiceClient;
use proto::GetUserRequest;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
 let mut client = UserServiceClient::connect("http://[::1]:50051").await?;

 let request = tonic::Request::new(GetUserRequest {
 id: 123,
 });

 let response = client.get_user(request).await?;
 println!("Response: {:?}", response.get_ref());

 Ok(())
}
```

## Data Processing Examples

### JSON Processing

#### Python (FastAPI)
```python
import json
from typing import List, Dict, Any
from pathlib import Path

def process_json_data(file_path: str) -> List[Dict[str, Any]]:
 """Process JSON file and return structured data"""
 with open(file_path, 'r') as f:
 data = json.load(f)
 
 # Transform data
 processed = []
 for item in data:
 processed.append({
 'id': item['id'],
 'name': item['name'].upper(),
 'email': item['email'].lower(),
 'processed_at': datetime.now().isoformat()
 })
 
 return processed

# Async batch processing
async def process_large_json_async(file_path: str) -> List[Dict]:
 with open(file_path, 'r') as f:
 data = json.load(f)
 
 # Process in batches
 batch_size = 1000
 results = []
 
 for i in range(0, len(data), batch_size):
 batch = data[i:i + batch_size]
 batch_results = await process_batch(batch)
 results.extend(batch_results)
 
 return results
```

#### JavaScript (Node.js)
```javascript
// jsonProcessor.js
const fs = require('fs').promises;
const { pipeline } = require('stream/promises');
const { createReadStream, createWriteStream } = require('fs');

async function processJsonFile(inputPath, outputPath) {
 const readStream = createReadStream(inputPath, { encoding: 'utf8' });
 const writeStream = createWriteStream(outputPath, { encoding: 'utf8' });
 
 let buffer = '';
 let isFirstLine = true;
 
 // Process line by line for memory efficiency
 for await (const chunk of readStream) {
 buffer += chunk;
 let lineEnd = buffer.indexOf('\n');
 
 while (lineEnd >= 0) {
 const line = buffer.slice(0, lineEnd);
 buffer = buffer.slice(lineEnd + 1);
 lineEnd = buffer.indexOf('\n');
 
 try {
 const data = JSON.parse(line);
 const processed = {
 id: data.id,
 name: data.name.toUpperCase(),
 email: data.email.toLowerCase(),
 processedAt: new Date().toISOString()
 };
 
 if (!isFirstLine) {
 writeStream.write(',\n');
 }
 isFirstLine = false;
 writeStream.write(JSON.stringify(processed));
 } catch (error) {
 console.error('Error processing line:', error.message);
 }
 }
 }
 
 writeStream.end();
}

// Usage
processJsonFile('input.json', 'output.json')
 .then(() => console.log('Processing complete'))
 .catch(console.error);
```

#### Go
```go
// jsonProcessor.go
package main

import (
 "encoding/json"
 "fmt"
 "io"
 "os"
 "strings"
)

type InputData struct {
 ID int `json:"id"`
 Name string `json:"name"`
 Email string `json:"email"`
}

type ProcessedData struct {
 ID int `json:"id"`
 Name string `json:"name"`
 Email string `json:"email"`
 ProcessedAt string `json:"processed_at"`
}

func processJSONFile(inputPath, outputPath string) error {
 file, err := os.Open(inputPath)
 if err != nil {
 return err
 }
 defer file.Close()

 outputFile, err := os.Create(outputPath)
 if err != nil {
 return err
 }
 defer outputFile.Close()

 decoder := json.NewDecoder(file)
 
 outputFile.WriteString("[\n")
 isFirst := true
 
 for decoder.More() {
 var data InputData
 if err := decoder.Decode(&data); err != nil {
 return err
 }
 
 processed := ProcessedData{
 ID: data.ID,
 Name: strings.ToUpper(data.Name),
 Email: strings.ToLower(data.Email),
 ProcessedAt: time.Now().Format(time.RFC3339),
 }
 
 if !isFirst {
 outputFile.WriteString(",\n")
 }
 isFirst = false
 
 jsonData, err := json.MarshalIndent(processed, "", " ")
 if err != nil {
 return err
 }
 outputFile.Write(jsonData)
 }
 
 outputFile.WriteString("\n]")
 return nil
}

func main() {
 err := processJSONFile("input.json", "output.json")
 if err != nil {
 fmt.Printf("Error: %v\n", err)
 os.Exit(1)
 }
 fmt.Println("Processing complete")
}
```

## Testing Examples

### Unit Testing

#### Python (pytest)
```python
# test_user_service.py
import pytest
from unittest.mock import Mock, patch
from user_service import UserService, UserNotFoundError

@pytest.fixture
def user_service():
 return UserService()

@pytest.fixture
def mock_db():
 return Mock()

def test_create_user_success(user_service, mock_db):
 # Setup
 user_data = {"name": "John Doe", "email": "john@example.com"}
 mock_db.insert.return_value = {"id": 1, user_data}
 
 # Test
 with patch.object(user_service, 'db', mock_db):
 result = user_service.create_user(user_data)
 
 # Assert
 assert result["id"] == 1
 assert result["name"] == "John Doe"
 mock_db.insert.assert_called_once_with(user_data)

def test_get_user_not_found(user_service, mock_db):
 # Setup
 user_id = 999
 mock_db.get.return_value = None
 
 # Test
 with patch.object(user_service, 'db', mock_db):
 with pytest.raises(UserNotFoundError):
 user_service.get_user(user_id)
 
 # Assert
 mock_db.get.assert_called_once_with(user_id)

@pytest.mark.asyncio
async def test_async_user_operations(user_service, mock_db):
 # Setup
 user_data = {"name": "Jane Doe", "email": "jane@example.com"}
 mock_db.async_insert.return_value = {"id": 2, user_data}
 
 # Test
 with patch.object(user_service, 'db', mock_db):
 result = await user_service.create_user_async(user_data)
 
 # Assert
 assert result["id"] == 2
 assert result["name"] == "Jane Doe"
```

#### JavaScript (Vitest)
```javascript
// userService.test.js
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { UserService, UserNotFoundError } from './userService.js';

describe('UserService', () => {
 let userService;
 let mockDb;

 beforeEach(() => {
 mockDb = {
 insert: vi.fn(),
 get: vi.fn(),
 update: vi.fn(),
 delete: vi.fn()
 };
 userService = new UserService(mockDb);
 });

 it('should create user successfully', async () => {
 // Setup
 const userData = { name: 'John Doe', email: 'john@example.com' };
 const expectedUser = { id: 1, ...userData };
 mockDb.insert.mockResolvedValue(expectedUser);

 // Test
 const result = await userService.createUser(userData);

 // Assert
 expect(result).toEqual(expectedUser);
 expect(mockDb.insert).toHaveBeenCalledWith(userData);
 expect(mockDb.insert).toHaveBeenCalledTimes(1);
 });

 it('should throw UserNotFoundError when user not found', async () => {
 // Setup
 const userId = 999;
 mockDb.get.mockResolvedValue(null);

 // Test & Assert
 await expect(userService.getUser(userId))
 .rejects
 .toThrow(UserNotFoundError);
 
 expect(mockDb.get).toHaveBeenCalledWith(userId);
 });

 it('should handle database errors gracefully', async () => {
 // Setup
 const userData = { name: 'Jane Doe', email: 'jane@example.com' };
 mockDb.insert.mockRejectedValue(new Error('Database connection failed'));

 // Test & Assert
 await expect(userService.createUser(userData))
 .rejects
 .toThrow('Database connection failed');
 });
});
```

#### Go
```go
// user_service_test.go
package main

import (
 "errors"
 "testing"
 "github.com/stretchr/testify/assert"
 "github.com/stretchr/testify/mock"
)

type MockDB struct {
 mock.Mock
}

func (m *MockDB) Insert(user map[string]interface{}) (map[string]interface{}, error) {
 args := m.Called(user)
 return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDB) Get(id int) (map[string]interface{}, error) {
 args := m.Called(id)
 return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestCreateUser_Success(t *testing.T) {
 // Setup
 mockDB := new(MockDB)
 userService := NewUserService(mockDB)
 
 userData := map[string]interface{}{
 "name": "John Doe",
 "email": "john@example.com",
 }
 expectedUser := map[string]interface{}{
 "id": 1,
 "name": "John Doe",
 "email": "john@example.com",
 }
 
 mockDB.On("Insert", userData).Return(expectedUser, nil)
 
 // Test
 result, err := userService.CreateUser(userData)
 
 // Assert
 assert.NoError(t, err)
 assert.Equal(t, expectedUser, result)
 mockDB.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
 // Setup
 mockDB := new(MockDB)
 userService := NewUserService(mockDB)
 userID := 999
 
 mockDB.On("Get", userID).Return(nil, errors.New("user not found"))
 
 // Test
 _, err := userService.GetUser(userID)
 
 // Assert
 assert.Error(t, err)
 assert.Contains(t, err.Error(), "user not found")
 mockDB.AssertExpectations(t)
}

func BenchmarkCreateUser(b *testing.B) {
 // Setup
 mockDB := new(MockDB)
 userService := NewUserService(mockDB)
 userData := map[string]interface{}{
 "name": "Benchmark User",
 "email": "benchmark@example.com",
 }
 expectedUser := map[string]interface{}{
 "id": 1,
 "name": "Benchmark User",
 "email": "benchmark@example.com",
 }
 
 mockDB.On("Insert", userData).Return(expectedUser, nil)
 
 // Benchmark
 b.ResetTimer()
 for i := 0; i < b.N; i++ {
 _, _ = userService.CreateUser(userData)
 }
}
```

#### Rust
```rust
// src/user_service_test.rs
#[cfg(test)]
mod tests {
 use super::*;
 use tokio_test;
 
 #[derive(Debug)]
 struct MockDB {
 users: std::cell::RefCell<HashMap<i32, User>>,
 next_id: std::cell::Cell<i32>,
 }
 
 impl MockDB {
 fn new() -> Self {
 Self {
 users: std::cell::RefCell::new(HashMap::new()),
 next_id: std::cell::Cell::new(1),
 }
 }
 
 async fn create_user(&self, user: CreateUserRequest) -> User {
 let id = self.next_id.get();
 self.next_id.set(id + 1);
 
 let user = User {
 id,
 name: user.name,
 email: user.email,
 created_at: Some(chrono::Utc::now().to_string()),
 };
 
 self.users.borrow_mut().insert(id, user.clone());
 user
 }
 
 async fn get_user(&self, id: i32) -> Option<User> {
 self.users.borrow().get(&id).cloned()
 }
 }
 
 #[tokio::test]
 async fn test_create_user_success() {
 // Setup
 let mock_db = MockDB::new();
 let user_request = CreateUserRequest {
 name: "John Doe".to_string(),
 email: "john@example.com".to_string(),
 };
 
 // Test
 let result = mock_db.create_user(user_request.clone()).await;
 
 // Assert
 assert_eq!(result.name, "John Doe");
 assert_eq!(result.email, "john@example.com");
 assert_eq!(result.id, 1);
 assert!(result.created_at.is_some());
 }
 
 #[tokio::test]
 async fn test_get_user_not_found() {
 // Setup
 let mock_db = MockDB::new();
 let non_existent_id = 999;
 
 // Test
 let result = mock_db.get_user(non_existent_id).await;
 
 // Assert
 assert!(result.is_none());
 }
 
 #[tokio::test]
 async fn test_get_user_success() {
 // Setup
 let mock_db = MockDB::new();
 let user_request = CreateUserRequest {
 name: "Jane Doe".to_string(),
 email: "jane@example.com".to_string(),
 };
 
 let created_user = mock_db.create_user(user_request.clone()).await;
 
 // Test
 let result = mock_db.get_user(created_user.id).await;
 
 // Assert
 assert!(result.is_some());
 let user = result.unwrap();
 assert_eq!(user.name, "Jane Doe");
 assert_eq!(user.email, "jane@example.com");
 }
}
```

## Deployment Examples

### Docker Multi-Stage Builds

#### Python Dockerfile
```dockerfile
# Dockerfile.python
FROM python:3.13-slim as builder

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

FROM python:3.13-slim as runtime

WORKDIR /app
COPY --from=builder /root/.local /root/.local
COPY . .

ENV PATH=/root/.local/bin:$PATH

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

#### Go Dockerfile
```dockerfile
# Dockerfile.go
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 3000
CMD ["./main"]
```

#### Rust Dockerfile
```dockerfile
# Dockerfile.rust
FROM rust:1.91-alpine AS builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

COPY . .
RUN touch src/main.rs && cargo build --release

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/target/release/service .
EXPOSE 3000
CMD ["./service"]
```

### Kubernetes Deployment

#### Python Deployment
```yaml
# k8s-python.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
 name: python-api
spec:
 replicas: 3
 selector:
 matchLabels:
 app: python-api
 template:
 metadata:
 labels:
 app: python-api
 spec:
 containers:
 - name: python-api
 image: your-registry/python-api:latest
 ports:
 - containerPort: 8000
 env:
 - name: DATABASE_URL
 valueFrom:
 secretKeyRef:
 name: db-secret
 key: url
 resources:
 requests:
 memory: "256Mi"
 cpu: "250m"
 limits:
 memory: "512Mi"
 cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
 name: python-api-service
spec:
 selector:
 app: python-api
 ports:
 - port: 80
 targetPort: 8000
 type: LoadBalancer
```

#### Go Deployment
```yaml
# k8s-go.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
 name: go-api
spec:
 replicas: 2
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
 image: your-registry/go-api:latest
 ports:
 - containerPort: 3000
 env:
 - name: DB_HOST
 value: postgres-service
 - name: DB_PORT
 value: "5432"
 resources:
 requests:
 memory: "64Mi"
 cpu: "100m"
 limits:
 memory: "128Mi"
 cpu: "200m"
 livenessProbe:
 httpGet:
 path: /health
 port: 3000
 initialDelaySeconds: 10
 periodSeconds: 30
---
apiVersion: v1
kind: Service
metadata:
 name: go-api-service
spec:
 selector:
 app: go-api
 ports:
 - port: 80
 targetPort: 3000
 type: LoadBalancer
```

## Microservices Communication

### Message Queue Patterns

#### RabbitMQ Producer (Python)
```python
# producer.py
import pika
import json
import asyncio

class MessageProducer:
 def __init__(self, connection_params):
 self.connection = pika.BlockingConnection(connection_params)
 self.channel = self.connection.channel()
 
 # Declare queue
 self.channel.queue_declare(queue='user_events', durable=True)
 
 def publish_user_event(self, event_type, user_data):
 message = {
 'event_type': event_type,
 'user': user_data,
 'timestamp': datetime.now().isoformat()
 }
 
 self.channel.basic_publish(
 exchange='',
 routing_key='user_events',
 body=json.dumps(message),
 properties=pika.BasicProperties(
 delivery_mode=2, # make message persistent
 ))
 
 print(f" [x] Sent {event_type} event")
 
 def close(self):
 self.connection.close()

# Usage
producer = MessageProducer(pika.ConnectionParameters('localhost'))
producer.publish_user_event('user_created', {'id': 123, 'name': 'John Doe'})
producer.close()
```

#### RabbitMQ Consumer (Go)
```go
// consumer.go
package main

import (
 "encoding/json"
 "log"
 "github.com/streadway/amqp"
)

type UserEvent struct {
 EventType string `json:"event_type"`
 User interface{} `json:"user"`
 Timestamp string `json:"timestamp"`
}

func main() {
 conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
 if err != nil {
 log.Fatal(err)
 }
 defer conn.Close()

 ch, err := conn.Channel()
 if err != nil {
 log.Fatal(err)
 }
 defer ch.Close()

 q, err := ch.QueueDeclare(
 "user_events", // name
 true, // durable
 false, // delete when unused
 false, // exclusive
 false, // no-wait
 nil, // arguments
 )
 if err != nil {
 log.Fatal(err)
 }

 msgs, err := ch.Consume(
 q.Name, // queue
 "", // consumer
 true, // auto-ack
 false, // exclusive
 false, // no-local
 false, // no-wait
 nil, // args
 )
 if err != nil {
 log.Fatal(err)
 }

 log.Println("Waiting for messages. To exit press CTRL+C")
 
 forever := make(chan bool)
 
 go func() {
 for d := range msgs {
 log.Printf("Received a message: %s", d.Body)
 
 var event UserEvent
 if err := json.Unmarshal(d.Body, &event); err != nil {
 log.Printf("Error parsing message: %v", err)
 continue
 }
 
 log.Printf("Event type: %s, User: %v", event.EventType, event.User)
 
 // Process the event
 handleUserEvent(event)
 }
 }()
 
 <-forever
}

func handleUserEvent(event UserEvent) {
 switch event.EventType {
 case "user_created":
 log.Println("Handling user creation event")
 case "user_updated":
 log.Println("Handling user update event")
 case "user_deleted":
 log.Println("Handling user deletion event")
 default:
 log.Printf("Unknown event type: %s", event.EventType)
 }
}
```

---

Last Updated: 2025-11-25
Version: 1.0.0
