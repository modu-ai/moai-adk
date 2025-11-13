---
name: moai-readme-expert
version: 4.0.0
status: production
description: |
  README.md generation specialist with dynamic project analysis capabilities. 
  Master professional README creation, specialized templates, and automated 
  documentation generation. Create comprehensive, engaging README files that 
  showcase projects effectively and follow best practices.
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "WebFetch", "WebSearch"]
tags: ["readme", "documentation", "markdown", "project-templates", "automated-generation"]
---

# README.md Expert

## Level 1: Quick Reference

### Core Capabilities
- **Professional Templates**: Industry-standard README structures
- **Dynamic Analysis**: Automatic project detection and analysis
- **Multiple Languages**: Support for various programming projects
- **Visual Elements**: Badges, diagrams, and structured formatting
- **Best Practices**: SEO optimization and engagement techniques
- **Automated Generation**: Script-based README creation

### Quick Setup Examples

```python
# Basic README generation
from readme_expert import ReadmeGenerator

generator = ReadmeGenerator()
readme_content = generator.generate_basic_readme(
    project_name="My Awesome Project",
    description="A revolutionary application that changes everything",
    author="Your Name",
    license="MIT"
)

# Save to file
with open("README.md", "w") as f:
    f.write(readme_content)
```

```python
# Advanced README with automatic project analysis
generator = ReadmeGenerator()
readme_content = generator.generate_comprehensive_readme(
    project_path="./",
    include_sections=[
        "installation", "usage", "contributing", "license", "changelog"
    ],
    style="professional",
    include_badges=True,
    include_diagrams=True
)
```

## Level 2: Core Patterns

### Template Engine Architecture

```python
class ReadmeTemplate:
    def __init__(self):
        self.templates = self._load_templates()
        self.section_generators = {
            'installation': self._generate_installation,
            'usage': self._generate_usage,
            'contributing': self._generate_contributing,
            'license': self._generate_license,
            'changelog': self._generate_changelog,
            'api': self._generate_api_docs,
            'testing': self._generate_testing,
            'deployment': self._generate_deployment
        }
    
    def generate_readme(self, project_info: Dict[str, Any], 
                        template_type: str = 'professional',
                        sections: List[str] = None) -> str:
        """Generate README from template"""
        
        template = self.templates.get(template_type, self.templates['professional'])
        
        # Generate badges
        badges = self._generate_badges(project_info)
        
        # Generate sections
        generated_sections = {}
        if sections:
            for section in sections:
                if section in self.section_generators:
                    generated_sections[f"{section}_section"] = self.section_generators[section](project_info)
        
        # Fill template
        readme_content = template.format(
            project_name=project_info.get('name', 'Project Name'),
            description=project_info.get('description', 'Project description'),
            badges=badges,
            features_section=generated_sections.get('features', ''),
            installation_section=generated_sections.get('installation', ''),
            usage_section=generated_sections.get('usage', ''),
            api_section=generated_sections.get('api', ''),
            testing_section=generated_sections.get('testing', ''),
            deployment_section=generated_sections.get('deployment', ''),
            contributing_section=generated_sections.get('contributing', ''),
            changelog_section=generated_sections.get('changelog', ''),
            license_section=generated_sections.get('license', '')
        )
        
        return readme_content
```

### Project Analysis System

```python
class ProjectAnalyzer:
    def __init__(self):
        self.indicators = {
            'package.json': self._analyze_nodejs_project,
            'requirements.txt': self._analyze_python_project,
            'Cargo.toml': self._analyze_rust_project,
            'go.mod': self._analyze_go_project,
            'pom.xml': self._analyze_java_project
        }
    
    def analyze_project(self, project_path: str) -> Dict[str, Any]:
        """Analyze project and extract information"""
        
        project_info = {
            'path': project_path,
            'name': self._extract_project_name(project_path),
            'description': self._extract_description(project_path),
            'type': 'unknown',
            'dependencies': [],
            'version': '1.0.0',
            'license': None,
            'repository': None,
            'author': None
        }
        
        # Detect project type
        for indicator_file, analyzer in self.indicators.items():
            file_path = os.path.join(project_path, indicator_file)
            if os.path.exists(file_path):
                specific_info = analyzer(file_path)
                project_info.update(specific_info)
                break
        
        return project_info
```

## Level 3: Advanced Implementation

### Specialized Domain Templates

#### API Service Template
```python
def _api_service_template(self, project_info: Dict[str, Any]) -> str:
    return f"""
# {project_info.get('name', 'API Service')}

{self._generate_api_badges(project_info)}

## Features

- RESTful API design
- JWT authentication
- Rate limiting
- Comprehensive error handling

## API Endpoints

### Authentication

#### POST /api/auth/login
Authenticate user and return JWT token.

**Request Body:**
```json
{{
  "email": "user@example.com",
  "password": "password123"
}}
```

## Installation

```bash
# Clone the repository
git clone {project_info.get('repository_url', '')}

# Install dependencies
{self._get_install_command(project_info)}

# Set up environment variables
cp .env.example .env

# Run database migrations
{self._get_migration_command(project_info)}

# Start the server
{self._get_start_command(project_info)}
```
    """
```

#### Web Application Template
```python
def _web_app_template(self, project_info: Dict[str, Any]) -> str:
    return f"""
# {project_info.get('name', 'Web Application')}

{self._generate_web_app_badges(project_info)}

## Features

- 🚀 **Modern UI/UX**: Built with latest frontend technologies
- 📱 **Responsive Design**: Works on all devices
- 🔐 **Authentication**: Secure user authentication
- ⚡ **Performance**: Optimized for speed

## Technology Stack

### Frontend
- **Framework**: {project_info.get('frontend_framework', 'React')}
- **Styling**: {project_info.get('styling', 'Tailwind CSS')}
- **State Management**: {project_info.get('state_management', 'Redux Toolkit')}

### Backend
- **Runtime**: {project_info.get('backend_runtime', 'Node.js')}
- **Framework**: {project_info.get('backend_framework', 'Express.js')}
- **Database**: {project_info.get('database', 'PostgreSQL')}

## Quick Start

```bash
# Clone and install
git clone {project_info.get('repository_url', '')}
cd {project_info.get('name', 'my-app')}
npm install

# Set up environment
cp .env.example .env
npm run db:setup

# Start development
npm run dev
```
    """
```

### Automated README Generator

```python
class ReadmeGenerator:
    def __init__(self):
        self.template_engine = ReadmeTemplate()
        self.project_analyzer = ProjectAnalyzer()
    
    def generate_comprehensive_readme(self, project_path: str = "./",
                                    template_type: str = 'professional',
                                    sections: List[str] = None,
                                    **overrides) -> str:
        """Generate a comprehensive README by analyzing the project"""
        
        # Analyze project
        project_info = self.project_analyzer.analyze_project(project_path)
        
        # Apply overrides
        project_info.update(overrides)
        
        # Set default sections
        if sections is None:
            sections = [
                'installation', 'usage', 'api', 'testing', 
                'deployment', 'contributing', 'license'
            ]
        
        # Generate README
        readme_content = self.template_engine.generate_readme(
            project_info=project_info,
            template_type=template_type,
            sections=sections
        )
        
        return readme_content
    
    def validate_readme(self, readme_path: str = "README.md") -> Dict[str, Any]:
        """Validate README content for best practices"""
        
        if not os.path.exists(readme_path):
            return {'valid': False, 'errors': ['README.md file not found']}
        
        with open(readme_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        validation_results = {
            'valid': True,
            'warnings': [],
            'suggestions': [],
            'stats': {
                'character_count': len(content),
                'line_count': len(content.split('\n')),
                'word_count': len(content.split())
            }
        }
        
        # Check for required sections
        required_sections = ['installation', 'usage']
        for section in required_sections:
            if f"## {section.title()}" not in content:
                validation_results['warnings'].append(f"Missing '{section}' section")
        
        # Check for badges
        if '![' not in content:
            validation_results['suggestions'].append("Consider adding badges for build status, license, etc.")
        
        # Check for code examples
        if '```' not in content:
            validation_results['suggestions'].append("Add code examples to demonstrate usage")
        
        return validation_results
```

## Level 4: Reference & Integration

### Badge Generation System

```python
def _generate_badges(self, project_info: Dict[str, Any]) -> str:
    """Generate project badges"""
    badges = []
    
    # Build status badge
    if project_info.get('ci_platform') and project_info.get('repository'):
        ci_badge = self._get_ci_badge(project_info['ci_platform'], project_info['repository'])
        badges.append(ci_badge)
    
    # License badge
    if project_info.get('license'):
        license_badge = f"![License](https://img.shields.io/badge/license-{project_info['license'].lower()}-blue.svg)"
        badges.append(license_badge)
    
    # Version badge
    if project_info.get('version'):
        version_badge = f"![Version](https://img.shields.io/badge/version-{project_info['version']}-brightgreen.svg)"
        badges.append(version_badge)
    
    return " ".join(badges) if badges else ""

def _get_ci_badge(self, ci_platform: str, repository: str) -> str:
    """Get CI platform badge"""
    ci_badges = {
        'github': f"![Build Status](https://github.com/{repository}/workflows/CI/badge.svg)",
        'travis': f"![Build Status](https://travis-ci.org/{repository}.svg?branch=main)",
        'circleci': f"![Build Status](https://circleci.com/gh/{repository}.svg?style=shield)",
        'gitlab': f"![Build Status](https://gitlab.com/{repository}/badges/main/pipeline.svg)"
    }
    return ci_badges.get(ci_platform, "")
```

### Installation Command Generator

```python
def _generate_installation(self, project_info: Dict[str, Any]) -> str:
    """Generate installation section"""
    project_type = project_info.get('type', 'nodejs')
    
    install_commands = {
        'nodejs': """
```bash
# Using npm
npm install {package_name}

# Using yarn
yarn add {package_name}
```
        """.strip(),
        
        'python': """
```bash
# Using pip
pip install {package_name}

# Using conda
conda install {package_name}
```
        """.strip(),
        
        'rust': """
```bash
# Using Cargo
cargo install {package_name}

# Or add to Cargo.toml
[dependencies]
{package_name} = "{version}"
```
        """.strip(),
        
        'go': """
```bash
# Using go get
go get {package_path}

# Or add to go.mod
require {package_path} {version}
```
        """.strip()
    }
    
    template = install_commands.get(project_type, install_commands['nodejs'])
    
    return template.format(
        package_name=project_info.get('package_name', project_info.get('name', '')),
        package_path=project_info.get('package_path', ''),
        version=project_info.get('version', 'latest')
    )
```

## Related Skills

- **moai-document-processing**: Document generation and formatting
- **moai-domain-testing**: Documentation testing and validation
- **moai-alfred-workflow**: README generation automation
- **moai-essentials-refactor**: Content optimization and refactoring

## Quick Start Checklist

- [ ] Analyze project structure and dependencies
- [ ] Choose appropriate README template
- [ ] Generate badges for project status
- [ ] Include installation and usage instructions
- [ ] Add API documentation if applicable
- [ ] Include testing and deployment guides
- [ ] Add contributing guidelines
- [ ] Validate README for best practices

## README Best Practices

1. **Clear Title**: Make the project name prominent and descriptive
2. **Badges**: Add relevant badges for status and information
3. **Table of Contents**: Help users navigate long READMEs
4. **Installation**: Provide clear, copy-pasteable installation instructions
5. **Usage Examples**: Include practical examples that work out-of-the-box
6. **Screenshots**: Add visual elements for better understanding
7. **Contributing**: Encourage contributions with clear guidelines
8. **License**: Clearly state the project license
9. **Links**: Include links to documentation, issues, and discussions
10. **Regular Updates**: Keep README updated with project changes

---

**README.md Expert** - Create professional, comprehensive README files that effectively showcase your projects and follow industry best practices.
