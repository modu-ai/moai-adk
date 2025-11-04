---
title: config.json 수동 편집 가이드
description: .moai/config.json 파일 구조, 안전한 편집 방법, 유효성 검증 및 복구 절차
version: 1.0.0
created_at: 2025-11-04
last_updated: 2025-11-04
language: Korean
---

# config.json 수동 편집 가이드

> **대상**: MoAI-ADK 개발자 및 프로젝트 소유자
> **용도**: `.moai/config.json` 파일의 안전한 수동 편집
> **주의**: 잘못된 편집은 프로젝트 초기화를 방해할 수 있습니다

---

## 📋 목차

1. [config.json 개요](#configjson-개요)
2. [파일 구조](#파일-구조)
3. [필드별 가이드](#필드별-가이드)
4. [안전한 편집 절차](#안전한-편집-절차)
5. [유효성 검증](#유효성-검증)
6. [일반적인 실수](#일반적인-실수)
7. [문제 해결 및 복구](#문제-해결-및-복구)

---

## config.json 개요

### 역할

`.moai/config.json`은 MoAI-ADK 프로젝트의 **중요한 설정 파일**입니다:

- 🔧 프로젝트 메타데이터 (이름, 모드, 언어)
- 🎯 Alfred 워크플로우 설정
- 🏷️ TAG 시스템 구성
- 🔐 보안 및 권한 설정
- 📊 프로젝트 최적화 상태

### 중요성

| 영향도 | 내용 |
|------|------|
| **매우 높음** | git_strategy, project.mode, project.optimized |
| **높음** | language, tags, constitution |
| **중간** | hooks, github |
| **낮음** | version_check, cache_ttl |

---

## 파일 구조

### 전체 구조 예시

```json
{
  "_meta": {
    "@CODE:CONFIG-STRUCTURE-001": "@DOC:JSON-CONFIG-001"
  },
  "moai": {
    "version": "0.16.0",
    "update_check_frequency": "daily",
    "version_check": {
      "enabled": true,
      "cache_ttl_hours": 24
    }
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  },
  "constitution": {
    "enforce_tdd": true,
    "require_tags": true,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "personal": {
      "auto_checkpoint": "event-driven",
      "branch_prefix": "feature/"
    },
    "team": {
      "auto_pr": true,
      "use_gitflow": true,
      "default_pr_base": "develop"
    }
  },
  "project": {
    "name": "MyProject",
    "mode": "personal",
    "language": "python",
    "optimized": false
  },
  "tags": {
    "auto_sync": true,
    "storage_type": "code_scan"
  },
  "hooks": {
    "timeout_ms": 5000,
    "graceful_degradation": true
  }
}
```

---

## 필드별 가이드

### ✅ 안전하게 편집 가능한 필드

#### 1. language 섹션 (언어 설정)

```json
"language": {
  "conversation_language": "ko",        // ✅ 수정 가능
  "conversation_language_name": "한국어" // ✅ 수정 가능
}
```

**허용되는 언어 코드**:
- `"ko"` → 한국어
- `"en"` → English
- `"ja"` → 日本語
- `"zh"` → 中文
- `"es"` → Español

**수정 방법**:
```json
// 영어로 변경
"language": {
  "conversation_language": "en",
  "conversation_language_name": "English"
}
```

#### 2. project.name (프로젝트 이름)

```json
"project": {
  "name": "MyProject"  // ✅ 수정 가능 (디스플레이용)
}
```

**주의**: 프로젝트 폴더 이름과 다를 수 있습니다 (영향 없음)

#### 3. moai.version (버전)

```json
"moai": {
  "version": "0.16.0"  // ✅ 수정 가능 (정보용, 자동 갱신됨)
}
```

**주의**: 자동으로 갱신되므로 수동 편집 불필요

### ⚠️ 신중하게 편집해야 할 필드

#### 1. project.mode (프로젝트 모드)

```json
"project": {
  "mode": "personal"  // ⚠️ 신중하게 변경
}
```

**값**:
- `"personal"` - 개인 프로젝트 (기본값)
- `"team"` - 팀 프로젝트 (GitFlow 사용)

**변경 영향**:
- Git 워크플로우 변경
- PR 생성 규칙 변경
- 브랜치 전략 변경

**안전한 변경 절차**:
```bash
# 1. 백업
cp .moai/config.json .moai/config.json.backup

# 2. 편집 (personal → team)
# "mode": "team"

# 3. 확인
git status  # 워크플로우 변경 감지

# 4. 문제 발생 시 복구
cp .moai/config.json.backup .moai/config.json
```

#### 2. constitution.enforce_tdd (TDD 강제)

```json
"constitution": {
  "enforce_tdd": true  // ⚠️ 신중하게 변경
}
```

**영향**:
- TDD 워크플로우 강제 여부
- RED → GREEN → REFACTOR 커밋 확인

**변경 예시**:
```json
"constitution": {
  "enforce_tdd": false  // TDD 비활성화 (비권장)
}
```

#### 3. tags.auto_sync (TAG 자동 동기화)

```json
"tags": {
  "auto_sync": true  // ⚠️ 신중하게 변경
}
```

**영향**:
- @TAG 마커 자동 검증
- 코드-문서 추적성

**권장**: 항상 `true` 유지

### ❌ 편집하면 안 되는 필드

#### 1. git_strategy (Git 전략)

```json
"git_strategy": {
  "personal": { /* ... */ },
  "team": { /* ... */ }
}  // ❌ 수정하지 마세요
```

**이유**: Alfred 명령과 연동됨, 변경 시 워크플로우 오류

**복구 방법**:
```bash
# /alfred:0-project 실행하여 리셋
/alfred:0-project
```

#### 2. hooks 섹션

```json
"hooks": {
  "timeout_ms": 5000,
  "graceful_degradation": true
}  // ❌ 수정하지 마세요
```

**이유**: Claude Code Hook 설정과 연동

#### 3. _meta 섹션

```json
"_meta": {
  "@CODE:CONFIG-STRUCTURE-001": "@DOC:JSON-CONFIG-001"
}  // ❌ 수정하지 마세요
```

**이유**: TAG 추적 메타데이터

---

## 안전한 편집 절차

### Step 1: 백업 생성

```bash
# 편집 전 항상 백업
cp .moai/config.json .moai/config.json.backup

# 또는 git으로 추적
git add .moai/config.json
git commit -m "backup: Pre-edit config snapshot"
```

### Step 2: 필드 확인

편집 전 확인해야 할 항목:

- [ ] 수정할 필드가 "안전하게 편집 가능"에 있는가?
- [ ] 영향 범위를 이해하고 있는가?
- [ ] 백업이 준비되어 있는가?
- [ ] 유효성 검증 방법을 알고 있는가?

### Step 3: 파일 편집

**방법 1: 에디터로 열기**

```bash
# VS Code
code .moai/config.json

# Vim
vim .moai/config.json

# nano
nano .moai/config.json
```

**방법 2: 명령행으로 편집**

```bash
# jq 도구 사용 (설치: brew install jq)
jq '.language.conversation_language = "en"' .moai/config.json > tmp.json && mv tmp.json .moai/config.json
```

### Step 4: 유효성 검증

```bash
# JSON 문법 검증
python3 -m json.tool .moai/config.json > /dev/null && echo "✅ Valid JSON"

# 또는
cat .moai/config.json | jq . > /dev/null && echo "✅ Valid JSON"
```

### Step 5: 테스트

```bash
# 프로젝트 상태 확인
moai-adk status

# 또는 Claude Code에서
/alfred:0-project --dry-run
```

### Step 6: 커밋

```bash
# 변경사항 확인
git diff .moai/config.json

# 커밋
git add .moai/config.json
git commit -m "config: Update language to English"

# 푸시
git push origin develop
```

---

## 유효성 검증

### 자동 검증 명령

```bash
# 전체 프로젝트 상태 확인
moai-adk status

# config.json만 검증
python3 << 'EOF'
import json
from pathlib import Path

config_path = Path(".moai/config.json")
try:
    config = json.loads(config_path.read_text())
    print("✅ config.json 유효함")
    print(f"  - Project: {config.get('project', {}).get('name')}")
    print(f"  - Mode: {config.get('project', {}).get('mode')}")
    print(f"  - Language: {config.get('language', {}).get('conversation_language')}")
except json.JSONDecodeError as e:
    print(f"❌ JSON 문법 오류: {e}")
except Exception as e:
    print(f"❌ 오류: {e}")
EOF
```

### 필드별 검증

#### language 필드 검증

```bash
python3 << 'EOF'
import json

config = json.loads(open(".moai/config.json").read())
lang = config.get("language", {}).get("conversation_language")

valid_langs = ["ko", "en", "ja", "zh", "es"]
if lang in valid_langs:
    print(f"✅ 유효한 언어: {lang}")
else:
    print(f"❌ 유효하지 않은 언어: {lang}")
    print(f"   허용값: {valid_langs}")
EOF
```

#### project.mode 검증

```bash
python3 << 'EOF'
import json

config = json.loads(open(".moai/config.json").read())
mode = config.get("project", {}).get("mode")

valid_modes = ["personal", "team"]
if mode in valid_modes:
    print(f"✅ 유효한 모드: {mode}")
else:
    print(f"❌ 유효하지 않은 모드: {mode}")
    print(f"   허용값: {valid_modes}")
EOF
```

---

## 일반적인 실수

### ❌ 실수 1: JSON 문법 오류

**잘못된 예**:
```json
{
  "language": {
    "conversation_language": "ko"  // ← 마지막 쉼표 제거 필요
    "conversation_language_name": "한국어"
  }
}
```

**올바른 예**:
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**확인 방법**:
```bash
python3 -m json.tool .moai/config.json
```

### ❌ 실수 2: 필드 타입 불일치

**잘못된 예**:
```json
{
  "constitution": {
    "test_coverage_target": "85"  // ← 문자열이 아니라 숫자여야 함
  }
}
```

**올바른 예**:
```json
{
  "constitution": {
    "test_coverage_target": 85
  }
}
```

### ❌ 실수 3: 필수 필드 삭제

**잘못된 예**:
```json
{
  "moai": {
    // "version" 필드 삭제됨 ← 문제!
  }
}
```

**해결책**:
```bash
# 백업에서 복구
cp .moai/config.json.backup .moai/config.json

# 또는 /alfred:0-project 실행
/alfred:0-project
```

### ❌ 실수 4: 필드 경로 오류

**잘못된 예**:
```json
{
  "project_name": "MyProject"  // ← 잘못된 위치
}
```

**올바른 예**:
```json
{
  "project": {
    "name": "MyProject"  // ← 올바른 위치
  }
}
```

### ❌ 실수 5: 비활성화된 필드 편집

**잘못된 예**:
```json
{
  "git_strategy": {
    "personal": {
      "custom_option": "value"  // ← 존재하지 않는 필드
    }
  }
}
```

**해결책**:
```bash
# git_strategy는 복구
cp .moai/config.json.backup .moai/config.json
```

---

## 문제 해결 및 복구

### 상황 1: JSON 문법 오류로 프로젝트 실행 불가

**증상**:
```
Error: Failed to load config.json
json.JSONDecodeError: Expecting ',' delimiter
```

**해결책**:
```bash
# 1. 백업이 있는지 확인
ls -la .moai/config.json.backup

# 2. 복구
cp .moai/config.json.backup .moai/config.json

# 3. 검증
python3 -m json.tool .moai/config.json

# 4. 다시 시도
moai-adk status
```

### 상황 2: 잘못된 값으로 워크플로우 오류

**증상**:
```
Invalid project mode: invalid_mode
```

**원인**: 존재하지 않는 값으로 편집

**해결책**:
```bash
# 1. 현재 값 확인
jq '.project.mode' .moai/config.json

# 2. 유효한 값으로 수정
jq '.project.mode = "personal"' .moai/config.json > tmp.json && mv tmp.json .moai/config.json

# 3. 확인
jq '.project.mode' .moai/config.json  # "personal" 출력됨
```

### 상황 3: 실수로 중요 필드 삭제

**증상**:
```
KeyError: 'version'
```

**복구 방법**:

**방법 1: 백업에서 복구 (권장)**
```bash
cp .moai/config.json.backup .moai/config.json
```

**방법 2: Git 히스토리에서 복구**
```bash
# 커밋된 상태로 복구
git checkout HEAD -- .moai/config.json
```

**방법 3: /alfred:0-project로 재생성**
```bash
# 프로젝트 초기화 (사용자 설정 유지)
/alfred:0-project
```

### 상황 4: 언어 설정 오류

**증상**:
```
Invalid language: invalid_lang
```

**해결책**:
```bash
# 유효한 언어 목록으로 수정
jq '.language.conversation_language = "ko"' .moai/config.json > tmp.json && mv tmp.json .moai/config.json

# 언어 이름도 맞추기
jq '.language.conversation_language_name = "한국어"' .moai/config.json > tmp.json && mv tmp.json .moai/config.json
```

---

## 고급: 일괄 편집

### jq를 사용한 다중 필드 편집

```bash
# 여러 필드를 한 번에 수정
jq '
  .language.conversation_language = "en" |
  .language.conversation_language_name = "English" |
  .project.name = "NewProject"
' .moai/config.json > tmp.json && mv tmp.json .moai/config.json
```

### Python을 사용한 프로그래매틱 편집

```python
#!/usr/bin/env python3
import json
from pathlib import Path

config_path = Path(".moai/config.json")

# 읽기
config = json.loads(config_path.read_text())

# 수정
config["language"]["conversation_language"] = "en"
config["language"]["conversation_language_name"] = "English"
config["project"]["name"] = "NewProject"

# 백업
config_path.write_text(json.dumps(config, indent=2, ensure_ascii=False))

print("✅ 설정 업데이트 완료")
```

---

## 체크리스트

### 편집 전 확인

- [ ] 백업 생성했는가? (`cp .moai/config.json .moai/config.json.backup`)
- [ ] 수정할 필드가 "안전하게 편집 가능"에 있는가?
- [ ] 필드 경로가 올바른가?
- [ ] 값의 타입이 맞는가? (문자열 vs 숫자)

### 편집 후 검증

- [ ] JSON 문법이 유효한가? (`python3 -m json.tool .moai/config.json`)
- [ ] 필드 값이 유효한가? (`moai-adk status`)
- [ ] 워크플로우가 정상인가? (`/alfred:0-project --dry-run`)
- [ ] 변경사항이 예상대로 반영되었는가?

### 커밋 전 확인

- [ ] 테스트 완료했는가?
- [ ] 변경사항을 이해하고 있는가? (`git diff .moai/config.json`)
- [ ] 커밋 메시지가 명확한가?

---

## 참고: .moai/config.json vs .claude/settings.json

### 차이점

| 항목 | .moai/config.json | .claude/settings.json |
|------|-------------------|----------------------|
| **용도** | MoAI-ADK 설정 | Claude Code 설정 |
| **관리자** | 사용자 | Claude Code |
| **수정 빈도** | 낮음 | 거의 없음 |
| **위험도** | 높음 | 중간 |
| **백업 필요** | ✅ 필요 | △ 권장 |
| **자동 갱신** | O | O |

### 각각 편집 가능한 필드

**config.json**:
- ✅ language (언어)
- ✅ project.name (프로젝트명)
- ⚠️ project.mode (개인/팀)

**settings.json**:
- △ permissions (권한, 신중하게)
- ✅ env (환경변수)

---

## 정리

| 작업 | 난이도 | 위험도 | 필요시간 |
|------|------|------|--------|
| 언어 변경 | 낮음 | 낮음 | 1분 |
| 프로젝트명 변경 | 낮음 | 낮음 | 1분 |
| 모드 변경 | 중간 | 높음 | 5분 |
| TDD 설정 변경 | 중간 | 중간 | 3분 |
| 전체 재설정 | 높음 | 높음 | 10분 |

**권장**: 가능하면 `/alfred:0-project`를 사용하세요. 수동 편집은 작은 변경에만 사용하세요.

---

**문서 버전**: 1.0.0
**마지막 업데이트**: 2025-11-04
**작성자**: Alfred (MoAI-ADK SuperAgent)

🤖 Generated with Claude Code
Co-Authored-By: 🎩 Alfred@MoAI
