---
name: moai-readme-expert
description: README.md generation specialist with dynamic project analysis and professional templates
version: 1.0.0
modularized: false
tags:
  - enterprise
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: readme, expert, moai  


# README.md Expert

## Quick Reference

### Core Capabilities
- Professional README templates for all project types
- Dynamic project analysis and detection
- Multi-language support (Python, JavaScript, Rust, Go, Java)
- Badge and visual element generation
- Best practices for SEO and engagement

### Immediate Usage

```python
from readme_expert import ReadmeGenerator

# Auto-generate README from project analysis
generator = ReadmeGenerator()
readme = generator.generate_comprehensive_readme(
    project_path="./",
    template_type="professional",
    include_badges=True
)

# Save to file
with open("README.md", "w") as f:
    f.write(readme)
```

### Quick Templates

**Basic README**:
```markdown
# Project Name

Brief description of your project.

## Installation

```bash
npm install package-name
```

## Usage

```javascript
import { feature } from 'package-name';
feature();
```

## License

MIT
```


## Implementation Guide

### Template Engine Architecture

```python
class ReadmeTemplate:
    def __init__(self):
        self.section_generators = {
            'installation': self._generate_installation,
            'usage': self._generate_usage,
            'contributing': self._generate_contributing,
            'license': self._generate_license,
            'api': self._generate_api_docs
        }
    
    def generate_readme(self, project_info: Dict, sections: List[str]) -> str:
        """Generate README from project information"""
        
        # Generate badges
        badges = self._generate_badges(project_info)
        
        # Generate selected sections
        content_sections = {}
        for section in sections:
            if section in self.section_generators:
                content_sections[section] = self.section_generators[section](project_info)
        
        # Assemble final README
        readme = f"""# {project_info['name']}

{badges}

{project_info['description']}

{chr(10).join(content_sections.values())}
"""
        return readme
```

### Project Analysis System

```python
class ProjectAnalyzer:
    def analyze_project(self, project_path: str) -> Dict[str, Any]:
        """Analyze project structure and extract metadata"""
        
        project_info = {
            'name': self._extract_project_name(project_path),
            'description': self._extract_description(project_path),
            'type': 'unknown',
            'dependencies': [],
            'version': '1.0.0'
        }
        
        # Detect project type
        if os.path.exists(f"{project_path}/package.json"):
            project_info.update(self._analyze_nodejs_project(project_path))
        elif os.path.exists(f"{project_path}/requirements.txt"):
            project_info.update(self._analyze_python_project(project_path))
        elif os.path.exists(f"{project_path}/Cargo.toml"):
            project_info.update(self._analyze_rust_project(project_path))
        
        return project_info
```

### Badge Generation

```python
def _generate_badges(self, project_info: Dict) -> str:
    """Generate project status badges"""
    
    badges = []
    
    # Build status
    if project_info.get('ci_platform'):
        badges.append(f"![Build]({self._get_ci_badge(project_info)})")
    
    # License
    if project_info.get('license'):
        badges.append(f"![License](https://img.shields.io/badge/license-{project_info['license']}-blue.svg)")
    
    # Version
    if project_info.get('version'):
        badges.append(f"![Version](https://img.shields.io/badge/version-{project_info['version']}-brightgreen.svg)")
    
    return " ".join(badges)
```

### Installation Command Generator

```python
def _generate_installation(self, project_info: Dict) -> str:
    """Generate installation instructions"""
    
    project_type = project_info.get('type', 'nodejs')
    
    templates = {
        'nodejs': """
## Installation

```bash
# Using npm
npm install {package_name}

# Using yarn
yarn add {package_name}
```
        """,
        'python': """
## Installation

```bash
# Using pip
pip install {package_name}

# Using conda
conda install {package_name}
```
        """,
        'rust': """
## Installation

```bash
# Using Cargo
cargo install {package_name}

# Or add to Cargo.toml
[dependencies]
{package_name} = "{version}"
```
        """
    }
    
    template = templates.get(project_type, templates['nodejs'])
    return template.format(
        package_name=project_info.get('name', 'package'),
        version=project_info.get('version', 'latest')
    )
```

### Validation System

```python
def validate_readme(self, readme_path: str = "README.md") -> Dict:
    """Validate README against best practices"""
    
    with open(readme_path, 'r') as f:
        content = f.read()
    
    validation = {
        'valid': True,
        'warnings': [],
        'suggestions': []
    }
    
    # Check required sections
    if "## Installation" not in content:
        validation['warnings'].append("Missing 'Installation' section")
    
    if "## Usage" not in content:
        validation['warnings'].append("Missing 'Usage' section")
    
    # Check for code examples
    if '```' not in content:
        validation['suggestions'].append("Add code examples")
    
    # Check for badges
    if '![' not in content:
        validation['suggestions'].append("Consider adding badges")
    
    validation['valid'] = len(validation['warnings']) == 0
    return validation
```


## Advanced Patterns

### Specialized Templates

#### API Service Template

```python
def _api_service_template(self, project_info: Dict) -> str:
    return f"""
# {project_info['name']} API

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

**Request**:
```json
{{
  "email": "user@example.com",
  "password": "password123"
}}
```

**Response**:
```json
{{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {{"id": 1, "email": "user@example.com"}}
}}
```

## Quick Start

```bash
# Clone repository
git clone {project_info.get('repository_url', '')}

# Install dependencies
{self._get_install_command(project_info)}

# Configure environment
cp .env.example .env

# Start server
{self._get_start_command(project_info)}
```
    """
```

#### Web Application Template

```python
def _web_app_template(self, project_info: Dict) -> str:
    return f"""
# {project_info['name']}

{self._generate_web_app_badges(project_info)}

## Features

- Modern UI/UX with {project_info.get('frontend_framework', 'React')}
- Responsive design
- Secure authentication
- Performance optimized

## Technology Stack

### Frontend
- Framework: {project_info.get('frontend_framework', 'React')}
- Styling: {project_info.get('styling', 'Tailwind CSS')}
- State: {project_info.get('state_management', 'Redux')}

### Backend
- Runtime: {project_info.get('backend_runtime', 'Node.js')}
- Framework: {project_info.get('backend_framework', 'Express')}
- Database: {project_info.get('database', 'PostgreSQL')}

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```
    """
```

### Best Practices Checklist

```
README Best Practices:

✓ Clear, descriptive title
✓ Status badges (build, coverage, license)
✓ Table of contents (for long READMEs)
✓ Installation instructions (copy-paste ready)
✓ Usage examples (working code)
✓ Screenshots/demos (visual elements)
✓ API documentation (if applicable)
✓ Contributing guidelines
✓ License information
✓ Contact/support information
```

### Integration with MoAI Skills

```python
# Works well with:
Skill("moai-docs-generation")       # Documentation generation
Skill("moai-domain-testing")        # Testing documentation
Skill("moai-alfred-workflow")       # Workflow automation
Skill("moai-essentials-refactor")   # Content optimization
```


**End of Skill** | Create professional, comprehensive README files
