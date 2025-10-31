# Contributing to Backend Plugin

Thank you for your interest in contributing to the MoAI Backend Plugin! This document provides guidelines and instructions for contributing.

## Code of Conduct

We are committed to providing a welcoming and inspiring community for all. Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

### Prerequisites
- Python 3.13+
- FastAPI knowledge
- Git and GitHub familiarity

### Development Setup

```bash
# Clone the repository
git clone https://github.com/moai-adk/moai-marketplace.git
cd moai-marketplace/plugins/moai-plugin-backend

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install development dependencies
pip install -e ".[dev]"
```

### Running Tests

```bash
# Run all tests
pytest tests/ -v

# Run with coverage
pytest tests/ --cov=src --cov-report=html

# Run specific test file
pytest tests/test_fastapi_patterns.py -v
```

### Code Quality Checks

```bash
# Type checking with mypy
mypy src/

# Linting with ruff
ruff check src/

# Format code
ruff format src/
```

## Making Changes

### Branch Naming
- Feature: `feature/agent-name-description`
- Bug fix: `fix/short-description`
- Documentation: `docs/short-description`

Example: `feature/database-expert-pooling`

### Commit Messages

Follow conventional commits format:

```
type(scope): short description

Longer description explaining the change and why.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Process

1. **Create Branch**: `git checkout -b feature/your-feature`
2. **Make Changes**: Write code following style guidelines
3. **Add Tests**: Include tests for new functionality
4. **Run Tests**: Ensure all tests pass
5. **Push**: `git push origin feature/your-feature`
6. **Create PR**: Open pull request on GitHub
7. **Description**: Fill in PR template with details
8. **Review**: Address reviewer feedback

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New skill content
- [ ] Agent enhancement
- [ ] Bug fix
- [ ] Documentation update

## Related Issues
Fixes #(issue number)

## Testing
Describe how to test these changes

## Checklist
- [ ] Code follows style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
```

## Writing Skills

### Skill Structure

Each skill should follow this structure:

```markdown
# moai-domain-example

Brief description of what this skill covers.

## Quick Start

When to use this skill and basic overview.

## Core Patterns

### Pattern 1: [Pattern Name]

**Pattern**: Brief description of the pattern.

\`\`\`python
# Code example showing the pattern
\`\`\`

**When to use**:
- Specific use case 1
- Specific use case 2

**Key benefits**:
- Benefit 1
- Benefit 2

### Pattern 2: [Pattern Name]
...

## Progressive Disclosure

### Level 1: Basics
- Item 1
- Item 2

### Level 2: Advanced
- Item 1
- Item 2

### Level 3: Expert
- Item 1
- Item 2

## Works Well With

- **Library 1**: Description
- **Library 2**: Description

## References

- **Documentation**: URL
- **Examples**: URL
```

### Code Examples
- Include complete, runnable examples
- Add inline comments explaining key concepts
- Show realistic use cases
- Follow Python best practices (PEP 8)

## Documentation

### Updating Docs
- Keep README.md in sync with changes
- Update docstrings in code
- Add examples for new features
- Fix typos and improve clarity

### Style Guide
- Use clear, concise language
- Include code examples where helpful
- Link to related documentation
- Maintain consistent formatting

## Testing

### Test Requirements
- Write tests for all new functionality
- Achieve minimum 85% code coverage
- Use pytest fixtures for setup
- Test both success and error paths

### Test Example

```python
import pytest
from src.fastapi_patterns import create_app

@pytest.fixture
def app():
    return create_app()

def test_async_endpoint(app):
    """Test async endpoint handler."""
    with app.test_client() as client:
        response = client.get("/api/data")
        assert response.status_code == 200
        assert "data" in response.json()
```

## Reporting Issues

### Bug Reports
- Include Python version and OS
- Provide minimal reproduction case
- Include full error traceback
- Describe expected vs actual behavior

### Feature Requests
- Explain the use case
- Provide examples if possible
- Discuss alternatives considered
- Explain benefits and impact

## Community

- **Discussions**: Ask questions in GitHub Discussions
- **Discord**: Join our community Discord server
- **Email**: Contact team@example.com

## Recognition

Contributors are recognized in:
- Release notes
- Contributors list in README
- Special thanks section

## Questions?

Feel free to:
- Open an issue with your question
- Join our Discord community
- Email the maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to making MoAI Backend Plugin better! 🎉
