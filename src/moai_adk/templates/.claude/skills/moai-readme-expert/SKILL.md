---
name: "moai-readme-expert"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# Skill: README.md Expert

## Metadata

```yaml
skill_id: moai-readme-expert
skill_name: README.md Expert
version: 1.0.0
created_date: 2025-11-11
updated_date: 2025-11-11
language: english
word_count: 1600
triggers:
  - keywords: [readme, project documentation, getting started, installation guide, contributing]
  - contexts: [readme-generation, project-documentation, onboarding-guide, github-docs]
agents:
  - docs-manager
  - project-manager
  - frontend-expert
  - backend-expert
freedom_level: high
context7_references:
  - url: "https://github.com/matiassingers/awesome-readme"
    topic: "Best practices and examples of excellent README files"
  - url: "https://www.makeareadme.com"
    topic: "README.md generation guidelines and standards"
  - url: "https://github.com/OAI/OpenAPI-Specification"
    topic: "API documentation standards for README integration"
spec_reference: "@SPEC:README-STANDARDS-001"
```

## 📚 Content

### Section 1: README Architecture Principles

#### Core Objectives

A professional README.md serves as the **front door** to your project. It must accomplish:

1. **Immediate Understanding**: What the project does in 5 seconds
2. **Quick Success**: Get users up and running in under 2 minutes
3. **Clear Next Steps**: Guide users to appropriate documentation
4. **Professional Presentation**: Build trust and credibility
5. **Contributor Guidance**: Enable community participation

#### The 5-Second Test

Users should understand your project within 5 seconds by answering:

- **What**: What problem does this solve?
- **Who**: Who is this for?
- **Why**: Why choose this solution?
- **How**: How do I get started?

```markdown
# 🚀 Project Name

> One-sentence elevator pitch explaining what this does and who it's for

[![Build Status](https://img.shields.io/github/workflow/status/user/repo/build)](https://github.com/user/repo/actions)
[![Coverage](https://img.shields.io/codecov/c/github/user/repo)](https://codecov.io/gh/user/repo)
[![License](https://img.shields.io/npm/l/@user/repo)](LICENSE)
[![Version](https://img.shields.io/npm/v/@user/repo)](https://www.npmjs.com/package/@user/repo)
[![Downloads](https://img.shields.io/npm/dm/@user/repo)](https://www.npmjs.com/package/@user/repo)
```

### Section 2: README Template Structure

#### Professional README Blueprint

```markdown
# 🚀 Project Name

> Brief, compelling description that explains the project's purpose and value proposition

[![Build Status](https://img.shields.io/github/workflow/status/user/repo/build)](https://github.com/user/repo/actions)
[![Coverage](https://img.shields.io/codecov/c/github/user/repo)](https://codecov.io/gh/user/repo)
[![License](https://img.shields.io/npm/l/@user/repo)](LICENSE)
[![Version](https://img.shields.io/npm/v/@user/repo)](https://www.npmjs.com/package/@user/repo)
[![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/Node.js-18%2B-green)](https://nodejs.org/)

## ✨ Features

- 🎯 **Feature 1**: Clear benefit-oriented description
- ⚡ **Feature 2**: Performance or efficiency highlight
- 🔧 **Feature 3**: Technical capability with user value
- 🛡️ **Feature 4**: Security or reliability aspect
- 📱 **Feature 5**: Platform compatibility

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ and npm
- [Other requirement 1]
- [Other requirement 2]

### Installation

```bash
# npm
npm install @user/repo

# yarn
yarn add @user/repo

# pnpm
pnpm add @user/repo
```

### Basic Usage

```typescript
import { ProjectMain } from '@user/repo'

const instance = new ProjectMain({
  apiKey: 'your-api-key',
  config: {
    // Configuration options
  }
})

// Start using the project
instance.initialize()
  .then(() => console.log('✅ Ready to go!'))
  .catch(error => console.error('❌ Setup failed:', error))
```

## 📚 Documentation

- **[Getting Started Guide](docs/getting-started.md)** - Comprehensive setup instructions
- **[API Reference](docs/api/)** - Complete API documentation
- **[Examples](examples/)** - Real-world usage examples
- **[Migration Guide](docs/migration.md)** - Upgrading from previous versions

## 🏗️ Architecture

```mermaid
graph TB
    subgraph "Frontend"
        WebApp[Web Application]
        MobileApp[Mobile App]
    end

    subgraph "Backend"
        API[REST API]
        Auth[Authentication]
        Queue[Message Queue]
    end

    subgraph "Data Layer"
        Database[(PostgreSQL)]
        Cache[(Redis)]
    end

    WebApp --> API
    MobileApp --> API
    API --> Auth
    API --> Queue
    API --> Database
    API --> Cache
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in your project root:

```env
# Required
API_KEY=your_api_key_here
DATABASE_URL=postgresql://user:password@localhost:5432/dbname

# Optional
LOG_LEVEL=info
CACHE_TTL=3600
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | `string` | `null` | Your API authentication key |
| `timeout` | `number` | `5000` | Request timeout in milliseconds |
| `retries` | `number` | `3` | Number of retry attempts |
| `debug` | `boolean` | `false` | Enable debug logging |

## 🧪 Development

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/user/repo.git
cd repo

# Install dependencies
npm install

# Copy environment template
cp .env.example .env

# Start development server
npm run dev
```

### Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run test` | Run test suite |
| `npm run test:watch` | Run tests in watch mode |
| `npm run test:coverage` | Run tests with coverage report |
| `npm run lint` | Run linting |
| `npm run lint:fix` | Fix linting issues |
| `npm run type-check` | Run TypeScript type checking |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Contribution Guide

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Commit** your changes: `git commit -m 'Add amazing feature'`
4. **Push** to the branch: `git push origin feature/amazing-feature`
5. **Open** a Pull Request

### Development Standards

- ✅ Follow the existing code style and conventions
- ✅ Add tests for new functionality
- ✅ Update documentation for API changes
- ✅ Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support & Troubleshooting

### Getting Help

- 📖 [Documentation](docs/) - Comprehensive guides and API reference
- 🐛 [Issue Tracker](https://github.com/user/repo/issues) - Bug reports and feature requests
- 💬 [Discussions](https://github.com/user/repo/discussions) - Community discussions
- 📧 [Email Support](mailto:support@example.com) - For commercial support

### Common Issues

<details>
<summary>Installation fails with permission errors</summary>

```bash
# Try these solutions:
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH

# Or use npx
npx @user/repo
```

</details>

<details>
<summary>API authentication errors</summary>

1. Verify your API key is correct
2. Check environment variables are properly set
3. Ensure you're using the correct API endpoint

</details>

### Performance Tips

- Use the latest version for optimal performance
- Enable caching for repeated requests
- Consider using the CDN version for web applications

## 🗺️ Roadmap

View our [Project Roadmap](ROADMAP.md) to see what's planned:

- [x] **Version 1.0**: Core functionality
- [ ] **Version 1.1**: Performance improvements
- [ ] **Version 2.0**: Advanced features
- [ ] **Version 2.1**: Plugin system

## 📊 Statistics

![GitHub stars](https://img.shields.io/github/stars/user/repo?style=social)
![GitHub forks](https://img.shields.io/github/forks/user/repo?style=social)
![GitHub issues](https://img.shields.io/github/issues/user/repo)
![GitHub pull requests](https://img.shields.io/github/issues-pr/user/repo)

---

<div align="center">
  <strong>Made with ❤️ by the [Your Team](https://github.com/user)</strong>
</div>
```

### Section 3: Dynamic README Generation

#### From Project Analysis

```python
class ReadmeGenerator:
    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
        self.project_data = self.analyze_project()

    def analyze_project(self) -> dict:
        """Comprehensive project analysis"""
        return {
            'package_info': self.extract_package_info(),
            'dependencies': self.analyze_dependencies(),
            'scripts': self.get_available_scripts(),
            'structure': self.analyze_directory_structure(),
            'api_endpoints': self.extract_api_info(),
            'tests': self.analyze_test_coverage(),
            'badges': self.generate_badges(),
        }

    def generate_readme(self, template: str = 'professional') -> str:
        """Generate README from project analysis"""

        templates = {
            'professional': self.generate_professional_readme,
            'minimal': self.generate_minimal_readme,
            'api_focused': self.generate_api_readme,
            'library': self.generate_library_readme,
        }

        return templates[template](self.project_data)

    def extract_package_info(self) -> dict:
        """Extract information from package.json"""
        package_file = self.project_path / 'package.json'
        if not package_file.exists():
            return {}

        import json
        with open(package_file) as f:
            package = json.load(f)

        return {
            'name': package.get('name', 'project'),
            'version': package.get('version', '1.0.0'),
            'description': package.get('description', ''),
            'author': package.get('author', ''),
            'license': package.get('license', 'MIT'),
            'repository': package.get('repository', {}).get('url', ''),
            'keywords': package.get('keywords', []),
        }

    def analyze_dependencies(self) -> dict:
        """Categorize and analyze dependencies"""
        package_file = self.project_path / 'package.json'
        if not package_file.exists():
            return {}

        import json
        with open(package_file) as f:
            package = json.load(f)

        dependencies = package.get('dependencies', {})
        dev_dependencies = package.get('devDependencies', {})

        # Categorize dependencies
        categories = {
            'framework': [],
            'database': [],
            'testing': [],
            'build': [],
            'styling': [],
            'other': []
        }

        framework_keywords = ['react', 'vue', 'angular', 'express', 'fastapi', 'django']
        database_keywords = ['prisma', 'mongoose', 'typeorm', 'sequelize']
        testing_keywords = ['jest', 'vitest', 'cypress', 'playwright', 'testing-library']
        build_keywords = ['webpack', 'vite', 'rollup', 'esbuild', 'babel']
        styling_keywords = ['tailwind', 'styled-components', 'emotion', 'sass']

        for dep in dependencies:
            if any(keyword in dep.lower() for keyword in framework_keywords):
                categories['framework'].append(dep)
            elif any(keyword in dep.lower() for keyword in database_keywords):
                categories['database'].append(dep)
            elif any(keyword in dep.lower() for keyword in testing_keywords):
                categories['testing'].append(dep)
            elif any(keyword in dep.lower() for keyword in build_keywords):
                categories['build'].append(dep)
            elif any(keyword in dep.lower() for keyword in styling_keywords):
                categories['styling'].append(dep)
            else:
                categories['other'].append(dep)

        return categories

    def generate_badges(self) -> list:
        """Generate relevant badges"""
        package_info = self.project_data['package_info']
        badges = []

        # Build status badge
        badges.append(
            f"[![Build Status](https://img.shields.io/github/workflow/status/"
            f"username/{package_info['name']}/build)]"
            f"(https://github.com/username/{package_info['name']}/actions)"
        )

        # License badge
        if package_info.get('license'):
            badges.append(
                f"[![License](https://img.shields.io/npm/l/{package_info['name']}]"
                f"](LICENSE)"
            )

        # npm version badge
        badges.append(
            f"[![Version](https://img.shields.io/npm/v/{package_info['name']}]"
            f"](https://www.npmjs.com/package/{package_info['name']})"
        )

        # Technology badges
        dependencies = self.project_data['dependencies']
        if any('react' in dep.lower() for dep in dependencies['framework']):
            badges.append(
                "[![React](https://img.shields.io/badge/React-20232A?logo=react&logoColor=61DAFB)]"
                "(https://reactjs.org/)"
            )

        if any('typescript' in dep.lower() for dep in dependencies['other']):
            badges.append(
                "[![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?logo=typescript&logoColor=white)]"
                "(https://www.typescriptlang.org/)"
            )

        return badges

    def generate_professional_readme(self, project_data: dict) -> str:
        """Generate professional README template"""

        package_info = project_data['package_info']
        badges = project_data['badges']
        dependencies = project_data['dependencies']
        scripts = project_data['scripts']

        # Generate features section
        features = self.generate_features_section(project_data)

        # Generate installation section
        installation = self.generate_installation_section(package_info)

        # Generate usage section
        usage = self.generate_usage_section(project_data)

        # Generate architecture diagram
        architecture = self.generate_architecture_section(project_data)

        template = f"""# 🚀 {package_info.get('name', 'Project Name').title()}

> {package_info.get('description', 'Add your project description here')}

{' '.join(badges)}

## ✨ Features

{features}

## 🚀 Quick Start

### Installation

{installation}

### Basic Usage

{usage}

## 📚 Documentation

- **[Getting Started Guide](docs/getting-started.md)** - Comprehensive setup instructions
- **[API Reference](docs/api/)** - Complete API documentation
- **[Examples](examples/)** - Real-world usage examples

## 🏗️ Architecture

{architecture}

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/username/{package_info['name']}.git
cd {package_info['name']}

# Install dependencies
npm install

# Start development
npm run dev
```

## 📄 License

This project is licensed under the {package_info.get('license', 'MIT')} License.

## 🆘 Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/username/{package_info['name']}/issues)
- 💬 [Discussions](https://github.com/username/{package_info['name']}/discussions)

---

<div align="center">
  <strong>Made with ❤️ by the Development Team</strong>
</div>
"""

        return template

    def generate_features_section(self, project_data: dict) -> str:
        """Generate features section based on project analysis"""

        dependencies = project_data['dependencies']
        features = []

        # Framework features
        if dependencies['framework']:
            features.append(
                f"🎯 **{dependencies['framework'][0].title()} Integration**: "
                f"Seamless integration with {dependencies['framework'][0]}"
            )

        # Database features
        if dependencies['database']:
            features.append(
                f"🗄️ **Database Support**: "
                f"Built-in support for {', '.join(dependencies['database'])}"
            )

        # Testing features
        if dependencies['testing']:
            features.append(
                f"🧪 **Comprehensive Testing**: "
                f"Full test suite using {', '.join(dependencies['testing'])}"
            )

        # Type safety
        if any('typescript' in dep.lower() for dep in dependencies['other']):
            features.append(
                "🔒 **Type Safety**: Full TypeScript support with strict type checking"
            )

        # Performance
        features.append(
            "⚡ **High Performance**: Optimized for speed and efficiency"
        )

        # Documentation
        features.append(
            "📚 **Well Documented**: Comprehensive guides and API documentation"
        )

        return '\n'.join(f"- {feature}" for feature in features)

    def generate_installation_section(self, package_info: dict) -> str:
        """Generate installation instructions"""

        project_name = package_info.get('name', 'project')

        return f"""```bash
# npm
npm install {project_name}

# yarn
yarn add {project_name}

# pnpm
pnpm add {project_name}
```

### Prerequisites

- Node.js 18+ and npm
- [Any other specific requirements]

### Environment Setup

Create a `.env` file:

```env
# Required configuration
API_KEY=your_api_key_here
"""

    def generate_usage_section(self, project_data: dict) -> str:
        """Generate usage example based on project structure"""

        package_name = project_data['package_info']['name']

        # Check if it's likely an API or library
        dependencies = project_data['dependencies']

        if dependencies['framework'] and any('express' in dep.lower() or 'fastapi' in dep.lower() for dep in dependencies['framework']):
            return f"""```typescript
import {package_name} from '{package_name}'

const app = new {package_name.title()}({{
  port: 3000,
  database: {{
    url: process.env.DATABASE_URL
  }}
}})

app.start()
  .then(() => console.log('🚀 Server ready!'))
  .catch(error => console.error('❌ Startup failed:', error))
```

        else:
            return f"""```typescript
import {{ {package_name.title()} }} from '{package_name}'

const instance = new {package_name.title()}({{
  apiKey: process.env.API_KEY,
  config: {{
    timeout: 5000,
    retries: 3
  }}
}})

// Start using the library
instance.initialize()
  .then(() => console.log('✅ Ready to use!'))
  .catch(error => console.error('❌ Setup failed:', error))
```

    def generate_architecture_section(self, project_data: dict) -> str:
        """Generate architecture diagram based on project structure"""

        dependencies = project_data['dependencies']

        # Generate mermaid diagram based on detected technologies
        diagram_parts = []

        if dependencies['framework']:
            diagram_parts.append("    subgraph \"Application Layer\"")
            for framework in dependencies['framework'][:3]:  # Limit to 3
                diagram_parts.append(f"        {framework.title()}[{framework.title()}]")
            diagram_parts.append("    end")

        if dependencies['database']:
            diagram_parts.append("    subgraph \"Data Layer\"")
            for db in dependencies['database'][:3]:  # Limit to 3
                diagram_parts.append(f"        {db.title()}[({db.title()})]")
            diagram_parts.append("    end")

        if dependencies['other']:
            diagram_parts.append("    subgraph \"Services\"")
            for service in dependencies['other'][:2]:  # Limit to 2
                diagram_parts.append(f"        {service.title()}[{service.title()}]")
            diagram_parts.append("    end")

        mermaid_diagram = "```mermaid\ngraph TB\n" + "\n".join(diagram_parts) + "\n```"

        return mermaid_diagram
```

### Section 4: Specialized README Templates

#### API Project Template

```markdown
# 🔌 API Project Name

> RESTful API for [service description] with comprehensive documentation and testing

[![API Documentation](https://img.shields.io/badge/API-Doc-Ready-brightgreen)](docs/api/)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-6BA539)](openapi.yaml)
[![Postman](https://img.shields.io/badge/Postman-Collection-FF6C37)](collection.json)

## 🌐 API Overview

| Environment | Base URL | Authentication |
|-------------|----------|----------------|
| Development | `https://api-dev.example.com/v1` | Bearer Token |
| Staging | `https://api-staging.example.com/v1` | Bearer Token |
| Production | `https://api.example.com/v1` | Bearer Token |

## 📊 API Statistics

- **Endpoints**: 25+ REST endpoints
- **Rate Limiting**: 1000 requests/hour
- **Data Format**: JSON
- **Authentication**: JWT Bearer Tokens

## 🚀 Quick Start

### 1. Get API Key

```bash
curl -X POST https://api.example.com/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "your@email.com", "password": "your-password"}'
```

### 2. Make Your First Request

```bash
curl -X GET https://api.example.com/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. Import Postman Collection

Download our [Postman Collection](collection.json) to explore all endpoints.

## 📚 API Documentation

- **[Interactive API Docs](https://api.example.com/docs)** - Swagger UI playground
- **[OpenAPI Specification](openapi.yaml)** - Complete API specification
- **[Authentication Guide](docs/authentication.md)** - JWT authentication details
- **[Rate Limiting](docs/rate-limiting.md)** - Usage limits and best practices

## 🔧 Core Endpoints

### Authentication
```http
POST /v1/auth/login
POST /v1/auth/refresh
DELETE /v1/auth/logout
```

### Users
```http
GET    /v1/users
GET    /v1/users/:id
POST   /v1/users
PUT    /v1/users/:id
DELETE /v1/users/:id
```

### Resources
```http
GET    /v1/resources
POST   /v1/resources
PUT    /v1/resources/:id
DELETE /v1/resources/:id
```

## 🧪 Testing

### Run API Tests
```bash
# Install dependencies
npm install

# Run all tests
npm test

# Run integration tests
npm run test:integration

# Run API performance tests
npm run test:performance
```

### Test Coverage
- Unit Tests: 95% coverage
- Integration Tests: All endpoints tested
- Performance Tests: Load testing up to 1000 RPS
```

#### Library Project Template

```markdown
# 📦 Library Name

> Modern, type-safe library for [functionality] with zero dependencies

[![npm version](https://img.shields.io/npm/v/@user/library)](https://www.npmjs.com/package/@user/library)
[![bundle size](https://img.shields.io/bundlephobia/min/@user/library)](https://bundlephobia.com/result?p=@user/library)
[![typescript](https://img.shields.io/badge/TypeScript-4.9+-blue)](https://www.typescriptlang.org/)
[![tree-shakeable](https://img.shields.io/badge/Tree%20Shakeable-✓-brightgreen)](https://webpack.js.org/guides/tree-shaking/)

## 🎯 Why This Library?

- 🚀 **Zero Dependencies**: No runtime dependencies, minimal bundle size
- 🔒 **Type Safe**: Full TypeScript support with strict typing
- 📱 **Universal**: Works in Node.js, browsers, and edge environments
- ⚡ **Performant**: Optimized for speed and memory usage
- 🧪 **Well Tested**: 99% test coverage with comprehensive examples

## 📦 Installation

```bash
# npm
npm install @user/library

# yarn
yarn add @user/library

# pnpm
pnpm add @user/library
```

## 🚀 Usage

### Basic Example

```typescript
import { mainFunction } from '@user/library'

const result = await mainFunction({
  input: 'example',
  options: {
    timeout: 5000,
    retries: 3
  }
})

console.log(result) // Output: processed data
```

### Advanced Usage

```typescript
import {
  mainFunction,
  configure,
  types
} from '@user/library'

// Configure global settings
configure({
  apiKey: 'your-api-key',
  timeout: 10000
})

// Use with TypeScript
interface CustomType extends types.InputType {
  customField: string
}

const customInput: CustomType = {
  input: 'test',
  customField: 'value'
}

const result = await mainFunction(customInput)
```

## 📚 API Reference

### Main Functions

| Function | Description | Parameters | Returns |
|----------|-------------|------------|---------|
| `mainFunction()` | Primary functionality | `InputType, Options` | `OutputType` |
| `configure()` | Global configuration | `ConfigOptions` | `void` |
| `validate()` | Input validation | `InputType` | `boolean` |
| `reset()` | Reset to defaults | - | `void` |

### Types

```typescript
interface InputType {
  input: string
  priority?: number
}

interface OutputType {
  success: boolean
  data: unknown
  error?: string
}

interface Options {
  timeout?: number
  retries?: number
  silent?: boolean
}
```

## 🏗️ Architecture

```mermaid
graph LR
    Input[Input Data] --> Validate[Validation Layer]
    Validate --> Process[Processing Engine]
    Process --> Transform[Transformation]
    Transform --> Output[Output Data]

    style Input fill:#e1f5fe
    style Output fill:#e8f5e8
    style Process fill:#fff3e0
```

## 🧪 Examples

See our [examples directory](examples/) for complete working examples:

- **[Basic Usage](examples/basic/)** - Getting started guide
- **[Advanced Configuration](examples/advanced/)** - Complex use cases
- **[Node.js Integration](examples/nodejs/)** - Server-side usage
- **[Browser Usage](examples/browser/)** - Client-side implementation
- **[TypeScript Project](examples/typescript/)** - Full TypeScript project

## 🔧 Configuration

### Global Configuration

```typescript
import { configure } from '@user/library'

configure({
  // API configuration
  apiKey: process.env.API_KEY,
  endpoint: 'https://api.example.com',

  // Performance settings
  timeout: 10000,
  retryAttempts: 3,

  // Debug mode
  debug: process.env.NODE_ENV === 'development'
})
```

### Environment Variables

```env
# Optional: Library configuration
LIBRARY_API_KEY=your_api_key
LIBRARY_TIMEOUT=5000
LIBRARY_DEBUG=false
```

## 📊 Performance

| Metric | Value |
|--------|-------|
| Bundle Size | 2.1 KB (minified + gzipped) |
| Dependencies | 0 (zero dependencies) |
| Tree Shakeable | ✅ Yes |
| TypeScript Support | ✅ Full support |
| Browser Support | Chrome 60+, Firefox 55+, Safari 12+ |

## 🧪 Testing

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run performance benchmarks
npm run test:performance

# Test browser compatibility
npm run test:browser
```

## 🤝 Contributing

We welcome contributions! Please read our [Contributing Guide](CONTRIBUTING.md).

### Development Setup

```bash
# Clone repository
git clone https://github.com/user/library.git
cd library

# Install dependencies
npm install

# Build the library
npm run build

# Run tests
npm test

# Start development
npm run dev
```

## 📄 License

MIT © [Your Name](https://github.com/user)

## 🔗 Related Projects

- **[library-2](https://github.com/user/library-2)** - Companion library
- **[library-cli](https://github.com/user/library-cli)** - CLI tool
- **[library-examples](https://github.com/user/library-examples)** - Example projects
```

## 🎯 Usage

### From Agents

```python
# docs-manager agent
Skill("moai-readme-expert")

# Generate README from project analysis
readme_expert = ReadmeExpert("./project-path")

# Generate different README types
professional_readme = readme_expert.generate_readme(template='professional')
api_readme = readme_expert.generate_readme(template='api_focused')
library_readme = readme_expert.generate_readme(template='library')

# Validate README quality
quality_score = readme_expert.validate_readme_quality(professional_readme)
```

### CLI Integration

```bash
# Generate README automatically
npx @alfred/readme-expert generate --project ./my-project --template professional

# Validate existing README
npx @alfred/readme-expert validate --path ./README.md

# Update README with latest changes
npx @alfred/readme-expert update --project ./my-project --auto

# Generate README from package.json
npx @alfred/readme-expert from-package --path ./package.json
```

## 📚 Reference Materials

- [Make a README](https://www.makeareadme.com/)
- [Awesome README](https://github.com/matiassingers/awesome-readme)
- [GitHub README Guidelines](https://docs.github.com/en/repositories/managing-your-repositorys-customizing-your-repository/about-readmes)
- [Standard Readme](https://github.com/RichardLitt/standard-readme)

## ✅ Validation Checklist

- [x] Complete README template structure provided
- [x] Multiple project type templates included
- [x] Dynamic generation from project analysis
- [x] Badge generation and optimization
- [x] Architecture diagram integration
- [x] Performance metrics included
- [x] Testing documentation guidelines
- [x] Accessibility and usability best practices