# SPEC-009 수락 기준: TAG 시스템 SQLite 마이그레이션

**@TEST:ACCEPTANCE-CRITERIA-001** ← 수락 기준 정의  
**@TEST:PERFORMANCE-VALIDATION-001** ← 성능 검증 시나리오  
**@TEST:COMPATIBILITY-VALIDATION-001** ← 호환성 검증 시나리오

---

## 전체 수락 기준 개요

### 핵심 성공 지표
1. **성능 향상**: JSON 대비 10배 빠른 검색, 삽입, 업데이트
2. **호환성 유지**: 기존 API 100% 동일한 결과 반환
3. **안전한 마이그레이션**: 데이터 손실 0건, 롤백 가능
4. **메모리 효율성**: 현재 대비 50% 메모리 사용량 감소

### 데이터 규모
- **현재 상태**: 441개 TAG, 4,747줄, 136KB JSON 파일
- **테스트 규모**: 100개, 1,000개, 10,000개 TAG 시나리오
- **성능 목표**: 1,000개 TAG 검색 시 < 10ms

---

## 성능 검증 시나리오

### @TEST:SEARCH-PERFORMANCE-001 검색 성능 테스트

**Given**: 1,000개의 TAG가 SQLite 데이터베이스에 저장되어 있고
**When**: 특정 TAG를 검색할 때
**Then**: 
- 검색 시간이 10ms 이내여야 함
- JSON 방식 대비 10배 이상 빠른 성능을 보여야 함
- 메모리 사용량이 현재 대비 50% 이하여야 함

**테스트 시나리오**:
```python
def test_search_performance():
    # Given: 1,000개 TAG 준비
    db_manager = TagDatabaseManager(":memory:")
    tags = generate_test_tags(1000)
    db_manager.bulk_insert(tags)
    
    # When: 특정 TAG 검색
    start_time = time.perf_counter()
    result = db_manager.search_tag("@SPEC:USER-AUTH-001")
    end_time = time.perf_counter()
    
    # Then: 성능 기준 검증
    search_time = (end_time - start_time) * 1000  # ms
    assert search_time < 10.0, f"Search took {search_time}ms, expected < 10ms"
    assert len(result) > 0, "Search result should not be empty"
    assert result[0]['tag_key'] == "@SPEC:USER-AUTH-001"
```

### @TEST:INSERT-PERFORMANCE-001 삽입 성능 테스트

**Given**: 비어있는 SQLite 데이터베이스가 준비되고
**When**: 100개의 새로운 TAG를 삽입할 때
**Then**: 
- 전체 삽입 시간이 500ms 이내여야 함
- 개별 TAG 삽입 시간이 5ms 이내여야 함
- 중복 TAG 삽입 시 적절한 오류 처리가 되어야 함

**테스트 시나리오**:
```python
def test_insert_performance():
    # Given: 빈 데이터베이스
    db_manager = TagDatabaseManager(":memory:")
    new_tags = generate_test_tags(100)
    
    # When: 100개 TAG 삽입
    start_time = time.perf_counter()
    for tag in new_tags:
        db_manager.insert_tag(tag)
    end_time = time.perf_counter()
    
    # Then: 성능 기준 검증
    total_time = (end_time - start_time) * 1000
    avg_time_per_tag = total_time / 100
    
    assert total_time < 500.0, f"Total insert time {total_time}ms > 500ms"
    assert avg_time_per_tag < 5.0, f"Avg insert time {avg_time_per_tag}ms > 5ms"
    assert db_manager.count_tags() == 100, "All tags should be inserted"
```

### @TEST:MEMORY-USAGE-001 메모리 사용량 테스트

**Given**: 10,000개의 TAG가 있는 큰 데이터셋이 준비되고
**When**: SQLite 시스템과 JSON 시스템을 각각 로딩할 때
**Then**: 
- SQLite 시스템의 메모리 사용량이 JSON 대비 50% 이하여야 함
- 지연 로딩으로 초기 메모리 사용량이 최소화되어야 함
- 가비지 컬렉션 후에도 메모리 누수가 없어야 함

**테스트 시나리오**:
```python
def test_memory_usage():
    # Given: 대용량 테스트 데이터
    large_tags = generate_test_tags(10000)
    
    # When: JSON vs SQLite 메모리 사용량 비교
    json_memory = measure_json_memory_usage(large_tags)
    sqlite_memory = measure_sqlite_memory_usage(large_tags)
    
    # Then: 메모리 효율성 검증
    memory_reduction = (json_memory - sqlite_memory) / json_memory
    assert memory_reduction >= 0.5, f"Memory reduction {memory_reduction:.1%} < 50%"
    
    # 가비지 컬렉션 후 메모리 누수 확인
    gc.collect()
    final_memory = get_current_memory_usage()
    assert final_memory < sqlite_memory * 1.1, "Memory leak detected"
```

---

## 호환성 검증 시나리오

### @TEST:API-COMPATIBILITY-001 기존 API 호환성 테스트

**Given**: 기존 JSON 시스템과 새로운 SQLite 시스템이 같은 데이터로 초기화되고
**When**: 동일한 API 호출을 각각 실행할 때
**Then**: 
- 반환 결과가 100% 동일해야 함
- JSON 구조와 키 이름이 완전히 일치해야 함
- 타임스탬프 형식이 기존과 동일해야 함

**테스트 시나리오**:
```python
def test_api_compatibility():
    # Given: 동일한 데이터로 초기화
    json_system = JsonTagSystem("tags.json")
    sqlite_system = TagIndexAdapter("tags.db")
    
    test_data = load_test_tags()
    json_system.load_data(test_data)
    sqlite_system.migrate_from_json(test_data)
    
    # When: 동일한 API 호출
    json_result = json_system.get_tags()
    sqlite_result = sqlite_system.get_tags()
    
    # Then: 결과 100% 일치
    assert json_result.keys() == sqlite_result.keys()
    assert json_result['version'] == sqlite_result['version']
    assert json_result['statistics'] == sqlite_result['statistics']
    
    # 개별 TAG 데이터 비교
    for tag_key in json_result['index']:
        json_refs = json_result['index'][tag_key]
        sqlite_refs = sqlite_result['index'][tag_key]
        assert len(json_refs) == len(sqlite_refs)
        
        for json_ref, sqlite_ref in zip(json_refs, sqlite_refs):
            assert json_ref['file'] == sqlite_ref['file']
            assert json_ref['line'] == sqlite_ref['line']
            assert json_ref['context'] == sqlite_ref['context']
```

### @TEST:EXISTING-TOOLS-001 기존 도구 호환성 테스트

**Given**: `validate_tags.py` 스크립트가 SQLite 시스템과 통합되고
**When**: 동일한 TAG 데이터에 대해 스크립트를 실행할 때
**Then**: 
- JSON 버전과 동일한 출력 형식이어야 함
- 모든 검증 규칙이 동일하게 적용되어야 함
- 오류 메시지와 경고가 일치해야 함

**테스트 시나리오**:
```python
def test_validate_tags_compatibility():
    # Given: JSON과 SQLite 버전 준비
    setup_test_environment()
    
    # When: validate_tags.py 실행
    json_output = run_validate_script("--backend=json")
    sqlite_output = run_validate_script("--backend=sqlite")
    
    # Then: 출력 결과 비교
    assert json_output['summary'] == sqlite_output['summary']
    assert json_output['errors'] == sqlite_output['errors']
    assert json_output['warnings'] == sqlite_output['warnings']
```

### @TEST:MOAI-SYNC-001 `/moai:3-sync` 명령어 호환성

**Given**: SQLite로 전환된 TAG 시스템이 활성화되고
**When**: `/moai:3-sync` 명령어를 실행할 때
**Then**: 
- 기존과 동일한 sync-report.md가 생성되어야 함
- TAG 통계 정보가 정확해야 함
- 모든 TAG 참조가 올바르게 추적되어야 함

**테스트 시나리오**:
```python
def test_moai_sync_compatibility():
    # Given: SQLite 시스템 활성화
    migrate_to_sqlite()
    
    # When: sync 명령 실행
    result = execute_moai_sync()
    
    # Then: 보고서 내용 검증
    report = load_sync_report()
    assert report['total_tags'] == 441
    assert report['total_references'] == 770
    assert 'migration_info' in report
    assert report['performance_improvement'] > 10.0
```

---

## 마이그레이션 검증 시나리오

### @TEST:MIGRATION-INTEGRITY-001 데이터 무결성 테스트

**Given**: 441개 TAG를 포함한 실제 `tags.json` 파일이 있고
**When**: JSON → SQLite → JSON 완전한 라운드트립을 수행할 때
**Then**: 
- 원본과 결과가 100% 동일해야 함
- TAG 개수, 참조 개수가 정확해야 함
- 모든 메타데이터가 보존되어야 함

**테스트 시나리오**:
```python
def test_migration_integrity():
    # Given: 원본 JSON 파일
    original_data = load_json("tags.json")
    original_hash = calculate_data_hash(original_data)
    
    # When: JSON → SQLite → JSON 라운드트립
    migration_tool = TagMigrationTool()
    migration_tool.migrate_json_to_sqlite("tags.json", "test.db")
    migration_tool.migrate_sqlite_to_json("test.db", "restored.json")
    
    restored_data = load_json("restored.json")
    restored_hash = calculate_data_hash(restored_data)
    
    # Then: 데이터 완전성 검증
    assert original_hash == restored_hash, "Data integrity check failed"
    assert original_data['statistics'] == restored_data['statistics']
    
    # 개별 TAG 검증
    for tag_key in original_data['index']:
        assert tag_key in restored_data['index']
        assert len(original_data['index'][tag_key]) == len(restored_data['index'][tag_key])
```

### @TEST:ATOMIC-MIGRATION-001 원자적 마이그레이션 테스트

**Given**: 마이그레이션이 중간에 중단되는 상황을 시뮬레이션하고
**When**: 부분적으로 완료된 마이그레이션 상태에서
**Then**: 
- 원본 데이터가 손상되지 않아야 함
- 중간 상태의 SQLite 파일이 올바르게 정리되어야 함
- 재시작 시 처음부터 다시 마이그레이션이 가능해야 함

**테스트 시나리오**:
```python
def test_atomic_migration():
    # Given: 마이그레이션 중단 시뮬레이션
    original_backup = backup_file("tags.json")
    
    # When: 마이그레이션 중 예외 발생
    migration_tool = TagMigrationTool()
    try:
        with simulate_interruption(after_tags=200):
            migration_tool.migrate_json_to_sqlite("tags.json", "test.db")
    except MigrationInterrupted:
        pass
    
    # Then: 원자성 검증
    assert file_unchanged("tags.json", original_backup), "Original file was modified"
    assert not os.path.exists("test.db"), "Partial database file should be cleaned up"
    
    # 재시작 가능성 검증
    migration_result = migration_tool.migrate_json_to_sqlite("tags.json", "test.db")
    assert migration_result.success, "Migration should succeed on retry"
    assert migration_result.tags_migrated == 441
```

### @TEST:ROLLBACK-FUNCTIONALITY-001 롤백 기능 테스트

**Given**: SQLite로 성공적으로 마이그레이션이 완료된 상태에서
**When**: 사용자가 JSON으로 롤백을 요청할 때
**Then**: 
- 10초 이내에 롤백이 완료되어야 함
- 원본 JSON과 100% 동일한 데이터가 복원되어야 함
- 시스템이 JSON 모드로 자동 전환되어야 함

**테스트 시나리오**:
```python
def test_rollback_functionality():
    # Given: 마이그레이션 완료 상태
    original_json = load_json("tags.json")
    migrate_to_sqlite()
    
    # When: 롤백 실행
    start_time = time.time()
    rollback_result = execute_rollback("--to-json")
    rollback_time = time.time() - start_time
    
    # Then: 롤백 성능 및 정확성 검증
    assert rollback_time < 10.0, f"Rollback took {rollback_time}s > 10s"
    assert rollback_result.success, "Rollback should succeed"
    
    restored_json = load_json("tags.json")
    assert original_json == restored_json, "Rollback data mismatch"
    
    # 시스템 모드 전환 확인
    current_backend = get_current_backend()
    assert current_backend == "json", "System should switch back to JSON mode"
```

---

## 예외 상황 처리 시나리오

### @TEST:ERROR-HANDLING-001 데이터베이스 잠금 상황 테스트

**Given**: SQLite 데이터베이스가 다른 프로세스에 의해 잠겨있고
**When**: TAG 읽기/쓰기 작업을 시도할 때
**Then**: 
- 적절한 오류 메시지가 표시되어야 함
- 3회 재시도 후 graceful failure 처리
- 읽기 전용 모드로 fallback 가능해야 함

**테스트 시나리오**:
```python
def test_database_lock_handling():
    # Given: 데이터베이스 잠금 상황
    with database_lock_simulation("test.db"):
        db_manager = TagDatabaseManager("test.db")
        
        # When: 쓰기 작업 시도
        with pytest.raises(DatabaseLockError) as exc_info:
            db_manager.insert_tag(create_test_tag())
        
        # Then: 적절한 오류 처리
        assert "database is locked" in str(exc_info.value)
        assert exc_info.value.retry_count == 3
        
        # 읽기 전용 모드 fallback 확인
        readonly_manager = db_manager.get_readonly_mode()
        assert readonly_manager.can_read(), "Should allow read operations"
```

### @TEST:DISK-SPACE-001 디스크 공간 부족 상황 테스트

**Given**: 디스크 공간이 부족한 상황이고
**When**: 대용량 마이그레이션을 시도할 때
**Then**: 
- 마이그레이션이 시작되기 전에 공간 검사를 해야 함
- 공간 부족 시 명확한 오류 메시지 제공
- 부분적으로 생성된 파일이 자동으로 정리되어야 함

**테스트 시나리오**:
```python
def test_disk_space_handling():
    # Given: 디스크 공간 부족 시뮬레이션
    with limited_disk_space(available_mb=10):
        migration_tool = TagMigrationTool()
        
        # When: 대용량 마이그레이션 시도
        with pytest.raises(InsufficientDiskSpaceError) as exc_info:
            migration_tool.migrate_json_to_sqlite("large_tags.json", "test.db")
        
        # Then: 적절한 오류 처리
        assert "insufficient disk space" in str(exc_info.value)
        assert exc_info.value.required_mb > 10
        assert not os.path.exists("test.db"), "Partial file should be cleaned up"
```

### @TEST:CORRUPTED-DATABASE-001 손상된 데이터베이스 복구 테스트

**Given**: SQLite 데이터베이스 파일이 손상된 상태에서
**When**: 시스템이 데이터베이스에 접근을 시도할 때
**Then**: 
- 손상 감지가 즉시 이루어져야 함
- 자동 백업에서 복구 시도
- 복구 불가능 시 JSON 모드로 안전하게 fallback

**테스트 시나리오**:
```python
def test_corrupted_database_recovery():
    # Given: 손상된 데이터베이스
    create_corrupted_database("corrupted.db")
    
    # When: 데이터베이스 접근 시도
    db_manager = TagDatabaseManager("corrupted.db")
    
    with pytest.raises(DatabaseCorruptionError) as exc_info:
        db_manager.get_tags()
    
    # Then: 복구 시도 확인
    assert exc_info.value.corruption_detected, "Should detect corruption"
    assert exc_info.value.backup_attempted, "Should attempt backup recovery"
    
    # Fallback 모드 확인
    fallback_system = db_manager.get_fallback_system()
    assert fallback_system.backend == "json", "Should fallback to JSON"
    assert fallback_system.is_operational(), "Fallback should be operational"
```

---

## 성능 벤치마크 시나리오

### @TEST:SCALABILITY-001 확장성 테스트

**Given**: 100개, 1,000개, 10,000개 TAG를 가진 데이터셋이 준비되고
**When**: 각 규모에서 검색 성능을 측정할 때
**Then**: 
- 모든 규모에서 선형 성능을 유지해야 함
- 10,000개 TAG에서도 100ms 이내 응답
- 메모리 사용량이 예측 가능하게 증가해야 함

**테스트 시나리오**:
```python
def test_scalability():
    test_sizes = [100, 1000, 10000]
    performance_results = {}
    
    for size in test_sizes:
        # Given: 각 크기별 데이터셋
        db_manager = create_test_database(size)
        
        # When: 검색 성능 측정
        search_times = []
        for _ in range(10):  # 10회 반복 측정
            start_time = time.perf_counter()
            result = db_manager.search_random_tag()
            end_time = time.perf_counter()
            search_times.append((end_time - start_time) * 1000)
        
        avg_time = sum(search_times) / len(search_times)
        performance_results[size] = avg_time
    
    # Then: 확장성 검증
    assert performance_results[100] < 5.0, "100 tags: should be < 5ms"
    assert performance_results[1000] < 10.0, "1000 tags: should be < 10ms"
    assert performance_results[10000] < 100.0, "10000 tags: should be < 100ms"
    
    # 선형성 검증 (10배 데이터가 10배 이상 느려지지 않아야 함)
    ratio_1k_100 = performance_results[1000] / performance_results[100]
    ratio_10k_1k = performance_results[10000] / performance_results[1000]
    assert ratio_1k_100 < 5.0, "Performance should scale sub-linearly"
    assert ratio_10k_1k < 5.0, "Performance should scale sub-linearly"
```

### @TEST:CONCURRENT-ACCESS-001 동시 접근 테스트

**Given**: 여러 스레드가 동시에 SQLite 데이터베이스에 접근하고
**When**: 동시에 읽기/쓰기 작업을 수행할 때
**Then**: 
- 데이터 무결성이 보장되어야 함
- 데드락이 발생하지 않아야 함
- 성능 저하가 합리적인 수준 이내여야 함

**테스트 시나리오**:
```python
def test_concurrent_access():
    # Given: 공유 데이터베이스
    db_path = "concurrent_test.db"
    setup_test_database(db_path, 1000)
    
    # When: 동시 액세스 시뮬레이션
    def worker_thread(thread_id, results):
        db_manager = TagDatabaseManager(db_path)
        start_time = time.time()
        
        for i in range(50):  # 각 스레드가 50개 작업 수행
            if i % 2 == 0:
                # 읽기 작업
                result = db_manager.get_random_tag()
                assert result is not None
            else:
                # 쓰기 작업
                tag = create_test_tag(f"THREAD-{thread_id}-{i}")
                db_manager.insert_tag(tag)
        
        end_time = time.time()
        results[thread_id] = end_time - start_time
    
    # 5개 스레드 동시 실행
    threads = []
    results = {}
    
    for i in range(5):
        thread = threading.Thread(
            target=worker_thread, 
            args=(i, results)
        )
        threads.append(thread)
        thread.start()
    
    # 모든 스레드 완료 대기
    for thread in threads:
        thread.join()
    
    # Then: 동시성 검증
    assert len(results) == 5, "All threads should complete"
    avg_time = sum(results.values()) / len(results)
    assert avg_time < 10.0, f"Average thread time {avg_time}s too slow"
    
    # 데이터 무결성 확인
    db_manager = TagDatabaseManager(db_path)
    final_count = db_manager.count_tags()
    expected_count = 1000 + (5 * 25)  # 원본 + 새로 추가된 TAG
    assert final_count == expected_count, "Data integrity violation detected"
```

---

## 최종 수락 조건

### @TEST:FINAL-ACCEPTANCE-001 전체 시스템 수락 테스트

**Given**: 완전히 구현된 SQLite 마이그레이션 시스템이 있고
**When**: 실제 production 데이터(441개 TAG)로 전체 워크플로우를 수행할 때
**Then**: 모든 수락 기준을 만족해야 함

**최종 체크리스트**:

#### ✅ 성능 요구사항
- [ ] JSON 대비 10배 빠른 검색 속도 달성
- [ ] 1,000개 TAG 검색이 10ms 이내 완료
- [ ] TAG 삽입이 5ms 이내 완료
- [ ] 메모리 사용량 50% 감소 달성

#### ✅ 호환성 요구사항
- [ ] 기존 JSON API와 100% 동일한 결과 반환
- [ ] `validate_tags.py` 스크립트 정상 동작
- [ ] `/moai:3-sync` 명령어 정상 동작
- [ ] 모든 기존 도구 호환성 유지

#### ✅ 안정성 요구사항
- [ ] 441개 TAG 마이그레이션 100% 성공률
- [ ] 라운드트립 마이그레이션 데이터 손실 0건
- [ ] 롤백 기능 10초 이내 완료
- [ ] 예외 상황 적절한 처리 및 복구

#### ✅ 사용성 요구사항
- [ ] 마이그레이션 진행률 실시간 표시
- [ ] 명확한 오류 메시지 및 해결 방안 제공
- [ ] 사용자 문서 완성도 90% 이상
- [ ] 새로운 사용자도 문서만으로 마이그레이션 성공

**최종 검증 시나리오**:
```python
def test_final_system_acceptance():
    # 전체 워크플로우 테스트
    original_system = JsonTagSystem("tags.json")
    
    # 1. 마이그레이션 실행
    migration_tool = TagMigrationTool()
    migration_result = migration_tool.migrate_json_to_sqlite(
        "tags.json", "production.db"
    )
    assert migration_result.success
    assert migration_result.tags_migrated == 441
    
    # 2. 성능 검증
    sqlite_system = TagIndexAdapter("production.db")
    performance = measure_performance_improvement(original_system, sqlite_system)
    assert performance.search_speedup >= 10.0
    assert performance.memory_reduction >= 0.5
    
    # 3. 호환성 검증
    compatibility = verify_api_compatibility(original_system, sqlite_system)
    assert compatibility.match_percentage == 100.0
    
    # 4. 안정성 검증
    rollback_result = migration_tool.migrate_sqlite_to_json(
        "production.db", "restored.json"
    )
    assert rollback_result.success
    assert files_identical("tags.json", "restored.json")
    
    print("🎉 SPEC-009 SQLite 마이그레이션 모든 수락 기준 통과!")
```

---

**수락 기준 요약**: 성능 10배 향상, 호환성 100% 유지, 안전한 마이그레이션, 완전한 롤백 기능을 모두 만족하는 TAG 시스템 SQLite 마이그레이션이 성공적으로 구현되어야 합니다.