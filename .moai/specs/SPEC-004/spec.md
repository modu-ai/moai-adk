# SPEC-004: Claude Hooks 최적화 및 불필요한 것들 제거

## @SPEC:HOOKS-CLEANUP-001

### Environment (환경 및 가정사항)

**현재 환경**

- `.claude/hooks/moai/` 디렉토리에 10개 hook 파일 존재
- 총 코드량: 3,853줄
- 주요 문제: session_start_notice.py (2,133줄, 전체의 55%)
- Claude Code hooks 시스템 기반 동작

**가정사항**

- Claude Code hooks API는 변경되지 않음
- 기존 hook 실행 순서와 이벤트 트리거 유지
- Python 3.11+ 환경에서 동작
- TRUST 5원칙 준수 (특히 R-Readable: 파일당 300줄 이하)

### Assumptions (전제 조건)

**기능 분류 기준**

1. **핵심 기능**: 세션 운영에 필수적인 기능
2. **보조 기능**: 편의성 향상 기능
3. **중복 기능**: 다른 모듈과 역할 겹침
4. **불필요 기능**: 사용 빈도 낮거나 외부 도구로 대체 가능

**성능 목표**

- 세션 시작 시간 50% 단축
- 전체 코드량 74% 감소 (3,853줄 → ~1,000줄)
- 메모리 사용량 최적화

### Requirements (기능 요구사항)

#### @REQ:HOOKS-CORE-001 핵심 Hook 유지

- **session_start_notice.py**: 세션 상태 알림 (2,133줄 → ~200줄)
- **pre_write_guard.py**: 파일 쓰기 전 보안 검사 (131줄 → 유지)
- **policy_block.py**: 위험 명령 차단 (95줄 → 유지)
- **language_detector.py**: 프로젝트 언어 감지 (108줄 → 유지)
- **steering_guard.py**: 프롬프트 검사 (79줄 → 유지)
- **run_tests_and_report.py**: 테스트 실행 (91줄 → 유지)

#### @REQ:HOOKS-CLEANUP-001 불필요한 Hook 제거

- **tag_validator.py**: TAG 검증 (430줄) → 제거 (사용 빈도 낮음)
- **check_style.py**: 코드 스타일 검사 (241줄) → 제거 (CI/CD로 이전)

#### @REQ:HOOKS-MERGE-001 중복 기능 통합

- **auto_checkpoint.py** + **file_watcher.py** → **checkpoint_manager.py** (통합)
  - auto_checkpoint.py (222줄) + file_watcher.py (323줄) → ~150줄

#### @REQ:HOOKS-OPTIMIZE-001 성능 최적화

- 불필요한 import 제거
- 중복 로직 함수화
- 조건부 실행으로 불필요한 연산 제거
- 메모리 효율적인 데이터 구조 사용

### Specifications (상세 명세)

#### @DESIGN:SESSION-START-001 session_start_notice.py 리팩토링

**현재 문제점**

- 2,133줄의 비대한 코드
- 중복된 설정 검사 로직
- 불필요한 세부 출력

**리팩토링 전략**

```python
# Before (2,133줄)
# 복잡한 설정 검사, 상세한 출력, 중복 로직

# After (~200줄)
class SessionNotifier:
    def __init__(self):
        self.config = load_minimal_config()

    def notify_session_start(self):
        """핵심 세션 정보만 출력"""
        self._show_project_status()
        self._show_recent_activity()
        self._show_next_actions()

    def _show_project_status(self):
        """프로젝트 상태 요약"""
        pass

    def _show_recent_activity(self):
        """최근 활동 요약"""
        pass

    def _show_next_actions(self):
        """다음 액션 제안"""
        pass
```

#### @DESIGN:CHECKPOINT-MANAGER-001 checkpoint_manager.py 통합 설계

**통합 대상**

- auto_checkpoint.py (222줄): 자동 체크포인트 생성
- file_watcher.py (323줄): 파일 변경 감지

**통합 후 구조**

```python
# checkpoint_manager.py (~150줄)
class CheckpointManager:
    def __init__(self):
        self.watcher = FileWatcher()  # 경량화된 파일 감지
        self.auto_save = AutoSave()   # 간소화된 체크포인트

    def start_monitoring(self):
        """파일 변경 감지 및 자동 체크포인트"""
        pass
```

#### @TASK:CLEANUP-SEQUENCE-001 제거 순서

1. **1단계**: tag_validator.py, check_style.py 제거
2. **2단계**: session_start_notice.py 리팩토링
3. **3단계**: auto_checkpoint.py + file_watcher.py 통합
4. **4단계**: 전체 성능 검증

#### @TASK:MIGRATION-SAFETY-001 안전한 마이그레이션

**백업 전략**

```bash
# 기존 hooks 백업
cp -r .claude/hooks/moai .claude/hooks/moai.backup.$(date +%Y%m%d)
```

**단계별 테스트**

- 각 단계마다 기능 동작 확인
- 롤백 시나리오 준비
- 세션 시작 시간 측정

### Traceability (추적성 태그)

- **연결**: @REQ:HOOKS-CORE-001 → @DESIGN:SESSION-START-001 → @TASK:CLEANUP-SEQUENCE-001
- **품질**: @PERF:SESSION-START-001 (세션 시작 시간 50% 단축)
- **보안**: @SEC:HOOKS-SAFETY-001 (기능 유지하면서 최적화)
- **문서**: @DOCS:HOOKS-GUIDE-001 (리팩토링된 hook 사용 가이드)

---

**관련 문서**

- @.moai/project/tech.md: 기술 스택 및 품질 정책
- @.moai/memory/development-guide.md: TRUST 5원칙 및 코딩 규칙
- @CLAUDE.md: MoAI-ADK 전체 워크플로우 컨텍스트
