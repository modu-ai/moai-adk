# MoAI-ADK 명령어 시스템

## 🎯 핵심 명령어 개요

MoAI-ADK는 4단계 워크플로우 명령어, Git 전용 명령어 5종, 그리고 디버깅 명령어를 제공합니다.

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
| `/moai:git:commit`     | 개발 가이드 기반 커밋(RED/GREEN/REFACTOR 등) |
| `/moai:git:sync`       | 원격 저장소 동기화 및 충돌 보조 |

### 디버깅 명령어

| 명령어         | 기능 설명 |
|----------------|----------|
| `/moai:debug`  | 일반 오류 디버깅 + 개발 가이드 위반 검사 |

## 모델 사용 가이드

| 명령어            | 권장 모델 | 비고 |
|-------------------|----------|------|
| `/moai:0-project` | sonnet   | 프로젝트 문서 갱신 + CLAUDE 메모리 로드 |
| `/moai:debug`     | sonnet   | 오류 분석 + 개발 가이드 검사 |
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

## 디버깅 명령어 상세

### `/moai:debug`
```bash
# 일반 오류 디버깅
/moai:debug "TypeError: 'NoneType' object has no attribute 'name'"
/moai:debug "ImportError: No module named 'requests'"
/moai:debug "fatal: refusing to merge unrelated histories"
/moai:debug "PermissionError: [Errno 13] Permission denied"

# 개발 가이드 위반 검사
/moai:debug --constitution-check
```

#### 출력 예시: 일반 오류
```markdown
🐛 디버그 분석 결과
━━━━━━━━━━━━━━━━━━━
📍 오류 위치: src/auth/login.py:45
🔍 오류 유형: TypeError
📝 오류 내용: 'NoneType' object has no attribute 'name'

🔬 원인 분석:
- 직접 원인: user 객체가 None 상태
- 근본 원인: 인증 실패 시 예외 처리 누락
- 영향 범위: 로그인 플로우 전체

🛠️ 해결 방안:
1. 즉시 조치: None 체크 추가 (user and user.name)
2. 권장 수정: Optional chaining 활용 (user?.name)
3. 예방 대책: 인증 실패 예외 처리 강화

🎯 다음 단계:
→ code-builder 호출 권장
→ 예상 명령: /moai:2-build (코드 수정)
```

#### 출력 예시: 개발 가이드 검사
```markdown
🏛️ 개발 가이드 검사 결과
━━━━━━━━━━━━━━━━━━━━━
📊 전체 준수율: 85%

❌ 위반 사항:
1. Simplicity (파일 크기)
   - 현재: src/core/analyzer.py 420줄 (목표: ≤300줄)
   - 권장: 모듈 분리

2. Testing (커버리지)
   - 현재: 72% (목표: ≥85%)
   - 권장: 누락 테스트 추가

✅ 준수 사항:
- Observability: 구조화 로깅 ✓
- Versioning: 시맨틱 버전 ✓

🎯 개선 우선순위:
1. 테스트 커버리지 향상
2. 큰 파일 분리
3. 순환 의존성 해결

🔄 권장 다음 단계:
→ /moai:2-build (테스트 코드 추가)
→ /moai:1-spec (아키텍처 개선 명세)
```

## 참고

- `/moai:3-sync`는 TAG 인덱스를 갱신하고 `docs/status/sync-report.md`를 생성하며 `docs/sections/index.md` 갱신일을 반영합니다.
- GitHub PR 자동화는 Anthropic GitHub App 설치 및 시크릿 설정 후 권장됩니다.
- 상세 워크플로우와 제약 사항은 `docs/MOAI-ADK-GUIDE.md`를 참고하세요.
