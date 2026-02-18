# MoAI Worktree Reference

Purpose: External resources, documentation, and additional learning materials for moai-worktree skill.

Version: 1.0.0
Last Updated: 2025-11-29

---

## External Documentation

### Git Worktree Official Documentation

- Git Worktree Documentation: [https://git-scm.com/docs/git-worktree](https://git-scm.com/docs/git-worktree)
 - Official Git worktree command reference
 - Advanced worktree patterns and workflows
 - Troubleshooting and best practices

- Pro Git Book - Worktrees: [https://git-scm.com/book/en/v2/Git-Tools-Worktree](https://git-scm.com/book/en/v2/Git-Tools-Worktree)
 - Comprehensive guide to Git worktrees
 - Multiple workflow management strategies
 - Worktree lifecycle management

### Development Tools Documentation

- VS Code Multi-Root Workspaces: [https://code.visualstudio.com/docs/editor/multi-root-workspaces](https://code.visualstudio.com/docs/editor/multi-root-workspaces)
 - Setting up multi-root workspaces for worktree development
 - Workspace configuration and task automation
 - Extension management across worktrees

- Click Framework: [https://click.palletsprojects.com/](https://click.palletsprojects.com/)
 - Command-line interface framework used by moai-worktree
 - Advanced CLI patterns and argument parsing
 - Custom command development

### Python Development Resources

- Rich Library: [https://rich.readthedocs.io/](https://rich.readthedocs.io/)
 - Terminal output formatting used by moai-worktree
 - Tables, progress bars, and syntax highlighting
 - Advanced terminal UI patterns

- Pathlib Documentation: [https://docs.python.org/3/library/pathlib.html](https://docs.python.org/3/library/pathlib.html)
 - Modern path manipulation for cross-platform compatibility
 - File system operations and directory traversal
 - Path validation and security considerations

---

## Related Skills and Integrations

### MoAI-ADK Skills

- moai-foundation-core: Foundation principles and delegation patterns
 - Reference: [moai-foundation-core modules](moai-foundation-core/modules/)
 - Integration: Worktree creation follows foundation delegation patterns

- moai-workflow-project: Project management and configuration
 - Reference: [moai-workflow-project modules](moai-workflow-project/modules/)
 - Integration: Project setup with worktree support

- moai-foundation-claude: Claude Code execution patterns
 - Reference: [moai-foundation-claude modules](moai-foundation-claude/modules/)
 - Integration: Command and agent execution patterns

### Complementary Tools

- Git Hooks: Custom Git hooks for worktree automation
 - Pre-commit hooks for worktree validation
 - Post-checkout hooks for environment setup
 - Pre-push hooks for worktree synchronization

- Docker: Containerized development environments
 - Docker Compose for service orchestration in worktrees
 - Volume mounts for persistent data across worktrees
 - Environment isolation for different worktree configurations

- Makefiles: Build automation across worktrees
 - Parallel build processes for multiple worktrees
 - Shared build targets and dependencies
 - Cleanup and optimization automation

---

## Community Resources

### Open Source Projects

- Git Worktree Managers:
 - [git-worktree](https://github.com/charmbracelet/git-worktree) - Go implementation
 - [git-worktree.nvim](https://github.com/ThePrimeagen/git-worktree.nvim) - Neovim plugin
 - [git-worktree.el](https://github.com/magit/git-worktree.el) - Emacs integration

### Blog Posts and Articles

- Parallel Development with Git Worktrees: [Medium Article](https://medium.com/@username/parallel-development-git-worktrees)
- Git Worktree Best Practices: [Dev.to Article](https://dev.to/username/git-worktree-best-practices)
- Advanced Worktree Workflows: [GitHub Gist](https://gist.github.com/username/worktree-patterns)

### Video Tutorials

- Git Worktree Tutorial: [YouTube Tutorial](https://www.youtube.com/watch?v=worktree-tutorial)
- Parallel Development Setup: [Screencast](https://www.youtube.com/watch?v=parallel-dev-setup)
- Worktree Automation: [Conference Talk](https://www.youtube.com/watch?v=worktree-automation)

---

## Performance and Optimization

### Performance Monitoring Tools

- htop: Process monitoring for worktree resource usage
 - Monitor CPU and memory usage across worktrees
 - Identify resource-intensive worktree operations
 - Track disk usage and I/O patterns

- ncdu: Disk usage analysis for worktree cleanup
 - Analyze disk usage patterns across worktrees
 - Identify large files and directories for optimization
 - Generate cleanup recommendations

- git-branchless: Faster Git operations for large repositories
 - Optimize performance for worktree operations
 - Reduce overhead for large codebase worktrees
 - Improve sync performance across multiple worktrees

### Optimization Techniques

1. Shallow Worktrees: For fast prototyping and testing
 ```bash
 moai-worktree new SPEC-PROTO-001 "Prototype" --shallow --depth 1
 ```

2. Selective Synchronization: Sync only essential files
 ```bash
 moai-worktree sync SPEC-001 --include "src/" --exclude "node_modules/"
 ```

3. Background Operations: Non-blocking worktree operations
 ```bash
 moai-worktree sync --all --background
 ```

---

## Security Considerations

### Worktree Security Best Practices

1. Isolation: Ensure worktrees don't share sensitive data
 - Use separate environment files (`.env.local`) for each worktree
 - Avoid sharing API keys or credentials between worktrees
 - Regularly audit worktree configurations

2. Access Control: Implement proper file permissions
 - Set appropriate permissions on worktree directories (755)
 - Restrict access to sensitive worktree data
 - Use Git hooks to prevent accidental secrets commits

3. Backup Strategy: Implement worktree backup and recovery
 - Regular backups of worktree registry and configurations
 - Version control for worktree templates and scripts
 - Recovery procedures for corrupted worktrees

### Security Tools

- git-secrets: Detect sensitive data in worktrees
 - Scan worktrees for accidental secrets or credentials
 - Integrate with pre-commit hooks for worktree security
 - Automatic remediation suggestions

- truffleHog: Security scanning for worktree files
 - Comprehensive security audit across worktrees
 - API key and credential detection
 - Custom pattern matching for project-specific secrets

---

## Integration Examples

### CI/CD Integration

#### GitHub Actions Workflow

```yaml
# .github/workflows/worktree-testing.yml
name: Worktree Testing

on:
 push:
 branches: [ "feature/SPEC-*" ]

jobs:
 test-worktree:
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v3
 with:
 fetch-depth: 0

 - name: Setup moai-worktree
 run: |
 go install github.com/modu-ai/moai-adk/cmd/moai@latest
 echo "Setting up worktree environment..."

 - name: Test Worktree Operations
 run: |
 # Test worktree creation
 moai-worktree new test-spec "Test Worktree"

 # Test worktree synchronization
 moai-worktree sync test-spec

 # Test worktree cleanup
 moai-worktree remove test-spec
```

#### Jenkins Pipeline

```groovy
// Jenkinsfile for worktree testing
pipeline {
 agent any

 stages {
 stage('Setup') {
 steps {
 sh '''
 go install github.com/modu-ai/moai-adk/cmd/moai@latest
 moai-worktree config set worktree_root $WORKSPACE/worktrees
 '''
 }
 }

 stage('Test') {
 parallel {
 stage('Auth Worktree') {
 steps {
 sh '''
 moai-worktree new SPEC-AUTH-001 "Authentication Worktree"
 cd $(moai-worktree go SPEC-AUTH-001)
 npm test
 '''
 }
 }

 stage('Payment Worktree') {
 steps {
 sh '''
 moai-worktree new SPEC-PAY-001 "Payment Worktree"
 cd $(moai-worktree go SPEC-PAY-001)
 npm test
 '''
 }
 }
 }
 }

 stage('Cleanup') {
 steps {
 sh '''
 moai-worktree clean --force
 rm -rf $WORKSPACE/worktrees
 '''
 }
 }
 }
}
```

---

## Troubleshooting Resources

### Common Error Patterns

1. Permission Denied Errors:
 - Check file permissions on worktree directories
 - Verify user has write access to worktree root
 - Ensure Git repository permissions are correct

2. Disk Space Issues:
 - Monitor disk usage with `df -h`
 - Clean up large files with `ncdu`
 - Implement automated cleanup schedules

3. Network Connectivity:
 - Verify Git remote repository accessibility
 - Check network connectivity for external dependencies
 - Use offline worktree operations when possible

### Debugging Tools

- git worktree prune: Clean up stale worktree references
- git fsck: Repository integrity checking
- strace: System call tracing for debugging
- lsof: Open file tracking and process monitoring

---

## Version History

### Current Version (1.0.0)
- Initial release with core worktree management features
- Integration with MoAI-ADK workflow commands
- Comprehensive documentation and examples

### Future Roadmap
- v1.1.0: Team collaboration features
 - Shared worktree registries
 - Multi-developer coordination
 - Conflict resolution algorithms

- v1.2.0: Advanced automation features
 - Worktree templates marketplace
 - Custom automation scripts
 - Performance optimization tools

- v1.3.0: Cloud integration
 - Cloud-based worktree storage
 - Remote worktree synchronization
 - Distributed development workflows

---

## Contributing

### How to Contribute

1. Report Issues: Submit bug reports and feature requests
2. Submit Pull Requests: Contribute code and documentation improvements
3. Share Examples: Add usage examples and integration patterns
4. Document Workflows: Create tutorials and guides

### Development Setup

```bash
# Clone the repository
git clone https://github.com/your-org/moai-adk.git

# Set up development environment
cd moai-adk
pip install -e .
pip install -r requirements-dev.txt

# Run tests
pytest tests/test_worktree.py -v

# Build documentation
mkdocs build
```

---

Version: 1.0.0
Last Updated: 2025-11-29
Reference: External resources and additional learning materials for moai-worktree
