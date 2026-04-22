---
title: 시작하기
description: /moai db init으로 프로젝트 데이터베이스 메타데이터 초기화
weight: 10
draft: false
---

## 사전 요구사항

데이터베이스 워크플로우를 시작하기 전에 다음이 필요합니다:

1. `/moai project` 명령어로 생성된 `.moai/project/product.md`와 `.moai/project/tech.md` 파일
2. 지원되는 데이터베이스 엔진 (PostgreSQL, MySQL, SQLite, MongoDB 등)
3. ORM 또는 쿼리 빌더 (GORM, sqlc, Prisma, SQLAlchemy, ActiveRecord 등)
4. 마이그레이션 도구 (golang-migrate, Flyway, Liquibase, Alembic 등)

## 단계별 초기화 가이드

### 1단계: 프로젝트 메타데이터 확인

먼저 `/moai project`를 실행하여 필수 파일이 생성되었는지 확인합니다:

```bash
ls -la .moai/project/
# 다음 파일이 존재해야 합니다:
# - product.md
# - tech.md
# - structure.md
```

이 파일들이 없으면 먼저 `/moai project`를 실행하세요.

### 2단계: 데이터베이스 메타데이터 초기화

이제 `/moai db init` 명령어를 실행합니다:

```bash
/moai db init
```

### 3단계: 인터뷰 질문 응답

MoAI는 다음 4가지 항목에 대해 대화형으로 질문합니다:

1. **데이터베이스 엔진** — 사용 중인 데이터베이스 (PostgreSQL, MySQL, SQLite, MongoDB 등)
2. **ORM/쿼리 빌더** — 데이터 접근 계층 도구
3. **멀티테넌트 전략** — 싱글 스키마, 스키마당 테넌트, DB당 테넌트, 또는 없음
4. **마이그레이션 도구** — 스키마 변경 관리 도구

각 질문에 대해 적절한 옵션을 선택합니다.

### 4단계: 생성된 파일 검토

초기화 후 `.moai/project/db/` 디렉토리에 다음 파일들이 생성됩니다:

```
.moai/project/db/
├── README.md              # DB 섹션 개요
├── schema.md              # 자동 생성되는 테이블 레지스트리
├── erd.mmd                # 엔티티 관계 다이어그램
├── migrations.md          # 마이그레이션 파일 인덱스
├── rls-policies.md        # Row-level security 규칙 (Supabase/Postgres)
├── queries.md             # 일반적인 쿼리 라이브러리
└── seed-data.md           # 시드 데이터 패턴
```

각 파일의 역할:

- `schema.md` — 모든 테이블, 컬럼, 데이터타입, 제약조건을 자동으로 문서화
- `erd.mmd` — Mermaid 문법으로 테이블 관계를 시각화
- `migrations.md` — 적용된 마이그레이션 파일 타임라인
- `queries.md` — AI 에이전트가 참고하는 일반 쿼리 예제 모음

### 5단계: 첫 번째 마이그레이션 작성 및 동기화

새로운 마이그레이션 파일을 프로젝트에 추가합니다. 예를 들어 Go/golang-migrate인 경우:

```bash
# db/migrations/ 디렉토리에 마이그레이션 파일 생성
touch db/migrations/001_create_users_table.sql
```

마이그레이션 파일을 작성한 후, 다음 명령어로 스키마 문서를 갱신합니다:

```bash
/moai db refresh
```

이 명령어는:
- 모든 마이그레이션 파일을 스캔
- schema.md에 새로운 테이블 정보 추가
- erd.mmd 다이어그램 업데이트
- migrations.md 타임라인 갱신

### 6단계: 드리프트 검증 (선택사항)

드리프트가 있는지 확인하려면:

```bash
/moai db verify
```

결과:

- `스키마 문서가 동기화되어 있습니다` — 마이그레이션과 문서가 일치
- 드리프트 보고서 출력 — 차이점 상세 표시 (exit code: 1)

## 문제 해결

### "Missing prerequisite files" 오류

`.moai/project/product.md`와 `.moai/project/tech.md`가 없는 경우:

```bash
/moai project
```

위 명령어를 먼저 실행하여 프로젝트 메타데이터를 생성하세요.

### 마이그레이션 파일이 인식되지 않음

프로젝트의 언어와 마이그레이션 도구가 올바르게 감지되었는지 확인합니다:

```bash
cat .moai/config/sections/language.yaml
```

`language` 필드를 확인하고, 필요하면 `.moai/config/sections/db.yaml`에서 `migration_patterns`을 수동으로 지정할 수 있습니다.

### 자동 동기화가 작동하지 않음

PostToolUse 훅이 올바르게 등록되었는지 확인합니다:

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

훅이 없으면 `/moai db init`을 다시 실행하거나 `.claude/settings.json`에 수동으로 등록하세요.
