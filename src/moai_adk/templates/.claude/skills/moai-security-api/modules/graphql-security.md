# GraphQL Security

## Overview

Security patterns for GraphQL API implementations.

## Query Depth Limiting

```python
from graphql import GraphQLSchema, GraphQLDepthLimitRule

schema = GraphQLSchema(
    query=Query,
    validation_rules=[
        GraphQLDepthLimitRule(max_depth=10)
    ]
)
```

## Rate Limiting

```python
from flask_limiter import Limiter

limiter = Limiter(
    key_func=get_remote_address,
    default_limits=["100 per hour"]
)

@limiter.limit("10 per minute")
@app.route("/graphql", methods=["POST"])
def graphql_handler():
    # GraphQL execution
    pass
```

## Input Validation

```python
from graphene import ObjectType, String, Field

class Query(ObjectType):
    user = Field(User, user_id=String(required=True))

    def resolve_user(self, info, user_id):
        # Validate input
        if not is_valid_uuid(user_id):
            raise ValidationError("Invalid user ID format")

        return get_user(user_id)
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
