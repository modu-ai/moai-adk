---
name: moai:git
description: 🎯 MoAI Git 통합 관리자
argument-hint: <서브커맨드> [옵션] - branch, commit, checkpoint, rollback, sync
allowed-tools: Bash(git:*), Bash(python3:*), Read, Write, Glob, Grep
model: sonnet
---

# MoAI Git 통합 관리 시스템

Constitution 5원칙을 준수하는 단일 Git 관리 인터페이스입니다.

## 작업 요청

Git 서브커맨드 실행: **$ARGUMENTS**

## 서브커맨드 처리

### 현재 상태 확인

현재 Git 상태를 확인합니다:

!`echo "=== Git 상태 ==="`
!`echo "📍 브랜치: $(git branch --show-current)"`
!`echo "📝 변경사항: $(git status --porcelain | wc -l)개 파일"`
!`echo "🎯 모드: $(python3 -c "import json; config=json.load(open('.moai/config.json')); print(config['project']['mode'])" 2>/dev/null || echo "unknown")"`

### Git 관리자 실행

통합 Git 관리자로 요청을 처리합니다:

!`python3 .moai/scripts/git_manager.py $ARGUMENTS`

## 지원되는 서브커맨드

### branch - 브랜치 관리

```bash
/moai:git branch                    # 현재 브랜치 및 목록 표시
/moai:git branch create feature/새기능  # 새 브랜치 생성
/moai:git branch switch main        # 브랜치 전환
/moai:git branch delete old-branch  # 브랜치 삭제
```

### commit - 스마트 커밋

```bash
/moai:git commit --auto             # 자동 메시지 생성 커밋
/moai:git commit "구체적인 메시지"    # 사용자 지정 메시지 커밋
```

### checkpoint - 체크포인트 관리 (개인 모드 전용)

```bash
/moai:git checkpoint                # 자동 체크포인트 생성
/moai:git checkpoint "실험 시작"     # 메시지와 함께 체크포인트
/moai:git checkpoint --list         # 체크포인트 목록 확인
```

### rollback - 안전한 롤백

```bash
/moai:git rollback --last           # 마지막 체크포인트로 롤백
/moai:git rollback checkpoint_20240122_143000  # 특정 체크포인트로 롤백
```

### sync - 원격 동기화

```bash
/moai:git sync --auto               # 모드별 자동 동기화
/moai:git sync                      # 상태 확인만
```

### help - 도움말

```bash
/moai:git help                      # 전체 도움말 표시
```

## Constitution 5원칙 준수

1. **Simplicity**: 5개 파일 → 1개 파일, 복잡한 로직은 Python 스크립트로 분리
2. **Architecture**: 명확한 서브커맨드 구조, 모듈화된 처리
3. **Testing**: 안전한 Git 명령어 사용, 오류 처리 내장
4. **Observability**: 모든 실행 과정이 투명하게 출력됨
5. **Versioning**: 표준 Git 워크플로우 준수, 체계적인 커밋 관리

## 특징

- **통일된 인터페이스**: 모든 Git 작업을 단일 명령어로 처리
- **모드별 최적화**: 개인/팀 모드에 따른 차별화된 워크플로우
- **자동 메시지 생성**: 변경사항 분석을 통한 지능적 커밋 메시지
- **안전한 실험**: 체크포인트 기반 롤백 시스템
- **Constitution 준수**: 모든 커밋에 표준 footer 자동 추가

이제 Git 작업이 **`/moai:git <서브커맨드>`** 하나로 통합되어 더욱 직관적이고 효율적입니다!
