# statusline-config.yaml 필요성 분석

**문제**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.moai/config/statusline-config.yaml` 파일이 필요한가?

**결론**: ✅ **필요함** (마이그레이션 및 레거시 호환성을 위해)

---

## 1️⃣ 현재 상황 분석

### 파일 위치

```
패키지 템플릿:
└─ src/moai_adk/templates/.moai/config/statusline-config.yaml

프로젝트 로컬:
└─ .moai/config/statusline-config.yaml
```

### 파일 크기
- **패키지 템플릿**: 86줄
- **내용**: Statusline 표시 모드, 색상, 캐시 설정 등

---

## 2️⃣ 사용처 분석

### 코드에서의 참조

#### 1. `statusline/config.py` (설정 로더)
```python
# 설정 검색 위치 (우선순위 순)
locations = [
    Path.cwd() / ".moai" / "config" / "statusline-config.yaml",     # 1순위 (NEW)
    Path.cwd() / ".moai" / "config" / "statusline-config.yml",      # 2순위
    Path.home() / ".moai" / "config" / "statusline-config.yaml",    # 3순위
]
```

**역할**: 프로젝트별 statusline 커스터마이제이션 지원

#### 2. `core/migration/version_detector.py` (마이그레이션)
```python
old_statusline = self.project_root / ".claude" / "statusline-config.yaml"
new_statusline = self.project_root / ".moai" / "config" / "statusline-config.yaml"

# v0.23.0 → v0.24.0 마이그레이션 시
# .claude/statusline-config.yaml → .moai/config/statusline-config.yaml 이동
```

**역할**: 레거시 위치에서 새 위치로 마이그레이션

#### 3. `core/migration/backup_manager.py` (백업 대상)
```python
backup_targets = [
    self.project_root / ".moai" / "config.json",
    self.project_root / ".moai" / "config" / "config.json",
    self.project_root / ".claude" / "statusline-config.yaml",      # 레거시
    self.project_root / ".moai" / "config" / "statusline-config.yaml", # 신규
]
```

**역할**: 마이그레이션 중 설정 파일 백업

#### 4. `core/migration/file_migrator.py` (마이그레이션 안전 패턴)
```python
safe_patterns = [
    ".moai/config.json",
    ".claude/statusline-config.yaml",  # 안전 패턴 (레거시)
]
```

**역할**: 레거시 파일 안전 복사 패턴

#### 5. `cli/commands/update.py` (사용자 안내)
```python
console.print("   • .moai/config.json → .moai/config/config.json")
console.print("   • .claude/statusline-config.yaml → .moai/config/statusline-config.yaml")
```

**역할**: 마이그레이션 안내 메시지

---

## 3️⃣ 마이그레이션 히스토리

### v0.23.0 (레거시)
```
프로젝트 구조:
├─ .claude/
│  └─ statusline-config.yaml  ← 원래 위치
└─ .moai/
   └─ config.json
```

### v0.24.0+ (현재)
```
프로젝트 구조:
├─ .claude/
│  └─ statusline-config.yaml  ← 레거시 (호환성)
└─ .moai/
   └─ config/
      ├─ config.json
      └─ statusline-config.yaml  ← 새 위치
```

### 마이그레이션 로직
```python
if old_statusline.exists() and not new_statusline.exists():
    # v0.23 → v0.24 마이그레이션 수행
    move_file(old_statusline, new_statusline)
```

---

## 4️⃣ 왜 패키지 템플릿에 필요한가?

### 이유 1️⃣: 신규 프로젝트 초기화
```
moai-adk init
│
├─ TemplateProcessor.copy_templates()
│  ├─ .moai/config/ 디렉토리 생성
│  └─ statusline-config.yaml 복사 ← **패키지 템플릿에서**
│
└─ .moai/config/statusline-config.yaml (초기 설정)
```

**효과**: 신규 프로젝트가 즉시 커스터마이제이션 가능

### 이유 2️⃣: 마이그레이션 대상 파일
```
update 명령어 (v0.23 → v0.24+)
│
├─ BackupManager.create_backup()
│  └─ .claude/statusline-config.yaml 백업 ← **레거시**
│
├─ VersionMigrator.needs_migration()
│  └─ 감지: "old statusline 파일이 존재한다"
│
└─ MigrationPlan: 이 파일을 이동할 예정
```

**효과**: 마이그레이션 대상 파일 존재 여부 판단

### 이유 3️⃣: 마이그레이션 검증
```
마이그레이션 완료 후:
├─ BackupManager에 기록 (backup_metadata.json)
└─ 파일 존재 여부 검증
```

**효과**: 마이그레이션 성공 여부 확인

---

## 5️⃣ 파일 내용 분석

### statusline-config.yaml의 역할

```yaml
statusline:
  enabled: true
  mode: "extended"        # 표시 모드 (compact/extended/minimal)

  colors:
    enabled: true
    theme: "auto"         # 색상 테마 (auto/light/dark)
    palette:
      model: "38;5;33"    # 색상 코드
      ...

  cache:
    git_ttl_seconds: 5    # 캐시 TTL 설정
    ...

  display:
    model: true           # 어떤 정보를 표시할지
    version: true
    branch: true
    ...

  format:
    max_branch_length: 20 # 포맷 설정
    icons:
      git: "🔀"
      ...
```

### 커스터마이제이션 예시

**사용자가 프로젝트별로 커스터마이즈 가능**:
```yaml
# .moai/config/statusline-config.yaml (프로젝트 로컬)
statusline:
  mode: "minimal"         # compact 대신 minimal 선택
  colors:
    theme: "dark"         # 다크 테마
  display:
    model: false          # 모델 정보 숨기기
    duration: false       # 시간 표시 안 함
```

---

## 6️⃣ 제거 시 영향

### ❌ 만약 패키지 템플릿에서 제거한다면?

#### 문제 1: 신규 프로젝트
```
moai-adk init
│
└─ .moai/config/ 디렉토리는 생성됨
   └─ statusline-config.yaml 는 없음 ❌
      └─ statusline/config.py가 기본값 사용

결과: 커스터마이제이션 불가능
```

#### 문제 2: 마이그레이션 감지 불가
```
update 명령어 (레거시 프로젝트)
│
└─ VersionMigrator.needs_migration()
   └─ old_statusline = ".claude/statusline-config.yaml"
      └─ 이 파일의 위치만 추적, 템플릿 필요 없음 ✓
```

마이그레이션 감지는 영향 없음 (파일이 프로젝트에 있으면 감지됨)

#### 문제 3: 초기화 실패
```
init 실패 시나리오 (만약 파일이 필수라면)
│
└─ FAIL: "statusline-config.yaml not found in templates"
```

현재: 파일 없으면 기본값 사용 (실패 없음)

---

## 7️⃣ 권장사항

### 현재 상태: ✅ 유지해야 함

**이유**:
1. **신규 프로젝트 초기화**: 기본 설정 제공
2. **마이그레이션 검증**: 대상 파일 목록에 포함
3. **백업 대상**: BackupManager의 backup_targets 목록
4. **레거시 호환성**: v0.23 → v0.24+ 마이그레이션 지원

### 개선 방안 (선택사항)

#### 1. 문서화 강화
```markdown
# statusline-config.yaml 커스터마이제이션

프로젝트별로 statusline 표시를 커스터마이즈할 수 있습니다:

1. 기본 설정: .moai/config/statusline-config.yaml
2. 복사 후 수정: 프로젝트 루트 `.moai/config/statusline-config.yaml` 수정
3. 저장: 다음 세션부터 적용됨
```

#### 2. 기본값 최적화
현재 `extended` 모드는 정보가 많을 수 있으니:
```yaml
# 현재
mode: "extended"    # 120글자

# 추천
mode: "compact"     # 80글자 (Claude Code에서 적절)
```

#### 3. 캐시 설정 조정
```yaml
# 현재 git_ttl: 5초는 자주 변경되는 프로젝트에서 과도할 수 있음
git_ttl_seconds: 10  # 10초로 증가 (캐시 히트율 ⬆)
```

---

## 8️⃣ 결론

| 항목 | 평가 |
|------|------|
| 필요성 | ✅ **필요함** |
| 유지 | ✅ **계속 유지** |
| 레거시 지원 | ✅ **필수** |
| 신규 프로젝트 | ✅ **기본값 제공 필요** |
| 사용자 커스터마이제이션 | ✅ **지원 필요** |

### 최종 권장사항

1. **현재 파일 유지** ✅
   - 패키지 템플릿에 계속 포함
   - 신규 프로젝트에 복사됨

2. **마이그레이션 검증 유지** ✅
   - `version_detector.py`의 마이그레이션 로직 계속 작동
   - 레거시 프로젝트 (v0.23) → 신규 (v0.24+) 전환 지원

3. **문서화 추가** (선택사항)
   - `.moai/memory/statusline-customization.md` 작성
   - 사용자가 statusline 커스터마이제이션 방법 학습

4. **기본값 미세 조정** (선택사항)
   - `mode: "extended"` → `"compact"` 검토
   - 캐시 TTL 최적화

---

**최종 답변**: 이 파일은 **제거하면 안 됩니다**. 신규 프로젝트 초기화와 마이그레이션 시스템에 필수적입니다.

