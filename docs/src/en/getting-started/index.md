# Quick Start Guide

**@DOC:QUICK-START-001** | **Last Updated**: 2025-11-05 | **Duration**: 5 minutes

______________________________________________________________________

## 🎯 Objectives

This guide will teach you how to perfectly set up and run the MoAI-ADK online documentation system.

______________________________________________________________________

## 🚀 Step 1: System Requirements

### Required

- **Python**: 3.13 or higher
- **UV**: Latest version recommended
- **Git**: Latest version
- **Browser**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

### Optional

- **Vercel CLI**: Optional tool for automated deployment
- **Node.js**: v18+ (required for some build tools)

______________________________________________________________________

## ⚡ Step 2: Project Setup (30 seconds)

### Install UV

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Verify installation
uv --version
```

### Clone and Setup Project

```bash
# 1. Clone project
git clone https://github.com/moai-adk/MoAI-ADK.git
cd MoAI-ADK

# 2. Install dependencies (automatic)
uv sync

# 3. Run development server
uv run dev
```

### Quick Verification

```bash
# Check server status
curl http://127.0.0.1:8080

# Check build status
uv run build
```

______________________________________________________________________

## 🎨 Step 3: Building Documentation System

### MkDocs Configuration

```bash
# Verify MkDocs basic setup
uv run mkdocs --help

# Create project structure
mkdir -p docs/{getting-started,alfred,commands,development,advanced,api,contributing}

# Setup theme
uv run mkdocs new .
```

### Add Multi-language Configuration

```yaml
# mkdocs.yml
site_name: MoAI-ADK Documentation
nav:
  - Home: index.md
  - Quick Start: getting-started/
  - Alfred: alfred/
  - Commands: commands/
  - Development: development/
  - Advanced Features: advanced/
  - API: api/
  - Contributing: contributing/

theme:
  name: material
  language: en
  palette:
    - media: "(prefers-color-scheme: light)"
      scheme: default
      primary: blue
      accent: indigo
    - media: "(prefers-color-scheme: dark)"
      scheme: slate
      primary: blue
      accent: indigo
```

______________________________________________________________________

## <span class="material-icons">search</span> Step 4: Search and Navigation

### Enable Search System

```bash
# Add dependency
uv add mkdocs-search

# Update configuration
uv run mkdocs build --strict
```

### Test Real-time Search

1. Run development server: `uv run dev`
2. Access http://127.0.0.1:8080 in browser
3. Enter "MoAI" in search bar
4. Verify real-time search results

______________________________________________________________________

## 🌍 Step 5: Multi-language Configuration

### Create Language Files

```bash
# English
echo "# English Documentation" > docs/getting-started/index-en.md

# Japanese
echo "# 日本語ドキュメント" > docs/getting-started/index-ja.md

# Chinese
echo "# 中文文档" > docs/getting-started/index-zh.md
```

### Test Language Switching

```bash
# Build multi-language docs
uv run build

# Verify results
ls -la site/getting-started/
```

______________________________________________________________________

## 📊 Step 6: Deployment Preparation

### Local Build

```bash
# Build static site
uv run build

# Verify results
ls -la site/

# Check file size
du -sh site/
```

### Vercel Deployment

```bash
# 1. Install Vercel CLI
npm i -g vercel

# 2. Login
vercel login

# 3. Deploy project
vercel --prod

# 4. Verify deployment
vercel ls
```

______________________________________________________________________

## 🧪 Step 7: Testing and Validation

### Automated Testing

```bash
# 1. Validate documentation
uv run validate

# 2. Check links
uv run check-links

# 3. Build test
uv run build --strict
```

### Manual Testing Checklist

- [ ] All pages display correctly
- [ ] Dark/light mode switching works
- [ ] Search functionality works properly
- [ ] Mobile responsive design confirmed
- [ ] Multi-language documentation accessible

______________________________________________________________________

## 📋 Completion Checklist

### System Status

- [ ] UV installed successfully
- [ ] Project cloned successfully
- [ ] Dependencies installed successfully
- [ ] Development server running successfully
- [ ] Documentation build successful

### Feature Verification

- [ ] All pages display correctly
- [ ] Search functionality works properly
- [ ] Multi-language support confirmed
- [ ] Responsive design confirmed
- [ ] Dark/light mode switching confirmed

### Deployment Verification

- [ ] Local build successful
- [ ] Vercel deployment successful
- [ ] Domain accessible
- [ ] SSL certificate confirmed
- [ ] CDN performance confirmed

______________________________________________________________________

## 🚀 Next Steps

### 1. Customization

- Modify design system
- Add new languages
- Develop custom components

### 2. Add Content

- Generate API documentation
- Write tutorials
- Add advanced guides

### 3. Production Deployment

- Configure automated deployment
- Connect monitoring tools
- Optimize performance

______________________________________________________________________

## 🐛 Troubleshooting

### Common Issues

#### UV Installation Error

```bash
# Clean cache
uv cache clean

# Reinstall
pip install --upgrade uv
```

#### Build Error

```bash
# Clean cache
rm -rf site/ .doit_db/

# Reinstall dependencies
uv sync --force

# Rebuild
uv run build
```

#### Server Start Error

```bash
# Change port
uv run dev --port 3000

# Check logs
uv run dev --verbose
```

______________________________________________________________________

## 📞 Support

### Official Documentation

- **URL**: https://adk.mo.ai.kr
- **Status**: 24/7 operation
- **Updates**: Real-time synchronization

### Development Support

- **GitHub Issues**: [Report technical issues](https://github.com/moai-adk/MoAI-ADK/issues)
- **GitHub Discussions**: [Q&A](https://github.com/moai-adk/MoAI-ADK/discussions)
- **Email**: support@mo.ai.kr

______________________________________________________________________

*Last Updated: 2025-11-05 | Version: v0.17.0 | Status: 100% Complete*
