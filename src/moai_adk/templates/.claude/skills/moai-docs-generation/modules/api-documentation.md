# API Documentation Generation

## Overview
Automated API documentation generation from code with OpenAPI/Swagger support, example generation, and multi-format output.

## Quick Implementation

```python
from typing import Dict, List, Optional
import inspect
from fastapi import FastAPI

class APIDocGenerator:
 def __init__(self, app: FastAPI):
 self.app = app
 self.openapi_schema = app.openapi()

 def generate_openapi_spec(self) -> dict:
 """Generate complete OpenAPI specification."""
 return {
 "openapi": "3.1.0",
 "info": {
 "title": self.openapi_schema["info"]["title"],
 "version": self.openapi_schema["info"]["version"],
 "description": self.openapi_schema["info"].get("description", ""),
 "contact": {
 "name": "API Support",
 "email": "support@example.com"
 }
 },
 "paths": self.enrich_paths_with_examples(),
 "components": self.generate_components()
 }

 def enrich_paths_with_examples(self) -> dict:
 """Add examples and detailed descriptions to API paths."""
 enriched_paths = {}

 for path, path_item in self.openapi_schema["paths"].items():
 enriched_paths[path] = {}

 for method, operation in path_item.items():
 if method.upper() in ["GET", "POST", "PUT", "DELETE"]:
 enriched_paths[path][method] = {
 operation,
 "responses": self.enrich_responses(operation.get("responses", {})),
 "examples": self.generate_examples(path, method, operation)
 }

 return enriched_paths

 def generate_examples(self, path: str, method: str, operation: dict) -> dict:
 """Generate request/response examples."""
 examples = {}

 if method.upper() == "POST":
 examples["application/json"] = {
 "summary": "Example request",
 "value": self.generate_request_example(operation)
 }

 return examples

 def generate_markdown_docs(self) -> str:
 """Generate comprehensive Markdown documentation."""
 template = """
# {title}

{description}

## Base URL
```
{base_url}
```

## Authentication
{authentication}

## Endpoints

{endpoints}

## Data Models

{models}
 """

 endpoints = self.generate_endpoint_markdown()
 models = self.generate_model_markdown()

 return template.format(
 title=self.openapi_schema["info"]["title"],
 description=self.openapi_schema["info"].get("description", ""),
 base_url="https://api.example.com/v1",
 authentication=self.generate_auth_docs(),
 endpoints=endpoints,
 models=models
 )

# Usage example
def generate_api_docs():
 app = FastAPI(title="User Management API", version="1.0.0")

 doc_generator = APIDocGenerator(app)

 # Generate different formats
 openapi_spec = doc_generator.generate_openapi_spec()
 markdown_docs = doc_generator.generate_markdown_docs()
 html_docs = doc_generator.generate_html_docs()

 return {
 "openapi": openapi_spec,
 "markdown": markdown_docs,
 "html": html_docs
 }
```

## Key Features
- OpenAPI 3.1.0 specification generation
- Automatic example generation
- Multi-format output (OpenAPI, Markdown, HTML)
- Request/response enrichment
- Authentication documentation
- Interactive API docs generation

## Integration Points
- FastAPI applications
- Express.js applications
- Django REST Framework
- Spring Boot applications
