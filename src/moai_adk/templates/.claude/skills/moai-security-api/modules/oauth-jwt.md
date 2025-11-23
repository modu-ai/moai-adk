# OAuth & JWT Security

## Overview

OAuth 2.0 and JWT implementation security patterns.

## OAuth 2.0 Patterns

### Authorization Code Flow

```python
from authlib.integrations.flask_client import OAuth

oauth = OAuth()
oauth.register(
    'google',
    client_id='CLIENT_ID',
    client_secret='CLIENT_SECRET',
    authorize_url='https://accounts.google.com/o/oauth2/auth',
    access_token_url='https://accounts.google.com/o/oauth2/token'
)
```

## JWT Security

### Token Generation

```python
import jwt
from datetime import datetime, timedelta

def generate_token(user_id: str) -> str:
    payload = {
        'sub': user_id,
        'exp': datetime.utcnow() + timedelta(hours=1),
        'iat': datetime.utcnow()
    }
    return jwt.encode(payload, SECRET_KEY, algorithm='HS256')
```

### Token Validation

```python
def validate_token(token: str) -> dict:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=['HS256'])
        return payload
    except jwt.ExpiredSignatureError:
        raise AuthenticationError("Token expired")
    except jwt.InvalidTokenError:
        raise AuthenticationError("Invalid token")
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
