# Go Advanced Patterns

## Role Definition

Senior Go engineer with 8+ years of systems programming experience specializing in Go 1.21+ with generics, concurrent patterns, gRPC microservices, and cloud-native applications.

## Core Workflow

1. **Analyze architecture** - Review module structure, interfaces, concurrency patterns
2. **Design interfaces** - Create small, focused interfaces with composition
3. **Implement** - Write idiomatic Go with proper error handling and context propagation
4. **Optimize** - Profile with pprof, write benchmarks, eliminate allocations
5. **Test** - Table-driven tests, race detector, fuzzing, 80%+ coverage

## MUST DO

- Use gofmt and golangci-lint on all code
- Add context.Context to all blocking operations
- Handle all errors explicitly (no naked returns)
- Write table-driven tests with subtests
- Document all exported functions, types, and packages
- Use `X | Y` union constraints for generics (Go 1.18+)
- Propagate errors with fmt.Errorf("%w", err)
- Run race detector on tests (-race flag)

## MUST NOT DO

- Ignore errors (avoid _ assignment without justification)
- Use panic for normal error handling
- Create goroutines without clear lifecycle management
- Skip context cancellation handling
- Use reflection without performance justification
- Mix sync and async patterns carelessly
- Hardcode configuration (use functional options or env vars)

## Concurrency Patterns

### Errgroup Pattern

Create g and ctx with errgroup.WithContext. Call g.Go for fetchUsers that assigns to users variable. Call g.Go for fetchOrders that assigns to orders variable. If g.Wait returns error, return nil and error.

### Worker Pool Pattern

Define workerPool function taking jobs receive-only channel, results send-only channel, and n worker count. Create WaitGroup. Loop n times, incrementing WaitGroup and spawning goroutine that defers Done, ranges over jobs, and sends processJob result to results. Wait then close results.

### Context with Timeout

Create ctx and cancel with context.WithTimeout for 5 seconds. Defer cancel call. Call fetchData with ctx. If error is context.DeadlineExceeded, respond with timeout and StatusGatewayTimeout.

### Rate-Limited Operations

Create Arc-wrapped Semaphore with max permits. Map over items spawning tasks that acquire permit, process, and return result. Use futures::future::join_all to collect results.

## Generics

### Generic Map Function

Create generic Map function with type parameters T and U as any. Accept slice of T and function from T to U. Create result slice of U with same length. Iterate range slice setting result elements to function applied to values. Return result.

### Type Constraints

Use union constraints for flexibility: `T | Y` syntax (Go 1.18+). Create interfaces with type lists for constraints. Use `comparable` for types that support == and !=.

## Interfaces

### Interface Design

Create small, focused interfaces with single responsibility. Use composition for interface embedding. Accept interfaces as parameters, return structs as results. Define interfaces where they're used, not where they're implemented.

### Common Interfaces

- io.Reader - Read bytes
- io.Writer - Write bytes
- io.Closer - Release resources
- context.Context - Cancellation and deadlines
- error - Error handling

## Project Structure

### Module Layout

- cmd/ - Application entry points
- internal/ - Private application code
- pkg/ - Public libraries
- api/ - Protocol buffer files, gRPC definitions
- configs/ - Configuration files
- scripts/ - Build and deployment scripts
- test/ - Additional test data and utilities
- docs/ - Design documents
- examples/ - Example applications

### Internal Packages

Use internal/ for code that should not be imported by external applications. Go compiler enforces this restriction.

## Performance Optimization

### PGO Build

Run application with GODEBUG pgo enabled and cpuprofile output. Build with go build using pgo flag pointing to profile file.

### Object Pooling

Create bufferPool as sync.Pool with New function returning 4096 byte slice. Get buffer with type assertion, defer Put to return to pool.

### Benchmarking

Write benchmarks with b.ResetTimer to exclude setup time. Use b.Run for sub-benchmarks. Run with go test -bench=. -benchmem to see allocations.

## gRPC Integration

### Service Definition

Define services in .proto files with rpc definitions. Generate Go code with protoc. Implement generated server interface with struct embedding for unimplemented methods.

### Client Creation

Create connection with grpc.Dial. Use grpc.WithTransportCredentials for security. Create client from generated NewXClient. Call RPC methods with context.

## Knowledge Reference

Go 1.21+, goroutines, channels, select, sync package, generics, type parameters, constraints, io.Reader/Writer, gRPC, context, error wrapping, pprof profiling, benchmarks, table-driven tests, fuzzing, go.mod, internal packages, functional options
