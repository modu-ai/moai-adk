# Claude Code Python SDK

## Overview

The Claude Code Python SDK provides a native Python interface for building AI agents with advanced tool access and conversation management. Perfect for data science workflows, automation scripts, and Python applications.

## Installation

### Prerequisites

- Python 3.10 or newer
- Node.js 18+ (for Claude Code runtime)

### Install Steps

```bash
# Install Python SDK
pip install claude-code-sdk

# Install Claude Code runtime
npm install -g @anthropic-ai/claude-code
```

### Verify Installation

```python
import claude_code_sdk
print(claude_code_sdk.__version__)
```

## Authentication

Set up authentication before using the SDK:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Alternative Providers:**

```bash
export CLAUDE_CODE_USE_BEDROCK=1    # For AWS Bedrock
export CLAUDE_CODE_USE_VERTEX=1     # For Google Vertex AI
```

## Quick Start

### Simple Query

```python
import asyncio
from claude_code_sdk import query

async def main():
    async for message in query("Analyze this Python file for potential improvements"):
        if message.type == "result":
            print(message.result)

asyncio.run(main())
```

### Basic Client Usage

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def analyze_code():
    async with ClaudeSDKClient(
        options=ClaudeCodeOptions(
            system_prompt="You are a Python expert",
            max_turns=3
        )
    ) as client:

        await client.query("Review the main.py file for performance issues")

        async for message in client.receive_response():
            if hasattr(message, 'content'):
                for block in message.content:
                    if hasattr(block, 'text'):
                        print(block.text)

asyncio.run(analyze_code())
```

## Core Components

### ClaudeSDKClient

The main client for multi-turn conversations:

```python
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

# Configure the client
options = ClaudeCodeOptions(
    system_prompt="You are a data scientist",
    max_turns=5,
    allowed_tools=["Read", "Write", "Bash", "WebFetch"],
    permission_mode="acceptEdits",
    model="claude-3-5-sonnet",
    verbose=True
)

# Use as context manager (recommended)
async with ClaudeSDKClient(options=options) as client:
    # Your code here
    pass

# Or manage lifecycle manually
client = ClaudeSDKClient(options=options)
await client.start()
# ... use client ...
await client.close()
```

### Query Function

For simple, one-off queries:

```python
from claude_code_sdk import query

# Basic query
async for message in query("Explain this algorithm"):
    print(message)

# With options
async for message in query(
    prompt="Optimize database queries",
    options=ClaudeCodeOptions(
        max_turns=2,
        allowed_tools=["Read", "Grep"]
    )
):
    if message.type == "result":
        print(message.result)
```

## Configuration Options

### ClaudeCodeOptions

```python
from claude_code_sdk import ClaudeCodeOptions

options = ClaudeCodeOptions(
    # Core settings
    system_prompt="You are an expert Python developer",
    max_turns=10,                    # Limit conversation rounds

    # Tool control
    allowed_tools=["Read", "Edit", "Bash", "WebFetch"],
    permission_mode="acceptEdits",   # ask, acceptEdits, acceptAll, bypassPermissions

    # Model settings
    model="claude-3-5-sonnet",       # Override default model

    # Output control
    verbose=True,                    # Enable debug output
    stream=True,                     # Enable streaming responses

    # Session management
    resume_session_id=None,          # Resume specific session
    working_directories=["./", "../docs/"],  # Additional work dirs

    # Advanced
    environment_variables={"DEBUG": "1"},
    timeout=300                      # Operation timeout in seconds
)
```

### Permission Modes

- `ask`: Request approval for each tool (not practical for automation)
- `acceptEdits`: Auto-approve file operations, ask for others
- `acceptAll`: Approve all tools automatically
- `bypassPermissions`: Skip permission system (advanced use)

## Practical Examples

### Data Analysis Workflow

```python
import asyncio
import pandas as pd
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def data_analysis():
    options = ClaudeCodeOptions(
        system_prompt="You are a data scientist. Analyze data and provide insights.",
        allowed_tools=["Read", "Write", "Bash"],
        max_turns=10,
        permission_mode="acceptAll"
    )

    async with ClaudeSDKClient(options=options) as client:
        # Load and analyze data
        await client.query("""
        Load the sales_data.csv file and perform:
        1. Basic statistical analysis
        2. Identify trends and patterns
        3. Create visualizations
        4. Generate a summary report
        """)

        # Process the streaming response
        async for message in client.receive_response():
            if message.type == "tool_result":
                print(f"Tool: {message.tool_name}")
                print(f"Result: {message.result}")
            elif message.type == "result":
                print(f"Final result: {message.content}")

asyncio.run(data_analysis())
```

### Code Review Agent

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

class CodeReviewer:
    def __init__(self):
        self.options = ClaudeCodeOptions(
            system_prompt="""
            You are a senior code reviewer. For each file, check for:
            1. Code quality and maintainability
            2. Security vulnerabilities
            3. Performance issues
            4. Best practice violations
            Provide specific suggestions for improvement.
            """,
            allowed_tools=["Read", "Grep", "Glob", "Bash"],
            max_turns=5,
            permission_mode="acceptEdits"
        )

    async def review_files(self, file_patterns):
        async with ClaudeSDKClient(options=self.options) as client:
            for pattern in file_patterns:
                await client.query(f"Review all files matching: {pattern}")

                async for message in client.receive_response():
                    if message.type == "result":
                        yield {
                            'pattern': pattern,
                            'review': message.content,
                            'timestamp': message.timestamp
                        }

# Usage
async def main():
    reviewer = CodeReviewer()

    patterns = ["*.py", "*.js", "*.ts"]
    async for review in reviewer.review_files(patterns):
        print(f"Review for {review['pattern']}:")
        print(review['review'])
        print("-" * 50)

asyncio.run(main())
```

### Automated Testing

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def smart_testing():
    """
    Intelligent test generation and execution
    """
    options = ClaudeCodeOptions(
        system_prompt="You are a testing expert. Write comprehensive tests and fix failures.",
        allowed_tools=["Read", "Write", "Edit", "Bash"],
        permission_mode="acceptAll",
        max_turns=15
    )

    async with ClaudeSDKClient(options=options) as client:
        # Generate tests
        await client.query("""
        1. Analyze the codebase structure
        2. Identify functions that need tests
        3. Write comprehensive test cases using pytest
        4. Run the tests and fix any failures
        5. Ensure good test coverage
        """)

        # Track test results
        test_results = []

        async for message in client.receive_response():
            if message.type == "tool_result" and message.tool_name == "Bash":
                if "pytest" in message.input.get("command", ""):
                    test_results.append({
                        'command': message.input["command"],
                        'output': message.result,
                        'success': message.success
                    })

        return test_results

# Run the testing automation
asyncio.run(smart_testing())
```

### Document Generation

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

class DocumentationGenerator:
    def __init__(self, project_type="python"):
        self.options = ClaudeCodeOptions(
            system_prompt=f"""
            You are a technical writer specializing in {project_type} projects.
            Generate clear, comprehensive documentation including:
            1. API documentation
            2. Usage examples
            3. Installation instructions
            4. Contributing guidelines
            """,
            allowed_tools=["Read", "Write", "Glob", "Grep"],
            permission_mode="acceptEdits",
            max_turns=8
        )

    async def generate_docs(self, output_format="markdown"):
        async with ClaudeSDKClient(options=self.options) as client:
            await client.query(f"""
            Generate comprehensive documentation for this project in {output_format} format:
            1. Create/update README.md with overview and setup
            2. Generate API documentation from code comments
            3. Create usage examples
            4. Add contributing guidelines
            """)

            # Collect generated documents
            documents = []

            async for message in client.receive_response():
                if message.type == "tool_result" and message.tool_name == "Write":
                    documents.append({
                        'file_path': message.input.get('file_path'),
                        'content': message.input.get('content'),
                        'type': 'documentation'
                    })

            return documents

# Usage
async def main():
    doc_gen = DocumentationGenerator("python")
    docs = await doc_gen.generate_docs()

    for doc in docs:
        print(f"Generated: {doc['file_path']}")

asyncio.run(main())
```

## Advanced Features

### Session Management

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def session_management_example():
    # Start a session
    options = ClaudeCodeOptions(system_prompt="You are a debugging assistant")

    async with ClaudeSDKClient(options=options) as client:
        await client.query("Help me debug the authentication system")

        # Get session ID for later resumption
        session_id = client.session_id
        print(f"Session ID: {session_id}")

        # Continue conversation...
        await client.query("Check the login endpoint for issues")

    # Resume the session later
    resume_options = ClaudeCodeOptions(
        resume_session_id=session_id,
        system_prompt="You are a debugging assistant"
    )

    async with ClaudeSDKClient(options=resume_options) as client:
        await client.query("Now check the user registration flow")

asyncio.run(session_management_example())
```

### Custom Tool Restrictions

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def restricted_analysis():
    # Only allow read-only operations
    options = ClaudeCodeOptions(
        system_prompt="Analyze code without making changes",
        allowed_tools=["Read", "Grep", "Glob"],  # No editing tools
        max_turns=5
    )

    async with ClaudeSDKClient(options=options) as client:
        await client.query("Analyze the codebase for security vulnerabilities")

        async for message in client.receive_response():
            if message.type == "result":
                print(message.content)

asyncio.run(restricted_analysis())
```

### Stream Processing

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def stream_processing():
    options = ClaudeCodeOptions(
        system_prompt="Process large datasets incrementally",
        stream=True,
        allowed_tools=["Read", "Bash"]
    )

    async with ClaudeSDKClient(options=options) as client:
        await client.query("Process the large logfile.txt and extract error patterns")

        # Handle streaming responses in real-time
        async for message in client.receive_response():
            if message.type == "thinking":
                print(f"Thinking: {message.content[:100]}...")
            elif message.type == "tool_use":
                print(f"Using tool: {message.tool_name}")
            elif message.type == "partial_result":
                print(f"Partial: {message.content}")
            elif message.type == "result":
                print(f"Final: {message.content}")

asyncio.run(stream_processing())
```

## Error Handling

### Basic Error Handling

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions, ClaudeCodeError

async def error_handling_example():
    options = ClaudeCodeOptions(
        allowed_tools=["Read", "Bash"],
        permission_mode="acceptAll"
    )

    try:
        async with ClaudeSDKClient(options=options) as client:
            await client.query("Perform risky operation")

            async for message in client.receive_response():
                if message.type == "error":
                    print(f"Tool error: {message.error}")
                elif message.type == "result":
                    print(message.content)

    except ClaudeCodeError as e:
        print(f"Claude Code error: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")

asyncio.run(error_handling_example())
```

### Retry Logic

```python
import asyncio
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

async def retry_operation(max_retries=3):
    options = ClaudeCodeOptions(
        allowed_tools=["Bash"],
        permission_mode="acceptAll"
    )

    for attempt in range(max_retries):
        try:
            async with ClaudeSDKClient(options=options) as client:
                await client.query("Run unstable command that might fail")

                async for message in client.receive_response():
                    if message.type == "result":
                        return message.content  # Success

        except Exception as e:
            print(f"Attempt {attempt + 1} failed: {e}")
            if attempt == max_retries - 1:
                raise
            await asyncio.sleep(2 ** attempt)  # Exponential backoff

# Usage
try:
    result = asyncio.run(retry_operation())
    print(f"Success: {result}")
except Exception as e:
    print(f"All retries failed: {e}")
```

## Integration Patterns

### With FastAPI

```python
from fastapi import FastAPI, BackgroundTasks
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions
import asyncio

app = FastAPI()

class CodeAnalysisService:
    def __init__(self):
        self.options = ClaudeCodeOptions(
            system_prompt="You are a code analyzer API",
            allowed_tools=["Read", "Grep"],
            max_turns=3
        )

    async def analyze_repository(self, repo_path: str):
        async with ClaudeSDKClient(options=self.options) as client:
            await client.query(f"Analyze the repository at {repo_path}")

            results = []
            async for message in client.receive_response():
                if message.type == "result":
                    results.append(message.content)

            return results

service = CodeAnalysisService()

@app.post("/analyze")
async def analyze_code(repo_path: str, background_tasks: BackgroundTasks):
    analysis = await service.analyze_repository(repo_path)
    return {"analysis": analysis}
```

### With Jupyter Notebooks

```python
# In Jupyter cell
import asyncio
from claude_code_sdk import query, ClaudeCodeOptions

# Use in notebook with proper event loop handling
def run_claude_query(prompt):
    try:
        # Check if event loop is running
        loop = asyncio.get_event_loop()
        if loop.is_running():
            # Use nest_asyncio for Jupyter compatibility
            import nest_asyncio
            nest_asyncio.apply()
    except:
        pass

    async def _query():
        results = []
        async for message in query(prompt, options=ClaudeCodeOptions(
            allowed_tools=["Read", "Write"],
            max_turns=3
        )):
            if message.type == "result":
                results.append(message.content)
        return results

    return asyncio.run(_query())

# Usage in notebook
results = run_claude_query("Analyze this dataset and create visualizations")
for result in results:
    print(result)
```

### With Django Management Commands

```python
# management/commands/ai_audit.py
from django.core.management.base import BaseCommand
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions
import asyncio

class Command(BaseCommand):
    help = 'Perform AI-powered code audit'

    def add_arguments(self, parser):
        parser.add_argument('--app', type=str, help='Django app to audit')
        parser.add_argument('--fix', action='store_true', help='Auto-fix issues')

    def handle(self, *args, **options):
        asyncio.run(self.audit_code(options))

    async def audit_code(self, options):
        permission_mode = "acceptAll" if options['fix'] else "acceptEdits"

        claude_options = ClaudeCodeOptions(
            system_prompt="You are a Django expert auditing code quality",
            allowed_tools=["Read", "Edit", "Grep", "Glob"],
            permission_mode=permission_mode,
            max_turns=10
        )

        async with ClaudeSDKClient(options=claude_options) as client:
            if options['app']:
                query = f"Audit the Django app: {options['app']}"
            else:
                query = "Audit the entire Django project for issues"

            await client.query(query)

            async for message in client.receive_response():
                if message.type == "result":
                    self.stdout.write(message.content)
```

## Best Practices

### Resource Management

```python
# Always use context managers
async with ClaudeSDKClient(options=options) as client:
    # Client automatically cleaned up
    pass

# Or handle manually if needed
client = ClaudeSDKClient(options=options)
try:
    await client.start()
    # Use client
finally:
    await client.close()
```

### Performance Optimization

```python
# Limit turns for cost control
options = ClaudeCodeOptions(max_turns=5)

# Use specific tools only
options = ClaudeCodeOptions(allowed_tools=["Read", "Grep"])

# Enable streaming for real-time feedback
options = ClaudeCodeOptions(stream=True)

# Set timeouts for long operations
options = ClaudeCodeOptions(timeout=300)
```

### Security Considerations

```python
# Restrict file access
options = ClaudeCodeOptions(
    allowed_tools=["Read", "Grep"],  # No write operations
    working_directories=["./safe_dir/"]  # Limit access scope
)

# Use read-only analysis mode
options = ClaudeCodeOptions(
    permission_mode="ask",  # Require explicit approval
    allowed_tools=["Read", "Grep", "Glob"]  # No modification tools
)
```

### Logging and Monitoring

```python
import logging
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def monitored_analysis():
    options = ClaudeCodeOptions(verbose=True)  # Enable debug output

    async with ClaudeSDKClient(options=options) as client:
        logger.info("Starting code analysis")

        await client.query("Analyze codebase for issues")

        async for message in client.receive_response():
            if message.type == "tool_use":
                logger.info(f"Tool used: {message.tool_name}")
            elif message.type == "error":
                logger.error(f"Error occurred: {message.error}")
            elif message.type == "result":
                logger.info("Analysis completed")
                return message.content

result = asyncio.run(monitored_analysis())
```

This comprehensive guide covers the Python SDK's capabilities for building sophisticated AI agents with Claude Code's powerful tool ecosystem.
