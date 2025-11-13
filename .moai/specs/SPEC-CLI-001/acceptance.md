
> **Given-When-Then 형식의 구체적인 시나리오**
>
> 모든 인수 기준은 자동화된 테스트로 검증 가능해야 합니다.

---

## AC-1: doctor 명령어 - Python 프로젝트 도구 감지

### Scenario 1.1: pytest 설치, mypy 미설치

**Given**:
- Python 3.13.1 프로젝트
- pyproject.toml 존재
- pytest 8.4.2 설치됨
- mypy 미설치

**When**:
```bash
moai doctor
```

**Then**:
- 출력에 "Language: Python" 표시
- pytest ✓ (installed) 표시
- mypy ✗ (not installed) 표시
- 설치 명령어 제안: `pip install mypy`
- Exit code: 0 (경고는 있지만 성공)

### Scenario 1.2: doctor --verbose 버전 표시

**Given**:
- Python 프로젝트
- pytest 8.4.2, mypy 1.13.0 설치됨

**When**:
```bash
moai doctor --verbose
```

**Then**:
- pytest ✓ (8.4.2) 표시
- mypy ✓ (1.13.0) 표시
- ruff 상태 표시
- 전체 실행 시간 < 5초

### Scenario 1.3: doctor --fix 자동 수정 제안

**Given**:
- TypeScript 프로젝트
- package.json 존재
- Vitest 미설치

**When**:
```bash
moai doctor --fix
```

**Then**:
- "Vitest not installed" 메시지
- 설치 명령어 제안: `npm install -D vitest`
- 사용자 승인 요청: "Install Vitest? (y/n)"
- "y" 입력 시 설치 실행
- 설치 후 재검증

---

## AC-2: doctor 명령어 - 언어 미감지 처리

### Scenario 2.1: 언어 감지 실패

**Given**:
- pyproject.toml, package.json 모두 없음
- 언어 식별 불가

**When**:
```bash
moai doctor
```

**Then**:
- "Language: Unknown" 표시
- 기본 도구만 검증 (Python, Git)
- 가이드 메시지: ".moai/config.json에서 언어를 수동 설정할 수 있습니다"
- Exit code: 0

---

## AC-3: doctor 명령어 - 성능 제약

### Scenario 3.1: 진단 시간 5초 이하

**Given**:
- Python 프로젝트
- 20개 도구 체크 필요

**When**:
```bash
time moai doctor --verbose
```

**Then**:
- 실행 시간 < 5초
- 진행 상황 표시줄 표시 (5초 가까이 소요 시)
- 모든 도구 체크 완료

---

## AC-4: status 명령어 - SPEC 체인 무결성

### Scenario 4.1: 완벽한 SPEC 체인

**Given**:
- 1개 SPEC: `.moai/specs/SPEC-AUTH-001/spec.md`

**When**:
```bash
moai status --detail
```

**Then**:
- "SPEC Chain: 100% (0 orphans, 0 broken)" 표시
- 정상 상태 아이콘 ✓

### Scenario 4.2: 끊어진 SPEC 체인 감지

**Given**:

**When**:
```bash
moai status --detail
```

**Then**:
- "SPEC Chain: ⚠ 50% (0 orphans, 1 broken)" 표시
- 끊어진 SPEC 상세: "AUTH-001: CODE exists but SPEC missing"
- 수정 가이드: "Run `/alfred:1-plan AUTH-001` to create missing SPEC"
- Exit code: 0 (경고지만 실행 성공)

### Scenario 4.3: 고아 SPEC 감지

**Given**:

**When**:
```bash
moai status --detail
```

**Then**:
- "SPEC Chain: ⚠ 50% (1 orphan, 0 broken)" 표시
- 고아 SPEC 상세: "AUTH-001: SPEC exists but CODE not implemented"
- 수정 가이드: "Run `/alfred:2-run AUTH-001` to implement"

---

## AC-5: status 명령어 - 테스트 커버리지

### Scenario 5.1: 목표 달성 (85% 이상)

**Given**:
- pytest-cov 결과: 85.61%
- `.coverage.json` 파일 존재

**When**:
```bash
moai status --detail
```

**Then**:
- "Test Coverage: 85.61% ✓ (goal: 85%)" 표시
- 초록색 체크 마크
- Exit code: 0

### Scenario 5.2: 목표 미달 (85% 미만)

**Given**:
- pytest-cov 결과: 72.06%

**When**:
```bash
moai status --detail
```

**Then**:
- "Test Coverage: 72.06% ✗ (goal: 85%)" 표시
- 빨간색 X 마크
- 부족 비율 표시: "Need +12.94% coverage"
- 가이드: "Run `pytest --cov` to improve coverage"
- Exit code: 0 (경고지만 성공)

### Scenario 5.3: 커버리지 데이터 없음

**Given**:
- `.coverage.json` 파일 없음

**When**:
```bash
moai status --detail
```

**Then**:
- "Test Coverage: N/A" 표시
- 가이드: "Run `pytest --cov --cov-report=json` to generate coverage"

---

## AC-6: status 명령어 - 코드 품질 지표

### Scenario 6.1: 완벽한 품질 (경고 0개)

**Given**:
- ruff 실행 결과: 0개 경고

**When**:
```bash
moai status --detail
```

**Then**:
- "Code Quality: 0 warnings, 0 complexity issues" 표시
- 초록색 ✓

### Scenario 6.2: 린터 경고 존재

**Given**:
- ruff 실행 결과: 3개 경고

**When**:
```bash
moai status --detail
```

**Then**:
- "Code Quality: 3 warnings" 표시
- 노란색 ⚠
- 가이드: "Run `ruff check src --fix` to auto-fix"

---

## AC-7: status 명령어 - JSON 출력 (CI/CD)

### Scenario 7.1: JSON 형식 출력

**Given**:
- 프로젝트 초기화됨

**When**:
```bash
moai status --json
```

**Then**:
- JSON 형식 출력 (Rich UI 없음)
- 파싱 가능한 구조:
  ```json
  {
    "mode": "team",
    "locale": "ko",
    "specs": 1,
    "branch": "develop",
    "git_status": "clean"
  }
  ```
- Exit code: 0

### Scenario 7.2: JSON + --detail 옵션

**Given**:
- 프로젝트 초기화됨

**When**:
```bash
moai status --json --detail
```

**Then**:
- JSON에 추가 필드:
  ```json
  {
    "mode": "team",
    "specs": 1,
    "tag_chain": {
      "integrity": 100.0,
      "orphans": 0,
      "broken": 0
    },
    "coverage": 85.61,
    "quality": {
      "warnings": 0,
      "complexity_issues": 0
    }
  }
  ```

---

## AC-8: status 명령어 - 미초기화 프로젝트

### Scenario 8.1: .moai 디렉토리 없음

**Given**:
- `.moai/` 디렉토리 없음

**When**:
```bash
moai status
```

**Then**:
- "⚠ No .moai/config.json found" 메시지
- 가이드: "Run `moai init .` to initialize project"
- Exit code: 1 (실패)

---

## AC-9: restore 명령어 - 백업 목록 표시

### Scenario 9.1: 여러 백업 존재

**Given**:
- 3개 백업 존재:
  - backup-001 (2025-10-15 10:00, 50 files, 2.3MB)
  - backup-002 (2025-10-15 12:00, 52 files, 2.5MB)
  - backup-003 (2025-10-15 14:00, 48 files, 2.1MB)

**When**:
```bash
moai restore --list
```

**Then**:
- Rich 테이블 출력:
  ```
  ┏━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━┳━━━━━━━┓
  ┃ Backup ID  ┃ Timestamp        ┃ Files ┃ Size  ┃
  ┡━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━╇━━━━━━━╇━━━━━━━┩
  │ backup-003 │ 2025-10-15 14:00 │ 48    │ 2.1MB │
  │ backup-002 │ 2025-10-15 12:00 │ 52    │ 2.5MB │
  │ backup-001 │ 2025-10-15 10:00 │ 50    │ 2.3MB │
  └────────────┴──────────────────┴───────┴───────┘
  ```
- 최신 백업이 상단 (역순 정렬)

### Scenario 9.2: 백업 없음

**Given**:
- `.moai-backups/*` 백업 디렉토리가 존재하지 않음

**When**:
```bash
moai restore --list
```

**Then**:
- "No backups found" 메시지
- 가이드: "Run `moai backup` to create a backup"
- Exit code: 0

---

## AC-10: restore 명령어 - dry-run 복원 미리보기

### Scenario 10.1: 복원 미리보기

**Given**:
- backup-001 존재 (5개 파일)
- 현재 config.json이 수정됨

**When**:
```bash
moai restore --dry-run backup-001
```

**Then**:
- "Will restore 5 files:" 메시지
- 파일 목록:
  ```
  - .moai/config.json (modified)
  - .moai/project/product.md
  - .moai/project/structure.md
  - .moai/project/tech.md
  - .moai/specs/SPEC-AUTH-001/spec.md
  ```
- config.json diff 표시 (변경 사항)
- 가이드: "Run `moai restore backup-001` to apply"
- Exit code: 0 (실제 복원 안 함)

---

## AC-11: restore 명령어 - 선택적 복원

### Scenario 11.1: 특정 파일만 복원

**Given**:
- backup-001 존재 (10개 파일)

**When**:
```bash
moai restore backup-001 --select config.json,product.md
```

**Then**:
- "Restoring 2 selected files..." 메시지
- config.json 복원됨
- product.md 복원됨
- 나머지 8개 파일은 건드리지 않음
- "Restored 2 files successfully" 메시지
- Exit code: 0

### Scenario 11.2: 존재하지 않는 파일 선택

**Given**:
- backup-001에 nonexistent.md 없음

**When**:
```bash
moai restore backup-001 --select nonexistent.md
```

**Then**:
- "⚠ File not found in backup: nonexistent.md" 경고
- 복원 건너뜀
- Exit code: 1

---

## AC-12: restore 명령어 - Git dirty state 경고

### Scenario 12.1: 미커밋 변경사항 감지

**Given**:
- Git 저장소
- config.json 수정 후 미커밋

**When**:
```bash
moai restore backup-001
```

**Then**:
- "⚠ Uncommitted changes detected" 경고
- Git status 요약:
  ```
  Modified files:
    - .moai/config.json
  ```
- 확인 요청: "Proceed? (y/n)"
- "n" 입력 시 복원 중단, Exit code: 1
- "y" 입력 시 복원 진행, Exit code: 0

### Scenario 12.2: Clean working tree

**Given**:
- Git 저장소
- 미커밋 변경사항 없음

**When**:
```bash
moai restore backup-001
```

**Then**:
- Git 경고 없이 즉시 복원
- Exit code: 0

---

## AC-13: restore 명령어 - 백업 전 스냅샷

### Scenario 13.1: 복원 전 자동 백업

**Given**:
- backup-001 복원 요청

**When**:
```bash
moai restore backup-001
```

**Then**:
- 복원 전 현재 상태 자동 백업
- 백업 ID 생성: `backup-pre-restore-<timestamp>`
- 복원 진행
- 롤백 안내: "To rollback, run `moai restore backup-pre-restore-<timestamp>`"

---

## AC-14: doctor 명령어 - 도구 버전 파싱 실패

### Scenario 14.1: 버전 정보 가져오기 실패

**Given**:
- pytest 설치되어 있지만 `pytest --version` 실패

**When**:
```bash
moai doctor --verbose
```

**Then**:
- pytest ✓ (version: unknown) 표시
- 경고 없이 계속 진행
- Exit code: 0

---

## AC-15: doctor 명령어 - 오프라인 동작

### Scenario 15.1: 인터넷 연결 없음

**Given**:
- 인터넷 연결 끊김
- 로컬 도구들은 설치됨

**When**:
```bash
moai doctor
```

**Then**:
- 로컬 도구 체크 정상 수행
- 온라인 기능 스킵 표시: "⚠ Skipping online version check (offline)"
- Exit code: 0

---

## AC-16: doctor 명령어 - JSON 출력

### Scenario 16.1: JSON 형식으로 진단 결과 저장

**Given**:
- Python 프로젝트

**When**:
```bash
moai doctor --export diagnosis.json
```

**Then**:
- `diagnosis.json` 파일 생성
- JSON 내용:
  ```json
  {
    "language": "python",
    "timestamp": "2025-10-15T14:30:00",
    "tools": {
      "python3": {"installed": true, "version": "3.13.1"},
      "pytest": {"installed": true, "version": "8.4.2"},
      "mypy": {"installed": false, "version": null}
    },
    "recommendations": [
      "Install mypy: pip install mypy"
    ]
  }
  ```
- Exit code: 0

---

## 검증 방법

### 자동화된 테스트

모든 AC는 다음 테스트로 검증:
1. **단위 테스트**: `tests/unit/test_doctor_advanced.py`, `test_status_advanced.py`, `test_restore_advanced.py`
2. **통합 테스트**: `tests/integration/test_cli_advanced_integration.py`
3. **E2E 테스트**: `tests/e2e/test_full_cli_workflow.py`

### 수동 검증 체크리스트

- [ ] AC-1: doctor 명령어 Python 도구 감지
- [ ] AC-2: doctor 언어 미감지 처리
- [ ] AC-3: doctor 성능 제약 (< 5초)
- [ ] AC-4: status SPEC 체인 무결성
- [ ] AC-5: status 테스트 커버리지
- [ ] AC-6: status 코드 품질 지표
- [ ] AC-7: status JSON 출력
- [ ] AC-8: status 미초기화 프로젝트
- [ ] AC-9: restore 백업 목록
- [ ] AC-10: restore dry-run
- [ ] AC-11: restore 선택적 복원
- [ ] AC-12: restore Git dirty state
- [ ] AC-13: restore 백업 전 스냅샷
- [ ] AC-14: doctor 버전 파싱 실패
- [ ] AC-15: doctor 오프라인 동작
- [ ] AC-16: doctor JSON 출력

---

## 성공 기준

**완료 조건**:
- ✅ 모든 AC (16개) 자동화 테스트 통과
- ✅ 수동 검증 체크리스트 100% 완료
- ✅ 테스트 커버리지 ≥85%
- ✅ Exit code 정확성 검증

**품질 기준**:
- 모든 AC는 독립적으로 실행 가능
- 각 AC는 3초 이내 검증 완료
- 실패 시 명확한 에러 메시지 제공
