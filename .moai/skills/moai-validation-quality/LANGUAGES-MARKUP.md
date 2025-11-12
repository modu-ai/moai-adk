# Markup & Configuration Languages Validation Patterns

> Referenced from: `moai-validation-quality/SKILL.md`

This document covers validation patterns for markup and configuration languages: **HTML, CSS, Tailwind CSS, Markdown, Shell, SQL**.

---

## HTML / CSS

### Technology Stack
- **HTML**: HTML5
- **CSS**: CSS3, SCSS, PostCSS
- **Package Manager**: npm (for tools)
- **Validation**: htmlhint, W3C validator
- **Linting**: stylelint

### Validation Tools

#### HTML Validation
```bash
# htmlhint (HTML validator)
brew install htmlhint

# Run validation
htmlhint src/**/*.html

# With specific rules
htmlhint src/ --config .htmlhintrc

# W3C validator (online or local)
# https://validator.w3.org/
```

#### CSS Linting (stylelint)
```bash
# Install stylelint
npm install -g stylelint stylelint-config-standard

# Run linting
stylelint "src/**/*.css"

# With config
stylelint "src/**/*.css" --config .stylelintrc.json
```

#### Formatting (Prettier)
```bash
# Format check
npx prettier --check "src/**/*.{html,css}"

# Auto-format
npx prettier --write "src/**/*.{html,css}"
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🌐 HTML/CSS Validation Starting..."

echo "▶ HTML Validation"
htmlhint src/**/*.html

echo "▶ CSS Linting"
stylelint "src/**/*.css" --config .stylelintrc.json

echo "▶ Formatting"
npx prettier --check "src/**/*.{html,css}"

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
html:
  validator: [htmlhint, html-validate]
  formatter: [prettier]

css:
  linter: [stylelint, csslint]
  formatter: [prettier, postcss]
```

---

## Tailwind CSS

### Technology Stack
- **Framework**: Tailwind CSS 3.x
- **Configuration**: tailwind.config.js
- **Building**: PostCSS, autoprefixer
- **Linting**: Tailwind-specific stylelint config

### Validation Tools

#### CSS Linting (Tailwind-specific)
```bash
# stylelint with Tailwind config
stylelint "src/**/*.css" --config .stylelintrc-tailwind.json

# Tailwind configuration validation
# Validate tailwind.config.js manually
```

#### HTML/JSX/TSX with Tailwind Classes
```bash
# Format all files
npx prettier --check "src/**/*.{html,jsx,tsx,css}"

# With Tailwind plugin
# npx prettier-plugin-tailwindcss
npx prettier --write "src/**/*.{jsx,tsx}" --plugin=prettier-plugin-tailwindcss
```

#### Content Scanning
```bash
# Ensure all Tailwind classes in content are scanned
# In tailwind.config.js:
# content: [
#   './src/**/*.{html,js,jsx,ts,tsx}',
# ]

# Test build
npx tailwindcss build src/input.css -o dist/output.css
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "💨 Tailwind CSS Validation Starting..."

echo "▶ Tailwind Build"
npx tailwindcss build src/input.css -o /dev/null

echo "▶ CSS Linting"
stylelint "src/**/*.css" --config .stylelintrc-tailwind.json

echo "▶ Class Formatting"
npx prettier --check "src/**/*.{jsx,tsx}" --plugin=prettier-plugin-tailwindcss

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
tailwind_css:
  linter: [stylelint]
  formatter: [prettier, prettier-plugin-tailwindcss]
  builder: [tailwindcss-cli]
```

---

## Markdown

### Technology Stack
- **Format**: Markdown, MDX (Markdown + JSX)
- **Validation**: markdownlint
- **Mermaid Diagrams**: @mermaid-js/mermaid-cli
- **Formatting**: Prettier, Remark

### Validation Tools

#### Markdown Linting
```bash
# Install markdownlint-cli2
npm install -g markdownlint-cli2

# Run linting
markdownlint-cli2 "**/*.md" --config .markdownlint.json

# Fix issues
markdownlint-cli2 "**/*.md" --fix --config .markdownlint.json
```

#### Mermaid Diagram Validation
```bash
# Install mermaid-cli
npm install -g @mermaid-js/mermaid-cli

# Validate diagram
mmdc --input diagram.mmd --output /dev/null --validate

# Render to PNG
mmdc --input architecture.mmd --output architecture.png
```

#### Formatting (Prettier)
```bash
# Check format
npx prettier --check "**/*.md"

# Auto-format
npx prettier --write "**/*.md"
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "📝 Markdown Validation Starting..."

echo "▶ Linting"
markdownlint-cli2 "**/*.md" --config .markdownlint.json

echo "▶ Mermaid Diagrams"
find . -name "*.mmd" -exec mmdc --input {} --output /dev/null --validate \;

echo "▶ Formatting"
npx prettier --check "**/*.md"

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
markdown:
  linter: [markdownlint-cli2, markdownlint, remark-lint]
  formatter: [prettier, remark]
  mermaid: [mermaid-cli]
```

---

## Shell Script

### Technology Stack
- **Shells**: bash, sh, zsh
- **POSIX Compliance**: Required
- **Testing**: BATS (Bash Automated Testing System), shunit2
- **Package Manager**: Homebrew, apt

### Validation Tools

#### Linting (ShellCheck)
```bash
# Install ShellCheck
brew install shellcheck

# Check single file
shellcheck script.sh

# Check all shell scripts
find . -name "*.sh" -exec shellcheck {} \;

# Report format
shellcheck --format=gcc script.sh
```

#### Formatting (shfmt)
```bash
# Install shfmt
brew install shfmt

# Check format
shfmt -d -i 2 **/*.sh

# Auto-format (2-space indent)
shfmt -i 2 -w **/*.sh
```

#### Testing (BATS)
```bash
# Install BATS
brew install bats-core

# Run tests
bats tests/*.bats

# With verbose output
bats tests/ --tap
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🐚 Shell Script Validation Starting..."

echo "▶ ShellCheck"
find . -name "*.sh" -exec shellcheck {} \;

echo "▶ Formatting"
shfmt -d -i 2 **/*.sh || (echo "Format issues found" && exit 1)

echo "▶ Testing"
bats tests/*.bats

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
shell:
  linter: [shellcheck]
  formatter: [shfmt, bash-formatter]
  test: [bats, shunit2]
```

---

## SQL

### Technology Stack
- **Databases**: PostgreSQL, MySQL, SQLite, others
- **Validation**: sqlfluff, sqlfmt
- **Formatting**: PostgreSQL formatter, pgformatter

### Validation Tools

#### Linting (sqlfluff)
```bash
# Install sqlfluff
pip install sqlfluff

# Lint SQL files
sqlfluff lint sql/

# Specific database dialect
sqlfluff lint sql/ --dialect postgres

# Fix issues
sqlfluff fix sql/ --dialect postgres
```

#### Formatting (pgformatter)
```bash
# Install pgformatter
brew install pgformatter  # macOS
# or pip install pgformatter

# Format SQL file
pg_format -o formatted.sql input.sql

# Check format
pg_format -o /dev/null input.sql  # validates syntax
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🗄️ SQL Validation Starting..."

echo "▶ Linting"
sqlfluff lint sql/ --dialect postgres

echo "▶ Format Check"
for file in sql/**/*.sql; do
  pgformatter --check "$file" || exit 1
done

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
sql:
  linter: [sqlfluff, sqlint]
  formatter: [pgformatter, sqlfmt, sqlparse]
```

---

## Installation Commands

### HTML/CSS Tools
```bash
# htmlhint
npm install -g htmlhint

# stylelint
npm install -g stylelint stylelint-config-standard

# Prettier
npm install -g prettier
```

### Markdown Tools
```bash
# markdownlint-cli2
npm install -g markdownlint-cli2

# Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Prettier
npm install -g prettier
```

### Shell Tools
```bash
# ShellCheck
brew install shellcheck  # macOS
sudo apt-get install shellcheck  # Linux

# shfmt
brew install shfmt

# BATS
brew install bats-core
```

### SQL Tools
```bash
# sqlfluff
pip install sqlfluff

# pgformatter
brew install pgformatter  # macOS
pip install pgformatter  # Python alternative
```

### Tailwind CSS Tools
```bash
# Tailwind CSS
npm install -g tailwindcss

# Prettier + Tailwind plugin
npm install -g prettier prettier-plugin-tailwindcss

# stylelint
npm install -g stylelint stylelint-config-standard
```

---

**Last Updated**: 2025-11-12
**Related**: [SKILL.md](SKILL.md), [LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md), [TOOL-REFERENCE.md](TOOL-REFERENCE.md)
