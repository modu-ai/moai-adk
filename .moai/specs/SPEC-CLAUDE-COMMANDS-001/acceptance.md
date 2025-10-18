# 수락 기준 - CLAUDE-COMMANDS-001

> **SPEC**: `.moai/specs/SPEC-CLAUDE-COMMANDS-001/spec.md`

---

## 수락 기준 개요

이 SPEC의 구현이 완료되었다고 판단하려면 다음 3가지 시나리오가 모두 통과해야 합니다.

---

## Scenario 1: 정상 커맨드 로드

### Given (전제 조건)

- `.claude/commands/` 디렉토리에 유효한 `.md` 파일 4개가 있음:
  - `alfred-0-project.md`
  - `alfred-1-spec.md`
  - `alfred-2-build.md`
  - `alfred-3-sync.md`
- 각 파일은 올바른 Front Matter와 필수 필드(name, description)를 포함

### When (실행 동작)

- Claude Code를 시작함

### Then (예상 결과)

- 로그에 다음 메시지 출력:
  ```
  [DEBUG] Total plugin commands loaded: 4
  [DEBUG] Slash commands included in SlashCommand tool:
    - /alfred:0-project
    - /alfred:1-spec
    - /alfred:2-build
    - /alfred:3-sync
  ```
- 각 커매드를 `/` 접두사로 호출 가능함
- 커맨드 실행 시 오류 없이 정상 동작함

### 검증 방법

```bash
# 1. Claude Code 재시작 후 로그 확인
grep "Total plugin commands loaded" ~/.local/share/claude-code/logs/debug.log

# 2. 커맨드 호출 테스트
# Claude Code UI에서 `/alfred:1-spec` 입력 후 자동완성 확인

# 3. 실제 실행 테스트
# 커맨드 실행 후 에러 없이 정상 동작 확인
```

---

## Scenario 2: 일부 커맨드 오류

### Given (전제 조건)

- `.claude/commands/` 디렉토리에 5개 파일이 있음
- 그중 2개 파일이 형식 오류:
  - `broken-1.md`: Front Matter 누락
  - `broken-2.md`: 필수 필드 (name) 누락
- 나머지 3개 파일은 유효함

### When (실행 동작)

- Claude Code를 시작함

### Then (예상 결과)

- 로그에 다음 메시지 출력:
  ```
  [DEBUG] Total plugin commands loaded: 3
  [WARNING] Failed to load command from broken-1.md: Missing YAML front matter
  [WARNING] Failed to load command from broken-2.md: Missing required field: name
  ```
- 유효한 3개 커맨드는 정상 로드됨
- 오류 파일명과 이유가 명확하게 출력됨

### 검증 방법

```bash
# 1. 진단 도구 실행
python -m moai_adk doctor --check-commands

# 예상 출력:
# ✅ Valid: 3 files
# ❌ Invalid: 2 files
#   - broken-1.md: Missing YAML front matter
#   - broken-2.md: Missing required field: name

# 2. Claude Code 로그 확인
grep "WARNING.*Failed to load command" ~/.local/share/claude-code/logs/debug.log
```

---

## Scenario 3: 진단 도구 실행

### Given (전제 조건)

- 슬래시 커맨드가 로드되지 않는 상황
- `.claude/commands/` 디렉토리에 문제가 있는 파일들 존재

### When (실행 동작)

- 진단 도구 실행:
  ```bash
  python -m moai_adk doctor --check-commands
  ```

### Then (예상 결과)

- 각 `.md` 파일의 검증 결과 출력 (테이블 형식):
  ```
  ╔════════════════════════════════╦════════╦══════════════════════════════════════╗
  ║ File                           ║ Status ║ Errors                               ║
  ╠════════════════════════════════╬════════╬══════════════════════════════════════╣
  ║ alfred-0-project.md            ║   ✅   ║ -                                    ║
  ║ alfred-1-spec.md               ║   ✅   ║ -                                    ║
  ║ alfred-2-build.md              ║   ✅   ║ -                                    ║
  ║ alfred-3-sync.md               ║   ✅   ║ -                                    ║
  ║ broken-1.md                    ║   ❌   ║ Missing YAML front matter            ║
  ║ broken-2.md                    ║   ❌   ║ Missing required field: name         ║
  ╚════════════════════════════════╩════════╩══════════════════════════════════════╝

  Summary:
    Total files: 6
    Valid: 4
    Invalid: 2

  Recommendations:
    1. Fix broken-1.md: Add YAML front matter (---)
    2. Fix broken-2.md: Add 'name' field in front matter
  ```

- 오류 원인과 해결 방법 제시
- JSON 출력 옵션 지원:
  ```bash
  python -m moai_adk doctor --check-commands --json
  ```
  ```json
  {
    "total_files": 6,
    "valid_commands": 4,
    "invalid_commands": 2,
    "details": [
      {
        "file": "alfred-0-project.md",
        "valid": true,
        "errors": []
      },
      {
        "file": "broken-1.md",
        "valid": false,
        "errors": ["Missing YAML front matter"]
      }
    ]
  }
  ```

### 검증 방법

```bash
# 1. 진단 도구 실행 및 출력 확인
python -m moai_adk doctor --check-commands | grep "Total files"

# 2. JSON 출력 확인
python -m moai_adk doctor --check-commands --json | jq '.total_files'

# 3. 권장사항 확인
python -m moai_adk doctor --check-commands | grep "Recommendations"
```

---

## 품질 게이트 기준

### 기능 요구사항

- ✅ **진단 도구 구현**: `doctor --check-commands` 명령어 동작
- ✅ **파일 검증**: 모든 `.md` 파일 형식 검증
- ✅ **오류 보고**: 명확한 오류 메시지 및 해결 방법 제시
- ✅ **Claude Code 통합**: 커맨드 정상 로드 확인

### 비기능 요구사항

- ✅ **성능**: 진단 도구 실행 시간 < 2초 (파일 100개 기준)
- ✅ **사용성**: 사용자가 오류를 쉽게 이해하고 수정 가능
- ✅ **확장성**: 새로운 검증 규칙 추가 용이
- ✅ **유지보수성**: 코드 복잡도 ≤ 10, 함수당 ≤ 50 LOC

### 테스트 커버리지

- ✅ **단위 테스트**: 커버리지 ≥ 85%
- ✅ **통합 테스트**: 3가지 시나리오 모두 통과
- ✅ **회귀 테스트**: 기존 기능 영향 없음

---

## 검증 체크리스트

### Phase 1: 진단

- [ ] `.claude/commands/` 디렉토리 스캔 성공
- [ ] 모든 `.md` 파일 발견
- [ ] 각 파일 형식 검증 완료
- [ ] 오류 파일 식별 완료

### Phase 2: 수정

- [ ] 형식 오류 파일 수정 완료
- [ ] 필수 메타데이터 추가 완료
- [ ] 인코딩 통일 (UTF-8) 완료
- [ ] 권한 확인 및 수정 완료

### Phase 3: 검증

- [ ] Claude Code 재시작 후 커맨드 로드 확인
- [ ] 로그에 정확한 개수 출력 확인
- [ ] 각 커맨드 실행 테스트 통과
- [ ] 오류 메시지 명확성 확인

### Phase 4: 자동화

- [ ] Pre-commit hook 추가 (선택사항)
- [ ] CI/CD 파이프라인 통합 (선택사항)
- [ ] 문서 업데이트 완료

---

## 완료 조건 (Definition of Done)

다음 조건이 **모두** 충족되어야 SPEC 구현 완료로 간주합니다:

1. **기능 완성**:
   - 진단 도구가 정상 동작함
   - Claude Code에서 슬래시 커맨드가 정상 로드됨
   - 모든 커맨드가 실행 가능함

2. **품질 기준**:
   - 테스트 커버리지 ≥ 85%
   - 코드 리뷰 승인 (최소 1명)
   - TRUST 5원칙 준수 확인

3. **문서화**:
   - `docs/slash-commands-guide.md` 작성 완료
   - README.md 업데이트 (진단 도구 사용법)
   - CHANGELOG.md 업데이트

4. **배포 준비**:
   - 통합 테스트 통과
   - 스테이징 환경 검증 완료
   - 릴리스 노트 작성

---

**작성일**: 2025-10-18
**작성자**: spec-builder 에이전트
