---
title: "보안 예제"
description: "애플리케이션 보안 강화 기법"
---

# 보안 예제

안전한 웹 애플리케이션을 만들기 위한 보안 예제입니다.

## 📚 예제 목록

### [입력 검증](/ko/examples/security/input-validation)
**난이도**: 초급 | **태그**: `pydantic`, `validation`, `security`

Pydantic을 사용한 강력한 입력 데이터 검증

### [SQL 인젝션 방지](/ko/examples/security/sql-injection-prevention)
**난이도**: 초급 | **태그**: `sql`, `security`, `sqlalchemy`

SQL 인젝션 공격을 방지하는 안전한 쿼리 작성

### [XSS 방어](/ko/examples/security/xss-protection)
**난이도**: 중급 | **태그**: `xss`, `security`, `sanitization`

크로스 사이트 스크립팅(XSS) 공격 방어

### [속도 제한](/ko/examples/security/rate-limiting)
**난이도**: 중급 | **태그**: `rate-limit`, `security`, `redis`

API 남용 및 DDoS 공격 방지를 위한 속도 제한

---

## 🎯 보안 체크리스트

### 필수 사항
- ✅ **HTTPS 사용**: 모든 통신 암호화
- ✅ **입력 검증**: 모든 사용자 입력 검증
- ✅ **SQL 인젝션 방지**: ORM 사용, Raw query 지양
- ✅ **XSS 방어**: 출력 시 이스케이핑
- ✅ **CSRF 토큰**: 상태 변경 작업에 토큰 사용
- ✅ **Rate Limiting**: API 속도 제한
- ✅ **비밀번호 해싱**: bcrypt, argon2 사용
- ✅ **SECRET_KEY 관리**: 환경 변수로 관리

### 권장 사항
- ⚡ **CORS 설정**: 적절한 출처 제한
- ⚡ **보안 헤더**: CSP, X-Frame-Options 등
- ⚡ **감사 로깅**: 보안 이벤트 기록
- ⚡ **정기 업데이트**: 의존성 보안 패치

## 🛡️ OWASP Top 10

MoAI-ADK 예제는 OWASP Top 10 보안 위협에 대응합니다:

1. **Broken Access Control** → RBAC 예제
2. **Cryptographic Failures** → JWT, 비밀번호 해싱
3. **Injection** → SQL 인젝션 방지
4. **Insecure Design** → 보안 설계 패턴
5. **Security Misconfiguration** → 안전한 설정
6. **Vulnerable Components** → 의존성 관리
7. **Identification & Auth** → JWT 인증
8. **Software & Data Integrity** → 입력 검증
9. **Logging & Monitoring** → 감사 로그
10. **SSRF** → 외부 요청 검증

## 📖 관련 문서

- [JWT 인증](/ko/examples/authentication/jwt-basic)
- [에러 처리](/ko/examples/rest-api/error-handling)
- [OWASP Cheat Sheet](https://cheatsheetseries.owasp.org/)

---

**시작하기**: [입력 검증](/ko/examples/security/input-validation) 예제부터 시작하세요!
