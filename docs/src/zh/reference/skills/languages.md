# Language Skills 完整指南

MoAI-ADK 支持的 20+ 编程语言技能参考。

## 🎯 概览

每个语言技能都包含该语言的最新版本、测试框架、代码检查工具和最佳实践。

## 🐍 Python (最受欢迎)

### 基本信息

- **版本**: Python 3.13+
- **包管理**: uv 0.9.3
- **测试**: pytest 8.4.2
- **类型检查**: mypy 1.8.0
- **代码检查**: ruff 0.13.1
- **框架**: FastAPI, Flask, Django

### 结构

```
python_project/
├── src/
│   ├── models.py
│   ├── services.py
│   └── __init__.py
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conftest.py
├── pyproject.toml
└── .python-version
```

### 核心工具

```bash
# 测试
pytest tests/ -v --cov=src

# 类型检查
mypy src/

# 代码检查
ruff check src/

# 格式化
black src/

# 包安装
uv add pytest
uv sync
```

### 示例

```python
# @CODE:SPEC-001:example
def add(a: int, b: int) -> int:
    """两数相加"""
    return a + b

# @TEST:SPEC-001:test_add
def test_add():
    assert add(2, 3) == 5
```

______________________________________________________________________

## 📘 TypeScript (Web 开发)

### 基本信息

- **版本**: TypeScript 5.7+
- **运行时**: Node.js 20+
- **包管理**: npm, pnpm, bun
- **测试**: Vitest 2.1
- **代码检查**: Biome 1.9
- **框架**: React 19, Vue 3.5, Angular 19

### 结构

```
typescript_project/
├── src/
│   ├── components/
│   ├── types/
│   ├── utils/
│   └── main.ts
├── tests/
│   ├── unit/
│   └── integration/
├── tsconfig.json
└── package.json
```

### 核心工具

```bash
# 测试
vitest run

# 类型检查
tsc --noEmit

# 代码检查
biome check src/

# 格式化
biome format src/
```

______________________________________________________________________

## 🦀 Rust (系统编程)

### 基本信息

- **版本**: Rust 1.84+
- **构建工具**: Cargo
- **测试**: cargo test
- **代码检查**: clippy
- **格式化**: rustfmt
- **Web 框架**: Axum, Rocket

### 结构

```
rust_project/
├── src/
│   ├── lib.rs
│   ├── main.rs
│   └── mod.rs
├── tests/
├── Cargo.toml
└── Cargo.lock
```

### 核心工具

```bash
# 测试
cargo test

# 代码检查
cargo clippy

# 格式化
cargo fmt

# 构建
cargo build --release
```

______________________________________________________________________

## 🐹 Go (后端)

### 基本信息

- **版本**: Go 1.24+
- **包管理**: go modules
- **测试**: go test
- **代码检查**: golangci-lint
- **格式化**: gofmt
- **Web 框架**: Gin, Beego

### 结构

```
go_project/
├── cmd/
│   └── main/main.go
├── internal/
│   ├── handlers/
│   └── services/
├── pkg/
├── tests/
└── go.mod
```

### 核心工具

```bash
# 测试
go test ./...

# 代码检查
golangci-lint run

# 格式化
gofmt -w .

# 构建
go build -o app
```

______________________________________________________________________

## 📜 其他热门语言

### Java (21+)

- 测试: JUnit 5, Mockito
- 构建: Maven, Gradle
- 代码检查: Checkstyle, SpotBugs
- 框架: Spring Boot, Quarkus

### C# (13+)

- 测试: xUnit, NUnit
- 构建: .NET SDK
- 代码检查: StyleCop, FxCop
- 框架: ASP.NET Core, Blazor

### PHP (8.4+)

- 测试: PHPUnit 11
- 包管理: Composer
- 代码检查: PHP_CodeSniffer, PHPStan
- 框架: Laravel 11, Symfony 7

### Ruby (3.4+)

- 测试: RSpec 4
- 包管理: Bundler, RubyGems
- 代码检查: RuboCop 2
- 框架: Rails 8

### Kotlin (2.1+)

- 测试: KUnit, Mockk
- 构建: Gradle, Maven
- 代码检查: ktlint, detekt
- 框架: Spring Boot, Ktor

### Dart (3.x)

- 测试: Dart test
- 构建: Pub, Flutter
- 代码检查: analysis server
- 框架: Flutter 3.27

## 🗂️ 按语言激活技能

### 自动激活

```
在 SPEC 中检测语言
    ├─ 检测到 "Python"
    │   └─→ 自动加载 Skill("moai-lang-python")
    ├─ 检测到 "TypeScript"
    │   └─→ 自动加载 Skill("moai-lang-typescript")
    ├─ 检测到 "Rust"
    │   └─→ 自动加载 Skill("moai-lang-rust")
    └─ 检测到 "Go"
        └─→ 自动加载 Skill("moai-lang-go")
```

### 手动调用

```python
# 显式调用特定语言技能
Skill("moai-lang-python")
Skill("moai-lang-typescript")
Skill("moai-lang-go")
Skill("moai-lang-rust")
Skill("moai-lang-java")
Skill("moai-lang-kotlin")
Skill("moai-lang-csharp")
Skill("moai-lang-ruby")
Skill("moai-lang-php")
Skill("moai-lang-sql")
Skill("moai-lang-shell")
```

## 📊 语言技能对比

| 语言       | 版本  | 测试       | 代码检查        | 类型     | 性能       |
| ---------- | ----- | ---------- | --------------- | -------- | ---------- |
| Python     | 3.13+ | pytest     | ruff            | mypy     | ⭐⭐⭐     |
| TypeScript | 5.7+  | Vitest     | Biome           | Built-in | ⭐⭐⭐⭐   |
| Rust       | 1.84+ | cargo test | clippy          | Built-in | ⭐⭐⭐⭐⭐ |
| Go         | 1.24+ | go test    | golangci        | Built-in | ⭐⭐⭐⭐⭐ |
| Java       | 21+   | JUnit      | Checkstyle      | Built-in | ⭐⭐⭐⭐   |
| C#         | 13+   | xUnit      | StyleCop        | Built-in | ⭐⭐⭐⭐   |
| Ruby       | 3.4+  | RSpec      | RuboCop         | Optional | ⭐⭐⭐     |
| PHP        | 8.4+  | PHPUnit    | PHP_CodeSniffer | PHPStan  | ⭐⭐⭐     |
| Kotlin     | 2.1+  | KUnit      | ktlint          | Built-in | ⭐⭐⭐⭐   |

## 🎯 语言选择指南

### API/后端

推荐: **Python** (FastAPI) > **Go** > **Rust** > **Java** > **TypeScript** (Node.js)

### 前端

推荐: **TypeScript** > **JavaScript** > **Dart** (Flutter)

### 移动

推荐: **Kotlin** (Android) > **Swift** (iOS) > **Dart** (Flutter - 跨平台)

### 系统编程

推荐: **Rust** > **Go** > **C++** > **C**

### 数据科学/ML

推荐: **Python** > **R** > **Julia**

### Web 全栈

推荐: **TypeScript** (React/Node.js) > **Ruby** (Rails) > **Python** (Django)

## 🔗 各语言最佳实践

所有语言技能都包括:

- ✅ 明确最新版本
- ✅ 项目结构指南
- ✅ 测试框架集成
- ✅ 代码检查和格式化自动化
- ✅ 保证类型安全
- ✅ 性能优化技术
- ✅ 部署自动化
- ✅ 最佳实践和反模式

## 📚 详细文档

每个语言技能都提供:

- **结构**: 项目目录结构
- **工具**: 测试、代码检查、格式化工具
- **示例**: 实际代码示例
- **性能**: 性能优化指南
- **测试**: 测试编写模式
- **部署**: CI/CD 自动化

______________________________________________________________________

**下一步**: [Alfred Skills](alfred.md) 或 [Skills 概览](index.md)
