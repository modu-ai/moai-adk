---
title: /moai db
weight: 50
draft: false
---

프로젝트 데이터베이스 메타데이터를 관리하는 워크플로우 명령어입니다.

{{< callout type="info" >}}
**슬래시 커맨드**: Claude Code에서 `/moai:db`을 입력하면 이 명령어를 바로 실행할 수 있습니다.
{{< /callout >}}

## 개요

`/moai db`는 데이터베이스 스키마, 마이그레이션, 시드를 관리하는 4개 하위 명령어를
제공합니다. 대화형 인터뷰로 DB 엔진, ORM, 멀티테넌트 설정을 구성하고
`.moai/project/db/` 아티팩트를 생성합니다.

## 하위 명령어

| 명령어 | 설명 | 읽기 전용 |
|--------|------|-----------|
| `init` | 대화형 초기 설정 | 아니오 |
| `refresh` | 마이그레이션 재스캔 + 스키마 문서 재생성 | 아니오 |
| `verify` | 드리프트 감지 (스키마 vs 마이그레이션) | 예 |
| `list` | 테이블 목록 출력 | 예 |

```bash
/moai db init       # 대화형 설정 마법사
/moai db refresh    # 마이그레이션 재스캔
/moai db verify     # 드리프트 검사
/moai db list       # 테이블 목록
/moai db            # 하위 명령어 선택
```

## init — 대화형 초기 설정

### Phase 1: 사전 확인

기존 `.moai/project/db/` 확인 후 인터뷰를 진행합니다.

### Phase 2: 인터뷰

AskUserQuestion으로 다음 항목을 설정합니다:

- **DB 엔진**: PostgreSQL, MySQL, SQLite, MongoDB 등
- **ORM**: Prisma, TypeORM, SQLAlchemy, GORM 등
- **멀티테넌트**: 스키마 분리 / 행 수준 분리 / 단일 테넌트
- **마이그레이션 도구**: 선택한 ORM의 기본 도구 또는 커스텀

### Phase 3: 템플릿 렌더링

`.moai/project/db/` 디렉터리에 다음 파일을 생성합니다:

```
.moai/project/db/
├── schema.md          # 스키마 문서 (Markdown)
├── migrations.md      # 마이그레이션 추적
└── seeds.md           # 시드 데이터 가이드
```

## refresh — 마이그레이션 재스캔

마이그레이션 파일을 다시 스캔하고 스키마 문서를 재생성합니다.

- 새 마이그레이션 파일 감지
- `schema.md` 자동 업데이트
- 변경사항 요약 출력

## verify — 드리프트 검사

`schema.md`의 테이블 목록과 실제 마이그레이션 파일을 비교합니다.

- **일치**: 종료 코드 0
- **드리프트 감지**: 종료 코드 1 + 상세 보고서

CI 파이프라인에서 사용하여 스키마 동기화를 검증할 수 있습니다.

```yaml
# GitHub Actions 예시
- name: DB Schema Drift Check
  run: claude -p "/moai db verify"
```

## list — 테이블 목록

`schema.md`의 모든 테이블을 Markdown 정렬 표로 출력합니다.

## --mode 호환성

이 명령어는 Multi-Mode Router의 `--mode` 축에 참여하지 않습니다. `--mode` 값을
전달하면 조용히 무시됩니다. 단, `--mode pipeline`은 `MODE_PIPELINE_ONLY_UTILITY`
오류를 발생시킵니다.

## 관련 문서

- [데이터베이스 시작하기](/ko/db/getting-started) — DB 관리 상세 가이드
- [스키마 동기화](/ko/db/schema-sync) — 스키마 동기화 패턴
- [마이그레이션 패턴](/ko/db/migration-patterns) — 마이그레이션 전략
- [/moai plan](/ko/workflow-commands/moai-plan) — SPEC 문서 생성
