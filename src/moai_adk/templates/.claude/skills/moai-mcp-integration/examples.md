# MCP Integration Examples

## 10+ Real-World MCP Usage Patterns

---

### Example 1: Playwright Web Automation Server

**Use Case**: E2E testing and web scraping automation

```python
from fastmcp import FastMCP

server = FastMCP("web-automation-server")

@server.tool()
def automate_purchase_flow(url: str, username: str, password: str) -> dict:
    """Complete automated purchase flow."""
    browser = launch_browser()
    page = browser.new_page()

    # Navigate and login
    page.navigate(url)
    page.fill("#username", username)
    page.fill("#password", password)
    page.click("#login-btn")

    # Add to cart and checkout
    page.click(".product.featured")
    page.click("#add-to-cart")
    page.click("#checkout")

    screenshot = page.screenshot()
    browser.close()

    return {
        "status": "success",
        "screenshot": screenshot,
        "message": "Purchase flow completed"
    }

if __name__ == "__main__":
    server.run()
```

**Claude Usage**: "Automate the purchase flow on the test site with these credentials..."

---

### Example 2: GitHub Repository Management

**Use Case**: Automated PR creation and issue management

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "${GITHUB_CLIENT_ID}",
        "clientSecret": "${GITHUB_CLIENT_SECRET}",
        "scopes": ["repo", "issues"]
      }
    }
  }
}
```

**Claude Workflow**:
1. "Search for issues labeled 'bug' in my repository"
2. "Create a PR to fix issue #42 using branch 'fix/issue-42'"
3. "Merge the PR after CI passes"

---

### Example 3: Notion Database Integration

**Use Case**: Knowledge base management and documentation

```python
from fastmcp import FastMCP
from notion_client import Client

server = FastMCP("notion-automation-server")
notion = Client(auth=os.getenv("NOTION_API_KEY"))

@server.tool()
def create_meeting_notes(
    title: str,
    attendees: list[str],
    topics: list[str],
    database_id: str
) -> dict:
    """Create structured meeting notes in Notion."""
    page_data = {
        "parent": {"database_id": database_id},
        "properties": {
            "Title": {"title": [{"text": {"content": title}}]},
            "Date": {"date": {"start": datetime.now().isoformat()}},
            "Attendees": {"multi_select": [{"name": a} for a in attendees]},
        },
        "children": [
            {
                "object": "block",
                "type": "heading_2",
                "heading_2": {
                    "rich_text": [{"type": "text", "text": {"content": "Topics"}}]
                }
            },
            *[
                {
                    "object": "block",
                    "type": "bulleted_list_item",
                    "bulleted_list_item": {
                        "rich_text": [{"type": "text", "text": {"content": topic}}]
                    }
                }
                for topic in topics
            ]
        ]
    }

    result = notion.pages.create(**page_data)
    return {"page_id": result["id"], "url": result["url"]}

if __name__ == "__main__":
    server.run()
```

---

### Example 4: Firebase Firestore Integration

**Use Case**: Real-time database operations

```python
from fastmcp import FastMCP
import firebase_admin
from firebase_admin import firestore

server = FastMCP("firebase-mcp-server")
db = firestore.client()

@server.tool()
def create_user_profile(
    user_id: str,
    email: str,
    profile_data: dict
) -> dict:
    """Create user profile in Firestore."""
    try:
        db.collection("users").document(user_id).set({
            "email": email,
            "profile": profile_data,
            "created_at": firestore.SERVER_TIMESTAMP,
            "updated_at": firestore.SERVER_TIMESTAMP
        })

        return {
            "status": "success",
            "user_id": user_id,
            "message": "Profile created"
        }
    except Exception as e:
        raise ValueError(f"Failed to create profile: {str(e)}")

@server.resource("firebase://{collection}/{document_id}")
def get_document(collection: str, document_id: str) -> dict:
    """Retrieve document from Firestore."""
    doc = db.collection(collection).document(document_id).get()
    if doc.exists:
        return doc.to_dict()
    raise ValueError(f"Document not found: {collection}/{document_id}")

if __name__ == "__main__":
    server.run()
```

---

### Example 5: Multi-Server Orchestration

**Use Case**: Complex workflows using multiple MCP servers

```yaml
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {"clientId": "${GITHUB_CLIENT_ID}", "clientSecret": "${GITHUB_CLIENT_SECRET}"}
    },
    "slack": {
      "command": "npx",
      "args": ["-y", "@slack/mcp-server-slack"],
      "auth": {"type": "bearer", "token": "${SLACK_BOT_TOKEN}"}
    },
    "notion": {
      "command": "npx",
      "args": ["-y", "@notionhq/mcp-server-notion"],
      "auth": {"type": "bearer", "token": "${NOTION_API_KEY}"}
    }
  }
}
```

**Claude Workflow**:
1. Search GitHub issues labeled "ready for review"
2. Create Notion page with PR details
3. Send Slack notification to the team

---

### Example 6: Custom Authentication Server

**Use Case**: Internal tools with custom auth

```python
from fastmcp import FastMCP
from fastmcp.auth import OAuth2Provider

server = FastMCP("internal-tools-server")

oauth = OAuth2Provider(
    authorize_url="https://internal-auth.company.com/authorize",
    token_url="https://internal-auth.company.com/token",
    scopes=["read:internal", "write:internal"]
)

@server.auth(oauth)
@server.tool()
def query_sales_data(
    quarter: str,
    region: str,
    metric: str = "revenue"
) -> dict:
    """Query sales data (requires authentication)."""
    data = execute_query(f"""
        SELECT {metric}
        FROM sales
        WHERE quarter = '{quarter}' AND region = '{region}'
    """)

    return {"data": data, "query_time_ms": 45}

if __name__ == "__main__":
    server.run()
```

---

### Example 7: Figma Design System Server

**Use Case**: Design-to-code workflow automation

```python
from fastmcp import FastMCP
from figma_api import FigmaClient

server = FastMCP("figma-design-system-server")

@server.tool()
def extract_design_tokens(figma_file_key: str) -> dict:
    """Extract design tokens from Figma."""
    client = FigmaClient(os.getenv("FIGMA_API_KEY"))
    file_data = client.get_file(figma_file_key)

    tokens = {
        "colors": extract_colors(file_data),
        "typography": extract_typography(file_data),
        "spacing": extract_spacing(file_data),
        "components": extract_components(file_data)
    }

    return tokens

@server.tool()
def generate_tailwind_config(design_tokens: dict) -> str:
    """Generate Tailwind CSS config from tokens."""
    config = """export default {
  theme: {
    extend: {
      colors: {
    %s
      },
      fontSize: {
    %s
      }
    }
  }
}""" % (
        format_colors(design_tokens["colors"]),
        format_typography(design_tokens["typography"])
    )
    return config

if __name__ == "__main__":
    server.run()
```

---

### Example 8: Kubernetes Deployment with Health Checks

**Use Case**: Production-grade orchestration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
  namespace: default
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
    spec:
      containers:
      - name: mcp-server
        image: mcp-server:v2.1.0
        ports:
        - containerPort: 8000
          name: http
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: database-url
        - name: LOG_LEVEL
          value: "info"
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-server-service
spec:
  selector:
    app: mcp-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
  type: LoadBalancer
```

---

### Example 9: Error Handling and Validation

**Use Case**: Robust production server

```python
from fastmcp import FastMCP
from pydantic import BaseModel, Field, validator
from typing import Optional

server = FastMCP("robust-server")

class SearchQuery(BaseModel):
    query: str = Field(..., min_length=1, max_length=100)
    limit: int = Field(default=10, ge=1, le=100)
    offset: int = Field(default=0, ge=0)

    @validator('query')
    def query_not_empty(cls, v):
        if not v.strip():
            raise ValueError("Query cannot be only whitespace")
        return v.strip()

@server.tool()
def search(params: SearchQuery) -> dict:
    """Search with comprehensive validation."""
    try:
        results = execute_search(
            query=params.query,
            limit=params.limit,
            offset=params.offset
        )

        return {
            "status": "success",
            "count": len(results),
            "results": results,
            "total_available": get_total_count(params.query)
        }

    except DatabaseError as e:
        raise ValueError(f"Database error: {str(e)}")
    except Exception as e:
        raise ValueError(f"Search failed: {str(e)}")

if __name__ == "__main__":
    server.run()
```

---

### Example 10: Streaming Large Data

**Use Case**: Handling large datasets efficiently

```python
from fastmcp import FastMCP
from typing import AsyncGenerator

server = FastMCP("streaming-server")

@server.resource("analytics://{report_id}/stream")
async def stream_analytics_report(report_id: str) -> AsyncGenerator[str, None]:
    """Stream large analytics report in chunks."""
    query = f"SELECT * FROM reports WHERE id = '{report_id}'"

    async for row in execute_streaming_query(query):
        # Process and yield chunks instead of loading all in memory
        yield json.dumps(row) + "\n"

if __name__ == "__main__":
    server.run()
```

---

## Configuration Examples

### Simple Server (Single Tool)

```json
{
  "mcpServers": {
    "calculator": {
      "command": "python",
      "args": ["./calculator_server.py"],
      "capabilities": ["tools"]
    }
  }
}
```

### Enterprise Server (Multiple Auth Methods)

```json
{
  "mcpServers": {
    "enterprise": {
      "command": "python",
      "args": ["./enterprise_server.py"],
      "env": {
        "LOG_LEVEL": "info",
        "ENABLE_METRICS": "true"
      },
      "oauth": {
        "clientId": "${OAUTH_CLIENT_ID}",
        "clientSecret": "${OAUTH_CLIENT_SECRET}"
      },
      "apiKey": {
        "header": "X-API-Key",
        "value": "${API_KEY}"
      },
      "capabilities": ["tools", "resources", "prompts"]
    }
  }
}
```

### Distributed Deployment

```yaml
# 3 independent servers in Kubernetes
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-servers-config
data:
  servers.json: |
    {
      "mcpServers": {
        "github": { "command": "npx", "args": [...] },
        "slack": { "command": "npx", "args": [...] },
        "notion": { "command": "npx", "args": [...] }
      }
    }
```

---

## Common Integration Patterns

### Database Query Server
- Use pagination for large result sets
- Validate input parameters with Pydantic
- Return consistent data structures
- Implement connection pooling

### API Gateway Server
- Route requests to multiple backends
- Handle authentication per endpoint
- Implement circuit breakers for resilience
- Log all requests for auditing

### Document Processing Server
- Stream large files to avoid memory issues
- Validate file format before processing
- Return progress updates for long operations
- Support batch operations

---

**Last Updated**: 2025-11-22 | Total Examples: 10+
