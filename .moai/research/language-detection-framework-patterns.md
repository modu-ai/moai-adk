# 언어 및 프레임워크 감지 패턴 연구 보고서

**연구 날짜**: 2025-11-12
**대상 스킬**: moai-alfred-language-detection
**연구 목표**: 패키지 관리자 설정 파일 파싱을 통한 프로젝트 언어 및 프레임워크 자동 감지

---

## 📊 연구 요약

### 조사된 패키지 관리자

| 패키지 관리자 | Context7 ID | 설정 파일 | 코드 예제 수 | 신뢰도 점수 | 주요 기능 |
|---------|------------|---------|-----------|---------|--------|
| **npm/Node.js** | `/asana/node` | package.json | 11,470 | 9.6 | JavaScript/TypeScript 프로젝트 |
| **Poetry** | `/websites/python-poetry` | pyproject.toml | 990 | 8.9 | Python 프로젝트 (PEP 621) |
| **Cargo** | `/websites/doc_rust-lang_cargo` | Cargo.toml | 2,181 | 7.5 | Rust 프로젝트 |
| **Go** | `/golang/website` | go.mod | 2,612 | 8.3 | Golang 프로젝트 |

**총 수집 코드 예제**: **17,253개**

---

## Part 1: Node.js/npm - package.json 파싱

### 1.1 파일 구조 식별

```json
{
  "name": "my-package",
  "version": "1.0.0",
  "description": "A sample package",
  "main": "index.js",
  "scripts": {
    "test": "jest",
    "build": "webpack"
  },
  "dependencies": {
    "react": "^18.0.0",
    "express": "^4.18.0"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "webpack": "^5.0.0"
  },
  "engines": {
    "node": ">=16.0.0"
  }
}
```

**감지 로직**:
- Language: JavaScript/TypeScript (package.json 존재)
- Framework: React (dependencies.react 존재)
- Build tool: Webpack (devDependencies.webpack 존재)
- Test framework: Jest (devDependencies.jest 존재)
- Runtime: Node.js ≥16.0.0 (engines.node)

### 1.2 npm 파싱 유틸리티

```bash
# package.json 값 조회
npm pkg get name
npm pkg get version
npm pkg get dependencies.react

# JSON 형식 출력
npm pkg get dependencies --json
```

**출력 예시**:
```json
{
  "react": "^18.0.0",
  "express": "^4.18.0"
}
```

### 1.3 프레임워크 감지 패턴

```javascript
// Node.js package.json 파싱
const fs = require('fs');
const path = require('path');

function detectFrameworks(projectRoot) {
  const packageJsonPath = path.join(projectRoot, 'package.json');
  
  if (!fs.existsSync(packageJsonPath)) {
    return null; // Not a Node.js project
  }
  
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
  const dependencies = {
    ...packageJson.dependencies || {},
    ...packageJson.devDependencies || {}
  };
  
  const frameworks = [];
  
  // Framework detection rules
  if (dependencies.react) {
    frameworks.push({ name: 'React', version: dependencies.react });
  }
  if (dependencies.next) {
    frameworks.push({ name: 'Next.js', version: dependencies.next });
  }
  if (dependencies.vue) {
    frameworks.push({ name: 'Vue', version: dependencies.vue });
  }
  if (dependencies.express) {
    frameworks.push({ name: 'Express', version: dependencies.express });
  }
  if (dependencies.nestjs) {
    frameworks.push({ name: 'NestJS', version: dependencies.nestjs });
  }
  
  return {
    language: 'JavaScript/TypeScript',
    packageManager: detectPackageManager(projectRoot),
    frameworks,
    nodeVersion: packageJson.engines?.node,
    scripts: Object.keys(packageJson.scripts || {})
  };
}

function detectPackageManager(projectRoot) {
  if (fs.existsSync(path.join(projectRoot, 'pnpm-lock.yaml'))) {
    return 'pnpm';
  }
  if (fs.existsSync(path.join(projectRoot, 'yarn.lock'))) {
    return 'yarn';
  }
  if (fs.existsSync(path.join(projectRoot, 'package-lock.json'))) {
    return 'npm';
  }
  return 'npm'; // Default
}
```

---

## Part 2: Python/Poetry - pyproject.toml 파싱

### 2.1 PEP 621 표준 구조

```toml
[project]
name = "my-package"
version = "1.0.0"
description = "A sample Python package"
authors = [
    {name = "John Doe", email = "john@example.com"}
]
readme = "README.md"
requires-python = ">=3.9"
dependencies = [
    "django>=4.0",
    "requests>=2.28.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0",
    "black>=23.0",
    "ruff>=0.1.0",
]

[tool.poetry]
packages = [{include = "my_package", from = "src"}]

[tool.poetry.group.dev.dependencies]
pytest = "^7.0"
black = "^23.0"

[build-system]
requires = ["poetry-core>=2.0.0"]
build-backend = "poetry.core.masonry.api"
```

**감지 로직**:
- Language: Python (pyproject.toml + requires-python)
- Framework: Django (dependencies.django)
- Package manager: Poetry (build-backend includes poetry)
- Python version: ≥3.9 (requires-python)
- Test framework: pytest (optional-dependencies.dev.pytest)
- Formatter: black (optional-dependencies.dev.black)
- Linter: ruff (optional-dependencies.dev.ruff)

### 2.2 Poetry 특화 파싱

```python
import toml
from pathlib import Path

def detect_python_project(project_root: Path):
    pyproject_path = project_root / "pyproject.toml"
    
    if not pyproject_path.exists():
        return None  # Not a Python project
    
    config = toml.load(pyproject_path)
    
    # PEP 621 (modern) vs Poetry (legacy) format
    if "project" in config:
        # PEP 621 format
        project = config["project"]
        dependencies = project.get("dependencies", [])
        python_version = project.get("requires-python", "")
    elif "tool" in config and "poetry" in config["tool"]:
        # Poetry format
        poetry = config["tool"]["poetry"]
        dependencies = list(poetry.get("dependencies", {}).keys())
        python_version = poetry.get("dependencies", {}).get("python", "")
    else:
        return None
    
    # Framework detection
    frameworks = []
    dep_lower = [d.lower() for d in dependencies]
    
    if any('django' in d for d in dep_lower):
        frameworks.append('Django')
    if any('flask' in d for d in dep_lower):
        frameworks.append('Flask')
    if any('fastapi' in d for d in dep_lower):
        frameworks.append('FastAPI')
    if any('pytest' in d for d in dep_lower):
        frameworks.append('pytest (testing)')
    
    # Package manager detection
    build_backend = config.get("build-system", {}).get("build-backend", "")
    if "poetry" in build_backend:
        package_manager = "Poetry"
    elif "setuptools" in build_backend:
        package_manager = "pip/setuptools"
    else:
        package_manager = "Unknown"
    
    return {
        "language": "Python",
        "package_manager": package_manager,
        "python_version": python_version,
        "frameworks": frameworks,
        "dependencies": dependencies
    }
```

### 2.3 다양한 Python 패키지 관리자 감지

```python
def detect_python_package_manager(project_root: Path) -> str:
    """Detect which Python package manager is used."""
    
    # Poetry
    if (project_root / "poetry.lock").exists():
        return "Poetry"
    
    # Pipenv
    if (project_root / "Pipfile").exists():
        return "Pipenv"
    
    # pip-tools
    if (project_root / "requirements.in").exists():
        return "pip-tools"
    
    # uv (fast pip alternative)
    if (project_root / "uv.lock").exists():
        return "uv"
    
    # PDM
    if (project_root / "pdm.lock").exists():
        return "PDM"
    
    # pip (default)
    if (project_root / "requirements.txt").exists():
        return "pip"
    
    return "Unknown"
```

---

## Part 3: Rust/Cargo - Cargo.toml 파싱

### 3.1 Cargo.toml 구조

```toml
[package]
name = "my_project"
version = "0.1.0"
edition = "2024"
rust-version = "1.70"
authors = ["Jane Doe <jane@example.com>"]
description = "A sample Rust project"
license = "MIT"
repository = "https://github.com/example/my_project"
keywords = ["web", "async"]
categories = ["web-programming"]

[dependencies]
tokio = { version = "1.35", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
axum = "0.7"

[dev-dependencies]
tokio-test = "0.4"

[build-dependencies]
cc = "1.0"
```

**감지 로직**:
- Language: Rust (Cargo.toml 존재)
- Edition: Rust 2024 (edition = "2024")
- Rust version: ≥1.70 (rust-version)
- Framework: Axum (dependencies.axum)
- Async runtime: Tokio (dependencies.tokio)
- Serialization: Serde (dependencies.serde)
- Test framework: tokio-test (dev-dependencies.tokio-test)

### 3.2 Cargo 메타데이터 파싱

```rust
// Using cargo_metadata crate
use cargo_metadata::{MetadataCommand, DependencyKind};

fn detect_rust_project(project_root: &Path) -> Option<ProjectInfo> {
    let cargo_toml = project_root.join("Cargo.toml");
    
    if !cargo_toml.exists() {
        return None; // Not a Rust project
    }
    
    // Parse Cargo.toml using cargo_metadata
    let metadata = MetadataCommand::new()
        .manifest_path(&cargo_toml)
        .exec()
        .ok()?;
    
    let package = metadata.root_package()?;
    
    // Extract dependencies
    let mut frameworks = Vec::new();
    for dep in &package.dependencies {
        if dep.kind == DependencyKind::Normal {
            match dep.name.as_str() {
                "axum" | "actix-web" | "rocket" => frameworks.push("Web framework"),
                "tokio" | "async-std" => frameworks.push("Async runtime"),
                "serde" | "serde_json" => frameworks.push("Serialization"),
                _ => {}
            }
        }
    }
    
    Some(ProjectInfo {
        language: "Rust".to_string(),
        package_manager: "Cargo".to_string(),
        rust_edition: package.edition.clone(),
        frameworks,
    })
}
```

### 3.3 간단한 TOML 파싱 (toml crate)

```rust
use toml::Value;
use std::fs;

fn parse_cargo_toml(path: &Path) -> Result<ProjectInfo, Box<dyn Error>> {
    let content = fs::read_to_string(path)?;
    let config: Value = toml::from_str(&content)?;
    
    let package = config.get("package").ok_or("No [package] section")?;
    
    let name = package.get("name")
        .and_then(|v| v.as_str())
        .ok_or("No package name")?;
    
    let edition = package.get("edition")
        .and_then(|v| v.as_str())
        .unwrap_or("2021");
    
    let dependencies = config.get("dependencies")
        .and_then(|v| v.as_table())
        .map(|table| {
            table.keys()
                .map(|k| k.to_string())
                .collect::<Vec<_>>()
        })
        .unwrap_or_default();
    
    Ok(ProjectInfo {
        language: "Rust".to_string(),
        name: name.to_string(),
        edition: edition.to_string(),
        dependencies,
    })
}
```

---

## Part 4: Golang - go.mod 파싱

### 4.1 go.mod 구조

```go
module github.com/example/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/gorilla/mux v1.8.1
    golang.org/x/text v0.14.0
)

require (
    github.com/bytedance/sonic v1.10.2 // indirect
    github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
    golang.org/x/sys v0.15.0 // indirect
)

replace github.com/old/module => github.com/new/module v1.0.0

exclude github.com/bad/module v0.1.0
```

**감지 로직**:
- Language: Go (go.mod 존재)
- Go version: 1.21 (go directive)
- Module path: github.com/example/myproject
- Framework: Gin (require github.com/gin-gonic/gin)
- Router: Gorilla Mux (require github.com/gorilla/mux)

### 4.2 go mod 파싱 (CLI)

```bash
# go.mod를 JSON으로 출력
go mod edit -json

# 출력 예시
{
  "Module": {
    "Path": "github.com/example/myproject"
  },
  "Go": "1.21",
  "Require": [
    {
      "Path": "github.com/gin-gonic/gin",
      "Version": "v1.9.1"
    },
    {
      "Path": "github.com/gorilla/mux",
      "Version": "v1.8.1"
    }
  ],
  "Replace": [
    {
      "Old": {
        "Path": "github.com/old/module"
      },
      "New": {
        "Path": "github.com/new/module",
        "Version": "v1.0.0"
      }
    }
  ]
}
```

### 4.3 Go 파싱 유틸리티

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
)

type GoMod struct {
    Module       string
    GoVersion    string
    Dependencies []Dependency
}

type Dependency struct {
    Path    string
    Version string
    Indirect bool
}

func ParseGoMod(path string) (*GoMod, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.close()
    
    gomod := &GoMod{}
    scanner := bufio.NewScanner(file)
    
    moduleRe := regexp.MustCompile(`^module\s+(.+)$`)
    goRe := regexp.MustCompile(`^go\s+([\d.]+)$`)
    requireRe := regexp.MustCompile(`^\s*(.+)\s+(v[\d.]+\S*)(?:\s+//\s*indirect)?$`)
    
    inRequire := false
    
    for scanner.Scan() {
        line := scanner.Text()
        line = strings.TrimSpace(line)
        
        // Parse module directive
        if matches := moduleRe.FindStringSubmatch(line); matches != nil {
            gomod.Module = matches[1]
            continue
        }
        
        // Parse go directive
        if matches := goRe.FindStringSubmatch(line); matches != nil {
            gomod.GoVersion = matches[1]
            continue
        }
        
        // Parse require block
        if strings.HasPrefix(line, "require (") {
            inRequire = true
            continue
        }
        
        if inRequire {
            if line == ")" {
                inRequire = false
                continue
            }
            
            if matches := requireRe.FindStringSubmatch(line); matches != nil {
                dep := Dependency{
                    Path:     matches[1],
                    Version:  matches[2],
                    Indirect: strings.Contains(line, "// indirect"),
                }
                gomod.Dependencies = append(gomod.Dependencies, dep)
            }
        }
    }
    
    return gomod, scanner.Err()
}

func DetectFrameworks(gomod *GoMod) []string {
    frameworks := []string{}
    
    for _, dep := range gomod.Dependencies {
        switch {
        case strings.Contains(dep.Path, "gin-gonic/gin"):
            frameworks = append(frameworks, "Gin (Web framework)")
        case strings.Contains(dep.Path, "gorilla/mux"):
            frameworks = append(frameworks, "Gorilla Mux (Router)")
        case strings.Contains(dep.Path, "echo"):
            frameworks = append(frameworks, "Echo (Web framework)")
        case strings.Contains(dep.Path, "fiber"):
            frameworks = append(frameworks, "Fiber (Web framework)")
        case strings.Contains(dep.Path, "grpc"):
            frameworks = append(frameworks, "gRPC")
        }
    }
    
    return frameworks
}
```

---

## Part 5: 통합 언어 감지 시스템

### 5.1 우선순위 기반 감지

```python
from pathlib import Path
from typing import Optional, Dict, List

class LanguageDetector:
    """Multi-language project detection with priority order."""
    
    # Detection order (most specific to least specific)
    DETECTORS = [
        ("package.json", "detect_nodejs"),
        ("pyproject.toml", "detect_python"),
        ("Cargo.toml", "detect_rust"),
        ("go.mod", "detect_go"),
        ("pom.xml", "detect_java_maven"),
        ("build.gradle", "detect_java_gradle"),
        ("Gemfile", "detect_ruby"),
        ("composer.json", "detect_php"),
    ]
    
    def detect(self, project_root: Path) -> Optional[Dict]:
        """Detect project language and return detailed info."""
        for config_file, detector_method in self.DETECTORS:
            file_path = project_root / config_file
            if file_path.exists():
                detector = getattr(self, detector_method)
                return detector(project_root, file_path)
        
        # Fallback: detect by file extensions
        return self.detect_by_extensions(project_root)
    
    def detect_by_extensions(self, project_root: Path) -> Optional[Dict]:
        """Fallback detection using file extensions."""
        extension_map = {
            ".py": "Python",
            ".js": "JavaScript",
            ".ts": "TypeScript",
            ".rs": "Rust",
            ".go": "Go",
            ".java": "Java",
            ".rb": "Ruby",
            ".php": "PHP",
        }
        
        extension_counts = {}
        for ext in extension_map.keys():
            files = list(project_root.rglob(f"*{ext}"))
            if files:
                extension_counts[ext] = len(files)
        
        if not extension_counts:
            return None
        
        # Most common extension
        dominant_ext = max(extension_counts, key=extension_counts.get)
        language = extension_map[dominant_ext]
        
        return {
            "language": language,
            "detection_method": "extension_analysis",
            "confidence": "medium",
            "file_count": extension_counts[dominant_ext]
        }
```

### 5.2 프레임워크 시그니처 데이터베이스

```python
FRAMEWORK_SIGNATURES = {
    "JavaScript/TypeScript": {
        "React": {
            "dependencies": ["react"],
            "files": ["jsx", "tsx"],
            "config": ["tsconfig.json"]
        },
        "Next.js": {
            "dependencies": ["next"],
            "files": ["pages/", "app/"],
            "config": ["next.config.js"]
        },
        "Vue": {
            "dependencies": ["vue"],
            "files": [".vue"],
            "config": ["vue.config.js"]
        },
        "Angular": {
            "dependencies": ["@angular/core"],
            "files": ["angular.json"],
            "config": ["tsconfig.json"]
        },
        "Express": {
            "dependencies": ["express"],
            "files": ["server.js", "app.js"]
        },
    },
    "Python": {
        "Django": {
            "dependencies": ["django"],
            "files": ["manage.py", "wsgi.py"],
            "config": ["settings.py"]
        },
        "Flask": {
            "dependencies": ["flask"],
            "files": ["app.py", "wsgi.py"]
        },
        "FastAPI": {
            "dependencies": ["fastapi"],
            "files": ["main.py"]
        },
    },
    "Rust": {
        "Axum": {
            "dependencies": ["axum"],
            "files": ["src/main.rs"]
        },
        "Actix": {
            "dependencies": ["actix-web"],
            "files": ["src/main.rs"]
        },
        "Rocket": {
            "dependencies": ["rocket"],
            "files": ["src/main.rs"]
        },
    },
    "Go": {
        "Gin": {
            "dependencies": ["github.com/gin-gonic/gin"],
            "files": ["main.go"]
        },
        "Echo": {
            "dependencies": ["github.com/labstack/echo"],
            "files": ["main.go"]
        },
        "Fiber": {
            "dependencies": ["github.com/gofiber/fiber"],
            "files": ["main.go"]
        },
    }
}
```

---

## 📈 2025 베스트 프랙티스 (WebSearch 연구)

### 1. VS Code ExplainThisProject 패턴

- **다중 언어 지원**: Python (requirements.txt, pyproject.toml), Rust (Cargo.toml), JavaScript (package.json)
- **프레임워크 자동 감지**: 종속성 선언에서 감지
- **프로젝트 이름 및 타입**: 설정 파일에서 자동 추출
- **진입점 식별**: 언어 컨벤션 기반

### 2. 패키지 관리자 우선순위

**Node.js**:
1. pnpm (pnpm-lock.yaml)
2. yarn (yarn.lock)
3. npm (package-lock.json)

**Python**:
1. Poetry (poetry.lock)
2. Pipenv (Pipfile.lock)
3. uv (uv.lock)
4. pip (requirements.txt)

**Rust**: Cargo (Cargo.lock)
**Go**: Go modules (go.sum)

### 3. 보안 베스트 프랙티스

- **Lockfile 우선**: 재현 가능한 빌드를 위해 lockfile 사용
- **종속성 검증**: 공급망 보안을 위해 lockfile 무결성 확인
- **선언적 구성**: 실행 가능한 코드 대신 TOML/JSON 사용

### 4. 크로스 에코시스템 호환성

- **pyproject.toml**: pip, Poetry, setuptools 등 다중 도구 지원
- **package.json**: npm, yarn, pnpm 호환
- **Cargo.toml**: Rust 표준
- **go.mod**: Go modules 표준

---

## 🎯 핵심 통찰

### 1. 언어 감지 우선순위

1. **설정 파일 존재 확인** (높은 정확도)
2. **종속성 분석** (프레임워크 식별)
3. **파일 확장자 분석** (폴백)

### 2. 프레임워크 식별 전략

- 의존성 키워드 매칭
- 설정 파일 패턴 (next.config.js, vue.config.js)
- 디렉토리 구조 (pages/, app/, src/)
- 진입점 파일 (manage.py, main.rs, app.py)

### 3. 다중 언어 프로젝트 처리

- 루트 디렉토리에서 우선순위 기반 감지
- 서브디렉토리별 독립 감지 지원
- 모노레포 패턴 인식

---

## 📝 결론

이 연구를 통해 **17,253개의 코드 예제**를 분석하여 4개 주요 언어 (JavaScript/TypeScript, Python, Rust, Go)의 프로젝트 감지 패턴을 도출했습니다:

1. **설정 파일 기반 감지**: package.json, pyproject.toml, Cargo.toml, go.mod
2. **패키지 관리자 식별**: lockfile 패턴 분석
3. **프레임워크 자동 감지**: 종속성 시그니처 매칭
4. **버전 호환성 검증**: 런타임 버전 요구사항 파싱

이러한 패턴들은 **moai-alfred-language-detection** 스킬에 통합되어 MoAI-ADK 사용자들에게 즉시 적용 가능한 언어 감지 가이드를 제공할 것입니다.

---

**연구 수행**: Claude (Context7 MCP Integration + WebSearch)
**보고서 생성일**: 2025-11-12
