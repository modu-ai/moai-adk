# API Documentation Generation

## Overview

Generate comprehensive API documentation using established tools: FastAPI auto-docs, swagger-jsdoc, Redoc, and OpenAPI specification standards.

## FastAPI Automatic Documentation

FastAPI provides built-in OpenAPI documentation. Access at /docs (Swagger UI) and /redoc (ReDoc).

### Enhancing FastAPI Documentation

Add detailed descriptions to route handlers:
- Include docstrings that describe endpoint purpose
- Use response_model parameter for typed responses
- Add summary and description parameters to decorators
- Define examples in Pydantic model Config class

Organize endpoints with tags:
- Group related endpoints under meaningful tag names
- Add tag descriptions in app configuration
- Use consistent naming conventions

Enhance Pydantic models:
- Add Field descriptions with Field(description="...")
- Include examples with Field(example="...")
- Document validation constraints
- Use Config class for schema customization

### Exporting OpenAPI Specification

Access the generated OpenAPI JSON at /openapi.json endpoint.

For programmatic access, use app.openapi() method to get the specification dictionary.

Save to file by serializing the openapi() output to JSON format.

## Express/Node.js with swagger-jsdoc

### Setup

Install swagger-jsdoc and swagger-ui-express packages.

Create swagger configuration object:
- Define openapi version (3.0.0 or 3.1.0)
- Set info section with title, version, description
- Specify servers array with URLs
- Define component schemas for reuse
- Add security definitions if needed

### Documenting Endpoints

Add @openapi comments above route handlers:
- Document path with HTTP method
- Specify summary and description
- Define parameters (path, query, body)
- Document request body schema
- List all response codes with schemas
- Add tags for organization

### Serving Documentation

Use swagger-ui-express to serve interactive documentation at /api-docs endpoint.

Generate static OpenAPI JSON file for external tools.

## OpenAPI Specification Best Practices

Schema Organization:
- Define reusable schemas in components section
- Use $ref references to avoid duplication
- Create separate schemas for request and response when different
- Document nullable fields explicitly

Documentation Quality:
- Write clear, concise summaries (under 80 characters)
- Provide detailed descriptions for complex operations
- Include realistic examples for all schemas
- Document error responses with problem details

Security Documentation:
- Define security schemes in components
- Apply security requirements at operation or global level
- Document required scopes for OAuth flows
- Include authentication examples

## Rendering Tools

Swagger UI:
- Interactive documentation with try-it-out feature
- Customizable with CSS and configuration
- Supports OAuth authentication flows

Redoc:
- Three-panel layout for better readability
- Generates static HTML for hosting
- Better for large API specifications

Stoplight Elements:
- Modern React-based documentation
- Built-in mock server capabilities
- Supports OpenAPI 3.1 features

## Integration Points

CI/CD Integration:
- Generate OpenAPI spec during build
- Validate spec with spectral or openapi-validator
- Deploy documentation alongside API
- Version documentation with API releases

External Tools:
- Postman collection generation from OpenAPI
- SDK generation with openapi-generator
- API testing with tools like Dredd

---

Version: 2.0.0
Last Updated: 2025-12-30
