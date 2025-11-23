---
name: moai-lang-rust
description: Rust 1.91+ Systems Programming with Tokio 1.48, ownership, async/await, and safety guarantees
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: rust, lang, moai  


## Quick Reference (30 seconds)

# Rust Systems Programming — Enterprise

**Primary Focus**: Rust 1.91+ with ownership model, zero-cost abstractions, async/await patterns
**Best For**: Systems programming, performance-critical code, concurrent applications, safety-first development
**Key Libraries**: Tokio 1.48, Axum 0.8, async-trait 0.1, parking_lot 0.12
**Auto-triggers**: Rust, .rs files, Tokio, Axum, async/await, systems code

| Version | Release | Support |
|---------|---------|---------|
| Rust 1.91.1 | 2025-11 | Active |
| Tokio 1.48 | 2025-11 | ✅ |
| Axum 0.8 | 2025-11 | ✅ |
| Cargo | Latest | Built-in |

---

## What It Does

Rust 1.91+ systems programming with memory safety, ownership rules, and async/await. Zero-cost abstractions, fearless concurrency, and compile-time guarantees for production systems.

**Key capabilities**:
- ✅ Ownership model and borrow checker
- ✅ Zero-cost abstractions with traits and generics
- ✅ Async/await with Tokio runtime
- ✅ Macro system and metaprogramming
- ✅ Error handling with Result and Option
- ✅ Memory safety without garbage collection
- ✅ High-performance concurrent applications

---

## When to Use

**Automatic triggers**:
- Rust source files (*.rs, *.toml)
- Systems programming tasks
- Performance-critical applications
- Concurrent/parallel processing

**Manual invocation**:
- Design memory-safe architecture
- Implement async patterns with Tokio
- Optimize performance bottlenecks
- Review Rust code for safety issues

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core Rust concepts with practical patterns:
- **Ownership & Borrowing**: Move semantics, references, lifetimes
- **Pattern Matching**: Destructuring, exhaustive matching, option/result
- **Error Handling**: Result wrapper, ? operator, custom errors
- **Basic Async**: async/await, futures, task spawning
- **Testing**: Unit tests, integration tests, test modules

### Level 2: Advanced Patterns (See modules/advanced-patterns.md)

Production-ready enterprise patterns:
- **Traits & Generics**: Polymorphism, trait bounds, associated types
- **Async Patterns**: Tokio runtime, channels, select macros
- **Macro System**: Derive macros, procedural macros, custom DSLs
- **Concurrency Primitives**: Mutexes, RwLocks, atomic operations
- **Advanced Error Handling**: Custom error types, error propagation

### Level 3: Performance & Systems (See modules/optimization.md)

Production deployment and optimization:
- **Memory Optimization**: Stack vs heap, SIMD, inline assembly
- **Async Performance**: Tokio tuning, buffer management, backpressure
- **Zero-Cost Abstractions**: Monomorphization, inlining, compile-time evaluation
- **Unsafe Code Patterns**: FFI, low-level I/O, optimization
- **Profiling & Debugging**: Flamegraph, valgrind, miri

---

## Best Practices

✅ **DO**:
- Use ownership rules to ensure memory safety
- Prefer Result over panicking
- Use async/await with Tokio for I/O
- Leverage pattern matching for clarity
- Write comprehensive tests
- Use type system to encode invariants
- Profile before optimizing

❌ **DON'T**:
- Fight the borrow checker - understand it
- Use unwrap() in production
- Create unbounded async tasks
- Ignore compiler warnings
- Use unsafe without justification
- Clone excessively (use references)
- Ignore error handling

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **Rust** | 1.91.1 | Language |
| **Cargo** | Latest | Package manager |
| **Tokio** | 1.48 | Async runtime |
| **Axum** | 0.8 | Web framework |
| **async-trait** | 0.1 | Async traits |
| **parking_lot** | 0.12 | Sync primitives |

---

## Installation & Setup

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Create new Cargo project
cargo new myapp
cd myapp

# Add dependencies
cargo add tokio --features full
cargo add axum
cargo add serde --features derive

# Build and run
cargo build
cargo run
```

---

## Works Well With

- `moai-essentials-perf` (Performance optimization)
- `moai-essentials-debug` (Debugging with gdb)
- `moai-security-backend` (Security patterns)
- `moai-foundation-testing` (Test strategies)

---

## Learn More

- **Practical Examples**: See `examples.md` for 15+ real-world patterns
- **Advanced Patterns**: See `modules/advanced-patterns.md` for traits, async, macros
- **Performance Tuning**: See `modules/optimization.md` for memory, concurrency optimization
- **Official Docs**: https://doc.rust-lang.org/
- **Tokio Guide**: https://tokio.rs/
- **Axum**: https://github.com/tokio-rs/axum

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized structure with advanced patterns and optimization modules
- **v3.0.0** (2025-11-13): Rust 1.91 features and Tokio 1.48 updates
- **v2.0.0** (2025-10-01): Async/await and ownership deep dive
- **v1.0.0** (2025-03-01): Initial release

---

## Context7 Integration

### Related Libraries & Tools
- [Rust](/rust-lang/rust): Systems programming language with memory safety
- [Tokio](/tokio-rs/tokio): Asynchronous runtime for Rust
- [Axum](/tokio-rs/axum): Ergonomic web framework
- [Serde](/serde-rs/serde): Serialization framework

### Official Documentation
- [Rust Documentation](https://doc.rust-lang.org/)
- [Tokio Tutorial](https://tokio.rs/tokio/tutorial)
- [Rust Book](https://doc.rust-lang.org/book/)
- [Axum Examples](https://github.com/tokio-rs/axum/tree/main/examples)

### Version-Specific Guides
Latest stable version: Rust 1.91.1, Tokio 1.48, Axum 0.8
- [Rust Release Notes](https://github.com/rust-lang/rust/releases)
- [Tokio Changelog](https://github.com/tokio-rs/tokio/releases)
- [Axum Changelog](https://github.com/tokio-rs/axum/releases)

---

**Skills**: Skill("moai-essentials-perf"), Skill("moai-essentials-debug"), Skill("moai-security-backend")
**Auto-loads**: Rust projects with Tokio, Axum, async patterns

