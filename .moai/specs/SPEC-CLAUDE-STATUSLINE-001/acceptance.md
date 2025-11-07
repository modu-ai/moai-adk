---
spec_id: CLAUDE-STATUSLINE-001
version: 1.2.0
created: 2025-11-07
---

# 수용 기준 및 테스트 시나리오

## 1. 기본 동작 테스트

### AC 1.1: Compact 모드 기본 렌더링
**Given**: Claude Code 세션이 활성화되고, `.moai/config.json` 에 `statusline_mode: "compact"` 설정
**When**: 상태줄 렌더러가 호출될 때
**Then**: 다음 형식의 텍스트가 생성되어야 함:
```
[MODEL] [DURATION] | [DIR] | [VERSION] [UPDATE-INDICATOR] | [BRANCH] | [GIT-STATUS] | [TASK]
```

**테스트 케이스:**
```python
def test_compact_rendering():
    renderer = StatuslineRenderer(mode="compact")
    output = renderer.render()

    assert output  # 비어있지 않음
    assert len(output) <= 80  # 80자 이하
    assert "H " in output or "S " in output  # 모델명 포함
    assert "0." in output  # 버전 정보 포함 (0.20.1 형태)
    assert "|" in output  # 구분자 포함
```

**수용 기준:**
- ✓ 출력이 비어있지 않음
- ✓ 80자 이하
- ✓ 모듈별 구분자 존재
- ✓ 버전 정보 포함
- ✓ 각 정보가 해당 위치에 있음

---

### AC 1.2: Extended 모드 상세 렌더링
**Given**: Claude Code 세션이 활성화되고, `.moai/config.json` 에 `statusline_mode: "extended"` 설정
**When**: 상태줄 렌더러가 호출될 때
**Then**: 다음 형식의 텍스트가 생성되어야 함:
```
[FULL_MODEL_NAME] | [FULL_DURATION] | [FULL_PATH] | [VERSION] (optional: latest version) | [BRANCH] | [GIT_DETAILS] | [TASK_DETAIL]
```

**테스트 케이스:**
```python
def test_extended_rendering():
    renderer = StatuslineRenderer(mode="extended")
    output = renderer.render()

    assert len(output) <= 120  # 120자 이하
    assert "Haiku" in output or "Sonnet" in output  # 전체 모델명
    assert "h " in output  # 시간 단위 표시 (예: "1h 30m")
    assert "v0." in output  # 버전 정보 포함 (v0.20.1 형태)
```

**수용 기준:**
- ✓ 출력이 120자 이하
- ✓ 전체 모델명 표시
- ✓ 시간 단위 정보 상세 표시
- ✓ 버전 정보 표시 (v prefix 포함 가능)

---

### AC 1.3: Minimal 모드 축약 렌더링
**Given**: Claude Code 세션이 활성화되고, `.moai/config.json` 에 `statusline_mode: "minimal"` 설정
**When**: 상태줄 렌더러가 호출될 때
**Then**: 다음 형식의 텍스트가 생성되어야 함:
```
[M] [T] [V] [B] [S]
```
(M=Model, T=Time, V=Version, B=Branch, S=Status)

**테스트 케이스:**
```python
def test_minimal_rendering():
    renderer = StatuslineRenderer(mode="minimal")
    output = renderer.render()

    assert len(output) <= 40  # 40자 이하
    assert output.count("|") >= 2  # 최소 구분자 포함
    assert "0." in output or "?" in output  # 버전 정보 포함
```

**수용 기준:**
- ✓ 출력이 40자 이하
- ✓ 핵심 정보만 표시 (버전 포함)
- ✓ 읽기 쉬운 형식

---

## 2. Git 정보 수집 테스트

### AC 2.1: Branch 이름 정확한 감지
**Given**: 로컬 Git 저장소에서 현재 branch가 `feature/SPEC-AUTH-001` 일 때
**When**: GitCollector의 collect_git_info() 메서드 호출
**Then**: branch 이름이 정확히 `feature/SPEC-AUTH-001` 로 반환되어야 함

**테스트 케이스:**
```python
def test_git_branch_detection():
    git_collector = GitCollector(repo_path="/path/to/repo")
    info = git_collector.collect_git_info()

    assert info.branch == "feature/SPEC-AUTH-001"
    assert info.branch_type == "feature"  # feature 감지
```

**테스트 환경:**
- 임시 Git 저장소 생성 (`pytest fixture`)
- 다양한 branch 이름으로 테스트:
  - `feature/SPEC-XXX` (feature branch)
  - `develop` (develop branch)
  - `main` (main branch)
  - `bugfix/issue-123` (다른 branch)

**수용 기준:**
- ✓ 현재 branch 정확히 감지
- ✓ branch 타입 분류 (feature/develop/main/other)

---

### AC 2.2: 변경 사항 개수 정확한 계산
**Given**: 다음의 파일 변경이 있을 때:
- Staged: 3개 파일 (+)
- Unstaged: 2개 파일 (M)
- Untracked: 1개 파일 (?)

**When**: GitCollector의 collect_git_info() 메서드 호출
**Then**: 다음 값들이 정확히 반환되어야 함:
- `staged_count = 3`
- `modified_count = 2`
- `untracked_count = 1`

**테스트 케이스:**
```python
def test_git_changes_counting():
    # 임시 저장소에서 파일 변경 생성
    git_collector = GitCollector(repo_path=temp_repo)
    info = git_collector.collect_git_info()

    assert info.staged_count == 3
    assert info.modified_count == 2
    assert info.untracked_count == 1

    # 포맷팅도 검증
    assert info.formatted_status == "+3 M2 ?1"
```

**테스트 시나리오:**
| Staged | Unstaged | Untracked | Expected Format |
|--------|----------|-----------|-----------------|
| 0 | 0 | 0 | (clean) |
| 5 | 0 | 0 | +5 |
| 2 | 3 | 1 | +2 M3 ?1 |
| 0 | 10 | 5 | M10 ?5 |

**수용 기준:**
- ✓ 모든 변경 사항 정확히 계산
- ✓ 포맷팅 정확함 (형식: `+N MN ?N`)
- ✓ 변경 없음 시 clean 표시

---

### AC 2.3: Git 명령 캐싱 효율성
**Given**: GitCollector가 초기화되었을 때
**When**: collect_git_info()를 5번 연속 호출
**Then**: 첫 번째 호출만 Git 명령 실행, 나머지 4개는 캐시에서 읽어야 함 (5초 TTL 내)

**테스트 케이스:**
```python
def test_git_caching():
    git_collector = GitCollector(repo_path=temp_repo, cache_ttl=5)

    with patch('subprocess.run') as mock_run:
        # 첫 호출 - Git 실행
        info1 = git_collector.collect_git_info()
        assert mock_run.call_count == 1

        # 2-5번 호출 - 캐시 사용 (Git 실행 안 됨)
        for _ in range(4):
            git_collector.collect_git_info()

        assert mock_run.call_count == 1  # 여전히 1 (캐시 사용)

        # 5초 후 - 캐시 만료, Git 다시 실행
        time.sleep(5.1)
        git_collector.collect_git_info()
        assert mock_run.call_count == 2
```

**수용 기준:**
- ✓ 캐시 유효 기간 내에는 Git 명령 실행 안 함
- ✓ TTL 만료 후 새로 조회
- ✓ 캐시 성능: 첫 조회 200ms 이상, 캐시 조회 <5ms

---

## 3. 세션 메트릭 추적 테스트

### AC 3.1: 세션 경과 시간 정확한 계산
**Given**: 세션이 30분 45초 경과했을 때
**When**: MetricsTracker의 track_duration() 메서드 호출
**Then**: 다음 형식으로 반환되어야 함:
- Compact 모드: `30m`
- Extended 모드: `30m 45s`

**테스트 케이스:**
```python
def test_duration_formatting():
    # Mock 세션 시작 시간
    start_time = datetime.now() - timedelta(minutes=30, seconds=45)

    tracker = MetricsTracker(session_start=start_time)

    # Compact 모드
    compact_duration = tracker.format_duration(mode="compact")
    assert compact_duration == "30m"

    # Extended 모드
    extended_duration = tracker.format_duration(mode="extended")
    assert extended_duration == "30m 45s"
```

**테스트 케이스 (경계값):**
| 경과 시간 | Compact | Extended |
|---------|---------|----------|
| 30초 | <1m | 30s |
| 5분 30초 | 5m | 5m 30s |
| 1시간 30분 | 1h | 1h 30m |
| 2시간 15분 45초 | 2h | 2h 15m |

**수용 기준:**
- ✓ 시간/분/초 계산 정확함
- ✓ Compact 모드: 분 단위 (30초 이상)
- ✓ Extended 모드: 분:초 단위

---

## 4. Alfred 작업 상태 감지 테스트

### AC 4.1: 활성 Alfred 명령 감지
**Given**: 현재 실행 중인 Alfred 명령이 `/alfred:2-run SPEC-AUTH-001` 일 때
**When**: AlfredDetector의 detect_current_task() 메서드 호출
**Then**: 다음 정보가 반환되어야 함:
- `command = "/alfred:2-run"`
- `spec_id = "SPEC-AUTH-001"`
- `status = "running"`

**테스트 케이스:**
```python
def test_alfred_command_detection():
    # Mock 세션 상태
    session_state = {
        "alfred_command": "/alfred:2-run",
        "spec_id": "SPEC-AUTH-001",
        "status": "running"
    }

    detector = AlfredDetector()
    with patch.object(detector, '_read_session_state', return_value=session_state):
        task = detector.detect_current_task()

        assert task.command == "/alfred:2-run"
        assert task.spec_id == "SPEC-AUTH-001"
        assert task.status == "running"
```

**테스트 케이스 (모든 명령):**
| 명령 | 예상 상태 | 표시 형식 |
|------|---------|---------|
| `/alfred:0-project` | initializing | [0-PROJECT] |
| `/alfred:1-plan` | planning | [1-PLAN] |
| `/alfred:2-run SPEC-AUTH-001` | running | [2-RUN: AUTH-001] |
| `/alfred:3-sync` | syncing | [3-SYNC] |

**수용 기준:**
- ✓ 모든 Alfred 명령 감지 가능
- ✓ SPEC ID 정확히 추출
- ✓ 작업 상태 정확함

---

### AC 4.2: TDD 단계 감지
**Given**: 현재 `/alfred:2-run` 중에 RED 단계일 때
**When**: AlfredDetector의 detect_tdd_stage() 메서드 호출
**Then**: 다음 정보가 반환되어야 함:
- `stage = "RED"`
- `description = "Writing failing tests"`

**테스트 케이스:**
```python
def test_tdd_stage_detection():
    detector = AlfredDetector()

    # TodoWrite에서 현재 in_progress 작업 읽기
    # 또는 Git commit history에서 마지막 커밋 메시지 확인

    stage = detector.detect_tdd_stage()

    assert stage in ["RED", "GREEN", "REFACTOR"]
    assert stage.description is not None
```

**테스트 데이터:**
| Stage | 기준 | 표시 |
|-------|------|------|
| RED | 테스트 파일 수정 커밋, TodoWrite in_progress | 🔴 |
| GREEN | 코드 파일 수정 커밋 | 🟡 |
| REFACTOR | 리팩토링 커밋 | 🟢 |

**수용 기준:**
- ✓ RED/GREEN/REFACTOR 정확히 판별
- ✓ 단계별 설명 명확함

---

## 5. 버전 정보 및 업데이트 안내 테스트

### AC 5.1: MoAI-ADK 버전 정보 읽기
**Given**: `.moai/config.json` 파일에 `version: "0.20.1"` 이 설정되어 있을 때
**When**: VersionReader의 read_version() 메서드 호출
**Then**: `version = "0.20.1"` 이 정확히 반환되어야 함

**테스트 케이스:**
```python
def test_version_reading():
    reader = VersionReader(config_path="/path/to/.moai/config.json")
    version_info = reader.read_version()

    assert version_info.current_version == "0.20.1"
    assert version_info.current_version is not None
    assert isinstance(version_info.current_version, str)
```

**테스트 데이터:**
| 설정 값 | 예상 반환 | 비고 |
|--------|---------|------|
| "0.20.1" | "0.20.1" | 표준 형식 |
| "v0.20.1" | "0.20.1" | v prefix 제거 |
| "0.20.1-beta" | "0.20.1-beta" | 프리릴리스 버전 |

**수용 기준:**
- ✓ 버전 문자열 정확히 읽기
- ✓ 파일 변경 감지 시 즉시 갱신 (60초 캐싱)
- ✓ 오류 처리: 파일 없음 시 `[???]` 반환

---

### AC 5.2: 업데이트 가용성 확인
**Given**: 현재 버전이 "0.20.1" 이고 최신 버전이 "0.21.0" 일 때
**When**: UpdateChecker의 check_for_update() 메서드 호출
**Then**: 다음 정보가 반환되어야 함:
- `latest_version = "0.21.0"`
- `update_available = True`
- `update_icon = "⬆️"` (또는 `[UPDATE]`)

**테스트 케이스:**
```python
def test_update_check():
    checker = UpdateChecker(current_version="0.20.1")

    with patch('requests.get') as mock_get:
        # PyPI API 응답 시뮬레이션
        mock_get.return_value.json.return_value = {
            "releases": {
                "0.21.0": [...]  # 최신 버전
            }
        }

        update_info = checker.check_for_update()

        assert update_info.latest_version == "0.21.0"
        assert update_info.available == True
        assert update_info.update_icon in ["⬆️", "↑", "[UPDATE]"]
```

**테스트 시나리오:**
| 현재 버전 | 최신 버전 | 업데이트 필요 | 예상 아이콘 |
|---------|---------|------------|-----------|
| 0.20.1 | 0.20.1 | False | (없음) |
| 0.20.1 | 0.21.0 | True | ⬆️ |
| 0.20.0 | 0.21.0 | True | ⬆️ |
| 0.21.0 | 0.20.1 | False | (없음) |

**수용 기준:**
- ✓ 버전 비교 알고리즘 정확함
- ✓ 300초 캐싱으로 API 호출 최소화
- ✓ API 실패 시 조용히 무시 (오류 표시 안 함)
- ✓ 아이콘 형식 통일

---

### AC 5.3: 상태줄에서 버전 + 업데이트 표시
**Given**: 현재 버전이 "0.20.1" 이고 업데이트가 "0.21.0" 으로 가능할 때
**When**: StatuslineRenderer의 render() 메서드 호출
**Then**: 다음 중 하나의 형식으로 표시:
```
Compact: 0.20.1 ⬆️ 0.21.0
Extended: v0.20.1 (latest: v0.21.0)
Minimal: 0.20.1↑
```

**테스트 케이스:**
```python
def test_version_update_display():
    renderer = StatuslineRenderer(
        mode="compact",
        current_version="0.20.1",
        update_available=True,
        latest_version="0.21.0"
    )

    output = renderer.render()

    assert "0.20.1" in output
    assert "⬆️" in output or "[UPDATE]" in output
    assert "0.21.0" in output
```

**색상 기준:**
- 버전 정보: 파란색 (38;5;33)
- 업데이트 아이콘: 주황색 (38;5;208)
- 최신 버전: 주황색 (38;5;208)

**수용 기준:**
- ✓ 버전과 업데이트 아이콘 모두 표시
- ✓ 포맷팅 정확함
- ✓ 색상 구분 명확함

---

## 6. 색상 및 포맷팅 테스트

### AC 6.1: ANSI 색상 정확한 적용
**Given**: ColorManager가 256-color 팔레트로 초기화되었을 때
**When**: apply_color("text", "feature_branch") 호출
**Then**: 다음 형식의 ANSI 코드가 포함된 텍스트가 반환되어야 함:
```
\033[38;5;226mtext\033[0m
```
(226 = 노란색, feature branch)

**테스트 케이스:**
```python
def test_ansi_color_codes():
    color_mgr = ColorManager(palette="256-color")

    # Feature branch 색상 (노란색)
    colored_text = color_mgr.apply_color("feature/SPEC-AUTH-001", "feature_branch")
    assert "\033[38;5;226m" in colored_text  # 노란색 코드
    assert "\033[0m" in colored_text  # 리셋 코드
```

**색상 매핑 검증:**
| 요소 | 색상 코드 | ANSI |
|------|---------|------|
| Feature branch | 226 | Yellow |
| Develop branch | 51 | Cyan |
| Main branch | 46 | Green |
| Staged (+) | 46 | Green |
| Modified (M) | 208 | Orange |
| Untracked (?) | 196 | Red |

**수용 기준:**
- ✓ 모든 색상 코드 정확함
- ✓ ANSI 리셋 코드 포함
- ✓ 16-color fallback 작동

---

### AC 6.2: 이모지 vs 기호 자동 선택
**Given**: ColorManager가 환경에 따라 초기화될 때
**When**: 터미널이 이모지 지원할 때
**Then**: 이모지 사용 (🔵, 🟢, 🔴)
**When**: 터미널이 이모지 미지원할 때
**Then**: 기호 사용 (●, ✓, ✗)

**테스트 케이스:**
```python
def test_emoji_fallback():
    # 이모지 지원
    color_mgr = ColorManager(emoji_support=True)
    icon = color_mgr.get_icon("active")
    assert icon == "🔵"

    # 이모지 미지원
    color_mgr = ColorManager(emoji_support=False)
    icon = color_mgr.get_icon("active")
    assert icon == "●"
```

**수용 기준:**
- ✓ 자동 감지 또는 수동 설정 가능
- ✓ Fallback 명확함
- ✓ 모든 환경에서 읽을 수 있음

---

## 7. 성능 테스트

### AC 7.1: 렌더링 속도 (300ms 주기 지원)
**Given**: 일반적인 프로젝트 상태일 때
**When**: StatuslineRenderer의 render() 메서드 실행
**Then**: 총 실행 시간이 50ms 이하여야 함

**테스트 케이스:**
```python
def test_rendering_performance():
    renderer = StatuslineRenderer()

    start = time.time()
    for _ in range(100):
        renderer.render()
    elapsed = (time.time() - start) / 100  # 평균 시간

    assert elapsed < 0.05  # 50ms 이하
```

**성능 기준:**
| 작업 | 목표 | 실제 |
|------|------|------|
| Git 정보 수집 (캐시) | <5ms | ? |
| 세션 메트릭 조회 (캐시) | <5ms | ? |
| 렌더링 (Compact) | <30ms | ? |
| 렌더링 (Extended) | <40ms | ? |
| 전체 주기 (300ms) | <50ms | ? |

**수용 기준:**
- ✓ 평균 렌더링 시간 <50ms
- ✓ 99th percentile <100ms
- ✓ 300ms 주기로 최대 6회 업데이트 가능

---

### AC 7.2: 메모리 사용량 제약
**Given**: 상태줄이 계속 업데이트될 때
**When**: 메모리 프로파일러로 측정
**Then**: 메모리 사용량이 5MB 이하여야 함

**테스트 케이스:**
```python
def test_memory_usage():
    import psutil
    process = psutil.Process()

    initial_memory = process.memory_info().rss / 1024 / 1024  # MB

    renderer = StatuslineRenderer()
    for _ in range(1000):
        renderer.render()

    final_memory = process.memory_info().rss / 1024 / 1024
    peak_memory = max(process.memory_info().rss) / 1024 / 1024

    assert peak_memory - initial_memory < 5  # 5MB 이내
```

**수용 기준:**
- ✓ 초기 메모리: <2MB
- ✓ 장시간 실행 후: <5MB
- ✓ 메모리 누수 없음 (증가 추세 없음)

---

## 8. 오류 처리 및 복원력 테스트

### AC 8.1: Git 오류 시 graceful 처리
**Given**: Git 명령이 실패했을 때 (예: not a git repo)
**When**: GitCollector의 collect_git_info() 호출
**Then**: Exception 대신 기본값을 반환해야 함:
- `branch = "[GIT N/A]"` (회색)
- `staged_count = 0`
- `modified_count = 0`

**테스트 케이스:**
```python
def test_git_error_handling():
    git_collector = GitCollector(repo_path="/not/a/repo")

    # Exception이 발생하지 않음
    info = git_collector.collect_git_info()

    assert info.branch == "[GIT N/A]"
    assert info.staged_count == 0
    assert info.is_error == True
```

**테스트 시나리오:**
| 오류 상황 | 예상 반환값 | 표시 |
|---------|-----------|------|
| Git repo 아님 | branch="[GIT N/A]" | 회색 |
| 권한 오류 | status="[RESTRICTED]" | 회색 |
| 명령 타임아웃 | status="[TIMEOUT]" | 회색 |

**수용 기준:**
- ✓ Exception 발생 안 함
- ✓ 기본값 명확함
- ✓ 오류 상황을 사용자에게 표시

---

### AC 8.2: 파일 읽기 오류 시 복원
**Given**: `.moai/memory/last-session-state.json` 읽기 실패했을 때
**When**: SessionReader의 read_session_state() 호출
**Then**: 기본값으로 폴백해야 함:
- `duration = "0s"`

**테스트 케이스:**
```python
def test_session_file_read_error():
    reader = SessionReader(state_file="/nonexistent/file.json")

    state = reader.read_session_state()

    assert state.duration == "0s"
    assert state.error_occurred == True
```

**수용 기준:**
- ✓ 파일 없음: 기본값 사용
- ✓ JSON 파싱 오류: 기본값 사용
- ✓ 권한 오류: 기본값 사용

---

### AC 8.3: 버전 정보 읽기 오류 시 복원
**Given**: `.moai/config.json` 읽기 또는 파싱이 실패했을 때
**When**: VersionReader의 read_version() 호출
**Then**: 예외 발생 없이 기본값을 반환해야 함:
- `current_version = "[???]"` (회색)

**테스트 케이스:**
```python
def test_version_read_error():
    reader = VersionReader(config_path="/nonexistent/file.json")

    version_info = reader.read_version()

    assert version_info.current_version == "[???]"
    assert version_info.error_occurred == True
```

**테스트 시나리오:**
| 오류 상황 | 예상 반환 | 색상 |
|---------|---------|------|
| 파일 없음 | [???] | 회색 |
| JSON 파싱 오류 | [???] | 회색 |
| 권한 오류 | [???] | 회색 |

**수용 기준:**
- ✓ 예외 발생 안 함
- ✓ 기본값 명확함
- ✓ 오류 로깅 (디버깅용)

---

### AC 8.4: 업데이트 확인 API 실패 시 처리
**Given**: PyPI API 또는 네트워크 오류로 인해 업데이트 확인 실패했을 때
**When**: UpdateChecker의 check_for_update() 호출
**Then**: 예외 발생 없이 아무것도 표시하지 않음 (무시)

**테스트 케이스:**
```python
def test_update_check_api_failure():
    checker = UpdateChecker(current_version="0.20.1")

    with patch('requests.get') as mock_get:
        mock_get.side_effect = requests.ConnectionError()

        update_info = checker.check_for_update()

        assert update_info.available == False
        assert update_info.latest_version is None
        # 상태줄에는 아무것도 표시 안 됨
```

**테스트 시나리오:**
| 오류 상황 | 예상 반환 | 상태줄 표시 |
|---------|---------|-----------|
| 연결 오류 | available=False | (없음) |
| 타임아웃 | available=False | (없음) |
| JSON 파싱 오류 | available=False | (없음) |

**수용 기준:**
- ✓ 예외 발생 안 함
- ✓ 조용히 무시 (사용자에게 알리지 않음)
- ✓ 300초 캐싱 유지 (이전 결과 사용)

---

## 9. 통합 테스트 (E2E)

### AC 9.1: 전체 상태줄 렌더링 통합 테스트
**Given**: 실제 MoAI-ADK 프로젝트 환경
**When**: StatuslineRenderer가 모든 정보를 수집하고 렌더링
**Then**: 완전한 상태줄이 생성되어야 함

**테스트 시나리오:**
```python
def test_full_statusline_integration(moai_project):
    """실제 프로젝트에서 상태줄 생성"""

    renderer = StatuslineRenderer(
        project_root=moai_project,
        mode="compact"
    )

    output = renderer.render()

    # 필수 요소 검증
    assert output
    assert "H " in output or "S " in output  # 모델명
    assert "|" in output  # 구분자
    assert "feature/" in output or "develop" in output or "main" in output  # branch

    # 길이 검증
    assert len(output) <= 80

    # 포맷팅 검증
    assert output[0] != " "  # 앞의 공백 없음
    assert output[-1] != " "  # 뒤의 공백 없음
```

**테스트 데이터:**
```python
@pytest.fixture
def moai_project(tmp_path):
    """임시 MoAI-ADK 프로젝트 생성"""
    # Git 저장소 초기화
    # .moai/config.json 생성
    # .moai/memory/last-session-state.json 생성
    # feature/SPEC-TEST-001 branch 생성
    return tmp_path
```

**수용 기준:**
- ✓ 모든 정보 수집 성공
- ✓ 포맷팅 정확함
- ✓ 길이 제약 지킴
- ✓ 색상 정확함

---

### AC 9.2: 장시간 실행 테스트
**Given**: 상태줄이 계속 업데이트될 때 (예: 1시간)
**When**: 300ms 주기로 12,000회 업데이트 실행
**Then**: 다음 조건을 만족해야 함:
- Memory: <10MB (증가 추세 없음)
- CPU: <2% (평균)
- 오류: 0건

**테스트 케이스:**
```python
def test_long_running_operation():
    """1시간 분량의 업데이트 시뮬레이션"""

    renderer = StatuslineRenderer()
    errors = []
    peak_memory = 0

    for i in range(12000):  # 12,000 * 300ms = 3,600s = 1h
        try:
            output = renderer.render()
            assert output

            if i % 100 == 0:  # 매 100회마다 메모리 체크
                current_memory = get_memory_usage()
                peak_memory = max(peak_memory, current_memory)

        except Exception as e:
            errors.append((i, e))

    assert len(errors) == 0  # 오류 없음
    assert peak_memory < 10  # 10MB 이하
```

**수용 기준:**
- ✓ 12,000회 무중단 실행
- ✓ 메모리 누수 없음
- ✓ CPU 사용률 안정적

---

## 10. 사용자 수용 테스트 (UAT)

### AC 10.1: 개발자 피드백 (알파)
**Given**: 내부 팀이 상태줄을 사용 중일 때
**When**: 1주일간 피드백 수집
**Then**: 다음을 달성해야 함:
- 만족도 4.0/5.0 이상
- 주요 기능 버그 0건
- 개선 제안 수집

**평가 항목:**
- [ ] 상태줄 정보가 유용한가?
- [ ] 읽기 쉬운가?
- [ ] 성능이 만족스러운가?
- [ ] 색상이 구분되는가?
- [ ] 오류 메시지가 명확한가?

---

### AC 10.2: 베타 테스트 (공개)
**Given**: GitHub에서 선택 사용자 참여
**When**: 2주간 테스트
**Then**: 다음을 달성해야 함:
- 최소 5명의 사용자 피드백
- 각 OS에서 성공적으로 작동 (macOS, Linux, Windows)
- 주요 버그 모두 해결

**환경 검증:**
- [ ] macOS 13.0+
- [ ] Linux (Ubuntu 20.04+)
- [ ] Windows 11 (WSL2)
- [ ] 256-color terminal support
- [ ] Python 3.10+

---

## Definition of Done (DoD)

상태줄 SPEC이 완료되었으려면 다음을 만족해야 함:

### 기능성
- [ ] 6가지 핵심 정보 모두 표시
- [ ] 3가지 디스플레이 모드 (Compact, Extended, Minimal)
- [ ] 색상 팔레트 완전 구현
- [ ] 이모지/기호 자동 fallback

### 성능
- [ ] 평균 렌더링 시간 <50ms
- [ ] 메모리 사용량 <5MB
- [ ] 캐싱 효율 >90%
- [ ] CPU 사용률 <2%

### 신뢰성
- [ ] 모든 오류 상황에서 graceful 처리
- [ ] 99% 이상의 uptime
- [ ] 12,000회 연속 실행 무오류
- [ ] 로깅 상세함

### 호환성
- [ ] macOS, Linux, Windows 모두 지원
- [ ] Python 3.10+ 지원
- [ ] Git 2.20+ 지원
- [ ] 256-color 및 16-color 지원

### 품질
- [ ] 코드 커버리지 85% 이상
- [ ] Pylint 점수 8.0 이상
- [ ] 타입 체킹 통과 (mypy)
- [ ] 모든 테스트 통과

### 문서
- [ ] 설정 가이드 작성
- [ ] 사용 예시 3개 이상
- [ ] 트러블슈팅 가이드
- [ ] API 문서 자동 생성
