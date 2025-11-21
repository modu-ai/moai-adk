    }
    
    # Build system detection
    if 'tool' in toml_data and 'poetry' in toml_data['tool']:
        detection['build_system'] = 'poetry'
        detection['package_manager'] = 'poetry'
        
        # Extract Python version from Poetry
        python_dep = toml_data['tool']['poetry'].get('dependencies', {}).get('python')
        if python_dep:
            detection['python_version'] = python_dep.replace('^', '>=')
    elif 'build-system' in toml_data:
        build_requires = toml_data['build-system'].get('requires', [])
        if 'setuptools' in str(build_requires):
            detection['build_system'] = 'setuptools'
        elif 'flit' in str(build_requires):
            detection['build_system'] = 'flit'
        elif 'hatch' in str(build_requires):
            detection['build_system'] = 'hatch'
    
    # Framework detection from dependencies
    deps = []
    if 'project' in toml_data and 'dependencies' in toml_data['project']:
        deps = toml_data['project']['dependencies']
    elif 'tool' in toml_data and 'poetry' in toml_data['tool']:
        deps = list(toml_data['tool']['poetry'].get('dependencies', {}).keys())
    
    deps_str = ' '.join(deps).lower()
    if 'fastapi' in deps_str:
        detection['frameworks'].append('fastapi')
    if 'django' in deps_str:
        detection['frameworks'].append('django')
    if 'flask' in deps_str:
        detection['frameworks'].append('flask')
    if 'pydantic' in deps_str:
        detection['frameworks'].append('pydantic')
    
    # Test framework detection
    dev_deps = []
    if 'project' in toml_data and 'optional-dependencies' in toml_data['project']:
        dev_deps = toml_data['project']['optional-dependencies'].get('dev', [])
    elif 'tool' in toml_data and 'poetry' in toml_data:
        dev_deps = list(toml_data['tool']['poetry'].get('group', {}).get('dev', {}).get('dependencies', {}).keys())
    
    dev_deps_str = ' '.join(dev_deps).lower()
    if 'pytest' in dev_deps_str:
        detection['test_frameworks'].append('pytest')
    if 'unittest' in dev_deps_str:
        detection['test_frameworks'].append('unittest')
    
    # Linting tools detection
    if 'black' in dev_deps_str:
        detection['linting_tools'].append('black')
    if 'flake8' in dev_deps_str:
        detection['linting_tools'].append('flake8')
    if 'mypy' in dev_deps_str:
        detection['linting_tools'].append('mypy')
    if 'ruff' in dev_deps_str:
        detection['linting_tools'].append('ruff')
    
    return detection
```


#### Pattern 3: Rust Project Detection - Cargo.toml

**Objective**: Identify Rust projects and extract crate information, dependencies, and edition.

**Cargo.toml Structure**:

```toml
[package]
name = "my-rust-app"
version = "0.1.0"
edition = "2021"
description = "Rust application with web framework"

[dependencies]
tokio = { version = "1.0", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
axum = "0.7"
sqlx = { version = "0.7", features = ["postgres", "runtime-tokio-rustls"] }

[dev-dependencies]
tokio-test = "0.4"
criterion = "0.5"

[features]
default = ["axum"]
rocket = ["dep:rocket", "tokio/sync"]

[workspace]
members = [
    "crates/core",
    "crates/utils",
    "crates/api"
]

[profile.release]
lto = true
codegen-units = 1
panic = "abort"
```

**Detection Logic**:
- **Edition**: 2021, 2021, or 2018 from package.edition
- **Frameworks**: axum, rocket, actix-web (from dependencies)
- **Async Runtime**: tokio, async-std (from dependencies)
- **Database**: sqlx, diesel (from dependencies)
- **Serialization**: serde, serde_json (from dependencies)
- **Workspace**: Multi-crate project if workspace members exist

**Implementation**:

```python
def detect_rust_project():
    """Detect Rust project details from Cargo.toml."""
    
    # Check for Cargo.toml
    if not os.path.exists('Cargo.toml'):
        return None
    
    with open('Cargo.toml', 'r') as f:
        cargo_data = toml.load(f)
    
    detection = {
        'language': 'rust',
        'frameworks': [],
        'async_runtime': None,
        'database_libs': [],
        'serialization': [],
        'edition': '2021',
        'is_workspace': False
    }
    
    # Package information
    if 'package' in cargo_data:
        package_info = cargo_data['package']
        detection['edition'] = package_info.get('edition', '2021')
    
    # Workspace detection
    if 'workspace' in cargo_data:
        detection['is_workspace'] = True
    
    # Dependencies analysis
    deps = cargo_data.get('dependencies', {})
    deps_str = ' '.join(deps.keys()).lower()
    
    # Framework detection
    if 'axum' in deps_str:
        detection['frameworks'].append('axum')
    if 'rocket' in deps_str:
        detection['frameworks'].append('rocket')
    if 'actix-web' in deps_str or 'actix' in deps_str:
        detection['frameworks'].append('actix-web')
    
    # Async runtime detection
    if 'tokio' in deps_str:
        detection['async_runtime'] = 'tokio'
    elif 'async-std' in deps_str:
        detection['async_runtime'] = 'async-std'
    
    # Database libraries detection
    if 'sqlx' in deps_str:
        detection['database_libs'].append('sqlx')
    if 'diesel' in deps_str:
        detection['database_libs'].append('diesel')
    if 'sea-orm' in deps_str:
        detection['database_libs'].append('sea-orm')
    
    # Serialization detection
    if 'serde' in deps_str:
        detection['serialization'].append('serde')
    if 'serde_json' in deps_str:
        detection['serialization'].append('serde_json')
    
    return detection
```


#### Pattern 4: Go Project Detection - go.mod

**Objective**: Identify Go projects and extract module information, Go version, and dependencies.

**go.mod Structure**:

```go
module github.com/username/my-go-app

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/lib/pq v1.10.9
    github.com/spf13/viper v1.17.0
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
    github.com/gabriel-vasile/mimetype v1.4.2 // indirect
)
```

**Detection Logic**:
- **Go Version**: From go directive (1.21, 1.20, etc.)
- **Frameworks**: gin, echo, fiber (from require statements)
- **Database**: lib/pq (PostgreSQL), go-sqlite3, gorm (from require)
- **Configuration**: viper, envconfig (from require)
- **Testing**: testify, gomock (from require)

**Implementation**:

```python
def detect_go_project():
    """Detect Go project details from go.mod."""
    
    # Check for go.mod
    if not os.path.exists('go.mod'):
        return None
    
    with open('go.mod', 'r') as f:
        go_mod_content = f.read()
    
    detection = {
        'language': 'go',
        'frameworks': [],
        'database_libs': [],
        'config_libs': [],
        'test_libs': [],
        'go_version': None,
        'module_name': None
    }
    
    lines = go_mod_content.split('\n')
    
    for line in lines:


## Implementation Guide




## Advanced Patterns



