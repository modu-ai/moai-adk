# MoAI-ADK 명령어 시스템

## 🎯 핵심 명령어 개요

MoAI-ADK는 4단계 워크플로우 명령어와 Git 전용 명령어 5종을 제공합니다.

### 워크플로우 명령어 (0→3)

| 순서 | 명령어            | 담당 에이전트 | 기능 |
|-----:|-------------------|---------------|------|
| 0    | `/moai:0-project` | cc-manager    | 프로젝트 문서 초기화/갱신 + 메모리 반영 |
| 1    | `/moai:1-spec`    | spec-builder  | 프로젝트 문서 기반 SPEC auto 제안/생성 |
| 2    | `/moai:2-build`   | code-builder  | TDD 구현 (개인: 체크포인트 / 팀: 7단계 커밋) |
| 3    | `/moai:3-sync`    | doc-syncer    | 문서/PR 동기화 + TAG 인덱스 갱신 |

### Git 전용 명령어 (5종)

| 명령어                 | 기능 설명 |
|------------------------|----------|
| `/moai:git:checkpoint` | 자동/수동 체크포인트 생성(개인) |
| `/moai:git:rollback`   | 체크포인트 기반 안전 롤백 |
| `/moai:git:branch`     | 모드별 브랜치 전략(개인/팀) |
| `/moai:git:commit`     | Constitution 기반 커밋(RED/GREEN/REFACTOR 등) |
| `/moai:git:sync`       | 원격 저장소 동기화 및 충돌 보조 |

## 모델 사용 가이드

| 명령어            | 권장 모델 | 비고 |
|-------------------|----------|------|
| `/moai:0-project` | sonnet   | 프로젝트 문서 갱신 + CLAUDE 메모리 로드 |
| `/moai:1-spec`    | sonnet   | auto 제안 후 생성(개인: 로컬, 팀: GitHub Issue) |
| `/moai:2-build`   | sonnet   | TDD (개인: 체크포인트, 팀: 7단계 커밋) |
| `/moai:3-sync`    | haiku    | 문서/PR 동기화 + TAG 인덱스 갱신 |

## 사용 예시

### 개인 모드

```bash
/moai:0-project
/moai:1-spec                    # 로컬 SPEC 생성
/moai:git:checkpoint "작업 시작"
/moai:2-build                   # TDD + 자동 체크포인트
/moai:3-sync                    # 문서/상태 보고
```

### 팀 모드

```bash
/moai:0-project update
/moai:1-spec                    # GitHub Issue 생성
/moai:git:branch --team SPEC-001
/moai:2-build SPEC-001          # 7단계 커밋 패턴
/moai:git:sync --pull
/moai:3-sync                    # PR Ready 전환(옵션)
```

## Git 명령어 상세

### `/moai:git:checkpoint`
```bash
/moai:git:checkpoint                   # 자동 체크포인트
/moai:git:checkpoint "메시지"         # 수동 메시지 포함
/moai:git:checkpoint --list            # 체크포인트 목록
/moai:git:checkpoint --status          # 상태 확인
```

### `/moai:git:rollback`
```bash
/moai:git:rollback --list
/moai:git:rollback --checkpoint checkpoint_YYYYMMDD_HHMMSS
/moai:git:rollback --time "30분전" | --last | --safe
```

### `/moai:git:branch`
```bash
/moai:git:branch --status
/moai:git:branch --personal "새-기능"     # → feature/새-기능
/moai:git:branch --team SPEC-001          # → feature/SPEC-001-설명
/moai:git:branch --cleanup
```

### `/moai:git:commit`
```bash
/moai:git:commit --auto
/moai:git:commit --red|--green|--refactor "메시지"
/moai:git:commit --spec SPEC-001 --message "설명"
```

### `/moai:git:sync`
```bash
/moai:git:sync --auto | --pull | --push | --resolve
```

## 참고

- `/moai:3-sync`는 TAG 인덱스를 갱신하고 `docs/status/sync-report.md`를 생성하며 `docs/sections/index.md` 갱신일을 반영합니다.
- GitHub PR 자동화는 Anthropic GitHub App 설치 및 시크릿 설정 후 권장됩니다.
- 상세 워크플로우와 제약 사항은 `docs/MOAI-ADK-0.2.2-GUIDE.md`를 참고하세요.
