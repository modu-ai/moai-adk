---

name: moai-core-env-security
description: Environment variable security, secret management, and credential protection patterns
allowed-tools: [Read, Bash]

---

# Environment Variable Security & Secret Management

## Quick Reference (30 seconds)

환경 변수 보안은 애플리케이션 구성을 안전하게 관리하고 민감한 자격증명(API 키, 데이터베이스 비밀번호, JWT 토큰)이
버전 관리, 로그, 메모리 덤프에서 노출되지 않도록 하는 필수 보안 패턴입니다.
`.env` 파일, 클라우드 비밀 관리자, IAM 역할을 조합하여 다층 방어를 구현합니다.

**핵심 원칙**:
- `.env` 파일을 버전 관리에 절대 커밋하지 않기
- 프로덕션 환경에서는 전용 비밀 관리자 사용
- 30-90일 주기로 비밀번호 로테이션
- 최소 권한 원칙(Least Privilege, IAM 정책)


## Implementation Guide

### 1. 개발 환경 설정

**`.env` 파일 패턴**:
```bash
# .env (개발용 - 버전 관리 제외)
DATABASE_URL=postgresql://user:password@localhost:5432/devdb
API_KEY=sk_dev_abc123xyz789
JWT_SECRET=dev-secret-key-not-for-production
DEBUG=true
```

**`.env.example` (템플릿 - 버전 관리 포함)**:
```bash
# .env.example (실제 값 없음, 필드만 정의)
DATABASE_URL=
API_KEY=
JWT_SECRET=
DEBUG=
```

**`.gitignore` 설정**:
```bash
# 비밀 파일 보호
.env
.env.local
.env.*.local
.env.development
.env.test.local
.env.production

# 기타 민감 파일
*.pem
*.key
secrets.yaml
credentials.json
.aws/
.gcloud/
```

### 2. Python 환경 변수 로딩

**python-dotenv 사용**:
```python
from dotenv import load_dotenv
import os

# .env 파일 로드
load_dotenv()

# 환경 변수 접근
DATABASE_URL = os.getenv('DATABASE_URL')
API_KEY = os.getenv('API_KEY', default='fallback-value')
```

**필수 변수 검증**:
```python
import os
from typing import List

def validate_env_variables(required: List[str]) -> None:
    """필수 환경 변수 검증"""
    missing = [var for var in required if not os.getenv(var)]
    if missing:
        raise ValueError(f"Missing required environment variables: {missing}")

# 사용
validate_env_variables([
    'DATABASE_URL',
    'API_KEY',
    'JWT_SECRET'
])
```

**Pydantic를 사용한 타입-안전 검증**:
```python
from pydantic import BaseSettings, Field

class Settings(BaseSettings):
    database_url: str = Field(..., description="PostgreSQL connection string")
    api_key: str = Field(..., min_length=20)
    jwt_secret: str = Field(..., min_length=32)
    debug: bool = Field(default=False)

    class Config:
        env_file = ".env"

settings = Settings()  # 자동 검증
```

### 3. 클라우드 비밀 관리 (프로덕션)

**AWS Secrets Manager**:
```python
import boto3
import json

def get_aws_secret(secret_name: str, region: str = 'us-east-1') -> dict:
    """AWS에서 비밀 조회"""
    client = boto3.client('secretsmanager', region_name=region)
    try:
        response = client.get_secret_value(SecretId=secret_name)
        return json.loads(response['SecretString'])
    except Exception as e:
        raise RuntimeError(f"Failed to retrieve secret: {e}")

# 사용
db_creds = get_aws_secret('prod/database')
DATABASE_URL = db_creds['url']
```

**Google Cloud Secret Manager**:
```python
from google.cloud import secretmanager

def get_gcp_secret(project_id: str, secret_id: str, version: str = 'latest') -> str:
    """Google Cloud에서 비밀 조회"""
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{project_id}/secrets/{secret_id}/versions/{version}"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode("UTF-8")

# 사용
api_key = get_gcp_secret('my-project', 'api-key')
```

**HashiCorp Vault**:
```python
import hvac

def get_vault_secret(path: str, vault_addr: str, token: str) -> dict:
    """Vault에서 비밀 조회"""
    client = hvac.Client(url=vault_addr, token=token)
    response = client.secrets.kv.read_secret_version(path)
    return response['data']['data']

# 사용
db_secret = get_vault_secret('secret/db/prod', 'https://vault.example.com', token)
```

### 4. 비밀 로테이션 (Secret Rotation)

**자동 로테이션 패턴**:
```python
import os
from datetime import datetime, timedelta

class SecretRotationManager:
    """비밀 로테이션 관리"""

    def __init__(self, rotation_days: int = 90):
        self.rotation_days = rotation_days

    def should_rotate(self, secret_name: str) -> bool:
        """로테이션 필요 여부 확인"""
        last_rotated = self._get_last_rotation_date(secret_name)
        if not last_rotated:
            return True

        days_since = (datetime.now() - last_rotated).days
        return days_since >= self.rotation_days

    def rotate_secret(self, secret_name: str) -> str:
        """비밀 로테이션 실행"""
        new_secret = self._generate_new_secret()
        self._update_secret_in_manager(secret_name, new_secret)
        self._record_rotation_timestamp(secret_name)
        return new_secret

    def _generate_new_secret(self) -> str:
        import secrets
        return secrets.token_urlsafe(32)

    def _update_secret_in_manager(self, name: str, value: str):
        # AWS Secrets Manager, Vault 등에 업데이트
        pass

    def _record_rotation_timestamp(self, name: str):
        # 로테이션 시간 기록
        pass
```

### 5. 비밀 감지 및 방지

**Git Pre-commit Hook** (git-secrets):
```bash
#!/bin/bash
# .git/hooks/pre-commit

# git-secrets 설치
git secrets --install

# 패턴 정의
git secrets --register-aws

# 모든 커밋 스캔
git secrets --scan
```

**Python Pre-commit 검증**:
```python
import re

SECRET_PATTERNS = [
    r'(?i)(password|passwd|pwd)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'(?i)(api_?key|api_?token)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'(?i)(secret|token)\s*=\s*[\'"]?([^\s\'\"]+)',
    r'Bearer\s+[A-Za-z0-9\-\._~\+\/]+=*',
]

def detect_secrets_in_code(code: str) -> list:
    """코드에서 비밀 패턴 감지"""
    found = []
    for pattern in SECRET_PATTERNS:
        matches = re.finditer(pattern, code)
        found.extend([m.group() for m in matches])
    return found
```


## Best Practices

### ✅ DO
- `.env.example` 템플릿 사용 (실제 값 없음)
- 환경별 비밀 분리 (dev, staging, prod)
- 환경 변수 로드 시 기본값 설정 금지
- 필수 환경 변수 시작 시 검증
- 비밀번호/토큰 접근 감시 로깅
- IAM 역할 사용 (장기 자격증명 대신)
- 비밀 접근 권한 최소화

### ❌ DON'T
- 환경 변수를 로그에 출력
- `.env` 파일을 이메일/Slack으로 공유
- 프로덕션 비밀을 개발 환경에서 사용
- 암호화되지 않은 전송 (HTTP 대신 HTTPS)
- 하드코드된 비밀번호/키
- 비밀 만료 정책 무시
- 비밀 접근 감사 로깅 생략


## Works Well With

- `moai-domain-security` (전체 보안 아키텍처)
- `moai-lang-python` (Python 환경 변수 패턴)
- `moai-domain-devops` (배포 시 비밀 관리)
- `moai-baas-foundation` (BaaS 비밀 통합)


**Version**: 2.0.0 | **Last Updated**: 2025-11-21 | **Lines**: 180
