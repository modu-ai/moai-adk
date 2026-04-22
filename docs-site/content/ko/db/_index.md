---
title: 데이터베이스 스키마 관리
description: 스키마, 마이그레이션, 시드 데이터를 자동으로 추적하고 관리하는 워크플로우
weight: 15
draft: false
---

MoAI-ADK의 데이터베이스 워크플로우는 프로젝트의 스키마 메타데이터를 중앙에서 관리합니다. `/moai db` 명령어를 통해 마이그레이션 파일을 스캔하고, 스키마 문서를 자동 생성하며, 드리프트를 감지할 수 있습니다.

## 주요 기능

- **대화형 초기화** — `/moai db init`으로 데이터베이스 엔진, ORM, 마이그레이션 도구를 선택하고 메타데이터 템플릿 자동 생성
- **자동 동기화** — PostToolUse 훅으로 마이그레이션 파일 변경 감지 후 자동 리프레시
- **드리프트 감지** — `/moai db verify`로 스키마 문서와 마이그레이션 파일 간 불일치 검사
- **16개 언어 지원** — Go, Python, TypeScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift

## 4개 서브명령어

```bash
/moai db init      # 대화형 인터뷰로 DB 메타데이터 초기화
/moai db refresh   # 마이그레이션 파일 리스캔 및 스키마 문서 재생성
/moai db verify    # 드리프트 확인 (읽기 전용)
/moai db list      # 모든 테이블을 Markdown 표로 표시
```

## 언제 사용하는가

- 새 프로젝트 시작 시 데이터베이스 메타데이터 설정
- 마이그레이션 파일 추가/편집 후 문서 자동 갱신
- 팀원에게 현재 스키마 상태 공유
- 스키마 문서와 실제 마이그레이션 상태 간 일관성 검증

## 다음 단계

- **[시작하기](./getting-started.md)** — `/moai db init` 실행 및 첫 마이그레이션
- **[스키마 동기화](./schema-sync.md)** — PostToolUse 훅과 자동 리프레시 메커니즘
- **[마이그레이션 패턴](./migration-patterns.md)** — 16개 언어의 기본 마이그레이션 경로
- **[프로젝트 DB 디렉토리](./project-db-directory.md)** — 7개 파일 템플릿 세트 소개

## 관련 문서

더 자세한 내용은 [/moai db 명령어 가이드](../../reference/moai-db.md)를 참조하세요.
