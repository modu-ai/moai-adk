---
name: moai-streaming-ui
description: Enhanced streaming UI system with progress indicators and status displays
version: 1.0.0
modularized: false
last_updated: 2025-11-22
compliance_score: 61
auto_trigger_keywords:
  - streaming
  - ui
category_tier: special
---

# Enhanced Streaming UI System

## Quick Reference

### Core Capabilities
- Real-time progress indicators
- Status displays (success, warning, error)
- Interactive user prompts
- Workflow visualization
- Multi-step operation tracking

### When to Use
- Long-running operations (>5 seconds)
- Multi-step workflows
- File processing tasks
- Background operations
- Test execution monitoring

### Immediate Usage

```python
# Basic progress bar
Skill("moai-streaming-ui")
progress_bar(65, 100, "Installing dependencies")
# → [████████████████████                    ] 65%

# Success indicator
success_indicator("Build completed successfully")
# → ✅ Build completed successfully

# Workflow steps
workflow_steps([
    ("✅", "Setup completed"),
    ("🔄", "Running tests..."),
    ("⏸️", "Build pending")
])
```

---

## Implementation Guide

### Progress Indicators

```python
class ProgressIndicator:
    def show_progress(self, current: int, total: int, message: str):
        """Display progress bar with percentage"""
        
        percentage = (current / total) * 100
        bar_length = 40
        filled = int(bar_length * current / total)
        
        bar = "█" * filled + " " * (bar_length - filled)
        print(f"\r{message}: [{bar}] {percentage:.1f}%", end='', flush=True)
    
    def show_spinner(self, message: str):
        """Animated spinner for indeterminate progress"""
        
        frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']
        for frame in frames:
            print(f'\r{frame} {message}', end='', flush=True)
            time.sleep(0.1)
```

### Status Indicators

```python
class StatusIndicator:
    def success(self, message: str, details: str = None):
        """Display success status"""
        output = f"✅ {message}"
        if details:
            output += f" ({details})"
        print(output)
    
    def warning(self, message: str, action: str = None):
        """Display warning with optional action"""
        output = f"⚠️ {message}"
        if action:
            output += f" → Action: {action}"
        print(output)
    
    def error(self, message: str, recovery: str = None):
        """Display error with recovery suggestion"""
        output = f"❌ {message}"
        if recovery:
            output += f" → Recovery: {recovery}"
        print(output)
```

### Interactive Elements

```python
class InteractiveUI:
    def user_prompt(self, question: str, options: List[str]) -> str:
        """Display user prompt with options"""
        
        prompt = f"❓ {question} "
        prompt += "[" + "/".join([f"[{opt[0].upper()}]{opt[1:]}" for opt in options]) + "]"
        
        print(prompt)
        response = input().strip().lower()
        
        # Match first letter
        for opt in options:
            if response.startswith(opt[0].lower()):
                return opt
        
        return options[0]  # Default to first option
    
    def confirmation(self, action: str, details: str = None) -> bool:
        """Request confirmation for action"""
        
        prompt = f"⚠️ Confirm: {action}"
        if details:
            prompt += f" ({details})"
        prompt += "? [y/N] "
        
        response = input(prompt).strip().lower()
        return response in ['y', 'yes']
```

### Workflow Visualization

```python
class WorkflowDisplay:
    def show_workflow(self, steps: List[Tuple[str, str]]):
        """Display multi-step workflow status"""
        
        for i, (status, description) in enumerate(steps, 1):
            print(f"Step {i}: {status} {description}")
    
    def show_phase(self, phase_name: str, current: int, total: int, percentage: int):
        """Display current phase with progress"""
        
        bar = self._create_progress_bar(percentage)
        print(f"🎯 Phase {current}/{total}: {phase_name} {bar} {percentage}%")
    
    def _create_progress_bar(self, percentage: int, length: int = 30) -> str:
        """Create progress bar string"""
        
        filled = int(length * percentage / 100)
        return f"[{'█' * filled}{' ' * (length - filled)}]"
```

### Integration Patterns

```python
# Example: Test execution with progress
def run_tests_with_progress(tests: List[Test]):
    """Run tests with visual progress feedback"""
    
    Skill("moai-streaming-ui")
    show_operation_start("Running test suite", f"{len(tests)} tests")
    
    passed = 0
    failed = 0
    
    for i, test in enumerate(tests):
        # Show progress
        progress = (i + 1) / len(tests) * 100
        show_progress(progress, f"Test {i+1}/{len(tests)}")
        
        # Run test
        result = execute_test(test)
        
        # Update status
        if result.passed:
            passed += 1
            show_success(f"✓ {test.name}")
        else:
            failed += 1
            show_error(f"✗ {test.name}: {result.error}")
    
    # Final summary
    show_complete(f"Tests: {passed} passed, {failed} failed")
```

---

## Advanced Patterns

### Multi-Step Operations

```python
class MultiStepOperation:
    def __init__(self, steps: List[str]):
        self.steps = steps
        self.current = 0
    
    def execute(self):
        """Execute multi-step operation with progress"""
        
        Skill("moai-streaming-ui")
        show_workflow_start(self.steps)
        
        for i, step_name in enumerate(self.steps):
            self.current = i + 1
            
            # Show current step
            show_current_step(self.current, len(self.steps), step_name)
            
            try:
                # Execute step
                result = self._execute_step(step_name)
                
                # Success
                show_step_success(step_name)
            
            except Exception as e:
                # Error
                show_step_error(step_name, str(e))
                raise
        
        # Completion
        show_workflow_complete()
```

### File Operations

```python
def process_files_with_ui(files: List[Path], operation: Callable):
    """Process files with progress tracking"""
    
    Skill("moai-streaming-ui")
    show_operation_start(f"Processing {len(files)} files")
    
    for i, file_path in enumerate(files):
        # Calculate progress
        progress = (i + 1) / len(files) * 100
        
        # Show file progress
        show_file_progress(
            file_name=file_path.name,
            current=i + 1,
            total=len(files),
            percentage=progress
        )
        
        try:
            # Process file
            operation(file_path)
            show_file_success(file_path.name)
        
        except Exception as e:
            show_file_error(file_path.name, str(e))
```

### Error Handling

```python
class ErrorRecoveryUI:
    def show_retry(self, operation: str, attempt: int, max_attempts: int):
        """Display retry attempt"""
        
        message = f"🔄 {operation} - Retry {attempt}/{max_attempts}"
        if attempt > 1:
            message += " (previous attempts failed)"
        
        print(message)
    
    def show_recovery_options(self, error: str, options: List[str]):
        """Display recovery options"""
        
        print(f"❌ Error: {error}")
        print("🔧 Recovery options:")
        
        for i, option in enumerate(options, 1):
            print(f"   {i}. {option}")
        
        choice = input(f"Choose [1-{len(options)}]: ")
        return int(choice) - 1
```

### Performance Optimization

```python
# Update throttling
MIN_UPDATE_INTERVAL = 0.5  # 500ms minimum

last_update = 0

def update_progress(progress: int):
    """Throttled progress update"""
    
    global last_update
    current_time = time.time()
    
    if current_time - last_update < MIN_UPDATE_INTERVAL:
        return  # Skip update
    
    last_update = current_time
    show_progress(progress)

# Memory efficiency
MAX_DISPLAY_ITEMS = 50

def display_items(items: List[str]):
    """Display limited number of items"""
    
    if len(items) > MAX_DISPLAY_ITEMS:
        display = items[:MAX_DISPLAY_ITEMS]
        display.append(f"... and {len(items) - MAX_DISPLAY_ITEMS} more")
        items = display
    
    for item in items:
        print(item)
```

---

**End of Skill** | Rich visual feedback for enhanced user experience