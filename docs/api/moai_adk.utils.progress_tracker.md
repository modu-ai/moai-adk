# moai_adk.utils.progress_tracker

@FEATURE:PROGRESS-TRACKER-001 Progress tracking and display utilities for MoAI-ADK installation
@TASK:UI-PROGRESS-001 Provides clean, professional progress display with color coding and customizable progress callbacks for different UI contexts

This module handles:
- Real-time progress visualization with progress bars
- Multi-stage progress tracking for complex operations
- Color-coded status indicators
- Customizable progress callbacks for different UI contexts

## Functions

### __init__

```python
__init__(self)
```

### update_progress

@TASK:PROGRESS-UPDATE-001 Update installation progress with clean, professional display

Args:
    step: Description of current step
    callback: Optional custom progress callback

```python
update_progress(self, step, callback)
```

### _display_default_progress

Display default progress with color coding and progress bar.

```python
_display_default_progress(self, step)
```

### _get_progress_color

Get color based on progress percentage.

```python
_get_progress_color(self, percentage)
```

### _create_progress_bar

Create clean visual progress bar with accurate calculation.

```python
_create_progress_bar(self, width)
```

### reset

Reset progress tracker to initial state.

```python
reset(self)
```

### set_total_steps

Update total steps count.

```python
set_total_steps(self, total_steps)
```

### get_progress_info

Get current progress information as dictionary.

```python
get_progress_info(self)
```

### is_complete

Check if progress is complete.

```python
is_complete(self)
```

### skip_to_step

Skip directly to a specific step number.

```python
skip_to_step(self, step_number)
```

### add_steps

Add additional steps to the total.

```python
add_steps(self, additional_steps)
```

### create_sub_tracker

Create a sub-tracker for nested operations.

Args:
    sub_steps: Number of steps in the sub-operation

Returns:
    ProgressTracker: New tracker for sub-operation

```python
create_sub_tracker(self, sub_steps)
```

### display_completion_message

Display completion message with formatting.

```python
display_completion_message(self, message)
```

### display_error_message

Display error message with formatting.

```python
display_error_message(self, message)
```

### display_warning_message

Display warning message with formatting.

```python
display_warning_message(self, message)
```

### add_stage

Add a new stage with specified step count.

```python
add_stage(self, stage_name, steps)
```

### start_stage

Start a specific stage and return its tracker.

```python
start_stage(self, stage_name)
```

### complete_stage

Mark a stage as completed.

```python
complete_stage(self, stage_name)
```

### get_overall_progress

Get overall progress across all stages.

```python
get_overall_progress(self)
```

## Classes

### ProgressTracker

@TASK:PROGRESS-MAIN-001 Handles progress tracking and display for installation processes

Provides clean, color-coded progress visualization for MoAI-ADK operations.

### MultiStageProgressTracker

@TASK:MULTI-STAGE-PROGRESS-001 Progress tracker that handles multiple stages with different step counts

Manages complex operations that span multiple phases with independent step counts.
