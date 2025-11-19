# Setup Guide - Mermaid Diagram Expert v5.0.0

Complete setup instructions for local development and CI/CD integration.

## Local Installation

### 1. Basic Setup (5 minutes)

```bash
# Clone or navigate to your project
cd your-project

# Sync Python dependencies with uv
uv sync

# Copy CLI script
cp mermaid-to-svg-png.py ./scripts/
chmod +x ./scripts/mermaid-to-svg-png.py
```

### 2. Create Diagram Directory

```bash
mkdir -p diagrams
mkdir -p docs/images
```

### 3. Test Installation

```bash
# Create test diagram
cat > diagrams/test.mmd << 'EOF'
flowchart TD
    A[Start] --> B[Test]
    B --> C[End]
EOF

# Convert to SVG
uv run mermaid-to-svg-png.py diagrams/test.mmd --output docs/images/test.svg

# Convert to PNG
uv run mermaid-to-svg-png.py diagrams/test.mmd --output docs/images/test.png --format png
```

---

## Git Pre-commit Hook Setup

### 1. Copy Hook Script

```bash
cp ci-cd/pre-commit-hook.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### 2. Test Hook

```bash
# Create a test diagram
cat > diagrams/new-feature.mmd << 'EOF'
graph TD
    A[Feature] --> B[Implementation]
EOF

# Stage it
git add diagrams/new-feature.mmd

# Commit (hook will auto-convert)
git commit -m "chore: add new feature diagram"

# Check that docs/images/new-feature.png was created
ls -lh docs/images/new-feature.png
```

---

## GitHub Actions Setup

### 1. Create Workflow Directory

```bash
mkdir -p .github/workflows
```

### 2. Copy Workflow

```bash
cp ci-cd/github-actions-workflow.yml .github/workflows/mermaid-diagrams.yml
```

### 3. Enable Workflow

- Push to your repository
- Go to Actions tab
- Authorize workflow if needed

### 4. Test

```bash
# Create a feature branch
git checkout -b feature/test-mermaid

# Add diagram
cat > diagrams/api-flow.mmd << 'EOF'
sequenceDiagram
    User->>API: Request
    API->>Database: Query
    Database-->>API: Data
    API-->>User: Response
EOF

# Push branch
git add diagrams/api-flow.mmd
git commit -m "docs: add API flow diagram"
git push -u origin feature/test-mermaid

# Create PR
# Workflow will run and convert diagrams
# Check PR comments for results
```

---

## Project Structure

Recommended directory layout:

```
your-project/
├── .github/
│   └── workflows/
│       └── mermaid-diagrams.yml       # GitHub Actions workflow
├── .git/
│   └── hooks/
│       └── pre-commit                 # Git pre-commit hook
├── diagrams/                          # Source .mmd files
│   ├── architecture.mmd
│   ├── api-flow.mmd
│   ├── database-schema.mmd
│   └── deployment.mmd
├── docs/
│   └── images/                        # Generated PNG images
│       ├── architecture.png
│       ├── api-flow.png
│       ├── database-schema.png
│       └── deployment.png
├── scripts/
│   └── mermaid-to-svg-png.py         # CLI tool
└── README.md                          # Document with mermaid diagrams
```

---

## CLI Usage Reference

### Single File Conversion

```bash
# Convert to SVG (default)
uv run mermaid-to-svg-png.py diagrams/architecture.mmd

# Convert to PNG
uv run mermaid-to-svg-png.py diagrams/architecture.mmd \
  --format png \
  --output docs/images/architecture.png

# Custom dimensions
uv run mermaid-to-svg-png.py diagrams/flowchart.mmd \
  --width 1400 \
  --height 900 \
  --output diagram.png \
  --format png

# Custom theme
uv run mermaid-to-svg-png.py diagrams/dark-diagram.mmd \
  --theme dark \
  --format png \
  --output diagram-dark.png
```

### Batch Conversion

```bash
# Convert all diagrams in directory
uv run mermaid-to-svg-png.py diagrams \
  --batch \
  --format png \
  --output docs/images

# Batch with custom options
uv run mermaid-to-svg-png.py diagrams \
  --batch \
  --format png \
  --width 1200 \
  --height 800 \
  --output docs/images

# Skip existing files
uv run mermaid-to-svg-png.py diagrams \
  --batch \
  --no-overwrite \
  --output docs/images
```

### Watch Mode

```bash
# Auto-convert on changes (useful during development)
uv run mermaid-to-svg-png.py diagrams \
  --watch \
  --output docs/images
```

### Validation Only

```bash
# Check syntax without converting
uv run mermaid-to-svg-png.py diagrams/diagram.mmd --validate

# Validate all diagrams
uv run mermaid-to-svg-png.py diagrams --batch --validate
```

### CI/CD Usage

```bash
# Output as JSON for CI/CD parsing
uv run mermaid-to-svg-png.py diagrams \
  --batch \
  --format png \
  --json \
  --quiet > results.json

# Check exit code
echo $?  # 0 = success, 1 = failure
```

---

## Troubleshooting

### Issue: "Playwright: chromium not found"

**Solution**: Ensure dependencies are synced with uv

```bash
uv sync
uv run playwright install chromium
```

### Issue: "Permission denied" on pre-commit hook

**Solution**: Make hook executable

```bash
chmod +x .git/hooks/pre-commit
```

### Issue: PNG rendering is slow

**Solution**: Reduce image dimensions or use SVG format

```bash
# Use SVG instead
uv run mermaid-to-svg-png.py diagrams \
  --format svg \
  --batch

# Or reduce PNG size
uv run mermaid-to-svg-png.py diagrams \
  --format png \
  --width 800 \
  --height 600
```

### Issue: GitHub Actions timeout

**Solution**: Check diagram complexity and split into smaller diagrams

```bash
# List large diagrams
find diagrams -name "*.mmd" -size +50k

# Split into smaller files
# Large diagram with 100+ nodes should be split
```

---

## Advanced: Continuous Deployment

### Auto-commit diagram changes

```yaml
# In .github/workflows/mermaid-diagrams.yml
- name: Auto-commit diagrams
  if: github.event_name == 'push'
  uses: stefanzweifel/git-auto-commit-action@v4
  with:
    commit_message: "chore: auto-generated diagrams [skip ci]"
    commit_options: '--no-verify'
    file_pattern: 'docs/images/*'
```

### Publish to GitHub Pages

```yaml
- name: Publish diagrams
  uses: peaceiris/actions-gh-pages@v3
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: ./docs
```

---

## Next Steps

1. **Create diagrams**: Add .mmd files to `diagrams/` folder
2. **Setup automation**: Configure pre-commit hook or GitHub Actions
3. **Document changes**: Reference generated images in README.md
4. **Version control**: Commit both .mmd and generated images
5. **Review regularly**: Keep diagrams updated with code

---

**Version**: 5.0.0 | **Last Updated**: 2025-11-20
