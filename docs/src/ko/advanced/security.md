# 보안 고급 가이드

MoAI-ADK 프로젝트의 보안을 강화하는 종합 가이드입니다.

## 🎯 보안 원칙

1. **기본값은 안전하게**: 안전한 설정을 기본값으로
2. **입력 검증**: 모든 외부 입력 검증
3. **최소 권한**: 필요한 권한만 부여
4. **심층 방어**: 여러 계층의 방어 수단

## 🛡️ OWASP Top 10 대응

### 1. SQL Injection

```python
# ❌ 위험
user = db.execute(f"SELECT * FROM users WHERE email = '{email}'")

# ✅ 안전
user = db.execute(
    "SELECT * FROM users WHERE email = ?",
    [email]
)
```

### 2. 인증 & 세션 관리

```python
# ✅ JWT 토큰
token = jwt.encode(
    {"user_id": user.id, "exp": datetime.utcnow() + timedelta(hours=1)},
    SECRET_KEY,
    algorithm="HS256"
)

# ✅ 비밀번호 해싱
hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())
```

### 3. Cross-Site Scripting (XSS)

```python
# ✅ 자동 이스케이프 (Jinja2)
<p>{{ user_input }}</p>  <!-- 자동으로 이스케이프됨 -->

# ✅ 수동 이스케이프
from markupsafe import escape
safe_html = escape(user_input)
```

### 4. Cross-Site Request Forgery (CSRF)

```python
# ✅ CSRF 토큰
<form method="post">
    <input type="hidden" name="csrf_token" value="{{ csrf_token() }}">
</form>
```

## 🔐 보안 체크리스트

### 개발 단계

- [ ] 입력 검증 구현
- [ ] 민감한 데이터 로깅 금지
- [ ] 에러 메시지 최소화 (정보 유출 방지)
- [ ] 인증/인가 구현
- [ ] 비밀번호 해싱 (bcrypt/scrypt)

### 배포 단계

- [ ] HTTPS 활성화
- [ ] CORS 정책 설정
- [ ] 보안 헤더 추가
- [ ] 의존성 보안 검사
- [ ] 데이터베이스 백업 암호화

### 운영 단계

- [ ] 접근 로그 모니터링
- [ ] 의존성 업데이트 확인
- [ ] 침입 탐지 시스템
- [ ] 정기 보안 감사
- [ ] 장애 대응 계획

## 🔑 암호화

### 데이터 암호화

```python
from cryptography.fernet import Fernet

# 대칭 암호화
cipher = Fernet(key)
encrypted = cipher.encrypt(b"sensitive data")
decrypted = cipher.decrypt(encrypted)
```

### 통신 암호화

```bash
# HTTPS 인증서 생성
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem

# Nginx 설정
server {
    listen 443 ssl;
    ssl_certificate cert.pem;
    ssl_certificate_key key.pem;
}
```

## 🚨 보안 취약점 스캔

### 자동 스캔 도구

```bash
# Python 의존성 보안 검사
safety check

# Snyk 스캔
snyk test

# Bandit (정적 분석)
bandit -r src/
```

### 스캔 결과 처리

```
vulnerability_db: 1.0.2 → 1.0.3 (update)
unpatched_dep: 2.0.0 (보안 업데이트 필요)

CRITICAL: SQL injection 가능성 (src/api.py:45)
→ 즉시 수정 필요
```

## 📋 보안 정책 수립

### 암호 정책

```
- 최소 12자
- 대문자, 소문자, 숫자, 특수문자 포함
- 90일마다 변경
- 이전 5개 암호 재사용 금지
```

### 접근 제어

```
RBAC (Role-Based Access Control):
- admin: 모든 권한
- manager: 데이터 읽기/쓰기
- user: 자신의 데이터만
```

______________________________________________________________________

**다음**: [확장 및 커스터마이제이션](extensions.md) 또는 [성능 최적화](performance.md)
