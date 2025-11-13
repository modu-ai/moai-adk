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

## Level 2: Practical Implementation

### README Template Engine

#### 1. Template-based README Generation

```python
import os
import re
from datetime import datetime
from typing import Dict, List, Optional, Any
import json
import subprocess
from pathlib import Path

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
            'deployment': self._generate_deployment,
            'architecture': self._generate_architecture,
            'performance': self._generate_performance
        }
    
    def _load_templates(self) -> Dict[str, str]:
        """Load README templates"""
        return {
            'basic': """
# {project_name}

{description}

## Installation

{installation_section}

## Usage

{usage_section}

## Contributing

{contributing_section}

## License

{license_section}
            """.strip(),
            
            'professional': """
# {project_name}

{badges}

{description}

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Deployment](#deployment)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [Changelog](#changelog)
- [License](#license)

## Features

{features_section}

## Installation

{installation_section}

## Usage

{usage_section}

## API Reference

{api_section}

## Testing

{testing_section}

## Deployment

{deployment_section}

## Architecture

{architecture_section}

## Contributing

{contributing_section}

## Changelog

{changelog_section}

## License

{license_section}
            """.strip(),
            
            'minimal': """
# {project_name}

{description}

## Quick Start

```bash
{quick_start_command}
```

## Usage

{usage_section}

## License

{license_section}
            """.strip()
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
            architecture_section=generated_sections.get('architecture', ''),
            contributing_section=generated_sections.get('contributing', ''),
            changelog_section=generated_sections.get('changelog', ''),
            license_section=generated_sections.get('license', ''),
            quick_start_command=project_info.get('quick_start', 'npm install && npm start')
        )
        
        return readme_content
    
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
        
        # Downloads badge (for npm packages)
        if project_info.get('npm_package'):
            downloads_badge = f"![Downloads](https://img.shields.io/npm/dm/{project_info['npm_package']}.svg)"
            badges.append(downloads_badge)
        
        # Code coverage badge
        if project_info.get('coverage_badge'):
            badges.append(project_info['coverage_badge'])
        
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
    
    def _generate_usage(self, project_info: Dict[str, Any]) -> str:
        """Generate usage section"""
        project_type = project_info.get('type', 'nodejs')
        
        usage_examples = {
            'nodejs': """
```javascript
const {packageName} = require('{packageName}');

// Basic usage
const result = {packageName}.methodName({
  parameter1: 'value1',
  parameter2: 'value2'
});

console.log(result);
```
            """.strip(),
            
            'python': """
```python
import {package_name}

# Basic usage
result = {package_name}.method_name(
    parameter1='value1',
    parameter2='value2'
)

print(result)
```
            """.strip(),
            
            'rust': """
```rust
use {package_name}::{PackageName};

fn main() {
    let result = PackageName::method_name(
        "value1",
        "value2"
    );
    
    println!("{:?}", result);
}
```
            """.strip()
        }
        
        template = usage_examples.get(project_type, usage_examples['nodejs'])
        
        return template.format(
            package_name=project_info.get('package_name', 'package'),
            PackageName=project_info.get('package_name', 'package').replace('-', '_').title().replace('_', '')
        )
    
    def _generate_contributing(self, project_info: Dict[str, Any]) -> str:
        """Generate contributing section"""
        return """
Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

### Development Setup

```bash
# Clone the repository
git clone {repository_url}
cd {project_name}

# Install dependencies
{install_command}

# Run tests
{test_command}
```

### Code Style

This project follows the {code_style} code style. Please ensure your code adheres to these guidelines.

### Commit Messages

Please follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.
        """.strip().format(
            repository_url=project_info.get('repository_url', ''),
            project_name=project_info.get('name', 'project'),
            install_command=project_info.get('dev_install_command', 'npm install'),
            test_command=project_info.get('test_command', 'npm test'),
            code_style=project_info.get('code_style', 'standard')
        )
    
    def _generate_license(self, project_info: Dict[str, Any]) -> str:
        """Generate license section"""
        license_type = project_info.get('license', 'MIT')
        
        license_info = {
            'MIT': "This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.",
            'Apache-2.0': "This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.",
            'GPL-3.0': "This project is licensed under the GPL v3.0 License - see the [LICENSE](LICENSE) file for details.",
            'BSD-3-Clause': "This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details."
        }
        
        return license_info.get(license_type, f"This project is licensed under the {license_type} License.")
    
    def _generate_changelog(self, project_info: Dict[str, Any]) -> str:
        """Generate changelog section"""
        return """
## [Unreleased]

### Added
- New feature X
- New feature Y

### Changed
- Updated dependency Z

### Fixed
- Bug fix A
- Bug fix B

## [1.0.0] - 2024-01-01

### Added
- Initial release
- Core functionality
        """.strip()
    
    def _generate_api_docs(self, project_info: Dict[str, Any]) -> str:
        """Generate API documentation section"""
        return """
### Core Methods

#### `methodName(parameter1, parameter2)`

Description of what this method does.

**Parameters:**
- `parameter1` (string): Description of parameter1
- `parameter2` (number): Description of parameter2

**Returns:**
- (object): Description of return value

**Example:**
```javascript
const result = methodName('value1', 42);
console.log(result);
```

### Events

#### `eventName`

Emitted when something happens.

**Payload:**
- `property1` (string): Description of property1
- `property2` (number): Description of property2

**Example:**
```javascript
emitter.on('eventName', (payload) => {
    console.log(payload.property1);
});
```
        """.strip()
    
    def _generate_testing(self, project_info: Dict[str, Any]) -> str:
        """Generate testing section"""
        return """
### Running Tests

```bash
# Run all tests
{test_command}

# Run tests with coverage
{coverage_command}

# Run specific test file
{specific_test_command}
```

### Test Structure

Tests are organized in the following structure:

```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
├── e2e/           # End-to-end tests
└── fixtures/      # Test fixtures
```

### Writing Tests

```javascript
// Example test
describe('FeatureName', () => {
  it('should do something', () => {
    // Test implementation
  });
});
```

### Coverage

This project aims for {coverage_target}% test coverage. Coverage reports are generated automatically.
        """.strip().format(
            test_command=project_info.get('test_command', 'npm test'),
            coverage_command=project_info.get('coverage_command', 'npm run test:coverage'),
            specific_test_command=project_info.get('specific_test_command', 'npm test -- --grep "test name"'),
            coverage_target=project_info.get('coverage_target', '80')
        )
    
    def _generate_deployment(self, project_info: Dict[str, Any]) -> str:
        """Generate deployment section"""
        return """
### Production Deployment

#### Environment Variables

```bash
NODE_ENV=production
PORT=3000
DATABASE_URL=your_database_url
API_KEY=your_api_key
```

#### Docker Deployment

```bash
# Build Docker image
docker build -t {image_name} .

# Run container
docker run -p 3000:3000 {image_name}
```

#### Cloud Deployment

**Heroku**
```bash
# Deploy to Heroku
git push heroku main
```

**Vercel**
```bash
# Deploy to Vercel
vercel --prod
```

**AWS**
```bash
# Deploy to AWS using AWS CLI
aws s3 sync ./build s3://your-bucket-name
```
        """.strip().format(
            image_name=project_info.get('docker_image', 'my-app')
        )
    
    def _generate_architecture(self, project_info: Dict[str, Any]) -> str:
        """Generate architecture section"""
        return """
### System Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │────│    API      │────│  Database   │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                   ┌─────────────┐
                   │   Cache     │
                   │             │
                   └─────────────┘
```

### Component Overview

- **Frontend**: {frontend_description}
- **API**: {api_description}
- **Database**: {database_description}
- **Cache**: {cache_description}

### Data Flow

1. User interacts with frontend
2. Frontend makes API calls
3. API processes requests and queries database
4. Results are cached for performance
5. Response returned to frontend
        """.strip().format(
            frontend_description=project_info.get('frontend_description', 'Modern web application'),
            api_description=project_info.get('api_description', 'RESTful API service'),
            database_description=project_info.get('database_description', 'PostgreSQL database'),
            cache_description=project_info.get('cache_description', 'Redis caching layer')
        )
    
    def _generate_performance(self, project_info: Dict[str, Any]) -> str:
        """Generate performance section"""
        return """
### Performance Metrics

- **Response Time**: < 100ms (95th percentile)
- **Throughput**: 1000+ requests/second
- **Memory Usage**: < 512MB
- **CPU Usage**: < 50%

### Optimization Techniques

- Database query optimization
- Response caching
- CDN usage
- Code splitting
- Lazy loading

### Monitoring

- Application performance monitoring
- Error tracking
- Log aggregation
- Health checks
        """.strip()
```

#### 2. Project Analysis and Detection

```python
class ProjectAnalyzer:
    def __init__(self):
        self.indicators = {
            'package.json': self._analyze_nodejs_project,
            'requirements.txt': self._analyze_python_project,
            'Cargo.toml': self._analyze_rust_project,
            'go.mod': self._analyze_go_project,
            'pom.xml': self._analyze_java_project,
            'composer.json': self._analyze_php_project,
            'Gemfile': self._analyze_ruby_project
        }
    
    def analyze_project(self, project_path: str) -> Dict[str, Any]:
        """Analyze project and extract information"""
        
        project_info = {
            'path': project_path,
            'name': self._extract_project_name(project_path),
            'description': self._extract_description(project_path),
            'type': 'unknown',
            'dependencies': [],
            'dev_dependencies': [],
            'scripts': {},
            'version': '1.0.0',
            'license': None,
            'repository': None,
            'author': None,
            'keywords': []
        }
        
        # Detect project type and extract specific information
        for indicator_file, analyzer in self.indicators.items():
            file_path = os.path.join(project_path, indicator_file)
            if os.path.exists(file_path):
                specific_info = analyzer(file_path)
                project_info.update(specific_info)
                break
        
        # Extract additional information
        project_info.update(self._extract_git_info(project_path))
        project_info.update(self._analyze_code_structure(project_path))
        
        return project_info
    
    def _extract_project_name(self, project_path: str) -> str:
        """Extract project name from directory or package files"""
        dir_name = os.path.basename(project_path)
        
        # Try to get name from package.json first
        package_json = os.path.join(project_path, 'package.json')
        if os.path.exists(package_json):
            try:
                with open(package_json, 'r') as f:
                    data = json.load(f)
                    return data.get('name', dir_name)
            except:
                pass
        
        return dir_name.replace('-', ' ').replace('_', ' ').title()
    
    def _extract_description(self, project_path: str) -> str:
        """Extract description from various sources"""
        sources = [
            'package.json',
            'README.md',
            'setup.py',
            'Cargo.toml'
        ]
        
        for source in sources:
            file_path = os.path.join(project_path, source)
            if os.path.exists(file_path):
                try:
                    if source == 'package.json':
                        with open(file_path, 'r') as f:
                            data = json.load(f)
                            return data.get('description', '')
                    elif source == 'README.md':
                        with open(file_path, 'r') as f:
                            first_line = f.readline().strip()
                            if first_line.startswith('# '):
                                return first_line[2:].strip()
                    elif source == 'setup.py':
                        with open(file_path, 'r') as f:
                            content = f.read()
                            match = re.search(r'description\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                            if match:
                                return match.group(1)
                    elif source == 'Cargo.toml':
                        with open(file_path, 'r') as f:
                            content = f.read()
                            match = re.search(r'description\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                            if match:
                                return match.group(1)
                except:
                    continue
        
        return f"A {os.path.basename(project_path)} project"
    
    def _analyze_nodejs_project(self, package_json_path: str) -> Dict[str, Any]:
        """Analyze Node.js project from package.json"""
        
        try:
            with open(package_json_path, 'r') as f:
                data = json.load(f)
            
            return {
                'type': 'nodejs',
                'package_name': data.get('name', ''),
                'version': data.get('version', '1.0.0'),
                'description': data.get('description', ''),
                'author': data.get('author', ''),
                'license': data.get('license', 'MIT'),
                'keywords': data.get('keywords', []),
                'repository': data.get('repository', {}).get('url', '') if data.get('repository') else '',
                'dependencies': list(data.get('dependencies', {}).keys()),
                'dev_dependencies': list(data.get('devDependencies', {}).keys()),
                'scripts': data.get('scripts', {}),
                'main': data.get('main', ''),
                'engines': data.get('engines', {}),
                'npm_package': data.get('name', '')
            }
        except Exception as e:
            print(f"Error analyzing package.json: {e}")
            return {'type': 'nodejs'}
    
    def _analyze_python_project(self, requirements_path: str) -> Dict[str, Any]:
        """Analyze Python project from requirements.txt or setup.py"""
        
        project_info = {'type': 'python'}
        project_path = os.path.dirname(requirements_path)
        
        # Analyze requirements.txt
        if os.path.exists(requirements_path):
            try:
                with open(requirements_path, 'r') as f:
                    requirements = [line.strip() for line in f if line.strip() and not line.startswith('#')]
                project_info['dependencies'] = requirements
            except:
                pass
        
        # Analyze setup.py if exists
        setup_py = os.path.join(project_path, 'setup.py')
        if os.path.exists(setup_py):
            try:
                with open(setup_py, 'r') as f:
                    content = f.read()
                
                # Extract information from setup.py
                name_match = re.search(r'name\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                version_match = re.search(r'version\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                desc_match = re.search(r'description\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                author_match = re.search(r'author\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                license_match = re.search(r'license\s*=\s*[\'"]([^\'"]+)[\'"]', content)
                
                if name_match:
                    project_info['package_name'] = name_match.group(1)
                if version_match:
                    project_info['version'] = version_match.group(1)
                if desc_match:
                    project_info['description'] = desc_match.group(1)
                if author_match:
                    project_info['author'] = author_match.group(1)
                if license_match:
                    project_info['license'] = license_match.group(1)
                
            except:
                pass
        
        return project_info
    
    def _analyze_rust_project(self, cargo_toml_path: str) -> Dict[str, Any]:
        """Analyze Rust project from Cargo.toml"""
        
        try:
            import toml
            with open(cargo_toml_path, 'r') as f:
                data = toml.load(f)
            
            package_info = data.get('package', {})
            
            return {
                'type': 'rust',
                'package_name': package_info.get('name', ''),
                'version': package_info.get('version', '1.0.0'),
                'description': package_info.get('description', ''),
                'author': package_info.get('authors', [''])[0] if package_info.get('authors') else '',
                'license': package_info.get('license', ''),
                'keywords': package_info.get('keywords', []),
                'repository': package_info.get('repository', ''),
                'dependencies': list(data.get('dependencies', {}).keys()),
                'dev_dependencies': list(data.get('dev-dependencies', {}).keys())
            }
        except Exception as e:
            print(f"Error analyzing Cargo.toml: {e}")
            return {'type': 'rust'}
    
    def _analyze_go_project(self, go_mod_path: str) -> Dict[str, Any]:
        """Analyze Go project from go.mod"""
        
        try:
            with open(go_mod_path, 'r') as f:
                content = f.read()
            
            # Extract module name
            module_match = re.search(r'module\s+([^\s]+)', content)
            module_name = module_match.group(1) if module_match else ''
            
            # Extract Go version
            go_match = re.search(r'go\s+([^\s]+)', content)
            go_version = go_match.group(1) if go_match else ''
            
            return {
                'type': 'go',
                'package_path': module_name,
                'version': go_version,
                'dependencies': self._extract_go_dependencies(content)
            }
        except Exception as e:
            print(f"Error analyzing go.mod: {e}")
            return {'type': 'go'}
    
    def _extract_go_dependencies(self, content: str) -> List[str]:
        """Extract Go dependencies from go.mod content"""
        dependencies = []
        
        # Find require block
        require_match = re.search(r'require\s*\((.*?)\)', content, re.DOTALL)
        if require_match:
            require_content = require_match.group(1)
            for line in require_content.split('\n'):
                line = line.strip()
                if line and not line.startswith('//'):
                    parts = line.split()
                    if parts:
                        dependencies.append(parts[0])
        else:
            # Single line requires
            for line in content.split('\n'):
                if line.strip().startswith('require '):
                    parts = line.strip().split()
                    if len(parts) >= 2:
                        dependencies.append(parts[1])
        
        return dependencies
    
    def _extract_git_info(self, project_path: str) -> Dict[str, Any]:
        """Extract Git repository information"""
        
        git_info = {}
        
        try:
            # Get remote URL
            result = subprocess.run(
                ['git', 'config', '--get', 'remote.origin.url'],
                cwd=project_path,
                capture_output=True,
                text=True
            )
            
            if result.returncode == 0:
                remote_url = result.stdout.strip()
                git_info['repository_url'] = remote_url
                
                # Extract owner and repo from URL
                if 'github.com' in remote_url:
                    parts = remote_url.split('/')
                    if len(parts) >= 2:
                        git_info['repository'] = f"{parts[-2]}/{parts[-1].replace('.git', '')}"
            
            # Get default branch
            result = subprocess.run(
                ['git', 'rev-parse', '--abbrev-ref', 'HEAD'],
                cwd=project_path,
                capture_output=True,
                text=True
            )
            
            if result.returncode == 0:
                git_info['default_branch'] = result.stdout.strip()
                
        except:
            pass
        
        return git_info
    
    def _analyze_code_structure(self, project_path: str) -> Dict[str, Any]:
        """Analyze project code structure"""
        
        structure_info = {
            'directories': [],
            'file_types': {},
            'total_files': 0
        }
        
        try:
            for root, dirs, files in os.walk(project_path):
                # Skip hidden directories and common ignore patterns
                dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ['node_modules', 'target', '__pycache__']]
                
                if root != project_path:
                    structure_info['directories'].append(os.path.relpath(root, project_path))
                
                for file in files:
                    if not file.startswith('.'):
                        structure_info['total_files'] += 1
                        
                        # Count file types
                        ext = os.path.splitext(file)[1].lower()
                        if ext:
                            structure_info['file_types'][ext] = structure_info['file_types'].get(ext, 0) + 1
            
        except:
            pass
        
        return structure_info
```

#### 3. Automated README Generation Script

```python
class ReadmeGenerator:
    def __init__(self):
        self.template_engine = ReadmeTemplate()
        self.project_analyzer = ProjectAnalyzer()
    
    def generate_basic_readme(self, **kwargs) -> str:
        """Generate a basic README with provided information"""
        
        project_info = {
            'name': kwargs.get('project_name', 'Project Name'),
            'description': kwargs.get('description', 'Project description'),
            'author': kwargs.get('author', ''),
            'license': kwargs.get('license', 'MIT'),
            'version': kwargs.get('version', '1.0.0'),
            'repository_url': kwargs.get('repository_url', ''),
            'package_name': kwargs.get('package_name', kwargs.get('project_name', '').lower().replace(' ', '-')),
            'type': kwargs.get('project_type', 'nodejs')
        }
        
        return self.template_engine.generate_readme(
            project_info=project_info,
            template_type='basic',
            sections=['installation', 'usage', 'license']
        )
    
    def generate_comprehensive_readme(self, project_path: str = "./",
                                    template_type: str = 'professional',
                                    sections: List[str] = None,
                                    **overrides) -> str:
        """Generate a comprehensive README by analyzing the project"""
        
        # Analyze project
        project_info = self.project_analyzer.analyze_project(project_path)
        
        # Apply overrides
        project_info.update(overrides)
        
        # Set default sections if not provided
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
    
    def update_existing_readme(self, project_path: str = "./",
                              sections_to_update: List[str] = None) -> str:
        """Update specific sections of an existing README"""
        
        # Analyze project
        project_info = self.project_analyzer.analyze_project(project_path)
        
        # Read existing README
        readme_path = os.path.join(project_path, 'README.md')
        if os.path.exists(readme_path):
            with open(readme_path, 'r') as f:
                existing_content = f.read()
        else:
            existing_content = ""
        
        # Generate new sections
        updated_sections = {}
        if sections_to_update:
            for section in sections_to_update:
                if section in self.template_engine.section_generators:
                    updated_sections[section] = self.template_engine.section_generators[section](project_info)
        
        # Replace sections in existing content
        updated_content = existing_content
        for section, content in updated_sections.items():
            section_pattern = rf"(?s)## {section.title()}\s*\n.*?(?=\n## |\n# |\Z)"
            updated_content = re.sub(
                section_pattern,
                f"## {section.title()}\n\n{content}\n",
                updated_content
            )
        
        return updated_content
    
    def save_readme(self, content: str, output_path: str = "README.md"):
        """Save README content to file"""
        
        try:
            with open(output_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"README.md saved to {output_path}")
            return True
        except Exception as e:
            print(f"Error saving README: {e}")
            return False
    
    def validate_readme(self, readme_path: str = "README.md") -> Dict[str, Any]:
        """Validate README content for best practices"""
        
        if not os.path.exists(readme_path):
            return {'valid': False, 'errors': ['README.md file not found']}
        
        try:
            with open(readme_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except Exception as e:
            return {'valid': False, 'errors': [f'Error reading README: {e}']}
        
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
        
        # Check for contributing guidelines
        if 'contributing' not in content.lower():
            validation_results['suggestions'].append("Add contributing guidelines")
        
        # Check length
        if validation_results['stats']['character_count'] < 500:
            validation_results['warnings'].append("README seems too short, consider adding more details")
        elif validation_results['stats']['character_count'] > 10000:
            validation_results['suggestions'].append("Consider moving some content to separate documentation files")
        
        # Check for links
        if '[[' in content:  # Double brackets often indicate broken links
            validation_results['warnings'].append("Check for broken or malformed links")
        
        return validation_results

# CLI interface
def main():
    import argparse
    
    parser = argparse.ArgumentParser(description='Generate README.md files')
    parser.add_argument('--project-path', default='.', help='Project path to analyze')
    parser.add_argument('--output', default='README.md', help='Output file path')
    parser.add_argument('--template', default='professional', 
                       choices=['basic', 'professional', 'minimal'],
                       help='README template type')
    parser.add_argument('--sections', nargs='+', 
                       help='Sections to include in README')
    parser.add_argument('--validate', action='store_true', 
                       help='Validate existing README')
    parser.add_argument('--update', nargs='+', 
                       help='Update specific sections in existing README')
    
    args = parser.parse_args()
    
    generator = ReadmeGenerator()
    
    if args.validate:
        # Validate existing README
        validation = generator.validate_readme(args.output)
        print(f"README validation results:")
        print(f"Valid: {validation['valid']}")
        
        if validation.get('warnings'):
            print("\nWarnings:")
            for warning in validation['warnings']:
                print(f"  - {warning}")
        
        if validation.get('suggestions'):
            print("\nSuggestions:")
            for suggestion in validation['suggestions']:
                print(f"  - {suggestion}")
        
        print(f"\nStats: {validation['stats']}")
    
    elif args.update:
        # Update specific sections
        updated_content = generator.update_existing_readme(
            args.project_path, 
            args.update
        )
        generator.save_readme(updated_content, args.output)
    
    else:
        # Generate new README
        readme_content = generator.generate_comprehensive_readme(
            project_path=args.project_path,
            template_type=args.template,
            sections=args.sections
        )
        generator.save_readme(readme_content, args.output)

if __name__ == "__main__":
    main()
```

## Level 3: Advanced Integration

### Specialized README Templates

#### 4. Domain-Specific Templates

```python
class SpecializedTemplates:
    def __init__(self):
        self.templates = {
            'api_service': self._api_service_template,
            'web_app': self._web_app_template,
            'cli_tool': self._cli_tool_template,
            'library': self._library_template,
            'data_science': self._data_science_template,
            'mobile_app': self._mobile_app_template
        }
    
    def get_template(self, project_type: str, project_info: Dict[str, Any]) -> str:
        """Get specialized template for project type"""
        
        template_func = self.templates.get(project_type, self._web_app_template)
        return template_func(project_info)
    
    def _api_service_template(self, project_info: Dict[str, Any]) -> str:
        """Template for API services"""
        return f"""
# {project_info.get('name', 'API Service')}

{project_info.get('description', 'A RESTful API service')}

{self._generate_api_badges(project_info)}

## Table of Contents

- [Features](#features)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Installation](#installation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring](#monitoring)

## Features

- RESTful API design
- JWT authentication
- Rate limiting
- Comprehensive error handling
- API documentation
- Request validation
- Response caching

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

**Response:**
```json
{{
  "token": "jwt_token_here",
  "expires_in": 3600
}}
```

### Users

#### GET /api/users
Get list of users (requires authentication).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{{
  "users": [
    {{
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe"
    }}
  ],
  "pagination": {{
    "page": 1,
    "limit": 20,
    "total": 100
  }}
}}
```

## Authentication

This API uses JWT (JSON Web Token) authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Rate Limiting

API requests are rate limited to prevent abuse:
- **Standard users**: 100 requests per hour
- **Premium users**: 1000 requests per hour

Rate limit headers are included in each response:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

```json
{{
  "error": {{
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {{
        "field": "email",
        "message": "Invalid email format"
      }}
    ]
  }}
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

## Development

```bash
# Install development dependencies
{project_info.get('dev_install_command', 'npm install')}

# Run in development mode
{project_info.get('dev_command', 'npm run dev')}

# Run tests
{project_info.get('test_command', 'npm test')}

# Run linting
{project_info.get('lint_command', 'npm run lint')}
```

## Testing

```bash
# Run all tests
{project_info.get('test_command', 'npm test')}

# Run tests with coverage
{project_info.get('coverage_command', 'npm run test:coverage')}

# Run integration tests
npm run test:integration

# Run API tests
npm run test:api
```

## Deployment

### Docker

```bash
# Build image
docker build -t {project_info.get('docker_image', 'api-service')} .

# Run container
docker run -p 3000:3000 --env-file .env {project_info.get('docker_image', 'api-service')}
```

### Docker Compose

```bash
# Deploy with database
docker-compose up -d
```

### Cloud Deployment

**AWS ECS**
```bash
# Deploy to AWS ECS
aws ecs deploy --service my-api-service
```

**Google Cloud Run**
```bash
# Deploy to Cloud Run
gcloud run deploy --image gcr.io/project-id/api-service
```

## Monitoring

- Health check endpoint: `GET /health`
- Metrics endpoint: `GET /metrics`
- Logs: Structured JSON logging
- Monitoring: Prometheus + Grafana
- Error tracking: Sentry

### Health Check Response

```json
{{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "{project_info.get('version', '1.0.0')}",
  "database": "connected",
  "redis": "connected"
}}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

{self._get_license_text(project_info)}
        """.strip()
    
    def _web_app_template(self, project_info: Dict[str, Any]) -> str:
        """Template for web applications"""
        return f"""
# {project_info.get('name', 'Web Application')}

{project_info.get('description', 'A modern web application')}

{self._generate_web_app_badges(project_info)}

## Table of Contents

- [Features](#features)
- [Technology Stack](#technology-stack)
- [Screenshots](#screenshots)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)

## Features

- 🚀 **Modern UI/UX**: Built with latest frontend technologies
- 📱 **Responsive Design**: Works on all devices
- 🔐 **Authentication**: Secure user authentication
- 🌍 **Internationalization**: Multi-language support
- ⚡ **Performance**: Optimized for speed
- 🔍 **SEO**: Search engine optimized
- 📊 **Analytics**: Built-in analytics tracking

## Technology Stack

### Frontend
- **Framework**: {project_info.get('frontend_framework', 'React')}
- **Styling**: {project_info.get('styling', 'Tailwind CSS')}
- **State Management**: {project_info.get('state_management', 'Redux Toolkit')}
- **Routing**: {project_info.get('routing', 'React Router')}
- **Build Tool**: {project_info.get('build_tool', 'Vite')}

### Backend
- **Runtime**: {project_info.get('backend_runtime', 'Node.js')}
- **Framework**: {project_info.get('backend_framework', 'Express.js')}
- **Database**: {project_info.get('database', 'PostgreSQL')}
- **ORM**: {project_info.get('orm', 'Prisma')}
- **Authentication**: {project_info.get('auth', 'JWT')}

### Infrastructure
- **Hosting**: {project_info.get('hosting', 'Vercel')}
- **Database**: {project_info.get('db_hosting', 'Heroku Postgres')}
- **CDN**: {project_info.get('cdn', 'Cloudflare')}
- **Monitoring**: {project_info.get('monitoring', 'Sentry')}

## Screenshots

### Desktop View
![Desktop Screenshot](./screenshots/desktop.png)

### Mobile View
![Mobile Screenshot](./screenshots/mobile.png)

## Installation

### Prerequisites

- {project_info.get('node_version', 'Node.js 18+')}
- {project_info.get('database_requirement', 'PostgreSQL 14+')}

### Setup

```bash
# Clone the repository
git clone {project_info.get('repository_url', '')}

# Navigate to project directory
cd {project_info.get('name', 'my-app')}

# Install dependencies
npm install

# Set up environment variables
cp .env.example .env

# Set up database
npm run db:setup

# Run the application
npm run dev
```

## Usage

1. **Sign Up**: Create a new account
2. **Login**: Authenticate with your credentials
3. **Dashboard**: Access your personalized dashboard
4. **Features**: Explore various features
5. **Settings**: Configure your preferences

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/myapp

# Authentication
JWT_SECRET=your_jwt_secret_here
JWT_EXPIRES_IN=7d

# Application
NODE_ENV=development
PORT=3000

# External Services
API_KEY=your_api_key
WEBHOOK_URL=your_webhook_url
```

### Database Configuration

```bash
# Run migrations
npm run db:migrate

# Seed database
npm run db:seed

# Reset database
npm run db:reset
```

## Development

### Available Scripts

```bash
# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Run linting
npm run lint

# Run type checking
npm run type-check

# Format code
npm run format
```

### Code Style

This project uses:
- **ESLint** for code linting
- **Prettier** for code formatting
- **Husky** for git hooks
- **lint-staged** for pre-commit checks

### Git Hooks

Pre-commit hooks run automatically:
- ESLint check
- Prettier formatting
- Type checking
- Unit tests

## Testing

### Test Structure

```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
├── e2e/           # End-to-end tests
└── fixtures/      # Test data
```

### Running Tests

```bash
# Run all tests
npm test

# Run specific test file
npm test -- login.test.js

# Run tests with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

### Coverage Reports

Coverage reports are generated in `coverage/` directory.
Target coverage: **90%**

## Deployment

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

### Docker Deployment

```bash
# Build Docker image
docker build -t {project_info.get('docker_image', 'my-app')} .

# Run container
docker run -p 3000:3000 {project_info.get('docker_image', 'my-app')}
```

### Vercel Deployment

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy to Vercel
vercel --prod
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    depends_on:
      - db

  db:
    image: postgres:14
    environment:
      - POSTGRES_DB=myapp
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
# Deploy with Docker Compose
docker-compose up -d
```

## Performance

### Optimization Features

- **Code Splitting**: Automatic code splitting
- **Lazy Loading**: Components loaded on demand
- **Image Optimization**: WebP format, lazy loading
- **Caching**: Browser and CDN caching
- **Bundle Analysis**: Optimized bundle sizes

### Performance Metrics

- **Lighthouse Score**: 95+
- **First Contentful Paint**: < 1.5s
- **Largest Contentful Paint**: < 2.5s
- **Cumulative Layout Shift**: < 0.1

## Security

### Security Measures

- **Authentication**: JWT-based authentication
- **Authorization**: Role-based access control
- **Input Validation**: Comprehensive input validation
- **SQL Injection Prevention**: Parameterized queries
- **XSS Protection**: Content Security Policy
- **HTTPS**: SSL/TLS encryption

### Security Headers

```
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Ensure all tests pass
6. Submit a pull request

## License

{self._get_license_text(project_info)}

## Support

- 📧 **Email**: support@example.com
- 💬 **Discord**: [Join our Discord](https://discord.gg/example)
- 🐦 **Twitter**: [@example](https://twitter.com/example)
- 📖 **Documentation**: [docs.example.com](https://docs.example.com)
        """.strip()
    
    def _cli_tool_template(self, project_info: Dict[str, Any]) -> str:
        """Template for CLI tools"""
        return f"""
# {project_info.get('name', 'CLI Tool')}

{project_info.get('description', 'A powerful command-line tool')}

{self._generate_cli_badges(project_info)}

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Configuration](#configuration)
- [Examples](#examples)
- [Development](#development)
- [Contributing](#contributing)

## Features

- ⚡ **Fast**: Optimized for performance
- 🔧 **Configurable**: Flexible configuration options
- 📝 **Well-documented**: Comprehensive help and documentation
- 🎨 **Colored output**: Beautiful terminal output
- 🔌 **Extensible**: Plugin architecture
- 🧪 **Tested**: Comprehensive test suite

## Installation

### npm

```bash
npm install -g {project_info.get('package_name', 'my-cli')}
```

### yarn

```bash
yarn global add {project_info.get('package_name', 'my-cli')}
```

### Download Binary

Download the appropriate binary for your system from the [releases page](https://github.com/{project_info.get('repository', 'user/repo')}/releases).

### Build from Source

```bash
git clone https://github.com/{project_info.get('repository', 'user/repo')}.git
cd {project_info.get('name', 'my-cli')}
npm install
npm run build
npm link
```

## Usage

### Basic Usage

```bash
# Show help
{project_info.get('command_name', 'my-cli')} --help

# Show version
{project_info.get('command_name', 'my-cli')} --version

# Run command
{project_info.get('command_name', 'my-cli')} [command] [options]
```

### Global Options

```bash
  -h, --help              Show help
  -v, --version           Show version
  -c, --config <path>     Path to config file
  -q, --quiet             Quiet mode
  -v, --verbose           Verbose output
  --no-color              Disable colored output
```

## Commands

### {project_info.get('main_command', 'build')}

Build the project.

```bash
{project_info.get('command_name', 'my-cli')} {project_info.get('main_command', 'build')} [options]
```

**Options:**
```bash
  -o, --output <dir>      Output directory (default: dist)
  -w, --watch             Watch for changes
  -m, --minify            Minify output
  -s, --sourcemap         Generate sourcemaps
```

**Example:**
```bash
{project_info.get('command_name', 'my-cli')} {project_info.get('main_command', 'build')} --output build --minify
```

### init

Initialize a new project.

```bash
{project_info.get('command_name', 'my-cli')} init [project-name] [options]
```

**Options:**
```bash
  -t, --template <name>   Template to use
  -f, --force             Force overwrite existing files
  -g, --git               Initialize git repository
```

**Example:**
```bash
{project_info.get('command_name', 'my-cli')} init my-project --template react --git
```

### serve

Start development server.

```bash
{project_info.get('command_name', 'my-cli')} serve [options]
```

**Options:**
```bash
  -p, --port <number>     Port number (default: 3000)
  -h, --host <host>       Host address (default: localhost)
  -o, --open              Open browser automatically
```

**Example:**
```bash
{project_info.get('command_name', 'my-cli')} serve --port 8080 --open
```

## Configuration

### Config File

Create a configuration file in one of these locations:

- `{project_info.get('config_file', '.myclirc')}` in project directory
- `~/.{project_info.get('config_file', 'myclirc')}` in home directory
- `{project_info.get('config_file', 'mycli')}.config.js` for JavaScript config

### Example Configuration

```json
{{
  "build": {{
    "outputDir": "dist",
    "minify": true,
    "sourcemap": true
  }},
  "serve": {{
    "port": 3000,
    "host": "localhost",
    "open": true
  }},
  "plugins": [
    "@my-cli/plugin-typescript",
    "@my-cli/plugin-tailwind"
  ]
}}
```

### Environment Variables

```bash
MY_CLI_CONFIG_PATH=/path/to/config
MY_CLI_LOG_LEVEL=info
MY_CLI_NO_COLOR=1
```

## Examples

### Example 1: Basic Project Setup

```bash
# Initialize new project
{project_info.get('command_name', 'my-cli')} init my-awesome-project

# Navigate to project
cd my-awesome-project

# Start development
{project_info.get('command_name', 'my-cli')} serve --open

# Build for production
{project_info.get('command_name', 'my-cli')} build --minify
```

### Example 2: Custom Configuration

```bash
# Use custom config
{project_info.get('command_name', 'my-cli')} build --config ./my-config.json

# Override config options
{project_info.get('command_name', 'my-cli')} build --output ./build --no-minify

# Use environment variables
MY_CLI_LOG_LEVEL=debug {project_info.get('command_name', 'my-cli')} build
```

### Example 3: Advanced Usage

```bash
# Watch mode with custom port
{project_info.get('command_name', 'my-cli')} serve --port 8080 --watch

# Build specific targets
{project_info.get('command_name', 'my-cli')} build --target es2020 --format esm

# Run with plugins
{project_info.get('command_name', 'my-cli')} build --plugin @my-cli/plugin-typescript
```

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/{project_info.get('repository', 'user/repo')}.git
cd {project_info.get('name', 'my-cli')}

# Install dependencies
npm install

# Run tests
npm test

# Run in development mode
npm run dev

# Build for development
npm run build:dev

# Run CLI from source
node dist/cli.js --help
```

### Project Structure

```
src/
├── cli.js              # CLI entry point
├── commands/           # Command implementations
│   ├── build.js
│   ├── serve.js
│   └── init.js
├── utils/              # Utility functions
├── templates/          # Project templates
└── plugins/            # Built-in plugins

tests/
├── unit/               # Unit tests
├── integration/        # Integration tests
└── fixtures/           # Test fixtures
```

### Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run integration tests
npm run test:integration
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md).

### Adding New Commands

1. Create command file in `src/commands/`
2. Export command function
3. Add command to CLI configuration
4. Write tests
5. Update documentation

### Plugin Development

Create a plugin by extending the CLI:

```javascript
// my-plugin.js
module.exports = function(cli) {{
  cli.command('my-command', 'My custom command', (args, options) => {{
    // Command implementation
  }});
}};
```

## License

{self._get_license_text(project_info)}

## Support

- 📖 **Documentation**: [docs.example.com](https://docs.example.com)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/{project_info.get('repository', 'user/repo')}/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/{project_info.get('repository', 'user/repo')}/discussions)
- 📧 **Email**: support@example.com
        """.strip()
    
    def _generate_api_badges(self, project_info: Dict[str, Any]) -> str:
        """Generate badges for API projects"""
        badges = []
        
        if project_info.get('repository'):
            badges.append(f"![Build Status](https://github.com/{project_info['repository']}/workflows/CI/badge.svg)")
        
        badges.append("![API](https://img.shields.io/badge/API-REST-blue.svg)")
        badges.append("![Authentication](https://img.shields.io/badge/Auth-JWT-green.svg)")
        
        return " ".join(badges)
    
    def _generate_web_app_badges(self, project_info: Dict[str, Any]) -> str:
        """Generate badges for web applications"""
        badges = []
        
        if project_info.get('repository'):
            badges.append(f"![Build Status](https://github.com/{project_info['repository']}/workflows/CI/badge.svg)")
        
        badges.append("![Web App](https://img.shields.io/badge/Web-Application-blue.svg)")
        badges.append("![Responsive](https://img.shields.io/badge/Design-Responsive-green.svg)")
        
        return " ".join(badges)
    
    def _generate_cli_badges(self, project_info: Dict[str, Any]) -> str:
        """Generate badges for CLI tools"""
        badges = []
        
        if project_info.get('repository'):
            badges.append(f"![Build Status](https://github.com/{project_info['repository']}/workflows/CI/badge.svg)")
        
        badges.append("![CLI](https://img.shields.io/badge/CLI-Tool-orange.svg)")
        badges.append("![Node](https://img.shields.io/badge/Node.js-18%2B-green.svg)")
        
        return " ".join(badges)
    
    def _get_install_command(self, project_info: Dict[str, Any]) -> str:
        """Get install command based on project type"""
        project_type = project_info.get('type', 'nodejs')
        
        commands = {
            'nodejs': 'npm install',
            'python': 'pip install -r requirements.txt',
            'rust': 'cargo build --release',
            'go': 'go mod tidy && go build'
        }
        
        return commands.get(project_type, 'npm install')
    
    def _get_migration_command(self, project_info: Dict[str, Any]) -> str:
        """Get database migration command"""
        project_type = project_info.get('type', 'nodejs')
        
        commands = {
            'nodejs': 'npm run db:migrate',
            'python': 'python manage.py migrate',
            'rust': 'sqlx migrate run',
            'go': 'go run cmd/migrate/main.go'
        }
        
        return commands.get(project_type, 'npm run db:migrate')
    
    def _get_start_command(self, project_info: Dict[str, Any]) -> str:
        """Get start command"""
        project_type = project_info.get('type', 'nodejs')
        
        commands = {
            'nodejs': 'npm start',
            'python': 'python manage.py runserver',
            'rust': './target/release/myapp',
            'go': './myapp'
        }
        
        return commands.get(project_type, 'npm start')
    
    def _get_license_text(self, project_info: Dict[str, Any]) -> str:
        """Get license text"""
        license_type = project_info.get('license', 'MIT')
        
        if license_type == 'MIT':
            return "This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details."
        else:
            return f"This project is licensed under the {license_type} License - see the [LICENSE](LICENSE) file for details."
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
