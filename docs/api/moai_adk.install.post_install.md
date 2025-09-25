# moai_adk.install.post_install

@FEATURE:POST-INSTALL-001 🗿 MoAI-ADK Post-Installation Script

@TASK:POST-INSTALL-001 Provides post-installation functionality for MoAI-ADK.
@TASK:POST-INSTALL-002 Since pyproject.toml doesn't support post-install hooks directly,
this script serves as an alternative for setting up global resources.

Usage:
    pip install moai-adk
    moai-install  # Run after installation

## Functions

### print_post_install_banner

Print post-installation banner.

```python
print_post_install_banner()
```

### main

MoAI-ADK post-installation setup.

Note: As of v0.1.13+, resources are embedded in the package.
No separate installation needed.

```python
main(force, quiet)
```

### auto_install_on_first_run

Check if resources are available (실제 리소스 존재 여부 검증).

Returns:
    bool: 리소스가 실제로 사용 가능한지 여부

```python
auto_install_on_first_run()
```
