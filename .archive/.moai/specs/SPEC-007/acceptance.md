# SPEC-007 Acceptance Criteria: .claude/hooks 최적화

## @TEST:ACCEPTANCE-OVERVIEW-001 수락 기준 개요

### 성공 정의

- 전체 hook 코드량 74% 감소 (3,853줄 → ~1,000줄)
- 세션 시작 시간 50% 단축 (3-5초 → 1.5-2.5초)
- 모든 핵심 기능 정상 동작 유지
- TDD 테스트 커버리지 85% 이상 달성
- TRUST 5원칙 완전 준수

### 검증 방법

- 자동화된 성능 벤치마크 테스트
- Given-When-Then 형식의 상세 시나리오 테스트
- 기존 기능 호환성 회귀 테스트
- 사용자 수용 테스트 (UAT)

## @TEST:PERFORMANCE-ACCEPTANCE-001 성능 수락 기준

### AC1-P1: 세션 시작 시간 단축

**Given** Claude Code 세션이 시작될 때
**When** 모든 hook 파일이 로딩되면
**Then** 총 로딩 시간이 2.5초 이하여야 함

**측정 방법**:

```python
def test_session_start_performance():
    start_time = time.time()
    load_all_hooks()
    elapsed = time.time() - start_time
    assert elapsed <= 2.5, f"Session start took {elapsed}s, expected ≤2.5s"
```

**성공 기준**:

- 평균 세션 시작 시간 ≤ 2.5초
- 최대 세션 시작 시간 ≤ 3.0초
- 95% 신뢰구간에서 일관된 성능

### AC1-P2: 메모리 사용량 감소

**Given** hook 파일들이 메모리에 로딩될 때
**When** 메모리 사용량을 측정하면
**Then** 기존 대비 30% 이상 감소해야 함

**측정 방법**:

```python
@memory_profiler.profile
def test_memory_usage_reduction():
    baseline_memory = get_baseline_memory_usage()
    optimized_memory = measure_optimized_hooks_memory()
    reduction_ratio = (baseline_memory - optimized_memory) / baseline_memory
    assert reduction_ratio >= 0.3, f"Memory reduction {reduction_ratio:.2%}, expected ≥30%"
```

**성공 기준**:

- 메모리 사용량 30% 이상 감소
- 메모리 누수 없음 (장기간 실행 테스트)
- 가비지 컬렉션 빈도 감소

### AC1-P3: 파일 I/O 최적화

**Given** hook이 파일 시스템에 접근할 때
**When** 파일 I/O 횟수를 계산하면
**Then** 기존 대비 50% 이상 감소해야 함

**측정 방법**:

```python
def test_file_io_optimization():
    with patch('builtins.open') as mock_open:
        load_optimized_hooks()
        io_count = mock_open.call_count
        baseline_io = 100  # 기존 측정값
        reduction_ratio = (baseline_io - io_count) / baseline_io
        assert reduction_ratio >= 0.5, f"I/O reduction {reduction_ratio:.2%}, expected ≥50%"
```

## @TEST:FUNCTIONALITY-ACCEPTANCE-001 기능 수락 기준

### AC2-F1: session_start_notice.py 핵심 기능 보존

**Given** 세션이 시작될 때
**When** session_start_notice.py가 실행되면
**Then** 다음 핵심 기능들이 정상 동작해야 함:

1. **MoAI 개발 가이드 위반 감지**

```python
def test_constitution_violation_detection():
    # Given: 개발 가이드 위반 상황 설정
    create_violation_scenario()

    # When: 세션 시작 알림 실행
    result = session_start_notice.check_violations()

    # Then: 위반 사항이 정확히 감지됨
    assert result.has_violations is True
    assert "TRUST principles" in result.violations
```

2. **프로젝트 초기화 상태 확인**

```python
def test_project_initialization_check():
    # Given: 미초기화된 프로젝트
    setup_uninitialized_project()

    # When: 세션 시작 체크 실행
    result = session_start_notice.check_project_status()

    # Then: 초기화 필요 알림 표시
    assert result.needs_initialization is True
    assert "/moai:0-project" in result.recommendations
```

3. **중요 설정 누락 알림**

```python
def test_missing_config_notification():
    # Given: 설정 파일이 누락된 상태
    remove_critical_configs()

    # When: 설정 검사 실행
    result = session_start_notice.check_configurations()

    # Then: 누락된 설정에 대한 알림
    assert result.missing_configs
    assert ".claude/settings.json" in result.missing_configs
```

### AC2-F2: file_monitor.py 통합 기능 검증

**Given** 파일 변경이 발생할 때
**When** file_monitor.py가 이벤트를 처리하면
**Then** 파일 감지와 체크포인트 생성이 통합 동작해야 함

**시나리오 1: 파일 변경 감지**

```python
def test_file_change_detection():
    # Given: 파일 모니터가 활성화된 상태
    monitor = FileMonitor()
    monitor.start()

    # When: 파일 변경 발생
    modify_test_file("test.py")
    time.sleep(0.1)  # 이벤트 처리 대기

    # Then: 변경 사항이 감지됨
    assert monitor.last_change_detected is not None
    assert "test.py" in monitor.changed_files
```

**시나리오 2: 조건부 체크포인트 생성**

```python
def test_conditional_checkpoint_creation():
    # Given: 체크포인트 생성 조건 설정
    monitor = FileMonitor(checkpoint_threshold=5)

    # When: 설정된 임계값만큼 파일 변경
    for i in range(5):
        modify_test_file(f"test_{i}.py")

    # Then: 자동 체크포인트 생성됨
    assert monitor.checkpoint_created is True
    assert git_has_new_commit() is True
```

### AC2-F3: 보안 기능 유지 (pre_write_guard.py)

**Given** 파일 쓰기 요청이 발생할 때
**When** pre_write_guard.py가 검사를 수행하면
**Then** 핵심 보안 검사가 정상 동작해야 함

**시나리오 1: 민감정보 검사**

```python
def test_sensitive_info_detection():
    # Given: 민감정보가 포함된 콘텐츠
    sensitive_content = "API_KEY=abc123secret"

    # When: 사전 쓰기 검사 실행
    result = pre_write_guard.check_content(sensitive_content)

    # Then: 민감정보가 감지되고 차단됨
    assert result.blocked is True
    assert "API_KEY" in result.sensitive_patterns
```

**시나리오 2: 위험 파일 차단**

```python
def test_dangerous_file_blocking():
    # Given: 위험한 시스템 파일 경로
    dangerous_path = "/etc/passwd"

    # When: 파일 쓰기 시도
    result = pre_write_guard.check_file_path(dangerous_path)

    # Then: 접근이 차단됨
    assert result.allowed is False
    assert "system file" in result.block_reason
```

## @TEST:COMPATIBILITY-ACCEPTANCE-001 호환성 수락 기준

### AC3-C1: Claude Code API 호환성

**Given** 기존 에이전트가 hook과 통신할 때
**When** 최적화된 hook이 응답하면
**Then** 기존과 동일한 인터페이스를 유지해야 함

```python
def test_api_compatibility():
    # Given: 기존 에이전트 호출 패턴
    agent_request = {
        "action": "session_start",
        "parameters": {"project_path": "/test"}
    }

    # When: 최적화된 hook 호출
    response = optimized_hook.handle_request(agent_request)

    # Then: 기존 응답 형식 유지
    expected_fields = ["status", "message", "recommendations"]
    assert all(field in response for field in expected_fields)
    assert response["status"] in ["success", "warning", "error"]
```

### AC3-C2: 설정 파일 호환성

**Given** 기존 .claude/settings.json 설정이 있을 때
**When** 최적화된 hook들이 로딩되면
**Then** 모든 설정이 정상적으로 적용되어야 함

```python
def test_settings_compatibility():
    # Given: 기존 설정 파일
    existing_settings = load_existing_settings()

    # When: 최적화된 hook에서 설정 로딩
    loaded_settings = optimized_hooks.load_settings()

    # Then: 모든 설정값이 일치
    assert loaded_settings == existing_settings
    assert all_hooks_configured_correctly(loaded_settings)
```

## @TEST:CODE-QUALITY-ACCEPTANCE-001 코드 품질 수락 기준

### AC4-Q1: TRUST 5원칙 준수

**Test First (T)**

```python
def test_coverage_requirement():
    # Given: 테스트 커버리지 측정
    coverage = measure_test_coverage()

    # Then: 85% 이상 달성
    assert coverage >= 0.85, f"Coverage {coverage:.1%}, expected ≥85%"
```

**Readable (R)**

```python
def test_file_size_limits():
    # Given: 모든 hook 파일 검사
    hook_files = get_all_hook_files()

    # Then: 모든 파일이 300줄 이하
    for file_path in hook_files:
        line_count = count_lines(file_path)
        assert line_count <= 300, f"{file_path} has {line_count} lines, expected ≤300"
```

**Unified (U)**

```python
def test_no_circular_dependencies():
    # Given: 모듈 의존성 그래프 분석
    dependency_graph = analyze_dependencies()

    # Then: 순환 의존성 없음
    circular_deps = find_circular_dependencies(dependency_graph)
    assert len(circular_deps) == 0, f"Circular dependencies found: {circular_deps}"
```

**Secured (S)**

```python
def test_security_features_preserved():
    # Given: 보안 기능 테스트 스위트
    security_tests = [
        test_secret_detection,
        test_dangerous_command_blocking,
        test_file_permission_checks
    ]

    # Then: 모든 보안 테스트 통과
    for test in security_tests:
        assert test() is True
```

**Trackable (T)**

```python
def test_traceability_tags():
    # Given: TAG 시스템 검증
    tag_index = load_tag_index()

    # Then: 모든 주요 기능에 TAG 존재
    required_tags = ["@SPEC:HOOKS-OPTIMIZE-001", "@CODE:SESSION-START-001"]
    for tag in required_tags:
        assert tag in tag_index, f"Missing traceability tag: {tag}"
```

### AC4-Q2: 정적 분석 통과

```python
def test_static_analysis():
    # Given: 정적 분석 도구 실행
    mypy_result = run_mypy("hooks/")
    flake8_result = run_flake8("hooks/")

    # Then: 모든 도구 통과
    assert mypy_result.exit_code == 0, f"mypy errors: {mypy_result.errors}"
    assert flake8_result.exit_code == 0, f"flake8 errors: {flake8_result.errors}"
```

## @TEST:USER-ACCEPTANCE-001 사용자 수용 테스트

### AC5-U1: 개발자 경험 개선

**Given** 개발자가 Claude Code 세션을 시작할 때
**When** 최적화된 hook들이 실행되면
**Then** 체감 성능 개선이 명확해야 함

**사용자 시나리오 1: 빠른 세션 시작**

```python
def test_perceived_performance_improvement():
    # Given: 사용자 세션 시작 타이밍 측정
    start_times = []
    for _ in range(10):
        start_time = measure_user_session_start()
        start_times.append(start_time)

    # Then: 평균 시작 시간이 목표 달성
    avg_start_time = sum(start_times) / len(start_times)
    assert avg_start_time <= 2.0, f"Average start time {avg_start_time}s, expected ≤2.0s"
```

**사용자 시나리오 2: 정보 품질 유지**

```python
def test_information_quality_preserved():
    # Given: 사용자가 중요한 정보를 기대하는 상황
    scenarios = [
        "missing_project_initialization",
        "constitution_violations",
        "security_warnings"
    ]

    # Then: 모든 중요 정보가 여전히 제공됨
    for scenario in scenarios:
        result = simulate_user_scenario(scenario)
        assert result.important_info_provided is True
        assert result.user_satisfaction >= 4.0  # 5점 척도
```

### AC5-U2: 기능 완성도

**Given** 개발자가 기존 워크플로우를 수행할 때
**When** 최적화된 hook들과 상호작용하면
**Then** 모든 기능이 기대한 대로 동작해야 함

```python
def test_workflow_completion():
    workflows = [
        "/moai:0-project",
        "/moai:1-spec",
        "/moai:2-build",
        "/moai:3-sync"
    ]

    for workflow in workflows:
        # When: 워크플로우 실행
        result = execute_workflow(workflow)

        # Then: 성공적 완료
        assert result.success is True
        assert result.hook_interactions_normal is True
```

## @TEST:REGRESSION-PREVENTION-001 회귀 방지 테스트

### AC6-R1: 기존 기능 회귀 없음

```python
def test_no_functionality_regression():
    # Given: 기존 기능들의 기준 동작
    baseline_behaviors = load_baseline_test_results()

    # When: 최적화 후 동일 테스트 실행
    current_behaviors = run_comprehensive_regression_tests()

    # Then: 모든 기능이 기존과 동일하게 동작
    for function_name, baseline in baseline_behaviors.items():
        current = current_behaviors[function_name]
        assert current.equivalent_to(baseline), f"Regression in {function_name}"
```

### AC6-R2: 에러 처리 일관성

```python
def test_error_handling_consistency():
    error_scenarios = [
        "invalid_file_permissions",
        "missing_dependencies",
        "corrupted_config_files",
        "network_unavailable"
    ]

    for scenario in error_scenarios:
        # When: 에러 상황 발생
        result = simulate_error_scenario(scenario)

        # Then: 적절한 에러 처리 및 사용자 피드백
        assert result.error_handled is True
        assert result.user_feedback_clear is True
        assert result.system_recovery_possible is True
```

## @TEST:INTEGRATION-ACCEPTANCE-001 통합 테스트

### AC7-I1: 전체 시스템 통합

**Given** MoAI-ADK 전체 시스템이 구동될 때
**When** 최적화된 hook들과 다른 컴포넌트가 상호작용하면
**Then** 모든 통합 지점이 정상 동작해야 함

```python
def test_end_to_end_integration():
    # Given: 전체 시스템 구동
    system = start_moai_adk_system()

    # When: 복잡한 사용자 워크플로우 실행
    workflow_result = execute_complex_workflow([
        "project_initialization",
        "spec_creation",
        "tdd_implementation",
        "documentation_sync"
    ])

    # Then: 모든 단계 성공적 완료
    assert workflow_result.all_steps_completed is True
    assert workflow_result.hook_integrations_successful is True
    assert workflow_result.performance_acceptable is True
```

---

## 최종 검증 체크리스트

### 필수 통과 기준 (모든 항목 ✅ 필요)

#### 성능 기준

- [ ] 세션 시작 시간 ≤ 2.5초 달성
- [ ] 메모리 사용량 30% 이상 감소
- [ ] 파일 I/O 50% 이상 감소
- [ ] 코드량 74% 감소 달성 (3,853줄 → ~1,000줄)

#### 기능 기준

- [ ] 모든 핵심 기능 정상 동작
- [ ] 기존 에이전트와 100% 호환성
- [ ] 보안 기능 완전 보존
- [ ] 사용자 워크플로우 중단 없음

#### 품질 기준

- [ ] 테스트 커버리지 85% 이상
- [ ] 모든 파일 300줄 이하
- [ ] 정적 분석 도구 통과
- [ ] TRUST 5원칙 완전 준수

#### 사용자 수용 기준

- [ ] 체감 성능 개선 확인
- [ ] 기능 완성도 검증
- [ ] 회귀 테스트 통과
- [ ] 사용자 만족도 4.0/5.0 이상

### 선택적 향상 기준 (추가 점수)

- [ ] 세션 시작 시간 2.0초 이하 달성
- [ ] 메모리 사용량 50% 이상 감소
- [ ] 사용자 만족도 4.5/5.0 이상
- [ ] 테스트 커버리지 90% 이상

---

_이 수락 기준은 SPEC-007의 모든 목표가 성공적으로 달성되었음을 객관적으로 검증하며, TDD 접근 방식을 통해 품질 있는 리팩토링을 보장합니다._
