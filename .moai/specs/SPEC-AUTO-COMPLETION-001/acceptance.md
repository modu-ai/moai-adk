---
@META: {
  "id": "ACCEPT-AUTO-COMPLETION-001",
  "spec_id": "SPEC-AUTO-COMPLETION-001",
  "title": "자동 SPEC 생성 완성화 시스템 검수 기준",
  "title_en": "Automated SPEC Completion System Acceptance Criteria",
  "version": "1.0.0",
  "status": "draft",
  "created": "2025-11-11",
  "author": "@user"
}
---

# @ACCEPT:AUTO-COMPLETION-001: 검수 기준
## Acceptance Criteria

### 검수 개요 (Acceptance Overview)

자동 SPEC 생성 완성화 시스템이 요구사항을 모두 충족하는지 검증하기 위한 상세 기준입니다. 기능적, 비기능적, 사용자 경험 관점에서의 검증 항목을 포함합니다.

### 기능적 검수 기준 (Functional Acceptance Criteria)

#### AC-001: 코드 변경 감지 기능
**Given** 사용자가 Python/JavaScript/TypeScript/Go 코드 파일을 생성하거나 수정할 때
**When** Write/Edit/MultiEdit 툴이 실행된 후
**Then** 시스템은 변경된 파일 목록을 정확히 식별해야 함

**검증 방법:**
```bash
# 테스트 시나리오 1: 신규 파일 생성
echo "def hello_world(): pass" > new_feature.py

# Expected: post_tool__auto_spec_completion.py 실행됨
# Expected: .moai/specs/SPEC-NEW-FEATURE-001/ 디렉토리 생성됨
```

#### AC-002: SPEC 미존재 확인 기능
**Given** 코드 파일이 생성되었지만 해당하는 SPEC이 없을 때
**When** 자동 생성 프로세스가 실행될 때
**Then** 시스템은 SPEC 부재를 확인하고 생성을 진행해야 함

**검증 방법:**
```bash
# 테스트 시나리오 2: 기존 파일 수정 (SPEC 없음)
echo "def calculate_sum(a, b): return a + b" > calculator.py

# Expected: SPEC-CALCULATOR-001 생성됨
# Not Expected: 기존 SPEC이 있으면 생성 안 됨
```

#### AC-003: 완전한 SPEC 문서 생성 기능
**Given** 코드 분석이 완료되고 confidence score가 임계치 이상일 때
**When** SPEC 생성이 실행될 때
**Then** 3개의 완전한 파일(spec.md, plan.md, acceptance.md)이 생성되어야 함

**검증 방법:**
```bash
# 검증 항목
ls .moai/specs/SPEC-CALCULATOR-001/
# Expected 출력:
# - spec.md (EARS 형식, 500+ 단어)
# - plan.md (구현 계획, 300+ 단어)
# - acceptance.md (검수 기준, 200+ 단어)
```

#### AC-004: EARS 형식 준수 기능
**Given** 생성된 spec.md 파일을 열었을 때
**When** EARS 형식 요구사항을 검증할 때
**Then** 모든 필수 섹션(환경, 가정, 요구사항, 명세)이 포함되어야 함

**검증 방법:**
```python
def validate_ears_format(spec_path: str) -> bool:
    """EARS 형식 준수 검증"""
    required_sections = [
        "## 개요 (Overview)",
        "## 환경 (Environment)",
        "## 가정 (Assumptions)",
        "## 요구사항 (Requirements)",
        "## 명세 (Specifications)",
        "### 보편적 요구사항",
        "### 상태 기반 요구사항",
        "### 이벤트 기반 요구사항"
    ]

    with open(spec_path, 'r') as f:
        content = f.read()

    for section in required_sections:
        if section not in content:
            return False
    return True
```

#### AC-005: 중복 생성 방지 기능
**Given** 이미 SPEC이 존재하는 코드 파일을 수정할 때
**When** 자동 생성 프로세스가 실행될 때
**Then** 기존 SPEC을 덮어쓰지 않고 알림만 표시해야 함

**검증 방법:**
```bash
# 기존 SPEC 있는 상태에서
echo "# Updated function" >> calculator.py

# Expected: "SPEC-CALCULATOR-001 already exists" 메시지
# Not Expected: spec.md 파일 덮어쓰기
```

### 비기능적 검수 기준 (Non-Functional Acceptance Criteria)

#### AC-006: 성능 요구사항
**Given** 자동 SPEC 생성 프로세스가 실행될 때
**When** 전체 실행 시간을 측정할 때
**Then** 총 실행 시간이 2초 이내여야 함

**검증 방법:**
```python
import time
from subprocess import run

def measure_performance():
    start = time.time()

    # Hook 실행 시뮬레이션
    result = run([
        "python3",
        ".claude/hooks/alfred/post_tool__auto_spec_completion.py"
    ], capture_output=True, text=True)

    duration = time.time() - start

    assert duration < 2.0, f"Too slow: {duration:.2f}s > 2.0s"
    assert result.returncode == 0, f"Hook failed: {result.stderr}"
```

#### AC-007: 신뢰도 필터링 기능
**Given** 코드 분석 결과 confidence score가 계산되었을 때
**When** 설정된 최소 신뢰도(0.7)와 비교할 때
**Then** 낮은 confidence score일 경우 생성을 건너뛰어야 함

**검증 방법:**
```python
def test_confidence_filtering():
    # Low confidence 코드 (단순 파일)
    low_conf_code = "print('hello')"
    assert should_generate_spec(low_conf_code) == False

    # High confidence 코드 (구조화된 모듈)
    high_conf_code = '''
class Calculator:
    """Advanced calculator with multiple operations."""

    def __init__(self):
        self.history = []

    def add(self, a, b):
        """Add two numbers."""
        result = a + b
        self.history.append(f"{a} + {b} = {result}")
        return result
'''
    assert should_generate_spec(high_conf_code) == True
```

#### AC-008: 오류 처리 기능
**Given** SPEC 생성 중 예외 상황이 발생할 때
**When** 오류가 처리될 때
**Then** 그레이스풀 데그레이션으로 시스템이 계속 동작해야 함

**검증 방법:**
```python
def test_error_handling():
    # 시나리오 1: 권한 없는 디렉토리
    with patch('os.makedirs', side_effect=PermissionError()):
        result = execute_hook(tool_data)
        assert result['continue'] == True
        assert 'error' in result.get('hookSpecificOutput', {})

    # 시나리오 2: 분석 실패
    with patch('spec_generator.SpecGenerator', side_effect=Exception()):
        result = execute_hook(tool_data)
        assert result['continue'] == True
        assert 'error' in result.get('hookSpecificOutput', {})
```

### 사용자 경험 검수 기준 (User Experience Acceptance Criteria)

#### AC-009: 명확한 알림 기능
**Given** SPEC 자동 생성이 성공적으로 완료되었을 때
**When** 사용자에게 결과가 표시될 때
**Then** 성공 여부, 생성 위치, 신뢰도 정보가 명확히 표시되어야 함

**검증 방법:**
```python
def test_success_notification():
    result = execute_completion_hook({
        'toolName': 'Write',
        'toolArguments': {'file_path': 'new_service.py'},
        'result': {'success': True}
    })

    assert 'systemMessage' in result
    message = result['systemMessage']

    # 필수 정보 포함 확인
    assert '✅ 자동 SPEC 생성 완료' in message
    assert '.moai/specs/' in message
    assert '신뢰도' in message
    assert '%' in message
```

#### AC-010: 편집 가이드 제공 기능
**Given** 생성된 SPEC의 품질이 완벽하지 않을 때
**When** 사용자에게 편집 가이드가 표시될 때
**Then** 구체적인 개선 제안이 포함되어야 함

**검증 방법:**
```python
def test_editing_guidance():
    result = execute_completion_hook(sample_tool_data)

    if result['hookSpecificOutput']['spec_proposal']['confidence'] < 0.9:
        message = result['systemMessage']

        # 편집 가이드 항목 확인
        assert '📝 편집 가이드' in message or '📝 추천 편집' in message
        assert '1.' in message  # 번호 리스트
        assert len(message.split('\n')) >= 5  # 최소 5줄 이상
```

#### AC-011: 설정 기반 제어 기능
**Given** 사용자가 config.json에서 자동 생성 설정을 변경했을 때
**When** Hook이 실행될 때
**Then** 설정 값이 올바르게 적용되어야 함

**검증 방법:**
```json
// 테스트 설정 1: 비활성화
{
  "tags": {
    "policy": {
      "auto_spec_completion": {
        "enabled": false
      }
    }
  }
}

// Expected: Hook이 실행되지 않고 continue: True 반환
```

```python
def test_config_disabled():
    # config.json에 enabled: false 설정
    set_config_value('tags.policy.auto_spec_completion.enabled', False)

    result = execute_completion_hook(tool_data)

    assert result == {'continue': True}
```

### 통합 검수 기준 (Integration Acceptance Criteria)

#### AC-012: 기존 시스템과의 호환성
**Given** 기존의 pre_tool__auto_spec_proposal.py가 동작하는 환경일 때
**When** 두 개의 Hook이 함께 실행될 때
**Then** 충돌 없이 각자의 기능을 수행해야 함

**검증 방법:**
```bash
# 순서 테스트
echo "def feature(): pass" > test_feature.py

# Expected 순서:
# 1. pre_tool__auto_spec_proposal.py 실행 (제안 표시)
# 2. 코드 파일 생성됨
# 3. post_tool__auto_spec_completion.py 실행 (실제 생성)

# 두 Hook 모두 정상 실행되고 충돌 없음
```

#### AC-013: Git 워크플로우 통합
**Given** 생성된 SPEC 파일이 Git 추적 대상일 때
**When** git status를 실행할 때
**Then** 새로 생성된 파일이 untracked 상태로 표시되어야 함

**검증 방법:**
```bash
# Hook 실행 후
git status

# Expected 출력:
# Untracked files:
#   (use "git add <file>..." to include in what will be committed)
#         .moai/specs/SPEC-TEST-FEATURE-001/
```

### 정의의 완료 (Definition of Done)

각 기능은 다음 조건을 모두 만족해야 "완료"로 간주됩니다:

1. **기능 완성**: 모든 AC (Acceptance Criteria) 통과
2. **테스트 커버리지**: 85% 이상 단위/통합 테스트覆盖
3. **문서화**: 코드 주석, README 업데이트 완료
4. **성능 검증**: 모든 비기능적 AC 통과
5. **사용자 테스트**: 실제 사용자 피드백 수렴
6. **코드 리뷰**: 팀 멤버 코드 리뷰 완료
7. **통합 테스트**: 전체 시스템과의 연동 테스트 통과
8. **배포 준비**: 프로덕션 환경 배포 가능 상태

### 검수 체크리스트 (Acceptance Checklist)

- [ ] **AC-001**: 코드 변경 감지 정확성 (단위 테스트 100% 통과)
- [ ] **AC-002**: SPEC 미존재 확인 로직 (엣지 케이스 포함)
- [ ] **AC-003**: 3종 파일 완전성 검증 (템플릿完整性)
- [ ] **AC-004**: EARS 형식 엄격 준수 (자동 검증 도구)
- [ ] **AC-005**: 중복 생성 방지 (경쟁 상태 포함)
- [ ] **AC-006**: 성능 기준 충족 (2초 이내)
- [ ] **AC-007**: 신뢰도 필터링 정확성
- [ ] **AC-008**: 모든 예외 상황 처리
- [ ] **AC-009**: 사용자 친화적 알림 시스템
- [ ] **AC-010**: 유용한 편집 가이드 제공
- [ ] **AC-011**: 설정 시스템 완벽 통합
- [ ] **AC-012**: 기존 Hook과의 호환성
- [ ] **AC-013**: Git 워크플로우 자연스러운 통합