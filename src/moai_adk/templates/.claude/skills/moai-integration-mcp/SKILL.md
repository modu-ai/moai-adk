---
name: moai-integration-mcp
description: Universal MCP integration specialist combining connector patterns for Figma, Notion, Nano-Banana, and custom MCP server development
version: 1.0.0
category: integration
tags:
  - integration
  - mcp
  - figma
  - notion
  - connectors
  - server-development
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# MCP Integration Specialist

## Quick Reference (30 seconds)

**Universal MCP Integration** - Comprehensive MCP development combining all connector patterns (Figma, Notion, Nano-Banana) with custom server development capabilities.

**Core Capabilities**:
- 🔌 **Universal Connectors**: Figma, Notion, Nano-Banana integration patterns
- 🛠️ **MCP Server Development**: FastMCP framework, custom tools and resources
- 🎨 **Design Integration**: Figma design systems, component extraction, tokens
- 📝 **Knowledge Management**: Notion databases, docs, wikis integration
- 🤖 **AI Services**: Nano-Banana AI model integration and orchestration
- 🔒 **Security Patterns**: OAuth2, API keys, authentication workflows

**When to Use**:
- Building custom MCP servers for internal tools
- Integrating external services via MCP protocol
- Design-to-code workflows with Figma
- Knowledge base integration with Notion
- AI service integration and orchestration

---

## Implementation Guide

### Universal MCP Server Architecture

**Base MCP Server with Multiple Connectors**:
```python
from fastmcp import FastMCP
from typing import Dict, List, Optional
import asyncio

class UniversalMCPServer:
    def __init__(self, server_name: str):
        self.server = FastMCP(server_name)
        self.connectors = {}
        self.setup_connectors()
        self.register_tools()

    def setup_connectors(self):
        """Initialize all available connectors."""
        self.connectors = {
            'figma': FigmaConnector(),
            'notion': NotionConnector(),
            'nano_banana': NanoBananaConnector(),
            'custom': CustomConnector()
        }

    def register_tools(self):
        """Register all tools from connectors."""
        for connector_name, connector in self.connectors.items():
            connector.register_tools(self.server)

# Figma Connector
class FigmaConnector:
    def __init__(self, api_key: str = None):
        self.api_key = api_key
        self.figma_client = None

    async def initialize(self):
        if self.api_key:
            self.figma_client = FigmaClient(self.api_key)

    def register_tools(self, server):
        @server.tool()
        async def extract_figma_components(
            file_id: str,
            component_names: Optional[List[str]] = None
        ) -> dict:
            """Extract React components from Figma design."""
            if not self.figma_client:
                raise ValueError("Figma API key not configured")

            file_data = await self.figma_client.get_file(file_id)
            components = self.parse_components(file_data, component_names)

            return {
                "components": components,
                "design_tokens": self.extract_design_tokens(file_data),
                "metadata": self.extract_metadata(file_data)
            }

        @server.tool()
        async def sync_figma_tokens(
            file_id: str,
            output_path: str = "src/styles/tokens.json"
        ) -> dict:
            """Sync design tokens from Figma to code."""
            file_data = await self.figma_client.get_file(file_id)
            tokens = self.extract_design_tokens(file_data)

            # Convert to W3C DTCG 2.0 format
            formatted_tokens = self.format_design_tokens(tokens)

            return {
                "tokens": formatted_tokens,
                "file_path": output_path,
                "sync_status": "success"
            }

# Notion Connector
class NotionConnector:
    def __init__(self, api_key: str = None):
        self.api_key = api_key
        self.notion_client = None

    async def initialize(self):
        if self.api_key:
            self.notion_client = NotionClient(auth=self.api_key)

    def register_tools(self, server):
        @server.tool()
        async def query_notion_database(
            database_id: str,
            query: Optional[Dict] = None
        ) -> dict:
            """Query Notion database with filtering."""
            if not self.notion_client:
                raise ValueError("Notion API key not configured")

            database = self.notion_client.databases.retrieve(database_id)
            results = self.notion_client.databases.query(
                database_id=database_id,
                query=query or {}
            )

            return {
                "results": results.get("results", []),
                "database_info": database,
                "total_count": len(results.get("results", []))
            }

        @server.tool()
        async def create_notion_page(
            database_id: str,
            properties: Dict,
            content: Optional[List[Dict]] = None
        ) -> dict:
            """Create new page in Notion database."""
            page_data = {
                "parent": {"database_id": database_id},
                "properties": properties
            }

            if content:
                page_data["children"] = content

            new_page = self.notion_client.pages.create(**page_data)
            return {"page": new_page, "status": "created"}

# Nano-Banana Connector
class NanoBananaConnector:
    def __init__(self, api_key: str = None):
        self.api_key = api_key
        self.nb_client = None

    async def initialize(self):
        if self.api_key:
            self.nb_client = NanoBananaClient(api_key=self.api_key)

    def register_tools(self, server):
        @server.tool()
        async def generate_ai_content(
            prompt: str,
            model: str = "claude-3-5-sonnet-20241022",
            max_tokens: int = 4000
        ) -> dict:
            """Generate content using Nano-Banana AI models."""
            if not self.nb_client:
                raise ValueError("Nano-Banana API key not configured")

            response = await self.nb_client.chat.completions.create(
                model=model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=max_tokens
            )

            return {
                "content": response.choices[0].message.content,
                "model": model,
                "usage": response.usage
            }

        @server.tool()
        async def analyze_with_ai(
            content: str,
            analysis_type: str = "summary"
        ) -> dict:
            """Analyze content using AI models."""
            analysis_prompts = {
                "summary": "Please provide a comprehensive summary of the following content:",
                "sentiment": "Analyze the sentiment and emotional tone of this content:",
                "key_points": "Extract the key points and main ideas from this content:",
                "action_items": "Identify action items and next steps from this content:"
            }

            prompt = f"{analysis_prompts.get(analysis_type, analysis_prompts['summary'])}\n\n{content}"

            response = await self.nb_client.chat.completions.create(
                model="claude-3-5-sonnet-20241022",
                messages=[{"role": "user", "content": prompt}],
                max_tokens=2000
            )

            return {
                "analysis": response.choices[0].message.content,
                "type": analysis_type,
                "content_length": len(content)
            }
```

### Advanced Integration Patterns

**Multi-Service Orchestration**:
```python
class ServiceOrchestrator:
    def __init__(self, mcp_server: UniversalMCPServer):
        self.mcp_server = mcp_server
        self.workflow_engine = WorkflowEngine()

    def register_orchestration_tools(self):
        @self.mcp_server.server.tool()
        async def design_to_code_workflow(
            figma_file_id: str,
            component_library: str = "shadcn"
        ) -> dict:
            """Complete design-to-code workflow from Figma to React components."""

            # Step 1: Extract components from Figma
            figma_result = await self.mcp_server.server.invoke_tool(
                "extract_figma_components",
                {"file_id": figma_file_id}
            )

            # Step 2: Generate React components
            components = figma_result["components"]
            generated_code = []

            for component in components:
                ai_prompt = f"""
                Generate a React component using {component_library} for this Figma design:
                Component Name: {component['name']}
                Design: {component['design_data']}
                Tokens: {component['tokens']}

                Requirements:
                - Use TypeScript
                - Include proper props interface
                - Apply design tokens
                - Make responsive with Tailwind CSS
                - Include accessibility attributes
                """

                ai_result = await self.mcp_server.server.invoke_tool(
                    "generate_ai_content",
                    {"prompt": ai_prompt, "max_tokens": 2000}
                )

                generated_code.append({
                    "name": component['name'],
                    "code": ai_result["content"],
                    "tokens": component['tokens']
                })

            # Step 3: Create documentation
            doc_prompt = f"""
            Create comprehensive documentation for these React components:
            {json.dumps([c['name'] for c in components], indent=2)}

            Include:
            - Component descriptions
            - Props documentation
            - Usage examples
            - Design token references
            """

            docs = await self.mcp_server.server.invoke_tool(
                "generate_ai_content",
                {"prompt": doc_prompt, "max_tokens": 3000}
            )

            return {
                "components": generated_code,
                "documentation": docs["content"],
                "workflow_status": "completed",
                "generated_files": len(generated_code)
            }

        @self.mcp_server.server.tool()
        async def knowledge_extraction_workflow(
            notion_database_id: str,
            analysis_goal: str = "extract_best_practices"
        ) -> dict:
            """Extract knowledge from Notion and create structured documentation."""

            # Step 1: Query Notion database
            notion_result = await self.mcp_server.server.invoke_tool(
                "query_notion_database",
                {"database_id": notion_database_id}
            )

            # Step 2: Analyze content with AI
            all_content = "\n\n".join([
                self.extract_text_from_page(page)
                for page in notion_result["results"]
            ])

            analysis_prompt = f"""
            Analyze this content for {analysis_goal}:

            Content:
            {all_content[:8000]}  # Limit content length

            Please provide:
            1. Key insights and patterns
            2. Structured summary
            3. Actionable recommendations
            4. Categorization of information
            """

            analysis = await self.mcp_server.server.invoke_tool(
                "analyze_with_ai",
                {"content": all_content, "analysis_type": analysis_goal}
            )

            # Step 3: Create structured knowledge base
            knowledge_base = self.structure_knowledge(
                analysis["analysis"],
                notion_result["results"]
            )

            return {
                "analysis": analysis["analysis"],
                "knowledge_base": knowledge_base,
                "source_count": len(notion_result["results"]),
                "workflow_status": "completed"
            }
```

### Resource Management

**Advanced Resource Patterns**:
```python
class ResourceManager:
    def __init__(self, mcp_server: UniversalMCPServer):
        self.mcp_server = mcp_server
        self.cache = {}

    def register_resources(self):
        @mcp_server.server.resource("figma://file/{file_id}/components")
        async def get_figma_components(file_id: str) -> dict:
            """Get Figma file components with caching."""
            cache_key = f"figma_components_{file_id}"

            if cache_key in self.cache:
                cached_data, timestamp = self.cache[cache_key]
                if time.time() - timestamp < 300:  # 5 minutes cache
                    return cached_data

            components = await self.mcp_server.connectors['figma'].get_components(file_id)
            self.cache[cache_key] = (components, time.time())

            return components

        @mcp_server.server.resource("notion://database/{database_id}/schema")
        async def get_notion_database_schema(database_id: str) -> dict:
            """Get Notion database schema with properties."""
            schema = await self.mcp_server.connectors['notion'].get_database_schema(database_id)
            return {
                "database_id": database_id,
                "schema": schema,
                "last_updated": time.time()
            }

        @mcp_server.server.resource("ai://models")
        async def get_available_ai_models() -> dict:
            """Get available AI models and capabilities."""
            models = await self.mcp_server.connectors['nano_banana'].list_models()
            return {
                "models": models,
                "capabilities": {
                    "text_generation": True,
                    "analysis": True,
                    "multimodal": True
                }
            }

        @mcp_server.server.resource("workflow://templates")
        async def get_workflow_templates() -> dict:
            """Get available workflow templates."""
            return {
                "design_to_code": {
                    "description": "Convert Figma designs to React components",
                    "steps": ["extract_components", "generate_code", "create_docs"]
                },
                "knowledge_extraction": {
                    "description": "Extract insights from Notion databases",
                    "steps": ["query_data", "analyze_content", "structure_knowledge"]
                },
                "content_generation": {
                    "description": "Generate content with AI assistance",
                    "steps": ["define_requirements", "generate_content", "review_and_refine"]
                }
            }
```

---

## Advanced Implementation

### Security and Authentication

**Multi-Provider Authentication**:
```python
class AuthManager:
    def __init__(self):
        self.providers = {
            'figma': FigmaAuthProvider(),
            'notion': NotionAuthProvider(),
            'nano_banana': NanoBananaAuthProvider()
        }

    async def authenticate_service(self, service: str, credentials: dict) -> bool:
        """Authenticate with specific service provider."""
        if service not in self.providers:
            raise ValueError(f"Unknown service: {service}")

        provider = self.providers[service]
        return await provider.authenticate(credentials)

    async def get_auth_url(self, service: str, redirect_uri: str) -> str:
        """Get OAuth URL for service authentication."""
        provider = self.providers[service]
        return provider.get_auth_url(redirect_uri)

    async def exchange_code_for_token(self, service: str, code: str, redirect_uri: str) -> dict:
        """Exchange OAuth code for access token."""
        provider = self.providers[service]
        return await provider.exchange_code_for_token(code, redirect_uri)

class FigmaAuthProvider:
    def __init__(self):
        self.client_id = os.getenv('FIGMA_CLIENT_ID')
        self.client_secret = os.getenv('FIGMA_CLIENT_SECRET')
        self.redirect_uri = os.getenv('FIGMA_REDIRECT_URI')

    async def authenticate(self, credentials: dict) -> bool:
        """Validate Figma API credentials."""
        try:
            client = FigmaClient(credentials['access_token'])
            # Test with a simple API call
            await client.get_user_info()
            return True
        except Exception:
            return False

    def get_auth_url(self, redirect_uri: str) -> str:
        """Generate Figma OAuth URL."""
        params = {
            'client_id': self.client_id,
            'redirect_uri': redirect_uri,
            'scope': 'file_read',
            'response_type': 'code',
            'state': self.generate_state()
        }
        return f"https://www.figma.com/oauth?{urlencode(params)}"

    async def exchange_code_for_token(self, code: str, redirect_uri: str) -> dict:
        """Exchange OAuth code for access token."""
        # Implementation for token exchange
        pass
```

### Error Handling and Resilience

**Robust Error Management**:
```python
class ErrorHandler:
    def __init__(self):
        self.retry_config = {
            'max_retries': 3,
            'backoff_factor': 2,
            'retryable_errors': ['timeout', 'rate_limit', 'temporary_failure']
        }

    async def handle_api_error(self, service: str, error: Exception, attempt: int = 1) -> dict:
        """Handle API errors with retry logic."""
        error_type = self.classify_error(error)

        if error_type in self.retry_config['retryable_errors'] and attempt <= self.retry_config['max_retries']:
            wait_time = self.retry_config['backoff_factor'] ** attempt
            await asyncio.sleep(wait_time)

            return {
                'should_retry': True,
                'wait_time': wait_time,
                'attempt': attempt + 1
            }

        return {
            'should_retry': False,
            'error_type': error_type,
            'message': str(error),
            'attempt': attempt
        }

    def classify_error(self, error: Exception) -> str:
        """Classify error type for retry determination."""
        if 'rate limit' in str(error).lower():
            return 'rate_limit'
        elif 'timeout' in str(error).lower():
            return 'timeout'
        elif 'temporary' in str(error).lower():
            return 'temporary_failure'
        else:
            return 'permanent_error'
```

---

## Works Well With

- **moai-domain-backend** - Backend integration for MCP servers
- **moai-domain-frontend** - Frontend tools and component generation
- **moai-foundation-core** - Core MCP principles and patterns
- **moai-quality-security** - Security validation for integrations
- **moai-system-universal** - Performance optimization strategies

---

## Integration Examples

### Design System Automation
```bash
# Extract design tokens and generate component library
mcp-tools extract_figma_components --file-id "abc123" --output ./src/components
mcp-tools sync_figma_tokens --file-id "abc123" --format "typescript"
```

### Knowledge Base Management
```bash
# Extract insights from Notion and create documentation
mcp-tools query_notion_database --database-id "xyz789" --query '{"filter": {"property": "Status", "select": {"equals": "Published"}}}'
mcp-tools analyze_with_ai --content-file "./notion-export.json" --analysis-type "key_points"
```

### AI-Powered Workflows
```bash
# Generate content with AI and store in Notion
mcp-tools generate_ai_content --prompt "Create technical documentation" --model "claude-3-5-sonnet"
mcp-tools create_notion_page --database-id "docs-db" --properties '{"Title": {"title": [{"text": {"content": "API Documentation"}}]}}'
```

---

## Technology Stack

**Core Frameworks**:
- **FastMCP**: Python MCP server development framework
- **Figma API**: Design system and component extraction
- **Notion API**: Database queries and content management
- **Nano-Banana API**: AI model integration and content generation

**Authentication**:
- OAuth 2.0 for user authentication
- API keys for service-to-service communication
- JWT tokens for session management

**Infrastructure**:
- Docker containerization
- FastAPI/ASGI server deployment
- Redis for caching and session management
- PostgreSQL for data persistence

---

**Status**: Production Ready
**Last Updated**: 2025-11-30
**Maintained by**: MoAI-ADK Integration Team