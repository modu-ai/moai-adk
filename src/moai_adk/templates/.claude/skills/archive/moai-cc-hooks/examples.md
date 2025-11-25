# Claude Code Hooks - Practical Examples

## Example 1: Pre-commit Hook

```python
# hooks/pre_commit_hook.py
from moai_cc_hooks import Hook, HookRegistry

class PreCommitHook(Hook):
    """Pre-commit validation hook"""

    event = "pre-commit"
    priority = 100

    async def execute(self, context):
        """Pre-commit validation"""

        # 1. Linting validation
        lint_result = await self.run_linter()
        if not lint_result.passed:
            return {
                "success": False,
                "message": f"Linting failed: {lint_result.errors}"
            }

        # 2. Run tests
        test_result = await self.run_tests()
        if not test_result.passed:
            return {
                "success": False,
                "message": f"Tests failed: {test_result.failures}"
            }

        # 3. Type checking
        type_check = await self.run_type_checker()
        if not type_check.passed:
            return {
                "success": False,
                "message": f"Type validation failed: {type_check.errors}"
            }

        return {"success": True}

    async def run_linter(self):
        # Linting logic
        pass

    async def run_tests(self):
        # Test logic
        pass

    async def run_type_checker(self):
        # Type validation logic
        pass

# Register
registry = HookRegistry()
registry.register(PreCommitHook())
```

## Example 2: Post-merge Hook

```python
# hooks/post_merge_hook.py
from moai_cc_hooks import Hook

class PostMergeHook(Hook):
    """Post-merge automation hook"""

    event = "post-merge"

    async def execute(self, context):
        """Post-merge processing"""

        merged_branch = context['source_branch']

        # 1. Update dependencies
        await self.update_dependencies()

        # 2. Generate documentation
        await self.generate_documentation()

        # 3. Trigger deployment
        if merged_branch == "main":
            await self.trigger_deployment()

        return {"success": True}
```

## Example 3: Custom Hook

```python
# hooks/custom_validation_hook.py
from moai_cc_hooks import Hook

class CustomValidationHook(Hook):
    """Custom validation hook"""

    event = "pre-push"

    async def execute(self, context):
        """Custom validation before push"""

        # Check SPEC file existence
        spec_files = await self.find_spec_files()
        if not spec_files:
            return {
                "success": False,
                "message": "SPEC file is required"
            }

        # Check test coverage
        coverage = await self.check_coverage()
        if coverage < 80:
            return {
                "success": False,
                "message": f"Test coverage insufficient: {coverage}%"
            }

        return {"success": True}
```

## Example 4: Hook Chain

```python
# hooks/hook_chain.py
from moai_cc_hooks import HookChain

class ValidationChain(HookChain):
    """Validation chain"""

    async def execute(self):
        """Execute validation chain"""

        hooks = [
            LintingHook(),
            TestingHook(),
            SecurityScanHook(),
            CoverageCheckHook()
        ]

        results = []
        for hook in hooks:
            result = await hook.execute()
            results.append(result)

            # Stop on failure
            if not result.success:
                return {
                    "success": False,
                    "failed_at": hook.name,
                    "results": results
                }

        return {"success": True, "results": results}
```

## Example 5: Async Hook

```python
# hooks/async_notification_hook.py
from moai_cc_hooks import Hook

class AsyncNotificationHook(Hook):
    """Async notification hook"""

    event = "post-deploy"
    async_execution = True

    async def execute(self, context):
        """Async notification after deployment"""

        # Slack notification
        await self.send_slack_notification(
            f"Deployment complete: {context['version']}"
        )

        # Email notification
        await self.send_email_notification(
            recipients=context['team'],
            message=f"Version {context['version']} deployed"
        )

        # Monitoring update
        await self.update_monitoring(
            version=context['version'],
            status="deployed"
        )

        return {"success": True}
```

## Example 6: Conditional Hook

```python
# hooks/conditional_hook.py
from moai_cc_hooks import Hook

class ConditionalHook(Hook):
    """Conditional execution hook"""

    event = "pre-build"

    async def should_execute(self, context):
        """Check execution condition"""

        # Execute only on main branch
        if context['branch'] != 'main':
            return False

        # Execute only when code changes exist
        if not context['files_changed']:
            return False

        return True

    async def execute(self, context):
        """Execute when condition is met"""

        if not await self.should_execute(context):
            return {"skipped": True}

        await self.run_full_build()

        return {"success": True}
```

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
