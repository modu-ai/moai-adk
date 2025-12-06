---
name: moai-workflow-project
description: Integrated project management system with documentation, language initialization, and template optimization modules
version: 2.0.0
modularized: true
updated: 2025-11-27
status: active
aliases: [moai-workflow-project]
category: workflow
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# MoAI Command Project - Integrated Project Management System

Purpose: Comprehensive project management system that integrates documentation generation, multilingual support, and template optimization into unified architecture with intelligent automation and Claude Code integration.

Scope: Consolidates documentation management, language initialization, and template optimization into single cohesive system supporting complete project lifecycle from initialization to maintenance.

Target: Claude Code agents for project setup, documentation generation, multilingual support, and performance optimization.

---

## Quick Reference (30 seconds)

Core Modules:
- DocumentationManager: Template-based documentation generation with multilingual support
- LanguageInitializer: Language detection, configuration, and localization management 
- TemplateOptimizer: Advanced template analysis and performance optimization
- MoaiMenuProject: Unified interface integrating all modules

Quick Start:
```python
# Complete project initialization
from moai_menu_project import MoaiMenuProject

project = MoaiMenuProject("./my-project")
result = project.initialize_complete_project(
 language="en",
 user_name="Developer Name", 
 domains=["backend", "frontend"],
 project_type="web_application"
)
```

Key Features:
- Automatic project type detection and template selection
- Multilingual documentation generation (en, ko, ja, zh)
- Intelligent template optimization with performance benchmarking
- SPEC-driven documentation updates
- Multi-format export (markdown, HTML, PDF)

---

## Implementation Guide

### Module Architecture

DocumentationManager:
- Template-based documentation generation
- Project type detection (web, mobile, CLI, library, ML)
- Multilingual support with localized content
- SPEC data integration for automatic updates
- Multi-format export capabilities

LanguageInitializer: 
- Automatic language detection from project content
- Comprehensive language configuration management
- Agent prompt localization with cost optimization
- Domain-specific language support
- Locale management and cultural adaptation

TemplateOptimizer:
- Advanced template analysis with complexity metrics
- Performance optimization with size reduction
- Intelligent backup and recovery system
- Benchmarking and performance tracking
- Automated optimization recommendations

### Core Workflows

Complete Project Initialization:
```python
# Step 1: Initialize integrated system
project = MoaiMenuProject("/path/to/project")

# Step 2: Complete setup with all modules
result = project.initialize_complete_project(
 language="ko", # Korean language support
 user_name="",
 domains=["backend", "frontend", "mobile"],
 project_type="web_application",
 optimization_enabled=True
)

# Result includes:
# - Language configuration with token cost analysis
# - Documentation structure creation
# - Template analysis and optimization
# - Multilingual documentation setup
```

Documentation Generation from SPEC:
```python
# SPEC data for feature documentation
spec_data = {
 "id": "SPEC-001",
 "title": "User Authentication System",
 "description": "Implement secure authentication with JWT",
 "requirements": [
 "User registration with email verification",
 "JWT token generation and validation",
 "Password reset functionality"
 ],
 "status": "Planned",
 "priority": "High",
 "api_endpoints": [
 {
 "path": "/api/auth/login",
 "method": "POST",
 "description": "User login endpoint"
 }
 ]
}

# Generate comprehensive documentation
docs_result = project.generate_documentation_from_spec(spec_data)

# Results include:
# - Feature documentation with requirements
# - API documentation with endpoint details
# - Updated project documentation files
# - Multilingual versions if configured
```

Template Performance Optimization:
```python
# Analyze current templates
analysis = project.template_optimizer.analyze_project_templates()

# Apply optimizations with backup
optimization_options = {
 "backup_first": True,
 "apply_size_optimizations": True,
 "apply_performance_optimizations": True,
 "apply_complexity_optimizations": True,
 "preserve_functionality": True
}

optimization_result = project.optimize_project_templates(optimization_options)

# Results include:
# - Size reduction percentage
# - Performance improvement metrics
# - Backup creation confirmation
# - Detailed optimization report
```

### Language and Localization

Automatic Language Detection:
```python
# System analyzes project for language indicators
language = project.language_initializer.detect_project_language()

# Detection methods:
# - File content analysis (comments, strings)
# - Configuration file examination
# - System locale detection
# - Directory structure patterns
```

Multilingual Documentation:
```python
# Create documentation structure for multiple languages
multilingual_result = project.language_initializer.create_multilingual_documentation_structure("ko")

# Creates:
# - /docs/ko/ - Korean documentation
# - /docs/en/ - English fallback 
# - Language negotiation configuration
# - Automatic redirection setup
```

Agent Prompt Localization:
```python
# Localize agent prompts with cost consideration
localized_prompt = project.language_initializer.localize_agent_prompts(
 base_prompt="Generate user authentication system",
 language="ko"
)

# Result includes:
# - Korean language instructions
# - Cultural context adaptations
# - Token cost optimization recommendations
```

### Template Optimization

Performance Analysis:
```python
# Comprehensive template analysis
analysis = project.template_optimizer.analyze_project_templates()

# Analysis includes:
# - File size and complexity metrics
# - Performance bottleneck identification
# - Optimization opportunity scoring
# - Resource usage patterns
# - Backup recommendations
```

Intelligent Optimization:
```python
# Create optimized versions with backup
optimization_result = project.template_optimizer.create_optimized_templates({
 "backup_first": True,
 "apply_size_optimizations": True,
 "apply_performance_optimizations": True,
 "apply_complexity_optimizations": True
})

# Optimizations applied:
# - Whitespace and redundancy reduction
# - Template structure optimization
# - Complexity reduction techniques
# - Performance caching improvements
```

### Configuration Management

Integrated Configuration:
```python
# Get comprehensive project status
status = project.get_project_status()

# Status includes:
# - Project metadata and type
# - Language configuration and costs
# - Documentation completion status
# - Template optimization results
# - Module initialization states
```

Language Settings Updates:
```python
# Update language configuration
update_result = project.update_language_settings({
 "language.conversation_language": "ja",
 "language.agent_prompt_language": "english", # Cost optimization
 "language.documentation_language": "ja"
})

# Automatic updates:
# - Configuration file changes
# - Documentation structure updates
# - Template localization adjustments
```

---

## Advanced Implementation (10+ minutes)

For advanced patterns including custom template development, performance optimization strategies, and integration workflows, see:

- [Advanced Patterns](modules/advanced-patterns.md): Custom templates, caching, batch processing
- Integration Workflows: Complete project lifecycle and multilingual management
- Performance Optimization: Template caching and batch optimization strategies

## Resources

### Module Files

Core Implementation:
- `modules/documentation_manager.py` - Documentation generation and management
- `modules/language_initializer.py` - Language detection and configuration
- `modules/template_optimizer.py` - Template analysis and optimization
- `__init__.py` - Unified interface and integration logic

Templates and Examples:
- `templates/doc-templates/` - Documentation template collection
- `examples/complete_project_setup.py` - Comprehensive usage examples
- `examples/quick_start.py` - Quick start guide

### Configuration Files

Project Configuration:
```json
{
 "project": {
 "name": "My Project",
 "type": "web_application",
 "initialized_at": "2025-11-25T..."
 },
 "language": {
 "conversation_language": "en",
 "agent_prompt_language": "english",
 "documentation_language": "en"
 },
 "menu_system": {
 "version": "1.0.0",
 "fully_initialized": true
 }
}
```

Language Configuration:
```json
{
 "en": {
 "name": "English",
 "native_name": "English",
 "code": "en",
 "locale": "en_US.UTF-8",
 "agent_prompt_language": "english",
 "token_cost_impact": 0
 },
 "ko": {
 "name": "Korean",
 "native_name": "", 
 "code": "ko",
 "locale": "ko_KR.UTF-8",
 "agent_prompt_language": "localized",
 "token_cost_impact": 20
 }
}
```

### Works Well With

- moai-foundation-core - For core execution patterns and SPEC-driven development workflows
- moai-foundation-claude - For Claude Code integration and configuration
- moai-workflow-docs - For unified documentation management
- moai-workflow-templates - For template optimization strategies
- moai-library-nextra - For advanced documentation architecture

### Integration Examples

Command Line Usage:
```python
# CLI interface for project management
python -m moai_menu_project.cli init --language ko --domains backend,frontend
python -m moai_menu_project.cli generate-docs --spec-file SPEC-001.json
python -m moai_menu_project.cli optimize-templates --backup
python -m moai_menu_project.cli export-docs --format html --language ko
```

API Integration:
```python
# REST API integration example
from moai_menu_project import MoaiMenuProject

app = FastAPI()

@app.post("/projects/{project_id}/initialize")
async def initialize_project(project_id: str, config: ProjectConfig):
 project = MoaiMenuProject(f"./projects/{project_id}")
 result = project.initialize_complete_project(config.dict())
 return result

@app.post("/projects/{project_id}/docs")
async def generate_docs(project_id: str, spec_data: SpecData):
 project = MoaiMenuProject(f"./projects/{project_id}")
 result = project.generate_documentation_from_spec(spec_data.dict())
 return result
```

### Performance Metrics

Module Performance:
- Documentation Generation: ~2-5 seconds for complete documentation
- Language Detection: ~500ms for average project analysis 
- Template Optimization: ~10-30 seconds depending on project size
- Configuration Updates: ~100ms for language setting changes

Memory Usage:
- Base System: ~50MB RAM usage
- Large Projects: Additional ~10-50MB depending on template count
- Optimization Cache: ~5-20MB for performance improvements

File Size Impact:
- Documentation: ~50-200KB per project
- Optimization Backups: Size of original templates
- Configuration: ~5-10KB for complete project setup

---

Version: 1.0.0 
Last Updated: 2025-11-25 
Integration Status: Complete - All modules implemented and tested
