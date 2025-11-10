# Contributing

Thank you for your interest in the MoAI-ADK project. Your contributions make the project even stronger.

## How to Contribute

### Bug Reports

Found a bug? Let us know via GitHub Issues:

1. **Search**: Check if the bug has already been reported
2. **Details**: Include reproduction steps, environment info, screenshots
3. **Minimal Example**: Minimal code that demonstrates the issue

### Feature Suggestions

Have a new feature idea?

1. **Problem Definition**: Clearly explain what problem it solves
2. **Solution**: Analyze pros and cons of proposed solution
3. **Alternatives**: Consider other possible solutions

### Code Contributions

Want to contribute code directly?

1. **Fork**: Fork the repository
2. **Branch**: Create a branch per feature (`git checkout -b feature/amazing-feature`)
3. **Commit**: Commit changes (`git commit -m 'Add amazing feature'`)
4. **Push**: Push branch (`git push origin feature/amazing-feature`)
5. **PR**: Create Pull Request

## Development Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 2. Install Development Environment

```bash
# Using UV (recommended)
uv sync

# Or using pip
pip install -e ".[dev]"
```

### 3. Run Tests

```bash
pytest

# With coverage
pytest --cov=moai_adk
```

### 4. Code Style Checks

```bash
# Code formatting
black .
ruff check .
ruff format .

# Type checking
mypy .
```

## Documentation Contributions

Documentation improvements are also important contributions:

- **Fix Typos**: Fix typos or grammar errors you find
- **Translation**: Translate documentation to other languages
- **Add Examples**: Add more usage examples
- **Improve Explanations**: Explain complex concepts more clearly

## Coding Standards

- **PEP 8**: Follow Python style guide
- **Type Hints**: Include type hints for all functions
- **Docstrings**: Include docstrings for all modules and functions
- **Tests**: Include test code for new features

## Community

- **GitHub Discussions**: Questions and discussions
- **Issues**: Bug reports and feature suggestions
- **Pull Requests**: Code reviews and collaboration

## License

Code you contribute will be distributed under the project's MIT license.




