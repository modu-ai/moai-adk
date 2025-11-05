# 다국어 린트/포맷 설치 가이드

## 빠른 설치

각 언어별 필수 도구를 설치하세요.

### Python 프로젝트

```bash
# 선택사항: pyproject.toml에 추가
[project.optional-dependencies]
dev = ["ruff>=0.1.0", "mypy>=1.0.0"]

# uv 사용 (권장)
uv add --optional ruff mypy

# pip 사용
pip install ruff mypy

# 설치 확인
ruff --version    # e.g., ruff 0.1.0
mypy --version    # e.g., mypy 1.0.0
```

### JavaScript/TypeScript 프로젝트

```bash
# 추천 도구 설치
npm install --save-dev eslint prettier typescript

# 또는 pnpm
pnpm add -D eslint prettier typescript

# 또는 yarn
yarn add --dev eslint prettier typescript

# 설치 확인
npx eslint --version      # e.g., v8.50.0
npx prettier --version    # e.g., 3.0.3
npx tsc --version         # e.g., Version 5.2.2
```

### Go 프로젝트

```bash
# golangci-lint 설치
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 설치 확인
golangci-lint version    # e.g., v1.54.2

# go.mod에서 버전 고정 (선택사항)
# go 1.16 이상
```

### Rust 프로젝트

```bash
# Rust 설치 (이미 설치된 경우 건너뛰기)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# clippy와 rustfmt 업데이트
rustup update

# 설치 확인
cargo clippy --version    # e.g., clippy 0.1.73
rustfmt --version        # e.g., 1.6.0
```

### Java 프로젝트

**Maven 사용:**

```bash
# pom.xml에 플러그인 추가
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-checkstyle-plugin</artifactId>
    <version>3.3.1</version>
</plugin>

# 또는 명령줄에서
mvn checkstyle:check
```

**Gradle 사용:**

```bash
# build.gradle에 플러그인 추가
plugins {
    id 'checkstyle'
}

# 또는 명령줄에서
gradle checkstyle
```

### Ruby 프로젝트

```bash
# Gemfile에 추가
group :development do
  gem 'rubocop', '~> 1.56'
end

# 또는 직접 설치
gem install rubocop

# 설치 확인
rubocop --version    # e.g., 1.56.0
```

### PHP 프로젝트

```bash
# composer.json 업데이트
composer require --dev phpstan/phpstan friendsofphp/php-cs-fixer

# 또는 개별 설치
composer require --dev phpstan/phpstan
composer require --dev friendsofphp/php-cs-fixer

# 설치 확인
vendor/bin/phpstan --version      # e.g., PHPStan 1.10.0
vendor/bin/php-cs-fixer --version # e.g., PHP CS Fixer 3.13.0
```

### C# 프로젝트

```bash
# .NET SDK 설치 (이미 설치된 경우 건너뛰기)
dotnet --version

# 프로젝트에서 Roslyn 도구 사용
dotnet tool install -g dotnet-codeanalyzer
```

### Kotlin 프로젝트

```bash
# build.gradle.kts에 플러그인 추가
plugins {
    id("org.jlleitschuh.gradle.ktlint") version "11.6.1"
}

# 또는 직접 설치
brew install ktlint  # macOS
apt-get install ktlint  # Linux
```

## 멀티언어 프로젝트 빠른 설정

여러 언어를 사용하는 프로젝트:

### 예: Python + TypeScript 프로젝트

```bash
# Python 부분
uv add --optional ruff mypy

# JavaScript/TypeScript 부분
npm install --save-dev eslint prettier typescript
```

### 예: Go + Rust 프로젝트

```bash
# Go 부분
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Rust 부분
rustup update
```

### 예: Java + Ruby 프로젝트

**Java (Maven):**
```bash
mvn checkstyle:check
```

**Ruby:**
```bash
gem install rubocop
```

## 버전 관리

### 권장 버전

| 도구 | 최소 버전 | 권장 버전 |
|------|----------|---------|
| ruff | 0.1.0 | 최신 |
| eslint | 8.0.0 | 최신 |
| prettier | 2.8.0 | 3.0.0+ |
| typescript | 4.9.0 | 5.0.0+ |
| golangci-lint | 1.50.0 | 최신 |
| cargo clippy | - | 최신 |
| checkstyle | 10.0.0 | 최신 |
| rubocop | 1.50.0 | 최신 |
| phpstan | 1.9.0 | 최신 |

### 버전 락 설정

**Python (pyproject.toml):**

```toml
[project.optional-dependencies]
dev = [
    "ruff>=0.1.0,<1.0.0",
    "mypy>=1.0.0,<2.0.0"
]
```

**JavaScript (package.json):**

```json
{
  "devDependencies": {
    "eslint": "^8.50.0",
    "prettier": "^3.0.0",
    "typescript": "^5.0.0"
  }
}
```

## 도구별 설정 파일

각 도구의 설정 파일을 프로젝트 루트에 생성하세요:

### Python (ruff)

`pyproject.toml` 또는 `ruff.toml`:

```toml
[tool.ruff]
line-length = 120
target-version = "py39"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "N"]
ignore = ["E501"]

[tool.mypy]
python_version = "3.9"
warn_return_any = true
warn_unused_configs = true
```

### JavaScript/TypeScript

`.eslintrc.json`:

```json
{
  "env": {
    "browser": true,
    "es2021": true
  },
  "extends": ["eslint:recommended"],
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 12,
    "sourceType": "module"
  },
  "rules": {
    "semi": ["error", "always"],
    "quotes": ["error", "single"]
  }
}
```

`.prettierrc.json`:

```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5"
}
```

### Go

`.golangci.yml`:

```yaml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck

linters-settings:
  gofmt:
    simplify: true
```

### Rust

`rustfmt.toml` 또는 `.rustfmt.toml`:

```toml
max_width = 100
hard_tabs = false
tab_spaces = 4
newline_style = "Auto"
```

### Ruby

`.rubocop.yml`:

```yaml
AllCops:
  TargetRubyVersion: 3.0

Style/FrozenStringLiteralComment:
  Enabled: false

Metrics/LineLength:
  Max: 100
```

### PHP

`php-cs-fixer.dist.php`:

```php
<?php

return (new PhpCsFixer\Config())
    ->setRiskyAllowed(true)
    ->setRules([
        '@PSR12' => true,
        'array_syntax' => ['syntax' => 'short'],
    ])
    ->setFinder(
        PhpCsFixer\Finder::create()
            ->in(__DIR__ . '/src')
    );
?>
```

## 문제 해결

### 1. 도구가 설치되어 있지 않음

**증상:** `⚠️ [Tool] not installed`

**해결책:**

1. 해당 언어의 설치 명령어를 실행하세요.
2. 도구가 PATH에 있는지 확인하세요:

```bash
# Python 도구
which ruff
which mypy

# Node 도구
which npx
npx eslint --version

# Go 도구
which golangci-lint

# Rust 도구
which cargo
```

### 2. 권한 오류 (Permission Denied)

**증상:** `permission denied`

**해결책:**

```bash
# Linux/macOS
chmod +x /usr/local/bin/ruff

# 또는 sudo 사용
sudo gem install rubocop
```

### 3. Python 가상 환경 문제

**증상:** `ModuleNotFoundError: No module named 'ruff'`

**해결책:**

```bash
# 현재 가상 환경 확인
python -m venv venv
source venv/bin/activate  # Linux/macOS
venv\Scripts\activate     # Windows

# 다시 설치
pip install ruff mypy
```

### 4. Node 모듈 경로 문제

**증상:** `Command not found: eslint`

**해결책:**

```bash
# 모듈이 설치되었는지 확인
npm ls eslint

# 로컬 설치되어 있으면 npx 사용
npx eslint src/

# 또는 전역 설치
npm install -g eslint prettier
```

### 5. Go 경로 문제

**증상:** `command not found: golangci-lint`

**해결책:**

```bash
# Go bin 디렉토리를 PATH에 추가
export PATH=$PATH:$(go env GOPATH)/bin

# .bashrc 또는 .zshrc에 추가
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

## 자동화된 설치 스크립트

### 다국어 프로젝트 일괄 설정

`install-linters.sh`:

```bash
#!/bin/bash

PROJECT_ROOT=$(pwd)

# Python
if [ -f "$PROJECT_ROOT/pyproject.toml" ]; then
    echo "Installing Python tools..."
    uv add --optional ruff mypy
fi

# JavaScript/TypeScript
if [ -f "$PROJECT_ROOT/package.json" ]; then
    echo "Installing JavaScript tools..."
    npm install --save-dev eslint prettier typescript
fi

# Go
if [ -f "$PROJECT_ROOT/go.mod" ]; then
    echo "Installing Go tools..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Rust
if [ -f "$PROJECT_ROOT/Cargo.toml" ]; then
    echo "Installing Rust tools..."
    rustup update
fi

# Java (Maven)
if [ -f "$PROJECT_ROOT/pom.xml" ]; then
    echo "Maven project detected. Use: mvn checkstyle:check"
fi

# Ruby
if [ -f "$PROJECT_ROOT/Gemfile" ]; then
    echo "Installing Ruby tools..."
    gem install rubocop
fi

# PHP
if [ -f "$PROJECT_ROOT/composer.json" ]; then
    echo "Installing PHP tools..."
    composer require --dev phpstan/phpstan friendsofphp/php-cs-fixer
fi

echo "Installation complete!"
```

실행 권한 부여 및 실행:

```bash
chmod +x install-linters.sh
./install-linters.sh
```

## CI/CD 통합

### GitHub Actions

`.github/workflows/lint.yml`:

```yaml
name: Linting

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install Python tools
        run: pip install ruff mypy

      - name: Run ruff
        run: ruff check .

      - name: Set up Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install Node tools
        run: npm install eslint prettier

      - name: Run eslint
        run: npx eslint src/
```

## 참고 자료

- [MoAI-ADK Linting Guide](./MULTILINGUAL_LINTING_GUIDE.md)
- [ruff Documentation](https://docs.astral.sh/ruff/)
- [eslint Documentation](https://eslint.org/docs/)
- [golangci-lint Documentation](https://golangci-lint.run/usage/install/)
